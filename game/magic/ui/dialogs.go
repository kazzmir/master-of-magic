package ui

import (
    // "log"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/hajimehoshi/ebiten/v2"
)

func MakeHelpElement(ui *UI, cache *lbx.LbxCache, imageCache *util.ImageCache, help lbx.HelpEntry, helpEntries ...lbx.HelpEntry) *UIElement {
    return MakeHelpElementWithLayer(ui, cache, imageCache, UILayer(1), help, helpEntries...)
}

func MakeHelpElementWithLayer(ui *UI, cache *lbx.LbxCache, imageCache *util.ImageCache, layer UILayer, help lbx.HelpEntry, helpEntries ...lbx.HelpEntry) *UIElement {

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
    // infoY := 30
    infoWidth := helpTop.Bounds().Dx()
    // infoHeight := screen.HelpTop.Bounds().Dy()
    infoLeftMargin := 18
    infoTopMargin := 26
    infoBodyMargin := 3
    maxInfoWidth := infoWidth - infoLeftMargin - infoBodyMargin - 14

    // fmt.Printf("Help text: %v\n", []byte(help.Text))

    wrapped := helpFont.CreateWrappedText(float64(maxInfoWidth), 1, help.Text)

    helpTextY := infoTopMargin
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

    var moreHelp []font.WrappedText

    // add in more help entries
    for _, entry := range helpEntries {
        bottom += 2
        bottom += float64(helpTitleFont.Height()) + 1
        moreWrapped := helpFont.CreateWrappedText(float64(maxInfoWidth), 1, entry.Text)
        moreHelp = append(moreHelp, moreWrapped)
        bottom += moreWrapped.TotalHeight
    }

    // only draw as much of the top scroll as there are lines of text
    topImage := helpTop.SubImage(image.Rect(0, 0, helpTop.Bounds().Dx(), int(bottom))).(*ebiten.Image)
    helpBottom, err := imageCache.GetImage("help.lbx", 1, 0)
    if err != nil {
        return nil
    }

    infoY := (float64(data.ScreenHeight) - bottom - float64(helpBottom.Bounds().Dy())) / 2

    infoElement := &UIElement{
        // Rect: image.Rect(infoX, infoY, infoX + infoWidth, infoY + infoHeight),
        Rect: image.Rect(0, 0, data.ScreenWidth, data.ScreenHeight),
        Draw: func (infoThis *UIElement, window *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(infoX), float64(infoY))
            window.DrawImage(topImage, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(infoX), float64(bottom) + infoY)
            window.DrawImage(helpBottom, &options)

            // for debugging
            // vector.StrokeRect(window, float32(infoX), float32(infoY), float32(infoWidth), float32(infoHeight), 1, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, true)
            // vector.StrokeRect(window, float32(infoX + infoLeftMargin), float32(infoY + infoTopMargin), float32(maxInfoWidth), float32(screen.HelpTitleFont.Height() + 20 + 1), 1, color.RGBA{R: 0, G: 0xff, B: 0, A: 0xff}, false)

            titleX := infoX + infoLeftMargin

            if extraImage != nil {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(titleX), infoY + float64(infoTopMargin))
                window.DrawImage(extraImage, &options)
                titleX += extraImage.Bounds().Dx() + 5
            }

            helpTitleFont.Print(window, float64(titleX), infoY + float64(infoTopMargin + titleYAdjust), 1, help.Headline)
            helpFont.RenderWrapped(window, float64(infoX + infoLeftMargin + infoBodyMargin), float64(helpTextY) + infoY, wrapped, false)

            yPos := float64(helpTextY) + infoY + wrapped.TotalHeight + 2
            for i, moreWrapped := range moreHelp {
                helpTitleFont.Print(window, float64(titleX), float64(yPos), 1, helpEntries[i].Headline)
                helpFont.RenderWrapped(window, float64(infoX + infoLeftMargin + infoBodyMargin), yPos + float64(helpTitleFont.Height()) + 1, moreWrapped, false)
                yPos += float64(helpTitleFont.Height()) + 1 + float64(moreWrapped.TotalHeight) + 2
            }

        },
        LeftClick: func(infoThis *UIElement){
            ui.RemoveElement(infoThis)
        },
        Layer: layer,
    }

    return infoElement
}

func MakeErrorElement(ui *UI, cache *lbx.LbxCache, imageCache *util.ImageCache, message string) *UIElement {
    errorX := 67
    errorY := 73

    errorMargin := 15
    errorTopMargin := 10

    errorTop, err := imageCache.GetImage("newgame.lbx", 44, 0)
    if err != nil {
        return nil
    }

    errorBottom, err := imageCache.GetImage("newgame.lbx", 45, 0)
    if err != nil {
        return nil
    }

    // FIXME: this should be a fade from bright yellow to dark yellow/orange
    yellowFade := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0xb2, G: 0x8c, B: 0x05, A: 0xff},
        color.RGBA{R: 0xc9, G: 0xa1, B: 0x26, A: 0xff},
        color.RGBA{R: 0xff, G: 0xd3, B: 0x5b, A: 0xff},
        color.RGBA{R: 0xff, G: 0xe8, B: 0x6f, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
    }

    fontLbx, err := cache.GetLbxFile("FONTS.LBX")
    if err != nil {
        return nil
    }

    fonts, err := fontLbx.ReadFonts(0)
    if err != nil {
        return nil
    }

    errorFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowFade)

    maxWidth := errorTop.Bounds().Dx() - errorMargin * 2

    wrapped := errorFont.CreateWrappedText(float64(maxWidth), 1, message)

    bottom := float64(errorY + errorTopMargin) + wrapped.TotalHeight

    topDraw := errorTop.SubImage(image.Rect(0, 0, errorTop.Bounds().Dx(), int(bottom) - errorY)).(*ebiten.Image)

    element := &UIElement{
        Rect: image.Rect(0, 0, data.ScreenWidth, data.ScreenHeight),
        Layer: 1,
        LeftClick: func(this *UIElement){
            ui.RemoveElement(this)
        },
        Draw: func(this *UIElement, window *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(errorX), float64(errorY))
            window.DrawImage(topDraw, &options)

            errorFont.RenderWrapped(window, float64(errorX + errorMargin + maxWidth / 2), float64(errorY + errorTopMargin), wrapped, true)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(errorX), float64(bottom))
            window.DrawImage(errorBottom, &options)
        },
    }

    return element
}

func MakeConfirmDialog(ui *UI, cache *lbx.LbxCache, imageCache *util.ImageCache, message string, confirm func(), cancel func()) []*UIElement {
    confirmX := 67
    confirmY := 73

    confirmMargin := 15
    confirmTopMargin := 10

    confirmTop, err := imageCache.GetImage("resource.lbx", 0, 0)
    if err != nil {
        return nil
    }

    confirmBottom, err := imageCache.GetImage("resource.lbx", 1, 0)
    if err != nil {
        return nil
    }

    // FIXME: this should be a fade from bright yellow to dark yellow/orange
    yellowFade := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0xb2, G: 0x8c, B: 0x05, A: 0xff},
        color.RGBA{R: 0xc9, G: 0xa1, B: 0x26, A: 0xff},
        color.RGBA{R: 0xff, G: 0xd3, B: 0x5b, A: 0xff},
        color.RGBA{R: 0xff, G: 0xe8, B: 0x6f, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
    }

    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        return nil
    }

    fonts, err := fontLbx.ReadFonts(0)
    if err != nil {
        return nil
    }

    confirmFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowFade)

    maxWidth := confirmTop.Bounds().Dx() - confirmMargin * 2

    wrapped := confirmFont.CreateWrappedText(float64(maxWidth), 1, message)

    bottom := float64(confirmY + confirmTopMargin) + wrapped.TotalHeight

    topDraw := confirmTop.SubImage(image.Rect(0, 0, confirmTop.Bounds().Dx(), int(bottom) - confirmY)).(*ebiten.Image)

    var elements []*UIElement

    elements = append(elements, &UIElement{
        Rect: image.Rect(0, 0, data.ScreenWidth, data.ScreenHeight),
        Layer: 1,
        LeftClick: func(this *UIElement){
            // ui.RemoveElement(this)
        },
        Draw: func(this *UIElement, window *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(confirmX), float64(confirmY))
            window.DrawImage(topDraw, &options)

            confirmFont.RenderWrapped(window, float64(confirmX + confirmMargin + maxWidth / 2), float64(confirmY + confirmTopMargin), wrapped, true)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(confirmX), float64(bottom))
            window.DrawImage(confirmBottom, &options)
        },
    })

    // add yes/no buttons
    yesButtons, err := imageCache.GetImages("resource.lbx", 3)
    if err == nil {
        yesX := confirmX + 101
        yesY := bottom + 5

        clicked := false
        elements = append(elements, &UIElement{
            Rect: image.Rect(int(yesX), int(yesY), int(yesX) + yesButtons[0].Bounds().Dx(), int(yesY) + yesButtons[0].Bounds().Dy()),
            Layer: 1,
            LeftClick: func(this *UIElement){
                clicked = true
            },
            LeftClickRelease: func(this *UIElement){
                clicked = false
                confirm()
            },
            Draw: func(this *UIElement, window *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(yesX), float64(yesY))

                index := 0
                if clicked {
                    index = 1
                }
                window.DrawImage(yesButtons[index], &options)
            },
        })
    }

    noButtons, err := imageCache.GetImages("resource.lbx", 4)
    if err == nil {
        noX := confirmX + 18
        noY := bottom + 5

        clicked := false
        elements = append(elements, &UIElement{
            Rect: image.Rect(int(noX), int(noY), int(noX) + noButtons[0].Bounds().Dx(), int(noY) + noButtons[0].Bounds().Dy()),
            Layer: 1,
            LeftClick: func(this *UIElement){
                clicked = true
            },
            LeftClickRelease: func(this *UIElement){
                clicked = false
                cancel()
            },
            Draw: func(this *UIElement, window *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(noX), float64(noY))

                index := 0
                if clicked {
                    index = 1
                }
                window.DrawImage(noButtons[index], &options)
            },
        })
    }

    return elements
}
