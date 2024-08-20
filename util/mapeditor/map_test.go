package main

// $ go test -cpuprofile cpu.out -bench=. ./util/mapeditor
// $ go tool pprof cpu.out

import (
    "testing"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
)

func BenchmarkGeneration(bench *testing.B){
    map_ := terrain.MakeMap(100, 200)

    for i := 0; i < bench.N; i++ {
        map_.GenerateLandCellularAutomata()
        map_.RemoveSmallIslands(100)
        map_.PlaceRandomTerrainTiles()
    }
}
