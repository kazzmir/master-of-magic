package setup

import (
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
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
    Cache *lbx.LbxCache
    ImageCache util.ImageCache

    State NewGameState

    Settings NewGameSettings

    UI *uilib.UI
}

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

    var elements []*uilib.UIElement

    okX := 161 + 91
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
        Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(this.Rect.Min.X), float64(this.Rect.Min.Y))
            scale.DrawScaled(screen, okButtons[okIndex], &options)
        },
    })

    cancelX := 161 + 10
    cancelY := 179

    cancelButtons, _ := newGameScreen.ImageCache.GetImages("newgame.lbx", 3)
    cancelIndex := 0
    elements = append(elements, &uilib.UIElement{
        Rect: util.ImageRect(cancelX, cancelY, cancelButtons[0]),
        LeftClick: func(element *uilib.UIElement) {
            cancelIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement) {
            cancelIndex = 0
            newGameScreen.State = NewGameStateCancel
        },
        Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(this.Rect.Min.X), float64(this.Rect.Min.Y))
            scale.DrawScaled(screen, cancelButtons[cancelIndex], &options)
        },
    })

    difficultyX := 160 + 91
    difficultyY := 39
    difficultyBlock, _ := newGameScreen.ImageCache.GetImage("newgame.lbx", 4, 0)

    elements = append(elements, &uilib.UIElement{
        Rect: util.ImageRect(difficultyX, difficultyY, difficultyBlock),
        IsOffsetWhenPressed: true, 
        LeftClick: func(element *uilib.UIElement) {
            newGameScreen.Settings.DifficultyNext()
        },
        Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(this.Rect.Min.X), float64(this.Rect.Min.Y))
            scale.DrawScaled(screen, difficultyBlock, &options)

            x := this.Rect.Min.X + difficultyBlock.Bounds().Dx() / 2
            y := this.Rect.Min.Y + 3
            buttonFont.PrintOptions(screen, float64(x), float64(y), font.FontOptions{Scale: scale.ScaleAmount, Justify: font.FontJustifyCenter}, newGameScreen.Settings.DifficultyString())
        },
    })

    opponentsX := 160 + 91
    opponentsY := 66
    opponentsBlock, _ := newGameScreen.ImageCache.GetImage("newgame.lbx", 5, 0)

    elements = append(elements, &uilib.UIElement{
        Rect: util.ImageRect(opponentsX, opponentsY, opponentsBlock),
        IsOffsetWhenPressed: true, 
        LeftClick: func(element *uilib.UIElement) {
            newGameScreen.Settings.OpponentsNext()
        },
        Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(this.Rect.Min.X), float64(this.Rect.Min.Y))
            scale.DrawScaled(screen, opponentsBlock, &options)

            x := this.Rect.Min.X + opponentsBlock.Bounds().Dx() / 2
            y := this.Rect.Min.Y + 4
            buttonFont.PrintOptions(screen, float64(x), float64(y), font.FontOptions{Scale: scale.ScaleAmount, Justify: font.FontJustifyCenter}, newGameScreen.Settings.OpponentsString())
        },
    })

    landsizeX := 160 + 91
    landsizeY := 93
    landSizeBlock, _ := newGameScreen.ImageCache.GetImage("newgame.lbx", 6, 0)

    elements = append(elements, &uilib.UIElement{
        Rect: util.ImageRect(landsizeX, landsizeY, landSizeBlock),
        IsOffsetWhenPressed: true, 
        LeftClick: func(element *uilib.UIElement) {
            newGameScreen.Settings.LandSizeNext()
        },
        Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(this.Rect.Min.X), float64(this.Rect.Min.Y))
            scale.DrawScaled(screen, landSizeBlock, &options)

            x := this.Rect.Min.X + landSizeBlock.Bounds().Dx() / 2
            y := this.Rect.Min.Y + 4

            buttonFont.PrintOptions(screen, float64(x), float64(y), font.FontOptions{Scale: scale.ScaleAmount, Justify: font.FontJustifyCenter}, newGameScreen.Settings.LandSizeString())
        },
    })

    magicX := 160 + 91
    magicY := 120
    magicBlock, _ := newGameScreen.ImageCache.GetImage("newgame.lbx", 7, 0)

    elements = append(elements, &uilib.UIElement{
        Rect: util.ImageRect(magicX, magicY, magicBlock),
        IsOffsetWhenPressed: true, 
        LeftClick: func(element *uilib.UIElement) {
            newGameScreen.Settings.MagicNext()
        },
        Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(this.Rect.Min.X), float64(this.Rect.Min.Y))
            scale.DrawScaled(screen, magicBlock, &options)
            x := this.Rect.Min.X + magicBlock.Bounds().Dx() / 2
            y := this.Rect.Min.Y + 4
            buttonFont.PrintOptions(screen, float64(x), float64(y), font.FontOptions{Scale: scale.ScaleAmount, Justify: font.FontJustifyCenter}, newGameScreen.Settings.MagicString())
        },
    })

    ui := uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions

            background, _ := newGameScreen.ImageCache.GetImage("newgame.lbx", 0, 0)
            scale.DrawScaled(screen, background, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64((160 + 5)), 0)

            ui.StandardDraw(screen)

            // Draw buttons overlay above the buttons themselves
            optionsImage, _ := newGameScreen.ImageCache.GetImage("newgame.lbx", 1, 0)
            scale.DrawScaled(screen, optionsImage, &options)

        },
        HandleKeys: func(keys []ebiten.Key){
            for _, key := range keys {
                if inputmanager.IsQuitKey(key) {
                    newGameScreen.State = NewGameStateCancel
                }
            }
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
