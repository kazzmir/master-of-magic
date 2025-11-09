package maplib

import (
    "testing"
    "image"

    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

func TestMap(t *testing.T) {
    rawMap := terrain.MakeMap(10, 10)
    xmap := Map{
        Map: rawMap,
    }

    type Test struct {
        x1, x2, distance int
    }

    tests := []Test{
        {3,2,-1},
        {3,0,-3},
        {2,3,1},
        {3,2,-1},
        {9,0,1},
        {0,9,-1},
    }

    for _, test := range tests {
        distance := xmap.XDistance(test.x1, test.x2)
        if distance != test.distance {
            t.Errorf("Distance from %d to %d is %d, expected %d", test.x1, test.x2, distance, test.distance)
        }
    }
}

type FakeWizard struct {
}

func (fake *FakeWizard) GetBanner() data.BannerType {
    return data.BannerGreen
}

func TestInfluence(test *testing.T) {
    xmap := Map{
        ExtraMap: make(map[image.Point]map[ExtraKind]ExtraTile),
    }

    xmap.ExtraMap[image.Pt(2, 2)] = map[ExtraKind]ExtraTile{
        ExtraKindMagicNode: &ExtraMagicNode{
            Kind: MagicNodeNature,
            MeldingWizard: &FakeWizard{},
            Zone: []image.Point{image.Pt(0, 0), image.Pt(1, 0), image.Pt(3, 3)},
        },
    }

    node := xmap.GetMagicInfluence(5, 5)
    if node == nil || node.Kind.MagicType() != data.NatureMagic {
        test.Errorf("Expected NatureMagic at 5,5")
    }

    node = xmap.GetMagicInfluence(4, 4)
    if node != nil {
        test.Errorf("Expected no magic at 4,4")
    }
}

type TestCityProvider struct {
    Cities map[image.Point]bool
}

func (provider *TestCityProvider) ContainsCity(x int, y int, plane data.Plane) bool {
    _, exists := provider.Cities[image.Pt(x, y)]
    return exists
}

func TestCatchmentArea(test *testing.T) {
    testCityProvider := TestCityProvider{Cities: make(map[image.Point]bool)}
    terrainData := terrain.MakeTerrainData([]image.Image{nil}, []terrain.TerrainTile{terrain.TerrainTile{TileIndex: 0, Tile: terrain.TileLand}})
    rawMap := terrain.MakeMap(20, 20)
    xmap := Map{
        Data: terrainData,
        Plane: data.PlaneArcanus,
        Map: rawMap,
        CityProvider: &testCityProvider,
        ExtraMap: make(map[image.Point]map[ExtraKind]ExtraTile),
    }

    catchment := xmap.GetCatchmentArea(5, 5)
    if len(catchment) != 21 {
        test.Errorf("Expected catchment area of size 21, got %d", len(catchment))
    }

    for _, tile := range catchment {
        if tile.IsShared {
            test.Errorf("Expected no shared tiles in catchment area")
        }
    }

    // two cities do not overlap at all
    testCityProvider.Cities[image.Pt(5, 5)] = true
    testCityProvider.Cities[image.Pt(19, 5)] = true

    catchment = xmap.GetCatchmentArea(5, 5)
    for _, tile := range catchment {
        if tile.IsShared {
            test.Errorf("Expected no shared tiles in catchment area with two non-overlapping cities")
        }
    }

    testCityProvider.Cities[image.Pt(6, 6)] = true
    catchment = xmap.GetCatchmentArea(5, 5)

    count := 0
    for _, tile := range catchment {
        if tile.IsShared {
            count += 1
        }
    }

    if count != 14 {
        test.Errorf("Expected 6 shared tiles in catchment area with overlapping cities, got %d", count)
    }
}

func TestTile(test *testing.T) {
    tileOcean := FullTile{
        Tile: terrain.TileOcean,
    }

    if !tileOcean.CanTraverse(terrain.East, TraverseWater) {
        test.Errorf("Expected ocean to be traversable by water")
    }

    if tileOcean.CanTraverse(terrain.West, TraverseLand) {
        test.Errorf("Expected ocean to not be traversable by land")
    }

    tile2 := FullTile{
        Tile: terrain.TileLand,
    }

    if !tile2.CanTraverse(terrain.North, TraverseLand) {
        test.Errorf("Expected land to be traversable by land")
    }

    if tile2.CanTraverse(terrain.South, TraverseWater) {
        test.Errorf("Expected land to not be traversable by water")
    }

    // land on all east positions
    tileShoreRight2 := FullTile{
        Tile: terrain.TileShore1_00011100,
    }

    if !tileShoreRight2.CanTraverse(terrain.East, TraverseLand) {
        test.Errorf("Expected shore tile to be traversable by land on east side")
    }

    if tileShoreRight2.CanTraverse(terrain.West, TraverseLand) {
        test.Errorf("Expected shore tile to not be traversable by land on west side")
    }

    if tileShoreRight2.CanTraverse(terrain.East, TraverseWater) {
        test.Errorf("Expected shore tile to not be traversable by water on east side")
    }

    if !tileShoreRight2.CanTraverse(terrain.West, TraverseWater) {
        test.Errorf("Expected shore tile to be traversable by water on west side")
    }

    // land on all west positions
    tileShoreLeft := FullTile{
        Tile: terrain.TileShore1_11000001,
    }

    if !tileShoreLeft.CanTraverse(terrain.West, TraverseLand) {
        test.Errorf("Expected shore tile to be traversable by land on west side")
    }

    if tileShoreLeft.CanTraverse(terrain.East, TraverseLand) {
        test.Errorf("Expected shore tile to not be traversable by land on east side")
    }

    if tileShoreLeft.CanTraverse(terrain.West, TraverseWater) {
        test.Errorf("Expected shore tile to not be traversable by water on west side")
    }

    if !tileShoreLeft.CanTraverse(terrain.East, TraverseWater) {
        test.Errorf("Expected shore tile to be traversable by water on east side")
    }

    /*
    tile1 := FullTile{
        X: 0,
        Y: 0,
        Tile: terrain.TileOcean,
    }

    tile2 := FullTile{
        X: 1,
        Y: 0,
        Tile: terrain.TileOcean,
    }

    tile3 := FullTile{
        X: 1,
        Y: 0,
        Tile: terrain.TileLand,
    }

    if !tile1.IsConnected(&tile2, ConnectedWater) {
        test.Errorf("Expected tile1 and tile2 to be connected")
    }

    if tile1.IsConnected(&tile3, ConnectedWater) {
        test.Errorf("Expected tile1 and tile3 to not be connected")
    }

    // a shore tile where the land is on the right side
    tileShoreRight := FullTile{
        X: 1,
        Y: 0,
        Tile: terrain.TileShore1_00011100,
    }

    if !tile1.IsConnected(&tileShoreRight, ConnectedWater) {
        test.Errorf("Expected tile1 and tileShoreRight to be connected")
    }

    if !tileShoreRight.IsConnected(&tile1, ConnectedWater) {
        test.Errorf("Expected tileShoreRight and tile1 to be connected")
    }

    // land on all east positions
    tileShoreRight2 := FullTile{
        X: 2,
        Y: 0,
        Tile: terrain.TileShore1_00011100,
    }

    if tile3.IsConnected(&tileShoreRight2, ConnectedLand) {
        test.Errorf("Expected tile3 and tileShoreRight2 to not be connected")
    }

    // land on all west positions
    tileShoreLeft := FullTile{
        X: 3,
        Y: 0,
        Tile: terrain.TileShore1_11000001,
    }

    if tileShoreRight.IsConnected(&tileShoreLeft, ConnectedWater) {
        test.Errorf("Expected tileShoreRight and tileShoreLeft to not be connected")
    }

    if !tileShoreRight.IsConnected(&tileShoreLeft, ConnectedLand) {
        test.Errorf("Expected tileShoreRight and tileShoreLeft to be connected by land")
    }
    */
}
