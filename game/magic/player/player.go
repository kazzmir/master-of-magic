package player

import (
    "slices"

    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
)

type Unit struct {
    Unit units.Unit
    Banner data.BannerType
    Plane data.Plane
    MovesLeft fraction.Fraction
    Patrol bool
    X int
    Y int
    Id uint64
}

const MovementLimit = 10

func (unit *Unit) ResetMoves(){
    unit.MovesLeft = fraction.FromInt(unit.Unit.MovementSpeed)
}

func (unit *Unit) Move(dx int, dy int, cost fraction.Fraction){
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

type UnitStack struct {
    units []*Unit
    active map[*Unit]bool

    // non-zero while animating movement on the overworld
    offsetX float64
    offsetY float64
}

func MakeUnitStack() *UnitStack {
    return MakeUnitStackFromUnits(nil)
}

func MakeUnitStackFromUnits(units []*Unit) *UnitStack {
    stack := &UnitStack{
        units: units,
        active: make(map[*Unit]bool),
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

func (stack *UnitStack) Units() []*Unit {
    return stack.units
}

func (stack *UnitStack) ActiveUnits() []*Unit {
    var out []*Unit
    for unit, active := range stack.active {
        if active {
            out = append(out, unit)
        }
    }

    return out
}

func (stack *UnitStack) InactiveUnits() []*Unit {
    var inactive []*Unit
    for unit, active := range stack.active {
        if !active {
            inactive = append(inactive, unit)
        }
    }

    return inactive
}

func (stack *UnitStack) AllFlyers() bool {
    for _, unit := range stack.ActiveUnits() {
        if !unit.Unit.Flying {
            return false
        }
    }

    return true
}

func (stack *UnitStack) ToggleActive(unit *Unit){
    value, ok := stack.active[unit]
    if ok {
        // if unit is active then set to inactive
        // if unit is inactive, then only set to active if the unit has moves left

        if value {
            stack.active[unit] = false
        } else if unit.MovesLeft.GreaterThan(fraction.Zero()) {
            stack.active[unit] = true
            unit.Patrol = false
        }
    }
}

func (stack *UnitStack) AddUnit(unit *Unit){
    stack.units = append(stack.units, unit)
    stack.active[unit] = true
}

func (stack *UnitStack) IsActive(unit *Unit) bool {
    val, ok := stack.active[unit]
    if !ok {
        return false
    }
    return val
}

func (stack *UnitStack) RemoveUnits(units []*Unit){
    for _, unit := range units {
        stack.RemoveUnit(unit)
    }
}

func (stack *UnitStack) RemoveUnit(unit *Unit){
    stack.units = slices.DeleteFunc(stack.units, func(u *Unit) bool {
        return u == unit
    })

    delete(stack.active, unit)
}

func (stack *UnitStack) ContainsUnit(unit *Unit) bool {
    return slices.Contains(stack.units, unit)
}

func (stack *UnitStack) Plane() data.Plane {
    if len(stack.units) > 0 {
        return stack.units[0].Plane
    }

    return data.PlaneArcanus
}

func (stack *UnitStack) ExhaustMoves(){
    for _, unit := range stack.units {
        unit.MovesLeft = fraction.Zero()
        stack.active[unit] = false
    }
}

func (stack *UnitStack) EnableMovers(){
    for _, unit := range stack.units {
        if unit.MovesLeft.GreaterThan(fraction.Zero()) && !unit.Patrol {
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
        if unit.MovesLeft.GreaterThan(fraction.Zero()) {
            return false
        }
    }

    return true
}

// true if any unit in the stack has moves left
func (stack *UnitStack) HasMoves() bool {
    return !stack.OutOfMoves()
}

func (stack *UnitStack) Leader() *Unit {
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
        return stack.units[0].X
    }

    return 0
}

func (stack *UnitStack) Y() int {
    if len(stack.units) > 0 {
        return stack.units[0].Y
    }

    return 0
}

type Player struct {
    // matrix the same size as the map, where true means the player can see the tile
    // and false means the tile has not yet been discovered
    ArcanusFog [][]bool
    MyrrorFog [][]bool

    TaxRate float32

    Gold int
    Food int
    Mana int

    // known spells
    Spells spellbook.Spells

    CastingSkill int

    Wizard setup.WizardCustom

    Units []*Unit
    Stacks []*UnitStack
    Cities []*citylib.City

    // counter for the next created unit owned by this player
    UnitId uint64
    SelectedStack *UnitStack
}

func (player *Player) GetFog(plane data.Plane) [][]bool {
    if plane == data.PlaneArcanus {
        return player.ArcanusFog
    } else {
        return player.MyrrorFog
    }
}

func (player *Player) SetSelectedStack(stack *UnitStack){
    player.SelectedStack = stack
}

/* make anything within the given radius viewable by the player */
func (player *Player) LiftFog(x int, y int, radius int){

    // FIXME: make this a parameter
    fog := player.ArcanusFog

    for dx := -radius; dx <= radius; dx++ {
        for dy := -radius; dy <= radius; dy++ {
            if x + dx < 0 || x + dx >= len(fog) || y + dy < 0 || y + dy >= len(fog[0]) {
                continue
            }

            if dx * dx + dy * dy <= radius * radius {
                fog[x + dx][y + dy] = true
            }
        }
    }

}

func (player *Player) FindStackByUnit(unit *Unit) *UnitStack {
    for _, stack := range player.Stacks {
        if stack.ContainsUnit(unit) {
            return stack
        }
    }

    return nil
}

func (player *Player) FindStack(x int, y int) *UnitStack {
    for _, stack := range player.Stacks {
        if stack.X() == x && stack.Y() == y {
            return stack
        }
    }

    return nil
}

func (player *Player) MergeStacks(stack1 *UnitStack, stack2 *UnitStack) *UnitStack {
    stack1.units = append(stack1.units, stack2.units...)

    for unit, active := range stack2.active {
        stack1.active[unit] = active
    }

    player.Stacks = slices.DeleteFunc(player.Stacks, func (s *UnitStack) bool {
        return s == stack2
    })

    return stack1
}

func (player *Player) RemoveUnit(unit *Unit) {
    player.Units = slices.DeleteFunc(player.Units, func (u *Unit) bool {
        return u == unit
    })

    stack := player.FindStack(unit.X, unit.Y)
    if stack != nil {
        stack.RemoveUnit(unit)

        if stack.IsEmpty() {
            player.Stacks = slices.DeleteFunc(player.Stacks, func (s *UnitStack) bool {
                return s == stack
            })
        }
    }
}

func (player *Player) AddCity(city *citylib.City) *citylib.City {
    player.Cities = append(player.Cities, city)
    return city
}

func (player *Player) AddStack(stack *UnitStack){
    player.Stacks = append(player.Stacks, stack)
}

func (player *Player) AddUnit(unit Unit) *Unit {
    unit.Id = player.UnitId
    player.UnitId += 1
    unit_ptr := &unit
    player.Units = append(player.Units, unit_ptr)

    stack := player.FindStack(unit.X, unit.Y)
    if stack == nil {
        stack = MakeUnitStack()
        player.Stacks = append(player.Stacks, stack)
    } else {
    }

    stack.AddUnit(unit_ptr)

    return unit_ptr
}
