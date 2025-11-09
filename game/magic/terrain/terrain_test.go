package terrain

import (
    "testing"
    "github.com/kazzmir/master-of-magic/lib/set"
)


func TestTerrainType(test *testing.T) {
    for _, tile := range allTiles {
        if tile.TerrainType() == Unknown {
            test.Errorf("TerrainType of %v is unknown", tile)
        }
    }
}

func TestTerrainMatches(test *testing.T) {
    compatibilities := make(map[Direction]Compatibility)
    match := make(map[Direction]TerrainType)
    tile := Tile{
        index: 0,
        Compatibilities: compatibilities,
    }

    // No compatibilities
    if !tile.Matches(match) {
        test.Errorf("%v should match %v", match, tile)
    }

    // AnyOf
    tile.Compatibilities[North] = Compatibility{
        Terrains: set.NewSet([]TerrainType{Ocean, Shore}...),
        Type: AnyOf,
    }
    match[North] = Ocean
    if !tile.Matches(match) {
        test.Errorf("%v should match %v", match, tile)
    }

    tile.Compatibilities[North] = Compatibility{
        Terrains: set.NewSet([]TerrainType{Ocean, Shore}...),
        Type: AnyOf,
    }
    match[North] = Shore
    if !tile.Matches(match) {
        test.Errorf("%v should match %v", match, tile)
    }

    tile.Compatibilities[North] = Compatibility{
        Terrains: set.NewSet([]TerrainType{Ocean, Shore}...),
        Type: AnyOf,
    }
    match[North] = Grass
    if tile.Matches(match) {
        test.Errorf("%v should not match %v", match, tile)
    }

    // NoneOf
    tile.Compatibilities[North] = Compatibility{
        Terrains: set.NewSet([]TerrainType{Ocean, Shore}...),
        Type: NoneOf,
    }
    match[North] = Ocean
    if tile.Matches(match) {
        test.Errorf("%v should not match %v", match, tile)
    }

    tile.Compatibilities[North] = Compatibility{
        Terrains: set.NewSet([]TerrainType{Ocean, Shore}...),
        Type: NoneOf,
    }
    match[North] = Shore
    if tile.Matches(match) {
        test.Errorf("%v should not match %v", match, tile)
    }

    tile.Compatibilities[North] = Compatibility{
        Terrains: set.NewSet([]TerrainType{Ocean, Shore}...),
        Type: NoneOf,
    }
    match[North] = Grass
    if !tile.Matches(match) {
        test.Errorf("%v should match %v", match, tile)
    }

    // two directions
    tile.Compatibilities[North] = Compatibility{
        Terrains: set.NewSet([]TerrainType{Ocean, Shore}...),
        Type: AnyOf,
    }
    tile.Compatibilities[South] = Compatibility{
        Terrains: set.NewSet([]TerrainType{Ocean, Shore}...),
        Type: NoneOf,
    }

    match[North] = Shore
    match[South] = Grass
    if !tile.Matches(match) {
        test.Errorf("%v should match %v", match, tile)
    }

    match[North] = Shore
    match[South] = Shore
    if tile.Matches(match) {
        test.Errorf("%v should not match %v", match, tile)
    }

    match[North] = Grass
    match[South] = Grass
    if tile.Matches(match) {
        test.Errorf("%v should not match %v", match, tile)
    }
}
