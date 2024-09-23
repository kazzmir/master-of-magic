package city

import (
    "math"

    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/building"
)

type CityEvent interface {
}

type CityEventPopulationGrowth struct {
    Size int
}

type CityEventNewUnit struct {
    Unit units.Unit
}

type CityEventNewBuilding struct {
    Building building.Building
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
        return "Hamlet"
    case CitySizeVillage:
        return "Village"
    case CitySizeTown:
        return "Town"
    case CitySizeCity:
        return "City"
    case CitySizeCapital:
        return "Capital"
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
    Race data.Race
    X int
    Y int
    // the turn this city was created on
    BirthTurn uint64
    Banner data.BannerType
    Buildings *set.Set[building.Building]

    TaxRate fraction.Fraction

    // reset every turn, keeps track of whether the player sold a building
    SoldBuilding bool

    // how many hammers the city has produced towards the current project
    Production float32
    ProducingBuilding building.Building
    ProducingUnit units.Unit

    BuildingInfo building.BuildingInfos
}

func MakeCity(name string, x int, y int, birth uint64, race data.Race, taxRate fraction.Fraction, buildingInfo building.BuildingInfos) *City {
    city := City{
        Name: name,
        X: x,
        Y: y,
        BirthTurn: birth,
        Race: race,
        Buildings: set.MakeSet[building.Building](),
        TaxRate: taxRate,
        BuildingInfo: buildingInfo,
    }

    return &city
}

func (city *City) UpdateTaxRate(taxRate fraction.Fraction, garrison []*units.OverworldUnit){
    city.TaxRate = taxRate
    city.UpdateUnrest(garrison)
}

func (city *City) AddBuilding(building building.Building){
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

func (city *City) ResetCitizens(garrison []*units.OverworldUnit) {
    // try to leave farmers alone, but adjust them if necessary
    minimumFarmers := city.ComputeSubsistenceFarmers()
    if city.Farmers < minimumFarmers {
        city.Farmers = minimumFarmers
    }
    if city.Farmers > city.Citizens() {
        city.Farmers = city.Citizens()
    }
    city.Workers = city.Citizens() - city.Farmers
    city.Rebels = 0
    city.UpdateUnrest(garrison)
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

func (city *City) UpdateUnrest(garrison []*units.OverworldUnit) {
    rebels := city.ComputeUnrest(garrison)

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

func (city *City) ComputeUnrest(garrison []*units.OverworldUnit) int {
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
    for _, unit := range garrison {
        if unit.Unit.Race != data.RaceFantastic {
            garrisonSupression += 1
        }
    }

    // pacification from buildings

    pacificationRetort := float64(1)
    // if has Divine Power or Infernal Power
    // pacificationRetort = 1.5

    pacification := float64(0)
    if city.Buildings.Contains(building.BuildingShrine) {
        pacification += 1 * pacificationRetort
    }

    if city.Buildings.Contains(building.BuildingTemple) {
        pacification += float64(templePacification(city.Race)) * pacificationRetort
    }

    if city.Buildings.Contains(building.BuildingParthenon) {
        pacification += float64(parthenonPacification(city.Race)) * pacificationRetort
    }

    if city.Buildings.Contains(building.BuildingCathedral) {
        pacification += float64(cathedralPaclification(city.Race)) * pacificationRetort
    }

    if city.Buildings.Contains(building.BuildingAnimistsGuild) {
        pacification += float64(animistsGuildPacification(city.Race))
    }

    if city.Buildings.Contains(building.BuildingOracle) {
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

    if city.Buildings.Contains(building.BuildingGranary) {
        bonus += 1
    }

    if city.Buildings.Contains(building.BuildingFarmersMarket) {
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

    if city.Buildings.Contains(building.BuildingGranary) {
        base += 20
    }

    if city.Buildings.Contains(building.BuildingFarmersMarket) {
        base += 30
    }

    return base
}

func (city *City) ResearchProduction() int {
    research := 0

    for _, building := range city.Buildings.Values() {
        research += city.BuildingInfo.ResearchProduction(building)
    }

    return research
}

func (city *City) ManaCost() int {
    mana := 0

    for _, building := range city.Buildings.Values() {
        mana += city.BuildingInfo.ManaCost(building)
    }

    return mana

}

func (city *City) ManaSurplus() int {
    return city.ManaProduction() - city.ManaCost()
}

func (city *City) ManaProduction() int {
    mana := 0

    for _, building := range city.Buildings.Values() {
        mana += city.BuildingInfo.ManaProduction(building)
    }

    return mana
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
    return 20
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

    if city.Buildings.Contains(building.BuildingForestersGuild) {
        baseRate += 2
    }

    // TODO: if famine is active then base rate is halved

    baseLevel := float32(city.BaseFoodLevel())
    if baseRate > baseLevel {
        baseRate = baseLevel + (baseLevel - baseRate) / 2
    }

    bonus := 0

    if city.Buildings.Contains(building.BuildingGranary) {
        bonus += 2
    }

    if city.Buildings.Contains(building.BuildingFarmersMarket) {
        bonus += 3
    }

    // TODO: add 2 for each wild game tile in the catchment area

    return int(baseRate) + bonus
}

func (city *City) ComputeUpkeep() int {
    costs := 0

    for _, building := range city.Buildings.Values() {
        costs += city.BuildingInfo.UpkeepCost(building)
    }

    return costs
}

func (city *City) GoldSurplus() int {
    citizenIncome := float32(city.NonRebels()) * float32(city.TaxRate.ToFloat())

    bonus := float32(0)

    if city.ProducingBuilding == building.BuildingTradeGoods {
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
func (city *City) DoNextTurn(garrison []*units.OverworldUnit) []CityEvent {
    var cityEvents []CityEvent

    city.SoldBuilding = false

    oldPopulation := city.Population
    city.Population += city.PopulationGrowthRate()
    if city.Population > city.MaximumCitySize() * 1000 {
        city.Population = city.MaximumCitySize() * 1000
    }

    if math.Abs(float64(city.Population/1000 - oldPopulation/1000)) > 0 {
        cityEvents = append(cityEvents, &CityEventPopulationGrowth{Size: (city.Population - oldPopulation)/1000})
    }

    buildingCost := city.BuildingInfo.ProductionCost(city.ProducingBuilding)

    if buildingCost != 0 || !city.ProducingUnit.Equals(units.UnitNone) {
        city.Production += city.WorkProductionRate()

        if buildingCost != 0 {
            if city.Production >= float32(buildingCost) {
                city.Buildings.Insert(city.ProducingBuilding)
                cityEvents = append(cityEvents, &CityEventNewBuilding{Building: city.ProducingBuilding})

                city.Production = 0
                city.ProducingBuilding = building.BuildingHousing
            }
        } else if !city.ProducingUnit.Equals(units.UnitNone) && city.Production >= float32(city.ProducingUnit.ProductionCost) {
            cityEvents = append(cityEvents, &CityEventNewUnit{Unit: city.ProducingUnit})
            city.Production = 0

            if city.ProducingUnit.IsSettlers() {
                city.Population -= 1000
            }
        }
    }

    /*
    if city.Farmers < city.ComputeSubsistenceFarmers() {
        city.Farmers = city.ComputeSubsistenceFarmers()
        city.Workers = city.Citizens() - city.Rebels
    }
    */

    city.ResetCitizens(garrison)

    return cityEvents
}

func (city *City) AllowedBuildings(what building.Building) []building.Building {
    return city.BuildingInfo.Allows(what)
}

func (city *City) AllowedUnits(what building.Building) []units.Unit {
    var out []units.Unit

    for _, unit := range units.AllUnits {
        if unit.Race == data.RaceNone || unit.Race == city.Race {
            canBuild := false

            for _, required := range unit.RequiredBuildings {
                if required == what {
                    canBuild = true
                }
            }

            if canBuild {
                out = append(out, unit)
            }
        }
    }

    return out
}
