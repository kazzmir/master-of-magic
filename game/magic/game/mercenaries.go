package game

import (
    "fmt"

    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/unitview"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"

    "github.com/hajimehoshi/ebiten/v2"
)

func MakeHireMercenariesScreenUI(cache *lbx.LbxCache, ui *uilib.UI, unit *units.OverworldUnit, count int, goldToHire int, action func(bool)) *uilib.UIElementGroup {
    imageCache := util.MakeImageCache(cache)

    yTop := float64(10)

    fonts := fontslib.MakeMercenariesFonts(cache)

    var elements []*uilib.UIElement

    const fadeSpeed = 7

    getAlpha := ui.MakeFadeIn(fadeSpeed)

    uiGroup := uilib.MakeGroup()

    uiGroup.AddElement(&uilib.UIElement{
        Layer: 1,
        Order: -1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("unitview.lbx", 1, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(float64(31), float64(6))
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, background, &options)

            options.GeoM.Translate(float64(24), float64(28))
            unitview.RenderUnitViewImage(screen, &imageCache, unit, options, false, 0)

            options.GeoM.Reset()
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(float64(31), float64(6))
            options.GeoM.Translate(float64(51), float64(7))
            unitview.RenderUnitInfoNormal(screen, &imageCache, unit, "", unit.Unit.Race.String(), fonts.DescriptionFont, fonts.SmallFont, options)

            /*
            options.GeoM.Reset()
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(float64(31), float64(6))
            options.GeoM.Translate(float64(10), float64(50))
            unitview.RenderUnitInfoStats(screen, &imageCache, unit, 15, fonts.DescriptionFont, fonts.SmallFont, options)
            */
        },
    })

    var statsOptions ebiten.DrawImageOptions
    statsOptions.GeoM.Translate(0, yTop)
    statsOptions.GeoM.Translate(float64(31), float64(6))
    statsOptions.GeoM.Translate(float64(10), float64(50))

    uiGroup.AddElements(unitview.CreateUnitInfoStatsElements(&imageCache, unit, 15, fonts.DescriptionFont, fonts.SmallFont, statsOptions, &getAlpha, 1))

    uiGroup.AddElements(unitview.MakeUnitAbilitiesElements(uiGroup, cache, &imageCache, unit, fonts.MediumFont, 40, 124, &ui.Counter, 1, &getAlpha, false, 0, false))

    uiGroup.AddElement(&uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            box, _ := imageCache.GetImage("unitview.lbx", 2, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(float64(248), float64(139))
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, box, &options)
        },
    })

    buttonBackgrounds, _ := imageCache.GetImages("backgrnd.lbx", 24)

    hireRect := util.ImageRect(257, 149 + int(yTop), buttonBackgrounds[0])
    hireIndex := 0
    uiGroup.AddElement(&uilib.UIElement{
        Layer: 1,
        Rect: hireRect,
        LeftClick: func(this *uilib.UIElement){
            hireIndex = 1
        },
        LeftClickRelease: func(this *uilib.UIElement){
            hireIndex = 0
            getAlpha = ui.MakeFadeOut(fadeSpeed)
            ui.AddDelay(fadeSpeed, func(){
                ui.RemoveElements(elements)
                ui.RemoveGroup(uiGroup)
                action(true)
            })
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(hireRect.Min.X), float64(hireRect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, buttonBackgrounds[hireIndex], &options)

            x := float64(hireRect.Min.X + hireRect.Max.X) / 2
            y := float64(hireRect.Min.Y + hireRect.Max.Y) / 2
            fonts.OkDismissFont.PrintOptions(screen, x, y - float64(5), font.FontOptions{Options: &options, Justify: font.FontJustifyCenter, Scale: scale.ScaleAmount}, "Hire")
        },
    })

    rejectRect := util.ImageRect(257, 169 + int(yTop), buttonBackgrounds[0])
    rejectIndex := 0
    uiGroup.AddElement(&uilib.UIElement{
        Layer: 1,
        Rect: rejectRect,
        LeftClick: func(this *uilib.UIElement){
            rejectIndex = 1
        },
        LeftClickRelease: func(this *uilib.UIElement){
            getAlpha = ui.MakeFadeOut(fadeSpeed)

            ui.AddDelay(fadeSpeed, func(){
                ui.RemoveElements(elements)
                ui.RemoveGroup(uiGroup)
                action(false)
            })
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(rejectRect.Min.X), float64(rejectRect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, buttonBackgrounds[rejectIndex], &options)

            x := float64(rejectRect.Min.X + rejectRect.Max.X) / 2
            y := float64(rejectRect.Min.Y + rejectRect.Max.Y) / 2
            fonts.OkDismissFont.PrintOptions(screen, x, y - float64(5), font.FontOptions{Options: &options, Justify: font.FontJustifyCenter, Scale: scale.ScaleAmount}, "Reject")
        },
    })

    uiGroup.AddElement(&uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            banner, _ := imageCache.GetImage("hire.lbx", 0, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(0, 0)
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, banner, &options)

            message := fmt.Sprintf("Mercenaries for Hire: %v gold", goldToHire)
            if count > 1 {
                message = fmt.Sprintf("%v Mercenaries for Hire: %v gold", count, goldToHire)
            }
            fonts.OkDismissFont.PrintOptions(screen, float64(135), float64(6), font.FontOptions{Options: &options, Justify: font.FontJustifyCenter, Scale: scale.ScaleAmount}, message)
        },
    })

    return uiGroup
}
