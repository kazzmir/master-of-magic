package main

import (
    "log"

    "image/color"

    "github.com/hajimehoshi/ebiten/v2"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type MagicGame struct {
}

func (game *MagicGame) Update() error {
    return nil
}

func (game *MagicGame) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return ScreenWidth, ScreenHeight
}

func (game *MagicGame) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
    ebiten.SetWindowTitle("magic")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    var game MagicGame

    err := ebiten.RunGame(&game)
    if err != nil {
        log.Printf("Error: %v", err)
    }

}
