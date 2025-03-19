package player

import (
    "testing"

    "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
)

func TestHeroNames(test *testing.T) {
    names := make(map[hero.HeroType]string)
    names[hero.HeroFang] = "goofy"

    player := MakePlayer(setup.WizardCustom{}, true, 5, 5, names, nil)
    fangName := player.HeroPool[hero.HeroFang].GetName()

    if fangName != "goofy" {
        test.Errorf("Expected hero name to be 'goofy', got '%s'", fangName)
    }
}

func TestSkillReduction(test *testing.T) {
    setPlayerInvestedPower := func (player *Player, desiredSkill int) { player.CastingSkillPower = desiredSkill * desiredSkill - desiredSkill + 1} // skill^2 - skill + 1, simplification of (skill-1)^2 + skill
    names := make(map[hero.HeroType]string)
    player := MakePlayer(setup.WizardCustom{}, true, 5, 5, names, nil)

    // Trying to reduce for various values
    for initialSkill := 1; initialSkill < 1000; initialSkill++ {
        for reduceBy := 1; reduceBy <= initialSkill; reduceBy++ {
            setPlayerInvestedPower(player, initialSkill)
            actualReduction := player.ReduceCastingSkill(reduceBy)
            if actualReduction != reduceBy || player.ComputeCastingSkill() != initialSkill - reduceBy {
                test.Errorf("Skill reduction from %d by %d doesn't work: reduced by %d, actual skill after reduction is %d", initialSkill, reduceBy, actualReduction, player.ComputeCastingSkill())
            }
        }
    }

    // Repeatedly reducing the skill
    for reduceBy := 1; reduceBy <= 100; reduceBy++ {
        setPlayerInvestedPower(player, 5050) // 5050 is a sum of all the reductions
        actualReduction := player.ReduceCastingSkill(reduceBy)
        if actualReduction != reduceBy {
            test.Errorf("Cumulative skill reduction by %d doesn't work: reduced by %d, actual skill after reduction is %d", reduceBy, actualReduction, player.ComputeCastingSkill())
        }
    }

    // Trying to reduce more than the player has
    setPlayerInvestedPower(player, 5)
    actualReduction := player.ReduceCastingSkill(7)
    if actualReduction != 5 {
        test.Errorf("Reduction of 5 by 7 doesn't work: reduced by %d, actual skill after reduction is %d", actualReduction, player.ComputeCastingSkill())
    }
}
