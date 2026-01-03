package maplib

import (
    "image"
    "fmt"

    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"

    "github.com/hajimehoshi/ebiten/v2"
)

type ExtraMapData struct {
    X int `json:"x"`
    Y int `json:"y"`
    Data map[string]any `json:"data"`
}

type SerializedMap struct {
    Width int `json:"width"`
    Height int `json:"height"`
    Map [][]int `json:"map"`
    Extra []ExtraMapData `json:"extra"`
}

func SerializeMap(useMap *Map) SerializedMap {

    extraData := []ExtraMapData{}
    for point, extras := range useMap.ExtraMap {

        if len(extras) > 0 {
            data := make(map[string]any)
            for kind, tile := range extras {
                data[kind.String()] = tile.Serialize()
            }

            extraData = append(extraData, ExtraMapData{
                X: point.X,
                Y: point.Y,
                Data: data,
            })
        }
    }

    return SerializedMap{
        Width: useMap.Width(),
        Height: useMap.Height(),
        Map: useMap.Map.Terrain,
        Extra: extraData,
    }
}

func (bonus *ExtraBonus) Serialize() map[string]any {
    return map[string]any{
        "bonus": bonus.Bonus.String(),
    }
}

func (bonus ExtraBonus) Reconstruct(raw map[string]any) *ExtraBonus {
    bonusTypeStr := raw["bonus"].(string)
    bonusType := data.GetBonusByName(bonusTypeStr)
    return &ExtraBonus{
        Bonus: bonusType,
    }
}

func (encounter *ExtraEncounter) Serialize() map[string]any {

    serializedUnits := make([]units.SerializedUnit, 0)
    for _, unit := range encounter.Units {
        serializedUnits = append(serializedUnits, units.SerializeUnit(unit))
    }

    exploredBy := []string{}

    for _, player := range encounter.ExploredBy.Values() {
        exploredBy = append(exploredBy, player.GetBanner().String())
    }

    return map[string]any{
        "type": encounter.Type.Name(),
        "budget": encounter.Budget,
        "units": serializedUnits,
        "explored_by": exploredBy,
    }
}

func (node *ExtraMagicNode) Serialize() map[string]any {
    out := map[string]any{
        "kind": node.Kind.Name(),
        "zone": node.Zone,
    }

    if node.MeldingWizard != nil {
        out["melder"] = node.MeldingWizard.GetBanner().String()
        out["guardian spirit"] = node.GuardianSpiritMeld

        if node.Warped {
            out["warped"] = node.Warped
            out["warped owner"] = node.WarpedOwner.GetBanner().String()
        }
    }

    return out
}

func (volcano *ExtraVolcano) Serialize() map[string]any {
    return map[string]any{
        "caster": volcano.CastingWizard.GetBanner().String(),
    }
}

func (road *ExtraRoad) Serialize() map[string]any {
    return map[string]any{
        "enchanted": road.Enchanted,
    }
}

func (tower *ExtraOpenTower) Serialize() map[string]any {
    return map[string]any{}
}

func (corruption *ExtraCorruption) Serialize() map[string]any {
    return map[string]any{}
}

func DeserializeMap(data map[string]any) *Map {
    /*
    width := data["width"].(int)
    height := data["height"].(int)
    */
    terrainData := data["map"].([][]int)

    return &Map{
        Map: &terrain.Map{
            Terrain: terrainData,
        },
    }
}

func ReconstructExtraTile(kind ExtraKind, data map[string]any) ExtraTile {
    switch kind {
        case ExtraKindBonus: ExtraBonus{}.Reconstruct(data)
        case ExtraKindOpenTower: return &ExtraOpenTower{}
    }

    panic(fmt.Sprintf("unsupported extra tile kind: %v", kind))
}

func ReconstructMap(mapData SerializedMap, terrainData *terrain.TerrainData, cityProvider CityProvider) *Map {
    extras := make(map[image.Point]map[ExtraKind]ExtraTile)

    for _, extra := range mapData.Extra {
        point := image.Pt(extra.X, extra.Y)

        extraData, ok := extras[point]
        if !ok {
            extraData = make(map[ExtraKind]ExtraTile)
        }

        for kind, raw := range extra.Data {
            tileData := raw.(map[string]any)

            extraKind := extraKindFromString(kind)

            extraData[extraKind] = ReconstructExtraTile(extraKind, tileData)
        }

        extras[point] = extraData
    }

    return &Map{
        Map: &terrain.Map{
            Terrain: mapData.Map,
        },
        TileCache: make(map[int]*ebiten.Image),
        Data: terrainData,
        CityProvider: cityProvider,
        ExtraMap: extras,
    }
}
