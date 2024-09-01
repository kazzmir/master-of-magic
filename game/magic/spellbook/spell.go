package spellbook

import (
    "image"
    "image/color"
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/hajimehoshi/ebiten/v2"
)

func MakeSpellBookUI(ui *uilib.UI, cache *lbx.LbxCache) []*uilib.UIElement {
    var elements []*uilib.UIElement

    imageCache := util.MakeImageCache(cache)

    fadeSpeed := uint64(7)

    spells, err := ReadSpellsFromCache(cache)
    if err != nil {
        log.Printf("Unable to read spells: %v", err)
        return nil
    }

    spellDescriptions, err := ReadSpellDescriptionsFromCache(cache)
    if err != nil {
        log.Printf("Unable to read spell descriptions: %v", err)
        return nil
    }

    getAlpha := ui.MakeFadeIn(fadeSpeed)

    bookFlip, _ := imageCache.GetImages("book.lbx", 1)
    bookFlipIndex := uint64(0)
    bookFlipReverse := false

    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts: %v", err)
        return nil
    }

    red := color.RGBA{R: 0x5a, G: 0, B: 0, A: 0xff}
    redPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        red, red, red,
        red, red, red,
        red, red, red,
        red, red, red,
        red, red, red,
        red, red, red,
    }

    titleFont := font.MakeOptimizedFontWithPalette(fonts[5], redPalette)

    grey := color.RGBA{R: 35, G: 35, B: 35, A: 0xff}
    textPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        util.PremultiplyAlpha(color.RGBA{R: 35, G: 35, B: 35, A: 64}),
        grey, grey, grey,
        grey, grey, grey,
        grey, grey, grey,
        grey, grey, grey,
    }

    greyLight := util.PremultiplyAlpha(color.RGBA{R: 35, G: 35, B: 35, A: 164})
    textPaletteLighter := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        util.PremultiplyAlpha(color.RGBA{R: 35, G: 35, B: 35, A: 64}),
        greyLight, greyLight, greyLight,
        greyLight, greyLight, greyLight,
    }

    spellTitleNormalFont := font.MakeOptimizedFontWithPalette(fonts[4], textPalette)
    spellTextNormalFont := font.MakeOptimizedFontWithPalette(fonts[0], textPaletteLighter)

    spellTitleAlienFont := font.MakeOptimizedFontWithPalette(fonts[7], textPalette)
    spellTextAlienFont := font.MakeOptimizedFontWithPalette(fonts[6], textPaletteLighter)

    // mystery font title: fonts[7]
    // mystery font normal: fonts[6]

    showSection := SectionUnitSpell

    wrapWidth := float64(130)

    spellDescriptionNormalCache := make(map[int]font.WrappedText)

    getSpellDescriptionNormalText := func(index int) font.WrappedText {
        text, ok := spellDescriptionNormalCache[index]
        if ok {
            return text
        }

        wrapped := spellTextNormalFont.CreateWrappedText(wrapWidth, 1, spellDescriptions[index])
        spellDescriptionNormalCache[index] = wrapped
        return wrapped
    }

    spellDescriptionAlienCache := make(map[int]font.WrappedText)

    getSpellDescriptionAlienText := func(index int) font.WrappedText {
        text, ok := spellDescriptionAlienCache[index]
        if ok {
            return text
        }

        wrapped := spellTextAlienFont.CreateWrappedText(wrapWidth, 1, spellDescriptions[index])
        spellDescriptionAlienCache[index] = wrapped
        return wrapped
    }

    knownSpell := func(spell Spell) bool {
        return true
        // return spell.Index <= 2
    }

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        NotLeftClicked: func(this *uilib.UIElement){
            getAlpha = ui.MakeFadeOut(fadeSpeed)
            ui.AddDelay(fadeSpeed, func(){
                ui.RemoveElements(elements)
            })
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("scroll.lbx", 6, 0)

            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(background, &options)

            titleFont.PrintCenter(screen, 90, 11, 1, options.ColorScale, showSection.Name())

            sectionSpells := spells.GetSpellsBySection(showSection)

            x := float64(25)
            y := float64(35)
            for i, spell := range sectionSpells.Spells {
                if i >= 4 {
                    break
                }

                if knownSpell(spell) {
                    spellTitleNormalFont.Print(screen, x, y, 1, options.ColorScale, spell.Name)
                    wrapped := getSpellDescriptionNormalText(spell.Index)
                    spellTextNormalFont.RenderWrapped(screen, x, y + 10, wrapped, options.ColorScale, false)
                } else {
                    spellTitleAlienFont.Print(screen, x, y, 1, options.ColorScale, spell.Name)
                    wrapped := getSpellDescriptionAlienText(spell.Index)
                    spellTextAlienFont.RenderWrapped(screen, x, y + 10, wrapped, options.ColorScale, false)
                }

                y += 35
            }

            animationIndex := ui.Counter
            if bookFlipIndex > 0 && (animationIndex - bookFlipIndex) / 6 < uint64(len(bookFlip)) {
                index := (animationIndex - bookFlipIndex) / 6
                if bookFlipReverse {
                    index = uint64(len(bookFlip)) - 1 - index
                }
                options.GeoM.Translate(0, 0)
                screen.DrawImage(bookFlip[index], &options)
            }

        },
    })

    // left page turn
    leftTurn, _ := imageCache.GetImage("scroll.lbx", 7, 0)
    leftRect := image.Rect(15, 9, 15 + leftTurn.Bounds().Dx(), 9 + leftTurn.Bounds().Dy())
    elements = append(elements, &uilib.UIElement{
        Rect: leftRect,
        Layer: 1,
        LeftClick: func(this *uilib.UIElement){
            bookFlipIndex = ui.Counter
            bookFlipReverse = true

            showSection = showSection.PreviousSection()
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(leftRect.Min.X), float64(leftRect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(leftTurn, &options)
        },
    })

    // right page turn
    rightTurn, _ := imageCache.GetImage("scroll.lbx", 8, 0)
    rightRect := image.Rect(289, 9, 295 + rightTurn.Bounds().Dx(), 5 + rightTurn.Bounds().Dy())
    elements = append(elements, &uilib.UIElement{
        Rect: rightRect,
        Layer: 1,
        LeftClick: func(this *uilib.UIElement){
            bookFlipIndex = ui.Counter
            bookFlipReverse = false
            showSection = showSection.NextSection()
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(rightRect.Min.X), float64(rightRect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(rightTurn, &options)
        },
    })

    return elements
}
