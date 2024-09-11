package combat

import (
    "fmt"
    "log"
    "math"
    "math/rand"
    "image"
    "image/color"
    "slices"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/mouse"
    "github.com/kazzmir/master-of-magic/lib/colorconv"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/vector"
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
    TeamEither
)

func (team Team) String() string {
    switch team {
        case TeamAttacker: return "Attacker"
        case TeamDefender: return "Defender"
        case TeamEither: return "Either"
    }
    return "Unknown"
}

func oppositeTeam(a Team) Team {
    if a == TeamAttacker {
        return TeamDefender
    }
    return TeamAttacker
}

type MouseState int
const (
    CombatClickHud MouseState = iota
    CombatMoveOk
    CombatMeleeAttackOk
    CombatRangeAttackOk
    CombatNotOk
    CombatCast
)

type Tile struct {
    // index of grass/floor
    Index int
    // tree/rock on top, or -1 if nothing
    ExtraObject int
    Mud bool
}

type ArmyUnit struct {
    Unit units.Unit
    Facing units.Facing
    Moving bool
    X int
    Y int
    Health int
    MovesLeft fraction.Fraction

    Team Team

    Attacking bool
    AttackingCounter uint64

    MovementTick uint64
    MoveX float64
    MoveY float64

    // the path the unit is currently following
    MovementPath pathfinding.Path

    LastTurn int

    // ugly to need this, but this caches paths computed for the unit
    Paths map[image.Point]pathfinding.Path
}

// cost to move one tile in one of the 8 directions
func pathCost(from image.Point, to image.Point) fraction.Fraction {
    xDiff := int(math.Abs(float64(from.X - to.X)))
    yDiff := int(math.Abs(float64(from.Y - to.Y)))

    if xDiff == 0 && yDiff == 1 {
        return fraction.FromInt(1)
    }

    if xDiff == 1 && yDiff == 0 {
        return fraction.FromInt(1)
    }

    if xDiff == 1 && yDiff == 1 {
        return fraction.Make(3, 2)
    }

    if xDiff == 0 && yDiff == 0 {
        return fraction.FromInt(0)
    }

    // shouldn't ever really get here
    return fraction.Make(xDiff + yDiff, 1)
}

func (unit *ArmyUnit) CanFollowPath(path pathfinding.Path) bool {
    movesLeft := unit.MovesLeft

    /*
    var start image.Point
    var end image.Point
    if len(path) > 0 {
        start = path[0]
        end = path[len(path) - 1]
    }

    log.Printf("Can move from %v,%v to %v,%v path %v", start.X, start.Y, end.X, end.Y, path)
    */

    for i := 1; i < len(path); i++ {
        if movesLeft.GreaterThan(fraction.FromInt(0)) {
            movesLeft = movesLeft.Subtract(pathCost(path[i-1], path[i]))
        } else {
            return false
        }
    }

    return true
}

func computeMoves(x1 int, y1 int, x2 int, y2 int) fraction.Fraction {
    movesNeeded := fraction.Fraction{}

    for x1 != x2 || y1 != y2 {
        // movesNeeded += 1

        xDiff := int(math.Abs(float64(x1 - x2)))
        yDiff := int(math.Abs(float64(y1 - y2)))

        // move diagonally
        if xDiff > 0 && yDiff > 0 {
            movesNeeded = movesNeeded.Add(fraction.Make(3, 2))
        } else {
            movesNeeded = movesNeeded.Add(fraction.FromInt(1))
        }

        // a move can be made in any of the 8 available directions
        if x1 < x2 {
            x1 += 1
        }
        if x1 > x2 {
            x1 -= 1
        }
        if y1 < y2 {
            y1 += 1
        }
        if y1 > y2 {
            y1 -= 1
        }
    }

    return movesNeeded
}

func (combat *CombatScreen) computePath(x1 int, y1 int, x2 int, y2 int) (pathfinding.Path, bool) {

    containsUnit := make(map[image.Point]bool)

    tileEmpty := func (x int, y int) bool {
        // check if the tile is empty
        where := image.Pt(x, y)
        contains, ok := containsUnit[where]
        if ok {
            return !contains
        } else {
            unit := combat.GetUnit(x, y)
            if unit == nil {
                containsUnit[where] = false
                return true
            } else {
                containsUnit[where] = true
                return false
            }
        }
    }

    // FIXME: take into account mud, hills, other types of terrain obstacles
    tileCost := func (x1 int, y1 int, x2 int, y2 int) float64 {

        if x2 < 0 || y2 < 0 || y2 >= len(combat.Tiles) || x2 >= len(combat.Tiles[y2]) {
            return pathfinding.Infinity
        }

        if !tileEmpty(x2, y2) {
            return pathfinding.Infinity
        }

        xDiff := int(math.Abs(float64(x1 - x2)))
        yDiff := int(math.Abs(float64(y1 - y2)))

        if xDiff == 0 && yDiff == 1 {
            return 1
        }

        if xDiff == 1 && yDiff == 0 {
            return 1
        }

        if xDiff == 1 && yDiff == 1 {
            return 1.5
        }

        if xDiff == 0 && yDiff == 0 {
            return 0
        }

        // shouldn't ever really get here
        return float64(xDiff + yDiff)
    }

    neighbors := func(cx int, cy int) []image.Point {
        // var out []image.Point
        out := make([]image.Point, 0, 8)
        for dx := -1; dx <= 1; dx++ {
            for dy := -1; dy <= 1; dy++ {
                if dx == 0 && dy == 0 {
                    continue
                }

                x := cx + dx
                y := cy + dy

                if x >= 0 && y >= 0 && y < len(combat.Tiles) && x < len(combat.Tiles[y]) {
                    // ignore non-empty tiles entirely
                    if tileEmpty(x, y) {
                        out = append(out, image.Pt(x, y))
                    }
                }
            }
        }
        return out
    }

    return pathfinding.FindPath(image.Pt(x1, y1), image.Pt(x2, y2), 50, tileCost, neighbors)
}

/* return a valid path that the given unit can take to reach tile position x, y
 * this caches the path such that the next call to FindPath() will return the same path without computing it
 */
func (combat *CombatScreen) FindPath(unit *ArmyUnit, x int, y int) (pathfinding.Path, bool) {
    end := image.Pt(x, y)
    path, ok := unit.Paths[end]
    if ok {
        return path, len(path) > 0
    }

    path, ok = combat.computePath(unit.X, unit.Y, x, y)
    if !ok {
        unit.Paths[end] = nil
        // log.Printf("No such path from %v,%v -> %v,%v", unit.X, unit.Y, x, y)
        return nil, false
    }

    canMove := unit.CanFollowPath(path)

    if canMove {
        unit.Paths[end] = path
    } else {
        unit.Paths[end] = nil
    }

    return path, canMove
}

func (combat *CombatScreen) CanMoveTo(unit *ArmyUnit, x int, y int) bool {
    _, ok := combat.FindPath(unit, x, y)
    return ok
}

// this allows a unit to move a space diagonally even if they only have 0.5 movement points left
func (unit *ArmyUnit) CanMoveTo_legacy(x int, y int) bool {
    /*
    movesNeeded := computeMoves(unit.X, unit.Y, x, y)
    // log.Printf("CanMoveTo: %v,%v -> %v,%v moves left %v need %v", unit.X, unit.Y, x, y, unit.MovesLeft, movesNeeded)

    return movesNeeded.LessThanEqual(unit.MovesLeft)
    */

    moves := unit.MovesLeft

    // movesA := moves

    x1 := unit.X
    y1 := unit.Y

    for (x1 != x || y1 != y) && moves.GreaterThan(fraction.FromInt(0)) {
        xDiff := int(math.Abs(float64(x1 - x)))
        yDiff := int(math.Abs(float64(y1 - y)))

        // move diagonally
        if xDiff > 0 && yDiff > 0 {
            moves = moves.Subtract(fraction.Make(3, 2))
        } else {
            moves = moves.Subtract(fraction.FromInt(1))
        }

        // a move can be made in any of the 8 available directions
        if x1 < x {
            x1 += 1
        }
        if x1 > x {
            x1 -= 1
        }
        if y1 < y {
            y1 += 1
        }
        if y1 > y {
            y1 -= 1
        }
    }

    // movesB := moves

    // log.Printf("CanMoveTo: %v,%v -> %v,%v moves start %v left %v", unit.X, unit.Y, x, y, movesA, movesB)

    return x1 == x && y1 == y

    // return movesRemaining(unit.X, unit.Y, x, y, unit.MovesLeft).GreaterThanEqual(fraction.FromInt(0))

    /*
    xDiff := math.Abs(float64(unit.X - x))
    yDiff := math.Abs(float64(unit.Y - y))
    return int(xDiff + yDiff) <= unit.MovesLeft
    */
}

type Army struct {
    Units []*ArmyUnit
    Player *player.Player
}

// represents a unit that is not part of the army, for things like magic vortex, for things like magic vortex
type CombatUnit struct {
    X int
    Y int
    Animation *util.Animation
}

type Projectile struct {
    X float64
    Y float64
    Speed float64
    Angle float64
    TargetX float64
    TargetY float64
    Exploding bool
    Animation *util.Animation
    Explode *util.Animation
}

type CombatScreen struct {
    Counter uint64
    Cache *lbx.LbxCache
    ImageCache util.ImageCache
    DefendingArmy *Army
    AttackingArmy *Army
    Tiles [][]Tile
    // order to draw tiles in such that they are drawn from the top of the screen to the bottom (painter's order)
    TopDownOrder []image.Point
    SelectedUnit *ArmyUnit

    TurnAttacker int
    TurnDefender int

    OtherUnits []*CombatUnit

    MouseState MouseState

    Mouse *mouse.MouseData

    Turn Team
    CurrentTurn int

    UI *uilib.UI

    Projectiles []*Projectile

    DebugFont *font.Font
    HudFont *font.Font
    InfoFont *font.Font
    WhiteFont *font.Font

    // when the user hovers over a unit, that unit should be shown in a little info box at the upper right
    HighlightedUnit *ArmyUnit

    AttackingWizardFont *font.Font
    DefendingWizardFont *font.Font

    Coordinates ebiten.GeoM
    ScreenToTile ebiten.GeoM

    WhitePixel *ebiten.Image

    MouseTileX int
    MouseTileY int

    // if true then the player should select a tile to cast a spell on
    DoSelectTile bool
    SelectTile func(int, int)

    // if true then the player should select a unit to cast a spell on
    DoSelectUnit bool
    // which team to pick a unit from
    SelectTeam Team
    // invoke this function on the unit that is selected
    SelectTarget func(*ArmyUnit)
    CanTarget func(*ArmyUnit) bool
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

func MakeCombatScreen(cache *lbx.LbxCache, defendingArmy *Army, attackingArmy *Army, player *playerlib.Player) *CombatScreen {
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

    orange := color.RGBA{R: 0xf6, G: 0x9c, B: 0x22, A: 0xff}
    orangePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        orange, orange, orange,
        orange, orange, orange,
    }

    infoFont := font.MakeOptimizedFontWithPalette(fonts[0], orangePalette)

    whiteFont := font.MakeOptimizedFontWithPalette(fonts[0], whitePalette)

    defendingWizardFont := font.MakeOptimizedFontWithPalette(fonts[4], makePaletteFromBanner(defendingArmy.Player.Wizard.Banner))
    attackingWizardFont := font.MakeOptimizedFontWithPalette(fonts[4], makePaletteFromBanner(attackingArmy.Player.Wizard.Banner))

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

    mouseData, err := mouse.MakeMouseData(cache)
    if err != nil {
        log.Printf("Error loading mouse data: %v", err)
        return nil
    }

    // FIXME: do layout of armys

    combat := &CombatScreen{
        Cache: cache,
        ImageCache: imageCache,
        Mouse: mouseData,
        Turn: TeamDefender,
        CurrentTurn: 0,
        DefendingArmy: defendingArmy,
        TurnDefender: 0,
        AttackingArmy: attackingArmy,
        TurnAttacker: 0,
        Tiles: makeTiles(30, 30),
        SelectedUnit: nil,
        DebugFont: debugFont,
        HudFont: hudFont,
        InfoFont: infoFont,
        WhiteFont: whiteFont,
        Coordinates: coordinates,
        ScreenToTile: screenToTile,
        WhitePixel: whitePixel,
        AttackingWizardFont: attackingWizardFont,
        DefendingWizardFont: defendingWizardFont,
    }

    combat.TopDownOrder = combat.computeTopDownOrder()

    /*
    log.Printf("Top down order: %v", combat.TopDownOrder)

    for i := 0; i < 5; i++ {
        x1, y1 := combat.Coordinates.Apply(float64(combat.TopDownOrder[i].X), float64(combat.TopDownOrder[i].Y))
        log.Printf("tile %v (%v,%v): %v,%v", i, combat.TopDownOrder[i].X, combat.TopDownOrder[i].Y, x1, y1)
    }
    */

    combat.UI = combat.MakeUI(player)
    combat.NextTurn()
    combat.SelectedUnit = combat.ChooseNextUnit(TeamDefender)

    return combat
}

func (combat *CombatScreen) computeTopDownOrder() []image.Point {
    var points []image.Point
    for y := 0; y < len(combat.Tiles); y++ {
        for x := 0; x < len(combat.Tiles[y]); x++ {
            points = append(points, image.Pt(x, y))
        }
    }

    compare := func(a image.Point, b image.Point) int {
        ax, ay := combat.Coordinates.Apply(float64(a.X), float64(a.Y))
        bx, by := combat.Coordinates.Apply(float64(b.X), float64(b.Y))

        if ay < by {
            return -1
        }

        if ay > by {
            return 1
        }

        if ax < bx {
            return -1
        }

        if ax > bx {
            return 1
        }

        return 0
    }

    slices.SortFunc(points, compare)
    return points
}

func (combat *CombatScreen) AddProjectile(projectile *Projectile){
    combat.Projectiles = append(combat.Projectiles, projectile)
}

/* a projectile that shoots down from the sky at an angle
 */
func (combat *CombatScreen) createSkyProjectile(target *ArmyUnit, images []*ebiten.Image, explodeImages []*ebiten.Image) *Projectile {
    // find where on the screen the unit is
    screenX, screenY := combat.Coordinates.Apply(float64(target.X), float64(target.Y))
    screenY -= 10
    screenX += 2

    x := screenX + 80 + rand.Float64() * 60
    y := -rand.Float64() * 40

    // FIXME: make this a parameter?
    speed := 2.5

    angle := math.Atan2(screenY - y, screenX - x)

    // log.Printf("Create fireball projectile at %v,%v -> %v,%v", x, y, screenX, screenY)

    projectile := &Projectile{
        X: x,
        Y: y,
        Speed: speed,
        Angle: angle,
        TargetX: screenX,
        TargetY: screenY,
        Animation: util.MakeAnimation(images, true),
        Explode: util.MakeAnimation(explodeImages, false),
    }

    return projectile
}

/* a projectile that shoots down from the sky vertically
 */
func (combat *CombatScreen) createVerticalSkyProjectile(target *ArmyUnit, images []*ebiten.Image, explodeImages []*ebiten.Image) *Projectile {
    // find where on the screen the unit is
    screenX, screenY := combat.Coordinates.Apply(float64(target.X), float64(target.Y))
    screenY -= 10
    screenX += 2

    x := screenX
    y := -40.0

    // FIXME: make this a parameter?
    speed := 2.5

    angle := math.Pi / 2

    // log.Printf("Create fireball projectile at %v,%v -> %v,%v", x, y, screenX, screenY)

    projectile := &Projectile{
        X: x,
        Y: y,
        Speed: speed,
        Angle: angle,
        TargetX: screenX,
        TargetY: screenY,
        Animation: util.MakeAnimation(images, true),
        Explode: util.MakeAnimation(explodeImages, false),
    }

    return projectile
}

type UnitPosition int
const (
    UnitPositionMiddle UnitPosition = iota
    UnitPositionUnder
)

type Targeting int
const (
    TargetFriend Targeting = iota
    TargetEnemy
    TargetEither
)

/* needs a new name, but creates a projectile that is already at the target
 */
func (combat *CombatScreen) createUnitProjectile(target *ArmyUnit, images []*ebiten.Image, explodeImages []*ebiten.Image, position UnitPosition) *Projectile {
    // find where on the screen the unit is
    screenX, screenY := combat.Coordinates.Apply(float64(target.X), float64(target.Y))

    var useImage *ebiten.Image
    if len(images) > 0 {
        useImage = images[0]
    } else if len(explodeImages) > 0 {
        useImage = explodeImages[0]
    }

    switch position {
        case UnitPositionMiddle:
            screenY += 3
            screenY -= float64(useImage.Bounds().Dy()/2)
            screenX += 14
            screenX -= float64(useImage.Bounds().Dx()/2)
        case UnitPositionUnder:
            screenY += 15
            screenY -= float64(useImage.Bounds().Dy())
    }


    // log.Printf("Create fireball projectile at %v,%v -> %v,%v", x, y, screenX, screenY)

    projectile := &Projectile{
        X: screenX,
        Y: screenY,
        Speed: 0,
        Angle: 0,
        TargetX: screenX,
        TargetY: screenY,
        Animation: util.MakeAnimation(images, true),
        Explode: util.MakeAnimation(explodeImages, false),
    }

    return projectile
}

func (combat *CombatScreen) CreateIceBoltProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 11)

    loopImages := images[0:3]
    explodeImages := images[3:]

    combat.Projectiles = append(combat.Projectiles, combat.createSkyProjectile(target, loopImages, explodeImages))
}

func (combat *CombatScreen) CreateFireBoltProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 0)
    loopImages := images[0:3]
    explodeImages := images[3:]

    combat.Projectiles = append(combat.Projectiles, combat.createSkyProjectile(target, loopImages, explodeImages))
}

func (combat *CombatScreen) CreateFireballProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 23)

    loopImages := images[0:11]
    explodeImages := images[11:]

    combat.Projectiles = append(combat.Projectiles, combat.createSkyProjectile(target, loopImages, explodeImages))
}

func (combat *CombatScreen) CreateStarFiresProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 9)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle))
}

func (combat *CombatScreen) CreateDispelEvilProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 10)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle))
}

func (combat *CombatScreen) CreatePsionicBlastProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 16)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle))
}

func (combat *CombatScreen) CreateDoomBoltProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 5)
    loopImages := images[0:3]
    explodeImages := images[3:]

    combat.Projectiles = append(combat.Projectiles, combat.createVerticalSkyProjectile(target, loopImages, explodeImages))
}

func (combat *CombatScreen) CreateLightningBoltProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 24)
    // loopImages := images
    explodeImages := images

    screenX, screenY := combat.Coordinates.Apply(float64(target.X), float64(target.Y))
    screenY += 3
    screenX += 5

    screenY -= float64(images[0].Bounds().Dy())

    projectile := &Projectile{
        X: screenX,
        Y: screenY,
        Speed: 0,
        Angle: 0,
        TargetX: screenX,
        TargetY: screenY,
        Animation: util.MakeAnimation(images, true),
        Explode: util.MakeRepeatAnimation(explodeImages, 2),
        Exploding: true,
    }

    combat.Projectiles = append(combat.Projectiles, projectile)
}

func (combat *CombatScreen) CreateWarpLightningProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 3)
    // loopImages := images
    explodeImages := images

    screenX, screenY := combat.Coordinates.Apply(float64(target.X), float64(target.Y))
    screenY += 13
    screenX += 3

    screenY -= float64(images[0].Bounds().Dy())

    projectile := &Projectile{
        X: screenX,
        Y: screenY,
        Speed: 0,
        Angle: 0,
        TargetX: screenX,
        TargetY: screenY,
        Animation: util.MakeAnimation(images, true),
        Explode: util.MakeRepeatAnimation(explodeImages, 2),
        Exploding: true,
    }

    combat.Projectiles = append(combat.Projectiles, projectile)
}

func (combat *CombatScreen) CreateLifeDrainProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 6)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle))
}

func (combat *CombatScreen) CreateFlameStrikeProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 33)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle))
}

func (combat *CombatScreen) CreateRecallHeroProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 5)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle))
}

func (combat *CombatScreen) CreateHealingProjectile(target *ArmyUnit) {
    // FIXME: the images should be mostly with with transparency
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 3)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle))
}

func (combat *CombatScreen) CreateHolyWordProjectile(target *ArmyUnit) {
    // FIXME: the images should be mostly with with transparency
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 3)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle))
}

func (combat *CombatScreen) CreateWebProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 13)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle))
}

func (combat *CombatScreen) CreateDeathSpellProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 14)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle))
}

func (combat *CombatScreen) CreateWordOfDeathProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 14)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle))
}

func (combat *CombatScreen) CreateWarpWoodProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 2)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle))
}

func (combat *CombatScreen) CreateDisintegrateProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 4)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle))
}

func (combat *CombatScreen) CreateWordOfRecallProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 1)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle))
}

func (combat *CombatScreen) CreateDispelMagicProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 26)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle))
}

func (combat *CombatScreen) CreateCracksCallProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 15)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionUnder))
}

func (combat *CombatScreen) CreateBanishProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 19)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionUnder))
}

func (combat *CombatScreen) CreateDisruptProjectile(x int, y int) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 1)

    var loopImages []*ebiten.Image
    explodeImages := images

    fakeTarget := ArmyUnit{
        X: x,
        Y: y,
    }

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(&fakeTarget, loopImages, explodeImages, UnitPositionUnder))
}

func (combat *CombatScreen) CreateSummoningCircle(x int, y int) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 22)
    var loopImages []*ebiten.Image
    explodeImages := images

    fakeTarget := ArmyUnit{
        X: x,
        Y: y,
    }

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(&fakeTarget, loopImages, explodeImages, UnitPositionUnder))
}

func (combat *CombatScreen) CreateMagicVortex(x int, y int) {
    images, _ := combat.ImageCache.GetImages("cmbmagic.lbx", 120)

    unit := &CombatUnit{
        X: x,
        Y: y,
        Animation: util.MakeAnimation(images, true),
    }

    combat.OtherUnits = append(combat.OtherUnits, unit)
}

func (combat *CombatScreen) addNewUnit(player *playerlib.Player, x int, y int, unit units.Unit, facing units.Facing) {
    newUnit := ArmyUnit{
        Unit: unit,
        Facing: facing,
        Moving: false,
        X: x,
        Y: y,
        Health: 10, // FIXME: figures * hitpoints?
        MovesLeft: fraction.FromInt(unit.MovementSpeed),
        LastTurn: combat.CurrentTurn-1,
    }

    if player == combat.DefendingArmy.Player {
        newUnit.Team = TeamDefender
        combat.DefendingArmy.Units = append(combat.DefendingArmy.Units, &newUnit)
    } else {
        newUnit.Team = TeamAttacker
        combat.AttackingArmy.Units = append(combat.AttackingArmy.Units, &newUnit)
    }
}

func (combat *CombatScreen) CreatePhantomWarriors(player *playerlib.Player, x int, y int) {
    // FIXME: compute facing based on player
    combat.addNewUnit(player, x, y, units.PhantomWarrior, units.FacingDown)
}

func (combat *CombatScreen) CreatePhantomBeast(player *playerlib.Player, x int, y int) {
    combat.addNewUnit(player, x, y, units.PhantomBeast, units.FacingDown)
}

func (combat *CombatScreen) CreateEarthElemental(player *playerlib.Player, x int, y int) {
    combat.addNewUnit(player, x, y, units.EarthElemental, units.FacingDown)
}

func (combat *CombatScreen) CreateAirElemental(player *playerlib.Player, x int, y int) {
    combat.addNewUnit(player, x, y, units.AirElemental, units.FacingDown)
}

func (combat *CombatScreen) CreateFireElemental(player *playerlib.Player, x int, y int) {
    combat.addNewUnit(player, x, y, units.FireElemental, units.FacingDown)
}

func (combat *CombatScreen) CreateDemon(player *playerlib.Player, x int, y int) {
    combat.addNewUnit(player, x, y, units.Demon, units.FacingDown)
}

/* let the user select a target, then cast the spell on that target
 */
func (combat *CombatScreen) DoTargetUnitSpell(player *playerlib.Player, spell spellbook.Spell, targetKind Targeting, onTarget func(*ArmyUnit), canTarget func(*ArmyUnit) bool) {
    teamAttacked := TeamAttacker

    if targetKind == TargetFriend {
        /* if the player is the defender and we are targeting a friend then the team should be the defenders */
        if combat.DefendingArmy.Player == player {
            teamAttacked = TeamDefender
        }
    } else if targetKind == TargetEnemy {
        /* if the player is the attacker and we are targeting an enemy then the team should be the defenders */
        if combat.AttackingArmy.Player == player {
            teamAttacked = TeamDefender
        }
    } else if targetKind == TargetEither {
        teamAttacked = TeamEither
    }

    // log.Printf("Create sound for spell %v: %v", spell.Name, spell.Sound)

    x := 250
    if player == combat.DefendingArmy.Player {
        x = 3
    }

    y := 168

    var elements []*uilib.UIElement

    removeElements := func(){
        combat.UI.RemoveElements(elements)
    }

    selectElement := &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            combat.WhiteFont.PrintWrap(screen, float64(x), float64(y), 75, 1, ebiten.ColorScale{}, fmt.Sprintf("Select a target for a %v spell.", spell.Name))
        },
    }

    cancelImages, _ := combat.ImageCache.GetImages("compix.lbx", 22)
    cancelRect := image.Rect(0, 0, cancelImages[0].Bounds().Dx(), cancelImages[0].Bounds().Dy()).Add(image.Point{x + 15, y + 15})
    cancelIndex := 0
    cancelElement := &uilib.UIElement{
        Rect: cancelRect,
        LeftClick: func(element *uilib.UIElement){
            cancelIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            cancelIndex = 0
            combat.DoSelectUnit = false
            combat.SelectTarget = func(target *ArmyUnit){}
            combat.CanTarget = func(target *ArmyUnit) bool { return false }
            removeElements()
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(cancelRect.Min.X), float64(cancelRect.Min.Y))
            screen.DrawImage(cancelImages[cancelIndex], &options)
        },
    }

    elements = append(elements, selectElement, cancelElement)

    combat.UI.AddElements(elements)

    combat.DoSelectUnit = true
    combat.SelectTeam = teamAttacked
    combat.CanTarget = canTarget
    combat.SelectTarget = func(target *ArmyUnit){
        sound, err := audio.LoadSound(combat.Cache, spell.Sound)
        if err == nil {
            sound.Play()
        } else {
            log.Printf("No such sound %v for %v: %v", spell.Sound, spell.Name, err)
        }

        removeElements()
        onTarget(target)

        combat.SelectTarget = func(*ArmyUnit){}
        combat.CanTarget = func(*ArmyUnit) bool { return false }
    }
}

// FIXME: take in a canTarget function to check if the tile is legal
func (combat *CombatScreen) DoTargetTileSpell(player *playerlib.Player, spell spellbook.Spell, onTarget func(int, int)){
    // log.Printf("Create sound for spell %v: %v", spell.Name, spell.Sound)

    x := 250
    if player == combat.DefendingArmy.Player {
        x = 3
    }

    y := 168

    var elements []*uilib.UIElement

    removeElements := func(){
        combat.UI.RemoveElements(elements)
    }

    selectElement := &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            combat.WhiteFont.PrintWrap(screen, float64(x), float64(y), 75, 1, ebiten.ColorScale{}, fmt.Sprintf("Select a target for a %v spell.", spell.Name))
        },
    }

    cancelImages, _ := combat.ImageCache.GetImages("compix.lbx", 22)
    cancelRect := image.Rect(0, 0, cancelImages[0].Bounds().Dx(), cancelImages[0].Bounds().Dy()).Add(image.Point{x + 15, y + 15})
    cancelIndex := 0
    cancelElement := &uilib.UIElement{
        Rect: cancelRect,
        LeftClick: func(element *uilib.UIElement){
            cancelIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            cancelIndex = 0
            combat.DoSelectTile = false
            combat.SelectTile = func(x int, y int){}
            removeElements()
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(cancelRect.Min.X), float64(cancelRect.Min.Y))
            screen.DrawImage(cancelImages[cancelIndex], &options)
        },
    }

    elements = append(elements, selectElement, cancelElement)

    combat.UI.AddElements(elements)

    combat.DoSelectTile = true
    combat.SelectTile = func(x int, y int){
        sound, err := audio.LoadSound(combat.Cache, spell.Sound)
        if err == nil {
            sound.Play()
        } else {
            log.Printf("No such sound %v for %v: %v", spell.Sound, spell.Name, err)
        }

        removeElements()
        onTarget(x, y)

        combat.SelectTile = func(int, int){}
    }
}

func (combat *CombatScreen) DoSummoningSpell(player *playerlib.Player, spell spellbook.Spell, onTarget func(int, int)){
    // FIXME: pass in a canTarget function that only allows summoning on an empty tile on the casting wizards side of the battlefield
    combat.DoTargetTileSpell(player, spell, func (x int, y int){
        combat.CreateSummoningCircle(x, y)
        // FIXME: there should be a delay between the summoning circle appearing and when the unit appears
        onTarget(x, y)
    })
}

/* create projectiles on all units immediately, no targeting required
 */
func (combat *CombatScreen) DoAllUnitsSpell(player *playerlib.Player, spell spellbook.Spell, targetKind Targeting, onTarget func(*ArmyUnit), canTarget func(*ArmyUnit) bool) {
    var units []*ArmyUnit

    if player == combat.DefendingArmy.Player && targetKind == TargetEnemy {
        units = combat.AttackingArmy.Units
    } else if player == combat.AttackingArmy.Player && targetKind == TargetEnemy {
        units = combat.DefendingArmy.Units
    } else if player == combat.DefendingArmy.Player && targetKind == TargetFriend {
        units = combat.DefendingArmy.Units
    } else if player == combat.AttackingArmy.Player && targetKind == TargetFriend {
        units = combat.AttackingArmy.Units
    }

    sound, err := audio.LoadSound(combat.Cache, spell.Sound)
    if err == nil {
        sound.Play()
    } else {
        log.Printf("No such sound %v for %v: %v", spell.Sound, spell.Name, err)
    }

    for _, unit := range units {
        if canTarget(unit){
            onTarget(unit)
        }
    }
}

/* makes a 5x5 square of tiles have mud on them
 */
func (combat *CombatScreen) CreateEarthToMud(centerX int, centerY int){
    // log.Printf("Create earth to mud at %v, %v", centerX, centerY)

    for x := centerX - 2; x <= centerX + 2; x++ {
        for y := centerY - 2; y <= centerY + 2; y++ {
            if x >= 0 && x < len(combat.Tiles[0]) && y >= 0 && y < len(combat.Tiles) {
                combat.Tiles[y][x].Mud = true
            }
        }
    }
}

func (combat *CombatScreen) FindEmptyTile() (int, int, error) {

    middleX := len(combat.Tiles[0]) / 2
    middleY := len(combat.Tiles) / 2

    distance := 3
    tries := 0
    for tries < 100 {
        x := middleX + rand.Intn(distance) - distance/2
        y := middleY + rand.Intn(distance) - distance/2

        if x >= 0 && x < len(combat.Tiles[0]) && y >= 0 && y < len(combat.Tiles) && combat.GetUnit(x, y) == nil {
            return x, y, nil
        }

        distance += 1
        if distance > len(combat.Tiles) * 2 {
            distance = len(combat.Tiles) * 2
        }
    }

    return -1, -1, fmt.Errorf("unable to find a free tile")
}

func (combat *CombatScreen) InvokeSpell(player *playerlib.Player, spell spellbook.Spell){
    targetAny := func (target *ArmyUnit) bool { return true }

    switch spell.Name {
        case "Fireball":
            combat.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                combat.CreateFireballProjectile(target)
            }, targetAny)
        case "Ice Bolt":
            combat.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                combat.CreateIceBoltProjectile(target)
            }, targetAny)
        case "Star Fires":
            combat.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                combat.CreateStarFiresProjectile(target)
            }, func (target *ArmyUnit) bool {
                // FIXME: can only target fantastic creatures that are death or chaos
                return true
            })
        case "Psionic Blast":
            combat.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                combat.CreatePsionicBlastProjectile(target)
            }, targetAny)
        case "Doom Bolt":
            combat.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                combat.CreateDoomBoltProjectile(target)
            }, targetAny)
        case "Fire Bolt":
            combat.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                combat.CreateFireBoltProjectile(target)
            }, targetAny)
        case "Lightning Bolt":
            combat.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                combat.CreateLightningBoltProjectile(target)
            }, targetAny)
        case "Warp Lightning":
            combat.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                combat.CreateWarpLightningProjectile(target)
            }, targetAny)
        case "Flame Strike":
            combat.DoAllUnitsSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                combat.CreateFlameStrikeProjectile(target)
            }, targetAny)
        case "Life Drain":
            combat.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                combat.CreateLifeDrainProjectile(target)
            }, targetAny)
        case "Dispel Evil":
            combat.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                combat.CreateDispelEvilProjectile(target)
            }, func (target *ArmyUnit) bool {
                // FIXME: can only target units that are death or chaos
                return true
            })
        case "Healing":
            combat.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                combat.CreateHealingProjectile(target)
            }, func (target *ArmyUnit) bool {
                // FIXME: can only target units that are not death
                return true
            })
        case "Holy Word":
            combat.DoAllUnitsSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                combat.CreateHolyWordProjectile(target)
            }, func (target *ArmyUnit) bool {
                // FIXME: can only target fantastic units, chaos channeled and undead
                return true
            })
        case "Recall Hero":
            combat.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                combat.CreateRecallHeroProjectile(target)
            }, func (target *ArmyUnit) bool {
                // FIXME: can only target heros
                return true
            })
        case "Mass Healing":
            combat.DoAllUnitsSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                combat.CreateHealingProjectile(target)
            }, targetAny)
        case "Cracks Call":
            combat.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                combat.CreateCracksCallProjectile(target)
            }, targetAny)
        case "Earth to Mud":
            combat.DoTargetTileSpell(player, spell, func (x int, y int){
                combat.CreateEarthToMud(x, y)
            })
        case "Web":
            combat.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                combat.CreateWebProjectile(target)
            }, targetAny)
        case "Banish":
            combat.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                combat.CreateBanishProjectile(target)
            }, func (target *ArmyUnit) bool {
                // FIXME: must be a fantastic unit
                return true
            })
        case "Dispel Magic True":
            combat.DoTargetUnitSpell(player, spell, TargetEither, func(target *ArmyUnit){
                combat.CreateDispelMagicProjectile(target)
            }, targetAny)
        case "Word of Recall":
            combat.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                combat.CreateWordOfRecallProjectile(target)
            }, targetAny)
        case "Disintegrate":
            combat.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                combat.CreateDisintegrateProjectile(target)
            }, targetAny)
        case "Disrupt":
            // FIXME: can only target city walls
            combat.DoTargetTileSpell(player, spell, func (x int, y int){
                combat.CreateDisruptProjectile(x, y)
            })
        case "Magic Vortex":
            combat.DoTargetTileSpell(player, spell, func (x int, y int){
                combat.CreateMagicVortex(x, y)
            })
        case "Warp Wood":
            combat.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                combat.CreateWarpWoodProjectile(target)
            }, func (target *ArmyUnit) bool {
                // FIXME: can be cast on a normal unit or hero that has a ranged missle attack
                return true
            })
        case "Death Spell":
            combat.DoAllUnitsSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                combat.CreateDeathSpellProjectile(target)
            }, targetAny)
        case "Word of Death":
            combat.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                combat.CreateWordOfDeathProjectile(target)
            }, targetAny)
        case "Phantom Warriors":
            combat.DoSummoningSpell(player, spell, func(x int, y int){
                combat.CreatePhantomWarriors(player, x, y)
            })
        case "Phantom Beast":
            combat.DoSummoningSpell(player, spell, func(x int, y int){
                combat.CreatePhantomBeast(player, x, y)
            })
        case "Earth Elemental":
            combat.DoSummoningSpell(player, spell, func(x int, y int){
                combat.CreateEarthElemental(player, x, y)
            })
        case "Air Elemental":
            combat.DoSummoningSpell(player, spell, func(x int, y int){
                combat.CreateAirElemental(player, x, y)
            })
        case "Fire Elemental":
            combat.DoSummoningSpell(player, spell, func(x int, y int){
                combat.CreateFireElemental(player, x, y)
            })
        case "Summon Demon":
            // FIXME: the tile should be near the middle of the map
            x, y, err := combat.FindEmptyTile()
            if err == nil {
                combat.CreateSummoningCircle(x, y)
                combat.CreateDemon(player, x, y)
            }

            /*
Disenchant Area - need picture
Dispel Magic - need picture
Raise Dead - need picture
Petrify	- need picture
Disenchant True - need picture
Call Chaos - need picture
Animate Dead - need picture
            */

    }
}

func (combat *CombatScreen) MakeUI(player *playerlib.Player) *uilib.UI {
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

            if combat.AttackingArmy.Player == player && (combat.DoSelectUnit || combat.DoSelectTile) {
            } else {
                combat.AttackingWizardFont.Print(screen, 265, 170, 1, ebiten.ColorScale{}, combat.AttackingArmy.Player.Wizard.Name)
            }

            if combat.DefendingArmy.Player == player && (combat.DoSelectUnit || combat.DoSelectTile) {
            } else {
                combat.DefendingWizardFont.Print(screen, 30, 170, 1, ebiten.ColorScale{}, combat.DefendingArmy.Player.Wizard.Name)
            }

            rightImage, _ := combat.ImageCache.GetImage(combat.SelectedUnit.Unit.CombatLbxFile, combat.SelectedUnit.Unit.GetCombatIndex(units.FacingRight), 0)
            options.GeoM.Reset()
            options.GeoM.Translate(89, 170)
            screen.DrawImage(rightImage, &options)

            combat.HudFont.Print(screen, 92, 167, 1, ebiten.ColorScale{}, combat.SelectedUnit.Unit.Name)

            plainAttack, _ := combat.ImageCache.GetImage("compix.lbx", 29, 0)
            options.GeoM.Reset()
            options.GeoM.Translate(126, 173)
            screen.DrawImage(plainAttack, &options)
            combat.HudFont.PrintRight(screen, 126, 174, 1, ebiten.ColorScale{}, fmt.Sprintf("%v", combat.SelectedUnit.Unit.MeleeAttackPower))

            var movementImage *ebiten.Image
            if combat.SelectedUnit.Unit.Flying {
                movementImage, _ = combat.ImageCache.GetImage("compix.lbx", 39, 0)
            } else {
                movementImage, _ = combat.ImageCache.GetImage("compix.lbx", 38, 0)
            }

            options.GeoM.Reset()
            options.GeoM.Translate(126, 188)
            screen.DrawImage(movementImage, &options)
            combat.HudFont.PrintRight(screen, 126, 190, 1, ebiten.ColorScale{}, fmt.Sprintf("%v", combat.SelectedUnit.MovesLeft.ToFloat()))

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })
        },
    }

    buttonX := float64(139)
    buttonY := float64(167)

    makeButton := func(lbxIndex int, x int, y int, action func()) *uilib.UIElement {
        buttons, _ := combat.ImageCache.GetImages("compix.lbx", lbxIndex)
        rect := image.Rect(0, 0, buttons[0].Bounds().Dx(), buttons[0].Bounds().Dy()).Add(image.Point{int(buttonX) + buttons[0].Bounds().Dx() * x, int(buttonY) + buttons[0].Bounds().Dy() * y})
        index := 0
        return &uilib.UIElement{
            Rect: rect,
            LeftClick: func(element *uilib.UIElement){
                action()
                index = 1
            },
            LeftClickRelease: func(element *uilib.UIElement){
                index = 0
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
                screen.DrawImage(buttons[index], &options)
            },
        }
    }

    // spell
    elements = append(elements, makeButton(1, 0, 0, func(){
        spellUI := spellbook.MakeSpellBookCastUI(ui, combat.Cache, player.Spells, player.CastingSkill, func (spell spellbook.Spell, picked bool){
            if picked {
                // player mana and skill should go down accordingly
                combat.InvokeSpell(player, spell)
            }
        })
        ui.AddElements(spellUI)
    }))

    // wait
    elements = append(elements, makeButton(2, 1, 0, func(){
        combat.NextUnit()
    }))

    // info
    elements = append(elements, makeButton(20, 0, 1, func(){
        // FIXME
    }))

    // auto
    elements = append(elements, makeButton(4, 1, 1, func(){
        // FIXME
    }))

    // flee
    elements = append(elements, makeButton(21, 0, 2, func(){
        // FIXME
    }))

    // done
    elements = append(elements, makeButton(3, 1, 2, func(){
        combat.SelectedUnit.LastTurn = combat.CurrentTurn
        combat.NextUnit()
    }))

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

/* choose a unit from the given team such that
 * the unit's LastTurn is less than the current turn
 */
func (combat *CombatScreen) ChooseNextUnit(team Team) *ArmyUnit {

    switch team {
        case TeamAttacker:
            for i := 0; i < len(combat.AttackingArmy.Units); i++ {
                combat.TurnAttacker = (combat.TurnAttacker + 1) % len(combat.AttackingArmy.Units)
                unit := combat.AttackingArmy.Units[combat.TurnAttacker]
                if unit.LastTurn < combat.CurrentTurn {
                    unit.Paths = make(map[image.Point]pathfinding.Path)
                    return unit
                }
            }
            return nil
        case TeamDefender:
            for i := 0; i < len(combat.DefendingArmy.Units); i++ {
                combat.TurnDefender = (combat.TurnDefender + 1) % len(combat.DefendingArmy.Units)
                unit := combat.DefendingArmy.Units[combat.TurnDefender]
                if unit.LastTurn < combat.CurrentTurn {
                    unit.Paths = make(map[image.Point]pathfinding.Path)
                    return unit
                }
            }
            return nil
    }

    return nil
}

func (combat *CombatScreen) NextTurn() {
    combat.CurrentTurn += 1

    /* reset movement */
    for _, unit := range combat.DefendingArmy.Units {
        unit.MovesLeft = fraction.FromInt(unit.Unit.MovementSpeed)
        unit.Paths = make(map[image.Point]pathfinding.Path)
    }

    for _, unit := range combat.AttackingArmy.Units {
        unit.MovesLeft = fraction.FromInt(unit.Unit.MovementSpeed)
        unit.Paths = make(map[image.Point]pathfinding.Path)
    }
}

func (combat *CombatScreen) NextUnit() {

    var nextChoice *ArmyUnit
    for i := 0; i < 2; i++ {
        // find a unit on the same team
        nextChoice = combat.ChooseNextUnit(combat.Turn)
        if nextChoice == nil {
            // if there are no available units then the team must be out of moves, so try the next team
            combat.Turn = oppositeTeam(combat.Turn)
            nextChoice = combat.ChooseNextUnit(combat.Turn)

            if nextChoice == nil {
                // if the other team still has nothing available then the entire turn has finished
                // so go to the next turn and try again
                combat.NextTurn()
                combat.SelectedUnit = nil
            }
        }

        // found something so break the loop
        if nextChoice != nil {
            break
        }
    }

    combat.SelectedUnit = nextChoice
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

func (combat *CombatScreen) withinArrowRange(attacker *ArmyUnit, defender *ArmyUnit) bool {
    /*
    xDiff := math.Abs(float64(attacker.X - defender.X))
    yDiff := math.Abs(float64(attacker.Y - defender.Y))

    return xDiff <= 1 && yDiff <= 1
    */
    // FIXME: what is the actual range distance?
    return false
}

func (combat *CombatScreen) canAttack(attacker *ArmyUnit, defender *ArmyUnit) bool {
    if attacker.MovesLeft.LessThanEqual(fraction.FromInt(0)) {
        return false
    }

    if defender.Unit.Flying && !attacker.Unit.Flying {
        return false
    }

    if attacker.Team == defender.Team {
        return false
    }

    return true
}

func distanceInRange(x1 float64, y1 float64, x2 float64, y2 float64, r float64) bool {
    xDiff := x2 - x1
    yDiff := y2 - y1
    return xDiff * xDiff + yDiff * yDiff <= r*r
}

func (combat *CombatScreen) UpdateProjectiles() bool {
    animationSpeed := uint64(5)

    alive := len(combat.Projectiles) > 0

    var projectilesOut []*Projectile
    for _, projectile := range combat.Projectiles {
        keep := false
        if projectile.Exploding || distanceInRange(projectile.X, projectile.Y, projectile.TargetX, projectile.TargetY, 5) {
            projectile.Exploding = true
            keep = true
            if combat.Counter % animationSpeed == 0 && !projectile.Explode.Next() {
                keep = false
            }
        } else {
            projectile.X += math.Cos(projectile.Angle) * projectile.Speed
            projectile.Y += math.Sin(projectile.Angle) * projectile.Speed
            if combat.Counter % animationSpeed == 0 {
                projectile.Animation.Next()
            }
            keep = true
        }

        if keep {
            projectilesOut = append(projectilesOut, projectile)
        }
    }

    combat.Projectiles = projectilesOut

    return alive
}

func (combat *CombatScreen) Update() CombatState {
    combat.Counter += 1

    combat.UI.StandardUpdate()

    mouseX, mouseY := ebiten.CursorPosition()
    hudImage, _ := combat.ImageCache.GetImage("cmbtfx.lbx", 28, 0)

    tileX, tileY := combat.ScreenToTile.Apply(float64(mouseX), float64(mouseY))
    combat.MouseTileX = int(math.Round(tileX))
    combat.MouseTileY = int(math.Round(tileY))

    for _, unit := range combat.OtherUnits {
        if combat.Counter % 6 == 0 {
            unit.Animation.Next()
        }
    }

    hudY := data.ScreenHeight - hudImage.Bounds().Dy()

    if combat.DoSelectTile {
        combat.MouseState = CombatCast

        if mouseY >= hudY {
            combat.MouseState = CombatClickHud
            return CombatStateRunning
        }

        if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && mouseY < hudY {
            combat.SelectTile(combat.MouseTileX, combat.MouseTileY)
            combat.DoSelectTile = false
        }

        return CombatStateRunning
    }

    if combat.DoSelectUnit {
        combat.MouseState = CombatCast

        if mouseY >= hudY {
            combat.MouseState = CombatClickHud
            return CombatStateRunning
        }

        unit := combat.GetUnit(combat.MouseTileX, combat.MouseTileY)
        if unit == nil || (combat.SelectTeam != TeamEither && unit.Team != combat.SelectTeam) || !combat.CanTarget(unit){
            combat.MouseState = CombatNotOk
        }

        if combat.CanTarget(unit) && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && mouseY < hudY {
            // log.Printf("Click unit at %v,%v -> %v", combat.MouseTileX, combat.MouseTileY, unit)
            if unit != nil && (combat.SelectTeam == TeamEither || unit.Team == combat.SelectTeam) {
                combat.SelectTarget(unit)
                combat.DoSelectUnit = false

                // shouldn't need to set the mouse state here
                combat.MouseState = CombatClickHud
            }
        }

        return CombatStateRunning
    }

    if combat.UpdateProjectiles() {
        combat.UI.Disable()
        return CombatStateRunning
    }

    combat.UI.Enable()

    if combat.UI.GetHighestLayerValue() > 0 || mouseY >= hudY {
        combat.MouseState = CombatClickHud
    } else if combat.SelectedUnit != nil && combat.SelectedUnit.Moving {
        combat.MouseState = CombatClickHud
    } else {
        who := combat.GetUnit(combat.MouseTileX, combat.MouseTileY)
        if who == nil {
            if combat.CanMoveTo(combat.SelectedUnit, combat.MouseTileX, combat.MouseTileY) {
                combat.MouseState = CombatMoveOk
            } else {
                combat.MouseState = CombatNotOk
            }
        } else {
            if combat.canAttack(combat.SelectedUnit, who){
                if combat.withinMeleeRange(combat.SelectedUnit, who) {
                    combat.MouseState = CombatMeleeAttackOk
                } else if combat.withinArrowRange(combat.SelectedUnit, who) {
                    combat.MouseState = CombatRangeAttackOk
                }
            } else {
                combat.MouseState = CombatNotOk
            }
        }
    }

    // if there is no unit at the tile position then the highlighted unit will be nil
    combat.HighlightedUnit = combat.GetUnit(combat.MouseTileX, combat.MouseTileY)

    // dont allow clicks into the hud area
    // also don't allow clicks into the game if the ui is showing some overlay
    if combat.UI.GetHighestLayerValue() == 0 &&
       inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) &&
       mouseY < hudY &&
       combat.SelectedUnit.Moving == false && combat.SelectedUnit.Attacking == false {

        if combat.TileIsEmpty(combat.MouseTileX, combat.MouseTileY) && combat.CanMoveTo(combat.SelectedUnit, combat.MouseTileX, combat.MouseTileY){
            path, _ := combat.FindPath(combat.SelectedUnit, combat.MouseTileX, combat.MouseTileY)
            combat.SelectedUnit.MovementTick = combat.Counter
            combat.SelectedUnit.MovementPath = path[1:]
            /*
            combat.SelectedUnit.TargetX = combat.MouseTileX
            combat.SelectedUnit.TargetY = combat.MouseTileY
            */
            combat.SelectedUnit.Moving = true
            combat.SelectedUnit.MoveX = float64(combat.SelectedUnit.X)
            combat.SelectedUnit.MoveY = float64(combat.SelectedUnit.Y)
            /*
            combat.SelectedUnit.MovesLeft = combat.SelectedUnit.MovesLeft.Subtract(computeMoves(combat.SelectedUnit.X, combat.SelectedUnit.Y, combat.MouseTileX, combat.MouseTileY))
            if combat.SelectedUnit.MovesLeft.LessThan(fraction.FromInt(0)) {
                combat.SelectedUnit.MovesLeft = fraction.FromInt(0)
            }
            */
       } else {

           defender := combat.GetUnit(combat.MouseTileX, combat.MouseTileY)

           if defender != nil && defender.Team != combat.SelectedUnit.Team && combat.withinMeleeRange(combat.SelectedUnit, defender) && combat.canAttack(combat.SelectedUnit, defender){
               combat.SelectedUnit.Attacking = true
               combat.SelectedUnit.AttackingCounter = combat.Counter
               // attacking takes 50% of movement points
               // FIXME: in some cases an extra 0.5 movements points is lost, possibly due to counter attacks?
               combat.SelectedUnit.MovesLeft = combat.SelectedUnit.MovesLeft.Subtract(fraction.FromInt(combat.SelectedUnit.Unit.MovementSpeed).Divide(fraction.FromInt(2)))
               if combat.SelectedUnit.MovesLeft.LessThan(fraction.FromInt(0)) {
                   combat.SelectedUnit.MovesLeft = fraction.FromInt(0)
               }

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

            if combat.SelectedUnit.MovesLeft.LessThanEqual(fraction.FromInt(0)) {
                combat.SelectedUnit.LastTurn = combat.CurrentTurn
                combat.NextUnit()
            }
        }
    }

    if combat.SelectedUnit.Moving {
        targetX, targetY := combat.SelectedUnit.MovementPath[0].X, combat.SelectedUnit.MovementPath[0].Y

        angle := math.Atan2(float64(targetY) - combat.SelectedUnit.MoveY, float64(targetX) - combat.SelectedUnit.MoveX)

        // rotate by 45 degrees to get the on screen facing angle
        // have to negate the angle because the y axis is flipped (higher y values are lower on the screen)
        useAngle := -(angle - math.Pi/4)

        // log.Printf("Angle: %v from (%v,%v) to (%v,%v)", useAngle, combat.SelectedUnit.X, combat.SelectedUnit.Y, combat.SelectedUnit.TargetX, combat.SelectedUnit.TargetY)

        combat.SelectedUnit.Facing = computeFacing(useAngle)

        // speed := float64(combat.Counter - combat.SelectedUnit.MovementTick) / 4
        speed := float64(0.09)
        combat.SelectedUnit.MoveX += math.Cos(angle) * speed
        combat.SelectedUnit.MoveY += math.Sin(angle) * speed

        // log.Printf("Moving %v,%v -> %v,%v", combat.SelectedUnit.X, combat.SelectedUnit.Y, combat.SelectedUnit.MoveX, combat.SelectedUnit.MoveY)

        /*
        combat.SelectedUnit.MoveX = newX
        combat.SelectedUnit.MoveY = newY
        */

        // if math.Abs(combat.SelectedUnit.MoveX - float64(targetX)) < speed*2 && math.Abs(combat.SelectedUnit.MoveY - float64(targetY)) < 0.5 {
        if distanceInRange(combat.SelectedUnit.MoveX, combat.SelectedUnit.MoveY, float64(targetX), float64(targetY), speed * 3) {
            combat.SelectedUnit.MovesLeft = combat.SelectedUnit.MovesLeft.Subtract(pathCost(image.Pt(combat.SelectedUnit.X, combat.SelectedUnit.Y), image.Pt(targetX, targetY)))
            combat.SelectedUnit.X = targetX
            combat.SelectedUnit.Y = targetY
            combat.SelectedUnit.MoveX = float64(targetX)
            combat.SelectedUnit.MoveY = float64(targetY)
            combat.SelectedUnit.MovementPath = combat.SelectedUnit.MovementPath[1:]

            if len(combat.SelectedUnit.MovementPath) == 0 {
                combat.SelectedUnit.Moving = false
                // reset path computations
                combat.SelectedUnit.Paths = make(map[image.Point]pathfinding.Path)

                if combat.SelectedUnit.MovesLeft.LessThanEqual(fraction.FromInt(0)) {
                    combat.SelectedUnit.LastTurn = combat.CurrentTurn
                    combat.NextUnit()
                }
            }
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

    for _, point := range combat.TopDownOrder {
        x := point.X
        y := point.Y

        image, _ := combat.ImageCache.GetImage("cmbgrass.lbx", combat.Tiles[y][x].Index, 0)
        options.GeoM.Reset()
        tx, ty := tilePosition(x, y)
        options.GeoM.Translate(tx, ty)
        screen.DrawImage(image, &options)

        if combat.Tiles[y][x].Mud {
            mudTiles, _ := combat.ImageCache.GetImages("cmbtcity.lbx", 118)
            index := animationIndex % uint64(len(mudTiles))
            screen.DrawImage(mudTiles[index], &options)
        }

        if combat.Tiles[y][x].ExtraObject != -1 {
            extraImage, _ := combat.ImageCache.GetImage("cmbgrass.lbx", 48 + combat.Tiles[y][x].ExtraObject, 0)
            screen.DrawImage(extraImage, &options)
        }
    }


    // draw base land
    /*
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

            if combat.Tiles[y][x].Mud {
                mudTiles, _ := combat.ImageCache.GetImages("cmbtcity.lbx", 118)
                index := animationIndex % uint64(len(mudTiles))
                screen.DrawImage(mudTiles[index], &options)
            }

            if combat.Tiles[y][x].ExtraObject != -1 {
                extraImage, _ := combat.ImageCache.GetImage("cmbgrass.lbx", 48 + combat.Tiles[y][x].ExtraObject, 0)
                screen.DrawImage(extraImage, &options)
            }
        }
    }
    */

    combat.DrawHighlightedTile(screen, combat.MouseTileX, combat.MouseTileY, color.RGBA{R: 0, G: 0x67, B: 0x78, A: 255}, color.RGBA{R: 0, G: 0xef, B: 0xff, A: 255})

    if combat.SelectedUnit != nil {
        path, ok := combat.FindPath(combat.SelectedUnit, combat.MouseTileX, combat.MouseTileY)
        if ok {
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(0.8)
            for i := 1; i < len(path); i++ {
                tileX, tileY := path[i].X, path[i].Y

                tx, ty := tilePosition(tileX, tileY)
                tx += float64(tile0.Bounds().Dx())/2
                ty += float64(tile0.Bounds().Dy())/2
                movementImage, _ := combat.ImageCache.GetImage("compix.lbx", 72, 0)
                tx -= float64(movementImage.Bounds().Dx())/2
                ty -= float64(movementImage.Bounds().Dy())/2

                options.GeoM.Reset()
                options.GeoM.Translate(tx, ty)
                screen.DrawImage(movementImage, &options)
            }
        }
    }

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
            var unitOptions ebiten.DrawImageOptions
            unitOptions.GeoM.Reset()
            var tx float64
            var ty float64

            if unit.Moving {
                tx, ty = combat.Coordinates.Apply(unit.MoveX, unit.MoveY)
            } else {
                tx, ty = tilePosition(unit.X, unit.Y)
            }
            unitOptions.GeoM.Translate(tx, ty)
            unitOptions.GeoM.Translate(float64(tile0.Bounds().Dx()/2), float64(tile0.Bounds().Dy()/2))

            index := uint64(0)
            if unit.Unit.Flying || unit.Moving {
                index = animationIndex % (uint64(len(combatImages)) - 1)
            }

            if unit.Attacking {
                index = 2 + animationIndex % 2
            }

            if combat.HighlightedUnit == unit {
                scaleValue := 1.5 + math.Sin(float64(combat.Counter)/5)/2
                unitOptions.ColorScale.Scale(float32(scaleValue), 1, 1, 1)
            }

            RenderCombatUnit(screen, combatImages[index], unitOptions, unit.Unit.Count)
        }
    }

    for _, unit := range combat.DefendingArmy.Units {
        renderUnit(unit)
    }

    for _, unit := range combat.AttackingArmy.Units {
        renderUnit(unit)
    }

    for _, unit := range combat.OtherUnits {
        var unitOptions ebiten.DrawImageOptions
        tx, ty := tilePosition(unit.X, unit.Y)
        unitOptions.GeoM.Translate(tx, ty)
        unitOptions.GeoM.Translate(float64(tile0.Bounds().Dx()/2), float64(tile0.Bounds().Dy()/2))

        frame := unit.Animation.Frame()
        unitOptions.GeoM.Translate(float64(-frame.Bounds().Dx()/2), float64(-frame.Bounds().Dy()))
        screen.DrawImage(frame, &unitOptions)
    }

    combat.UI.Draw(combat.UI, screen)

    if combat.HighlightedUnit != nil {
        x1 := 255 - 1
        y1 := 5
        width := 65
        height := 45
        vector.DrawFilledRect(screen, float32(x1), float32(y1), float32(width), float32(height), color.RGBA{R: 0, G: 0, B: 0, A: 100}, false)
        vector.StrokeRect(screen, float32(x1), float32(y1), float32(width), float32(height), 1, util.PremultiplyAlpha(color.RGBA{R: 0x27, G: 0x4e, B: 0xdc, A: 100}), false)
        combat.InfoFont.PrintCenter(screen, float64(x1 + 35), float64(y1 + 2), 1, ebiten.ColorScale{}, fmt.Sprintf("%v", combat.HighlightedUnit.Unit.Name))

        meleeImage, _ := combat.ImageCache.GetImage("compix.lbx", 61, 0)
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(x1 + 14), float64(y1 + 10))
        screen.DrawImage(meleeImage, &options)
        combat.InfoFont.PrintRight(screen, float64(x1 + 14), float64(y1 + 10 + 2), 1, ebiten.ColorScale{}, fmt.Sprintf("%v", combat.HighlightedUnit.Unit.MeleeAttackPower))

        movementImage, _ := combat.ImageCache.GetImage("compix.lbx", 72, 0)
        if combat.HighlightedUnit.Unit.Flying {
            movementImage, _ = combat.ImageCache.GetImage("compix.lbx", 73, 0)
        }

        options.GeoM.Reset()
        options.GeoM.Translate(float64(x1 + 14), float64(y1 + 26))
        screen.DrawImage(movementImage, &options)
        combat.InfoFont.PrintRight(screen, float64(x1 + 14), float64(y1 + 26 + 2), 1, ebiten.ColorScale{}, fmt.Sprintf("%v", combat.HighlightedUnit.MovesLeft.ToFloat()))

        armorImage, _ := combat.ImageCache.GetImage("compix.lbx", 70, 0)
        options.GeoM.Reset()
        options.GeoM.Translate(float64(x1 + 48), float64(y1 + 10))
        screen.DrawImage(armorImage, &options)
        combat.InfoFont.PrintRight(screen, float64(x1 + 48), float64(y1 + 10 + 2), 1, ebiten.ColorScale{}, fmt.Sprintf("%v", combat.HighlightedUnit.Unit.Defense))

        resistanceImage, _ := combat.ImageCache.GetImage("compix.lbx", 75, 0)
        options.GeoM.Reset()
        options.GeoM.Translate(float64(x1 + 48), float64(y1 + 18))
        screen.DrawImage(resistanceImage, &options)
        combat.InfoFont.PrintRight(screen, float64(x1 + 48), float64(y1 + 18 + 2), 1, ebiten.ColorScale{}, fmt.Sprintf("%v", combat.HighlightedUnit.Unit.Resistance))

        combat.InfoFont.PrintCenter(screen, float64(x1 + 14), float64(y1 + 37), 1, ebiten.ColorScale{}, "Hits")

        healthyColor := color.RGBA{R: 0, G: 0xff, B: 0, A: 0xff}
        // medium healthy is yellow
        // unhealthy is red
        healthWidth := 15

        vector.StrokeLine(screen, float32(x1 + 25), float32(y1 + 40), float32(x1 + 25 + healthWidth), float32(y1 + 40), 1, healthyColor, false)
    }

    for _, projectile := range combat.Projectiles {
        var frame *ebiten.Image
        if projectile.Exploding {
            frame = projectile.Explode.Frame()
        } else {
            frame = projectile.Animation.Frame()
        }
        if frame != nil {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(projectile.X, projectile.Y)
            screen.DrawImage(frame, &options)
        }
    }

    var mouseOptions ebiten.DrawImageOptions
    mouseX, mouseY := ebiten.CursorPosition()
    mouseOptions.GeoM.Translate(float64(mouseX), float64(mouseY))
    switch combat.MouseState {
        case CombatMoveOk:
            screen.DrawImage(combat.Mouse.Move, &mouseOptions)
        case CombatClickHud:
            screen.DrawImage(combat.Mouse.Normal, &mouseOptions)
        case CombatMeleeAttackOk:
            mouseOptions.GeoM.Translate(-1, -1)
            screen.DrawImage(combat.Mouse.Attack, &mouseOptions)
        case CombatRangeAttackOk:
            screen.DrawImage(combat.Mouse.Arrow, &mouseOptions)
        case CombatNotOk:
            mouseOptions.GeoM.Translate(-1, -1)
            screen.DrawImage(combat.Mouse.Error, &mouseOptions)
        case CombatCast:
            index := (combat.Counter / 8) % uint64(len(combat.Mouse.Cast))
            screen.DrawImage(combat.Mouse.Cast[index], &mouseOptions)
    }
}
