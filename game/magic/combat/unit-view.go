package combat

import (
    "log"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/unitview"

    "github.com/hajimehoshi/ebiten/v2"
)

func MakeUnitView(cache *lbx.LbxCache, ui *uilib.UI, unit *ArmyUnit) *uilib.UIElementGroup {
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

    _ = descriptionFont
    _ = smallFont
    _ = mediumFont
    _ = yellowGradient

    group := uilib.MakeGroup()

    const fadeSpeed = 7
    getAlpha := ui.MakeFadeIn(fadeSpeed)

    group.AddElement(&uilib.UIElement{
        Layer: 1,
        NotLeftClicked: func(element *uilib.UIElement) {
            getAlpha = ui.MakeFadeOut(fadeSpeed)
            ui.AddDelay(fadeSpeed, func(){
                ui.RemoveGroup(group)
            })
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("unitview.lbx", 1, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(31 * data.ScreenScale), float64(6 * data.ScreenScale))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(background, &options)

            options.GeoM.Translate(float64(25 * data.ScreenScale), float64(30 * data.ScreenScale))
            var generic interface{} = unit
            portaitUnit, ok := generic.(unitview.PortraitUnit)
            if ok {
                lbxFile, index := portaitUnit.GetPortraitLbxInfo()
                portait, err := imageCache.GetImage(lbxFile, index, 0)
                if err == nil {
                    options.GeoM.Translate(0, float64(-7 * data.ScreenScale))
                    options.GeoM.Translate(float64(-portait.Bounds().Dx()/2), float64(-portait.Bounds().Dy()/2))
                    screen.DrawImage(portait, &options)
                }
            } else {
                unitview.RenderUnitViewImage(screen, &imageCache, unit, options, ui.Counter)
            }

            options.GeoM.Reset()
            options.GeoM.Translate(float64(31 * data.ScreenScale), float64(6 * data.ScreenScale))
            options.GeoM.Translate(float64(51 * data.ScreenScale), float64(6 * data.ScreenScale))

            // RenderUnitInfoNormal(screen, &imageCache, unit, unit.GetTitle(), "", descriptionFont, smallFont, options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(31 * data.ScreenScale), float64(6 * data.ScreenScale))
            options.GeoM.Translate(float64(10 * data.ScreenScale), float64(50 * data.ScreenScale))
            unitview.RenderUnitInfoStats(screen, &imageCache, unit, 15, descriptionFont, smallFont, options)

            /*
            options.GeoM.Translate(0, 60)
            RenderUnitAbilities(screen, &imageCache, unit, mediumFont, options, false, 0)
            */
        },
    })

    return group
}
