package unitview

import (
    "log"
    "fmt"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"

    "github.com/hajimehoshi/ebiten/v2"
)

type UnitView interface {
    GetName() string
    GetTitle() string // for heroes. normal units will not have a title
    GetBanner() data.BannerType
    GetCombatLbxFile() string
    GetCombatIndex(units.Facing) int
    GetLbxFile() string
    GetLbxIndex() int
    GetCount() int
    GetUpkeepGold() int
    GetUpkeepFood() int
    GetUpkeepMana() int
    GetMovementSpeed() int
    GetProductionCost() int
    GetEnchantments() []data.UnitEnchantment
    GetWeaponBonus() data.WeaponBonus
    GetExperience() int
    GetExperienceData() units.ExperienceData
    GetBaseMeleeAttackPower() int
    GetMeleeAttackPower() int
    GetBaseRangedAttackPower() int
    GetRangedAttackPower() int
    GetRangedAttackDamageType() units.Damage
    GetBaseDefense() int
    GetDefense() int
    GetResistance() int
    GetBaseResistance() int
    GetHitPoints() int
    GetBaseHitPoints() int
    GetAbilities() []data.Ability
    GetArtifactSlots() []artifact.ArtifactSlot
    GetArtifacts() []*artifact.Artifact
}

type PortraitUnit interface {
    // returns the lbx file and index that the portrait is in
    GetPortraitLbxInfo() (string, int)
}

func UnitDisbandMessage(unit UnitView) string {
    return fmt.Sprintf("Do you wish to disband the unit of %v?", unit.GetName())
}

func MakeUnitContextMenu(cache *lbx.LbxCache, ui *uilib.UI, unit UnitView, doDisband func()) []*uilib.UIElement {
    maybeHero, ok := unit.(*herolib.Hero)
    if ok {
        return makeHeroContextMenu(cache, ui, maybeHero, doDisband)
    }

    return MakeGenericContextMenu(cache, ui, unit, UnitDisbandMessage(unit), doDisband)
}

func makeHeroContextMenu(cache *lbx.LbxCache, ui *uilib.UI, hero *herolib.Hero, doDisband func()) []*uilib.UIElement {
    return MakeGenericContextMenu(cache, ui, hero, fmt.Sprintf("Do you wish to dismiss %v?", hero.GetName()), doDisband)
}

func MakeGenericContextMenu(cache *lbx.LbxCache, ui *uilib.UI, unit UnitView, disbandMessage string, doDisband func()) []*uilib.UIElement {
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

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("unitview.lbx", 1, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(31, 6)
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(background, &options)

            options.GeoM.Translate(25, 30)
            portaitUnit, ok := unit.(PortraitUnit)
            if ok {
                lbxFile, index := portaitUnit.GetPortraitLbxInfo()
                portait, err := imageCache.GetImage(lbxFile, index, 0)
                if err == nil {
                    options.GeoM.Translate(0, -7)
                    options.GeoM.Translate(float64(-portait.Bounds().Dx()/2), float64(-portait.Bounds().Dy()/2))
                    screen.DrawImage(portait, &options)
                }
            } else {
                RenderCombatImage(screen, &imageCache, unit, options, ui.Counter)
            }

            options.GeoM.Reset()
            options.GeoM.Translate(31, 6)
            options.GeoM.Translate(51, 6)

            RenderUnitInfoNormal(screen, &imageCache, unit, unit.GetTitle(), "", descriptionFont, smallFont, options)

            options.GeoM.Reset()
            options.GeoM.Translate(31, 6)
            options.GeoM.Translate(10, 50)
            RenderUnitInfoStats(screen, &imageCache, unit, 15, descriptionFont, smallFont, options)

            /*
            options.GeoM.Translate(0, 60)
            RenderUnitAbilities(screen, &imageCache, unit, mediumFont, options, false, 0)
            */
        },
    })

    elements = append(elements, MakeUnitAbilitiesElements(&imageCache, unit, mediumFont, 40, 114, &ui.Counter, 1, &getAlpha, false)...)

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            box, _ := imageCache.GetImage("unitview.lbx", 2, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(248, 139)
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(box, &options)
        },
    })

    buttonBackgrounds, _ := imageCache.GetImages("backgrnd.lbx", 24)
    // dismiss button
    cancelRect := util.ImageRect(257, 149, buttonBackgrounds[0])
    cancelIndex := 0
    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Rect: cancelRect,
        LeftClick: func(this *uilib.UIElement){
            cancelIndex = 1

            var confirmElements []*uilib.UIElement

            yes := func(){
                ui.RemoveElements(elements)
                doDisband()
            }

            no := func(){
            }

            confirmElements = uilib.MakeConfirmDialogWithLayer(ui, cache, &imageCache, 2, disbandMessage, yes, no)

            ui.AddElements(confirmElements)
        },
        LeftClickRelease: func(this *uilib.UIElement){
            cancelIndex = 0
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(cancelRect.Min.X), float64(cancelRect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(buttonBackgrounds[cancelIndex], &options)

            x := float64(cancelRect.Min.X + cancelRect.Max.X) / 2
            y := float64(cancelRect.Min.Y + cancelRect.Max.Y) / 2
            okDismissFont.PrintCenter(screen, x, y - 5, 1, options.ColorScale, "Dismiss")
        },
    })

    okRect := util.ImageRect(257, 169, buttonBackgrounds[0])
    okIndex := 0
    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Rect: okRect,
        LeftClick: func(this *uilib.UIElement){
            okIndex = 1
        },
        LeftClickRelease: func(this *uilib.UIElement){
            getAlpha = ui.MakeFadeOut(fadeSpeed)

            ui.AddDelay(fadeSpeed, func(){
                ui.RemoveElements(elements)
            })
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(okRect.Min.X), float64(okRect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(buttonBackgrounds[okIndex], &options)

            x := float64(okRect.Min.X + okRect.Max.X) / 2
            y := float64(okRect.Min.Y + okRect.Max.Y) / 2
            okDismissFont.PrintCenter(screen, x, y - 5, 1, options.ColorScale, "Ok")
        },
    })

    return elements
}

// FIXME: this was copied from combat/combat-screen.go
func makePaletteFromBanner(banner data.BannerType) color.Palette {
    var topColor color.RGBA

    switch banner {
        case data.BannerGreen: topColor = color.RGBA{R: 0x20, G: 0x80, B: 0x2c, A: 0xff}
        case data.BannerBlue: topColor = color.RGBA{R: 0x15, G: 0x1d, B: 0x9d, A: 0xff}
        case data.BannerRed: topColor = color.RGBA{R: 0x9d, G: 0x15, B: 0x15, A: 0xff}
        case data.BannerPurple: topColor = color.RGBA{R: 0x6d, G: 0x15, B: 0x9d, A: 0xff}
        case data.BannerYellow: topColor = color.RGBA{R: 0x9d, G: 0x9d, B: 0x15, A: 0xff}
        case data.BannerBrown: topColor = color.RGBA{R: 0x82, G: 0x60, B: 0x12, A: 0xff}
    }

    // red := color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}

    solidColor := util.Lighten(topColor, 80)
    return color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        solidColor, solidColor, solidColor, solidColor,
        solidColor, solidColor, solidColor, solidColor,
    }
}

// list of units that shows up when you right click on an enemy unit stack
func MakeSmallListView(cache *lbx.LbxCache, ui *uilib.UI, stack []UnitView, title string, clicked func()) []*uilib.UIElement {
    imageCache := util.MakeImageCache(cache)

    titleHeight := 22
    unitHeight := 19

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

    black := color.RGBA{R: 0, G: 0, B: 0, A: 0xff}
    descriptionPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        black, black, black, black,
        black, black, black, black,
    }

    brightPalette := color.Palette{
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

    titleFont := font.MakeOptimizedFontWithPalette(fonts[4], makePaletteFromBanner(stack[0].GetBanner()))
    smallFont := font.MakeOptimizedFontWithPalette(fonts[1], descriptionPalette)
    mediumFont := font.MakeOptimizedFontWithPalette(fonts[2], brightPalette)

    // title bar + 1 for each unit
    height := titleHeight + unitHeight * len(stack)

    fullBackground, _ := imageCache.GetImage("unitview.lbx", 28, 0)

    background := fullBackground.SubImage(image.Rect(0, 0, fullBackground.Bounds().Dx(), height)).(*ebiten.Image)
    bottom, _ := imageCache.GetImage("unitview.lbx", 29, 0)

    posX := 30
    posY := data.ScreenHeight / 2 - background.Bounds().Dy() / 2

    var elements []*uilib.UIElement

    getAlpha := ui.MakeFadeIn(7)

    // cut the border off
    cut1PixelFunc := func (input *image.Paletted) image.Image {
        bounds := input.Bounds()
        return input.SubImage(image.Rect(bounds.Min.X+1, bounds.Min.Y+1, bounds.Max.X-1, bounds.Max.Y-1))
    }

    meleeImage, _ := imageCache.GetImageTransform("unitview.lbx", 13, 0, "cut1", cut1PixelFunc)
    rangeMagicImage, _ := imageCache.GetImageTransform("unitview.lbx", 14, 0, "cut1", cut1PixelFunc)
    rangeBowImage, _ := imageCache.GetImageTransform("unitview.lbx", 18, 0, "cut1", cut1PixelFunc)
    rangeBoulderImage, _ := imageCache.GetImageTransform("unitview.lbx", 19, 0, "cut1", cut1PixelFunc)
    defenseImage, _ := imageCache.GetImageTransform("unitview.lbx", 22, 0, "cut1", cut1PixelFunc)
    healthImage, _ := imageCache.GetImageTransform("unitview.lbx", 23, 0, "cut1", cut1PixelFunc)
    moveImage, _ := imageCache.GetImageTransform("unitview.lbx", 24, 0, "cut1", cut1PixelFunc)

    rect := util.ImageRect(posX, posY, background)
    element := &uilib.UIElement{
        Rect: rect,
        LeftClick: func(this *uilib.UIElement){
            getAlpha = ui.MakeFadeOut(7)
            ui.AddDelay(7, func(){
                ui.RemoveElements(elements)
                clicked()
            })
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(posX), float64(posY))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(background, &options)

            titleX, titleY := options.GeoM.Apply(float64(background.Bounds().Dx() / 2), 8)
            titleFont.PrintCenter(screen, titleX, titleY, 1, options.ColorScale, title)

            /*
            util.DrawRect(screen, image.Rect(posX, posY, posX+1, posY + titleHeight), color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff})
            util.DrawRect(screen, image.Rect(posX, posY + titleHeight, posX+1, posY + titleHeight + unitHeight), color.RGBA{R: 0, G: 0xff, B: 0, A: 0xff})
            */

            options.GeoM.Translate(0, float64(background.Bounds().Dy()))
            screen.DrawImage(bottom, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(posX), float64(posY + titleHeight))

            var unitOptions ebiten.DrawImageOptions
            for _, unit := range stack {
                banner := unit.GetBanner()
                unitBack, err := units.GetUnitBackgroundImage(unit.GetBanner(), &imageCache)
                if err != nil {
                    continue
                }

                unitImage, err := imageCache.GetImageTransform(unit.GetLbxFile(), unit.GetLbxIndex(), 0, banner.String(), units.MakeUpdateUnitColorsFunc(banner))
                if err != nil {
                    continue
                }

                var x, y float64

                unitOptions = options
                unitOptions.GeoM.Translate(8, 2)
                screen.DrawImage(unitBack, &unitOptions)
                unitOptions.GeoM.Translate(1, 1)
                screen.DrawImage(unitImage, &unitOptions)

                x, y = unitOptions.GeoM.Apply(float64(unitBack.Bounds().Dx() + 2), 5)
                mediumFont.Print(screen, x, y, 1, options.ColorScale, unit.GetName())

                unitOptions.GeoM.Translate(133, 5)
                x, y = unitOptions.GeoM.Apply(0, 1)
                smallFont.PrintRight(screen, x, y, 1, options.ColorScale, fmt.Sprintf("%v", unit.GetMeleeAttackPower()))
                // FIXME: show mythril/adamantium weapons?
                screen.DrawImage(meleeImage, &unitOptions)

                unitOptions.GeoM.Translate(20, 0)
                x, y = unitOptions.GeoM.Apply(0, 1)
                smallFont.PrintRight(screen, x, y, 1, options.ColorScale, fmt.Sprintf("%v", unit.GetRangedAttackPower()))
                switch unit.GetRangedAttackDamageType() {
                    case units.DamageNone: // nothing
                    case units.DamageRangedMagical:
                        screen.DrawImage(rangeMagicImage, &unitOptions)
                    case units.DamageRangedPhysical:
                        screen.DrawImage(rangeBowImage, &unitOptions)
                    case units.DamageRangedBoulder:
                        screen.DrawImage(rangeBoulderImage, &unitOptions)
                }

                unitOptions.GeoM.Translate(20, 0)
                x, y = unitOptions.GeoM.Apply(0, 1)
                smallFont.PrintRight(screen, x, y, 1, options.ColorScale, fmt.Sprintf("%v", unit.GetDefense()))
                screen.DrawImage(defenseImage, &unitOptions)

                unitOptions.GeoM.Translate(20, 0)
                x, y = unitOptions.GeoM.Apply(0, 1)
                smallFont.PrintRight(screen, x, y, 1, options.ColorScale, fmt.Sprintf("%v", unit.GetHitPoints()))

                screen.DrawImage(healthImage, &unitOptions)

                unitOptions.GeoM.Translate(20, 0)
                x, y = unitOptions.GeoM.Apply(0, 1)
                smallFont.PrintRight(screen, x, y, 1, options.ColorScale, fmt.Sprintf("%v", unit.GetMovementSpeed()))

                screen.DrawImage(moveImage, &unitOptions)

                options.GeoM.Translate(0, float64(unitHeight))
            }
        },
    }

    elements = append(elements, element)

    return elements
}
