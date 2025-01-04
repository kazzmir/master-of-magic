package cityview

import (
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"

    "github.com/hajimehoshi/ebiten/v2"
)

func MakeEnchantmentView(cache *lbx.LbxCache, city *citylib.City, player *playerlib.Player, spellName string) (*uilib.UI, error) {
    imageCache := util.MakeImageCache(cache)

    background, _ := imageCache.GetImage("spellscr.lbx", 73, 0)
    buildingSlots := makeBuildingSlots(city)
    fonts, err := makeFonts(cache)
    if err != nil {
        return nil, err
    }

    geom := ebiten.GeoM{}
    geom.Translate(30, 30)

    // getAlpha := ui.MakeFadeIn(fadeSpeed)

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions

            options.GeoM = geom

            screen.DrawImage(background, &options)

            geom2 := geom
            geom2.Translate(5, 28)
            drawCityScape(screen, buildingSlots, buildinglib.BuildingNone, buildinglib.BuildingNone, ui.Counter / 8, &imageCache, fonts, city.BuildingInfo, player, geom2, 1.0)
        },
    }

    ui.SetElementsFromArray(nil)

    // just to silence log warning
    ui.AddElement(&uilib.UIElement{
    })

    return ui, nil
}
