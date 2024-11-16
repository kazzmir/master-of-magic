package city

import (
    _ "log"
    "math"
    "math/rand/v2"
    "image"

    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
)

type CityEvent interface {
}

type CityEventPopulationGrowth struct {
    Size int
}

type CityEventNewUnit struct {
    Unit units.Unit
    WeaponBonus data.WeaponBonus
}

type CityEventOutpostDestroyed struct {
}

type CityEventOutpostHamlet struct {
}

type CityEventNewBuilding struct {
    Building buildinglib.Building
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

type CatchmentProvider interface {
    GetCatchmentArea(x int, y int) map[image.Point]maplib.FullTile
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
    Outpost bool
    Banner data.BannerType
    Buildings *set.Set[buildinglib.Building]

    CatchmentProvider CatchmentProvider

    TaxRate fraction.Fraction

    // reset every turn, keeps track of whether the player sold a building
    SoldBuilding bool

    // how many hammers the city has produced towards the current project
    Production float32
    ProducingBuilding buildinglib.Building
    ProducingUnit units.Unit

    BuildingInfo buildinglib.BuildingInfos
}

func MakeCity(name string, x int, y int, race data.Race, banner data.BannerType, taxRate fraction.Fraction, buildingInfo buildinglib.BuildingInfos, catchmentProvider CatchmentProvider) *City {
    city := City{
        Name: name,
        X: x,
        Y: y,
        Banner: banner,
        Race: race,
        Buildings: set.MakeSet[buildinglib.Building](),
        TaxRate: taxRate,
        CatchmentProvider: catchmentProvider,
        BuildingInfo: buildingInfo,
    }

    return &city
}

func (city *City) GetX() int {
    return city.X
}

func (city *City) GetY() int {
    return city.Y
}

func (city *City) GetBanner() data.BannerType {
    return city.Banner
}

func (city *City) GetOutpostHouses() int {
    // every 100 population is 1 house
    return city.Population / 100
}

func (city *City) UpdateTaxRate(taxRate fraction.Fraction, garrison []units.StackUnit){
    city.TaxRate = taxRate
    city.UpdateUnrest(garrison)
}

func (city *City) AddBuilding(building buildinglib.Building){
    city.Buildings.Insert(building)
}

func (city *City) HasSummoningCircle() bool {
    return city.Buildings.Contains(buildinglib.BuildingSummoningCircle)
}

func (city *City) HasFortress() bool {
    return city.Buildings.Contains(buildinglib.BuildingFortress)
}

func (city *City) ProducingString() string {
    if city.ProducingBuilding != buildinglib.BuildingNone {
        return city.BuildingInfo.Name(city.ProducingBuilding)
    }

    if !city.ProducingUnit.Equals(units.UnitNone) {
        return city.ProducingUnit.Name
    }

    return ""
}

func (city *City) ProducingTurnsLeft() int {
    if city.ProducingBuilding != buildinglib.BuildingNone {
        switch city.ProducingBuilding {
            case buildinglib.BuildingHousing, buildinglib.BuildingTradeGoods: return 1
        }

        cost := city.BuildingInfo.ProductionCost(city.ProducingBuilding) - int(city.Production)
        if cost < 0 {
            cost = 0
        }

        return int(math.Ceil(float64(cost) / float64(city.WorkProductionRate())))
    }

    if !city.ProducingUnit.Equals(units.UnitNone) {
        cost := city.ProducingUnit.ProductionCost - int(city.Production)
        if cost < 0 {
            cost = 0
        }

        return int(math.Ceil(float64(cost) / float64(city.WorkProductionRate())))
    }

    return 0
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

func (city *City) ResetCitizens(garrison []units.StackUnit) {
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

func (city *City) PowerCitizens() int {
    citizenPower := float64(0)
    if city.Race == data.RaceDraconian || city.Race == data.RaceHighElf || city.Race == data.RaceBeastmen {
        citizenPower = 0.5
    } else if city.Race == data.RaceDarkElf {
        citizenPower = 1
    }
    return int(citizenPower * float64(city.Citizens()))
}

func (city *City) PowerMinerals() int {
    catchment := city.CatchmentProvider.GetCatchmentArea(city.X, city.Y)

    extra := 0
    for _, tile := range catchment {
        bonus := tile.GetBonus()
        switch bonus {
            case data.BonusMithrilOre: extra += 1
            case data.BonusAdamantiumOre: extra += 2
            case data.BonusQuorkCrystal: extra += 3
            case data.BonusCrysxCrystal: extra += 5
        }
    }

    if city.Race == data.RaceDwarf {
        extra *= 2
    }

    if city.Buildings.Contains(buildinglib.BuildingMinersGuild) {
        extra = int(float64(extra) * 1.5)
    }

    return extra
}

/* power production from buildings and citizens
 */
func (city *City) ComputePower() int {
    power := 0

    religiousPower := 0

    for _, buildingValue := range city.Buildings.Values() {
        switch buildingValue {
            case buildinglib.BuildingShrine: religiousPower += 1
            case buildinglib.BuildingTemple: religiousPower += 2
            case buildinglib.BuildingParthenon: religiousPower += 3
            case buildinglib.BuildingCathedral: religiousPower += 4
            case buildinglib.BuildingAlchemistsGuild: power += 3
            case buildinglib.BuildingWizardsGuild: power -= 3
            case buildinglib.BuildingFortress:
                if city.Plane == data.PlaneMyrror {
                    power += 5
                }
        }
    }

    // FIXME: take enchantments and bonuses for religious power into account

    return power + religiousPower + city.PowerCitizens() + city.PowerMinerals()
}

func (city *City) UpdateUnrest(garrison []units.StackUnit) {
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

func (city *City) ComputeUnrest(garrison []units.StackUnit) int {
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
        if unit.GetRace() != data.RaceFantastic {
            garrisonSupression += 1
        }
    }

    // pacification from buildings

    pacificationRetort := float64(1)
    // if has Divine Power or Infernal Power
    // pacificationRetort = 1.5

    pacification := float64(0)
    if city.Buildings.Contains(buildinglib.BuildingShrine) {
        pacification += 1 * pacificationRetort
    }

    if city.Buildings.Contains(buildinglib.BuildingTemple) {
        pacification += float64(templePacification(city.Race)) * pacificationRetort
    }

    if city.Buildings.Contains(buildinglib.BuildingParthenon) {
        pacification += float64(parthenonPacification(city.Race)) * pacificationRetort
    }

    if city.Buildings.Contains(buildinglib.BuildingCathedral) {
        pacification += float64(cathedralPaclification(city.Race)) * pacificationRetort
    }

    if city.Buildings.Contains(buildinglib.BuildingAnimistsGuild) {
        pacification += float64(animistsGuildPacification(city.Race))
    }

    if city.Buildings.Contains(buildinglib.BuildingOracle) {
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

    if city.Buildings.Contains(buildinglib.BuildingGranary) {
        bonus += 1
    }

    if city.Buildings.Contains(buildinglib.BuildingFarmersMarket) {
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

    if city.Buildings.Contains(buildinglib.BuildingGranary) {
        base += 20
    }

    if city.Buildings.Contains(buildinglib.BuildingFarmersMarket) {
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
    catchment := city.CatchmentProvider.GetCatchmentArea(city.X, city.Y)
    food := fraction.Zero()

    for _, tile := range catchment {
        switch tile.Tile.TerrainType() {
            case terrain.Ocean, terrain.Mountain, terrain.Desert, terrain.Tundra, terrain.Volcano: // nothing
            case terrain.Grass: food = food.Add(fraction.Make(3, 2))
            case terrain.Forest, terrain.Hill, terrain.Shore, terrain.Swamp: food = food.Add(fraction.Make(1, 2))
            case terrain.SorceryNode, terrain.River: food = food.Add(fraction.FromInt(2))
            case terrain.NatureNode: food = food.Add(fraction.Make(5, 2))
            case terrain.ChaosNode: // nothing
            case terrain.Lake:
                switch tile.Tile.Index(city.Plane) {
                    case terrain.IndexLake1, terrain.IndexLake2, terrain.IndexLake3, terrain.IndexLake4:
                        food = food.Add(fraction.Make(3, 2))
                }
        }

        if tile.HasWildGame() {
            food = food.Add(fraction.FromInt(2))
        }
    }

    return int(food.ToFloat())
}

func (city *City) FoodProductionRate() int {
    base := city.foodProductionRate(city.Farmers)

    // foresters guild doesn't contribute to the food needed to support the town, instead the food is added to the global surplus
    if city.Buildings.Contains(buildinglib.BuildingForestersGuild) {
        base += 2
    }

    return base
}

func (city *City) FarmerFoodProduction(farmers int) int {
    rate := 2

    switch city.Race {
        case data.RaceHalfling: rate = 3
    }

    return rate * farmers
}

func (city *City) foodProductionRate(farmers int) int {
    baseRate := float32(city.FarmerFoodProduction(farmers))

    if city.Buildings.Contains(buildinglib.BuildingForestersGuild) {
        baseRate += 2
    }

    // TODO: if famine is active then base rate is halved

    baseLevel := float32(city.BaseFoodLevel())
    if baseRate > baseLevel {
        baseRate = baseLevel + (baseLevel - baseRate) / 2
    }

    bonus := 0

    if city.Buildings.Contains(buildinglib.BuildingGranary) {
        bonus += 2
    }

    if city.Buildings.Contains(buildinglib.BuildingFarmersMarket) {
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

func (city *City) GoldTaxation() int {
    return int(float32(city.NonRebels()) * float32(city.TaxRate.ToFloat()))
}

func (city *City) GoldTradeGoods() int {
    if city.ProducingBuilding == buildinglib.BuildingTradeGoods {
        return int(city.WorkProductionRate() / 2)
    }

    return 0
}

func (city *City) GoldMinerals() int {
    catchment := city.CatchmentProvider.GetCatchmentArea(city.X, city.Y)

    extra := 0

    for _, tile := range catchment {
        bonus := tile.GetBonus()
        switch bonus {
            case data.BonusSilverOre: extra += 2
            case data.BonusGoldOre: extra += 3
            case data.BonusGem: extra += 5
        }
    }

    if city.Race == data.RaceDwarf {
        extra *= 2
    }

    if city.Buildings.Contains(buildinglib.BuildingMinersGuild) {
        extra = int(float64(extra) * 1.5)
    }

    return extra
}

func (city *City) GoldMarketplace() int {
    if city.Buildings.Contains(buildinglib.BuildingMarketplace) {
        return (city.GoldTaxation() + city.GoldMinerals()) / 2
    }

    return 0
}

func (city *City) GoldBank() int {
    if city.Buildings.Contains(buildinglib.BuildingBank) {
        return (city.GoldTaxation() + city.GoldMinerals()) / 2
    }

    return 0
}

func (city *City) GoldMerchantsGuild() int {
    if city.Buildings.Contains(buildinglib.BuildingMerchantsGuild) {
        return city.GoldTaxation() + city.GoldMinerals()
    }

    return 0
}

func (city *City) GoldSurplus() int {
    income := city.GoldTaxation()
    income += city.GoldTradeGoods()
    income += city.GoldMinerals()
    income += city.GoldMarketplace()
    income += city.GoldBank()
    income += city.GoldMerchantsGuild()

    upkeepCosts := city.ComputeUpkeep()

    out := income - upkeepCosts

    if out < 0 {
        out = 0
    }

    return out
}

func (city *City) ProductionWorkers() float32 {
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

    return float32(workerRate * city.Workers)
}

func (city *City) ProductionFarmers() float32 {
    return 0.5 * float32(city.Farmers)
}

func (city *City) productionBuildingBonus(building buildinglib.Building, percent float32) float32 {
    if city.Buildings.Contains(building) {
        return percent * (city.ProductionWorkers() + city.ProductionFarmers())
    }

    return 0
}

func (city *City) ProductionMinersGuild() float32 {
    return city.productionBuildingBonus(buildinglib.BuildingMinersGuild, 0.5)
}

func (city *City) ProductionSawmill() float32 {
    return city.productionBuildingBonus(buildinglib.BuildingSawmill, 0.25)
}

func (city *City) ProductionForestersGuild() float32 {
    return city.productionBuildingBonus(buildinglib.BuildingForestersGuild, 0.25)
}

func (city *City) ProductionMechaniciansGuild() float32 {
    return city.productionBuildingBonus(buildinglib.BuildingMechaniciansGuild, 0.5)
}

func (city *City) ProductionTerrain() float32 {
    catchment := city.CatchmentProvider.GetCatchmentArea(city.X, city.Y)
    production := float32(0)
    mineralProduction := float32(0)

    for _, tile := range catchment {
        switch tile.Tile.TerrainType() {
            case terrain.Mountain, terrain.ChaosNode: production += 0.05
            case terrain.Desert, terrain.Forest, terrain.Hill, terrain.NatureNode: production += 0.03
        }

        switch tile.GetBonus() {
            case data.BonusIronOre: mineralProduction += 0.05
            case data.BonusCoal: mineralProduction += 0.1
        }
    }

    if city.Race == data.RaceDwarf {
        mineralProduction *= 2
    }

    if city.Buildings.Contains(buildinglib.BuildingMinersGuild) {
        mineralProduction *= 2
    }

    return (production + mineralProduction) * (city.ProductionWorkers() + city.ProductionFarmers())
}

func (city *City) WorkProductionRate() float32 {
    return city.ProductionWorkers() +
           city.ProductionFarmers() +
           city.ProductionMinersGuild() +
           city.ProductionMechaniciansGuild() +
           city.ProductionTerrain() +
           city.ProductionSawmill() +
           city.ProductionForestersGuild()
}

func (city *City) GrowOutpost() CityEvent {

    growRaceBonus := float64(0.0)
    growTerrainChance := 0.0
    growSpellChance := 0.0

    switch city.Race {
        case data.RaceBarbarian: growRaceBonus = 15
        case data.RaceGnoll: growRaceBonus = 5
        case data.RaceHalfling: growRaceBonus = 15
        case data.RaceHighElf: growRaceBonus = 5
        case data.RaceHighMen: growRaceBonus = 10
        case data.RaceKlackon: growRaceBonus = 5
        case data.RaceLizard: growRaceBonus = 10
        case data.RaceNomad: growRaceBonus = 10
        case data.RaceOrc: growRaceBonus = 10
        case data.RaceBeastmen: growRaceBonus = 5
        case data.RaceDarkElf: growRaceBonus = 2
        case data.RaceDraconian: growRaceBonus = 5
        case data.RaceDwarf: growRaceBonus = 7
        case data.RaceTroll: growRaceBonus = 3
    }

    // FIXME: take terrain into account
    // FIXME: take global enchantments into account for growth

    growChance := 0.01 * float64(city.MaximumCitySize()) + growRaceBonus / 100.0 + growTerrainChance + growSpellChance

    shrinkSpellChance := 0.0

    // FIXME: take global enchantments into account for shrinkage

    shrinkChance := 0.05 + shrinkSpellChance

    for range 3 {
        if rand.Float64() < growChance {
            city.Population += 100
        }
    }

    for range 2 {
        if rand.Float64() < shrinkChance {
            city.Population -= 100
        }
    }

    if city.Population < 100 {
        return &CityEventOutpostDestroyed{}
    } else if city.Population >= 1000 {
        city.Outpost = false
        return &CityEventOutpostHamlet{}
    }

    return nil
}

// if the city contains an alchemist's guild then new units get one of the following bonuses
//  * magic weapon
//  * mythril weapon (if mythril ore in catchment area)
//  * adamantium weapon (if adamantium ore in catchment area)
func (city *City) GetWeaponBonus() data.WeaponBonus {
    hasMythril := false
    hasAdamantium := false

    if city.Buildings.Contains(buildinglib.BuildingAlchemistsGuild) {
        catchment := city.CatchmentProvider.GetCatchmentArea(city.X, city.Y)

        for _, tile := range catchment {
            if tile.GetBonus() == data.BonusMithrilOre {
                hasMythril = true
            }

            if tile.GetBonus() == data.BonusAdamantiumOre {
                hasAdamantium = true
            }
        }

        if hasAdamantium {
            return data.WeaponAdamantium
        } else if hasMythril {
            return data.WeaponMythril
        } else {
            return data.WeaponMagic
        }
    }

    return data.WeaponNone
}

// do all the stuff needed per turn
// increase population, add production, add food/money, etc
func (city *City) DoNextTurn(garrison []units.StackUnit) []CityEvent {
    var cityEvents []CityEvent
    if city.Outpost {
        event := city.GrowOutpost()
        if event != nil {
            cityEvents = append(cityEvents, event)
        }
    } else {

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
                    city.ProducingBuilding = buildinglib.BuildingHousing
                }
            } else if !city.ProducingUnit.Equals(units.UnitNone) && city.Production >= float32(city.ProducingUnit.ProductionCost) {
                cityEvents = append(cityEvents, &CityEventNewUnit{Unit: city.ProducingUnit, WeaponBonus: city.GetWeaponBonus()})
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

    }

    return cityEvents
}

func (city *City) AllowedBuildings(what buildinglib.Building) []buildinglib.Building {
    return city.BuildingInfo.Allows(what)
}

func (city *City) AllowedUnits(what buildinglib.Building) []units.Unit {
    var out []units.Unit

    for _, unit := range units.AllUnits {
        if unit.Race == data.RaceAll || unit.Race == city.Race {
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
