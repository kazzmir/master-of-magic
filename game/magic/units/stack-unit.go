package units

import (
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/lbx"
)

type StackUnit interface {
    SetId(id uint64)
    ResetMoves()
    NaturalHeal(rate float64)
    GetPatrol() bool
    SetPatrol(bool)
    IsFlying() bool
    GetName() string
    GetTitle() string
    GetPlane() data.Plane
    SetPlane(data.Plane)
    GetMovesLeft() fraction.Fraction
    SetMovesLeft(fraction.Fraction)
    GetRace() data.Race
    GetUpkeepGold() int
    GetUpkeepFood() int
    GetUpkeepMana() int
    GetEnchantments() []data.UnitEnchantment
    AddEnchantment(data.UnitEnchantment)
    RemoveEnchantment(data.UnitEnchantment)
    GetBanner() data.BannerType
    SetWeaponBonus(data.WeaponBonus)
    GetWeaponBonus() data.WeaponBonus
    GetX() int
    GetY() int
    Move(int, int, fraction.Fraction)
    GetLbxFile() string
    GetLbxIndex() int
    HasAbility(AbilityType) bool
    GetAbilities() []Ability
    GetBaseDefense() int
    GetDefense() int
    GetBaseHitPoints() int
    GetHitPoints() int
    GetBaseMeleeAttackPower() int
    GetMeleeAttackPower() int
    GetBaseRangedAttackPower() int
    GetBaseResistance() int
    GetCombatLbxFile() string
    GetCombatIndex(Facing) int
    GetCount() int
    GetMovementSpeed() int
    GetProductionCost() int
    GetRangedAttackPower() int
    GetResistance() int
    AdjustHealth(amount int)
    GetAttackSound() AttackSound
    GetCombatRangeIndex(Facing) int
    GetHealth() int
    GetMaxHealth() int
    GetMovementSound() MovementSound
    GetRangeAttackSound() RangeAttackSound
    GetRangedAttackDamageType() Damage
    GetRangedAttacks() int
    AddExperience(int)
    GetExperience() int
    GetExperienceData() ExperienceData
    GetRawUnit() Unit
    GetToHitMelee() int
    GetArtifactSlots() []artifact.ArtifactSlot
    GetArtifacts() []*artifact.Artifact
    GetSpells(*lbx.LbxCache) spellbook.Spells
}

