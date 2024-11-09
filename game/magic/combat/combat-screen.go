package combat

import (
    "fmt"
    "log"
    "cmp"
    "math"
    "math/rand/v2"
    "image"
    "image/color"
    "time"
    "slices"
    "context"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/mouse"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    globalMouse "github.com/kazzmir/master-of-magic/game/magic/mouse"
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
    CombatStateAttackerWin
    CombatStateDefenderWin
    CombatStateDone
)

type CombatEvent interface {
}

type CombatEventSelectTile struct {
    SelectTile func(int, int)
    Spell spellbook.Spell
    Selecter Team
}

type CombatEventSelectUnit struct {
    SelectTarget func(*ArmyUnit)
    CanTarget func(*ArmyUnit) bool
    Spell spellbook.Spell
    Selecter Team
    SelectTeam Team
}

type ZoneType struct {
    // fighting in a city
    City *citylib.City

    AncientTemple bool
    FallenTemple bool
    Ruins bool
    AbandonedKeep bool
    Lair bool
    Tower bool
    Dungeon bool

    // one of the three node types
    ChaosNode bool
    NatureNode bool
    SorceryNode bool
}

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

type TileAlignment int
const (
    TileAlignMiddle TileAlignment = iota
    TileAlignBottom
)

type TileDraw func(*ebiten.Image, *util.ImageCache, *ebiten.DrawImageOptions, uint64)

type TileTop struct {
    Drawer TileDraw
    Lbx string
    Index int
    Alignment TileAlignment
}

type Tile struct {
    // a unit standing on this tile, if any
    Unit *ArmyUnit
    Lbx string
    // index of grass/floor
    Index int
    // tree/rock on top, or -1 if nothing
    ExtraObject TileTop
    Mud bool
}

type CombatUnit interface {
    HasAbility(units.AbilityType) bool
    GetDefense() int
    GetResistance() int
    AdjustHealth(int)
    GetBanner() data.BannerType
    GetRangedAttackDamageType() units.Damage
    GetRangedAttackPower() int
    GetMeleeAttackPower() int
    GetMaxHealth() int
    GetEnchantments() []data.UnitEnchantment
    GetCount() int
    GetHealth() int
    GetToHitMelee() int
    GetRangedAttacks() int
    GetCombatLbxFile() string
    GetCombatIndex(units.Facing) int
    GetCombatRangeIndex(units.Facing) int
    GetMovementSound() units.MovementSound
    GetRangeAttackSound() units.RangeAttackSound
    GetAttackSound() units.AttackSound
    GetName() string
    GetMovementSpeed() int
    IsFlying() bool
}

type ArmyUnit struct {
    Unit CombatUnit
    Facing units.Facing
    Moving bool
    X int
    Y int
    // Health int
    MovesLeft fraction.Fraction

    Team Team

    RangedAttacks int

    Attacking bool
    Defending bool

    MovementTick uint64
    MoveX float64
    MoveY float64

    LastTurn int

    // ugly to need this, but this caches paths computed for the unit
    Paths map[image.Point]pathfinding.Path
}

func (unit *ArmyUnit) ComputeDefense(damage units.Damage) int {
    defense := 0

    toDefend := unit.ToDefend()

    switch damage {
        case units.DamageRangedMagical:
            if unit.Unit.HasAbility(units.AbilityMagicImmunity) {
                toDefend = 50
            }
        case units.DamageRangedPhysical:
            if unit.Unit.HasAbility(units.AbilityMissileImmunity) {
                toDefend = 50
            }
    }

    for figure := 0; figure < unit.Figures(); figure++ {
        for i := 0; i < unit.Unit.GetDefense(); i++ {
            if rand.N(100) < toDefend {
                defense += 1
            }
        }
    }

    return defense
}

func (unit *ArmyUnit) TakeDamage(damage int) {
    // the first figure should take damage, and if it dies then the next unit takes damage, etc
    unit.Unit.AdjustHealth(-damage)
}

func (unit *ArmyUnit) Heal(amount int){
    unit.Unit.AdjustHealth(amount)
}

// given the distance to the target in tiles, return the amount of range damage done
func (unit *ArmyUnit) ComputeRangeDamage(tileDistance int) int {

    toHit := unit.Unit.GetToHitMelee()

    // magical attacks don't suffer a to-hit penalty
    if unit.Unit.GetRangedAttackDamageType() != units.DamageRangedMagical {

        if unit.Unit.HasAbility(units.AbilityLongRange) {
            if tileDistance >= 3 {
                toHit -= 10
            }
        } else {
            if tileDistance >= 3 && tileDistance <= 5 {
                toHit -= 10
            } else if tileDistance >= 6 && tileDistance <= 8 {
                toHit -= 20
            } else if tileDistance > 8 {
                toHit = 10
            }
        }

    }

    damage := 0
    for figure := 0; figure < unit.Figures(); figure++ {
        for i := 0; i < unit.Unit.GetRangedAttackPower(); i++ {
            if rand.N(100) < toHit {
                damage += 1
            }
        }
    }

    return damage
}

func (unit *ArmyUnit) ComputeMeleeDamage() int {
    damage := 0
    for figure := 0; figure < unit.Figures(); figure++ {
        for i := 0; i < unit.Unit.GetMeleeAttackPower(); i++ {
            if rand.N(100) < unit.Unit.GetToHitMelee() {
                damage += 1
            }
        }
    }

    return damage
}

func (unit *ArmyUnit) ToDefend() int {
    return 30
}

// number of alive figures in this unit
func (unit *ArmyUnit) Figures() int {

    // health per figure = max health / figures
    // figures = health / health per figure

    health_per_figure := float64(unit.Unit.GetMaxHealth()) / float64(unit.Unit.GetCount())
    return int(math.Ceil(float64(unit.Unit.GetHealth()) / health_per_figure))
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

/* compute the distance between two tiles by moving in one of the 8 directions
 */
func computeTileDistance(x1 int, y1 int, x2 int, y2 int) int {
    distance := 0

    for x1 != x2 || y1 != y2 {
        xDiff := int(math.Abs(float64(x1 - x2)))
        yDiff := int(math.Abs(float64(y1 - y2)))
        if xDiff > 0 && yDiff > 0 {
            distance += 1
            if x1 < x2 {
                x1 += 1
            } else {
                x1 -= 1
            }
            if y1 < y2 {
                y1 += 1
            } else {
                y1 -= 1
            }
        } else if xDiff > 0 && yDiff == 0 {
            distance += 1
            if x1 < x2 {
                x1 += 1
            } else {
                x1 -= 1
            }
        } else if yDiff > 0 && xDiff == 0 {
            distance += 1
            if y1 < y2 {
                y1 += 1
            } else {
                y1 -= 1
            }
        }
    }

    return distance
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

    tileEmpty := func (x int, y int) bool {
        return combat.GetUnit(x, y) == nil
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

type Army struct {
    Player *player.Player
    Units []*ArmyUnit
    Auto bool
}

func (army *Army) IsAI() bool {
    return army.Auto || army.Player.IsAI()
}

/* must call LayoutUnits() some time after invoking AddUnit() to ensure
 * the units are laid out correctly
 */
func (army *Army) AddUnit(unit CombatUnit){
    army.Units = append(army.Units, &ArmyUnit{
        Unit: unit,
        Facing: units.FacingDownRight,
        // Health: unit.GetMaxHealth(),
    })
}

func (army *Army) LayoutUnits(team Team){
    x := 10
    y := 10

    facing := units.FacingDownRight

    if team == TeamAttacker {
        x = 10
        y = 17
        facing = units.FacingUpLeft
    }

    cx := x
    cy := y

    row := 0
    for _, unit := range army.Units {
        unit.X = cx
        unit.Y = cy
        unit.Facing = facing

        cx += 1
        row += 1
        if row >= 5 {
            row = 0
            cx = x
            cy += 1
        }
    }
}

func (army *Army) RemoveUnit(remove *ArmyUnit){
    var units []*ArmyUnit

    for _, unit := range army.Units {
        if remove != unit {
            units = append(units, unit)
        }
    }

    army.Units = units
}

// represents a unit that is not part of the army, for things like magic vortex, for things like magic vortex
type OtherUnit struct {
    Animation *util.Animation
    X int
    Y int
}

type ProjectileEffect func(*ArmyUnit)

type Projectile struct {
    Target *ArmyUnit
    Animation *util.Animation
    Explode *util.Animation
    Effect ProjectileEffect
    X float64
    Y float64
    Speed float64
    Angle float64
    TargetX float64
    TargetY float64
    Exploding bool
}

type CombatScreen struct {
    ImageCache util.ImageCache
    Events chan CombatEvent
    Cache *lbx.LbxCache
    SelectedUnit *ArmyUnit
    Mouse *mouse.MouseData
    DefendingArmy *Army
    AttackingArmy *Army
    AttackingWizardFont *font.Font
    DefendingWizardFont *font.Font
    WhitePixel *ebiten.Image
    UI *uilib.UI
    DebugFont *font.Font
    HudFont *font.Font
    InfoFont *font.Font
    WhiteFont *font.Font
    // when the user hovers over a unit, that unit should be shown in a little info box at the upper right
    HighlightedUnit *ArmyUnit
    Tiles [][]Tile
    DrawRoad bool
    OtherUnits []*OtherUnit
    Projectiles []*Projectile
    // order to draw tiles in such that they are drawn from the top of the screen to the bottom (painter's order)
    TopDownOrder []image.Point

    Coordinates ebiten.GeoM
    // ScreenToTile ebiten.GeoM
    MouseState MouseState

    CameraScale float64

    Counter uint64

    TurnAttacker int
    TurnDefender int


    // track how many units were killed on each side, so experience
    // can be given out after combat ends
    DefeatedDefenders int
    DefeatedAttackers int

    Turn Team
    CurrentTurn int

    MouseTileX int
    MouseTileY int

    // if true then the player should select a tile to cast a spell on
    /*
    SelectTile func(int, int)
    */
    DoSelectTile bool

    // if true then the player should select a unit to cast a spell on
    DoSelectUnit bool
    // which team to pick a unit from
    // SelectTeam Team
    // invoke this function on the unit that is selected
    /*
    SelectTarget func(*ArmyUnit)
    CanTarget func(*ArmyUnit) bool
    */
}

type CombatLandscape int

const (
    CombatLandscapeGrass CombatLandscape = iota
    CombatLandscapeDesert
    CombatLandscapeMountain
    CombatLandscapeTundra
)

const TownCenterX = 11
const TownCenterY = 9

func makeTiles(width int, height int, landscape CombatLandscape, plane data.Plane, zone ZoneType) [][]Tile {

    baseLbx := "cmbgrass.lbx"

    switch landscape {
        case CombatLandscapeGrass:
            if plane == data.PlaneArcanus {
                baseLbx = "cmbgrass.lbx"
            } else {
                baseLbx = "cmbgrasc.lbx"
            }
        case CombatLandscapeDesert:
            if plane == data.PlaneArcanus {
                baseLbx = "cmbdesrt.lbx"
            } else {
                baseLbx = "cmbdesrc.lbx"
            }
        case CombatLandscapeMountain:
            if plane == data.PlaneArcanus {
                baseLbx = "cmbmount.lbx"
            } else {
                baseLbx = "cmbmounc.lbx"
            }
        case CombatLandscapeTundra:
            if plane == data.PlaneArcanus {
                baseLbx = "cmbtundr.lbx"
            } else {
                baseLbx = "cmbtundc.lbx"
            }
    }

    maybeExtraTile := func() TileTop {
        if rand.N(10) == 0 {
            // trees/rocks
            return TileTop{
                Lbx: baseLbx,
                Index: 48 + rand.N(10),
                Alignment: TileAlignMiddle,
            }
        }
        return TileTop{Index: -1}
    }

    tiles := make([][]Tile, height)
    for y := 0; y < len(tiles); y++ {
        tiles[y] = make([]Tile, width)
        for x := 0; x < len(tiles[y]); x++ {
            tiles[y][x] = Tile{
                // Index: rand.N(48),
                Lbx: baseLbx,
                Index: rand.N(32),
                ExtraObject: maybeExtraTile(),
            }
        }
    }

    // defending city, so place city tiles around
    if zone.City != nil {

        // clear all space around the city
        for x := -2; x <= 2; x++ {
            for y := -2; y <= 2; y++ {
                mx := x + TownCenterX
                my := y + TownCenterY
                tiles[my][mx].ExtraObject.Index = -1
            }
        }

        for range 8 {
            x := TownCenterX + rand.N(5) - 2
            y := TownCenterY + rand.N(5) - 2

            tiles[y][x].ExtraObject = TileTop{
                Lbx: "cmbtcity.lbx",
                Index: 2 + rand.N(5),
                Alignment: TileAlignBottom,
            }
        }

        if zone.City.HasFortress() {
            tiles[TownCenterY][TownCenterX].ExtraObject = TileTop{
                Lbx: "cmbtcity.lbx",
                Index: 17,
                Alignment: TileAlignBottom,
            }
        }
    } else if zone.Tower {
        tiles[TownCenterY][TownCenterX].ExtraObject = TileTop{
            Lbx: "cmbtcity.lbx",
            Index: 20,
            Alignment: TileAlignBottom,
        }
    } else if zone.AbandonedKeep {
        tiles[TownCenterY][TownCenterX].ExtraObject = TileTop{
            Lbx: "cmbtcity.lbx",
            Index: 22,
            Alignment: TileAlignBottom,
        }
    } else if zone.AncientTemple {
        tiles[TownCenterY][TownCenterX].ExtraObject = TileTop{
            Lbx: "cmbtcity.lbx",
            Index: 23,
            Alignment: TileAlignBottom,
        }
    } else if zone.FallenTemple {
        tiles[TownCenterY][TownCenterX].ExtraObject = TileTop{
            Lbx: "cmbtcity.lbx",
            // FIXME: check on this
            Index: 21,
            Alignment: TileAlignBottom,
        }
    } else if zone.Lair {
        tiles[TownCenterY][TownCenterX].ExtraObject = TileTop{
            Lbx: "cmbtcity.lbx",
            Index: 19,
            Alignment: TileAlignBottom,
        }
    } else if zone.Ruins || zone.Dungeon {
        tiles[TownCenterY][TownCenterX].ExtraObject = TileTop{
            Lbx: "cmbtcity.lbx",
            Index: 21,
            Alignment: TileAlignBottom,
        }
    } else if zone.NatureNode {
        tiles[TownCenterY][TownCenterX].ExtraObject = TileTop{
            Lbx: "cmbtcity.lbx",
            Index: 65,
            Alignment: TileAlignBottom,
        }
    } else if zone.SorceryNode {
        tiles[TownCenterY][TownCenterX].ExtraObject = TileTop{
            Lbx: "cmbtcity.lbx",
            Index: 66,
            Alignment: TileAlignBottom,
        }
    } else if zone.ChaosNode {
        tiles[TownCenterY-1][TownCenterX].ExtraObject = TileTop{
            Drawer: func(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64) {
                base, _ := imageCache.GetImageTransform("chriver.lbx", 32, 0, "crop", util.AutoCrop)
                screen.DrawImage(base, options)

                top, _ := imageCache.GetImage("chriver.lbx", 24 + int((counter / 4) % 8), 0)
                options.GeoM.Translate(16, -3)
                screen.DrawImage(top, options)

            },
        }
    }

    return tiles
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
        util.Lighten(topColor, 25), util.Lighten(topColor, 18), util.Lighten(topColor, 12),
        topColor, topColor, topColor,
        topColor, topColor, topColor,
    }
}

// player is always the human player
func MakeCombatScreen(cache *lbx.LbxCache, defendingArmy *Army, attackingArmy *Army, player *playerlib.Player, landscape CombatLandscape, plane data.Plane, zone ZoneType) *CombatScreen {
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

    whitePixel := ebiten.NewImage(1, 1)
    whitePixel.Fill(color.RGBA{R: 255, G: 255, B: 255, A: 255})

    mouseData, err := mouse.MakeMouseData(cache)
    if err != nil {
        log.Printf("Error loading mouse data: %v", err)
        return nil
    }

    // FIXME: do layout of armys
    var coordinates ebiten.GeoM

    tile0, _ := imageCache.GetImage("cmbgrass.lbx", 0, 0)

    // the battlefield is rotated by 45 degrees
    coordinates.Rotate(-math.Pi / 4)
    // coordinates.Scale(float64(tile0.Bounds().Dx())/2, float64(tile0.Bounds().Dy())/2)
    // FIXME: this math is hacky, but it works for now
    coordinates.Scale(float64(tile0.Bounds().Dx()) * 3 / 4 - 2, float64(tile0.Bounds().Dy()) * 3 / 4 - 1)
    coordinates.Translate(-220, 80)

    combat := &CombatScreen{
        Cache: cache,
        ImageCache: imageCache,
        Mouse: mouseData,
        Turn: TeamDefender,
        CameraScale: 1,
        CurrentTurn: 0,
        DefendingArmy: defendingArmy,
        TurnDefender: 0,
        AttackingArmy: attackingArmy,
        TurnAttacker: 0,
        Events: make(chan CombatEvent, 1000),
        Tiles: makeTiles(30, 30, landscape, plane, zone),
        DrawRoad: zone.City != nil,
        SelectedUnit: nil,
        DebugFont: debugFont,
        HudFont: hudFont,
        InfoFont: infoFont,
        WhiteFont: whiteFont,
        Coordinates: coordinates,
        // ScreenToTile: screenToTile,
        WhitePixel: whitePixel,
        AttackingWizardFont: attackingWizardFont,
        DefendingWizardFont: defendingWizardFont,
    }

    for _, unit := range defendingArmy.Units {
        unit.Team = TeamDefender
        unit.RangedAttacks = unit.Unit.GetRangedAttacks()
        combat.Tiles[unit.Y][unit.X].Unit = unit
    }

    for _, unit := range attackingArmy.Units {
        unit.Team = TeamAttacker
        unit.RangedAttacks = unit.Unit.GetRangedAttacks()
        combat.Tiles[unit.Y][unit.X].Unit = unit
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

func (combat *CombatScreen) GetCameraMatrix() ebiten.GeoM {
    return combat.Coordinates
}

func (combat *CombatScreen) ScreenToTile(x float64, y float64) (float64, float64) {
    // tile0, _ := combat.ImageCache.GetImage("cmbgrass.lbx", 0, 0)
    screenToTile := combat.GetCameraMatrix()
    screenToTile.Invert()

    // return screenToTile.Apply(x - float64(tile0.Bounds().Dx()/3) * combat.CameraScale, y - float64(tile0.Bounds().Dy()/3) * combat.CameraScale)
    return screenToTile.Apply(x, y)
}

func (combat *CombatScreen) computeTopDownOrder() []image.Point {
    var points []image.Point
    for y := 0; y < len(combat.Tiles); y++ {
        for x := 0; x < len(combat.Tiles[y]); x++ {
            points = append(points, image.Pt(x, y))
        }
    }

    matrix := combat.GetCameraMatrix()

    compare := func(a image.Point, b image.Point) int {
        ax, ay := matrix.Apply(float64(a.X), float64(a.Y))
        bx, by := matrix.Apply(float64(b.X), float64(b.Y))

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
func (combat *CombatScreen) createSkyProjectile(target *ArmyUnit, images []*ebiten.Image, explodeImages []*ebiten.Image, effect ProjectileEffect) *Projectile {
    // find where on the screen the unit is
    matrix := combat.GetCameraMatrix()
    screenX, screenY := matrix.Apply(float64(target.X), float64(target.Y))
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
        Target: target,
        TargetX: screenX,
        TargetY: screenY,
        Animation: util.MakeAnimation(images, true),
        Explode: util.MakeAnimation(explodeImages, false),
        Effect: effect,
    }

    return projectile
}

/* a projectile that shoots down from the sky vertically
 */
func (combat *CombatScreen) createVerticalSkyProjectile(target *ArmyUnit, images []*ebiten.Image, explodeImages []*ebiten.Image) *Projectile {
    // find where on the screen the unit is
    matrix := combat.GetCameraMatrix()
    screenX, screenY := matrix.Apply(float64(target.X), float64(target.Y))
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
        Target: target,
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
func (combat *CombatScreen) createUnitProjectile(target *ArmyUnit, images []*ebiten.Image, explodeImages []*ebiten.Image, position UnitPosition, effect ProjectileEffect) *Projectile {
    // find where on the screen the unit is
    matrix := combat.GetCameraMatrix()

    var geom1 ebiten.GeoM

    var useImage *ebiten.Image
    if len(images) > 0 {
        useImage = images[0]
    } else if len(explodeImages) > 0 {
        useImage = explodeImages[0]
    }

    switch position {
        case UnitPositionMiddle:
            // geom1.Translate(14, 3)
            geom1.Translate(-float64(useImage.Bounds().Dx()/2), -float64(useImage.Bounds().Dy()/2))
        case UnitPositionUnder:
            geom1.Translate(0, 9)
            geom1.Translate(-float64(useImage.Bounds().Dx()/2), -float64(useImage.Bounds().Dy()))
    }

    geom1.Scale(combat.CameraScale, combat.CameraScale)
    tx, ty := matrix.Apply(float64(target.X), float64(target.Y))
    geom1.Translate(tx, ty)

    screenX, screenY := geom1.Apply(0, 0)

    // log.Printf("Create fireball projectile at %v,%v -> %v,%v", x, y, screenX, screenY)

    projectile := &Projectile{
        X: screenX,
        Y: screenY,
        Target: target,
        Speed: 0,
        Angle: 0,
        Effect: effect,
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

    // FIXME: made up
    damage := func(unit *ArmyUnit) {
        unit.TakeDamage(3)
        if unit.Unit.GetHealth() <= 0 {
            combat.RemoveUnit(unit)
        }
    }

    combat.Projectiles = append(combat.Projectiles, combat.createSkyProjectile(target, loopImages, explodeImages, damage))
}

func (combat *CombatScreen) CreateFireBoltProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 0)
    loopImages := images[0:3]
    explodeImages := images[3:]

    // FIXME: made up
    damage := func(unit *ArmyUnit) {
        unit.TakeDamage(3)
        if unit.Unit.GetHealth() <= 0 {
            combat.RemoveUnit(unit)
        }
    }

    combat.Projectiles = append(combat.Projectiles, combat.createSkyProjectile(target, loopImages, explodeImages, damage))
}

func (combat *CombatScreen) CreateFireballProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 23)

    loopImages := images[0:11]
    explodeImages := images[11:]

    // FIXME: made up
    damage := func(unit *ArmyUnit) {
        unit.TakeDamage(3)
        if unit.Unit.GetHealth() <= 0 {
            combat.RemoveUnit(unit)
        }
    }

    combat.Projectiles = append(combat.Projectiles, combat.createSkyProjectile(target, loopImages, explodeImages, damage))
}

func (combat *CombatScreen) CreateStarFiresProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 9)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle, func (*ArmyUnit){}))
}

func (combat *CombatScreen) CreateDispelEvilProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 10)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle, func (*ArmyUnit){}))
}

func (combat *CombatScreen) CreatePsionicBlastProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 16)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle, func (*ArmyUnit){}))
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

    matrix := combat.GetCameraMatrix()
    screenX, screenY := matrix.Apply(float64(target.X), float64(target.Y))
    screenY += 3
    screenX += 5

    screenY -= float64(images[0].Bounds().Dy())

    projectile := &Projectile{
        X: screenX,
        Y: screenY,
        Target: target,
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

    matrix := combat.GetCameraMatrix()
    screenX, screenY := matrix.Apply(float64(target.X), float64(target.Y))
    screenY += 13
    screenX += 3

    screenY -= float64(images[0].Bounds().Dy())

    projectile := &Projectile{
        X: screenX,
        Y: screenY,
        Target: target,
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

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle, func (*ArmyUnit){}))
}

func (combat *CombatScreen) CreateFlameStrikeProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 33)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle, func (*ArmyUnit){}))
}

func (combat *CombatScreen) CreateRecallHeroProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 5)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle, func (*ArmyUnit){}))
}

func (combat *CombatScreen) CreateHealingProjectile(target *ArmyUnit) {
    // FIXME: the images should be mostly with with transparency
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 3)
    var loopImages []*ebiten.Image
    explodeImages := images

    heal := func (unit *ArmyUnit){
        unit.Heal(5)
    }

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle, heal))
}

func (combat *CombatScreen) CreateHolyWordProjectile(target *ArmyUnit) {
    // FIXME: the images should be mostly with with transparency
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 3)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle, func (*ArmyUnit){}))
}

func (combat *CombatScreen) CreateWebProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 13)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle, func (*ArmyUnit){}))
}

func (combat *CombatScreen) CreateDeathSpellProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 14)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle, func (*ArmyUnit){}))
}

func (combat *CombatScreen) CreateWordOfDeathProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 14)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle, func (*ArmyUnit){}))
}

func (combat *CombatScreen) CreateWarpWoodProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 2)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle, func (*ArmyUnit){}))
}

func (combat *CombatScreen) CreateDisintegrateProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 4)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle, func (*ArmyUnit){}))
}

func (combat *CombatScreen) CreateWordOfRecallProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 1)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle, func (*ArmyUnit){}))
}

func (combat *CombatScreen) CreateDispelMagicProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 26)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionMiddle, func (*ArmyUnit){}))
}

func (combat *CombatScreen) CreateCracksCallProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 15)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionUnder, func (*ArmyUnit){}))
}

func (combat *CombatScreen) CreateBanishProjectile(target *ArmyUnit) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 19)
    var loopImages []*ebiten.Image
    explodeImages := images

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(target, loopImages, explodeImages, UnitPositionUnder, func (*ArmyUnit){}))
}

func (combat *CombatScreen) CreateDisruptProjectile(x int, y int) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 1)

    var loopImages []*ebiten.Image
    explodeImages := images

    fakeTarget := ArmyUnit{
        X: x,
        Y: y,
    }

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(&fakeTarget, loopImages, explodeImages, UnitPositionUnder, func (*ArmyUnit){}))
}

func (combat *CombatScreen) CreateSummoningCircle(x int, y int) {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 22)
    var loopImages []*ebiten.Image
    explodeImages := images

    fakeTarget := ArmyUnit{
        X: x,
        Y: y,
    }

    combat.Projectiles = append(combat.Projectiles, combat.createUnitProjectile(&fakeTarget, loopImages, explodeImages, UnitPositionUnder, func (*ArmyUnit){}))
}

func (combat *CombatScreen) CreateMagicVortex(x int, y int) {
    images, _ := combat.ImageCache.GetImages("cmbmagic.lbx", 120)

    unit := &OtherUnit{
        X: x,
        Y: y,
        Animation: util.MakeAnimation(images, true),
    }

    combat.OtherUnits = append(combat.OtherUnits, unit)
}

func (combat *CombatScreen) addNewUnit(player *playerlib.Player, x int, y int, unit units.Unit, facing units.Facing) {
    newUnit := ArmyUnit{
        Unit: &units.OverworldUnit{
            Unit: unit,
            Health: unit.GetMaxHealth(),
        },
        Facing: facing,
        Moving: false,
        X: x,
        Y: y,
        MovesLeft: fraction.FromInt(unit.MovementSpeed),
        LastTurn: combat.CurrentTurn-1,
    }

    combat.Tiles[y][x].Unit = &newUnit

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

    selecter := TeamAttacker
    if player == combat.DefendingArmy.Player {
        selecter = TeamDefender
    }

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

    event := &CombatEventSelectUnit{
        Selecter: selecter,
        Spell: spell,
        SelectTeam: teamAttacked,
        CanTarget: canTarget,
        SelectTarget: func(target *ArmyUnit){
            sound, err := audio.LoadSound(combat.Cache, spell.Sound)
            if err == nil {
                sound.Play()
            } else {
                log.Printf("No such sound %v for %v: %v", spell.Sound, spell.Name, err)
            }

            onTarget(target)
        },
    }

    select {
        case combat.Events <- event:
        default:
    }
}

// FIXME: take in a canTarget function to check if the tile is legal
func (combat *CombatScreen) DoTargetTileSpell(player *playerlib.Player, spell spellbook.Spell, onTarget func(int, int)){
    // log.Printf("Create sound for spell %v: %v", spell.Name, spell.Sound)

    selecter := TeamAttacker
    if player == combat.DefendingArmy.Player {
        selecter = TeamDefender
    }

    event := &CombatEventSelectTile{
        Selecter: selecter,
        Spell: spell,
        SelectTile: func(x int, y int){
            sound, err := audio.LoadSound(combat.Cache, spell.Sound)
            if err == nil {
                sound.Play()
            } else {
                log.Printf("No such sound %v for %v: %v", spell.Sound, spell.Name, err)
            }

            onTarget(x, y)
        },
    }

    select {
        case combat.Events <- event:
        default:
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
        x := middleX + rand.N(distance) - distance/2
        y := middleY + rand.N(distance) - distance/2

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

            if combat.SelectedUnit != nil {

                rightImage, _ := combat.ImageCache.GetImageTransform(combat.SelectedUnit.Unit.GetCombatLbxFile(), combat.SelectedUnit.Unit.GetCombatIndex(units.FacingRight), 0, player.Wizard.Banner.String(), units.MakeUpdateUnitColorsFunc(player.Wizard.Banner))
                options.GeoM.Reset()
                options.GeoM.Translate(89, 170)
                screen.DrawImage(rightImage, &options)

                combat.HudFont.Print(screen, 92, 167, 1, ebiten.ColorScale{}, combat.SelectedUnit.Unit.GetName())

                plainAttack, _ := combat.ImageCache.GetImage("compix.lbx", 29, 0)
                options.GeoM.Reset()
                options.GeoM.Translate(126, 173)
                screen.DrawImage(plainAttack, &options)
                combat.HudFont.PrintRight(screen, 126, 174, 1, ebiten.ColorScale{}, fmt.Sprintf("%v", combat.SelectedUnit.Unit.GetMeleeAttackPower()))

                if combat.SelectedUnit.RangedAttacks > 0 {
                    y := float64(180)
                    switch combat.SelectedUnit.Unit.GetRangedAttackDamageType() {
                        case units.DamageRangedPhysical:
                            arrow, _ := combat.ImageCache.GetImage("compix.lbx", 34, 0)
                            options.GeoM.Reset()
                            options.GeoM.Translate(126, y)
                            screen.DrawImage(arrow, &options)
                            combat.HudFont.PrintRight(screen, 126, y+2, 1, ebiten.ColorScale{}, fmt.Sprintf("%v", combat.SelectedUnit.Unit.GetRangedAttackPower()))
                        case units.DamageRangedMagical:
                            magic, _ := combat.ImageCache.GetImage("compix.lbx", 30, 0)
                            options.GeoM.Reset()
                            options.GeoM.Translate(126, y)
                            screen.DrawImage(magic, &options)
                            combat.HudFont.PrintRight(screen, 126, y+2, 1, ebiten.ColorScale{}, fmt.Sprintf("%v", combat.SelectedUnit.Unit.GetRangedAttackPower()))
                    }
                }

                var movementImage *ebiten.Image
                if combat.SelectedUnit.Unit.IsFlying() {
                    movementImage, _ = combat.ImageCache.GetImage("compix.lbx", 39, 0)
                } else {
                    movementImage, _ = combat.ImageCache.GetImage("compix.lbx", 38, 0)
                }

                options.GeoM.Reset()
                options.GeoM.Translate(126, 188)
                screen.DrawImage(movementImage, &options)
                combat.HudFont.PrintRight(screen, 126, 190, 1, ebiten.ColorScale{}, fmt.Sprintf("%v", combat.SelectedUnit.MovesLeft.ToFloat()))
            }

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
                index = 1
            },
            LeftClickRelease: func(element *uilib.UIElement){
                action()
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
        spellUI := spellbook.MakeSpellBookCastUI(ui, combat.Cache, player.KnownSpells, player.ComputeCastingSkill(), spellbook.Spell{}, 0, false, func (spell spellbook.Spell, picked bool){
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
        if combat.AttackingArmy.Player == player {
            combat.AttackingArmy.Auto = true
        } else {
            combat.DefendingArmy.Auto = true
        }
    }))

    // flee
    elements = append(elements, makeButton(21, 0, 2, func(){
        // FIXME: choose the right side
        combat.AttackingArmy.Units = nil
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
    unit := combat.GetUnit(x, y)
    if unit != nil && unit.Unit.GetHealth() > 0 {
        return false
    }
    /*
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
    */

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
        unit.MovesLeft = fraction.FromInt(unit.Unit.GetMovementSpeed())
        unit.Paths = make(map[image.Point]pathfinding.Path)
    }

    for _, unit := range combat.AttackingArmy.Units {
        unit.MovesLeft = fraction.FromInt(unit.Unit.GetMovementSpeed())
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
    if x >= 0 && y >= 0 && y < len(combat.Tiles) && x < len(combat.Tiles[0]) {
        return combat.Tiles[y][x].Unit
    }

    /*
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
    */

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
    return true
}

func (combat *CombatScreen) canRangeAttack(attacker *ArmyUnit, defender *ArmyUnit) bool {
    if attacker.RangedAttacks <= 0 {
        return false
    }

    if attacker.MovesLeft.LessThanEqual(fraction.FromInt(0)) {
        return false
    }

    if attacker.Team == defender.Team {
        return false
    }

    // FIXME: check if defender has missle immunity and attacker is using regular non-magical attacks
    // FIXME: check if defender has magic immunity and attacker is using magical attacks
    // FIXME: check if defender has invisible, and attacker doesn't have illusions immunity

    return true
}

func (combat *CombatScreen) canMeleeAttack(attacker *ArmyUnit, defender *ArmyUnit) bool {
    if attacker.MovesLeft.LessThanEqual(fraction.FromInt(0)) {
        return false
    }

    if defender.Unit.IsFlying() && !attacker.Unit.IsFlying() {
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

func distanceAboveRange(x1 float64, y1 float64, x2 float64, y2 float64, r float64) bool {
    xDiff := x2 - x1
    yDiff := y2 - y1
    return xDiff * xDiff + yDiff * yDiff >= r*r
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

                if projectile.Target != nil && projectile.Effect != nil {
                    projectile.Effect(projectile.Target)
                }
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

func (combat *CombatScreen) doProjectiles(yield coroutine.YieldFunc) {
    for combat.UpdateProjectiles() {
        combat.Counter += 1
        yield()
    }
}

/* attacker is performing a physical melee attack against defender
 */
func (combat *CombatScreen) meleeAttack(attacker *ArmyUnit, defender *ArmyUnit){
    // for each figure in attacker, choose a random number from 1-100, if lower than the ToHit percent then
    // add 1 damage point. do this random roll for however much the melee attack power is

    attackerDamage := attacker.ComputeMeleeDamage()
    defenderCounterDamage := defender.ComputeMeleeDamage()

    originalAttackerDamage := attackerDamage
    originalDefenderCounterDamage := defenderCounterDamage

    defenderDefense := defender.ComputeDefense(units.DamageMeleePhysical)
    attackerDefense := attacker.ComputeDefense(units.DamageMeleePhysical)

    attackerDamage -= defenderDefense
    defenderCounterDamage -= attackerDefense

    if attackerDamage < 0 {
        attackerDamage = 0
    }

    if defenderCounterDamage < 0 {
        defenderCounterDamage = 0
    }

    log.Printf("Attacker damage roll %v, defender defended %v, attacker damage to defender %v", originalAttackerDamage, originalAttackerDamage - attackerDamage, attackerDamage)
    log.Printf("Defender counter damage roll %v, attacker defended %v, defender damage to attacker %v", originalDefenderCounterDamage, originalDefenderCounterDamage - defenderCounterDamage, defenderCounterDamage)

    attacker.TakeDamage(defenderCounterDamage)
    defender.TakeDamage(attackerDamage)

    if attacker.Unit.GetHealth() <= 0 {
        combat.RemoveUnit(attacker)
    }

    if defender.Unit.GetHealth() <= 0 {
        combat.RemoveUnit(defender)
    }
}

func (combat *CombatScreen) createUnitToUnitProjectile(attacker *ArmyUnit, target *ArmyUnit, offset image.Point, images []*ebiten.Image, explodeImages []*ebiten.Image, effect ProjectileEffect) *Projectile {
    matrix := combat.GetCameraMatrix()
    // find where on the screen the unit is
    screenX, screenY := matrix.Apply(float64(attacker.X), float64(attacker.Y))
    targetX, targetY := matrix.Apply(float64(target.X), float64(target.Y))

    var useImage *ebiten.Image
    if len(images) > 0 {
        useImage = images[0]
    } else if len(explodeImages) > 0 {
        useImage = explodeImages[0]
    }

    // FIXME: these coordinates should be incorporated into a geom

    screenY += 3
    screenY -= float64(useImage.Bounds().Dy()/2)
    screenX += 14
    screenX -= float64(useImage.Bounds().Dx()/2)

    screenY += float64(offset.Y)
    screenX += float64(offset.X)

    targetY += 3
    targetY -= float64(useImage.Bounds().Dy()/2)
    targetX += 14
    targetX -= float64(useImage.Bounds().Dx()/2)

    targetY += rand.Float64() * 6 - 3
    targetX += rand.Float64() * 6 - 3

    /*
    switch position {
        case UnitPositionMiddle:)
            screenY += 3
            screenY -= float64(useImage.Bounds().Dy()/2)
            screenX += 14
            screenX -= float64(useImage.Bounds().Dx()/2)
        case UnitPositionUnder:
            screenY += 15
            screenY -= float64(useImage.Bounds().Dy())
    }
    */

    speed := 2.8

    angle := math.Atan2(targetY - screenY, targetX - screenX)

    // log.Printf("Create fireball projectile at %v,%v -> %v,%v", x, y, screenX, screenY)

    projectile := &Projectile{
        X: screenX,
        Y: screenY,
        Target: target,
        Speed: speed,
        Angle: angle,
        Effect: effect,
        TargetX: targetX,
        TargetY: targetY,
        Animation: util.MakeAnimation(images, true),
        Explode: util.MakeAnimation(explodeImages, false),
    }

    return projectile
}

func (combat *CombatScreen) createRangeAttack(attacker *ArmyUnit, defender *ArmyUnit){
    index := attacker.Unit.GetCombatRangeIndex(attacker.Facing)
    images, err := combat.ImageCache.GetImages("cmbmagic.lbx", index)
    if err != nil {
        log.Printf("Unable to load attacker range images for %v index %v: %v", attacker.Unit.GetName(), index, err)
        return
    }

    if len(images) != 4 {
        log.Printf("Invalid number of attack range animation images for %v: %v", attacker.Unit.GetName(), len(images))
        return
    }

    animation := images[0:3]
    explode := images[3:]

    tileDistance := computeTileDistance(attacker.X, attacker.Y, defender.X, defender.Y)

    effect := func (target *ArmyUnit){
        if target.Unit.GetHealth() <= 0 {
            return
        }

        // FIXME: apply defenses for magic immunity or missle immunity

        damage := attacker.ComputeRangeDamage(tileDistance)
        defense := target.ComputeDefense(attacker.Unit.GetRangedAttackDamageType())

        // log.Printf("Ranged attack from %v: damage=%v defense=%v distance=%v", attacker.Unit.Name, damage, defense, tileDistance)

        damage -= defense
        if damage < 0 {
            damage = 0
        }
        target.TakeDamage(damage)
        if target.Unit.GetHealth() <= 0 {
            combat.RemoveUnit(target)
        }
    }

    for _, offset := range combatPoints(attacker.Figures()) {
        combat.Projectiles = append(combat.Projectiles, combat.createUnitToUnitProjectile(attacker, defender, offset, animation, explode, effect))
    }
}

func (combat *CombatScreen) RemoveUnit(unit *ArmyUnit){
    if unit.Team == TeamDefender {
        combat.DefeatedDefenders += 1
        combat.DefendingArmy.RemoveUnit(unit)
    } else {
        combat.DefeatedAttackers += 1
        combat.AttackingArmy.RemoveUnit(unit)
    }

    combat.Tiles[unit.Y][unit.X].Unit = nil

    if unit == combat.SelectedUnit {
        combat.NextUnit()
    }
}

func (combat *CombatScreen) doSelectTile(yield coroutine.YieldFunc, selecter Team, spell spellbook.Spell, selectTile func(int, int)) {
    combat.DoSelectTile = true
    defer func(){
        combat.DoSelectTile = false
    }()

    hudImage, _ := combat.ImageCache.GetImage("cmbtfx.lbx", 28, 0)

    hudY := data.ScreenHeight - hudImage.Bounds().Dy()

    quit := false

    x := 250
    if selecter == TeamDefender {
        x = 3
    }

    y := 168

    var elements []*uilib.UIElement

    removeElements := func(){
        combat.UI.RemoveElements(elements)
    }

    defer removeElements()

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
            quit = true
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(cancelRect.Min.X), float64(cancelRect.Min.Y))
            screen.DrawImage(cancelImages[cancelIndex], &options)
        },
    }

    elements = append(elements, selectElement, cancelElement)

    combat.UI.AddElements(elements)

    for !quit {
        combat.Counter += 1

        for _, unit := range combat.OtherUnits {
            if combat.Counter % 6 == 0 {
                unit.Animation.Next()
            }
        }

        combat.UI.StandardUpdate()
        mouseX, mouseY := ebiten.CursorPosition()
        tileX, tileY := combat.ScreenToTile(float64(mouseX), float64(mouseY))
        combat.MouseTileX = int(math.Round(tileX))
        combat.MouseTileY = int(math.Round(tileY))

        if mouseY >= hudY {
            combat.MouseState = CombatClickHud
        } else {
            combat.MouseState = CombatCast

            if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && mouseY < hudY {
                selectTile(combat.MouseTileX, combat.MouseTileY)
                break
            }
        }

        yield()
    }
}

func (combat *CombatScreen) doSelectUnit(yield coroutine.YieldFunc, selecter Team, spell spellbook.Spell, selectTarget func (*ArmyUnit), canTarget func (*ArmyUnit) bool, selectTeam Team) {
    combat.DoSelectUnit = true
    defer func(){
        combat.DoSelectUnit = false
    }()

    hudImage, _ := combat.ImageCache.GetImage("cmbtfx.lbx", 28, 0)

    hudY := data.ScreenHeight - hudImage.Bounds().Dy()

    x := 250
    if selecter == TeamDefender {
        x = 3
    }

    y := 168

    var elements []*uilib.UIElement

    removeElements := func(){
        combat.UI.RemoveElements(elements)
    }

    defer removeElements()

    selectElement := &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            combat.WhiteFont.PrintWrap(screen, float64(x), float64(y), 75, 1, ebiten.ColorScale{}, fmt.Sprintf("Select a target for a %v spell.", spell.Name))
        },
    }

    quit := false

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
            quit = true
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(cancelRect.Min.X), float64(cancelRect.Min.Y))
            screen.DrawImage(cancelImages[cancelIndex], &options)
        },
    }

    elements = append(elements, selectElement, cancelElement)

    combat.UI.AddElements(elements)

    for !quit {
        combat.Counter += 1

        for _, unit := range combat.OtherUnits {
            if combat.Counter % 6 == 0 {
                unit.Animation.Next()
            }
        }

        combat.UI.StandardUpdate()
        mouseX, mouseY := ebiten.CursorPosition()
        tileX, tileY := combat.ScreenToTile(float64(mouseX), float64(mouseY))
        combat.MouseTileX = int(math.Round(tileX))
        combat.MouseTileY = int(math.Round(tileY))

        combat.MouseState = CombatCast

        if mouseY >= hudY {
            combat.MouseState = CombatClickHud
        } else {
            unit := combat.GetUnit(combat.MouseTileX, combat.MouseTileY)
            if unit == nil || (selectTeam != TeamEither && unit.Team != selectTeam) || !canTarget(unit){
                combat.MouseState = CombatNotOk
            }

            if canTarget(unit) && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && mouseY < hudY {
                // log.Printf("Click unit at %v,%v -> %v", combat.MouseTileX, combat.MouseTileY, unit)
                if unit != nil && (selectTeam == TeamEither || unit.Team == selectTeam) {
                    selectTarget(unit)

                    // shouldn't need to set the mouse state here
                    combat.MouseState = CombatClickHud
                    return
                }
            }
        }

        yield()
    }
}

func (combat *CombatScreen) ProcessEvents(yield coroutine.YieldFunc) {
    for {
        select {
            case event := <-combat.Events:
                switch event.(type) {
                    case *CombatEventSelectTile:
                        use := event.(*CombatEventSelectTile)
                        combat.doSelectTile(yield, use.Selecter, use.Spell, use.SelectTile)
                    case *CombatEventSelectUnit:
                        use := event.(*CombatEventSelectUnit)
                        combat.doSelectUnit(yield, use.Selecter, use.Spell, use.SelectTarget, use.CanTarget, use.SelectTeam)
                }
            default:
                return
        }
    }
}

func (combat *CombatScreen) UpdateAnimations(){
    for _, unit := range combat.OtherUnits {
        if combat.Counter % 6 == 0 {
            unit.Animation.Next()
        }
    }
}

func (combat *CombatScreen) doMoveUnit(yield coroutine.YieldFunc, mover *ArmyUnit, path pathfinding.Path){
    if len(path) == 0 {
        return
    }

    mover.MovementTick = combat.Counter
    mover.Moving = true
    mover.MoveX = float64(mover.X)
    mover.MoveY = float64(mover.Y)

    quit, cancel := context.WithCancel(context.Background())
    defer cancel()

    sound, err := audio.LoadSound(combat.Cache, mover.Unit.GetMovementSound().LbxIndex())
    if err == nil {
        // keep playing movement sound in a loop until the unit stops moving
        go func(){
            // defer sound.Pause()
            for quit.Err() == nil {
                err = sound.Rewind()
                if err != nil {
                    log.Printf("Unable to rewind sound for %v: %v", mover.Unit.GetMovementSound(), err)
                }
                sound.Play()
                for sound.IsPlaying() {
                    select {
                    case <-quit.Done():
                        return
                    case <-time.After(10 * time.Millisecond):
                    }
                }
            }
        }()
    }

    for len(path) > 0 {
        combat.UpdateAnimations()
        combat.Counter += 1

        targetX, targetY := path[0].X, path[0].Y

        angle := math.Atan2(float64(targetY) - mover.MoveY, float64(targetX) - mover.MoveX)

        // rotate by 45 degrees to get the on screen facing angle
        // have to negate the angle because the y axis is flipped (higher y values are lower on the screen)
        useAngle := -(angle - math.Pi/4)

        // log.Printf("Angle: %v from (%v,%v) to (%v,%v)", useAngle, combat.SelectedUnit.X, combat.SelectedUnit.Y, combat.SelectedUnit.TargetX, combat.SelectedUnit.TargetY)

        mover.Facing = computeFacing(useAngle)

        // speed := float64(combat.Counter - combat.SelectedUnit.MovementTick) / 4
        speed := float64(0.08)
        mover.MoveX += math.Cos(angle) * speed
        mover.MoveY += math.Sin(angle) * speed

        // log.Printf("Moving %v,%v -> %v,%v", combat.SelectedUnit.X, combat.SelectedUnit.Y, combat.SelectedUnit.MoveX, combat.SelectedUnit.MoveY)

        // if math.Abs(combat.SelectedUnit.MoveX - float64(targetX)) < speed*2 && math.Abs(combat.SelectedUnit.MoveY - float64(targetY)) < 0.5 {
        if distanceInRange(mover.MoveX, mover.MoveY, float64(targetX), float64(targetY), speed * 3) ||
        // a stop gap to ensure the unit doesn't fly off the screen somehow
        distanceAboveRange(float64(mover.X), float64(mover.Y), float64(targetX), float64(targetY), 2.5) {

            // tile where the unit came from is now empty
            combat.Tiles[mover.Y][mover.X].Unit = nil

            mover.MovesLeft = mover.MovesLeft.Subtract(pathCost(image.Pt(mover.X, mover.Y), image.Pt(targetX, targetY)))
            if mover.MovesLeft.LessThan(fraction.FromInt(0)) {
                mover.MovesLeft = fraction.FromInt(0)
            }

            mover.X = targetX
            mover.Y = targetY
            mover.MoveX = float64(targetX)
            mover.MoveY = float64(targetY)
            // new tile the unit landed on is now occupied
            combat.Tiles[mover.Y][mover.X].Unit = mover
            path = path[1:]
        }

        yield()
    }

    mover.Moving = false
    mover.Paths = make(map[image.Point]pathfinding.Path)
}

func (combat *CombatScreen) doRangeAttack(yield coroutine.YieldFunc, attacker *ArmyUnit, defender *ArmyUnit){
    attacker.MovesLeft = attacker.MovesLeft.Subtract(fraction.FromInt(10))
    if attacker.MovesLeft.LessThan(fraction.FromInt(0)) {
        attacker.MovesLeft = fraction.FromInt(0)
    }

    attacker.RangedAttacks -= 1

    attacker.Facing = faceTowards(attacker.X, attacker.Y, defender.X, defender.Y)

    // FIXME: could use a for/yield loop here to update projectiles
    combat.createRangeAttack(attacker, defender)

    sound, err := audio.LoadSound(combat.Cache, attacker.Unit.GetRangeAttackSound().LbxIndex())
    if err == nil {
        sound.Play()
    }

    combat.doProjectiles(yield)
}

func (combat *CombatScreen) doMelee(yield coroutine.YieldFunc, attacker *ArmyUnit, defender *ArmyUnit){
    // create defer scope
    attacker.Attacking = true
    defender.Defending = true
    defer func(){
        attacker.Attacking = false
        defender.Defending = false
    }()

    // attacking takes 50% of movement points
    // FIXME: in some cases an extra 0.5 movements points is lost, possibly due to counter attacks?

    pointsUsed := fraction.FromInt(attacker.Unit.GetMovementSpeed()).Divide(fraction.FromInt(2))
    if pointsUsed.LessThan(fraction.FromInt(1)) {
        pointsUsed = fraction.FromInt(1)
    }

    attacker.MovesLeft = attacker.MovesLeft.Subtract(pointsUsed)
    if attacker.MovesLeft.LessThan(fraction.FromInt(0)) {
        attacker.MovesLeft = fraction.FromInt(0)
    }

    attacker.Facing = faceTowards(attacker.X, attacker.Y, defender.X, defender.Y)
    defender.Facing = faceTowards(defender.X, defender.Y, attacker.X, attacker.Y)

    // FIXME: sound is based on attacker type, and possibly defender type
    sound, err := audio.LoadCombatSound(combat.Cache, attacker.Unit.GetAttackSound().LbxIndex())
    if err == nil {
        sound.Play()
    }

    for i := 0; i < 60; i++ {
        combat.Counter += 1
        combat.UpdateAnimations()

        if i == 20 {
            combat.meleeAttack(combat.SelectedUnit, defender)
        }

        yield()
    }
}

func (combat *CombatScreen) IsAIControlled(unit *ArmyUnit) bool {
    if unit.Team == TeamDefender {
        return combat.DefendingArmy.IsAI()
    } else {
        return combat.AttackingArmy.IsAI()
    }
}

func (combat *CombatScreen) GetArmy(unit *ArmyUnit) *Army {
    if unit.Team == TeamDefender {
        return combat.DefendingArmy
    }

    return combat.AttackingArmy
}

func (combat *CombatScreen) GetOtherArmy(unit *ArmyUnit) *Army {
    if unit.Team == TeamDefender {
        return combat.AttackingArmy
    }

    return combat.DefendingArmy
}

func (combat *CombatScreen) doAI(yield coroutine.YieldFunc, aiUnit *ArmyUnit) {
    // aiArmy := combat.GetArmy(combat.SelectedUnit)
    otherArmy := combat.GetOtherArmy(aiUnit)

    // try a ranged attack first
    if aiUnit.RangedAttacks > 0 {
        candidates := slices.Clone(otherArmy.Units)
        slices.SortFunc(candidates, func (a *ArmyUnit, b *ArmyUnit) int {
            return cmp.Compare(computeTileDistance(aiUnit.X, aiUnit.Y, a.X, a.Y), computeTileDistance(aiUnit.X, aiUnit.Y, b.X, b.Y))
        })

        for _, candidate := range candidates {
           if combat.withinArrowRange(aiUnit, candidate) && combat.canRangeAttack(aiUnit, candidate) {
               combat.doRangeAttack(yield, aiUnit, candidate)
               return
           }
        }
    }

    aiUnit.Paths = make(map[image.Point]pathfinding.Path)

    // if the selected unit has ranged attacks, then try to use that
    // otherwise, if in melee range of some enemy then attack them
    // otherwise walk towards the enemy

    paths := make(map[*ArmyUnit]pathfinding.Path)

    getPath := func (unit *ArmyUnit) pathfinding.Path {
        path, found := paths[unit]
        if !found {
            combat.Tiles[unit.Y][unit.X].Unit = nil
            var ok bool
            path, ok = combat.computePath(aiUnit.X, aiUnit.Y, unit.X, unit.Y)
            combat.Tiles[unit.Y][unit.X].Unit = unit
            if ok {
                paths[unit] = path
            } else {
                paths[unit] = nil
            }
        }

        return path
    }

    filterReachable := func (units []*ArmyUnit) []*ArmyUnit {
        var out []*ArmyUnit
        for _, unit := range units {
            path := getPath(unit)
            if len(path) > 0 {
                out = append(out, unit)
            }
        }
        return out
    }

    // should filter by enemies that we can attack, so non-flyers do not move toward flyers
    candidates := filterReachable(slices.Clone(otherArmy.Units))

    slices.SortFunc(candidates, func (a *ArmyUnit, b *ArmyUnit) int {
        aPath := getPath(a)
        bPath := getPath(b)

        return cmp.Compare(len(aPath), len(bPath))
    })


    // find a path to some enemy
    for _, closestEnemy := range candidates {
        // pretend that there is no unit at the tile. this is a sin of the highest order

        path := getPath(closestEnemy)

        // a path of length 2 contains the position of the aiUnit and the position of the enemy, so they are right next to each other
        if len(path) == 2 && combat.canMeleeAttack(aiUnit, closestEnemy) {
            combat.doMelee(yield, aiUnit, closestEnemy)
            return
        } else if len(path) > 2 {
            // ignore path[0], thats where we are now. also ignore the last element, since we can't move onto the enemy

            last := path[len(path)-1]
            if last.X == closestEnemy.X && last.Y == closestEnemy.Y {
                path = path[:len(path)-1]
            }

            lastIndex := 0
            for lastIndex < len(path) {
                if aiUnit.CanFollowPath(path[0:lastIndex]) {
                    lastIndex += 1
                } else {
                    lastIndex -= 1
                    break
                }
            }

            if lastIndex >= 1 && lastIndex <= len(path) {
                combat.doMoveUnit(yield, aiUnit, path[1:lastIndex])
                return
            }
        }
    }

    // didn't make a choice, just exhaust moves left
    aiUnit.MovesLeft = fraction.FromInt(0)
}

func (combat *CombatScreen) UpdateMouseState() {
    switch combat.MouseState {
        case CombatMoveOk:
            globalMouse.Mouse.SetImage(combat.Mouse.Move)
        case CombatClickHud:
            globalMouse.Mouse.SetImage(combat.Mouse.Normal)
        case CombatMeleeAttackOk:
            // mouseOptions.GeoM.Translate(-1, -1)
            globalMouse.Mouse.SetImage(combat.Mouse.Attack)
        case CombatRangeAttackOk:
            globalMouse.Mouse.SetImage(combat.Mouse.Arrow)
        case CombatNotOk:
            globalMouse.Mouse.SetImage(combat.Mouse.Error)
        case CombatCast:
            index := (combat.Counter / 8) % uint64(len(combat.Mouse.Cast))
            globalMouse.Mouse.SetImage(combat.Mouse.Cast[index])
    }
}

func (combat *CombatScreen) Update(yield coroutine.YieldFunc) CombatState {
    // defender wins in a tie
    if len(combat.AttackingArmy.Units) == 0 {
        return CombatStateDefenderWin
    }

    if len(combat.DefendingArmy.Units) == 0 {
        return CombatStateAttackerWin
    }

    combat.Counter += 1
    combat.UI.StandardUpdate()

    combat.UpdateMouseState()

    mouseX, mouseY := ebiten.CursorPosition()
    hudImage, _ := combat.ImageCache.GetImage("cmbtfx.lbx", 28, 0)

    tileX, tileY := combat.ScreenToTile(float64(mouseX), float64(mouseY))
    combat.MouseTileX = int(math.Round(tileX))
    combat.MouseTileY = int(math.Round(tileY))

    combat.UpdateAnimations()

    hudY := data.ScreenHeight - hudImage.Bounds().Dy()

    var keys []ebiten.Key
    keys = inpututil.AppendPressedKeys(keys)
    for _, key := range keys {
        speed := 1.5
        switch key {
            case ebiten.KeyDown:
                combat.Coordinates.Translate(0, -speed)
            case ebiten.KeyUp:
                combat.Coordinates.Translate(0, speed)
            case ebiten.KeyLeft:
                combat.Coordinates.Translate(speed, 0)
            case ebiten.KeyRight:
                combat.Coordinates.Translate(-speed, 0)
            case ebiten.KeyEqual:
                if combat.CameraScale < 3 {
                    combat.CameraScale *= 1 + 0.01
                    combat.Coordinates.Scale(1.01, 1.01)
                }
            case ebiten.KeyMinus:
                if combat.CameraScale > 0.5 {
                    combat.CameraScale *= 1.0 - 0.01
                    combat.Coordinates.Scale(0.99, 0.99)
                }
            case ebiten.KeySpace:
                normalized := 1 / combat.CameraScale
                combat.CameraScale *= normalized
                combat.Coordinates.Scale(normalized, normalized)
        }
    }

    // FIXME: handle right-click drag to move the camera

    _, wheelY := ebiten.Wheel()
    wheelScale := 1 + float64(wheelY) / 10
    combat.CameraScale *= wheelScale
    combat.Coordinates.Scale(wheelScale, wheelScale)

    combat.ProcessEvents(yield)

    if len(combat.Projectiles) > 0 {
        combat.doProjectiles(yield)
    }

    if combat.SelectedUnit != nil && combat.IsAIControlled(combat.SelectedUnit) {
        aiUnit := combat.SelectedUnit

        // keep making choices until the unit runs out of moves
        for aiUnit.MovesLeft.GreaterThan(fraction.FromInt(0)) {
            combat.doAI(yield, aiUnit)
        }
        aiUnit.LastTurn = combat.CurrentTurn
        combat.NextUnit()
        return CombatStateRunning
    }

    if combat.UI.GetHighestLayerValue() > 0 || mouseY >= hudY {
        combat.MouseState = CombatClickHud
    } else if combat.SelectedUnit != nil && combat.SelectedUnit.Moving {
        combat.MouseState = CombatClickHud
    } else if combat.SelectedUnit != nil {
        who := combat.GetUnit(combat.MouseTileX, combat.MouseTileY)
        if who == nil {
            if combat.CanMoveTo(combat.SelectedUnit, combat.MouseTileX, combat.MouseTileY) {
                combat.MouseState = CombatMoveOk
            } else {
                combat.MouseState = CombatNotOk
            }
        } else {
            newState := CombatNotOk
            // prioritize range attack over melee
            if combat.canRangeAttack(combat.SelectedUnit, who) && combat.withinArrowRange(combat.SelectedUnit, who) {
                newState = CombatRangeAttackOk
            } else if combat.canMeleeAttack(combat.SelectedUnit, who) && combat.withinMeleeRange(combat.SelectedUnit, who) {
                newState = CombatMeleeAttackOk
            }

            combat.MouseState = newState
        }
    }

    // if there is no unit at the tile position then the highlighted unit will be nil
    if combat.UI.GetHighestLayerValue() == 0 {
        combat.HighlightedUnit = combat.GetUnit(combat.MouseTileX, combat.MouseTileY)
    }

    // dont allow clicks into the hud area
    // also don't allow clicks into the game if the ui is showing some overlay
    if combat.UI.GetHighestLayerValue() == 0 &&
       inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) &&
       mouseY < hudY {

        if combat.TileIsEmpty(combat.MouseTileX, combat.MouseTileY) && combat.CanMoveTo(combat.SelectedUnit, combat.MouseTileX, combat.MouseTileY){
            path, _ := combat.FindPath(combat.SelectedUnit, combat.MouseTileX, combat.MouseTileY)
            path = path[1:]
            combat.doMoveUnit(yield, combat.SelectedUnit, path)
        } else {

           defender := combat.GetUnit(combat.MouseTileX, combat.MouseTileY)
           attacker := combat.SelectedUnit

           // try a ranged attack first
           if defender != nil && combat.withinArrowRange(attacker, defender) && combat.canRangeAttack(attacker, defender) {
               combat.doRangeAttack(yield, attacker, defender)
           // then fall back to melee
           } else if defender != nil && defender.Team != attacker.Team && combat.withinMeleeRange(attacker, defender) && combat.canMeleeAttack(attacker, defender){
               combat.doMelee(yield, attacker, defender)
               attacker.Paths = make(map[image.Point]pathfinding.Path)
           }
       }
    }

    // the unit died or is out of moves
    if combat.SelectedUnit != nil && (combat.SelectedUnit.Unit.GetHealth() <= 0 || combat.SelectedUnit.MovesLeft.LessThanEqual(fraction.FromInt(0))) {
        combat.SelectedUnit.LastTurn = combat.CurrentTurn
        combat.NextUnit()
    }

    // log.Printf("Mouse original %v,%v %v,%v -> %v,%v", mouseX, mouseY, tileX, tileY, combat.MouseTileX, combat.MouseTileY)

    return CombatStateRunning
}

func (combat *CombatScreen) DrawHighlightedTile(screen *ebiten.Image, x int, y int, matrix *ebiten.GeoM, minColor color.RGBA, maxColor color.RGBA){
    tile0, _ := combat.ImageCache.GetImage("cmbgrass.lbx", 0, 0)

    var useMatrix ebiten.GeoM

    tx, ty := matrix.Apply(float64(x), float64(y))
    useMatrix.Scale(combat.CameraScale, combat.CameraScale)
    useMatrix.Translate(tx, ty)

    // left
    x1, y1 := useMatrix.Apply(-float64(tile0.Bounds().Dx()/2), 0)
    // top
    x2, y2 := useMatrix.Apply(0, -float64(tile0.Bounds().Dy()/2))
    // right
    x3, y3 := useMatrix.Apply(float64(tile0.Bounds().Dx()/2), 0)
    // bottom
    x4, y4 := useMatrix.Apply(0, float64(tile0.Bounds().Dy()/2))

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

func (combat *CombatScreen) ShowUnitInfo(screen *ebiten.Image, unit *ArmyUnit){
    x1 := 255 - 1
    y1 := 5
    width := 65
    height := 45
    vector.DrawFilledRect(screen, float32(x1), float32(y1), float32(width), float32(height), color.RGBA{R: 0, G: 0, B: 0, A: 100}, false)
    vector.StrokeRect(screen, float32(x1), float32(y1), float32(width), float32(height), 1, util.PremultiplyAlpha(color.RGBA{R: 0x27, G: 0x4e, B: 0xdc, A: 100}), false)
    combat.InfoFont.PrintCenter(screen, float64(x1 + 35), float64(y1 + 2), 1, ebiten.ColorScale{}, fmt.Sprintf("%v", unit.Unit.GetName()))

    meleeImage, _ := combat.ImageCache.GetImage("compix.lbx", 61, 0)
    var options ebiten.DrawImageOptions
    options.GeoM.Translate(float64(x1 + 14), float64(y1 + 10))
    screen.DrawImage(meleeImage, &options)
    combat.InfoFont.PrintRight(screen, float64(x1 + 14), float64(y1 + 10 + 2), 1, ebiten.ColorScale{}, fmt.Sprintf("%v", unit.Unit.GetMeleeAttackPower()))

    switch unit.Unit.GetRangedAttackDamageType() {
        case units.DamageRangedMagical:
            fire, _ := combat.ImageCache.GetImage("compix.lbx", 62, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(x1 + 14), float64(y1 + 18))
            screen.DrawImage(fire, &options)
            combat.InfoFont.PrintRight(screen, float64(x1 + 14), float64(y1 + 18 + 2), 1, ebiten.ColorScale{}, fmt.Sprintf("%v", unit.Unit.GetRangedAttackPower()))
        case units.DamageRangedPhysical:
            arrow, _ := combat.ImageCache.GetImage("compix.lbx", 66, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(x1 + 14), float64(y1 + 18))
            screen.DrawImage(arrow, &options)
            combat.InfoFont.PrintRight(screen, float64(x1 + 14), float64(y1 + 18 + 2), 1, ebiten.ColorScale{}, fmt.Sprintf("%v", unit.Unit.GetRangedAttackPower()))
    }

    movementImage, _ := combat.ImageCache.GetImage("compix.lbx", 72, 0)
    if unit.Unit.IsFlying() {
        movementImage, _ = combat.ImageCache.GetImage("compix.lbx", 73, 0)
    }

    options.GeoM.Reset()
    options.GeoM.Translate(float64(x1 + 14), float64(y1 + 26))
    screen.DrawImage(movementImage, &options)
    combat.InfoFont.PrintRight(screen, float64(x1 + 14), float64(y1 + 26 + 2), 1, ebiten.ColorScale{}, fmt.Sprintf("%v", unit.MovesLeft.ToFloat()))

    armorImage, _ := combat.ImageCache.GetImage("compix.lbx", 70, 0)
    options.GeoM.Reset()
    options.GeoM.Translate(float64(x1 + 48), float64(y1 + 10))
    screen.DrawImage(armorImage, &options)
    combat.InfoFont.PrintRight(screen, float64(x1 + 48), float64(y1 + 10 + 2), 1, ebiten.ColorScale{}, fmt.Sprintf("%v", unit.Unit.GetDefense()))

    resistanceImage, _ := combat.ImageCache.GetImage("compix.lbx", 75, 0)
    options.GeoM.Reset()
    options.GeoM.Translate(float64(x1 + 48), float64(y1 + 18))
    screen.DrawImage(resistanceImage, &options)
    combat.InfoFont.PrintRight(screen, float64(x1 + 48), float64(y1 + 18 + 2), 1, ebiten.ColorScale{}, fmt.Sprintf("%v", unit.Unit.GetResistance()))

    combat.InfoFont.PrintCenter(screen, float64(x1 + 14), float64(y1 + 37), 1, ebiten.ColorScale{}, "Hits")

    highHealth := color.RGBA{R: 0, G: 0xff, B: 0, A: 0xff}
    mediumHealth := color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff}
    lowHealth := color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}
    healthWidth := 15

    vector.StrokeLine(screen, float32(x1 + 25), float32(y1 + 40), float32(x1 + 25 + healthWidth), float32(y1 + 40), 1, color.RGBA{R: 0, G: 0, B: 0, A: 0xff}, false)

    healthPercent := float64(unit.Unit.GetHealth()) / float64(unit.Unit.GetMaxHealth())
    healthLength := float64(healthWidth) * healthPercent

    // always show at least one point of health
    if healthLength < 1 {
        healthLength = 1
    }

    useColor := highHealth
    if healthPercent < 0.33 {
        useColor = lowHealth
    } else if healthPercent < 0.66 {
        useColor = mediumHealth
    } else {
        useColor = highHealth
    }

    vector.StrokeLine(screen, float32(x1 + 25), float32(y1 + 40), float32(x1 + 25) + float32(healthLength), float32(y1 + 40), 1, useColor, false)
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

    useMatrix := combat.GetCameraMatrix()

    tilePosition := func(x float64, y float64) (float64, float64){
        return useMatrix.Apply(x, y)
    }

    // draw base land first
    for _, point := range combat.TopDownOrder {
        x := point.X
        y := point.Y

        image, _ := combat.ImageCache.GetImage(combat.Tiles[y][x].Lbx, combat.Tiles[y][x].Index, 0)
        options.GeoM.Reset()
        // tx,ty is the middle of the tile
        tx, ty := tilePosition(float64(x), float64(y))
        options.GeoM.Scale(combat.CameraScale, combat.CameraScale)
        options.GeoM.Translate(tx, ty)
        screen.DrawImage(image, &options)

        if combat.Tiles[y][x].Mud {
            mudTiles, _ := combat.ImageCache.GetImages("cmbtcity.lbx", 118)
            index := animationIndex % uint64(len(mudTiles))
            screen.DrawImage(mudTiles[index], &options)
        }

        // vector.DrawFilledCircle(screen, float32(tx), float32(ty), 2, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, false)
    }

    if combat.DrawRoad {
        tx, ty := tilePosition(TownCenterX+1, TownCenterY-4)

        road, _ := combat.ImageCache.GetImageTransform("cmbtcity.lbx", 0, 0, "crop", util.AutoCrop)
        options.GeoM.Reset()
        options.GeoM.Scale(combat.CameraScale, combat.CameraScale)
        options.GeoM.Translate(tx, ty)
        options.GeoM.Translate(0, float64(tile0.Bounds().Dy())/2)
        screen.DrawImage(road, &options)

    }

    // then draw extra stuff on top
    for _, point := range combat.TopDownOrder {
        x := point.X
        y := point.Y

        extra := combat.Tiles[y][x].ExtraObject
        if extra.Drawer != nil {
            options.GeoM.Reset()
            tx, ty := tilePosition(float64(x), float64(y))
            options.GeoM.Scale(combat.CameraScale, combat.CameraScale)
            options.GeoM.Translate(tx, ty)

            extra.Drawer(screen, &combat.ImageCache, &options, combat.Counter)
        } else if extra.Index != -1 {
            options.GeoM.Reset()
            // tx,ty is the middle of the tile
            tx, ty := tilePosition(float64(x), float64(y))

            var geom ebiten.GeoM

            geom.Scale(combat.CameraScale, combat.CameraScale)
            geom.Translate(tx, ty)

            extraImages, _ := combat.ImageCache.GetImagesTransform(extra.Lbx, extra.Index, "crop", util.AutoCrop)

            index := animationIndex % uint64(len(extraImages))
            extraImage := extraImages[index]

            switch extra.Alignment {
                case TileAlignBottom:
                    options.GeoM.Translate(0, float64(tile0.Bounds().Dy())/2)
                    options.GeoM.Translate(-float64(extraImage.Bounds().Dy())/2, -float64(extraImage.Bounds().Dy()))
                case TileAlignMiddle:
                    options.GeoM.Translate(-float64(extraImage.Bounds().Dy())/2, -float64(extraImage.Bounds().Dy()/2))
            }

            options.GeoM.Concat(geom)

            screen.DrawImage(extraImage, &options)

            // vector.DrawFilledCircle(screen, float32(tx), float32(ty), 2, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, false)
        }
    }

    combat.DrawHighlightedTile(screen, combat.MouseTileX, combat.MouseTileY, &useMatrix, color.RGBA{R: 0, G: 0x67, B: 0x78, A: 255}, color.RGBA{R: 0, G: 0xef, B: 0xff, A: 255})

    if combat.SelectedUnit != nil {
        path, ok := combat.FindPath(combat.SelectedUnit, combat.MouseTileX, combat.MouseTileY)
        if ok {
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(0.8)
            for i := 1; i < len(path); i++ {
                tileX, tileY := path[i].X, path[i].Y

                tx, ty := tilePosition(float64(tileX), float64(tileY))
                // tx += float64(tile0.Bounds().Dx())/2
                // ty += float64(tile0.Bounds().Dy())/2
                movementImage, _ := combat.ImageCache.GetImage("compix.lbx", 72, 0)
                tx -= float64(movementImage.Bounds().Dx())/2
                ty -= float64(movementImage.Bounds().Dy())/2

                options.GeoM.Reset()
                options.GeoM.Scale(combat.CameraScale, combat.CameraScale)
                options.GeoM.Translate(tx, ty)
                screen.DrawImage(movementImage, &options)
            }
        }

        minColor := color.RGBA{R: 32, G: 0, B: 0, A: 255}
        maxColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}

        if !combat.SelectedUnit.Moving {
            combat.DrawHighlightedTile(screen, combat.SelectedUnit.X, combat.SelectedUnit.Y, &useMatrix, minColor, maxColor)
        }
    }

    renderUnit := func(unit *ArmyUnit){
        banner := unit.Unit.GetBanner()
        combatImages, _ := combat.ImageCache.GetImagesTransform(unit.Unit.GetCombatLbxFile(), unit.Unit.GetCombatIndex(unit.Facing), banner.String(), units.MakeUpdateUnitColorsFunc(banner))

        if combatImages != nil {
            var unitOptions ebiten.DrawImageOptions
            unitOptions.GeoM.Reset()
            var tx float64
            var ty float64

            if unit.Moving {
                tx, ty = tilePosition(unit.MoveX, unit.MoveY)
            } else {
                tx, ty = tilePosition(float64(unit.X), float64(unit.Y))
            }

            /*
            tx, ty := tilePosition(float64(x), float64(y))
            options.GeoM.Scale(combat.CameraScale, combat.CameraScale)
            options.GeoM.Translate(tx, ty)
            */

            unitOptions.GeoM.Scale(combat.CameraScale, combat.CameraScale)
            unitOptions.GeoM.Translate(tx, ty)
            // unitOptions.GeoM.Translate(float64(tile0.Bounds().Dx()/2) * combat.CameraScale, 0)
            // unitOptions.GeoM.Translate(float64(tile0.Bounds().Dx()/2), float64(tile0.Bounds().Dy()/2))
            // unitOptions.GeoM.Translate(0, float64(tile0.Bounds().Dy()/2))

            index := uint64(0)
            if unit.Unit.IsFlying() || unit.Moving {
                index = animationIndex % (uint64(len(combatImages)) - 1)
            }

            if unit.Attacking || unit.Defending {
                index = 2 + animationIndex % 2
            }

            if combat.SelectedUnit == unit {
                scaleValue := 1.5 + math.Sin(float64(combat.Counter)/6)/2
                unitOptions.ColorScale.Scale(float32(scaleValue), float32(scaleValue), 1, 1)
            }

            if combat.HighlightedUnit == unit {
                scaleValue := 1.5 + math.Sin(float64(combat.Counter)/6)/2
                unitOptions.ColorScale.Scale(float32(scaleValue), 1, 1, 1)
            }

            /*
            x, y := unitOptions.GeoM.Apply(0, 0)
            vector.DrawFilledCircle(screen, float32(x), float32(y), 2, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, false)
            x, y = unitOptions.GeoM.Apply(float64(tile0.Bounds().Dx()/2), 0)
            vector.DrawFilledCircle(screen, float32(x), float32(y), 2, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, false)
            */

            // _ = index
            enchantment := util.First(unit.Unit.GetEnchantments(), data.UnitEnchantmentNone)
            RenderCombatUnit(screen, combatImages[index], unitOptions, unit.Figures(), enchantment, combat.Counter, &combat.ImageCache)
        }
    }

    // sort units in top down order before drawing them
    allUnits := make([]*ArmyUnit, 0, len(combat.DefendingArmy.Units) + len(combat.AttackingArmy.Units))
    for _, unit := range combat.DefendingArmy.Units {
        allUnits = append(allUnits, unit)
    }

    for _, unit := range combat.AttackingArmy.Units {
        allUnits = append(allUnits, unit)
    }

    compareUnit := func(unitA *ArmyUnit, unitB *ArmyUnit) int {
        ax, ay := tilePosition(float64(unitA.X), float64(unitA.Y))
        bx, by := tilePosition(float64(unitB.X), float64(unitB.Y))

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

    slices.SortFunc(allUnits, compareUnit)

    for _, unit := range allUnits {
        renderUnit(unit)
    }

    for _, unit := range combat.OtherUnits {
        var unitOptions ebiten.DrawImageOptions
        tx, ty := tilePosition(float64(unit.X), float64(unit.Y))
        unitOptions.GeoM.Scale(combat.CameraScale, combat.CameraScale)
        unitOptions.GeoM.Translate(tx, ty)
        unitOptions.GeoM.Translate(float64(tile0.Bounds().Dx()/2), float64(tile0.Bounds().Dy()/2))

        frame := unit.Animation.Frame()
        unitOptions.GeoM.Translate(float64(-frame.Bounds().Dx()/2), float64(-frame.Bounds().Dy()))
        screen.DrawImage(frame, &unitOptions)
    }

    combat.UI.Draw(combat.UI, screen)

    if combat.HighlightedUnit != nil {
        combat.ShowUnitInfo(screen, combat.HighlightedUnit)
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
            options.GeoM.Scale(combat.CameraScale, combat.CameraScale)
            options.GeoM.Translate(projectile.X, projectile.Y)
            screen.DrawImage(frame, &options)
        }
    }
}
