package main

import (
    "os"
    "fmt"
    "log"
    "image/color"
    "math/rand"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type Editor struct {
    Data *terrain.TerrainData

    Terrain [][]int

    TileGpuCache map[int]*ebiten.Image

    TileX int
    TileY int
}

func chooseRandomElement(values []int) int {
    index := rand.Intn(len(values))
    return values[index]
}

func (editor *Editor) removeMyrror(tiles []int) []int {
    var out []int

    for _, tile := range tiles {
        if ! editor.Data.Tiles[tile].IsMyrror() {
            out = append(out, tile)
        }
    }

    return out
}

// given a position in the terrain matrix, find a tile that fits all the neighbors of the tile
func (editor *Editor) ResolveTile(x int, y int) (int, error) {

    matching := make(map[terrain.Direction]terrain.TerrainType)

    getDirection := func(x int, y int, direction terrain.Direction) terrain.TerrainType {
        index := editor.Terrain[y][x]
        return editor.Data.Tiles[index].Tile.GetDirection(direction)
    }

    if x > 0 {
        matching[terrain.West] = getDirection(x-1, y, terrain.East)
    }

    if x > 0 && y > 0 {
        matching[terrain.NorthWest] = getDirection(x-1, y-1, terrain.SouthEast)
    }

    if x > 0 && y < len(editor.Terrain) - 1 {
        matching[terrain.SouthWest] = getDirection(x-1, y+1, terrain.NorthEast)
    }

    if x < len(editor.Terrain[0]) - 1 {
        matching[terrain.East] = getDirection(x+1, y, terrain.West)
    }

    if y > 0 {
        matching[terrain.North] = getDirection(x, y-1, terrain.South)
    }

    if y < len(editor.Terrain) - 1 {
        matching[terrain.South] = getDirection(x, y+1, terrain.North)
    }

    if x < len(editor.Terrain[0]) - 1 && y > 0 {
        matching[terrain.NorthEast] = getDirection(x+1, y-1, terrain.SouthWest)
    }

    if x < len(editor.Terrain[0]) - 1 && y < len(editor.Terrain) - 1 {
        matching[terrain.SouthEast] = getDirection(x+1, y+1, terrain.NorthWest)
    }

    tiles := editor.Data.FindMatchingAllTiles(matching)
    if tiles == nil {
        return -1, fmt.Errorf("no matching tile for %v", matching)
    }

    return chooseRandomElement(editor.removeMyrror(tiles)), nil
}

func (editor *Editor) Update() error {
    var keys []ebiten.Key

    keys = make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyEscape, ebiten.KeyCapsLock:
                return ebiten.Termination
        }
    }

    leftClick := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
    rightClick := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)

    xSize := editor.GetTileImage(0, 0).Bounds().Dx()
    ySize := editor.GetTileImage(0, 0).Bounds().Dy()

    x, y := ebiten.CursorPosition()
    x -= 10
    y -= 10
    x /= xSize
    y /= ySize

    editor.TileX = x
    editor.TileY = y

    if leftClick {
        if x >= 0 && x < len(editor.Terrain[0]) && y >= 0 && y < len(editor.Terrain) {
            editor.Terrain[y][x] = terrain.TileLand.Index
        }
    } else if rightClick {
        resolved, err := editor.ResolveTile(x, y)
        if err == nil {
            editor.Terrain[y][x] = resolved
        } else {
            fmt.Printf("Unable to resolve tile %v, %v: %v\n", x, y, err)
        }
    }

    // fmt.Printf("TPS: %v\n", ebiten.ActualTPS())

    return nil
}

func (editor *Editor) GetTileImage(x int, y int) *ebiten.Image {
    index := editor.Terrain[y][x]

    use, ok := editor.TileGpuCache[index]
    if ok {
        return use
    }

    useImage := editor.Data.Tiles[index].Images[0]
    use = ebiten.NewImageFromImage(useImage)

    editor.TileGpuCache[index] = use

    return use
}

func (editor *Editor) Draw(screen *ebiten.Image){
    xSize := editor.GetTileImage(0, 0).Bounds().Dx()
    ySize := editor.GetTileImage(0, 0).Bounds().Dy()

    startX := 10
    startY := 10

    // log.Printf("Draw start")

    for y := 0; y < len(editor.Terrain); y++ {
        for x := 0; x < len(editor.Terrain[y]); x++ {
            xPos := startX + x * xSize
            yPos := startY + y * ySize

            tileImage := editor.GetTileImage(x, y)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(xPos), float64(yPos))
            screen.DrawImage(tileImage, &options)

            if editor.TileX == x && editor.TileY == y {
                vector.StrokeRect(screen, float32(xPos), float32(yPos), float32(xSize), float32(ySize), 1.5, color.White, true)
            }
        }
    }

    // log.Printf("Draw end")
}

func (editor *Editor) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return ScreenWidth, ScreenHeight
}

func createTerrain(rows int, columns int) [][]int {
    out := make([][]int, columns)
    for i := 0; i < columns; i++ {
        out[i] = make([]int, rows)
    }

    return out
}

func MakeEditor(lbxFile *lbx.LbxFile) *Editor {
    data, err := terrain.ReadTerrainData(lbxFile)
    if err != nil {
        fmt.Printf("Could not read terrain data: %v\n", err)
        os.Exit(0)
    }

    return &Editor{
        Data: data,
        Terrain: createTerrain(50, 50),
        TileGpuCache: make(map[int]*ebiten.Image),
        TileX: -1,
        TileY: -1,
    }
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    if len(os.Args) < 2 {
        fmt.Printf("Give the terrain.lbx file as an argument\n")
        return
    }

    path := os.Args[1]
    file, err := os.Open(path)
    if err != nil {
        fmt.Printf("Could not open lbx file %v: %v\n", path, err)
        return
    }
    lbxData, err := lbx.ReadLbx(file)
    if err != nil {
        fmt.Printf("Could read lbx file %v: %v\n", path, err)
        return
    }
    file.Close()

    editor := MakeEditor(&lbxData)

    ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
    ebiten.SetWindowTitle("map editor")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
    
    err = ebiten.RunGame(editor)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
