package main

import (
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/ebitenui/ebitenui"
    "github.com/ebitenui/ebitenui/widget"
)

type Engine struct {
    Cache *lbx.LbxCache
}

func MakeEngine(cache *lbx.LbxCache) *Engine {
    return &Engine{
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

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (int, int) {
    // Layout logic here
    return outsideWidth, outsideHeight
}

func (engine *Engine) MakeUI() *ebitenui.UI {
    rootContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(12),
            widget.RowLayoutOpts.Padding(widget.Insets{Top: 10, Left: 10, Right: 10}),
        )),
        // widget.ContainerOpts.BackgroundImage(backgroundImage),
        // widget.ContainerOpts.BackgroundImage(backgroundImageNine),
    )

    ui := ebitenui.UI{
        Container: rootContainer,
    }

    return &ui
}

func main(){
    cache := lbx.AutoCache()

    engine := MakeEngine(cache)
    ebiten.SetWindowSize(1200, 900)
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    err := ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
