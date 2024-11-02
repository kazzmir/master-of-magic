package units

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
)

// the file that stores the icon for this ability
func (ability Ability) LbxFile() string {
    switch ability.Ability {
        case AbilityArmorPiercing: return ""
        case AbilityCauseFear: return ""
        case AbilityColdImmunity: return ""
        case AbilityConstruction: return ""
        case AbilityCreateOutpost: return "special.lbx"
        case AbilityCreateUndead: return ""
        case AbilityDeathGaze: return ""
        case AbilityDeathImmunity: return ""
        case AbilityDispelEvil: return ""
        case AbilityDoomBoltSpell: return ""
        case AbilityDoomGaze: return "special.lbx"
        case AbilityFireballSpell: return ""
        case AbilityFireBreath: return ""
        case AbilityFireImmunity: return ""
        case AbilityFirstStrike: return ""
        case AbilityForester: return "special.lbx"
        case AbilityHealer: return ""
        case AbilityHealingSpell: return ""
        case AbilityHolyBonus: return ""
        case AbilityIllusion: return ""
        case AbilityIllusionsImmunity: return ""
        case AbilityInvisibility: return ""
        case AbilityLargeShield: return ""
        case AbilityLifeSteal: return ""
        case AbilityLightningBreath: return ""
        case AbilityLongRange: return ""
        case AbilityMagicImmunity: return ""
        case AbilityMeld: return ""
        case AbilityMerging: return ""
        case AbilityMissileImmunity: return ""
        case AbilityMountaineer: return ""
        case AbilityNegateFirstStrike: return ""
        case AbilityNonCorporeal: return ""
        case AbilityPathfinding: return ""
        case AbilityPlaneShift: return ""
        case AbilityPoisonImmunity: return ""
        case AbilityPoisonTouch: return ""
        case AbilityPurify: return ""
        case AbilityRegeneration: return ""
        case AbilityResistanceToAll: return ""
        case AbilityScouting: return ""
        case AbilityStoningGaze: return ""
        case AbilityStoningImmunity: return ""
        case AbilityStoningTouch: return ""
        case AbilitySummonDemons: return ""
        case AbilityTeleporting: return ""
        case AbilityThrown: return ""
        case AbilityWallCrusher: return ""
        case AbilityWeaponImmunity: return ""
        case AbilityWebSpell: return ""
        case AbilityWindWalking: return ""
        case AbilityAgility, AbilitySuperAgility: return "special2.lbx"
        case AbilityArcanePower, AbilitySuperArcanePower: return "special.lbx"
        case AbilityArmsmaster, AbilitySuperArmsmaster: return "special.lbx"
        case AbilityBlademaster: return "special.lbx"
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
        case AbilityArmorPiercing: return ""
        case AbilityCauseFear: return ""
        case AbilityColdImmunity: return ""
        case AbilityConstruction: return ""
        case AbilityCreateOutpost: return "Create Outpost"
        case AbilityCreateUndead: return ""
        case AbilityDeathGaze: return ""
        case AbilityDeathImmunity: return ""
        case AbilityDispelEvil: return ""
        case AbilityDoomBoltSpell: return "Doom Bolt Spell"
        case AbilityDoomGaze: return "Doom Gaze"
        case AbilityFireballSpell: return ""
        case AbilityFireBreath: return ""
        case AbilityFireImmunity: return ""
        case AbilityFirstStrike: return ""
        case AbilityForester: return "Forester"
        case AbilityHealer: return ""
        case AbilityHealingSpell: return ""
        case AbilityHolyBonus: return ""
        case AbilityIllusion: return ""
        case AbilityIllusionsImmunity: return ""
        case AbilityInvisibility: return ""
        case AbilityLargeShield: return ""
        case AbilityLifeSteal: return ""
        case AbilityLightningBreath: return ""
        case AbilityLongRange: return ""
        case AbilityMagicImmunity: return ""
        case AbilityMeld: return "Meld"
        case AbilityMerging: return "Merging"
        case AbilityMissileImmunity: return "Missile Immunity"
        case AbilityMountaineer: return "Mountaineer"
        case AbilityNegateFirstStrike: return ""
        case AbilityNonCorporeal: return ""
        case AbilityPathfinding: return ""
        case AbilityPlaneShift: return "Plane Shift"
        case AbilityPoisonImmunity: return "Poison Immunity"
        case AbilityPoisonTouch: return "Poison Touch"
        case AbilityPurify: return "Purify"
        case AbilityRegeneration: return "Regeneration"
        case AbilityResistanceToAll: return ""
        case AbilityScouting: return ""
        case AbilityStoningGaze: return ""
        case AbilityStoningImmunity: return ""
        case AbilityStoningTouch: return ""
        case AbilitySummonDemons: return ""
        case AbilityTeleporting: return ""
        case AbilityThrown: return ""
        case AbilityWallCrusher: return ""
        case AbilityWeaponImmunity: return ""
        case AbilityWebSpell: return ""
        case AbilityWindWalking: return "Wind Walking"
        case AbilityAgility: return "Agility"
        case AbilitySuperAgility: return "Super Agility"
        case AbilityArcanePower: return "Arcane Power"
        case AbilitySuperArcanePower: return "Super Arcane Power"
        case AbilityArmsmaster: return "Armsmaster"
        case AbilitySuperArmsmaster: return "Super Armsmaster"
        case AbilityBlademaster: return "Blademaster"
        case AbilitySuperBlademaster: return "Super Blademaster"
        case AbilityCaster: return "Caster"
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
    }

    return "?"
}

// the index in the lbx file for this icon
func (ability Ability) LbxIndex() int {
    switch ability.Ability {
        case AbilityArmorPiercing: return -1
        case AbilityCauseFear: return -1
        case AbilityColdImmunity: return -1
        case AbilityConstruction: return -1
        case AbilityCreateOutpost: return 17
        case AbilityCreateUndead: return -1
        case AbilityDeathGaze: return -1
        case AbilityDeathImmunity: return -1
        case AbilityDispelEvil: return -1
        case AbilityDoomBoltSpell: return -1
        case AbilityDoomGaze: return 26
        case AbilityFireballSpell: return -1
        case AbilityFireBreath: return -1
        case AbilityFireImmunity: return -1
        case AbilityFirstStrike: return -1
        case AbilityForester: return 1
        case AbilityHealer: return -1
        case AbilityHealingSpell: return -1
        case AbilityHolyBonus: return -1
        case AbilityIllusion: return -1
        case AbilityIllusionsImmunity: return -1
        case AbilityInvisibility: return -1
        case AbilityLargeShield: return -1
        case AbilityLifeSteal: return -1
        case AbilityLightningBreath: return -1
        case AbilityLongRange: return -1
        case AbilityMagicImmunity: return -1
        case AbilityMeld: return -1
        case AbilityMerging: return -1
        case AbilityMissileImmunity: return -1
        case AbilityMountaineer: return -1
        case AbilityNegateFirstStrike: return -1
        case AbilityNonCorporeal: return -1
        case AbilityPathfinding: return -1
        case AbilityPlaneShift: return -1
        case AbilityPoisonImmunity: return -1
        case AbilityPoisonTouch: return -1
        case AbilityPurify: return -1
        case AbilityRegeneration: return -1
        case AbilityResistanceToAll: return -1
        case AbilityScouting: return -1
        case AbilityStoningGaze: return -1
        case AbilityStoningImmunity: return -1
        case AbilityStoningTouch: return -1
        case AbilitySummonDemons: return -1
        case AbilityTeleporting: return -1
        case AbilityThrown: return -1
        case AbilityWallCrusher: return -1
        case AbilityWeaponImmunity: return -1
        case AbilityWebSpell: return -1
        case AbilityWindWalking: return -1
        case AbilityAgility, AbilitySuperAgility: return 32
        case AbilityArcanePower, AbilitySuperArcanePower: return 54
        case AbilityArmsmaster, AbilitySuperArmsmaster: return 46
        case AbilityBlademaster: return 47
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
