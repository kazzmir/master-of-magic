package main

import (
    "log"
    "strconv"
    "os"
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/mastery"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
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

    /*
    player2 := playerlib.Player{
        Wizard: setup.WizardCustom{
            Base: randomWizard(),
            Name: "Kali",
        },
    }
    */

    switch scenario {
        case 0:
            var currentDraw func (*ebiten.Image)

            draw := func (screen *ebiten.Image) {
                currentDraw(screen)
            }

            logic := func (yield coroutine.YieldFunc) error {
                logic1, draw1 := mastery.ShowSpellOfMasteryScreen(cache, player1.Wizard.Name)
                currentDraw = draw1

                logic1(yield)

                logic2, draw2 := mastery.LabVortexScreen(cache, player1.Wizard.Base, []data.WizardBase{data.WizardMerlin, data.WizardRaven /*,data.WizardSharee */})
                currentDraw = draw2
                logic2(yield)

                logic3, draw3 := mastery.SpellOfMasteryEndScreen(cache, player1.Wizard.Base)
                currentDraw = draw3
                logic3(yield)

                return nil
            }

            return &Engine{
                LbxCache: cache,
                DrawScene: draw,
                Coroutine: coroutine.MakeCoroutine(logic),
            }, nil


        case 1:
            logic, draw := mastery.ShowSpellOfMasteryScreen(cache, player1.Wizard.Name)

            return &Engine{
                LbxCache: cache,
                DrawScene: draw,
                Coroutine: coroutine.MakeCoroutine(logic),
            }, nil
        case 2:
            logic, draw := mastery.LabVortexScreen(cache, player1.Wizard.Base, []data.WizardBase{data.WizardMerlin, data.WizardRaven /*,data.WizardSharee */})

            return &Engine{
                LbxCache: cache,
                DrawScene: draw,
                Coroutine: coroutine.MakeCoroutine(logic),
            }, nil

        case 3:
            logic, draw := mastery.SpellOfMasteryEndScreen(cache, player1.Wizard.Base)

            return &Engine{
                LbxCache: cache,
                DrawScene: draw,
                Coroutine: coroutine.MakeCoroutine(logic),
            }, nil
        default:
            logic, draw := mastery.ShowSpellOfMasteryScreen(cache, player1.Wizard.Name)

            return &Engine{
                LbxCache: cache,
                DrawScene: draw,
                Coroutine: coroutine.MakeCoroutine(logic),
            }, nil
    }
}

func (engine *Engine) Update() error {

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
