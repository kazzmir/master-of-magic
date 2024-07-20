package main

import (
    "log"
    "os"
    "fmt"
    "sync"
    "math"
    "bytes"
    _ "embed"

    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed futura.ttf
var FuturaTTF []byte

const ScreenWidth = 1024
const ScreenHeight = 768

func LoadFont() (*text.GoTextFaceSource, error) {
    return text.NewGoTextFaceSource(bytes.NewReader(FuturaTTF))
}

type Viewer struct {
    Lbx *lbx.LbxFile
    Images []*ebiten.Image
    LoadImages sync.Once
    Scale float64
    CurrentImage int
    Font *text.GoTextFaceSource
}

func (viewer *Viewer) Update() error {
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendPressedKeys(keys)

    scaleAmount := 0.06

    for _, key := range keys {
        switch key {
            case ebiten.KeyUp:
                viewer.Scale *= 1 + scaleAmount
            case ebiten.KeyDown:
                viewer.Scale *= 1 - scaleAmount
                if viewer.Scale < 1 {
                    viewer.Scale = 1
                }
            case ebiten.KeySpace:
                if len(viewer.Images) > 0 {
                    bounds := viewer.Images[viewer.CurrentImage].Bounds()
                    viewer.Scale = 100.0 / math.Max(float64(bounds.Dx()), float64(bounds.Dy()))
                }
        }

    }

    keys = make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyLeft:
                viewer.CurrentImage -= 1
                if viewer.CurrentImage < 0 {
                    viewer.CurrentImage = len(viewer.Images) - 1
                }
            case ebiten.KeyRight:
                viewer.CurrentImage += 1
                if viewer.CurrentImage >= len(viewer.Images) {
                    viewer.CurrentImage = 0
                }
            case ebiten.KeyEscape, ebiten.KeyCapsLock:
                return ebiten.Termination
        }
    }

    viewer.LoadImages.Do(func(){
        rawImages, err := viewer.Lbx.ReadImages(0)
        if err != nil {
            log.Printf("Unable to load images: %v", err)
            return
        }
        var images []*ebiten.Image
        for _, rawImage := range rawImages {
            images = append(images, ebiten.NewImageFromImage(rawImage))
        }

        viewer.Images = images
    })

    return nil
}

func (viewer *Viewer) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return ScreenWidth, ScreenHeight
}

func (viewer *Viewer) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})

    face := &text.GoTextFace{Source: viewer.Font, Size: 15}

    op := &text.DrawOptions{}
    op.GeoM.Translate(1, 1)
    op.ColorScale.ScaleWithColor(color.White)
    text.Draw(screen, fmt.Sprintf("Image: %v/%v", viewer.CurrentImage+1, len(viewer.Images)), face, op)
    op.GeoM.Translate(0, 20)
    text.Draw(screen, fmt.Sprintf("Scale: %.2f", viewer.Scale), face, op)

    middleX := ScreenWidth / 2
    middleY := ScreenHeight / 2

    if len(viewer.Images) > 0 {
        var options ebiten.DrawImageOptions
        bounds := viewer.Images[viewer.CurrentImage].Bounds()
        options.GeoM.Translate(float64(-bounds.Dx()) / 2.0, float64(-bounds.Dy()) / 2.0)
        options.GeoM.Scale(viewer.Scale, viewer.Scale)
        options.GeoM.Translate(float64(middleX), float64(middleY))
        screen.DrawImage(viewer.Images[viewer.CurrentImage], &options)
    }
}

func MakeViewer(lbxFile *lbx.LbxFile) (*Viewer, error) {
    font, err := LoadFont()
    if err != nil {
        return nil, err
    }

    return &Viewer{
        Lbx: lbxFile,
        Scale: 5,
        Font: font,
    }, nil
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    if len(os.Args) < 2 {
        log.Printf("Give an lbx file to view")
        return
    }

    file := os.Args[1]

    ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
    ebiten.SetWindowTitle("lbx viewer")
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
