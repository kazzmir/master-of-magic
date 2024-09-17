package game

import (
    "log"
    "math"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"

    "github.com/hajimehoshi/ebiten/v2"
)

type Map struct {
    Map *terrain.Map

    Data *terrain.TerrainData

    TileCache map[int]*ebiten.Image

    miniMapPixels []byte
}

func MakeMap(data *terrain.TerrainData) *Map {
    return &Map{
        Data: data,
        Map: terrain.GenerateLandCellularAutomata(100, 200, data),
        TileCache: make(map[int]*ebiten.Image),
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

func (mapObject *Map) DrawMinimap(screen *ebiten.Image, cities []*citylib.City, cameraX int, cameraY int, fog [][]bool, counter uint64){
    if len(mapObject.miniMapPixels) != screen.Bounds().Dx() * screen.Bounds().Dy() * 4 {
        mapObject.miniMapPixels = make([]byte, screen.Bounds().Dx() * screen.Bounds().Dy() * 4)
    }

    rowSize := screen.Bounds().Dx()

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

            tileX := x + cameraX - screen.Bounds().Dx() / 2
            tileY := y + cameraY - screen.Bounds().Dy() / 2

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
            posX := city.X - cameraX + screen.Bounds().Dx() / 2
            posY := city.Y - cameraY + screen.Bounds().Dy() / 2

            if posX >= 0 && posX < screen.Bounds().Dx() && posY >= 0 && posY < screen.Bounds().Dy() {
                set(posX, posY, color.RGBA{R: 255, G: 255, B: 255, A: 255})
            }
        }
    }

    cursorColorBlue := math.Sin(float64(counter) / 10.0) * 127.0 + 127.0
    if cursorColorBlue > 255 {
        cursorColorBlue = 255
    }
    cursorColor := util.PremultiplyAlpha(color.RGBA{R: 255, G: 255, B: byte(cursorColorBlue), A: 180})

    cursorRadius := 5
    x1 := screen.Bounds().Dx() / 2 - cursorRadius
    y1 := screen.Bounds().Dy() / 2 - cursorRadius
    x2 := screen.Bounds().Dx() / 2 + cursorRadius
    y2 := y1
    x3 := x1
    y3 := screen.Bounds().Dy() / 2 + cursorRadius
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

    screen.WritePixels(mapObject.miniMapPixels)
}

func (mapObject *Map) Draw(cameraX int, cameraY int, animationCounter uint64, screen *ebiten.Image, geom ebiten.GeoM){

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

            image, err := mapObject.GetTileImage(tileX, tileY, animationCounter)
            if err == nil {
                options.GeoM = geom
                // options.GeoM.Reset()
                options.GeoM.Translate(float64(x * tileWidth), float64(y * tileHeight))
                screen.DrawImage(image, &options)
            } else {
                log.Printf("Unable to render tilte at %d, %d: %v", tileX, tileY, err)
            }
        }
    }
}
