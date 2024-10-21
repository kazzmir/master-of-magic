package console

import (
    "log"

    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/game/magic/util"

    "github.com/hajimehoshi/ebiten/v2/text/v2"
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
    Lines []string
    State ConsoleState
    PosY int
    Font *text.GoTextFaceSource
}

func MakeConsole() *Console {
    font, err := LoadFont()
    if err != nil {
        log.Printf("Unable to load console font: %v", err)
        return nil
    }

    return &Console{
        Font: font,
        State: ConsoleClosed,
    }
}

func (console *Console) IsActive() bool {
    return console.State == ConsoleOpen
}

func (console *Console) Update() {
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    control := ebiten.IsKeyPressed(ebiten.KeyControl)

    for _, key := range keys {
        switch key {
            case ebiten.KeyBackquote:
                if console.State == ConsoleClosed {
                    console.State = ConsoleOpen
                } else if console.State == ConsoleOpen {
                    console.State = ConsoleClosed
                }
            case ebiten.KeyBackspace:
                if console.State == ConsoleOpen {
                    if len(console.CurrentLine) > 0 {
                        console.CurrentLine = console.CurrentLine[:len(console.CurrentLine) - 1]
                    }
                }
            case ebiten.KeyEnter:
                if console.State == ConsoleOpen {
                    if len(console.CurrentLine) > 0 {
                        console.Lines = append(console.Lines, console.CurrentLine)
                        console.CurrentLine = ""
                    }
                }
            case ebiten.KeyU:
                if console.State == ConsoleOpen {
                    if control {
                        console.CurrentLine = ""
                    }
                }
            case ebiten.KeyW:
                if console.State == ConsoleOpen {
                    if control {
                        if len(console.CurrentLine) > 0 {
                            i := len(console.CurrentLine) - 1
                            for i >= 0 && console.CurrentLine[i] == ' ' {
                                i -= 1
                            }

                            for i >= 0 && console.CurrentLine[i] != ' ' {
                                i -= 1
                            }

                            console.CurrentLine = console.CurrentLine[:i+1]
                        }
                    }
                }
            case ebiten.KeyUp:
                if console.State == ConsoleOpen {
                    if len(console.Lines) > 0 {
                        console.CurrentLine = console.Lines[len(console.Lines) - 1]
                    }
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
            if r == '`' {
                continue
            }

            if len(console.CurrentLine) < 120 {
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

        face := &text.GoTextFace{
            Source: console.Font,
            Size: 8,
        }

        textOptions := text.DrawOptions{}
        /*
        textOptions.Blend.BlendFactorSourceAlpha = ebiten.BlendFactorOne
        textOptions.Blend.BlendFactorDestinationAlpha = ebiten.BlendFactorZero
        */
        textOptions.GeoM.Translate(2, float64(console.PosY) - face.Size - 1)
        text.Draw(screen, "> " + console.CurrentLine + "|", face, &textOptions)

        for i, count := len(console.Lines) - 1, 0; i >= 0; i, count = i-1, count+1 {
            if count > 20 {
                break
            }
            line := console.Lines[i]
            textOptions.GeoM.Translate(0, -float64(face.Size))
            text.Draw(screen, line, face, &textOptions)
        }
    }
}
