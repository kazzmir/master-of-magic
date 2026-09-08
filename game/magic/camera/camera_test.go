package camera

import (
    "math"
    "testing"

    "github.com/kazzmir/master-of-magic/game/magic/data"
)

func TestGetZoomedYClampsNorthEdge(test *testing.T) {
    cam := MakeCameraAt(10, 0)
    cam.SetMapHeight(50)

    if got := cam.GetZoomedY(); got != 0 {
        test.Errorf("north edge GetZoomedY=%v, want 0 (map edge on viewport edge)", got)
    }
}

func TestGetZoomedYClampsSouthEdge(test *testing.T) {
    height := 50
    cam := MakeCameraAt(10, height-1)
    cam.SetMapHeight(height)

    bottom := cam.GetZoomedMaxY()
    if bottom > float64(height)+1e-9 {
        test.Errorf("south edge viewport bottom %v past map height %d", bottom, height)
    }
    if cam.GetZoomedY() < 0 {
        test.Errorf("south edge GetZoomedY=%v should not go north of 0", cam.GetZoomedY())
    }

    visible := float64(data.ScreenHeight) / 18.0
    wantY := float64(height) - visible
    if math.Abs(cam.GetZoomedY()-wantY) > 1e-9 {
        test.Errorf("south edge GetZoomedY=%v, want %v so the last row sits on the bottom of the screen", cam.GetZoomedY(), wantY)
    }
}

func TestGetZoomedYLeavesInteriorUnclamped(test *testing.T) {
    cam := MakeCameraAt(10, 25)
    cam.SetMapHeight(50)

    // default zoom 1, SizeY 10 → unclamped Y - 5
    want := 20.0
    if got := cam.GetZoomedY(); got != want {
        test.Errorf("interior GetZoomedY=%v, want %v (old centering)", got, want)
    }
}

func TestGetZoomedYWithoutMapHeightKeepsOldFormula(test *testing.T) {
    cam := MakeCameraAt(10, 0)

    // city-preview / unset height: Y=0 used to show 5 tiles of empty space
    want := -5.0
    if got := cam.GetZoomedY(); got != want {
        test.Errorf("unset MapHeight GetZoomedY=%v, want %v", got, want)
    }
}

func TestCenterClampsToMapRows(test *testing.T) {
    cam := MakeCamera()
    cam.SetMapHeight(40)

    cam.Center(5, 99)
    if cam.GetY() != 39 {
        test.Errorf("Center past south: Y=%d, want 39", cam.GetY())
    }

    cam.Center(5, -3)
    if cam.GetY() != 0 {
        test.Errorf("Center past north: Y=%d, want 0", cam.GetY())
    }
}
