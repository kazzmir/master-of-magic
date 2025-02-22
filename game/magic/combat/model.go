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
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"

    "github.com/hajimehoshi/ebiten/v2"
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
    // whether to show wall of darkness on this tile
    Darkness *set.Set[DarknessSide]

    Wall *set.Set[WallKind]

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

        // clear all space around the city
        for x := townSquare.Min.X; x <= townSquare.Max.X; x++ {
            for y := townSquare.Min.Y; y <= townSquare.Max.Y; y++ {
                tiles[y][x].ExtraObject.Index = -1
                tiles[y][x].InsideTown = true
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
            createWallOfFire(tiles, TownCenterX, TownCenterY, 4)
        }

        if zone.City.HasWallOfDarkness() {
            createWallOfDarkness(tiles, TownCenterX, TownCenterY, 4)
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
                screen.DrawImage(base, options)

                top, _ := imageCache.GetImage("chriver.lbx", 24 + int((counter / 4) % 8), 0)
                options.GeoM.Translate(16, -3)
                screen.DrawImage(top, options)

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
func createWallOfFire(tiles [][]Tile, centerX int, centerY int, sideLength int) {
    set := func(x int, y int, direction CardinalDirection) {
        tile := &tiles[y][x]
        if tile.Fire == nil {
            tile.Fire = set.MakeSet[FireSide]()
        }

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

func createWallOfDarkness(tiles [][]Tile, centerX int, centerY int, sideLength int) {
    set := func(x int, y int, direction CardinalDirection) {
        tile := &tiles[y][x]
        if tile.Darkness == nil {
            tile.Darkness = set.MakeSet[DarknessSide]()
        }

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

type CombatUnit interface {
    HasAbility(data.AbilityType) bool
    HasItemAbility(data.ItemAbility) bool
    GetAbilityValue(data.AbilityType) float32
    GetDefense() int
    GetResistance() int
    AdjustHealth(int)
    GetAbilities() []data.Ability
    GetBanner() data.BannerType
    GetRangedAttackDamageType() units.Damage
    GetRangedAttackPower() int
    GetMeleeAttackPower() int
    GetMaxHealth() int
    GetHitPoints() int
    GetEnchantments() []data.UnitEnchantment
    HasEnchantment(data.UnitEnchantment) bool
    GetCount() int
    GetHealth() int
    GetToHitMelee() int
    GetKnownSpells() []string
    GetRangedAttacks() int
    GetCombatLbxFile() string
    GetCombatIndex(units.Facing) int
    GetCombatRangeIndex(units.Facing) int
    GetMovementSound() units.MovementSound
    GetRangeAttackSound() units.RangeAttackSound
    GetAttackSound() units.AttackSound
    GetName() string
    GetMovementSpeed() int
    CanTouchAttack(units.Damage) bool
    IsFlying() bool
    IsHero() bool
    IsUndead() bool
    GetRace() data.Race
    GetRealm() data.MagicType
    GetSpellChargeSpells() map[spellbook.Spell]int
}

type ArmyUnit struct {
    Unit CombatUnit
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

    Model *CombatModel

    Team Team

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

    // ugly to need this, but this caches paths computed for the unit
    Paths map[image.Point]pathfinding.Path
}

func (unit *ArmyUnit) GetResistance() int {
    modifier := 0

    if unit.Unit.GetRace() == data.RaceFantastic && unit.Model.IsEnchantmentActive(data.CombatEnchantmentTrueLight, TeamEither) {
        switch unit.Unit.GetRealm() {
            case data.LifeMagic: modifier += 1
            case data.DeathMagic: modifier -= 1
        }
    }

    return max(0, unit.Unit.GetResistance() + modifier)
}

func (unit *ArmyUnit) GetDefense() int {
    modifier := 0

    if unit.Unit.GetRace() == data.RaceFantastic && unit.Model.IsEnchantmentActive(data.CombatEnchantmentTrueLight, TeamEither) {
        switch unit.Unit.GetRealm() {
            case data.LifeMagic: modifier += 1
            case data.DeathMagic: modifier -= 1
        }
    }

    return max(0, unit.Unit.GetDefense() + modifier)
}

func (unit *ArmyUnit) GetMeleeAttackPower() int {
    modifier := 0

    if unit.Unit.GetRace() == data.RaceFantastic && unit.Model.IsEnchantmentActive(data.CombatEnchantmentTrueLight, TeamEither) {
        switch unit.Unit.GetRealm() {
            case data.LifeMagic: modifier += 1
            case data.DeathMagic: modifier -= 1
        }
    }

    return max(0, unit.Unit.GetMeleeAttackPower() + modifier)
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
    power += unit.Unit.GetRangedAttackPower() * unit.Figures()
    power += unit.GetMeleeAttackPower() * unit.Figures()

    return power
}

// true if this unit can move through a tile with a wall tile
func (unit *ArmyUnit) CanTraverseWall() bool {
    return unit.Unit.IsFlying() || unit.HasAbility(data.AbilityMerging) || unit.HasAbility(data.AbilityTeleporting)
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

func (unit *ArmyUnit) GetEnchantments() []data.UnitEnchantment {
    return append(slices.Clone(unit.Unit.GetEnchantments()), unit.Enchantments...)
}

func (unit *ArmyUnit) AddEnchantment(enchantment data.UnitEnchantment) {
    // skip duplicates
    for _, check := range unit.Enchantments {
        if check == enchantment {
            return
        }
    }
    unit.Enchantments = append(unit.Enchantments, enchantment)
}

func (unit *ArmyUnit) GetResistances(enchantments... data.UnitEnchantment) int {
    resistance := 0

    for _, enchantment := range enchantments {

        if unit.HasEnchantment(enchantment) {
            switch enchantment {
                case data.UnitEnchantmentBless: resistance += 3
                case data.UnitEnchantmentElementalArmor: resistance += 10
                case data.UnitEnchantmentRighteousness: resistance += 30
                case data.UnitEnchantmentResistMagic: resistance += 5
                case data.UnitEnchantmentResistElements: resistance += 3
            }
        }
    }

    return resistance
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

func (unit *ArmyUnit) ResetTurnData() {
    unit.MovesLeft = fraction.FromInt(unit.Unit.GetMovementSpeed())
    unit.Paths = make(map[image.Point]pathfinding.Path)
    unit.Casted = false
}

func (unit *ArmyUnit) ComputeDefense(damage units.Damage, armorPiercing bool, wallDefense int) int {
    toDefend := unit.ToDefend()
    defenseRolls := unit.GetDefense()

    hasImmunity := false

    switch damage {
        case units.DamageRangedMagical:
            if unit.HasAbility(data.AbilityLargeShield) {
                defenseRolls += 2
            }

            if unit.HasAbility(data.AbilityMagicImmunity) {
                hasImmunity = true
            }
        case units.DamageRangedPhysical:
            if unit.HasAbility(data.AbilityLargeShield) {
                defenseRolls += 2
            }

            if unit.HasAbility(data.AbilityMissileImmunity) {
                hasImmunity = true
            }
        case units.DamageImmolation:
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

            defenseRolls += unit.GetResistances(data.UnitEnchantmentResistElements, data.UnitEnchantmentBless, data.UnitEnchantmentElementalArmor)

        case units.DamageFire:
            if unit.HasAbility(data.AbilityLargeShield) {
                defenseRolls += 2
            }

            if unit.HasAbility(data.AbilityMagicImmunity) || unit.HasAbility(data.AbilityFireImmunity) {
                hasImmunity = true
            }
        case units.DamageCold:
            if unit.HasAbility(data.AbilityLargeShield) {
                defenseRolls += 2
            }

            if unit.HasAbility(data.AbilityMagicImmunity) || unit.HasAbility(data.AbilityColdImmunity) {
                hasImmunity = true
            }
        case units.DamageThrown:
            if unit.HasAbility(data.AbilityLargeShield) {
                defenseRolls += 2
            }
    }

    if armorPiercing {
        defenseRolls /= 2
    }

    // after armor piercing, wall defense is applied
    switch damage {
        case units.DamageRangedMagical,
             units.DamageRangedPhysical,
             units.DamageMeleePhysical:
            defenseRolls += wallDefense
    }

    if hasImmunity {
        defenseRolls = 50
    }

    // log.Printf("Unit %v has %v defense", unit.Unit.GetName(), defenseRolls)

    defense := 0

    for range defenseRolls {
        if rand.N(100) < toDefend {
            defense += 1
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

// apply damage to each individual figure such that each figure gets to individually block damage.
// this could potentially allow a damage of 5 to destroy a unit with 4 figures of 1HP each
func (unit *ArmyUnit) ApplyAreaDamage(attackStrength int, damageType units.Damage, wallDefense int) int {
    totalDamage := 0
    health_per_figure := unit.Unit.GetMaxHealth() / unit.Unit.GetCount()

    for range unit.Figures() {
        damage := 0
        // FIXME: should this toHit be based on the unit's toHitMelee?
        toHit := 30
        for range attackStrength {
            if rand.N(100) < toHit {
                damage += 1
            }
        }

        defense := unit.ComputeDefense(damageType, false, wallDefense)
        // can't do more damage than a single figure has HP
        figureDamage := min(damage - defense, health_per_figure)
        if figureDamage > 0 {
            totalDamage += figureDamage
        }
    }

    totalDamage = min(totalDamage, unit.Unit.GetHealth())
    unit.TakeDamage(totalDamage)
    return totalDamage
}

// apply damage to lead figure, and if it dies then keep applying remaining damage to the next figure
func (unit *ArmyUnit) ApplyDamage(damage int, damageType units.Damage, armorPiercing bool, wallDefense int) int {
    taken := 0
    for damage > 0 && unit.Unit.GetHealth() > 0 {
        // compute defense, apply damage to lead figure. if lead figure dies, apply damage to next figure
        defense := unit.ComputeDefense(damageType, armorPiercing, wallDefense)
        damage -= defense
        if damage > 0 {
            health_per_figure := unit.Unit.GetMaxHealth() / unit.Unit.GetCount()
            healthLeft := unit.Unit.GetHealth() % unit.Figures()
            if healthLeft == 0 {
                healthLeft = health_per_figure
            }

            take := min(healthLeft, damage)
            unit.TakeDamage(take)
            damage -= take

            taken += take
        }
    }

    return taken
}

func (unit *ArmyUnit) InitializeSpells(allSpells spellbook.Spells, player *playerlib.Player) {
    unit.CastingSkill = 0
    unit.SpellCharges = make(map[spellbook.Spell]int)
    for _, ability := range unit.Unit.GetAbilities() {
        switch ability.Ability {
            case data.AbilityDoomBoltSpell:
                doomBolt := allSpells.FindByName("Doom Bolt")
                unit.SpellCharges[doomBolt] = int(ability.Value)
                // unit.Spells.AddSpell(doomBolt)
                // unit.CastingSkill += float32(doomBolt.CastCost)
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
        unit.Spells.AddAllSpells(player.KnownSpells)
        unit.SpellCharges = unit.Unit.GetSpellChargeSpells()
    }
}

// given the distance to the target in tiles, return the amount of range damage done
func (unit *ArmyUnit) ComputeRangeDamage(tileDistance int) int {

    toHit := unit.Unit.GetToHitMelee()

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

    damage := 0
    for range unit.Figures() {
        for range unit.Unit.GetRangedAttackPower() {
            if rand.N(100) < toHit {
                damage += 1
            }
        }
    }

    return damage
}

func (unit *ArmyUnit) ComputeMeleeDamage(fearFigure int) (int, bool) {

    if unit.GetMeleeAttackPower() == 0 {
        return 0, false
    }

    damage := 0
    hit := false
    for range unit.Figures() - fearFigure {
        // even if all figures fail to cause damage, it still counts as a hit for touch purposes
        hit = true
        for range unit.GetMeleeAttackPower() {
            if rand.N(100) < unit.Unit.GetToHitMelee() {
                damage += 1
            }
        }
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

    resistance := unit.GetResistance()

    resistance += unit.GetResistances(data.UnitEnchantmentBless, data.UnitEnchantmentResistMagic)

    for range unit.Figures() {
        if rand.N(10) + 1 > resistance {
            fear += 1
        }
    }

    return fear
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

type Army struct {
    Player *playerlib.Player
    Units []*ArmyUnit
    Auto bool
    Fled bool

    Enchantments []data.CombatEnchantment
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

// a number that mostly represents the strength of this army
func (army *Army) GetPower() int {
    power := 0

    for _, unit := range army.Units {
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
func (army *Army) AddUnit(unit CombatUnit){
    army.Units = append(army.Units, &ArmyUnit{
        Unit: unit,
        Facing: units.FacingDownRight,
        // Health: unit.GetMaxHealth(),
    })
}

func (army *Army) LayoutUnits(team Team){
    x := TownCenterX - 2
    y := 10

    facing := units.FacingDownRight

    if team == TeamAttacker {
        x = TownCenterX - 2
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

    TurnAttacker int
    TurnDefender int

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
}

func MakeCombatModel(cache *lbx.LbxCache, defendingArmy *Army, attackingArmy *Army, landscape CombatLandscape, plane data.Plane, zone ZoneType) *CombatModel {
    model := &CombatModel{
        Turn: TeamDefender,
        SelectedUnit: nil,
        Tiles: makeTiles(30, 30, landscape, plane, zone),
        TurnAttacker: 0,
        TurnDefender: 0,
        AttackingArmy: attackingArmy,
        DefendingArmy: defendingArmy,
        CurrentTurn: 0,
    }

    allSpells, err := spellbook.ReadSpellsFromCache(cache)
    if err != nil {
        log.Printf("Error: unable to read spells: %v", err)
        allSpells = spellbook.Spells{}
    }

    for _, unit := range defendingArmy.Units {
        unit.Model = model
        unit.Team = TeamDefender
        unit.RangedAttacks = unit.Unit.GetRangedAttacks()
        unit.InitializeSpells(allSpells, defendingArmy.Player)
        model.Tiles[unit.Y][unit.X].Unit = unit
    }

    for _, unit := range attackingArmy.Units {
        unit.Model = model
        unit.Team = TeamAttacker
        unit.RangedAttacks = unit.Unit.GetRangedAttacks()
        unit.InitializeSpells(allSpells, attackingArmy.Player)
        model.Tiles[unit.Y][unit.X].Unit = unit
    }

    model.NextTurn()
    model.SelectedUnit = model.ChooseNextUnit(TeamDefender)

    return model
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
            for i := 0; i < len(model.AttackingArmy.Units); i++ {
                model.TurnAttacker = (model.TurnAttacker + 1) % len(model.AttackingArmy.Units)
                unit := model.AttackingArmy.Units[model.TurnAttacker]
                if unit.LastTurn < model.CurrentTurn {
                    unit.Paths = make(map[image.Point]pathfinding.Path)
                    return unit
                }
            }
            return nil
        case TeamDefender:
            for i := 0; i < len(model.DefendingArmy.Units); i++ {
                model.TurnDefender = (model.TurnDefender + 1) % len(model.DefendingArmy.Units)
                unit := model.DefendingArmy.Units[model.TurnDefender]
                if unit.LastTurn < model.CurrentTurn {
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

    /* reset movement */
    for _, unit := range model.DefendingArmy.Units {
        unit.ResetTurnData()
    }

    for _, unit := range model.AttackingArmy.Units {
        // increase collateral damage to the town for each unit that is within the town area
        if model.InsideTown(unit.X, unit.Y) {
            model.CollateralDamage += 1
        }

        unit.ResetTurnData()
    }
}

func (model *CombatModel) computePath(x1 int, y1 int, x2 int, y2 int, canTraverseWall bool) (pathfinding.Path, bool) {

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

    path, ok = model.computePath(unit.X, unit.Y, x, y, unit.CanTraverseWall())
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
    _, ok := model.FindPath(unit, x, y)
    return ok
}

func (model *CombatModel) GetObserver() CombatObserver {
    return &model.Observer
}

func (model *CombatModel) IsEnchantmentActive(enchantment data.CombatEnchantment, team Team) bool {
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

func (model *CombatModel) AddEnchantment(player *playerlib.Player, enchantment data.CombatEnchantment) bool {
    if player == model.DefendingArmy.Player {
        return model.DefendingArmy.AddEnchantment(enchantment)
    } else {
        return model.AttackingArmy.AddEnchantment(enchantment)
    }
}

func (model *CombatModel) addNewUnit(player *playerlib.Player, x int, y int, unit units.Unit, facing units.Facing) {
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
        LastTurn: model.CurrentTurn-1,
    }

    model.Tiles[y][x].Unit = &newUnit

    if player == model.DefendingArmy.Player {
        newUnit.Team = TeamDefender
        model.DefendingArmy.Units = append(model.DefendingArmy.Units, &newUnit)
    } else {
        newUnit.Team = TeamAttacker
        model.AttackingArmy.Units = append(model.AttackingArmy.Units, &newUnit)
    }
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

func (model *CombatModel) FindEmptyTile() (int, int, error) {

    middleX := len(model.Tiles[0]) / 2
    middleY := len(model.Tiles) / 2

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
        strength := int(attacker.Unit.GetAbilityValue(data.AbilityFireBreath))
        hit = true

        damage = append(damage, func(){
            fireDamage := defender.ApplyDamage(strength, units.DamageFire, false, 0)
            model.AddLogEvent(fmt.Sprintf("%v uses fire breath on %v for %v damage", attacker.Unit.GetName(), defender.Unit.GetName(), fireDamage))
            // damage += fireDamage
            model.Observer.FireBreathAttack(attacker, defender, fireDamage)
        })
    }

    if attacker.HasAbility(data.AbilityLightningBreath) {
        strength := int(attacker.Unit.GetAbilityValue(data.AbilityLightningBreath))
        hit = true

        damage = append(damage, func(){
            lightningDamage := defender.ApplyDamage(strength, units.DamageRangedMagical, true, 0)
            model.AddLogEvent(fmt.Sprintf("%v uses lightning breath on %v for %v damage", attacker.Unit.GetName(), defender.Unit.GetName(), lightningDamage))
            // damage += lightningDamage
            model.Observer.LightningBreathAttack(attacker, defender, lightningDamage)
        })
    }

    return damage, hit
}

func (model *CombatModel) doGazeAttack(attacker *ArmyUnit, defender *ArmyUnit) (int, bool) {
    // FIXME: take into account the attack strength of the unit, and modifiers from spells/magic nodes

    damage := 0
    hit := false
    if attacker.HasAbility(data.AbilityStoningGaze) {
        if !defender.HasAbility(data.AbilityStoningImmunity) && !defender.HasAbility(data.AbilityMagicImmunity) {
            resistance := int(attacker.Unit.GetAbilityValue(data.AbilityStoningGaze))

            stoneDamage := 0

            for range defender.Figures() {
                if rand.N(10) + 1 > defender.GetResistance() - resistance {
                    stoneDamage += defender.Unit.GetHitPoints()
                }
            }

            // FIXME: this should be irreversable damage
            damage += stoneDamage
            hit = true

            model.Observer.StoneGazeAttack(attacker, defender, stoneDamage)

            model.AddLogEvent(fmt.Sprintf("%v uses stone gaze on %v for %v damage", attacker.Unit.GetName(), defender.Unit.GetName(), stoneDamage))
        }
    }

    if attacker.HasAbility(data.AbilityDeathGaze) {
        if !defender.HasAbility(data.AbilityDeathImmunity) && !defender.HasAbility(data.AbilityMagicImmunity) {
            resistance := int(attacker.Unit.GetAbilityValue(data.AbilityDeathGaze))

            deathDamage := 0

            for range defender.Figures() {
                if rand.N(10) + 1 > defender.GetResistance() - resistance {
                    deathDamage += defender.Unit.GetHitPoints()
                }
            }

            damage += deathDamage
            hit = true

            model.Observer.DeathGazeAttack(attacker, defender, deathDamage)

            model.AddLogEvent(fmt.Sprintf("%v uses death gaze on %v for %v damage", attacker.Unit.GetName(), defender.Unit.GetName(), deathDamage))
        }
    }

    if attacker.HasAbility(data.AbilityDoomGaze) {
        doomDamage := int(attacker.Unit.GetAbilityValue(data.AbilityDoomGaze))
        damage += doomDamage
        hit = true
        model.Observer.DoomGazeAttack(attacker, defender, doomDamage)
        model.AddLogEvent(fmt.Sprintf("%v uses doom gaze on %v for %v damage", attacker.Unit.GetName(), defender.Unit.GetName(), doomDamage))
    }

    return damage, hit
}

func (model *CombatModel) doThrowAttack(attacker *ArmyUnit, defender *ArmyUnit) (int, bool) {
    if attacker.HasAbility(data.AbilityThrown) {
        strength := int(attacker.Unit.GetAbilityValue(data.AbilityThrown))
        damage := 0
        for range attacker.Figures() {
            if rand.N(100) < attacker.Unit.GetToHitMelee() {
                // damage += defender.ApplyDamage(strength, units.DamageThrown, false)
                damage += strength
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
        for range int(attacker.Unit.GetAbilityValue(data.AbilityPoisonTouch)) {
            if rand.N(10) + 1 > defender.GetResistance() {
                damage += 1
            }
        }

        damageFuncs = append(damageFuncs, func(){
            defender.TakeDamage(damage)
            model.Observer.PoisonTouchAttack(attacker, defender, damage)
            model.AddLogEvent(fmt.Sprintf("%v is poisoned for %v damage. HP now %v", defender.Unit.GetName(), damage, defender.Unit.GetHealth()))
        })
    }

    if attacker.HasAbility(data.AbilityLifeSteal) {
        if !defender.HasAbility(data.AbilityDeathImmunity) && !defender.HasAbility(data.AbilityMagicImmunity) {
            modifier := int(attacker.Unit.GetAbilityValue(data.AbilityLifeSteal))
            // if vampiric, modifier will just be 0
            damage := 0
            defenderResistance := defender.GetResistance() + defender.GetResistances(data.UnitEnchantmentResistMagic, data.UnitEnchantmentBless, data.UnitEnchantmentRighteousness)

            for range attacker.Figures() - fearFigure {
                more := rand.N(10) + 1 - (defenderResistance + modifier)
                if more > 0 {
                    damage += more
                }
            }

            if damage > 0 {
                // cannot steal more life than the target has
                damage = min(damage, defender.Unit.GetHealth())

                damageFuncs = append(damageFuncs, func(){
                    // FIXME: if the unit dies they can become undead
                    defender.TakeDamage(damage)
                    attacker.Heal(damage)
                    model.AddLogEvent(fmt.Sprintf("%v steals %v life from %v. HP now %v", attacker.Unit.GetName(), damage, defender.Unit.GetName(), defender.Unit.GetHealth()))

                    model.Observer.LifeStealTouchAttack(attacker, defender, damage)
                })
            }
        }
    }

    if attacker.HasAbility(data.AbilityStoningTouch) {
        if !defender.HasAbility(data.AbilityStoningImmunity) && !defender.HasAbility(data.AbilityMagicImmunity) {
            damage := 0

            defenderResistance := defender.GetResistance() + defender.GetResistances(data.UnitEnchantmentElementalArmor, data.UnitEnchantmentResistElements, data.UnitEnchantmentResistMagic)

            modifier := int(attacker.Unit.GetAbilityValue(data.AbilityStoningTouch))

            // for each failed resistance roll, the defender takes damage equal to one figure's hit points
            for range attacker.Figures() - fearFigure {
                if rand.N(10) + 1 > defenderResistance - modifier {
                    damage += defender.Unit.GetHitPoints()
                }
            }

            damageFuncs = append(damageFuncs, func(){
                defender.TakeDamage(damage)

                model.AddLogEvent(fmt.Sprintf("%v turns %v to stone for %v damage. HP now %v", attacker.Unit.GetName(), defender.Unit.GetName(), damage, defender.Unit.GetHealth()))

                model.Observer.StoningTouchAttack(attacker, defender, damage)
            })
        }
    }

    if attacker.HasAbility(data.AbilityDispelEvil) {
        immune := true

        if defender.Unit.GetRace() == data.RaceFantastic {
            if defender.Unit.GetRealm() == data.ChaosMagic || defender.Unit.GetRealm() == data.DeathMagic {
                immune = false
            }
        }

        if defender.Unit.IsUndead() {
            immune = false
        }

        if defender.HasAbility(data.AbilityMagicImmunity) {
            immune = true
        }

        if !immune {
            damage := 0

            defenderResistance := defender.GetResistance()
            if defender.Unit.IsUndead() {
                defenderResistance -= 9
            } else {
                defenderResistance -= 4
            }

            defenderResistance += defender.GetResistances(data.UnitEnchantmentResistMagic)

            for range attacker.Figures() - fearFigure {
                if rand.N(10) + 1 > defenderResistance {
                    damage += defender.Unit.GetHitPoints()
                }
            }

            damageFuncs = append(damageFuncs, func(){
                defender.TakeDamage(damage)
                model.AddLogEvent(fmt.Sprintf("%v dispels evil from %v for %v damage. HP now %v", attacker.Unit.GetName(), defender.Unit.GetName(), damage, defender.Unit.GetHealth()))

                model.Observer.DispelEvilTouchAttack(attacker, defender, damage)
            })
        }
    }

    if attacker.HasAbility(data.AbilityDeathTouch) {
        if !defender.HasAbility(data.AbilityDeathImmunity) && !defender.HasAbility(data.AbilityMagicImmunity) {
            damage := 0
            defenderResistance := defender.GetResistance() + defender.GetResistances(data.UnitEnchantmentResistMagic, data.UnitEnchantmentBless, data.UnitEnchantmentRighteousness)
            modifier := 3

            for range attacker.Figures() - fearFigure {
                if rand.N(10) + 1 > defenderResistance - modifier {
                    damage += defender.Unit.GetHitPoints()
                }
            }

            damageFuncs = append(damageFuncs, func(){
                defender.TakeDamage(damage)

                model.AddLogEvent(fmt.Sprintf("%v uses death touch on %v for %v damage. HP now %v", attacker.Unit.GetName(), defender.Unit.GetName(), damage, defender.Unit.GetHealth()))

                model.Observer.DeathTouchAttack(attacker, defender, damage)
            })
        }
    }

    if attacker.Unit.HasItemAbility(data.ItemAbilityDestruction) {
        if !defender.HasAbility(data.AbilityMagicImmunity) {
            defenderResistance := defender.GetResistance() + defender.GetResistances(
                data.UnitEnchantmentResistMagic, data.UnitEnchantmentBless,
                data.UnitEnchantmentRighteousness, data.UnitEnchantmentElementalArmor,
                data.UnitEnchantmentResistElements)

            damage := 0
            for range attacker.Figures() - fearFigure {
                if rand.N(10) + 1 > defenderResistance {
                    damage += defender.Unit.GetHitPoints()
                }
            }

            damageFuncs = append(damageFuncs, func(){
                defender.TakeDamage(damage)
                model.AddLogEvent(fmt.Sprintf("%v uses destruction on %v for %v damage. HP now %v", attacker.Unit.GetName(), defender.Unit.GetName(), damage, defender.Unit.GetHealth()))

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
        hurt := defender.ApplyAreaDamage(immolationDamage, units.DamageImmolation, 0)
        model.AddLogEvent(fmt.Sprintf("%v is immolated for %v damage. HP now %v", defender.Unit.GetName(), hurt, defender.Unit.GetHealth()))
    }
}

func (model *CombatModel) ApplyMeleeDamage(attacker *ArmyUnit, defender *ArmyUnit, damage int) {
    hurt := defender.ApplyDamage(damage, units.DamageMeleePhysical, false, model.ComputeWallDefense(attacker, defender))
    model.AddLogEvent(fmt.Sprintf("%v damage roll %v, %v took %v damage. HP now %v", attacker.Unit.GetName(), damage, defender.Unit.GetName(), hurt, defender.Unit.GetHealth()))
}

func (model *CombatModel) ApplyWallOfFireDamage(defender *ArmyUnit) {
    model.ApplyImmolationDamage(defender, 5)
    model.Observer.WallOfFire(defender, 5)
}

func (model *CombatModel) canMeleeAttack(attacker *ArmyUnit, defender *ArmyUnit) bool {
    if attacker.MovesLeft.LessThanEqual(fraction.FromInt(0)) {
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

    if defender.Unit.IsFlying() && !attacker.Unit.IsFlying() {
        // a unit with Thrown can attack a flying unit
        if attacker.HasAbility(data.AbilityThrown) ||
           attacker.HasAbility(data.AbilityFireBreath) ||
           attacker.HasAbility(data.AbilityLightningBreath) {
            return true
        }
        return false
    }

    if attacker.Team == defender.Team {
        return false
    }

    return true
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

                gazeDamage, hit := model.doGazeAttack(attacker, defender)
                if hit {
                    immolationDamage += model.immolationDamage(attacker, defender)
                    if attacker.Unit.CanTouchAttack(units.DamageMeleePhysical) {
                        damageFuncs = append(damageFuncs, model.doTouchAttack(attacker, defender, 0)...)
                    }
                }

                if throwDamage > 0 {
                    damage := defender.ApplyDamage(throwDamage, units.DamageThrown, attacker.HasAbility(data.AbilityArmorPiercing), 0)
                    model.Observer.ThrowAttack(attacker, defender, damage)
                    model.AddLogEvent(fmt.Sprintf("%v throws %v at %v. HP now %v", attacker.Unit.GetName(), damage, defender.Unit.GetName(), defender.Unit.GetHealth()))
                }

                model.ApplyImmolationDamage(defender, immolationDamage)
                for _, f := range damageFuncs {
                    f()
                }

                defender.TakeDamage(gazeDamage)

            case 1:
                immolationDamage := 0
                damageFuncs := []func(){}

                gazeDamage, hit := model.doGazeAttack(defender, attacker)
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
                attacker.TakeDamage(gazeDamage)

            case 2:

                // if attacker is outside the wall of fire and the defender is inside (or vice-versa), then both side take immolation damage.
                // if either side is flying then they do not take damage.
                // for this to be false, either both are inside the wall of fire, or both are outside.
                if model.InsideWallOfFire(defender.X, defender.Y) != model.InsideWallOfFire(attacker.X, attacker.Y) {
                    if !attacker.Unit.IsFlying() {
                        model.ApplyWallOfFireDamage(attacker)
                    }

                    if !defender.Unit.IsFlying() {
                        model.ApplyWallOfFireDamage(defender)
                    }
                }

            case 3:
                if defender.HasAbility(data.AbilityCauseFear) {
                    attackerFear = attacker.CauseFear()
                    model.AddLogEvent(fmt.Sprintf("%v causes fear in %v for %v figures", defender.Unit.GetName(), attacker.Unit.GetName(), attackerFear))
                    model.Observer.CauseFear(defender, attacker, attackerFear)
                }
            case 4:
                if attacker.HasAbility(data.AbilityFirstStrike) && !defender.HasAbility(data.AbilityNegateFirstStrike) {
                    attackerDamage, hit := attacker.ComputeMeleeDamage(attackerFear)

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
                if attacker.HasAbility(data.AbilityCauseFear) || attacker.HasEnchantment(data.UnitEnchantmentCloakOfFear) {
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
                    attackerDamage, hit := attacker.ComputeMeleeDamage(attackerFear)

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

                // defender does counter-attack
                for range counters {
                    defenderDamage, hit := defender.ComputeMeleeDamage(defenderFear)

                    if hit {
                        model.Observer.MeleeAttack(defender, attacker, defenderDamage)
                        attackerMeleeDamage += defenderDamage
                        attackerImmolationDamage += model.immolationDamage(defender, attacker)
                        if defender.Unit.CanTouchAttack(units.DamageMeleePhysical) {
                            damageFuncs = append(damageFuncs, model.doTouchAttack(defender, attacker, defenderFear)...)
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
        if defender.Unit.GetHealth() <= 0 {
            model.AddLogEvent(fmt.Sprintf("%v is killed", defender.Unit.GetName()))
            model.RemoveUnit(defender)
            end = true
            model.Observer.UnitKilled(defender)
        }

        if attacker.Unit.GetHealth() <= 0 {
            model.AddLogEvent(fmt.Sprintf("%v is killed", attacker.Unit.GetName()))
            model.RemoveUnit(attacker)
            end = true
            model.Observer.UnitKilled(attacker)
        }

        if end {
            break
        }
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

func (model *CombatModel) IsAIControlled(unit *ArmyUnit) bool {
    if unit.Team == TeamDefender {
        return model.DefendingArmy.IsAI()
    } else {
        return model.AttackingArmy.IsAI()
    }
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

    for _, unit := range attackingArmy.Units {
        unit.Model = &fakeModel
    }

    for _, unit := range defendingArmy.Units {
        unit.Model = &fakeModel
    }

    attackingPower := attackingArmy.GetPower()
    defendingPower := defendingArmy.GetPower()

    log.Printf("strategic combat: attacking power: %v, defending power: %v", attackingPower, defendingPower)

    // FIXME: Allow fleeing?

    if attackingPower > defendingPower {
        for _, unit := range defendingArmy.Units {
            unit.TakeDamage(unit.Unit.GetMaxHealth())
        }

        return CombatStateAttackerWin, 0, len(defendingArmy.Units)
    } else {
        for _, unit := range attackingArmy.Units {
            unit.TakeDamage(unit.Unit.GetMaxHealth())
        }

        return CombatStateDefenderWin, len(attackingArmy.Units), 0
    }
}

func (model *CombatModel) flee(army *Army) {
    for _, unit := range army.Units {
        // FIXME: units unable to move always die

        // heroes have a 25% chance to die, normal units 50%
        chance := 50
        if unit.Unit.IsHero() {
            chance = 25
        }

        if rand.IntN(100) < chance {
            unit.TakeDamage(unit.Unit.GetHealth())
            model.RemoveUnit(unit)
            model.DiedWhileFleeing += 1
        }
    }
}
