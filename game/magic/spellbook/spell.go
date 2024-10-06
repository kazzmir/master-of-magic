package spellbook

import (
    "fmt"
    "image"
    "image/color"
    "math"
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

type Page struct {
    Title string
    Spells Spells
    // true if this page should render a title even if the spells are empty
    ForceRender bool
    // true if the text for the spell should always use normal font rather than alien
    IsResearch bool
}

func computeHalfPages(spells Spells, max int) []Page {
    var halfPages []Page

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
                halfPages = append(halfPages, Page{
                    Title: section.Name(),
                    Spells: pageSpells,
                })
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

/* two modes:
 * 1. when a new spell is learned, flip to the page where the spell would go and show the sparkle animation,
 *    then flip to the 'research spells' page and let the user pick a new spell
 * 2. show book and let user flip between pages. on the 'research spells' page, show currently
*     researching spell as glowing text
 */
func ShowSpellBook(yield coroutine.YieldFunc, cache *lbx.LbxCache, allSpells Spells, knownSpells Spells, researchSpells Spells, researchingSpell Spell, researchProgress int, researchPoints int, castingSkill int, learnedSpell Spell, pickResearchSpell bool, chosenSpell *Spell, drawFunc *func(screen *ebiten.Image)) {
    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })
        },
    }

    if researchPoints < 1 {
        researchPoints = 1
    }

    var elements []*uilib.UIElement

    imageCache := util.MakeImageCache(cache)

    fadeSpeed := uint64(7)

    /*
    spells, err := ReadSpellsFromCache(cache)
    if err != nil {
        log.Printf("Unable to read spells: %v", err)
        return
    }
    */

    spellDescriptions, err := ReadSpellDescriptionsFromCache(cache)
    if err != nil {
        log.Printf("Unable to read spell descriptions: %v", err)
        return
    }

    getAlpha := ui.MakeFadeIn(fadeSpeed)

    bookFlip, _ := imageCache.GetImages("book.lbx", 1)
    bookFlipIndex := uint64(0)
    bookFlipReverse := false

    bookFlipSpeed := uint64(7)

    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts: %v", err)
        return
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts: %v", err)
        return
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

    white := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
    whitePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        white,
        util.Lighten(white, -10),
        util.Lighten(white, -20),
        util.Lighten(white, -30),
        util.Lighten(white, -40),
        util.Lighten(white, -50),
        white, white, white, white, white, white,
        white, white, white, white, white, white,
    }

    chooseFont := font.MakeOptimizedFontWithPalette(fonts[5], whitePalette)

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
        return knownSpells.Contains(spell)
    }

    // compute half pages
    var halfPages []Page
    if !pickResearchSpell {
        halfPages = computeHalfPages(allSpells, 4)
    }

    researchPage1 := Page{
        Title: "Research",
        Spells: researchSpells.Sub(0, 4),
        ForceRender: true,
        IsResearch: true,
    }

    researchPage2 := Page{
        Title: "Spells",
        Spells: researchSpells.Sub(4, 8),
        ForceRender: true,
        IsResearch: true,
    }

    // insert an empty page so that the research pages are on their own
    if len(halfPages) % 2 != 0 {
        halfPages = append(halfPages, Page{})
    }

    halfPages = append(halfPages, researchPage1, researchPage2)

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

    learnSpellPics, _ := imageCache.GetImages("specfx.lbx", 49)
    learnSpellAnimation := util.MakeAnimation(learnSpellPics, false)

    renderPage := func(page Page, flipping bool, pageImage *ebiten.Image, options ebiten.DrawImageOptions){
        if len(page.Spells.Spells) > 0 || page.ForceRender {
            // var options ebiten.DrawImageOptions
            // options.GeoM.Translate(0, 0)

            // section := pageSpells.Spells[0].Section
            titleX, titleY := options.GeoM.Apply(90, 11)

            titleFont.PrintCenter(pageImage, titleX, titleY, 1, options.ColorScale, page.Title)

            x, topY := options.GeoM.Apply(25, 35)

            for i, spell := range page.Spells.Spells {
                if i >= 4 {
                    break
                }

                y := topY + float64(i * 35)

                spellY := y

                if page.IsResearch || knownSpell(spell) || learnedSpell.Name == spell.Name {
                    scale := options.ColorScale

                    if !flipping && researchingSpell.Name == spell.Name {
                        v := 1.5 + (math.Cos(float64(ui.Counter) / 7) * 64 + 64) / float64(64)
                        scale.SetR(float32(v))
                        scale.SetG(float32(v))
                        scale.SetB(float32(v) * 1.8)
                    }

                    if !flipping && learnedSpell.Name == spell.Name && !learnSpellAnimation.Done() {
                        v := 1.5 + float64(ui.Counter) / 32
                        scale.SetR(float32(v))
                        scale.SetG(float32(v))
                        scale.SetB(float32(v) * 1.8)
                    }

                    spellTitleNormalFont.Print(pageImage, x, y, 1, scale, spell.Name)
                    y += float64(spellTitleNormalFont.Height())

                    if page.IsResearch {
                        turns := spell.ResearchCost / researchPoints
                        if spell.Name == researchingSpell.Name {
                            turns = (spell.ResearchCost - researchProgress) / researchPoints
                        }
                        if turns < 1 {
                            turns = 1
                        }
                        turnString := "turn"
                        if turns > 1 {
                            turnString = "turns"
                        }
                        spellTextNormalFont.Print(pageImage, x, y, 1, scale, fmt.Sprintf("Research Cost:%v (%v %v)", spell.ResearchCost, turns, turnString))
                        y += float64(spellTextNormalFont.Height())
                    } else {
                        turns := spell.Cost(true) / castingSkill
                        if turns < 1 {
                            turns = 1
                        }
                        turnString := "turn"
                        if turns > 1 {
                            turnString = "turns"
                        }
                        spellTextNormalFont.Print(pageImage, x, y, 1, scale, fmt.Sprintf("Casting cost:%v (%v %v)", spell.Cost(true), turns, turnString))
                        y += float64(spellTextNormalFont.Height())
                    }

                    wrapped := getSpellDescriptionNormalText(spell.Index)
                    spellTextNormalFont.RenderWrapped(pageImage, x, y, wrapped, scale, false)

                    if !flipping && learnedSpell.Name == spell.Name && !learnSpellAnimation.Done() {
                        animationOptions := options
                        animationOptions.GeoM.Reset()
                        animationOptions.GeoM.Translate(x, spellY - 2)
                        pageImage.DrawImage(learnSpellAnimation.Frame(), &animationOptions)
                    }

                } else {
                    spellTitleAlienFont.Print(pageImage, x, y, 1, options.ColorScale, spell.Name)
                    wrapped := getSpellDescriptionAlienText(spell.Index)
                    spellTextAlienFont.RenderWrapped(pageImage, x, y + 10, wrapped, options.ColorScale, false)
                }
            }
        }
    }

    // lazily construct the page graphics, which consists of the section title and 4 spell descriptions
    getHalfPageImage := func(halfPage int) *ebiten.Image {
        image, ok := halfPageCache[halfPage]
        if ok {
            return image
        }

        pageImage := ebiten.NewImage(155, 170)
        pageImage.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 0})

        if halfPage < len(halfPages) {
            renderPage(halfPages[halfPage], true, pageImage, ebiten.DrawImageOptions{})
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

    if researchingSpell.Valid() {
        currentPage := 0

        loop:
        for page, halfPage := range halfPages {
            for _, spell := range halfPage.Spells.Spells {
                if spell.Name == researchingSpell.Name {
                    currentPage = page
                    break loop
                }
            }
        }

        // force it to be even
        currentPage -= currentPage % 2
        showLeftPage = currentPage
        showRightPage = currentPage + 1
    }

    flipping := false

    quit := false

    posX := 0
    posY := 0
    elements = append(elements, &uilib.UIElement{
        Rect: image.Rect(0, 0, data.ScreenWidth, data.ScreenHeight),
        Inside: func (this *uilib.UIElement, x, y int){
            posX = x
            posY = y
        },
        LeftClick: func(this *uilib.UIElement){
            if pickResearchSpell {
                if posY >= 35 && posY < 35 + 35 * 4 {
                    spellIndex := (posY - 35) / 35

                    // left page
                    if posX < 160 {
                        if spellIndex >= 0 && spellIndex < len(researchPage1.Spells.Spells) {
                            *chosenSpell = researchPage1.Spells.Spells[spellIndex]
                            quit = true
                        }
                    } else {
                        // right page
                        if spellIndex >= 0 && spellIndex < len(researchPage2.Spells.Spells) {
                            *chosenSpell = researchPage2.Spells.Spells[spellIndex]
                            quit = true
                        }
                    }
                }
            } else {
                getAlpha = ui.MakeFadeOut(fadeSpeed)
                ui.AddDelay(fadeSpeed, func(){
                    // ui.RemoveElements(elements)
                    quit = true
                })
            }
        },
        NotLeftClicked: func(this *uilib.UIElement){
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("scroll.lbx", 6, 0)

            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(background, &options)

            if showLeftPage >= 0 {
                renderPage(halfPages[showLeftPage], false, screen, options)
            }

            if showRightPage < len(halfPages) {
                rightOptions := options
                rightOptions.GeoM.Translate(148, 0)
                rightPage := screen.SubImage(image.Rect(148, 0, screen.Bounds().Dx(), screen.Bounds().Dy())).(*ebiten.Image)
                renderPage(halfPages[showRightPage], false, rightPage, rightOptions)
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

            if pickResearchSpell {
                chooseFont.PrintCenter(screen, 160, 180, 1, options.ColorScale, "Choose a new spell to research")
            }

        },
    })

    doLeftPageTurn := func(){
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
    }

    doRightPageTurn := func(){
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
    }

    // left page turn
    leftTurn, _ := imageCache.GetImage("scroll.lbx", 7, 0)
    leftRect := image.Rect(15, 9, 15 + leftTurn.Bounds().Dx(), 9 + leftTurn.Bounds().Dy())
    elements = append(elements, &uilib.UIElement{
        Rect: leftRect,
        LeftClick: func(this *uilib.UIElement){
            doLeftPageTurn()
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
        LeftClick: func(this *uilib.UIElement){
            doRightPageTurn()
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

    ui.SetElementsFromArray(elements)

    *drawFunc = func(screen *ebiten.Image){
        ui.Draw(ui, screen)
    }

    yield()

    for !quit {
        ui.StandardUpdate()
        if ui.Counter % 4 == 0 {
            learnSpellAnimation.Next()
        }
        yield()
    }

    yield()
}

func CastLeftSideDistortions1(page *ebiten.Image) util.Distortion {
    return util.Distortion{
        Top: image.Pt(page.Bounds().Dx()/2 + 25, 12),
        Bottom: image.Pt(page.Bounds().Dx()/2 + 25, page.Bounds().Dy() - 4),
        Segments: []util.Segment{
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 40, -5),
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
                Top: image.Pt(page.Bounds().Dx()/2 + 100, -3),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 100, page.Bounds().Dy() - 22),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 135, 16),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 135, page.Bounds().Dy() - 3),
            },
        },
    }
}

func CastLeftSideDistortions2(page *ebiten.Image) util.Distortion {
    return util.Distortion{
        Top: image.Pt(page.Bounds().Dx()/2 + 25, 9),
        Bottom: image.Pt(page.Bounds().Dx()/2 + 25, page.Bounds().Dy() - 2),
        Segments: []util.Segment{
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 40, -12),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 40, page.Bounds().Dy() - 33),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 58, -13),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 58, page.Bounds().Dy() - 43),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 73, -20),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 73, page.Bounds().Dy() - 43),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 90, -3),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 90, page.Bounds().Dy() - 34),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 99, 1),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 99, page.Bounds().Dy() - 24),
            },
        },
    }
}

func CastRightSideDistortions1(page *ebiten.Image) util.Distortion {
    offset := 60
    return util.Distortion{
        Top: image.Pt(page.Bounds().Dx()/2 - 110 + offset - 3, 2),
        Bottom: image.Pt(page.Bounds().Dx()/2 - 110 + offset, page.Bounds().Dy() - 28),
        Segments: []util.Segment{
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 90 + offset, -0),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 90 + offset, page.Bounds().Dy() - 43),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 73 + offset, -10),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 73 + offset, page.Bounds().Dy() - 43),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 58 + offset, -13),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 58 + offset, page.Bounds().Dy() - 35),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 40 + offset, 4),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 40 + offset, page.Bounds().Dy() - 21),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 38 + offset, 18),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 38 + offset, page.Bounds().Dy() - 10),
            },
        },
    }
}

func CastRightSideDistortions2(page *ebiten.Image) util.Distortion {
    offset := 40
    return util.Distortion{
        Top: image.Pt(page.Bounds().Dx()/2 - 130 + offset + 2, 15),
        Bottom: image.Pt(page.Bounds().Dx()/2 - 130 + offset + 2, page.Bounds().Dy() - 3),

        Segments: []util.Segment{
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 100 + offset, -0),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 100 + offset, page.Bounds().Dy() - 18),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 80 + offset, -10),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 80 + offset, page.Bounds().Dy() - 27),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 60 + offset, -10),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 60 + offset, page.Bounds().Dy() - 31),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 40 + offset, -3),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 40 + offset, page.Bounds().Dy() - 28),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 30 + offset, -2),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 30 + offset, page.Bounds().Dy() - 22),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 20 + offset + 3, 15),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 20 + offset + 1, page.Bounds().Dy() - 5),
            },
        },
    }
}

// FIXME: take in the wizard/player that is casting the spell
// chosenCallback is invoked when the spellbook ui goes away, either because the user
// selected a spell or because they canceled the ui
// if a spell is chosen then it will be passed in as the first argument to the callback along with true
// if the ui is cancelled then the second argument will be false
func MakeSpellBookCastUI(ui *uilib.UI, cache *lbx.LbxCache, spells Spells, castingSkill int, currentSpell Spell, currentProgress int, overland bool, chosenCallback func(Spell, bool)) []*uilib.UIElement {
    var elements []*uilib.UIElement

    imageCache := util.MakeImageCache(cache)

    getAlpha := ui.MakeFadeIn(7)

    // value needs to be one so that the ColorScale later on works
    black := color.RGBA{R: 1, G: 1, B: 1, A: 0xff}

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

    renderPage := func(screen *ebiten.Image, options ebiten.DrawImageOptions, page Page, highlightedSpell Spell){
        // section := spells.Spells[0].Section

        titleX, titleY := options.GeoM.Apply(60, 1)

        titleFont.PrintCenter(screen, titleX, titleY, 1, options.ColorScale, page.Title)
        gibberish, _ := imageCache.GetImage("spells.lbx", 10, 0)
        gibberishHeight := 18

        options2 := options
        options2.GeoM.Translate(0, 15)
        for _, spell := range page.Spells.Spells {

            // invalid spell?
            if spell.Invalid() {
                continue
            }

            spellOptions := options2

            textColorScale := spellOptions.ColorScale

            if currentSpell.Name == spell.Name {
                v := math.Cos(float64(ui.Counter) / 5) * 64 + 128
                textColorScale.SetR(float32(v))
                textColorScale.SetG(float32(v))
                textColorScale.SetB(float32(v))
            } else if highlightedSpell.Name == spell.Name {
                // spellOptions.ColorScale.Scale(1.5, 1, 1, 1)
                r := math.Cos(float64(ui.Counter) / 5) * 128 + 128
                textColorScale.SetR(float32(r))
            }

            spellX, spellY := spellOptions.GeoM.Apply(0, 0)

            costRemaining := spell.Cost(overland)
            if spell.Name == currentSpell.Name {
                costRemaining -= currentProgress
            }

            infoFont.Print(screen, spellX, spellY, 1, textColorScale, spell.Name)
            infoFont.PrintRight(screen, spellX + 124, spellY, 1, textColorScale, fmt.Sprintf("%v MP", costRemaining))
            icon := getMagicIcon(spell)

            nameLength := infoFont.MeasureTextWidth(spell.Name, 1) + 1
            mpLength := infoFont.MeasureTextWidth(fmt.Sprintf("%v MP", costRemaining), 1)

            gibberishPart := gibberish.SubImage(image.Rect(0, 0, gibberish.Bounds().Dx(), gibberishHeight)).(*ebiten.Image)

            partIndex := 0
            partHeight := 20

            subLines := 6

            part1 := gibberishPart.SubImage(image.Rect(int(nameLength), partIndex * partHeight, int(nameLength) + gibberishPart.Bounds().Dx() - int(nameLength + mpLength), partIndex * partHeight + subLines)).(*ebiten.Image)

            part1Options := options2
            part1Options.GeoM.Translate(nameLength, 0)
            screen.DrawImage(part1, &part1Options)

            iconCount := costRemaining / int(math.Max(1, float64(castingSkill)))
            if iconCount < 1 {
                iconCount = 1
            }

            iconOptions := spellOptions
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

            icon1Width := 0

            for i := 0; i < icons1; i++ {
                screen.DrawImage(icon, &iconOptions)
                iconOptions.GeoM.Translate(float64(icon.Bounds().Dx()) + 1, 0)
                icon1Width += icon.Bounds().Dx() + 1
            }

            if costRemaining < castingSkill {
                x, y := iconOptions.GeoM.Apply(0, 0)
                x += 2
                infoFont.Print(screen, x, y, 1, spellOptions.ColorScale, "Instant")
                icon1Width += int(infoFont.MeasureTextWidth("Instant", 1)) + 2
                iconOptions.GeoM.Translate(infoFont.MeasureTextWidth("Instant", 1) + 2, 0)
            }

            part2 := gibberishPart.SubImage(image.Rect(icon1Width + 3, partIndex * partHeight + subLines, gibberish.Bounds().Dx(), partIndex * partHeight + subLines * 2)).(*ebiten.Image)
            part2Options := iconOptions
            part2Options.GeoM.Translate(3, 0)
            screen.DrawImage(part2, &part2Options)

            part3Options.GeoM.Translate(0, float64(icon.Bounds().Dy()+1))

            for i := 0; i < iconCount; i++ {
                screen.DrawImage(icon, &part3Options)
                part3Options.GeoM.Translate(float64(icon.Bounds().Dx()) + 1, 0)
            }

            part3 := gibberishPart.SubImage(image.Rect((icon.Bounds().Dx() + 1) * iconCount, partIndex * partHeight + subLines * 2, gibberish.Bounds().Dx(), partIndex * partHeight + subLines * 3)).(*ebiten.Image)
            screen.DrawImage(part3, &part3Options)

            options2.GeoM.Translate(0, 22)
        }
    }

    // lazily construct the page graphics, which consists of the section title and 4 spell descriptions
    getPageImage := func(page int) *ebiten.Image {
        cached, ok := pageCache[page]
        if ok {
            return cached
        }

        out := ebiten.NewImage(120, 154)
        out.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 0})

        if page < len(spellPages) {
            var options ebiten.DrawImageOptions
            renderPage(out, options, spellPages[page], Spell{})
            // out.Fill(color.RGBA{R: 255, G: 0, B: 0, A: 255})

            /*
            alpha := uint8(64)
            vector.DrawFilledRect(out, 0, 0, 30, float32(out.Bounds().Dy()), util.PremultiplyAlpha(color.RGBA{R: 255, G: 0, B: 0, A: alpha}), false)
            vector.DrawFilledRect(out, 30, 0, 30, float32(out.Bounds().Dy()), util.PremultiplyAlpha(color.RGBA{R: 0, G: 255, B: 0, A: alpha}), false)
            vector.DrawFilledRect(out, 60, 0, 30, float32(out.Bounds().Dy()), util.PremultiplyAlpha(color.RGBA{R: 0, G: 0, B: 255, A: alpha}), false)
            vector.DrawFilledRect(out, 90, 0, 30, float32(out.Bounds().Dy()), util.PremultiplyAlpha(color.RGBA{R: 255, G: 255, B: 0, A: alpha}), false)
            vector.StrokeLine(out, 0, 80, float32(out.Bounds().Dx()), 80, 2, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
            */
        }

        // vector.StrokeRect(out, 1, 1, float32(out.Bounds().Dx()-1), float32(out.Bounds().Dy()-10), 1, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
        pageCache[page] = out
        return out
    }

    // the spell the user is mousing over
    var highlightedSpell Spell

    // invoke to shut down the ui and return a result
    var shutdown func(Spell, bool)

    var spellButtons []*uilib.UIElement
    setupSpells := func(page int) {
        ui.RemoveElements(spellButtons)
        spellButtons = nil

        if page < 0 || page >= len(spellPages) {
            return
        }

        makeSpell := func(rect image.Rectangle, spell Spell) *uilib.UIElement {
            return &uilib.UIElement{
                Rect: rect,
                Layer: 1,
                Inside: func(this *uilib.UIElement, x, y int){
                    highlightedSpell = spell
                },
                NotInside: func(this *uilib.UIElement){
                    if highlightedSpell.Name == spell.Name {
                        highlightedSpell = Spell{}
                    }
                },
                LeftClick: func(this *uilib.UIElement){
                    // if the user is already casting a spell then ask them if they want to abort that spell
                    if currentSpell.Valid() {
                        confirm := func(){
                            // if the user clicked on the same spell being cast then select an invalid spell, which
                            // is the same thing as not casting any spell
                            if spell.Name == currentSpell.Name {
                                shutdown(Spell{}, true)
                            } else {
                                shutdown(spell, true)
                            }
                        }
                        message := fmt.Sprintf("Do you wish to abort your %v spell?", currentSpell.Name)
                        ui.AddElements(uilib.MakeConfirmDialogWithLayer(ui, cache, &imageCache, 2, message, confirm, func(){}))
                    } else {
                        // log.Printf("Click on spell %v", spell)
                        shutdown(spell, true)
                    }
                },
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    // vector.StrokeRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Max.X - rect.Min.X), float32(rect.Max.Y - rect.Min.Y), 1, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
                },
            }
        }

        leftSpells := spellPages[page].Spells.Spells

        for i, spell := range leftSpells {

            x1 := 24
            y1 := 30 + i * 22
            width := 122
            height := 20

            rect := image.Rect(0, 0, width, height).Add(image.Pt(x1, y1))
            spellButtons = append(spellButtons, makeSpell(rect, spell))
        }

        if page + 1 < len(spellPages) {
            rightSpells := spellPages[page+1].Spells.Spells

            for i, spell := range rightSpells {
                x1 := 159
                y1 := 30 + i * 22
                width := 122
                height := 20

                rect := image.Rect(0, 0, width, height).Add(image.Pt(x1, y1))
                spellButtons = append(spellButtons, makeSpell(rect, spell))
            }
        }

        ui.AddElements(spellButtons)
    }

    currentPage := 0

    if currentSpell.Valid() {
        loop:
        for page, halfPage := range spellPages {
            for _, spell := range halfPage.Spells.Spells {
                if spell.Name == currentSpell.Name {
                    currentPage = page
                    break loop
                }
            }
        }

        // force it to be even
        currentPage -= currentPage % 2
    }

    bookFlip, _ := imageCache.GetImages("book.lbx", 0)
    bookFlipIndex := uint64(0)
    bookFlipReverse := false
    bookFlipSpeed := uint64(6)
    flipping := false

    showPageLeft := 0
    showPageRight := 0
    pageSideLeft := 0
    pageSideRight := 0

    _ = pageSideLeft
    _ = pageSideRight

    shutdown = func(spell Spell, picked bool){
        getAlpha = ui.MakeFadeOut(7)
        ui.AddDelay(7, func(){
            setupSpells(-1)
            ui.RemoveElements(elements)
            chosenCallback(spell, picked)
        })
    }

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        /*
        NotLeftClicked: func(this *uilib.UIElement){
            getAlpha = ui.MakeFadeOut(7)
            ui.AddDelay(7, func(){
                ui.RemoveElements(elements)
                setupSpells(-1)

                log.Printf("Chose spell %+v", chosenSpell)
            })
        },
        */
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            // FIXME: do the whole page flipping thing with distorted pages
            vector.DrawFilledRect(screen, 0, 0, float32(screen.Bounds().Dx()), float32(screen.Bounds().Dy()), color.RGBA{R: 0, G: 0, B: 0, A: 128}, false)

            background, _ := imageCache.GetImage("spells.lbx", 0, 0)
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            options.GeoM.Translate(10, 10)
            screen.DrawImage(background, &options)

            flipOptions := options

            if flipping {
                options.GeoM.Translate(15, 5)
                renderPage(screen, options, spellPages[showPageLeft], Spell{})

                if showPageRight < len(spellPages) {
                    options.GeoM.Translate(134, 0)
                    renderPage(screen, options, spellPages[showPageRight], Spell{})
                }

                if bookFlipIndex > 0 && (ui.Counter - bookFlipIndex) / bookFlipSpeed < uint64(len(bookFlip)) {
                    index := (ui.Counter - bookFlipIndex) / bookFlipSpeed
                    if bookFlipReverse {
                        index = uint64(len(bookFlip)) - 1 - index
                    }

                    // index = 3

                    flipOptions.GeoM.Translate(-17, -10)
                    screen.DrawImage(bookFlip[index], &flipOptions)

                    if index == 0 {
                        if pageSideLeft >= 0 {
                            leftSide := getPageImage(pageSideLeft)
                            util.DrawDistortion(screen, bookFlip[index], leftSide, CastLeftSideDistortions1(bookFlip[index]), flipOptions)
                        }
                    } else if index == 1 {
                        if pageSideLeft >= 0 {
                            leftSide := getPageImage(pageSideLeft)
                            util.DrawDistortion(screen, bookFlip[index], leftSide, CastLeftSideDistortions2(bookFlip[index]), flipOptions)
                        }
                    } else if index == 2 {
                        if pageSideRight < len(spellPages) {
                            rightSide := getPageImage(pageSideRight)
                            util.DrawDistortion(screen, bookFlip[index], rightSide, CastRightSideDistortions1(bookFlip[index]), flipOptions)
                        }
                    } else if index == 3 {
                        if pageSideRight < len(spellPages) {
                            rightSide := getPageImage(pageSideRight)
                            util.DrawDistortion(screen, bookFlip[index], rightSide, CastRightSideDistortions2(bookFlip[index]), flipOptions)
                        }
                    }
                }

            } else {
                options.GeoM.Translate(15, 5)
                if currentPage < len(spellPages) {
                    renderPage(screen, options, spellPages[currentPage], highlightedSpell)
                }

                if currentPage + 1 < len(spellPages) {
                    options.GeoM.Translate(134, 0)
                    // screen.DrawImage(right, &options)
                    renderPage(screen, options, spellPages[currentPage+1], highlightedSpell)
                }
            }
        },
    })

    cancelRect := image.Rect(0, 0, 18, 25).Add(image.Pt(170, 170))
    elements = append(elements, &uilib.UIElement{
        Rect: cancelRect,
        Layer: 1,
        LeftClick: func(this *uilib.UIElement){
            shutdown(Spell{}, false)
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            // vector.StrokeRect(screen, float32(cancelRect.Min.X), float32(cancelRect.Min.Y), float32(cancelRect.Dx()), float32(cancelRect.Dy()), 1, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
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
            if currentPage + 2 < len(spellPages) && !flipping {
                flipping = true
                bookFlipReverse = false
                bookFlipIndex = ui.Counter

                showPageRight = currentPage + 3
                pageSideLeft = currentPage + 1
                pageSideRight = currentPage + 2
                showPageLeft = currentPage

                ui.AddDelay(bookFlipSpeed * uint64(len(bookFlip)), func (){
                    flipping = false
                    currentPage += 2
                    setupSpells(currentPage)
                })
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            if currentPage + 2 < len(spellPages) {
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(getAlpha())
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
            if currentPage >= 2 && !flipping {
                flipping = true
                bookFlipReverse = true
                bookFlipIndex = ui.Counter

                showPageRight = currentPage + 1
                showPageLeft = currentPage - 2
                pageSideLeft = currentPage - 1
                pageSideRight = currentPage

                ui.AddDelay(bookFlipSpeed * uint64(len(bookFlip) - 1), func (){
                    flipping = false
                    currentPage -= 2
                    setupSpells(currentPage)
                })
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            if currentPage > 0 {
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(getAlpha())
                options.GeoM.Translate(float64(pageTurnLeftRect.Min.X), float64(pageTurnLeftRect.Min.Y))
                screen.DrawImage(pageTurnLeft, &options)
            }

        },
    })

    return elements
}
