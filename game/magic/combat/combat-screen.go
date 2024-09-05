package combat

import (
    "fmt"
    "log"
    "math"
    "math/rand"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/colorconv"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/player"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    // "github.com/hajimehoshi/ebiten/v2/vector"
)

type CombatState int

const (
    CombatStateRunning CombatState = iota
    CombatStateDone
)

type Team int

const (
    TeamAttacker Team = iota
    TeamDefender
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
    Moving bool
    X int
    Y int
    Health int

    Team Team

    Attacking bool
    AttackingCounter uint64

    Movement uint64
    MoveX float64
    MoveY float64

    TargetX int
    TargetY int

    LastTurn int
}

type Army struct {
    Units []*ArmyUnit
    Player *player.Player
}

type CombatScreen struct {
    Counter uint64
    Cache *lbx.LbxCache
    ImageCache util.ImageCache
    DefendingArmy *Army
    AttackingArmy *Army
    Tiles [][]Tile
    SelectedUnit *ArmyUnit

    Turn Team
    CurrentTurn int

    UI *uilib.UI

    DebugFont *font.Font
    HudFont *font.Font

    AttackingWizardFont *font.Font
    DefendingWizardFont *font.Font

    Coordinates ebiten.GeoM
    ScreenToTile ebiten.GeoM

    WhitePixel *ebiten.Image

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

func lighten(c color.RGBA, amount float64) color.Color {
    h, s, l := colorconv.ColorToHSL(c)
    l += amount/100
    if l > 1 {
        l = 1
    }
    out, err := colorconv.HSLToColor(h, s, l)
    if err != nil {
        log.Printf("Error in lighten: %v", err)
        return c
    }
    return out
}

func makePaletteFromBanner(banner data.BannerType) color.Palette {
    var topColor color.RGBA

    switch banner {
        case data.BannerGreen: topColor = color.RGBA{R: 0x20, G: 0x80, B: 0x2c, A: 0xff}
        case data.BannerBlue: topColor = color.RGBA{R: 0x15, G: 0x1d, B: 0x9d, A: 0xff}
        case data.BannerRed: topColor = color.RGBA{R: 0x9d, G: 0x15, B: 0x15, A: 0xff}
        case data.BannerPurple: topColor = color.RGBA{R: 0x6d, G: 0x15, B: 0x9d, A: 0xff}
        case data.BannerYellow: topColor = color.RGBA{R: 0x9d, G: 0x9d, B: 0x15, A: 0xff}
        case data.BannerBrown: topColor = color.RGBA{R: 0x82, G: 0x60, B: 0x12, A: 0xff}
    }

    // red := color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}

    return color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        lighten(topColor, 15), lighten(topColor, 10), lighten(topColor, 5),
        topColor, topColor, topColor,
        topColor, topColor, topColor,
    }
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

    black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
    blackPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        black, black, black,
        black, black, black,
    }

    hudFont := font.MakeOptimizedFontWithPalette(fonts[0], blackPalette)

    defendingWizardFont := font.MakeOptimizedFontWithPalette(fonts[4], makePaletteFromBanner(defendingArmy.Player.Wizard.Banner))
    attackingWizardFont := font.MakeOptimizedFontWithPalette(fonts[4], makePaletteFromBanner(attackingArmy.Player.Wizard.Banner))

    var selectedUnit *ArmyUnit
    if len(defendingArmy.Units) > 0 {
        selectedUnit = defendingArmy.Units[0]
    } else {
        log.Printf("Error: No defending units")
        return nil
    }

    imageCache := util.MakeImageCache(cache)

    tile0, _ := imageCache.GetImage("cmbgrass.lbx", 0, 0)

    var coordinates ebiten.GeoM

    // the battlefield is rotated by 45 degrees
    coordinates.Rotate(-math.Pi / 4)
    // coordinates.Scale(float64(tile0.Bounds().Dx())/2, float64(tile0.Bounds().Dy())/2)
    // FIXME: this math is hacky, but it works for now
    coordinates.Scale(float64(tile0.Bounds().Dx()) * 3 / 4 - 2, float64(tile0.Bounds().Dy()) * 3 / 4 - 1)
    coordinates.Translate(-220, 80)

    screenToTile := coordinates
    screenToTile.Translate(float64(tile0.Bounds().Dx())/2, float64(tile0.Bounds().Dy())/2)
    screenToTile.Invert()

    whitePixel := ebiten.NewImage(1, 1)
    whitePixel.Fill(color.RGBA{R: 255, G: 255, B: 255, A: 255})

    for _, unit := range defendingArmy.Units {
        unit.Team = TeamDefender
    }

    for _, unit := range attackingArmy.Units {
        unit.Team = TeamAttacker
    }

    // FIXME: do layout of armys

    combat := &CombatScreen{
        Cache: cache,
        ImageCache: imageCache,
        Turn: TeamDefender,
        CurrentTurn: 1,
        DefendingArmy: defendingArmy,
        AttackingArmy: attackingArmy,
        Tiles: makeTiles(30, 30),
        SelectedUnit: selectedUnit,
        DebugFont: debugFont,
        HudFont: hudFont,
        Coordinates: coordinates,
        ScreenToTile: screenToTile,
        WhitePixel: whitePixel,
        AttackingWizardFont: attackingWizardFont,
        DefendingWizardFont: defendingWizardFont,
    }

    combat.UI = combat.MakeUI()
    return combat
}

func (combat *CombatScreen) MakeUI() *uilib.UI {
    var elements []*uilib.UIElement

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            hudImage, _ := combat.ImageCache.GetImage("cmbtfx.lbx", 28, 0)
            options.GeoM.Reset()
            options.GeoM.Translate(0, float64(data.ScreenHeight - hudImage.Bounds().Dy()))
            for i := 0; i < 4; i++ {
                screen.DrawImage(hudImage, &options)
                options.GeoM.Translate(float64(hudImage.Bounds().Dx()), 0)
            }

            combat.AttackingWizardFont.Print(screen, 265, 170, 1, ebiten.ColorScale{}, combat.AttackingArmy.Player.Wizard.Name)
            combat.DefendingWizardFont.Print(screen, 30, 170, 1, ebiten.ColorScale{}, combat.DefendingArmy.Player.Wizard.Name)

            rightImage, _ := combat.ImageCache.GetImage(combat.SelectedUnit.Unit.CombatLbxFile, combat.SelectedUnit.Unit.GetCombatIndex(units.FacingRight), 0)
            options.GeoM.Reset()
            options.GeoM.Translate(90, 170)
            screen.DrawImage(rightImage, &options)

            combat.HudFont.Print(screen, 92, 167, 1, ebiten.ColorScale{}, combat.SelectedUnit.Unit.Name)

            plainAttack, _ := combat.ImageCache.GetImage("compix.lbx", 29, 0)
            options.GeoM.Reset()
            options.GeoM.Translate(126, 173)
            screen.DrawImage(plainAttack, &options)
            combat.HudFont.Print(screen, 121, 174, 1, ebiten.ColorScale{}, fmt.Sprintf("%v", combat.SelectedUnit.Unit.MeleeAttackPower))

            var movementImage *ebiten.Image
            if combat.SelectedUnit.Unit.Flying {
                movementImage, _ = combat.ImageCache.GetImage("compix.lbx", 39, 0)
            } else {
                movementImage, _ = combat.ImageCache.GetImage("compix.lbx", 38, 0)
            }

            options.GeoM.Reset()
            options.GeoM.Translate(126, 188)
            screen.DrawImage(movementImage, &options)
            combat.HudFont.Print(screen, 121, 190, 1, ebiten.ColorScale{}, fmt.Sprintf("%v", combat.SelectedUnit.Unit.MovementSpeed))

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })
        },
    }

    buttonX := float64(139)
    buttonY := float64(167)

    elements = append(elements, &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            spellButtons, _ := combat.ImageCache.GetImages("compix.lbx", 1)
            options.GeoM.Reset()
            options.GeoM.Translate(buttonX, buttonY)
            screen.DrawImage(spellButtons[0], &options)
        },
    })


    elements = append(elements, &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            waitButtons, _ := combat.ImageCache.GetImages("compix.lbx", 2)
            options.GeoM.Translate(buttonX, buttonY)
            options.GeoM.Translate(float64(waitButtons[0].Bounds().Dx()), 0)
            screen.DrawImage(waitButtons[0], &options)
        },
    })

    elements = append(elements, &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions

            infoButtons, _ := combat.ImageCache.GetImages("compix.lbx", 20)
            options.GeoM.Translate(buttonX, buttonY)
            options.GeoM.Translate(0, float64(infoButtons[0].Bounds().Dy()))
            screen.DrawImage(infoButtons[0], &options)
        },
    })

    elements = append(elements, &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            autoButtons, _ := combat.ImageCache.GetImages("compix.lbx", 4)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(buttonX, buttonY)
            options.GeoM.Translate(float64(autoButtons[0].Bounds().Dx()), float64(autoButtons[0].Bounds().Dy()))
            screen.DrawImage(autoButtons[0], &options)
        },
    })

    elements = append(elements, &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            fleeButtons, _ := combat.ImageCache.GetImages("compix.lbx", 21)
            options.GeoM.Translate(buttonX, buttonY)
            options.GeoM.Translate(0, float64(fleeButtons[0].Bounds().Dy()) * 2)
            screen.DrawImage(fleeButtons[0], &options)
        },
    })

    elements = append(elements, &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            doneButtons, _ := combat.ImageCache.GetImages("compix.lbx", 3)
            options.GeoM.Translate(buttonX, buttonY)
            options.GeoM.Translate(float64(doneButtons[0].Bounds().Dx()), float64(doneButtons[0].Bounds().Dy()) * 2)
            screen.DrawImage(doneButtons[0], &options)
        },
    })

    ui.SetElementsFromArray(elements)

    return ui
}

/* check that 'check' is between angle-spread and angle+spread
 */
func betweenAngle(check float64, angle float64, spread float64) bool {
    minAngle := angle - spread
    maxAngle := angle + spread

    for minAngle < 0 {
        minAngle += math.Pi * 2
        maxAngle += math.Pi * 2
    }

    // at least make 'check' positive
    for check < 0 {
        check += math.Pi * 2
    }

    // if minAngle is above pi, then try to move 'check' to be in the same range
    if check < minAngle {
        check += math.Pi * 2
    // same for max angle
    } else if check > maxAngle {
        check -= math.Pi * 2
    }

    // now check if 'check' is between min and max
    return check >= minAngle && check <= maxAngle
}

func (combat *CombatScreen) TileIsEmpty(x int, y int) bool {
    for _, unit := range combat.DefendingArmy.Units {
        if unit.Health > 0 && unit.X == x && unit.Y == y {
            return false
        }
    }

    for _, unit := range combat.AttackingArmy.Units {
        if unit.Health > 0 && unit.X == x && unit.Y == y {
            return false
        }
    }

    return true
}

// angle in radians
func computeFacing(angle float64) units.Facing {
    // right
    if betweenAngle(angle, 0, math.Pi/8){
        return units.FacingRight
    }

    // left
    if betweenAngle(angle, math.Pi, math.Pi/8){
        return units.FacingLeft
    }

    // up
    if betweenAngle(angle, math.Pi/2, math.Pi/8){
        return units.FacingUp
    }

    // up-left
    if betweenAngle(angle, math.Pi - math.Pi/4, math.Pi/8){
        return units.FacingUpLeft
    }

    // up-right
    if betweenAngle(angle, math.Pi/4, math.Pi/8){
        return units.FacingUpRight
    }

    // down-left
    if betweenAngle(angle, math.Pi + math.Pi/4, math.Pi/8){
        return units.FacingDownLeft
    }

    // down
    if betweenAngle(angle, math.Pi + math.Pi/2, math.Pi/8){
        return units.FacingDown
    }

    // down-right
    if betweenAngle(angle, math.Pi + math.Pi/2 + math.Pi/4, math.Pi/8){
        return units.FacingDownRight
    }

    // should be impossible to get here
    return units.FacingRight
}

func (combat *CombatScreen) NextUnit() {
    /*
    if combat.Turn == TurnDefending {
        combat.Turn = TurnAttacking
    } else {
        combat.Turn = TurnDefending
    }
    */

    combat.SelectedUnit = nil

    canMove := false
    for _, unit := range combat.DefendingArmy.Units {
        if unit.LastTurn < combat.CurrentTurn {
            canMove = true
            break
        }
    }

    if !canMove {
        for _, unit := range combat.AttackingArmy.Units {
            if unit.LastTurn < combat.CurrentTurn {
                canMove = true
                break
            }
        }
    }

    // no one left can move in this turn, go to next turn
    if !canMove {
        combat.CurrentTurn += 1

        if combat.Turn == TeamDefender {
            combat.Turn = TeamAttacker
        } else {
            combat.Turn = TeamDefender
        }
    }

    for combat.SelectedUnit == nil {
        switch combat.Turn {
        case TeamDefender:
            found := false
            for _, unit := range combat.DefendingArmy.Units {
                if unit.LastTurn < combat.CurrentTurn {
                    combat.SelectedUnit = unit
                    found = true
                    break
                }
            }
            if !found {
                combat.Turn = TeamAttacker
            }
        case TeamAttacker:
            found := false
            for _, unit := range combat.AttackingArmy.Units {
                if unit.LastTurn < combat.CurrentTurn {
                    combat.SelectedUnit = unit
                    found = true
                    break
                }
            }
            if !found {
                combat.Turn = TeamDefender
            }
        }
    }
}

func (combat *CombatScreen) GetUnit(x int, y int) *ArmyUnit {
    for _, unit := range combat.DefendingArmy.Units {
        if unit.Health > 0 && unit.X == x && unit.Y == y {
            return unit
        }
    }

    for _, unit := range combat.AttackingArmy.Units {
        if unit.Health > 0 && unit.X == x && unit.Y == y {
            return unit
        }
    }

    return nil
}

func (combat *CombatScreen) ContainsOppositeArmy(x int, y int, team Team) bool {
    unit := combat.GetUnit(x, y)
    if unit == nil {
        return false
    }
    return unit.Team != team
}

func faceTowards(x1 int, y1 int, x2 int, y2 int) units.Facing {
    angle := math.Atan2(float64(y2 - y1), float64(x2 - x1))

    // rotate by 45 degrees to get the on screen facing angle
    // have to negate the angle because the y axis is flipped (higher y values are lower on the screen)
    useAngle := -(angle - math.Pi/4)

    // log.Printf("Angle: %v from (%v,%v) to (%v,%v)", useAngle, combat.SelectedUnit.X, combat.SelectedUnit.Y, combat.SelectedUnit.TargetX, combat.SelectedUnit.TargetY)

    return computeFacing(useAngle)
}

func (combat *CombatScreen) withinMeleeRange(attacker *ArmyUnit, defender *ArmyUnit) bool {
    xDiff := math.Abs(float64(attacker.X - defender.X))
    yDiff := math.Abs(float64(attacker.Y - defender.Y))

    return xDiff <= 1 && yDiff <= 1
}

func (combat *CombatScreen) canAttack(attacker *ArmyUnit, defender *ArmyUnit) bool {
    if defender.Unit.Flying && !attacker.Unit.Flying {
        return false
    }

    return true
}

func (combat *CombatScreen) Update() CombatState {
    combat.Counter += 1

    mouseX, mouseY := ebiten.CursorPosition()

    tileX, tileY := combat.ScreenToTile.Apply(float64(mouseX), float64(mouseY))
    combat.MouseTileX = int(math.Round(tileX))
    combat.MouseTileY = int(math.Round(tileY))

    hudImage, _ := combat.ImageCache.GetImage("cmbtfx.lbx", 28, 0)

    // dont allow clicks into the hud area
    if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) &&
       mouseY < data.ScreenHeight - hudImage.Bounds().Dy() &&
       combat.SelectedUnit.Moving == false && combat.SelectedUnit.Attacking == false {

        if combat.TileIsEmpty(combat.MouseTileX, combat.MouseTileY) {
            combat.SelectedUnit.Movement = combat.Counter
            combat.SelectedUnit.TargetX = combat.MouseTileX
            combat.SelectedUnit.TargetY = combat.MouseTileY
            combat.SelectedUnit.Moving = true
       } else {

           defender := combat.GetUnit(combat.MouseTileX, combat.MouseTileY)

           if defender != nil && defender.Team != combat.SelectedUnit.Team && combat.withinMeleeRange(combat.SelectedUnit, defender) && combat.canAttack(combat.SelectedUnit, defender){
               combat.SelectedUnit.Attacking = true
               combat.SelectedUnit.AttackingCounter = combat.Counter

               combat.SelectedUnit.Facing = faceTowards(combat.SelectedUnit.X, combat.SelectedUnit.Y, combat.MouseTileX, combat.MouseTileY)
               defender.Facing = faceTowards(defender.X, defender.Y, combat.SelectedUnit.X, combat.SelectedUnit.Y)

               // FIXME: sound is based on attacker type, and possibly defender type
               sound, err := audio.LoadCombatSound(combat.Cache, 1)
               if err == nil {
                   sound.Play()
               }
           }
       }
    }

    if combat.SelectedUnit.Attacking {
        if combat.Counter - combat.SelectedUnit.AttackingCounter > 60 {
            combat.SelectedUnit.Attacking = false
            combat.SelectedUnit.AttackingCounter = 0
        }
    }

    if combat.SelectedUnit.Moving {
        angle := math.Atan2(float64(combat.SelectedUnit.TargetY - combat.SelectedUnit.Y), float64(combat.SelectedUnit.TargetX - combat.SelectedUnit.X))

        // rotate by 45 degrees to get the on screen facing angle
        // have to negate the angle because the y axis is flipped (higher y values are lower on the screen)
        useAngle := -(angle - math.Pi/4)

        // log.Printf("Angle: %v from (%v,%v) to (%v,%v)", useAngle, combat.SelectedUnit.X, combat.SelectedUnit.Y, combat.SelectedUnit.TargetX, combat.SelectedUnit.TargetY)

        combat.SelectedUnit.Facing = computeFacing(useAngle)

        speed := float64(combat.Counter - combat.SelectedUnit.Movement) / 4
        newX := float64(combat.SelectedUnit.X) + math.Cos(angle) * speed
        newY := float64(combat.SelectedUnit.Y) + math.Sin(angle) * speed

        combat.SelectedUnit.MoveX = newX
        combat.SelectedUnit.MoveY = newY

        if math.Abs(newX - float64(combat.SelectedUnit.TargetX)) < 0.5 && math.Abs(newY - float64(combat.SelectedUnit.TargetY)) < 0.5 {
            combat.SelectedUnit.LastTurn = combat.CurrentTurn
            combat.SelectedUnit.Moving = false
            combat.SelectedUnit.X = combat.SelectedUnit.TargetX
            combat.SelectedUnit.Y = combat.SelectedUnit.TargetY

            combat.NextUnit()
        }
    }

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


    rFloat := float32(lineColor.R) / 255
    gFloat := float32(lineColor.G) / 255
    bFloat := float32(lineColor.B) / 255
    aFloat := float32(lineColor.A) / 255

    vertices := []ebiten.Vertex{
        ebiten.Vertex{DstX: float32(x1), DstY: float32(y1), SrcX: 0, SrcY: 0, ColorR: rFloat, ColorG: gFloat, ColorB: bFloat, ColorA: aFloat},
        ebiten.Vertex{DstX: float32(x2), DstY: float32(y2), SrcX: 0, SrcY: 0, ColorR: rFloat, ColorG: gFloat, ColorB: bFloat, ColorA: aFloat},
        ebiten.Vertex{DstX: float32(x3), DstY: float32(y3), SrcX: 0, SrcY: 0, ColorR: rFloat, ColorG: gFloat, ColorB: bFloat, ColorA: aFloat},
        ebiten.Vertex{DstX: float32(x4), DstY: float32(y4), SrcX: 0, SrcY: 0, ColorR: rFloat, ColorG: gFloat, ColorB: bFloat, ColorA: aFloat},
    }

    indicies := []uint16{0, 1, 2, 2, 3, 0}

    screen.DrawTriangles(vertices, indicies, combat.WhitePixel, &ebiten.DrawTrianglesOptions{})


        /*
    vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 1, lineColor, false)
    vector.StrokeLine(screen, float32(x2), float32(y2), float32(x3), float32(y3), 1, lineColor, false)
    vector.StrokeLine(screen, float32(x3), float32(y3), float32(x4), float32(y4), 1, lineColor, false)
    vector.StrokeLine(screen, float32(x4), float32(y4), float32(x1), float32(y1), 1, lineColor, false)
    */
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

        if !combat.SelectedUnit.Moving {
            combat.DrawHighlightedTile(screen, combat.SelectedUnit.X, combat.SelectedUnit.Y, minColor, maxColor)
        }
    }

    renderUnit := func(unit *ArmyUnit){
        combatImages, _ := combat.ImageCache.GetImages(unit.Unit.CombatLbxFile, unit.Unit.GetCombatIndex(unit.Facing))

        if combatImages != nil {
            options.GeoM.Reset()
            var tx float64
            var ty float64

            if unit.Moving {
                tx, ty = combat.Coordinates.Apply(unit.MoveX, unit.MoveY)
            } else {
                tx, ty = tilePosition(unit.X, unit.Y)
            }
            options.GeoM.Translate(tx, ty)
            options.GeoM.Translate(float64(tile0.Bounds().Dx()/2), float64(tile0.Bounds().Dy()/2))

            index := uint64(0)
            if unit.Unit.Flying {
                index = animationIndex % (uint64(len(combatImages)) - 1)
            }

            if unit.Attacking {
                index = 2 + animationIndex % 2
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

    combat.UI.Draw(combat.UI, screen)
}
