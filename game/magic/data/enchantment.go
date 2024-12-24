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
    UnitEnchantmentResistElements
    UnitEnchantmentResistMagic
    UnitEnchantmentElementalArmor
    UnitEnchantmentBless
    UnitEnchantmentRighteousness
    UnitEnchantmentCloakOfFear
)

var natureColor = color.RGBA{R: 0, G: 180, B: 0, A: 255}
var chaosColor = color.RGBA{R: 180, G: 0, B: 0, A: 255}
var sorceryColor = color.RGBA{R: 0, G: 0, B: 180, A: 255}
var deathColor = color.RGBA{R: 0x62, G: 0x11, B: 0xba, A: 255}
var lifeColor = color.RGBA{R: 180, G: 180, B: 180, A: 255}

/*
Endurance	Life
Heroism	Life
Holy Armor	Life
Holy Weapon	Life
Invulnerability	Life
Planar Travel	Life
True Sight	Life
Iron Skin	Nature
Path Finding	Nature
Regeneration	Nature
Stone Skin	Nature
Water Walking	Nature
Flight	Sorcery
Guardian Wind	Sorcery
Invisibility	Sorcery
Magic Immunity	Sorcery
Spell Lock	Sorcery
Wind Walking	Sorcery
Eldritch Weapon	Chaos
Flame Blade	Chaos
Berserk	Death
Black Channels	Death
Wraith Form	Death
 */

func (enchantment UnitEnchantment) Color() color.Color {
    switch enchantment {
        case UnitEnchantmentGiantStrength: return natureColor
        case UnitEnchantmentLionHeart: return lifeColor
        case UnitEnchantmentHaste: return sorceryColor
        case UnitEnchantmentImmolation: return chaosColor
        case UnitEnchantmentResistElements: return natureColor
        case UnitEnchantmentResistMagic: return sorceryColor
        case UnitEnchantmentElementalArmor: return natureColor
        case UnitEnchantmentBless: return lifeColor
        case UnitEnchantmentRighteousness: return lifeColor
        case UnitEnchantmentCloakOfFear: return deathColor
    }

    return color.RGBA{R: 0, G: 0, B: 0, A: 0}
}

func (enchantment UnitEnchantment) Name() string {
    switch enchantment {
        case UnitEnchantmentGiantStrength: return "Giant Strength"
        case UnitEnchantmentLionHeart: return "Lion Heart"
        case UnitEnchantmentHaste: return "Haste"
        case UnitEnchantmentImmolation: return "Immolation"
        case UnitEnchantmentResistElements: return "Resist Elements"
        case UnitEnchantmentResistMagic: return "Resist Magic"
        case UnitEnchantmentElementalArmor: return "Elemental Armor"
        case UnitEnchantmentBless: return "Bless"
        case UnitEnchantmentRighteousness: return "Righteousness"
        case UnitEnchantmentCloakOfFear: return "Cloak of Fear"
    }

    return ""
}

func (enchantment UnitEnchantment) LbxFile() string {
    switch enchantment {
        case UnitEnchantmentGiantStrength: return "special.lbx"
        case UnitEnchantmentLionHeart: return "special.lbx"
        case UnitEnchantmentHaste: return "special.lbx"
        case UnitEnchantmentImmolation: return "special.lbx"
        case UnitEnchantmentResistElements: return "special.lbx"
        case UnitEnchantmentResistMagic: return "special.lbx"
        case UnitEnchantmentElementalArmor: return "special.lbx"
        case UnitEnchantmentBless: return "special.lbx"
        case UnitEnchantmentRighteousness: return "special.lbx"
        case UnitEnchantmentCloakOfFear: return "special2.lbx"
    }

    return ""
}

func (enchantment UnitEnchantment) LbxIndex() int {
    switch enchantment {
        case UnitEnchantmentGiantStrength: return 65
        case UnitEnchantmentLionHeart: return 89
        case UnitEnchantmentHaste: return 77
        case UnitEnchantmentImmolation: return 32
        case UnitEnchantmentResistElements: return 72
        case UnitEnchantmentResistMagic: return 81
        case UnitEnchantmentElementalArmor: return 73
        case UnitEnchantmentBless: return 88
        case UnitEnchantmentRighteousness: return 93
        case UnitEnchantmentCloakOfFear: return 21
    }

    return -1
}
