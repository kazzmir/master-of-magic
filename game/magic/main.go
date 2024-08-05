package main

import (
    "log"
    _ "fmt"
    "image"
    "image/color"

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

func pointInRect(x int, y int, rect image.Rectangle) bool {
    return x >= rect.Min.X && x < rect.Max.X && y >= rect.Min.Y && y < rect.Max.Y
}

type Game struct {
    active bool
}

func MakeGame(wizard wizardCustom) *Game {
    return &Game{}
}

func (game *Game) IsActive() bool {
    return game.active
}

func (game *Game) Activate() {
    game.active = true
}

func (game *Game) Update(){
}

func (game *Game) Draw(screen *ebiten.Image){
}

type MagicGame struct {
    LbxCache *lbx.LbxCache

    NewGameScreen *NewGameScreen
    NewWizardScreen *NewWizardScreen

    Game *Game
}

func NewMagicGame() (*MagicGame, error) {
    game := &MagicGame{
        LbxCache: lbx.MakeLbxCache("magic-data"),
        NewGameScreen: MakeNewGameScreen(),
        NewWizardScreen: MakeNewWizardScreen(),
    }

    /*
    err := game.NewGameScreen.Load(game.LbxCache)
    if err != nil {
        return nil, err
    }
    game.NewGameScreen.Activate()
    */

    err := game.NewWizardScreen.Load(game.LbxCache)
    if err != nil {
        return nil, err
    }
    game.NewWizardScreen.Activate()

    return game, err
}

func (game *MagicGame) Update() error {
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    if game.NewGameScreen != nil && game.NewGameScreen.IsActive() {
        switch game.NewGameScreen.Update() {
            case NewGameStateRunning:
            case NewGameStateOk:
                game.NewGameScreen.Deactivate()
                err := game.NewWizardScreen.Load(game.LbxCache)
                if err != nil {
                    return err
                }
                game.NewWizardScreen.Activate()
            case NewGameStateCancel:
                return ebiten.Termination
        }
    }

    if game.NewWizardScreen != nil && game.NewWizardScreen.IsActive() {
        switch game.NewWizardScreen.Update() {
            case NewWizardScreenStateFinished:
                game.NewWizardScreen.Deactivate()
                wizard := game.NewWizardScreen.CustomWizard
                log.Printf("Launch game with wizard: %+v\n", wizard)
                game.Game = MakeGame(wizard)
                game.Game.Activate()
                game.NewWizardScreen = nil
        }
    }

    if game.Game != nil && game.Game.IsActive() {
        game.Game.Update()
    }

    return nil
}

func (game *MagicGame) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return ScreenWidth, ScreenHeight
}

func (game *MagicGame) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})

    if game.NewGameScreen.IsActive() {
        game.NewGameScreen.Draw(screen)
    }

    if game.NewWizardScreen != nil && game.NewWizardScreen.IsActive() {
        game.NewWizardScreen.Draw(screen)
    }

    if game.Game != nil && game.Game.IsActive() {
        game.Game.Draw(screen)
    }
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(ScreenWidth * 5, ScreenHeight * 5)
    ebiten.SetWindowTitle("magic")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    game, err := NewMagicGame()
    
    if err != nil {
        log.Printf("Error: unable to load game: %v", err)
        return
    }

    err = ebiten.RunGame(game)
    if err != nil {
        log.Printf("Error: %v", err)
    }

}
