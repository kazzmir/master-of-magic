package main

import (
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

func getUnitCost(unit *units.Unit) uint64 {
    // Settlers always cost 10
    if unit.IsSettlers() {
        return 10
    }

    if unit.Race == data.RaceHero {
        return 500000
    }

    // Base cost: health, attack, defense, abilities
    health := unit.GetHitPoints() * unit.GetCount() / 2
    melee := unit.GetMeleeAttackPower() * unit.GetCount()
    ranged := unit.GetRangedAttackPower() * unit.RangedAttacks * unit.GetCount()
    var rangedMultiplier float32 = 1.0
    if unit.GetRangedAttackDamageType() == units.DamageRangedMagical {
        rangedMultiplier = 2
    }

    defense := unit.GetDefense() * unit.GetCount()
    resistance := unit.GetResistance() * unit.GetCount()

    // Ability modifier: +10% per ability, +20% for some strong ones
    var abilityValue float32 = 1
    for _, ability := range unit.GetAbilities() {
        switch ability.Ability {
            case data.AbilityArmorPiercing: abilityValue = 1.3
            case data.AbilityCauseFear: abilityValue = 1.8
            case data.AbilityColdImmunity: abilityValue = 1.1
            case data.AbilityDeathGaze: abilityValue = 1.5 * ability.Value
            case data.AbilityDeathImmunity: abilityValue = 2
            case data.AbilityDispelEvil: abilityValue = 1.1
            case data.AbilityDoomBoltSpell: abilityValue = 1.5
            case data.AbilityDoomGaze: abilityValue = 1.8 * ability.Value
            case data.AbilityDeathTouch: abilityValue = 2
            case data.AbilityFireballSpell: abilityValue = 2
            case data.AbilityFireBreath: abilityValue = 1.3 * ability.Value
            case data.AbilityFireImmunity: abilityValue = 1.2
            case data.AbilityFirstStrike: abilityValue = 1.5
            case data.AbilityHealingSpell: abilityValue = 1.4
            case data.AbilityHolyBonus: abilityValue = 1.3
            case data.AbilityIllusion: abilityValue = 1.6
            case data.AbilityIllusionsImmunity: abilityValue = 1.7
            case data.AbilityImmolation: abilityValue = 1.9
            case data.AbilityInvisibility: abilityValue = 2.4
            case data.AbilityLargeShield: abilityValue = 1.1
            case data.AbilityLifeSteal: abilityValue = 1.5 * -ability.Value
            case data.AbilityLightningBreath: abilityValue = 1.8 * ability.Value
            case data.AbilityLongRange: abilityValue = 1.1
            case data.AbilityMagicImmunity: abilityValue = 2.5
            case data.AbilityMerging: abilityValue = 2
            case data.AbilityMissileImmunity: abilityValue = 2
            case data.AbilityNegateFirstStrike: abilityValue = 1.3
            case data.AbilityNonCorporeal: abilityValue = 1.2
            case data.AbilityPathfinding: abilityValue = 1.1
            case data.AbilityPoisonImmunity: abilityValue = 1.3
            case data.AbilityPoisonTouch: abilityValue = 1.7 * ability.Value
            case data.AbilityRegeneration: abilityValue = 2.5
            case data.AbilityResistanceToAll: abilityValue = 1.9
            case data.AbilityStoningGaze: abilityValue = 1.6 * ability.Value
            case data.AbilityStoningImmunity: abilityValue = 1.5
            case data.AbilityStoningTouch: abilityValue = 1.6 * ability.Value
            case data.AbilitySummonDemons: abilityValue = 1.3
            case data.AbilityToHit: abilityValue = 1.3 * ability.Value / 10
            case data.AbilityTeleporting: abilityValue = 2
            case data.AbilityThrown: abilityValue = 1.4 * ability.Value
            case data.AbilityWallCrusher: abilityValue = 1.1
            case data.AbilityWeaponImmunity: abilityValue = 1.3
            case data.AbilityWebSpell: abilityValue = 1.3
        }
    }

    // Magic/fantastic units: add casting cost if present
    magicCost := 0
    if unit.CastingCost > 0 {
        magicCost = unit.CastingCost * 5
    }

    // Main cost formula
    cost := (float32(health)*3 + float32(melee)*4 + float32(ranged)*3*rangedMultiplier + float32(defense)*2 + float32(resistance) + float32(magicCost)) * abilityValue
    if cost < 10 {
        cost = 10
    }
    return uint64(cost)
}
