package building

import (
    "image"
    "slices"
    randomlib "math/rand/v2"
    "cmp"
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
    X int
    Y int
    Buildings []BuildingPosition
    // true if this rect should contain the fortress
    Fortress bool
}

func (rect *Rect) Equals(other *Rect) bool {
    if rect.Width != other.Width {
        return false
    }

    if rect.Height != other.Height {
        return false
    }

    if rect.Id != other.Id {
        return false
    }

    if rect.X != other.X {
        return false
    }

    if rect.Y != other.Y {
        return false
    }

    if rect.Fortress != other.Fortress {
        return false
    }

    if len(rect.Buildings) != len(other.Buildings) {
        return false
    }

    for i := range len(rect.Buildings) {
        if rect.Buildings[i] != other.Buildings[i] {
            return false
        }
    }

    return true
}

func (rect *Rect) Clone() *Rect {
    newRect := &Rect{Width: rect.Width, Height: rect.Height, Id: rect.Id, Fortress: rect.Fortress, X: rect.X, Y: rect.Y}
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
func (rect *Rect) Add(building Building, width, height int, random *randomlib.Rand) bool {

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

// compute the positions of the buildings given a set of rectangles that the buildings could be placed into
func doLayoutIterative(buildings []Building, rects []*Rect, random *randomlib.Rand, count *int, maxIterations int) ([]*Rect, bool) {
    if len(buildings) == 0 {
        return rects, true
    }

    // fmt.Printf("Trying to add %v (%v,%v)\n", building, width, height)

    type State struct {
        Buildings []Building
        Rects []*Rect
        Index int
    }

    var buildingsUse []Building
    for _, building := range buildings {
        if building == BuildingFortress {
            rects = cloneRects(rects)
            // find the rect that should contain the fortress and add it immediately
            for _, rect := range rects {
                if rect.Fortress {
                    width, height := building.Size()
                    if !rect.Add(building, width, height, random) {
                        return nil, false
                    }
                    break
                }
            }

        } else {
            buildingsUse = append(buildingsUse, building)
        }
    }
    buildings = buildingsUse

    // sort from biggest to smallest, or by index
    buildings = slices.SortedFunc(slices.Values(buildings), func (a, b Building) int {
        aWidth, aHeight := a.Size()
        bWidth, bHeight := b.Size()
        areasDifference := (bWidth * bHeight) - (aWidth * aHeight)
        // If areas are the same, sort by index, otherwise we'll get unpredictable order for buildings with equal areas.
        if areasDifference == 0 {
            return cmp.Compare(a, b)
        }
        return areasDifference
    })

    var stack []State

    for _, i := range random.Perm(len(rects)) {
        stack = append(stack, State{Buildings: buildings, Rects: rects, Index: i})
    }

    for len(stack) > 0 && *count < maxIterations {
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

            for _, i := range random.Perm(len(rects)) {
                stack = append(stack, State{Buildings: use.Buildings[1:], Rects: moreRects, Index: i})
            }

            // fmt.Printf("Removed %v (%v,%v) from rect %v empty=%v\n", building, width, height, rect.Id, rect.EmptySpace())
        }
    }

    return nil, false
}

func LayoutBuildings(buildings []Building, rects []*Rect, random *randomlib.Rand) ([]*Rect, bool) {
    return doLayoutIterative(buildings, rects, random, new(int), 500)
}

func StandardRects() []*Rect {
    // these rows represent the sizes of the standard patches of land in a cityscape
    row1 := []*Rect{
        {Width: 3, Height: 4, Id: 0, X: 0, Y: 0},
        {Width: 4, Height: 4, Id: 1, X: 1, Y: 0},
        {Width: 3, Height: 4, Id: 2, X: 2, Y: 0},
        {Width: 4, Height: 4, Id: 3, X: 3, Y: 0},
        // FIXME: this might be 2x2
        {Width: 1, Height: 1, Id: 4, X: 4, Y: 0},
    }
    row2 := []*Rect{
        // {Width: 3, Height: 3, Id: 4}, this is always shipyard etc.
        {Width: 4, Height: 3, Id: 5, X: 1, Y: 1},
        {Width: 3, Height: 3, Id: 6, Fortress: true, X: 2, Y: 1},
        {Width: 4, Height: 3, Id: 7, X: 3, Y: 1},
        {Width: 2, Height: 3, Id: 8, X: 4, Y: 1},
    }
    row3 := []*Rect{
        {Width: 1, Height: 4, Id: 9, X: 0, Y: 2},
        {Width: 4, Height: 4, Id: 10, X: 1, Y: 2},
        {Width: 3, Height: 4, Id: 11, X: 2, Y: 2},
        {Width: 4, Height: 4, Id: 12, X: 3, Y: 2},
        {Width: 4, Height: 4, Id: 13, X: 4, Y: 2},
    }

    return append(append(row1, row2...), row3...)
}
