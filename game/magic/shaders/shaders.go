package shaders

import (
    _ "embed"
    "github.com/hajimehoshi/ebiten/v2"
)

type Shader int
const (
    ShaderEdgeGlow Shader = iota
)

//go:embed edge-glow.kage
var edgeGlowShader []byte

func LoadEdgeGlowShader() (*ebiten.Shader, error) {
    return ebiten.NewShader(edgeGlowShader)
}
