package main

import (
    "os"
    "log"
    "strconv"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    mouselib "github.com/kazzmir/master-of-magic/lib/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/music"
    "github.com/kazzmir/master-of-magic/game/magic/console"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    // "github.com/kazzmir/master-of-magic/game/magic/terrain"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    Game *gamelib.Game
    Coroutine *coroutine.Coroutine
    Console *console.Console
}

type NodeInfo struct {
    X int
    Y int
    Node *maplib.ExtraMagicNode
}

func createScenario1(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 1")

    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerBlue,
        Race: data.RaceTroll,
        Retorts: []data.Retort{
            data.RetortAlchemy,
            data.RetortSageMaster,
        },
        Books: []data.WizardBook{
            data.WizardBook{
                Magic: data.LifeMagic,
                Count: 3,
            },
            data.WizardBook{
                Magic: data.SorceryMagic,
                Count: 8,
            },
        },
    }

    game := gamelib.MakeGame(cache, music.MakeMusic(cache), setup.NewGameSettings{})

    game.Model.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)
    player.TaxRate = fraction.Zero()

    x, y, _ := game.FindValidCityLocation(game.Model.Plane)

    /*
    x = 20
    y = 20
    */

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.Model.BuildingInfo, game.Model.CurrentMap(), game.Model, player)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 2600

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    // log.Printf("City at %v, %v", x, y)

    player.LiftFog(x, y, 30, data.PlaneArcanus)

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil, &units.NoEnchantments{}))

    for i := 0; i < 5; i++ {
        fireElemental := player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil, &units.NoEnchantments{}))
        _ = fireElemental
    }

    // player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, 30, 30, data.PlaneArcanus, wizard.Banner))

    stack := player.FindStackByUnit(drake)
    player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)

    enemy1 := game.AddPlayer(setup.WizardCustom{
        Name: "dingus",
        Banner: data.BannerRed,
    }, false)

    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, nil, &units.NoEnchantments{}))
    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, nil, &units.NoEnchantments{}))

    game.Camera.Center(stack.X(), stack.Y())

    return game
}

func NewEngine(scenario int) (*Engine, error) {
    cache := lbx.AutoCache()

    var game *gamelib.Game

    switch scenario {
        case 1: game = createScenario1(cache)
        default: game = createScenario1(cache)
    }

    game.DoNextTurn()

    run := func(yield coroutine.YieldFunc) error {
        for game.Update(yield) != gamelib.GameStateQuit {
            yield()
        }

        return ebiten.Termination
    }

    normalMouse, err := mouselib.GetMouseNormal(cache, &game.ImageCache)
    if err == nil {
        mouse.Mouse.SetImage(normalMouse)
    }

    return &Engine{
        LbxCache: cache,
        Coroutine: coroutine.MakeCoroutine(run),
        Game: game,
        Console: console.MakeConsole(game),
    }, nil
}

func (engine *Engine) ChangeScale(scaleAmount int, algorithm scale.ScaleAlgorithm) {
    /*
    data.ScreenScale = 1
    data.ScreenScaleAlgorithm = algorithm
    data.ScreenWidth = 320 * scale
    data.ScreenHeight = 200 * scale
    */
    scale.UpdateScale(float64(scaleAmount))

    log.Printf("Changing scale to %v %v", scaleAmount, algorithm)

    engine.Game.UpdateImages()
    engine.Game.RefreshUI()
}

func (engine *Engine) Update() error {

    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyEscape, ebiten.KeyCapsLock: return ebiten.Termination
            case ebiten.KeyF1: engine.ChangeScale(1, scale.ScaleAlgorithmNormal)
            case ebiten.KeyF2: engine.ChangeScale(2, scale.ScaleAlgorithmNormal)
            case ebiten.KeyF3: engine.ChangeScale(3, scale.ScaleAlgorithmNormal)
            case ebiten.KeyF4: engine.ChangeScale(4, scale.ScaleAlgorithmNormal)
            case ebiten.KeyF5: engine.ChangeScale(2, scale.ScaleAlgorithmScale)
            case ebiten.KeyF6: engine.ChangeScale(3, scale.ScaleAlgorithmScale)
            case ebiten.KeyF7: engine.ChangeScale(4, scale.ScaleAlgorithmScale)
            case ebiten.KeyF8: engine.ChangeScale(2, scale.ScaleAlgorithmXbr)
            case ebiten.KeyF9: engine.ChangeScale(3, scale.ScaleAlgorithmXbr)
            case ebiten.KeyF10: engine.ChangeScale(4, scale.ScaleAlgorithmXbr)
        }
    }

    inputmanager.Update()

    engine.Console.Update()

    select {
        case event := <-engine.Console.Events:
            _, ok := event.(*console.ConsoleQuit)
            if ok {
                return ebiten.Termination
            }
        default:
    }

    /*
    switch engine.Game.Update() {
        case gamelib.GameStateRunning:
    }
    */
    if engine.Coroutine.Run() != nil {
        return ebiten.Termination
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    engine.Game.Draw(screen)
    mouse.Mouse.Draw(screen)
    engine.Console.Draw(screen)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return scale.Scale2(data.ScreenWidth, data.ScreenHeight)
}

func main(){

    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    monitorWidth, _ := ebiten.Monitor().Size()

    size := monitorWidth / 390

    ebiten.SetWindowSize(data.ScreenWidth * size, data.ScreenHeight * size)
    ebiten.SetWindowTitle("new screen")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    scenario := 1

    if len(os.Args) > 1 {
        var err error
        scenario, err = strconv.Atoi(os.Args[1])
        if err != nil {
            log.Printf("Error choosing scenario: %v", err)
            return
        }
    }

    /*
    data.ScreenWidth = data.ScreenWidthOriginal * 3
    data.ScreenHeight = data.ScreenHeightOriginal * 3
    */

    audio.Initialize()
    mouse.Initialize()

    engine, err := NewEngine(scenario)

    if err != nil {
        log.Printf("Error: unable to load engine: %v", err)
        return
    }

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
