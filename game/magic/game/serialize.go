package game

import (
    "time"

    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/serialize"
)

const SerializeVersion = 1

type SerializedTargetCity struct {
    X int `json:"x"`
    Y int `json:"y"`
    Plane data.Plane `json:"plane"`
    Banner data.BannerType `json:"banner"`
}

// we only really have to keep track of the non-instant events
type SerializedRandomEvent struct {
    Type RandomEventType `json:"type"`
    Year uint64 `json:"year"`

    TargetCity *SerializedTargetCity `json:"city,omitempty"`

    // events that target players are all instant, so this will almost always be nil
    TargetPlayer *data.BannerType `json:"player,omitempty"`
}

func serializeRandomEvents(events []*RandomEvent) []SerializedRandomEvent {
    out := make([]SerializedRandomEvent, 0, len(events))
    for _, event := range events {
        serialized := SerializedRandomEvent{
            Type: event.Type,
            Year: event.BirthYear,
        }

        if event.TargetCity != nil {
            serialized.TargetCity = &SerializedTargetCity{
                X: event.TargetCity.X,
                Y: event.TargetCity.Y,
                Plane: event.TargetCity.Plane,
                Banner: event.TargetCity.ReignProvider.GetBanner(),
            }
        }

        if event.TargetPlayer != nil {
            banner := event.TargetPlayer.GetBanner()
            serialized.TargetPlayer = &banner
        }

        out = append(out, serialized)
    }

    return out
}

func SerializeModel(model *GameModel, saveName string) map[string]any {
    var players []playerlib.SerializedPlayer
    for _, player := range model.Players {
        players = append(players, playerlib.SerializePlayer(player))
    }

    return map[string]any{
        "metadata": serialize.SaveMetadata{
            Version: SerializeVersion,
            Date: time.Now(),
            Name: saveName,
        },
        "arcanus": maplib.SerializeMap(model.ArcanusMap),
        "myrror":  maplib.SerializeMap(model.MyrrorMap),
        "plane":  model.Plane.String(),
        "settings": model.Settings,
        "current-player": model.CurrentPlayer,
        "turn": model.TurnNumber,
        "last-event-turn": model.LastEventTurn,
        "players": players,
        "events": serializeRandomEvents(model.RandomEvents),
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
