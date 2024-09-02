package util

import (
    "image"
    "image/color"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

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
    Top image.Point
    Bottom image.Point
    Segments []Segment
}

/* draw 'source' onto 'screen' with a distortion effect
 * for each segment in 'distortion', make a quad from 'source' where the width is 'source width' / segments in the distortion
 * the distortion segments define where to place each corner of the quad
 */
func DrawDistortion(screen *ebiten.Image, page *ebiten.Image, source *ebiten.Image, distortion Distortion, options ebiten.DrawImageOptions){
    ax0, ay0 := options.GeoM.Apply(0, 0)
    ax1, ay1 := options.GeoM.Apply(float64(page.Bounds().Dx()), float64(page.Bounds().Dy()))
    subScreen := screen.SubImage(image.Rect(int(ax0), int(ay0), int(ax1), int(ay1))).(*ebiten.Image)

    // top left
    x1, y1 := options.GeoM.Apply(float64(distortion.Top.X), float64(distortion.Top.Y))
    // bottom left
    x4, y4 := options.GeoM.Apply(float64(distortion.Bottom.X), float64(distortion.Bottom.Y))

    segmentWidth := float32(source.Bounds().Dx()) / float32(len(distortion.Segments))

    for i := 0; i < len(distortion.Segments); i++ {
        segment := distortion.Segments[i]
        // top right
        x2, y2 := options.GeoM.Apply(float64(segment.Top.X), float64(segment.Top.Y))
        // bottom right
        x3, y3 := options.GeoM.Apply(float64(segment.Bottom.X), float64(segment.Bottom.Y))

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
