package font

import (
    "math"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/hajimehoshi/ebiten/v2"
)

type Font struct {
    Image *ebiten.Image
}

func MakeGPUSpriteMap(font *lbx.Font) *ebiten.Image {
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

    return sheet
}

func MakeOptimizedFont(font *lbx.Font) *Font {
    sheet := MakeGPUSpriteMap(font)
    _ = sheet
    return nil
}
