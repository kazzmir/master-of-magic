package ui

import (
    "image/color"
    "math"

    // "github.com/ebitenui/ebitenui"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
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

// IconStrip scales and packs its images to fit the allocated area without
// resizing source images or allocating new GPU images when the window changes.
type IconStrip struct {
    *widget.Widget
    images []*ebiten.Image
    dirty bool
}

func NewIconStrip() *IconStrip {
    return &IconStrip{Widget: widget.NewWidget()}
}

func (strip *IconStrip) GetWidget() *widget.Widget { return strip.Widget }
func (strip *IconStrip) Validate() {}

func (strip *IconStrip) SetImages(images []*ebiten.Image) {
    strip.images = images
    strip.dirty = true
}

func (strip *IconStrip) imageSize() (width, height, count, minWidth int) {
    for _, img := range strip.images {
        if img != nil {
            width += img.Bounds().Dx()
            height = max(height, img.Bounds().Dy())
            if count == 0 || img.Bounds().Dx() < minWidth {
                minWidth = img.Bounds().Dx()
            }
            count++
        }
    }
    return
}

func (strip *IconStrip) PreferredSize() (int, int) {
    width, height, count, _ := strip.imageSize()
    return width + max(0, count-1)*4, height
}

func (strip *IconStrip) Update(update *widget.UpdateObject) {
    strip.Widget.Update(update)
    if strip.dirty {
        update.RelayoutRequested = true
        strip.dirty = false
    }
}

func fitIconStrip(totalWidth, minWidth, maxHeight, count, width, height int) (scale, gap float64) {
    if totalWidth <= 0 || minWidth <= 0 || maxHeight <= 0 || width <= 0 || height <= 0 || count <= 0 {
        return 0, 0
    }
    // A negative gap overlaps books. Bound it by half the narrowest book's
    // width so at least half of every image remains visible, even with mixed
    // image sizes. Only shrink once that maximum overlap is insufficient.
    maxOverlap := float64(minWidth)/2
    packedWidth := float64(totalWidth) - float64(count-1)*maxOverlap
    scale = math.Min(1, math.Min(float64(height)/float64(maxHeight), float64(width)/packedWidth))
    if count > 1 {
        gap = math.Max(-maxOverlap*scale, math.Min(4, (float64(width)-float64(totalWidth)*scale)/float64(count-1)))
    }
    return
}

func (strip *IconStrip) Render(screen *ebiten.Image) {
    if !strip.Widget.IsVisible() {
        return
    }
    strip.Widget.Render(screen)
    width, height, count, minWidth := strip.imageSize()
    rect := strip.Widget.Rect
    scale, gap := fitIconStrip(width, minWidth, height, count, rect.Dx(), rect.Dy())
    if scale == 0 {
        return
    }
    x := float64(rect.Min.X)
    for _, img := range strip.images {
        if img == nil {
            continue
        }
        var options ebiten.DrawImageOptions
        options.GeoM.Scale(scale, scale)
        options.GeoM.Translate(x, float64(rect.Min.Y)+(float64(rect.Dy())-float64(img.Bounds().Dy())*scale)/2)
        screen.DrawImage(img, &options)
        x += float64(img.Bounds().Dx())*scale + gap
    }
}

// Checkbox is a labeled, keyboard-accessible checkbox using the arena palette.
func Checkbox(label string, face *text.Face, checked bool, changed func(bool)) *widget.Checkbox {
    makeImage := func(selected, hover bool) *ui_image.NineSlice {
        img := ebiten.NewImage(22, 22)
        img.Fill(color.NRGBA{R: 32, G: 40, B: 48, A: 255})
        border := color.NRGBA{R: 125, G: 153, B: 145, A: 255}
        if hover {
            border = color.NRGBA{R: 160, G: 220, B: 185, A: 255}
        }
        vector.StrokeRect(img, 1, 1, 20, 20, 2, border, false)
        if selected {
            tick := color.NRGBA{R: 145, G: 225, B: 165, A: 255}
            vector.StrokeLine(img, 5, 11, 9, 16, 3, tick, true)
            vector.StrokeLine(img, 9, 16, 17, 6, 3, tick, true)
        }
        return ui_image.NewFixedNineSlice(img)
    }
    state := widget.WidgetUnchecked
    if checked {
        state = widget.WidgetChecked
    }
    return widget.NewCheckbox(
        widget.CheckboxOpts.Image(&widget.CheckboxImage{
            Unchecked: makeImage(false, false),
            UncheckedHovered: makeImage(false, true),
            Checked: makeImage(true, false),
            CheckedHovered: makeImage(true, true),
        }),
        widget.CheckboxOpts.Text(label, face, &widget.LabelColor{Idle: color.White, Disabled: color.Gray{Y: 128}}),
        widget.CheckboxOpts.Spacing(10),
        widget.CheckboxOpts.InitialState(state),
        widget.CheckboxOpts.StateChangedHandler(func(args *widget.CheckboxChangedEventArgs) {
            if changed != nil {
                changed(args.State == widget.WidgetChecked)
            }
        }),
    )
}

// NavigationButtonImage gives each section its own palette, with a brighter
// fill and border for the selected tab (the button's pressed state).
func NavigationButtonImage(base, accent color.NRGBA) *widget.ButtonImage {
    blend := func(a, b color.NRGBA, amount int) color.NRGBA {
        channel := func(x, y uint8) uint8 {
            return uint8((int(x)*(100-amount) + int(y)*amount)/100)
        }
        return color.NRGBA{R: channel(a.R, b.R), G: channel(a.G, b.G), B: channel(a.B, b.B), A: 255}
    }
    return &widget.ButtonImage{
        Idle: ui_image.NewBorderedNineSliceColor(base, blend(base, accent, 45), 2),
        Hover: ui_image.NewBorderedNineSliceColor(blend(base, accent, 20), accent, 2),
        Pressed: ui_image.NewBorderedNineSliceColor(blend(base, accent, 40), accent, 2),
        PressedHover: ui_image.NewBorderedNineSliceColor(blend(base, accent, 50), accent, 2),
        Disabled: SolidImage(38, 40, 46),
    }
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
