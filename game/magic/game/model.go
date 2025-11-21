package game

import (
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
)

type GameModel struct {
    ArcanusMap *maplib.Map
    MyrrorMap *maplib.Map
}

func MakeGameModel(terrainData *terrain.TerrainData, settings setup.NewGameSettings, cityProvider maplib.CityProvider) *GameModel {

    planeTowers := maplib.GeneratePlaneTowerPositions(settings.LandSize, 6)

    return &GameModel{
        ArcanusMap: maplib.MakeMap(terrainData, settings.LandSize, settings.Magic, settings.Difficulty, data.PlaneArcanus, cityProvider, planeTowers),
        MyrrorMap: maplib.MakeMap(terrainData, settings.LandSize, settings.Magic, settings.Difficulty, data.PlaneMyrror, cityProvider, planeTowers),
    }
}
