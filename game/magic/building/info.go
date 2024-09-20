package building

import (
    "fmt"
    "bytes"
    "github.com/kazzmir/master-of-magic/lib/lbx"
)

func parseName(name []byte) string {
    for i, b := range name {
        if b == 0 {
            return string(name[:i])
        }
    }
    return string(name)
}

const ForrestTerrain = 101
const WaterTerrain = 110
const MineralTerrain = 200

type BuildingInfo struct {
    Name string
    BuildingDependency1 int
    BuildingDependency2 int
    // -1 for no terrain, otherwise specifies a tile index that the building can be built on
    TerrainDependency int
    BuildingReplace int
}

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

        dependency1, err := lbx.ReadUint16(buildingReader)
        if err != nil {
            return fmt.Errorf("unable to read building dependency1 %v: %v", i, err)
        }

        dependency2, err := lbx.ReadUint16(buildingReader)
        if err != nil {
            return fmt.Errorf("unable to read building dependency2 %v: %v", i, err)
        }

        replace, err := lbx.ReadUint16(buildingReader)
        if err != nil {
            return fmt.Errorf("unable to read building replace %v: %v", i, err)
        }

        // fmt.Printf("Building %v: name='%v' dependency1=%v dependency2=%v\n", i, parseName(name), dependency1, dependency2)

        terrainDependency := uint16(0)
        if dependency1 > 100 {
            terrainDependency = dependency1
            dependency1 = 0
        }

        info := BuildingInfo{
            Name: parseName(name),
            BuildingDependency1: int(dependency1),
            BuildingDependency2: int(dependency2),
            TerrainDependency: int(terrainDependency),
            BuildingReplace: int(replace),
        }

        fmt.Printf("Building %v: %+v\n", i, info)
    }

    return nil
}
