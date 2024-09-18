package main

import (
    "os"
    "log"
    "strconv"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    // "github.com/kazzmir/master-of-magic/game/magic/terrain"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    Game *gamelib.Game
    Coroutine *coroutine.Coroutine
}

func createScenario1(cache *lbx.LbxCache) *gamelib.Game {
    wizard := setup.WizardCustom{
        Name: "player",
        Banner: data.BannerBlue,
        Race: data.RaceHighMen,
        Abilities: []setup.WizardAbility{
            setup.AbilityAlchemy,
            setup.AbilitySageMaster,
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

    game := gamelib.MakeGame(cache)

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard)

    city := citylib.MakeCity("Test City", 4, 5, data.RaceHighElf)
    city.Population = 6000
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = citylib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3
    city.Wall = false

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(4, 5, 3)

    drake := player.AddUnit(playerlib.Unit{
        Unit: units.GreatDrake,
        Plane: data.PlaneArcanus,
        Banner: wizard.Banner,
        X: 5,
        Y: 5,
    })

    for i := 0; i < 5; i++ {
        fireElemental := player.AddUnit(playerlib.Unit{
            Unit: units.FireElemental,
            Plane: data.PlaneArcanus,
            Banner: wizard.Banner,
            X: 5,
            Y: 5,
        })
        _ = fireElemental
    }

    stack := player.FindStackByUnit(drake)
    player.SetSelectedStack(stack)

    player.LiftFog(5, 5, 2)

    enemy1 := game.AddPlayer(setup.WizardCustom{
        Name: "dingus",
        Banner: data.BannerRed,
    })

    enemy1.AddUnit(playerlib.Unit{
        Unit: units.Warlocks,
        Plane: data.PlaneArcanus,
        Banner: enemy1.Wizard.Banner,
        X: 6,
        Y: 6,
    })

    enemy1.AddUnit(playerlib.Unit{
        Unit: units.HighMenBowmen,
        Plane: data.PlaneArcanus,
        Banner: enemy1.Wizard.Banner,
        X: 6,
        Y: 6,
    })

    return game
}

// test the starting city name input
func createScenario2(cache *lbx.LbxCache) *gamelib.Game {
    wizard := setup.WizardCustom{
        Name: "player",
        Banner: data.BannerBlue,
        Abilities: []setup.WizardAbility{
            setup.AbilityAlchemy,
            setup.AbilitySageMaster,
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

    game := gamelib.MakeGame(cache)

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard)

    introCity := citylib.MakeCity("Test City", 4, 5, data.RaceHighElf)
    introCity.Population = 6000
    introCity.Plane = data.PlaneArcanus
    introCity.ProducingBuilding = citylib.BuildingHousing
    introCity.ProducingUnit = units.UnitNone
    introCity.Wall = false

    player.AddCity(introCity)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(4, 5, 3)

    drake := player.AddUnit(playerlib.Unit{
        Unit: units.GreatDrake,
        Plane: data.PlaneArcanus,
        Banner: wizard.Banner,
        X: 5,
        Y: 5,
    })

    stack := player.FindStackByUnit(drake)
    player.SetSelectedStack(stack)

    game.Events <- gamelib.StartingCityEvent(introCity)

    player.LiftFog(5, 5, 2)

    return game
}

// put starting city on a valid map tile
func createScenario3(cache *lbx.LbxCache) *gamelib.Game {
    wizard := setup.WizardCustom{
        Name: "player",
        Banner: data.BannerBlue,
        Race: data.RaceHighMen,
        Abilities: []setup.WizardAbility{
            setup.AbilityAlchemy,
            setup.AbilitySageMaster,
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

    game := gamelib.MakeGame(cache)

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard)

    x, y := game.FindValidCityLocation()

    introCity := citylib.MakeCity("Test City", x, y, data.RaceHighElf)
    introCity.Population = 6000
    introCity.Plane = data.PlaneArcanus
    introCity.ProducingBuilding = citylib.BuildingHousing
    introCity.ProducingUnit = units.UnitNone
    introCity.Wall = false

    player.AddCity(introCity)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3)

    player.AddUnit(playerlib.Unit{
        Unit: units.HighMenBowmen,
        Plane: data.PlaneArcanus,
        Banner: wizard.Banner,
        X: x+1,
        Y: y,
    })

    settlers := player.AddUnit(playerlib.Unit{
        Unit: units.HighMenSettlers,
        Plane: data.PlaneArcanus,
        Banner: wizard.Banner,
        X: x+1,
        Y: y,
    })

    stack := player.FindStackByUnit(settlers)
    player.SetSelectedStack(stack)

    _ = introCity
    // game.Events <- gamelib.StartingCityEvent(introCity)

    player.LiftFog(x, y, 2)

    game.CenterCamera(x, y)

    return game
}

func NewEngine(scenario int) (*Engine, error) {
    cache := lbx.AutoCache()

    var game *gamelib.Game

    switch scenario {
        case 1: game = createScenario1(cache)
        case 2: game = createScenario2(cache)
        case 3: game = createScenario3(cache)
        default: game = createScenario1(cache)
    }

    game.DoNextTurn()

    run := func(yield coroutine.YieldFunc) error {
        for game.Update(yield) != gamelib.GameStateQuit {
            yield()
        }

        return ebiten.Termination
    }

    return &Engine{
        LbxCache: cache,
        Coroutine: coroutine.MakeCoroutine(run),
        Game: game,
    }, nil
}

func (engine *Engine) Update() error {

    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
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
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return data.ScreenWidth, data.ScreenHeight
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

    engine, err := NewEngine(scenario)

    audio.Initialize()

    if err != nil {
        log.Printf("Error: unable to load engine: %v", err)
        return
    }

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }

}
