package game

import (
    "log"
    "math"

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

    if image, ok := mapObject.TileCache[tile]; ok {
        return image, nil
    }

    tileInfo := mapObject.Data.Tiles[tile]

    gpuImage := ebiten.NewImageFromImage(tileInfo.Images[animationCounter % uint64(len(tileInfo.Images))])

    mapObject.TileCache[tile] = gpuImage
    return gpuImage, nil
}

func (mapObject *Map) Draw(cameraX int, cameraY int, animationCounter uint64, screen *ebiten.Image, geom ebiten.GeoM){

    tileWidth := mapObject.TileWidth()
    tileHeight := mapObject.TileHeight()

    tilesPerRow := int(math.Ceil(float64(screen.Bounds().Dx()) / float64(tileWidth)))
    tilesPerColumn := int(math.Ceil(float64(screen.Bounds().Dy()) / float64(tileHeight)))

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
