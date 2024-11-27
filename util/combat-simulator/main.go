package main

import (
    "log"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/util/common"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/text/v2"

    "github.com/ebitenui/ebitenui"
    "github.com/ebitenui/ebitenui/input"
    "github.com/ebitenui/ebitenui/widget"
)

const EngineWidth = 800
const EngineHeight = 600

type EngineMode int
const (
    EngineModeMenu EngineMode = iota
    EngineModeCombat
)

type HoverData struct {
    OnHover func()
    OnUnhover func()
}

type Engine struct {
    Cache *lbx.LbxCache
    Mode EngineMode
    UI *ebitenui.UI

    Hovers map[*widget.Widget]HoverData
    StopHovers []func()
}

func MakeEngine(cache *lbx.LbxCache) *Engine {
    engine := Engine{
        Cache: cache,
        Hovers: make(map[*widget.Widget]HoverData),
    }

    engine.UI = engine.MakeUI()

    return &engine
}

func (engine *Engine) Update() error {
    keys := inpututil.AppendJustPressedKeys(nil)

    for _, key := range keys {
        switch key {
            case ebiten.KeyEscape, ebiten.KeyCapsLock:
                return ebiten.Termination
        }
    }

    engine.UI.Update()

    for _, stopHover := range engine.StopHovers {
        stopHover()
    }

    engine.StopHovers = nil

    if input.UIHovered {
        x, y := ebiten.CursorPosition()
        find := engine.UI.Container.WidgetAt(x, y)
        if find != nil {
            useWidget := find.GetWidget()
            if hoverData, ok := engine.Hovers[useWidget]; ok {
                hoverData.OnHover()
                engine.StopHovers = append(engine.StopHovers, hoverData.OnUnhover)
            }
        }
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    engine.UI.Draw(screen)
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

func (engine *Engine) MakeUI() *ebitenui.UI {
    face, _ := loadFont(18)

    rootContainer := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))))

    label1 := widget.NewText(
        widget.TextOpts.Text("Hello!", face, color.White),
        widget.TextOpts.WidgetOpts(widget.WidgetOpts.TrackHover(true)),
    )

    engine.Hovers[label1.GetWidget()] = HoverData{
        OnHover: func(){
            label1.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
        },
        OnUnhover: func(){
            label1.Color = color.White
        },
    }

    rootContainer.AddChild(label1)

    label2 := widget.NewText(
        widget.TextOpts.Text("Everyone!", face, color.White),
        widget.TextOpts.WidgetOpts(widget.WidgetOpts.TrackHover(true)),
    )
    rootContainer.AddChild(label2)

    engine.Hovers[label2.GetWidget()] = HoverData{
        OnHover: func(){
            label2.Color = color.RGBA{R: 0, G: 255, B: 0, A: 255}
        },
        OnUnhover: func(){
            label2.Color = color.White
        },
    }

    ui := ebitenui.UI{
        Container: rootContainer,
    }

    return &ui
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

func main(){
    cache := lbx.AutoCache()

    engine := MakeEngine(cache)
    ebiten.SetWindowSize(800, 600)

    err := ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
