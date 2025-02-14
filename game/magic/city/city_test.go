package city

import (
    "testing"
    "image"
    "math"

    "github.com/kazzmir/master-of-magic/game/magic/building"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/lib/fraction"
)

func makeSimpleMap() map[image.Point]maplib.FullTile {
    out := make(map[image.Point]maplib.FullTile)
    for x := -2; x <= 2; x++ {
        for y := -2; y <= 2; y++ {
            out[image.Point{x, y}] = maplib.FullTile{
                Tile: terrain.TileGrasslands1,
            }
        }
    }
    return out
}

type Catchment struct {
    Map map[image.Point]maplib.FullTile
}

func (catchment *Catchment) GetCatchmentArea(x int, y int) map[image.Point]maplib.FullTile {
    return catchment.Map
}

func (catchment *Catchment) OnShore(x int, y int) bool {
    return false
}

type NoCities struct {
}

func (provider *NoCities) FindRoadConnectedCities(city *City) []*City {
    return nil
}

func (provider *NoCities) GoodMoonActive() bool {
    return false
}

func (provider *NoCities) BadMoonActive() bool {
    return false
}

func (provider *NoCities) PopulationBoomActive(city *City) bool {
    return false
}

func (provider *NoCities) PlagueActive(city *City) bool {
    return false
}

func TestBasicCity(test *testing.T){
    city := MakeCity("Test City", 10, 10, data.RaceHighMen, data.BannerBlue, fraction.Make(3, 2), nil, &Catchment{Map: makeSimpleMap()}, &NoCities{})
    city.Population = 6000
    city.Farmers = 6
    city.Workers = 0
    city.ResetCitizens(nil)
    if city.Name != "Test City" {
        test.Error("City name is not correct")
    }
    if city.X != 10 {
        test.Error("City X is not correct")
    }
    if city.Y != 10 {
        test.Error("City Y is not correct")
    }

    if city.Citizens() != 6 {
        test.Errorf("Citizens should have been 6 but was %v", city.Citizens())
    }

    if city.ComputeSubsistenceFarmers() != 3 {
        test.Errorf("Subsistence farmers should have been 3 but was %v", city.ComputeSubsistenceFarmers())
    }

    if city.ComputeUnrest(nil) != 1 {
        test.Errorf("Unrest should have been 1 but was %v", city.ComputeUnrest(nil))
    }

    if city.Rebels != 1 {
        test.Errorf("Rebels should have been 1 but was %v", city.Rebels)
    }

    city.UpdateTaxRate(fraction.Make(3, 1), nil)

    if city.ComputeUnrest(nil) != 4 {
        test.Errorf("Unrest should have been 4 but was %v", city.ComputeUnrest(nil))
    }

    if city.Rebels != 3 {
        test.Errorf("Rebels should have been 3 but was %v", city.Rebels)
    }

    if city.Workers != 0 {
        test.Errorf("Workers should have been 0 but was %v", city.Workers)
    }

    if city.Farmers != 3 {
        test.Errorf("Farmers should have been 3 but was %v", city.Farmers)
    }
}

type AllConnected struct {
    Cities []*City
}

func (provider *AllConnected) FindRoadConnectedCities(city *City) []*City {
    var out []*City

    for _, other := range provider.Cities {
        if other != city {
            out = append(out, other)
        }
    }

    return out
}

func (provider *AllConnected) GoodMoonActive() bool {
    return false
}

func (provider *AllConnected) BadMoonActive() bool {
    return false
}

func (provider *AllConnected) PopulationBoomActive(city *City) bool {
    return false
}

func (provider *AllConnected) PlagueActive(city *City) bool {
    return false
}

func closeFloat(a float64, b float64) bool {
    return math.Abs(a - b) < 0.0001
}

func TestForeignTrade(test *testing.T){
    var connected AllConnected
    city1 := MakeCity("Test City", 10, 10, data.RaceHighMen, data.BannerBlue, fraction.Make(3, 2), nil, &Catchment{Map: makeSimpleMap()}, &connected)
    city1.Population = 6000
    city1.Farmers = 6
    city1.Workers = 0
    city1.ResetCitizens(nil)

    city2 := MakeCity("Test City 2", 10, 10, data.RaceHighMen, data.BannerBlue, fraction.Make(3, 2), nil, &Catchment{Map: makeSimpleMap()}, &connected)
    city2.Population = 7000
    city2.Farmers = 7
    city2.Workers = 0
    city2.ResetCitizens(nil)

    connected.Cities = []*City{city1, city2}

    city1Trade := city1.ComputeForeignTrade()
    if !closeFloat(city1Trade, 7 * 0.5) {
        test.Errorf("City1 foreign trade expected %v but was %v", 7 * 0.5, city1Trade)
    }

    // different race
    city3 := MakeCity("Test City 3 elf", 10, 10, data.RaceHighElf, data.BannerBlue, fraction.Make(3, 2), nil, &Catchment{Map: makeSimpleMap()}, &connected)
    city3.Population = 5000
    city3.Farmers = 5
    city3.Workers = 0
    city3.ResetCitizens(nil)

    connected.Cities = []*City{city1, city2, city3}

    city1Trade = city1.ComputeForeignTrade()
    expected := 7 * 0.5 + 5 * 1
    if !closeFloat(city1Trade, expected) {
        test.Errorf("City1 foreign trade expected %v but was %v", expected, city1Trade)
    }
}

func TestEnchantments(test *testing.T){
    banner := data.BannerBlue

    map_ := make(map[image.Point]maplib.FullTile)
    for x := -2; x <= 2; x++ {
        for y := -2; y <= 2; y++ {
            map_[image.Point{x, y}] = maplib.FullTile{
                Tile: terrain.TileForest1,
            }
        }
    }
    catchment := Catchment{Map: map_}

    city := MakeCity("Test City", 10, 10, data.RaceHighMen, banner, fraction.FromInt(1), nil, &catchment, &NoCities{})
    city.Population = 10100
    city.Farmers = 5
    city.Workers = 3
    city.Rebels = 2
    city.ProducingBuilding = building.BuildingTradeGoods
    // FIXME: mock BuildingInfos, add buildings below and update power values
    // city.AddBuilding(building.BuildingShrine)
    // city.AddBuilding(building.BuildingTemple)

    stack := []units.StackUnit{}

    if city.FoodProductionRate() != 10 {
        // 5 * 2 farmer
        test.Errorf("City FoodProductionRate is not correct: %v", city.FoodProductionRate())
    }

    if int(city.WorkProductionRate()) != int(math.Floor(15.75)) {
        // 3 x 2 worker + 5 x 0.5 farmer + 6.75 terrain
        test.Errorf("City WorkProductionRate is not correct: %v", city.WorkProductionRate())
    }

    if city.GoldSurplus() != 15 {
        // 8 taxation + 15.75 / 2 trade goods
        test.Errorf("City GoldSurplus is not correct: %v", city.GoldSurplus())
    }

    if city.ComputeUnrest(stack) != 2 {
        // 0.2 * 10 race
        test.Errorf("City ComputeUnrest is not correct: %v", city.ComputeUnrest(stack))
    }

    if city.PopulationGrowthRate() != 15 {
        // 10 * (12 - 10 + 1) / 2 max city size and population
        test.Errorf("City PopulationGrowthRate is not correct: %v", city.PopulationGrowthRate())
    }

    // Prosperity
    city.AddEnchantment(data.CityEnchantmentProsperity, banner)

    if city.FoodProductionRate() != 10 {
        // 5 * 2 farmer
        test.Errorf("City FoodProductionRate is not correct: %v", city.FoodProductionRate())
    }

    if int(city.WorkProductionRate()) != int(math.Floor(15.75)) {
        // 3 x 2 worker + 5 x 0.5 farmer + 6.75 terrain
        test.Errorf("City WorkProductionRate is not correct: %v", city.WorkProductionRate())
    }

    if city.GoldSurplus() != 23 {
        // 2 x 8 taxation + 15.75 / 2 trade goods
        test.Errorf("City GoldSurplus is not correct: %v", city.GoldSurplus())
    }

    if city.ComputeUnrest(stack) != 2 {
        // 0.2 * 10 race
        test.Errorf("City ComputeUnrest is not correct: %v", city.ComputeUnrest(stack))
    }

    if city.PopulationGrowthRate() != 15 {
        // 10 * (12 - 10 + 1) / 2 max city size and population
        test.Errorf("City PopulationGrowthRate is not correct: %v", city.PopulationGrowthRate())
    }

    // Inspirations
    city.AddEnchantment(data.CityEnchantmentInspirations, banner)

    if city.FoodProductionRate() != 10 {
        // 5 * 2 farmer
        test.Errorf("City FoodProductionRate is not correct: %v", city.FoodProductionRate())
    }

    if int(city.WorkProductionRate()) != int(math.Floor(24.75)) {
        // 2 x (3 x 2 worker + 5 x 0.5 farmer) + 6.75 terrain
        test.Errorf("City WorkProductionRate is not correct: %v", city.WorkProductionRate())
    }

    if city.GoldSurplus() != 28 {
        // 2 x 8 taxation + 24.75 / 2 trade goods
        test.Errorf("City GoldSurplus is not correct: %v", city.GoldSurplus())
    }

    if city.ComputeUnrest(stack) != 2 {
        // 0.2 * 10 race
        test.Errorf("City ComputeUnrest is not correct: %v", city.ComputeUnrest(stack))
    }

    if city.PopulationGrowthRate() != 15 {
        // 10 * (12 - 10 + 1) / 2 max city size and population
        test.Errorf("City PopulationGrowthRate is not correct: %v", city.PopulationGrowthRate())
    }

    // Cursed Lands
    city.AddEnchantment(data.CityEnchantmentCursedLands, banner)

    if city.FoodProductionRate() != 10 {
        // 5 * 2 farmer
        test.Errorf("City FoodProductionRate is not correct: %v", city.FoodProductionRate())
    }

    if int(city.WorkProductionRate()) != int(math.Floor(12.375)) {
        // (2 x (3 x 2 worker + 5 x 0.5 farmer) + 6.75 terrain) / 2
        test.Errorf("City WorkProductionRate is not correct: %v", city.WorkProductionRate())
    }

    if city.GoldSurplus() != 22 {
        // 2 x 8 taxation + 12.375/2 trade goods
        test.Errorf("City GoldSurplus is not correct: %v", city.GoldSurplus())
    }

    if city.ComputeUnrest(stack) != 3 {
        // 0.2 * 10 race + 1 cursed lands
        test.Errorf("City ComputeUnrest is not correct: %v", city.ComputeUnrest(stack))
    }

    if city.PopulationGrowthRate() != 15 {
        // 10 * (12 - 10 + 1) / 2 max city size and population
        test.Errorf("City PopulationGrowthRate is not correct: %v", city.PopulationGrowthRate())
    }

    // Gaias Blessing
    city.AddEnchantment(data.CityEnchantmentGaiasBlessing, banner)

    if city.FoodProductionRate() != 12 {
        // 5 * 2 farmer + 0.2 * 10
        test.Errorf("City FoodProductionRate is not correct: %v", city.FoodProductionRate())
    }

    if int(city.WorkProductionRate()) != int(math.Floor(15.75)) {
        // (2 x (3 x 2 worker + 5 x 0.5 farmer) + 13.5 terrain) / 2 = 12.375
        test.Errorf("City WorkProductionRate is not correct: %v", city.WorkProductionRate())
    }

    if city.GoldSurplus() != 23 {
        // 2 x 8 taxation + 15.75/2 trade goods
        test.Errorf("City GoldSurplus is not correct: %v", city.GoldSurplus())
    }

    if city.ComputeUnrest(stack) != 1 {
        // 0.2 * 10 race + 1 cursed lands - 2 gaias blessing
        test.Errorf("City ComputeUnrest is not correct: %v", city.ComputeUnrest(stack))
    }

    if city.PopulationGrowthRate() != 85 {
        // 10 * (18 - 10 + 1) / 2 max city size and population + (2.5 * 18) rounded to 10s gaias blessing
        test.Errorf("City PopulationGrowthRate is not correct: %v", city.PopulationGrowthRate())
    }

    // Dark Rituals
    city.AddEnchantment(data.CityEnchantmentDarkRituals, banner)

    if city.FoodProductionRate() != 12 {
        // 5 * 2 farmer + 0.2 * 10
        test.Errorf("City FoodProductionRate is not correct: %v", city.FoodProductionRate())
    }

    if int(city.WorkProductionRate()) != int(math.Floor(15.75)) {
        // (2 x (3 x 2 worker + 5 x 0.5 farmer) + 13.5 terrain) / 2 = 12.375
        test.Errorf("City WorkProductionRate is not correct: %v", city.WorkProductionRate())
    }

    if city.GoldSurplus() != 23 {
        // 2 x 8 taxation + 15.75/2 trade goods
        test.Errorf("City GoldSurplus is not correct: %v", city.GoldSurplus())
    }

    if city.ComputeUnrest(stack) != 2 {
        // 0.2 * 10 race + 1 cursed lands - 2 gaias blessing + 1 dark rituals
        test.Errorf("City ComputeUnrest is not correct: %v", city.ComputeUnrest(stack))
    }

    if city.PopulationGrowthRate() != 63 {
        // (10 * (18 - 10 + 1) / 2 max city size and population + (2.5 * 18) rounded to 10s gaias blessing) * 0.75 dark rituals
        test.Errorf("City PopulationGrowthRate is not correct: %v", city.PopulationGrowthRate())
    }

    // Famine
    city.AddEnchantment(data.CityEnchantmentFamine, banner)

    if city.FoodProductionRate() != 5 {
        // ((5 * 2 farmer + 0.2 * 10) with halved excess) / 2  = 5.25
        test.Errorf("City FoodProductionRate is not correct: %v", city.FoodProductionRate())
    }

    if int(city.WorkProductionRate()) != int(math.Floor(15.75)) {
        // (2 x (3 x 2 worker + 5 x 0.5 farmer) + 13.5 terrain) / 2 = 12.375
        test.Errorf("City WorkProductionRate is not correct: %v", city.WorkProductionRate())
    }

    if city.GoldSurplus() != 23 {
        // 2 x 8 taxation + 15.75/2 trade goods
        test.Errorf("City GoldSurplus is not correct: %v", city.GoldSurplus())
    }

    if city.ComputeUnrest(stack) != 4 {
        // (0.2 race + 0.25 famine) * 10 + 1 cursed lands - 2 gaias blessing + 1 dark rituals
        test.Errorf("City ComputeUnrest is not correct: %v", city.ComputeUnrest(stack))
    }

    if city.PopulationGrowthRate() != -250 {
        // 5 * (5 - 10) food surplus
        test.Errorf("City PopulationGrowthRate is not correct: %v", city.PopulationGrowthRate())
    }

    // Pestilence
    city.AddEnchantment(data.CityEnchantmentPestilence, banner)

    if city.FoodProductionRate() != 5 {
        // ((5 * 2 farmer + 0.2 * 10) with halved excess) / 2  = 5.25
        test.Errorf("City FoodProductionRate is not correct: %v", city.FoodProductionRate())
    }

    if int(city.WorkProductionRate()) != int(math.Floor(15.75)) {
        // (2 x (3 x 2 worker + 5 x 0.5 farmer) + 13.5 terrain) / 2 = 12.375
        test.Errorf("City WorkProductionRate is not correct: %v", city.WorkProductionRate())
    }

    if city.GoldSurplus() != 23 {
        // 2 x 8 taxation + 15.75/2 trade goods
        test.Errorf("City GoldSurplus is not correct: %v", city.GoldSurplus())
    }

    if city.ComputeUnrest(stack) != 6 {
        // (0.2 race + 0.25 famine) * 10 + 1 cursed lands - 2 gaias blessing + 1 dark rituals + 2 pestilence
        test.Errorf("City ComputeUnrest is not correct: %v", city.ComputeUnrest(stack))
    }

    if city.PopulationGrowthRate() != -250 {
        // 5 * (5 - 10) food surplus
        test.Errorf("City PopulationGrowthRate is not correct: %v", city.PopulationGrowthRate())
    }
}


func makeScenarioMap() map[image.Point]maplib.FullTile {
    out := make(map[image.Point]maplib.FullTile)

    out[image.Point{-2, -2}] = maplib.FullTile{Tile: terrain.TileHills1}
    out[image.Point{-2, -1}] = maplib.FullTile{Tile: terrain.TileForest1}
    out[image.Point{-2,  0}] = maplib.FullTile{Tile: terrain.TileHills1, Extras: make(map[maplib.ExtraKind]maplib.ExtraTile)}
    out[image.Point{-2,  1}] = maplib.FullTile{Tile: terrain.TileRiver0001}
    out[image.Point{-2,  2}] = maplib.FullTile{Tile: terrain.TileForest1}

    out[image.Point{-1, -2}] = maplib.FullTile{Tile: terrain.TileHills1}
    out[image.Point{-1, -1}] = maplib.FullTile{Tile: terrain.TileForest1}
    out[image.Point{-1,  0}] = maplib.FullTile{Tile: terrain.TileHills1, Extras: make(map[maplib.ExtraKind]maplib.ExtraTile)}
    out[image.Point{-1,  1}] = maplib.FullTile{Tile: terrain.TileRiver0001}
    out[image.Point{-1,  2}] = maplib.FullTile{Tile: terrain.TileRiver0001}

    out[image.Point{ 0, -2}] = maplib.FullTile{Tile: terrain.TileShore1_00000001}
    out[image.Point{ 0, -1}] = maplib.FullTile{Tile: terrain.TileShore1_00000001}
    out[image.Point{ 0,  0}] = maplib.FullTile{Tile: terrain.TileRiver0001}
    out[image.Point{ 0,  1}] = maplib.FullTile{Tile: terrain.TileRiver0001}
    out[image.Point{ 0,  2}] = maplib.FullTile{Tile: terrain.TileForest1}

    out[image.Point{ 1, -2}] = maplib.FullTile{Tile: terrain.TileShore1_00000001}
    out[image.Point{ 1, -1}] = maplib.FullTile{Tile: terrain.TileShore1_00000001}
    out[image.Point{ 1,  0}] = maplib.FullTile{Tile: terrain.TileForest1}
    out[image.Point{ 1,  1}] = maplib.FullTile{Tile: terrain.TileForest1}
    out[image.Point{ 1,  2}] = maplib.FullTile{Tile: terrain.TileForest1}

    out[image.Point{ 2, -2}] = maplib.FullTile{Tile: terrain.TileShore1_00000001}
    out[image.Point{ 2, -1}] = maplib.FullTile{Tile: terrain.TileGrasslands1}
    out[image.Point{ 2,  0}] = maplib.FullTile{Tile: terrain.TileTundra}
    out[image.Point{ 2,  1}] = maplib.FullTile{Tile: terrain.TileHills1}
    out[image.Point{ 2,  2}] = maplib.FullTile{Tile: terrain.TileHills1}

    out[image.Point{-2,  0}].Extras[maplib.ExtraKindBonus] = &maplib.ExtraBonus{Bonus: data.BonusGoldOre}
    out[image.Point{-1,  0}].Extras[maplib.ExtraKindBonus] = &maplib.ExtraBonus{Bonus: data.BonusIronOre}

    return out
}

func TestScenario1(test *testing.T) {
    // Test against values from a city screen of original MoM v1.60

    // City
    city := MakeCity("Schleswig", 10, 10, data.RaceBarbarian, data.BannerGreen, fraction.FromInt(1), nil, &Catchment{Map: makeScenarioMap()}, &NoCities{})
    city.Population = 4600
    city.Farmers = 3
    city.Workers = 1
    city.AddBuilding(building.BuildingBarracks)
    city.AddBuilding(building.BuildingBuildersHall)
    city.AddBuilding(building.BuildingSmithy)
    city.AddBuilding(building.BuildingFortress)
    city.ProducingBuilding = building.BuildingHousing
    // maybe add 2 units garrison and call city.ResetCitizens(nil)

    // Food
    if city.FarmerFoodProduction(city.Farmers) != 6 {
        test.Errorf("City FarmerFoodProduction is not correct: %v", city.FarmerFoodProduction(city.Farmers))
    }
    if city.RequiredFood() != 4 {
        test.Errorf("City RequiredFood is not correct: %v", city.RequiredFood())
    }
    if city.SurplusFood() != 2 {
        test.Errorf("City SurplusFood is not correct: %v", city.SurplusFood())
    }

    // Production
    if int(city.WorkProductionRate()) != 5 {
        test.Errorf("City WorkProductionRate is not correct: %v", city.WorkProductionRate())
    }
    if int(city.ProductionWorkers()) != 2 {
        test.Errorf("City ProductionWorkers is not correct: %v", city.ProductionWorkers())
    }
    if int(city.ProductionFarmers()) != 2 {
        test.Errorf("City ProductionFarmers is not correct: %v", city.ProductionFarmers())
    }
    if int(city.ProductionTerrain()) != 1 {
        test.Errorf("City ProductionTerrain is not correct: %v", city.ProductionTerrain())
    }

    // Gold
    if city.GoldTaxation() != 4 {
        test.Errorf("City GoldTaxation is not correct: %v", city.GoldTaxation())
    }
    if city.GoldMinerals() != 3 {
        test.Errorf("City GoldMinerals is not correct: %v", city.GoldMinerals())
    }

    // Power
    books := []data.WizardBook{data.WizardBook{Magic: data.LifeMagic, Count: 11}}
    if city.ComputePower(books) != 11 {
        test.Logf("City ComputePower is not correct: %v", city.ComputePower(books))
    }

    if city.PopulationGrowthRate() != 120 {
        test.Errorf("City PopulationGrowthRate is not correct: %v", city.PopulationGrowthRate())
    }
}

func TestScenario2(test *testing.T) {
    // Test against values from a city screen of original MoM v1.60

    // City
    city := MakeCity("Schleswig", 10, 10, data.RaceBarbarian, data.BannerGreen, fraction.FromInt(1), nil, &Catchment{Map: makeScenarioMap()}, &NoCities{})
    city.Population = 10110
    city.Farmers = 7
    city.Workers = 3
    city.AddBuilding(building.BuildingBarracks)
    city.AddBuilding(building.BuildingBuildersHall)
    city.AddBuilding(building.BuildingSmithy)
    city.AddBuilding(building.BuildingFortress)
    city.AddBuilding(building.BuildingCityWalls)
    city.AddBuilding(building.BuildingGranary)
    city.AddBuilding(building.BuildingMarketplace)
    city.AddBuilding(building.BuildingMinersGuild)
    city.ProducingBuilding = building.BuildingTradeGoods
    // maybe add 6 units garrison and call city.ResetCitizens(nil)

    // Food
    if city.FarmerFoodProduction(city.Farmers) != 14 {
        test.Errorf("City FarmerFoodProduction is not correct: %v", city.FarmerFoodProduction(city.Farmers))
    }
    if city.RequiredFood() != 10 {
        test.Errorf("City RequiredFood is not correct: %v", city.RequiredFood())
    }
    if city.SurplusFood() != 6 {
        test.Errorf("City SurplusFood is not correct: %v", city.SurplusFood())
    }

    // Production
    if int(city.WorkProductionRate()) != 18 {
        // see "FIXME: This should be only when producing units
        test.Logf("City WorkProductionRate is not correct: %v", city.WorkProductionRate())
    }
    if int(city.ProductionWorkers()) != 6 {
        test.Errorf("City ProductionWorkers is not correct: %v", city.ProductionWorkers())
    }
    if int(city.ProductionFarmers()) != 4 {
        test.Errorf("City ProductionFarmers is not correct: %v", city.ProductionFarmers())
    }
    if int(city.ProductionTerrain()) != 3 {
        // see "FIXME: This should be only when producing units
        test.Logf("City ProductionTerrain is not correct: %v", city.ProductionTerrain())
    }
    if int(city.ProductionMinersGuild()) != 5 {
        test.Errorf("City ProductionTerrain is not correct: %v", city.ProductionTerrain())
    }

    // Gold
    if city.GoldTaxation() != 10 {
        test.Errorf("City GoldTaxation is not correct: %v", city.GoldTaxation())
    }
    if city.GoldTradeGoods() != 9 {
        test.Errorf("City GoldTradeGoods is not correct: %v", city.GoldTradeGoods())
    }
    if city.GoldMinerals() != 4 { // Gold Mine + Miner's Guild
        test.Errorf("City GoldMinerals is not correct: %v", city.GoldMinerals())
    }
    if city.GoldBonus(city.ComputeTotalBonusPercent()) != 4 {
        // see "FIXME: add river/shore bonus"
        test.Logf("City GoldBonus is not correct: %v", city.GoldBonus(city.ComputeTotalBonusPercent()))
    }
    if city.GoldMarketplace() != 7 {
        test.Errorf("City GoldMarketplace is not correct: %v", city.GoldMarketplace())
    }

    // Power
    books := []data.WizardBook{data.WizardBook{Magic: data.LifeMagic, Count: 11}}
    if city.ComputePower(books) != 11 {
        test.Logf("City ComputePower is not correct: %v", city.ComputePower(books))
    }

    if city.PopulationGrowthRate() != 90 {
        // Dos MoM seems to be doing this other than explained in the wiki
        test.Logf("City PopulationGrowthRate is not correct: %v", city.PopulationGrowthRate())
    }
}
