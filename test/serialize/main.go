package main

import (
    "os"
    "encoding/json"
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/music"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

func main() {
    cache := lbx.AutoCache()
    game := gamelib.MakeGame(cache, music.MakeMusic(cache), setup.NewGameSettings{LandSize: 0})

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

    player.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, x + 1, y + 1, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

    player.AddHero(hero.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroRakir, x, y, data.PlaneArcanus, wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()), hero.HeroRakir, "Rakir"), x + 1, y + 2, data.PlaneArcanus)

    serialized := gamelib.SerializeModel(game.Model)

    log.Printf("Serialized model: %v", serialized)
    jsonData, err := json.Marshal(serialized)
    if err != nil {
        log.Fatalf("Failed to marshal serialized model to JSON: %v", err)
    }
    log.Printf("Serialized model JSON:\n%s", string(jsonData))

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
}
