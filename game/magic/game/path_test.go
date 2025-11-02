package game

import (
    "testing"
    "image"

    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
)

func TestPathBasic(test *testing.T) {
    var game Game

    // just have two tiles, land and ocean
    terrainData := terrain.MakeTerrainData([]image.Image{nil, nil}, []terrain.TerrainTile{
        terrain.TerrainTile{TileIndex: 0, Tile: terrain.TileLand},
        terrain.TerrainTile{TileIndex: 1, Tile: terrain.TileOcean},
    })

    xmap := maplib.Map{
        Map: terrain.MakeMap(1, 4),
        Data: terrainData,
        Plane: data.PlaneArcanus,
    }

    // ocean, ocean, land, land
    xmap.Map.Terrain[0][0] = 1
    xmap.Map.Terrain[1][0] = 1
    xmap.Map.Terrain[2][0] = 0
    xmap.Map.Terrain[3][0] = 0

    game.ArcanusMap = &xmap

    makeFog := func(width int, height int) data.FogMap {
        fog := make(data.FogMap, width)
        for x := 0; x < width; x++ {
            fog[x] = make([]data.FogType, height)

            for y := range height {
                fog[x][y] = data.FogTypeVisible
            }
        }
        return fog
    }

    fog := makeFog(3, 1)

    checkValidPath := func (unit units.Unit, fromX, toX int) bool {
        player1 := playerlib.MakePlayer(setup.WizardCustom{}, true, 3, 1, map[herolib.HeroType]string{}, &game)
        walker := player1.AddUnit(units.MakeOverworldUnit(unit, fromX, 0, data.PlaneArcanus))

        return len(game.FindPath(walker.GetX(), walker.GetY(), toX, 0, player1, player1.FindStack(walker.GetX(), walker.GetY(), data.PlaneArcanus), fog)) > 0
    }

    // land walking unit can move from one land tile to another
    if !checkValidPath(units.HighMenSwordsmen, 2, 3) {
        test.Errorf("Expected path from land to land")
    }

    // land walking unit without swimming ability cannot move from land -> water
    if checkValidPath(units.HighMenSwordsmen, 2, 1) {
        test.Errorf("Land walker cannot move from land to water")
    }

    // land walking unit with swimming ability can move from land -> water
    if !checkValidPath(units.LizardSwordsmen, 2, 1) {
        test.Errorf("Swimmer can move from land to water")
    }

    // flying unit can walk from land -> water, and water->land
    if !checkValidPath(units.SkyDrake, 2, 1) {
        test.Errorf("Flying unit can move from land to water")
    }

    if !checkValidPath(units.SkyDrake, 1, 2) {
        test.Errorf("Flying unit can move from water to land")
    }

    // stack with flying + land unit can walk land->land but not land->water
    // stack with flying + land unit with swimming ability can walk land->water

    // land walking unit with flight can walk land->land, land->water, and water->land

    // sailing unit can walk water->water, but not water->land
    // sailing unit with flying can walk water->water and water->land

    // land walking unit can move onto sailing unit if sailing unit is in water
    // land walking unit as part of a stack with a sailing unit that has flight can move into water
}
