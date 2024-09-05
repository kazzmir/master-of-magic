package main

import (
    "log"
    "fmt"
    "bytes"

    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type View struct {
    Cache *lbx.LbxCache
    Mice []*ebiten.Image
}

func MakeView(cache *lbx.LbxCache) (*View, error) {

    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        return nil, err
    }

    data, err := fontLbx.RawData(2)
    if err != nil {
        return nil, err
    }

    var mainPalette color.Palette

    paletteData := data[0:256*3]
    for i := 0; i < 256; i++ {
        r := paletteData[i*3]
        g := paletteData[i*3+1]
        b := paletteData[i*3+2]
        // log.Printf("palette[%d] = %d, %d, %d", i, r, g, b)
        mainPalette = append(mainPalette, color.RGBA{R: r, G: g, B: b, A: 255})
    }

    // make transparent
    mainPalette[0] = color.RGBA{R: 0, G: 0, B: 0, A: 0}

    // 32 arrays of 16 colors
    fontColors := data[256*3:256*3 + 1280-768]
    _ = fontColors

    // each pointer is 0x100 bytes
    mouseData := data[1280:5376]

    length := 0x100

    var mousePics []*ebiten.Image

    for i := 0; i < 16; i++ {
        mouse := mouseData[i*length:i*length + length]
        pic := ebiten.NewImage(16, 16)

        reader := bytes.NewReader(mouse)

        for x := 0; x < 16; x++ {
            for y := 0; y < 16; y++ {
                colorIndex, err := reader.ReadByte()
                if err != nil {
                    return nil, err
                }

                color := mainPalette[colorIndex]
                pic.Set(x, y, color)
            }
        }

        mousePics = append(mousePics, pic)
    }

    return &View{
        Cache: cache,
        Mice: mousePics,
    }, nil
}

func (view *View) Update() error {
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
    for _, pic := range view.Mice {
        options.GeoM.Translate(60, 0)
        screen.DrawImage(pic, &options)
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
    
    err = ebiten.RunGame(editor)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
