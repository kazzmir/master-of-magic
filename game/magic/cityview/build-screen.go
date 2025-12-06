package cityview

import (
    "fmt"
    "image"
    "slices"
    "cmp"
    "log"
    // "strings"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
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

type BuildScreenFonts struct {
    TitleFont *font.Font
    TitleFontWhite *font.Font
    DescriptionFont *font.Font
    OkCancelFont *font.Font
    SmallFont *font.Font
    MediumFont *font.Font
}

func MakeBuildScreenFonts(cache *lbx.LbxCache) *BuildScreenFonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Unable to load fonts: %v", err)
        return nil
    }

    return &BuildScreenFonts{
        TitleFont: loader(fontslib.TitleFont),
        TitleFontWhite: loader(fontslib.TitleFontWhite),
        DescriptionFont: loader(fontslib.WhiteBig),
        OkCancelFont: loader(fontslib.YellowBig),
        SmallFont: loader(fontslib.SmallWhite),
        MediumFont: loader(fontslib.MediumWhite2),
    }
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

    fonts := MakeBuildScreenFonts(cache)

    // var elements []*uilib.UIElement

    ui := &uilib.UI{
        Cache: cache,
        Draw: func(ui *uilib.UI, screen *ebiten.Image) {
            mainInfo, err := imageCache.GetImage("unitview.lbx", 0, 0)
            if err == nil {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(75), 0)
                scale.DrawScaled(screen, mainInfo, &options)
            }

            ui.StandardDraw(screen)
        },
    }

    group := uilib.MakeGroup()
    ui.AddGroup(group)

    mainGroup := uilib.MakeGroup()

    buildingInfo, err := imageCache.GetImage("unitview.lbx", 31, 0)

    var selectedElement *uilib.UIElement

    updateMainElementBuilding := func(building buildinglib.Building){
        descriptionWrapped := fonts.DescriptionFont.CreateWrappedText(float64(155), 1, buildDescriptions.Get(building))

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

        allowsWrapped := fonts.MediumFont.CreateWrappedText(float64(100), 1, allows)

        ui.RemoveGroup(mainGroup)
        mainGroup = uilib.MakeGroup()
        ui.AddGroup(mainGroup)
        mainGroup.AddElement(&uilib.UIElement{
            Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                images, err := imageCache.GetImages("cityscap.lbx", GetProducingBuildingIndex(building))
                if err == nil {
                    middleX := float64(103)
                    middleY := float64(22)
                    index := (ui.Counter / 7) % uint64(len(images))

                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(middleX, middleY)
                    options.GeoM.Translate(-float64(images[index].Bounds().Dx() / 2), -float64(images[index].Bounds().Dy() / 2))

                    width := 44
                    height := 36
                    clipRect := image.Rect(int(middleX) - width / 2, int(middleY) - height / 2, int(middleX) + width / 2, int(middleY) + height / 2)
                    clip := screen.SubImage(scale.ScaleRect(clipRect)).(*ebiten.Image)

                    // vector.DrawFilledRect(clip, float32(clipRect.Min.X), float32(clipRect.Min.Y), float32(clipRect.Bounds().Dx()), float32(clipRect.Bounds().Dy()), color.RGBA{R: 255, G: 255, B: 255, A: 255}, true)

                    scale.DrawScaled(clip, images[index], &options)

                    // vector.DrawFilledCircle(screen, float32(middleX), float32(middleY), 1, color.RGBA{255, 255, 255, 255}, true)
                }

                fonts.DescriptionFont.PrintOptions(screen, float64(130), float64(12), font.FontOptions{Scale: scale.ScaleAmount}, city.BuildingInfo.Name(building))
                fonts.SmallFont.PrintOptions(screen, float64(130), float64(33), font.FontOptions{Scale: scale.ScaleAmount}, fmt.Sprintf("Cost %v", city.BuildingInfo.ProductionCost(building)))

                fonts.DescriptionFont.PrintOptions(screen, float64(85), float64(48), font.FontOptions{Scale: scale.ScaleAmount}, "Maintenance")

                buildingMaintenance := city.BuildingInfo.UpkeepCost(building)

                if buildingMaintenance == 0 {
                    fonts.MediumFont.PrintOptions(screen, float64(85) + fonts.DescriptionFont.MeasureTextWidth("Maintenance", 1) + float64(4), float64(49), font.FontOptions{Scale: scale.ScaleAmount}, "0")
                } else {
                    smallCoin, err := imageCache.GetImage("backgrnd.lbx", 42, 0)
                    if err == nil {
                        var options ebiten.DrawImageOptions
                        options.GeoM.Translate(float64(85) + fonts.DescriptionFont.MeasureTextWidth("Maintenance", 1) + float64(3), float64(50))
                        for i := 0; i < buildingMaintenance; i++ {
                            scale.DrawScaled(screen, smallCoin, &options)
                            options.GeoM.Translate(float64(smallCoin.Bounds().Dx() + 2), 0)
                        }
                    }
                }

                fonts.DescriptionFont.PrintOptions(screen, 85, 58, font.FontOptions{Scale: scale.ScaleAmount}, "Allows")

                fonts.MediumFont.RenderWrapped(screen, float64(85) + fonts.DescriptionFont.MeasureTextWidth("Allows", 1) + float64(10), float64(59), allowsWrapped, font.FontOptions{Scale: scale.ScaleAmount})

                fonts.DescriptionFont.RenderWrapped(screen, 85, 108, descriptionWrapped, font.FontOptions{Scale: scale.ScaleAmount})

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
        bannerUnit := units.MakeOverworldUnitFromUnit(unit, 0, 0, city.Plane, city.GetBanner(), nil, &units.NoEnchantments{})
        productionCost := city.UnitProductionCost(&unit)
        mainGroup.AddElement(&uilib.UIElement{
            Order: -1,
            Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(104), float64(28))
                unitview.RenderUnitViewImage(screen, imageCache, bannerUnit, options, false, 0)

                options.GeoM.Reset()
                options.GeoM.Translate(float64(130), float64(7))
                unitview.RenderUnitInfoBuild(screen, imageCache, bannerUnit, fonts.DescriptionFont, fonts.SmallFont, options, productionCost)

                /*
                options.GeoM.Reset()
                options.GeoM.Translate(float64(85), float64(48))
                unitview.RenderUnitInfoStats(screen, imageCache, bannerUnit, 10, fonts.DescriptionFont, fonts.SmallFont, options)
                */

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

        var defaultOptions ebiten.DrawImageOptions
        defaultOptions.GeoM.Translate(float64(85), float64(48))

        mainGroup.AddElements(unitview.CreateUnitInfoStatsElements(imageCache, bannerUnit, 10, fonts.DescriptionFont, fonts.SmallFont, defaultOptions, &getAlpha, 0))

        mainGroup.AddElements(unitview.MakeUnitAbilitiesElements(mainGroup, cache, imageCache, bannerUnit, fonts.MediumFont, 85, 108, &ui.Counter, 0, &getAlpha, true, 0, false))
        // ui.AddElements(mainElements)
    }

    if err == nil {
        possibleBuildings := city.ComputePossibleBuildings(false)
        for i, building := range slices.SortedFunc(slices.Values(possibleBuildings.Values()), func (a, b buildinglib.Building) int {
            return cmp.Compare(city.BuildingInfo.GetBuildingIndex(a), city.BuildingInfo.GetBuildingIndex(b))
        }) {
            x1 := 0
            y1 := 4 + i * (buildingInfo.Bounds().Dy() + 1)
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
                    scale.DrawScaled(screen, buildingInfo, &options)

                    use := fonts.TitleFont

                    // show highlight if element was clicked on
                    if selectedElement == this {
                        use = fonts.TitleFontWhite
                    }

                    use.PrintOptions(screen, float64(x1 + 2), float64(y1 + 1), font.FontOptions{Scale: scale.ScaleAmount}, city.BuildingInfo.Name(building))
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

            x1 := 240
            y1 := 4 + i * (buildingInfo.Bounds().Dy() + 1)
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
                    scale.DrawScaled(screen, unitInfo, &options)

                    use := fonts.TitleFont

                    // show highlight if element was clicked on
                    if selectedElement == this {
                        use = fonts.TitleFontWhite
                    }

                    use.PrintOptions(screen, float64(x1 + 2), float64(y1 + 1), font.FontOptions{Scale: scale.ScaleAmount}, unit.String())
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
        cancelX := 100
        cancelY := 181

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
                scale.DrawScaled(screen, buttonBackground, &options)

                fonts.OkCancelFont.PrintOptions(screen, float64(cancelX + buttonBackground.Bounds().Dx() / 2), float64(cancelY + 1), font.FontOptions{Scale: scale.ScaleAmount, Justify: font.FontJustifyCenter}, "Cancel")
            },
        })

        okX := 173
        okY := 181
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
                scale.DrawScaled(screen, buttonBackground, &options)

                fonts.OkCancelFont.PrintOptions(screen, float64(okX + buttonBackground.Bounds().Dx() / 2), float64(okY + 1), font.FontOptions{Scale: scale.ScaleAmount, Justify: font.FontJustifyCenter}, "Ok")
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
