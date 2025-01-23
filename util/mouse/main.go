package main

import (
    "log"
    "fmt"
    "image"

    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/util"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type MouseImage int

const (
    MouseNormal MouseImage = iota
    MouseMagic
    MouseError
    MouseArrow
    MouseAttack
    MouseWait
    MouseMove
    MouseCast
)

func nextMouse(mouse MouseImage) MouseImage {
    switch mouse {
        case MouseNormal: return MouseMagic
        case MouseMagic: return MouseError
        case MouseError: return MouseArrow
        case MouseArrow: return MouseAttack
        case MouseAttack: return MouseWait
        case MouseWait: return MouseMove
        case MouseMove: return MouseCast
        case MouseCast: return MouseNormal
    }

    return MouseNormal
}

type View struct {
    Cache *lbx.LbxCache
    Mice []*ebiten.Image
    Cast []*ebiten.Image
    Counter uint64

    Mouse MouseImage

    Images map[MouseImage]*ebiten.Image
}

type ScaleNone struct {
}

func (scaler *ScaleNone) ApplyScale(input image.Image) image.Image {
    return util.Scale2x(input, false)
}

func MakeView(cache *lbx.LbxCache) (*View, error) {

    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        return nil, err
    }

    scaler := &ScaleNone{}

    var all []*ebiten.Image
    for i := 2; i < 9; i++ {
        /*
        data, err := fontLbx.RawData(i)
        if err != nil {
            return nil, err
        }

        mousePics, err := readMousePics(data)
        */
        mousePics, err := mouse.ReadMouseImages(fontLbx, scaler, i)
        if err != nil {
            return nil, err
        }
        all = append(all, mousePics...)
    }

    castImages, err := mouse.GetMouseCast(cache, scaler)
    if err != nil {
        return nil, err
    }

    images := make(map[MouseImage]*ebiten.Image)
    x, err := mouse.GetMouseNormal(cache, scaler)
    if err == nil {
        images[MouseNormal] = x
    }
    x, err = mouse.GetMouseMagic(cache, scaler)
    if err == nil {
        images[MouseMagic] = x
    }
    x, err = mouse.GetMouseError(cache, scaler)
    if err == nil {
        images[MouseError] = x
    }
    x, err = mouse.GetMouseArrow(cache, scaler)
    if err == nil {
        images[MouseArrow] = x
    }
    x, err = mouse.GetMouseAttack(cache, scaler)
    if err == nil {
        images[MouseAttack] = x
    }
    x, err = mouse.GetMouseWait(cache, scaler)
    if err == nil {
        images[MouseWait] = x
    }
    x, err = mouse.GetMouseMove(cache, scaler)
    if err == nil {
        images[MouseMove] = x
    }

    return &View{
        Cache: cache,
        Mice: all,
        Mouse: MouseNormal,
        Cast: castImages,
        Images: images,
    }, nil
}

func (view *View) Update() error {
    view.Counter += 1
    var keys []ebiten.Key
    keys = make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyTab:
                view.Mouse = nextMouse(view.Mouse)
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

    options.GeoM.Reset()
    options.GeoM.Scale(3, 3)
    mouseX, mouseY := ebiten.CursorPosition()
    options.GeoM.Translate(float64(mouseX), float64(mouseY))
    switch view.Mouse {
        case MouseCast:
            index := int((view.Counter/4) % uint64(len(view.Cast)))
            screen.DrawImage(view.Cast[index], &options)
        default:
            img, ok := view.Images[view.Mouse]
            if ok {
                screen.DrawImage(img, &options)
            }
    }
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
