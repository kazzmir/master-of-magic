package main

import (
    "log"
    // "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/magicview"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    MagicScreen *magicview.MagicScreen
    ImageCache util.ImageCache
}

type NoDiplomacy struct{}
func (no *NoDiplomacy) EnterDiplomacy(player *playerlib.Player, enemy *playerlib.Player) {
}

func NewEngine() (*Engine, error) {
    cache := lbx.AutoCache()

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    player := playerlib.MakePlayer(setup.WizardCustom{
        Base: data.WizardHorus,
        Name: "Horus",
        Banner: data.BannerRed,
    }, true, 0, 0, nil, &playerlib.NoGlobalEnchantments{})

    player.CastingSkillPower = 280
    player.Gold = 234
    player.Mana = 981

    player.Wizard.ToggleRetort(data.RetortAlchemy, 2)
    player.GlobalEnchantments.Insert(data.EnchantmentNatureAwareness)
    player.GlobalEnchantments.Insert(data.EnchantmentDetectMagic)

    enemy1 := playerlib.MakePlayer(setup.WizardCustom{
        Base: data.WizardMerlin,
        Name: "Merlin",
        Banner: data.BannerPurple,
    }, false, 0, 0, nil, &playerlib.NoGlobalEnchantments{})

    enemy1.GlobalEnchantments.Insert(data.EnchantmentCrusade)
    enemy1.CastingSpell = allSpells.FindByName("Eldritch Weapon")

    player.AwarePlayer(enemy1)

    enemy2 := playerlib.MakePlayer(setup.WizardCustom{
        Base: data.WizardFreya,
        Name: "Freya",
        Banner: data.BannerGreen,
    }, false, 0, 0, nil, &playerlib.NoGlobalEnchantments{})

    enemy2.GlobalEnchantments.Insert(data.EnchantmentNaturesWrath)

    enemy3 := playerlib.MakePlayer(setup.WizardCustom{
        Base: data.WizardHorus,
        Name: "Horus",
        Banner: data.BannerYellow,
    }, false, 0, 0, nil, &playerlib.NoGlobalEnchantments{})

    enemy3.Defeated = true

    player.AwarePlayer(enemy3)

    enemy4 := playerlib.MakePlayer(setup.WizardCustom{
        Base: data.WizardJafar,
        Name: "Jafar",
        Banner: data.BannerBlue,
    }, false, 0, 0, nil, &playerlib.NoGlobalEnchantments{})

    enemy4.Defeated = true

    magicScreen := magicview.MakeMagicScreen(cache, player, []*playerlib.Player{enemy1, enemy2, enemy3, enemy4}, 100, &NoDiplomacy{})

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
    return scale.Scale2(data.ScreenWidth, data.ScreenHeight)
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    monitorWidth, _ := ebiten.Monitor().Size()

    size := monitorWidth / 390

    ebiten.SetWindowSize(data.ScreenWidth * size, data.ScreenHeight * size)

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
