package pathfinding

import (
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

type PathStack interface {
    AllFlyers() bool
    AnyLandWalkers() bool
    GetBanner() data.BannerType
    Plane() data.Plane
    HasSailingUnits(bool) bool
    ActiveUnitsDoesntHaveAbility(data.AbilityType) bool
    ActiveUnitsHasAbility(data.AbilityType) bool
    HasPathfinding() bool
}

