package units

import (
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
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
    IsInvisible() bool
    IsTransport() bool
    GetName() string
    GetFullName() string
    GetTitle() string
    GetPlane() data.Plane
    SetPlane(data.Plane)
    GetMovesLeft(bool) fraction.Fraction
    SetMovesLeft(bool, fraction.Fraction)
    GetRace() data.Race
    GetRealm() data.MagicType
    GetUpkeepGold() int
    GetUpkeepFood() int
    GetUpkeepMana() int
    GetEnchantments() []data.UnitEnchantment
    GetUpkeepEnchantments() []data.UnitEnchantment
    AddEnchantment(data.UnitEnchantment)
    HasEnchantment(data.UnitEnchantment) bool
    RemoveEnchantment(data.UnitEnchantment)
    SetEnchantmentProvider(EnchantmentProvider)

    MeleeEnchantmentBonus(data.UnitEnchantment) int
    DefenseEnchantmentBonus(data.UnitEnchantment) int
    RangedEnchantmentBonus(data.UnitEnchantment) int
    ResistanceEnchantmentBonus(data.UnitEnchantment) int
    MovementSpeedEnchantmentBonus(fraction.Fraction, []data.UnitEnchantment) fraction.Fraction
    HitPointsEnchantmentBonus(data.UnitEnchantment) int

    IsUndead() bool
    SetUndead()
    GetBanner() data.BannerType
    SetBanner(data.BannerType)
    SetGlobalEnchantmentProvider(GlobalEnchantmentProvider)
    SetExperienceInfo(ExperienceInfo)

    SetWeaponBonus(data.WeaponBonus)
    GetWeaponBonus() data.WeaponBonus
    GetX() int
    GetY() int
    SetX(int)
    SetY(int)
    SetBuildRoadPath(pathfinding.Path)
    GetBuildRoadPath() pathfinding.Path
    IsHero() bool
    Move(int, int, fraction.Fraction, NormalizeCoordinateFunc)
    GetLbxFile() string
    GetLbxIndex() int
    GetKnownSpells() []string
    HasAbility(data.AbilityType) bool
    HasItemAbility(data.ItemAbility) bool
    GetAbilityValue(data.AbilityType) float32
    GetAbilities() []data.Ability
    GetBaseDefense() int
    GetDefense() int
    GetFullHitPoints() int
    GetBaseHitPoints() int
    GetHitPoints() int
    GetBaseMeleeAttackPower() int
    GetMeleeAttackPower() int
    GetBaseRangedAttackPower() int
    GetBaseResistance() int
    GetCombatLbxFile() string
    GetCombatIndex(Facing) int
    GetCount() int
    GetVisibleCount() int
    VisibleFigures() int
    GetMovementSpeed(bool) fraction.Fraction
    GetProductionCost() int
    GetRangedAttackPower() int
    GetResistance() int
    AdjustHealth(amount int)
    GetAttackSound() AttackSound
    GetCombatRangeIndex(Facing) int
    GetHealth() int
    GetDamage() int
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
    GetToDefend() int
    CanTouchAttack(Damage) bool
    GetArtifactSlots() []artifact.ArtifactSlot
    GetArtifacts() []*artifact.Artifact
    GetSpellChargeSpells() map[spellbook.Spell]int
    GetBusy() BusyStatus
    SetBusy(BusyStatus)
    GetSightRange() int
}

