package unitview

import (
    "log"
    "fmt"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/util"

    "github.com/hajimehoshi/ebiten/v2"
)

func MakeUnitContextMenu(cache *lbx.LbxCache, ui *uilib.UI, unit *units.OverworldUnit) []*uilib.UIElement {
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
            options.GeoM.Translate(31, 6)
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(background, &options)

            options.GeoM.Translate(25, 30)
            RenderCombatImage(screen, &imageCache, &unit.Unit, options)

            options.GeoM.Reset()
            options.GeoM.Translate(31, 6)
            options.GeoM.Translate(51, 8)

            RenderUnitInfoNormal(screen, &imageCache, &unit.Unit, descriptionFont, smallFont, options)

            options.GeoM.Reset()
            options.GeoM.Translate(31, 6)
            options.GeoM.Translate(10, 50)
            RenderUnitInfoStats(screen, &imageCache, &unit.Unit, descriptionFont, smallFont, options)

            options.GeoM.Translate(0, 60)
            RenderUnitAbilities(screen, &imageCache, &unit.Unit, mediumFont, options)
        },
    })

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            box, _ := imageCache.GetImage("unitview.lbx", 2, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(248, 139)
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(box, &options)
        },
    })

    buttonBackgrounds, _ := imageCache.GetImages("backgrnd.lbx", 24)
    // dismiss button
    cancelRect := util.ImageRect(257, 149, buttonBackgrounds[0])
    cancelIndex := 0
    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Rect: cancelRect,
        LeftClick: func(this *uilib.UIElement){
            cancelIndex = 1

            var confirmElements []*uilib.UIElement

            yes := func(){
                ui.RemoveElements(elements)
                // FIXME: disband unit
            }

            no := func(){
            }

            confirmElements = uilib.MakeConfirmDialogWithLayer(ui, cache, &imageCache, 2, fmt.Sprintf("Do you wish to disband the unit of %v?", unit.Unit.Name), yes, no)

            ui.AddElements(confirmElements)
        },
        LeftClickRelease: func(this *uilib.UIElement){
            cancelIndex = 0
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(cancelRect.Min.X), float64(cancelRect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(buttonBackgrounds[cancelIndex], &options)

            x := float64(cancelRect.Min.X + cancelRect.Max.X) / 2
            y := float64(cancelRect.Min.Y + cancelRect.Max.Y) / 2
            okDismissFont.PrintCenter(screen, x, y - 5, 1, options.ColorScale, "Dismiss")
        },
    })

    okRect := util.ImageRect(257, 169, buttonBackgrounds[0])
    okIndex := 0
    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Rect: okRect,
        LeftClick: func(this *uilib.UIElement){
            okIndex = 1
        },
        LeftClickRelease: func(this *uilib.UIElement){
            getAlpha = ui.MakeFadeOut(fadeSpeed)

            ui.AddDelay(fadeSpeed, func(){
                ui.RemoveElements(elements)
            })
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(okRect.Min.X), float64(okRect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(buttonBackgrounds[okIndex], &options)

            x := float64(okRect.Min.X + okRect.Max.X) / 2
            y := float64(okRect.Min.Y + okRect.Max.Y) / 2
            okDismissFont.PrintCenter(screen, x, y - 5, 1, options.ColorScale, "Ok")
        },
    })

    return elements
}
