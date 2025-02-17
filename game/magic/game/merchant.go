package game

import (
    "log"
    "fmt"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/hajimehoshi/ebiten/v2"
)

func MakeMerchantScreenUI(cache *lbx.LbxCache, ui *uilib.UI, artifactToBuy *artifact.Artifact, goldToBuy int, action func(bool)) []*uilib.UIElement {
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

    lightFont := font.MakeOptimizedFontWithPalette(fonts[4], lightPalette)

    vaultFonts := fontslib.MakeVaultFonts(cache)

    var elements []*uilib.UIElement

    const fadeSpeed = 7

    getAlpha := ui.MakeFadeIn(fadeSpeed)

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("hire.lbx", 2, 0)
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            options.GeoM.Translate(float64(4 * data.ScreenScale), float64(15 * data.ScreenScale))
            screen.DrawImage(background, &options)
        },
    })

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            colorScale := ebiten.ColorScale{}
            colorScale.ScaleAlpha(getAlpha())
            text := fmt.Sprintf("A merchant arrives and offers a magic %v for sale. The price is only %v gold pieces.", artifactToBuy.Name, goldToBuy)
            lightFont.PrintWrap(screen, float64(60 * data.ScreenScale), float64(23 * data.ScreenScale), float64(180 * data.ScreenScale), float64(data.ScreenScale), colorScale, text)
        },
    })

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            options.GeoM.Translate(float64(18 * data.ScreenScale), float64(80 * data.ScreenScale))
            artifact.RenderArtifactBox(screen, &imageCache, *artifactToBuy, ui.Counter / 8, vaultFonts.ItemName, vaultFonts.PowerFont, options)
        },
    })

    buttonBackgrounds, _ := imageCache.GetImages("backgrnd.lbx", 24)
    buyRect := util.ImageRect(256 * data.ScreenScale, 136 * data.ScreenScale, buttonBackgrounds[0])
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
            lightFont.PrintCenter(screen, x, y - float64(5 * data.ScreenScale), float64(data.ScreenScale), options.ColorScale, "Buy")
        },
    })

    rejectRect := util.ImageRect(256 * data.ScreenScale, 155 * data.ScreenScale, buttonBackgrounds[0])
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
            lightFont.PrintCenter(screen, x, y - float64(5 * data.ScreenScale), float64(data.ScreenScale), options.ColorScale, "Reject")
        },
    })

    return elements
}
