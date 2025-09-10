package combat

import (
    "image"
    "log"
    "fmt"
    "slices"
    "math"
    "math/rand/v2"
    "time"

    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    // playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"

    "github.com/hajimehoshi/ebiten/v2"
)

const MAX_TURNS = 50
const LostUnitsMax = 50

type DamageSource int
const (
    DamageSourceNormal DamageSource = iota
    DamageSourceHero
    DamageSourceFantastic
    DamageSourceSpell
)

type DamageType int
const (
    DamageNormal DamageType = iota
    DamageIrreversable
    DamageUndead
)

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

func (zone *ZoneType) GetMagic() data.MagicType {
    if zone.ChaosNode {
        return data.ChaosMagic
    }

    if zone.NatureNode {
        return data.NatureMagic
    }

    if zone.SorceryNode {
        return data.SorceryMagic
    }

    return data.MagicNone
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

type ArmyPlayer interface {
    spellbook.SpellCaster

    GetKnownSpells() spellbook.Spells
    GetWizard() *setup.WizardCustom
    FindFortressCity() *citylib.City
    MakeExperienceInfo() units.ExperienceInfo
    MakeUnitEnchantmentProvider() units.GlobalEnchantmentProvider
    HasEnchantment(data.Enchantment) bool
    GetMana() int
    UseMana(int)
    ComputeCastingSkill() int
    IsHuman() bool
    IsAI() bool
}

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

type FireSide int
const (
    FireSideSouth FireSide = iota
    FireSideEast
    FireSideNorth
    FireSideWest
)

type DarknessSide int
const (
    DarknessSideSouth DarknessSide = iota
    DarknessSideEast
    DarknessSideNorth
    DarknessSideWest
)

type WallKind int
const (
    WallKindNone WallKind = iota
    WallKindNorth
    WallKindEast
    WallKindSouth
    WallKindWest
    WallKindGate
)

type Tile struct {
    // a unit standing on this tile, if any
    Unit *ArmyUnit
    Lbx string
    // index of grass/floor
    Index int
    // tree/rock on top, or -1 if nothing
    ExtraObject TileTop
    Mud bool
    // whether to show fire on this tile
    Fire *set.Set[FireSide]
    // the counter when the fire was last activated
    FireActive uint64
    // whether to show wall of darkness on this tile
    Darkness *set.Set[DarknessSide]
    DarknessActive uint64

    Wall *set.Set[WallKind]

    // if combat is in a city with flying fortress then this tile might be flying (in the clouds)
    Flying bool

    // true if this tile is inside the wall of fire/darkness
    InsideTown bool
    InsideFire bool
    InsideDarkness bool
    InsideWall bool
}

type CombatLandscape int

const (
    CombatLandscapeGrass CombatLandscape = iota
    CombatLandscapeWater
    CombatLandscapeDesert
    CombatLandscapeMountain
    CombatLandscapeTundra
)

const TownCenterX = 11
const TownCenterY = 10

func makeTiles(width int, height int, landscape CombatLandscape, plane data.Plane, zone ZoneType) [][]Tile {

    baseLbx := "cmbgrass.lbx"

    // most tile sets have 32 possible tiles to choose from
    tileMax := 32
    tileStart := 0

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
        case CombatLandscapeWater:
            // water only has 4
            tileMax = 4
            if plane == data.PlaneArcanus {
                baseLbx = "cmbtcity.lbx"
                tileStart = 109
            } else {
                baseLbx = "chriver.lbx"
                tileStart = 12
            }
    }

    maybeExtraTile := func() TileTop {
        // water never has trees/rocks
        if landscape == CombatLandscapeWater {
            return TileTop{Index: -1}
        }

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
                Index: tileStart + rand.N(tileMax),
                ExtraObject: maybeExtraTile(),
            }
        }
    }

    // defending city, so place city tiles around
    if zone.City != nil {

        townSquare := image.Rect(TownCenterX - 2, TownCenterY - 2, TownCenterX + 1, TownCenterY + 1)

        randTownSquare := func() (int, int) {
            x := rand.N(townSquare.Dx())
            y := rand.N(townSquare.Dy())
            return townSquare.Min.X + x, townSquare.Min.Y + y
        }

        flyingFortress := zone.City != nil && zone.City.HasEnchantment(data.CityEnchantmentFlyingFortress)

        // clear all space around the city
        for x := townSquare.Min.X; x <= townSquare.Max.X; x++ {
            for y := townSquare.Min.Y; y <= townSquare.Max.Y; y++ {
                tiles[y][x].ExtraObject.Index = -1
                tiles[y][x].InsideTown = true
                tiles[y][x].Flying = flyingFortress
            }
        }

        // add random houses
        for range 8 {
            x, y := randTownSquare()

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

        if zone.City.HasWallOfFire() {
            createWallOfFire(tiles, TownCenterX, TownCenterY, 4, 0)
        }

        if zone.City.HasWallOfDarkness() {
            createWallOfDarkness(tiles, TownCenterX, TownCenterY, 4, 0)
        }

        if zone.City.HasWall() {
            createCityWall(tiles, TownCenterX, TownCenterY, 4)
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
                scale.DrawScaled(screen, base, options)

                top, _ := imageCache.GetImage("chriver.lbx", 24 + int((counter / 4) % 8), 0)
                var geom2 ebiten.GeoM
                geom2.Translate(float64(16), float64(-3))
                geom2.Concat(options.GeoM)

                saved := options.GeoM
                options.GeoM = geom2
                scale.DrawScaled(screen, top, options)
                options.GeoM = saved
            },
        }
    }

    return tiles
}

type CardinalDirection int
const (
    DirectionNorth CardinalDirection = iota
    DirectionEast
    DirectionSouth
    DirectionWest
)

func createWallArea(centerX int, centerY int, sideLength int, set func(int, int, CardinalDirection), inside func(int, int)) {
    minX := centerX - sideLength/2
    maxX := minX + sideLength - 1
    minY := centerY - sideLength/2
    maxY := minY + sideLength - 1

    for x := minX; x <= maxX; x++ {
        set(x, minY, DirectionWest)
        set(x, maxY, DirectionEast)
    }

    for y := minY; y <= maxY; y++ {
        set(minX, y, DirectionSouth)
        set(maxX, y, DirectionNorth)
    }

    for x := minX; x <= maxX; x++ {
        for y := minY; y <= maxY; y++ {
            inside(x, y)
        }
    }
}

// update the Fire set on the tiles centered around x/y with a length of sideLength
func createWallOfFire(tiles [][]Tile, centerX int, centerY int, sideLength int, activateCounter uint64) {
    set := func(x int, y int, direction CardinalDirection) {
        tile := &tiles[y][x]
        if tile.Fire == nil {
            tile.Fire = set.MakeSet[FireSide]()
        }

        tile.FireActive = activateCounter

        switch direction {
            case DirectionNorth: tile.Fire.Insert(FireSideNorth)
            case DirectionEast: tile.Fire.Insert(FireSideEast)
            case DirectionSouth: tile.Fire.Insert(FireSideSouth)
            case DirectionWest: tile.Fire.Insert(FireSideWest)
        }
    }

    inside := func(x int, y int) {
        tile := &tiles[y][x]
        tile.InsideFire = true
    }

    createWallArea(centerX, centerY, sideLength, set, inside)
}

func createCityWall(tiles [][]Tile, centerX int, centerY int, sideLength int) {
    set := func(x int, y int, direction CardinalDirection) {
        tile := &tiles[y][x]
        if tile.Wall == nil {
            tile.Wall = set.MakeSet[WallKind]()
        }

        switch direction {
            case DirectionNorth: tile.Wall.Insert(WallKindNorth)
            case DirectionEast: tile.Wall.Insert(WallKindEast)
            case DirectionSouth: tile.Wall.Insert(WallKindSouth)
            case DirectionWest: tile.Wall.Insert(WallKindWest)
        }
    }

    inside := func(x int, y int) {
        tile := &tiles[y][x]
        tile.InsideWall = true
    }

    createWallArea(centerX, centerY, sideLength, set, inside)

    // set gate tile

    minY := centerY - sideLength/2
    maxY := minY + sideLength - 1
    tiles[maxY][centerX-1].Wall.Clear()
    tiles[maxY][centerX-1].Wall.Insert(WallKindGate)
}

func createWallOfDarkness(tiles [][]Tile, centerX int, centerY int, sideLength int, activateCounter uint64) {
    set := func(x int, y int, direction CardinalDirection) {
        tile := &tiles[y][x]
        if tile.Darkness == nil {
            tile.Darkness = set.MakeSet[DarknessSide]()
        }

        tile.DarknessActive = activateCounter

        switch direction {
            case DirectionNorth: tile.Darkness.Insert(DarknessSideNorth)
            case DirectionEast: tile.Darkness.Insert(DarknessSideEast)
            case DirectionSouth: tile.Darkness.Insert(DarknessSideSouth)
            case DirectionWest: tile.Darkness.Insert(DarknessSideWest)
        }
    }

    inside := func(x int, y int) {
        tile := &tiles[y][x]
        tile.InsideDarkness = true
    }

    createWallArea(centerX, centerY, sideLength, set, inside)
}

// for units that are cursed with confusion, this is the action they will take on their turn
type ConfusionAction int
const (
    ConfusionActionNone ConfusionAction = iota
    ConfusionActionDoNothing
    ConfusionActionMoveRandomly
    ConfusionActionEnemyControl
)

type ArmyUnit struct {
    Unit units.StackUnit
    Facing units.Facing
    Moving bool
    X int
    Y int
    // Health int
    MovesLeft fraction.Fraction

    Spells spellbook.Spells
    SpellCharges map[spellbook.Spell]int
    CastingSkill float32
    Casted bool

    // track total damage applied to this unit
    NormalDamage int
    IrreversableDamage int
    UndeadDamage int

    // health of the web spell cast on this unit
    WebHealth int

    Model *CombatModel

    Team Team

    // how many units were just lost due to an attack/spell/something that caused damage
    LostUnits int
    LostUnitsTime int

    // height above the ground, negative for partially below ground
    Height int
    // how much alpha to apply to the unit, 1 is invisible, 0 is fully visible
    Fade float32

    // true if this unit was summoned via a spell
    Summoned bool

    // whether this unit is currently under the effects of the Confusion curse
    ConfusionAction ConfusionAction

    // number of times this unit was attacked this turn
    Attacked int

    // number of ranged attacks remaining
    RangedAttacks int

    Attacking bool
    Defending bool

    MovementTick uint64
    MoveX float64
    MoveY float64
    CurrentPath pathfinding.Path

    LastTurn int

    // enchantments applied to the unit during combat, usually by a spell
    Enchantments []data.UnitEnchantment
    // separate list of enchantments cast by the opposite wizard
    Curses []data.UnitEnchantment

    // ugly to need this, but this caches paths computed for the unit
    Paths map[image.Point]pathfinding.Path
}

func (unit *ArmyUnit) CanTeleport() bool {
    return unit.HasAbility(data.AbilityTeleporting) || unit.HasAbility(data.AbilityMerging)
}

func (unit *ArmyUnit) IsUndead() bool {
    return unit.Unit.IsUndead()
}

func (unit *ArmyUnit) SetFade(fade float32) {
    unit.Fade = fade
}

func (unit *ArmyUnit) SetHeight(height int) {
    unit.Height = height
}

func (unit *ArmyUnit) RaiseFromDead() {
    // make sure health is 0 first
    unit.Unit.AdjustHealth(-unit.GetMaxHealth())
    // then raise to 1/2
    unit.Heal(unit.GetMaxHealth()/2)
    unit.Enchantments = nil
    unit.Curses = nil
    unit.WebHealth = 0
}

func (unit *ArmyUnit) IsAsleep() bool {
    return unit.HasCurse(data.UnitCurseBlackSleep)
}

func (unit *ArmyUnit) IsWebbed() bool {
    return unit.WebHealth > 0
}

func (unit *ArmyUnit) GetCastingSkill() float32 {
    return unit.CastingSkill
}

func (unit *ArmyUnit) UseRangeAttack() {
    if unit.GetRangedAttackDamageType() == units.DamageRangedMagical && unit.CastingSkill >= 3 {
        unit.CastingSkill = max(0, unit.CastingSkill - 3)
    } else if unit.RangedAttacks > 0 {
        unit.RangedAttacks -= 1
    }
}

func (unit *ArmyUnit) GetDamageSource() DamageSource {
    if unit.Unit.IsHero() {
        return DamageSourceHero
    }

    if unit.GetRace() == data.RaceFantastic {
        return DamageSourceFantastic
    }

    return DamageSourceNormal
}

func (unit *ArmyUnit) CanNegateWeaponImmunity() bool {
    // any non-normal weapon
    if unit.GetWeaponBonus() != data.WeaponNone {
        return true
    }

    enchantments := []data.UnitEnchantment{
        data.UnitEnchantmentEldritchWeapon,
        data.UnitEnchantmentFlameBlade,
        data.UnitEnchantmentHolyWeapon,
    }

    for _, enchantment := range enchantments {
        if unit.HasEnchantment(enchantment) {
            return true
        }
    }

    if unit.Model.IsEnchantmentActive(data.CombatEnchantmentMetalFires, unit.Team) {
        return true
    }

    return false
}

func (unit *ArmyUnit) ProcessWeb() {
    if unit.WebHealth > 0 {
        damage := max(unit.GetMeleeAttackPower(), unit.GetRangedAttackPower())
        damage += int(unit.GetAbilityValue(data.AbilityFireBreath))
        unit.WebHealth -= damage
    }
}

func (unit *ArmyUnit) ReduceInvulnerability(damage int) int {
    if unit.HasEnchantment(data.UnitEnchantmentInvulnerability) {
        return max(0, damage - 2)
    }

    return damage
}

func (unit *ArmyUnit) IsFlying() bool {
    // a webbed unit is not flying
    return unit.Unit.IsFlying() && !unit.HasCurse(data.UnitCurseWeb)
}

func (unit *ArmyUnit) IsInvisible() bool {
    return unit.HasAbility(data.AbilityInvisibility) || unit.Model.IsEnchantmentActive(data.CombatEnchantmentMassInvisibility, unit.Team)
}

func (unit *ArmyUnit) IsSwimmer() bool {
    return unit.Unit.IsSwimmer()
}

func (unit *ArmyUnit) GetAbilities() []data.Ability {
    var enchantmentAbilities []data.Ability
    for _, enchantment := range unit.Enchantments {
        enchantmentAbilities = append(enchantmentAbilities, enchantment.Abilities()...)
    }

    return append(unit.Unit.GetAbilities(), enchantmentAbilities...)
}

func (unit *ArmyUnit) GetArtifactSlots() []artifact.ArtifactSlot {
    return unit.Unit.GetArtifactSlots()
}

func (unit *ArmyUnit) GetHeroExperienceLevel() units.HeroExperienceLevel {
    return unit.Unit.GetHeroExperienceLevel()
}

func (unit *ArmyUnit) GetExperienceLevel() units.NormalExperienceLevel {
    return unit.Unit.GetExperienceLevel()
}

func (unit *ArmyUnit) GetExperience() int {
    return unit.Unit.GetExperience()
}

func (unit *ArmyUnit) GetExperienceData() units.ExperienceData {
    return unit.Unit.GetExperienceData()
}

func (unit *ArmyUnit) GetRace() data.Race {
    return unit.Unit.GetRace()
}

func (unit *ArmyUnit) GetArtifacts() []*artifact.Artifact {
    return unit.Unit.GetArtifacts()
}

func (unit *ArmyUnit) GetBanner() data.BannerType {
    return unit.Unit.GetBanner()
}

func (unit *ArmyUnit) GetCombatIndex(facing units.Facing) int {
    return unit.Unit.GetCombatIndex(facing)
}

func (unit *ArmyUnit) GetCombatLbxFile() string {
    return unit.Unit.GetCombatLbxFile()
}

func (unit *ArmyUnit) GetCount() int {
    return unit.Unit.GetCount()
}

func (unit *ArmyUnit) GetVisibleCount() int {
    return unit.Unit.GetVisibleCount()
}

func (unit *ArmyUnit) GetBaseDefense() int {
    return unit.Unit.GetBaseDefense()
}

func (unit *ArmyUnit) GetBaseHitPoints() int {
    return unit.Unit.GetBaseHitPoints()
}

func (unit *ArmyUnit) GetBaseMeleeAttackPower() int {
    return unit.Unit.GetBaseMeleeAttackPower()
}

func (unit *ArmyUnit) GetBaseRangedAttackPower() int {
    return unit.Unit.GetBaseRangedAttackPower()
}

func (unit *ArmyUnit) GetBaseResistance() int {
    return unit.Unit.GetBaseResistance()
}

func (unit *ArmyUnit) GetFullHitPoints() int {
    base := unit.GetBaseHitPoints()
    for _, enchantment := range unit.Enchantments {
        base += unit.Unit.HitPointsEnchantmentBonus(enchantment)
    }
    return base
}

func (unit *ArmyUnit) GetHealth() int {
    return unit.GetMaxHealth() - unit.GetDamage()
}

func (unit *ArmyUnit) GetMaxHealth() int {
    return unit.GetFullHitPoints() * unit.GetCount()
}

func (unit *ArmyUnit) GetHitPoints() int {
    return (unit.GetMaxHealth() - unit.GetDamage()) / unit.GetCount()
    /*
    base := unit.Unit.GetHitPoints()

    for _, enchantment := range unit.Enchantments {
        base += unit.Unit.HitPointsEnchantmentBonus(enchantment)
    }

    return base
    */

    // return unit.Unit.GetHealth() / unit.Unit.GetCount()
}

func (unit *ArmyUnit) GetRangedAttackDamageType() units.Damage {
    return unit.Unit.GetRangedAttackDamageType()
}

func (unit *ArmyUnit) GetDamage() int {
    return unit.Unit.GetDamage()
}

func (unit *ArmyUnit) GetRealm() data.MagicType {
    if unit.Unit.IsUndead() {
        return data.DeathMagic
    }

    return unit.Unit.GetRealm()
}

func (unit *ArmyUnit) GetWeaponBonus() data.WeaponBonus {
    if unit.Unit.GetRace() != data.RaceFantastic && unit.Model.IsEnchantmentActive(data.CombatEnchantmentMetalFires, unit.Team) && !unit.HasEnchantment(data.UnitEnchantmentFlameBlade){
        if unit.Unit.GetWeaponBonus() == data.WeaponNone {
            return data.WeaponMagic
        }
    }

    return unit.Unit.GetWeaponBonus()
}

// true if this unit is immune to all effects from the given magic realm
func (unit *ArmyUnit) IsMagicImmune(magic data.MagicType) bool {
    if unit.HasAbility(data.AbilityMagicImmunity) {
        return true
    }

    switch magic {
        case data.DeathMagic:
            if unit.HasAbility(data.AbilityDeathImmunity) || unit.HasEnchantment(data.UnitEnchantmentRighteousness) {
                return true
            }
        case data.ChaosMagic:
            if unit.HasEnchantment(data.UnitEnchantmentRighteousness) {
                return true
            }
    }

    return false
}

// true if this unit has the same realm as the current magic node
func (unit *ArmyUnit) UnderNodeInfluence() bool {
    realm := unit.GetRealm()
    return realm != data.MagicNone && realm == unit.Model.Influence
}

func (unit *ArmyUnit) GetAbilityValue(ability data.AbilityType) float32 {
    // metal fires adds 1 to thrown attacks
    if ability == data.AbilityThrown {
        value := unit.Unit.GetAbilityValue(ability)
        if value > 0 {
            modifier := float32(0)

            if unit.UnderNodeInfluence() {
                modifier += 2
            }

            for _, enchantment := range unit.Enchantments {
                modifier += float32(unit.Unit.MeleeEnchantmentBonus(enchantment))
            }

            shattered := false

            for _, curse := range unit.Curses {
                switch curse {
                    case data.UnitCurseMindStorm: modifier -= 5
                    case data.UnitCurseShatter: shattered = true
                }
            }

            if unit.Model.IsEnchantmentActive(data.CombatEnchantmentBlackPrayer, oppositeTeam(unit.Team)) {
                modifier -= 1
            }

            // count metal fires, but only if flame blade is not active
            if unit.Unit.GetRace() != data.RaceFantastic && unit.Model.IsEnchantmentActive(data.CombatEnchantmentMetalFires, unit.Team) && !unit.HasEnchantment(data.UnitEnchantmentFlameBlade) {
                modifier += 1
            }

            if unit.Unit.GetRace() == data.RaceFantastic && unit.Model.IsEnchantmentActive(data.CombatEnchantmentDarkness, TeamEither) {
                switch unit.GetRealm() {
                    case data.DeathMagic: modifier += 1
                    case data.LifeMagic: modifier -= 1
                }
            }

            final := value + modifier

            if shattered && final > 0 {
                return 1
            }

            return max(0, final)
        }
    }

    if ability == data.AbilityFireBreath || ability == data.AbilityLightningBreath {
        value := unit.Unit.GetAbilityValue(ability)
        if value > 0 {
            modifier := float32(0)

            shattered := false

            if unit.UnderNodeInfluence() {
                modifier += 2
            }

            for _, curse := range unit.Curses {
                switch curse {
                    case data.UnitCurseMindStorm: modifier -= 5
                    case data.UnitCurseShatter: shattered = true
                }
            }

            if unit.Unit.GetRace() == data.RaceFantastic && unit.Model.IsEnchantmentActive(data.CombatEnchantmentDarkness, TeamEither) {
                switch unit.GetRealm() {
                    case data.DeathMagic: modifier += 1
                    case data.LifeMagic: modifier -= 1
                }
            }

            final := value + modifier

            if shattered && final > 0 {
                return 1
            }

            return max(0, final)
        }

        return value
    }

    if ability == data.AbilityPoisonTouch /* || ability == data.AbilityLifeSteal */ || ability == data.AbilityStoningTouch ||
       ability == data.AbilityDispelEvil || ability == data.AbilityDeathTouch /* || ability == data.AbilityDestruction */ {

        // FIXME: how does the value get used for dispel evil, death touch, destruction?
        // FIXME: life steal is already negative, so subtracting 1 would make it even more powerful

        value := unit.Unit.GetAbilityValue(ability)
        if value > 0 {
            modifier := float32(0)

            if unit.Model.IsEnchantmentActive(data.CombatEnchantmentBlackPrayer, oppositeTeam(unit.Team)) {
                modifier -= 1
            }

            return max(0, value + modifier)
        }

        return value
    }

    // FIXME: add magic influence to doom gaze (and maybe other gazes)

    return unit.Unit.GetAbilityValue(ability)
}

func (unit *ArmyUnit) GetCounterAttackToHit(defender *ArmyUnit) int {
    base := unit.GetToHitMelee(defender)
    // if somehow the unit already has <10% tohit then just return that
    if base < 10 {
        return base
    }

    // for every 2 attacks against this unit, reduce tohit by 10%
    reduction := 10 * (unit.Attacked / 2)

    // tohit cannot go below 10%
    return max(10, base - reduction)
}

// FIXME: needs a type passed in (melee, ranged, etc)
func (unit *ArmyUnit) GetToHitMelee(defender *ArmyUnit) int {
    modifier := 0

    if unit.Model.IsEnchantmentActive(data.CombatEnchantmentHighPrayer, unit.Team) ||
       unit.Model.IsEnchantmentActive(data.CombatEnchantmentPrayer, unit.Team) {
        modifier += 10
    }

    if unit.HasCurse(data.UnitCurseVertigo) {
        modifier -= 20
    }

    if defender.HasAbility(data.AbilityLucky) {
        modifier -= 10
    }

    if defender.IsInvisible() && !unit.HasAbility(data.AbilityIllusionsImmunity) {
        modifier -= 10
    }

    // FIXME: blur doesn't affect to hit, instead it directly reduces damage points
    if unit.Model.IsEnchantmentActive(data.CombatEnchantmentBlur, oppositeTeam(unit.Team)) {
        modifier -= 10
    }

    if unit.Model.IsEnchantmentActive(data.CombatEnchantmentWarpReality, TeamEither) {
        // all non chaos fantastic units get -10 to hit
        isChaos := unit.Unit.GetRace() == data.RaceFantastic && unit.GetRealm() == data.ChaosMagic
        if !isChaos {
            modifier -= 10
        }
    }

    return max(0, unit.Unit.GetToHitMelee() + modifier)
}

type UnitResistance interface {
    GetResistance() int
    HasEnchantment(data.UnitEnchantment) bool
    HasAbility(data.AbilityType) bool
}

// get the resistance of the unit, taking into account enchantments and curses that apply to the specific magic type
func GetResistanceFor(unit UnitResistance, magic data.MagicType) int {
    base := unit.GetResistance()
    modifier := 0

    if unit.HasEnchantment(data.UnitEnchantmentBless) {
        switch magic {
            case data.DeathMagic, data.ChaosMagic: modifier += 3
        }
    }

    if unit.HasEnchantment(data.UnitEnchantmentRighteousness) {
        switch magic {
            case data.DeathMagic, data.ChaosMagic: modifier += 30
        }
    }

    if unit.HasEnchantment(data.UnitEnchantmentResistMagic) {
        modifier += 5
    }

    if unit.HasEnchantment(data.UnitEnchantmentElementalArmor) {
        switch magic {
            case data.NatureMagic, data.ChaosMagic: modifier += 10
        }
    }

    if unit.HasEnchantment(data.UnitEnchantmentResistElements) {
        switch magic {
            case data.NatureMagic, data.ChaosMagic: modifier += 3
        }
    }

    if unit.HasAbility(data.AbilityDeathImmunity) && magic == data.DeathMagic {
        modifier = 50
    }

    if unit.HasAbility(data.AbilityMagicImmunity) {
        modifier = 50
    }

    return base + modifier
}

func (unit *ArmyUnit) GetResistance() int {
    modifier := 0

    // magic node influencing fantastic creatures
    if unit.UnderNodeInfluence() {
        modifier += 2
    }

    // charmed heroes have +30 resistance during battle
    if unit.HasAbility(data.AbilityCharmed) {
        modifier += 30
    }

    for _, enchantment := range unit.Enchantments {
        modifier += unit.Unit.ResistanceEnchantmentBonus(enchantment)
    }

    for _, curse := range unit.Curses {
        switch curse {
            case data.UnitCurseMindStorm: modifier -= 5
            // FIXME: it could be the case that warp creature sets the base resistance to 0, but leaves the modifier alone
            // should the various enchantments in effect still apply if the unit has warp creature resistance?
            case data.UnitCurseWarpCreatureResistance: return 0
        }
    }

    hasRighteousness := unit.HasEnchantment(data.UnitEnchantmentRighteousness)

    if !hasRighteousness && unit.Model.IsEnchantmentActive(data.CombatEnchantmentBlackPrayer, oppositeTeam(unit.Team)) {
        modifier -= 2
    }

    if unit.Model.IsEnchantmentActive(data.CombatEnchantmentHighPrayer, unit.Team) ||
       unit.Model.IsEnchantmentActive(data.CombatEnchantmentPrayer, unit.Team) {
        modifier += 3
    }

    if unit.Unit.GetRace() == data.RaceFantastic && unit.Model.IsEnchantmentActive(data.CombatEnchantmentDarkness, TeamEither) {
        switch unit.GetRealm() {
            case data.DeathMagic: modifier += 1
            case data.LifeMagic:
                if !hasRighteousness {
                    modifier -= 1
                }
        }
    }

    if unit.Unit.GetRace() == data.RaceFantastic && unit.Model.IsEnchantmentActive(data.CombatEnchantmentTrueLight, TeamEither) {
        switch unit.GetRealm() {
            case data.LifeMagic: modifier += 1
            case data.DeathMagic: modifier -= 1
        }
    }

    return max(0, unit.Unit.GetResistance() + modifier)
}

// get defense against a specific magic type
func GetDefenseFor(unit UnitDamage, magic data.MagicType) int {
    // berserk prevents any enchantments from applying
    if unit.HasEnchantment(data.UnitEnchantmentBerserk) {
        return 0
    }

    base := unit.GetDefense()
    modifier := 0

    if unit.HasEnchantment(data.UnitEnchantmentBless) {
        switch magic {
            case data.DeathMagic, data.ChaosMagic: modifier += 3
        }
    }

    if unit.HasEnchantment(data.UnitEnchantmentElementalArmor) {
        switch magic {
            case data.NatureMagic, data.ChaosMagic: modifier += 10
        }
    }

    if unit.HasEnchantment(data.UnitEnchantmentResistElements) {
        switch magic {
            case data.NatureMagic, data.ChaosMagic: modifier += 3
        }
    }

    return base + modifier
}

func (unit *ArmyUnit) GetDefense() int {
    if unit.HasEnchantment(data.UnitEnchantmentBerserk) {
        return 0
    }

    modifier := 0

    // magic node influencing fantastic creatures
    if unit.UnderNodeInfluence() {
        modifier += 2
    }

    for _, enchantment := range unit.Enchantments {
        modifier += unit.Unit.DefenseEnchantmentBonus(enchantment)
    }

    warpCreatureDefense := false

    for _, curse := range unit.Curses {
        switch curse {
            case data.UnitCurseMindStorm: modifier -= 5
            case data.UnitCurseVertigo: modifier -= 1
            case data.UnitCurseWarpCreatureDefense: warpCreatureDefense = true
        }
    }

    if unit.Model.IsEnchantmentActive(data.CombatEnchantmentHighPrayer, unit.Team) {
        modifier += 2
    }

    if unit.Model.IsEnchantmentActive(data.CombatEnchantmentBlackPrayer, oppositeTeam(unit.Team)) {
        modifier -= 1
    }

    if unit.Unit.GetRace() == data.RaceFantastic && unit.Model.IsEnchantmentActive(data.CombatEnchantmentDarkness, TeamEither) {
        switch unit.GetRealm() {
            case data.DeathMagic: modifier += 1
            case data.LifeMagic: modifier -= 1
        }
    }

    if unit.Unit.GetRace() == data.RaceFantastic && unit.Model.IsEnchantmentActive(data.CombatEnchantmentTrueLight, TeamEither) {
        switch unit.GetRealm() {
            case data.LifeMagic: modifier += 1
            case data.DeathMagic: modifier -= 1
        }
    }

    final := unit.Unit.GetDefense() + modifier
    if warpCreatureDefense {
        final /= 2
    }

    return max(0, final)
}

func (unit *ArmyUnit) GetRangedAttackPower() int {
    if unit.Unit.GetRangedAttackPower() == 0 {
        return 0
    }

    modifier := 0

    // magic node influencing fantastic creatures
    if unit.UnderNodeInfluence() {
        modifier += 2
    }

    for _, enchantment := range unit.Enchantments {
        modifier += unit.Unit.RangedEnchantmentBonus(enchantment)
    }

    shattered := false

    for _, curse := range unit.Curses {
        switch curse {
            case data.UnitCurseMindStorm: modifier -= 5
            case data.UnitCurseWeakness:
                if unit.Unit.GetRangedAttackDamageType() == units.DamageRangedPhysical {
                    modifier -= 2
                }
            case data.UnitCurseShatter: shattered = true
        }
    }

    if unit.Unit.GetRace() != data.RaceFantastic && unit.Model.IsEnchantmentActive(data.CombatEnchantmentMetalFires, unit.Team) && !unit.HasEnchantment(data.UnitEnchantmentFlameBlade) {
        if unit.GetRangedAttackDamageType() == units.DamageRangedPhysical {
            modifier += 1
        }
    }

    final := unit.Unit.GetRangedAttackPower() + modifier

    if shattered && final > 0 {
        return 1
    }

    return max(0, final)
}

func (unit *ArmyUnit) GetMeleeAttackPower() int {
    modifier := 0

    // magic node influencing fantastic creatures
    if unit.UnderNodeInfluence() {
        modifier += 2
    }

    for _, enchantment := range unit.Enchantments {
        modifier += unit.Unit.MeleeEnchantmentBonus(enchantment)
    }

    shattered := false
    warpCreatureMelee := false

    for _, curse := range unit.Curses {
        switch curse {
            case data.UnitCurseMindStorm: modifier -= 5
            case data.UnitCurseWeakness: modifier -= 2
            case data.UnitCurseShatter: shattered = true
            case data.UnitCurseWarpCreatureMelee: warpCreatureMelee = true
        }
    }

    if unit.Model.IsEnchantmentActive(data.CombatEnchantmentHighPrayer, unit.Team) {
        if unit.Unit.GetMeleeAttackPower() > 0 {
            modifier += 2
        }
    }

    if (unit.Unit.GetRace() == data.RaceFantastic && unit.GetRealm() == data.DeathMagic) &&
        unit.Model.IsEnchantmentActive(data.CombatEnchantmentDarkness, TeamEither) {
        modifier += 1
    }

    if (unit.Unit.GetRace() == data.RaceFantastic && unit.GetRealm() == data.LifeMagic) &&
        unit.Model.IsEnchantmentActive(data.CombatEnchantmentDarkness, TeamEither) {
        modifier -= 1
    }

    if unit.Unit.GetRace() != data.RaceFantastic && unit.Model.IsEnchantmentActive(data.CombatEnchantmentMetalFires, unit.Team) && !unit.HasEnchantment(data.UnitEnchantmentFlameBlade) {
        if unit.Unit.GetMeleeAttackPower() > 0 {
            modifier += 1
        }
    }

    if unit.Unit.GetRace() == data.RaceFantastic && unit.Model.IsEnchantmentActive(data.CombatEnchantmentTrueLight, TeamEither) {
        switch unit.GetRealm() {
            case data.LifeMagic: modifier += 1
            case data.DeathMagic: modifier -= 1
        }
    }

    if unit.Model.IsEnchantmentActive(data.CombatEnchantmentBlackPrayer, oppositeTeam(unit.Team)) {
        modifier -= 1
    }

    berserkModifier := 1
    if unit.HasEnchantment(data.UnitEnchantmentBerserk) {
        berserkModifier = 2
    }

    final := (unit.Unit.GetMeleeAttackPower() + modifier) * berserkModifier

    if shattered && final > 0 {
        return 1
    }

    if warpCreatureMelee {
        final /= 2
    }

    return max(0, final)
}

func (unit *ArmyUnit) HasAbility(ability data.AbilityType) bool {
    if unit.Unit.HasAbility(ability) {
        return true
    }

    for _, enchantment := range unit.Enchantments {
        for _, grantedAbility := range enchantment.Abilities() {
            if grantedAbility.Ability == ability {
                return true
            }
        }
    }

    return false
}

func (unit *ArmyUnit) HasEnchantmentOnly(enchantment data.UnitEnchantment) bool {
    return slices.Contains(unit.Enchantments, enchantment)
}

func (unit *ArmyUnit) HasEnchantment(enchantment data.UnitEnchantment) bool {
    if unit.Unit.HasEnchantment(enchantment) {
        return true
    }

    for _, check := range unit.Enchantments {
        if check == enchantment {
            return true
        }
    }

    return false
}

// roughly represents the strength of this unit, used for strategic combat
func (unit *ArmyUnit) GetPower() int {
    power := 0

    power += unit.Unit.GetMaxHealth()
    power += unit.GetDefense()
    power += unit.GetResistance()
    power += unit.GetRangedAttackPower() * unit.Figures()
    power += unit.GetMeleeAttackPower() * unit.Figures()

    return power
}

// true if this unit can move through a tile with a wall tile
func (unit *ArmyUnit) CanTraverseWall() bool {
    return unit.IsFlying() || unit.HasAbility(data.AbilityMerging) || unit.HasAbility(data.AbilityTeleporting)
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

func (unit *ArmyUnit) GetCurses() []data.UnitEnchantment{
    return unit.Curses
}

func (unit *ArmyUnit) HasCurse(curse data.UnitEnchantment) bool {
    return slices.Contains(unit.Curses, curse)
}

func (unit *ArmyUnit) AddCurse(curse data.UnitEnchantment) {
    // skip duplicates
    if unit.HasCurse(curse) {
        return
    }

    unit.Curses = append(unit.Curses, curse)
}

func (unit *ArmyUnit) RemoveCurse(curse data.UnitEnchantment) {
    unit.Curses = slices.DeleteFunc(unit.Curses, func(check data.UnitEnchantment) bool {
        return check == curse
    })
}

func (unit *ArmyUnit) GetEnchantments() []data.UnitEnchantment {
    return append(append(slices.Clone(unit.Unit.GetEnchantments()), unit.Enchantments...), unit.Curses...)
}

func (unit *ArmyUnit) RemoveEnchantment(enchantment data.UnitEnchantment) {
    unit.Enchantments = slices.DeleteFunc(unit.Enchantments, func(check data.UnitEnchantment) bool {
        return enchantment == check
    })

    unit.Unit.RemoveEnchantment(enchantment)
}

func (unit *ArmyUnit) AddEnchantment(enchantment data.UnitEnchantment) {
    // skip duplicates
    if unit.HasEnchantment(enchantment) {
        return
    }

    unit.Enchantments = append(unit.Enchantments, enchantment)
}

func (unit *ArmyUnit) CanCast() bool {
    if unit.Casted {
        return false
    }

    for _, charges := range unit.SpellCharges {
        if charges > 0 {
            return true
        }
    }

    for _, spell := range unit.Spells.Spells {
        if int(unit.CastingSkill) >= spell.CastCost {
            return true
        }
    }

    return false
}

func (unit *ArmyUnit) GetMovementSpeed() fraction.Fraction {
    modifier := fraction.Zero()
    base := unit.Unit.GetMovementSpeed()

    base = unit.Unit.MovementSpeedEnchantmentBonus(base, unit.Enchantments)

    if unit.Model.IsEnchantmentActive(data.CombatEnchantmentEntangle, oppositeTeam(unit.Team)) {
        unaffected := unit.IsFlying() || unit.HasAbility(data.AbilityNonCorporeal)

        if !unaffected {
            modifier = modifier.Subtract(fraction.FromInt(1))
        }
    }

    return fraction.Zero().Max(base.Add(modifier))
}

func (unit *ArmyUnit) ResetTurnData() {
    unit.MovesLeft = unit.GetMovementSpeed()
    unit.Paths = make(map[image.Point]pathfinding.Path)
    unit.Casted = false
    unit.Attacked = 0

    unit.ConfusionAction = ConfusionActionNone

    if unit.HasCurse(data.UnitCurseConfusion) {
        actions := []ConfusionAction{ConfusionActionDoNothing, ConfusionActionMoveRandomly, ConfusionActionEnemyControl, ConfusionActionNone}
        unit.ConfusionAction = actions[rand.N(len(actions))]
    }

}

type DamageModifiers struct {
    // 50% less defense
    ArmorPiercing bool
    // 100% less defense
    Illusion bool

    // any protection offered by the city wall
    WallDefense int

    NegateWeaponImmunity bool

    EldritchWeapon bool

    // if the damage comes from a specific realm (for spells or magic damage)
    Magic data.MagicType

    DamageType DamageType
}

type UnitDamage interface {
    IsAsleep() bool
    HasAbility(ability data.AbilityType) bool
    HasEnchantment(enchantment data.UnitEnchantment) bool
    GetDefense() int
    ToDefend(modifiers DamageModifiers) int
    // returns the number of figures lost
    TakeDamage(damage int, damageType DamageType) int
    ReduceInvulnerability(damage int) int
    GetHealth() int
    GetMaxHealth() int
    GetCount() int
    Figures() int
}

func ComputeDefense(unit UnitDamage, damage units.Damage, source DamageSource, modifiers DamageModifiers) int {
    if unit.IsAsleep() {
        return 0
    }

    toDefend := unit.ToDefend(modifiers)
    var defenseRolls int

    hasImmunity := false

    switch damage {
        case units.DamageRangedMagical:
            defenseRolls = GetDefenseFor(unit, modifiers.Magic)
            if unit.HasAbility(data.AbilityLargeShield) {
                defenseRolls += 2
            }

            if unit.HasAbility(data.AbilityMagicImmunity) {
                hasImmunity = true
            }
        case units.DamageRangedPhysical:
            defenseRolls = unit.GetDefense()
            if unit.HasAbility(data.AbilityLargeShield) {
                defenseRolls += 2
            }

            if unit.HasAbility(data.AbilityMissileImmunity) {
                hasImmunity = true
            }
        case units.DamageImmolation:
            defenseRolls = GetDefenseFor(unit, data.ChaosMagic)
            if unit.HasAbility(data.AbilityLargeShield) {
                defenseRolls += 2
            }

            if unit.HasAbility(data.AbilityMagicImmunity) || unit.HasEnchantment(data.UnitEnchantmentRighteousness) {
                // always completely immune to immolation
                return 1000
            }

            if unit.HasAbility(data.AbilityFireImmunity) {
                hasImmunity = true
            }

        case units.DamageFire:
            defenseRolls = GetDefenseFor(unit, modifiers.Magic)
            if unit.HasAbility(data.AbilityLargeShield) {
                defenseRolls += 2
            }

            if unit.HasAbility(data.AbilityMagicImmunity) || unit.HasAbility(data.AbilityFireImmunity) {
                hasImmunity = true
            }
        case units.DamageCold:
            defenseRolls = GetDefenseFor(unit, data.NatureMagic)
            if unit.HasAbility(data.AbilityLargeShield) {
                defenseRolls += 2
            }

            if unit.HasAbility(data.AbilityMagicImmunity) || unit.HasAbility(data.AbilityColdImmunity) {
                hasImmunity = true
            }
        case units.DamageThrown:
            defenseRolls = unit.GetDefense()
            if unit.HasAbility(data.AbilityLargeShield) {
                defenseRolls += 2
            }
        default:
            defenseRolls = unit.GetDefense()
    }

    switch damage {
        case units.DamageMeleePhysical, units.DamageRangedPhysical, units.DamageThrown:
            if unit.HasAbility(data.AbilityWeaponImmunity) && !modifiers.NegateWeaponImmunity && source == DamageSourceNormal {
                defenseRolls = max(defenseRolls, 10)
            }
    }

    if modifiers.ArmorPiercing {
        defenseRolls /= 2
    }

    // after armor piercing, wall defense is applied
    switch damage {
        case units.DamageRangedMagical,
             units.DamageRangedPhysical,
             units.DamageMeleePhysical:
            defenseRolls += modifiers.WallDefense
    }

    if hasImmunity {
        defenseRolls = 50
    }

    if modifiers.Illusion && !unit.HasAbility(data.AbilityIllusionsImmunity) {
        defenseRolls = 0
    }

    // log.Printf("Unit %v has %v defense", unit.Unit.GetName(), defenseRolls)

    defense := ComputeRoll(defenseRolls, toDefend)

    return defense
}

// returns the damage reason why this unit died, which is the largest source of damage
func (unit *ArmyUnit) DeathReason() DamageType {
    if unit.NormalDamage >= unit.IrreversableDamage && unit.NormalDamage >= unit.UndeadDamage {
        return DamageNormal
    }

    if unit.UndeadDamage >= unit.IrreversableDamage && unit.UndeadDamage >= unit.NormalDamage {

        invalid := unit.Unit.IsHero() || unit.GetRealm() == data.DeathMagic || unit.Summoned || unit.HasAbility(data.AbilityMagicImmunity)
        if !invalid {
            return DamageUndead
        }
    }

    if unit.IrreversableDamage >= unit.NormalDamage && unit.IrreversableDamage >= unit.UndeadDamage {
        return DamageIrreversable
    }

    return DamageNormal
}

func (unit *ArmyUnit) TakeDamage(damage int, damageType DamageType) int {
    // keep track of how many figures were in this unit before damage occurred, so
    // that we can show a death animation for the lost figures
    visibleFigures := unit.VisibleFigures()
    // unit.LastVisibleFigures = unit.VisibleFigures()

    // the first figure should take damage, and if it dies then the next unit takes damage, etc
    unit.Unit.AdjustHealth(-damage)

    switch damageType {
        case DamageNormal: unit.NormalDamage += damage
        case DamageIrreversable: unit.IrreversableDamage += damage
        case DamageUndead: unit.UndeadDamage += damage
    }

    lost := visibleFigures - unit.VisibleFigures()

    if unit.LostUnitsTime == 0 {
        // length of time (in updates) to show the lost units
        unit.LostUnitsTime = LostUnitsMax
        unit.LostUnits = lost
    } else {
        unit.LostUnits += lost
    }

    return lost
}

func (unit *ArmyUnit) Heal(amount int){
    maxHeal := unit.GetDamage() - unit.IrreversableDamage
    unit.Unit.AdjustHealth(max(0, min(amount, maxHeal)))
}

// apply damage to each individual figure such that each figure gets to individually block damage.
// this could potentially allow a damage of 5 to destroy a unit with 4 figures of 1HP each
// returns damage taken and number of visible units lost
func ApplyAreaDamage(unit UnitDamage, attackStrength int, damageType units.Damage, wallDefense int) (int, int) {
    totalDamage := 0
    health_per_figure := unit.GetMaxHealth() / unit.GetCount()

    modifiers := DamageModifiers{WallDefense: wallDefense}

    for range unit.Figures() {
        // FIXME: should this toHit=30 be based on the unit's toHitMelee?
        damage := ComputeRoll(attackStrength, 30)

        defense := ComputeDefense(unit, damageType, DamageSourceSpell, modifiers)

        // can't do more damage than a single figure has HP
        figureDamage := unit.ReduceInvulnerability(min(damage - defense, health_per_figure))
        if figureDamage > 0 {
            totalDamage += figureDamage
        }
    }

    totalDamage = min(totalDamage, unit.GetHealth())
    lost := unit.TakeDamage(totalDamage, DamageNormal)
    return totalDamage, lost
}

// apply damage to lead figure, and if it dies then keep applying remaining damage to the next figure
// FIXME: its possible that the damage can be passed to ComputeRoll() to determine how much actual damage is done
// returns damage taken and the number of visible figures lost
func ApplyDamage(unit UnitDamage, damage int, damageType units.Damage, source DamageSource, modifiers DamageModifiers) (int, int) {
    isMagic := damageType == units.DamageRangedMagical || damageType == units.DamageFire || damageType == units.DamageCold
    if isMagic && unit.HasAbility(data.AbilityMagicImmunity) {
        return 0, 0
    }

    taken := 0
    lost := 0
    for damage > 0 && unit.GetHealth() > 0 {
        // compute defense, apply damage to lead figure. if lead figure dies, apply damage to next figure
        defense := ComputeDefense(unit, damageType, source, modifiers)
        damage -= defense

        if damage > 0 {
            health_per_figure := unit.GetMaxHealth() / unit.GetCount()
            healthLeft := unit.GetHealth() % unit.Figures()
            if healthLeft == 0 {
                healthLeft = health_per_figure
            }

            take := min(healthLeft, damage)
            lost += unit.TakeDamage(take, modifiers.DamageType)
            damage -= take

            taken += take
        }
    }

    return taken, lost
}

func (unit *ArmyUnit) InitializeSpells(allSpells spellbook.Spells, player ArmyPlayer) {
    unit.CastingSkill = 0
    unit.SpellCharges = make(map[spellbook.Spell]int)
    for _, ability := range unit.Unit.GetAbilities() {
        switch ability.Ability {
            case data.AbilityDoomBoltSpell:
                doomBolt := allSpells.FindByName("Doom Bolt")
                unit.SpellCharges[doomBolt] = int(ability.Value)
                // unit.Spells.AddSpell(doomBolt)
                // unit.CastingSkill += float32(doomBolt.CastCost)
            case data.AbilityFireballSpell:
                fireball := allSpells.FindByName("Fireball")
                unit.SpellCharges[fireball] = int(ability.Value)
            case data.AbilityHealingSpell:
                healing := allSpells.FindByName("Healing")
                unit.SpellCharges[healing] = int(ability.Value)
            case data.AbilityWebSpell:
                web := allSpells.FindByName("Web")
                unit.SpellCharges[web] = int(ability.Value)
            case data.AbilityCaster:
                unit.CastingSkill = ability.Value
        }
    }

    // for units that are casters
    for _, knownSpell := range unit.Unit.GetKnownSpells() {
        spell := allSpells.FindByName(knownSpell)
        if spell.Valid() {
            unit.Spells.AddSpell(spell)
        } else {
            log.Printf("Error: unable to find spell %v for %v", knownSpell, unit.Unit.GetName())
        }
    }

    if unit.Unit.IsHero() {
        unit.Spells.AddAllSpells(player.GetKnownSpells())
        unit.SpellCharges = unit.Unit.GetSpellChargeSpells()
    } else {
        if unit.GetRealm() != data.MagicNone {
            unit.Spells.AddAllSpells(allSpells.GetSpellsByMagic(unit.GetRealm()))
        }
    }
}

func (unit *ArmyUnit) GetRangedAttacks() int {
    return unit.RangedAttacks
}

func (unit *ArmyUnit) SetRangedAttacks(attacks int) {
    unit.RangedAttacks = attacks
}

// given the distance to the target in tiles, return the amount of range damage done
func (unit *ArmyUnit) ComputeRangeDamage(defender *ArmyUnit, tileDistance int) int {

    toHit := unit.GetToHitMelee(defender)

    // FIXME: if the unit has Holy Weapon and this is a magic ranged attack then the +10% to-hit should not apply

    // magical attacks don't suffer a to-hit penalty
    if unit.Unit.GetRangedAttackDamageType() != units.DamageRangedMagical {

        if unit.HasAbility(data.AbilityLongRange) {
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

    return defender.ReduceInvulnerability(ComputeRoll(unit.GetRangedAttackPower(), toHit))
}

func (unit *ArmyUnit) ComputeMeleeDamage(defender *ArmyUnit, fearFigure int, counterAttack bool) (int, bool) {

    if unit.GetMeleeAttackPower() == 0 {
        return 0, false
    }

    damage := 0
    hit := false
    for range unit.Figures() - fearFigure {
        // even if all figures fail to cause damage, it still counts as a hit for touch purposes
        hit = true

        toHit := unit.GetToHitMelee(defender)
        if counterAttack {
            // counter attack to-hit might be penalized
            toHit = unit.GetCounterAttackToHit(defender)
        }
        damage += defender.ReduceInvulnerability(ComputeRoll(unit.GetMeleeAttackPower(), toHit))
    }

    return damage, hit
}

// return how many units should become afraid
func (unit *ArmyUnit) CauseFear() int {
    fear := 0

    if unit.HasAbility(data.AbilityMagicImmunity) || unit.HasAbility(data.AbilityDeathImmunity) || unit.HasAbility(data.AbilityCharmed) {
        return 0
    }

    if unit.HasEnchantment(data.UnitEnchantmentRighteousness) {
        return 0
    }

    resistance := GetResistanceFor(unit, data.DeathMagic)

    for range unit.Figures() {
        if rand.N(10) + 1 > resistance {
            fear += 1
        }
    }

    return fear
}

func (unit *ArmyUnit) ToDefend(modifiers DamageModifiers) int {
    modifier := 0

    if unit.Model.IsEnchantmentActive(data.CombatEnchantmentHighPrayer, unit.Team) ||
       unit.Model.IsEnchantmentActive(data.CombatEnchantmentPrayer, unit.Team) {
        modifier += 10
    }

    if modifiers.EldritchWeapon {
        modifier -= 10
    }

    return unit.Unit.GetToDefend() + modifier
}

// number of alive figures in this unit
func (unit *ArmyUnit) Figures() int {

    // health per figure = max health / figures
    // figures = health / health per figure

    health_per_figure := float64(unit.GetMaxHealth()) / float64(unit.GetCount())
    return int(math.Ceil(float64(unit.GetHealth()) / health_per_figure))
}

// lame to need another function just for visible figures
func (unit *ArmyUnit) VisibleFigures() int {
    health_per_figure := float64(unit.GetMaxHealth()) / float64(unit.GetVisibleCount())
    return int(math.Ceil(float64(unit.GetHealth()) / health_per_figure))
}

type Army struct {
    Player ArmyPlayer
    ManaPool int
    Range fraction.Fraction
    // when counter magic is cast, this field tracks how much 'counter magic' strength is available to dispel
    CounterMagic int
    units []*ArmyUnit
    KilledUnits []*ArmyUnit
    RegeneratedUnits []*ArmyUnit
    Auto bool
    Fled bool
    Casted bool
    RecalledUnits []*ArmyUnit

    Enchantments []data.CombatEnchantment
    Cleanups []func()
}

func (army *Army) AddEnchantment(enchantment data.CombatEnchantment) bool {
    for _, check := range army.Enchantments {
        if check == enchantment {
            return false
        }
    }
    army.Enchantments = append(army.Enchantments, enchantment)
    return true
}

func (army *Army) HasEnchantment(enchantment data.CombatEnchantment) bool {
    return slices.ContainsFunc(army.Enchantments, func(check data.CombatEnchantment) bool {
        return check == enchantment
    })
}

func (army *Army) RemoveEnchantment(enchamtent data.CombatEnchantment) {
    army.Enchantments = slices.DeleteFunc(army.Enchantments, func(check data.CombatEnchantment) bool {
        return check == enchamtent
    })
}

// remove mutations done to the underlying stack units
func (army *Army) Cleanup() {
    for _, cleanup := range army.Cleanups {
        cleanup()
    }
}

func (army *Army) GetUnits() []*ArmyUnit {
    return army.units
}

// a number that mostly represents the strength of this army
func (army *Army) GetPower() int {
    power := 0

    for _, unit := range army.units {
        power += unit.GetPower()
    }

    return power
}

func (army *Army) IsAI() bool {
    return army.Auto || army.Player.IsAI()
}

/* must call LayoutUnits() some time after invoking AddUnit() to ensure
 * the units are laid out correctly
 */
func (army *Army) AddUnit(unit units.StackUnit){
    armyUnit := &ArmyUnit{
        Unit: unit,
        Facing: units.FacingDownRight,
        // Health: unit.GetMaxHealth(),
    }
    // Warning: it is imperative that unit.SetEnchantmentProvider(nil) is called when combat ends
    unit.SetEnchantmentProvider(armyUnit)
    army.Cleanups = append(army.Cleanups, func(){
        unit.SetEnchantmentProvider(nil)
    })
    army.AddArmyUnit(armyUnit)
}

func (army *Army) AddArmyUnit(unit *ArmyUnit){
    army.units = append(army.units, unit)
}

func (army *Army) LayoutUnits(team Team){
    x := TownCenterX - 2
    y := 11
    rowDirection := -1

    facing := units.FacingDownRight

    if team == TeamAttacker {
        x = TownCenterX - 2
        y = 17
        rowDirection = 1
        facing = units.FacingUpLeft
    }

    cx := x
    cy := y

    columns := int(math.Round(math.Log(float64(len(army.units))) * 2))
    if columns < 4 {
        columns = 4
    }

    row := 0
    for _, unit := range army.units {
        unit.X = cx
        unit.Y = cy
        unit.Facing = facing

        cx += 1
        row += 1
        if row >= columns {
            row = 0
            cx = x
            cy += rowDirection
        }
    }
}

func (army *Army) RaiseDeadUnit(unit *ArmyUnit, x int, y int){
    unit.RaiseFromDead()
    unit.X = x
    unit.Y = y
    army.units = append(army.units, unit)
    army.KilledUnits = slices.DeleteFunc(army.KilledUnits, func(check *ArmyUnit) bool {
        return check == unit
    })
}

func (army *Army) KillUnit(kill *ArmyUnit){
    // units that died due to irreversable damage are gone forever
    if kill.DeathReason() != DamageIrreversable {
        army.KilledUnits = append(army.KilledUnits, kill)
    }
    army.RemoveUnit(kill)
}

func (army *Army) RemoveUnit(remove *ArmyUnit){
    army.units = slices.DeleteFunc(army.units, func(check *ArmyUnit) bool {
        return check == remove
    })
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

type CombatLogEvent struct {
    Turn int
    Text string
    AbsoluteTime time.Time
}

type CombatModel struct {
    SelectedUnit *ArmyUnit
    DefendingArmy *Army
    AttackingArmy *Army
    Tiles [][]Tile
    // when the user hovers over a unit, that unit should be shown in a little info box at the upper right
    HighlightedUnit *ArmyUnit
    OtherUnits []*OtherUnit
    Projectiles []*Projectile
    Plane data.Plane
    Zone ZoneType
    // the type of magic that is influencing this combat because the combat takes place near a magic node
    Influence data.MagicType

    // units that became undead once combat ends
    UndeadUnits []*ArmyUnit

    Events chan CombatEvent

    TurnAttacker int
    TurnDefender int

    Cleanups []func()

    // track how many units were killed on each side, so experience
    // can be given out after combat ends
    DefeatedDefenders int
    DefeatedAttackers int

    // track how many units were killed when fleeing, so the number
    // can be reported after combands ends
    DiedWhileFleeing int

    Turn Team
    CurrentTurn int

    Log []CombatLogEvent
    Observer CombatObservers

    // cached location of city wall gate
    CityWallGate image.Point

    // incremented for each unit that is inside the town area (when fighting in a town)
    CollateralDamage int

    // enchantments applied to the battle usually by a town enchantment (heavenly light or cloud of darkness)
    // these enchantments cannot be removed by Dispel, but can be removed by Disenchant Area/True
    GlobalEnchantments []data.CombatEnchantment
}

func MakeCombatModel(allSpells spellbook.Spells, defendingArmy *Army, attackingArmy *Army, landscape CombatLandscape, plane data.Plane, zone ZoneType, influence data.MagicType, overworldX int, overworldY int, events chan CombatEvent) *CombatModel {
    model := &CombatModel{
        Turn: TeamDefender,
        Plane: plane,
        SelectedUnit: nil,
        Tiles: makeTiles(30, 30, landscape, plane, zone),
        TurnAttacker: 0,
        TurnDefender: 0,
        AttackingArmy: attackingArmy,
        DefendingArmy: defendingArmy,
        CurrentTurn: 0,
        Events: events,
        Zone: zone,
        Influence: influence,
    }

    model.Initialize(allSpells, overworldX, overworldY)

    model.NextTurn()
    model.SelectedUnit = model.ChooseNextUnit(TeamDefender)

    return model
}

// do N rolls (n=strength) where each roll has 'chance' success. return number of successful rolls
func ComputeRoll(strength int, chance int) int {
    out := 0

    for range strength {
        if rand.N(100) < chance {
            out += 1
        }
    }

    return out
}

// distance -> cost multiplier:
// 0 -> 0.5
// <=5 -> 1
// <=10 -> 1.5
// <=15 -> 2
// <=20 -> 2.5
// >20 or other plane -> 3.0
func computeRangeToFortress(plane data.Plane, x int, y int, player ArmyPlayer) fraction.Fraction {
    // channeler menas the maximum range is 1.0
    minRange := fraction.FromInt(3)
    if player.GetWizard().RetortEnabled(data.RetortChanneler) {
        minRange = fraction.FromInt(1)
    }

    fortressCity := player.FindFortressCity()
    if fortressCity == nil || fortressCity.Plane != plane {
        return fraction.FromInt(3).Min(minRange)
    }

    distance := fortressCity.TileDistance(x, y)
    switch {
        case distance == 0: return fraction.Make(1, 2)
        case distance <= 5: return fraction.FromInt(1)
        case distance <= 10: return fraction.Make(3, 2).Min(minRange)
        case distance <= 15: return fraction.Make(2, 1).Min(minRange)
        case distance <= 20: return fraction.Make(5, 2).Min(minRange)
        default: return fraction.Make(3, 1).Min(minRange)
    }
}

func (model *CombatModel) Initialize(allSpells spellbook.Spells, overworldX int, overworldY int) {
    model.AttackingArmy.ManaPool = min(model.AttackingArmy.Player.GetMana(), model.AttackingArmy.Player.ComputeCastingSkill())
    model.DefendingArmy.ManaPool = min(model.DefendingArmy.Player.GetMana(), model.DefendingArmy.Player.ComputeCastingSkill())

    model.DefendingArmy.Range = computeRangeToFortress(model.Plane, overworldX, overworldY, model.DefendingArmy.Player)
    model.AttackingArmy.Range = computeRangeToFortress(model.Plane, overworldX, overworldY, model.AttackingArmy.Player)

    for _, unit := range model.DefendingArmy.units {
        unit.Model = model
        unit.Team = TeamDefender
        unit.RangedAttacks = unit.Unit.GetRangedAttacks()
        unit.InitializeSpells(allSpells, model.DefendingArmy.Player)
        model.Tiles[unit.Y][unit.X].Unit = unit
    }

    for _, unit := range model.AttackingArmy.units {
        unit.Model = model
        unit.Team = TeamAttacker
        unit.RangedAttacks = unit.Unit.GetRangedAttacks()
        unit.InitializeSpells(allSpells, model.AttackingArmy.Player)
        model.Tiles[unit.Y][unit.X].Unit = unit
    }
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

/* choose a unit from the given team such that
 * the unit's LastTurn is less than the current turn
 */
func (model *CombatModel) ChooseNextUnit(team Team) *ArmyUnit {

    switch team {
        case TeamAttacker:
            for i := 0; i < len(model.AttackingArmy.units); i++ {
                unit := model.AttackingArmy.units[model.TurnAttacker]
                model.TurnAttacker = (model.TurnAttacker + 1) % len(model.AttackingArmy.units)

                if unit.IsAsleep() || unit.ConfusionAction == ConfusionActionDoNothing {
                    unit.LastTurn = model.CurrentTurn
                }

                if unit.LastTurn < model.CurrentTurn {

                    // spend a turn to remove the web
                    if unit.IsWebbed() {
                        unit.ProcessWeb()
                        continue
                    }

                    unit.Paths = make(map[image.Point]pathfinding.Path)
                    return unit
                }
            }
            return nil
        case TeamDefender:
            for i := 0; i < len(model.DefendingArmy.units); i++ {
                unit := model.DefendingArmy.units[model.TurnDefender]
                model.TurnDefender = (model.TurnDefender + 1) % len(model.DefendingArmy.units)

                if unit.IsAsleep() || unit.ConfusionAction == ConfusionActionDoNothing {
                    unit.LastTurn = model.CurrentTurn
                }

                if unit.LastTurn < model.CurrentTurn {
                    if unit.IsWebbed() {
                        unit.ProcessWeb()
                        continue
                    }

                    unit.Paths = make(map[image.Point]pathfinding.Path)
                    return unit
                }
            }
            return nil
    }

    return nil
}

func (model *CombatModel) NextTurn() {
    model.CurrentTurn += 1

    model.DefendingArmy.Casted = false
    model.AttackingArmy.Casted = false

    defenderLeakMana := false

    if model.IsEnchantmentActive(data.CombatEnchantmentManaLeak, TeamAttacker) {
        model.DefendingArmy.ManaPool = max(0, model.DefendingArmy.ManaPool - 5)
        model.DefendingArmy.Player.UseMana(5)
        defenderLeakMana = true
    }

    defenderTerror := model.IsEnchantmentActive(data.CombatEnchantmentTerror, TeamAttacker)
    defenderWrack := model.IsEnchantmentActive(data.CombatEnchantmentWrack, TeamAttacker)

    /* reset movement */
    for _, unit := range model.DefendingArmy.units {
        unit.ResetTurnData()

        if defenderLeakMana {
            // FIXME: magic ranged attacks should go down by 1 as well
            unit.CastingSkill = max(0, unit.CastingSkill - 5)
        }

        if defenderTerror {
            if rand.N(10) + 1 > unit.GetResistance() + 1 {
                unit.MovesLeft = fraction.Zero()
            }
        }

        if unit.HasAbility(data.AbilityRegeneration) {
            unit.Heal(1)
        }

        if defenderWrack {
            damage := 0
            for range unit.Figures() {
                if rand.N(10) + 1 > unit.GetResistance() + 1 {
                    damage += 1
                }
            }
            model.MakeGibs(unit, unit.TakeDamage(damage, DamageNormal))
            if unit.GetHealth() <= 0 {
                model.KillUnit(unit)
            }
        }
    }

    attackerLeakMana := false

    if model.IsEnchantmentActive(data.CombatEnchantmentManaLeak, TeamDefender) {
        model.AttackingArmy.ManaPool = max(0, model.AttackingArmy.ManaPool - 5)
        model.AttackingArmy.Player.UseMana(5)
        attackerLeakMana = true
    }

    attackerTerror := model.IsEnchantmentActive(data.CombatEnchantmentTerror, TeamDefender)
    attackerWrack := model.IsEnchantmentActive(data.CombatEnchantmentWrack, TeamDefender)

    for _, unit := range model.AttackingArmy.units {
        // increase collateral damage to the town for each unit that is within the town area
        if model.InsideTown(unit.X, unit.Y) {
            model.CollateralDamage += 1
        }

        unit.ResetTurnData()

        if attackerLeakMana {
            // FIXME: magic ranged attacks should go down by 1 as well
            unit.CastingSkill = max(0, unit.CastingSkill - 5)
        }

        if attackerTerror {
            if rand.N(10) + 1 > unit.GetResistance() + 1 {
                unit.MovesLeft = fraction.Zero()
            }
        }

        if unit.HasAbility(data.AbilityRegeneration) {
            unit.Heal(1)
        }

        if attackerWrack {
            damage := 0
            for range unit.Figures() {
                if rand.N(10) + 1 > unit.GetResistance() + 1 {
                    damage += 1
                }
            }
            model.MakeGibs(unit, unit.TakeDamage(damage, DamageNormal))
            if unit.GetHealth() <= 0 {
                model.KillUnit(unit)
            }
        }
    }
}

func (model *CombatModel) IsTeamAlive(team Team) bool {
    switch team {
        case TeamDefender: return len(model.DefendingArmy.units) > 0
        case TeamAttacker: return len(model.AttackingArmy.units) > 0
    }

    return false
}

// one side finished its turn
func (model *CombatModel) FinishTurn(team Team) {
    if model.IsTeamAlive(team) && model.IsEnchantmentActive(data.CombatEnchantmentCallLightning, team) {
        switch team {
            case TeamDefender: model.doCallLightning(model.AttackingArmy)
            case TeamAttacker: model.doCallLightning(model.DefendingArmy)
        }
    }
}

func (model *CombatModel) doCallLightning(army *Army) {
    /* not sure about this
    units := slices.DeleteFunc(slices.Clone(army.Units), func (unit *ArmyUnit) bool {
        // even though call lightning is a nature enchantment, the lightning bolts themselves are chaos in nature
        return unit.IsMagicImmune(data.ChaosMagic)
    })
    */

    if len(army.units) == 0 {
        return
    }

    count := rand.N(3) + 3

    for range count {
        choice := rand.N(len(army.units))

        model.Events <- &CombatEventCreateLightningBolt{
            Target: army.units[choice],
            Strength: 8,
        }
    }
}

func (model *CombatModel) computePath(x1 int, y1 int, x2 int, y2 int, canTraverseWall bool, isFlying bool) (pathfinding.Path, bool) {

    tileEmpty := func (x int, y int) bool {
        return model.GetUnit(x, y) == nil
    }

    // FIXME: take into account mud, hills, other types of terrain obstacles
    tileCost := func (x1 int, y1 int, x2 int, y2 int) float64 {

        if x2 < 0 || y2 < 0 || y2 >= len(model.Tiles) || x2 >= len(model.Tiles[y2]) {
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

                if x >= 0 && y >= 0 && y < len(model.Tiles) && x < len(model.Tiles[y]) {

                    canMove := tileEmpty(x, y)

                    // towers are impassable to all units
                    if canMove && model.ContainsWallTower(x, y) {
                        canMove = false
                    }

                    // if the unit is not flying, then it can't move through a cloud tile
                    if canMove && !isFlying && model.IsCloudTile(x, y) != model.IsCloudTile(cx, cy) {
                        canMove = false
                    }

                    // can't move through a city wall
                    if canMove && !canTraverseWall && model.InsideCityWall(cx, cy) != model.InsideCityWall(x, y) {
                        // FIXME: handle destroyed walls here
                        if !model.IsCityWallGate(x, y) {
                            canMove = false
                        }
                    }

                    // ignore non-empty tiles entirely
                    if canMove {
                        out = append(out, image.Pt(x, y))
                    }
                }
            }
        }
        return out
    }

    return pathfinding.FindPath(image.Pt(x1, y1), image.Pt(x2, y2), 50, tileCost, neighbors, pathfinding.PointEqual)
}

/* return a valid path that the given unit can take to reach tile position x, y
 * this caches the path such that the next call to FindPath() will return the same path without computing it
 */
func (model *CombatModel) FindPath(unit *ArmyUnit, x int, y int) (pathfinding.Path, bool) {
    end := image.Pt(x, y)
    path, ok := unit.Paths[end]
    if ok {
        return path, len(path) > 0
    }

    path, ok = model.computePath(unit.X, unit.Y, x, y, unit.CanTraverseWall(), unit.IsFlying())
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

func (model *CombatModel) GetUnit(x int, y int) *ArmyUnit {
    if x >= 0 && y >= 0 && y < len(model.Tiles) && x < len(model.Tiles[0]) {
        return model.Tiles[y][x].Unit
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

func (model *CombatModel) CanMoveTo(unit *ArmyUnit, x int, y int) bool {

    if unit.CanTeleport() {
        return distance(float64(unit.X), float64(unit.Y), float64(x), float64(y)) <= 10
    }

    _, ok := model.FindPath(unit, x, y)
    return ok
}

func (model *CombatModel) GetObserver() CombatObserver {
    return &model.Observer
}

// do a dispel roll on all enchantments owned by the other player
// presumption: disenchantStrength should already have the runemaster bonus applied
func (model *CombatModel) DoDisenchantArea(allSpells spellbook.Spells, caster ArmyPlayer, disenchantStrength int) {
    targetArmy := model.GetOppositeArmyForPlayer(caster)

    // enemy combat enchantments
    var removedEnchantments []data.CombatEnchantment
    for _, enchantment := range targetArmy.Enchantments {
        spell := allSpells.FindByName(enchantment.SpellName())
        cost := spell.Cost(false)
        dispellChance := spellbook.ComputeDispelChance(disenchantStrength, cost, spell.Magic, targetArmy.Player.GetWizard())
        if spellbook.RollDispelChance(dispellChance) {
            removedEnchantments = append(removedEnchantments, enchantment)
        }
    }

    for _, enchantment := range removedEnchantments {
        targetArmy.RemoveEnchantment(enchantment)
    }

    // enemy unit enchantments
    for _, unit := range targetArmy.units {
        if unit.GetHealth() > 0 {
            model.DoDisenchantUnit(allSpells, unit, targetArmy.Player, disenchantStrength)
        }
    }

    // friendly unit curses
    playerArmy := model.GetArmyForPlayer(caster)
    for _, unit := range playerArmy.units {
        if unit.GetHealth() > 0 {
            model.DoDisenchantUnitCurses(allSpells, unit, targetArmy.Player, disenchantStrength)
        }
    }
}

// only removes enchantments (not curses)
func (model *CombatModel) DoDisenchantUnit(allSpells spellbook.Spells, unit *ArmyUnit, owner ArmyPlayer, disenchantStrength int) {
    var removedEnchantments []data.UnitEnchantment

    choices := append(unit.Unit.GetEnchantments(), unit.Enchantments...)

    // if the unit has spell lock then only that spell can be dispelled. once it is dispelled then the
    // other choices become valid targets
    if unit.HasEnchantment(data.UnitEnchantmentSpellLock) {
        choices = []data.UnitEnchantment{data.UnitEnchantmentSpellLock}
    }

    for _, enchantment := range choices {
        spell := allSpells.FindByName(enchantment.SpellName())
        cost := spell.Cost(false)
        // spell lock has a unique cost for the purposes of dispelling
        if enchantment == data.UnitEnchantmentSpellLock {
            cost = 150
        }
        dispellChance := spellbook.ComputeDispelChance(disenchantStrength, cost, spell.Magic, owner.GetWizard())
        if spellbook.RollDispelChance(dispellChance) {
            removedEnchantments = append(removedEnchantments, enchantment)
        }
    }

    for _, enchantment := range removedEnchantments {
        unit.RemoveEnchantment(enchantment)
    }
}

func (model *CombatModel) DoDisenchantUnitCurses(allSpells spellbook.Spells, unit *ArmyUnit, owner ArmyPlayer, disenchantStrength int) {
    var removedEnchantments []data.UnitEnchantment
    for _, enchantment := range unit.GetCurses() {
        spell := allSpells.FindByName(enchantment.SpellName())
        cost := spell.Cost(false)
        dispellChance := spellbook.ComputeDispelChance(disenchantStrength, cost, spell.Magic, owner.GetWizard())
        if spellbook.RollDispelChance(dispellChance) {
            removedEnchantments = append(removedEnchantments, enchantment)
        }
    }

    // if the unit had Creature Bind then when it is dispelled the unit should be moved to the other army
    swapArmy := false
    for _, enchantment := range removedEnchantments {
        // these enchantments will basically never be dispelled because if a wizard loses control of a unit and casts
        // disenchant on the unit, then only positive enchantments will be removed, not curses
        if enchantment == data.UnitCurseCreatureBinding || enchantment == data.UnitCursePossession {
            swapArmy = true
        }

        unit.RemoveCurse(enchantment)

        // remove confusion effects
        if enchantment == data.UnitCurseConfusion {
            unit.ConfusionAction = ConfusionActionNone
        }
    }

    if swapArmy {
        model.SwitchTeams(unit)
    }
}

func (model *CombatModel) IsEnchantmentActive(enchantment data.CombatEnchantment, team Team) bool {
    // global enchantments affect both sides no matter what
    if slices.ContainsFunc(model.GlobalEnchantments, func(check data.CombatEnchantment) bool {
        return enchantment == check
    }) {
        return true
    }

    if team == TeamEither {
        return model.DefendingArmy.HasEnchantment(enchantment) || model.AttackingArmy.HasEnchantment(enchantment)
    }

    if team == TeamDefender {
        return model.DefendingArmy.HasEnchantment(enchantment)
    } else {
        return model.AttackingArmy.HasEnchantment(enchantment)
    }
}

func (model *CombatModel) AddLogEvent(text string) {
    log.Printf(text)
    model.Log = append(model.Log, CombatLogEvent{
        Turn: model.CurrentTurn,
        Text: text,
        AbsoluteTime: time.Now(),
    })
}

func (model *CombatModel) AddProjectile(projectile *Projectile){
    model.Projectiles = append(model.Projectiles, projectile)
}

func (model *CombatModel) AddGlobalEnchantment(enchantment data.CombatEnchantment) {
    model.GlobalEnchantments = append(model.GlobalEnchantments, enchantment)
}

func (model *CombatModel) AddEnchantment(player ArmyPlayer, enchantment data.CombatEnchantment) bool {
    if slices.Contains(model.GlobalEnchantments, enchantment) {
        return false
    }

    if player == model.DefendingArmy.Player {
        return model.DefendingArmy.AddEnchantment(enchantment)
    } else {
        return model.AttackingArmy.AddEnchantment(enchantment)
    }
}

func (model *CombatModel) summonUnit(player ArmyPlayer, x int, y int, unit units.Unit, facing units.Facing, summoned bool) *ArmyUnit {
    newUnit := model.addNewUnit(player, x, y, unit, facing, summoned)
    // this offset is chosen somewhat arbitrarily. maybe there is a better way to compute it, possibly based on the unit's image size?
    newUnit.SetHeight(-18)
    model.Events <- &CombatEventSummonUnit{
        Unit: newUnit,
    }
    return newUnit
}

func (model *CombatModel) addNewUnit(player ArmyPlayer, x int, y int, unit units.Unit, facing units.Facing, summoned bool) *ArmyUnit {
    newUnit := ArmyUnit{
        Unit: units.MakeOverworldUnitFromUnit(unit, 0, 0, model.Plane, player.GetWizard().Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()),
        Facing: facing,
        Moving: false,
        X: x,
        Y: y,
        MovesLeft: fraction.FromInt(unit.MovementSpeed),
        LastTurn: model.CurrentTurn-1,
        Summoned: summoned,
    }

    newUnit.Model = model
    newUnit.Unit.SetEnchantmentProvider(&newUnit)
    model.Cleanups = append(model.Cleanups, func (){
        newUnit.Unit.SetEnchantmentProvider(nil)
    })

    model.Tiles[y][x].Unit = &newUnit

    if player == model.DefendingArmy.Player {
        newUnit.Team = TeamDefender
        model.DefendingArmy.units = append(model.DefendingArmy.units, &newUnit)
    } else {
        newUnit.Team = TeamAttacker
        model.AttackingArmy.units = append(model.AttackingArmy.units, &newUnit)
    }

    return &newUnit
}

/* makes a 5x5 square of tiles have mud on them
 */
func (model *CombatModel) CreateEarthToMud(centerX int, centerY int){
    // log.Printf("Create earth to mud at %v, %v", centerX, centerY)

    for x := centerX - 2; x <= centerX + 2; x++ {
        for y := centerY - 2; y <= centerY + 2; y++ {
            if x >= 0 && x < len(model.Tiles[0]) && y >= 0 && y < len(model.Tiles) {
                model.Tiles[y][x].Mud = true
            }
        }
    }
}

type MapSide int
const (
    MapSideAttacker MapSide = iota
    MapSideDefender
    MapSideMiddle
)

func (model *CombatModel) GetSideForPlayer(player ArmyPlayer) MapSide {
    if player == model.DefendingArmy.Player {
        return MapSideDefender
    }

    return MapSideAttacker
}

// returns true if the given coordinates are on the given side
func (model *CombatModel) IsOnSide(x int, y int, side MapSide) bool {
    // FIXME: verify these coordinates
    switch side {
        case MapSideAttacker: return y >= 15
        case MapSideDefender: return y <= 12
    }

    return true
}

func (model *CombatModel) FindEmptyTile(side MapSide) (int, int, error) {

    middleX := len(model.Tiles[0]) / 2
    middleY := len(model.Tiles) / 2

    switch side {
        case MapSideMiddle: // already set
        case MapSideAttacker:
            middleX = TownCenterX - 2
            middleY = 17
        case MapSideDefender:
            middleX = TownCenterX - 2
            middleY = 10
    }

    distance := 3
    tries := 0
    for tries < 100 {
        x := middleX + rand.N(distance) - distance/2
        y := middleY + rand.N(distance) - distance/2

        if x >= 0 && x < len(model.Tiles[0]) && y >= 0 && y < len(model.Tiles) && model.GetUnit(x, y) == nil {
            return x, y, nil
        }

        distance += 1
        if distance > len(model.Tiles) * 2 {
            distance = len(model.Tiles) * 2
        }
    }

    return -1, -1, fmt.Errorf("unable to find a free tile")
}

func (model *CombatModel) DoneTurn() {
    model.SelectedUnit.LastTurn = model.CurrentTurn
    model.NextUnit()
}

func (model *CombatModel) NextUnit() {

    var nextChoice *ArmyUnit
    for range 2 {
        // find a unit on the same team
        nextChoice = model.ChooseNextUnit(model.Turn)
        if nextChoice == nil {
            model.FinishTurn(model.Turn)
            // if there are no available units then the team must be out of moves, so try the next team
            model.Turn = oppositeTeam(model.Turn)
            nextChoice = model.ChooseNextUnit(model.Turn)

            if nextChoice == nil {
                // if the other team still has nothing available then the entire turn has finished
                // so go to the next turn and try again
                model.NextTurn()
                model.SelectedUnit = nil
            }
        }

        // found something so break the loop
        if nextChoice != nil {
            break
        }
    }

    model.SelectedUnit = nextChoice
}

// return true if x,y is within the bounds of the enclosed wall of fire space
func (model *CombatModel) InsideWallOfFire(x int, y int) bool {
    if x < 0 || y < 0 || y >= len(model.Tiles) || x >= len(model.Tiles[0]) {
        return false
    }

    return model.Tiles[y][x].InsideFire
}

func (model *CombatModel) InsideWallOfDarkness(x int, y int) bool {
    if x < 0 || y < 0 || y >= len(model.Tiles) || x >= len(model.Tiles[0]) {
        return false
    }

    return model.Tiles[y][x].InsideDarkness
}

func (model *CombatModel) InsideTown(x int, y int) bool {
    if x < 0 || y < 0 || y >= len(model.Tiles) || x >= len(model.Tiles[0]) {
        return false
    }

    return model.Tiles[y][x].InsideTown
}

func (model *CombatModel) InsideCityWall(x int, y int) bool {
    if x < 0 || y < 0 || y >= len(model.Tiles) || x >= len(model.Tiles[0]) {
        return false
    }

    return model.Tiles[y][x].InsideWall
}

func (model *CombatModel) ContainsWall(x int, y int) bool {
    if x < 0 || y < 0 || y >= len(model.Tiles) || x >= len(model.Tiles[0]) {
        return false
    }

    wall := model.Tiles[y][x].Wall
    if wall != nil && wall.Size() > 0 {
        return true
    }

    return false
}

// a wall tower is a wall with two sides
func (model *CombatModel) ContainsWallTower(x int, y int) bool {
    if x < 0 || y < 0 || y >= len(model.Tiles) || x >= len(model.Tiles[0]) {
        return false
    }

    wall := model.Tiles[y][x].Wall
    if wall != nil && wall.Size() == 2 {
        return true
    }

    return false
}

func (model *CombatModel) IsCityWallGate(x int, y int) bool {
    if x < 0 || y < 0 || y >= len(model.Tiles) || x >= len(model.Tiles[0]) {
        return false
    }

    wall := model.Tiles[y][x].Wall

    if wall != nil && wall.Contains(WallKindGate) {
        return true
    }

    return false
}

func (model *CombatModel) GetCityGateCoordinates() (int, int) {

    if !model.CityWallGate.Eq(image.Point{}) {
        return model.CityWallGate.X, model.CityWallGate.Y
    }

    for y := 0; y < len(model.Tiles); y++ {
        for x := 0; x < len(model.Tiles[y]); x++ {
            if model.IsCityWallGate(x, y) {
                model.CityWallGate = image.Pt(x, y)
                return x, y
            }
        }
    }

    return -1, -1
}

func (model *CombatModel) InsideAnyWall(x int, y int) bool {
    return model.InsideWallOfFire(x, y) || model.InsideWallOfDarkness(x, y) || model.InsideCityWall(x, y)
}

func (model *CombatModel) IsCloudTile(x int, y int) bool {
    if x < 0 || y < 0 || y >= len(model.Tiles) || x >= len(model.Tiles[0]) {
        return false
    }

    return model.Tiles[y][x].Flying
}

func distance(x1 float64, y1 float64, x2 float64, y2 float64) float64 {
    xDiff := x2 - x1
    yDiff := y2 - y1
    return math.Sqrt(xDiff * xDiff + yDiff * yDiff)
}

func (model *CombatModel) UpdateProjectiles(counter uint64) bool {
    animationSpeed := uint64(5)

    alive := len(model.Projectiles) > 0

    var projectilesOut []*Projectile
    for _, projectile := range model.Projectiles {
        keep := false
        if projectile.Exploding {
            projectile.Exploding = true
            keep = true
            if counter % animationSpeed == 0 && !projectile.Explode.Next() {
                keep = false

                if projectile.Target != nil && projectile.Effect != nil {
                    projectile.Effect(projectile.Target)
                }
            }
        } else {
            previousDistance := distance(projectile.X, projectile.Y, projectile.TargetX, projectile.TargetY)

            projectile.X += math.Cos(projectile.Angle) * projectile.Speed
            projectile.Y += math.Sin(projectile.Angle) * projectile.Speed

            newDistance := distance(projectile.X, projectile.Y, projectile.TargetX, projectile.TargetY)

            // when the distance between the projectile and its target increases then we know the projectile has gone too far,
            // so it should explode
            if newDistance > previousDistance {
                // possibly just set projectile.X/Y to TargetX/Y?
                projectile.Exploding = true
            }

            if counter % animationSpeed == 0 {
                projectile.Animation.Next()
            }
            keep = true
        }

        if keep {
            projectilesOut = append(projectilesOut, projectile)
        }
    }

    model.Projectiles = projectilesOut

    return alive
}

func (model *CombatModel) doBreathAttack(attacker *ArmyUnit, defender *ArmyUnit) ([]func(), bool) {
    damage := []func(){}
    hit := false

    if attacker.HasAbility(data.AbilityFireBreath) {
        strength := int(attacker.GetAbilityValue(data.AbilityFireBreath))
        hit = true

        // FIXME: the tohit value here should not include the Holy Weapon bonus

        damage = append(damage, func(){
            fireDamage := 0
            lost := 0
            // one breath attack per figure
            for range attacker.Figures() {
                attackerDamage := ComputeRoll(strength, attacker.GetToHitMelee(defender))
                moreDamage, moreLost := ApplyDamage(defender, defender.ReduceInvulnerability(attackerDamage), units.DamageFire, attacker.GetDamageSource(), DamageModifiers{Magic: data.ChaosMagic})
                fireDamage += moreDamage
                lost += moreLost
            }
            model.MakeGibs(defender, lost)
            model.AddLogEvent(fmt.Sprintf("%v uses fire breath on %v for %v damage", attacker.Unit.GetName(), defender.Unit.GetName(), fireDamage))
            // damage += fireDamage
            model.Observer.FireBreathAttack(attacker, defender, fireDamage)
        })
    }

    if attacker.HasAbility(data.AbilityLightningBreath) {
        strength := int(attacker.GetAbilityValue(data.AbilityLightningBreath))
        hit = true

        damage = append(damage, func(){
            lightningDamage := 0
            lost := 0
            for range attacker.Figures() {
                attackerDamage := ComputeRoll(strength, attacker.GetToHitMelee(defender))
                moreLightningDamage, mostLost := ApplyDamage(defender, defender.ReduceInvulnerability(attackerDamage), units.DamageRangedMagical, attacker.GetDamageSource(), DamageModifiers{ArmorPiercing: true, Magic: data.ChaosMagic})
                lightningDamage += moreLightningDamage
                lost += mostLost
            }
            model.MakeGibs(defender, lost)
            model.AddLogEvent(fmt.Sprintf("%v uses lightning breath on %v for %v damage", attacker.Unit.GetName(), defender.Unit.GetName(), lightningDamage))
            // damage += lightningDamage
            model.Observer.LightningBreathAttack(attacker, defender, lightningDamage)
        })
    }

    return damage, hit
}

// returns normal damage, irreversable damage, and whether or not the attack hit
func (model *CombatModel) doGazeAttack(attacker *ArmyUnit, defender *ArmyUnit) (int, int, bool) {
    // FIXME: take into account the attack strength of the unit, and modifiers from spells/magic nodes

    irreversableDamage := 0
    damage := 0
    hit := false
    if attacker.HasAbility(data.AbilityStoningGaze) {
        if !defender.HasAbility(data.AbilityStoningImmunity) && !defender.HasAbility(data.AbilityMagicImmunity) {
            resistance := int(attacker.GetAbilityValue(data.AbilityStoningGaze))

            stoneDamage := 0

            for range defender.Figures() {
                if rand.N(10) + 1 > GetResistanceFor(defender, data.NatureMagic) - resistance {
                    stoneDamage += defender.GetHitPoints()
                }
            }

            irreversableDamage += stoneDamage
            hit = true

            model.Observer.StoneGazeAttack(attacker, defender, stoneDamage)

            model.AddLogEvent(fmt.Sprintf("%v uses stone gaze on %v for %v damage", attacker.Unit.GetName(), defender.Unit.GetName(), stoneDamage))
        }
    }

    if attacker.HasAbility(data.AbilityDeathGaze) {
        if !defender.HasAbility(data.AbilityDeathImmunity) && !defender.HasAbility(data.AbilityMagicImmunity) {
            resistance := int(attacker.GetAbilityValue(data.AbilityDeathGaze))

            deathDamage := 0

            for range defender.Figures() {
                if rand.N(10) + 1 > GetResistanceFor(defender, data.DeathMagic) - resistance {
                    deathDamage += defender.GetHitPoints()
                }
            }

            damage += deathDamage
            hit = true

            model.Observer.DeathGazeAttack(attacker, defender, deathDamage)

            model.AddLogEvent(fmt.Sprintf("%v uses death gaze on %v for %v damage", attacker.Unit.GetName(), defender.Unit.GetName(), deathDamage))
        }
    }

    if attacker.HasAbility(data.AbilityDoomGaze) {
        doomDamage := int(attacker.GetAbilityValue(data.AbilityDoomGaze))
        damage += doomDamage
        hit = true
        model.Observer.DoomGazeAttack(attacker, defender, doomDamage)
        model.AddLogEvent(fmt.Sprintf("%v uses doom gaze on %v for %v damage", attacker.Unit.GetName(), defender.Unit.GetName(), doomDamage))
    }

    return damage, irreversableDamage, hit
}

func (model *CombatModel) doThrowAttack(attacker *ArmyUnit, defender *ArmyUnit) (int, bool) {
    if attacker.HasAbility(data.AbilityThrown) {
        strength := int(attacker.GetAbilityValue(data.AbilityThrown))
        damage := 0
        for range attacker.Figures() {
            if rand.N(100) < attacker.GetToHitMelee(defender) {
                // damage += defender.ApplyDamage(strength, units.DamageThrown, false)
                damage += defender.ReduceInvulnerability(strength)
            }
        }

        return damage, true
    }

    return 0, false
}

func (model *CombatModel) immolationDamage(attacker *ArmyUnit, defender *ArmyUnit) int {
    if attacker.HasAbility(data.AbilityImmolation) {
        damage := 4
        model.Observer.ImmolationAttack(attacker, defender, damage)
        return damage
    }

    return 0
}

func (model *CombatModel) doTouchAttack(attacker *ArmyUnit, defender *ArmyUnit, fearFigure int) []func() {
    damageFuncs := []func(){}

    if attacker.HasAbility(data.AbilityPoisonTouch) && !defender.HasAbility(data.AbilityPoisonImmunity) {
        damage := 0
        for range int(attacker.GetAbilityValue(data.AbilityPoisonTouch)) {
            if rand.N(10) + 1 > defender.GetResistance() {
                damage += 1
            }
        }

        damageType := DamageNormal
        if attacker.HasAbility(data.AbilityCreateUndead) {
            damageType = DamageUndead
        }

        damageFuncs = append(damageFuncs, func(){
            model.MakeGibs(defender, defender.TakeDamage(damage, damageType))
            model.Observer.PoisonTouchAttack(attacker, defender, damage)
            model.AddLogEvent(fmt.Sprintf("%v is poisoned for %v damage. HP now %v", defender.Unit.GetName(), damage, defender.GetHealth()))
        })
    }

    if attacker.HasAbility(data.AbilityLifeSteal) {
        if !defender.HasAbility(data.AbilityDeathImmunity) && !defender.HasAbility(data.AbilityMagicImmunity) {
            modifier := int(attacker.GetAbilityValue(data.AbilityLifeSteal))
            // if vampiric, modifier will just be 0
            damage := 0
            defenderResistance := GetResistanceFor(defender, data.DeathMagic)

            for range attacker.Figures() - fearFigure {
                more := rand.N(10) + 1 - (defenderResistance + modifier)
                if more > 0 {
                    damage += more
                }
            }

            if damage > 0 {
                // cannot steal more life than the target has
                damage = min(damage, defender.GetHealth())

                damageFuncs = append(damageFuncs, func(){
                    model.MakeGibs(defender, defender.TakeDamage(damage, DamageUndead))
                    attacker.Heal(damage)
                    model.AddLogEvent(fmt.Sprintf("%v steals %v life from %v. HP now %v", attacker.Unit.GetName(), damage, defender.Unit.GetName(), defender.GetHealth()))

                    model.Observer.LifeStealTouchAttack(attacker, defender, damage)
                })
            }
        }
    }

    if attacker.HasAbility(data.AbilityStoningTouch) {
        if !defender.HasAbility(data.AbilityStoningImmunity) && !defender.HasAbility(data.AbilityMagicImmunity) {
            damage := 0

            defenderResistance := GetResistanceFor(defender, data.NatureMagic)

            modifier := int(attacker.GetAbilityValue(data.AbilityStoningTouch))

            // for each failed resistance roll, the defender takes damage equal to one figure's hit points
            for range attacker.Figures() - fearFigure {
                if rand.N(10) + 1 > defenderResistance - modifier {
                    damage += defender.GetHitPoints()
                }
            }

            damageFuncs = append(damageFuncs, func(){
                model.MakeGibs(defender, defender.TakeDamage(damage, DamageIrreversable))

                model.AddLogEvent(fmt.Sprintf("%v turns %v to stone for %v damage. HP now %v", attacker.Unit.GetName(), defender.Unit.GetName(), damage, defender.GetHealth()))

                model.Observer.StoningTouchAttack(attacker, defender, damage)
            })
        }
    }

    if attacker.HasAbility(data.AbilityDispelEvil) {
        immune := true

        if defender.Unit.GetRace() == data.RaceFantastic {
            if defender.GetRealm() == data.ChaosMagic || defender.GetRealm() == data.DeathMagic {
                immune = false
            }
        }

        if defender.HasAbility(data.AbilityMagicImmunity) {
            immune = true
        }

        if !immune {
            damage := 0

            defenderResistance := GetResistanceFor(defender, data.LifeMagic)
            if defender.Unit.IsUndead() {
                defenderResistance -= 9
            } else {
                defenderResistance -= 4
            }

            for range attacker.Figures() - fearFigure {
                if rand.N(10) + 1 > defenderResistance {
                    damage += defender.GetHitPoints()
                }
            }

            damageFuncs = append(damageFuncs, func(){
                model.MakeGibs(defender, defender.TakeDamage(damage, DamageIrreversable))
                model.AddLogEvent(fmt.Sprintf("%v dispels evil from %v for %v damage. HP now %v", attacker.Unit.GetName(), defender.Unit.GetName(), damage, defender.GetHealth()))

                model.Observer.DispelEvilTouchAttack(attacker, defender, damage)
            })
        }
    }

    if attacker.HasAbility(data.AbilityDeathTouch) {
        if !defender.HasAbility(data.AbilityDeathImmunity) && !defender.HasAbility(data.AbilityMagicImmunity) {
            damage := 0
            defenderResistance := GetResistanceFor(defender, data.DeathMagic)
            modifier := 3

            for range attacker.Figures() - fearFigure {
                if rand.N(10) + 1 > defenderResistance - modifier {
                    damage += defender.GetHitPoints()
                }
            }

            damageFuncs = append(damageFuncs, func(){
                model.MakeGibs(defender, defender.TakeDamage(damage, DamageNormal))

                model.AddLogEvent(fmt.Sprintf("%v uses death touch on %v for %v damage. HP now %v", attacker.Unit.GetName(), defender.Unit.GetName(), damage, defender.GetHealth()))

                model.Observer.DeathTouchAttack(attacker, defender, damage)
            })
        }
    }

    if attacker.Unit.HasItemAbility(data.ItemAbilityDestruction) {
        if !defender.HasAbility(data.AbilityMagicImmunity) {
            defenderResistance := GetResistanceFor(defender, data.ChaosMagic)

            damage := 0
            for range attacker.Figures() - fearFigure {
                if rand.N(10) + 1 > defenderResistance {
                    damage += defender.GetHitPoints()
                }
            }

            damageFuncs = append(damageFuncs, func(){
                model.MakeGibs(defender, defender.TakeDamage(damage, DamageIrreversable))
                model.AddLogEvent(fmt.Sprintf("%v uses destruction on %v for %v damage. HP now %v", attacker.Unit.GetName(), defender.Unit.GetName(), damage, defender.GetHealth()))

                model.Observer.DestructionAttack(attacker, defender, damage)
            })
        }
    }

    return damageFuncs
}

// if defender is in city wall but attacker is outside, then defense is
//  1 if defender is not adjacent a wall
//  3 if defender is adjacent to a wall
func (model *CombatModel) ComputeWallDefense(attacker *ArmyUnit, defender *ArmyUnit) int {
    if model.InsideCityWall(defender.X, defender.Y) && !model.InsideCityWall(attacker.X, attacker.Y) {

        wall := model.Tiles[defender.Y][defender.X].Wall
        if wall != nil {
            if wall.Contains(WallKindGate) {
                return 1
            }

            return 3
        }

        return 1
    }

    return 0
}

func (model *CombatModel) ApplyImmolationDamage(defender *ArmyUnit, immolationDamage int) {
    if immolationDamage > 0 {
        hurt, lost := ApplyAreaDamage(defender, immolationDamage, units.DamageImmolation, 0)
        model.MakeGibs(defender, lost)
        model.AddLogEvent(fmt.Sprintf("%v is immolated for %v damage. HP now %v", defender.Unit.GetName(), hurt, defender.GetHealth()))
    }
}

func (model *CombatModel) ApplyMeleeDamage(attacker *ArmyUnit, defender *ArmyUnit, damage int) {
    modifiers := DamageModifiers{
        WallDefense: model.ComputeWallDefense(attacker, defender),
        ArmorPiercing: attacker.HasAbility(data.AbilityArmorPiercing),
        Illusion: attacker.HasAbility(data.AbilityIllusion),
        NegateWeaponImmunity: attacker.CanNegateWeaponImmunity(),
        EldritchWeapon: attacker.HasEnchantment(data.UnitEnchantmentEldritchWeapon),
        DamageType: DamageNormal,
    }

    if attacker.HasAbility(data.AbilityCreateUndead) {
        modifiers.DamageType = DamageUndead
    }

    hurt, lost := ApplyDamage(defender, damage, units.DamageMeleePhysical, attacker.GetDamageSource(), modifiers)
    model.MakeGibs(defender, lost)
    model.AddLogEvent(fmt.Sprintf("%v damage roll %v, %v took %v damage. HP now %v", attacker.Unit.GetName(), damage, defender.Unit.GetName(), hurt, defender.GetHealth()))
}

func (model *CombatModel) ApplyWallOfFireDamage(defender *ArmyUnit) {
    model.ApplyImmolationDamage(defender, 5)
    model.Observer.WallOfFire(defender, 5)
}

func (model *CombatModel) canRangeAttack(attacker *ArmyUnit, defender *ArmyUnit) bool {
    if attacker == defender {
        return false
    }

    if attacker.IsAsleep() {
        return false
    }

    hasRangedAttacks := attacker.RangedAttacks > 0
    hasCastingAttacks := attacker.GetRangedAttackDamageType() == units.DamageRangedMagical && attacker.CastingSkill >= 3

    if !hasRangedAttacks && !hasCastingAttacks {
        return false
    }

    if attacker.GetRangedAttackPower() <= 0 {
        return false
    }

    if attacker.MovesLeft.LessThanEqual(fraction.FromInt(0)) {
        return false
    }

    if attacker.Team == defender.Team && attacker.ConfusionAction != ConfusionActionEnemyControl {
        return false
    }

    if defender.IsInvisible() && !attacker.HasAbility(data.AbilityIllusionsImmunity) {
        return false
    }

    // FIXME: check if defender has invisible, and attacker doesn't have illusions immunity

    if model.InsideWallOfDarkness(defender.X, defender.Y) && !model.InsideWallOfDarkness(attacker.X, attacker.Y) {
        // attacker can't target a defender inside a wall of darkness, unless the attacker has True Sight or Illusions Immunity

        if attacker.HasAbility(data.AbilityIllusionsImmunity) {
            return true
        }

        return false
    }

    return true
}

func (model *CombatModel) canMeleeAttack(attacker *ArmyUnit, defender *ArmyUnit) bool {
    if attacker == defender {
        return false
    }

    if attacker.IsAsleep() {
        return false
    }

    if attacker.MovesLeft.LessThanEqual(fraction.FromInt(0)) {
        return false
    }

    if attacker.GetMeleeAttackPower() <= 0 {
        return false
    }

    if !attacker.IsFlying() && model.IsCloudTile(attacker.X, attacker.Y) != model.IsCloudTile(defender.X, defender.Y) {
        return false
    }

    containsWall := func(x int, y int) bool {
        wall := model.Tiles[y][x].Wall
        return wall != nil && !wall.Contains(WallKindGate)
    }

    // cannot attack through a wall
    if model.InsideCityWall(attacker.X, attacker.Y) != model.InsideCityWall(defender.X, defender.Y) {
        // if the attacker normally cannot move through the wall, then they can only attack if either the attacker or defender
        // is adjacent to the gate
        if !attacker.CanTraverseWall() {
            var insideWall *ArmyUnit
            var outsideWall *ArmyUnit

            if model.InsideCityWall(attacker.X, attacker.Y) {
                insideWall = attacker
                outsideWall = defender
            } else {
                insideWall = defender
                outsideWall = attacker
            }

            // north
            if outsideWall.X == insideWall.X && outsideWall.Y + 1 == insideWall.Y {
                if containsWall(outsideWall.X, outsideWall.Y + 1) {
                    return false
                }
            }

            // north east
            if outsideWall.X + 1 == insideWall.X && outsideWall.Y + 1 == insideWall.Y {
                if containsWall(attacker.X, attacker.Y + 1) || model.ContainsWallTower(insideWall.X, insideWall.Y) {
                    return false
                }
            }

            // east
            if outsideWall.X + 1 == insideWall.X && outsideWall.Y == insideWall.Y {
                if containsWall(outsideWall.X + 1, outsideWall.Y) {
                    return false
                }
            }

            // south east
            if outsideWall.X + 1 == insideWall.X && outsideWall.Y - 1 == insideWall.Y {
                if containsWall(outsideWall.X, outsideWall.Y - 1) || model.ContainsWallTower(insideWall.X, insideWall.Y) {
                    return false
                }
            }

            // south
            if outsideWall.X == insideWall.X && outsideWall.Y - 1 == insideWall.Y {
                if containsWall(outsideWall.X, outsideWall.Y - 1) {
                    return false
                }
            }

            // south west
            if outsideWall.X - 1 == insideWall.X && outsideWall.Y - 1 == insideWall.Y {
                if containsWall(outsideWall.X, outsideWall.Y - 1) || model.ContainsWallTower(insideWall.X, insideWall.Y) {
                    return false
                }
            }

            // west
            if outsideWall.X - 1 == insideWall.X && outsideWall.Y == insideWall.Y {
                if containsWall(outsideWall.X - 1, outsideWall.Y) {
                    return false
                }
            }

            // north west
            if outsideWall.X - 1 == insideWall.X && outsideWall.Y + 1 == insideWall.Y {
                if containsWall(outsideWall.X, outsideWall.Y + 1) || model.ContainsWallTower(insideWall.X, insideWall.Y) {
                    return false
                }
            }
        }
    }

    if defender.IsFlying() && !attacker.IsFlying() {
        // a unit with Thrown can attack a flying unit
        if attacker.HasAbility(data.AbilityThrown) ||
           attacker.HasAbility(data.AbilityFireBreath) ||
           attacker.HasAbility(data.AbilityLightningBreath) {
        } else {
            return false
        }
    }

    if attacker.Team == defender.Team && attacker.ConfusionAction != ConfusionActionEnemyControl {
        return false
    }

    return true
}

func (model *CombatModel) MakeGibs(defender *ArmyUnit, lost int) {
    if lost <= 0 {
        return
    }

    event := CombatEventMakeGibs{
        Unit: defender,
        Count: lost,
    }

    select {
        case model.Events <- &event:
        default:
    }
}

/* attacker is performing a physical melee attack against defender
 */
func (model *CombatModel) meleeAttack(attacker *ArmyUnit, defender *ArmyUnit){
    // for each figure in attacker, choose a random number from 1-100, if lower than the ToHit percent then
    // add 1 damage point. do this random roll for however much the melee attack power is

    // number of attacking units in fear
    attackerFear := 0
    // number of defending units in fear
    defenderFear := 0

    doRound := func (round int) {
        switch round {
            case 0:
                attacks := 1
                if attacker.HasEnchantment(data.UnitEnchantmentHaste) {
                    attacks = 2
                }

                immolationDamage := 0
                throwDamage := 0
                damageFuncs := []func(){}

                for range attacks {
                    damage, throwHit := model.doThrowAttack(attacker, defender)
                    if throwHit {
                        throwDamage += damage
                        immolationDamage += model.immolationDamage(attacker, defender)
                        if attacker.Unit.CanTouchAttack(units.DamageMeleePhysical) {
                            damageFuncs = append(damageFuncs, model.doTouchAttack(attacker, defender, 0)...)
                        }
                    }

                    breathFuncs, breathHit := model.doBreathAttack(attacker, defender)
                    damageFuncs = append(damageFuncs, breathFuncs...)

                    if breathHit {
                        immolationDamage += model.immolationDamage(attacker, defender)
                        if attacker.Unit.CanTouchAttack(units.DamageMeleePhysical) {
                            damageFuncs = append(damageFuncs, model.doTouchAttack(attacker, defender, 0)...)
                        }
                    }
                }

                gazeDamage, gazeIrreversableDamage, hit := model.doGazeAttack(attacker, defender)
                if hit {
                    immolationDamage += model.immolationDamage(attacker, defender)
                    if attacker.Unit.CanTouchAttack(units.DamageMeleePhysical) {
                        damageFuncs = append(damageFuncs, model.doTouchAttack(attacker, defender, 0)...)
                    }
                }

                if throwDamage > 0 {
                    damage, lost := ApplyDamage(defender, throwDamage, units.DamageThrown, attacker.GetDamageSource(), DamageModifiers{
                        ArmorPiercing: attacker.HasAbility(data.AbilityArmorPiercing),
                        NegateWeaponImmunity: attacker.CanNegateWeaponImmunity(),
                        EldritchWeapon: attacker.HasEnchantment(data.UnitEnchantmentEldritchWeapon),
                    })

                    model.MakeGibs(defender, lost)

                    model.Observer.ThrowAttack(attacker, defender, damage)
                    model.AddLogEvent(fmt.Sprintf("%v throws %v at %v. HP now %v", attacker.Unit.GetName(), damage, defender.Unit.GetName(), defender.GetHealth()))
                }

                model.ApplyImmolationDamage(defender, immolationDamage)
                for _, f := range damageFuncs {
                    f()
                }

                model.MakeGibs(defender, defender.TakeDamage(gazeDamage, DamageNormal))
                model.MakeGibs(defender, defender.TakeDamage(gazeIrreversableDamage, DamageIrreversable))

            case 1:
                immolationDamage := 0
                damageFuncs := []func(){}

                gazeDamage, gazeIrreversableDamage, hit := model.doGazeAttack(defender, attacker)
                if hit {
                    immolationDamage += model.immolationDamage(defender, attacker)
                    if defender.Unit.CanTouchAttack(units.DamageMeleePhysical) {
                        damageFuncs = append(damageFuncs, model.doTouchAttack(defender, attacker, 0)...)
                    }
                }

                model.ApplyImmolationDamage(attacker, immolationDamage)
                for _, f := range damageFuncs {
                    f()
                }
                model.MakeGibs(attacker, attacker.TakeDamage(gazeDamage, DamageNormal))
                model.MakeGibs(attacker, attacker.TakeDamage(gazeIrreversableDamage, DamageIrreversable))

            case 2:

                // if attacker is outside the wall of fire and the defender is inside (or vice-versa), then both side take immolation damage.
                // if either side is flying then they do not take damage.
                // for this to be false, either both are inside the wall of fire, or both are outside.
                if model.InsideWallOfFire(defender.X, defender.Y) != model.InsideWallOfFire(attacker.X, attacker.Y) {
                    if !attacker.IsFlying() {
                        model.ApplyWallOfFireDamage(attacker)
                    }

                    if !defender.IsFlying() {
                        model.ApplyWallOfFireDamage(defender)
                    }
                }

            case 3:
                if !defender.IsAsleep() && defender.HasAbility(data.AbilityCauseFear) {
                    attackerFear = attacker.CauseFear()
                    model.AddLogEvent(fmt.Sprintf("%v causes fear in %v for %v figures", defender.Unit.GetName(), attacker.Unit.GetName(), attackerFear))
                    model.Observer.CauseFear(defender, attacker, attackerFear)
                }
            case 4:
                if attacker.HasAbility(data.AbilityFirstStrike) && !defender.HasAbility(data.AbilityNegateFirstStrike) {
                    attackerDamage, hit := attacker.ComputeMeleeDamage(defender, attackerFear, false)

                    // asleep units take full attack power as damage
                    if defender.IsAsleep() {
                        hit = true
                        attackerDamage = attacker.GetMeleeAttackPower() * max(0, attacker.Figures() - attackerFear)
                    }

                    immolationDamage := 0

                    damageFuncs := []func(){}

                    if hit {
                        model.Observer.MeleeAttack(attacker, defender, attackerDamage)
                        immolationDamage += model.immolationDamage(attacker, defender)
                        if attacker.Unit.CanTouchAttack(units.DamageMeleePhysical) {
                            damageFuncs = append(damageFuncs, model.doTouchAttack(attacker, defender, attackerFear)...)
                        }

                        model.ApplyMeleeDamage(attacker, defender, attackerDamage)
                        model.ApplyImmolationDamage(defender, immolationDamage)
                        for _, f := range damageFuncs {
                            f()
                        }
                    }
                }
            case 5:
                // attacker fear attack
                if attacker.HasAbility(data.AbilityCauseFear) {
                    defenderFear = defender.CauseFear()
                    model.AddLogEvent(fmt.Sprintf("%v causes fear in %v for %v figures", attacker.Unit.GetName(), defender.Unit.GetName(), defenderFear))
                    model.Observer.CauseFear(attacker, defender, defenderFear)
                }
            case 6:
                didFirstStrike := attacker.HasAbility(data.AbilityFirstStrike) && !defender.HasAbility(data.AbilityNegateFirstStrike)

                attacks := 1

                if didFirstStrike {
                    if attacker.HasEnchantment(data.UnitEnchantmentHaste) {
                        attacks = 1
                    } else {
                        // already melee attacked and doesn't have haste, so no more melee attacks
                        attacks = 0
                    }
                } else {
                    // didn't do first strike for whatever reason (either the attacker doesn't have the ability or the defender negated it)
                    if attacker.HasEnchantment(data.UnitEnchantmentHaste) {
                        attacks = 2
                    }
                }

                defenderImmolationDamage := 0
                defenderMeleeDamage := 0

                damageFuncs := []func(){}

                // attacker has not melee attacked yet, so let them do it now, or they have haste so they can attack again
                for range attacks {
                    attackerDamage, hit := attacker.ComputeMeleeDamage(defender, attackerFear, false)

                    if defender.IsAsleep() {
                        hit = true
                        attackerDamage = attacker.GetMeleeAttackPower() * max(0, attacker.Figures() - attackerFear)
                    }

                    if hit {
                        model.Observer.MeleeAttack(attacker, defender, attackerDamage)
                        defenderMeleeDamage += attackerDamage
                        defenderImmolationDamage += model.immolationDamage(attacker, defender)
                        if attacker.Unit.CanTouchAttack(units.DamageMeleePhysical) {
                            damageFuncs = append(damageFuncs, model.doTouchAttack(attacker, defender, attackerFear)...)
                        }
                    }
                }

                counters := 1
                if defender.HasEnchantment(data.UnitEnchantmentHaste) {
                    counters = 2
                }

                attackerImmolationDamage := 0
                attackerMeleeDamage := 0

                if !defender.IsAsleep() {
                    // defender does counter-attack
                    for range counters {
                        defenderDamage, hit := defender.ComputeMeleeDamage(attacker, defenderFear, true)

                        // attacker can't possibly be asleep, so no need to check here

                        if hit {
                            model.Observer.MeleeAttack(defender, attacker, defenderDamage)
                            attackerMeleeDamage += defenderDamage
                            attackerImmolationDamage += model.immolationDamage(defender, attacker)
                            if defender.Unit.CanTouchAttack(units.DamageMeleePhysical) {
                                damageFuncs = append(damageFuncs, model.doTouchAttack(defender, attacker, defenderFear)...)
                            }
                        }
                    }
                }

                model.ApplyImmolationDamage(defender, defenderImmolationDamage)
                model.ApplyMeleeDamage(attacker, defender, defenderMeleeDamage)

                model.ApplyImmolationDamage(attacker, attackerImmolationDamage)
                model.ApplyMeleeDamage(defender, attacker, attackerMeleeDamage)

                for _, f := range damageFuncs {
                    f()
                }
            }
    }

    for round := range 7 {
        doRound(round)
        end := false
        if defender.GetHealth() <= 0 {
            model.AddLogEvent(fmt.Sprintf("%v is killed", defender.Unit.GetName()))
            model.KillUnit(defender)
            end = true
            model.Observer.UnitKilled(defender)
        }

        if attacker.GetHealth() <= 0 {
            model.AddLogEvent(fmt.Sprintf("%v is killed", attacker.Unit.GetName()))
            model.KillUnit(attacker)
            end = true
            model.Observer.UnitKilled(attacker)
        }

        if end {
            break
        }
    }

    defender.Attacked += 1
}

func (model *CombatModel) KillUnit(unit *ArmyUnit){
    if unit.Team == TeamDefender {
        model.DefeatedDefenders += 1
        model.DefendingArmy.KillUnit(unit)
    } else {
        model.DefeatedAttackers += 1
        model.AttackingArmy.KillUnit(unit)
    }

    model.Tiles[unit.Y][unit.X].Unit = nil

    if unit == model.SelectedUnit {
        model.NextUnit()
    }
}

func (model *CombatModel) RemoveUnit(unit *ArmyUnit){
    if unit.Team == TeamDefender {
        model.DefeatedDefenders += 1
        model.DefendingArmy.RemoveUnit(unit)
    } else {
        model.DefeatedAttackers += 1
        model.AttackingArmy.RemoveUnit(unit)
    }

    model.Tiles[unit.Y][unit.X].Unit = nil

    if unit == model.SelectedUnit {
        model.NextUnit()
    }
}

func (model *CombatModel) RecallUnit(unit *ArmyUnit) {
    if unit.Team == TeamDefender {
        model.DefendingArmy.RecalledUnits = append(model.DefendingArmy.RecalledUnits, unit)
    } else {
        model.AttackingArmy.RecalledUnits = append(model.AttackingArmy.RecalledUnits, unit)
    }
    model.RemoveUnit(unit)
}

func (model *CombatModel) IsAIControlled(unit *ArmyUnit) bool {
    isConfused := unit.ConfusionAction == ConfusionActionEnemyControl
    if unit.Team == TeamDefender {
        return (model.DefendingArmy.IsAI() && !isConfused) || (model.AttackingArmy.IsAI() && isConfused)
    } else {
        return (model.AttackingArmy.IsAI() && !isConfused) || (model.DefendingArmy.IsAI() && isConfused)
    }
}

func (model *CombatModel) IsAdjacentToEnemy(unit *ArmyUnit) bool {
    for dx := -1; dx <= 1; dx++ {
        for dy := -1; dy <= 1; dy++ {
            x := unit.X + dx
            y := unit.Y + dy

            otherUnit := model.GetUnit(x, y)
            // confused units can see their own teammates
            if otherUnit != nil && (otherUnit.ConfusionAction == ConfusionActionEnemyControl || otherUnit.Team != unit.Team) {
                return true
            }
        }
    }

    return false
}

func (model *CombatModel) GetArmyForPlayer(player ArmyPlayer) *Army {
    if model.DefendingArmy.Player == player {
        return model.DefendingArmy
    }

    return model.AttackingArmy
}

func (model *CombatModel) GetOppositeArmyForPlayer(player ArmyPlayer) *Army {
    if model.DefendingArmy.Player == player {
        return model.AttackingArmy
    }

    return model.DefendingArmy
}

func (model *CombatModel) GetArmyForTeam(team Team) *Army {
    if team == TeamDefender {
        return model.DefendingArmy
    }

    return model.AttackingArmy
}

func (model *CombatModel) GetOppositeArmyForTeam(team Team) *Army {
    if team == TeamDefender {
        return model.AttackingArmy
    }

    return model.DefendingArmy
}

func (model *CombatModel) GetTeamForArmy(army *Army) Team {
    if army == model.DefendingArmy {
        return TeamDefender
    }

    return TeamAttacker
}

func (model *CombatModel) GetArmy(unit *ArmyUnit) *Army {
    if unit.Team == TeamDefender {
        return model.DefendingArmy
    }

    return model.AttackingArmy
}

func (model *CombatModel) GetOtherArmy(unit *ArmyUnit) *Army {
    if unit.Team == TeamDefender {
        return model.AttackingArmy
    }

    return model.DefendingArmy
}

// predicts the outcome of a battle just by comparing the relative power level of each army
func DoStrategicCombat(attackingArmy *Army, defendingArmy *Army) (CombatState, int, int) {
    fakeModel := CombatModel{
        AttackingArmy: attackingArmy,
        DefendingArmy: defendingArmy,
    }

    for _, unit := range attackingArmy.units {
        unit.Model = &fakeModel
    }

    for _, unit := range defendingArmy.units {
        unit.Model = &fakeModel
    }

    attackingPower := attackingArmy.GetPower()
    defendingPower := defendingArmy.GetPower()

    log.Printf("strategic combat: attacking power: %v, defending power: %v", attackingPower, defendingPower)

    // FIXME: Allow fleeing?

    if attackingPower > defendingPower {
        for _, unit := range defendingArmy.units {
            unit.TakeDamage(unit.GetMaxHealth(), DamageNormal)
        }

        return CombatStateAttackerWin, 0, len(defendingArmy.units)
    } else {
        for _, unit := range attackingArmy.units {
            unit.TakeDamage(unit.GetMaxHealth(), DamageNormal)
        }

        return CombatStateDefenderWin, len(attackingArmy.units), 0
    }
}

func (model *CombatModel) flee(army *Army) {
    for _, unit := range army.units {
        // FIXME: units unable to move always die

        // heroes have a 25% chance to die, normal units 50%
        chance := 50
        if unit.Unit.IsHero() {
            chance = 25
        }

        // units that are still under the web will always be lost
        if unit.IsWebbed() {
            chance = 100
        }

        if rand.IntN(100) < chance {
            unit.TakeDamage(unit.GetHealth(), DamageNormal)
            model.RemoveUnit(unit)
            model.DiedWhileFleeing += 1
        }
    }
}

// called when the battle ends
func (model *CombatModel) FinishCombat(state CombatState) {
    // kill all units that are bound or possessed, or summoned units
    // also regenerate units with the regeneration ability
    killUnits := func(army *Army, team Team) {
        wonBattle := state.IsWinner(team)

        for _, unit := range army.units {
            if wonBattle && unit.HasAbility(data.AbilityRegeneration) {
                unit.Heal(unit.GetMaxHealth())
            }

            if unit.GetHealth() > 0 {
                if unit.HasCurse(data.UnitCurseCreatureBinding) || unit.HasCurse(data.UnitCursePossession) || unit.Summoned {
                    unit.TakeDamage(unit.GetHealth(), DamageNormal)
                }
            }
        }

        var regeneratedUnits []*ArmyUnit
        var undeadUnits []*ArmyUnit

        var stillKilledUnits []*ArmyUnit

        for _, unit := range army.KilledUnits {
            killed := true
            if wonBattle && unit.HasAbility(data.AbilityRegeneration) {
                unit.Heal(unit.GetMaxHealth())
                regeneratedUnits = append(regeneratedUnits, unit)
                killed = false
            }

            if !wonBattle {
                // raise unit as an undead unit for the opposing team
                if unit.DeathReason() == DamageUndead && unit.GetRace() != data.RaceHero {
                    model.UndeadUnits = append(model.UndeadUnits, unit)
                    killed = false
                    undeadUnits = append(undeadUnits, unit)
                }
            }

            if killed {
                stillKilledUnits = append(stillKilledUnits, unit)
            }
        }

        army.KilledUnits = stillKilledUnits
        army.RegeneratedUnits = regeneratedUnits
    }

    killUnits(model.DefendingArmy, TeamDefender)
    killUnits(model.AttackingArmy, TeamAttacker)

    for _, unit := range model.UndeadUnits {
        unit.Unit.SetUndead()
        unit.Unit.AdjustHealth(unit.GetMaxHealth())
    }

    model.DefendingArmy.Cleanup()
    model.AttackingArmy.Cleanup()

    for _, cleanups := range model.Cleanups {
        cleanups()
    }
}

func (model *CombatModel) InsideMagicNode() bool {
    return model.Zone.GetMagic() != data.MagicNone
}

// returns true if the spell should be dispelled (due to counter magic, magic nodes, etc)
func (model *CombatModel) CheckDispel(spell spellbook.Spell, caster ArmyPlayer) bool {
    // FIXME: what should come first, counter magic or node dispel?
    if model.InsideMagicNode() && !caster.GetWizard().RetortEnabled(data.RetortNodeMastery) {
        nodeMagic := model.Zone.GetMagic()
        if spell.Magic != nodeMagic {
            chance := spellbook.ComputeDispelChance(50, spell.Cost(false), spell.Magic, caster.GetWizard())
            if spellbook.RollDispelChance(chance) {
                return true
            }
        }
    }

    opposite := model.GetOppositeArmyForPlayer(caster)
    if opposite.CounterMagic > 0 {
        // FIXME: should runemaster add to the counter magic dispel strength?
        chance := spellbook.ComputeDispelChance(opposite.CounterMagic, spell.Cost(false), spell.Magic, caster.GetWizard())
        opposite.CounterMagic = max(0, opposite.CounterMagic - 5)

        if opposite.CounterMagic == 0 {
            opposite.RemoveEnchantment(data.CombatEnchantmentCounterMagic)
        }

        return spellbook.RollDispelChance(chance)
    }

    return false
}

func (model *CombatModel) SwitchTeams(target *ArmyUnit) {
    newArmy := model.GetOtherArmy(target)
    oldArmy := model.GetArmy(target)
    oldArmy.RemoveUnit(target)

    newArmy.AddArmyUnit(target)
    target.Team = model.GetTeamForArmy(newArmy)
}

func (model *CombatModel) ApplyCreatureBinding(target *ArmyUnit){
    target.AddCurse(data.UnitCurseCreatureBinding)
    model.SwitchTeams(target)
}

func (model *CombatModel) ApplyPossession(target *ArmyUnit){
    target.AddCurse(data.UnitCursePossession)
    model.SwitchTeams(target)
}

/* let the user select a target, then cast the spell on that target
 */
func (model *CombatModel) DoTargetUnitSpell(player ArmyPlayer, spell spellbook.Spell, targetKind Targeting, onTarget func(*ArmyUnit), canTarget func(*ArmyUnit) bool) {
    teamAttacked := TeamAttacker

    selecter := TeamAttacker
    if player == model.DefendingArmy.Player {
        selecter = TeamDefender
    }

    if targetKind == TargetFriend {
        /* if the player is the defender and we are targeting a friend then the team should be the defenders */
        if model.DefendingArmy.Player == player {
            teamAttacked = TeamDefender
        }
    } else if targetKind == TargetEnemy {
        /* if the player is the attacker and we are targeting an enemy then the team should be the defenders */
        if model.AttackingArmy.Player == player {
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
        SelectTarget: onTarget,
    }

    select {
        case model.Events <- event:
        default:
    }
}

func (model *CombatModel) DoTargetTileSpell(player ArmyPlayer, spell spellbook.Spell, canTarget func(int, int) bool, onTarget func(int, int)){
    selecter := TeamAttacker
    if player == model.DefendingArmy.Player {
        selecter = TeamDefender
    }

    event := &CombatEventSelectTile{
        Selecter: selecter,
        Spell: spell,
        SelectTile: onTarget,
        CanTarget: canTarget,
    }

    select {
        case model.Events <- event:
        default:
    }
}

/* create projectiles on all units immediately, no targeting required
 */
func (model *CombatModel) DoAllUnitsSpell(player ArmyPlayer, spell spellbook.Spell, targetKind Targeting, onTarget func(*ArmyUnit), canTarget func(*ArmyUnit) bool) {
    var units []*ArmyUnit

    if player == model.DefendingArmy.Player && targetKind == TargetEnemy {
        units = model.AttackingArmy.units
    } else if player == model.AttackingArmy.Player && targetKind == TargetEnemy {
        units = model.DefendingArmy.units
    } else if player == model.DefendingArmy.Player && targetKind == TargetFriend {
        units = model.DefendingArmy.units
    } else if player == model.AttackingArmy.Player && targetKind == TargetFriend {
        units = model.AttackingArmy.units
    }

    model.Events <- &CombatPlaySound{
        Sound: spell.Sound,
    }

    for _, unit := range units {
        if canTarget(unit){
            onTarget(unit)
        }
    }
}

type SpellSystem interface {
    CreateFireballProjectile(target *ArmyUnit, cost int) *Projectile
    CreateIceBoltProjectile(target *ArmyUnit, cost int) *Projectile
    CreateStarFiresProjectile(target *ArmyUnit) *Projectile
    CreatePsionicBlastProjectile(target *ArmyUnit, cost int) *Projectile
    CreateDoomBoltProjectile(target *ArmyUnit) *Projectile
    CreateFireBoltProjectile(target *ArmyUnit, cost int) *Projectile
    CreateLightningBoltProjectile(target *ArmyUnit, cost int) *Projectile
    CreateWarpLightningProjectile(target *ArmyUnit) *Projectile
    CreateFlameStrikeProjectile(target *ArmyUnit) *Projectile
    CreateLifeDrainProjectile(target *ArmyUnit, reduceResistance int, player ArmyPlayer, unitCaster *ArmyUnit) *Projectile
    CreateDispelEvilProjectile(target *ArmyUnit) *Projectile
    CreateHealingProjectile(target *ArmyUnit) *Projectile
    CreateHolyWordProjectile(target *ArmyUnit) *Projectile
    CreateRecallHeroProjectile(target *ArmyUnit) *Projectile
    CreateCracksCallProjectile(target *ArmyUnit) *Projectile
    CreateWebProjectile(target *ArmyUnit) *Projectile
    CreateBanishProjectile(target *ArmyUnit, reduceResistance int) *Projectile
    CreateDispelMagicProjectile(target *ArmyUnit, caster ArmyPlayer, dispelStrength int) *Projectile
    CreateWordOfRecallProjectile(target *ArmyUnit) *Projectile
    CreateDisintegrateProjectile(target *ArmyUnit) *Projectile
    CreateDisruptProjectile(x int, y int) *Projectile
    CreateMagicVortex(x int, y int) *OtherUnit
    CreateWarpWoodProjectile(target *ArmyUnit) *Projectile
    CreateDeathSpellProjectile(target *ArmyUnit) *Projectile
    CreateWordOfDeathProjectile(target *ArmyUnit) *Projectile
    CreateSummoningCircle(x int, y int) *Projectile
    CreateMindStormProjectile(target *ArmyUnit) *Projectile
    CreateBlessProjectile(target *ArmyUnit) *Projectile
    CreateWeaknessProjectile(target *ArmyUnit) *Projectile
    CreateBlackSleepProjectile(target *ArmyUnit) *Projectile
    CreateVertigoProjectile(target *ArmyUnit) *Projectile
    CreateShatterProjectile(target *ArmyUnit) *Projectile
    CreateWarpCreatureProjectile(target *ArmyUnit) *Projectile
    CreateConfusionProjectile(target *ArmyUnit) *Projectile
    CreatePossessionProjectile(target *ArmyUnit) *Projectile
    CreateCreatureBindingProjectile(target *ArmyUnit) *Projectile
    CreatePetrifyProjectile(target *ArmyUnit) *Projectile
    CreateChaosChannelsProjectile(target *ArmyUnit) *Projectile
    CreateHeroismProjectile(target *ArmyUnit) *Projectile
    CreateHolyArmorProjectile(target *ArmyUnit) *Projectile
    CreateHolyWeaponProjectile(target *ArmyUnit) *Projectile
    CreateInvulnerabilityProjectile(target *ArmyUnit) *Projectile
    CreateLionHeartProjectile(target *ArmyUnit) *Projectile
    CreateRighteousnessProjectile(target *ArmyUnit) *Projectile
    CreateTrueSightProjectile(target *ArmyUnit) *Projectile
    CreateElementalArmorProjectile(target *ArmyUnit) *Projectile
    CreateGiantStrengthProjectile(target *ArmyUnit) *Projectile
    CreateIronSkinProjectile(target *ArmyUnit) *Projectile
    CreateStoneSkinProjectile(target *ArmyUnit) *Projectile
    CreateRegenerationProjectile(target *ArmyUnit) *Projectile
    CreateResistElementsProjectile(target *ArmyUnit) *Projectile
    CreateFlightProjectile(target *ArmyUnit) *Projectile
    CreateGuardianWindProjectile(target *ArmyUnit) *Projectile
    CreateHasteProjectile(target *ArmyUnit) *Projectile
    CreateInvisibilityProjectile(target *ArmyUnit) *Projectile
    CreateMagicImmunityProjectile(target *ArmyUnit) *Projectile
    CreateResistMagicProjectile(target *ArmyUnit) *Projectile
    CreateSpellLockProjectile(target *ArmyUnit) *Projectile
    CreateEldritchWeaponProjectile(target *ArmyUnit) *Projectile
    CreateFlameBladeProjectile(target *ArmyUnit) *Projectile
    CreateImmolationProjectile(target *ArmyUnit) *Projectile
    CreateBerserkProjectile(target *ArmyUnit) *Projectile
    CreateCloakOfFearProjectile(target *ArmyUnit) *Projectile
    CreateWraithFormProjectile(target *ArmyUnit) *Projectile

    GetAllSpells() spellbook.Spells
}

// playerCasted is true if the player cast the spell, or false if a unit cast the spell
func (model *CombatModel) InvokeSpell(spellSystem SpellSystem, player ArmyPlayer, unitCaster *ArmyUnit, spell spellbook.Spell, castedCallback func()){

    if model.CheckDispel(spell, player) {
        model.Events <- &CombatEventMessage{
            Message: fmt.Sprintf("%v fizzled", spell.Name),
        }
        castedCallback()
        return
    }

    // shared eligible functions

    // any target is eligible
    targetAny := func (target *ArmyUnit) bool { return true }

    // only fantastic units are eligible
    targetFantastic := func (target *ArmyUnit) bool {
        return target != nil && target.Unit.GetRace() == data.RaceFantastic
    }

    targetHero := func (target *ArmyUnit) bool {
        return target.Unit.IsHero()
    }

    // anything non-death related
    targetNonDeath := func (target *ArmyUnit) bool {
        return target.GetRealm() != data.DeathMagic
    }

    targetNotImmune := func (target *ArmyUnit) bool {
        return !target.IsMagicImmune(spell.Magic)
    }

    healingTarget := targetNonDeath
    chaosChannelsTarget := func (target *ArmyUnit) bool {
        return target.Unit.GetRace() != data.RaceFantastic
    }

    warpCreatureTarget := func (target *ArmyUnit) bool {
        if target.GetRace() == data.RaceFantastic {
            return false
        }

        if target.IsMagicImmune(spell.Magic) {
            return false
        }

        // warp creature can be cast on the same unit multiple times, where a different curse effect
        // is applied each time. if a unit has all 3 curses then potentially we could return false here

        return true
    }

    fireBoltTarget := targetNotImmune
    warpLightningTarget := targetNotImmune
    doomBoltTarget := targetNotImmune
    disintegrateTarget := targetNotImmune

    switch spell.Name {
        case "Fireball":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateFireballProjectile(target, spell.Cost(false) / 3))
                castedCallback()
            }, targetNotImmune)
        case "Ice Bolt":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateIceBoltProjectile(target, spell.Cost(false)))
                castedCallback()
            }, targetAny)
        case "Star Fires":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateStarFiresProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                realm := target.Unit.GetRealm()
                if target.Unit.GetRace() == data.RaceFantastic && (realm == data.ChaosMagic || realm == data.DeathMagic) {
                    return true
                }

                return false
            })
        case "Psionic Blast":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreatePsionicBlastProjectile(target, spell.Cost(false) / 2))
                castedCallback()
            }, targetAny)
        case "Doom Bolt":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateDoomBoltProjectile(target))
                castedCallback()
            }, doomBoltTarget)
        case "Fire Bolt":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateFireBoltProjectile(target, spell.Cost(false)))
                castedCallback()
            }, fireBoltTarget)
        case "Lightning Bolt":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateLightningBoltProjectile(target, spell.Cost(false) - 5))
                castedCallback()
            }, targetNotImmune)
        case "Warp Lightning":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateWarpLightningProjectile(target))
                castedCallback()
            }, warpLightningTarget)
        case "Flame Strike":
            model.DoAllUnitsSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateFlameStrikeProjectile(target))
            }, targetAny)
            castedCallback()
        case "Life Drain":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateLifeDrainProjectile(target, spell.SpentAdditionalCost(false) / 5, player, unitCaster))
                castedCallback()
            }, targetNotImmune)
        case "Dispel Evil":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateDispelEvilProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.Unit.GetRace() == data.RaceFantastic &&
                   (target.Unit.GetRealm() == data.ChaosMagic || target.Unit.GetRealm() == data.DeathMagic) {
                    return true
                }

                return false
            })
        case "Healing":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateHealingProjectile(target))
                castedCallback()
            }, healingTarget)
        case "Holy Word":
            model.DoAllUnitsSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateHolyWordProjectile(target))
            }, targetFantastic)
            castedCallback()
        case "Recall Hero":
            // FIXME:  check planar seal and summoning circle?
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateRecallHeroProjectile(target))
                castedCallback()
            }, targetHero)
        case "Mass Healing":
            model.DoAllUnitsSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateHealingProjectile(target))
            }, targetNonDeath)
            castedCallback()
        case "Cracks Call":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateCracksCallProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.IsFlying() {
                    return false
                }

                if target.HasAbility(data.AbilityNonCorporeal) {
                    return false
                }

                if target.HasAbility(data.AbilityMerging) {
                    return false
                }

                return true
            })
        case "Earth to Mud":
            model.DoTargetTileSpell(player, spell, func (x int, y int) bool { return true}, func (x int, y int){
                model.CreateEarthToMud(x, y)
                castedCallback()
            })
        case "Web":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateWebProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                return !target.HasAbility(data.AbilityNonCorporeal)
            })
        case "Banish":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateBanishProjectile(target, spell.SpentAdditionalCost(false) / 15))
                castedCallback()
            }, targetFantastic)
        case "Dispel Magic True":
            model.DoTargetUnitSpell(player, spell, TargetEither, func(target *ArmyUnit){
                disenchantStrength := spell.Cost(false) * 3

                if player.GetWizard().RetortEnabled(data.RetortRunemaster) {
                    disenchantStrength *= 2
                }

                model.AddProjectile(spellSystem.CreateDispelMagicProjectile(target, player, disenchantStrength))
                castedCallback()
            }, targetAny)
        case "Dispel Magic":
            model.DoTargetUnitSpell(player, spell, TargetEither, func(target *ArmyUnit){
                disenchantStrength := spell.Cost(false)

                if player.GetWizard().RetortEnabled(data.RetortRunemaster) {
                    disenchantStrength *= 2
                }

                model.AddProjectile(spellSystem.CreateDispelMagicProjectile(target, player, disenchantStrength))
                castedCallback()
            }, targetAny)
        case "Word of Recall":
            // FIXME: check planar seal and summoning circle?
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateWordOfRecallProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.Summoned {
                    return false
                }

                return true
            })
        case "Disintegrate":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateDisintegrateProjectile(target))
                castedCallback()
            }, disintegrateTarget)
        case "Disrupt":
            model.DoTargetTileSpell(player, spell, model.ContainsWall, func (x int, y int){
                model.AddProjectile(spellSystem.CreateDisruptProjectile(x, y))
                castedCallback()
            })
        case "Magic Vortex":
            // FIXME: should this also take walls into account?
            unoccupied := func (x int, y int) bool {
                return model.GetUnit(x, y) == nil
            }

            model.DoTargetTileSpell(player, spell, unoccupied, func (x int, y int){
                model.OtherUnits = append(model.OtherUnits, spellSystem.CreateMagicVortex(x, y))
                castedCallback()
            })
        case "Warp Wood":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateWarpWoodProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.IsMagicImmune(spell.Magic) {
                    return false
                }

                if target.GetRangedAttacks() > 0 && target.GetRangedAttackDamageType() == units.DamageRangedPhysical {
                    return true
                }

                return false
            })
        case "Death Spell":
            model.DoAllUnitsSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateDeathSpellProjectile(target))
            }, targetNotImmune)
            castedCallback()
        case "Word of Death":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateWordOfDeathProjectile(target))
                castedCallback()
            }, targetNotImmune)
        case "Phantom Warriors":
            model.DoSummoningSpell(spellSystem, player, spell, func(x int, y int){
                model.summonUnit(player, x, y, units.PhantomWarrior, units.FacingDown, true)
                castedCallback()
            })
        case "Phantom Beast":
            model.DoSummoningSpell(spellSystem, player, spell, func(x int, y int){
                model.summonUnit(player, x, y, units.PhantomBeast, units.FacingDown, true)
                castedCallback()
            })
        case "Earth Elemental":
            model.DoSummoningSpell(spellSystem, player, spell, func(x int, y int){
                model.summonUnit(player, x, y, units.EarthElemental, units.FacingDown, true)
                castedCallback()
            })
        case "Air Elemental":
            model.DoSummoningSpell(spellSystem, player, spell, func(x int, y int){
                model.summonUnit(player, x, y, units.AirElemental, units.FacingDown, true)
                castedCallback()
            })
        case "Fire Elemental":
            model.DoSummoningSpell(spellSystem, player, spell, func(x int, y int){
                model.summonUnit(player, x, y, units.FireElemental, units.FacingDown, true)
                castedCallback()
            })
        case "Summon Demon":
            x, y, err := model.FindEmptyTile(MapSideMiddle)
            if err == nil {
                spellSystem.CreateSummoningCircle(x, y)
                model.summonUnit(player, x, y, units.Demon, units.FacingDown, true)
                castedCallback()
            }
        case "Disenchant Area", "Disenchant True":
            // show some animation and play sound

            model.Events <- &CombatEventGlobalSpell{
                Caster: player,
                Magic: spell.Magic,
                Name: spell.Name,
            }

            disenchantStrength := spell.Cost(false)
            if spell.Name == "Disenchant True" {
                // strength is 3x mana spent
                disenchantStrength = spell.Cost(false) * 3
            }

            if player.GetWizard().RetortEnabled(data.RetortRunemaster) {
                disenchantStrength *= 2
            }

            model.DoDisenchantArea(spellSystem.GetAllSpells(), player, disenchantStrength)

            castedCallback()

        case "High Prayer":
            model.CastEnchantment(player, data.CombatEnchantmentHighPrayer, castedCallback)
        case "Prayer":
            model.CastEnchantment(player, data.CombatEnchantmentPrayer, castedCallback)
        case "True Light":
            model.CastEnchantment(player, data.CombatEnchantmentTrueLight, castedCallback)
        case "Call Lightning":
            model.CastEnchantment(player, data.CombatEnchantmentCallLightning, castedCallback)
        case "Entangle":
            model.CastEnchantment(player, data.CombatEnchantmentEntangle, castedCallback)
        case "Blur":
            model.CastEnchantment(player, data.CombatEnchantmentBlur, castedCallback)
        case "Counter Magic":
            model.CastEnchantment(player, data.CombatEnchantmentCounterMagic, castedCallback)
            // set counter magic counter for the player to be the spell strength
            model.GetArmyForPlayer(player).CounterMagic = spell.Cost(false)
        case "Mass Invisibility":
            model.CastEnchantment(player, data.CombatEnchantmentMassInvisibility, castedCallback)
        case "Metal Fires":
            model.CastEnchantment(player, data.CombatEnchantmentMetalFires, castedCallback)
        case "Warp Reality":
            model.CastEnchantment(player, data.CombatEnchantmentWarpReality, castedCallback)
        case "Black Prayer":
            model.CastEnchantment(player, data.CombatEnchantmentBlackPrayer, castedCallback)
        case "Darkness":
            model.CastEnchantment(player, data.CombatEnchantmentDarkness, castedCallback)
        case "Mana Leak":
            model.CastEnchantment(player, data.CombatEnchantmentManaLeak, castedCallback)
        case "Terror":
            model.CastEnchantment(player, data.CombatEnchantmentTerror, castedCallback)
        case "Wrack":
            model.CastEnchantment(player, data.CombatEnchantmentWrack, castedCallback)

        case "Creature Binding":
            selectable := func(target *ArmyUnit) bool {
                if target == nil {
                    return false
                }

                if target.Unit.GetRace() == data.RaceFantastic {
                    if target.HasAbility(data.AbilityIllusionsImmunity) {
                        return false
                    }

                    if target.HasAbility(data.AbilityMagicImmunity) {
                        return false
                    }

                    return true
                }

                return false
            }

            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateCreatureBindingProjectile(target))
                castedCallback()
            }, selectable)

        case "Mind Storm":
            selectable := func(target *ArmyUnit) bool {
                if target != nil {
                    if target.HasAbility(data.AbilityIllusionsImmunity) || target.HasAbility(data.AbilityMagicImmunity) {
                        return false
                    }

                    return true
                }

                return false
            }

            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateMindStormProjectile(target))
                castedCallback()
            }, selectable)

        case "Bless":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateBlessProjectile(target))
                castedCallback()
            }, targetAny)
        case "Weakness":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateWeaknessProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.HasCurse(data.UnitCurseWeakness) {
                    return false
                }

                if target.IsMagicImmune(spell.Magic) {
                    return false
                }

                if target.HasAbility(data.AbilityCharmed) {
                    return false
                }

                return true
            })
        case "CurseBlackSleep":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateBlackSleepProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.HasCurse(data.UnitCurseBlackSleep) {
                    return false
                }

                if target.IsMagicImmune(spell.Magic) {
                    return false
                }

                if target.HasEnchantment(data.UnitEnchantmentRighteousness) {
                    return false
                }

                if target.HasAbility(data.AbilityCharmed) {
                    return false
                }

                return true
            })
        case "Vertigo":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateVertigoProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.HasCurse(data.UnitCurseVertigo) {
                    return false
                }

                if target.HasAbility(data.AbilityIllusionsImmunity) || target.IsMagicImmune(spell.Magic) {
                    return false
                }

                return true
            })
        case "Shatter":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateShatterProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.GetRace() == data.RaceFantastic {
                    return false
                }

                if target.IsMagicImmune(spell.Magic) {
                    return false
                }

                if target.HasCurse(data.UnitCurseShatter) {
                    return false
                }

                return true
            })
        case "Warp Creature":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateWarpCreatureProjectile(target))
                castedCallback()
            }, warpCreatureTarget)
        case "Confusion":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateConfusionProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.HasAbility(data.AbilityIllusionsImmunity) || target.IsMagicImmune(spell.Magic) {
                    return false
                }

                if target.HasCurse(data.UnitCurseConfusion) {
                    return false
                }

                return true
            })
        case "Possession":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreatePossessionProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.Unit.IsHero() || target.GetRace() == data.RaceFantastic {
                    return false
                }

                if target.IsMagicImmune(spell.Magic) {
                    return false
                }

                return true
            })
        case "Petrify":
            model.DoTargetUnitSpell(player, spell, TargetEnemy, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreatePetrifyProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.IsMagicImmune(spell.Magic) || target.HasAbility(data.AbilityStoningImmunity) {
                    return false
                }

                return true
            })
        case "Call Chaos":
            otherArmy := model.GetOppositeArmyForPlayer(player)
            for _, unit := range otherArmy.units {
                switch rand.N(8) {
                    // nothing
                    case 0:

                    // healing
                    case 1:
                        if healingTarget(unit) {
                            model.AddProjectile(spellSystem.CreateHealingProjectile(unit))
                        }

                    // chaos channels
                    case 2:
                        if chaosChannelsTarget(unit) {
                            model.AddProjectile(spellSystem.CreateChaosChannelsProjectile(unit))
                        }

                    // warp creature
                    case 3:
                        if warpCreatureTarget(unit) {
                            model.AddProjectile(spellSystem.CreateWarpCreatureProjectile(unit))
                        }

                    // fire bolt
                    case 4:
                        if fireBoltTarget(unit) {
                            model.AddProjectile(spellSystem.CreateFireBoltProjectile(unit, 15))
                        }

                    // warp lightning
                    case 5:
                        if warpLightningTarget(unit) {
                            model.AddProjectile(spellSystem.CreateWarpLightningProjectile(unit))
                        }

                    // doom bolt
                    case 6:
                        if doomBoltTarget(unit) {
                            model.AddProjectile(spellSystem.CreateDoomBoltProjectile(unit))
                        }

                    // disintegrate
                    case 7:
                        if disintegrateTarget(unit) {
                            model.AddProjectile(spellSystem.CreateDisintegrateProjectile(unit))
                        }
                }
            }

            castedCallback()
        case "Raise Dead":
            army := model.GetArmyForPlayer(player)
            failed := true

            doRaiseDead := func (killedUnit *ArmyUnit){
                model.DoSummoningSpell(spellSystem, player, spell, func(x int, y int){
                    // shouldn't really be necessary because the model should already be set, but just in case
                    killedUnit.Model = model
                    army.RaiseDeadUnit(killedUnit, x, y)
                    model.Tiles[y][x].Unit = killedUnit
                    castedCallback()
                })

            }

            if len(army.KilledUnits) > 0 {
                if len(army.KilledUnits) == 1 {

                    killedUnit := army.KilledUnits[0]

                    if killedUnit.GetRace() != data.RaceFantastic {
                        failed = false
                        doRaiseDead(killedUnit)
                    }
                } else {
                    // show selection box to choose a dead unit

                    var targets []*ArmyUnit
                    for _, killedUnit := range army.KilledUnits {
                        if killedUnit.GetRace() != data.RaceFantastic {
                            targets = append(targets, killedUnit)
                        }
                    }

                    if len(targets) > 0 {
                        failed = false
                        model.Events <- &CombatSelectTargets{
                            Title: "Select a unit to Revive",
                            Targets: targets,
                            Select: func (target *ArmyUnit){
                                doRaiseDead(target)
                            },
                        }
                    }
                }
            }

            if failed {
                model.Events <- &CombatEventMessage{
                    Message: "There are no available units to revive.",
                }
            }
        case "Animate Dead":
            // first filter possible candidates to revive
            allKilledUnits := slices.DeleteFunc(slices.Clone(append(model.AttackingArmy.KilledUnits, model.DefendingArmy.KilledUnits...)), func (unit *ArmyUnit) bool {
                if unit.Unit.IsHero() {
                    return true
                }

                if unit.GetRealm() == data.DeathMagic {
                    return true
                }

                if unit.Summoned {
                    return true
                }

                return false
            })

            // remove unit from the army's killed units list
            // add unit to caster's army
            // make unit undead
            // restore health
            doAnimateDead := func (killedUnit *ArmyUnit){
                x, y, err := model.FindEmptyTile(model.GetSideForPlayer(player))
                if err != nil {
                    log.Printf("Unable to find empty tile to animate unit")
                    return
                }
                model.AddProjectile(spellSystem.CreateSummoningCircle(x, y))
                ownerArmy := model.GetArmy(killedUnit)
                ownerArmy.KilledUnits = slices.DeleteFunc(ownerArmy.KilledUnits, func (unit *ArmyUnit) bool {
                    return unit == killedUnit
                })
                army := model.GetArmyForPlayer(player)
                army.AddArmyUnit(killedUnit)
                killedUnit.Unit.SetUndead()
                killedUnit.Heal(killedUnit.GetMaxHealth())
                killedUnit.X = x
                killedUnit.Y = y
                killedUnit.Team = model.GetTeamForArmy(army)
                killedUnit.Model = model
                killedUnit.Enchantments = nil
                killedUnit.Curses = nil
                killedUnit.WebHealth = 0
                model.Tiles[y][x].Unit = killedUnit
            }

            if len(allKilledUnits) > 0 {
                if len(allKilledUnits) == 1 {
                    raised := allKilledUnits[0]
                    doAnimateDead(raised)
                    castedCallback()
                } else {
                    model.Events <- &CombatSelectTargets{
                        Title: "Select a unit to Animate",
                        Targets: allKilledUnits,
                        Select: func (target *ArmyUnit){
                            doAnimateDead(target)
                            castedCallback()
                        },
                    }
                }
            } else {
                model.Events <- &CombatEventMessage{
                    Message: "There are no available units to raise.",
                }
            }

        case "Heroism":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateHeroismProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.GetRace() == data.RaceFantastic {
                    return false
                }

                if target.GetRealm() == data.DeathMagic {
                    return false
                }

                // elite level for both units and heroes is 3
                if target.GetExperienceData().ToInt() >= 3 {
                    return false
                }

                if target.HasEnchantment(data.UnitEnchantmentHeroism) {
                    return false
                }

                return true
            })
        case "Holy Armor":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateHolyArmorProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.GetRace() == data.RaceFantastic {
                    return false
                }

                if target.GetRealm() == data.DeathMagic {
                    return false
                }

                if target.HasEnchantment(data.UnitEnchantmentHolyArmor) {
                    return false
                }

                return true
            })
        case "Holy Weapon":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateHolyWeaponProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.GetRace() == data.RaceFantastic {
                    return false
                }

                if target.GetRealm() == data.DeathMagic {
                    return false
                }

                if target.HasEnchantment(data.UnitEnchantmentHolyWeapon) {
                    return false
                }

                return true
            })
        case "Invulnerability":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateInvulnerabilityProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.GetRealm() == data.DeathMagic {
                    return false
                }

                if target.HasEnchantment(data.UnitEnchantmentInvulnerability) {
                    return false
                }

                return true
            })
        case "Lionheart":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateLionHeartProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.GetRealm() == data.DeathMagic {
                    return false
                }

                if target.HasEnchantment(data.UnitEnchantmentLionHeart) {
                    return false
                }

                return true
            })
        case "Righteousness":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateRighteousnessProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.GetRealm() == data.DeathMagic {
                    return false
                }

                if target.HasEnchantment(data.UnitEnchantmentRighteousness) {
                    return false
                }

                return true
            })
        case "True Sight":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateTrueSightProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.GetRealm() == data.DeathMagic {
                    return false
                }

                if target.HasEnchantment(data.UnitEnchantmentTrueSight) {
                    return false
                }

                return true
            })
        case "Elemental Armor":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateElementalArmorProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.HasEnchantment(data.UnitEnchantmentElementalArmor) {
                    return false
                }

                return true
            })
        case "Giant Strength":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateGiantStrengthProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.HasEnchantment(data.UnitEnchantmentGiantStrength) {
                    return false
                }

                return true
            })
        case "Iron Skin":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateIronSkinProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                // if a target has stone skin they cannot also have iron skin
                if target.HasEnchantment(data.UnitEnchantmentIronSkin) || target.HasEnchantment(data.UnitEnchantmentStoneSkin) {
                    return false
                }

                return true
            })
        case "Regeneration":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateRegenerationProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.HasAbility(data.AbilityRegeneration){
                    return false
                }

                return true
            })
        case "Resist Elements":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateResistElementsProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.HasEnchantment(data.UnitEnchantmentResistElements) {
                    return false
                }

                return true
            })
        case "Stone Skin":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateStoneSkinProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                // if a target has iron skin they cannot also have stone skin
                if target.HasEnchantment(data.UnitEnchantmentIronSkin) || target.HasEnchantment(data.UnitEnchantmentStoneSkin) {
                    return false
                }

                return true
            })
        case "Flight":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateFlightProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                // we could check if the unit has the flight ability, but the Flight spell also grants
                // 3 movement, so its still worthwhile to cast Flight on a unit that has less than 3 movement speed
                if target.HasEnchantment(data.UnitEnchantmentFlight) {
                    return false
                }

                return true
            })
        case "Guardian Wind":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateGuardianWindProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.HasEnchantment(data.UnitEnchantmentGuardianWind) {
                    return false
                }

                return true
            })
        case "Haste":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateHasteProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.HasEnchantment(data.UnitEnchantmentHaste) {
                    return false
                }

                return true
            })
        case "Invisiblity":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateInvisibilityProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.HasEnchantment(data.UnitEnchantmentInvisibility) {
                    return false
                }

                return true
            })
        case "Magic Immunity":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateMagicImmunityProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.HasAbility(data.AbilityMagicImmunity) {
                    return false
                }

                return true
            })
        case "Resist Magic":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateResistMagicProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.HasEnchantment(data.UnitEnchantmentResistMagic) {
                    return false
                }

                return true
            })
        case "Spell Lock":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateSpellLockProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.HasEnchantment(data.UnitEnchantmentSpellLock) {
                    return false
                }

                return true
            })
        case "Eldritch Weapon":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateEldritchWeaponProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.GetRace() == data.RaceFantastic {
                    return false
                }

                if target.HasEnchantment(data.UnitEnchantmentEldritchWeapon) {
                    return false
                }

                return true
            })
        case "Flame Blade":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateFlameBladeProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.GetRace() == data.RaceFantastic {
                    return false
                }

                if target.HasEnchantment(data.UnitEnchantmentFlameBlade) {
                    return false
                }

                return true
            })
        case "Immolation":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateImmolationProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.HasAbility(data.AbilityImmolation) {
                    return false
                }

                return true
            })
        case "Berserk":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateBerserkProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.HasEnchantment(data.UnitEnchantmentBerserk) {
                    return false
                }

                if target.GetRealm() == data.LifeMagic {
                    return false
                }

                return true
            })
        case "Cloak of Fear":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateCloakOfFearProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.HasEnchantment(data.UnitEnchantmentCloakOfFear) {
                    return false
                }

                if target.GetRealm() == data.LifeMagic {
                    return false
                }

                return true
            })
        case "Wraith Form":
            model.DoTargetUnitSpell(player, spell, TargetFriend, func(target *ArmyUnit){
                model.AddProjectile(spellSystem.CreateWraithFormProjectile(target))
                castedCallback()
            }, func (target *ArmyUnit) bool {
                if target.HasEnchantment(data.UnitEnchantmentWraithForm) {
                    return false
                }

                if target.GetRealm() == data.LifeMagic {
                    return false
                }

                return true
            })

        case "Wall of Fire":
            model.Events <- &CombatCreateWallOfFire{
                Sound: spell.Sound,
            }
        case "Wall of Darkness":
            model.Events <- &CombatCreateWallOfDarkness{
                Sound: spell.Sound,
            }


        default:
            log.Printf("Unhandled spell %v", spell.Name)
    }
}

func (model *CombatModel) DoSummoningSpell(spellSystem SpellSystem, player ArmyPlayer, spell spellbook.Spell, onTarget func(int, int)){

    side := model.GetSideForPlayer(player)

    allowed := func (x int, y int) bool {
        return model.IsOnSide(x, y, side)
    }

    // FIXME: pass in a canTarget function that only allows summoning on an empty tile on the casting wizards side of the battlefield
    model.DoTargetTileSpell(player, spell, allowed, func (x int, y int){
        model.AddProjectile(spellSystem.CreateSummoningCircle(x, y))
        // FIXME: there should be a delay between the summoning circle appearing and when the unit appears
        onTarget(x, y)
    })
}

func (model *CombatModel) CastEnchantment(player ArmyPlayer, enchantment data.CombatEnchantment, castedCallback func()){
    if model.AddEnchantment(player, enchantment) {
        model.Events <- &CombatEventGlobalSpell{
            Caster: player,
            Magic: enchantment.Magic(),
            Name: enchantment.Name(),
        }
        castedCallback()
    } else {
        model.Events <- &CombatEventMessage{
            Message: "That combat enchantment is already in effect",
        }
    }
}

func (model *CombatModel) GetHumanArmy() *Army {
    if model.AttackingArmy.Player.IsHuman() {
        return model.AttackingArmy
    } else {
        return model.DefendingArmy
    }
}
