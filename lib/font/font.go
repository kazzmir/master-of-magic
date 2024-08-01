package font

import (
    "math"
    _ "fmt"
    "image"
    "image/color"
    "strings"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/hajimehoshi/ebiten/v2"
)

type Font struct {
    Image *ebiten.Image
    GlyphWidth int
    GlyphHeight int
    Rows int
    Columns int
    Glyphs []lbx.Glyph
    internalFont *lbx.Font
}

func MakeGPUSpriteMap(font *lbx.Font) (*ebiten.Image, int, int, int, int) {
    return MakeGPUSpriteMapWithPalette(font, lbx.GetDefaultPalette())
}

func MakeGPUSpriteMapWithPalette(font *lbx.Font, palette color.Palette) (*ebiten.Image, int, int, int, int) {
    // make one huge ebiten.Image and draw all glyphs onto it, keeping track of the location of each glyph

    totalGlyphs := font.GlyphCount()

    // find max width of glyphs
    maxWidth := 0
    for _, glyph := range font.Glyphs {
        if glyph.Width > maxWidth {
            maxWidth = glyph.Width
        }
    }

    columns := int(math.Ceil(math.Sqrt(float64(totalGlyphs))))
    rows := int(math.Ceil(float64(totalGlyphs) / float64(columns)))

    sheet := ebiten.NewImage(columns * maxWidth, rows * font.Height)

    x := 0
    y := 0

    for _, glyph := range font.Glyphs {
        raw := glyph.MakeImageWithPalette(palette)
        if raw != nil {
            posX := x * maxWidth
            posY := y * font.Height

            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(posX), float64(posY))
            sheet.DrawImage(ebiten.NewImageFromImage(raw), &options)
        }

        x += 1
        if x >= columns {
            x = 0
            y += 1
        }
    }

    return sheet, maxWidth, font.Height, rows, columns
}

func MakeOptimizedFont(font *lbx.Font) *Font {
    return MakeOptimizedFontWithPalette(font, lbx.GetDefaultPalette())
}

func MakeOptimizedFontWithPalette(font *lbx.Font, palette color.Palette) *Font {
    sheet, width, height, rows, columns := MakeGPUSpriteMapWithPalette(font, palette)

    return &Font{
        Image: sheet,
        GlyphWidth: width,
        GlyphHeight: height,
        Rows: rows,
        Columns: columns,
        Glyphs: font.Glyphs,
        internalFont: font,
    }
}

func (font *Font) Height() int {
    return font.internalFont.Height
}

func (font *Font) getGlyphImage(index int) *ebiten.Image {
    x := index % font.Columns
    y := index / font.Columns

    x1 := x * font.GlyphWidth
    y1 := y * font.GlyphHeight
    x2 := (x+1) * font.GlyphWidth
    y2 := (y+1) * font.GlyphHeight

    return font.Image.SubImage(image.Rect(x1, y1, x2, y2)).(*ebiten.Image)
}

func (font *Font) Print(image *ebiten.Image, x float64, y float64, scale float64, text string) {
    useX := x
    for _, c := range text {
        if c == '\n' {
            y += float64(font.GlyphHeight + font.internalFont.VerticalSpacing)
            useX = 0
            continue
        }

        glyphIndex := int(c) - 32
        if glyphIndex >= len(font.Glyphs) || glyphIndex < 0 {
            continue
        }

        glyph := font.Glyphs[glyphIndex]

        var options ebiten.DrawImageOptions
        options.GeoM.Scale(scale, scale)
        options.GeoM.Translate(useX, y)
        glyphImage := font.getGlyphImage(glyphIndex)
        image.DrawImage(glyphImage, &options)

        useX += float64(glyph.Width + font.internalFont.HorizontalSpacing) * scale
    }
}

func (font *Font) MeasureTextWidth(text string, scale float64) float64 {
    width := 0
    for _, c := range text {
        if c == '\n' {
            continue
        }

        glyphIndex := int(c) - 32
        if glyphIndex >= len(font.Glyphs) || glyphIndex < 0 {
            continue
        }

        glyph := font.Glyphs[glyphIndex]
        width += glyph.Width + font.internalFont.HorizontalSpacing
    }

    return float64(width) * scale
}

func (font *Font) PrintCenter(image *ebiten.Image, x float64, y float64, scale float64, text string) {
    width := font.MeasureTextWidth(text, scale)
    font.Print(image, x - width / 2, y, scale, text)
}

/* split the input text ABCD into two substrings AB and CD such that the pixel width of AB is less than maxWidth */
func (font *Font) splitText(text string, maxWidth float64, scale float64) (string, string) {
    size := font.MeasureTextWidth(text, scale)
    if size < maxWidth {
        return text, ""
    }

    parts := strings.Split(text, " ")
    // FIXME: use binary search for this
    for i := len(parts) - 1; i >= 0; i-- {
        sofar := strings.Join(parts[0:i], " ")
        if font.MeasureTextWidth(sofar, scale) < maxWidth {
            return sofar, strings.Join(parts[i:len(parts)], " ")
        }
    }

    return "", text
}

func (font *Font) PrintWrap(image *ebiten.Image, x float64, y float64, maxWidth float64, scale float64, text string) {
    wrapped := font.CreateWrappedText(maxWidth, scale, text)
    font.RenderWrapped(image, x, y, wrapped)
}

type WrappedText struct {
    Lines []string
    TotalHeight float64
    MaxWidth float64
    Scale float64
}

func (font *Font) RenderWrapped(image *ebiten.Image, x float64, y float64, wrapped WrappedText) {
    yPos := y
    for _, line := range wrapped.Lines {
        font.Print(image, x, yPos, wrapped.Scale, line)
        yPos += float64(font.Height()) * wrapped.Scale + 1
    }
}

// precompute an object that can be used to render a wrapped string
func (font *Font) CreateWrappedText(maxWidth float64, scale float64, text string) WrappedText {
    var lines []string
    var yPos float64 = 0

    for text != "" {
        show, rest := font.splitText(text, maxWidth, scale)

        // we were unable to split the text, just bail
        if show == "" {
            break
        }

        lines = append(lines, show)

        yPos += float64(font.Height()) * scale + 1

        text = rest
    }

    return WrappedText{
        Lines: lines,
        TotalHeight: yPos,
        MaxWidth: maxWidth,
        Scale: scale,
    }
}
