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
}

func MakeUnitFromUnit(unit Unit, x int, y int, plane data.Plane, banner data.BannerType) OverworldUnit {
    return OverworldUnit{
        Unit: unit,
        Banner: banner,
        Plane: plane,
        MovesLeft: fraction.FromInt(unit.MovementSpeed),
        Patrol: false,
        X: x,
        Y: y,
    }
}

func (unit *OverworldUnit) ResetMoves(){
    unit.MovesLeft = fraction.FromInt(unit.Unit.MovementSpeed)
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
