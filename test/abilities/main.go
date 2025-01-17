package main

import (
    "log"
    "strconv"
    "os"
    // "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/unitview"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    mouselib "github.com/kazzmir/master-of-magic/lib/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/mouse"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    UI *uilib.UI
}

type Experience struct {
}

func (experience *Experience) HasWarlord() bool {
    return false
}

func (experience *Experience) Crusade() bool {
    return false
}

func NewEngine(scenario int) (*Engine, error) {
    cache := lbx.AutoCache()

    normalMouse, err := mouselib.GetMouseNormal(cache)
    if err == nil {
        mouse.Mouse.SetImage(normalMouse)
    }

    var ui *uilib.UI

    switch scenario {
        default:
            ui = &uilib.UI{
                Draw: func(ui *uilib.UI, screen *ebiten.Image) {
                    ui.IterateElementsByLayer(func (element *uilib.UIElement) {
                        if element.Draw != nil {
                            element.Draw(element, screen)
                        }
                    })
                },
            }

            ui.SetElementsFromArray(nil)

            baseUnit := units.HeroRakir

            hero := herolib.MakeHero(units.MakeOverworldUnitFromUnit(baseUnit, 1, 1, data.PlaneArcanus, data.BannerBrown, &Experience{}), herolib.HeroRakir, "rakir")

            ui.AddElements(unitview.MakeUnitContextMenu(cache, ui, hero, func(){}))
    }

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

    inputmanager.Update()

    engine.UI.StandardUpdate()

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    engine.UI.Draw(engine.UI, screen)

    mouse.Mouse.Draw(screen)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return data.ScreenWidth, data.ScreenHeight
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    scenario := 1
    if len(os.Args) > 1 {
        scenario, _ = strconv.Atoi(os.Args[1])
    }

    ebiten.SetWindowSize(data.ScreenWidth * 3, data.ScreenHeight * 3)
    ebiten.SetWindowTitle("abilities")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
    ebiten.SetCursorMode(ebiten.CursorModeHidden)

    audio.Initialize()
    mouse.Initialize()

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
