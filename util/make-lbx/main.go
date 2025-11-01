package main

import (
    "os"
    "log"
)

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    files := os.Args[1:]

    if len(files) == 0 {
        log.Fatal("No input files provided")
    }

    log.Printf("Processing %d files\n", len(files))
}
