package artifact

import (
    "fmt"
    "image"
    "strings"
    "slices"
    "cmp"
    "image/color"
    "log"
    "bytes"

    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/lib/functional"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
    // "github.com/hajimehoshi/ebiten/v2/inpututil"
)

// the screen can be invoked as either the 'Enchant Item' spell or 'Create Artifact'
type CreationScreen int
const (
    CreationEnchantItem CreationScreen = iota
    CreationCreateArtifact
)

const CreationScreenCostThreshold = 200  // TODO: validate this again with abilities

// a way to know how many books of magic the wizard has
type MagicLevel interface {
    // for a given magic type, returns the number of books of that magic
    MagicLevel(kind data.MagicType) int
}

func ReadPowers(cache *lbx.LbxCache) ([]Power, map[Power]int, map[Power]set.Set[ArtifactType], error) {
    itemData, err := cache.GetLbxFile("itempow.lbx")
    if err != nil {
        return nil, nil, nil, fmt.Errorf("unable to read itempow.lbx: %v", err)
    }

    reader, err := itemData.GetReader(0)
    if err != nil {
        return nil, nil, nil, fmt.Errorf("unable to read entry 0 in itempow.lbx: %v", err)
    }

    numEntries, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil, nil, nil, fmt.Errorf("read error: %v", err)
    }

    entrySize, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil, nil, nil, fmt.Errorf("read error: %v", err)
    }
    if entrySize != 30 {
        return nil, nil, nil, fmt.Errorf("unsupported itempow.lbx")
    }

    artifactTypeMap := map[uint16]ArtifactType{
        1 << 0: ArtifactTypeSword,
        1 << 1: ArtifactTypeMace,
        1 << 2: ArtifactTypeAxe,
        1 << 3: ArtifactTypeBow,
        1 << 4: ArtifactTypeStaff,
        1 << 5: ArtifactTypeWand,
        1 << 6: ArtifactTypeMisc,
        1 << 7: ArtifactTypeShield,
        1 << 8: ArtifactTypeChain,
        1 << 9: ArtifactTypePlate,
    }

    powerTypeMap := map[byte]PowerType{
        0: PowerTypeAttack,
        1: PowerTypeToHit,
        2: PowerTypeDefense,
        3: PowerTypeSpellSkill,
        4: PowerTypeSpellSave,
        5: PowerTypeMovement,
        6: PowerTypeResistance,
        7: PowerTypeAbility1,
        8: PowerTypeAbility2,
        9: PowerTypeAbility3,
    }

    magicTypeMap := map[byte]data.MagicType{
        0: data.NatureMagic,
        1: data.SorceryMagic,
        2: data.ChaosMagic,
        3: data.LifeMagic,
        4: data.DeathMagic,
    }

    abilityMap := map[uint32]data.ItemAbility{
        1 << 0:  data.ItemAbilityVampiric,
        1 << 1:  data.ItemAbilityGuardianWind,
        1 << 2:  data.ItemAbilityLightning,
        1 << 3:  data.ItemAbilityCloakOfFear,
        1 << 4:  data.ItemAbilityDestruction,
        1 << 5:  data.ItemAbilityWraithform,
        1 << 6:  data.ItemAbilityRegeneration,
        1 << 7:  data.ItemAbilityPathfinding,
        1 << 8:  data.ItemAbilityWaterWalking,
        1 << 9:  data.ItemAbilityResistElements,
        1 << 10: data.ItemAbilityElementalArmor,
        1 << 11: data.ItemAbilityChaos,
        1 << 12: data.ItemAbilityStoning,
        1 << 13: data.ItemAbilityEndurance,
        1 << 14: data.ItemAbilityHaste,
        1 << 15: data.ItemAbilityInvisibility,
        1 << 16: data.ItemAbilityDeath,
        1 << 17: data.ItemAbilityFlight,
        1 << 18: data.ItemAbilityResistMagic,
        1 << 19: data.ItemAbilityMagicImmunity,
        1 << 20: data.ItemAbilityFlaming,
        1 << 21: data.ItemAbilityHolyAvenger,
        1 << 22: data.ItemAbilityTrueSight,
        1 << 23: data.ItemAbilityPhantasmal,
        1 << 24: data.ItemAbilityPowerDrain,
        1 << 25: data.ItemAbilityBless,
        1 << 26: data.ItemAbilityLionHeart,
        1 << 27: data.ItemAbilityGiantStrength,
        1 << 28: data.ItemAbilityPlanarTravel,
        1 << 29: data.ItemAbilityMerging,
        1 << 30: data.ItemAbilityRighteousness,
        1 << 31: data.ItemAbilityInvulnerability,
    }

    var powers []Power
    costs := make(map[Power]int)
    compatibilities := make(map[Power]set.Set[ArtifactType])

    for i := 0; i < int(numEntries); i++ {
        // Name
        name := make([]byte, 18)
        n, err := reader.Read(name)
        if err != nil || n != int(18) {
            return nil, nil, nil, fmt.Errorf("unable to read item name %v: %v", i, err)
        }
        name = bytes.Trim(name, " \x00")

        // Artifact types
        artifactTypeValue, err := lbx.ReadUint16(reader)
        if err != nil {
            return nil, nil, nil, fmt.Errorf("read error: %v", err)
        }
        if artifactTypeValue == 0 {
            continue  // there seems to be empty entries
        }

        artifactTypes := set.MakeSet[ArtifactType]()
        for mask, artifactType := range artifactTypeMap {
            if artifactTypeValue & mask != 0 {
                artifactTypes.Insert(artifactType)
            }
        }

        // Cost
        cost, err := lbx.ReadUint16(reader)
        if err != nil {
            return nil, nil, nil, fmt.Errorf("read error: %v", err)
        }

        // Power type
        powerTypeValue, err := lbx.ReadByte(reader)
        if err != nil {
            return nil, nil, nil, fmt.Errorf("read error: %v", err)
        }
        powerType, exists := powerTypeMap[powerTypeValue]
        if !exists {
            return nil, nil, nil, fmt.Errorf("Invalid power type %v", powerTypeValue)
        }

        // Magic type
        magicTypeValue, err := lbx.ReadByte(reader)
        if err != nil {
            return nil, nil, nil, fmt.Errorf("read error: %v", err)
        }
        magicType, exists := magicTypeMap[magicTypeValue]
        if !exists {
            return nil, nil, nil, fmt.Errorf("Invalid magic type %v", magicTypeValue)
        }

        // Amount
        amount, err := lbx.ReadUint16(reader)
        if err != nil {
            return nil, nil, nil, fmt.Errorf("read error: %v", err)
        }
        if string(name) == "+6 Defense" {
            amount = 6
        }

        // Abilities
        abilitiesValue, err := lbx.ReadUint32(reader)
        if err != nil {
            return nil, nil, nil, fmt.Errorf("read error: %v", err)
        }

        var ability data.ItemAbility
        for mask, current := range abilityMap {
            if abilitiesValue&mask != 0 {
                ability = current
            }
        }

        // Create power
        if amount == 0 {
            continue // Spell Charges
        }

        if powerType != PowerTypeNone {
            power := Power{
                Type: powerType,
                Name: string(name),
                Amount: int(amount),
                Ability: ability,
                Magic: magicType,
                Index: i,
            }
            powers = append(powers, power)
            costs[power] = int(cost)
            compatibilities[power] = *artifactTypes
        }
    }
    return powers, costs, compatibilities, nil
}

func groupPowers(powers []Power, costs map[Power]int, compatibilities map[Power]set.Set[ArtifactType], artifactType ArtifactType, creationType CreationScreen) [][]Power {
    grouped := make(map[PowerType][]Power)
    for _, power := range powers {
        artifactTypes := compatibilities[power]
        cost := costs[power]
        allowed := artifactTypes.Contains(artifactType)
        if creationType == CreationEnchantItem {
            allowed = allowed && cost <= CreationScreenCostThreshold
        }

        if allowed {
            grouped[power.Type] = append(grouped[power.Type], power)
        }
    }

    order := []PowerType{
        PowerTypeAttack,
        PowerTypeDefense,
        PowerTypeToHit,
        PowerTypeMovement,
        PowerTypeResistance,
        PowerTypeSpellSkill,
        PowerTypeSpellSave,
    }

    var result [][]Power
    for _, powerType := range order {
        if grouped[powerType] != nil {
            result = append(result, grouped[powerType])
        }
    }

    return result
}

func getName(artifact *Artifact, customName string) string {
    if customName != "" {
        return customName
    }

    base := artifact.Type.Name()
    if artifact.Type == ArtifactTypeMisc {
        switch {
            case artifact.Image >= 101: base = "Orb"
            case artifact.Image >= 94: base = "Helm"
            case artifact.Image >= 90: base = "Gauntlet"
            case artifact.Image >= 84: base = "Cloak"
            case artifact.Image >= 78: base = "Ring"
            default: base = "Amulet"
        }
    }

    // attack is added as "+X" prefix
    prefix := ""
    attack := artifact.MeleeBonus()
    if attack == 0 {
        attack = artifact.RangedAttackBonus()
    }
    if attack == 0 {
        attack = artifact.MagicAttackBonus()
    }
    if attack != 0 {
        prefix = fmt.Sprintf("+%v ", attack)
    }

    // other powers are added as "of X" postfix (only one)
    postfix := ""
    switch {
        case artifact.HasAbilities(): postfix = fmt.Sprintf(" of %v", artifact.LastAbility().Name())
        case artifact.HasSpellCharges():
            spell, charges := artifact.GetSpellCharge()
            postfix = fmt.Sprintf(" of %v x%v", spell.Name, charges)
        case artifact.HasSpellSavePower(): postfix = " of Power"
        case artifact.HasSpellSkillPower(): postfix = " of Wizardry"
        case artifact.HasResistancePower(): postfix = " of Protection"
        case artifact.HasMovementPower(): postfix = " of Speed"
        case artifact.HasToHitPower(): postfix = " of Accuracy"
        case artifact.HasDefensePower(): postfix = " of Defense"
    }

    return prefix + base + postfix
}

func calculateCost(artifact *Artifact, costs map[Power]int) int {
    base := 0
    switch artifact.Type {
        case ArtifactTypeSword: base = 100
        case ArtifactTypeMace: base = 100
        case ArtifactTypeAxe: base = 100
        case ArtifactTypeBow: base = 100
        case ArtifactTypeStaff: base = 300
        case ArtifactTypeWand: base = 200
        case ArtifactTypeMisc: base = 50
        case ArtifactTypeShield: base = 100
        case ArtifactTypeChain: base = 100
        case ArtifactTypePlate: base = 300
    }

    powerCost := 0
    spellCost := 0
    for _, power := range artifact.Powers {
        if power.Type == PowerTypeSpellCharges {
            spellCost += power.Amount * power.Spell.Cost(false) * 20
        } else {
            powerCost += costs[power]
        }
    }

    // jewelry costs are 2x
    if artifact.Type == ArtifactTypeMisc {
        powerCost *= 2
    }

    cost := base + powerCost + spellCost
    return cost
}

func makePowersFull(ui *uilib.UI, cache *lbx.LbxCache, imageCache *util.ImageCache, fonts ArtifactFonts, picLow int, picHigh int, powerGroups [][]Power, costs map[Power]int, artifact *Artifact, customName *string, selectCount *int) []*uilib.UIElement {
    var elements []*uilib.UIElement

    // image
    artifact.Image = picLow

    elements = append(elements, &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(7), float64(6))

            /*
            image, _ := imageCache.GetImage("items.lbx", artifact.Image, 0)
            screen.DrawImage(image, &options)
            */

            RenderArtifactImage(screen, imageCache, *artifact, ui.Counter / 8, options)
        },
    })

    leftIndex := 0
    leftImages, _ := imageCache.GetImages("spellscr.lbx", 35)
    leftRect := util.ImageRect(5, 24, leftImages[leftIndex])
    elements = append(elements, &uilib.UIElement{
        Rect: leftRect,
        PlaySoundLeftClick: true,
        LeftClick: func(element *uilib.UIElement){
            leftIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            leftIndex = 0
            artifact.Image = artifact.Image - 1
            if artifact.Image < picLow {
                artifact.Image = picHigh
            }
            artifact.Name = getName(artifact, *customName)
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(leftRect.Min.X), float64(leftRect.Min.Y))
            image := leftImages[leftIndex]
            scale.DrawScaled(screen, image, &options)
        },
    })

    rightIndex := 0
    rightImages, _ := imageCache.GetImages("spellscr.lbx", 36)
    rightRect := util.ImageRect(17, 24, leftImages[rightIndex])
    elements = append(elements, &uilib.UIElement{
        Rect: rightRect,
        PlaySoundLeftClick: true,
        LeftClick: func(element *uilib.UIElement){
            rightIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            rightIndex = 0
            artifact.Image = artifact.Image + 1
            if artifact.Image > picHigh {
                artifact.Image = picLow
            }
            artifact.Name = getName(artifact, *customName)
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(rightRect.Min.X), float64(rightRect.Min.Y))
            image := rightImages[rightIndex]
            scale.DrawScaled(screen, image, &options)
        },
    })

    // name field
    nameRect := image.Rect(30, 12, (30 + 130), (12 + fonts.NameFont.Height() + 2))
    nameFocused := false
    artifact.Name = getName(artifact, *customName)
    nameColorSource := ebiten.NewImage(1, 1)
    nameColorSource.Fill(color.RGBA{R: 0xf3, G: 0xb3, B: 0x47, A: 0xff})

    nameEntry := &uilib.UIElement{
        Rect: nameRect,
        GainFocus: func(element *uilib.UIElement){
            nameFocused = true
            ui.FocusElement(element, artifact.Name)
        },
        LoseFocus: func(element *uilib.UIElement){
            nameFocused = false
        },
        TextEntry: func(element *uilib.UIElement, text string) string {
            newName := text
            if len(newName) > 25 {
                newName = newName[0:25]
            }

            if artifact.Name != newName {
                artifact.Name = newName
                *customName = newName
            }

            return newName
        },
        HandleKeys: func(keys []ebiten.Key){
            u := false
            w := false

            for _, key := range keys {
                if key == ebiten.KeyU {
                    u = true
                }

                if key == ebiten.KeyW {
                    w = true
                }

                if key == ebiten.KeyBackspace {
                    if len(artifact.Name) > 0 {
                        artifact.Name = artifact.Name[0:len(artifact.Name) - 1]
                    }
                }
            }

            if ebiten.IsKeyPressed(ebiten.KeyControl) && w {
                for len(artifact.Name) > 0 && artifact.Name[len(artifact.Name) - 1] != ' ' {
                    artifact.Name = artifact.Name[0:len(artifact.Name) - 1]
                }

                for len(artifact.Name) > 0 && artifact.Name[len(artifact.Name) - 1] == ' ' {
                    artifact.Name = artifact.Name[0:len(artifact.Name) - 1]
                }
            }

            if ebiten.IsKeyPressed(ebiten.KeyControl) && u {
                artifact.Name = ""
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            if nameFocused {
                options.ColorScale.SetR(3)
                options.ColorScale.SetG(3)
            }

            fonts.NameFont.PrintOptions(screen, float64(nameRect.Min.X + 1), float64(nameRect.Min.Y + 1), font.FontOptions{Options: &options, Scale: scale.ScaleAmount}, artifact.Name)

            if nameFocused {
                util.DrawTextCursor(screen, nameColorSource, float64(nameRect.Min.X) + 1 + fonts.NameFont.MeasureTextWidth(artifact.Name, 1), float64(nameRect.Min.Y) + 1, ui.Counter)
            }
        },
    }

    elements = append(elements, nameEntry)

    // powers
    x := 7
    y := 40
    printRight := false
    for _, group := range powerGroups {
        groupSelect := -1

        // goto the next column
        if y + (fonts.PowerFont.Height() + 1) * len(group) > data.ScreenHeight - 10 {
            y = 40
            x = 170
            printRight = true
        }

        groupRight := printRight

        var lastPower *Power = nil
        for i, power := range group {
            rect := image.Rect(x, y, x + int(fonts.PowerFont.MeasureTextWidth(power.Name, 1)), y + fonts.PowerFont.Height())
            if groupRight {
                rect = image.Rect(x - int(fonts.PowerFont.MeasureTextWidth(power.Name, 1)), y, x, y + fonts.PowerFont.Height())
            }

            elements = append(elements, &uilib.UIElement{
                Rect: rect,
                PlaySoundLeftClick: true,
                LeftClick: func(element *uilib.UIElement){
                    if groupSelect != -1 {
                        if groupSelect == i {
                            groupSelect = -1
                            *selectCount -= 1
                            artifact.RemovePower(power)
                            artifact.Name = getName(artifact, *customName)
                            artifact.Cost = calculateCost(artifact, costs)
                            lastPower = nil
                        } else {
                            // something was already selected in this group, so the count doesn't change
                            groupSelect = i
                            artifact.RemovePower(*lastPower)
                            artifact.AddPower(power)
                            artifact.Name = getName(artifact, *customName)
                            artifact.Cost = calculateCost(artifact, costs)
                            lastPower = &power
                        }
                    } else {
                        if *selectCount < 4 {
                            *selectCount += 1
                            groupSelect = i
                            artifact.AddPower(power)
                            artifact.Name = getName(artifact, *customName)
                            artifact.Cost = calculateCost(artifact, costs)
                            lastPower = &power
                        } else {
                            ui.AddElement(uilib.MakeErrorElement(ui, cache, imageCache, "Only four powers may be enchanted into an item", func(){}))
                        }

                    }
                },
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    // draw in bright yellow if selected
                    var options ebiten.DrawImageOptions

                    if groupSelect == i {
                        options.ColorScale.SetR(3)
                        options.ColorScale.SetG(3)
                    }

                    if groupRight {
                        fonts.PowerFont.PrintOptions(screen, float64(rect.Max.X), float64(rect.Min.Y), font.FontOptions{Justify: font.FontJustifyRight, Options: &options, Scale: scale.ScaleAmount}, power.Name)
                    } else {
                        fonts.PowerFont.PrintOptions(screen, float64(rect.Min.X), float64(rect.Min.Y), font.FontOptions{Options: &options, Scale: scale.ScaleAmount}, power.Name)
                    }
                },
            })

            y += (fonts.PowerFont.Height() + 1)
        }

        y += 5
    }

    return elements
}

// for choosing a spell to embed in an item with the spell charges power
func makeSpellChoiceElements(ui *uilib.UI, imageCache *util.ImageCache, fonts ArtifactFonts, spells spellbook.Spells, selected *bool, chosen func(spellbook.Spell, int)) []*uilib.UIElement {
    var elements []*uilib.UIElement

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            vector.DrawFilledRect(screen, 0, 0, float32(data.ScreenWidth), float32(data.ScreenHeight), color.RGBA{R: 0, G: 0, B: 0, A: 0x80}, false)

            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(28), float64(12))
            background, _ := imageCache.GetImage("spells.lbx", 0, 0)
            scale.DrawScaled(screen, background, &options)

            // print text "Choose a spell to embed in this item"
            fonts.TitleSpellFont.PrintOptions(screen, float64(data.ScreenWidth / 2), float64(2), font.FontOptions{Scale: scale.ScaleAmount, Justify: font.FontJustifyCenter}, "Choose a spell to embed in this item")
        },
    })

    xLeft := 47
    xRight := 187
    y := 29

    var pages []spellbook.Spells

    // slice up the spells into chunks of 13
    for i, spell := range spells.Spells {
        if i % 13 == 0 {
            pages = append(pages, spellbook.Spells{})
        }

        pages[len(pages) - 1].AddSpell(spell)
    }

    showPage := 0
    spellFont := fonts.SpellFont

    var pageElements []*uilib.UIElement

    shutdown := func(){
        ui.RemoveElements(elements)
        ui.RemoveElements(pageElements)
    }

    makeSelectChargesElements := func (spell spellbook.Spell, x int, picked func(int)) []*uilib.UIElement {
        var moreElements []*uilib.UIElement

        background, _ := imageCache.GetImage("spellscr.lbx", 37, 0)

        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(x), float64(80))

        moreElements = append(moreElements, &uilib.UIElement{
            Layer: 2,
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                scale.DrawScaled(screen, background, &options)

                ax, ay := options.GeoM.Apply(float64(background.Bounds().Dx()) / 2, float64(5))
                spellFont.PrintOptions(screen, ax, ay, font.FontOptions{Scale: scale.ScaleAmount, Justify: font.FontJustifyCenter}, "Charges")
            },
        })

        button0, _ := imageCache.GetImage("spellscr.lbx", 38, 0)

        // 1x, 2x, 3x, 4x
        for i := range 4 {
            buttons, _ := imageCache.GetImages("spellscr.lbx", 38 + i)
            buttonOptions := options
            buttonOptions.GeoM.Translate(float64(button0.Bounds().Dx() * i + 3), float64(14))
            x, y := buttonOptions.GeoM.Apply(0, 0)
            rect := util.ImageRect(int(x), int(y), buttons[0])
            pressed := false
            moreElements = append(moreElements, &uilib.UIElement{
                Layer: 2,
                Rect: rect,
                LeftClick: func(element *uilib.UIElement){
                    pressed = true
                },
                LeftClickRelease: func(element *uilib.UIElement){
                    pressed = false
                    ui.RemoveElements(moreElements)
                    picked(i+1)
                },
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    if pressed {
                        scale.DrawScaled(screen, buttons[1], &buttonOptions)
                    } else {
                        scale.DrawScaled(screen, buttons[0], &buttonOptions)
                    }
                },
            })
        }

        return moreElements
    }

    makePageElements := func () []*uilib.UIElement {
        ui.RemoveElements(pageElements)

        pageElements = nil
        // left page
        for i, spell := range pages[showPage].Spells {
            yPos := y + (spellFont.Height() + 1) * i
            rect := image.Rect(xLeft, yPos, xLeft + int(spellFont.MeasureTextWidth(spell.Name, 1)), yPos + spellFont.Height())

            pickCharges := func (count int) {
                shutdown()
                chosen(spell, count)
            }

            pageElements = append(pageElements, &uilib.UIElement{
                Layer: 1,
                Rect: rect,
                LeftClick: func(element *uilib.UIElement){
                    ui.AddElements(makeSelectChargesElements(spell, xRight, pickCharges))
                },
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    spellFont.PrintOptions(screen, float64(xLeft), float64(yPos), font.FontOptions{Scale: scale.ScaleAmount}, spell.Name)
                },
            })
        }

        // right page
        if showPage + 1 < len(pages) {
            for i, spell := range pages[showPage+1].Spells {
                yPos := y + (spellFont.Height() + 1) * i
                rect := image.Rect(xRight, yPos, xRight + int(spellFont.MeasureTextWidth(spell.Name, 1)), yPos + spellFont.Height())

                pickCharges := func (count int) {
                    shutdown()
                    chosen(spell, count)
                }

                pageElements = append(pageElements, &uilib.UIElement{
                    Layer: 1,
                    Rect: rect,
                    LeftClick: func(element *uilib.UIElement){
                        ui.AddElements(makeSelectChargesElements(spell, xLeft, pickCharges))
                    },
                    Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                        spellFont.PrintOptions(screen, float64(xRight), float64(yPos), font.FontOptions{Scale: scale.ScaleAmount}, spell.Name)
                    },
                })
            }
        }

        return pageElements
    }

    elements = append(elements, makePageElements()...)

    if len(pages) > 2 {
        // TODO: use cool page flip animation from the spellbook code

        // left dogear
        leftEar, _ := imageCache.GetImage("spells.lbx", 1, 0)
        leftRect := util.ImageRect(41, 16, leftEar)
        elements = append(elements, &uilib.UIElement{
            Layer: 1,
            Rect: leftRect,
            LeftClick: func(element *uilib.UIElement){
                if showPage - 2 >= 0 {
                    showPage -= 2
                    ui.AddElements(makePageElements())
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(41, 16)
                scale.DrawScaled(screen, leftEar, &options)
            },
        })

        // right dogear
        rightEar, _ := imageCache.GetImage("spells.lbx", 2, 0)
        rightRect := util.ImageRect(286, 16, rightEar)
        elements = append(elements, &uilib.UIElement{
            Layer: 1,
            Rect: rightRect,
            LeftClick: func(element *uilib.UIElement){
                if showPage + 2 < len(pages) {
                    showPage += 2
                    ui.AddElements(makePageElements())
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(rightRect.Min.X), float64(rightRect.Min.Y))
                scale.DrawScaled(screen, rightEar, &options)
            },
        })
    }

    // X button at bottom to cancel
    cancelRect := image.Rect(0, 0, 18, 24).Add(image.Pt(188, 172))
    elements = append(elements, &uilib.UIElement{
        Rect: cancelRect,
        Layer: 1,
        LeftClick: func(this *uilib.UIElement){
            shutdown()
            *selected = false
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            // vector.StrokeRect(screen, float32(cancelRect.Min.X), float32(cancelRect.Min.Y), float32(cancelRect.Dx()), float32(cancelRect.Dy()), 1, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
        },
    })

    return elements
}

func makeAbilityElements(ui *uilib.UI, cache *lbx.LbxCache, imageCache *util.ImageCache, artifact *Artifact, customName *string, fonts ArtifactFonts, powers []Power, compatibilities map[Power]set.Set[ArtifactType], costs map[Power]int, selectCount *int, wizard spellbook.Wizard, availableSpells spellbook.Spells) []*uilib.UIElement {
    var elements []*uilib.UIElement

    var group1 []Power
    var group2 []Power
    var group3 []Power

    for _, power := range powers {
        switch power.Type {
            case PowerTypeAbility1: group1 = append(group1, power)
            case PowerTypeAbility2: group2 = append(group2, power)
            case PowerTypeAbility3: group3 = append(group3, power)
        }
    }

    /*
    minItem := 0
    */
    maxItem := 11

    // currentItem := 0
    y := 39
    x := 200

    // true if the rect is within the bounds of where the abilities should be
    inBounds := func (rect image.Rectangle) bool {
        if rect.Min.Y >= 39 && rect.Min.Y <= 160 {
            return true
        }

        return false
    }

    totalItems := 0

    for groupNum, group := range [][]Power{group1, group2, group3} {
        groupSelect := -1

        mutuallyExclusive := groupNum == 0 || groupNum == 1

        slices.SortFunc(group, func(a, b Power) int {
            return cmp.Compare(a.Index, b.Index)
        })

        var lastPower *Power = nil
        selected := make([]bool, len(group))
        for i, power := range group {
            artifactTypes := compatibilities[power]
            if artifactTypes.Contains(artifact.Type) && wizard.MagicLevel(power.Magic) >= power.Amount {
                totalItems += 1
                xRect := image.Rect(x, y, x + int(fonts.PowerFont.MeasureTextWidth(power.Name, 1)), y + fonts.PowerFont.Height())
                elements = append(elements, &uilib.UIElement{
                    Rect: xRect,
                    PlaySoundLeftClick: inBounds(xRect),
                    LeftClick: func(element *uilib.UIElement){
                        if !inBounds(element.Rect) {
                            return
                        }

                        if mutuallyExclusive {
                            // can only pick one in the group
                            if groupSelect != -1 {
                                if groupSelect == i {
                                    groupSelect = -1
                                    *selectCount -= 1
                                    artifact.RemovePower(power)
                                    artifact.Name = getName(artifact, *customName)
                                    artifact.Cost = calculateCost(artifact, costs)
                                    lastPower = nil
                                } else {
                                    // something was already selected in this group, so the count doesn't change
                                    groupSelect = i
                                    artifact.RemovePower(*lastPower)
                                    artifact.AddPower(power)
                                    artifact.Name = getName(artifact, *customName)
                                    artifact.Cost = calculateCost(artifact, costs)
                                    lastPower = &power
                                }
                            } else {
                                if *selectCount < 4 {
                                    *selectCount += 1
                                    groupSelect = i
                                    artifact.AddPower(power)
                                    artifact.Name = getName(artifact, *customName)
                                    artifact.Cost = calculateCost(artifact, costs)
                                    lastPower = &power
                                } else {
                                    ui.AddElement(uilib.MakeErrorElement(ui, cache, imageCache, "Only four powers may be enchanted into an item", func(){}))
                                }

                            }
                        } else {
                            // can pick multiple
                            if selected[i] {
                                artifact.RemovePower(power)
                                artifact.Name = getName(artifact, *customName)
                                artifact.Cost = calculateCost(artifact, costs)
                                selected[i] = false
                                *selectCount -= 1
                            } else if *selectCount < 4 {
                                *selectCount += 1
                                selected[i] = true

                                artifact.AddPower(power)
                                artifact.Name = getName(artifact, *customName)
                                artifact.Cost = calculateCost(artifact, costs)
                            } else {
                                ui.AddElement(uilib.MakeErrorElement(ui, cache, imageCache, "Only four powers may be enchanted into an item", func(){}))
                            }
                        }
                    },
                    Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                        if !inBounds(element.Rect) {
                            return
                        }

                        var options ebiten.DrawImageOptions

                        if (mutuallyExclusive && groupSelect == i) || (!mutuallyExclusive && selected[i]) {
                            options.ColorScale.SetR(3)
                            options.ColorScale.SetG(3)
                        }

                        fonts.PowerFont.PrintOptions(screen, float64(element.Rect.Min.X), float64(element.Rect.Min.Y), font.FontOptions{Scale: scale.ScaleAmount, Options: &options}, power.Name)
                    },
                })

                y += (fonts.PowerFont.Height() + 1)
            }
        }

        y += 5
    }

    // spell charges
    if len(availableSpells.Spells) > 0 && (artifact.Type == ArtifactTypeWand || artifact.Type == ArtifactTypeStaff) {
        xRect := image.Rect(x, y, x + int(fonts.PowerFont.MeasureTextWidth("Spell Charges", 1)), y + fonts.PowerFont.Height())
        selected := false
        totalItems += 1

        addedPower := Power{Type: PowerTypeNone}

        elements = append(elements, &uilib.UIElement{
            Rect: xRect,
            PlaySoundLeftClick: inBounds(xRect),
            LeftClick: func(element *uilib.UIElement){
                if !inBounds(element.Rect) {
                    return
                }

                selected = !selected

                if selected {
                    ui.AddElements(makeSpellChoiceElements(ui, imageCache, fonts, availableSpells, &selected, func (spell spellbook.Spell, charges int){
                        addedPower = Power{
                            Type: PowerTypeSpellCharges,
                            Spell: spell,
                            Amount: charges,
                            Name: fmt.Sprintf("%v x%v", spell.Name, charges),
                        }
                        artifact.AddPower(addedPower)
                        artifact.Name = getName(artifact, *customName)
                        artifact.Cost = calculateCost(artifact, costs)
                    }))
                } else {
                    artifact.RemovePower(addedPower)
                    addedPower = Power{Type: PowerTypeNone}
                    artifact.Name = getName(artifact, *customName)
                    artifact.Cost = calculateCost(artifact, costs)
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                if !inBounds(element.Rect) {
                    return
                }

                var options ebiten.DrawImageOptions

                if selected {
                    options.ColorScale.SetR(3)
                    options.ColorScale.SetG(3)
                }

                if addedPower.Type == PowerTypeSpellCharges {
                    fonts.PowerFont.PrintOptions(screen, float64(element.Rect.Min.X), float64(element.Rect.Min.Y), font.FontOptions{Scale: scale.ScaleAmount, Options: &options}, addedPower.Name)
                } else {
                    fonts.PowerFont.PrintOptions(screen, float64(element.Rect.Min.X), float64(element.Rect.Min.Y), font.FontOptions{Scale: scale.ScaleAmount, Options: &options}, "Spell Charges")
                }
            },
        })
    }

    // show up/down scroll arrows if there are too many abilities to choose
    if totalItems > maxItem {
        upArrows, _ := imageCache.GetImages("spellscr.lbx", 43)
        downArrows, _ := imageCache.GetImages("spellscr.lbx", 44)

        abilityElements := slices.Clone(elements)

        minItem := 0

        // scrolling is a slight hack in that the elements always exist but they are only displayed
        // if their Rect field is within the bounds of the scroll window.
        // scrolling mutates the Y values of the Rect fields of the elements, so that scrolling up means
        // all Y values go up by some value, and scrolling down means all Y values go down by some value.
        // The play sound has to be manually enabled/disabled based on whether the element is in view

        doScroll := func (direction int) {
            move := direction * fonts.PowerFont.Height()
            for _, element := range abilityElements {
                element.Rect.Min.Y += move
                element.Rect.Max.Y += move

                element.PlaySoundLeftClick = inBounds(element.Rect)
            }
        }

        scrollUp := func() {
            if minItem > 0 {
                doScroll(1)
                minItem -= 1
            }
        }

        scrollDown := func() {
            if minItem < totalItems - maxItem {
                doScroll(-1)
                minItem += 1
            }
        }

        upX := 308
        upY := 43
        upPressed := false
        elements = append(elements, &uilib.UIElement{
            Rect: util.ImageRect(upX, upY, upArrows[0]),
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(upX), float64(upY))

                var image *ebiten.Image
                if upPressed {
                    image = upArrows[1]
                } else {
                    image = upArrows[0]
                }

                scale.DrawScaled(screen, image, &options)
            },
            PlaySoundLeftClick: true,
            LeftClick: func(element *uilib.UIElement){
                upPressed = true
            },
            LeftClickRelease: func(element *uilib.UIElement){
                upPressed = false

                scrollUp()
            },
        })

        downX := upX
        downY := 165
        downPressed := false
        elements = append(elements, &uilib.UIElement{
            Rect: util.ImageRect(downX, downY, downArrows[0]),
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(downX), float64(downY))

                var image *ebiten.Image
                if downPressed {
                    image = downArrows[1]
                } else {
                    image = downArrows[0]
                }
                scale.DrawScaled(screen, image, &options)
            },
            PlaySoundLeftClick: true,
            LeftClick: func(element *uilib.UIElement){
                downPressed = true
            },
            LeftClickRelease: func(element *uilib.UIElement){
                downPressed = false
                scrollDown()
            },
        })

        elements = append(elements, &uilib.UIElement{
            Rect: image.Rect(200, 39, (200 + 110), 170),
            Scroll: func (element *uilib.UIElement, x float64, y float64) {
                if y < 0 {
                    scrollDown()
                } else if y > 0 {
                    scrollUp()
                }
            },
        })
    }

    return elements
}

type ArtifactFonts struct {
    PowerFont *font.Font
    PowerFontWhite *font.Font
    NameFont *font.Font
    TitleSpellFont *font.Font
    SpellFont *font.Font
}

func makeFonts(cache *lbx.LbxCache) ArtifactFonts {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return ArtifactFonts{}
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return ArtifactFonts{}
    }

    // solid := util.Lighten(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, -40)
    solid := util.Lighten(color.RGBA{R: 0xca, G: 0x8a, B: 0x4a, A: 0xff}, -10)

    palette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        solid, solid, solid,
        solid, solid, solid,
        solid, solid, solid,
    }

    powerFont := font.MakeOptimizedFontWithPalette(fonts[3], palette)

    grey := util.Lighten(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, -40)
    greyPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        grey, grey, grey,
        grey, grey, grey,
    }

    nameFont := font.MakeOptimizedFontWithPalette(fonts[1], greyPalette)

    powerFontWhite := font.MakeOptimizedFontWithPalette(fonts[3], greyPalette)

    orange := util.Lighten(color.RGBA{R: 0xe5, G: 0x7b, B: 0x12, A: 0xff}, 10)
    // red := color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}
    titlePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        util.Lighten(orange, 20),
        util.Lighten(orange, 12),
        util.Lighten(orange, 8),
        orange,
    }

    titleSpellFont := font.MakeOptimizedFontWithPalette(fonts[4], titlePalette)

    darkRed := color.RGBA{R: 0x6d, G: 0x09, B: 0x0c, A: 0xff}

    spellPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        darkRed, darkRed, darkRed,
        darkRed, darkRed, darkRed,
    }

    spellFont := font.MakeOptimizedFontWithPalette(fonts[3], spellPalette)

    return ArtifactFonts{
        PowerFont: powerFont,
        PowerFontWhite: powerFontWhite,
        NameFont: nameFont,
        TitleSpellFont: titleSpellFont,
        SpellFont: spellFont,
    }
}

/* returns the artifact that was created and true,
 * otherwise false for cancelled
 */
func ShowCreateArtifactScreen(yield coroutine.YieldFunc, cache *lbx.LbxCache, creationType CreationScreen, wizard spellbook.Wizard, availableSpells spellbook.Spells, draw *func(*ebiten.Image)) (*Artifact, bool) {
    fonts := makeFonts(cache)

    imageCache := util.MakeImageCache(cache)

    ui := &uilib.UI{
        Cache: cache,
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            background, _ := imageCache.GetImage("spellscr.lbx", 13, 0)
            scale.DrawScaled(screen, background, &options)

            ui.StandardDraw(screen)
        },
    }

    ui.SetElementsFromArray(nil)

    var currentArtifact *Artifact
    var customName string

    type PowerArtifact struct {
        Elements []*uilib.UIElement
        AbilityElements []*uilib.UIElement
        Artifact *Artifact
    }

    // ui elements for powers that can be selected, based on what item is selected
    powers := make(map[ArtifactType]PowerArtifact)

    // manually curry
    makePowers := func(picLow int, picHigh int, artifactType ArtifactType, powers []Power, costs map[Power]int, compatibilities map[Power]set.Set[ArtifactType]) PowerArtifact {
        var artifact Artifact
        artifact.Type = artifactType
        groups := groupPowers(powers, costs, compatibilities, artifactType, creationType)
        selectCount := 0
        elements := makePowersFull(ui, cache, &imageCache, fonts, picLow, picHigh, groups, costs, &artifact, &customName, &selectCount)
        abilityElements := makeAbilityElements(ui, cache, &imageCache, &artifact, &customName, fonts, powers, compatibilities, costs, &selectCount, wizard, availableSpells)
        return PowerArtifact{
            Elements: elements,
            AbilityElements: abilityElements,
            Artifact: &artifact,
        }
    }

    powerEntries, costs, compatibilities, err := ReadPowers(cache)
    if err != nil {
        return nil, true
    }

    powers[ArtifactTypeSword] = makePowers(0, 8, ArtifactTypeSword, powerEntries, costs, compatibilities)
    powers[ArtifactTypeMace] = makePowers(9, 19, ArtifactTypeMace, powerEntries, costs, compatibilities)
    powers[ArtifactTypeAxe] = makePowers(20, 28, ArtifactTypeAxe, powerEntries, costs, compatibilities)
    powers[ArtifactTypeBow] = makePowers(29, 37, ArtifactTypeBow, powerEntries, costs, compatibilities)
    powers[ArtifactTypeStaff] = makePowers(38, 46, ArtifactTypeStaff, powerEntries, costs, compatibilities)
    powers[ArtifactTypeWand] = makePowers(107, 115, ArtifactTypeWand, powerEntries, costs, compatibilities)
    powers[ArtifactTypeMisc] = makePowers(72, 106, ArtifactTypeMisc, powerEntries, costs, compatibilities)
    powers[ArtifactTypeShield] = makePowers(62, 71, ArtifactTypeShield, powerEntries, costs, compatibilities)
    powers[ArtifactTypeChain] = makePowers(47, 54, ArtifactTypeChain, powerEntries, costs, compatibilities)
    powers[ArtifactTypePlate] = makePowers(55, 61, ArtifactTypePlate,  powerEntries, costs, compatibilities)

    updatePowers := func(index ArtifactType){
        for _, each := range powers {
            ui.RemoveElements(each.Elements)
            ui.RemoveElements(each.AbilityElements)
        }

        ui.AddElements(powers[index].Elements)
        ui.AddElements(powers[index].AbilityElements)
        currentArtifact = powers[index].Artifact
        currentArtifact.Name = getName(currentArtifact, customName)
        currentArtifact.Cost = calculateCost(currentArtifact, costs)
    }

    var selectedButton *uilib.UIElement

    makeButton := func(x int, y int, unselected int, selected int, item ArtifactType) *uilib.UIElement {
        index := 0
        imageRect, _ := imageCache.GetImage("spellscr.lbx", unselected, 0)
        rect := util.ImageRect(x, y, imageRect)
        return &uilib.UIElement{
            Rect: rect,
            PlaySoundLeftClick: true,
            LeftClick: func(element *uilib.UIElement){
                index = 1
            },
            LeftClickRelease: func(element *uilib.UIElement){
                selectedButton = element
                index = 0

                updatePowers(item)
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
                use := unselected
                if selectedButton == element {
                    use = selected
                }
                image, _ := imageCache.GetImage("spellscr.lbx", use, index)
                scale.DrawScaled(screen, image, &options)
            },
        }
    }

    unselectedImageStart := 14
    selectedImageStart := 25
    // to get dimensions
    tmpImage, _ := imageCache.GetImage("spellscr.lbx", unselectedImageStart, 0)

    // 10 item types
    for i := 0; i < 10; i++ {
        x := 156 + (i % 5) * (tmpImage.Bounds().Dx() + 2)
        y := 3 + (i / 5) * (tmpImage.Bounds().Dy() + 2)

        button := makeButton(x, y, unselectedImageStart + i, selectedImageStart + i, ArtifactType(i+1))
        if selectedButton == nil {
            selectedButton = button
            updatePowers(ArtifactType(i+1))
        }

        ui.AddElement(button)
    }

    // the final artifact cost can be cached because the final cost is only a function of the base artifact cost
    artifactCost := functional.Memoize(func (cost int) int {
        spellName := ""
        switch creationType {
            case CreationEnchantItem: spellName = "Enchant Item"
            case CreationCreateArtifact: spellName = "Create Artifact"
        }
        spell := spellbook.Spell{
            Name: spellName,
            OverrideCost: currentArtifact.Cost,
            Magic: data.ArcaneMagic,
            Eligibility: spellbook.EligibilityOverlandOnly,
            Section: spellbook.SectionSpecial,
        }

        return spellbook.ComputeSpellCost(wizard, spell, false, false)
    })

    ui.AddElement(&uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            fonts.PowerFontWhite.PrintOptions(screen, 198, 185, font.FontOptions{Scale: scale.ScaleAmount}, fmt.Sprintf("Cost: %v", artifactCost(currentArtifact.Cost)))
        },
    })

    quit := false
    canceled := false

    okButtons, _ := imageCache.GetImages("spellscr.lbx", 24)
    okIndex := 0
    okRect := util.ImageRect(281, 180, okButtons[0])
    ui.AddElement(&uilib.UIElement{
        Rect: okRect,
        PlaySoundLeftClick: true,
        LeftClick: func(element *uilib.UIElement){
            okIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            okIndex = 0

            if strings.TrimSpace(currentArtifact.Name) == "" {
                ui.AddElement(uilib.MakeErrorElement(ui, cache, &imageCache, "An item must have a name", func(){}))
                return
            }

            if len(currentArtifact.Powers) == 0 {
                group := uilib.MakeGroup()
                yes := func(){
                    quit = true
                    canceled = true
                    ui.RemoveGroup(group)
                }

                no := func(){
                    ui.RemoveGroup(group)
                }

                group.AddElements(uilib.MakeConfirmDialog(group, cache, &imageCache, "This item has no powers. Do you wish to abort the enchantment?", true, yes, no))
                ui.AddGroup(group)
                return
            }

            quit = true
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(okRect.Min.X), float64(okRect.Min.Y))
            image := okButtons[okIndex]
            scale.DrawScaled(screen, image, &options)
        },
    })

    *draw = func(screen *ebiten.Image) {
        ui.Draw(ui, screen)
    }

    for !quit {
        ui.StandardUpdate()
        yield()
    }

    ui.UnfocusElement()

    return currentArtifact, canceled
}
