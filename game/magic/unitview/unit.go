package unitview

import (
    // "image/color"
    // "log"
    "fmt"

    "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/lib/font"

    "github.com/hajimehoshi/ebiten/v2"
    // "github.com/hajimehoshi/ebiten/v2/vector"
)

func RenderCombatImage(screen *ebiten.Image, imageCache *util.ImageCache, unit UnitView, options ebiten.DrawImageOptions) {
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
        combat.RenderCombatUnit(screen, use, options, unit.GetCount())
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

    descriptionFont.Print(screen, x, y, 1, defaultOptions.ColorScale, unit.GetName())

    if extraTitle != "" {
        y += float64(descriptionFont.Height())
        defaultOptions.GeoM.Translate(0, float64(descriptionFont.Height()))
        descriptionFont.Print(screen, x, y, 1, defaultOptions.ColorScale, extraTitle)

        y += float64(descriptionFont.Height())
        defaultOptions.GeoM.Translate(0, float64(descriptionFont.Height()))

    } else {
        y += 15
        defaultOptions.GeoM.Translate(0, 14)
    }

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

func RenderUnitAbilities(screen *ebiten.Image, imageCache *util.ImageCache, unit UnitView, mediumFont *font.Font, defaultOptions ebiten.DrawImageOptions) {
    // FIXME: handle more than 4 abilities by using more columns
    for _, ability := range unit.GetAbilities() {
        pic, err := imageCache.GetImage(ability.LbxFile(), ability.LbxIndex(), 0)
        if err == nil {
            screen.DrawImage(pic, &defaultOptions)
            x, y := defaultOptions.GeoM.Apply(0, 0)
            mediumFont.Print(screen, x + float64(pic.Bounds().Dx() + 2), float64(y) + 5, 1, defaultOptions.ColorScale, ability.Name())
            defaultOptions.GeoM.Translate(0, float64(pic.Bounds().Dy() + 1))
        }
    }
}
