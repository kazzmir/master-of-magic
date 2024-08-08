package main

import (
    "os"
    "fmt"
    "bytes"
    "image"
    "image/color"
    _ "embed"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed futura.ttf
var FuturaTTF []byte

const ScreenWidth = 1024
const ScreenHeight = 768

func LoadFont() (*text.GoTextFaceSource, error) {
    return text.NewGoTextFaceSource(bytes.NewReader(FuturaTTF))
}

type ImageGPU struct {
    Raw image.Image
    GPU *ebiten.Image
}

type Viewer struct {
    Images []ImageGPU
    Font *text.GoTextFaceSource
    Choice int
    Counter uint64
    StartingRow int
}

func MakeViewer(images []image.Image) *Viewer {
    var use []ImageGPU

    for _, img := range images {
        use = append(use, ImageGPU{
            Raw: img,
            GPU: nil,
        })
    }

    font, err := LoadFont()
    if err != nil {
        fmt.Printf("Could not load font: %v\n", err)
    }

    return &Viewer{
        Images: use,
        Font: font,
        Choice: 0,
        StartingRow: 0,
    }
}

func (viewer *Viewer) TilesPerRow() int {
    // hack: why is +1 needed?
    return (ScreenWidth - 3) / (viewer.Images[0].Raw.Bounds().Dx() + 5) + 1
}

func (viewer *Viewer) TilesPerColumn() int {
    return (ScreenHeight - 110) / (viewer.Images[0].Raw.Bounds().Dy() + 5)
}

func (viewer *Viewer) Update() error {
    viewer.Counter += 1
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendPressedKeys(keys)

    moveRight := false
    moveLeft := false
    moveUp := false
    moveDown := false

    leftShift := inpututil.KeyPressDuration(ebiten.KeyShiftLeft) > 0

    if viewer.Counter % 3 == 0 && leftShift{

        for _, key := range keys {
            switch key {
                case ebiten.KeyRight: moveRight = true
                case ebiten.KeyLeft: moveLeft = true
                case ebiten.KeyUp: moveUp = true
                case ebiten.KeyDown: moveDown = true
            }
        }
    }

    keys = make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyRight: moveRight = true
            case ebiten.KeyLeft: moveLeft = true
            case ebiten.KeyUp: moveUp = true
            case ebiten.KeyDown: moveDown = true
            case ebiten.KeyEscape, ebiten.KeyCapsLock:
                return ebiten.Termination
        }
    }

    if moveRight {
        viewer.Choice += 1
        if viewer.Choice >= len(viewer.Images) {
            viewer.Choice = len(viewer.Images) - 1
        }
    }

    if moveLeft {
        viewer.Choice -= 1
        if viewer.Choice < 0 {
            viewer.Choice = 0
        }
    }

    if moveUp {
        viewer.Choice -= viewer.TilesPerRow()
        if viewer.Choice < 0 {
            viewer.Choice = 0
        }
    }

    if moveDown {
        viewer.Choice += viewer.TilesPerRow()
        if viewer.Choice >= len(viewer.Images) {
            viewer.Choice = len(viewer.Images) - 1
        }
    }

    for viewer.Choice < viewer.StartingRow * viewer.TilesPerRow() {
        viewer.StartingRow -= 1
    }

    for viewer.Choice >= (viewer.StartingRow + viewer.TilesPerColumn()) * viewer.TilesPerRow() {
        viewer.StartingRow += 1
    }

    return nil
}

func (viewer *Viewer) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return ScreenWidth, ScreenHeight
}

func (viewer *Viewer) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{0x10, 0x10, 0x10, 0xff})

    face := &text.GoTextFace{Source: viewer.Font, Size: 15}
    op := &text.DrawOptions{}
    op.GeoM.Translate(1, 1)
    op.ColorScale.ScaleWithColor(color.White)
    text.Draw(screen, fmt.Sprintf("Terrain entry: %v/%v", viewer.Choice, len(viewer.Images)-1), face, op)

    var options ebiten.DrawImageOptions
    x := float64(3)
    y := float64(110)

    options.GeoM.Scale(4, 4)
    options.GeoM.Translate(ScreenWidth/2, 10)
    if viewer.Images[viewer.Choice].GPU != nil {
        screen.DrawImage(viewer.Images[viewer.Choice].GPU, &options)
    }

    startPosition := viewer.StartingRow * viewer.TilesPerRow()

    for i := startPosition; i < len(viewer.Images); i++ {
        img := viewer.Images[i]
        if img.GPU == nil {
            img.GPU = ebiten.NewImageFromImage(img.Raw)
            viewer.Images[i] = img
        }

        options.GeoM.Reset()
        options.GeoM.Translate(x, y)
        screen.DrawImage(img.GPU, &options)

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
    /*
    images, err := lbxData.ReadTerrainImages(0)
    if err != nil {
        return err
    }
    */
    data, err := terrain.ReadTerrainData(&lbxData)
    if err != nil {
        return err
    }
    
    ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
    ebiten.SetWindowTitle("terrain viewer")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    viewer := MakeViewer(data.Images)
    
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
