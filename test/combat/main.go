package main

import (
    "log"
    // "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    CombatScreen *combat.CombatScreen
}

func NewEngine() (*Engine, error) {
    cache := lbx.AutoCache()

    defendingPlayer := player.Player{
        Wizard: setup.WizardCustom{
            Name: "Lair",
            Banner: data.BannerBrown,
        },
    }

    defendingArmy := combat.Army{
        Player: &defendingPlayer,
        Units: []*combat.ArmyUnit{
            &combat.ArmyUnit{
                Unit: units.HighElfSpearmen,
                Facing: units.FacingDownRight,
                X: 12,
                Y: 10,
                Health: 10,
            },
            /*
            &combat.ArmyUnit{
                Unit: units.HighElfSpearmen,
                Facing: units.FacingDownRight,
                X: 13,
                Y: 10,
                Health: 10,
            },
            &combat.ArmyUnit{
                Unit: units.HighElfSpearmen,
                Facing: units.FacingDownRight,
                X: 14,
                Y: 10,
                Health: 10,
            },
            */
        },
    }

    allSpells, err := spellbook.ReadSpellsFromCache(cache)
    if err != nil {
        log.Printf("Unable to read spells: %v", err)
        allSpells = spellbook.Spells{}
    }

    attackingPlayer := player.Player{
        Wizard: setup.WizardCustom{
            Name: "Merlin",
            Banner: data.BannerGreen,
        },
        CastingSkill: 10,
    }

    attackingPlayer.Spells.AddSpell(allSpells.FindByName("Fireball"))

    defendingPlayer.Spells.AddSpell(allSpells.FindByName("Fireball"))
    defendingPlayer.Spells.AddSpell(allSpells.FindByName("Ice Bolt"))
    defendingPlayer.Spells.AddSpell(allSpells.FindByName("Star Fires"))
    defendingPlayer.Spells.AddSpell(allSpells.FindByName("Psionic Blast"))
    defendingPlayer.Spells.AddSpell(allSpells.FindByName("Doom Bolt"))

    attackingArmy := combat.Army{
        Player: &attackingPlayer,
        Units: []*combat.ArmyUnit{
            &combat.ArmyUnit{
                Unit: units.GreatDrake,
                Facing: units.FacingUpLeft,
                X: 11,
                Y: 19,
                Health: 10,
            },
        },
    }

    return &Engine{
        LbxCache: cache,
        CombatScreen: combat.MakeCombatScreen(cache, &defendingArmy, &attackingArmy, &defendingPlayer),
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
    ebiten.SetCursorMode(ebiten.CursorModeHidden)

    audio.Initialize()

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
