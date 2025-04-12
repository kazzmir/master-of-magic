package game

import (
    "math"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/combat"
)

// an implementation of combat/UnitDamage that wraps a StackUnit for the purpose of applying damage to a StackUnit
type UnitDamageWrapper struct {
    Unit units.StackUnit
}

func (wrapper *UnitDamageWrapper) GetMaxHealth() int {
    return wrapper.Unit.GetMaxHealth()
}

func (wrapper *UnitDamageWrapper) GetHealth() int {
    return wrapper.Unit.GetHealth()
}

func (wrapper *UnitDamageWrapper) GetCount() int {
    return wrapper.Unit.GetCount()
}

func (wrapper *UnitDamageWrapper) GetDefense() int {
    return wrapper.Unit.GetDefense()
}

func (wrapper *UnitDamageWrapper) HasAbility(ability data.AbilityType) bool {
    return wrapper.Unit.HasAbility(ability)
}

func (wrapper *UnitDamageWrapper) HasEnchantment(enchantment data.UnitEnchantment) bool {
    return wrapper.Unit.HasEnchantment(enchantment)
}

func (wrapper *UnitDamageWrapper) IsAsleep() bool {
    return false
}

func (wrapper *UnitDamageWrapper) TakeDamage(damage int, damageType combat.DamageType) int {
    wrapper.Unit.AdjustHealth(-damage)
    return 0
}

func (wrapper *UnitDamageWrapper) ToDefend(modifiers combat.DamageModifiers) int {
    return 30
}

func (wrapper *UnitDamageWrapper) ReduceInvulnerability(damage int) int {
    if wrapper.Unit.HasEnchantment(data.UnitEnchantmentInvulnerability) {
        return max(0, damage - 2)
    }

    return damage
}

func (wrapper *UnitDamageWrapper) Figures() int {
    health_per_figure := float64(wrapper.GetMaxHealth()) / float64(wrapper.GetCount())
    return int(math.Ceil(float64(wrapper.GetHealth()) / health_per_figure))
}
