package load


import (
    "image"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/setup"

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

    for _, tower := range saveGame.Towers {
        point := image.Pt(int(tower.X), int(tower.Y))
        if map_.ExtraMap[point] == nil {
            map_.ExtraMap[point] = make(map[maplib.ExtraKind]maplib.ExtraTile)
        }
        if tower.Owner == -1 {
            continue
        }
        map_.ExtraMap[point][maplib.ExtraKindOpenTower] = &maplib.ExtraOpenTower{}
    }

    for _, lair := range saveGame.Lairs {
        if lair.Intact == 0 {
            continue
        }

        if lair.Plane == 0 && plane != data.PlaneArcanus || lair.Plane != 0 && plane != data.PlaneMyrror {
            continue
        }

        var encounterType maplib.EncounterType
        switch lair.Kind {
            case 0: encounterType = maplib.EncounterTypePlaneTower
            case 1: encounterType = maplib.EncounterTypeChaosNode
            case 2: encounterType = maplib.EncounterTypeNatureNode
            case 3: encounterType = maplib.EncounterTypeSorceryNode
            case 4: encounterType = maplib.EncounterTypeCave
            case 5: encounterType = maplib.EncounterTypeDungeon
            case 6: encounterType = maplib.EncounterTypeAncientTemple
            case 7: encounterType = maplib.EncounterTypeAbandonedKeep
            case 8: encounterType = maplib.EncounterTypeLair
            case 9: encounterType = maplib.EncounterTypeRuins
            case 10: encounterType = maplib.EncounterTypeFallenTemple
        }

        point := image.Pt(int(lair.X), int(lair.Y))
        if map_.ExtraMap[point] == nil {
            map_.ExtraMap[point] = make(map[maplib.ExtraKind]maplib.ExtraTile)
        }
        map_.ExtraMap[point][maplib.ExtraKindEncounter] = &maplib.ExtraEncounter{
            Type: encounterType,
            // FIXME: set budget, units and explored by
            Budget: 0,
            Units: nil,
            ExploredBy: nil,
        }
    }

    return &map_
}

func (saveGame *SaveGame) ToSettings() setup.NewGameSettings {
    return setup.NewGameSettings{
        Difficulty:  data.DifficultySetting(saveGame.Difficulty),
        Opponents: int(saveGame.NumPlayers) - 1,
        LandSize: int(saveGame.LandSize),
        Magic: data.MagicSetting(saveGame.Magic),
    }
}