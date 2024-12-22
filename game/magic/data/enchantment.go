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
    UnitEnchantmentLionHeart
    UnitEnchantmentHaste
    UnitEnchantmentImmolation
)

var natureColor = color.RGBA{R: 0, G: 180, B: 0, A: 255}
var chaosColor = color.RGBA{R: 180, G: 0, B: 0, A: 255}
var sorceryColor = color.RGBA{R: 0, G: 0, B: 180, A: 255}
var deathColor = color.RGBA{R: 0x62, G: 0x11, B: 0xba, A: 255}
var lifeColor = color.RGBA{R: 180, G: 180, B: 180, A: 255}

/*
Bless	Life
Endurance	Life
Heroism	Life
Holy Armor	Life
Holy Weapon	Life
Invulnerability	Life
Planar Travel	Life
Righteousness	Life
True Sight	Life
Elemental Armor	Nature
Iron Skin	Nature
Path Finding	Nature
Regeneration	Nature
Resist Elements	Nature
Stone Skin	Nature
Water Walking	Nature
Flight	Sorcery
Guardian Wind	Sorcery
Invisibility	Sorcery
Magic Immunity	Sorcery
Resist Magic	Sorcery
Spell Lock	Sorcery
Wind Walking	Sorcery
Eldritch Weapon	Chaos
Flame Blade	Chaos
Berserk	Death
Black Channels	Death
Cloak of Fear	Death
Wraith Form	Death
 */

func (enchantment UnitEnchantment) Color() color.Color {
    switch enchantment {
        case UnitEnchantmentGiantStrength: return natureColor
        case UnitEnchantmentLionHeart: return lifeColor
        case UnitEnchantmentHaste: return sorceryColor
        case UnitEnchantmentImmolation: return chaosColor
    }

    return color.RGBA{R: 0, G: 0, B: 0, A: 0}
}

func (enchantment UnitEnchantment) Name() string {
    switch enchantment {
        case UnitEnchantmentGiantStrength: return "Giant Strength"
        case UnitEnchantmentLionHeart: return "Lion Heart"
        case UnitEnchantmentHaste: return "Haste"
        case UnitEnchantmentImmolation: return "Immolation"
    }

    return ""
}

func (enchantment UnitEnchantment) LbxFile() string {
    switch enchantment {
        case UnitEnchantmentGiantStrength: return "special.lbx"
        case UnitEnchantmentLionHeart: return "special.lbx"
        case UnitEnchantmentHaste: return "special.lbx"
        case UnitEnchantmentImmolation: return "special.lbx"
    }

    return ""
}

func (enchantment UnitEnchantment) LbxIndex() int {
    switch enchantment {
        case UnitEnchantmentGiantStrength: return 65
        case UnitEnchantmentLionHeart: return 89
        case UnitEnchantmentHaste: return 77
        case UnitEnchantmentImmolation: return 32
    }

    return -1
}
