package main

import (
    "log"
    "os"
    "fmt"
    "sync"
    "math"
    "slices"

    "image/color"
    "image"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/util/common"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type LbxData struct {
    Lbx *lbx.LbxFile
    Name string
}

type CacheData struct {
    Image *ebiten.Image
    Time uint64
}

// LRU cache
type ImageCache struct {
    Images map[string]CacheData
    MaxSize int
}

func (cache *ImageCache) GetImage(key string, raw *image.Paletted, time uint64) *ebiten.Image {
    data, ok := cache.Images[key]
    if !ok {
        data.Image = ebiten.NewImageFromImage(raw)
    }

    data.Time = time
    cache.Images[key] = data
    return data.Image
}

func (cache *ImageCache) Cleanup(){
    maxExtra := 5

    if len(cache.Images) > cache.MaxSize + maxExtra {
        // var oldestKey string
        // var oldestTime uint64

        type removeKey struct {
            Time uint64
            Key string
        }

        var toRemove []removeKey

        for key, data := range cache.Images {
            /*
            if oldestTime == 0 || data.Time < oldestTime {
                oldestTime = data.Time
                oldestKey = key
            }
            */

            added := false

            if len(toRemove) > 0 && data.Time < toRemove[len(toRemove)-1].Time {
                for i := 0; i < len(toRemove); i++ {
                    if data.Time < toRemove[i].Time {
                        toRemove = slices.Insert(toRemove, i, removeKey{Time: data.Time, Key: key})
                        added = true
                        break
                    }
                }
            }

            if !added && len(toRemove) < maxExtra {
                toRemove = append(toRemove, removeKey{Time: data.Time, Key: key})
            }

            if len(toRemove) > maxExtra {
                toRemove = toRemove[:maxExtra]
            }
        }

        for _, key := range toRemove {
            // log.Printf("Cache eviction: %v", key.Key)
            delete(cache.Images, key.Key)
        }
    }
}

func MakeImageCache(size int) ImageCache {
    return ImageCache{
        Images: make(map[string]CacheData),
        MaxSize: size,
    }
}

type LbxImages struct {
    Keys []string
    Images []*image.Paletted
    LbxData *LbxData
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
    Data []*LbxData
    StartingRow int
    Indexes map[string]int
    Images []*LbxImages
    Scale float64
    CurrentImage int
    CurrentTile int
    State ViewerState
    Font *text.GoTextFaceSource
    AnimationFrame int
    AnimationCount int
    ShiftCount int
    Time uint64
    ImageCache ImageCache
}

const TileWidth = 50
const TileHeight = 50

func tilesPerRow() int {
    width := ScreenWidth - 1
    return width / TileWidth
}

func (viewer *Viewer) Update() error {
    viewer.ImageCache.Cleanup()
    viewer.Time += 1
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendPressedKeys(keys)

    const AnimationSpeed = 30

    shift_pressed := false

    scaleAmount := 0.06

    press_up := false
    press_down := false
    press_left := false
    press_right := false

    shiftSpeed := 4

    for _, key := range keys {
        switch key {
            case ebiten.KeyUp:
                if viewer.State == ViewStateImage {
                    viewer.Scale *= 1 + scaleAmount
                }
                if viewer.State == ViewStateTiles && viewer.ShiftCount % shiftSpeed == 1 {
                    press_up = true
                }
            case ebiten.KeyDown:
                if viewer.State == ViewStateImage {
                    viewer.Scale *= 1 - scaleAmount
                    if viewer.Scale < 1 {
                        viewer.Scale = 1
                    }
                }

                if viewer.State == ViewStateTiles && viewer.ShiftCount % shiftSpeed == 1 {
                    press_down = true
                }
            case ebiten.KeyLeft:
                if viewer.State == ViewStateTiles && viewer.ShiftCount % shiftSpeed == 1 {
                    press_left = true
                }
            case ebiten.KeyRight:
                if viewer.State == ViewStateTiles && viewer.ShiftCount % shiftSpeed == 1 {
                    press_right = true
                }
            case ebiten.KeyShiftLeft:
                shift_pressed = true
            case ebiten.KeySpace:
                if viewer.State == ViewStateImage {
                    if len(viewer.Images[viewer.CurrentTile].Images) > 0 {
                        bounds := viewer.Images[viewer.CurrentTile].Images[viewer.CurrentImage].Bounds()
                        viewer.Scale = 200.0 / math.Max(float64(bounds.Dx()), float64(bounds.Dy()))
                    }
                }
        }
    }

    if shift_pressed {
        viewer.ShiftCount += 1
    } else {
        viewer.ShiftCount = 0
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
                        press_left = true
                    case ViewStateImage:
                        viewer.CurrentImage -= 1
                        if viewer.CurrentImage < 0 {
                            viewer.CurrentImage = len(viewer.Images[viewer.CurrentTile].Images) - 1
                        }
                }

            case ebiten.KeyRight:
                switch viewer.State {
                    case ViewStateTiles:
                        press_right = true
                    case ViewStateImage:
                        viewer.CurrentImage += 1
                        if viewer.CurrentImage >= len(viewer.Images[viewer.CurrentTile].Images) {
                            viewer.CurrentImage = 0
                        }
                }

            case ebiten.KeyUp:
                switch viewer.State {
                    case ViewStateTiles:
                        press_up = true
                }

            case ebiten.KeyDown:
                switch viewer.State {
                    case ViewStateTiles:
                        press_down = true
                }

            case ebiten.KeyA:
                if viewer.AnimationFrame == -1 {
                    viewer.AnimationFrame = 0
                    viewer.AnimationCount = AnimationSpeed
                } else {
                    viewer.AnimationFrame = -1
                }

            case ebiten.KeyEscape, ebiten.KeyCapsLock:
                switch viewer.State {
                    case ViewStateTiles: return ebiten.Termination
                    case ViewStateImage: viewer.State = ViewStateTiles
                }
        }
    }

    if press_left {
        if viewer.CurrentTile > 0 {
            viewer.CurrentTile -= 1
            viewer.CurrentImage = 0
        }
    }

    if press_right {
        if viewer.CurrentTile < len(viewer.Images) - 1 {
            viewer.CurrentTile += 1
            viewer.CurrentImage = 0
        }
    }

    if press_up {
        position := viewer.CurrentTile - tilesPerRow()
        if position >= 0 {
            viewer.CurrentTile = position
            viewer.CurrentImage = 0
        }
    }

    if press_down {
        position := viewer.CurrentTile + tilesPerRow()
        if position < len(viewer.Images) {
            viewer.CurrentTile = position
            viewer.CurrentImage = 0
        }
    }

    tilesPerRow := (ScreenWidth - 1) / TileWidth
    tilesPerColumn := (ScreenHeight - 100) / TileHeight - 1

    if viewer.CurrentTile > (viewer.StartingRow + tilesPerColumn) * tilesPerRow {
        viewer.StartingRow = viewer.CurrentTile / tilesPerRow - tilesPerColumn
    }

    if viewer.CurrentTile < viewer.StartingRow * tilesPerRow {
        viewer.StartingRow = viewer.CurrentTile / tilesPerRow
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
    text.Draw(screen, fmt.Sprintf("%v entry: %v/%v", viewer.Images[viewer.CurrentTile].LbxData.Name, viewer.CurrentTile - viewer.Indexes[viewer.Images[viewer.CurrentTile].LbxData.Name], viewer.Images[viewer.CurrentTile].LbxData.Lbx.TotalEntries() - 1), face, op)
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

    for i, image := range viewer.Images {
        if i < viewer.StartingRow * tilesPerRow() {
            continue
        }
        if image.IsLoaded() && len(image.Images) > 0 {
            var options ebiten.DrawImageOptions

            draw := viewer.ImageCache.GetImage(image.Keys[0], image.Images[0], viewer.Time)

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

            if y >= ScreenHeight {
                break
            }
        }
    }

    if viewer.State == ViewStateImage {
        if len(viewer.Images[viewer.CurrentTile].Images) > 0 {
            vector.DrawFilledRect(screen, 0, float32(startY), float32(ScreenWidth), float32(ScreenHeight - startY), color.RGBA{0, 0, 0, 64}, false)
            middleX := ScreenWidth / 2
            middleY := ScreenHeight / 2

            tile := viewer.Images[viewer.CurrentTile]

            var options ebiten.DrawImageOptions
            var useImage *ebiten.Image

            if viewer.AnimationFrame != -1 && viewer.AnimationFrame < len(tile.Images) {
                useImage = viewer.ImageCache.GetImage(tile.Keys[viewer.AnimationFrame], tile.Images[viewer.AnimationFrame], viewer.Time)
            } else {
                useImage = viewer.ImageCache.GetImage(tile.Keys[viewer.CurrentImage], tile.Images[viewer.CurrentImage], viewer.Time)
            }

            bounds := useImage.Bounds()
            options.GeoM.Translate(float64(-bounds.Dx()) / 2.0, float64(-bounds.Dy()) / 2.0)
            options.GeoM.Scale(viewer.Scale, viewer.Scale)
            options.GeoM.Translate(float64(middleX), float64(middleY))
            screen.DrawImage(useImage, &options)

            x1, y1 := options.GeoM.Apply(0, 0)
            x2, y2 := options.GeoM.Apply(float64(bounds.Dx()), float64(bounds.Dy()))

            vector.StrokeRect(screen, float32(x1), float32(y1), float32(x2 - x1), float32(y2 - y1), 1, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, true)
        }
    }
}

func MakeViewer(data []*LbxData) (*Viewer, error) {
    font, err := common.LoadFont()
    if err != nil {
        return nil, err
    }

    viewer := &Viewer{
        Data: data,
        Scale: 1,
        Font: font,
        CurrentImage: 0,
        AnimationFrame: -1,
        AnimationCount: 0,
        State: ViewStateTiles,
        ImageCache: MakeImageCache(300),
    }

    maxLoad := make(chan bool, 4)
    for i := 0; i < 4; i++ {
        maxLoad <- true
    }

    indexes := make(map[string]int)

    imageIndex := 0
    for _, lbxData := range data {
        indexes[lbxData.Name] = imageIndex
        imageIndex += lbxData.Lbx.TotalEntries()
        for i := 0; i < lbxData.Lbx.TotalEntries(); i++ {
            loader := &LbxImages{}
            loader.LbxData = lbxData
            viewer.Images = append(viewer.Images, loader)

            go func(){
                <-maxLoad

                loader.Load.Do(func(){
                    rawImages, err := lbxData.Lbx.ReadImages(i)
                    if err != nil {
                        log.Printf("Unable to load images from %v at index %v: %v", lbxData.Name, i, err)
                        return
                    }

                    var keys []string
                    for z := 0; z < len(rawImages); z++ {
                        keys = append(keys, fmt.Sprintf("%v-%v-%v", lbxData.Name, i, z))
                    }

                    loader.Keys = keys
                    loader.Images = rawImages

                    loader.Lock.Lock()
                    loader.Loaded = true
                    loader.Lock.Unlock()
                })

                maxLoad <- true
            }()
        }
    }

    viewer.Indexes = indexes

    return viewer, nil
}

func MakeSingleFileViewer(path string) (*Viewer, error) {
    var lbxFile lbx.LbxFile

    open, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer open.Close()
    lbxFile, err = lbx.ReadLbx(open)
    if err != nil {
        return nil, err
    }
    log.Printf("Loaded lbx file: %v\n", path)

    return MakeViewer([]*LbxData{&LbxData{Lbx: &lbxFile, Name: path}})
}

func MakeMultiViewer(path string) (*Viewer, error) {
    entries, err := os.ReadDir(path)
    if err != nil {
        return nil, err
    }

    var data []*LbxData

    for _, entry := range entries {
        open, err := os.Open(path + "/" + entry.Name())
        if err != nil {
            continue
        }

        lbxFile, err := lbx.ReadLbx(open)
        if err == nil {
            data = append(data, &LbxData{Lbx: &lbxFile, Name: entry.Name()})
        }

        open.Close()
    }

    return MakeViewer(data)
}

func isFile(path string) bool {
    info, err := os.Stat(path)
    if err != nil {
        return false
    }
    return !info.IsDir()
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

    var viewer *Viewer
    var err error

    if isFile(file) {
        viewer, err = MakeSingleFileViewer(file)
        if err != nil {
            log.Printf("Error: %v", err)
            return
        }
    } else {
        viewer, err = MakeMultiViewer(file)
        if err != nil {
            log.Printf("Error: %v", err)
            return
        }
    }

    err = ebiten.RunGame(viewer)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
