package combat

import (
    "math/rand"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/hajimehoshi/ebiten/v2"
)

type CombatState int

const (
    CombatStateRunning CombatState = iota
    CombatStateDone
)

type Tile struct {
    // index of grass/floor
    Index int
    // tree/rock on top, or -1 if nothing
    ExtraObject int
}

type Army struct {
    Units []*units.Unit
}

type CombatScreen struct {
    Cache *lbx.LbxCache
    ImageCache util.ImageCache
    DefendingArmy *Army
    AttackingArmy *Army
    Tiles [][]Tile
}

func makeTiles(width int, height int) [][]Tile {

    maybeExtraTile := func() int {
        if rand.Intn(10) == 0 {
            return rand.Intn(10)
        }
        return -1
    }

    tiles := make([][]Tile, width)
    for x := 0; x < len(tiles); x++ {
        tiles[x] = make([]Tile, height)
        for y := 0; y < len(tiles[x]); y++ {
            tiles[x][y] = Tile{
                Index: rand.Intn(48),
                ExtraObject: maybeExtraTile(),
            }
        }

    }

    return tiles
}

func MakeCombatScreen(cache *lbx.LbxCache, defendingArmy *Army, attackingArmy *Army) *CombatScreen {
    return &CombatScreen{
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
        DefendingArmy: defendingArmy,
        AttackingArmy: attackingArmy,
        Tiles: makeTiles(50, 50),
    }
}

func (combat *CombatScreen) Update() CombatState {
    return CombatStateRunning
}

func (combat *CombatScreen) Draw(screen *ebiten.Image){
}
