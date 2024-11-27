package main

import (
    "log"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const EngineWidth = 800
const EngineHeight = 600

type EngineMode int
const (
    EngineModeMenu EngineMode = iota
    EngineModeCombat
)

type Engine struct {
    Cache *lbx.LbxCache
    Mode EngineMode
}

func MakeEngine(cache *lbx.LbxCache) Engine {
    return Engine{
        Cache: cache,
    }
}

func (engine *Engine) Update() error {
    keys := inpututil.AppendJustPressedKeys(nil)

    for _, key := range keys {
        switch key {
            case ebiten.KeyEscape, ebiten.KeyCapsLock:
                return ebiten.Termination
        }
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    switch engine.Mode {
        case EngineModeMenu:
            return EngineWidth, EngineHeight
        case EngineModeCombat:
            return data.ScreenWidth, data.ScreenHeight
    }

    return 0, 0
}

func main(){
    cache := lbx.AutoCache()

    engine := MakeEngine(cache)
    ebiten.SetWindowSize(800, 600)

    err := ebiten.RunGame(&engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
