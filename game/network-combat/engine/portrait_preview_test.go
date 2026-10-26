package engine

import (
	"image"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestPortraitAspectFit(t *testing.T) {
	for _, test := range []struct {
		bounds      image.Rectangle
		scale, x, y float64
	}{
		{image.Rect(0, 0, 100, 100), 1.25, 10, 0},
		{image.Rect(0, 0, 200, 100), 1.25, 60, 0},
		{image.Rect(0, 0, 100, 200), 1.5625, 0, 37.5},
		{image.Rect(10, 20, 110, 120), 1.25, 20, 20},
	} {
		scale, x, y := fitPortrait(image.Pt(64, 80), test.bounds)
		if math.Abs(scale-test.scale) > 1e-9 || math.Abs(x-test.x) > 1e-9 || math.Abs(y-test.y) > 1e-9 {
			t.Fatal("portrait stretched or not centered", scale, x, y, test)
		}
	}
	if scale, _, _ := fitPortrait(image.Pt(64, 80), image.Rectangle{}); scale != 0 {
		t.Fatal("empty component should not draw")
	}
}

func TestPortraitComponentExpands(t *testing.T) {
	game, _, _ := setupEditorGame(t, 1)
	editor := game.matches.editor
	game.Draw(ebiten.NewImage(1100, 700))
	bounds := editor.portraitButton.GetWidget().Rect.Inset(4)
	if bounds.Dx() <= 64 && bounds.Dy() <= 80 {
		t.Fatal("portrait component did not expand with its panel", bounds)
	}
	original := editor.portraitButton.portrait
	scale, _, _ := fitPortrait(original.Bounds().Size(), bounds)
	width, height := float64(original.Bounds().Dx())*scale, float64(original.Bounds().Dy())*scale
	if math.Abs(width-float64(bounds.Dx())) > 1e-9 && math.Abs(height-float64(bounds.Dy())) > 1e-9 {
		t.Fatal("portrait does not fill either component dimension")
	}
	game.matches.window.SetLocation(image.Rect(0, 0, 900, 620))
	game.Draw(ebiten.NewImage(1100, 700))
	if editor.portraitButton.portrait != original {
		t.Fatal("resizing replaced cached portrait")
	}
}
