package terrain

import (
    "bytes"

    "github.com/kazzmir/master-of-magic/lib/lbx"
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
