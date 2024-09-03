package main

import (
    "log"
    // "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    CombatScreen *combat.CombatScreen
}

func NewEngine() (*Engine, error) {
    cache := lbx.AutoCache()

    defendingArmy := combat.Army{
        Units: []*combat.ArmyUnit{
            &combat.ArmyUnit{
                Unit: units.HighElfSpearmen,
                Facing: units.FacingDownRight,
                X: 12,
                Y: 10,
            },
            &combat.ArmyUnit{
                Unit: units.HighElfSpearmen,
                Facing: units.FacingDownRight,
                X: 13,
                Y: 10,
            },
        },
    }

    attackingArmy := combat.Army{
        Units: []*combat.ArmyUnit{
            &combat.ArmyUnit{
                Unit: units.GreatDrake,
                Facing: units.FacingUpLeft,
                X: 11,
                Y: 19,
            },
        },
    }

    return &Engine{
        LbxCache: cache,
        CombatScreen: combat.MakeCombatScreen(cache, &defendingArmy, &attackingArmy),
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

    switch engine.CombatScreen.Update() {
        case combat.CombatStateRunning:
        case combat.CombatStateDone:
            return ebiten.Termination
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    engine.CombatScreen.Draw(screen)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return data.ScreenWidth, data.ScreenHeight
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(data.ScreenWidth * 5, data.ScreenHeight * 5)
    ebiten.SetWindowTitle("combat screen")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    engine, err := NewEngine()

    if err != nil {
        log.Printf("Error: unable to load engine: %v", err)
        return
    }

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
