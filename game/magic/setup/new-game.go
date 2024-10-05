package setup

import (
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/util"

    "github.com/hajimehoshi/ebiten/v2"
)

const DifficultyMax = 4
const OpponentsMax = 4
const LandSizeMax = 2
const MagicMax = 2

type NewGameSettings struct {
    Difficulty data.DifficultySetting
    Opponents int
    LandSize int
    Magic data.MagicSetting
}

func (settings *NewGameSettings) DifficultyNext() {
    difficulties := []data.DifficultySetting{
        data.DifficultyIntro,
        data.DifficultyEasy,
        data.DifficultyAverage,
        data.DifficultyHard,
        data.DifficultyExtreme,
        data.DifficultyImpossible,
    }

    for i, diff := range difficulties {
        if diff == settings.Difficulty {
            settings.Difficulty = difficulties[(i + 1) % len(difficulties)]
            return
        }
    }
}

func (settings *NewGameSettings) OpponentsNext() {
    settings.Opponents += 1
    if settings.Opponents > OpponentsMax {
        settings.Opponents = 1
    }
}

func (settings *NewGameSettings) LandSizeNext() {
    settings.LandSize += 1
    if settings.LandSize > LandSizeMax {
        settings.LandSize = 0
    }
}

func (settings *NewGameSettings) MagicNext() {
    switch settings.Magic {
        case data.MagicSettingWeak: settings.Magic = data.MagicSettingNormal
        case data.MagicSettingNormal: settings.Magic = data.MagicSettingPowerful
        case data.MagicSettingPowerful: settings.Magic = data.MagicSettingWeak
    }
}

func (settings *NewGameSettings) DifficultyString() string {
    names := map[data.DifficultySetting]string{
        data.DifficultyIntro: "Intro",
        data.DifficultyEasy: "Easy",
        data.DifficultyAverage: "Average",
        data.DifficultyHard: "Hard",
        data.DifficultyExtreme: "Extreme",
        data.DifficultyImpossible: "Impossible",
    }
    return names[settings.Difficulty]
}

func (settings *NewGameSettings) OpponentsString() string {
    kinds := []string{"One", "Two", "Three", "Four"}
    return kinds[settings.Opponents - 1]
}

func (settings *NewGameSettings) LandSizeString() string {
    kinds := []string{"Small", "Medium", "Large"}
    return kinds[settings.LandSize]
}

func (settings *NewGameSettings) MagicString() string {
    kinds := map[data.MagicSetting]string{
        data.MagicSettingWeak: "Weak",
        data.MagicSettingNormal: "Normal",
        data.MagicSettingPowerful: "Powerful",
    }
    return kinds[settings.Magic]
}

type NewGameState int
const (
    NewGameStateRunning NewGameState = iota
    NewGameStateOk
    NewGameStateCancel
)

type NewGameScreen struct {
    /*
    LbxFile *lbx.LbxFile
    Background *ebiten.Image
    Options *ebiten.Image
    OkButtons []*ebiten.Image
    CancelButtons []*ebiten.Image
    DifficultyBlock *ebiten.Image
    OpponentsBlock *ebiten.Image
    LandSizeBlock *ebiten.Image
    MagicBlock *ebiten.Image
    loaded sync.Once
    Font *font.Font
    */

    Cache *lbx.LbxCache
    ImageCache util.ImageCache

    State NewGameState

    Settings NewGameSettings
    // Active bool

    UI *uilib.UI
}

/*
func (newGameScreen *NewGameScreen) Activate() {
    newGameScreen.Active = true
}

func (newGameScreen *NewGameScreen) Deactivate() {
    newGameScreen.Active = false
}

func (newGameScreen *NewGameScreen) IsActive() bool {
    return newGameScreen.Active
}
*/

/*
func (newGameScreen *NewGameScreen) Load(cache *lbx.LbxCache) error {
    var outError error = nil

    newGameScreen.loaded.Do(func() {
        fontLbx, err := cache.GetLbxFile("FONTS.LBX")
        if err != nil {
            outError = fmt.Errorf("Unable to read FONTS.LBX: %v", err)
            return
        }

        fonts, err := font.ReadFonts(fontLbx, 0)
        if err != nil {
            outError = fmt.Errorf("Unable to read fonts from FONTS.LBX: %v", err)
            return
        }

        newGameScreen.Font = font.MakeOptimizedFont(fonts[3])

        newGameLbx, err := cache.GetLbxFile("NEWGAME.LBX")
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

        newGameScreen.Background = loadImage(0, 0)
        newGameScreen.Options = loadImage(1, 0)

        newGameScreen.OkButtons = make([]*ebiten.Image, 2)
        newGameScreen.OkButtons[0] = loadImage(2, 0)
        newGameScreen.OkButtons[1] = loadImage(2, 1)

        newGameScreen.CancelButtons = make([]*ebiten.Image, 2)
        newGameScreen.CancelButtons[0] = loadImage(3, 0)
        newGameScreen.CancelButtons[1] = loadImage(3, 1)

        newGameScreen.DifficultyBlock = loadImage(4, 0)
        newGameScreen.OpponentsBlock = loadImage(5, 0)
        newGameScreen.LandSizeBlock = loadImage(6, 0)
        newGameScreen.MagicBlock = loadImage(7, 0)

        newGameScreen.UI = newGameScreen.MakeUI()
    })

    return outError
}
*/

func (newGameScreen *NewGameScreen) MakeUI() *uilib.UI {
    fontLbx, err := newGameScreen.Cache.GetLbxFile("FONTS.LBX")
    if err != nil {
        log.Printf("Unable to open fonts: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from FONTS.LBX: %v", err)
        return nil
    }

    buttonFont := font.MakeOptimizedFont(fonts[3])

    _ = buttonFont

    var elements []*uilib.UIElement

    okX := 160 + 91
    okY := 179

    okButtons, _ := newGameScreen.ImageCache.GetImages("newgame.lbx", 2)

    okIndex := 0
    elements = append(elements, &uilib.UIElement{
        Rect: util.ImageRect(okX, okY, okButtons[0]),
        LeftClick: func(element *uilib.UIElement) {
            okIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement) {
            okIndex = 0
            newGameScreen.State = NewGameStateOk
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(okX), float64(okY))
            screen.DrawImage(okButtons[okIndex], &options)
        },
    })

    /*
    cancelX := 160 + 10
    cancelY := 179

    elements = append(elements, &uilib.UIElement{
        Rect: image.Rect(cancelX, cancelY, cancelX + newGameScreen.CancelButtons[0].Bounds().Dx(), cancelY + newGameScreen.CancelButtons[0].Bounds().Dy()),
        LeftClick: func(element *uilib.UIElement) {
            newGameScreen.State = NewGameStateCancel
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(cancelX), float64(cancelY))
            screen.DrawImage(newGameScreen.CancelButtons[0], &options)
        },
    })

    difficultyX := 160 + 91
    difficultyY := 39

    elements = append(elements, &uilib.UIElement{
        Rect: image.Rect(difficultyX, difficultyY, difficultyX + newGameScreen.DifficultyBlock.Bounds().Dx(), difficultyY + newGameScreen.DifficultyBlock.Bounds().Dy()),
        LeftClick: func(element *uilib.UIElement) {
            newGameScreen.Settings.DifficultyNext()
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(difficultyX), float64(difficultyY))
            screen.DrawImage(newGameScreen.DifficultyBlock, &options)

            x := difficultyX + newGameScreen.DifficultyBlock.Bounds().Dx() / 2
            y := difficultyY + 3
            newGameScreen.Font.PrintCenter(screen, float64(x), float64(y), 1, ebiten.ColorScale{}, newGameScreen.Settings.DifficultyString())
        },
    })

    opponentsX := 160 + 91
    opponentsY := 66

    elements = append(elements, &uilib.UIElement{
        Rect: image.Rect(opponentsX, opponentsY, opponentsX + newGameScreen.OpponentsBlock.Bounds().Dx(), opponentsY + newGameScreen.OpponentsBlock.Bounds().Dy()),
        LeftClick: func(element *uilib.UIElement) {
            newGameScreen.Settings.OpponentsNext()
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(opponentsX), float64(opponentsY))
            screen.DrawImage(newGameScreen.OpponentsBlock, &options)
            x := opponentsX + newGameScreen.OpponentsBlock.Bounds().Dx() / 2
            y := opponentsY + 4
            newGameScreen.Font.PrintCenter(screen, float64(x), float64(y), 1, ebiten.ColorScale{}, newGameScreen.Settings.OpponentsString())
        },
    })

    landsizeX := 160 + 91
    landsizeY := 93

    elements = append(elements, &uilib.UIElement{
        Rect: image.Rect(landsizeX, landsizeY, landsizeX + newGameScreen.LandSizeBlock.Bounds().Dx(), landsizeY + newGameScreen.LandSizeBlock.Bounds().Dy()),
        LeftClick: func(element *uilib.UIElement) {
            newGameScreen.Settings.LandSizeNext()
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(landsizeX), float64(landsizeY))
            screen.DrawImage(newGameScreen.LandSizeBlock, &options)

            x := landsizeX + newGameScreen.LandSizeBlock.Bounds().Dx() / 2
            y := landsizeY + 4

            newGameScreen.Font.PrintCenter(screen, float64(x), float64(y), 1, ebiten.ColorScale{}, newGameScreen.Settings.LandSizeString())
        },
    })

    magicX := 160 + 91
    magicY := 120

    elements = append(elements, &uilib.UIElement{
        Rect: image.Rect(magicX, magicY, magicX + newGameScreen.MagicBlock.Bounds().Dx(), magicY + newGameScreen.MagicBlock.Bounds().Dy()),
        LeftClick: func(element *uilib.UIElement) {
            newGameScreen.Settings.MagicNext()
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(magicX), float64(magicY))
            screen.DrawImage(newGameScreen.MagicBlock, &options)
            x := magicX + newGameScreen.MagicBlock.Bounds().Dx() / 2
            y := magicY + 4
            newGameScreen.Font.PrintCenter(screen, float64(x), float64(y), 1, ebiten.ColorScale{}, newGameScreen.Settings.MagicString())
        },
    })
    */

    ui := uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions

            background, _ := newGameScreen.ImageCache.GetImage("newgame.lbx", 0, 0)
            screen.DrawImage(background, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(160 + 5, 0)

            optionsImage, _ := newGameScreen.ImageCache.GetImage("newgame.lbx", 1, 0)
            screen.DrawImage(optionsImage, &options)

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                element.Draw(element, screen)
            })
        },
    }

    ui.SetElementsFromArray(elements)

    return &ui
}

func (newGameScreen *NewGameScreen) Update() NewGameState {

    if newGameScreen.UI != nil {
        newGameScreen.UI.StandardUpdate()
    }

    return newGameScreen.State
}

func (newGameScreen *NewGameScreen) Draw(screen *ebiten.Image) {
    newGameScreen.UI.Draw(newGameScreen.UI, screen)
}

func MakeNewGameScreen(cache *lbx.LbxCache) *NewGameScreen {
    out := &NewGameScreen{
        State: NewGameStateRunning,
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
        Settings: NewGameSettings{
            Difficulty: 0,
            Opponents: 3,
            LandSize: 1,
            Magic: 1,
        },
    }
    out.UI = out.MakeUI()
    return out
}
