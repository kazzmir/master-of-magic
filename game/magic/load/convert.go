package load


import (
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

func ConvertTerrainSpecial(terrainSpecial uint8) data.BonusType {
    switch (terrainSpecial) {
        case 1: return data.BonusIronOre
        case 2: return data.BonusCoal
        case 3: return data.BonusSilverOre
        case 4: return data.BonusGoldOre
        case 5: return data.BonusGem
        case 6: return data.BonusMithrilOre
        case 7: return data.BonusAdamantiumOre
        case 8: return data.BonusQuorkCrystal
        case 9: return data.BonusCrysxCrystal
        case 64: return data.BonusWildGame
        case 128: return data.BonusNightshade
    }

    return data.BonusNone
}

