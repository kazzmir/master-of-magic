package cityview

import (
    "fmt"
    "log"
    "context"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/audio"

    "github.com/hajimehoshi/ebiten/v2"
)

func PlayEnchantmentSound(cache *lbx.LbxCache) {
    sound, err := audio.LoadSound(cache, 30)
    if err != nil {
        log.Printf("Error loading city enchantment sound: %v", err)
    } else {
        sound.Play()
    }
}

func MakeEnchantmentView(cache *lbx.LbxCache, city *citylib.City, player *playerlib.Player, spellName string) (*uilib.UI, context.Context, error) {
    imageCache := util.MakeImageCache(cache)

    background, _ := imageCache.GetImage("spellscr.lbx", 73, 0)
    buildingSlots := makeBuildingSlots(city)
    fonts, err := makeFonts(cache)
    if err != nil {
        return nil, context.Background(), err
    }

    geom := ebiten.GeoM{}
    geom.Translate(30, 30)

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

            screen.DrawImage(background, &options)

            titleX, titleY := options.GeoM.Apply(float64(background.Bounds().Dx()) / 2, 7)

            fonts.BigFont.PrintCenter(screen, titleX, titleY, 1, options.ColorScale, fmt.Sprintf("%v of %s", city.GetSize(), city.Name))

            descriptionX, descriptionY := options.GeoM.Apply(float64(background.Bounds().Dx()) / 2, float64(background.Bounds().Dy() - fonts.CastFont.Height() - 2))
            fonts.CastFont.PrintCenter(screen, descriptionX, descriptionY, 1, options.ColorScale, fmt.Sprintf("You cast %v", spellName))

            geom2 := geom
            geom2.Translate(5, 28)
            drawCityScape(screen, buildingSlots, buildinglib.BuildingNone, buildinglib.BuildingNone, ui.Counter / 8, &imageCache, fonts, city.BuildingInfo, player, city.Enchantments.Values(), geom2, getAlpha())
        },
    }

    getAlpha = ui.MakeFadeIn(fadeSpeed)

    ui.SetElementsFromArray(nil)

    // just to silence log warning
    ui.AddElement(&uilib.UIElement{
    })

    return ui, quit, nil
}
