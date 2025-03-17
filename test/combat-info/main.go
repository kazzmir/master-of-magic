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
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    // "github.com/kazzmir/master-of-magic/game/magic/hero"
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

    model := combat.MakeCombatModel(spellbook.Spells{}, defendingArmy, attackingArmy, combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, data.MagicNone, 0, 0, make(chan combat.CombatEvent))

    /*
    rakir := hero.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroRakir, 1, 1, data.PlaneArcanus, data.BannerRed, &ExperienceInfo{}), hero.HeroRakir, "Rakir")
    rakir.AddExperience(100)
    rakir.NaturalHeal(1)
    */

    // warlock := units.MakeOverworldUnitFromUnit(units.Warlocks, 1, 1, data.PlaneArcanus, data.BannerRed, &ExperienceInfo{})
    slingers := units.MakeOverworldUnitFromUnit(units.Slingers, 1, 1, data.PlaneArcanus, data.BannerRed, &ExperienceInfo{})

    // angel := units.MakeOverworldUnitFromUnit(units.ArchAngel, 1, 1, data.PlaneArcanus, data.BannerRed, &ExperienceInfo{})

    unit := &combat.ArmyUnit{
        Unit: slingers,
        Model: model,
    }

    unit.AddEnchantment(data.UnitEnchantmentBless)
    unit.AddEnchantment(data.UnitEnchantmentGiantStrength)
    unit.AddCurse(data.UnitCurseMindStorm)
    unit.AddEnchantment(data.UnitEnchantmentEndurance)
    unit.AddEnchantment(data.UnitEnchantmentLionHeart)
    unit.TakeDamage(5, combat.DamageNormal)

    // log.Printf("Base %v Defense %v Full Defense %v", unit.GetBaseDefense(), unit.GetDefense(), unit.GetFullDefense())

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
    return scale.Scale2(data.ScreenWidth, data.ScreenHeight)
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    scenario := 1
    if len(os.Args) > 1 {
        scenario, _ = strconv.Atoi(os.Args[1])
    }

    monitorWidth, _ := ebiten.Monitor().Size()
    size := monitorWidth / 390
    ebiten.SetWindowSize(data.ScreenWidth * size, data.ScreenHeight * size)

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
