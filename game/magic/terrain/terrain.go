package terrain

import (
    "bytes"

    "github.com/kazzmir/master-of-magic/lib/lbx"
)

// terrain tiles are indicies 0-0x259 for arcanus, and 0x25A - 0x5f4 for myrror

type TerrainIndex int

const (
    IndexOcean1 TerrainIndex      = 0x0
    IndexBugGrass    = 0x1
    IndexShore1_1st  = 0x2
    IndexLake        = 0x12
    IndexShore1_end  = 0x0A1
    IndexGrass1      = 0x0A2
    IndexForest1     = 0x0A3
    IndexMountain1   = 0x0A4
    IndexDesert1     = 0x0A5
    IndexSwamp1      = 0x0A6
    IndexTundra1     = 0x0A7
    IndexSorcNode    = 0x0A8
    IndexNatNode     = 0x0A9
    IndexChaosNode   = 0x0AA
    IndexHills1      = 0x0AB
    IndexGrass2      = 0x0AC
    IndexGrass3      = 0x0AD
    IndexDesert2     = 0x0AE
    IndexDesert3     = 0x0AF
    IndexDesert4     = 0x0B0
    IndexSwamp2      = 0x0B1
    IndexSwamp3      = 0x0B2
    IndexVolcano     = 0x0B3
    IndexGrass4      = 0x0B4
    IndexTundra2     = 0x0B5
    IndexTundra3     = 0x0B6
    IndexForest2     = 0x0B7
    IndexForest3     = 0x0B8
    IndexRiverMStart  = 0x0B9
    IndexRiverMEnd  = 0x0C4
    IndexLake1       = 0x0C5
    IndexLake2       = 0x0C6
    IndexLake3       = 0x0C7
    IndexLake4       = 0x0C8
    IndexShore2FStart = 0x0C9
    IndexShore2FEnd = 0x0E8
    IndexRiversStart  = 0x0E9
    IndexRiversEnd  = 0x102
    IndexMountainsStart   = 0x103
    IndexMountainsEnd = 0x112
    IndexHillsStart   = 0x113
    IndexHillsEnd   = 0x123
    IndexDesertStart  = 0x124
    IndexDesertEnd  = 0x1C3
    IndexShore2Start  = 0x1C4
    IndexShore2End  = 0x1D3
    Index4WRiver1    = 0x1D4
    Index4WRiver2    = 0x1D5
    Index4WRiver3    = 0x1D6
    Index4WRiver4    = 0x1D7
    Index4WRiver5    = 0x1D8
    IndexShore3Start  = 0x1D9
    IndexShore3End  = 0x258
    IndexOcean2      = 0x259
    IndexTundra_1st  = 0x25A
    IndexTundra_Last = 0x2F9
)

type TerrainData struct {
    Images []image.Image
    Tiles []TerrainTile
}

type TerrainTile struct {
    Image int
    Animiation bool
}

// pass in terrain.lbx
func ReadTerrainData(lbxFile *LbxFile) (*TerrainData, error) {
    data, err := lbxFile.RawData(1)
    if err != nil {
        return nil, err
    }

    reader := bytes.NewReader(data)

    var tiles []TerrainTile

    for reader.Len() > 0 {
        var animation = false
        value1, err := reader.ReadByte()
        if err != nil {
            return nil, err
        }

        value2, err := reader.ReadByte()
        if err != nil {
            return nil, err
        }

        if value1 & 0x80 != 0 {
            value1 &= 0x7f
            animation = true
            // value2 should be incremented based on an animation counter
        }

        // the index into the terrain image
        var index int

        if value1 == 0 {
            index = value2 - 2
        } else {
            // this formula comes from
            // terrain_lbx_000_offset = (terrain_001_0 * 16384) + (terrain_001_1 * 384) - 0xC0 - 384;
            // this only works if value1 is a multiple of 3
            // 3 -> 126, 6 -> 254, 9 -> 382
            index = value1 * 16384 / 384 + value2 - 2
        }

        tiles = append(tiles, TerrainTile{
            Image: index,
            Animation: animation,
        })
    }

    images, err := lbxFile.ReadTerrainImages(0)
    if err != nil {
        return nil, err
    }

    return &TerrainData{
        Images: images,
        Tiles: tiles,
    }
}
