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
    GetPlane() data.Plane
    GetMovesLeft() fraction.Fraction
    SetMovesLeft(fraction.Fraction)
    GetRace() data.Race
    GetUpkeepGold() int
    GetUpkeepFood() int
    GetUpkeepMana() int
    GetX() int
    GetY() int
    Move(int, int, fraction.Fraction)
}

