package game

import (
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
)

func SerializeModel(model *GameModel) map[string]any {
    return map[string]any{
        "arcanus": maplib.SerializeMap(model.ArcanusMap),
        "myrror":  maplib.SerializeMap(model.MyrrorMap),
    }
}
