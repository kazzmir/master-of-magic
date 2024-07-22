package main

import (
    "log"
    "os"
    _ "fmt"
    // "sync"
    // "math"
    // "bytes"
    // _ "embed"

    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    // "github.com/hajimehoshi/ebiten/v2/text/v2"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type Viewer struct {
    Lbx *lbx.LbxFile
    FontIndex int
    Fonts []*lbx.Font
    Optimized *font.Font
    Scale float64
}

func MakeViewer(lbxFile *lbx.LbxFile) (*Viewer, error) {
    fonts, err := lbxFile.ReadFonts(0)
    if err != nil {
        return nil, err
    }

    optimized := font.MakeOptimizedFont(fonts[0])

    /*
    pGlyph := fonts[0].GlyphForRune('p')
    data := pGlyph.Data
    log.Printf("Glyph data for 'p' width=%v height=%v", pGlyph.Width, pGlyph.Height)
    for _, v := range data {
        fmt.Printf("0x%x ", v)
    }
    fmt.Println()
    pGlyph.MakeImage()
    */

    return &Viewer{
        Lbx: lbxFile,
        Optimized: optimized,
        FontIndex: 0,
        Fonts: fonts,
        Scale: 4,
    }, nil
}

func (viewer *Viewer) Update() error {
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyUp:
                viewer.Scale *= 1.05
            case ebiten.KeyDown:
                viewer.Scale *= 0.95
                if viewer.Scale < 1 {
                    viewer.Scale = 1
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
                viewer.FontIndex -= 1
                if viewer.FontIndex < 0 {
                    viewer.FontIndex = len(viewer.Fonts) - 1
                }

                viewer.Optimized = font.MakeOptimizedFont(viewer.Fonts[viewer.FontIndex])
            case ebiten.KeyRight:
                viewer.FontIndex = (viewer.FontIndex + 1) % len(viewer.Fonts)
                viewer.Optimized = font.MakeOptimizedFont(viewer.Fonts[viewer.FontIndex])
        }
    }

    return nil
}

func (viewer *Viewer) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return ScreenWidth, ScreenHeight
}

func (viewer *Viewer) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})

    // vector.DrawFilledRect(screen, 90, 90, 100, 100, &color.RGBA{R: 0xff, A: 0xff}, true)

    var options ebiten.DrawImageOptions
    options.GeoM.Scale(viewer.Scale, viewer.Scale)
    options.GeoM.Translate(50, 50)

    screen.DrawImage(viewer.Optimized.Image, &options)

    vector.StrokeRect(screen, 50, 500, float32(float64(viewer.Optimized.GlyphWidth) * 20 * viewer.Scale), float32(float64(viewer.Optimized.GlyphHeight) * viewer.Scale), 1, &color.RGBA{R: 0xff, A: 0xff}, true)
    viewer.Optimized.Print(screen, 50, 500, viewer.Scale, "Hello, potato!")

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

    if len(os.Args) < 2 {
        log.Printf("Give an lbx font file to view")
        return
    }

    file := os.Args[1]

    ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
    ebiten.SetWindowTitle("lbx font viewer")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    var lbxFile lbx.LbxFile

    func(){
        open, err := os.Open(file)
        if err != nil {
            log.Printf("Error: %v", err)
            return
        }
        defer open.Close()
        lbxFile, err = lbx.ReadLbx(open)
        if err != nil {
            log.Printf("Error: %v\n", err)
            return
        }
        log.Printf("Loaded lbx file: %v\n", file)
    }()

    viewer, err := MakeViewer(&lbxFile)
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }

    err = ebiten.RunGame(viewer)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
