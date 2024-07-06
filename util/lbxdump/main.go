package main

import (
    "os"
    "io"
    "fmt"
    "bytes"
    "strings"
    "archive/zip"

    "github.com/kazzmir/master-of-magic/lib/lbx"
)

func dumpLbx(reader io.ReadSeeker) error {
    file, err := lbx.ReadLbx(reader)
    if err != nil {
        return err
    }
    
    fmt.Printf("Number of files: %v\n", len(file.Data))
    // fmt.Printf("Signature: 0x%x\n", signature)

    for i, data := range file.Data {
        fmt.Printf("File %v: 0x%x (%v) bytes\n", i, len(data), len(data))
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

                    var memory bytes.Buffer
                    io.Copy(&memory, opened)

                    err := dumpLbx(bytes.NewReader(memory.Bytes()))
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
