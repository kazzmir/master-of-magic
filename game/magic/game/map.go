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
    Draw(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int)
}

type ExtraRoad struct {
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

func (node *ExtraMagicNode) Draw(screen *ebiten.Image, imageCache *util.ImageCache, options *ebiten.DrawImageOptions, counter uint64, tileWidth int, tileHeight int){
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
        if node.GuardianSpiritMeld && rand.Intn(4) != 0 {
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

/* choose X points surrounding the node. 0,0 is the node itself. for arcanus, choose 5-10 points from a 4x4 square.
 * for myrror choose 10-20 points from a 5x5 square.
 */
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

/* budget for making encounter monsters is zone size + bonus
 */
func computeEncounterBudget(magicSetting data.MagicSetting, difficultySetting data.DifficultySetting, zoneSize int) int {
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

    return budget + int(float64(budget) * bonus)
}

/* divide budget by some divisor in range 1 to numChoices. find the enemy with the largest cost
 * that fits in the divided result.
 */
func chooseEnemy[E comparable](enemyCosts map[E]int, budget int, numChoices int) E {
    if numChoices < 0 {
        numChoices = 0
    }

    choices := rand.Perm(numChoices)
    var zero E

    for _, choice := range choices {
        divisor := choice + 1

        var enemyChoice E
        maxCost := 0

        for unit, cost := range enemyCosts {
            if cost > maxCost && cost <= budget / divisor {
                enemyChoice = unit
                maxCost = cost
            }
        }

        if enemyChoice != zero {
            return enemyChoice
        }
    }

    return zero
}

func chooseGuardianAndSecondary[E comparable](enemyCosts map[E]int, makeUnit func(E) units.Unit, budget int) ([]units.Unit, []units.Unit) {
    var guardians []units.Unit
    var secondary []units.Unit

    enemyChoice := chooseEnemy(enemyCosts, budget, 4)

    var zero E

    // chose no enemies!
    if enemyChoice == zero {
        return nil, nil
    }

    numGuardians := budget / enemyCosts[enemyChoice]
    if numGuardians > 9 {
        numGuardians = 9
    }

    for i := 0; i < numGuardians; i++ {
        guardians = append(guardians, makeUnit(enemyChoice))
    }

    remainingBudget := budget - numGuardians * enemyCosts[enemyChoice]

    enemyChoice = chooseEnemy(enemyCosts, remainingBudget, 10 - numGuardians)

    if enemyChoice != zero {
        secondaryCount := remainingBudget / enemyCosts[enemyChoice]

        if secondaryCount > 9 - numGuardians {
            secondaryCount = 9 - numGuardians
        }

        for i := 0; i < secondaryCount; i++ {
            secondary = append(secondary, makeUnit(enemyChoice))
        }
    }

    return guardians, secondary
}

/* returns guardian units and secondary units
 */
func computeNatureNodeEnemies(magicSetting data.MagicSetting, difficultySetting data.DifficultySetting, zoneSize int) ([]units.Unit, []units.Unit) {
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

    return chooseGuardianAndSecondary(enemyCosts, makeUnit, computeEncounterBudget(magicSetting, difficultySetting, zoneSize))
}

func computeSorceryNodeEnemies(magicSetting data.MagicSetting, difficultySetting data.DifficultySetting, zoneSize int) ([]units.Unit, []units.Unit) {
    type Enemy int
    const (
        None Enemy = iota
        PhantomWarriors
        Naga
        AirElemental
        PhantomBeast
        StormGiant
        Djinn
        SkyDrake
    )

    makeUnit := func(enemy Enemy) units.Unit {
        switch enemy {
            case PhantomWarriors: return units.PhantomWarrior
            case Naga: return units.Nagas
            case AirElemental: return units.AirElemental
            case PhantomBeast: return units.PhantomBeast
            case StormGiant: return units.StormGiant
            case Djinn: return units.Djinn
            case SkyDrake: return units.SkyDrake
        }

        return units.UnitNone
    }

    enemyCosts := map[Enemy]int{
        None: 0,
        PhantomWarriors: 20,
        Naga: 120,
        AirElemental: 170,
        PhantomBeast: 225,
        StormGiant: 500,
        Djinn: 650,
        SkyDrake: 1000,
    }

    return chooseGuardianAndSecondary(enemyCosts, makeUnit, computeEncounterBudget(magicSetting, difficultySetting, zoneSize))
}

func computeChaosNodeEnemies(magicSetting data.MagicSetting, difficultySetting data.DifficultySetting, zoneSize int) ([]units.Unit, []units.Unit) {
    type Enemy int
    const (
        None Enemy = iota
        HellHounds
        FireElemental
        FireGiant
        Gargoyles
        DoomBat
        Chimera
        ChaosSpawn
        Efreet
        Hydra
        GreatDrake
    )

    makeUnit := func(enemy Enemy) units.Unit {
        switch enemy {
            case HellHounds: return units.HellHounds
            case FireElemental: return units.FireElemental
            case FireGiant: return units.FireGiant
            case Gargoyles: return units.Gargoyle
            case DoomBat: return units.DoomBat
            case Chimera: return units.Chimeras
            case ChaosSpawn: return units.ChaosSpawn
            case Efreet: return units.Efreet
            case Hydra: return units.Hydra
            case GreatDrake: return units.GreatDrake
        }

        return units.UnitNone
    }

    enemyCosts := map[Enemy]int{
        None: 0,
        HellHounds: 40,
        FireElemental: 100,
        FireGiant: 150,
        Gargoyles: 200,
        DoomBat: 300,
        Chimera: 350,
        ChaosSpawn: 400,
        Efreet: 550,
        Hydra: 650,
        GreatDrake: 900,
    }

    return chooseGuardianAndSecondary(enemyCosts, makeUnit, computeEncounterBudget(magicSetting, difficultySetting, zoneSize))
}

func MakeMagicNode(kind MagicNode, magicSetting data.MagicSetting, difficulty data.DifficultySetting, plane data.Plane) *ExtraMagicNode {
    zone := makeZone(plane)
    var guardians []units.Unit
    var secondary []units.Unit

    switch kind {
        case MagicNodeNature:
            guardians, secondary = computeNatureNodeEnemies(magicSetting, difficulty, len(zone))
            log.Printf("Created nature node guardians: %v secondary: %v", guardians, secondary)
        case MagicNodeSorcery:
            guardians, secondary = computeSorceryNodeEnemies(magicSetting, difficulty, len(zone))
            log.Printf("Created sorcery node guardians: %v secondary: %v", guardians, secondary)
        case MagicNodeChaos:
            guardians, secondary = computeChaosNodeEnemies(magicSetting, difficulty, len(zone))
            log.Printf("Created chaos node guardians: %v secondary: %v", guardians, secondary)
    }


    return &ExtraMagicNode{
        Kind: kind,
        Empty: false,
        Guardians: guardians,
        Secondary: secondary,
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
                    extra.Draw(screen, imageCache, &options, animationCounter, tileWidth, tileHeight)
                }
            } else {
                log.Printf("Unable to render tilte at %d, %d: %v", tileX, tileY, err)
            }
        }
    }
}
