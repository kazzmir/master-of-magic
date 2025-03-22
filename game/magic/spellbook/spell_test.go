package spellbook

import (
    "testing"
    "math"
    "slices"

    "github.com/kazzmir/master-of-magic/game/magic/data"
)

type testWizard struct {
    Books []data.WizardBook
    Retorts []data.Retort
}

func (wizard *testWizard) RetortEnabled(retort data.Retort) bool {
    return slices.Contains(wizard.Retorts, retort)
}

func (wizard *testWizard) MagicLevel(magic data.MagicType) int {
    for _, book := range wizard.Books {
        if book.Magic == magic {
            return book.Count
        }
    }

    return 0
}

func TestSpellCost(test *testing.T) {
    natureSpell := Spell{
        CastCost: 50,
        Magic: data.NatureMagic,
        Section: SectionSpecial,
    }

    wizard := testWizard{
        Books: []data.WizardBook{
            data.WizardBook{
                Magic: data.NatureMagic,
                Count: 1,
            },
        },
        Retorts: []data.Retort{
        },
    }

    value := ComputeSpellCost(&wizard, natureSpell, true, false)
    if value != 50*5 {
        test.Errorf("casting cost for nature spell with no modifiers was wrong. expected=%v actual=%v", 50*5, value)
    }

    // 30% reduction
    wizard = testWizard{
        Books: []data.WizardBook{
            data.WizardBook{
                Magic: data.NatureMagic,
                Count: 10,
            },
        },
        Retorts: []data.Retort{
        },
    }

    value = ComputeSpellCost(&wizard, natureSpell, true, false)
    if value != 50*5*7/10 {
        test.Errorf("casting cost for nature spell was wrong. expected=%v actual=%v", 50*5*7/10, value)
    }

    // 30% reduction + nature mastery
    wizard = testWizard{
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

    value = ComputeSpellCost(&wizard, natureSpell, true, false)
    expected := 50*5*(100 - (30 + 15))/100
    if value != expected {
        test.Errorf("casting cost for nature spell was wrong. expected=%v actual=%v", expected, value)
    }

    // evil omens
    value = ComputeSpellCost(&wizard, natureSpell, true, true)
    expected = int(math.Floor(50.0*5*(100 - (30 + 15))/100 * 3/2))
    if value != expected {
        test.Errorf("casting cost for nature spell with evil omens. expected=%v actual=%v", expected, value)
    }
}
