package main

import (
    "log"
    "fmt"
    "flag"
    "errors"
    // "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    introlib "github.com/kazzmir/master-of-magic/game/magic/intro"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
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

        if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
            return
        }
    }
}

func runNewGame(yield coroutine.YieldFunc, game *MagicGame) setup.NewGameSettings {
    newGame := setup.MakeNewGameScreen(game.Cache)

    game.Drawer = func(screen *ebiten.Image) {
        newGame.Draw(screen)
    }

    for newGame.Update() == setup.NewGameStateRunning {
        yield()
    }

    return newGame.Settings
}

func runNewWizard(yield coroutine.YieldFunc, game *MagicGame) setup.WizardCustom {
    newWizard := setup.MakeNewWizardScreen(game.Cache)

    game.Drawer = func(screen *ebiten.Image) {
        newWizard.Draw(screen)
    }

    for newWizard.Update() != setup.NewWizardScreenStateFinished {
        yield()
    }

    return newWizard.CustomWizard
}

func runMainMenu(yield coroutine.YieldFunc, game *MagicGame) mainview.MainScreenState {
    menu := mainview.MakeMainScreen(game.Cache)

    game.Drawer = func(screen *ebiten.Image) {
        menu.Draw(screen)
    }

    for menu.Update() == mainview.MainScreenStateRunning {

        if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyCapsLock) {
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

func runGameInstance(yield coroutine.YieldFunc, magic *MagicGame, settings setup.NewGameSettings, wizard setup.WizardCustom) error {
    game := gamelib.MakeGame(magic.Cache, settings)
    game.Plane = data.PlaneArcanus

    magic.Drawer = func(screen *ebiten.Image) {
        game.Draw(screen)
    }

    player := game.AddPlayer(wizard, true)

    cityX, cityY := game.FindValidCityLocation()

    introCity := citylib.MakeCity("City1", cityX, cityY, player.Wizard.Race, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, player)
    introCity.Population = 4000
    introCity.Wall = false
    introCity.Plane = data.PlaneArcanus
    introCity.Buildings.Insert(buildinglib.BuildingSmithy)
    introCity.Buildings.Insert(buildinglib.BuildingBarracks)
    introCity.Buildings.Insert(buildinglib.BuildingBuildersHall)
    introCity.Buildings.Insert(buildinglib.BuildingFortress)
    introCity.ProducingBuilding = buildinglib.BuildingHousing
    introCity.ProducingUnit = units.UnitNone
    introCity.Farmers = 4

    introCity.ResetCitizens(player.GetUnits(cityX, cityY))

    player.AddCity(introCity)

    for _, unit := range startingUnits(player.Wizard.Race) {
        player.AddUnit(units.MakeOverworldUnitFromUnit(unit, cityX, cityY, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    }

    player.LiftFog(cityX, cityY, 3)

    game.Events <- gamelib.StartingCityEvent(introCity)

    game.CenterCamera(cityX, cityY)

    game.DoNextTurn()

    for game.Update(yield) != gamelib.GameStateQuit {
        if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyCapsLock) {
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

        if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyCapsLock) {
            return ebiten.Termination
        }
    }

    game.Cache = cache

    return nil
}

func runGame(yield coroutine.YieldFunc, game *MagicGame, dataPath string) error {

    err := loadData(yield, game, dataPath)
    if err != nil {
        return err
    }

    shutdown := func (screen *ebiten.Image){
        ebitenutil.DebugPrintAt(screen, "Shutting down", 10, 10)
    }

    runIntro(yield, game)

    for {
        state := runMainMenu(yield, game)
        switch state {
            case mainview.MainScreenStateQuit:
                game.Drawer = shutdown
                yield()
                return ebiten.Termination
            case mainview.MainScreenStateNewGame:
                // yield so that clicks from the menu don't bleed into the next part
                yield()
                settings := runNewGame(yield, game)
                yield()
                wizard := runNewWizard(yield, game)
                yield()
                err := runGameInstance(yield, game, settings, wizard)

                if err != nil {
                    game.Drawer = shutdown
                    yield()
                    return err
                }
        }
    }
}

func NewMagicGame(dataPath string) (*MagicGame, error) {
    var game *MagicGame

    run := func(yield coroutine.YieldFunc) error {
        return runGame(yield, game, dataPath)
    }

    game = &MagicGame{
        MainCoroutine: coroutine.MakeCoroutine(run),
        Drawer: nil,
    }

    return game, nil
}

func (game *MagicGame) Update() error {

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
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    var dataPath string
    flag.StringVar(&dataPath, "data", "", "path to master of magic lbx data files. Give either a directory or a zip file. Data is searched for in the current directory if not given.")
    flag.Parse()

    ebiten.SetWindowSize(data.ScreenWidth * 5, data.ScreenHeight * 5)
    ebiten.SetWindowTitle("magic")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    audio.Initialize()

    game, err := NewMagicGame(dataPath)
    
    if err != nil {
        log.Printf("Error: unable to load game: %v", err)
        return
    }

    err = ebiten.RunGame(game)
    if err != nil {
        log.Printf("Error: %v", err)
    }

}
