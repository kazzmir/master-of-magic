package player

import (
    "slices"

    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/lib/fraction"
)

type ActiveMap map[units.StackUnit]bool

type UnitStack struct {
    units []units.StackUnit
    active ActiveMap

    CurrentPath pathfinding.Path

    // non-zero while animating movement on the overworld
    offsetX float64
    offsetY float64
}

func MakeUnitStack() *UnitStack {
    return MakeUnitStackFromUnits(nil)
}

func MakeUnitStackFromUnits(units []units.StackUnit) *UnitStack {
    stack := &UnitStack{
        units: units,
        active: make(ActiveMap),
    }

    for _, unit := range units {
        stack.active[unit] = true
    }

    return stack
}

func (stack *UnitStack) ResetMoves(){
    for _, unit := range stack.units {
        unit.ResetMoves()
    }
}

func (stack *UnitStack) NaturalHeal(rate float64){
    for _, unit := range stack.units {
        unit.NaturalHeal(rate)
    }
}

func (stack *UnitStack) SetOffset(x float64, y float64) {
    stack.offsetX = x
    stack.offsetY = y
}

func (stack *UnitStack) OffsetX() float64 {
    return stack.offsetX
}

func (stack *UnitStack) OffsetY() float64 {
    return stack.offsetY
}

func (stack *UnitStack) IsEmpty() bool {
    return len(stack.units) == 0
}

func (stack *UnitStack) Units() []units.StackUnit {
    return slices.Clone(stack.units)
}

func (stack *UnitStack) ActiveUnits() []units.StackUnit {
    var out []units.StackUnit
    for unit, active := range stack.active {
        if active {
            out = append(out, unit)
        }
    }

    return out
}

func (stack *UnitStack) InactiveUnits() []units.StackUnit {
    var inactive []units.StackUnit
    for unit, active := range stack.active {
        if !active {
            inactive = append(inactive, unit)
        }
    }

    return inactive
}

func (stack *UnitStack) AllFlyers() bool {
    for _, unit := range stack.ActiveUnits() {
        if !unit.IsFlying() {
            return false
        }
    }

    return true
}

func (stack *UnitStack) ToggleActive(unit units.StackUnit){
    value, ok := stack.active[unit]
    if ok {
        // if unit is active then set to inactive
        // if unit is inactive, then only set to active if the unit has moves left

        if value {
            stack.active[unit] = false
        } else if unit.GetMovesLeft().GreaterThan(fraction.Zero()) {
            stack.active[unit] = true
            unit.SetPatrol(false)
        }
    }
}

func (stack *UnitStack) AddUnit(unit units.StackUnit){
    stack.units = append(stack.units, unit)
    stack.active[unit] = true
}

func (stack *UnitStack) IsActive(unit units.StackUnit) bool {
    val, ok := stack.active[unit]
    if !ok {
        return false
    }
    return val
}

func (stack *UnitStack) RemoveUnits(units []units.StackUnit){
    for _, unit := range units {
        stack.RemoveUnit(unit)
    }
}

func (stack *UnitStack) RemoveUnit(unit units.StackUnit){
    stack.units = slices.DeleteFunc(stack.units, func(u units.StackUnit) bool {
        return u == unit
    })

    delete(stack.active, unit)
}

func (stack *UnitStack) ContainsUnit(unit units.StackUnit) bool {
    return slices.Contains(stack.units, unit)
}

func (stack *UnitStack) SetPlane(plane data.Plane) {
    for _, unit := range stack.units {
        unit.SetPlane(plane)
    }
}

func (stack *UnitStack) Plane() data.Plane {
    if len(stack.units) > 0 {
        return stack.units[0].GetPlane()
    }

    return data.PlaneArcanus
}

func (stack *UnitStack) ExhaustMoves(){
    for _, unit := range stack.units {
        unit.SetMovesLeft(fraction.Zero())
        stack.active[unit] = false
    }
}

func (stack *UnitStack) EnableMovers(){
    for _, unit := range stack.units {
        if unit.GetMovesLeft().GreaterThan(fraction.Zero()) && !unit.GetPatrol() {
            stack.active[unit] = true
        } else {
            stack.active[unit] = false
        }
    }
}

func (stack *UnitStack) Move(dx int, dy int, cost fraction.Fraction){
    for _, unit := range stack.units {
        unit.Move(dx, dy, cost)
    }
}

// true if no unit has any moves left
func (stack *UnitStack) OutOfMoves() bool {
    for _, unit := range stack.units {
        if unit.GetMovesLeft().GreaterThan(fraction.Zero()) {
            return false
        }
    }

    return true
}

// true if any unit in the stack has moves left
func (stack *UnitStack) HasMoves() bool {
    return !stack.OutOfMoves()
}

func (stack *UnitStack) Leader() units.StackUnit {
    // return the first active unit
    for _, unit := range stack.units {
        if stack.active[unit] {
            return unit
        }
    }

    // otherwise just return any unit
    if len(stack.units) > 0 {
        return stack.units[0]
    }

    return nil
}

func (stack *UnitStack) X() int {
    if len(stack.units) > 0 {
        return stack.units[0].GetX()
    }

    return 0
}

func (stack *UnitStack) Y() int {
    if len(stack.units) > 0 {
        return stack.units[0].GetY()
    }

    return 0
}
