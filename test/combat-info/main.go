package main

import (
    "log"
    "strconv"
    "os"
    // "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    UI *uilib.UI
}

type ExperienceInfo struct {
}

func (info *ExperienceInfo) Crusade() bool {
    return false
}

func (info *ExperienceInfo) HasWarlord() bool {
    return false
}

func NewEngine(scenario int) (*Engine, error) {
    cache := lbx.AutoCache()

    defendingPlayer := &playerlib.Player{}
    attackingPlayer := &playerlib.Player{}

    defendingArmy := &combat.Army{
        Player: defendingPlayer,
    }

    attackingArmy := &combat.Army{
        Player: attackingPlayer,
    }

    model := combat.MakeCombatModel(cache, defendingArmy, attackingArmy, combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, 0, 0, make(chan combat.CombatEvent))

    unit := &combat.ArmyUnit{
        Unit: units.MakeOverworldUnitFromUnit(units.ArchAngel, 1, 1, data.PlaneArcanus, data.BannerRed, &ExperienceInfo{}),
        Model: model,
    }

    unit.AddCurse(data.CurseMindStorm)
    unit.TakeDamage(5)

    ui := &uilib.UI{
        Draw: func(this *uilib.UI, screen *ebiten.Image) {
            this.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })
        },
    }

    ui.SetElementsFromArray(nil)
    ui.AddGroup(combat.MakeUnitView(cache, ui, unit))

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

    monitorWidth, _ := ebiten.Monitor().Size()
    size := monitorWidth / 390
    ebiten.SetWindowSize(data.ScreenWidth / data.ScreenScale * size, data.ScreenHeight / data.ScreenScale * size)

    ebiten.SetWindowTitle("combat info")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
    // ebiten.SetCursorMode(ebiten.CursorModeHidden)

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
