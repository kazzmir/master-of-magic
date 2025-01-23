package main

import (
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/data"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    NewWizardScreen *setup.NewWizardScreen
}

func NewEngine() (*Engine, error) {
    cache := lbx.AutoCache()

    screen := setup.MakeNewWizardScreen(cache)

    // screen.Activate()

    return &Engine{
        LbxCache: cache,
        NewWizardScreen: screen,
    }, nil
}

func (engine *Engine) Update() error {

    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    /*
    for _, key := range keys {
        if key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }
    */

    switch engine.NewWizardScreen.Update() {
        case setup.NewWizardScreenStateFinished:
            wizard := engine.NewWizardScreen.CustomWizard
            log.Printf("New wizard: %+v", wizard)
            return ebiten.Termination
        case setup.NewWizardScreenStateCanceled:
            return ebiten.Termination
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    engine.NewWizardScreen.Draw(screen)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return data.ScreenWidth, data.ScreenHeight
}

func main(){

    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(data.ScreenWidth * 2, data.ScreenHeight * 2)
    ebiten.SetWindowTitle("new wizard")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    engine, err := NewEngine()

    if err != nil {
        log.Printf("Error: unable to load engine: %v", err)
        return
    }

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }

}
