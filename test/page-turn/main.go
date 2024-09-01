package main

import (
    "log"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type Engine struct {
    Cache *lbx.LbxCache
    ImageCache util.ImageCache
}

func (engine *Engine) Update() error {
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image){
    screen.Fill(color.RGBA{R: 0, G: 100, B: 200, A: 255})
    pages, _ := engine.ImageCache.GetImages("book.lbx", 0)
    var options ebiten.DrawImageOptions
    options.GeoM.Translate(10, 20)
    screen.DrawImage(pages[0], &options)

    options.GeoM.Translate(float64(pages[0].Bounds().Dx() + 10), 0)
    screen.DrawImage(pages[1], &options)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return ScreenWidth, ScreenHeight
}

func NewEngine() (*Engine, error){
    cache := lbx.AutoCache()
    return &Engine{
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
    }, nil
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
    ebiten.SetWindowTitle("page turn")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    engine, err := NewEngine()

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
