package main

import (
    "log"
    "image/color"
    "slices"
    "cmp"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/functional"
    fontlib "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/util/common"
    "github.com/kazzmir/master-of-magic/game/magic/fonts"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/colorm"

    "github.com/ebitenui/ebitenui"
    "github.com/ebitenui/ebitenui/widget"
    ui_image "github.com/ebitenui/ebitenui/image"
)

type Engine struct {
    Cache *lbx.LbxCache
    UI *ebitenui.UI
}

func MakeEngine(cache *lbx.LbxCache) *Engine {
    engine := &Engine{
        Cache: cache,
    }

    engine.UI = engine.MakeUI()

    return engine
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

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    engine.UI.Draw(screen)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (int, int) {
    // Layout logic here
    return outsideWidth, outsideHeight
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

func lighten(c color.Color, amount float64) color.Color {
    var change colorm.ColorM
    change.ChangeHSV(0, 1 - amount/100, 1 + amount/100)
    return change.Apply(c)
}


func makeNineImage(img *ebiten.Image, border int) *ui_image.NineSlice {
    width := img.Bounds().Dx()
    return ui_image.NewNineSliceSimple(img, border, width - border * 2)
}

func makeNineRoundedButtonImage(width int, height int, border int, col color.Color) *widget.ButtonImage {
    return &widget.ButtonImage{
        Idle: makeNineImage(makeRoundedButtonImage(width, height, border, col), border),
        Hover: makeNineImage(makeRoundedButtonImage(width, height, border, lighten(col, 50)), border),
        Pressed: makeNineImage(makeRoundedButtonImage(width, height, border, lighten(col, 20)), border),
    }
}

func padding(n int) widget.Insets {
    return widget.Insets{Top: n, Bottom: n, Left: n, Right: n}
}

func makeRoundedButtonImage(width int, height int, border int, col color.Color) *ebiten.Image {
    img := ebiten.NewImage(width, height)

    vector.DrawFilledRect(img, float32(border), 0, float32(width - border * 2), float32(height), col, true)
    vector.DrawFilledRect(img, 0, float32(border), float32(width), float32(height - border * 2), col, true)
    vector.DrawFilledCircle(img, float32(border), float32(border), float32(border), col, true)
    vector.DrawFilledCircle(img, float32(width-border), float32(border), float32(border), col, true)
    vector.DrawFilledCircle(img, float32(border), float32(height-border), float32(border), col, true)
    vector.DrawFilledCircle(img, float32(width-border), float32(height-border), float32(border), col, true)

    return img
}

func (engine *Engine) MakeUI() *ebitenui.UI {
    face, _ := loadFont(19)
    backgroundImage := ui_image.NewNineSliceColor(color.NRGBA{R: 32, G: 32, B: 32, A: 128})

    rootContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
            widget.RowLayoutOpts.Spacing(12),
            widget.RowLayoutOpts.Padding(widget.Insets{Top: 10, Left: 10, Right: 10}),
        )),
        widget.ContainerOpts.BackgroundImage(backgroundImage),
        // widget.ContainerOpts.BackgroundImage(backgroundImageNine),
    )

    fakeImage := ui_image.NewNineSliceColor(color.NRGBA{R: 32, G: 32, B: 32, A: 255})

    textArea := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(4),
            widget.RowLayoutOpts.Padding(padding(5)),
        )),
        widget.ContainerOpts.BackgroundImage(makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 32, G: 32, B: 32, A: 255}), 5)),
    )

    makeFont := functional.Memoize(func (name string) *fontlib.Font {
        loadedFonts, err := fonts.LoadFonts(engine.Cache, name)
        if err != nil {
            log.Printf("Error loading font: %v", err)
            return nil
        }

        return loadedFonts[name]
    })

    updateTextFont := func (name string) {
        textArea.RemoveChildren()

        graphic := widget.NewGraphic()

        surface := ebiten.NewImage(700, 200)
        font := makeFont(name)
        if font != nil {
            scale := 3.0
            font.PrintWrap(surface, 1, 1, float64(surface.Bounds().Dx() - 2) / scale, fontlib.FontOptions{Scale: scale}, "This is sample text. I am proud of it")
            graphic.Image = surface
        }

        textArea.AddChild(graphic)
    }

    fontList := widget.NewList(
        widget.ListOpts.EntryFontFace(face),

        widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(
            widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                MaxHeight: 850,
            }),
            widget.WidgetOpts.MinSize(0, 850),
        )),

        widget.ListOpts.SliderOpts(
            widget.SliderOpts.Images(&widget.SliderTrackImage{
                    Idle: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                    Hover: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                },
                makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xad, G: 0x8d, B: 0x55, A: 0xff}),
            ),
        ),

        widget.ListOpts.HideHorizontalSlider(),

        widget.ListOpts.EntryLabelFunc(
            func (e any) string {
                item := e.(string)
                return item
            },
        ),

        widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
            entry := args.Entry.(string)
            log.Printf("Entry Selected: %v", entry)
            updateTextFont(entry)
        }),

        widget.ListOpts.EntryColor(&widget.ListEntryColor{
            Selected: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
            Unselected: color.NRGBA{R: 0, G: 255, B: 0, A: 255},
        }),

        widget.ListOpts.ScrollContainerOpts(
            widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
                Idle: ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 64, B: 64, A: 255}),
                Disabled: fakeImage,
                Mask: fakeImage,
            }),
        ),
    )

    for _, name := range slices.SortedFunc(slices.Values(fonts.GetFontList()), cmp.Compare) {
        fontList.AddEntry(name)
    }

    rootContainer.AddChild(fontList)
    rootContainer.AddChild(textArea)

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
