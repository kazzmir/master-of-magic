package main

import (
    "os"
    "io"
    "fmt"
    "encoding/binary"
    "strings"
    "archive/zip"
)

const LbxSignature = 0x0000fead

func readUint16(reader io.Reader) (uint16, error) {
    var value uint16
    err := binary.Read(reader, binary.LittleEndian, &value)
    return value, err
}

func readUint32(reader io.Reader) (uint32, error) {
    var value uint32
    err := binary.Read(reader, binary.LittleEndian, &value)
    return value, err
}

func readInt32(reader io.Reader) (int32, error) {
    v, err := readUint32(reader)
    return int32(v), err
}

func dumpLbx(reader io.Reader) error {
    numFiles, err := readUint16(reader)
    if err != nil {
        return err
    }

    signature, err := readUint32(reader)
    if err != nil {
        return err
    }

    fmt.Printf("Number of files: %v\n", numFiles)
    fmt.Printf("Signature: 0x%x\n", signature)

    if signature != LbxSignature {
        return fmt.Errorf("Invalid lbx signature, was 0x%x but expected 0x%x\n", signature, LbxSignature)
    }

    version, err := readUint16(reader)
    if err != nil {
        return err
    }

    fmt.Printf("Version: %v\n", version)

    for i := 0; i < int(numFiles); i++ {
        offset, err := readUint32(reader)
        if err != nil {
            return err
        }

        fmt.Printf("Offset %v: 0x%x\n", i, offset)
    }

    return nil
}

func main(){
    if len(os.Args) < 2 {
        fmt.Println("Give an lbx file, or a zip file and the name of an lbx file inside it")
        return
    }

    if len(os.Args) == 2 {
        fmt.Printf("Opening %v as an lbx file\n", os.Args[1])
    } else if len(os.Args) == 3 {
        zipFile, err := zip.OpenReader(os.Args[1])
        if err != nil {
            fmt.Printf("Error opening zip file: %s\n", err)
            return
        }
        defer zipFile.Close()

        searchName := os.Args[2]

        var matches []string
        for _, file := range zipFile.File {
            // fmt.Printf("Entry: %s\n", file.Name)

            lower := strings.ToLower(file.Name)
            check := strings.ToLower(searchName)

            if strings.Contains(lower, check) {
                matches = append(matches, file.Name)
            }
        }

        if len(matches) == 0 {
            fmt.Printf("No such entry with name '%v'\n", searchName)
            return
        }

        if len(matches) > 1 {
            fmt.Printf("More than one match found for '%v'\n", searchName)
            for _, name := range matches {
                fmt.Printf("  %v\n", name)
            }
            return
        }

        match := matches[0]
        for _, file := range zipFile.File {
            if file.Name == match {
                opened, err := file.Open()
                if err != nil {
                    fmt.Printf("Unable to open entry %v: %v\n", file.Name, err)
                } else {
                    fmt.Printf("Dumping %v\n", file.Name)
                    err := dumpLbx(opened)
                    if err != nil {
                        fmt.Printf("Error dumping lbx file: %v\n", err)
                    }
                    opened.Close()
                }
            }
        }

    } else {
        fmt.Println("Too many arguments")
        return
    }

}
