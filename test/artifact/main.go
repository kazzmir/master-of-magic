package main

import (
    "log"
    "fmt"

    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    Counter uint64
    Drawer func(*ebiten.Image)
    Cache *lbx.LbxCache
    Coroutine *coroutine.Coroutine

    ShowUpdate int
    Books Books
}

type Books struct {
    Artificer bool
    Runemaster bool
}

func (books *Books) MagicLevel(magic data.MagicType) int {
    switch magic {
        case data.ChaosMagic: return 11
        case data.MagicNone: return 0
    }

    return 0
}

func (books *Books) RetortEnabled(retort data.Retort) bool {
    if retort == data.RetortArtificer {
        return books.Artificer
    }

    if retort == data.RetortRunemaster {
        return books.Runemaster
    }

    return false
}

func NewEngine() (*Engine, error) {
    cache := lbx.AutoCache()
    engine := &Engine{
        Counter: 0,
        Cache: cache,
        Drawer: func(*ebiten.Image){},
    }

    engine.Coroutine = coroutine.MakeCoroutine(engine.ArtifactRoutine())

    return engine, nil
}

func (engine *Engine) ArtifactRoutine() func (coroutine.YieldFunc) error {
    spells, err := spellbook.ReadSpellsFromCache(engine.Cache)
    if err != nil {
        log.Printf("Unable to read spells: %v", err)
    }

    return func(yield coroutine.YieldFunc) error {
        create, cancel := artifact.ShowCreateArtifactScreen(yield, engine.Cache, artifact.CreationCreateArtifact, &engine.Books, spells.CombatSpells(false), &engine.Drawer)
        if !cancel {
            log.Printf("Create artifact: %+v", create)
        } else {
            log.Printf("Aborted")
        }
        return nil
    }
}

func (engine *Engine) Update() error {
    engine.Counter += 1
    inputmanager.Update()
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyEscape, ebiten.KeyCapsLock: return ebiten.Termination
            case ebiten.KeyF1:
                engine.Books.Artificer = !engine.Books.Artificer
                engine.ShowUpdate = 60
                engine.Coroutine = coroutine.MakeCoroutine(engine.ArtifactRoutine())
                log.Printf("Artificer %v Runemaster %v", engine.Books.Artificer, engine.Books.Runemaster)
            case ebiten.KeyF2:
                engine.Books.Runemaster = !engine.Books.Runemaster
                engine.ShowUpdate = 60
                engine.Coroutine = coroutine.MakeCoroutine(engine.ArtifactRoutine())
                log.Printf("Artificer %v Runemaster %v", engine.Books.Artificer, engine.Books.Runemaster)
        }
    }

    if engine.ShowUpdate > 0 {
        engine.ShowUpdate -= 1
    }

    if engine.Coroutine.Run() != nil {
        return ebiten.Termination
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image){
    engine.Drawer(screen)

    if engine.ShowUpdate > 0 {
        ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Artificer %v Runemaster %v", engine.Books.Artificer, engine.Books.Runemaster), 0, 0)
    }
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return scale.Scale2(data.ScreenWidth, data.ScreenHeight)
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    monitorWidth, _ := ebiten.Monitor().Size()

    size := monitorWidth / 390

    ebiten.SetWindowSize(data.ScreenWidth * size, data.ScreenHeight * size)

    ebiten.SetWindowTitle("page turn")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    audio.Initialize()

    engine, err := NewEngine()

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
