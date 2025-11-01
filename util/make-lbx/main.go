package main

import (
    "os"
    "log"
    "image"
    "bufio"
    _ "image/png"

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
            file, err := os.Open(filePath)
            if err != nil {
                log.Printf("Error reading file %s: %v\n", filePath, err)
            } else {
                defer file.Close()
                img, _, err := image.Decode(bufio.NewReader(file))
                if err != nil {
                    log.Printf("Error decoding image %s: %v\n", filePath, err)
                } else {
                    encoded, err := lbx.EncodeImages([]image.Image{img}, lbx.GetDefaultPalette())
                    if err != nil {
                        log.Printf("Error encoding image %s: %v\n", filePath, err)
                        return
                    }
                    lbxFile.Data = append(lbxFile.Data, encoded)
                }
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

    log.Printf("Successfully created output.lbx with %d files\n", len(lbxFile.Data))
}
