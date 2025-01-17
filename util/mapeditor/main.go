package main

import (
    "os"
    "fmt"
    "log"
    "time"
    "image/color"
    "math/rand"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/util/common"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/data"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
)

const ScreenWidth = 1024
const ScreenHeight = 768


type Editor struct {
    Data *terrain.TerrainData
    Font *text.GoTextFaceSource

    Map *terrain.Map
    Plane data.Plane

    TileGpuCache map[int]*ebiten.Image

    TileX int
    TileY int

    CameraX float64
    CameraY float64

    Counter uint64
    Scale float64

    ShowInfo bool
    InfoImage *ebiten.Image
}

func chooseRandomElement[T any](values []T) T {
    index := rand.Intn(len(values))
    return values[index]
}

func (editor *Editor) clear() {
    for column := range(editor.Map.Columns()) {
        for row := range(editor.Map.Rows()) {
            editor.Map.Terrain[column][row] = terrain.TileOcean.Index(editor.Plane)
        }
    }
}

func (editor *Editor) togglePlane() {
    if editor.Plane == data.PlaneArcanus {
        editor.Plane = data.PlaneMyrror
    } else {
        editor.Plane = data.PlaneArcanus
    }
    editor.clear()
}

func (editor *Editor) Update() error {
    editor.Counter += 1

    var keys []ebiten.Key

    canScroll := editor.Counter % 2 == 0

    keys = make([]ebiten.Key, 0)
    keys = inpututil.AppendPressedKeys(keys)
    for _, key := range keys {
        switch key {
            case ebiten.KeyUp:
                if editor.CameraY > 0 && canScroll {
                    editor.CameraY -= 1.0 / editor.Scale
                }
            case ebiten.KeyDown:
                if int(editor.CameraY) < editor.Map.Rows() && canScroll {
                    editor.CameraY += 1.0 / editor.Scale
                }
            case ebiten.KeyLeft:
                if editor.CameraX > 0 && canScroll {
                    editor.CameraX -= 1.0 / editor.Scale
                }
            case ebiten.KeyRight:
                if int(editor.CameraX) < editor.Map.Columns() && canScroll {
                    editor.CameraX += 1.0 / editor.Scale
                }
            case ebiten.KeyMinus:
                editor.Scale *= 0.98
            case ebiten.KeyEqual:
                editor.Scale *= 1.02
        }
    }

    _, wheelY := ebiten.Wheel()
    editor.Scale *= 1 + float64(wheelY) * 0.1

    if editor.Scale < 0.2 {
        editor.Scale = 0.2
    }
    if editor.Scale > 2 {
        editor.Scale = 2
    }

    keys = make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyP:
                editor.togglePlane()
            case ebiten.KeyC:
                editor.clear()
            case ebiten.KeyG:
                start := time.Now()
                editor.Map = terrain.GenerateLandCellularAutomata(editor.Map.Rows(), editor.Map.Columns(), editor.Data, editor.Plane)
                end := time.Now()
                log.Printf("Generate land took %v", end.Sub(start))
            case ebiten.KeyS:
                start := time.Now()
                editor.Map.ResolveTiles(editor.Data, editor.Plane)
                end := time.Now()
                log.Printf("Resolve tiles took %v", end.Sub(start))
            case ebiten.KeyTab:
                editor.ShowInfo = !editor.ShowInfo
            case ebiten.KeyEscape, ebiten.KeyCapsLock:
                return ebiten.Termination
        }
    }

    leftShift := inpututil.KeyPressDuration(ebiten.KeyShiftLeft) > 0

    leftClick := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
    rightClick := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)

    xSize := editor.GetTileImage(0, 0).Bounds().Dx()
    ySize := editor.GetTileImage(0, 0).Bounds().Dy()

    x, y := ebiten.CursorPosition()
    x -= 10
    y -= 10
    x = int(float64(x) / (float64(xSize) * editor.Scale))
    y = int(float64(y) / (float64(ySize) * editor.Scale))

    x += int(editor.CameraX)
    y += int(editor.CameraY)

    editor.TileX = x
    editor.TileY = y

    if leftClick {
        if x >= 0 && x < editor.Map.Columns() && y >= 0 && y < editor.Map.Rows() {
            use := terrain.TileLand.Index(editor.Plane)

            if leftShift {
                use = terrain.TileOcean.Index(editor.Plane)
            }

            editor.Map.Terrain[x][y] = use
        }
    } else if rightClick {
        if x >= 0 && x < editor.Map.Columns() && y >= 0 && y < editor.Map.Rows() {
            resolved, err := editor.Map.ResolveTile(x, y, editor.Data, editor.Plane)
            if err == nil {
                editor.Map.Terrain[x][y] = resolved
            } else {
                fmt.Printf("Unable to resolve tile %v, %v: %v\n", x, y, err)
            }
        }
    }

    // fmt.Printf("TPS: %v\n", ebiten.ActualTPS())

    return nil
}

func (editor *Editor) GetTileImage(x int, y int) *ebiten.Image {
    index := editor.Map.Terrain[x][y]

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

    startX := 10.0
    startY := 10.0

    // log.Printf("Draw start")

    for y := 0; y < editor.Map.Rows(); y++ {
        for x := 0; x < editor.Map.Columns(); x++ {
            // xPos := startX + float64(x * xSize) //  * editor.Scale
            // yPos := startY + float64(y * ySize) // * editor.Scale
            xPos := float64(x * xSize)
            yPos := float64(y * ySize)

            xUse := x + int(editor.CameraX)
            yUse := y + int(editor.CameraY)

            if xUse >= 0 && xUse < editor.Map.Columns() && yUse >= 0 && yUse < editor.Map.Rows() {
                tileImage := editor.GetTileImage(xUse, yUse)
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(xPos), float64(yPos))
                options.GeoM.Scale(editor.Scale, editor.Scale)
                options.GeoM.Translate(startX, startY)
                screen.DrawImage(tileImage, &options)

                if editor.TileX == xUse && editor.TileY == yUse {
                    vector.StrokeRect(screen, float32(startX) + float32(xPos * editor.Scale), float32(startY) + float32(yPos * editor.Scale), float32(xSize) * float32(editor.Scale), float32(ySize) * float32(editor.Scale), 1.5, color.White, true)
                }
            }
        }
    }

    if editor.ShowInfo {
        editor.InfoImage.Fill(color.RGBA{32, 32, 32, 128})

        face := &text.GoTextFace{Source: editor.Font, Size: 13}
        op := &text.DrawOptions{}
        op.GeoM.Translate(1, 1)
        op.ColorScale.ScaleWithColor(color.White)
        text.Draw(editor.InfoImage, fmt.Sprintf("Map Dimensions: %vx%v", editor.Map.Columns(), editor.Map.Rows()), face, op)
        op.GeoM.Translate(0, face.Size + 2)
        value := -1
        var type_ terrain.TerrainType = terrain.Unknown

        if editor.TileX >= 0 && editor.TileX < editor.Map.Columns() && editor.TileY >= 0 && editor.TileY < editor.Map.Rows() {
            value = editor.Map.Terrain[editor.TileX][editor.TileY]
            type_ = editor.Data.Tiles[value].Tile.TerrainType()
        }

        text.Draw(editor.InfoImage, fmt.Sprintf("Tile: %v,%v: 0x%x %v", editor.TileX, editor.TileY, value, type_), face, op)

        if editor.TileX >= 0 && editor.TileX < editor.Map.Columns() && editor.TileY >= 0 && editor.TileY < editor.Map.Rows() {
            tileImage := editor.GetTileImage(editor.TileX, editor.TileY)
            var options ebiten.DrawImageOptions
            options.GeoM.Scale(1.5, 1.5)
            options.GeoM.Translate(1, face.Size * 3)
            editor.InfoImage.DrawImage(tileImage, &options)
        }

        var options ebiten.DrawImageOptions
        options.GeoM.Translate(2, 2)
        scale := 0.9
        options.ColorM.Scale(scale, scale, scale, scale)
        screen.DrawImage(editor.InfoImage, &options)
    }

    // log.Printf("Draw end")
}

func (editor *Editor) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return ScreenWidth, ScreenHeight
}

func MakeEditor(lbxFile *lbx.LbxFile) *Editor {
    font, err := common.LoadFont()
    if err != nil {
        fmt.Printf("Could not load font: %v\n", err)
        os.Exit(0)
    }

    data, err := terrain.ReadTerrainData(lbxFile)
    if err != nil {
        fmt.Printf("Could not read terrain data: %v\n", err)
        os.Exit(0)
    }

    return &Editor{
        Data: data,
        Font: font,
        Map: terrain.MakeMap(100, 200),
        TileGpuCache: make(map[int]*ebiten.Image),
        TileX: -1,
        TileY: -1,
        Scale: 1.0,
        CameraX: 0,
        CameraY: 0,
        ShowInfo: true,
        InfoImage: ebiten.NewImage(200, 100),
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
