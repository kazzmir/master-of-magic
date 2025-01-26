package shaders

import (
    "testing"
)

func TestLoadShaders(test *testing.T) {
    _, err := LoadEdgeGlowShader()

    if err != nil {
        test.Errorf("Error loading edge glow shader: %v", err)
    }

    _, err = LoadWarpShader()

    if err != nil {
        test.Errorf("Error loading distortion shader: %v", err)
    }
}
