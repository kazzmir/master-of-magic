package main

import (
    "log"
    "os"
    "fmt"
    "sync"
    "math"
    "flag"
    // "slices"

    "image/color"
    "image"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/util/common"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
)

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

// scan through entire cache and remove keys that haven't been accessed in 60 seconds
func (cache *ImageCache) Cleanup(currentTime uint64){
    var toRemove []string

    for key, data := range cache.Images {

        if data.Time + 60 < currentTime {
            toRemove = append(toRemove, key)
        }
    }

    // log.Printf("Evicting %v keys", len(toRemove))
    for _, key := range toRemove {
        // log.Printf("Cache eviction: %v", key.Key)
        delete(cache.Images, key)
    }
}

func MakeImageCache() ImageCache {
    return ImageCache{
        Images: make(map[string]CacheData),
    }
}

type LbxImages struct {
    Keys []string
    Images []*image.Paletted
    CustomPalette bool
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
    // Data []*LbxData
    StartingRow int
    Indexes map[string]int
    Images []*LbxImages
    Scale float64
    CurrentImage int
    CurrentTile int
    State ViewerState
    ShowPalette bool
    Font *text.GoTextFaceSource
    AnimationFrame int
    AnimationCount int
    ShiftCount int
    Time uint64
    ImageCache ImageCache

    ScreenWidth int
    ScreenHeight int
}

const TileWidth = 60
const TileHeight = 60

func (viewer *Viewer) tilesPerRow() int {
    width := viewer.ScreenWidth - 1
    return width / TileWidth
}

func (viewer *Viewer) Update() error {
    if viewer.Time % 60 == 0 {
        viewer.ImageCache.Cleanup(viewer.Time)
    }
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

    quick := viewer.ShiftCount % shiftSpeed == 1

    if ebiten.IsKeyPressed(ebiten.KeyControlLeft) {
        quick = true
    }

    for _, key := range keys {
        switch key {
            case ebiten.KeyUp, ebiten.KeyK:
                if viewer.State == ViewStateImage {
                    viewer.Scale *= 1 + scaleAmount
                }
                if viewer.State == ViewStateTiles && quick {
                    press_up = true
                }
            case ebiten.KeyDown, ebiten.KeyJ:
                if viewer.State == ViewStateImage {
                    viewer.Scale *= 1 - scaleAmount
                    if viewer.Scale < 1 {
                        viewer.Scale = 1
                    }
                }

                if viewer.State == ViewStateTiles && quick {
                    press_down = true
                }
            case ebiten.KeyLeft, ebiten.KeyH:
                if viewer.State == ViewStateTiles && quick {
                    press_left = true
                }
            case ebiten.KeyRight, ebiten.KeyL:
                if viewer.State == ViewStateTiles && quick {
                    press_right = true
                }
            case ebiten.KeyPageDown:
                if viewer.State == ViewStateTiles {
                    press_down = true
                }
            case ebiten.KeyPageUp:
                if viewer.State == ViewStateTiles {
                    press_up = true
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

            case ebiten.KeyTab:
                if viewer.State == ViewStateImage {
                    viewer.ShowPalette = !viewer.ShowPalette
                }

            case ebiten.KeyLeft, ebiten.KeyH:
                switch viewer.State {
                    case ViewStateTiles:
                        press_left = true
                    case ViewStateImage:
                        viewer.CurrentImage -= 1
                        if viewer.CurrentImage < 0 {
                            viewer.CurrentImage = len(viewer.Images[viewer.CurrentTile].Images) - 1
                        }
                }

            case ebiten.KeyRight, ebiten.KeyL:
                switch viewer.State {
                    case ViewStateTiles:
                        press_right = true
                    case ViewStateImage:
                        viewer.CurrentImage += 1
                        if viewer.CurrentImage >= len(viewer.Images[viewer.CurrentTile].Images) {
                            viewer.CurrentImage = 0
                        }
                }

            case ebiten.KeyUp, ebiten.KeyK:
                switch viewer.State {
                    case ViewStateTiles:
                        press_up = true
                }

            case ebiten.KeyDown, ebiten.KeyJ:
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
        position := viewer.CurrentTile - viewer.tilesPerRow()
        if position >= 0 {
            viewer.CurrentTile = position
            viewer.CurrentImage = 0
        }
    }

    if press_down {
        position := viewer.CurrentTile + viewer.tilesPerRow()
        if position < len(viewer.Images) {
            viewer.CurrentTile = position
            viewer.CurrentImage = 0
        }
    }

    tilesPerRow := (viewer.ScreenWidth - 1) / TileWidth
    tilesPerColumn := (viewer.ScreenHeight - 100) / TileHeight - 1

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
    viewer.ScreenWidth = outsideWidth
    viewer.ScreenHeight = outsideHeight
    return outsideWidth, outsideHeight
    // return ScreenWidth, ScreenHeight
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

    if viewer.CurrentTile >= len(viewer.Images) {
        return
    }

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
    op.GeoM.Translate(0, 20)
    text.Draw(screen, fmt.Sprintf("Has Palette: %v", viewer.Images[viewer.CurrentTile].CustomPalette), face, op)

    startX := 1
    startY := 100

    x := startX
    y := startY

    for i, image := range viewer.Images {
        if i < viewer.StartingRow * viewer.tilesPerRow() {
            continue
        }
        if image.IsLoaded() && len(image.Images) > 0 {
            var options ebiten.DrawImageOptions

            var draw *ebiten.Image
            if i == viewer.CurrentTile && viewer.AnimationFrame != -1 && viewer.AnimationFrame < len(image.Images) {
                draw = viewer.ImageCache.GetImage(image.Keys[viewer.AnimationFrame], image.Images[viewer.AnimationFrame], viewer.Time)
            } else {
                draw = viewer.ImageCache.GetImage(image.Keys[0], image.Images[0], viewer.Time)
            }

            // draw := viewer.ImageCache.GetImage(image.Keys[0], image.Images[0], viewer.Time)

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
        if x + TileWidth >= viewer.ScreenWidth {
            x = 1
            y += TileHeight

            if y >= viewer.ScreenHeight {
                break
            }
        }
    }

    if viewer.State == ViewStateImage {
        if len(viewer.Images[viewer.CurrentTile].Images) > 0 {
            vector.DrawFilledRect(screen, 0, float32(startY), float32(viewer.ScreenWidth), float32(viewer.ScreenHeight - startY), color.RGBA{0, 0, 0, 92}, false)
            middleX := viewer.ScreenWidth / 2
            middleY := viewer.ScreenHeight / 2

            tile := viewer.Images[viewer.CurrentTile]

            if viewer.ShowPalette {
                useImage := tile.Images[viewer.CurrentImage]
                bounds := useImage.Bounds()
                var options ebiten.DrawImageOptions
                tileSize := 8
                options.GeoM.Translate(float64(-bounds.Dx() * tileSize) / 2.0, float64(-bounds.Dy() * tileSize) / 2.0)
                options.GeoM.Scale(viewer.Scale, viewer.Scale)
                options.GeoM.Translate(float64(middleX), float64(middleY))

                x1, y1 := options.GeoM.Apply(0, 0)
                x2, y2 := options.GeoM.Apply(float64(bounds.Dx() * tileSize), float64(bounds.Dy() * tileSize))
                vector.DrawFilledRect(screen, float32(x1), float32(y1), float32(x2 - x1), float32(y2 - y1), color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, false)

                for x := x1; x < x2; x += float64(tileSize) * viewer.Scale {
                    vector.StrokeLine(screen, float32(x), float32(y1), float32(x), float32(y2), 1, color.RGBA{R: 0, G: 0, B: 0, A: 0xff}, false)
                }

                for y := y1; y < y2; y += float64(tileSize) * viewer.Scale {
                    vector.StrokeLine(screen, float32(x1), float32(y), float32(x2), float32(y), 1, color.RGBA{R: 0, G: 0, B: 0, A: 0xff}, false)
                }

                face := &text.GoTextFace{Source: viewer.Font, Size: 3 * viewer.Scale}
                op := text.DrawOptions{}
                op.GeoM.Translate(1, 1)
                op.ColorScale.ScaleWithColor(color.Black)

                for x := 0; x < bounds.Dx(); x++ {
                    for y := 0; y < bounds.Dy(); y++ {
                        posX, posY := options.GeoM.Apply(float64(x * tileSize) + 2, float64(y * tileSize) + 2)

                        /*
                        r, g, b, a := useImage.Palette[image.At(x, y)].RGBA()
                        useColor := color.RGBA{
                            R: uint8(r),
                            G: uint8(g),
                            B: uint8(b),
                            A: uint8(a),
                        }
                        */
                        useColor := useImage.At(x, y)
                        vector.DrawFilledCircle(screen, float32(posX), float32(posY), 2 * float32(viewer.Scale), useColor, false)

                        textX, textY := options.GeoM.Apply(float64(x * tileSize) + 1, float64(y * tileSize) + 4)

                        op.GeoM.Reset()
                        op.GeoM.Translate(textX, textY)
                        index := useImage.ColorIndexAt(x, y)
                        text.Draw(screen, fmt.Sprintf("%v", index), face, &op)
                    }
                }

            } else {
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

                vector.StrokeRect(screen, float32(x1), float32(y1), float32(x2 - x1), float32(y2 - y1), 1, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, false)
            }
        }
    }
}

func MakeViewer(dataPath string, names []string) (*Viewer, error) {
    font, err := common.LoadFont()
    if err != nil {
        return nil, err
    }

    viewer := &Viewer{
        // Data: data,
        Scale: 1,
        Font: font,
        CurrentImage: 0,
        AnimationFrame: -1,
        AnimationCount: 0,
        State: ViewStateTiles,
        ImageCache: MakeImageCache(),
    }

    maxLoad := make(chan bool, 4)
    for i := 0; i < 4; i++ {
        maxLoad <- true
    }

    indexes := make(map[string]int)

    var cache *lbx.LbxCache
    if dataPath == "" {
        cache = lbx.AutoCache()
    } else {
        if isFile(dataPath) {
            file, err := os.Open(dataPath)
            if err != nil {
                return nil, err
            }
            defer file.Close()
            lbxFile, err := lbx.ReadLbx(file)
            if err != nil {
                return nil, err
            }

            lbxFiles := make(map[string]*lbx.LbxFile)
            lbxFiles[dataPath] = &lbxFile

            cache = lbx.MakeCacheFromLbxFiles(lbxFiles)
        } else {

            cache = lbx.CacheFromPath(dataPath)
            if cache == nil {
                return nil, fmt.Errorf("No lbx files found at %v", dataPath)
            }
        }
    }

    if len(names) == 0 {
        // get all lbx files to show up
        names = append(names, "")
    }

    imageIndex := 0
    for _, name := range names {
        allLbx := cache.GetLbxFilesSimilarName(name)
        if len(allLbx) == 0 {
            log.Printf("No LBX files found for name '%v'", name)
            continue
        }

        for _, lbxDataName := range allLbx {
            lbxFile, err := cache.GetLbxFile(lbxDataName)
            if err != nil {
                log.Printf("Unable to load lbx file %v: %v", lbxDataName, err)
                continue
            }

            indexes[lbxDataName] = imageIndex
            imageIndex += lbxFile.TotalEntries()

            customPaletteMap, err := lbx.GetPaletteOverrideMap(cache, lbxFile, lbxDataName)
            if err != nil {
                return nil, err
            }

            for i := 0; i < lbxFile.TotalEntries(); i++ {
                loader := &LbxImages{}
                loader.LbxData = &LbxData{Name: lbxDataName, Lbx: lbxFile}
                viewer.Images = append(viewer.Images, loader)

                go func(){
                    <-maxLoad

                    loader.Load.Do(func(){

                        palette := customPaletteMap[i]
                        if palette == nil {
                            palette = customPaletteMap[-1]
                        }

                        rawImages, err := lbxFile.ReadImagesWithPalette(i, palette, palette != nil)
                        if err != nil {
                            log.Printf("Unable to load images from %v at index %v: %v", lbxDataName, i, err)
                            return
                        }

                        var keys []string
                        for z := 0; z < len(rawImages); z++ {
                            keys = append(keys, fmt.Sprintf("%v-%v-%v", lbxDataName, i, z))
                        }

                        loader.Keys = keys
                        loader.Images = rawImages

                        hasPalette, err := lbxFile.GetPalette(i)
                        if err == nil && hasPalette != nil {
                            loader.CustomPalette = true
                        } else {
                            loader.CustomPalette = false
                        }

                        loader.Lock.Lock()
                        loader.Loaded = true
                        loader.Lock.Unlock()
                    })

                    maxLoad <- true
                }()
            }
        }
    }

    viewer.Indexes = indexes

    return viewer, nil
}

/*
func MakeViewerFromFiles(paths []string) (*Viewer, error) {
    return MakeViewer(paths)
}
*/

func isFile(path string) bool {
    info, err := os.Stat(path)
    if err != nil {
        return false
    }
    return !info.IsDir()
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    var dataPath string
    flag.StringVar(&dataPath, "data", "", "path to master of magic lbx files. Given either a directory, zip file, or a single lbx file. Data is searched for in the current directory if not given.")

    flag.Parse()

    ebiten.SetWindowSize(1100, 1000)
    ebiten.SetWindowTitle("lbx viewer")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    restArgs := os.Args[len(os.Args) - flag.NArg():]

    viewer, err := MakeViewer(dataPath, restArgs)
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }

    err = ebiten.RunGame(viewer)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
