package city

import (
    _ "log"
    "bytes"
    "fmt"
    "math"
    "math/rand/v2"
    "image"

    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
)

type CityEvent interface {
}

type CityEventPopulationGrowth struct {
    Size int
    Grow bool
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
    OnShore(x int, y int) bool
}

type ConnectedCityProvider interface {
    FindRoadConnectedCities(city *City) []*City
}

const MAX_CITY_CITIZENS = 25

type Enchantment struct {
    Enchantment data.CityEnchantment
    // this keeps track of the owning wizard by implicitly associating a banner to the wizard that cast it
    Owner data.BannerType
}

type City struct {
    Population int
    Farmers int
    Workers int
    Rebels int
    Name string
    Wall bool
    Plane data.Plane
    // the race of the towns people
    Race data.Race
    // the race of the controlling wizard
    RulingRace data.Race
    X int
    Y int
    Outpost bool
    Banner data.BannerType
    Buildings *set.Set[buildinglib.Building]

    Enchantments *set.Set[Enchantment]

    CatchmentProvider CatchmentProvider
    ConnectedProvider ConnectedCityProvider

    TaxRate fraction.Fraction

    // reset every turn, keeps track of whether the player sold a building
    SoldBuilding bool

    // how many hammers the city has produced towards the current project
    Production float32
    ProducingBuilding buildinglib.Building
    ProducingUnit units.Unit

    BuildingInfo buildinglib.BuildingInfos
}

// FIXME: Add plane?
func MakeCity(name string, x int, y int, race data.Race, banner data.BannerType, taxRate fraction.Fraction, buildingInfo buildinglib.BuildingInfos, catchmentProvider CatchmentProvider, connectedProvider ConnectedCityProvider) *City {
    city := City{
        Name: name,
        X: x,
        Y: y,
        Banner: banner,
        Race: race,
        RulingRace: race,
        Buildings: set.MakeSet[buildinglib.Building](),
        Enchantments: set.MakeSet[Enchantment](),
        TaxRate: taxRate,
        CatchmentProvider: catchmentProvider,
        ConnectedProvider: connectedProvider,
        BuildingInfo: buildingInfo,
    }

    return &city
}

func (city *City) String() string {
    return fmt.Sprintf("%v of %v", city.GetSize(), city.Name)
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

/* returns the set of buildings that could possibly be built by this city, taking terrain dependencies into account
 */
func (city *City) GetBuildableBuildings() *set.Set[buildinglib.Building] {
    // add all buildings at first
    out := set.NewSet[buildinglib.Building](
        buildinglib.BuildingBarracks, buildinglib.BuildingArmory, buildinglib.BuildingFightersGuild,
        buildinglib.BuildingArmorersGuild, buildinglib.BuildingWarCollege, buildinglib.BuildingSmithy,
        buildinglib.BuildingStables, buildinglib.BuildingAnimistsGuild, buildinglib.BuildingFantasticStable,
        buildinglib.BuildingShipwrightsGuild, buildinglib.BuildingShipYard, buildinglib.BuildingMaritimeGuild,
        buildinglib.BuildingSawmill, buildinglib.BuildingLibrary, buildinglib.BuildingSagesGuild,
        buildinglib.BuildingOracle, buildinglib.BuildingAlchemistsGuild, buildinglib.BuildingUniversity,
        buildinglib.BuildingWizardsGuild, buildinglib.BuildingShrine, buildinglib.BuildingTemple,
        buildinglib.BuildingParthenon, buildinglib.BuildingCathedral, buildinglib.BuildingMarketplace,
        buildinglib.BuildingBank, buildinglib.BuildingMerchantsGuild, buildinglib.BuildingGranary,
        buildinglib.BuildingFarmersMarket, buildinglib.BuildingForestersGuild, buildinglib.BuildingBuildersHall,
        buildinglib.BuildingMechaniciansGuild, buildinglib.BuildingMinersGuild, buildinglib.BuildingCityWalls,
    )

    // remove all buildings that depend on being near a shore
    if !city.OnShore() {
        out.RemoveMany(buildinglib.BuildingShipYard, buildinglib.BuildingShipwrightsGuild, buildinglib.BuildingMaritimeGuild, buildinglib.BuildingMaritimeGuild)
    }

    hasForest := false
    minersGuildOk := false

    for _, tile := range city.CatchmentProvider.GetCatchmentArea(city.X, city.Y) {
        switch tile.Tile.TerrainType() {
            case terrain.Forest, terrain.NatureNode: hasForest = true
            case terrain.Mountain, terrain.Volcano, terrain.Hill, terrain.ChaosNode:
                minersGuildOk = true
        }
    }

    if !minersGuildOk {
        out.RemoveMany(buildinglib.BuildingMinersGuild)
    }

    if !hasForest {
        out.RemoveMany(buildinglib.BuildingSawmill)
    }

    switch city.Race {
        case data.RaceLizard:
            out.RemoveMany(
                buildinglib.BuildingAnimistsGuild, buildinglib.BuildingUniversity, buildinglib.BuildingFantasticStable,
                buildinglib.BuildingMechaniciansGuild, buildinglib.BuildingWizardsGuild, buildinglib.BuildingMaritimeGuild,
                buildinglib.BuildingOracle, buildinglib.BuildingWarCollege, buildinglib.BuildingBank, buildinglib.BuildingMerchantsGuild,
                buildinglib.BuildingShipYard, buildinglib.BuildingAlchemistsGuild, buildinglib.BuildingShipwrightsGuild,
                buildinglib.BuildingCathedral, buildinglib.BuildingParthenon, buildinglib.BuildingSagesGuild,
                buildinglib.BuildingSawmill, buildinglib.BuildingForestersGuild, buildinglib.BuildingMinersGuild,
            )
        case data.RaceNomad:
            out.RemoveMany(buildinglib.BuildingWizardsGuild, buildinglib.BuildingMaritimeGuild)

        case data.RaceOrc:

        case data.RaceTroll:
            out.RemoveMany(
                buildinglib.BuildingAlchemistsGuild, buildinglib.BuildingUniversity, buildinglib.BuildingFantasticStable, buildinglib.BuildingMechaniciansGuild,
                buildinglib.BuildingWizardsGuild, buildinglib.BuildingMaritimeGuild, buildinglib.BuildingOracle, buildinglib.BuildingWarCollege,
                buildinglib.BuildingBank, buildinglib.BuildingMerchantsGuild, buildinglib.BuildingShipYard, buildinglib.BuildingSagesGuild,
                buildinglib.BuildingMinersGuild,
            )

        case data.RaceBarbarian:
            out.RemoveMany(
                buildinglib.BuildingAnimistsGuild, buildinglib.BuildingUniversity, buildinglib.BuildingFantasticStable, buildinglib.BuildingMechaniciansGuild,
                buildinglib.BuildingWizardsGuild, buildinglib.BuildingCathedral, buildinglib.BuildingOracle, buildinglib.BuildingWarCollege,
                buildinglib.BuildingBank, buildinglib.BuildingMerchantsGuild,
            )

        case data.RaceBeastmen:
            out.RemoveMany(buildinglib.BuildingFantasticStable, buildinglib.BuildingMerchantsGuild, buildinglib.BuildingShipYard, buildinglib.BuildingMaritimeGuild)

        case data.RaceDarkElf:
            out.RemoveMany(buildinglib.BuildingCathedral, buildinglib.BuildingMaritimeGuild)

        case data.RaceDraconian:
            out.RemoveMany(buildinglib.BuildingMechaniciansGuild, buildinglib.BuildingMaritimeGuild, buildinglib.BuildingFantasticStable)

        case data.RaceDwarf:
            out.RemoveMany(
                buildinglib.BuildingAnimistsGuild, buildinglib.BuildingUniversity, buildinglib.BuildingFantasticStable, buildinglib.BuildingMechaniciansGuild,
                buildinglib.BuildingWizardsGuild, buildinglib.BuildingMaritimeGuild, buildinglib.BuildingOracle, buildinglib.BuildingWarCollege,
                buildinglib.BuildingBank, buildinglib.BuildingMerchantsGuild, buildinglib.BuildingShipYard, buildinglib.BuildingStables,
                buildinglib.BuildingParthenon, buildinglib.BuildingCathedral,
            )

        case data.RaceGnoll:
            out.RemoveMany(
                buildinglib.BuildingMaritimeGuild, buildinglib.BuildingArmorersGuild, buildinglib.BuildingSagesGuild, buildinglib.BuildingAnimistsGuild,
                buildinglib.BuildingUniversity, buildinglib.BuildingFantasticStable, buildinglib.BuildingParthenon, buildinglib.BuildingAlchemistsGuild,
                buildinglib.BuildingCathedral, buildinglib.BuildingOracle, buildinglib.BuildingWarCollege, buildinglib.BuildingBank,
                buildinglib.BuildingMerchantsGuild, buildinglib.BuildingMechaniciansGuild, buildinglib.BuildingWizardsGuild,
            )

        case data.RaceHalfling:
            out.RemoveMany(
                buildinglib.BuildingAnimistsGuild, buildinglib.BuildingUniversity, buildinglib.BuildingFantasticStable, buildinglib.BuildingMechaniciansGuild,
                buildinglib.BuildingWizardsGuild, buildinglib.BuildingMaritimeGuild, buildinglib.BuildingOracle, buildinglib.BuildingWarCollege,
                buildinglib.BuildingBank, buildinglib.BuildingMerchantsGuild, buildinglib.BuildingShipYard, buildinglib.BuildingArmorersGuild,
                buildinglib.BuildingStables,
            )

        case data.RaceHighElf:
            out.RemoveMany(
                buildinglib.BuildingParthenon, buildinglib.BuildingMaritimeGuild, buildinglib.BuildingOracle, buildinglib.BuildingCathedral,
            )

        case data.RaceHighMen:
            out.RemoveMany(buildinglib.BuildingFantasticStable)

        case data.RaceKlackon:
            out.RemoveMany(
                buildinglib.BuildingAnimistsGuild, buildinglib.BuildingUniversity, buildinglib.BuildingFantasticStable, buildinglib.BuildingMechaniciansGuild,
                buildinglib.BuildingWizardsGuild, buildinglib.BuildingMaritimeGuild, buildinglib.BuildingOracle, buildinglib.BuildingWarCollege,
                buildinglib.BuildingBank, buildinglib.BuildingMerchantsGuild, buildinglib.BuildingShipYard, buildinglib.BuildingAlchemistsGuild,
                buildinglib.BuildingTemple, buildinglib.BuildingCathedral, buildinglib.BuildingParthenon, buildinglib.BuildingSagesGuild,
            )
    }

    return out
}

/* return the buildings that can be built, based on what the city already has and what dependencies are met
 */
func (city *City) ComputePossibleBuildings() *set.Set[buildinglib.Building] {
    possibleBuildings := set.NewSet[buildinglib.Building]()

    allowedBuildings := city.GetBuildableBuildings()

    for _, building := range buildinglib.Buildings() {
        if city.Buildings.Contains(building) {
            continue
        }

        if !allowedBuildings.Contains(building) {
            continue
        }

        canBuild := true
        for _, dependency := range city.BuildingInfo.Dependencies(building) {
            if !city.Buildings.Contains(dependency) {
                canBuild = false
                break
            }
        }

        if canBuild {
            possibleBuildings.Insert(building)
        }
    }

    possibleBuildings.Insert(buildinglib.BuildingTradeGoods)
    possibleBuildings.Insert(buildinglib.BuildingHousing)

    return possibleBuildings
}

func (city *City) ComputePossibleUnits() []units.Unit {
    var out []units.Unit
    for _, unit := range units.AllUnits {
        if unit.Race == data.RaceAll || unit.Race == city.Race {

            canBuild := true
            for _, building := range unit.RequiredBuildings {
                if !city.Buildings.Contains(building) {
                    canBuild = false
                }
            }

            if canBuild {
                out = append(out, unit)
            }
        }
    }

    return out
}

// pass in true to compute fame for capturing, and false to compute fame for razing/losing
func (city *City) FameForCaptureOrRaze(captured bool) int {
    if city.Outpost {
        return 0
    }

    switch city.GetSize() {
        case CitySizeHamlet:
            if captured {
                return 0
            } else {
                return -1
            }
        case CitySizeVillage:
            if captured {
                return 0
            } else {
                return -2
            }
        case CitySizeTown:
            if captured {
                return 1
            } else {
                return -2
            }
        case CitySizeCity:
            if captured {
                return 2
            } else {
                return -4
            }
        case CitySizeCapital:
            if captured {
                return 3
            } else {
                return -5
            }
    }

    return 0
}

// true if the city is adjacent to a water tile
func (city *City) OnShore() bool {
    return city.CatchmentProvider.OnShore(city.X, city.Y)
}

func (city *City) AddEnchantment(enchantment data.CityEnchantment, owner data.BannerType) {
    city.Enchantments.Insert(Enchantment{
        Enchantment: enchantment,
        Owner: owner,
    })
}

func (city *City) RemoveEnchantment(enchantment data.CityEnchantment, owner data.BannerType) {
    city.Enchantments.Remove(Enchantment{
        Enchantment: enchantment,
        Owner: owner,
    })
}

func (city *City) HasEnchantment(check data.CityEnchantment) bool {
    for _, enchantment := range city.Enchantments.Values() {
        if enchantment.Enchantment == check {
            return true
        }
    }

    return false
}

func (city *City) GetEnchantmentsCastBy(banner data.BannerType) []Enchantment {
    var enchantments []Enchantment

    for _, enchantment := range city.Enchantments.Values() {
        if enchantment.Owner == banner {
            enchantments = append(enchantments, enchantment)
        }
    }

    return enchantments
}

func (city *City) HasWallOfFire() bool {
    return city.HasEnchantment(data.CityEnchantmentWallOfFire)
}

func (city *City) HasWallOfDarkness() bool {
    return city.HasEnchantment(data.CityEnchantmentWallOfDarkness)
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
        extra += tile.GetBonus().PowerBonus()
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
                // FIXME: Wizards spell books should be reported in the city with the fortress
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

func cathedralPacification(race data.Race) int {
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

// https://masterofmagic.fandom.com/wiki/Town#Unrest_Reduction
func (city *City) InteracialUnrest() float64 {
    unrest := make(map[data.Race]map[data.Race]float64)

    set := func (race1 data.Race, race2 data.Race, value float64) {
        if _, ok := unrest[race1]; !ok {
            unrest[race1] = make(map[data.Race]float64)
        }
        unrest[race1][race2] = value

        if _, ok := unrest[race2]; !ok {
            unrest[race2] = make(map[data.Race]float64)
        }
        unrest[race2][race1] = value
    }

    set(data.RaceBarbarian, data.RaceBarbarian, 0)
    set(data.RaceBarbarian, data.RaceBeastmen, 0.1)
    set(data.RaceBarbarian, data.RaceDarkElf, 0.1)
    set(data.RaceBarbarian, data.RaceDraconian, 0.1)
    set(data.RaceBarbarian, data.RaceDwarf, 0.1)
    set(data.RaceBarbarian, data.RaceGnoll, 0.1)
    set(data.RaceBarbarian, data.RaceHalfling, 0.1)
    set(data.RaceBarbarian, data.RaceHighElf, 0.1)
    set(data.RaceBarbarian, data.RaceHighMen, 0.1)
    set(data.RaceBarbarian, data.RaceKlackon, 0.2)
    set(data.RaceBarbarian, data.RaceLizard, 0.1)
    set(data.RaceBarbarian, data.RaceNomad, 0)
    set(data.RaceBarbarian, data.RaceOrc, 0)
    set(data.RaceBarbarian, data.RaceTroll, 0.1)

    set(data.RaceBeastmen, data.RaceBeastmen, 0)
    set(data.RaceBeastmen, data.RaceDarkElf, 0.2)
    set(data.RaceBeastmen, data.RaceDraconian, 0.2)
    set(data.RaceBeastmen, data.RaceDwarf, 0.2)
    set(data.RaceBeastmen, data.RaceGnoll, 0)
    set(data.RaceBeastmen, data.RaceHalfling, 0.1)
    set(data.RaceBeastmen, data.RaceHighElf, 0.2)
    set(data.RaceBeastmen, data.RaceHighMen, 0.1)
    set(data.RaceBeastmen, data.RaceKlackon, 0.2)
    set(data.RaceBeastmen, data.RaceLizard, 0.1)
    set(data.RaceBeastmen, data.RaceNomad, 0.1)
    set(data.RaceBeastmen, data.RaceOrc, 0.1)
    set(data.RaceBeastmen, data.RaceTroll, 0.2)

    set(data.RaceDarkElf, data.RaceDarkElf, 0)
    set(data.RaceDarkElf, data.RaceDraconian, 0.2)
    set(data.RaceDarkElf, data.RaceDwarf, 0.3)
    set(data.RaceDarkElf, data.RaceGnoll, 0.2)
    set(data.RaceDarkElf, data.RaceHalfling, 0.2)
    set(data.RaceDarkElf, data.RaceHighElf, 0.4)
    set(data.RaceDarkElf, data.RaceHighMen, 0.2)
    set(data.RaceDarkElf, data.RaceKlackon, 0.2)
    set(data.RaceDarkElf, data.RaceLizard, 0.2)
    set(data.RaceDarkElf, data.RaceNomad, 0.2)
    set(data.RaceDarkElf, data.RaceOrc, 0.2)
    set(data.RaceDarkElf, data.RaceTroll, 0.3)

    set(data.RaceDraconian, data.RaceDraconian, 0)
    set(data.RaceDraconian, data.RaceDwarf, 0.2)
    set(data.RaceDraconian, data.RaceGnoll, 0.2)
    set(data.RaceDraconian, data.RaceHalfling, 0.1)
    set(data.RaceDraconian, data.RaceHighElf, 0.1)
    set(data.RaceDraconian, data.RaceHighMen, 0.1)
    set(data.RaceDraconian, data.RaceKlackon, 0.2)
    set(data.RaceDraconian, data.RaceLizard, 0.1)
    set(data.RaceDraconian, data.RaceNomad, 0.1)
    set(data.RaceDraconian, data.RaceOrc, 0.1)
    set(data.RaceDraconian, data.RaceTroll, 0.2)

    set(data.RaceDwarf, data.RaceDwarf, 0)
    set(data.RaceDwarf, data.RaceGnoll, 0.1)
    set(data.RaceDwarf, data.RaceHalfling, 0)
    set(data.RaceDwarf, data.RaceHighElf, 0.3)
    set(data.RaceDwarf, data.RaceHighMen, 0)
    set(data.RaceDwarf, data.RaceKlackon, 0.2)
    set(data.RaceDwarf, data.RaceLizard, 0.1)
    set(data.RaceDwarf, data.RaceNomad, 0)
    set(data.RaceDwarf, data.RaceOrc, 0.3)
    set(data.RaceDwarf, data.RaceTroll, 0.4)

    set(data.RaceGnoll, data.RaceGnoll, 0)
    set(data.RaceGnoll, data.RaceHalfling, 0)
    set(data.RaceGnoll, data.RaceHighElf, 0.3)
    set(data.RaceGnoll, data.RaceHighMen, 0)
    set(data.RaceGnoll, data.RaceKlackon, 0.2)
    set(data.RaceGnoll, data.RaceLizard, 0.1)
    set(data.RaceGnoll, data.RaceNomad, 0.1)
    set(data.RaceGnoll, data.RaceOrc, 0)
    set(data.RaceGnoll, data.RaceTroll, 0)

    set(data.RaceHalfling, data.RaceHalfling, 0)
    set(data.RaceHalfling, data.RaceHighElf, 0)
    set(data.RaceHalfling, data.RaceHighMen, 0)
    set(data.RaceHalfling, data.RaceKlackon, 0.2)
    set(data.RaceHalfling, data.RaceLizard, 0)
    set(data.RaceHalfling, data.RaceNomad, 0)
    set(data.RaceHalfling, data.RaceOrc, 0)
    set(data.RaceHalfling, data.RaceTroll, 0)

    set(data.RaceHighElf, data.RaceHighElf, 0)
    set(data.RaceHighElf, data.RaceHighMen, 0)
    set(data.RaceHighElf, data.RaceKlackon, 0.2)
    set(data.RaceHighElf, data.RaceLizard, 0.1)
    set(data.RaceHighElf, data.RaceNomad, 0)
    set(data.RaceHighElf, data.RaceOrc, 0.2)
    set(data.RaceHighElf, data.RaceTroll, 0.3)

    set(data.RaceHighMen, data.RaceHighMen, 0)
    set(data.RaceHighMen, data.RaceKlackon, 0.2)
    set(data.RaceHighMen, data.RaceLizard, 0.1)
    set(data.RaceHighMen, data.RaceNomad, 0)
    set(data.RaceHighMen, data.RaceOrc, 0)
    set(data.RaceHighMen, data.RaceTroll, 0.1)

    set(data.RaceKlackon, data.RaceKlackon, -0.2)
    set(data.RaceKlackon, data.RaceLizard, 0.2)
    set(data.RaceKlackon, data.RaceNomad, 0.2)
    set(data.RaceKlackon, data.RaceOrc, 0.2)
    set(data.RaceKlackon, data.RaceTroll, 0.2)

    set(data.RaceLizard, data.RaceLizard, 0)
    set(data.RaceLizard, data.RaceNomad, 0.1)
    set(data.RaceLizard, data.RaceOrc, 0.1)
    set(data.RaceLizard, data.RaceTroll, 0.1)

    set(data.RaceNomad, data.RaceNomad, 0)
    set(data.RaceNomad, data.RaceOrc, 0)
    set(data.RaceNomad, data.RaceTroll, 0.1)

    set(data.RaceOrc, data.RaceOrc, 0)
    set(data.RaceOrc, data.RaceTroll, 0)

    set(data.RaceTroll, data.RaceTroll, 0)

    return unrest[city.Race][city.RulingRace]
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

    unrestPercent += city.InteracialUnrest()

    // unrest from curses
    if city.HasEnchantment(data.CityEnchantmentFamine) {
        unrestPercent += 0.25
    }

    unrestAbsolute := float64(0)

    if city.HasEnchantment(data.CityEnchantmentCursedLands) {
        unrestAbsolute += 1
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
        pacification += float64(cathedralPacification(city.Race)) * pacificationRetort
    }

    if city.Buildings.Contains(buildinglib.BuildingAnimistsGuild) {
        pacification += float64(animistsGuildPacification(city.Race))
    }

    if city.Buildings.Contains(buildinglib.BuildingOracle) {
        pacification += float64(oraclePacification(city.Race))
    }

    // pacification from enchantments
    if city.HasEnchantment(data.CityEnchantmentGaiasBlessing) {
        pacification += 2
    }

    total := unrestPercent * float64(city.Citizens()) + unrestAbsolute - pacification - garrisonSupression / 2

    return int(math.Max(0, total))
}

/* returns the maximum number of citizens. population is citizens * 1000
 */
func (city *City) MaximumCitySize() int {
    foodAvailability := city.BaseFoodLevel()

    // TODO: 1/2 if famine is active

    bonus := 0

    if city.Buildings.Contains(buildinglib.BuildingGranary) {
        bonus += 2
    }

    if city.Buildings.Contains(buildinglib.BuildingFarmersMarket) {
        bonus += 3
    }

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

    if city.HasEnchantment(data.CityEnchantmentGaiasBlessing) {
        base += int(2.5 * float32(city.MaximumCitySize()) / 10) * 10
    }

    // FIXME: Add Famine, Stream of Life, Dark Rituals and Population Boom event


    if city.SurplusFood() < 0 {
        base = 50 * city.SurplusFood()
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
        food = food.Add(tile.FoodBonus())
    }

    if city.HasEnchantment(data.CityEnchantmentGaiasBlessing) {
        food = food.Add(food.Divide(fraction.FromInt(2)))
    }

    for _, tile := range catchment {
        food = food.Add(fraction.FromInt(tile.GetBonus().FoodBonus()))
    }

    return int(food.ToFloat())
}

func (city *City) FoodProductionRate() int {
    base := city.foodProductionRate(city.Farmers)

    // foresters guild doesn't contribute to the food needed to support the town, instead the food is added to the global surplus
    if city.Buildings.Contains(buildinglib.BuildingForestersGuild) {
        base += 2
    }

    if city.HasEnchantment(data.CityEnchantmentFamine) {
        base /= 2
    }

    return base
}

func (city *City) FarmerFoodProduction(farmers int) int {
    food := 2 * farmers

    if city.Race == data.RaceHalfling {
        food += farmers
    }

    if city.HasEnchantment(data.CityEnchantmentGaiasBlessing) {
        food += int(float32(food) * 0.2)
    }

    return food
}

func (city *City) foodProductionRate(farmers int) int {
    baseRate := float32(city.FarmerFoodProduction(farmers))

    /*
    if city.Buildings.Contains(buildinglib.BuildingForestersGuild) {
        baseRate += 2
    }
    */

    // FIXME: check if animists guild applies to the base rate or to the bonus
    if city.Buildings.Contains(buildinglib.BuildingAnimistsGuild) {
        baseRate += float32(farmers)
    }

    // TODO: if famine is active then base rate is halved

    baseLevel := float32(city.BaseFoodLevel())
    if baseRate > baseLevel {
        baseRate = baseLevel + (baseRate - baseLevel) / 2
    }

    bonus := 0

    if city.Buildings.Contains(buildinglib.BuildingGranary) {
        bonus += 2
    }

    if city.Buildings.Contains(buildinglib.BuildingFarmersMarket) {
        bonus += 3
    }

    bonus += city.ComputeWildGame()

    return int(baseRate) + bonus
}

func (city *City) ComputeWildGame() int {
    bonus := 0

    catchment := city.CatchmentProvider.GetCatchmentArea(city.X, city.Y)

    for _, tile := range catchment {
        if !tile.Corrupted() {
            bonus += tile.GetBonus().FoodBonus()
        }
    }

    return bonus
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
        extra += tile.GetBonus().GoldBonus()
    }

    if city.Race == data.RaceDwarf {
        extra *= 2
    }

    if city.Buildings.Contains(buildinglib.BuildingMinersGuild) {
        extra = int(float64(extra) * 1.5)
    }

    return extra
}

// return the percent of foreign trade bonus
//   population of other city * 0.5% if same race
//   population of other city * 1% if different
func (city *City) ComputeForeignTrade() float64 {
    connected := city.ConnectedProvider.FindRoadConnectedCities(city)
    percent := float64(0)

    for _, other := range connected {
        if other.Race == city.Race {
            percent += float64(other.Citizens()) * 0.5
        } else {
            percent += float64(other.Citizens()) * 1
        }
    }

    return percent
}

// return the bonus gold percent, capped at citizens*3
func (city *City) ComputeTotalBonusPercent() float64 {
    percent := city.ComputeForeignTrade()
    if city.Race == data.RaceNomad {
        percent += 50
    }

    // FIXME: add river/shore bonus
    // +10 if adjacent to a shore
    // +20 if on a river
    // +30 if on a river and adjacent to a shore, or on a river mouth

    return min(percent, float64(city.Citizens() * 3))
}

// gold from cities connected via roads, ocean, and river
func (city *City) GoldBonus(percent float64) int {
    return int(float64(city.GoldTaxation() + city.GoldMinerals()) * percent / 100)
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

func (city *City) GoldProsperity() int {
    if city.HasEnchantment(data.CityEnchantmentProsperity) {
        return city.GoldTaxation() + city.GoldMinerals()
    }

    return 0
}

func (city *City) GoldSurplus() int {
    income := city.GoldTaxation()
    income += city.GoldTradeGoods()
    income += city.GoldMinerals()
    income += city.GoldMarketplace()
    income += city.GoldBonus(city.ComputeTotalBonusPercent())
    income += city.GoldBank()
    income += city.GoldMerchantsGuild()
    income += city.GoldProsperity()

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
    return float32(math.Ceil(0.5 * float64(city.Farmers)))
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
    hasGaiasBlessing := city.HasEnchantment(data.CityEnchantmentGaiasBlessing)

    for _, tile := range catchment {
        production += float32(tile.ProductionBonus(hasGaiasBlessing)) / 100

        // FIXME: This should be only when producing units
        mineralProduction += float32(tile.GetBonus().UnitReductionBonus()) / 100
    }

    if city.Race == data.RaceDwarf {
        mineralProduction *= 2
    }

    if city.Buildings.Contains(buildinglib.BuildingMinersGuild) {
        mineralProduction *= 2
    }

    return (production + mineralProduction) * (city.ProductionWorkers() + city.ProductionFarmers())
}

func (city *City) ProductionInspirations() float32 {
    if city.HasEnchantment(data.CityEnchantmentInspirations) {
        return city.ProductionWorkers() + city.ProductionFarmers()
    }

    return 0
}

func (city *City) WorkProductionRate() float32 {
    result := city.ProductionWorkers() +
              city.ProductionFarmers() +
              city.ProductionMinersGuild() +
              city.ProductionMechaniciansGuild() +
              city.ProductionTerrain() +
              city.ProductionSawmill() +
              city.ProductionForestersGuild() +
              city.ProductionInspirations()

    if city.HasEnchantment(data.CityEnchantmentCursedLands) {
        result /= 2
    }

    return result
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
            cityEvents = append(cityEvents, &CityEventPopulationGrowth{Size: (city.Population - oldPopulation)/1000, Grow: city.Population > oldPopulation})
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
    }

    // update minimum farmers
    city.ResetCitizens(garrison)

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

func (city *City) GetSightRange() int {
    if city.HasEnchantment(data.CityEnchantmentNaturesEye) {
        return 5
    }

    if city.Buildings.Contains(buildinglib.BuildingOracle) {
        return 4
    }

    return 2
}

func ReadCityNames(cache *lbx.LbxCache) ([]string, error) {
    data, err := cache.GetLbxFile("cityname.lbx")
    if err != nil {
        return nil, fmt.Errorf("unable to read cityname.lbx: %v", err)
    }

    reader, err := data.GetReader(0)
    if err != nil {
        return nil, fmt.Errorf("unable to read entry 0 in cityname.lbx: %v", err)
    }

    entrySize, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil, fmt.Errorf("read error: %v", err)
    }

    numEntries, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil, fmt.Errorf("read error: %v", err)
    }

    var out []string

    for i := 0; i < int(numEntries); i++ {
        entry := make([]byte, entrySize)
        n, err := reader.Read(entry)
        if err != nil || n != int(entrySize) {
            return nil, fmt.Errorf("unable to read entry %v: %v", i, err)
        }
        out = append(out, string(bytes.Trim(entry, "\x00")))
    }

    return out, nil
}
