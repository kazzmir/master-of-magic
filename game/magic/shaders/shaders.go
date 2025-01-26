package shaders

import (
    _ "embed"
    "github.com/hajimehoshi/ebiten/v2"
)

type Shader int
const (
    ShaderEdgeGlow Shader = iota
    ShaderGlitch
)

//go:embed edge-glow.kage
var edgeGlowShader []byte

//go:embed glitch.kage
var distortionShader []byte

func LoadEdgeGlowShader() (*ebiten.Shader, error) {
    return ebiten.NewShader(edgeGlowShader)
}

func LoadGlitchShader() (*ebiten.Shader, error) {
    return ebiten.NewShader(distortionShader)
}
