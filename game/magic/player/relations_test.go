package player

import (
    "testing"

    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

func TestStartingRelation(test *testing.T) {
    ariel := setup.WizardCustom{
        Books: []data.WizardBook{
            data.WizardBook{
                Magic: data.LifeMagic,
                Count: 10,
            },
        },
    }

    tauron := setup.WizardCustom{
        Books: []data.WizardBook{
            data.WizardBook{
                Magic: data.ChaosMagic,
                Count: 10,
            },
        },
    }

    tlaloc := setup.WizardCustom{
        Books: []data.WizardBook{
            data.WizardBook{
                Magic: data.NatureMagic,
                Count: 4,
            },
            data.WizardBook{
                Magic: data.DeathMagic,
                Count: 5,
            },
        },
    }

    relation := computeStartingRelation(ariel, tauron)
    if relation != -48 {
        test.Errorf("expected -48, got %d", relation)
    }

    relation = computeStartingRelation(ariel, tlaloc)
    if relation != -21 {
        test.Errorf("expected -21, got %d", relation)
    }

    relation = computeStartingRelation(tauron, tlaloc)
    if relation != -15 {
        test.Errorf("expected -15, got %d", relation)
    }
}
