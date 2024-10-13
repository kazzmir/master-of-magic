package artifact

import (
    "fmt"
    "image"
    "strings"
    "slices"
    "image/color"
    "log"

    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
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

func (a ArtifactType) Name() string {
    switch a {
        case ArtifactTypeSword: return "Sword"
        case ArtifactTypeMace: return "Mace"
        case ArtifactTypeAxe: return "Axe"
        case ArtifactTypeBow: return "Bow"
        case ArtifactTypeStaff: return "Staff"
        case ArtifactTypeWand: return "Wand"
        case ArtifactTypeMisc: return "Misc"
        case ArtifactTypeShield: return "Shield"
        case ArtifactTypeChain: return "Chain"
        case ArtifactTypePlate: return "Plate"
    }

    return ""
}

// the screen can be invoked as either the 'Enchant Item' spell or 'Create Artifact'
type CreationScreen int
const (
    CreationEnchantItem CreationScreen = iota
    CreationCreateArtifact
)

type Power interface {
    String() string
    Cost() int
}

type PowerAttack struct {
    Amount int
}

func (p *PowerAttack) Cost() int {
    switch p.Amount {
        case 1: return 50
        case 2: return 100
        case 3: return 200
        case 4: return 350
        case 5: return 550
        case 6: return 800
    }

    return 800
}

func (p *PowerAttack) String() string {
    return fmt.Sprintf("+%v Attack", p.Amount)
}

type PowerDefense struct {
    Amount int
}

func (p *PowerDefense) Cost() int {
    switch p.Amount {
        case 1: return 50
        case 2: return 100
        case 3: return 200
        case 4: return 350
        case 5: return 550
        case 6: return 800
    }

    return 800
}

func (p *PowerDefense) String() string {
    return fmt.Sprintf("+%v Defense", p.Amount)
}

type PowerToHit struct {
    Amount int
}

func (p *PowerToHit) Cost() int {
    switch p.Amount {
        case 1: return 400
        case 2: return 800
        case 3: return 1200
    }

    return 1200
}

func (p *PowerToHit) String() string {
    return fmt.Sprintf("+%v To Hit", p.Amount)
}

type PowerSpellSkill struct {
    Amount int
}

func (p *PowerSpellSkill) Cost() int {
    switch p.Amount {
        case 5: return 200
        case 10: return 400
        case 15: return 800
        case 20: return 1600
    }

    return 1600
}

func (p *PowerSpellSkill) String() string {
    return fmt.Sprintf("+%v Spell Skill", p.Amount)
}

type PowerSpellSave struct {
    Amount int
}

func (p *PowerSpellSave) Cost() int {
    switch p.Amount {
        case -1: return 100
        case -2: return 200
        case -3: return 400
        case -4: return 800
    }

    return 800
}

func (p *PowerSpellSave) String() string {
    return fmt.Sprintf("%v Spell Save", p.Amount)
}

type PowerResistance struct {
    Amount int
}

func (p *PowerResistance) Cost() int {
    switch p.Amount {
        case 1: return 50
        case 2: return 100
        case 3: return 200
        case 4: return 350
        case 5: return 550
        case 6: return 800
    }

    return 800
}

func (p *PowerResistance) String() string {
    return fmt.Sprintf("+%v Resistance", p.Amount)
}

func (p *PowerMovement) Cost() int {
    switch p.Amount {
        case 1: return 100
        case 2: return 200
        case 3: return 400
        case 4: return 800
    }

    return 800
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

func (p *PowerSpellCharges) Cost() int {
    // FIXME: depends on the spell and the charges
    return 0
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

func (artifact *Artifact) AddPower(power Power) {
    artifact.Powers = append(artifact.Powers, power)
}

func (artifact *Artifact) RemovePower(remove Power) {
    artifact.Powers = slices.DeleteFunc(artifact.Powers, func (power Power) bool {
        return remove == power
    })
}

func (artifact *Artifact) Cost() int {
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
        spell, isSpell := power.(*PowerSpellCharges)
        if isSpell {
            spellCost += spell.Cost()
        } else {
            powerCost += power.Cost()
        }
    }

    // jewelry costs are 2x
    if artifact.Type == ArtifactTypeMisc {
        powerCost *= 2
    }

    return base + powerCost + spellCost
}

func makePowersFull(ui *uilib.UI, cache *lbx.LbxCache, imageCache *util.ImageCache, nameFont *font.Font, powerFont *font.Font, artifactType ArtifactType, picLow int, picHigh int, powerGroups [][]Power, artifact *Artifact) []*uilib.UIElement {
    var elements []*uilib.UIElement

    artifact.Image = picLow

    elements = append(elements, &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(7, 6)
            image, _ := imageCache.GetImage("items.lbx", artifact.Image, 0)
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
            artifact.Image = artifact.Image - 1
            if artifact.Image < picLow {
                artifact.Image = picHigh
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
            artifact.Image = artifact.Image + 1
            if artifact.Image > picHigh {
                artifact.Image = picLow
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(rightRect.Min.X), float64(rightRect.Min.Y))
            image := rightImages[rightIndex]
            screen.DrawImage(image, &options)
        },
    })

    // name field
    nameRect := image.Rect(30, 12, 30 + 130, 12 + nameFont.Height() + 2)
    nameFocused := false
    artifact.Name = artifactType.Name()
    nameColorSource := ebiten.NewImage(1, 1)
    nameColorSource.Fill(color.RGBA{R: 0xf3, G: 0xb3, B: 0x47, A: 0xff})
    elements = append(elements, &uilib.UIElement{
        Rect: nameRect,
        GainFocus: func(element *uilib.UIElement){
            nameFocused = true
        },
        LoseFocus: func(element *uilib.UIElement){
            nameFocused = false
        },
        TextEntry: func(element *uilib.UIElement, char []rune){
            for _, r := range char {
                if len(artifact.Name) < 25 {
                    artifact.Name += string(r)
                }
            }
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
            scale := ebiten.ColorScale{}
            if nameFocused {
                scale.SetR(3)
                scale.SetG(3)
            }

            nameFont.Print(screen, float64(nameRect.Min.X + 1), float64(nameRect.Min.Y + 1), 1, scale, artifact.Name)

            if nameFocused {
                util.DrawTextCursor(screen, nameColorSource, float64(nameRect.Min.X) + 1 + nameFont.MeasureTextWidth(artifact.Name, 1), float64(nameRect.Min.Y) + 1, ui.Counter)
            }
        },
    })

    x := 7
    y := 40
    selectCount := 0
    printRight := false
    for _, group := range powerGroups {
        groupSelect := -1

        // goto the next column
        if y + (powerFont.Height() + 1) * len(group) > data.ScreenHeight - 10 {
            y = 40
            x = 170
            printRight = true
        }

        groupRight := printRight

        var lastPower Power = nil
        for i, power := range group {
            rect := image.Rect(x, y, x + int(powerFont.MeasureTextWidth(power.String(), 1)), y + powerFont.Height())
            if groupRight {
                rect = image.Rect(x - int(powerFont.MeasureTextWidth(power.String(), 1)), y, x, y + powerFont.Height())
            }

            elements = append(elements, &uilib.UIElement{
                Rect: rect,
                LeftClick: func(element *uilib.UIElement){
                    if groupSelect != -1 {
                        if groupSelect == i {
                            groupSelect = -1
                            selectCount -= 1

                            artifact.RemovePower(power)
                            lastPower = nil
                        } else {
                            // something was already selected in this group, so the count doesn't change
                            groupSelect = i
                            artifact.RemovePower(lastPower)
                            artifact.AddPower(power)
                            lastPower = power
                        }
                    } else {
                        if selectCount < 4 {
                            selectCount += 1
                            groupSelect = i
                            artifact.AddPower(power)
                            lastPower = power
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

                    if groupRight {
                        powerFont.PrintRight(screen, float64(rect.Max.X), float64(rect.Min.Y), 1, scale, power.String())
                    } else {
                        powerFont.Print(screen, float64(rect.Min.X), float64(rect.Min.Y), 1, scale, power.String())
                    }
                },
            })

            y += powerFont.Height() + 1
        }

        y += 5
    }

    return elements
}

func makeFonts(cache *lbx.LbxCache) (*font.Font, *font.Font, *font.Font) {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil, nil, nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil, nil, nil
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

    return powerFont, powerFontWhite, nameFont
}

func getSwordPowers(creationType CreationScreen) [][]Power {
    switch creationType {
        case CreationCreateArtifact: return [][]Power{
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
        }
        case CreationEnchantItem: return [][]Power {
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
                &PowerSpellSkill{Amount: 5},
            },
        }
    }

    return nil
}

func getMacePowers(creationType CreationScreen) [][]Power {
    switch creationType {
        case CreationCreateArtifact: return [][]Power{
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
        }
        case CreationEnchantItem: return [][]Power {
            []Power{
                &PowerAttack{Amount: 1},
                &PowerAttack{Amount: 2},
                &PowerAttack{Amount: 3},
            },
            []Power{
                &PowerDefense{Amount: 1},
            },
            []Power{
                &PowerSpellSkill{Amount: 5},
            },
        }
    }

    return nil
}

func getAxePowers(creationType CreationScreen) [][]Power {
    switch creationType {
        case CreationCreateArtifact: return [][]Power{
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
        }
        case CreationEnchantItem: return [][]Power {
            []Power{
                &PowerAttack{Amount: 1},
                &PowerAttack{Amount: 2},
                &PowerAttack{Amount: 3},
            },
            []Power{
                &PowerSpellSkill{Amount: 5},
            },
        }
    }

    return nil
}

func getBowPowers(creationType CreationScreen) [][]Power {
    switch creationType {
        case CreationCreateArtifact: return [][]Power{
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
        }
        case CreationEnchantItem: return [][]Power{
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
                &PowerSpellSkill{Amount: 5},
            },
        }
    }

    return nil
}

func getStaffPowers(creationType CreationScreen) [][]Power {
    switch creationType {
        case CreationCreateArtifact: return [][]Power{
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
        }
        case CreationEnchantItem: return [][]Power{
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
                &PowerSpellSkill{Amount: 5},
            },
            []Power{
                &PowerSpellSave{Amount: -1},
                &PowerSpellSave{Amount: -2},
            },
        }
    }

    return nil
}

func getWandPowers(creationType CreationScreen) [][]Power {
    switch creationType {
        case CreationCreateArtifact: return [][]Power{
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
        }
        case CreationEnchantItem: return [][]Power{
            []Power{
                &PowerAttack{Amount: 1},
                &PowerAttack{Amount: 2},
            },
            []Power{
                &PowerSpellSkill{Amount: 5},
            },
            []Power{
                &PowerSpellSave{Amount: -1},
                &PowerSpellSave{Amount: -2},
            },
        }
    }

    return nil
}

func getMiscPowers(creationType CreationScreen) [][]Power {
    switch creationType {
        case CreationCreateArtifact: return [][]Power{
            []Power{
                &PowerAttack{Amount: 1},
                &PowerAttack{Amount: 2},
                &PowerAttack{Amount: 3},
                &PowerAttack{Amount: 4},
            },
            []Power{
                &PowerDefense{Amount: 1},
                &PowerDefense{Amount: 2},
                &PowerDefense{Amount: 3},
                &PowerDefense{Amount: 4},
            },
            []Power{
                &PowerToHit{Amount: 1},
                &PowerToHit{Amount: 2},
            },
            []Power{
                &PowerMovement{Amount: 1},
                &PowerMovement{Amount: 2},
                &PowerMovement{Amount: 3},
            },
            []Power{
                &PowerResistance{Amount: 1},
                &PowerResistance{Amount: 2},
                &PowerResistance{Amount: 3},
                &PowerResistance{Amount: 4},
                &PowerResistance{Amount: 5},
                &PowerResistance{Amount: 6},
            },
            []Power{
                &PowerSpellSkill{Amount: 5},
                &PowerSpellSkill{Amount: 10},
                &PowerSpellSkill{Amount: 15},
            },
            []Power{
                &PowerSpellSave{Amount: -1},
                &PowerSpellSave{Amount: -2},
                &PowerSpellSave{Amount: -3},
                &PowerSpellSave{Amount: -4},
            },
        }
        case CreationEnchantItem: return [][]Power{
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
                &PowerMovement{Amount: 1},
                &PowerMovement{Amount: 2},
            },
            []Power{
                &PowerResistance{Amount: 1},
                &PowerResistance{Amount: 2},
                &PowerResistance{Amount: 3},
            },
            []Power{
                &PowerSpellSkill{Amount: 5},
            },
            []Power{
                &PowerSpellSave{Amount: -1},
                &PowerSpellSave{Amount: -2},
            },
        }
    }

    return nil
}

func getShieldPowers(creationType CreationScreen) [][]Power {
    switch creationType {
        case CreationCreateArtifact: return [][]Power{
            []Power{
                &PowerDefense{Amount: 1},
                &PowerDefense{Amount: 2},
                &PowerDefense{Amount: 3},
                &PowerDefense{Amount: 4},
                &PowerDefense{Amount: 5},
                &PowerDefense{Amount: 6},
            },
            []Power{
                &PowerMovement{Amount: 1},
                &PowerMovement{Amount: 2},
                &PowerMovement{Amount: 3},
                &PowerMovement{Amount: 4},
            },
            []Power{
                &PowerResistance{Amount: 1},
                &PowerResistance{Amount: 2},
                &PowerResistance{Amount: 3},
                &PowerResistance{Amount: 4},
                &PowerResistance{Amount: 5},
                &PowerResistance{Amount: 6},
            },
        }
        case CreationEnchantItem: return [][]Power{
            []Power{
                &PowerDefense{Amount: 1},
                &PowerDefense{Amount: 2},
                &PowerDefense{Amount: 3},
            },
            []Power{
                &PowerMovement{Amount: 1},
                &PowerMovement{Amount: 2},
            },
            []Power{
                &PowerResistance{Amount: 1},
                &PowerResistance{Amount: 2},
                &PowerResistance{Amount: 3},
            },
        }
    }

    return nil
}

func getChainPowers(creationType CreationScreen) [][]Power {
    switch creationType {
        case CreationCreateArtifact: return [][]Power{
            []Power{
                &PowerDefense{Amount: 1},
                &PowerDefense{Amount: 2},
                &PowerDefense{Amount: 3},
                &PowerDefense{Amount: 4},
                &PowerDefense{Amount: 5},
                &PowerDefense{Amount: 6},
            },
            []Power{
                &PowerMovement{Amount: 1},
                &PowerMovement{Amount: 2},
                &PowerMovement{Amount: 3},
                &PowerMovement{Amount: 4},
            },
            []Power{
                &PowerResistance{Amount: 1},
                &PowerResistance{Amount: 2},
                &PowerResistance{Amount: 3},
                &PowerResistance{Amount: 4},
                &PowerResistance{Amount: 5},
                &PowerResistance{Amount: 6},
            },
        }
        case CreationEnchantItem: return [][]Power{
            []Power{
                &PowerDefense{Amount: 1},
                &PowerDefense{Amount: 2},
                &PowerDefense{Amount: 3},
            },
            []Power{
                &PowerMovement{Amount: 1},
                &PowerMovement{Amount: 2},
            },
            []Power{
                &PowerResistance{Amount: 1},
                &PowerResistance{Amount: 2},
                &PowerResistance{Amount: 3},
            },
        }
    }

    return nil
}

func getPlatePowers(creationType CreationScreen) [][]Power {
    switch creationType {
        case CreationCreateArtifact: return [][]Power{
            []Power{
                &PowerDefense{Amount: 1},
                &PowerDefense{Amount: 2},
                &PowerDefense{Amount: 3},
                &PowerDefense{Amount: 4},
                &PowerDefense{Amount: 5},
                &PowerDefense{Amount: 6},
            },
            []Power{
                &PowerMovement{Amount: 1},
                &PowerMovement{Amount: 2},
                &PowerMovement{Amount: 3},
                &PowerMovement{Amount: 4},
            },
            []Power{
                &PowerResistance{Amount: 1},
                &PowerResistance{Amount: 2},
                &PowerResistance{Amount: 3},
                &PowerResistance{Amount: 4},
                &PowerResistance{Amount: 5},
                &PowerResistance{Amount: 6},
            },
        }
        case CreationEnchantItem: return [][]Power{
            []Power{
                &PowerDefense{Amount: 1},
                &PowerDefense{Amount: 2},
                &PowerDefense{Amount: 3},
            },
            []Power{
                &PowerMovement{Amount: 1},
                &PowerMovement{Amount: 2},
            },
            []Power{
                &PowerResistance{Amount: 1},
                &PowerResistance{Amount: 2},
                &PowerResistance{Amount: 3},
            },
        }
    }

    return nil
}

/* returns the artifact that was created and true,
 * otherwise false for cancelled
 */
func ShowCreateArtifactScreen(yield coroutine.YieldFunc, cache *lbx.LbxCache, creationType CreationScreen, draw *func(*ebiten.Image)) (*Artifact, bool) {
    powerFont, powerFontWhite, nameFont := makeFonts(cache)

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

    var currentArtifact *Artifact

    type PowerArtifact struct {
        Elements []*uilib.UIElement
        Artifact *Artifact
    }

    // ui elements for powers that can be selected, based on what item is selected
    powers := make(map[ArtifactType]PowerArtifact)

    // manually curry
    makePowers := func(picLow int, picHigh int, artifactType ArtifactType, groups [][]Power) PowerArtifact {
        var artifact Artifact
        artifact.Type = artifactType
        elements := makePowersFull(ui, cache, &imageCache, nameFont, powerFont, artifactType, picLow, picHigh, groups, &artifact)
        return PowerArtifact{
            Elements: elements,
            Artifact: &artifact,
        }
    }

    powers[ArtifactTypeSword] = makePowers(0, 8, ArtifactTypeSword, getSwordPowers(creationType))
    powers[ArtifactTypeMace] = makePowers(9, 19, ArtifactTypeMace, getMacePowers(creationType))
    powers[ArtifactTypeAxe] = makePowers(20, 28, ArtifactTypeAxe, getAxePowers(creationType))
    powers[ArtifactTypeBow] = makePowers(29, 37, ArtifactTypeBow, getBowPowers(creationType))
    powers[ArtifactTypeStaff] = makePowers(38, 46, ArtifactTypeStaff, getStaffPowers(creationType))
    powers[ArtifactTypeWand] = makePowers(107, 115, ArtifactTypeWand, getWandPowers(creationType))
    powers[ArtifactTypeMisc] = makePowers(72, 106, ArtifactTypeMisc, getMiscPowers(creationType))
    powers[ArtifactTypeShield] = makePowers(62, 71, ArtifactTypeShield, getShieldPowers(creationType))
    powers[ArtifactTypeChain] = makePowers(47, 54, ArtifactTypeChain, getChainPowers(creationType))
    powers[ArtifactTypePlate] = makePowers(55, 61, ArtifactTypePlate, getPlatePowers(creationType))

    updatePowers := func(index ArtifactType){
        for _, each := range powers {
            ui.RemoveElements(each.Elements)
        }

        ui.AddElements(powers[index].Elements)
        currentArtifact = powers[index].Artifact
    }

    var selectedButton *uilib.UIElement

    makeButton := func(x int, y int, unselected int, selected int, item ArtifactType) *uilib.UIElement {
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

        button := makeButton(x, y, unselectedImageStart + i, selectedImageStart + i, ArtifactType(i+1))
        if selectedButton == nil {
            selectedButton = button
            updatePowers(ArtifactType(i+1))
        }

        ui.AddElement(button)
    }

    ui.AddElement(&uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            powerFontWhite.Print(screen, 198, 185, 1, ebiten.ColorScale{}, fmt.Sprintf("Cost: %v", currentArtifact.Cost()))
        },
    })

    quit := false
    canceled := false

    okButtons, _ := imageCache.GetImages("spellscr.lbx", 24)
    okIndex := 0
    okRect := util.ImageRect(281, 180, okButtons[0])
    ui.AddElement(&uilib.UIElement{
        Rect: okRect,
        LeftClick: func(element *uilib.UIElement){
            okIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            okIndex = 0

            if strings.TrimSpace(currentArtifact.Name) == "" {
                ui.AddElement(uilib.MakeErrorElement(ui, cache, &imageCache, "An item must have a name"))
                return
            }

            if len(currentArtifact.Powers) == 0 {
                yes := func(){
                    quit = true
                    canceled = true
                }

                no := func(){
                }

                ui.AddElements(uilib.MakeConfirmDialog(ui, cache, &imageCache, "This item has no powers. Do you wish to abort the enchantment?", yes, no))
                return
            }

            quit = true
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(okRect.Min.X), float64(okRect.Min.Y))
            image := okButtons[okIndex]
            screen.DrawImage(image, &options)
        },
    })

    *draw = func(screen *ebiten.Image) {
        ui.Draw(ui, screen)
    }

    for !quit {
        ui.StandardUpdate()
        yield()
    }

    return currentArtifact, canceled
}
