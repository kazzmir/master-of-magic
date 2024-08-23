package city

import (
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/hajimehoshi/ebiten/v2"
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
    
    titleFont := font.MakeOptimizedFont(fonts[2])

    var elements []*uilib.UIElement

    buildingInfo, err := imageCache.GetImage("unitview.lbx", 31, 0)

    if err == nil {
        possibleBuildings := []Building{BuildingTradeGoods, BuildingHousing, BuildingBarracks, BuildingStables}
        for i, building := range possibleBuildings {

            element := &uilib.UIElement{
                Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(0, 4 + float64(i * (buildingInfo.Bounds().Dy() + 1)))
                    screen.DrawImage(buildingInfo, &options)
                    x, y := options.GeoM.Apply(0, 0)
                    titleFont.Print(screen, x + 2, y + 1, 1, building.String())
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

    ui.SetElementsFromArray(elements)

    return ui
}

func (build *BuildScreen) Update() {
    build.UI.StandardUpdate()
}

func (build *BuildScreen) Draw(screen *ebiten.Image) {
    build.UI.Draw(build.UI, screen)
}
