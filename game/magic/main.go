package main

import (
    "log"
    "fmt"
    "flag"
    "io"
    "errors"
    "math"
    "math/rand/v2"
    "slices"
    "cmp"
    "bufio"
    "compress/gzip"
    "encoding/json"
    "image"
    // "image/color"

    // for trace/pprof
    "net/http"
    _ "net/http/pprof"
    /*
    "runtime"
    "runtime/debug"
    "os"
    */

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/system"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    introlib "github.com/kazzmir/master-of-magic/game/magic/intro"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    musiclib "github.com/kazzmir/master-of-magic/game/magic/music"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/ai"
    "github.com/kazzmir/master-of-magic/game/magic/load"
    "github.com/kazzmir/master-of-magic/game/magic/serialize"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    mouselib "github.com/kazzmir/master-of-magic/lib/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/mainview"
    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type DrawFunc func(*ebiten.Image)

type MagicGame struct {
    Cache *lbx.LbxCache

    MainCoroutine *coroutine.Coroutine
    Drawer DrawFunc

    Music *musiclib.Music
}

func randomChoose[T any](choices... T) T {
    return choices[rand.N(len(choices))]
}

func runIntro(yield coroutine.YieldFunc, game *MagicGame) {
    mouse.Mouse.Disable()
    defer mouse.Mouse.Enable()

    intro, err := introlib.MakeIntro(game.Cache, introlib.DefaultAnimationSpeed)
    if err != nil {
        log.Printf("Unable to run intro: %v", err)
        return
    }

    game.Drawer = func(screen *ebiten.Image) {
        intro.Draw(screen)
    }

    for intro.Update() == introlib.IntroStateRunning {
        if yield() != nil {
            return
        }

        if inputmanager.LeftClick() ||
           inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
           inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
            return
        }
    }
}

func runNewGame(yield coroutine.YieldFunc, game *MagicGame) (bool, setup.NewGameSettings) {
    newGame := setup.MakeNewGameScreen(game.Cache)

    game.Drawer = func(screen *ebiten.Image) {
        newGame.Draw(screen)
    }

    state := newGame.Update()
    for state == setup.NewGameStateRunning {
        yield()
        state = newGame.Update()
    }

    return state == setup.NewGameStateCancel, newGame.Settings
}

func runNewWizard(yield coroutine.YieldFunc, game *MagicGame) (bool, setup.WizardCustom) {
    newWizard := setup.MakeNewWizardScreen(game.Cache)

    game.Drawer = func(screen *ebiten.Image) {
        newWizard.Draw(screen)
    }

    state := newWizard.Update()
    for state != setup.NewWizardScreenStateFinished && state != setup.NewWizardScreenStateCanceled {
        yield()
        state = newWizard.Update()
    }

    return state == setup.NewWizardScreenStateCanceled, newWizard.CustomWizard
}

func runMainMenu(yield coroutine.YieldFunc, game *MagicGame, gameLoader *OriginalGameLoader) (*gamelib.Game, mainview.MainScreenState) {
    menu := mainview.MakeMainScreen(game.Cache, gameLoader, game.Music)

    game.Drawer = func(screen *ebiten.Image) {
        menu.Draw(screen)
    }

    for menu.Update(yield) == mainview.MainScreenStateRunning {

        select {
            case newGame := <-gameLoader.NewGame:
                return newGame, mainview.MainScreenStateLoadGame
            default:
        }

        if inputmanager.IsQuitPressed() {
            return nil, mainview.MainScreenStateQuit
        }

        if yield() != nil {
            return nil, mainview.MainScreenStateQuit
        }
    }

    return nil, menu.State
}

/* starting units are swordsmen and spearmen of the appropriate race
 */
func startingUnits(race data.Race) []units.Unit {
    switch race {
        case data.RaceLizard: return []units.Unit{units.LizardSwordsmen, units.LizardSpearmen}
        case data.RaceNomad: return []units.Unit{units.NomadSwordsmen, units.NomadSpearmen}
        case data.RaceOrc: return []units.Unit{units.OrcSwordsmen, units.OrcSpearmen}
        case data.RaceTroll: return []units.Unit{units.TrollSwordsmen, units.TrollSpearmen}
        case data.RaceBarbarian: return []units.Unit{units.BarbarianSwordsmen, units.BarbarianSpearmen}
        case data.RaceBeastmen: return []units.Unit{units.BeastmenSwordsmen, units.BeastmenSpearmen}
        case data.RaceDarkElf: return []units.Unit{units.DarkElfSwordsmen, units.DarkElfSpearmen}
        case data.RaceDraconian: return []units.Unit{units.DraconianSwordsmen, units.DraconianSpearmen}
        case data.RaceDwarf: return []units.Unit{units.DwarfSwordsmen, units.DwarfSwordsmen}
        case data.RaceGnoll: return []units.Unit{units.GnollSwordsmen, units.GnollSpearmen}
        case data.RaceHalfling: return []units.Unit{units.HalflingSwordsmen, units.HalflingSpearmen}
        case data.RaceHighElf: return []units.Unit{units.HighElfSwordsmen, units.HighElfSpearmen}
        case data.RaceHighMen: return []units.Unit{units.HighMenSwordsmen, units.HighMenSpearmen}
        case data.RaceKlackon: return []units.Unit{units.KlackonSwordsmen, units.KlackonSpearmen}
        default: return nil
    }
}

func euclideanDistance(x1, y1, x2, y2 int) float64 {
    dx := float64(x1 - x2)
    dy := float64(y1 - y2)

    return math.Sqrt(dx*dx + dy*dy)
}

func findCityLocation(game *gamelib.Game, startingPlane data.Plane, cityArea gamelib.CityValidArea) (int, int) {
    allCities := game.Model.AllCities()

    closestDistance := func(x, y int) int {
        distance := -1

        for _, city := range allCities {
            if city.Plane == startingPlane {
                d := int(euclideanDistance(x, y, city.X, city.Y))
                if distance == -1 || d < distance {
                    distance = d
                }
            }
        }

        if distance == -1 {
            return 0
        } else {
            return distance
        }
    }

    type CityLocation struct {
        X, Y int
        // distance to closest city
        Distance int
        Population int
    }

    var locations []CityLocation
    for range 10 {
        // x, y, ok := game.FindValidCityLocation(startingPlane)
        x, y, ok := cityArea.FindLocation()
        if ok {
            distance := closestDistance(x, y)
            // either there are no other cities nearby (distance=0) or the closest city is farther than 10 squares away
            if distance == 0 || distance > 10 {
                locations = append(locations, CityLocation{X: x, Y: y, Distance: distance, Population: game.Model.ComputeMaximumPopulation(x, y, startingPlane)})
            }
        }
    }

    // compute a weighted sum of distance to other cities and maximum population of the location
    computeValue := func (point CityLocation) float64 {
        distance := point.Distance
        // assume a distance if no other cities are nearby
        if distance == 0 {
            distance = 150
        }

        return math.Log2(float64(distance)) * 2 + float64(point.Population) * 0.5
    }

    slices.SortFunc(locations, func(pointA, pointB CityLocation) int {
        return cmp.Compare(computeValue(pointA), computeValue(pointB))
    })

    if len(locations) > 0 {
        // choose furthest point
        return locations[len(locations) - 1].X, locations[len(locations) - 1].Y
    } else {
        // couldn't find a good spot, just pick anything
        for range 100 {
            cityX, cityY, ok := cityArea.FindLocation()
            if ok {
                return cityX, cityY
            }
        }
    }

    return -1, -1
}

func initializePlayer(game *gamelib.Game, wizard setup.WizardCustom, isHuman bool, arcanusCityArea gamelib.CityValidArea, myrrorCityArea gamelib.CityValidArea) *playerlib.Player {
    area := arcanusCityArea
    startingPlane := data.PlaneArcanus
    if wizard.RetortEnabled(data.RetortMyrran) {
        startingPlane = data.PlaneMyrror
        area = myrrorCityArea
    }

    player := game.AddPlayer(wizard, isHuman)

    cityName := game.SuggestCityName(player.Wizard.Race)

    cityX, cityY := findCityLocation(game, startingPlane, area)
    area[image.Pt(cityX, cityY)] = false

    introCity := citylib.MakeCity(cityName, cityX, cityY, player.Wizard.Race, game.Model.BuildingInfo, game.GetMap(startingPlane), game.Model, player)
    introCity.Population = 4000
    introCity.Plane = startingPlane

    for _, building := range []buildinglib.Building{buildinglib.BuildingSmithy, buildinglib.BuildingBarracks, buildinglib.BuildingBuildersHall} {
        if introCity.GetBuildableBuildings().Contains(building) {
            introCity.Buildings.Insert(building)
        }
    }

    introCity.Buildings.Insert(buildinglib.BuildingFortress)
    introCity.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    introCity.ProducingBuilding = buildinglib.BuildingHousing
    introCity.ProducingUnit = units.UnitNone
    introCity.Farmers = 4

    introCity.ResetCitizens()

    player.AddCity(introCity)

    for _, unit := range startingUnits(player.Wizard.Race) {
        player.AddUnit(units.MakeOverworldUnitFromUnit(unit, cityX, cityY, startingPlane, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    }

    // for debugging purposes, to start the game super powered
    /*
    if isHuman {
        player.Admin = true
        player.Mana = 90000
        for range 3 {
            player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, cityX, cityY, startingPlane, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
        }
    }
    */

    player.LiftFog(cityX, cityY, 3, introCity.Plane)

    if isHuman {
        game.Events <- gamelib.StartingCityEvent(introCity)
        game.Camera.Center(cityX, cityY)
        game.Model.Plane = startingPlane
    }

    return player
}

func initializeNeutralPlayer(game *gamelib.Game, arcanusCityArea gamelib.CityValidArea, myrrorCityArea gamelib.CityValidArea) *playerlib.Player {
    wizard := setup.WizardCustom{
        Name: "Raiders",
        Base: data.WizardMerlin, // doesn't really matter
        Race: data.RaceBarbarian, // doesn't really matter
        Banner: data.BannerBrown,
    }

    player := game.AddPlayer(wizard, false)
    player.AIBehavior = ai.MakeRaiderAI()
    player.TaxRate = fraction.Zero()

    for _, plane := range []data.Plane{data.PlaneArcanus, data.PlaneMyrror} {
        randomRace := func() data.Race {
            switch plane {
                case data.PlaneArcanus: return randomChoose(data.ArcanianRaces()...)
                case data.PlaneMyrror: return randomChoose(data.MyrranRaces()...)
            }

            return data.RaceNone
        }

        area := arcanusCityArea
        if plane == data.PlaneMyrror {
            area = myrrorCityArea
        }

        for range 5 {
            cityX, cityY := findCityLocation(game, plane, area)

            // should every neutral town be a random race, or should they all be related?
            race := randomRace()
            cityName := game.SuggestCityName(race)
            city := citylib.MakeCity(cityName, cityX, cityY, race, game.Model.BuildingInfo, game.GetMap(plane), game.Model, player)
            city.Population = rand.N(5) * 1000 + 2000
            city.ProducingBuilding = buildinglib.BuildingHousing
            city.Plane = plane
            city.Farmers = city.Citizens()
            city.ResetCitizens()

            area[image.Pt(cityX, cityY)] = false

            player.AddCity(city)
        }
    }

    return player
}

type OriginalGameLoader struct {
    Cache *lbx.LbxCache
    NewGame chan *gamelib.Game
    FS system.WriteableFS
}

func (loader *OriginalGameLoader) LoadMetadata(path string) (serialize.SaveMetadata, bool) {
    return serialize.LoadMetadata(loader.FS, path)
}

func (loader *OriginalGameLoader) LoadNew(path string) error {
    file, err := loader.FS.Open(path)
    if err != nil {
        return fmt.Errorf("Could not open save game file '%v': %v", path, err)
    }
    defer file.Close()

    err = loader.LoadNewReader(file)
    if err != nil {
        return fmt.Errorf("Could not load save game file '%v': %v", path, err)
    }

    return nil
}

func (loader *OriginalGameLoader) GetFS() system.WriteableFS {
    return loader.FS
}

// we assume the reader is still compressed
func (loader *OriginalGameLoader) LoadNewReader(readerOriginal io.Reader) error {
    reader := bufio.NewReader(readerOriginal)
    gzipReader, err := gzip.NewReader(reader)
    if err != nil {
        log.Printf("Error: unable to create gzip reader for save game: %v", err)
        return fmt.Errorf("Could not load")
    }
    defer gzipReader.Close()

    decoder := json.NewDecoder(gzipReader)

    var serializedGame gamelib.SerializedGame
    err = decoder.Decode(&serializedGame)
    if err != nil {
        log.Printf("Error: unable to decode save game: %v", err)
        return fmt.Errorf("Could not load")
    }

    newGame := gamelib.MakeGameFromSerialized(loader.Cache, musiclib.MakeMusic(loader.Cache), &serializedGame)
    select {
        case loader.NewGame <- newGame:
        default:
            log.Printf("Warning: unable to send new game to channel")
    }

    return nil
}

func (loader *OriginalGameLoader) Load(reader io.Reader) error {
    saved, err := load.LoadSaveGame(reader)
    if err != nil {
        return err
    }

    newGame := saved.Convert(loader.Cache)

    if newGame != nil {
        select {
            case loader.NewGame <- newGame:
            default:
        }

        return nil
    }

    return fmt.Errorf("Could not convert saved game")
}

/*
func writeHeapDump(filename string) {
    runtime.GC()
    out, err := os.Create(filename)
    if err == nil {
        debug.WriteHeapDump(out.Fd())
        out.Close()
    }
}
*/

func centerOnCity(game *gamelib.Game) {
    humanPlayer := game.Model.GetHumanPlayer()
    if humanPlayer != nil {
        if len(humanPlayer.Cities) > 0 {
            for _, city := range humanPlayer.Cities {
                game.Camera.Center(city.X, city.Y)
                game.Model.Plane = city.Plane
            }
        }
    }
}

func runGameInstance(game *gamelib.Game, yield coroutine.YieldFunc, magic *MagicGame, gameLoader *OriginalGameLoader) error {
    defer func() {
        // wrap the game variable so that only the remaining reference is shutdown
        game.Shutdown()
    }()
    game.GameLoader = gameLoader

    magic.Drawer = func(screen *ebiten.Image) {
        game.Draw(screen)
    }

    game.RefreshUI()

    /*
    runtime.AddCleanup(game, func(x int){
        log.Printf("Cleaned up game instance %v", x)
    }, 0)
    writeHeapDump("g1.dump")
    */

    for game.Update(yield) != gamelib.GameStateQuit {
        if inputmanager.IsQuitPressed() {
            return ebiten.Termination
        }

        select {
            case newGame := <-gameLoader.NewGame:
                game.Shutdown()
                game = newGame
                game.GameLoader = gameLoader
                game.Model.CurrentPlayer = 0

                centerOnCity(game)
                humanPlayer := game.Model.GetHumanPlayer()
                if humanPlayer != nil {
                    game.DoNextUnit(humanPlayer)
                }

                game.RefreshUI()

                magic.Drawer = func(screen *ebiten.Image) {
                    game.Draw(screen)
                }

                // writeHeapDump("g2.dump")
            default:
        }

        yield()
    }

    return nil
}

func initializeGame(magic *MagicGame, settings setup.NewGameSettings, humanWizard setup.WizardCustom) *gamelib.Game {
    game := gamelib.MakeGame(magic.Cache, magic.Music, settings)

    game.RefreshUI()

    arcanusCityArea := game.MakeCityValidArea(data.PlaneArcanus)
    myrrorCityArea := game.MakeCityValidArea(data.PlaneMyrror)

    human := initializePlayer(game, humanWizard, true, arcanusCityArea, myrrorCityArea)

    for range settings.Opponents {
        wizard, ok := game.ChooseWizard()
        if ok {
            initializePlayer(game, wizard, false, arcanusCityArea, myrrorCityArea)
        } else {
            log.Printf("Warning: unable to add another wizard to the game")
        }
    }

    log.Printf("Create neutral player")
    neutral := initializeNeutralPlayer(game, arcanusCityArea, myrrorCityArea)
    log.Printf("done create neutral player with %v cities", len(neutral.Cities))

    // hack
    // human.Admin = true
    game.Model.CurrentPlayer = 0
    game.StartPlayerTurn(human)

    // game.DoNextTurn()
    return game
}

func loadData(yield coroutine.YieldFunc, game *MagicGame, dataPath string) error {
    game.Drawer = func(screen *ebiten.Image) {
        ebitenutil.DebugPrintAt(screen, "Drag and drop a zip file that contains", 10, 10)
        ebitenutil.DebugPrintAt(screen, "the master of magic data files", 10, 30)
    }

    var cache *lbx.LbxCache
    if dataPath != "" {
        cache = lbx.CacheFromPath(dataPath)
        if cache == nil {
            return fmt.Errorf("Could not load data from '%v'", dataPath)
        }
        log.Printf("Loaded data from '%v'", dataPath)
    }

    for cache == nil {
        cache = lbx.AutoCache()
        if cache == nil {
            yield()
        }

        if inputmanager.IsQuitPressed() {
            return ebiten.Termination
        }
    }

    game.Cache = cache

    imageCache := util.MakeImageCache(cache)
    normalMouse, err := mouselib.GetMouseNormal(cache, &imageCache)
    if err == nil {
        mouse.Mouse.SetImage(normalMouse)
    }

    return nil
}

func startQuickGame(yield coroutine.YieldFunc, game *MagicGame, gameLoader *OriginalGameLoader) error {
    settings := setup.NewGameSettings{
        Opponents: rand.N(4) + 1,
        Difficulty: data.DifficultyAverage,
        Magic: data.MagicSettingNormal,
        LandSize: rand.N(3),
    }

    spells, err := spellbook.ReadSpellsFromCache(game.Cache)
    if err != nil {
        return err
    }

    wizard, ok := gamelib.ChooseUniqueWizard(nil, spells)
    if !ok {
        return fmt.Errorf("Could not choose a wizard")
    }

    log.Printf("Starting game with settings=%+v wizard=%v race=%v", settings, wizard.Name, wizard.Race)

    realGame := initializeGame(game, settings, wizard)
    return runGameInstance(realGame, yield, game, gameLoader)
}

func runGame(yield coroutine.YieldFunc, game *MagicGame, dataPath string, startGame bool, loadSave string, enableMusic bool) error {

    err := loadData(yield, game, dataPath)
    if err != nil {
        return err
    }

    game.Music = musiclib.MakeMusic(game.Cache)
    game.Music.Enabled = enableMusic
    defer game.Music.Stop()

    shutdown := func (screen *ebiten.Image){
        ebitenutil.DebugPrintAt(screen, "Shutting down", 10, 10)
    }

    gameLoader := &OriginalGameLoader{
        Cache: game.Cache,
        NewGame: make(chan *gamelib.Game, 1),
        FS: system.MakeFS(),
    }

    // start a game immediately
    if startGame {
        return startQuickGame(yield, game, gameLoader)
    }

    if loadSave == "" {
        game.Music.PlaySong(musiclib.SongIntro)
        runIntro(yield, game)

        yield()

        game.Music.PlaySong(musiclib.SongTitle)
    }

    if loadSave != "" {
        err := gameLoader.LoadNew(loadSave)
        // couldn't load game, just play title music
        if err != nil {
            game.Music.PlaySong(musiclib.SongTitle)
        }
    }

    for {
        newGame, state := runMainMenu(yield, game, gameLoader)
        switch state {
            case mainview.MainScreenStateQuit:
                game.Drawer = shutdown
                yield()
                return ebiten.Termination
            case mainview.MainScreenStateLoadGame:
                if newGame != nil {
                    game.Music.Stop()
                    // FIXME: should this go here?
                    newGame.Model.CurrentPlayer = 0
                    centerOnCity(newGame)

                    humanPlayer := newGame.Model.GetHumanPlayer()
                    if humanPlayer != nil {
                        newGame.DoNextUnit(humanPlayer)
                    }

                    err := runGameInstance(newGame, yield, game, gameLoader)
                    if err != nil {
                        game.Drawer = shutdown
                        yield()
                        return err
                    }

                    game.Music.PlaySong(musiclib.SongTitle)
                }
            case mainview.MainScreenStateQuickGame:
                game.Music.Stop()
                yield()
                err := startQuickGame(yield, game, gameLoader)
                if err != nil {
                    game.Drawer = shutdown
                    yield()
                    return err
                }
                game.Music.PlaySong(musiclib.SongTitle)
            case mainview.MainScreenStateNewGame:
                var settings setup.NewGameSettings
                var wizard setup.WizardCustom
                restart := true
                cancel := false
                for restart && !cancel {
                    // yield so that clicks from the menu don't bleed into the next part
                    yield()
                    cancel, settings = runNewGame(yield, game)
                    if cancel {
                        break
                    }
                    yield()
                    restart, wizard = runNewWizard(yield, game)
                }
                yield()
                if cancel {
                    break
                }

                game.Music.Stop()

                realGame := initializeGame(game, settings, wizard)
                err := runGameInstance(realGame, yield, game, gameLoader)

                if err != nil {
                    game.Drawer = shutdown
                    yield()
                    return err
                }

                game.Music.PlaySong(musiclib.SongTitle)
        }
    }
}

func NewMagicGame(dataPath string, startGame bool, loadSave string, enableMusic bool) (*MagicGame, error) {
    var game *MagicGame

    run := func(yield coroutine.YieldFunc) error {
        return runGame(yield, game, dataPath, startGame, loadSave, enableMusic)
    }

    game = &MagicGame{
        MainCoroutine: coroutine.MakeCoroutine(run),
        Drawer: nil,
    }

    return game, nil
}

func (game *MagicGame) Update() error {
    inputmanager.Update()

    if ebiten.IsWindowBeingClosed() {
        game.MainCoroutine.Stop()
    }

    err := game.MainCoroutine.Run()
    if err != nil {
        if errors.Is(err, coroutine.CoroutineFinished) || errors.Is(err, coroutine.CoroutineCancelled) {
            return ebiten.Termination
        }

        return err
    }

    return nil
}

func (game *MagicGame) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return scale.Scale2(data.ScreenWidth, data.ScreenHeight)
}

func (game *MagicGame) Draw(screen *ebiten.Image) {
    // screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})

    if game.Drawer != nil {
        game.Drawer(screen)
    }

    mouse.Mouse.Draw(screen)
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    var dataPath string
    var startGame bool
    var trace bool
    var enableMusic bool
    var loadSave string
    var watchMode bool
    flag.StringVar(&dataPath, "data", "", "path to master of magic lbx data files. Give either a directory or a zip file. Data is searched for in the current directory if not given.")
    flag.BoolVar(&enableMusic, "music", true, "enable music playback")
    flag.BoolVar(&startGame, "start", false, "start the game immediately with a random wizard")
    flag.BoolVar(&trace, "trace", false, "enable profiling (pprof)")
    flag.StringVar(&loadSave, "load", "", "load a saved game from the given file and start immediately")
    flag.BoolVar(&watchMode, "watch", false, "run in watch mode, where you can watch the AI play against itself (no human players)")
    flag.Parse()

    if trace {
        go func() {
            log.Printf("Starting pprof server on localhost:8000")
            log.Println(http.ListenAndServe("localhost:8000", nil))
        }()
    }

    ebiten.SetWindowSize(data.ScreenWidth * 4, data.ScreenHeight * 4)
    ebiten.SetWindowTitle("magic")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
    ebiten.SetWindowClosingHandled(true)

    audio.Initialize()
    mouse.Initialize()

    ebiten.SetCursorMode(ebiten.CursorModeHidden)

    game, err := NewMagicGame(dataPath, startGame, loadSave, enableMusic)

    if err != nil {
        log.Printf("Error: unable to load game: %v", err)
        return
    }

    err = ebiten.RunGame(game)
    if err != nil {
        log.Printf("Error: %v", err)
    }

    log.Printf("Bye")
}
