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
    "github.com/kazzmir/master-of-magic/lib/coroutine"
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
    "github.com/hajimehoshi/ebiten/v2/inpututil"
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

// represents how much of a resource is being used/produced, such as '2 granary' for 'the granary produces 2 food'
type ResourceUsage struct {
    Count int // can be negative
    Name string
}

type CityScreenState int

const (
    CityScreenStateRunning CityScreenState = iota
    CityScreenStateDone
)

type Fonts struct {
    BigFont *font.Font
    DescriptionFont *font.Font
    ProducingFont *font.Font
    SmallFont *font.Font
    RubbleFont *font.Font
    CastFont *font.Font
    BannerFonts map[data.BannerType]*font.Font
}

type CityScreen struct {
    LbxCache *lbx.LbxCache
    ImageCache util.ImageCache
    Fonts *Fonts
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

func makeFonts(cache *lbx.LbxCache) (*Fonts, error) {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        return nil, err
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil, err
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

    // FIXME: this palette isn't exactly right. It should be a yellow-orange fade. Probably it exists somewhere else in the codebase
    yellow := color.RGBA{R: 0xef, G: 0xce, B: 0x4e, A: 0xff}
    fadePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        util.RotateHue(yellow, -0.6),
        // color.RGBA{R: 0xd5, G: 0x88, B: 0x25, A: 0xff},
        util.RotateHue(yellow, -0.3),
        util.RotateHue(yellow, -0.1),
        yellow,
    }

    castFont := font.MakeOptimizedFontWithPalette(fonts[4], fadePalette)

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

    makeBannerPalette := func(banner data.BannerType) color.Palette {
        var bannerColor color.RGBA

        switch banner {
            case data.BannerBlue: bannerColor = color.RGBA{R: 0x00, G: 0x00, B: 0xff, A: 0xff}
            case data.BannerGreen: bannerColor = color.RGBA{R: 0x00, G: 0xf0, B: 0x00, A: 0xff}
            case data.BannerPurple: bannerColor = color.RGBA{R: 0x8f, G: 0x30, B: 0xff, A: 0xff}
            case data.BannerRed: bannerColor = color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}
            case data.BannerYellow: bannerColor = color.RGBA{R: 0xff, G: 0xff, B: 0x00, A: 0xff}
        }

        return color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0x0},
            color.RGBA{R: 0, G: 0, B: 0, A: 0x0},
            bannerColor, bannerColor, bannerColor,
            bannerColor, bannerColor, bannerColor,
            bannerColor, bannerColor, bannerColor,
        }
    }

    bannerFonts := make(map[data.BannerType]*font.Font)

    for _, banner := range []data.BannerType{data.BannerGreen, data.BannerBlue, data.BannerRed, data.BannerPurple, data.BannerYellow} {
        bannerFonts[banner] = font.MakeOptimizedFontWithPalette(fonts[0], makeBannerPalette(banner))
    }

    return &Fonts{
        BigFont: bigFont,
        DescriptionFont: descriptionFont,
        ProducingFont: producingFont,
        SmallFont: smallFont,
        RubbleFont: rubbleFont,
        BannerFonts: bannerFonts,
        CastFont: castFont,
    }, nil
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

func makeBuildingSlots(city *citylib.City) []BuildingSlot {
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

    fonts, err := makeFonts(cache)
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

func makeCityScapeElement(cache *lbx.LbxCache, ui *uilib.UI, city *citylib.City, help *lbx.Help, imageCache *util.ImageCache, doSell func(buildinglib.Building), buildings []BuildingSlot, newBuilding buildinglib.Building, x1 int, y1 int, fonts *Fonts, player *playerlib.Player, getAlpha *util.AlphaFadeFunc) *uilib.UIElement {
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

        rawImageCache[index] = images[0]

        return images[0], nil
    }

    roadX := 0.0
    roadY := 18.0

    buildingLook := buildinglib.BuildingNone
    buildingView := image.Rect(x1, y1, x1 + 208, y1 + 195)
    element := &uilib.UIElement{
        Rect: buildingView,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            var geom ebiten.GeoM
            geom.Translate(float64(x1), float64(y1))
            drawCityScape(screen, buildings, buildingLook, newBuilding, ui.Counter / 8, imageCache, fonts, city.BuildingInfo, player, city.Enchantments.Values(), geom, (*getAlpha)())
            // vector.StrokeRect(screen, float32(buildingView.Min.X), float32(buildingView.Min.Y), float32(buildingView.Dx()), float32(buildingView.Dy()), 1, color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff}, true)
        },
        RightClick: func(element *uilib.UIElement) {
            if buildingLook != buildinglib.BuildingNone {
                helpEntries := help.GetEntriesByName(city.BuildingInfo.Name(buildingLook))
                if helpEntries != nil {
                    ui.AddElement(uilib.MakeHelpElement(ui, cache, imageCache, helpEntries[0]))
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
            buildingLook = buildinglib.BuildingNone
            // log.Printf("inside building view: %v, %v", x, y)

            // go in reverse order so we select the one in front first
            for i := len(buildings) - 1; i >= 0; i-- {
                slot := buildings[i]

                buildingName := city.BuildingInfo.Name(slot.Building)

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

                    // do pixel perfect detection
                    if image.Pt(x, y).In(image.Rect(x1, y1, x2, y2)) {
                        pixelX := x - x1
                        pixelY := y - y1

                        _, _, _, a := pic.At(pixelX, pixelY).RGBA()
                        if a > 0 {
                            buildingLook = slot.Building
                            // log.Printf("look at building %v (%v,%v) in (%v,%v,%v,%v)", slot.Building, useX, useY, x1, y1, x2, y2)
                            break
                        }
                    }
                }
            }
        },
    }

    return element
}

func (cityScreen *CityScreen) MakeUI(newBuilding buildinglib.Building) *uilib.UI {
    ui := &uilib.UI{
        Cache: cityScreen.LbxCache,
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

    sellBuilding := func (toSell buildinglib.Building) {
        if cityScreen.City.SoldBuilding {
            ui.AddElement(uilib.MakeErrorElement(ui, cityScreen.LbxCache, &cityScreen.ImageCache, "You can only sell back one building per turn.", func(){}))
        } else {
            var confirmElements []*uilib.UIElement

            yes := func(){
                cityScreen.SellBuilding(toSell)
            }

            no := func(){
            }

            confirmElements = uilib.MakeConfirmDialog(ui, cityScreen.LbxCache, &cityScreen.ImageCache, fmt.Sprintf("Are you sure you want to sell back the %v for %v gold?", cityScreen.City.BuildingInfo.Name(toSell), sellAmount(cityScreen.City, toSell)), yes, no)
            ui.AddElements(confirmElements)
        }
    }

    var getAlpha util.AlphaFadeFunc = func() float32 { return 1 }

    elements = append(elements, makeCityScapeElement(cityScreen.LbxCache, ui, cityScreen.City, &help, &cityScreen.ImageCache, sellBuilding, cityScreen.Buildings, newBuilding, 4, 102, cityScreen.Fonts, cityScreen.Player, &getAlpha))

    // FIXME: show disabled buy button if the item is not buyable (not enough money, or the item is trade goods/housing)
    // buy button
    buyButton, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 7, 0)
    if err == nil {
        buyX := 214
        buyY := 188
        elements = append(elements, &uilib.UIElement{
            Rect: image.Rect(buyX, buyY, buyX + buyButton.Bounds().Dx(), buyY + buyButton.Bounds().Dy()),
            PlaySoundLeftClick: true,
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
            PlaySoundLeftClick: true,
            LeftClick: func(element *uilib.UIElement) {
                if cityScreen.BuildScreen == nil {
                    cityScreen.BuildScreen = MakeBuildScreen(cityScreen.LbxCache, cityScreen.City)
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
            PlaySoundLeftClick: true,
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

    // FIXME: what happens where there are too many enchantments such that the text goes beyond the enchantment ui box?
    for i, enchantment := range slices.SortedFunc(slices.Values(cityScreen.City.Enchantments.Values()), func (a citylib.Enchantment, b citylib.Enchantment) int {
        return cmp.Compare(a.Enchantment.Name(), b.Enchantment.Name())
    }) {
        useFont := cityScreen.Fonts.BannerFonts[enchantment.Owner]
        x := 140
        y := 51 + i * useFont.Height()
        rect := image.Rect(x, y, x + int(useFont.MeasureTextWidth(enchantment.Enchantment.Name(), 1)), y + useFont.Height())
        elements = append(elements, &uilib.UIElement{
            Rect: rect,
            LeftClickRelease: func(element *uilib.UIElement) {
                if enchantment.Owner == cityScreen.Player.GetBanner() {
                    yes := func(){
                        cityScreen.City.RemoveEnchantment(enchantment.Enchantment, enchantment.Owner)
                        cityScreen.UI = cityScreen.MakeUI(buildinglib.BuildingNone)
                    }

                    no := func(){
                    }

                    confirmElements := uilib.MakeConfirmDialog(ui, cityScreen.LbxCache, &cityScreen.ImageCache, fmt.Sprintf("Do you wish to turn off the %v spell?", enchantment.Enchantment.Name()), yes, no)
                    ui.AddElements(confirmElements)
                } else {
                    ui.AddElement(uilib.MakeErrorElement(ui, cityScreen.LbxCache, &cityScreen.ImageCache, "You cannot cancel another wizard's enchantment.", func(){}))
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                useFont.Print(screen, float64(rect.Min.X), float64(rect.Min.Y), 1, ebiten.ColorScale{}, enchantment.Enchantment.Name())
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
                citizenX += 3
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

            resetResourceIcons()
        }
    } else {
        setupWorkers = func(){
        }
    }

    ui.SetElementsFromArray(elements)

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
                        ui.AddElements(unitview.MakeUnitContextMenu(cityScreen.LbxCache, ui, useUnit, disband))
                    },
                    Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                        var options colorm.DrawImageOptions
                        var matrix colorm.ColorM
                        options.GeoM.Translate(float64(posX), float64(posY))
                        colorm.DrawImage(screen, garrisonBackground, matrix, &options)
                        options.GeoM.Translate(1, 1)

                        // draw in grey scale if the unit is on patrol
                        if useUnit.GetBusy() == units.BusyStatusPatrol {
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

func drawCityScape(screen *ebiten.Image, buildings []BuildingSlot, buildingLook buildinglib.Building, newBuilding buildinglib.Building, animationCounter uint64, imageCache *util.ImageCache, fonts *Fonts, buildingInfo buildinglib.BuildingInfos, player *playerlib.Player, enchantments []citylib.Enchantment, baseGeoM ebiten.GeoM, alphaScale float32) {
    // 5 is grasslands
    // FIXME: make the land type and sky configurable
    landBackground, err := imageCache.GetImage("cityscap.lbx", 0, 4)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.ColorScale.ScaleAlpha(alphaScale)
        options.GeoM = baseGeoM
        screen.DrawImage(landBackground, &options)
    }

    hills1, err := imageCache.GetImage("cityscap.lbx", 7, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.ColorScale.ScaleAlpha(alphaScale)
        options.GeoM = baseGeoM
        options.GeoM.Translate(0, -1)
        screen.DrawImage(hills1, &options)
    }

    roadX := 0.0
    roadY := 18.0

    normalRoad, err := imageCache.GetImage("cityscap.lbx", 5, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.ColorScale.ScaleAlpha(alphaScale)
        options.GeoM = baseGeoM
        options.GeoM.Translate(roadX, roadY)
        screen.DrawImage(normalRoad, &options)
    }

    drawName := func(){
    }

    for _, building := range buildings {

        index := GetBuildingIndex(building.Building)

        if building.IsRubble {
            index = 105 + building.RubbleIndex
        }

        x, y := building.Point.X, building.Point.Y

        images, err := imageCache.GetImages("cityscap.lbx", index)
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
            // x,y position is the bottom left of the sprite
            options.GeoM.Translate(float64(x) + roadX, float64(y - use.Bounds().Dy()) + roadY)
            screen.DrawImage(use, &options)

            if buildingLook == building.Building {
                drawName = func(){
                    useFont := fonts.SmallFont
                    text := buildingInfo.Name(building.Building)
                    if building.IsRubble {
                        text = "Destroyed " + text
                        useFont = fonts.RubbleFont
                    }

                    if building.Building == buildinglib.BuildingFortress {
                        text = fmt.Sprintf("%v's Fortress", player.Wizard.Name)
                    }

                    printX, printY := baseGeoM.Apply(float64(x + 10) + roadX, float64(y + 1) + roadY)

                    useFont.PrintCenter(screen, printX, printY, 1, options.ColorScale, text)
                }
            }
        }
    }

    // FIXME: make this configurable
    river, err := imageCache.GetImages("cityscap.lbx", 3)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.ColorScale.ScaleAlpha(alphaScale)
        options.GeoM = baseGeoM
        options.GeoM.Translate(1, -2)
        index := animationCounter % uint64(len(river))
        screen.DrawImage(river[index], &options)
    }

    for _, enchantment := range slices.SortedFunc(slices.Values(enchantments), func (a citylib.Enchantment, b citylib.Enchantment) int {
        return cmp.Compare(a.Enchantment.Name(), b.Enchantment.Name())
    }) {
        switch enchantment.Enchantment {
            case data.CityEnchantmentWallOfFire:
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(alphaScale)
                options.GeoM = baseGeoM
                options.GeoM.Translate(0, 85)
                images, _ := imageCache.GetImages("cityscap.lbx", 77)
                index := animationCounter % uint64(len(images))
                screen.DrawImage(images[index], &options)
            case data.CityEnchantmentWallOfDarkness:
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(alphaScale)
                options.GeoM = baseGeoM
                options.GeoM.Translate(0, 85)
                images, _ := imageCache.GetImages("cityscap.lbx", 79)
                index := animationCounter % uint64(len(images))
                screen.DrawImage(images[index], &options)
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
            })
        }
    }

    return usages
}

// pass screen=nil to compute how wide the icons are without drawing them
func (cityScreen *CityScreen) drawIcons(total int, small *ebiten.Image, large *ebiten.Image, options ebiten.DrawImageOptions, screen *ebiten.Image) ebiten.GeoM {
    largeGap := large.Bounds().Dx()

    if total / 10 > 3 {
        largeGap -= 1
    }

    if total / 10 > 6 {
        largeGap -= 1
    }

    for range total / 10 {
        if screen != nil {
            screen.DrawImage(large, &options)
        }
        options.GeoM.Translate(float64(largeGap), 0)
    }

    smallGap := small.Bounds().Dx() + 1
    if total % 10 > 3 {
        smallGap -= 1
    }
    if total % 10 >= 6 {
        smallGap -= 1
    }

    for range total % 10 {
        if screen != nil {
            screen.DrawImage(small, &options)
        }
        options.GeoM.Translate(float64(smallGap), 0)
    }

    return options.GeoM
}

// copied heavily from ui/dialogs.go:MakeHelpElementWithLayer
func (cityScreen *CityScreen) MakeResourceDialog(title string, smallIcon *ebiten.Image, bigIcon *ebiten.Image, ui *uilib.UI, resources []ResourceUsage) *uilib.UIElement {
    helpTop, err := cityScreen.ImageCache.GetImage("help.lbx", 0, 0)
    if err != nil {
        return nil
    }

    fontLbx, err := cityScreen.LbxCache.GetLbxFile("fonts.lbx")
    if err != nil {
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        return nil
    }

    const fadeSpeed = 7

    getAlpha := ui.MakeFadeIn(fadeSpeed)

    helpPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0x5e, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
    }

    helpFont := font.MakeOptimizedFontWithPalette(fonts[1], helpPalette)

    titleRed := color.RGBA{R: 0x50, G: 0x00, B: 0x0e, A: 0xff}
    titlePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        titleRed,
        titleRed,
        titleRed,
        titleRed,
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
    }

    helpTitleFont := font.MakeOptimizedFontWithPalette(fonts[4], titlePalette)

    infoX := 55
    // infoY := 30
    infoWidth := helpTop.Bounds().Dx()
    // infoHeight := screen.HelpTop.Bounds().Dy()
    infoLeftMargin := 18
    infoTopMargin := 26
    infoBodyMargin := 3
    maxInfoWidth := infoWidth - infoLeftMargin - infoBodyMargin - 14

    // fmt.Printf("Help text: %v\n", []byte(help.Text))

    helpTextY := infoTopMargin
    titleYAdjust := 0
    helpTextY += helpTitleFont.Height() + 1

    textHeight := len(resources) * (helpFont.Height() + 1)

    bottom := helpTextY + textHeight

    // only draw as much of the top scroll as there are lines of text
    topImage := helpTop.SubImage(image.Rect(0, 0, helpTop.Bounds().Dx(), int(bottom))).(*ebiten.Image)
    helpBottom, err := cityScreen.ImageCache.GetImage("help.lbx", 1, 0)
    if err != nil {
        return nil
    }

    infoY := (data.ScreenHeight - bottom - helpBottom.Bounds().Dy()) / 2

    widestResources := float64(0)
    for _, usage := range resources {
        var options ebiten.DrawImageOptions
        geom := cityScreen.drawIcons(int(math.Abs(float64(usage.Count))), smallIcon, bigIcon, options, nil)
        x, _ := geom.Apply(0, 0)
        if x > widestResources {
            widestResources = x
        }
    }

    infoElement := &uilib.UIElement{
        // Rect: image.Rect(infoX, infoY, infoX + infoWidth, infoY + infoHeight),
        Rect: image.Rect(0, 0, data.ScreenWidth, data.ScreenHeight),
        Draw: func (infoThis *uilib.UIElement, window *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(infoX), float64(infoY))
            options.ColorScale.ScaleAlpha(getAlpha())
            window.DrawImage(topImage, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(infoX), float64(bottom + infoY))
            options.ColorScale.ScaleAlpha(getAlpha())
            window.DrawImage(helpBottom, &options)

            // for debugging
            // vector.StrokeRect(window, float32(infoX), float32(infoY), float32(infoWidth), float32(infoHeight), 1, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, true)
            // vector.StrokeRect(window, float32(infoX + infoLeftMargin), float32(infoY + infoTopMargin), float32(maxInfoWidth), float32(screen.HelpTitleFont.Height() + 20 + 1), 1, color.RGBA{R: 0, G: 0xff, B: 0, A: 0xff}, false)

            titleX := infoX + infoLeftMargin + maxInfoWidth / 2

            helpTitleFont.PrintCenter(window, float64(titleX), float64(infoY + infoTopMargin + titleYAdjust), 1, options.ColorScale, title)

            yPos := infoY + infoTopMargin + helpTitleFont.Height() + 1
            xPos := infoX + infoLeftMargin

            options.GeoM.Reset()
            options.GeoM.Translate(float64(xPos), float64(yPos))

            for _, usage := range resources {
                if usage.Count < 0 {
                    x, y := options.GeoM.Apply(0, 1)
                    helpFont.PrintRight(window, x, y, 1, options.ColorScale, "-")
                }

                cityScreen.drawIcons(int(math.Abs(float64(usage.Count))), smallIcon, bigIcon, options, window)

                x, y := options.GeoM.Apply(widestResources + 5, 0)

                helpFont.Print(window, x, y, 1, options.ColorScale, fmt.Sprintf("%v (%v)", usage.Name, usage.Count))
                yPos += helpFont.Height() + 1
                options.GeoM.Translate(0, float64(helpFont.Height() + 1))
            }

        },
        LeftClick: func(infoThis *uilib.UIElement){
            getAlpha = ui.MakeFadeOut(fadeSpeed)
            ui.AddDelay(fadeSpeed, func(){
                ui.RemoveElement(infoThis)
            })
        },
        Layer: 1,
    }

    return infoElement
}

func (cityScreen *CityScreen) WorkProducers() []ResourceUsage {
    var usage []ResourceUsage

    add := func(count int, name string){
        if count > 0 {
            usage = append(usage, ResourceUsage{
                Count: count,
                Name: name,
            })
        }
    }

    add(int(cityScreen.City.Workers), "Workers")
    add(int(cityScreen.City.ProductionFarmers()), "Farmers")
    add(int(cityScreen.City.ProductionTerrain()), "Terrain")
    add(int(cityScreen.City.ProductionSawmill()), "Sawmill")
    add(int(cityScreen.City.ProductionForestersGuild()), "Forester's Guild")
    add(int(cityScreen.City.ProductionMinersGuild()), "Miner's Guild")
    add(int(cityScreen.City.ProductionMechaniciansGuild()), "Mechanician's Guild")

    return usage
}

func (cityScreen *CityScreen) BuildingMaintenanceResources() []ResourceUsage {
    var usage []ResourceUsage

    for _, building := range cityScreen.City.Buildings.Values() {
        if building == buildinglib.BuildingFortress || building == buildinglib.BuildingSummoningCircle {
            continue
        }

        maintenance := cityScreen.City.BuildingInfo.UpkeepCost(building)
        usage = append(usage, ResourceUsage{
            Count: maintenance,
            Name: cityScreen.City.BuildingInfo.Name(building),
        })
    }

    return usage
}

func (cityScreen *CityScreen) GoldProducers() []ResourceUsage {
    var usage []ResourceUsage

    add := func(count int, name string){
        if count > 0 {
            usage = append(usage, ResourceUsage{
                Count: count,
                Name: name,
            })
        }
    }

    // FIXME: add tiles (road/river/ocean)
    add(int(cityScreen.City.GoldTaxation()), "Taxes")
    add(int(cityScreen.City.GoldTradeGoods()), "Trade Goods")
    add(int(cityScreen.City.GoldMinerals()), "Minerals")
    add(int(cityScreen.City.GoldMarketplace()), "Marketplace")
    add(int(cityScreen.City.GoldBank()), "Bank")
    add(int(cityScreen.City.GoldMerchantsGuild()), "Merchant's Guild")

    return usage
}

func (cityScreen *CityScreen) PowerProducers() []ResourceUsage {
    var usage []ResourceUsage

    if int(cityScreen.City.PowerCitizens()) > 0 {
        usage = append(usage, ResourceUsage{
            Count: int(cityScreen.City.PowerCitizens()),
            Name: "Townsfolk",
        })
    }

    if cityScreen.City.Buildings.Contains(buildinglib.BuildingShrine) {
        usage = append(usage, ResourceUsage{
            Count: 1,
            Name: "Shrine",
        })
    }

    if cityScreen.City.Buildings.Contains(buildinglib.BuildingTemple) {
        usage = append(usage, ResourceUsage{
            Count: 2,
            Name: "Temple",
        })
    }

    if cityScreen.City.Buildings.Contains(buildinglib.BuildingParthenon) {
        usage = append(usage, ResourceUsage{
            Count: 3,
            Name: "Parthenon",
        })
    }

    if cityScreen.City.Buildings.Contains(buildinglib.BuildingCathedral) {
        usage = append(usage, ResourceUsage{
            Count: 4,
            Name: "Cathedral",
        })
    }

    if cityScreen.City.Buildings.Contains(buildinglib.BuildingAlchemistsGuild) {
        usage = append(usage, ResourceUsage{
            Count: 3,
            Name: "Alchemist's Guild",
        })
    }

    if cityScreen.City.Buildings.Contains(buildinglib.BuildingWizardsGuild) {
        usage = append(usage, ResourceUsage{
            Count: -3,
            Name: "Wizard's Guild",
        })
    }

    if cityScreen.City.Buildings.Contains(buildinglib.BuildingFortress) && cityScreen.City.Plane == data.PlaneMyrror {
        usage = append(usage, ResourceUsage{
            Count: 5,
            Name: "Fortress",
        })
    }

    if int(cityScreen.City.PowerMinerals()) > 0 {
        usage = append(usage, ResourceUsage{
            Count: int(cityScreen.City.PowerMinerals()),
            Name: "Minerals",
        })
    }

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
            })
        }
    }

    return usage
}

func (cityScreen *CityScreen) CreateResourceIcons(ui *uilib.UI) []*uilib.UIElement {
    foodRequired := cityScreen.City.RequiredFood()
    foodSurplus := cityScreen.City.SurplusFood()

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
            ui.AddElement(cityScreen.MakeResourceDialog("Food", smallFood, bigFood, ui, foodProducers))
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(foodRect.Min.X), float64(foodRect.Min.Y))
            options.GeoM = cityScreen.drawIcons(foodRequired, smallFood, bigFood, options, screen)
            options.GeoM.Translate(5, 0)
            cityScreen.drawIcons(foodSurplus, smallFood, bigFood, options, screen)
        },
    })

    production := cityScreen.City.WorkProductionRate()
    workRect := image.Rect(6, 60, 6 + 9 * bigHammer.Bounds().Dx(), 60 + bigHammer.Bounds().Dy())
    elements = append(elements, &uilib.UIElement{
        Rect: workRect,
        LeftClick: func(element *uilib.UIElement) {
            workProducers := cityScreen.WorkProducers()
            ui.AddElement(cityScreen.MakeResourceDialog("Production", smallHammer, bigHammer, ui, workProducers))
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
            ui.AddElement(cityScreen.MakeResourceDialog("Building Maintenance", smallCoin, bigCoin, ui, maintenance))
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
            ui.AddElement(cityScreen.MakeResourceDialog("Gold", smallCoin, bigCoin, ui, gold))
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
            ui.AddElement(cityScreen.MakeResourceDialog("Power", smallMagic, bigMagic, ui, power))
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
            ui.AddElement(cityScreen.MakeResourceDialog("Spell Research", smallResearch, bigResearch, ui, research))
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
        screen.DrawImage(ui, &options)
    }

    cityScreen.Fonts.BigFont.Print(screen, 20, 3, 1, ebiten.ColorScale{}, fmt.Sprintf("%v of %s", cityScreen.City.GetSize(), cityScreen.City.Name))
    cityScreen.Fonts.DescriptionFont.Print(screen, 6, 19, 1, ebiten.ColorScale{}, fmt.Sprintf("%v", cityScreen.City.Race))

    deltaNumber := func(n int) string {
        if n > 0 {
            return fmt.Sprintf("+%v", n)
        } else if n == 0 {
            return "0"
        } else {
            return fmt.Sprintf("%v", n)
        }
    }

    cityScreen.Fonts.DescriptionFont.PrintRight(screen, 210, 19, 1, ebiten.ColorScale{}, fmt.Sprintf("Population: %v (%v)", cityScreen.City.Population, deltaNumber(cityScreen.City.PopulationGrowthRate())))

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

        cityScreen.Fonts.ProducingFont.PrintCenter(screen, 237, 179, 1, ebiten.ColorScale{}, cityScreen.City.BuildingInfo.Name(cityScreen.City.ProducingBuilding))

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

            cityScreen.Fonts.ProducingFont.PrintWrapCenter(screen, 285, 155, 60, 1, ebiten.ColorScale{}, description)
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
            combat.RenderCombatUnit(screen, use, options, cityScreen.City.ProducingUnit.Count, data.UnitEnchantmentNone, 0, nil)
            cityScreen.Fonts.ProducingFont.PrintCenter(screen, 237, 179, 1, ebiten.ColorScale{}, cityScreen.City.ProducingUnit.Name)
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

        cityScreen.Fonts.DescriptionFont.PrintRight(screen, 318, 140, 1, ebiten.ColorScale{}, turn)

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
func SimplifiedView(cache *lbx.LbxCache, city *citylib.City, player *playerlib.Player) (func(coroutine.YieldFunc, func()), func(*ebiten.Image)) {
    imageCache := util.MakeImageCache(cache)

    fonts, err := makeFonts(cache)
    if err != nil {
        log.Printf("Could not make fonts: %v", err)
        return func(yield coroutine.YieldFunc, update func()){}, func(*ebiten.Image){}
    }

    helpLbx, err := cache.GetLbxFile("help.lbx")
    if err != nil {
        log.Printf("Error with help: %v", err)
        return func(yield coroutine.YieldFunc, update func()){}, func(*ebiten.Image){}
    }

    help, err := helpLbx.ReadHelp(2)
    if err != nil {
        log.Printf("Error with help: %v", err)
        return func(yield coroutine.YieldFunc, update func()){}, func(*ebiten.Image){}
    }

    buildings := makeBuildingSlots(city)

    background, _ := imageCache.GetImage("reload.lbx", 26, 0)
    var options ebiten.DrawImageOptions
    options.GeoM.Translate(float64(data.ScreenWidth / 2 - background.Bounds().Dx() / 2), 0)

    var getAlpha util.AlphaFadeFunc

    currentUnitName := ""

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            useOptions := options
            useOptions.ColorScale.ScaleAlpha(getAlpha())

            screen.DrawImage(background, &useOptions)

            titleX, titleY := options.GeoM.Apply(20, 3)
            fonts.BigFont.Print(screen, titleX, titleY, 1, useOptions.ColorScale, fmt.Sprintf("%v of %s", city.GetSize(), city.Name))
            raceX, raceY := options.GeoM.Apply(6, 19)
            fonts.DescriptionFont.Print(screen, raceX, raceY, 1, useOptions.ColorScale, fmt.Sprintf("%v", city.Race))

            unitsX, unitsY := options.GeoM.Apply(6, 43)
            fonts.DescriptionFont.Print(screen, unitsX, unitsY, 1, useOptions.ColorScale, fmt.Sprintf("Units   %v", currentUnitName))

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })
        },
    }

    getAlpha = ui.MakeFadeIn(7)

    ui.SetElementsFromArray(nil)

    x1, y1 := options.GeoM.Apply(5, 102)

    cityScapeElement := makeCityScapeElement(cache, ui, city, &help, &imageCache, func(buildinglib.Building){}, buildings, buildinglib.BuildingNone, int(x1), int(y1), fonts, player, &getAlpha)

    ui.AddElement(cityScapeElement)

    // draw all farmers/workers/rebels
    ui.AddElement(&uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            localOptions := options
            localOptions.ColorScale.ScaleAlpha(getAlpha())
            localOptions.GeoM.Translate(6, 27)
            farmer, _ := imageCache.GetImage("backgrnd.lbx", getRaceFarmerIndex(city.Race), 0)
            worker, _ := imageCache.GetImage("backgrnd.lbx", getRaceWorkerIndex(city.Race), 0)
            rebel, _ := imageCache.GetImage("backgrnd.lbx", getRaceRebelIndex(city.Race), 0)
            subsistenceFarmers := city.ComputeSubsistenceFarmers()

            for range subsistenceFarmers {
                screen.DrawImage(farmer, &localOptions)
                localOptions.GeoM.Translate(float64(farmer.Bounds().Dx()), 0)
            }

            localOptions.GeoM.Translate(3, 0)

            for range city.Farmers - subsistenceFarmers {
                screen.DrawImage(farmer, &localOptions)
                localOptions.GeoM.Translate(float64(farmer.Bounds().Dx()), 0)
            }

            for range city.Workers {
                screen.DrawImage(worker, &localOptions)
                localOptions.GeoM.Translate(float64(worker.Bounds().Dx()), 0)
            }

            localOptions.GeoM.Translate(3, -2)

            for range city.Rebels {
                screen.DrawImage(rebel, &localOptions)
                localOptions.GeoM.Translate(float64(rebel.Bounds().Dx()), 0)
            }
        },
    })

    stack := player.FindStack(city.X, city.Y, city.Plane)
    if stack != nil {
        inside := 0
        for i, unit := range stack.Units() {
            x, y := options.GeoM.Apply(8, 52)

            x += float64(i % 6) * 20
            y += float64(i / 6) * 20

            pic, _ := imageCache.GetImageTransform(unit.GetLbxFile(), unit.GetLbxIndex(), 0, unit.GetBanner().String(), units.MakeUpdateUnitColorsFunc(unit.GetBanner()))
            rect := image.Rect(int(x), int(y), int(x) + pic.Bounds().Dx(), int(y) + pic.Bounds().Dy())
            ui.AddElement(&uilib.UIElement{
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
                    screen.DrawImage(pic, &localOptions)
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
        x, y := options.GeoM.Apply(142, float64(51 + i * useFont.Height()))
        ui.AddElement(&uilib.UIElement{
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var scale ebiten.ColorScale
                scale.ScaleAlpha(getAlpha())
                useFont.Print(screen, x, y, 1, scale, enchantment.Enchantment.Name())
            },
        })
    }

    draw := func(screen *ebiten.Image){
        ui.Draw(ui, screen)
    }

    logic := func(yield coroutine.YieldFunc, update func()){
        countdown := 0
        for countdown != 1 {
            update()
            ui.StandardUpdate()

            if countdown == 0 && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
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
