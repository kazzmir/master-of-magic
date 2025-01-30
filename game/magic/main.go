package main

import (
    "log"
    "fmt"
    "flag"
    "errors"
    "math"
    "math/rand/v2"
    "slices"
    "cmp"
    // "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    introlib "github.com/kazzmir/master-of-magic/game/magic/intro"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    mouselib "github.com/kazzmir/master-of-magic/lib/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/mainview"
    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

func stretchImage(screen *ebiten.Image, sprite *ebiten.Image){
    var options ebiten.DrawImageOptions
    options.GeoM.Scale(float64(data.ScreenWidth) / float64(sprite.Bounds().Dx()), float64(data.ScreenHeight) / float64(sprite.Bounds().Dy()))
    screen.DrawImage(sprite, &options)
}

type DrawFunc func(*ebiten.Image)

type MagicGame struct {
    Cache *lbx.LbxCache

    MainCoroutine *coroutine.Coroutine
    Drawer DrawFunc

    NewGameScreen *setup.NewGameScreen
    NewWizardScreen *setup.NewWizardScreen

    Game *gamelib.Game
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
        yield()

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

func runMainMenu(yield coroutine.YieldFunc, game *MagicGame) mainview.MainScreenState {
    menu := mainview.MakeMainScreen(game.Cache)

    game.Drawer = func(screen *ebiten.Image) {
        menu.Draw(screen)
    }

    for menu.Update() == mainview.MainScreenStateRunning {

        if inputmanager.IsQuitPressed() {
            return mainview.MainScreenStateQuit
        }

        yield()
    }

    return menu.State
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

func initializePlayer(game *gamelib.Game, wizard setup.WizardCustom, isHuman bool) {
    startingPlane := data.PlaneArcanus
    if wizard.AbilityEnabled(setup.AbilityMyrran) {
        startingPlane = data.PlaneMyrror
    }

    player := game.AddPlayer(wizard, isHuman)

    allCities := game.AllCities()

    closestDistance := func(x, y int) int {
        distance := -1

        for _, city := range allCities {
            d := int(euclideanDistance(x, y, city.X, city.Y))
            if distance == -1 || d < distance {
                distance = d
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
    }

    var cityX int
    var cityY int
    var locations []CityLocation
    for range 10 {
        x, y, ok := game.FindValidCityLocation(startingPlane)
        if ok {
            locations = append(locations, CityLocation{X: x, Y: y, Distance: closestDistance(x, y)})
        }
    }

    slices.SortFunc(locations, func(pointA, pointB CityLocation) int {
        return cmp.Compare(pointA.Distance, pointB.Distance)
    })

    if len(locations) > 0 {
        // choose furthest point
        cityX = locations[len(locations) - 1].X
        cityY = locations[len(locations) - 1].Y
    } else {
        // couldn't find a good spot, just pick anything
        for range 100 {
            var ok bool
            cityX, cityY, ok = game.FindValidCityLocation(startingPlane)
            if ok {
                break
            }
        }
    }

    cityName := game.SuggestCityName(player.Wizard.Race)

    introCity := citylib.MakeCity(cityName, cityX, cityY, player.Wizard.Race, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.GetMap(startingPlane), game)
    introCity.Population = 4000
    introCity.Wall = false
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

    introCity.ResetCitizens(player.GetUnits(cityX, cityY, startingPlane))

    player.AddCity(introCity)

    for _, unit := range startingUnits(player.Wizard.Race) {
        player.AddUnit(units.MakeOverworldUnitFromUnit(unit, cityX, cityY, startingPlane, wizard.Banner, player.MakeExperienceInfo()))
    }

    player.LiftFog(cityX, cityY, 2, introCity.Plane)

    if isHuman {
        game.Events <- gamelib.StartingCityEvent(introCity)
        game.Camera.Center(cityX, cityY)
        game.Plane = startingPlane
    }
}

func runGameInstance(yield coroutine.YieldFunc, magic *MagicGame, settings setup.NewGameSettings, humanWizard setup.WizardCustom) error {
    game := gamelib.MakeGame(magic.Cache, settings)

    magic.Drawer = func(screen *ebiten.Image) {
        game.Draw(screen)
    }

    initializePlayer(game, humanWizard, true)

    for range settings.Opponents {
        wizard, ok := game.ChooseWizard()
        if ok {
            initializePlayer(game, wizard, false)
        } else {
            log.Printf("Warning: unable to add another wizard to the game")
        }
    }

    game.DoNextTurn()

    for game.Update(yield) != gamelib.GameStateQuit {
        if inputmanager.IsQuitPressed() {
            return ebiten.Termination
        }

        yield()
    }

    return nil
}

func loadData(yield coroutine.YieldFunc, game *MagicGame, dataPath string) error {
    game.Drawer = func(screen *ebiten.Image) {
        ebitenutil.DebugPrintAt(screen, "Drag and drop a zip file that contains", 10, 10)
        ebitenutil.DebugPrintAt(screen, "the master of magic data files", 10, 30)
    }

    if dataPath != "" {
        cache := lbx.CacheFromPath(dataPath)
        if cache == nil {
            return fmt.Errorf("Could not load data from %v", dataPath)
        }
        log.Printf("Loaded data from %v", dataPath)
        game.Cache = cache
        return nil
    }

    var cache *lbx.LbxCache
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

func runGame(yield coroutine.YieldFunc, game *MagicGame, dataPath string, startGame bool) error {

    err := loadData(yield, game, dataPath)
    if err != nil {
        return err
    }

    shutdown := func (screen *ebiten.Image){
        ebitenutil.DebugPrintAt(screen, "Shutting down", 10, 10)
    }

    // start a game immediately
    if startGame {
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

        return runGameInstance(yield, game, settings, wizard)
    }

    runIntro(yield, game)

    yield()

    for {
        state := runMainMenu(yield, game)
        switch state {
            case mainview.MainScreenStateQuit:
                game.Drawer = shutdown
                yield()
                return ebiten.Termination
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
                err := runGameInstance(yield, game, settings, wizard)

                if err != nil {
                    game.Drawer = shutdown
                    yield()
                    return err
                }
        }
    }
}

func NewMagicGame(dataPath string, startGame bool) (*MagicGame, error) {
    var game *MagicGame

    run := func(yield coroutine.YieldFunc) error {
        return runGame(yield, game, dataPath, startGame)
    }

    game = &MagicGame{
        MainCoroutine: coroutine.MakeCoroutine(run),
        Drawer: nil,
    }

    return game, nil
}

func (game *MagicGame) Update() error {
    inputmanager.Update()

    err := game.MainCoroutine.Run()
    if err != nil {
        if errors.Is(err, coroutine.CoroutineFinished) {
            return ebiten.Termination
        }

        return err
    }

    return nil
}

func (game *MagicGame) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return data.ScreenWidth, data.ScreenHeight
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
    flag.StringVar(&dataPath, "data", "", "path to master of magic lbx data files. Give either a directory or a zip file. Data is searched for in the current directory if not given.")
    flag.BoolVar(&startGame, "start", false, "start the game immediately with a random wizard")
    flag.Parse()

    ebiten.SetWindowSize(data.ScreenWidth * 2, data.ScreenHeight * 2)
    ebiten.SetWindowTitle("magic")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    audio.Initialize()
    mouse.Initialize()

    ebiten.SetCursorMode(ebiten.CursorModeHidden)

    game, err := NewMagicGame(dataPath, startGame)

    if err != nil {
        log.Printf("Error: unable to load game: %v", err)
        return
    }

    err = ebiten.RunGame(game)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
