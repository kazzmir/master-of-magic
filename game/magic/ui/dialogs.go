package ui

import (
    "log"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/help"
    "github.com/hajimehoshi/ebiten/v2"
)

func MakeHelpElement(ui *UI, cache *lbx.LbxCache, imageCache *util.ImageCache, help help.HelpEntry, helpEntries ...help.HelpEntry) *UIElement {
    return MakeHelpElementWithLayer(ui, cache, imageCache, UILayer(1), help, helpEntries...)
}

func MakeHelpElementWithLayer(ui *UI, cache *lbx.LbxCache, imageCache *util.ImageCache, layer UILayer, help help.HelpEntry, helpEntries ...help.HelpEntry) *UIElement {

    helpTop, err := imageCache.GetImage("help.lbx", 0, 0)
    if err != nil {
        return nil
    }

    fontLbx, err := cache.GetLbxFile("FONTS.LBX")
    if err != nil {
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        return nil
    }

    const fadeSpeed = 7

    getAlpha := ui.MakeFadeIn(fadeSpeed)

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
    maxInfoWidth := infoWidth / data.ScreenScale - infoLeftMargin - infoBodyMargin - 14

    // fmt.Printf("Help text: %v\n", []byte(help.Text))

    wrapped := helpFont.CreateWrappedText(float64(maxInfoWidth * data.ScreenScale), float64(data.ScreenScale), help.Text)

    helpTextY := infoTopMargin * data.ScreenScale
    titleYAdjust := 0

    var extraImage *ebiten.Image
    if help.Lbx != "" {
        // fmt.Printf("Load extra image from %v index %v\n", help.Lbx, help.LbxIndex)
        use, err := imageCache.GetImageTransform(help.Lbx, help.LbxIndex, 0, "crop", util.AutoCrop)
        if err == nil && use != nil {
            extraImage = use
        }
    }

    if extraImage != nil {
        titleYAdjust = extraImage.Bounds().Dy() / 2 - helpTitleFont.Height() * data.ScreenScale / 2

        if extraImage.Bounds().Dy() > helpTitleFont.Height() * data.ScreenScale {
            helpTextY += extraImage.Bounds().Dy() + 1
        } else {
            helpTextY += (helpTitleFont.Height() + 1) * data.ScreenScale
        }
    } else {
        helpTextY += (helpTitleFont.Height() + 1) * data.ScreenScale
    }

    bottom := float64(helpTextY) + wrapped.TotalHeight

    var moreHelp []font.WrappedText

    // add in more help entries
    for _, entry := range helpEntries {
        bottom += 2
        bottom += float64(helpTitleFont.Height() * data.ScreenScale) + 1
        moreWrapped := helpFont.CreateWrappedText(float64(maxInfoWidth * data.ScreenScale), float64(data.ScreenScale), entry.Text)
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
            options.GeoM.Translate(float64(infoX * data.ScreenScale), float64(infoY))
            options.ColorScale.ScaleAlpha(getAlpha())
            window.DrawImage(topImage, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(infoX * data.ScreenScale), float64(bottom) + infoY)
            options.ColorScale.ScaleAlpha(getAlpha())
            window.DrawImage(helpBottom, &options)

            // for debugging
            // vector.StrokeRect(window, float32(infoX), float32(infoY), float32(infoWidth), float32(infoHeight), 1, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, true)
            // vector.StrokeRect(window, float32(infoX + infoLeftMargin), float32(infoY + infoTopMargin), float32(maxInfoWidth), float32(screen.HelpTitleFont.Height() + 20 + 1), 1, color.RGBA{R: 0, G: 0xff, B: 0, A: 0xff}, false)

            titleX := (infoX + infoLeftMargin) * data.ScreenScale

            if extraImage != nil {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(titleX), infoY + float64(infoTopMargin * data.ScreenScale))
                options.ColorScale.ScaleAlpha(getAlpha())
                window.DrawImage(extraImage, &options)
                titleX += extraImage.Bounds().Dx() + 5 * data.ScreenScale
            }

            helpTitleFont.Print(window, float64(titleX), infoY + float64(infoTopMargin * data.ScreenScale + titleYAdjust), float64(data.ScreenScale), options.ColorScale, help.Headline)
            helpFont.RenderWrapped(window, float64(infoX + infoLeftMargin + infoBodyMargin) * float64(data.ScreenScale), float64(helpTextY) + infoY, wrapped, options.ColorScale, false)

            yPos := float64(helpTextY) + infoY + wrapped.TotalHeight + 2
            for i, moreWrapped := range moreHelp {
                helpTitleFont.Print(window, float64(titleX), float64(yPos), float64(data.ScreenScale), options.ColorScale, helpEntries[i].Headline)
                helpFont.RenderWrapped(window, float64(infoX + infoLeftMargin + infoBodyMargin) * float64(data.ScreenScale), yPos + float64(helpTitleFont.Height() * data.ScreenScale) + 1, moreWrapped, options.ColorScale, false)
                yPos += float64(helpTitleFont.Height() * data.ScreenScale) + 1 + float64(moreWrapped.TotalHeight) + 2
            }

        },
        LeftClick: func(infoThis *UIElement){
            getAlpha = ui.MakeFadeOut(fadeSpeed)
            ui.AddDelay(fadeSpeed, func(){
                ui.RemoveElement(infoThis)
            })
        },
        Layer: layer,
    }

    return infoElement
}

func MakeErrorElement(ui *UI, cache *lbx.LbxCache, imageCache *util.ImageCache, message string, clicked func()) *UIElement {
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

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        return nil
    }

    errorFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowFade)

    maxWidth := errorTop.Bounds().Dx() - errorMargin * 2 * data.ScreenScale

    wrapped := errorFont.CreateWrappedText(float64(maxWidth), float64(data.ScreenScale), message)

    bottom := float64(errorY + errorTopMargin) * float64(data.ScreenScale) + wrapped.TotalHeight

    topDraw := errorTop.SubImage(image.Rect(0, 0, errorTop.Bounds().Dx(), int(bottom) - errorY * data.ScreenScale)).(*ebiten.Image)

    element := &UIElement{
        Rect: image.Rect(0, 0, data.ScreenWidth, data.ScreenHeight),
        Layer: 1,
        LeftClick: func(this *UIElement){
            ui.RemoveElement(this)
            clicked()
        },
        Draw: func(this *UIElement, window *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(errorX * data.ScreenScale), float64(errorY * data.ScreenScale))
            window.DrawImage(topDraw, &options)

            errorFont.RenderWrapped(window, float64((errorX + errorMargin) * data.ScreenScale + maxWidth / 2), float64(errorY + errorTopMargin) * float64(data.ScreenScale), wrapped, ebiten.ColorScale{}, true)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(errorX * data.ScreenScale), float64(bottom))
            window.DrawImage(errorBottom, &options)
        },
    }

    return element
}

func MakeConfirmDialog(group *UIElementGroup, cache *lbx.LbxCache, imageCache *util.ImageCache, message string, center bool, confirm func(), cancel func()) []*UIElement {
    return MakeConfirmDialogWithLayer(group, cache, imageCache, 1, message, center, confirm, cancel)
}

func MakeConfirmDialogWithLayer(group *UIElementGroup, cache *lbx.LbxCache, imageCache *util.ImageCache, layer UILayer, message string, center bool, confirm func(), cancel func()) []*UIElement {
    // a button that says 'Yes'
    yesButtons, _ := imageCache.GetImages("resource.lbx", 3)

    // a button that says 'No'
    noButtons, _ := imageCache.GetImages("resource.lbx", 4)
    return MakeConfirmDialogWithLayerFull(group, cache, imageCache, layer, message, center, confirm, cancel, yesButtons, noButtons)
}

func MakeConfirmDialogWithLayerFull(group *UIElementGroup, cache *lbx.LbxCache, imageCache *util.ImageCache, layer UILayer, message string, center bool, confirm func(), cancel func(), yesButtons []*ebiten.Image, noButtons []*ebiten.Image) []*UIElement {
    confirmX := 67 * data.ScreenScale
    confirmY := 68 * data.ScreenScale

    confirmMargin := 15 * data.ScreenScale
    confirmTopMargin := 10 * data.ScreenScale

    const fadeSpeed = 7

    getAlpha := group.MakeFadeIn(fadeSpeed)

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

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        return nil
    }

    confirmFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowFade)

    maxWidth := confirmTop.Bounds().Dx() - confirmMargin * 2

    wrapped := confirmFont.CreateWrappedText(float64(maxWidth), float64(data.ScreenScale), message)

    bottom := float64(confirmY + confirmTopMargin) + wrapped.TotalHeight

    topDraw := confirmTop.SubImage(image.Rect(0, 0, confirmTop.Bounds().Dx(), int(bottom) - confirmY)).(*ebiten.Image)

    var elements []*UIElement

    elements = append(elements, &UIElement{
        Rect: image.Rect(0, 0, data.ScreenWidth, data.ScreenHeight),
        Layer: layer,
        LeftClick: func(this *UIElement){
            // ui.RemoveElement(this)
        },
        Draw: func(this *UIElement, window *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(confirmX), float64(confirmY))
            options.ColorScale.ScaleAlpha(getAlpha())
            window.DrawImage(topDraw, &options)

            if center {
                confirmFont.RenderWrapped(window, float64(confirmX + confirmMargin + maxWidth / 2), float64(confirmY + confirmTopMargin), wrapped, options.ColorScale, true)
            } else {
                confirmFont.RenderWrapped(window, float64(confirmX + confirmMargin), float64(confirmY + confirmTopMargin), wrapped, options.ColorScale, false)
            }

            options.GeoM.Reset()
            options.GeoM.Translate(float64(confirmX), float64(bottom))
            window.DrawImage(confirmBottom, &options)
        },
    })

    // add yes/no buttons
    if err == nil {
        yesX := confirmX + 101 * data.ScreenScale
        yesY := bottom + float64(5 * data.ScreenScale)

        clicked := false
        elements = append(elements, &UIElement{
            Rect: util.ImageRect(int(yesX), int(yesY), yesButtons[0]),
            Layer: layer,
            PlaySoundLeftClick: true,
            LeftClick: func(this *UIElement){
                clicked = true
            },
            LeftClickRelease: func(this *UIElement){
                clicked = false

                getAlpha = group.MakeFadeOut(fadeSpeed)
                group.AddDelay(fadeSpeed, func(){
                    group.RemoveElements(elements)
                    confirm()
                })
            },
            Draw: func(this *UIElement, window *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(yesX), float64(yesY))
                options.ColorScale.ScaleAlpha(getAlpha())

                index := 0
                if clicked {
                    index = 1
                }
                window.DrawImage(yesButtons[index], &options)
            },
        })
    }

    if err == nil {
        noX := confirmX + 18 * data.ScreenScale
        noY := bottom + float64(5 * data.ScreenScale)

        clicked := false
        elements = append(elements, &UIElement{
            Rect: util.ImageRect(int(noX), int(noY), noButtons[0]),
            Layer: layer,
            PlaySoundLeftClick: true,
            LeftClick: func(this *UIElement){
                clicked = true
            },
            LeftClickRelease: func(this *UIElement){
                clicked = false

                getAlpha = group.MakeFadeOut(fadeSpeed)
                group.AddDelay(fadeSpeed, func(){
                    group.RemoveElements(elements)
                    cancel()
                })
            },
            Draw: func(this *UIElement, window *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(noX), float64(noY))
                options.ColorScale.ScaleAlpha(getAlpha())

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

func MakeLairConfirmDialog(ui *UI, cache *lbx.LbxCache, imageCache *util.ImageCache, lairPicture *util.Animation, message string, confirm func(), cancel func()) []*UIElement {
    return MakeLairConfirmDialogWithLayer(ui, cache, imageCache, lairPicture, 1, message, confirm, cancel)
}

func MakeLairConfirmDialogWithLayer(ui *UI, cache *lbx.LbxCache, imageCache *util.ImageCache, lairPicture *util.Animation, layer UILayer, message string, confirm func(), cancel func()) []*UIElement {
    confirmX := 67 * data.ScreenScale
    confirmY := 40 * data.ScreenScale

    confirmMargin := 55 * data.ScreenScale
    confirmTopMargin := 10 * data.ScreenScale

    const fadeSpeed = 7

    getAlpha := ui.MakeFadeIn(fadeSpeed)

    confirmTop, err := imageCache.GetImage("backgrnd.lbx", 25, 0)
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

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        return nil
    }

    confirmFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowFade)

    maxWidth := confirmTop.Bounds().Dx() - confirmMargin - 5 * data.ScreenScale

    wrapped := confirmFont.CreateWrappedText(float64(maxWidth), float64(data.ScreenScale), message)

    bottom := float64(confirmY + confirmTopMargin) + wrapped.TotalHeight

    topDraw := confirmTop.SubImage(image.Rect(0, 0, confirmTop.Bounds().Dx(), int(bottom) - confirmY)).(*ebiten.Image)

    var elements []*UIElement

    elements = append(elements, &UIElement{
        Rect: image.Rect(0, 0, data.ScreenWidth, data.ScreenHeight),
        Layer: layer,
        LeftClick: func(this *UIElement){
            // ui.RemoveElement(this)
        },
        Draw: func(this *UIElement, window *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(confirmX), float64(confirmY))
            options.ColorScale.ScaleAlpha(getAlpha())
            window.DrawImage(topDraw, &options)

            options.GeoM.Translate(float64(7 * data.ScreenScale), float64(7 * data.ScreenScale))
            window.DrawImage(lairPicture.Frame(), &options)

            confirmFont.RenderWrapped(window, float64(confirmX + confirmMargin + maxWidth / 2), float64(confirmY + confirmTopMargin), wrapped, options.ColorScale, true)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(confirmX - 1 * data.ScreenScale), float64(bottom))
            window.DrawImage(confirmBottom, &options)
        },
    })

    // add yes/no buttons
    yesButtons, err := imageCache.GetImages("resource.lbx", 3)
    if err == nil {
        yesX := confirmX + 101 * data.ScreenScale
        yesY := bottom + float64(5 * data.ScreenScale)

        clicked := false
        elements = append(elements, &UIElement{
            Rect: util.ImageRect(int(yesX), int(yesY), yesButtons[0]),
            Layer: layer,
            PlaySoundLeftClick: true,
            LeftClick: func(this *UIElement){
                clicked = true
            },
            LeftClickRelease: func(this *UIElement){
                clicked = false

                getAlpha = ui.MakeFadeOut(fadeSpeed)
                ui.AddDelay(fadeSpeed, func(){
                    ui.RemoveElements(elements)
                    confirm()
                })
            },
            Draw: func(this *UIElement, window *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(yesX), float64(yesY))
                options.ColorScale.ScaleAlpha(getAlpha())

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
        noX := confirmX + 18 * data.ScreenScale
        noY := bottom + float64(5 * data.ScreenScale)

        clicked := false
        elements = append(elements, &UIElement{
            Rect: util.ImageRect(int(noX), int(noY), noButtons[0]),
            Layer: layer,
            PlaySoundLeftClick: true,
            LeftClick: func(this *UIElement){
                clicked = true
            },
            LeftClickRelease: func(this *UIElement){
                clicked = false

                getAlpha = ui.MakeFadeOut(fadeSpeed)
                ui.AddDelay(fadeSpeed, func(){
                    ui.RemoveElements(elements)
                    cancel()
                })
            },
            Draw: func(this *UIElement, window *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(noX), float64(noY))
                options.ColorScale.ScaleAlpha(getAlpha())

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

func MakeLairShowDialogWithLayer(ui *UI, cache *lbx.LbxCache, imageCache *util.ImageCache, lairPicture *util.Animation, layer UILayer, message string, dismiss func()) []*UIElement {
    confirmX := 67 * data.ScreenScale
    confirmY := 40 * data.ScreenScale

    confirmMargin := 55 * data.ScreenScale
    confirmTopMargin := 10 * data.ScreenScale

    const fadeSpeed = 7

    getAlpha := ui.MakeFadeIn(fadeSpeed)

    confirmTop, err := imageCache.GetImage("backgrnd.lbx", 25, 0)
    if err != nil {
        return nil
    }

    confirmBottom, err := imageCache.GetImage("backgrnd.lbx", 27, 0)
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

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        return nil
    }

    confirmFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowFade)

    maxWidth := confirmTop.Bounds().Dx() - confirmMargin - 5 * data.ScreenScale

    wrapped := confirmFont.CreateWrappedText(float64(maxWidth), float64(data.ScreenScale), message)

    bottom := float64(confirmY + confirmTopMargin) + max(wrapped.TotalHeight, float64(lairPicture.Frame().Bounds().Dy()))

    topDraw := confirmTop.SubImage(image.Rect(0, 0, confirmTop.Bounds().Dx(), int(bottom) - confirmY)).(*ebiten.Image)

    var elements []*UIElement

    elements = append(elements, &UIElement{
        Rect: image.Rect(0, 0, data.ScreenWidth, data.ScreenHeight),
        Layer: layer,
        LeftClick: func(this *UIElement){
            getAlpha = ui.MakeFadeOut(fadeSpeed)
            ui.AddDelay(fadeSpeed, func(){
                dismiss()
            })
        },
        Draw: func(this *UIElement, window *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(confirmX), float64(confirmY))
            options.ColorScale.ScaleAlpha(getAlpha())
            window.DrawImage(topDraw, &options)

            options.GeoM.Translate(float64(7 * data.ScreenScale), float64(7 * data.ScreenScale))
            window.DrawImage(lairPicture.Frame(), &options)

            confirmFont.RenderWrapped(window, float64(confirmX + confirmMargin + maxWidth / 2), float64(confirmY + confirmTopMargin), wrapped, options.ColorScale, true)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(confirmX), float64(bottom))
            window.DrawImage(confirmBottom, &options)
        },
    })

    return elements
}

type Selection struct {
    Name string
    Action func()
    Hotkey string
}

func MakeSelectionUI(ui *UI, lbxCache *lbx.LbxCache, imageCache *util.ImageCache, cornerX int, cornerY int, selectionTitle string, choices []Selection) []*UIElement {
    var elements []*UIElement

    fontLbx, err := lbxCache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    font4, err := font.ReadFont(fontLbx, 0, 4)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
    }

    fadeSpeed := uint64(6)

    getAlpha := ui.MakeFadeIn(fadeSpeed)

    blackPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
    }

    // FIXME: this is too bright
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

    buttonFont := font.MakeOptimizedFontWithPalette(font4, blackPalette)
    topFont := font.MakeOptimizedFontWithPalette(font4, yellowGradient)

    buttonBackground1, _ := imageCache.GetImage("resource.lbx", 13, 0)
    left, _ := imageCache.GetImage("resource.lbx", 5, 0)
    top, _ := imageCache.GetImage("resource.lbx", 7, 0)

    requiredWidth := buttonFont.MeasureTextWidth(selectionTitle, float64(data.ScreenScale)) + 2

    for _, choice := range choices {
        width := buttonFont.MeasureTextWidth(choice.Name, float64(data.ScreenScale)) + 2
        if choice.Hotkey != "" {
            width += buttonFont.MeasureTextWidth(choice.Hotkey, float64(data.ScreenScale)) + 2
        }
        if width > requiredWidth {
            requiredWidth = width
        }
    }

    totalHeight := buttonBackground1.Bounds().Dy() * len(choices)

    elements = append(elements, &UIElement{
        Layer: 1,
        NotLeftClicked: func(this *UIElement){
            getAlpha = ui.MakeFadeOut(fadeSpeed)

            ui.AddDelay(fadeSpeed, func(){
                ui.RemoveElements(elements)
            })
        },
        Draw: func(element *UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            bottom, _ := imageCache.GetImage("resource.lbx", 9, 0)
            options.GeoM.Reset()
            // FIXME: figure out why -3 is needed
            options.GeoM.Translate(float64(cornerX * data.ScreenScale + left.Bounds().Dx()), float64(cornerY * data.ScreenScale + top.Bounds().Dy() + totalHeight - 3))
            bottomSub := bottom.SubImage(image.Rect(0, 0, int(requiredWidth), bottom.Bounds().Dy())).(*ebiten.Image)
            screen.DrawImage(bottomSub, &options)

            bottomLeft, _ := imageCache.GetImage("resource.lbx", 6, 0)
            options.GeoM.Reset()
            options.GeoM.Translate(float64(cornerX * data.ScreenScale), float64(cornerY * data.ScreenScale + totalHeight))
            screen.DrawImage(bottomLeft, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(cornerX * data.ScreenScale), float64(cornerY * data.ScreenScale))
            leftSub := left.SubImage(image.Rect(0, 0, left.Bounds().Dx(), totalHeight)).(*ebiten.Image)
            screen.DrawImage(leftSub, &options)

            topSub := top.SubImage(image.Rect(0, 0, int(requiredWidth), top.Bounds().Dy())).(*ebiten.Image)
            options.GeoM.Reset()
            options.GeoM.Translate(float64(cornerX * data.ScreenScale + left.Bounds().Dx()), float64(cornerY * data.ScreenScale))
            screen.DrawImage(topSub, &options)

            right, _ := imageCache.GetImage("resource.lbx", 8, 0)
            options.GeoM.Reset()
            options.GeoM.Translate(float64(cornerX * data.ScreenScale + left.Bounds().Dx()) + requiredWidth, float64(cornerY * data.ScreenScale))
            rightSub := right.SubImage(image.Rect(0, 0, right.Bounds().Dx(), totalHeight)).(*ebiten.Image)
            screen.DrawImage(rightSub, &options)

            bottomRight, _ := imageCache.GetImage("resource.lbx", 10, 0)
            options.GeoM.Reset()
            options.GeoM.Translate((float64(cornerX * data.ScreenScale + left.Bounds().Dx()) + requiredWidth), float64(cornerY * data.ScreenScale + totalHeight))
            screen.DrawImage(bottomRight, &options)

            topFont.Print(screen, float64(cornerX * data.ScreenScale + left.Bounds().Dx() + 4 * data.ScreenScale), float64(cornerY * data.ScreenScale + 4 * data.ScreenScale), float64(data.ScreenScale), options.ColorScale, selectionTitle)
        },
    })

    x1 := cornerX * data.ScreenScale + left.Bounds().Dx()
    y1 := cornerY * data.ScreenScale + top.Bounds().Dy()

    // FIXME: handle more than 9 choices
    for choiceIndex, choice := range choices {
        images, _ := imageCache.GetImages("resource.lbx", 12 + choiceIndex)
        // the ends are all the same image
        ends, _ := imageCache.GetImages("resource.lbx", 22)

        myX := x1
        myY := y1

        rect := image.Rect(myX, myY, myX + int(requiredWidth), myY + images[0].Bounds().Dy())
        imageIndex := 0
        elements = append(elements, &UIElement{
            Rect: rect,
            Layer: 1,
            Inside: func(this *UIElement, x int, y int){
                imageIndex = 1
            },
            NotInside: func(this *UIElement){
                imageIndex = 0
            },
            PlaySoundLeftClick: true,
            LeftClick: func(this *UIElement){
                getAlpha = ui.MakeFadeOut(fadeSpeed)

                ui.AddDelay(fadeSpeed, func(){
                    ui.RemoveElements(elements)
                    choice.Action()
                })
            },
            Draw: func(element *UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(getAlpha())
                options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))

                use := images[imageIndex].SubImage(image.Rect(0, 0, int(requiredWidth), images[imageIndex].Bounds().Dy())).(*ebiten.Image)
                screen.DrawImage(use, &options)

                options.GeoM.Translate(float64(use.Bounds().Dx()), 0)
                screen.DrawImage(ends[imageIndex], &options)

                y := float64(myY + 2 * data.ScreenScale)

                buttonFont.Print(screen, float64(myX + 2), y, float64(data.ScreenScale), options.ColorScale, choice.Name)
                if choice.Hotkey != "" {
                    buttonFont.PrintRight(screen, float64(myX) + requiredWidth - 2, y, float64(data.ScreenScale), options.ColorScale, choice.Hotkey)
                }
            },
        })

        y1 += images[0].Bounds().Dy()
    }

    return elements
}
