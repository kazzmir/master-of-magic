package main

import (
    "log"
    "fmt"

    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type Editor struct {
    Lbx *lbx.LbxFile
    Palette color.Palette
    GlyphImage *image.Paletted
    FontIndex int
    Fonts []*font.LbxFont
    Optimized *font.Font
    Scale float64
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
    }, nil
}

func (editor *Editor) Update() error {
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyUp:
                editor.Scale *= 1.05
            case ebiten.KeyDown:
                editor.Scale *= 0.95
                if editor.Scale < 1 {
                    editor.Scale = 1
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
                editor.FontIndex -= 1
                if editor.FontIndex < 0 {
                    editor.FontIndex = len(editor.Fonts) - 1
                }

                editor.Optimized = font.MakeOptimizedFont(editor.Fonts[editor.FontIndex])

                fmt.Printf("Font: %v\n", editor.FontIndex)
            case ebiten.KeyRight:
                editor.FontIndex = (editor.FontIndex + 1) % len(editor.Fonts)
                editor.Optimized = font.MakeOptimizedFont(editor.Fonts[editor.FontIndex])
                fmt.Printf("Font: %v\n", editor.FontIndex)
        }
    }

    return nil
}

func (editor *Editor) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return ScreenWidth, ScreenHeight
}

func (editor *Editor) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})

    // vector.DrawFilledRect(screen, 90, 90, 100, 100, &color.RGBA{R: 0xff, A: 0xff}, true)

    paletteRect := image.Rect(800, 0, 1024, 768)
    paletteArea := screen.SubImage(paletteRect).(*ebiten.Image)
    paletteArea.Fill(color.RGBA{0xff, 0xff, 0xff, 0xff})

    for i, c := range editor.Palette {
        area := image.Rect(paletteRect.Min.X, paletteRect.Min.Y + 10 + i * 20, paletteRect.Max.X, paletteRect.Min.Y + 10 + (i + 1) * 20)
        vector.DrawFilledRect(paletteArea, float32(area.Min.X), float32(area.Min.Y), float32(area.Dx()), float32(area.Dy()), c, true)
        vector.StrokeRect(paletteArea, float32(area.Min.X), float32(area.Min.Y), float32(area.Dx()), float32(area.Dy()), 1, &color.RGBA{R: 0xff, A: 0xff}, true)
    }

    editor.Optimized.Print(screen, 50, 300, editor.Scale, ebiten.ColorScale{}, "This is a test")

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
