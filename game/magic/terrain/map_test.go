package terrain

import (
    "testing"
)


func TestTerrainType(test *testing.T) {
    for _, tile := range allTiles {
        if tile.TerrainType() == Unknown {
            test.Errorf("TerrainType of %v is unknown", tile)
        }
    }
}
