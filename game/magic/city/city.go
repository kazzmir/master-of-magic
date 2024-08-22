package city

import (
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

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
    BuildingWizardTower
    BuildingSummoningCircle

    BuildingHousing
    BuildingTradeGoods

    // not a real building, just a marker
    BuildingLast
)

func (building Building) String() string {
    switch building {
        case BuildingBarracks: return "Barracks"
        case BuildingArmory: return "Armory"
        case BuildingFightersGuild: return "Fighters Guild"
        case BuildingArmorersGuild: return "Armorers Guild"
        case BuildingWarCollege: return "War College"
        case BuildingSmithy: return "Smithy"
        case BuildingStables: return "Stables"
        case BuildingAnimistsGuild: return "Animists Guild"
        case BuildingFantasticStable: return "Fantastic Stable"
        case BuildingShipwrightsGuild: return "Shipwrights Guild"
        case BuildingShipYard: return "Ship Yard"
        case BuildingMaritimeGuild: return "Maritime Guild"
        case BuildingSawmill: return "Sawmill"
        case BuildingLibrary: return "Library"
        case BuildingSagesGuild: return "Sages Guild"
        case BuildingOracle: return "Oracle"
        case BuildingAlchemistsGuild: return "Alchemists Guild"
        case BuildingUniversity: return "University"
        case BuildingWizardsGuild: return "Wizards Guild"
        case BuildingShrine: return "Shrine"
        case BuildingTemple: return "Temple"
        case BuildingParthenon: return "Parthenon"
        case BuildingCathedral: return "Cathedral"
        case BuildingMarketplace: return "Marketplace"
        case BuildingBank: return "Bank"
        case BuildingMerchantsGuild: return "Merchants Guild"
        case BuildingGranary: return "Granary"
        case BuildingFarmersMarket: return "Farmers Market"
        case BuildingForestersGuild: return "Foresters Guild"
        case BuildingBuildersHall: return "Builders Hall"
        case BuildingMechaniciansGuild: return "Mechanicians Guild"
        case BuildingMinersGuild: return "Miners Guild"
        case BuildingCityWalls: return "City Walls"
        case BuildingWizardTower: return "Wizard Tower"
        case BuildingSummoningCircle: return "Summoning Circle"

        case BuildingHousing: return "Housing"
        case BuildingTradeGoods: return "Trade Goods"
    }

    return "?"
}

type CitySize int
const (
    CitySizeHamlet CitySize = iota
    CitySizeVillage
    CitySizeTown
    CitySizeCity
    CitySizeCapital
)

func (citySize CitySize) String() string {
    switch citySize {
    case CitySizeHamlet:
        return "hamlet"
    case CitySizeVillage:
        return "village"
    case CitySizeTown:
        return "town"
    case CitySizeCity:
        return "city"
    case CitySizeCapital:
        return "capital"
    }

    return "Unknown"
}

type City struct {
    Population int
    Name string
    Wall bool
    FoodProduction int
    WorkProduction int
    MoneyProduction int
    MagicProduction int
    Race data.Race
    X int
    Y int
    Buildings *set.Set[Building]

    Producing Building
}

func (city *City) AddBuilding(building Building){
    city.Buildings.Insert(building)
}

func (city *City) GetSize() CitySize {
    if city.Population < 5000 {
        return CitySizeHamlet
    }

    if city.Population < 9000 {
        return CitySizeVillage
    }

    if city.Population < 13000 {
        return CitySizeTown
    }

    if city.Population < 17000 {
        return CitySizeCity
    }

    return CitySizeCapital
}

