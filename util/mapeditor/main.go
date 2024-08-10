package main

import (
    "os"
    "fmt"
    "log"
    "image/color"

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

    if leftClick {
        x, y := ebiten.CursorPosition()
        x -= 10
        y -= 10
        x /= 20
        y /= 20

        if x >= 0 && x < len(editor.Terrain[0]) && y >= 0 && y < len(editor.Terrain) {
            editor.Terrain[y][x] = terrain.TileLand.Index
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
    size := 20

    startX := 10
    startY := 10

    // log.Printf("Draw start")

    for y := 0; y < len(editor.Terrain); y++ {
        for x := 0; x < len(editor.Terrain[y]); x++ {
            xPos := startX + x * size
            yPos := startY + y * size

            tileImage := editor.GetTileImage(x, y)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(xPos), float64(yPos))
            screen.DrawImage(tileImage, &options)

            if 2 < 1 {
                vector.StrokeRect(screen, float32(xPos), float32(yPos), float32(size), float32(size), 1.5, color.White, true)
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
