package main

import (
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    // "github.com/kazzmir/master-of-magic/game/magic/data"

    "github.com/hajimehoshi/ebiten/v2"
)

type Engine struct {
}

func (engine *Engine) Update() error {
    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (int, int) {
    return outsideWidth, outsideHeight
}

func MakeEngine(cache *lbx.LbxCache) *Engine {
    return &Engine{}
}

func main() {
    cache := lbx.AutoCache()

    engine := MakeEngine(cache)

    err := ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
