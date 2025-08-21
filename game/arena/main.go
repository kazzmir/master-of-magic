package main

import (
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

/*
 * Start with a small army (single unit?), fight a battle. If you win then you get money/score that you can use to buy more units and spells.
 *  1. start game, pick a wizard portrait, name, etc
 *  2. pick an army from a small set of units
 *  3. fight a battle against an equivalent foe
 *  4. use money to buy more units and spells
 *  5. repeat from step 3
 */

type GameMode int

const (
    GameModeUI GameMode = iota
    GameModeBattle
)

type Player struct {
    Name string
    Money uint64

    Units []units.StackUnit
}

func MakePlayer() *Player {
    return &Player{
        Name: "Player",
        Money: 1000,
        Units: []units.StackUnit{
            units.MakeOverworldUnitFromUnit(units.LizardSwordsmen, 0, 0, data.PlaneArcanus, data.BannerGreen, &units.NoExperienceInfo{}, &units.NoEnchantments{}),
        },
    }
}

type Engine struct {
    GameMode GameMode
    Player *Player
}

func (engine *Engine) Update() error {
    keys := inpututil.AppendJustPressedKeys(nil)

    for _, key := range keys {
        switch key {
            case ebiten.KeyEscape, ebiten.KeyCapsLock:
                return ebiten.Termination
        }
    }

    return nil
}

func (engine *Engine) DrawUI(screen *ebiten.Image) {
}

func (engine *Engine) DrawBattle(screen *ebiten.Image) {
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    switch engine.GameMode {
        case GameModeUI:
            engine.DrawUI(screen)
        case GameModeBattle:
            engine.DrawBattle(screen)
    }
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (int, int) {
    switch engine.GameMode {
        case GameModeUI: return outsideWidth, outsideHeight
        case GameModeBattle: return data.ScreenWidth, data.ScreenHeight
    }

    return outsideWidth, outsideHeight
}

func MakeEngine(cache *lbx.LbxCache) *Engine {
    return &Engine{
        GameMode: GameModeUI,
        Player: MakePlayer(),
    }
}

func main() {
    log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)

    cache := lbx.AutoCache()

    engine := MakeEngine(cache)

    err := ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
