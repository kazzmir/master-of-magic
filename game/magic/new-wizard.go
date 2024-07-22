package main

import (
    "fmt"
    "sync"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"

    "github.com/hajimehoshi/ebiten/v2"
)

type wizardSlot struct {
    Name string
    Background *ebiten.Image
    X int
    Y int
}

type NewWizardScreen struct {
    Background *ebiten.Image
    Slots *ebiten.Image
    Font *font.Font
    loaded sync.Once
    WizardSlots []wizardSlot
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

        screen.Background = loadImage(0, 0)
        screen.Slots = loadImage(8, 0)

        screen.WizardSlots = []wizardSlot{
            wizardSlot{
                Name: "Merlin",
                Background: loadImage(9, 0),
                X: 170,
                Y: 10,
            },
            wizardSlot{
                Name: "Ssr'ra",
                Background: loadImage(16, 0),
                X: 246,
                Y: 10,
            },

        }
    })

    return outError
}

func (screen *NewWizardScreen) Draw(window *ebiten.Image) {
    var options ebiten.DrawImageOptions
    window.DrawImage(screen.Background, &options)

    options.GeoM.Reset()
    options.GeoM.Translate(166, 0)
    window.DrawImage(screen.Slots, &options)

    for _, wizard := range screen.WizardSlots {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(wizard.X), float64(wizard.Y))
        window.DrawImage(wizard.Background, &options)
        screen.Font.PrintCenter(window, float64(wizard.X) + float64(wizard.Background.Bounds().Dx()) / 2, float64(wizard.Y) + 3, 1, wizard.Name)
    }
}

func MakeNewWizardScreen() *NewWizardScreen {
    return &NewWizardScreen{}
}
