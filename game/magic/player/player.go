package player

import (
    "slices"

    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
)

type Unit struct {
    Unit units.Unit
    Banner data.BannerType
    Plane data.Plane
    X int
    Y int
    Id uint64

    MovementAnimation int
    // the tile the unit was just on in order to animate moving around
    MoveX int
    MoveY int
}

const MovementLimit = 10

func (unit *Unit) Move(dx int, dy int){
    unit.MovementAnimation = MovementLimit

    unit.MoveX = unit.X
    unit.MoveY = unit.Y

    unit.X += dx
    unit.Y += dy

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
    active []bool
}

func (stack *UnitStack) IsEmpty() bool {
    return len(stack.units) == 0
}

func (stack *UnitStack) MakeStack(units []*Unit){
    stack.units = units
    stack.active = make([]bool, len(units))
    for i := range stack.active {
        stack.active[i] = true
    }
}

func (stack *UnitStack) Units() []*Unit {
    return stack.units
}

func (stack *UnitStack) AddUnit(unit *Unit){
    stack.units = append(stack.units, unit)
    stack.active = append(stack.active, true)
}

func (stack *UnitStack) RemoveUnit(unit *Unit){
    index := -1

    for i := 0; i < len(stack.units); i++ {
        if stack.units[i] == unit {
            index = i
        }
    }

    stack.units = slices.Delete(stack.units, index, index+1)
    stack.active = slices.Delete(stack.active, index, index+1)
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

func (stack *UnitStack) Move(dx int, dy int){
    for _, unit := range stack.units {
        unit.Move(dx, dy)
    }
}

func (stack *UnitStack) Leader() *Unit {
    if len(stack.units) > 0 {
        return stack.units[0]
    }

    return nil
}

// reduce movement of all units, return true if units are done moving
func (stack *UnitStack) UpdateMovement() bool {
    for _, unit := range stack.units {
        if unit.MovementAnimation > 0 {
            unit.MovementAnimation -= 1
        }
    }

    if len(stack.units) > 0 {
        return stack.units[0].MovementAnimation == 0
    }

    return true
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

func (player *Player) AddCity(city citylib.City) {
    player.Cities = append(player.Cities, &city)
}

func (player *Player) AddUnit(unit Unit) *Unit {
    unit.Id = player.UnitId
    player.UnitId += 1
    unit_ptr := &unit
    player.Units = append(player.Units, unit_ptr)

    stack := player.FindStack(unit.X, unit.Y)
    if stack == nil {
        stack = &UnitStack{}
        player.Stacks = append(player.Stacks, stack)
    } else {
    }

    stack.AddUnit(unit_ptr)

    return unit_ptr
}
