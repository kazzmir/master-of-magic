package player

import (
    "image"

    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

// stores information about every stack and city for fast lookups
// Warning: do not store this information for long periods of time as it will become out of date
type CityStackInfo struct {
    ArcanusStacks map[image.Point]*UnitStack
    MyrrorStacks map[image.Point]*UnitStack
    ArcanusCities map[image.Point]*citylib.City
    MyrrorCities map[image.Point]*citylib.City
}

func (info CityStackInfo) FindStack(x int, y int, plane data.Plane) *UnitStack {
    var use map[image.Point]*UnitStack

    switch plane {
        case data.PlaneArcanus: use = info.ArcanusStacks
        case data.PlaneMyrror: use = info.MyrrorStacks
    }

    stack, ok := use[image.Pt(x, y)]
    if ok {
        return stack
    }

    return nil
}

func (info CityStackInfo) FindCity(x int, y int, plane data.Plane) *citylib.City {
    var use map[image.Point]*citylib.City

    switch plane {
        case data.PlaneArcanus: use = info.ArcanusCities
        case data.PlaneMyrror: use = info.MyrrorCities
    }

    city, ok := use[image.Pt(x, y)]
    if ok {
        return city
    }

    return nil
}

func (info CityStackInfo) ContainsEnemy(x int, y int, plane data.Plane, player *Player) bool {
    stack := info.FindStack(x, y, plane)
    if stack != nil && stack.GetBanner() != player.GetBanner() {
        return true
    }

    city := info.FindCity(x, y, plane)
    if city != nil && city.GetBanner() != player.GetBanner() {
        return true
    }

    return false
}

func (info CityStackInfo) FindFriendlyStack(x int, y int, plane data.Plane, player *Player) *UnitStack {
    stack := info.FindStack(x, y, plane)

    if stack != nil && stack.GetBanner() == player.GetBanner() {
        return stack
    }

    return nil
}

