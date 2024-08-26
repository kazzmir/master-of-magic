package city

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
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

const (
    // not a real building, just something that shows up in the city view screen
    BuildingTree1 Building = iota + BuildingLast
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
    Building Building
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
    City *City

    UI *uilib.UI

    Buildings []BuildingSlot
    BuildScreen *BuildScreen

    Counter uint64
    State CityScreenState
}

type BuildingNativeSort []Building
func (b BuildingNativeSort) Len() int {
    return len(b)
}
func (b BuildingNativeSort) Less(i, j int) bool {
    return b[i] < b[j]
}
func (b BuildingNativeSort) Swap(i, j int) {
    b[i], b[j] = b[j], b[i]
}

func sortBuildings(buildings []Building) []Building {
    sort.Sort(BuildingNativeSort(buildings))
    return buildings
}

func hash(str string) uint64 {
    hasher := fnv.New64a()
    hasher.Write([]byte(str))
    return hasher.Sum64()
}

func MakeCityScreen(cache *lbx.LbxCache, city *City) *CityScreen {

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

        tree := []Building{BuildingTree1, BuildingTree2, BuildingTree3}[random.IntN(3)]

        buildings = append(buildings, BuildingSlot{Building: tree, Point: image.Pt(x, y)})
    }

    // FIXME: this is based on the population of the city
    maxHouses := random.IntN(15) + 20

    for i := 0; i < maxHouses; i++ {
        x := random.IntN(150) + 20
        y := random.IntN(60) + 10

        // house types are based on population size (village vs capital, etc)
        house := []Building{BuildingTreeHouse1, BuildingTreeHouse2, BuildingTreeHouse3, BuildingTreeHouse4, BuildingTreeHouse5}[random.IntN(5)]

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
        Buildings: buildings,
        State: CityScreenStateRunning,
    }

    cityScreen.UI = cityScreen.MakeUI()

    return cityScreen
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

    buildingView := image.Rect(5, 103, 208, 195)
    elements = append(elements, &uilib.UIElement{
        Rect: buildingView,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            // vector.StrokeRect(screen, float32(buildingView.Min.X), float32(buildingView.Min.Y), float32(buildingView.Dx()), float32(buildingView.Dy()), 1, color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff}, true)
        },
        Inside: func(element *uilib.UIElement, x int, y int){
            // log.Printf("inside building view: %v, %v", x, y)

            for _, slot := range cityScreen.Buildings {
                // if image.Pt(x, y).In(image.Rect(slot.Point.X + buildingView.Min.X, slot.Point.Y + buildingView.Min.Y, slot.Point.X + buildingView.Min.X + 32, slot.Point.Y + buildingView.Min.Y + 32)) {
                tx := (slot.Point.X + roadX - buildingView.Min.X) - x
                ty := (slot.Point.Y + roadY - buildingView.Min.Y) - y

                if tx * tx + ty * ty < 20 {
                    log.Printf("look at building %v", slot.Building)
                }
            }
        },
    })

    // FIXME: show disabled buy button if the item is not buyable (not enough money, or the item is trade goods/housing)
    // buy button
    // confirmation images are in resource.lbx
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
                    ui.RemoveElements(elements)
                }

                no := func(){
                    ui.RemoveElements(elements)
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

func GetBuildingIndex(building Building) int {
    switch building {
        case BuildingBarracks: return 45
        case BuildingArmory: return 46
        case BuildingFightersGuild: return 47
        case BuildingArmorersGuild: return 48
        case BuildingWarCollege: return 49
        case BuildingSmithy: return 50
        case BuildingStables: return 51
        case BuildingAnimistsGuild: return 52
        case BuildingFantasticStable: return 53
        case BuildingShipwrightsGuild: return 54
        case BuildingShipYard: return 55
        case BuildingMaritimeGuild: return 56
        case BuildingSawmill: return 57
        case BuildingLibrary: return 58
        case BuildingSagesGuild: return 59
        case BuildingOracle: return 60
        case BuildingAlchemistsGuild: return 61
        case BuildingUniversity: return 62
        case BuildingWizardsGuild: return 63
        case BuildingShrine: return 64
        case BuildingTemple: return 65
        case BuildingParthenon: return 66
        case BuildingCathedral: return 67
        case BuildingMarketplace: return 68
        case BuildingBank: return 69
        case BuildingMerchantsGuild: return 70
        case BuildingGranary: return 71
        case BuildingFarmersMarket: return 72
        case BuildingBuildersHall: return 73
        case BuildingMechaniciansGuild: return 74
        case BuildingMinersGuild: return 75
        case BuildingCityWalls: return 76
        case BuildingForestersGuild: return 78
        case BuildingWizardTower: return 40
        case BuildingSummoningCircle: return 6
        case BuildingTree1: return 19
        case BuildingTree2: return 20
        case BuildingTree3: return 21

        case BuildingTradeGoods: return 41

        // FIXME: housing is indices 42-44, based on the race of the city
        case BuildingHousing: return 43

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

    for _, building := range cityScreen.Buildings {

        index := GetBuildingIndex(building.Building)
        x, y := building.Point.X, building.Point.Y

        images, err := cityScreen.ImageCache.GetImages("cityscap.lbx", index)
        if err == nil {
            animationIndex := animationCounter % uint64(len(images))
            use := images[animationIndex]
            var options ebiten.DrawImageOptions
            // x,y position is the bottom left of the sprite
            options.GeoM.Translate(float64(x) + roadX, float64(y - use.Bounds().Dy()) + roadY)
            screen.DrawImage(use, &options)
        }
    }

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

    cityScreen.BigFont.Print(screen, 20, 3, 1, fmt.Sprintf("%v of %s", cityScreen.City.GetSize(), cityScreen.City.Name))

    cityScreen.DescriptionFont.Print(screen, 6, 19, 1, fmt.Sprintf("%v", cityScreen.City.Race))

    deltaNumber := func(n int) string {
        if n > 0 {
            return fmt.Sprintf("+%v", n)
        } else if n == 0 {
            return "0"
        } else {
            return fmt.Sprintf("-%v", n)
        }
    }

    cityScreen.DescriptionFont.PrintRight(screen, 210, 19, 1, fmt.Sprintf("Population: %v (%v)", cityScreen.City.Population, deltaNumber(80)))

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

    if cityScreen.City.ProducingBuilding != BuildingNone {
        producingPics, err := cityScreen.ImageCache.GetImages("cityscap.lbx", GetBuildingIndex(cityScreen.City.ProducingBuilding))
        if err == nil {
            index := animationCounter % uint64(len(producingPics))

            var options ebiten.DrawImageOptions
            options.GeoM.Translate(217, 144)
            screen.DrawImage(producingPics[index], &options)
        }

        cityScreen.ProducingFont.PrintCenter(screen, 237, 179, 1, fmt.Sprintf("%v", cityScreen.City.ProducingBuilding))

        // for all buildings besides trade goods and housing, show amount of work required to build

        if cityScreen.City.ProducingBuilding == BuildingTradeGoods || cityScreen.City.ProducingBuilding == BuildingHousing {
            producingBackground, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 13, 0)
            if err == nil {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(260, 149)
                screen.DrawImage(producingBackground, &options)
            }

            description := ""
            switch cityScreen.City.ProducingBuilding {
                case BuildingTradeGoods: description = "Trade Goods"
                case BuildingHousing: description = "Increases population growth rate."
            }

            cityScreen.ProducingFont.PrintWrapCenter(screen, 285, 155, 60, 1, description)
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
            cityScreen.ProducingFont.PrintCenter(screen, 237, 179, 1, cityScreen.City.ProducingUnit.Name)
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

        cityScreen.DescriptionFont.PrintRight(screen, 318, 140, 1, turn)

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
