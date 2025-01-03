package main

import (
    "log"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/util/common"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/text/v2"

    "github.com/ebitenui/ebitenui"
    "github.com/ebitenui/ebitenui/widget"
    ui_image "github.com/ebitenui/ebitenui/image"
)

type Engine struct {
    cache *lbx.LbxCache
    Artifacts []artifact.Artifact
    UI *ebitenui.UI
}

func MakeEngine(cache *lbx.LbxCache) (*Engine, error) {
    artifacts, err := artifact.ReadArtifacts(cache)

    if err != nil {
        return nil, err
    }

    log.Printf("Loaded %d artifacts", len(artifacts))

    engine := &Engine{
        cache: cache,
        Artifacts: artifacts,
    }

    engine.UI = engine.MakeUI()

    return engine, nil
}

func padding(n int) widget.Insets {
    return widget.Insets{Top: n, Bottom: n, Left: n, Right: n}
}

func loadFont(size float64) (text.Face, error) {
    source, err := common.LoadFont()

    if err != nil {
        log.Fatal(err)
        return nil, err
    }

    return &text.GoTextFace{
        Source: source,
        Size:   size,
    }, nil
}

func (engine *Engine) MakeUI() *ebitenui.UI {
    // face, _ := loadFont(19)
    backgroundImage := ui_image.NewNineSliceColor(color.NRGBA{R: 32, G: 32, B: 32, A: 128})

    rootContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(12),
            widget.RowLayoutOpts.Padding(padding(5)),
        )),
        widget.ContainerOpts.BackgroundImage(backgroundImage),
        // widget.ContainerOpts.BackgroundImage(backgroundImageNine),
    )

    ui := ebitenui.UI{
        Container: rootContainer,
    }

    return &ui
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
