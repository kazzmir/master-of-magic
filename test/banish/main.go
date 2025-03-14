package main

import (
    "log"
    "strconv"
    "os"
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/banish"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/setup"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    Coroutine *coroutine.Coroutine
    DrawScene func (*ebiten.Image)
}

func randomWizard() data.WizardBase {
    choices := []data.WizardBase{
        data.WizardMerlin, data.WizardRaven, data.WizardSharee,
        data.WizardLoPan, data.WizardJafar, data.WizardOberic,
        data.WizardRjak, data.WizardSssra, data.WizardTauron,
        data.WizardFreya, data.WizardHorus, data.WizardAriel,
        data.WizardTlaloc, data.WizardKali,
    }

    return choices[rand.IntN(len(choices))]
}

func NewEngine(scenario int) (*Engine, error) {
    cache := lbx.AutoCache()

    player1 := playerlib.Player{
        Wizard: setup.WizardCustom{
            Base: randomWizard(),
            Name: "bob",
        },
    }

    player2 := playerlib.Player{
        Wizard: setup.WizardCustom{
            Base: randomWizard(),
            Name: "Kali",
        },
    }

    logic, draw := banish.ShowBanishAnimation(cache, &player1, &player2)

    return &Engine{
        LbxCache: cache,
        DrawScene: draw,
        Coroutine: coroutine.MakeCoroutine(logic),
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

    if engine.Coroutine.Run() != nil {
        return ebiten.Termination
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    engine.DrawScene(screen)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return scale.Scale2(data.ScreenWidth, data.ScreenHeight)
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    monitorWidth, _ := ebiten.Monitor().Size()
    size := monitorWidth / 390

    ebiten.SetWindowSize(data.ScreenWidth * size, data.ScreenHeight * size)
    ebiten.SetWindowTitle("banish")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    audio.Initialize()

    scenario := 1

    if len(os.Args) >= 2 {
        x, err := strconv.Atoi(os.Args[1])
        if err != nil {
            log.Fatalf("Error with scenario: %v", err)
        }

        scenario = x
    }

    engine, err := NewEngine(scenario)

    if err != nil {
        log.Printf("Error: unable to load engine: %v", err)
        return
    }

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
