package main

import (
    "log"
    "fmt"

    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/mouse"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type View struct {
    Cache *lbx.LbxCache
    Mice []*ebiten.Image
    Cast []*ebiten.Image
    Counter uint64
}

func MakeView(cache *lbx.LbxCache) (*View, error) {

    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        return nil, err
    }

    var all []*ebiten.Image
    for i := 2; i < 9; i++ {
        /*
        data, err := fontLbx.RawData(i)
        if err != nil {
            return nil, err
        }

        mousePics, err := readMousePics(data)
        */
        mousePics, err := mouse.ReadMouseImages(fontLbx, i)
        if err != nil {
            return nil, err
        }
        all = append(all, mousePics...)
    }

    castImages, err := mouse.GetMouseCast(cache)
    if err != nil {
        return nil, err
    }

    return &View{
        Cache: cache,
        Mice: all,
        Cast: castImages,
    }, nil
}

func (view *View) Update() error {
    view.Counter += 1
    var keys []ebiten.Key
    keys = make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyEscape, ebiten.KeyCapsLock:
                return ebiten.Termination
        }
    }

    return nil
}

func (view *View) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return ScreenWidth, ScreenHeight
}

func (view *View) Draw(screen *ebiten.Image){
    screen.Fill(color.RGBA{R: 0x4c, G: 0xe2, B: 0xed, A: 0xff})
    var options ebiten.DrawImageOptions
    options.GeoM.Scale(3, 3)
    options.GeoM.Translate(10, 10)
    x := 0
    spacing := 55
    for _, pic := range view.Mice {
        options.GeoM.Translate(float64(spacing), 0)
        screen.DrawImage(pic, &options)
        x += 1
        if x >= 16 {
            options.GeoM.Translate(float64(-(x * spacing)), 60)
            x = 0
        }
    }

    mouseX, mouseY := ebiten.CursorPosition()
    index := int((view.Counter/4) % uint64(len(view.Cast)))
    options.GeoM.Reset()
    options.GeoM.Scale(3, 3)
    options.GeoM.Translate(float64(mouseX), float64(mouseY))
    screen.DrawImage(view.Cast[index], &options)
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    cache := lbx.AutoCache()

    editor, err := MakeView(cache)

    if err != nil {
        log.Printf("Error: %v", err)
        return
    }

    ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
    ebiten.SetWindowTitle("mouse viewer")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
    ebiten.SetCursorMode(ebiten.CursorModeHidden)
    
    err = ebiten.RunGame(editor)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
