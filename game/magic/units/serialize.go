package units

import (
    "github.com/kazzmir/master-of-magic/game/magic/data"
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
