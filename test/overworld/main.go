package main

import (
    "os"
    "log"
    "fmt"
    "strconv"
    "math/rand"
    "runtime/pprof"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    mouselib "github.com/kazzmir/master-of-magic/lib/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/console"
    "github.com/kazzmir/master-of-magic/game/magic/ai"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
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

func findNodes(mapObject *maplib.Map) map[terrain.TerrainType][]NodeInfo {
    out := make(map[terrain.TerrainType][]NodeInfo)
    for x := 0; x < mapObject.Height(); x++ {
        for y := 0; y < mapObject.Width(); y++ {
            if mapObject.GetTile(x, y).Tile.IsMagic() {
                type_ := mapObject.GetTile(x, y).Tile.TerrainType()
                out[type_] = append(out[type_], NodeInfo{x, y, mapObject.GetMagicNode(x, y)})
            }
        }
    }
    return out
}

func createScenario1(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 1")

    wizard := setup.WizardCustom{
        Name: "bob",
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    /*
    x = 20
    y = 20
    */

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, fraction.Zero(), game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3
    city.Wall = false

    city.ResetCitizens(nil)

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 2600

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    // log.Printf("City at %v, %v", x, y)

    player.LiftFog(x, y, 30, data.PlaneArcanus)

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))

    for i := 0; i < 5; i++ {
        fireElemental := player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))
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

    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, nil))
    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, nil))

    game.Camera.Center(stack.X(), stack.Y())

    return game
}

// test the starting city name input
func createScenario2(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 2")
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    introCity := citylib.MakeCity("Test City", 4, 5, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    introCity.Population = 6000
    introCity.Plane = data.PlaneArcanus
    introCity.ProducingBuilding = buildinglib.BuildingHousing
    introCity.ProducingUnit = units.UnitNone
    introCity.Wall = false

    player.AddCity(introCity)

    player.Gold = 83
    player.Mana = 2600

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(4, 5, 3, data.PlaneArcanus)

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, 5, 5, data.PlaneArcanus, wizard.Banner, nil))

    stack := player.FindStackByUnit(drake)
    player.SetSelectedStack(stack)

    game.Events <- gamelib.StartingCityEvent(introCity)

    player.LiftFog(5, 5, 2, data.PlaneArcanus)

    return game
}

// put starting city on a valid map tile
func createScenario3(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 3")
    wizard := setup.WizardCustom{
        Name: "player",
        Banner: data.BannerRed,
        Race: data.RaceLizard,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    introCity := citylib.MakeCity("Test City", x, y, player.Wizard.Race, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    introCity.Population = 9000
    introCity.Plane = data.PlaneArcanus
    introCity.ProducingBuilding = buildinglib.BuildingHousing
    introCity.ProducingUnit = units.UnitNone
    introCity.Farmers = 9
    introCity.Wall = false

    introCity.AddBuilding(buildinglib.BuildingShrine)
    introCity.AddBuilding(buildinglib.BuildingGranary)

    introCity.ResetCitizens(nil)

    player.AddCity(introCity)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x+1, y, data.PlaneArcanus, wizard.Banner, nil))

    settlers := player.AddUnit(units.MakeOverworldUnitFromUnit(units.LizardSettlers, x+1, y, data.PlaneArcanus, wizard.Banner, nil))

    stack := player.FindStackByUnit(settlers)
    player.SetSelectedStack(stack)

    _ = introCity
    // game.Events <- gamelib.StartingCityEvent(introCity)

    // game.Events <- &gamelib.GameEventNewOutpost{City: introCity, Stack: nil}

    player.LiftFog(x, y, 2, data.PlaneArcanus)

    game.Camera.Center(x, y)

    return game
}

// show new building event
func createScenario4(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 4")
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
                Magic: data.ChaosMagic,
                Count: 8,
            },
        },
    }

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    introCity := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    introCity.Population = 10000
    introCity.Plane = data.PlaneArcanus
    introCity.ProducingBuilding = buildinglib.BuildingHousing
    introCity.ProducingUnit = units.UnitNone
    introCity.Wall = false
    introCity.Farmers = 10

    introCity.AddBuilding(buildinglib.BuildingShrine)

    introCity.ResetCitizens(nil)

    player.AddCity(introCity)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x+1, y, data.PlaneArcanus, wizard.Banner, nil))

    settlers := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSettlers, x+1, y, data.PlaneArcanus, wizard.Banner, nil))

    stack := player.FindStackByUnit(settlers)
    player.SetSelectedStack(stack)

    _ = introCity
    // game.Events <- gamelib.StartingCityEvent(introCity)

    player.LiftFog(x, y, 2, data.PlaneArcanus)

    game.Camera.Center(x, y)

    introCity.Buildings.Insert(buildinglib.BuildingSmithy)

    game.Events <- &gamelib.GameEventNewBuilding{
        City: introCity,
        Player: player,
        Building: buildinglib.BuildingSmithy,
    }

    return game
}

func createScenario5(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 5")
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    introCity := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
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

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    _ = introCity

    game.Camera.Center(x, y)

    game.Events <- &gamelib.GameEventScroll{
        Title: "CITY GROWTH",
        Text: "New Haven has grown to a population of 8",
        // Text: "this is a really long piece of text that has something to do with city growth. what will it look like on the screen? lets keep going until this runs out of space",
    }

    return game
}

func createScenario6(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 6")
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    introCity := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    introCity.Population = 14000
    introCity.Plane = data.PlaneArcanus
    introCity.ProducingBuilding = buildinglib.BuildingHousing
    introCity.ProducingUnit = units.UnitNone
    introCity.Wall = false
    introCity.Farmers = 14

    introCity.AddBuilding(buildinglib.BuildingShrine)

    introCity.ResetCitizens(nil)

    player.AddCity(introCity)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    _ = introCity

    for i := 0; i < 3; i++ {
        player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x + i, y + 1, data.PlaneArcanus, wizard.Banner, nil))
    }

    for i := 0; i < 3; i++ {
        player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 3 + i, y + 1, data.PlaneArcanus, wizard.Banner, nil))
    }

    for i := 0; i < 3; i++ {
        player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, x + 6 + i, y + 1, data.PlaneArcanus, wizard.Banner, nil))
    }

    x, y, _ = game.FindValidCityLocation(game.Plane)

    city2 := citylib.MakeCity("utah", x, y, data.RaceDarkElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    city2.Population = 7000
    city2.Plane = data.PlaneArcanus
    city2.ProducingBuilding = buildinglib.BuildingShrine
    city2.ProducingUnit = units.UnitNone
    city2.Wall = false

    city2.ResetCitizens(nil)

    player.AddCity(city2)

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    game.Camera.Center(x, y)

    return game
}

func createScenario7(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 7")
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    for i := 0; i < 15; i++ {
        x, y, _ := game.FindValidCityLocation(game.Plane)
        player.LiftFog(x, y, 3, data.PlaneArcanus)

        introCity := citylib.MakeCity(fmt.Sprintf("city%v", i), x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
        introCity.Population = rand.Intn(5000) + 5000
        introCity.Plane = data.PlaneArcanus
        introCity.ProducingBuilding = buildinglib.BuildingHousing
        introCity.ProducingUnit = units.UnitNone
        introCity.Wall = false

        introCity.AddBuilding(buildinglib.BuildingShrine)

        introCity.ResetCitizens(nil)

        player.AddCity(introCity)
    }

    for i := 0; i < 4; i++ {
        x, y, _ := game.FindValidCityLocation(data.PlaneMyrror)
        player.LiftFog(x, y, 3, data.PlaneMyrror)

        introCity := citylib.MakeCity(fmt.Sprintf("city%v myr", i), x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.GetMap(data.PlaneMyrror), game)
        introCity.Population = rand.Intn(5000) + 5000
        introCity.Plane = data.PlaneMyrror
        introCity.ProducingBuilding = buildinglib.BuildingHousing
        introCity.ProducingUnit = units.UnitNone
        introCity.Wall = false

        introCity.AddBuilding(buildinglib.BuildingShrine)

        introCity.ResetCitizens(nil)

        player.AddCity(introCity)
    }

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    return game
}

// test entering a node with a unit
func createScenario8(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 8")
    wizard := setup.WizardCustom{
        Name: "bob",
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
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
    player.Mana = 2600

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))

    nodes := findNodes(game.CurrentMap())

    node := nodes[terrain.SorceryNode][0]
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, node.X + 1, node.Y + 1, data.PlaneArcanus, wizard.Banner, nil))
    player.LiftFog(node.X, node.Y, 3, data.PlaneArcanus)

    node = nodes[terrain.ChaosNode][0]
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, node.X + 1, node.Y + 1, data.PlaneArcanus, wizard.Banner, nil))
    player.LiftFog(node.X, node.Y, 3, data.PlaneArcanus)

    node = nodes[terrain.NatureNode][0]
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, node.X + 1, node.Y + 1, data.PlaneArcanus, wizard.Banner, nil))
    player.LiftFog(node.X, node.Y, 3, data.PlaneArcanus)

    stack := player.FindStackByUnit(drake)
    player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)

    return game
}

// show summon unit animation
func createScenario9(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 9")
    wizard := setup.WizardCustom{
        Name: "bob",
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
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

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))

    for i := 0; i < 1; i++ {
        fireElemental := player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))
        _ = fireElemental
    }

    stack := player.FindStackByUnit(drake)
    player.SetSelectedStack(stack)

    game.Events <- &gamelib.GameEventSummonUnit{
        Unit: units.FireGiant,
        Player: player,
    }

    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)

    return game
}

// show summon hero animation
func createScenario10(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 10")
    wizard := setup.WizardCustom{
        Name: "bob",
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
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

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))

    for i := 0; i < 1; i++ {
        fireElemental := player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))
        _ = fireElemental
    }

    stack := player.FindStackByUnit(drake)
    player.SetSelectedStack(stack)

    game.Events <- &gamelib.GameEventSummonHero{
        Wizard: wizard.Base,
        Champion: false,
    }

    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)

    return game
}

// test meld ability
func createScenario11(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 11")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
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

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    nodes := findNodes(game.CurrentMap())
    node := nodes[terrain.SorceryNode][0]
    game.CurrentMap().RemoveEncounter(node.X, node.Y)

    player.LiftFog(node.X, node.Y, 3, data.PlaneArcanus)

    spirit := player.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, node.X + 1, node.Y + 1, data.PlaneArcanus, wizard.Banner, nil))

    stack := player.FindStackByUnit(spirit)
    player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)

    return game
}

// show map tile bonuses
func createScenario12(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 12")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
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

    player.LiftFog(x, y, 10, data.PlaneArcanus)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))

    game.CurrentMap().SetBonus(x-3, y-1, data.BonusSilverOre)
    game.CurrentMap().SetBonus(x-2, y-1, data.BonusGem)
    game.CurrentMap().SetBonus(x-1, y-1, data.BonusWildGame)
    game.CurrentMap().SetBonus(x, y-1, data.BonusQuorkCrystal)
    game.CurrentMap().SetBonus(x+1, y-1, data.BonusIronOre)

    game.Camera.Center(x, y)

    game.Events <- &gamelib.GameEventSurveyor{}

    return game
}

// overland cast
func createScenario13(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 13")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.CastingSkillPower += 50000

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player.KnownSpells.AddSpell(allSpells.FindByName("Earth Lore"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Giant Strength"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Ice Bolt"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Enchant Item"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Create Artifact"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Magic Spirit"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Wraiths"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Dark Rituals"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Wall of Fire"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Change Terrain"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Transmute"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Summon Hero"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Summon Champion"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Enchant Road"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Raise Volcano"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Corruption"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Warp Node"))

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3
    city.Wall = false

    city.AddBuilding(buildinglib.BuildingShrine)
    city.AddBuilding(buildinglib.BuildingGranary)

    city.ResetCitizens(nil)

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 5000

    player.LiftFog(x, y, 4, data.PlaneArcanus)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))

    game.CurrentMap().SetRoad(x, y+1, false)
    game.CurrentMap().SetRoad(x, y+2, false)

    return game
}

// research a new spell
func createScenario14(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 14")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
        Race: data.RaceTroll,
        Base: data.WizardAriel,
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
                Magic: data.DeathMagic,
                Count: 8,
            },
        },
    }

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.CastingSkillPower += 500

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    /*
    for _, name := range []string{"Earth Lore", "Giant Strength", "Ice Bolt", "Enchant Item", "Dark Rituals", "Spell Blast", "Time Stop", "Web", "Magic Spirit"} {
        player.ResearchPoolSpells.AddSpell(allSpells.FindByName(name))
    }
    */
    // player.ResearchPoolSpells.AddAllSpells(allSpells)

    player.KnownSpells.AddSpell(allSpells.FindByName("Earth Lore"))

    /*
    player.KnownSpells.AddSpell(allSpells.FindByName("Giant Strength"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Ice Bolt"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Enchant Item"))
    // player.Spells.AddSpell(allSpells.FindByName("Magic Spirit"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Dark Rituals"))
    */

    /*
    player.ResearchCandidateSpells.AddSpell(allSpells.FindByName("Magic Spirit"))
    player.ResearchCandidateSpells.AddSpell(allSpells.FindByName("Endurance"))
    player.ResearchCandidateSpells.AddSpell(allSpells.FindByName("Hell Hounds"))
    player.ResearchCandidateSpells.AddSpell(allSpells.FindByName("Healing"))
    player.ResearchCandidateSpells.AddSpell(allSpells.FindByName("Corruption"))
    player.ResearchCandidateSpells.AddSpell(allSpells.FindByName("Dispel Magic"))
    player.ResearchCandidateSpells.AddSpell(allSpells.FindByName("Summoning Circle"))
    player.ResearchCandidateSpells.AddSpell(allSpells.FindByName("Just Cause"))
    player.ResearchCandidateSpells.AddSpell(allSpells.FindByName("Detect Magic"))
    */

    // player.ResearchingSpell = allSpells.FindByName("Magic Spirit")
    // player.ResearchingSpell = allSpells.FindByName("Earth Lore")
    // player.ResearchProgress = 10

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
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
    player.Mana = 50

    player.LiftFog(x, y, 4, data.PlaneArcanus)
    game.Camera.Center(x, y)

    game.Events <- &gamelib.GameEventLearnedSpell{
        Player: player,
        // Spell: allSpells.FindByName("Earth Lore"),
        Spell: allSpells.FindByName("Magic Spirit"),
    }

    return game
}

// show units with low health in overland
func createScenario15(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 15")

    wizard := setup.WizardCustom{
        Name: "bob",
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 20190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.Buildings.Insert(buildinglib.BuildingGranary)
    city.Buildings.Insert(buildinglib.BuildingFarmersMarket)
    city.Buildings.Insert(buildinglib.BuildingForestersGuild)
    city.Buildings.Insert(buildinglib.BuildingShrine)
    city.Buildings.Insert(buildinglib.BuildingTemple)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 20
    city.Workers = 0
    city.Wall = false

    city.ResetCitizens(nil)

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    spear1 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, nil))
    spear2 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, nil))
    spear3 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, nil))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, nil))
    spear1.AdjustHealth(-spear1.GetHealth() / 2)
    spear2.AdjustHealth(-spear2.GetHealth() / 3)
    spear3.AdjustHealth(-1)

    stack := player.FindStackByUnit(spear1)
    player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)

    enemy1 := game.AddPlayer(setup.WizardCustom{
        Name: "dingus",
        Banner: data.BannerRed,
    }, false)

    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, nil))
    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, nil))

    return game
}

// show units with various levels of experience
func createScenario16(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 16")

    wizard := setup.WizardCustom{
        Name: "bob",
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.Fame = 40

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, fraction.Zero(), game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 10190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.Buildings.Insert(buildinglib.BuildingWizardsGuild)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = city.Citizens() - 1
    city.Workers = 0
    city.Wall = false

    city.ResetCitizens(nil)

    player.AddCity(city)

    player.Gold = 1283
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    spear1 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    spear2 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    spear3 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    spear4 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    spear5 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    spear2.AddExperience(30)
    spear3.AddExperience(60)
    spear4.AddExperience(100)
    spear5.AddExperience(200)

    stack := player.FindStackByUnit(spear1)
    player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)

    return game
}

// create a new artifact
func createScenario17(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 17")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.CastingSkillPower += 5000

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player.KnownSpells.AddSpell(allSpells.FindByName("Earth Lore"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Giant Strength"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Ice Bolt"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Enchant Item"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Create Artifact"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Magic Spirit"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Dark Rituals"))

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
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
    player.Mana = 5000

    player.LiftFog(x, y, 4, data.PlaneArcanus)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))

    player.Heroes[0] = hero.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroRakir, 1, 1, data.PlaneArcanus, wizard.Banner, nil), hero.HeroRakir, "Rakir")
    player.Heroes[1] = hero.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroTorin, 1, 1, data.PlaneArcanus, wizard.Banner, nil), hero.HeroTorin, "Torin")
    player.Heroes[2] = hero.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroWarrax, 1, 1, data.PlaneArcanus, wizard.Banner, nil), hero.HeroWarrax, "Warrax")
    player.Heroes[3] = hero.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroRavashack, 1, 1, data.PlaneArcanus, wizard.Banner, nil), hero.HeroRavashack, "Ravashack")
    player.Heroes[4] = hero.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroSirHarold, 1, 1, data.PlaneArcanus, wizard.Banner, nil), hero.HeroSirHarold, "Sir Harold")
    player.Heroes[5] = hero.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroAlorra, 1, 1, data.PlaneArcanus, wizard.Banner, nil), hero.HeroAlorra, "Alorra")

    player.VaultEquipment[0] = &artifact.Artifact{
        Name: "Baloney",
        Image: 7,
        Type: artifact.ArtifactTypeSword,
        Powers: []artifact.Power{
            {
                Type: artifact.PowerTypeAttack,
                Amount: 1,
                Name: "+1 Attack",
            },
            {
                Type: artifact.PowerTypeDefense,
                Amount: 2,
                Name: "+2 Defense",
            },
        },
        Cost: 250,
    }

    player.VaultEquipment[1] = &artifact.Artifact{
        Name: "Pizza",
        Image: 31,
        Type: artifact.ArtifactTypeBow,
        Powers: []artifact.Power{
            {
                Type: artifact.PowerTypeAttack,
                Amount: 1,
                Name: "+1 Attack",
            },
            {
                Type: artifact.PowerTypeMovement,
                Amount: 2,
                Name: "+2 Movement",
            },
        },
        Cost: 300,
    }

    player.VaultEquipment[1] = game.ArtifactPool["Pummel Mace"]

    testArtifact := artifact.Artifact{
        Name: "Sword",
        Image: 5,
        Type: artifact.ArtifactTypeSword,
        Powers: []artifact.Power{
            {
                Type: artifact.PowerTypeAttack,
                Amount: 2,
                Name: "+2 Attack",
            },
            {
                Type: artifact.PowerTypeDefense,
                Amount: 2,
                Name: "+2 Defense",
            },
            {
                Type: artifact.PowerTypeMovement,
                Amount: 3,
                Name: "+3 Movement",
            },
            {
                Type: artifact.PowerTypeAbility1,
                Amount: 2,
                Ability: data.AbilityFlaming,
                Name: "Flaming",
            },
            /*
            {
                Type: artifact.PowerTypeSpellCharges,
                Amount: 2,
                Spell: allSpells.FindByName("Ice Bolt"),
                Name: "Ice Bolt x1",
            },
            */
        },
        Cost: 1000,
    }

    game.Events <- &gamelib.GameEventVault{
        CreatedArtifact: &testArtifact,
    }

    return game
}

// show units with low health in overland
func createScenario18(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 18")

    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerBlue,
        Race: data.RaceTroll,
        Abilities: []setup.WizardAbility{
            setup.AbilityAlchemy,
            setup.AbilitySageMaster,
            setup.AbilityWarlord,
            setup.AbilityChanneler,
            setup.AbilityMyrran,
            setup.AbilityFamous,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.Fame = 12

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 8190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 5
    city.Workers = 3
    city.Wall = false

    city.ResetCitizens(nil)

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)
    player.LiftFog(20, 20, 100, data.PlaneMyrror)

    rakir := hero.MakeHero(units.MakeOverworldUnit(units.HeroRakir), hero.HeroRakir, "bubba")
    player.AddHero(rakir)
    rakir.AddExperience(528)

    mysticX := hero.MakeHero(units.MakeOverworldUnit(units.HeroMysticX), hero.HeroMysticX, "bubba")
    player.AddHero(mysticX)
    mysticX.SetExtraAbilities()
    mysticX.AddAbility(data.AbilityArmsmaster)
    mysticX.AddExperience(528)

    warlock := player.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, x, y, data.PlaneArcanus, player.GetBanner(), nil))
    // warlock.AddEnchantment(data.UnitEnchantmentGiantStrength)
    warlock.AddEnchantment(data.UnitEnchantmentLionHeart)

    stack := player.FindStackByUnit(mysticX)
    player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)

    enemy1 := game.AddPlayer(setup.WizardCustom{
        Name: "herby",
        Banner: data.BannerBlue,
    }, false)

    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, nil))
    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, nil))
    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, nil))

    return game
}

// hire hero event
func createScenario19(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 19")

    wizard := setup.WizardCustom{
        Name: "bob",
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3
    city.Wall = false

    city.ResetCitizens(nil)

    player.AddCity(city)

    player.Gold = 683
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)
    game.Camera.Center(x, y)

    enemy1 := game.AddPlayer(setup.WizardCustom{
        Name: "dingus",
        Banner: data.BannerRed,
    }, false)

    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, nil))
    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, nil))

    game.Events <- &gamelib.GameEventHireHero{
        Player: player,
        Hero: game.Heroes[hero.HeroRakir],
        Cost: 200,
    }

    return game
}

// enemy neutral town
func createScenario20(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 20")

    wizard := setup.WizardCustom{
        Name: "bob",
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.Fame = 5

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, fraction.Zero(), game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 10190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.Buildings.Insert(buildinglib.BuildingWizardsGuild)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = city.Citizens() - 1
    city.Workers = 0
    city.Wall = false

    city.ResetCitizens(nil)

    player.AddCity(city)

    player.Gold = 1283
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    spear1 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    spear2 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    spear3 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    spear4 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    spear5 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    spear2.AddExperience(30)
    spear3.AddExperience(60)
    spear4.AddExperience(100)
    spear5.AddExperience(200)

    spear2.SetWeaponBonus(data.WeaponMagic)
    spear3.SetWeaponBonus(data.WeaponMythril)
    spear4.SetWeaponBonus(data.WeaponAdamantium)

    stack := player.FindStackByUnit(spear1)
    player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)

    enemyWizard := setup.WizardCustom{
        Name: "enemy",
        Banner: data.BannerBrown,
        Race: data.RaceDraconian,
    }

    enemy := game.AddPlayer(enemyWizard, false)

    city2 := citylib.MakeCity("Test City", x, y-1, enemy.Wizard.Race, enemy.Wizard.Banner, fraction.Make(1, 1), game.BuildingInfo, game.CurrentMap(), game)
    city2.Population = 8000
    city2.Plane = data.PlaneArcanus
    city2.ProducingBuilding = buildinglib.BuildingHousing
    city2.ProducingUnit = units.UnitNone
    city2.Farmers = city2.Citizens() - 1
    city2.Workers = 1
    city2.Wall = false
    city2.Buildings.Insert(buildinglib.BuildingSmithy)
    city2.Buildings.Insert(buildinglib.BuildingOracle)

    // cant use brown banner because neutral cities will never cast a city enchantment
    city2.AddEnchantment(data.CityEnchantmentWallOfFire, data.BannerRed)

    city2.ResetCitizens(nil)

    city2.Farmers = 5
    city2.Workers = 2
    city2.Rebels = 1

    for range 8 {
        randomUnit := units.ChooseRandomUnit(enemy.Wizard.Race)
        enemy.AddUnit(units.MakeOverworldUnitFromUnit(randomUnit, city2.X, city2.Y, data.PlaneArcanus, enemyWizard.Banner, enemy.MakeExperienceInfo()))
    }

    enemy.AddCity(city2)

    return game
}

// enemy neutral town
func createScenario21(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 21")

    wizard := setup.WizardCustom{
        Name: "bob",
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{LandSize: 0})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.Fame = 5

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, fraction.Zero(), game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 10190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.Buildings.Insert(buildinglib.BuildingWizardsGuild)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = city.Citizens() - 1
    city.Workers = 0
    city.Wall = false

    city.ResetCitizens(nil)

    player.AddCity(city)

    player.Gold = 1283
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    spear1 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    spear2 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    spear3 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    spear4 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    spear5 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    spear2.AddExperience(30)
    spear3.AddExperience(60)
    spear4.AddExperience(100)
    spear5.AddExperience(200)

    stack := player.FindStackByUnit(spear1)
    player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 20, data.PlaneArcanus)

    enemyWizard := setup.WizardCustom{
        Name: "enemy",
        Banner: data.BannerBrown,
        Race: data.RaceHighMen,
    }

    enemy := game.AddPlayer(enemyWizard, false)

    enemy.AIBehavior = ai.MakeRaiderAI()

    x2, y2 := game.FindValidCityLocationOnContinent(game.Plane, city.X, city.Y)
    log.Printf("enemy city at %v, %v", x2, y2)
    city2 := citylib.MakeCity("Test City", x2, y2, enemy.Wizard.Race, enemy.Wizard.Banner, fraction.Make(1, 1), game.BuildingInfo, game.CurrentMap(), game)
    city2.Population = 8000
    city2.Plane = data.PlaneArcanus
    city2.ProducingBuilding = buildinglib.BuildingHousing
    city2.ProducingUnit = units.UnitNone
    city2.Farmers = city2.Citizens() - 1
    city2.Workers = 1
    city2.Wall = false
    city2.Buildings.Insert(buildinglib.BuildingSmithy)
    city2.Buildings.Insert(buildinglib.BuildingOracle)

    city2.ResetCitizens(nil)

    city2.Farmers = 5
    city2.Workers = 2
    city2.Rebels = 1

    enemy.LiftFog(city2.X, city2.Y, 10, data.PlaneArcanus)

    for range 8 {
        randomUnit := units.ChooseRandomUnit(enemy.Wizard.Race)
        enemy.AddUnit(units.MakeOverworldUnitFromUnit(randomUnit, city2.X, city2.Y, data.PlaneArcanus, enemyWizard.Banner, enemy.MakeExperienceInfo()))
    }

    enemy.AddCity(city2)

    return game
}

// test entering a lair with a unit
func createScenario22(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 22")
    wizard := setup.WizardCustom{
        Name: "bob",
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    // game.Map.CreateEncounter(x, y+1, maplib.EncounterTypeLair, game.Settings.Difficulty, false, game.Plane)
    game.CurrentMap().CreateEncounterRandom(x, y+1, game.Settings.Difficulty, game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
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
    player.Mana = 2600

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))

    for i := 0; i < 1; i++ {
        fireElemental := player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))
        _ = fireElemental
    }

    stack := player.FindStackByUnit(drake)
    player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)

    return game
}

// enemy ai
func createScenario23(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 23")

    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerBlue,
        Race: data.RaceTroll,
        Base: data.WizardRaven,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{LandSize: 0})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.Fame = 5

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, fraction.Zero(), game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 10190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.Buildings.Insert(buildinglib.BuildingWizardsGuild)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = city.Citizens() - 1
    city.Workers = 0
    city.Wall = false

    city.ResetCitizens(nil)

    player.AddCity(city)

    player.Gold = 1283
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    spear1 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    spear2 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    spear3 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    spear4 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    spear5 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo()))
    spear2.AddExperience(30)
    spear3.AddExperience(60)
    spear4.AddExperience(100)
    spear5.AddExperience(200)

    stack := player.FindStackByUnit(spear1)
    player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 20, data.PlaneArcanus)

    enemyWizard, _ := game.ChooseWizard()

    enemy := game.AddPlayer(enemyWizard, false)

    enemy.AwarePlayer(player)
    player.AwarePlayer(enemy)

    enemy.WarWithPlayer(player)
    player.WarWithPlayer(enemy)

    enemyWizard2, _ := game.ChooseWizard()
    enemy2 := game.AddPlayer(enemyWizard2, false)
    enemy.AwarePlayer(enemy2)
    enemy2.AwarePlayer(enemy)

    player.AwarePlayer(enemy2)
    enemy2.AwarePlayer(player)

    enemy.PactWithPlayer(enemy2)
    enemy2.PactWithPlayer(enemy)

    enemy2.AllianceWithPlayer(player)
    player.AllianceWithPlayer(enemy2)

    enemyWizard3, _ := game.ChooseWizard()
    enemy3 := game.AddPlayer(enemyWizard3, false)

    enemy3.AwarePlayer(player)
    player.AwarePlayer(enemy3)

    enemy.WarWithPlayer(enemy3)
    enemy3.WarWithPlayer(enemy)

    enemyWizard4, _ := game.ChooseWizard()
    enemy4 := game.AddPlayer(enemyWizard4, false)

    enemy4.AwarePlayer(player)
    player.AwarePlayer(enemy4)

    enemy4.PactWithPlayer(enemy)
    enemy.PactWithPlayer(enemy4)

    enemy.AIBehavior = ai.MakeRaiderAI()

    x2, y2 := game.FindValidCityLocationOnContinent(game.Plane, city.X, city.Y)
    log.Printf("enemy city at %v, %v", x2, y2)
    city2 := citylib.MakeCity("Test City", x2, y2, enemy.Wizard.Race, enemy.Wizard.Banner, fraction.Make(1, 1), game.BuildingInfo, game.CurrentMap(), game)
    city2.Population = 8000
    city2.Plane = data.PlaneArcanus
    city2.ProducingBuilding = buildinglib.BuildingHousing
    city2.ProducingUnit = units.UnitNone
    city2.Farmers = city2.Citizens() - 1
    city2.Workers = 1
    city2.Wall = false
    city2.Buildings.Insert(buildinglib.BuildingSmithy)
    city2.Buildings.Insert(buildinglib.BuildingOracle)

    city2.ResetCitizens(nil)

    city2.Farmers = 5
    city2.Workers = 2
    city2.Rebels = 1

    enemy.LiftFog(city2.X, city2.Y, 10, data.PlaneArcanus)

    /*
    for range 8 {
        randomUnit := units.ChooseRandomUnit(enemy.Wizard.Race)
        enemy.AddUnit(units.MakeOverworldUnitFromUnit(randomUnit, city2.X, city2.Y, data.PlaneArcanus, enemyWizard.Banner, enemy.MakeExperienceInfo()))
    }
    */

    enemy.AddCity(city2)

    return game
}

// hire mercenaries
func createScenario24(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 24")
    wizard := setup.WizardCustom{
        Name: "bob",
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 5
    city.Workers = 1
    city.Wall = false

    city.ResetCitizens(nil)

    player.AddCity(city)

    player.Gold = 15772
    player.Mana = 26

    player.LiftFog(x, y, 3, data.PlaneArcanus)
    game.Camera.Center(x, y)

    enemy1 := game.AddPlayer(setup.WizardCustom{
        Name: "dingus",
        Banner: data.BannerRed,
    }, false)

    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, x + 1, y + 1, data.PlaneArcanus, enemy1.Wizard.Banner, nil))
    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x + 1, y + 1, data.PlaneArcanus, enemy1.Wizard.Banner, nil))

    game.Events <- &gamelib.GameEventHireMercenaries{
        Player: player,
        Units: []*units.OverworldUnit{
            units.MakeOverworldUnitFromUnit(units.BarbarianSpearmen, x, y, data.PlaneArcanus, player.Wizard.Banner, nil),
            units.MakeOverworldUnitFromUnit(units.BarbarianSpearmen, x, y, data.PlaneArcanus, player.Wizard.Banner, nil),
        },
        Cost: 200,
    }

    return game
}

// merchant
func createScenario25(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 25")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerBlue,
        Race: data.RaceTroll,
        Abilities: []setup.WizardAbility{
            setup.AbilityAlchemy,
            setup.AbilitySageMaster,
        },
        Books: []data.WizardBook{
            {
                Magic: data.LifeMagic,
                Count: 3,
            },
            {
                Magic: data.SorceryMagic,
                Count: 8,
            },
        },
    }

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})
    game.Plane = data.PlaneArcanus

    x, y, _ := game.FindValidCityLocation(game.Plane)
    game.Camera.Center(x, y)

    player := game.AddPlayer(wizard, true)
    player.Gold = 15772
    player.Mana = 26
    player.LiftFog(x, y, 3, data.PlaneArcanus)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 5
    city.Workers = 1
    city.Wall = false
    city.ResetCitizens(nil)
    player.AddCity(city)

    player.AddHero(hero.MakeHero(units.MakeOverworldUnit(units.HeroGunther), hero.HeroGunther, "Gunther"))

    enemy := game.AddPlayer(setup.WizardCustom{
        Name: "dingus",
        Banner: data.BannerRed,
    }, false)
    enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.KlackonSpearmen, x + 1, y + 1, data.PlaneArcanus, enemy.Wizard.Banner, nil))

    artifact := game.ArtifactPool["Pummel Mace"]

    game.Events <- &gamelib.GameEventMerchant{
        Player: player,
        Artifact: artifact,
        Cost: artifact.Cost,
    }

    return game
}

// hero level up
func createScenario26(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 26")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerBlue,
        Race: data.RaceTroll,
        Abilities: []setup.WizardAbility{
            setup.AbilityAlchemy,
            setup.AbilitySageMaster,
        },
        Books: []data.WizardBook{
            {
                Magic: data.LifeMagic,
                Count: 3,
            },
            {
                Magic: data.SorceryMagic,
                Count: 8,
            },
        },
    }

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})
    game.Plane = data.PlaneArcanus

    x, y, _ := game.FindValidCityLocation(game.Plane)
    game.Camera.Center(x, y)

    player := game.AddPlayer(wizard, true)
    player.Gold = 15772
    player.Mana = 26
    player.LiftFog(x, y, 3, data.PlaneArcanus)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 5
    city.Workers = 1
    city.Wall = false
    city.ResetCitizens(nil)
    player.AddCity(city)

    gunther := hero.MakeHero(units.MakeOverworldUnit(units.HeroGunther), hero.HeroGunther, "Gunther")
    reywind := hero.MakeHero(units.MakeOverworldUnit(units.HeroReywind), hero.HeroReywind, "Reywind")
    mysticX := hero.MakeHero(units.MakeOverworldUnit(units.HeroMysticX), hero.HeroMysticX, "Mystic X")
    mysticX.SetExtraAbilities()
    player.AddHero(mysticX)
    // player.AddHero(gunther)
    // player.AddHero(reywind)
    gunther.AddExperience(19)
    reywind.AddExperience(58)
    mysticX.AddExperience(19)

    enemy := game.AddPlayer(setup.WizardCustom{
        Name: "dingus",
        Banner: data.BannerRed,
    }, false)
    enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.KlackonSpearmen, x + 1, y + 1, data.PlaneArcanus, enemy.Wizard.Banner, nil))

    return game
}

// multiple stacks on top of each other
func createScenario27(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 1")

    wizard := setup.WizardCustom{
        Name: "bob",
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    /*
    x = 20
    y = 20
    */

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, fraction.Zero(), game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 10000
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.Buildings.Insert(buildinglib.BuildingGranary)
    city.Buildings.Insert(buildinglib.BuildingFarmersMarket)
    city.Buildings.Insert(buildinglib.BuildingForestersGuild)
    city.ProducingBuilding = buildinglib.BuildingNone
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 7
    city.Workers = 3
    city.Wall = false

    city.ResetCitizens(nil)

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    // log.Printf("City at %v, %v", x, y)

    player.LiftFog(x, y, 30, data.PlaneArcanus)

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))

    for i := 0; i < 2; i++ {
        fireElemental := player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))
        _ = fireElemental
    }

    // player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, 30, 30, data.PlaneArcanus, wizard.Banner))

    stack := player.FindStackByUnit(drake)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x + 2, y + 2, data.PlaneArcanus, player.Wizard.Banner, nil))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenCavalry, x + 2, y + 2, data.PlaneArcanus, player.Wizard.Banner, nil))

    stack2 := player.FindStack(x + 2, y + 2, data.PlaneArcanus)
    stack2.Move(-1, -1, fraction.Zero(), game.GetNormalizeCoordinateFunc())

    // player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)

    enemy1 := game.AddPlayer(setup.WizardCustom{
        Name: "dingus",
        Banner: data.BannerRed,
    }, false)

    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, x + 3, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, nil))
    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x + 3, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, nil))

    game.Camera.Center(stack.X(), stack.Y())

    return game
}

// show roads
func createScenario28(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 28")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneMyrror

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 10190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 10
    city.Wall = false

    city.ResetCitizens(nil)

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 26

    player.LiftFog(x, y, 4, game.Plane)

    nodes := findNodes(game.CurrentMap())
    node := nodes[terrain.SorceryNode][0]
    game.CurrentMap().RemoveEncounter(node.X, node.Y)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, node.X + 1, node.Y + 1, game.Plane, wizard.Banner, nil))
    // has pathfinding
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.NomadRangers, node.X + 1, node.Y + 1, game.Plane, wizard.Banner, nil))

    player.LiftFog(node.X, node.Y, 4, game.Plane)

    game.CurrentMap().SetRoad(node.X+3, node.Y+1, true)
    game.CurrentMap().SetRoad(node.X+2, node.Y+1, true)
    game.CurrentMap().SetRoad(node.X+3, node.Y, true)
    game.CurrentMap().SetRoad(node.X+3, node.Y+2, true)

    return game
}

// build roads
func createScenario29(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 29")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneMyrror

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 12190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Buildings.Insert(buildinglib.BuildingFarmersMarket)
    city.Race = wizard.Race
    city.Farmers = 9
    city.Workers = 3
    city.Wall = false

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 26

    player.LiftFog(x, y, 4, city.Plane)

    nodes := findNodes(game.CurrentMap())
    node := nodes[terrain.SorceryNode][0]
    game.CurrentMap().RemoveEncounter(node.X, node.Y)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.OrcSwordsmen, node.X + 1, node.Y + 1, game.Plane, wizard.Banner, nil))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.OrcEngineers, node.X + 1, node.Y + 1, game.Plane, wizard.Banner, nil))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.OrcEngineers, node.X + 1, node.Y + 1, game.Plane, wizard.Banner, nil))

    player.LiftFog(node.X, node.Y, 4, game.Plane)

    game.CurrentMap().SetRoad(node.X+3, node.Y+1, true)
    game.CurrentMap().SetRoad(node.X+2, node.Y+1, true)
    game.CurrentMap().SetRoad(node.X+3, node.Y, true)
    game.CurrentMap().SetRoad(node.X+3, node.Y+2, true)

    return game
}

// connected cities via roads
func createScenario30(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 30")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 12190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 9
    city.Workers = 3
    city.Wall = false

    city.ResetCitizens(nil)

    player.AddCity(city)

    x2, y2 := x + 3, y

    city2 := citylib.MakeCity("City2", x2, y2, data.RaceHighElf, player.GetBanner(), player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    city2.Plane = city.Plane
    city2.Population = 6000
    city2.ResetCitizens(nil)

    player.AddCity(city2)

    city3 := citylib.MakeCity("City3", x + 1, y2+2, data.RaceHighElf, player.GetBanner(), player.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    city3.Plane = city.Plane
    city3.Population = 6000
    city3.ResetCitizens(nil)

    player.AddCity(city3)

    player.Gold = 83
    player.Mana = 26

    player.LiftFog(x, y, 8, game.Plane)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.OrcEngineers, x + 1, y + 1, game.Plane, wizard.Banner, nil))

    for i := x; i <= x2; i++ {
        game.CurrentMap().SetRoad(i, y, false)
    }

    log.Printf("Connected city->city2: %v", game.IsCityRoadConnected(city, city2))
    log.Printf("Connected city2->city: %v", game.IsCityRoadConnected(city, city2))
    log.Printf("Connected city->city3: %v", game.IsCityRoadConnected(city, city3))
    log.Printf("Connected city3->city2: %v", game.IsCityRoadConnected(city3, city2))

    return game
}

// test getting treasure
func createScenario31(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 31")

    wizard := setup.WizardCustom{
        Name: "bob",
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)
    player.StrategicCombat = true

    x, y, _ := game.FindValidCityLocation(game.Plane)

    /*
    x = 20
    y = 20
    */

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, fraction.Zero(), game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3
    city.Wall = false
    city.Buildings.Insert(buildinglib.BuildingFortress)

    city.ResetCitizens(nil)

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 2600

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    // log.Printf("City at %v, %v", x, y)

    player.LiftFog(x, y, 30, data.PlaneArcanus)

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))

    for i := 0; i < 4; i++ {
        player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))
    }

    // player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, 30, 30, data.PlaneArcanus, wizard.Banner))

    stack := player.FindStackByUnit(drake)
    player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)

    if !game.CurrentMap().CreateEncounter(game.CurrentMap().WrapX(stack.X() + 1), stack.Y(), maplib.EncounterTypeLair, data.DifficultyAverage, false, data.PlaneArcanus) {
        log.Printf("Unable to create encounter")
    }

    game.Camera.Center(stack.X(), stack.Y())

    game.Events <- &gamelib.GameEventTreasure{
        Player: player,
        Treasure: gamelib.Treasure{
            Treasures: []gamelib.TreasureItem{
                /*
                &gamelib.TreasureGold{
                    Amount: 300,
                },
                */
                &gamelib.TreasurePrisonerHero{
                    Hero: game.Heroes[0],
                },
            },
        },
    }

    return game
}

// test a hero dying in combat
func createScenario32(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 32")

    wizard := setup.WizardCustom{
        Name: "bob",
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    /*
    x = 20
    y = 20
    */

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, fraction.Zero(), game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3
    city.Wall = false
    city.Buildings.Insert(buildinglib.BuildingFortress)

    city.ResetCitizens(nil)

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 2600

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    // log.Printf("City at %v, %v", x, y)

    player.LiftFog(x, y, 30, data.PlaneArcanus)

    gunther := hero.MakeHero(units.MakeOverworldUnit(units.HeroGunther), hero.HeroGunther, "Gunther")
    gunther.Equipment[0] = &artifact.Artifact{
        Name: "Baloney",
        Image: 7,
        Type: artifact.ArtifactTypeSword,
        Powers: []artifact.Power{
            {
                Type: artifact.PowerTypeAttack,
                Amount: 1,
                Name: "+1 Attack",
            },
            {
                Type: artifact.PowerTypeDefense,
                Amount: 2,
                Name: "+2 Defense",
            },
        },
        Cost: 250,
    }
    player.AddHero(gunther)

    // player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, 30, 30, data.PlaneArcanus, wizard.Banner))

    stack := player.FindStackByUnit(gunther)
    player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)

    x = game.CurrentMap().WrapX(stack.X() + 1)
    y = stack.Y()

    if !game.CurrentMap().CreateEncounter(x, y, maplib.EncounterTypeLair, data.DifficultyAverage, false, data.PlaneArcanus) {
        log.Printf("Unable to create encounter")
    }

    encounter := game.CurrentMap().GetEncounter(x, y)
    encounter.Units = []units.Unit{units.SkyDrake, units.SkyDrake}

    game.Camera.Center(stack.X(), stack.Y())

    return game
}

// priest purifies a tile
func createScenario33(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 33")

    wizard := setup.WizardCustom{
        Name: "bob",
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, fraction.Zero(), game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.ProducingBuilding = buildinglib.BuildingTradeGoods
    city.ProducingUnit = units.UnitNone
    city.Buildings.Insert(buildinglib.BuildingGranary)
    city.Buildings.Insert(buildinglib.BuildingFarmersMarket)
    city.Buildings.Insert(buildinglib.BuildingForestersGuild)
    city.Race = wizard.Race
    city.Farmers = 13
    city.Workers = 3
    city.Wall = false

    city.ResetCitizens(nil)

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 2600

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    // log.Printf("City at %v, %v", x, y)

    player.LiftFog(x, y, 30, data.PlaneArcanus)

    priest1 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.BeastmenPriest, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))
    priest2 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.BeastmenPriest, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))

    _ = priest2

    game.CurrentMap().SetCorruption(game.CurrentMap().WrapX(x + 1), y)

    // player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, 30, 30, data.PlaneArcanus, wizard.Banner))

    stack := player.FindStackByUnit(priest1)
    player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)

    game.Camera.Center(stack.X(), stack.Y())

    return game
}

// warp node
func createScenario34(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 34")

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    wizard1 := setup.WizardCustom{
        Name: "Rjak",
        Base: data.WizardRjak,
        Banner: data.BannerPurple,
    }

    player1 := game.AddPlayer(wizard1, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("", x, y, data.RaceBarbarian, player1.Wizard.Banner, player1.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard1.Banner
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard1.Race
    city.Farmers = 3
    city.Workers = 3
    city.Wall = false

    city.ResetCitizens(nil)

    player1.AddCity(city)

    player1.Gold = 830
    player1.Mana = 26557
    player1.CastingSkillPower = 10000

    player1.LiftFog(x, y, 3, data.PlaneArcanus)

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player1.KnownSpells.AddSpell(allSpells.FindByName("Warp Node"))

    wizard2 := setup.WizardCustom{
        Name: "Merlin",
        Base: data.WizardMerlin,
        Banner: data.BannerYellow,
    }
    player2 := game.AddPlayer(wizard2, true)

    nodes := findNodes(game.CurrentMap())
    node := nodes[terrain.NatureNode][0]
    game.CurrentMap().RemoveEncounter(node.X, node.Y)
    node.Node.MeldingWizard = player2
    player1.LiftFog(node.X, node.Y, 3, data.PlaneArcanus)

    spirit := player1.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, node.X + 1, node.Y + 1, data.PlaneArcanus, wizard1.Banner, nil))
    stack := player1.FindStackByUnit(spirit)
    player1.SetSelectedStack(stack)
    player1.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)

    node = nodes[terrain.SorceryNode][0]
    game.CurrentMap().RemoveEncounter(node.X, node.Y)
    node.Node.MeldingWizard = player2
    node.Node.Warped = true
    player1.LiftFog(node.X, node.Y, 3, data.PlaneArcanus)

    node = nodes[terrain.ChaosNode][0]
    game.CurrentMap().RemoveEncounter(node.X, node.Y)
    node.Node.MeldingWizard = player2
    node.Node.Warped = true
    player1.LiftFog(node.X, node.Y, 3, data.PlaneArcanus)

    return game
}

// all spells in spellbook
func createScenario35(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 35")

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    wizard1 := setup.WizardCustom{
        Name: "Rjak",
        Base: data.WizardRjak,
        Banner: data.BannerPurple,
    }

    player1 := game.AddPlayer(wizard1, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("", x, y, data.RaceBarbarian, player1.Wizard.Banner, player1.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard1.Banner
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard1.Race
    city.Farmers = 3
    city.Workers = 3
    city.Wall = false

    city.ResetCitizens(nil)

    player1.AddCity(city)

    player1.Gold = 830
    player1.Mana = 26557
    player1.CastingSkillPower = 10000

    player1.LiftFog(x, y, 3, data.PlaneArcanus)

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)
    for _, spell := range allSpells.Spells {
        player1.KnownSpells.AddSpell(spell)
    }

    spirit := player1.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x + 1, y + 1, data.PlaneArcanus, wizard1.Banner, nil))
    stack := player1.FindStackByUnit(spirit)
    player1.SetSelectedStack(stack)
    player1.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)

    return game
}

// enemy wizard controlled town
func createScenario36(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 36")

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    wizard2 := setup.WizardCustom{
        Name: "Merlin",
        Base: data.WizardMerlin,
        Banner: data.BannerYellow,
    }
    human := game.AddPlayer(wizard2, true)
    human.Admin = true

    wizard1 := setup.WizardCustom{
        Name: "Rjak",
        Base: data.WizardRjak,
        Banner: data.BannerPurple,
        Race: data.RaceBarbarian,
    }

    enemy1 := game.AddPlayer(wizard1, false)

    enemy1.AIBehavior = ai.MakeEnemyAI()

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("ai1", x, y, data.RaceBarbarian, enemy1.Wizard.Banner, enemy1.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Banner = wizard1.Banner
    city.ProducingBuilding = buildinglib.BuildingTradeGoods
    city.ProducingUnit = units.UnitNone
    city.Farmers = 3
    city.Workers = 3
    city.Wall = false

    city.ResetCitizens(nil)

    enemy1.AddCity(city)

    enemy1.Gold = 830
    enemy1.Mana = 26557
    enemy1.CastingSkillPower = 10000

    enemy1.LiftFog(x, y, 3, data.PlaneArcanus)

    // allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    human.LiftFog(20, 20, 100, data.PlaneArcanus)

    return game
}

// fleeing from an encounter
func createScenario37(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 37")

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })
    game.Plane = data.PlaneArcanus

    wizard := setup.WizardCustom{
        Name: "Rjak",
        Base: data.WizardRjak,
        Banner: data.BannerPurple,
    }

    nodes := findNodes(game.CurrentMap())
    node := nodes[terrain.NatureNode][0]

    player := game.AddPlayer(wizard, true)
    player.Gold = 830
    player.Mana = 26557
    player.CastingSkillPower = 10000
    player.LiftFog(node.X, node.Y, 3, data.PlaneArcanus)

    unit := player.AddUnit(units.MakeOverworldUnitFromUnit(units.SkyDrake, node.X + 1, node.Y + 1, data.PlaneArcanus, wizard.Banner, nil))
    stack := player.FindStackByUnit(unit)
    player.SetSelectedStack(stack)

    return game
}

// fleeing from an enemy
func createScenario38(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 38")

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })
    game.Plane = data.PlaneArcanus

    wizard1 := setup.WizardCustom{
        Name: "Rjak",
        Base: data.WizardRjak,
        Banner: data.BannerPurple,
    }


    player1 := game.AddPlayer(wizard1, true)
    player1.Gold = 830
    player1.Mana = 26557
    player1.CastingSkillPower = 10000

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city1 := citylib.MakeCity("", x, y, data.RaceBarbarian, player1.Wizard.Banner, player1.TaxRate, game.BuildingInfo, game.CurrentMap(), game)
    city1.Population = 6190
    city1.Plane = data.PlaneArcanus
    city1.Banner = wizard1.Banner
    city1.ProducingBuilding = buildinglib.BuildingGranary
    city1.ProducingUnit = units.UnitNone
    city1.Race = wizard1.Race
    city1.Farmers = 3
    city1.Workers = 3
    city1.Wall = false
    city1.ResetCitizens(nil)
    player1.AddCity(city1)
    player1.LiftFog(x, y, 3, data.PlaneArcanus)

    x, y, _ = game.FindValidCityLocation(game.Plane)
    player1.LiftFog(x, y, 3, data.PlaneArcanus)

    unit1 := player1.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x, y, data.PlaneArcanus, wizard1.Banner, nil))
    stack1 := player1.FindStackByUnit(unit1)
    player1.SetSelectedStack(stack1)

    wizard2 := setup.WizardCustom{
        Name: "Merlin",
        Base: data.WizardMerlin,
        Banner: data.BannerYellow,
    }

    player2 := game.AddPlayer(wizard2, false)
    player2.AIBehavior = ai.MakeRaiderAI()
    player2.Mana = 26557

    player2.AddUnit(units.MakeOverworldUnitFromUnit(units.Basilisk, x+1, y-2, data.PlaneArcanus, wizard2.Banner, nil))
    player2.AddUnit(units.MakeOverworldUnitFromUnit(units.Basilisk, x+1, y-1, data.PlaneArcanus, wizard2.Banner, nil))
    player2.AddUnit(units.MakeOverworldUnitFromUnit(units.Basilisk, x+1, y, data.PlaneArcanus, wizard2.Banner, nil))
    player2.AddUnit(units.MakeOverworldUnitFromUnit(units.Basilisk, x-1, y-2, data.PlaneArcanus, wizard2.Banner, nil))
    player2.AddUnit(units.MakeOverworldUnitFromUnit(units.Basilisk, x-1, y-1, data.PlaneArcanus, wizard2.Banner, nil))
    player2.AddUnit(units.MakeOverworldUnitFromUnit(units.Basilisk, x-1, y, data.PlaneArcanus, wizard2.Banner, nil))

    return game
}

// town doesn't produce enough food
func createScenario39(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 39")

    wizard := setup.WizardCustom{
        Name: "bob",
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    /*
    x = 20
    y = 20
    */

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, fraction.Make(3, 1), game.BuildingInfo, game.CurrentMap(), game)
    city.Population = 6300
    city.Plane = data.PlaneArcanus
    city.Banner = wizard.Banner
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.AddEnchantment(data.CityEnchantmentFamine, player.GetBanner())
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3
    city.Wall = false

    city.ResetCitizens(nil)

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 2600

    // remove all bonuses surrounding the city
    // and force food availability to be very low
    mapUse := game.CurrentMap()
    for dx := -2; dx <= 2; dx++ {
        for dy := -2; dy <= 2; dy++ {
            cx := mapUse.WrapX(x + dx)
            cy := y + dy
            mapUse.SetBonus(cx, cy, data.BonusNone)
            if (cx+cy) % 2 == 0 {
                mapUse.Map.Terrain[cx][cy] = int(terrain.IndexForest1)
            } else {
                mapUse.Map.Terrain[cx][cy] = int(terrain.IndexDesert1)
            }
        }
    }

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    // log.Printf("City at %v, %v", x, y)

    player.LiftFog(x, y, 30, data.PlaneArcanus)
    game.Camera.Center(x, y)

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
        case 6: game = createScenario6(cache)
        case 7: game = createScenario7(cache)
        case 8: game = createScenario8(cache)
        case 9: game = createScenario9(cache)
        case 10: game = createScenario10(cache)
        case 11: game = createScenario11(cache)
        case 12: game = createScenario12(cache)
        case 13: game = createScenario13(cache)
        case 14: game = createScenario14(cache)
        case 15: game = createScenario15(cache)
        case 16: game = createScenario16(cache)
        case 17: game = createScenario17(cache)
        case 18: game = createScenario18(cache)
        case 19: game = createScenario19(cache)
        case 20: game = createScenario20(cache)
        case 21: game = createScenario21(cache)
        case 22: game = createScenario22(cache)
        case 23: game = createScenario23(cache)
        case 24: game = createScenario24(cache)
        case 25: game = createScenario25(cache)
        case 26: game = createScenario26(cache)
        case 27: game = createScenario27(cache)
        case 28: game = createScenario28(cache)
        case 29: game = createScenario29(cache)
        case 30: game = createScenario30(cache)
        case 31: game = createScenario31(cache)
        case 32: game = createScenario32(cache)
        case 33: game = createScenario33(cache)
        case 34: game = createScenario34(cache)
        case 35: game = createScenario35(cache)
        case 36: game = createScenario36(cache)
        case 37: game = createScenario37(cache)
        case 38: game = createScenario38(cache)
        case 39: game = createScenario39(cache)
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
        Console: console.MakeConsole(),
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
    return data.ScreenWidth, data.ScreenHeight
}

func main(){

    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    profile, err := os.Create("profile.cpu.overworld")
    if err != nil {
        log.Printf("Error creating profile: %v", err)
    } else {
        defer profile.Close()
        pprof.StartCPUProfile(profile)
        defer pprof.StopCPUProfile()
    }

    monitorWidth, _ := ebiten.Monitor().Size()

    size := monitorWidth / 390

    ebiten.SetWindowSize(data.ScreenWidth / data.ScreenScale * size, data.ScreenHeight / data.ScreenScale * size)
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

    memoryProfile, err := os.Create("profile.mem.overworld")
    if err != nil {
        log.Printf("Error creating memory profile: %v", err)
    } else {
        defer memoryProfile.Close()
        pprof.WriteHeapProfile(memoryProfile)
    }

}
