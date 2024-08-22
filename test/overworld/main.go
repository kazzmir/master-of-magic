package main

import (
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/data"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    Game *gamelib.Game
}

func NewEngine() (*Engine, error) {
    cache := lbx.AutoCache()

    wizard := setup.WizardCustom{
        Banner: data.BannerBlue,
    }

    game := gamelib.MakeGame(wizard, cache)
    err := game.Load(cache)

    if err != nil {
        return nil, err
    }

    game.Activate()

    player := game.AddPlayer(wizard)

    player.AddCity(citylib.City{
        Population: 6000,
        Wall: false,
        X: 4,
        Y: 5,
    })

    player.LiftFog(4, 5, 3)

    drake := player.AddUnit(gamelib.Unit{
        Unit: units.GreatDrake,
        Banner: wizard.Banner,
        X: 5,
        Y: 5,
    })

    player.SetSelectedUnit(drake)

    player.LiftFog(5, 5, 2)

    return &Engine{
        LbxCache: cache,
        Game: game,
    }, nil
}

func (engine *Engine) Update() error {

    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    switch engine.Game.Update() {
        case gamelib.GameStateRunning:
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    engine.Game.Draw(screen)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return data.ScreenWidth, data.ScreenHeight
}

func main(){

    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(data.ScreenWidth * 5, data.ScreenHeight * 5)
    ebiten.SetWindowTitle("new screen")
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
