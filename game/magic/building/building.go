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

// FIXME: these values can come from buildat.lbx
func (building Building) ProductionCost_legacy() int {
    switch building {
        case BuildingBarracks: return 30
        case BuildingArmory: return 80
        case BuildingFightersGuild: return 200
        case BuildingArmorersGuild: return 350
        case BuildingWarCollege: return 500
        case BuildingSmithy: return 40
        case BuildingStables: return 80
        case BuildingAnimistsGuild: return 300
        case BuildingFantasticStable: return 600
        case BuildingShipwrightsGuild: return 100
        case BuildingShipYard: return 200
        case BuildingMaritimeGuild: return 400
        case BuildingSawmill: return 100
        case BuildingLibrary: return 60
        case BuildingSagesGuild: return 120
        case BuildingOracle: return 500
        case BuildingAlchemistsGuild: return 250
        case BuildingUniversity: return 300
        case BuildingWizardsGuild: return 1000
        case BuildingShrine: return 100
        case BuildingTemple: return 200
        case BuildingParthenon: return 400
        case BuildingCathedral: return 800
        case BuildingMarketplace: return 100
        case BuildingBank: return 250
        case BuildingMerchantsGuild: return 600
        case BuildingGranary: return 40
        case BuildingFarmersMarket: return 100
        case BuildingForestersGuild: return 200
        case BuildingBuildersHall: return 60
        case BuildingMechaniciansGuild: return 600
        case BuildingMinersGuild: return 300
        case BuildingCityWalls: return 150
    }

    return 0
}

func (building Building) UpkeepCost() int {
    switch building {
        case BuildingBarracks: return 0
        case BuildingArmory: return 2
        case BuildingFightersGuild: return 3
        case BuildingArmorersGuild: return 4
        case BuildingWarCollege: return 5
        case BuildingSmithy: return 1
        case BuildingStables: return 2
        case BuildingAnimistsGuild: return 5
        case BuildingFantasticStable: return 6
        case BuildingShipwrightsGuild: return 1
        case BuildingShipYard: return 2
        case BuildingMaritimeGuild: return 4
        case BuildingSawmill: return 2
        case BuildingLibrary: return 1
        case BuildingSagesGuild: return 2
        case BuildingOracle: return 4
        case BuildingAlchemistsGuild: return 3
        case BuildingUniversity: return 3
        case BuildingWizardsGuild: return 5
        case BuildingShrine: return 1
        case BuildingTemple: return 2
        case BuildingParthenon: return 3
        case BuildingCathedral: return 4
        case BuildingMarketplace: return 1
        case BuildingBank: return 3
        case BuildingMerchantsGuild: return 5
        case BuildingGranary: return 1
        case BuildingFarmersMarket: return 2
        case BuildingForestersGuild: return 2
        case BuildingBuildersHall: return 1
        case BuildingMechaniciansGuild: return 5
        case BuildingMinersGuild: return 3
        case BuildingCityWalls: return 2
    }

    return 0
}

func (building Building) String() string {
    switch building {
        case BuildingBarracks: return "Barracks"
        case BuildingArmory: return "Armory"
        case BuildingFightersGuild: return "Fighters Guild"
        case BuildingArmorersGuild: return "Armorers Guild"
        case BuildingWarCollege: return "War College"
        case BuildingSmithy: return "Smithy"
        case BuildingStables: return "Stables"
        case BuildingAnimistsGuild: return "Animist's Guild"
        case BuildingFantasticStable: return "Fantastic Stable"
        case BuildingShipwrightsGuild: return "Ship Wrights Guild"
        case BuildingShipYard: return "Shipyard"
        case BuildingMaritimeGuild: return "Maritime Guild"
        case BuildingSawmill: return "Sawmill"
        case BuildingLibrary: return "Library"
        case BuildingSagesGuild: return "Sage's Guild"
        case BuildingOracle: return "Oracle"
        case BuildingAlchemistsGuild: return "Alchemists Guild"
        case BuildingUniversity: return "University"
        case BuildingWizardsGuild: return "Wizard's Guild"
        case BuildingShrine: return "Shrine"
        case BuildingTemple: return "Temple"
        case BuildingParthenon: return "Parthenon"
        case BuildingCathedral: return "Cathedral"
        case BuildingMarketplace: return "Marketplace"
        case BuildingBank: return "Bank"
        case BuildingMerchantsGuild: return "Merchant's Guild"
        case BuildingGranary: return "Granary"
        case BuildingFarmersMarket: return "Farmer's Market"
        case BuildingForestersGuild: return "Forester's Guild"
        case BuildingBuildersHall: return "Builder's Hall"
        case BuildingMechaniciansGuild: return "Mechanician's Guild"
        case BuildingMinersGuild: return "Miner's Guild"
        case BuildingCityWalls: return "City Walls"
        case BuildingFortress: return "Fortress"
        case BuildingSummoningCircle: return "Summoning Circle"

        case BuildingHousing: return "Housing"
        case BuildingTradeGoods: return "Trade Goods"
    }

    return "?"
}

// FIXME: these values come from buildat.lbx
func GetBuildingMaintenance(building Building) int {
    switch building {
        case BuildingBarracks: return 0
        case BuildingArmory: return 2
        case BuildingFightersGuild: return 3
        case BuildingArmorersGuild: return 4
        case BuildingWarCollege: return 5
        case BuildingSmithy: return 1
        case BuildingStables: return 2
        case BuildingAnimistsGuild: return 5
        case BuildingFantasticStable: return 6
        case BuildingShipwrightsGuild: return 1
        case BuildingShipYard: return 2
        case BuildingMaritimeGuild: return 4
        case BuildingSawmill: return 2
        case BuildingLibrary: return 1
        case BuildingSagesGuild: return 2
        case BuildingOracle: return 4
        case BuildingAlchemistsGuild: return 3
        case BuildingUniversity: return 3
        case BuildingWizardsGuild: return 5 // FIXME: also requires 3 power
        case BuildingShrine: return 1
        case BuildingTemple: return 2
        case BuildingParthenon: return 3
        case BuildingCathedral: return 4
        case BuildingMarketplace: return 1
        case BuildingBank: return 3
        case BuildingMerchantsGuild: return 5
        case BuildingGranary: return 1
        case BuildingFarmersMarket: return 2
        case BuildingForestersGuild: return 2
        case BuildingBuildersHall: return 1
        case BuildingMechaniciansGuild: return 5
        case BuildingMinersGuild: return 3
        case BuildingCityWalls: return 2
    }

    return 0
}
