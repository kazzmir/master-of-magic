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
    Race: data.RaceLizard,
}

var LizardSwordsmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 1,
    Name: "Swordsmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceLizard,
}

var LizardHalberdiers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 2,
    Name: "Halberdiers",
    RequiredBuildings: []building.Building{building.BuildingArmory},
    Race: data.RaceLizard,
}

var LizardJavelineers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 3,
    Name: "Javelineers",
    RequiredBuildings: []building.Building{building.BuildingFightersGuild},
    Race: data.RaceLizard,
}

var LizardShamans Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 4,
    Name: "Shamans",
    RequiredBuildings: []building.Building{building.BuildingShrine},
    Race: data.RaceLizard,
}

var LizardSettlers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 5,
    Name: "Settlers",
    Race: data.RaceLizard,
}

var DragonTurtle Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 6,
    Name: "Dragon Turtle",
    RequiredBuildings: []building.Building{building.BuildingArmorersGuild, building.BuildingStables},
    Race: data.RaceLizard,
}

var NomadSpearmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 7,
    Name: "Spearmen",
    Race: data.RaceNomad,
}

var NomadSwordsmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 8,
    Name: "Swordsmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceNomad,
}

var NomadBowmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 9,
    Name: "Bowmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSawmill},
    Race: data.RaceNomad,
}

var NomadPriest Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 10,
    Name: "Priests",
    RequiredBuildings: []building.Building{building.BuildingParthenon},
    Race: data.RaceNomad,
}

// what is units2.lbx index 11?
// its some nomad unit holding a sword or something

var NomadSettlers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 12,
    Name: "Settlers",
    Race: data.RaceNomad,
}

var NomadHorsebowemen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 13,
    Name: "Horsebowmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingStables},
    Race: data.RaceNomad,
}

var NomadPikemen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 14,
    Name: "Pikemen",
    RequiredBuildings: []building.Building{building.BuildingFightersGuild},
    Race: data.RaceNomad,
}

var NomadRangers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 15,
    Name: "Rangers",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingAnimistsGuild},
    Race: data.RaceNomad,
}

var Griffin Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 16,
    Name: "Griffins",
    RequiredBuildings: []building.Building{building.BuildingFantasticStable},
    // maybe race magical?
    Race: data.RaceNomad,
}

var OrcSpearmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 17,
    Name: "Spearmen",
    Race: data.RaceOrc,
}

var OrcSwordsmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 18,
    Name: "Swordsmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceOrc,
}

var OrcHalberdiers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 19,
    Name: "Halberdiers",
    RequiredBuildings: []building.Building{building.BuildingArmory},
    Race: data.RaceOrc,
}

var OrcBowmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 20,
    Name: "Bowmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSawmill},
    Race: data.RaceOrc,
}

var OrcCavalry Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 21,
    Name: "Calvary",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingStables},
    Race: data.RaceOrc,
}

var OrcShamans Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 22,
    Name: "Shamans",
    RequiredBuildings: []building.Building{building.BuildingShrine},
    Race: data.RaceOrc,
}

var OrcMagicians Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 23,
    Name: "Magicians",
    RequiredBuildings: []building.Building{building.BuildingWizardsGuild},
    Race: data.RaceOrc,
}

var OrcEngineers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 24,
    Name: "Engineers",
    RequiredBuildings: []building.Building{building.BuildingBuildersHall},
    Race: data.RaceOrc,
}

var OrcSettlers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 25,
    Name: "Settlers",
    Race: data.RaceOrc,
}

var WyvernRiders Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 26,
    Name: "Wyvern Riders",
    RequiredBuildings: []building.Building{building.BuildingFantasticStable},
    Race: data.RaceOrc,
}

var TrollSpearmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 27,
    Name: "Spearmen",
    Race: data.RaceTroll,
}

var TrollSwordsmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 28,
    Name: "Swordsmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceTroll,
}

var TrollHalberdiers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 29,
    Name: "Halberdiers",
    RequiredBuildings: []building.Building{building.BuildingArmory},
    Race: data.RaceTroll,
}

var TrollShamans Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 30,
    Name: "Shamans",
    RequiredBuildings: []building.Building{building.BuildingShrine},
    Race: data.RaceTroll,
}

var TrollSettlers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 31,
    Name: "Settlers",
    Race: data.RaceTroll,
}

var WarTrolls Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 32,
    Name: "War Trolls",
    RequiredBuildings: []building.Building{building.BuildingFightersGuild},
    Race: data.RaceTroll,
}

var WarMammoths Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 33,
    Name: "War Mammoths",
    RequiredBuildings: []building.Building{building.BuildingArmorersGuild, building.BuildingStables},
    Race: data.RaceTroll,
}

var MagicSpirit Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 34,
    Realm: data.ArcaneMagic,
    Abilities: []Ability{AbilityMeld, AbilityNonCorporeal},
    MovementSpeed: 1,
    // FIXME: check on this
    Race: data.RaceFantastic,
}

var HellHounds Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 35,
    Name: "Hell Hounds",
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 40,
    MovementSpeed: 2,
    Realm: data.ChaosMagic,
    MeleeAttackPower: 3,
    Count: 4,
    Defense: 2,
    Resistance: 6,
    HitPoints: 4,
    Race: data.RaceFantastic,
}

var Gargoyle Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 36,
    Name: "Gargoyles",
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 48,
    Realm: data.ChaosMagic,
    MovementSpeed: 2,
    Flying: true,
    MeleeAttackPower: 4,
    Count: 4,
    Defense: 8,
    Resistance: 7,
    HitPoints: 4,
    Race: data.RaceFantastic,
}

var FireGiant Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 37,
    Name: "Fire Giant",
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 56,
    MovementSpeed: 2,
    Realm: data.ChaosMagic,
    Count: 1,
    MeleeAttackPower: 10,
    RangedAttackPower: 10,
    RangedAttackDamageType: DamageRangedBoulder,
    RangedAttacks: 2,
    HitPoints: 15,
    Defense: 5,
    Resistance: 7,
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
}

var ChaosSpawn Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 39,
    Name: "Chaos Spawn",
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 72,
    Realm: data.ChaosMagic,
    MovementSpeed: 1,
    Flying: true,
    Count: 1,
    MeleeAttackPower: 1,
    Defense: 6,
    Resistance: 10,
    HitPoints: 15,
    Race: data.RaceFantastic,
}

var Chimeras Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 40,
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 80,
    Realm: data.ChaosMagic,
    MovementSpeed: 2,
    Flying: true,
    Count: 4,
    MeleeAttackPower: 7,
    Defense: 5,
    Resistance: 8,
    HitPoints: 8,
    Race: data.RaceFantastic,
}

var DoomBat Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 41,
    Name: "Doom Bat",
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 88,
    MovementSpeed: 4,
    Realm: data.ChaosMagic,
    Flying: true,
    MeleeAttackPower: 10,
    Defense: 5,
    Resistance: 9,
    Count: 1,
    HitPoints: 20,
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
    Flying: true,
    Count: 1,
    // caster 20
    MeleeAttackPower: 9,
    RangedAttackPower: 9,
    RangedAttackDamageType: DamageRangedMagical,
    Defense: 7,
    Resistance: 10,
    HitPoints: 12,
    Race: data.RaceFantastic,
}

// FIXME: hydra has 9 virtual figures, one for each head
var Hydra Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 43,
    Name: "Hydra",
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 104,
    Count: 1,
    Realm: data.ChaosMagic,
    MovementSpeed: 1,
    MeleeAttackPower: 6,
    Defense: 4,
    Resistance:11,
    HitPoints: 10,
    Race: data.RaceFantastic,
}

var GreatDrake Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 44,
    CombatLbxFile: "figure11.lbx",
    CombatIndex: 112,
    Name: "Great Drake",
    Race: data.RaceFantastic,
    Abilities: []Ability{AbilityForester, AbilityDoomGaze},
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
    Race: data.RaceFantastic,
    Realm: data.DeathMagic,
}

var Ghoul Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 46,
    Realm: data.DeathMagic,
    Race: data.RaceFantastic,
}

var NightStalker Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 47,
    Realm: data.DeathMagic,
    Race: data.RaceFantastic,
}

var WereWolf Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 48,
    Realm: data.DeathMagic,
    Race: data.RaceFantastic,
}

var Demon Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 49,
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
}

var Wraith Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 50,
    Realm: data.DeathMagic,
    Race: data.RaceFantastic,
}

var ShadowDemon Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 51,
    Realm: data.DeathMagic,
    Race: data.RaceFantastic,
}

var DeathKnight Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 52,
    Realm: data.DeathMagic,
    Race: data.RaceFantastic,
}

var DemonLord Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 53,
    Realm: data.DeathMagic,
    Race: data.RaceFantastic,
}

var Zombie Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 54,
    Realm: data.DeathMagic,
    Race: data.RaceFantastic,
}

var Unicorn Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 55,
    Race: data.RaceFantastic,
    Realm: data.LifeMagic,
}

var GuardianSpirit Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 56,
    Race: data.RaceFantastic,
    Abilities: []Ability{AbilityMeld, AbilityNonCorporeal, AbilityResistanceToAll},
    Realm: data.LifeMagic,
}

var Angel Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 57,
    Race: data.RaceFantastic,
    Realm: data.LifeMagic,
}

var ArchAngel Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 58,
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
    Race: data.RaceHero,
}

var HeroGunther Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 1,
    Race: data.RaceHero,
}

var HeroZaldron Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 2,
    Race: data.RaceHero,
}

var HeroBShan Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 3,
    Race: data.RaceHero,
}

var HeroRakir Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 4,
    Name: "Rakir",
    Race: data.RaceHero,
}

var HeroValana Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 5,
    Race: data.RaceHero,
}

var HeroBahgtru Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 6,
    Race: data.RaceHero,
}

var HeroSerena Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 7,
    Race: data.RaceHero,
}

var HeroShuri Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 8,
    Race: data.RaceHero,
}

var HeroTheria Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 9,
    Race: data.RaceHero,
}

var HeroGreyfairer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 10,
    Race: data.RaceHero,
}

var HeroTaki Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 11,
    Race: data.RaceHero,
}

var HeroReywind Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 12,
    Race: data.RaceHero,
}

var HeroMalleus Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 13,
    Race: data.RaceHero,
}

var HeroTumu Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 14,
    Race: data.RaceHero,
}

var HeroJaer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 15,
    Race: data.RaceHero,
}

var HeroMarcus Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 16,
    Race: data.RaceHero,
}

var HeroFang Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 17,
    Race: data.RaceHero,
}

var HeroMorgana Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 18,
    Race: data.RaceHero,
}

var HeroAureus Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 19,
    Race: data.RaceHero,
}

var HeroShinBo Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 20,
    Race: data.RaceHero,
}

var HeroSpyder Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 21,
    Race: data.RaceHero,
}

var HeroShalla Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 22,
    Race: data.RaceHero,
}

var HeroYramrag Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 23,
    Race: data.RaceHero,
}

var HeroMysticX Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 24,
    Race: data.RaceHero,
}

var HeroAerie Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 25,
    Race: data.RaceHero,
}

var HeroDethStryke Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 26,
    Race: data.RaceHero,
}

var HeroElana Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 27,
    Race: data.RaceHero,
}

var HeroRoland Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 28,
    Race: data.RaceHero,
}

var HeroMortu Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 29,
    Race: data.RaceHero,
}

var HeroAlorra Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 30,
    Race: data.RaceHero,
}

var HeroSirHarold Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 31,
    Race: data.RaceHero,
}

var HeroRavashack Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 32,
    Race: data.RaceHero,
}

var HeroWarrax Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 33,
    Race: data.RaceHero,
}

var HeroTorin Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 34,
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
    Race: data.RaceBarbarian,
}

var BarbarianSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 40,
    Name: "Swordsmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceBarbarian,
}

var BarbarianBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 41,
    Name: "Bowmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSawmill},
    Race: data.RaceBarbarian,
}

var BarbarianCavalry Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 42,
    Name: "Cavalry",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingStables},
    Race: data.RaceBarbarian,
}

var BarbarianShaman Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 43,
    Name: "Shaman",
    RequiredBuildings: []building.Building{building.BuildingShrine},
    Race: data.RaceBarbarian,
}

var BarbarianSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 44,
    Name: "Settlers",
    Race: data.RaceBarbarian,
}

var Berserkers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 45,
    Name: "Berserkers",
    RequiredBuildings: []building.Building{building.BuildingArmorersGuild},
    Race: data.RaceBarbarian,
}

var BeastmenSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 46,
    Name: "Spearmen",
    Race: data.RaceBeastmen,
}

var BeastmenSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 47,
    Name: "Swordsmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceBeastmen,
}

var BeastmenHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 48,
    Name: "Halberdiers",
    RequiredBuildings: []building.Building{building.BuildingArmory},
    Race: data.RaceBeastmen,
}

var BeastmenBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 49,
    Name: "Bowmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSawmill},
    Race: data.RaceBeastmen,
}

var BeastmenPriest Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 50,
    Name: "Priests",
    RequiredBuildings: []building.Building{building.BuildingParthenon},
    Race: data.RaceBeastmen,
}

var BeastmenMagician Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 51,
    Name: "Magicians",
    RequiredBuildings: []building.Building{building.BuildingWizardsGuild},
    Race: data.RaceBeastmen,
}

var BeastmenEngineer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 52,
    Name: "Engineers",
    RequiredBuildings: []building.Building{building.BuildingBuildersHall},
    Race: data.RaceBeastmen,
}

var BeastmenSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 53,
    Name: "Settlers",
    Race: data.RaceBeastmen,
}

var Centaur Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 54,
    Name: "Centaurs",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingStables},
    Race: data.RaceBeastmen,
}

var Manticore Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 55,
    Name: "Manticores",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingAnimistsGuild},
    Race: data.RaceBeastmen,
}

var Minotaur Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 56,
    Name: "Minotaurs",
    RequiredBuildings: []building.Building{building.BuildingArmorersGuild},
    Race: data.RaceBeastmen,
}

var DarkElfSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 57,
    Name: "Spearmen",
    Race: data.RaceDarkElf,
}

var DarkElfSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 58,
    Name: "Swordsmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceDarkElf,
}

var DarkElfHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 59,
    Name: "Halberdiers",
    RequiredBuildings: []building.Building{building.BuildingArmory},
    Race: data.RaceDarkElf,
}

var DarkElfCavalry Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 60,
    Name: "Cavalry",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingStables},
    Race: data.RaceDarkElf,
}

var DarkElfPriests Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 61,
    Name: "Priests",
    RequiredBuildings: []building.Building{building.BuildingParthenon},
    Race: data.RaceDarkElf,
}

var DarkElfSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 62,
    Name: "Settlers",
    Race: data.RaceDarkElf,
}

var Nightblades Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 63,
    Name: "Nightblades",
    RequiredBuildings: []building.Building{building.BuildingFightersGuild},
    Race: data.RaceDarkElf,
}

var Warlocks Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 64,
    Race: data.RaceDarkElf,
    Count: 4,
    Name: "Warlocks",
    RequiredBuildings: []building.Building{building.BuildingWizardsGuild},
    RangedAttackDamageType: DamageRangedMagical,
    RangeAttackIndex: 16,
    MeleeAttackPower: 1,
    RangedAttackPower: 7,
    RangeAttackSound: RangeAttackSoundFireball,
    Defense: 4,
    Resistance: 9,
    HitPoints: 1,
    RangedAttacks: 4,
    MovementSpeed: 1,
    MovementSound: MovementSoundShuffle,
    Abilities: []Ability{AbilityDoomBoltSpell, AbilityMissileImmunity},
    CombatLbxFile: "figures5.lbx",
    CombatIndex: 32,
}

var Nightmares Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 65,
    Name: "Nightmares",
    RequiredBuildings: []building.Building{building.BuildingFantasticStable},
    Race: data.RaceDarkElf,
}

var DraconianSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 66,
    Name: "Spearmen",
    Race: data.RaceDraconian,
}

var DraconianSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 67,
    Name: "Swordsmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceDraconian,
}

var DraconianHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 68,
    Name: "Halberdiers",
    RequiredBuildings: []building.Building{building.BuildingArmory},
    Race: data.RaceDraconian,
}

var DraconianBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 69,
    Name: "Bowmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSawmill},
    Race: data.RaceDraconian,
}

var DraconianShaman Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 70,
    Name: "Shaman",
    RequiredBuildings: []building.Building{building.BuildingShrine},
    Race: data.RaceDraconian,
}

var DraconianMagician Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 71,
    Name: "Magicians",
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
    Name: "Settlers",
    Race: data.RaceDraconian,
}

var DoomDrake Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 74,
    Name: "Doom Drake",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingStables},
    Race: data.RaceDraconian,
}

var AirShip Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 75,
    Name: "Air Ship",
    RequiredBuildings: []building.Building{building.BuildingShipYard},
    Race: data.RaceDraconian,
}

var DwarfSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 76,
    Name: "Swordsmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceDwarf,
}

var DwarfHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 77,
    Name: "Halberdiers",
    RequiredBuildings: []building.Building{building.BuildingArmory},
    Race: data.RaceDwarf,
}

var DwarfEngineer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 78,
    Name: "Engineers",
    RequiredBuildings: []building.Building{building.BuildingBuildersHall},
    Race: data.RaceDwarf,
}

var Hammerhands Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 79,
    Name: "Hammerhands",
    RequiredBuildings: []building.Building{building.BuildingFightersGuild},
    Race: data.RaceDwarf,
}

var SteamCannon Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 80,
    Name: "Steam Cannon",
    RequiredBuildings: []building.Building{building.BuildingMinersGuild},
    Race: data.RaceDwarf,
}

var Golem Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 81,
    Name: "Golem",
    RequiredBuildings: []building.Building{building.BuildingArmorersGuild},
    Race: data.RaceDwarf,
}

var DwarfSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 82,
    Name: "Settlers",
    Race: data.RaceDwarf,
}

var GnollSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 83,
    Name: "Spearmen",
    Race: data.RaceGnoll,
}

var GnollSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 84,
    Name: "Swordsmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceGnoll,
}

var GnollHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 85,
    Name: "Halberdiers",
    RequiredBuildings: []building.Building{building.BuildingArmory},
    Race: data.RaceGnoll,
}

var GnollBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 86,
    Name: "Bowmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSawmill},
    Race: data.RaceGnoll,
}

var GnollSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 87,
    Name: "Settlers",
    Race: data.RaceGnoll,
}

var WolfRiders Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 88,
    Name: "Wolf Riders",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingStables},
    Race: data.RaceGnoll,
}

var HalflingSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 89,
    Name: "Spearmen",
    Race: data.RaceHalfling,
}

var HalflingSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 90,
    Name: "Swordsmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceHalfling,
}

var HalflingBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 91,
    Name: "Bowmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSawmill},
    Race: data.RaceHalfling,
}

var HalflingShamans Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 92,
    Name: "Shamans",
    RequiredBuildings: []building.Building{building.BuildingShrine},
    Race: data.RaceHalfling,
}

var HalflingSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 93,
    Name: "Settlers",
    Race: data.RaceHalfling,
}

var Slingers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 94,
    Name: "Slingers",
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
    Abilities: []Ability{AbilityForester},
    MovementSpeed: 1,
    Race: data.RaceHighElf,
    HitPoints: 1,
    UpkeepFood: 1,
}

var HighElfSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Name: "Swordsmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Index: 96,
    Race: data.RaceHighElf,
}

var HighElfHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Name: "Halberdiers",
    RequiredBuildings: []building.Building{building.BuildingArmory},
    Index: 97,
    Race: data.RaceHighElf,
}

var HighElfCavalry Unit = Unit{
    LbxFile: "units1.lbx",
    Name: "Cavalry",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingStables},
    Index: 98,
    Race: data.RaceHighElf,
}

var HighElfMagician Unit = Unit{
    LbxFile: "units1.lbx",
    Name: "Magicians",
    RequiredBuildings: []building.Building{building.BuildingWizardsGuild},
    Index: 99,
    Race: data.RaceHighElf,
}

var HighElfSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Name: "Settlers",
    CombatLbxFile: "figures7.lbx",
    CombatIndex: 80,
    Count: 1,
    Index: 100,
    Race: data.RaceHighElf,
    ProductionCost: 90,
    MovementSpeed: 1,
    Resistance: 6,
    MeleeAttackPower: 1,
    HitPoints: 1,
    Defense: 1,
    Abilities: []Ability{AbilityCreateOutpost},
}

var Longbowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 101,
    Name: "Longbowmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSawmill},
    Race: data.RaceHighElf,
}

var ElvenLord Unit = Unit{
    LbxFile: "units1.lbx",
    Name: "ElvenLord",
    RequiredBuildings: []building.Building{building.BuildingArmorersGuild},
    Index: 102,
    Race: data.RaceHighElf,
}

var Pegasai Unit = Unit{
    LbxFile: "units1.lbx",
    Name: "Pegasai",
    RequiredBuildings: []building.Building{building.BuildingFantasticStable},
    Index: 103,
    Race: data.RaceHighElf,
}

var HighMenSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Name: "Spearmen",
    Index: 104,
    CombatLbxFile: "figures7.lbx",
    CombatIndex: 112,
    Race: data.RaceHighMen,
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
    Name: "Swordsmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceHighMen,
}

var HighMenBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 106,
    Race: data.RaceHighMen,
    Name: "Bowmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSawmill},
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
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingStables},
    Race: data.RaceHighMen,
}

var HighMenPriest Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 108,
    Name: "Priests",
    RequiredBuildings: []building.Building{building.BuildingParthenon},
    Race: data.RaceHighMen,
}

var HighMenMagician Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 109,
    Name: "Magicians",
    RequiredBuildings: []building.Building{building.BuildingWizardsGuild},
    Race: data.RaceHighMen,
}

var HighMenEngineer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 110,
    Name: "Engineers",
    RequiredBuildings: []building.Building{building.BuildingBuildersHall},
    Race: data.RaceHighMen,
}

var HighMenSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Name: "Settlers",
    Index: 111,
    MovementSpeed: 1,
    ProductionCost: 60,
    Abilities: []Ability{AbilityCreateOutpost},
    Race: data.RaceHighMen,
}

var HighMenPikemen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 112,
    Name: "Pikemen",
    RequiredBuildings: []building.Building{building.BuildingFightersGuild},
    Race: data.RaceHighMen,
}

var Paladin Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 113,
    Name: "Paladins",
    RequiredBuildings: []building.Building{building.BuildingArmorersGuild, building.BuildingCathedral},
    Race: data.RaceHighMen,
}

var KlackonSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 114,
    Name: "Spearmen",
    Race: data.RaceKlackon,
}

var KlackonSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 115,
    Name: "Swordsmen",
    RequiredBuildings: []building.Building{building.BuildingBarracks, building.BuildingSmithy},
    Race: data.RaceKlackon,
}

var KlackonHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 116,
    Name: "Halberdiers",
    RequiredBuildings: []building.Building{building.BuildingArmory},
    Race: data.RaceKlackon,
}

var KlackonEngineer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 117,
    Name: "Engineer",
    RequiredBuildings: []building.Building{building.BuildingBuildersHall},
    Race: data.RaceKlackon,
}

var KlackonSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 118,
    Name: "Settlers",
    Race: data.RaceKlackon,
}

var StagBeetle Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 119,
    Name: "Stag Beetle",
    RequiredBuildings: []building.Building{building.BuildingArmorersGuild, building.BuildingFantasticStable},
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
