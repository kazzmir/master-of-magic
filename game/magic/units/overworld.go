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
    return unit.GetMeleeAttackPower()
}

func (unit *OverworldUnit) GetMeleeAttackPower() int {
    return unit.Unit.GetMeleeAttackPower()
}

func (unit *OverworldUnit) GetBaseRangedAttackPower() int {
    return unit.GetRangedAttackPower()
}

func (unit *OverworldUnit) GetRangedAttackPower() int {
    return unit.Unit.GetRangedAttackPower()
}

func (unit *OverworldUnit) GetBaseDefense() int {
    return unit.Unit.GetDefense()
}

func (unit *OverworldUnit) GetDefense() int {
    return unit.Unit.GetDefense()
}

func (unit *OverworldUnit) GetResistance() int {
    return unit.Unit.GetResistance()
}

func (unit *OverworldUnit) GetHitPoints() int {
    return unit.Unit.GetMaxHealth()
}

func (unit *OverworldUnit) GetAbilities() []Ability {
    return unit.Unit.GetAbilities()
}

func MakeOverworldUnit(unit Unit) *OverworldUnit {
    return MakeOverworldUnitFromUnit(unit, 0, 0, data.PlaneArcanus, data.BannerBrown)
}

func MakeOverworldUnitFromUnit(unit Unit, x int, y int, plane data.Plane, banner data.BannerType) *OverworldUnit {
    return &OverworldUnit{
        Unit: unit,
        Banner: banner,
        Plane: plane,
        MovesLeft: fraction.FromInt(unit.MovementSpeed),
        Patrol: false,
        Health: unit.GetMaxHealth(),
        X: x,
        Y: y,
    }
}

/* restore health points on the overworld
 * FIXME: take bonuses into account (city garrison, healer ability, etc)
 */
func (unit *OverworldUnit) NaturalHeal() {
    maxHealth := unit.Unit.GetMaxHealth()
    amount := float64(maxHealth) * 5 / 100
    if amount < 1 {
        amount = 1
    }
    unit.Health += int(amount)
    if unit.Health >= maxHealth {
        unit.Health = maxHealth
    }
}

func (unit *OverworldUnit) ResetMoves(){
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
