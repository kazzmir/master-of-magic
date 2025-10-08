package console

import (
    "log"

    "strings"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"

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

type ConsoleEvent interface {
}

type ConsoleQuit struct {
}

type Console struct {
    CurrentLine string
    Lines []string
    InputLines []string
    State ConsoleState
    PosY int
    Font *text.GoTextFaceSource
    Events chan ConsoleEvent
    Game *gamelib.Game
}

func MakeConsole(game *gamelib.Game) *Console {
    font, err := LoadFont()
    if err != nil {
        log.Printf("Unable to load console font: %v", err)
        return nil
    }

    return &Console{
        Font: font,
        State: ConsoleClosed,
        Events: make(chan ConsoleEvent, 10),
        Game: game,
    }
}

func (console *Console) Run(command string) {
    parts := strings.Fields(command)

    if len(parts) == 0 {
        return
    }

    switch strings.ToLower(parts[0]) {
        case "quit":
            select {
                case console.Events <- &ConsoleQuit{}:
                default:
            }
        case "help":
            console.Lines = append(console.Lines, "Available commands:")
            console.Lines = append(console.Lines, "  quit - Quit the game")
            console.Lines = append(console.Lines, "  help - This help")
            console.Lines = append(console.Lines, "  cast <term> ... - Cast a spell. Search for spells given the terms,")
            console.Lines = append(console.Lines, "    such as 'cast gua wi' will cast Guardian Wind.")
        case "cast":
            if len(parts) == 1 {
                console.Lines = append(console.Lines, "Cast a spell. Give the name of the spell to cast. Partial names are ok.")
            } else {
                spells := console.Game.AllSpells()
                var use spellbook.Spells

                for _, spell := range spells.Spells {
                    if len(parts[1:]) == 1 && strings.ToLower(spell.Name) == strings.ToLower(parts[1]) {
                        use.Spells = nil
                        use.AddSpell(spell)
                        break
                    }

                    keep := true
                    for _, part := range parts[1:] {
                        if !strings.Contains(strings.ToLower(spell.Name), strings.ToLower(part)) {
                            keep = false
                            break
                        }
                    }

                    if keep {
                        use.AddSpell(spell)
                    }
                }

                if len(use.Spells) == 1 {
                    console.Lines = append(console.Lines, "Casting " + use.Spells[0].Name)
                    console.Game.Events <- &gamelib.GameEventCastSpell{
                        Player: console.Game.Players[0],
                        Spell: use.Spells[0],
                    }
                } else {
                    for _, spell := range use.Spells {
                        console.Lines = append(console.Lines, "  " + spell.Name)
                    }
                }
            }

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
                        console.InputLines = append(console.InputLines, console.CurrentLine)

                        console.Run(console.CurrentLine)

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
                    if len(console.InputLines) > 0 {
                        console.CurrentLine = console.InputLines[len(console.InputLines) - 1]
                    }
                }
        }
    }

    const speed = 9

    if console.State == ConsoleOpen {
        if console.PosY < ConsoleHeight * int(scale.ScaleAmount) {
            console.PosY += speed * int(scale.ScaleAmount)
            if console.PosY > ConsoleHeight * int(scale.ScaleAmount) {
                console.PosY = ConsoleHeight * int(scale.ScaleAmount)
            }
        }
    } else {
        if console.PosY > 0 {
            console.PosY -= speed * int(scale.ScaleAmount)
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
        vector.FillRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), backgroundColor, false)

        face := &text.GoTextFace{
            Source: console.Font,
            Size: float64(8 * scale.ScaleAmount),
        }

        textOptions := text.DrawOptions{}
        /*
        textOptions.Blend.BlendFactorSourceAlpha = ebiten.BlendFactorOne
        textOptions.Blend.BlendFactorDestinationAlpha = ebiten.BlendFactorZero
        */
        textOptions.GeoM.Translate(2, float64(console.PosY) - face.Size - float64(1 * scale.ScaleAmount))
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
