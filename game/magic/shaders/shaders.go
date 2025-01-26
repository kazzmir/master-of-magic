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

func LoadEdgeGlowShader() (*ebiten.Shader, error) {
    return ebiten.NewShader(edgeGlowShader)
}

func LoadWarpShader() (*ebiten.Shader, error) {
    return ebiten.NewShader(warpShader)
}
