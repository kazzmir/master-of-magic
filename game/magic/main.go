package main

import (
    "log"

    "github.com/hajimehoshi/ebiten/v2"
)

const ScreenWidth = 1024
const ScreenHeight = 768

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
    ebiten.SetWindowTitle("magic")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
}
