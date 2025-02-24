package game

import (
	"testing"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
	"github.com/kazzmir/master-of-magic/game/magic/setup"
)

func TestComputeDispelChance(test *testing.T) {
    // No retorts
    expect := func(testName string, dispelChance int, expectedChance int) {
        if (dispelChance != expectedChance) {
            test.Errorf("Test %s: chance is %d, but %d expected", testName, dispelChance, expectedChance)
        }
    }
    playerWithoutRetorts := setup.WizardCustom{Banner: data.BannerRed, Race: data.RaceDraconian}

    expect("No retorts 1", spellbook.ComputeDispelChance(1, 200, data.ChaosMagic, &playerWithoutRetorts), 1)
    expect("No retorts 2", spellbook.ComputeDispelChance(10, 20, data.ChaosMagic, &playerWithoutRetorts), 83)
    expect("No retorts 3", spellbook.ComputeDispelChance(30, 20, data.ChaosMagic, &playerWithoutRetorts), 150)
    expect("No retorts 4", spellbook.ComputeDispelChance(500, 20, data.ChaosMagic, &playerWithoutRetorts), 240)
    expect("No retorts 5", spellbook.ComputeDispelChance(10000, 20, data.ChaosMagic, &playerWithoutRetorts), 249)

    // Archmage
    // playerArchmage := playerlib.MakePlayer(setup.WizardCustom{Banner: data.BannerRed, Race: data.RaceDraconian}, true, 1, 1, make(map[herolib.HeroType]string))
    playerArchmage := setup.WizardCustom{Banner: data.BannerRed, Race: data.RaceDraconian}
    playerArchmage.EnableRetort(data.RetortArchmage)
    playerChaosMastery := setup.WizardCustom{Banner: data.BannerRed, Race: data.RaceDraconian}
    playerChaosMastery.EnableRetort(data.RetortChaosMastery)
    playerArchmageAndChaosMastery := setup.WizardCustom{Banner: data.BannerRed, Race: data.RaceDraconian}
    playerArchmageAndChaosMastery.EnableRetort(data.RetortArchmage)
    playerArchmageAndChaosMastery.EnableRetort(data.RetortChaosMastery)

    for cost := 100; cost <= 1000; cost += 300 {
        // Archmage and Mastery should both double the cost for calculation relative to no-retorts mage.
        expect("Archmage",
                spellbook.ComputeDispelChance(10, cost, data.ChaosMagic, &playerArchmage), 
                spellbook.ComputeDispelChance(10, 2*cost, data.ChaosMagic, &playerWithoutRetorts))
        expect("Chaos Mastery",
                spellbook.ComputeDispelChance(10, cost, data.ChaosMagic, &playerChaosMastery), 
                spellbook.ComputeDispelChance(10, 2*cost, data.ChaosMagic, &playerWithoutRetorts))
        // Archmage and Mastery together should triple the cost for calculation relative to no-retorts mage.
        expect("Archmage and Chaos Mastery",
                spellbook.ComputeDispelChance(10, cost, data.ChaosMagic, &playerArchmageAndChaosMastery), 
                spellbook.ComputeDispelChance(10, 3*cost, data.ChaosMagic, &playerWithoutRetorts))
        // Non-matching realm should make no effect
        expect("Non-matching realm mastery",
                spellbook.ComputeDispelChance(10, cost, data.NatureMagic, &playerChaosMastery), 
                spellbook.ComputeDispelChance(10, cost, data.NatureMagic, &playerWithoutRetorts))
    }
}

