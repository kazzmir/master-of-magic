package combat

import (
    "github.com/kazzmir/master-of-magic/game/magic/units"
)

type Army struct {
    Units []*units.Unit
}

type CombatScreen struct {
    DefendingArmy *Army
    AttackingArmy *Army
}
