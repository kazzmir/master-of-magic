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
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
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

func (catchment *NoCatchment) GetGoldBonus(x int, y int) int {
    return 0
}

func (no *NoCatchment) OnShore(x int, y int) bool {
    return false
}

func (no *NoCatchment) ByRiver(x int, y int) bool {
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

func (no *NoServices) GetSpellByName(name string) spellbook.Spell {
    return spellbook.Spell{}
}

func (no *NoServices) GetAllGlobalEnchantments() map[data.BannerType]*set.Set[data.Enchantment] {
    enchantments := make(map[data.BannerType]*set.Set[data.Enchantment])
    return enchantments
}

func TestChangeCityOwner(test *testing.T){
    player1 := playerlib.MakePlayer(setup.WizardCustom{Banner: data.BannerRed, Race: data.RaceDraconian}, true, 1, 1, make(map[herolib.HeroType]string), nil)
    player2 := playerlib.MakePlayer(setup.WizardCustom{Banner: data.BannerGreen, Race: data.RaceDarkElf}, true, 1, 1, make(map[herolib.HeroType]string), nil)
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

func TestCastTurns(test *testing.T){
    // enough mana to cast in one turn
    cast1 := CastPlayer{
        player: nil,
        remainingCastingSkill: 10,
        castingSkill: 10,
        manaPerTurn: 5,
        mana: 100,
    }

    if cast1.ComputeTurnsToCast(5) != 0 {
        test.Errorf("Expected 0 turns but got %v", cast1.ComputeTurnsToCast(5))
    }

    // take multiple turns, but doesn't run out of mana
    cast2 := CastPlayer{
        player: nil,
        remainingCastingSkill: 10,
        castingSkill: 10,
        manaPerTurn: 0,
        mana: 100,
    }

    if cast2.ComputeTurnsToCast(15) != 1 {
        test.Errorf("Expected 1 turns but got %v", cast2.ComputeTurnsToCast(15))
    }

    // runs out of mana before casting
    cast3 := CastPlayer{
        player: nil,
        remainingCastingSkill: 10,
        castingSkill: 10,
        manaPerTurn: 4,
        mana: 10,
    }

    // first turn: 10 (casting skill), then 4 mana per turn
    if cast3.ComputeTurnsToCast(50) != 10 {
        test.Errorf("Expected 10 turns but got %v", cast3.ComputeTurnsToCast(50))
    }

    cast4 := CastPlayer{
        player: nil,
        remainingCastingSkill: 10,
        castingSkill: 10,
        manaPerTurn: 20,
        mana: 0,
    }

    // first turn: 0, then 10 for each turn
    if cast4.ComputeTurnsToCast(50) != 5 {
        test.Errorf("Expected 5 turns but got %v", cast4.ComputeTurnsToCast(50))
    }

    cast5 := CastPlayer{
        player: nil,
        remainingCastingSkill: 10,
        castingSkill: 10,
        manaPerTurn: -5,
        mana: 10,
    }

    // infinite turns
    if cast5.ComputeTurnsToCast(50) != 1000 {
        test.Errorf("Expected 1000 turns but got %v", cast5.ComputeTurnsToCast(50))
    }

    // if the player doesn't have casting skill somehow then it is infinite
    cast6 := CastPlayer{
        player: nil,
        castingSkill: 0,
        manaPerTurn: 5,
        mana: 10,
    }

    if cast6.ComputeTurnsToCast(50) != 1000 {
        test.Errorf("Expected 1000 turns but got %v", cast6.ComputeTurnsToCast(1000))
    }
}
