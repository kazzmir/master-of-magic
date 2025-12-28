package main

import (
    "os"
    "encoding/json"
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"
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

    game.AddPlayer(wizard, true)

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
