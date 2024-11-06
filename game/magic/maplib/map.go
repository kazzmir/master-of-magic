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

    "github.com/hajimehoshi/ebiten/v2"
)

type MiniMapCity interface {
    GetX() int
    GetY() int
    GetBanner() data.BannerType
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

type ExtraRoad struct {
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
}

// choices is a map from a name to the chance of choosing that name, where all the int values should add up to 100
// an individual int is a percentage chance to choose the given key
// for example, choices might be the map {"a": 30, "b": 30, "c": 40}
// which means that a and b should both have a 30% chance of being picked, and c has a 40% chance of being picked
func chooseValue(choices map[string]int) string {
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

    return ""
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
        case EncounterTypeLair: index = 71
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
    Extra ExtraTile
    Tile terrain.Tile
    X int
    Y int
}

func (tile *FullTile) Valid() bool {
    return tile.Tile.Index != -1
}

func (tile *FullTile) GetBonus() data.BonusType {
    if tile.Extra == nil {
        return data.BonusNone
    }

    if bonus, ok := tile.Extra.(*ExtraBonus); ok {
        return bonus.Bonus
    }

    return data.BonusNone
}

func (tile *FullTile) HasWildGame() bool {
    if tile.Extra == nil {
        return false
    }

    if bonus, ok := tile.Extra.(*ExtraBonus); ok {
        return bonus.Bonus == data.BonusWildGame
    }

    return false
}

type Map struct {
    Map *terrain.Map

    // contains information about map squares that contain extra features on top
    // such as a road, enchantment, encounter place (plane tower, lair, etc)
    ExtraMap map[image.Point]ExtraTile

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

func MakeMap(data *terrain.TerrainData, landSize int) *Map {
    landWidth, landHeight := getLandSize(landSize)

    return &Map{
        Data: data,
        Map: terrain.GenerateLandCellularAutomata(landWidth, landHeight, data),
        TileCache: make(map[int]*ebiten.Image),
        ExtraMap: make(map[image.Point]ExtraTile),
    }
}

func (mapObject *Map) GetMeldedNodes(melder Melder) []*ExtraMagicNode {
    var out []*ExtraMagicNode

    for _, extra := range mapObject.ExtraMap {
        if node, ok := extra.(*ExtraMagicNode); ok {
            if node.MeldingWizard == melder {
                out = append(out, node)
            }
        }
    }

    return out
}

func (mapObject *Map) SetBonus(x int, y int, bonus data.BonusType) {
    mapObject.ExtraMap[image.Pt(x, y)] = &ExtraBonus{Bonus: bonus}
}

func (mapObject *Map) GetBonusTile(x int, y int) data.BonusType {
    if extra, ok := mapObject.ExtraMap[image.Pt(x, y)]; ok {
        if bonus, ok := extra.(*ExtraBonus); ok {
            return bonus.Bonus
        }
    }

    return data.BonusNone
}

func (mapObject *Map) CreateEncounter(x int, y int, encounterType EncounterType, difficulty data.DifficultySetting, weakStrength bool, plane data.Plane) bool {
    _, ok := mapObject.ExtraMap[image.Pt(x, y)]
    if ok {
        return false
    }

    mapObject.ExtraMap[image.Pt(x, y)] = makeEncounter(encounterType, difficulty, weakStrength, plane)
    return true
}

func (mapObject *Map) CreateEncounterRandom(x int, y int, difficulty data.DifficultySetting, plane data.Plane) bool {
    return mapObject.CreateEncounter(x, y, randomEncounterType(), difficulty, rand.N(2) == 0, plane)
}

func (mapObject *Map) CreateNode(x int, y int, node MagicNode, plane data.Plane, magicSetting data.MagicSetting, difficulty data.DifficultySetting) *ExtraMagicNode {
    tileType := 0
    switch node {
        case MagicNodeNature: tileType = terrain.TileNatureForest.Index
        case MagicNodeSorcery: tileType = terrain.TileSorceryLake.Index
        case MagicNodeChaos: tileType = terrain.TileChaosVolcano.Index
    }

    mapObject.Map.Terrain[x][y] = tileType

    out := MakeMagicNode(node, magicSetting, difficulty, plane)

    mapObject.ExtraMap[image.Pt(x, y)] = out

    return out
}

func (mapObject *Map) GetLair(x int, y int) *ExtraEncounter {
    if extra, ok := mapObject.ExtraMap[image.Pt(x, y)]; ok {
        if encounter, ok := extra.(*ExtraEncounter); ok {
            return encounter
        }
    }

    return nil
}

func (mapObject *Map) GetMagicNode(x int, y int) *ExtraMagicNode {
    if extra, ok := mapObject.ExtraMap[image.Pt(x, y)]; ok {
        if node, ok := extra.(*ExtraMagicNode); ok {
            return node
        }
    }

    return nil
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

func (mapObject *Map) GetTile(tileX int, tileY int) FullTile {
    if tileX >= 0 && tileX < mapObject.Map.Columns() && tileY >= 0 && tileY < mapObject.Map.Rows() {
        tile := mapObject.Data.Tiles[mapObject.Map.Terrain[tileX][tileY]].Tile

        extra, ok := mapObject.ExtraMap[image.Pt(tileX, tileY)]
        if !ok {
            extra = nil
        }

        return FullTile{
            Tile: tile,
            X: tileX,
            Y: tileY,
            Extra: extra,
        }
    }

    return FullTile{
        Tile: terrain.Tile{Index: -1},
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

            tileX := x + dx
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

func (mapObject *Map) DrawMinimap(screen *ebiten.Image, cities []MiniMapCity, centerX int, centerY int, fog [][]bool, counter uint64, crosshairs bool){
    if len(mapObject.miniMapPixels) != screen.Bounds().Dx() * screen.Bounds().Dy() * 4 {
        mapObject.miniMapPixels = make([]byte, screen.Bounds().Dx() * screen.Bounds().Dy() * 4)
    }

    rowSize := screen.Bounds().Dx()

    cameraX := centerX - screen.Bounds().Dx() / 2
    cameraY := centerY - screen.Bounds().Dy() / 2

    if cameraX < 0 {
        cameraX = 0
    }
    if cameraY < 0 {
        cameraY = 0
    }

    if cameraX > mapObject.Map.Columns() - screen.Bounds().Dx() {
        cameraX = mapObject.Map.Columns() - screen.Bounds().Dx()
    }
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

    for x := 0; x < screen.Bounds().Dx(); x++ {
        for y := 0; y < screen.Bounds().Dy(); y++ {

            tileX := x + cameraX
            tileY := y + cameraY

            if tileX < 0 || tileX >= mapObject.Map.Columns() || tileY < 0 || tileY >= mapObject.Map.Rows() || !fog[tileX][tileY] {
                set(x, y, black)
                continue
            }

            var use color.RGBA

            switch mapObject.Map.Terrain[tileX][tileY] {
                case terrain.TileLand.Index: use = color.RGBA{R: 0, G: 255, B: 0, A: 255}
                case terrain.TileOcean.Index: use = color.RGBA{R: 0, G: 0, B: 255, A: 255}
                default: use = color.RGBA{R: 64, G: 64, B: 64, A: 255}
            }

            set(x, y, use)
        }
    }

    for _, city := range cities {
        if fog[city.GetX()][city.GetY()] {
            posX := city.GetX() - cameraX
            posY := city.GetY() - cameraY

            if posX >= 0 && posX < screen.Bounds().Dx() && posY >= 0 && posY < screen.Bounds().Dy() {
                set(posX, posY, bannerColor(city.GetBanner()))
            }
        }
    }

    if crosshairs {
        cursorColorBlue := math.Sin(float64(counter) / 10.0) * 127.0 + 127.0
        if cursorColorBlue > 255 {
            cursorColorBlue = 255
        }
        cursorColor := util.PremultiplyAlpha(color.RGBA{R: 255, G: 255, B: byte(cursorColorBlue), A: 180})

        cursorRadius := 5
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
            if point.X >= 0 && point.Y >= 0 && point.X < screen.Bounds().Dx() && point.Y < screen.Bounds().Dy(){
                set(point.X, point.Y, cursorColor)
            }
        }
    }

    screen.WritePixels(mapObject.miniMapPixels)
}

// draw base map tiles, in general stuff that should go under cities/units
func (mapObject *Map) DrawLayer1(cameraX int, cameraY int, animationCounter uint64, imageCache *util.ImageCache, screen *ebiten.Image, geom ebiten.GeoM){
    tileWidth := mapObject.TileWidth()
    tileHeight := mapObject.TileHeight()

    tilesPerRow := mapObject.TilesPerRow(screen.Bounds().Dx())
    tilesPerColumn := mapObject.TilesPerColumn(screen.Bounds().Dy())

    var options ebiten.DrawImageOptions

    // draw all tiles first
    for x := 0; x < tilesPerRow; x++ {
        for y := 0; y < tilesPerColumn; y++ {

            tileX := cameraX + x
            tileY := cameraY + y

            if tileX < 0 || tileX >= mapObject.Map.Columns() || tileY < 0 || tileY >= mapObject.Map.Rows() {
                continue
            }

            tileImage, err := mapObject.GetTileImage(tileX, tileY, animationCounter)
            if err == nil {
                options.GeoM = geom
                // options.GeoM.Reset()
                options.GeoM.Translate(float64(x * tileWidth), float64(y * tileHeight))
                screen.DrawImage(tileImage, &options)

                extra, ok := mapObject.ExtraMap[image.Pt(tileX, tileY)]
                if ok {
                    extra.DrawLayer1(screen, imageCache, &options, animationCounter, tileWidth, tileHeight)
                }
            } else {
                log.Printf("Unable to render tilte at %d, %d: %v", tileX, tileY, err)
            }
        }
    }
}

// give the extra nodes a chance to draw on top of cities/units, but still under the fog
func (mapObject *Map) DrawLayer2(cameraX int, cameraY int, animationCounter uint64, imageCache *util.ImageCache, screen *ebiten.Image, geom ebiten.GeoM){
    tileWidth := mapObject.TileWidth()
    tileHeight := mapObject.TileHeight()

    tilesPerRow := mapObject.TilesPerRow(screen.Bounds().Dx())
    tilesPerColumn := mapObject.TilesPerColumn(screen.Bounds().Dy())

    var options ebiten.DrawImageOptions

    // then draw all extra nodes on top
    for x := 0; x < tilesPerRow; x++ {
        for y := 0; y < tilesPerColumn; y++ {

            tileX := cameraX + x
            tileY := cameraY + y

            if tileX < 0 || tileX >= mapObject.Map.Columns() || tileY < 0 || tileY >= mapObject.Map.Rows() {
                continue
            }

            extra, ok := mapObject.ExtraMap[image.Pt(tileX, tileY)]
            if ok {
                options.GeoM = geom
                options.GeoM.Translate(float64(x * tileWidth), float64(y * tileHeight))
                extra.DrawLayer2(screen, imageCache, &options, animationCounter, tileWidth, tileHeight)
            }
        }
    }
}
