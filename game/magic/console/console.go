package console

import (
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/game/magic/util"

    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2"
)

type Console struct {
    Active bool
    CurrentLine string
}

func MakeConsole() *Console {
    return &Console{
        Active: false,
    }
}

func (console *Console) Update() {
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        switch key {
        case ebiten.KeyBackquote:
            console.Active = !console.Active
        }
    }

    if console.Active {
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
    if console.Active {
        rect := image.Rect(0, 0, screen.Bounds().Dx(), 160)
        backgroundColor := util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 120})
        vector.DrawFilledRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), backgroundColor, false)
    }
}
