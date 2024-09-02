package combat

import (
    // "log"
    // "math"
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

type ArmyUnit struct {
    Unit units.Unit
    Facing units.Facing
    X int
    Y int
}

type Army struct {
    Units []*ArmyUnit
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

    tiles := make([][]Tile, height)
    for y := 0; y < len(tiles); y++ {
        tiles[y] = make([]Tile, width)
        for x := 0; x < len(tiles[y]); x++ {
            tiles[y][x] = Tile{
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
        Tiles: makeTiles(20, 30),
    }
}

func (combat *CombatScreen) Update() CombatState {
    return CombatStateRunning
}

func (combat *CombatScreen) Draw(screen *ebiten.Image){

    var options ebiten.DrawImageOptions

    tile0, _ := combat.ImageCache.GetImage("cmbgrass.lbx", 0, 0)

    tilePosition := func(x int, y int) (float64, float64){
        startX := 0
        if y % 2 == 1 {
            startX = -tile0.Bounds().Dx() / 2
        }

        return float64(x * tile0.Bounds().Dx() + startX), float64(y * tile0.Bounds().Dy() / 2 - tile0.Bounds().Dy() / 2)
    }

    // draw base land
    for y := 0; y < len(combat.Tiles); y++ {
        for x := 0; x < len(combat.Tiles[y]); x++ {
            image, _ := combat.ImageCache.GetImage("cmbgrass.lbx", combat.Tiles[y][x].Index, 0)
            options.GeoM.Reset()
            // options.GeoM.Rotate(math.Pi/2)
            tx, ty := tilePosition(x, y)
            options.GeoM.Translate(tx, ty)
            screen.DrawImage(image, &options)
        }
    }

    // draw extra trees/rocks on top
    for y := 0; y < len(combat.Tiles); y++ {
        for x := 0; x < len(combat.Tiles[y]); x++ {
            options.GeoM.Reset()
            tx, ty := tilePosition(x, y)
            options.GeoM.Translate(tx, ty)

            if combat.Tiles[y][x].ExtraObject != -1 {
                extraImage, _ := combat.ImageCache.GetImage("cmbgrass.lbx", 48 + combat.Tiles[y][x].ExtraObject, 0)
                screen.DrawImage(extraImage, &options)
            }
        }
    }

    for _, unit := range combat.DefendingArmy.Units {
        combatImage, _ := combat.ImageCache.GetImage(unit.Unit.CombatLbxFile, unit.Unit.GetCombatIndex(unit.Facing), 0)

        if combatImage != nil {
            options.GeoM.Reset()
            tx, ty := tilePosition(unit.X, unit.Y)
            options.GeoM.Translate(tx, ty)
            RenderCombatUnit(screen, combatImage, options, unit.Unit.Count)
        }
    }

    for _, unit := range combat.AttackingArmy.Units {
        combatImage, _ := combat.ImageCache.GetImage(unit.Unit.CombatLbxFile, unit.Unit.GetCombatIndex(unit.Facing), 0)

        if combatImage != nil {
            options.GeoM.Reset()
            tx, ty := tilePosition(unit.X, unit.Y)
            options.GeoM.Translate(tx, ty)
            RenderCombatUnit(screen, combatImage, options, unit.Unit.Count)
        }
    }

}
