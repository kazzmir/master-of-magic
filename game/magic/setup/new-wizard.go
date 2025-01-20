package setup

import (
    "fmt"
    "math"
    "math/rand/v2"
    "image"
    "image/color"
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/draw"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    _ "github.com/hajimehoshi/ebiten/v2/vector"

    "github.com/hajimehoshi/ebiten/v2"
)

const MaxPicks = 11

type WizardSlot struct {
    Name string
    // the block that the wizard's name is printed on in the ui
    Background int
    Base data.WizardBase
    // the portrait of the wizard shown when the user's cursor is on top of their name
    Portrait int
    Books []data.WizardBook
    ExtraAbility WizardAbility
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

func (ability WizardAbility) DependencyExplanation() string {
    switch ability {
        case AbilityAlchemy: return ""
        case AbilityWarlord: return ""
        case AbilityChanneler: return ""
        case AbilityArchmage: return "To select Archmage you need: 4 picks in any Realm of Magic"
        case AbilityArtificer: return ""
        case AbilityConjurer: return ""
        case AbilitySageMaster: return "To select Sage Master you need: 1 pick in any 2 Realms of Magic"
        case AbilityMyrran: return ""
        case AbilityDivinePower: return "To select Divine Power you need: 4 picks in Life Magic"
        case AbilityFamous: return ""
        case AbilityRunemaster: return "To select Runemaster you need: 2 picks in any 3 Realms of Magic"
        case AbilityCharismatic: return ""
        case AbilityChaosMastery: return "To select Chaos Mastery you need: 4 picks in Chaos Magic"
        case AbilityNatureMastery: return "To select Nature Mastery you need: 4 picks in Nature Magic"
        case AbilitySorceryMastery: return "To select Sorcery Mastery you need: 4 picks in Sorcery Magic"
        case AbilityInfernalPower: return "To select Infernal Power you need: 4 picks in Death Magic"
        case AbilityManaFocusing: return "To select Mana Focusing you need: 4 picks in any Realm of Magic"
        case AbilityNodeMastery: return "To select Node Mastery you need: 1 pick in Chaos Magic, 1 pick in Nature Magic, 1 pick in Sorcery Magic"
        case AbilityNone: return ""
        default: return ""
    }
}

// some abilities can only be selected if other properties of the wizard are set
func (ability WizardAbility) SatisifiedDependencies(wizard *WizardCustom) bool {
    switch ability {
        case AbilityAlchemy: return true
        case AbilityWarlord: return true
        case AbilityChanneler: return true
        case AbilityArchmage:
            // need at least 4 books of some magic type
            for _, book := range wizard.Books {
                if book.Count >= 4 {
                    return true
                }
            }
            return false
        case AbilityArtificer: return true
        case AbilityConjurer: return true
        case AbilitySageMaster:
            // need at least 2 books of different magic types
            count := 0
            for _, book := range wizard.Books {
                if book.Count > 0 {
                    count += 1
                }
            }
            return count >= 2
        case AbilityMyrran: return true
        case AbilityDivinePower: return wizard.MagicLevel(data.LifeMagic) >= 4
        case AbilityFamous: return true
        case AbilityRunemaster:
            // need at least 3 books of different magic types with 2 picks per type
            count := 0
            for _, book := range wizard.Books {
                if book.Count >= 2 {
                    count += 1
                }
            }
            return count >= 3
        case AbilityCharismatic: return true
        case AbilityChaosMastery: return wizard.MagicLevel(data.ChaosMagic) >= 4
        case AbilityNatureMastery: return wizard.MagicLevel(data.NatureMagic) >= 4
        case AbilitySorceryMastery: return wizard.MagicLevel(data.SorceryMagic) >= 4
        case AbilityInfernalPower: return wizard.MagicLevel(data.DeathMagic) >= 4
        case AbilityManaFocusing:
            // need at least 4 books of some magic type
            for _, book := range wizard.Books {
                if book.Count >= 4 {
                    return true
                }
            }
            return false
        case AbilityNodeMastery:
            // one pick in chaos, nature, and sorcery
            return wizard.MagicLevel(data.ChaosMagic) >= 1 && wizard.MagicLevel(data.NatureMagic) >= 1 && wizard.MagicLevel(data.SorceryMagic) >= 1
        case AbilityNone: return true
    }

    return true
}

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
    switch ability {
        case AbilityAlchemy: return 1
        case AbilityWarlord: return 2
        case AbilityChanneler: return 2
        case AbilityArchmage: return 1
        case AbilityArtificer: return 1
        case AbilityConjurer: return 1
        case AbilitySageMaster: return 1
        case AbilityMyrran: return 3
        case AbilityDivinePower: return 2
        case AbilityFamous: return 2
        case AbilityRunemaster: return 1
        case AbilityCharismatic: return 1
        case AbilityChaosMastery: return 1
        case AbilityNatureMastery: return 1
        case AbilitySorceryMastery: return 1
        case AbilityInfernalPower: return 2
        case AbilityManaFocusing: return 1
        case AbilityNodeMastery: return 1
        case AbilityNone: return 0
    }

    return 1
}

type NewWizardScreenState int

const (
    NewWizardScreenStateSelectWizard NewWizardScreenState = iota
    NewWizardScreenStateCustomPicture
    NewWizardScreenStateCustomName
    NewWizardScreenStateCustomBooks
    NewWizardScreenStateSelectSpells
    NewWizardScreenStateSelectRace
    NewWizardScreenStateSelectBanner
    NewWizardScreenStateFinished
    NewWizardScreenStateCanceled
)

func (state NewWizardScreenState) String() string {
    switch state {
        case NewWizardScreenStateSelectWizard: return "select wizard"
        case NewWizardScreenStateCustomPicture: return "custom picture"
        case NewWizardScreenStateCustomName: return "custom name"
        case NewWizardScreenStateCustomBooks: return "custom books"
        case NewWizardScreenStateSelectSpells: return "select spells"
        case NewWizardScreenStateSelectRace: return "select race"
        case NewWizardScreenStateSelectBanner: return "select banner"
        case NewWizardScreenStateFinished: return "finished"
    }

    return "?"
}

func DefaultWizardSlots() []WizardSlot {
    return []WizardSlot{
        WizardSlot{
            Name: "Merlin",
            Background: 9,
            Portrait: 0,
            Base: data.WizardMerlin,
            Books: []data.WizardBook{
                data.WizardBook{Magic: data.LifeMagic, Count: 5},
                data.WizardBook{Magic: data.NatureMagic, Count: 5},
            },
            ExtraAbility: AbilitySageMaster,
        },
        WizardSlot{
            Name: "Raven",
            Background: 10,
            Portrait: 1,
            Base: data.WizardRaven,
            Books: []data.WizardBook{
                data.WizardBook{Magic: data.SorceryMagic, Count: 6},
                data.WizardBook{Magic: data.NatureMagic, Count: 5},
            },
            ExtraAbility: AbilityNone,
        },
        WizardSlot{
            Name: "Sharee",
            Background: 11,
            Portrait: 2,
            Base: data.WizardSharee,
            Books: []data.WizardBook{
                data.WizardBook{Magic: data.DeathMagic, Count: 5},
                data.WizardBook{Magic: data.ChaosMagic, Count: 5},
            },
            ExtraAbility: AbilityConjurer,
        },
        WizardSlot{
            Name: "Lo Pan",
            Background: 12,
            Portrait: 3,
            Base: data.WizardLoPan,
            Books: []data.WizardBook{
                data.WizardBook{Magic: data.SorceryMagic, Count: 5},
                data.WizardBook{Magic: data.ChaosMagic, Count: 5},
            },
            ExtraAbility: AbilityChanneler,
        },
        WizardSlot{
            Name: "Jafar",
            Background: 13,
            Portrait: 4,
            Base: data.WizardJafar,
            Books: []data.WizardBook{
                data.WizardBook{Magic: data.SorceryMagic, Count: 10},
            },
            ExtraAbility: AbilityAlchemy,
        },
        WizardSlot{
            Name: "Oberic",
            Background: 14,
            Portrait: 5,
            Base: data.WizardOberic,
            Books: []data.WizardBook{
                data.WizardBook{Magic: data.NatureMagic, Count: 5},
                data.WizardBook{Magic: data.ChaosMagic, Count: 5},
            },
            ExtraAbility: AbilityManaFocusing,
        },
        WizardSlot{
            Name: "Rjak",
            Background: 15,
            Portrait: 6,
            Base: data.WizardRjak,
            Books: []data.WizardBook{
                data.WizardBook{Magic: data.DeathMagic, Count: 9},
            },
            ExtraAbility: AbilityInfernalPower,
        },
        WizardSlot{
            Name: "Sss'ra",
            Background: 16,
            Portrait: 7,
            Base: data.WizardSssra,
            Books: []data.WizardBook{
                data.WizardBook{Magic: data.LifeMagic, Count: 4},
                data.WizardBook{Magic: data.ChaosMagic, Count: 4},
            },
            ExtraAbility: AbilityMyrran,
        },
        WizardSlot{
            Name: "Tauron",
            Background: 17,
            Portrait: 8,
            Base: data.WizardTauron,
            Books: []data.WizardBook{
                data.WizardBook{Magic: data.ChaosMagic, Count: 10},
            },
            ExtraAbility: AbilityChaosMastery,
        },
        WizardSlot{
            Name: "Freya",
            Background: 18,
            Portrait: 9,
            Base: data.WizardFreya,
            Books: []data.WizardBook{
                data.WizardBook{Magic: data.NatureMagic, Count: 10},
            },
            ExtraAbility: AbilityNatureMastery,
        },
        WizardSlot{
            Name: "Horus",
            Background: 19,
            Portrait: 10,
            Base: data.WizardHorus,
            Books: []data.WizardBook{
                data.WizardBook{Magic: data.LifeMagic, Count: 5},
                data.WizardBook{Magic: data.SorceryMagic, Count: 5},
            },
            ExtraAbility: AbilityArchmage,
        },
        WizardSlot{
            Name: "Ariel",
            Background: 20,
            Portrait: 11,
            Base: data.WizardAriel,
            Books: []data.WizardBook{
                data.WizardBook{Magic: data.LifeMagic, Count: 10},
            },
            ExtraAbility: AbilityCharismatic,
        },
        WizardSlot{
            Name: "Tlaloc",
            Background: 21,
            Portrait: 12,
            Base: data.WizardTlaloc,
            Books: []data.WizardBook{
                data.WizardBook{Magic: data.NatureMagic, Count: 4},
                data.WizardBook{Magic: data.DeathMagic, Count: 5},
            },
            ExtraAbility: AbilityWarlord,
        },
        WizardSlot{
            Name: "Kali",
            Background: 22,
            Portrait: 13,
            Base: data.WizardKali,
            Books: []data.WizardBook{
                data.WizardBook{Magic: data.SorceryMagic, Count: 5},
                data.WizardBook{Magic: data.DeathMagic, Count: 5},
            },
            ExtraAbility: AbilityArtificer,
        },
        WizardSlot{
            Name: "Custom",
            Background: 23,
            Books: nil,
        },
    }
}

type WizardCustom struct {
    Name string
    // FIXME: remove portrait
    Portrait int
    Base data.WizardBase
    Abilities []WizardAbility
    Books []data.WizardBook
    StartingSpells spellbook.Spells
    Race data.Race
    Banner data.BannerType
}

func (wizard *WizardCustom) MostBooks() data.MagicType {
    max := 0
    kind := data.LifeMagic

    for _, book := range wizard.Books {
        if book.Count > max {
            max = book.Count
            kind = book.Magic
        }
    }

    return kind
}

/* total number of all magic books combined
 */
func (wizard *WizardCustom) TotalBooks() int {
    count := 0
    for _, book := range wizard.Books {
        count += book.Count
    }

    return count
}

func (wizard *WizardCustom) AbilityEnabled(ability WizardAbility) bool {
    for _, check := range wizard.Abilities {
        if check == ability {
            return true
        }
    }

    return false
}

func (wizard *WizardCustom) EnableAbility(ability WizardAbility) {
    for _, check := range wizard.Abilities {
        if check == ability {
            return
        }
    }

    wizard.Abilities = append(wizard.Abilities, ability)
}

func (wizard *WizardCustom) ToggleAbility(ability WizardAbility, picksLeft int){
    var out []WizardAbility

    found := false

    for _, check := range wizard.Abilities {
        if check == ability {
            found = true
        } else {
            out = append(out, check)
        }
    }

    if !found && ability.PickCost() <= picksLeft {
        out = append(out, ability)
    }

    wizard.Abilities = out
}

func (wizard *WizardCustom) AddMagicLevel(kind data.MagicType, count int){
    for i := range wizard.Books {
        if wizard.Books[i].Magic == kind {
            wizard.Books[i].Count += count
            return
        }
    }

    wizard.Books = append(wizard.Books, data.WizardBook{
        Magic: kind,
        Count: count,
    })
}

func (wizard *WizardCustom) SetMagicLevel(kind data.MagicType, count int){
    var out []data.WizardBook

    found := false

    for _, book := range wizard.Books {
        if book.Magic == kind {
            found = true
            if count != 0 {
                book.Count = count
                out = append(out, book)
            }
        } else {
            out = append(out, book)
        }
    }

    if !found {
        out = append(out, data.WizardBook{
            Magic: kind,
            Count: count,
        })
    }

    // fmt.Printf("Books: %+v\n", out)

    wizard.Books = out
}

func (wizard *WizardCustom) MagicLevel(kind data.MagicType) int {
    for _, book := range wizard.Books {
        if book.Magic == kind {
            return book.Count
        }
    }

    return 0
}


type NewWizardScreen struct {
    LbxCache *lbx.LbxCache

    LbxFonts []*font.LbxFont
    Font *font.Font
    AbilityFont *font.Font
    AbilityFontSelected *font.Font
    AbilityFontAvailable *font.Font
    ErrorFont *font.Font
    NameFont *font.Font
    NameFontBright *font.Font
    SelectFont *font.Font
    WizardSlots []WizardSlot
    ImageCache util.ImageCache

    Spells spellbook.Spells

    Help lbx.Help

    UI *uilib.UI

    BooksOrderSeed1 uint64
    BooksOrderSeed2 uint64

    State NewWizardScreenState

    CustomWizard WizardCustom

    CurrentWizard int
    Active bool

    counter uint64
}

func (screen *NewWizardScreen) MakeCustomNameUI() *uilib.UI {
    const portraitX = 24
    const portraitY = 10

    const nameX = 75
    const nameY = 120

    ui := &uilib.UI{
        Elements: make(map[uilib.UILayer][]*uilib.UIElement),
        Draw: func(this *uilib.UI, window *ebiten.Image){
            var options ebiten.DrawImageOptions
            background, _ := screen.ImageCache.GetImage("newgame.lbx", 0, 0)
            window.DrawImage(background, &options)

            options.GeoM.Translate(portraitX, portraitY)
            portrait, _ := screen.ImageCache.GetImage("wizards.lbx", screen.CustomWizard.Portrait, 0)
            window.DrawImage(portrait, &options)
            screen.Font.PrintCenter(window, nameX, nameY, 1, ebiten.ColorScale{}, screen.CustomWizard.Name)
            screen.SelectFont.PrintCenter(window, 245, 2, 1, ebiten.ColorScale{}, "Wizard's Name")

            options.GeoM.Reset()
            options.GeoM.Translate(184, 20)
            nameBox, _ := screen.ImageCache.GetImage("newgame.lbx", 40, 0)
            window.DrawImage(nameBox, &options)

            name := screen.CustomWizard.Name

            // add blinking _ to show cursor position
            if (screen.counter / 30) % 2 == 0 {
                name += "_"
            }

            screen.NameFontBright.Print(window, 195, 39, 1, ebiten.ColorScale{}, name)

            return
        },
        HandleKeys: func(keys []ebiten.Key){
            for _, key := range keys {
                if inputmanager.IsQuitKey(key) {
                    screen.State = NewWizardScreenStateCustomPicture
                    screen.UI = screen.MakeCustomPictureUI()
                }
            }
        },
    }

    nameRect := image.Rect(184, 20, 184 + 128, 60)

    nameElement := &uilib.UIElement{
        Rect: nameRect,
        /*
        TextEntry: func(this *uilib.UIElement, text []rune){
            str := string(text)
            if str != "" && validNameString(str) {
                screen.CustomWizard.Name += str
                if len(screen.CustomWizard.Name) > MaxNameLength {
                    screen.CustomWizard.Name = screen.CustomWizard.Name[0:MaxNameLength]
                }
            }
        },
        */
        LeftClick: func(this *uilib.UIElement){
        },
        TextEntry: func(this *uilib.UIElement, text string) string {
            // log.Printf("set name to '%v'", text)
            screen.CustomWizard.Name = text

            if len(screen.CustomWizard.Name) > MaxNameLength {
                screen.CustomWizard.Name = screen.CustomWizard.Name[0:MaxNameLength]
            }

            return screen.CustomWizard.Name
        },
        HandleKeys: func(keys []ebiten.Key){
            for _, key := range keys {
                if inputmanager.IsQuitKey(key) {
                    screen.State = NewWizardScreenStateCustomPicture
                    screen.UI = screen.MakeCustomPictureUI()
                }

                switch key {
                    case ebiten.KeyBackspace:
                        length := len(screen.CustomWizard.Name)
                        if length > 0 {
                            length -= 1
                        }
                        screen.CustomWizard.Name = screen.CustomWizard.Name[0:length]
                    case ebiten.KeyEnter:
                        if len(screen.CustomWizard.Name) > 0 {
                            screen.State = NewWizardScreenStateCustomBooks
                            ui.UnfocusElement()
                            screen.UI = screen.MakeCustomWizardBooksUI()
                        }
                        /*
                    case ebiten.KeySpace:
                        screen.CustomWizard.Name += " "
                    default:
                        str := strings.ToLower(key.String())
                        if str != "" && validNameString(str) {
                            screen.CustomWizard.Name += str
                        }
                        */
                }
            }
        },
    }

    ui.AddElement(nameElement)
    ui.FocusElement(nameElement, screen.CustomWizard.Name)

    return ui
}

func (screen *NewWizardScreen) MakeCustomPictureUI() *uilib.UI {

    clickFunc := func(wizard int){
        screen.State = NewWizardScreenStateCustomName
        screen.UI = screen.MakeCustomNameUI()
    }

    insideFunc := func(wizard int){
        screen.CustomWizard.Portrait = screen.WizardSlots[wizard].Portrait
        screen.CustomWizard.Name = screen.WizardSlots[wizard].Name
        screen.CustomWizard.Base = screen.WizardSlots[wizard].Base
    }

    elements := screen.MakeWizardUIElements(clickFunc, insideFunc)

    ui := &uilib.UI{
        Draw: func(this *uilib.UI, window *ebiten.Image){
            var options ebiten.DrawImageOptions
            background, _ := screen.ImageCache.GetImage("newgame.lbx", 0, 0)
            window.DrawImage(background, &options)

            screen.SelectFont.PrintCenter(window, 245, 2, 1, ebiten.ColorScale{}, "Select Wizard")

            const portraitX = 24
            const portraitY = 10

            options.GeoM.Reset()
            options.GeoM.Translate(166, 18)
            customPictureBackground, _ := screen.ImageCache.GetImage("newgame.lbx", 39, 0)
            window.DrawImage(customPictureBackground, &options)

            this.IterateElementsByLayer(func (element *uilib.UIElement){
                element.Draw(element, window)
            })

            portrait, _ := screen.ImageCache.GetImage("wizards.lbx", screen.CustomWizard.Portrait, 0)
            if portrait != nil {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(portraitX, portraitY)
                window.DrawImage(portrait, &options)
            }
        },
        HandleKeys: func(keys []ebiten.Key){
            for _, key := range keys {
                if inputmanager.IsQuitKey(key) {
                    screen.State = NewWizardScreenStateSelectWizard
                    screen.UI = screen.MakeSelectWizardUI()
                }
            }
        },
    }

    ui.SetElementsFromArray(elements)

    return ui
}

func (screen *NewWizardScreen) MakeWizardUIElements(clickFunc func(wizard int), insideFunc func(wizard int)) []*uilib.UIElement {
    var elements []*uilib.UIElement

    top := 28
    space := 22
    columnSpace := 76

    left := 170

    counter := 0
    for column := 0; column < 2; column += 1 {
        for row := 0; row < 7; row++ {
            wizard := counter
            background, _ := screen.ImageCache.GetImage("newgame.lbx", screen.WizardSlots[counter].Background, 0)
            name := screen.WizardSlots[counter].Name
            counter += 1

            x1 := left + column * columnSpace
            y1 := top + row * space
            x2 := x1 + background.Bounds().Dx()
            y2 := y1 + background.Bounds().Dy()

            elements = append(elements, &uilib.UIElement{
                Rect: image.Rect(x1, y1, x2, y2),
                LeftClick: func(this *uilib.UIElement){
                    clickFunc(wizard)
                },
                Inside: func(this *uilib.UIElement, x int, y int){
                    insideFunc(wizard)
                    // screen.CurrentWizard = wizard
                },
                Draw: func(this *uilib.UIElement, window *ebiten.Image){
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(x1), float64(y1))
                    window.DrawImage(background, &options)
                    screen.Font.PrintCenter(window, float64(x1) + float64(background.Bounds().Dx()) / 2, float64(y1) + 3, 1, ebiten.ColorScale{}, name)
                },
            })
        }
    }

    return elements
}

func (screen *NewWizardScreen) MakeSelectWizardUI() *uilib.UI {
    top := 28
    space := 22
    columnSpace := 76

    left := 170

    clickFunc := func(wizard int){
        screen.CustomWizard.Name = screen.WizardSlots[wizard].Name
        screen.CustomWizard.Books = screen.WizardSlots[wizard].Books
        screen.CustomWizard.Base = screen.WizardSlots[wizard].Base
        screen.CustomWizard.Abilities = make([]WizardAbility, 0)
        screen.CustomWizard.Portrait = screen.WizardSlots[wizard].Portrait
        if screen.WizardSlots[wizard].ExtraAbility != AbilityNone {
            screen.CustomWizard.Abilities = append(screen.CustomWizard.Abilities, screen.WizardSlots[wizard].ExtraAbility)
        }

        screen.CustomWizard.StartingSpells.AddAllSpells(GetStartingSpells(&screen.CustomWizard, screen.Spells))

        screen.State = NewWizardScreenStateSelectRace
        screen.UI = screen.MakeSelectRaceUI()
    }

    insideFunc := func(wizard int){
        screen.CurrentWizard = wizard
    }

    elements := screen.MakeWizardUIElements(clickFunc, insideFunc)

    // custom element
    elements = append(elements, (func () *uilib.UIElement {
        background, _ := screen.ImageCache.GetImage("newgame.lbx", screen.WizardSlots[len(elements)].Background, 0)
        x1 := left + columnSpace
        y1 := top + 7 * space
        x2 := x1 + background.Bounds().Dx()
        y2 := y1 + background.Bounds().Dy()

        return &uilib.UIElement{
            Rect: image.Rect(x1, y1, x2, y2),
            LeftClick: func(this *uilib.UIElement){
                screen.State = NewWizardScreenStateCustomPicture
                screen.UI = screen.MakeCustomPictureUI()
            },
            Inside: func(this *uilib.UIElement, x int, y int){
                screen.CurrentWizard = -1
            },
            Draw: func(this *uilib.UIElement, window *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(x1), float64(y1))
                window.DrawImage(background, &options)
                screen.Font.PrintCenter(window, float64(x1) + float64(background.Bounds().Dx()) / 2, float64(y1) + 3, 1, ebiten.ColorScale{}, "Custom")
            },
        }
    })())

    ui := &uilib.UI{
        Draw: func(this *uilib.UI, window *ebiten.Image){
            var options ebiten.DrawImageOptions
            background, _ := screen.ImageCache.GetImage("newgame.lbx", 0, 0)
            window.DrawImage(background, &options)
            screen.SelectFont.PrintCenter(window, 245, 2, 1, ebiten.ColorScale{}, "Select Wizard")

            this.IterateElementsByLayer(func (element *uilib.UIElement){
                element.Draw(element, window)
            })

            if screen.CurrentWizard >= 0 && screen.CurrentWizard < len(screen.WizardSlots) {
                const portraitX = 24
                const portraitY = 10

                const nameX = 75
                const nameY = 120

                portrait, _ := screen.ImageCache.GetImage("wizards.lbx", screen.WizardSlots[screen.CurrentWizard].Portrait, 0)
                if portrait != nil {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(portraitX, portraitY)
                    window.DrawImage(portrait, &options)
                    screen.Font.PrintCenter(window, nameX, nameY, 1, ebiten.ColorScale{}, screen.WizardSlots[screen.CurrentWizard].Name)

                    // screen.DrawBooks(window, 36, 135, screen.WizardSlots[screen.CurrentWizard].Books)
                    options.GeoM.Reset()
                    options.GeoM.Translate(36, 135)
                    draw.DrawBooks(window, options, &screen.ImageCache, screen.WizardSlots[screen.CurrentWizard].Books, screen.BooksOrderRandom())
                    if screen.WizardSlots[screen.CurrentWizard].ExtraAbility != AbilityNone {
                        screen.AbilityFontSelected.Print(window, 12, 180, 1, ebiten.ColorScale{}, screen.WizardSlots[screen.CurrentWizard].ExtraAbility.String())
                    }
                }
            }
        },
        HandleKeys: func(keys []ebiten.Key){
            for _, key := range keys {
                if inputmanager.IsQuitKey(key) {
                    screen.State = NewWizardScreenStateCanceled
                }
            }
        },
    }

    ui.SetElementsFromArray(elements)

    return ui
}

func (screen *NewWizardScreen) BooksOrderRandom() *rand.Rand {
    return rand.New(rand.NewPCG(screen.BooksOrderSeed1, screen.BooksOrderSeed2))
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

func (screen *NewWizardScreen) Update() NewWizardScreenState {
    screen.counter += 1

    if screen.UI != nil {
        screen.UI.StandardUpdate()
    }

    return screen.State
}

func (screen *NewWizardScreen) LoadHelp(cache *lbx.LbxCache) error {
    helpLbx, err := cache.GetLbxFile("HELP.LBX")
    if err != nil {
        return err
    }

    screen.Help, err = helpLbx.ReadHelp(2)
    if err != nil {
        return err
    }

    return nil
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

    fontLbx, err := cache.GetLbxFile("FONTS.LBX")
    if err != nil {
        return err
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        return err
    }

    err = screen.LoadHelp(cache)
    if err != nil {
        return err
    }

    spellsLbx, err := cache.GetLbxFile("SPELLDAT.LBX")
    if err != nil {
        return err
    }

    screen.Spells, err = spellbook.ReadSpells(spellsLbx, 0)
    if err != nil {
        return err
    }

    // FIXME: this is a fudged palette to look like the original, but its probably slightly wrong
    brightYellowPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        // orangish
        color.RGBA{R: 0xff, G: 0xaa, B: 0x00, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
    }

    // FIXME: also a fudged palette
    whitishPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0xc6, G: 0x9d, B: 0x65, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
    }

    pickPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0xc6, G: 0x9d, B: 0x65, A: 0xff},
        color.RGBA{R: 0xc6, G: 0x9d, B: 0x65, A: 0xff},
        color.RGBA{R: 0xc6, G: 0x9d, B: 0x65, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
    }

    transparentPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: uint8(math.Round(0xc6 * 0.3)), G: uint8(math.Round(0x9d * 0.3)), B: uint8(math.Round(0x65 * 0.3)), A: uint8(math.Round(0xff * 0.3))},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
    }

    // FIXME: move this to the font module
    selectYellowPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xfa, G: 0xe1, B: 0x16, A: 0xff},
        color.RGBA{R: 0xff, G: 0xef, B: 0x2f, A: 0xff},
        color.RGBA{R: 0xff, G: 0xef, B: 0x2f, A: 0xff},
        color.RGBA{R: 0xe0, G: 0x8a, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xd2, G: 0x7f, B: 0x0, A: 0xff},
        color.RGBA{R: 0xd2, G: 0x7f, B: 0x0, A: 0xff},
        color.RGBA{R: 0x99, G: 0x4f, B: 0x0, A: 0xff},
        color.RGBA{R: 0x99, G: 0x4f, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0x99, G: 0x4f, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
    }

    screen.LbxFonts = fonts

    screen.Font = font.MakeOptimizedFont(fonts[3])
    screen.WizardSlots = DefaultWizardSlots()

    // FIXME: load with a yellowish palette
    screen.SelectFont = font.MakeOptimizedFontWithPalette(fonts[5], selectYellowPalette)

    screen.AbilityFont = font.MakeOptimizedFontWithPalette(fonts[0], transparentPalette)
    screen.AbilityFontSelected = font.MakeOptimizedFontWithPalette(fonts[0], brightYellowPalette)
    screen.AbilityFontAvailable = font.MakeOptimizedFontWithPalette(fonts[0], whitishPalette)

    // FIXME: use a monochrome color scheme, light-brownish
    screen.NameFont = font.MakeOptimizedFont(fonts[3])
    screen.NameFontBright = font.MakeOptimizedFontWithPalette(fonts[3], pickPalette)

    // FIXME: this should be a fade from bright yellow to dark yellow/orange
    yellowFade := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0xb2, G: 0x8c, B: 0x05, A: 0xff},
        color.RGBA{R: 0xc9, G: 0xa1, B: 0x26, A: 0xff},
        color.RGBA{R: 0xff, G: 0xd3, B: 0x5b, A: 0xff},
        color.RGBA{R: 0xff, G: 0xe8, B: 0x6f, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
    }

    screen.ErrorFont = font.MakeOptimizedFontWithPalette(fonts[4], yellowFade)


    if screen.State == NewWizardScreenStateSelectWizard {
        screen.UI = screen.MakeSelectWizardUI()
    } else if screen.State == NewWizardScreenStateCustomBooks {
        screen.UI = screen.MakeCustomWizardBooksUI()
    } else if screen.State == NewWizardScreenStateSelectSpells {
        screen.UI = screen.MakeSelectSpellsUI()
    } else if screen.State == NewWizardScreenStateSelectRace {
        screen.UI = screen.MakeSelectRaceUI()
    } else if screen.State == NewWizardScreenStateSelectBanner {
        screen.UI = screen.MakeSelectBannerUI()
    }

    return nil
}

func JoinAbilities(abilities []WizardAbility) string {
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

func (screen *NewWizardScreen) MakeCustomWizardBooksUI() *uilib.UI {

    screen.CustomWizard.Abilities = []WizardAbility{}
    screen.CustomWizard.Books = []data.WizardBook{}

    imageCache := util.MakeImageCache(screen.LbxCache)

    picksLeft := func() int {
        picks := MaxPicks

        for _, ability := range screen.CustomWizard.Abilities {
            picks -= ability.PickCost()
        }

        for _, book := range screen.CustomWizard.Books {
            picks -= book.Count
        }

        return picks
    }

    var elements []*uilib.UIElement

    const bookWidth = 8
    const bookHeight = 20
    const bookX = 197

    type bookData struct {
        Kind data.MagicType
        Help string
        Image *ebiten.Image
        Y int
    }

    lifeBook, _ := imageCache.GetImage("newgame.lbx", 24, 0)
    sorceryBook, _ := imageCache.GetImage("newgame.lbx", 27, 0)
    natureBook, _ := imageCache.GetImage("newgame.lbx", 30, 0)
    deathBook, _ := imageCache.GetImage("newgame.lbx", 33, 0)
    chaosBook, _ := imageCache.GetImage("newgame.lbx", 36, 0)

    books := []bookData{
        bookData{
            Kind: data.LifeMagic,
            Help: "Life Spells",
            Image: lifeBook,
            Y: 49,
        },
        bookData{
            Kind: data.DeathMagic,
            Help: "Death Spells",
            Image: deathBook,
            Y: 75,
        },
        bookData{
            Kind: data.ChaosMagic,
            Help: "Chaos Spells",
            Image: chaosBook,
            Y: 101,
        },
        bookData{
            Kind: data.NatureMagic,
            Help: "Nature Spells",
            Image: natureBook,
            Y: 127,
        },
        bookData{
            Kind: data.SorceryMagic,
            Help: "Sorcery Spells",
            Image: sorceryBook,
            Y: 153,
        },
    }

    // for each magic book, create a UI element that contains the book dimensions and can draw the book

    for _, book := range books {
        bookY := book.Y
        bookMagic := book.Kind
        bookImage := book.Image

        x1 := bookX - bookWidth
        y1 := bookY
        x2 := x1 + bookWidth
        y2 := y1 + bookHeight

        // element to remove all books
        elements = append(elements, &uilib.UIElement{
            Rect: image.Rect(x1, y1, x2, y2),
            LeftClick: func(this *uilib.UIElement){
                screen.CustomWizard.SetMagicLevel(bookMagic, 0)
            },
            Draw: func(this *uilib.UIElement, window *ebiten.Image){
            },
        })

        ghostBooks := -1

        minX := bookX
        maxX := bookX

        for i := 0; i < 11; i++ {
            // Rect image.Rectangle
            // Inside UIInsideElementFunc
            // Click UIClickElementFunc
            // Draw UIDrawFunc

            x1 := bookX + bookWidth * i
            y1 := bookY
            x2 := x1 + bookWidth
            y2 := y1 + bookHeight

            if x1 < minX {
                minX = x1
            }

            if x2 > maxX {
                maxX = x2
            }

            level := i

            element := &uilib.UIElement{
                Rect: image.Rect(x1, y1, x2, y2),
                LeftClick: func(this *uilib.UIElement){

                    // user cannot hold both life and death magic
                    if bookMagic == data.LifeMagic && screen.CustomWizard.MagicLevel(data.DeathMagic) > 0 {
                        screen.UI.AddElement(uilib.MakeErrorElement(screen.UI, screen.LbxCache, &imageCache, "You can not select both Life and Death magic", func(){}))
                        return
                    }

                    if bookMagic == data.DeathMagic && screen.CustomWizard.MagicLevel(data.LifeMagic) > 0 {
                        screen.UI.AddElement(uilib.MakeErrorElement(screen.UI, screen.LbxCache, &imageCache, "You can not select both Life and Death magic", func(){}))
                        return
                    }

                    if level + 1 <= screen.CustomWizard.MagicLevel(bookMagic) {
                        screen.CustomWizard.SetMagicLevel(bookMagic, level+1)
                    } else if picksLeft() == 0 {
                        screen.UI.AddElement(uilib.MakeErrorElement(screen.UI, screen.LbxCache, &imageCache, "You have already made all your picks", func(){}))
                    } else {
                        screen.CustomWizard.SetMagicLevel(bookMagic, level+1)
                        if picksLeft() < 0 {
                            screen.CustomWizard.SetMagicLevel(bookMagic, screen.CustomWizard.MagicLevel(bookMagic) + picksLeft())
                        }
                    }
                },
                RightClick: func(this *uilib.UIElement){
                    helpEntries := screen.Help.GetEntriesByName(book.Help)
                    if helpEntries == nil {
                        return
                    }

                    screen.UI.AddElement(uilib.MakeHelpElement(screen.UI, screen.LbxCache, &imageCache, helpEntries[0]))
                },
                Inside: func(this *uilib.UIElement, x int, y int){
                    // if the user hovers over this element, then draw partially transparent books
                    ghostBooks = level
                },
                Draw: func(this *uilib.UIElement, window *ebiten.Image){
                    if screen.CustomWizard.MagicLevel(bookMagic) > level {
                        var options ebiten.DrawImageOptions
                        options.GeoM.Translate(float64(x1), float64(y1))
                        window.DrawImage(bookImage, &options)
                    } else if ghostBooks >= level {
                        // draw a transparent book that shows what the user would have if they selected this
                        // TODO: use a fragment shader to draw the book in a different color
                        var options ebiten.DrawImageOptions
                        options.ColorScale.Scale(1.4 * 0.5, 1 * 0.5, 1 * 0.5, 0.5)
                        options.GeoM.Translate(float64(x1), float64(y1))
                        window.DrawImage(bookImage, &options)
                    }
                },
            }

            elements = append(elements, element)
        }

        // add a non-drawing UI element that is used to detect if the user is pointing at any of the books
        elements = append(elements, &uilib.UIElement{
            Rect: image.Rect(minX, bookY, maxX, bookY + bookHeight),
            NotInside: func(this *uilib.UIElement){
                ghostBooks = -1
            },
        })
    }

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

    // FIXME: compute this based on the largest string in a single column
    tabs := []float64{172, 210, 260, 320}

    type abilityUI struct {
        Ability WizardAbility
        Length int
        X float64
        Y float64
    }

    // iterator to produce all ability ui element positions
    produceAbilityPositions := func() chan abilityUI {
        out := make(chan abilityUI)

        go func(){
            defer close(out)
            const topY = 5
            const veriticalGap = 7
            const maxY = 45

            tab := 0
            y := topY

            for _, ability := range abilities {
                out <- abilityUI{
                    Ability: ability,
                    Length: int(tabs[tab+1] - tabs[tab]),
                    X: tabs[tab],
                    Y: float64(y),
                }

                y += veriticalGap
                if y >= maxY {
                    tab += 1
                    y = topY
                }
            }
        }()

        return out
    }

    isAbilityAvailable := func(ability WizardAbility) bool {
        if picksLeft() < ability.PickCost() {
            return false
        }

        return ability.SatisifiedDependencies(&screen.CustomWizard)
    }

    for ability := range produceAbilityPositions() {
        elements = append(elements, &uilib.UIElement{
            Rect: image.Rect(int(ability.X), int(ability.Y), int(ability.X) + ability.Length, int(ability.Y) + screen.AbilityFont.Height()),
            LeftClick: func(this *uilib.UIElement){
                if screen.CustomWizard.AbilityEnabled(ability.Ability) {
                    screen.CustomWizard.ToggleAbility(ability.Ability, picksLeft())
                } else if isAbilityAvailable(ability.Ability) {
                    screen.CustomWizard.ToggleAbility(ability.Ability, picksLeft())
                } else if picksLeft() == 0 {
                    screen.UI.AddElement(uilib.MakeErrorElement(screen.UI, screen.LbxCache, &imageCache, "You have already made all your picks", func(){}))
                } else {
                    if ability.Ability.SatisifiedDependencies(&screen.CustomWizard) {
                        screen.UI.AddElement(uilib.MakeErrorElement(screen.UI, screen.LbxCache, &imageCache, fmt.Sprintf("You don't have enough picks left to make this selection. You need %v picks", 3 - picksLeft()), func(){}))
                    } else {
                        screen.UI.AddElement(uilib.MakeErrorElement(screen.UI, screen.LbxCache, &imageCache, ability.Ability.DependencyExplanation(), func(){}))
                    }
                }
            },
            RightClick: func(this *uilib.UIElement){

                helpEntries := screen.Help.GetEntriesByName(ability.Ability.String())
                if helpEntries == nil {
                    return
                }

                // Hack! There are two FAMOUS entries in help.lbx, one for the ability and one for the spell
                if ability.Ability == AbilityFamous {
                    helpEntries = []lbx.HelpEntry{screen.Help.GetRawEntry(702)}
                }

                screen.UI.AddElement(uilib.MakeHelpElement(screen.UI, screen.LbxCache, &imageCache, helpEntries[0]))
            },
            Draw: func(this *uilib.UIElement, window *ebiten.Image){
                font := screen.AbilityFont

                if screen.CustomWizard.AbilityEnabled(ability.Ability) {
                    var options ebiten.DrawImageOptions
                    checkMark, _ := screen.ImageCache.GetImage("newgame.lbx", 52, 0)
                    options.GeoM.Translate(ability.X - float64(checkMark.Bounds().Dx()) - 1, ability.Y + 1)
                    window.DrawImage(checkMark, &options)
                    font = screen.AbilityFontSelected
                } else if isAbilityAvailable(ability.Ability) {
                    font = screen.AbilityFontAvailable
                }

                font.Print(window, ability.X, ability.Y, 1, ebiten.ColorScale{}, ability.Ability.String())
            },
        })
    }

    // ok button
    okReady, _ := screen.ImageCache.GetImage("newgame.lbx", 42, 0)
    okNotReady, _ := screen.ImageCache.GetImage("newgame.lbx", 43, 0)
    elements = append(elements, &uilib.UIElement{
        Rect: image.Rect(252, 182, 252 + okReady.Bounds().Dx(), 182 + okReady.Bounds().Dy()),
        LeftClick: func(this *uilib.UIElement){
            if picksLeft() == 0 {
                screen.State = NewWizardScreenStateSelectSpells
                screen.UI = screen.MakeSelectSpellsUI()
            } else {
                screen.UI.AddElement(uilib.MakeErrorElement(screen.UI, screen.LbxCache, &imageCache, "You need to make all your picks before you can continue", func(){}))
            }
        },
        RightClick: func(this *uilib.UIElement){
            helpEntries := screen.Help.GetEntriesByName("ok button")
            if helpEntries == nil {
                return
            }

            screen.UI.AddElement(uilib.MakeHelpElement(screen.UI, screen.LbxCache, &imageCache, helpEntries[0]))
        },
        Draw: func(this *uilib.UIElement, window *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(252, 182)
            if picksLeft() == 0 {
                window.DrawImage(okReady, &options)
            } else {
                window.DrawImage(okNotReady, &options)
            }
        },
    })

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, window *ebiten.Image){
            const portraitX = 24
            const portraitY = 10

            const nameX = 75
            const nameY = 120

            var options ebiten.DrawImageOptions
            customWizardBooks, _ := screen.ImageCache.GetImage("newgame.lbx", 41, 0)
            window.DrawImage(customWizardBooks, &options)

            options.GeoM.Translate(portraitX, portraitY)
            portrait, _ := screen.ImageCache.GetImage("wizards.lbx", screen.CustomWizard.Portrait, 0)
            window.DrawImage(portrait, &options)
            screen.Font.PrintCenter(window, nameX, nameY, 1, ebiten.ColorScale{}, screen.CustomWizard.Name)

            options.GeoM.Reset()
            options.GeoM.Translate(37, 135)
            draw.DrawBooks(window, options, &imageCache, screen.CustomWizard.Books, screen.BooksOrderRandom())

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, window)
                }
            })

            screen.AbilityFontSelected.Print(window, 12, 180, 1, ebiten.ColorScale{}, JoinAbilities(screen.CustomWizard.Abilities))
            screen.NameFontBright.PrintCenter(window, 223, 185, 1, ebiten.ColorScale{}, fmt.Sprintf("%v picks", picksLeft()))
        },
        HandleKeys: func(keys []ebiten.Key){
            for _, key := range keys {
                if inputmanager.IsQuitKey(key) {
                    screen.State = NewWizardScreenStateCustomName
                    screen.UI = screen.MakeCustomNameUI()
                }
            }
        },
    }

    ui.SetElementsFromArray(elements)

    return ui

}

// a struct that stores how many spells the user can choose from each rarity
type ChooseSpellInfo struct {
    CommonMax int
    UncommonMax int
    RareMax int

    CommonSpells spellbook.Spells
    UncommonSpells spellbook.Spells
    RareSpells spellbook.Spells
}

func MakeChooseSpellInfo(allSpells spellbook.Spells, magic data.MagicType, books int) ChooseSpellInfo {
    computeCommon := func(books int) int {
        if books == 0 || books == 11 {
            return 0
        }

        if books < 11 {
            return books - 1
        }

        return 0
    }

    computeUncommon := func(books int) int {
        if books == 11 {
            return 2
        }

        return 0
    }

    computeRare := func(books int) int {
        if books == 11 {
            return 1
        }

        return 0
    }

    chooseSpells := func(rarity spellbook.SpellRarity) spellbook.Spells {
        return allSpells.GetSpellsByMagic(magic).GetSpellsByRarity(rarity)
    }

    commonMax := computeCommon(books)
    uncommonMax := computeUncommon(books)
    rareMax := computeRare(books)

    commonSpells := chooseSpells(spellbook.SpellRarityCommon)
    uncommonSpells := chooseSpells(spellbook.SpellRarityUncommon)
    rareSpells := chooseSpells(spellbook.SpellRarityRare)

    return ChooseSpellInfo{
        CommonMax: commonMax,
        UncommonMax: uncommonMax,
        RareMax: rareMax,

        CommonSpells: commonSpells,
        UncommonSpells: uncommonSpells,
        RareSpells: rareSpells,
    }
}

func GetStartingSpells(wizard *WizardCustom, allSpells spellbook.Spells) spellbook.Spells {
    var spellsOut spellbook.Spells
    magicOrder := []data.MagicType{data.LifeMagic, data.DeathMagic, data.ChaosMagic, data.NatureMagic, data.SorceryMagic}

    for _, magic := range magicOrder {
        spellInfo := MakeChooseSpellInfo(allSpells, magic, wizard.MagicLevel(magic))

        if wizard.MagicLevel(magic) == 11 {
            spellsOut.AddAllSpells(spellInfo.CommonSpells)
        }

        // assign common spells
        for _, index := range rand.Perm(len(spellInfo.CommonSpells.Spells))[0:spellInfo.CommonMax] {
            spellsOut.AddSpell(spellInfo.CommonSpells.Spells[index])
        }

        // assign uncommon spells
        for _, index := range rand.Perm(len(spellInfo.UncommonSpells.Spells))[0:spellInfo.UncommonMax] {
            spellsOut.AddSpell(spellInfo.UncommonSpells.Spells[index])
        }

        // assign rare spells
        for _, index := range rand.Perm(len(spellInfo.RareSpells.Spells))[0:spellInfo.RareMax] {
            spellsOut.AddSpell(spellInfo.RareSpells.Spells[index])
        }
    }

    return spellsOut
}

func (screen *NewWizardScreen) MakeSelectSpellsUI() *uilib.UI {

    // for each book of magic the user has create a spell ui that allows the user to select
    // some set of spells, so if the user has 4 nature and 4 chaos, then the user would see
    // 2 separate UI's, one for nature and one for chaos

    // 2 picks = 1 common
    // 3 picks = 2 common
    // 4 picks = 3 common
    // 5 picks = 4 common
    // 6 picks = 5 common
    // 7 picks = 6 common
    // 8 picks = 7 common
    // 9 picks = 8 common
    // 10 picks = 9 common
    // 11 picks = 2 uncommon, 1 rare

    imageCache := util.MakeImageCache(screen.LbxCache)

    magicOrder := []data.MagicType{data.LifeMagic, data.DeathMagic, data.ChaosMagic, data.NatureMagic, data.SorceryMagic}

    // an all black palette
    black := color.RGBA{R: 0, G: 0, B: 0, A: 0xff}
    blackPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        black, black, black, black, black,
    }

    // create a mono-color palette where the color depends on the magic type
    getPalette := func(magic data.MagicType) color.Palette {
        var use color.RGBA
        switch magic {
            case data.LifeMagic: use = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
            case data.DeathMagic: use = color.RGBA{R: 0x80, G: 0x25, B: 0xca, A: 0xff}
            case data.ChaosMagic: use = color.RGBA{R: 0xcc, G: 0x16, B: 0x27, A: 0xff}
            case data.NatureMagic: use = color.RGBA{R: 0x15, G: 0xa5, B: 0x1b, A: 0xff}
            case data.SorceryMagic: use = color.RGBA{R: 0x00, G: 0x60, B: 0xd6, A: 0xff}
        }

        return color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            use, use, use, use, use,
        }
    }

    blackFont := font.MakeOptimizedFontWithPalette(screen.LbxFonts[4], blackPalette)
    shadowDescriptionFont := font.MakeOptimizedFontWithPalette(screen.LbxFonts[3], blackPalette)

    var doNextMagicUI func (magic data.MagicType)

    makeUIForMagic := func (magic data.MagicType) *uilib.UI {
        spellInfo := MakeChooseSpellInfo(screen.Spells, magic, screen.CustomWizard.MagicLevel(magic))

        // if the wizard has all 11 books then they start with knowing all common spells
        if screen.CustomWizard.MagicLevel(magic) == 11 {
            screen.CustomWizard.StartingSpells.AddAllSpells(spellInfo.CommonSpells)
        }

        // number of remaining picks in each rarity category
        commonPicks := spellInfo.CommonMax
        uncommonPicks := spellInfo.UncommonMax
        rarePicks := spellInfo.RareMax

        // assign common spells
        for _, index := range rand.Perm(len(spellInfo.CommonSpells.Spells))[0:spellInfo.CommonMax] {
            screen.CustomWizard.StartingSpells.AddSpell(spellInfo.CommonSpells.Spells[index])
            commonPicks -= 1
        }

        // assign uncommon spells
        for _, index := range rand.Perm(len(spellInfo.UncommonSpells.Spells))[0:spellInfo.UncommonMax] {
            screen.CustomWizard.StartingSpells.AddSpell(spellInfo.UncommonSpells.Spells[index])
            uncommonPicks -= 1
        }

        // assign rare spells
        for _, index := range rand.Perm(len(spellInfo.RareSpells.Spells))[0:spellInfo.RareMax] {
            screen.CustomWizard.StartingSpells.AddSpell(spellInfo.RareSpells.Spells[index])
            rarePicks -= 1
        }

        picksLeft := func() int {
            return commonPicks + uncommonPicks + rarePicks
        }

        var elements []*uilib.UIElement

        // make fonts in the color of the magic book (blue for sorcery, etc)
        titleFont := font.MakeOptimizedFontWithPalette(screen.LbxFonts[4], getPalette(magic))
        descriptionFont := font.MakeOptimizedFontWithPalette(screen.LbxFonts[3], getPalette(magic))

        createSpellElements := func(spells spellbook.Spells, x int, yTop int, picks *int){
            y := yTop
            checkMark, _ := screen.ImageCache.GetImage("newgame.lbx", 52, 0)
            margin := checkMark.Bounds().Dx() + 1
            spellBackground1, _ := screen.ImageCache.GetImage("newgame.lbx", 48, 0)

            width := (spellBackground1.Bounds().Dx() - 2) / 2
            for i, spell := range spells.Spells {
                useX := x
                useY := y

                elements = append(elements, &uilib.UIElement{
                    Rect: image.Rect(int(x), int(y), int(x) + width, int(y) + screen.AbilityFontAvailable.Height()),
                    LeftClick: func(this *uilib.UIElement){
                        if screen.CustomWizard.StartingSpells.HasSpell(spell) {
                            screen.CustomWizard.StartingSpells.RemoveSpell(spell)
                            *picks += 1
                        } else if *picks > 0 {
                            screen.CustomWizard.StartingSpells.AddSpell(spell)
                            *picks -= 1
                        } else {
                            screen.UI.AddElement(uilib.MakeErrorElement(screen.UI, screen.LbxCache, &imageCache, "You have no picks left in this area, to deselect click on a selected item", func(){}))
                        }
                    },
                    RightClick: func(this *uilib.UIElement){
                        helpEntries := screen.Help.GetEntriesByName(spell.Name)
                        if helpEntries == nil {
                            return
                        }

                        screen.UI.AddElement(uilib.MakeHelpElement(screen.UI, screen.LbxCache, &imageCache, helpEntries[0]))
                    },
                    Draw: func(this *uilib.UIElement, window *ebiten.Image){
                        if screen.CustomWizard.StartingSpells.HasSpell(spell) {
                            var options ebiten.DrawImageOptions
                            options.GeoM.Translate(float64(useX), float64(useY))
                            window.DrawImage(checkMark, &options)
                            screen.AbilityFontSelected.Print(window, float64(useX + margin), float64(useY), 1, ebiten.ColorScale{}, spell.Name)
                        } else {
                            screen.AbilityFontAvailable.Print(window, float64(useX + margin), float64(useY), 1, ebiten.ColorScale{}, spell.Name)
                        }
                    },
                })

                y += screen.AbilityFontAvailable.Height() + 1
                if i == 4 {
                    y = yTop
                    x += width
                }
            }
        }

        if spellInfo.CommonMax > 0 {
            x := 169
            top := 28 + descriptionFont.Height() + 3
            createSpellElements(spellInfo.CommonSpells, x, top, &commonPicks)
        }

        if spellInfo.UncommonMax > 0 {
            x := 169
            top := 28 + descriptionFont.Height() + 3
            createSpellElements(spellInfo.UncommonSpells, x, top, &uncommonPicks)
        }

        if spellInfo.RareMax > 0 {
            x := 169
            top := 78 + descriptionFont.Height() + 3
            createSpellElements(spellInfo.RareSpells, x, top, &rarePicks)
        }

        // ok button
        okReady, _ := screen.ImageCache.GetImage("newgame.lbx", 42, 0)
        okNotReady, _ := screen.ImageCache.GetImage("newgame.lbx", 43, 0)
        elements = append(elements, &uilib.UIElement{
            Rect: image.Rect(252, 182, 252 + okReady.Bounds().Dx(), 182 + okReady.Bounds().Dy()),
            LeftClick: func(this *uilib.UIElement){
                if picksLeft() == 0 {
                    doNextMagicUI(magic)
                } else {
                    screen.UI.AddElement(uilib.MakeErrorElement(screen.UI, screen.LbxCache, &imageCache, "You need to make all your picks before you can continue", func(){}))
                }
            },
            RightClick: func(this *uilib.UIElement){
                /*
                helpEntries := screen.Help.GetEntriesByName("ok button")
                if helpEntries == nil {
                    return
                }

                screen.UI.AddElement(screen.makeHelpElement(helpEntries[0]))
                */
            },
            Draw: func(this *uilib.UIElement, window *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(252, 182)
                if picksLeft() == 0 {
                    window.DrawImage(okReady, &options)
                } else {
                    window.DrawImage(okNotReady, &options)
                }
            },
        })

        ui := &uilib.UI{
            Draw: func(ui *uilib.UI, window *ebiten.Image){
                const portraitX = 24
                const portraitY = 10

                const nameX = 75
                const nameY = 120

                var options ebiten.DrawImageOptions
                background, _ := screen.ImageCache.GetImage("newgame.lbx", 0, 0)
                window.DrawImage(background, &options)

                options.GeoM.Translate(portraitX, portraitY)
                portrait, _ := screen.ImageCache.GetImage("wizards.lbx", screen.CustomWizard.Portrait, 0)
                window.DrawImage(portrait, &options)
                screen.Font.PrintCenter(window, nameX, nameY, 1, ebiten.ColorScale{}, screen.CustomWizard.Name)

                options.GeoM.Reset()
                options.GeoM.Translate(36, 135)
                draw.DrawBooks(window, options, &imageCache, screen.CustomWizard.Books, screen.BooksOrderRandom())

                options.GeoM.Reset()
                options.GeoM.Translate(196, 180)
                pickOkSlot, _  := screen.ImageCache.GetImage("newgame.lbx", 51, 0)
                window.DrawImage(pickOkSlot, &options)

                titleX := 240
                titleY := 5

                blackFont.PrintCenter(window, float64(titleX + 1), float64(titleY + 1), 1, ebiten.ColorScale{}, fmt.Sprintf("Select %v Spells", magic.String()))
                titleFont.PrintCenter(window, float64(titleX), float64(titleY), 1, ebiten.ColorScale{}, fmt.Sprintf("Select %v Spells", magic.String()))

                options.GeoM.Reset()
                options.GeoM.Translate(180, 18)
                windyBorder, _ := screen.ImageCache.GetImage("newgame.lbx", 47, 0)
                window.DrawImage(windyBorder, &options)

                screen.AbilityFontSelected.Print(window, 12, 180, 1, ebiten.ColorScale{}, JoinAbilities(screen.CustomWizard.Abilities))
                screen.NameFontBright.PrintCenter(window, 223, 185, 1, ebiten.ColorScale{}, fmt.Sprintf("%v picks", picksLeft()))

                showDescription := func(y float64, text string, background *ebiten.Image){
                    descriptionX := float64(167)

                    shadowDescriptionFont.Print(window, descriptionX+1, y + 1, 1, ebiten.ColorScale{}, text)
                    descriptionFont.Print(window, descriptionX, y, 1, ebiten.ColorScale{}, text)

                    boxY := y + float64(descriptionFont.Height()) + 1

                    options.GeoM.Reset()
                    options.GeoM.Translate(descriptionX, boxY)
                    window.DrawImage(background, &options)
                }

                spellBackground1, _ := screen.ImageCache.GetImage("newgame.lbx", 48, 0)
                if spellInfo.CommonMax > 0 {
                    showDescription(28, fmt.Sprintf("Common: %v", spellInfo.CommonMax), spellBackground1)
                }

                if spellInfo.UncommonMax > 0 {
                    showDescription(28, fmt.Sprintf("Uncommon: %v", spellInfo.UncommonMax), spellBackground1)
                }

                if spellInfo.RareMax > 0 {
                    spellBackground2, _ := screen.ImageCache.GetImage("newgame.lbx", 49, 0)
                    showDescription(78, fmt.Sprintf("Rare: %v", spellInfo.RareMax), spellBackground2)
                }

                ui.IterateElementsByLayer(func (element *uilib.UIElement){
                    if element.Draw != nil {
                        element.Draw(element, window)
                    }
                })
            },
            HandleKeys: func(keys []ebiten.Key){
                for _, key := range keys {
                    if inputmanager.IsQuitKey(key) {
                        screen.State = NewWizardScreenStateCustomBooks
                        screen.UI = screen.MakeCustomWizardBooksUI()
                    }
                }
            },
        }

        ui.SetElementsFromArray(elements)

        return ui
    }

    // user has clicked ok, so go to next magic spell selection screen
    // example: wizard has 3 life, 4 chaos. show life screen, show chaos screen, then goto race selection screen
    doNextMagicUI = func(current data.MagicType){
        for i := 0; i < len(magicOrder); i++ {
            if current == magicOrder[i] {
                for j := i + 1; j < len(magicOrder); j++ {
                    if screen.CustomWizard.MagicLevel(magicOrder[j]) > 1 {
                        screen.UI = makeUIForMagic(magicOrder[j])
                        return
                    }
                }
            }
        }

        screen.State = NewWizardScreenStateSelectRace
        screen.UI = screen.MakeSelectRaceUI()
    }

    for _, magic := range magicOrder {
        if screen.CustomWizard.MagicLevel(magic) > 1 {
            return makeUIForMagic(magic)
        }
    }

    // player doesn't have any magic, just go directly to race ui
    return screen.MakeSelectRaceUI()
}

func premultiplyAlpha(c color.RGBA, alpha float32) color.RGBA {
    return color.RGBA{
        R: uint8(float32(c.R) * alpha),
        G: uint8(float32(c.G) * alpha),
        B: uint8(float32(c.B) * alpha),
        A: uint8(float32(c.A) * alpha),
    }
}

func (screen *NewWizardScreen) MakeSelectRaceUI() *uilib.UI {

    imageCache := util.MakeImageCache(screen.LbxCache)

    black := color.RGBA{R: 0, G: 0, B: 0, A: 0xff}
    blackPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        black, black, black, black, black,
    }

    raceColor := color.RGBA{R: 0xc1, G: 0x7a, B: 0x23, A: 0xff}
    racePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        raceColor, raceColor, raceColor,
        raceColor, raceColor, raceColor,
    }

    raceShadowFont := font.MakeOptimizedFontWithPalette(screen.LbxFonts[3], blackPalette)
    raceFont := font.MakeOptimizedFontWithPalette(screen.LbxFonts[3], racePalette)

    yellow1 := color.RGBA{R: 0xd6, G: 0xb3, B: 0x85, A: 0xff}
    availablePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        premultiplyAlpha(color.RGBA{R: 0xd6, G: 0xb3, B: 0x85, A: 0xff}, 0.5),
        // color.RGBA{R: 0x85, G: 0x68, B: 0x3d, A: 0xff},
        yellow1, yellow1, yellow1,
        yellow1, yellow1, yellow1,
    }

    yellow2 := premultiplyAlpha(yellow1, 0.3)
    raceUnavailablePalette := color.Palette {
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        yellow2, yellow2, yellow2,
        yellow2, yellow2, yellow2,
    }

    selectColor := color.RGBA{R: 0xfc, G: 0xf3, B: 0x1c, A: 0xff}
    selectPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        premultiplyAlpha(selectColor, 0.5),
        selectColor, selectColor, selectColor,
        selectColor, selectColor, selectColor,
    }

    raceAvailable := font.MakeOptimizedFontWithPalette(screen.LbxFonts[2], availablePalette)
    raceUnavailable := font.MakeOptimizedFontWithPalette(screen.LbxFonts[2], raceUnavailablePalette)
    raceSelect := font.MakeOptimizedFontWithPalette(screen.LbxFonts[2], selectPalette)

    var elements []*uilib.UIElement

    arcanianRaces := data.ArcanianRaces()
    myrranRaces := data.MyrranRaces()

    raceBackground, _ := screen.ImageCache.GetImage("newgame.lbx", 55, 0)
    for i, race := range arcanianRaces {
        yPos := 35 + 1 + i * (raceFont.Height() + 1)

        highlight := false

        elements = append(elements, &uilib.UIElement{
            Rect: image.Rect(210, yPos, 210 + raceBackground.Bounds().Dx(), yPos + raceAvailable.Height()),
            Inside: func(this *uilib.UIElement, x int, y int){
                highlight = true
            },
            NotInside: func(this *uilib.UIElement){
                highlight = false
            },
            LeftClick: func(this *uilib.UIElement){
                screen.CustomWizard.Race = race
                screen.UI = screen.MakeSelectBannerUI()
                screen.State = NewWizardScreenStateSelectBanner
            },
            RightClick: func(this *uilib.UIElement){
                helpEntries := screen.Help.GetEntriesByName(fmt.Sprintf("%v townsfolk", race))
                if helpEntries == nil {
                    log.Printf("Warning: no help found for race '%v'", race)
                    return
                }

                screen.UI.AddElement(uilib.MakeHelpElement(screen.UI, screen.LbxCache, &imageCache, helpEntries[0], helpEntries[1:]...))
            },
            Draw: func(this *uilib.UIElement, window *ebiten.Image){
                if highlight {
                    raceSelect.Print(window, 215, float64(yPos), 1, ebiten.ColorScale{}, race.String())
                } else {
                    raceAvailable.Print(window, 215, float64(yPos), 1, ebiten.ColorScale{}, race.String())
                }
            },
        })
    }

    for i, race := range myrranRaces {
        yPos := 145 + 1 + i * (raceFont.Height() + 1)
        fontUse := raceUnavailable

        if screen.CustomWizard.AbilityEnabled(AbilityMyrran){
            fontUse = raceAvailable
        }

        highlight := false

        elements = append(elements, &uilib.UIElement{
            Rect: image.Rect(210, yPos, 210 + raceBackground.Bounds().Dx(), yPos + raceAvailable.Height()),
            Inside: func(this *uilib.UIElement, x int, y int){
                highlight = true
            },
            NotInside: func(this *uilib.UIElement){
                highlight = false
            },
            LeftClick: func(this *uilib.UIElement){
                if screen.CustomWizard.AbilityEnabled(AbilityMyrran) {
                    screen.CustomWizard.Race = race
                    screen.UI = screen.MakeSelectBannerUI()
                    screen.State = NewWizardScreenStateSelectBanner
                } else {
                    screen.UI.AddElement(uilib.MakeErrorElement(screen.UI, screen.LbxCache, &imageCache, "You can not select a Myrran race unless you have the Myrran special.", func(){}))
                }
            },
            RightClick: func(this *uilib.UIElement){
                helpEntries := screen.Help.GetEntriesByName(fmt.Sprintf("%v townsfolk", race))
                if helpEntries == nil {
                    log.Printf("Warning: no help found for race '%v'", race)
                    return
                }

                screen.UI.AddElement(uilib.MakeHelpElement(screen.UI, screen.LbxCache, &imageCache, helpEntries[0], helpEntries[1:]...))
            },
            Draw: func(this *uilib.UIElement, window *ebiten.Image){
                fontDraw := fontUse
                if screen.CustomWizard.AbilityEnabled(AbilityMyrran) {
                    if highlight {
                        fontDraw = raceSelect
                    } else {
                        fontDraw = raceAvailable
                    }
                }

                fontDraw.Print(window, 215, float64(yPos), 1, ebiten.ColorScale{}, race.String())
            },
        })
    }

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, window *ebiten.Image){
            const portraitX = 24
            const portraitY = 10

            const nameX = 75
            const nameY = 120

            var options ebiten.DrawImageOptions
            background, _ := screen.ImageCache.GetImage("newgame.lbx", 0, 0)
            window.DrawImage(background, &options)

            options.GeoM.Translate(portraitX, portraitY)
            portrait, _ := screen.ImageCache.GetImage("wizards.lbx", screen.CustomWizard.Portrait, 0)
            window.DrawImage(portrait, &options)
            screen.Font.PrintCenter(window, nameX, nameY, 1, ebiten.ColorScale{}, screen.CustomWizard.Name)

            options.GeoM.Reset()
            options.GeoM.Translate(36, 135)
            draw.DrawBooks(window, options, &imageCache, screen.CustomWizard.Books, screen.BooksOrderRandom())

            screen.SelectFont.PrintCenter(window, 245, 2, 1, ebiten.ColorScale{}, "Select Race")

            options.GeoM.Reset()
            options.GeoM.Translate(180, 18)
            windyBorder, _ := screen.ImageCache.GetImage("newgame.lbx", 47, 0)
            window.DrawImage(windyBorder, &options)

            raceShadowFont.PrintCenter(window, 243 + 1, 25, 1, ebiten.ColorScale{}, "Arcanian Races:")
            raceFont.PrintCenter(window, 243, 25, 1, ebiten.ColorScale{}, "Arcanian Races:")

            options.GeoM.Reset()
            options.GeoM.Translate(210, 33)
            window.DrawImage(raceBackground, &options)

            raceShadowFont.PrintCenter(window, 243 + 1, 132, 1, ebiten.ColorScale{}, "Myrran Races:")
            raceFont.PrintCenter(window, 243, 132, 1, ebiten.ColorScale{}, "Myrran Races:")

            screen.AbilityFontSelected.Print(window, 12, 180, 1, ebiten.ColorScale{}, JoinAbilities(screen.CustomWizard.Abilities))

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, window)
                }
            })

        },
        HandleKeys: func(keys []ebiten.Key){
            for _, key := range keys {
                if inputmanager.IsQuitKey(key) {
                    if screen.CurrentWizard == -1 {
                        screen.State = NewWizardScreenStateSelectSpells
                        screen.UI = screen.MakeSelectSpellsUI()
                    } else {
                        screen.State = NewWizardScreenStateSelectWizard
                        screen.UI = screen.MakeSelectWizardUI()
                    }
                }
            }
        },
    }

    ui.SetElementsFromArray(elements)

    return ui
}

func (screen *NewWizardScreen) MakeSelectBannerUI() *uilib.UI {
    var elements []*uilib.UIElement
    imageCache := util.MakeImageCache(screen.LbxCache)

    for i, banner := range []data.BannerType{data.BannerGreen, data.BannerBlue, data.BannerRed, data.BannerPurple, data.BannerYellow} {
        height := 34
        yPos := 24 + i * height
        elements = append(elements, &uilib.UIElement{
            Rect: image.Rect(160, yPos, 320, yPos + height),
            Draw: func(this *uilib.UIElement, window *ebiten.Image){
                // vector.StrokeRect(window, 160, float32(yPos), 160, float32(height), 1, color.RGBA{R: 0xff, G: uint8(i * 20), B: uint8(i * 20), A: 0xff}, true)
            },
            LeftClick: func(this *uilib.UIElement){
                screen.CustomWizard.Banner = banner
                screen.State = NewWizardScreenStateFinished
                // fmt.Printf("choose banner %v\n", banner)
            },
            RightClick: func(this *uilib.UIElement){
                helpEntries := screen.Help.GetEntriesByName("Select a banner")
                if helpEntries == nil {
                    return
                }

                screen.UI.AddElement(uilib.MakeHelpElement(screen.UI, screen.LbxCache, &imageCache, helpEntries[0]))
            },
        })
    }

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, window *ebiten.Image){
            const portraitX = 24
            const portraitY = 10

            const nameX = 75
            const nameY = 120

            var options ebiten.DrawImageOptions
            background, _ := screen.ImageCache.GetImage("newgame.lbx", 0, 0)
            window.DrawImage(background, &options)

            options.GeoM.Translate(portraitX, portraitY)
            portrait, _ := screen.ImageCache.GetImage("wizards.lbx", screen.CustomWizard.Portrait, 0)
            window.DrawImage(portrait, &options)
            screen.Font.PrintCenter(window, nameX, nameY, 1, ebiten.ColorScale{}, screen.CustomWizard.Name)

            options.GeoM.Reset()
            options.GeoM.Translate(36, 135)
            draw.DrawBooks(window, options, &imageCache, screen.CustomWizard.Books, screen.BooksOrderRandom())

            options.GeoM.Reset()
            options.GeoM.Translate(158, 0)
            bannerBackground, _ := screen.ImageCache.GetImage("newgame.lbx", 46, 0)
            window.DrawImage(bannerBackground, &options)

            screen.SelectFont.PrintCenter(window, 245, 2, 1, ebiten.ColorScale{}, "Select Banner")

            screen.AbilityFontSelected.Print(window, 12, 180, 1, ebiten.ColorScale{}, JoinAbilities(screen.CustomWizard.Abilities))

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, window)
                }
            })
        },
        HandleKeys: func(keys []ebiten.Key){
            for _, key := range keys {
                if inputmanager.IsQuitKey(key) {
                    screen.State = NewWizardScreenStateSelectRace
                    screen.UI = screen.MakeSelectRaceUI()
                }
            }
        },
    }

    ui.SetElementsFromArray(elements)

    return ui
}

func (screen *NewWizardScreen) Draw(window *ebiten.Image) {
    if screen.UI != nil {
        screen.UI.Draw(screen.UI, window)
    }
}

func MakeNewWizardScreen(cache *lbx.LbxCache) *NewWizardScreen {
    out := &NewWizardScreen{
        LbxCache: cache,
        ImageCache: util.MakeImageCache(cache),
        CurrentWizard: 0,
        State: NewWizardScreenStateSelectWizard,
    }

    err := out.Load(cache)
    if err != nil {
        log.Printf("Error loading new wizard screen: %v", err)
        return nil
    }

    return out
}
