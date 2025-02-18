package hero

// read the names of the heroes per wizard from names.lbx
// each wizarc has their own set of heroes, and each wizard gives a unique name to their hero

import (
    "log"
    "fmt"
    "io"

    "github.com/kazzmir/master-of-magic/lib/lbx"
)

func ReadNames(cache *lbx.LbxCache) ([]string, error) {
    lbxFile, err := cache.GetLbxFile("names.lbx")
    if err != nil {
        return nil, err
    }

    reader, err := lbxFile.GetReader(0)
    if err != nil {
        return nil, err
    }

    count, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    if count > 10000 {
        return nil, fmt.Errorf("Name count was too high: %v", count)
    }

    size, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    if size > 10000 {
        return nil, fmt.Errorf("Size of each name entry was too high: %v", size)
    }

    var out []string

    data := make([]byte, size)
    for range count {
        n, err := io.ReadFull(reader, data)
        if err != nil {
            return nil, err
        }
        if n != len(data) {
            return nil, fmt.Errorf("Failed to read all of the name data (%v)", n)
        }

        log.Printf("Name: '%v'", string(data))
        out = append(out, string(data))
    }

    return out, nil
}
