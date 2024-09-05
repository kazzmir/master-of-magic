package main

import (
    "log"
    "fmt"

    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type View struct {
    Cache *lbx.LbxCache
}

func MakeView(cache *lbx.LbxCache) *View {
    return &View{
        Cache: cache,
    }
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
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    cache := lbx.AutoCache()

    editor := MakeView(cache)

    ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
    ebiten.SetWindowTitle("mouse viewer")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
    
    err := ebiten.RunGame(editor)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
