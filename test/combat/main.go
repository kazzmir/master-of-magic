package main

import (
    "log"
    "strconv"
    "os"
    // "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/player"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    CombatScreen *combat.CombatScreen
    CombatEndScreen *combat.CombatEndScreen
    Coroutine *coroutine.Coroutine
}

func createWarlockArmy(player *player.Player) combat.Army {
    return combat.Army{
        Player: player,
        Units: []*combat.ArmyUnit{
            /*
            &combat.ArmyUnit{
                Unit: units.HighElfSpearmen,
                Facing: units.FacingDownRight,
                X: 12,
                Y: 10,
                Health: units.HighElfSpearmen.GetMaxHealth(),
            },
            */
            &combat.ArmyUnit{
                Unit: units.MakeOverworldUnitFromUnit(units.Warlocks, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()),
                Facing: units.FacingDownRight,
                X: 12,
                Y: 10,
            },
            /*
            &combat.ArmyUnit{
                Unit: units.HighElfSpearmen,
                Facing: units.FacingDownRight,
                X: 13,
                Y: 10,
                Health: units.HighElfSpearmen.GetMaxHealth(),
            },
            &combat.ArmyUnit{
                Unit: units.HighElfSpearmen,
                Facing: units.FacingDownRight,
                X: 14,
                Y: 10,
                Health: units.HighElfSpearmen.GetMaxHealth(),
            },
            */
        },
    }
}

func createWarlockArmyN(player *player.Player, count int) *combat.Army {
    army := combat.Army{
        Player: player,
    }

    for i := 0; i < count; i++ {
        army.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()))
    }

    return &army
}

func createHighMenBowmanArmyN(player *player.Player, count int) combat.Army {
    army := combat.Army{
        Player: player,
    }

    for i := 0; i < count; i++ {
        army.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()))
    }

    return army
}

func createHighMenBowmanArmy(player *player.Player) combat.Army {
    return combat.Army{
        Player: player,
        Units: []*combat.ArmyUnit{
            &combat.ArmyUnit{
                Unit: units.MakeOverworldUnitFromUnit(units.HighMenBowmen, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()),
                Facing: units.FacingDownRight,
                X: 12,
                Y: 10,
            },
        },
    }
}

func createGreatDrakeArmy(player *player.Player) *combat.Army{
    return &combat.Army{
        Player: player,
        Units: []*combat.ArmyUnit{
            &combat.ArmyUnit{
                Unit: units.MakeOverworldUnitFromUnit(units.GreatDrake, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()),
                Facing: units.FacingUpLeft,
                X: 10,
                Y: 17,
            },
            &combat.ArmyUnit{
                Unit: units.MakeOverworldUnitFromUnit(units.GreatDrake, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()),
                Facing: units.FacingUpLeft,
                X: 9,
                Y: 18,
            },
        },
    }
}

func createSettlerArmy(player *player.Player, count int) *combat.Army{
    var armyUnits []*combat.ArmyUnit

    for i := 0; i < count; i++ {
        armyUnits = append(armyUnits, &combat.ArmyUnit{
            Unit: units.MakeOverworldUnitFromUnit(units.HighElfSettlers, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()),
            Facing: units.FacingUpLeft,
            X: 10,
            Y: 17,
        })
    }

    return &combat.Army{
        Player: player,
        Units: armyUnits,
    }
}

func createHeroArmy(player *player.Player) *combat.Army{
    var armyUnits []*combat.ArmyUnit

    armyUnits = append(armyUnits, &combat.ArmyUnit{
        Unit: herolib.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroRakir, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()), herolib.HeroRakir, "bubba"),
    })

    armyUnits = append(armyUnits, &combat.ArmyUnit{
        Unit: herolib.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroWarrax, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()), herolib.HeroWarrax, "warby"),
    })

    return &combat.Army{
        Player: player,
        Units: armyUnits,
    }
}

func makeScenario1(cache *lbx.LbxCache) *combat.CombatScreen {
    defendingPlayer := player.Player{
        Wizard: setup.WizardCustom{
            Name: "Lair",
            Banner: data.BannerBrown,
        },
        Human: false,
    }

    // defendingArmy := createWarlockArmy(&defendingPlayer)
    defendingArmy := createHighMenBowmanArmyN(&defendingPlayer, 3)
    defendingArmy.LayoutUnits(combat.TeamDefender)

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
        CastingSkillPower: 10,
        Human: true,
    }

    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Fireball"))

    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Fireball"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Ice Bolt"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Fire Bolt"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Lightning Bolt"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Star Fires"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Psionic Blast"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Doom Bolt"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Warp Lightning"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Flame Strike"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Life Drain"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Dispel Evil"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Healing"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Holy Word"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Recall Hero"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Mass Healing"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Cracks Call"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Earth To Mud"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Web"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Banish"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Dispel Magic True"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Word of Recall"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Disintegrate"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Disrupt"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Magic Vortex"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Warp Wood"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Death Spell"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Word of Death"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Phantom Warriors"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Phantom Beast"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Earth Elemental"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Air Elemental"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Fire Elemental"))

    // attackingArmy := createGreatDrakeArmy(&attackingPlayer)
    attackingArmy := createWarlockArmyN(&attackingPlayer, 3)
    attackingArmy.LayoutUnits(combat.TeamAttacker)

    return combat.MakeCombatScreen(cache, &defendingArmy, attackingArmy, &attackingPlayer)
}

func makeScenario2(cache *lbx.LbxCache) *combat.CombatScreen {
    defendingPlayer := player.Player{
        Wizard: setup.WizardCustom{
            Name: "Lair",
            Banner: data.BannerBlue,
        },
        Human: false,
    }

    // defendingArmy := createWarlockArmy(&defendingPlayer)
    defendingArmy := createSettlerArmy(&defendingPlayer, 3)
    defendingArmy.LayoutUnits(combat.TeamDefender)

    /*
    allSpells, err := spellbook.ReadSpellsFromCache(cache)
    if err != nil {
        log.Printf("Unable to read spells: %v", err)
        allSpells = spellbook.Spells{}
    }
    */

    attackingPlayer := player.Player{
        Wizard: setup.WizardCustom{
            Name: "Merlin",
            Banner: data.BannerRed,
        },
        CastingSkillPower: 10,
        Human: true,
    }

    // attackingArmy := createGreatDrakeArmy(&attackingPlayer)
    attackingArmy := createSettlerArmy(&attackingPlayer, 3)
    attackingArmy.LayoutUnits(combat.TeamAttacker)

    return combat.MakeCombatScreen(cache, defendingArmy, attackingArmy, &attackingPlayer)
}

func makeScenario3(cache *lbx.LbxCache) *combat.CombatScreen {
    defendingPlayer := player.Player{
        Wizard: setup.WizardCustom{
            Name: "Lair",
            Banner: data.BannerBlue,
        },
        Human: false,
    }

    // defendingArmy := createWarlockArmy(&defendingPlayer)
    defendingArmy := createSettlerArmy(&defendingPlayer, 3)
    defendingArmy.LayoutUnits(combat.TeamDefender)

    /*
    allSpells, err := spellbook.ReadSpellsFromCache(cache)
    if err != nil {
        log.Printf("Unable to read spells: %v", err)
        allSpells = spellbook.Spells{}
    }
    */

    attackingPlayer := player.Player{
        Wizard: setup.WizardCustom{
            Name: "Merlin",
            Banner: data.BannerRed,
        },
        CastingSkillPower: 10,
        Human: true,
    }

    // attackingArmy := createGreatDrakeArmy(&attackingPlayer)
    attackingArmy := createHeroArmy(&attackingPlayer)
    attackingArmy.LayoutUnits(combat.TeamAttacker)

    return combat.MakeCombatScreen(cache, defendingArmy, attackingArmy, &attackingPlayer)
}

func NewEngine(scenario int) (*Engine, error) {
    cache := lbx.AutoCache()

    var combatScreen *combat.CombatScreen

    switch scenario {
        case 1: combatScreen = makeScenario1(cache)
        case 2: combatScreen = makeScenario2(cache)
        case 3: combatScreen = makeScenario3(cache)
        default: combatScreen = makeScenario1(cache)
    }

    run := func(yield coroutine.YieldFunc) error {
        for combatScreen.Update(yield) == combat.CombatStateRunning {
            yield()
        }

        return ebiten.Termination
    }

    return &Engine{
        LbxCache: cache,
        CombatScreen: combatScreen,
        CombatEndScreen: nil,
        Coroutine: coroutine.MakeCoroutine(run),
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

    if engine.CombatEndScreen != nil {
        switch engine.CombatEndScreen.Update() {
            case combat.CombatEndScreenRunning:
            case combat.CombatEndScreenDone:
                return ebiten.Termination
        }
    } else {
        if engine.Coroutine.Run() != nil {
            return ebiten.Termination
        }
        /*
        switch engine.CombatScreen.Update() {
            case combat.CombatStateRunning:
            case combat.CombatStateAttackerWin:
                log.Printf("Attackers win")
                engine.CombatEndScreen = combat.MakeCombatEndScreen(engine.LbxCache, engine.CombatScreen, true)
            case combat.CombatStateDefenderWin:
                log.Printf("Defenders win")
                engine.CombatEndScreen = combat.MakeCombatEndScreen(engine.LbxCache, engine.CombatScreen, false)
            case combat.CombatStateDone:
                return ebiten.Termination
        }
        */
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    if engine.CombatEndScreen != nil {
        engine.CombatEndScreen.Draw(screen)
    } else {
        engine.CombatScreen.Draw(screen)
    }
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
    ebiten.SetWindowTitle("combat screen")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
    ebiten.SetCursorMode(ebiten.CursorModeHidden)

    audio.Initialize()

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
