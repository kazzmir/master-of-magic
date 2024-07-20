package main

import (
    "log"
    "os"
    "sync"
    "math"

    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type Viewer struct {
    Lbx *lbx.LbxFile
    Images []*ebiten.Image
    LoadImages sync.Once
    Scale float64
    CurrentImage int
}

func (viewer *Viewer) Update() error {
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyUp:
                viewer.Scale *= 1.1
            case ebiten.KeyDown:
                viewer.Scale *= 0.9
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

func MakeViewer(lbxFile *lbx.LbxFile) Viewer {
    return Viewer{
        Lbx: lbxFile,
        Scale: 5,
    }
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

    viewer := MakeViewer(&lbxFile)

    err := ebiten.RunGame(&viewer)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
