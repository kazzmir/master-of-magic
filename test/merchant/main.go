package main

import (
    "log"
    "image/color"
    "strconv"
    "os"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    UI *uilib.UI
}

func NewEngine(scenario int) (*Engine, error) {
    cache := lbx.AutoCache()

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){

        ui.IterateElementsByLayer(func (element *uilib.UIElement){
            if element.Draw != nil {
                element.Draw(element, screen)
            }
        })

        },
    }
    ui.SetElementsFromArray(nil)

    artifact := &artifact.Artifact{
        Name: "Excalibur",
        Image: 7,
        Type: artifact.ArtifactTypeSword,
        Powers: []artifact.Power{
            {
                Type: artifact.PowerTypeAttack,
                Amount: 1,
                Name: "+3 Attack",
            },
            {
                Type: artifact.PowerTypeDefense,
                Amount: 2,
                Name: "+2 Defense",
            },
            {
                Type: artifact.PowerTypeSpellSkill,
                Amount: 2,
                Name: "+2 Spell Skill",
            },
            /*
            {
                Type: artifact.PowerTypeSpellSave,
                Amount: 2,
                Name: "-2 Spell Save",
            },
            */
            {
                Type: artifact.PowerTypeAbility1,
                Name: "Flaming",
                Ability: data.ItemAbilityFlaming,
            },
        },
        Cost: 250,
    }

    ui.AddElements(gamelib.MakeMerchantScreenUI(cache, ui, artifact, 250, func (bought bool){
        log.Printf("bought %v", bought)
    }))

    return &Engine{
        LbxCache: cache,
        UI: ui,
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

    engine.UI.StandardUpdate()

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{R: 0, G: 150, B: 150, A: 0xff})

    engine.UI.Draw(engine.UI, screen)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return scale.Scale2(data.ScreenWidth, data.ScreenHeight)
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    monitorWidth, _ := ebiten.Monitor().Size()
    size := monitorWidth / 390

    ebiten.SetWindowSize(data.ScreenWidth * size, data.ScreenHeight * size)
    ebiten.SetWindowTitle("summon unit")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

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
