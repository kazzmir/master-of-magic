package game

import (
    "time"

    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
)

const SerializeVersion = 1

func SerializeModel(model *GameModel) map[string]any {
    var players []playerlib.SerializedPlayer
    for _, p := range model.Players {
        players = append(players, playerlib.SerializePlayer(p))
    }

    return map[string]any{
        "version": SerializeVersion,
        "date": time.Now(),
        "arcanus": maplib.SerializeMap(model.ArcanusMap),
        "myrror":  maplib.SerializeMap(model.MyrrorMap),
        "plane":  model.Plane.String(),
        "settings": model.Settings,
        "current-player": model.CurrentPlayer,
        "turn": model.TurnNumber,
        "last-event-turn": model.LastEventTurn,
        "players": players,
        // FIXME: handle random events
        // RandomEvents []*RandomEvent
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
