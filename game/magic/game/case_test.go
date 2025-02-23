package game

import (
	"testing"

	"github.com/kazzmir/master-of-magic/game/magic/data"
	herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
	playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
	"github.com/kazzmir/master-of-magic/game/magic/setup"
)

func TestComputeDispelChance(test *testing.T) {
    // No retorts
    expect := func(testName string, dispelChance int, expectedChance int) {
        if (dispelChance != expectedChance) {
            test.Errorf("Test %s: chance is %d, but %d expected", testName, dispelChance, expectedChance)
        }
    }
    playerWithoutRetorts := playerlib.MakePlayer(setup.WizardCustom{Banner: data.BannerRed, Race: data.RaceDraconian}, true, 1, 1, make(map[herolib.HeroType]string))

    expect("No retorts 1", ComputeDispelChance(1, 200, data.ChaosMagic, playerWithoutRetorts), 1)
    expect("No retorts 2", ComputeDispelChance(10, 20, data.ChaosMagic, playerWithoutRetorts), 83)
    expect("No retorts 3", ComputeDispelChance(30, 20, data.ChaosMagic, playerWithoutRetorts), 150)
    expect("No retorts 4", ComputeDispelChance(500, 20, data.ChaosMagic, playerWithoutRetorts), 240)
    expect("No retorts 5", ComputeDispelChance(10000, 20, data.ChaosMagic, playerWithoutRetorts), 249)

    // Archmage
    playerArchmage := playerlib.MakePlayer(setup.WizardCustom{Banner: data.BannerRed, Race: data.RaceDraconian}, true, 1, 1, make(map[herolib.HeroType]string))
    playerArchmage.Wizard.EnableAbility(setup.AbilityArchmage)
    playerChaosMastery := playerlib.MakePlayer(setup.WizardCustom{Banner: data.BannerRed, Race: data.RaceDraconian}, true, 1, 1, make(map[herolib.HeroType]string))
    playerChaosMastery.Wizard.EnableAbility(setup.AbilityChaosMastery)
    playerArchmageAndChaosMastery := playerlib.MakePlayer(setup.WizardCustom{Banner: data.BannerRed, Race: data.RaceDraconian}, true, 1, 1, make(map[herolib.HeroType]string))
    playerArchmageAndChaosMastery.Wizard.EnableAbility(setup.AbilityArchmage)
    playerArchmageAndChaosMastery.Wizard.EnableAbility(setup.AbilityChaosMastery)

    for cost := 100; cost <= 1000; cost += 300 {
        // Archmage and Mastery should both double the cost for calculation relative to no-retorts mage.
        expect("Archmage",
                ComputeDispelChance(10, cost, data.ChaosMagic, playerArchmage), 
                ComputeDispelChance(10, 2*cost, data.ChaosMagic, playerWithoutRetorts))
        expect("Chaos Mastery",
                ComputeDispelChance(10, cost, data.ChaosMagic, playerChaosMastery), 
                ComputeDispelChance(10, 2*cost, data.ChaosMagic, playerWithoutRetorts))
        // Archmage and Mastery together should triple the cost for calculation relative to no-retorts mage.
        expect("Archmage and Chaos Mastery",
                ComputeDispelChance(10, cost, data.ChaosMagic, playerArchmageAndChaosMastery), 
                ComputeDispelChance(10, 3*cost, data.ChaosMagic, playerWithoutRetorts))
        // Non-matching realm should make no effect
        expect("Non-matching realm mastery",
                ComputeDispelChance(10, cost, data.DeathMagic, playerChaosMastery), 
                ComputeDispelChance(10, cost, data.DeathMagic, playerWithoutRetorts))
    }
}

