package main

import (
    "log"
    // "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/magicview"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    MagicScreen *magicview.MagicScreen
    ImageCache util.ImageCache
}

func NewEngine() (*Engine, error) {
    cache := lbx.AutoCache()

    player := &playerlib.Player{
        CastingSkillPower: 28,
        PowerDistribution: playerlib.PowerDistribution{
            Mana: 1.0/3,
            Research: 1.0/3,
            Skill: 1.0/3,
        },
    }

    player.Gold = 234
    player.Mana = 981

    player.Wizard.ToggleAbility(setup.AbilityAlchemy, 2)

    enemy1 := &playerlib.Player{
        Human: false,
        Wizard: setup.WizardCustom{
            Base: data.WizardMerlin,
            Name: "Merlin",
            Banner: data.BannerPurple,
        },
    }

    magicScreen := magicview.MakeMagicScreen(cache, player, []*playerlib.Player{enemy1}, 100)

    return &Engine{
        LbxCache: cache,
        MagicScreen: magicScreen,
        ImageCache: util.MakeImageCache(cache),
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

    switch engine.MagicScreen.Update() {
        case magicview.MagicScreenStateRunning:
        case magicview.MagicScreenStateDone:
            return ebiten.Termination
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    engine.MagicScreen.Draw(screen)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return data.ScreenWidth, data.ScreenHeight
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(data.ScreenWidth * 5, data.ScreenHeight * 5)
    ebiten.SetWindowTitle("magic view")
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
