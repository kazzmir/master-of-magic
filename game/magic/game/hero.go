package game

import (
    "fmt"
    "image"
    "log"
    "math"
    "slices"
    "cmp"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/unitview"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type HireHeroFonts struct {
    DescriptionFont *font.Font
    SmallFont *font.Font
    MediumFont *font.Font
    OkDismissFont *font.Font
}

func MakeHireHeroFonts(cache *lbx.LbxCache) *HireHeroFonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Unable to load fonts: %v", err)
        return nil
    }

    return &HireHeroFonts{
        DescriptionFont: loader(fontslib.WhiteBig),
        SmallFont: loader(fontslib.SmallWhite),
        MediumFont: loader(fontslib.MediumWhite2),
        OkDismissFont: loader(fontslib.LightFont),
    }
}

func MakeHireHeroScreenUI(cache *lbx.LbxCache, ui *uilib.UI, hero *herolib.Hero, goldToHire int, action func(bool), onFadeOut func()) *uilib.UIElementGroup {
    imageCache := util.MakeImageCache(cache)

    yTop := float64(10)

    fonts := MakeHireHeroFonts(cache)

    const fadeSpeed = 7

    getAlpha := ui.MakeFadeIn(fadeSpeed)

    hireText := "Hire"
    titleText := fmt.Sprintf("Hero for Hire: %v gold", goldToHire)
    if hero.HeroType == herolib.HeroTorin {
        hireText = "Accept"
        titleText = "Hero Summoned"
    }

    uiGroup := uilib.MakeGroup()

    background, _ := imageCache.GetImage("unitview.lbx", 1, 0)
    uiGroup.AddElement(&uilib.UIElement{
        Layer: 1,
        Order: 0,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(float64(31), float64(6))
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, background, &options)

            options.GeoM.Translate(float64(9), float64(7))
            portraitLbx, portraitIndex := hero.GetPortraitLbxInfo()
            portrait, err := imageCache.GetImage(portraitLbx, portraitIndex, 0)
            if err == nil {
                scale.DrawScaled(screen, portrait, &options)
            }

            // unitview.RenderCombatImage(screen, &imageCache, &hero.Unit.Unit, options)

            options.GeoM.Reset()
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(float64(31), float64(6))
            options.GeoM.Translate(float64(51), float64(7))

            unitview.RenderUnitInfoNormal(screen, &imageCache, hero, hero.GetTitle(), "", fonts.DescriptionFont, fonts.SmallFont, options)

            /*
            options.GeoM.Reset()
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(float64(31), float64(6))
            options.GeoM.Translate(float64(10), float64(50))
            unitview.RenderUnitInfoStats(screen, &imageCache, hero, 15, fonts.DescriptionFont, fonts.SmallFont, options)
            */

            /*
            options.GeoM.Translate(0, 60)
            unitview.RenderUnitAbilities(screen, &imageCache, hero, mediumFont, options, true, 0)
            */
        },
    })

    /*
    uiGroup.AddElement(&uilib.UIElement{
        Layer: 1,
        Order: 1,
        Tooltip: func(element *uilib.UIElement) (string, *font.Font) {
            return "3", fonts.DescriptionFont
        },
        Rect: image.Rect(31, int(yTop) + 6 + 50, 31 + 150, int(yTop) + 6 + 50 + 50),
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(float64(31), float64(6))
            options.GeoM.Translate(float64(10), float64(50))
            options.ColorScale.ScaleAlpha(getAlpha())
            unitview.RenderUnitInfoStats(screen, &imageCache, hero, 15, fonts.DescriptionFont, fonts.SmallFont, options)
        },
    })
    */

    var statsOptions ebiten.DrawImageOptions
    statsOptions.GeoM.Translate(0, yTop)
    statsOptions.GeoM.Translate(float64(31), float64(6))
    statsOptions.GeoM.Translate(float64(10), float64(50))

    uiGroup.AddElements(unitview.CreateUnitInfoStatsElements(&imageCache, hero, 15, fonts.DescriptionFont, fonts.SmallFont, statsOptions, &getAlpha, background, 1))

    uiGroup.AddElements(unitview.MakeUnitAbilitiesElements(uiGroup, cache, &imageCache, hero, fonts.MediumFont, 40, 124, &ui.Counter, 1, &getAlpha, true, 0, false))

    uiGroup.AddElement(&uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            box, _ := imageCache.GetImage("unitview.lbx", 2, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(float64(248), float64(139))
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, box, &options)
        },
    })

    buttonBackgrounds, _ := imageCache.GetImages("backgrnd.lbx", 24)
    // hire button
    hireRect := util.ImageRect(257, 149 + int(yTop), buttonBackgrounds[0])
    hireIndex := 0
    uiGroup.AddElement(&uilib.UIElement{
        Layer: 1,
        Order: 1,
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
                ui.RemoveGroup(uiGroup)
                onFadeOut()
            })
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(hireRect.Min.X), float64(hireRect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, buttonBackgrounds[hireIndex], &options)

            x := float64(hireRect.Min.X + hireRect.Max.X) / 2
            y := float64(hireRect.Min.Y + hireRect.Max.Y) / 2
            fonts.OkDismissFont.PrintOptions(screen, x, y - float64(5), font.FontOptions{Justify: font.FontJustifyCenter, Options: &options, Scale: scale.ScaleAmount}, hireText)
        },
    })

    rejectRect := util.ImageRect(257, 169 + int(yTop), buttonBackgrounds[0])
    rejectIndex := 0
    uiGroup.AddElement(&uilib.UIElement{
        Layer: 1,
        Order: 1,
        Rect: rejectRect,
        LeftClick: func(this *uilib.UIElement){
            rejectIndex = 1
        },
        LeftClickRelease: func(this *uilib.UIElement){
            getAlpha = ui.MakeFadeOut(fadeSpeed)

            ui.AddDelay(fadeSpeed, func(){
                ui.RemoveGroup(uiGroup)
                action(false)
                onFadeOut()
            })
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(rejectRect.Min.X), float64(rejectRect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, buttonBackgrounds[rejectIndex], &options)

            x := float64(rejectRect.Min.X + rejectRect.Max.X) / 2
            y := float64(rejectRect.Min.Y + rejectRect.Max.Y) / 2
            fonts.OkDismissFont.PrintOptions(screen, x, y - float64(5), font.FontOptions{Options: &options, Scale: scale.ScaleAmount, Justify: font.FontJustifyCenter}, "Reject")
        },
    })

    uiGroup.AddElement(&uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            banner, _ := imageCache.GetImage("hire.lbx", 0, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(0, 0)
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, banner, &options)

            fonts.OkDismissFont.PrintOptions(screen, float64(135), float64(6), font.FontOptions{Options: &options, Scale: scale.ScaleAmount, Justify: font.FontJustifyCenter}, titleText)
        },
    })

    return uiGroup
}

type HeroLevelUpFonts struct {
    TitleFont *font.Font
    SmallFont *font.Font
}

func MakeHeroLevelUpFonts(cache *lbx.LbxCache) *HeroLevelUpFonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Unable to load fonts: %v", err)
        return nil
    }

    return &HeroLevelUpFonts{
        TitleFont: loader(fontslib.LightFont),
        SmallFont: loader(fontslib.LightFontSmall),
    }
}

func (game *Game) showHeroLevelUpPopup(yield coroutine.YieldFunc, hero *herolib.Hero) {
    fonts := MakeHeroLevelUpFonts(game.Cache)

    top := float64(40)
    left := float64(30)

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

    height := (50 + abilityRows * 20)

    backgroundTop, _ := game.ImageCache.GetImage("reload.lbx", 23, 0)
    backgroundTop = backgroundTop.SubImage(image.Rect(0, 0, backgroundTop.Bounds().Dx(), height)).(*ebiten.Image)

    backgroundBottom, _ := game.ImageCache.GetImage("reload.lbx", 24, 0)

    portraitLbx, portraitIndex := hero.GetPortraitLbxInfo()
    portrait, _ := game.ImageCache.GetImage(portraitLbx, portraitIndex, 0)

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
        scale.DrawScaled(screen, backgroundTop, &options)

        options.GeoM.Reset()
        options.GeoM.Translate(left, top + float64(height))
        scale.DrawScaled(screen, backgroundBottom, &options)

        // portrait
        options.GeoM.Reset()
        options.GeoM.Translate(left + float64(10), top + float64(10))
        scale.DrawScaled(screen, portrait, &options)

        // text
        fonts.TitleFont.PrintOptions(screen, left + float64(48), top + float64(10), font.FontOptions{Options: &options, Scale: scale.ScaleAmount}, fmt.Sprintf("%v has made a level.", hero.Name))

        // stats progression
        for index, progression := range hero.GetBaseProgression() {
            xOffset := 95 * float64(index / 2)
            yOffset := 10 * float64(index % 2)

            options.GeoM.Reset()
            options.GeoM.Translate(left + (48 + xOffset), top + (25 + yOffset))
            scale.DrawScaled(screen, dot, &options)

            fonts.SmallFont.PrintOptions(screen, left + (55 + xOffset), top + (24 + yOffset), font.FontOptions{Options: &options, Scale: scale.ScaleAmount, DropShadow: true}, progression)
        }

        row := 0
        column := 0
        abilityWidth := 115

        // level
        options.GeoM.Reset()
        options.GeoM.Translate(left + float64((10 + abilityWidth * column)), top + float64((50 + row * 20)))
        unitview.RenderExperienceBadge(screen, &game.ImageCache, hero, fonts.SmallFont, options, false)

        // start in second column because the badge is in the first
        column = 1

        for _, ability := range haveAbilities {

            pic, err := game.ImageCache.GetImage(ability.LbxFile(), ability.LbxIndex(), 0)
            if err == nil {
                options.GeoM.Reset()
                options.GeoM.Translate(left + float64((10 + abilityWidth * column)), top + float64((50 + row * 20)))
                scale.DrawScaled(screen, pic, &options)

                x, y := options.GeoM.Apply(float64(pic.Bounds().Dx() + 2), float64(5))

                fonts.SmallFont.PrintOptions(screen, x, y, font.FontOptions{Options: &options, Scale: scale.ScaleAmount, DropShadow: true}, ability.Name())
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
