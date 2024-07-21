package lbx

import (
    "os"
)

type LbxCache struct {
    lbxFiles map[string]*LbxFile
}

func MakeLbxCache() *LbxCache {
    return &LbxCache{
        lbxFiles: make(map[string]*LbxFile),
    }
}

func (cache *LbxCache) GetLbxFile(filename string) (*LbxFile, error) {
    if lbxFile, ok := cache.lbxFiles[filename]; ok {
        return lbxFile, nil
    }

    file, err := os.Open(filename)
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
