package maplib

import (
    "log"
    "math"
    "math/rand/v2"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    cameralib "github.com/kazzmir/master-of-magic/game/magic/camera"

    "github.com/hajimehoshi/ebiten/v2"
)

type MiniMapCity interface {
    GetX() int
    GetY() int
    GetBanner() data.BannerType
}

type CityProvider interface {
    ContainsCity(x int, y int, plane data.Plane) bool
}

type Melder interface {
    GetBanner() data.BannerType
}

type MagicNode int
const (
    MagicNodeNature MagicNode = iota
    MagicNodeSorcery
    MagicNodeChaos
)

type ExtraTile interface {
    DrawLayer1(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int)
    DrawLayer2(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int)
}

type ExtraKind int
const (
    ExtraKindRoad ExtraKind = iota
    ExtraKindBonus
    ExtraKindMagicNode
    ExtraKindEncounter
)

var ExtraDrawOrder = []ExtraKind{
    ExtraKindRoad,
    ExtraKindBonus,
    ExtraKindMagicNode,
    ExtraKindEncounter,
}

type Direction int
const (
    DirectionNorth Direction = iota
    DirectionNorthEast
    DirectionEast
    DirectionSouthEast
    DirectionSouth
    DirectionSouthWest
    DirectionWest
    DirectionNorthWest
)

type ExtraRoad struct {
    Map *Map
    // either cast Enchant Road, or build the road in Myrror
    Enchanted bool
    X int
    Y int
}

func (road *ExtraRoad) DrawLayer1(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int){
    neighbors := road.Map.GetRoadNeighbors(road.X, road.Y)

    connected := false

    baseIndex := 45
    if road.Enchanted {
        baseIndex = 54
    }

    directionToIndex := map[Direction]int{
        DirectionNorth: 1,
        DirectionNorthEast: 2,
        DirectionEast: 3,
        DirectionSouthEast: 4,
        DirectionSouth: 5,
        DirectionSouthWest: 6,
        DirectionWest: 7,
        DirectionNorthWest: 8,
    }

    for direction, has := range neighbors {
        if has {
            index := directionToIndex[direction]
            pics, err := imageCache.GetImages("mapback.lbx", baseIndex + index)
            if err == nil {
                pic := pics[counter % uint64(len(pics))]
                screen.DrawImage(pic, options)
            }

            connected = true
        }
    }

    if !connected {
        pics, err := imageCache.GetImages("mapback.lbx", baseIndex)
        if err == nil {
            pic := pics[counter % uint64(len(pics))]
            screen.DrawImage(pic, options)
        }
    }
}

func (road *ExtraRoad) DrawLayer2(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int){
}

// wild game, gold ore, mithril, etc
type ExtraBonus struct {
    Bonus data.BonusType
}

func (bonus *ExtraBonus) DrawLayer1(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int){
    index := -1

    switch bonus.Bonus {
        case data.BonusWildGame: index = 92
        case data.BonusNightshade: index = 91
        case data.BonusSilverOre: index = 80
        case data.BonusGoldOre: index = 81
        case data.BonusIronOre: index = 78
        case data.BonusCoal: index = 79
        case data.BonusMithrilOre: index = 83
        case data.BonusAdamantiumOre: index = 84
        case data.BonusGem: index = 82
        case data.BonusQuorkCrystal: index = 85
        case data.BonusCrysxCrystal: index = 86
    }

    if index == -1 {
        return
    }

    pic, err := imageCache.GetImage("mapback.lbx", index, 0)
    if err == nil {
        screen.DrawImage(pic, options)
    }
}

func (bonus *ExtraBonus) DrawLayer2(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int){
    // nothing
}

type EncounterType int

const (
    // lair, plane tower, ancient temple, fallen temple, ruins, abandoned keep, dungeon
    EncounterTypeLair EncounterType = iota
    EncounterTypeCave
    EncounterTypePlaneTower
    EncounterTypePlaneTowerOpen
    EncounterTypeAncientTemple
    EncounterTypeFallenTemple
    EncounterTypeRuins
    EncounterTypeAbandonedKeep
    EncounterTypeDungeon
)

func randomEncounterType() EncounterType {
    all := []EncounterType{
        EncounterTypeLair,
        EncounterTypeCave,
        EncounterTypePlaneTower,
        EncounterTypeAncientTemple,
        EncounterTypeFallenTemple,
        EncounterTypeRuins,
        EncounterTypeAbandonedKeep,
        EncounterTypeDungeon,
    }

    return all[rand.N(len(all))]
}

// lair, plane tower, etc
type ExtraEncounter struct {
    Type EncounterType
    Units []units.Unit
    Empty bool
}

// choices is a map from a name to the chance of choosing that name, where all the int values should add up to 100
// an individual int is a percentage chance to choose the given key
// for example, choices might be the map {"a": 30, "b": 30, "c": 40}
// which means that a and b should both have a 30% chance of being picked, and c has a 40% chance of being picked
func chooseValue[T comparable](choices map[T]int) T {
    total := 0
    for _, value := range choices {
        total += value
    }

    pick := rand.N(total)
    for key, value := range choices {
        if pick < value {
            return key
        }

        pick -= value
    }

    var out T
    return out
}

func makeEncounter(encounterType EncounterType, difficulty data.DifficultySetting, weakStrength bool, plane data.Plane) *ExtraEncounter {
    var guardians []units.Unit
    var secondary []units.Unit

    budget := 0
    if weakStrength {
        if plane == data.PlaneArcanus {
            budget = (rand.N(20) + 1) * 30
        } else {
            budget = (rand.N(30) + 1) * 30
        }
    } else {
        if plane == data.PlaneArcanus {
            budget = (rand.N(80) + 1) * 50 + 250
        } else {
            budget = (rand.N(90) + 1) * 50 + 250
        }
    }

    bonus := float64(0)
    switch difficulty {
        case data.DifficultyIntro: bonus = -0.75
        case data.DifficultyEasy: bonus = -0.5
        case data.DifficultyAverage: bonus = -0.25
        case data.DifficultyHard: bonus = 0
        case data.DifficultyExtreme: bonus = 0.25
        case data.DifficultyImpossible: bonus = 0.50
    }

    budget = int(float64(budget) * (1.0 + bonus))

    chooseRealm := func() string {
        switch encounterType {
            case EncounterTypeLair: return chooseValue(map[string]int{"chaos": 40, "death": 40, "nature": 20})
            case EncounterTypeCave: return chooseValue(map[string]int{"chaos": 40, "death": 40, "nature": 20})
            case EncounterTypePlaneTower: return chooseValue(map[string]int{"chaos": 10, "death": 20, "nature": 10, "life": 10, "sorcery": 10})
            case EncounterTypeAncientTemple: return chooseValue(map[string]int{"death": 75, "life": 25})
            case EncounterTypeFallenTemple: return chooseValue(map[string]int{"death": 75, "life": 25})
            case EncounterTypeRuins: return chooseValue(map[string]int{"death": 75, "life": 25})
            case EncounterTypeAbandonedKeep: return chooseValue(map[string]int{"chaos": 40, "death": 40, "nature": 20})
            case EncounterTypeDungeon: return chooseValue(map[string]int{"chaos": 40, "death": 40, "nature": 20})
        }

        return ""
    }

    switch chooseRealm() {
        case "chaos":
            guardians, secondary = computeChaosNodeEnemies(budget)
        case "death":
            guardians, secondary = computeDeathNodeEnemies(budget)
        case "nature":
            guardians, secondary = computeNatureNodeEnemies(budget)
        case "life":
            guardians, secondary = computeLifeNodeEnemies(budget)
        case "sorcery":
            guardians, secondary = computeSorceryNodeEnemies(budget)
    }

    return &ExtraEncounter{
        Type: encounterType,
        Units: append(guardians, secondary...),
    }
}

func (extra *ExtraEncounter) DrawLayer1(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int){
    index := -1

    switch extra.Type {
        case EncounterTypeLair, EncounterTypeCave: index = 71
        case EncounterTypePlaneTower: index = 69
        case EncounterTypePlaneTowerOpen: index = 70
        case EncounterTypeAncientTemple: index = 72
        case EncounterTypeFallenTemple: index = 75
        case EncounterTypeRuins: index = 74
        case EncounterTypeAbandonedKeep: index = 73
        case EncounterTypeDungeon: index = 74
    }

    if index == -1 {
        return
    }

    pic, err := imageCache.GetImage("mapback.lbx", index, 0)
    if err == nil {
        screen.DrawImage(pic, options)
    }
}

func (extra *ExtraEncounter) DrawLayer2(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int){
    // nuthin'
}

type ExtraMagicNode struct {
    Kind MagicNode
    Empty bool
    Guardians []units.Unit
    Secondary []units.Unit
    // list of points that are affected by the node
    Zone []image.Point

    // if this node is melded, then this player receives the power
    MeldingWizard Melder
    // true if melded by a guardian spirit, otherwise false if melded by a magic spirit
    GuardianSpiritMeld bool

    // also contains treasure
}

func (node *ExtraMagicNode) DrawLayer1(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int){
}

func (node *ExtraMagicNode) DrawLayer2(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int){
    // if the node is melded then show the zone of influence with the sparkly images

    if node.Empty && node.MeldingWizard != nil {
        index := 63
        switch node.MeldingWizard.GetBanner() {
            case data.BannerBlue: index = 63
            case data.BannerGreen: index = 64
            case data.BannerPurple: index = 65
            case data.BannerRed: index = 66
            case data.BannerYellow: index = 67
        }

        sparkle, _ := imageCache.GetImages("mapback.lbx", index)
        use := sparkle[counter % uint64(len(sparkle))]

        for _, point := range node.Zone {
            options2 := *options
            options2.GeoM.Translate(float64(point.X * tileWidth), float64(point.Y * tileHeight))
            screen.DrawImage(use, &options2)
        }
    }
}

func (node *ExtraMagicNode) Meld(meldingWizard Melder, spirit units.Unit) bool {
    if node.MeldingWizard == nil {
        node.MeldingWizard = meldingWizard
        if spirit.Equals(units.GuardianSpirit) {
            node.GuardianSpiritMeld = true
        } else {
            node.GuardianSpiritMeld = false
        }

        return true
    } else {
        // can't meld the same node twice
        if node.MeldingWizard == meldingWizard {
            return false
        }

        successful := true
        // 25% chance to meld if guardian spirit already melded it
        if node.GuardianSpiritMeld && rand.IntN(4) != 0 {
            successful = false
        }

        if successful {
            node.MeldingWizard = meldingWizard
            if spirit.Equals(units.GuardianSpirit) {
                node.GuardianSpiritMeld = true
            } else {
                node.GuardianSpiritMeld = false
            }

            return true
        }

        return false
    }
}

type FullTile struct {
    Extras map[ExtraKind]ExtraTile
    Tile terrain.Tile
    X int
    Y int
}

func (tile *FullTile) Valid() bool {
    return tile.Tile.Valid()
}

func (tile *FullTile) GetBonus() data.BonusType {
    bonus, ok := tile.Extras[ExtraKindBonus]
    if ok {
        return bonus.(*ExtraBonus).Bonus
    }

    return data.BonusNone
}

func (tile *FullTile) HasWildGame() bool {
    return tile.GetBonus() == data.BonusWildGame
}

type Map struct {
    Map *terrain.Map

    Plane data.Plane

    CityProvider CityProvider

    // contains information about map squares that contain extra features on top
    // such as a road, enchantment, encounter place (plane tower, lair, etc)
    ExtraMap map[image.Point]map[ExtraKind]ExtraTile

    Data *terrain.TerrainData

    TileCache map[int]*ebiten.Image

    miniMapPixels []byte
}

func getLandSize(size int) (int, int) {
    switch size {
        case 0: return 50, 50
        case 1: return 100, 100
        case 2: return 200, 150
    }

    return 100, 100
}

func MakeMap(terrainData *terrain.TerrainData, landSize int, magicSetting data.MagicSetting, difficulty data.DifficultySetting, plane data.Plane, cityProvider CityProvider) *Map {
    landWidth, landHeight := getLandSize(landSize)

    map_ := terrain.GenerateLandCellularAutomata(landWidth, landHeight, terrainData, plane)

    extraMap := make(map[image.Point]map[ExtraKind]ExtraTile)
    for x := range landWidth {
        for y := range landHeight {
            point := image.Pt(x, y)
            extraMap[point] = make(map[ExtraKind]ExtraTile)

            tile := terrainData.Tiles[map_.Terrain[x][y]].Tile
            switch tile.TerrainType() {
                case terrain.SorceryNode:
                    extraMap[point][ExtraKindMagicNode] = MakeMagicNode(MagicNodeSorcery, magicSetting, difficulty, plane)
                case terrain.NatureNode:
                    extraMap[point][ExtraKindMagicNode] = MakeMagicNode(MagicNodeNature, magicSetting, difficulty, plane)
                case terrain.ChaosNode:
                    extraMap[point][ExtraKindMagicNode] = MakeMagicNode(MagicNodeChaos, magicSetting, difficulty, plane)
            }
        }
    }

    canPlaceEncounter := func (x int, y int) bool {
        // an encounter can be placed if the specified tile is plain land, and is not near some other node/encounter

        tile := terrainData.Tiles[map_.Terrain[x][y]].Tile
        if !tile.IsLand() || tile.IsMagic() {
            return false
        }

        switch tile.TerrainType() {
            // FIXME: what types of terrain should we not place encounters on?
            case terrain.River: return false
        }

        // check that no surrounding tile is special
        for dx := -3; dx <= 3; dx++ {
            for dy := -3; dy <= 3; dy++ {
                cx := map_.WrapX(dx + x)
                cy := dy + y

                if cy < 0 || cy >= landHeight {
                    continue
                }

                extra, hasExtra := extraMap[image.Pt(cx, cy)]
                if hasExtra && extra != nil && len(extra) > 0 {
                    return false
                }
            }
        }

        return true
    }

    // returns a map of bonus types and the percent chance to get that bonus
    // https://masterofmagic.fandom.com/wiki/Mineral
    bonusTypeMap := func (x int, y int) map[data.BonusType]int {
        out := make(map[data.BonusType]int)

        cast := func (v float64) int {
            return int(v)
        }

        tile := terrainData.Tiles[map_.Terrain[x][y]].Tile
        switch tile.TerrainType() {
            case terrain.Hill:
                if plane == data.PlaneArcanus {
                    out[data.BonusIronOre] = cast(2 * 1000 / 6)
                    out[data.BonusCoal] = cast(1 * 1000 / 6)
                    out[data.BonusSilverOre] = cast(1.33 * 1000 / 6)
                    out[data.BonusGoldOre] = cast(1.33 * 1000 / 6)
                    out[data.BonusMithrilOre] = cast(0.33 * 1000 / 6)
                } else {
                    out[data.BonusIronOre] = cast(1 * 100 / 10)
                    out[data.BonusCoal] = cast(1 * 100 / 10)
                    out[data.BonusSilverOre] = cast(1 * 100 / 10)
                    out[data.BonusGoldOre] = cast(4 * 100 / 10)
                    out[data.BonusMithrilOre] = cast(2 * 100 / 10)
                    out[data.BonusAdamantiumOre] = cast(1 * 100 / 10)
                }
            case terrain.Forest:
                out[data.BonusWildGame] = 100
            case terrain.Mountain:
                if plane == data.PlaneArcanus {
                    out[data.BonusSilverOre] = cast(1 * 1000 / 6)
                    out[data.BonusGoldOre] = cast(1 * 1000 / 6)
                    out[data.BonusIronOre] = cast(1.33 * 1000 / 6)
                    out[data.BonusCoal] = cast(1.67 * 1000 / 6)
                    out[data.BonusMithrilOre] = cast(1 * 1000 / 6)
                } else {
                    out[data.BonusSilverOre] = cast(1 * 100 / 10)
                    out[data.BonusGoldOre] = cast(2 * 100 / 10)
                    out[data.BonusIronOre] = cast(1 * 100 / 10)
                    out[data.BonusCoal] = cast(1 * 100 / 10)
                    out[data.BonusMithrilOre] = cast(3 * 100 / 10)
                    out[data.BonusAdamantiumOre] = cast(2 * 100 / 10)
                }

                /*
            case terrain.Grass:
                if plane == data.PlaneArcanus {
                    out[data.BonusGoldOre] = 100
                } else {
                    out[data.BonusGoldOre] = int(1 * 100 / 2)
                    out[data.BonusCoal] = int(1 * 100 / 2)
                }
                */

            case terrain.Swamp:
                out[data.BonusNightshade] = 100
            case terrain.Desert:
                if plane == data.PlaneArcanus {
                    out[data.BonusGem] = cast(4 * 100 / 6)
                    out[data.BonusQuorkCrystal] = cast(2 * 100 / 6)
                } else {
                    out[data.BonusGem] = cast(2 * 100 / 10)
                    out[data.BonusQuorkCrystal] = cast(6 * 100 / 10)
                    out[data.BonusCrysxCrystal] = cast(2 * 100 / 10)
                }

            case terrain.Tundra: // none

            case terrain.Volcano, terrain.Lake, terrain.Ocean, terrain.River,
                terrain.Shore, terrain.NatureNode, terrain.SorceryNode, terrain.ChaosNode:
                // none

        }

        return out
    }

    canContainMineral := func (x int, y int) bool {
        tile := terrainData.Tiles[map_.Terrain[x][y]].Tile
        if !tile.IsLand() || tile.IsMagic() {
            return false
        }

        // FIXME: not 100% sure on this, can there be a bonus under a lair?
        _, hasLair := extraMap[image.Pt(x, y)][ExtraKindEncounter]
        if hasLair {
            return false
        }

        switch tile.TerrainType() {
            case terrain.Hill, terrain.Forest, terrain.Mountain,
                 terrain.Swamp, terrain.Desert: return true
        }

        return false
    }

    continents := map_.FindContinents()

    // place some encounter nodes down (lair, cave, etc)
    for i := range len(continents) {

        // try to place N encounters. if we can't place them all, then we just place as many as we can
        maxEncounters := len(continents[i]) / 10
        for _, index := range rand.Perm(len(continents[i])) {

            if maxEncounters == 0 {
                break
            }

            x, y := continents[i][index].X, continents[i][index].Y

            if canPlaceEncounter(x, y) {
                // log.Printf("Place encounter at %v, %v", x, y)
                extraMap[image.Pt(x, y)][ExtraKindEncounter] = makeEncounter(randomEncounterType(), difficulty, rand.N(2) == 0, plane)
                maxEncounters -= 1
            }
        }

        var candidates []image.Point
        for _, point := range continents[i] {
            if canContainMineral(point.X, point.Y) {
                candidates = append(candidates, point)
            }
        }

        fraction := 0.03
        if plane == data.PlaneMyrror {
            fraction = 0.07
        }

        maxBonuses := int(float64(len(candidates)) * fraction)
        // log.Printf("Candidates %v max bonuses %v", len(candidates), maxBonuses)
        for count, index := range rand.Perm(len(candidates)) {
            if count > maxBonuses {
                break
            }

            point := candidates[index]
            x, y := point.X, point.Y

            bonusTypes := bonusTypeMap(x, y)
            if len(bonusTypes) > 0 {
                chosen := chooseValue(bonusTypes)
                // if one didn't get picked because the values in the map don't add to 100 then just pick one randomly
                if chosen == data.BonusNone {
                    var choices []data.BonusType
                    for choice, _ := range bonusTypes {
                        choices = append(choices, choice)
                    }
                    chosen = choices[rand.N(len(choices))]
                }
                extraMap[point][ExtraKindBonus] = &ExtraBonus{Bonus: chosen}
            }
        }
    }

    return &Map{
        Data: terrainData,
        Map: map_,
        Plane: plane,
        TileCache: make(map[int]*ebiten.Image),
        ExtraMap: extraMap,
        CityProvider: cityProvider,
    }
}

// returns a map where for each direction, if the value is true then there is a road there
func (mapObject *Map) GetRoadNeighbors(x int, y int) map[Direction]bool {
    out := make(map[Direction]bool)

    convert := func(dx int, dy int) Direction {
        switch {
            case dx == -1 && dy == -1: return DirectionNorthWest
            case dx == -1 && dy == 0: return DirectionWest
            case dx == -1 && dy == 1: return DirectionSouthWest

            case dx == 1 && dy == -1: return DirectionNorthEast
            case dx == 1 && dy == 0: return DirectionEast
            case dx == 1 && dy == 1: return DirectionSouthEast

            case dx == 0 && dy == -1: return DirectionNorth
            case dx == 0 && dy == 1: return DirectionSouth
        }

        return DirectionNorth
    }

    for dx := -1; dx <= 1; dx++ {
        for dy := -1; dy <= 1; dy++ {
            // skip center
            if dx == 0 && dy == 0 {
                continue
            }

            out[convert(dx, dy)] = mapObject.ContainsRoad(x + dx, y + dy) || mapObject.CityProvider.ContainsCity(x + dx, y + dy, mapObject.Plane)
        }
    }

    return out
}

func (mapObject *Map) GetMeldedNodes(melder Melder) []*ExtraMagicNode {
    var out []*ExtraMagicNode

    for _, extras := range mapObject.ExtraMap {
        node, ok := extras[ExtraKindMagicNode]
        if ok {
            magic := node.(*ExtraMagicNode)
            if magic.MeldingWizard == melder {
                out = append(out, magic)
            }
        }
    }

    return out
}

func (mapObject *Map) SetBonus(x int, y int, bonus data.BonusType) {
    point := image.Pt(x, y)
    extras := mapObject.ExtraMap[point]
    extras[ExtraKindBonus] = &ExtraBonus{Bonus: bonus}
}

func getExtra[T any](extras map[ExtraKind]ExtraTile, kind ExtraKind) T {
    if obj, ok := extras[kind]; ok {
        if obj != nil {
            return obj.(T)
        }
    }

    var out T
    return out
}

func (mapObject *Map) SetRoad(x int, y int, enchanted bool) {
    mapObject.ExtraMap[image.Pt(x, y)][ExtraKindRoad] = &ExtraRoad{Map: mapObject, X: x, Y: y, Enchanted: enchanted}
}

func (mapObject *Map) ContainsRoad(x int, y int) bool {
    _, ok := mapObject.ExtraMap[image.Pt(x, y)][ExtraKindRoad]
    return ok
}

func (mapObject *Map) RemoveRoad(x int, y int) {
    delete(mapObject.ExtraMap[image.Pt(x, y)], ExtraKindRoad)
}

func (mapObject *Map) GetLair(x int, y int) *ExtraEncounter {
    return getExtra[*ExtraEncounter](mapObject.ExtraMap[image.Pt(x, y)], ExtraKindEncounter)
}

func (mapObject *Map) GetMagicNode(x int, y int) *ExtraMagicNode {
    return getExtra[*ExtraMagicNode](mapObject.ExtraMap[image.Pt(x, y)], ExtraKindMagicNode)
}

func (mapObject *Map) GetBonusTile(x int, y int) data.BonusType {
    bonus := getExtra[*ExtraBonus](mapObject.ExtraMap[image.Pt(x, y)], ExtraKindBonus)
    if bonus != nil {
        return bonus.Bonus
    }

    return data.BonusNone
}

func (mapObject *Map) CreateEncounter(x int, y int, encounterType EncounterType, difficulty data.DifficultySetting, weakStrength bool, plane data.Plane) bool {
    _, ok := mapObject.ExtraMap[image.Pt(x, y)]
    if ok {
        return false
    }

    mapObject.ExtraMap[image.Pt(x, y)][ExtraKindEncounter] = makeEncounter(encounterType, difficulty, weakStrength, plane)
    return true
}

func (mapObject *Map) CreateEncounterRandom(x int, y int, difficulty data.DifficultySetting, plane data.Plane) bool {
    return mapObject.CreateEncounter(x, y, randomEncounterType(), difficulty, rand.N(2) == 0, plane)
}

// for testing purposes
func (mapObject *Map) CreateNode(x int, y int, node MagicNode, plane data.Plane, magicSetting data.MagicSetting, difficulty data.DifficultySetting) *ExtraMagicNode {
    tileType := 0
    switch node {
        case MagicNodeNature: tileType = terrain.TileNatureForest.Index(plane)
        case MagicNodeSorcery: tileType = terrain.TileSorceryLake.Index(plane)
        case MagicNodeChaos: tileType = terrain.TileChaosVolcano.Index(plane)
    }

    mapObject.Map.Terrain[x][y] = tileType

    out := MakeMagicNode(node, magicSetting, difficulty, plane)

    mapObject.ExtraMap[image.Pt(x, y)][ExtraKindMagicNode] = out

    return out
}

func (mapObject *Map) Width() int {
    return mapObject.Map.Columns()
}

func (mapObject *Map) Height() int {
    return mapObject.Map.Rows()
}

func (mapObject *Map) TileWidth() int {
    return mapObject.Data.TileWidth()
}

func (mapObject *Map) TileHeight() int {
    return mapObject.Data.TileHeight()
}

func (mapObject *Map) WrapX(x int) int {
    return mapObject.Map.WrapX(x)
}

// return the shortest x distance between two points, taking into account the map wrapping
// result: WrapX(x1 + distance) = x2
func (mapObject *Map) XDistance(x1 int, x2 int) int {
    abs := func(x int) int {
        if x < 0 {
            return -x
        }

        return x
    }

    value := x2 - x1

    /*
    absValue := value
    if absValue < 0 {
        absValue = -absValue
    }
    */

    // cross over map boundary from x1 towards x2
    value2 := (mapObject.Map.Columns() - x1) + x2

    // cross over map boundary from x2 towards x1
    value3 := -x1 - (mapObject.Map.Columns() - x2)

    if abs(value) < abs(value2) && abs(value) < abs(value3) {
        return value
    } else if abs(value2) < abs(value3) && abs(value2) < abs(value) {
        return value2
    } else {
        return value3
    }
}

func (mapObject *Map) GetTile(tileX int, tileY int) FullTile {
    tileX = mapObject.WrapX(tileX)

    if tileX >= 0 && tileX < mapObject.Map.Columns() && tileY >= 0 && tileY < mapObject.Map.Rows() {
        tile := mapObject.Data.Tiles[mapObject.Map.Terrain[tileX][tileY]].Tile

        extras, ok := mapObject.ExtraMap[image.Pt(tileX, tileY)]
        if !ok {
            extras = make(map[ExtraKind]ExtraTile)
        }

        return FullTile{
            Tile: tile,
            X: tileX,
            Y: tileY,
            Extras: extras,
        }
    }

    return FullTile{
        Tile: terrain.InvalidTile(),
    }
}

func (mapObject *Map) GetTileImage(tileX int, tileY int, animationCounter uint64) (*ebiten.Image, error) {
    tile := mapObject.Map.Terrain[tileX][tileY]
    tileInfo := mapObject.Data.Tiles[tile]

    animationIndex := animationCounter % uint64(len(tileInfo.Images))

    if image, ok := mapObject.TileCache[tile * 0x1000 + int(animationIndex)]; ok {
        return image, nil
    }

    gpuImage := ebiten.NewImageFromImage(tileInfo.Images[animationCounter % uint64(len(tileInfo.Images))])

    mapObject.TileCache[tile * 0x1000 + int(animationIndex)] = gpuImage
    return gpuImage, nil
}

func (mapObject *Map) GetCatchmentArea(x int, y int) map[image.Point]FullTile {

    area := make(map[image.Point]FullTile)

    for dx := -2; dx <= 2; dx++ {
        for dy := -2; dy <= 2; dy++ {
            if int(math.Abs(float64(dx)) + math.Abs(float64(dy))) == 4 {
                continue
            }

            tileX := mapObject.WrapX(x + dx)
            tileY := y + dy

            tile := mapObject.GetTile(tileX, tileY)
            if tile.Valid() {
                area[image.Pt(tileX, tileY)] = tile
            }
        }
    }

    return area
}

func (mapObject *Map) TilesPerRow(screenWidth int) int {
    return int(math.Ceil(float64(screenWidth) / float64(mapObject.TileWidth())))
}

func (mapObject *Map) TilesPerColumn(screenHeight int) int {
    return int(math.Ceil(float64(screenHeight) / float64(mapObject.TileHeight())))
}

func bannerColor(banner data.BannerType) color.RGBA {
    switch banner {
        case data.BannerBlue: return color.RGBA{R: 0, G: 0, B: 255, A: 255}
        case data.BannerGreen: return color.RGBA{R: 0, G: 255, B: 0, A: 255}
        case data.BannerPurple: return color.RGBA{R: 255, G: 0, B: 255, A: 255}
        case data.BannerRed: return color.RGBA{R: 255, G: 0, B: 0, A: 255}
        case data.BannerYellow: return color.RGBA{R: 255, G: 255, B: 0, A: 255}
        case data.BannerBrown: return color.RGBA{R: 0xdb, G: 0x7e, B: 0x1f, A: 255}
    }

    return color.RGBA{R: 0, G: 0, B: 0, A: 255}
}

func (mapObject *Map) DrawMinimap(screen *ebiten.Image, cities []MiniMapCity, centerX int, centerY int, zoom float64, fog [][]bool, counter uint64, crosshairs bool){
    if len(mapObject.miniMapPixels) != screen.Bounds().Dx() * screen.Bounds().Dy() * 4 {
        mapObject.miniMapPixels = make([]byte, screen.Bounds().Dx() * screen.Bounds().Dy() * 4)
    }

    rowSize := screen.Bounds().Dx()

    cameraX := centerX - screen.Bounds().Dx() / 2
    cameraY := centerY - screen.Bounds().Dy() / 2

    /*
    if cameraX < 0 {
        cameraX = 0
    }
    */

    if cameraY < 0 {
        cameraY = 0
    }

    /*
    if cameraX > mapObject.Map.Columns() - screen.Bounds().Dx() {
        cameraX = mapObject.Map.Columns() - screen.Bounds().Dx()
    }
    */

    if cameraY > mapObject.Map.Rows() - screen.Bounds().Dy() {
        cameraY = mapObject.Map.Rows() - screen.Bounds().Dy()
    }

    set := func(x int, y int, c color.RGBA){
        baseIndex := (y * rowSize + x) * 4

        /*
        if baseIndex > len(mapObject.miniMapPixels) {
            return
        }
        */

        r, g, b, a := c.RGBA()

        mapObject.miniMapPixels[baseIndex + 0] = byte(r >> 8)
        mapObject.miniMapPixels[baseIndex + 1] = byte(g >> 8)
        mapObject.miniMapPixels[baseIndex + 2] = byte(b >> 8)
        mapObject.miniMapPixels[baseIndex + 3] = byte(a >> 8)
    }

    black := color.RGBA{R: 0, G: 0, B: 0, A: 255}

    cityLocations := make(map[image.Point]color.RGBA)

    for _, city := range cities {
        if fog[city.GetX()][city.GetY()] {
            cityLocations[image.Pt(city.GetX(), city.GetY())] = bannerColor(city.GetBanner())
        }
    }

    for x := 0; x < screen.Bounds().Dx(); x++ {
        for y := 0; y < screen.Bounds().Dy(); y++ {
            tileX := mapObject.WrapX(x + cameraX)
            tileY := y + cameraY

            if tileX < 0 || tileX >= mapObject.Map.Columns() || tileY < 0 || tileY >= mapObject.Map.Rows() || !fog[tileX][tileY] {
                set(x, y, black)
                continue
            }

            var use color.RGBA

            landColor := color.RGBA{R: 0, G: 0xad, B: 0x00, A: 255}

            switch terrain.GetTile(mapObject.Map.Terrain[tileX][tileY]).TerrainType() {
                case terrain.Grass: use = landColor
                case terrain.Ocean: use = color.RGBA{R: 0, G: 0, B: 255, A: 255}
                case terrain.River: use = color.RGBA{R: 0x3f, G: 0x88, B: 0xd3, A: 255}
                case terrain.Shore: use = landColor
                case terrain.Mountain: use = color.RGBA{R: 0xbc, G: 0xd0, B: 0xe4, A: 255}
                case terrain.Hill: use = landColor
                case terrain.Swamp: use = landColor
                case terrain.Forest: use = landColor
                case terrain.Desert: use = color.RGBA{R: 0xdb, G: 0xbd, B: 0x29, A: 255}
                case terrain.Tundra: use = color.RGBA{R: 0xd6, G: 0xd4, B: 0xc9, A: 255}
                case terrain.Volcano: use = color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 255}
                case terrain.Lake: use = color.RGBA{R: 0x3f, G: 0x88, B: 0xd3, A: 255}
                case terrain.NatureNode: use = color.RGBA{R: 0, G: 255, B: 0, A: 255}
                case terrain.SorceryNode: use = color.RGBA{R: 0, G: 0, B: 255, A: 255}
                case terrain.ChaosNode: use = color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 255}
                default: use = color.RGBA{R: 64, G: 64, B: 64, A: 255}
            }

            if cityColor, ok := cityLocations[image.Pt(tileX, tileY)]; ok {
                use = cityColor
            }

            set(x, y, use)
        }
    }

    if crosshairs {
        cursorColorBlue := math.Sin(float64(counter) / 10.0) * 127.0 + 127.0
        if cursorColorBlue > 255 {
            cursorColorBlue = 255
        }
        cursorColor := util.PremultiplyAlpha(color.RGBA{R: 255, G: 255, B: byte(cursorColorBlue), A: 180})

        cursorRadius := int(5.0 / zoom)
        x1 := centerX - cursorRadius - cameraX
        y1 := centerY - cursorRadius - cameraY
        x2 := centerX + cursorRadius - cameraX
        y2 := y1
        x3 := x1
        y3 := centerY + cursorRadius - cameraY
        x4 := x2
        y4 := y3
        points := []image.Point{
            image.Pt(x1, y1),
            image.Pt(x1+1, y1),
            image.Pt(x1, y1+1),

            image.Pt(x2, y2),
            image.Pt(x2-1, y2),
            image.Pt(x2, y2+1),

            image.Pt(x3, y3),
            image.Pt(x3+1, y3),
            image.Pt(x3, y3-1),

            image.Pt(x4, y4),
            image.Pt(x4-1, y4),
            image.Pt(x4, y4-1),
        }

        for _, point := range points {
            x := mapObject.WrapX(point.X)
            y := point.Y

            if x >= 0 && y >= 0 && x < screen.Bounds().Dx() && y < screen.Bounds().Dy(){
                set(x, y, cursorColor)
            }
        }
    }

    screen.WritePixels(mapObject.miniMapPixels)
}

// draw base map tiles, in general stuff that should go under cities/units
func (mapObject *Map) DrawLayer1(camera cameralib.Camera, animationCounter uint64, imageCache *util.ImageCache, screen *ebiten.Image, geom ebiten.GeoM){
    tileWidth := mapObject.TileWidth()
    tileHeight := mapObject.TileHeight()

    /*
    tilesPerRow := mapObject.TilesPerRow(screen.Bounds().Dx())
    tilesPerColumn := mapObject.TilesPerColumn(screen.Bounds().Dy())
    */

    var options ebiten.DrawImageOptions

    minX, minY, maxX, maxY := camera.GetTileBounds()

    // draw all tiles first
    for x := minX; x < maxX; x++ {
        for y := minY; y < maxY; y++ {
            tileX := mapObject.WrapX(x)
            tileY := y

            // for debugging
            // util.DrawRect(screen, image.Rect(x * tileWidth, y * tileHeight, (x + 1) * tileWidth, (y + 1) * tileHeight), color.RGBA{R: 255, G: 0, B: 0, A: 255})

            if tileX < 0 || tileX >= mapObject.Map.Columns() || tileY < 0 || tileY >= mapObject.Map.Rows() {
                continue
            }

            tileImage, err := mapObject.GetTileImage(tileX, tileY, animationCounter)
            if err == nil {
                options.GeoM.Reset()
                // options.GeoM = geom
                // options.GeoM.Reset()
                options.GeoM.Translate(float64(x * tileWidth), float64(y * tileHeight))
                options.GeoM.Concat(geom)

                screen.DrawImage(tileImage, &options)

                for _, extraKind := range ExtraDrawOrder {
                    extra, ok := mapObject.ExtraMap[image.Pt(tileX, tileY)][extraKind]
                    if ok {
                        extra.DrawLayer1(screen, imageCache, &options, animationCounter, tileWidth, tileHeight)
                    }
                }
            } else {
                log.Printf("Unable to render tile at %d, %d: %v", tileX, tileY, err)
            }
        }
    }
}

// give the extra nodes a chance to draw on top of cities/units, but still under the fog
func (mapObject *Map) DrawLayer2(cameraX int, cameraY int, animationCounter uint64, imageCache *util.ImageCache, screen *ebiten.Image, geom ebiten.GeoM){
    tileWidth := mapObject.TileWidth()
    tileHeight := mapObject.TileHeight()

    /*
    tilesPerRow := mapObject.TilesPerRow(screen.Bounds().Dx())
    tilesPerColumn := mapObject.TilesPerColumn(screen.Bounds().Dy())
    */

    var options ebiten.DrawImageOptions

    // then draw all extra nodes on top
    x_loop:
    for x := 0; x < mapObject.Map.Columns(); x++ {
        y_loop:
        for y := 0; y < mapObject.Map.Rows(); y++ {
            options.GeoM.Reset()
            options.GeoM.Translate(float64(x * tileWidth), float64(y * tileHeight))
            options.GeoM.Concat(geom)

            posX, posY := options.GeoM.Apply(0, 0)
            if int(posX) > screen.Bounds().Max.X {
                break x_loop
            }
            if int(posY) > screen.Bounds().Max.Y {
                break y_loop
            }

            tileX := mapObject.WrapX(x)
            tileY := y

            if tileX < 0 || tileX >= mapObject.Map.Columns() || tileY < 0 || tileY >= mapObject.Map.Rows() {
                continue
            }

            for _, extraKind := range ExtraDrawOrder {
                extra, ok := mapObject.ExtraMap[image.Pt(tileX, tileY)][extraKind]
                if ok {
                    extra.DrawLayer2(screen, imageCache, &options, animationCounter, tileWidth, tileHeight)
                }
            }
        }
    }
}
