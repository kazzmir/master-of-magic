package main

import (
    "log"

    "image/color"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type Viewer struct {
}

func (viewer *Viewer) Update() error {
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    return nil
}

func (viewer *Viewer) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return ScreenWidth, ScreenHeight
}

func (viewer *Viewer) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
    ebiten.SetWindowTitle("lbx viewer")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    var viewer Viewer

    err := ebiten.RunGame(&viewer)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
