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
}

func createScenario1(cache *lbx.LbxCache) *gamelib.Game {
    wizard := setup.WizardCustom{
        Name: "player",
        Banner: data.BannerBlue,
        Race: data.RaceTroll,
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

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.TaxRate, game.BuildingInfo)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3
    city.Wall = false

    city.ResetCitizens(nil)

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3)

    drake := player.AddUnit(units.OverworldUnit{
        Unit: units.GreatDrake,
        Plane: data.PlaneArcanus,
        Banner: wizard.Banner,
        X: x + 1,
        Y: y + 1,
    })

    for i := 0; i < 5; i++ {
        fireElemental := player.AddUnit(units.OverworldUnit{
            Unit: units.FireElemental,
            Plane: data.PlaneArcanus,
            Banner: wizard.Banner,
            X: x + 1,
            Y: y + 1,
        })
        _ = fireElemental
    }

    stack := player.FindStackByUnit(drake)
    player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 2)

    enemy1 := game.AddPlayer(setup.WizardCustom{
        Name: "dingus",
        Banner: data.BannerRed,
    })

    enemy1.AddUnit(units.OverworldUnit{
        Unit: units.Warlocks,
        Plane: data.PlaneArcanus,
        Banner: enemy1.Wizard.Banner,
        X: x + 2,
        Y: y + 2,
    })

    enemy1.AddUnit(units.OverworldUnit{
        Unit: units.HighMenBowmen,
        Plane: data.PlaneArcanus,
        Banner: enemy1.Wizard.Banner,
        X: x + 2,
        Y: y + 2,
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

    introCity := citylib.MakeCity("Test City", 4, 5, data.RaceHighElf, player.TaxRate, game.BuildingInfo)
    introCity.Population = 6000
    introCity.Plane = data.PlaneArcanus
    introCity.ProducingBuilding = buildinglib.BuildingHousing
    introCity.ProducingUnit = units.UnitNone
    introCity.Wall = false

    player.AddCity(introCity)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(4, 5, 3)

    drake := player.AddUnit(units.OverworldUnit{
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

    introCity := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.TaxRate, game.BuildingInfo)
    introCity.Population = 6000
    introCity.Plane = data.PlaneArcanus
    introCity.ProducingBuilding = buildinglib.BuildingHousing
    introCity.ProducingUnit = units.UnitNone
    introCity.Wall = false

    introCity.AddBuilding(buildinglib.BuildingShrine)

    introCity.ResetCitizens(nil)

    player.AddCity(introCity)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3)

    player.AddUnit(units.OverworldUnit{
        Unit: units.HighMenBowmen,
        Plane: data.PlaneArcanus,
        Banner: wizard.Banner,
        X: x+1,
        Y: y,
    })

    settlers := player.AddUnit(units.OverworldUnit{
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

// put starting city on a valid map tile
func createScenario4(cache *lbx.LbxCache) *gamelib.Game {
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

    introCity := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.TaxRate, game.BuildingInfo)
    introCity.Population = 6000
    introCity.Plane = data.PlaneArcanus
    introCity.ProducingBuilding = buildinglib.BuildingHousing
    introCity.ProducingUnit = units.UnitNone
    introCity.Wall = false

    introCity.AddBuilding(buildinglib.BuildingShrine)

    introCity.ResetCitizens(nil)

    player.AddCity(introCity)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3)

    player.AddUnit(units.OverworldUnit{
        Unit: units.HighMenBowmen,
        Plane: data.PlaneArcanus,
        Banner: wizard.Banner,
        X: x+1,
        Y: y,
    })

    settlers := player.AddUnit(units.OverworldUnit{
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

    game.Events <- &gamelib.GameEventNewBuilding{
        City: introCity,
        Building: buildinglib.BuildingSmithy,
    }

    return game
}

func createScenario5(cache *lbx.LbxCache) *gamelib.Game {
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

    introCity := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.TaxRate, game.BuildingInfo)
    introCity.Population = 6000
    introCity.Plane = data.PlaneArcanus
    introCity.ProducingBuilding = buildinglib.BuildingHousing
    introCity.ProducingUnit = units.UnitNone
    introCity.Wall = false

    introCity.AddBuilding(buildinglib.BuildingShrine)

    introCity.ResetCitizens(nil)

    player.AddCity(introCity)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3)

    _ = introCity

    game.CenterCamera(x, y)

    game.Events <- &gamelib.GameEventScroll{
        Title: "CITY GROWTH",
        Text: "New Haven has grown to a population of 8",
        // Text: "this is a really long piece of text that has something to do with city growth. what will it look like on the screen? lets keep going until this runs out of space",
    }

    return game
}

func NewEngine(scenario int) (*Engine, error) {
    cache := lbx.AutoCache()

    var game *gamelib.Game

    switch scenario {
        case 1: game = createScenario1(cache)
        case 2: game = createScenario2(cache)
        case 3: game = createScenario3(cache)
        case 4: game = createScenario4(cache)
        case 5: game = createScenario5(cache)
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
