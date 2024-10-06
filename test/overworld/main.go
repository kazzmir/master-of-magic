package main

import (
    "os"
    "log"
    "fmt"
    "strconv"
    "math/rand"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
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

    player.AddUnit(units.OverworldUnit{
        Unit: units.HighMenSpearmen,
        Plane: data.PlaneArcanus,
        Banner: wizard.Banner,
        X: 30,
        Y: 30,
    })

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
    log.Printf("Running scenario 3")
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

    for i := 0; i < 3; i++ {
        player.AddUnit(units.OverworldUnit{
            Unit: units.HighMenSpearmen,
            Plane: data.PlaneArcanus,
            Banner: wizard.Banner,
            X: x + i,
            Y: y + 1,
        })
    }

    for i := 0; i < 3; i++ {
        player.AddUnit(units.OverworldUnit{
            Unit: units.GreatDrake,
            Plane: data.PlaneArcanus,
            Banner: wizard.Banner,
            X: x + 3 + i,
            Y: y + 1,
        })
    }

    for i := 0; i < 3; i++ {
        player.AddUnit(units.OverworldUnit{
            Unit: units.FireElemental,
            Plane: data.PlaneArcanus,
            Banner: wizard.Banner,
            X: x + 6 + i,
            Y: y + 1,
        })
    }

    x, y = game.FindValidCityLocation()

    city2 := citylib.MakeCity("utah", x, y, data.RaceDarkElf, player.TaxRate, game.BuildingInfo)
    city2.Population = 7000
    city2.Plane = data.PlaneArcanus
    city2.ProducingBuilding = buildinglib.BuildingShrine
    city2.ProducingUnit = units.UnitNone
    city2.Wall = false

    city2.ResetCitizens(nil)

    player.AddCity(city2)

    player.LiftFog(x, y, 3)

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

    player := game.AddPlayer(wizard)

    for i := 0; i < 15; i++ {
        x, y := game.FindValidCityLocation()
        player.LiftFog(x, y, 3)

        introCity := citylib.MakeCity(fmt.Sprintf("city%v", i), x, y, data.RaceHighElf, player.TaxRate, game.BuildingInfo)
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

    player := game.AddPlayer(wizard)

    x, y := game.FindValidCityLocation()

    game.Map.CreateNode(x, y+1, gamelib.MagicNodeNature, game.Plane, game.Settings.Magic, game.Settings.Difficulty)
    game.Map.CreateNode(x+1, y, gamelib.MagicNodeChaos, game.Plane, game.Settings.Magic, game.Settings.Difficulty)
    game.Map.CreateNode(x+2, y+1, gamelib.MagicNodeSorcery, game.Plane, game.Settings.Magic, game.Settings.Difficulty)

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

    for i := 0; i < 1; i++ {
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

    return game
}

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

    for i := 0; i < 1; i++ {
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

    game.Events <- &gamelib.GameEventSummonUnit{
        Wizard: wizard.Base,
        Unit: units.FireGiant,
    }

    player.LiftFog(stack.X(), stack.Y(), 2)

    return game
}

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

    for i := 0; i < 1; i++ {
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

    game.Events <- &gamelib.GameEventSummonHero{
        Wizard: wizard.Base,
        Champion: false,
    }

    player.LiftFog(stack.X(), stack.Y(), 2)

    return game
}

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

    player.LiftFog(x, y, 3)

    node := game.Map.CreateNode(x, y+2, gamelib.MagicNodeNature, game.Plane, game.Settings.Magic, game.Settings.Difficulty)
    node.Empty = true

    spirit := player.AddUnit(units.OverworldUnit{
        Unit: units.MagicSpirit,
        Plane: data.PlaneArcanus,
        Banner: wizard.Banner,
        X: x + 1,
        Y: y + 1,
    })

    stack := player.FindStackByUnit(spirit)
    player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 2)

    return game
}

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

    player.LiftFog(x, y, 4)

    player.AddUnit(units.OverworldUnit{
        Unit: units.MagicSpirit,
        Plane: data.PlaneArcanus,
        Banner: wizard.Banner,
        X: x + 1,
        Y: y + 1,
    })

    node := game.Map.CreateNode(x, y+2, gamelib.MagicNodeNature, game.Plane, game.Settings.Magic, game.Settings.Difficulty)
    node.Empty = true

    game.Map.SetBonus(x-3, y-1, gamelib.BonusSilverOre)
    game.Map.SetBonus(x-2, y-1, gamelib.BonusGem)
    game.Map.SetBonus(x-1, y-1, gamelib.BonusWildGame)
    game.Map.SetBonus(x, y-1, gamelib.BonusQuorkCrystal)
    game.Map.SetBonus(x+1, y-1, gamelib.BonusIronOre)

    game.Events <- &gamelib.GameEventSurveyor{}

    return game
}

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

    player := game.AddPlayer(wizard)

    player.CastingSkillPower += 500

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player.KnownSpells.AddSpell(allSpells.FindByName("Earth Lore"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Giant Strength"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Ice Bolt"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Enchant Item"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Magic Spirit"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Dark Rituals"))

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
    player.Mana = 50

    player.LiftFog(x, y, 4)

    player.AddUnit(units.OverworldUnit{
        Unit: units.MagicSpirit,
        Plane: data.PlaneArcanus,
        Banner: wizard.Banner,
        X: x + 1,
        Y: y + 1,
    })

    return game
}

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

    player := game.AddPlayer(wizard)

    player.CastingSkillPower += 500

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    for _, name := range []string{"Earth Lore", "Giant Strength", "Ice Bolt", "Enchant Item", "Dark Rituals", "Spell Blast", "Time Stop", "Web", "Magic Spirit"} {
        player.ResearchPoolSpells.AddSpell(allSpells.FindByName(name))
    }

    // player.KnownSpells.AddSpell(allSpells.FindByName("Earth Lore"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Giant Strength"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Ice Bolt"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Enchant Item"))
    // player.Spells.AddSpell(allSpells.FindByName("Magic Spirit"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Dark Rituals"))

    player.ResearchCandidateSpells.AddSpell(allSpells.FindByName("Magic Spirit"))
    player.ResearchCandidateSpells.AddSpell(allSpells.FindByName("Endurance"))
    player.ResearchCandidateSpells.AddSpell(allSpells.FindByName("Hell Hounds"))
    player.ResearchCandidateSpells.AddSpell(allSpells.FindByName("Healing"))
    player.ResearchCandidateSpells.AddSpell(allSpells.FindByName("Corruption"))
    player.ResearchCandidateSpells.AddSpell(allSpells.FindByName("Dispel Magic"))
    player.ResearchCandidateSpells.AddSpell(allSpells.FindByName("Summoning Circle"))
    player.ResearchCandidateSpells.AddSpell(allSpells.FindByName("Just Cause"))
    player.ResearchCandidateSpells.AddSpell(allSpells.FindByName("Detect Magic"))
    // player.ResearchingSpell = allSpells.FindByName("Magic Spirit")
    // player.ResearchingSpell = allSpells.FindByName("Earth Lore")
    // player.ResearchProgress = 10

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
    player.Mana = 50

    player.LiftFog(x, y, 4)

    /*
    game.Events <- &gamelib.GameEventLearnedSpell{
        Player: player,
        // Spell: allSpells.FindByName("Earth Lore"),
        Spell: allSpells.FindByName("Magic Spirit"),
    }
    */

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
