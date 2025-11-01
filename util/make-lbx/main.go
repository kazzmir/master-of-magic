package main

import (
    "os"
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
)

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    files := os.Args[1:]

    if len(files) == 0 {
        log.Fatal("No input files provided")
    }

    log.Printf("Processing %d files\n", len(files))

    var lbxFile lbx.LbxFile
    for _, filePath := range files {
        func() {
            bytes, err := os.ReadFile(filePath)
            if err != nil {
                log.Printf("Error reading file %s: %v\n", filePath, err)
            } else {
                lbxFile.Data = append(lbxFile.Data, bytes)
            }
        }()
    }

    output, err := os.Create("output.lbx")
    if err != nil {
        log.Fatalf("Error creating output file: %v\n", err)
    }
    defer output.Close()

    err = lbx.SaveLbx(lbxFile, output)
    if err != nil {
        log.Fatalf("Error saving LBX file: %v\n", err)
    }
}
