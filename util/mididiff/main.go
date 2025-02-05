package main

import (
    "os"
    "fmt"
    "log"
    "reflect"

    "gitlab.com/gomidi/midi/v2/smf"
)

func main() {
    if len(os.Args) < 3 {
        fmt.Printf("Give two midi files\n")
        return
    }

    file1, err := os.Open(os.Args[1])
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }
    defer file1.Close()
    file2, err := os.Open(os.Args[2])
    if err != nil {
        log.Printf("Error: %v", err)
    }
    defer file2.Close()

    midi1, err := smf.ReadFrom(file1)
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }

    midi2, err := smf.ReadFrom(file2)
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }

    for i := range midi1.Tracks[0] {
        event1 := midi1.Tracks[0][i]
        if i >= len(midi2.Tracks[0]) {
            fmt.Printf("End of track for %v\n", file2)
            // fmt.Printf("[%d] Event1: %v\n", i, event1)
            break
        }
        event2 := midi2.Tracks[0][i]

        fmt.Printf("[%d] Event1: %v\n", i, event1)
        fmt.Printf("[%d] Event2: %v\n", i, event2)
        if !reflect.DeepEqual(event1, event2) {
            fmt.Printf("  not equal!\n")
            // break
        }
    }

    if len(midi1.Tracks[0]) != len(midi2.Tracks[0]) {
        fmt.Printf("Error: Tracks are not equal length!\n")
    }

}
