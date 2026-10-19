package ui

import (
    "image/color"
    "math"

    // "github.com/ebitenui/ebitenui"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
    "github.com/ebitenui/ebitenui/widget"
    ui_image "github.com/ebitenui/ebitenui/image"
)

func MakeButtonImage(baseImage *ui_image.NineSlice) *widget.ButtonImage {
    return &widget.ButtonImage{
        Idle: baseImage,
        Hover: baseImage,
        Pressed: baseImage,
        Disabled: baseImage,
    }
}

func SolidImage(r uint8, g uint8, b uint8) *ui_image.NineSlice {
    return ui_image.NewNineSliceColor(color.NRGBA{R: r, G: g, B: b, A: 255})
}

func HBox(opts... widget.ContainerOpt) *widget.Container {
    all := append(opts, widget.ContainerOpts.Layout(widget.NewRowLayout(
        widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
        widget.RowLayoutOpts.Spacing(4),
    )))

    return widget.NewContainer(all...)
}

func VBox(opts ...widget.ContainerOpt) *widget.Container {

    allArgs := append(opts, widget.ContainerOpts.Layout(widget.NewRowLayout(
        widget.RowLayoutOpts.Direction(widget.DirectionVertical),
        widget.RowLayoutOpts.Spacing(4),
        widget.RowLayoutOpts.Padding(&widget.Insets{Top: 2, Bottom: 2, Left: 2, Right: 2}),
    )))

    return widget.NewContainer(allArgs...)
}

// Column fills the available width. Optional row weights stretch selected rows
// into the remaining height; zero-weight rows keep their preferred height.
func Column(rowWeights ...float64) *widget.Container {
    return widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(
        widget.GridLayoutOpts.Columns(1),
        widget.GridLayoutOpts.Weights([]float64{1}, rowWeights),
        widget.GridLayoutOpts.Spacing(0, 8),
        widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(6)),
    )))
}

// ScrollRow reserves a narrow scrollbar beside flexible content.
func ScrollRow(rowWeights ...float64) *widget.Container {
    return widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(
        widget.GridLayoutOpts.Columns(2),
        widget.GridLayoutOpts.Weights([]float64{1, 0}, rowWeights),
        widget.GridLayoutOpts.Spacing(6, 0),
    )))
}

func CenteredText(textStr string, face *text.Face, textColor color.Color) *widget.Text {
    return widget.NewText(
        widget.TextOpts.Text(textStr, face, textColor),
        widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionStart),
        widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
            Stretch: true,
        })))
}

func BorderedImage(borderColor color.RGBA, borderSize int) *ui_image.NineSlice {
    return ui_image.NewBorderedNineSliceColor(color.RGBA{}, borderColor, borderSize)
}

// BattleButton is the primary preparation action. Cache the glow frames rather
// than creating GPU images during updates. A gentle, two-second pulse provides
// a hint without flashing or changing the button's size.
func BattleButton(face *text.Face, clicked func()) *widget.Button {
    const frameCount = 32
    const period = 120 // Ebiten's default 60 updates per second.
    frames := make([]*widget.ButtonImage, frameCount)
    pressed := ui_image.NewBorderedNineSliceColor(
        color.NRGBA{R: 119, G: 79, B: 29, A: 255},
        color.NRGBA{R: 255, G: 220, B: 144, A: 255}, 2,
    )
    disabled := ui_image.NewBorderedNineSliceColor(
        color.NRGBA{R: 38, G: 40, B: 46, A: 255},
        color.NRGBA{R: 77, G: 77, B: 73, A: 255}, 2,
    )
    for i := range frames {
        glow := (1 - math.Cos(2*math.Pi*float64(i)/frameCount)) / 2
        border := color.NRGBA{R: 210 + uint8(35*glow), G: 160 + uint8(40*glow), B: 72 + uint8(40*glow), A: 255}
        frames[i] = &widget.ButtonImage{
            Idle: ui_image.NewBorderedNineSliceColor(
                color.NRGBA{R: 86 + uint8(15*glow), G: 63 + uint8(12*glow), B: 29 + uint8(6*glow), A: 255}, border, 2),
            Hover: ui_image.NewBorderedNineSliceColor(
                color.NRGBA{R: 119 + uint8(12*glow), G: 88 + uint8(10*glow), B: 39 + uint8(5*glow), A: 255}, border, 2),
            Pressed: pressed,
            Disabled: disabled,
        }
    }

    tick, currentFrame := 0, 0
    return widget.NewButton(
        widget.ButtonOpts.WidgetOpts(
            widget.WidgetOpts.MinSize(230, 64),
            widget.WidgetOpts.OnUpdate(func(w widget.HasWidget) {
                if w.GetWidget().Disabled {
                    tick = 0
                } else {
                    tick = (tick + 1) % period
                }
                frame := tick * frameCount / period
                if frame != currentFrame {
                    button := w.(*widget.Button)
                    widget.ButtonOpts.Image(frames[frame])(button)
                    button.Validate()
                    currentFrame = frame
                }
            }),
        ),
        widget.ButtonOpts.TextPadding(&widget.Insets{Top: 16, Bottom: 16, Left: 28, Right: 28}),
        widget.ButtonOpts.Image(frames[0]),
        widget.ButtonOpts.Text("Enter Battle", face, &widget.ButtonTextColor{
            Idle: color.NRGBA{R: 255, G: 234, B: 182, A: 255},
            Hover: color.White,
            Pressed: color.White,
            Disabled: color.NRGBA{R: 140, G: 140, B: 140, A: 255},
        }),
        widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) { clicked() }),
    )
}
