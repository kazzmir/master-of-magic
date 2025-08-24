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
    ranged := unit.GetRangedAttackPower() * unit.RangedAttacks * unit.GetCount() / 2
    defense := unit.GetDefense() * unit.GetCount()
    resistance := unit.GetResistance() * unit.GetCount()

    // Ability modifier: +10% per ability, +20% for some strong ones
    var abilityValue float32 = 0
    for _, ability := range unit.GetAbilities() {
        switch ability.Ability {
            case data.AbilityArmorPiercing: abilityValue = 100
            case data.AbilityCauseFear: abilityValue = 100
            case data.AbilityColdImmunity: abilityValue = 100
            case data.AbilityDeathGaze: abilityValue = 100 * ability.Value
            case data.AbilityDeathImmunity: abilityValue = 100
            case data.AbilityDispelEvil: abilityValue = 100
            case data.AbilityDoomBoltSpell: abilityValue = 150
            case data.AbilityDoomGaze: abilityValue = 100 * ability.Value
            case data.AbilityDeathTouch: abilityValue = 200
            case data.AbilityFireballSpell: abilityValue = 100
            case data.AbilityFireBreath: abilityValue = 30 * ability.Value
            case data.AbilityFireImmunity: abilityValue = 100
            case data.AbilityFirstStrike: abilityValue = 50
            case data.AbilityHealingSpell: abilityValue = 100
            case data.AbilityHolyBonus: abilityValue = 50
            case data.AbilityIllusion: abilityValue = 100
            case data.AbilityIllusionsImmunity: abilityValue = 100
            case data.AbilityImmolation: abilityValue = 150
            case data.AbilityInvisibility: abilityValue = 250
            case data.AbilityLargeShield: abilityValue = 50
            case data.AbilityLifeSteal: abilityValue = 100 * -ability.Value
            case data.AbilityLightningBreath: abilityValue = 50 * ability.Value
            case data.AbilityLongRange: abilityValue = 50
            case data.AbilityMagicImmunity: abilityValue = 300
            case data.AbilityMerging: abilityValue = 300
            case data.AbilityMissileImmunity: abilityValue = 200
            case data.AbilityNegateFirstStrike: abilityValue = 100
            case data.AbilityNonCorporeal: abilityValue = 200
            case data.AbilityPathfinding: abilityValue = 50
            case data.AbilityPoisonImmunity: abilityValue = 100
            case data.AbilityPoisonTouch: abilityValue = 50 * ability.Value
            case data.AbilityRegeneration: abilityValue = 100
            case data.AbilityResistanceToAll: abilityValue = 200
            case data.AbilityStoningGaze: abilityValue = 50 * ability.Value
            case data.AbilityStoningImmunity: abilityValue = 100
            case data.AbilityStoningTouch: abilityValue = 50 * ability.Value
            case data.AbilitySummonDemons: abilityValue = 200
            case data.AbilityToHit: abilityValue = 50 * ability.Value
            case data.AbilityTeleporting: abilityValue = 200
            case data.AbilityThrown: abilityValue = 50 * ability.Value
            case data.AbilityWallCrusher: abilityValue = 100
            case data.AbilityWeaponImmunity: abilityValue = 200
            case data.AbilityWebSpell: abilityValue = 100
        }
    }

    // Magic/fantastic units: add casting cost if present
    magicCost := 0
    if unit.CastingCost > 0 {
        magicCost = unit.CastingCost * 5
    }

    // Main cost formula
    cost := health*3 + melee*4 + ranged*3 + defense*2 + resistance + magicCost + int(abilityValue)
    if cost < 10 {
        cost = 10
    }
    return uint64(cost)
}
