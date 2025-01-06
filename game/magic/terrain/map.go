package terrain

import (
    "fmt"
    "image"
    "math/rand/v2"
    "math"

    "github.com/kazzmir/master-of-magic/game/magic/data"
)

// a continent is a list of points that are indices into the Terrain matrix
type Continent []image.Point

func (continent Continent) Size() int {
    return len(continent)
}

func (continent Continent) Contains(point image.Point) bool {
    for _, check := range continent {
        if check == point {
            return true
        }
    }

    return false
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

type FloodFunc func(x int, y int) bool
func (map_ *Map) FloodWalk(x int, y int, f FloodFunc){
    var walk func(x int, y int)

    walk = func(x int, y int){
        for dx := -1; dx <= 1; dx++ {
            for dy := -1; dy <= 1; dy++ {
                nx := x + dx
                ny := y + dy

                if nx >= 0 && nx < map_.Columns() && ny >= 0 && ny < map_.Rows() {
                    if f(nx, ny) {
                        walk(nx, ny)
                    }
                }
            }
        }
    }

    walk(x, y)
}

func (map_ *Map) FindContinents(plane data.Plane) []Continent {

    seen := makeCells(map_.Rows(), map_.Columns())

    var searchTiles func(x int, y int, continent *Continent)

    rows := map_.Rows()
    columns := map_.Columns()

    // count all tiles connected to this one of the same kind
    searchTiles = func(startX int, startY int, continent *Continent){

        search := []image.Point{image.Pt(startX, startY)}

        for len(search) > 0 {

            point := search[0]
            search = search[1:]
            x := point.X
            y := point.Y

            if seen[x][y] == true {
                continue
            }

            seen[x][y] = true

            for dx := -1; dx <= 1; dx++ {
                for dy := -1; dy <= 1; dy++ {
                    nx := x + dx
                    ny := y + dy

                    if nx >= 0 && nx < columns && ny >= 0 && ny < rows {
                        if map_.Terrain[nx][ny] == TileLand.Index(plane) {
                            *continent = append(*continent, image.Pt(nx, ny))
                            // searchTiles(nx, ny, continent)
                            search = append(search, image.Pt(nx, ny))
                        }
                    }
                }
            }
        }
    }

    var continents []Continent

    for x := 0; x < map_.Columns(); x++ {
        for y := 0; y < map_.Rows(); y++ {
            if map_.Terrain[x][y] == TileLand.Index(plane) && seen[x][y] == false {
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

func chooseRandomElement[T any](values []T) T {
    index := rand.IntN(len(values))
    return values[index]
}

/*
func (editor *Editor) removeMyrror(tiles []int) []int {
    var out []int

    for _, tile := range tiles {
        if ! editor.Data.Tiles[tile].IsMyrror() {
            out = append(out, tile)
        }
    }

    return out
}
*/

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
        nx := x + dx

        for nx < 0 {
            nx += len(cells)
        }

        nx = nx % len(cells)

        for dy := -1; dy <= 1; dy++ {

            if dx == 0 && dy == 0 {
                continue
            }

            ny := y + dy

            if ny < 0 || ny >= len(cells[0]) {
                continue
            }

            if cells[nx][ny] {
                count += 1
            }
        }
    }

    return count
}

func (map_ *Map) GenerateLandCellularAutomata(plane data.Plane){
    cells := makeCells(map_.Rows(), map_.Columns())
    tmpCells := makeCells(map_.Rows(), map_.Columns())

    cellRounds := 5

    const deathRate = 3
    const birthRate = 3

    stepCells := func(cells [][]bool, tmpCells [][]bool) {
        for x := 0; x < map_.Columns(); x++ {
            for y := 1; y < map_.Rows() - 1; y++ {
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
        x := rand.IntN(map_.Columns())
        y := rand.IntN(map_.Rows())

        cells[x][y] = rand.IntN(2) == 1
    }

    for i := 0; i < cellRounds; i++ {
        // kill some cells randomly
        for z := 0; z < int(float64(len(cells[0]) * len(cells)) * 0.03); z++ {
            x := rand.IntN(map_.Columns())
            y := rand.IntN(map_.Rows())

            cells[x][y] = false
        }

        stepCells(cells, tmpCells)
        cells, tmpCells = tmpCells, cells
    }

    for x := 0; x < map_.Columns(); x++ {
        for y := 0; y < map_.Rows(); y++ {
            if cells[x][y] {
                map_.Terrain[x][y] = TileLand.Index(plane)
            } else {
                map_.Terrain[x][y] = TileOcean.Index(plane)
            }
        }
    }
}

func GenerateLandCellularAutomata(rows int, columns int, data *TerrainData, plane data.Plane) *Map {
    // run a cellular automata simulation for a few rounds to generate
    // land and ocean tiles. then call ResolveTiles() to clean up the edges
    map_ := MakeMap(rows, columns)
    map_.GenerateLandCellularAutomata(plane)

    map_.RemoveSmallIslands(100, plane)

    /*
    continents := editor.Map.FindContinents()
    log.Printf("Continents: %v\n", len(continents))
    */

    map_.PlaceRandomTerrainTiles(plane)

    // start := time.Now()
    map_.ResolveTiles(data, plane)
    return map_
    // end := time.Now()
    // log.Printf("Resolve tiles took %v", end.Sub(start))
}

// put down other tiles like forests, mountains, special nodes, etc
func (map_ *Map) PlaceRandomTerrainTiles(plane data.Plane){

    continents := map_.FindContinents(plane)

    randomForest := func() int {
        choices := []int{
            TileForest1.Index(plane),
            TileForest2.Index(plane),
            TileForest3.Index(plane),
        }

        return chooseRandomElement(choices)
    }

    for _, continent := range continents {

        for i := 0; i < int(math.Sqrt(float64(continent.Size()))) / 4; i++ {
            point := chooseRandomElement(continent)

            use := TileSorceryLake.Index(plane)
            switch rand.IntN(2) {
                case 0: use = randomForest()
                case 1: use = TileMountain1.Index(plane)
            }

            map_.Terrain[point.X][point.Y] = use
        }

        for i := 0; i < int(math.Sqrt(float64(continent.Size()))) / 8; i++ {
            point := chooseRandomElement(continent)

            use := TileSorceryLake.Index(plane)
            switch rand.IntN(4) {
                case 0: use = TileSorceryLake.Index(plane)
                case 1: use = TileNatureForest.Index(plane)
                case 2: use = TileChaosVolcano.Index(plane)
                case 3: use = TileLake.Index(plane)
            }

            map_.Terrain[point.X][point.Y] = use
        }
    }
}

// remove land masses that contain less squares than 'area'
func (map_ *Map) RemoveSmallIslands(area int, plane data.Plane){
    continents := map_.FindContinents(plane)

    for _, continent := range continents {
        if continent.Size() < area {
            for _, point := range continent {
                map_.Terrain[point.X][point.Y] = TileOcean.Index(plane)
            }
        }
    }
}

// given a position in the terrain matrix, find a tile that fits all the neighbors of the tile
func (map_ *Map) ResolveTile(x int, y int, data *TerrainData, plane data.Plane) (int, error) {

    matching := make(map[Direction]TerrainType)

    getDirection := func(x int, y int, direction Direction) TerrainType {
        for x < 0 {
            x += map_.Columns()
        }

        x = x % map_.Columns()

        index := map_.Terrain[x][y]
        if index < 0 || index >= len(data.Tiles) {
            fmt.Printf("Error: invalid index in terrain %v at %v,%v\n", index, x, y)
            return Unknown
        }
        return data.Tiles[index].Tile.GetDirection(direction)
    }

    matching[West] = getDirection(x-1, y, East)

    matching[NorthWest] = Ocean
    if y > 0 {
        matching[NorthWest] = getDirection(x-1, y-1, SouthEast)
    }

    matching[SouthWest] = Ocean
    if y < map_.Rows() - 1 {
        matching[SouthWest] = getDirection(x-1, y+1, NorthEast)
    }

    matching[East] = getDirection(x+1, y, West)

    matching[North] = Ocean
    if y > 0 {
        matching[North] = getDirection(x, y-1, South)
    }

    matching[South] = Ocean
    if y < map_.Rows() - 1 {
        matching[South] = getDirection(x, y+1, North)
    }

    matching[NorthEast] = Ocean
    if y > 0 {
        matching[NorthEast] = getDirection(x+1, y-1, SouthWest)
    }

    matching[SouthEast] = Ocean
    if y < map_.Rows() - 1 {
        matching[SouthEast] = getDirection(x+1, y+1, NorthWest)
    }

    if data.Tiles[map_.Terrain[x][y]].Tile.Matches(matching) {
        return map_.Terrain[x][y], nil
    }

    tile := data.FindMatchingTile(matching, plane)
    if tile == -1 {
        // try to fill edges if tiles are not found
        if matching[North] == matching[East] {
            matching[NorthEast] = matching[North]
        }
        if matching[North] == matching[West] {
            matching[NorthWest] = matching[North]
        }
        if matching[South] == matching[East] {
            matching[SouthEast] = matching[South]
        }
        if matching[South] == matching[West] {
            matching[SouthWest] = matching[South]
        }
        tile = data.FindMatchingTile(matching, plane)
    }

    if tile == -1 {
        return -1, fmt.Errorf("no matching tile for %v", matching)
    }

    return tile, nil

    // return chooseRandomElement(editor.removeMyrror(tiles)), nil
    // return editor.removeMyrror(tiles)[0], nil
}

func (map_ *Map) ResolveTiles(data *TerrainData, plane data.Plane){
    // go through every tile and try to resolve it, keep doing this in a loop until there are no more tiles to resolve

    var unresolved []image.Point
    for x := 0; x < map_.Columns(); x++ {
        for y := 0; y < map_.Rows(); y++ {
            unresolved = append(unresolved, image.Pt(x, y))
        }
    }

    count := 0
    for len(unresolved) > 0 && count < 5 {
        count += 1
        var more []image.Point

        for _, index := range rand.Perm(len(unresolved)) {
            point := unresolved[index]
            choice, err := map_.ResolveTile(point.X, point.Y, data, plane)
            if err != nil {
                more = append(more, point)
            } else if choice != map_.Terrain[point.X][point.Y] {
                map_.Terrain[point.X][point.Y] = choice
            }
        }

        unresolved = more

        // fmt.Printf("resolve loop %d\n", count)
    }
}
