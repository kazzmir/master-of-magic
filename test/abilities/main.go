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
    "github.com/kazzmir/master-of-magic/game/magic/util"
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

func allAbilities() []data.Ability {
    abilities := []data.AbilityType{
        data.AbilityArmorPiercing,
        data.AbilityCauseFear,
        data.AbilityColdImmunity,
        data.AbilityConstruction,
        data.AbilityCreateOutpost,
        data.AbilityCreateUndead,
        data.AbilityDeathGaze,
        data.AbilityDeathImmunity,
        data.AbilityDispelEvil,
        data.AbilityDoomBoltSpell,
        data.AbilityDoomGaze,
        data.AbilityFireballSpell,
        data.AbilityFireBreath,
        data.AbilityFireImmunity,
        data.AbilityFirstStrike,
        data.AbilityForester,
        data.AbilityHealer,
        data.AbilityHealingSpell,
        data.AbilityHolyBonus,
        data.AbilityIllusion,
        data.AbilityIllusionsImmunity,
        data.AbilityImmolation,
        data.AbilityInvisibility,
        data.AbilityLargeShield,
        data.AbilityLifeSteal,
        data.AbilityLightningBreath,
        data.AbilityLongRange,
        data.AbilityMagicImmunity,
        data.AbilityMeld,
        data.AbilityMerging,
        data.AbilityMissileImmunity,
        data.AbilityMountaineer,
        data.AbilityNegateFirstStrike,
        data.AbilityNonCorporeal,
        data.AbilityPathfinding,
        data.AbilityPlaneShift,
        data.AbilityPoisonImmunity,
        data.AbilityPoisonTouch,
        data.AbilityPurify,
        data.AbilityRegeneration,
        data.AbilityResistanceToAll,
        data.AbilityScouting,
        data.AbilityStoningGaze,
        data.AbilityStoningImmunity,
        data.AbilityStoningTouch,
        data.AbilitySummonDemons,
        data.AbilityToHit,
        data.AbilityTransport,
        data.AbilityTeleporting,
        data.AbilityThrown,
        data.AbilityWallCrusher,
        data.AbilityWeaponImmunity,
        data.AbilityWebSpell,
        data.AbilityWindWalking,

        // hero abilities
        data.AbilityAgility,
        data.AbilitySuperAgility,
        data.AbilityArcanePower,
        data.AbilitySuperArcanePower,
        data.AbilityArmsmaster,
        data.AbilitySuperArmsmaster,
        data.AbilityBlademaster,
        data.AbilitySuperBlademaster,
        data.AbilityCaster,
        data.AbilityCharmed,
        data.AbilityConstitution,
        data.AbilitySuperConstitution,
        data.AbilityLeadership,
        data.AbilitySuperLeadership,
        data.AbilityLegendary,
        data.AbilitySuperLegendary,
        data.AbilityLucky,
        data.AbilityMight,
        data.AbilitySuperMight,
        data.AbilityNoble,
        data.AbilityPrayermaster,
        data.AbilitySuperPrayermaster,
        data.AbilitySage,
        data.AbilitySuperSage,
    }

    var out []data.Ability

    for _, ability := range abilities {
        out = append(out, data.MakeAbility(ability))
    }

    return out
}

func NewEngine(scenario int) (*Engine, error) {
    cache := lbx.AutoCache()

    imageCache := util.MakeImageCache(cache)
    normalMouse, err := mouselib.GetMouseNormal(cache, &imageCache)
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

            baseUnit.Abilities = allAbilities()

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
