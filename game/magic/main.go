package main

import (
    "log"
    "fmt"
    "flag"
    "io"
    "errors"
    "math/rand/v2"
    "os"
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
    settingslib "github.com/kazzmir/master-of-magic/game/magic/settings"
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
    "github.com/kazzmir/master-of-magic/game/magic/auto"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    mouselib "github.com/kazzmir/master-of-magic/lib/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/mainview"
    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
    "github.com/hajimehoshi/ebiten/v2/inpututil"

    "gopkg.in/yaml.v3"
)

type DrawFunc func(*ebiten.Image)

type GameConfig struct {
    DataPath string `yaml:"datapath"`
    StartGame bool `yaml:"startgame"`
    Trace bool `yaml:"trace"`
    EnableMusic bool `yaml:"enablemusic"`
    LoadSave string `yaml:"loadsave"`
    WatchMode bool `yaml:"watchmode"`
    TrainMode bool `yaml:"trainmode"`
    AIMode string `yaml:"aimode"`
    Opponents int `yaml:"opponents"`
    Raiders bool `yaml:"raiders"`
    AutoMode bool `yaml:"automode"`
}

func DefaultGameConfig() GameConfig {
    // set default values that are different from the zero values of the struct
    return GameConfig{
        EnableMusic: true,
        Opponents: 4,
        Raiders: true,
    }
}

type MagicGame struct {
    Cache *lbx.LbxCache

    MainCoroutine *coroutine.Coroutine
    Drawer DrawFunc

    Settings *settingslib.Settings
    Music *musiclib.Music

    Config GameConfig
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

func runMainMenu(yield coroutine.YieldFunc, game *MagicGame, gameLoader *OriginalGameLoader, music *musiclib.Music) (*gamelib.Game, mainview.MainScreenState) {
    menu := mainview.MakeMainScreen(game.Cache, gameLoader, music, game.Settings)

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

func initializePlayer(game *gamelib.Game, wizard setup.WizardCustom, isHuman bool, arcanusCityArea gamelib.CityValidArea, myrrorCityArea gamelib.CityValidArea) *playerlib.Player {
    area := arcanusCityArea
    startingPlane := data.PlaneArcanus
    if wizard.RetortEnabled(data.RetortMyrran) {
        startingPlane = data.PlaneMyrror
        area = myrrorCityArea
    }

    player := game.Model.AddPlayer(wizard, isHuman)

    cityName := game.Model.SuggestCityName(player.Wizard.Race)

    cityX, cityY := gamelib.FindCityLocation(game.Model, startingPlane, area)
    area[image.Pt(cityX, cityY)] = false

    game.GetMap(startingPlane).SetRoad(cityX, cityY, startingPlane == data.PlaneMyrror)

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

    for _, unit := range gamelib.StartingUnits(player.Wizard.Race) {
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
            cityX, cityY := gamelib.FindCityLocation(game.Model, plane, area)

            // should every neutral town be a random race, or should they all be related?
            race := randomRace()
            cityName := game.Model.SuggestCityName(race)
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
    Music *musiclib.Music
    Settings *settingslib.Settings
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

    newGame := gamelib.MakeGameFromSerialized(loader.Cache, loader.Music, loader.Settings, &serializedGame)
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

    newGame := saved.Convert(loader.Cache, loader.Music, loader.Settings)

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

func runGameInstance(game *gamelib.Game, yield coroutine.YieldFunc, magic *MagicGame, gameLoader *OriginalGameLoader, maxTurns uint64) error {
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

        if maxTurns > 0 && game.Model.TurnNumber >= maxTurns {
            log.Printf("Max turns reached: %v", maxTurns)
            return nil
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
                    game.Model.DoNextUnit(humanPlayer)
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
    game := gamelib.MakeGame(magic.Cache, magic.Music, magic.Settings, settings)

    switch magic.Config.AIMode {
        case "", "enemy2": game.Model.AIMode = ai.AIEnemy2
        case "net": game.Model.AIMode = ai.AINet
        default:
            log.Printf("Warning: unknown ai mode '%v', using default", magic.Config.AIMode)
    }

    game.RefreshUI()

    arcanusCityArea := game.Model.MakeCityValidArea(data.PlaneArcanus)
    myrrorCityArea := game.Model.MakeCityValidArea(data.PlaneMyrror)

    human := initializePlayer(game, humanWizard, true, arcanusCityArea, myrrorCityArea)

    for range settings.Opponents {
        wizard, ok := game.Model.ChooseWizard()
        if ok {
            initializePlayer(game, wizard, false, arcanusCityArea, myrrorCityArea)
        } else {
            log.Printf("Warning: unable to add another wizard to the game")
        }
    }

    if magic.Config.Raiders {
        log.Printf("Create neutral player")
        neutral := initializeNeutralPlayer(game, arcanusCityArea, myrrorCityArea)
        log.Printf("done create neutral player with %v cities", len(neutral.Cities))
    }

    // hack
    // human.Admin = true
    game.Model.CurrentPlayer = 0
    game.Model.StartPlayerTurn(human)

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

func startAITrainMode(yield coroutine.YieldFunc, game *MagicGame) error {
    game.Config.AIMode = "net"

    trainingRuns := 5

    for range trainingRuns {

        settings := setup.NewGameSettings{
            // Opponents: rand.N(4) + 1,
            Opponents: game.Config.Opponents,
            Difficulty: data.DifficultyAverage,
            Magic: data.MagicSettingNormal,
            LandSize: rand.N(3),
            DisableRaiders: !game.Config.Raiders,
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

        var netAI *ai.EnemyNetAI

        if len(realGame.Model.Players) >= 2 {
            aiPlayer := realGame.Model.Players[1]
            if aiPlayer.AIBehavior != nil {
                var ok bool
                netAI, ok = aiPlayer.AIBehavior.(*ai.EnemyNetAI)
                if !ok {
                    return fmt.Errorf("AI player is not a net AI, cannot train")
                }
            }
        }

        if netAI == nil {
            return fmt.Errorf("No AI player found, cannot train")
        }

        (func (){
            weights, err := os.Create("ai.json")
            if err == nil {
                defer weights.Close()
                err = netAI.LoadNeuralNet(weights)
                if err != nil {
                    log.Printf("Unable to load neural net: %v", err)
                }
            }
        })()

        realGame.SetWatchMode(500)

        human := realGame.Model.GetHumanPlayer()
        if human != nil {
            // make the human player an AI
            human.Admin = true
            human.Banished = true

            for _, city := range human.GetCities() {
                human.RemoveCity(city)
            }

            human.Stacks = nil
            human.SelectedStack = nil
            human.Skip = true

            // make sure all fog is visible
            human.UpdateFogVisibility()

            // consume initial events
            for range 10 {
                select {
                case <-realGame.Events:
                default:
                }
            }
        }

        // focus on the ai player
        realGame.FocusPlayerCity(1)

        // FIXME: we shouldn't need this
        gameLoader := &OriginalGameLoader{
            Cache: game.Cache,
            NewGame: make(chan *gamelib.Game, 1),
            FS: system.MakeFS(),
            Music: game.Music,
            Settings: game.Settings,
        }

        err = runGameInstance(realGame, yield, game, gameLoader, 800)
        if err != nil {
            return err
        }

        netAI.ApplyTraining()
        steps := netAI.Steps
        steps = steps[len(steps) - 10:]
        var rewards []float64
        for _, step := range steps {
            rewards = append(rewards, step.Reward)
        }
        log.Printf("AI last 10 step reward: %v", rewards)

        // log.Printf("Layer 0 weights: %v", netAI.NeuralNet.Layers[0])

        (func (){
            output, err := os.Create("ai.json")
            if err == nil {
                defer output.Close()
                err = netAI.SaveNeuralNet(output)
                if err != nil {
                    log.Printf("Unable to save neural net: %v", err)
                }
            } else {
                log.Printf("Unable to create ai.json: %v", err)
            }
        })()
    }

    return nil
}

// run a game with only AI players
func startWatchMode(yield coroutine.YieldFunc, game *MagicGame) error {
    settings := setup.NewGameSettings{
        // Opponents: rand.N(4) + 1,
        Opponents: game.Config.Opponents,
        Difficulty: data.DifficultyAverage,
        Magic: data.MagicSettingNormal,
        LandSize: rand.N(3),
        DisableRaiders: !game.Config.Raiders,
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

    realGame.SetWatchMode(100)

    human := realGame.Model.GetHumanPlayer()
    if human != nil {
        // make the human player an AI
        human.Admin = true
        human.Banished = true

        for _, city := range human.GetCities() {
            human.RemoveCity(city)
        }

        human.Stacks = nil
        human.SelectedStack = nil
        human.Skip = true

        // make sure all fog is visible
        human.UpdateFogVisibility()

        // consume initial events
        for range 10 {
            select {
                case <-realGame.Events:
                default:
            }
        }
    }

    // FIXME: we shouldn't need this
    gameLoader := &OriginalGameLoader{
        Cache: game.Cache,
        NewGame: make(chan *gamelib.Game, 1),
        FS: system.MakeFS(),
        Music: game.Music,
        Settings: game.Settings,
    }

    return runGameInstance(realGame, yield, game, gameLoader, 0)
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
    return runGameInstance(realGame, yield, game, gameLoader, 0)
}

func runGame(yield coroutine.YieldFunc, game *MagicGame, config GameConfig) error {

    err := loadData(yield, game, config.DataPath)
    if err != nil {
        return err
    }

    game.Music = musiclib.MakeMusic(game.Cache)

    game.Settings = settingslib.MakeSettings(game.Cache)
    game.Music.Enabled = config.EnableMusic
    defer game.Music.Stop()

    shutdown := func (screen *ebiten.Image){
        ebitenutil.DebugPrintAt(screen, "Shutting down", 10, 10)
    }

    if config.WatchMode {
        return startWatchMode(yield, game)
    }

    if config.TrainMode {
        return startAITrainMode(yield, game)
    }

    gameLoader := &OriginalGameLoader{
        Cache: game.Cache,
        NewGame: make(chan *gamelib.Game, 1),
        FS: system.MakeFS(),
        Music: game.Music,
        Settings: game.Settings,
    }

    // start a game immediately
    if config.StartGame {
        return startQuickGame(yield, game, gameLoader)
    }

    if config.LoadSave == "" {
        game.Music.PlaySong(musiclib.SongIntro)
        runIntro(yield, game)

        yield()

        game.Music.PlaySong(musiclib.SongTitle)
    }

    if config.LoadSave != "" {
        err := gameLoader.LoadNew(config.LoadSave)
        // couldn't load game, just play title music
        if err != nil {
            game.Music.PlaySong(musiclib.SongTitle)
        }
    }

    for {
        newGame, state := runMainMenu(yield, game, gameLoader, game.Music)
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
                        newGame.Model.DoNextUnit(humanPlayer)
                    }

                    err := runGameInstance(newGame, yield, game, gameLoader, 0)
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
                err := runGameInstance(realGame, yield, game, gameLoader, 0)

                if err != nil {
                    game.Drawer = shutdown
                    yield()
                    return err
                }

                game.Music.PlaySong(musiclib.SongTitle)
        }
    }
}

func NewMagicGame(config GameConfig) (*MagicGame, error) {
    var game *MagicGame

    run := func(yield coroutine.YieldFunc) error {
        return runGame(yield, game, config)
    }

    game = &MagicGame{
        MainCoroutine: coroutine.MakeCoroutine(run),
        Drawer: nil,
        Config: config,
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

func runAutoMode(config GameConfig) error {
    return auto.RunAutoGame()
}

func loadGameConfigFromFile(path string) GameConfig {
    // try to load as yaml file, if not then try to load as ini file (x=y format)

    out := DefaultGameConfig()

    file, err := os.Open(path)
    if err != nil {
        log.Printf("Warning: unable to open config file '%v': %v", path, err)
        return out
    }

    defer file.Close()

    decoder := yaml.NewDecoder(file)

    err = decoder.Decode(&out)
    if err != nil {
        log.Printf("Warning: unable to decode config file '%v': %v", path, err)
        return out
    }

    return out
}

// load cli arguments and return a GameConfig struct
// if -config "path" is given then load the config from the given file
// but cli arguments override the config file settings
func loadGameConfig() GameConfig {
    var dataPath string
    var startGame bool
    var trace bool
    var enableMusic bool
    var loadSave string
    var watchMode bool
    var trainMode bool
    var aiMode string
    var autoMode bool
    var config string

    flag.StringVar(&dataPath, "data", "", "path to master of magic lbx data files. Give either a directory or a zip file. Data is searched for in the current directory if not given.")
    flag.BoolVar(&enableMusic, "music", true, "enable music playback")
    flag.BoolVar(&startGame, "start", false, "start the game immediately with a random wizard")
    flag.BoolVar(&trace, "trace", false, "enable profiling (pprof)")
    flag.StringVar(&loadSave, "load", "", "load a saved game from the given file and start immediately")
    flag.BoolVar(&watchMode, "watch", false, "run in watch mode, where you can watch the AI play against itself (no human players)")
    flag.BoolVar(&trainMode, "train", false, "train the AI in watch mode. model weights are saved as ai.json")
    flag.StringVar(&aiMode, "ai", "", "select ai mode. 'default', 'enemy2', 'net'")
    flag.StringVar(&config, "config", "", "path to config file (yaml or ini)")
    flag.BoolVar(&autoMode, "auto", false, "run headless ai game")
    flag.Parse()

    out := DefaultGameConfig()

    if config != "" {
        out = loadGameConfigFromFile(config)
    }

    flag.Visit(func (f *flag.Flag) {
        switch {
            case f.Name == "data" && dataPath != "": out.DataPath = dataPath
            case f.Name == "music": out.EnableMusic = enableMusic
            case f.Name == "start": out.StartGame = startGame
            case f.Name == "trace": out.Trace = trace
            case f.Name == "load": out.LoadSave = loadSave
            case f.Name == "watch": out.WatchMode = watchMode
            case f.Name == "ai" && aiMode != "": out.AIMode = aiMode
            case f.Name == "train": out.TrainMode = trainMode
            case f.Name == "auto": out.AutoMode = autoMode
        }
    })

    return out
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    config := loadGameConfig()

    if config.Trace {
        go func() {
            log.Printf("Starting pprof server on localhost:8000")
            log.Println(http.ListenAndServe("localhost:8000", nil))
        }()
    }

    if config.AutoMode {
        runAutoMode(config)
        return
    }

    ebiten.SetWindowSize(data.ScreenWidth * 4, data.ScreenHeight * 4)
    ebiten.SetWindowTitle("magic")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
    ebiten.SetWindowClosingHandled(true)

    audio.Initialize()
    mouse.Initialize()

    ebiten.SetCursorMode(ebiten.CursorModeHidden)

    game, err := NewMagicGame(config)

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
