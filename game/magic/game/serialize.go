package game

import (
    "time"

    "github.com/kazzmir/master-of-magic/game/magic/maplib"
)

const SerializeVersion = 1

func SerializeModel(model *GameModel) map[string]any {
    return map[string]any{
        "version": SerializeVersion,
        "date": time.Now(),
        "arcanus": maplib.SerializeMap(model.ArcanusMap),
        "myrror":  maplib.SerializeMap(model.MyrrorMap),
    }
}

func DeserializeModel(data map[string]any) *GameModel {
    arcanusMap := maplib.DeserializeMap(data["arcanus"].(map[string]any))
    myrrorMap := maplib.DeserializeMap(data["myrror"].(map[string]any))

    return &GameModel{
        ArcanusMap: arcanusMap,
        MyrrorMap:  myrrorMap,
    }
}
