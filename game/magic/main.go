package main

import (
    "log"
    _ "fmt"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

func stretchImage(screen *ebiten.Image, sprite *ebiten.Image){
    var options ebiten.DrawImageOptions
    options.GeoM.Scale(float64(data.ScreenWidth) / float64(sprite.Bounds().Dx()), float64(data.ScreenHeight) / float64(sprite.Bounds().Dy()))
    screen.DrawImage(sprite, &options)
}

type MagicGame struct {
    LbxCache *lbx.LbxCache

    NewGameScreen *setup.NewGameScreen
    NewWizardScreen *setup.NewWizardScreen

    Game *gamelib.Game
}

func NewMagicGame() (*MagicGame, error) {
    game := &MagicGame{
        LbxCache: lbx.MakeLbxCache("magic-data"),
        NewGameScreen: setup.MakeNewGameScreen(),
        NewWizardScreen: setup.MakeNewWizardScreen(),
    }

    err := game.NewGameScreen.Load(game.LbxCache)
    if err != nil {
        return nil, err
    }
    game.NewGameScreen.Activate()

    /*
    err := game.NewWizardScreen.Load(game.LbxCache)
    if err != nil {
        return nil, err
    }
    game.NewWizardScreen.Activate()
    */

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
            case setup.NewGameStateRunning:
            case setup.NewGameStateOk:
                game.NewGameScreen.Deactivate()
                err := game.NewWizardScreen.Load(game.LbxCache)
                if err != nil {
                    return err
                }
                game.NewWizardScreen.Activate()
            case setup.NewGameStateCancel:
                return ebiten.Termination
        }
    }

    if game.NewWizardScreen != nil && game.NewWizardScreen.IsActive() {
        switch game.NewWizardScreen.Update() {
            case setup.NewWizardScreenStateFinished:
                game.NewWizardScreen.Deactivate()
                wizard := game.NewWizardScreen.CustomWizard
                log.Printf("Launch game with wizard: %+v\n", wizard)
                game.Game = gamelib.MakeGame(wizard)
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
    return data.ScreenWidth, data.ScreenHeight
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

    ebiten.SetWindowSize(data.ScreenWidth * 5, data.ScreenHeight * 5)
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
