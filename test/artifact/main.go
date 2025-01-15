package main

import (
    "log"

    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ScreenWidth = 320
const ScreenHeight = 200

type Engine struct {
    Counter uint64
    Drawer func(*ebiten.Image)
    Cache *lbx.LbxCache
    Coroutine *coroutine.Coroutine

    Artificer bool
    Runemaster bool
}

type Books struct {
}

func (books *Books) MagicLevel(magic data.MagicType) int {
    switch magic {
        case data.ChaosMagic: return 11
    }

    return 11
}

func NewEngine() (*Engine, error) {
    cache := lbx.AutoCache()
    engine := &Engine{
        Counter: 0,
        Cache: cache,
        Drawer: func(*ebiten.Image){},
    }

    run := func(yield coroutine.YieldFunc) error {
        create, cancel := artifact.ShowCreateArtifactScreen(yield, engine.Cache, artifact.CreationCreateArtifact, &Books{}, engine.Artificer, engine.Runemaster, &engine.Drawer)
        if !cancel {
            log.Printf("Create artifact: %+v", create)
        } else {
            log.Printf("Aborted")
        }
        return nil
    }

    engine.Coroutine = coroutine.MakeCoroutine(run)

    return engine, nil
}

func (engine *Engine) Update() error {
    engine.Counter += 1
    inputmanager.Update()
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    if engine.Coroutine.Run() != nil {
        return ebiten.Termination
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image){
    engine.Drawer(screen)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return ScreenWidth, ScreenHeight
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(ScreenWidth * 5, ScreenHeight * 5)
    ebiten.SetWindowTitle("page turn")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    audio.Initialize()

    engine, err := NewEngine()

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
