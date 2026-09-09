package player

import (
    "testing"

    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
)

func TestIsAwareOf(test *testing.T) {
    human := MakePlayer(setup.WizardCustom{Banner: data.BannerRed}, true, 8, 8, make(map[herolib.HeroType]string), nil)
    enemy := MakePlayer(setup.WizardCustom{Banner: data.BannerGreen}, false, 8, 8, make(map[herolib.HeroType]string), nil)

    if human.IsAwareOf(enemy) {
        test.Errorf("players should not know each other at the start")
    }

    human.AwarePlayer(enemy)
    if !human.IsAwareOf(enemy) {
        test.Errorf("AwarePlayer should make the human aware of the enemy")
    }
    if enemy.IsAwareOf(human) {
        test.Errorf("awareness is one-way until both sides call AwarePlayer")
    }
}

func TestCanSeePlayerCity(test *testing.T) {
    human := MakePlayer(setup.WizardCustom{Banner: data.BannerRed}, true, 8, 8, make(map[herolib.HeroType]string), nil)
    enemy := MakePlayer(setup.WizardCustom{Banner: data.BannerGreen}, false, 8, 8, make(map[herolib.HeroType]string), nil)

    city := &citylib.City{Name: "Enemy", X: 3, Y: 4, Plane: data.PlaneArcanus}
    enemy.AddCity(city)

    if human.CanSeePlayer(enemy) {
        test.Errorf("should not see an unexplored enemy city")
    }

    human.LiftFogSquare(3, 4, 0, data.PlaneArcanus)
    if !human.CanSeePlayer(enemy) {
        test.Errorf("should see an enemy city on a visible tile")
    }
}

func TestCanSeePlayerStack(test *testing.T) {
    human := MakePlayer(setup.WizardCustom{Banner: data.BannerRed}, true, 8, 8, make(map[herolib.HeroType]string), nil)
    enemy := MakePlayer(setup.WizardCustom{Banner: data.BannerGreen}, false, 8, 8, make(map[herolib.HeroType]string), nil)

    unit := units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 2, 2, data.PlaneArcanus, data.BannerGreen, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    enemy.AddUnit(unit)

    if human.CanSeePlayer(enemy) {
        test.Errorf("should not see an enemy stack in the fog")
    }

    human.LiftFogSquare(2, 2, 0, data.PlaneArcanus)
    if !human.CanSeePlayer(enemy) {
        test.Errorf("should see an enemy stack on a visible tile")
    }
}

func TestCanSeePlayerExploredCityIsHidden(test *testing.T) {
    human := MakePlayer(setup.WizardCustom{Banner: data.BannerRed}, true, 8, 8, make(map[herolib.HeroType]string), nil)
    enemy := MakePlayer(setup.WizardCustom{Banner: data.BannerGreen}, false, 8, 8, make(map[herolib.HeroType]string), nil)

    city := &citylib.City{Name: "Enemy", X: 1, Y: 1, Plane: data.PlaneArcanus}
    enemy.AddCity(city)

    human.ExploreFogSquare(1, 1, 0, data.PlaneArcanus)
    if human.CanSeePlayer(enemy) {
        test.Errorf("an explored but not currently visible city should not count as seeing the wizard")
    }
}
