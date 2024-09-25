package game

import (
    "log"
    "math"
    "math/rand"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
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
    Draw(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions)
}

type ExtraRoad struct {
}

type ExtraMagicNode struct {
    Kind MagicNode
    Empty bool
    Units []units.Unit
    // list of points that are affected by the node
    Zone []image.Point

    // also contains treasure
}

func (node *ExtraMagicNode) Draw(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions){
}

func makeZone(plane data.Plane) []image.Point {
    // choose X points
    maxSize := 4
    numPoints := 0
    if plane == data.PlaneArcanus {
        maxSize = 4
        numPoints = 5 + rand.Intn(5)
    } else if plane == data.PlaneMyrror {
        maxSize = 5
        numPoints = 10 + rand.Intn(10)
    }

    chosen := make(map[image.Point]bool)
    out := make([]image.Point, 0, numPoints)

    // always choose the center, which is where the node itself is
    chosen[image.Pt(0, 0)] = true
    out = append(out, image.Pt(0, 0))

    possible := make([]image.Point, 0, maxSize * maxSize)
    for x := -maxSize / 2; x < maxSize / 2; x++ {
        for y := -maxSize / 2; y < maxSize / 2; y++ {
            if x == 0 && y == 0 {
                continue
            }
            possible = append(possible, image.Pt(x, y))
        }
    }

    // choose N points from the possible points
    choices := rand.Perm(len(possible))[:numPoints]

    for _, choice := range choices {
        out = append(out, possible[choice])
    }

    return out
}

func computeNatureNodeEnemies(magicSetting data.MagicSetting, difficultySetting data.DifficultySetting, zoneSize int) []units.Unit {
    budget := 0

    // these formulas come from the master of magic wiki
    switch magicSetting {
        case data.MagicSettingWeak:
            budget = (rand.Intn(11) + 4) * (zoneSize * zoneSize) / 2
        case data.MagicSettingNormal:
            budget = (rand.Intn(11) + 4) * (zoneSize * zoneSize)
        case data.MagicSettingPowerful:
            budget = (rand.Intn(11) + 4) * (zoneSize * zoneSize) * 3 / 2
    }

    bonus := float64(0)
    switch difficultySetting {
        case data.DifficultyIntro: bonus = -0.75
        case data.DifficultyEasy: bonus = -0.5
        case data.DifficultyAverage: bonus = -0.25
        case data.DifficultyHard: bonus = 0
        case data.DifficultyExtreme: bonus = 0.25
        case data.DifficultyImpossible: bonus = 0.50
    }

    budget = budget + int(float64(budget) * bonus)

    type Enemy int
    const (
        None Enemy = iota
        WarBear
        Sprite
        EarthElemental
        Spiders
        Cockatrice
        Basilisk
        StoneGiant
        Gorgons
        Behemoth
        Colossus
        GreatWyrm
    )

    makeUnit := func(enemy Enemy) units.Unit {
        switch enemy {
            case WarBear: return units.WarBear
            case Sprite: return units.Sprite
            case EarthElemental: return units.EarthElemental
            case Spiders: return units.GiantSpider
            case Cockatrice: return units.Cockatrice
            case Basilisk: return units.Basilisk
            case StoneGiant: return units.StoneGiant
            case Gorgons: return units.Gorgon
            case Behemoth: return units.Behemoth
            case Colossus: return units.Colossus
            case GreatWyrm: return units.GreatWyrm
        }

        return units.UnitNone
    }

    enemyCosts := map[Enemy]int{
        None: 0,
        WarBear: 70,
        Sprite: 100,
        EarthElemental: 160,
        Spiders: 200,
        Cockatrice: 275,
        Basilisk: 325,
        StoneGiant: 450,
        Gorgons: 600,
        Behemoth: 700,
        Colossus: 800,
        GreatWyrm: 1000,
    }

    choices := rand.Perm(4)

    enemyChoice := None
    maxCost := 0

    // divide the budget by the divisor, then choose the most expensive unit that fits
    for _, choice := range choices {
        divisor := choice + 1

        enemyChoice = None
        maxCost = 0

        for unit, cost := range enemyCosts {
            if cost > maxCost && cost <= budget / divisor {
                enemyChoice = unit
                maxCost = cost
            }
        }

        if enemyChoice != None {
            break
        }
    }

    // chose no enemies!
    if enemyChoice == None {
        return nil
    }

    numGuardians := budget / enemyCosts[enemyChoice]

    var out []units.Unit

    for i := 0; i < numGuardians; i++ {
        out = append(out, makeUnit(enemyChoice))
    }

    remainingBudget := budget - numGuardians * enemyCosts[enemyChoice]

    enemyChoice = None
    maxCost = 0

    // divide the budget by the divisor, then choose the most expensive unit that fits
    for _, choice := range choices {
        divisor := choice + 1

        enemyChoice = None
        maxCost = 0

        for unit, cost := range enemyCosts {
            if cost > maxCost && cost <= remainingBudget / divisor {
                enemyChoice = unit
                maxCost = cost
            }
        }

        if enemyChoice != None {
            break
        }
    }

    if enemyChoice != None {
        secondary := remainingBudget / enemyCosts[enemyChoice]
        for i := 0; i < secondary; i++ {
            out = append(out, makeUnit(enemyChoice))
        }
    }

    return out
}

func MakeMagicNode(kind MagicNode, magicSetting data.MagicSetting, difficulty data.DifficultySetting, plane data.Plane) *ExtraMagicNode {
    zone := makeZone(plane)
    var enemies []units.Unit

    switch kind {
        case MagicNodeNature:
            enemies = computeNatureNodeEnemies(magicSetting, difficulty, len(zone))
        case MagicNodeSorcery:
        case MagicNodeChaos:
    }

    log.Printf("Created nature node enemies: %v", enemies)

    return &ExtraMagicNode{
        Kind: kind,
        Empty: false,
        Units: enemies,
        Zone: zone,
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

func (mapObject *Map) CreateNode(x int, y int, node MagicNode, plane data.Plane, magicSetting data.MagicSetting, difficulty data.DifficultySetting) {
    tileType := 0
    switch node {
        case MagicNodeNature: tileType = terrain.TileNatureForest.Index
        case MagicNodeSorcery: tileType = terrain.TileSorceryLake.Index
        case MagicNodeChaos: tileType = terrain.TileChaosVolcano.Index
    }

    mapObject.Map.Terrain[x][y] = tileType

    mapObject.ExtraMap[image.Pt(x, y)] = MakeMagicNode(node, magicSetting, difficulty, plane)
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
                set(posX, posY, color.RGBA{R: 255, G: 255, B: 255, A: 255})
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

func (mapObject *Map) Draw(cameraX int, cameraY int, animationCounter uint64, imageCache *util.ImageCache, screen *ebiten.Image, geom ebiten.GeoM){
    tileWidth := mapObject.TileWidth()
    tileHeight := mapObject.TileHeight()

    tilesPerRow := mapObject.TilesPerRow(screen.Bounds().Dx())
    tilesPerColumn := mapObject.TilesPerColumn(screen.Bounds().Dy())

    var options ebiten.DrawImageOptions

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
                    extra.Draw(screen, imageCache, &options)
                }
            } else {
                log.Printf("Unable to render tilte at %d, %d: %v", tileX, tileY, err)
            }
        }
    }
}
