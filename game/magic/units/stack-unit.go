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
    IsSailing() bool
    IsLandWalker() bool
    GetName() string
    GetFullName() string
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

    MeleeEnchantmentBonus(data.UnitEnchantment) int
    DefenseEnchantmentBonus(data.UnitEnchantment) int
    RangedEnchantmentBonus(data.UnitEnchantment) int
    ResistanceEnchantmentBonus(data.UnitEnchantment) int
    MovementSpeedEnchantmentBonus(int, []data.UnitEnchantment) int

    IsUndead() bool
    GetBanner() data.BannerType
    SetBanner(data.BannerType)
    SetWeaponBonus(data.WeaponBonus)
    GetWeaponBonus() data.WeaponBonus
    GetX() int
    GetY() int
    SetX(int)
    SetY(int)
    IsHero() bool
    Move(int, int, fraction.Fraction, NormalizeCoordinateFunc)
    GetLbxFile() string
    GetLbxIndex() int
    GetKnownSpells() []string
    HasAbility(data.AbilityType) bool
    HasItemAbility(data.ItemAbility) bool
    GetAbilityValue(data.AbilityType) float32
    GetAbilities() []data.Ability
    GetFullDefense() int
    GetBaseDefense() int
    GetDefense() int
    GetFullHitPoints() int
    GetBaseHitPoints() int
    GetHitPoints() int
    GetFullMeleeAttackPower() int
    GetBaseMeleeAttackPower() int
    GetMeleeAttackPower() int
    GetBaseRangedAttackPower() int
    GetBaseResistance() int
    GetCombatLbxFile() string
    GetCombatIndex(Facing) int
    GetCount() int
    GetMovementSpeed() int
    GetProductionCost() int
    GetFullRangedAttackPower() int
    GetRangedAttackPower() int
    GetFullResistance() int
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
    GetExperienceLevel() NormalExperienceLevel
    GetHeroExperienceLevel() HeroExperienceLevel
    GetRawUnit() Unit
    GetToHitMelee() int
    CanTouchAttack(Damage) bool
    GetArtifactSlots() []artifact.ArtifactSlot
    GetArtifacts() []*artifact.Artifact
    GetSpellChargeSpells() map[spellbook.Spell]int
    GetBusy() BusyStatus
    SetBusy(BusyStatus)
    GetSightRange() int
}

