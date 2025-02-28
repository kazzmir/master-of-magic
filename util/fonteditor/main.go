package main

import (
    "log"
    "fmt"
    "math"
    "math/rand/v2"

    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/colorm"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type State int
const (
    NormalState State = iota
    ChooseColorState
)

type HSVColor struct {
    H, S, V float64
}

func (hsv HSVColor) ToColor() color.Color {
    var colorm colorm.ColorM
    colorm.ChangeHSV(hsv.H, hsv.S, hsv.V)
    return colorm.Apply(color.RGBA{R: 0xff, A: 0xff})
}

type Editor struct {
    Lbx *lbx.LbxFile
    Palette color.Palette
    PaletteIndex int
    GlyphImage *image.Paletted
    FontIndex int
    Fonts []*font.LbxFont
    Optimized *font.Font
    Scale float64
    State State

    CurrentColor HSVColor
    ColorBand *ebiten.Image
}

// go through each glyph and find the highest palette index used, then make a palette of that many entries
func makePaletteForFont(font *font.LbxFont) color.Palette {
    highest := 0
    for i, glyph := range font.Glyphs {
        img := glyph.MakeImage()
        if img != nil {
            for _, pixel := range img.Pix {
                if int(pixel) > highest {
                    highest = int(pixel)
                }
            }
        } else {
            log.Printf("Warning: nil image for glyph %v", i)
        }
    }

    out := make(color.Palette, highest + 1)

    out[0] = color.RGBA{0, 0, 0, 0}
    for i := range len(out) - 1 {
        out[i+1] = color.RGBA{0, 0, 0, 0xff}
    }

    return out
}

func randomColor() color.RGBA {
    return color.RGBA{R: uint8(rand.N(256)), G: uint8(rand.N(256)), B: uint8(rand.N(256)), A: 0xff}
}

func MakeEditor() (*Editor, error) {
    cache := lbx.AutoCache()

    lbxFile, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        fmt.Printf("Could not load fonts.lbx: %v\n", err)
        return nil, err
    }

    fonts, err := font.ReadFonts(lbxFile, 0)
    if err != nil {
        return nil, err
    }

    palette := makePaletteForFont(fonts[0])
    optimized := font.MakeOptimizedFontWithPalette(fonts[0], palette)

    log.Printf("Palette length: %v", len(palette))
    // log.Printf("Palette: %v", palette)

    return &Editor{
        Lbx: lbxFile,
        Optimized: optimized,
        FontIndex: 0,
        Fonts: fonts,
        GlyphImage: fonts[0].GlyphForRune('T').MakeImage(),
        Scale: 4,
        Palette: palette,
        PaletteIndex: 0,
        CurrentColor: HSVColor{math.Pi * 90 / 180, 1, 1},
    }, nil
}

func (editor *Editor) Update() error {
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyLeft:
                switch editor.State {
                    case ChooseColorState:
                        editor.CurrentColor.H -= math.Pi * 1 / 180
                        editor.ColorBand = nil
                }
            case ebiten.KeyRight:
                switch editor.State {
                    case ChooseColorState:
                        editor.CurrentColor.H += math.Pi * 1 / 180
                        editor.ColorBand = nil
                }
        }
    }

    keys = make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyEscape, ebiten.KeyCapsLock:
                return ebiten.Termination
            case ebiten.KeyLeft:
                switch editor.State {
                    case NormalState:
                        editor.FontIndex -= 1
                        if editor.FontIndex < 0 {
                            editor.FontIndex = len(editor.Fonts) - 1
                        }

                        editor.Optimized = font.MakeOptimizedFont(editor.Fonts[editor.FontIndex])

                        fmt.Printf("Font: %v\n", editor.FontIndex)
                }
            case ebiten.KeyRight:
                switch editor.State {
                    case NormalState:
                        editor.FontIndex = (editor.FontIndex + 1) % len(editor.Fonts)
                        editor.Optimized = font.MakeOptimizedFont(editor.Fonts[editor.FontIndex])
                        fmt.Printf("Font: %v\n", editor.FontIndex)
                }
            case ebiten.KeyUp:
                editor.PaletteIndex = max(0, editor.PaletteIndex - 1)
                // editor.Scale *= 1.05
            case ebiten.KeyDown:
                editor.PaletteIndex = min(len(editor.Palette) - 1, editor.PaletteIndex + 1)
                /*
                editor.Scale *= 0.95
                if editor.Scale < 1 {
                    editor.Scale = 1
                }
                */
            case ebiten.KeyEnter:
                switch editor.State {
                    case NormalState:
                        editor.State = ChooseColorState
                    case ChooseColorState:
                        editor.State = NormalState
                        editor.Palette[editor.PaletteIndex] = editor.CurrentColor.ToColor()
                        editor.Optimized = font.MakeOptimizedFontWithPalette(editor.Fonts[editor.FontIndex], editor.Palette)
                }

        }
    }

    return nil
}

func (editor *Editor) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return ScreenWidth, ScreenHeight
}

func (editor *Editor) CreateColorBand(color HSVColor, width int, height int) *ebiten.Image {
    out := ebiten.NewImage(width, height)

    middle := width / 2
    for x := range width {
        use := color
        use.H = color.H + (math.Pi / 2 * float64(x - middle) / float64(width / 2))
        vector.DrawFilledRect(out, float32(x), 0, float32(x+1), float32(height), use.ToColor(), true)
    }

    return out
}

func (editor *Editor) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})

    // vector.DrawFilledRect(screen, 90, 90, 100, 100, &color.RGBA{R: 0xff, A: 0xff}, true)

    paletteRect := image.Rect(800, 0, 1024, 768)
    paletteArea := screen.SubImage(paletteRect).(*ebiten.Image)
    paletteArea.Fill(color.RGBA{32, 32, 32, 0xff})

    colorSize := 20

    for i, c := range editor.Palette {
        area := image.Rect(paletteRect.Min.X, paletteRect.Min.Y + 10 + i * colorSize, paletteRect.Max.X, paletteRect.Min.Y + 10 + (i + 1) * colorSize)
        vector.DrawFilledRect(paletteArea, float32(area.Min.X), float32(area.Min.Y), float32(area.Dx()), float32(area.Dy()), c, true)

        borderColor := color.RGBA{R: 0xff, A: 0xff}
        if i == editor.PaletteIndex {
            borderColor = color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff}
        }

        vector.StrokeRect(paletteArea, float32(area.Min.X), float32(area.Min.Y), float32(area.Dx()), float32(area.Dy()), 1, borderColor, true)
    }

    editor.Optimized.Print(screen, 50, 300, editor.Scale, ebiten.ColorScale{}, "This is a test")

    if editor.State == ChooseColorState {
        vector.DrawFilledRect(screen, 600, 50, 100, 50, editor.CurrentColor.ToColor(), true)
        if editor.ColorBand == nil {
            editor.ColorBand = editor.CreateColorBand(editor.CurrentColor, 200, 40)
        }
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(650 - editor.ColorBand.Bounds().Dx() / 2), float64(50 - editor.ColorBand.Bounds().Dy()))
        screen.DrawImage(editor.ColorBand, &options)
    }

    /*
    var options ebiten.DrawImageOptions
    options.GeoM.Scale(editor.Scale, editor.Scale)
    options.GeoM.Translate(50, 50)

    screen.DrawImage(editor.Optimized.Image, &options)

    vector.StrokeRect(screen, 50, 500, float32(float64(editor.Optimized.GlyphWidth) * 20 * editor.Scale), float32(float64(editor.Optimized.GlyphHeight) * editor.Scale), 1, &color.RGBA{R: 0xff, A: 0xff}, true)
    editor.Optimized.Print(screen, 50, 500, editor.Scale, ebiten.ColorScale{}, "Hello, potato! money")
    */

    /*
    yPos := 1

    for _, font := range viewer.Fonts {

        for i := 0; i < font.GlyphCount(); i++ {
            raw := font.Glyphs[i].MakeImage()
            if raw == nil {
                continue
            }
            glyph1 := ebiten.NewImageFromImage(raw)
            screen.DrawImage(glyph1, &options)
            options.GeoM.Translate(25, 0)
        }

        yPos += 80
        options.GeoM.Reset()
        options.GeoM.Scale(4, 4)
        options.GeoM.Translate(1, float64(yPos))
    }
    */
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
    ebiten.SetWindowTitle("font palette editor")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    editor, err := MakeEditor()
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }

    err = ebiten.RunGame(editor)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
