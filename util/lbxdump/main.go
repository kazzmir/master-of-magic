package main

import (
    "os"
    "io"
    "fmt"
    "log"
    "image/png"
    "path/filepath"
    "strconv"
    "bytes"
    "strings"
    "archive/zip"

    "github.com/kazzmir/master-of-magic/lib/lbx"
)

func dumpLbx(reader io.ReadSeeker, lbxName string, onlyIndex int, rawDump bool) error {
    file, err := lbx.ReadLbx(reader)
    if err != nil {
        return err
    }
    
    fmt.Printf("Number of files: %v\n", len(file.Data))
    // fmt.Printf("Signature: 0x%x\n", signature)

    dir := fmt.Sprintf("%v_output", lbxName)

    os.Mkdir(dir, 0755)

    if lbxName == "terrain.lbx" {
        index := 0
        images, err := file.ReadTerrainImages(index)
        if err != nil {
            return err
        }

        fmt.Printf("Loaded %v images\n", len(images))
        for i, image := range images {
            func (){
                name := filepath.Join(dir, fmt.Sprintf("image_%v_%v.png", index, i))
                out, err := os.Create(name)
                if err != nil {
                    fmt.Printf("Error creating image file: %v\n", err)
                    return
                }
                defer out.Close()

                png.Encode(out, image)
                fmt.Printf("Saved image %v to %v\n", i, name)
            }()
        }
    } else {
        for index, data := range file.Data {

            if onlyIndex != -1 && index != onlyIndex {
                continue
            }

            fmt.Printf("File %v: 0x%x (%v) bytes\n", index, len(data), len(data))

            if rawDump {
                func (){
                    name := filepath.Join(dir, fmt.Sprintf("file_%v.bin", index))
                    out, err := os.Create(name)
                    if err != nil {
                        fmt.Printf("Error creating file: %v\n", err)
                        return
                    }
                    defer out.Close()

                    out.Write(data)
                    fmt.Printf("Saved raw data to %v\n", name)
                }()
            } else if len(data) > 0 {
                images, err := file.ReadImages(index)
                if err != nil {
                    log.Printf("Unable to load entry %v as images: %v\n", index, err)
                    continue
                }

                fmt.Printf("Loaded %v images\n", len(images))
                for i, image := range images {
                    func (){
                        name := filepath.Join(dir, fmt.Sprintf("image_%v_%v.png", index, i))
                        out, err := os.Create(name)
                        if err != nil {
                            fmt.Printf("Error creating image file: %v\n", err)
                            return
                        }
                        defer out.Close()

                        png.Encode(out, image)
                        fmt.Printf("Saved image %v to %v\n", i, name)
                    }()
                }
            }
        }

    }

    return nil
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    if len(os.Args) < 2 {
        fmt.Println("Give an lbx file, or a zip file and the name of an lbx file inside it")
        return
    }

    if len(os.Args) == 2 {
        path := os.Args[1]
        fmt.Printf("Opening %v as an lbx file\n", path)

        file, err := os.Open(path)
        if err != nil {
            log.Printf("Error opening %v: %v\n", path, err)
            return
        }
        onlyIndex := -1
        rawDump := true
        err = dumpLbx(file, strings.ToLower(path), onlyIndex, rawDump)
        if err != nil {
            log.Printf("Error dumping lbx file: %v\n", err)
        }
    } else if len(os.Args) >= 3 {
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

            // exact match
            if lower == check {
                matches = []string{file.Name}
                break
            }

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

        rawDump := false
        onlyIndex := -1
        if len(os.Args) >= 4 {
            arg := os.Args[3]
            switch arg {
                case "raw", "--raw", "-raw":
                    rawDump = true
                default:
                    onlyIndex, err = strconv.Atoi(arg)
                    if err != nil {
                        fmt.Printf("Expected index to be an integer: %v\n", arg)
                        onlyIndex = -1
                    }
            }
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

                    err := dumpLbx(bytes.NewReader(memory.Bytes()), strings.ToLower(file.Name), onlyIndex, rawDump)
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
