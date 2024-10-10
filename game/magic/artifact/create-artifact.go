package artifact

import (
    "fmt"
    "image"
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

func makePowers(imageCache *util.ImageCache, powerFont *font.Font, powers []Power) []*uilib.UIElement {
    var elements []*uilib.UIElement
    x := 7
    y := 40
    selectCount := 0
    for _, power := range powers {
        rect := image.Rect(x, y, x + int(powerFont.MeasureTextWidth(power.String(), 1)), y + powerFont.Height())
        selected := false
        elements = append(elements, &uilib.UIElement{
            Rect: rect,
            LeftClick: func(element *uilib.UIElement){
                if selected {
                    selectCount -= 1
                    selected = false
                } else {
                    // FIXME: max of 4
                    selectCount += 1
                    selected = true
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                // draw in bright yellow if selected
                powerFont.Print(screen, float64(rect.Min.X), float64(rect.Min.Y), 1, ebiten.ColorScale{}, power.String())
            },
        })

        y += powerFont.Height()
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

    return font.MakeOptimizedFont(fonts[3])
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

    powers[ItemSword] = makePowers(&imageCache, powerFont, []Power{
        &PowerAttack{Amount: 1},
        &PowerAttack{Amount: 2},
        &PowerAttack{Amount: 3},
        &PowerDefense{Amount: 1},
        &PowerDefense{Amount: 2},
        &PowerDefense{Amount: 3},
        &PowerToHit{Amount: 1},
        &PowerToHit{Amount: 2},
        &PowerToHit{Amount: 3},
        &PowerSpellSkill{Amount: 5},
        &PowerSpellSkill{Amount: 10},
    })

    powers[ItemWand] = makePowers(&imageCache, powerFont, []Power{
        &PowerAttack{Amount: 1},
        &PowerAttack{Amount: 2},
        &PowerToHit{Amount: 1},
        &PowerSpellSkill{Amount: 5},
        &PowerSpellSkill{Amount: 10},
        &PowerSpellSave{Amount: -1},
        &PowerSpellSave{Amount: -2},
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

        ui.AddElement(makeButton(x, y, unselectedImageStart + i, selectedImageStart + i, ItemIndex(i)))
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
