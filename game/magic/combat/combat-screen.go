package combat

import (
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/hajimehoshi/ebiten/v2"
)

type Army struct {
    Units []*units.Unit
}

type CombatScreen struct {
    DefendingArmy *Army
    AttackingArmy *Army
}

func NewCombatScreen(defendingArmy *Army, attackingArmy *Army) *CombatScreen {
    return &CombatScreen{
        DefendingArmy: defendingArmy,
        AttackingArmy: attackingArmy,
    }
}

func (combat *CombatScreen) Update() error {
    return nil
}

func (combat *CombatScreen) Draw(screen *ebiten.Image) error {
    return nil
}
