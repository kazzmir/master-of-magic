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

    checkValidPathOverworldUnit := func (fromX, toX int, unit... *units.OverworldUnit) bool {
        player1 := playerlib.MakePlayer(setup.WizardCustom{}, true, 3, 1, map[herolib.HeroType]string{}, &game)
        for _, u := range unit {
            newUnit := player1.AddUnit(u)
            newUnit.SetX(fromX)
            newUnit.SetY(0)
        }

        return len(game.FindPath(fromX, 0, toX, 0, player1, player1.FindStack(fromX, 0, data.PlaneArcanus), fog)) > 0
    }

    checkValidPath := func (fromX, toX int, unit... units.Unit) bool {
        var overworldUnits []*units.OverworldUnit
        for _, u := range unit {
            overworldUnits = append(overworldUnits, units.MakeOverworldUnit(u, fromX, 0, data.PlaneArcanus))
        }

        return checkValidPathOverworldUnit(fromX, toX, overworldUnits...)
    }

    // land walking unit can move from one land tile to another
    if !checkValidPath(2, 3, units.HighMenSwordsmen) {
        test.Errorf("Expected path from land to land")
    }

    // land walking unit without swimming ability cannot move from land -> water
    if checkValidPath(2, 1, units.HighMenSwordsmen) {
        test.Errorf("Land walker cannot move from land to water")
    }

    // land walking unit with swimming ability can move from land -> water
    if !checkValidPath(2, 1, units.LizardSwordsmen) {
        test.Errorf("Swimmer can move from land to water")
    }

    // flying unit can walk from land -> water, and water->land
    if !checkValidPath(2, 1, units.SkyDrake) {
        test.Errorf("Flying unit can move from land to water")
    }

    if !checkValidPath(1, 2, units.SkyDrake) {
        test.Errorf("Flying unit can move from water to land")
    }

    // stack with flying + land unit can walk land->land but not land->water
    if checkValidPath(2, 1, units.SkyDrake, units.HighMenSwordsmen) {
        test.Errorf("Flying+land cannot move from land to water")
    }

    // stack with flying + land unit with swimming ability can walk land->water
    if !checkValidPath(2, 1, units.SkyDrake, units.LizardSwordsmen) {
        test.Errorf("Flying+land cannot move from land to water")
    }

    // land walking unit with flight enchantment can walk land->land, land->water, and water->land
    flyingHighMen := units.MakeOverworldUnit(units.HighMenSwordsmen, 0, 0, data.PlaneArcanus)
    flyingHighMen.AddEnchantment(data.UnitEnchantmentFlight)
    if !checkValidPathOverworldUnit(2, 1, flyingHighMen) {
        test.Errorf("Flight enchanted land unit should be move from land to water")
    }

    if !checkValidPathOverworldUnit(2, 3, flyingHighMen) {
        test.Errorf("Flight enchanted land unit should be move from land to land")
    }

    if !checkValidPathOverworldUnit(1, 2, flyingHighMen) {
        test.Errorf("Flight enchanted land unit should be move from water to land")
    }

    // sailing unit can walk water->water, but not water->land

    // sailing unit with flying can walk water->water and water->land

    // land walking unit can move onto sailing unit if sailing unit is in water
    // land walking unit as part of a stack with a sailing unit that has flight can move into water
}
