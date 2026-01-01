package maplib

import (
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/units"
)

type ExtraMapData struct {
    X int `json:"x"`
    Y int `json:"y"`
    Data map[string]any `json:"data"`
}

func SerializeMap(useMap *Map) map[string]any {

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

    return map[string]any{
        "width": useMap.Width(),
        "height": useMap.Height(),
        "map": useMap.Map.Terrain,
        "extra": extraData,
    }
}

func (bonus *ExtraBonus) Serialize() map[string]any {
    return map[string]any{
        "bonus": bonus.Bonus.String(),
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
