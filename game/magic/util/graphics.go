package util

import (
    "image"
    "image/color"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

func PremultiplyAlpha(c color.RGBA) color.RGBA {
    a := float64(c.A) / 255.0
    return color.RGBA{
        R: uint8(float64(c.R) * a),
        G: uint8(float64(c.G) * a),
        B: uint8(float64(c.B) * a),
        A: c.A,
    }
}

func DrawRect(screen *ebiten.Image, rect image.Rectangle, color_ color.Color){
    vector.StrokeRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), 1, color_, false)
}

func ImageRect(x int, y int, img *ebiten.Image) image.Rectangle {
    return image.Rect(x, y, x + img.Bounds().Dx(), y + img.Bounds().Dy())
}
