package main

import (
    "fmt"
    "sync"
    "math/rand"
    "strings"
    "image"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/hajimehoshi/ebiten/v2/inpututil"

    "github.com/hajimehoshi/ebiten/v2"
)

type MagicType int

const (
    LifeMagic MagicType = iota
    SorceryMagic
    NatureMagic
    DeathMagic
    ChaosMagic
)

/* the number of books a wizard has of a specific magic type */
type wizardBook struct {
    Magic MagicType
    Count int
}

type wizardSlot struct {
    Name string
    // the block that the wizard's name is printed on in the ui
    Background *ebiten.Image
    // the portrait of the wizard shown when the user's cursor is on top of their name
    Portrait *ebiten.Image
    Books []wizardBook
    ExtraAbility WizardAbility
    X int
    Y int
}

type WizardAbility int
const (
        AbilityAlchemy WizardAbility = iota
        AbilityWarlord
        AbilityChanneler
        AbilityArchmage
        AbilityArtificer
        AbilityConjurer
        AbilitySageMaster
        AbilityMyrran
        AbilityDivinePower
        AbilityFamous
        AbilityRunemaster
        AbilityCharismatic
        AbilityChaosMastery
        AbilityNatureMastery
        AbilitySorceryMastery
        AbilityInfernalPower
        AbilityManaFocusing
        AbilityNodeMastery
        AbilityNone
)

func (ability WizardAbility) String() string {
    switch ability {
        case AbilityAlchemy: return "Alchemy"
        case AbilityWarlord: return "Warlord"
        case AbilityChanneler: return "Channeler"
        case AbilityArchmage: return "Archmage"
        case AbilityArtificer: return "Artificer"
        case AbilityConjurer: return "Conjurer"
        case AbilitySageMaster: return "Sage Master"
        case AbilityMyrran: return "Myrran"
        case AbilityDivinePower: return "Divine Power"
        case AbilityFamous: return "Famous"
        case AbilityRunemaster: return "Runemaster"
        case AbilityCharismatic: return "Charismatic"
        case AbilityChaosMastery: return "Chaos Mastery"
        case AbilityNatureMastery: return "Nature Mastery"
        case AbilitySorceryMastery: return "Sorcery Mastery"
        case AbilityInfernalPower: return "Infernal Power"
        case AbilityManaFocusing: return "Mana Focusing"
        case AbilityNodeMastery: return "Node Mastery"
        case AbilityNone: return "invalid"
    }

    return "?"
}

// number of picks this ability costs when choosing a custom wizard
func (ability WizardAbility) PickCost() int {
    // TODO
    return 1
}

type NewWizardScreenState int

const (
    NewWizardScreenStateSelectWizard NewWizardScreenState = iota
    NewWizardScreenStateCustomPicture
    NewWizardScreenStateCustomName
    NewWizardScreenStateCustomAbility
    NewWizardScreenStateCustomBooks
)

type wizardCustom struct {
    Name string
    Portrait *ebiten.Image
    Abilities []WizardAbility
    Books []wizardBook
}

type UIInsideElementFunc func(element *UIElement)
type UIClickElementFunc func(element *UIElement)

type UIElement struct {
    Rect image.Rectangle
    Inside UIInsideElementFunc
    Click UIClickElementFunc
}

type UI struct {
    Elements []*UIElement
}

type NewWizardScreen struct {
    Background *ebiten.Image
    CustomPictureBackground *ebiten.Image
    CustomWizardBooks *ebiten.Image
    Slots *ebiten.Image
    Font *font.Font
    AbilityFont *font.Font
    NameFont *font.Font
    SelectFont *font.Font
    loaded sync.Once
    WizardSlots []wizardSlot

    UI *UI

    NameBox *ebiten.Image

    LifeBooks [3]*ebiten.Image
    SorceryBooks [3]*ebiten.Image
    NatureBooks [3]*ebiten.Image
    DeathBooks [3]*ebiten.Image
    ChaosBooks [3]*ebiten.Image

    BooksOrder []int

    State NewWizardScreenState

    CustomWizard wizardCustom

    CurrentWizard int
    Active bool

    counter uint64
}

func (screen *NewWizardScreen) MakeSelectWizardUI() *UI {
    var elements []*UIElement

    top := 28
    space := 22
    columnSpace := 76

    left := 170

    background := screen.WizardSlots[0].Background

    counter := 0
    for column := 0; column < 2; column += 1 {
        for row := 0; row < 7; row++ {
            wizard := counter
            counter += 1
            elements = append(elements, &UIElement{
                Rect: image.Rect(left + column * columnSpace, top + row * space, left + column * columnSpace + background.Bounds().Dx(), top + row * space + background.Bounds().Dy()),
                Click: func(this *UIElement){
                },
                Inside: func(this *UIElement){
                    screen.CurrentWizard = wizard
                },
            })
        }
    }

    /*
        screen.WizardSlots = []wizardSlot{
            wizardSlot{
                Name: "Merlin",
                Background: loadImage(9, 0),
                Portrait: loadWizardPortrait(0),
                Books: []wizardBook{
                    wizardBook{Magic: LifeMagic, Count: 5},
                    wizardBook{Magic: NatureMagic, Count: 5},
                },
                ExtraAbility: AbilitySageMaster,
                X: 170,
                Y: top,
            },
            wizardSlot{
                Name: "Raven",
                Background: loadImage(10, 0),
                Portrait: loadWizardPortrait(1),
                Books: []wizardBook{
                    wizardBook{Magic: SorceryMagic, Count: 6},
                    wizardBook{Magic: NatureMagic, Count: 5},
                },
                ExtraAbility: AbilityNone,
                X: 170,
                Y: top + 1 * space,
            },
            wizardSlot{
                Name: "Sharee",
                Background: loadImage(11, 0),
                Portrait: loadWizardPortrait(2),
                Books: []wizardBook{
                    wizardBook{Magic: DeathMagic, Count: 5},
                    wizardBook{Magic: ChaosMagic, Count: 5},
                },
                ExtraAbility: AbilityConjurer,
                X: 170,
                Y: top + 2 * space,
            },
            wizardSlot{
                Name: "Lo Pan",
                Background: loadImage(12, 0),
                Portrait: loadWizardPortrait(3),
                Books: []wizardBook{
                    wizardBook{Magic: SorceryMagic, Count: 5},
                    wizardBook{Magic: ChaosMagic, Count: 5},
                },
                ExtraAbility: AbilityChanneler,
                X: 170,
                Y: top + 3 * space,
            },
            wizardSlot{
                Name: "Jafar",
                Background: loadImage(13, 0),
                Portrait: loadWizardPortrait(4),
                Books: []wizardBook{
                    wizardBook{Magic: SorceryMagic, Count: 10},
                },
                ExtraAbility: AbilityAlchemy,
                X: 170,
                Y: top + 4 * space,
            },
            wizardSlot{
                Name: "Oberic",
                Background: loadImage(14, 0),
                Portrait: loadWizardPortrait(5),
                Books: []wizardBook{
                    wizardBook{Magic: NatureMagic, Count: 5},
                    wizardBook{Magic: ChaosMagic, Count: 5},
                },
                ExtraAbility: AbilityManaFocusing,
                X: 170,
                Y: top + 5 * space,
            },
            wizardSlot{
                Name: "Rjak",
                Background: loadImage(15, 0),
                Portrait: loadWizardPortrait(6),
                Books: []wizardBook{
                    wizardBook{Magic: DeathMagic, Count: 9},
                },
                ExtraAbility: AbilityInfernalPower,
                X: 170,
                Y: top + 6 * space,
            },
            wizardSlot{
                Name: "Ssr'ra",
                Background: loadImage(16, 0),
                Portrait: loadWizardPortrait(7),
                Books: []wizardBook{
                    wizardBook{Magic: LifeMagic, Count: 4},
                    wizardBook{Magic: ChaosMagic, Count: 4},
                },
                ExtraAbility: AbilityMyrran,
                X: 246,
                Y: top + 0 * space,
            },
            wizardSlot{
                Name: "Tauron",
                Background: loadImage(17, 0),
                Portrait: loadWizardPortrait(8),
                Books: []wizardBook{
                    wizardBook{Magic: ChaosMagic, Count: 10},
                },
                ExtraAbility: AbilityChaosMastery,
                X: 246,
                Y: top + 1 * space,
            },
            wizardSlot{
                Name: "Freya",
                Background: loadImage(18, 0),
                Portrait: loadWizardPortrait(9),
                Books: []wizardBook{
                    wizardBook{Magic: NatureMagic, Count: 10},
                },
                ExtraAbility: AbilityNatureMastery,
                X: 246,
                Y: top + 2 * space,
            },
            wizardSlot{
                Name: "Horus",
                Background: loadImage(19, 0),
                Portrait: loadWizardPortrait(10),
                Books: []wizardBook{
                    wizardBook{Magic: LifeMagic, Count: 5},
                    wizardBook{Magic: SorceryMagic, Count: 5},
                },
                ExtraAbility: AbilityArchmage,
                X: 246,
                Y: top + 3 * space,
            },
            wizardSlot{
                Name: "Ariel",
                Background: loadImage(20, 0),
                Portrait: loadWizardPortrait(11),
                Books: []wizardBook{
                    wizardBook{Magic: LifeMagic, Count: 10},
                },
                ExtraAbility: AbilityCharismatic,
                X: 246,
                Y: top + 4 * space,
            },
            wizardSlot{
                Name: "Tlaloc",
                Background: loadImage(21, 0),
                Portrait: loadWizardPortrait(12),
                Books: []wizardBook{
                    wizardBook{Magic: NatureMagic, Count: 4},
                    wizardBook{Magic: DeathMagic, Count: 5},
                },
                ExtraAbility: AbilityWarlord,
                X: 246,
                Y: top + 5 * space,
            },
            wizardSlot{
                Name: "Kali",
                Background: loadImage(22, 0),
                Portrait: loadWizardPortrait(13),
                Books: []wizardBook{
                    wizardBook{Magic: SorceryMagic, Count: 5},
                    wizardBook{Magic: DeathMagic, Count: 5},
                },
                ExtraAbility: AbilityArtificer,
                X: 246,
                Y: top + 6 * space,
            },
            wizardSlot{
                Name: "Custom",
                Background: loadImage(23, 0),
                Books: nil,
                Portrait: nil,
                X: 246,
                Y: top + 7 * space,
            },
        }
        */

    return &UI{
        Elements: elements,
    }
}

func (screen *NewWizardScreen) IsActive() bool {
    return screen.Active
}

func (screen *NewWizardScreen) Activate() {
    screen.Active = true
}

func (screen *NewWizardScreen) Deactivate() {
    screen.Active = false
}

func validNameString(text string) bool {
    // FIXME: only allow a-zA-Z0-9 and maybe a few extra chars lke ' and ,
    return true
}

const MaxNameLength = 18

func (screen *NewWizardScreen) Update() {
    screen.counter += 1

    if screen.State == NewWizardScreenStateCustomName {
        keys := make([]ebiten.Key, 0)
        keys = inpututil.AppendJustPressedKeys(keys)

        for _, key := range keys {
            switch key {
                case ebiten.KeyBackspace:
                    length := len(screen.CustomWizard.Name)
                    if length > 0 {
                        length -= 1
                    }
                    screen.CustomWizard.Name = screen.CustomWizard.Name[0:length]
                case ebiten.KeyEnter:
                    screen.State = NewWizardScreenStateCustomBooks
                case ebiten.KeySpace:
                    screen.CustomWizard.Name += " "
                default:
                    str := strings.ToLower(key.String())
                    if str != "" && validNameString(str) {
                        screen.CustomWizard.Name += str
                    }
            }
        }

        if len(screen.CustomWizard.Name) > MaxNameLength {
            screen.CustomWizard.Name = screen.CustomWizard.Name[0:MaxNameLength]
        }

    } else if screen.State == NewWizardScreenStateSelectWizard {
        leftClick := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

        mouseX, mouseY := ebiten.CursorPosition()

        if screen.UI != nil {
            for _, element := range screen.UI.Elements {
                if mouseX >= element.Rect.Min.X && mouseY >= element.Rect.Min.Y && mouseX < element.Rect.Max.X && mouseY <= element.Rect.Max.Y {
                    element.Inside(element)
                    if leftClick {
                        element.Click(element)
                    }
                }
            }
        }

        /*
        for i, wizard := range screen.WizardSlots {
            if mouseX >= wizard.X && mouseX < wizard.X + wizard.Background.Bounds().Dx() &&
                mouseY >= wizard.Y && mouseY < wizard.Y + wizard.Background.Bounds().Dy() {
                screen.CurrentWizard = i

                if leftClick && wizard.Name == "Custom" {
                    screen.State = NewWizardScreenStateCustomPicture
                }

                return
            }
        }
        */
    } else if screen.State == NewWizardScreenStateCustomPicture {
        leftClick := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

        mouseX, mouseY := ebiten.CursorPosition()

        for _, wizard := range screen.WizardSlots {
            if mouseX >= wizard.X && mouseX < wizard.X + wizard.Background.Bounds().Dx() &&
                mouseY >= wizard.Y && mouseY < wizard.Y + wizard.Background.Bounds().Dy() {
                screen.CustomWizard.Portrait = wizard.Portrait
                screen.CustomWizard.Name = wizard.Name
            }
        }

        if leftClick {
            screen.State = NewWizardScreenStateCustomName
        }
    }
}

func (screen *NewWizardScreen) Load(cache *lbx.LbxCache) error {
    // NEWGAME.LBX entry 8 contains boxes for wizard names
    // 9-23 are backgrounds for names
    // 24-26 are life books
    // 27-29 are sorcery/blue books
    // 30-32 are nature/green books
    // 33-35 are death books
    // 36-38 are chaos/red books
    // 41 is custom screen

    var outError error

    screen.loaded.Do(func() {
        fontLbx, err := cache.GetLbxFile("magic-data/FONTS.LBX")
        if err != nil {
            outError = fmt.Errorf("Unable to read FONTS.LBX: %v", err)
            return
        }

        fonts, err := fontLbx.ReadFonts(0)
        if err != nil {
            outError = fmt.Errorf("Unable to read fonts from FONTS.LBX: %v", err)
            return
        }

        screen.Font = font.MakeOptimizedFont(fonts[4])

        // FIXME: load with a yellowish palette
        screen.SelectFont = font.MakeOptimizedFont(fonts[5])

        // FIXME: load with a yellowish palette
        screen.AbilityFont = font.MakeOptimizedFont(fonts[0])

        // FIXME: use a monochrome color scheme, light-brownish
        screen.NameFont = font.MakeOptimizedFont(fonts[3])

        newGameLbx, err := cache.GetLbxFile("magic-data/NEWGAME.LBX")
        if err != nil {
            outError = fmt.Errorf("Unable to load NEWGAME.LBX: %v", err)
            return
        }

        loadImage := func(index int, subIndex int) *ebiten.Image {
            if outError != nil {
                return nil
            }

            sprites, err := newGameLbx.ReadImages(index)
            if err != nil {
                outError = fmt.Errorf("Unable to read background image from NEWGAME.LBX: %v", err)
                return nil
            }

            if len(sprites) <= subIndex {
                outError = fmt.Errorf("Unable to read background image from NEWGAME.LBX: index %d out of range", subIndex)
                return nil
            }

            return ebiten.NewImageFromImage(sprites[subIndex])
        }

        wizardsLbx, err := cache.GetLbxFile("magic-data/WIZARDS.LBX")
        if err != nil {
            outError = err
            return
        }

        loadWizardPortrait := func(index int) *ebiten.Image {
            if outError != nil {
                return nil
            }

            sprites, err := wizardsLbx.ReadImages(index)
            if err != nil {
                outError = fmt.Errorf("Unable to read wizard portrait from WIZARDS.LBX: %v", err)
                return nil
            }

            if len(sprites) == 0 {
                outError = fmt.Errorf("Unable to read wizard portrait from WIZARDS.LBX: no images found")
                return nil
            }

            return ebiten.NewImageFromImage(sprites[0])
        }

        screen.Background = loadImage(0, 0)
        screen.Slots = loadImage(8, 0)
        screen.NameBox = loadImage(40, 0)

        screen.CustomPictureBackground = loadImage(39, 0)
        screen.CustomWizardBooks = loadImage(41, 0)

        for i := 0; i < 3; i++ {
            screen.LifeBooks[i] = loadImage(24 + i, 0)
        }

        for i := 0; i < 3; i++ {
            screen.SorceryBooks[i] = loadImage(27 + i, 0)
        }

        for i := 0; i < 3; i++ {
            screen.NatureBooks[i] = loadImage(30 + i, 0)
        }

        for i := 0; i < 3; i++ {
            screen.DeathBooks[i] = loadImage(33 + i, 0)
        }

        for i := 0; i < 3; i++ {
            screen.ChaosBooks[i] = loadImage(36 + i, 0)
        }

        top := 28
        space := 22

        screen.WizardSlots = []wizardSlot{
            wizardSlot{
                Name: "Merlin",
                Background: loadImage(9, 0),
                Portrait: loadWizardPortrait(0),
                Books: []wizardBook{
                    wizardBook{Magic: LifeMagic, Count: 5},
                    wizardBook{Magic: NatureMagic, Count: 5},
                },
                ExtraAbility: AbilitySageMaster,
                X: 170,
                Y: top,
            },
            wizardSlot{
                Name: "Raven",
                Background: loadImage(10, 0),
                Portrait: loadWizardPortrait(1),
                Books: []wizardBook{
                    wizardBook{Magic: SorceryMagic, Count: 6},
                    wizardBook{Magic: NatureMagic, Count: 5},
                },
                ExtraAbility: AbilityNone,
                X: 170,
                Y: top + 1 * space,
            },
            wizardSlot{
                Name: "Sharee",
                Background: loadImage(11, 0),
                Portrait: loadWizardPortrait(2),
                Books: []wizardBook{
                    wizardBook{Magic: DeathMagic, Count: 5},
                    wizardBook{Magic: ChaosMagic, Count: 5},
                },
                ExtraAbility: AbilityConjurer,
                X: 170,
                Y: top + 2 * space,
            },
            wizardSlot{
                Name: "Lo Pan",
                Background: loadImage(12, 0),
                Portrait: loadWizardPortrait(3),
                Books: []wizardBook{
                    wizardBook{Magic: SorceryMagic, Count: 5},
                    wizardBook{Magic: ChaosMagic, Count: 5},
                },
                ExtraAbility: AbilityChanneler,
                X: 170,
                Y: top + 3 * space,
            },
            wizardSlot{
                Name: "Jafar",
                Background: loadImage(13, 0),
                Portrait: loadWizardPortrait(4),
                Books: []wizardBook{
                    wizardBook{Magic: SorceryMagic, Count: 10},
                },
                ExtraAbility: AbilityAlchemy,
                X: 170,
                Y: top + 4 * space,
            },
            wizardSlot{
                Name: "Oberic",
                Background: loadImage(14, 0),
                Portrait: loadWizardPortrait(5),
                Books: []wizardBook{
                    wizardBook{Magic: NatureMagic, Count: 5},
                    wizardBook{Magic: ChaosMagic, Count: 5},
                },
                ExtraAbility: AbilityManaFocusing,
                X: 170,
                Y: top + 5 * space,
            },
            wizardSlot{
                Name: "Rjak",
                Background: loadImage(15, 0),
                Portrait: loadWizardPortrait(6),
                Books: []wizardBook{
                    wizardBook{Magic: DeathMagic, Count: 9},
                },
                ExtraAbility: AbilityInfernalPower,
                X: 170,
                Y: top + 6 * space,
            },
            wizardSlot{
                Name: "Ssr'ra",
                Background: loadImage(16, 0),
                Portrait: loadWizardPortrait(7),
                Books: []wizardBook{
                    wizardBook{Magic: LifeMagic, Count: 4},
                    wizardBook{Magic: ChaosMagic, Count: 4},
                },
                ExtraAbility: AbilityMyrran,
                X: 246,
                Y: top + 0 * space,
            },
            wizardSlot{
                Name: "Tauron",
                Background: loadImage(17, 0),
                Portrait: loadWizardPortrait(8),
                Books: []wizardBook{
                    wizardBook{Magic: ChaosMagic, Count: 10},
                },
                ExtraAbility: AbilityChaosMastery,
                X: 246,
                Y: top + 1 * space,
            },
            wizardSlot{
                Name: "Freya",
                Background: loadImage(18, 0),
                Portrait: loadWizardPortrait(9),
                Books: []wizardBook{
                    wizardBook{Magic: NatureMagic, Count: 10},
                },
                ExtraAbility: AbilityNatureMastery,
                X: 246,
                Y: top + 2 * space,
            },
            wizardSlot{
                Name: "Horus",
                Background: loadImage(19, 0),
                Portrait: loadWizardPortrait(10),
                Books: []wizardBook{
                    wizardBook{Magic: LifeMagic, Count: 5},
                    wizardBook{Magic: SorceryMagic, Count: 5},
                },
                ExtraAbility: AbilityArchmage,
                X: 246,
                Y: top + 3 * space,
            },
            wizardSlot{
                Name: "Ariel",
                Background: loadImage(20, 0),
                Portrait: loadWizardPortrait(11),
                Books: []wizardBook{
                    wizardBook{Magic: LifeMagic, Count: 10},
                },
                ExtraAbility: AbilityCharismatic,
                X: 246,
                Y: top + 4 * space,
            },
            wizardSlot{
                Name: "Tlaloc",
                Background: loadImage(21, 0),
                Portrait: loadWizardPortrait(12),
                Books: []wizardBook{
                    wizardBook{Magic: NatureMagic, Count: 4},
                    wizardBook{Magic: DeathMagic, Count: 5},
                },
                ExtraAbility: AbilityWarlord,
                X: 246,
                Y: top + 5 * space,
            },
            wizardSlot{
                Name: "Kali",
                Background: loadImage(22, 0),
                Portrait: loadWizardPortrait(13),
                Books: []wizardBook{
                    wizardBook{Magic: SorceryMagic, Count: 5},
                    wizardBook{Magic: DeathMagic, Count: 5},
                },
                ExtraAbility: AbilityArtificer,
                X: 246,
                Y: top + 6 * space,
            },
            wizardSlot{
                Name: "Custom",
                Background: loadImage(23, 0),
                Books: nil,
                Portrait: nil,
                X: 246,
                Y: top + 7 * space,
            },
        }

        // set custom wizard to merlin for now
        screen.CustomWizard.Portrait = screen.WizardSlots[0].Portrait
        screen.CustomWizard.Name = screen.WizardSlots[0].Name
        screen.CustomWizard.Abilities = []WizardAbility{AbilityAlchemy, AbilityConjurer, AbilityFamous}
        screen.CustomWizard.Books = []wizardBook{
            wizardBook{
                Magic: NatureMagic,
                Count: 2,
            },
            wizardBook{
                Magic: DeathMagic,
                Count: 3,
            },
            wizardBook{
                Magic: SorceryMagic,
                Count: 2,
            },
        }

        if screen.State == NewWizardScreenStateSelectWizard {
            screen.UI = screen.MakeSelectWizardUI()
        }
    })

    return outError
}

func joinAbilities(abilities []WizardAbility) string {
    // this could be simplified by iterating backwards through the array and
    // preprending 'and' or ', ' for each element, depending on its index

    if len(abilities) == 0 {
        return ""
    }

    // x
    if len(abilities) == 1 {
        return abilities[0].String()
    }

    // x and y
    if len(abilities) == 2 {
        return fmt.Sprintf("%v and %v", abilities[0], abilities[1])
    }

    // x, y, and z
    out := ""
    for i := 0; i < len(abilities) - 2; i++ {
        out += fmt.Sprintf("%v, ", abilities[i])
    }
    out += fmt.Sprintf("%v and %v", abilities[len(abilities) - 2], abilities[len(abilities) - 1])
    return out
}

func (screen *NewWizardScreen) DrawBooks(window *ebiten.Image, x float64, y float64, books []wizardBook){
    index := 0
    offsetX := 0
    for _, book := range books {
        for i := 0; i < book.Count; i++ {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(x + float64(offsetX), y)
            var img *ebiten.Image
            switch book.Magic {
                case LifeMagic: img = screen.LifeBooks[screen.BooksOrder[index]]
                case SorceryMagic: img = screen.SorceryBooks[screen.BooksOrder[index]]
                case NatureMagic: img = screen.NatureBooks[screen.BooksOrder[index]]
                case DeathMagic: img = screen.DeathBooks[screen.BooksOrder[index]]
                case ChaosMagic: img = screen.ChaosBooks[screen.BooksOrder[index]]
            }

            window.DrawImage(img, &options)
            offsetX += img.Bounds().Dx() - 1
            index += 1
        }
    }
}

func (screen *NewWizardScreen) Draw(window *ebiten.Image) {
    const portraitX = 24
    const portraitY = 10

    const nameX = 75
    const nameY = 120

    if screen.State == NewWizardScreenStateCustomBooks {
        var options ebiten.DrawImageOptions
        window.DrawImage(screen.CustomWizardBooks, &options)

        options.GeoM.Translate(portraitX, portraitY)
        window.DrawImage(screen.CustomWizard.Portrait, &options)
        screen.Font.PrintCenter(window, nameX, nameY, 1, screen.CustomWizard.Name)

        screen.DrawBooks(window, 37, 135, screen.CustomWizard.Books)

        screen.AbilityFont.Print(window, 12, 180, 1, joinAbilities(screen.CustomWizard.Abilities))

        abilities := []WizardAbility{
            AbilityAlchemy,
            AbilityWarlord,
            AbilityChanneler,
            AbilityArchmage,
            AbilityArtificer,
            AbilityConjurer,
            AbilitySageMaster,
            AbilityMyrran,
            AbilityDivinePower,
            AbilityFamous,
            AbilityRunemaster,
            AbilityCharismatic,
            AbilityChaosMastery,
            AbilityNatureMastery,
            AbilitySorceryMastery,
            AbilityInfernalPower,
            AbilityManaFocusing,
            AbilityNodeMastery,
        }

        const topY = 5
        const veriticalGap = 7
        const maxY = 45

        tab := 0
        y := topY

        // FIXME: compute this based on the largest string in a single column
        tabs := []float64{172, 210, 260}

        for _, ability := range abilities {
            screen.AbilityFont.Print(window, tabs[tab], float64(y), 1, ability.String())
            y += veriticalGap
            if y >= maxY {
                tab += 1
                y = topY
            }
        }

        return
    }

    var options ebiten.DrawImageOptions
    window.DrawImage(screen.Background, &options)

    if screen.State == NewWizardScreenStateCustomName {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(portraitX, portraitY)
        window.DrawImage(screen.CustomWizard.Portrait, &options)
        screen.Font.PrintCenter(window, nameX, nameY, 1, screen.CustomWizard.Name)
        screen.SelectFont.PrintCenter(window, 245, 2, 1, "Wizard's Name")

        options.GeoM.Reset()
        options.GeoM.Translate(184, 20)
        window.DrawImage(screen.NameBox, &options)

        name := screen.CustomWizard.Name

        // add blinking _ to show cursor position
        if (screen.counter / 30) % 2 == 0 {
            name += "_"
        }

        screen.NameFont.Print(window, 195, 39, 1, name)

        return
    }

    options.GeoM.Reset()
    options.GeoM.Translate(166, 18)

    switch screen.State {
        case NewWizardScreenStateSelectWizard: 
            window.DrawImage(screen.Slots, &options)
        case NewWizardScreenStateCustomPicture:
            window.DrawImage(screen.CustomPictureBackground, &options)
    }

    screen.SelectFont.PrintCenter(window, 245, 2, 1, "Select Wizard")

    if screen.State == NewWizardScreenStateCustomPicture {
        if screen.CustomWizard.Portrait != nil {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(portraitX, portraitY)
            window.DrawImage(screen.CustomWizard.Portrait, &options)
        }
    }

    for _, wizard := range screen.WizardSlots {
        if screen.State != NewWizardScreenStateCustomPicture || wizard.Name != "Custom" {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(wizard.X), float64(wizard.Y))
            window.DrawImage(wizard.Background, &options)
            screen.Font.PrintCenter(window, float64(wizard.X) + float64(wizard.Background.Bounds().Dx()) / 2, float64(wizard.Y) + 3, 1, wizard.Name)
        }
    }

    if screen.CurrentWizard >= 0 && screen.CurrentWizard < len(screen.WizardSlots) {
        portrait := screen.WizardSlots[screen.CurrentWizard].Portrait
        if portrait != nil {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(portraitX, portraitY)
            window.DrawImage(portrait, &options)
            screen.Font.PrintCenter(window, nameX, nameY, 1, screen.WizardSlots[screen.CurrentWizard].Name)

            if screen.State == NewWizardScreenStateSelectWizard {

                screen.DrawBooks(window, 36, 135, screen.WizardSlots[screen.CurrentWizard].Books)

                if screen.WizardSlots[screen.CurrentWizard].ExtraAbility != AbilityNone {
                    screen.AbilityFont.Print(window, 12, 180, 1, screen.WizardSlots[screen.CurrentWizard].ExtraAbility.String())
                }
            }
        }
    }
}

// create an array of N integers where each integer is some value between 0 and 2
// these values correlate to the index of the book image to draw under the wizard portrait
func randomizeBookOrder(books int) []int {
    order := make([]int, books)
    for i := 0; i < books; i++ {
        order[i] = rand.Intn(3)
    }
    return order
}

func MakeNewWizardScreen() *NewWizardScreen {
    return &NewWizardScreen{
        CurrentWizard: 0,
        BooksOrder: randomizeBookOrder(12),
        State: NewWizardScreenStateSelectWizard,
    }
}
