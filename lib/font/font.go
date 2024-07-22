package font

import (
    "math"

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
    }
}
