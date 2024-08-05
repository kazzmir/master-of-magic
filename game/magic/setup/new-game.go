package setup

import (
    "fmt"
    _ "log"
    "image"
    "sync"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const DifficultyMax = 4
const OpponentsMax = 4
const LandSizeMax = 2
const MagicMax = 2

func pointInRect(x int, y int, rect image.Rectangle) bool {
    return x >= rect.Min.X && x < rect.Max.X && y >= rect.Min.Y && y < rect.Max.Y
}

type NewGameSettings struct {
    Difficulty int
    Opponents int
    LandSize int
    Magic int
}

func (settings *NewGameSettings) DifficultyNext() {
    settings.Difficulty += 1
    if settings.Difficulty > DifficultyMax {
        settings.Difficulty = 0
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
    settings.Magic += 1
    if settings.Magic > MagicMax {
        settings.Magic = 0
    }
}

func (settings *NewGameSettings) DifficultyString() string {
    kinds := []string{"Intro", "Easy", "Normal", "Hard", "Impossible"}
    return kinds[settings.Difficulty]
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
    kinds := []string{"Weak", "Normal", "Powerful"}
    return kinds[settings.Magic]
}

type NewGameUI int

const (
    NewGameDifficulty NewGameUI = iota
    NewGameOppoonents
    NewGameLandSize
    NewGameMagic
    NewGameOk
    NewGameCancel
)

type NewGameState int
const (
    NewGameStateRunning NewGameState = iota
    NewGameStateOk
    NewGameStateCancel
)

type NewGameScreen struct {
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

    Settings NewGameSettings
    Active bool
}

func (newGameScreen *NewGameScreen) Activate() {
    newGameScreen.Active = true
}

func (newGameScreen *NewGameScreen) Deactivate() {
    newGameScreen.Active = false
}

func (newGameScreen *NewGameScreen) IsActive() bool {
    return newGameScreen.Active
}

func (newGameScreen *NewGameScreen) Load(cache *lbx.LbxCache) error {
    var outError error = nil

    newGameScreen.loaded.Do(func() {
        fontLbx, err := cache.GetLbxFile("FONTS.LBX")
        if err != nil {
            outError = fmt.Errorf("Unable to read FONTS.LBX: %v", err)
            return
        }

        fonts, err := fontLbx.ReadFonts(0)
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
    })

    return outError
}

func (newGameScreen *NewGameScreen) GetUIRect(component NewGameUI) image.Rectangle {
    switch component {
        case NewGameDifficulty:
            bounds := newGameScreen.DifficultyBlock.Bounds()
            x := 160 + 91
            y := 39
            return image.Rect(x, y, x + bounds.Dx(), y + bounds.Dy())
        case NewGameOppoonents:
            bounds := newGameScreen.OpponentsBlock.Bounds()
            x := 160 + 91
            y := 66
            return image.Rect(x, y, x + bounds.Dx(), y + bounds.Dy())
        case NewGameLandSize:
            bounds := newGameScreen.LandSizeBlock.Bounds()
            x := 160 + 91
            y := 93
            return image.Rect(x, y, x + bounds.Dx(), y + bounds.Dy())
        case NewGameMagic:
            bounds := newGameScreen.MagicBlock.Bounds()
            x := 160 + 91
            y := 120
            return image.Rect(x, y, x + bounds.Dx(), y + bounds.Dy())
        case NewGameOk:
            bounds := newGameScreen.OkButtons[0].Bounds()
            x := 160 + 91
            y := 179
            return image.Rect(x, y, x + bounds.Dx(), y + bounds.Dy())
        case NewGameCancel:
            bounds := newGameScreen.CancelButtons[0].Bounds()
            x := 160 + 10
            y := 179
            return image.Rect(x, y, x + bounds.Dx(), y + bounds.Dy())
    }

    return image.Rectangle{}
}

func (newGameScreen *NewGameScreen) Update() NewGameState {
    if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
        x, y := ebiten.CursorPosition()
        if pointInRect(x, y, newGameScreen.GetUIRect(NewGameDifficulty)) {
            newGameScreen.Settings.DifficultyNext()
        }

        if pointInRect(x, y, newGameScreen.GetUIRect(NewGameOppoonents)) {
            newGameScreen.Settings.OpponentsNext()
        }

        if pointInRect(x, y, newGameScreen.GetUIRect(NewGameLandSize)) {
            newGameScreen.Settings.LandSizeNext()
        }

        if pointInRect(x, y, newGameScreen.GetUIRect(NewGameMagic)) {
            newGameScreen.Settings.MagicNext()
        }

        if pointInRect(x, y, newGameScreen.GetUIRect(NewGameOk)) {
            return NewGameStateOk
        }

        if pointInRect(x, y, newGameScreen.GetUIRect(NewGameCancel)) {
            return NewGameStateCancel
        }
    }

    return NewGameStateRunning
}

func (newGameScreen *NewGameScreen) Draw(screen *ebiten.Image) {
    if newGameScreen.Background != nil {
        var options ebiten.DrawImageOptions
        screen.DrawImage(newGameScreen.Background, &options)
    }

    if newGameScreen.Options != nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(160 + 5, 0)
        screen.DrawImage(newGameScreen.Options, &options)
    }

    if newGameScreen.OkButtons[0] != nil {
        var options ebiten.DrawImageOptions
        rect := newGameScreen.GetUIRect(NewGameOk)
        options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
        screen.DrawImage(newGameScreen.OkButtons[0], &options)
    }

    if newGameScreen.CancelButtons[0] != nil {
        var options ebiten.DrawImageOptions
        rect := newGameScreen.GetUIRect(NewGameCancel)
        options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
        screen.DrawImage(newGameScreen.CancelButtons[0], &options)
    }

    if newGameScreen.DifficultyBlock != nil {
        var options ebiten.DrawImageOptions
        rect := newGameScreen.GetUIRect(NewGameDifficulty)
        options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
        screen.DrawImage(newGameScreen.DifficultyBlock, &options)
    }

    if newGameScreen.OpponentsBlock != nil {
        var options ebiten.DrawImageOptions
        rect := newGameScreen.GetUIRect(NewGameOppoonents)
        options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
        screen.DrawImage(newGameScreen.OpponentsBlock, &options)
    }

    if newGameScreen.LandSizeBlock != nil {
        var options ebiten.DrawImageOptions
        rect := newGameScreen.GetUIRect(NewGameLandSize)
        options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
        screen.DrawImage(newGameScreen.LandSizeBlock, &options)
    }

    if newGameScreen.MagicBlock != nil {
        var options ebiten.DrawImageOptions
        rect := newGameScreen.GetUIRect(NewGameMagic)
        options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
        screen.DrawImage(newGameScreen.MagicBlock, &options)
    }

    if newGameScreen.Font != nil {
        newGameScreen.Font.PrintCenter(screen, 160 + 91 + float64(newGameScreen.DifficultyBlock.Bounds().Dx()) / 2, 39 + 3, 1, newGameScreen.Settings.DifficultyString())
        newGameScreen.Font.PrintCenter(screen, 160 + 91 + float64(newGameScreen.OpponentsBlock.Bounds().Dx()) / 2, 66 + 4, 1, newGameScreen.Settings.OpponentsString())
        newGameScreen.Font.PrintCenter(screen, 160 + 91 + float64(newGameScreen.LandSizeBlock.Bounds().Dx()) / 2, 93 + 4, 1, newGameScreen.Settings.LandSizeString())
        newGameScreen.Font.PrintCenter(screen, 160 + 91 + float64(newGameScreen.LandSizeBlock.Bounds().Dx()) / 2, 120 + 4, 1, newGameScreen.Settings.MagicString())
    }
}

func MakeNewGameScreen() *NewGameScreen {
    return &NewGameScreen{
        Active: false,
        Settings: NewGameSettings{
            Difficulty: 0,
            Opponents: 3,
            LandSize: 1,
            Magic: 1,
        },
    }
}
