package game

import (
    "fmt"
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/hajimehoshi/ebiten/v2"
)

type MerchantFonts struct {
    LightFont *font.Font
}

func MakeMerchantFonts(cache *lbx.LbxCache) *MerchantFonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    return &MerchantFonts{
        LightFont: loader(fontslib.LightFont),
    }
}

func MakeMerchantScreenUI(cache *lbx.LbxCache, ui *uilib.UI, artifactToBuy *artifact.Artifact, goldToBuy int, action func(bool)) []*uilib.UIElement {
    imageCache := util.MakeImageCache(cache)

    fonts := MakeMerchantFonts(cache)
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
            options.GeoM.Translate(float64(4), float64(15))
            scale.DrawScaled(screen, background, &options)
        },
    })

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            text := fmt.Sprintf("A merchant arrives and offers a magic %v for sale. The price is only %v gold pieces.", artifactToBuy.Name, goldToBuy)
            fonts.LightFont.PrintWrap(screen, float64(60), float64(23), float64(180), font.FontOptions{Scale: scale.ScaleAmount, Options: &options}, text)
        },
    })

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            options.GeoM.Translate(float64(18), float64(80))
            artifact.RenderArtifactBox(screen, &imageCache, *artifactToBuy, ui.Counter / 8, vaultFonts.ItemName, vaultFonts.PowerFont, options)
        },
    })

    buttonBackgrounds, _ := imageCache.GetImages("backgrnd.lbx", 24)
    buyRect := util.ImageRect(256, 136, buttonBackgrounds[0])
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
            scale.DrawScaled(screen, buttonBackgrounds[buyIndex], &options)

            x := float64(buyRect.Min.X + buyRect.Max.X) / 2
            y := float64(buyRect.Min.Y + buyRect.Max.Y) / 2
            fonts.LightFont.PrintOptions(screen, x, y - float64(5), font.FontOptions{Justify: font.FontJustifyCenter, Options: &options, Scale: scale.ScaleAmount}, "Buy")
        },
    })

    rejectRect := util.ImageRect(256, 155, buttonBackgrounds[0])
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
            scale.DrawScaled(screen, buttonBackgrounds[rejectIndex], &options)

            x := float64(rejectRect.Min.X + rejectRect.Max.X) / 2
            y := float64(rejectRect.Min.Y + rejectRect.Max.Y) / 2
            fonts.LightFont.PrintOptions(screen, x, y - float64(5), font.FontOptions{Justify: font.FontJustifyCenter, Options: &options, Scale: scale.ScaleAmount}, "Reject")
        },
    })

    return elements
}
