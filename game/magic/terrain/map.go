package terrain

import (
    "fmt"
    "log"
    "time"
    "image"
    "math/rand/v2"
    "math"

    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

// a continent is a list of points that are indices into the Terrain matrix
type Continent = *set.Set[image.Point]

func MakeContinent() Continent {
    return set.MakeSet[image.Point]()
}

/*
func (continent Continent) Size() int {
    return continent.Size()
}

func (continent Continent) Contains(point image.Point) bool {
    for _, check := range continent.Values() {
        if check == point {
            return true
        }
    }

    return false
}
*/

type Map struct {
    Terrain [][]int
}

func (map_ *Map) Rows() int {
    return len(map_.Terrain[0])
}

func (map_ *Map) Columns() int {
    return len(map_.Terrain)
}

func (map_ *Map) WrapX(x int) int {
    for x < 0 {
        x += map_.Columns()
    }

    return x % map_.Columns()
}

func (map_ *Map) Copy() *Map {
    columns := map_.Columns()
    rows := map_.Rows()

    terrain := make([][]int, columns)
    for x := 0; x < columns; x++ {
        terrain[x] = make([]int, rows)
        for y := 0; y < rows; y++ {
            terrain[x][y] = map_.Terrain[x][y]
        }
    }

    return &Map{
        Terrain: terrain,
    }
}

func (map_ *Map) findContinent(seen map[image.Point]struct{}, x int, y int) Continent {
    rows := map_.Rows()
    columns := map_.Columns()

    continent := MakeContinent()

    // count all tiles connected to this one if those tiles are land
    searchTiles := func(startX int, startY int){
        search := []image.Point{image.Pt(startX, startY)}

        for len(search) > 0 {
            point := search[0]
            search = search[1:]
            x := point.X
            y := point.Y

            _, ok := seen[image.Pt(x, y)]
            if ok {
                continue
            }

            seen[image.Pt(x, y)] = struct{}{}

            if GetTile(map_.Terrain[x][y]).IsLand() {
                continent.Insert(image.Pt(x, y))

                for dx := -1; dx <= 1; dx++ {
                    for dy := -1; dy <= 1; dy++ {
                        if dx == 0 && dy == 0 {
                            continue
                        }

                        nx := map_.WrapX(x + dx)
                        ny := y + dy

                        if nx >= 0 && nx < columns && ny >= 0 && ny < rows {
                            search = append(search, image.Pt(nx, ny))
                        }
                    }
                }
            }
        }
    }

    searchTiles(x, y)
    return continent
}

// given a point on a continent, return all the points for that continent
func (map_ *Map) FindContinent(x int, y int) Continent {
    return map_.findContinent(make(map[image.Point]struct{}), x, y)
}

func (map_ *Map) FindContinents() []Continent {
    seen := make(map[image.Point]struct{})

    var continents []Continent

    for x := 0; x < map_.Columns(); x++ {
        for y := 0; y < map_.Rows(); y++ {
            continent := map_.findContinent(seen, x, y)
            if continent.Size() > 0 {
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

func (map_ *Map) generateLandCellularAutomata(plane data.Plane){
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

// put down other tiles like forests, mountains, special nodes, etc
func (map_ *Map) placeRandomTerrainTiles(plane data.Plane, continents []Continent) {

    // continents := map_.FindContinents()

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

        points := continent.Values()
        for i := 0; i < continent.Size() * 2; i++ {
            point := chooseRandomElement(points)

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

        magicTiles := []int{TileSorceryLake.Index(plane), TileNatureForest.Index(plane), TileChaosVolcano.Index(plane)}

        magicNodeOk := func(point image.Point) bool {
            size := 3
            for dx := -size; dx <= size; dx++ {
                for dy := -size; dy <= size; dy++ {
                    cx := map_.WrapX(point.X + dx)
                    cy := point.Y + dy

                    if cy < 0 || cy >= map_.Rows() {
                        continue
                    }

                    if GetTile(map_.Terrain[cx][cy]).IsMagic() {
                        return false
                    }
                }
            }

            return true
        }

        maxNodes := int(math.Sqrt(float64(continent.Size()))) / 7
        for _, index := range rand.Perm(len(points)) {
            // fmt.Printf("Index: %v, maxNodes: %v\n", index, maxNodes)
            if maxNodes <= 0 {
                break
            }

            point := points[index]
            if magicNodeOk(point) {
                map_.Terrain[point.X][point.Y] = chooseRandomElement(magicTiles)
                maxNodes -= 1
            }
        }
    }
}

func (map_ *Map) placeRivers(area int, data *TerrainData, plane data.Plane, continents []Continent) {
    // continents := map_.FindContinents()

    getSides := func(point image.Point) (*set.Set[image.Point], *set.Set[TerrainType]) {
        // get points and terrains on cardinal sides of a point
        points := set.MakeSet[image.Point]()
        terrains := set.MakeSet[TerrainType]()

        points.Insert(image.Pt(map_.WrapX(point.X-1), point.Y))
        terrains.Insert(map_.getTerrainAt(map_.WrapX(point.X-1), point.Y, data))

        points.Insert(image.Pt(map_.WrapX(point.X+1), point.Y))
        terrains.Insert(map_.getTerrainAt(map_.WrapX(point.X+1), point.Y, data))

        if point.Y > 0 {
            points.Insert(image.Pt(point.X, point.Y-1))
            terrains.Insert(map_.getTerrainAt(point.X, point.Y-1, data))
        }
        if point.Y < map_.Rows() {
            points.Insert(image.Pt(point.X, point.Y+1))
            terrains.Insert(map_.getTerrainAt(point.X, point.Y+1, data))
        }
        return points, terrains
    }

    isFinished := func(point image.Point) bool {
        // check if the point is adjacent to shore, ocean or lake
        _, sides := getSides(point)
        return sides.Contains(Ocean) || sides.Contains(Shore) || sides.Contains(Lake)
    }

    isValid := func(point image.Point, path []image.Point) bool {
        // check if th given point is valid

        // outside map
        if point.Y < 0 || point.Y > map_.Rows() - 1 {
            return false
        }

        // next to path (to prevent loops)
        sidePoints, sideTerrains := getSides(point)
        for _, current := range path[:len(path)-1] {
            if sidePoints.Contains(current) {
                return false
            }
        }

        // next to existing river (to prevent loops)
        if sideTerrains.Contains(River) {
            return false
        }

        return true
    }

    // keep one in-memory copy of the original map
    mapCopy := map_.Copy()

    resolves := func(path []image.Point) bool {
        // check if the rendered path would resolve
        checked := make(map[image.Point]bool)

        // work with a copy where the path is rendered to allow resolving and discarding it
        // reset the copy here back to the original map
        for x := range mapCopy.Columns() {
            copy(mapCopy.Terrain[x], map_.Terrain[x])
        }

        for _, point := range path {
            mapCopy.Terrain[point.X][point.Y] = TileRiver0001.Index(plane)
        }

        for _, point := range path {
            if !checked[point] {
                _, err := mapCopy.ResolveTile(point.X, point.Y, data, plane)
                if err != nil {
                    return false
                }
                checked[point] = true
            }

            sides, _ := getSides(point)
            for _, side := range sides.Values() {
                if !checked[side] {
                    _, err := mapCopy.ResolveTile(side.X, side.Y, data, plane)
                    if err != nil {
                        return false
                    }
                    checked[side] = true
                }
            }
        }
        return true
    }

    walk := func(start image.Point) (bool, []image.Point) {
        // perform a random walk
        path := []image.Point{start}
        current := image.Pt(start.X, start.Y)

        if !isValid(current, path) {
            return false, []image.Point{}
        }

        finished := isFinished(current)
        visited := make(map[image.Point]bool)
        step, maxSteps := 0, 100
        for !finished {
            // drop out if not successful after 100 steps
            step += 1
            if step >= maxSteps {
                return false, []image.Point{}
            }

            // find valid next point
            next := image.Pt(current.X, current.Y)
            switch rand.IntN(4) {
                case 0: next.Y++
                case 1: next.Y--
                case 2: next.X = map_.WrapX(next.X - 1)
                case 3: next.X = map_.WrapX(next.X + 1)
            }
            if !isValid(next, path) {
                continue
            }
            if visited[next] {
                continue
            }

            // update
            visited[next] = true
            current = next
            path = append(path, current)
            finished = isFinished(current)
        }

        // disallow 1-tile rivers
        if len(path) <= 1 {
            return false, []image.Point{}
        }

        // check if tiles around river would resolve
        if !resolves(path) {
            return false, []image.Point{}
        }

        return true, path
    }

    for _, continent := range continents {
        points := continent.Values()
        for i := 0; i < continent.Size() / area; i++ {
            point := chooseRandomElement(points)
            successful, path := walk(point)
            if successful {
                for _, point := range path {
                    map_.Terrain[point.X][point.Y] = TileRiver0001.Index(plane)
                }
            }
        }
    }
}

// remove land masses that contain less squares than 'area'
func (map_ *Map) removeSmallIslands(area int, plane data.Plane){
    continents := map_.FindContinents()

    for _, continent := range continents {
        if continent.Size() < area {
            points := continent.Values()
            for _, point := range points {
                map_.Terrain[point.X][point.Y] = TileOcean.Index(plane)
            }
        }
    }
}

func (map_ *Map) getTerrainAt(x int, y int, data *TerrainData) TerrainType {
    if y < 0 || y > map_.Rows() - 1 {
        return Ocean
    }

    x = map_.WrapX(x)

    index := map_.Terrain[x][y]
    if index < 0 || index >= len(data.Tiles) {
        fmt.Printf("Error: invalid index in terrain %v at %v,%v\n", index, x, y)
        return Unknown
    }
    return data.Tiles[index].Tile.TerrainType()
}

func (map_ *Map) ResolveTile(x int, y int, data *TerrainData, plane data.Plane) (int, error) {
    return map_.resolveTile(x, y, data, plane, make(map[Direction]TerrainType))
}

// given a position in the terrain matrix, find a tile that fits the tile and all its neighbors
func (map_ *Map) resolveTile(x int, y int, data *TerrainData, plane data.Plane, region map[Direction]TerrainType) (int, error) {
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
    if data.Tiles[map_.Terrain[x][y]].Tile.matches(region) {
        return map_.Terrain[x][y], nil
    }

    // resolve tile
    tile := data.FindMatchingTile(region, plane)

    if tile == -1 {
        return -1, fmt.Errorf("no matching tile for %v", region)
    }

    return tile, nil
}

func (map_ *Map) ResolveTiles(data *TerrainData, plane data.Plane) {
    region := make(map[Direction]TerrainType)
    for x := 0; x < map_.Columns(); x++ {
        for y := 0; y < map_.Rows(); y++ {
            choice, err := map_.resolveTile(x, y, data, plane, region)
            if err == nil {
                map_.Terrain[x][y] = choice
            }
        }
    }
}

func (map_ *Map) resolveTileWithNeighbors(x int, y int, data *TerrainData, plane data.Plane) {
    for dx := x - 1; dx <= x + 1; dx++ {
        for dy := y - 1; dy <= y + 1; dy++ {
            if dy >= 0 && dy < map_.Rows() {
                wx := map_.WrapX(dx)
                choice, err := map_.ResolveTile(wx, dy, data, plane)
                if err == nil {
                    map_.Terrain[wx][dy] = choice
                }
            }
        }
    }
}

func (map_ *Map) SetTerrainAt(x int, y int, terrainType TerrainType, data *TerrainData, plane data.Plane) {
    if y >= 0 || y < map_.Rows() {
        x = map_.WrapX(x)

        index := -1
        switch terrainType {
            case Ocean: index = TileOcean.Index(plane)
            case River: index = TileRiver0001.Index(plane)
            case Shore: index = TileShore1_00000001.Index(plane)
            case Mountain: index = TileMountain1.Index(plane)
            case Hill: index = TileHills1.Index(plane)
            case Grass: index = TileGrasslands1.Index(plane)
            case Swamp: index = TileSwamp1.Index(plane)
            case Forest: index = TileForest1.Index(plane)
            case Desert: index = TileAllDesert1.Index(plane)
            case Tundra: index = TileTundra.Index(plane)
            case Volcano: index = TileVolcano.Index(plane)
            case Lake: index = TileLake.Index(plane)
            case NatureNode: index = TileNatureForest.Index(plane)
            case SorceryNode: index = TileSorceryLake.Index(plane)
            case ChaosNode: index = TileChaosVolcano.Index(plane)
        }
        if index != -1 {
            map_.Terrain[x][y] = index
            map_.resolveTileWithNeighbors(x, y, data, plane)
        }
    }
}


func GenerateLandCellularAutomata(columns int, rows int, data *TerrainData, plane data.Plane) *Map {
    // run a cellular automata simulation for a few rounds to generate
    // land and ocean tiles. then call ResolveTiles() to clean up the edges
    start := time.Now()
    map_ := MakeMap(rows, columns)
    map_.generateLandCellularAutomata(plane)
    map_.removeSmallIslands(100, plane)
    continents := map_.FindContinents()
    map_.placeRandomTerrainTiles(plane, continents)
    map_.placeRivers(100, data, plane, continents)
    map_.ResolveTiles(data, plane)
    end := time.Now()
    log.Printf("Generated %vx%v %v map in %v", columns, rows, plane, end.Sub(start))
    return map_
}
