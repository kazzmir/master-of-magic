package magicview

import (
    // "log"
    "image"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/hajimehoshi/ebiten/v2"
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

    ManaLocked bool
    ResearchLocked bool
    SkillLocked bool
}

func MakeMagicScreen(cache *lbx.LbxCache) *MagicScreen {
    magic := &MagicScreen{
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),

        ManaLocked: false,
        ResearchLocked: false,
        SkillLocked: false,
    }

    magic.UI = magic.MakeUI()

    return magic
}

func (magic *MagicScreen) MakeUI() *uilib.UI {
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
        },
    }

    var elements []*uilib.UIElement

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

    return MagicScreenStateRunning
}

func (magic *MagicScreen) Draw(screen *ebiten.Image){
    magic.UI.Draw(magic.UI, screen)
}
