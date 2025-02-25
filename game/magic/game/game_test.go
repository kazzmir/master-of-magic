package game

import (
    "testing"
    "image"
    "strings"

    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/set"
)

func TestCityName(test *testing.T){
    names := []string{"A", "B", "C", "A 1", "B 1"}
    choices := []string{"A", "B", "C"}

    chosen := chooseCityName(names, choices)
    if chosen != "C 1" {
        test.Errorf("Expected C 1 but got %v", chosen)
    }

    names = []string{"A", "B", "C", "A 1", "B 1", "C 1"}
    choices = []string{"A", "B", "C"}
    chosen = chooseCityName(names, choices)
    if !strings.Contains(chosen, "2") {
        test.Errorf("Expected 2 but got %v", chosen)
    }
}

type NoCatchment struct {}

func (no *NoCatchment) GetCatchmentArea(x int, y int) map[image.Point]maplib.FullTile {
    return make(map[image.Point]maplib.FullTile)
}

func (no *NoCatchment) OnShore(x int, y int) bool {
    return false
}

func (no *NoCatchment) TileDistance(x1 int, y1 int, x2 int, y2 int) int {
    return 0
}

type NoServices struct {}

func (no *NoServices) FindRoadConnectedCities(city *citylib.City) []*citylib.City {
    return nil
}

func (no *NoServices) GoodMoonActive() bool {
    return false
}

func (no *NoServices) BadMoonActive() bool {
    return false
}

func (no *NoServices) PopulationBoomActive(city *citylib.City) bool {
    return false
}

func (no *NoServices) PlagueActive(city *citylib.City) bool {
    return false
}

func (no *NoServices) GetAllGlobalEnchantments() map[data.BannerType]*set.Set[data.Enchantment] {
    enchantments := make(map[data.BannerType]*set.Set[data.Enchantment])
    return enchantments
}

func TestChangeCityOwner(test *testing.T){
    player1 := playerlib.MakePlayer(setup.WizardCustom{Banner: data.BannerRed, Race: data.RaceDraconian}, true, 1, 1, make(map[herolib.HeroType]string))
    player2 := playerlib.MakePlayer(setup.WizardCustom{Banner: data.BannerGreen, Race: data.RaceDarkElf}, true, 1, 1, make(map[herolib.HeroType]string))
    player1.TaxRate = fraction.Zero()
    player2.TaxRate = fraction.FromInt(2)

    city := citylib.MakeCity("xyz", 1, 1, player1.Wizard.Race, nil, &NoCatchment{}, &NoServices{}, player1)
    city.Population = 6000
    city.ResetCitizens()
    city.AddBuilding(buildinglib.BuildingFortress)
    city.AddEnchantment(data.CityEnchantmentAltarOfBattle, player1.GetBanner())
    player1.AddCity(city)

    if city.ComputeUnrest() != 0 {
        test.Errorf("Unrest is unexpected")
    }

    ChangeCityOwner(city, player1, player2, ChangeCityKeepEnchantments)

    if player1.OwnsCity(city) {
        test.Errorf("Player 1 still owns the city")
    }

    if !player2.OwnsCity(city) {
        test.Errorf("Player 2 does not own the city")
    }

    if !city.HasEnchantment(data.CityEnchantmentAltarOfBattle) {
        test.Errorf("City does not have the altar of battle")
    }

    if city.Buildings.Contains(buildinglib.BuildingFortress) {
        test.Errorf("City still has the fortress")
    }

    if city.ComputeUnrest() != 3 {
        test.Errorf("Unrest is not updated")
    }

    city.AddEnchantment(data.CityEnchantmentGaiasBlessing, player2.GetBanner())

    ChangeCityOwner(city, player2, player1, ChangeCityRemoveOwnerEnchantments)

    if player2.OwnsCity(city) {
        test.Errorf("Player 2 still owns the city")
    }

    if city.HasEnchantment(data.CityEnchantmentGaiasBlessing) {
        test.Errorf("gaias blessing still on city")
    }
}
