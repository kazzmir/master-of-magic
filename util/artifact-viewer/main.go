package main

import (
    "log"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    cache *lbx.LbxCache
    Artifacts []artifact.Artifact
}

func MakeEngine(cache *lbx.LbxCache) (*Engine, error) {
    artifacts, err := artifact.ReadArtifacts(cache)

    if err != nil {
        return nil, err
    }

    log.Printf("Loaded %d artifacts", len(artifacts))

    return &Engine{
        cache: cache,
        Artifacts: artifacts,
    }, nil
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
    return 1200, 900
}

func main(){
    cache := lbx.AutoCache()

    engine, err := MakeEngine(cache)
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    ebiten.SetWindowSize(1200, 900)
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
