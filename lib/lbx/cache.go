package lbx

import (
    "os"
    "fmt"
    "log"
    "bytes"
    "io"
    "io/fs"
    "strings"
    "archive/zip"
    // "path/filepath"

    "github.com/kazzmir/master-of-magic/data"
    "github.com/hajimehoshi/ebiten/v2"
)

type LbxCache struct {
    lbxFiles map[string]*LbxFile
    Base fs.FS
}

func MakeLbxCache(base fs.FS) *LbxCache {
    return &LbxCache{
        Base: base,
        lbxFiles: make(map[string]*LbxFile),
    }
}

/* true if the fs contains the proper lbx files */
func validateData(entries []fs.DirEntry) bool {
    /*
    entries, err := fs.ReadDir(data, ".")
    if err != nil {
        return false
    }
    */

    // subset of the required files, add more if necessary
    required := make(map[string]bool)
    required["MAGIC.LBX"] = true
    required["BACKGRND.LBX"] = true
    required["INTRO.LBX"] = true
    required["CITYSCAP.LBX"] = true
    required["FIGURE10.LBX"] = true
    required["UNITS1.LBX"] = true
    required["UNITS2.LBX"] = true
    required["TERRAIN.LBX"] = true

    count := 0

    for _, entry := range entries {
        _, in := required[strings.ToUpper(entry.Name())]
        if in {
            count += 1
        }
    }

    if count == len(required) {
        return true
    }

    return false
}

func maybeOpenZip(reader io.Reader) *LbxCache {

    /*
    info, err := reader.Stat()
    if err != nil {
        reader.Close()
        return nil
    }
    */

    byteReader, err := makeByteReader(reader)
    if err != nil {
        return nil
    }

    zipper, err := zip.NewReader(byteReader, byteReader.Size())
    if err == nil {
        entries, err := fs.ReadDir(zipper, ".")
        if err == nil && validateData(entries) {
            return MakeLbxCache(zipper)
        }
    }

    return nil
}

func searchFs(here fs.FS, context string) *LbxCache {
    entries, err := fs.ReadDir(here, ".")
    if err == nil {
        if validateData(entries){
            return MakeLbxCache(here)
        }

        // log.Printf("Check entries: %v", entries)
        for _, entry := range entries {
            if entry.IsDir() {
                // check := os.DirFS(entry.Name())
                check, err := fs.Sub(here, entry.Name())
                if err == nil {
                    entries, err := fs.ReadDir(check, ".")
                    if err == nil && validateData(entries) {
                        log.Printf("[%v] Found data in %v", context, entry.Name())
                        return MakeLbxCache(check)
                    }
                }
            } else if strings.HasSuffix(strings.ToLower(entry.Name()), ".zip") {
                // log.Printf("Check zip file '%v'", entry.Name())

                reader, err := here.Open(entry.Name())
                if err == nil {
                    cache := maybeOpenZip(reader)
                    if cache != nil {
                        log.Printf("[%v] Found data in zip file %v", context, entry.Name())
                        reader.Close()
                        return cache
                    }

                    reader.Close()
                } else {
                    log.Printf("[%v] Unable to open zip file %v: %v", context, entry.Name(), err)
                }
            }
        }
    }

    return nil
}

func CacheFromPath(path string) *LbxCache {
    if strings.HasSuffix(strings.ToLower(path), ".zip") {
        reader, err := os.Open(path)
        if err != nil {
            log.Printf("Unable to open %v: %v", path, err)
            return nil
        }
        defer reader.Close()
        return maybeOpenZip(reader)
    } else {
        return searchFs(os.DirFS(path), "filesystem")
    }
}

func AutoCache() *LbxCache {
    // Find where the data is
    // 1. possibly use an embedded fs with all data in it, if it exists
    // 2. check whats in the dropped files (for desktop and browsers)
    // 3. check in the current working directory for all existing lbx files
    // 4. check all the directories in the current working directory to see if any of them contain lbx files
    // 5. look at all zip files in the current working directory to see if any of the zip files contain lbx files

    /*
    byteReader := bytes.NewReader(data.DataZip)
    zipReader, err := zip.NewReader(byteReader, int64(len(data.DataZip)))
    if err == nil {
        entries, err := fs.ReadDir(zipReader, ".")
        if err == nil && validateData(entries) {
            log.Printf("Found data in embedded zip")
            return MakeLbxCache(zipReader)
        }
    }
    */
    embeddedCache := searchFs(data.Data, "embedded")
    if embeddedCache != nil {
        return embeddedCache
    }

    droppedFiles := ebiten.DroppedFiles()
    if droppedFiles != nil {
        cache := searchFs(droppedFiles, "dropped")
        if cache != nil {
            return cache
        }
    }

    cache := searchFs(os.DirFS("."), "filesystem")
    if cache != nil {
        return cache
    }

    // log.Printf("Unable to find data")

    return nil
}

func makeByteReader(reader io.Reader) (*bytes.Reader, error) {
    rawData, err := io.ReadAll(reader)
    if err != nil {
        return nil, err
    }

    data := bytes.NewReader(rawData)

    return data, nil
}

func createReadSeeker(reader fs.File) (io.ReadSeeker, error) {
    if readSeeker, ok := reader.(io.ReadSeeker); ok {
        return readSeeker, nil
    }

    return makeByteReader(reader)
}

func (cache *LbxCache) GetLbxFile(filename string) (*LbxFile, error) {
    filename = strings.ToUpper(filename)

    if lbxFile, ok := cache.lbxFiles[filename]; ok {
        return lbxFile, nil
    }

    entries, err := fs.ReadDir(cache.Base, ".")
    if err != nil {
        return nil, err
    }

    for _, entry := range entries {
        name := strings.ToUpper(entry.Name())
        if name == filename {
            file, err := cache.Base.Open(entry.Name())
            if err != nil {
                return nil, err
            }
            defer file.Close()

            reader, err := createReadSeeker(file)
            if err != nil {
                return nil, err
            }

            lbxFile, err := ReadLbx(reader)
            if err != nil {
                return nil, err
            }

            cache.lbxFiles[filename] = &lbxFile

            return cache.lbxFiles[filename], nil
        }
    }

    return nil, fmt.Errorf("%v lbx file not found", filename)
}
