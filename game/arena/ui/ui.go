package ui

import (
    "image/color"

    // "github.com/ebitenui/ebitenui"
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
