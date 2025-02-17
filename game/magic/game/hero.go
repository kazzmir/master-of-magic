package game

import (
    "log"
    "fmt"
    "image"
    "image/color"
    "math"
    "slices"
    "cmp"

    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/unitview"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

func MakeHireHeroScreenUI(cache *lbx.LbxCache, ui *uilib.UI, hero *herolib.Hero, goldToHire int, action func(bool), onFadeOut func()) []*uilib.UIElement {
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

    yTop := float64(10 * data.ScreenScale)

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

    okDismissFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowGradient)

    var elements []*uilib.UIElement

    const fadeSpeed = 7

    getAlpha := ui.MakeFadeIn(fadeSpeed)

    hireText := "Hire"
    titleText := fmt.Sprintf("Hero for Hire: %v gold", goldToHire)
    if hero.HeroType == herolib.HeroTorin {
        hireText = "Accept"
        titleText = "Hero Summoned"
    }

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("unitview.lbx", 1, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(float64(31 * data.ScreenScale), float64(6 * data.ScreenScale))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(background, &options)

            options.GeoM.Translate(float64(9 * data.ScreenScale), float64(7 * data.ScreenScale))
            portraitLbx, portraitIndex := hero.GetPortraitLbxInfo()
            portrait, err := imageCache.GetImage(portraitLbx, portraitIndex, 0)
            if err == nil {
                screen.DrawImage(portrait, &options)
            }

            // unitview.RenderCombatImage(screen, &imageCache, &hero.Unit.Unit, options)

            options.GeoM.Reset()
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(float64(31 * data.ScreenScale), float64(6 * data.ScreenScale))
            options.GeoM.Translate(float64(51 * data.ScreenScale), float64(7 * data.ScreenScale))

            unitview.RenderUnitInfoNormal(screen, &imageCache, hero, hero.GetTitle(), "", descriptionFont, smallFont, options)

            options.GeoM.Reset()
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(float64(31 * data.ScreenScale), float64(6 * data.ScreenScale))
            options.GeoM.Translate(float64(10 * data.ScreenScale), float64(50 * data.ScreenScale))
            unitview.RenderUnitInfoStats(screen, &imageCache, hero, 15, descriptionFont, smallFont, options)

            /*
            options.GeoM.Translate(0, 60)
            unitview.RenderUnitAbilities(screen, &imageCache, hero, mediumFont, options, true, 0)
            */
        },
    })

    elements = append(elements, unitview.MakeUnitAbilitiesElements(ui, &imageCache, hero, mediumFont, 40 * data.ScreenScale, 124 * data.ScreenScale, &ui.Counter, 1, &getAlpha, true)...)

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            box, _ := imageCache.GetImage("unitview.lbx", 2, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(float64(248 * data.ScreenScale), float64(139 * data.ScreenScale))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(box, &options)
        },
    })

    buttonBackgrounds, _ := imageCache.GetImages("backgrnd.lbx", 24)
    // hire button
    hireRect := util.ImageRect(257 * data.ScreenScale, 149 * data.ScreenScale + int(yTop), buttonBackgrounds[0])
    hireIndex := 0
    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Rect: hireRect,
        LeftClick: func(this *uilib.UIElement){
            hireIndex = 1

            /*
            var confirmElements []*uilib.UIElement

            yes := func(){
                ui.RemoveElements(elements)
                // FIXME: disband unit
            }

            no := func(){
            }

            confirmElements = uilib.MakeConfirmDialogWithLayer(ui, cache, &imageCache, 2, fmt.Sprintf("Do you wish to disband the unit of %v?", unit.Unit.Name), yes, no)

            ui.AddElements(confirmElements)
            */
        },
        LeftClickRelease: func(this *uilib.UIElement){
            action(true)
            hireIndex = 0
            getAlpha = ui.MakeFadeOut(fadeSpeed)
            ui.AddDelay(fadeSpeed, func(){
                ui.RemoveElements(elements)
                onFadeOut()
            })
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(hireRect.Min.X), float64(hireRect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(buttonBackgrounds[hireIndex], &options)

            x := float64(hireRect.Min.X + hireRect.Max.X) / 2
            y := float64(hireRect.Min.Y + hireRect.Max.Y) / 2
            okDismissFont.PrintCenter(screen, x, y - float64(5 * data.ScreenScale), float64(data.ScreenScale), options.ColorScale, hireText)
        },
    })

    rejectRect := util.ImageRect(257 * data.ScreenScale, 169 * data.ScreenScale + int(yTop), buttonBackgrounds[0])
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
            screen.DrawImage(buttonBackgrounds[rejectIndex], &options)

            x := float64(rejectRect.Min.X + rejectRect.Max.X) / 2
            y := float64(rejectRect.Min.Y + rejectRect.Max.Y) / 2
            okDismissFont.PrintCenter(screen, x, y - float64(5 * data.ScreenScale), float64(data.ScreenScale), options.ColorScale, "Reject")
        },
    })

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            banner, _ := imageCache.GetImage("hire.lbx", 0, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(0, 0)
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(banner, &options)

            okDismissFont.PrintCenter(screen, float64(135 * data.ScreenScale), float64(6 * data.ScreenScale), float64(1 * data.ScreenScale), options.ColorScale, titleText)
        },
    })

    return elements
}

func (game *Game) showHeroLevelUpPopup(yield coroutine.YieldFunc, hero *herolib.Hero) {
    fontLbx, err := game.Cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return
    }

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

    top := float64(40 * data.ScreenScale)
    left := float64(30 * data.ScreenScale)

    // the set of abilities that can possibly show an improvement
    progressAbilities := set.MakeSet[data.AbilityType]()
    for _, ability := range []data.AbilityType{
        data.AbilityConstitution, data.AbilitySuperConstitution,
        data.AbilityAgility, data.AbilitySuperAgility,
        data.AbilityLeadership, data.AbilitySuperLeadership,
        data.AbilitySage, data.AbilitySuperSage,
        data.AbilityPrayermaster, data.AbilitySuperPrayermaster,
        data.AbilityArcanePower, data.AbilitySuperArcanePower,
        data.AbilityMight, data.AbilitySuperMight,
        data.AbilityArmsmaster, data.AbilitySuperArmsmaster,
        data.AbilityBlademaster, data.AbilitySuperBlademaster,
        data.AbilityLegendary, data.AbilitySuperLegendary,
    } {
        progressAbilities.Insert(ability)
    }

    var haveAbilities []data.Ability
    for _, ability := range hero.GetAbilities() {
        if progressAbilities.Contains(ability.Ability) {
            haveAbilities = append(haveAbilities, ability)
        }
    }

    slices.SortFunc(haveAbilities, func(a, b data.Ability) int {
        return cmp.Compare(a.Name(), b.Name())
    })

    maxAbilitiesPerRow := 2

    abilityRows := int(math.Ceil(float64(1 + len(haveAbilities)) / float64(maxAbilitiesPerRow)))

    height := (50 + abilityRows * 20) * data.ScreenScale

    titleFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowGradient)
    smallFont := font.MakeOptimizedFontWithPalette(fonts[2], yellowGradient)

    backgroundTop, _ := game.ImageCache.GetImage("reload.lbx", 23, 0)
    backgroundTop = backgroundTop.SubImage(image.Rect(0, 0, backgroundTop.Bounds().Dx(), height)).(*ebiten.Image)

    backgroundBottom, _ := game.ImageCache.GetImage("reload.lbx", 24, 0)

    portraitLbx, portraitIndex := hero.GetPortraitLbxInfo()
    portrait, err := game.ImageCache.GetImage(portraitLbx, portraitIndex, 0)

    dot, _ := game.ImageCache.GetImage("itemisc.lbx", 26, 0)

    drawer := game.Drawer
    defer func(){
        game.Drawer = drawer
    }()

    getAlpha := util.MakeFadeIn(7, &game.Counter)

    game.Drawer = func (screen *ebiten.Image, game *Game){
        drawer(screen, game)

        var options ebiten.DrawImageOptions

        // background
        options.GeoM.Translate(left, top)
        options.ColorScale.ScaleAlpha(getAlpha())
        screen.DrawImage(backgroundTop, &options)

        options.GeoM.Reset()
        options.GeoM.Translate(left, top + float64(height))
        screen.DrawImage(backgroundBottom, &options)

        // portrait
        options.GeoM.Reset()
        options.GeoM.Translate(left + float64(10 * data.ScreenScale), top + float64(10 * data.ScreenScale))
        screen.DrawImage(portrait, &options)

        // text
        titleFont.Print(screen, left + float64(48 * data.ScreenScale), top + float64(10 * data.ScreenScale), float64(data.ScreenScale), options.ColorScale, fmt.Sprintf("%v has made a level.", hero.Name))

        // stats progression
        for index, progression := range hero.GetBaseProgression() {
            xOffset := 95 * float64(index / 2)
            yOffset := 10 * float64(index % 2)

            options.GeoM.Reset()
            options.GeoM.Translate(left + (48 + xOffset) * float64(data.ScreenScale), top + (25 + yOffset) * float64(data.ScreenScale))
            screen.DrawImage(dot, &options)

            smallFont.Print(screen, left + (55 + xOffset) * float64(data.ScreenScale), top + (24 + yOffset) * float64(data.ScreenScale), float64(data.ScreenScale), options.ColorScale, progression)
        }

        row := 0
        column := 0
        abilityWidth := 115

        // level
        options.GeoM.Reset()
        options.GeoM.Translate(left + float64((10 + abilityWidth * column) * data.ScreenScale), top + float64((50 + row * 20) * data.ScreenScale))
        unitview.RenderExperienceBadge(screen, &game.ImageCache, hero, smallFont, options, false)

        // start in second column because the badge is in the first
        column = 1

        for _, ability := range haveAbilities {

            pic, err := game.ImageCache.GetImage(ability.LbxFile(), ability.LbxIndex(), 0)
            if err == nil {
                options.GeoM.Reset()
                options.GeoM.Translate(left + float64((10 + abilityWidth * column) * data.ScreenScale), top + float64((50 + row * 20) * data.ScreenScale))
                screen.DrawImage(pic, &options)

                x, y := options.GeoM.Apply(float64(pic.Bounds().Dx() + 2 * data.ScreenScale), float64(5 * data.ScreenScale))

                abilityBonus := hero.GetAbilityBonus(ability.Ability)
                if abilityBonus > 0 {
                    smallFont.Print(screen, x, y, float64(data.ScreenScale), options.ColorScale, fmt.Sprintf("%v +%v", ability.Name(), abilityBonus))
                } else {
                    smallFont.Print(screen, x, y, float64(data.ScreenScale), options.ColorScale, ability.Name())
                }
            }

            column += 1
            if column >= maxAbilitiesPerRow {
                row += 1
                column = 0
            }
        }
    }

    quit := false

    // absorb clicks and key presses
    yield()

    // fade in
    getAlpha = util.MakeFadeIn(7, &game.Counter)
    for i := 0; i < 7; i++ {
        game.Counter += 1
        yield()
    }


    for !quit {
        game.Counter += 1

        if inputmanager.LeftClick() || inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
            quit = true
        }

        yield()
    }

    // fade out
    getAlpha = util.MakeFadeOut(7, &game.Counter)
    for i := 0; i < 7; i++ {
        game.Counter += 1
        yield()
    }
}
