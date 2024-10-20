package cityview

import (
    "log"
    "fmt"
    "image"
    "image/color"
    // "strings"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    // "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/game/magic/unitview"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/hajimehoshi/ebiten/v2"
    // "github.com/hajimehoshi/ebiten/v2/vector"
)

type BuildScreenState int
const (
    BuildScreenRunning BuildScreenState = iota
    BuildScreenCanceled
    BuildScreenOk
)

type BuildScreen struct {
    LbxCache *lbx.LbxCache
    ImageCache *util.ImageCache
    City *citylib.City
    UI *uilib.UI
    State BuildScreenState
    ProducingBuilding buildinglib.Building
    ProducingUnit units.Unit
}

func MakeBuildScreen(cache *lbx.LbxCache, city *citylib.City, producingBuilding buildinglib.Building, producingUnit units.Unit) *BuildScreen {
    imageCache := util.MakeImageCache(cache)

    var buildScreen *BuildScreen

    doCancel := func(){
        buildScreen.Cancel()
    }

    doOk := func(){
        buildScreen.Ok()
    }

    buildScreen = &BuildScreen{
        LbxCache: cache,
        ImageCache: &imageCache,
        City: city,
        State: BuildScreenRunning,
        ProducingBuilding: producingBuilding,
        ProducingUnit: producingUnit,
    }

    ui := makeBuildUI(cache, &imageCache, city, buildScreen, doCancel, doOk)

    buildScreen.UI = ui

    return buildScreen
}

func (buildScreen *BuildScreen) Cancel() {
    buildScreen.State = BuildScreenCanceled
}

func (buildScreen *BuildScreen) Ok() {
    buildScreen.State = BuildScreenOk
}

/* return the buildings that can be built, based on what the city already has
 */
func computePossibleBuildings(city *citylib.City) []buildinglib.Building {
    // FIXME: take race into account
    var possibleBuildings []buildinglib.Building

    for _, building := range buildinglib.Buildings() {
        if city.Buildings.Contains(building) {
            continue
        }

        canBuild := true
        for _, dependency := range city.BuildingInfo.Dependencies(building) {
            if !city.Buildings.Contains(dependency) {
                canBuild = false
            }
        }

        // FIXME: take terrain dependency into account

        if canBuild {
            possibleBuildings = append(possibleBuildings, building)
        }
    }

    return possibleBuildings
}

func combineStrings(all []string) string {
    if len(all) == 0 {
        return ""
    }

    if len(all) == 1 {
        return all[0]
    }

    if len(all) == 2 {
        return all[0] + " and " + all[1]
    }

    out := ""
    for i := 0; i < len(all) - 1; i++ {
        out += all[i] + ", "
    }

    out += "and " + all[len(all) - 1]
    return out
}

func getPossibleUnits(city *citylib.City) []units.Unit {
    var out []units.Unit
    for _, unit := range units.AllUnits {
        if unit.Race == data.RaceAll || unit.Race == city.Race {

            canBuild := true
            for _, building := range unit.RequiredBuildings {
                if !city.Buildings.Contains(building) {
                    canBuild = false
                }
            }

            if canBuild {
                out = append(out, unit)
            }
        }
    }

    return out
}

func makeBuildUI(cache *lbx.LbxCache, imageCache *util.ImageCache, city *citylib.City, buildScreen *BuildScreen, doCancel func(), doOk func()) *uilib.UI {

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

    helpLbx, err := cache.GetLbxFile("HELP.LBX")
    if err != nil {
        return nil
    }

    help, err := helpLbx.ReadHelp(2)
    if err != nil {
        return nil
    }

    buildDescriptions := buildinglib.MakeBuildDescriptions(cache)
    
    titleFont := font.MakeOptimizedFont(fonts[2])

    alphaWhite := util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 180})

    whitePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        alphaWhite, alphaWhite, alphaWhite,
    }

    titleFontWhite := font.MakeOptimizedFontWithPalette(fonts[2], whitePalette)

    descriptionPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 90}),
        util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}),
        util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 200}),
        util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 200}),
        util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 200}),
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
    }

    descriptionFont := font.MakeOptimizedFontWithPalette(fonts[4], descriptionPalette)

    yellowGradient := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
    }

    okCancelFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowGradient)

    smallFont := font.MakeOptimizedFontWithPalette(fonts[1], descriptionPalette)

    mediumFont := font.MakeOptimizedFontWithPalette(fonts[2], descriptionPalette)

    var elements []*uilib.UIElement

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image) {
            mainInfo, err := imageCache.GetImage("unitview.lbx", 0, 0)
            if err == nil {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(75, 0)
                screen.DrawImage(mainInfo, &options)
            }

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })

        },
    }

    mainElement := &uilib.UIElement{
    }

    elements = append(elements, mainElement)

    buildingInfo, err := imageCache.GetImage("unitview.lbx", 31, 0)

    var selectedElement *uilib.UIElement

    updateMainElementBuilding := func(building buildinglib.Building){
        descriptionWrapped := descriptionFont.CreateWrappedText(155, 1, buildDescriptions.Get(building))

        allowedBuildings := city.AllowedBuildings(building)
        allowedUnits := city.AllowedUnits(building)
        var allowStrings []string
        for _, allowed := range allowedBuildings {
            allowStrings = append(allowStrings, city.BuildingInfo.Name(allowed))
        }
        for _, allowed := range allowedUnits {
            allowStrings = append(allowStrings, allowed.Name)
        }

        allows := combineStrings(allowStrings)

        allowsWrapped := mediumFont.CreateWrappedText(100, 1, allows)

        ui.RemoveElement(mainElement)
        mainElement = &uilib.UIElement{
            Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                images, err := imageCache.GetImages("cityscap.lbx", GetBuildingIndex(building))
                if err == nil {
                    middleX := float64(104)
                    middleY := float64(20)
                    index := (ui.Counter / 7) % uint64(len(images))

                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(middleX, middleY)
                    options.GeoM.Translate(-float64(images[index].Bounds().Dx() / 2), -float64(images[index].Bounds().Dy() / 2))
                    screen.DrawImage(images[index], &options)

                    // vector.DrawFilledCircle(screen, float32(middleX), float32(middleY), 1, color.RGBA{255, 255, 255, 255}, true)
                }

                descriptionFont.Print(screen, 130, 12, 1, ebiten.ColorScale{}, city.BuildingInfo.Name(building))
                smallFont.Print(screen, 130, 33, 1, ebiten.ColorScale{}, fmt.Sprintf("Cost %v", city.BuildingInfo.ProductionCost(building)))

                descriptionFont.Print(screen, 85, 48, 1, ebiten.ColorScale{}, "Maintenance")

                buildingMaintenance := city.BuildingInfo.UpkeepCost(building)

                if buildingMaintenance == 0 {
                    mediumFont.Print(screen, 85 + descriptionFont.MeasureTextWidth("Maintenance", 1) + 4, 49, 1, ebiten.ColorScale{}, "0")
                } else {
                    smallCoin, err := imageCache.GetImage("backgrnd.lbx", 42, 0)
                    if err == nil {
                        var options ebiten.DrawImageOptions
                        options.GeoM.Translate(85 + descriptionFont.MeasureTextWidth("Maintenance", 1) + 3, 50)
                        for i := 0; i < buildingMaintenance; i++ {
                            screen.DrawImage(smallCoin, &options)
                            options.GeoM.Translate(float64(smallCoin.Bounds().Dx() + 2), 0)
                        }
                    }
                }

                descriptionFont.Print(screen, 85, 58, 1, ebiten.ColorScale{}, "Allows")

                mediumFont.RenderWrapped(screen, 85 + descriptionFont.MeasureTextWidth("Allows", 1) + 10, 59, allowsWrapped, ebiten.ColorScale{}, false)

                descriptionFont.RenderWrapped(screen, 85, 108, descriptionWrapped, ebiten.ColorScale{}, false)

                /*
                helpEntries := help.GetEntriesByName(building.String())
                if helpEntries != nil {
                    entry := helpEntries[0].Text
                    splitIndex := strings.IndexRune(entry, 0x14)
                    if splitIndex != -1 {
                        entry = entry[splitIndex+1:]
                    }
                    descriptionFont.Print(screen, 85, 90, 1, entry)
                }
                */
            },
        }
        ui.AddElement(mainElement)
    }

    updateMainElementUnit := func(unit units.Unit){
        ui.RemoveElement(mainElement)
        bannerUnit := units.MakeOverworldUnitFromUnit(unit, 0, 0, city.Plane, city.Banner, nil)
        mainElement = &uilib.UIElement{
            Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(104, 28)
                unitview.RenderCombatImage(screen, imageCache, bannerUnit, options)

                options.GeoM.Reset()
                options.GeoM.Translate(130, 7)
                unitview.RenderUnitInfoBuild(screen, imageCache, bannerUnit, descriptionFont, smallFont, options)

                options.GeoM.Reset()
                options.GeoM.Translate(85, 48)
                unitview.RenderUnitInfoStats(screen, imageCache, bannerUnit, 10, descriptionFont, smallFont, options)

                options.GeoM.Reset()
                options.GeoM.Translate(85, 108)
                unitview.RenderUnitAbilities(screen, imageCache, bannerUnit, mediumFont, options)
            },
        }
        ui.AddElement(mainElement)
    }

    if err == nil {
        possibleBuildings := computePossibleBuildings(city)
        for i, building := range possibleBuildings {

            x1 := 0
            y1 := 4 + i * (buildingInfo.Bounds().Dy() + 1)
            x2 := x1 + buildingInfo.Bounds().Dx()
            y2 := y1 + buildingInfo.Bounds().Dy()

            element := &uilib.UIElement{
                Rect: image.Rect(x1, y1, x2, y2),
                DoubleLeftClick: func(this *uilib.UIElement) {
                    doOk()
                },
                LeftClick: func(this *uilib.UIElement) {
                    selectedElement = this
                    buildScreen.ProducingBuilding = building
                    buildScreen.ProducingUnit = units.UnitNone
                    updateMainElementBuilding(building)
                },
                RightClick: func(this *uilib.UIElement) {
                    helpEntries := help.GetEntriesByName("Building Options")
                    if helpEntries != nil {
                        ui.AddElement(uilib.MakeHelpElement(ui, cache, imageCache, helpEntries[0]))
                    }
                },
                Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(x1), float64(y1))
                    screen.DrawImage(buildingInfo, &options)

                    use := titleFont

                    // show highlight if element was clicked on
                    if selectedElement == this {
                        use = titleFontWhite
                    }

                    use.Print(screen, float64(x1 + 2), float64(y1 + 1), 1, ebiten.ColorScale{}, city.BuildingInfo.Name(building))
                },
            }

            defer func(){
                if city.ProducingBuilding == building {
                    selectedElement = element
                    updateMainElementBuilding(building)
                }
            }()

            elements = append(elements, element)
        }
    }

    unitInfo, err := imageCache.GetImage("unitview.lbx", 32, 0)
    if err == nil {
        possibleUnits := getPossibleUnits(city)
        for i, unit := range possibleUnits {

            x1 := 240
            y1 := 4 + i * (buildingInfo.Bounds().Dy() + 1)
            x2 := x1 + unitInfo.Bounds().Dx()
            y2 := y1 + unitInfo.Bounds().Dy()

            element := &uilib.UIElement{
                Rect: image.Rect(x1, y1, x2, y2),
                DoubleLeftClick: func(this *uilib.UIElement) {
                    doOk()
                },
                LeftClick: func(this *uilib.UIElement) {
                    selectedElement = this
                    buildScreen.ProducingBuilding = buildinglib.BuildingNone
                    buildScreen.ProducingUnit = unit
                    updateMainElementUnit(unit)
                },
                RightClick: func(this *uilib.UIElement) {
                    helpEntries := help.GetEntriesByName("Unit Options")
                    if helpEntries != nil {
                        ui.AddElement(uilib.MakeHelpElement(ui, cache, imageCache, helpEntries[0]))
                    }
                },
                Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(x1), float64(y1))
                    screen.DrawImage(unitInfo, &options)

                    use := titleFont

                    // show highlight if element was clicked on
                    if selectedElement == this {
                        use = titleFontWhite
                    }

                    use.Print(screen, float64(x1 + 2), float64(y1 + 1), 1, ebiten.ColorScale{}, unit.String())
                },
            }

            defer func(){
                if city.ProducingUnit.Equals(unit) {
                    selectedElement = element
                    updateMainElementUnit(unit)
                }
            }()

            elements = append(elements, element)
        }
    }

    // FIXME: get both background images, index 1 is when the button is clicked
    buttonBackground, err := imageCache.GetImage("backgrnd.lbx", 24, 0)

    if err == nil {
        cancelX := 100
        cancelY := 181

        // cancel button
        elements = append(elements, &uilib.UIElement{
            Rect: image.Rect(cancelX, cancelY, cancelX + buttonBackground.Bounds().Dx(), cancelY + buttonBackground.Bounds().Dy()),
            LeftClick: func(this *uilib.UIElement) {
                doCancel()
            },
            RightClick: func(this *uilib.UIElement) {
                helpEntries := help.GetEntries(376)
                if helpEntries != nil {
                    ui.AddElement(uilib.MakeHelpElement(ui, cache, imageCache, helpEntries[0]))
                }
            },
            Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(cancelX), float64(cancelY))
                screen.DrawImage(buttonBackground, &options)

                okCancelFont.PrintCenter(screen, float64(cancelX + buttonBackground.Bounds().Dx() / 2), float64(cancelY + 1), 1, ebiten.ColorScale{}, "Cancel")
            },
        })

        okX := 173
        okY := 181
        // ok button
        elements = append(elements, &uilib.UIElement{
            Rect: image.Rect(okX, okY, okX + buttonBackground.Bounds().Dx(), okY + buttonBackground.Bounds().Dy()),
            LeftClick: func(this *uilib.UIElement) {
                doOk()
            },
            RightClick: func(this *uilib.UIElement) {
                helpEntries := help.GetEntries(377)
                if helpEntries != nil {
                    ui.AddElement(uilib.MakeHelpElement(ui, cache, imageCache, helpEntries[0]))
                }
            },
            Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(okX), float64(okY))
                screen.DrawImage(buttonBackground, &options)

                okCancelFont.PrintCenter(screen, float64(okX + buttonBackground.Bounds().Dx() / 2), float64(okY + 1), 1, ebiten.ColorScale{}, "Ok")
            },
        })
    }

    ui.SetElementsFromArray(elements)

    return ui
}

func (build *BuildScreen) Update() BuildScreenState {
    build.UI.StandardUpdate()
    return build.State
}

func (build *BuildScreen) Draw(screen *ebiten.Image) {
    build.UI.Draw(build.UI, screen)
}
