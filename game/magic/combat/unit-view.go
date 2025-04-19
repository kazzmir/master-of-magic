package combat

import (
    "log"
    "fmt"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
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
    loader, err := fontslib.Loader(cache)
    if err != nil {
        return nil, err
    }

    return &UnitViewFonts{
        DescriptionFont: loader(fontslib.WhiteBig),
        SmallFont: loader(fontslib.SmallWhite),
        MediumFont: loader(fontslib.MediumWhite2),
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

    var layer uilib.UILayer = 1

    group.AddElement(&uilib.UIElement{
        Layer: layer,
        Order: -1,
        NotLeftClicked: func(element *uilib.UIElement) {
            getAlpha = ui.MakeFadeOut(fadeSpeed)
            ui.AddDelay(fadeSpeed, func(){
                ui.RemoveGroup(group)
            })
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("unitview.lbx", 1, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(31), float64(6))
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, background, &options)

            options.GeoM.Translate(float64(25), float64(30))
            portaitUnit, ok := unit.Unit.(unitview.PortraitUnit)
            if ok {
                lbxFile, index := portaitUnit.GetPortraitLbxInfo()
                portait, err := imageCache.GetImage(lbxFile, index, 0)
                if err == nil {
                    options.GeoM.Translate(0, float64(-7))
                    options.GeoM.Translate(float64(-portait.Bounds().Dx()/2), float64(-portait.Bounds().Dy()/2))
                    scale.DrawScaled(screen, portait, &options)
                }
            } else {
                unitview.RenderUnitViewImage(screen, &imageCache, unit, options, unit.IsAsleep(), ui.Counter)
            }

            options.GeoM.Reset()
            options.GeoM.Translate(float64(31), float64(6))
            options.GeoM.Translate(float64(51), float64(6))

            RenderUnitInfo(screen, &imageCache, unit, fonts, options)

            /*
            options.GeoM.Reset()
            options.GeoM.Translate(float64(31), float64(6))
            options.GeoM.Translate(float64(10), float64(50))
            unitview.RenderUnitInfoStats(screen, &imageCache, unit, 15, fonts.DescriptionFont, fonts.SmallFont, options)
            */

            /*
            options.GeoM.Translate(0, 60)
            RenderUnitAbilities(screen, &imageCache, unit, mediumFont, options, false, 0)
            */
        },
    })

    var defaultOptions ebiten.DrawImageOptions
    defaultOptions.GeoM.Translate(float64(31), float64(6))
    defaultOptions.GeoM.Translate(float64(10), float64(50))

    group.AddElements(unitview.CreateUnitInfoStatsElements(&imageCache, unit, 15, fonts.DescriptionFont, fonts.SmallFont, defaultOptions, &getAlpha, layer))

    group.AddElements(unitview.MakeUnitAbilitiesElements(group, cache, &imageCache, unit, fonts.MediumFont, 40, 114, &ui.Counter, layer, &getAlpha, false, 0, false))

    return group
}

func RenderUnitInfo(screen *ebiten.Image, imageCache *util.ImageCache, unit *ArmyUnit, fonts *UnitViewFonts, defaultOptions ebiten.DrawImageOptions) {
    x, y := defaultOptions.GeoM.Apply(0, 0)

    // FIXME: if the unit is a hero and has a title then the title should show up on the next line after the name
    name := unit.Unit.GetFullName()

    fonts.DescriptionFont.PrintOptions(screen, x, y + float64(2), font.FontOptions{DropShadow: true, Options: &defaultOptions, Scale: scale.ScaleAmount}, name)

    y += float64((fonts.DescriptionFont.Height() + 6))

    fonts.SmallFont.PrintOptions(screen, x, y, font.FontOptions{DropShadow: true, Options: &defaultOptions, Scale: scale.ScaleAmount}, "Moves")

    unitMoves := unit.GetMovementSpeed()

    movementImage, err := imageCache.GetImage("unitview.lbx", 24, 0)
    if unit.IsFlying() {
        movementImage, _ = imageCache.GetImage("unitview.lbx", 25, 0)
    } else if unit.IsSwimmer() {
        movementImage, _ = imageCache.GetImage("unitview.lbx", 26, 0)
    }
    if err == nil {
        var options ebiten.DrawImageOptions
        options = defaultOptions
        options.GeoM.Reset()
        options.GeoM.Translate(x + fonts.SmallFont.MeasureTextWidth("Damage ", 1), y)

        // FIXME: draw half a movement image if the unit has a half movement point?
        for i := 0; i < unitMoves.ToInt(); i++ {
            scale.DrawScaled(screen, movementImage, &options)
            options.GeoM.Translate(float64(movementImage.Bounds().Dx()), 0)
        }
    }

    y += float64((fonts.SmallFont.Height() + 3))
    fonts.SmallFont.PrintOptions(screen, x, y, font.FontOptions{DropShadow: true, Options: &defaultOptions, Scale: scale.ScaleAmount}, fmt.Sprintf("Damage %v", unit.GetDamage()))
}
