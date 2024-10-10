package artifact

import (
    "fmt"
    "image"
    "image/color"
    "log"

    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"

    "github.com/hajimehoshi/ebiten/v2"
    // "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type ArtifactType int
const (
    ArtifactTypeNone ArtifactType = iota
    ArtifactTypeSword
    ArtifactTypeMace
    ArtifactTypeAxe
    ArtifactTypeBow
    ArtifactTypeStaff
    ArtifactTypeWand
    ArtifactTypeMisc
    ArtifactTypeShield
    ArtifactTypeChain
    ArtifactTypePlate
)

type Power interface {
    String() string
}

type PowerAttack struct {
    Amount int
}

func (p *PowerAttack) String() string {
    return fmt.Sprintf("+%v Attack", p.Amount)
}

type PowerDefense struct {
    Amount int
}

func (p *PowerDefense) String() string {
    return fmt.Sprintf("+%v Defense", p.Amount)
}

type PowerToHit struct {
    Amount int
}

func (p *PowerToHit) String() string {
    return fmt.Sprintf("+%v To Hit", p.Amount)
}

type PowerSpellSkill struct {
    Amount int
}

func (p *PowerSpellSkill) String() string {
    return fmt.Sprintf("+%v Spell Skill", p.Amount)
}

type PowerSpellSave struct {
    Amount int
}

func (p *PowerSpellSave) String() string {
    return fmt.Sprintf("%v Spell Save", p.Amount)
}

type PowerResistance struct {
    Amount int
}

func (p *PowerResistance) String() string {
    return fmt.Sprintf("+%v Resistance", p.Amount)
}

type PowerMovement struct {
    Amount int
}

func (p *PowerMovement) String() string {
    return fmt.Sprintf("+%v Movement", p.Amount)
}

type PowerSpellCharges struct {
    Spell spellbook.Spell
    Charges int
}

func (p *PowerSpellCharges) String() string {
    return "Spell Charges"
}

type Artifact struct {
    Type ArtifactType
    Image int
    Name string
    Powers []Power
}

func makePowersFull(ui *uilib.UI, cache *lbx.LbxCache, imageCache *util.ImageCache, powerFont *font.Font, picLow int, picHigh int, powerGroups [][]Power) []*uilib.UIElement {
    var elements []*uilib.UIElement

    currentPicture := picLow

    elements = append(elements, &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(7, 6)
            image, _ := imageCache.GetImage("items.lbx", currentPicture, 0)
            screen.DrawImage(image, &options)
        },
    })

    leftIndex := 0
    leftImages, _ := imageCache.GetImages("spellscr.lbx", 35)
    leftRect := util.ImageRect(5, 24, leftImages[leftIndex])
    elements = append(elements, &uilib.UIElement{
        Rect: leftRect,
        LeftClick: func(element *uilib.UIElement){
            leftIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            leftIndex = 0
            currentPicture = currentPicture - 1
            if currentPicture < picLow {
                currentPicture = picHigh
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(leftRect.Min.X), float64(leftRect.Min.Y))
            image := leftImages[leftIndex]
            screen.DrawImage(image, &options)
        },
    })

    rightIndex := 0
    rightImages, _ := imageCache.GetImages("spellscr.lbx", 36)
    rightRect := util.ImageRect(17, 24, leftImages[rightIndex])
    elements = append(elements, &uilib.UIElement{
        Rect: rightRect,
        LeftClick: func(element *uilib.UIElement){
            rightIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            rightIndex = 0
            currentPicture = currentPicture + 1
            if currentPicture > picHigh {
                currentPicture = picLow
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(rightRect.Min.X), float64(rightRect.Min.Y))
            image := rightImages[rightIndex]
            screen.DrawImage(image, &options)
        },
    })

    x := 7
    y := 40
    selectCount := 0
    for _, group := range powerGroups {

        groupSelect := -1

        for i, power := range group {
            rect := image.Rect(x, y, x + int(powerFont.MeasureTextWidth(power.String(), 1)), y + powerFont.Height())
            elements = append(elements, &uilib.UIElement{
                Rect: rect,
                LeftClick: func(element *uilib.UIElement){
                    if groupSelect != -1 {
                        if groupSelect == i {
                            groupSelect = -1
                            selectCount -= 1
                        } else {
                            // something was already selected in this group, so the count doesn't change
                            groupSelect = i
                        }
                    } else {
                        if selectCount < 4 {
                            selectCount += 1
                            groupSelect = i
                        } else {
                            ui.AddElement(uilib.MakeErrorElement(ui, cache, imageCache, "Only four powers may be enchanted into an item"))
                        }

                    }
                },
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    // draw in bright yellow if selected
                    scale := ebiten.ColorScale{}

                    if groupSelect == i {
                        scale.SetR(3)
                        scale.SetG(3)
                    }

                    powerFont.Print(screen, float64(rect.Min.X), float64(rect.Min.Y), 1, scale, power.String())
                },
            })

            y += powerFont.Height()
        }

        y += 5
    }

    return elements
}

func makePowerFont(cache *lbx.LbxCache) *font.Font {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
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

    return font.MakeOptimizedFontWithPalette(fonts[3], palette)
}

/* returns the artifact that was created and true,
 * otherwise false for cancelled
 */
func ShowCreateArtifactScreen(yield coroutine.YieldFunc, cache *lbx.LbxCache, draw *func(*ebiten.Image)) (*Artifact, bool) {
    quit := false

    powerFont := makePowerFont(cache)

    imageCache := util.MakeImageCache(cache)

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            background, _ := imageCache.GetImage("spellscr.lbx", 13, 0)
            screen.DrawImage(background, &options)

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })
        },
    }

    ui.SetElementsFromArray(nil)

    type ItemIndex int
    const (
        ItemSword ItemIndex = iota
        ItemMace
        ItemAxe
        ItemBow
        ItemStaff
        ItemWand
        ItemMisc
        ItemShield
        ItemChain
        ItemPlate
    )

    // ui elements for powers that can be selected, based on what item is selected
    powers := make(map[ItemIndex][]*uilib.UIElement)

    // manually curry
    makePowers := func(picLow int, picHigh int, groups [][]Power) []*uilib.UIElement {
        return makePowersFull(ui, cache, &imageCache, powerFont, picLow, picHigh, groups)
    }

    powers[ItemSword] = makePowers(0, 8, [][]Power{
        []Power{
            &PowerAttack{Amount: 1},
            &PowerAttack{Amount: 2},
            &PowerAttack{Amount: 3},
        },
        []Power{
            &PowerDefense{Amount: 1},
            &PowerDefense{Amount: 2},
            &PowerDefense{Amount: 3},
        },
        []Power{
            &PowerToHit{Amount: 1},
            &PowerToHit{Amount: 2},
            &PowerToHit{Amount: 3},
        },
        []Power{
            &PowerSpellSkill{Amount: 5},
            &PowerSpellSkill{Amount: 10},
        },
    })

    powers[ItemMace] = makePowers(9, 19, [][]Power{
        []Power{
            &PowerAttack{Amount: 1},
            &PowerAttack{Amount: 2},
            &PowerAttack{Amount: 3},
            &PowerAttack{Amount: 4},
        },
        []Power{
            &PowerDefense{Amount: 1},
        },
        []Power{
            &PowerToHit{Amount: 1},
            &PowerToHit{Amount: 2},
            &PowerToHit{Amount: 3},
        },
        []Power{
            &PowerSpellSkill{Amount: 5},
            &PowerSpellSkill{Amount: 10},
        },
    })

    powers[ItemAxe] = makePowers(20, 28, [][]Power{
        []Power{
            &PowerAttack{Amount: 1},
            &PowerAttack{Amount: 2},
            &PowerAttack{Amount: 3},
            &PowerAttack{Amount: 4},
            &PowerAttack{Amount: 5},
            &PowerAttack{Amount: 6},
        },
        []Power{
            &PowerToHit{Amount: 1},
            &PowerToHit{Amount: 2},
        },
        []Power{
            &PowerSpellSkill{Amount: 5},
            &PowerSpellSkill{Amount: 10},
        },
    })

    powers[ItemBow] = makePowers(29, 37, [][]Power{
        []Power{
            &PowerAttack{Amount: 1},
            &PowerAttack{Amount: 2},
            &PowerAttack{Amount: 3},
            &PowerAttack{Amount: 4},
            &PowerAttack{Amount: 5},
            &PowerAttack{Amount: 6},
        },
        []Power{
            &PowerDefense{Amount: 1},
            &PowerDefense{Amount: 2},
            &PowerDefense{Amount: 3},
        },
        []Power{
            &PowerToHit{Amount: 1},
            &PowerToHit{Amount: 2},
            &PowerToHit{Amount: 3},
        },
        []Power{
            &PowerSpellSkill{Amount: 5},
            &PowerSpellSkill{Amount: 10},
        },
    })

    powers[ItemStaff] = makePowers(38, 46, [][]Power{
        []Power{
            &PowerAttack{Amount: 1},
            &PowerAttack{Amount: 2},
            &PowerAttack{Amount: 3},
            &PowerAttack{Amount: 4},
            &PowerAttack{Amount: 5},
            &PowerAttack{Amount: 6},
        },
        []Power{
            &PowerDefense{Amount: 1},
            &PowerDefense{Amount: 2},
            &PowerDefense{Amount: 3},
        },
        []Power{
            &PowerToHit{Amount: 1},
            &PowerToHit{Amount: 2},
            &PowerToHit{Amount: 3},
        },
        []Power{
            &PowerSpellSkill{Amount: 5},
            &PowerSpellSkill{Amount: 10},
            &PowerSpellSkill{Amount: 15},
            &PowerSpellSkill{Amount: 20},
        },
        []Power{
            &PowerSpellSave{Amount: -1},
            &PowerSpellSave{Amount: -2},
            &PowerSpellSave{Amount: -3},
            &PowerSpellSave{Amount: -4},
        },
    })

    powers[ItemWand] = makePowers(107, 115, [][]Power{
        []Power{
            &PowerAttack{Amount: 1},
            &PowerAttack{Amount: 2},
        },
        []Power{
            &PowerToHit{Amount: 1},
        },
        []Power{
            &PowerSpellSkill{Amount: 5},
            &PowerSpellSkill{Amount: 10},
        },
        []Power{
            &PowerSpellSave{Amount: -1},
            &PowerSpellSave{Amount: -2},
        },
    })

    updatePowers := func(index ItemIndex){
        for _, elements := range powers {
            ui.RemoveElements(elements)
        }

        ui.AddElements(powers[index])
    }

    var selectedButton *uilib.UIElement

    makeButton := func(x int, y int, unselected int, selected int, item ItemIndex) *uilib.UIElement {
        index := 0
        imageRect, _ := imageCache.GetImage("spellscr.lbx", unselected, 0)
        rect := util.ImageRect(x, y, imageRect)
        return &uilib.UIElement{
            Rect: rect,
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
                screen.DrawImage(image, &options)
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

        button := makeButton(x, y, unselectedImageStart + i, selectedImageStart + i, ItemIndex(i))
        if selectedButton == nil {
            selectedButton = button
            updatePowers(ItemIndex(i))
        }

        ui.AddElement(button)
    }

    *draw = func(screen *ebiten.Image) {
        ui.Draw(ui, screen)
    }

    for !quit {
        ui.StandardUpdate()
        yield()
    }

    return nil, false
}
