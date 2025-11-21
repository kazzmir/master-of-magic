package game

import (
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
)

type GameModel struct {
    ArcanusMap *maplib.Map
    MyrrorMap *maplib.Map
    Players []*playerlib.Player
}

func MakeGameModel(terrainData *terrain.TerrainData, settings setup.NewGameSettings, cityProvider maplib.CityProvider) *GameModel {

    planeTowers := maplib.GeneratePlaneTowerPositions(settings.LandSize, 6)

    return &GameModel{
        ArcanusMap: maplib.MakeMap(terrainData, settings.LandSize, settings.Magic, settings.Difficulty, data.PlaneArcanus, cityProvider, planeTowers),
        MyrrorMap: maplib.MakeMap(terrainData, settings.LandSize, settings.Magic, settings.Difficulty, data.PlaneMyrror, cityProvider, planeTowers),
    }
}

func (model *GameModel) AddPlayer(newPlayer *playerlib.Player) {
    model.Players = append(model.Players, newPlayer)
}
