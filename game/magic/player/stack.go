package player

import (
    "slices"

    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/set"
)

type ActiveMap map[units.StackUnit]bool

type PathStack interface {
    AllFlyers() bool
    AnyLandWalkers() bool
    GetBanner() data.BannerType
    Plane() data.Plane
    HasSailingUnits(bool) bool
    CanMoveOnLand(bool) bool
    ActiveUnitsDoesntHaveAbility(data.AbilityType) bool
    ActiveUnitsHasAbility(data.AbilityType) bool
    HasPathfinding() bool
}

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

func (stack *UnitStack) SetBuildRoadPath(path pathfinding.Path) {
    for _, unit := range stack.units {
        unit.SetBuildRoadPath(path)
    }
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

func (stack *UnitStack) Size() int {
    return len(stack.units)
}

func (stack *UnitStack) Units() []units.StackUnit {
    return slices.Clone(stack.units)
}

// remove the given units from this stack and put them in a new stack, then return the new stack
func (stack *UnitStack) SplitUnits(newUnits []units.StackUnit) *UnitStack {
    if len(newUnits) == 0 {
        return stack
    }

    var out []units.StackUnit
    var oldUnits []units.StackUnit

    newUnitSet := set.NewSet(newUnits...)

    for _, unit := range stack.units {
        if newUnitSet.Contains(unit) {
            out = append(out, unit)
        } else {
            oldUnits = append(oldUnits, unit)
        }
    }

    if len(out) == len(stack.units) {
        return stack
    }

    stack.units = oldUnits

    for _, unit := range out {
        delete(stack.active, unit)
    }

    return MakeUnitStackFromUnits(out)
}

// return a new stack that only contains the active units
// this might return the same stack if all units are active
func (stack *UnitStack) SplitActiveUnits() *UnitStack {
    return stack.SplitUnits(stack.ActiveUnits())

    /*
    var out []units.StackUnit
    var oldUnits []units.StackUnit
    for unit, active := range stack.active {
        if active {
            out = append(out, unit)
        } else {
            oldUnits = append(oldUnits, unit)
        }
    }

    if len(out) == len(stack.units) {
        return stack
    }

    stack.units = oldUnits

    for _, unit := range out {
        delete(stack.active, unit)
    }

    return MakeUnitStackFromUnits(out)
    */
}

func (stack *UnitStack) ActiveUnits() []units.StackUnit {
    out := make([]units.StackUnit, 0, len(stack.active))
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

func (stack *UnitStack) CanMoveOnLand(onlyActive bool) bool {
    use := stack.units
    if onlyActive {
        use = stack.ActiveUnits()
    }

    // sailing units that can't fly cannot move on land
    for _, unit := range use {
        if unit.GetRawUnit().Sailing && !unit.IsFlying() {
            return false
        }
    }

    // otherwise every other unit can move on land

    return true
}

// pass in true to only check active units
// FIXME: this should probably be HasTransport
func (stack *UnitStack) HasSailingUnits(onlyActive bool) bool {
    use := stack.units
    if onlyActive {
        use = stack.ActiveUnits()
    }
    for _, unit := range use {
        if unit.GetRawUnit().Sailing {
            return true
        }
    }

    return false
}

// true if every active unit can only walk on land
func (stack *UnitStack) AllLandWalkers() bool {
    for _, unit := range stack.ActiveUnits() {
        if !unit.IsLandWalker() {
            return false
        }
    }

    return true
}

func (stack *UnitStack) AnyLandWalkers() bool {
    return slices.ContainsFunc(stack.ActiveUnits(), units.StackUnit.IsLandWalker)
}

func (stack *UnitStack) AllFlyers() bool {
    // wind walking gives every unit in the stack the ability to fly
    if stack.ActiveUnitsHasAbility(data.AbilityWindWalking) {
        return true
    }

    for _, unit := range stack.ActiveUnits() {
        if !unit.IsFlying() {
            return false
        }
    }

    return true
}

func (stack *UnitStack) AllSwimmers() bool {
    for _, unit := range stack.ActiveUnits() {
        if !unit.IsSwimmer() {
            return false
        }
    }

    return true
}

func (stack *UnitStack) HasHero() bool {
    return slices.ContainsFunc(stack.units, units.StackUnit.IsHero)
}

// returns true if any of the active units in the stack have the given ability
func (stack *UnitStack) ActiveUnitsHasAbility(ability data.AbilityType) bool {
    for unit, active := range stack.active {
        if active && unit.HasAbility(ability) {
            return true
        }
    }

    return false
}

func (stack *UnitStack) ActiveUnitsHasEnchantment(ability data.UnitEnchantment) bool {
    for unit, active := range stack.active {
        if active && unit.HasEnchantment(ability) {
            return true
        }
    }

    return false
}

// returns true if none of the active units in the stack have the given ability
// if a single unit has the ability then return false
func (stack *UnitStack) ActiveUnitsDoesntHaveAbility(ability data.AbilityType) bool {
    for unit, active := range stack.active {
        if active && unit.HasAbility(ability) {
            return false
        }
    }

    return true
}

func (stack *UnitStack) HasPathfinding() bool {
    return stack.ActiveUnitsHasAbility(data.AbilityPathfinding) ||
           (stack.ActiveUnitsHasAbility(data.AbilityMountaineer) && stack.ActiveUnitsHasAbility(data.AbilityForester))
}

func (stack *UnitStack) AllActive() bool {
    count := 0
    for _, active := range stack.active {
        if active {
            count += 1
        }
    }

    return count == len(stack.units)
}

func (stack *UnitStack) SetActive(unit units.StackUnit, active bool){
    _, ok := stack.active[unit]
    if ok {
        stack.active[unit] = active
    }
}

func (stack *UnitStack) ToggleActive(unit units.StackUnit){
    value, ok := stack.active[unit]
    if ok {
        // if unit is active then set to inactive
        // if unit is inactive, then only set to active if the unit has moves left

        if value {
            // if there are multiple units in the stack and they are all active, then toggling this unit
            // should activate this unit and deactivate all the others

            if len(stack.units) > 1 && stack.AllActive() {
                for _, unit := range stack.units {
                    stack.active[unit] = false
                }
                stack.active[unit] = true
            } else {
                stack.active[unit] = false
            }
        } else if unit.GetMovesLeft(true).GreaterThan(fraction.Zero()) {
            stack.active[unit] = true
            unit.SetBusy(units.BusyStatusNone)
        }
    }
}

func (stack *UnitStack) ResetActive() {
    for _, unit := range stack.units {
        if unit.GetMovesLeft(true).GreaterThan(fraction.Zero()) && unit.GetBusy() == units.BusyStatusNone {
            stack.active[unit] = true
        }

        if unit.GetMovesLeft(true).LessThanEqual(fraction.Zero()) || unit.GetBusy() == units.BusyStatusNone {
            stack.active[unit] = false
        }
    }
}

func (stack *UnitStack) AddUnit(unit units.StackUnit){
    _, existing := stack.active[unit]
    if existing {
        return
    }
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
        unit.SetMovesLeft(true, fraction.Zero())
        stack.active[unit] = false
    }
}

func (stack *UnitStack) EnableMovers(){
    for _, unit := range stack.units {
        if unit.GetMovesLeft(true).GreaterThan(fraction.Zero()) && unit.GetBusy() == units.BusyStatusNone {
            stack.active[unit] = true
        } else {
            stack.active[unit] = false
        }
    }
}

// use up movement points for all active units, but don't move anywhere
func (stack *UnitStack) UseMovement(cost fraction.Fraction){
    normalize := units.NormalizeCoordinateFunc(func (x int, y int) (int, int) {
        return x, y
    })
    for _, unit := range stack.units {
        if stack.active[unit] {
            unit.Move(0, 0, cost, normalize)
        }
    }
}

// make all units that are not transport units inactive
func (stack *UnitStack) DisableNonTransport() {
    if stack.HasSailingUnits(true) {
        for _, unit := range stack.units {
            if !unit.IsTransport() {
                stack.active[unit] = false
            }
        }
    }
}

func (stack *UnitStack) Move(dx int, dy int, cost fraction.Fraction, normalize units.NormalizeCoordinateFunc){
    transport := stack.HasSailingUnits(true)
    for _, unit := range stack.units {
        unitCost := cost
        // land walking units that are being transported do not use up any movement points
        if unit.IsLandWalker() && transport {
            unitCost = fraction.Zero()
        }

        unit.Move(dx, dy, unitCost, normalize)
    }
}

// true if no unit has any moves left
func (stack *UnitStack) OutOfMoves() bool {
    for _, unit := range stack.units {
        if unit.GetBusy() == units.BusyStatusNone && unit.GetMovesLeft(true).GreaterThan(fraction.Zero()) {
            return false
        }
    }

    return true
}

func (stack *UnitStack) AnyOutOfMoves() bool {
    for _, unit := range stack.units {
        if unit.GetBusy() == units.BusyStatusNone && unit.GetMovesLeft(true).Equals(fraction.Zero()) {
            return true
        }
    }

    return false
}

func (stack *UnitStack) GetRemainingMoves() fraction.Fraction {
    hasMoves := false
    moves := fraction.Make(10000, 1)
    transport := stack.HasSailingUnits(true)
    for _, unit := range stack.units {
        // ignore units being transported
        if transport && unit.IsLandWalker() {
            continue
        }
        if unit.GetBusy() == units.BusyStatusNone && stack.active[unit] && unit.GetMovesLeft(true).LessThan(moves) {
            moves = unit.GetMovesLeft(true)
            hasMoves = true
        }
    }

    if !hasMoves {
        return fraction.Zero()
    } else {
        return moves
    }
}

// true if any unit in the stack has moves left
func (stack *UnitStack) HasMoves() bool {
    return !stack.OutOfMoves()
}

func (stack *UnitStack) GetBanner() data.BannerType {
    if len(stack.units) > 0 {
        return stack.units[0].GetBanner()
    }

    // bogus..
    return data.BannerBrown
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

func (stack *UnitStack) SetX(x int) {
    for _, unit := range stack.units {
        unit.SetX(x)
    }
}

func (stack *UnitStack) SetY(y int) {
    for _, unit := range stack.units {
        unit.SetY(y)
    }
}

func (stack *UnitStack) GetSightRange() int {
    sightRange := 0
    for _, unit := range stack.units {
        sightRange = max(sightRange, unit.GetSightRange())
    }
    return sightRange
}
