package game

import (
    "fmt"
    "image/color"
    "log"

    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/unitview"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
)

func MakeHireMercenariesScreenUI(cache *lbx.LbxCache, ui *uilib.UI, unit *units.OverworldUnit, count int, goldToHire int, action func(bool)) *uilib.UIElementGroup {
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

    imageCache := util.MakeImageCache(cache)

    yTop := float64(10 * data.ScreenScale)

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
    smallFont := font.MakeOptimizedFontWithPalette(fonts[1], descriptionPalette)
    mediumFont := font.MakeOptimizedFontWithPalette(fonts[2], descriptionPalette)

    yellowGradient := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0},
        color.RGBA{R: 0xed, G: 0xa4, B: 0x00, A: 0xff},
        color.RGBA{R: 0xff, G: 0xbc, B: 0x00, A: 0xff},
        color.RGBA{R: 0xff, G: 0xd6, B: 0x11, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
    }

    okDismissFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowGradient)

    var elements []*uilib.UIElement

    const fadeSpeed = 7

    getAlpha := ui.MakeFadeIn(fadeSpeed)

    uiGroup := uilib.MakeGroup()

    uiGroup.AddElement(&uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("unitview.lbx", 1, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(float64(31 * data.ScreenScale), float64(6 * data.ScreenScale))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(background, &options)

            options.GeoM.Translate(float64(24 * data.ScreenScale), float64(28 * data.ScreenScale))
            unitview.RenderCombatImage(screen, &imageCache, unit, options, 0)

            options.GeoM.Reset()
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(float64(31 * data.ScreenScale), float64(6 * data.ScreenScale))
            options.GeoM.Translate(float64(51 * data.ScreenScale), float64(7 * data.ScreenScale))
            unitview.RenderUnitInfoNormal(screen, &imageCache, unit, "", unit.Unit.Race.String(), descriptionFont, smallFont, options)

            options.GeoM.Reset()
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(float64(31 * data.ScreenScale), float64(6 * data.ScreenScale))
            options.GeoM.Translate(float64(10 * data.ScreenScale), float64(50 * data.ScreenScale))
            unitview.RenderUnitInfoStats(screen, &imageCache, unit, 15, descriptionFont, smallFont, options)
        },
    })

    uiGroup.AddElements(unitview.MakeUnitAbilitiesElements(uiGroup, &imageCache, unit, mediumFont, 40 * data.ScreenScale, 124 * data.ScreenScale, &ui.Counter, 1, &getAlpha, false))

    uiGroup.AddElement(&uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            box, _ := imageCache.GetImage("unitview.lbx", 2, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(float64(248 * data.ScreenScale), float64(139 * data.ScreenScale))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(box, &options)
        },
    })

    buttonBackgrounds, _ := imageCache.GetImages("backgrnd.lbx", 24)

    hireRect := util.ImageRect(257 * data.ScreenScale, 149 * data.ScreenScale + int(yTop), buttonBackgrounds[0])
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
            screen.DrawImage(buttonBackgrounds[hireIndex], &options)

            x := float64(hireRect.Min.X + hireRect.Max.X) / 2
            y := float64(hireRect.Min.Y + hireRect.Max.Y) / 2
            okDismissFont.PrintCenter(screen, x, y - float64(5 * data.ScreenScale), float64(data.ScreenScale), options.ColorScale, "Hire")
        },
    })

    rejectRect := util.ImageRect(257 * data.ScreenScale, 169 * data.ScreenScale + int(yTop), buttonBackgrounds[0])
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
            screen.DrawImage(buttonBackgrounds[rejectIndex], &options)

            x := float64(rejectRect.Min.X + rejectRect.Max.X) / 2
            y := float64(rejectRect.Min.Y + rejectRect.Max.Y) / 2
            okDismissFont.PrintCenter(screen, x, y - float64(5 * data.ScreenScale), float64(data.ScreenScale), options.ColorScale, "Reject")
        },
    })

    uiGroup.AddElement(&uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            banner, _ := imageCache.GetImage("hire.lbx", 0, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(0, 0)
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(banner, &options)

            message := fmt.Sprintf("Mercenaries for Hire: %v gold", goldToHire)
            if count > 1 {
                message = fmt.Sprintf("%v Mercenaries for Hire: %v gold", count, goldToHire)
            }
            okDismissFont.PrintCenter(screen, float64(135 * data.ScreenScale), float64(6 * data.ScreenScale), float64(data.ScreenScale), options.ColorScale, message)
        },
    })

    return uiGroup
}
