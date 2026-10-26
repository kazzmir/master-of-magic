package engine

import (
	"image"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// portraitPreview retains normal button input/focus behavior, but draws the
// original portrait at the largest aspect-preserving size inside the button.
// Resizing changes only the draw transform, not the cached image.
type portraitPreview struct {
	*widget.Button
	portrait *ebiten.Image
}

func newPortraitPreview(portrait *ebiten.Image, clicked func()) *portraitPreview {
	return &portraitPreview{portrait: portrait, Button: widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(64, 80)),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) { clicked() }),
	)}
}

func fitPortrait(source image.Point, bounds image.Rectangle) (scale, x, y float64) {
	if source.X <= 0 || source.Y <= 0 || bounds.Empty() {
		return 0, 0, 0
	}
	scale = min(float64(bounds.Dx())/float64(source.X), float64(bounds.Dy())/float64(source.Y))
	x = float64(bounds.Min.X) + (float64(bounds.Dx())-float64(source.X)*scale)/2
	y = float64(bounds.Min.Y) + (float64(bounds.Dy())-float64(source.Y)*scale)/2
	return
}

func (preview *portraitPreview) Render(screen *ebiten.Image) {
	preview.Button.Render(screen)
	if preview.portrait == nil {
		return
	}
	scale, x, y := fitPortrait(preview.portrait.Bounds().Size(), preview.GetWidget().Rect.Inset(4))
	if scale == 0 {
		return
	}
	options := ebiten.DrawImageOptions{}
	options.GeoM.Scale(scale, scale)
	options.GeoM.Translate(x, y)
	screen.DrawImage(preview.portrait, &options)
}
