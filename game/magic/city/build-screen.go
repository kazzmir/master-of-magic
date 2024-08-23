package city

import (
    "log"
    "image"
    "image/color"
    "bytes"
    // "strings"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/hajimehoshi/ebiten/v2"
    // "github.com/hajimehoshi/ebiten/v2/vector"
)

type BuildScreen struct {
    LbxCache *lbx.LbxCache
    ImageCache *util.ImageCache
    City *City
    UI *uilib.UI
}

func MakeBuildScreen(cache *lbx.LbxCache, city *City) *BuildScreen {
    imageCache := util.MakeImageCache(cache)

    ui := makeBuildUI(cache, &imageCache, city)

    return &BuildScreen{
        LbxCache: cache,
        ImageCache: &imageCache,
        City: city,
        UI: ui,
    }
}

func premultiplyAlpha(c color.RGBA) color.RGBA {
    a := float64(c.A) / 255.0
    return color.RGBA{
        R: uint8(float64(c.R) * a),
        G: uint8(float64(c.G) * a),
        B: uint8(float64(c.B) * a),
        A: c.A,
    }
}

type BuildingDescriptions struct {
    Descriptions []string
}

func (descriptions *BuildingDescriptions) Get(building Building) string {
    switch building {
        case BuildingTradeGoods: return descriptions.Descriptions[1]
        case BuildingHousing: return descriptions.Descriptions[2]
        case BuildingBarracks: return descriptions.Descriptions[3]
        case BuildingArmory: return descriptions.Descriptions[4]
        case BuildingFightersGuild: return descriptions.Descriptions[5]
        case BuildingArmorersGuild: return descriptions.Descriptions[6]
        case BuildingWarCollege: return descriptions.Descriptions[7]
        case BuildingSmithy: return descriptions.Descriptions[8]
        case BuildingStables: return descriptions.Descriptions[9]
        case BuildingAnimistsGuild: return descriptions.Descriptions[10]
        case BuildingFantasticStable: return descriptions.Descriptions[11]
        case BuildingShipwrightsGuild: return descriptions.Descriptions[12]
        case BuildingShipYard: return descriptions.Descriptions[13]
        case BuildingMaritimeGuild: return descriptions.Descriptions[14]
        case BuildingSawmill: return descriptions.Descriptions[15]
        case BuildingLibrary: return descriptions.Descriptions[16]
        case BuildingSagesGuild: return descriptions.Descriptions[17]
        case BuildingOracle: return descriptions.Descriptions[18]
        case BuildingAlchemistsGuild: return descriptions.Descriptions[19]
        case BuildingUniversity: return descriptions.Descriptions[20]
        case BuildingWizardsGuild: return descriptions.Descriptions[21]
        case BuildingShrine: return descriptions.Descriptions[22]
        case BuildingTemple: return descriptions.Descriptions[23]
        case BuildingParthenon: return descriptions.Descriptions[24]
        case BuildingCathedral: return descriptions.Descriptions[25]
        case BuildingMarketplace: return descriptions.Descriptions[26]
        case BuildingBank: return descriptions.Descriptions[27]
        case BuildingMerchantsGuild: return descriptions.Descriptions[28]
        case BuildingGranary: return descriptions.Descriptions[29]
        case BuildingFarmersMarket: return descriptions.Descriptions[30]
        case BuildingForestersGuild: return descriptions.Descriptions[31]
        case BuildingBuildersHall: return descriptions.Descriptions[32]
        case BuildingMechaniciansGuild: return descriptions.Descriptions[33]
        case BuildingMinersGuild: return descriptions.Descriptions[34]
        case BuildingCityWalls: return descriptions.Descriptions[35]
    }

    return ""
}

func readBuildDescriptions(buildDescriptionLbx *lbx.LbxFile) []string {
    entries, err := buildDescriptionLbx.RawData(0)
    if err != nil {
        return nil
    }

    reader := bytes.NewReader(entries)

    count, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil
    }

    if count > 10000 {
        return nil
    }

    size, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil
    }

    if size > 10000 {
        return nil
    }

    var descriptions []string

    for i := 0; i < int(count); i++ {
        data := make([]byte, size)
        _, err := reader.Read(data)

        if err != nil {
            break
        }

        nullByte := bytes.IndexByte(data, 0)
        if nullByte != -1 {
            descriptions = append(descriptions, string(data[0:nullByte]))
        } else {
            descriptions = append(descriptions, string(data))
        }
    }

    return descriptions
}

func MakeBuildDescriptions(cache *lbx.LbxCache) *BuildingDescriptions {
    buildDescriptionLbx, err := cache.GetLbxFile("buildesc.lbx")
    if err == nil {
    } else {
        log.Printf("Unable to read building descriptions")
    }

    descriptions := readBuildDescriptions(buildDescriptionLbx)

    return &BuildingDescriptions{
        Descriptions: descriptions,
    }
}

func makeBuildUI(cache *lbx.LbxCache, imageCache *util.ImageCache, city *City) *uilib.UI {

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

    helpLbx, err := cache.GetLbxFile("HELP.LBX")
    if err != nil {
        return nil
    }

    help, err := helpLbx.ReadHelp(2)
    if err != nil {
        return nil
    }

    _ = help

    buildDescriptions := MakeBuildDescriptions(cache)
    
    titleFont := font.MakeOptimizedFont(fonts[2])

    alphaWhite := premultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 180})

    whitePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        alphaWhite, alphaWhite, alphaWhite,
    }

    titleFontWhite := font.MakeOptimizedFontWithPalette(fonts[2], whitePalette)

    descriptionPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        premultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 90}),
        premultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}),
        premultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 200}),
        premultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 200}),
        premultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 200}),
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
    }

    descriptionFont := font.MakeOptimizedFontWithPalette(fonts[4], descriptionPalette)

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

    updateMainElementBuilding := func(building Building){
        descriptionWrapped := descriptionFont.CreateWrappedText(155, 1, buildDescriptions.Get(building))
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

                descriptionFont.Print(screen, 130, 14, 1, building.String())

                descriptionFont.Print(screen, 85, 48, 1, "Maintenance")
                descriptionFont.Print(screen, 85, 58, 1, "Allows")

                descriptionFont.RenderWrapped(screen, 85, 108, descriptionWrapped, false)

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

    if err == nil {
        possibleBuildings := []Building{BuildingTradeGoods, BuildingHousing, BuildingBarracks, BuildingStables}
        for i, building := range possibleBuildings {

            x1 := 0
            y1 := 4 + i * (buildingInfo.Bounds().Dy() + 1)
            x2 := x1 + buildingInfo.Bounds().Dx()
            y2 := y1 + buildingInfo.Bounds().Dy()

            element := &uilib.UIElement{
                Rect: image.Rect(x1, y1, x2, y2),
                LeftClick: func(this *uilib.UIElement) {
                    selectedElement = this
                    updateMainElementBuilding(building)
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

                    use.Print(screen, float64(x1 + 2), float64(y1 + 1), 1, building.String())
                },
            }

            elements = append(elements, element)
        }
    }

    unitInfo, err := imageCache.GetImage("unitview.lbx", 32, 0)
    if err == nil {
        possibleUnits := []units.Unit{units.LizardSpearmen, units.LizardSettlers}
        for i, unit := range possibleUnits {
            element := &uilib.UIElement{
                Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(240, 4 + float64(i * (unitInfo.Bounds().Dy() + 1)))
                    screen.DrawImage(unitInfo, &options)
                    x, y := options.GeoM.Apply(0, 0)
                    titleFont.Print(screen, x + 2, y + 1, 1, unit.String())
                },
            }

            elements = append(elements, element)
        }
    }

    ui.SetElementsFromArray(elements)

    return ui
}

func (build *BuildScreen) Update() {
    build.UI.StandardUpdate()
}

func (build *BuildScreen) Draw(screen *ebiten.Image) {
    build.UI.Draw(build.UI, screen)
}
