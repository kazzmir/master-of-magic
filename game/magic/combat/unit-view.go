package combat

import (
    "log"
    "fmt"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/unitview"

    "github.com/hajimehoshi/ebiten/v2"
)

type UnitViewFonts struct {
    DescriptionFont *font.Font
    SmallFont *font.Font
    MediumFont *font.Font
    YellowGradient *font.Font
}

func MakeFonts(cache *lbx.LbxCache) (*UnitViewFonts, error) {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        return nil, err
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        return nil, err
    }

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

    return &UnitViewFonts{
        DescriptionFont: descriptionFont,
        SmallFont: smallFont,
        MediumFont: mediumFont,
    }, nil
}

func MakeUnitView(cache *lbx.LbxCache, ui *uilib.UI, unit *ArmyUnit) *uilib.UIElementGroup {
    fonts, err := MakeFonts(cache)
    if err != nil {
        log.Printf("Unable to make fonts: %v", err)
        return nil
    }

    imageCache := util.MakeImageCache(cache)

    group := uilib.MakeGroup()

    const fadeSpeed = 7
    getAlpha := ui.MakeFadeIn(fadeSpeed)

    group.AddElement(&uilib.UIElement{
        Layer: 1,
        NotLeftClicked: func(element *uilib.UIElement) {
            getAlpha = ui.MakeFadeOut(fadeSpeed)
            ui.AddDelay(fadeSpeed, func(){
                ui.RemoveGroup(group)
            })
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("unitview.lbx", 1, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(31 * data.ScreenScale), float64(6 * data.ScreenScale))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(background, &options)

            options.GeoM.Translate(float64(25 * data.ScreenScale), float64(30 * data.ScreenScale))
            portaitUnit, ok := unit.Unit.(unitview.PortraitUnit)
            if ok {
                lbxFile, index := portaitUnit.GetPortraitLbxInfo()
                portait, err := imageCache.GetImage(lbxFile, index, 0)
                if err == nil {
                    options.GeoM.Translate(0, float64(-7 * data.ScreenScale))
                    options.GeoM.Translate(float64(-portait.Bounds().Dx()/2), float64(-portait.Bounds().Dy()/2))
                    screen.DrawImage(portait, &options)
                }
            } else {
                unitview.RenderUnitViewImage(screen, &imageCache, unit, options, ui.Counter)
            }

            options.GeoM.Reset()
            options.GeoM.Translate(float64(31 * data.ScreenScale), float64(6 * data.ScreenScale))
            options.GeoM.Translate(float64(51 * data.ScreenScale), float64(6 * data.ScreenScale))

            RenderUnitInfo(screen, &imageCache, unit, fonts, options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(31 * data.ScreenScale), float64(6 * data.ScreenScale))
            options.GeoM.Translate(float64(10 * data.ScreenScale), float64(50 * data.ScreenScale))
            unitview.RenderUnitInfoStats(screen, &imageCache, unit, 15, fonts.DescriptionFont, fonts.SmallFont, options)

            /*
            options.GeoM.Translate(0, 60)
            RenderUnitAbilities(screen, &imageCache, unit, mediumFont, options, false, 0)
            */
        },
    })

    group.AddElements(unitview.MakeUnitAbilitiesElements(group, cache, &imageCache, unit, fonts.MediumFont, 40 * data.ScreenScale, 114 * data.ScreenScale, &ui.Counter, 1, &getAlpha, false, 0, false))

    return group
}

func RenderUnitInfo(screen *ebiten.Image, imageCache *util.ImageCache, unit *ArmyUnit, fonts *UnitViewFonts, defaultOptions ebiten.DrawImageOptions) {
    x, y := defaultOptions.GeoM.Apply(0, 0)

    name := unit.Unit.GetFullName()

    fonts.DescriptionFont.PrintOptions(screen, x, y + float64(2 * data.ScreenScale), float64(data.ScreenScale), defaultOptions.ColorScale, font.FontOptions{DropShadow: true}, name)

    y += float64((fonts.DescriptionFont.Height() + 6) * data.ScreenScale)

    fonts.SmallFont.PrintOptions(screen, x, y, float64(data.ScreenScale), defaultOptions.ColorScale, font.FontOptions{DropShadow: true}, "Moves")

    unitMoves := unit.GetMovementSpeed()

    // FIXME: show wings if flying, or the water thing if can walk on water
    smallBoot, err := imageCache.GetImage("unitview.lbx", 24, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options = defaultOptions
        options.GeoM.Reset()
        options.GeoM.Translate(x + fonts.SmallFont.MeasureTextWidth("Damage ", float64(data.ScreenScale)), y)

        for i := 0; i < unitMoves; i++ {
            screen.DrawImage(smallBoot, &options)
            options.GeoM.Translate(float64(smallBoot.Bounds().Dx()), 0)
        }
    }

    y += float64((fonts.SmallFont.Height() + 3) * data.ScreenScale)
    fonts.SmallFont.PrintOptions(screen, x, y, float64(data.ScreenScale), defaultOptions.ColorScale, font.FontOptions{DropShadow: true}, fmt.Sprintf("Damage %v", unit.GetDamage()))
}
