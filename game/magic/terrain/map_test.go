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
