package units

// FIXME: probably move this somewhere else
type Race int

const (
    RaceLizard Race = iota
    RaceNomad
    RaceOrc
    RaceTroll
    RaceFantastic
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
