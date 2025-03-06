package load


import (
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

func GetTerrainSpecialLookupTable() *map[uint8]data.BonusType {
    return &map[uint8]data.BonusType{
        0: data.BonusNone,
        1: data.BonusIronOre,
        2: data.BonusCoal,
        3: data.BonusSilverOre,
        4: data.BonusGoldOre,
        5: data.BonusGem,
        6: data.BonusMithrilOre,
        7: data.BonusAdamantiumOre,
        8: data.BonusQuorkCrystal,
        9: data.BonusCrysxCrystal,
        64: data.BonusWildGame,
        128: data.BonusNightshade,
    }
}

