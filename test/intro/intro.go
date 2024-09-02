package main

import (
    "os"
    "log"
    "strconv"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    introlib "github.com/kazzmir/master-of-magic/game/magic/intro"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    Intro *introlib.Intro
}

func NewEngine(speed int) (*Engine, error) {
    cache := lbx.AutoCache()

    intro, err := introlib.MakeIntro(cache, uint64(speed))

    if err != nil {
        return nil, err
    }

    return &Engine{
        LbxCache: cache,
        Intro: intro,
    }, nil
}

func (engine *Engine) Update() error {

    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    switch engine.Intro.Update() {
        case introlib.IntroStateRunning:
        case introlib.IntroStateDone:
            return ebiten.Termination
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    engine.Intro.Draw(screen)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return data.ScreenWidth, data.ScreenHeight
}

func main(){

    speed := introlib.DefaultAnimationSpeed

    if len(os.Args) > 1 {
        var err error
        speed, err = strconv.Atoi(os.Args[1])
        if err != nil {
            log.Printf("Not a number for animation speed %v: %v", os.Args[1], err)
            return
        }
    }

    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(data.ScreenWidth * 5, data.ScreenHeight * 5)
    ebiten.SetWindowTitle("intro")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    audio.Initialize()

    engine, err := NewEngine(speed)

    if err != nil {
        log.Printf("Error: unable to load engine: %v", err)
        return
    }

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }

}
