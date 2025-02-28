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
    "github.com/kazzmir/master-of-magic/lib/colorconv"
    "github.com/kazzmir/master-of-magic/util/common"

    "github.com/hajimehoshi/ebiten/v2"
    // "github.com/hajimehoshi/ebiten/v2/colorm"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
)

const ScreenWidth = 1300
const ScreenHeight = 900

type State int
const (
    NormalState State = iota
    ChooseColorState
)

type BandChoice int
const (
    BandH BandChoice = iota
    BandS
    BandV

    BandRed
    BandGreen
    BandBlue

    BandMax
)

type HSVColor struct {
    H, S, V float64
}

func (hsv HSVColor) ToColor() color.Color {
    /*
    var colorm colorm.ColorM
    colorm.ChangeHSV(hsv.H, hsv.S, hsv.V)
    return colorm.Apply(color.RGBA{R: 0xff, A: 0xff})
    */

    h := hsv.H * 180 / math.Pi
    if h < 0 {
        h += 360
    }
    if h >= 360 {
        h -= 360
    }
    s := hsv.S
    v := hsv.V

    // log.Printf("H: %.2f S: %.2f V: %.2f", h, s, v)

    out, err := colorconv.HSVToColor(h, s, v)
    if err != nil {
        panic(err)
    }
    return out
}

type Editor struct {
    Lbx *lbx.LbxFile
    Palette color.Palette
    GlyphPosition image.Point
    GlyphImage *image.Paletted
    Rune rune
    FontIndex int
    Fonts []*font.LbxFont
    Optimized *font.Font
    Scale float64
    State State
    TextFont *text.GoTextFaceSource

    Band BandChoice

    CurrentColor HSVColor
    ColorBand *ebiten.Image
    SaturationBand *ebiten.Image
    ValueBand *ebiten.Image

    RedBand *ebiten.Image
    GreenBand *ebiten.Image
    BlueBand *ebiten.Image
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

    textFont, err := common.LoadFont()
    if err != nil {
        fmt.Printf("Could not load font: %v\n", err)
    }

    return &Editor{
        Lbx: lbxFile,
        Optimized: optimized,
        FontIndex: 0,
        Fonts: fonts,
        Rune: 'F',
        GlyphPosition: image.Pt(0, 0),
        GlyphImage: fonts[0].GlyphForRune('F').MakeImage(),
        Scale: 6,
        Band: BandH,
        Palette: palette,
        CurrentColor: HSVColor{math.Pi * 90 / 180, 1, 1},
        TextFont: textFont,
    }, nil
}

func (editor *Editor) UpdateFont() {
    lbxFont := editor.Fonts[editor.FontIndex]
    palette := makePaletteForFont(lbxFont)

    for i := range len(editor.Palette) {
        if i < len(palette) {
            palette[i] = editor.Palette[i]
        }
    }

    editor.Palette = palette

    editor.Optimized = font.MakeOptimizedFontWithPalette(lbxFont, editor.Palette)
    editor.GlyphImage = lbxFont.GlyphForRune(editor.Rune).MakeImageWithPalette(editor.Palette)

    if editor.GlyphPosition.X >= editor.GlyphImage.Bounds().Dx() {
        editor.GlyphPosition.X = editor.GlyphImage.Bounds().Dx() - 1
    }

    if editor.GlyphPosition.Y >= editor.GlyphImage.Bounds().Dy() {
        editor.GlyphPosition.Y = editor.GlyphImage.Bounds().Dy() - 1
    }
}

func (editor *Editor) UpdateBandColor(delta float64) {
    updateRGB := func (deltaR int, deltaG int, deltaB int) {
        r, g, b, _ := editor.CurrentColor.ToColor().RGBA()
        r >>= 8
        g >>= 8
        b >>= 8
        r = uint32(min(0xff, max(0, int(r) + int(deltaR))))
        g = uint32(min(0xff, max(0, int(g) + int(deltaG))))
        b = uint32(min(0xff, max(0, int(b) + int(deltaB))))

        h, s, v := colorconv.RGBToHSV(uint8(r), uint8(g), uint8(b))
        editor.CurrentColor.H = h * math.Pi / 180
        editor.CurrentColor.S = s
        editor.CurrentColor.V = v
    }

    switch editor.Band {
        case BandH:
            editor.CurrentColor.H += math.Pi * delta / 180 / 2
            if editor.CurrentColor.H < 0 {
                editor.CurrentColor.H += 2 * math.Pi
            }
            if editor.CurrentColor.H >= 2 * math.Pi {
                editor.CurrentColor.H -= 2 * math.Pi
            }
        case BandS:
            editor.CurrentColor.S = min(1, max(0, editor.CurrentColor.S + 0.01 * delta))
        case BandV:
            editor.CurrentColor.V = min(1, max(0, editor.CurrentColor.V + 0.01 * delta))
        case BandRed:
            updateRGB(int(delta), 0, 0)
        case BandGreen:
            updateRGB(0, int(delta), 0)
        case BandBlue:
            updateRGB(0, 0, int(delta))
    }

    editor.ColorBand = nil
    editor.SaturationBand = nil
    editor.ValueBand = nil

    /*
    log.Printf("H: %.2f S: %.2f V: %.2f", editor.CurrentColor.H, editor.CurrentColor.S, editor.CurrentColor.V)
    log.Printf("Color: %v", editor.CurrentColor.ToColor())
    */
}

func (editor *Editor) ChangeBand(current BandChoice, delta int) BandChoice {
    current += BandChoice(delta)
    if current >= BandMax {
        current = 0
    }
    if current < 0 {
        current = BandMax - 1
    }

    return current
}

func (editor *Editor) NextBand() {
    editor.Band = editor.ChangeBand(editor.Band, 1)
}

func (editor *Editor) PreviousBand() {
    editor.Band = editor.ChangeBand(editor.Band, -1)
}

func (editor *Editor) SavePalette() {
    fmt.Printf("color.Palette{\n")
    for _, c := range editor.Palette {
        r, g, b, a := c.RGBA()
        fmt.Printf("    color.RGBA{R: 0x%x, G: 0x%x, B: 0x%x, A: 0x%x},\n", r >> 8, g >> 8, b >> 8, a >> 8)
    }
    fmt.Printf("}\n")
}

func (editor *Editor) Update() error {
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendPressedKeys(keys)

    leftShift := ebiten.IsKeyPressed(ebiten.KeyShift)

    speed := 1.0
    if leftShift {
        speed = 3
    }

    for _, key := range keys {
        switch key {
            case ebiten.KeyLeft:
                switch editor.State {
                    case ChooseColorState:
                        editor.UpdateBandColor(-speed)
                }
            case ebiten.KeyRight:
                switch editor.State {
                    case ChooseColorState:
                        editor.UpdateBandColor(speed)
                }
        }
    }

    inputs := ebiten.AppendInputChars(nil)
    for _, r := range inputs {
        if r == ' ' {
            continue
        }
        editor.Rune = r
        editor.UpdateFont()
    }

    keys = make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyEscape, ebiten.KeyCapsLock:
                return ebiten.Termination
                /*
            case ebiten.KeyLeft:
                switch editor.State {
                    case NormalState:
                        editor.FontIndex -= 1
                        if editor.FontIndex < 0 {
                            editor.FontIndex = len(editor.Fonts) - 1
                        }

                        editor.UpdateFont()

                        fmt.Printf("Font: %v\n", editor.FontIndex)
                }
                */
            case ebiten.KeyF1:
                editor.SavePalette()

            case ebiten.KeySpace:
                switch editor.State {
                    case NormalState:
                        paletteIndex := editor.GlyphImage.ColorIndexAt(editor.GlyphPosition.X, editor.GlyphPosition.Y)
                        // reset to alpha=0
                        editor.Palette[paletteIndex] = color.RGBA{}
                        editor.UpdateFont()
                }

            case ebiten.KeyTab:
                switch editor.State {
                    case NormalState:
                        if leftShift {
                            editor.FontIndex -= 1
                            if editor.FontIndex < 0 {
                                editor.FontIndex = len(editor.Fonts) - 1
                            }
                        } else {
                            editor.FontIndex = (editor.FontIndex + 1) % len(editor.Fonts)
                        }
                        editor.UpdateFont()
                        fmt.Printf("Font: %v\n", editor.FontIndex)
                }
            case ebiten.KeyLeft:
                switch editor.State {
                    case NormalState:
                        editor.GlyphPosition.X -= 1
                        if editor.GlyphPosition.X < 0 {
                            editor.GlyphPosition.X = editor.GlyphImage.Bounds().Dx() - 1
                        }
                }
            case ebiten.KeyRight:
                switch editor.State {
                    case NormalState:
                        editor.GlyphPosition.X += 1
                        if editor.GlyphPosition.X >= editor.GlyphImage.Bounds().Dx() {
                            editor.GlyphPosition.X = 0
                        }
                }
            case ebiten.KeyUp:
                switch editor.State {
                    case NormalState:
                        editor.GlyphPosition.Y -= 1
                        if editor.GlyphPosition.Y < 0 {
                            editor.GlyphPosition.Y = editor.GlyphImage.Bounds().Dy() - 1
                        }
                    case ChooseColorState:
                        editor.PreviousBand()
                }
                // editor.Scale *= 1.05
            case ebiten.KeyDown:
                switch editor.State {
                    case NormalState:
                        editor.GlyphPosition.Y += 1
                        if editor.GlyphPosition.Y >= editor.GlyphImage.Bounds().Dy() {
                            editor.GlyphPosition.Y = 0
                        }
                    case ChooseColorState:
                        editor.NextBand()
                }
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

                        paletteIndex := editor.GlyphImage.ColorIndexAt(editor.GlyphPosition.X, editor.GlyphPosition.Y)

                        editor.Palette[paletteIndex] = editor.CurrentColor.ToColor()
                        editor.Optimized = font.MakeOptimizedFontWithPalette(editor.Fonts[editor.FontIndex], editor.Palette)
                        editor.GlyphImage = editor.Fonts[editor.FontIndex].GlyphForRune(editor.Rune).MakeImageWithPalette(editor.Palette)
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
        vector.DrawFilledRect(out, float32(x), 0, float32(1), float32(height), use.ToColor(), true)
    }

    return out
}

func (editor *Editor) CreateSaturationBand(hsv HSVColor, width int, height int) *ebiten.Image {
    out := ebiten.NewImage(width, height)

    for x := range width {
        use := hsv
        use.S = float64(x) / float64(width)
        vector.DrawFilledRect(out, float32(x), 0, float32(1), float32(height), use.ToColor(), true)
    }

    // draw yellow line at current saturation
    x := float32(hsv.S * float64(width))
    vector.DrawFilledRect(out, x, 0, 1, float32(height), color.RGBA{R: 0xff, G: 0xff, A: 0xff}, true)

    return out
}

func (editor *Editor) CreateValueBand(hsv HSVColor, width int, height int) *ebiten.Image {
    out := ebiten.NewImage(width, height)

    for x := range width {
        use := hsv
        use.V = float64(x) / float64(width)
        vector.DrawFilledRect(out, float32(x), 0, float32(1), float32(height), use.ToColor(), true)
    }

    x := float32(hsv.V * float64(width))
    vector.DrawFilledRect(out, x, 0, 1, float32(height), color.RGBA{R: 0xff, G: 0xff, A: 0xff}, true)

    return out
}

// for r,g,b values
func (editor *Editor) CreateSolidBand(rgb color.RGBA, width int, height int) *ebiten.Image {
    out := ebiten.NewImage(width, height)

    r, g, b, a := rgb.RGBA()
    r = r >> 8
    g = g >> 8
    b = b >> 8
    a = a >> 8

    for x := range width {
        r1 := uint8(float64(r) * float64(x) / float64(width))
        g1 := uint8(float64(g) * float64(x) / float64(width))
        b1 := uint8(float64(b) * float64(x) / float64(width))

        vector.DrawFilledRect(out, float32(x), 0, float32(1), float32(height), color.RGBA{R: r1, G: g1, B: b1, A: uint8(a)}, true)
    }

    return out
}

func (editor *Editor) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})

    // vector.DrawFilledRect(screen, 90, 90, 100, 100, &color.RGBA{R: 0xff, A: 0xff}, true)

    paletteRect := image.Rect(ScreenWidth - 200, 0, ScreenWidth, ScreenHeight)
    paletteArea := screen.SubImage(paletteRect).(*ebiten.Image)
    paletteArea.Fill(color.RGBA{32, 32, 32, 0xff})

    colorSize := 20

    minY := 20

    {
        var opts text.DrawOptions
        opts.GeoM.Translate(float64(paletteRect.Min.X + 1), float64(paletteRect.Min.Y + 1))
        opts.ColorScale.ScaleWithColor(color.White)
        face := &text.GoTextFace{Source: editor.TextFont, Size: 15}
        text.Draw(screen, fmt.Sprintf("Palette %v", len(editor.Palette)), face, &opts)

        opts.GeoM.Reset()
        opts.GeoM.Translate(1, 1)
        text.Draw(screen, fmt.Sprintf("Font %v", editor.FontIndex), face, &opts)
    }

    paletteIndex := int(editor.GlyphImage.ColorIndexAt(editor.GlyphPosition.X, editor.GlyphPosition.Y))
    for i, c := range editor.Palette {
        area := image.Rect(paletteRect.Min.X, paletteRect.Min.Y + minY + i * colorSize, paletteRect.Max.X, paletteRect.Min.Y + minY + (i + 1) * colorSize)
        vector.DrawFilledRect(paletteArea, float32(area.Min.X), float32(area.Min.Y+1), float32(area.Dx()), float32(area.Dy()-2), c, true)

        borderColor := color.RGBA{R: 0xff, A: 0xff}
        if i == paletteIndex {
            borderColor = color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff}
        }

        vector.StrokeRect(screen, float32(area.Min.X-1), float32(area.Min.Y-1), float32(area.Dx()+2), float32(area.Dy()+2), 2, borderColor, true)

        var opts text.DrawOptions
        opts.GeoM.Translate(float64(area.Min.X - 22), float64(area.Min.Y + 1))
        opts.ColorScale.ScaleWithColor(color.White)
        face := &text.GoTextFace{Source: editor.TextFont, Size: 15}
        text.Draw(screen, fmt.Sprintf("%v", i), face, &opts)
    }

    minX := 50
    minY = 20

    width := 7
    height := 7
    for x := range editor.GlyphImage.Bounds().Dx() {
        for y := range editor.GlyphImage.Bounds().Dy() {

            x1 := float32(minX + x * width * int(editor.Scale))
            y1 := float32(minY + y * height * int(editor.Scale))
            x2 := x1 + float32(width * int(editor.Scale))
            y2 := y1 + float32(height * int(editor.Scale))

            value := editor.GlyphImage.At(x, y)
            vector.DrawFilledRect(screen, x1 + 3, y1 + 3, x2 - x1 - 6, y2 - y1 - 6, value, true)

            borderColor := color.RGBA{A: 0xff}
            if image.Pt(x, y) == editor.GlyphPosition {
                borderColor = color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff}
            }

            vector.StrokeRect(screen, x1+1, y1+1, x2 - x1 - 2, y2 - y1 - 2, 1, borderColor, true)
            // vector.StrokeRect(screen, x1, y1, 3, 3, 1, color.RGBA{A: 0xff}, true)

            var opts text.DrawOptions
            opts.GeoM.Translate(float64(x1 + 1), float64(y1 + 1))
            opts.ColorScale.ScaleWithColor(color.White)
            face := &text.GoTextFace{Source: editor.TextFont, Size: 13}
            text.Draw(screen, fmt.Sprintf("%v", editor.GlyphImage.ColorIndexAt(x, y)), face, &opts)
        }
    }

    yPos := editor.GlyphImage.Bounds().Dy() * height * int(editor.Scale) + 10 + 10

    editor.Optimized.Print(screen, 50, float64(yPos), editor.Scale, ebiten.ColorScale{}, "abcdefghijkl")
    yPos += editor.Optimized.Height() * int(editor.Scale) + 2
    editor.Optimized.Print(screen, 50, float64(yPos), editor.Scale, ebiten.ColorScale{}, "ABCDEFGHIJKL")

    if editor.State == ChooseColorState {
        yellow := color.RGBA{R: 0xff, G: 0xff, A: 0xff}

        theColor := editor.CurrentColor.ToColor()

        vector.DrawFilledRect(screen, 850, 50, 100, 50, theColor, true)
        if editor.ColorBand == nil {
            editor.ColorBand = editor.CreateColorBand(editor.CurrentColor, 200, 40)
        }
        if editor.SaturationBand == nil {
            editor.SaturationBand = editor.CreateSaturationBand(editor.CurrentColor, 200, 40)
        }
        if editor.ValueBand == nil {
            editor.ValueBand = editor.CreateValueBand(editor.CurrentColor, 200, 40)
        }

        var opts text.DrawOptions
        opts.GeoM.Translate(600, 20)
        opts.ColorScale.ScaleWithColor(color.White)
        face := &text.GoTextFace{Source: editor.TextFont, Size: 15}
        text.Draw(screen, fmt.Sprintf("H: %.2f S: %.2f V: %.2f", editor.CurrentColor.H, editor.CurrentColor.S, editor.CurrentColor.V), face, &opts)

        var options ebiten.DrawImageOptions
        options.GeoM.Translate(600, 50)
        screen.DrawImage(editor.ColorBand, &options)

        if editor.Band == BandH {
            x1, y1 := options.GeoM.Apply(-1, -1)
            vector.StrokeRect(screen, float32(x1), float32(y1), float32(editor.ColorBand.Bounds().Dx() + 2), float32(editor.ColorBand.Bounds().Dy() + 2), 1, yellow, true)
        }

        opts.GeoM = options.GeoM
        opts.GeoM.Translate(-20, 8)
        text.Draw(screen, "H", face, &opts)

        options.GeoM.Translate(0, float64(editor.ColorBand.Bounds().Dy() + 1))
        screen.DrawImage(editor.SaturationBand, &options)

        if editor.Band == BandS {
            x1, y1 := options.GeoM.Apply(-1, -1)
            vector.StrokeRect(screen, float32(x1), float32(y1), float32(editor.SaturationBand.Bounds().Dx() + 2), float32(editor.SaturationBand.Bounds().Dy() + 2), 1, yellow, true)
        }

        opts.GeoM = options.GeoM
        opts.GeoM.Translate(-20, 8)
        text.Draw(screen, "S", face, &opts)

        options.GeoM.Translate(0, float64(editor.SaturationBand.Bounds().Dy() + 1))
        screen.DrawImage(editor.ValueBand, &options)

        if editor.Band == BandV {
            x1, y1 := options.GeoM.Apply(-1, -1)
            vector.StrokeRect(screen, float32(x1), float32(y1), float32(editor.ValueBand.Bounds().Dx() + 2), float32(editor.ValueBand.Bounds().Dy() + 2), 1, yellow, true)
        }

        opts.GeoM = options.GeoM
        opts.GeoM.Translate(-20, 8)
        text.Draw(screen, "V", face, &opts)

        r, g, b, a := theColor.RGBA()
        r >>= 8
        g >>= 8
        b >>= 8
        a >>= 8

        options.GeoM.Translate(0, float64(editor.ValueBand.Bounds().Dy() + 20))
        opts.GeoM = options.GeoM
        text.Draw(screen, fmt.Sprintf("R: 0x%x (%v) G: 0x%x (%v) B: 0x%x (%v)", r, r, g, g, b, b), face, &opts)

        if editor.RedBand == nil {
            editor.RedBand = editor.CreateSolidBand(color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, 200, 40)
        }
        if editor.GreenBand == nil {
            editor.GreenBand = editor.CreateSolidBand(color.RGBA{R: 0, G: 0xff, B: 0, A: 0xff}, 200, 40)
        }
        if editor.BlueBand == nil {
            editor.BlueBand = editor.CreateSolidBand(color.RGBA{R: 0, G: 0, B: 0xff, A: 0xff}, 200, 40)
        }

        options.GeoM.Translate(0, 20)
        screen.DrawImage(editor.RedBand, &options)

        if editor.Band == BandRed {
            x1, y1 := options.GeoM.Apply(-1, -1)
            vector.StrokeRect(screen, float32(x1), float32(y1), float32(editor.RedBand.Bounds().Dx() + 2), float32(editor.RedBand.Bounds().Dy() + 2), 1, yellow, true)
        }

        opts.GeoM = options.GeoM
        opts.GeoM.Translate(-20, 8)
        text.Draw(screen, "R", face, &opts)

        px, py := options.GeoM.Apply(float64(int(r) * editor.RedBand.Bounds().Dx()) / 0xff, 0)
        vector.StrokeLine(screen, float32(px), float32(py), float32(px), float32(py) + float32(editor.RedBand.Bounds().Dy()), 1, yellow, true)

        options.GeoM.Translate(0, float64(editor.RedBand.Bounds().Dy() + 2))
        screen.DrawImage(editor.GreenBand, &options)

        if editor.Band == BandGreen {
            x1, y1 := options.GeoM.Apply(-1, -1)
            vector.StrokeRect(screen, float32(x1), float32(y1), float32(editor.GreenBand.Bounds().Dx() + 2), float32(editor.GreenBand.Bounds().Dy() + 2), 1, yellow, true)
        }

        opts.GeoM = options.GeoM
        opts.GeoM.Translate(-20, 8)
        text.Draw(screen, "G", face, &opts)

        px, py = options.GeoM.Apply(float64(int(g) * editor.RedBand.Bounds().Dx()) / 0xff, 0)
        vector.StrokeLine(screen, float32(px), float32(py), float32(px), float32(py) + float32(editor.GreenBand.Bounds().Dy()), 1, yellow, true)

        options.GeoM.Translate(0, float64(editor.GreenBand.Bounds().Dy() + 2))
        screen.DrawImage(editor.BlueBand, &options)

        if editor.Band == BandBlue {
            x1, y1 := options.GeoM.Apply(-1, -1)
            vector.StrokeRect(screen, float32(x1), float32(y1), float32(editor.BlueBand.Bounds().Dx() + 2), float32(editor.BlueBand.Bounds().Dy() + 2), 1, yellow, true)
        }

        opts.GeoM = options.GeoM
        opts.GeoM.Translate(-20, 8)
        text.Draw(screen, "B", face, &opts)

        px, py = options.GeoM.Apply(float64(int(b) * editor.RedBand.Bounds().Dx()) / 0xff, 0)
        vector.StrokeLine(screen, float32(px), float32(py), float32(px), float32(py) + float32(editor.BlueBand.Bounds().Dy()), 1, yellow, true)
    } else {
        var opts text.DrawOptions
        opts.GeoM.Translate(600, 20)
        opts.ColorScale.ScaleWithColor(color.White)
        face := &text.GoTextFace{Source: editor.TextFont, Size: 15}

        lines := []string{
            "Keys",
            "Tab: Change font",
            "Space: Clear color (alpha=0)",
            "Enter: Choose color",
            "Up/Down: Change band while choosing color",
            "Up/Down/Left/Right: Move cursor while not choosing color",
            "Any letter key: Change glyph",
        }

        for _, line := range lines {
            text.Draw(screen, line, face, &opts)
            opts.GeoM.Translate(0, 20)
        }
    }

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
