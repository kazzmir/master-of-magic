package main

import (
    "os"
    "fmt"
    "image"

    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type Viewer struct {
    Images []image.Image
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
}

func display(lbxData lbx.LbxFile) error {
    images, err := lbxData.ReadTerrainImages(0)
    if err != nil {
        return err
    }
    
    ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
    ebiten.SetWindowTitle("terrain viewer")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    viewer := &Viewer{
        Images: images,
    }
    
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
