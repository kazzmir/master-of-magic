package pathfinding

import (
    "testing"
    "strings"
    "image"
    "slices"
    "math/rand/v2"
)

func makeMap(data string) [][]float64 {
    lines := strings.Split(data, "\n")

    var out [][]float64

    for _, line := range lines {
        var row []float64

        for _, c := range line {
            if c >= '0' && c <= '9' {
                row = append(row, float64(c - '0'))
            } else if c == 'X' {
                row = append(row, Infinity)
            }
        }

        if len(row) > 0 {
            out = append(out, row)
        }
    }

    return out
}

func makeTileCost(tiles [][]float64) TileCostFunc {
    // just consider cost of new tile
    return func(x1 int, y1 int, x2 int, y2 int) float64 {
        if x2 < 0 || y2 < 0 || y2 >= len(tiles) || x2 >= len(tiles[y2]) {
            return Infinity
        }

        return tiles[y2][x2]
    }
}

func makeNeighbors(tiles [][]float64) NeighborsFunc {
    return func(cx int, cy int) []image.Point {
        // var out []image.Point
        out := make([]image.Point, 0, 8)

        for x := -1; x <= 1; x++ {
            for y := -1; y <= 1; y++ {
                dx := cx + x
                dy := cy + y
                if dx == cx && dy == cy {
                    continue
                }

                if dx >= 0 && dy >= 0 && dy < len(tiles) && dx < len(tiles[dy]) {
                    out = append(out, image.Pt(dx, dy))
                }
            }
        }

        return out
    }
}

func TestPath1(test *testing.T){

    tiles := makeMap(`
1223
8123
2153
2111
`)

    start := image.Pt(0, 0)
    end := image.Pt(3, 3)

    expectedPath := []image.Point{
        image.Pt(0, 0),
        image.Pt(1, 1),
        image.Pt(1, 2),
        image.Pt(2, 3),
        image.Pt(3, 3),
    }

    tileCost := makeTileCost(tiles)
    neighbors := makeNeighbors(tiles)

    path, ok := FindPath(start, end, 10, tileCost, neighbors, PointEqual)
    if !ok {
        test.Errorf("unable to find path")
    }

    equalPoints := func (a image.Point, b image.Point) bool {
        return a.Eq(b)
    }

    if !slices.EqualFunc(path, expectedPath, equalPoints) {
        test.Errorf("path not as expected: expected=%v actual=%v", expectedPath, path)
    }

    path2, ok := FindPath(start, end, 3, tileCost, neighbors, PointEqual)
    if ok {
        test.Errorf("expected no path, but found one: %v", path2)
    }
}

/* can't move through walls (X) */
func TestPathObstacles(test *testing.T){
    tiles := makeMap(`
1123
XXX3
213X
2111
`)

    start := image.Pt(0, 0)
    end := image.Pt(3, 3)

    expectedPath := []image.Point{
        image.Pt(0, 0),
        image.Pt(1, 0),
        image.Pt(2, 0),
        image.Pt(3, 1),
        image.Pt(2, 2),
        image.Pt(3, 3),
    }

    tileCost := makeTileCost(tiles)
    neighbors := makeNeighbors(tiles)

    path, ok := FindPath(start, end, 12, tileCost, neighbors, PointEqual)
    if !ok {
        test.Errorf("unable to find path")
    }

    equalPoints := func (a image.Point, b image.Point) bool {
        return a.Eq(b)
    }

    if !slices.EqualFunc(path, expectedPath, equalPoints) {
        test.Errorf("path not as expected: expected=%v actual=%v", expectedPath, path)
    }

}

/* no path at all because there are no openings in the walls */
func TestPathFullBlocked(test *testing.T){
    tiles := makeMap(`
1123
XXXX
2132
2111
`)

    start := image.Pt(0, 0)
    end := image.Pt(3, 3)

    tileCost := makeTileCost(tiles)
    neighbors := makeNeighbors(tiles)

    _, ok := FindPath(start, end, Infinity, tileCost, neighbors, PointEqual)
    if ok {
        test.Errorf("able to find path through blocked map")
    }
}

func makeRandomMap(rows int, columns int, value int) [][]float64 {
    var out [][]float64

    for y := 0; y < rows; y++ {
        var row []float64
        for x := 0; x < columns; x++ {
            row = append(row, float64(rand.IntN(value)))
        }
        out = append(out, row)
    }

    return out
}

func TestStress(test *testing.T){
    tiles := makeRandomMap(100, 100, 10)

    start := image.Pt(0, 0)
    end := image.Pt(99, 99)

    _, ok := FindPath(start, end, 100000, makeTileCost(tiles), makeNeighbors(tiles), PointEqual)
    if !ok {
        test.Errorf("unable to find path")
    }

    for range 10 {
        start = image.Pt(rand.IntN(100), rand.IntN(100))
        end = image.Pt(rand.IntN(100), rand.IntN(100))
        if start.Eq(end) {
            continue
        }
        _, ok := FindPath(start, end, 100000, makeTileCost(tiles), makeNeighbors(tiles), PointEqual)
        if !ok {
            test.Errorf("unable to find path")
        }
    }
}

func BenchmarkLarge(bench *testing.B){
    tiles := makeRandomMap(100, 100, 10)
    tileCost := makeTileCost(tiles)
    neighbors := makeNeighbors(tiles)
    bench.ResetTimer()
    for bench.Loop() {
        FindPath(image.Pt(0, 0), image.Pt(99, 99), 100000, tileCost, neighbors, PointEqual)
    }
}

func BenchmarkSmall(bench *testing.B){
    tiles := makeRandomMap(20, 20, 10)
    tileCost := makeTileCost(tiles)
    neighbors := makeNeighbors(tiles)
    bench.ResetTimer()
    for bench.Loop() {
        FindPath(image.Pt(0, 0), image.Pt(19, 19), 100000, tileCost, neighbors, PointEqual)
    }
}
