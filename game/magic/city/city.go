package city

import (
    "math"

    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
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
    BuildingFortress
    BuildingSummoningCircle

    BuildingHousing
    BuildingTradeGoods

    // not a real building, just a marker
    BuildingLast
)

// FIXME: these values can come from buildat.lbx
func (building Building) ProductionCost() int {
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

type CityEvent interface {
}

type CityEventPopulationGrowth struct {
    Size int
}

type CityEventNewUnit struct {
    Unit units.Unit
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

const MAX_CITY_POPULATION = 25000

type City struct {
    Population int
    Farmers int
    Workers int
    Rebels int
    Name string
    Wall bool
    Plane data.Plane
    FoodProductionRate int
    MoneyProductionRate int
    MagicProductionRate int
    Race data.Race
    X int
    Y int
    Buildings *set.Set[Building]

    // reset every turn, keeps track of whether the player sold a building
    SoldBuilding bool

    // how many hammers the city has produced towards the current project
    Production float32
    ProducingBuilding Building
    ProducingUnit units.Unit
}

func MakeCity(name string, x int, y int, race data.Race) *City {
    city := City{
        Name: name,
        X: x,
        Y: y,
        Race: race,
        Buildings: set.MakeSet[Building](),
    }

    return &city
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

func (city *City) Citizens() int {
    return city.Population / 1000
}

/* FIXME: take enchantments into account
 * https://masterofmagic.fandom.com/wiki/Farmer
 */
func (city *City) ComputeSubsistenceFarmers() int {
    // FIXME: take buildings into account (granary, farmers market, etc)
    // each citizen needs 2 food
    // round up in case of an odd number of citizens
    return int(math.Ceil(float64(city.Citizens()) / 2.0))
}

/* returns the maximum number of citizens. population is citizens * 1000
 */
func (city *City) MaximumCitySize() int {
    // FIXME: based on surrounding tiles
    foodAvailability := 20
    bonus := 0

    if city.Buildings.Contains(BuildingGranary) {
        bonus += 1
    }

    if city.Buildings.Contains(BuildingFarmersMarket) {
        bonus += 3
    }

    return foodAvailability + bonus
}

func (city *City) PopulationGrowthRate() int {
    base := 10 * (city.MaximumCitySize() - city.Citizens() + 1) / 2
    switch city.Race {
        case data.RaceBarbarian: base += 20
        case data.RaceGnoll: base -= 10
        case data.RaceHalfling:
        case data.RaceHighElf: base -= 20
        case data.RaceHighMen:
        case data.RaceKlackon: base -= 10
        case data.RaceLizard: base += 10
        case data.RaceNomad: base -= 10
        case data.RaceOrc:
        case data.RaceBeastmen:
        case data.RaceDarkElf: base -= 10
        case data.RaceDraconian: base -= 10
        case data.RaceDwarf: base -= 20
        case data.RaceTroll: base -= 20
    }

    if city.Buildings.Contains(BuildingGranary) {
        base += 20
    }

    if city.Buildings.Contains(BuildingFarmersMarket) {
        base += 30
    }

    return base
}

func (city *City) WorkProductionRate() float32 {
    workerRate := 2

    switch city.Race {
        case data.RaceBarbarian: workerRate = 2
        case data.RaceGnoll: workerRate = 2
        case data.RaceHalfling: workerRate = 2
        case data.RaceHighElf: workerRate = 2
        case data.RaceHighMen: workerRate = 2
        case data.RaceKlackon: workerRate = 3
        case data.RaceLizard: workerRate = 2
        case data.RaceNomad: workerRate = 2
        case data.RaceOrc: workerRate = 2
        case data.RaceBeastmen: workerRate = 2
        case data.RaceDarkElf: workerRate = 2
        case data.RaceDraconian: workerRate = 2
        case data.RaceDwarf: workerRate = 2
        case data.RaceTroll: workerRate = 3
    }

    return float32(workerRate * city.Workers) + 0.5 * float32(city.Farmers)
}

// do all the stuff needed per turn
// increase population, add production, add food/money, etc
func (city *City) DoNextTurn() []CityEvent {
    var cityEvents []CityEvent

    city.SoldBuilding = false

    oldPopulation := city.Population
    city.Population += city.PopulationGrowthRate()
    if city.Population > city.MaximumCitySize() * 1000 {
        city.Population = city.MaximumCitySize() * 1000
    }

    if math.Abs(float64(city.Population - oldPopulation)) >= 1000 {
        cityEvents = append(cityEvents, CityEventPopulationGrowth{Size: city.Population - oldPopulation})
    }

    if city.ProducingBuilding.ProductionCost() != 0 || !city.ProducingUnit.Equals(units.UnitNone) {
        city.Production += city.WorkProductionRate()

        if city.ProducingBuilding.ProductionCost() != 0 {
            if city.Production >= float32(city.ProducingBuilding.ProductionCost()) {
                city.Buildings.Insert(city.ProducingBuilding)
                city.Production = 0
                city.ProducingBuilding = BuildingTradeGoods
            } else if !city.ProducingUnit.Equals(units.UnitNone) && city.Production >= float32(city.ProducingUnit.ProductionCost) {
                cityEvents = append(cityEvents, CityEventNewUnit{Unit: city.ProducingUnit})
            }
        }
    }

    return cityEvents
}
