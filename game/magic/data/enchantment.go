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
    UnitEnchantmentChaosChannelsDemonSkin
    UnitEnchantmentChaosChannelsFireBreath
    UnitEnchantmentEndurance
    UnitEnchantmentHeroism
    UnitEnchantmentHolyArmor
    UnitEnchantmentHolyWeapon
    UnitEnchantmentInvulnerability
    UnitEnchantmentPlanarTravel
    UnitEnchantmentIronSkin
    UnitEnchantmentRegeneration
    UnitEnchantmentStoneSkin
    UnitEnchantmentWaterWalking
    UnitEnchantmentGuardianWind
    UnitEnchantmentInvisibility
    UnitEnchantmentMagicImmunity
    UnitEnchantmentSpellLock
    UnitEnchantmentWindWalking
    UnitEnchantmentEldritchWeapon
    UnitEnchantmentFlameBlade
    UnitEnchantmentBerserk
    UnitEnchantmentBlackChannels
    UnitEnchantmentWraithForm

    // curses
    UnitEnchantmentVertigo
    UnitEnchantmentShatter
)

var natureColor = color.RGBA{R: 0, G: 180, B: 0, A: 255}
var chaosColor = color.RGBA{R: 180, G: 0, B: 0, A: 255}
var sorceryColor = color.RGBA{R: 0, G: 0, B: 180, A: 255}
var deathColor = color.RGBA{R: 0x62, G: 0x11, B: 0xba, A: 255}
var lifeColor = color.RGBA{R: 180, G: 180, B: 180, A: 255}
var arcaneColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}

func GetMagicColor(magic MagicType) color.RGBA {
    switch magic {
        case NatureMagic: return natureColor
        case ChaosMagic: return chaosColor
        case SorceryMagic: return sorceryColor
        case DeathMagic: return deathColor
        case LifeMagic: return lifeColor
        case ArcaneMagic: return arcaneColor
    }

    return color.RGBA{}
}

func (enchantment UnitEnchantment) Color() color.Color {
    return GetMagicColor(enchantment.Magic())
}

// the magic realm this enchantment belongs to
func (enchantment UnitEnchantment) Magic() MagicType {
    switch enchantment {
        case UnitEnchantmentIronSkin: return NatureMagic
        case UnitEnchantmentGiantStrength: return NatureMagic
        case UnitEnchantmentResistElements: return NatureMagic
        case UnitEnchantmentElementalArmor: return NatureMagic
        case UnitEnchantmentPathFinding: return NatureMagic
        case UnitEnchantmentRegeneration: return NatureMagic
        case UnitEnchantmentStoneSkin: return NatureMagic
        case UnitEnchantmentWaterWalking: return NatureMagic

        case UnitEnchantmentFlight: return SorceryMagic
        case UnitEnchantmentHaste: return SorceryMagic
        case UnitEnchantmentResistMagic: return SorceryMagic
        case UnitEnchantmentGuardianWind: return SorceryMagic
        case UnitEnchantmentInvisibility: return SorceryMagic
        case UnitEnchantmentMagicImmunity: return SorceryMagic
        case UnitEnchantmentSpellLock: return SorceryMagic
        case UnitEnchantmentWindWalking: return SorceryMagic
        case UnitEnchantmentVertigo: return SorceryMagic

        case UnitEnchantmentCloakOfFear: return DeathMagic
        case UnitEnchantmentBerserk: return DeathMagic
        case UnitEnchantmentBlackChannels: return DeathMagic
        case UnitEnchantmentWraithForm: return DeathMagic

        case UnitEnchantmentImmolation: return ChaosMagic
        case UnitEnchantmentChaosChannelsDemonWings: return ChaosMagic
        case UnitEnchantmentChaosChannelsDemonSkin: return ChaosMagic
        case UnitEnchantmentChaosChannelsFireBreath: return ChaosMagic
        case UnitEnchantmentEldritchWeapon: return ChaosMagic
        case UnitEnchantmentFlameBlade: return ChaosMagic
        case UnitEnchantmentShatter: return ChaosMagic

        case UnitEnchantmentBless: return LifeMagic
        case UnitEnchantmentLionHeart: return LifeMagic
        case UnitEnchantmentEndurance: return LifeMagic
        case UnitEnchantmentHeroism: return LifeMagic
        case UnitEnchantmentTrueSight: return LifeMagic
        case UnitEnchantmentHolyArmor: return LifeMagic
        case UnitEnchantmentRighteousness: return LifeMagic
        case UnitEnchantmentHolyWeapon: return LifeMagic
        case UnitEnchantmentInvulnerability: return LifeMagic
        case UnitEnchantmentPlanarTravel: return LifeMagic
    }

    return MagicNone
}

// granted abilities, if any
func (enchantment UnitEnchantment) Abilities() []Ability {
    switch enchantment {
        case UnitEnchantmentImmolation: return []Ability{MakeAbility(AbilityImmolation)}
        case UnitEnchantmentCloakOfFear: return []Ability{MakeAbility(AbilityCauseFear)}
        case UnitEnchantmentTrueSight: return []Ability{MakeAbility(AbilityIllusionsImmunity)}
        case UnitEnchantmentPathFinding: return []Ability{MakeAbility(AbilityPathfinding)}
        case UnitEnchantmentChaosChannelsFireBreath: return []Ability{MakeAbilityValue(AbilityFireBreath, 2)}
        case UnitEnchantmentInvulnerability: return []Ability{MakeAbility(AbilityWeaponImmunity)}
        case UnitEnchantmentPlanarTravel: return []Ability{MakeAbility(AbilityPlaneShift)}
        case UnitEnchantmentRegeneration: return []Ability{MakeAbility(AbilityRegeneration)}
        case UnitEnchantmentGuardianWind: return []Ability{MakeAbility(AbilityMissileImmunity)}
        case UnitEnchantmentInvisibility: return []Ability{MakeAbility(AbilityInvisibility)}
        case UnitEnchantmentMagicImmunity: return []Ability{MakeAbility(AbilityMagicImmunity)}
        case UnitEnchantmentWindWalking: return []Ability{MakeAbility(AbilityWindWalking)}
        case UnitEnchantmentWraithForm: return []Ability{MakeAbility(AbilityWeaponImmunity), MakeAbility(AbilityNonCorporeal)}
    }

    return nil
}

func (enchantment UnitEnchantment) UpkeepMana() int {
    switch enchantment {
        case UnitEnchantmentGiantStrength: return 1
        case UnitEnchantmentLionHeart: return 4
        // combat only
        case UnitEnchantmentHaste: return 0
        case UnitEnchantmentImmolation: return 2
        case UnitEnchantmentResistElements: return 1
        case UnitEnchantmentResistMagic: return 1
        case UnitEnchantmentElementalArmor: return 5
        case UnitEnchantmentBless: return 1
        case UnitEnchantmentRighteousness: return 2
        case UnitEnchantmentCloakOfFear: return 1
        case UnitEnchantmentTrueSight: return 2
        case UnitEnchantmentFlight: return 3
        case UnitEnchantmentChaosChannelsDemonWings: return 0
        case UnitEnchantmentChaosChannelsDemonSkin: return 0
        case UnitEnchantmentChaosChannelsFireBreath: return 0
        case UnitEnchantmentEndurance: return 1
        case UnitEnchantmentHeroism: return 2
        case UnitEnchantmentHolyArmor: return 2
        case UnitEnchantmentHolyWeapon: return 1
        case UnitEnchantmentInvulnerability: return 5
        case UnitEnchantmentPlanarTravel: return 5
        case UnitEnchantmentIronSkin: return 5
        case UnitEnchantmentPathFinding: return 1
        case UnitEnchantmentRegeneration: return 10
        case UnitEnchantmentStoneSkin: return 1
        case UnitEnchantmentWaterWalking: return 1
        case UnitEnchantmentGuardianWind: return 2
        case UnitEnchantmentInvisibility: return 10
        case UnitEnchantmentMagicImmunity: return 5
        case UnitEnchantmentSpellLock: return 1
        case UnitEnchantmentWindWalking: return 10
        case UnitEnchantmentEldritchWeapon: return 1
        case UnitEnchantmentFlameBlade: return 2
        // combat only
        case UnitEnchantmentBerserk: return 0
        case UnitEnchantmentBlackChannels: return 1
        case UnitEnchantmentWraithForm: return 3
    }

    return 0
}

// the equivalent spell that would have cast this spell
func (enchantment UnitEnchantment) SpellName() string {
    switch enchantment {
        case UnitEnchantmentChaosChannelsDemonWings,
             UnitEnchantmentChaosChannelsDemonSkin,
             UnitEnchantmentChaosChannelsFireBreath: return "Chaos Channels"
        default: return enchantment.Name()
    }
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
        case UnitEnchantmentChaosChannelsDemonSkin: return "Demon Skin"
        case UnitEnchantmentChaosChannelsFireBreath: return "Fire Breath"
        case UnitEnchantmentEndurance: return "Endurance"
        case UnitEnchantmentHeroism: return "Heroism"
        case UnitEnchantmentHolyArmor: return "Holy Armor"
        case UnitEnchantmentHolyWeapon: return "Holy Weapon"
        case UnitEnchantmentInvulnerability: return "Invulnerability"
        case UnitEnchantmentPlanarTravel: return "Planar Travel"
        case UnitEnchantmentIronSkin: return "Iron Skin"
        case UnitEnchantmentPathFinding: return "Path Finding"
        case UnitEnchantmentRegeneration: return "Regeneration"
        case UnitEnchantmentStoneSkin: return "Stone Skin"
        case UnitEnchantmentWaterWalking: return "Water Walking"
        case UnitEnchantmentGuardianWind: return "Guardian Wind"
        case UnitEnchantmentInvisibility: return "Invisibility"
        case UnitEnchantmentMagicImmunity: return "Magic Immunity"
        case UnitEnchantmentSpellLock: return "Spell Lock"
        case UnitEnchantmentWindWalking: return "Wind Walking"
        case UnitEnchantmentEldritchWeapon: return "Eldritch Weapon"
        case UnitEnchantmentFlameBlade: return "Flame Blade"
        case UnitEnchantmentBerserk: return "Berserk"
        case UnitEnchantmentBlackChannels: return "Black Channels"
        case UnitEnchantmentWraithForm: return "Wraith Form"

        case UnitEnchantmentVertigo: return "Vertigo"
        case UnitEnchantmentShatter: return "Shatter"
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
        case UnitEnchantmentChaosChannelsDemonSkin: return "special.lbx"
        case UnitEnchantmentChaosChannelsFireBreath: return "special.lbx"
        case UnitEnchantmentEndurance: return "special.lbx"
        case UnitEnchantmentHeroism: return "special.lbx"
        case UnitEnchantmentHolyArmor: return "special.lbx"
        case UnitEnchantmentHolyWeapon: return "special.lbx"
        case UnitEnchantmentInvulnerability: return "special.lbx"
        case UnitEnchantmentPlanarTravel: return "special.lbx"
        case UnitEnchantmentIronSkin: return "special.lbx"
        case UnitEnchantmentPathFinding: return "special.lbx"
        case UnitEnchantmentRegeneration: return "special.lbx"
        case UnitEnchantmentStoneSkin: return "special.lbx"
        case UnitEnchantmentWaterWalking: return "special.lbx"
        case UnitEnchantmentGuardianWind: return "special2.lbx"
        case UnitEnchantmentInvisibility: return "special.lbx"
        case UnitEnchantmentMagicImmunity: return "special.lbx"
        case UnitEnchantmentSpellLock: return "special2.lbx"
        case UnitEnchantmentWindWalking: return "special.lbx"
        case UnitEnchantmentEldritchWeapon: return "special.lbx"
        case UnitEnchantmentFlameBlade: return "special.lbx"
        case UnitEnchantmentBerserk: return "special2.lbx"
        case UnitEnchantmentBlackChannels: return "special.lbx"
        case UnitEnchantmentWraithForm: return "special.lbx"
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
        case UnitEnchantmentChaosChannelsDemonSkin: return 62
        case UnitEnchantmentChaosChannelsDemonWings: return 63
        case UnitEnchantmentChaosChannelsFireBreath: return 64
        case UnitEnchantmentEndurance: return 76
        case UnitEnchantmentHeroism: return 87
        case UnitEnchantmentHolyArmor: return 92
        case UnitEnchantmentHolyWeapon: return 86
        case UnitEnchantmentInvulnerability: return 94
        case UnitEnchantmentPlanarTravel: return 91
        case UnitEnchantmentIronSkin: return 75
        case UnitEnchantmentPathFinding: return 70
        case UnitEnchantmentRegeneration: return 69
        case UnitEnchantmentStoneSkin: return 74
        case UnitEnchantmentWaterWalking: return 71
        case UnitEnchantmentGuardianWind: return 7
        case UnitEnchantmentInvisibility: return 78
        case UnitEnchantmentMagicImmunity: return 82
        case UnitEnchantmentSpellLock: return 8
        case UnitEnchantmentWindWalking: return 79
        case UnitEnchantmentEldritchWeapon: return 84
        case UnitEnchantmentFlameBlade: return 83
        case UnitEnchantmentBerserk: return 17
        case UnitEnchantmentBlackChannels: return 67
        case UnitEnchantmentWraithForm: return 68
    }

    return -1
}

// the index in specfx.lbx for when this enchantment is casted
func (enchantment UnitEnchantment) CastAnimationIndex() int {
    switch enchantment {
        // nature
        // FIXME: verify
        case UnitEnchantmentGiantStrength: return 0
        case UnitEnchantmentElementalArmor: return 0
        case UnitEnchantmentResistElements: return 0
        case UnitEnchantmentIronSkin: return 0
        case UnitEnchantmentPathFinding: return 0
        case UnitEnchantmentRegeneration: return 0
        case UnitEnchantmentStoneSkin: return 0
        case UnitEnchantmentWaterWalking: return 0

        // death
        case UnitEnchantmentCloakOfFear: return 4
        case UnitEnchantmentBerserk: return 4
        case UnitEnchantmentBlackChannels: return 4
        case UnitEnchantmentWraithForm: return 4

        // chaos
        case UnitEnchantmentImmolation: return 2
        case UnitEnchantmentChaosChannelsDemonSkin: return 2
        case UnitEnchantmentChaosChannelsDemonWings: return 2
        case UnitEnchantmentChaosChannelsFireBreath: return 2
        case UnitEnchantmentEldritchWeapon: return 2
        case UnitEnchantmentFlameBlade: return 2

        // sorcery
        case UnitEnchantmentHaste: return 1
        case UnitEnchantmentFlight: return 1
        case UnitEnchantmentResistMagic: return 1
        case UnitEnchantmentGuardianWind: return 1
        case UnitEnchantmentInvisibility: return 1
        case UnitEnchantmentMagicImmunity: return 1
        case UnitEnchantmentSpellLock: return 1
        case UnitEnchantmentWindWalking: return 1

        // life
        case UnitEnchantmentBless: return 3
        case UnitEnchantmentHeroism: return 3
        case UnitEnchantmentLionHeart: return 3
        case UnitEnchantmentEndurance: return 3
        case UnitEnchantmentTrueSight: return 3
        case UnitEnchantmentRighteousness: return 3
        case UnitEnchantmentHolyArmor: return 3
        case UnitEnchantmentHolyWeapon: return 3
        case UnitEnchantmentInvulnerability: return 3
        case UnitEnchantmentPlanarTravel: return 3
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
    CityEnchantmentWallOfStone
)

func (enchantment CityEnchantment) SpellName() string {
    switch enchantment {
        case CityEnchantmentLifeWard,
             CityEnchantmentSorceryWard,
             CityEnchantmentNatureWard,
             CityEnchantmentDeathWard,
             CityEnchantmentChaosWard: return "Spell Ward"
        default: return enchantment.Name()
    }
}

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
        case CityEnchantmentWallOfStone: return "Wall of Stone"
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
            CityEnchantmentProsperity, CityEnchantmentStreamOfLife, CityEnchantmentAstralGate:
            return 31
        // case CityEnchantmentChaosRift: return 0
        case CityEnchantmentCloudOfShadow, CityEnchantmentEvilPresence, CityEnchantmentFamine,
            CityEnchantmentPestilence, CityEnchantmentWallOfDarkness:
            return 32
        case CityEnchantmentConsecration: return 31
        case CityEnchantmentCursedLands: return 61
        case CityEnchantmentDarkRituals: return 60
        // case CityEnchantmentEarthGate: return 0
        // case CityEnchantmentFlyingFortress: return 0
        case CityEnchantmentGaiasBlessing, CityEnchantmentNaturesEye, CityEnchantmentWallOfStone:
            return 28
        // case CityEnchantmentLifeWard: return 0
        // case CityEnchantmentSorceryWard: return 0
        // case CityEnchantmentNatureWard: return 0
        // case CityEnchantmentDeathWard: return 0
        // case CityEnchantmentChaosWard: return 0
        case CityEnchantmentWallOfFire: return 30
    }

    return 0
}

// this is the offset in the city view of where to draw the enchantment sprite
// for enchantments that have 'normal' buildings associated with them (such as altar of battle), this is not needed
func (enchantment CityEnchantment) IconOffset() int {
    // FIXME: Add other offsets
    switch enchantment {
        // case CityEnchantmentCloudOfShadow: return 0
        case CityEnchantmentConsecration: return 177
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

type CombatEnchantment int

const (
    CombatEnchantmentNone CombatEnchantment = iota
    CombatEnchantmentHighPrayer
    CombatEnchantmentPrayer
    CombatEnchantmentTrueLight
    CombatEnchantmentCallLightning
    CombatEnchantmentEntangle
    CombatEnchantmentBlur
    CombatEnchantmentCounterMagic
    CombatEnchantmentMassInvisibility
    CombatEnchantmentMetalFires
    CombatEnchantmentWarpReality
    CombatEnchantmentBlackPrayer
    CombatEnchantmentDarkness
    CombatEnchantmentManaLeak
    CombatEnchantmentTerror
    CombatEnchantmentWrack
)

// in compix.lbx
func (enchantment CombatEnchantment) LbxIndex() int {
    switch enchantment {
        case CombatEnchantmentHighPrayer: return 12
        case CombatEnchantmentPrayer: return 11
        case CombatEnchantmentTrueLight: return 5
        case CombatEnchantmentCallLightning: return 14
        case CombatEnchantmentEntangle: return 60
        case CombatEnchantmentBlur: return 80
        case CombatEnchantmentCounterMagic: return 15
        case CombatEnchantmentMassInvisibility: return 41
        case CombatEnchantmentMetalFires: return 10
        case CombatEnchantmentWarpReality: return 7
        case CombatEnchantmentBlackPrayer: return 8
        case CombatEnchantmentDarkness: return 6
        case CombatEnchantmentManaLeak: return 79
        case CombatEnchantmentTerror: return 13
        case CombatEnchantmentWrack: return 9
    }

    return -1
}

func (enchantment CombatEnchantment) Magic() MagicType {
    switch enchantment {
        case CombatEnchantmentHighPrayer: return LifeMagic
        case CombatEnchantmentPrayer: return LifeMagic
        case CombatEnchantmentTrueLight: return LifeMagic
        case CombatEnchantmentCallLightning: return NatureMagic
        case CombatEnchantmentEntangle: return NatureMagic
        case CombatEnchantmentBlur: return SorceryMagic
        case CombatEnchantmentCounterMagic: return SorceryMagic
        case CombatEnchantmentMassInvisibility: return SorceryMagic
        case CombatEnchantmentMetalFires: return ChaosMagic
        case CombatEnchantmentWarpReality: return ChaosMagic
        case CombatEnchantmentBlackPrayer: return DeathMagic
        case CombatEnchantmentDarkness: return DeathMagic
        case CombatEnchantmentManaLeak: return DeathMagic
        case CombatEnchantmentTerror: return DeathMagic
        case CombatEnchantmentWrack: return DeathMagic
    }

    return MagicNone
}

func (enchantment CombatEnchantment) SpellName() string {
    return enchantment.Name()
}

func (enchantment CombatEnchantment) Name() string {
    switch enchantment {
        case CombatEnchantmentHighPrayer: return "High Prayer"
        case CombatEnchantmentPrayer: return "Prayer"
        case CombatEnchantmentTrueLight: return "True Light"
        case CombatEnchantmentCallLightning: return "Call Lightning"
        case CombatEnchantmentEntangle: return "Entangle"
        case CombatEnchantmentBlur: return "Blur"
        case CombatEnchantmentCounterMagic: return "Counter Magic"
        case CombatEnchantmentMassInvisibility: return "Mass Invisibility"
        case CombatEnchantmentMetalFires: return "Metal Fires"
        case CombatEnchantmentWarpReality: return "Warp Reality"
        case CombatEnchantmentBlackPrayer: return "Black Prayer"
        case CombatEnchantmentDarkness: return "Darkness"
        case CombatEnchantmentManaLeak: return "Mana Leak"
        case CombatEnchantmentTerror: return "Terror"
        case CombatEnchantmentWrack: return "Wrack"
    }

    return ""
}
