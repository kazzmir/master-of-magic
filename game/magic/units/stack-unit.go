package units

import (
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/lib/fraction"
)

type StackUnit interface {
    SetId(id uint64)
    ResetMoves()
    NaturalHeal(rate float64)
    IsFlying() bool
    IsSwimmer() bool
    GetName() string
    GetTitle() string
    GetPlane() data.Plane
    SetPlane(data.Plane)
    GetMovesLeft() fraction.Fraction
    SetMovesLeft(fraction.Fraction)
    GetRace() data.Race
    GetRealm() data.MagicType
    GetUpkeepGold() int
    GetUpkeepFood() int
    GetUpkeepMana() int
    GetEnchantments() []data.UnitEnchantment
    AddEnchantment(data.UnitEnchantment)
    HasEnchantment(data.UnitEnchantment) bool
    RemoveEnchantment(data.UnitEnchantment)
    IsUndead() bool
    GetBanner() data.BannerType
    SetWeaponBonus(data.WeaponBonus)
    GetWeaponBonus() data.WeaponBonus
    GetX() int
    GetY() int
    IsHero() bool
    Move(int, int, fraction.Fraction, NormalizeCoordinateFunc)
    GetLbxFile() string
    GetLbxIndex() int
    GetKnownSpells() []string
    HasAbility(data.AbilityType) bool
    GetAbilityValue(data.AbilityType) float32
    GetAbilities() []data.Ability
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
    CanTouchAttack(Damage) bool
    GetArtifactSlots() []artifact.ArtifactSlot
    GetArtifacts() []*artifact.Artifact
    GetSpellChargeSpells() map[spellbook.Spell]int
    GetBusy() BusyStatus
    SetBusy(BusyStatus)
}

