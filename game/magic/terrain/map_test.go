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

func TestPolarIceRowsSkippedOnTinyMaps(test *testing.T) {
    if PolarIceRows(3) != 0 {
        test.Errorf("3-row maps should not be painted as ice")
    }
    if PolarIceRows(50) != 1 {
        test.Errorf("standard maps should have a 1-row ice cap")
    }
    if IsPolarIceRow(0, 50) != true || IsPolarIceRow(49, 50) != true {
        test.Errorf("y=0 and y=height-1 should be the ice cap")
    }
    if IsPolarIceRow(1, 50) || IsPolarIceRow(48, 50) {
        test.Errorf("tiles adjacent to the ice cap should not be the cap")
    }
}

func TestPlacePolarIcePaintsFullWidthCaps(test *testing.T) {
    rows := 10
    columns := 12
    map_ := MakeMap(rows, columns)
    ocean := TileOcean.Index(data.PlaneArcanus)
    for x := 0; x < columns; x++ {
        for y := 0; y < rows; y++ {
            map_.Terrain[x][y] = ocean
        }
    }

    map_.placePolarIce(data.PlaneArcanus)

    ice := PolarIceRows(rows)
    tundra := TileTundra.Index(data.PlaneArcanus)
    for x := 0; x < columns; x++ {
        for y := 0; y < ice; y++ {
            if map_.Terrain[x][y] != tundra {
                test.Errorf("north cap %d,%d want tundra", x, y)
            }
            if map_.Terrain[x][rows-1-y] != tundra {
                test.Errorf("south cap %d,%d want tundra", x, rows-1-y)
            }
        }
        if map_.Terrain[x][ice] != ocean {
            test.Errorf("row inside the north cap should stay ocean at x=%d", x)
        }
    }
}

func TestGenerateLandHasPolarIceCaps(test *testing.T) {
    terrainData := createTerrainData()
    map_ := GenerateLandCellularAutomata(20, 20, terrainData, data.PlaneArcanus)
    ice := PolarIceRows(20)
    for x := 0; x < 20; x++ {
        for y := 0; y < ice; y++ {
            if GetTile(map_.Terrain[x][y]).TerrainType() != Tundra {
                test.Errorf("north cap %d,%d is %v", x, y, GetTile(map_.Terrain[x][y]).TerrainType())
            }
            if GetTile(map_.Terrain[x][19-y]).TerrainType() != Tundra {
                test.Errorf("south cap %d,%d is %v", x, 19-y, GetTile(map_.Terrain[x][19-y]).TerrainType())
            }
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
