package game

import (
    "fmt"
    "image/color"
    "log"

    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/unitview"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
)

func MakeHireMercenariesScreenUI(cache *lbx.LbxCache, ui *uilib.UI, unit *units.OverworldUnit, count int, goldToHire int, action func(bool)) []*uilib.UIElement {
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

    yTop := float64(10)

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

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("unitview.lbx", 1, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(31, 6)
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(background, &options)

            options.GeoM.Translate(24, 28)
            unitview.RenderCombatImage(screen, &imageCache, unit, options, 0)

            options.GeoM.Reset()
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(31, 6)
            options.GeoM.Translate(51, 7)
            unitview.RenderUnitInfoNormal(screen, &imageCache, unit, "", unit.Unit.Race.String(), descriptionFont, smallFont, options)

            options.GeoM.Reset()
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(31, 6)
            options.GeoM.Translate(10, 50)
            unitview.RenderUnitInfoStats(screen, &imageCache, unit, 15, descriptionFont, smallFont, options)
        },
    })

    elements = append(elements, unitview.MakeUnitAbilitiesElements(&imageCache, unit, mediumFont, 40, 124, 1, &getAlpha, false)...)

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            box, _ := imageCache.GetImage("unitview.lbx", 2, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(248, 139)
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(box, &options)
        },
    })

    buttonBackgrounds, _ := imageCache.GetImages("backgrnd.lbx", 24)

    hireRect := util.ImageRect(257, 149 + int(yTop), buttonBackgrounds[0])
    hireIndex := 0
    elements = append(elements, &uilib.UIElement{
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
            okDismissFont.PrintCenter(screen, x, y - 5, 1, options.ColorScale, "Hire")
        },
    })

    rejectRect := util.ImageRect(257, 169 + int(yTop), buttonBackgrounds[0])
    rejectIndex := 0
    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Rect: rejectRect,
        LeftClick: func(this *uilib.UIElement){
            rejectIndex = 1
        },
        LeftClickRelease: func(this *uilib.UIElement){
            getAlpha = ui.MakeFadeOut(fadeSpeed)

            ui.AddDelay(fadeSpeed, func(){
                ui.RemoveElements(elements)
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
            okDismissFont.PrintCenter(screen, x, y - 5, 1, options.ColorScale, "Reject")
        },
    })

    elements = append(elements, &uilib.UIElement{
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
            okDismissFont.PrintCenter(screen, 135, 6, 1, options.ColorScale, message)
        },
    })

    return elements
}
