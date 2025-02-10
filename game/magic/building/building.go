package building

type Building int
const (
    BuildingNone Building = iota

    BuildingBarracks
    BuildingArmory
    BuildingFightersGuild
    BuildingArmorersGuild
    BuildingWarCollege
    BuildingSmithy
    BuildingStables
    BuildingAnimistsGuild
    BuildingFantasticStable
    BuildingShipwrightsGuild
    BuildingShipYard
    BuildingMaritimeGuild
    BuildingSawmill
    BuildingLibrary
    BuildingSagesGuild
    BuildingOracle
    BuildingAlchemistsGuild
    BuildingUniversity
    BuildingWizardsGuild
    BuildingShrine
    BuildingTemple
    BuildingParthenon
    BuildingCathedral
    BuildingMarketplace
    BuildingBank
    BuildingMerchantsGuild
    BuildingGranary
    BuildingFarmersMarket
    BuildingForestersGuild
    BuildingBuildersHall
    BuildingMechaniciansGuild
    BuildingMinersGuild
    BuildingCityWalls
    BuildingFortress
    BuildingSummoningCircle

    BuildingHousing
    BuildingTradeGoods

    // not a real building, just a marker
    BuildingLast
)

// the buildings that can be built in the order that they usually show up in the build screen
func Buildings() []Building {
    return []Building{
        BuildingHousing,
        BuildingTradeGoods,
        BuildingBarracks,
        BuildingArmory,
        BuildingFightersGuild,
        BuildingArmorersGuild,
        BuildingWarCollege,
        BuildingSmithy,
        BuildingStables,
        BuildingAnimistsGuild,
        BuildingFantasticStable,
        BuildingShipwrightsGuild,
        BuildingShipYard,
        BuildingMaritimeGuild,
        BuildingSawmill,
        BuildingLibrary,
        BuildingSagesGuild,
        BuildingOracle,
        BuildingAlchemistsGuild,
        BuildingUniversity,
        BuildingWizardsGuild,
        BuildingShrine,
        BuildingTemple,
        BuildingParthenon,
        BuildingCathedral,
        BuildingMarketplace,
        BuildingBank,
        BuildingMerchantsGuild,
        BuildingGranary,
        BuildingFarmersMarket,
        BuildingForestersGuild,
        BuildingBuildersHall,
        BuildingMechaniciansGuild,
        BuildingMinersGuild,
        BuildingCityWalls,
    }
}

// the index in cityscap.lbx for the picture of this building
func (building Building) Index() int {
    switch building {
        case BuildingBarracks: return 45
        case BuildingArmory: return 46
        case BuildingFightersGuild: return 47
        case BuildingArmorersGuild: return 48
        case BuildingWarCollege: return 49
        case BuildingSmithy: return 50
        case BuildingStables: return 51
        case BuildingAnimistsGuild: return 52
        case BuildingFantasticStable: return 53
        case BuildingShipwrightsGuild: return 54
        case BuildingShipYard: return 55
        case BuildingMaritimeGuild: return 56
        case BuildingSawmill: return 57
        case BuildingLibrary: return 58
        case BuildingSagesGuild: return 59
        case BuildingOracle: return 60
        case BuildingAlchemistsGuild: return 61
        case BuildingUniversity: return 62
        case BuildingWizardsGuild: return 63
        case BuildingShrine: return 64
        case BuildingTemple: return 65
        case BuildingParthenon: return 66
        case BuildingCathedral: return 67
        case BuildingMarketplace: return 68
        case BuildingBank: return 69
        case BuildingMerchantsGuild: return 70
        case BuildingGranary: return 71
        case BuildingFarmersMarket: return 72
        case BuildingBuildersHall: return 73
        case BuildingMechaniciansGuild: return 74
        case BuildingMinersGuild: return 75
        case BuildingCityWalls: return 76
        case BuildingForestersGuild: return 78
        case BuildingFortress: return 40
        case BuildingSummoningCircle: return 6
    }

    return -1
}

// the building which is shown in the city scape instead
func (building Building) ReplacedBy() Building {
    switch building {
        case BuildingBarracks: return BuildingArmory
        case BuildingFightersGuild: return BuildingArmorersGuild
        case BuildingArmorersGuild: return BuildingWarCollege
        case BuildingStables: return BuildingFantasticStable
        case BuildingLibrary: return BuildingUniversity
        case BuildingAlchemistsGuild: return BuildingWizardsGuild
        case BuildingShrine: return BuildingTemple
        case BuildingTemple: return BuildingParthenon
        case BuildingParthenon: return BuildingCathedral
        case BuildingMarketplace: return BuildingBank
        case BuildingBank: return BuildingMerchantsGuild
        case BuildingGranary: return BuildingFarmersMarket
        case BuildingShipwrightsGuild: return BuildingShipYard
        case BuildingShipYard: return BuildingMaritimeGuild
    }

    return BuildingNone
}

// the size of the picture for this building (in squares)
func (building Building) Size() (int, int) {
    switch building {
        case BuildingBarracks: return 2, 3
        case BuildingArmory: return 2, 2
        case BuildingFightersGuild: return 3, 2
        case BuildingArmorersGuild: return 4, 2
        case BuildingWarCollege: return 3, 2
        case BuildingSmithy: return 2, 2
        case BuildingStables: return 3, 3
        case BuildingFantasticStable: return 3, 3
        case BuildingAnimistsGuild: return 2, 2
        case BuildingSawmill: return 2, 2
        case BuildingLibrary: return 3, 2
        case BuildingUniversity: return 3, 2
        case BuildingSagesGuild: return 2, 2
        case BuildingOracle: return 2, 2
        case BuildingAlchemistsGuild: return 1, 1
        case BuildingWizardsGuild: return 2, 2
        case BuildingShrine: return 2, 2
        case BuildingTemple: return 3, 3
        case BuildingParthenon: return 3, 3
        case BuildingCathedral: return 3, 3
        case BuildingMarketplace: return 2, 2
        case BuildingBank: return 2, 2
        case BuildingMerchantsGuild: return 2, 2
        case BuildingGranary: return 2, 2
        case BuildingFarmersMarket: return 2, 2
        case BuildingForestersGuild: return 2, 2
        case BuildingBuildersHall: return 2, 2
        case BuildingMechaniciansGuild: return 2, 2
        case BuildingMinersGuild: return 2, 1
        case BuildingFortress: return 3, 3
        case BuildingSummoningCircle: return 3, 2
    }

    return 0, 0
}