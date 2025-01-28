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
    helplib "github.com/kazzmir/master-of-magic/game/magic/help"
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

func MakeTransmuteElements(ui *uilib.UI, smallFont *font.Font, player *playerlib.Player, help *helplib.Help, cache *lbx.LbxCache, imageCache *util.ImageCache) []*uilib.UIElement {
    var elements []*uilib.UIElement

    ok, _ := imageCache.GetImages("magic.lbx", 54)
    okRect := util.ImageRect(176 * data.ScreenScale, 99 * data.ScreenScale, ok[0])
    okIndex := 0

    arrowRight, _ := imageCache.GetImage("magic.lbx", 55, 0)
    arrowLeft, _ := imageCache.GetImage("magic.lbx", 56, 0)

    // if true, then convert power to gold, otherwise convert gold to power
    isRight := true

    arrowRect := util.ImageRect(146 * data.ScreenScale, 99 * data.ScreenScale, arrowRight)

    cancel, _ := imageCache.GetImages("magic.lbx", 53)
    cancelRect := util.ImageRect(93 * data.ScreenScale, 99 * data.ScreenScale, cancel[0])
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
            options.GeoM.Translate(float64(75 * data.ScreenScale), float64(60 * data.ScreenScale))
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
                options.GeoM.Translate(float64(87 * data.ScreenScale), float64(70 * data.ScreenScale))
                screen.DrawImage(powerToGold, &options)

                flip = true
            } else {
                screen.DrawImage(arrowLeft, &options)
            }

            // the conveyor belt should only be drawn until the position of the cursor
            conveyorArea := screen.SubImage(image.Rect(131 * data.ScreenScale, 85 * data.ScreenScale, 131 * data.ScreenScale + cursorPosition, (85 + 7) * data.ScreenScale)).(*ebiten.Image)

            // draw an animated conveyor belt by drawing the same image twice with a slight offset
            movement := int((ui.Counter / 6) % 8)
            if flip {
                options.GeoM.Reset()
                options.GeoM.Scale(-1, 1)
                options.GeoM.Translate(float64(131 * data.ScreenScale), float64(85 * data.ScreenScale))
                options.GeoM.Translate(float64(cursorPosition + conveyor.Bounds().Dx()), 0)
                options.GeoM.Translate(float64(-movement * data.ScreenScale), 0)
                conveyorArea.DrawImage(conveyor, &options)

                options.GeoM.Reset()
                options.GeoM.Scale(-1, 1)
                options.GeoM.Translate(float64(131 * data.ScreenScale), float64(85 * data.ScreenScale))
                options.GeoM.Translate(float64(cursorPosition), 0)
                options.GeoM.Translate(float64(-movement * data.ScreenScale), 0)
                conveyorArea.DrawImage(conveyor, &options)
            } else {
                options.GeoM.Reset()
                options.GeoM.Translate(float64(131 * data.ScreenScale), float64(85 * data.ScreenScale))
                options.GeoM.Translate(float64(-conveyor.Bounds().Dx()), 0)
                options.GeoM.Translate(float64(movement * data.ScreenScale), 0)
                conveyorArea.DrawImage(conveyor, &options)

                options.GeoM.Reset()
                options.GeoM.Translate(float64(131 * data.ScreenScale), float64(85 * data.ScreenScale))
                options.GeoM.Translate(float64(movement * data.ScreenScale), 0)
                conveyorArea.DrawImage(conveyor, &options)
            }

            // draw the cursor itself
            options.GeoM.Reset()
            options.GeoM.Translate(float64((131 - 3) * data.ScreenScale) + float64(cursorPosition), float64(85 * data.ScreenScale))
            screen.DrawImage(cursor, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(cancelRect.Min.X), float64(cancelRect.Min.Y))
            screen.DrawImage(cancel[cancelIndex], &options)

            leftSide := float64(122 * data.ScreenScale)
            rightSide := float64(224 * data.ScreenScale)
            if isRight {
                smallFont.PrintRight(screen, leftSide, float64(86 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("%v GP", int(float64(totalMana) * changePercent * alchemyConversion)))
                smallFont.PrintRight(screen, rightSide, float64(86 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("%v PP", int(float64(totalMana) * changePercent)))
            } else {
                smallFont.PrintRight(screen, leftSide, float64(86 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("%v GP", int(float64(totalGold) * changePercent)))
                smallFont.PrintRight(screen, rightSide, float64(86 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("%v PP", int(float64(totalGold) * changePercent * alchemyConversion)))
            }
        },
    })

    {
        posX := 0
        conveyorRect := image.Rect(131 * data.ScreenScale, 85 * data.ScreenScale, (131 + 56) * data.ScreenScale, (85 + 7) * data.ScreenScale)
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
        Rect: image.Rect(94 * data.ScreenScale, 85 * data.ScreenScale, 123 * data.ScreenScale, 92 * data.ScreenScale),
        Layer: 1,
        RightClick: func(element *uilib.UIElement){
            helpEntries := help.GetEntriesByName("Alchemy Gold")
            if helpEntries != nil {
                ui.AddElement(uilib.MakeHelpElementWithLayer(ui, cache, imageCache, 2, helpEntries[0], helpEntries[1:]...))
            }
        },
    })

    elements = append(elements, &uilib.UIElement{
        Rect: image.Rect(195 * data.ScreenScale, 85 * data.ScreenScale, 225 * data.ScreenScale, 92 * data.ScreenScale),
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
                screen.DrawImage(background, &options)
            }

            mana := int(math.Round(player.PowerDistribution.Mana * float64(magic.Power)))
            research := int(math.Round(player.PowerDistribution.Research * float64(magic.Power)))
            skill := magic.Power - (mana + research)

            normalFont.PrintRight(screen, float64(56 * data.ScreenScale), float64(160 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("%v MP", mana))
            normalFont.PrintRight(screen, float64(103 * data.ScreenScale), float64(160 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("%v RP", research))
            normalFont.PrintRight(screen, float64(151 * data.ScreenScale), float64(160 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("%v SP", skill))

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
        rect := util.ImageRect(position.X * data.ScreenScale, position.Y * data.ScreenScale, gemUnknown)
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
                options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))

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
                positionStart.X += gemUnknown.Bounds().Dx() + 2 * data.ScreenScale
                positionStart.Y -= 2 * data.ScreenScale
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
                                options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))
                                screen.DrawImage(treatyIcon, &options)
                            },
                        })

                        positionStart.Y += treatyIcon.Bounds().Dy() + 1 * data.ScreenScale
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
            Rect: image.Rect(27 * data.ScreenScale, 81 * data.ScreenScale, 27 * data.ScreenScale + manaLocked.Bounds().Dx(), 81 * data.ScreenScale + manaLocked.Bounds().Dy() - 2 * data.ScreenScale),
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
        staffRect := image.Rect(33 * data.ScreenScale, 102 * data.ScreenScale, 38 * data.ScreenScale, 102 * data.ScreenScale + manaPowerStaff.Bounds().Dy())

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
                options.GeoM.Translate(float64(29 * data.ScreenScale), float64(83 * data.ScreenScale))
                screen.DrawImage(manaStaff, &options)

                if player.PowerDistribution.Mana > 0 {
                    length := manaPowerStaff.Bounds().Dy() - int(float64(manaPowerStaff.Bounds().Dy()) * player.PowerDistribution.Mana)
                    part := manaPowerStaff.SubImage(image.Rect(0, length, manaPowerStaff.Bounds().Dx(), manaPowerStaff.Bounds().Dy())).(*ebiten.Image)
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(32 * data.ScreenScale), float64(staffRect.Min.Y + length))
                    screen.DrawImage(part, &options)
                }

                if magic.ManaLocked {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(27 * data.ScreenScale), float64(81 * data.ScreenScale))
                    screen.DrawImage(manaLocked, &options)
                }
            },
        })
    }

    // transmute button
    transmuteRect := image.Rect(235 * data.ScreenScale, 185 * data.ScreenScale, 290 * data.ScreenScale, 195 * data.ScreenScale)
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
    okRect := image.Rect(296 * data.ScreenScale, 185 * data.ScreenScale, 316 * data.ScreenScale, 195 * data.ScreenScale)
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
            Rect: image.Rect(74 * data.ScreenScale, 81 * data.ScreenScale, 74 * data.ScreenScale + researchLocked.Bounds().Dx(), 81 * data.ScreenScale + researchLocked.Bounds().Dy() - 2 * data.ScreenScale),
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
        staffRect := image.Rect(79 * data.ScreenScale, 102 * data.ScreenScale, 86 * data.ScreenScale, 102 * data.ScreenScale + researchPowerStaff.Bounds().Dy())

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
                options.GeoM.Translate(float64(75 * data.ScreenScale), float64(85 * data.ScreenScale))
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
                    options.GeoM.Translate(float64(74 * data.ScreenScale), float64(81 * data.ScreenScale))
                    screen.DrawImage(researchLocked, &options)
                }

            },
        })
    }

    skillLocked, err := magic.ImageCache.GetImage("magic.lbx", 17, 0)
    if err == nil {
        elements = append(elements, &uilib.UIElement{
            Rect: image.Rect(121 * data.ScreenScale, 81 * data.ScreenScale, 121 * data.ScreenScale + skillLocked.Bounds().Dx(), 81 * data.ScreenScale + skillLocked.Bounds().Dy() - 3 * data.ScreenScale),
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
        staffRect := image.Rect(126 * data.ScreenScale, 102 * data.ScreenScale, 132 * data.ScreenScale, 102 * data.ScreenScale + skillPowerStaff.Bounds().Dy())

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
                    options.GeoM.Translate(float64(122 * data.ScreenScale), float64(83 * data.ScreenScale))
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
                    options.GeoM.Translate(float64(121 * data.ScreenScale), float64(81 * data.ScreenScale))
                    screen.DrawImage(skillLocked, &options)
                }
            },
        })
    }

    spellCastUIRect := image.Rect(5 * data.ScreenScale, 175 * data.ScreenScale, 99 * data.ScreenScale, 196 * data.ScreenScale)
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

            smallerFont.Print(screen, float64(5 * data.ScreenScale), float64(176 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("Casting Skill: %v(%v)", player.RemainingCastingSkill, player.ComputeCastingSkill()))
            smallerFont.Print(screen, float64(5 * data.ScreenScale), float64(183 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("Magic Reserve: %v", player.Mana))
            smallerFont.Print(screen, float64(5 * data.ScreenScale), float64(190 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("Power Base: %v", magic.Power))
        },
    })

    castingRect := image.Rect(100 * data.ScreenScale, 175 * data.ScreenScale, 220 * data.ScreenScale, 196 * data.ScreenScale)
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
            smallerFont.Print(screen, float64(100 * data.ScreenScale), float64(176 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("Casting: %v", player.CastingSpell.Name))
            smallerFont.Print(screen, float64(100 * data.ScreenScale), float64(183 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("Researching: %v", player.ResearchingSpell.Name))

            summonCity := player.FindSummoningCity()
            if summonCity == nil {
                summonCity = &citylib.City{
                    Name: "",
                }
            }

            smallerFont.Print(screen, float64(100 * data.ScreenScale), float64(190 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("Summon To: %v", summonCity.Name))
        },
    })

    enchantmentsRect := image.Rect(168 * data.ScreenScale, 67 * data.ScreenScale, 310 * data.ScreenScale, 172 * data.ScreenScale)
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
