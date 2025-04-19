package ui

import (
    "log"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/help"
    "github.com/hajimehoshi/ebiten/v2"
)

func MakeHelpElement(container UIContainer, cache *lbx.LbxCache, imageCache *util.ImageCache, help help.HelpEntry, helpEntries ...help.HelpEntry) *UIElement {
    return MakeHelpElementWithLayer(container, cache, imageCache, UILayer(1), help, helpEntries...)
}

type HelpFonts struct {
    HelpFont *font.Font
    HelpTitleFont *font.Font
}

func MakeHelpFonts(cache *lbx.LbxCache) HelpFonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Unable to read fonts: %v", err)
        return HelpFonts{}
    }

    return HelpFonts{
        HelpFont: loader(fontslib.HelpFont),
        HelpTitleFont: loader(fontslib.HelpTitleFont),
    }
}

func MakeHelpElementWithLayer(container UIContainer, cache *lbx.LbxCache, imageCache *util.ImageCache, layer UILayer, help help.HelpEntry, helpEntries ...help.HelpEntry) *UIElement {

    helpTop, err := imageCache.GetImage("help.lbx", 0, 0)
    if err != nil {
        return nil
    }

    helpFonts := MakeHelpFonts(cache)

    const fadeSpeed = 7

    getAlpha := container.MakeFadeIn(fadeSpeed)

    infoX := 55
    // infoY := 30
    infoWidth := helpTop.Bounds().Dx()
    // infoHeight := screen.HelpTop.Bounds().Dy()
    infoLeftMargin := 18
    infoTopMargin := 26
    infoBodyMargin := 3
    maxInfoWidth := infoWidth - infoLeftMargin - infoBodyMargin - 14

    // fmt.Printf("Help text: %v\n", []byte(help.Text))

    wrapped := helpFonts.HelpFont.CreateWrappedText(float64(maxInfoWidth), 1, help.Text)

    helpTextY := infoTopMargin
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
        titleYAdjust = extraImage.Bounds().Dy() / 2 - helpFonts.HelpTitleFont.Height() / 2

        if extraImage.Bounds().Dy() > helpFonts.HelpTitleFont.Height() {
            helpTextY += extraImage.Bounds().Dy() + 1
        } else {
            helpTextY += (helpFonts.HelpTitleFont.Height() + 1)
        }
    } else {
        helpTextY += (helpFonts.HelpTitleFont.Height() + 1)
    }

    bottom := float64(helpTextY) + wrapped.TotalHeight

    var moreHelp []font.WrappedText

    // add in more help entries
    for _, entry := range helpEntries {
        bottom += 2
        bottom += float64(helpFonts.HelpTitleFont.Height()) + 1
        moreWrapped := helpFonts.HelpFont.CreateWrappedText(float64(maxInfoWidth), 1, entry.Text)
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
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(window, topImage, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(infoX), float64(bottom) + infoY)
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(window, helpBottom, &options)

            // for debugging
            // vector.StrokeRect(window, float32(infoX), float32(infoY), float32(infoWidth), float32(infoHeight), 1, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, true)
            // vector.StrokeRect(window, float32(infoX + infoLeftMargin), float32(infoY + infoTopMargin), float32(maxInfoWidth), float32(screen.HelpTitleFont.Height() + 20 + 1), 1, color.RGBA{R: 0, G: 0xff, B: 0, A: 0xff}, false)

            titleX := (infoX + infoLeftMargin)

            if extraImage != nil {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(titleX), infoY + float64(infoTopMargin))
                options.ColorScale.ScaleAlpha(getAlpha())
                scale.DrawScaled(window, extraImage, &options)
                titleX += extraImage.Bounds().Dx() + 5
            }

            helpFonts.HelpTitleFont.PrintOptions(window, float64(titleX), infoY + float64(infoTopMargin + titleYAdjust), font.FontOptions{Options: &options, Scale: scale.ScaleAmount}, help.Headline)
            helpFonts.HelpFont.RenderWrapped(window, float64(infoX + infoLeftMargin + infoBodyMargin), float64(helpTextY) + infoY, wrapped, font.FontOptions{Options: &options, Scale: scale.ScaleAmount})

            yPos := float64(helpTextY) + infoY + wrapped.TotalHeight + 2
            for i, moreWrapped := range moreHelp {
                helpFonts.HelpTitleFont.PrintOptions(window, float64(titleX), yPos, font.FontOptions{Options: &options, Scale: scale.ScaleAmount}, helpEntries[i].Headline)
                helpFonts.HelpFont.RenderWrapped(window, float64(infoX + infoLeftMargin + infoBodyMargin), yPos + float64(helpFonts.HelpTitleFont.Height()) + 1, moreWrapped, font.FontOptions{Options: &options, Scale: scale.ScaleAmount})
                yPos += float64(helpFonts.HelpTitleFont.Height()) + 1 + float64(moreWrapped.TotalHeight) + 2
            }

        },
        LeftClick: func(infoThis *UIElement){
            getAlpha = container.MakeFadeOut(fadeSpeed)
            container.AddDelay(fadeSpeed, func(){
                container.RemoveElement(infoThis)
            })
        },
        Layer: layer,
    }

    return infoElement
}

type UIFonts struct {
    Yellow *font.Font
}

func MakeUIFonts(cache *lbx.LbxCache) UIFonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Unable to read fonts: %v", err)
        return UIFonts{}
    }

    return UIFonts{
        Yellow: loader(fontslib.LightFont),
    }
}

func MakeErrorElement(ui UIContainer, cache *lbx.LbxCache, imageCache *util.ImageCache, message string, clicked func()) *UIElement {
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

    fonts := MakeUIFonts(cache)

    maxWidth := errorTop.Bounds().Dx() - errorMargin * 2

    wrapped := fonts.Yellow.CreateWrappedText(float64(maxWidth), 1, message)

    bottom := float64(errorY + errorTopMargin) + wrapped.TotalHeight

    topDraw := errorTop.SubImage(image.Rect(0, 0, errorTop.Bounds().Dx(), int(bottom) - errorY)).(*ebiten.Image)

    element := &UIElement{
        Rect: image.Rect(0, 0, data.ScreenWidth, data.ScreenHeight),
        Layer: 1,
        LeftClick: func(this *UIElement){
            ui.RemoveElement(this)
            clicked()
        },
        Draw: func(this *UIElement, window *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(errorX), float64(errorY))
            scale.DrawScaled(window, topDraw, &options)

            fontOptions := font.FontOptions{Justify: font.FontJustifyCenter, Scale: scale.ScaleAmount, DropShadow: true}
            fonts.Yellow.RenderWrapped(window, float64((errorX + errorMargin) + maxWidth / 2), float64(errorY + errorTopMargin), wrapped, fontOptions)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(errorX), float64(bottom))
            scale.DrawScaled(window, errorBottom, &options)
        },
    }

    return element
}

func MakeConfirmDialog(container UIContainer, cache *lbx.LbxCache, imageCache *util.ImageCache, message string, center bool, confirm func(), cancel func()) []*UIElement {
    return MakeConfirmDialogWithLayer(container, cache, imageCache, 1, message, center, confirm, cancel)
}

func MakeConfirmDialogWithLayer(container UIContainer, cache *lbx.LbxCache, imageCache *util.ImageCache, layer UILayer, message string, center bool, confirm func(), cancel func()) []*UIElement {
    // a button that says 'Yes'
    yesButtons, _ := imageCache.GetImages("resource.lbx", 3)

    // a button that says 'No'
    noButtons, _ := imageCache.GetImages("resource.lbx", 4)
    return MakeConfirmDialogWithLayerFull(container, cache, imageCache, layer, message, center, confirm, cancel, yesButtons, noButtons)
}

func MakeConfirmDialogWithLayerFull(container UIContainer, cache *lbx.LbxCache, imageCache *util.ImageCache, layer UILayer, message string, center bool, confirm func(), cancel func(), yesButtons []*ebiten.Image, noButtons []*ebiten.Image) []*UIElement {
    confirmX := 67
    confirmY := 68

    confirmMargin := 15
    confirmTopMargin := 10

    const fadeSpeed = 7

    getAlpha := container.MakeFadeIn(fadeSpeed)

    confirmTop, err := imageCache.GetImage("resource.lbx", 0, 0)
    if err != nil {
        return nil
    }

    confirmBottom, err := imageCache.GetImage("resource.lbx", 1, 0)
    if err != nil {
        return nil
    }

    fonts := MakeUIFonts(cache)

    maxWidth := confirmTop.Bounds().Dx() - confirmMargin * 2

    wrapped := fonts.Yellow.CreateWrappedText(float64(maxWidth), 1, message)

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
            scale.DrawScaled(window, topDraw, &options)

            if center {
                fonts.Yellow.RenderWrapped(window, float64(confirmX + confirmMargin + maxWidth / 2), float64(confirmY + confirmTopMargin), wrapped, font.FontOptions{Justify: font.FontJustifyCenter, Scale: scale.ScaleAmount, Options: &options})
            } else {
                fonts.Yellow.RenderWrapped(window, float64(confirmX + confirmMargin), float64(confirmY + confirmTopMargin), wrapped, font.FontOptions{Scale: scale.ScaleAmount, Options: &options})
            }

            options.GeoM.Reset()
            options.GeoM.Translate(float64(confirmX), float64(bottom))
            scale.DrawScaled(window, confirmBottom, &options)
        },
    })

    // add yes/no buttons
    if err == nil {
        yesX := confirmX + 101
        yesY := bottom + float64(5)

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

                getAlpha = container.MakeFadeOut(fadeSpeed)
                container.AddDelay(fadeSpeed, func(){
                    container.RemoveElements(elements)
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
                scale.DrawScaled(window, yesButtons[index], &options)
            },
        })
    }

    if err == nil {
        noX := confirmX + 18
        noY := bottom + float64(5)

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

                getAlpha = container.MakeFadeOut(fadeSpeed)
                container.AddDelay(fadeSpeed, func(){
                    container.RemoveElements(elements)
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
                scale.DrawScaled(window, noButtons[index], &options)
            },
        })
    }

    return elements
}

func MakeLairConfirmDialog(ui UIContainer, cache *lbx.LbxCache, imageCache *util.ImageCache, lairPicture *util.Animation, message string, confirm func(), cancel func()) []*UIElement {
    return MakeLairConfirmDialogWithLayer(ui, cache, imageCache, lairPicture, 1, message, confirm, cancel)
}

func MakeLairConfirmDialogWithLayer(ui UIContainer, cache *lbx.LbxCache, imageCache *util.ImageCache, lairPicture *util.Animation, layer UILayer, message string, confirm func(), cancel func()) []*UIElement {
    confirmX := 67
    confirmY := 40

    confirmMargin := 55
    confirmTopMargin := 10

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

    fonts := MakeUIFonts(cache)

    maxWidth := confirmTop.Bounds().Dx() - confirmMargin - 5

    wrapped := fonts.Yellow.CreateWrappedText(float64(maxWidth), 1, message)

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
            scale.DrawScaled(window, topDraw, &options)

            options.GeoM.Translate(float64(7), float64(7))
            scale.DrawScaled(window, lairPicture.Frame(), &options)

            fonts.Yellow.RenderWrapped(window, float64(confirmX + confirmMargin + maxWidth / 2), float64(confirmY + confirmTopMargin), wrapped, font.FontOptions{Justify: font.FontJustifyCenter, Scale: scale.ScaleAmount, Options: &options})

            options.GeoM.Reset()
            options.GeoM.Translate(float64(confirmX - 1), float64(bottom))
            scale.DrawScaled(window, confirmBottom, &options)
        },
    })

    // add yes/no buttons
    yesButtons, err := imageCache.GetImages("resource.lbx", 3)
    if err == nil {
        yesX := confirmX + 101
        yesY := bottom + float64(5)

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
                scale.DrawScaled(window, yesButtons[index], &options)
            },
        })
    }

    noButtons, err := imageCache.GetImages("resource.lbx", 4)
    if err == nil {
        noX := confirmX + 18
        noY := bottom + float64(5)

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
                scale.DrawScaled(window, noButtons[index], &options)
            },
        })
    }

    return elements
}

func MakeLairShowDialogWithLayer(ui UIContainer, cache *lbx.LbxCache, imageCache *util.ImageCache, lairPicture *util.Animation, layer UILayer, message string, dismiss func()) []*UIElement {
    confirmX := 67
    confirmY := 40

    confirmMargin := 55
    confirmTopMargin := 10

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

    fonts := MakeUIFonts(cache)

    maxWidth := confirmTop.Bounds().Dx() - confirmMargin - 5

    wrapped := fonts.Yellow.CreateWrappedText(float64(maxWidth), 1, message)

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
            scale.DrawScaled(window, topDraw, &options)

            options.GeoM.Translate(float64(7), float64(7))
            scale.DrawScaled(window, lairPicture.Frame(), &options)

            fonts.Yellow.RenderWrapped(window, float64(confirmX + confirmMargin + maxWidth / 2), float64(confirmY + confirmTopMargin), wrapped, font.FontOptions{Justify: font.FontJustifyCenter, Scale: scale.ScaleAmount, Options: &options})

            options.GeoM.Reset()
            options.GeoM.Translate(float64(confirmX), float64(bottom))
            scale.DrawScaled(window, confirmBottom, &options)
        },
    })

    return elements
}

type Selection struct {
    Name string
    Action func()
    Hotkey string
}

type SelectionFonts struct {
    Black *font.Font
    Title *font.Font
}

func MakeSelectionFonts(cache *lbx.LbxCache) SelectionFonts {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return SelectionFonts{}
    }

    font4, err := font.ReadFont(fontLbx, 0, 4)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return SelectionFonts{}
    }

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

    return SelectionFonts{
        Black: buttonFont,
        Title: topFont,
    }
}

func MakeSelectionUI(ui UIContainer, lbxCache *lbx.LbxCache, imageCache *util.ImageCache, cornerX int, cornerY int, selectionTitle string, choices []Selection, canCancel bool) []*UIElement {
    var elements []*UIElement

    fadeSpeed := uint64(6)

    getAlpha := ui.MakeFadeIn(fadeSpeed)

    fonts := MakeSelectionFonts(lbxCache)

    buttonBackground1, _ := imageCache.GetImage("resource.lbx", 13, 0)
    left, _ := imageCache.GetImage("resource.lbx", 5, 0)
    top, _ := imageCache.GetImage("resource.lbx", 7, 0)

    requiredWidth := fonts.Black.MeasureTextWidth(selectionTitle, 1) + 2

    for _, choice := range choices {
        width := fonts.Black.MeasureTextWidth(choice.Name, 1) + 2
        if choice.Hotkey != "" {
            width += fonts.Black.MeasureTextWidth(choice.Hotkey, 1) + 2
        }
        if width > requiredWidth {
            requiredWidth = width
        }
    }

    totalHeight := buttonBackground1.Bounds().Dy() * len(choices)

    elements = append(elements, &UIElement{
        Layer: 1,
        NotLeftClicked: func(this *UIElement){
            // if the user cannot cancel then they have to select one of the options
            if canCancel {
                getAlpha = ui.MakeFadeOut(fadeSpeed)

                ui.AddDelay(fadeSpeed, func(){
                    ui.RemoveElements(elements)
                })
            }
        },
        Draw: func(element *UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            bottom, _ := imageCache.GetImage("resource.lbx", 9, 0)
            options.GeoM.Reset()
            // FIXME: figure out why -3 is needed
            options.GeoM.Translate(float64(cornerX + left.Bounds().Dx()), float64(cornerY + top.Bounds().Dy() + totalHeight - 3))
            bottomSub := bottom.SubImage(image.Rect(0, 0, int(requiredWidth), bottom.Bounds().Dy())).(*ebiten.Image)
            scale.DrawScaled(screen, bottomSub, &options)

            bottomLeft, _ := imageCache.GetImage("resource.lbx", 6, 0)
            options.GeoM.Reset()
            options.GeoM.Translate(float64(cornerX), float64(cornerY + totalHeight))
            scale.DrawScaled(screen, bottomLeft, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(cornerX), float64(cornerY))
            leftSub := left.SubImage(image.Rect(0, 0, left.Bounds().Dx(), totalHeight)).(*ebiten.Image)
            scale.DrawScaled(screen, leftSub, &options)

            topSub := top.SubImage(image.Rect(0, 0, int(requiredWidth), top.Bounds().Dy())).(*ebiten.Image)
            options.GeoM.Reset()
            options.GeoM.Translate(float64(cornerX + left.Bounds().Dx()), float64(cornerY))
            scale.DrawScaled(screen, topSub, &options)

            right, _ := imageCache.GetImage("resource.lbx", 8, 0)
            options.GeoM.Reset()
            options.GeoM.Translate(float64(cornerX + left.Bounds().Dx()) + requiredWidth, float64(cornerY))
            rightSub := right.SubImage(image.Rect(0, 0, right.Bounds().Dx(), totalHeight)).(*ebiten.Image)
            scale.DrawScaled(screen, rightSub, &options)

            bottomRight, _ := imageCache.GetImage("resource.lbx", 10, 0)
            options.GeoM.Reset()
            options.GeoM.Translate((float64(cornerX + left.Bounds().Dx()) + requiredWidth), float64(cornerY + totalHeight))
            scale.DrawScaled(screen, bottomRight, &options)

            fonts.Title.PrintOptions(screen, float64(cornerX + left.Bounds().Dx() + 4), float64(cornerY + 4), font.FontOptions{Options: &options, Scale: scale.ScaleAmount}, selectionTitle)
        },
    })

    x1 := cornerX + left.Bounds().Dx()
    y1 := cornerY + top.Bounds().Dy()

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
                scale.DrawScaled(screen, use, &options)

                options.GeoM.Translate(float64(use.Bounds().Dx()), 0)
                scale.DrawScaled(screen, ends[imageIndex], &options)

                y := float64(myY + 2)

                fonts.Black.PrintOptions(screen, float64(myX + 2), y, font.FontOptions{Options: &options, Scale: scale.ScaleAmount}, choice.Name)
                if choice.Hotkey != "" {
                    fonts.Black.PrintOptions(screen, float64(myX) + requiredWidth - 2, y, font.FontOptions{Options: &options, Scale: scale.ScaleAmount, Justify: font.FontJustifyRight}, choice.Hotkey)
                }
            },
        })

        y1 += images[0].Bounds().Dy()
    }

    return elements
}
