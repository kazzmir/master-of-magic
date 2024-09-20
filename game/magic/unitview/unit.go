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

func RenderCombatImage(screen *ebiten.Image, imageCache *util.ImageCache, unit *units.Unit, options ebiten.DrawImageOptions) {
    images, err := imageCache.GetImages(unit.CombatLbxFile, unit.GetCombatIndex(units.FacingRight))
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
        combat.RenderCombatUnit(screen, use, options, unit.Count)
    }
}

func RenderUnitInfoBuild(screen *ebiten.Image, imageCache *util.ImageCache, unit *units.Unit, descriptionFont *font.Font, smallFont *font.Font, defaultOptions ebiten.DrawImageOptions) {
    x, y := defaultOptions.GeoM.Apply(0, 0)

    descriptionFont.Print(screen, x, y, 1, defaultOptions.ColorScale, unit.Name)

    smallFont.Print(screen, x, y + 11, 1, defaultOptions.ColorScale, "Moves")

    unitMoves := 2

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

    unitCostMoney := 2
    unitCostFood := 2

    smallCoin, err1 := imageCache.GetImage("backgrnd.lbx", 42, 0)
    smallFood, err2 := imageCache.GetImage("backgrnd.lbx", 40, 0)
    if err1 == nil && err2 == nil {
        var options ebiten.DrawImageOptions
        options = defaultOptions
        options.GeoM.Translate(smallFont.MeasureTextWidth("Upkeep ", 1), 18)
        for i := 0; i < unitCostMoney; i++ {
            screen.DrawImage(smallCoin, &options)
            options.GeoM.Translate(float64(smallCoin.Bounds().Dx() + 1), 0)
        }

        for i := 0; i < unitCostFood; i++ {
            screen.DrawImage(smallFood, &options)
            options.GeoM.Translate(float64(smallFood.Bounds().Dx() + 1), 0)
        }
    }

    cost := unit.ProductionCost
    smallFont.Print(screen, x, y + 27, 1, defaultOptions.ColorScale, fmt.Sprintf("Cost %v(%v)", cost, cost))
}

func RenderUnitInfoStats(screen *ebiten.Image, imageCache *util.ImageCache, unit *units.Unit, descriptionFont *font.Font, smallFont *font.Font, defaultOptions ebiten.DrawImageOptions) {
    width := descriptionFont.MeasureTextWidth("Armor", 1)

    x, y := defaultOptions.GeoM.Apply(0, 0)

    descriptionFont.Print(screen, x, y, 1, defaultOptions.ColorScale, "Melee")

    // show rows of icons. the second row is offset a bit to the right and down
    showNIcons := func(icon *ebiten.Image, count int, x, y float64) {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(x, y)
        options.GeoM.Translate(width + 1, 0)
        saveGeoM := options.GeoM
        for i := 0; i < count; i++ {
            if i > 0 && i % 5 == 0 {
                options.GeoM.Translate(3, 0)
            }

            if i > 0 && i % 15 == 0 {
                options.GeoM = saveGeoM
                options.GeoM.Translate(float64(3 * (i/15)), 2 * float64(i/15))
            }

            screen.DrawImage(icon, &options)
            // FIXME: if a stat is given due to an ability/spell then render the icon in gold
            options.GeoM.Translate(float64(icon.Bounds().Dx() + 1), 0)
        }
    }

    weaponIcon, err := imageCache.GetImage("unitview.lbx", 13, 0)
    if err == nil {
        showNIcons(weaponIcon, unit.MeleeAttackPower, x, y)
    }

    y += float64(descriptionFont.Height())
    descriptionFont.Print(screen, x, y, 1, defaultOptions.ColorScale, "Range")

    // FIXME: use the rock icon for sling, or the magic icon fire magic damage
    rangeBow, err := imageCache.GetImage("unitview.lbx", 18, 0)
    if err == nil {
        showNIcons(rangeBow, unit.RangedAttackPower, x, y)
    }

    y += float64(descriptionFont.Height())
    descriptionFont.Print(screen, x, float64(y), 1, defaultOptions.ColorScale, "Armor")

    armorIcon, err := imageCache.GetImage("unitview.lbx", 22, 0)
    if err == nil {
        showNIcons(armorIcon, unit.Defense, x, y)
    }

    y += float64(descriptionFont.Height())
    descriptionFont.Print(screen, x, float64(y), 1, defaultOptions.ColorScale, "Resist")

    resistIcon, err := imageCache.GetImage("unitview.lbx", 27, 0)
    if err == nil {
        showNIcons(resistIcon, unit.Resistance, x, y)
    }

    y += float64(descriptionFont.Height())
    descriptionFont.Print(screen, x, float64(y), 1, defaultOptions.ColorScale, "Hits")

    healthIcon, err := imageCache.GetImage("unitview.lbx", 23, 0)
    if err == nil {
        showNIcons(healthIcon, unit.HitPoints, x, y)
    }
}

func RenderUnitAbilities(screen *ebiten.Image, imageCache *util.ImageCache, unit *units.Unit, mediumFont *font.Font, defaultOptions ebiten.DrawImageOptions) {
    // FIXME: handle more than 4 abilities by using more columns
    for _, ability := range unit.Abilities {
        pic, err := imageCache.GetImage(ability.LbxFile(), ability.LbxIndex(), 0)
        if err == nil {
            screen.DrawImage(pic, &defaultOptions)
            x, y := defaultOptions.GeoM.Apply(0, 0)
            mediumFont.Print(screen, x + float64(pic.Bounds().Dx() + 2), float64(y) + 5, 1, defaultOptions.ColorScale, ability.Name())
            defaultOptions.GeoM.Translate(0, float64(pic.Bounds().Dy() + 1))
        }
    }
}
