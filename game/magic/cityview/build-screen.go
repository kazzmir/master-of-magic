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

func allowedByRace(building buildinglib.Building, race data.Race) bool {
    switch building {
        case buildinglib.BuildingBarracks,
             buildinglib.BuildingArmory,
             buildinglib.BuildingSmithy,
             buildinglib.BuildingLibrary,
             buildinglib.BuildingShrine,
             buildinglib.BuildingMarketplace,
             buildinglib.BuildingGranary,
             buildinglib.BuildingFarmersMarket,
             buildinglib.BuildingBuildersHall,
             buildinglib.BuildingCityWalls,
             buildinglib.BuildingTradeGoods,
             buildinglib.BuildingHousing,
             buildinglib.BuildingFightersGuild:
                return true
        case buildinglib.BuildingArmorersGuild:
            if race == data.RaceGnoll || race == data.RaceHalfling {
                return false
            }
            return true
        case buildinglib.BuildingStables:
            // not halfling or dwarf
            if race == data.RaceHalfling || race == data.RaceDwarf {
                return false
            }

            return true
        case buildinglib.BuildingWarCollege:
            return race == data.RaceHighElf ||
                   race == data.RaceHighMen ||
                   race == data.RaceNomad ||
                   race == data.RaceOrc ||
                   race == data.RaceBeastmen ||
                   race == data.RaceDarkElf ||
                   race == data.RaceDraconian
        case buildinglib.BuildingAnimistsGuild:
            return race == data.RaceHighElf ||
                   race == data.RaceHighMen ||
                   race == data.RaceNomad ||
                   race == data.RaceOrc ||
                   race == data.RaceBeastmen ||
                   race == data.RaceDarkElf ||
                   race == data.RaceDraconian ||
                   race == data.RaceTroll
        case buildinglib.BuildingFantasticStable:
            return race == data.RaceHighElf ||
                   race == data.RaceNomad ||
                   race == data.RaceOrc ||
                   race == data.RaceDarkElf
        case buildinglib.BuildingShipYard:
            return race == data.RaceBarbarian ||
                   race == data.RaceGnoll ||
                   race == data.RaceHighElf ||
                   race == data.RaceHighMen ||
                   race == data.RaceNomad ||
                   race == data.RaceOrc ||
                   race == data.RaceDarkElf ||
                   race == data.RaceDraconian
        case buildinglib.BuildingMaritimeGuild:
            return race == data.RaceBarbarian ||
                   race == data.RaceHighMen ||
                   race == data.RaceOrc
        case buildinglib.BuildingSawmill,
             buildinglib.BuildingShipwrightsGuild,
             buildinglib.BuildingForestersGuild:
            return race != data.RaceLizard
        case buildinglib.BuildingSagesGuild,
             buildinglib.BuildingAlchemistsGuild:
            if race == data.RaceGnoll ||
               race == data.RaceKlackon ||
               race == data.RaceLizard ||
               race == data.RaceTroll {
                   return false
            }

            return true
        case buildinglib.BuildingOracle:
            return race == data.RaceHighMen ||
                   race == data.RaceNomad ||
                   race == data.RaceOrc ||
                   race == data.RaceBeastmen ||
                   race == data.RaceDarkElf
        case buildinglib.BuildingUniversity:
            return race == data.RaceHighElf ||
                   race == data.RaceHighMen ||
                   race == data.RaceNomad ||
                   race == data.RaceOrc ||
                   race == data.RaceBeastmen ||
                   race == data.RaceDarkElf ||
                   race == data.RaceDraconian
        case buildinglib.BuildingWizardsGuild:
            return race == data.RaceHighElf ||
                   race == data.RaceHighMen ||
                   race == data.RaceOrc ||
                   race == data.RaceBeastmen ||
                   race == data.RaceDarkElf ||
                   race == data.RaceDraconian
        case buildinglib.BuildingTemple:
            return race != data.RaceKlackon
        case buildinglib.BuildingParthenon:
            return race == data.RaceBarbarian ||
                   race == data.RaceHalfling ||
                   race == data.RaceHighMen ||
                   race == data.RaceNomad ||
                   race == data.RaceOrc ||
                   race == data.RaceBeastmen ||
                   race == data.RaceDarkElf ||
                   race == data.RaceDraconian ||
                   race == data.RaceTroll
        case buildinglib.BuildingCathedral:
            return race == data.RaceHalfling ||
                   race == data.RaceHighMen ||
                   race == data.RaceNomad ||
                   race == data.RaceOrc ||
                   race == data.RaceBeastmen ||
                   race == data.RaceDraconian ||
                   race == data.RaceTroll
        case buildinglib.BuildingBank:
            return race == data.RaceHighElf ||
                   race == data.RaceHighMen ||
                   race == data.RaceNomad ||
                   race == data.RaceOrc ||
                   race == data.RaceBeastmen ||
                   race == data.RaceDarkElf ||
                   race == data.RaceDraconian
        case buildinglib.BuildingMerchantsGuild:
            return race == data.RaceHighElf ||
                   race == data.RaceHighMen ||
                   race == data.RaceNomad ||
                   race == data.RaceOrc ||
                   race == data.RaceDarkElf ||
                   race == data.RaceDraconian
        case buildinglib.BuildingMechaniciansGuild:
            return race == data.RaceHighElf ||
                   race == data.RaceHighMen ||
                   race == data.RaceNomad ||
                   race == data.RaceOrc ||
                   race == data.RaceBeastmen ||
                   race == data.RaceDarkElf
        case buildinglib.BuildingMinersGuild:
            if race == data.RaceLizard || race == data.RaceTroll {
                return false
            }

            return true
    }

    return false
}

/* return the buildings that can be built, based on what the city already has
 */
func computePossibleBuildings(city *citylib.City) []buildinglib.Building {
    var possibleBuildings []buildinglib.Building

    for _, building := range buildinglib.Buildings() {
        if city.Buildings.Contains(building) {
            continue
        }

        if !allowedByRace(building, city.Race) {
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

    var mainElements []*uilib.UIElement

    // elements = append(elements, mainElement)

    buildingInfo, err := imageCache.GetImage("unitview.lbx", 31, 0)

    var selectedElement *uilib.UIElement

    updateMainElementBuilding := func(building buildinglib.Building){
        descriptionWrapped := descriptionFont.CreateWrappedText(float64(155 * data.ScreenScale), float64(data.ScreenScale), buildDescriptions.Get(building))

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

        allowsWrapped := mediumFont.CreateWrappedText(float64(100 * data.ScreenScale), float64(data.ScreenScale), allows)

        ui.RemoveElements(mainElements)
        mainElements = nil
        mainElements = append(mainElements, &uilib.UIElement{
            Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                images, err := imageCache.GetImages("cityscap.lbx", GetBuildingIndex(building))
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

                descriptionFont.Print(screen, float64(130 * data.ScreenScale), float64(12 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, city.BuildingInfo.Name(building))
                smallFont.Print(screen, float64(130 * data.ScreenScale), float64(33 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("Cost %v", city.BuildingInfo.ProductionCost(building)))

                descriptionFont.Print(screen, float64(85 * data.ScreenScale), float64(48 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, "Maintenance")

                buildingMaintenance := city.BuildingInfo.UpkeepCost(building)

                if buildingMaintenance == 0 {
                    mediumFont.Print(screen, float64(85 * data.ScreenScale) + descriptionFont.MeasureTextWidth("Maintenance", float64(data.ScreenScale)) + float64(4 * data.ScreenScale), float64(49 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, "0")
                } else {
                    smallCoin, err := imageCache.GetImage("backgrnd.lbx", 42, 0)
                    if err == nil {
                        var options ebiten.DrawImageOptions
                        options.GeoM.Translate(float64(85 * data.ScreenScale) + descriptionFont.MeasureTextWidth("Maintenance", float64(data.ScreenScale)) + float64(3 * data.ScreenScale), float64(50 * data.ScreenScale))
                        for i := 0; i < buildingMaintenance; i++ {
                            screen.DrawImage(smallCoin, &options)
                            options.GeoM.Translate(float64(smallCoin.Bounds().Dx() + 2 * data.ScreenScale), 0)
                        }
                    }
                }

                descriptionFont.Print(screen, float64(85 * data.ScreenScale), float64(58 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, "Allows")

                mediumFont.RenderWrapped(screen, float64(85 * data.ScreenScale) + descriptionFont.MeasureTextWidth("Allows", float64(data.ScreenScale)) + float64(10 * data.ScreenScale), float64(59 * data.ScreenScale), allowsWrapped, ebiten.ColorScale{}, false)

                descriptionFont.RenderWrapped(screen, float64(85 * data.ScreenScale), float64(108 * data.ScreenScale), descriptionWrapped, ebiten.ColorScale{}, false)

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
        ui.AddElements(mainElements)
    }

    updateMainElementUnit := func(unit units.Unit){
        ui.RemoveElements(mainElements)
        mainElements = nil
        bannerUnit := units.MakeOverworldUnitFromUnit(unit, 0, 0, city.Plane, city.Banner, nil)
        mainElements = append(mainElements, &uilib.UIElement{
            Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(104, 28)
                unitview.RenderCombatImage(screen, imageCache, bannerUnit, options, 0)

                options.GeoM.Reset()
                options.GeoM.Translate(130, 7)
                unitview.RenderUnitInfoBuild(screen, imageCache, bannerUnit, descriptionFont, smallFont, options)

                options.GeoM.Reset()
                options.GeoM.Translate(85, 48)
                unitview.RenderUnitInfoStats(screen, imageCache, bannerUnit, 10, descriptionFont, smallFont, options)

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
        mainElements = append(mainElements, unitview.MakeUnitAbilitiesElements(imageCache, bannerUnit, mediumFont, 85, 108, &ui.Counter, 0, &getAlpha, true)...)
        ui.AddElements(mainElements)
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
        cancelX := 100 * data.ScreenScale
        cancelY := 181 * data.ScreenScale

        // cancel button
        elements = append(elements, &uilib.UIElement{
            Rect: util.ImageRect(cancelX, cancelY, buttonBackground),
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

                okCancelFont.PrintCenter(screen, float64(cancelX + buttonBackground.Bounds().Dx() / 2), float64(cancelY + 1), float64(data.ScreenScale), ebiten.ColorScale{}, "Cancel")
            },
        })

        okX := 173 * data.ScreenScale
        okY := 181 * data.ScreenScale
        // ok button
        elements = append(elements, &uilib.UIElement{
            Rect: util.ImageRect(okX, okY, buttonBackground),
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

                okCancelFont.PrintCenter(screen, float64(okX + buttonBackground.Bounds().Dx() / 2), float64(okY + 1), float64(data.ScreenScale), ebiten.ColorScale{}, "Ok")
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
