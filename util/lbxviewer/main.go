package main

import (
    "log"
    "os"
    "sync"

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
}

func (viewer *Viewer) Update() error {
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
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


    var options ebiten.DrawImageOptions
    screen.DrawImage(viewer.Images[0], &options)
}

func MakeViewer(lbxFile *lbx.LbxFile) Viewer {
    return Viewer{
        Lbx: lbxFile,
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
