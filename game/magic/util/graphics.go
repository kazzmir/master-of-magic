package util

import (
    "log"
    "image"
    "image/color"
    "math"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/colorconv"
    "github.com/kazzmir/master-of-magic/game/magic/shaders"
    "github.com/kazzmir/master-of-magic/game/magic/scale"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/colorm"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

type AlphaFadeFunc func() float32

func PremultiplyAlpha(c color.RGBA) color.RGBA {
    a := float64(c.A) / 255.0
    return color.RGBA{
        R: uint8(float64(c.R) * a),
        G: uint8(float64(c.G) * a),
        B: uint8(float64(c.B) * a),
        A: c.A,
    }
}

func DrawRect(screen *ebiten.Image, rect image.Rectangle, color_ color.Color){
    vector.StrokeRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), 1, color_, false)
}

func ImageRect(x int, y int, img *ebiten.Image) image.Rectangle {
    return image.Rect(x, y, x + img.Bounds().Dx(), y + img.Bounds().Dy())
}

func toFloatArray(color color.Color) []float32 {
    r, g, b, a := color.RGBA()
    var max float32 = 65535.0
    return []float32{float32(r) / max, float32(g) / max, float32(b) / max, float32(a) / max}
}

func DrawOutline(screen *ebiten.Image, imageCache *ImageCache, pic *ebiten.Image, geom ebiten.GeoM, scale ebiten.ColorScale, time uint64, baseColor color.Color) {
    color1 := Lighten(baseColor, -20)
    color2 := Lighten(baseColor, 10)
    color3 := Lighten(baseColor, 50)

    shader, err := imageCache.GetShader(shaders.ShaderEdgeGlow)

    if err != nil {
        log.Printf("Error: unable to get edge glow shader: %v", err)
        return
    }

    var options ebiten.DrawRectShaderOptions
    // FIXME: this geom doesn't take the screen scale into account. The image passed in is already scaled
    // so if we simply apply the screen scale here to the geom then the resulting shader will be much too large
    options.GeoM = geom
    options.Images[0] = pic
    options.Uniforms = make(map[string]interface{})
    options.Uniforms["Color1"] = toFloatArray(color1)
    options.Uniforms["Color2"] = toFloatArray(color2)
    options.Uniforms["Color3"] = toFloatArray(color3)
    options.Uniforms["Time"] = float32(math.Abs(float64(time)))
    options.ColorScale = scale
    screen.DrawRectShader(pic.Bounds().Dx(), pic.Bounds().Dy(), shader, &options)
}

func drawDistortedImage(destination *ebiten.Image, source *ebiten.Image, vertices [4]ebiten.Vertex){
    for i := 0; i < 4; i++ {
        vertices[i].ColorA = 1
        vertices[i].ColorR = 1
        vertices[i].ColorG = 1
        vertices[i].ColorB = 1
    }

    // the order probably doesn't matter all that much but we draw two triangles:
    // 1: top-left, top-right, bottom right
    // 2: bottom right, bottom left, top left
    destination.DrawTriangles(vertices[:], []uint16{0, 1, 2, 2, 3, 0}, source, nil)
}

type Segment struct {
    Top image.Point
    Bottom image.Point
}

type Distortion struct {
    Segments []Segment
    Top image.Point
    Bottom image.Point
}

/* draw 'source' onto 'screen' with a distortion effect
 * for each segment in 'distortion', make a quad from 'source' where the width is 'source width' / segments in the distortion
 * the distortion segments define where to place each corner of the quad
 */
func DrawDistortion(screen *ebiten.Image, page *ebiten.Image, source *ebiten.Image, distortion Distortion, geom ebiten.GeoM){
    ax0, ay0 := geom.Apply(0, 0)
    ax1, ay1 := geom.Apply(float64(page.Bounds().Dx()), float64(page.Bounds().Dy()))
    subScreen := screen.SubImage(image.Rect(int(ax0), int(ay0), int(ax1), int(ay1))).(*ebiten.Image)

    // top left
    x1, y1 := geom.Apply(float64(distortion.Top.X), float64(distortion.Top.Y))
    // bottom left
    x4, y4 := geom.Apply(float64(distortion.Bottom.X), float64(distortion.Bottom.Y))

    segmentWidth := float32(source.Bounds().Dx()) / float32(len(distortion.Segments))

    for i := 0; i < len(distortion.Segments); i++ {
        segment := distortion.Segments[i]
        // top right
        x2, y2 := geom.Apply(float64(segment.Top.X), float64(segment.Top.Y))
        // bottom right
        x3, y3 := geom.Apply(float64(segment.Bottom.X), float64(segment.Bottom.Y))

        sx := float32(0)
        sy := float32(source.Bounds().Dy())

        // 'sx + segmentWidth * i' defines the portion of the source image that we want to draw
        drawDistortedImage(subScreen, source, [4]ebiten.Vertex{
            ebiten.Vertex{
                DstX: float32(x1),
                DstY: float32(y1),
                SrcX: sx + segmentWidth * float32(i),
                SrcY: 0,
            },
            ebiten.Vertex{
                DstX: float32(x2),
                DstY: float32(y2),
                SrcX: sx + segmentWidth * float32(i+1),
                SrcY: 0,
            },
            ebiten.Vertex{
                DstX: float32(x3),
                DstY: float32(y3),
                SrcX: sx + segmentWidth * float32(i+1),
                SrcY: sy,
            },
            ebiten.Vertex{
                DstX: float32(x4),
                DstY: float32(y4),
                SrcX: sx + segmentWidth * float32(i),
                SrcY: sy,
            },
        })

        // old left becomes right
        x1 = x2
        y1 = y2
        x4 = x3
        y4 = y3
    }
}

func Lighten(c color.Color, amount float64) color.Color {
    var change colorm.ColorM
    change.ChangeHSV(0, 1 - amount/100, 1 + amount/100)
    return change.Apply(c)
}

func RotateHue(c color.Color, radian float64) color.Color {
    var rotate colorm.ColorM
    rotate.ChangeHSV(radian, 1, 1)
    return rotate.Apply(c)
}

func ToRGBA(c color.Color) color.RGBA {
    r, g, b, a := c.RGBA()
    return color.RGBA{
        R: uint8(r >> 8),
        G: uint8(g >> 8),
        B: uint8(b >> 8),
        A: uint8(a >> 8),
    }
}

// a more sane version of lighten
func Lighten2(c color.RGBA, amount float64) color.Color {
    h, s, v := colorconv.ColorToHSV(c)
    v += amount/100
    if v > 1 {
        v = 1
    }
    out, err := colorconv.HSVToColor(h, s, v)
    if err != nil {
        log.Printf("Error in lighten: %v", err)
        return c
    }
    return out
}

// just for convenience
func Darken2(c color.RGBA, amount float64) color.Color {
    return Lighten2(c, -amount)
}

func MakeFadeIn(time uint64, counter *uint64) AlphaFadeFunc {
    start := *counter
    return func() float32 {
        diff := *counter - start
        if diff > time {
            return 1.0
        }

        return float32(diff) / float32(time)
    }
}

func MakeFadeOut(time uint64, counter *uint64) AlphaFadeFunc {
    start := *counter
    return func() float32 {
        diff := *counter - start
        if diff > time {
            return 0.0
        }

        return 1.0 - (float32(diff) / float32(time))
    }
}


/* create an animation by rotating the colors in a palette for a given lbx/index pair.
 * all the colors between indexLow and indexHigh will be rotated once in the animation
 */
func MakePaletteRotateAnimation(lbxFile *lbx.LbxFile, index int, rotateIndexLow int, rotateIndexHigh int) *Animation {
    basePalette, err := lbxFile.GetPalette(index)
    if err != nil {
        return nil
    }

    var images []*ebiten.Image

    for range rotateIndexHigh - rotateIndexLow {

        /*
        rotatedPalette := make(color.Palette, len(basePalette))
        copy(rotatedPalette, basePalette)

        for c := 245; c <= 254; c++ {
            v := math.Sin(float64(i + c) / 3) * 60
            rotatedPalette[c] = util.Lighten(basePalette[c], v)
        }
        */

        RotateSlice(basePalette[rotateIndexLow:rotateIndexHigh], false)

        newImages, err := lbxFile.ReadImagesWithPalette(index, basePalette, true)
        if err != nil || len(newImages) != 1 {
            return nil
        }

        images = append(images, ebiten.NewImageFromImage(newImages[0]))
    }

    return MakeAnimation(images, true)
}

func DrawTextCursor(screen *ebiten.Image, source *ebiten.Image, cursorX float64, y float64, counter uint64) {
    width := float64(4)
    height := float64(8)

    yOffset := float64((counter/3) % 16) - height

    vertices := [4]ebiten.Vertex{
        ebiten.Vertex{
            DstX: float32(scale.Scale(cursorX)),
            DstY: float32(scale.Scale(y - yOffset)),
            SrcX: 0,
            SrcY: 0,
            ColorA: 1,
            ColorB: 1 ,
            ColorG: 1,
            ColorR: 1,
        },
        ebiten.Vertex{
            DstX: float32(scale.Scale(cursorX + width)),
            DstY: float32(scale.Scale(y - yOffset)),
            SrcX: 0,
            SrcY: 0,
            ColorA: 1,
            ColorB: 1 ,
            ColorG: 1,
            ColorR: 1,
        },
        ebiten.Vertex{
            DstX: float32(scale.Scale(cursorX + width)),
            DstY: float32(scale.Scale(y + height - yOffset)),
            SrcX: 0,
            SrcY: 0,
            ColorA: 0.1,
            ColorB: 1 ,
            ColorG: 1,
            ColorR: 1,
        },
        ebiten.Vertex{
            DstX: float32(scale.Scale(cursorX)),
            DstY: float32(scale.Scale(y + height - yOffset)),
            SrcX: 0,
            SrcY: 0,
            ColorA: 0.1,
            ColorB: 1 ,
            ColorG: 1,
            ColorR: 1,
        },
    }

    cursorArea := screen.SubImage(scale.ScaleRect(image.Rect(int(cursorX), int(y), int(cursorX + width), int(y + height)))).(*ebiten.Image)
    cursorArea.DrawTriangles(vertices[:], []uint16{0, 1, 2, 2, 3, 0}, source, nil)
}

func ClonePalette(p color.Palette) color.Palette {
    newPalette := make(color.Palette, len(p))
    copy(newPalette, p)
    return newPalette
}
