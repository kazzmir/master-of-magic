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

// a continent is a list of points that are indices into the Terrain matrix
type Continent []image.Point

func (continent Continent) Size() int {
    return len(continent)
}

type Map struct {
    Terrain [][]int
}

func (map_ *Map) Rows() int {
    return len(map_.Terrain[0])
}

func (map_ *Map) Columns() int {
    return len(map_.Terrain)
}

func (map_ *Map) FindContinents() []Continent {

    seen := makeCells(map_.Rows(), map_.Columns())

    var searchTiles func(x int, y int, continent *Continent)

    // count all tiles connected to this one of the same kind
    searchTiles = func(x int, y int, continent *Continent){
        if seen[x][y] == true {
            return
        }

        seen[x][y] = true

        for dx := -1; dx <= 1; dx++ {
            for dy := -1; dy <= 1; dy++ {
                nx := x + dx
                ny := y + dy

                if nx >= 0 && nx < map_.Columns() && ny >= 0 && ny < map_.Rows() {
                    if map_.Terrain[nx][ny] == terrain.TileLand.Index {
                        *continent = append(*continent, image.Pt(nx, ny))
                        searchTiles(nx, ny, continent)
                    }
                }
            }
        }
    }

    var continents []Continent

    for x := 0; x < map_.Columns(); x++ {
        for y := 0; y < map_.Rows(); y++ {
            if map_.Terrain[x][y] == terrain.TileLand.Index && seen[x][y] == false {
                var continent Continent
                continent = append(continent, image.Pt(x, y))
                searchTiles(x, y, &continent)
                continents = append(continents, continent)
            }
        }
    }

    return continents
}

func MakeMap(rows int, columns int) *Map {
    terrain := make([][]int, columns)
    for i := 0; i < columns; i++ {
        terrain[i] = make([]int, rows)
    }

    return &Map{
        Terrain: terrain,
    }
}

type Editor struct {
    Data *terrain.TerrainData
    Font *text.GoTextFaceSource

    Map *Map

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

func makeCells(rows int, columns int) [][]bool {
    out := make([][]bool, columns)
    for i := 0; i < columns; i++ {
        out[i] = make([]bool, rows)
    }

    return out
}

func countNeighbors(cells [][]bool, x int, y int) int {
    count := 0

    for dx := -1; dx <= 1; dx++ {
        for dy := -1; dy <= 1; dy++ {

            if dx == 0 && dy == 0 {
                continue
            }

            nx := x + dx
            ny := y + dy

            if nx >= 0 && nx < len(cells) && ny >= 0 && ny < len(cells[0]) {
                if cells[nx][ny] {
                    count += 1
                }
            }
        }
    }

    return count
}

func (editor *Editor) GenerateLandCellularAutomata(){
    // run a cellular automata simulation for a few rounds to generate
    // land and ocean tiles. then call ResolveTiles() to clean up the edges

    cells := makeCells(editor.Map.Rows(), editor.Map.Columns())
    tmpCells := makeCells(editor.Map.Rows(), editor.Map.Columns())

    cellRounds := 5

    const deathRate = 3
    const birthRate = 3

    stepCells := func(cells [][]bool, tmpCells [][]bool) {
        for x := 0; x < editor.Map.Columns(); x++ {
            for y := 0; y < editor.Map.Rows(); y++ {
                neighbors := countNeighbors(cells, x, y)

                if cells[x][y] {
                    if neighbors < deathRate {
                        tmpCells[x][y] = false
                    } else {
                        tmpCells[x][y] = true
                    }
                } else {
                    if neighbors < birthRate {
                        tmpCells[x][y] = false
                    } else {
                        tmpCells[x][y] = true
                    }
                }
            }
        }
    }

    // set some cells to be alive
    max := float64(len(cells) * len(cells[0])) * 0.6
    for i := 0; i < int(max); i++ {
        x := rand.Intn(editor.Map.Columns())
        y := rand.Intn(editor.Map.Rows())

        cells[x][y] = rand.Intn(2) == 1
    }

    for i := 0; i < cellRounds; i++ {
        // kill some cells randomly
        for z := 0; z < int(float64(len(cells[0]) * len(cells)) * 0.03); z++ {
            x := rand.Intn(editor.Map.Columns())
            y := rand.Intn(editor.Map.Rows())

            cells[x][y] = false
        }

        stepCells(cells, tmpCells)
        cells, tmpCells = tmpCells, cells
    }

    for x := 0; x < editor.Map.Columns(); x++ {
        for y := 0; y < editor.Map.Rows(); y++ {
            if cells[x][y] {
                editor.Map.Terrain[x][y] = terrain.TileLand.Index
            } else {
                editor.Map.Terrain[x][y] = terrain.TileOcean.Index
            }
        }
    }

    editor.RemoveSmallIslands(100)

    continents := editor.Map.FindContinents()
    log.Printf("Continents: %v\n", len(continents))

    editor.PlaceRandomTerrainTiles()

    editor.ResolveTiles()
}

// put down other tiles like forests, mountains, special nodes, etc
func (editor *Editor) PlaceRandomTerrainTiles(){

    for i := 0; i < 10; i++ {

        for try := 0; try < 10; try++ {
            x := rand.Intn(editor.Map.Columns())
            y := rand.Intn(editor.Map.Rows())

            if editor.Map.Terrain[x][y] == terrain.TileLand.Index {
                editor.Map.Terrain[x][y] = terrain.TileLake.Index
                break
            }
        }

        for try := 0; try < 10; try++ {
            x := rand.Intn(editor.Map.Columns())
            y := rand.Intn(editor.Map.Rows())

            if editor.Map.Terrain[x][y] == terrain.TileLand.Index {
                use := terrain.TileSorceryLake.Index
                switch rand.Intn(3) {
                    case 0: use = terrain.TileSorceryLake.Index
                    case 1: use = terrain.TileNatureForest.Index
                    case 2: use = terrain.TileChaosVolcano.Index
                }

                editor.Map.Terrain[x][y] = use
            }
        }

    }

}

// remove land masses that contain less squares than 'area'
func (editor *Editor) RemoveSmallIslands(area int){

    seen := makeCells(editor.Map.Rows(), editor.Map.Columns())

    var countTiles func(x int, y int, kind int) int

    // count all tiles connected to this one of the same kind
    countTiles = func(x int, y int, kind int) int {
        if seen[x][y] == true {
            return 0
        }

        count := 1
        seen[x][y] = true

        for dx := -1; dx <= 1; dx++ {
            for dy := -1; dy <= 1; dy++ {
                nx := x + dx
                ny := y + dy

                if nx >= 0 && nx < editor.Map.Columns() && ny >= 0 && ny < editor.Map.Rows() {
                    if editor.Map.Terrain[nx][ny] == kind {
                        count += countTiles(nx, ny, kind)
                    }
                }
            }
        }

        return count
    }

    var floodFill func(x int, y int, what int, kind int)

    floodFill = func(x int, y int, what int, kind int) {
        for dx := -1; dx <= 1; dx++ {
            for dy := -1; dy <= 1; dy++ {
                nx := x + dx
                ny := y + dy

                if nx >= 0 && nx < editor.Map.Columns() && ny >= 0 && ny < editor.Map.Rows() {
                    if editor.Map.Terrain[nx][ny] == what {
                        editor.Map.Terrain[nx][ny] = kind
                        floodFill(nx, ny, what, kind)
                    }
                }
            }
        }
    }

    for x := 0; x < editor.Map.Columns(); x++ {
        for y := 0; y < editor.Map.Rows(); y++ {
            if seen[x][y] {
                continue
            }

            if editor.Map.Terrain[x][y] == terrain.TileLand.Index {
                count := countTiles(x, y, terrain.TileLand.Index)
                if count < area {
                    floodFill(x, y, terrain.TileLand.Index, terrain.TileOcean.Index)
                }
            }
        }
    }

}

func (editor *Editor) GenerateLand1() {
    // create a matrix of floats the same dimensions as the terrain
    // fill in matrix with random values between -1,1
    // do a few rounds of averaging out the cells with their neighbors
    // for every cell below some threshold, put an ocean tile there.
    // every cell above the threshold, put a land tile
    // finally, end by calling ResolveTiles() to clean up edges

    const threshold = 0.0
    const smoothRounds = 4

    data := make([][]float32, editor.Map.Columns())
    for x := 0; x < len(data); x++ {
        data[x] = make([]float32, editor.Map.Rows())

        for y := 0; y < len(data[x]); y++ {
            data[x][y] = rand.Float32() * 2 - 1
        }
    }

    for i := 0; i < smoothRounds; i++ {
        data = averageCells(data)
    }

    for x := 0; x < len(data); x++ {
        for y := 0; y < len(data[0]); y++ {
            if data[x][y] < threshold {
                editor.Map.Terrain[x][y] = terrain.TileOcean.Index
            } else {
                editor.Map.Terrain[x][y] = terrain.TileLand.Index
            }
        }
    }

    editor.ResolveTiles()
}

// given a position in the terrain matrix, find a tile that fits all the neighbors of the tile
func (editor *Editor) ResolveTile(x int, y int) (int, error) {

    matching := make(map[terrain.Direction]terrain.TerrainType)

    getDirection := func(x int, y int, direction terrain.Direction) terrain.TerrainType {
        index := editor.Map.Terrain[x][y]
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

    if x > 0 && y < editor.Map.Rows() - 1 {
        matching[terrain.SouthWest] = getDirection(x-1, y+1, terrain.NorthEast)
    }

    if x < editor.Map.Columns() - 1 {
        matching[terrain.East] = getDirection(x+1, y, terrain.West)
    }

    if y > 0 {
        matching[terrain.North] = getDirection(x, y-1, terrain.South)
    }

    if y < editor.Map.Rows() - 1 {
        matching[terrain.South] = getDirection(x, y+1, terrain.North)
    }

    if x < editor.Map.Columns() - 1 && y > 0 {
        matching[terrain.NorthEast] = getDirection(x+1, y-1, terrain.SouthWest)
    }

    if x < editor.Map.Columns() - 1 && y < editor.Map.Rows() - 1 {
        matching[terrain.SouthEast] = getDirection(x+1, y+1, terrain.NorthWest)
    }

    if editor.Data.Tiles[editor.Map.Terrain[x][y]].Tile.Matches(matching) {
        return editor.Map.Terrain[x][y], nil
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
    for x := 0; x < editor.Map.Columns(); x++ {
        for y := 0; y < editor.Map.Rows(); y++ {
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
            } else if choice != editor.Map.Terrain[point.X][point.Y] {
                editor.Map.Terrain[point.X][point.Y] = choice
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
            case ebiten.KeyG:
                start := time.Now()
                editor.GenerateLandCellularAutomata()
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

    x += int(editor.CameraX)
    y += int(editor.CameraY)

    editor.TileX = x
    editor.TileY = y

    if leftClick {
        if x >= 0 && x < editor.Map.Columns() && y >= 0 && y < editor.Map.Rows() {
            use := terrain.TileLand.Index

            if leftShift {
                use = terrain.TileOcean.Index
            }

            editor.Map.Terrain[x][y] = use
        }
    } else if rightClick {
        if x >= 0 && x < editor.Map.Columns() && y >= 0 && y < editor.Map.Rows() {
            resolved, err := editor.ResolveTile(x, y)
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

        if editor.TileX >= 0 && editor.TileX < editor.Map.Columns() && editor.TileY >= 0 && editor.TileY < editor.Map.Rows() {
            value = editor.Map.Terrain[editor.TileX][editor.TileY]
        }

        text.Draw(editor.InfoImage, fmt.Sprintf("Tile: %v,%v: %v (0x%x)", editor.TileX, editor.TileY, value, value), face, op)

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
        Map: MakeMap(100, 200),
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
