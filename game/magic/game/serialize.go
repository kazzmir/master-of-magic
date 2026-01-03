package game

import (
    "time"
    "maps"
    "slices"

    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/serialize"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
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

type SerializedGame struct {
    Metadata serialize.SaveMetadata `json:"metadata"`
    Arcanus maplib.SerializedMap `json:"arcanus"`
    Myrror maplib.SerializedMap `json:"myrror"`
    Plane data.Plane `json:"plane"`
    ArtifactPool []string `json:"artifact-pool"`
    Settings setup.NewGameSettings `json:"settings"`
    CurrentPlayer int `json:"current-player"`
    Turn uint64 `json:"turn"`
    LastEventTurn uint64 `json:"last-event-turn"`
    Players []playerlib.SerializedPlayer `json:"players"`
    Events []SerializedRandomEvent `json:"events"`
}

func SerializeModel(model *GameModel, saveName string) SerializedGame {
    var players []playerlib.SerializedPlayer
    for _, player := range model.Players {
        players = append(players, playerlib.SerializePlayer(player))
    }

    return SerializedGame{
        Metadata: serialize.SaveMetadata{
            Version: SerializeVersion,
            Date: time.Now(),
            Name: saveName,
        },
        Arcanus: maplib.SerializeMap(model.ArcanusMap),
        Myrror: maplib.SerializeMap(model.MyrrorMap),
        ArtifactPool: slices.Collect(maps.Keys(model.ArtifactPool)),
        Plane:  model.Plane,
        Settings: model.Settings,
        CurrentPlayer: model.CurrentPlayer,
        Turn: model.TurnNumber,
        LastEventTurn: model.LastEventTurn,
        Players: players,
        Events: serializeRandomEvents(model.RandomEvents),
    }
}

/*
func MakeModelFromSerialize(decoder json.Decoder) *GameModel {
    arcanusMap := maplib.DeserializeMap(data["arcanus"].(map[string]any))
    myrrorMap := maplib.DeserializeMap(data["myrror"].(map[string]any))

    return &GameModel{
        ArcanusMap: arcanusMap,
        MyrrorMap:  myrrorMap,
    }
}
*/

func reconstructArtifactPool(items []string, artifactPool map[string]*artifact.Artifact) map[string]*artifact.Artifact {
    out := make(map[string]*artifact.Artifact)

    for _, itemName := range items {
        if item, ok := artifactPool[itemName]; ok {
            out[itemName] = item
        }
    }

    return out
}

func MakeModelFromSerialized(
    serializedGame *SerializedGame, events chan GameEvent,
    heroNames map[int]map[herolib.HeroType]string,
    allSpells spellbook.Spells,
    artifactPool map[string]*artifact.Artifact,
    buildingInfo buildinglib.BuildingInfos,
    terrainData *terrain.TerrainData,
) *GameModel {
    model := &GameModel{
        Plane: serializedGame.Plane,
        ArtifactPool: reconstructArtifactPool(serializedGame.ArtifactPool, artifactPool),
        Settings: serializedGame.Settings,
        heroNames: heroNames,
        allSpells: allSpells,
        CurrentPlayer: serializedGame.CurrentPlayer,
        TurnNumber: serializedGame.Turn,
        Events: events,

        BuildingInfo: buildingInfo,
        LastEventTurn: serializedGame.LastEventTurn,

        /*
        Players []*playerlib.Player
        RandomEvents []*RandomEvent
        */
    }

    model.ArcanusMap = maplib.ReconstructMap(serializedGame.Arcanus, terrainData, model)
    model.MyrrorMap = maplib.ReconstructMap(serializedGame.Myrror, terrainData, model)

    return model
}
