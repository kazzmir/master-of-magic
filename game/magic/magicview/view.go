package magicview

import (
    "fmt"
    "log"
    "image"
    "image/color"
    "math"
    "math/rand/v2"
    "slices"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/mirror"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
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

func MakeTransmuteElements(ui *uilib.UI, smallFont *font.Font, player *playerlib.Player, help *lbx.Help, cache *lbx.LbxCache, imageCache *util.ImageCache) []*uilib.UIElement {
    var elements []*uilib.UIElement

    ok, _ := imageCache.GetImages("magic.lbx", 54)
    okRect := image.Rect(176, 99, 176 + ok[0].Bounds().Dx(), 99 + ok[0].Bounds().Dy())
    okIndex := 0

    arrowRight, _ := imageCache.GetImage("magic.lbx", 55, 0)
    arrowLeft, _ := imageCache.GetImage("magic.lbx", 56, 0)

    // if true, then convert power to gold, otherwise convert gold to power
    isRight := true

    arrowRect := image.Rect(146, 99, 146 + arrowRight.Bounds().Dx(), 99 + arrowRight.Bounds().Dy())

    cancel, _ := imageCache.GetImages("magic.lbx", 53)
    cancelRect := image.Rect(93, 99, 93 + cancel[0].Bounds().Dx(), 99 + cancel[0].Bounds().Dy())
    cancelIndex := 0

    powerToGold, _ := imageCache.GetImage("magic.lbx", 59, 0)

    totalGold := player.Gold
    totalMana := player.Mana

    alchemyConversion := 0.5
    if player.Wizard.AbilityEnabled(setup.AbilityAlchemy) {
        alchemyConversion = 1
    }

    conveyor, _ := imageCache.GetImage("magic.lbx", 57, 0)
    cursor, _ := imageCache.GetImage("magic.lbx", 58, 0)

    cursorPosition := 0
    changePercent := float64(0)

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("magic.lbx", 52, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(75, 60)
            screen.DrawImage(background, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(okRect.Min.X), float64(okRect.Min.Y))
            screen.DrawImage(ok[okIndex], &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(arrowRect.Min.X), float64(arrowRect.Min.Y))

            flip := false
            if isRight {
                screen.DrawImage(arrowRight, &options)

                options.GeoM.Reset()
                options.GeoM.Translate(87, 70)
                screen.DrawImage(powerToGold, &options)

                flip = true
            } else {
                screen.DrawImage(arrowLeft, &options)
            }

            // the conveyor belt should only be drawn until the position of the cursor
            conveyorArea := screen.SubImage(image.Rect(131, 85, 131 + cursorPosition, 85 + 7)).(*ebiten.Image)

            // draw an animated conveyor belt by drawing the same image twice with a slight offset
            movement := int((ui.Counter / 6) % 8)
            if flip {
                options.GeoM.Reset()
                options.GeoM.Scale(-1, 1)
                options.GeoM.Translate(131, 85)
                options.GeoM.Translate(float64(cursorPosition + conveyor.Bounds().Dx()), 0)
                options.GeoM.Translate(float64(-movement), 0)
                conveyorArea.DrawImage(conveyor, &options)

                options.GeoM.Reset()
                options.GeoM.Scale(-1, 1)
                options.GeoM.Translate(131, 85)
                options.GeoM.Translate(float64(cursorPosition), 0)
                options.GeoM.Translate(float64(-movement), 0)
                conveyorArea.DrawImage(conveyor, &options)
            } else {
                options.GeoM.Reset()
                options.GeoM.Translate(131, 85)
                options.GeoM.Translate(float64(-conveyor.Bounds().Dx()), 0)
                options.GeoM.Translate(float64(movement), 0)
                conveyorArea.DrawImage(conveyor, &options)

                options.GeoM.Reset()
                options.GeoM.Translate(131, 85)
                options.GeoM.Translate(float64(movement), 0)
                conveyorArea.DrawImage(conveyor, &options)
            }

            // draw the cursor itself
            options.GeoM.Reset()
            options.GeoM.Translate(131 + float64(cursorPosition) - 3, 85)
            screen.DrawImage(cursor, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(cancelRect.Min.X), float64(cancelRect.Min.Y))
            screen.DrawImage(cancel[cancelIndex], &options)

            leftSide := float64(122)
            rightSide := float64(224)
            if isRight {
                smallFont.PrintRight(screen, leftSide, 86, 1, ebiten.ColorScale{}, fmt.Sprintf("%v GP", int(float64(totalMana) * changePercent * alchemyConversion)))
                smallFont.PrintRight(screen, rightSide, 86, 1, ebiten.ColorScale{}, fmt.Sprintf("%v PP", int(float64(totalMana) * changePercent)))
            } else {
                smallFont.PrintRight(screen, leftSide, 86, 1, ebiten.ColorScale{}, fmt.Sprintf("%v GP", int(float64(totalGold) * changePercent)))
                smallFont.PrintRight(screen, rightSide, 86, 1, ebiten.ColorScale{}, fmt.Sprintf("%v PP", int(float64(totalGold) * changePercent * alchemyConversion)))
            }
        },
    })

    {
        posX := 0
        conveyorRect := image.Rect(131, 85, 131 + 56, 85 + 7)
        elements = append(elements, &uilib.UIElement{
            Rect: conveyorRect,
            Layer: 1,
            PlaySoundLeftClick: true,
            RightClick: func(element *uilib.UIElement){
                helpEntries := help.GetEntriesByName("Alchemy Ratio")
                if helpEntries != nil {
                    ui.AddElement(uilib.MakeHelpElementWithLayer(ui, cache, imageCache, 2, helpEntries[0], helpEntries[1:]...))
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

    elements = append(elements, &uilib.UIElement{
        Rect: image.Rect(94, 85, 123, 92),
        Layer: 1,
        RightClick: func(element *uilib.UIElement){
            helpEntries := help.GetEntriesByName("Alchemy Gold")
            if helpEntries != nil {
                ui.AddElement(uilib.MakeHelpElementWithLayer(ui, cache, imageCache, 2, helpEntries[0], helpEntries[1:]...))
            }
        },
    })

    elements = append(elements, &uilib.UIElement{
        Rect: image.Rect(195, 85, 225, 92),
        Layer: 1,
        RightClick: func(element *uilib.UIElement){
            helpEntries := help.GetEntriesByName("Alchemy Power")
            if helpEntries != nil {
                ui.AddElement(uilib.MakeHelpElementWithLayer(ui, cache, imageCache, 2, helpEntries[0], helpEntries[1:]...))
            }
        },
    })

    // cancel button
    elements = append(elements, &uilib.UIElement{
        Rect: cancelRect,
        Layer: 1,
        PlaySoundLeftClick: true,
        LeftClick: func(element *uilib.UIElement){
            cancelIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            cancelIndex = 0
            ui.RemoveElements(elements)
        },
        RightClick: func(element *uilib.UIElement){
            helpEntries := help.GetEntries(371)
            if helpEntries != nil {
                ui.AddElement(uilib.MakeHelpElementWithLayer(ui, cache, imageCache, 2, helpEntries[0], helpEntries[1:]...))
            }
        },
    })

    // arrow button
    elements = append(elements, &uilib.UIElement{
        Rect: arrowRect,
        Layer: 1,
        PlaySoundLeftClick: true,
        LeftClick: func(element *uilib.UIElement){
            isRight = !isRight
        },
        RightClick: func(element *uilib.UIElement){
            helpEntries := help.GetEntries(372)
            if helpEntries != nil {
                ui.AddElement(uilib.MakeHelpElementWithLayer(ui, cache, imageCache, 2, helpEntries[0], helpEntries[1:]...))
            }
        },
    })

    // ok button
    elements = append(elements, &uilib.UIElement{
        Rect: okRect,
        Layer: 1,
        PlaySoundLeftClick: true,
        LeftClick: func(element *uilib.UIElement){
            okIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            okIndex = 0
            ui.RemoveElements(elements)

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
                ui.AddElement(uilib.MakeHelpElementWithLayer(ui, cache, imageCache, 2, helpEntries[0], helpEntries[1:]...))
            }
        },

    })

    return elements
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

    fontLbx, err := magic.Cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
    }

    blue := color.RGBA{R: 0x6e, G: 0x79, B: 0xe6, A: 0xff}
    bluishPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        blue, blue, blue, blue,
    }

    normalFont := font.MakeOptimizedFontWithPalette(fonts[2], bluishPalette)

    blue2Palette := bluishPalette
    blue2Palette[1] = color.RGBA{R: 0x52, G: 0x61, B: 0xca, A: 0xff}
    smallerFont := font.MakeOptimizedFontWithPalette(fonts[1], blue2Palette)

    translucentWhite := util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 80})
    transmutePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        translucentWhite, translucentWhite, translucentWhite,
        translucentWhite, translucentWhite, translucentWhite,
    }

    transmuteFont := font.MakeOptimizedFontWithPalette(fonts[0], transmutePalette)

    helpLbx, err := magic.Cache.GetLbxFile("help.lbx")
    if err != nil {
        return nil
    }

    help, err := helpLbx.ReadHelp(2)
    if err != nil {
        return nil
    }

    ui := &uilib.UI{
        Cache: magic.Cache,
        Draw: func(ui *uilib.UI, screen *ebiten.Image) {
            background, err := magic.ImageCache.GetImage("magic.lbx", 0, 0)
            if err == nil {
                var options ebiten.DrawImageOptions
                screen.DrawImage(background, &options)
            }

            mana := int(math.Round(player.PowerDistribution.Mana * float64(magic.Power)))
            research := int(math.Round(player.PowerDistribution.Research * float64(magic.Power)))
            skill := magic.Power - (mana + research)

            normalFont.PrintRight(screen, 56, 160, 1, ebiten.ColorScale{}, fmt.Sprintf("%v MP", mana))
            normalFont.PrintRight(screen, 103, 160, 1, ebiten.ColorScale{}, fmt.Sprintf("%v RP", research))
            normalFont.PrintRight(screen, 151, 160, 1, ebiten.ColorScale{}, fmt.Sprintf("%v SP", skill))

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })
        },
    }

    var elements []*uilib.UIElement

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
            if otherPlayer == player || slices.Contains(enemies, otherPlayer) {
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
        elements = append(elements, &uilib.UIElement{
            Rect: rect,
            LeftClick: func(element *uilib.UIElement){
                // show diplomatic dialogue screen
            },
            RightClick: func(element *uilib.UIElement){
                // show mirror ui with extra enemy info: relations, treaties, personality, objective

                if i < len(enemies) && !enemies[i].Defeated {
                    mirrorElement := mirror.MakeMirrorUI(magic.Cache, enemies[i], ui)
                    ui.AddElement(mirrorElement)
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(position.X), float64(position.Y))

                if i < len(enemies) {
                    enemy := enemies[i]

                    if enemy.Defeated {
                        screen.DrawImage(gemDefeated, &options)
                    } else {
                        portraitIndex := mirror.GetWizardPortraitIndex(enemy.Wizard.Base, enemy.Wizard.Banner)
                        portrait, _ := magic.ImageCache.GetImage("lilwiz.lbx", portraitIndex, 0)
                        if portrait != nil {
                            screen.DrawImage(portrait, &options)
                        }
                    }
                } else {
                    screen.DrawImage(gemUnknown, &options)
                }
            },
        })

        if i < len(enemies) {
            enemy := enemies[i]
            if !enemy.Defeated {
                positionStart := gemPositions[i]
                positionStart.X += gemUnknown.Bounds().Dx() + 2
                positionStart.Y -= 2
                for otherPlayer, relationship := range enemy.PlayerRelations {
                    treatyIcon := getTreatyIcon(otherPlayer, relationship.Treaty)
                    if treatyIcon != nil {
                        usePosition := positionStart
                        rect := util.ImageRect(usePosition.X, usePosition.Y, treatyIcon)
                        elements = append(elements, &uilib.UIElement{
                            Rect: rect,
                            RightClick: func(element *uilib.UIElement){
                                helpEntries := help.GetEntriesByName("TREATIES")
                                if helpEntries != nil {
                                    ui.AddElement(uilib.MakeHelpElement(ui, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
                                }
                            },
                            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                                var options ebiten.DrawImageOptions
                                options.GeoM.Translate(float64(usePosition.X), float64(usePosition.Y))
                                screen.DrawImage(treatyIcon, &options)
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
        elements = append(elements, &uilib.UIElement{
            Rect: image.Rect(27, 81, 27 + manaLocked.Bounds().Dx(), 81 + manaLocked.Bounds().Dy() - 2),
            RightClick: func(element *uilib.UIElement){
                helpEntries := help.GetEntriesByName("Mana Points Ratio")
                if helpEntries != nil {
                    ui.AddElement(uilib.MakeHelpElement(ui, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
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

        elements = append(elements, &uilib.UIElement{
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
                    ui.AddElement(uilib.MakeHelpElement(ui, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
                }
            },
            Inside: func(element *uilib.UIElement, x, y int){
                posY = y
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(29, 83)
                screen.DrawImage(manaStaff, &options)

                if player.PowerDistribution.Mana > 0 {
                    length := manaPowerStaff.Bounds().Dy() - int(float64(manaPowerStaff.Bounds().Dy()) * player.PowerDistribution.Mana)
                    part := manaPowerStaff.SubImage(image.Rect(0, length, manaPowerStaff.Bounds().Dx(), manaPowerStaff.Bounds().Dy())).(*ebiten.Image)
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(32, float64(staffRect.Min.Y + length))
                    screen.DrawImage(part, &options)
                }

                if magic.ManaLocked {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(27, 81)
                    screen.DrawImage(manaLocked, &options)
                }
            },
        })
    }

    // transmute button
    transmuteRect := image.Rect(235, 185, 290, 195)
    elements = append(elements, &uilib.UIElement{
        Rect: transmuteRect,
        PlaySoundLeftClick: true,
        LeftClick: func(element *uilib.UIElement){
            transmuteElements := MakeTransmuteElements(ui, transmuteFont, player, &help, magic.Cache, &magic.ImageCache)
            ui.AddElements(transmuteElements)
        },
        RightClick: func(element *uilib.UIElement){
            helpEntries := help.GetEntries(247)
            if helpEntries != nil {
                ui.AddElement(uilib.MakeHelpElement(ui, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            // vector.StrokeRect(screen, float32(transmuteRect.Min.X), float32(transmuteRect.Min.Y), float32(transmuteRect.Dx()), float32(transmuteRect.Bounds().Dy()), 1, color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff}, true)
        },
    })

    // ok button
    okRect := image.Rect(296, 185, 316, 195)
    elements = append(elements, &uilib.UIElement{
        Rect: okRect,
        PlaySoundLeftClick: true,
        LeftClick: func(element *uilib.UIElement){
            magic.State = MagicScreenStateDone
        },
        RightClick: func(element *uilib.UIElement){
            helpEntries := help.GetEntries(248)
            if helpEntries != nil {
                ui.AddElement(uilib.MakeHelpElement(ui, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            // vector.StrokeRect(screen, float32(okRect.Min.X), float32(okRect.Min.Y), float32(okRect.Dx()), float32(okRect.Bounds().Dy()), 1, color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff}, true)
        },
    })

    researchLocked, err := magic.ImageCache.GetImage("magic.lbx", 16, 0)
    if err == nil {
        researchStaff, _ := magic.ImageCache.GetImage("magic.lbx", 9, 0)

        elements = append(elements, &uilib.UIElement{
            Rect: image.Rect(74, 81, 74 + researchLocked.Bounds().Dx(), 81 + researchLocked.Bounds().Dy() - 1),
            PlaySoundLeftClick: true,
            LeftClick: func(element *uilib.UIElement){
                magic.ResearchLocked = !magic.ResearchLocked
            },
            RightClick: func(element *uilib.UIElement){
                helpEntries := help.GetEntriesByName("Research Ratio")
                if helpEntries != nil {
                    ui.AddElement(uilib.MakeHelpElement(ui, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
                }
            },
        })

        researchPowerStaff, _ := magic.ImageCache.GetImage("magic.lbx", 10, 0)
        posY := 0
        staffRect := image.Rect(79, 102, 86, 102 + researchPowerStaff.Bounds().Dy())

        elements = append(elements, &uilib.UIElement{
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
                    ui.AddElement(uilib.MakeHelpElement(ui, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
                }
            },
            Inside: func(element *uilib.UIElement, x, y int){
                posY = y
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(75, 85)
                screen.DrawImage(researchStaff, &options)

                if player.PowerDistribution.Research > 0 {
                    length := researchPowerStaff.Bounds().Dy() - int(float64(researchPowerStaff.Bounds().Dy()) * player.PowerDistribution.Research)
                    part := researchPowerStaff.SubImage(image.Rect(0, length, researchPowerStaff.Bounds().Dx(), researchPowerStaff.Bounds().Dy())).(*ebiten.Image)
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(staffRect.Min.X), float64(staffRect.Min.Y + length))
                    screen.DrawImage(part, &options)
                }

                if magic.ResearchLocked {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(74, 81)
                    screen.DrawImage(researchLocked, &options)
                }

            },
        })
    }

    skillLocked, err := magic.ImageCache.GetImage("magic.lbx", 17, 0)
    if err == nil {
        elements = append(elements, &uilib.UIElement{
            Rect: image.Rect(121, 81, 121 + skillLocked.Bounds().Dx(), 81 + skillLocked.Bounds().Dy() - 3),
            PlaySoundLeftClick: true,
            LeftClick: func(element *uilib.UIElement){
                magic.SkillLocked = !magic.SkillLocked
            },
            RightClick: func(element *uilib.UIElement){
                helpEntries := help.GetEntriesByName("Casting Skill Ratio")
                if helpEntries != nil {
                    ui.AddElement(uilib.MakeHelpElement(ui, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
                }
            },
        })

        skillPowerStaff, _ := magic.ImageCache.GetImage("magic.lbx", 10, 0)
        posY := 0
        staffRect := image.Rect(126, 102, 132, 102 + skillPowerStaff.Bounds().Dy())

        elements = append(elements, &uilib.UIElement{
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
                    ui.AddElement(uilib.MakeHelpElement(ui, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
                }
            },
            Inside: func(element *uilib.UIElement, x, y int){
                posY = y
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                skillStaff, err := magic.ImageCache.GetImage("magic.lbx", 11, 0)
                if err == nil {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(122, 83)
                    screen.DrawImage(skillStaff, &options)
                }

                if player.PowerDistribution.Skill > 0 {
                    length := skillPowerStaff.Bounds().Dy() - int(float64(skillPowerStaff.Bounds().Dy()) * player.PowerDistribution.Skill)
                    part := skillPowerStaff.SubImage(image.Rect(0, length, skillPowerStaff.Bounds().Dx(), skillPowerStaff.Bounds().Dy())).(*ebiten.Image)
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(staffRect.Min.X), float64(staffRect.Min.Y + length))
                    screen.DrawImage(part, &options)
                }

                if magic.SkillLocked {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(121, 81)
                    screen.DrawImage(skillLocked, &options)
                }
            },
        })
    }

    spellCastUIRect := image.Rect(5, 175, 99, 196)
    elements = append(elements, &uilib.UIElement{
        Rect: spellCastUIRect,
        RightClick: func(element *uilib.UIElement){
            helpEntries := help.GetEntriesByName("Spell Casting Skill")
            if helpEntries != nil {
                ui.AddElement(uilib.MakeHelpElement(ui, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            // vector.StrokeRect(screen, float32(spellCastUIRect.Min.X), float32(spellCastUIRect.Min.Y), float32(spellCastUIRect.Dx()), float32(spellCastUIRect.Dy()), 1, color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff}, false)

            smallerFont.Print(screen, 5, 176, 1, ebiten.ColorScale{}, fmt.Sprintf("Casting Skill: %v(%v)", player.RemainingCastingSkill, player.ComputeCastingSkill()))
            smallerFont.Print(screen, 5, 183, 1, ebiten.ColorScale{}, fmt.Sprintf("Magic Reserve: %v", player.Mana))
            smallerFont.Print(screen, 5, 190, 1, ebiten.ColorScale{}, fmt.Sprintf("Power Base: %v", magic.Power))
        },
    })

    castingRect := image.Rect(100, 175, 220, 196)
    elements = append(elements, &uilib.UIElement{
        Rect: castingRect,
        RightClick: func(element *uilib.UIElement){
            helpEntries := help.GetEntries(276)
            if helpEntries != nil {
                ui.AddElement(uilib.MakeHelpElement(ui, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            // util.DrawRect(screen, castingRect, color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff})
            smallerFont.Print(screen, 100, 176, 1, ebiten.ColorScale{}, fmt.Sprintf("Casting: %v", player.CastingSpell.Name))
            smallerFont.Print(screen, 100, 183, 1, ebiten.ColorScale{}, fmt.Sprintf("Researching: %v", player.ResearchingSpell.Name))

            summonCity := player.FindSummoningCity()
            if summonCity == nil {
                summonCity = &citylib.City{
                    Name: "",
                }
            }

            smallerFont.Print(screen, 100, 190, 1, ebiten.ColorScale{}, fmt.Sprintf("Summon To: %v", summonCity.Name))
        },
    })

    enchantmentsRect := image.Rect(168, 67, 310, 172)
    elements = append(elements, &uilib.UIElement{
        Rect: enchantmentsRect,
        RightClick: func(element *uilib.UIElement){
            helpEntries := help.GetEntriesByName("Enchantments")
            if helpEntries != nil {
                ui.AddElement(uilib.MakeHelpElement(ui, magic.Cache, &magic.ImageCache, helpEntries[0], helpEntries[1:]...))
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            // util.DrawRect(screen, enchantmentsRect, color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff})
        },
    })

    ui.SetElementsFromArray(elements)

    return ui
}

func (magic *MagicScreen) Update() MagicScreenState {
    magic.UI.StandardUpdate()

    return magic.State
}

func (magic *MagicScreen) Draw(screen *ebiten.Image){
    magic.UI.Draw(magic.UI, screen)
}
