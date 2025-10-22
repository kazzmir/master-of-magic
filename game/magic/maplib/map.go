package maplib

import (
    "log"
    "math"
    "math/rand/v2"
    "image"
    "image/color"
    "slices"

    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/lib/functional"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/shaders"
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

type Wizard interface {
    GetBanner() data.BannerType
}

type MagicNode int
const (
    MagicNodeNature MagicNode = iota
    MagicNodeSorcery
    MagicNodeChaos

    MagicNodeNone
)

func (magicNode MagicNode) MagicType() data.MagicType {
    switch magicNode {
        case MagicNodeNature: return data.NatureMagic
        case MagicNodeSorcery: return data.SorceryMagic
        case MagicNodeChaos: return data.ChaosMagic
    }

    return data.MagicNone
}

func (magicNode MagicNode) Name() string {
    switch magicNode {
        case MagicNodeNature: return "Nature Node"
        case MagicNodeSorcery: return "Sorcery Node"
        case MagicNodeChaos: return "Chaos Node"
    }
    return ""
}

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
    ExtraKindOpenTower
    ExtraKindVolcano
    ExtraKindCorruption
)

var ExtraDrawOrder = []ExtraKind{
    ExtraKindVolcano,
    ExtraKindCorruption,
    ExtraKindRoad,
    ExtraKindBonus,
    ExtraKindMagicNode,
    ExtraKindEncounter,
    ExtraKindOpenTower,
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
                scale.DrawScaled(screen, pic, options)
            }

            connected = true
        }
    }

    if !connected {
        pics, err := imageCache.GetImages("mapback.lbx", baseIndex)
        if err == nil {
            pic := pics[counter % uint64(len(pics))]
            scale.DrawScaled(screen, pic, options)
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
    index := bonus.Bonus.LbxIndex()

    if index == -1 {
        return
    }

    pic, err := imageCache.GetImage("mapback.lbx", index, 0)
    if err == nil {
        scale.DrawScaled(screen, pic, options)
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
    EncounterTypeAncientTemple
    EncounterTypeFallenTemple
    EncounterTypeRuins
    EncounterTypeAbandonedKeep
    EncounterTypeDungeon
    EncounterTypeChaosNode
    EncounterTypeNatureNode
    EncounterTypeSorceryNode
)

func (encounterType EncounterType) Name() string {
    switch encounterType {
        case EncounterTypeLair: return "Monster Lair"
        case EncounterTypeCave: return "Mysterious Cave"
        case EncounterTypePlaneTower: return "Tower"
        case EncounterTypeAncientTemple: return "Ancient Temple"
        case EncounterTypeFallenTemple: return "Fallen Temple"
        case EncounterTypeRuins: return "Ruins"
        case EncounterTypeAbandonedKeep: return "Abandoned Keep"
        case EncounterTypeDungeon: return "Dungeon"
        case EncounterTypeChaosNode: return "Chaos Node"
        case EncounterTypeNatureNode: return "Nature Node"
        case EncounterTypeSorceryNode: return "Sorcery Node"
    }

    return ""
}

func (encounterType EncounterType) ShortName() string {
    switch encounterType {
        case EncounterTypeLair: return "Lair"
        case EncounterTypeCave: return "Cave"
        case EncounterTypePlaneTower: return "Tower"
        case EncounterTypeAncientTemple: return "Temple"
        case EncounterTypeFallenTemple: return "Temple"
        case EncounterTypeRuins: return "Ruins"
        case EncounterTypeAbandonedKeep: return "Keep"
        case EncounterTypeDungeon: return "Dungeon"
        case EncounterTypeChaosNode: return "Node"
        case EncounterTypeNatureNode: return "Node"
        case EncounterTypeSorceryNode: return "Node"
    }

    return ""
}

func (encounterType EncounterType) LbxIndex() int {
    switch encounterType {
        case EncounterTypeLair, EncounterTypeCave: return 71
        case EncounterTypePlaneTower: return 69
        case EncounterTypeAncientTemple: return 72
        case EncounterTypeFallenTemple: return 75
        case EncounterTypeRuins: return 74
        case EncounterTypeAbandonedKeep: return 73
        case EncounterTypeDungeon: return 74
    }

    return -1
}

func randomEncounterType() EncounterType {
    all := []EncounterType{
        EncounterTypeLair,
        EncounterTypeCave,
        EncounterTypeAncientTemple,
        EncounterTypeFallenTemple,
        EncounterTypeRuins,
        EncounterTypeAbandonedKeep,
        EncounterTypeDungeon,
    }

    return all[rand.N(len(all))]
}

// lair, plane tower, etc and also the initial enemies on nodes
type ExtraEncounter struct {
    Type EncounterType
    Units []units.Unit
    Budget int // used for treasure
    ExploredBy *set.Set[Wizard]
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
        if plane == data.PlaneArcanus || encounterType == EncounterTypePlaneTower {
            budget = (rand.N(20) + 1) * 30
        } else {
            budget = (rand.N(30) + 1) * 30
        }
    } else {
        if plane == data.PlaneArcanus || encounterType == EncounterTypePlaneTower {
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
        Budget: budget,
        Units: append(guardians, secondary...),
        ExploredBy: set.MakeSet[Wizard](),
    }
}

func (extra *ExtraEncounter) DrawLayer1(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int){
    index := extra.Type.LbxIndex()

    if index == -1 {
        return
    }

    pic, err := imageCache.GetImage("mapback.lbx", index, 0)
    if err == nil {
        scale.DrawScaled(screen, pic, options)
    }
}

func (extra *ExtraEncounter) DrawLayer2(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int){
    // nuthin'
}

type ExtraOpenTower struct {
}

func (extra *ExtraOpenTower) DrawLayer1(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int){
    index := 70

    pic, err := imageCache.GetImage("mapback.lbx", index, 0)
    if err == nil {
        scale.DrawScaled(screen, pic, options)
    }
}

func (extra *ExtraOpenTower) DrawLayer2(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int){
}

type ExtraMagicNode struct {
    Kind MagicNode

    // list of points that are affected by the node
    // this could be a map to be more optimal for lookups
    Zone []image.Point

    // if this node is melded, then this player receives the power
    MeldingWizard Wizard
    // true if melded by a guardian spirit, otherwise false if melded by a magic spirit
    GuardianSpiritMeld bool

    Warped bool
    // the wizard that cast warp node (so it can be dispelled)
    WarpedOwner Wizard
}

func (node *ExtraMagicNode) ContainsPoint(x int, y int) bool {
    return slices.Contains(node.Zone, image.Pt(x, y))
}

func (node *ExtraMagicNode) DrawLayer1(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int){
}

func (node *ExtraMagicNode) DrawLayer2(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int){
    // if the node is melded then show the zone of influence with the sparkly images

    if node.MeldingWizard != nil {
        index := 63
        switch node.MeldingWizard.GetBanner() {
            case data.BannerBlue: index = 63
            case data.BannerGreen: index = 64
            case data.BannerPurple: index = 65
            case data.BannerRed: index = 66
            case data.BannerYellow: index = 67
        }

        sparkle, err := imageCache.GetImages("mapback.lbx", index)
        if err != nil {
            return
        }

        for i, point := range node.Zone {
            use := sparkle[(counter + uint64(i)) % uint64(len(sparkle))]

            options2 := *options
            options2.GeoM.Reset()
            options2.GeoM.Translate(float64(point.X * tileWidth), float64(point.Y * tileHeight))
            options2.GeoM.Concat(options.GeoM)
            scale.DrawScaled(screen, use, &options2)
        }
    }

    if node.Warped && node.MeldingWizard != nil {
        shader, err := imageCache.GetShader(shaders.ShaderWarp)
        if err != nil {
            log.Printf("error getting shader %v", err)
            return
        }

        // FIXME get the mask to work, but scale it based on what the options.GeoM has
        /*
        mask, err := imageCache.GetImage("mapback.lbx", 93, 0)
        if err != nil {
            return
        }
        */

        x1, y1 := options.GeoM.Apply(0, 0)
        x2, y2 := options.GeoM.Apply(float64(tileWidth), float64(tileHeight))

        // log.Printf("x1 %v y1 %v x2 %v y2 %v", x1, y1, x2, y2)

        if x1 < 0 || y1 < 0 || x2 <= x1 || y2 <= y1 {
            return
        }
        rect := image.Rect(int(x1), int(y1), int(x2), int(y2))
        image := screen.SubImage(scale.ScaleRect(rect))
        if image.Bounds().Dx() == 0 || image.Bounds().Dy() == 0 {
            return
        }

        // the mask needs to be the same size as the subimage. This is a hack for now
        /*
        mask := ebiten.NewImage(image.Bounds().Dx(), image.Bounds().Dy())
        mask.Fill(color.RGBA{0, 0, 0, 255})
        */

        sourceImage := ebiten.NewImageFromImage(image)

        // log.Printf("warp at %v, %v bounds image=%v source=%v", point.X, point.Y, image.Bounds(), sourceImage.Bounds())

        var options2 ebiten.DrawRectShaderOptions
        options2.GeoM.Translate(scale.Scale(x1), scale.Scale(y1))

        options2.Images[0] = sourceImage
        // options2.Images[1] = mask
        options2.Uniforms = make(map[string]interface{})
        options2.Uniforms["Time"] = float32(math.Abs(float64(counter/5)))
        screen.DrawRectShader(image.Bounds().Dx(), image.Bounds().Dy(), shader, &options2)
    }
}

func (node *ExtraMagicNode) Meld(meldingWizard Wizard, spirit units.Unit) bool {
    if node.Warped {
        return false
    } else if node.MeldingWizard == nil {
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

func (node *ExtraMagicNode) GetPower(magicBonus float64) float64 {
    if node.Warped {
        return -5
    }

    return float64(len(node.Zone)) * magicBonus
}

type ExtraVolcano struct {
    CastingWizard Wizard
}

func (node *ExtraVolcano) DrawLayer1(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int){
}

func (node *ExtraVolcano) DrawLayer2(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int){
}

type ExtraCorruption struct {
}

func (node *ExtraCorruption) DrawLayer1(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int){
    pic, err := imageCache.GetImage("mapback.lbx", 77, 0)
    if err == nil {
        scale.DrawScaled(screen, pic, options)
    }
}

func (node *ExtraCorruption) DrawLayer2(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int){
}

type FullTile struct {
    Extras map[ExtraKind]ExtraTile
    Tile terrain.Tile
    X int
    Y int
    IsHalf bool
}

func (tile *FullTile) Name(mapObject *Map) string {
    if tile.IsRiverMouth(mapObject) {
        return "River Mouth"
    }

    return tile.Tile.Name()
}

func (tile *FullTile) Valid() bool {
    return tile.Tile.Valid()
}

func (tile *FullTile) Corrupted() bool {
    _, ok := tile.Extras[ExtraKindCorruption]
    return ok
}

func (tile *FullTile) HasEncounter() bool {
    _, ok := tile.Extras[ExtraKindEncounter]
    return ok
}

func (tile *FullTile) FoodBonus() fraction.Fraction {
    if tile.Corrupted() {
        return fraction.Zero()
    }

    if tile.Tile.IsLakeWithFlow() {
        return fraction.FromInt(2)
    }
    // FIXME: Shore with two river deltas = 2
    // FIXME: Shore with single river delta surrounded by ocean = 2

    switch tile.Tile.TerrainType() {
        case terrain.Ocean: return fraction.Zero()
        case terrain.Grass: return fraction.Make(3, 2)
        case terrain.Forest: return fraction.Make(1, 2)
        case terrain.Mountain: return fraction.Zero()
        case terrain.Desert: return fraction.Zero()
        case terrain.Swamp: return fraction.Zero()
        case terrain.Tundra: return fraction.Zero()
        case terrain.SorceryNode: return fraction.FromInt(2)
        case terrain.NatureNode: return fraction.Make(5, 2)
        case terrain.ChaosNode: return fraction.Zero()
        case terrain.Hill: return fraction.Make(1, 2)
        case terrain.Volcano: return fraction.Zero()
        case terrain.Lake: return fraction.Zero()
        case terrain.River: return fraction.FromInt(2)
        case terrain.Shore: return fraction.Make(1, 2)
    }

    return fraction.Zero()
}

// percent bonus increase, 3 = 3%
func (tile *FullTile) GoldBonus(mapObject *Map) int {
    if tile.Corrupted() {
        return 0
    }

    switch {
        case tile.IsRiverMouth(mapObject): return 30
        case tile.Tile.TerrainType() == terrain.River: return 20
        case tile.IsTouchingShore(mapObject, false): return 10
    }

    return 0
}

// percent bonus increase, 3 = 3%
func (tile *FullTile) ProductionBonus(hasGaiasBlessing bool) int {
    if tile.Corrupted() {
        return 0
    }

    terrainType := tile.Tile.TerrainType()
    switch {
        case terrainType == terrain.Forest && hasGaiasBlessing: return 6
        case terrainType == terrain.Forest: return 3
        case terrainType == terrain.Mountain: return 5
        case terrainType == terrain.Desert: return 3
        case terrainType == terrain.NatureNode: return 3
        case terrainType == terrain.ChaosNode: return 5
        case terrainType == terrain.Hill: return 3
    }

    return 0
}

func (tile *FullTile) IsRiverMouth(mapObject *Map) bool {
    return tile.Tile.TerrainType() == terrain.River && tile.IsTouchingShore(mapObject, true)
}

func (tile *FullTile) IsTouchingShore(mapObject *Map, cardinal bool) bool {
    for dx := -1; dx <= 1; dx++ {
        for dy := -1; dy <= 1; dy++ {
            if dx == 0 && dy == 0 {
                continue
            }

            if cardinal && dx != 0 && dy != 0 {
                continue
            }

            tile := mapObject.GetTile(tile.X + dx, tile.Y + dy)
            if tile.Tile.IsShore() {
                return true
            }
        }
    }

    return false
}

func (tile *FullTile) IsByRiver(mapObject *Map) bool {
    for dx := -1; dx <= 1; dx++ {
        for dy := -1; dy <= 1; dy++ {
            tile := mapObject.GetTile(tile.X + dx, tile.Y + dy)
            if tile.Tile.IsRiver() {
                return true
            }
        }
    }

    return false
}

func (tile *FullTile) GetBonus() data.BonusType {
    if tile.Corrupted() {
        return data.BonusNone
    }

    bonus, ok := tile.Extras[ExtraKindBonus]
    if ok {
        return bonus.(*ExtraBonus).Bonus
    }

    return data.BonusNone
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
    // miniMapImage *image.Paletted
}

func getLandSize(size int) (int, int) {
    switch size {
        case 0: return 50, 50
        case 1: return 100, 100
        case 2: return 200, 150
    }

    return 100, 100
}

func MakeMap(terrainData *terrain.TerrainData, landSize int, magicSetting data.MagicSetting, difficulty data.DifficultySetting, plane data.Plane, cityProvider CityProvider, planeTowers []image.Point) *Map {
    landWidth, landHeight := getLandSize(landSize)

    map_ := terrain.GenerateLandCellularAutomata(landWidth, landHeight, terrainData, plane)

    extraMap := make(map[image.Point]map[ExtraKind]ExtraTile)

    // place towers, and then re-resolve tiles
    for _, point := range planeTowers {
        map_.Terrain[point.X][point.Y] = terrain.GetTile(terrain.IndexGrass1).Index(plane)
        extraMap[point] = make(map[ExtraKind]ExtraTile)
        extraMap[point][ExtraKindEncounter] = makeEncounter(EncounterTypePlaneTower, difficulty, false, plane)
    }

    map_.ResolveTiles(terrainData, plane)

    for x := range landWidth {
        for y := range landHeight {
            point := image.Pt(x, y)
            _, exists := extraMap[point]
            if !exists {
                extraMap[point] = make(map[ExtraKind]ExtraTile)
            }

            tile := terrainData.Tiles[map_.Terrain[x][y]].Tile
            switch tile.TerrainType() {
                case terrain.SorceryNode:
                    extraMap[point][ExtraKindMagicNode], extraMap[point][ExtraKindEncounter] = MakeMagicNode(MagicNodeSorcery, magicSetting, difficulty, plane)
                case terrain.NatureNode:
                    extraMap[point][ExtraKindMagicNode], extraMap[point][ExtraKindEncounter] = MakeMagicNode(MagicNodeNature, magicSetting, difficulty, plane)
                case terrain.ChaosNode:
                    extraMap[point][ExtraKindMagicNode], extraMap[point][ExtraKindEncounter] = MakeMagicNode(MagicNodeChaos, magicSetting, difficulty, plane)
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
        maxEncounters := continents[i].Size() / 10
        points := continents[i].Values()
        for _, index := range rand.Perm(len(points)) {

            if maxEncounters == 0 {
                break
            }

            x, y := points[index].X, points[index].Y

            if canPlaceEncounter(x, y) {
                // log.Printf("Place encounter at %v, %v", x, y)
                extraMap[image.Pt(x, y)][ExtraKindEncounter] = makeEncounter(randomEncounterType(), difficulty, rand.N(2) == 0, plane)
                maxEncounters -= 1
            }
        }

        var candidates []image.Point
        for _, point := range points {
            if canContainMineral(point.X, point.Y) {
                candidates = append(candidates, point)
            }
        }

        fraction := 0.06
        if plane == data.PlaneMyrror {
            fraction = 0.10
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

func GeneratePlaneTowerPositions(landSize int, count int) []image.Point {
    width, height := getLandSize(landSize)

    var out []image.Point

    for range count {

        success := false
        tries := 100
        for !success && tries > 0 {
            tries -= 1
            x := rand.N(width)
            y := rand.N(height)

            // not too close to the edges
            if y < 3 || y >= height - 3 {
                continue
            }

            point := image.Pt(x, y)

            ok := true
            for _, other := range out {
                vector := other.Sub(point)
                distance := math.Sqrt(float64(vector.X * vector.X + vector.Y * vector.Y))
                if distance < 10 {
                    ok = false
                    break
                }
            }

            if ok {
                out = append(out, point)
                success = true
            }
        }
    }

    return out
}

func (mapObject *Map) GetContinents() []terrain.Continent {
    return mapObject.Map.FindContinents()
}

// get all the tiles that are part of the continent that contains the given x, y
func (mapObject *Map) GetContinentTiles(x int, y int) []FullTile {
    continent := mapObject.Map.FindContinent(x, y)
    var out []FullTile
    for _, point := range continent.Values() {
        tile := mapObject.GetTile(point.X, point.Y)
        out = append(out, tile)
    }
    return out
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

func (mapObject *Map) GetMeldedNodes(melder Wizard) []*ExtraMagicNode {
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

func (mapObject *Map) GetCastedVolcanoes(caster Wizard) []*ExtraVolcano {
    var out []*ExtraVolcano

    for _, extras := range mapObject.ExtraMap {
        extra, ok := extras[ExtraKindVolcano]
        if ok {
            volcano := extra.(*ExtraVolcano)
            if volcano.CastingWizard == caster {
                out = append(out, volcano)
            }
        }
    }

    return out
}

func (mapObject *Map) RemoveBonus(x int, y int) {
    delete(mapObject.ExtraMap[image.Pt(x, y)], ExtraKindBonus)
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

func (mapObject *Map) GetEncounter(x int, y int) *ExtraEncounter {
    return getExtra[*ExtraEncounter](mapObject.ExtraMap[image.Pt(x, y)], ExtraKindEncounter)
}

func (mapObject *Map) GetEncounterLocations() []image.Point {
    var out []image.Point

    for point, extras := range mapObject.ExtraMap {
        _, exists := extras[ExtraKindEncounter]
        if exists {
            out = append(out, point)
        }
    }

    return out
}

func (mapObject *Map) RemoveEncounter(x int, y int) {
    _, exists := mapObject.ExtraMap[image.Pt(x, y)]
    if !exists {
        return
    }

    _, exists = mapObject.ExtraMap[image.Pt(x, y)][ExtraKindEncounter]
    if exists {
        delete(mapObject.ExtraMap[image.Pt(x, y)], ExtraKindEncounter)
    }
}

// an open plane tower
func (mapObject *Map) SetPlaneTower(x int, y int) {
    mapObject.ExtraMap[image.Pt(x, y)][ExtraKindOpenTower] = &ExtraOpenTower{}
}

func (mapObject *Map) HasOpenTower(x int, y int) bool {
    _, exists := mapObject.ExtraMap[image.Pt(x, y)][ExtraKindOpenTower]
    return exists
}

func (mapObject *Map) HasMagicNode(x int, y int) bool {
    _, exists := mapObject.ExtraMap[image.Pt(x, y)][ExtraKindMagicNode]
    return exists
}

func (mapObject *Map) GetMagicNode(x int, y int) *ExtraMagicNode {
    return getExtra[*ExtraMagicNode](mapObject.ExtraMap[image.Pt(x, y)], ExtraKindMagicNode)
}

// return the node that contains x/y in its influence zone
// this is a bit slow in that it checks the entire map
func (mapObject* Map) GetMagicInfluence(x int, y int) *ExtraMagicNode {
    for point, extras := range mapObject.ExtraMap {
        magicNode := getExtra[*ExtraMagicNode](extras, ExtraKindMagicNode)
        if magicNode != nil && magicNode.ContainsPoint(x - point.X, y - point.Y) {
            return magicNode
        }
    }

    return nil
}

func (mapObject *Map) SetVolcano(x int, y int, caster Wizard) {
    mapObject.ExtraMap[image.Pt(x, y)][ExtraKindVolcano] = &ExtraVolcano{CastingWizard: caster}
    mapObject.Map.SetTerrainAt(x, y, terrain.Volcano, mapObject.Data, mapObject.Plane)
}

func (mapObject *Map) HasVolcano(x int, y int) bool {
    _, exists := mapObject.ExtraMap[image.Pt(x, y)][ExtraKindVolcano]
    return exists
}

func (mapObject *Map) RemoveVolcano(x int, y int) {
    _, exists := mapObject.ExtraMap[image.Pt(x, y)][ExtraKindVolcano]
    if exists {
        delete(mapObject.ExtraMap[image.Pt(x, y)], ExtraKindVolcano)

        mapObject.Map.SetTerrainAt(x, y, terrain.Mountain, mapObject.Data, mapObject.Plane)

        // chance of 5% to generate a mineral
        if rand.N(100) < 5 {
            choices := []data.BonusType{data.BonusSilverOre, data.BonusGoldOre, data.BonusIronOre, data.BonusCoal, data.BonusMithrilOre}
            if mapObject.Plane == data.PlaneMyrror {
                choices = append(choices, data.BonusAdamantiumOre)
            }
            mapObject.SetBonus(x, y, choices[rand.IntN(len(choices))])
        }
    }
}

func (mapObject *Map) GetBonusTile(x int, y int) data.BonusType {
    bonus := getExtra[*ExtraBonus](mapObject.ExtraMap[image.Pt(x, y)], ExtraKindBonus)
    if bonus != nil {
        return bonus.Bonus
    }

    return data.BonusNone
}

func (mapObject *Map) CreateEncounter(x int, y int, encounterType EncounterType, difficulty data.DifficultySetting, weakStrength bool, plane data.Plane) bool {
    if mapObject.GetEncounter(x, y) != nil {
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

    out, _ := MakeMagicNode(node, magicSetting, difficulty, plane)

    mapObject.ExtraMap[image.Pt(x, y)][ExtraKindMagicNode] = out

    return out
}

func (mapObject *Map) HasCorruption(x int, y int) bool {
    _, exists := mapObject.ExtraMap[image.Pt(x, y)]
    if exists {
        _, exists = mapObject.ExtraMap[image.Pt(x, y)][ExtraKindCorruption]
        return exists
    }
    return false
}

func (mapObject *Map) SetCorruption(x int, y int) {
    mapObject.ExtraMap[image.Pt(x, y)][ExtraKindCorruption] = &ExtraCorruption{}
}

func (mapObject *Map) RemoveCorruption(x int, y int) {
    if mapObject.HasCorruption(x, y) {
        delete(mapObject.ExtraMap[image.Pt(x, y)], ExtraKindCorruption)
    }
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

// distance in tiles between two points
func (mapObject *Map) TileDistance(x1 int, y1 int, x2 int, y2 int) int {
    dx := mapObject.XDistance(x1, x2)
    dy := y2 - y1

    return int(math.Sqrt(float64(dx * dx + dy * dy)))
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

func (mapObject *Map) ResetCache() {
    mapObject.TileCache = make(map[int]*ebiten.Image)
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

func (mapObject *Map) OnShore(x int, y int) bool {
    tile := mapObject.GetTile(x, y)
    return tile.IsTouchingShore(mapObject, false)
}

func (mapObject *Map) ByRiver(x int, y int) bool {
    tile := mapObject.GetTile(x, y)
    return tile.IsByRiver(mapObject)
}

func (mapObject *Map) GetGoldBonus(x int, y int) int {
    tile := mapObject.GetTile(x, y)
    return tile.GoldBonus(mapObject)
}

// Returns city catchment area (5x5 square minus the corners)
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

// Returns a square with a side of (radius * 2 + 1), and x, y in the center. May be useful for some spells or area effects
// Radius has to be >= 0
func (mapObject *Map) GetSquareAreaAround(x int, y int, radius int) map[image.Point]FullTile {

    area := make(map[image.Point]FullTile)

    for dx := -radius; dx <= radius; dx++ {
        for dy := -radius; dy <= radius; dy++ {
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

/* an experiment to draw the entire minimap as a paletted image
func (mapObject *Map) DrawMinimap2(screen *ebiten.Image, cities []MiniMapCity, centerX int, centerY int, zoom float64, fog [][]bool, counter uint64, crosshairs bool){
    // draw entire minimap as a paletted image
    // find a subimage that corresponds to the centerX/centerY position
    // use writePixels to copy pixels from sub image to the screen

    cursorColorBlue := math.Sin(float64(counter) / 10.0) * 127.0 + 127.0
    if cursorColorBlue > 255 {
        cursorColorBlue = 255
    }
    cursorColor := util.PremultiplyAlpha(color.RGBA{R: 255, G: 255, B: byte(cursorColorBlue), A: 180})

    landColor := color.RGBA{R: 0, G: 0xad, B: 0x00, A: 255}
    oceanColor := color.RGBA{R: 0, G: 0, B: 255, A: 255}
    riverColor := color.RGBA{R: 0x3f, G: 0x88, B: 0xd3, A: 255}
    mountainColor := color.RGBA{R: 0xbc, G: 0xd0, B: 0xe4, A: 255}
    desertColor := color.RGBA{R: 0xdb, G: 0xbd, B: 0x29, A: 255}
    tundraColor := color.RGBA{R: 0xd6, G: 0xd4, B: 0xc9, A: 255}
    volcanoColor := color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 255}
    lakeColor := color.RGBA{R: 0x3f, G: 0x88, B: 0xd3, A: 255}
    natureNodeColor := color.RGBA{R: 0, G: 255, B: 0, A: 255}
    sorceryNodeColor := color.RGBA{R: 0, G: 0, B: 255, A: 255}
    chaosNodeColor := color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 255}
    defaultColor := color.RGBA{R: 64, G: 64, B: 64, A: 255}
    blackColor := color.RGBA{R: 0, G: 0, B: 0, A: 255}

    bannerGreen := bannerColor(data.BannerGreen)
    bannerBlue := bannerColor(data.BannerBlue)
    bannerPurple := bannerColor(data.BannerPurple)
    bannerRed := bannerColor(data.BannerRed)
    bannerYellow := bannerColor(data.BannerYellow)
    bannerBrown := bannerColor(data.BannerBrown)

    mapPalette := color.Palette{
        cursorColor,
        landColor,
        oceanColor,
        riverColor,
        mountainColor,
        desertColor,
        tundraColor,
        volcanoColor,
        lakeColor,
        natureNodeColor,
        sorceryNodeColor,
        chaosNodeColor,
        defaultColor,
        blackColor,

        bannerGreen,
        bannerBlue,
        bannerPurple,
        bannerRed,
        bannerYellow,
        bannerBrown,
    }

    const cursorIndex = 0
    const grassIndex = 1
    const shoreIndex = 1
    const hillIndex = 1
    const SwampIndex = 1
    const ForestIndex = 1
    const oceanIndex = 2
    const riverIndex = 3
    const mountainIndex = 4
    const desertIndex = 5
    const tundraIndex = 6
    const volcanoIndex = 7
    const lakeIndex = 8
    const natureNodeIndex = 9
    const sorceryNodeIndex = 10
    const chaosNodeIndex = 11
    const defaultIndex = 12
    const blackIndex = 13
    const bannerGreenIndex = 14
    const bannerBlueIndex = 15
    const bannerPurpleIndex = 16
    const bannerRedIndex = 17
    const bannerYellowIndex = 18
    const bannerBrownIndex = 19

    if len(mapObject.miniMapPixels) != screen.Bounds().Dx() * screen.Bounds().Dy() * 4 {
        log.Printf("set minimap pixels to %v", screen.Bounds().Dx() * screen.Bounds().Dy() * 4)
        mapObject.miniMapPixels = make([]byte, screen.Bounds().Dx() * screen.Bounds().Dy() * 4)
        mapObject.miniMapImage = image.NewPaletted(image.Rect(0, 0, mapObject.Map.Columns(), mapObject.Map.Rows()), mapPalette)
    }

    mapObject.miniMapImage.Palette = mapPalette

    for x := 0; x < mapObject.Map.Columns(); x++ {
        for y := 0; y < mapObject.Map.Rows(); y++ {
            if !fog[x][y] {
                mapObject.miniMapImage.SetColorIndex(x, y, uint8(blackIndex))
            } else {
                var index = 0
                switch terrain.GetTile(mapObject.Map.Terrain[x][y]).TerrainType() {
                    case terrain.Grass: index = grassIndex
                    case terrain.Ocean: index = oceanIndex
                    case terrain.River: index = riverIndex
                    case terrain.Shore: index = shoreIndex
                    case terrain.Mountain: index = mountainIndex
                    case terrain.Hill: index = hillIndex
                    case terrain.Swamp: index = SwampIndex
                    case terrain.Forest: index = ForestIndex
                    case terrain.Desert: index = desertIndex
                    case terrain.Tundra: index = tundraIndex
                    case terrain.Volcano: index = volcanoIndex
                    case terrain.Lake: index = lakeIndex
                    case terrain.NatureNode: index = natureNodeIndex
                    case terrain.SorceryNode: index = sorceryNodeIndex
                    case terrain.ChaosNode: index = chaosNodeIndex
                    default: index = defaultIndex
                }

                mapObject.miniMapImage.SetColorIndex(x, y, uint8(index))
            }
        }
    }

    for _, city := range cities {
        x, y := city.GetX(), city.GetY()
        if fog[x][y] {
            index := bannerGreenIndex
            switch city.GetBanner() {
                case data.BannerBlue: index = bannerBlueIndex
                case data.BannerGreen: index = bannerGreenIndex
                case data.BannerPurple: index = bannerPurpleIndex
                case data.BannerRed: index = bannerRedIndex
                case data.BannerYellow: index = bannerYellowIndex
                case data.BannerBrown: index = bannerBrownIndex
            }
            mapObject.miniMapImage.SetColorIndex(x, y, uint8(index))
        }
    }

    pixels := mapObject.miniMapPixels
    rowSize := screen.Bounds().Dx()

    set := func(x int, y int, c color.Color){
        baseIndex := (y * rowSize + x) * 4

        / *
        if baseIndex > len(mapObject.miniMapPixels) {
            return
        }
        * /

        r, g, b, a := c.RGBA()

        pixels[baseIndex + 0] = byte(r >> 8)
        pixels[baseIndex + 1] = byte(g >> 8)
        pixels[baseIndex + 2] = byte(b >> 8)
        pixels[baseIndex + 3] = byte(a >> 8)
    }


    for x := range mapObject.Map.Columns() {
        for y := range mapObject.Map.Rows() {
            c := mapObject.miniMapImage.At(x, y)
            // set all the pixels in a block

            if data.ScreenScale == 1 {
                set(x, y, c)
            } else {
                for dx := range data.ScreenScale {
                    for dy := range data.ScreenScale {
                        cx := (mapObject.WrapX(x + centerX)) * data.ScreenScale + dx
                        cy := y + dy + centerY
                        // log.Printf("set at %v, %v %v, %v", x, y, cx, cy)
                        if cx < 0 || cx >= screen.Bounds().Dx() || cy < 0 || cy >= screen.Bounds().Dy() {
                            continue
                        }
                        set(cx, cy, c)
                    }
                }
            }
        }
    }

    screen.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 255})
    screen.WritePixels(mapObject.miniMapPixels)
}
*/

func (mapObject *Map) DrawMinimap(screen *ebiten.Image, cities []MiniMapCity, centerX int, centerY int, zoom float64, fog data.FogMap, counter uint64, crosshairs bool){
    if len(mapObject.miniMapPixels) != screen.Bounds().Dx() * screen.Bounds().Dy() * 4 {
        // log.Printf("set minimap pixels to %v", screen.Bounds().Dx() * screen.Bounds().Dy() * 4)
        mapObject.miniMapPixels = make([]byte, screen.Bounds().Dx() * screen.Bounds().Dy() * 4)
    }

    rowSize := screen.Bounds().Dx()

    cameraX := scale.Scale(centerX) - screen.Bounds().Dx() / 2
    cameraY := scale.Scale(centerY) - screen.Bounds().Dy() / 2

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

    if cameraY > scale.Scale(mapObject.Map.Rows()) - screen.Bounds().Dy() {
        cameraY = scale.Scale(mapObject.Map.Rows()) - screen.Bounds().Dy()
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
        // FIXME: how should this behave for different fog types?
        if fog[city.GetX()][city.GetY()] != data.FogTypeUnexplored {
            cityLocations[image.Pt(city.GetX(), city.GetY())] = bannerColor(city.GetBanner())
        }
    }

    getMapColor := functional.Memoize2(func (terrainType terrain.TerrainType, explored data.FogType) color.RGBA {
        var use color.RGBA

        landColor := color.RGBA{R: 0, G: 0xad, B: 0x00, A: 255}

        switch terrainType {
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

        if explored == data.FogTypeExplored {
            use = util.ToRGBA(util.Lighten(use, -50))
        }

        return use
    })

    getCityColor := functional.Memoize2(func (cityColor color.RGBA, explored data.FogType) color.RGBA {
        use := cityColor
        if explored == data.FogTypeExplored {
            use = util.ToRGBA(util.Lighten(use, -50))
        }
        return use
    })

    for x := 0; x < screen.Bounds().Dx(); x++ {
        for y := 0; y < screen.Bounds().Dy(); y++ {
            tileX := mapObject.WrapX(scale.Unscale(x + cameraX))
            tileY := scale.Unscale(y + cameraY)

            if tileX < 0 || tileX >= mapObject.Map.Columns() || tileY < 0 || tileY >= mapObject.Map.Rows() || fog[tileX][tileY] == data.FogTypeUnexplored {
                set(x, y, black)
                continue
            }

            use := getMapColor(terrain.GetTile(mapObject.Map.Terrain[tileX][tileY]).TerrainType(), fog[tileX][tileY])

            if cityColor, ok := cityLocations[image.Pt(tileX, tileY)]; ok {
                use = getCityColor(cityColor, fog[tileX][tileY])
            }

            set(x, y, use)
        }
    }

    if crosshairs {
        cursorColorBlue := math.Sin(float64(counter) * 3 * math.Pi / 180) * 127.0 + 127.0
        if cursorColorBlue > 255 {
            cursorColorBlue = 255
        }
        cursorColor := util.PremultiplyAlpha(color.RGBA{R: 255, G: 255, B: byte(cursorColorBlue), A: 180})

        cursorRadius := int(float64(scale.Scale(5.0)) / zoom)
        x1 := scale.Scale(centerX) - cursorRadius - cameraX
        y1 := scale.Scale(centerY) - cursorRadius - cameraY
        x2 := scale.Scale(centerX) + cursorRadius - cameraX
        y2 := y1
        x3 := x1
        y3 := scale.Scale(centerY) + cursorRadius - cameraY
        x4 := x2
        y4 := y3
        points := []image.Point{
            image.Pt(x1, y1),
            image.Pt(x1+scale.Scale(1), y1),
            image.Pt(x1, y1+scale.Scale(1)),

            image.Pt(x2, y2),
            image.Pt(x2-scale.Scale(1), y2),
            image.Pt(x2, y2+scale.Scale(1)),

            image.Pt(x3, y3),
            image.Pt(x3+scale.Scale(1), y3),
            image.Pt(x3, y3-scale.Scale(1)),

            image.Pt(x4, y4),
            image.Pt(x4-scale.Scale(1), y4),
            image.Pt(x4, y4-scale.Scale(1)),
        }

        drawSquare := func(x int, y int, c color.RGBA){
            if scale.ScaleAmount == 1 {
                set(x, y, cursorColor)
            } else {
                for dx := range int(scale.ScaleAmount) {
                    for dy := range int(scale.ScaleAmount) {
                        cx := x + dx
                        cy := y + dy

                        if cx >= 0 && cy >= 0 && cx < screen.Bounds().Dx() && cy < screen.Bounds().Dy() {
                            set(cx, cy, cursorColor)
                        }
                    }
                }
            }
        }

        for _, point := range points {
            x := scale.Scale(mapObject.WrapX(scale.Unscale(point.X)))
            y := point.Y

            if x >= 0 && y >= 0 && x < screen.Bounds().Dx() && y < screen.Bounds().Dy(){
                drawSquare(x, y, cursorColor)
            }
        }
    }

    screen.WritePixels(mapObject.miniMapPixels)
    /*
    red := ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
    red.Fill(color.RGBA{R: 255, G: 0, B: 0, A: 255})
    var options ebiten.DrawImageOptions
    screen.DrawImage(red, &options)
    */
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

                scale.DrawScaled(screen, tileImage, &options)

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
func (mapObject *Map) DrawLayer2(camera cameralib.Camera, animationCounter uint64, imageCache *util.ImageCache, screen *ebiten.Image, geom ebiten.GeoM){
    tileWidth := mapObject.TileWidth()
    tileHeight := mapObject.TileHeight()

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

            if len(mapObject.ExtraMap[image.Pt(tileX, tileY)]) == 0 {
                continue
            }

            options.GeoM.Reset()
            options.GeoM.Translate(float64(x * tileWidth), float64(y * tileHeight))
            options.GeoM.Concat(geom)

            for _, extraKind := range ExtraDrawOrder {
                extra, ok := mapObject.ExtraMap[image.Pt(tileX, tileY)][extraKind]
                if ok {
                    extra.DrawLayer2(screen, imageCache, &options, animationCounter, tileWidth, tileHeight)
                }
            }

        }
    }
}
