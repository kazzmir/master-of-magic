package font

import (
    "math"
    _ "fmt"
    "image"
    "image/color"
    "strings"
    "bytes"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/hajimehoshi/ebiten/v2"
)

type FontJustify int
const (
    FontJustifyLeft FontJustify = iota
    FontJustifyCenter
    FontJustifyRight
)

// for rendering, to specify justification, wrapping and shadows
type FontOptions struct {
    // left is default
    Justify FontJustify
    // no shadow is default
    DropShadow bool
    // if DropShadow is true, this is the color of the shadow, which defaults to black
    ShadowColor color.Color
    // if nil then the default options are used (no scaling, no color scaling)
    Options *ebiten.DrawImageOptions
    Scale float64
}

type Font struct {
    Image *ebiten.Image
    GlyphWidth int
    GlyphHeight int
    Rows int
    Columns int
    Glyphs []Glyph
    internalFont *LbxFont
    GlyphImages map[int]*ebiten.Image
}

func MakeGPUSpriteMap(font *LbxFont) (*ebiten.Image, int, int, int, int) {
    return MakeGPUSpriteMapWithPalette(font, lbx.GetDefaultPalette())
}

func MakeGPUSpriteMapWithPalette(font *LbxFont, palette color.Palette) (*ebiten.Image, int, int, int, int) {
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

        // FIXME: check that all the pixels referenced by the glyph are actually in the palette
        // meaning, if the palette doesn't have enough values in it then we will get a crash
        // at the ebiten.NewImageFromImage() call

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

func MakeOptimizedFont(font *LbxFont) *Font {
    return MakeOptimizedFontWithPalette(font, lbx.GetDefaultPalette())
}

func MakeOptimizedFontWithPalette(font *LbxFont, palette color.Palette) *Font {
    sheet, width, height, rows, columns := MakeGPUSpriteMapWithPalette(font, palette)

    return &Font{
        Image: sheet,
        GlyphWidth: width,
        GlyphHeight: height,
        Rows: rows,
        Columns: columns,
        Glyphs: font.Glyphs,
        internalFont: font,
        GlyphImages: make(map[int]*ebiten.Image),
    }
}

func (font *Font) Height() int {
    return font.internalFont.Height
}

func (font *Font) getGlyphImage(index int) *ebiten.Image {
    cached, ok := font.GlyphImages[index]
    if ok {
        return cached
    }

    x := index % font.Columns
    y := index / font.Columns

    x1 := x * font.GlyphWidth
    y1 := y * font.GlyphHeight
    x2 := (x+1) * font.GlyphWidth
    y2 := (y+1) * font.GlyphHeight

    sub := font.Image.SubImage(image.Rect(x1, y1, x2, y2)).(*ebiten.Image)
    font.GlyphImages[index] = sub
    return sub
}

func toFloatArray(color color.Color) []float32 {
    r, g, b, a := color.RGBA()
    var max float32 = 65535.0
    return []float32{float32(r) / max, float32(g) / max, float32(b) / max, float32(a) / max}
}

// draws a 1px outline around each glyph
// the edge shader should be shaders.ShaderEdgeGlow, although it should probably be changed to shaders.ShaderOutline
// just use PrintDropShadow for now
func (font *Font) PrintOutline(destination *ebiten.Image, edgeShader *ebiten.Shader, x float64, y float64, scale float64, colorScale ebiten.ColorScale, text string) {
    useX := x

    black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
    var shaderOptions ebiten.DrawRectShaderOptions
    shaderOptions.Uniforms = make(map[string]interface{})
    shaderOptions.Uniforms["Color1"] = toFloatArray(black)
    shaderOptions.Uniforms["Color2"] = toFloatArray(black)
    shaderOptions.Uniforms["Color3"] = toFloatArray(black)
    shaderOptions.Uniforms["Time"] = float32(0)
    shaderOptions.ColorScale = colorScale

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
        options.ColorScale = colorScale
        glyphImage := font.getGlyphImage(glyphIndex)
        destination.DrawImage(glyphImage, &options)

        shaderOptions.GeoM = options.GeoM
        shaderOptions.Images[0] = glyphImage
        destination.DrawRectShader(glyphImage.Bounds().Dx(), glyphImage.Bounds().Dy(), edgeShader, &shaderOptions)

        useX += float64(glyph.Width + font.internalFont.HorizontalSpacing) * scale
    }
}

func (font *Font) doPrint(destination *ebiten.Image, x float64, y float64, scale float64, colorScale ebiten.ColorScale, dropShadow bool, shadowColor color.Color, text string) {
    useX := x

    if shadowColor == nil {
        shadowColor = color.Black
    }

    // TODO: this shadow distance could be a parameter
    distance := 3.0/4

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
        glyphImage := font.getGlyphImage(glyphIndex)

        // draw the shadow first
        var options ebiten.DrawImageOptions
        options.GeoM.Scale(scale, scale)

        options.GeoM.Translate(useX, y)
        if dropShadow {
            options.GeoM.Translate(scale*distance, scale*distance)
            options.ColorScale = colorScale
            options.ColorScale.ScaleWithColor(shadowColor)
            destination.DrawImage(glyphImage, &options)

            // then draw the normal glyph on top
            options.GeoM.Translate(-scale*distance, -scale*distance)
        }

        options.ColorScale = colorScale
        destination.DrawImage(glyphImage, &options)

        useX += float64(glyph.Width + font.internalFont.HorizontalSpacing) * scale
    }
}

func (font *Font) PrintDropShadow(destination *ebiten.Image, x float64, y float64, scale float64, colorScale ebiten.ColorScale, text string) {
    font.doPrint(destination, x, y, scale, colorScale, true, color.Black, text)
}

// print the text with no border/outline
func (font *Font) Print(image *ebiten.Image, x float64, y float64, scale float64, colorScale ebiten.ColorScale, text string) {
    font.doPrint(image, x, y, scale, colorScale, false, color.Black, text)
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
        // FIXME: technically we don't need to add the horizontal spacing for the last character
        width += glyph.Width + font.internalFont.HorizontalSpacing
    }

    return float64(width) * scale
}

func (font *Font) PrintCenter(image *ebiten.Image, x float64, y float64, scale float64, colorScale ebiten.ColorScale, text string) {
    width := font.MeasureTextWidth(text, scale)
    font.Print(image, x - width / 2, y, scale, colorScale, text)
}

func (font *Font) PrintRight(image *ebiten.Image, x float64, y float64, scale float64, colorScale ebiten.ColorScale, text string) {
    width := font.MeasureTextWidth(text, scale)
    font.Print(image, x - width, y, scale, colorScale, text)
}

func (font *Font) PrintOptions(image *ebiten.Image, x float64, y float64, scale float64, colorScale ebiten.ColorScale, options FontOptions, text string) {
    useX := x
    useY := y

    switch options.Justify {
        case FontJustifyLeft:
        case FontJustifyCenter:
            width := font.MeasureTextWidth(text, scale)
            useX = x - width / 2
        case FontJustifyRight:
            width := font.MeasureTextWidth(text, scale)
            useX = x - width
    }

    if options.DropShadow {
        font.doPrint(image, useX, useY, scale, colorScale, true, options.ShadowColor, text)
    } else {
        font.doPrint(image, useX, useY, scale, colorScale, false, options.ShadowColor, text)
    }
}

func (font *Font) PrintOptions2(image *ebiten.Image, x float64, y float64, options FontOptions, text string) {
    useOptions := options.Options
    if useOptions == nil {
        useOptions = &ebiten.DrawImageOptions{}
    }

    scale := options.Scale
    if scale == 0 {
        scale = 1
    }

    useX, useY := x * options.Scale, y * options.Scale

    switch options.Justify {
        case FontJustifyLeft:
        case FontJustifyCenter:
            width := font.MeasureTextWidth(text, scale)
            useX = useX - width / 2
        case FontJustifyRight:
            width := font.MeasureTextWidth(text, scale)
            useX = useX - width
    }

    if options.DropShadow {
        font.doPrint(image, useX, useY, scale, useOptions.ColorScale, true, options.ShadowColor, text)
    } else {
        font.doPrint(image, useX, useY, scale, useOptions.ColorScale, false, options.ShadowColor, text)
    }
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

func (font *Font) PrintWrap(image *ebiten.Image, x float64, y float64, maxWidth float64, scale float64, colorScale ebiten.ColorScale, options FontOptions, text string) {
    wrapped := font.CreateWrappedText(maxWidth, scale, text)
    font.RenderWrapped(image, x, y, wrapped, colorScale, options)
}

func (font *Font) PrintWrapCenter(image *ebiten.Image, x float64, y float64, maxWidth float64, scale float64, colorScale ebiten.ColorScale, text string) {
    font.PrintWrap(image, x, y, maxWidth, scale, colorScale, FontOptions{Justify: FontJustifyCenter}, text)
    /*
    wrapped := font.CreateWrappedText(maxWidth, scale, text)
    font.RenderWrapped(image, x, y, wrapped, colorScale, FontOptions{Justify: FontJustifyCenter})
    */
}

type WrappedText struct {
    Lines []string
    TotalHeight float64
    MaxWidth float64
    Scale float64
}

func (text *WrappedText) Clear() {
    text.Lines = nil
}

// FIXME: remove colorScale argument
func (font *Font) RenderWrapped(image *ebiten.Image, x float64, y float64, wrapped WrappedText, colorScale ebiten.ColorScale, options FontOptions) {
    yPos := y
    for _, line := range wrapped.Lines {
        font.PrintOptions2(image, x, yPos, options, line)
        yPos += float64(font.Height()) + 1
    }
}

// FIXME: put this somewhere else
const NewLine = 0x14

// precompute an object that can be used to render a wrapped string
func (font *Font) CreateWrappedText(maxWidth float64, scale float64, text string) WrappedText {
    var lines []string
    var yPos float64 = 0

    text = string(bytes.ReplaceAll([]byte(text), []byte{'\n'}, []byte{NewLine}))

    textLines := bytes.Split([]byte(text), []byte{NewLine})

    for _, lineByte := range textLines {
        // line := strings.TrimSpace(string(lineByte))
        line := string(lineByte)
        for line != "" {
            show, rest := font.splitText(line, maxWidth, scale)

            // we were unable to split the text, just bail
            if show == "" {
                break
            }

            lines = append(lines, show)

            yPos += float64(font.Height()) * scale + 1

            line = rest
        }
    }

    return WrappedText{
        Lines: lines,
        TotalHeight: yPos,
        MaxWidth: maxWidth,
        Scale: scale,
    }
}
