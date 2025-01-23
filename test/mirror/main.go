package main

import (
    "image/color"
    "log"
    "os"
    "strconv"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/mirror"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/hero"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    UI *uilib.UI
}

func NewEngine(scenario int) (*Engine, error) {
    cache := lbx.AutoCache()

    player:= playerlib.Player{
        Wizard: setup.WizardCustom{
            Base: data.WizardLoPan,
            Name: "Lo Pan",
            Books: []data.WizardBook{
                data.WizardBook{
                    Magic: data.LifeMagic,
                    Count: 3,
                },
                data.WizardBook{
                    Magic: data.NatureMagic,
                    Count: 4,
                },
            },
            Abilities: []setup.WizardAbility{
                setup.AbilityArchmage,
                setup.AbilityFamous,
            },
        },
        Fame: 100,
        Gold: 2123,
        Mana: 23455,
    }
    player.Heroes[0] = hero.MakeHeroSimple(hero.HeroTheria)
    player.Heroes[1] = hero.MakeHeroSimple(hero.HeroRakir)

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

    ui.AddElement(mirror.MakeMirrorUI(cache, &player, ui))

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
    return data.ScreenWidth, data.ScreenHeight
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    monitorWidth, _ := ebiten.Monitor().Size()

    size := monitorWidth / 390

    ebiten.SetWindowSize(data.ScreenWidth / data.ScreenScale * size, data.ScreenHeight / data.ScreenScale * size)

    ebiten.SetWindowTitle("mirror")
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
