package main

import (
    "os"
    "log"
    "fmt"
    "strconv"
    "math/rand"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    mouselib "github.com/kazzmir/master-of-magic/lib/mouse"
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

    x, y := game.FindValidCityLocation()

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, fraction.Zero(), game.BuildingInfo, game.CurrentMap())
    city.Population = 16190
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

    game.CenterCamera(x, y)

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

    introCity := citylib.MakeCity("Test City", 4, 5, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap())
    introCity.Population = 6000
    introCity.Plane = data.PlaneArcanus
    introCity.ProducingBuilding = buildinglib.BuildingHousing
    introCity.ProducingUnit = units.UnitNone
    introCity.Wall = false

    player.AddCity(introCity)

    player.Gold = 83
    player.Mana = 26

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

    x, y := game.FindValidCityLocation()

    introCity := citylib.MakeCity("Test City", x, y, player.Wizard.Race, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap())
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

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x+1, y, data.PlaneArcanus, wizard.Banner, nil))

    settlers := player.AddUnit(units.MakeOverworldUnitFromUnit(units.LizardSettlers, x+1, y, data.PlaneArcanus, wizard.Banner, nil))

    stack := player.FindStackByUnit(settlers)
    player.SetSelectedStack(stack)

    _ = introCity
    // game.Events <- gamelib.StartingCityEvent(introCity)

    // game.Events <- &gamelib.GameEventNewOutpost{City: introCity, Stack: nil}

    player.LiftFog(x, y, 2, data.PlaneArcanus)

    game.CenterCamera(x, y)

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
                Magic: data.SorceryMagic,
                Count: 8,
            },
        },
    }

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y := game.FindValidCityLocation()

    introCity := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap())
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

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x+1, y, data.PlaneArcanus, wizard.Banner, nil))

    settlers := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSettlers, x+1, y, data.PlaneArcanus, wizard.Banner, nil))

    stack := player.FindStackByUnit(settlers)
    player.SetSelectedStack(stack)

    _ = introCity
    // game.Events <- gamelib.StartingCityEvent(introCity)

    player.LiftFog(x, y, 2, data.PlaneArcanus)

    game.CenterCamera(x, y)

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

    x, y := game.FindValidCityLocation()

    introCity := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap())
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

    game.CenterCamera(x, y)

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

    x, y := game.FindValidCityLocation()

    introCity := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap())
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

    for i := 0; i < 3; i++ {
        player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x + i, y + 1, data.PlaneArcanus, wizard.Banner, nil))
    }

    for i := 0; i < 3; i++ {
        player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 3 + i, y + 1, data.PlaneArcanus, wizard.Banner, nil))
    }

    for i := 0; i < 3; i++ {
        player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, x + 6 + i, y + 1, data.PlaneArcanus, wizard.Banner, nil))
    }

    x, y = game.FindValidCityLocation()

    city2 := citylib.MakeCity("utah", x, y, data.RaceDarkElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap())
    city2.Population = 7000
    city2.Plane = data.PlaneArcanus
    city2.ProducingBuilding = buildinglib.BuildingShrine
    city2.ProducingUnit = units.UnitNone
    city2.Wall = false

    city2.ResetCitizens(nil)

    player.AddCity(city2)

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    game.CenterCamera(x, y)

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
        x, y := game.FindValidCityLocation()
        player.LiftFog(x, y, 3, data.PlaneArcanus)

        introCity := citylib.MakeCity(fmt.Sprintf("city%v", i), x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap())
        introCity.Population = rand.Intn(5000) + 5000
        introCity.Plane = data.PlaneArcanus
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

    x, y := game.FindValidCityLocation()

    game.CurrentMap().CreateNode(x, y+1, maplib.MagicNodeNature, game.Plane, game.Settings.Magic, game.Settings.Difficulty)
    game.CurrentMap().CreateNode(x+1, y, maplib.MagicNodeChaos, game.Plane, game.Settings.Magic, game.Settings.Difficulty)
    game.CurrentMap().CreateNode(x+2, y+1, maplib.MagicNodeSorcery, game.Plane, game.Settings.Magic, game.Settings.Difficulty)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap())
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

    x, y := game.FindValidCityLocation()

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap())
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
        Wizard: wizard.Base,
        Unit: units.FireGiant,
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

    x, y := game.FindValidCityLocation()

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap())
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

    x, y := game.FindValidCityLocation()

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap())
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

    node := game.CurrentMap().CreateNode(x, y+2, maplib.MagicNodeNature, game.Plane, game.Settings.Magic, game.Settings.Difficulty)
    node.Empty = true

    spirit := player.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))

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

    x, y := game.FindValidCityLocation()

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap())
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

    player.LiftFog(x, y, 4, data.PlaneArcanus)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, nil))

    node := game.CurrentMap().CreateNode(x, y+2, maplib.MagicNodeNature, game.Plane, game.Settings.Magic, game.Settings.Difficulty)
    node.Empty = true

    game.CurrentMap().SetBonus(x-3, y-1, data.BonusSilverOre)
    game.CurrentMap().SetBonus(x-2, y-1, data.BonusGem)
    game.CurrentMap().SetBonus(x-1, y-1, data.BonusWildGame)
    game.CurrentMap().SetBonus(x, y-1, data.BonusQuorkCrystal)
    game.CurrentMap().SetBonus(x+1, y-1, data.BonusIronOre)

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

    player.CastingSkillPower += 5000

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player.KnownSpells.AddSpell(allSpells.FindByName("Earth Lore"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Giant Strength"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Ice Bolt"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Enchant Item"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Create Artifact"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Magic Spirit"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Dark Rituals"))

    x, y := game.FindValidCityLocation()

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap())
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

    x, y := game.FindValidCityLocation()

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap())
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
    game.CenterCamera(x, y)

    /*
    game.Events <- &gamelib.GameEventLearnedSpell{
        Player: player,
        // Spell: allSpells.FindByName("Earth Lore"),
        Spell: allSpells.FindByName("Magic Spirit"),
    }
    */

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

    x, y := game.FindValidCityLocation()

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap())
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

    x, y := game.FindValidCityLocation()

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, fraction.Zero(), game.BuildingInfo, game.CurrentMap())
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

    x, y := game.FindValidCityLocation()

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap())
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
            &artifact.PowerAttack{
                Amount: 1,
            },
            &artifact.PowerDefense{
                Amount: 2,
            },
        },
    }

    player.VaultEquipment[1] = &artifact.Artifact{
        Name: "Pizza",
        Image: 31,
        Type: artifact.ArtifactTypeBow,
        Powers: []artifact.Power{
            &artifact.PowerAttack{
                Amount: 1,
            },
            &artifact.PowerMovement{
                Amount: 2,
            },
        },
    }

    testArtifact := artifact.Artifact{
        Name: "Sword",
        Image: 5,
        Type: artifact.ArtifactTypeSword,
        Powers: []artifact.Power{
            &artifact.PowerAttack{
                Amount: 2,
            },
            &artifact.PowerDefense{
                Amount: 2,
            },
            &artifact.PowerMovement{
                Amount: 3,
            },
            &artifact.PowerResistance{
                Amount: 2,
            },
        },
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

    x, y := game.FindValidCityLocation()

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap())
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

    /*
    rakir := hero.MakeHero(units.MakeOverworldUnit(units.HeroRakir), hero.HeroRakir, "bubba")
    player.AddHero(rakir)
    rakir.AddExperience(528)
    */
    mysticX := hero.MakeHero(units.MakeOverworldUnit(units.HeroMysticX), hero.HeroMysticX, "bubba")
    player.AddHero(mysticX)
    mysticX.SetExtraAbilities()
    mysticX.AddExperience(528)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, x, y, data.PlaneArcanus, player.GetBanner(), nil))

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

    x, y := game.FindValidCityLocation()

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap())
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
    game.CenterCamera(x, y)

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

    x, y := game.FindValidCityLocation()

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, fraction.Zero(), game.BuildingInfo, game.CurrentMap())
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

    enemyWizard := setup.WizardCustom{
        Name: "enemy",
        Banner: data.BannerBrown,
        Race: data.RaceDraconian,
    }

    enemy := game.AddPlayer(enemyWizard, false)

    city2 := citylib.MakeCity("Test City", x, y-1, enemy.Wizard.Race, enemy.Wizard.Banner, fraction.Make(1, 1), game.BuildingInfo, game.CurrentMap())
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

    x, y := game.FindValidCityLocation()

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, fraction.Zero(), game.BuildingInfo, game.CurrentMap())
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

    x2, y2 := game.FindValidCityLocationOnContinent(city.X, city.Y)
    log.Printf("enemy city at %v, %v", x2, y2)
    city2 := citylib.MakeCity("Test City", x2, y2, enemy.Wizard.Race, enemy.Wizard.Banner, fraction.Make(1, 1), game.BuildingInfo, game.CurrentMap())
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

    x, y := game.FindValidCityLocation()

    // game.Map.CreateEncounter(x, y+1, maplib.EncounterTypeLair, game.Settings.Difficulty, false, game.Plane)
    game.CurrentMap().CreateEncounterRandom(x, y+1, game.Settings.Difficulty, game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, player.Wizard.Banner, player.TaxRate, game.BuildingInfo, game.CurrentMap())
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

    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)

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
        default: game = createScenario1(cache)
    }

    game.DoNextTurn()

    run := func(yield coroutine.YieldFunc) error {
        for game.Update(yield) != gamelib.GameStateQuit {
            yield()
        }

        return ebiten.Termination
    }

    normalMouse, err := mouselib.GetMouseNormal(cache)
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
