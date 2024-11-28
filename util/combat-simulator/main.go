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
    // "github.com/ebitenui/ebitenui/input"
    "github.com/ebitenui/ebitenui/widget"
    ui_image "github.com/ebitenui/ebitenui/image"
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
}

func MakeEngine(cache *lbx.LbxCache) *Engine {
    engine := Engine{
        Cache: cache,
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

    switch engine.Mode {
        case EngineModeMenu:
            engine.UI.Update()
        case EngineModeCombat:
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    switch engine.Mode {
        case EngineModeMenu:
            engine.UI.Draw(screen)
        case EngineModeCombat:
    }
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

func (engine *Engine) EnterCombat() {
    engine.Mode = EngineModeCombat
}

func (engine *Engine) MakeUI() *ebitenui.UI {
    face, _ := loadFont(18)

    rootContainer := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))))

    var label1 *widget.Text

    label1 = widget.NewText(
        widget.TextOpts.Text("Hello!", face, color.White),
        widget.TextOpts.WidgetOpts(
            widget.WidgetOpts.CursorEnterHandler(func(args *widget.WidgetCursorEnterEventArgs) {
                label1.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
            }),
            widget.WidgetOpts.CursorExitHandler(func(args *widget.WidgetCursorExitEventArgs) {
                label1.Color = color.White
            }),
        ),
    )

    rootContainer.AddChild(label1)

    var label2 *widget.Text
    label2 = widget.NewText(
        widget.TextOpts.Text("Everyone!", face, color.White),
        widget.TextOpts.WidgetOpts(
            widget.WidgetOpts.CursorEnterHandler(func(args *widget.WidgetCursorEnterEventArgs) {
                label2.Color = color.RGBA{R: 0, G: 255, B: 0, A: 255}
            }),
            widget.WidgetOpts.CursorExitHandler(func(args *widget.WidgetCursorExitEventArgs) {
                label2.Color = color.White
            }),
        ),
    )
    rootContainer.AddChild(label2)

    fakeImage := ui_image.NewNineSliceColor(color.NRGBA{R: 255, G: 0, B: 0, A: 255})
    buttonImage := ui_image.NewNineSliceColor(color.NRGBA{R: 150, G: 150, B: 0, A: 255})

    unitList1 := widget.NewListComboButton(
        widget.ListComboButtonOpts.SelectComboButtonOpts(
            widget.SelectComboButtonOpts.ComboButtonOpts(
                widget.ComboButtonOpts.MaxContentHeight(200),
                widget.ComboButtonOpts.ButtonOpts(
                    widget.ButtonOpts.Image(&widget.ButtonImage{
                        Idle: buttonImage,
                        Hover: buttonImage,
                        Pressed: buttonImage,
                    }),
                    widget.ButtonOpts.Text("Select Unit", face, &widget.ButtonTextColor{
                        Idle: color.White,
                        Disabled: color.White,
                    }),
                ),
            ),
        ),
        widget.ListComboButtonOpts.EntryLabelFunc(
            func (e any) string {
                return "Button " + e.(string)
            },
            func (e any) string {
                return "List " + e.(string)
            },
        ),
        widget.ListComboButtonOpts.ListOpts(
            widget.ListOpts.EntryFontFace(face),
            widget.ListOpts.Entries([]any{"x", "y", "z", "a", "b"}),
            widget.ListOpts.EntryColor(&widget.ListEntryColor{
                Selected: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
                Unselected: color.NRGBA{R: 0, G: 255, B: 0, A: 255},
            }),
            widget.ListOpts.SliderOpts(
                widget.SliderOpts.Images(&widget.SliderTrackImage{
                    Idle: fakeImage,
                    Hover: fakeImage,
                }, &widget.ButtonImage{
                    Idle: fakeImage,
                    Hover: fakeImage,
                    Pressed: fakeImage,
                }),
            ),
            widget.ListOpts.ScrollContainerOpts(
                widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
                    Idle: fakeImage,
                    Disabled: fakeImage,
                    Mask: fakeImage,
                }),
            ),
        ),
    )
    rootContainer.AddChild(unitList1)

    rootContainer.AddChild(widget.NewButton(
        widget.ButtonOpts.Image(&widget.ButtonImage{
            Idle: buttonImage,
            Hover: buttonImage,
            Pressed: buttonImage,
        }),
        widget.ButtonOpts.Text("Enter Combat!", face, &widget.ButtonTextColor{
            Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
            Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            Pressed: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
        }),
        widget.ButtonOpts.PressedHandler(func (args *widget.ButtonPressedEventArgs) {
            engine.EnterCombat()
        }),
    ))

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
