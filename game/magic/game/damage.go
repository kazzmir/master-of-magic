package game

import (
    "math"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/combat"
)

// an implementation of combat/UnitDamage that wraps a StackUnit for the purpose of applying damage to a StackUnit
type UnitDamageWrapper struct {
    units.StackUnit
}

func (wrapper *UnitDamageWrapper) IsAsleep() bool {
    return false
}

func (wrapper *UnitDamageWrapper) TakeDamage(damage int, damageType combat.DamageType) int {
    wrapper.StackUnit.AdjustHealth(-damage)
    return 0
}

func (wrapper *UnitDamageWrapper) ToDefend(modifiers combat.DamageModifiers) int {
    return 30
}

func (wrapper *UnitDamageWrapper) ReduceInvulnerability(damage int) int {
    if wrapper.StackUnit.HasEnchantment(data.UnitEnchantmentInvulnerability) {
        return max(0, damage - 2)
    }

    return damage
}

func (wrapper *UnitDamageWrapper) Figures() int {
    health_per_figure := float64(wrapper.GetMaxHealth()) / float64(wrapper.GetCount())
    return int(math.Ceil(float64(wrapper.GetHealth()) / health_per_figure))
}
