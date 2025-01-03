package game

import (
    "log"
    "fmt"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/artifactview"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/hajimehoshi/ebiten/v2"
)

func MakeMerchantScreenUI(cache *lbx.LbxCache, ui *uilib.UI, artifact *artifact.Artifact, goldToBuy int, action func(bool)) []*uilib.UIElement {
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

    lightPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0},
        color.RGBA{R: 0xed, G: 0xa4, B: 0x00, A: 0xff},
        color.RGBA{R: 0xff, G: 0xbc, B: 0x00, A: 0xff},
        color.RGBA{R: 0xff, G: 0xd6, B: 0x11, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
    }
    darkPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        util.Lighten(color.RGBA{R: 0xc7, G: 0x82, B: 0x1b, A: 0xff}, 0),
        util.Lighten(color.RGBA{R: 0xc7, G: 0x82, B: 0x1b, A: 0xff}, 20),
        util.Lighten(color.RGBA{R: 0xc7, G: 0x82, B: 0x1b, A: 0xff}, 50),
        util.Lighten(color.RGBA{R: 0xc7, G: 0x82, B: 0x1b, A: 0xff}, 80),
        color.RGBA{R: 0xc7, G: 0x82, B: 0x1b, A: 0xff},
        color.RGBA{R: 0xc7, G: 0x82, B: 0x1b, A: 0xff},
    }

    lightFont := font.MakeOptimizedFontWithPalette(fonts[4], lightPalette)
    darkFont := font.MakeOptimizedFontWithPalette(fonts[4], darkPalette)

    var elements []*uilib.UIElement

    const fadeSpeed = 7

    getAlpha := ui.MakeFadeIn(fadeSpeed)

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("hire.lbx", 2, 0)
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(background, &options)
        },
    })

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            colorScale := ebiten.ColorScale{}
            colorScale.ScaleAlpha(getAlpha())
            text := fmt.Sprintf("A merchant arrives and offers a magic %v for sale. The price is only %v gold pieces.", artifact.Name, goldToBuy)
            lightFont.PrintWrap(screen, 56, 8, 180, 1, colorScale, text)
        },
    })

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            options.GeoM.Translate(14, 65)
            artifactview.RenderArtifactBox(screen, &imageCache, *artifact, darkFont, options)
        },
    })

    buttonBackgrounds, _ := imageCache.GetImages("backgrnd.lbx", 24)
    buyRect := util.ImageRect(252, 121, buttonBackgrounds[0])
    buyIndex := 0
    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Rect: buyRect,
        LeftClick: func(this *uilib.UIElement){
            buyIndex = 1
        },
        LeftClickRelease: func(this *uilib.UIElement){
            buyIndex = 0
            getAlpha = ui.MakeFadeOut(fadeSpeed)
            ui.AddDelay(fadeSpeed, func(){
                ui.RemoveElements(elements)
                action(true)
            })
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(buyRect.Min.X), float64(buyRect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(buttonBackgrounds[buyIndex], &options)

            x := float64(buyRect.Min.X + buyRect.Max.X) / 2
            y := float64(buyRect.Min.Y + buyRect.Max.Y) / 2
            lightFont.PrintCenter(screen, x, y - 5, 1, options.ColorScale, "Buy")
        },
    })

    rejectRect := util.ImageRect(252, 140, buttonBackgrounds[0])
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
            lightFont.PrintCenter(screen, x, y - 5, 1, options.ColorScale, "Reject")
        },
    })

    return elements
}
