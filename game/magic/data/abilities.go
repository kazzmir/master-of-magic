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

    // artifact abilities
    AbilityBless
    AbilityChaos
    AbilityCloakOfFear
    AbilityDeath
    AbilityDeathTouch
    AbilityDestruction
    AbilityElementalArmor
    AbilityEndurance
    AbilityFlaming
    AbilityFlight
    AbilityGiantStrength
    AbilityGuardianWind
    AbilityHaste
    AbilityHolyAvenger
    AbilityInvulnerability
    AbilityLightning
    AbilityLionHeart
    AbilityPhantasmal
    AbilityPlanarTravel
    AbilityPowerDrain
    AbilityResistElements
    AbilityResistMagic
    AbilityRighteousness
    AbilityStoning
    AbilityTrueSight
    AbilityVampiric
    AbilityWaterWalking
    AbilityWraithform
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
        case AbilityHolyAvenger: return "itemisc.lbx"
        case AbilityDoomBoltSpell: return "special.lbx"
        case AbilityDoomGaze: return "special.lbx"
        case AbilityDeathTouch: return "special2.lbx"
        case AbilityDestruction: return "special2.lbx"
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
        case AbilityLifeSteal, AbilityVampiric: return "special.lbx"
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
        case AbilityStoningTouch, AbilityStoning: return "special.lbx"
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
        case AbilityHolyAvenger: return "Holy Avenger"
        case AbilityDoomBoltSpell: return "Doom Bolt Spell"
        case AbilityDoomGaze: return "Doom Gaze"
        case AbilityDeathTouch: return "Death Touch"
        case AbilityDestruction: return "Destruction"
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
        case AbilityVampiric: return "Vampiric"
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
        case AbilityStoning: return "Stoning"
        case AbilitySummonDemons: return "Summon Demons"
        case AbilityTeleporting: return "Teleporting"
        case AbilityThrown: return "Thrown"
        case AbilityToHit: return fmt.Sprintf("+%v To Hit", ability.Value/10)
        case AbilityTransport: return "Transport"
        case AbilityWallCrusher: return "Wall Crusher"
        case AbilityWeaponImmunity: return "Weapon Immunity"
        case AbilityWebSpell: return "Web Spell"
        case AbilityWindWalking: return "Wind Walking"
        case AbilityAgility: return "Agility"
        case AbilitySuperAgility: return "Super Agility"
        case AbilityArcanePower: return "Arcane Power"
        case AbilitySuperArcanePower: return "Super Arcane Power"
        case AbilityArmsmaster: return "Armsmaster"
        case AbilitySuperArmsmaster: return "Super Armsmaster"
        case AbilityBlademaster: return "Blademaster"
        case AbilitySuperBlademaster: return "Super Blademaster"
        case AbilityCaster: return fmt.Sprintf("Caster %v mp", ability.Value)
        case AbilityCharmed: return "Charmed"
        case AbilityConstitution: return "Constitution"
        case AbilitySuperConstitution: return "Super Constitution"
        case AbilityLeadership: return "Leadership"
        case AbilitySuperLeadership: return "Super Leadership"
        case AbilityLegendary: return "Legendary"
        case AbilitySuperLegendary: return "Super Legendary"
        case AbilityLucky: return "Lucky"
        case AbilityMight: return "Might"
        case AbilitySuperMight: return "Super Might"
        case AbilityNoble: return "Noble"
        case AbilityPrayermaster: return "Prayermaster"
        case AbilitySuperPrayermaster: return "Super Prayermaster"
        case AbilitySage: return "Sage"
        case AbilitySuperSage: return "Super Sage"

        case AbilityBless: return "Bless"
        case AbilityChaos: return "Chaos"
        case AbilityCloakOfFear: return "Cloak Of Fear"
        case AbilityDeath: return "Death"
        case AbilityElementalArmor: return "Elemental Armor"
        case AbilityEndurance: return "Endurance"
        case AbilityFlaming: return "Flaming"
        case AbilityFlight: return "Flight"
        case AbilityGiantStrength: return "Giant Strength"
        case AbilityGuardianWind: return "Guardian Wind"
        case AbilityHaste: return "Haste"
        case AbilityInvulnerability: return "Invulnerability"
        case AbilityLightning: return "Lightning"
        case AbilityLionHeart: return "Lion Heart"
        case AbilityPhantasmal: return "Phantasmal"
        case AbilityPlanarTravel: return "Planar Travel"
        case AbilityPowerDrain: return "Power Drain"
        case AbilityResistElements: return "Resist Elements"
        case AbilityResistMagic: return "Resist Magic"
        case AbilityRighteousness: return "Righteousness"
        case AbilityTrueSight: return "True Sight"
        case AbilityWaterWalking: return "Water Walking"
        case AbilityWraithform: return "Wraithform"
    }

    return "?"
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
        case AbilityHolyAvenger: return 33
        case AbilityDoomBoltSpell: return 41
        case AbilityDoomGaze: return 26
        case AbilityDeathTouch: return 30
        case AbilityDestruction: return 5
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
        case AbilityLifeSteal, AbilityVampiric: return 31
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
        case AbilityStoningTouch, AbilityStoning: return 27
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

func (ability Ability) MagicType() MagicType {
    switch ability.Ability {
        case AbilityWraithform, AbilityVampiric,
             AbilityDeath, AbilityPowerDrain,
             AbilityCloakOfFear: return DeathMagic

        case AbilityGuardianWind, AbilityHaste,
             AbilityInvisibility, AbilityFlight,
             AbilityResistMagic, AbilityMagicImmunity,
             AbilityPhantasmal: return SorceryMagic

        case AbilityWaterWalking, AbilityRegeneration,
             AbilityPathfinding, AbilityMerging,
             AbilityResistElements, AbilityElementalArmor,
             AbilityGiantStrength, AbilityStoning: return NatureMagic

        case AbilityHolyAvenger, AbilityTrueSight,
             AbilityBless, AbilityRighteousness,
             AbilityInvulnerability, AbilityEndurance,
             AbilityPlanarTravel, AbilityLionHeart: return LifeMagic

        case AbilityFlaming, AbilityLightning,
             AbilityChaos, AbilityDestruction: return ChaosMagic
    }

    return MagicNone
}
