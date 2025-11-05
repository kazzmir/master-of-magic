package main

import (
    "os"
    "io"
    "log"
    "fmt"
    "image"
    "bufio"
    "bytes"
    _ "image/png"

    "github.com/kazzmir/master-of-magic/lib/lbx"
)

type LbxType int
const (
    LbxRaw LbxType = iota
    LbxImage
    LbxTerrain
)

type LbxBuilder struct {
    Data [][]byte
}

func (builder *LbxBuilder) NumData() int {
    return len(builder.Data)
}

func (builder *LbxBuilder) AddImage(img image.Image) error {
    encoded, err := lbx.EncodeImages([]image.Image{img}, lbx.GetDefaultPalette())
    if err != nil {
        return err
    }
    builder.Data = append(builder.Data, encoded)
    return nil
}

func (builder *LbxBuilder) AddTerrain(img image.Image) error {
    encoded, err := lbx.EncodeTerrainImage(img)
    if err != nil {
        return err
    }

    builder.Data = append(builder.Data, encoded)
    return nil
}

func (builder *LbxBuilder) AddFileType(path string, lbxType LbxType) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()

    switch lbxType {
        case LbxRaw:
            var buf bytes.Buffer
            _, err = io.Copy(&buf, file)
            if err != nil {
                return err
            }
            builder.Data = append(builder.Data, buf.Bytes())
        case LbxImage:
            img, _, err := image.Decode(bufio.NewReader(file))
            if err != nil {
                return err
            }
            return builder.AddImage(img)
        case LbxTerrain:
            img, _, err := image.Decode(bufio.NewReader(file))
            if err != nil {
                return err
            }
            return builder.AddTerrain(img)
    }

    return fmt.Errorf("Unknown LBX type")
}

func (builder *LbxBuilder) MakeLbx() *lbx.LbxFile {
    return &lbx.LbxFile{
        Data: builder.Data,
    }
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    files := os.Args[1:]

    if len(files) == 0 {
        log.Fatal("No input files provided")
    }

    log.Printf("Processing %d files\n", len(files))

    var builder LbxBuilder

    for _, filePath := range files {
        err := builder.AddFileType(filePath, LbxImage)
        if err != nil {
            log.Printf("Error processing file %s: %v\n", filePath, err)
        }
    }

    output, err := os.Create("output.lbx")
    if err != nil {
        log.Fatalf("Error creating output file: %v\n", err)
    }
    defer output.Close()

    err = lbx.SaveLbx(builder.MakeLbx(), output)
    if err != nil {
        log.Fatalf("Error saving LBX file: %v\n", err)
    }

    log.Printf("Successfully created output.lbx with %d files\n", builder.NumData())
}
