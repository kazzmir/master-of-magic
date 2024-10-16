package units

import (
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/lib/fraction"
)

type StackUnit interface {
    SetId(id uint64)
    ResetMoves()
    NaturalHeal()
    GetPatrol() bool
    SetPatrol(bool)
    IsFlying() bool
    GetName() string
    GetPlane() data.Plane
    GetMovesLeft() fraction.Fraction
    SetMovesLeft(fraction.Fraction)
    GetRace() data.Race
    GetUpkeepGold() int
    GetUpkeepFood() int
    GetUpkeepMana() int
    GetBanner() data.BannerType
    GetX() int
    GetY() int
    Move(int, int, fraction.Fraction)
    GetLbxFile() string
    GetLbxIndex() int
    HasAbility(Ability) bool
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
    GetRawUnit() Unit
    GetToHitMelee() int
}

