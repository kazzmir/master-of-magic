package combat

import (
    // "fmt"
    "log"
    "math"
    "math/rand"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
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
    Counter uint64
    Cache *lbx.LbxCache
    ImageCache util.ImageCache
    DefendingArmy *Army
    AttackingArmy *Army
    Tiles [][]Tile

    DebugFont *font.Font
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
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
    }

    white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
    whitePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        white, white, white,
        white, white, white,
    }

    debugFont := font.MakeOptimizedFontWithPalette(fonts[0], whitePalette)

    return &CombatScreen{
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
        DefendingArmy: defendingArmy,
        AttackingArmy: attackingArmy,
        Tiles: makeTiles(35, 35),
        DebugFont: debugFont,
    }
}

func (combat *CombatScreen) Update() CombatState {
    combat.Counter += 1
    return CombatStateRunning
}

func (combat *CombatScreen) Draw(screen *ebiten.Image){

    animationIndex := combat.Counter / 8

    var options ebiten.DrawImageOptions

    tile0, _ := combat.ImageCache.GetImage("cmbgrass.lbx", 0, 0)

    var coordinates ebiten.GeoM

    coordinates.Rotate(-math.Pi / 4)
    coordinates.Scale(float64(tile0.Bounds().Dx()/2), float64(tile0.Bounds().Dy()/2))
    coordinates.Translate(-220, 80)

    screenToTile := coordinates
    screenToTile.Invert()

    /*
    a, b := screenToTile.Apply(160, 100)
    log.Printf("(160,100) -> (%f, %f)", a, b)
    */

    /*
    a, b := coordinates.Apply(3, 0)
    log.Printf("(3,3) -> (%f, %f)", a, b)
    */

    tilePosition := func(x int, y int) (float64, float64){
        return coordinates.Apply(float64(x), float64(y))
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

            // combat.DebugFont.Print(screen, tx, ty, 1, ebiten.ColorScale{}, fmt.Sprintf("%v,%v", x, y))
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
        combatImages, _ := combat.ImageCache.GetImages(unit.Unit.CombatLbxFile, unit.Unit.GetCombatIndex(unit.Facing))

        if combatImages != nil {
            options.GeoM.Reset()
            tx, ty := tilePosition(unit.X, unit.Y)
            options.GeoM.Translate(tx, ty)

            index := uint64(0)
            if unit.Unit.Flying {
                index = animationIndex % (uint64(len(combatImages)) - 1)
            }

            RenderCombatUnit(screen, combatImages[index], options, unit.Unit.Count)
        }
    }

    for _, unit := range combat.AttackingArmy.Units {
        combatImages, _ := combat.ImageCache.GetImages(unit.Unit.CombatLbxFile, unit.Unit.GetCombatIndex(unit.Facing))

        if combatImages != nil {
            options.GeoM.Reset()
            tx, ty := tilePosition(unit.X, unit.Y)
            options.GeoM.Translate(tx, ty)

            index := uint64(0)
            if unit.Unit.Flying {
                index = animationIndex % (uint64(len(combatImages)) - 1)
            }

            RenderCombatUnit(screen, combatImages[index], options, unit.Unit.Count)
        }
    }

}
