package main

import (
    "os"
    "encoding/json"
    "log"
    "bytes"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/music"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
)

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    cache := lbx.AutoCache()
    useMusic := music.MakeMusic(cache)
    useMusic.Enabled = false
    game := gamelib.MakeGame(cache, useMusic, setup.NewGameSettings{LandSize: 0})

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

    player := game.AddPlayer(wizard, true)

    x := 3
    y := 8

    city := citylib.MakeCity("Test City", x, y, data.RaceHighElf, game.Model.BuildingInfo, game.Model.CurrentMap(), game.Model, player)
    city.Population = 16190
    city.Plane = data.PlaneArcanus
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
    city.Buildings.Insert(buildinglib.BuildingGranary)
    city.Buildings.Insert(buildinglib.BuildingFarmersMarket)
    city.ProducingBuilding = buildinglib.BuildingBank
    city.ProducingUnit = units.UnitNone
    city.Farmers = 14
    city.Workers = 3

    city.ResetCitizens()

    player.AddCity(city)

    game.Model.RandomEvents = append(game.Model.RandomEvents, gamelib.MakeDisjunctionEvent(18))
    game.Model.RandomEvents = append(game.Model.RandomEvents, gamelib.MakePopulationBoomEvent(17, city))

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    rakir := hero.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroRakir, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()), hero.HeroRakir, "Rakir")
    player.AddHero(rakir, x + 1, y + 2, data.PlaneArcanus)

    rakir.Equipment[0] = &artifact.Artifact{
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

    player.CreateArtifact = &artifact.Artifact{
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

    serialized := gamelib.SerializeModel(game.Model, "test")

    // log.Printf("Serialized model: %v", serialized)
    jsonData, err := json.Marshal(serialized)
    if err != nil {
        log.Fatalf("Failed to marshal serialized model to JSON: %v", err)
    }
    // log.Printf("Serialized model JSON:\n%s", string(jsonData))

    out, err := os.Create("serialized_model.json")
    if err != nil {
        log.Fatalf("Failed to create output file: %v", err)
    }
    defer out.Close()

    _, err = out.Write(jsonData)
    if err != nil {
        log.Fatalf("Failed to write JSON data to file: %v", err)
    }

    log.Println("Serialized model written to serialized_model.json")

    log.Println("Stage 1: PASSED")

    reader := bytes.NewReader(jsonData)
    decoder := json.NewDecoder(reader)
    var loadedData gamelib.SerializedGame
    err = decoder.Decode(&loadedData)
    if err != nil {
        log.Fatalf("Failed to decode JSON data: %v", err)
    }
    log.Println("JSON data decoded successfully")

    log.Println("Stage 2: PASSED")

    newGame := gamelib.MakeGameFromSerialized(cache, useMusic, &loadedData)
    _ = newGame

    log.Println("Stage 3: PASSED")
}
