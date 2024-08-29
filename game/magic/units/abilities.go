package units

type Ability int

const (
    AbilityNone Ability = iota
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
    AbilityInvisibility
    AbilityLargeShield
    AbilityLifeSteal
    AbilityLightningBreath
    AbilityLongRange
    AbilityMagicImmunity
    AbilityMeld
    AbilityMerging
    AbilityMissleImmunity
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
    AbilityTeleporting
    AbilityThrown
    AbilityWallCrusher
    AbilityWeaponImmunity
    AbilityWebSpell
    AbilityWindWalking

    // hero abilities
    AbilityAgility
    AbilityArcanePower
    AbilityArmsmaster
    AbilityBlademaster
    AbilityCaster
    AbilityCharmed
    AbilityConstitution
    AbilityLeadership
    AbilityLegendary
    AbilityLucky
    AbilityMight
    AbilityNoble
    AbilityPrayermaster
    AbilitySage
)

func (ability Ability) LbxFile() string {
    switch ability {
        case AbilityArmorPiercing: return ""
        case AbilityCauseFear: return ""
        case AbilityColdImmunity: return ""
        case AbilityConstruction: return ""
        case AbilityCreateOutpost: return ""
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
        case AbilityMissleImmunity: return ""
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
        case AbilityAgility: return ""
        case AbilityArcanePower: return ""
        case AbilityArmsmaster: return ""
        case AbilityBlademaster: return ""
        case AbilityCaster: return ""
        case AbilityCharmed: return ""
        case AbilityConstitution: return ""
        case AbilityLeadership: return ""
        case AbilityLegendary: return ""
        case AbilityLucky: return ""
        case AbilityMight: return ""
        case AbilityNoble: return ""
        case AbilityPrayermaster: return ""
        case AbilitySage: return ""
    }

    return ""

}

func (ability Ability) Name() string {
    switch ability {
        case AbilityArmorPiercing: return ""
        case AbilityCauseFear: return ""
        case AbilityColdImmunity: return ""
        case AbilityConstruction: return ""
        case AbilityCreateOutpost: return ""
        case AbilityCreateUndead: return ""
        case AbilityDeathGaze: return ""
        case AbilityDeathImmunity: return ""
        case AbilityDispelEvil: return ""
        case AbilityDoomBoltSpell: return ""
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
        case AbilityMeld: return ""
        case AbilityMerging: return ""
        case AbilityMissleImmunity: return ""
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
        case AbilityAgility: return ""
        case AbilityArcanePower: return ""
        case AbilityArmsmaster: return ""
        case AbilityBlademaster: return ""
        case AbilityCaster: return ""
        case AbilityCharmed: return ""
        case AbilityConstitution: return ""
        case AbilityLeadership: return ""
        case AbilityLegendary: return ""
        case AbilityLucky: return ""
        case AbilityMight: return ""
        case AbilityNoble: return ""
        case AbilityPrayermaster: return ""
        case AbilitySage: return ""
    }

    return ""

}

func (ability Ability) LbxIndex() int {
    switch ability {
        case AbilityArmorPiercing: return -1
        case AbilityCauseFear: return -1
        case AbilityColdImmunity: return -1
        case AbilityConstruction: return -1
        case AbilityCreateOutpost: return -1
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
        case AbilityMissleImmunity: return -1
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
        case AbilityAgility: return -1
        case AbilityArcanePower: return -1
        case AbilityArmsmaster: return -1
        case AbilityBlademaster: return -1
        case AbilityCaster: return -1
        case AbilityCharmed: return -1
        case AbilityConstitution: return -1
        case AbilityLeadership: return -1
        case AbilityLegendary: return -1
        case AbilityLucky: return -1
        case AbilityMight: return -1
        case AbilityNoble: return -1
        case AbilityPrayermaster: return -1
        case AbilitySage: return -1
    }

    return -1
}
