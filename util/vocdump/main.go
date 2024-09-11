package main

import (
    "os"
    "fmt"
    "path/filepath"
    "io"
    "strconv"
    "bytes"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/voc"
)

func saveVoc(lbxFile *lbx.LbxFile, entryIndex int) error {
    outputDir := "voc"

    os.Mkdir(outputDir, 0755)

    entry, err := lbxFile.RawData(entryIndex)
    if err != nil {
        return err
    }

    reader := bytes.NewReader(entry)
    // the voc data starts at offset 16
    reader.Seek(16, io.SeekStart)

    vocFile, err := voc.Load(reader)

    if err != nil {
        return fmt.Errorf("Unable to read voc entry %v: %v", entryIndex, err)
    }

    // fmt.Printf("Voc file: %v\n", vocFile)

    output, err := os.Create(filepath.Join(outputDir, fmt.Sprintf("output-%03d.voc", entryIndex)))
    if err != nil {
        return err
    }
    defer output.Close()
    voc.Save(output, vocFile.SampleRate(), vocFile.AllSamples())

    fmt.Printf("Wrote voc file to %v/output-%03d.voc\n", outputDir, entryIndex)

    return nil
}

func main(){
    if len(os.Args) < 2 {
        fmt.Printf("Usage: %s <lbx file> [<entry index>]\n", os.Args[0])
        return
    }

    lbxPath := os.Args[1]

    opened, err := os.Open(lbxPath)
    if err != nil {
        fmt.Printf("Error opening file %v: %v\n", lbxPath, err)
        return
    }
    defer opened.Close()

    lbxFile, err := lbx.ReadLbx(opened)
    if err != nil {
        fmt.Printf("Error reading lbx file %v: %v\n", lbxPath, err)
        return
    }

    if len(os.Args) == 3 {
        entryIndex, err := strconv.Atoi(os.Args[2])
        if err != nil {
            fmt.Printf("Error parsing entry index %v: %v\n", os.Args[2], err)
            return
        }

        err = saveVoc(&lbxFile, entryIndex)
        if err != nil {
            fmt.Printf("Error saving voc file: %v\n", err)
        }
    } else {
        for i := 0; i < lbxFile.TotalEntries(); i++ {
            err = saveVoc(&lbxFile, i)
            if err != nil {
                fmt.Printf("Error saving voc file %v: %v\n", i, err)
            }
        }
    }
}
