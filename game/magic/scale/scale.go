package scale

import (
    "golang.org/x/exp/constraints"
    "image"

    "github.com/hajimehoshi/ebiten/v2"
)

var ScaleAmount = 3.0
var ScaledGeom ebiten.GeoM

var ScreenScaleAlgorithm = ScaleAlgorithmNormal

type ScaleAlgorithm int

const (
    // the scale2x
    // https://www.scale2x.it/
    ScaleAlgorithmScale ScaleAlgorithm = iota
    ScaleAlgorithmXbr
    ScaleAlgorithmNormal
)

func (algorithm ScaleAlgorithm) String() string {
    switch algorithm {
        case ScaleAlgorithmScale: return "scale"
        case ScaleAlgorithmXbr: return "xbr"
        case ScaleAlgorithmNormal: return "normal"
    }

    return ""
}

type UnscaledGeoM ebiten.GeoM

func (unscaled *UnscaledGeoM) Scaled() ebiten.GeoM {
    var scaled ebiten.GeoM = ebiten.GeoM(*unscaled)
    scaled.Scale(ScaleAmount, ScaleAmount)
    return scaled
}

type Number interface {
    constraints.Integer | constraints.Float
}

func Scale[T Number](x T) T {
    return T(float64(x) * ScaleAmount)
}

func Scale2[T Number](x, y T) (T, T) {
    return T(float64(x) * ScaleAmount), T(float64(y) * ScaleAmount)
}

func Unscale[T Number](x T) T {
    return T(float64(x) / ScaleAmount)
}

func Unscale2[T Number](x, y T) (T, T) {
    return T(float64(x) / ScaleAmount), T(float64(y) / ScaleAmount)
}

func ScaleGeom(geom ebiten.GeoM) ebiten.GeoM {
    geom.Scale(ScaleAmount, ScaleAmount)
    return geom
}

func ScaleOptions(options ebiten.DrawImageOptions) *ebiten.DrawImageOptions {
    options.GeoM.Scale(ScaleAmount, ScaleAmount)
    return &options
}

func DefaultScaleOptions() *ebiten.DrawImageOptions {
    return ScaleOptions(ebiten.DrawImageOptions{})
}

func init(){
    ScaledGeom.Scale(ScaleAmount, ScaleAmount)
}

func UpdateScale(amount float64) {
    ScaleAmount = amount
    ScaledGeom.Reset()
    ScaledGeom.Scale(ScaleAmount, ScaleAmount)
}

// draw the image using the current scale, but avoid allocating an entire DrawImageOptions
func DrawScaled(screen *ebiten.Image, img *ebiten.Image, options *ebiten.DrawImageOptions) {
    oldGeom := options.GeoM
    options.GeoM.Concat(ScaledGeom)
    screen.DrawImage(img, options)
    options.GeoM = oldGeom
}

func ScaleRect(rect image.Rectangle) image.Rectangle {
    return image.Rectangle{
        Min: rect.Min.Mul(int(ScaleAmount)),
        Max: rect.Max.Mul(int(ScaleAmount)),
    }
}

func UnscaleRect(rect image.Rectangle) image.Rectangle {
    return image.Rectangle{
        Min: rect.Min.Div(int(ScaleAmount)),
        Max: rect.Max.Div(int(ScaleAmount)),
    }
}
