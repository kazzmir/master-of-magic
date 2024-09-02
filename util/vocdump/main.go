package main

import (
    "os"
    "fmt"
    "io"
    "strconv"
    "bytes"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/voc"
)

func main(){
    if len(os.Args) < 3 {
        fmt.Printf("Usage: %s <lbx file> <entry index>\n", os.Args[0])
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

    entryIndex, err := strconv.Atoi(os.Args[2])
    if err != nil {
        fmt.Printf("Error parsing entry index %v: %v\n", os.Args[2], err)
        return
    }

    entry, err := lbxFile.RawData(entryIndex)
    if err != nil {
        fmt.Printf("Error reading entry %v: %v\n", entryIndex, err)
        return
    }

    reader := bytes.NewReader(entry)
    // the voc data starts at offset 16
    reader.Seek(16, io.SeekStart)

    vocFile, err := voc.Load(reader)

    if err != nil {
        fmt.Printf("Error reading voc file: %v\n", err)
        return
    }

    // fmt.Printf("Voc file: %v\n", vocFile)

    output, err := os.Create(fmt.Sprintf("output-%d.voc", entryIndex))
    if err != nil {
        fmt.Printf("Error creating output file: %v\n", err)
        return
    }
    defer output.Close()
    voc.Save(output, vocFile.SampleRate(), vocFile.AllSamples())

    fmt.Printf("Wrote voc file to output-%d.voc\n", entryIndex)
}
