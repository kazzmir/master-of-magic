package hero

import (
    "log"
    "fmt"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/unitview"
    "github.com/kazzmir/master-of-magic/game/magic/util"

    "github.com/hajimehoshi/ebiten/v2"
)

type Hero struct {
    Unit *units.OverworldUnit
    Title string
}

func getHeroPortraitIndex(hero *Hero) int {
    if hero.Unit.Unit.Equals(units.HeroTorin) {
        return 0
    }

    if hero.Unit.Unit.Equals(units.HeroFang) {
        return 1
    }

    if hero.Unit.Unit.Equals(units.HeroBShan) {
        return 2
    }

    if hero.Unit.Unit.Equals(units.HeroMorgana) {
        return 3
    }

    if hero.Unit.Unit.Equals(units.HeroWarrax) {
        return 4
    }

    if hero.Unit.Unit.Equals(units.HeroMysticX) {
        return 5
    }

    if hero.Unit.Unit.Equals(units.HeroBahgtru) {
        return 6
    }

    if hero.Unit.Unit.Equals(units.HeroDethStryke) {
        return 7
    }

    if hero.Unit.Unit.Equals(units.HeroSpyder) {
        return 8
    }

    if hero.Unit.Unit.Equals(units.HeroSirHarold) {
        return 9
    }

    if hero.Unit.Unit.Equals(units.HeroBrax) {
        return 10
    }

    if hero.Unit.Unit.Equals(units.HeroRavashack) {
        return 11
    }

    if hero.Unit.Unit.Equals(units.HeroGreyfairer) {
        return 12
    }

    if hero.Unit.Unit.Equals(units.HeroShalla) {
        return 13
    }

    if hero.Unit.Unit.Equals(units.HeroRoland) {
        return 14
    }

    if hero.Unit.Unit.Equals(units.HeroMalleus) {
        return 15
    }

    if hero.Unit.Unit.Equals(units.HeroMortu) {
        return 16
    }

    if hero.Unit.Unit.Equals(units.HeroGunther) {
        return 17
    }

    if hero.Unit.Unit.Equals(units.HeroRakir) {
        return 18
    }

    if hero.Unit.Unit.Equals(units.HeroJaer) {
        return 19
    }

    if hero.Unit.Unit.Equals(units.HeroTaki) {
        return 20
    }

    if hero.Unit.Unit.Equals(units.HeroYramrag) {
        return 21
    }

    if hero.Unit.Unit.Equals(units.HeroValana) {
        return 22
    }

    if hero.Unit.Unit.Equals(units.HeroElana) {
        return 23
    }

    if hero.Unit.Unit.Equals(units.HeroAerie) {
        return 24
    }

    if hero.Unit.Unit.Equals(units.HeroMarcus) {
        return 25
    }

    if hero.Unit.Unit.Equals(units.HeroReywind) {
        return 26
    }

    if hero.Unit.Unit.Equals(units.HeroAlorra) {
        return 27
    }

    if hero.Unit.Unit.Equals(units.HeroZaldron) {
        return 28
    }

    if hero.Unit.Unit.Equals(units.HeroShinBo) {
        return 29
    }

    if hero.Unit.Unit.Equals(units.HeroSerena) {
        return 30
    }

    if hero.Unit.Unit.Equals(units.HeroShuri) {
        return 31
    }

    if hero.Unit.Unit.Equals(units.HeroTheria) {
        return 32
    }

    if hero.Unit.Unit.Equals(units.HeroTumu) {
        return 33
    }

    if hero.Unit.Unit.Equals(units.HeroAureus) {
        return 34
    }

    return -1
}

func MakeHireScreenUI(cache *lbx.LbxCache, ui *uilib.UI, hero *Hero, action func(bool)) []*uilib.UIElement {
    goldToHire := 100

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

    yTop := float64(10)

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

    heroPortraitIndex := getHeroPortraitIndex(hero)

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("unitview.lbx", 1, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(31, 6)
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(background, &options)

            options.GeoM.Translate(9, 7)
            portrait, err := imageCache.GetImage("portrait.lbx", heroPortraitIndex, 0)
            if err == nil {
                screen.DrawImage(portrait, &options)
            }

            // unitview.RenderCombatImage(screen, &imageCache, &hero.Unit.Unit, options)

            options.GeoM.Reset()
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(31, 6)
            options.GeoM.Translate(51, 7)

            unitview.RenderUnitInfoNormal(screen, &imageCache, &hero.Unit.Unit, hero.Title, descriptionFont, smallFont, options)

            options.GeoM.Reset()
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(31, 6)
            options.GeoM.Translate(10, 50)
            unitview.RenderUnitInfoStats(screen, &imageCache, &hero.Unit.Unit, descriptionFont, smallFont, options)

            options.GeoM.Translate(0, 60)
            unitview.RenderUnitAbilities(screen, &imageCache, &hero.Unit.Unit, mediumFont, options)
        },
    })

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            box, _ := imageCache.GetImage("unitview.lbx", 2, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(0, yTop)
            options.GeoM.Translate(248, 139)
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(box, &options)
        },
    })

    buttonBackgrounds, _ := imageCache.GetImages("backgrnd.lbx", 24)
    // hire button
    hireRect := util.ImageRect(257, 149 + int(yTop), buttonBackgrounds[0])
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
            hireIndex = 0
            action(true)
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(hireRect.Min.X), float64(hireRect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(buttonBackgrounds[hireIndex], &options)

            x := float64(hireRect.Min.X + hireRect.Max.X) / 2
            y := float64(hireRect.Min.Y + hireRect.Max.Y) / 2
            okDismissFont.PrintCenter(screen, x, y - 5, 1, options.ColorScale, "Hire")
        },
    })

    rejectRect := util.ImageRect(257, 169 + int(yTop), buttonBackgrounds[0])
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
            okDismissFont.PrintCenter(screen, x, y - 5, 1, options.ColorScale, "Reject")
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

            okDismissFont.PrintCenter(screen, 135, 6, 1, options.ColorScale, fmt.Sprintf("Hero for Hire: %v gold", goldToHire))
        },
    })

    return elements
}
