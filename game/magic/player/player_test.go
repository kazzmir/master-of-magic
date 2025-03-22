package player

import (
    "testing"
    "math"

    "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/data"
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

func TestResearchPoints(test *testing.T) {
    natureSpell := spellbook.Spell{
        ResearchCost: 200,
        Magic: data.NatureMagic,
    }

    // 10% from nature magic books, 25% from sage master
    wizard := setup.WizardCustom{
        Books: []data.WizardBook{
            data.WizardBook{
                Magic: data.NatureMagic,
                Count: 8,
            },
        },
        Retorts: []data.Retort{
            data.RetortSageMaster,
        },
    }

    points := 15.0

    if computeEffectiveResearchPerTurn(&wizard, points, natureSpell) != int(points * (1 + 0.1 + 0.25)) {
        test.Errorf("Research points computation doesn't work")
    }
}

func TestSpellCost(test *testing.T) {
    natureSpell := spellbook.Spell{
        CastCost: 50,
        Magic: data.NatureMagic,
        Section: spellbook.SectionSpecial,
    }

    wizard := setup.WizardCustom{
        Books: []data.WizardBook{
            data.WizardBook{
                Magic: data.NatureMagic,
                Count: 1,
            },
        },
        Retorts: []data.Retort{
        },
    }

    value := computeEffectiveSpellCost(&wizard, natureSpell, true, false)
    if value != 50*5 {
        test.Errorf("casting cost for nature spell with no modifiers was wrong. expected=%v actual=%v", 50*5, value)
    }

    // 30% reduction
    wizard = setup.WizardCustom{
        Books: []data.WizardBook{
            data.WizardBook{
                Magic: data.NatureMagic,
                Count: 10,
            },
        },
        Retorts: []data.Retort{
        },
    }

    value = computeEffectiveSpellCost(&wizard, natureSpell, true, false)
    if value != 50*5*7/10 {
        test.Errorf("casting cost for nature spell was wrong. expected=%v actual=%v", 50*5*7/10, value)
    }

    // 30% reduction + nature mastery
    wizard = setup.WizardCustom{
        Books: []data.WizardBook{
            data.WizardBook{
                Magic: data.NatureMagic,
                Count: 10,
            },
        },
        Retorts: []data.Retort{
            data.RetortNatureMastery,
        },
    }

    value = computeEffectiveSpellCost(&wizard, natureSpell, true, false)
    expected := 50*5*(100 - (30 + 15))/100
    if value != expected {
        test.Errorf("casting cost for nature spell was wrong. expected=%v actual=%v", expected, value)
    }

    // evil omens
    value = computeEffectiveSpellCost(&wizard, natureSpell, true, true)
    expected = int(math.Floor(50.0*5*(100 - (30 + 15))/100 * 3/2))
    if value != expected {
        test.Errorf("casting cost for nature spell with evil omens. expected=%v actual=%v", expected, value)
    }

}
