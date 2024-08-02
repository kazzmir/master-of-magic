package lbx

import (
    "os"
    "strings"
    "path/filepath"
)

type LbxCache struct {
    lbxFiles map[string]*LbxFile
    Base string
}

func MakeLbxCache(basedir string) *LbxCache {
    return &LbxCache{
        Base: basedir,
        lbxFiles: make(map[string]*LbxFile),
    }
}

func (cache *LbxCache) GetLbxFile(filename string) (*LbxFile, error) {
    filename = strings.ToUpper(filename)

    if lbxFile, ok := cache.lbxFiles[filename]; ok {
        return lbxFile, nil
    }

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
}
