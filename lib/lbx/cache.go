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
func validateData(data fs.FS) bool {
    entries, err := fs.ReadDir(data, ".")
    if err != nil {
        return false
    }

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

func AutoCache() *LbxCache {
    // Find where the data is
    // 1. check in the current working directory for all existing lbx files
    // 2. check all the directories in the current working directory to see if any of them contain lbx files
    // 3. look at all zip files in the current working directory to see if any of the zip files contain lbx files
    // 4. possibly use an embedded fs with all data in it

    byteReader := bytes.NewReader(data.DataZip)
    zipReader, err := zip.NewReader(byteReader, int64(len(data.DataZip)))
    if err == nil && validateData(zipReader) {
        log.Printf("Found data in embedded zip")
        return MakeLbxCache(zipReader)
    }

    here := os.DirFS(".")
    if validateData(here) {
        log.Printf("Found data in .")
        return MakeLbxCache(here)
    }

    entries, err := os.ReadDir(".")
    if err == nil {
        for _, entry := range entries {
            if entry.IsDir() {
                check := os.DirFS(entry.Name())
                if validateData(check) {
                    log.Printf("Found data in %v", entry.Name())
                    return MakeLbxCache(check)
                }
            } else if strings.HasSuffix(entry.Name(), ".zip") {
                zipper, err := zip.OpenReader(entry.Name())
                if err == nil {
                    if validateData(zipper) {
                        log.Printf("Found data in zip file %v", entry.Name())
                        return MakeLbxCache(zipper)
                    } else {
                        zipper.Close()
                    }
                }
            }
        }
    }

    log.Printf("Unable to find data")

    return nil
}

func createReadSeeker(reader fs.File) (io.ReadSeeker, error) {
    if readSeeker, ok := reader.(io.ReadSeeker); ok {
        return readSeeker, nil
    }

    rawData, err := io.ReadAll(reader)
    if err != nil {
        return nil, err
    }

    data := bytes.NewReader(rawData)

    return data, nil
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
