package main

import (
    "os"
    "fmt"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type ImageGPU struct {
    Raw image.Image
    GPU *ebiten.Image
}

type Viewer struct {
    Images []ImageGPU
    Choice int
}

func MakeViewer(images []image.Image) *Viewer {
    var use []ImageGPU

    for _, img := range images {
        use = append(use, ImageGPU{
            Raw: img,
            GPU: nil,
        })
    }

    return &Viewer{
        Images: use,
        Choice: 0,
    }
}

func (viewer *Viewer) TilesPerRow() int {
    // hack: why is +1 needed?
    return (ScreenWidth - 3) / (viewer.Images[0].Raw.Bounds().Dx() + 5) + 1
}

func (viewer *Viewer) Update() error {
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyRight:
                viewer.Choice += 1
                if viewer.Choice >= len(viewer.Images) {
                    viewer.Choice = len(viewer.Images) - 1
                }
            case ebiten.KeyLeft:
                viewer.Choice -= 1
                if viewer.Choice < 0 {
                    viewer.Choice = 0
                }
            case ebiten.KeyUp:
                viewer.Choice -= viewer.TilesPerRow()
                if viewer.Choice < 0 {
                    viewer.Choice = 0
                }
            case ebiten.KeyDown:
                viewer.Choice += viewer.TilesPerRow()
                if viewer.Choice >= len(viewer.Images) {
                    viewer.Choice = len(viewer.Images) - 1
                }
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
    screen.Fill(color.RGBA{0x10, 0x10, 0x10, 0xff})

    var options ebiten.DrawImageOptions
    x := float64(3)
    y := float64(100)

    for i, img := range viewer.Images {
        if img.GPU == nil {
            img.GPU = ebiten.NewImageFromImage(img.Raw)
            viewer.Images[i] = img
        }

        screen.DrawImage(img.GPU, &options)
        options.GeoM.Reset()
        options.GeoM.Translate(x, y)

        if i == viewer.Choice {
            width := float32(img.Raw.Bounds().Dx())
            height := float32(img.Raw.Bounds().Dy())
            vector.StrokeRect(screen, float32(x-1), float32(y-1), width+2, height+2, 1.5, color.White, true)
        }

        x += float64(img.Raw.Bounds().Dx()) + 5
        if x >= float64(ScreenWidth - img.Raw.Bounds().Dx()) {
            x = 3
            y += float64(img.Raw.Bounds().Dy()) + 5
        }

        if y >= float64(ScreenHeight) {
            break
        }
    }
}

func display(lbxData lbx.LbxFile) error {
    images, err := lbxData.ReadTerrainImages(0)
    if err != nil {
        return err
    }
    
    ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
    ebiten.SetWindowTitle("terrain viewer")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    viewer := MakeViewer(images)
    
    err = ebiten.RunGame(viewer)

    return err
}

func main(){
    if len(os.Args) < 2 {
        fmt.Printf("Give an lbx file to read terrain data from\n")
        return
    }

    path := os.Args[1]
    file, err := os.Open(path)
    if err != nil {
        fmt.Printf("Error opening file: %v\n", err)
        return
    }

    lbxData, err := lbx.ReadLbx(file)
    if err != nil {
        fmt.Printf("Error reading lbx: %v\n", err)
        return
    }

    file.Close()

    err = display(lbxData)
    if err != nil {
        fmt.Printf("Error displaying lbx: %v\n", err)
        return
    }
}
