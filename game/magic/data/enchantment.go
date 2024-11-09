package data

import (
    "image/color"
)

type Enchantment int

const (
    EnchantmentNone Enchantment = iota
    EnchantmentAwareness
    EnchantmentDetectMagic
    EnchantmentCharmOfLife
    EnchantmentCrusade
    EnchantmentHolyArms
    EnchantmentJustCause
    EnchantmentLifeForce
    EnchantmentPlanarSeal
    EnchantmentTranquility
    EnchantmentHerbMastery
    EnchantmentNatureAwareness
    EnchantmentNaturesWrath
    EnchantmentAuraOfMajesty
    EnchantmentSuppressMagic
    EnchantmentTimeStop
    EnchantmentWindMastery
    EnchantmentArmageddon
    EnchantmentChaosSurge
    EnchantmentDoomMastery
    EnchantmentGreatWasting
    EnchantmentMeteorStorm
    EnchantmentEternalNight
    EnchantmentEvilOmens
    EnchantmentZombieMastery
)

type UnitEnchantment int

const (
    UnitEnchantmentNone UnitEnchantment = iota
    UnitEnchantmentGiantStrength
)

func (enchantment UnitEnchantment) Color() color.Color {
    switch enchantment {
        case UnitEnchantmentGiantStrength: return color.RGBA{R: 0, G: 180, B: 0, A: 255}
    }

    return color.RGBA{R: 0, G: 0, B: 0, A: 0}
}

func (enchantment UnitEnchantment) Name() string {
    switch enchantment {
        case UnitEnchantmentGiantStrength: return "Giant Strength"
    }

    return ""
}

func (enchantment UnitEnchantment) LbxFile() string {
    switch enchantment {
        case UnitEnchantmentGiantStrength: return "special.lbx"
    }

    return ""
}

func (enchantment UnitEnchantment) LbxIndex() int {
    switch enchantment {
        case UnitEnchantmentGiantStrength: return 65
    }

    return -1
}
