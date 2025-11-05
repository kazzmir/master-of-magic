package main

import (
    "os"
    "io"
    "log"
    "fmt"
    "flag"
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
    TerrainData [][]byte
}

func (builder *LbxBuilder) NumData() int {
    return len(builder.Data) + len(builder.TerrainData)
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

    builder.TerrainData = append(builder.TerrainData, encoded)
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
            return nil
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
    out := lbx.LbxFile{
        Data: builder.Data,
    }

    if len(builder.TerrainData) > 0 {
        var total bytes.Buffer
        total.Write(make([]byte, 192))
        for _, terrain := range builder.TerrainData {
            total.Write(terrain)
        }

        out.Data = append(out.Data, total.Bytes())
    }

    return &out
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    lbxType := LbxRaw

    useImage := flag.Bool("image", false, "Treat input files as images")
    useTerrain := flag.Bool("terrain", false, "Treat input files as terrain images")
    flag.Parse()

    if *useImage {
        lbxType = LbxImage
    }

    if *useTerrain {
        lbxType = LbxTerrain
    }

    files := os.Args[1 + flag.NFlag():]

    if len(files) == 0 {
        log.Fatal("No input files provided")
    }

    log.Printf("Processing %d files\n", len(files))

    var builder LbxBuilder

    for _, filePath := range files {
        err := builder.AddFileType(filePath, lbxType)
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
