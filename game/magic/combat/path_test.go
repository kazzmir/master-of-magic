package combat

import (
    "testing"
    "strings"
    "image"
    "slices"
)

func makeMap(data string) [][]int {
    lines := strings.Split(data, "\n")

    var out [][]int

    for _, line := range lines {
        var row []int

        for _, c := range line {
            if c >= '0' && c <= '9' {
                row = append(row, int(c - '0'))
            }
        }

        if len(row) > 0 {
            out = append(out, row)
        }
    }

    return out
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

    tileCost := func(x int, y int) float64 {
        if x < 0 || y < 0 || y >= len(tiles) || x >= len(tiles[y]) {
            return Infinity
        }

        return float64(tiles[y][x])
    }

    neighbors := func(cx int, cy int) []image.Point {
        var out []image.Point

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

    path, ok := FindPath(start, end, 10, tileCost, neighbors)
    if !ok {
        test.Errorf("unable to find path")
    }

    equalPoints := func (a image.Point, b image.Point) bool {
        return a.Eq(b)
    }

    if !slices.EqualFunc(path, expectedPath, equalPoints) {
        test.Errorf("path not as expected: expected=%v actual=%v", expectedPath, path)
    }

    path2, ok := FindPath(start, end, 3, tileCost, neighbors)
    if ok {
        test.Errorf("expected no path, but found one: %v", path2)
    }
}
