package building

import (
    "testing"
    "fmt"
    "time"
    "slices"
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/lib/set"
)

const MAX_ITERATIONS = 500

// recursive algorithm that tries to layout each building in some patch of land
// if a building fails to be placed, then the algorithm backtracks and tries a different rect
// for the previous building
func doLayoutRecursive(buildings []Building, rects []*Rect, random *rand.Rand, count *int) ([]*Rect, bool) {
    *count += 1

    if *count > MAX_ITERATIONS {
        return nil, false
    }

    if len(buildings) == 0 {
        return rects, true
    }

    building := buildings[0]
    // fmt.Printf("Check %v\n", building)
    width, height := building.Size()
    for width == 0 && height == 0 {
        buildings = buildings[1:]
        if len(buildings) == 0 {
            return rects, true
        }
        building = buildings[0]
        width, height = building.Size()
    }

    // fmt.Printf("Trying to add %v (%v,%v)\n", building, width, height)

    clone := cloneRects(rects)

    // the order the patches of land are tried is random
    for _, i := range rand.Perm(len(clone)) {
        rect := clone[i]
        // fmt.Printf("Rect %v empty space %v buildings %v\n", rect.Id, rect.EmptySpace(), len(rect.Buildings))
        if rect.Add(building, width, height, random) {
            // fmt.Printf("Added %v (%v,%v) to rect %v\n", building, width, height, rect.Id)
            solution, ok := doLayoutRecursive(buildings[1:], clone, random, count)
            if ok {
                return solution, true
            }

            rect.Remove(building)
            // fmt.Printf("Removed %v (%v,%v) from rect %v empty=%v\n", building, width, height, rect.Id, rect.EmptySpace())
        }
    }

    return nil, false
}

func TestLayout2(test *testing.T){
    rects := []*Rect{&Rect{Width: 4, Height: 4, Id: 0}}
    count := 0
    solution, ok := doLayoutIterative([]Building{BuildingArmorersGuild}, rects, rand.New(rand.NewPCG(0, 1)), &count, 1000)
    if !ok {
        fmt.Printf("No solution\n")
    } else {
        fmt.Printf("Empty space: %v\n", solution[0].EmptySpace())
    }
}

// remove buildings that have been replaced
func filterReplaced(buildings []Building) []Building {
    var wasBuildingReplaced func (building Building) bool
    wasBuildingReplaced = func (building Building) bool {
        if building == BuildingNone {
            return false
        }

        replacedBy := building.ReplacedBy()
        return replacedBy != BuildingNone && (slices.Contains(buildings, replacedBy) || wasBuildingReplaced(replacedBy))
    }

    var out []Building
    for _, building := range buildings {
        width, height := building.Size()
        if width == 0 && height == 0 {
            continue
        }

        if !wasBuildingReplaced(building) {
            out = append(out, building)
        }
    }

    return out
}

func TestLayout(test *testing.T){
    // these rows represent the sizes of the standard patches of land in a cityscape
    row1 := []*Rect{&Rect{Width: 3, Height: 4, Id: 0}, &Rect{Width: 4, Height: 4, Id: 1}, &Rect{Width: 3, Height: 4, Id: 2}, &Rect{Width: 4, Height: 4, Id: 3}}
    row2 := []*Rect{&Rect{Width: 3, Height: 3, Id: 4}, &Rect{Width: 4, Height: 3, Id: 5}, &Rect{Width: 3, Height: 3, Id: 6}, &Rect{Width: 4, Height: 3, Id: 7}}
    row3 := []*Rect{&Rect{Width: 1, Height: 4, Id: 8}, &Rect{Width: 4, Height: 4, Id: 9}, &Rect{Width: 3, Height: 4, Id: 10}, &Rect{Width: 4, Height: 4, Id: 11}, &Rect{Width: 3, Height: 4, Id: 12}}

    rects := append(append(row1, row2...), row3...)

    totalSpace := 0
    for _, rect := range rects {
        totalSpace += rect.Area()
    }

    fmt.Printf("Total space: %d\n", totalSpace)

    totalBuildings := 0
    for _, building := range Buildings() {
        width, height := building.Size()
        totalBuildings += width * height
    }

    fmt.Printf("Total building space: %d\n", totalBuildings)

    try := Buildings()

    /*
    try := []Building{
        BuildingBarracks,
        BuildingArmory,
        BuildingFightersGuild,
        BuildingArmorersGuild,
        BuildingWarCollege,
        BuildingSmithy,

        BuildingStables,
        BuildingAnimistsGuild,
        BuildingFantasticStable,
        BuildingShipwrightsGuild,
        BuildingShipYard,
        BuildingMaritimeGuild,
        BuildingSawmill,
        BuildingLibrary,
        BuildingSagesGuild,
        BuildingOracle,
        BuildingAlchemistsGuild,
        BuildingUniversity,
        BuildingWizardsGuild,
        BuildingShrine,
        BuildingTemple,
        BuildingParthenon,
        BuildingCathedral,
        BuildingMarketplace,
        BuildingBank,
        BuildingMerchantsGuild,
        BuildingGranary,
        BuildingFarmersMarket,
        BuildingForestersGuild,
        BuildingBuildersHall,
        BuildingMechaniciansGuild,
        BuildingMinersGuild,
        BuildingCityWalls,
    }
    */

    a := time.Now()
    time.Sleep(1 * time.Millisecond)
    b := time.Now()

    fmt.Printf("Layout %v buildings\n", len(filterReplaced(try)))

    count := 0
    solution, ok := doLayoutIterative(filterReplaced(try), rects, rand.New(rand.NewPCG(uint64(a.UnixNano()), uint64(b.UnixNano()))), &count, 1000)

    if !ok {
        fmt.Printf("No solution found\n")
    } else {
        emptySpace := 0
        for _, rect := range solution {
            emptySpace += rect.EmptySpace()
        }

        fmt.Printf("Count: %v Empty space: %v\n", count, emptySpace)
    }

    for i := range 5 {
        v1 := uint64(i) + uint64(time.Now().UnixNano())
        start := time.Now()
        count := 0
        _, ok := doLayoutIterative(filterReplaced(try), rects, rand.New(rand.NewPCG(v1, v1 + 1)), &count, 1000)
        end := time.Now()
        if !ok {
            fmt.Printf("[%v] No solution\n", i)
        } else {
            fmt.Printf("[%v] Success in %v iterations %v\n", i, end.Sub(start), count)
        }
    }
}

func TestLayout3(test *testing.T){
    rects := StandardRects()

    totalSpace := 0
    for _, rect := range rects {
        totalSpace += rect.Area()
    }

    fmt.Printf("Total space: %d\n", totalSpace)

    buildings := []Building{
        BuildingBarracks,
        BuildingArmory,
        BuildingFightersGuild,
        BuildingArmorersGuild,
        BuildingWarCollege,
        BuildingSmithy,
        BuildingStables,
        BuildingAnimistsGuild,
        BuildingFantasticStable,
        // BuildingShipwrightsGuild,
        // BuildingShipYard,
        // BuildingMaritimeGuild,
        BuildingSawmill,
        BuildingLibrary,
        BuildingSagesGuild,
        BuildingOracle,
        BuildingAlchemistsGuild,
        BuildingUniversity,
        BuildingWizardsGuild,
        BuildingShrine,
        BuildingTemple,
        BuildingParthenon,
        BuildingCathedral,
        BuildingMarketplace,
        BuildingBank,
        BuildingMerchantsGuild,
        BuildingGranary,
        BuildingFarmersMarket,
        BuildingForestersGuild,
        BuildingBuildersHall,
        BuildingMechaniciansGuild,
        BuildingMinersGuild,
        BuildingCityWalls,
        BuildingFortress,
        BuildingSummoningCircle,
        BuildingAltarOfBattle,
        BuildingAstralGate,
        BuildingStreamOfLife,
        BuildingEarthGate,
        // BuildingDarkRituals, // cannot mix death and life
    }

    totalBuildings := 0
    for _, building := range buildings {
        width, height := building.Size()
        totalBuildings += width * height
    }

    fmt.Printf("Total building space: %d\n", totalBuildings)

    fmt.Printf("Layout %v buildings\n", len(filterReplaced(buildings)))

    count := 0
    var solution []*Rect
    var ok bool
    tries := 0
    for range 50 {
        tries += 1
        count = 0
        solution, ok = doLayoutRecursive(filterReplaced(buildings), rects, rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())), &count)
        if ok {
            break
        }
    }

    if !ok {
        test.Errorf("No solution found in 10 tries\n")
    } else {
        emptySpace := 0
        for _, rect := range solution {
            emptySpace += rect.EmptySpace()
        }

        fmt.Printf("Recursive: Tries %v Count: %v Empty space: %v\n", tries, count, emptySpace)
    }

    maxIterations := 500

    start := time.Now()
    tries = 0
    for range 20 {
        tries += 1
        count = 0
        solution, ok = doLayoutIterative(filterReplaced(buildings), rects, rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())), &count, maxIterations)
        if ok {
            break
        }
    }
    end := time.Now()

    if !ok {
        test.Errorf("No solution found in 10 tries\n")
    } else {
        emptySpace := 0
        for _, rect := range solution {
            emptySpace += rect.EmptySpace()
        }

        fmt.Printf("Iterative: Tries %v Count: %v Empty space: %v. Took %v\n", tries, count, emptySpace, end.Sub(start))

        var foundBuildings []Building
        for _, rect := range solution {
            for _, building := range rect.Buildings {
                if slices.Contains(foundBuildings, building.Building) {
                    fmt.Printf("Duplicate building %v!\n", building.Building)
                    break
                }
                foundBuildings = append(foundBuildings, building.Building)
            }
        }

        if len(foundBuildings) != len(filterReplaced(buildings)) {
            fmt.Printf("Not all buildings placed!\n")
        }

    }

    successes := 0
    total := 50
    for range total {
        count = 0
        solution, ok = doLayoutIterative(filterReplaced(buildings), rects, rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())), &count, maxIterations)
        if ok {
            successes += 1

            var foundBuildings []Building
            for _, rect := range solution {
                for _, building := range rect.Buildings {
                    if slices.Contains(foundBuildings, building.Building) {
                        fmt.Printf("Duplicate building %v!\n", building.Building)
                        break
                    }
                    foundBuildings = append(foundBuildings, building.Building)
                }
            }

            checkBuildings := filterReplaced(buildings)

            if len(foundBuildings) != len(checkBuildings) {
                fmt.Printf("Not all buildings placed! %v vs %v\n", len(foundBuildings), len(filterReplaced(buildings)))

                for _, building := range checkBuildings {
                    if !slices.Contains(foundBuildings, building) {
                        fmt.Printf("Missing building %v (fortress %v)\n", building, BuildingFortress)
                    }
                }

            }
        }
    }

    fmt.Printf("Success rate %v/%v %.2f%%\n", successes, total, float64(successes) / float64(total) * 100)

    for i := range 10 {
        v1 := uint64(i) + uint64(time.Now().UnixNano())
        start := time.Now()
        count := 0
        _, ok := doLayoutRecursive(filterReplaced(buildings), rects, rand.New(rand.NewPCG(v1, v1 + 1)), &count)
        end := time.Now()
        if !ok {
            fmt.Printf("[%v] No solution\n", i)
        } else {
            fmt.Printf("[%v] Success in %v iterations %v\n", i, end.Sub(start), count)
        }
    }
}

// laying out the same buildings twice with the same initial random should produce the same layout
func TestLayoutEquality(test *testing.T){
    rects := StandardRects()

    buildings := set.NewSet(
        BuildingBarracks,
        BuildingArmory,
        BuildingFightersGuild,
        BuildingArmorersGuild,
        BuildingWarCollege,
        BuildingSmithy,
        BuildingStables,
        BuildingAnimistsGuild,
        BuildingFantasticStable,
        BuildingSawmill,
        BuildingCathedral,
        BuildingMarketplace,
        BuildingBank,
        BuildingMerchantsGuild,
        BuildingGranary,
        BuildingFarmersMarket,
        BuildingForestersGuild,
        BuildingBuildersHall,
        BuildingMechaniciansGuild,
        BuildingMinersGuild,
        BuildingCityWalls,
        BuildingFortress,
        BuildingSummoningCircle,
        BuildingAltarOfBattle,
        BuildingAstralGate,
        BuildingStreamOfLife,
        BuildingEarthGate,
    )

    random := rand.New(rand.NewPCG(0, 1))

    layout, ok := LayoutBuildings(filterReplaced(buildings.Values()), rects, random)
    if !ok {
        test.Errorf("No solution found\n")
    }

    if len(layout) == 0 {
        test.Errorf("No layout found\n")
    }

    // do layout check 5 times
    for range 5 {
        random = rand.New(rand.NewPCG(0, 1))
        layout2, ok := LayoutBuildings(filterReplaced(buildings.Values()), rects, random)
        if !ok {
            test.Errorf("No solution found\n")
            break
        }

        if len(layout2) == 0 {
            test.Errorf("No second layout found\n")
            break
        }

        for _, rect1 := range layout {
            found := false
            for _, rect2 := range layout2 {
                if rect1.Equals(rect2) {
                    found = true
                    break
                }
            }

            if !found {
                test.Errorf("Rect %v not found in second layout\n", rect1.Id)
                break
            }
        }
    }
}
