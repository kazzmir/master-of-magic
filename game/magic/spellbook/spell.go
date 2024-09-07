package spellbook

import (
    "fmt"
    "image"
    "image/color"
    "math"
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

func computeHalfPages(spells Spells, max int) []Spells {
    var halfPages []Spells

    sections := []Section{SectionSummoning, SectionSpecial, SectionCitySpell, SectionEnchantment, SectionUnitSpell, SectionCombatSpell}

    for _, section := range sections {
        sectionSpells := spells.GetSpellsBySection(section)
        numSpells := len(sectionSpells.Spells)

        for i := 0; i < int(math.Ceil(float64(numSpells) / float64(max))); i++ {
            var pageSpells Spells

            for j := 0; j < max; j++ {
                index := i * max + j
                if index < numSpells {
                    pageSpells.AddSpell(sectionSpells.Spells[index])
                }
            }

            if len(pageSpells.Spells) > 0 {
                halfPages = append(halfPages, pageSpells)
            }
        }
    }

    return halfPages
}

// flipping the page to the left
func LeftSideDistortions1(page *ebiten.Image) util.Distortion {
    return util.Distortion{
        Top: image.Pt(page.Bounds().Dx()/2 + 20, 5),
        Bottom: image.Pt(page.Bounds().Dx()/2 + 20, page.Bounds().Dy() - 12),
        Segments: []util.Segment{
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 40, 0),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 40, page.Bounds().Dy() - 25),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 60, -10),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 60, page.Bounds().Dy() - 33),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 80, -10),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 80, page.Bounds().Dy() - 30),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 100, -0),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 100, page.Bounds().Dy() - 22),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 130, -10),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 130, page.Bounds().Dy() - 12),
            },
        },
    }
}

func LeftSideDistortions2(page *ebiten.Image) util.Distortion {
    return util.Distortion{
        Top: image.Pt(page.Bounds().Dx()/2 + 20, 5),
        Bottom: image.Pt(page.Bounds().Dx()/2 + 20, page.Bounds().Dy() - 15),
        Segments: []util.Segment{
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 40, 0),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 40, page.Bounds().Dy() - 28),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 58, -13),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 58, page.Bounds().Dy() - 35),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 73, -20),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 73, page.Bounds().Dy() - 35),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 90, -0),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 90, page.Bounds().Dy() - 22),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 120, -10),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 120, page.Bounds().Dy() - 12),
            },
        },
    }
}

func RightSideDistortions2(page *ebiten.Image) util.Distortion {
    offset := 30
    return util.Distortion{
        Top: image.Pt(page.Bounds().Dx()/2 - 130 + offset, 5),
        Bottom: image.Pt(page.Bounds().Dx()/2 - 130 + offset, page.Bounds().Dy() - 0),

        Segments: []util.Segment{
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 100 + offset, -0),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 100 + offset, page.Bounds().Dy() - 22),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 80 + offset, -10),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 80 + offset, page.Bounds().Dy() - 30),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 60 + offset, -10),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 60 + offset, page.Bounds().Dy() - 33),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 40 + offset, 0),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 40 + offset, page.Bounds().Dy() - 25),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 20 + offset, 5),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 20 + offset, page.Bounds().Dy() - 12),
            },
        },
    }
}

func RightSideDistortions1(page *ebiten.Image) util.Distortion {
    offset := 50
    return util.Distortion{
        Top: image.Pt(page.Bounds().Dx()/2 - 110 + offset, -10),
        Bottom: image.Pt(page.Bounds().Dx()/2 - 110 + offset, page.Bounds().Dy() - 12),
        Segments: []util.Segment{
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 90 + offset, -0),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 90 + offset, page.Bounds().Dy() - 22),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 73 + offset, -20),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 73 + offset, page.Bounds().Dy() - 35),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 58 + offset, -13),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 58 + offset, page.Bounds().Dy() - 35),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 40 + offset, 0),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 40 + offset, page.Bounds().Dy() - 28),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 20 + offset, 5),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 20 + offset, page.Bounds().Dy() - 15),
            },
        },
    }
}

// flipping the page to the right
func RightSideFlipRightDistortions1(page *ebiten.Image) util.Distortion {
    return util.Distortion{
        Top: image.Pt(page.Bounds().Dx()/2 + 20, 5),
        Bottom: image.Pt(page.Bounds().Dx()/2 + 20, page.Bounds().Dy() - 12),
        Segments: []util.Segment{
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 40, 0),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 40, page.Bounds().Dy() - 25),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 60, -10),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 60, page.Bounds().Dy() - 33),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 80, -10),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 80, page.Bounds().Dy() - 30),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 100, -0),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 100, page.Bounds().Dy() - 22),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 130, -10),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 130, page.Bounds().Dy() - 12),
            },
        },
    }
}

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

    // use index 0 for a smaller book (like when casting?)
    bookFlip, _ := imageCache.GetImages("book.lbx", 1)
    bookFlipIndex := uint64(0)
    bookFlipReverse := false

    bookFlipSpeed := uint64(7)

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

    // showSection := SectionSpecial
    // page N refers to both left and right sides of the book
    // given 5 summoning spells and 2 unit spells
    // page 0 would be left: summoning spells 1-4, right: summoning spell 5
    // page 1 would be left: unit spells 1-2, right: empty

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

    // compute half pages
    halfPages := computeHalfPages(spells, 4)

    // for debugging
    /*
    for i, halfPage := range halfPages {
        log.Printf("Half page %d: length=%v %+v", i, len(halfPage.Spells), halfPage)
    }
    */

    hasNextPage := func(page int) bool {
        return page < len(halfPages) - 1
    }

    hasPreviousPage := func(page int) bool {
        return page > 0
    }

    // create images of each page
    halfPageCache := make(map[int]*ebiten.Image)

    // lazily construct the page graphics, which consists of the section title and 4 spell descriptions
    getHalfPageImage := func(halfPage int) *ebiten.Image {
        image, ok := halfPageCache[halfPage]
        if ok {
            return image
        }

        var pageSpells Spells
        if halfPage < len(halfPages) {
            pageSpells = halfPages[halfPage]
        }

        pageImage := ebiten.NewImage(155, 170)
        pageImage.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 0})

        if len(pageSpells.Spells) > 0 {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(0, 0)

            section := pageSpells.Spells[0].Section
            titleFont.PrintCenter(pageImage, 90, 11, 1, options.ColorScale, section.Name())

            x := float64(25)
            y := float64(35)
            for i, spell := range pageSpells.Spells {
                if i >= 4 {
                    break
                }

                if knownSpell(spell) {
                    spellTitleNormalFont.Print(pageImage, x, y, 1, options.ColorScale, spell.Name)
                    wrapped := getSpellDescriptionNormalText(spell.Index)
                    spellTextNormalFont.RenderWrapped(pageImage, x, y + 10, wrapped, options.ColorScale, false)
                } else {
                    spellTitleAlienFont.Print(pageImage, x, y, 1, options.ColorScale, spell.Name)
                    wrapped := getSpellDescriptionAlienText(spell.Index)
                    spellTextAlienFont.RenderWrapped(pageImage, x, y + 10, wrapped, options.ColorScale, false)
                }

                y += 35
            }
        }

        halfPageCache[halfPage] = pageImage
        return pageImage
    }

    // FIXME: this page could be passed in, so that it is stored for a while
    // page := 0
    showLeftPage := 0
    showRightPage := 1

    flipLeftSide := 0
    flipRightSide := 1

    flipping := false

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

            if showLeftPage >= 0 {
                leftPageImage := getHalfPageImage(showLeftPage)
                screen.DrawImage(leftPageImage, &options)
            }

            if showRightPage < len(halfPages) {
                rightPageImage := getHalfPageImage(showRightPage)
                rightOptions := options
                rightOptions.GeoM.Translate(148, 0)
                screen.DrawImage(rightPageImage, &rightOptions)
            }

            animationIndex := ui.Counter
            if bookFlipIndex > 0 && (animationIndex - bookFlipIndex) / bookFlipSpeed < uint64(len(bookFlip)) {
                index := (animationIndex - bookFlipIndex) / bookFlipSpeed
                if bookFlipReverse {
                    index = uint64(len(bookFlip)) - 1 - index
                }
                options.GeoM.Translate(0, 0)
                screen.DrawImage(bookFlip[index], &options)

                if index == 0 {
                    if flipLeftSide >= 0 {
                        leftSide := getHalfPageImage(flipLeftSide)
                        util.DrawDistortion(screen, bookFlip[index], leftSide, LeftSideDistortions1(bookFlip[index]), options)
                    }
                } else if index == 1 {
                    if flipLeftSide >= 0 {
                        leftSide := getHalfPageImage(flipLeftSide)
                        util.DrawDistortion(screen, bookFlip[index], leftSide, LeftSideDistortions2(bookFlip[index]), options)
                    }
                } else if index == 2 {
                    if flipRightSide < len(halfPages) {
                        rightSide := getHalfPageImage(flipRightSide)
                        util.DrawDistortion(screen, bookFlip[index], rightSide, RightSideDistortions1(bookFlip[index]), options)
                    }
                } else if index == 3 {
                    if flipRightSide < len(halfPages) {
                        rightSide := getHalfPageImage(flipRightSide)
                        util.DrawDistortion(screen, bookFlip[index], rightSide, RightSideDistortions2(bookFlip[index]), options)
                    }
                }
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
            if !flipping && hasPreviousPage(showLeftPage){
                bookFlipIndex = ui.Counter
                bookFlipReverse = true

                flipLeftSide = showLeftPage - 1
                flipRightSide = showLeftPage
                showLeftPage -= 2
                flipping = true

                ui.AddDelay(bookFlipSpeed * uint64(len(bookFlip)) - 1, func(){
                    showRightPage -= 2
                    flipping = false
                })
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            if hasPreviousPage(showLeftPage){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(leftRect.Min.X), float64(leftRect.Min.Y))
                options.ColorScale.ScaleAlpha(getAlpha())
                screen.DrawImage(leftTurn, &options)
            }
        },
    })

    // right page turn
    rightTurn, _ := imageCache.GetImage("scroll.lbx", 8, 0)
    rightRect := image.Rect(289, 9, 295 + rightTurn.Bounds().Dx(), 5 + rightTurn.Bounds().Dy())
    elements = append(elements, &uilib.UIElement{
        Rect: rightRect,
        Layer: 1,
        LeftClick: func(this *uilib.UIElement){
            if !flipping && hasNextPage(showRightPage){
                bookFlipIndex = ui.Counter
                bookFlipReverse = false

                flipLeftSide = showRightPage
                flipRightSide = showRightPage + 1
                showRightPage += 2

                flipping = true

                ui.AddDelay(bookFlipSpeed * uint64(len(bookFlip)) - 1, func(){
                    showLeftPage += 2
                    flipping = false
                })
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            if hasNextPage(showRightPage){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(rightRect.Min.X), float64(rightRect.Min.Y))
                options.ColorScale.ScaleAlpha(getAlpha())
                screen.DrawImage(rightTurn, &options)
            }
        },
    })

    return elements
}

func maxRange[T any](slice []T, max int) []T {
    if len(slice) <= max {
        return slice
    }

    return slice[:max]
}

// FIXME: take in the wizard/player that is casting the spell
// somehow return the spell chosen
func MakeSpellBookCastUI(ui *uilib.UI, cache *lbx.LbxCache, spells Spells, castingSkill int) []*uilib.UIElement {
    var elements []*uilib.UIElement

    imageCache := util.MakeImageCache(cache)

    getAlpha := ui.MakeFadeIn(7)

    black := color.RGBA{R: 0, G: 0, B: 0, A: 0xff}

    paletteBlack := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        black, black, black,
        black, black, black,
    }

    getMagicIcon := func(spell Spell) *ebiten.Image {
        index := -1
        switch spell.Magic {
            case data.LifeMagic: index = 7
            case data.SorceryMagic: index = 5
            case data.NatureMagic: index = 4
            case data.DeathMagic: index = 8
            case data.ChaosMagic: index = 6
            case data.ArcaneMagic: index = 9
        }

        if index == -1 {
            return nil
        }

        img, _ := imageCache.GetImage("spells.lbx", index, 0)
        return img
    }

    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Could not read fonts: %v", err)
        return nil
    }

    fonts, _ := font.ReadFonts(fontLbx, 0)

    infoFont := font.MakeOptimizedFontWithPalette(fonts[1], paletteBlack)

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

    titleFont := font.MakeOptimizedFontWithPalette(fonts[4], redPalette)

    pageCache := make(map[int]*ebiten.Image)

    spellPages := computeHalfPages(spells, 6)

    var chosenSpell Spell

    // lazily construct the page graphics, which consists of the section title and 4 spell descriptions
    getPageImage := func(page int) *ebiten.Image {
        cached, ok := pageCache[page]
        if ok {
            return cached
        }

        out := ebiten.NewImage(120, 154)
        out.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 0})

        var options ebiten.DrawImageOptions
        if page < len(spellPages) {
            titleFont.PrintCenter(out, 60, 1, 1, options.ColorScale, "Summoning")
            gibberish, _ := imageCache.GetImage("spells.lbx", 10, 0)
            gibberishHeight := 18

            options2 := options
            options2.GeoM.Translate(0, 15)
            for _, spell := range spellPages[page].Spells {
                spellX, spellY := options2.GeoM.Apply(0, 0)

                infoFont.Print(out, spellX, spellY, 1, options.ColorScale, spell.Name)
                infoFont.PrintRight(out, spellX + 124, spellY, 1, options.ColorScale, fmt.Sprintf("%v MP", spell.CastCost))
                icon := getMagicIcon(spell)

                nameLength := infoFont.MeasureTextWidth(spell.Name, 1) + 1
                mpLength := infoFont.MeasureTextWidth(fmt.Sprintf("%v MP", spell.CastCost), 1)

                gibberishPart := gibberish.SubImage(image.Rect(0, 0, gibberish.Bounds().Dx(), gibberishHeight)).(*ebiten.Image)

                partIndex := 0
                partHeight := 20

                subLines := 6

                part1 := gibberishPart.SubImage(image.Rect(int(nameLength), partIndex * partHeight, int(nameLength) + gibberishPart.Bounds().Dx() - int(nameLength + mpLength), partIndex * partHeight + subLines)).(*ebiten.Image)

                part1Options := options2
                part1Options.GeoM.Translate(nameLength, 0)
                out.DrawImage(part1, &part1Options)

                iconCount := spell.CastCost / int(math.Max(1, float64(castingSkill)))
                if iconCount < 1 {
                    iconCount = 1
                }

                iconOptions := options2
                iconOptions.GeoM.Translate(0, float64(infoFont.Height())+1)
                part3Options := iconOptions

                icons1 := iconCount
                if icons1 > 20 {
                    icons1 = 20
                    iconCount -= icons1
                    // FIXME: what to do if there is still overflow here?
                    if iconCount > 20 {
                        iconCount = 20
                    }
                } else {
                    iconCount = 0
                }

                for i := 0; i < icons1; i++ {
                    out.DrawImage(icon, &iconOptions)
                    iconOptions.GeoM.Translate(float64(icon.Bounds().Dx()) + 1, 0)
                }

                part2 := gibberishPart.SubImage(image.Rect((icon.Bounds().Dx() + 1) * icons1 + 3, partIndex * partHeight + subLines, gibberish.Bounds().Dx(), partIndex * partHeight + subLines * 2)).(*ebiten.Image)
                part2Options := iconOptions
                part2Options.GeoM.Translate(3, 0)
                out.DrawImage(part2, &part2Options)

                part3Options.GeoM.Translate(0, float64(icon.Bounds().Dy()+1))

                for i := 0; i < iconCount; i++ {
                    out.DrawImage(icon, &part3Options)
                    part3Options.GeoM.Translate(float64(icon.Bounds().Dx()) + 1, 0)
                }

                part3 := gibberishPart.SubImage(image.Rect((icon.Bounds().Dx() + 1) * iconCount, partIndex * partHeight + subLines * 2, gibberish.Bounds().Dx(), partIndex * partHeight + subLines * 3)).(*ebiten.Image)
                out.DrawImage(part3, &part3Options)

                options2.GeoM.Translate(0, 22)
            }
        }

        // vector.StrokeRect(out, 1, 1, float32(out.Bounds().Dx()-1), float32(out.Bounds().Dy()-10), 1, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
        pageCache[page] = out
        return out
    }

    var spellButtons []*uilib.UIElement
    setupSpells := func(page int) {
        ui.RemoveElements(spellButtons)
        spellButtons = nil

        if page < 0 || page >= len(spellPages) {
            return
        }

        leftSpells := spellPages[page].Spells

        for i, spell := range leftSpells {

            x1 := 24
            y1 := 30 + i * 22
            width := 122
            height := 20

            rect := image.Rect(0, 0, width, height).Add(image.Pt(x1, y1))
            spellButtons = append(spellButtons, &uilib.UIElement{
                Rect: rect,
                Layer: 1,
                LeftClick: func(this *uilib.UIElement){
                    log.Printf("Click on spell %v", spell)
                    chosenSpell = spell
                },
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    vector.StrokeRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Max.X - rect.Min.X), float32(rect.Max.Y - rect.Min.Y), 1, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
                },
            })
        }

        if page + 1 < len(spellPages) {
            rightSpells := spellPages[page+1].Spells

            for i, spell := range rightSpells {

                x1 := 159
                y1 := 30 + i * 22
                width := 122
                height := 20

                rect := image.Rect(0, 0, width, height).Add(image.Pt(x1, y1))
                spellButtons = append(spellButtons, &uilib.UIElement{
                    Rect: rect,
                    Layer: 1,
                    LeftClick: func(this *uilib.UIElement){
                        log.Printf("Click on spell %v", spell)
                        chosenSpell = spell
                    },
                    Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                        vector.StrokeRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Max.X - rect.Min.X), float32(rect.Max.Y - rect.Min.Y), 1, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
                    },
                })
            }
        }

        ui.AddElements(spellButtons)
    }

    currentPage := 0

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        NotLeftClicked: func(this *uilib.UIElement){
            getAlpha = ui.MakeFadeOut(7)
            ui.AddDelay(7, func(){
                ui.RemoveElements(elements)
                setupSpells(-1)

                log.Printf("Chose spell %+v", chosenSpell)
            })
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            // FIXME: do the whole page flipping thing with distorted pages
            vector.DrawFilledRect(screen, 0, 0, float32(screen.Bounds().Dx()), float32(screen.Bounds().Dy()), color.RGBA{R: 0, G: 0, B: 0, A: 128}, false)

            background, _ := imageCache.GetImage("spells.lbx", 0, 0)
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            options.GeoM.Translate(10, 10)
            screen.DrawImage(background, &options)

            left := getPageImage(currentPage)
            right := getPageImage(currentPage+1)

            options.GeoM.Translate(15, 5)
            screen.DrawImage(left, &options)

            options.GeoM.Translate(134, 0)
            screen.DrawImage(right, &options)
        },
    })


    // hack to add the spell ui elements after the main element
    ui.AddDelay(0, func(){
        setupSpells(currentPage)
    })

    pageTurnRight, _ := imageCache.GetImage("spells.lbx", 2, 0)
    pageTurnRightRect := image.Rect(0, 0, pageTurnRight.Bounds().Dx(), pageTurnRight.Bounds().Dy()).Add(image.Pt(268, 14))
    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Rect: pageTurnRightRect,
        LeftClick: func(this *uilib.UIElement){
            if currentPage + 2 < len(spellPages) {
                currentPage += 2
                setupSpells(currentPage)
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            if currentPage + 2 < len(spellPages) {
                options.GeoM.Translate(float64(pageTurnRightRect.Min.X), float64(pageTurnRightRect.Min.Y))
                screen.DrawImage(pageTurnRight, &options)
            }
        },
    })

    pageTurnLeft, _ := imageCache.GetImage("spells.lbx", 1, 0)
    pageTurnLeftRect := image.Rect(0, 0, pageTurnLeft.Bounds().Dx(), pageTurnLeft.Bounds().Dy()).Add(image.Pt(23, 14))
    elements = append(elements, &uilib.UIElement{
        Rect: pageTurnLeftRect,
        Layer: 1,
        LeftClick: func(this *uilib.UIElement){
            if currentPage >= 2 {
                currentPage -= 2
                setupSpells(currentPage)
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            if currentPage > 0 {
                options.GeoM.Translate(float64(pageTurnLeftRect.Min.X), float64(pageTurnLeftRect.Min.Y))
                screen.DrawImage(pageTurnLeft, &options)
            }

        },
    })

    return elements
}
