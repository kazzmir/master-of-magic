//go:build js
package inputmanager

import (
    "github.com/hajimehoshi/ebiten/v2"
)

func IsQuitKey(key ebiten.Key) bool {
    return false
}

func IsQuitPressed() bool {
    return false
}
