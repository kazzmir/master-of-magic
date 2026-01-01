package city

import (
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
)

type SerializedCity struct {
    Population int
    Farmers int
    Workers int
    Rebels int
    Name string
    Plane data.Plane
    Race data.Race
    X int
    Y int
    Outpost bool
    Buildings []buildinglib.Building
    Enchantments []Enchantment
    SoldBuilding bool
    Production float32
    ProducingBuilding buildinglib.Building
    ProducingUnit units.SerializedUnit
}

func SerializeCity(city *City) SerializedCity {
    return SerializedCity{
        Population: city.Population,
        Farmers: city.Farmers,
        Workers: city.Workers,
        Rebels: city.Rebels,
        Name: city.Name,
        Plane: city.Plane,
        Race: city.Race,
        X: city.X,
        Y: city.Y,
        Outpost: city.Outpost,
        Buildings: append(make([]buildinglib.Building, 0), city.Buildings.Values()...),
        Enchantments: append(make([]Enchantment, 0), city.Enchantments.Values()...),
        SoldBuilding: city.SoldBuilding,
        Production: city.Production,
        ProducingBuilding: city.ProducingBuilding,
        ProducingUnit: units.SerializeUnit(city.ProducingUnit),
    }
}
