package units

import (
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

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
    LbxFile string
    CombatLbxFile string
    Index int
    // first index of combat tiles, order is always up, up-right, right, down-right, down, down-left, left, up-left
    CombatIndex int
    Name string
    Race data.Race

    // number of figures that are drawn in a single combat tile
    Count int
    // FIXME: add construction cost, building requirements to build this unit
    //  upkeep cost, how many figures appear in the battlefield, movement speed,
    //  attack power, ranged attack, defense, magic resistance, hit points, special power
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

func (unit *Unit) String() string {
    return unit.Name
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
    Race: data.RaceLizard,
}

var LizardHalberdiers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 2,
    Race: data.RaceLizard,
}

var LizardJavelineers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 3,
    Race: data.RaceLizard,
}

var LizardShamans Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 4,
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
    Race: data.RaceLizard,
}

var NomadSpearmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 7,
    Race: data.RaceNomad,
}

var NomadSwordsmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 8,
    Race: data.RaceNomad,
}

var NomadBowmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 9,
    Race: data.RaceNomad,
}

var NomadPriest Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 10,
    Race: data.RaceNomad,
}

// what is units2.lbx index 11?
// its some nomad unit holding a sword or something

var NomadSettlers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 12,
    Race: data.RaceNomad,
}

var NomadHorsebowemen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 13,
    Race: data.RaceNomad,
}

var NomadPikemen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 14,
    Race: data.RaceNomad,
}

var NomadRangers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 15,
    Race: data.RaceNomad,
}

var Griffin Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 16,
    // maybe race magical?
    Race: data.RaceNomad,
}

var OrcSpearmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 17,
    Race: data.RaceOrc,
}

var OrcSwordsmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 18,
    Race: data.RaceOrc,
}

var OrcHalberdiers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 19,
    Race: data.RaceOrc,
}

var OrcBowmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 20,
    Race: data.RaceOrc,
}

var OrcCavalry Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 21,
    Race: data.RaceOrc,
}

var OrcShamans Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 22,
    Race: data.RaceOrc,
}

var OrcMagicians Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 23,
    Race: data.RaceOrc,
}

var OrcEngineers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 24,
    Race: data.RaceOrc,
}

var OrcSettlers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 25,
    Race: data.RaceOrc,
}

var WyvernRiders Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 26,
    Race: data.RaceOrc,
}

var TrollSpearmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 27,
    Race: data.RaceTroll,
}

var TrollSwordsmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 28,
    Race: data.RaceTroll,
}

var TrollHalberdiers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 29,
    Race: data.RaceTroll,
}

var TrollShamans Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 30,
    Race: data.RaceTroll,
}

var TrollSettlers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 31,
    Race: data.RaceTroll,
}

var WarTrolls Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 32,
    Race: data.RaceTroll,
}

var WarMammoths Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 33,
    Race: data.RaceTroll,
}

var MagicSpirit Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 34,
    // FIXME: check on this
    Race: data.RaceFantastic,
}

var HellHounds Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 35,
    Race: data.RaceFantastic,
}

var Gargoyle Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 36,
    Race: data.RaceFantastic,
}

var FireGiant Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 37,
    Race: data.RaceFantastic,
}

var FireElemental Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 38,
    Race: data.RaceFantastic,
}

var ChaosSpawn Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 39,
    Race: data.RaceFantastic,
}

var Chimeras Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 40,
    Race: data.RaceFantastic,
}

var DoomBat Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 41,
    Race: data.RaceFantastic,
}

var Efreet Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 42,
    Race: data.RaceFantastic,
}

var Hydra Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 43,
    Race: data.RaceFantastic,
}

var GreatDrake Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 44,
    Race: data.RaceFantastic,
}

var Skeleton Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 45,
    Race: data.RaceFantastic,
}

var Ghoul Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 46,
    Race: data.RaceFantastic,
}

var NightStalker Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 47,
    Race: data.RaceFantastic,
}

var WereWolf Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 48,
    Race: data.RaceFantastic,
}

var Demon Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 49,
    Race: data.RaceFantastic,
}

var Wraith Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 50,
    Race: data.RaceFantastic,
}

var ShadowDemon Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 51,
    Race: data.RaceFantastic,
}

var DeathKnight Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 52,
    Race: data.RaceFantastic,
}

var DemonLord Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 53,
    Race: data.RaceFantastic,
}

var Zombie Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 54,
    Race: data.RaceFantastic,
}

var Unicorn Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 55,
    Race: data.RaceFantastic,
}

var GuardianSpirit Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 56,
    Race: data.RaceFantastic,
}

var Angel Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 57,
    Race: data.RaceFantastic,
}

var ArchAngel Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 58,
    Race: data.RaceFantastic,
}

var WarBear Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 59,
    Race: data.RaceFantastic,
}

var Sprite Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 60,
    Race: data.RaceFantastic,
}

var Cockatrice Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 61,
    Race: data.RaceFantastic,
}

var Basilisk Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 62,
    Race: data.RaceFantastic,
}

var GiantSpider Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 63,
    Race: data.RaceFantastic,
}

var StoneGiant Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 64,
    Race: data.RaceFantastic,
}

var Colossus Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 65,
    Race: data.RaceFantastic,
}

var Gorgon Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 66,
    Race: data.RaceFantastic,
}

var EarthElemental Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 67,
    Race: data.RaceFantastic,
}

var Behemoth Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 68,
    Race: data.RaceFantastic,
}

var GreatWyrm Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 69,
    Race: data.RaceFantastic,
}

var FloatingIsland Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 70,
    Race: data.RaceFantastic,
}

var PhantomBeast Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 71,
    Race: data.RaceFantastic,
}

var PhantomWarrior Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 72,
    Race: data.RaceFantastic,
}

var StormGiant Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 73,
    Race: data.RaceFantastic,
}

var AirElemental Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 74,
    Race: data.RaceFantastic,
}

var Djinn Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 75,
    Race: data.RaceFantastic,
}

var SkyDrake Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 76,
    Race: data.RaceFantastic,
}

var Nagas Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 77,
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

var HeroAeirie Unit = Unit{
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
    Race: data.RaceNone,
}

var Galley Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 36,
    Race: data.RaceNone,
}

var Catapult Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 37,
    Race: data.RaceNone,
}

var Warship Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 38,
    Race: data.RaceNone,
}

var BarbarianSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 39,
    Race: data.RaceBarbarian,
}

var BarbarianSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 40,
    Race: data.RaceBarbarian,
}

var BarbarianBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 41,
    Race: data.RaceBarbarian,
}

var BarbarianCavalry Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 42,
    Race: data.RaceBarbarian,
}

var BarbarianShaman Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 43,
    Race: data.RaceBarbarian,
}

var BarbarianSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 44,
    Race: data.RaceBarbarian,
}

var Berserkers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 45,
    Race: data.RaceBarbarian,
}

var BeastmenSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 46,
    Race: data.RaceBeastmen,
}

var BeastmenSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 47,
    Race: data.RaceBeastmen,
}

var BeastmenHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 48,
    Race: data.RaceBeastmen,
}

var BeastmenBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 49,
    Race: data.RaceBeastmen,
}

var BeastmenPriest Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 50,
    Race: data.RaceBeastmen,
}

var BeastmenMagician Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 51,
    Race: data.RaceBeastmen,
}

var BeastmenEngineer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 52,
    Race: data.RaceBeastmen,
}

var BeastmenSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 53,
    Race: data.RaceBeastmen,
}

var Centaur Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 54,
    Race: data.RaceBeastmen,
}

var Manticore Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 55,
    Race: data.RaceBeastmen,
}

var Minotaur Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 56,
    Race: data.RaceBeastmen,
}

var DarkElfSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 57,
    Race: data.RaceDarkElf,
}

var DarkElfSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 58,
    Race: data.RaceDarkElf,
}

var DarkElfHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 59,
    Race: data.RaceDarkElf,
}

var DarkElfCavalry Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 60,
    Race: data.RaceDarkElf,
}

var DarkElfPriests Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 61,
    Race: data.RaceDarkElf,
}

var DarkElfSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 62,
    Race: data.RaceDarkElf,
}

var Nightblades Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 63,
    Race: data.RaceDarkElf,
}

var Warlocks Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 64,
    Race: data.RaceDarkElf,
}

var Nightmares Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 65,
    Race: data.RaceDarkElf,
}

var DraconianSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 66,
    Race: data.RaceDraconian,
}

var DraconianSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 67,
    Race: data.RaceDraconian,
}

var DraconianHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 68,
    Race: data.RaceDraconian,
}

var DraconianBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 69,
    Race: data.RaceDraconian,
}

var DraconianShaman Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 70,
    Race: data.RaceDraconian,
}

var DraconianMagician Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 71,
    Race: data.RaceDraconian,
}

var DraconianEngineer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 72,
    Race: data.RaceDraconian,
}

var DraconianSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 73,
    Race: data.RaceDraconian,
}

var DoomDrake Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 74,
    Race: data.RaceDraconian,
}

var AirShip Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 75,
    Race: data.RaceDraconian,
}

var DwarfSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 76,
    Race: data.RaceDwarf,
}

var DwarfHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 77,
    Race: data.RaceDwarf,
}

var DwarfEngineer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 78,
    Race: data.RaceDwarf,
}

var Hammerhands Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 79,
    Race: data.RaceDwarf,
}

var SteamCannon Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 80,
    Race: data.RaceDwarf,
}

var Golem Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 81,
    Race: data.RaceDwarf,
}

var DwarfSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 82,
    Race: data.RaceDwarf,
}

var GnollSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 83,
    Race: data.RaceGnoll,
}

var GnollSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 84,
    Race: data.RaceGnoll,
}

var GnollHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 85,
    Race: data.RaceGnoll,
}

var GnollBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 86,
    Race: data.RaceGnoll,
}

var GnollSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 87,
    Race: data.RaceGnoll,
}

var WolfRiders Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 88,
    Race: data.RaceGnoll,
}

var HalflingSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 89,
    Race: data.RaceHalfling,
}

var HalflingSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 90,
    Race: data.RaceHalfling,
}

var HalflingBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 91,
    Race: data.RaceHalfling,
}

var HalflingShamans Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 92,
    Race: data.RaceHalfling,
}

var HalflingSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 93,
    Race: data.RaceHalfling,
}

var Slingers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 94,
    Race: data.RaceHalfling,
}

var HighElfSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 95,
    CombatLbxFile: "figures7.lbx",
    CombatIndex: 40,
    Count: 8,
    Name: "Spearmen",
    Race: data.RaceHighElf,
}

var HighElfSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 96,
    Race: data.RaceHighElf,
}

var HighElfHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 97,
    Race: data.RaceHighElf,
}

var HighElfCavalry Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 98,
    Race: data.RaceHighElf,
}

var HighElfMagician Unit = Unit{
    LbxFile: "units1.lbx",
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
}

var Longbowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 101,
    Race: data.RaceHighElf,
}

var ElvenLord Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 102,
    Race: data.RaceHighElf,
}

var Pegasai Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 103,
    Race: data.RaceHighElf,
}

var HighMenSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Name: "Spearmen",
    Index: 104,
    Race: data.RaceHighMen,
}

var HighMenSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 105,
    Race: data.RaceHighMen,
}

var HighMenBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 106,
    Race: data.RaceHighMen,
}

var HighMenCavalry Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 107,
    Race: data.RaceHighMen,
}

var HighMenPreist Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 108,
    Race: data.RaceHighMen,
}

var HighMenMagician Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 109,
    Race: data.RaceHighMen,
}

var HighMenEngineer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 110,
    Race: data.RaceHighMen,
}

var HighMenSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Name: "Settlers",
    Index: 111,
    Race: data.RaceHighMen,
}

var HighMenPikemen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 112,
    Race: data.RaceHighMen,
}

var Paladin Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 113,
    Race: data.RaceHighMen,
}

var KlackonSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 114,
    Race: data.RaceKlackon,
}

var KlackonSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 115,
    Race: data.RaceKlackon,
}

var KlackonHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 116,
    Race: data.RaceKlackon,
}

var KlackonEngineer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 117,
    Race: data.RaceKlackon,
}

var KlackonSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 118,
    Race: data.RaceKlackon,
}

var StagBeetle Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 119,
    Race: data.RaceKlackon,
}
