package game

import (
    "log"
    "math"
    "math/rand/v2"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/units"

    "github.com/hajimehoshi/ebiten/v2"
)

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

type BonusType int
const (
    BonusNone BonusType = iota
    BonusWildGame
    BonusNightshade
    BonusSilverOre
    BonusGoldOre
    BonusIronOre
    BonusCoal
    BonusMithrilOre
    BonusAdamantiumOre
    BonusGem
    BonusQuorkCrystal
    BonusCrysxCrystal
)

// wild game, gold ore, mithril, etc
type ExtraBonus struct {
    Bonus BonusType
}

func (bonus BonusType) FoodBonus() int {
    if bonus == BonusWildGame {
        return 2
    }

    return 0
}

func (bonus BonusType) GoldBonus() int {
    switch bonus {
        case BonusSilverOre: return 2
        case BonusGoldOre: return 3
        case BonusGem: return 5
        default: return 0
    }
}

func (bonus BonusType) PowerBonus() int {
    switch bonus {
        case BonusMithrilOre: return 1
        case BonusAdamantiumOre: return 2
        case BonusQuorkCrystal: return 3
        case BonusCrysxCrystal: return 5
        default: return 0
    }
}

// returns a percent that unit costs are reduced by, 10 -> -10%
func (bonus BonusType) UnitReductionBonus() int {
    switch bonus {
        case BonusIronOre: return 5
        case BonusCoal: return 10
        default: return 0
    }
}

func (bonus *ExtraBonus) DrawLayer1(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int){
    index := -1

    switch bonus.Bonus {
        case BonusWildGame: index = 92
        case BonusNightshade: index = 91
        case BonusSilverOre: index = 80
        case BonusGoldOre: index = 81
        case BonusIronOre: index = 78
        case BonusCoal: index = 79
        case BonusMithrilOre: index = 83
        case BonusAdamantiumOre: index = 84
        case BonusGem: index = 82
        case BonusQuorkCrystal: index = 85
        case BonusCrysxCrystal: index = 86
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

type ExtraMagicNode struct {
    Kind MagicNode
    Empty bool
    Guardians []units.Unit
    Secondary []units.Unit
    // list of points that are affected by the node
    Zone []image.Point

    // if this node is melded, then this player receives the power
    MeldingWizard *playerlib.Player
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
        switch node.MeldingWizard.Wizard.Banner {
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

func (node *ExtraMagicNode) Meld(meldingWizard *playerlib.Player, spirit units.Unit) bool {
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

type Map struct {
    Map *terrain.Map

    // contains information about map squares that contain extra features on top
    // such as a road, enchantment, encounter place (plane tower, lair, etc)
    ExtraMap map[image.Point]ExtraTile

    Data *terrain.TerrainData

    TileCache map[int]*ebiten.Image

    miniMapPixels []byte
}

func MakeMap(data *terrain.TerrainData) *Map {
    return &Map{
        Data: data,
        Map: terrain.GenerateLandCellularAutomata(100, 200, data),
        TileCache: make(map[int]*ebiten.Image),
        ExtraMap: make(map[image.Point]ExtraTile),
    }
}

func (mapObject *Map) GetMeldedNodes(player *playerlib.Player) []*ExtraMagicNode {
    var out []*ExtraMagicNode

    for _, extra := range mapObject.ExtraMap {
        if node, ok := extra.(*ExtraMagicNode); ok {
            if node.MeldingWizard == player {
                out = append(out, node)
            }
        }
    }

    return out
}

func (mapObject *Map) SetBonus(x int, y int, bonus BonusType) {
    mapObject.ExtraMap[image.Pt(x, y)] = &ExtraBonus{Bonus: bonus}
}

func (mapObject *Map) GetBonusTile(x int, y int) BonusType {
    if extra, ok := mapObject.ExtraMap[image.Pt(x, y)]; ok {
        if bonus, ok := extra.(*ExtraBonus); ok {
            return bonus.Bonus
        }
    }

    return BonusNone
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

func (mapObject *Map) GetTile(tileX int, tileY int) terrain.Tile {
    if tileX >= 0 && tileX < mapObject.Map.Columns() && tileY >= 0 && tileY < mapObject.Map.Rows() {
        return mapObject.Data.Tiles[mapObject.Map.Terrain[tileX][tileY]].Tile
    }

    return terrain.Tile{Index: -1}
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

func (mapObject *Map) DrawMinimap(screen *ebiten.Image, cities []*citylib.City, centerX int, centerY int, fog [][]bool, counter uint64, crosshairs bool){
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
        if fog[city.X][city.Y] {
            posX := city.X - cameraX
            posY := city.Y - cameraY

            if posX >= 0 && posX < screen.Bounds().Dx() && posY >= 0 && posY < screen.Bounds().Dy() {
                set(posX, posY, bannerColor(city.Banner))
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
