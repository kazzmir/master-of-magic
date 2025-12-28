package main

import (
    "encoding/json"
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"
    "github.com/kazzmir/master-of-magic/game/magic/music"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
)

func main() {
    cache := lbx.AutoCache()
    game := gamelib.MakeGame(cache, music.MakeMusic(cache), setup.NewGameSettings{LandSize: 0})

    serialized := gamelib.SerializeModel(game.Model)

    log.Printf("Serialized model: %v", serialized)
    jsonData, err := json.Marshal(serialized)
    if err != nil {
        log.Fatalf("Failed to marshal serialized model to JSON: %v", err)
    }
    log.Printf("Serialized model JSON:\n%s", string(jsonData))
}
