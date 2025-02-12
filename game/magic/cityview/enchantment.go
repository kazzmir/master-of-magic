package cityview

import (
    "fmt"
    "context"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"

    "github.com/hajimehoshi/ebiten/v2"
)

func MakeEnchantmentView(cache *lbx.LbxCache, city *citylib.City, player *playerlib.Player, enchantment data.CityEnchantment) (*uilib.UI, context.Context, error) {
    imageCache := util.MakeImageCache(cache)

    background, _ := imageCache.GetImage("spellscr.lbx", 73, 0)
    buildingSlots := makeBuildingSlots(city)
    fonts, err := makeFonts(cache)
    if err != nil {
        return nil, context.Background(), err
    }

    geom := ebiten.GeoM{}
    geom.Translate(float64(30 * data.ScreenScale), float64(30 * data.ScreenScale))

    var getAlpha util.AlphaFadeFunc
    fadeSpeed := uint64(7)

    quit, cancel := context.WithCancel(context.Background())

    var ui *uilib.UI

    enchantmentBuilding, ok := enchantmentBuildings()[enchantment]
    if !ok {
        enchantmentBuilding = buildinglib.BuildingNone
    }

    ui = &uilib.UI{
        LeftClick: func() {
            getAlpha = ui.MakeFadeOut(fadeSpeed)
            ui.AddDelay(fadeSpeed, func(){
                cancel()
            })
        },
        Draw: func(ui *uilib.UI, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions

            options.ColorScale.ScaleAlpha(getAlpha())
            options.GeoM = geom

            screen.DrawImage(background, &options)

            titleX, titleY := options.GeoM.Apply(float64(background.Bounds().Dx()) / 2, float64(7 * data.ScreenScale))

            fonts.BigFont.PrintCenter(screen, titleX, titleY, float64(data.ScreenScale), options.ColorScale, fmt.Sprintf("%v of %s", city.GetSize(), city.Name))

            descriptionX, descriptionY := options.GeoM.Apply(float64(background.Bounds().Dx()) / 2, float64(background.Bounds().Dy() - fonts.CastFont.Height() * data.ScreenScale - 2 * data.ScreenScale))
            fonts.CastFont.PrintCenter(screen, descriptionX, descriptionY, float64(data.ScreenScale), options.ColorScale, fmt.Sprintf("You cast %v", enchantment.Name()))

            geom2 := geom
            geom2.Translate(float64(5 * data.ScreenScale), float64(28 * data.ScreenScale))
            drawCityScape(screen, city, buildingSlots, buildinglib.BuildingNone, 0, enchantmentBuilding, ui.Counter / 8, &imageCache, fonts, player, geom2, getAlpha())
        },
    }

    getAlpha = ui.MakeFadeIn(fadeSpeed)

    ui.SetElementsFromArray(nil)

    // just to silence log warning
    ui.AddElement(&uilib.UIElement{
    })

    return ui, quit, nil
}
