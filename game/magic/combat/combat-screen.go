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
    "github.com/hajimehoshi/ebiten/v2/vector"
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
    SelectedUnit *ArmyUnit

    DebugFont *font.Font

    Coordinates ebiten.GeoM
    ScreenToTile ebiten.GeoM

    MouseTileX int
    MouseTileY int
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

    var selectedUnit *ArmyUnit
    if len(attackingArmy.Units) > 0 {
        selectedUnit = attackingArmy.Units[0]
    } else if len(defendingArmy.Units) > 0 {
        selectedUnit = defendingArmy.Units[0]
    }

    imageCache := util.MakeImageCache(cache)

    tile0, _ := imageCache.GetImage("cmbgrass.lbx", 0, 0)

    var coordinates ebiten.GeoM

    coordinates.Rotate(-math.Pi / 4)
    coordinates.Scale(float64(tile0.Bounds().Dx())/2, float64(tile0.Bounds().Dy())/2)
    coordinates.Translate(-220, 80)

    screenToTile := coordinates
    screenToTile.Translate(float64(tile0.Bounds().Dx())/2, float64(tile0.Bounds().Dy())/2)
    screenToTile.Invert()

    return &CombatScreen{
        Cache: cache,
        ImageCache: imageCache,
        DefendingArmy: defendingArmy,
        AttackingArmy: attackingArmy,
        Tiles: makeTiles(35, 35),
        SelectedUnit: selectedUnit,
        DebugFont: debugFont,
        Coordinates: coordinates,
        ScreenToTile: screenToTile,
    }
}

func (combat *CombatScreen) Update() CombatState {
    combat.Counter += 1

    mouseX, mouseY := ebiten.CursorPosition()

    tileX, tileY := combat.ScreenToTile.Apply(float64(mouseX), float64(mouseY))
    combat.MouseTileX = int(math.Round(tileX))
    combat.MouseTileY = int(math.Round(tileY))

    // log.Printf("Mouse original %v,%v %v,%v -> %v,%v", mouseX, mouseY, tileX, tileY, combat.MouseTileX, combat.MouseTileY)

    return CombatStateRunning
}

func (combat *CombatScreen) DrawHighlightedTile(screen *ebiten.Image, x int, y int, minColor color.RGBA, maxColor color.RGBA){
    tile0, _ := combat.ImageCache.GetImage("cmbgrass.lbx", 0, 0)

    tx, ty := combat.Coordinates.Apply(float64(x), float64(y))
    x1 := tx
    y1 := ty + float64(tile0.Bounds().Dy()/2)

    x2 := tx + float64(tile0.Bounds().Dx()/2)
    y2 := ty

    x3 := tx + float64(tile0.Bounds().Dx())
    y3 := ty + float64(tile0.Bounds().Dy()/2)

    x4 := tx + float64(tile0.Bounds().Dx()/2)
    y4 := ty + float64(tile0.Bounds().Dy())

    gradient := (math.Sin(float64(combat.Counter)/6) + 1)

    lerp := func(minC uint8, maxC uint8) uint8 {
        out := float64(minC) + gradient * float64(maxC - minC)/2
        if out > 255 {
            out = 255
        }
        if out < 0 {
            out = 0
        }

        return uint8(out)
    }

    lineColor := util.PremultiplyAlpha(color.RGBA{
        R: lerp(minColor.R, maxColor.R),
        G: lerp(minColor.G, maxColor.G),
        B: lerp(minColor.B, maxColor.B),
        A: 190})

    vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 1, lineColor, false)
    vector.StrokeLine(screen, float32(x2), float32(y2), float32(x3), float32(y3), 1, lineColor, false)
    vector.StrokeLine(screen, float32(x3), float32(y3), float32(x4), float32(y4), 1, lineColor, false)
    vector.StrokeLine(screen, float32(x4), float32(y4), float32(x1), float32(y1), 1, lineColor, false)
}

func (combat *CombatScreen) Draw(screen *ebiten.Image){

    animationIndex := combat.Counter / 8

    var options ebiten.DrawImageOptions

    tile0, _ := combat.ImageCache.GetImage("cmbgrass.lbx", 0, 0)

    /*
    a, b := screenToTile.Apply(160, 100)
    log.Printf("(160,100) -> (%f, %f)", a, b)
    */

    /*
    a, b := coordinates.Apply(3, 0)
    log.Printf("(3,3) -> (%f, %f)", a, b)
    */

    tilePosition := func(x int, y int) (float64, float64){
        return combat.Coordinates.Apply(float64(x), float64(y))
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

    combat.DrawHighlightedTile(screen, combat.MouseTileX, combat.MouseTileY, color.RGBA{R: 0, G: 0x67, B: 0x78, A: 255}, color.RGBA{R: 0, G: 0xef, B: 0xff, A: 255})

    if combat.SelectedUnit != nil {
        minColor := color.RGBA{R: 32, G: 0, B: 0, A: 255}
        maxColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
        combat.DrawHighlightedTile(screen, combat.SelectedUnit.X, combat.SelectedUnit.Y, minColor, maxColor)

        /*
        tx, ty := tilePosition(combat.SelectedUnit.X, combat.SelectedUnit.Y)
        x1 := tx
        y1 := ty + float64(tile0.Bounds().Dy()/2)

        x2 := tx + float64(tile0.Bounds().Dx()/2)
        y2 := ty

        x3 := tx + float64(tile0.Bounds().Dx())
        y3 := ty + float64(tile0.Bounds().Dy()/2)

        x4 := tx + float64(tile0.Bounds().Dx()/2)
        y4 := ty + float64(tile0.Bounds().Dy())

        minR := float64(32)
        r := minR + (math.Sin(float64(combat.Counter)/6) + 1) * (256-minR)/2

        if r > 255 {
            r = 255
        }

        if r < 0 {
            r = 0
        }

        lineColor := util.PremultiplyAlpha(color.RGBA{R: uint8(r), G: 0, B: 0, A: 190})

        vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 1, lineColor, false)
        vector.StrokeLine(screen, float32(x2), float32(y2), float32(x3), float32(y3), 1, lineColor, false)
        vector.StrokeLine(screen, float32(x3), float32(y3), float32(x4), float32(y4), 1, lineColor, false)
        vector.StrokeLine(screen, float32(x4), float32(y4), float32(x1), float32(y1), 1, lineColor, false)
        */
    }

    renderUnit := func(unit *ArmyUnit){
        combatImages, _ := combat.ImageCache.GetImages(unit.Unit.CombatLbxFile, unit.Unit.GetCombatIndex(unit.Facing))

        if combatImages != nil {
            options.GeoM.Reset()
            tx, ty := tilePosition(unit.X, unit.Y)
            options.GeoM.Translate(tx, ty)
            options.GeoM.Translate(float64(tile0.Bounds().Dx()/2), float64(tile0.Bounds().Dy()/2))

            index := uint64(0)
            if unit.Unit.Flying {
                index = animationIndex % (uint64(len(combatImages)) - 1)
            }
            RenderCombatUnit(screen, combatImages[index], options, unit.Unit.Count)
        }
    }

    for _, unit := range combat.DefendingArmy.Units {
        renderUnit(unit)
    }

    for _, unit := range combat.AttackingArmy.Units {
        renderUnit(unit)
    }

}
