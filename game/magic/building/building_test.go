package building

import (
    "testing"
    "fmt"
    "time"
    "image"
    "slices"
    "math/rand/v2"
)

type BuildingPosition struct {
    Building Building
    Area image.Rectangle
}

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

func (rect *Rect) EmptySpace() int {
    total := rect.Area()

    for _, building := range rect.Buildings {
        total -= building.Area.Dx() * building.Area.Dy()
    }

    return total
}

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

func doLayout(buildings []Building, rects []*Rect, random *rand.Rand) ([]*Rect, bool) {
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

    clone := make([]*Rect, len(rects))
    for i, rect := range rects {
        clone[i] = rect.Clone()
    }

    for _, rect := range clone {
        // fmt.Printf("Rect %v empty space %v buildings %v\n", rect.Id, rect.EmptySpace(), len(rect.Buildings))
        if rect.Add(building, width, height, random) {
            // fmt.Printf("Added %v (%v,%v) to rect %v\n", building, width, height, rect.Id)
            solution, ok := doLayout(buildings[1:], clone, random)
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
    solution, ok := doLayout([]Building{BuildingArmorersGuild}, rects, rand.New(rand.NewPCG(0, 1)))
    if !ok {
        fmt.Printf("No solution\n")
    } else {
        fmt.Printf("Empty space: %v\n", solution[0].EmptySpace())
    }
}

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
        if !wasBuildingReplaced(building) {
            out = append(out, building)
        }
    }

    return out
}

func TestLayout(test *testing.T){
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

    solution, ok := doLayout(filterReplaced(try), rects, rand.New(rand.NewPCG(uint64(a.UnixNano()), uint64(b.UnixNano()))))

    if !ok {
        test.Errorf("No solution found\n")
    } else {
        emptySpace := 0
        for _, rect := range solution {
            emptySpace += rect.EmptySpace()
        }

        fmt.Printf("Empty space: %v\n", emptySpace)
    }

    for i := range 50 {
        v1 := uint64(i) + uint64(time.Now().UnixNano())
        start := time.Now()
        _, ok := doLayout(filterReplaced(try), rects, rand.New(rand.NewPCG(v1, v1 + 1)))
        end := time.Now()
        if !ok {
            test.Errorf("[%v] No solution\n", i)
        }
        fmt.Printf("[%v] Success in %v\n", i, end.Sub(start))
    }
}
