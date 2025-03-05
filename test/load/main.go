package main

import (
    "os"
    "fmt"

    "github.com/kazzmir/master-of-magic/game/magic/load"
)

func main(){
    if len(os.Args) < 2 {
        fmt.Printf("Give a GAM file to load\n")
        return
    }

    reader, err := os.Open(os.Args[1])
    if err != nil {
        fmt.Printf("Error opening file: %v\n", err)
        return
    }
    defer reader.Close()

    data, err := load.LoadSaveGame(reader)
    if err != nil {
        fmt.Printf("Error loading save game: %v\n", err)
        return
    }

    fmt.Printf("Loaded saved game\n")
    fmt.Printf("Num players: %d\n", data.NumPlayers)
}
