//go:build !js
package inputmanager

import (
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

func IsQuitKey(key ebiten.Key) bool {
    return key == ebiten.KeyEscape || key == ebiten.KeyCapsLock
}

func IsQuitPressed() bool {
    return inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyCapsLock)
}
