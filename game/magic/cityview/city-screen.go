package cityview

import (
    "log"
    "fmt"
    "math"
    "math/rand/v2"
    "cmp"
    "slices"
    "image"
    "image/color"
    "hash/fnv"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/unitview"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    helplib "github.com/kazzmir/master-of-magic/game/magic/help"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"

    "github.com/hajimehoshi/ebiten/v2"
    // "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/colorm"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

const (
    // not a real building, just something that shows up in the city view screen
    BuildingTree1 buildinglib.Building = iota + buildinglib.BuildingLast
    BuildingTree2
    BuildingTree3

    BuildingTreeHouse1
    BuildingTreeHouse2
    BuildingTreeHouse3
    BuildingTreeHouse4
    BuildingTreeHouse5

    BuildingNormalHouse1
    BuildingNormalHouse2
    BuildingNormalHouse3
    BuildingNormalHouse4
    BuildingNormalHouse5

    BuildingHutHouse1
    BuildingHutHouse2
    BuildingHutHouse3
    BuildingHutHouse4
    BuildingHutHouse5
)

// buildings can appear in certain well-defined places around the city
func buildingSlots() []image.Point {

    return []image.Point{
        // row 1
        image.Pt(30, 23),
        image.Pt(70, 23),
        image.Pt(110, 23),
        image.Pt(150, 23),

        // row 2
        image.Pt(50, 43),
        image.Pt(94, 43),
        image.Pt(135, 43),

        // row 3
        image.Pt(35, 64),
        image.Pt(75, 64),
        image.Pt(115, 64),

        /*
        image.Pt(92, -4),
        image.Pt(129, 6),
        image.Pt(38, 25),
        image.Pt(10, 10),
        */
    }
}

func randomSlots(random *rand.Rand) []image.Point {
    slots := buildingSlots()
    random.Shuffle(len(slots), func(i, j int) {
        slots[i], slots[j] = slots[j], slots[i]
    })
    return slots
}

type BuildingSlot struct {
    Building buildinglib.Building
    IsRubble bool // in case of rubble
    RubbleIndex int
    Point image.Point
}

// represents how much of a resource is being used/produced, such as '2 granary' for 'the granary produces 2 food'
type ResourceUsage struct {
    Count int // can be negative
    Name string
    Replaced bool // true if the building has been replaced by another
}

type CityScreenState int

const (
    CityScreenStateRunning CityScreenState = iota
    CityScreenStateDone
)

type CityScreen struct {
    LbxCache *lbx.LbxCache
    ImageCache util.ImageCache
    Fonts *fontslib.CityViewFonts
    City *citylib.City

    UI *uilib.UI
    Player *playerlib.Player

    Buildings []BuildingSlot
    BuildScreen *BuildScreen

    // the building that was just built
    // NewBuilding buildinglib.Building

    Counter uint64
    State CityScreenState
}

func sortBuildings(buildings []buildinglib.Building) []buildinglib.Building {
    slices.SortFunc(buildings, func(a buildinglib.Building, b buildinglib.Building) int {
        return cmp.Compare(a, b)
    })
    return buildings
}

func hash(str string) uint64 {
    hasher := fnv.New64a()
    hasher.Write([]byte(str))
    return hasher.Sum64()
}

func makeRectComputePoint(rect *buildinglib.Rect) func(x int, y int) image.Point {
    startX := 0
    startY := 0

    row0Y := 25
    row1Y := 42
    row2Y := 65

    // the following equations almost work:
    // row1: startX = 44.9 * rect.X + 17.6
    // row2: startX = 44.7 * rect.X + 1.0
    // row3: startX = 38.4 * rect.X + -3.39

    switch {
        case rect.X == 0 && rect.Y == 0:
            startX = 15
            startY = row0Y
        case rect.X == 1 && rect.Y == 0:
            startX = 64
            startY = row0Y
        case rect.X == 2 && rect.Y == 0:
            startX = 112
            startY = row0Y
        case rect.X == 3 && rect.Y == 0:
            startX = 149
            startY = row0Y
        case rect.X == 4 && rect.Y == 0:
            startX = 197
            startY = row0Y

        case rect.X == 0 && rect.Y == 1:
            startX = 0
            startY = row1Y
        case rect.X == 1 && rect.Y == 1:
            startX = 45
            startY = row1Y
        case rect.X == 2 && rect.Y == 1:
            startX = 95
            startY = row1Y
        case rect.X == 3 && rect.Y == 1:
            startX = 132
            startY = row1Y
        case rect.X == 4 && rect.Y == 1:
            startX = 180
            startY = row1Y

        case rect.X == 0 && rect.Y == 2:
            startX = 8
            startY = row2Y
        case rect.X == 1 && rect.Y == 2:
            startX = 23
            startY = row2Y
        case rect.X == 2 && rect.Y == 2:
            startX = 70
            startY = row2Y
        case rect.X == 3 && rect.Y == 2:
            startX = 109
            startY = row2Y
        case rect.X == 4 && rect.Y == 2:
            startX = 157
            startY = row2Y

    }

    // the first row and column has to be pushed closer to the right side
    offsetX := 0
    if rect.X == 0 && rect.Y == 0 {
        offsetX += 1
    }

    computePoint := func (x int, y int) image.Point {
        x += offsetX
        return image.Pt(startX + (x*2+y) * 5, startY - y * 5)
    }

    return computePoint
}

// for testing the positions of the points in the city view
func makeBuildingSlots2(city *citylib.City) []BuildingSlot {
    rects := buildinglib.StandardRects()

    var slots []BuildingSlot

    for _, rect := range rects {
        computePoint := makeRectComputePoint(rect)

        /*
        if rect.Y != 2 {
            continue
        }
        */

        /*
        var geom ebiten.GeoM
        geom.Rotate(float64(-95 / 180.0) * math.Pi)
        geom.Scale(2, 2)
        geom.Translate(10, 5)
        geom.Translate(float64(rect.X * 10), float64(rect.Y * 8))
        geom.Scale(3, 3)

        computePoint := func(x int, y int) image.Point {
            ax, ay := geom.Apply(float64(x), float64(y))
            return image.Pt(int(ax), int(ay))
            // return image.Pt(rect.X * 20 + x * 5, rect.Y * 20 + y * 5)
        }
        */

        for x := range rect.Width {
            for y := range rect.Height {
                slots = append(slots, BuildingSlot{Building: buildinglib.BuildingShrine, Point: computePoint(x, y)})
            }
        }

        // slots = append(slots, BuildingSlot{Building: buildinglib.BuildingShrine, Point: computePoint(0 + offsetX, 0)})

        // distance to next plot is based on the current plot's width
    }

    return slots
}

func makeBuildingSlots(city *citylib.City) []BuildingSlot {
    // use a random seed based on the position and name of the city so that each game gets
    // a different city view, but within the same game the city view is consistent
    random := rand.New(rand.NewPCG(uint64(city.X), uint64(city.Y) + hash(city.Name)))

    toLayout := city.Buildings.Clone()

    for _, building := range toLayout.Values() {
        width, height := building.Size()
        if width == 0 || height == 0 || wasBuildingReplaced(building, city) {
            toLayout.Remove(building)
        }
    }

    enchantmentBuildings := buildinglib.EnchantmentBuildings()

    for enchantment, building := range enchantmentBuildings {
        if city.HasEnchantment(enchantment) {
            toLayout.Insert(building)
        }
    }

    toLayout.RemoveMany(buildinglib.BuildingCityWalls, buildinglib.BuildingShipwrightsGuild, buildinglib.BuildingShipYard, buildinglib.BuildingMaritimeGuild)

    var result []*buildinglib.Rect
    ok := false
    // start := time.Now()
    for range 40 {
        result, ok = buildinglib.LayoutBuildings(toLayout.Values(), buildinglib.StandardRects(), random)
        if ok {
            break
        }
    }
    // end := time.Now()

    if !ok {
        log.Printf("Warning: could not layout buildings")
        // FIXME: use some random layout?
        return nil
    }

    // log.Printf("Building layout took %v", end.Sub(start))

    getHouseTypes := func() []buildinglib.Building {
        trees := []buildinglib.Building{BuildingTreeHouse1, BuildingTreeHouse2, BuildingTreeHouse3, BuildingTreeHouse4, BuildingTreeHouse5}
        huts := []buildinglib.Building{BuildingHutHouse1, BuildingHutHouse2, BuildingHutHouse3, BuildingHutHouse4, BuildingHutHouse5}
        normal := []buildinglib.Building{BuildingNormalHouse1, BuildingNormalHouse2, BuildingNormalHouse3, BuildingNormalHouse4, BuildingNormalHouse5}

        switch city.Race.HouseType() {
            case data.HouseTypeTree: return trees
            case data.HouseTypeHut: return huts
            case data.HouseTypeNormal: return normal
        }

        return normal
    }

    houseTypes := getHouseTypes()
    treeTypes := []buildinglib.Building{BuildingTree1, BuildingTree2, BuildingTree3}

    var slots []BuildingSlot
    for _, rect := range result {
        computePoint := makeRectComputePoint(rect)

        tiles := set.MakeSet[image.Point]()
        for x := range rect.Width {
            for y := range rect.Height {
                tiles.Insert(image.Pt(x, y))
            }
        }

        for _, building := range rect.Buildings {
            point := computePoint(building.Area.Min.X, building.Area.Min.Y)
            slots = append(slots, BuildingSlot{Building: building.Building, Point: point})

            for x := range building.Area.Dx() {
                for y := range building.Area.Dy() {
                    tiles.Remove(image.Pt(x, y))
                }
            }
        }

        // fill in unused tiles with trees and houses
        for _, tile := range tiles.Values() {
            switch random.IntN(4) {
                case 0:
                    tree := treeTypes[random.IntN(len(treeTypes))]
                    slots = append(slots, BuildingSlot{Building: tree, Point: computePoint(tile.X, tile.Y)})
                case 1:
                    house := houseTypes[random.IntN(len(houseTypes))]
                    slots = append(slots, BuildingSlot{Building: house, Point: computePoint(tile.X, tile.Y)})
            }
        }
    }

    if city.Buildings.Contains(buildinglib.BuildingCityWalls) {
        slots = append(slots, BuildingSlot{Building: buildinglib.BuildingCityWalls, Point: image.Pt(0, 75)})
    }

    // water buildings always go in the same place
    for _, waterBuilding := range []buildinglib.Building{buildinglib.BuildingShipwrightsGuild, buildinglib.BuildingShipYard, buildinglib.BuildingMaritimeGuild} {
        if city.Buildings.Contains(waterBuilding) && !wasBuildingReplaced(waterBuilding, city) {
            slots = append(slots, BuildingSlot{Building: waterBuilding, Point: image.Pt(12, 45)})
        }
    }

    slices.SortFunc(slots, func(a BuildingSlot, b BuildingSlot) int {
        if a.Point.Y < b.Point.Y {
            return -1
        }

        if a.Point.Y == b.Point.Y {
            if a.Point.X < b.Point.X {
                return -1
            }

            if a.Point.X == b.Point.X {
                return 0
            }
        }

        return 1
    })

    return slots
}

func makeBuildingSlotsOld(city *citylib.City) []BuildingSlot {
    // use a random seed based on the position and name of the city so that each game gets
    // a different city view, but within the same game the city view is consistent
    random := rand.New(rand.NewPCG(uint64(city.X), uint64(city.Y) + hash(city.Name)))

    // for testing purposes, use a random seed
    // random = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
    openSlots := randomSlots(random)
    // openSlots := buildingSlots()

    var buildings []BuildingSlot

    for _, building := range sortBuildings(city.Buildings.Values()) {
        // city walls is handled specially
        if building == buildinglib.BuildingCityWalls {
            buildings = append(buildings, BuildingSlot{Building: building, Point: image.Pt(0, 75)})
            continue
        }

        if wasBuildingReplaced(building, city) {
            continue
        }

        if len(openSlots) == 0 {
            log.Printf("Ran out of open slots in city view for %+v for building %v", city, city.BuildingInfo.Name(building))
            continue
        }

        point := openSlots[0]
        openSlots = openSlots[1:]

        buildings = append(buildings, BuildingSlot{Building: building, Point: point})
    }

    maxTrees := random.IntN(15) + 20
    for i := 0; i < maxTrees; i++ {
        x := random.IntN(150) + 20
        y := random.IntN(60) + 10

        tree := []buildinglib.Building{BuildingTree1, BuildingTree2, BuildingTree3}[random.IntN(3)]

        buildings = append(buildings, BuildingSlot{Building: tree, Point: image.Pt(x, y)})
    }

    // FIXME: this is based on the population of the city
    maxHouses := random.IntN(15) + 20

    for i := 0; i < maxHouses; i++ {
        x := random.IntN(150) + 20
        y := random.IntN(60) + 10

        // house types are based on population size (village vs capital, etc)
        house := []buildinglib.Building{BuildingTreeHouse1, BuildingTreeHouse2, BuildingTreeHouse3, BuildingTreeHouse4, BuildingTreeHouse5}[random.IntN(5)]

        buildings = append(buildings, BuildingSlot{Building: house, Point: image.Pt(x, y)})
    }

    slices.SortFunc(buildings, func(a BuildingSlot, b BuildingSlot) int {
        if a.Point.Y < b.Point.Y {
            return -1
        }

        if a.Point.Y == b.Point.Y {
            if a.Point.X < b.Point.X {
                return -1
            }

            if a.Point.X == b.Point.X {
                return 0
            }
        }

        return 1
    })

    return buildings
}

func MakeCityScreen(cache *lbx.LbxCache, city *citylib.City, player *playerlib.Player, newBuilding buildinglib.Building) *CityScreen {

    fonts, err := fontslib.MakeCityViewFonts(cache)
    if err != nil {
        log.Printf("Could not make fonts: %v", err)
        return nil
    }

    buildings := makeBuildingSlots(city)

    cityScreen := &CityScreen{
        LbxCache: cache,
        ImageCache: util.MakeImageCache(cache),
        City: city,
        Fonts: fonts,
        Buildings: buildings,
        State: CityScreenStateRunning,
        Player: player,
    }

    cityScreen.UI = cityScreen.MakeUI(newBuilding)

    return cityScreen
}

func canSellBuilding(city *citylib.City, building buildinglib.Building) bool {
    return city.BuildingInfo.ProductionCost(building) > 0
}

func sellAmount(city *citylib.City, building buildinglib.Building) int {
    cost := city.BuildingInfo.ProductionCost(building) / 3
    if building == buildinglib.BuildingCityWalls {
        cost /= 2
    }

    if cost < 1 {
        cost = 1
    }

    return cost
}

func wasBuildingReplaced(building buildinglib.Building, city *citylib.City) bool {
    if building == buildinglib.BuildingNone {
        return false
    }

    replacedBy := building.ReplacedBy()
    return replacedBy != buildinglib.BuildingNone && (city.Buildings.Contains(replacedBy) || wasBuildingReplaced(replacedBy, city))
}

func (cityScreen *CityScreen) SellBuilding(building buildinglib.Building) {
    // convert the building pic to one of the rubble ones
    // give player back the gold for the building
    // remove building from the city

    cityScreen.City.SoldBuilding = true

    cityScreen.Player.Gold += sellAmount(cityScreen.City, building)

    cityScreen.City.Buildings.Remove(building)

    for i, _ := range cityScreen.Buildings {
        if cityScreen.Buildings[i].Building == building {
            cityScreen.Buildings[i].IsRubble = true
            cityScreen.Buildings[i].RubbleIndex = rand.IntN(4)
            break
        }
    }
}

func makeCityScapeElement(cache *lbx.LbxCache, group *uilib.UIElementGroup, city *citylib.City, help *helplib.Help, imageCache *util.ImageCache, doSell func(buildinglib.Building), buildings []BuildingSlot, newBuilding buildinglib.Building, x1 int, y1 int, fonts *fontslib.CityViewFonts, player *playerlib.Player, getAlpha *util.AlphaFadeFunc) *uilib.UIElement {
    // this stores the in-memory image because we don't need the gpu ebiten.Image to do pixel perfect detection
    rawImageCache := make(map[int]image.Image)

    getRawImage := func(index int) (image.Image, error) {
        if pic, ok := rawImageCache[index]; ok {
            return pic, nil
        }

        cityScapLbx, err := cache.GetLbxFile("cityscap.lbx")
        if err != nil {
            return nil, err
        }
        images, err := cityScapLbx.ReadImages(index)
        if err != nil {
            return nil, err
        }

        rawImageCache[index] = util.AutoCrop(images[0])

        return images[0], nil
    }

    roadX := 0.0
    roadY := 18.0

    buildingLook := buildinglib.BuildingNone
    buildingLookTime := uint64(0)
    buildingView := image.Rect(x1, y1, x1 + 206, y1 + 96)
    element := &uilib.UIElement{
        Rect: buildingView,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            var geom ebiten.GeoM
            geom.Translate(float64(x1), float64(y1))
            cityScapeScreen := screen.SubImage(scale.ScaleRect(buildingView)).(*ebiten.Image)
            drawCityScape(cityScapeScreen, city, buildings, buildingLook, buildingLookTime, newBuilding, group.Counter / 8, imageCache, fonts, player, geom, (*getAlpha)())
            // vector.StrokeRect(screen, float32(buildingView.Min.X), float32(buildingView.Min.Y), float32(buildingView.Dx()), float32(buildingView.Dy()), 1, color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff}, true)
        },
        RightClick: func(element *uilib.UIElement) {
            if buildingLook != buildinglib.BuildingNone {
                helpEntries := help.GetEntriesByName(getBuildingName(city.BuildingInfo, buildingLook))
                if helpEntries != nil {
                    group.AddElement(uilib.MakeHelpElement(group, cache, imageCache, helpEntries[0]))
                }
            }
        },
        LeftClick: func(element *uilib.UIElement) {
            if buildingLook != buildinglib.BuildingNone && canSellBuilding(city, buildingLook) {
                doSell(buildingLook)
            }
        },
        // if the user hovers over a building then show the name of the building
        Inside: func(element *uilib.UIElement, x int, y int){
            oldBuildingLook := buildingLook

            buildingLook = buildinglib.BuildingNone
            // log.Printf("inside building view: %v, %v", x, y)

            // go in reverse order so we select the one in front first
            for i := len(buildings) - 1; i >= 0; i-- {
                slot := buildings[i]

                buildingName := getBuildingName(city.BuildingInfo, slot.Building)

                if buildingName == "?" || buildingName == "" {
                    continue
                }

                index := GetBuildingIndex(slot.Building)

                pic, err := getRawImage(index)
                if err == nil {
                    x1 := int(roadX) + slot.Point.X
                    y1 := int(roadY) + slot.Point.Y - pic.Bounds().Dy()
                    x2 := x1 + pic.Bounds().Dx()
                    y2 := y1 + pic.Bounds().Dy()

                    // log.Printf("compare %v, %v to %v, %v, %v, %v for %v", x, y, x1, y1, x2, y2, buildingName)

                    // do pixel perfect detection
                    if image.Pt(x, y).In(image.Rect(x1, y1, x2, y2)) {
                        pixelX := x - x1
                        pixelY := y - y1

                        // log.Printf("look x=%v y=%v inside %v %v,%v,%v,%v at %v,%v", x, y, buildingName, x1, y1, x2, y2, pixelX, pixelY)

                        _, _, _, a := pic.At(pic.Bounds().Min.X + pixelX, pic.Bounds().Min.Y + pixelY).RGBA()
                        // log.Printf("  pixel value %v", pic.At(pixelX, pixelY))
                        if a > 0 {
                            buildingLook = slot.Building
                            // log.Printf("look at building %v (%v,%v) in (%v,%v,%v,%v)", slot.Building, useX, useY, x1, y1, x2, y2)
                            break
                        }
                    }
                }
            }

            if oldBuildingLook != buildingLook {
                buildingLookTime = 0
            } else {
                buildingLookTime += 1
            }

        },
    }

    return element
}

func (cityScreen *CityScreen) MakeUI(newBuilding buildinglib.Building) *uilib.UI {
    ui := &uilib.UI{
        Cache: cityScreen.LbxCache,
        Draw: func(ui *uilib.UI, screen *ebiten.Image) {
            ui.StandardDraw(screen)
        },
    }

    helpLbx, err := cityScreen.LbxCache.GetLbxFile("HELP.LBX")
    if err != nil {
        return nil
    }

    help, err := helplib.ReadHelp(helpLbx, 2)
    if err != nil {
        return nil
    }

    group := uilib.MakeGroup()
    ui.AddGroup(group)
    // var elements []*uilib.UIElement

    sellBuilding := func (toSell buildinglib.Building) {
        // FIXME: Check if building is needed for other building
        if cityScreen.City.SoldBuilding {
            ui.AddElement(uilib.MakeErrorElement(ui, cityScreen.LbxCache, &cityScreen.ImageCache, "You can only sell back one building per turn.", func(){}))
        } else {
            group := uilib.MakeGroup()
            var confirmElements []*uilib.UIElement

            yes := func(){
                cityScreen.SellBuilding(toSell)
                // update the ui
                cityScreen.UI = cityScreen.MakeUI(buildinglib.BuildingNone)
                ui.RemoveGroup(group)
            }

            no := func(){
                ui.RemoveGroup(group)
            }

            confirmElements = uilib.MakeConfirmDialog(group, cityScreen.LbxCache, &cityScreen.ImageCache, fmt.Sprintf("Are you sure you want to sell back the %v for %v gold?", cityScreen.City.BuildingInfo.Name(toSell), sellAmount(cityScreen.City, toSell)), true, yes, no)
            group.AddElements(confirmElements)
            ui.AddGroup(group)
        }
    }

    var getAlpha util.AlphaFadeFunc = func() float32 { return 1 }

    group.AddElement(makeCityScapeElement(cityScreen.LbxCache, group, cityScreen.City, &help, &cityScreen.ImageCache, sellBuilding, cityScreen.Buildings, newBuilding, 4, 101, cityScreen.Fonts, cityScreen.Player, &getAlpha))

    // returns the amount of gold and the amount of production that will be used to buy a building
    computeBuyAmount := func(cost int) (int, float32) {
        // FIXME: take terrain/race/buildings into effect, such as the miners guild for dwarves

        remaining := float32(cost) - cityScreen.City.Production
        if remaining > 0 {
            modifier := float32(4)
            if cityScreen.City.Production == 0 {
                modifier = 4
            } else if cityScreen.City.Production / float32(cost) < 1.0/3 {
                modifier = 3
            } else {
                modifier = 2
            }

            buyAmount := int(remaining * modifier)
            buyProduction := remaining

            return buyAmount, buyProduction
        }

        return 0, 0
    }

    buyAmount := 0
    buyProduction := float32(0)

    if cityScreen.City.ProducingBuilding != buildinglib.BuildingNone && cityScreen.City.ProducingBuilding != buildinglib.BuildingTradeGoods && cityScreen.City.ProducingBuilding != buildinglib.BuildingHousing {
        cost := float32(cityScreen.City.BuildingInfo.ProductionCost(cityScreen.City.ProducingBuilding))
        buyAmount, buyProduction = computeBuyAmount(int(cost))
    }

    if !cityScreen.City.ProducingUnit.IsNone() {
        buyAmount, buyProduction = computeBuyAmount(cityScreen.City.UnitProductionCost(&cityScreen.City.ProducingUnit))
    }

    // true if there is something to buy
    canBuy := buyProduction > 0 && buyAmount <= cityScreen.Player.Gold

    buyButtons, _ := cityScreen.ImageCache.GetImages("backgrnd.lbx", 7)
    if !canBuy {
        buyButtons, _ = cityScreen.ImageCache.GetImages("backgrnd.lbx", 14)
    }

    buyIndex := 0
    buyX := 214
    buyY := 188
    group.AddElement(&uilib.UIElement{
        Rect: image.Rect(buyX, buyY, buyX + buyButtons[0].Bounds().Dx(), buyY + buyButtons[0].Bounds().Dy()),
        PlaySoundLeftClick: true,
        LeftClick: func(element *uilib.UIElement) {
            if !canBuy {
                return
            }

            buyIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement) {
            if !canBuy {
                return
            }

            buyIndex = 0

            group := uilib.MakeGroup()

            yes := func(){
                cityScreen.City.Production += buyProduction
                cityScreen.Player.Gold -= buyAmount
                cityScreen.UI = cityScreen.MakeUI(buildinglib.BuildingNone)
                ui.RemoveGroup(group)
            }

            no := func(){
                ui.RemoveGroup(group)
            }

            var name string
            if cityScreen.City.ProducingBuilding != buildinglib.BuildingNone {
                name = cityScreen.City.BuildingInfo.Name(cityScreen.City.ProducingBuilding)
            } else {
                name = cityScreen.City.ProducingUnit.Name
            }
            message := fmt.Sprintf("Do you wish to spend %v by purchasing a %v?", buyAmount, name)
            elements := uilib.MakeConfirmDialog(group, cityScreen.LbxCache, &cityScreen.ImageCache, message, false, yes, no)
            group.AddElements(elements)
            ui.AddGroup(group)

        },
        RightClick: func(element *uilib.UIElement) {
            helpEntries := help.GetEntries(305)
            if helpEntries != nil {
                group.AddElement(uilib.MakeHelpElement(group, cityScreen.LbxCache, &cityScreen.ImageCache, helpEntries[0]))
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))

            use := 0
            if buyIndex < len(buyButtons) {
                use = buyIndex
            }

            scale.DrawScaled(screen, buyButtons[use], &options)
        },
    })

    // change button
    changeButton, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 8, 0)
    if err == nil {
        changeX := 247
        changeY := 188
        group.AddElement(&uilib.UIElement{
            Rect: image.Rect(changeX, changeY, changeX + changeButton.Bounds().Dx(), changeY + changeButton.Bounds().Dy()),
            PlaySoundLeftClick: true,
            LeftClick: func(element *uilib.UIElement) {
                if cityScreen.BuildScreen == nil {
                    cityScreen.BuildScreen = MakeBuildScreen(cityScreen.LbxCache, cityScreen.City)
                }
            },
            RightClick: func(element *uilib.UIElement) {
                helpEntries := help.GetEntries(306)
                if helpEntries != nil {
                    group.AddElement(uilib.MakeHelpElement(group, cityScreen.LbxCache, &cityScreen.ImageCache, helpEntries[0]))
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))
                scale.DrawScaled(screen, changeButton, &options)
            },
        })
    }

    // ok button
    okButton, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 9, 0)
    if err == nil {
        okX := 286
        okY := 188
        group.AddElement(&uilib.UIElement{
            Rect: image.Rect(okX, okY, okX + okButton.Bounds().Dx(), okY + okButton.Bounds().Dy()),
            PlaySoundLeftClick: true,
            LeftClick: func(element *uilib.UIElement) {
                cityScreen.State = CityScreenStateDone
            },
            RightClick: func(element *uilib.UIElement) {
                helpEntries := help.GetEntries(307)
                if helpEntries != nil {
                    group.AddElement(uilib.MakeHelpElement(group, cityScreen.LbxCache, &cityScreen.ImageCache, helpEntries[0]))
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))
                scale.DrawScaled(screen, okButton, &options)
            },
        })
    }

    enchantmentAreaRect := image.Rect(140, 51, 140 + 60, 93)
    maxEnchantments := enchantmentAreaRect.Dy() / (cityScreen.Fonts.BannerFonts[data.BannerGreen].Height())

    var enchantmentElements []*uilib.UIElement

    // if there are too many enchantments then up/down arrows will appear that let the user scroll the enchantment view
    for i, enchantment := range slices.SortedFunc(slices.Values(cityScreen.City.Enchantments.Values()), func (a citylib.Enchantment, b citylib.Enchantment) int {
        return cmp.Compare(a.Enchantment.Name(), b.Enchantment.Name())
    }) {
        useFont := cityScreen.Fonts.BannerFonts[enchantment.Owner]
        x := 140
        y := (51 + i * useFont.Height())

        rect := image.Rect(x, y, x + int(useFont.MeasureTextWidth(enchantment.Enchantment.Name(), 1)), y + useFont.Height())
        inside := false
        enchantmentElement := &uilib.UIElement{
            Rect: rect,
            Inside: func(element *uilib.UIElement, x int, y int){
                if image.Pt(x, y).In(enchantmentAreaRect.Sub(element.Rect.Min)) {
                    inside = true
                }
            },
            NotInside: func(element *uilib.UIElement){
                inside = false
            },
            LeftClickRelease: func(element *uilib.UIElement) {
                if !inside {
                    return
                }

                if enchantment.Owner == cityScreen.Player.GetBanner() {
                    group := uilib.MakeGroup()
                    yes := func(){
                        defer ui.RemoveGroup(group)
                        cityScreen.City.CancelEnchantment(enchantment.Enchantment, enchantment.Owner)
                        cityScreen.City.UpdateUnrest()

                        enchantmentBuildings := buildinglib.EnchantmentBuildings()
                        building, ok := enchantmentBuildings[enchantment.Enchantment]
                        if ok {
                            cityScreen.City.Buildings.Remove(building)

                            cityScreen.Buildings = slices.DeleteFunc(cityScreen.Buildings, func(slot BuildingSlot) bool {
                                return slot.Building == building
                            })
                        }

                        cityScreen.UI = cityScreen.MakeUI(buildinglib.BuildingNone)
                    }

                    no := func(){
                        ui.RemoveGroup(group)
                    }

                    confirmElements := uilib.MakeConfirmDialog(group, cityScreen.LbxCache, &cityScreen.ImageCache, fmt.Sprintf("Do you wish to turn off the %v spell?", enchantment.Enchantment.Name()), true, yes, no)
                    group.AddElements(confirmElements)
                    ui.AddGroup(group)
                } else {
                    ui.AddElement(uilib.MakeErrorElement(ui, cityScreen.LbxCache, &cityScreen.ImageCache, "You cannot cancel another wizard's enchantment.", func(){}))
                }
            },
            RightClick: func(element *uilib.UIElement){
                if !inside {
                    return
                }

                helpEntries := help.GetEntriesByName(enchantment.Enchantment.Name())
                if helpEntries != nil {
                    group.AddElement(uilib.MakeHelpElementWithLayer(group, cityScreen.LbxCache, &cityScreen.ImageCache, 2, helpEntries[0], helpEntries[1:]...))
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                area := screen.SubImage(scale.ScaleRect(enchantmentAreaRect)).(*ebiten.Image)

                useFont.PrintOptions(area, float64(element.Rect.Min.X), float64(element.Rect.Min.Y), font.FontOptions{Scale: scale.ScaleAmount}, enchantment.Enchantment.Name())
            },
        }

        group.AddElement(enchantmentElement)
        enchantmentElements = append(enchantmentElements, enchantmentElement)
    }

    if cityScreen.City.Enchantments.Size() > maxEnchantments {
        enchantmentMin := 0
        // shift all enchantment rect's around depending on the scroll position
        updateElements := func(){
            fontHeight := cityScreen.Fonts.BannerFonts[data.BannerGreen].Height()
            yOffset := enchantmentMin * fontHeight
            for i, element := range enchantmentElements {
                y := (51 + i * fontHeight)
                element.Rect.Min.Y = y - yOffset
                element.Rect.Max.Y = element.Rect.Min.Y + fontHeight
            }
        }

        upArrow, _ := cityScreen.ImageCache.GetImages("resource.lbx", 32)
        upArrowRect := util.ImageRect(200, 51, upArrow[0])
        upUse := 1
        group.AddElement(&uilib.UIElement{
            Rect: upArrowRect,
            LeftClick: func(element *uilib.UIElement){
                upUse = 0
            },
            LeftClickRelease: func(element *uilib.UIElement){
                enchantmentMin = max(0, enchantmentMin - 1)
                upUse = 1

                updateElements()
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(upArrowRect.Min.X), float64(upArrowRect.Min.Y))
                scale.DrawScaled(screen, upArrow[upUse], &options)
            },
        })

        downArrow, _ := cityScreen.ImageCache.GetImages("resource.lbx", 33)
        downArrowRect := util.ImageRect(200, 83, downArrow[0])
        downUse := 1
        group.AddElement(&uilib.UIElement{
            Rect: downArrowRect,
            LeftClick: func(element *uilib.UIElement){
                downUse = 0
            },
            LeftClickRelease: func(element *uilib.UIElement){
                downUse = 1
                enchantmentMin = min(cityScreen.City.Enchantments.Size() - maxEnchantments, enchantmentMin + 1)

                updateElements()
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(downArrowRect.Min.X), float64(downArrowRect.Min.Y))
                scale.DrawScaled(screen, downArrow[downUse], &options)
            },
        })

        // scroll wheel element
        group.AddElement(&uilib.UIElement{
            Rect: enchantmentAreaRect,
            Scroll: func (element *uilib.UIElement, x float64, y float64) {
                if y < 0 {
                    enchantmentMin = min(cityScreen.City.Enchantments.Size() - maxEnchantments, enchantmentMin + 1)
                    updateElements()
                } else if y > 0 {
                    enchantmentMin = max(0, enchantmentMin - 1)
                    updateElements()
                }
            },
        })

    }

    // FIXME: show Nightshade as a city enchantment if a nightshade tile is in the city catchment area and an appropriate building exists

    var resourceIcons []*uilib.UIElement
    resetResourceIcons := func(){
        ui.RemoveElements(resourceIcons)
        resourceIcons = cityScreen.CreateResourceIcons(ui)
        ui.AddElements(resourceIcons)
    }

    farmer, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", getRaceFarmerIndex(cityScreen.City.Race), 0)
    var setupWorkers func()
    if err == nil {
        workerY := float64(27)
        var workerElements []*uilib.UIElement
        setupWorkers = func(){
            ui.RemoveElements(workerElements)
            workerElements = nil
            citizenX := 6

            // the city might not have enough the required subsistence farmers, so only show what is available
            subsistenceFarmers := min(cityScreen.City.ComputeSubsistenceFarmers(), cityScreen.City.Farmers)

            for i := 0; i < subsistenceFarmers; i++ {
                posX := citizenX
                workerElements = append(workerElements, &uilib.UIElement{
                    Rect: util.ImageRect(posX, int(workerY), farmer),
                    Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                        var options ebiten.DrawImageOptions
                        options.GeoM.Translate(float64(posX), workerY)
                        scale.DrawScaled(screen, farmer, &options)
                    },
                    LeftClick: func(element *uilib.UIElement) {
                        cityScreen.City.Farmers = subsistenceFarmers
                        cityScreen.City.Workers = cityScreen.City.Citizens() - cityScreen.City.Rebels - cityScreen.City.Farmers
                        setupWorkers()
                    },
                })

                citizenX += farmer.Bounds().Dx()
            }

            // the farmers that can be changed to workers
            citizenX += 3
            for i := subsistenceFarmers; i < cityScreen.City.Farmers; i++ {
                posX := citizenX

                extraFarmer := i

                workerElements = append(workerElements, &uilib.UIElement{
                    Rect: util.ImageRect(posX, int(workerY), farmer),
                    Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                        var options ebiten.DrawImageOptions
                        options.GeoM.Translate(float64(posX), workerY)
                        scale.DrawScaled(screen, farmer, &options)
                    },
                    LeftClick: func(element *uilib.UIElement) {
                        cityScreen.City.Farmers = extraFarmer
                        cityScreen.City.Workers = cityScreen.City.Citizens() - cityScreen.City.Rebels - cityScreen.City.Farmers
                        setupWorkers()
                    },
                })

                citizenX += farmer.Bounds().Dx()
            }

            worker, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", getRaceWorkerIndex(cityScreen.City.Race), 0)
            if err == nil {
                for i := 0; i < cityScreen.City.Workers; i++ {
                    posX := citizenX

                    workerNum := i
                    workerElements = append(workerElements, &uilib.UIElement{
                        Rect: util.ImageRect(posX, int(workerY), worker),
                        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                            var options ebiten.DrawImageOptions
                            options.GeoM.Translate(float64(posX), workerY)
                            scale.DrawScaled(screen, worker, &options)
                        },
                        LeftClick: func(element *uilib.UIElement) {
                            cityScreen.City.Workers -= workerNum + 1
                            cityScreen.City.Farmers += workerNum + 1
                            setupWorkers()
                        },
                    })

                    citizenX += worker.Bounds().Dx()
                }
            }

            rebel, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", getRaceRebelIndex(cityScreen.City.Race), 0)
            if err == nil {
                citizenX += 3
                for i := 0; i < cityScreen.City.Rebels; i++ {
                    posX := citizenX

                    workerElements = append(workerElements, &uilib.UIElement{
                        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                            var options ebiten.DrawImageOptions
                            options.GeoM.Translate(float64(posX), workerY - float64(2))
                            scale.DrawScaled(screen, rebel, &options)
                        },
                    })

                    citizenX += rebel.Bounds().Dx()
                }
            }

            ui.AddElements(workerElements)

            resetResourceIcons()
        }
    } else {
        setupWorkers = func(){
        }
    }

    ui.SetElementsFromArray(nil)

    ui.AddElements(resourceIcons)

    var resetUnits func()

    var garrisonUnits []*uilib.UIElement
    resetUnits = func(){
        ui.RemoveElements(garrisonUnits)
        garrisonX := 216
        garrisonY := 103

        garrisonRow := 0

        var garrison []units.StackUnit

        cityStack := cityScreen.Player.FindStack(cityScreen.City.X, cityScreen.City.Y, cityScreen.City.Plane)
        if cityStack != nil {
            garrison = cityStack.Units()
        }

        for _, unit := range garrison {
            func (){
                garrisonBackground, err := units.GetUnitBackgroundImage(unit.GetBanner(), &cityScreen.ImageCache)
                if err != nil {
                    return
                }
                pic, err := cityScreen.ImageCache.GetImageTransform(unit.GetLbxFile(), unit.GetLbxIndex(), 0, unit.GetBanner().String(), units.MakeUpdateUnitColorsFunc(unit.GetBanner()))
                if err != nil {
                    return
                }

                posX := garrisonX
                posY := garrisonY
                useUnit := unit

                disband := func(){
                    cityScreen.Player.RemoveUnit(unit)
                    resetUnits()
                }

                garrisonElement := &uilib.UIElement{
                    Rect: util.ImageRect(posX, posY, garrisonBackground),
                    LeftClick: func(element *uilib.UIElement) {
                        cityScreen.State = CityScreenStateDone
                        cityScreen.Player.SelectedStack = cityScreen.Player.FindStackByUnit(useUnit)
                    },
                    RightClick: func(element *uilib.UIElement) {
                        ui.AddGroup(unitview.MakeUnitContextMenu(cityScreen.LbxCache, ui, useUnit, disband))
                    },
                    Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                        var options colorm.DrawImageOptions
                        var matrix colorm.ColorM
                        options.GeoM.Translate(float64(posX), float64(posY))
                        options.GeoM.Concat(scale.ScaledGeom)
                        colorm.DrawImage(screen, garrisonBackground, matrix, &options)
                        options.GeoM.Translate(1, 1)

                        // draw in grey scale if the unit is on patrol
                        if useUnit.GetBusy() == units.BusyStatusPatrol || useUnit.GetBusy() == units.BusyStatusStasis {
                            matrix.ChangeHSV(0, 0, 1)
                        }

                        colorm.DrawImage(screen, pic, matrix, &options)
                    },
                }

                garrisonUnits = append(garrisonUnits, garrisonElement)
                ui.AddElement(garrisonElement)

                garrisonX += pic.Bounds().Dx() + 1
                garrisonRow += 1
                if garrisonRow >= 5 {
                    garrisonRow = 0
                    garrisonX = 216
                    garrisonY += pic.Bounds().Dy() + 1
                }
            }()
        }
    }

    resetUnits()
    setupWorkers()

    return ui
}

func (cityScreen *CityScreen) Update() CityScreenState {
    cityScreen.Counter += 1

    if cityScreen.BuildScreen != nil {
        // allow city scape ui to update animations
        cityScreen.UI.Counter += 1

        switch cityScreen.BuildScreen.Update() {
            case BuildScreenRunning:
            case BuildScreenCanceled:
                cityScreen.BuildScreen = nil
                cityScreen.UI = cityScreen.MakeUI(buildinglib.BuildingNone)
            case BuildScreenOk:
                cityScreen.City.ProducingBuilding = cityScreen.BuildScreen.ProducingBuilding
                cityScreen.City.ProducingUnit = cityScreen.BuildScreen.ProducingUnit
                cityScreen.BuildScreen = nil
                cityScreen.UI = cityScreen.MakeUI(buildinglib.BuildingNone)
        }
    } else {
        cityScreen.UI.StandardUpdate()
    }

    return cityScreen.State
}

// the sprite used to show when building the thing (like when selecting buildings in the build screen)
func GetProducingBuildingIndex(building buildinglib.Building) int {
    if building == buildinglib.BuildingCityWalls {
        return 114
    }

    return GetBuildingIndex(building)
}

// the index in cityscap.lbx for the picture of this building
func GetBuildingIndex(building buildinglib.Building) int {
    index := building.Index()

    if index != -1 {
        return index
    }

    switch building {
        case BuildingTree1: return 19
        case BuildingTree2: return 20
        case BuildingTree3: return 21

        case buildinglib.BuildingTradeGoods: return 41

        // race specific housing is handled elsewhere
        case buildinglib.BuildingHousing: return 43

        case BuildingTreeHouse1: return 30
        case BuildingTreeHouse2: return 31
        case BuildingTreeHouse3: return 32
        case BuildingTreeHouse4: return 33
        case BuildingTreeHouse5: return 34

        case BuildingNormalHouse1: return 25
        case BuildingNormalHouse2: return 26
        case BuildingNormalHouse3: return 27
        case BuildingNormalHouse4: return 28
        case BuildingNormalHouse5: return 29

        case BuildingHutHouse1: return 35
        case BuildingHutHouse2: return 36
        case BuildingHutHouse3: return 37
        case BuildingHutHouse4: return 38
        case BuildingHutHouse5: return 39

        case buildinglib.BuildingAstralGate: return 85
        case buildinglib.BuildingAltarOfBattle: return 12
        case buildinglib.BuildingStreamOfLife: return 84
        case buildinglib.BuildingEarthGate: return 83
        case buildinglib.BuildingDarkRituals: return 81
    }

    return -1
}

// index into backgrnd.lbx for the farmer image of the given race
func getRaceFarmerIndex(race data.Race) int {
    switch race {
        case data.RaceNone: return -1
        case data.RaceBarbarian: return 59
        case data.RaceBeastmen: return 60
        case data.RaceDarkElf: return 61
        case data.RaceDraconian: return 62
        case data.RaceDwarf: return 63
        case data.RaceGnoll: return 64
        case data.RaceHalfling: return 65
        case data.RaceHighElf: return 66
        case data.RaceHighMen: return 67
        case data.RaceKlackon: return 68
        case data.RaceLizard: return 69
        case data.RaceNomad: return 70
        case data.RaceOrc: return 71
        case data.RaceTroll: return 72
    }

    return -1
}

func getRaceWorkerIndex(race data.Race) int {
    switch race {
        case data.RaceNone: return -1
        case data.RaceBarbarian: return 45
        case data.RaceBeastmen: return 46
        case data.RaceDarkElf: return 47
        case data.RaceDraconian: return 48
        case data.RaceDwarf: return 49
        case data.RaceGnoll: return 50
        case data.RaceHalfling: return 51
        case data.RaceHighElf: return 52
        case data.RaceHighMen: return 53
        case data.RaceKlackon: return 54
        case data.RaceLizard: return 55
        case data.RaceNomad: return 56
        case data.RaceOrc: return 57
        case data.RaceTroll: return 58
    }

    return -1
}

func getRaceRebelIndex(race data.Race) int {
    switch race {
        case data.RaceNone: return -1
        case data.RaceBarbarian: return 74
        case data.RaceBeastmen: return 75
        case data.RaceDarkElf: return 76
        case data.RaceDraconian: return 77
        case data.RaceDwarf: return 78
        case data.RaceGnoll: return 79
        case data.RaceHalfling: return 80
        case data.RaceHighElf: return 81
        case data.RaceHighMen: return 82
        case data.RaceKlackon: return 83
        case data.RaceLizard: return 84
        case data.RaceNomad: return 85
        case data.RaceOrc: return 86
        case data.RaceTroll: return 87
    }

    return -1
}

func getBuildingName(info buildinglib.BuildingInfos, building buildinglib.Building) string {
    switch building {
        case buildinglib.BuildingAstralGate: return "Astral Gate"
        case buildinglib.BuildingAltarOfBattle: return "Altar of Battle"
        case buildinglib.BuildingStreamOfLife: return "Stream of Life"
        case buildinglib.BuildingEarthGate: return "Earth Gate"
        case buildinglib.BuildingDarkRituals: return "Dark Rituals"
    }

    return info.Name(building)
}

func drawCityScape(screen *ebiten.Image, city *citylib.City, buildings []BuildingSlot, buildingLook buildinglib.Building, buildingLookTime uint64, newBuilding buildinglib.Building, animationCounter uint64, imageCache *util.ImageCache, fonts *fontslib.CityViewFonts, player *playerlib.Player, baseGeoM ebiten.GeoM, alphaScale float32) {
    onMyrror := city.Plane == data.PlaneMyrror

    // background
    spriteIndex := 0
    if onMyrror {
        spriteIndex = 8
    }

    animationIndex := 4
    hasFlyingFortress := city.HasEnchantment(data.CityEnchantmentFlyingFortress)
    switch {
        case hasFlyingFortress: animationIndex = 1
        case city.HasEnchantment(data.CityEnchantmentFamine): animationIndex = 2
        case city.HasEnchantment(data.CityEnchantmentCursedLands): animationIndex = 0
        case city.HasEnchantment(data.CityEnchantmentGaiasBlessing): animationIndex = 3
    }

    landBackground, err := imageCache.GetImage("cityscap.lbx", spriteIndex, animationIndex)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.ColorScale.ScaleAlpha(alphaScale)
        options.GeoM = baseGeoM
        scale.DrawScaled(screen, landBackground, &options)
    }

    // horizon
    spriteIndex = 7
    hasChaosRift := city.HasEnchantment(data.CityEnchantmentChaosRift)
    hasHeavenlyLight := city.HasEnchantment(data.CityEnchantmentHeavenlyLight)
    hasCloudOfShadow := city.HasEnchantment(data.CityEnchantmentCloudOfShadow)
    switch {
        case hasFlyingFortress: spriteIndex = -1
        case hasChaosRift && onMyrror: spriteIndex = 112
        case hasChaosRift: spriteIndex = 92
        case hasHeavenlyLight && onMyrror: spriteIndex = 113
        case hasHeavenlyLight: spriteIndex = 93
        case hasCloudOfShadow && onMyrror: spriteIndex = 111
        case hasCloudOfShadow: spriteIndex = 91
        // FIXME: hills and mountains seem to depend on the tiles north of the town
        //      hills && onMyrror: 9
        //      hills: 1
        //      mountains && onMyrror: 10
        //      mountains: 2
        case onMyrror: spriteIndex = 11
    }

    if spriteIndex > 0 {
        horizon, err := imageCache.GetImage("cityscap.lbx", spriteIndex, 0)
        if err == nil {
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(alphaScale)
            options.GeoM = baseGeoM
            options.GeoM.Translate(0, -1)
            scale.DrawScaled(screen, horizon, &options)
        }
    }

    // roads
    roadX := float64(0.0)
    roadY := float64(18.0)

    animationIndex = 0
    if onMyrror {
        animationIndex = 4
    }

    normalRoad, err := imageCache.GetImage("cityscap.lbx", 5, animationIndex)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.ColorScale.ScaleAlpha(alphaScale)
        options.GeoM = baseGeoM
        options.GeoM.Translate(roadX, roadY)
        scale.DrawScaled(screen, normalRoad, &options)
    }

    drawName := func(){
    }

    var buildingLookPoint image.Point
    if buildingLook != buildinglib.BuildingNone {
        for _, building := range buildings {
            if building.Building == buildingLook {
                buildingLookPoint = building.Point
                break
            }
        }
    }

    // river / shore
    spriteIndex = -1
    onRiver := city.ByRiver()
    onShore := city.OnShore()
    switch {
        case onShore && onMyrror: spriteIndex = 116
        case onShore: spriteIndex = 4
        case onRiver && onMyrror: spriteIndex = 115
        case onRiver: spriteIndex = 3
    }
    if spriteIndex != -1 {
        river, err := imageCache.GetImages("cityscap.lbx", spriteIndex)
        if err == nil {
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(alphaScale)
            options.GeoM = baseGeoM
            options.GeoM.Translate(0, -2)
            index := animationCounter % uint64(len(river))
            scale.DrawScaled(screen, river[index], &options)
        }
    }

    // buildings
    for _, building := range buildings {
        index := GetBuildingIndex(building.Building)

        if building.IsRubble {
            index = 105 + building.RubbleIndex
        }

        x, y := building.Point.X, building.Point.Y

        images, err := imageCache.GetImagesTransform("cityscap.lbx", index, "crop", util.AutoCrop)
        // images, err := imageCache.GetImages("cityscap.lbx", index)
        if err == nil {
            animationIndex := animationCounter % uint64(len(images))
            use := images[animationIndex]
            var options ebiten.DrawImageOptions

            useAlpha := alphaScale
            if newBuilding == building.Building {
                if animationCounter < 10 {
                    useAlpha *= float32(animationCounter) / 10
                }
            }

            options.ColorScale.ScaleAlpha(useAlpha)
            options.GeoM = baseGeoM

            if buildingLook == building.Building && buildingLookTime > 0 {
                options.ColorScale.Scale(1.1, 1.1, 1, 1)
            }

            if buildingLook != buildinglib.BuildingNone && buildingLook != building.Building && buildingLookTime > 0 {
                xDiff := building.Point.X - buildingLookPoint.X
                yDiff := building.Point.Y - buildingLookPoint.Y
                if xDiff * xDiff + yDiff * yDiff < 600 {
                    options.ColorScale.ScaleAlpha(max(1 - float32(buildingLookTime) / 20, 0.5))
                }
            }

            /*
            options.GeoM.Translate(float64(x) + roadX, float64(y) + roadY)
            dx, dy := options.GeoM.Apply(0, 0)
            vector.DrawFilledCircle(screen, float32(dx), float32(dy), 2, color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff}, false)
            */

            // x,y position is the bottom left of the sprite
            options.GeoM.Translate(float64(x) + roadX, float64(y) + roadY)
            // dx, dy := options.GeoM.Apply(0, 0)
            // vector.DrawFilledCircle(screen, float32(dx), float32(dy), 2, color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff}, false)
            // options.GeoM.Translate(float64(x) + roadX, float64(y - use.Bounds().Dy()) + roadY)
            options.GeoM.Translate(0, float64(-use.Bounds().Dy()))
            scale.DrawScaled(screen, use, &options)

            if buildingLook == building.Building {
                drawName = func(){
                    useFont := fonts.SmallFont
                    text := getBuildingName(city.BuildingInfo, building.Building)

                    if building.IsRubble {
                        text = "Destroyed " + text
                        useFont = fonts.RubbleFont
                    }

                    if building.Building == buildinglib.BuildingFortress {
                        text = fmt.Sprintf("%v's Fortress", player.Wizard.Name)
                    }

                    printX, printY := baseGeoM.Apply(float64(x + use.Bounds().Dx() / 2) + roadX, float64(y + 1) + roadY)

                    useFont.PrintOptions(screen, printX, printY, font.FontOptions{Justify: font.FontJustifyCenter, DropShadow: true, Scale: scale.ScaleAmount, Options: &options}, text)
                }
            }
        }
    }

    // evil presence
    if city.HasEnchantment(data.CityEnchantmentEvilPresence) {
        for _, building := range buildings {
            if building.Building.IsReligous() && !building.IsRubble {
                index := GetBuildingIndex(building.Building)
                images, err := imageCache.GetImagesTransform("cityscap.lbx", index, "crop", util.AutoCrop)
                if err != nil {
                    continue
                }
                dx := images[0].Bounds().Dx()

                index = data.CityEnchantmentEvilPresence.LbxIndex()
                images, err = imageCache.GetImagesTransform("cityscap.lbx", index, "crop", util.AutoCrop)
                if err != nil {
                    continue
                }

                use := images[0]
                x, y := building.Point.X, building.Point.Y

                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(0.6)
                options.GeoM = baseGeoM
                options.GeoM.Translate(float64(x) + roadX, float64(y) + roadY)
                options.GeoM.Translate(float64(dx - use.Bounds().Dx())/2, float64(-use.Bounds().Dy()))

                scale.DrawScaled(screen, use, &options)
            }
        }
    }

    // magic walls and enchantment icons
    enchantments := city.Enchantments.Values()
    for _, enchantment := range slices.SortedFunc(slices.Values(enchantments), func (a citylib.Enchantment, b citylib.Enchantment) int {
        return cmp.Compare(a.Enchantment.Name(), b.Enchantment.Name())
    }) {
        switch enchantment.Enchantment {
            case data.CityEnchantmentWallOfFire, data.CityEnchantmentWallOfDarkness:
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(alphaScale)
                options.GeoM = baseGeoM
                options.GeoM.Translate(0, 85)
                images, _ := imageCache.GetImages("cityscap.lbx", enchantment.Enchantment.LbxIndex())
                index := animationCounter % uint64(len(images))
                scale.DrawScaled(screen, images[index], &options)
            case data.CityEnchantmentNaturesEye, data.CityEnchantmentProsperity, data.CityEnchantmentConsecration,
                data.CityEnchantmentInspirations, data.CityEnchantmentLifeWard, data.CityEnchantmentSorceryWard,
                data.CityEnchantmentChaosWard, data.CityEnchantmentDeathWard, data.CityEnchantmentNatureWard:
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(alphaScale)
                options.GeoM = baseGeoM
                options.GeoM.Translate(float64(enchantment.Enchantment.IconOffset()), 82)
                image, _ := imageCache.GetImage("cityscap.lbx", enchantment.Enchantment.LbxIndex(), 0)
                scale.DrawScaled(screen, image, &options)
        }
    }

    drawName()

    /*
    x, y := baseGeoM.Apply(0, 0)
    vector.DrawFilledCircle(screen, float32(x), float32(y), 2, color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff}, false)
    */
}

func (cityScreen *CityScreen) FoodProducers() []ResourceUsage {
    var usages []ResourceUsage

    // FIXME: should this take enchantments like famine into account and only show half the food production?
    usages = append(usages, ResourceUsage{
        Count: cityScreen.City.FarmerFoodProduction(cityScreen.City.Farmers),
        Name: "Farmers",
    })

    buildings := cityScreen.City.Buildings.Values()
    for _, building := range sortBuildings(buildings) {
        value := 0
        switch building {
            case buildinglib.BuildingForestersGuild: value = 2
            case buildinglib.BuildingGranary: value = 2
            case buildinglib.BuildingFarmersMarket: value = 3
        }

        if value > 0 {
            usages = append(usages, ResourceUsage{
                Count: value,
                Name: cityScreen.City.BuildingInfo.Name(building),
                Replaced: wasBuildingReplaced(building, cityScreen.City),
            })
        }
    }

    wildGame := cityScreen.City.ComputeWildGame()
    if wildGame > 0 {
        usages = append(usages, ResourceUsage{
            Count: wildGame,
            Name: "Wild Game",
        })
    }

    return usages
}

// pass screen=nil to compute how wide the icons are without drawing them
func (cityScreen *CityScreen) drawIcons(total int, small *ebiten.Image, large *ebiten.Image, options ebiten.DrawImageOptions, screen *ebiten.Image) ebiten.GeoM {
    largeGap := large.Bounds().Dx()

    var optionsM colorm.DrawImageOptions
    optionsM.GeoM = options.GeoM
    var matrix colorm.ColorM
    matrix.Scale(float64(options.ColorScale.R()), float64(options.ColorScale.G()), float64(options.ColorScale.B()), float64(options.ColorScale.A()))

    // draw icons in grey scale
    if total < 0 {
        matrix.ChangeHSV(0, 0, 1)
        total = -total
    }

    totalIcons := total / 10 + total % 10

    if totalIcons > 3 {
        largeGap -= 1
    }

    if totalIcons > 5 {
        largeGap -= 4
    }

    for range total / 10 {
        if screen != nil {
            oldGeom := optionsM.GeoM
            // screen.DrawImage(large, &options)
            optionsM.GeoM.Concat(scale.ScaledGeom)
            colorm.DrawImage(screen, large, matrix, &optionsM)
            optionsM.GeoM = oldGeom
        }
        optionsM.GeoM.Translate(float64(largeGap), 0)
    }

    smallGap := small.Bounds().Dx() + 1
    if totalIcons > 4 {
        smallGap -= 1
    }
    if totalIcons >= 8 {
        smallGap -= 1
    }

    for range total % 10 {
        if screen != nil {
            // screen.DrawImage(small, &options)
            oldGeom := optionsM.GeoM
            optionsM.GeoM.Concat(scale.ScaledGeom)
            colorm.DrawImage(screen, small, matrix, &optionsM)
            optionsM.GeoM = oldGeom
        }
        optionsM.GeoM.Translate(float64(smallGap), 0)
    }

    return optionsM.GeoM
}

// copied heavily from ui/dialogs.go:MakeHelpElementWithLayer
func (cityScreen *CityScreen) MakeResourceDialog(title string, smallIcon *ebiten.Image, bigIcon *ebiten.Image, ui *uilib.UI, resources []ResourceUsage) []*uilib.UIElement {
    helpTop, err := cityScreen.ImageCache.GetImage("help.lbx", 0, 0)
    if err != nil {
        return nil
    }

    fonts := fontslib.MakeCityViewResourceFonts(cityScreen.LbxCache)

    const fadeSpeed = 7

    getAlpha := ui.MakeFadeIn(fadeSpeed)

    infoX := 55
    // infoY := 30
    infoWidth := helpTop.Bounds().Dx()
    // infoHeight := screen.HelpTop.Bounds().Dy()
    infoLeftMargin := 18
    infoTopMargin := 24
    infoBodyMargin := 3
    maxInfoWidth := infoWidth - infoLeftMargin - infoBodyMargin - 14

    // fmt.Printf("Help text: %v\n", []byte(help.Text))

    helpTextY := infoTopMargin
    helpTextY += fonts.HelpTitleFont.Height() + 1

    maxResources := 14

    textHeight := (min(len(resources), maxResources) + 1) * (fonts.HelpFont.Height() + 1)

    bottom := helpTextY + textHeight

    // only draw as much of the top scroll as there are lines of text
    topImage := helpTop.SubImage(image.Rect(0, 0, helpTop.Bounds().Dx(), int(bottom))).(*ebiten.Image)
    helpBottom, err := cityScreen.ImageCache.GetImage("help.lbx", 1, 0)
    if err != nil {
        return nil
    }

    infoY := (data.ScreenHeight - bottom - helpBottom.Bounds().Dy()) / 2

    makeRenderPage := func(resources []ResourceUsage) func (screen *ebiten.Image) {
        widestResources := float64(0)
        for _, usage := range resources {
            var options ebiten.DrawImageOptions
            geom := cityScreen.drawIcons(int(math.Abs(float64(usage.Count))), smallIcon, bigIcon, options, nil)
            x, _ := geom.Apply(0, 0)
            if x > widestResources {
                widestResources = x
            }
        }

        return func (window *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(infoX), float64(infoY))
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(window, topImage, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(infoX), float64(bottom + infoY))
            scale.DrawScaled(window, helpBottom, &options)

            // for debugging
            // vector.StrokeRect(window, float32(infoX), float32(infoY), float32(infoWidth), float32(infoHeight), 1, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, true)
            // vector.StrokeRect(window, float32(infoX + infoLeftMargin), float32(infoY + infoTopMargin), float32(maxInfoWidth), float32(screen.HelpTitleFont.Height() + 20 + 1), 1, color.RGBA{R: 0, G: 0xff, B: 0, A: 0xff}, false)

            titleX := infoX + infoLeftMargin + maxInfoWidth / 2

            fonts.HelpTitleFont.PrintOptions(window, float64(titleX), float64(infoY + infoTopMargin), font.FontOptions{Justify: font.FontJustifyCenter, Options: &options, Scale: scale.ScaleAmount}, title)

            yPos := infoY + (infoTopMargin + fonts.HelpTitleFont.Height() + 1)
            xPos := (infoX + infoLeftMargin)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(xPos), float64(yPos))

            for _, usage := range resources {
                if usage.Count < 0 {
                    x, y := options.GeoM.Apply(0, 1)
                    fonts.HelpFont.PrintOptions(window, x, y, font.FontOptions{Justify: font.FontJustifyRight, Options: &options, Scale: scale.ScaleAmount}, "-")
                }

                cityScreen.drawIcons(int(math.Abs(float64(usage.Count))), smallIcon, bigIcon, options, window)

                x, y := options.GeoM.Apply(widestResources + float64(5), 0)

                text := usage.Name
                if usage.Replaced {
                    text = fmt.Sprintf("%v (Replaced)", usage.Name)
                }
                text += fmt.Sprintf(" (%v)", usage.Count)
                fonts.HelpFont.PrintOptions(window, x, y, font.FontOptions{Options: &options, Scale: scale.ScaleAmount}, text)
                yPos += (fonts.HelpFont.Height() + 1)
                options.GeoM.Translate(0, float64((fonts.HelpFont.Height() + 1)))
            }
        }
    }

    var renderPages []func (screen *ebiten.Image)
    for page := 0; page < len(resources) / maxResources + 1; page++ {
        start := page * maxResources
        end := min(len(resources), (page + 1) * maxResources)

        renderPages = append(renderPages, makeRenderPage(resources[start: end]))
    }
    // renderPage := makeRenderPage(resources[:min(len(resources), maxResources)])

    currentPage := 0

    var elements []*uilib.UIElement

    elements = append(elements, &uilib.UIElement{
        // Rect: image.Rect(infoX, infoY, infoX + infoWidth, infoY + infoHeight),
        Rect: image.Rect(0, 0, data.ScreenWidth, data.ScreenHeight),
        Draw: func (infoThis *uilib.UIElement, window *ebiten.Image){
            renderPages[currentPage](window)
        },
        LeftClick: func(infoThis *uilib.UIElement){
            getAlpha = ui.MakeFadeOut(fadeSpeed)
            ui.AddDelay(fadeSpeed, func(){
                ui.RemoveElements(elements)
            })
        },
        Layer: 1,
    })

    if len(renderPages) > 1 {
        width := fonts.HelpFont.MeasureTextWidth("More", 1)
        height := float64(fonts.HelpFont.Height())

        var geom ebiten.GeoM
        geom.Translate(float64(infoX), float64(bottom + infoY))
        geom.Translate(float64(infoWidth) - width - float64(18), -height)
        x1, y1 := geom.Apply(0, 0)
        x2, y2 := geom.Apply(width, height)

        moreRect := image.Rect(int(x1), int(y1), int(x2), int(y2))
        elements = append(elements, &uilib.UIElement{
            Rect: moreRect,
            Draw: func (this *uilib.UIElement, window *ebiten.Image){
                // vector.StrokeRect(window, float32(moreRect.Min.X), float32(moreRect.Min.Y), float32(moreRect.Bounds().Dx()), float32(moreRect.Bounds().Dy()), 2, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, true)

                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(getAlpha())
                options.GeoM.Reset()
                options.GeoM.Translate(float64(infoX), float64(bottom + infoY))
                options.GeoM.Translate(float64(infoWidth) - fonts.HelpFont.MeasureTextWidth("More", 1) - float64(18), -float64(fonts.HelpFont.Height()))
                x, y := options.GeoM.Apply(0, 0)

                fonts.HelpFont.PrintOptions(window, x, y, font.FontOptions{Options: &options, Scale: scale.ScaleAmount}, "More")
            },
            LeftClick: func(this *uilib.UIElement){
                currentPage = (currentPage + 1) % len(renderPages)
            },
            Order: 1,
            Layer: 1,
        })
    }

    return elements
}

func (cityScreen *CityScreen) WorkProducers() []ResourceUsage {
    var usage []ResourceUsage

    add := func(count int, name string, building buildinglib.Building){
        if count > 0 {
            usage = append(usage, ResourceUsage{
                Count: count,
                Name: name,
                Replaced: wasBuildingReplaced(building, cityScreen.City),
            })
        }
    }

    add(int(cityScreen.City.ProductionWorkers()), "Workers", buildinglib.BuildingNone)
    add(int(cityScreen.City.ProductionFarmers()), "Farmers", buildinglib.BuildingNone)
    add(int(cityScreen.City.ProductionTerrain()), "Terrain", buildinglib.BuildingNone)
    add(int(cityScreen.City.ProductionSawmill()), "Sawmill", buildinglib.BuildingSawmill)
    add(int(cityScreen.City.ProductionForestersGuild()), "Forester's Guild", buildinglib.BuildingForestersGuild)
    add(int(cityScreen.City.ProductionMinersGuild()), "Miner's Guild", buildinglib.BuildingMinersGuild)
    add(int(cityScreen.City.ProductionMechaniciansGuild()), "Mechanician's Guild", buildinglib.BuildingMechaniciansGuild)
    add(int(cityScreen.City.ProductionInspirations()), "Inspirations", buildinglib.BuildingNone)
    // FIXME: should this show halfed production in case of cursed lands?

    return usage
}

func (cityScreen *CityScreen) BuildingMaintenanceResources() []ResourceUsage {
    var usage []ResourceUsage

    for _, building := range cityScreen.City.Buildings.Values() {
        maintenance := cityScreen.City.BuildingInfo.UpkeepCost(building)
        if maintenance == 0 {
            continue
        }

        usage = append(usage, ResourceUsage{
            Count: maintenance,
            Name: cityScreen.City.BuildingInfo.Name(building),
            Replaced: wasBuildingReplaced(building, cityScreen.City),
        })
    }

    return usage
}

func (cityScreen *CityScreen) GoldProducers() []ResourceUsage {
    var usage []ResourceUsage

    add := func(count int, name string, building buildinglib.Building){
        if count > 0 {
            usage = append(usage, ResourceUsage{
                Count: count,
                Name: name,
                Replaced: wasBuildingReplaced(building, cityScreen.City),
            })
        }
    }

    // FIXME: add tiles (road/river/ocean)
    add(int(cityScreen.City.GoldTaxation()), "Taxes", buildinglib.BuildingNone)
    add(int(cityScreen.City.GoldTradeGoods()), "Trade Goods", buildinglib.BuildingTradeGoods)
    add(int(cityScreen.City.GoldMinerals()), "Minerals", buildinglib.BuildingNone)
    add(int(cityScreen.City.GoldMarketplace()), "Marketplace", buildinglib.BuildingMarketplace)
    add(int(cityScreen.City.GoldBank()), "Bank", buildinglib.BuildingBank)
    add(int(cityScreen.City.GoldMerchantsGuild()), "Merchant's Guild", buildinglib.BuildingMerchantsGuild)
    add(int(cityScreen.City.GoldProsperity()), "Prosperity", buildinglib.BuildingNone)

    return usage
}

func (cityScreen *CityScreen) PowerProducers() []ResourceUsage {
    var usage []ResourceUsage

    add := func(count int, name string, building buildinglib.Building){
        if count != 0 {
            usage = append(usage, ResourceUsage{
                Count: count,
                Name: name,
                Replaced: wasBuildingReplaced(building, cityScreen.City),
            })
        }
    }

    add(cityScreen.City.PowerCitizens(), "Townsfolk", buildinglib.BuildingNone)
    add(cityScreen.City.PowerFortress(), "Fortress", buildinglib.BuildingFortress)
    add(int(cityScreen.City.PowerShrine()), "Shrine", buildinglib.BuildingShrine)
    add(int(cityScreen.City.PowerTemple()), "Temple", buildinglib.BuildingTemple)
    add(int(cityScreen.City.PowerParthenon()), "Parthenon", buildinglib.BuildingParthenon)
    add(int(cityScreen.City.PowerCathedral()), "Cathedral", buildinglib.BuildingCathedral)
    add(cityScreen.City.PowerAlchemistsGuild(), "Alchemist's Guild", buildinglib.BuildingAlchemistsGuild)
    add(cityScreen.City.PowerWizardsGuild(), "Wizard's Guild", buildinglib.BuildingWizardsGuild)
    add(cityScreen.City.PowerMinerals(), "Minerals", buildinglib.BuildingNone)
    add(int(cityScreen.City.PowerDarkRituals()), "Dark Rituals", buildinglib.BuildingNone)

    // FIXME: add tiles (adamantium mine) and miner's guild

    return usage
}

func (cityScreen *CityScreen) ResearchProducers() []ResourceUsage {
    var usage []ResourceUsage

    for _, building := range sortBuildings(cityScreen.City.Buildings.Values()) {
        research := cityScreen.City.BuildingInfo.ResearchProduction(building)

        if research > 0 {
            usage = append(usage, ResourceUsage{
                Count: research,
                Name: cityScreen.City.BuildingInfo.Name(building),
                Replaced: wasBuildingReplaced(building, cityScreen.City),
            })
        }
    }

    return usage
}

func (cityScreen *CityScreen) CreateResourceIcons(ui *uilib.UI) []*uilib.UIElement {
    foodRequired := cityScreen.City.RequiredFood()
    foodSurplus := cityScreen.City.SurplusFood()

    foodRequired = int(max(0, min(foodRequired, foodRequired + foodSurplus)))

    smallFood, _ := cityScreen.ImageCache.GetImage("backgrnd.lbx", 40, 0)
    bigFood, _ := cityScreen.ImageCache.GetImage("backgrnd.lbx", 88, 0)

    smallHammer, _ := cityScreen.ImageCache.GetImage("backgrnd.lbx", 41, 0)
    bigHammer, _ := cityScreen.ImageCache.GetImage("backgrnd.lbx", 89, 0)

    smallCoin, _ := cityScreen.ImageCache.GetImage("backgrnd.lbx", 42, 0)
    bigCoin, _ := cityScreen.ImageCache.GetImage("backgrnd.lbx", 90, 0)

    smallMagic, _ := cityScreen.ImageCache.GetImage("backgrnd.lbx", 43, 0)
    bigMagic, _ := cityScreen.ImageCache.GetImage("backgrnd.lbx", 91, 0)

    smallResearch, _ := cityScreen.ImageCache.GetImage("backgrnd.lbx", 44, 0)
    bigResearch, _ := cityScreen.ImageCache.GetImage("backgrnd.lbx", 92, 0)

    var elements []*uilib.UIElement

    foodRect := image.Rect(6, 52, 6 + 9 * bigFood.Bounds().Dx(), 52 + bigFood.Bounds().Dy())
    elements = append(elements, &uilib.UIElement{
        Rect: foodRect,
        LeftClick: func(element *uilib.UIElement) {
            foodProducers := cityScreen.FoodProducers()
            ui.AddElements(cityScreen.MakeResourceDialog("Food", smallFood, bigFood, ui, foodProducers))
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(foodRect.Min.X), float64(foodRect.Min.Y))
            options.GeoM = cityScreen.drawIcons(foodRequired, smallFood, bigFood, options, screen)
            options.GeoM.Translate(float64(5), 0)
            cityScreen.drawIcons(foodSurplus, smallFood, bigFood, options, screen)
        },
    })

    production := cityScreen.City.WorkProductionRate()
    workRect := image.Rect(6, 60, 6 + 9 * bigHammer.Bounds().Dx(), 60 + bigHammer.Bounds().Dy())
    elements = append(elements, &uilib.UIElement{
        Rect: workRect,
        LeftClick: func(element *uilib.UIElement) {
            workProducers := cityScreen.WorkProducers()
            ui.AddElements(cityScreen.MakeResourceDialog("Production", smallHammer, bigHammer, ui, workProducers))
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(workRect.Min.X), float64(workRect.Min.Y))
            cityScreen.drawIcons(int(production), smallHammer, bigHammer, options, screen)
        },
    })

    var goldUpkeepOptions ebiten.DrawImageOptions
    goldGeom := cityScreen.drawIcons(cityScreen.City.ComputeUpkeep(), smallCoin, bigCoin, goldUpkeepOptions, nil)

    x, _ := goldGeom.Apply(0, 0)

    // FIXME: if income - upkeep < 0 then show greyed out icons for gold

    goldMaintenanceRect := image.Rect(6, 68, 6 + int(x), 68 + bigCoin.Bounds().Dy())
    elements = append(elements, &uilib.UIElement{
        Rect: goldMaintenanceRect,
        LeftClick: func(element *uilib.UIElement) {
            maintenance := cityScreen.BuildingMaintenanceResources()
            ui.AddElements(cityScreen.MakeResourceDialog("Building Maintenance", smallCoin, bigCoin, ui, maintenance))
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(goldMaintenanceRect.Min.X), float64(goldMaintenanceRect.Min.Y))
            cityScreen.drawIcons(cityScreen.City.ComputeUpkeep(), smallCoin, bigCoin, options, screen)

            // util.DrawRect(screen, goldMaintenanceRect, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff})
        },
    })

    goldSurplus := cityScreen.City.GoldSurplus()

    goldGeom = cityScreen.drawIcons(goldSurplus, smallCoin, bigCoin, goldUpkeepOptions, nil)
    x, _ = goldGeom.Apply(0, 0)
    goldSurplusRect := image.Rect(goldMaintenanceRect.Max.X + 6, 68, goldMaintenanceRect.Max.X + 6 + int(x), 68 + bigCoin.Bounds().Dy())
    elements = append(elements, &uilib.UIElement{
        Rect: goldSurplusRect,
        LeftClick: func(element *uilib.UIElement) {
            gold := cityScreen.GoldProducers()
            ui.AddElements(cityScreen.MakeResourceDialog("Gold", smallCoin, bigCoin, ui, gold))
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(goldSurplusRect.Min.X), float64(goldSurplusRect.Min.Y))
            cityScreen.drawIcons(goldSurplus, smallCoin, bigCoin, options, screen)
            // util.DrawRect(screen, goldSurplusRect, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff})
        },
    })

    powerRect := image.Rect(6, 76, 6 + 9 * bigMagic.Bounds().Dx(), 76 + bigMagic.Bounds().Dy())
    elements = append(elements, &uilib.UIElement{
        Rect: powerRect,
        LeftClick: func(element *uilib.UIElement) {
            power := cityScreen.PowerProducers()
            ui.AddElements(cityScreen.MakeResourceDialog("Power", smallMagic, bigMagic, ui, power))
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(powerRect.Min.X), float64(powerRect.Min.Y))
            cityScreen.drawIcons(cityScreen.City.ComputePower(), smallMagic, bigMagic, options, screen)
        },
    })

    researchRect := image.Rect(6, 84, 6 + 9 * bigResearch.Bounds().Dx(), 84 + bigResearch.Bounds().Dy())
    elements = append(elements, &uilib.UIElement{
        Rect: researchRect,
        LeftClick: func(element *uilib.UIElement) {
            research := cityScreen.ResearchProducers()
            ui.AddElements(cityScreen.MakeResourceDialog("Spell Research", smallResearch, bigResearch, ui, research))
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(researchRect.Min.X), float64(researchRect.Min.Y))
            cityScreen.drawIcons(cityScreen.City.ResearchProduction(), smallResearch, bigResearch, options, screen)
        },
    })

    return elements
}

func (cityScreen *CityScreen) Draw(screen *ebiten.Image, mapView func (screen *ebiten.Image, geom ebiten.GeoM, counter uint64), tileWidth int, tileHeight int) {
    animationCounter := cityScreen.Counter / 8

    ui, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 6, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        scale.DrawScaled(screen, ui, &options)
    }

    cityScreen.Fonts.BigFont.PrintOptions(screen, 20, 3, font.FontOptions{DropShadow: true, Scale: scale.ScaleAmount}, fmt.Sprintf("%v of %s", cityScreen.City.GetSize(), cityScreen.City.Name))
    cityScreen.Fonts.DescriptionFont.PrintOptions(screen, 6, 19, font.FontOptions{Scale: scale.ScaleAmount}, fmt.Sprintf("%v", cityScreen.City.Race))

    deltaNumber := func(n int) string {
        if n > 0 {
            return fmt.Sprintf("+%v", n)
        } else if n == 0 {
            return "0"
        } else {
            return fmt.Sprintf("%v", n)
        }
    }

    cityScreen.Fonts.DescriptionFont.PrintOptions(screen, 210, 19, font.FontOptions{Justify: font.FontJustifyRight, Scale: scale.ScaleAmount}, fmt.Sprintf("Population: %v (%v)", cityScreen.City.Population, deltaNumber(cityScreen.City.PopulationGrowthRate())))

    showWork := false
    workRequired := 0

    if cityScreen.City.ProducingBuilding != buildinglib.BuildingNone {
        lbxIndex := GetProducingBuildingIndex(cityScreen.City.ProducingBuilding)

        if cityScreen.City.ProducingBuilding == buildinglib.BuildingHousing {
            switch cityScreen.City.Race.HouseType() {
                case data.HouseTypeTree: lbxIndex = 43
                case data.HouseTypeHut: lbxIndex = 44
                case data.HouseTypeNormal: lbxIndex = 42
            }
        }

        producingPics, err := cityScreen.ImageCache.GetImages("cityscap.lbx", lbxIndex)
        if err == nil {
            index := animationCounter % uint64(len(producingPics))

            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(217), float64(144))
            scale.DrawScaled(screen, producingPics[index], &options)
        }

        cityScreen.Fonts.ProducingFont.PrintOptions(screen, 237, 179, font.FontOptions{Justify: font.FontJustifyCenter, DropShadow: true, Scale: scale.ScaleAmount}, cityScreen.City.BuildingInfo.Name(cityScreen.City.ProducingBuilding))

        // for all buildings besides trade goods and housing, show amount of work required to build

        if cityScreen.City.ProducingBuilding == buildinglib.BuildingTradeGoods || cityScreen.City.ProducingBuilding == buildinglib.BuildingHousing {
            producingBackground, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 13, 0)
            if err == nil {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(260), float64(149))
                scale.DrawScaled(screen, producingBackground, &options)
            }

            description := ""
            switch cityScreen.City.ProducingBuilding {
                case buildinglib.BuildingTradeGoods: description = "Trade Goods"
                case buildinglib.BuildingHousing: description = "Increases population growth rate."
            }

            cityScreen.Fonts.ProducingFont.PrintWrap(screen, 285, 155, 60, font.FontOptions{Justify: font.FontJustifyCenter, DropShadow: true, Scale: scale.ScaleAmount}, description)
        } else {
            showWork = true
            workRequired = cityScreen.City.BuildingInfo.ProductionCost(cityScreen.City.ProducingBuilding)
        }
    } else if !cityScreen.City.ProducingUnit.IsNone() {
        images, err := cityScreen.ImageCache.GetImages(cityScreen.City.ProducingUnit.CombatLbxFile, cityScreen.City.ProducingUnit.GetCombatIndex(units.FacingRight))
        if err == nil {
            var options ebiten.DrawImageOptions
            use := images[2]

            options.GeoM.Translate(238, 168)
            unitview.RenderCombatTile(screen, &cityScreen.ImageCache, options)
            unitview.RenderCombatUnit(screen, use, options, cityScreen.City.ProducingUnit.Count, data.UnitEnchantmentNone, 0, nil)
            cityScreen.Fonts.ProducingFont.PrintOptions(screen, 237, 179, font.FontOptions{Justify: font.FontJustifyCenter, DropShadow: true, Scale: scale.ScaleAmount}, cityScreen.City.ProducingUnit.Name)
        }

        showWork = true
        workRequired = cityScreen.City.UnitProductionCost(&cityScreen.City.ProducingUnit)
    }

    if showWork {
        turn := ""
        turns := (float64(workRequired) - float64(cityScreen.City.Production)) / float64(cityScreen.City.WorkProductionRate())
        if turns <= 0 {
            turn = "1 Turn"
        } else {
            turn = fmt.Sprintf("%v Turns", int(math.Ceil(turns)))
        }

        cityScreen.Fonts.DescriptionFont.PrintOptions(screen, 318, 140, font.FontOptions{Justify: font.FontJustifyRight, Scale: scale.ScaleAmount}, turn)

        workEmpty, err1 := cityScreen.ImageCache.GetImage("backgrnd.lbx", 11, 0)
        workFull, err2 := cityScreen.ImageCache.GetImage("backgrnd.lbx", 12, 0)
        if err1 == nil && err2 == nil {
            startX := 262

            x := startX
            y := 151

            coinsPerRow := 10
            xSpacing := 5

            if workRequired / 10 > 50 {
                coinsPerRow = 20
                xSpacing = 2
            }

            coinsProduced := float64(cityScreen.City.Production) / 10.0

            row := 0
            for i := 0; i < workRequired / 10; i++ {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(x), float64(y))

                if coinsProduced > float64(i) {
                    leftOver := coinsProduced - float64(i)
                    if leftOver >= 1 {
                        scale.DrawScaled(screen, workFull, &options)
                    } else if leftOver > 0.05 {
                        scale.DrawScaled(screen, workEmpty, &options)
                        part := workFull.SubImage(image.Rect(0, 0, int(float64(workFull.Bounds().Dx()) * leftOver), workFull.Bounds().Dy())).(*ebiten.Image)
                        scale.DrawScaled(screen, part, &options)
                    }

                } else {
                    scale.DrawScaled(screen, workEmpty, &options)
                }

                row += 1
                if row >= coinsPerRow {
                    y += workFull.Bounds().Dy()
                    x = startX
                    row = 0
                } else {
                    x += xSpacing
                }
            }
        }
    }

    // draw a few squares of the map
    mapX := 215
    mapY := 4
    mapWidth := 100
    mapHeight := 88
    mapPart := screen.SubImage(scale.ScaleRect(image.Rect(mapX, mapY, mapX + mapWidth, mapY + mapHeight))).(*ebiten.Image)
    var mapGeom ebiten.GeoM
    mapGeom.Translate(float64(mapX), float64(mapY))
    mapView(mapPart, mapGeom, cityScreen.Counter)

    // darken the 4 corners of the small map view
    drawDarkTile := func(x int, y int){
        x1, y1 := mapGeom.Apply(float64(x * tileWidth), float64(y * tileHeight))
        vector.DrawFilledRect(mapPart, float32(x1), float32(y1), float32(tileWidth), float32(tileHeight), color.RGBA{R: 0, G: 0, B: 0, A: 0x80}, false)
    }

    drawDarkTile(0, 0)
    drawDarkTile(4, 0)
    drawDarkTile(0, 4)
    drawDarkTile(4, 4)

    cityScreen.UI.Draw(cityScreen.UI, screen)

    if cityScreen.BuildScreen != nil {
        // screen.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 0x80})
        vector.DrawFilledRect(screen, 0, 0, float32(screen.Bounds().Dx()), float32(screen.Bounds().Dy()), color.RGBA{R: 0, G: 0, B: 0, A: 0x80}, true)
        cityScreen.BuildScreen.Draw(screen)
    }
}

// when right clicking on an enemy city, this just shows the population, garrison, and city scape for that city
func SimplifiedView(cache *lbx.LbxCache, city *citylib.City, player *playerlib.Player, otherPlayer *playerlib.Player) (func(coroutine.YieldFunc, func()), func(*ebiten.Image)) {
    imageCache := util.MakeImageCache(cache)

    fonts, err := fontslib.MakeCityViewFonts(cache)
    if err != nil {
        log.Printf("Could not make fonts: %v", err)
        return func(yield coroutine.YieldFunc, update func()){}, func(*ebiten.Image){}
    }

    helpLbx, err := cache.GetLbxFile("help.lbx")
    if err != nil {
        log.Printf("Error with help: %v", err)
        return func(yield coroutine.YieldFunc, update func()){}, func(*ebiten.Image){}
    }

    help, err := helplib.ReadHelp(helpLbx, 2)
    if err != nil {
        log.Printf("Error with help: %v", err)
        return func(yield coroutine.YieldFunc, update func()){}, func(*ebiten.Image){}
    }

    quit := false

    buildings := makeBuildingSlots(city)

    background, _ := imageCache.GetImage("reload.lbx", 26, 0)
    var options ebiten.DrawImageOptions
    options.GeoM.Translate(float64(data.ScreenWidth / 2 - background.Bounds().Dx() / 2), 0)

    var getAlpha util.AlphaFadeFunc

    currentUnitName := ""

    ui := &uilib.UI{
        LeftClick: func(){
            quit = true
        },
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            useOptions := options
            useOptions.ColorScale.ScaleAlpha(getAlpha())

            scale.DrawScaled(screen, background, &useOptions)

            fontOptions := font.FontOptions{Options: &useOptions, Scale: scale.ScaleAmount}

            titleX, titleY := options.GeoM.Apply(float64(20), float64(3))
            fonts.BigFont.PrintOptions(screen, titleX, titleY, fontOptions, fmt.Sprintf("%v of %s", city.GetSize(), city.Name))
            raceX, raceY := options.GeoM.Apply(float64(6), float64(19))
            fonts.DescriptionFont.PrintOptions(screen, raceX, raceY, fontOptions, fmt.Sprintf("%v", city.Race))

            unitsX, unitsY := options.GeoM.Apply(float64(6), float64(43))
            fonts.DescriptionFont.PrintOptions(screen, unitsX, unitsY, fontOptions, fmt.Sprintf("Units   %v", currentUnitName))

            ui.StandardDraw(screen)
        },
    }

    group := uilib.MakeGroup()

    getAlpha = ui.MakeFadeIn(7)

    ui.SetElementsFromArray(nil)

    var setupUI func()
    setupUI = func(){
        ui.RemoveGroup(group)
        group = uilib.MakeGroup()
        ui.AddGroup(group)
        x1, y1 := options.GeoM.Apply(5, 102)

        cityScapeElement := makeCityScapeElement(cache, group, city, &help, &imageCache, func(buildinglib.Building){}, buildings, buildinglib.BuildingNone, int(x1), int(y1), fonts, otherPlayer, &getAlpha)

        group.AddElement(cityScapeElement)

        // draw all farmers/workers/rebels
        group.AddElement(&uilib.UIElement{
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                localOptions := options
                localOptions.ColorScale.ScaleAlpha(getAlpha())
                localOptions.GeoM.Translate(float64(6), float64(27))
                farmer, _ := imageCache.GetImage("backgrnd.lbx", getRaceFarmerIndex(city.Race), 0)
                worker, _ := imageCache.GetImage("backgrnd.lbx", getRaceWorkerIndex(city.Race), 0)
                rebel, _ := imageCache.GetImage("backgrnd.lbx", getRaceRebelIndex(city.Race), 0)
                subsistenceFarmers := city.ComputeSubsistenceFarmers()

                for range subsistenceFarmers {
                    scale.DrawScaled(screen, farmer, &localOptions)
                    localOptions.GeoM.Translate(float64(farmer.Bounds().Dx()), 0)
                }

                localOptions.GeoM.Translate(float64(3), 0)

                for range city.Farmers - subsistenceFarmers {
                    scale.DrawScaled(screen, farmer, &localOptions)
                    localOptions.GeoM.Translate(float64(farmer.Bounds().Dx()), 0)
                }

                for range city.Workers {
                    scale.DrawScaled(screen, worker, &localOptions)
                    localOptions.GeoM.Translate(float64(worker.Bounds().Dx()), 0)
                }

                localOptions.GeoM.Translate(float64(3), float64(-2))

                for range city.Rebels {
                    scale.DrawScaled(screen, rebel, &localOptions)
                    localOptions.GeoM.Translate(float64(rebel.Bounds().Dx()), 0)
                }
            },
        })

        stack := otherPlayer.FindStack(city.X, city.Y, city.Plane)
        if stack != nil && player.IsVisible(city.X, city.Y, city.Plane) {
            inside := 0
            for i, unit := range stack.Units() {
                x, y := options.GeoM.Apply(float64(8), float64(52))

                x += float64((i % 6) * 20)
                y += float64((i / 6) * 20)

                pic, _ := imageCache.GetImageTransform(unit.GetLbxFile(), unit.GetLbxIndex(), 0, unit.GetBanner().String(), units.MakeUpdateUnitColorsFunc(unit.GetBanner()))
                rect := image.Rect(int(x), int(y), int(x) + pic.Bounds().Dx(), int(y) + pic.Bounds().Dy())
                group.AddElement(&uilib.UIElement{
                    Rect: rect,
                    Inside: func(element *uilib.UIElement, x int, y int){
                        currentUnitName = unit.GetName()
                        inside = i
                    },
                    NotInside: func(element *uilib.UIElement){
                        if inside == i {
                            currentUnitName = ""
                        }
                    },
                    Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                        var localOptions ebiten.DrawImageOptions
                        localOptions.ColorScale.ScaleAlpha(getAlpha())
                        localOptions.GeoM.Translate(x, y)
                        scale.DrawScaled(screen, pic, &localOptions)
                    },
                })
            }
        }

        for i, enchantment := range slices.SortedFunc(slices.Values(city.Enchantments.Values()), func (a citylib.Enchantment, b citylib.Enchantment) int {
            return cmp.Compare(a.Enchantment.Name(), b.Enchantment.Name())
        }) {
            useFont := fonts.BannerFonts[enchantment.Owner]
            // failsafe, but should never happen
            if useFont == nil {
                continue
            }
            x, y := options.GeoM.Apply(float64(142), float64((51 + i * useFont.Height())))
            rect := image.Rect(int(x), int(y), int(x + useFont.MeasureTextWidth(enchantment.Enchantment.Name(), 1)), int(y) + useFont.Height())
            group.AddElement(&uilib.UIElement{
                Rect: rect,
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    var options ebiten.DrawImageOptions
                    options.ColorScale.ScaleAlpha(getAlpha())
                    useFont.PrintOptions(screen, x, y, font.FontOptions{Options: &options, Scale: scale.ScaleAmount}, enchantment.Enchantment.Name())
                },
                LeftClick: func(element *uilib.UIElement) {
                    if enchantment.Owner == player.GetBanner() {
                        yes := func(){
                            city.CancelEnchantment(enchantment.Enchantment, enchantment.Owner)
                            city.UpdateUnrest()
                            setupUI()
                        }

                        no := func(){
                        }

                        confirmElements := uilib.MakeConfirmDialog(group, cache, &imageCache, fmt.Sprintf("Do you wish to turn off the %v spell?", enchantment.Enchantment.Name()), true, yes, no)
                        group.AddElements(confirmElements)
                    }
                },
                RightClick: func(element *uilib.UIElement){
                    helpEntries := help.GetEntriesByName(enchantment.Enchantment.Name())
                    if helpEntries != nil {
                        group.AddElement(uilib.MakeHelpElementWithLayer(group, cache, &imageCache, 2, helpEntries[0], helpEntries[1:]...))
                    }
                },
            })
        }
    }

    setupUI()

    draw := func(screen *ebiten.Image){
        ui.Draw(ui, screen)
    }

    logic := func(yield coroutine.YieldFunc, update func()){
        countdown := 0
        for countdown != 1 {
            update()
            ui.StandardUpdate()

            if countdown == 0 && quit {
                countdown = 8
                getAlpha = ui.MakeFadeOut(7)
            }

            if countdown > 1 {
                countdown -= 1
            }

            yield()
        }
    }

    return logic, draw
}
