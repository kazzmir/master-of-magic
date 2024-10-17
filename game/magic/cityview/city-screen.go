package cityview

import (
    "log"
    "fmt"
    "math"
    "math/rand/v2"
    "sort"
    "image"
    "image/color"
    "hash/fnv"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/unitview"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/hajimehoshi/ebiten/v2"
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

type BuildingSlotSort []BuildingSlot

func (b BuildingSlotSort) Len() int {
    return len(b)
}

func (b BuildingSlotSort) Less(i, j int) bool {
    return b[i].Point.Y < b[j].Point.Y || (b[i].Point.Y == b[j].Point.Y && b[i].Point.X < b[j].Point.X)
}

func (b BuildingSlotSort) Swap(i, j int) {
    b[i], b[j] = b[j], b[i]
}

type CityScreenState int

const (
    CityScreenStateRunning CityScreenState = iota
    CityScreenStateDone
)

type CityScreen struct {
    LbxCache *lbx.LbxCache
    ImageCache util.ImageCache
    BigFont *font.Font
    DescriptionFont *font.Font
    ProducingFont *font.Font
    SmallFont *font.Font
    RubbleFont *font.Font
    City *citylib.City

    UI *uilib.UI
    Player *playerlib.Player

    Buildings []BuildingSlot
    BuildScreen *BuildScreen
    // the building the user is currently hovering their mouse over
    BuildingLook buildinglib.Building

    Counter uint64
    State CityScreenState
}

type BuildingNativeSort []buildinglib.Building
func (b BuildingNativeSort) Len() int {
    return len(b)
}
func (b BuildingNativeSort) Less(i, j int) bool {
    return b[i] < b[j]
}
func (b BuildingNativeSort) Swap(i, j int) {
    b[i], b[j] = b[j], b[i]
}

func sortBuildings(buildings []buildinglib.Building) []buildinglib.Building {
    sort.Sort(BuildingNativeSort(buildings))
    return buildings
}

func hash(str string) uint64 {
    hasher := fnv.New64a()
    hasher.Write([]byte(str))
    return hasher.Sum64()
}

func MakeCityScreen(cache *lbx.LbxCache, city *citylib.City, player *playerlib.Player) *CityScreen {

    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
    }

    yellowPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xfa, G: 0xe1, B: 0x16, A: 0xff},
            color.RGBA{R: 0xff, G: 0xef, B: 0x2f, A: 0xff},
            color.RGBA{R: 0xff, G: 0xef, B: 0x2f, A: 0xff},
            color.RGBA{R: 0xe0, G: 0x8a, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xd2, G: 0x7f, B: 0x0, A: 0xff},
            color.RGBA{R: 0xd2, G: 0x7f, B: 0x0, A: 0xff},
            color.RGBA{R: 0x99, G: 0x4f, B: 0x0, A: 0xff},
            color.RGBA{R: 0x99, G: 0x4f, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0x99, G: 0x4f, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        }

    bigFont := font.MakeOptimizedFontWithPalette(fonts[5], yellowPalette)

    brownPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0xe1, G: 0x8e, B: 0x32, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
    }

    // fixme: make shadow font as well
    descriptionFont := font.MakeOptimizedFontWithPalette(fonts[1], brownPalette)

    whitePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
    }

    producingFont := font.MakeOptimizedFontWithPalette(fonts[1], whitePalette)

    smallFontPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0x0},
        color.RGBA{R: 128, G: 128, B: 128, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
    }

    // FIXME: this font should have a black outline around all the glyphs
    smallFont := font.MakeOptimizedFontWithPalette(fonts[1], smallFontPalette)

    rubbleFontPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0x0},
        color.RGBA{R: 128, G: 0, B: 0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
    }

    // FIXME: this font should have a black outline around all the glyphs
    rubbleFont := font.MakeOptimizedFontWithPalette(fonts[1], rubbleFontPalette)

    // use a random seed based on the position and name of the city so that each game gets
    // a different city view, but within the same game the city view is consistent
    random := rand.New(rand.NewPCG(uint64(city.X), uint64(city.Y) + hash(city.Name)))

    // for testing purposes, use a random seed
    // random = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
    openSlots := randomSlots(random)
    // openSlots := buildingSlots()

    var buildings []BuildingSlot

    for _, building := range sortBuildings(city.Buildings.Values()) {
        if len(openSlots) == 0 {
            log.Printf("Ran out of open slots in city view for %+v", city)
            break
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

    sort.Sort(BuildingSlotSort(buildings))

    cityScreen := &CityScreen{
        LbxCache: cache,
        ImageCache: util.MakeImageCache(cache),
        City: city,
        BigFont: bigFont,
        DescriptionFont: descriptionFont,
        ProducingFont: producingFont,
        SmallFont: smallFont,
        RubbleFont: rubbleFont,
        Buildings: buildings,
        State: CityScreenStateRunning,
        Player: player,
    }

    cityScreen.UI = cityScreen.MakeUI()

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

func (cityScreen *CityScreen) MakeUI() *uilib.UI {
    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image) {
            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })
        },
    }

    helpLbx, err := cityScreen.LbxCache.GetLbxFile("HELP.LBX")
    if err != nil {
        return nil
    }

    help, err := helpLbx.ReadHelp(2)
    if err != nil {
        return nil
    }

    var elements []*uilib.UIElement

    roadX := 4
    roadY := 120

    rawImageCache := make(map[int]image.Image)

    getRawImage := func(index int) (image.Image, error) {
        if pic, ok := rawImageCache[index]; ok {
            return pic, nil
        }

        cityScapLbx, err := cityScreen.LbxCache.GetLbxFile("cityscap.lbx")
        if err != nil {
            return nil, err
        }
        images, err := cityScapLbx.ReadImages(index)
        if err != nil {
            return nil, err
        }

        rawImageCache[index] = images[0]

        return images[0], nil
    }

    buildingView := image.Rect(5, 103, 208, 195)
    elements = append(elements, &uilib.UIElement{
        Rect: buildingView,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            // vector.StrokeRect(screen, float32(buildingView.Min.X), float32(buildingView.Min.Y), float32(buildingView.Dx()), float32(buildingView.Dy()), 1, color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff}, true)
        },
        RightClick: func(element *uilib.UIElement) {
            if cityScreen.BuildingLook != buildinglib.BuildingNone {
                helpEntries := help.GetEntriesByName(cityScreen.City.BuildingInfo.Name(cityScreen.BuildingLook))
                if helpEntries != nil {
                    ui.AddElement(uilib.MakeHelpElement(ui, cityScreen.LbxCache, &cityScreen.ImageCache, helpEntries[0]))
                }
            }
        },
        LeftClick: func(element *uilib.UIElement) {
            if cityScreen.BuildingLook != buildinglib.BuildingNone && canSellBuilding(cityScreen.City, cityScreen.BuildingLook) {
                if cityScreen.City.SoldBuilding {
                    ui.AddElement(uilib.MakeErrorElement(cityScreen.UI, cityScreen.LbxCache, &cityScreen.ImageCache, "You can only sell back one building per turn."))
                } else {
                    var confirmElements []*uilib.UIElement

                    yes := func(){
                        cityScreen.SellBuilding(cityScreen.BuildingLook)
                    }

                    no := func(){
                    }

                    confirmElements = uilib.MakeConfirmDialog(cityScreen.UI, cityScreen.LbxCache, &cityScreen.ImageCache, fmt.Sprintf("Are you sure you want to sell back the %v for %v gold?", cityScreen.City.BuildingInfo.Name(cityScreen.BuildingLook), sellAmount(cityScreen.City, cityScreen.BuildingLook)), yes, no)
                    ui.AddElements(confirmElements)
                }
            }
        },
        // if the user hovers over a building then show the name of the building
        Inside: func(element *uilib.UIElement, x int, y int){
            cityScreen.BuildingLook = buildinglib.BuildingNone
            // log.Printf("inside building view: %v, %v", x, y)

            // go in reverse order so we select the one in front first
            for i := len(cityScreen.Buildings) - 1; i >= 0; i-- {
                slot := cityScreen.Buildings[i]

                buildingName := cityScreen.City.BuildingInfo.Name(slot.Building)

                if buildingName == "?" || buildingName == "" {
                    continue
                }

                index := GetBuildingIndex(slot.Building)

                pic, err := getRawImage(index)
                if err == nil {
                    x1 := roadX + slot.Point.X
                    y1 := roadY + slot.Point.Y - pic.Bounds().Dy()
                    x2 := x1 + pic.Bounds().Dx()
                    y2 := y1 + pic.Bounds().Dy()

                    useX := x + buildingView.Min.X
                    useY := y + buildingView.Min.Y

                    // do pixel perfect detection
                    if image.Pt(useX, useY).In(image.Rect(x1, y1, x2, y2)) {
                        pixelX := useX - x1
                        pixelY := useY - y1

                        _, _, _, a := pic.At(pixelX, pixelY).RGBA()
                        if a > 0 {
                            cityScreen.BuildingLook = slot.Building
                            // log.Printf("look at building %v (%v,%v) in (%v,%v,%v,%v)", slot.Building, useX, useY, x1, y1, x2, y2)
                            break
                        }
                    }
                }
            }
        },
    })

    // FIXME: show disabled buy button if the item is not buyable (not enough money, or the item is trade goods/housing)
    // buy button
    buyButton, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 7, 0)
    if err == nil {
        buyX := 214
        buyY := 188
        elements = append(elements, &uilib.UIElement{
            Rect: image.Rect(buyX, buyY, buyX + buyButton.Bounds().Dx(), buyY + buyButton.Bounds().Dy()),
            LeftClick: func(element *uilib.UIElement) {

                var elements []*uilib.UIElement

                yes := func(){
                    // FIXME: buy the thing being produced
                }

                no := func(){
                }

                elements = uilib.MakeConfirmDialog(cityScreen.UI, cityScreen.LbxCache, &cityScreen.ImageCache, "Are you sure you want to buy this building?", yes, no)
                ui.AddElements(elements)
            },
            RightClick: func(element *uilib.UIElement) {
                helpEntries := help.GetEntries(305)
                if helpEntries != nil {
                    ui.AddElement(uilib.MakeHelpElement(ui, cityScreen.LbxCache, &cityScreen.ImageCache, helpEntries[0]))
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(buyX), float64(buyY))
                screen.DrawImage(buyButton, &options)
            },
        })
    }

    // change button
    changeButton, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 8, 0)
    if err == nil {
        changeX := 247
        changeY := 188
        elements = append(elements, &uilib.UIElement{
            Rect: image.Rect(changeX, changeY, changeX + changeButton.Bounds().Dx(), changeY + changeButton.Bounds().Dy()),
            LeftClick: func(element *uilib.UIElement) {
                if cityScreen.BuildScreen == nil {
                    cityScreen.BuildScreen = MakeBuildScreen(cityScreen.LbxCache, cityScreen.City, cityScreen.City.ProducingBuilding, cityScreen.City.ProducingUnit)
                }
            },
            RightClick: func(element *uilib.UIElement) {
                helpEntries := help.GetEntries(306)
                if helpEntries != nil {
                    ui.AddElement(uilib.MakeHelpElement(ui, cityScreen.LbxCache, &cityScreen.ImageCache, helpEntries[0]))
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(changeX), float64(changeY))
                screen.DrawImage(changeButton, &options)
            },
        })
    }

    // ok button
    okButton, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 9, 0)
    if err == nil {
        okX := 286
        okY := 188
        elements = append(elements, &uilib.UIElement{
            Rect: image.Rect(okX, okY, okX + okButton.Bounds().Dx(), okY + okButton.Bounds().Dy()),
            LeftClick: func(element *uilib.UIElement) {
                cityScreen.State = CityScreenStateDone
            },
            RightClick: func(element *uilib.UIElement) {
                helpEntries := help.GetEntries(307)
                if helpEntries != nil {
                    ui.AddElement(uilib.MakeHelpElement(ui, cityScreen.LbxCache, &cityScreen.ImageCache, helpEntries[0]))
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(okX), float64(okY))
                screen.DrawImage(okButton, &options)
            },
        })
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

            subsistenceFarmers := cityScreen.City.ComputeSubsistenceFarmers()

            for i := 0; i < subsistenceFarmers; i++ {
                posX := citizenX
                workerElements = append(workerElements, &uilib.UIElement{
                    Rect: util.ImageRect(posX, int(workerY), farmer),
                    Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                        var options ebiten.DrawImageOptions
                        options.GeoM.Translate(float64(posX), workerY)
                        screen.DrawImage(farmer, &options)
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
                        screen.DrawImage(farmer, &options)
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
                            screen.DrawImage(worker, &options)
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
                for i := 0; i < cityScreen.City.Rebels; i++ {
                    posX := citizenX

                    workerElements = append(workerElements, &uilib.UIElement{
                        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                            var options ebiten.DrawImageOptions
                            options.GeoM.Translate(float64(posX), workerY-2)
                            screen.DrawImage(rebel, &options)
                        },
                    })

                    citizenX += rebel.Bounds().Dx()
                }
            }

            ui.AddElements(workerElements)
        }
    } else {
        setupWorkers = func(){
        }
    }

    garrisonX := 216
    garrisonY := 103

    garrisonRow := 0

    var garrison []units.StackUnit

    cityStack := cityScreen.Player.FindStack(cityScreen.City.X, cityScreen.City.Y)
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
                // FIXME: implement disband
            }

            elements = append(elements, &uilib.UIElement{
                Rect: util.ImageRect(posX, posY, garrisonBackground),
                LeftClick: func(element *uilib.UIElement) {
                    cityScreen.State = CityScreenStateDone
                    cityScreen.Player.SelectedStack = cityScreen.Player.FindStackByUnit(useUnit)
                },
                RightClick: func(element *uilib.UIElement) {
                    ui.AddElements(unitview.MakeUnitContextMenu(cityScreen.LbxCache, ui, useUnit, disband))
                },
                Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                    var options colorm.DrawImageOptions
                    var matrix colorm.ColorM
                    options.GeoM.Translate(float64(posX), float64(posY))
                    colorm.DrawImage(screen, garrisonBackground, matrix, &options)
                    options.GeoM.Translate(1, 1)

                    // draw in grey scale if the unit is on patrol
                    if useUnit.GetPatrol() {
                        matrix.ChangeHSV(0, 0, 1)
                    }

                    colorm.DrawImage(screen, pic, matrix, &options)
                },
            })

            garrisonX += pic.Bounds().Dx() + 1
            garrisonRow += 1
            if garrisonRow >= 5 {
                garrisonRow = 0
                garrisonX = 216
                garrisonY += pic.Bounds().Dy() + 1
            }
        }()
    }

    ui.SetElementsFromArray(elements)
    setupWorkers()

    return ui
}

func (cityScreen *CityScreen) Update() CityScreenState {
    cityScreen.Counter += 1

    if cityScreen.BuildScreen != nil {
        switch cityScreen.BuildScreen.Update() {
            case BuildScreenRunning:
            case BuildScreenCanceled:
                cityScreen.BuildScreen = nil
            case BuildScreenOk:
                cityScreen.City.ProducingBuilding = cityScreen.BuildScreen.ProducingBuilding
                cityScreen.City.ProducingUnit = cityScreen.BuildScreen.ProducingUnit
                cityScreen.BuildScreen = nil
        }
    } else {
        cityScreen.UI.StandardUpdate()
    }

    return cityScreen.State
}

// the index in cityscap.lbx for the picture of this building
func GetBuildingIndex(building buildinglib.Building) int {
    index := buildinglib.GetBuildingIndex(building)

    if index != -1 {
        return index
    }

    switch building {
        case BuildingTree1: return 19
        case BuildingTree2: return 20
        case BuildingTree3: return 21

        case buildinglib.BuildingTradeGoods: return 41

        // FIXME: housing is indices 42-44, based on the race of the city
        case buildinglib.BuildingHousing: return 43

        case BuildingTreeHouse1: return 30
        case BuildingTreeHouse2: return 31
        case BuildingTreeHouse3: return 32
        case BuildingTreeHouse4: return 33
        case BuildingTreeHouse5: return 34
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

func (cityScreen *CityScreen) Draw(screen *ebiten.Image, mapView func (screen *ebiten.Image, geom ebiten.GeoM, counter uint64)) {
    animationCounter := cityScreen.Counter / 8

    // 5 is grasslands
    landBackground, err := cityScreen.ImageCache.GetImage("cityscap.lbx", 0, 4)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(4, 102)
        screen.DrawImage(landBackground, &options)
    }

    hills1, err := cityScreen.ImageCache.GetImage("cityscap.lbx", 7, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(4, 101)
        screen.DrawImage(hills1, &options)
    }

    roadX := 4.0
    roadY := 120.0

    normalRoad, err := cityScreen.ImageCache.GetImage("cityscap.lbx", 5, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(roadX, roadY)
        screen.DrawImage(normalRoad, &options)
    }

    drawName := func(){
    }

    for _, building := range cityScreen.Buildings {

        index := GetBuildingIndex(building.Building)

        if building.IsRubble {
            index = 105 + building.RubbleIndex
        }

        x, y := building.Point.X, building.Point.Y

        images, err := cityScreen.ImageCache.GetImages("cityscap.lbx", index)
        if err == nil {
            animationIndex := animationCounter % uint64(len(images))
            use := images[animationIndex]
            var options ebiten.DrawImageOptions
            // x,y position is the bottom left of the sprite
            options.GeoM.Translate(float64(x) + roadX, float64(y - use.Bounds().Dy()) + roadY)
            screen.DrawImage(use, &options)

            if cityScreen.BuildingLook == building.Building {
                drawName = func(){
                    useFont := cityScreen.SmallFont
                    text := cityScreen.City.BuildingInfo.Name(building.Building)
                    if building.IsRubble {
                        text = "Destroyed " + text
                        useFont = cityScreen.RubbleFont
                    }

                    if building.Building == buildinglib.BuildingFortress {
                        text = fmt.Sprintf("%v's Fortress", cityScreen.Player.Wizard.Name)
                    }

                    useFont.PrintCenter(screen, float64(x + 10) + roadX, float64(y + 1) + roadY, 1, ebiten.ColorScale{}, text)
                }
            }
        }
    }

    drawName()

    river, err := cityScreen.ImageCache.GetImages("cityscap.lbx", 3)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(5, 100)
        index := animationCounter % uint64(len(river))
        screen.DrawImage(river[index], &options)
    }

    ui, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 6, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        screen.DrawImage(ui, &options)
    }

    cityScreen.BigFont.Print(screen, 20, 3, 1, ebiten.ColorScale{}, fmt.Sprintf("%v of %s", cityScreen.City.GetSize(), cityScreen.City.Name))

    cityScreen.DescriptionFont.Print(screen, 6, 19, 1, ebiten.ColorScale{}, fmt.Sprintf("%v", cityScreen.City.Race))

    deltaNumber := func(n int) string {
        if n > 0 {
            return fmt.Sprintf("+%v", n)
        } else if n == 0 {
            return "0"
        } else {
            return fmt.Sprintf("%v", n)
        }
    }

    cityScreen.DescriptionFont.PrintRight(screen, 210, 19, 1, ebiten.ColorScale{}, fmt.Sprintf("Population: %v (%v)", cityScreen.City.Population, deltaNumber(cityScreen.City.PopulationGrowthRate())))

    drawIcons := func(total int, small *ebiten.Image, large *ebiten.Image, x float64, y float64) float64 {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(x, y)

        for i := 0; i < total / 10; i++ {
            screen.DrawImage(large, &options)
            options.GeoM.Translate(float64(large.Bounds().Dx() + 1), 0)
        }

        for i := 0; i < total % 10; i++ {
            screen.DrawImage(small, &options)
            options.GeoM.Translate(float64(small.Bounds().Dx() + 1), 0)
        }

        endX, _ := options.GeoM.Apply(0, 0)
        return endX
    }

    foodRequired := cityScreen.City.RequiredFood()
    foodSurplus := cityScreen.City.SurplusFood()

    smallFood, _ := cityScreen.ImageCache.GetImage("backgrnd.lbx", 40, 0)
    bigFood, _ := cityScreen.ImageCache.GetImage("backgrnd.lbx", 88, 0)

    foodX := drawIcons(foodRequired, smallFood, bigFood, 6, 52)
    foodX += 5
    drawIcons(foodSurplus, smallFood, bigFood, foodX, 52)

    smallHammer, _ := cityScreen.ImageCache.GetImage("backgrnd.lbx", 41, 0)
    bigHammer, _ := cityScreen.ImageCache.GetImage("backgrnd.lbx", 89, 0)

    drawIcons(int(cityScreen.City.WorkProductionRate()), smallHammer, bigHammer, 6, 60)

    smallCoin, _ := cityScreen.ImageCache.GetImage("backgrnd.lbx", 42, 0)
    bigCoin, _ := cityScreen.ImageCache.GetImage("backgrnd.lbx", 90, 0)

    coinX := drawIcons(cityScreen.City.ComputeUpkeep(), smallCoin, bigCoin, 6, 68)
    drawIcons(cityScreen.City.GoldSurplus(), smallCoin, bigCoin, coinX + 6, 68)

    smallMagic, _ := cityScreen.ImageCache.GetImage("backgrnd.lbx", 43, 0)
    bigMagic, _ := cityScreen.ImageCache.GetImage("backgrnd.lbx", 91, 0)

    magicX := drawIcons(cityScreen.City.ManaCost(), smallMagic, bigMagic, 6, 76)
    drawIcons(cityScreen.City.ManaSurplus(), smallMagic, bigMagic, magicX + 6, 76)

    showWork := false
    workRequired := 0

    if cityScreen.City.ProducingBuilding != buildinglib.BuildingNone {
        producingPics, err := cityScreen.ImageCache.GetImages("cityscap.lbx", GetBuildingIndex(cityScreen.City.ProducingBuilding))
        if err == nil {
            index := animationCounter % uint64(len(producingPics))

            var options ebiten.DrawImageOptions
            options.GeoM.Translate(217, 144)
            screen.DrawImage(producingPics[index], &options)
        }

        cityScreen.ProducingFont.PrintCenter(screen, 237, 179, 1, ebiten.ColorScale{}, cityScreen.City.BuildingInfo.Name(cityScreen.City.ProducingBuilding))

        // for all buildings besides trade goods and housing, show amount of work required to build

        if cityScreen.City.ProducingBuilding == buildinglib.BuildingTradeGoods || cityScreen.City.ProducingBuilding == buildinglib.BuildingHousing {
            producingBackground, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 13, 0)
            if err == nil {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(260, 149)
                screen.DrawImage(producingBackground, &options)
            }

            description := ""
            switch cityScreen.City.ProducingBuilding {
                case buildinglib.BuildingTradeGoods: description = "Trade Goods"
                case buildinglib.BuildingHousing: description = "Increases population growth rate."
            }

            cityScreen.ProducingFont.PrintWrapCenter(screen, 285, 155, 60, 1, ebiten.ColorScale{}, description)
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
            combat.RenderCombatTile(screen, &cityScreen.ImageCache, options)
            combat.RenderCombatUnit(screen, use, options, cityScreen.City.ProducingUnit.Count)
            cityScreen.ProducingFont.PrintCenter(screen, 237, 179, 1, ebiten.ColorScale{}, cityScreen.City.ProducingUnit.Name)
        }

        showWork = true
        workRequired = cityScreen.City.ProducingUnit.ProductionCost
    }

    if showWork {
        turn := ""
        turns := (float64(workRequired) - float64(cityScreen.City.Production)) / float64(cityScreen.City.WorkProductionRate())
        if turns <= 0 {
            turn = "1 Turn"
        } else {
            turn = fmt.Sprintf("%v Turns", int(math.Ceil(turns)))
        }

        cityScreen.DescriptionFont.PrintRight(screen, 318, 140, 1, ebiten.ColorScale{}, turn)

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
                        screen.DrawImage(workFull, &options)
                    } else if leftOver > 0.05 {
                        screen.DrawImage(workEmpty, &options)
                        part := workFull.SubImage(image.Rect(0, 0, int(float64(workFull.Bounds().Dx()) * leftOver), workFull.Bounds().Dy())).(*ebiten.Image)
                        screen.DrawImage(part, &options)
                    }

                } else {
                    screen.DrawImage(workEmpty, &options)
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
    mapPart := screen.SubImage(image.Rect(mapX, mapY, mapX + mapWidth, mapY + mapHeight)).(*ebiten.Image)
    var mapGeom ebiten.GeoM
    mapGeom.Translate(float64(mapX), float64(mapY))
    mapView(mapPart, mapGeom, cityScreen.Counter)
    // FIXME: draw black translucent squares on the corner of the map to show the catchment area

    cityScreen.UI.Draw(cityScreen.UI, screen)

    if cityScreen.BuildScreen != nil {
        // screen.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 0x80})
        vector.DrawFilledRect(screen, 0, 0, float32(screen.Bounds().Dx()), float32(screen.Bounds().Dy()), color.RGBA{R: 0, G: 0, B: 0, A: 0x80}, true)
        cityScreen.BuildScreen.Draw(screen)
    }
}
