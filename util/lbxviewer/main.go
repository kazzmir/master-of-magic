package main

import (
    "log"
    "os"
    "fmt"
    "sync"
    "math"

    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/util/common"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type LbxImages struct {
    Images []*ebiten.Image
    Load sync.Once
    Loaded bool
    Lock sync.Mutex
}

func (loader *LbxImages) IsLoaded() bool {
    loader.Lock.Lock()
    defer loader.Lock.Unlock()
    return loader.Loaded
}

type ViewerState int

const (
    ViewStateTiles ViewerState = iota
    ViewStateImage
)

type Viewer struct {
    Lbx *lbx.LbxFile
    Images []*LbxImages
    Scale float64
    CurrentImage int
    CurrentTile int
    State ViewerState
    Font *text.GoTextFaceSource
    AnimationFrame int
    AnimationCount int
}

const TileWidth = 50
const TileHeight = 50

func tilesPerRow() int {
    width := ScreenWidth - 1
    return width / TileWidth
}

func (viewer *Viewer) Update() error {
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendPressedKeys(keys)

    const AnimationSpeed = 30

    // control_pressed := false

    scaleAmount := 0.06

    for _, key := range keys {
        switch key {
            case ebiten.KeyUp:
                if viewer.State == ViewStateImage {
                    viewer.Scale *= 1 + scaleAmount
                }
            case ebiten.KeyDown:
                if viewer.State == ViewStateImage {
                    viewer.Scale *= 1 - scaleAmount
                    if viewer.Scale < 1 {
                        viewer.Scale = 1
                    }
                }
                /*
            case ebiten.KeyControlLeft:
                control_pressed = true
                */
            case ebiten.KeySpace:
                if viewer.State == ViewStateImage {
                    if len(viewer.Images[viewer.CurrentTile].Images) > 0 {
                        bounds := viewer.Images[viewer.CurrentTile].Images[viewer.CurrentImage].Bounds()
                        viewer.Scale = 200.0 / math.Max(float64(bounds.Dx()), float64(bounds.Dy()))
                    }
                }
        }

    }

    keys = make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyEnter:
                if viewer.State == ViewStateTiles {
                    viewer.State = ViewStateImage
                } else {
                    viewer.State = ViewStateTiles
                }

            case ebiten.KeyLeft:
                switch viewer.State {
                    case ViewStateTiles:
                        if viewer.CurrentTile > 0 {
                            viewer.CurrentTile -= 1
                            viewer.CurrentImage = 0
                        }
                    case ViewStateImage:
                        viewer.CurrentImage -= 1
                        if viewer.CurrentImage < 0 {
                            viewer.CurrentImage = len(viewer.Images[viewer.CurrentTile].Images) - 1
                        }
                }

            case ebiten.KeyRight:
                switch viewer.State {
                    case ViewStateTiles:
                        if viewer.CurrentTile < len(viewer.Images) - 1 {
                            viewer.CurrentTile += 1
                            viewer.CurrentImage = 0
                        }
                    case ViewStateImage:
                        viewer.CurrentImage += 1
                        if viewer.CurrentImage >= len(viewer.Images[viewer.CurrentTile].Images) {
                            viewer.CurrentImage = 0
                        }
                }

            case ebiten.KeyUp:
                switch viewer.State {
                    case ViewStateTiles:
                        position := viewer.CurrentTile - tilesPerRow()
                        if position >= 0 {
                            viewer.CurrentTile = position
                            viewer.CurrentImage = 0
                        }
                }

            case ebiten.KeyDown:
                switch viewer.State {
                    case ViewStateTiles:
                        position := viewer.CurrentTile + tilesPerRow()
                        if position < len(viewer.Images) {
                            viewer.CurrentTile = position
                            viewer.CurrentImage = 0
                        }
                }

            case ebiten.KeyA:
                if viewer.AnimationFrame == -1 {
                    viewer.AnimationFrame = 0
                    viewer.AnimationCount = AnimationSpeed
                } else {
                    viewer.AnimationFrame = -1
                }

            case ebiten.KeyEscape, ebiten.KeyCapsLock:
                return ebiten.Termination
        }
    }

    if viewer.AnimationFrame != -1 {
        if viewer.AnimationCount > 0 {
            viewer.AnimationCount -= 1
        } else {
            viewer.AnimationFrame += 1
            if viewer.AnimationFrame >= len(viewer.Images[viewer.CurrentTile].Images) {
                viewer.AnimationFrame = 0
            }
            viewer.AnimationCount = AnimationSpeed
        }
    }

    return nil
}

func (viewer *Viewer) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return ScreenWidth, ScreenHeight
}

func aspectScale(width, height, maxWidth, maxHeight int) (float64, float64) {
    scaleX := float64(maxWidth) / float64(width)
    scaleY := float64(maxHeight) / float64(height)
    if scaleX < scaleY {
        return scaleX, scaleX
    }
    return scaleY, scaleY
}

func (viewer *Viewer) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})

    face := &text.GoTextFace{Source: viewer.Font, Size: 15}

    op := &text.DrawOptions{}
    op.GeoM.Translate(1, 1)
    op.ColorScale.ScaleWithColor(color.White)
    text.Draw(screen, fmt.Sprintf("Lbx entry: %v/%v", viewer.CurrentTile, viewer.Lbx.TotalEntries() - 1), face, op)
    op.GeoM.Translate(1, 20)
    if viewer.AnimationFrame != -1 {
        text.Draw(screen, fmt.Sprintf("Animation : %v/%v", viewer.AnimationFrame+1, len(viewer.Images[viewer.CurrentTile].Images)), face, op)
    } else {
        text.Draw(screen, fmt.Sprintf("Image: %v/%v", viewer.CurrentImage+1, len(viewer.Images[viewer.CurrentTile].Images)), face, op)
    }
    op.GeoM.Translate(0, 20)
    text.Draw(screen, fmt.Sprintf("Scale: %.2f", viewer.Scale), face, op)
    op.GeoM.Translate(0, 20)
    if len(viewer.Images[viewer.CurrentTile].Images) > 0 {
        img := viewer.Images[viewer.CurrentTile].Images[viewer.CurrentImage]
        text.Draw(screen, fmt.Sprintf("Dimensions: %v x %v", img.Bounds().Dx(), img.Bounds().Dy()), face, op)
    }

    startX := 1
    startY := 100

    x := startX
    y := startY

    // FIXME: handle the case when there are more images than can fit on the screen
    for i, image := range viewer.Images {
        if image.IsLoaded() && len(image.Images) > 0 {
            var options ebiten.DrawImageOptions

            draw := image.Images[0]

            scaleX, scaleY := aspectScale(draw.Bounds().Dx(), draw.Bounds().Dy(), TileWidth, TileHeight)

            options.GeoM.Scale(scaleX, scaleY)
            options.GeoM.Translate(float64(x), float64(y))
            screen.DrawImage(draw, &options)
            /*
            text.Draw(screen, fmt.Sprintf("%v", i), face, &text.DrawOptions{
                GeoM: ebiten.GeoM.Translate(float64(x), float64(y)),
            })
            */
        }

        if i == viewer.CurrentTile {
            vector.StrokeRect(screen, float32(x), float32(y), float32(TileWidth), float32(TileHeight), 1.5, color.White, true)
        }

        x += TileWidth
        if x + TileWidth >= ScreenWidth {
            x = 1
            y += TileHeight
        }
    }

    if viewer.State == ViewStateImage {
        if len(viewer.Images[viewer.CurrentTile].Images) > 0 {
            vector.DrawFilledRect(screen, 0, float32(startY), float32(ScreenWidth), float32(ScreenHeight - startY), color.RGBA{0, 0, 0, 64}, false)
            middleX := ScreenWidth / 2
            middleY := ScreenHeight / 2

            var options ebiten.DrawImageOptions
            useImage := viewer.Images[viewer.CurrentTile].Images[viewer.CurrentImage]
            if viewer.AnimationFrame != -1 && viewer.AnimationFrame < len(viewer.Images[viewer.CurrentTile].Images) {
                useImage = viewer.Images[viewer.CurrentTile].Images[viewer.AnimationFrame]
            }
            bounds := useImage.Bounds()
            options.GeoM.Translate(float64(-bounds.Dx()) / 2.0, float64(-bounds.Dy()) / 2.0)
            options.GeoM.Scale(viewer.Scale, viewer.Scale)
            options.GeoM.Translate(float64(middleX), float64(middleY))
            screen.DrawImage(useImage, &options)
        }
    }
}

func MakeViewer(lbxFile *lbx.LbxFile) (*Viewer, error) {
    font, err := common.LoadFont()
    if err != nil {
        return nil, err
    }

    viewer := &Viewer{
        Lbx: lbxFile,
        Scale: 1,
        Font: font,
        CurrentImage: 0,
        AnimationFrame: -1,
        AnimationCount: 0,
        State: ViewStateTiles,
    }

    for i := 0; i < lbxFile.TotalEntries(); i++ {
        loader := &LbxImages{}
        viewer.Images = append(viewer.Images, loader)

        go func(){
            loader.Load.Do(func(){
                rawImages, err := lbxFile.ReadImages(i)
                if err != nil {
                    log.Printf("Unable to load images: %v", err)
                    return
                }
                var images []*ebiten.Image
                for _, rawImage := range rawImages {
                    images = append(images, ebiten.NewImageFromImage(rawImage))
                }
                loader.Images = images

                loader.Lock.Lock()
                loader.Loaded = true
                loader.Lock.Unlock()
            })
        }()
    }

    return viewer, nil
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
