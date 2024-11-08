package main

import (
    "log"
    "image/color"

    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ScreenWidth = 320
const ScreenHeight = 200

type Engine struct {
    Counter uint64
    Cache *lbx.LbxCache
    ImageCache *util.ImageCache
}

func NewEngine() (*Engine, error) {
    cache := lbx.AutoCache()
    imageCache := util.MakeImageCache(cache)
    engine := &Engine{
        Counter: 0,
        Cache: cache,
        ImageCache: &imageCache,
    }

    return engine, nil
}

func (engine *Engine) Update() error {
    engine.Counter += 1
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
    screen.Fill(color.RGBA{R: 60, G: 60, B: 120, A: 255})

    lizardUnit, _ := engine.ImageCache.GetImage("units2.lbx", 0, 0)

    var options ebiten.DrawImageOptions
    options.GeoM.Translate(20, 20)
    screen.DrawImage(lizardUnit, &options)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return ScreenWidth, ScreenHeight
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(ScreenWidth * 4, ScreenHeight * 4)
    ebiten.SetWindowTitle("glow test")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    engine, err := NewEngine()

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
