package cityview

import (
    "fmt"
    "context"
    "image"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"

    "github.com/hajimehoshi/ebiten/v2"
)

func MakeNewBuildingView(cache *lbx.LbxCache, city *citylib.City, player *playerlib.Player, newBuilding buildinglib.Building, name string) (*uilib.UI, context.Context, error) {
    imageCache := util.MakeImageCache(cache)

    background, _ := imageCache.GetImage("spellscr.lbx", 73, 0)
    buildingSlots := makeBuildingSlots(city)
    fonts, err := fontslib.MakeCityViewFonts(cache)
    if err != nil {
        return nil, context.Background(), err
    }

    geom := ebiten.GeoM{}
    geom.Translate(float64(30), float64(30))

    var getAlpha util.AlphaFadeFunc
    fadeSpeed := uint64(7)

    quit, cancel := context.WithCancel(context.Background())

    var ui *uilib.UI

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

            scale.DrawScaled(screen, background, &options)

            titleX, titleY := options.GeoM.Apply(float64(background.Bounds().Dx()) / 2, float64(7))

            fonts.BigFont.PrintOptions2(screen, titleX, titleY, font.FontOptions{Justify: font.FontJustifyCenter, Options: &options, Scale: scale.ScaleAmount}, fmt.Sprintf("%v of %s", city.GetSize(), city.Name))

            descriptionX, descriptionY := options.GeoM.Apply(float64(background.Bounds().Dx()) / 2, float64(background.Bounds().Dy() - fonts.CastFont.Height() - 2))
            fonts.CastFont.PrintOptions2(screen, descriptionX, descriptionY, font.FontOptions{Justify: font.FontJustifyCenter, Scale: scale.ScaleAmount, Options: &options}, fmt.Sprintf("You cast %v", name))

            geom2 := geom
            geom2.Translate(5, 27)

            x1, y1 := geom2.Apply(0, 0)
            // FIXME: get this rectangle from city-screen.go
            x2, y2 := geom2.Apply(206, 96)

            cityScapeScreen := screen.SubImage(scale.ScaleRect(image.Rect(int(x1), int(y1), int(x2), int(y2)))).(*ebiten.Image)
            drawCityScape(cityScapeScreen, city, buildingSlots, buildinglib.BuildingNone, 0, newBuilding, ui.Counter / 8, &imageCache, fonts, player, geom2, getAlpha())
        },
    }

    getAlpha = ui.MakeFadeIn(fadeSpeed)

    ui.SetElementsFromArray(nil)

    // just to silence log warning
    ui.AddElement(&uilib.UIElement{
    })

    return ui, quit, nil
}

func MakeEnchantmentView(cache *lbx.LbxCache, city *citylib.City, player *playerlib.Player, enchantment data.CityEnchantment) (*uilib.UI, context.Context, error) {
    enchantmentBuilding, ok := buildinglib.EnchantmentBuildings()[enchantment]
    if !ok {
        enchantmentBuilding = buildinglib.BuildingNone
    }

    return MakeNewBuildingView(cache, city, player, enchantmentBuilding, enchantment.Name())
}
