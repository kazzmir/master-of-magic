package player

import (
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

// compute an integer value that represents how much the two wizards like each other
// https://masterofmagic.fandom.com/wiki/Starting_Relation
func computeStartingRelation(wizard1 setup.WizardCustom, wizard2 setup.WizardCustom) int {
    // 1.50 patch formula

    wizard1Books := make(map[data.MagicType]int)
    wizard2Books := make(map[data.MagicType]int)

    for _, book := range wizard1.Books {
        wizard1Books[book.Magic] = book.Count
    }

    for _, book := range wizard2.Books {
        wizard2Books[book.Magic] = book.Count
    }

    sharedBooks := 0

    allMagic := []data.MagicType{data.LifeMagic, data.SorceryMagic, data.NatureMagic, data.DeathMagic, data.ChaosMagic}
    wizard1Alignment := 0
    wizard2Alignemnt := 0

    for _, magic := range allMagic {
        if wizard1Books[magic] > 0 && wizard2Books[magic] > 0 {
            sharedBooks += 1
        }

        switch magic {
            case data.LifeMagic, data.NatureMagic:
                wizard1Alignment += wizard1Books[magic]
                wizard2Alignemnt += wizard2Books[magic]
            case data.ChaosMagic, data.DeathMagic:
                wizard1Alignment -= wizard1Books[magic]
                wizard2Alignemnt -= wizard2Books[magic]
        }
    }

    abs := func (x int) int {
        if x < 0 {
            return -x
        }

        return x
    }

    return 2 * sharedBooks - 3 * (abs(wizard1Alignment - wizard2Alignemnt) - 4)
}
