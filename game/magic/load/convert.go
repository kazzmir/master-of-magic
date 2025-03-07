package load


import (
    "image"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"

    "github.com/hajimehoshi/ebiten/v2"
)

func (saveGame *SaveGame) ToMap(terrainData *terrain.TerrainData, plane data.Plane, cityProvider maplib.CityProvider) *maplib.Map {

    map_ := maplib.Map{
        Data: terrainData,
        Map: terrain.MakeMap(WorldHeight, WorldWidth),
        Plane: plane,
        TileCache: make(map[int]*ebiten.Image),
        ExtraMap: make(map[image.Point]map[maplib.ExtraKind]maplib.ExtraTile),
        CityProvider: cityProvider,
    }

    terrainSource := saveGame.ArcanusMap
    terrainOffset := 0
    terrainSpecials := saveGame.ArcanusTerrainSpecials
    if plane == data.PlaneMyrror {
        terrainSource = saveGame.MyrrorMap
        terrainOffset = terrain.MyrrorStart
        terrainSpecials = saveGame.MyrrorTerrainSpecials
    }

    for y := range(WorldHeight) {
        for x := range(WorldWidth) {
            point := image.Pt(x, y)

            map_.Map.Terrain[x][y] = int(terrainSource.Data[x][y]) + terrainOffset

            var bonus data.BonusType
            switch terrainSpecials[x][y] {
                case 1: bonus = data.BonusIronOre
                case 2: bonus = data.BonusCoal
                case 3: bonus = data.BonusSilverOre
                case 4: bonus = data.BonusGoldOre
                case 5: bonus = data.BonusGem
                case 6: bonus = data.BonusMithrilOre
                case 7: bonus = data.BonusAdamantiumOre
                case 8: bonus = data.BonusQuorkCrystal
                case 9: bonus = data.BonusCrysxCrystal
                case 64: bonus = data.BonusWildGame
                case 128: bonus = data.BonusNightshade
            }
            if bonus != data.BonusNone {
                map_.ExtraMap[point] = make(map[maplib.ExtraKind]maplib.ExtraTile)
                map_.ExtraMap[point][maplib.ExtraKindBonus] = &maplib.ExtraBonus{Bonus: bonus}
            }
        }
    }

    return &map_
}