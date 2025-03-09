package scale

import (
    "golang.org/x/exp/constraints"

    "github.com/hajimehoshi/ebiten/v2"
)

var ScaleAmount = 3.0
var ScaledGeom ebiten.GeoM

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
    return x * T(ScaleAmount)
}

func Scale2[T Number](x, y T) (T, T) {
    return x * T(ScaleAmount), y * T(ScaleAmount)
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
