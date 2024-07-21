package main

import (
    "log"
    "image/color"
    "sync"

    "github.com/kazzmir/master-of-magic/lib/lbx"

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

type NewGameScreen struct {
    LbxFile *lbx.LbxFile
    Background *ebiten.Image
    Options *ebiten.Image
    loaded sync.Once
}

func (newGameScreen *NewGameScreen) Load(cache *lbx.LbxCache) error {
    var outError error = nil

    newGameScreen.loaded.Do(func() {
        newGameLbx, err := cache.GetLbxFile("magic-data/NEWGAME.LBX")
        if err != nil {
            log.Printf("Unable to load NEWGAME.LBX: %v", err)
            outError = err
            return
        }

        background, err := newGameLbx.ReadImages(0)
        if err != nil {
            log.Printf("Unable to read background image from NEWGAME.LBX: %v", err)
            outError = err
            return
        }

        newGameScreen.Background = ebiten.NewImageFromImage(background[0])

        options, err := newGameLbx.ReadImages(1)
        if err != nil {
            log.Printf("Unable to read options image from NEWGAME.LBX: %v", err)
            outError = err
            return
        }

        newGameScreen.Options = ebiten.NewImageFromImage(options[0])
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
}

type MagicGame struct {
    LbxCache *lbx.LbxCache

    NewGameScreen NewGameScreen
}

func NewMagicGame() *MagicGame {
    return &MagicGame{
        LbxCache: lbx.MakeLbxCache(),
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

    game.NewGameScreen.Load(game.LbxCache)

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
