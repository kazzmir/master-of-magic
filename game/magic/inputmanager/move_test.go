package inputmanager

import (
    "testing"

    "github.com/hajimehoshi/ebiten/v2"
)

func TestMoveDeltaArrows(test *testing.T) {
    checks := []struct {
        key ebiten.Key
        dx int
        dy int
    }{
        {ebiten.KeyUp, 0, -1},
        {ebiten.KeyDown, 0, 1},
        {ebiten.KeyLeft, -1, 0},
        {ebiten.KeyRight, 1, 0},
    }

    for _, check := range checks {
        dx, dy, ok := MoveDelta(check.key)
        if !ok || dx != check.dx || dy != check.dy {
            test.Errorf("%v: got (%d,%d) ok=%v, want (%d,%d)", check.key, dx, dy, ok, check.dx, check.dy)
        }
    }
}

func TestMoveDeltaNumpadEightWay(test *testing.T) {
    checks := []struct {
        key ebiten.Key
        dx int
        dy int
    }{
        {ebiten.KeyNumpad8, 0, -1},
        {ebiten.KeyNumpad2, 0, 1},
        {ebiten.KeyNumpad4, -1, 0},
        {ebiten.KeyNumpad6, 1, 0},
        {ebiten.KeyNumpad7, -1, -1},
        {ebiten.KeyNumpad9, 1, -1},
        {ebiten.KeyNumpad1, -1, 1},
        {ebiten.KeyNumpad3, 1, 1},
    }

    for _, check := range checks {
        dx, dy, ok := MoveDelta(check.key)
        if !ok || dx != check.dx || dy != check.dy {
            test.Errorf("%v: got (%d,%d) ok=%v, want (%d,%d)", check.key, dx, dy, ok, check.dx, check.dy)
        }
        if !IsNumpadMoveKey(check.key) {
            test.Errorf("%v should be a numpad move key", check.key)
        }
    }

    if _, _, ok := MoveDelta(ebiten.KeyNumpad5); ok {
        test.Errorf("numpad 5 should not move a unit")
    }
    if IsNumpadMoveKey(ebiten.KeyNumpad5) {
        test.Errorf("numpad 5 should not be a numpad move key")
    }
    if IsNumpadMoveKey(ebiten.KeyUp) {
        test.Errorf("arrow keys are not numpad move keys")
    }
}

func TestCombineMoveDeltas(test *testing.T) {
    dx, dy := CombineMoveDeltas([]ebiten.Key{ebiten.KeyUp, ebiten.KeyLeft})
    if dx != -1 || dy != -1 {
        test.Errorf("Up+Left should be northwest, got (%d,%d)", dx, dy)
    }

    dx, dy = CombineMoveDeltas([]ebiten.Key{ebiten.KeyNumpad3})
    if dx != 1 || dy != 1 {
        test.Errorf("numpad 3 should be southeast, got (%d,%d)", dx, dy)
    }

    dx, dy = CombineMoveDeltas([]ebiten.Key{ebiten.KeyN, ebiten.KeySpace})
    if dx != 0 || dy != 0 {
        test.Errorf("non-move keys should be ignored, got (%d,%d)", dx, dy)
    }
}
