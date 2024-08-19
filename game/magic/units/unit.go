package units

// FIXME: probably move this somewhere else
type Race int

const (
    RaceLizard Race = iota
    RaceNomad
    RaceOrc
    RaceTroll
    RaceFantastic
    RaceHero
    RaceNone
    RaceBarbarian
    RaceBeastmen
    RaceDarkElf
    RaceDraconian
    RaceDwarf
    RaceGnoll
    RaceHalfling
)

type Unit struct {
    LbxFile string
    Index int
    Race Race
    // FIXME: add construction cost, building requirements to build this unit
    //  upkeep cost, how many figures appear in the battlefield, movement speed,
    //  attack power, ranged attack, defense, magic resistance, hit points, special power
}

var LizardSpearmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 0,
    Race: RaceLizard,
}

var LizardSwordsmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 1,
    Race: RaceLizard,
}

var LizardHalberdiers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 2,
    Race: RaceLizard,
}

var LizardJavelineers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 3,
    Race: RaceLizard,
}

var LizardShamans Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 4,
    Race: RaceLizard,
}

var LizardSettlers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 5,
    Race: RaceLizard,
}

var DragonTurtle Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 6,
    Race: RaceLizard,
}

var NomadSpearmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 7,
    Race: RaceNomad,
}

var NomadSwordsmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 8,
    Race: RaceNomad,
}

var NomadBowmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 9,
    Race: RaceNomad,
}

var NomadPriest Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 10,
    Race: RaceNomad,
}

// what is units2.lbx index 11?
// its some nomad unit holding a sword or something

var NomadSettlers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 12,
    Race: RaceNomad,
}

var NomadHorsebowemen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 13,
    Race: RaceNomad,
}

var NomadPikemen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 14,
    Race: RaceNomad,
}

var NomadRangers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 15,
    Race: RaceNomad,
}

var Griffin Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 16,
    // maybe race magical?
    Race: RaceNomad,
}

var OrcSpearmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 17,
    Race: RaceOrc,
}

var OrcSwordsmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 18,
    Race: RaceOrc,
}

var OrcHalberdiers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 19,
    Race: RaceOrc,
}

var OrcBowmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 20,
    Race: RaceOrc,
}

var OrcCavalry Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 21,
    Race: RaceOrc,
}

var OrcShamans Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 22,
    Race: RaceOrc,
}

var OrcMagicians Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 23,
    Race: RaceOrc,
}

var OrcEngineers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 24,
    Race: RaceOrc,
}

var OrcSettlers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 25,
    Race: RaceOrc,
}

var WyvernRiders Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 26,
    Race: RaceOrc,
}

var TrollSpearmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 27,
    Race: RaceTroll,
}

var TrollSwordsmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 28,
    Race: RaceTroll,
}

var TrollHalberdiers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 29,
    Race: RaceTroll,
}

var TrollShamans Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 30,
    Race: RaceTroll,
}

var TrollSettlers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 31,
    Race: RaceTroll,
}

var WarTrolls Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 32,
    Race: RaceTroll,
}

var WarMammoths Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 33,
    Race: RaceTroll,
}

var MagicSpirit Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 34,
    // FIXME: check on this
    Race: RaceFantastic,
}

var HellHounds Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 35,
    Race: RaceFantastic,
}

var Gargoyle Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 36,
    Race: RaceFantastic,
}

var FireGiant Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 37,
    Race: RaceFantastic,
}

var FireElemental Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 38,
    Race: RaceFantastic,
}

var ChaosSpawn Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 39,
    Race: RaceFantastic,
}

var Chimeras Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 40,
    Race: RaceFantastic,
}

var DoomBat Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 41,
    Race: RaceFantastic,
}

var Efreet Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 42,
    Race: RaceFantastic,
}

var Hydra Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 43,
    Race: RaceFantastic,
}

var GreatDrake Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 44,
    Race: RaceFantastic,
}

var Skeleton Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 45,
    Race: RaceFantastic,
}

var Ghoul Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 46,
    Race: RaceFantastic,
}

var NightStalker Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 47,
    Race: RaceFantastic,
}

var WereWolf Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 48,
    Race: RaceFantastic,
}

var Demon Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 49,
    Race: RaceFantastic,
}

var Wraith Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 50,
    Race: RaceFantastic,
}

var ShadowDemon Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 51,
    Race: RaceFantastic,
}

var DeathKnight Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 52,
    Race: RaceFantastic,
}

var DemonLord Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 53,
    Race: RaceFantastic,
}

var Zombie Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 54,
    Race: RaceFantastic,
}

var Unicorn Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 55,
    Race: RaceFantastic,
}

var GuardianSpirit Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 56,
    Race: RaceFantastic,
}

var Angel Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 57,
    Race: RaceFantastic,
}

var ArchAngel Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 58,
    Race: RaceFantastic,
}

var WarBear Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 59,
    Race: RaceFantastic,
}

var Sprite Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 60,
    Race: RaceFantastic,
}

var Cockatrice Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 61,
    Race: RaceFantastic,
}

var Basilisk Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 62,
    Race: RaceFantastic,
}

var GiantSpider Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 63,
    Race: RaceFantastic,
}

var StoneGiant Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 64,
    Race: RaceFantastic,
}

var Colossus Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 65,
    Race: RaceFantastic,
}

var Gorgon Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 66,
    Race: RaceFantastic,
}

var EarthElemental Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 67,
    Race: RaceFantastic,
}

var Behemoth Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 68,
    Race: RaceFantastic,
}

var GreatWyrm Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 69,
    Race: RaceFantastic,
}

var FloatingIsland Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 70,
    Race: RaceFantastic,
}

var PhantomBeast Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 71,
    Race: RaceFantastic,
}

var PhantomWarrior Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 72,
    Race: RaceFantastic,
}

var StormGiant Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 73,
    Race: RaceFantastic,
}

var AirElemental Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 74,
    Race: RaceFantastic,
}

var Djinn Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 75,
    Race: RaceFantastic,
}

var SkyDrake Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 76,
    Race: RaceFantastic,
}

var Nagas Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 77,
    Race: RaceFantastic,
}

var HeroBrax Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 0,
    Race: RaceHero,
}

var HeroGunther Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 1,
    Race: RaceHero,
}

var HeroZaldron Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 2,
    Race: RaceHero,
}

var HeroBShan Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 3,
    Race: RaceHero,
}

var HeroRakir Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 4,
    Race: RaceHero,
}

var HeroValana Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 5,
    Race: RaceHero,
}

var HeroBahgtru Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 6,
    Race: RaceHero,
}

var HeroSerena Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 7,
    Race: RaceHero,
}

var HeroShuri Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 8,
    Race: RaceHero,
}

var HeroTheria Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 9,
    Race: RaceHero,
}

var HeroGreyfairer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 10,
    Race: RaceHero,
}

var HeroTaki Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 11,
    Race: RaceHero,
}

var HeroReywind Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 12,
    Race: RaceHero,
}

var HeroMalleus Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 13,
    Race: RaceHero,
}

var HeroTumu Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 14,
    Race: RaceHero,
}

var HeroJaer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 15,
    Race: RaceHero,
}

var HeroMarcus Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 16,
    Race: RaceHero,
}

var HeroFang Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 17,
    Race: RaceHero,
}

var HeroMorgana Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 18,
    Race: RaceHero,
}

var HeroAureus Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 19,
    Race: RaceHero,
}

var HeroShinBo Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 20,
    Race: RaceHero,
}

var HeroSpyder Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 21,
    Race: RaceHero,
}

var HeroShalla Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 22,
    Race: RaceHero,
}

var HeroYramrag Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 23,
    Race: RaceHero,
}

var HeroMysticX Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 24,
    Race: RaceHero,
}

var HeroAeirie Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 25,
    Race: RaceHero,
}

var HeroDethStryke Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 26,
    Race: RaceHero,
}

var HeroElana Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 27,
    Race: RaceHero,
}

var HeroRoland Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 28,
    Race: RaceHero,
}

var HeroMortu Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 29,
    Race: RaceHero,
}

var HeroAlorra Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 30,
    Race: RaceHero,
}

var HeroSirHarold Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 31,
    Race: RaceHero,
}

var HeroRavashack Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 32,
    Race: RaceHero,
}

var HeroWarrax Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 33,
    Race: RaceHero,
}

var HeroTorin Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 34,
    Race: RaceHero,
}

var Trireme Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 35,
    Race: RaceNone,
}

var Galley Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 36,
    Race: RaceNone,
}

var Catapult Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 37,
    Race: RaceNone,
}

var Warship Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 38,
    Race: RaceNone,
}

var BarbarianSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 39,
    Race: RaceBarbarian,
}

var BarbarianSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 40,
    Race: RaceBarbarian,
}

var BarbarianBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 41,
    Race: RaceBarbarian,
}

var BarbarianCavalry Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 42,
    Race: RaceBarbarian,
}

var BarbarianShaman Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 43,
    Race: RaceBarbarian,
}

var BarbarianSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 44,
    Race: RaceBarbarian,
}

var Berserkers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 45,
    Race: RaceBarbarian,
}

var BeastmenSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 46,
    Race: RaceBeastmen,
}

var BeastmenSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 47,
    Race: RaceBeastmen,
}

var BeastmenHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 48,
    Race: RaceBeastmen,
}

var BeastmenBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 49,
    Race: RaceBeastmen,
}

var BeastmenPriest Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 50,
    Race: RaceBeastmen,
}

var BeastmenMagician Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 51,
    Race: RaceBeastmen,
}

var BeastmenEngineer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 52,
    Race: RaceBeastmen,
}

var BeastmenSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 53,
    Race: RaceBeastmen,
}

var Centaur Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 54,
    Race: RaceBeastmen,
}

var Manticore Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 55,
    Race: RaceBeastmen,
}

var Minotaur Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 56,
    Race: RaceBeastmen,
}

var DarkElfSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 57,
    Race: RaceDarkElf,
}

var DarkElfSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 58,
    Race: RaceDarkElf,
}

var DarkElfHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 59,
    Race: RaceDarkElf,
}

var DarkElfCavalry Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 60,
    Race: RaceDarkElf,
}

var DarkElfPriests Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 61,
    Race: RaceDarkElf,
}

var DarkElfSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 62,
    Race: RaceDarkElf,
}

var Nightblades Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 63,
    Race: RaceDarkElf,
}

var Warlocks Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 64,
    Race: RaceDarkElf,
}

var Nightmares Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 65,
    Race: RaceDarkElf,
}

var DraconianSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 66,
    Race: RaceDraconian,
}

var DraconianSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 67,
    Race: RaceDraconian,
}

var DraconianHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 68,
    Race: RaceDraconian,
}

var DraconianBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 69,
    Race: RaceDraconian,
}

var DraconianShaman Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 70,
    Race: RaceDraconian,
}

var DraconianMagician Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 71,
    Race: RaceDraconian,
}

var DraconianEngineer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 72,
    Race: RaceDraconian,
}

var DraconianSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 73,
    Race: RaceDraconian,
}

var DoomDrake Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 74,
    Race: RaceDraconian,
}

var AirShip Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 75,
    Race: RaceDraconian,
}

var DwarfSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 76,
    Race: RaceDwarf,
}

var DwarfHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 77,
    Race: RaceDwarf,
}

var DwarfEngineer Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 78,
    Race: RaceDwarf,
}

var Hammerhands Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 79,
    Race: RaceDwarf,
}

var SteamCannon Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 80,
    Race: RaceDwarf,
}

var Golem Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 81,
    Race: RaceDwarf,
}

var DwarfSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 82,
    Race: RaceDwarf,
}

var GnollSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 83,
    Race: RaceGnoll,
}

var GnollSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 84,
    Race: RaceGnoll,
}

var GnollHalberdiers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 85,
    Race: RaceGnoll,
}

var GnollBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 86,
    Race: RaceGnoll,
}

var GnollSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 87,
    Race: RaceGnoll,
}

var WolfRiders Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 88,
    Race: RaceGnoll,
}

var HalflingSpearmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 89,
    Race: RaceHalfling,
}

var HalflingSwordsmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 90,
    Race: RaceHalfling,
}

var HalflingBowmen Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 91,
    Race: RaceHalfling,
}

var HalflingShamans Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 92,
    Race: RaceHalfling,
}

var HalflingSettlers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 93,
    Race: RaceHalfling,
}

var Slingers Unit = Unit{
    LbxFile: "units1.lbx",
    Index: 94,
    Race: RaceHalfling,
}

