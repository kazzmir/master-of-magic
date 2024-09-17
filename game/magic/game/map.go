package game

import (
    "log"
    "math"
    "image/color"

    "github.com/kazzmir/master-of-magic/game/magic/terrain"

    "github.com/hajimehoshi/ebiten/v2"
)

type Map struct {
    Map *terrain.Map

    Data *terrain.TerrainData

    TileCache map[int]*ebiten.Image
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

    if image, ok := mapObject.TileCache[tile * 100 + int(animationIndex)]; ok {
        return image, nil
    }


    gpuImage := ebiten.NewImageFromImage(tileInfo.Images[animationCounter % uint64(len(tileInfo.Images))])

    mapObject.TileCache[tile * 100 + int(animationIndex)] = gpuImage
    return gpuImage, nil
}

func (mapObject *Map) TilesPerRow(screenWidth int) int {
    return int(math.Ceil(float64(screenWidth) / float64(mapObject.TileWidth())))
}

func (mapObject *Map) TilesPerColumn(screenHeight int) int {
    return int(math.Ceil(float64(screenHeight) / float64(mapObject.TileHeight())))
}

func (mapObject *Map) DrawMinimap(screen *ebiten.Image, geom ebiten.GeoM){
    pixels := make([]byte, mapObject.Map.Columns() * mapObject.Map.Rows() * 4)
    for x := 0; x < mapObject.Map.Columns(); x++ {
        for y := 0; y < mapObject.Map.Rows(); y++ {

            var use color.RGBA

            switch mapObject.Map.Terrain[x][y] {
                case terrain.TileLand.Index: use = color.RGBA{R: 0, G: 255, B: 0, A: 255}
                case terrain.TileOcean.Index: use = color.RGBA{R: 0, G: 0, B: 255, A: 255}
                default: use = color.RGBA{R: 64, G: 64, B: 64, A: 255}
            }

            r, g, b, a := use.RGBA()

            /*
            pixels[(y * mapObject.Map.Columns() + x) * 4 + 0] = byte(r/255)
            pixels[(y * mapObject.Map.Columns() + x) * 4 + 1] = byte(g/255)
            pixels[(y * mapObject.Map.Columns() + x) * 4 + 2] = byte(b/255)
            pixels[(y * mapObject.Map.Columns() + x) * 4 + 3] = byte(a/255)
            */

            pixels[(y * mapObject.Map.Columns() + x) * 4 + 0] = byte(r >> 8)
            pixels[(y * mapObject.Map.Columns() + x) * 4 + 1] = byte(g >> 8)
            pixels[(y * mapObject.Map.Columns() + x) * 4 + 2] = byte(b >> 8)
            pixels[(y * mapObject.Map.Columns() + x) * 4 + 3] = byte(a >> 8)

            /*
            image, err := mapObject.GetTileImage(tileX, tileY, animationCounter)
            if err == nil {
                options.GeoM = geom
                // options.GeoM.Reset()
                options.GeoM.Translate(float64(x * tileWidth), float64(y * tileHeight))
                screen.DrawImage(image, &options)
            } else {
                log.Printf("Unable to render tilte at %d, %d: %v", tileX, tileY, err)
            }
            */
        }
    }

    var options ebiten.DrawImageOptions
    options.GeoM = geom
    mini := ebiten.NewImage(mapObject.Map.Columns(), mapObject.Map.Rows())
    mini.WritePixels(pixels)
    screen.DrawImage(mini, &options)
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
