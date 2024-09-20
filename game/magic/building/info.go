package building

import (
    "fmt"
    "bytes"
    "github.com/kazzmir/master-of-magic/lib/lbx"
)

func ReadBuildingInfo(cache *lbx.LbxCache) error {
    data, err := cache.GetLbxFile("builddat.lbx")
    if err != nil {
        return fmt.Errorf("Unable to read builddat.lbx: %v", err)
    }

    reader, err := data.GetReader(0)
    if err != nil {
        return fmt.Errorf("unable to read entry 0 in builddat.lbx: %v", err)
    }

    numBuildings, err := lbx.ReadUint16(reader)
    if err != nil {
        return fmt.Errorf("read error: %v", err)
    }

    entrySize, err := lbx.ReadUint16(reader)
    if err != nil {
        return fmt.Errorf("read error: %v", err)
    }

    for i := 0; i < int(numBuildings); i++ {
        buildingData := make([]byte, entrySize)
        n, err := reader.Read(buildingData)
        if err != nil || n != int(entrySize) {
            return fmt.Errorf("unable to read building info %v: %v", i, err)
        }

        buildingReader := bytes.NewReader(buildingData)
        name := make([]byte, 20)
        _, err = buildingReader.Read(name)
        if err != nil {
            return fmt.Errorf("unable to read building name %v: %v", i, err)
        }

        buildingBefore, err := lbx.ReadUint16(buildingReader)
        if err != nil {
            return fmt.Errorf("unable to read building before %v: %v", i, err)
        }

        fmt.Printf("Building %v: name='%v' before=%v\n", i, name, buildingBefore)
    }

    return nil
}
