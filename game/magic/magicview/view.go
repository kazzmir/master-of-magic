package magicview

import (
    "fmt"
    "image"
    "cmp"
    "math"
    "math/rand/v2"
    "slices"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/mirror"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    helplib "github.com/kazzmir/master-of-magic/game/magic/help"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    "github.com/hajimehoshi/ebiten/v2"
    // "github.com/hajimehoshi/ebiten/v2/vector"
)

type MagicScreenState int

const (
    MagicScreenStateRunning MagicScreenState = iota
    MagicScreenStateDone
)

type MagicScreen struct {
    Cache *lbx.LbxCache
    ImageCache util.ImageCache

    UI *uilib.UI
    State MagicScreenState

    Power int

    ManaLocked bool
    ResearchLocked bool
    SkillLocked bool
}

func MakeMagicScreen(cache *lbx.LbxCache, player *playerlib.Player, enemies []*playerlib.Player, power int) *MagicScreen {
    magic := &MagicScreen{
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
        State: MagicScreenStateRunning,

        Power: power,

        ManaLocked: false,
        ResearchLocked: false,
        SkillLocked: false,
    }

    magic.UI = magic.MakeUI(player, enemies)

    return magic
}

func MakeTransmuteElements(ui *uilib.UI, smallFont *font.Font, player *playerlib.Player, help *helplib.Help, cache *lbx.LbxCache, imageCache *util.ImageCache) *uilib.UIElementGroup {
    group := uilib.MakeGroup()

    fontOptions := font.FontOptions{Justify: font.FontJustifyRight, DropShadow: true, Scale: scale.ScaleAmount}

    ok, _ := imageCache.GetImages("magic.lbx", 54)
    okRect := util.ImageRect(176, 99, ok[0])
    okIndex := 0

    arrowRight, _ := imageCache.GetImage("magic.lbx", 55, 0)
    arrowLeft, _ := imageCache.GetImage("magic.lbx", 56, 0)

    // if true, then convert power to gold, otherwise convert gold to power
    isRight := true

    arrowRect := util.ImageRect(146, 99, arrowRight)

    cancel, _ := imageCache.GetImages("magic.lbx", 53)
    cancelRect := util.ImageRect(93, 99, cancel[0])
    cancelIndex := 0

    powerToGold, _ := imageCache.GetImage("magic.lbx", 59, 0)

    totalGold := player.Gold
    totalMana := player.Mana

    alchemyConversion := 0.5
    if player.Wizard.RetortEnabled(data.RetortAlchemy) {
        alchemyConversion = 1
    }

    conveyor, _ := imageCache.GetImage("magic.lbx", 57, 0)
    cursor, _ := imageCache.GetImage("magic.lbx", 58, 0)

    cursorPosition := 0
    changePercent := float64(0)

    group.AddElement(&uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("magic.lbx", 52, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(75, 60)
            scale.DrawScaled(screen, background, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(okRect.Min.X), float64(okRect.Min.Y))
            scale.DrawScaled(screen, ok[okIndex], &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(arrowRect.Min.X), float64(arrowRect.Min.Y))

            flip := false
            if isRight {
                scale.DrawScaled(screen, arrowRight, &options)

                options.GeoM.Reset()
                options.GeoM.Translate(87, 70)
                scale.DrawScaled(screen, powerToGold, &options)

                flip = true
            } else {
                scale.DrawScaled(screen, arrowLeft, &options)
            }

            // the conveyor belt should only be drawn until the position of the cursor
            conveyorArea := screen.SubImage(scale.ScaleRect(image.Rect(131, 85, 131 + cursorPosition, (85 + 7)))).(*ebiten.Image)

            // draw an animated conveyor belt by drawing the same image twice with a slight offset
            movement := int((ui.Counter / 4) % 8)
            if flip {
                options.GeoM.Reset()
                options.GeoM.Scale(-1, 1)
                options.GeoM.Translate(float64(131), float64(85))
                options.GeoM.Translate(float64(cursorPosition + conveyor.Bounds().Dx()), 0)
                options.GeoM.Translate(float64(-movement), 0)
                scale.DrawScaled(conveyorArea, conveyor, &options)

                options.GeoM.Reset()
                options.GeoM.Scale(-1, 1)
                options.GeoM.Translate(float64(131), float64(85))
                options.GeoM.Translate(float64(cursorPosition), 0)
                options.GeoM.Translate(float64(-movement), 0)
                scale.DrawScaled(conveyorArea, conveyor, &options)
            } else {
                options.GeoM.Reset()
                options.GeoM.Translate(float64(131), float64(85))
                options.GeoM.Translate(float64(-conveyor.Bounds().Dx()), 0)
                options.GeoM.Translate(float64(movement), 0)
                scale.DrawScaled(conveyorArea, conveyor, &options)

                options.GeoM.Reset()
                options.GeoM.Translate(float64(131), float64(85))
                options.GeoM.Translate(float64(movement), 0)
                scale.DrawScaled(conveyorArea, conveyor, &options)
            }

            // draw the cursor itself
            options.GeoM.Reset()
            options.GeoM.Translate(float64((131 - 3)) + float64(cursorPosition), float64(85))
            scale.DrawScaled(screen, cursor, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(cancelRect.Min.X), float64(cancelRect.Min.Y))
            scale.DrawScaled(screen, cancel[cancelIndex], &options)

            leftSide := float64(122)
            rightSide := float64(224)
            if isRight {
                smallFont.PrintOptions(screen, leftSide, float64(86), fontOptions, fmt.Sprintf("%v GP", int(float64(totalMana) * changePercent * alchemyConversion)))
                smallFont.PrintOptions(screen, rightSide, float64(86), fontOptions, fmt.Sprintf("%v PP", int(float64(totalMana) * changePercent)))
            } else {
                smallFont.PrintOptions(screen, leftSide, float64(86), fontOptions, fmt.Sprintf("%v GP", int(float64(totalGold) * changePercent)))
                smallFont.PrintOptions(screen, rightSide, float64(86), fontOptions, fmt.Sprintf("%v PP", int(float64(totalGold) * changePercent * alchemyConversion)))
            }
        },
    })

    {
        posX := 0
        conveyorRect := image.Rect(131, 85, (131 + 56), (85 + 7))
        group.AddElement(&uilib.UIElement{
            Rect: conveyorRect,
            Layer: 1,
            PlaySoundLeftClick: true,
            RightClick: func(element *uilib.UIElement){
                helpEntries := help.GetEntriesByName("Alchemy Ratio")
                if helpEntries != nil {
                    group.AddElement(uilib.MakeHelpElementWithLayer(group, cache, imageCache, 2, helpEntries[0], helpEntries[1:]...))
                }
            },
            LeftClick: func(element *uilib.UIElement){
                cursorPosition = posX
                if cursorPosition < 0 {
                    cursorPosition = 0
                }
                changePercent = float64(cursorPosition) / float64(conveyorRect.Dx())
            },
            Inside: func(element *uilib.UIElement, x, y int){
                posX = x
            },
        })
    }

    group.AddElement(&uilib.UIElement{
        Rect: image.Rect(94, 85, 123, 92),
        Layer: 1,
        RightClick: func(element *uilib.UIElement){
            helpEntries := help.GetEntriesByName("Alchemy Gold")
            if helpEntries != nil {
                group.AddElement(uilib.MakeHelpElementWithLayer(group, cache, imageCache, 2, helpEntries[0], helpEntries[1:]...))
            }
        },
    })

    group.AddElement(&uilib.UIElement{
        Rect: image.Rect(195, 85, 225, 92),
        Layer: 1,
        RightClick: func(element *uilib.UIElement){
            helpEntries := help.GetEntriesByName("Alchemy Power")
            if helpEntries != nil {
                group.AddElement(uilib.MakeHelpElementWithLayer(group, cache, imageCache, 2, helpEntries[0], helpEntries[1:]...))
            }
        },
    })

    // cancel button
    group.AddElement(&uilib.UIElement{
        Rect: cancelRect,
        Layer: 1,
        PlaySoundLeftClick: true,
        LeftClick: func(element *uilib.UIElement){
            cancelIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            cancelIndex = 0
            ui.RemoveGroup(group)
        },
        RightClick: func(element *uilib.UIElement){
            helpEntries := help.GetEntries(371)
            if helpEntries != nil {
                group.AddElement(uilib.MakeHelpElementWithLayer(group, cache, imageCache, 2, helpEntries[0], helpEntries[1:]...))
            }
        },
    })

    // arrow button
    group.AddElement(&uilib.UIElement{
        Rect: arrowRect,
        Layer: 1,
        PlaySoundLeftClick: true,
        LeftClick: func(element *uilib.UIElement){
            isRight = !isRight
        },
        RightClick: func(element *uilib.UIElement){
            helpEntries := help.GetEntries(372)
            if helpEntries != nil {
                group.AddElement(uilib.MakeHelpElementWithLayer(group, cache, imageCache, 2, helpEntries[0], helpEntries[1:]...))
            }
        },
    })

    // ok button
    group.AddElement(&uilib.UIElement{
        Rect: okRect,
        Layer: 1,
        PlaySoundLeftClick: true,
        LeftClick: func(element *uilib.UIElement){
            okIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            okIndex = 0
            ui.RemoveGroup(group)

            goldChange := 0
            manaChange := 0

            if isRight {
                goldChange = int(float64(totalMana) * changePercent * alchemyConversion)
                manaChange = -int(float64(totalMana) * changePercent)
            } else {
                goldChange = -int(float64(totalGold) * changePercent)
                manaChange = int(float64(totalGold) * changePercent * alchemyConversion)
            }

            player.Gold += goldChange
            player.Mana += manaChange

        },
        RightClick: func(element *uilib.UIElement){
            helpEntries := help.GetEntries(373)
            if helpEntries != nil {
                group.AddElement(uilib.MakeHelpElementWithLayer(group, cache, imageCache, 2, helpEntries[0], helpEntries[1:]...))
            }
        },

    })

    return group
}

// FIXME: move this into player
func randomizeBookOrder(books int) []int {
    order := make([]int, books)
    for i := 0; i < books; i++ {
        order[i] = rand.IntN(3)
    }
    return order
}

func (magic *MagicScreen) MakeUI(player *playerlib.Player, enemies []*playerlib.Player) *uilib.UI {
    knownPlayers := player.GetKnownPlayers()

    fonts := fontslib.MakeMagicViewFonts(magic.Cache)
    leftShadow := font.FontOptions{Justify: font.FontJustifyLeft, DropShadow: true, Scale: scale.ScaleAmount}
    centerShadow := font.FontOptions{Justify: font.FontJustifyCenter, DropShadow: true, Scale: scale.ScaleAmount}
    rightShadow := font.FontOptions{Justify: font.FontJustifyRight, DropShadow: true, Scale: scale.ScaleAmount}

    helpLbx, err := magic.Cache.GetLbxFile("help.lbx")
    if err != nil {
        return nil
    }

    help, err := helplib.ReadHelp(helpLbx, 2)
    if err != nil {
        return nil
    }

    ui := &uilib.UI{
        Cache: magic.Cache,
        Draw: func(ui *uilib.UI, screen *ebiten.Image) {
            background, err := magic.ImageCache.GetImage("magic.lbx", 0, 0)
            if err == nil {
                var options ebiten.DrawImageOptions
                scale.DrawScaled(screen, background, &options)
            }

            mana := int(math.Round(player.PowerDistribution.Mana * float64(magic.Power)))
            research := int(math.Round(player.PowerDistribution.Research * float64(magic.Power)))
            skill := magic.Power - (mana + research)

            fonts.NormalFont.PrintOptions(screen, float64(56), float64(160), rightShadow, fmt.Sprintf("%v MP", mana))
            fonts.NormalFont.PrintOptions(screen, float64(103), float64(160), rightShadow, fmt.Sprintf("%v RP", research))
            fonts.NormalFont.PrintOptions(screen, float64(151), float64(160), rightShadow, fmt.Sprintf("%v SP", skill))

            ui.StandardDraw(screen)
        },
    }

    group := uilib.MakeGroup()

    // gems with wizard info
    for i := range 4 {

        gemPositions := []image.Point{
            image.Pt(24, 4),
            image.Pt(101, 4),
            image.Pt(178, 4),
            image.Pt(255, 4),
        }

        // the treaty icon is the scroll/peace/war icon between the wizard being rendered and another wizard
        // each enemy wizard can have a treaty with any other wizard
        getTreatyIcon := func (otherPlayer *playerlib.Player, treaty data.TreatyType) *ebiten.Image {
            if treaty == data.TreatyNone {
                return nil
            }

            // only show treaties for other wizards that the main player already knows about
            if otherPlayer == player || slices.Contains(knownPlayers, otherPlayer) {
                base := 0
                // show the treaty icon in the color of the other player
                switch otherPlayer.GetBanner() {
                    case data.BannerBlue: base = 60
                    case data.BannerGreen: base = 63
                    case data.BannerPurple: base = 66
                    case data.BannerRed: base = 69
                    case data.BannerYellow: base = 72
                }

                switch treaty {
                    case data.TreatyPact: img, _ := magic.ImageCache.GetImage("magic.lbx", base + 0, 0); return img
                    case data.TreatyAlliance: img, _ := magic.ImageCache.GetImage("magic.lbx", base + 1, 0); return img
                    case data.TreatyWar: img, _ := magic.ImageCache.GetImage("magic.lbx", base + 2, 0); return img
                }
            }

            return nil
        }

        gemDefeated, _ := magic.ImageCache.GetImage("magic.lbx", 51, 0)
        gemUnknown, _ := magic.ImageCache.GetImage("magic.lbx", 6, 0)
        position := gemPositions[i]
        rect := util.ImageRect(position.X, position.Y, gemUnknown)
        group.AddElement(&uilib.UIElement{
            Rect: rect,
            LeftClick: func(element *uilib.UIElement){
                // show diplomatic dialogue screen
            },
            RightClick: func(element *uilib.UIElement){
                // show mirror ui with extra enemy info: relations, treaties, personality, objective

                if i < len(knownPlayers) && !knownPlayers[i].Defeated {
                    mirrorElement := mirror.MakeMirrorUI(magic.Cache, knownPlayers[i], ui)
                    ui.AddElement(mirrorElement)
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))

                // FIXME: should defeated players been shown regardless of if they are known?
                if i < len(knownPlayers) {
                    enemy := knownPlayers[i]

                    if enemy.Defeated {
                        scale.DrawScaled(screen, gemDefeated, &options)
                    } else {
                        portraitIndex := mirror.GetWizardPortraitIndex(enemy.Wizard.Base, enemy.Wizard.Banner)
                        portrait, _ := magic.ImageCache.GetImage("lilwiz.lbx", portraitIndex, 0)
                        if portrait != nil {
                            scale.DrawScaled(screen, portrait, &options)
                        }
                        if player.GlobalEnchantments.Contains(data.EnchantmentDetectMagic) {
                            text := enemy.CastingSpell.Name
                            if text == "" {
                                text = "None"
                            }
                            fonts.SpellFont.PrintOptions(screen, float64((position.X + 21)), float64(position.Y), centerShadow, text)
                        }
                    }
                } else {
                    scale.DrawScaled(screen, gemUnknown, &options)
                }
            },
        })

        if i < len(knownPlayers) {
            enemy := knownPlayers[i]
            if !enemy.Defeated {
                positionStart := gemPositions[i]
                positionStart.X += gemUnknown.Bounds().Dx() + 2
                positionStart.Y -= 2
                for otherPlayer, relationship := range enemy.PlayerRelations {
                    treatyIcon := getTreatyIcon(otherPlayer, relationship.Treaty)
                    if treatyIcon != nil {
                        usePosition := positionStart
                        rect := util.ImageRect(usePosition.X, usePosition.Y, treatyIcon)
                        group.AddElement(&uilib.UIElement{
                            Rect: rect,
                            RightClick: func(element *uilib.UIElement){
                                helpEntries := help.GetEntriesByName("TREATIES")
                                if helpEntries != nil {
                                    group.AddElement(uilib.MakeHelpElement(group, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
                                }
                            },
                            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                                var options ebiten.DrawImageOptions
                                options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))
                                scale.DrawScaled(screen, treatyIcon, &options)
                            },
                        })

                        positionStart.Y += treatyIcon.Bounds().Dy() + 1
                    }
                }
            }
        }

    }

    distribute := func(amount float64, update *float64, other1 *float64, other1Locked bool, other2 *float64, other2Locked bool){
        if other1Locked && other2Locked {
            return
        }

        diff := *update - amount
        *update = amount

        // distribute diff to the others
        if !other1Locked && !other2Locked {
            *other1 += diff / 2
            *other2 += diff / 2

            if *other1 < 0 {
                *other2 += *other1
                *other1 = 0
            } else if *other2 < 0 {
                *other1 += *other2
                *other2 = 0
            }

        } else if !other1Locked {
            *other1 += diff
            if *other1 < 0 {
                *update += *other1
                *other1 = 0
            }
        } else if !other2Locked {
            *other2 += diff

            if *other2 < 0 {
                *update += *other2
                *other2 = 0
            }
        }
    }

    adjustManaPercent := func(amount float64){
        distribute(amount, &player.PowerDistribution.Mana, &player.PowerDistribution.Research, magic.ResearchLocked, &player.PowerDistribution.Skill, magic.SkillLocked)
        // log.Printf("distribution: %+v", player.PowerDistribution)
    }

    adjustResearchPercent := func(amount float64){
        distribute(amount, &player.PowerDistribution.Research, &player.PowerDistribution.Mana, magic.ManaLocked, &player.PowerDistribution.Skill, magic.SkillLocked)
    }

    adjustSkillPercent := func(amount float64){
        distribute(amount, &player.PowerDistribution.Skill, &player.PowerDistribution.Mana, magic.ManaLocked, &player.PowerDistribution.Research, magic.ResearchLocked)
    }

    manaLocked, err := magic.ImageCache.GetImage("magic.lbx", 15, 0)
    if err == nil {
        group.AddElement(&uilib.UIElement{
            Rect: image.Rect(27, 81, 27 + manaLocked.Bounds().Dx(), 81 + manaLocked.Bounds().Dy() - 2),
            RightClick: func(element *uilib.UIElement){
                helpEntries := help.GetEntriesByName("Mana Points Ratio")
                if helpEntries != nil {
                    group.AddElement(uilib.MakeHelpElement(group, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
                }
            },
            PlaySoundLeftClick: true,
            LeftClick: func(element *uilib.UIElement){
                magic.ManaLocked = !magic.ManaLocked
            },
        })

        manaStaff, _ := magic.ImageCache.GetImage("magic.lbx", 7, 0)

        posY := 0
        manaPowerStaff, _ := magic.ImageCache.GetImage("magic.lbx", 8, 0)
        staffRect := image.Rect(33, 102, 38, 102 + manaPowerStaff.Bounds().Dy())

        group.AddElement(&uilib.UIElement{
            Rect: staffRect,
            PlaySoundLeftClick: true,
            LeftClick: func(element *uilib.UIElement){
                if !magic.ManaLocked {
                    // log.Printf("click mana staff at %v", manaStaff.Bounds().Dy() - posY)
                    amount := manaPowerStaff.Bounds().Dy() - posY
                    adjustManaPercent(float64(amount) / float64(manaPowerStaff.Bounds().Dy()))
                }
            },
            RightClick: func(element *uilib.UIElement){
                helpEntries := help.GetEntriesByName("Mana Points Ratio")
                if helpEntries != nil {
                    group.AddElement(uilib.MakeHelpElement(group, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
                }
            },
            Inside: func(element *uilib.UIElement, x, y int){
                posY = y
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(29), float64(83))
                scale.DrawScaled(screen, manaStaff, &options)

                if player.PowerDistribution.Mana > 0 {
                    length := manaPowerStaff.Bounds().Dy() - int(float64(manaPowerStaff.Bounds().Dy()) * player.PowerDistribution.Mana)
                    part := manaPowerStaff.SubImage(image.Rect(0, length, manaPowerStaff.Bounds().Dx(), manaPowerStaff.Bounds().Dy())).(*ebiten.Image)
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(32), float64(staffRect.Min.Y + length))
                    scale.DrawScaled(screen, part, &options)
                }

                if magic.ManaLocked {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(27), float64(81))
                    scale.DrawScaled(screen, manaLocked, &options)
                }
            },
        })
    }

    // transmute button
    transmuteRect := image.Rect(235, 185, 290, 195)
    group.AddElement(&uilib.UIElement{
        Rect: transmuteRect,
        PlaySoundLeftClick: true,
        LeftClick: func(element *uilib.UIElement){
            transmuteGroup := MakeTransmuteElements(ui, fonts.TransmuteFont, player, &help, magic.Cache, &magic.ImageCache)
            ui.AddGroup(transmuteGroup)
        },
        RightClick: func(element *uilib.UIElement){
            helpEntries := help.GetEntries(247)
            if helpEntries != nil {
                group.AddElement(uilib.MakeHelpElement(group, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            // vector.StrokeRect(screen, float32(transmuteRect.Min.X), float32(transmuteRect.Min.Y), float32(transmuteRect.Dx()), float32(transmuteRect.Bounds().Dy()), 1, color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff}, true)
        },
    })

    // ok button
    okRect := image.Rect(296, 185, 316, 195)
    group.AddElement(&uilib.UIElement{
        Rect: okRect,
        PlaySoundLeftClick: true,
        LeftClick: func(element *uilib.UIElement){
            magic.State = MagicScreenStateDone
        },
        RightClick: func(element *uilib.UIElement){
            helpEntries := help.GetEntries(248)
            if helpEntries != nil {
                group.AddElement(uilib.MakeHelpElement(group, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            // vector.StrokeRect(screen, float32(okRect.Min.X), float32(okRect.Min.Y), float32(okRect.Dx()), float32(okRect.Bounds().Dy()), 1, color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff}, true)
        },
    })

    researchLocked, err := magic.ImageCache.GetImage("magic.lbx", 16, 0)
    if err == nil {
        researchStaff, _ := magic.ImageCache.GetImage("magic.lbx", 9, 0)

        group.AddElement(&uilib.UIElement{
            Rect: image.Rect(74, 81, 74 + researchLocked.Bounds().Dx(), 81 + researchLocked.Bounds().Dy() - 2),
            PlaySoundLeftClick: true,
            LeftClick: func(element *uilib.UIElement){
                magic.ResearchLocked = !magic.ResearchLocked
            },
            RightClick: func(element *uilib.UIElement){
                helpEntries := help.GetEntriesByName("Research Ratio")
                if helpEntries != nil {
                    group.AddElement(uilib.MakeHelpElement(group, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
                }
            },
        })

        researchPowerStaff, _ := magic.ImageCache.GetImage("magic.lbx", 10, 0)
        posY := 0
        staffRect := image.Rect(79, 102, 86, 102 + researchPowerStaff.Bounds().Dy())

        group.AddElement(&uilib.UIElement{
            Rect: staffRect,
            PlaySoundLeftClick: true,
            LeftClick: func(element *uilib.UIElement){
                if !magic.ResearchLocked {
                    // log.Printf("click mana staff at %v", manaStaff.Bounds().Dy() - posY)
                    amount := researchPowerStaff.Bounds().Dy() - posY
                    adjustResearchPercent(float64(amount) / float64(researchPowerStaff.Bounds().Dy()))
                }
            },
            RightClick: func(element *uilib.UIElement){
                helpEntries := help.GetEntriesByName("Research Ratio")
                if helpEntries != nil {
                    group.AddElement(uilib.MakeHelpElement(group, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
                }
            },
            Inside: func(element *uilib.UIElement, x, y int){
                posY = y
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(75), float64(85))
                scale.DrawScaled(screen, researchStaff, &options)

                if player.PowerDistribution.Research > 0 {
                    length := researchPowerStaff.Bounds().Dy() - int(float64(researchPowerStaff.Bounds().Dy()) * player.PowerDistribution.Research)
                    part := researchPowerStaff.SubImage(image.Rect(0, length, researchPowerStaff.Bounds().Dx(), researchPowerStaff.Bounds().Dy())).(*ebiten.Image)
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(staffRect.Min.X), float64(staffRect.Min.Y + length))
                    scale.DrawScaled(screen, part, &options)
                }

                if magic.ResearchLocked {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(74), float64(81))
                    scale.DrawScaled(screen, researchLocked, &options)
                }

            },
        })
    }

    skillLocked, err := magic.ImageCache.GetImage("magic.lbx", 17, 0)
    if err == nil {
        group.AddElement(&uilib.UIElement{
            Rect: image.Rect(121, 81, 121 + skillLocked.Bounds().Dx(), 81 + skillLocked.Bounds().Dy() - 3),
            PlaySoundLeftClick: true,
            LeftClick: func(element *uilib.UIElement){
                magic.SkillLocked = !magic.SkillLocked
            },
            RightClick: func(element *uilib.UIElement){
                helpEntries := help.GetEntriesByName("Casting Skill Ratio")
                if helpEntries != nil {
                    group.AddElement(uilib.MakeHelpElement(group, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
                }
            },
        })

        skillPowerStaff, _ := magic.ImageCache.GetImage("magic.lbx", 10, 0)
        posY := 0
        staffRect := image.Rect(126, 102, 132, 102 + skillPowerStaff.Bounds().Dy())

        group.AddElement(&uilib.UIElement{
            Rect: staffRect,
            PlaySoundLeftClick: true,
            LeftClick: func(element *uilib.UIElement){
                if !magic.SkillLocked {
                    // log.Printf("click mana staff at %v", manaStaff.Bounds().Dy() - posY)
                    amount := skillPowerStaff.Bounds().Dy() - posY
                    adjustSkillPercent(float64(amount) / float64(skillPowerStaff.Bounds().Dy()))
                }
            },
            RightClick: func(element *uilib.UIElement){
                helpEntries := help.GetEntriesByName("Casting Skill Ratio")
                if helpEntries != nil {
                    group.AddElement(uilib.MakeHelpElement(group, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
                }
            },
            Inside: func(element *uilib.UIElement, x, y int){
                posY = y
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                skillStaff, err := magic.ImageCache.GetImage("magic.lbx", 11, 0)
                if err == nil {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(122), float64(83))
                    scale.DrawScaled(screen, skillStaff, &options)
                }

                if player.PowerDistribution.Skill > 0 {
                    length := skillPowerStaff.Bounds().Dy() - int(float64(skillPowerStaff.Bounds().Dy()) * player.PowerDistribution.Skill)
                    part := skillPowerStaff.SubImage(image.Rect(0, length, skillPowerStaff.Bounds().Dx(), skillPowerStaff.Bounds().Dy())).(*ebiten.Image)
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(staffRect.Min.X), float64(staffRect.Min.Y + length))
                    scale.DrawScaled(screen, part, &options)
                }

                if magic.SkillLocked {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(121), float64(81))
                    scale.DrawScaled(screen, skillLocked, &options)
                }
            },
        })
    }

    overworldCastingSkill := player.ComputeOverworldCastingSkill()
    castingSkill := player.ComputeCastingSkill()

    spellCastUIRect := image.Rect(5, 175, 99, 196)
    group.AddElement(&uilib.UIElement{
        Rect: spellCastUIRect,
        RightClick: func(element *uilib.UIElement){
            helpEntries := help.GetEntriesByName("Spell Casting Skill")
            if helpEntries != nil {
                group.AddElement(uilib.MakeHelpElement(group, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            // vector.StrokeRect(screen, float32(spellCastUIRect.Min.X), float32(spellCastUIRect.Min.Y), float32(spellCastUIRect.Dx()), float32(spellCastUIRect.Dy()), 1, color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff}, false)

            fonts.SmallerFont.PrintOptions(screen, float64(5), float64(176), leftShadow, fmt.Sprintf("Casting Skill: %v(%v)", overworldCastingSkill, castingSkill))
            fonts.SmallerFont.PrintOptions(screen, float64(5), float64(183), leftShadow, fmt.Sprintf("Magic Reserve: %v", player.Mana))
            fonts.SmallerFont.PrintOptions(screen, float64(5), float64(190), leftShadow, fmt.Sprintf("Power Base: %v", magic.Power))
        },
    })

    castingRect := image.Rect(100, 175, 220, 196)
    group.AddElement(&uilib.UIElement{
        Rect: castingRect,
        RightClick: func(element *uilib.UIElement){
            helpEntries := help.GetEntries(276)
            if helpEntries != nil {
                group.AddElement(uilib.MakeHelpElement(group, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            // util.DrawRect(screen, castingRect, color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff})
            fonts.SmallerFont.PrintOptions(screen, float64(100), float64(176), leftShadow, fmt.Sprintf("Casting: %v", player.CastingSpell.Name))
            fonts.SmallerFont.PrintOptions(screen, float64(100), float64(183), leftShadow, fmt.Sprintf("Researching: %v", player.ResearchingSpell.Name))

            summonCity := player.FindSummoningCity()
            if summonCity == nil {
                summonCity = &citylib.City{
                    Name: "",
                }
            }

            fonts.SmallerFont.PrintOptions(screen, float64(100), float64(190), leftShadow, fmt.Sprintf("Summon To: %v", summonCity.Name))
        },
    })

    type EnchantmentElement struct {
        Enchantment data.Enchantment
        Banner data.BannerType
    }

    allEnchantments := func() []EnchantmentElement {
        var out []EnchantmentElement
        for _, enchantment := range player.GlobalEnchantments.Values() {
            out = append(out, EnchantmentElement{
                Enchantment: enchantment,
                Banner: player.GetBanner(),
            })
        }

        for _, enemy := range knownPlayers {
            for _, enchantment := range enemy.GlobalEnchantments.Values() {
                out = append(out, EnchantmentElement{
                    Enchantment: enchantment,
                    Banner: enemy.GetBanner(),
                })
            }
        }
        return slices.SortedFunc(slices.Values(out), func (a, b EnchantmentElement) int {
            return cmp.Compare(a.Enchantment.String(), b.Enchantment.String())
        })
    }

    ui.SetElementsFromArray(nil)
    ui.AddGroup(group)

    var globalEnchantments []*uilib.UIElement
    var setupEnchantments func()
    setupEnchantments = func() {
        ui.RemoveElements(globalEnchantments)
        globalEnchantments = nil

        for i, enchantment := range allEnchantments() {
            name := enchantment.Enchantment.String()
            var useFont *font.Font
            switch enchantment.Banner {
                case data.BannerBlue: useFont = fonts.BannerBlueFont
                case data.BannerGreen: useFont = fonts.BannerGreenFont
                case data.BannerPurple: useFont = fonts.BannerPurpleFont
                case data.BannerRed: useFont = fonts.BannerRedFont
                case data.BannerYellow: useFont = fonts.BannerYellowFont
            }

            yStart := 80
            rect := image.Rect(170, (yStart + i * useFont.Height()), 310, (yStart + (i + 1) * useFont.Height()))
            globalEnchantments = append(globalEnchantments, &uilib.UIElement{
                Rect: rect,
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    useFont.PrintOptions(screen, float64(rect.Min.X), float64(rect.Min.Y), leftShadow, name)
                },
                LeftClick: func(element *uilib.UIElement){
                    // can only cancel the player's enchantments
                    if enchantment.Banner == player.GetBanner() {
                        group := uilib.MakeGroup()
                        no := func(){
                            ui.RemoveGroup(group)
                        }
                        yes := func(){
                            // FIXME: kill units that now have 0 health
                            player.GlobalEnchantments.Remove(enchantment.Enchantment)
                            player.UpdateUnrest()
                            for _, enemy := range enemies {
                                enemy.UpdateUnrest()
                            }
                            setupEnchantments()
                            ui.RemoveGroup(group)
                        }

                        group.AddElements(uilib.MakeConfirmDialog(group, magic.Cache, &magic.ImageCache, fmt.Sprintf("Do you wish to cancel your %v spell?", name), false, yes, no))
                        ui.AddGroup(group)
                    }
                },
                RightClick: func(element *uilib.UIElement) {
                    helpEntries := help.GetEntriesByName(name)
                    if helpEntries != nil {
                        group.AddElement(uilib.MakeHelpElementWithLayer(group, magic.Cache, &magic.ImageCache, 2, helpEntries[0], helpEntries[1:]...))
                    }
                },
            })
        }

        if len(globalEnchantments) == 0 {
            enchantmentsRect := image.Rect(168, 67, 310, 172)
            globalEnchantments = append(globalEnchantments, &uilib.UIElement{
                Rect: enchantmentsRect,
                RightClick: func(element *uilib.UIElement){
                    helpEntries := help.GetEntriesByName("Enchantments")
                    if helpEntries != nil {
                        group.AddElement(uilib.MakeHelpElement(group, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
                    }
                },
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    // util.DrawRect(screen, enchantmentsRect, color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff})
                },
            })
        }

        ui.AddElements(globalEnchantments)
    }

    setupEnchantments()

    return ui
}

func (magic *MagicScreen) Update() MagicScreenState {
    magic.UI.StandardUpdate()

    return magic.State
}

func (magic *MagicScreen) Draw(screen *ebiten.Image){
    magic.UI.Draw(magic.UI, screen)
}
