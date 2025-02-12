package data

import (
    "image/color"
)

// global overland enchantments
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

func (enchantment Enchantment) String() string {
    switch enchantment {
        case EnchantmentAwareness: return "Awareness"
        case EnchantmentDetectMagic: return "Detect Magic"
        case EnchantmentCharmOfLife: return "Charm of Life"
        case EnchantmentCrusade: return "Crusade"
        case EnchantmentHolyArms: return "Holy Arms"
        case EnchantmentJustCause: return "Just Cause"
        case EnchantmentLifeForce: return "Life Force"
        case EnchantmentPlanarSeal: return "Planar Seal"
        case EnchantmentTranquility: return "Tranquility"
        case EnchantmentHerbMastery: return "Herb Mastery"
        case EnchantmentNatureAwareness: return "Nature Awareness"
        case EnchantmentNaturesWrath: return "Nature's Wrath"
        case EnchantmentAuraOfMajesty: return "Aura of Majesty"
        case EnchantmentSuppressMagic: return "Suppress Magic"
        case EnchantmentTimeStop: return "Time Stop"
        case EnchantmentWindMastery: return "Wind Mastery"
        case EnchantmentArmageddon: return "Armageddon"
        case EnchantmentChaosSurge: return "Chaos Surge"
        case EnchantmentDoomMastery: return "Doom Mastery"
        case EnchantmentGreatWasting: return "Great Wasting"
        case EnchantmentMeteorStorm: return "Meteor Storm"
        case EnchantmentEternalNight: return "Eternal Night"
        case EnchantmentEvilOmens: return "Evil Omens"
        case EnchantmentZombieMastery: return "Zombie Mastery"
    }

    return ""
}

// how much mana per turn this enchantment costs
func (enchantment Enchantment) UpkeepMana() int {
    switch enchantment {
        case EnchantmentAwareness: return 3
        case EnchantmentDetectMagic: return 3
        case EnchantmentCharmOfLife: return 10
        case EnchantmentCrusade: return 10
        case EnchantmentHolyArms: return 10
        case EnchantmentJustCause: return 3
        case EnchantmentLifeForce: return 10
        case EnchantmentPlanarSeal: return 5
        case EnchantmentTranquility: return 10
        case EnchantmentHerbMastery: return 10
        case EnchantmentNatureAwareness: return 7
        case EnchantmentNaturesWrath: return 10
        case EnchantmentAuraOfMajesty: return 5
        case EnchantmentSuppressMagic: return 50
        case EnchantmentTimeStop: return 200
        case EnchantmentWindMastery: return 5
        case EnchantmentArmageddon: return 40
        case EnchantmentChaosSurge: return 40
        case EnchantmentDoomMastery: return 15
        case EnchantmentGreatWasting: return 20
        case EnchantmentMeteorStorm: return 10
        case EnchantmentEternalNight: return 15
        case EnchantmentEvilOmens: return 10
        case EnchantmentZombieMastery: return 40
    }

    return 0
}

// the index in specfx.lbx for when this enchantment is casted
func (enchantment Enchantment) LbxIndex() int {
    switch enchantment {
        case EnchantmentAwareness: return 56
        case EnchantmentDetectMagic: return 37
        case EnchantmentCharmOfLife: return 36
        case EnchantmentCrusade: return 32
        case EnchantmentHolyArms: return 34
        case EnchantmentJustCause: return 33
        case EnchantmentLifeForce: return 31
        case EnchantmentPlanarSeal: return 35
        case EnchantmentTranquility: return 30
        case EnchantmentHerbMastery: return 24
        case EnchantmentNatureAwareness: return 22
        case EnchantmentNaturesWrath: return 23
        case EnchantmentAuraOfMajesty: return 18
        case EnchantmentSuppressMagic: return 20
        case EnchantmentTimeStop: return 21
        case EnchantmentWindMastery: return 19
        case EnchantmentArmageddon: return 29
        case EnchantmentChaosSurge: return 25
        case EnchantmentDoomMastery: return 26
        case EnchantmentGreatWasting: return 27
        case EnchantmentMeteorStorm: return 28
        case EnchantmentEternalNight: return 15
        case EnchantmentEvilOmens: return 16
        case EnchantmentZombieMastery: return 17
    }

    return 0
}

// unit enchantments
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
    UnitEnchantmentTrueSight
    UnitEnchantmentPathFinding
    UnitEnchantmentFlight
    UnitEnchantmentChaosChannelsDemonWings
)

var natureColor = color.RGBA{R: 0, G: 180, B: 0, A: 255}
var chaosColor = color.RGBA{R: 180, G: 0, B: 0, A: 255}
var sorceryColor = color.RGBA{R: 0, G: 0, B: 180, A: 255}
var deathColor = color.RGBA{R: 0x62, G: 0x11, B: 0xba, A: 255}
var lifeColor = color.RGBA{R: 180, G: 180, B: 180, A: 255}

func GetMagicColor(magic MagicType) color.Color {
    switch magic {
        case NatureMagic: return natureColor
        case ChaosMagic: return chaosColor
        case SorceryMagic: return sorceryColor
        case DeathMagic: return deathColor
        case LifeMagic: return lifeColor
    }

    return color.RGBA{}
}

/*
Endurance	Life
Heroism	Life
Holy Armor	Life
Holy Weapon	Life
Invulnerability	Life
Planar Travel	Life
Iron Skin	Nature
Regeneration	Nature
Stone Skin	Nature
Water Walking	Nature
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
        case UnitEnchantmentTrueSight: return lifeColor
        case UnitEnchantmentFlight: return sorceryColor
        case UnitEnchantmentChaosChannelsDemonWings: return chaosColor
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
        case UnitEnchantmentTrueSight: return "True Sight"
        case UnitEnchantmentFlight: return "Flight"
        case UnitEnchantmentChaosChannelsDemonWings: return "Demon Wings"
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
        case UnitEnchantmentTrueSight: return "special.lbx"
        case UnitEnchantmentFlight: return "special.lbx"
        case UnitEnchantmentChaosChannelsDemonWings: return "special.lbx"
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
        case UnitEnchantmentTrueSight: return 85
        case UnitEnchantmentFlight: return 80
        case UnitEnchantmentChaosChannelsDemonWings: return 63
    }

    return -1
}

// city enchantments (also called Town Enchantments)
type CityEnchantment int
const (
    CityEnchantmentNone CityEnchantment = iota
    CityEnchantmentAltarOfBattle
    CityEnchantmentAstralGate
    CityEnchantmentChaosRift
    CityEnchantmentCloudOfShadow
    CityEnchantmentConsecration
    CityEnchantmentCursedLands
    CityEnchantmentDarkRituals
    CityEnchantmentEarthGate
    CityEnchantmentEvilPresence
    CityEnchantmentFamine
    CityEnchantmentFlyingFortress
    CityEnchantmentGaiasBlessing
    CityEnchantmentHeavenlyLight
    CityEnchantmentInspirations
    CityEnchantmentNaturesEye
    CityEnchantmentPestilence
    CityEnchantmentProsperity
    CityEnchantmentLifeWard
    CityEnchantmentSorceryWard
    CityEnchantmentNatureWard
    CityEnchantmentDeathWard
    CityEnchantmentChaosWard
    CityEnchantmentStreamOfLife
    CityEnchantmentWallOfDarkness
    CityEnchantmentWallOfFire
)

func (enchantment CityEnchantment) Name() string {
    switch enchantment {
        case CityEnchantmentAltarOfBattle: return "Altar of Battle"
        case CityEnchantmentAstralGate: return "Astral Gate"
        case CityEnchantmentChaosRift: return "Chaos Rift"
        case CityEnchantmentCloudOfShadow: return "Cloud of Shadow"
        case CityEnchantmentConsecration: return "Consecration"
        case CityEnchantmentCursedLands: return "Cursed Lands"
        case CityEnchantmentDarkRituals: return "Dark Rituals"
        case CityEnchantmentEarthGate: return "Earth Gate"
        case CityEnchantmentEvilPresence: return "Evil Presence"
        case CityEnchantmentFamine: return "Famine"
        case CityEnchantmentFlyingFortress: return "Flying Fortress"
        case CityEnchantmentGaiasBlessing: return "Gaia's Blessing"
        case CityEnchantmentHeavenlyLight: return "Heavenly Light"
        case CityEnchantmentInspirations: return "Inspirations"
        case CityEnchantmentNaturesEye: return "Nature's Eye"
        case CityEnchantmentPestilence: return "Pestilence"
        case CityEnchantmentProsperity: return "Prosperity"
        case CityEnchantmentLifeWard: return "Life Ward"
        case CityEnchantmentSorceryWard: return "Sorcery Ward"
        case CityEnchantmentNatureWard: return "Nature Ward"
        case CityEnchantmentDeathWard: return "Death Ward"
        case CityEnchantmentChaosWard: return "Chaos Ward"
        case CityEnchantmentStreamOfLife: return "Stream of Life"
        case CityEnchantmentWallOfDarkness: return "Wall of Darkness"
        case CityEnchantmentWallOfFire: return "Wall of Fire"
    }

    return ""
}

func (enchantment CityEnchantment) UpkeepMana() int {
    switch enchantment {
        case CityEnchantmentAltarOfBattle: return 5
        case CityEnchantmentAstralGate: return 5
        case CityEnchantmentChaosRift: return 10
        case CityEnchantmentCloudOfShadow: return 3
        case CityEnchantmentConsecration: return 8
        case CityEnchantmentCursedLands: return 2
        case CityEnchantmentDarkRituals: return 0
        case CityEnchantmentEarthGate: return 5
        case CityEnchantmentEvilPresence: return 4
        case CityEnchantmentFamine: return 5
        case CityEnchantmentFlyingFortress: return 25
        case CityEnchantmentGaiasBlessing: return 3
        case CityEnchantmentHeavenlyLight: return 2
        case CityEnchantmentInspirations: return 2
        case CityEnchantmentNaturesEye: return 1
        case CityEnchantmentPestilence: return 5
        case CityEnchantmentProsperity: return 2
        case CityEnchantmentLifeWard: return 5
        case CityEnchantmentSorceryWard: return 5
        case CityEnchantmentNatureWard: return 5
        case CityEnchantmentDeathWard: return 5
        case CityEnchantmentChaosWard: return 5
        case CityEnchantmentStreamOfLife: return 8
        case CityEnchantmentWallOfDarkness: return 5
        case CityEnchantmentWallOfFire: return 2
    }
    return 0
}

func (enchantment CityEnchantment) LbxIndex() int {
    switch enchantment {
        case CityEnchantmentConsecration: return 102
        case CityEnchantmentEvilPresence: return 82
        case CityEnchantmentInspirations: return 100
        case CityEnchantmentNaturesEye: return 99
        case CityEnchantmentProsperity: return 101
        case CityEnchantmentLifeWard: return 96
        case CityEnchantmentSorceryWard: return 97
        case CityEnchantmentNatureWard: return 98
        case CityEnchantmentDeathWard: return 94
        case CityEnchantmentChaosWard: return 95
        case CityEnchantmentWallOfDarkness: return 79
        case CityEnchantmentWallOfFire: return 77
    }

    return 0
}

func (enchantment CityEnchantment) SoundIndex() int {
    // FIXME: Add other sound indexes
    switch enchantment {
        case CityEnchantmentAltarOfBattle, CityEnchantmentHeavenlyLight, CityEnchantmentInspirations,
            CityEnchantmentProsperity, CityEnchantmentStreamOfLife:
            return 31
        // case CityEnchantmentAstralGate: return 0
        // case CityEnchantmentChaosRift: return 0
        case CityEnchantmentCloudOfShadow, CityEnchantmentEvilPresence, CityEnchantmentFamine,
            CityEnchantmentPestilence, CityEnchantmentWallOfDarkness:
            return 32
        // case CityEnchantmentConsecration: return 0
        case CityEnchantmentCursedLands: return 61
        case CityEnchantmentDarkRituals: return 60
        // case CityEnchantmentEarthGate: return 0
        // case CityEnchantmentFlyingFortress: return 0
        case CityEnchantmentGaiasBlessing, CityEnchantmentNaturesEye: return 28
        // case CityEnchantmentLifeWard: return 0
        // case CityEnchantmentSorceryWard: return 0
        // case CityEnchantmentNatureWard: return 0
        // case CityEnchantmentDeathWard: return 0
        // case CityEnchantmentChaosWard: return 0
        case CityEnchantmentWallOfFire: return 30
    }

    return 0
}

func (enchantment CityEnchantment) IconOffset() int {
    // FIXME: Add other offsets
    switch enchantment {
        // case CityEnchantmentCloudOfShadow: return 0
        // case CityEnchantmentConsecration: return 0
        case CityEnchantmentInspirations: return 134
        case CityEnchantmentNaturesEye: return 114
        case CityEnchantmentProsperity: return 153
        // case CityEnchantmentLifeWard: return 0
        // case CityEnchantmentSorceryWard: return 0
        // case CityEnchantmentNatureWard: return 0
        // case CityEnchantmentDeathWard: return 0
        // case CityEnchantmentChaosWard: return 0
    }

    return 0
}
