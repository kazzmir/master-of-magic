package main

import (
    "log"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/summon"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    Summon *summon.SummonUnit
}

func NewEngine() (*Engine, error) {
    cache := lbx.AutoCache()

    summonUnit := summon.MakeSummonUnit(cache, units.MagicSpirit, data.WizardMerlin)

    return &Engine{
        LbxCache: cache,
        Summon: summonUnit,
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

    switch engine.Summon.Update() {
        case summon.SummonStateRunning:
        case summon.SummonStateDone:
            return ebiten.Termination
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{R: 0, G: 150, B: 150, A: 0xff})

    engine.Summon.Draw(screen)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return data.ScreenWidth, data.ScreenHeight
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    monitorWidth, _ := ebiten.Monitor().Size()
    size := monitorWidth / 390

    ebiten.SetWindowSize(data.ScreenWidth * size, data.ScreenHeight * size)
    ebiten.SetWindowTitle("summon unit")
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
