package ui

import (
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/hajimehoshi/ebiten/v2"
)

func MakeHelpElement(ui *UI, cache *lbx.LbxCache, imageCache *util.ImageCache, help lbx.HelpEntry, helpEntries ...lbx.HelpEntry) *UIElement {

    helpTop, err := imageCache.GetImage("help.lbx", 0, 0)
    if err != nil {
        return nil
    }
        
    fontLbx, err := cache.GetLbxFile("FONTS.LBX")
    if err != nil {
        return nil
    }

    fonts, err := fontLbx.ReadFonts(0)
    if err != nil {
        return nil
    }

    helpPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0x5e, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
    }

    helpFont := font.MakeOptimizedFontWithPalette(fonts[1], helpPalette)

    titleRed := color.RGBA{R: 0x50, G: 0x00, B: 0x0e, A: 0xff}
    titlePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        titleRed,
        titleRed,
        titleRed,
        titleRed,
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
    }

    helpTitleFont := font.MakeOptimizedFontWithPalette(fonts[4], titlePalette)

    infoX := 55
    infoY := 30
    infoWidth := helpTop.Bounds().Dx()
    // infoHeight := screen.HelpTop.Bounds().Dy()
    infoLeftMargin := 18
    infoTopMargin := 26
    infoBodyMargin := 3
    maxInfoWidth := infoWidth - infoLeftMargin - infoBodyMargin - 15

    // fmt.Printf("Help text: %v\n", []byte(help.Text))

    wrapped := helpFont.CreateWrappedText(float64(maxInfoWidth), 1, help.Text)

    helpTextY := infoY + infoTopMargin
    titleYAdjust := 0

    var extraImage *ebiten.Image
    if help.Lbx != "" {
        // fmt.Printf("Load extra image from %v index %v\n", help.Lbx, help.LbxIndex)
        use, err := imageCache.GetImageTransform(help.Lbx, help.LbxIndex, 0, util.AutoCrop)
        if err == nil && use != nil {
            extraImage = use
        }
    }

    if extraImage != nil {
        titleYAdjust = extraImage.Bounds().Dy() / 2 - helpTitleFont.Height() / 2

        if extraImage.Bounds().Dy() > helpTitleFont.Height() {
            helpTextY += extraImage.Bounds().Dy() + 1
        } else {
            helpTextY += helpTitleFont.Height() + 1
        }
    } else {
        helpTextY += helpTitleFont.Height() + 1
    }

    bottom := float64(helpTextY) + wrapped.TotalHeight

    // only draw as much of the top scroll as there are lines of text
    topImage := helpTop.SubImage(image.Rect(0, 0, helpTop.Bounds().Dx(), int(bottom) - infoY)).(*ebiten.Image)
    helpBottom, err := imageCache.GetImage("help.lbx", 1, 0)
    if err != nil {
        return nil
    }

    infoElement := &UIElement{
        // Rect: image.Rect(infoX, infoY, infoX + infoWidth, infoY + infoHeight),
        Rect: image.Rect(0, 0, data.ScreenWidth, data.ScreenHeight),
        Draw: func (infoThis *UIElement, window *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(infoX), float64(infoY))
            window.DrawImage(topImage, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(infoX), float64(bottom))
            window.DrawImage(helpBottom, &options)

            // for debugging
            // vector.StrokeRect(window, float32(infoX), float32(infoY), float32(infoWidth), float32(infoHeight), 1, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, true)
            // vector.StrokeRect(window, float32(infoX + infoLeftMargin), float32(infoY + infoTopMargin), float32(maxInfoWidth), float32(screen.HelpTitleFont.Height() + 20 + 1), 1, color.RGBA{R: 0, G: 0xff, B: 0, A: 0xff}, false)

            titleX := infoX + infoLeftMargin

            if extraImage != nil {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(titleX), float64(infoY + infoTopMargin))
                window.DrawImage(extraImage, &options)
                titleX += extraImage.Bounds().Dx() + 5
            }

            helpTitleFont.Print(window, float64(titleX), float64(infoY + infoTopMargin + titleYAdjust), 1, help.Headline)
            helpFont.RenderWrapped(window, float64(infoX + infoLeftMargin + infoBodyMargin), float64(helpTextY), wrapped, false)
        },
        LeftClick: func(infoThis *UIElement){
            ui.RemoveElement(infoThis)
        },
        Layer: 1,
    }

    return infoElement
}
