package city

import (
    "math"

    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/lib/fraction"
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

const MAX_CITY_CITIZENS = 25

type City struct {
    Population int
    Farmers int
    Workers int
    Rebels int
    Name string
    Wall bool
    Plane data.Plane
    MagicProductionRate int
    Race data.Race
    X int
    Y int
    Buildings *set.Set[Building]

    Garrison []units.Unit

    TaxRate fraction.Fraction

    // reset every turn, keeps track of whether the player sold a building
    SoldBuilding bool

    // how many hammers the city has produced towards the current project
    Production float32
    ProducingBuilding Building
    ProducingUnit units.Unit
}

func MakeCity(name string, x int, y int, race data.Race, taxRate fraction.Fraction) *City {
    city := City{
        Name: name,
        X: x,
        Y: y,
        Race: race,
        Buildings: set.MakeSet[Building](),
        TaxRate: taxRate,
    }

    return &city
}

func (city *City) AddGarrisonUnit(unit units.Unit){
    city.Garrison = append(city.Garrison, unit)
}

func (city *City) RemoveGarrisonUnit(toRemove units.Unit){
    var out []units.Unit

    found := false
    for _, unit := range city.Garrison {
        if !found && unit.Equals(toRemove) {
            found = true
        } else {
            out = append(out, unit)
        }
    }

    city.Garrison = out
}

func (city *City) UpdateTaxRate(taxRate fraction.Fraction){
    city.TaxRate = taxRate
    city.UpdateUnrest()
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

func (city *City) NonRebels() int {
    return city.Citizens() - city.Rebels
}

func (city *City) ResetCitizens() {
    city.Farmers = city.ComputeSubsistenceFarmers()
    city.Workers = city.Citizens() - city.Rebels - city.Farmers
    city.UpdateUnrest()
}

/* FIXME: take enchantments into account
 * https://masterofmagic.fandom.com/wiki/Farmer
 */
func (city *City) ComputeSubsistenceFarmers() int {
    // each citizen needs 1 unit of food
    requiredFood := city.Citizens()

    maxFarmers := city.Citizens() - city.Rebels

    // compute how many farmers are needed to produce the given amount of food
    for i := 0; i < maxFarmers; i++ {
        food := city.foodProductionRate(i)
        if food >= requiredFood {
            return i
        }
    }

    return maxFarmers
}

func (city *City) UpdateUnrest() {
    rebels := city.ComputeUnrest()

    if rebels > city.Rebels {
        for i := city.Rebels; i < rebels && city.Workers > 0; i++ {
            city.Workers -= 1
            city.Rebels += 1
        }

        minimumFarmers := city.ComputeSubsistenceFarmers()
        for i := city.Rebels; i < rebels && city.Farmers > minimumFarmers; i++ {
            city.Farmers -= 1
            city.Rebels += 1
        }
    } else if rebels < city.Rebels {
        // turn rebels into workers
        city.Workers += city.Rebels - rebels
        city.Rebels = rebels
    }
}

func templePacification(race data.Race) int {
    switch race {
        case data.RaceKlackon: return 0
        default: return 1
    }
}

func parthenonPacification(race data.Race) int {
    switch race {
        case data.RaceGnoll, data.RaceHighElf, data.RaceKlackon, data.RaceLizard, data.RaceDwarf: return 0
        default: return 1
    }
}

func cathedralPaclification(race data.Race) int {
    switch race {
        case data.RaceBarbarian, data.RaceGnoll, data.RaceHighElf,
             data.RaceKlackon, data.RaceLizard, data.RaceDarkElf,
             data.RaceDwarf: return 0
        default: return 1
    }
}

func animistsGuildPacification(race data.Race) int {
    switch race {
        case data.RaceBarbarian, data.RaceGnoll, data.RaceHalfling,
             data.RaceKlackon, data.RaceLizard, data.RaceDwarf: return 0
        default: return 1
    }
}

func oraclePacification(race data.Race) int {
    switch race {
        case data.RaceBarbarian, data.RaceGnoll, data.RaceHalfling,
             data.RaceHighElf, data.RaceKlackon, data.RaceLizard,
             data.RaceDwarf, data.RaceTroll: return 0
        default: return 2
    }
}

func (city *City) ComputeUnrest() int {
    unrestPercent := float64(0)

    // unrest percent from taxes
    if city.TaxRate.Equals(fraction.Zero()) {
        unrestPercent = 0
    } else if city.TaxRate.Equals(fraction.Make(1,2)) {
        unrestPercent = 0.1
    } else if city.TaxRate.Equals(fraction.Make(1, 1)) {
        unrestPercent = 0.2
    } else if city.TaxRate.Equals(fraction.Make(3, 2)) {
        unrestPercent = 0.3
    } else if city.TaxRate.Equals(fraction.Make(2, 1)) {
        unrestPercent = 0.45
    } else if city.TaxRate.Equals(fraction.Make(5, 2)) {
        unrestPercent = 0.60
    } else if city.TaxRate.Equals(fraction.Make(3, 1)) {
        unrestPercent = 0.75
    }

    // capital race vs town race modifier
    // unrest from spells
    // supression from units
    garrisonSupression := float64(0)
    for _, unit := range city.Garrison {
        if unit.Race != data.RaceFantastic {
            garrisonSupression += 1
        }
    }

    // pacification from buildings

    pacificationRetort := float64(1)
    // if has Divine Power or Infernal Power
    // pacificationRetort = 1.5

    pacification := float64(0)
    if city.Buildings.Contains(BuildingShrine) {
        pacification += 1 * pacificationRetort
    }

    if city.Buildings.Contains(BuildingTemple) {
        pacification += float64(templePacification(city.Race)) * pacificationRetort
    }

    if city.Buildings.Contains(BuildingParthenon) {
        pacification += float64(parthenonPacification(city.Race)) * pacificationRetort
    }

    if city.Buildings.Contains(BuildingCathedral) {
        pacification += float64(cathedralPaclification(city.Race)) * pacificationRetort
    }

    if city.Buildings.Contains(BuildingAnimistsGuild) {
        pacification += float64(animistsGuildPacification(city.Race))
    }

    if city.Buildings.Contains(BuildingOracle) {
        pacification += float64(oraclePacification(city.Race))
    }

    total := unrestPercent * float64(city.Citizens()) - pacification - garrisonSupression / 2

    return int(math.Max(0, total))
}

/* returns the maximum number of citizens. population is citizens * 1000
 */
func (city *City) MaximumCitySize() int {
    foodAvailability := city.BaseFoodLevel()

    // TODO: 1/2 if famine is active

    bonus := 0

    if city.Buildings.Contains(BuildingGranary) {
        bonus += 1
    }

    if city.Buildings.Contains(BuildingFarmersMarket) {
        bonus += 3
    }

    // TODO: add 2 for each wild game in the city's catchment area

    return int(math.Min(MAX_CITY_CITIZENS, float64(foodAvailability + bonus)))
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

/* amount of food needed to feed the citizens
 */
func (city *City) RequiredFood() int {
    return city.Citizens()
}

func (city *City) SurplusFood() int {
    return city.FoodProductionRate() - city.RequiredFood()
}

/* compute amount of available food on tiles in catchment area
 */
func (city *City) BaseFoodLevel() int {
    // TODO
    return 10
}

func (city *City) FoodProductionRate() int {
    return city.foodProductionRate(city.Farmers)
}

func (city *City) foodProductionRate(farmers int) int {
    rate := 2

    switch city.Race {
        case data.RaceHalfling: rate = 3
    }

    baseRate := float32(rate * farmers)

    if city.Buildings.Contains(BuildingForestersGuild) {
        baseRate += 2
    }

    // TODO: if famine is active then base rate is halved

    baseLevel := float32(city.BaseFoodLevel())
    if baseRate > baseLevel {
        baseRate = baseLevel + (baseLevel - baseRate) / 2
    }

    bonus := 0

    if city.Buildings.Contains(BuildingGranary) {
        bonus += 2
    }

    if city.Buildings.Contains(BuildingFarmersMarket) {
        bonus += 3
    }

    // TODO: add 2 for each wild game tile in the catchment area

    return int(baseRate) + bonus
}

func (city *City) ComputeUpkeep() int {
    costs := 0

    for _, building := range city.Buildings.Values() {
        costs += building.UpkeepCost()
    }

    return costs
}

func (city *City) MoneyProductionRate() int {
    citizenIncome := float32(city.NonRebels()) * float32(city.TaxRate.ToFloat())

    bonus := float32(0)

    if city.ProducingBuilding == BuildingTradeGoods {
        bonus = city.WorkProductionRate() / 2
    }

    upkeepCosts := city.ComputeUpkeep()

    return int(citizenIncome + bonus) - upkeepCosts
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
        cityEvents = append(cityEvents, CityEventPopulationGrowth{Size: (city.Population - oldPopulation)/1000})
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

    if city.Farmers < city.ComputeSubsistenceFarmers() {
        city.Farmers = city.ComputeSubsistenceFarmers()
        city.Workers = city.Citizens() - city.Rebels
    }

    return cityEvents
}
