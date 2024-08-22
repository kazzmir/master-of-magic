package lbx

import (
    "os"
    "fmt"
    "log"
    "io"
    "io/fs"
    "strings"
    // "path/filepath"
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

func AutoCache() *LbxCache {
    // Find where the data is
    // 1. check in the current working directory for all existing lbx files
    // 2. check all the directories in the current working directory to see if any of them contain lbx files
    // 3. look at all zip files in the current working directory to see if any of the zip files contain lbx files
    // 4. possibly use an embedded fs with all data in it
    return MakeLbxCache(os.DirFS("magic-data"))
}

func createReadSeeker(reader fs.File) io.ReadSeeker {
    if readSeeker, ok := reader.(io.ReadSeeker); ok {
        return readSeeker
    }

    log.Printf("FIXME: create read seeker for %v", reader)
    return nil
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

            reader := createReadSeeker(file)

            lbxFile, err := ReadLbx(reader)
            if err != nil {
                return nil, err
            }

            cache.lbxFiles[filename] = &lbxFile

            return cache.lbxFiles[filename], nil
        }
    }

    return nil, fmt.Errorf("%v lbx file not found", filename)

    /*
    full := filepath.Join(cache.Base, filename)

    // FIXME: do this case-insensitive
    file, err := os.Open(full)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    lbxFile, err := ReadLbx(file)
    if err != nil {
        return nil, err
    }
    cache.lbxFiles[filename] = &lbxFile

    return cache.lbxFiles[filename], nil
    */
}
