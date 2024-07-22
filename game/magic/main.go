package main

import (
    "log"
    "fmt"
    "image/color"
    "sync"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ScreenWidth = 320
const ScreenHeight = 200

func stretchImage(screen *ebiten.Image, sprite *ebiten.Image){
    var options ebiten.DrawImageOptions
    options.GeoM.Scale(float64(ScreenWidth) / float64(sprite.Bounds().Dx()), float64(ScreenHeight) / float64(sprite.Bounds().Dy()))
    screen.DrawImage(sprite, &options)
}

type NewGameSettings struct {
    Difficulty int
    Opponents int
    LandSize int
    Magic int
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
}

func (newGameScreen *NewGameScreen) Load(cache *lbx.LbxCache) error {
    var outError error = nil

    newGameScreen.loaded.Do(func() {
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

        newGameScreen.Font = font.MakeOptimizedFont(fonts[3])

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
        options.GeoM.Translate(160 + 91, 179)
        screen.DrawImage(newGameScreen.OkButtons[0], &options)
    }

    if newGameScreen.CancelButtons[0] != nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(160 + 10, 179)
        screen.DrawImage(newGameScreen.CancelButtons[0], &options)
    }

    if newGameScreen.DifficultyBlock != nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(160 + 91, 39)
        screen.DrawImage(newGameScreen.DifficultyBlock, &options)
    }

    if newGameScreen.OpponentsBlock != nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(160 + 91, 66)
        screen.DrawImage(newGameScreen.OpponentsBlock, &options)
    }

    if newGameScreen.LandSizeBlock != nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(160 + 91, 93)
        screen.DrawImage(newGameScreen.LandSizeBlock, &options)
    }

    if newGameScreen.MagicBlock != nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(160 + 91, 120)
        screen.DrawImage(newGameScreen.MagicBlock, &options)
    }

    if newGameScreen.Font != nil {
        newGameScreen.Font.PrintCenter(screen, 160 + 91 + float64(newGameScreen.DifficultyBlock.Bounds().Dx()) / 2, 39 + 3, 1, newGameScreen.Settings.DifficultyString())
        newGameScreen.Font.PrintCenter(screen, 160 + 91 + float64(newGameScreen.OpponentsBlock.Bounds().Dx()) / 2, 66 + 4, 1, newGameScreen.Settings.OpponentsString())
        newGameScreen.Font.PrintCenter(screen, 160 + 91 + float64(newGameScreen.LandSizeBlock.Bounds().Dx()) / 2, 93 + 4, 1, newGameScreen.Settings.LandSizeString())
        newGameScreen.Font.PrintCenter(screen, 160 + 91 + float64(newGameScreen.LandSizeBlock.Bounds().Dx()) / 2, 120 + 4, 1, newGameScreen.Settings.MagicString())
    }
}

type MagicGame struct {
    LbxCache *lbx.LbxCache

    NewGameScreen NewGameScreen
}

func NewMagicGame() *MagicGame {
    return &MagicGame{
        LbxCache: lbx.MakeLbxCache(),
        NewGameScreen: NewGameScreen{
            Settings: NewGameSettings{
                Difficulty: 0,
                Opponents: 3,
                LandSize: 1,
                Magic: 1,
            },
        },
    }
}

func (game *MagicGame) Update() error {
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    err := game.NewGameScreen.Load(game.LbxCache)
    if err != nil {
        return err
    }

    return nil
}

func (game *MagicGame) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return ScreenWidth, ScreenHeight
}

func (game *MagicGame) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})

    game.NewGameScreen.Draw(screen)
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(ScreenWidth * 5, ScreenHeight * 5)
    ebiten.SetWindowTitle("magic")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    game := NewMagicGame()

    err := ebiten.RunGame(game)
    if err != nil {
        log.Printf("Error: %v", err)
    }

}
