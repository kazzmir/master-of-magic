package unitview

import (
    // "image/color"
    "log"
    "fmt"
    "math"
    "slices"

    "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/lib/font"

    "github.com/hajimehoshi/ebiten/v2"
    // "github.com/hajimehoshi/ebiten/v2/vector"
)

func RenderCombatImage(screen *ebiten.Image, imageCache *util.ImageCache, unit UnitView, options ebiten.DrawImageOptions, counter uint64) {
    images, err := imageCache.GetImagesTransform(unit.GetCombatLbxFile(), unit.GetCombatIndex(units.FacingRight), unit.GetBanner().String(), units.MakeUpdateUnitColorsFunc(unit.GetBanner()))
    if err == nil && len(images) > 2 {
        use := images[2]
        // log.Printf("unitview.RenderCombatImage: %v", use.Bounds())
        options.GeoM.Translate(float64(0), float64(0))

        /*
        x, y := options.GeoM.Apply(0, 0)
        log.Printf("render combat image at %v, %v", x, y)
        vector.DrawFilledCircle(screen, float32(x), float32(y), 3, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, false)
        */

        combat.RenderCombatTile(screen, imageCache, options)

        enchantment := util.First(unit.GetEnchantments(), data.UnitEnchantmentNone)
        combat.RenderCombatUnit(screen, use, options, unit.GetCount(), enchantment, counter, imageCache)
    }
}

func renderUpkeep(screen *ebiten.Image, imageCache *util.ImageCache, unit UnitView, options ebiten.DrawImageOptions) {
    unitCostMoney := unit.GetUpkeepGold()
    unitCostFood := unit.GetUpkeepFood()
    unitCostMana := unit.GetUpkeepMana()

    smallCoin, _ := imageCache.GetImage("backgrnd.lbx", 42, 0)
    smallFood, _ := imageCache.GetImage("backgrnd.lbx", 40, 0)
    smallMana, _ := imageCache.GetImage("backgrnd.lbx", 43, 0)

    bigCoin, _ := imageCache.GetImage("backgrnd.lbx", 90, 0)
    bigFood, _ := imageCache.GetImage("backgrnd.lbx", 88, 0)
    bigMana, _ := imageCache.GetImage("backgrnd.lbx", 91, 0)

    renderIcons := func(count int, small *ebiten.Image, big *ebiten.Image){
        for i := 0; i < count / 10; i++ {
            screen.DrawImage(big, &options)
            options.GeoM.Translate(float64(big.Bounds().Dx() + 1), 0)
        }

        for i := 0; i < count % 10; i++ {
            screen.DrawImage(small, &options)
            options.GeoM.Translate(float64(small.Bounds().Dx() + 1), 0)
        }
    }

    renderIcons(unitCostMoney, smallCoin, bigCoin)
    renderIcons(unitCostFood, smallFood, bigFood)
    renderIcons(unitCostMana, smallMana, bigMana)
}

func RenderUnitInfoNormal(screen *ebiten.Image, imageCache *util.ImageCache, unit UnitView, extraTitle string, descriptionFont *font.Font, smallFont *font.Font, defaultOptions ebiten.DrawImageOptions) {
    x, y := defaultOptions.GeoM.Apply(0, 0)

    if extraTitle != "" {
        descriptionFont.Print(screen, x, y, 1, defaultOptions.ColorScale, unit.GetName())
        y += float64(descriptionFont.Height())
        defaultOptions.GeoM.Translate(0, float64(descriptionFont.Height()))
        descriptionFont.Print(screen, x, y, 1, defaultOptions.ColorScale, "The " + extraTitle)

        y += float64(descriptionFont.Height())
        defaultOptions.GeoM.Translate(0, float64(descriptionFont.Height()))
    } else {
        descriptionFont.Print(screen, x, y+2, 1, defaultOptions.ColorScale, unit.GetName())
        y += 17
        defaultOptions.GeoM.Translate(0, 16)
    }

    defaultOptions.GeoM.Translate(0, -1)

    smallFont.Print(screen, x, y, 1, defaultOptions.ColorScale, "Moves")
    y += float64(smallFont.Height()) + 1

    unitMoves := unit.GetMovementSpeed()

    // FIXME: show wings if flying, or the water thing if can walk on water
    smallBoot, err := imageCache.GetImage("unitview.lbx", 24, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options = defaultOptions
        options.GeoM.Translate(smallFont.MeasureTextWidth("Upkeep ", 1), 0)

        for i := 0; i < unitMoves; i++ {
            screen.DrawImage(smallBoot, &options)
            options.GeoM.Translate(float64(smallBoot.Bounds().Dx()), 0)
        }
    }

    smallFont.Print(screen, x, y, 1, defaultOptions.ColorScale, "Upkeep")

    options := defaultOptions
    options.GeoM.Translate(smallFont.MeasureTextWidth("Upkeep ", 1), float64(smallFont.Height()) + 2)
    renderUpkeep(screen, imageCache, unit, options)
}

func RenderUnitInfoBuild(screen *ebiten.Image, imageCache *util.ImageCache, unit UnitView, descriptionFont *font.Font, smallFont *font.Font, defaultOptions ebiten.DrawImageOptions) {
    x, y := defaultOptions.GeoM.Apply(0, 0)

    descriptionFont.Print(screen, x, y, 1, defaultOptions.ColorScale, unit.GetName())

    smallFont.Print(screen, x, y + 11, 1, defaultOptions.ColorScale, "Moves")

    unitMoves := unit.GetMovementSpeed()

    // FIXME: show wings if flying or the water thing if water walking
    smallBoot, err := imageCache.GetImage("unitview.lbx", 24, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options = defaultOptions
        options.GeoM.Translate(smallFont.MeasureTextWidth("Upkeep ", 1), 9)

        for i := 0; i < unitMoves; i++ {
            screen.DrawImage(smallBoot, &options)
            options.GeoM.Translate(float64(smallBoot.Bounds().Dx()), 0)
        }
    }

    smallFont.Print(screen, x, y + 19, 1, defaultOptions.ColorScale, "Upkeep")

    options := defaultOptions
    options.GeoM.Translate(smallFont.MeasureTextWidth("Upkeep ", 1), 18)
    renderUpkeep(screen, imageCache, unit, options)

    cost := unit.GetProductionCost()
    // FIXME: compute discounted cost based on the unit being built and the tiles surrounding the city
    discountedCost := cost
    smallFont.Print(screen, x, y + 27, 1, defaultOptions.ColorScale, fmt.Sprintf("Cost %v(%v)", discountedCost, cost))
}

func RenderUnitInfoStats(screen *ebiten.Image, imageCache *util.ImageCache, unit UnitView, maxIconsPerLine int, descriptionFont *font.Font, smallFont *font.Font, defaultOptions ebiten.DrawImageOptions) {
    width := descriptionFont.MeasureTextWidth("Armor", 1)

    x, y := defaultOptions.GeoM.Apply(0, 0)

    descriptionFont.Print(screen, x, y, 1, defaultOptions.ColorScale, "Melee")

    // show rows of icons. the second row is offset a bit to the right and down
    showNIcons := func(icon *ebiten.Image, count int, icon2 *ebiten.Image, count2 int, x, y float64) {
        var options ebiten.DrawImageOptions
        options = defaultOptions
        options.GeoM.Reset()
        options.GeoM.Translate(x, y)
        options.GeoM.Translate(width + 1, 0)
        saveGeoM := options.GeoM

        draw := func (index int, icon *ebiten.Image) {
            if index > 0 && index % 5 == 0 {
                options.GeoM.Translate(3, 0)
            }

            if index > 0 && index % maxIconsPerLine == 0 {
                options.GeoM = saveGeoM
                options.GeoM.Translate(float64(3 * (index/maxIconsPerLine)), 2 * float64(index/maxIconsPerLine))
            }

            screen.DrawImage(icon, &options)
            // FIXME: if a stat is given due to an ability/spell then render the icon in gold
            options.GeoM.Translate(float64(icon.Bounds().Dx() + 1), 0)
        }

        index := 0
        for index < count {
            draw(index, icon)
            index += 1
        }

        for index < (count + count2) {
            draw(index, icon2)
            index += 1
        }
    }

    // change the melee type depending on the unit attributes (hero uses magic sword), but
    // mythril or admantanium is also possible
    weaponIcon, _ := imageCache.GetImage("unitview.lbx", 13, 0)
    weaponGold, _ := imageCache.GetImage("unitview.lbx", 35, 0)

    switch unit.GetWeaponBonus() {
        case data.WeaponMagic:
            weaponIcon, _ = imageCache.GetImage("unitview.lbx", 16, 0)
            weaponGold, _ = imageCache.GetImage("unitview.lbx", 38, 0)
        case data.WeaponMythril:
            weaponIcon, _ = imageCache.GetImage("unitview.lbx", 15, 0)
            weaponGold, _ = imageCache.GetImage("unitview.lbx", 37, 0)
        case data.WeaponAdamantium:
            weaponIcon, _ = imageCache.GetImage("unitview.lbx", 17, 0)
            weaponGold, _ = imageCache.GetImage("unitview.lbx", 39, 0)
    }

    showNIcons(weaponIcon, unit.GetBaseMeleeAttackPower(), weaponGold, unit.GetMeleeAttackPower() - unit.GetBaseMeleeAttackPower(), x, y)

    y += float64(descriptionFont.Height())
    descriptionFont.Print(screen, x, y, 1, defaultOptions.ColorScale, "Range")

    // FIXME: use the rock icon for sling, or the magic icon fire magic damage
    rangeBow, _ := imageCache.GetImage("unitview.lbx", 18, 0)
    rangeBowGold, _ := imageCache.GetImage("unitview.lbx", 40, 0)
    showNIcons(rangeBow, unit.GetBaseRangedAttackPower(), rangeBowGold, unit.GetRangedAttackPower() - unit.GetBaseRangedAttackPower(), x, y)

    y += float64(descriptionFont.Height())
    descriptionFont.Print(screen, x, float64(y), 1, defaultOptions.ColorScale, "Armor")

    armorIcon, _ := imageCache.GetImage("unitview.lbx", 22, 0)
    armorGold, _ := imageCache.GetImage("unitview.lbx", 44, 0)
    showNIcons(armorIcon, unit.GetBaseDefense(), armorGold, unit.GetDefense() - unit.GetBaseDefense(), x, y)

    y += float64(descriptionFont.Height())
    descriptionFont.Print(screen, x, float64(y), 1, defaultOptions.ColorScale, "Resist")

    resistIcon, _ := imageCache.GetImage("unitview.lbx", 27, 0)
    resistGold, _ := imageCache.GetImage("unitview.lbx", 49, 0)
    showNIcons(resistIcon, unit.GetResistance(), resistGold, unit.GetResistance() - unit.GetBaseResistance(), x, y)

    y += float64(descriptionFont.Height())
    descriptionFont.Print(screen, x, float64(y), 1, defaultOptions.ColorScale, "Hits")

    healthIcon, _ := imageCache.GetImage("unitview.lbx", 23, 0)
    healthIconGold, _ := imageCache.GetImage("unitview.lbx", 45, 0)
    showNIcons(healthIcon, unit.GetBaseHitPoints(), healthIconGold, unit.GetHitPoints() - unit.GetBaseHitPoints(), x, y)
}

func renderUnitAbilities(screen *ebiten.Image, imageCache *util.ImageCache, unit UnitView, mediumFont *font.Font, defaultOptions ebiten.DrawImageOptions, pureAbilities bool, page uint32) {
    var renders []func() float64

    if !pureAbilities {
        // experience badge
        renders = append(renders, func() float64 {
            data := unit.GetExperienceData()
            experienceIndex := 102 + data.ToInt()
            pic, _ := imageCache.GetImage("special.lbx", experienceIndex, 0)
            screen.DrawImage(pic, &defaultOptions)
            x, y := defaultOptions.GeoM.Apply(0, 0)
            mediumFont.Print(screen, x + float64(pic.Bounds().Dx() + 2), float64(y) + 5, 1, defaultOptions.ColorScale, fmt.Sprintf("%v (%v ep)", data.Name(), unit.GetExperience()))
            return float64(pic.Bounds().Dy() + 1)
        })

        artifacts := slices.Clone(unit.GetArtifacts())

        background, _ := imageCache.GetImage("special.lbx", 3, 0)

        for _, slot := range unit.GetArtifactSlots() {
            renders = append(renders, func() float64 {
                for i := 0; i < len(artifacts); i++ {
                    if artifacts[i] == nil {
                        continue
                    }

                    if slot.CompatibleWith(artifacts[i].Type) {
                        screen.DrawImage(background, &defaultOptions)

                        artifactPic, _ := imageCache.GetImage("items.lbx", artifacts[i].Image, 0)
                        screen.DrawImage(artifactPic, &defaultOptions)

                        artifacts = slices.Delete(artifacts, i, i+1)
                        return float64(artifactPic.Bounds().Dy() + 1)
                    }
                }

                pic, _ := imageCache.GetImage("itemisc.lbx", slot.ImageIndex() + 8, 0)
                screen.DrawImage(pic, &defaultOptions)

                return float64(pic.Bounds().Dy() + 1)
            })
        }

    }

    // FIXME: handle more than 4 abilities by using more columns
    for _, ability := range unit.GetAbilities() {
        renders = append(renders, func() float64 {
            pic, err := imageCache.GetImage(ability.LbxFile(), ability.LbxIndex(), 0)
            if err == nil {
                screen.DrawImage(pic, &defaultOptions)
                x, y := defaultOptions.GeoM.Apply(0, 0)
                mediumFont.Print(screen, x + float64(pic.Bounds().Dx() + 2), float64(y) + 5, 1, defaultOptions.ColorScale, ability.Name())
                return float64(pic.Bounds().Dy() + 1)
            } else {
                log.Printf("Error: unable to render ability %#v %v", ability, ability.Name())
            }

            return 0
        })
    }

    if len(renders) == 0 {
        return
    }

    pages := uint32(math.Ceil(float64(len(renders)) / 4))
    page = page % pages

    for i := int(page) * 4; i < len(renders) && i < (int(page) + 1) * 4; i++ {
        height := renders[i]()
        defaultOptions.GeoM.Translate(0, height)
    }
}

func MakeUnitAbilitiesElements(imageCache *util.ImageCache, unit UnitView, mediumFont *font.Font, x int, y int, layer uilib.UILayer, getAlpha *util.AlphaFadeFunc, pureAbilities bool) []*uilib.UIElement {
    var elements []*uilib.UIElement

    page := uint32(0)

    elements = append(elements, &uilib.UIElement{
        Layer: layer,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha((*getAlpha)())
            options.GeoM.Translate(float64(x), float64(y))
            renderUnitAbilities(screen, imageCache, unit, mediumFont, options, pureAbilities, page)
        },
    })

    upImages, _ := imageCache.GetImages("unitview.lbx", 3)
    downImages, _ := imageCache.GetImages("unitview.lbx", 4)

    abilityCount := len(unit.GetAbilities())
    if !pureAbilities {
        // 1 for experience
        abilityCount += 1
        // 3 more for items
        abilityCount += len(unit.GetArtifactSlots())
    }

    if abilityCount > 4 {
        pageUpRect := util.ImageRect(x + 195, y, upImages[0])
        pageUpIndex := 0
        elements = append(elements, &uilib.UIElement{
            Rect: pageUpRect,
            Layer: layer,
            LeftClick: func(element *uilib.UIElement){
                pageUpIndex = 1
            },
            LeftClickRelease: func(element *uilib.UIElement){
                pageUpIndex = 0
                page -= 1
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(pageUpRect.Min.X), float64(pageUpRect.Min.Y))
                screen.DrawImage(upImages[pageUpIndex], &options)
            },
        })

        pageDownRect := util.ImageRect(x + 195, y + 60, downImages[0])
        pageDownIndex := 0
        elements = append(elements, &uilib.UIElement{
            Rect: pageDownRect,
            Layer: layer,
            LeftClick: func(element *uilib.UIElement){
                pageDownIndex = 1
            },
            LeftClickRelease: func(element *uilib.UIElement){
                pageDownIndex = 0
                page += 1
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(pageDownRect.Min.X), float64(pageDownRect.Min.Y))
                screen.DrawImage(downImages[pageDownIndex], &options)
            },
        })
    }

    return elements
}
