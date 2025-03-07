package main

import (
    "os"
    "io"
    "fmt"
    "log"
    "image/color"
    "image/png"
    "path/filepath"
    "bytes"
    "strings"
    "archive/zip"
    "flag"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/game"
    "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/building"
    helplib "github.com/kazzmir/master-of-magic/game/magic/help"
)

func dumpLbx(reader io.ReadSeeker, lbxName string, onlyIndex int, rawDump bool, voc bool) error {
    file, err := lbx.ReadLbx(reader)
    if err != nil {
        return err
    }

    fmt.Printf("Number of files: %v\n", len(file.Data))
    // fmt.Printf("Signature: 0x%x\n", signature)

    dir := fmt.Sprintf("%v_output", lbxName)

    os.Mkdir(dir, 0755)

    soundFiles := set.NewSet("soundfx.lbx", "newsound.lbx", "introsfx.lbx", "cmbtsnd.lbx")

    if lbxName == "terrain.lbx" && !rawDump {
        index := 0
        images, err := file.ReadTerrainImages(index)
        if err != nil {
            return err
        }

        // fmt.Printf("Loaded %v images\n", len(images))
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
                // fmt.Printf("Saved image %v to %v\n", i, name)
            }()
        }
    } else if lbxName == "names.lbx" && !rawDump {
        files := make(map[string]*lbx.LbxFile)
        files["names.lbx"] = &file
        cache := lbx.MakeCacheFromLbxFiles(files)

        names := hero.ReadNamesPerWizard(cache)
        for i, name := range names {
            log.Printf("Choice %v: %v\n", i, name)
        }

        /*
        _, err := hero.ReadNames(cache)
        if err != nil {
            return err
        }
        */
    } else if lbxName == "fonts.lbx" && !rawDump {
        fonts, err := font.ReadFonts(&file, 0)
        if err != nil {
            return fmt.Errorf("Unable to read fonts: %v", err)
        }

        fmt.Printf("Fonts: %v\n", len(fonts))
        for i, font := range fonts {
            fmt.Printf("  font %v glyphs %v\n", i, font.GlyphCount())
        }
    } else if lbxName == "spelldat.lbx" && !rawDump {
        spells, err := spellbook.ReadSpells(&file, 0)
        if err != nil {
            return err
        }

        for i, spell := range spells.Spells {
            fmt.Printf("Spell %v: %+v\n", i, spell)
        }
    } else if lbxName == "eventmsg.lbx" && !rawDump {
        files := make(map[string]*lbx.LbxFile)
        files["eventmsg.lbx"] = &file
        cache := lbx.MakeCacheFromLbxFiles(files)
        events, err := game.ReadEventData(cache)
        if err != nil {
            return err
        }

        for i, event := range events.Events {
            fmt.Printf("Event %v: %+v\n", i, event)
        }

    } else if lbxName == "itemdata.lbx" && !rawDump {
        files := make(map[string]*lbx.LbxFile)
        files["itemdata.lbx"] = &file
        cache := lbx.MakeCacheFromLbxFiles(files)
        artifacts, err := artifact.ReadArtifacts(cache)

        if err != nil {
            return err
        }

        for _, use := range artifacts {
            fmt.Printf("Artifact: %+v\n", use)
        }

    } else if lbxName == "itempow.lbx" && !rawDump {
        files := make(map[string]*lbx.LbxFile)
        files["itempow.lbx"] = &file
        cache := lbx.MakeCacheFromLbxFiles(files)
        artifacts, costs, compatibilities, err := artifact.ReadPowers(cache)

        if err != nil {
            return err
        }

        for _, artifact := range artifacts {
            fmt.Printf("Power: %+v Cost: %v Artifact Types: %+v\n", artifact, costs[artifact], compatibilities[artifact])
        }
    } else if lbxName == "music.lbx" && !rawDump {

        for i, entry := range file.Data {
            (func(){
                name := fmt.Sprintf("music_%v.xmi", i)
                out, err := os.Create(filepath.Join(dir, name))
                if err != nil {
                    fmt.Printf("Error creating music file: %v\n", err)
                    return
                }
                defer out.Close()

                out.Write(entry[16:])
                fmt.Printf("Saved music %v to %v\n", i, name)
            })()
        }

    } else if lbxName == "help.lbx" && !rawDump {
        // uint16 number of entries
        // uint16 size of each entry
        help, err := helplib.ReadHelp(&file, 2)

        if err != nil {
            return err
        }

        // fmt.Printf("Help entries: %v\n", entries)
        /*
        for i, entry := range entries {
            if entry.AppendHelpIndex != -1 {
                fmt.Printf("Entry %v: %+v\n", i, entry)
            }
        }
        */

        /*
        fmt.Printf("Raw entry 365: %+v\n", help.GetRawEntry(365))
        fmt.Printf("Raw entry 366: %+v\n", help.GetRawEntry(366))
        fmt.Printf("Raw entry 367: %+v\n", help.GetRawEntry(367))

        fmt.Printf("Entry 365: %+v\n", help.GetEntries(365))
        */
        // fmt.Printf("Entry 'charismatic': %+v\n", help.GetEntriesByName("charismatic"))
        for i, entry := range help.Entries {
            fmt.Printf("Entry %v: %+v\n", i, entry)
        }

    } else if lbxName == "builddat.lbx" && !rawDump {
        files := make(map[string]*lbx.LbxFile)
        files["builddat.lbx"] = &file
        cache := lbx.MakeCacheFromLbxFiles(files)
        buildings, err := building.ReadBuildingInfo(cache)

        if err != nil {
            return err
        }

        for _, building := range buildings {
            fmt.Printf("%+v\n", building)
        }
    } else if soundFiles.Contains(lbxName) && !rawDump {
        for i := range len(file.Data) {
            func (){
                var name string
                if voc {
                    name = filepath.Join(dir, fmt.Sprintf("sound_%v.voc", i))
                } else {
                    name = filepath.Join(dir, fmt.Sprintf("sound_%v.wav", i))
                }
                out, err := os.Create(name)
                if err != nil {
                    fmt.Printf("Error creating sound file: %v\n", err)
                    return
                }
                defer out.Close()

                if voc {
                    err = audio.SaveVoc(out, &file, i)
                } else {
                    err = audio.SaveWav(out, &file, i)
                }
                if err != nil {
                    fmt.Printf("Error saving sound file: %v\n", err)
                    return
                }

                fmt.Printf("Saved sound %v to %v\n", i, name)
            }()
        }
    } else {
        func (){
            name := filepath.Join(dir, "strings.txt")
            out, err := os.Create(name)
            if err != nil {
                fmt.Printf("Error creating strings file: %v\n", err)
                return
            }
            defer out.Close()
            for i, str := range file.Strings {
                fmt.Fprintf(out, "%v: %v\n", i, str)
            }
        }()

        files := make(map[string]*lbx.LbxFile)
        files[lbxName] = &file
        cache := lbx.MakeCacheFromLbxFiles(files)

        customPaletteMap, err := lbx.GetPaletteOverrideMap(cache, &file, lbxName)
        if err != nil {
            log.Printf("Warning: unable to load custom palette map: %v\n", err)
            customPaletteMap = make(map[int]color.Palette)
        }

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

                palette := customPaletteMap[index]
                if palette == nil {
                    palette = customPaletteMap[-1]
                }

                images, err := file.ReadImagesWithPalette(index, palette, false)
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

    var zipName string
    var onlyIndex int
    var rawDump bool
    var voc bool

    flag.StringVar(&zipName, "zip", "", "Path to the zip file (optional)")
    flag.IntVar(&onlyIndex, "index", -1, "Only the file with the given index (optional)")
    flag.BoolVar(&rawDump, "raw", false, "Dump the files as binary (optional)")
    flag.BoolVar(&voc, "voc", false, "Dump audio files as voc rather than wav (optional)")
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: %v [options] filename\n\n", os.Args[0])
        fmt.Fprintln(os.Stderr, "Options:")
        flag.PrintDefaults()
        fmt.Fprintln(os.Stderr, "\nExample:")
        fmt.Fprintln(os.Stderr, "  ", os.Args[0], "--zip data.zip --index 0 --raw")
    }

    flag.Parse()

    positionalArgs := flag.Args()

    if len(positionalArgs) < 1 {
        flag.Usage()
        return
    }

    for _, path := range positionalArgs {

        if zipName == "" {
            fmt.Printf("Opening %v as an lbx file\n", path)

            file, err := os.Open(path)
            if err != nil {
                log.Printf("Error opening %v: %v\n", path, err)
                return
            }

            err = dumpLbx(file, strings.ToLower(filepath.Base(path)), onlyIndex, rawDump, voc)
            if err != nil {
                log.Printf("Error dumping lbx file: %v\n", err)
            }
        } else {
            zipFile, err := zip.OpenReader(zipName)
            if err != nil {
                fmt.Printf("Error opening zip file: %s\n", err)
                return
            }
            defer zipFile.Close()

            var matches []string
            for _, file := range zipFile.File {
                // fmt.Printf("Entry: %s\n", file.Name)

                lower := strings.ToLower(file.Name)
                check := strings.ToLower(path)

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
                fmt.Printf("No such entry with name '%v'\n", path)
                return
            }

            if len(matches) > 1 {
                fmt.Printf("More than one match found for '%v'\n", path)
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

                        err := dumpLbx(bytes.NewReader(memory.Bytes()), strings.ToLower(file.Name), onlyIndex, rawDump, voc)
                        if err != nil {
                            fmt.Printf("Error dumping lbx file: %v\n", err)
                        }
                        opened.Close()
                    }
                }
            }
        }
    }

}
