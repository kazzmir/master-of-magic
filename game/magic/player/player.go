package player

import (
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
)

type Unit struct {
    Unit units.Unit
    Banner data.BannerType
    Plane data.Plane
    X int
    Y int
    Id uint64

    Movement int
    MoveX int
    MoveY int
}

const MovementLimit = 10

func (unit *Unit) Move(dx int, dy int){
    unit.Movement = MovementLimit

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

type Player struct {
    // matrix the same size as the map, where true means the player can see the tile
    // and false means the tile has not yet been discovered
    ArcanusFog [][]bool
    MyrrorFog [][]bool

    Gold int
    Food int
    Mana int

    CastingSkill int

    Wizard setup.WizardCustom

    Units []*Unit
    Cities []*citylib.City

    // counter for the next created unit owned by this player
    UnitId uint64
    SelectedUnit *Unit
}

func (player *Player) GetFog(plane data.Plane) [][]bool {
    if plane == data.PlaneArcanus {
        return player.ArcanusFog
    } else {
        return player.MyrrorFog
    }
}

func (player *Player) SetSelectedUnit(unit *Unit){
    player.SelectedUnit = unit
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

func (player *Player) AddCity(city citylib.City) {
    player.Cities = append(player.Cities, &city)
}

func (player *Player) AddUnit(unit Unit) *Unit {
    unit.Id = player.UnitId
    player.UnitId += 1
    unit_ptr := &unit
    player.Units = append(player.Units, unit_ptr)
    return unit_ptr
}
