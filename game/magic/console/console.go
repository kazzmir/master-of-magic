package console

import (
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/game/magic/util"

    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2"
)

type ConsoleState int

const (
    ConsoleOpen ConsoleState = iota
    ConsoleClosed
)

const ConsoleHeight = 130

type Console struct {
    CurrentLine string
    State ConsoleState
    PosY int
}

func MakeConsole() *Console {
    return &Console{
        State: ConsoleClosed,
    }
}

func (console *Console) IsActive() bool {
    return console.State == ConsoleOpen
}

func (console *Console) Update() {
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        switch key {
        case ebiten.KeyBackquote:
            if console.State == ConsoleClosed {
                console.State = ConsoleOpen
            } else if console.State == ConsoleOpen {
                console.State = ConsoleClosed
            }
        }
    }

    const speed = 8

    if console.State == ConsoleOpen {
        if console.PosY < ConsoleHeight {
            console.PosY += speed
            if console.PosY > ConsoleHeight {
                console.PosY = ConsoleHeight
            }
        }
    } else {
        if console.PosY > 0 {
            console.PosY -= speed
        }
    }

    if console.IsActive() {
        runes := make([]rune, 0)
        runes = ebiten.AppendInputChars(runes)

        for _, r := range runes {
            if r == '\n' {
                console.CurrentLine = ""
            } else {
                console.CurrentLine += string(r)
            }
        }
    }
}

func (console *Console) Draw(screen *ebiten.Image) {
    if console.PosY > 0 {
        rect := image.Rect(0, 0, screen.Bounds().Dx(), console.PosY)
        backgroundColor := util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 120})
        vector.DrawFilledRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), backgroundColor, false)
    }
}
