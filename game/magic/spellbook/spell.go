package spellbook

import (
    "fmt"
    "image"
    "image/color"
    "math"
    "log"
    "slices"
    "cmp"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    helplib "github.com/kazzmir/master-of-magic/game/magic/help"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"

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

func magicToOrder(magic data.MagicType) int {
    switch magic {
        case data.NatureMagic: return 0
        case data.SorceryMagic: return 1
        case data.ChaosMagic: return 2
        case data.LifeMagic: return 3
        case data.DeathMagic: return 4
        case data.ArcaneMagic: return 5
    }

    return -1
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
        Top: image.Pt(page.Bounds().Dx()/2 + (offset - 130), 5),
        Bottom: image.Pt(page.Bounds().Dx()/2 + (offset - 130), page.Bounds().Dy() - 0),

        Segments: []util.Segment{
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + (offset - 100), -0),
                Bottom: image.Pt(page.Bounds().Dx()/2 + (offset - 100), page.Bounds().Dy() - 22),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + (offset - 80), -10),
                Bottom: image.Pt(page.Bounds().Dx()/2 + (offset - 80), page.Bounds().Dy() - 30),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + (offset - 60), -10),
                Bottom: image.Pt(page.Bounds().Dx()/2 + (offset - 60), page.Bounds().Dy() - 33),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + (offset - 40), 0),
                Bottom: image.Pt(page.Bounds().Dx()/2 + (offset - 40), page.Bounds().Dy() - 25),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + (offset - 20), 5),
                Bottom: image.Pt(page.Bounds().Dx()/2 + (offset - 20), page.Bounds().Dy() - 12),
            },
        },
    }
}

func RightSideDistortions1(page *ebiten.Image) util.Distortion {
    offset := 50
    return util.Distortion{
        Top: image.Pt(page.Bounds().Dx()/2 + (offset - 110), -10),
        Bottom: image.Pt(page.Bounds().Dx()/2 + (offset - 110), page.Bounds().Dy() - 12),
        Segments: []util.Segment{
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + (offset - 90), -0),
                Bottom: image.Pt(page.Bounds().Dx()/2 + (offset - 90), page.Bounds().Dy() - 22),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + (offset - 73), -20),
                Bottom: image.Pt(page.Bounds().Dx()/2 + (offset - 73), page.Bounds().Dy() - 35),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + (offset - 58), -13),
                Bottom: image.Pt(page.Bounds().Dx()/2 + (offset - 58), page.Bounds().Dy() - 35),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + (offset - 40), 0),
                Bottom: image.Pt(page.Bounds().Dx()/2 + (offset - 40), page.Bounds().Dy() - 28),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + (offset - 20), 5),
                Bottom: image.Pt(page.Bounds().Dx()/2 + (offset - 20), page.Bounds().Dy() - 15),
            },
        },
    }
}

// flipping the page to the right
/*
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
*/

type SpellCaster interface {
    // given research points and a spell, return the actual number of research points per turn
    // made towards that spell
    ComputeEffectiveResearchPerTurn(float64, Spell) int
    // the actual cost to cast the spell
    ComputeEffectiveSpellCost(spell Spell, overland bool) int
}

type Wizard interface {
    RetortEnabled(retort data.Retort) bool
    MagicLevel(magic data.MagicType) int
}

// the casting cost of a spell can be reduced based on retorts/spell books
func ComputeSpellCost(wizard Wizard, spell Spell, overland bool, hasEvilOmens bool) int {
    base := float64(spell.Cost(overland))
    modifier := float64(0)

    if wizard.RetortEnabled(data.RetortRunemaster) && spell.Magic == data.ArcaneMagic {
        modifier += 0.25
    }

    if wizard.RetortEnabled(data.RetortChaosMastery) && spell.Magic == data.ChaosMagic {
        modifier += 0.15
    }

    if wizard.RetortEnabled(data.RetortNatureMastery) && spell.Magic == data.NatureMagic {
        modifier += 0.15
    }

    if wizard.RetortEnabled(data.RetortSorceryMastery) && spell.Magic == data.SorceryMagic {
        modifier += 0.15
    }

    if wizard.RetortEnabled(data.RetortConjurer) && spell.IsSummoning() {
        modifier += 0.25
    }

    if wizard.RetortEnabled(data.RetortArtificer) && (spell.Name == "Enchant Item" || spell.Name == "Create Artifact") {
        modifier += 0.5
    }

    // for each book above 7, reduce cost by 10%
    realmBooks := max(0, wizard.MagicLevel(spell.Magic) - 7)
    modifier += float64(realmBooks) * 0.1

    evilOmens := float64(1.0)

    if hasEvilOmens {
        if spell.Magic == data.LifeMagic || spell.Magic == data.NatureMagic {
            evilOmens = 1.5
        }
    }

    return int(max(0, base * (1 - modifier) * evilOmens))
}

/* three modes:
 * 1. when a new spell is learned, flip to the page where the spell would go and show the sparkle animation over the new spell
 * 2. flip to the 'research spells' page and let the user pick a new spell
 * 3. show book and let user flip between pages. on the 'research spells' page, show currently
 *     researching spell as glowing text
 *
 * This function does all 3, which makes it kind of ugly and has too many parameters.
 */
func ShowSpellBook(yield coroutine.YieldFunc, cache *lbx.LbxCache, allSpells Spells, knownSpells Spells, researchSpells Spells, researchingSpell Spell, researchProgress int, researchPoints float64, castingSkill int, learnedSpell Spell, pickResearchSpell bool, chosenSpell *Spell, caster SpellCaster, drawFunc *func(screen *ebiten.Image)) {
    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            ui.StandardDraw(screen)
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

    helpLbx, err := cache.GetLbxFile("HELP.LBX")
    if err != nil {
        log.Printf("Unable to read help: %v", err)
        return
    }

    help, err := helplib.ReadHelp(helpLbx, 2)
    if err != nil {
        log.Printf("Unable to read help: %v", err)
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

    // sort research spells by turns to research
    slices.SortFunc(researchSpells.Spells, func(a, b Spell) int {

        if a.Magic != b.Magic {
            return cmp.Compare(magicToOrder(a.Magic), magicToOrder(b.Magic))
        }

        turnsA := a.ResearchCost / caster.ComputeEffectiveResearchPerTurn(researchPoints, a)
        turnsB := b.ResearchCost / caster.ComputeEffectiveResearchPerTurn(researchPoints, b)

        return cmp.Compare(turnsA, turnsB)
    })

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

    type HoverTime struct {
        // the first time the user hovered over the area
        On uint64
        // the first time the user left the area
        Off uint64
    }

    // map from spell number to its time, where number 0 means left page spell 0, and 4 means right page spell 0
    hoverTime := make(map[string]HoverTime)

    renderPage := func(page Page, flipping bool, pageImage *ebiten.Image, options ebiten.DrawImageOptions){
        if len(page.Spells.Spells) > 0 || page.ForceRender {
            // var options ebiten.DrawImageOptions
            // options.GeoM.Translate(0, 0)

            // section := pageSpells.Spells[0].Section
            titleX, titleY := options.GeoM.Apply(float64(90), float64(11))

            titleFont.PrintOptions(pageImage, titleX, titleY, font.FontOptions{Justify: font.FontJustifyCenter, Options: &options, Scale: scale.ScaleAmount}, page.Title)

            x, topY := options.GeoM.Apply(float64(25), float64(35))

            for i, spell := range page.Spells.Spells {
                if i >= 4 {
                    break
                }

                y := topY + float64(i * 35)

                spellY := y

                if page.IsResearch || knownSpell(spell) || learnedSpell.Name == spell.Name {
                    var scaleOptions ebiten.DrawImageOptions

                    if !flipping {

                        if researchingSpell.Name == spell.Name {
                            v := 1.5 + (math.Cos(float64(ui.Counter) / 7) * 64 + 64) / float64(64)
                            scaleOptions.ColorScale.SetR(float32(v))
                            scaleOptions.ColorScale.SetG(float32(v))
                            scaleOptions.ColorScale.SetB(float32(v) * 1.8)
                        } else if learnedSpell.Name == spell.Name && !learnSpellAnimation.Done() {
                            v := 1.5 + float64(ui.Counter) / 32
                            scaleOptions.ColorScale.SetR(float32(v))
                            scaleOptions.ColorScale.SetG(float32(v))
                            scaleOptions.ColorScale.SetB(float32(v) * 1.8)
                        } else if pickResearchSpell {
                            hover, ok := hoverTime[spell.Name]
                            if ok {
                                max := float32(5)
                                if hover.On > 0 {
                                    v := float32(ui.Counter - hover.On) / 2
                                    if v > max {
                                        v = max
                                    }
                                    scaleOptions.ColorScale.SetR(v)
                                } else if hover.Off > 0 {
                                    v := max - float32(ui.Counter - hover.Off) / 3
                                    if v > 0 {
                                        scaleOptions.ColorScale.SetR(v)
                                    }
                                }
                            }
                        }

                    }

                    spellTitleNormalFont.PrintOptions(pageImage, x, y, font.FontOptions{Scale: scale.ScaleAmount, Options: &scaleOptions}, spell.Name)
                    y += float64(spellTitleNormalFont.Height())

                    if page.IsResearch {
                        turns := spell.ResearchCost / caster.ComputeEffectiveResearchPerTurn(researchPoints, spell)
                        if spell.Name == researchingSpell.Name {
                            turns = (spell.ResearchCost - researchProgress) / caster.ComputeEffectiveResearchPerTurn(researchPoints, spell)
                        }
                        if turns < 1 {
                            turns = 1
                        }
                        turnString := "turn"
                        if turns > 1 {
                            turnString = "turns"
                        }
                        spellTextNormalFont.PrintOptions(pageImage, x, y, font.FontOptions{Scale: scale.ScaleAmount, Options: &scaleOptions}, fmt.Sprintf("Research Cost:%v (%v %v)", spell.ResearchCost, turns, turnString))
                        y += float64(spellTextNormalFont.Height())
                    } else {
                        spellCost := caster.ComputeEffectiveSpellCost(spell, true)

                        turns := spellCost / castingSkill
                        if turns < 1 {
                            turns = 1
                        }
                        turnString := "turn"
                        if turns > 1 {
                            turnString = "turns"
                        }
                        spellTextNormalFont.PrintOptions(pageImage, x, y, font.FontOptions{Scale: scale.ScaleAmount, Options: &scaleOptions}, fmt.Sprintf("Casting cost:%v (%v %v)", spellCost, turns, turnString))
                        y += float64(spellTextNormalFont.Height())
                    }

                    wrapped := getSpellDescriptionNormalText(spell.Index)
                    spellTextNormalFont.RenderWrapped(pageImage, x, y, wrapped, font.FontOptions{Scale: scale.ScaleAmount, Options: &scaleOptions})

                    if !flipping && learnedSpell.Name == spell.Name && !learnSpellAnimation.Done() {
                        animationOptions := options
                        animationOptions.GeoM.Reset()
                        animationOptions.GeoM.Translate(x, spellY - 2)
                        scale.DrawScaled(pageImage, learnSpellAnimation.Frame(), &animationOptions)
                    }

                } else {
                    spellTitleAlienFont.PrintOptions(pageImage, x, y, font.FontOptions{Scale: scale.ScaleAmount, Options: &options}, spell.Name)
                    wrapped := getSpellDescriptionAlienText(spell.Index)
                    spellTextAlienFont.RenderWrapped(pageImage, x, y + float64(10), wrapped, font.FontOptions{Scale: scale.ScaleAmount, Options: &options})
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

        pageImage := ebiten.NewImage(scale.Scale2(155, 170))
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

    openPage := func (findSpell Spell, isOk func(Page) bool){
        currentPage := 0
        found := false

        loop:
        for page, halfPage := range halfPages {
            for _, spell := range halfPage.Spells.Spells {
                if spell.Name == findSpell.Name && isOk(halfPage) {
                    currentPage = page
                    found = true
                    break loop
                }
            }
        }

        if found {
            // force it to be even
            currentPage -= currentPage % 2
            showLeftPage = currentPage
            showRightPage = currentPage + 1
        }
    }

    if researchingSpell.Valid() {
        openPage(researchingSpell, func (page Page) bool {
            return page.IsResearch
        })
    }

    if learnedSpell.Valid() {
        openPage(learnedSpell, func (page Page) bool {
            return !page.IsResearch
        })
    }

    flipping := false

    quit := false

    posX := 0
    posY := 0
    researchSpellIndex := -1

    var rect image.Rectangle
    if pickResearchSpell {
        rect = image.Rect(0, 0, data.ScreenWidth, data.ScreenHeight)
    }

    group := uilib.MakeGroup()
    ui.AddGroup(group)

    elements = append(elements, &uilib.UIElement{
        Rect: rect,
        Inside: func (this *uilib.UIElement, x int, y int){
            posX = x
            posY = y

            newIndex := -1

            if posY >= 35 && posY < (35 + 35 * 4) {
                newIndex = (posY - 35) / (35)

                if posX >= 160 {
                    newIndex += 4
                }
            } else {
                newIndex = -1
            }

            if pickResearchSpell && newIndex != researchSpellIndex {

                indexToName := func(index int) string {
                    if index >= 0 && index < len(researchPage1.Spells.Spells) {
                        return researchPage1.Spells.Spells[index].Name
                    } else if index >= 4 && index < len(researchPage2.Spells.Spells) + 4 {
                        return researchPage2.Spells.Spells[index - 4].Name
                    }

                    return ""
                }

                hoverTime[indexToName(researchSpellIndex)] = HoverTime{Off: ui.Counter}
                hoverTime[indexToName(newIndex)] = HoverTime{On: ui.Counter}
                researchSpellIndex = newIndex
            }
        },
        LeftClick: func(this *uilib.UIElement){
            if pickResearchSpell {
                if researchSpellIndex >= 0 && researchSpellIndex < len(researchPage1.Spells.Spells) {
                    *chosenSpell = researchPage1.Spells.Spells[researchSpellIndex]
                    getAlpha = ui.MakeFadeOut(fadeSpeed)
                    ui.AddDelay(fadeSpeed, func(){
                        // ui.RemoveElements(elements)
                        quit = true
                    })

                } else {
                    use := researchSpellIndex - 4
                    // right page
                    if use >= 0 && use < len(researchPage2.Spells.Spells) {
                        *chosenSpell = researchPage2.Spells.Spells[use]
                        getAlpha = ui.MakeFadeOut(fadeSpeed)
                        ui.AddDelay(fadeSpeed, func(){
                            // ui.RemoveElements(elements)
                            quit = true
                        })

                    }
                }
            }
        },
        RightClick: func(element *uilib.UIElement){
            // FIXME: when not in research mode, this should still allow the user to right click on a known spell to
            // view its help entry
            var spell *Spell
            if researchSpellIndex >= 0 && researchSpellIndex < len(researchPage1.Spells.Spells) {
                spell = &researchPage1.Spells.Spells[researchSpellIndex]
            } else {
                use := researchSpellIndex - 4
                if use >= 0 && use < len(researchPage2.Spells.Spells) {
                    spell = &researchPage2.Spells.Spells[use]
                }
            }
            if spell != nil {
                helpEntries := help.GetEntriesByName(spell.Name)
                if helpEntries != nil {
                    group.AddElement(uilib.MakeHelpElementWithLayer(group, cache, &imageCache, 2, helpEntries[0], helpEntries[1:]...))
                }
            }
        },
        NotLeftClicked: func(this *uilib.UIElement){
            if !pickResearchSpell {
                getAlpha = ui.MakeFadeOut(fadeSpeed)
                ui.AddDelay(fadeSpeed, func(){
                    // ui.RemoveElements(elements)
                    quit = true
                })
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("scroll.lbx", 6, 0)

            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, background, &options)

            if showLeftPage >= 0 {
                renderPage(halfPages[showLeftPage], false, screen, options)
            }

            if showRightPage < len(halfPages) {
                rightOptions := options
                rightOptions.GeoM.Translate(float64(148), 0)
                rightPage := screen.SubImage(scale.ScaleRect(image.Rect(148, 0, screen.Bounds().Dx(), screen.Bounds().Dy()))).(*ebiten.Image)
                renderPage(halfPages[showRightPage], false, rightPage, rightOptions)
            }

            animationIndex := ui.Counter
            if bookFlipIndex > 0 && (animationIndex - bookFlipIndex) / bookFlipSpeed < uint64(len(bookFlip)) {
                index := (animationIndex - bookFlipIndex) / bookFlipSpeed
                if bookFlipReverse {
                    index = uint64(len(bookFlip)) - 1 - index
                }
                options.GeoM.Translate(0, 0)
                scale.DrawScaled(screen, bookFlip[index], &options)

                if index == 0 {
                    if flipLeftSide >= 0 {
                        leftSide := getHalfPageImage(flipLeftSide)
                        util.DrawDistortion(screen, bookFlip[index], leftSide, LeftSideDistortions1(bookFlip[index]), scale.ScaleGeom(options.GeoM))
                    }
                } else if index == 1 {
                    if flipLeftSide >= 0 {
                        leftSide := getHalfPageImage(flipLeftSide)
                        util.DrawDistortion(screen, bookFlip[index], leftSide, LeftSideDistortions2(bookFlip[index]), scale.ScaleGeom(options.GeoM))
                    }
                } else if index == 2 {
                    if flipRightSide < len(halfPages) {
                        rightSide := getHalfPageImage(flipRightSide)
                        util.DrawDistortion(screen, bookFlip[index], rightSide, RightSideDistortions1(bookFlip[index]), scale.ScaleGeom(options.GeoM))
                    }
                } else if index == 3 {
                    if flipRightSide < len(halfPages) {
                        rightSide := getHalfPageImage(flipRightSide)
                        util.DrawDistortion(screen, bookFlip[index], rightSide, RightSideDistortions2(bookFlip[index]), scale.ScaleGeom(options.GeoM))
                    }
                }
            }

            if pickResearchSpell {
                chooseFont.PrintOptions(screen, 160, 180, font.FontOptions{Justify: font.FontJustifyCenter, Options: &options, Scale: scale.ScaleAmount}, "Choose a new spell to research")
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
    leftRect := util.ImageRect(15, 9, leftTurn)
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
                scale.DrawScaled(screen, leftTurn, &options)
            }
        },
    })

    // right page turn
    rightTurn, _ := imageCache.GetImage("scroll.lbx", 8, 0)
    rightRect := util.ImageRect(289, 9, rightTurn)
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
                scale.DrawScaled(screen, rightTurn, &options)
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

type SpellbookFonts struct {
    BigOrange *font.Font
}

func MakeSpellbookFonts(cache *lbx.LbxCache) *SpellbookFonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Error loading fonts: %v", err)
        return nil
    }

    return &SpellbookFonts{
        BigOrange: loader(fontslib.LightGradient1),
    }
}

// a popup that allows the user to select an additional amount of mana to add to the spell, upto maximum
func makeAdditionalPowerElements(cache *lbx.LbxCache, imageCache *util.ImageCache, maximum int, okCallback func(amount int)) *uilib.UIElementGroup {
    group := uilib.MakeGroup()

    var layer uilib.UILayer = 2

    x := 320 - 158
    y := 30

    fonts := MakeSpellbookFonts(cache)

    amount := float64(0)

    fadeSpeed := uint64(7)

    getAlpha := group.MakeFadeIn(fadeSpeed)

    group.AddElement(&uilib.UIElement{
        Layer: layer,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("spellscr.lbx", 5, 0)
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            options.GeoM.Translate(float64(x), float64(y))
            scale.DrawScaled(screen, background, &options)

            mx, my := options.GeoM.Apply(float64(8), float64(6))
            fonts.BigOrange.PrintOptions(screen, mx, my, font.FontOptions{Scale: scale.ScaleAmount, Options: &options}, "Additional Power:")
            mx, _ = options.GeoM.Apply(float64(background.Bounds().Dx() - 6), 0)
            fonts.BigOrange.PrintOptions(screen, mx, my, font.FontOptions{Scale: scale.ScaleAmount, Options: &options, Justify: font.FontJustifyRight}, fmt.Sprintf("+%v", math.Round(amount)))
        },
        Order: 0,
    })

    // conveyor
    conveyor, _ := imageCache.GetImageTransform("spellscr.lbx", 4, 0, "conveyor", func (img *image.Paletted) image.Image {
        return img.SubImage(image.Rect(0, 0, img.Bounds().Dx() - 6, img.Bounds().Dy()))
    })
    conveyorRect := util.ImageRect((x + 12), (y + 22), conveyor)
    conveyorX := 0
    doUpdate := false
    group.AddElement(&uilib.UIElement{
        Layer: layer,
        Order: 1,
        Rect: conveyorRect,
        Inside: func(element *uilib.UIElement, x int, y int){
            conveyorX = x
            if doUpdate {
                amount = float64(maximum) * float64(conveyorX) / float64(conveyor.Bounds().Dx())
            }
        },
        LeftClick: func(element *uilib.UIElement){
            doUpdate = true
        },
        LeftClickRelease: func(element *uilib.UIElement){
            doUpdate = false
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            showRect := conveyorRect
            showRect.Max.X = showRect.Min.X + int((float64(conveyor.Bounds().Dx()) * float64(amount) / float64(maximum)))
            area := screen.SubImage(scale.ScaleRect(showRect)).(*ebiten.Image)

            motion := (group.Counter / 2) % uint64(conveyor.Bounds().Dx())

            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            options.GeoM.Translate(float64(conveyorRect.Min.X - conveyor.Bounds().Dx() + int(motion)), float64(conveyorRect.Min.Y))
            scale.DrawScaled(area, conveyor, &options)
            options.GeoM.Reset()
            options.GeoM.Translate(float64(conveyorRect.Min.X + int(motion)), float64(conveyorRect.Min.Y))
            scale.DrawScaled(area, conveyor, &options)

            star, _ := imageCache.GetImage("spellscr.lbx", 3, 0)
            options.GeoM.Reset()
            options.GeoM.Translate(float64(conveyorRect.Min.X) + float64(conveyor.Bounds().Dx()) * amount / float64(maximum) - float64(3), float64(conveyorRect.Min.Y - 1))
            scale.DrawScaled(screen, star, &options)

            // vector.StrokeRect(area, float32(conveyorRect.Min.X), float32(conveyorRect.Min.Y), float32(conveyorRect.Bounds().Dx()), float32(conveyorRect.Bounds().Dy()), 2, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
        },
    })

    // ok button
    okIndex := 0
    okImages, _ := imageCache.GetImages("spellscr.lbx", 42)
    okRect := util.ImageRect((x + 127), (y + 17), okImages[0])
    group.AddElement(&uilib.UIElement{
        Layer: layer,
        Rect: okRect,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(okRect.Min.X), float64(okRect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, okImages[okIndex], &options)
        },
        LeftClick: func(element *uilib.UIElement){
            okIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            okIndex = 0

            getAlpha = group.MakeFadeOut(fadeSpeed)
            group.AddDelay(fadeSpeed, func(){
                okCallback(int(math.Round(amount)))
            })
        },
        Order: 1,
    })

    return group
}

// FIXME: take in the wizard/player that is casting the spell
// chosenCallback is invoked when the spellbook ui goes away, either because the user
// selected a spell or because they canceled the ui
// if a spell is chosen then it will be passed in as the first argument to the callback along with true
// if the ui is cancelled then the second argument will be false
func MakeSpellBookCastUI(ui *uilib.UI, cache *lbx.LbxCache, spells Spells, charges map[Spell]int, castingSkill int, currentSpell Spell, currentProgress int, overland bool, caster SpellCaster, currentPage *int, chosenCallback func(Spell, bool)) []*uilib.UIElement {
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

    helpLbx, err := cache.GetLbxFile("HELP.LBX")
    if err != nil {
        log.Printf("Unable to read help: %v", err)
        return nil
    }

    help, err := helplib.ReadHelp(helpLbx, 2)
    if err != nil {
        log.Printf("Unable to read help: %v", err)
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

    group := uilib.MakeGroup()
    ui.AddGroup(group)


    slices.SortFunc(spells.Spells, func(a, b Spell) int {
        // HACK: when casting Create Artifact, the override spell cost is stored in the currentSpell
        // so we just replace the spell being compared with the currentSpell to get the correct cost
        if a.Name == currentSpell.Name {
            a = currentSpell
        }

        if b.Name == currentSpell.Name {
            b = currentSpell
        }

        if a.Magic != b.Magic {
            return cmp.Compare(magicToOrder(a.Magic), magicToOrder(b.Magic))
        }

        costA := caster.ComputeEffectiveSpellCost(a, overland)
        costB := caster.ComputeEffectiveSpellCost(b, overland)

        // log.Printf("Comparing spell costs for %v (%v, override %v) and %v (%v, override %v)", a.Name, costA, a.OverrideCost, b.Name, costB, b.OverrideCost)

        return cmp.Compare(costA, costB)
    })

    useSpells := spells.Copy()
    for spell, charge := range charges {
        if charge > 0 {
            useSpells.AddSpell(spell)
        }
    }

    updateUseSpells := func(filter *set.Set[data.MagicType]) {
        useSpells = spells.Copy()
        for spell, charge := range charges {
            if charge > 0 {
                useSpells.AddSpell(spell)
            }
        }

        if filter.Size() > 0 {
            var out Spells

            for _, magic := range filter.Values() {
                out.AddAllSpells(useSpells.GetSpellsByMagic(magic))
            }

            useSpells = out
        }
    }

    canCast := func (spell Spell) bool {
        // in combat a spell is castable if the caster (a hero) has charges available for that spell,
        // or if the caster has the spell in their spellbook and the casting skill is high enough
        if !overland {
            charges, hasCharge := charges[spell]
            if hasCharge && charges > 0 {
                return true
            }

            // it could be the cast that the spell was granted by a charge, so the spellbook doesn't have it
            if spells.Contains(spell) {
                return caster.ComputeEffectiveSpellCost(spell, overland) <= castingSkill
            } else {
                return false
            }

        }

        return caster.ComputeEffectiveSpellCost(spell, overland) <= castingSkill
    }

    spellPages := computeHalfPages(useSpells, 6)

    renderPage := func(screen *ebiten.Image, options ebiten.DrawImageOptions, page Page, highlightedSpell Spell){
        // section := spells.Spells[0].Section

        titleX, titleY := options.GeoM.Apply(float64(60), float64(1))

        titleFont.PrintOptions(screen, titleX, titleY, font.FontOptions{Justify: font.FontJustifyCenter, Options: &options, Scale: scale.ScaleAmount}, page.Title)
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

            // if spell is too expensive in combat then it is not castable
            if !overland && !canCast(spell) {
                textColorScale.SetR(60)
                textColorScale.SetG(60)
                textColorScale.SetB(60)
            }

            spellX, spellY := spellOptions.GeoM.Apply(0, 0)

            costRemaining := caster.ComputeEffectiveSpellCost(spell, overland)
            if overland {
                if spell.Name == currentSpell.Name {
                    costRemaining = caster.ComputeEffectiveSpellCost(currentSpell, overland)
                    costRemaining -= currentProgress
                }
            }

            var textColorOptions ebiten.DrawImageOptions
            textColorOptions.ColorScale = textColorScale

            infoFont.PrintOptions(screen, spellX, spellY, font.FontOptions{Options: &textColorOptions, Scale: scale.ScaleAmount}, spell.Name)
            infoFont.PrintOptions(screen, spellX + float64(124), spellY, font.FontOptions{Options: &textColorOptions, Justify: font.FontJustifyRight, Scale: scale.ScaleAmount}, fmt.Sprintf("%v MP", costRemaining))
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
            scale.DrawScaled(screen, part1, &part1Options)

            iconCount := 0

            // casting on the overland shows N icons where N is the number of turns it takes to cast the spell
            if overland {
                iconCount = costRemaining / int(math.Max(1, float64(castingSkill)))
                if iconCount < 1 {
                    iconCount = 1
                }
            } else {
                // in combat the number of icons is how many times the spell can be cast given the casting cost of the spell
                // and the casting skill of the user

                // if this ever occurs this is a bug, the cost of a combat spell should never be 0
                if spell.Cost(false) == 0 {
                    iconCount = 1
                } else {
                    iconCount = castingSkill / caster.ComputeEffectiveSpellCost(spell, false)
                }
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

            for range icons1 {
                scale.DrawScaled(screen, icon, &iconOptions)
                iconOptions.GeoM.Translate(float64(icon.Bounds().Dx()) + 1, 0)
                icon1Width += icon.Bounds().Dx() + 1
            }

            if overland && costRemaining < castingSkill {
                x, y := iconOptions.GeoM.Apply(0, 0)
                x += 2
                infoFont.PrintOptions(screen, x, y, font.FontOptions{Options: &spellOptions, Scale: scale.ScaleAmount}, "Instant")
                icon1Width += int(infoFont.MeasureTextWidth("Instant", 1)) + 2
                iconOptions.GeoM.Translate(infoFont.MeasureTextWidth("Instant", 1) + 2, 0)
            }

            part2 := gibberishPart.SubImage(image.Rect(icon1Width + 3, partIndex * partHeight + subLines, gibberish.Bounds().Dx(), partIndex * partHeight + subLines * 2)).(*ebiten.Image)
            part2Options := iconOptions
            part2Options.GeoM.Translate(3, 0)
            scale.DrawScaled(screen, part2, &part2Options)

            part3Options.GeoM.Translate(0, float64(icon.Bounds().Dy()+1))

            for range iconCount {
                scale.DrawScaled(screen, icon, &part3Options)
                part3Options.GeoM.Translate(float64(icon.Bounds().Dx()) + 1, 0)
            }

            part3 := gibberishPart.SubImage(image.Rect((icon.Bounds().Dx() + 1) * iconCount, partIndex * partHeight + subLines * 2, gibberish.Bounds().Dx(), partIndex * partHeight + subLines * 3)).(*ebiten.Image)
            scale.DrawScaled(screen, part3, &part3Options)

            options2.GeoM.Translate(0, 22)
        }
    }

    // lazily construct the page graphics, which consists of the section title and 4 spell descriptions
    getPageImage := func(page int) *ebiten.Image {
        cached, ok := pageCache[page]
        if ok {
            return cached
        }

        out := ebiten.NewImage(scale.Scale2(120, 154))
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
                    if !overland && !canCast(spell) {
                        return
                    }

                    // if the user is already casting a spell then ask them if they want to abort that spell
                    if overland && currentSpell.Valid() {
                        group := uilib.MakeGroup()
                        confirm := func(){
                            // if the user clicked on the same spell being cast then select an invalid spell, which
                            // is the same thing as not casting any spell
                            if spell.Name == currentSpell.Name {
                                shutdown(Spell{}, true)
                            } else {
                                shutdown(spell, true)
                            }

                            ui.RemoveGroup(group)
                        }
                        message := fmt.Sprintf("Do you wish to abort your %v spell?", currentSpell.Name)
                        group.AddElements(uilib.MakeConfirmDialogWithLayer(group, cache, &imageCache, 2, message, true, confirm, func(){
                            ui.RemoveGroup(group)
                        }))
                        ui.AddGroup(group)
                    } else {
                        // log.Printf("Click on spell %v", spell)
                        shutdown(spell, true)
                    }
                },
                RightClick: func(element *uilib.UIElement){
                    helpEntries := help.GetEntriesByName(spell.Name)
                    if helpEntries != nil {
                        group.AddElement(uilib.MakeHelpElementWithLayer(group, cache, &imageCache, 2, helpEntries[0], helpEntries[1:]...))
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

    if *currentPage >= len(spellPages) - len(spellPages) % 2 {
        *currentPage = len(spellPages) - len(spellPages) % 2
    }

    if currentSpell.Valid() {
        loop:
        for page, halfPage := range spellPages {
            for _, spell := range halfPage.Spells.Spells {
                if spell.Name == currentSpell.Name {
                    *currentPage = page
                    // force it to be even
                    *currentPage -= *currentPage % 2
                    break loop
                }
            }
        }
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
        // FIXME: for some spells a confirmation dialog box will ask if you really want to cast the spell
        // since there may not be any valid targets for the spell. Disjunction/Disjunction True are like that
        // when there are no global enchantments

        shutdownFinal := func(){
            getAlpha = ui.MakeFadeOut(7)
            ui.AddDelay(7, func(){
                setupSpells(-1)
                ui.RemoveElements(elements)
                chosenCallback(spell, picked)
            })

        }

        if picked && spell.IsVariableCost() {
            var powerGroup *uilib.UIElementGroup
            // on the overland the user might take multiple turns to cast the spell, but the cost can be anything
            extraStrength := spell.Cost(overland) * 4
            if !overland {
                // in combat, the cost of the spell cannot exceed the casting skill
                // maximum additional strength is whatever the casting skill is minus the cost of the spell
                // extraStrength = min(spell.Cost(overland) * 4, castingSkill - spell.Cost(overland))

                // the extra strength that can be put into variable spells is based on the final casting skill
                // of the caster that has to take the cost modifiers into account
                for extraStrength > 0 {
                    spell.OverrideCost = spell.BaseCost(overland) + extraStrength
                    if caster.ComputeEffectiveSpellCost(spell, overland) <= castingSkill {
                        break
                    }

                    extraStrength -= 1
                }

                spell.OverrideCost = 0
            }

            if extraStrength > 0 {
                powerGroup = makeAdditionalPowerElements(cache, &imageCache, extraStrength, func(amount int){
                    // modifiers to the cost will be applied later
                    spell.OverrideCost = spell.Cost(overland) + amount
                    ui.RemoveGroup(powerGroup)
                    shutdownFinal()
                })
                ui.AddGroup(powerGroup)
            } else {
                shutdownFinal()
            }
        } else {
            shutdownFinal()
        }
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
            vector.FillRect(screen, 0, 0, float32(screen.Bounds().Dx()), float32(screen.Bounds().Dy()), color.RGBA{R: 0, G: 0, B: 0, A: 128}, false)

            background, _ := imageCache.GetImage("spells.lbx", 0, 0)
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            options.GeoM.Translate(10, 10)
            scale.DrawScaled(screen, background, &options)

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
                    scale.DrawScaled(screen, bookFlip[index], &flipOptions)

                    if index == 0 {
                        if pageSideLeft >= 0 {
                            leftSide := getPageImage(pageSideLeft)
                            util.DrawDistortion(screen, bookFlip[index], leftSide, CastLeftSideDistortions1(bookFlip[index]), scale.ScaleGeom(flipOptions.GeoM))
                        }
                    } else if index == 1 {
                        if pageSideLeft >= 0 {
                            leftSide := getPageImage(pageSideLeft)
                            util.DrawDistortion(screen, bookFlip[index], leftSide, CastLeftSideDistortions2(bookFlip[index]), scale.ScaleGeom(flipOptions.GeoM))
                        }
                    } else if index == 2 {
                        if pageSideRight < len(spellPages) {
                            rightSide := getPageImage(pageSideRight)
                            util.DrawDistortion(screen, bookFlip[index], rightSide, CastRightSideDistortions1(bookFlip[index]), scale.ScaleGeom(flipOptions.GeoM))
                        }
                    } else if index == 3 {
                        if pageSideRight < len(spellPages) {
                            rightSide := getPageImage(pageSideRight)
                            util.DrawDistortion(screen, bookFlip[index], rightSide, CastRightSideDistortions2(bookFlip[index]), scale.ScaleGeom(flipOptions.GeoM))
                        }
                    }
                }

            } else {
                options.GeoM.Translate(15, 5)
                if *currentPage < len(spellPages) {
                    renderPage(screen, options, spellPages[*currentPage], highlightedSpell)
                }

                if *currentPage + 1 < len(spellPages) {
                    options.GeoM.Translate(134, 0)
                    // screen.DrawImage(right, &options)
                    renderPage(screen, options, spellPages[*currentPage+1], highlightedSpell)
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
        setupSpells(*currentPage)
    })

    pageTurnRight, _ := imageCache.GetImage("spells.lbx", 2, 0)
    pageTurnRightRect := image.Rect(0, 0, pageTurnRight.Bounds().Dx(), pageTurnRight.Bounds().Dy()).Add(image.Pt(268, 14))
    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Order: 1,
        Rect: pageTurnRightRect,
        LeftClick: func(this *uilib.UIElement){
            if *currentPage + 2 < len(spellPages) && !flipping {
                flipping = true
                bookFlipReverse = false
                bookFlipIndex = ui.Counter

                showPageRight = *currentPage + 3
                pageSideLeft = *currentPage + 1
                pageSideRight = *currentPage + 2
                showPageLeft = *currentPage

                ui.AddDelay(bookFlipSpeed * uint64(len(bookFlip)), func (){
                    flipping = false
                    *currentPage += 2
                    setupSpells(*currentPage)
                })
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            if *currentPage + 2 < len(spellPages) {
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(getAlpha())
                options.GeoM.Translate(float64(pageTurnRightRect.Min.X), float64(pageTurnRightRect.Min.Y))
                scale.DrawScaled(screen, pageTurnRight, &options)
            }
        },
    })

    pageTurnLeft, _ := imageCache.GetImage("spells.lbx", 1, 0)
    pageTurnLeftRect := image.Rect(0, 0, pageTurnLeft.Bounds().Dx(), pageTurnLeft.Bounds().Dy()).Add(image.Pt(23, 14))
    elements = append(elements, &uilib.UIElement{
        Rect: pageTurnLeftRect,
        Layer: 1,
        Order: 1,
        LeftClick: func(this *uilib.UIElement){
            if *currentPage >= 2 && !flipping {
                flipping = true
                bookFlipReverse = true
                bookFlipIndex = ui.Counter

                showPageRight = *currentPage + 1
                showPageLeft = *currentPage - 2
                pageSideLeft = *currentPage - 1
                pageSideRight = *currentPage

                ui.AddDelay(bookFlipSpeed * uint64(len(bookFlip) - 1), func (){
                    flipping = false
                    *currentPage -= 2
                    setupSpells(*currentPage)
                })
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            if *currentPage > 0 {
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(getAlpha())
                options.GeoM.Translate(float64(pageTurnLeftRect.Min.X), float64(pageTurnLeftRect.Min.Y))
                scale.DrawScaled(screen, pageTurnLeft, &options)
            }

        },
    })

    filterMagic := set.NewSet[data.MagicType]()

    type filter struct {
        Magic data.MagicType
        LbxIndex int
    }

    filters := []filter{
        {Magic: data.NatureMagic, LbxIndex: 4},
        {Magic: data.SorceryMagic, LbxIndex: 5},
        {Magic: data.ChaosMagic, LbxIndex: 6},
        {Magic: data.LifeMagic, LbxIndex: 7},
        {Magic: data.DeathMagic, LbxIndex: 8},
        {Magic: data.ArcaneMagic, LbxIndex: 9},
    }

    hasMagic := set.NewSet[data.MagicType]()
    for _, spell := range useSpells.Spells {
        hasMagic.Insert(spell.Magic)
    }

    var filterButtons []*uilib.UIElement
    filterX := 130
    for _, filter := range filters {
        if !hasMagic.Contains(filter.Magic) {
            continue
        }
        // filter spells by their realm
        pic, _ := imageCache.GetImage("spells.lbx", filter.LbxIndex, 0)
        rect := util.ImageRect(filterX, 10, pic)
        filterX += pic.Bounds().Dx() + 3
        selected := false
        filterButtons = append(filterButtons, &uilib.UIElement{
            Rect: rect,
            Layer: 1,
            Order: 1,
            LeftClick: func(this *uilib.UIElement){
                selected = !selected

                if selected {
                    filterMagic.Insert(filter.Magic)
                } else {
                    filterMagic.Remove(filter.Magic)
                }

                updateUseSpells(filterMagic)
                spellPages = computeHalfPages(useSpells, 6)
                pageCache = make(map[int]*ebiten.Image)
                if *currentPage >= len(spellPages) {
                    *currentPage = max(0, len(spellPages) - 1)
                    *currentPage -= *currentPage % 2
                }

                setupSpells(*currentPage)
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(getAlpha())
                if !selected {
                    options.ColorScale.ScaleWithColor(color.RGBA{R: 164, G: 164, B: 164, A: 255})
                } else {
                    options.ColorScale.SetR(1.2)
                    options.ColorScale.SetG(1.2)
                    options.ColorScale.SetB(1.2)
                }

                vector.FillRect(screen, scale.Scale(float32(rect.Min.X-1)), scale.Scale(float32(rect.Min.Y-1)), scale.Scale(float32(rect.Dx()+2)), scale.Scale(float32(rect.Dy()+2)), color.RGBA{R: 32, G: 32, B: 32, A: 128}, true)
                if selected {
                    vector.StrokeRect(screen, scale.Scale(float32(rect.Min.X-1)), scale.Scale(float32(rect.Min.Y-1)), scale.Scale(float32(rect.Dx()+2)), scale.Scale(float32(rect.Dy()+2)), 1, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
                }

                options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
                scale.DrawScaled(screen, pic, &options)
            },
        })
    }

    if len(filterButtons) > 1 {
        elements = append(elements, filterButtons...)
    }

    return elements
}
