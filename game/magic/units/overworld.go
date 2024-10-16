package units

import (
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/lib/fraction"
)

type OverworldUnit struct {
    Unit Unit
    Banner data.BannerType
    Plane data.Plane
    MovesLeft fraction.Fraction
    Patrol bool
    X int
    Y int
    Id uint64
    Health int
    // to get the level, use the conversion functions in experience.go
    Experience int
    ExperienceInfo ExperienceInfo
}

func (unit *OverworldUnit) GetLbxFile() string {
    return unit.Unit.LbxFile
}

func (unit *OverworldUnit) GetLbxIndex() int {
    return unit.Unit.Index
}

func (unit *OverworldUnit) GetPatrol() bool {
    return unit.Patrol
}

func (unit *OverworldUnit) SetPatrol(patrol bool) {
    unit.Patrol = patrol
}

func (unit *OverworldUnit) GetPlane() data.Plane {
    return unit.Plane
}

func (unit *OverworldUnit) GetRace() data.Race {
    return unit.Unit.Race
}

func (unit *OverworldUnit) GetX() int {
    return unit.X
}

func (unit *OverworldUnit) GetY() int {
    return unit.Y
}

func (unit *OverworldUnit) SetId(id uint64) {
    unit.Id = id
}

func (unit *OverworldUnit) GetMovesLeft() fraction.Fraction {
    return unit.MovesLeft
}

func (unit *OverworldUnit) SetMovesLeft(moves fraction.Fraction) {
    unit.MovesLeft = moves
}

func (unit *OverworldUnit) IsFlying() bool {
    return unit.Unit.Flying
}

func (unit *OverworldUnit) AddExperience(amount int) {
    unit.Experience += amount

    // normal units max out at 120 experience
    if unit.GetRace() != data.RaceHero {
        if unit.Experience > 120 {
            unit.Experience = 120
        }
    }
}

func (unit *OverworldUnit) GetExperience() int {
    return unit.Experience
}

func (unit *OverworldUnit) GetRawUnit() Unit {
    return unit.Unit
}

func (unit *OverworldUnit) AdjustHealth(amount int) {
    unit.Health += amount
    if unit.Health < 0 {
        unit.Health = 0
    }

    if unit.Health > unit.GetMaxHealth() {
        unit.Health = unit.GetMaxHealth()
    }
}

func (unit *OverworldUnit) GetMovementSound() MovementSound {
    return unit.Unit.MovementSound
}

func (unit *OverworldUnit) GetRangeAttackSound() RangeAttackSound {
    return unit.Unit.RangeAttackSound
}

func (unit *OverworldUnit) GetAttackSound() AttackSound {
    return unit.Unit.AttackSound
}

func (unit *OverworldUnit) GetCombatRangeIndex(facing Facing) int {
    return unit.Unit.GetCombatRangeIndex(facing)
}

func (unit *OverworldUnit) GetHealth() int {
    return unit.Health
}

func (unit *OverworldUnit) GetMaxHealth() int {
    return unit.GetHitPoints() * unit.GetCount()
}

func (unit *OverworldUnit) GetToHitMelee() int {
    base := 30

    level := unit.GetExperienceLevel()
    switch level {
        case ExperienceRecruit:
        case ExperienceRegular:
        case ExperienceVeteran:
        case ExperienceElite: base += 10
        case ExperienceUltraElite: base += 20
        case ExperienceChampionNormal: base += 30
    }

    return base
}

func (unit *OverworldUnit) GetRangedAttackDamageType() Damage {
    return unit.Unit.RangedAttackDamageType
}

func (unit *OverworldUnit) GetRangedAttacks() int {
    return unit.Unit.RangedAttacks
}

func (unit *OverworldUnit) HasAbility(ability Ability) bool {
    return unit.Unit.HasAbility(ability)
}

func (unit *OverworldUnit) GetBanner() data.BannerType {
    return unit.Banner
}

func (unit *OverworldUnit) GetName() string {
    return unit.Unit.GetName()
}

func (unit *OverworldUnit) GetCombatLbxFile() string {
    return unit.Unit.GetCombatLbxFile()
}

func (unit *OverworldUnit) GetCombatIndex(facing Facing) int {
    return unit.Unit.GetCombatIndex(facing)
}

func (unit *OverworldUnit) GetCount() int {
    return unit.Unit.GetCount()
}

func (unit *OverworldUnit) GetUpkeepGold() int {
    return unit.Unit.GetUpkeepGold()
}

func (unit *OverworldUnit) GetUpkeepFood() int {
    return unit.Unit.GetUpkeepFood()
}

func (unit *OverworldUnit) GetUpkeepMana() int {
    return unit.Unit.GetUpkeepMana()
}

func (unit *OverworldUnit) GetMovementSpeed() int {
    return unit.Unit.GetMovementSpeed()
}

func (unit *OverworldUnit) GetProductionCost() int {
    return unit.Unit.GetProductionCost()
}

func (unit *OverworldUnit) GetBaseMeleeAttackPower() int {
    return unit.Unit.GetMeleeAttackPower()
}

func (unit *OverworldUnit) GetExperienceLevel() NormalExperienceLevel {
    // fantastic creatures can never gain any levels
    if unit.GetRace() == data.RaceFantastic {
        return ExperienceRecruit
    }

    if unit.ExperienceInfo != nil {
        return GetNormalExperienceLevel(unit.Experience, unit.ExperienceInfo.HasWarlord(), unit.ExperienceInfo.Crusade())
    }

    return ExperienceRecruit
}

func (unit *OverworldUnit) GetMeleeAttackPower() int {
    power := unit.GetBaseMeleeAttackPower()

    level := unit.GetExperienceLevel()
    switch level {
        case ExperienceRecruit:
        case ExperienceRegular: power += 1
        case ExperienceVeteran: power += 1
        case ExperienceElite: power += 2
        case ExperienceUltraElite: power += 2
        case ExperienceChampionNormal: power += 3
    }

    return power
}

func (unit *OverworldUnit) GetBaseRangedAttackPower() int {
    return unit.Unit.GetRangedAttackPower()
}

func (unit *OverworldUnit) GetRangedAttackPower() int {
    return unit.Unit.GetRangedAttackPower()
}

func (unit *OverworldUnit) GetBaseDefense() int {
    return unit.Unit.GetDefense()
}

func (unit *OverworldUnit) GetDefense() int {
    defense := unit.GetBaseDefense()

    level := unit.GetExperienceLevel()
    switch level {
        case ExperienceRecruit:
        case ExperienceRegular:
        case ExperienceVeteran: defense += 1
        case ExperienceElite: defense += 1
        case ExperienceUltraElite: defense += 2
        case ExperienceChampionNormal: defense += 2
    }

    return defense
}

func (unit *OverworldUnit) GetResistance() int {
    base := unit.GetBaseResistance()

    level := unit.GetExperienceLevel()
    switch level {
        case ExperienceRecruit:
        case ExperienceRegular: base += 1
        case ExperienceVeteran: base += 2
        case ExperienceElite: base += 3
        case ExperienceUltraElite: base += 4
        case ExperienceChampionNormal: base += 5
    }

    return base
}

func (unit *OverworldUnit) GetBaseResistance() int {
    return unit.Unit.GetResistance()
}

func (unit *OverworldUnit) GetHitPoints() int {
    base := unit.GetBaseHitPoints()

    level := unit.GetExperienceLevel()
    switch level {
        case ExperienceRecruit:
        case ExperienceRegular:
        case ExperienceVeteran:
        case ExperienceElite: base += 1
        case ExperienceUltraElite: base += 1
        case ExperienceChampionNormal: base += 2
    }

    return base
}

func (unit *OverworldUnit) GetBaseHitPoints() int {
    return unit.Unit.GetHitPoints()
}

func (unit *OverworldUnit) GetAbilities() []Ability {
    return unit.Unit.GetAbilities()
}

func MakeOverworldUnit(unit Unit) *OverworldUnit {
    return MakeOverworldUnitFromUnit(unit, 0, 0, data.PlaneArcanus, data.BannerBrown, nil)
}

func MakeOverworldUnitFromUnit(unit Unit, x int, y int, plane data.Plane, banner data.BannerType, experienceInfo ExperienceInfo) *OverworldUnit {
    return &OverworldUnit{
        Unit: unit,
        Banner: banner,
        Plane: plane,
        MovesLeft: fraction.FromInt(unit.MovementSpeed),
        Patrol: false,
        Health: unit.GetMaxHealth(),
        ExperienceInfo: experienceInfo,
        X: x,
        Y: y,
    }
}

/* restore health points on the overworld
 * FIXME: take bonuses into account (city garrison, healer ability, etc)
 */
func (unit *OverworldUnit) NaturalHeal() {
    maxHealth := unit.GetMaxHealth()
    amount := float64(maxHealth) * 5 / 100
    if amount < 1 {
        amount = 1
    }
    unit.Health += int(amount)
    if unit.Health >= maxHealth {
        unit.Health = maxHealth
    }
}

func (unit *OverworldUnit) ResetMoves() {
    unit.MovesLeft = fraction.FromInt(unit.Unit.MovementSpeed)
}

func (unit *OverworldUnit) HasMovesLeft() bool {
    return unit.MovesLeft.GreaterThan(fraction.Zero())
}

func (unit *OverworldUnit) Move(dx int, dy int, cost fraction.Fraction){
    unit.X += dx
    unit.Y += dy

    unit.MovesLeft = unit.MovesLeft.Subtract(cost)
    if unit.MovesLeft.LessThan(fraction.Zero()) {
        unit.MovesLeft = fraction.Zero()
    }

    // FIXME: can't move off of map

    if unit.X < 0 {
        unit.X = 0
    }

    if unit.Y < 0 {
        unit.Y = 0
    }
}
