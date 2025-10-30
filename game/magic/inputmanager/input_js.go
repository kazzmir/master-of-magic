//go:build js
package inputmanager

import (
    "github.com/hajimehoshi/ebiten/v2"
)

// don't support the quit key on browsers so that people don't accidentally quit, since it doesn't make a lot of sense in that case

func IsQuitKey(key ebiten.Key) bool {
    return false
}

func IsQuitPressed() bool {
    return false
}
