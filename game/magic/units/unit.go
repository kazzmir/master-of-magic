package units

import (
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/building"
)

type Damage int
const (
    DamageNone Damage = iota
    DamageRangedMagical
    DamageRangedPhysical
    DamageRangedBoulder
    DamageMeleePhysical
)

type AttackSound int

const (
    AttackSoundNone AttackSound = iota
    AttackSoundMonster1
    AttackSoundMonster2
    AttackSoundMonster3
    AttackSoundMonster4
    AttackSoundMonsterVsNormal
    AttackSoundMonster5
    AttackSoundNormal
    AttackSoundWeak
)

type RangeAttackSound int
const (
    RangeAttackSoundNone RangeAttackSound = iota
    RangeAttackSoundFireball
    RangeAttackSoundArrow
    RangeAttackSoundSling
    RangeAttackSoundLaunch // for catapult or other similar things
)

func (sound RangeAttackSound) LbxIndex() int {
    switch sound {
        case RangeAttackSoundNone: return -1
        case RangeAttackSoundFireball: return 20
        case RangeAttackSoundArrow: return 17
        case RangeAttackSoundSling: return 18
        case RangeAttackSoundLaunch: return 15
    }

    return -1
}

func (sound AttackSound) LbxIndex() int {
    switch sound {
        case AttackSoundNone: return -1
        case AttackSoundMonster1: return 0
        case AttackSoundMonster2: return 1
        case AttackSoundMonster3: return 2
        case AttackSoundMonster4: return 3
        case AttackSoundMonsterVsNormal: return 4
        case AttackSoundMonster5: return 5
        case AttackSoundNormal: return 6
        case AttackSoundWeak: return 7
    }

    return -1
}

type MovementSound int

const (
    MovementSoundNone MovementSound = iota
    MovementSoundMarching
    MovementSoundHorse
    MovementSoundFly
    MovementSoundBigSteps
    MovementSoundMerge
    MovementSoundShuffle
)

func (sound MovementSound) LbxIndex() int {
    switch sound {
        case MovementSoundNone: return -1
        case MovementSoundMarching: return 5
        case MovementSoundHorse: return 6
        case MovementSoundFly: return 7
        case MovementSoundBigSteps: return 10
        case MovementSoundMerge: return 12
        case MovementSoundShuffle: return 9
    }

    return -1
}

type Facing int
const (
    FacingUp Facing = iota
    FacingUpRight
    FacingRight
    FacingDownRight
    FacingDown
    FacingDownLeft
    FacingLeft
    FacingUpLeft
)

type Unit struct {
    // icon on the overworld and in various ui's
    LbxFile string
    Index int

    // sprites to use in combat
    CombatLbxFile string
    // first index of combat tiles, order is always up, up-right, right, down-right, down, down-left, left, up-left
    CombatIndex int
    Name string
    Race data.Race
    Flying bool
    Swimming bool
    Abilities []Ability

    // fantastic units belong to a specific magic realm
    Realm data.MagicType

    RequiredBuildings []building.Building

    RangedAttackDamageType Damage

    AttackSound AttackSound
    MovementSound MovementSound
    RangeAttackSound RangeAttackSound

    // first sprite index in cmbmagic.lbx for the range attack
    RangeAttackIndex int

    // number of figures that are drawn in a single combat tile
    Count int

    // cost in terms of production to build the unit
    ProductionCost int

    MeleeAttackPower int
    MovementSpeed int
    Defense int
    Resistance int
    HitPoints int
    RangedAttackPower int
    RangedAttacks int
    UpkeepGold int
    UpkeepFood int
    UpkeepMana int
    // FIXME: add construction cost, building requirements to build this unit
    //  upkeep cost, how many figures appear in the battlefield, movement speed,
    //  attack power, ranged attack, defense, magic resistance, hit points, special power
}

func (unit *Unit) Equals(other Unit) bool {
    return unit.LbxFile == other.LbxFile && unit.Index == other.Index
}

func (unit *Unit) GetName() string {
    return unit.Name
}

func (unit *Unit) GetCombatLbxFile() string {
    return unit.CombatLbxFile
}

func (unit *Unit) GetCount() int {
    return unit.Count
}

func (unit *Unit) GetUpkeepGold() int {
    return unit.UpkeepGold
}

func (unit *Unit) GetUpkeepFood() int {
    return unit.UpkeepFood
}

func (unit *Unit) GetUpkeepMana() int {
    return unit.UpkeepMana
}

func (unit *Unit) GetMovementSpeed() int {
    return unit.MovementSpeed
}

func (unit *Unit) GetProductionCost() int {
    return unit.ProductionCost
}

func (unit *Unit) GetBaseMeleeAttackPower() int {
    return unit.GetMeleeAttackPower()
}

func (unit *Unit) GetMeleeAttackPower() int {
    return unit.MeleeAttackPower
}

func (unit *Unit) GetBaseRangedAttackPower() int {
    return unit.GetRangedAttackPower()
}

func (unit *Unit) GetRangedAttackPower() int {
    return unit.RangedAttackPower
}

func (unit *Unit) GetBaseDefense() int {
    return unit.Defense
}

func (unit *Unit) GetDefense() int {
    return unit.Defense
}

func (unit *Unit) GetResistance() int {
    return unit.Resistance
}

func (unit *Unit) GetBaseResistance() int {
    return unit.Resistance
}

func (unit *Unit) GetHitPoints() int {
    return unit.HitPoints
}

func (unit *Unit) GetBaseHitPoints() int {
    return unit.HitPoints
}

func (unit *Unit) GetAbilities() []Ability {
    return unit.Abilities
}

func (unit *Unit) HasAbility(ability Ability) bool {
    for _, check := range unit.Abilities {
        if check == ability {
            return true
        }
    }

    return false
}

func (unit *Unit) IsSettlers() bool {
    return unit.HasAbility(AbilityCreateOutpost)
}

/* maximum health is the number of figures * the number of hit points per figure
 */
func (unit *Unit) GetMaxHealth() int {
    return unit.HitPoints * unit.Count
}

func (unit *Unit) GetCombatRangeIndex(facing Facing) int {
    switch facing {
        case FacingUp: return unit.RangeAttackIndex + 0
        case FacingUpRight: return unit.RangeAttackIndex + 1
        case FacingRight: return unit.RangeAttackIndex + 2
        case FacingDownRight: return unit.RangeAttackIndex + 3
        case FacingDown: return unit.RangeAttackIndex + 4
        case FacingDownLeft: return unit.RangeAttackIndex + 5
        case FacingLeft: return unit.RangeAttackIndex + 6
        case FacingUpLeft: return unit.RangeAttackIndex + 7
    }

    return unit.RangeAttackIndex
}

func (unit *Unit) GetCombatIndex(facing Facing) int {
    switch facing {
        case FacingUp: return unit.CombatIndex + 0
        case FacingUpRight: return unit.CombatIndex + 1
        case FacingRight: return unit.CombatIndex + 2
        case FacingDownRight: return unit.CombatIndex + 3
        case FacingDown: return unit.CombatIndex + 4
        case FacingDownLeft: return unit.CombatIndex + 5
        case FacingLeft: return unit.CombatIndex + 6
        case FacingUpLeft: return unit.CombatIndex + 7
    }

    return unit.CombatIndex
}

func (unit *Unit) IsNone() bool {
    return unit.Index == -1
}

func (unit *Unit) String() string {
    return unit.Name
}

var UnitNone Unit = Unit{
    Index: -1,
}

var LizardSpearmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 0,
    Name: "Spearmen",
    CombatLbxFile: "figures9.lbx",
    CombatIndex: 0,
    Race: data.RaceLizard,
    UpkeepFood: 1,
    ProductionCost: 10,
    Count: 8,
    MovementSpeed: 1,
    Swimming: true,
    MeleeAttackPower: 1,
    Defense: 3,
    Resistance: 4,
    HitPoints: 2,
}

var LizardSwordsmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 1,
    Name: "Swordsmen",
    CombatLbxFile: "figures9.lbx",
    CombatIndex: 8,
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    ProductionCost: 20,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    Swimming: true,
    MeleeAttackPower: 3,
    Defense: 3,
    Resistance: 4,
    HitPoints: 4,
    Abilities: []Ability{AbilityLargeShield},
    Race: data.RaceLizard,
}

var LizardHalberdiers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 2,
    Name: "Halberdiers",
    CombatLbxFile: "figures9.lbx",
    CombatIndex: 16,
    RequiredBuildings: []building.Building{building.BuildingArmory},
    Race: data.RaceLizard,
    Abilities: []Ability{AbilityNegateFirstStrike},
    ProductionCost: 40,
    UpkeepGold: 1,
    UpkeepFood: 1,
    MovementSpeed: 1,
    Swimming: true,
    MeleeAttackPower: 4,
    Defense: 4,
    Resistance: 4,
    HitPoints: 2,
    Count: 6,
}

var LizardJavelineers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 3,
    Name: "Javelineers",
    CombatLbxFile: "figures9.lbx",
    CombatIndex: 24,
    RequiredBuildings: []building.Building{building.BuildingFightersGuild},
    Race: data.RaceLizard,
    RangedAttacks: 6,
    ProductionCost: 80,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 6,
    MeleeAttackPower: 4,
    RangedAttackPower: 3,
    RangedAttackDamageType: DamageRangedPhysical,
    Defense: 4,
    Resistance: 5,
    HitPoints: 2,
    MovementSpeed: 1,
    Swimming: true,
}

var LizardShamans Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 4,
    Name: "Shamans",
    CombatLbxFile: "figures9.lbx",
    CombatIndex: 32,
    RequiredBuildings: []building.Building{building.BuildingShrine},
    Race: data.RaceLizard,
    RangedAttacks: 4,
    MeleeAttackPower: 2,
    RangedAttackPower: 2,
    RangedAttackDamageType: DamageRangedMagical,
    ProductionCost: 60,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 1,
    Swimming: true,
    Defense: 3,
    Resistance: 6,
    HitPoints: 2,
    Abilities: []Ability{AbilityHealer, AbilityPurify},
}

var LizardSettlers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 5,
    Name: "Settlers",
    CombatLbxFile: "figures9.lbx",
    CombatIndex: 40,
    MovementSpeed: 1,
    Defense: 2,
    Resistance: 4,
    HitPoints: 20,
    Count: 1,
    ProductionCost: 120,
    Swimming: true,
    UpkeepGold: 3,
    UpkeepFood: 1,
    Abilities: []Ability{AbilityCreateOutpost},
    Race: data.RaceLizard,
}

var DragonTurtle Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 6,
    Name: "Dragon Turtle",
    CombatLbxFile: "figures9.lbx",
    CombatIndex: 48,
    RequiredBuildings: []building.Building{building.BuildingArmorersGuild, building.BuildingStables},
    Race: data.RaceLizard,
    // fire breath 5
    Abilities: []Ability{AbilityFireBreath},
    ProductionCost: 100,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 1,
    MovementSpeed: 2,
    Swimming: true,
    MeleeAttackPower: 10,
    Defense: 8,
    Resistance: 8,
    HitPoints: 15,
}

var NomadSpearmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 7,
    CombatLbxFile: "figures9.lbx",
    CombatIndex: 56,
    Name: "Spearmen",
    ProductionCost: 10,
    UpkeepFood: 1,
    Count: 8,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    Defense: 2,
    Resistance: 4,
    HitPoints: 1,
    Race: data.RaceNomad,
}

var NomadSwordsmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 8,
    CombatLbxFile: "figures9.lbx",
    CombatIndex: 64,
    Name: "Swordsmen",
    ProductionCost: 20,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 3,
    Defense: 2,
    Resistance: 4,
    HitPoints: 1,
    Abilities: []Ability{AbilityLargeShield},
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceNomad,
}

var NomadBowmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 9,
    CombatLbxFile: "figures9.lbx",
    CombatIndex: 72,
    Name: "Bowmen",
    ProductionCost: 30,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    RangedAttackPower: 1,
    RangedAttacks: 8,
    RangedAttackDamageType: DamageRangedPhysical,
    Defense: 1,
    Resistance: 4,
    HitPoints: 1,
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSawmill},
    Race: data.RaceNomad,
}

var NomadPriest Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 10,
    CombatLbxFile: "figures9.lbx",
    CombatIndex: 80,
    Name: "Priests",
    ProductionCost: 100,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 1,
    MeleeAttackPower: 3,
    RangedAttackPower: 4,
    RangedAttacks: 4,
    // nature
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 4,
    Resistance: 7,
    HitPoints: 1,
    // healing spell 1x
    Abilities: []Ability{AbilityHealer, AbilityPurify, AbilityHealingSpell},
    RequiredBuildings: []building.Building{building.BuildingParthenon},
    Race: data.RaceNomad,
}

// what is units2.lbx index 11?
// its some nomad unit holding a sword or something

var NomadSettlers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 12,
    CombatLbxFile: "figures9.lbx",
    CombatIndex: 96,
    Name: "Settlers",
    ProductionCost: 60,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 1,
    MovementSpeed: 1,
    Defense: 1,
    Resistance: 4,
    HitPoints: 10,
    Abilities: []Ability{AbilityCreateOutpost},
    Race: data.RaceNomad,
}

var NomadHorsebowemen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 13,
    CombatLbxFile: "figures9.lbx",
    CombatIndex: 104,
    Name: "Horsebowmen",
    ProductionCost: 60,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 2,
    MeleeAttackPower: 4,
    RangedAttackPower: 2,
    RangedAttacks: 8,
    RangedAttackDamageType: DamageRangedPhysical,
    Defense: 2,
    Resistance: 4,
    HitPoints: 3,
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingStables},
    Race: data.RaceNomad,
}

var NomadPikemen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 14,
    CombatLbxFile: "figures9.lbx",
    CombatIndex: 112,
    Name: "Pikemen",
    ProductionCost: 80,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 8,
    MovementSpeed: 1,
    MeleeAttackPower: 5,
    Defense: 3,
    Resistance: 5,
    HitPoints: 1,
    Abilities: []Ability{AbilityNegateFirstStrike, AbilityArmorPiercing},
    RequiredBuildings: []building.Building{building.BuildingFightersGuild},
    Race: data.RaceNomad,
}

var NomadRangers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 15,
    CombatLbxFile: "figure10.lbx",
    CombatIndex: 0,
    Name: "Rangers",
    ProductionCost: 120,
    UpkeepGold: 3,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 2,
    MeleeAttackPower: 4,
    RangedAttackPower: 3,
    RangedAttacks: 8,
    RangedAttackDamageType: DamageRangedPhysical,
    Defense: 4,
    Resistance: 6,
    HitPoints: 2,
    Abilities: []Ability{AbilityPathfinding},
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingAnimistsGuild},
    Race: data.RaceNomad,
}

var Griffin Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 16,
    CombatLbxFile: "figure10.lbx",
    CombatIndex: 8,
    Name: "Griffins",
    ProductionCost: 200,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 2,
    Flying: true,
    MovementSpeed: 2,
    MeleeAttackPower: 9,
    Defense: 5,
    Resistance: 7,
    HitPoints: 10,
    Abilities: []Ability{AbilityArmorPiercing, AbilityFirstStrike},
    RequiredBuildings: []building.Building{building.BuildingFantasticStable},
    Race: data.RaceNomad,
}

var OrcSpearmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 17,
    CombatLbxFile: "figure10.lbx",
    CombatIndex: 16,
    Name: "Spearmen",
    ProductionCost: 10,
    UpkeepFood: 1,
    Count: 8,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    Defense: 2,
    Resistance: 4,
    HitPoints: 1,
    Race: data.RaceOrc,
}

var OrcSwordsmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 18,
    CombatLbxFile: "figure10.lbx",
    CombatIndex: 24,
    Name: "Swordsmen",
    ProductionCost: 20,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 3,
    Defense: 2,
    Resistance: 4,
    HitPoints: 1,
    Abilities: []Ability{AbilityLargeShield},
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceOrc,
}

var OrcHalberdiers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 19,
    CombatLbxFile: "figure10.lbx",
    CombatIndex: 32,
    Name: "Halberdiers",
    ProductionCost: 40,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 4,
    Defense: 3,
    Resistance: 4,
    HitPoints: 1,
    RequiredBuildings: []building.Building{building.BuildingArmory},
    Race: data.RaceOrc,
}

var OrcBowmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 20,
    CombatLbxFile: "figure10.lbx",
    CombatIndex: 40,
    Name: "Bowmen",
    ProductionCost: 30,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    RangedAttackPower: 1,
    RangedAttacks: 8,
    RangedAttackDamageType: DamageRangedPhysical,
    Defense: 1,
    Resistance: 4,
    HitPoints: 1,
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSawmill},
    Race: data.RaceOrc,
}

var OrcCavalry Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 21,
    CombatLbxFile: "figure10.lbx",
    CombatIndex: 48,
    Name: "Calvary",
    ProductionCost: 40,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 2,
    MeleeAttackPower: 4,
    Defense: 2,
    Resistance: 4,
    HitPoints: 3,
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingStables},
    Race: data.RaceOrc,
}

var OrcShamans Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 22,
    CombatLbxFile: "figure10.lbx",
    CombatIndex: 56,
    Name: "Shamans",
    ProductionCost: 60,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 1,
    MeleeAttackPower: 2,
    // nature
    RangedAttackPower: 2,
    RangedAttacks: 4,
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 3,
    Resistance: 6,
    HitPoints: 1,
    Abilities: []Ability{AbilityHealer, AbilityPurify},
    RequiredBuildings: []building.Building{building.BuildingShrine},
    Race: data.RaceOrc,
}

var OrcMagicians Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 23,
    CombatLbxFile: "figure10.lbx",
    CombatIndex: 64,
    Name: "Magicians",
    ProductionCost: 120,
    UpkeepGold: 3,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    // chaos
    RangedAttackPower: 5,
    RangedAttacks: 4,
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 3,
    Resistance: 8,
    HitPoints: 1,
    // fireball spell 1x
    Abilities: []Ability{AbilityMissileImmunity, AbilityFireballSpell},
    RequiredBuildings: []building.Building{building.BuildingWizardsGuild},
    Race: data.RaceOrc,
}

var OrcEngineers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 24,
    CombatLbxFile: "figure10.lbx",
    CombatIndex: 72,
    Name: "Engineers",
    ProductionCost: 40,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    Defense: 1,
    Resistance: 4,
    HitPoints: 1,
    Abilities: []Ability{AbilityConstruction, AbilityWallCrusher},
    RequiredBuildings: []building.Building{building.BuildingBuildersHall},
    Race: data.RaceOrc,
}

var OrcSettlers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 25,
    CombatLbxFile: "figure10.lbx",
    CombatIndex: 80,
    Name: "Settlers",
    ProductionCost: 60,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 1,
    MovementSpeed: 1,
    Defense: 1,
    Resistance: 4,
    HitPoints: 10,
    Abilities: []Ability{AbilityCreateOutpost},
    Race: data.RaceOrc,
}

var WyvernRiders Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 26,
    CombatLbxFile: "figure10.lbx",
    CombatIndex: 88,
    Name: "Wyvern Riders",
    ProductionCost: 200,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 2,
    Flying: true,
    MovementSpeed: 3,
    MeleeAttackPower: 5,
    Defense: 5,
    Resistance: 7,
    HitPoints: 10,
    // poison touch 6
    Abilities: []Ability{AbilityPoisonTouch},
    RequiredBuildings: []building.Building{building.BuildingFantasticStable},
    Race: data.RaceOrc,
}

var TrollSpearmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 27,
    CombatLbxFile: "figure10.lbx",
    CombatIndex: 96,
    Name: "Spearmen",
    ProductionCost: 30,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 1,
    MeleeAttackPower: 3,
    Defense: 2,
    Resistance: 7,
    HitPoints: 4,
    Abilities: []Ability{AbilityRegeneration},
    Race: data.RaceTroll,
}

var TrollSwordsmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 28,
    CombatLbxFile: "figure10.lbx",
    CombatIndex: 104,
    Name: "Swordsmen",
    ProductionCost: 60,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 1,
    MeleeAttackPower: 5,
    Defense: 2,
    Resistance: 7,
    HitPoints: 4,
    Abilities: []Ability{AbilityLargeShield, AbilityRegeneration},
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceTroll,
}

var TrollHalberdiers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 29,
    CombatLbxFile: "figure10.lbx",
    CombatIndex: 112,
    Name: "Halberdiers",
    ProductionCost: 120,
    UpkeepGold: 3,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 1,
    MeleeAttackPower: 6,
    Defense: 3,
    Resistance: 7,
    HitPoints: 4,
    Abilities: []Ability{AbilityNegateFirstStrike, AbilityRegeneration},
    RequiredBuildings: []building.Building{building.BuildingArmory},
    Race: data.RaceTroll,
}

var TrollShamans Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 30,
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 0,
    Name: "Shamans",
    ProductionCost: 180,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 1,
    MeleeAttackPower: 4,
    // nature
    RangedAttackPower: 2,
    RangedAttacks: 4,
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 3,
    Resistance: 8,
    HitPoints: 4,
    Abilities: []Ability{AbilityHealer, AbilityPurify, AbilityRegeneration},
    RequiredBuildings: []building.Building{building.BuildingShrine},
    Race: data.RaceTroll,
}

var TrollSettlers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 31,
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 8,
    Name: "Settlers",
    ProductionCost: 180,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 1,
    MovementSpeed: 1,
    Defense: 1,
    Resistance: 7,
    HitPoints: 40,
    Abilities: []Ability{AbilityCreateOutpost, AbilityRegeneration},
    Race: data.RaceTroll,
}

var WarTrolls Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 32,
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 16,
    Name: "War Trolls",
    ProductionCost: 160,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 2,
    MeleeAttackPower: 8,
    Defense: 4,
    Resistance: 8,
    HitPoints: 5,
    Abilities: []Ability{AbilityRegeneration},
    RequiredBuildings: []building.Building{building.BuildingFightersGuild},
    Race: data.RaceTroll,
}

var WarMammoths Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 33,
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 24,
    Name: "War Mammoths",
    ProductionCost: 240,
    UpkeepGold: 5,
    UpkeepFood: 1,
    Count: 2,
    MovementSpeed: 2,
    MeleeAttackPower: 10,
    Defense: 6,
    Resistance: 9,
    HitPoints: 12,
    Abilities: []Ability{AbilityWallCrusher, AbilityFirstStrike},
    RequiredBuildings: []building.Building{building.BuildingArmorersGuild, building.BuildingStables},
    Race: data.RaceTroll,
}

var MagicSpirit Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 34,
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 32,
    Name: "Magic Spirit",
    Realm: data.ArcaneMagic,
    UpkeepMana: 1,
    MovementSpeed: 1,
    Swimming: true,
    Count: 1,
    MeleeAttackPower: 5,
    Defense: 4,
    Resistance: 8,
    HitPoints: 10,
    Abilities: []Ability{AbilityMeld, AbilityNonCorporeal},
    Race: data.RaceFantastic,
}

var HellHounds Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 35,
    Name: "Hell Hounds",
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 40,
    UpkeepMana: 1,
    MovementSpeed: 2,
    Realm: data.ChaosMagic,
    MeleeAttackPower: 3,
    Count: 4,
    Defense: 2,
    Resistance: 6,
    HitPoints: 4,
    // fire breath 3
    // to hit +10%
    Abilities: []Ability{AbilityFireBreath, AbilityToHit},
    Race: data.RaceFantastic,
}

var Gargoyle Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 36,
    Name: "Gargoyles",
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 48,
    UpkeepMana: 5,
    Realm: data.ChaosMagic,
    MovementSpeed: 2,
    Flying: true,
    MeleeAttackPower: 4,
    Count: 4,
    Defense: 8,
    Resistance: 7,
    HitPoints: 4,
    // tohit +10%
    Abilities: []Ability{AbilityToHit, AbilityPoisonImmunity, AbilityStoningImmunity},
    Race: data.RaceFantastic,
}

var FireGiant Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 37,
    Name: "Fire Giant",
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 56,
    MovementSpeed: 2,
    UpkeepMana: 3,
    Realm: data.ChaosMagic,
    Count: 1,
    MeleeAttackPower: 10,
    RangedAttackPower: 10,
    RangedAttackDamageType: DamageRangedBoulder,
    RangedAttacks: 2,
    HitPoints: 15,
    Defense: 5,
    Resistance: 7,
    // tohit +10%
    Abilities: []Ability{AbilityToHit, AbilityMountaineer, AbilityWallCrusher, AbilityFireImmunity},
    Race: data.RaceFantastic,
}

var FireElemental Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 38,
    Race: data.RaceFantastic,
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 64,
    Realm: data.ChaosMagic,
    Name: "Fire Elemental",
    Count: 1,
    MovementSpeed: 1,
    MeleeAttackPower: 12,
    Defense: 4,
    Resistance: 6,
    HitPoints: 10,
    Abilities: []Ability{AbilityFireImmunity, AbilityPoisonImmunity, AbilityStoningImmunity},
}

var ChaosSpawn Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 39,
    Name: "Chaos Spawn",
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 72,
    Realm: data.ChaosMagic,
    MovementSpeed: 1,
    UpkeepMana: 12,
    Flying: true,
    Count: 1,
    MeleeAttackPower: 1,
    Defense: 6,
    Resistance: 10,
    HitPoints: 15,
    // poison touch 4
    // doom gaze 4
    // death gaze 4
    // stoning gaze 4
    Abilities: []Ability{AbilityCauseFear, AbilityPoisonTouch, AbilityDoomGaze, AbilityDeathGaze, AbilityStoningGaze},
    Race: data.RaceFantastic,
}

var Chimeras Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 40,
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 80,
    Realm: data.ChaosMagic,
    UpkeepMana: 10,
    MovementSpeed: 2,
    Flying: true,
    Count: 4,
    MeleeAttackPower: 7,
    Defense: 5,
    Resistance: 8,
    HitPoints: 8,
    // tohit +10%
    // firebreath 4
    Abilities: []Ability{AbilityFireBreath, AbilityToHit},
    Race: data.RaceFantastic,
}

var DoomBat Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 41,
    Name: "Doom Bat",
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 88,
    MovementSpeed: 4,
    UpkeepMana: 8,
    Realm: data.ChaosMagic,
    Flying: true,
    MeleeAttackPower: 10,
    Defense: 5,
    Resistance: 9,
    Count: 1,
    HitPoints: 20,
    // tohit +10%
    Abilities: []Ability{AbilityToHit, AbilityImmolation},
    Race: data.RaceFantastic,
}

var Efreet Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 42,
    Name: "Efreet",
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 96,
    MovementSpeed: 3,
    Realm: data.ChaosMagic,
    UpkeepMana: 15,
    Flying: true,
    Count: 1,
    // caster 20
    MeleeAttackPower: 9,
    RangedAttackPower: 9,
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 7,
    Resistance: 10,
    HitPoints: 12,
    // tohit +20%
    // caster 20
    Abilities: []Ability{AbilityFireImmunity, AbilityToHit, AbilityCaster},
    Race: data.RaceFantastic,
}

// FIXME: hydra has 9 virtual figures, one for each head
var Hydra Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 43,
    Name: "Hydra",
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 104,
    UpkeepMana: 14,
    Count: 1,
    Realm: data.ChaosMagic,
    MovementSpeed: 1,
    MeleeAttackPower: 6,
    Defense: 4,
    Resistance:11,
    HitPoints: 10,
    // tohit +10%
    // fire breath 5
    Abilities: []Ability{AbilityRegeneration, AbilityToHit, AbilityFireBreath},
    Race: data.RaceFantastic,
}

var GreatDrake Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 44,
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 112,
    Name: "Great Drake",
    Race: data.RaceFantastic,
    // fire breath 30, tohit 30
    Abilities: []Ability{AbilityFireBreath, AbilityToHit},
    Count: 1,
    Realm: data.ChaosMagic,
    HitPoints: 30,
    Flying: true,
    AttackSound: AttackSoundMonster2,
    MovementSound: MovementSoundFly,
    MeleeAttackPower: 30,
    MovementSpeed: 2,
    UpkeepMana: 30,
    Defense: 10,
    Resistance: 12,
}

var Skeleton Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 45,
    CombatLbxFile: "figure12.lbx",
    CombatIndex: 0,
    Name: "Skeletons",
    UpkeepMana: 1,
    Count: 6,
    MovementSpeed: 3,
    MeleeAttackPower: 3,
    Defense: 4,
    Resistance: 5,
    HitPoints: 1,
    Abilities: []Ability{AbilityToHit, AbilityPoisonImmunity, AbilityMissileImmunity, AbilityIllusionsImmunity, AbilityColdImmunity, AbilityDeathImmunity},
    Race: data.RaceFantastic,
    Realm: data.DeathMagic,
}

var Ghoul Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 46,
    CombatLbxFile: "figure12.lbx",
    CombatIndex: 8,
    Name: "Ghouls",
    UpkeepMana: 1,
    Count: 4,
    MovementSpeed: 1,
    MeleeAttackPower: 4,
    Defense: 3,
    Resistance: 6,
    HitPoints: 3,
    // tohit +10%
    // poison touch 1
    Abilities: []Ability{AbilityToHit, AbilityCreateUndead, AbilityPoisonImmunity, AbilityIllusionsImmunity, AbilityColdImmunity, AbilityDeathImmunity, AbilityPoisonTouch},
    Realm: data.DeathMagic,
    Race: data.RaceFantastic,
}

var NightStalker Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 47,
    Name: "Night Stalker",
    CombatLbxFile: "figure12.lbx",
    CombatIndex: 16,
    UpkeepMana: 1,
    Count: 1,
    MovementSpeed: 2,
    MeleeAttackPower: 7,
    Defense: 3,
    Resistance: 8,
    HitPoints: 10,
    // tohit +10%
    // death gaze 2
    Abilities: []Ability{AbilityToHit, AbilityInvisibility, AbilityPoisonImmunity, AbilityIllusionsImmunity, AbilityColdImmunity, AbilityDeathImmunity, AbilityDeathGaze},
    Realm: data.DeathMagic,
    Race: data.RaceFantastic,
}

var WereWolf Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 48,
    CombatLbxFile: "figure12.lbx",
    CombatIndex: 24,
    Name: "WereWolves",
    UpkeepMana: 5,
    Count: 6,
    MovementSpeed: 2,
    MeleeAttackPower: 5,
    Defense: 1,
    Resistance: 6,
    HitPoints: 5,
    // tohit +10%
    Abilities: []Ability{AbilityRegeneration, AbilityToHit, AbilityPoisonImmunity, AbilityIllusionsImmunity, AbilityColdImmunity, AbilityDeathImmunity, AbilityWeaponImmunity},
    Realm: data.DeathMagic,
    Race: data.RaceFantastic,
}

var Demon Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 49,
    Name: "Demon",
    Race: data.RaceFantastic,
    CombatLbxFile: "figure12.lbx",
    CombatIndex: 32,
    Count: 1,
    Realm: data.DeathMagic,
    MovementSpeed: 2,
    Flying: true,
    MeleeAttackPower: 14,
    Defense: 6,
    Resistance: 7,
    HitPoints: 12,
    // tohit +10%
    Abilities: []Ability{AbilityToHit, AbilityPoisonImmunity, AbilityIllusionsImmunity, AbilityColdImmunity, AbilityDeathImmunity, AbilityWeaponImmunity, AbilityMissileImmunity},
}

var Wraith Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 50,
    CombatLbxFile: "figure12.lbx",
    CombatIndex: 40,
    Name: "Wraiths",
    UpkeepMana: 5,
    Count: 4,
    Flying: true,
    MovementSpeed: 2,
    MeleeAttackPower: 7,
    Defense: 6,
    Resistance: 8,
    HitPoints: 8,
    // tohit +20%
    // life steal -3
    Abilities: []Ability{AbilityToHit, AbilityPoisonImmunity, AbilityIllusionsImmunity, AbilityColdImmunity, AbilityDeathImmunity, AbilityWeaponImmunity, AbilityNonCorporeal, AbilityLifeSteal},
    Realm: data.DeathMagic,
    Race: data.RaceFantastic,
}

var ShadowDemon Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 51,
    CombatLbxFile: "figure12.lbx",
    CombatIndex: 48,
    Name: "Shadow Demons",
    UpkeepMana: 7,
    Count: 4,
    Flying: true,
    MovementSpeed: 1,
    MeleeAttackPower: 5,
    RangedAttackPower: 4,
    RangedAttacks: 8,
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 4,
    Resistance: 8,
    HitPoints: 5,
    // tohit +20%
    Abilities: []Ability{
        AbilityToHit, AbilityPlaneShift, AbilityNonCorporeal,
        AbilityRegeneration,
        AbilityPoisonImmunity, AbilityWeaponImmunity,
        AbilityIllusionsImmunity, AbilityColdImmunity, AbilityDeathImmunity,
    },
    Realm: data.DeathMagic,
    Race: data.RaceFantastic,
}

var DeathKnight Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 52,
    CombatLbxFile: "figure12.lbx",
    CombatIndex: 56,
    Name: "Death Knights",
    UpkeepMana: 8,
    Count: 4,
    Flying: true,
    MovementSpeed: 3,
    MeleeAttackPower: 9,
    Defense: 8,
    Resistance: 10,
    HitPoints: 8,
    // tohit +30%
    // life steal -4
    Abilities: []Ability{
        AbilityToHit, AbilityArmorPiercing, AbilityFirstStrike,
        AbilityLifeSteal,
        AbilityPoisonImmunity, AbilityWeaponImmunity,
        AbilityIllusionsImmunity, AbilityColdImmunity, AbilityDeathImmunity,
    },
    Realm: data.DeathMagic,
    Race: data.RaceFantastic,
}

var DemonLord Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 53,
    CombatLbxFile: "figure12.lbx",
    CombatIndex: 64,
    Name: "Demon Lord",
    UpkeepMana: 15,
    Count: 1,
    Flying: true,
    MovementSpeed: 2,
    MeleeAttackPower: 20,
    RangedAttackPower: 10,
    RangedAttacks: 8,
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 10,
    Resistance: 12,
    HitPoints: 20,
    // tohit +30%
    // summon demons 3
    // life steal -5
    Abilities: []Ability{
        AbilityToHit, AbilitySummonDemons,
        AbilityPoisonImmunity, AbilityWeaponImmunity,
        AbilityIllusionsImmunity, AbilityColdImmunity, AbilityDeathImmunity,
        AbilityCauseFear, AbilityLifeSteal,
    },
    Realm: data.DeathMagic,
    Race: data.RaceFantastic,
}

var Zombie Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 54,
    CombatLbxFile: "figure12.lbx",
    CombatIndex: 72,
    Name: "Zombies",
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 4,
    Defense: 3,
    Resistance: 3,
    HitPoints: 3,
    // tohit +10%
    Abilities: []Ability{
        AbilityToHit, AbilityPoisonImmunity, AbilityIllusionsImmunity, AbilityColdImmunity, AbilityDeathImmunity,
    },
    Realm: data.DeathMagic,
    Race: data.RaceFantastic,
}

var Unicorn Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 55,
    CombatLbxFile: "figure12.lbx",
    CombatIndex: 80,
    Name: "Unicorns",
    UpkeepMana: 5,
    Count: 4,
    MovementSpeed: 2,
    MeleeAttackPower: 5,
    Defense: 3,
    Resistance: 7,
    HitPoints: 6,
    // tohit +20%
    // resistance to all +2
    Abilities: []Ability{AbilityToHit, AbilityTeleporting, AbilityPoisonImmunity, AbilityResistanceToAll},
    Race: data.RaceFantastic,
    Realm: data.LifeMagic,
}

var GuardianSpirit Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 56,
    CombatLbxFile: "figure12.lbx",
    CombatIndex: 88,
    Name: "Guardian Spirit",
    UpkeepMana: 1,
    Count: 1,
    MovementSpeed: 1,
    Swimming: true,
    MeleeAttackPower: 10,
    Defense: 4,
    Resistance: 10,
    HitPoints: 10,
    Race: data.RaceFantastic,
    // resistance to all +1
    Abilities: []Ability{AbilityMeld, AbilityNonCorporeal, AbilityResistanceToAll},
    Realm: data.LifeMagic,
}

var Angel Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 57,
    CombatLbxFile: "figure12.lbx",
    CombatIndex: 96,
    Name: "Angel",
    UpkeepMana: 15,
    Count: 1,
    Flying: true,
    MovementSpeed: 3,
    MeleeAttackPower: 13,
    Defense: 7,
    Resistance: 8,
    HitPoints: 15,
    // tohit +20%
    // holy bonus 1
    Abilities: []Ability{AbilityToHit, AbilityIllusionsImmunity, AbilityHolyBonus, AbilityDispelEvil},
    Race: data.RaceFantastic,
    Realm: data.LifeMagic,
}

var ArchAngel Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 58,
    CombatLbxFile: "figure12.lbx",
    CombatIndex: 104,
    Name: "Arch Angel",
    UpkeepMana: 20,
    Count: 1,
    Flying: true,
    MovementSpeed: 4,
    MeleeAttackPower: 15,
    Defense: 10,
    Resistance: 12,
    HitPoints: 18,
    // tohit +30%
    // caster 40
    // holy bonus 2
    Abilities: []Ability{
        AbilityToHit, AbilityCaster, AbilityIllusionsImmunity, AbilityHolyBonus,
    },
    Race: data.RaceFantastic,
    Realm: data.LifeMagic,
}

var WarBear Unit = Unit{
    LbxFile: "units2.lbx",
    Name: "War Bears",
    Index: 59,
    CombatLbxFile: "figure15.lbx",
    CombatIndex: 0,
    Count: 2,
    Realm: data.NatureMagic,
    MeleeAttackPower: 7,
    Defense: 3,
    Resistance: 6,
    HitPoints: 8,
    MovementSpeed: 2,
    Race: data.RaceFantastic,
}

var Sprite Unit = Unit{
    LbxFile: "units2.lbx",
    Name: "Sprites",
    Index: 60,
    CombatLbxFile: "figure15.lbx",
    CombatIndex: 8,
    Count: 4,
    Realm: data.NatureMagic,
    MeleeAttackPower: 4,
    RangedAttacks: 4,
    RangedAttackPower: 3,
    RangedAttackDamageType: DamageRangedMagical,
    RangeAttackIndex: 88,
    MovementSpeed: 2,
    Defense: 2,
    Resistance: 8,
    HitPoints: 1,
    Flying: true,
    Race: data.RaceFantastic,
}

var Cockatrice Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 61,
    Name: "Cockatrice",
    CombatLbxFile: "figure13.lbx",
    CombatIndex: 8,
    Realm: data.NatureMagic,
    Count: 4,
    MovementSpeed: 2,
    Flying: true,
    MeleeAttackPower: 4,
    Defense: 3,
    Resistance: 7,
    HitPoints: 3,
    Race: data.RaceFantastic,
}

var Basilisk Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 62,
    Name: "Basilisk",
    CombatLbxFile: "figure13.lbx",
    CombatIndex: 16,
    Count: 1,
    Realm: data.NatureMagic,
    MovementSpeed: 2,
    MeleeAttackPower: 15,
    Defense: 4,
    Resistance: 7,
    HitPoints: 30,
    Race: data.RaceFantastic,
}

var GiantSpider Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 63,
    Name: "Giant Spider",
    CombatLbxFile: "figure13.lbx",
    CombatIndex: 24,
    Realm: data.NatureMagic,
    Count: 2,
    MovementSpeed: 2,
    MeleeAttackPower: 4,
    Defense: 3,
    Resistance: 7,
    HitPoints: 10,
    Race: data.RaceFantastic,
}

var StoneGiant Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 64,
    Name: "Stone Giant",
    CombatLbxFile: "figure13.lbx",
    CombatIndex: 32,
    Count: 1,
    Realm: data.NatureMagic,
    MovementSpeed: 2,
    MeleeAttackPower: 15,
    RangedAttackPower: 15,
    RangedAttackDamageType: DamageRangedBoulder,
    RangedAttacks: 2, // FIXME
    Defense: 8,
    Resistance: 9,
    HitPoints: 20,
    Race: data.RaceFantastic,
}

var Colossus Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 65,
    Name: "Colossus",
    CombatLbxFile: "figure13.lbx",
    CombatIndex: 40,
    Count: 1,
    Realm: data.NatureMagic,
    MovementSpeed: 2,
    MeleeAttackPower: 20,
    Defense: 10,
    Resistance: 15,
    HitPoints: 30,
    Race: data.RaceFantastic,
}

var Gorgon Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 66,
    Name: "Gorgons",
    CombatLbxFile: "figure13.lbx",
    CombatIndex: 48,
    Count: 2,
    MovementSpeed: 2,
    Realm: data.NatureMagic,
    Flying: true,
    MeleeAttackPower: 8,
    Defense: 7,
    Resistance: 9,
    HitPoints: 9,
    Race: data.RaceFantastic,
}

var EarthElemental Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 67,
    Race: data.RaceFantastic,
    CombatLbxFile: "figure15.lbx",
    CombatIndex: 64,
    Realm: data.NatureMagic,
    Name: "Earth Elemental",
    Count: 1,
    MeleeAttackPower: 25,
    Defense: 4,
    MovementSpeed: 1,
    Resistance: 8,
    HitPoints: 30,
}

var Behemoth Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 68,
    Name: "Behemoth",
    CombatLbxFile: "figure13.lbx",
    CombatIndex: 64,
    Count: 1,
    Realm: data.NatureMagic,
    MovementSpeed: 2,
    MeleeAttackPower: 25,
    Defense: 9,
    Resistance: 10,
    HitPoints: 45,
    Race: data.RaceFantastic,
}

var GreatWyrm Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 69,
    Name: "Great Wyrm",
    CombatLbxFile: "figure13.lbx",
    CombatIndex: 72,
    Count: 1,
    Realm: data.NatureMagic,
    MovementSpeed: 3,
    MeleeAttackPower: 25,
    Defense: 12,
    Resistance: 12,
    HitPoints: 45,
    Race: data.RaceFantastic,
}

var FloatingIsland Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 70,
    Race: data.RaceFantastic,
    Realm: data.SorceryMagic,
}

var PhantomBeast Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 71,
    Race: data.RaceFantastic,
    CombatLbxFile: "figure13.lbx",
    CombatIndex: 88,
    Realm: data.SorceryMagic,
    Name: "Phantom Beast",
    Count: 1,
    MeleeAttackPower: 18,
    MovementSpeed: 2,
    Resistance: 8,
    HitPoints: 20,
}

var PhantomWarrior Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 72,
    CombatLbxFile: "figure13.lbx",
    CombatIndex: 96,
    Realm: data.SorceryMagic,
    Name: "Phantom Warriors",
    MovementSpeed: 1,
    Count: 6,
    MeleeAttackPower: 3,
    HitPoints: 1,
    Resistance: 6,
    Race: data.RaceFantastic,
}

var StormGiant Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 73,
    Name: "Storm Giant",
    CombatLbxFile: "figure13.lbx",
    CombatIndex: 104,
    Count: 1,
    Realm: data.SorceryMagic,
    MovementSpeed: 2,
    MeleeAttackPower: 12,
    RangedAttackPower: 10,
    RangedAttacks: 4,
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 7,
    Resistance: 9,
    HitPoints: 20,
    Race: data.RaceFantastic,
}

var AirElemental Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 74,
    Name: "Air Elemental",
    Race: data.RaceFantastic,
    CombatLbxFile: "figure13.lbx",
    CombatIndex: 112,
    Count: 1,
    Realm: data.SorceryMagic,
    Flying: true,
    MovementSpeed: 5,
    MeleeAttackPower: 15,
    Defense: 8,
    Resistance: 9,
    HitPoints: 10,
}

var Djinn Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 75,
    Name: "Djinn",
    CombatLbxFile: "figure14.lbx",
    CombatIndex: 0,
    Count: 1,
    Realm: data.SorceryMagic,
    MovementSpeed: 3,
    Flying: true,
    MeleeAttackPower: 15,
    RangedAttackPower: 8,
    // caster 20
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 8,
    Resistance: 10,
    HitPoints: 20,
    Race: data.RaceFantastic,
}

var SkyDrake Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 76,
    Name: "SkyDrake",
    Realm: data.SorceryMagic,
    CombatLbxFile: "figure14.lbx",
    CombatIndex: 8,
    Count: 1,
    MovementSpeed: 4,
    Flying: true,
    MeleeAttackPower: 20,
    Defense: 10,
    Resistance: 14,
    HitPoints: 25,
    Race: data.RaceFantastic,
}

var Nagas Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 77,
    Name: "Nagas",
    CombatLbxFile: "figure14.lbx",
    CombatIndex: 16,
    Realm: data.SorceryMagic,
    MovementSpeed: 1,
    Count: 2,
    MeleeAttackPower: 4,
    Defense: 3,
    Resistance: 7,
    HitPoints: 6,
    Race: data.RaceFantastic,
}

var HeroBrax Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 0,
    Name: "Brax",
    Race: data.RaceHero,
}

var HeroGunther Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 1,
    Name: "Gunther",
    Race: data.RaceHero,
}

var HeroZaldron Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 2,
    Name: "Zaldron",
    Race: data.RaceHero,
}

var HeroBShan Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 3,
    Name: "B'Shan",
    Race: data.RaceHero,
}

var HeroRakir Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 4,
    CombatLbxFile: "figures1.lbx",
    CombatIndex: 32,
    Name: "Rakir",
    Count: 1,
    HitPoints: 7,
    MeleeAttackPower: 5,
    Defense: 4,
    UpkeepGold: 2,
    Resistance: 6,
    MovementSpeed: 2,
    Race: data.RaceHero,
}

var HeroValana Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 5,
    Name: "Valana",
    Race: data.RaceHero,
}

var HeroBahgtru Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 6,
    Name: "Bahgtru",
    Race: data.RaceHero,
}

var HeroSerena Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 7,
    Name: "Serena",
    Race: data.RaceHero,
}

var HeroShuri Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 8,
    Name: "Shuri",
    Race: data.RaceHero,
}

var HeroTheria Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 9,
    Name: "Theria",
    Race: data.RaceHero,
}

var HeroGreyfairer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 10,
    Name: "Greyfairer",
    Race: data.RaceHero,
}

var HeroTaki Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 11,
    Name: "Taki",
    Race: data.RaceHero,
}

var HeroReywind Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 12,
    Name: "Reywind",
    Race: data.RaceHero,
}

var HeroMalleus Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 13,
    Name: "Malleus",
    Race: data.RaceHero,
}

var HeroTumu Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 14,
    Name: "Tumu",
    Race: data.RaceHero,
}

var HeroJaer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 15,
    Name: "Jaer",
    Race: data.RaceHero,
}

var HeroMarcus Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 16,
    Name: "Marcus",
    Race: data.RaceHero,
}

var HeroFang Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 17,
    Name: "Fang",
    Race: data.RaceHero,
}

var HeroMorgana Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 18,
    Name: "Morgana",
    Race: data.RaceHero,
}

var HeroAureus Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 19,
    Name: "Aureus",
    Race: data.RaceHero,
}

var HeroShinBo Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 20,
    Name: "Shin Bo",
    Race: data.RaceHero,
}

var HeroSpyder Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 21,
    Name: "Spyder",
    Race: data.RaceHero,
}

var HeroShalla Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 22,
    Name: "Shalla",
    Race: data.RaceHero,
}

var HeroYramrag Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 23,
    Name: "Yramrag",
    Race: data.RaceHero,
}

var HeroMysticX Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 24,
    Name: "Mystic X",
    Race: data.RaceHero,
}

var HeroAerie Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 25,
    Name: "Aerie",
    Race: data.RaceHero,
}

var HeroDethStryke Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 26,
    Name: "Deth Stryke",
    Race: data.RaceHero,
}

var HeroElana Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 27,
    Name: "Elana",
    Race: data.RaceHero,
}

var HeroRoland Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 28,
    Name: "Roland",
    Race: data.RaceHero,
}

var HeroMortu Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 29,
    Name: "Mortu",
    Race: data.RaceHero,
}

var HeroAlorra Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 30,
    Name: "Alorra",
    Race: data.RaceHero,
}

var HeroSirHarold Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 31,
    Name: "Sir Harold",
    Race: data.RaceHero,
}

var HeroRavashack Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 32,
    Name: "Ravashack",
    Race: data.RaceHero,
}

var HeroWarrax Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 33,
    CombatLbxFile: "figures3.lbx",
    CombatIndex: 24,
    Name: "Warrax",
    MovementSpeed: 2,
    MeleeAttackPower: 8,
    RangedAttackPower: 8,
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 5,
    Resistance: 9,
    HitPoints: 8,
    Count: 1,
    Race: data.RaceHero,
}

var HeroTorin Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 34,
    Name: "Torin",
    Race: data.RaceHero,
}

var Trireme Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 35,
    Name: "Trireme",
    RequiredBuildings: []building.Building{building.BuildingShipwrightsGuild},
    Race: data.RaceNone,
}

var Galley Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 36,
    Name: "Galley",
    RequiredBuildings: []building.Building{building.BuildingShipYard},
    Race: data.RaceNone,
}

var Catapult Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 37,
    Name: "Catapult",
    RequiredBuildings: []building.Building{building.BuildingMechaniciansGuild},
    Race: data.RaceNone,
}

var Warship Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 38,
    Name: "Warship",
    RequiredBuildings: []building.Building{building.BuildingMaritimeGuild},
    Race: data.RaceNone,
}

var BarbarianSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Name: "Spearmen",
    Index: 39,
    CombatLbxFile: "figures3.lbx",
    CombatIndex: 72,
    ProductionCost: 15,
    UpkeepFood: 1,
    Count: 8,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    Defense: 2,
    Resistance: 5,
    HitPoints: 1,
    // thrown 1
    Abilities: []Ability{AbilityThrown},
    Race: data.RaceBarbarian,
}

var BarbarianSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 40,
    Name: "Swordsmen",
    CombatLbxFile: "figures3.lbx",
    CombatIndex: 80,
    ProductionCost: 30,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 3,
    Defense: 2,
    Resistance: 5,
    HitPoints: 1,
    // thrown 1
    Abilities: []Ability{AbilityLargeShield, AbilityThrown},
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceBarbarian,
}

var BarbarianBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 41,
    Name: "Bowmen",
    CombatLbxFile: "figures3.lbx",
    CombatIndex: 88,
    ProductionCost: 30,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    RangedAttackPower: 1,
    RangedAttacks: 8,
    RangedAttackDamageType: DamageRangedPhysical,
    Defense: 1,
    Resistance: 5,
    HitPoints: 1,
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSawmill},
    Race: data.RaceBarbarian,
}

var BarbarianCavalry Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 42,
    Name: "Cavalry",
    CombatLbxFile: "figures3.lbx",
    CombatIndex: 96,
    ProductionCost: 60,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 2,
    MeleeAttackPower: 4,
    Defense: 2,
    Resistance: 5,
    HitPoints: 3,
    // thrown 1
    Abilities: []Ability{AbilityFirstStrike, AbilityThrown},
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingStables},
    Race: data.RaceBarbarian,
}

var BarbarianShaman Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 43,
    CombatLbxFile: "figures3.lbx",
    CombatIndex: 104,
    Name: "Shaman",
    ProductionCost: 50,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 1,
    MeleeAttackPower: 2,
    // nature
    RangedAttackPower: 2,
    RangedAttacks: 4,
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 3,
    Resistance: 7,
    HitPoints: 1,
    Abilities: []Ability{AbilityHealer, AbilityPurify},
    RequiredBuildings: []building.Building{building.BuildingShrine},
    Race: data.RaceBarbarian,
}

var BarbarianSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 44,
    CombatLbxFile: "figures3.lbx",
    CombatIndex: 112,
    Name: "Settlers",
    ProductionCost: 60,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 1,
    MovementSpeed: 1,
    Defense: 1,
    Resistance: 5,
    HitPoints: 10,
    Abilities: []Ability{AbilityCreateOutpost},
    Race: data.RaceBarbarian,
}

var Berserkers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 45,
    CombatLbxFile: "figures4.lbx",
    CombatIndex: 0,
    Name: "Berserkers",
    ProductionCost: 120,
    UpkeepGold: 3,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 7,
    Defense: 3,
    Resistance: 7,
    HitPoints: 3,
    // thrown 3
    Abilities: []Ability{AbilityThrown},
    RequiredBuildings: []building.Building{building.BuildingArmorersGuild},
    Race: data.RaceBarbarian,
}

var BeastmenSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 46,
    CombatLbxFile: "figures4.lbx",
    CombatIndex: 8,
    ProductionCost: 20,
    UpkeepFood: 1,
    Count: 8,
    MovementSpeed: 1,
    MeleeAttackPower: 2,
    Defense: 2,
    Resistance: 5,
    HitPoints: 2,
    Name: "Spearmen",
    Race: data.RaceBeastmen,
}

var BeastmenSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 47,
    CombatLbxFile: "figures4.lbx",
    CombatIndex: 16,
    Name: "Swordsmen",
    ProductionCost: 40,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 4,
    Defense: 2,
    Resistance: 5,
    HitPoints: 2,
    Abilities: []Ability{AbilityLargeShield},
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceBeastmen,
}

var BeastmenHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 48,
    CombatLbxFile: "figures4.lbx",
    CombatIndex: 24,
    ProductionCost: 80,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 5,
    Defense: 3,
    Resistance: 5,
    HitPoints: 2,
    Name: "Halberdiers",
    RequiredBuildings: []building.Building{building.BuildingArmory},
    Race: data.RaceBeastmen,
}

var BeastmenBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 49,
    CombatLbxFile: "figures4.lbx",
    CombatIndex: 32,
    Name: "Bowmen",
    ProductionCost: 60,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 2,
    RangedAttackPower: 1,
    RangedAttacks: 8,
    RangedAttackDamageType: DamageRangedPhysical,
    Defense: 1,
    Resistance: 5,
    HitPoints: 2,
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSawmill},
    Race: data.RaceBeastmen,
}

var BeastmenPriest Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 50,
    CombatLbxFile: "figures4.lbx",
    CombatIndex: 40,
    Name: "Priests",
    ProductionCost: 150,
    UpkeepGold: 3,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 1,
    MeleeAttackPower: 4,
    // nature
    RangedAttackPower: 4,
    RangedAttacks: 4,
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 4,
    Resistance: 8,
    HitPoints: 2,
    // healing spell 1x
    Abilities: []Ability{AbilityHealer, AbilityPurify, AbilityHealingSpell},
    RequiredBuildings: []building.Building{building.BuildingParthenon},
    Race: data.RaceBeastmen,
}

var BeastmenMagician Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 51,
    CombatLbxFile: "figures4.lbx",
    CombatIndex: 48,
    Name: "Magicians",
    ProductionCost: 180,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 1,
    MeleeAttackPower: 2,
    // chaos
    RangedAttackPower: 5,
    RangedAttacks: 4,
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 3,
    Resistance: 9,
    HitPoints: 2,
    // fireball spell 1x
    Abilities: []Ability{AbilityMissileImmunity, AbilityFireballSpell},
    RequiredBuildings: []building.Building{building.BuildingWizardsGuild},
    Race: data.RaceBeastmen,
}

var BeastmenEngineer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 52,
    CombatLbxFile: "figures4.lbx",
    CombatIndex: 56,
    Name: "Engineers",
    ProductionCost: 60,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 2,
    Defense: 1,
    Resistance: 5,
    HitPoints: 2,
    Abilities: []Ability{AbilityConstruction, AbilityWallCrusher},
    RequiredBuildings: []building.Building{building.BuildingBuildersHall},
    Race: data.RaceBeastmen,
}

var BeastmenSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 53,
    CombatLbxFile: "figures4.lbx",
    CombatIndex: 64,
    ProductionCost: 120,
    UpkeepGold: 3,
    UpkeepFood: 1,
    Count: 1,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    Defense: 1,
    Resistance: 5,
    HitPoints: 20,
    Abilities: []Ability{AbilityCreateOutpost},
    Name: "Settlers",
    Race: data.RaceBeastmen,
}

var Centaur Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 54,
    CombatLbxFile: "figures4.lbx",
    CombatIndex: 72,
    Name: "Centaurs",
    ProductionCost: 100,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 2,
    MeleeAttackPower: 3,
    RangedAttackPower: 2,
    RangedAttacks: 6,
    RangedAttackDamageType: DamageRangedPhysical,
    Defense: 3,
    Resistance: 5,
    HitPoints: 3,
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingStables},
    Race: data.RaceBeastmen,
}

var Manticore Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 55,
    CombatLbxFile: "figures4.lbx",
    CombatIndex: 80,
    Name: "Manticores",
    ProductionCost: 160,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 2,
    Flying: true,
    MovementSpeed: 2,
    MeleeAttackPower: 5,
    Defense: 3,
    Resistance: 6,
    HitPoints: 7,
    // poison touch 6
    Abilities: []Ability{AbilityScouting, AbilityPoisonTouch},
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingAnimistsGuild},
    Race: data.RaceBeastmen,
}

var Minotaur Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 56,
    CombatLbxFile: "figures4.lbx",
    CombatIndex: 88,
    Name: "Minotaurs",
    ProductionCost: 200,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 2,
    MovementSpeed: 1,
    MeleeAttackPower: 12,
    Defense: 4,
    Resistance: 7,
    HitPoints: 12,
    // tohit +20%
    Abilities: []Ability{AbilityToHit, AbilityLargeShield},
    RequiredBuildings: []building.Building{building.BuildingArmorersGuild},
    Race: data.RaceBeastmen,
}

var DarkElfSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 57,
    CombatLbxFile: "figures4.lbx",
    CombatIndex: 96,
    Name: "Spearmen",
    ProductionCost: 25,
    UpkeepFood: 1,
    Count: 8,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    // chaos
    RangedAttackPower: 1,
    RangedAttacks: 4,
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 2,
    Resistance: 7,
    HitPoints: 1,
    Race: data.RaceDarkElf,
}

var DarkElfSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 58,
    CombatLbxFile: "figures4.lbx",
    CombatIndex: 104,
    Name: "Swordsmen",
    ProductionCost: 50,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 3,
    // chaos
    RangedAttackPower: 1,
    RangedAttacks: 4,
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 2,
    Resistance: 7,
    HitPoints: 1,
    Abilities: []Ability{AbilityLargeShield},
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceDarkElf,
}

var DarkElfHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 59,
    CombatLbxFile: "figures4.lbx",
    CombatIndex: 112,
    Name: "Halberdiers",
    ProductionCost: 100,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 4,
    // chaos
    RangedAttackPower: 1,
    RangedAttacks: 4,
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 3,
    Resistance: 7,
    HitPoints: 1,
    RequiredBuildings: []building.Building{building.BuildingArmory},
    Race: data.RaceDarkElf,
}

var DarkElfCavalry Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 60,
    CombatLbxFile: "figures5.lbx",
    CombatIndex: 0,
    Name: "Cavalry",
    ProductionCost: 100,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 2,
    MeleeAttackPower: 4,
    // chaos
    RangedAttackPower: 1,
    RangedAttacks: 4,
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 2,
    Resistance: 7,
    HitPoints: 3,
    Abilities: []Ability{AbilityFirstStrike},
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingStables},
    Race: data.RaceDarkElf,
}

var DarkElfPriests Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 61,
    CombatLbxFile: "figures5.lbx",
    CombatIndex: 8,
    Name: "Priests",
    ProductionCost: 200,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 1,
    MeleeAttackPower: 3,
    // nature
    RangedAttackPower: 6,
    RangedAttacks: 4,
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 4,
    Resistance: 10,
    HitPoints: 1,
    // healing spell 1x
    Abilities: []Ability{AbilityHealer, AbilityPurify, AbilityHealingSpell},
    RequiredBuildings: []building.Building{building.BuildingParthenon},
    Race: data.RaceDarkElf,
}

var DarkElfSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 62,
    CombatLbxFile: "figures5.lbx",
    CombatIndex: 16,
    Name: "Settlers",
    ProductionCost: 150,
    UpkeepGold: 3,
    UpkeepFood: 1,
    Count: 1,
    MovementSpeed: 1,
    Defense: 1,
    Resistance: 7,
    HitPoints: 10,
    Abilities: []Ability{AbilityCreateOutpost},
    Race: data.RaceDarkElf,
}

var Nightblades Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 63,
    CombatLbxFile: "figures5.lbx",
    CombatIndex: 24,
    Name: "Nightblades",
    ProductionCost: 120,
    UpkeepGold: 3,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 4,
    Defense: 3,
    Resistance: 7,
    HitPoints: 1,
    // poison touch 1
    Abilities: []Ability{AbilityInvisibility, AbilityPoisonTouch},
    RequiredBuildings: []building.Building{building.BuildingFightersGuild},
    Race: data.RaceDarkElf,
}

var Warlocks Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 64,
    CombatLbxFile: "figures5.lbx",
    CombatIndex: 32,
    Name: "Warlocks",
    ProductionCost: 240,
    UpkeepGold: 5,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    RangedAttackDamageType: DamageRangedMagical,
    RangeAttackIndex: 16,
    RangedAttackPower: 7,
    RangedAttacks: 4,
    RangeAttackSound: RangeAttackSoundFireball,
    Defense: 4,
    Resistance: 9,
    HitPoints: 1,
    MovementSound: MovementSoundShuffle,
    // doom bolt 1x
    Abilities: []Ability{AbilityDoomBoltSpell, AbilityMissileImmunity},
    RequiredBuildings: []building.Building{building.BuildingWizardsGuild},
    Race: data.RaceDarkElf,
}

var Nightmares Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 65,
    CombatLbxFile: "figures5.lbx",
    CombatIndex: 40,
    Name: "Nightmares",
    ProductionCost: 160,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 2,
    MovementSpeed: 3,
    Flying: true,
    MeleeAttackPower: 8,
    RangedAttackPower: 5,
    RangedAttacks: 4,
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 4,
    Resistance: 8,
    HitPoints: 10,
    Abilities: []Ability{AbilityScouting},
    RequiredBuildings: []building.Building{building.BuildingFantasticStable},
    Race: data.RaceDarkElf,
}

var DraconianSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 66,
    CombatLbxFile: "figures5.lbx",
    CombatIndex: 48,
    Name: "Spearmen",
    ProductionCost: 25,
    UpkeepFood: 1,
    Count: 8,
    Flying: true,
    MovementSpeed: 2,
    MeleeAttackPower: 1,
    Defense: 3,
    Resistance: 6,
    HitPoints: 1,
    // fire breath 1x
    Abilities: []Ability{AbilityFireBreath},
    Race: data.RaceDraconian,
}

var DraconianSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 67,
    CombatLbxFile: "figures5.lbx",
    CombatIndex: 56,
    Name: "Swordsmen",
    ProductionCost: 50,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    Flying: true,
    MovementSpeed: 2,
    MeleeAttackPower: 3,
    Defense: 3,
    Resistance: 6,
    HitPoints: 1,
    // fire breath 1x
    Abilities: []Ability{AbilityLargeShield, AbilityFireBreath},
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceDraconian,
}

var DraconianHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 68,
    CombatLbxFile: "figures5.lbx",
    CombatIndex: 64,
    Name: "Halberdiers",
    ProductionCost: 100,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 6,
    Flying: true,
    MovementSpeed: 2,
    MeleeAttackPower: 4,
    Defense: 4,
    Resistance: 6,
    HitPoints: 1,
    // fire breath 1x
    Abilities: []Ability{AbilityFireBreath},
    RequiredBuildings: []building.Building{building.BuildingArmory},
    Race: data.RaceDraconian,
}

var DraconianBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 69,
    CombatLbxFile: "figures5.lbx",
    CombatIndex: 72,
    Name: "Bowmen",
    ProductionCost: 45,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    Flying: true,
    MovementSpeed: 2,
    MeleeAttackPower: 1,
    RangedAttackPower: 1,
    RangedAttacks: 1,
    RangedAttackDamageType: DamageRangedPhysical,
    Defense: 2,
    Resistance: 6,
    HitPoints: 1,
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSawmill},
    Race: data.RaceDraconian,
}

var DraconianShaman Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 70,
    CombatLbxFile: "figures5.lbx",
    CombatIndex: 80,
    Name: "Shaman",
    ProductionCost: 75,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 4,
    Flying: true,
    MovementSpeed: 2,
    MeleeAttackPower: 2,
    // nature
    RangedAttackPower: 2,
    RangedAttackDamageType: DamageRangedMagical,
    RangedAttacks: 4,
    Defense: 4,
    Resistance: 8,
    HitPoints: 1,
    Abilities: []Ability{AbilityHealer, AbilityPurify},
    RequiredBuildings: []building.Building{building.BuildingShrine},
    Race: data.RaceDraconian,
}

var DraconianMagician Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 71,
    CombatLbxFile: "figures5.lbx",
    CombatIndex: 88,
    Name: "Magicians",
    ProductionCost: 180,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 4,
    Flying: true,
    MovementSpeed: 2,
    MeleeAttackPower: 1,
    RangedAttackPower: 5,
    RangedAttacks: 4,
    // chaos
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 4,
    Resistance: 10,
    HitPoints: 1,
    // fireball 1x
    Abilities: []Ability{AbilityMissileImmunity, AbilityFireballSpell},
    RequiredBuildings: []building.Building{building.BuildingWizardsGuild},
    Race: data.RaceDraconian,
}

// removed from game
/*
var DraconianEngineer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 72,
    Name: "Engineers",
    RequiredBuildings: []building.Building{building.BuildingBuildersHall},
    Race: data.RaceDraconian,
}
*/

var DraconianSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 73,
    CombatLbxFile: "figures5.lbx",
    CombatIndex: 104,
    ProductionCost: 150,
    UpkeepGold: 3,
    UpkeepFood: 1,
    Count: 1,
    Flying: true,
    MovementSpeed: 1,
    Defense: 2,
    Resistance: 6,
    HitPoints: 10,
    Abilities: []Ability{AbilityCreateOutpost},
    Name: "Settlers",
    Race: data.RaceDraconian,
}

var DoomDrake Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 74,
    CombatLbxFile: "figures5.lbx",
    CombatIndex: 112,
    Name: "Doom Drake",
    ProductionCost: 160,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 2,
    Flying: true,
    MovementSpeed: 3,
    MeleeAttackPower: 8,
    Defense: 3,
    Resistance: 9,
    HitPoints: 10,
    // fire breath 6
    Abilities: []Ability{AbilityScouting, AbilityFireBreath},
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingStables},
    Race: data.RaceDraconian,
}

var AirShip Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 75,
    CombatLbxFile: "figures6.lbx",
    CombatIndex: 0,
    Name: "Air Ship",
    ProductionCost: 200,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 1,
    Flying: true,
    MovementSpeed: 4,
    MeleeAttackPower: 5,
    RangedAttackPower: 10,
    RangedAttacks: 10,
    RangedAttackDamageType: DamageRangedBoulder,
    Defense: 5,
    Resistance: 8,
    HitPoints: 20,
    Abilities: []Ability{AbilityScouting, AbilityWallCrusher},
    RequiredBuildings: []building.Building{building.BuildingShipYard},
    Race: data.RaceDraconian,
}

var DwarfSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 76,
    CombatLbxFile: "figures6.lbx",
    CombatIndex: 8,
    Name: "Swordsmen",
    ProductionCost: 50,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 3,
    Defense: 2,
    Resistance: 8,
    HitPoints: 3,
    Abilities: []Ability{AbilityLargeShield, AbilityMountaineer},
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceDwarf,
}

var DwarfHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 77,
    CombatLbxFile: "figures6.lbx",
    CombatIndex: 16,
    Name: "Halberdiers",
    ProductionCost: 100,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 4,
    Defense: 3,
    Resistance: 8,
    HitPoints: 3,
    Abilities: []Ability{AbilityMountaineer},
    RequiredBuildings: []building.Building{building.BuildingArmory},
    Race: data.RaceDwarf,
}

var DwarfEngineer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 78,
    CombatLbxFile: "figures6.lbx",
    CombatIndex: 24,
    Name: "Engineers",
    ProductionCost: 40,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    Defense: 1,
    Resistance: 8,
    HitPoints: 3,
    Abilities: []Ability{AbilityConstruction, AbilityWallCrusher, AbilityMountaineer},
    RequiredBuildings: []building.Building{building.BuildingBuildersHall},
    Race: data.RaceDwarf,
}

var Hammerhands Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 79,
    CombatLbxFile: "figures6.lbx",
    CombatIndex: 32,
    Name: "Hammerhands",
    ProductionCost: 160,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 8,
    Defense: 4,
    Resistance: 9,
    HitPoints: 4,
    Abilities: []Ability{AbilityMountaineer},
    RequiredBuildings: []building.Building{building.BuildingFightersGuild},
    Race: data.RaceDwarf,
}

var SteamCannon Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 80,
    CombatLbxFile: "figures6.lbx",
    CombatIndex: 40,
    Name: "Steam Cannon",
    ProductionCost: 180,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 1,
    MovementSpeed: 1,
    RangedAttackPower: 12,
    RangedAttacks: 10,
    RangedAttackDamageType: DamageRangedBoulder,
    Defense: 2,
    Resistance: 9,
    HitPoints: 12,
    RequiredBuildings: []building.Building{building.BuildingMinersGuild},
    Race: data.RaceDwarf,
}

var Golem Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 81,
    CombatLbxFile: "figures6.lbx",
    CombatIndex: 48,
    Name: "Golem",
    ProductionCost: 200,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 1,
    MovementSpeed: 1,
    MeleeAttackPower: 12,
    Defense: 8,
    Resistance: 15,
    HitPoints: 20,
    Abilities: []Ability{AbilityPoisonImmunity, AbilityDeathImmunity},
    RequiredBuildings: []building.Building{building.BuildingArmorersGuild},
    Race: data.RaceDwarf,
}

var DwarfSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 82,
    CombatLbxFile: "figures6.lbx",
    CombatIndex: 56,
    Name: "Settlers",
    ProductionCost: 150,
    UpkeepGold: 3,
    UpkeepFood: 1,
    Count: 1,
    MovementSpeed: 1,
    Defense: 1,
    Resistance: 8,
    HitPoints: 30,
    Abilities: []Ability{AbilityCreateOutpost},
    Race: data.RaceDwarf,
}

var GnollSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 83,
    CombatLbxFile: "figures6.lbx",
    CombatIndex: 64,
    Name: "Spearmen",
    ProductionCost: 10,
    UpkeepFood: 1,
    Count: 8,
    MovementSpeed: 1,
    MeleeAttackPower: 3,
    Defense: 2,
    Resistance: 4,
    HitPoints: 1,
    Race: data.RaceGnoll,
}

var GnollSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 84,
    CombatLbxFile: "figures6.lbx",
    CombatIndex: 72,
    Name: "Swordsmen",
    ProductionCost: 20,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 5,
    Defense: 2,
    Resistance: 4,
    HitPoints: 1,
    Abilities: []Ability{AbilityLargeShield},
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceGnoll,
}

var GnollHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 85,
    CombatLbxFile: "figures6.lbx",
    CombatIndex: 80,
    Name: "Halberdiers",
    ProductionCost: 40,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 6,
    Defense: 3,
    Resistance: 4,
    HitPoints: 1,
    Abilities: []Ability{AbilityNegateFirstStrike},
    RequiredBuildings: []building.Building{building.BuildingArmory},
    Race: data.RaceGnoll,
}

var GnollBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 86,
    CombatLbxFile: "figures6.lbx",
    CombatIndex: 88,
    Name: "Bowmen",
    ProductionCost: 30,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 3,
    RangedAttackPower: 1,
    RangedAttacks: 8,
    RangedAttackDamageType: DamageRangedPhysical,
    Defense: 1,
    Resistance: 4,
    HitPoints: 1,
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSawmill},
    Race: data.RaceGnoll,
}

var GnollSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 87,
    CombatLbxFile: "figures6.lbx",
    CombatIndex: 96,
    Name: "Settlers",
    ProductionCost: 60,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 1,
    MovementSpeed: 1,
    MeleeAttackPower: 2,
    Defense: 1,
    Resistance: 4,
    HitPoints: 10,
    Abilities: []Ability{AbilityCreateOutpost},
    Race: data.RaceGnoll,
}

var WolfRiders Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 88,
    CombatLbxFile: "figures6.lbx",
    CombatIndex: 104,
    Name: "Wolf Riders",
    ProductionCost: 100,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 3,
    MeleeAttackPower: 7,
    Defense: 3,
    Resistance: 4,
    HitPoints: 5,
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingStables},
    Race: data.RaceGnoll,
}

var HalflingSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 89,
    CombatLbxFile: "figures6.lbx",
    CombatIndex: 112,
    Name: "Spearmen",
    ProductionCost: 15,
    UpkeepFood: 1,
    Count: 8,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    Defense: 2,
    Resistance: 6,
    HitPoints: 1,
    Abilities: []Ability{AbilityLucky},
    Race: data.RaceHalfling,
}

var HalflingSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 90,
    CombatLbxFile: "figures7.lbx",
    CombatIndex: 0,
    Name: "Swordsmen",
    ProductionCost: 30,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 8,
    MovementSpeed: 1,
    MeleeAttackPower: 2,
    Defense: 2,
    Resistance: 6,
    HitPoints: 1,
    Abilities: []Ability{AbilityLargeShield, AbilityLucky},
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceHalfling,
}

var HalflingBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 91,
    CombatLbxFile: "figures7.lbx",
    CombatIndex: 8,
    Name: "Bowmen",
    ProductionCost: 45,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    RangedAttackPower: 1,
    RangedAttacks: 8,
    RangedAttackDamageType: DamageRangedPhysical,
    Defense: 1,
    Resistance: 6,
    HitPoints: 1,
    Abilities: []Ability{AbilityLucky},
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSawmill},
    Race: data.RaceHalfling,
}

var HalflingShamans Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 92,
    CombatLbxFile: "figures7.lbx",
    CombatIndex: 16,
    Name: "Shamans",
    ProductionCost: 75,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    RangedAttackPower: 2,
    RangedAttacks: 4,
    // nature
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 3,
    Resistance: 8,
    HitPoints: 1,
    Abilities: []Ability{AbilityHealer, AbilityPurify, AbilityLucky},
    RequiredBuildings: []building.Building{building.BuildingShrine},
    Race: data.RaceHalfling,
}

var HalflingSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 93,
    CombatLbxFile: "figures7.lbx",
    CombatIndex: 24,
    ProductionCost: 90,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 1,
    MovementSpeed: 1,
    Defense: 1,
    Resistance: 6,
    HitPoints: 10,
    Abilities: []Ability{AbilityCreateOutpost, AbilityLucky},
    Name: "Settlers",
    Race: data.RaceHalfling,
}

var Slingers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 94,
    CombatLbxFile: "figures7.lbx",
    CombatIndex: 32,
    Name: "Slingers",
    ProductionCost: 100,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 8,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    RangedAttackPower: 2,
    RangedAttacks: 6,
    RangedAttackDamageType: DamageRangedBoulder,
    Defense: 2,
    Resistance: 6,
    HitPoints: 1,
    Abilities: []Ability{AbilityLucky},
    RequiredBuildings: []building.Building{building.BuildingArmory},
    Race: data.RaceHalfling,
}

var HighElfSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 95,
    CombatLbxFile: "figures7.lbx",
    CombatIndex: 40,
    Count: 8,
    Name: "Spearmen",
    ProductionCost: 15,
    MeleeAttackPower: 1,
    Defense: 2,
    AttackSound: AttackSoundNormal,
    MovementSound: MovementSoundMarching,
    Resistance: 6,
    // tohit +10%
    Abilities: []Ability{AbilityForester, AbilityToHit},
    MovementSpeed: 1,
    Race: data.RaceHighElf,
    HitPoints: 1,
    UpkeepFood: 1,
}

var HighElfSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 96,
    CombatLbxFile: "figures7.lbx",
    CombatIndex: 48,
    Name: "Swordsmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    ProductionCost: 30,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 3,
    Defense: 2,
    Resistance: 6,
    HitPoints: 1,
    // tohit +10%
    Abilities: []Ability{AbilityLargeShield, AbilityToHit, AbilityForester},
    Race: data.RaceHighElf,
}

var HighElfHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 97,
    CombatLbxFile: "figures7.lbx",
    CombatIndex: 56,
    Name: "Halberdiers",
    ProductionCost: 60,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 4,
    Defense: 3,
    Resistance: 6,
    HitPoints: 1,
    RequiredBuildings: []building.Building{building.BuildingArmory},
    // tohit +10%
    Abilities: []Ability{AbilityForester, AbilityToHit},
    Race: data.RaceHighElf,
}

var HighElfCavalry Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 98,
    CombatLbxFile: "figures7.lbx",
    CombatIndex: 64,
    Name: "Cavalry",
    ProductionCost: 60,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 2,
    MeleeAttackPower: 4,
    Defense: 2,
    Resistance: 6,
    HitPoints: 3,
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingStables},
    // tohit +10%
    Abilities: []Ability{AbilityToHit, AbilityForester, AbilityFirstStrike},
    Race: data.RaceHighElf,
}

var HighElfMagician Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 99,
    CombatLbxFile: "figures7.lbx",
    CombatIndex: 72,
    Name: "Magicians",
    ProductionCost: 180,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 2,
    MeleeAttackPower: 1,
    RangedAttackPower: 5,
    RangedAttacks: 4,
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 3,
    Resistance: 10,
    HitPoints: 1,
    // fireball 1x, tohit +10%
    Abilities: []Ability{AbilityFireballSpell, AbilityMissileImmunity, AbilityForester, AbilityToHit},
    RequiredBuildings: []building.Building{building.BuildingWizardsGuild},
    Race: data.RaceHighElf,
}

var HighElfSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Name: "Settlers",
    CombatLbxFile: "figures7.lbx",
    CombatIndex: 80,
    Count: 1,
    Index: 100,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Race: data.RaceHighElf,
    ProductionCost: 90,
    MovementSpeed: 1,
    Resistance: 6,
    MeleeAttackPower: 1,
    HitPoints: 1,
    Defense: 1,
    // tohit +10%
    Abilities: []Ability{AbilityCreateOutpost, AbilityToHit, AbilityForester},
}

var Longbowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 101,
    CombatLbxFile: "figures7.lbx",
    CombatIndex: 88,
    Name: "Longbowmen",
    ProductionCost: 80,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    RangedAttackPower: 3,
    RangedAttacks: 8,
    RangedAttackDamageType: DamageRangedPhysical,
    Defense: 2,
    Resistance: 6,
    HitPoints: 1,
    // tohit +10%
    Abilities: []Ability{AbilityForester, AbilityToHit},
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSawmill},
    Race: data.RaceHighElf,
}

var ElvenLord Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 102,
    CombatLbxFile: "figures7.lbx",
    CombatIndex: 96,
    Name: "ElvenLord",
    ProductionCost: 160,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 2,
    MeleeAttackPower: 5,
    Defense: 4,
    Resistance: 9,
    HitPoints: 3,
    // tohit +20%
    Abilities: []Ability{AbilityForester, AbilityToHit, AbilityArmorPiercing, AbilityFirstStrike},
    RequiredBuildings: []building.Building{building.BuildingArmorersGuild},
    Race: data.RaceHighElf,
}

var Pegasai Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 103,
    CombatLbxFile: "figures7.lbx",
    CombatIndex: 104,
    Name: "Pegasai",
    ProductionCost: 160,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 2,
    Flying: true,
    MovementSpeed: 3,
    MeleeAttackPower: 5,
    RangedAttackPower: 3,
    RangedAttacks: 8,
    RangedAttackDamageType: DamageRangedPhysical,
    Defense: 4,
    Resistance: 8,
    HitPoints: 5,
    RequiredBuildings: []building.Building{building.BuildingFantasticStable},
    Race: data.RaceHighElf,
}

var HighMenSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Name: "Spearmen",
    Index: 104,
    CombatLbxFile: "figures7.lbx",
    CombatIndex: 112,
    Race: data.RaceHighMen,
    ProductionCost: 10,
    UpkeepFood: 1,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    Defense: 2,
    Resistance: 4,
    HitPoints: 1,
    Count: 8,
}

var HighMenSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 105,
    CombatLbxFile: "figures8.lbx",
    CombatIndex: 0,
    Name: "Swordsmen",
    ProductionCost: 20,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 3,
    Defense: 2,
    Resistance: 4,
    HitPoints: 1,
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Abilities: []Ability{AbilityLargeShield},
    Race: data.RaceHighMen,
}

var HighMenBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 106,
    Race: data.RaceHighMen,
    Name: "Bowmen",
    UpkeepGold: 1,
    UpkeepFood: 1,
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSawmill},
    ProductionCost: 30,
    Count: 6,
    Defense: 1,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    RangedAttackPower: 1,
    RangeAttackIndex: 8,
    RangedAttacks: 8,
    RangeAttackSound: RangeAttackSoundArrow,
    RangedAttackDamageType: DamageRangedPhysical,
    MovementSound: MovementSoundShuffle,
    Resistance: 4,
    HitPoints: 1,
    CombatLbxFile: "figures8.lbx",
    CombatIndex: 8,
}

var HighMenCavalry Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 107,
    Name: "Cavalry",
    CombatLbxFile: "figures8.lbx",
    CombatIndex: 16,
    ProductionCost: 40,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 2,
    MeleeAttackPower: 4,
    Defense: 2,
    Resistance: 4,
    HitPoints: 3,
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingStables},
    Abilities: []Ability{AbilityFirstStrike},
    Race: data.RaceHighMen,
}

var HighMenPriest Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 108,
    CombatLbxFile: "figures8.lbx",
    CombatIndex: 24,
    Name: "Priests",
    ProductionCost: 100,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 1,
    MeleeAttackPower: 3,
    RangedAttackPower: 4,
    RangedAttacks: 4,
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 4,
    Resistance: 7,
    HitPoints: 1,
    // healing spell x1
    Abilities: []Ability{AbilityHealer, AbilityPurify, AbilityHealingSpell},
    RequiredBuildings: []building.Building{building.BuildingParthenon},
    Race: data.RaceHighMen,
}

var HighMenMagician Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 109,
    CombatLbxFile: "figures8.lbx",
    CombatIndex: 32,
    Name: "Magicians",
    ProductionCost: 120,
    UpkeepGold: 3,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    RangedAttackPower: 5,
    RangedAttacks: 4,
    Defense: 3,
    Resistance: 8,
    HitPoints: 1,
    // fireball 1x
    Abilities: []Ability{AbilityMissileImmunity, AbilityFireballSpell},
    RequiredBuildings: []building.Building{building.BuildingWizardsGuild},
    Race: data.RaceHighMen,
}

var HighMenEngineer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 110,
    CombatLbxFile: "figures8.lbx",
    CombatIndex: 40,
    Name: "Engineers",
    ProductionCost: 40,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    MeleeAttackPower: 1,
    Defense: 1,
    Resistance: 4,
    HitPoints: 1,
    MovementSpeed: 1,
    RequiredBuildings: []building.Building{building.BuildingBuildersHall},
    Abilities: []Ability{AbilityConstruction, AbilityWallCrusher},
    Race: data.RaceHighMen,
}

var HighMenSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Name: "Settlers",
    Index: 111,
    CombatLbxFile: "figures8.lbx",
    CombatIndex: 48,
    MovementSpeed: 1,
    ProductionCost: 60,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 1,
    Defense: 1,
    Resistance: 4,
    HitPoints: 10,
    Abilities: []Ability{AbilityCreateOutpost},
    Race: data.RaceHighMen,
}

var HighMenPikemen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 112,
    CombatLbxFile: "figures8.lbx",
    CombatIndex: 56,
    Name: "Pikemen",
    ProductionCost: 80,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 8,
    MovementSpeed: 1,
    MeleeAttackPower: 5,
    Defense: 3,
    Resistance: 5,
    HitPoints: 1,
    RequiredBuildings: []building.Building{building.BuildingFightersGuild},
    Abilities: []Ability{AbilityNegateFirstStrike, AbilityArmorPiercing},
    Race: data.RaceHighMen,
}

var Paladin Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 113,
    CombatLbxFile: "figures8.lbx",
    CombatIndex: 64,
    Name: "Paladins",
    ProductionCost: 200,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 4,
    MovementSpeed: 2,
    MeleeAttackPower: 6,
    Defense: 5,
    Resistance: 8,
    HitPoints: 4,
    RequiredBuildings: []building.Building{building.BuildingArmorersGuild, building.BuildingCathedral},
    // holy bonus 1x
    Abilities: []Ability{AbilityMagicImmunity, AbilityHolyBonus, AbilityArmorPiercing, AbilityFirstStrike},
    Race: data.RaceHighMen,
}

var KlackonSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 114,
    CombatLbxFile: "figures8.lbx",
    CombatIndex: 72,
    Name: "Spearmen",
    ProductionCost: 20,
    UpkeepFood: 1,
    Count: 8,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    Defense: 4,
    Resistance: 5,
    HitPoints: 1,
    Race: data.RaceKlackon,
}

var KlackonSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 115,
    Name: "Swordsmen",
    CombatLbxFile: "figures8.lbx",
    CombatIndex: 80,
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    ProductionCost: 40,
    UpkeepGold: 1,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 3,
    Defense: 4,
    Resistance: 5,
    HitPoints: 1,
    Abilities: []Ability{AbilityLargeShield},
    Race: data.RaceKlackon,
}

var KlackonHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 116,
    CombatLbxFile: "figures8.lbx",
    CombatIndex: 88,
    Name: "Halberdiers",
    RequiredBuildings: []building.Building{building.BuildingArmory},
    ProductionCost: 80,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 4,
    Defense: 5,
    Resistance: 5,
    HitPoints: 1,
    Race: data.RaceKlackon,
}

var KlackonEngineer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 117,
    CombatLbxFile: "figures8.lbx",
    CombatIndex: 96,
    Name: "Engineer",
    RequiredBuildings: []building.Building{building.BuildingBuildersHall},
    ProductionCost: 80,
    UpkeepGold: 2,
    UpkeepFood: 1,
    Count: 6,
    MovementSpeed: 1,
    MeleeAttackPower: 1,
    Defense: 1,
    Resistance: 5,
    HitPoints: 1,
    Abilities: []Ability{AbilityConstruction, AbilityWallCrusher},
    Race: data.RaceKlackon,
}

var KlackonSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 118,
    CombatLbxFile: "figures8.lbx",
    CombatIndex: 104,
    Name: "Settlers",
    UpkeepGold: 3,
    UpkeepFood: 1,
    Count: 1,
    ProductionCost: 120,
    MovementSpeed: 1,
    Defense: 1,
    Resistance: 5,
    HitPoints: 20,
    Race: data.RaceKlackon,
}

var StagBeetle Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 119,
    CombatLbxFile: "figures8.lbx",
    CombatIndex: 112,
    Name: "Stag Beetle",
    RequiredBuildings: []building.Building{building.BuildingArmorersGuild, building.BuildingFantasticStable},
    ProductionCost: 160,
    UpkeepGold: 4,
    UpkeepFood: 1,
    Count: 1,
    MovementSpeed: 2,
    MeleeAttackPower: 15,
    Defense: 7,
    Resistance: 6,
    HitPoints: 20,
    // fire breath 5
    Abilities: []Ability{AbilityFireBreath},
    Race: data.RaceKlackon,
}

// -------------------------------------------------
// -------------------------------------------------
// -------------------------------------------------

var AllUnits []Unit = []Unit{
    LizardSettlers,
    LizardSpearmen,
    LizardSwordsmen,
    LizardHalberdiers,
    LizardJavelineers,
    LizardShamans,
    DragonTurtle,
    NomadSettlers,
    NomadSpearmen,
    NomadSwordsmen,
    NomadBowmen,
    NomadPriest,
    NomadHorsebowemen,
    NomadPikemen,
    NomadRangers,
    Griffin,
    OrcSettlers,
    OrcSpearmen,
    OrcSwordsmen,
    OrcHalberdiers,
    OrcBowmen,
    OrcCavalry,
    OrcShamans,
    OrcMagicians,
    OrcEngineers,
    WyvernRiders,
    TrollSettlers,
    TrollSpearmen,
    TrollSwordsmen,
    TrollHalberdiers,
    TrollShamans,
    WarTrolls,
    WarMammoths,
    MagicSpirit,
    HellHounds,
    Gargoyle,
    FireGiant,
    FireElemental,
    ChaosSpawn,
    Chimeras,
    DoomBat,
    Efreet,
    Hydra,
    GreatDrake,
    Skeleton,
    Ghoul,
    NightStalker,
    WereWolf,
    Demon,
    Wraith,
    ShadowDemon,
    DeathKnight,
    DemonLord,
    Zombie,
    Unicorn,
    GuardianSpirit,
    Angel,
    ArchAngel,
    WarBear,
    Sprite,
    Cockatrice,
    Basilisk,
    GiantSpider,
    StoneGiant,
    Colossus,
    Gorgon,
    EarthElemental,
    Behemoth,
    GreatWyrm,
    FloatingIsland,
    PhantomBeast,
    PhantomWarrior,
    StormGiant,
    AirElemental,
    Djinn,
    SkyDrake,
    Nagas,
    HeroBrax,
    HeroGunther,
    HeroZaldron,
    HeroBShan,
    HeroRakir,
    HeroValana,
    HeroBahgtru,
    HeroSerena,
    HeroShuri,
    HeroTheria,
    HeroGreyfairer,
    HeroTaki,
    HeroReywind,
    HeroMalleus,
    HeroTumu,
    HeroJaer,
    HeroMarcus,
    HeroFang,
    HeroMorgana,
    HeroAureus,
    HeroShinBo,
    HeroSpyder,
    HeroShalla,
    HeroYramrag,
    HeroMysticX,
    HeroAerie,
    HeroDethStryke,
    HeroElana,
    HeroRoland,
    HeroMortu,
    HeroAlorra,
    HeroSirHarold,
    HeroRavashack,
    HeroWarrax,
    HeroTorin,
    Trireme,
    Galley,
    Catapult,
    Warship,
    BarbarianSettlers,
    BarbarianSpearmen,
    BarbarianSwordsmen,
    BarbarianBowmen,
    BarbarianCavalry,
    BarbarianShaman,
    Berserkers,
    BeastmenSettlers,
    BeastmenSpearmen,
    BeastmenSwordsmen,
    BeastmenHalberdiers,
    BeastmenBowmen,
    BeastmenPriest,
    BeastmenMagician,
    BeastmenEngineer,
    Centaur,
    Manticore,
    Minotaur,
    DarkElfSettlers,
    DarkElfSpearmen,
    DarkElfSwordsmen,
    DarkElfHalberdiers,
    DarkElfCavalry,
    DarkElfPriests,
    Nightblades,
    Warlocks,
    Nightmares,
    DraconianSettlers,
    DraconianSpearmen,
    DraconianSwordsmen,
    DraconianHalberdiers,
    DraconianBowmen,
    DraconianShaman,
    DraconianMagician,
    DoomDrake,
    AirShip,
    DwarfSettlers,
    DwarfSwordsmen,
    DwarfHalberdiers,
    DwarfEngineer,
    Hammerhands,
    SteamCannon,
    Golem,
    GnollSettlers,
    GnollSpearmen,
    GnollSwordsmen,
    GnollHalberdiers,
    GnollBowmen,
    WolfRiders,
    HalflingSettlers,
    HalflingSpearmen,
    HalflingSwordsmen,
    HalflingBowmen,
    HalflingShamans,
    Slingers,
    HighElfSettlers,
    HighElfSpearmen,
    HighElfSwordsmen,
    HighElfHalberdiers,
    HighElfCavalry,
    HighElfMagician,
    Longbowmen,
    ElvenLord,
    Pegasai,
    HighMenSettlers,
    HighMenEngineer,
    HighMenSpearmen,
    HighMenSwordsmen,
    HighMenBowmen,
    HighMenCavalry,
    HighMenPriest,
    HighMenMagician,
    HighMenPikemen,
    Paladin,
    KlackonSettlers,
    KlackonSpearmen,
    KlackonSwordsmen,
    KlackonHalberdiers,
    KlackonEngineer,
    StagBeetle,
}
