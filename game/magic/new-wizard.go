package main

import (
    "fmt"
    "sync"
    "math/rand"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"

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
    ExtraAbility string
    X int
    Y int
}

type NewWizardScreen struct {
    Background *ebiten.Image
    Slots *ebiten.Image
    Font *font.Font
    AbilityFont *font.Font
    SelectFont *font.Font
    loaded sync.Once
    WizardSlots []wizardSlot

    LifeBooks [3]*ebiten.Image
    SorceryBooks [3]*ebiten.Image
    NatureBooks [3]*ebiten.Image
    DeathBooks [3]*ebiten.Image
    ChaosBooks [3]*ebiten.Image

    BooksOrder []int

    CurrentWizard int
    Active bool
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

func (screen *NewWizardScreen) Update() {
    mouseX, mouseY := ebiten.CursorPosition()

    for i, wizard := range screen.WizardSlots {
        if mouseX >= wizard.X && mouseX < wizard.X + wizard.Background.Bounds().Dx() &&
            mouseY >= wizard.Y && mouseY < wizard.Y + wizard.Background.Bounds().Dy() {
            screen.CurrentWizard = i
            return
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

        screen.Font = font.MakeOptimizedFont(fonts[3])

        // FIXME: load with a yellowish palette
        screen.SelectFont = font.MakeOptimizedFont(fonts[5])

        // FIXME: load with a yellowish palette
        screen.AbilityFont = font.MakeOptimizedFont(fonts[0])

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
                ExtraAbility: "Sage Master",
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
                ExtraAbility: "Conjurer",
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
                ExtraAbility: "Channeler",
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
                ExtraAbility: "Alchemy",
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
                ExtraAbility: "Mana Focusing",
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
                ExtraAbility: "Infernal Power",
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
                ExtraAbility: "Myrran",
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
                ExtraAbility: "Chaos Mastery",
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
                ExtraAbility: "Nature Mastery",
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
                ExtraAbility: "Archmage",
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
                ExtraAbility: "Charismatic",
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
                ExtraAbility: "Warlord",
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
                ExtraAbility: "Artificer",
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
    })

    return outError
}

func (screen *NewWizardScreen) Draw(window *ebiten.Image) {
    var options ebiten.DrawImageOptions
    window.DrawImage(screen.Background, &options)

    options.GeoM.Reset()
    options.GeoM.Translate(166, 18)
    window.DrawImage(screen.Slots, &options)

    screen.SelectFont.PrintCenter(window, 245, 2, 1, "Select Wizard")

    for _, wizard := range screen.WizardSlots {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(wizard.X), float64(wizard.Y))
        window.DrawImage(wizard.Background, &options)
        screen.Font.PrintCenter(window, float64(wizard.X) + float64(wizard.Background.Bounds().Dx()) / 2, float64(wizard.Y) + 3, 1, wizard.Name)
    }

    if screen.CurrentWizard >= 0 && screen.CurrentWizard < len(screen.WizardSlots) {
        portrait := screen.WizardSlots[screen.CurrentWizard].Portrait
        if portrait != nil {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(24, 10)
            window.DrawImage(portrait, &options)
            screen.Font.PrintCenter(window, 75, 120, 1, screen.WizardSlots[screen.CurrentWizard].Name)

            x := 0
            offsetX := 0
            for _, book := range screen.WizardSlots[screen.CurrentWizard].Books {
                for i := 0; i < book.Count; i++ {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(36 + float64(offsetX), 135)
                    var img *ebiten.Image
                    switch book.Magic {
                        case LifeMagic: img = screen.LifeBooks[screen.BooksOrder[x]]
                        case SorceryMagic: img = screen.SorceryBooks[screen.BooksOrder[x]]
                        case NatureMagic: img = screen.NatureBooks[screen.BooksOrder[x]]
                        case DeathMagic: img = screen.DeathBooks[screen.BooksOrder[x]]
                        case ChaosMagic: img = screen.ChaosBooks[screen.BooksOrder[x]]
                    }
                    window.DrawImage(img, &options)
                    offsetX += img.Bounds().Dx() - 1
                    x += 1
                }
            }

            if screen.WizardSlots[screen.CurrentWizard].ExtraAbility != "" {
                screen.AbilityFont.Print(window, 12, 180, 1, screen.WizardSlots[screen.CurrentWizard].ExtraAbility)
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
    }
}
