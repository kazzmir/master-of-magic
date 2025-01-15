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

func (map_ *Map) FindContinents() []Continent {

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
                        if GetTile(map_.Terrain[nx][ny]).IsLand() {
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
            if GetTile(map_.Terrain[x][y]).IsLand() && seen[x][y] == false {
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

func choseRandomWeightedElement[T any](values []T, weights []int) T {
    totalWeight := 0
    for _, weight := range weights {
        totalWeight += weight
    }

    totalWeight = rand.IntN(totalWeight)

    for index, value := range values {
        weight := weights[index]
        if totalWeight < weight {
            return value
        }
        totalWeight -= weight
    }

    return values[0]
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

    cellRounds := 4

    const deathRate = 3
    const birthRate = 3

    stepCells := func(cells [][]bool, tmpCells [][]bool) {
        for x := 0; x < map_.Columns(); x++ {
            for y := 5; y < map_.Rows() - 5; y++ {
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

    continents := map_.FindContinents()

    randomGrasslands := func(y int) int {
        choices := []int{
            TileGrasslands1.Index(plane),
            TileGrasslands2.Index(plane),
            TileGrasslands3.Index(plane),
            TileGrasslands4.Index(plane),
            TileTundra.Index(plane),
            TileAllDesert1.Index(plane),
        }

        yRel := float64(y) / float64(map_.Rows())

        tundraWeight := 0
        switch {
            case yRel < 0.2 || yRel > 0.8: tundraWeight = 10
            case yRel < 0.1 || yRel > 0.9: tundraWeight = 50
        }

        desertWeight := 0
        switch {
            case yRel > 0.4 && yRel < 0.6: desertWeight = 10
            case yRel > 0.45 && yRel < 0.55: desertWeight = 50
        }

        weights := []int{1, 1, 1, 1, tundraWeight, desertWeight}

        return choseRandomWeightedElement(choices, weights)
    }

    randomForest := func() int {
        choices := []int{
            TileForest1.Index(plane),
            TileForest2.Index(plane),
            TileForest3.Index(plane),
        }

        return chooseRandomElement(choices)
    }

    for _, continent := range continents {

        for i := 0; i < continent.Size() * 2; i++ {
            point := chooseRandomElement(continent)

            choices := []int{
                randomGrasslands(point.Y),
                randomForest(),
                TileSwamp2.Index(plane),
                TileHills1.Index(plane),
                TileMountain1.Index(plane),
            }
            weights := []int{20, 10, 1, 10, 5}

            map_.Terrain[point.X][point.Y] = choseRandomWeightedElement(choices, weights)
        }

        for i := 0; i < int(math.Sqrt(float64(continent.Size()))) / 8; i++ {
            point := chooseRandomElement(continent)

            var use int
            switch rand.IntN(3) {
                case 0: use = TileSorceryLake.Index(plane)
                case 1: use = TileNatureForest.Index(plane)
                case 2: use = TileChaosVolcano.Index(plane)
            }

            map_.Terrain[point.X][point.Y] = use
        }
    }
}

// remove land masses that contain less squares than 'area'
func (map_ *Map) RemoveSmallIslands(area int, plane data.Plane){
    continents := map_.FindContinents()

    for _, continent := range continents {
        if continent.Size() < area {
            for _, point := range continent {
                map_.Terrain[point.X][point.Y] = TileOcean.Index(plane)
            }
        }
    }
}

func (map_ *Map) getTerrainAt(x int, y int, data *TerrainData) TerrainType {
    if y < 0 || y > map_.Rows() - 1 {
        return  Ocean
    }

    for x < 0 {
        x += map_.Columns()
    }

    x = x % map_.Columns()

    index := map_.Terrain[x][y]
    if index < 0 || index >= len(data.Tiles) {
        fmt.Printf("Error: invalid index in terrain %v at %v,%v\n", index, x, y)
        return Unknown
    }
    return data.Tiles[index].Tile.TerrainType()
}

// given a position in the terrain matrix, find a tile that fits the tile and all its neighbors
func (map_ *Map) ResolveTile(x int, y int, data *TerrainData, plane data.Plane) (int, error) {
    region := make(map[Direction]TerrainType)
    region[Center] = map_.getTerrainAt(x, y, data)
    region[West] = map_.getTerrainAt(x-1, y, data)
    region[NorthWest] = map_.getTerrainAt(x-1, y-1, data)
    region[SouthWest] = map_.getTerrainAt(x-1, y+1, data)
    region[East] = map_.getTerrainAt(x+1, y, data)
    region[North] = map_.getTerrainAt(x, y-1, data)
    region[South] = map_.getTerrainAt(x, y+1, data)
    region[NorthEast] = map_.getTerrainAt(x+1, y-1, data)
    region[SouthEast] = map_.getTerrainAt(x+1, y+1, data)

    // convert ocean tiles to lake or shores
    isLand := func(t TerrainType) bool {
        return t != Ocean && t != Shore
    }

    if region[Center] == Ocean {
        if isLand(region[West]) && isLand(region[East]) && isLand(region[North]) && isLand(region[South]) {
            region[Center] = Lake
            map_.Terrain[x][y] = TileLake.Index(plane)
        } else if isLand(region[West]) || isLand(region[East]) || isLand(region[North]) || isLand(region[South]) || isLand(region[NorthWest]) || isLand(region[NorthEast]) || isLand(region[SouthWest]) || isLand(region[SouthEast]) {
            region[Center] = Shore
            map_.Terrain[x][y] = TileShore1_00000001.Index(plane)
        }
    }

    // check if tile is already resolved
    if data.Tiles[map_.Terrain[x][y]].Tile.Matches(region) {
        return map_.Terrain[x][y], nil
    }

    // resolve tile
    tile := data.FindMatchingTile(region, plane)

    if tile == -1 {
        fmt.Printf("no matching tile for %v", region)
        return -1, fmt.Errorf("no matching tile for %v", region)
    }

    return tile, nil
}

func (map_ *Map) ResolveTiles(data *TerrainData, plane data.Plane) {
    for x := 0; x < map_.Columns(); x++ {
        for y := 0; y < map_.Rows(); y++ {
            choice, err := map_.ResolveTile(x, y, data, plane)
            if err == nil {
                map_.Terrain[x][y] = choice
            }
        }
    }
}
