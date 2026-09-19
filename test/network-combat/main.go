package main

import (
    "log"

    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/mouse"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
}

func NewEngine() (*Engine, error) {
    return &Engine{}, nil
}

func (engine *Engine) Update() error {

    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (int, int) {
    return scale.Scale2(data.ScreenWidth, data.ScreenHeight)
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    log.Printf("Initializing")

    monitorWidth, _ := ebiten.Monitor().Size()
    size := monitorWidth / 390
    ebiten.SetWindowSize(data.ScreenWidth * size, data.ScreenHeight * size)

    ebiten.SetWindowTitle("combat screen")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
    ebiten.SetCursorMode(ebiten.CursorModeHidden)

    audio.Initialize()
    mouse.Initialize()

    engine, err := NewEngine()

    if err != nil {
        log.Printf("Error: unable to load engine: %v", err)
        return
    }

    log.Printf("Running")

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }

    log.Printf("Bye")
}
