package inputmanager

import (
    "github.com/hajimehoshi/ebiten/v2"
)

// MoveDelta returns the tile step for a unit-movement key. Arrow keys are the
// four cardinals; numpad 1-4 and 6-9 are the original MoM 8-direction keypad
// (5 is unused). Y grows south, matching the overland map.
func MoveDelta(key ebiten.Key) (dx int, dy int, ok bool) {
    switch key {
        case ebiten.KeyUp, ebiten.KeyNumpad8:
            return 0, -1, true
        case ebiten.KeyDown, ebiten.KeyNumpad2:
            return 0, 1, true
        case ebiten.KeyLeft, ebiten.KeyNumpad4:
            return -1, 0, true
        case ebiten.KeyRight, ebiten.KeyNumpad6:
            return 1, 0, true
        case ebiten.KeyNumpad7:
            return -1, -1, true
        case ebiten.KeyNumpad9:
            return 1, -1, true
        case ebiten.KeyNumpad1:
            return -1, 1, true
        case ebiten.KeyNumpad3:
            return 1, 1, true
    }

    return 0, 0, false
}

func IsNumpadMoveKey(key ebiten.Key) bool {
    switch key {
        case ebiten.KeyNumpad1, ebiten.KeyNumpad2, ebiten.KeyNumpad3, ebiten.KeyNumpad4,
             ebiten.KeyNumpad6, ebiten.KeyNumpad7, ebiten.KeyNumpad8, ebiten.KeyNumpad9:
            return true
    }

    return false
}

// CombineMoveDeltas folds movement keys into a single step. Cardinals set one
// axis so Up+Left in the same frame is northwest; a diagonal numpad key sets
// both. Later keys override the axes they touch.
func CombineMoveDeltas(keys []ebiten.Key) (dx int, dy int) {
    for _, key := range keys {
        kdx, kdy, ok := MoveDelta(key)
        if !ok {
            continue
        }
        if kdx != 0 {
            dx = kdx
        }
        if kdy != 0 {
            dy = kdy
        }
    }

    return dx, dy
}
