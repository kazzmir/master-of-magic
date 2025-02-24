package shaders

import (
    _ "embed"
    "github.com/hajimehoshi/ebiten/v2"
)

type Shader int
const (
    ShaderEdgeGlow Shader = iota
    ShaderWarp
)

//go:embed edge-glow.kage
var edgeGlowShader []byte

//go:embed warp.kage
var warpShader []byte

//go:embed drop-shadow.kage
var dropShadowShader []byte

//go:embed outline.kage
var outlineShader []byte

func LoadEdgeGlowShader() (*ebiten.Shader, error) {
    return ebiten.NewShader(edgeGlowShader)
}

func LoadWarpShader() (*ebiten.Shader, error) {
    return ebiten.NewShader(warpShader)
}

func LoadDropShadowShader() (*ebiten.Shader, error) {
    return ebiten.NewShader(dropShadowShader)
}

func LoadOutlineShader() (*ebiten.Shader, error) {
    return ebiten.NewShader(outlineShader)
}
