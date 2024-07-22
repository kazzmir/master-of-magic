package font

import (
    "math"
    _ "fmt"
    "image"

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
        raw := glyph.MakeImage()
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
    sheet, width, height, rows, columns := MakeGPUSpriteMap(font)

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
