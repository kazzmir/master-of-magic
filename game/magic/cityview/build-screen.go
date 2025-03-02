package cityview

import (
    "fmt"
    "image"
    "slices"
    "cmp"
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
    helplib "github.com/kazzmir/master-of-magic/game/magic/help"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
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

func MakeBuildScreen(cache *lbx.LbxCache, city *citylib.City) *BuildScreen {
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
        ProducingBuilding: city.ProducingBuilding,
        ProducingUnit: city.ProducingUnit,
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

func makeBuildUI(cache *lbx.LbxCache, imageCache *util.ImageCache, city *citylib.City, buildScreen *BuildScreen, doCancel func(), doOk func()) *uilib.UI {
    helpLbx, err := cache.GetLbxFile("HELP.LBX")
    if err != nil {
        return nil
    }

    help, err := helplib.ReadHelp(helpLbx, 2)
    if err != nil {
        return nil
    }

    buildDescriptions := buildinglib.MakeBuildDescriptions(cache)

    fonts := fontslib.MakeBuildScreenFonts(cache)

    // var elements []*uilib.UIElement

    ui := &uilib.UI{
        Cache: cache,
        Draw: func(ui *uilib.UI, screen *ebiten.Image) {
            mainInfo, err := imageCache.GetImage("unitview.lbx", 0, 0)
            if err == nil {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(75 * data.ScreenScale), 0)
                screen.DrawImage(mainInfo, &options)
            }

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })

        },
    }

    group := uilib.MakeGroup()
    ui.AddGroup(group)

    mainGroup := uilib.MakeGroup()

    buildingInfo, err := imageCache.GetImage("unitview.lbx", 31, 0)

    var selectedElement *uilib.UIElement

    updateMainElementBuilding := func(building buildinglib.Building){
        descriptionWrapped := fonts.DescriptionFont.CreateWrappedText(float64(155 * data.ScreenScale), float64(data.ScreenScale), buildDescriptions.Get(building))

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

        allowsWrapped := fonts.MediumFont.CreateWrappedText(float64(100 * data.ScreenScale), float64(data.ScreenScale), allows)

        ui.RemoveGroup(mainGroup)
        mainGroup = uilib.MakeGroup()
        ui.AddGroup(mainGroup)
        mainGroup.AddElement(&uilib.UIElement{
            Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                images, err := imageCache.GetImages("cityscap.lbx", GetProducingBuildingIndex(building))
                if err == nil {
                    middleX := float64(103 * data.ScreenScale)
                    middleY := float64(22 * data.ScreenScale)
                    index := (ui.Counter / 7) % uint64(len(images))

                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(middleX, middleY)
                    options.GeoM.Translate(-float64(images[index].Bounds().Dx() / 2), -float64(images[index].Bounds().Dy() / 2))

                    width := 44 * data.ScreenScale
                    height := 36 * data.ScreenScale
                    clipRect := image.Rect(int(middleX) - width / 2, int(middleY) - height / 2, int(middleX) + width / 2, int(middleY) + height / 2)
                    clip := screen.SubImage(clipRect).(*ebiten.Image)

                    // vector.DrawFilledRect(clip, float32(clipRect.Min.X), float32(clipRect.Min.Y), float32(clipRect.Bounds().Dx()), float32(clipRect.Bounds().Dy()), color.RGBA{R: 255, G: 255, B: 255, A: 255}, true)

                    clip.DrawImage(images[index], &options)

                    // vector.DrawFilledCircle(screen, float32(middleX), float32(middleY), 1, color.RGBA{255, 255, 255, 255}, true)
                }

                fonts.DescriptionFont.Print(screen, float64(130 * data.ScreenScale), float64(12 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, city.BuildingInfo.Name(building))
                fonts.SmallFont.Print(screen, float64(130 * data.ScreenScale), float64(33 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("Cost %v", city.BuildingInfo.ProductionCost(building)))

                fonts.DescriptionFont.Print(screen, float64(85 * data.ScreenScale), float64(48 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, "Maintenance")

                buildingMaintenance := city.BuildingInfo.UpkeepCost(building)

                if buildingMaintenance == 0 {
                    fonts.MediumFont.Print(screen, float64(85 * data.ScreenScale) + fonts.DescriptionFont.MeasureTextWidth("Maintenance", float64(data.ScreenScale)) + float64(4 * data.ScreenScale), float64(49 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, "0")
                } else {
                    smallCoin, err := imageCache.GetImage("backgrnd.lbx", 42, 0)
                    if err == nil {
                        var options ebiten.DrawImageOptions
                        options.GeoM.Translate(float64(85 * data.ScreenScale) + fonts.DescriptionFont.MeasureTextWidth("Maintenance", float64(data.ScreenScale)) + float64(3 * data.ScreenScale), float64(50 * data.ScreenScale))
                        for i := 0; i < buildingMaintenance; i++ {
                            screen.DrawImage(smallCoin, &options)
                            options.GeoM.Translate(float64(smallCoin.Bounds().Dx() + 2 * data.ScreenScale), 0)
                        }
                    }
                }

                fonts.DescriptionFont.Print(screen, float64(85 * data.ScreenScale), float64(58 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, "Allows")

                fonts.MediumFont.RenderWrapped(screen, float64(85 * data.ScreenScale) + fonts.DescriptionFont.MeasureTextWidth("Allows", float64(data.ScreenScale)) + float64(10 * data.ScreenScale), float64(59 * data.ScreenScale), allowsWrapped, ebiten.ColorScale{}, font.FontOptions{})

                fonts.DescriptionFont.RenderWrapped(screen, float64(85 * data.ScreenScale), float64(108 * data.ScreenScale), descriptionWrapped, ebiten.ColorScale{}, font.FontOptions{})

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
        })
    }

    updateMainElementUnit := func(unit units.Unit){
        ui.RemoveGroup(mainGroup)
        mainGroup = uilib.MakeGroup()
        ui.AddGroup(mainGroup)
        bannerUnit := units.MakeOverworldUnitFromUnit(unit, 0, 0, city.Plane, city.GetBanner(), nil)
        productionCost := city.UnitProductionCost(&unit)
        mainGroup.AddElement(&uilib.UIElement{
            Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(104 * data.ScreenScale), float64(28 * data.ScreenScale))
                unitview.RenderUnitViewImage(screen, imageCache, bannerUnit, options, false, 0)

                options.GeoM.Reset()
                options.GeoM.Translate(float64(130 * data.ScreenScale), float64(7 * data.ScreenScale))
                unitview.RenderUnitInfoBuild(screen, imageCache, bannerUnit, fonts.DescriptionFont, fonts.SmallFont, options, productionCost)

                options.GeoM.Reset()
                options.GeoM.Translate(float64(85 * data.ScreenScale), float64(48 * data.ScreenScale))
                unitview.RenderUnitInfoStats(screen, imageCache, bannerUnit, 10, fonts.DescriptionFont, fonts.SmallFont, options)

                /*
                options.GeoM.Reset()
                options.GeoM.Translate(85, 108)
                unitview.RenderUnitAbilities(screen, imageCache, bannerUnit, mediumFont, options, true)
                */
            },
        })
        var getAlpha util.AlphaFadeFunc = func () float32 {
            return 1
        }
        mainGroup.AddElements(unitview.MakeUnitAbilitiesElements(mainGroup, cache, imageCache, bannerUnit, fonts.MediumFont, 85 * data.ScreenScale, 108 * data.ScreenScale, &ui.Counter, 0, &getAlpha, true, 0, false))
        // ui.AddElements(mainElements)
    }

    if err == nil {
        possibleBuildings := city.ComputePossibleBuildings()
        for i, building := range slices.SortedFunc(slices.Values(possibleBuildings.Values()), func (a, b buildinglib.Building) int {
            return cmp.Compare(city.BuildingInfo.GetBuildingIndex(a), city.BuildingInfo.GetBuildingIndex(b))
        }) {
            x1 := 0
            y1 := 4 * data.ScreenScale + i * (buildingInfo.Bounds().Dy() + 1)
            x2 := x1 + buildingInfo.Bounds().Dx()
            y2 := y1 + buildingInfo.Bounds().Dy()

            element := &uilib.UIElement{
                Rect: image.Rect(x1, y1, x2, y2),
                DoubleLeftClick: func(this *uilib.UIElement) {
                    doOk()
                },
                PlaySoundLeftClick: true,
                LeftClick: func(this *uilib.UIElement) {
                    selectedElement = this
                    buildScreen.ProducingBuilding = building
                    buildScreen.ProducingUnit = units.UnitNone
                    updateMainElementBuilding(building)
                },
                RightClick: func(this *uilib.UIElement) {
                    helpEntries := help.GetEntriesByName("Building Options")
                    if helpEntries != nil {
                        group.AddElement(uilib.MakeHelpElement(group, cache, imageCache, helpEntries[0]))
                    }
                },
                Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(x1), float64(y1))
                    screen.DrawImage(buildingInfo, &options)

                    use := fonts.TitleFont

                    // show highlight if element was clicked on
                    if selectedElement == this {
                        use = fonts.TitleFontWhite
                    }

                    use.Print(screen, float64(x1 + 2 * data.ScreenScale), float64(y1 + 1 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, city.BuildingInfo.Name(building))
                },
            }

            defer func(){
                if city.ProducingBuilding == building {
                    selectedElement = element
                    updateMainElementBuilding(building)
                }
            }()

            group.AddElement(element)
        }
    }

    unitInfo, err := imageCache.GetImage("unitview.lbx", 32, 0)
    if err == nil {
        possibleUnits := city.ComputePossibleUnits()
        for i, unit := range slices.SortedFunc(slices.Values(possibleUnits), func (a, b units.Unit) int {
            return cmp.Compare(a.Name, b.Name)
        }) {

            x1 := 240 * data.ScreenScale
            y1 := 4 * data.ScreenScale + i * (buildingInfo.Bounds().Dy() + 1)
            x2 := x1 + unitInfo.Bounds().Dx()
            y2 := y1 + unitInfo.Bounds().Dy()

            element := &uilib.UIElement{
                Rect: image.Rect(x1, y1, x2, y2),
                DoubleLeftClick: func(this *uilib.UIElement) {
                    doOk()
                },
                PlaySoundLeftClick: true,
                LeftClick: func(this *uilib.UIElement) {
                    selectedElement = this
                    buildScreen.ProducingBuilding = buildinglib.BuildingNone
                    buildScreen.ProducingUnit = unit
                    updateMainElementUnit(unit)
                },
                RightClick: func(this *uilib.UIElement) {
                    helpEntries := help.GetEntriesByName("Unit Options")
                    if helpEntries != nil {
                        group.AddElement(uilib.MakeHelpElement(group, cache, imageCache, helpEntries[0]))
                    }
                },
                Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(x1), float64(y1))
                    screen.DrawImage(unitInfo, &options)

                    use := fonts.TitleFont

                    // show highlight if element was clicked on
                    if selectedElement == this {
                        use = fonts.TitleFontWhite
                    }

                    use.Print(screen, float64(x1 + 2 * data.ScreenScale), float64(y1 + 1 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, unit.String())
                },
            }

            defer func(){
                if city.ProducingUnit.Equals(unit) {
                    selectedElement = element
                    updateMainElementUnit(unit)
                }
            }()

            group.AddElement(element)
        }
    }

    // FIXME: get both background images, index 1 is when the button is clicked
    buttonBackground, err := imageCache.GetImage("backgrnd.lbx", 24, 0)

    if err == nil {
        cancelX := 100 * data.ScreenScale
        cancelY := 181 * data.ScreenScale

        // cancel button
        group.AddElement(&uilib.UIElement{
            Rect: util.ImageRect(cancelX, cancelY, buttonBackground),
            LeftClick: func(this *uilib.UIElement) {
                doCancel()
            },
            RightClick: func(this *uilib.UIElement) {
                helpEntries := help.GetEntries(376)
                if helpEntries != nil {
                    group.AddElement(uilib.MakeHelpElement(group, cache, imageCache, helpEntries[0]))
                }
            },
            Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(cancelX), float64(cancelY))
                screen.DrawImage(buttonBackground, &options)

                fonts.OkCancelFont.PrintCenter(screen, float64(cancelX + buttonBackground.Bounds().Dx() / 2), float64(cancelY + 1 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, "Cancel")
            },
        })

        okX := 173 * data.ScreenScale
        okY := 181 * data.ScreenScale
        // ok button
        group.AddElement(&uilib.UIElement{
            Rect: util.ImageRect(okX, okY, buttonBackground),
            LeftClick: func(this *uilib.UIElement) {
                doOk()
            },
            RightClick: func(this *uilib.UIElement) {
                helpEntries := help.GetEntries(377)
                if helpEntries != nil {
                    group.AddElement(uilib.MakeHelpElement(group, cache, imageCache, helpEntries[0]))
                }
            },
            Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(okX), float64(okY))
                screen.DrawImage(buttonBackground, &options)

                fonts.OkCancelFont.PrintCenter(screen, float64(okX + buttonBackground.Bounds().Dx() / 2), float64(okY + 1 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, "Ok")
            },
        })
    }

    ui.SetElementsFromArray(nil)

    return ui
}

func (build *BuildScreen) Update() BuildScreenState {
    build.UI.StandardUpdate()
    return build.State
}

func (build *BuildScreen) Draw(screen *ebiten.Image) {
    build.UI.Draw(build.UI, screen)
}
