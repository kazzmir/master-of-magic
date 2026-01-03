package maplib

import (
    "image"
    "fmt"
    "encoding/json"

    "github.com/kazzmir/master-of-magic/lib/set"
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

func (encounter ExtraEncounter) Reconstruct(raw map[string]any, wizards []Wizard) *ExtraEncounter {

    exploredBy := set.MakeSet[Wizard]()

    for _, values := range raw["explored_by"].([]any) {
        bannerRaw := values.(string)
        for _, wizard := range wizards {
            if wizard.GetBanner().String() == bannerRaw {
                exploredBy.Insert(wizard)
                break
            }
        }
    }

    var encounterUnits []units.Unit
    for _, rawUnit := range raw["units"].([]any) {

        // have to marshal and unmarshal again to convert from map[string]any to SerializedUnit
        rawData, err := json.Marshal(rawUnit)
        if err != nil {
            continue
        }
        var unitData units.SerializedUnit

        err = json.Unmarshal(rawData, &unitData)
        if err != nil {
            continue
        }

        encounterUnits = append(encounterUnits, units.DeserializeUnit(unitData))
    }

    return &ExtraEncounter{
        Type: encounterByName(raw["type"].(string)),
        Budget: int(raw["budget"].(float64)),
        ExploredBy: exploredBy,
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

func (node ExtraMagicNode) Reconstruct(raw map[string]any, wizards []Wizard) *ExtraMagicNode {

    var zone []image.Point
    for _, val := range raw["zone"].([]any) {
        pointData, err := json.Marshal(val)
        if err != nil {
            continue
        }
        var point image.Point
        err = json.Unmarshal(pointData, &point)
        if err != nil {
            continue
        }
        zone = append(zone, point)
    }

    warped, ok := raw["warped"].(bool)
    if !ok {
        warped = false
    }

    guardianSpiritMeld, ok := raw["guardian spirit"].(bool)
    if !ok {
        guardianSpiritMeld = false
    }

    var meldingWizard Wizard

    melderRaw, ok := raw["melder"].(string)
    if ok {
        for _, wizard := range wizards {
            if wizard.GetBanner().String() == melderRaw {
                meldingWizard = wizard
                break
            }
        }
    }

    warpedRaw, ok := raw["warped owner"].(string)
    var warpedOwner Wizard
    if ok {
        for _, wizard := range wizards {
            if wizard.GetBanner().String() == warpedRaw {
                warpedOwner = wizard
                break
            }
        }
    }

    return &ExtraMagicNode{
        Kind: magicNodeFromString(raw["kind"].(string)),
        Zone: zone,
        Warped: warped,
        GuardianSpiritMeld: guardianSpiritMeld,
        MeldingWizard: meldingWizard,
        WarpedOwner: warpedOwner,
    }
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

func ReconstructExtraTile(kind ExtraKind, data map[string]any, wizards []Wizard) ExtraTile {
    switch kind {
        case ExtraKindBonus: return ExtraBonus{}.Reconstruct(data)
        case ExtraKindMagicNode: return ExtraMagicNode{}.Reconstruct(data, wizards)
        case ExtraKindEncounter: return ExtraEncounter{}.Reconstruct(data, wizards)
        case ExtraKindOpenTower: return &ExtraOpenTower{}
    }

    panic(fmt.Sprintf("unsupported extra tile kind: %v", kind))
}

func ReconstructMap(mapData SerializedMap, terrainData *terrain.TerrainData, cityProvider CityProvider, wizards []Wizard) *Map {
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

            extraData[extraKind] = ReconstructExtraTile(extraKind, tileData, wizards)
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
