package main

import (
    "log"
    "os"
    // "fmt"
    // "sync"
    // "math"
    // "bytes"
    // _ "embed"

    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"

    "github.com/hajimehoshi/ebiten/v2"
    // "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    // "github.com/hajimehoshi/ebiten/v2/text/v2"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type Viewer struct {
    Lbx *lbx.LbxFile
    Fonts []*lbx.Font
}

func MakeViewer(lbxFile *lbx.LbxFile) (*Viewer, error) {
    fonts, err := lbxFile.ReadFonts(0)
    if err != nil {
        return nil, err
    }

    optimized := font.MakeOptimizedFont(fonts[0])
    _ = optimized

    return &Viewer{
        Lbx: lbxFile,
        Fonts: fonts,
    }, nil
}

func (viewer *Viewer) Update() error {
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyEscape, ebiten.KeyCapsLock:
                return ebiten.Termination
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
    options.GeoM.Scale(4, 4)
    options.GeoM.Translate(1, 1)

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
