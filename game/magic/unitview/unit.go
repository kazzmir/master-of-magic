package unitview

import (
    "log"
    "fmt"
    "math"
    "slices"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/fonts"
    helplib "github.com/kazzmir/master-of-magic/game/magic/help"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
    // "github.com/hajimehoshi/ebiten/v2/vector"
)

type CombatView interface {
    GetCombatLbxFile() string
    GetCombatIndex(facing units.Facing) int
    GetBanner() data.BannerType
    GetEnchantments() []data.UnitEnchantment
    GetCount() int
}

func RenderUnitViewImage(screen *ebiten.Image, imageCache *util.ImageCache, unit CombatView, options ebiten.DrawImageOptions, grey bool, counter uint64) {
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

        RenderCombatTile(screen, imageCache, options)

        first := util.First(unit.GetEnchantments(), data.UnitEnchantmentNone)
        if grey {
            RenderCombatUnitGrey(screen, use, options, unit.GetCount(), first, counter, imageCache)
        } else {
            RenderCombatUnit(screen, use, options, unit.GetCount(), first, counter, imageCache)
        }
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

func RenderUnitInfoNormal(screen *ebiten.Image, imageCache *util.ImageCache, unit UnitView, extraTitle string, namePrefix string, descriptionFont *font.Font, smallFont *font.Font, defaultOptions ebiten.DrawImageOptions) {
    x, y := defaultOptions.GeoM.Apply(0, 0)

    name := unit.GetName()
    if namePrefix != "" {
        name = fmt.Sprintf("%v %v", namePrefix, name)
    }

    if extraTitle != "" {
        descriptionFont.PrintOptions(screen, x, y, float64(data.ScreenScale), defaultOptions.ColorScale, font.FontOptions{DropShadow: true}, name)
        y += float64(descriptionFont.Height() * data.ScreenScale)
        defaultOptions.GeoM.Translate(0, float64(descriptionFont.Height() * data.ScreenScale))
        descriptionFont.PrintOptions(screen, x, y, float64(data.ScreenScale), defaultOptions.ColorScale, font.FontOptions{DropShadow: true}, "The " + extraTitle)

        y += float64(descriptionFont.Height() * data.ScreenScale)
        defaultOptions.GeoM.Translate(0, float64(descriptionFont.Height() * data.ScreenScale))
    } else {
        descriptionFont.PrintOptions(screen, x, y+float64(2 * data.ScreenScale), float64(data.ScreenScale), defaultOptions.ColorScale, font.FontOptions{DropShadow: true}, name)
        y += float64(17 * data.ScreenScale)
        defaultOptions.GeoM.Translate(0, float64(16 * data.ScreenScale))
    }

    defaultOptions.GeoM.Translate(0, float64(-1 * data.ScreenScale))

    smallFont.PrintOptions(screen, x, y, float64(data.ScreenScale), defaultOptions.ColorScale, font.FontOptions{DropShadow: true}, "Moves")
    y += float64((smallFont.Height() + 1) * data.ScreenScale)

    unitMoves := unit.GetMovementSpeed()

    // FIXME: show wings if flying, or the water thing if can walk on water
    smallBoot, err := imageCache.GetImage("unitview.lbx", 24, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options = defaultOptions
        options.GeoM.Translate(smallFont.MeasureTextWidth("Upkeep ", float64(data.ScreenScale)), 0)

        for i := 0; i < unitMoves; i++ {
            screen.DrawImage(smallBoot, &options)
            options.GeoM.Translate(float64(smallBoot.Bounds().Dx()), 0)
        }
    }

    smallFont.PrintOptions(screen, x, y, float64(data.ScreenScale), defaultOptions.ColorScale, font.FontOptions{DropShadow: true}, "Upkeep")

    options := defaultOptions
    options.GeoM.Translate(smallFont.MeasureTextWidth("Upkeep ", float64(data.ScreenScale)), float64((smallFont.Height() + 2) * data.ScreenScale))
    renderUpkeep(screen, imageCache, unit, options)
}

func RenderUnitInfoBuild(screen *ebiten.Image, imageCache *util.ImageCache, unit UnitView, descriptionFont *font.Font, smallFont *font.Font, defaultOptions ebiten.DrawImageOptions, discountedCost int) {
    x, y := defaultOptions.GeoM.Apply(0, 0)

    descriptionFont.PrintOptions(screen, x, y, float64(data.ScreenScale), defaultOptions.ColorScale, font.FontOptions{DropShadow: true}, unit.GetName())

    smallFont.PrintOptions(screen, x, y + float64(11 * data.ScreenScale), float64(data.ScreenScale), defaultOptions.ColorScale, font.FontOptions{DropShadow: true}, "Moves")

    unitMoves := unit.GetMovementSpeed()

    // FIXME: show wings if flying or the water thing if water walking
    smallBoot, err := imageCache.GetImage("unitview.lbx", 24, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options = defaultOptions
        options.GeoM.Translate(smallFont.MeasureTextWidth("Upkeep ", float64(data.ScreenScale)), float64(9 * data.ScreenScale))

        for i := 0; i < unitMoves; i++ {
            screen.DrawImage(smallBoot, &options)
            options.GeoM.Translate(float64(smallBoot.Bounds().Dx()), 0)
        }
    }

    smallFont.PrintOptions(screen, x, y + float64(19 * data.ScreenScale), float64(data.ScreenScale), defaultOptions.ColorScale, font.FontOptions{DropShadow: true}, "Upkeep")

    options := defaultOptions
    options.GeoM.Translate(smallFont.MeasureTextWidth("Upkeep ", float64(data.ScreenScale)), float64(18 * data.ScreenScale))
    renderUpkeep(screen, imageCache, unit, options)

    cost := unit.GetProductionCost()
    smallFont.PrintOptions(screen, x, y + float64(27 * data.ScreenScale), float64(data.ScreenScale), defaultOptions.ColorScale, font.FontOptions{DropShadow: true}, fmt.Sprintf("Cost %v(%v)", discountedCost, cost))
}

func RenderUnitInfoStats(screen *ebiten.Image, imageCache *util.ImageCache, unit UnitStats, maxIconsPerLine int, descriptionFont *font.Font, smallFont *font.Font, defaultOptions ebiten.DrawImageOptions) {
    width := descriptionFont.MeasureTextWidth("Armor", float64(data.ScreenScale))

    x, y := defaultOptions.GeoM.Apply(0, 0)

    descriptionFont.PrintOptions(screen, x, y, float64(data.ScreenScale), defaultOptions.ColorScale, font.FontOptions{DropShadow: true}, "Melee")

    // show rows of icons. the second row is offset a bit to the right and down
    showNIcons := func(icon *ebiten.Image, count int, icon2 *ebiten.Image, count2 int, negativeCount int, x, y float64) {
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
                options.GeoM.Translate(float64(3 * data.ScreenScale * index/maxIconsPerLine), 2 * float64(data.ScreenScale * index/maxIconsPerLine))
            }

            screen.DrawImage(icon, &options)
            // FIXME: if a stat is given due to an ability/spell then render the icon in gold
            options.GeoM.Translate(float64(icon.Bounds().Dx() + data.ScreenScale), 0)
        }

        index := 0
        for index < count {
            if index == count + count2 + negativeCount {
                options.ColorScale.ScaleWithColor(color.RGBA{R: 128, G: 128, B: 128, A: 255})
            }

            draw(index, icon)
            index += 1
        }

        for index < (count + count2) {
            if index == count + count2 + negativeCount {
                options.ColorScale.ScaleWithColor(color.RGBA{R: 128, G: 128, B: 128, A: 255})
            }

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

    showNIcons(weaponIcon, unit.GetBaseMeleeAttackPower(), weaponGold, unit.GetFullMeleeAttackPower() - unit.GetBaseMeleeAttackPower(), unit.GetMeleeAttackPower() - unit.GetFullMeleeAttackPower(), x, y)

    y += float64(descriptionFont.Height() * data.ScreenScale)
    descriptionFont.PrintOptions(screen, x, y, float64(data.ScreenScale), defaultOptions.ColorScale, font.FontOptions{DropShadow: true}, "Range")

    var rangeIcon *ebiten.Image
    var rangeIconGold *ebiten.Image

    switch unit.GetRangedAttackDamageType() {
        case units.DamageRangedMagical:
            rangeIcon, _ = imageCache.GetImage("unitview.lbx", 14, 0)
            rangeIconGold, _ = imageCache.GetImage("unitview.lbx", 36, 0)
        case units.DamageRangedPhysical:
            rangeIcon, _ = imageCache.GetImage("unitview.lbx", 18, 0)
            rangeIconGold, _ = imageCache.GetImage("unitview.lbx", 40, 0)
        case units.DamageRangedBoulder:
            rangeIcon, _ = imageCache.GetImage("unitview.lbx", 19, 0)
            rangeIconGold, _ = imageCache.GetImage("unitview.lbx", 41, 0)
    }

    showNIcons(rangeIcon, unit.GetBaseRangedAttackPower(), rangeIconGold, unit.GetFullRangedAttackPower() - unit.GetBaseRangedAttackPower(), unit.GetRangedAttackPower() - unit.GetFullRangedAttackPower(), x, y)

    y += float64(descriptionFont.Height() * data.ScreenScale)
    descriptionFont.PrintOptions(screen, x, float64(y), float64(data.ScreenScale), defaultOptions.ColorScale, font.FontOptions{DropShadow: true}, "Armor")

    armorIcon, _ := imageCache.GetImage("unitview.lbx", 22, 0)
    armorGold, _ := imageCache.GetImage("unitview.lbx", 44, 0)
    showNIcons(armorIcon, unit.GetBaseDefense(), armorGold, unit.GetFullDefense() - unit.GetBaseDefense(), unit.GetDefense() - unit.GetFullDefense(), x, y)

    y += float64(descriptionFont.Height() * data.ScreenScale)
    descriptionFont.PrintOptions(screen, x, float64(y), float64(data.ScreenScale), defaultOptions.ColorScale, font.FontOptions{DropShadow: true}, "Resist")

    resistIcon, _ := imageCache.GetImage("unitview.lbx", 27, 0)
    resistGold, _ := imageCache.GetImage("unitview.lbx", 49, 0)
    showNIcons(resistIcon, unit.GetBaseResistance(), resistGold, unit.GetFullResistance() - unit.GetBaseResistance(), unit.GetResistance() - unit.GetFullResistance(), x, y)

    y += float64(descriptionFont.Height() * data.ScreenScale)
    descriptionFont.PrintOptions(screen, x, float64(y), float64(data.ScreenScale), defaultOptions.ColorScale, font.FontOptions{DropShadow: true}, "Hits")

    healthIcon, _ := imageCache.GetImage("unitview.lbx", 23, 0)
    healthIconGold, _ := imageCache.GetImage("unitview.lbx", 45, 0)
    showNIcons(healthIcon, unit.GetBaseHitPoints(), healthIconGold, unit.GetFullHitPoints() - unit.GetBaseHitPoints(), unit.GetHitPoints() - unit.GetFullHitPoints(), x, y)
}

func RenderExperienceBadge(screen *ebiten.Image, imageCache *util.ImageCache, unit UnitExperience, showFont *font.Font, defaultOptions ebiten.DrawImageOptions, showExperience bool) (float64) {
    experience := unit.GetExperienceData()
    experienceIndex := 102 + experience.ToInt()
    pic, _ := imageCache.GetImage("special.lbx", experienceIndex, 0)
    screen.DrawImage(pic, &defaultOptions)
    x, y := defaultOptions.GeoM.Apply(0, 0)
    text := experience.Name()
    if showExperience {
        text = fmt.Sprintf("%v (%v ep)", experience.Name(), unit.GetExperience())
    }
    showFont.PrintOptions(screen, x + float64(pic.Bounds().Dx() + 2 * data.ScreenScale), y + float64(5 * data.ScreenScale), float64(data.ScreenScale), defaultOptions.ColorScale, font.FontOptions{DropShadow: true}, text)
    return float64(pic.Bounds().Dy() + 1 * data.ScreenScale)
}

func makeItemPopup(uiGroup *uilib.UIElementGroup, cache *lbx.LbxCache, imageCache *util.ImageCache, layer uilib.UILayer, item *artifact.Artifact) *uilib.UIElement {
    vaultFonts := fonts.MakeVaultFonts(cache)

    rect := image.Rect(0, 0, data.ScreenWidth, data.ScreenHeight)
    getAlpha := uiGroup.MakeFadeIn(7)
    element := &uilib.UIElement{
        Layer: layer+1,
        Rect: rect,
        LeftClick: func(element *uilib.UIElement){
            getAlpha = uiGroup.MakeFadeOut(7)
            uiGroup.AddDelay(7, func(){
                uiGroup.RemoveElement(element)
            })
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            options.GeoM.Translate(float64(48 * data.ScreenScale), float64(48 * data.ScreenScale))
            artifact.RenderArtifactBox(screen, imageCache, *item, uiGroup.Counter / 8, vaultFonts.ItemName, vaultFonts.PowerFont, options)
        },
    }

    return element
}

func createUnitAbilitiesElements(cache *lbx.LbxCache, imageCache *util.ImageCache, uiGroup *uilib.UIElementGroup, unit UnitAbilities, mediumFont *font.Font, x int, y int, counter *uint64, layer uilib.UILayer, getAlpha *util.AlphaFadeFunc, pureAbilities bool, page uint32, help *helplib.Help, allowDismiss bool, updateAbilities func()) []*uilib.UIElement {
    xStart := x
    yStart := y

    var elements []*uilib.UIElement

    background, _ := imageCache.GetImage("special.lbx", 3, 0)

    if !pureAbilities {
        // experience badge
        if unit.GetRace() != data.RaceFantastic {
            experienceX := x
            experienceY := y
            elements = append(elements, &uilib.UIElement{
                Layer: layer,
                Rect: util.ImageRect(experienceX, experienceY, background),
                RightClick: func(element *uilib.UIElement){
                    experience := unit.GetExperienceData()
                    helpEntries := help.GetEntriesByName(experience.Name())
                    if helpEntries != nil {
                        uiGroup.AddElement(uilib.MakeHelpElementWithLayer(uiGroup, cache, imageCache, layer+1, helpEntries[0], helpEntries[1:]...))
                    }
                },
                Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                    var options ebiten.DrawImageOptions
                    options.ColorScale.ScaleAlpha((*getAlpha)())
                    options.GeoM.Translate(float64(experienceX), float64(experienceY))
                    RenderExperienceBadge(screen, imageCache, unit, mediumFont, options, true)
                },
            })
        }

        y += background.Bounds().Dy() + 1

        artifacts := slices.Clone(unit.GetArtifacts())

        for _, slot := range unit.GetArtifactSlots() {
            rect := util.ImageRect(x, y, background)

            var showArtifact *artifact.Artifact
            for _, check := range artifacts {
                if check == nil {
                    continue
                }

                if slot.CompatibleWith(check.Type) {
                    showArtifact = check
                    break
                }
            }

            elements = append(elements, &uilib.UIElement{
                Rect: rect,
                Layer: layer,
                RightClick: func(element *uilib.UIElement){
                    if showArtifact != nil {
                        uiGroup.AddElement(makeItemPopup(uiGroup, cache, imageCache, layer+1, showArtifact))
                    }
                },
                Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))
                    options.ColorScale.ScaleAlpha((*getAlpha)())

                    screen.DrawImage(background, &options)

                    if showArtifact != nil {
                        artifactPic := artifact.RenderArtifactImage(screen, imageCache, *showArtifact, *counter, options)

                        x, y := options.GeoM.Apply(0, 0)
                        printX := x + float64(artifactPic.Bounds().Dx() + 2 * data.ScreenScale)
                        printY := y + float64(5 * data.ScreenScale)
                        mediumFont.PrintOptions(screen, printX, printY, float64(data.ScreenScale), options.ColorScale, font.FontOptions{DropShadow: true}, showArtifact.Name)
                    } else {
                        pic, _ := imageCache.GetImage("itemisc.lbx", slot.ImageIndex() + 8, 0)
                        screen.DrawImage(pic, &options)
                    }
                },
            })

            y += background.Bounds().Dy() + 1
        }
    }

    // FIXME: handle more than 4 abilities by using more columns
    for _, ability := range unit.GetAbilities() {
        pic, err := imageCache.GetImage(ability.LbxFile(), ability.LbxIndex(), 0)
        if err == nil {
            rect := util.ImageRect(x, y, pic)
            elements = append(elements, &uilib.UIElement{
                Layer: layer,
                Rect: rect,
                RightClick: func(element *uilib.UIElement){
                    helpEntries := help.GetEntriesByName(ability.Name())
                    if helpEntries != nil {
                        uiGroup.AddElement(uilib.MakeHelpElementWithLayer(uiGroup, cache, imageCache, layer+1, helpEntries[0], helpEntries[1:]...))
                    }
                },
                Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                    var options ebiten.DrawImageOptions
                    options.ColorScale.ScaleAlpha((*getAlpha)())
                    options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))
                    screen.DrawImage(pic, &options)
                    x, y := options.GeoM.Apply(0, 0)

                    printX := x + float64(pic.Bounds().Dx() + 2 * data.ScreenScale)
                    printY := y + float64(5 * data.ScreenScale)
                    mediumFont.PrintOptions(screen, printX, printY, float64(data.ScreenScale), options.ColorScale, font.FontOptions{DropShadow: true}, ability.Name())
                },
            })

            y += pic.Bounds().Dy() + 1
        } else {
            log.Printf("Error: unable to render ability %#v %v", ability, ability.Name())
        }
    }

    for _, enchantment := range unit.GetEnchantments() {
        pic, err := imageCache.GetImage(enchantment.LbxFile(), enchantment.LbxIndex(), 0)

        if err == nil {
            rect := util.ImageRect(x, y, pic)
            elements = append(elements, &uilib.UIElement{
                Layer: layer,
                Rect: rect,
                RightClick: func(element *uilib.UIElement){
                    helpEntries := help.GetEntriesByName(enchantment.Name())
                    if helpEntries != nil {
                        uiGroup.AddElement(uilib.MakeHelpElementWithLayer(uiGroup, cache, imageCache, layer+1, helpEntries[0], helpEntries[1:]...))
                    }
                },
                LeftClick: func(element *uilib.UIElement){
                    if allowDismiss {
                        message := fmt.Sprintf("Do you wish to turn off the %v spell?", enchantment.Name())
                        confirm := func(){
                            unit.RemoveEnchantment(enchantment)
                            updateAbilities()
                        }

                        cancel := func(){
                        }

                        uiGroup.AddElements(uilib.MakeConfirmDialogWithLayer(uiGroup, cache, imageCache, layer+1, message, false, confirm, cancel))
                    }
                },
                Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                    var options ebiten.DrawImageOptions
                    options.ColorScale.ScaleAlpha((*getAlpha)())
                    options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))
                    screen.DrawImage(pic, &options)
                    x, y := options.GeoM.Apply(0, 0)

                    printX := x + float64(pic.Bounds().Dx() + 2 * data.ScreenScale)
                    printY := y + float64(5 * data.ScreenScale)
                    mediumFont.PrintOptions(screen, printX, printY, float64(data.ScreenScale), options.ColorScale, font.FontOptions{DropShadow: true}, enchantment.Name())
                },
            })

            y += pic.Bounds().Dy() + 1
        } else {
            log.Printf("Error: unable to render enchantment %#v %v", enchantment, enchantment.Name())
        }
    }

    if len(elements) == 0 {
        return nil
    }

    pages := uint32(math.Ceil(float64(len(elements)) / 4))
    for page < 0 {
        page += pages
    }
    page = page % pages

    minElement := page * 4
    maxElement := int(min(float64(len(elements)), float64((page + 1) * 4)))

    outElements := elements[minElement:maxElement]
    for i, element := range outElements {
        element.Rect.Min.X = xStart
        element.Rect.Min.Y = yStart + i * (background.Bounds().Dy() + 1)
        element.Rect.Max.Y = element.Rect.Min.Y + background.Bounds().Dy()
    }

    return outElements
}

// if allowDismiss is true then enchantments can be dispelled when the user left clicks on them
func MakeUnitAbilitiesElements(group *uilib.UIElementGroup, cache *lbx.LbxCache, imageCache *util.ImageCache, unit UnitAbilities, mediumFont *font.Font, x int, y int, counter *uint64, layer uilib.UILayer, getAlpha *util.AlphaFadeFunc, pureAbilities bool, page uint32, allowDismiss bool) []*uilib.UIElement {
    var elements []*uilib.UIElement

    var abilityElements []*uilib.UIElement

    // possibly pass in the help data rather than reading it every time
    helpLbx, err := cache.GetLbxFile("help.lbx")
    if err != nil {
        return nil
    }

    help, err := helplib.ReadHelp(helpLbx, 2)
    if err != nil {
        return nil
    }

    // after removing an enchantment, recreate the entire ability list
    updateAbilities := func(){
        group.RemoveElements(elements)
        group.RemoveElements(abilityElements)
        // pass in the current page so the ui doesn't jump around
        group.AddElements(MakeUnitAbilitiesElements(group, cache, imageCache, unit, mediumFont, x, y, counter, layer, getAlpha, pureAbilities, page, allowDismiss))
    }

    abilityElements = createUnitAbilitiesElements(cache, imageCache, group, unit, mediumFont, x, y, counter, layer, getAlpha, pureAbilities, page, &help, allowDismiss, updateAbilities)

    elements = append(elements, abilityElements...)

    upImages, _ := imageCache.GetImages("unitview.lbx", 3)
    downImages, _ := imageCache.GetImages("unitview.lbx", 4)

    abilityCount := len(unit.GetAbilities())
    if !pureAbilities {
        // 1 for experience
        abilityCount += 1
        // 3 more for items
        abilityCount += len(unit.GetArtifactSlots())
    }

    abilityCount += len(unit.GetEnchantments())

    if abilityCount > 4 {
        pageUpRect := util.ImageRect(x + 195 * data.ScreenScale, y, upImages[0])
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

                group.RemoveElements(abilityElements)
                abilityElements = createUnitAbilitiesElements(cache, imageCache, group, unit, mediumFont, x, y, counter, layer, getAlpha, pureAbilities, page, &help, allowDismiss, updateAbilities)
                group.AddElements(abilityElements)
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha((*getAlpha)())
                options.GeoM.Translate(float64(pageUpRect.Min.X), float64(pageUpRect.Min.Y))
                screen.DrawImage(upImages[pageUpIndex], &options)
            },
        })

        pageDownRect := util.ImageRect(x + 195 * data.ScreenScale, y + 60 * data.ScreenScale, downImages[0])
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

                group.RemoveElements(abilityElements)
                abilityElements = createUnitAbilitiesElements(cache, imageCache, group, unit, mediumFont, x, y, counter, layer, getAlpha, pureAbilities, page, &help, allowDismiss, updateAbilities)
                group.AddElements(abilityElements)
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha((*getAlpha)())
                options.GeoM.Translate(float64(pageDownRect.Min.X), float64(pageDownRect.Min.Y))
                screen.DrawImage(downImages[pageDownIndex], &options)
            },
        })
    }

    return elements
}
