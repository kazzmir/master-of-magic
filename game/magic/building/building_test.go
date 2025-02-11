package building

import (
    "testing"
    "fmt"
    "time"
    "image"
    "slices"
    "math/rand/v2"
)

// a position within some patch of land where this building is located. the Area field
// implicitly contains the x,y position within the overall patch of land
type BuildingPosition struct {
    Building Building
    Area image.Rectangle
}

// represents a patch of land (in between roads) that can have buildings placed on it
type Rect struct {
    Width int
    Height int
    Id int
    Buildings []BuildingPosition
}

func (rect *Rect) Clone() *Rect {
    newRect := &Rect{Width: rect.Width, Height: rect.Height, Id: rect.Id}
    newRect.Buildings = make([]BuildingPosition, len(rect.Buildings))
    copy(newRect.Buildings, rect.Buildings)
    return newRect
}

func (rect *Rect) Remove(building Building) {
    rect.Buildings = slices.DeleteFunc(rect.Buildings, func(position BuildingPosition) bool {
        return position.Building == building
    })
}

// compute how much space is unused in this rectangle
func (rect *Rect) EmptySpace() int {
    total := rect.Area()

    for _, building := range rect.Buildings {
        total -= building.Area.Dx() * building.Area.Dy()
    }

    return total
}

// try to add the building to this patch of land. returns true if successful
// each possible point the new building could be placed is tried in a random order. if the building
// overlaps with any existing buildings then that point is skipped
func (rect *Rect) Add(building Building, width, height int, random *rand.Rand) bool {

    if rect.EmptySpace() < width * height {
        return false
    }

    check := func(use image.Rectangle) bool {
        for _, existing := range rect.Buildings {
            if existing.Area.Overlaps(use) {
                return false
            }
        }

        return true
    }

    // fmt.Printf("Add %v (%v, %v) to %v, %v\n", building, width, height, rect.Width, rect.Height)

    for _, x := range random.Perm(rect.Width - width + 1) {
        for _, y := range random.Perm(rect.Height - height + 1) {
            buildingRect := image.Rect(0, 0, width, height).Add(image.Pt(x, y))
            if check(buildingRect) {
                rect.Buildings = append(rect.Buildings, BuildingPosition{Building: building, Area: buildingRect})
                return true
            }
        }
    }

    return false
}

func (rect *Rect) Area() int {
    return rect.Width * rect.Height
}

func cloneRects(rects []*Rect) []*Rect {
    newRects := make([]*Rect, len(rects))
    for i, rect := range rects {
        newRects[i] = rect.Clone()
    }

    return newRects
}

const MAX_ITERATIONS = 2000

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

func doLayoutIterative(buildings []Building, rects []*Rect, random *rand.Rand, count *int) ([]*Rect, bool) {
    if len(buildings) == 0 {
        return rects, true
    }

    /*
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
    */

    // fmt.Printf("Trying to add %v (%v,%v)\n", building, width, height)

    type State struct {
        Buildings []Building
        Rects []*Rect
        Index int
    }

    var stack []State

    for _, i := range rand.Perm(len(rects)) {
        stack = append(stack, State{Buildings: buildings, Rects: rects, Index: i})
    }

    for len(stack) > 0 && *count < MAX_ITERATIONS {
        *count += 1

        // fmt.Printf("Stack length %v count %v\n", len(stack), *count)

        use := stack[len(stack) - 1]
        stack = stack[:len(stack) - 1]

        // fmt.Printf("Buildings left %v\n", len(use.Buildings))

        if len(use.Buildings) == 0 {
            return use.Rects, true
        }

        building := use.Buildings[0]
        width, height := building.Size()

        moreRects := cloneRects(use.Rects)

        // the order the patches of land are tried is random
        rect := moreRects[use.Index]
        // fmt.Printf("Rect %v empty space %v buildings %v\n", rect.Id, rect.EmptySpace(), len(rect.Buildings))
        if rect.Add(building, width, height, random) {
            // fmt.Printf("Added %v (%v,%v) to rect %v\n", building, width, height, rect.Id)

            for _, i := range rand.Perm(len(rects)) {
                stack = append(stack, State{Buildings: use.Buildings[1:], Rects: moreRects, Index: i})
            }

            // fmt.Printf("Removed %v (%v,%v) from rect %v empty=%v\n", building, width, height, rect.Id, rect.EmptySpace())
        }
    }

    return nil, false
}

func TestLayout2(test *testing.T){
    rects := []*Rect{&Rect{Width: 4, Height: 4, Id: 0}}
    count := 0
    solution, ok := doLayoutRecursive([]Building{BuildingArmorersGuild}, rects, rand.New(rand.NewPCG(0, 1)), &count)
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
    solution, ok := doLayoutRecursive(filterReplaced(try), rects, rand.New(rand.NewPCG(uint64(a.UnixNano()), uint64(b.UnixNano()))), &count)

    if !ok {
        fmt.Printf("No solution found\n")
    } else {
        emptySpace := 0
        for _, rect := range solution {
            emptySpace += rect.EmptySpace()
        }

        fmt.Printf("Count: %v Empty space: %v\n", count, emptySpace)
    }

    for i := range 50 {
        v1 := uint64(i) + uint64(time.Now().UnixNano())
        start := time.Now()
        count := 0
        _, ok := doLayoutRecursive(filterReplaced(try), rects, rand.New(rand.NewPCG(v1, v1 + 1)), &count)
        end := time.Now()
        if !ok {
            fmt.Printf("[%v] No solution\n", i)
        } else {
            fmt.Printf("[%v] Success in %v iterations %v\n", i, end.Sub(start), count)
        }
    }
}

func TestLayout3(test *testing.T){
    // these rows represent the sizes of the standard patches of land in a cityscape
    row1 := []*Rect{
        {Width: 3, Height: 4, Id: 0},
        {Width: 4, Height: 4, Id: 1},
        {Width: 3, Height: 4, Id: 2},
        {Width: 4, Height: 4, Id: 3},
        {Width: 1, Height: 1, Id: 4},
    }
    row2 := []*Rect{
        // {Width: 3, Height: 3, Id: 4}, this is always shipyard etc.
        {Width: 4, Height: 3, Id: 5},
        {Width: 3, Height: 3, Id: 6},
        {Width: 4, Height: 3, Id: 7},
        {Width: 4, Height: 3, Id: 7},
        {Width: 2, Height: 3, Id: 8},
    }
    row3 := []*Rect{
        {Width: 1, Height: 4, Id: 9},
        {Width: 4, Height: 4, Id: 10},
        {Width: 3, Height: 4, Id: 11},
        {Width: 4, Height: 4, Id: 12},
        {Width: 3, Height: 4, Id: 13},
    }

    rects := append(append(row1, row2...), row3...)

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
    for range 10 {
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

    tries = 0
    for range 10 {
        tries += 1
        count = 0
        solution, ok = doLayoutIterative(filterReplaced(buildings), rects, rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())), &count)
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

        fmt.Printf("Iterative: Tries %v Count: %v Empty space: %v\n", tries, count, emptySpace)
    }

    successes := 0
    total := 100
    for range total {
        count = 0
        _, ok = doLayoutIterative(filterReplaced(buildings), rects, rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())), &count)
        if ok {
            successes += 1
        }
    }

    fmt.Printf("Success rate %v/%v %v%%\n", successes, total, float64(successes) / float64(total) * 100)

    for i := range 5 {
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
