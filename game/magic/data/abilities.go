package data

import (
    "fmt"
)

type AbilityType int

type Ability struct {
    Ability AbilityType
    Value float32
    // here to make this struct not ==
    not_equal []int
}

func MakeAbility(ability AbilityType) Ability {
    return Ability{Ability: ability}
}

func MakeAbilityValue(ability AbilityType, value float32) Ability {
    return Ability{Ability: ability, Value: value}
}

func romanNumeral(value int) string {
    switch value {
        case 1: return "I"
        case 2: return "II"
        case 3: return "III"
        case 4: return "IV"
        case 5: return "V"
        case 6: return "VI"
        case 7: return "VII"
        case 8: return "VIII"
        case 9: return "IX"
        case 10: return "X"
    }

    return fmt.Sprintf("%v", value)
}

const (
    AbilityNone AbilityType = iota
    // unit abilities
    AbilityArmorPiercing
    AbilityCauseFear
    AbilityColdImmunity
    AbilityConstruction
    AbilityCreateOutpost
    AbilityCreateUndead
    AbilityDeathGaze
    AbilityDeathImmunity
    AbilityDispelEvil
    AbilityDoomBoltSpell
    AbilityDoomGaze
    AbilityDeathTouch
    AbilityFireballSpell
    AbilityFireBreath
    AbilityFireImmunity
    AbilityFirstStrike
    AbilityForester
    AbilityHealer
    AbilityHealingSpell
    AbilityHolyBonus
    AbilityIllusion
    AbilityIllusionsImmunity
    AbilityImmolation
    AbilityInvisibility
    AbilityLargeShield
    AbilityLifeSteal
    AbilityLightningBreath
    AbilityLongRange
    AbilityMagicImmunity
    AbilityMeld
    AbilityMerging
    AbilityMissileImmunity
    AbilityMountaineer
    AbilityNegateFirstStrike
    AbilityNonCorporeal
    AbilityPathfinding
    AbilityPlaneShift
    AbilityPoisonImmunity
    AbilityPoisonTouch
    AbilityPurify
    AbilityRegeneration
    AbilityResistanceToAll
    AbilityScouting
    AbilityStoningGaze
    AbilityStoningImmunity
    AbilityStoningTouch
    AbilitySummonDemons
    AbilityToHit
    AbilityTransport
    AbilityTeleporting
    AbilityThrown
    AbilityWallCrusher
    AbilityWeaponImmunity
    AbilityWebSpell
    AbilityWindWalking

    // hero abilities
    AbilityAgility
    AbilitySuperAgility
    AbilityArcanePower
    AbilitySuperArcanePower
    AbilityArmsmaster
    AbilitySuperArmsmaster
    AbilityBlademaster
    AbilitySuperBlademaster
    AbilityCaster
    AbilityCharmed
    AbilityConstitution
    AbilitySuperConstitution
    AbilityLeadership
    AbilitySuperLeadership
    AbilityLegendary
    AbilitySuperLegendary
    AbilityLucky
    AbilityMight
    AbilitySuperMight
    AbilityNoble
    AbilityPrayermaster
    AbilitySuperPrayermaster
    AbilitySage
    AbilitySuperSage
)

// the file that stores the icon for this ability
func (ability Ability) LbxFile() string {
    switch ability.Ability {
        case AbilityArmorPiercing: return "special.lbx"
        case AbilityCauseFear: return "special2.lbx"
        case AbilityColdImmunity: return "special.lbx"
        case AbilityConstruction: return "special.lbx"
        case AbilityCreateOutpost: return "special.lbx"
        case AbilityCreateUndead: return "special.lbx"
        case AbilityDeathGaze: return "special2.lbx"
        case AbilityDeathImmunity: return "special.lbx"
        case AbilityDispelEvil: return "special2.lbx"
        case AbilityDoomBoltSpell: return "special.lbx"
        case AbilityDoomGaze: return "special.lbx"
        case AbilityDeathTouch: return "special2.lbx"
        case AbilityFireballSpell: return "special.lbx"
        case AbilityFireBreath: return "special2.lbx"
        case AbilityFireImmunity: return "special.lbx"
        case AbilityFirstStrike: return "special.lbx"
        case AbilityForester: return "special.lbx"
        case AbilityHealer: return "special.lbx"
        case AbilityHealingSpell: return "special.lbx"
        case AbilityHolyBonus: return "special.lbx"
        case AbilityIllusion: return "special.lbx"
        case AbilityIllusionsImmunity: return "special.lbx"
        case AbilityImmolation: return "special.lbx"
        case AbilityInvisibility: return "special.lbx"
        case AbilityLargeShield: return "special.lbx"
        case AbilityLifeSteal: return "special.lbx"
        case AbilityLightningBreath: return "special2.lbx"
        case AbilityLongRange: return "special2.lbx"
        case AbilityMagicImmunity: return "special.lbx"
        case AbilityMeld: return "special.lbx"
        case AbilityMerging: return "itemisc.lbx"
        case AbilityMissileImmunity: return "special.lbx"
        case AbilityMountaineer: return "special.lbx"
        case AbilityNegateFirstStrike: return "special.lbx"
        case AbilityNonCorporeal: return "special.lbx"
        case AbilityPathfinding: return "special.lbx"
        case AbilityPlaneShift: return "special.lbx"
        case AbilityPoisonImmunity: return "special.lbx"
        case AbilityPoisonTouch: return "special.lbx"
        case AbilityPurify: return "special.lbx"
        case AbilityRegeneration: return "special.lbx"
        case AbilityResistanceToAll: return "special.lbx"
        case AbilityScouting: return "special.lbx"
        case AbilityStoningGaze: return "special.lbx"
        case AbilityStoningImmunity: return "special.lbx"
        case AbilityStoningTouch: return "special.lbx"
        case AbilitySummonDemons: return "special2.lbx"
        case AbilityTeleporting: return "special.lbx"
        case AbilityThrown: return "special2.lbx"
        case AbilityToHit: return "special2.lbx"
        // FIXME: this is not the right tile for transport, so we just pick a blank tile
        case AbilityTransport: return "special.lbx"
        case AbilityWallCrusher: return "special.lbx"
        case AbilityWeaponImmunity: return "special.lbx"
        case AbilityWebSpell: return "special2.lbx"
        case AbilityWindWalking: return "special.lbx"
        case AbilityAgility, AbilitySuperAgility: return "special2.lbx"
        case AbilityArcanePower, AbilitySuperArcanePower: return "special.lbx"
        case AbilityArmsmaster, AbilitySuperArmsmaster: return "special.lbx"
        case AbilityBlademaster, AbilitySuperBlademaster: return "special.lbx"
        case AbilityCaster: return "special.lbx"
        case AbilityCharmed: return "special.lbx"
        case AbilityConstitution, AbilitySuperConstitution: return "special.lbx"
        case AbilityLeadership, AbilitySuperLeadership: return "special.lbx"
        case AbilityLegendary, AbilitySuperLegendary: return "special.lbx"
        case AbilityLucky: return "special.lbx"
        case AbilityMight, AbilitySuperMight: return "special.lbx"
        case AbilityNoble: return "special.lbx"
        case AbilityPrayermaster, AbilitySuperPrayermaster: return "special.lbx"
        case AbilitySage, AbilitySuperSage: return "special.lbx"
    }

    return ""

}

func (ability Ability) String() string {
    return ability.Name()
}

func (ability Ability) Name() string {
    plusValue := func(name string) string {
        if ability.Value > 0 {
            return fmt.Sprintf("%v +%v", name, int(ability.Value))
        } else {
            return name
        }
    }

    // for blademaster, 30% -> +3
    plusValue10 := func(name string) string {
        if ability.Value > 0 {
            return fmt.Sprintf("%v +%v", name, int(ability.Value/10))
        } else {
            return name
        }
    }

    switch ability.Ability {
        case AbilityArmorPiercing: return "Armor Piercing"
        case AbilityCauseFear: return "Cause Fear"
        case AbilityColdImmunity: return "Cold Immunity"
        case AbilityConstruction: return "Construction"
        case AbilityCreateOutpost: return "Create Outpost"
        case AbilityCreateUndead: return "Create Undead"
        case AbilityDeathGaze: return fmt.Sprintf("Death Gaze %v", int(ability.Value))
        case AbilityDeathImmunity: return "Death Immunity"
        case AbilityDispelEvil: return "Dispel Evil"
        case AbilityDoomBoltSpell: return "Doom Bolt Spell"
        case AbilityDoomGaze: return "Doom Gaze"
        case AbilityDeathTouch: return "Death Touch"
        case AbilityFireballSpell: return fmt.Sprintf("Fireball Spell x%v", int(ability.Value))
        case AbilityFireBreath: return fmt.Sprintf("Fire Breath %v", int(ability.Value))
        case AbilityFireImmunity: return "Fire Immunity"
        case AbilityFirstStrike: return "First Strike"
        case AbilityForester: return "Forester"
        case AbilityHealer: return "Healer"
        case AbilityHealingSpell: return "Healing Spell"
        case AbilityHolyBonus: return "Holy Bonus"
        case AbilityIllusion: return "Illusion"
        case AbilityIllusionsImmunity: return "Illusions Immunity"
        case AbilityImmolation: return "Immolation"
        case AbilityInvisibility: return "Invisibility"
        case AbilityLargeShield: return "Large Shield"
        case AbilityLifeSteal: return fmt.Sprintf("Life Steal %v", int(ability.Value))
        case AbilityLightningBreath: return fmt.Sprintf("Lightning Breath %v", int(ability.Value))
        case AbilityLongRange: return "Long Range"
        case AbilityMagicImmunity: return "Magic Immunity"
        case AbilityMeld: return "Meld"
        case AbilityMerging: return "Merging"
        case AbilityMissileImmunity: return "Missile Immunity"
        case AbilityMountaineer: return "Mountaineer"
        case AbilityNegateFirstStrike: return "Negate First Strike"
        case AbilityNonCorporeal: return "NonCorporeal"
        case AbilityPathfinding: return "Pathfinding"
        case AbilityPlaneShift: return "Plane Shift"
        case AbilityPoisonImmunity: return "Poison Immunity"
        case AbilityPoisonTouch: return "Poison Touch"
        case AbilityPurify: return "Purify"
        case AbilityRegeneration: return "Regeneration"
        case AbilityResistanceToAll: return "Resistance to all"
        case AbilityScouting: return fmt.Sprintf("Scouting %v", romanNumeral(int(ability.Value)))
        case AbilityStoningGaze: return fmt.Sprintf("Stoning Gaze %v", int(ability.Value))
        case AbilityStoningImmunity: return "Stoning Immunity"
        case AbilityStoningTouch: return fmt.Sprintf("Stoning Touch %v", int(ability.Value))
        case AbilitySummonDemons: return "Summon Demons"
        case AbilityTeleporting: return "Teleporting"
        case AbilityThrown: return fmt.Sprintf("Thrown %v", int(ability.Value))
        case AbilityToHit: return fmt.Sprintf("+%v To Hit", ability.Value/10)
        case AbilityTransport: return "Transport"
        case AbilityWallCrusher: return "Wall Crusher"
        case AbilityWeaponImmunity: return "Weapon Immunity"
        case AbilityWebSpell: return "Web Spell"
        case AbilityWindWalking: return "Wind Walking"
        case AbilityAgility: return "Agility"

        // FIXME: add +X to each ability that has a value
        case AbilitySuperAgility: return plusValue("Super Agility")
        case AbilityArcanePower: return plusValue("Arcane Power")
        case AbilitySuperArcanePower: return plusValue("Super Arcane Power")
        case AbilityArmsmaster: return plusValue("Armsmaster")
        case AbilitySuperArmsmaster: return plusValue("Super Armsmaster")
        case AbilityBlademaster: return plusValue10("Blademaster")
        case AbilitySuperBlademaster: return plusValue10("Super Blademaster")
        case AbilityCaster: return fmt.Sprintf("Caster %v mp", ability.Value)
        case AbilityCharmed: return "Charmed"
        case AbilityConstitution: return plusValue("Constitution")
        case AbilitySuperConstitution: return plusValue("Super Constitution")
        case AbilityLeadership: return plusValue("Leadership")
        case AbilitySuperLeadership: return plusValue("Super Leadership")
        case AbilityLegendary: return plusValue("Legendary")
        case AbilitySuperLegendary: return plusValue("Super Legendary")
        case AbilityLucky: return "Lucky"
        case AbilityMight: return plusValue("Might")
        case AbilitySuperMight: return plusValue("Super Might")
        case AbilityNoble: return "Noble"
        case AbilityPrayermaster: return plusValue("Prayermaster")
        case AbilitySuperPrayermaster: return plusValue("Super Prayermaster")
        case AbilitySage: return plusValue("Sage")
        case AbilitySuperSage: return plusValue("Super Sage")
    }

    return "?"
}

func (ability Ability) IsHeroAbility() bool {
    switch ability.Ability {
        case AbilitySuperAgility, AbilityArcanePower,
             AbilitySuperArcanePower, AbilityArmsmaster,
             AbilitySuperArmsmaster, AbilityBlademaster,
             AbilitySuperBlademaster, AbilityCaster,
             AbilityCharmed, AbilityConstitution,
             AbilitySuperConstitution, AbilityLeadership,
             AbilitySuperLeadership, AbilityLegendary,
             AbilitySuperLegendary, AbilityLucky,
             AbilityMight, AbilitySuperMight,
             AbilityNoble, AbilityPrayermaster,
             AbilitySuperPrayermaster, AbilitySage,
             AbilitySuperSage: return true
        default: return false
    }
}

// the index in the lbx file for this icon
func (ability Ability) LbxIndex() int {
    switch ability.Ability {
        case AbilityArmorPiercing: return 28
        case AbilityCauseFear: return 21
        case AbilityColdImmunity: return 11
        case AbilityConstruction: return 36
        case AbilityCreateOutpost: return 17
        case AbilityCreateUndead: return 19
        case AbilityDeathGaze: return 24
        case AbilityDeathImmunity: return 49
        case AbilityDispelEvil: return 22
        case AbilityDoomBoltSpell: return 41
        case AbilityDoomGaze: return 26
        case AbilityDeathTouch: return 30
        case AbilityFireballSpell: return 39
        case AbilityFireBreath: return 27
        case AbilityFireImmunity: return 6
        case AbilityFirstStrike: return 29
        case AbilityForester: return 1
        case AbilityHealer: return 16
        case AbilityHealingSpell: return 38
        case AbilityHolyBonus: return 34
        case AbilityIllusion: return 35
        case AbilityIllusionsImmunity: return 10
        case AbilityImmolation: return 32
        case AbilityInvisibility: return 18
        case AbilityLargeShield: return 14
        case AbilityLifeSteal: return 31
        case AbilityLightningBreath: return 26
        case AbilityLongRange: return 18
        case AbilityMagicImmunity: return 12
        case AbilityMeld: return 40
        case AbilityMerging: return 18
        case AbilityMissileImmunity: return 9
        case AbilityMountaineer: return 2
        case AbilityNegateFirstStrike: return 48
        case AbilityNonCorporeal: return 22
        case AbilityPathfinding: return 20
        case AbilityPlaneShift: return 4
        case AbilityPoisonImmunity: return 5
        case AbilityPoisonTouch: return 30
        case AbilityPurify: return 25
        case AbilityRegeneration: return 24
        case AbilityResistanceToAll: return 33
        case AbilityScouting: return 37
        case AbilityStoningGaze: return 26
        case AbilityStoningImmunity: return 7
        case AbilityStoningTouch: return 27
        case AbilitySummonDemons: return 28
        case AbilityTeleporting: return 0
        case AbilityThrown: return 19
        case AbilityToHit: return 14
        // transport just uses a blank tile
        case AbilityTransport: return 3
        case AbilityWallCrusher: return 15
        case AbilityWeaponImmunity: return 8
        case AbilityWebSpell: return 20
        case AbilityWindWalking: return 23
        case AbilityAgility, AbilitySuperAgility: return 32
        case AbilityArcanePower, AbilitySuperArcanePower: return 54
        case AbilityArmsmaster, AbilitySuperArmsmaster: return 46
        case AbilityBlademaster, AbilitySuperBlademaster: return 47
        case AbilityCaster: return 55
        case AbilityCharmed: return 59
        case AbilityConstitution, AbilitySuperConstitution: return 50
        case AbilityLeadership, AbilitySuperLeadership: return 43
        case AbilityLegendary, AbilitySuperLegendary: return 45
        case AbilityLucky: return 58
        case AbilityMight, AbilitySuperMight: return 52
        case AbilityNoble: return 60
        case AbilityPrayermaster, AbilitySuperPrayermaster: return 57
        case AbilitySage, AbilitySuperSage: return 61
    }

    return -1
}

func (item ItemAbility) MagicType() MagicType {
    switch item {
        case ItemAbilityWraithform, ItemAbilityVampiric,
             ItemAbilityDeath, ItemAbilityPowerDrain,
             ItemAbilityCloakOfFear: return DeathMagic

        case ItemAbilityGuardianWind, ItemAbilityHaste,
             ItemAbilityInvisibility, ItemAbilityFlight,
             ItemAbilityResistMagic, ItemAbilityMagicImmunity,
             ItemAbilityPhantasmal: return SorceryMagic

        case ItemAbilityWaterWalking, ItemAbilityRegeneration,
             ItemAbilityPathfinding, ItemAbilityMerging,
             ItemAbilityResistElements, ItemAbilityElementalArmor,
             ItemAbilityGiantStrength, ItemAbilityStoning: return NatureMagic

        case ItemAbilityHolyAvenger, ItemAbilityTrueSight,
             ItemAbilityBless, ItemAbilityRighteousness,
             ItemAbilityInvulnerability, ItemAbilityEndurance,
             ItemAbilityPlanarTravel, ItemAbilityLionHeart: return LifeMagic

        case ItemAbilityFlaming, ItemAbilityLightning,
             ItemAbilityChaos, ItemAbilityDestruction: return ChaosMagic
    }

    return MagicNone
}

type ItemAbility int
const (
    ItemAbilityNone ItemAbility = iota
    ItemAbilityVampiric
    ItemAbilityGuardianWind
    ItemAbilityLightning
    ItemAbilityCloakOfFear
    ItemAbilityDestruction
    ItemAbilityWraithform
    ItemAbilityRegeneration
    ItemAbilityPathfinding
    ItemAbilityWaterWalking
    ItemAbilityResistElements
    ItemAbilityElementalArmor
    ItemAbilityChaos
    ItemAbilityStoning
    ItemAbilityEndurance
    ItemAbilityHaste
    ItemAbilityInvisibility
    ItemAbilityDeath
    ItemAbilityFlight
    ItemAbilityResistMagic
    ItemAbilityMagicImmunity
    ItemAbilityFlaming
    ItemAbilityHolyAvenger
    ItemAbilityTrueSight
    ItemAbilityPhantasmal
    ItemAbilityPowerDrain
    ItemAbilityBless
    ItemAbilityLionHeart
    ItemAbilityGiantStrength
    ItemAbilityPlanarTravel
    ItemAbilityMerging
    ItemAbilityRighteousness
    ItemAbilityInvulnerability
)

func (item ItemAbility) AbilityType() AbilityType {
    switch item {
        case ItemAbilityCloakOfFear: return AbilityCauseFear
        case ItemAbilityLightning: return AbilityArmorPiercing
        case ItemAbilityHolyAvenger: return AbilityDispelEvil
        // FIXME: should return AbilityValue(StoningTouch, 1)
        case ItemAbilityStoning: return AbilityStoningTouch
        case ItemAbilityPhantasmal: return AbilityIllusion

    }

    return AbilityNone
}

func (item ItemAbility) Enchantment() UnitEnchantment {
    switch item {
        case ItemAbilityBless: return UnitEnchantmentBless
        case ItemAbilityHolyAvenger: return UnitEnchantmentBless
        case ItemAbilityTrueSight: return UnitEnchantmentTrueSight
        case ItemAbilityResistElements: return UnitEnchantmentResistElements
        case ItemAbilityElementalArmor: return UnitEnchantmentElementalArmor
        // giant strength stacks with the spell, so we need a way to count the number of times this enchantment is applied to a unit
        case ItemAbilityGiantStrength: return UnitEnchantmentGiantStrength
        case ItemAbilityGuardianWind: return UnitEnchantmentGuardianWind
        case ItemAbilityHaste: return UnitEnchantmentHaste
        case ItemAbilityResistMagic: return UnitEnchantmentResistMagic
        case ItemAbilityMagicImmunity: return UnitEnchantmentMagicImmunity
        case ItemAbilityWraithform: return UnitEnchantmentWraithForm
        case ItemAbilityRighteousness: return UnitEnchantmentRighteousness
        case ItemAbilityInvulnerability: return UnitEnchantmentInvulnerability
        case ItemAbilityEndurance: return UnitEnchantmentEndurance
        case ItemAbilityPlanarTravel: return UnitEnchantmentPlanarTravel
        case ItemAbilityLionHeart: return UnitEnchantmentLionHeart
        case ItemAbilityWaterWalking: return UnitEnchantmentWaterWalking
        case ItemAbilityRegeneration: return UnitEnchantmentRegeneration
        case ItemAbilityPathfinding: return UnitEnchantmentPathFinding
        case ItemAbilityFlight: return UnitEnchantmentFlight
        case ItemAbilityInvisibility: return UnitEnchantmentInvisibility
    }

    return UnitEnchantmentNone
}

func (item ItemAbility) Name() string {
    switch item {
        case ItemAbilityBless: return "Bless"
        case ItemAbilityChaos: return "Chaos"
        case ItemAbilityCloakOfFear: return "Cloak Of Fear"
        case ItemAbilityDeath: return "Death"
        case ItemAbilityElementalArmor: return "Elemental Armor"
        case ItemAbilityEndurance: return "Endurance"
        case ItemAbilityFlaming: return "Flaming"
        case ItemAbilityFlight: return "Flight"
        case ItemAbilityGiantStrength: return "Giant Strength"
        case ItemAbilityGuardianWind: return "Guardian Wind"
        case ItemAbilityHaste: return "Haste"
        case ItemAbilityInvulnerability: return "Invulnerability"
        case ItemAbilityLightning: return "Lightning"
        case ItemAbilityLionHeart: return "Lion Heart"
        case ItemAbilityPhantasmal: return "Phantasmal"
        case ItemAbilityPlanarTravel: return "Planar Travel"
        case ItemAbilityPowerDrain: return "Power Drain"
        case ItemAbilityResistElements: return "Resist Elements"
        case ItemAbilityResistMagic: return "Resist Magic"
        case ItemAbilityRighteousness: return "Righteousness"
        case ItemAbilityTrueSight: return "True Sight"
        case ItemAbilityWaterWalking: return "Water Walking"
        case ItemAbilityWraithform: return "Wraithform"

        case ItemAbilityVampiric: return "Vampiric"
        case ItemAbilityDestruction: return "Destruction"
        case ItemAbilityRegeneration: return "Regeneration"
        case ItemAbilityPathfinding: return "Pathfinding"
        case ItemAbilityStoning: return "Stoning"
        case ItemAbilityInvisibility: return "Invisibility"
        case ItemAbilityMagicImmunity: return "Magic Immunity"
        case ItemAbilityHolyAvenger: return "Holy Avenger"
        case ItemAbilityMerging: return "Merging"
    }

    return ""
}

func (item ItemAbility) LbxFile() string {
    switch item {
        case ItemAbilityHolyAvenger: return "itemisc.lbx"
        case ItemAbilityVampiric: return "special.lbx"
        case ItemAbilityDestruction: return "special2.lbx"
        case ItemAbilityStoning: return "special.lbx"
    }

    return ""
}

func (item ItemAbility) LbxIndex() int {
    switch item {
        case ItemAbilityStoning: return 27
        case ItemAbilityHolyAvenger: return 33
        case ItemAbilityDestruction: return 5
        case ItemAbilityVampiric: return 31
    }

    return -1
}
