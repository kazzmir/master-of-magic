package terrain

import (
    "testing"
    "os"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/data"
)


func TestTerrainType(test *testing.T) {
    for _, tile := range allTiles {
        if tile.TerrainType() == Unknown {
            test.Errorf("TerrainType of %v is unknown", tile)
        }
    }
}

func TestResolveTile(test *testing.T) {
    file, err := os.Open("/Users/marc/Documents/Repos/master-of-magic/data/data/TERRAIN.LBX")
    if err != nil {
        test.Skip("data file not found")
    }
    lbxFile, _ := lbx.ReadLbx(file)
    terrainData, _ := ReadTerrainData(&lbxFile)

    map_ := MakeMap(3, 3)
    map_.Terrain[0][1] = TileRiver0001.index
    map_.Terrain[1][0] = TileGrasslands1.index
    map_.Terrain[1][1] = TileLake.index
    map_.Terrain[1][2] = TileGrasslands1.index
    map_.Terrain[2][1] = TileGrasslands1.index
    tile, _ := map_.ResolveTile(1, 1, terrainData, data.PlaneArcanus)
    if tile != TileLakeRiverWest.index {
        test.Errorf("should be TileLakeRiverWest")
    }

    map_ = MakeMap(3, 3)
    map_.Terrain[0][1] = TileGrasslands1.index
    map_.Terrain[1][0] = TileRiver0001.index
    map_.Terrain[1][1] = TileLake.index
    map_.Terrain[1][2] = TileGrasslands1.index
    map_.Terrain[2][1] = TileGrasslands1.index
    tile, _ = map_.ResolveTile(1, 1, terrainData, data.PlaneArcanus)
    if tile != TileLakeRiverNorth.index {
        test.Errorf("should be TileLakeRiverNorth")
    }

    map_ = MakeMap(3, 3)
    map_.Terrain[0][1] = TileGrasslands1.index
    map_.Terrain[1][0] = TileGrasslands1.index
    map_.Terrain[1][1] = TileLake.index
    map_.Terrain[1][2] = TileRiver0001.index
    map_.Terrain[2][1] = TileGrasslands1.index
    tile, _ = map_.ResolveTile(1, 1, terrainData, data.PlaneArcanus)
    if tile != TileLakeRiverSouth.index {
        test.Errorf("should be TileLakeRiverSouth")
    }

    map_ = MakeMap(3, 3)
    map_.Terrain[0][1] = TileGrasslands1.index
    map_.Terrain[1][0] = TileGrasslands1.index
    map_.Terrain[1][1] = TileLake.index
    map_.Terrain[1][2] = TileGrasslands1.index
    map_.Terrain[2][1] = TileRiver0001.index
    tile, _ = map_.ResolveTile(1, 1, terrainData, data.PlaneArcanus)
    if tile != TileLakeRiverEast.index {
        test.Errorf("should be TileLakeRiverEast")
    }

}
