package ui

import (
    "image/color"

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

func HBox() *widget.Container {
    return widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
            widget.RowLayoutOpts.Spacing(4),
        )),
    )
}

func VBox(opts ...widget.ContainerOpt) *widget.Container {

    allArgs := append(opts, widget.ContainerOpts.Layout(widget.NewRowLayout(
        widget.RowLayoutOpts.Direction(widget.DirectionVertical),
        widget.RowLayoutOpts.Spacing(4),
        widget.RowLayoutOpts.Padding(widget.Insets{Top: 2, Bottom: 2, Left: 2, Right: 2}),
    )))

    return widget.NewContainer(allArgs...)
}

func CenteredText(textStr string, face *text.GoTextFace, textColor color.Color) *widget.Text {
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
