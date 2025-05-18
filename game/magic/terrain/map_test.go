package terrain

import (
    "testing"
    "image"
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

func createTerrainData() *TerrainData {

    var tiles []TerrainTile
    for i, tile := range allTiles {
        tiles = append(tiles, TerrainTile{
            ImageIndex: 0,
            TileIndex: i,
            Tile: tile,
            Images: []image.Image{},
        })
    }

    return MakeTerrainData(nil, tiles)
}


func TestResolveRiverTiles (test *testing.T) {
    terrainData := createTerrainData()

    // TileRiver0101_1
    map_ := MakeMap(3, 3)
    map_.Terrain[0][0] = TileShore1_00000001.index
    map_.Terrain[0][1] = TileShore1_00000001.index
    map_.Terrain[0][2] = TileShore1_00000001.index
    map_.Terrain[1][0] = TileTundra.index
    map_.Terrain[1][1] = TileRiver0001.index
    map_.Terrain[1][2] = TileForest1.index
    map_.Terrain[2][0] = TileRiver0001.index
    map_.Terrain[2][1] = TileRiver0001.index
    map_.Terrain[2][2] = TileTundra.index
    tile, _ := map_.ResolveTile(1, 1, terrainData, data.PlaneArcanus)
    if tile != TileRiver0101_1.index {
        test.Errorf("should be TileRiver0101_1")
    }
}


func TestResolveLakeRiverTiles(test *testing.T) {
    terrainData := createTerrainData()

    // TileLakeRiverWest
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

    // TileLakeRiverNorth
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

    // TileLakeRiverSouth
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

    // TileLakeRiverEast
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


func TestResolveShoreRiverTiles(test *testing.T) {
    terrainData := createTerrainData()

    // TileShore_1R00000R
    map_ := MakeMap(3, 3)
    map_.Terrain[0][0] = TileDesert_00000000.index  // any land
    map_.Terrain[0][1] = TileRiver0001.index
    map_.Terrain[0][2] = TileOcean.index  // or shore
    map_.Terrain[1][0] = TileRiver0001.index
    map_.Terrain[1][1] = TileOcean.index  // or shore
    map_.Terrain[1][2] = TileOcean.index  // or shore
    map_.Terrain[2][0] = TileOcean.index  // or shore
    map_.Terrain[2][1] = TileOcean.index  // or shore
    map_.Terrain[2][2] = TileOcean.index  // or shore
    tile, _ := map_.ResolveTile(1, 1, terrainData, data.PlaneArcanus)
    if tile != TileShore_1R00000R.index {
        test.Errorf("should be TileShore_1R00000R not 0x%03x", tile)
    }

    // TileShore2_00011R11
    map_ = MakeMap(3, 3)
    map_.Terrain[0][0] = TileOcean.index  // or shore
    map_.Terrain[0][1] = TileDesert_00000000.index  // any land
    map_.Terrain[0][2] = TileDesert_00000000.index  // anything
    map_.Terrain[1][0] = TileOcean.index  // or shore
    map_.Terrain[1][1] = TileOcean.index  // or shore
    map_.Terrain[1][2] = TileRiver0001.index
    map_.Terrain[2][0] = TileOcean.index  // or shore
    map_.Terrain[2][1] = TileDesert_00000000.index  // any land
    map_.Terrain[2][2] = TileDesert_00000000.index  // anything
    tile, _ = map_.ResolveTile(1, 1, terrainData, data.PlaneArcanus)
    if tile != TileShore2_00011R11.index {
        test.Errorf("should be TileShore2_00011R11 not 0x%03x", tile)
    }
}

func TestContinents(test *testing.T) {
    use := MakeMap(30, 30)

    if len(use.FindContinents()) != 0 {
        test.Errorf("should be 0 continents")
    }


    for x := 5; x <= 6; x++ {
        for y := 5; y <= 6; y++ {
            use.Terrain[x][y] = TileLand.Index(data.PlaneArcanus)
        }
    }

    continents := use.FindContinents()
    if len(continents) != 1 {
        test.Errorf("should be 1 continent")
    }

    continent := continents[0]

    if continent.Size() != 4 {
        test.Errorf("should be 4 tiles in continent but was %v: %v", continent.Size(), continent)
    }

    if use.FindContinent(0, 0).Size() != 0 {
        test.Errorf("should be 0 tiles in continent")
    }

    if use.FindContinent(5, 5).Size() != 4 {
        test.Errorf("should be 4 tiles in continent")
    }

    for x := 10; x <= 12; x++ {
        for y := 10; y <= 12; y++ {
            use.Terrain[x][y] = TileLand.Index(data.PlaneArcanus)
        }
    }

    continents = use.FindContinents()
    if len(continents) != 2 {
        test.Errorf("should be 2 continents")
    }

    for _, continent := range continents {
        if continent.Size() != 4 && continent.Size() != 9 {
            test.Errorf("should be either 4 or 9 tiles in continent but was %v: %v", continent.Size(), continent)
        }
    }

}

func BenchmarkGeneration(bench *testing.B){
    terrainData := createTerrainData()
    plane := data.PlaneArcanus

    for i := 0; i < bench.N; i++ {
        GenerateLandCellularAutomata(100, 200, terrainData, plane)
    }
}
