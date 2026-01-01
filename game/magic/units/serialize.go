package units

import (
    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
)

type SerializedUnit struct {
    LbxFile string `json:"lbx_file"`
    LbxIndex int `json:"lbx_index"`
    Race data.Race `json:"race"`
    Name string `json:"name"`
}

func SerializeUnit(unit Unit) SerializedUnit {
    return SerializedUnit{
        LbxFile: unit.LbxFile,
        LbxIndex: unit.Index,
        Race: unit.Race,
        Name: unit.Name,
    }
}

func DeserializeUnit(serialized SerializedUnit) Unit {
    for _, unit := range AllUnits {
        if unit.LbxFile == serialized.LbxFile && unit.Index == serialized.LbxIndex {
            return unit
        }
    }

    return UnitNone
}

type SerializedOverworldUnit struct {
    Unit SerializedUnit `json:"unit"`
    MovesUsed fraction.Fraction `json:"moves-used"`
    Banner data.BannerType `json:"banner"`
    Plane data.Plane `json:"plane"`
    X int `json:"x"`
    Y int `json:"y"`
    Damage int `json:"damage"`
    Experience int `json:"experience"`
    WeaponBonus data.WeaponBonus `json:"weapon-bonus"`
    Undead bool `json:"undead"`

    Busy BusyStatus `json:"busy"`

    // for engineers to follow
    BuildRoadPath pathfinding.Path `json:"build-road-path"`

    Enchantments []data.UnitEnchantment `json:"enchantments"`
}

func SerializeOverworldUnit(overworldUnit *OverworldUnit) SerializedOverworldUnit {
    return SerializedOverworldUnit{
        Unit: SerializeUnit(overworldUnit.Unit),
        MovesUsed: overworldUnit.MovesUsed,
        Banner: overworldUnit.Banner,
        Plane: overworldUnit.Plane,
        X: overworldUnit.X,
        Y: overworldUnit.Y,
        Damage: overworldUnit.Damage,
        Experience: overworldUnit.Experience,
        WeaponBonus: overworldUnit.WeaponBonus,
        Undead: overworldUnit.Undead,
        Busy: overworldUnit.Busy,
        BuildRoadPath: append(make(pathfinding.Path, 0), overworldUnit.BuildRoadPath...),
        Enchantments: append(make([]data.UnitEnchantment, 0), overworldUnit.Enchantments...),
    }
}
