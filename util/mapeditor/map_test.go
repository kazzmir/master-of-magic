package main

// $ go test -cpuprofile cpu.out -bench=. ./util/mapeditor
// $ go tool pprof cpu.out

import (
    "testing"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

func BenchmarkGeneration(bench *testing.B){
    map_ := terrain.MakeMap(100, 200)

    plane := data.PlaneArcanus

    for i := 0; i < bench.N; i++ {
        map_.GenerateLandCellularAutomata(plane)
        map_.RemoveSmallIslands(100, plane)
        map_.PlaceRandomTerrainTiles(plane)
    }
}
