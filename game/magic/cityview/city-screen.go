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
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

const (
    // not a real building, just something that shows up in the city view screen
    BuildingTree1 citylib.Building = iota + citylib.BuildingLast
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
    Building citylib.Building
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
    BuildingLook citylib.Building

    Counter uint64
    State CityScreenState
}

type BuildingNativeSort []citylib.Building
func (b BuildingNativeSort) Len() int {
    return len(b)
}
func (b BuildingNativeSort) Less(i, j int) bool {
    return b[i] < b[j]
}
func (b BuildingNativeSort) Swap(i, j int) {
    b[i], b[j] = b[j], b[i]
}

func sortBuildings(buildings []citylib.Building) []citylib.Building {
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

    fonts, err := fontLbx.ReadFonts(0)
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

        tree := []citylib.Building{BuildingTree1, BuildingTree2, BuildingTree3}[random.IntN(3)]

        buildings = append(buildings, BuildingSlot{Building: tree, Point: image.Pt(x, y)})
    }

    // FIXME: this is based on the population of the city
    maxHouses := random.IntN(15) + 20

    for i := 0; i < maxHouses; i++ {
        x := random.IntN(150) + 20
        y := random.IntN(60) + 10

        // house types are based on population size (village vs capital, etc)
        house := []citylib.Building{BuildingTreeHouse1, BuildingTreeHouse2, BuildingTreeHouse3, BuildingTreeHouse4, BuildingTreeHouse5}[random.IntN(5)]

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

func canSellBuilding(building citylib.Building) bool {
    return building.ProductionCost() > 0
}

func sellAmount(building citylib.Building) int {
    cost := building.ProductionCost() / 3
    if building == citylib.BuildingCityWalls {
        cost /= 2
    }

    if cost < 1 {
        cost = 1
    }

    return cost
}

func (cityScreen *CityScreen) SellBuilding(building citylib.Building) {
    // convert the building pic to one of the rubble ones
    // give player back the gold for the building
    // remove building from the city

    cityScreen.City.SoldBuilding = true

    cityScreen.Player.Gold += sellAmount(building)

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
            if cityScreen.BuildingLook != citylib.BuildingNone {
                helpEntries := help.GetEntriesByName(cityScreen.BuildingLook.String())
                if helpEntries != nil {
                    ui.AddElement(uilib.MakeHelpElement(ui, cityScreen.LbxCache, &cityScreen.ImageCache, helpEntries[0]))
                }
            }
        },
        LeftClick: func(element *uilib.UIElement) {
            if cityScreen.BuildingLook != citylib.BuildingNone && canSellBuilding(cityScreen.BuildingLook) {
                if cityScreen.City.SoldBuilding {
                    ui.AddElement(uilib.MakeErrorElement(cityScreen.UI, cityScreen.LbxCache, &cityScreen.ImageCache, "You can only sell back one building per turn."))
                } else {
                    var confirmElements []*uilib.UIElement

                    yes := func(){
                        cityScreen.SellBuilding(cityScreen.BuildingLook)
                    }

                    no := func(){
                    }

                    confirmElements = uilib.MakeConfirmDialog(cityScreen.UI, cityScreen.LbxCache, &cityScreen.ImageCache, fmt.Sprintf("Are you sure you want to sell back the %v for %v gold?", cityScreen.BuildingLook, sellAmount(cityScreen.BuildingLook)), yes, no)
                    ui.AddElements(confirmElements)
                }
            }
        },
        // if the user hovers over a building then show the name of the building
        Inside: func(element *uilib.UIElement, x int, y int){
            cityScreen.BuildingLook = citylib.BuildingNone
            // log.Printf("inside building view: %v, %v", x, y)

            // go in reverse order so we select the one in front first
            for i := len(cityScreen.Buildings) - 1; i >= 0; i-- {
                slot := cityScreen.Buildings[i]

                if slot.Building.String() == "?" || slot.Building.String() == "" {
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
    ui.SetElementsFromArray(elements)

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

func GetBuildingIndex(building citylib.Building) int {
    switch building {
        case citylib.BuildingBarracks: return 45
        case citylib.BuildingArmory: return 46
        case citylib.BuildingFightersGuild: return 47
        case citylib.BuildingArmorersGuild: return 48
        case citylib.BuildingWarCollege: return 49
        case citylib.BuildingSmithy: return 50
        case citylib.BuildingStables: return 51
        case citylib.BuildingAnimistsGuild: return 52
        case citylib.BuildingFantasticStable: return 53
        case citylib.BuildingShipwrightsGuild: return 54
        case citylib.BuildingShipYard: return 55
        case citylib.BuildingMaritimeGuild: return 56
        case citylib.BuildingSawmill: return 57
        case citylib.BuildingLibrary: return 58
        case citylib.BuildingSagesGuild: return 59
        case citylib.BuildingOracle: return 60
        case citylib.BuildingAlchemistsGuild: return 61
        case citylib.BuildingUniversity: return 62
        case citylib.BuildingWizardsGuild: return 63
        case citylib.BuildingShrine: return 64
        case citylib.BuildingTemple: return 65
        case citylib.BuildingParthenon: return 66
        case citylib.BuildingCathedral: return 67
        case citylib.BuildingMarketplace: return 68
        case citylib.BuildingBank: return 69
        case citylib.BuildingMerchantsGuild: return 70
        case citylib.BuildingGranary: return 71
        case citylib.BuildingFarmersMarket: return 72
        case citylib.BuildingBuildersHall: return 73
        case citylib.BuildingMechaniciansGuild: return 74
        case citylib.BuildingMinersGuild: return 75
        case citylib.BuildingCityWalls: return 76
        case citylib.BuildingForestersGuild: return 78
        case citylib.BuildingFortress: return 40
        case citylib.BuildingSummoningCircle: return 6
        case BuildingTree1: return 19
        case BuildingTree2: return 20
        case BuildingTree3: return 21

        case citylib.BuildingTradeGoods: return 41

        // FIXME: housing is indices 42-44, based on the race of the city
        case citylib.BuildingHousing: return 43

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
                    text := building.Building.String()
                    if building.IsRubble {
                        text = "Destroyed " + text
                        useFont = cityScreen.RubbleFont
                    }

                    if building.Building == citylib.BuildingFortress {
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
            return fmt.Sprintf("-%v", n)
        }
    }

    cityScreen.DescriptionFont.PrintRight(screen, 210, 19, 1, ebiten.ColorScale{}, fmt.Sprintf("Population: %v (%v)", cityScreen.City.Population, deltaNumber(80)))

    smallFood, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 40, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(6, 52)
        for i := 0; i < cityScreen.City.FoodProductionRate; i++ {
            screen.DrawImage(smallFood, &options)
            options.GeoM.Translate(float64(smallFood.Bounds().Dx() + 1), 0)
        }
    }

    // big food is 88
    // hammer is 41
    smallWork, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 41, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(6, 60)
        for i := 0; i < cityScreen.City.WorkProductionRate; i++ {
            screen.DrawImage(smallWork, &options)
            options.GeoM.Translate(float64(smallWork.Bounds().Dx() + 1), 0)
        }
    }

    // big hammer is 89
    // coin is 42
    smallCoin, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 42, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(6, 68)
        for i := 0; i < cityScreen.City.MoneyProductionRate; i++ {
            screen.DrawImage(smallCoin, &options)
            options.GeoM.Translate(float64(smallCoin.Bounds().Dx() + 1), 0)
        }
    }

    // big coin is 90
    // small magic is 43
    smallMagic, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 43, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(6, 76)
        for i := 0; i < cityScreen.City.MagicProductionRate; i++ {
            screen.DrawImage(smallMagic, &options)
            options.GeoM.Translate(float64(smallMagic.Bounds().Dx() + 1), 0)
        }
    }
    // big magic is 91

    citizenX := 6

    requiredFarmers := cityScreen.City.ComputeSubsistenceFarmers()

    // FIXME: add gap between required farmers and extra workers
    farmer, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", getRaceFarmerIndex(cityScreen.City.Race), 0)
    if err == nil {
        i := 0
        for i = 0; i < requiredFarmers && i < cityScreen.City.Farmers; i++ {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(citizenX), 27)
            screen.DrawImage(farmer, &options)
            citizenX += farmer.Bounds().Dx()
        }

        citizenX += 3

        for ; i < cityScreen.City.Farmers; i++ {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(citizenX), 27)
            screen.DrawImage(farmer, &options)
            citizenX += farmer.Bounds().Dx()
        }
    }

    worker, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", getRaceWorkerIndex(cityScreen.City.Race), 0)
    if err == nil {
        for i := 0; i < cityScreen.City.Workers; i++ {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(citizenX), 27)
            screen.DrawImage(worker, &options)
            citizenX += worker.Bounds().Dx()
        }
    }

    showWork := false
    workRequired := 0

    if cityScreen.City.ProducingBuilding != citylib.BuildingNone {
        producingPics, err := cityScreen.ImageCache.GetImages("cityscap.lbx", GetBuildingIndex(cityScreen.City.ProducingBuilding))
        if err == nil {
            index := animationCounter % uint64(len(producingPics))

            var options ebiten.DrawImageOptions
            options.GeoM.Translate(217, 144)
            screen.DrawImage(producingPics[index], &options)
        }

        cityScreen.ProducingFont.PrintCenter(screen, 237, 179, 1, ebiten.ColorScale{}, fmt.Sprintf("%v", cityScreen.City.ProducingBuilding))

        // for all buildings besides trade goods and housing, show amount of work required to build

        if cityScreen.City.ProducingBuilding == citylib.BuildingTradeGoods || cityScreen.City.ProducingBuilding == citylib.BuildingHousing {
            producingBackground, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 13, 0)
            if err == nil {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(260, 149)
                screen.DrawImage(producingBackground, &options)
            }

            description := ""
            switch cityScreen.City.ProducingBuilding {
                case citylib.BuildingTradeGoods: description = "Trade Goods"
                case citylib.BuildingHousing: description = "Increases population growth rate."
            }

            cityScreen.ProducingFont.PrintWrapCenter(screen, 285, 155, 60, 1, ebiten.ColorScale{}, description)
        } else {
            showWork = true
            workRequired = cityScreen.City.ProducingBuilding.ProductionCost()
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
        turns := float64(workRequired - cityScreen.City.Production) / float64(cityScreen.City.WorkProductionRate)
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
