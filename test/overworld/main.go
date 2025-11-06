package main

import (
    "os"
    "log"
    "fmt"
    "image"
    "strconv"
    "math/rand/v2"
    "runtime/pprof"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    mouselib "github.com/kazzmir/master-of-magic/lib/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
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
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
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
        Retorts: []data.Retort{
            data.RetortAlchemy,
            data.RetortSageMaster,
            data.RetortRunemaster,
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
    player.Admin = true

    player.TaxRate = fraction.Zero()

    x, y, _ := game.FindValidCityLocation(game.Plane)

    /*
    x = 20
    y = 20
    */

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.Buildings.Insert(buildinglib.BuildingGranary)
    city.Buildings.Insert(buildinglib.BuildingFarmersMarket)
    city.ProducingBuilding = buildinglib.BuildingBank
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 14
    city.Workers = 3

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 2600
    player.CastingSkillPower = 10000

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player.KnownSpells.AddSpell(allSpells.FindByName("Animate Dead"))

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    // log.Printf("City at %v, %v", x, y)

    player.LiftFog(x, y, 30, data.PlaneArcanus)
    player.LiftFog(x, y, 30, data.PlaneMyrror)

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    for i := 0; i < 5; i++ {
        fireElemental := player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
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

    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))
    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))

    player.AddPowerHistory(playerlib.WizardPower{Army: 30, Magic: 4053, SpellResearch: 40})
    enemy1.AddPowerHistory(playerlib.WizardPower{Army: 15, Magic: 3088, SpellResearch: 30})

    game.Camera.Center(stack.X(), stack.Y())

    return game
}

// test the starting city name input
func createScenario2(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 2")
    wizard := setup.WizardCustom{
        Name: "player",
        Banner: data.BannerBlue,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    introCity := citylib.MakeCity("Test City", 4, 5, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    introCity.Population = 6000
    introCity.Plane = data.PlaneArcanus
    introCity.ProducingBuilding = buildinglib.BuildingHousing
    introCity.ProducingUnit = units.UnitNone

    player.AddCity(introCity)

    player.Gold = 83
    player.Mana = 2600

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(4, 5, 3, data.PlaneArcanus)

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, 5, 5, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    introCity := citylib.MakeCity("Test City", x, y, player.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, player)
    introCity.Population = 9000
    introCity.Plane = data.PlaneArcanus
    introCity.ProducingBuilding = buildinglib.BuildingHousing
    introCity.ProducingUnit = units.UnitNone
    introCity.Farmers = 9

    introCity.AddBuilding(buildinglib.BuildingShrine)
    introCity.AddBuilding(buildinglib.BuildingGranary)

    introCity.ResetCitizens()

    player.AddCity(introCity)

    player.Gold = 5000
    player.Mana = 260

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    for range 4 {
        newUnit := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
        newUnit.AddEnchantment(data.UnitEnchantmentGiantStrength)
        player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    }

    settlers := player.AddUnit(units.MakeOverworldUnitFromUnit(units.LizardSettlers, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    introCity := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    introCity.Population = 10000
    introCity.Plane = data.PlaneArcanus
    introCity.ProducingBuilding = buildinglib.BuildingHousing
    introCity.ProducingUnit = units.UnitNone
    introCity.Farmers = 10

    introCity.AddBuilding(buildinglib.BuildingShrine)

    introCity.ResetCitizens()

    player.AddCity(introCity)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    settlers := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSettlers, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

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
        // Building: buildinglib.BuildingSmithy,
        // Building: buildinglib.BuildingFantasticStable,
        // Building: buildinglib.BuildingWizardsGuild,
        Building: buildinglib.BuildingMaritimeGuild,
        // Building: buildinglib.BuildingArmorersGuild,
    }

    return game
}

// city growth message
func createScenario5(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 5")
    wizard := setup.WizardCustom{
        Name: "player",
        Banner: data.BannerBlue,
        Race: data.RaceHighMen,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    introCity := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    introCity.Population = 6000
    introCity.Plane = data.PlaneArcanus
    introCity.ProducingBuilding = buildinglib.BuildingHousing
    introCity.ProducingUnit = units.UnitNone

    introCity.AddBuilding(buildinglib.BuildingShrine)

    introCity.ResetCitizens()

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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    introCity := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    introCity.Population = 14000
    introCity.Plane = data.PlaneArcanus
    introCity.ProducingBuilding = buildinglib.BuildingHousing
    introCity.ProducingUnit = units.UnitNone
    introCity.Farmers = 14

    introCity.AddBuilding(buildinglib.BuildingShrine)

    introCity.ResetCitizens()

    player.AddCity(introCity)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    _ = introCity

    for i := 0; i < 3; i++ {
        player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x + i, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    }

    for i := 0; i < 3; i++ {
        player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 3 + i, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    }

    for i := 0; i < 3; i++ {
        player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, x + 6 + i, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    }

    x, y, _ = game.FindValidCityLocation(game.Plane)

    city2 := citylib.MakeCity("utah", x, y, data.RaceDarkElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city2.Population = 7000
    city2.Plane = data.PlaneArcanus
    city2.ProducingBuilding = buildinglib.BuildingShrine
    city2.ProducingUnit = units.UnitNone

    city2.ResetCitizens()

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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    for i := 0; i < 15; i++ {
        x, y, _ := game.FindValidCityLocation(game.Plane)
        player.LiftFog(x, y, 3, data.PlaneArcanus)

        introCity := citylib.MakeCity(fmt.Sprintf("city%v", i), x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
        introCity.Population = rand.N(5000) + 5000
        introCity.Plane = data.PlaneArcanus
        introCity.ProducingBuilding = buildinglib.BuildingHousing
        introCity.ProducingUnit = units.UnitNone

        introCity.AddBuilding(buildinglib.BuildingShrine)

        introCity.ResetCitizens()

        player.AddCity(introCity)
    }

    for i := 0; i < 4; i++ {
        x, y, _ := game.FindValidCityLocation(data.PlaneMyrror)
        player.LiftFog(x, y, 3, data.PlaneMyrror)

        introCity := citylib.MakeCity(fmt.Sprintf("city%v myr", i), x, y, data.RaceHighElf, game.BuildingInfo, game.GetMap(data.PlaneMyrror), game, player)
        introCity.Population = rand.N(5000) + 5000
        introCity.Plane = data.PlaneMyrror
        introCity.ProducingBuilding = buildinglib.BuildingHousing
        introCity.ProducingUnit = units.UnitNone

        introCity.AddBuilding(buildinglib.BuildingShrine)

        introCity.ResetCitizens()

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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6190
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

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    nodes := findNodes(game.CurrentMap())

    node := nodes[terrain.SorceryNode][0]
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, node.X + 1, node.Y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.LiftFog(node.X, node.Y, 3, data.PlaneArcanus)

    node = nodes[terrain.ChaosNode][0]
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, node.X + 1, node.Y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.LiftFog(node.X, node.Y, 3, data.PlaneArcanus)

    node = nodes[terrain.NatureNode][0]
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, node.X + 1, node.Y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    for i := 0; i < 1; i++ {
        fireElemental := player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    for i := 0; i < 1; i++ {
        fireElemental := player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
        _ = fireElemental
    }

    stack := player.FindStackByUnit(drake)
    player.SetSelectedStack(stack)

    game.Events <- &gamelib.GameEventSummonHero{
        Player: player,
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
        Retorts: []data.Retort{
            data.RetortAlchemy,
            data.RetortSageMaster,
            // data.RetortNodeMastery,
            data.RetortSorceryMastery,
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

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 26

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    nodes := findNodes(game.CurrentMap())
    node := nodes[terrain.SorceryNode][0]
    game.CurrentMap().RemoveEncounter(node.X, node.Y)

    player.LiftFog(node.X, node.Y, 3, data.PlaneArcanus)

    spirit := player.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, node.X + 1, node.Y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 26

    player.LiftFog(x, y, 10, data.PlaneArcanus)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.CastingSkillPower += 100000

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player.ResearchPoolSpells = allSpells

    // summoning
    player.KnownSpells.AddSpell(allSpells.FindByName("Guardian Spirit"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Unicorns"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Incarnation"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Angel"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Arch Angel"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Floating Island"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Nagas"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Storm Giant"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Djinn"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Sky Drake"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Phantom Beast"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Phantom Warriors"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Air Elemental"))
    player.KnownSpells.AddSpell(allSpells.FindByName("War Bears"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Sprites"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Giant Spiders"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Cockatrices"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Basilisk"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Stone Giant"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Gorgons"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Behemoth"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Colossus"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Great Wyrm"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Earth Elemental"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Skeletons"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Ghouls"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Night Stalker"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Shadow Demons"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Wraiths"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Death Knights"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Demon Lord"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Hell Hounds"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Fire Giant"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Gargoyles"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Doom Bat"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Chimeras"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Chaos Spawn"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Efreet"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Hydra"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Great Drake"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Fire Elemental"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Magic Spirit"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Summon Champion"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Summon Hero"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Consecration"))

    // special spells
    player.KnownSpells.AddSpell(allSpells.FindByName("Earth Lore"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Giant Strength"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Ice Bolt"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Enchant Item"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Create Artifact"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Enchant Road"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Raise Volcano"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Corruption"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Warp Node"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Call The Void"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Disenchant Area"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Disenchant True"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Change Terrain"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Transmute"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Spell Blast"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Recall Hero"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Word of Recall"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Disjunction"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Disjunction True"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Spell of Mastery"))

    // city spells
    player.KnownSpells.AddSpell(allSpells.FindByName("Wall of Fire"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Wall of Darkness"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Wall of Stone"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Nature's Eye"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Prosperity"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Inspirations"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Gaia's Blessing"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Cursed Lands"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Famine"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Pestilence"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Dark Rituals"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Evil Presence"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Stream of Life"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Altar of Battle"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Astral Gate"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Earth Gate"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Cloud of Shadow"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Summoning Circle"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Move Fortress"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Spell Ward"))

    // global enchantments
    player.KnownSpells.AddSpell(allSpells.FindByName("Nature Awareness"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Awareness"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Crusade"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Just Cause"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Armageddon"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Great Wasting"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Detect Magic"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Eternal Night"))

    // unit enchantments
    player.KnownSpells.AddSpell(allSpells.FindByName("Bless"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Heroism"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Giant Strength"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Lionheart"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Cloak of Fear"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Chaos Channels"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Holy Armor"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Holy Weapon"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Invulnerability"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Planar Travel"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Iron Skin"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Path Finding"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Regeneration"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Stone Skin"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Water Walking"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Guardian Wind"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Invisiblity"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Magic Immunity"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Spell Lock"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Wind Walking"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Eldritch Weapon"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Flame Blade"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Black Channels"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Wraith Form"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Lycanthropy"))

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3

    city.AddBuilding(buildinglib.BuildingFortress)
    city.AddBuilding(buildinglib.BuildingShrine)
    city.AddBuilding(buildinglib.BuildingGranary)

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 1000
    player.Mana = 10000

    player.LiftFog(x, y, 4, data.PlaneArcanus)

    spirit := player.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    spirit.SetBusy(units.BusyStatusStasis)

    game.CurrentMap().SetRoad(x, y+1, false)
    game.CurrentMap().SetRoad(x, y+2, false)

    x, y, _ = game.FindValidCityLocation(game.Plane)

    city2 := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city2.Population = 14000
    city2.Plane = data.PlaneArcanus
    city2.ProducingBuilding = buildinglib.BuildingHousing
    city2.ProducingUnit = units.UnitNone
    city2.AddBuilding(buildinglib.BuildingShrine)
    city2.AddBuilding(buildinglib.BuildingGranary)
    city2.Farmers = 10
    city2.Workers = 4
    city2.ResetCitizens()
    player.AddCity(city2)

    enemyWizard := setup.WizardCustom{
        Name: "enemy",
        Banner: data.BannerGreen,
        Race: data.RaceDraconian,
    }

    enemy := game.AddPlayer(enemyWizard, false)
    enemy.TaxRate = fraction.Make(1, 1)

    x, y, _ = game.FindValidCityLocation(game.Plane)

    city3 := citylib.MakeCity("Test City", x, y, enemy.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, enemy)
    city3.Population = 14000
    city3.Plane = data.PlaneArcanus
    city3.ProducingBuilding = buildinglib.BuildingHousing
    city3.ProducingUnit = units.UnitNone
    city3.AddBuilding(buildinglib.BuildingShrine)
    city3.AddBuilding(buildinglib.BuildingGranary)
    city3.AddBuilding(buildinglib.BuildingBank)
    city3.AddEnchantment(data.CityEnchantmentCloudOfShadow, enemy.GetBanner())
    city3.Farmers = 10
    city3.Workers = 4
    city3.ResetCitizens()
    enemy.AddCity(city3)

    city3.AddEnchantment(data.CityEnchantmentDeathWard, enemy.GetBanner())

    enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.DraconianSpearmen, x, y, data.PlaneArcanus, enemy.Wizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))
    enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.DraconianSpearmen, x, y - 1, data.PlaneArcanus, enemy.Wizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.DragonTurtle, x, y - 2, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.LiftFog(x, y, 2, data.PlaneArcanus)

    enemyWizard2 := setup.WizardCustom{
        Name: "enemy2",
        Banner: data.BannerPurple,
        Race: data.RaceBeastmen,
    }

    enemy2 := game.AddPlayer(enemyWizard2, false)
    enemy2.TaxRate = fraction.Make(1, 1)
    enemy2.Mana = 3000

    x2, y2, _ := game.FindValidCityLocation(game.Plane)

    city4 := citylib.MakeCity("Test City", x2, y2, enemy2.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, enemy2)
    city4.Population = 14000
    city4.Plane = data.PlaneArcanus
    city4.ProducingBuilding = buildinglib.BuildingHousing
    city4.ProducingUnit = units.UnitNone
    city4.AddBuilding(buildinglib.BuildingShrine)
    city4.AddBuilding(buildinglib.BuildingGranary)
    city4.AddBuilding(buildinglib.BuildingBank)
    city4.Farmers = 10
    city4.Workers = 4
    city4.ResetCitizens()
    city4.AddEnchantment(data.CityEnchantmentConsecration, enemy2.GetBanner())
    enemy2.AddCity(city4)
    enemy2.GlobalEnchantments.Insert(data.EnchantmentJustCause)
    enemy2.GlobalEnchantments.Insert(data.EnchantmentHerbMastery)
    enemy2.GlobalEnchantments.Insert(data.EnchantmentWindMastery)
    enemy2.GlobalEnchantments.Insert(data.EnchantmentArmageddon)

    player.LiftFog(x2, y2, 4, city4.Plane)
    player.AwarePlayer(enemy)
    player.AwarePlayer(enemy2)

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

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3

    city.ResetCitizens()

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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 20190
    city.Plane = data.PlaneArcanus
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

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    spear1 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    spear2 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    spear3 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
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

    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))
    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))

    return game
}

// show units with various levels of experience
func createScenario16(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 16")

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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.Fame = 40
    player.TaxRate = fraction.Zero()

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 10190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.Buildings.Insert(buildinglib.BuildingWizardsGuild)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = city.Citizens() - 1
    city.Workers = 0

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 1283
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    spear1 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    spear2 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    spear3 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    spear4 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    spear5 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.CastingSkillPower += 6000

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player.KnownSpells.AddSpell(allSpells.FindByName("Earth Lore"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Giant Strength"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Ice Bolt"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Enchant Item"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Create Artifact"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Magic Spirit"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Dark Rituals"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Enchant Road"))

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.AddBuilding(buildinglib.BuildingFortress)
    city.AddBuilding(buildinglib.BuildingShrine)
    city.AddBuilding(buildinglib.BuildingMarketplace)
    city.AddBuilding(buildinglib.BuildingWizardsGuild)
    city.AddBuilding(buildinglib.BuildingBank)
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3

    city.ResetCitizens()

    city.AddEnchantment(data.CityEnchantmentChaosRift, data.BannerGreen)

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 5000

    player.LiftFog(x, y, 4, data.PlaneArcanus)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    player.AddHeroToFortress(hero.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroRakir, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()), hero.HeroRakir, "Rakir"))
    player.AddHeroToFortress(hero.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroTorin, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()), hero.HeroTorin, "Torin"))
    player.AddHeroToFortress(hero.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroWarrax, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()), hero.HeroWarrax, "Warrax"))
    player.AddHeroToFortress(hero.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroRavashack, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()), hero.HeroRavashack, "Ravashack"))
    player.AddHeroToFortress(hero.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroSirHarold, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()), hero.HeroSirHarold, "Sir Harold"))

    alorra := hero.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroAlorra, x, y, data.PlaneMyrror, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()), hero.HeroAlorra, "Alorra")
    player.AddHeroToFortress(alorra)

    for _, xhero := range player.Heroes {
        if xhero != nil {
            xhero.Status = hero.StatusEmployed
        }
    }

    alorra.SetPlane(data.PlaneMyrror)

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
                Ability: data.ItemAbilityHaste,
                Name: "Haste",
            },
            /*
            {
                Type: artifact.PowerTypeAbility1,
                Amount: 2,
                Ability: data.ItemAbilityFlaming,
                Name: "Flaming",
            },
            */
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
        Player: player,
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
        Retorts: []data.Retort{
            data.RetortAlchemy,
            data.RetortSageMaster,
            data.RetortWarlord,
            data.RetortChanneler,
            data.RetortMyrran,
            data.RetortFamous,
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

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 8190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 5
    city.Workers = 3

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)
    player.LiftFog(20, 20, 100, data.PlaneMyrror)

    rakir := hero.MakeHero(units.MakeOverworldUnit(units.HeroRakir, x, y, data.PlaneArcanus), hero.HeroRakir, "bubba")
    player.AddHeroToFortress(rakir)
    rakir.AddExperience(528)
    rakir.Status = hero.StatusEmployed

    mysticX := hero.MakeHero(units.MakeOverworldUnit(units.HeroMysticX, x, y+1, data.PlaneArcanus), hero.HeroMysticX, "fred")
    player.AddHeroToFortress(mysticX)
    mysticX.SetExtraAbilities()
    mysticX.AddAbility(data.AbilityArmsmaster)
    mysticX.AddExperience(528)
    mysticX.Status = hero.StatusEmployed

    warlock := player.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, x, y, data.PlaneArcanus, player.GetBanner(), player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    // warlock.AddEnchantment(data.UnitEnchantmentGiantStrength)
    warlock.AddEnchantment(data.UnitEnchantmentLionHeart)

    stack := player.FindStackByUnit(mysticX)
    player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)

    enemy1 := game.AddPlayer(setup.WizardCustom{
        Name: "herby",
        Banner: data.BannerBlue,
    }, false)

    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))
    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))
    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))

    return game
}

// hire hero event
func createScenario19(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 19")

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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3

    city.ResetCitizens()

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

    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))
    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))

    game.Events <- &gamelib.GameEventHireHero{
        Player: player,
        Hero: player.HeroPool[hero.HeroRakir],
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.Fame = 5
    player.TaxRate = fraction.Zero()

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 10190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.Buildings.Insert(buildinglib.BuildingWizardsGuild)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = city.Citizens() - 1
    city.Workers = 0

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 1283
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    spear1 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    spear2 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    spear3 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    spear4 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    spear5 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
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
    enemy.TaxRate = fraction.Make(1, 1)

    enemy.Gold = 5000

    city2 := citylib.MakeCity("Test City", x, y-1, enemy.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, enemy)
    city2.Population = 14000
    city2.Plane = data.PlaneArcanus
    city2.ProducingBuilding = buildinglib.BuildingHousing
    city2.ProducingUnit = units.UnitNone
    city2.Farmers = city2.Citizens() - 1
    city2.Workers = 1
    city2.Buildings.Insert(buildinglib.BuildingSmithy)
    city2.Buildings.Insert(buildinglib.BuildingOracle)
    city2.Buildings.Insert(buildinglib.BuildingFortress)

    // cant use brown banner because neutral cities will never cast a city enchantment
    city2.AddEnchantment(data.CityEnchantmentWallOfFire, data.BannerRed)

    city2.Farmers = 5
    city2.Workers = 2
    city2.Rebels = 1

    city2.ResetCitizens()

    for range 1 {
        // randomUnit := units.ChooseRandomUnit(enemy.Wizard.Race)
        randomUnit := units.DraconianSettlers
        enemy.AddUnit(units.MakeOverworldUnitFromUnit(randomUnit, city2.X, city2.Y, data.PlaneArcanus, enemyWizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{LandSize: 0})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.Fame = 5
    player.TaxRate = fraction.Zero()

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 10190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.Buildings.Insert(buildinglib.BuildingWizardsGuild)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = city.Citizens() - 1
    city.Workers = 0

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 1283
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    spear1 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    spear2 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    spear3 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    spear4 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    spear5 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    spear2.AddExperience(30)
    spear3.AddExperience(60)
    spear4.AddExperience(100)
    spear5.AddExperience(200)

    stack := player.FindStackByUnit(spear1)
    player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 20, data.PlaneArcanus)
    player.Admin = true

    enemyWizard := setup.WizardCustom{
        Name: "enemy",
        Banner: data.BannerBrown,
        Race: data.RaceHighMen,
    }

    enemy := game.AddPlayer(enemyWizard, false)
    enemy.TaxRate = fraction.Make(1, 1)

    enemy.AIBehavior = ai.MakeRaiderAI()

    x2, y2 := game.FindValidCityLocationOnContinent(game.Plane, city.X, city.Y)
    log.Printf("enemy city at %v, %v", x2, y2)
    city2 := citylib.MakeCity("Test City", x2, y2, enemy.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, enemy)
    city2.Population = 8000
    city2.Plane = data.PlaneArcanus
    city2.ProducingBuilding = buildinglib.BuildingHousing
    city2.ProducingUnit = units.UnitNone
    city2.Farmers = city2.Citizens() - 1
    city2.Workers = 1
    city2.Buildings.Insert(buildinglib.BuildingSmithy)
    city2.Buildings.Insert(buildinglib.BuildingOracle)

    city2.ResetCitizens()

    city2.Farmers = 5
    city2.Workers = 2
    city2.Rebels = 1

    enemy.LiftFog(city2.X, city2.Y, 10, data.PlaneArcanus)

    /*
    for range 8 {
        randomUnit := units.ChooseRandomUnit(enemy.Wizard.Race)
        enemy.AddUnit(units.MakeOverworldUnitFromUnit(randomUnit, city2.X, city2.Y, data.PlaneArcanus, enemyWizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    // game.Map.CreateEncounter(x, y+1, maplib.EncounterTypeLair, game.Settings.Difficulty, false, game.Plane)
    game.CurrentMap().CreateEncounterRandom(x, y+1, game.Settings.Difficulty, game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6190
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

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    for i := 0; i < 1; i++ {
        fireElemental := player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{LandSize: 0})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.Fame = 5
    player.TaxRate = fraction.Zero()

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 10190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.Buildings.Insert(buildinglib.BuildingWizardsGuild)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = city.Citizens() - 1
    city.Workers = 0

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 1283
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    spear1 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    spear2 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    spear3 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    spear4 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    spear5 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
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

    // player.AwarePlayer(enemy2)
    // enemy2.AwarePlayer(player)
    player.WarWithPlayer(enemy2)
    enemy2.WarWithPlayer(player)

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
    enemy.TaxRate = fraction.Make(1, 1)

    x2, y2 := game.FindValidCityLocationOnContinent(game.Plane, city.X, city.Y)
    log.Printf("enemy city at %v, %v", x2, y2)
    city2 := citylib.MakeCity("Test City", x2, y2, enemy.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, enemy)
    city2.Population = 8000
    city2.Plane = data.PlaneArcanus
    city2.ProducingBuilding = buildinglib.BuildingHousing
    city2.ProducingUnit = units.UnitNone
    city2.Farmers = city2.Citizens() - 1
    city2.Workers = 1
    city2.Buildings.Insert(buildinglib.BuildingSmithy)
    city2.Buildings.Insert(buildinglib.BuildingOracle)

    city2.ResetCitizens()

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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 5
    city.Workers = 1

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 15772
    player.Mana = 26

    player.LiftFog(x, y, 3, data.PlaneArcanus)
    game.Camera.Center(x, y)

    enemy1 := game.AddPlayer(setup.WizardCustom{
        Name: "dingus",
        Banner: data.BannerRed,
    }, false)

    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, x + 1, y + 1, data.PlaneArcanus, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))
    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x + 1, y + 1, data.PlaneArcanus, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))

    game.Events <- &gamelib.GameEventHireMercenaries{
        Player: player,
        Units: []*units.OverworldUnit{
            units.MakeOverworldUnitFromUnit(units.BarbarianSpearmen, x, y, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()),
            units.MakeOverworldUnitFromUnit(units.BarbarianSpearmen, x, y, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()),
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
        Retorts: []data.Retort{
            data.RetortAlchemy,
            data.RetortSageMaster,
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

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 5
    city.Workers = 1
    city.ResetCitizens()
    player.AddCity(city)

    gunther := hero.MakeHero(units.MakeOverworldUnit(units.HeroGunther, 0, 0, data.PlaneArcanus), hero.HeroGunther, "Gunther")
    player.AddHeroToFortress(gunther)
    gunther.Status = hero.StatusEmployed

    enemy := game.AddPlayer(setup.WizardCustom{
        Name: "dingus",
        Banner: data.BannerRed,
    }, false)
    enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.KlackonSpearmen, x + 1, y + 1, data.PlaneArcanus, enemy.Wizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))

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
        Retorts: []data.Retort{
            data.RetortAlchemy,
            data.RetortSageMaster,
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

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 5
    city.Workers = 1
    city.ResetCitizens()
    player.AddCity(city)

    gunther := hero.MakeHero(units.MakeOverworldUnit(units.HeroGunther, 0, 0, data.PlaneArcanus), hero.HeroGunther, "Gunther")
    reywind := hero.MakeHero(units.MakeOverworldUnit(units.HeroReywind, 0, 0, data.PlaneArcanus), hero.HeroReywind, "Reywind")
    mysticX := hero.MakeHero(units.MakeOverworldUnit(units.HeroMysticX, 0, 0, data.PlaneArcanus), hero.HeroMysticX, "Mystic X")
    mysticX.SetExtraAbilities()
    player.AddHeroToFortress(mysticX)
    // player.AddHero(gunther)
    // player.AddHero(reywind)
    gunther.AddExperience(19)
    gunther.Status = hero.StatusEmployed
    reywind.AddExperience(58)
    reywind.Status = hero.StatusEmployed
    mysticX.AddExperience(19)
    mysticX.Status = hero.StatusEmployed

    enemy := game.AddPlayer(setup.WizardCustom{
        Name: "dingus",
        Banner: data.BannerRed,
    }, false)
    enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.KlackonSpearmen, x + 1, y + 1, data.PlaneArcanus, enemy.Wizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))

    return game
}

// multiple stacks on top of each other
func createScenario27(cache *lbx.LbxCache) *gamelib.Game {
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)
    player.TaxRate = fraction.Zero()

    x, y, _ := game.FindValidCityLocation(game.Plane)

    /*
    x = 20
    y = 20
    */

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 10000
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingGranary)
    city.Buildings.Insert(buildinglib.BuildingFarmersMarket)
    city.Buildings.Insert(buildinglib.BuildingForestersGuild)
    city.ProducingBuilding = buildinglib.BuildingNone
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 7
    city.Workers = 3

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    // log.Printf("City at %v, %v", x, y)

    player.LiftFog(x, y, 30, data.PlaneArcanus)

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    for i := 0; i < 2; i++ {
        fireElemental := player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
        _ = fireElemental
    }

    // player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, 30, 30, data.PlaneArcanus, wizard.Banner))

    stack := player.FindStackByUnit(drake)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x + 2, y + 2, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenCavalry, x + 2, y + 2, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    stack2 := player.FindStack(x + 2, y + 2, data.PlaneArcanus)
    stack2.Move(-1, -1, fraction.Zero(), game.GetNormalizeCoordinateFunc())

    // player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)

    enemy1 := game.AddPlayer(setup.WizardCustom{
        Name: "dingus",
        Banner: data.BannerRed,
    }, false)

    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, x + 3, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))
    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x + 3, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))

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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneMyrror

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 10190
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 10

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 26

    player.LiftFog(x, y, 4, game.Plane)

    nodes := findNodes(game.CurrentMap())
    node := nodes[terrain.SorceryNode][0]
    game.CurrentMap().RemoveEncounter(node.X, node.Y)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, node.X + 1, node.Y + 1, game.Plane, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    // has pathfinding
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.NomadRangers, node.X + 1, node.Y + 1, game.Plane, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneMyrror

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 12190
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Buildings.Insert(buildinglib.BuildingFarmersMarket)
    city.Race = wizard.Race
    city.Farmers = 9
    city.Workers = 3

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 26

    player.LiftFog(x, y, 4, city.Plane)

    nodes := findNodes(game.CurrentMap())
    node := nodes[terrain.SorceryNode][0]
    game.CurrentMap().RemoveEncounter(node.X, node.Y)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.OrcSwordsmen, node.X + 1, node.Y + 1, game.Plane, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.OrcEngineers, node.X + 1, node.Y + 1, game.Plane, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.OrcEngineers, node.X + 1, node.Y + 1, game.Plane, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 12190
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 9
    city.Workers = 3

    city.ResetCitizens()

    player.AddCity(city)

    x2, y2 := x + 3, y

    city2 := citylib.MakeCity("City2", x2, y2, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city2.Plane = city.Plane
    city2.Population = 6000
    city2.ResetCitizens()

    player.AddCity(city2)

    city3 := citylib.MakeCity("City3", x + 1, y2+2, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city3.Plane = city.Plane
    city3.Population = 6000
    city3.ResetCitizens()

    player.AddCity(city3)

    player.Gold = 83
    player.Mana = 26

    player.LiftFog(x, y, 8, game.Plane)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.OrcEngineers, x + 1, y + 1, game.Plane, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)
    player.StrategicCombat = true
    player.TaxRate = fraction.Zero()

    x, y, _ := game.FindValidCityLocation(game.Plane)

    /*
    x = 20
    y = 20
    */

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3
    city.Buildings.Insert(buildinglib.BuildingFortress)

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 2600

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    // log.Printf("City at %v, %v", x, y)

    player.LiftFog(x, y, 30, data.PlaneArcanus)

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    for i := 0; i < 4; i++ {
        player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
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
                    Hero: player.HeroPool[hero.HeroRakir],
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)
    player.TaxRate = fraction.Zero()

    x, y, _ := game.FindValidCityLocation(game.Plane)

    /*
    x = 20
    y = 20
    */

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3
    city.Buildings.Insert(buildinglib.BuildingFortress)

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 2600

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    // log.Printf("City at %v, %v", x, y)

    player.LiftFog(x, y, 30, data.PlaneArcanus)

    gunther := hero.MakeHero(units.MakeOverworldUnit(units.HeroGunther, 0, 0, data.PlaneArcanus), hero.HeroGunther, "Gunther")
    gunther.Status = hero.StatusEmployed
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
    player.AddHeroToFortress(gunther)

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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)
    player.TaxRate = fraction.Zero()

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = buildinglib.BuildingTradeGoods
    city.ProducingUnit = units.UnitNone
    city.Buildings.Insert(buildinglib.BuildingGranary)
    city.Buildings.Insert(buildinglib.BuildingFarmersMarket)
    city.Buildings.Insert(buildinglib.BuildingForestersGuild)
    city.Race = wizard.Race
    city.Farmers = 13
    city.Workers = 3

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 2600

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    // log.Printf("City at %v, %v", x, y)

    player.LiftFog(x, y, 30, data.PlaneArcanus)

    priest1 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.BeastmenPriest, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    priest2 := player.AddUnit(units.MakeOverworldUnitFromUnit(units.BeastmenPriest, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

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

    city := citylib.MakeCity("", x, y, data.RaceBarbarian, game.BuildingInfo, game.CurrentMap(), game, player1)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard1.Race
    city.Farmers = 3
    city.Workers = 3

    city.ResetCitizens()

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

    spirit := player1.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, node.X + 1, node.Y + 1, data.PlaneArcanus, wizard1.Banner, player1.MakeExperienceInfo(), player1.MakeUnitEnchantmentProvider()))
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

    city := citylib.MakeCity("", x, y, data.RaceBarbarian, game.BuildingInfo, game.CurrentMap(), game, player1)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard1.Race
    city.Farmers = 3
    city.Workers = 3

    city.ResetCitizens()

    player1.AddCity(city)

    player1.Gold = 830
    player1.Mana = 26557
    player1.CastingSkillPower = 10000

    player1.LiftFog(x, y, 3, data.PlaneArcanus)

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)
    for _, spell := range allSpells.Spells {
        player1.KnownSpells.AddSpell(spell)
    }

    spirit := player1.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x + 1, y + 1, data.PlaneArcanus, wizard1.Banner, player1.MakeExperienceInfo(), player1.MakeUnitEnchantmentProvider()))
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
    human.Mana = 10000

    human.GlobalEnchantments.Insert(data.EnchantmentNatureAwareness)

    wizard1 := setup.WizardCustom{
        Name: "Rjak",
        Base: data.WizardRjak,
        Banner: data.BannerPurple,
        Race: data.RaceBarbarian,
        Books: []data.WizardBook{
            data.WizardBook{
                Magic: data.LifeMagic,
                Count: 3,
            },
        },
    }

    enemy1 := game.AddPlayer(wizard1, false)

    enemy1.TaxRate = fraction.Zero()

    enemy1.AIBehavior = ai.MakeEnemyAI()

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("ai1", x, y, data.RaceBarbarian, game.BuildingInfo, game.CurrentMap(), game, enemy1)
    city.Population = 10190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingTradeGoods
    city.ProducingUnit = units.UnitNone
    city.Farmers = 8
    city.Workers = 3

    city.ResetCitizens()

    enemy1.AddCity(city)

    enemy1.Gold = 830
    enemy1.Mana = 26557
    enemy1.CastingSkillPower = 10000

    enemy1.LiftFog(x, y, 3, data.PlaneArcanus)

    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.BarbarianSettlers, x, y, data.PlaneArcanus, wizard1.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))

    // allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    human.LiftFog(20, 20, 100, data.PlaneArcanus)

    game.Camera.Center(x, y)

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
    player.Fame = 10
    player.LiftFog(node.X, node.Y, 3, data.PlaneArcanus)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.Cockatrices, node.X, node.Y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    unit := player.AddUnit(units.MakeOverworldUnitFromUnit(units.SkyDrake, node.X + 1, node.Y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
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

    city1 := citylib.MakeCity("", x, y, data.RaceBarbarian, game.BuildingInfo, game.CurrentMap(), game, player1)
    city1.Population = 6190
    city1.Plane = data.PlaneArcanus
    city1.ProducingBuilding = buildinglib.BuildingGranary
    city1.ProducingUnit = units.UnitNone
    city1.Race = wizard1.Race
    city1.Farmers = 3
    city1.Workers = 3
    city1.ResetCitizens()
    player1.AddCity(city1)
    player1.LiftFog(x, y, 3, data.PlaneArcanus)

    x, y, _ = game.FindValidCityLocation(game.Plane)
    player1.LiftFog(x, y, 3, data.PlaneArcanus)

    unit1 := player1.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x, y, data.PlaneArcanus, wizard1.Banner, player1.MakeExperienceInfo(), player1.MakeUnitEnchantmentProvider()))
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

    player2.AddUnit(units.MakeOverworldUnitFromUnit(units.Basilisk, x+1, y-2, data.PlaneArcanus, wizard2.Banner, player2.MakeExperienceInfo(), player2.MakeUnitEnchantmentProvider()))
    player2.AddUnit(units.MakeOverworldUnitFromUnit(units.Basilisk, x+1, y-1, data.PlaneArcanus, wizard2.Banner, player2.MakeExperienceInfo(), player2.MakeUnitEnchantmentProvider()))
    player2.AddUnit(units.MakeOverworldUnitFromUnit(units.Basilisk, x+1, y, data.PlaneArcanus, wizard2.Banner, player2.MakeExperienceInfo(), player2.MakeUnitEnchantmentProvider()))
    player2.AddUnit(units.MakeOverworldUnitFromUnit(units.Basilisk, x-1, y-2, data.PlaneArcanus, wizard2.Banner, player2.MakeExperienceInfo(), player2.MakeUnitEnchantmentProvider()))
    player2.AddUnit(units.MakeOverworldUnitFromUnit(units.Basilisk, x-1, y-1, data.PlaneArcanus, wizard2.Banner, player2.MakeExperienceInfo(), player2.MakeUnitEnchantmentProvider()))
    player2.AddUnit(units.MakeOverworldUnitFromUnit(units.Basilisk, x-1, y, data.PlaneArcanus, wizard2.Banner, player2.MakeExperienceInfo(), player2.MakeUnitEnchantmentProvider()))

    return game
}

// town doesn't produce enough food
func createScenario39(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 39")

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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)
    player.TaxRate = fraction.Make(3, 1)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    /*
    x = 20
    y = 20
    */

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6300
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.AddEnchantment(data.CityEnchantmentFamine, player.GetBanner())
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3

    city.ResetCitizens()

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

// cast global enchantment
func createScenario40(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 40")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.CastingSkillPower += 500000

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3

    city.AddBuilding(buildinglib.BuildingShrine)
    city.AddBuilding(buildinglib.BuildingGranary)

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 1000

    player.LiftFog(x, y, 4, data.PlaneArcanus)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    game.CurrentMap().SetRoad(x, y+1, false)
    game.CurrentMap().SetRoad(x, y+2, false)

    game.Camera.Center(x, y)

    game.Events <- &gamelib.GameEventCastGlobalEnchantment{
        Player: player,
        // Enchantment: data.EnchantmentNatureAwareness,
        // Enchantment: data.EnchantmentZombieMastery,
        Enchantment: data.EnchantmentPlanarSeal,
        // Enchantment: data.EnchantmentTimeStop,
    }

    return game
}

// shrink city to zero
func createScenario41(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 41")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
        Race: data.RaceTroll,
        Retorts: []data.Retort{},
        Books: []data.WizardBook{},
    }

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Erfurt", x, y, player.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 1000
    city.Plane = data.PlaneArcanus
    city.ProducingUnit = units.TrollSettlers
    city.Production = float32(city.ProducingUnit.ProductionCost)
    city.Farmers = 1
    city.ResetCitizens()

    player.AddCity(city)
    player.Gold = 1000
    player.Mana = 1000

    player.LiftFog(x, y, 4, data.PlaneArcanus)

    game.Camera.Center(x, y)

    return game
}

// water unit
func createScenario42(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 42")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.CastingSkillPower += 500000

    x, y, _ := game.FindValidCityLocationOnShore(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 4
    city.Workers = 2

    city.AddBuilding(buildinglib.BuildingShrine)
    city.AddBuilding(buildinglib.BuildingGranary)
    city.AddBuilding(buildinglib.BuildingOracle)

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 1000

    player.LiftFog(x, y, 200, data.PlaneArcanus)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighElfSwordsmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighElfSwordsmen, x+1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    warship := player.AddUnit(units.MakeOverworldUnitFromUnit(units.Warship, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    warship.AddEnchantment(data.UnitEnchantmentFlight)
    // player.AddUnit(units.MakeOverworldUnitFromUnit(units.Galley, x-1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    enemy1 := game.AddPlayer(setup.WizardCustom{
        Name: "dingus",
        Banner: data.BannerGreen,
    }, false)

    enemy1.AIBehavior = nil

    ex, ey, _ := game.FindValidCityLocation(game.Plane)

    city2 := citylib.MakeCity("Test City", ex, ey, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, enemy1)
    city2.Population = 20000
    city2.Plane = data.PlaneArcanus
    city2.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city2.ProducingBuilding = buildinglib.BuildingGranary
    city2.ProducingUnit = units.UnitNone
    city2.Race = wizard.Race
    city2.Farmers = 17
    city2.Workers = 3

    city2.AddBuilding(buildinglib.BuildingShrine)
    city2.AddBuilding(buildinglib.BuildingGranary)
    city2.AddBuilding(buildinglib.BuildingFarmersMarket)

    city2.ResetCitizens()

    enemy1.AddCity(city2)

    enemy1.Gold = 1000

    // add a water unit to test water combat
    mapUse := game.GetMap(data.PlaneArcanus)
    var candidates []image.Point
    for cx := range mapUse.Width() {
        for cy := range mapUse.Height() {
            if mapUse.GetTile(cx, cy).Tile.TerrainType() == terrain.Ocean && mapUse.TileDistance(cx, cy, x, y) < 8 {
                candidates = append(candidates, image.Pt(cx, cy))
            }
        }
    }

    if len(candidates) == 0 {
        log.Printf("No water tiles found")
        return game
    }

    use := candidates[rand.N(len(candidates))]
    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.Warship, use.X, use.Y, data.PlaneArcanus, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))
    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.Warship, use.X, use.Y, data.PlaneArcanus, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))

    game.Camera.Center(x, y)

    return game
}

// show random event
func createScenario43(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 43")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
        Race: data.RaceTroll,
        Retorts: []data.Retort{},
        Books: []data.WizardBook{
            data.WizardBook{
                Magic: data.DeathMagic,
                Count: 5,
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

    city := citylib.MakeCity("Erfurt", x, y, player.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 10000
    city.Plane = data.PlaneArcanus
    city.ProducingUnit = units.UnitNone
    city.Production = float32(city.ProducingUnit.ProductionCost)
    city.AddBuilding(buildinglib.BuildingFortress)
    city.AddBuilding(buildinglib.BuildingShrine)
    city.AddBuilding(buildinglib.BuildingTemple)
    city.Farmers = 1
    city.ResetCitizens()

    player.AddCity(city)
    player.Gold = 1000
    player.Mana = 1000

    player.LiftFog(x, y, 4, data.PlaneArcanus)

    game.Camera.Center(x, y)

    game.TurnNumber = 300

    game.Events <- &gamelib.GameEventShowRandomEvent{
        // Event: gamelib.MakeDisjunctionEvent(0),
        // Event: gamelib.MakeBadMoonEvent(0),
        // Event: gamelib.MakeConjunctionChaosEvent(0),
        // Event: gamelib.MakeConjunctionNatureEvent(0),
        // Event: gamelib.MakeConjunctionSorceryEvent(0),
        // Event: gamelib.MakeDepletionEvent(0, data.BonusMithrilOre, "nyc"),
        // Event: gamelib.MakeDiplomaticMarriageEvent(0, city),
        // Event: gamelib.MakeDonationEvent(0, 123),
        // Event: gamelib.MakeEarthquakeEvent(0, "nyc", 4, 8, 12),
        // Event: gamelib.MakeGiftEvent(0, "frisbee"),
        // Event: gamelib.MakeGoodMoonEvent(0),
        // Event: gamelib.MakeGreatMeteorEvent(0, "nyc", 2, 4, 6),
        // Event: gamelib.MakeManaShortEvent(0),
        // Event: gamelib.MakeNewMineralsEvent(0, data.BonusMithrilOre, city),
        // Event: gamelib.MakePiracyEvent(0, 456),
        // Event: gamelib.MakePlagueEvent(0, city),
        // Event: gamelib.MakePopulationBoomEvent(0, city),
        Event: gamelib.MakeRebellionEvent(0, city),
        Starting: true,
    }

    return game
}

// astral gate
func createScenario44(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 44: astral gate/plane shift")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.CastingSkillPower += 500000
    player.TaxRate = fraction.Zero()

    x, y, _ := game.FindValidCityLocationOnShore(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 16
    city.Workers = 0

    city.AddBuilding(buildinglib.BuildingShrine)
    city.AddBuilding(buildinglib.BuildingTemple)
    city.AddBuilding(buildinglib.BuildingParthenon)
    city.AddBuilding(buildinglib.BuildingGranary)
    city.AddBuilding(buildinglib.BuildingOracle)
    city.AddBuilding(buildinglib.BuildingFarmersMarket)

    city.AddEnchantment(data.CityEnchantmentAstralGate, player.GetBanner())

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 1000

    player.LiftFog(x, y, 200, data.PlaneArcanus)

    /*
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighElfSwordsmen, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighElfSwordsmen, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    */
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.Angel, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.ShadowDemons, x-1, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    city.ResetCitizens()

    enemy1 := game.AddPlayer(setup.WizardCustom{
        Name: "dingus",
        Banner: data.BannerGreen,
    }, false)

    enemy1.AIBehavior = nil

    // add an enemy at the same spot but on the opposite plane
    // enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x, y, data.PlaneMyrror, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))

    game.Camera.Center(x, y)

    return game
}

// combat in enemy city
func createScenario45(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 45")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.CastingSkillPower += 500000
    player.TaxRate = fraction.Zero()

    x, y, _ := game.FindValidCityLocationOnShore(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 16
    city.Workers = 0

    city.AddBuilding(buildinglib.BuildingShrine)
    city.AddBuilding(buildinglib.BuildingTemple)
    city.AddBuilding(buildinglib.BuildingParthenon)
    city.AddBuilding(buildinglib.BuildingGranary)
    city.AddBuilding(buildinglib.BuildingOracle)
    city.AddBuilding(buildinglib.BuildingFarmersMarket)

    city.AddEnchantment(data.CityEnchantmentAstralGate, player.GetBanner())

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 1000

    player.LiftFog(x, y, 200, data.PlaneArcanus)

    city.ResetCitizens()

    enemy1 := game.AddPlayer(setup.WizardCustom{
        Name: "dingus",
        Banner: data.BannerGreen,
        Race: data.RaceHighElf,
    }, false)

    enemy1.AIBehavior = nil
    enemy1.Gold = 10000
    enemy1.Mana = 10000

    x, y, _ = game.FindValidCityLocation(game.Plane)

    city2 := citylib.MakeCity("enemy city", x, y, enemy1.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, enemy1)
    city2.Population = 16190
    city2.Plane = data.PlaneArcanus
    city2.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city2.ProducingBuilding = buildinglib.BuildingGranary
    city2.ProducingUnit = units.UnitNone
    city2.Race = wizard.Race
    city2.Farmers = 16
    city2.Workers = 0

    city2.AddBuilding(buildinglib.BuildingShrine)
    city2.AddBuilding(buildinglib.BuildingTemple)
    city2.AddBuilding(buildinglib.BuildingParthenon)
    city2.AddBuilding(buildinglib.BuildingGranary)
    city2.AddBuilding(buildinglib.BuildingOracle)
    city2.AddBuilding(buildinglib.BuildingFarmersMarket)
    enemy1.AddCity(city2)

    // add an enemy at the same spot but on the opposite plane
    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x, y, data.PlaneArcanus, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x, y+1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x, y+1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    game.Camera.Center(x, y)

    return game
}

// building units with experience
func createScenario46(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 46")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
        Race: data.RaceTroll,
        Retorts: []data.Retort{
            data.RetortWarlord,
        },
        Books: []data.WizardBook{
            {
                Magic: data.LifeMagic,
                Count:11,
            },
        },
    }

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.CastingSkillPower += 500000

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player.KnownSpells.AddSpell(allSpells.FindByName("Guardian Spirit"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Altar of Battle"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Crusade"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Heroism"))

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingNone
    city.ProducingUnit = units.TrollSpearmen
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3

    city.AddBuilding(buildinglib.BuildingFortress)
    city.AddBuilding(buildinglib.BuildingFightersGuild)
    city.AddBuilding(buildinglib.BuildingWarCollege)

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 1000
    player.Mana = 10000

    player.LiftFog(x, y, 4, data.PlaneArcanus)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    return game
}

// disenchanting things
func createScenario47(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 47")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.CastingSkillPower += 500000

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)
    player.KnownSpells.AddSpell(allSpells.FindByName("Disenchant Area"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Disenchant True"))

    game.Events <- &gamelib.GameEventCastSpell{Player: player, Spell: allSpells.FindByName("Nature Awareness")}

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3

    city.AddBuilding(buildinglib.BuildingFortress)
    city.AddBuilding(buildinglib.BuildingShrine)
    city.AddBuilding(buildinglib.BuildingGranary)

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 1000
    player.Mana = 10000

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, player.GetBanner(), player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    drake.AddEnchantment(data.UnitEnchantmentGiantStrength)

    player.LiftFog(x, y, 4, data.PlaneArcanus)

    enemyWizard := setup.WizardCustom{
        Name: "enemy",
        Banner: data.BannerGreen,
        Race: data.RaceDraconian,
    }

    enemy := game.AddPlayer(enemyWizard, false)
    enemy.TaxRate = fraction.Make(1, 1)

    x, y, _ = game.FindValidCityLocation(game.Plane)

    city2 := citylib.MakeCity("Test City", x, y, enemy.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, enemy)
    city2.Population = 14000
    city2.Plane = data.PlaneArcanus
    city2.ProducingBuilding = buildinglib.BuildingHousing
    city2.ProducingUnit = units.UnitNone
    city2.AddBuilding(buildinglib.BuildingShrine)
    city2.AddBuilding(buildinglib.BuildingGranary)
    city2.AddBuilding(buildinglib.BuildingBank)
    city2.Farmers = 10
    city2.Workers = 4
    city2.ResetCitizens()
    enemy.AddCity(city2)

    enemy1Unit := enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x + 1, y + 1, data.PlaneArcanus, enemy.GetBanner(), enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))
    enemy1Unit.AddEnchantment(data.UnitEnchantmentGiantStrength)
    enemy1Unit.AddEnchantment(data.UnitEnchantmentLionHeart)
    enemy1Unit.AddEnchantment(data.UnitEnchantmentHaste)

    enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.DraconianSpearmen, x, y, data.PlaneArcanus, enemy.Wizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))
    enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.DraconianSpearmen, x + 2, y + 1, data.PlaneArcanus, enemy.Wizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.DragonTurtle, x + 2, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.LiftFog(x, y, 2, data.PlaneArcanus)

    enemyWizard2 := setup.WizardCustom{
        Name: "enemy2",
        Banner: data.BannerPurple,
        Race: data.RaceBeastmen,
    }

    enemy2 := game.AddPlayer(enemyWizard2, false)
    enemy2.TaxRate = fraction.Make(1, 1)

    x2, y2, _ := game.FindValidCityLocation(game.Plane)

    city3 := citylib.MakeCity("Test City", x2, y2, enemy2.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, enemy2)
    city3.Population = 14000
    city3.Plane = data.PlaneArcanus
    city3.ProducingBuilding = buildinglib.BuildingHousing
    city3.ProducingUnit = units.UnitNone
    city3.AddBuilding(buildinglib.BuildingShrine)
    city3.AddBuilding(buildinglib.BuildingGranary)
    city3.AddBuilding(buildinglib.BuildingBank)
    city3.Farmers = 10
    city3.Workers = 4
    city3.ResetCitizens()
    city3.AddEnchantment(data.CityEnchantmentConsecration, enemy2.GetBanner())
    enemy2.AddCity(city3)

    player.LiftFog(x2, y2, 4, city3.Plane)

    city.AddEnchantment(data.CityEnchantmentConsecration, player.GetBanner())
    city.AddEnchantment(data.CityEnchantmentFamine, enemy.GetBanner())

    // force a node to be melded and warped
    var node *maplib.ExtraMagicNode
    mapUse := game.GetMap(data.PlaneArcanus)
    nodeLook:
    for x := range mapUse.Width() {
        for y := range mapUse.Height() {
            if mapUse.GetMagicNode(x, y) != nil {
                node = mapUse.GetMagicNode(x, y)
                break nodeLook
            }
        }
    }

    if node != nil {
        node.Meld(enemy, units.MagicSpirit)
        node.Warped = true
        node.WarpedOwner = enemy
    }

    return game
}

// overland cast, but enemy has Tranquility and Life Force active
func createScenario48(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 48: enemy Tranquility")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.CastingSkillPower += 500000

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    // summoning
    player.KnownSpells.AddSpell(allSpells.FindByName("Storm Giant"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Cockatrices"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Basilisk"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Stone Giant"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Gorgons"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Behemoth"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Colossus"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Great Wyrm"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Earth Elemental"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Skeletons"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Ghouls"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Night Stalker"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Shadow Demons"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Wraiths"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Death Knights"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Demon Lord"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Hell Hounds"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Fire Giant"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Gargoyles"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Doom Bat"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Chimeras"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Chaos Spawn"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Efreet"))

    // special spells
    player.KnownSpells.AddSpell(allSpells.FindByName("Earth Lore"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Giant Strength"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Ice Bolt"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Enchant Item"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Create Artifact"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Enchant Road"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Raise Volcano"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Corruption"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Warp Node"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Call The Void"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Change Terrain"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Transmute"))

    // city spells
    player.KnownSpells.AddSpell(allSpells.FindByName("Wall of Fire"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Wall of Darkness"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Wall of Stone"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Nature's Eye"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Cursed Lands"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Famine"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Pestilence"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Astral Gate"))

    // global enchantments
    player.KnownSpells.AddSpell(allSpells.FindByName("Nature Awareness"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Awareness"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Crusade"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Just Cause"))

    // unit enchantments
    player.KnownSpells.AddSpell(allSpells.FindByName("Bless"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Heroism"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Chaos Channels"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Invulnerability"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Planar Travel"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Iron Skin"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Path Finding"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Regeneration"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Stone Skin"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Water Walking"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Guardian Wind"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Invisiblity"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Magic Immunity"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Spell Lock"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Wind Walking"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Eldritch Weapon"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Flame Blade"))

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3

    city.AddBuilding(buildinglib.BuildingFortress)
    city.AddBuilding(buildinglib.BuildingShrine)
    city.AddBuilding(buildinglib.BuildingGranary)

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 1000
    player.Mana = 10000

    player.LiftFog(x, y, 4, data.PlaneArcanus)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    game.CurrentMap().SetRoad(x, y+1, false)
    game.CurrentMap().SetRoad(x, y+2, false)

    enemyWizard := setup.WizardCustom{
        Name: "enemy",
        Banner: data.BannerGreen,
        Race: data.RaceDraconian,
    }

    enemy := game.AddPlayer(enemyWizard, false)
    enemy.TaxRate = fraction.Make(1, 1)
    enemy.GlobalEnchantments.Insert(data.EnchantmentTranquility)
    enemy.GlobalEnchantments.Insert(data.EnchantmentLifeForce)
    enemy.Mana += 10000

    x, y, _ = game.FindValidCityLocation(game.Plane)

    city2 := citylib.MakeCity("Test City", x, y, enemy.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, enemy)
    city2.Population = 14000
    city2.Plane = data.PlaneArcanus
    city2.ProducingBuilding = buildinglib.BuildingHousing
    city2.ProducingUnit = units.UnitNone
    city2.AddBuilding(buildinglib.BuildingShrine)
    city2.AddBuilding(buildinglib.BuildingGranary)
    city2.AddBuilding(buildinglib.BuildingBank)
    city2.Farmers = 10
    city2.Workers = 4
    city2.ResetCitizens()
    enemy.AddCity(city2)

    enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.DraconianSpearmen, x, y, data.PlaneArcanus, enemy.Wizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))
    enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.DraconianSpearmen, x + 2, y + 1, data.PlaneArcanus, enemy.Wizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.DragonTurtle, x + 2, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.LiftFog(x, y, 2, data.PlaneArcanus)

    enemyWizard2 := setup.WizardCustom{
        Name: "enemy2",
        Banner: data.BannerPurple,
        Race: data.RaceBeastmen,
    }

    enemy2 := game.AddPlayer(enemyWizard2, false)
    enemy2.TaxRate = fraction.Make(1, 1)

    x2, y2, _ := game.FindValidCityLocation(game.Plane)

    city3 := citylib.MakeCity("Test City", x2, y2, enemy2.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, enemy2)
    city3.Population = 14000
    city3.Plane = data.PlaneArcanus
    city3.ProducingBuilding = buildinglib.BuildingHousing
    city3.ProducingUnit = units.UnitNone
    city3.AddBuilding(buildinglib.BuildingShrine)
    city3.AddBuilding(buildinglib.BuildingGranary)
    city3.AddBuilding(buildinglib.BuildingBank)
    city3.Farmers = 10
    city3.Workers = 4
    city3.ResetCitizens()
    city3.AddEnchantment(data.CityEnchantmentConsecration, enemy2.GetBanner())
    enemy2.AddCity(city3)

    player.LiftFog(x2, y2, 4, city3.Plane)

    return game
}

// detect magic
func createScenario49(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 49")

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })
    game.Plane = data.PlaneArcanus

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player := game.AddPlayer(
        setup.WizardCustom{
            Name: "Merlin",
            Base: data.WizardMerlin,
            Banner: data.BannerRed,
            Race: data.RaceHalfling,
        },
        true,
    )

    player.CastingSkillPower += 500000
    player.KnownSpells.AddSpell(allSpells.FindByName("Detect Magic"))
    player.ResearchingSpell = allSpells.FindByName("Spell of Mastery")
    player.Gold = 1000
    player.Mana = 10000

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("City 1", x, y, player.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingWizardsGuild
    city.ProducingUnit = units.UnitNone
    city.Race = player.Wizard.Race
    city.Farmers = 3
    city.Workers = 3
    city.ResetCitizens()

    player.AddCity(city)
    player.LiftFog(x, y, 4, data.PlaneArcanus)
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x + 1, y + 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    enemy := game.AddPlayer(
        setup.WizardCustom{
            Name: "Horus",
            Base: data.WizardHorus,
            Banner: data.BannerYellow,
            Race: data.RaceHalfling,
        },
        false,
    )
    enemy.CastingSkillPower += 500000
    enemy.CastingSpell = allSpells.FindByName("Guardian Spirit")
    enemy.Gold = 1000
    enemy.Mana = 10000

    x, y, _ = game.FindValidCityLocation(game.Plane)

    city = citylib.MakeCity("City 2", x, y, enemy.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, enemy)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = enemy.Wizard.Race
    city.Farmers = 3
    city.Workers = 3
    city.ResetCitizens()

    enemy.AddCity(city)

    player.LiftFog(x, y, 4, data.PlaneArcanus)
    player.AwarePlayer(enemy)

    enemy = game.AddPlayer(
        setup.WizardCustom{
            Name: "Freya",
            Base: data.WizardFreya,
            Banner: data.BannerGreen,
            Race: data.RaceHalfling,
        },
        false,
    )
    enemy.CastingSkillPower += 500000
    enemy.CastingSpell = allSpells.FindByName("Awareness")
    enemy.Gold = 1000
    enemy.Mana = 10000

    x, y, _ = game.FindValidCityLocation(game.Plane)

    city = citylib.MakeCity("City 3", x, y, enemy.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, enemy)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = enemy.Wizard.Race
    city.Farmers = 3
    city.Workers = 3
    city.ResetCitizens()

    enemy.AddCity(city)

    player.LiftFog(x, y, 4, data.PlaneArcanus)
    player.AwarePlayer(enemy)

    return game
}

// Wizard-targeting spells (Spell Blast, Cruel Unminding, Drain Power) test
func createScenario50(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 49: Spell Blast")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.CastingSkillPower += 500000

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player.KnownSpells.AddSpell(allSpells.FindByName("Spell Blast"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Cruel Unminding"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Drain Power"))

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 6190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 3
    city.Workers = 3

    city.AddBuilding(buildinglib.BuildingFortress)
    city.AddBuilding(buildinglib.BuildingShrine)
    city.AddBuilding(buildinglib.BuildingGranary)

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 1000
    player.Mana = 10000

    player.LiftFog(x, y, 4, data.PlaneArcanus)

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    game.CurrentMap().SetRoad(x, y+1, false)
    game.CurrentMap().SetRoad(x, y+2, false)

    enemyWizard := setup.WizardCustom{
        Name: "enemy",
        Banner: data.BannerGreen,
        Race: data.RaceDraconian,
    }

    enemy := game.AddPlayer(enemyWizard, false)
    enemy.TaxRate = fraction.Make(1, 1)
    enemy.Mana += 10000
    enemy.CastingSkillPower = (100 * 100 - 100 + 1) // skill^2 - skill + 1, simplification of (skill-1)^2 + skill
    enemy.CastingSpell = allSpells.FindByName("Armageddon")
    enemy.CastingSpellProgress = 100

    x, y, _ = game.FindValidCityLocation(game.Plane)

    city2 := citylib.MakeCity("Test City", x, y, enemy.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, enemy)
    city2.Population = 14000
    city2.Plane = data.PlaneArcanus
    city2.ProducingBuilding = buildinglib.BuildingHousing
    city2.ProducingUnit = units.UnitNone
    city2.AddBuilding(buildinglib.BuildingShrine)
    city2.AddBuilding(buildinglib.BuildingGranary)
    city2.AddBuilding(buildinglib.BuildingBank)
    city2.Farmers = 10
    city2.Workers = 4
    city2.ResetCitizens()
    enemy.AddCity(city2)

    enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.DraconianSpearmen, x, y, data.PlaneArcanus, enemy.Wizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))
    enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.DraconianSpearmen, x + 2, y + 1, data.PlaneArcanus, enemy.Wizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.DragonTurtle, x + 2, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.LiftFog(x, y, 2, data.PlaneArcanus)

    enemyWizard2 := setup.WizardCustom{
        Name: "enemy2",
        Banner: data.BannerPurple,
        Race: data.RaceBeastmen,
    }

    enemy2 := game.AddPlayer(enemyWizard2, false)
    enemy2.TaxRate = fraction.Make(1, 1)

    x2, y2, _ := game.FindValidCityLocation(game.Plane)

    city3 := citylib.MakeCity("Test City", x2, y2, enemy2.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, enemy2)
    city3.Population = 14000
    city3.Plane = data.PlaneArcanus
    city3.ProducingBuilding = buildinglib.BuildingHousing
    city3.ProducingUnit = units.UnitNone
    city3.AddBuilding(buildinglib.BuildingShrine)
    city3.AddBuilding(buildinglib.BuildingGranary)
    city3.AddBuilding(buildinglib.BuildingBank)
    city3.Farmers = 10
    city3.Workers = 4
    city3.ResetCitizens()
    city3.AddEnchantment(data.CityEnchantmentConsecration, enemy2.GetBanner())
    enemy2.AddCity(city3)

    player.LiftFog(x2, y2, 4, city3.Plane)
    player.AwarePlayer(enemy)
    player.AwarePlayer(enemy2)
    enemy.AwarePlayer(player)

    relation, _ := enemy.GetDiplomaticRelation(player)
    relation.VisibleRelation = -40

    return game
}

// Relocate units
func createScenario51_52(cache *lbx.LbxCache, kill bool) *gamelib.Game {
    scenario := 51
    if kill {
        scenario = 52
    }
    log.Printf("Running scenario %v: relocate units", scenario)
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
        Race: data.RaceHalfling,
        Retorts: []data.Retort{},
        Books: []data.WizardBook{
            {
                Magic: data.SorceryMagic,
                Count: 11,
            },
        },
    }

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })
    game.Plane = data.PlaneArcanus

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player := game.AddPlayer(wizard, true)
    player.CastingSkillPower += 500000
    player.Gold = 1000
    player.Mana = 10000
    player.KnownSpells.AddSpell(allSpells.FindByName("Magic Spirit"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Summon Hero"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Recall Hero"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Word of Recall"))

    x, y, _ := game.FindValidCityLocationOnShore(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHalfling, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingNone
    city.ProducingUnit = units.HalflingSpearmen
    city.Race = wizard.Race
    city.Farmers = 13
    city.Workers = 3
    city.AddBuilding(buildinglib.BuildingFortress)
    city.AddBuilding(buildinglib.BuildingGranary)
    city.ResetCitizens()

    player.AddCity(city)
    player.LiftFog(x, y, 4, data.PlaneArcanus)
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.Slingers, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.HalflingSpearmen, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.HalflingShamans, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.MagicSpirit, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.Nagas, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.SkyDrake, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.HeroPool[hero.HeroRakir].AddExperience(0)
    player.HeroPool[hero.HeroShinBo].AddExperience(100)
    player.HeroPool[hero.HeroAerie].AddExperience(200)
    player.HeroPool[hero.HeroBShan].AddExperience(300)
    player.AddHeroToFortress(player.HeroPool[hero.HeroRakir])
    player.AddHeroToFortress(player.HeroPool[hero.HeroAerie])
    player.AddHeroToFortress(player.HeroPool[hero.HeroBShan])

    game.Events <- &gamelib.GameEventHireHero{
        Player: player,
        Hero: player.HeroPool[hero.HeroShinBo],
        Cost: 200,
    }

    game.Events <- &gamelib.GameEventHireMercenaries{
        Player: player,
        Units: []*units.OverworldUnit{
            units.MakeOverworldUnitFromUnit(units.Berserkers, x, y, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()),
            units.MakeOverworldUnitFromUnit(units.Berserkers, x, y, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()),
        },
        Cost: 200,
    }

    enemyWizard := setup.WizardCustom{
        Name: "enemy",
        Banner: data.BannerGreen,
        Race: data.RaceDraconian,
    }
    enemy := game.AddPlayer(enemyWizard, false)
    x1 := game.ArcanusMap.WrapX(x-1)
    x2 := game.ArcanusMap.WrapX(x+1)
    enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.DraconianSpearmen, x1, y-1, data.PlaneArcanus, enemyWizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))
    enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.DraconianSpearmen, x1, y, data.PlaneArcanus, enemyWizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))
    enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.DraconianSpearmen, x1, y+1, data.PlaneArcanus, enemyWizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))
    enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.DraconianSpearmen, x, y-1, data.PlaneArcanus, enemyWizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))
    enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.DraconianSpearmen, x, y+1, data.PlaneArcanus, enemyWizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))
    if kill {
        enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.DraconianSpearmen, x2, y-1, data.PlaneArcanus, enemyWizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))
        enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.DraconianSpearmen, x2, y, data.PlaneArcanus, enemyWizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))
        enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.DraconianSpearmen, x2, y+1, data.PlaneArcanus, enemyWizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))
    }

    x, y, _ = game.FindValidCityLocation(game.Plane)

    city = citylib.MakeCity("Test City", x, y, enemy.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, enemy)
    city.Population = 14000
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = buildinglib.BuildingHousing
    city.ProducingUnit = units.UnitNone
    city.AddBuilding(buildinglib.BuildingGranary)
    city.Farmers = 10
    city.Workers = 4
    city.ResetCitizens()
    enemy.AddCity(city)

    return game
}

// create undead units after combat ends
func createScenario53(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 53: create undead units")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
        Race: data.RaceHalfling,
        Retorts: []data.Retort{},
        Books: []data.WizardBook{
            {
                Magic: data.DeathMagic,
                Count: 11,
            },
        },
    }

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })
    game.Plane = data.PlaneArcanus

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player := game.AddPlayer(wizard, true)
    player.CastingSkillPower += 500000
    player.Gold = 1000
    player.Mana = 10000
    player.KnownSpells.AddSpell(allSpells.FindByName("Life Drain"))

    x, y, _ := game.FindValidCityLocationOnShore(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHalfling, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingNone
    city.ProducingUnit = units.HalflingSpearmen
    city.Race = wizard.Race
    city.Farmers = 13
    city.Workers = 3
    city.AddBuilding(buildinglib.BuildingFortress)
    city.AddBuilding(buildinglib.BuildingGranary)
    city.ResetCitizens()

    player.AddCity(city)
    player.LiftFog(x, y, 4, data.PlaneArcanus)
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.Ghoul, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.Ghoul, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.Ghoul, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    enemyWizard := setup.WizardCustom{
        Name: "enemy",
        Banner: data.BannerGreen,
        Race: data.RaceBeastmen,
    }
    enemy := game.AddPlayer(enemyWizard, false)
    x1 := game.ArcanusMap.WrapX(x-1)
    enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.BeastmenSpearmen, x1, y-1, data.PlaneArcanus, enemyWizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))

    x, y, _ = game.FindValidCityLocation(game.Plane)

    city = citylib.MakeCity("Test City", x, y, enemy.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, enemy)
    city.Population = 14000
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = buildinglib.BuildingHousing
    city.ProducingUnit = units.UnitNone
    city.AddBuilding(buildinglib.BuildingGranary)
    city.Farmers = 10
    city.Workers = 4
    city.ResetCitizens()
    enemy.AddCity(city)

    return game
}

// natures wrath
func createScenario54(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 54: natures wrath")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
        Race: data.RaceHalfling,
        Retorts: []data.Retort{},
        Books: []data.WizardBook{
            {
                Magic: data.DeathMagic,
                Count: 11,
            },
        },
    }

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })
    game.Plane = data.PlaneArcanus

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player := game.AddPlayer(wizard, true)
    player.CastingSkillPower += 500000
    player.Gold = 1000
    player.Mana = 10000
    player.KnownSpells.AddSpell(allSpells.FindByName("Corruption"))

    x, y, _ := game.FindValidCityLocationOnShore(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHalfling, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingNone
    city.ProducingUnit = units.HalflingSpearmen
    city.Race = wizard.Race
    city.Farmers = 13
    city.Workers = 3
    city.AddBuilding(buildinglib.BuildingFortress)
    city.AddBuilding(buildinglib.BuildingGranary)
    city.AddBuilding(buildinglib.BuildingMarketplace)
    city.AddBuilding(buildinglib.BuildingBank)
    city.AddBuilding(buildinglib.BuildingFarmersMarket)
    city.AddBuilding(buildinglib.BuildingWizardsGuild)
    city.AddBuilding(buildinglib.BuildingArmory)
    city.ResetCitizens()

    player.AddCity(city)
    player.LiftFog(x, y, 4, data.PlaneArcanus)
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.LizardSpearmen, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.Hydra, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    enemyWizard := setup.WizardCustom{
        Name: "enemy",
        Banner: data.BannerGreen,
        Race: data.RaceBeastmen,
    }
    enemy := game.AddPlayer(enemyWizard, false)
    x1 := game.ArcanusMap.WrapX(x-1)
    enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.BeastmenSpearmen, x1, y-1, data.PlaneArcanus, enemyWizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))
    enemy.Mana = 10000
    enemy.GlobalEnchantments.Insert(data.EnchantmentNaturesWrath)

    x, y, _ = game.FindValidCityLocation(game.Plane)

    city = citylib.MakeCity("Test City", x, y, enemy.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, enemy)
    city.Population = 14000
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = buildinglib.BuildingHousing
    city.ProducingUnit = units.UnitNone
    city.AddBuilding(buildinglib.BuildingGranary)
    city.Farmers = 10
    city.Workers = 4
    city.ResetCitizens()
    enemy.AddCity(city)

    return game
}

// suppress magic
func createScenario55(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 55: suppress magic")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
        Race: data.RaceHalfling,
        Retorts: []data.Retort{},
        Books: []data.WizardBook{
            {
                Magic: data.DeathMagic,
                Count: 11,
            },
        },
    }

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })
    game.Plane = data.PlaneArcanus

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player := game.AddPlayer(wizard, true)
    player.CastingSkillPower += 500000
    player.Gold = 1000
    player.Mana = 10000
    player.KnownSpells.AddSpell(allSpells.FindByName("Corruption"))

    x, y, _ := game.FindValidCityLocationOnShore(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHalfling, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.ProducingBuilding = buildinglib.BuildingNone
    city.ProducingUnit = units.HalflingSpearmen
    city.Race = wizard.Race
    city.Farmers = 13
    city.Workers = 3
    city.AddBuilding(buildinglib.BuildingFortress)
    city.AddBuilding(buildinglib.BuildingGranary)
    city.AddBuilding(buildinglib.BuildingMarketplace)
    city.AddBuilding(buildinglib.BuildingBank)
    city.AddBuilding(buildinglib.BuildingFarmersMarket)
    city.AddBuilding(buildinglib.BuildingWizardsGuild)
    city.AddBuilding(buildinglib.BuildingArmory)
    city.ResetCitizens()

    player.AddCity(city)
    player.LiftFog(x, y, 4, data.PlaneArcanus)
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.LizardSpearmen, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    player.AddUnit(units.MakeOverworldUnitFromUnit(units.Hydra, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    enemyWizard := setup.WizardCustom{
        Name: "enemy",
        Banner: data.BannerGreen,
        Race: data.RaceBeastmen,
    }
    enemy := game.AddPlayer(enemyWizard, false)
    x1 := game.ArcanusMap.WrapX(x-1)
    enemy.AddUnit(units.MakeOverworldUnitFromUnit(units.BeastmenSpearmen, x1, y-1, data.PlaneArcanus, enemyWizard.Banner, enemy.MakeExperienceInfo(), enemy.MakeUnitEnchantmentProvider()))
    enemy.Mana = 10000
    enemy.GlobalEnchantments.Insert(data.EnchantmentSuppressMagic)

    x, y, _ = game.FindValidCityLocation(game.Plane)

    city = citylib.MakeCity("Test City", x, y, enemy.Wizard.Race, game.BuildingInfo, game.CurrentMap(), game, enemy)
    city.Population = 14000
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = buildinglib.BuildingHousing
    city.ProducingUnit = units.UnitNone
    city.AddBuilding(buildinglib.BuildingGranary)
    city.Farmers = 10
    city.Workers = 4
    city.ResetCitizens()
    enemy.AddCity(city)

    return game
}

func createScenario56(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 56: regenerate after combat")

    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerBlue,
        Race: data.RaceTroll,
        Retorts: []data.Retort{
            data.RetortAlchemy,
            data.RetortSageMaster,
            data.RetortRunemaster,
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

    player.TaxRate = fraction.Zero()

    x, y, _ := game.FindValidCityLocation(game.Plane)

    /*
    x = 20
    y = 20
    */

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.Buildings.Insert(buildinglib.BuildingGranary)
    city.Buildings.Insert(buildinglib.BuildingFarmersMarket)
    city.ProducingBuilding = buildinglib.BuildingBank
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 14
    city.Workers = 3

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 2600
    player.CastingSkillPower = 10000

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player.KnownSpells.AddSpell(allSpells.FindByName("Animate Dead"))

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    // log.Printf("City at %v, %v", x, y)

    player.LiftFog(x, y, 30, data.PlaneArcanus)
    player.LiftFog(x, y, 30, data.PlaneMyrror)

    for range 8 {
        player.AddUnit(units.MakeOverworldUnitFromUnit(units.TrollSpearmen, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    }

    // player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, 30, 30, data.PlaneArcanus, wizard.Banner))

    /*
    stack := player.FindStackByUnit(drake)
    player.SetSelectedStack(stack)
    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)
    */

    enemy1 := game.AddPlayer(setup.WizardCustom{
        Name: "dingus",
        Banner: data.BannerRed,
    }, false)

    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))
    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))

    // game.Camera.Center(stack.X(), stack.Y())

    return game
}

func createScenario57(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 57: zombie mastery")

    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerBlue,
        Race: data.RaceTroll,
        Retorts: []data.Retort{
            data.RetortAlchemy,
            data.RetortSageMaster,
            data.RetortRunemaster,
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

    player.TaxRate = fraction.Zero()

    x, y, _ := game.FindValidCityLocation(game.Plane)

    /*
    x = 20
    y = 20
    */

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.Buildings.Insert(buildinglib.BuildingGranary)
    city.Buildings.Insert(buildinglib.BuildingFarmersMarket)
    city.ProducingBuilding = buildinglib.BuildingBank
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 14
    city.Workers = 3

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 2600
    player.CastingSkillPower = 10000

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player.KnownSpells.AddSpell(allSpells.FindByName("Animate Dead"))

    player.GlobalEnchantments.Insert(data.EnchantmentZombieMastery)

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    // log.Printf("City at %v, %v", x, y)

    player.LiftFog(x, y, 30, data.PlaneArcanus)
    player.LiftFog(x, y, 30, data.PlaneMyrror)

    for range 8 {
        player.AddUnit(units.MakeOverworldUnitFromUnit(units.BeastmenSwordsmen, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    }

    // player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, 30, 30, data.PlaneArcanus, wizard.Banner))

    /*
    stack := player.FindStackByUnit(drake)
    player.SetSelectedStack(stack)
    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)
    */

    enemy1 := game.AddPlayer(setup.WizardCustom{
        Name: "dingus",
        Banner: data.BannerRed,
    }, false)

    enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))
    // enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, x + 2, y + 2, data.PlaneArcanus, enemy1.Wizard.Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))

    // game.Camera.Center(stack.X(), stack.Y())

    return game
}

// resurrect heroes
func createScenario58(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 58: resurrect heroes")
    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerRed,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{
        Magic: data.MagicSettingNormal,
        Difficulty: data.DifficultyAverage,
    })

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.CastingSkillPower += 6000

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player.KnownSpells.AddSpell(allSpells.FindByName("Resurrection"))

    x, y, _ := game.FindValidCityLocation(game.Plane)

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.AddBuilding(buildinglib.BuildingFortress)
    city.AddBuilding(buildinglib.BuildingSummoningCircle)
    city.AddBuilding(buildinglib.BuildingShrine)
    city.AddBuilding(buildinglib.BuildingMarketplace)
    city.AddBuilding(buildinglib.BuildingWizardsGuild)
    city.AddBuilding(buildinglib.BuildingBank)
    city.Race = wizard.Race
    city.Farmers = 13
    city.Workers = 3

    city.ResetCitizens()

    player.AddCity(city)

    for range 7 {
        player.AddUnit(units.MakeOverworldUnitFromUnit(units.TrollSpearmen, city.X, city.Y, city.Plane, player.GetBanner(), player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    }

    player.Gold = 1000
    player.Mana = 5000

    player.LiftFog(x, y, 4, data.PlaneArcanus)

    for _, hero := range player.HeroPool {
        hero.AdjustHealth(-1000)
    }

    return game
}

func createScenario59(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 59: units on both planes")

    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerBlue,
        Race: data.RaceTroll,
        Retorts: []data.Retort{
            data.RetortAlchemy,
            data.RetortSageMaster,
            data.RetortRunemaster,
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

    player.TaxRate = fraction.Zero()

    x, y, _ := game.FindValidCityLocation(game.Plane)

    /*
    x = 20
    y = 20
    */

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.Buildings.Insert(buildinglib.BuildingGranary)
    city.Buildings.Insert(buildinglib.BuildingFarmersMarket)
    city.ProducingBuilding = buildinglib.BuildingBank
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 14
    city.Workers = 3

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 2600
    player.CastingSkillPower = 10000

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player.KnownSpells.AddSpell(allSpells.FindByName("Animate Dead"))

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    // log.Printf("City at %v, %v", x, y)

    player.LiftFog(x, y, 30, data.PlaneArcanus)
    player.LiftFog(x, y, 30, data.PlaneMyrror)

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    for i := 0; i < 5; i++ {
        fireElemental := player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, x + 1, y + 1, data.PlaneMyrror, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
        _ = fireElemental
    }

    // player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, 30, 30, data.PlaneArcanus, wizard.Banner))

    stack := player.FindStackByUnit(drake)
    player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)
    player.LiftFog(x, y, 2, data.PlaneMyrror)

    game.Camera.Center(stack.X(), stack.Y())

    return game
}

// tranquility fizzle
func createScenario60(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 60: tranquility fizzle spell")

    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerBlue,
        Race: data.RaceTroll,
        Retorts: []data.Retort{
            data.RetortAlchemy,
            data.RetortSageMaster,
            data.RetortRunemaster,
        },
        Books: []data.WizardBook{
            data.WizardBook{
                Magic: data.DeathMagic,
                Count: 3,
            },
            data.WizardBook{
                Magic: data.NatureMagic,
                Count: 8,
            },
        },
    }

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.TaxRate = fraction.Zero()

    x, y, _ := game.FindValidCityLocation(game.Plane)

    /*
    x = 20
    y = 20
    */

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.Buildings.Insert(buildinglib.BuildingGranary)
    city.Buildings.Insert(buildinglib.BuildingFarmersMarket)
    city.ProducingBuilding = buildinglib.BuildingBank
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 14
    city.Workers = 3

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 2600
    player.CastingSkillPower = 10000

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player.KnownSpells.AddSpell(allSpells.FindByName("Animate Dead"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Corruption"))

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    // log.Printf("City at %v, %v", x, y)

    player.LiftFog(x, y, 30, data.PlaneArcanus)
    player.LiftFog(x, y, 30, data.PlaneMyrror)

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    for i := 0; i < 5; i++ {
        fireElemental := player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, x + 1, y + 1, data.PlaneMyrror, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
        _ = fireElemental
    }

    // player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, 30, 30, data.PlaneArcanus, wizard.Banner))

    stack := player.FindStackByUnit(drake)
    player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)
    player.LiftFog(x, y, 2, data.PlaneMyrror)

    game.Camera.Center(stack.X(), stack.Y())

    enemy1 := game.AddPlayer(setup.WizardCustom{
        Name: "Tauron",
        Banner: data.BannerRed,
    }, false)

    enemy1.AddEnchantment(data.EnchantmentTranquility)

    // game.ShowTranquilityFizzle(player, enemy1, allSpells.FindByName("Healing"))

    return game
}

// power history
func createScenario61(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 60: power history")

    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerBlue,
        Race: data.RaceTroll,
        Retorts: []data.Retort{
            data.RetortAlchemy,
            data.RetortSageMaster,
            data.RetortRunemaster,
        },
        Books: []data.WizardBook{
            data.WizardBook{
                Magic: data.DeathMagic,
                Count: 3,
            },
            data.WizardBook{
                Magic: data.NatureMagic,
                Count: 8,
            },
        },
    }

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.TaxRate = fraction.Zero()

    x, y, _ := game.FindValidCityLocation(game.Plane)

    /*
    x = 20
    y = 20
    */

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.Buildings.Insert(buildinglib.BuildingGranary)
    city.Buildings.Insert(buildinglib.BuildingFarmersMarket)
    city.ProducingBuilding = buildinglib.BuildingBank
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 14
    city.Workers = 3

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 2600
    player.CastingSkillPower = 10000

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player.KnownSpells.AddSpell(allSpells.FindByName("Animate Dead"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Corruption"))

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    // log.Printf("City at %v, %v", x, y)

    player.LiftFog(x, y, 30, data.PlaneArcanus)
    player.LiftFog(x, y, 30, data.PlaneMyrror)

    drake := player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    for i := 0; i < 5; i++ {
        fireElemental := player.AddUnit(units.MakeOverworldUnitFromUnit(units.FireElemental, x + 1, y + 1, data.PlaneMyrror, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
        _ = fireElemental
    }

    // player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, 30, 30, data.PlaneArcanus, wizard.Banner))

    stack := player.FindStackByUnit(drake)
    player.SetSelectedStack(stack)

    player.LiftFog(stack.X(), stack.Y(), 2, data.PlaneArcanus)
    player.LiftFog(x, y, 2, data.PlaneMyrror)

    game.Camera.Center(stack.X(), stack.Y())

    enemy1 := game.AddPlayer(setup.WizardCustom{
        Name: "Tauron",
        Banner: data.BannerRed,
    }, false)

    game.TurnNumber = 300
    power := playerlib.WizardPower{
        Army: 5,
        Magic: 5,
        SpellResearch: 5,
    }

    enemyPower := playerlib.WizardPower{
        Army: 5,
        Magic: 5,
        SpellResearch: 5,
    }

    for range game.TurnNumber {
        player.AddPowerHistory(power)
        V := 10
        power.Army = max(0, power.Army + rand.N(V*2) - V)
        power.Magic = max(0, power.Magic + rand.N(V*3) - V)
        power.SpellResearch = max(0, power.SpellResearch + rand.N(V*2) - V)

        enemy1.AddPowerHistory(enemyPower)
        enemyPower.Army = max(0, enemyPower.Army + rand.N(V*3) - V)
        enemyPower.Magic = max(0, enemyPower.Magic + rand.N(V*2) - V)
        enemyPower.SpellResearch = max(0, enemyPower.SpellResearch + rand.N(V*2) - V)
    }

    player.AwarePlayer(enemy1)

    game.Events <- &gamelib.GameEventHistorian{}

    return game
}

// cities too close to each other with shared tiles
func createScenario62(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 62")
    wizard := setup.WizardCustom{
        Name: "player",
        Banner: data.BannerBlue,
        Race: data.RaceHighMen,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    x, y, _ := game.FindValidCityLocation(game.Plane)

    introCity := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    introCity.Population = 14000
    introCity.Plane = data.PlaneArcanus
    introCity.ProducingBuilding = buildinglib.BuildingHousing
    introCity.ProducingUnit = units.UnitNone
    introCity.Farmers = 14

    introCity.AddBuilding(buildinglib.BuildingShrine)

    introCity.ResetCitizens()

    player.AddCity(introCity)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    _ = introCity

    for i := 0; i < 3; i++ {
        player.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, x + i, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))
    }

    city2 := citylib.MakeCity("utah", x + 2, y + 1, data.RaceDarkElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city2.Population = 7000
    city2.Plane = data.PlaneArcanus
    city2.ProducingBuilding = buildinglib.BuildingShrine
    city2.ProducingUnit = units.UnitNone

    city2.ResetCitizens()

    player.AddCity(city2)

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    game.Camera.Center(x, y)

    return game
}

// nightshade protecting a town fizzles spells cast on it
func createScenario63(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 63")
    wizard := setup.WizardCustom{
        Name: "player",
        Banner: data.BannerBlue,
        Race: data.RaceHighMen,
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

    game := gamelib.MakeGame(cache, setup.NewGameSettings{})

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard, true)

    player.Mana = 5000
    player.CastingSkillPower = 20000

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player.KnownSpells.AddSpell(allSpells.FindByName("Chaos Rift"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Corruption"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Call the Void"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Evil Presence"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Famine"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Cursed Lands"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Pestilence"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Raise Volcano"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Ice Storm"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Stasis"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Fire Storm"))
    player.KnownSpells.AddSpell(allSpells.FindByName("Black Wind"))

    x, y, _ := game.FindValidCityLocation(game.Plane)

    introCity := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    introCity.Population = 14000
    introCity.Plane = data.PlaneArcanus
    introCity.ProducingBuilding = buildinglib.BuildingHousing
    introCity.ProducingUnit = units.UnitNone
    introCity.Farmers = 14

    introCity.AddBuilding(buildinglib.BuildingShrine)
    introCity.AddBuilding(buildinglib.BuildingOracle)

    introCity.ResetCitizens()

    player.AddCity(introCity)

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    _ = introCity

    enemy1 := game.AddPlayer(setup.WizardCustom{
        Name: "dingus",
        Banner: data.BannerRed,
    }, false)

    arcanusMap := game.GetMap(data.PlaneArcanus)
    city2 := citylib.MakeCity("utah", arcanusMap.WrapX(x + 2), y + 1, data.RaceDarkElf, game.BuildingInfo, game.CurrentMap(), game, enemy1)
    city2.Population = 14000
    city2.Farmers = 12
    city2.Plane = data.PlaneArcanus
    city2.ProducingBuilding = buildinglib.BuildingShrine
    city2.ProducingUnit = units.UnitNone

    city2.Buildings.Insert(buildinglib.BuildingShrine)

    city2.ResetCitizens()

    enemy1.AddCity(city2)

    for i := 0; i < 3; i++ {
        newUnit := enemy1.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenSpearmen, city2.X, city2.Y, data.PlaneArcanus, enemy1.GetWizard().Banner, enemy1.MakeExperienceInfo(), enemy1.MakeUnitEnchantmentProvider()))
        newUnit.SetBusy(units.BusyStatusPatrol)
    }

    arcanusMap.SetBonus(arcanusMap.WrapX(city2.X + 1), city2.Y + 1, data.BonusNightshade)
    arcanusMap.SetBonus(arcanusMap.WrapX(city2.X + 1), city2.Y + 2, data.BonusNightshade)

    player.LiftFog(x, y, 3, data.PlaneArcanus)

    game.Camera.Center(x, y)

    return game
}

// hero with a wand that grants water walking should be able to walk on water
func createScenario64(cache *lbx.LbxCache) *gamelib.Game {
    log.Printf("Running scenario 64")

    wizard := setup.WizardCustom{
        Name: "bob",
        Banner: data.BannerBlue,
        Race: data.RaceTroll,
        Retorts: []data.Retort{
            data.RetortAlchemy,
            data.RetortSageMaster,
            data.RetortWarlord,
            data.RetortChanneler,
            data.RetortMyrran,
            data.RetortFamous,
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

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.BuildingInfo, game.CurrentMap(), game, player)
    city.Population = 8190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.ProducingBuilding = buildinglib.BuildingGranary
    city.ProducingUnit = units.UnitNone
    city.Race = wizard.Race
    city.Farmers = 5
    city.Workers = 3

    city.ResetCitizens()

    player.AddCity(city)

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(x, y, 3, data.PlaneArcanus)
    player.LiftFog(20, 20, 100, data.PlaneMyrror)

    rakir := hero.MakeHero(units.MakeOverworldUnit(units.HeroRakir, x, y, data.PlaneArcanus), hero.HeroRakir, "bubba")
    player.AddHeroToFortress(rakir)
    rakir.AddExperience(528)
    rakir.Status = hero.StatusEmployed

    rakir.Equipment[0] = &artifact.Artifact{
        Name: "Baloney",
        Image: 7,
        Type: artifact.ArtifactTypeSword,
        Powers: []artifact.Power{
            {
                Type: artifact.PowerTypeAbility1,
                Amount: 1,
                Name: "Water Walking",
                Ability: data.ItemAbilityWaterWalking,
            },
        },
        Cost: 250,
    }

    stack := player.FindStackByUnit(rakir)
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
        case 40: game = createScenario40(cache)
        case 41: game = createScenario41(cache)
        case 42: game = createScenario42(cache)
        case 43: game = createScenario43(cache)
        case 44: game = createScenario44(cache)
        case 45: game = createScenario45(cache)
        case 46: game = createScenario46(cache)
        case 47: game = createScenario47(cache)
        case 48: game = createScenario48(cache)
        case 49: game = createScenario49(cache)
        case 50: game = createScenario50(cache)
        case 51: game = createScenario51_52(cache, false)
        case 52: game = createScenario51_52(cache, true)
        case 53: game = createScenario53(cache)
        case 54: game = createScenario54(cache)
        case 55: game = createScenario55(cache)
        case 56: game = createScenario56(cache)
        case 57: game = createScenario57(cache)
        case 58: game = createScenario58(cache)
        case 59: game = createScenario59(cache)
        case 60: game = createScenario60(cache)
        case 61: game = createScenario61(cache)
        case 62: game = createScenario62(cache)
        case 63: game = createScenario63(cache)
        case 64: game = createScenario64(cache)
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
    return scale.Scale2(data.ScreenWidth, data.ScreenHeight)
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

    engine.Game.Shutdown()

    memoryProfile, err := os.Create("profile.mem.overworld")
    if err != nil {
        log.Printf("Error creating memory profile: %v", err)
    } else {
        defer memoryProfile.Close()
        pprof.WriteHeapProfile(memoryProfile)
    }

}
