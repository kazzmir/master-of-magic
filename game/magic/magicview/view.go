package magicview

import (
    "fmt"
    "log"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
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

    ManaLocked bool
    ResearchLocked bool
    SkillLocked bool
}

func MakeMagicScreen(cache *lbx.LbxCache) *MagicScreen {
    magic := &MagicScreen{
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
        State: MagicScreenStateRunning,

        ManaLocked: false,
        ResearchLocked: false,
        SkillLocked: false,
    }

    magic.UI = magic.MakeUI()

    return magic
}

func (magic *MagicScreen) MakeTransmuteElements(ui *uilib.UI) []*uilib.UIElement {
    var elements []*uilib.UIElement

    ok, _ := magic.ImageCache.GetImages("magic.lbx", 54)
    okRect := image.Rect(176, 99, 176 + ok[0].Bounds().Dx(), 99 + ok[0].Bounds().Dy())
    okIndex := 0

    arrowRight, _ := magic.ImageCache.GetImage("magic.lbx", 55, 0)
    arrowLeft, _ := magic.ImageCache.GetImage("magic.lbx", 56, 0)

    // if true, then convert power to gold, otherwise convert gold to power
    isRight := true

    arrowRect := image.Rect(146, 99, 146 + arrowRight.Bounds().Dx(), 99 + arrowRight.Bounds().Dy())

    cancel, _ := magic.ImageCache.GetImages("magic.lbx", 53)
    cancelRect := image.Rect(93, 99, 93 + cancel[0].Bounds().Dx(), 99 + cancel[0].Bounds().Dy())
    cancelIndex := 0

    powerToGold, _ := magic.ImageCache.GetImage("magic.lbx", 59, 0)

    conveyor, _ := magic.ImageCache.GetImage("magic.lbx", 57, 0)
    cursor, _ := magic.ImageCache.GetImage("magic.lbx", 58, 0)

    cursorPosition := 20

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := magic.ImageCache.GetImage("magic.lbx", 52, 0)
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
            conveyorArea := screen.SubImage(image.Rect(131, 85, 131 + cursorPosition + 3, 85 + 7)).(*ebiten.Image)

            // draw an animated conveyor belt by drawing the same image twice with a slight offset
            movement := int((ui.Counter / 6) % 8)
            if flip {
                options.GeoM.Reset()
                options.GeoM.Scale(-1, 1)
                options.GeoM.Translate(131, 85)
                options.GeoM.Translate(float64(cursorPosition + 3 + conveyor.Bounds().Dx()), 0)
                options.GeoM.Translate(float64(-movement), 0)
                conveyorArea.DrawImage(conveyor, &options)

                options.GeoM.Reset()
                options.GeoM.Scale(-1, 1)
                options.GeoM.Translate(131 + float64(cursorPosition) + 3, 85)
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
            options.GeoM.Translate(131 + float64(cursorPosition), 85)
            screen.DrawImage(cursor, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(cancelRect.Min.X), float64(cancelRect.Min.Y))
            screen.DrawImage(cancel[cancelIndex], &options)
        },
    })

    {
        posX := 0
        conveyorRect := image.Rect(131, 85, 131 + 53, 85 + 7)
        elements = append(elements, &uilib.UIElement{
            Rect: conveyorRect,
            Layer: 1,
            LeftClick: func(element *uilib.UIElement){
                cursorPosition = posX - 3
                if cursorPosition < 0 {
                    cursorPosition = 0
                }
            },
            Inside: func(element *uilib.UIElement, x, y int){
                posX = x
            },
        })
    }

    // cancel button
    elements = append(elements, &uilib.UIElement{
        Rect: cancelRect,
        Layer: 1,
        LeftClick: func(element *uilib.UIElement){
            cancelIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            cancelIndex = 0
            ui.RemoveElements(elements)
        },
    })

    // arrow button
    elements = append(elements, &uilib.UIElement{
        Rect: arrowRect,
        Layer: 1,
        LeftClick: func(element *uilib.UIElement){
            isRight = !isRight
        },
    })

    // ok button
    elements = append(elements, &uilib.UIElement{
        Rect: okRect,
        Layer: 1,
        LeftClick: func(element *uilib.UIElement){
            okIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            okIndex = 0
            // FIXME: do transmutation
            ui.RemoveElements(elements)
        },
    })

    return elements
}

func (magic *MagicScreen) MakeUI() *uilib.UI {

    fontLbx, err := magic.Cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    fonts, err := fontLbx.ReadFonts(0)
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

    manaRate := 8
    researchRate := 4
    skillRate := 3

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image) {
            background, err := magic.ImageCache.GetImage("magic.lbx", 0, 0)
            if err == nil {
                var options ebiten.DrawImageOptions
                screen.DrawImage(background, &options)
            }

            gemPositions := []image.Point{
                image.Pt(24, 4),
                image.Pt(101, 4),
                image.Pt(178, 4),
                image.Pt(255, 4),
            }

            for _, position := range gemPositions {
                // FIXME: the gem color is based on what the banner color of the known wizard is
                gemUnknown, err := magic.ImageCache.GetImage("magic.lbx", 6, 0)
                if err == nil {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(position.X), float64(position.Y))
                    screen.DrawImage(gemUnknown, &options)
                }
            }

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })

            normalFont.PrintRight(screen, 56, 160, 1, fmt.Sprintf("%v MP", manaRate))
            normalFont.PrintRight(screen, 103, 160, 1, fmt.Sprintf("%v RP", researchRate))
            normalFont.PrintRight(screen, 151, 160, 1, fmt.Sprintf("%v SP", skillRate))

            smallerFont.Print(screen, 5, 176, 1, fmt.Sprintf("Casting Skill: %v(%v)", 20, 20))
            smallerFont.Print(screen, 5, 183, 1, fmt.Sprintf("Magic Reserve: %v", 90))
            smallerFont.Print(screen, 5, 190, 1, fmt.Sprintf("Power Base: %v", 12))

            smallerFont.Print(screen, 100, 176, 1, fmt.Sprintf("Casting: %v", "None"))
            smallerFont.Print(screen, 100, 183, 1, fmt.Sprintf("Researching: %v", "Whatever"))
            smallerFont.Print(screen, 100, 190, 1, fmt.Sprintf("Summon To: %v", "Somewhere"))
        },
    }

    var elements []*uilib.UIElement

    // FIXME: these default percent should come from the player
    var manaPercent float64 = 1.0 / 3
    var researchPercent float64 = 1.0 / 3
    var skillPercent float64 = 1.0 / 3

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
                *other2= 0
            }
        }
    }

    adjustManaPercent := func(amount float64){
        distribute(amount, &manaPercent, &researchPercent, magic.ResearchLocked, &skillPercent, magic.SkillLocked)
        // log.Printf("mana: %v, research: %v, skill: %v total %v", manaPercent, researchPercent, skillPercent, manaPercent + researchPercent + skillPercent)
    }

    adjustResearchPercent := func(amount float64){
        distribute(amount, &researchPercent, &manaPercent, magic.ManaLocked, &skillPercent, magic.SkillLocked)
    }

    adjustSkillPercent := func(amount float64){
        distribute(amount, &skillPercent, &manaPercent, magic.ManaLocked, &researchPercent, magic.ResearchLocked)
    }

    manaLocked, err := magic.ImageCache.GetImage("magic.lbx", 15, 0)
    if err == nil {
        elements = append(elements, &uilib.UIElement{
            Rect: image.Rect(27, 81, 27 + manaLocked.Bounds().Dx(), 81 + manaLocked.Bounds().Dy() - 2),
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
            LeftClick: func(element *uilib.UIElement){
                if !magic.ManaLocked {
                    // log.Printf("click mana staff at %v", manaStaff.Bounds().Dy() - posY)
                    amount := manaPowerStaff.Bounds().Dy() - posY
                    adjustManaPercent(float64(amount) / float64(manaPowerStaff.Bounds().Dy()))
                }
            },
            Inside: func(element *uilib.UIElement, x, y int){
                posY = y
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(29, 83)
                screen.DrawImage(manaStaff, &options)

                if manaPercent > 0 {
                    length := manaPowerStaff.Bounds().Dy() - int(float64(manaPowerStaff.Bounds().Dy()) * manaPercent)
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
        LeftClick: func(element *uilib.UIElement){
            transmuteElements := magic.MakeTransmuteElements(ui)
            ui.AddElements(transmuteElements)
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            // vector.StrokeRect(screen, float32(transmuteRect.Min.X), float32(transmuteRect.Min.Y), float32(transmuteRect.Dx()), float32(transmuteRect.Bounds().Dy()), 1, color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff}, true)
        },
    })

    // ok button
    okRect := image.Rect(296, 185, 316, 195)
    elements = append(elements, &uilib.UIElement{
        Rect: okRect,
        LeftClick: func(element *uilib.UIElement){
            magic.State = MagicScreenStateDone
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
            LeftClick: func(element *uilib.UIElement){
                magic.ResearchLocked = !magic.ResearchLocked
            },
        })

        researchPowerStaff, _ := magic.ImageCache.GetImage("magic.lbx", 10, 0)
        posY := 0
        staffRect := image.Rect(79, 102, 86, 102 + researchPowerStaff.Bounds().Dy())

        elements = append(elements, &uilib.UIElement{
            Rect: staffRect,
            LeftClick: func(element *uilib.UIElement){
                if !magic.ResearchLocked {
                    // log.Printf("click mana staff at %v", manaStaff.Bounds().Dy() - posY)
                    amount := researchPowerStaff.Bounds().Dy() - posY
                    adjustResearchPercent(float64(amount) / float64(researchPowerStaff.Bounds().Dy()))
                }
            },
            Inside: func(element *uilib.UIElement, x, y int){
                posY = y
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(75, 85)
                screen.DrawImage(researchStaff, &options)

                if researchPercent > 0 {
                    length := researchPowerStaff.Bounds().Dy() - int(float64(researchPowerStaff.Bounds().Dy()) * researchPercent)
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
            LeftClick: func(element *uilib.UIElement){
                magic.SkillLocked = !magic.SkillLocked
            },
        })

        skillPowerStaff, _ := magic.ImageCache.GetImage("magic.lbx", 10, 0)
        posY := 0
        staffRect := image.Rect(126, 102, 132, 102 + skillPowerStaff.Bounds().Dy())

        elements = append(elements, &uilib.UIElement{
            Rect: staffRect,
            LeftClick: func(element *uilib.UIElement){
                if !magic.SkillLocked {
                    // log.Printf("click mana staff at %v", manaStaff.Bounds().Dy() - posY)
                    amount := skillPowerStaff.Bounds().Dy() - posY
                    adjustSkillPercent(float64(amount) / float64(skillPowerStaff.Bounds().Dy()))
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

                if skillPercent > 0 {
                    length := skillPowerStaff.Bounds().Dy() - int(float64(skillPowerStaff.Bounds().Dy()) * skillPercent)
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
