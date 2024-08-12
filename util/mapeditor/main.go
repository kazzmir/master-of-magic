package main

import (
    "os"
    "fmt"
    "log"
    "time"
    "image"
    "image/color"
    "math/rand"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/util/common"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"

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

    Terrain [][]int

    TileGpuCache map[int]*ebiten.Image

    TileX int
    TileY int

    CameraX int
    CameraY int

    Counter uint64
    Scale float64

    ShowInfo bool
    InfoImage *ebiten.Image
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

func averageCell(data [][]float32, cx int, cy int) float32 {
    var total float32 = 0
    count := 0

    for x := -1; x <= 1; x++ {
        for y := -1; y <= 1; y++ {
            nx := cx + x
            ny := cy + y

            if nx >= 0 && nx < len(data[0]) && ny >= 0 && ny < len(data) {
                total += data[ny][nx]
                count += 1
            }
        }
    }

    return total / float32(count)
}

func averageCells(data [][]float32) [][]float32 {
    out := make([][]float32, len(data))
    for y := 0; y < len(data); y++ {
        out[y] = make([]float32, len(data[0]))
    }

    for x := 0; x < len(data[0]); x++ {
        for y := 0; y < len(data); y++ {
            out[y][x] = averageCell(data, x, y)
        }
    }

    return out
}

func (editor *Editor) GenerateLand() {
    // create a matrix of floats the same dimensions as the terrain
    // fill in matrix with random values between -1,1
    // do a few rounds of averaging out the cells with their neighbors
    // for every cell below some threshold, put an ocean tile there.
    // every cell above the threshold, put a land tile
    // finally, end by calling ResolveTiles() to clean up edges

    const threshold = 0.0
    const smoothRounds = 4

    data := make([][]float32, len(editor.Terrain))
    for y := 0; y < len(data); y++ {
        data[y] = make([]float32, len(editor.Terrain[0]))

        for x := 0; x < len(data[y]); x++ {
            data[y][x] = rand.Float32() * 2 - 1
        }
    }

    for i := 0; i < smoothRounds; i++ {
        data = averageCells(data)
    }

    for x := 0; x < len(data[0]); x++ {
        for y := 0; y < len(data); y++ {
            if data[y][x] < threshold {
                editor.Terrain[y][x] = terrain.TileOcean.Index
            } else {
                editor.Terrain[y][x] = terrain.TileLand.Index
            }
        }
    }

    editor.ResolveTiles()
}

// given a position in the terrain matrix, find a tile that fits all the neighbors of the tile
func (editor *Editor) ResolveTile(x int, y int) (int, error) {

    matching := make(map[terrain.Direction]terrain.TerrainType)

    getDirection := func(x int, y int, direction terrain.Direction) terrain.TerrainType {
        index := editor.Terrain[y][x]
        if index < 0 || index >= len(editor.Data.Tiles) {
            fmt.Printf("Error: invalid index in terrain %v at %v,%v\n", index, x, y)
            return terrain.Unknown
        }
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

    tile := editor.Data.FindMatchingTile(matching)
    if tile == -1 {
        return -1, fmt.Errorf("no matching tile for %v", matching)
    }

    return tile, nil

    // return chooseRandomElement(editor.removeMyrror(tiles)), nil
    // return editor.removeMyrror(tiles)[0], nil
}

func (editor *Editor) ResolveTiles(){
    // go through every tile and try to resolve it, keep doing this in a loop until there are no more tiles to resolve

    var unresolved []image.Point
    for x := 0; x < len(editor.Terrain[0]); x++ {
        for y := 0; y < len(editor.Terrain); y++ {
            unresolved = append(unresolved, image.Pt(x, y))
        }
    }

    count := 0
    for len(unresolved) > 0 && count < 5 {
        count += 1
        var more []image.Point

        for _, index := range rand.Perm(len(unresolved)) {
            point := unresolved[index]
            choice, err := editor.ResolveTile(point.X, point.Y)
            if err != nil {
                more = append(more, point)
            } else if choice != editor.Terrain[point.Y][point.X] {
                editor.Terrain[point.Y][point.X] = choice
            }
        }

        unresolved = more

        // fmt.Printf("resolve loop %d\n", count)
    }
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
                    editor.CameraY -= 1
                }
            case ebiten.KeyDown:
                if editor.CameraY < len(editor.Terrain[0]) && canScroll {
                    editor.CameraY += 1
                }
            case ebiten.KeyLeft:
                if editor.CameraX > 0 && canScroll {
                    editor.CameraX -= 1
                }
            case ebiten.KeyRight:
                if editor.CameraX < len(editor.Terrain) && canScroll {
                    editor.CameraX += 1
                }
            case ebiten.KeyMinus:
                editor.Scale *= 0.98
                if editor.Scale < 0.2 {
                    editor.Scale = 0.2
                }

            case ebiten.KeyEqual:
                editor.Scale *= 1.02
                if editor.Scale > 2 {
                    editor.Scale = 2
                }
        }
    }

    keys = make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyG:
                start := time.Now()
                editor.GenerateLand()
                end := time.Now()
                log.Printf("Generate land took %v", end.Sub(start))
            case ebiten.KeyS:
                start := time.Now()
                editor.ResolveTiles()
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
    // rightClick := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)
    rightClick := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)

    xSize := editor.GetTileImage(0, 0).Bounds().Dx()
    ySize := editor.GetTileImage(0, 0).Bounds().Dy()

    x, y := ebiten.CursorPosition()
    x -= 10
    y -= 10
    x = int(float64(x) / (float64(xSize) * editor.Scale))
    y = int(float64(y) / (float64(ySize) * editor.Scale))

    x += editor.CameraX
    y += editor.CameraY

    editor.TileX = x
    editor.TileY = y

    if leftClick {
        if x >= 0 && x < len(editor.Terrain[0]) && y >= 0 && y < len(editor.Terrain) {
            use := terrain.TileLand.Index

            if leftShift {
                use = terrain.TileOcean.Index
            }

            editor.Terrain[y][x] = use
        }
    } else if rightClick {
        if x >= 0 && x < len(editor.Terrain[0]) && y >= 0 && y < len(editor.Terrain) {
            resolved, err := editor.ResolveTile(x, y)
            if err == nil {
                editor.Terrain[y][x] = resolved
            } else {
                fmt.Printf("Unable to resolve tile %v, %v: %v\n", x, y, err)
            }
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

    startX := 10.0
    startY := 10.0

    // log.Printf("Draw start")

    for y := 0; y < len(editor.Terrain); y++ {
        for x := 0; x < len(editor.Terrain[y]); x++ {
            // xPos := startX + float64(x * xSize) //  * editor.Scale
            // yPos := startY + float64(y * ySize) // * editor.Scale
            xPos := float64(x * xSize)
            yPos := float64(y * ySize)

            xUse := x + editor.CameraX
            yUse := y + editor.CameraY

            if xUse >= 0 && xUse < len(editor.Terrain[0]) && yUse >= 0 && yUse < len(editor.Terrain) {
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
        text.Draw(editor.InfoImage, fmt.Sprintf("Map Dimensions: %vx%v", len(editor.Terrain[0]), len(editor.Terrain)), face, op)
        op.GeoM.Translate(0, face.Size + 2)
        text.Draw(editor.InfoImage, fmt.Sprintf("Tile: %v,%v", editor.TileX, editor.TileY), face, op)

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

func createTerrain(rows int, columns int) [][]int {
    out := make([][]int, columns)
    for i := 0; i < columns; i++ {
        out[i] = make([]int, rows)
    }

    return out
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
        Terrain: createTerrain(200, 100),
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
