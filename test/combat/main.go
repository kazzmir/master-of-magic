package main

import (
    "log"
    "strconv"
    "os"
    "math"
    "image"
    // "image/color"
    "runtime/pprof"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
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

func createWarlockArmy(player *player.Player) *combat.Army {
    army := &combat.Army{Player: player}

    /*
            &combat.ArmyUnit{
                Unit: units.HighElfSpearmen,
                Facing: units.FacingDownRight,
                X: 12,
                Y: 10,
                Health: units.HighElfSpearmen.GetMaxHealth(),
            },
            */
            /*
    unit1 := &combat.ArmyUnit{
        Unit: units.MakeOverworldUnitFromUnit(units.Warlocks, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()),
        Facing: units.FacingDownRight,
        X: 12,
        Y: 10,
    },
    */
    unit1 := units.MakeOverworldUnitFromUnit(units.Warlocks, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo())
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

    army.AddUnit(unit1)
    return army
}

func createWarlockArmyN(player *player.Player, count int) *combat.Army {
    army := combat.Army{
        Player: player,
    }

    for i := 0; i < count; i++ {
        warlock := units.MakeOverworldUnitFromUnit(units.Warlocks, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo())
        if i == 0 {
            warlock.AddEnchantment(data.UnitEnchantmentGiantStrength)
        }
        army.AddUnit(warlock)
    }

    return &army
}

func createArmyN(player *player.Player, unit units.Unit, count int) *combat.Army {
    army := combat.Army{
        Player: player,
    }

    for i := 0; i < count; i++ {
        made := units.MakeOverworldUnitFromUnit(unit, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo())
        army.AddUnit(made)
    }

    return &army
}

func createLizardmenArmy(player *player.Player, count int) *combat.Army {
    army := combat.Army{
        Player: player,
    }

    for range count {
        army.AddUnit(units.MakeOverworldUnitFromUnit(units.LizardSwordsmen, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()))
    }

    return &army
}

func createHighMenBowmanArmyN(player *player.Player, count int) *combat.Army {
    army := combat.Army{
        Player: player,
    }

    for i := 0; i < count; i++ {
        army.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()))
    }

    return &army
}

func createHighMenBowmanArmy(player *player.Player) *combat.Army {
    army := combat.Army{Player: player}
    unit := units.MakeOverworldUnitFromUnit(units.HighMenBowmen, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo())
    army.AddUnit(unit)
    return &army
}

func createGreatDrakeArmy(player *player.Player) *combat.Army{
    army := combat.Army{Player: player}
    army.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()))
    army.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()))
    return &army
}

func createSettlerArmy(player *player.Player, count int) *combat.Army{
    army := combat.Army{Player: player}

    for range count {
        army.AddUnit(units.MakeOverworldUnitFromUnit(units.HighElfSettlers, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()))
    }

    return &army
}

func createArchAngelArmy(player *player.Player) *combat.Army {
    army := combat.Army{Player: player}

    army.AddUnit(units.MakeOverworldUnitFromUnit(units.ArchAngel, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()))

    return &army
}

func createDeathCreatureArmy(player *player.Player) *combat.Army {
    army := combat.Army{Player: player}

    army.AddUnit(units.MakeOverworldUnitFromUnit(units.DemonLord, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()))

    // not death, but whatever
    army.AddUnit(units.MakeOverworldUnitFromUnit(units.HellHounds, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()))

    return &army
}

func createHeroArmy(player *player.Player, cache *lbx.LbxCache) *combat.Army {
    army := combat.Army{Player: player}

    rakir := herolib.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroRakir, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()), herolib.HeroRakir, "bubba")

    rakir.Equipment[0] = &artifact.Artifact{
        Type: artifact.ArtifactTypeSword,
        Name: "Sword of Sharpness",
        Image: 2,
        Powers: []artifact.Power{
            artifact.Power{
                Type: artifact.PowerTypeAttack,
                Amount: 3,
                Name: "+3 Attack",
            },
        },
    }

    army.AddUnit(rakir)

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    item := artifact.Artifact{
        Type: artifact.ArtifactTypeStaff,
        Name: "Staff of Power",
        Image: 40,
        Powers: []artifact.Power{
            artifact.Power{
                Type: artifact.PowerTypeSpellCharges,
                Amount: 2,
                Spell: allSpells.FindByName("Ice Bolt"),
                // Spell: allSpells.FindByName("Healing"),
            },
        },
    }

    torin := herolib.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroTorin, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()), herolib.HeroTorin, "warby")
    torin.Equipment[0] = &item

    army.AddUnit(torin)

    return &army
}

func createWeakArmy(player *player.Player) *combat.Army {
    army := combat.Army{Player: player}

    for range 2 {
        army.AddUnit(units.MakeOverworldUnitFromUnit(units.HighElfSpearmen, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo()))
    }

    return &army
}

type BasicCatchment struct {
}

func (basic *BasicCatchment) GetCatchmentArea(x int, y int) map[image.Point]maplib.FullTile {
    return map[image.Point]maplib.FullTile{}
}

func (catchment *BasicCatchment) GetGoldBonus(x int, y int) int {
    return 0
}

func (basic *BasicCatchment) OnShore(x int, y int) bool {
    return false
}

func (basic *BasicCatchment) ByRiver(x int, y int) bool {
    return false
}

func (basic *BasicCatchment) TileDistance(x1 int, y1 int, x2 int, y2 int) int {
    dx := x1 - x2
    dy := y1 - y2
    return int(math.Sqrt(float64(dx * dx + dy * dy)))
}

func makeScenario1(cache *lbx.LbxCache) *combat.CombatScreen {
    defendingPlayer := player.MakePlayer(setup.WizardCustom{
            Name: "Lair",
            Banner: data.BannerBrown,
        }, false, 0, 0, nil)

    // defendingArmy := createWarlockArmy(&defendingPlayer)
    // defendingArmy := createHighMenBowmanArmyN(defendingPlayer, 3)
    defendingArmy := createLizardmenArmy(defendingPlayer, 3)
    defendingArmy.LayoutUnits(combat.TeamDefender)

    defendingArmy.GetUnits()[0].AddCurse(data.UnitCurseBlackSleep)
    defendingArmy.GetUnits()[1].AddCurse(data.UnitCurseConfusion)

    /*
    defendingArmy.AddEnchantment(data.CombatEnchantmentCounterMagic)
    defendingArmy.CounterMagic = 50
    */

    allSpells, err := spellbook.ReadSpellsFromCache(cache)
    if err != nil {
        log.Printf("Unable to read spells: %v", err)
        allSpells = spellbook.Spells{}
    }

    attackingPlayer := player.MakePlayer(setup.WizardCustom{
            Name: "Merlin",
            Banner: data.BannerGreen,
        }, true, 0, 0, nil)

    fortressCity := citylib.MakeCity("xyz", 10, 10, attackingPlayer.Wizard.Race, nil, &BasicCatchment{}, nil, attackingPlayer)
    fortressCity.Buildings.Insert(buildinglib.BuildingFortress)
    attackingPlayer.AddCity(fortressCity)

    attackingPlayer.CastingSkillPower = 10000
    attackingPlayer.Mana = 1000

    // attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Fireball"))

    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Fireball"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Ice Bolt"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Fire Bolt"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Lightning Bolt"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Star Fires"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Psionic Blast"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Doom Bolt"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Warp Lightning"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Flame Strike"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Life Drain"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Dispel Evil"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Healing"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Holy Word"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Recall Hero"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Mass Healing"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Cracks Call"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Earth To Mud"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Mass Invisibility"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Web"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Banish"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Dispel Magic True"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Word of Recall"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Disintegrate"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Disrupt"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Magic Vortex"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Warp Wood"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Death Spell"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Word of Death"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Phantom Warriors"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Phantom Beast"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Earth Elemental"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Air Elemental"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Fire Elemental"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Bless"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Weakness"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Darkness"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Flame Strike"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Vertigo"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Shatter"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Warp Creature"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Confusion"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Possession"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Call Chaos"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Raise Dead"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Animate Dead"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Heroism"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Holy Armor"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Holy Weapon"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Invulnerability"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Lionheart"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Righteousness"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("True Sight"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Elemental Armor"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Giant Strength"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Iron Skin"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Regeneration"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Resist Elements"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Stone Skin"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Flight"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Guardian Wind"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Haste"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Invisiblity"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Magic Immunity"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Resist Magic"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Spell Lock"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Eldritch Weapon"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Flame Blade"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Immolation"))

    // attackingArmy := createGreatDrakeArmy(&attackingPlayer)
    // attackingArmy := createWarlockArmyN(attackingPlayer, 3)
    attackingArmy := createArmyN(attackingPlayer, units.HighElfMagician, 4)
    attackingArmy.LayoutUnits(combat.TeamAttacker)

    for range 2 {
        attackingArmy.KillUnit(attackingArmy.GetUnits()[0])
    }

    /*
    attackingArmy.GetUnits()[0].AddEnchantment(data.UnitEnchantmentTrueSight)
    attackingArmy.GetUnits()[1].Unit.AddEnchantment(data.UnitEnchantmentLionHeart)
    */

    // attackingArmy.Units[0].AddCurse(data.UnitCurseConfusion)

    // return combat.MakeCombatScreen(cache, &defendingArmy, attackingArmy, attackingPlayer, combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, 10, 25)
    combatScreen := combat.MakeCombatScreen(cache, defendingArmy, attackingArmy, attackingPlayer, combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, 10, 25)

    // lame but we have to do this after the model has been created
    defendingArmy.GetUnits()[2].AddEnchantment(data.UnitEnchantmentInvisibility)
    // attackingArmy.GetUnits()[0].AddEnchantment(data.UnitEnchantmentInvisibility)

    // combatScreen.Model.AddGlobalEnchantment(data.CombatEnchantmentDarkness)
    return combatScreen
}

func makeScenario2(cache *lbx.LbxCache) *combat.CombatScreen {
    defendingPlayer := player.MakePlayer(setup.WizardCustom{
            Name: "Lair",
            Banner: data.BannerBlue,
        }, false, 0, 0, nil)

    // defendingArmy := createWarlockArmy(&defendingPlayer)
    defendingArmy := createSettlerArmy(defendingPlayer, 3)
    defendingArmy.LayoutUnits(combat.TeamDefender)

    /*
    allSpells, err := spellbook.ReadSpellsFromCache(cache)
    if err != nil {
        log.Printf("Unable to read spells: %v", err)
        allSpells = spellbook.Spells{}
    }
    */

    attackingPlayer := player.MakePlayer(setup.WizardCustom{
            Name: "Merlin",
            Banner: data.BannerRed,
        }, true, 0, 0, nil)

    attackingPlayer.CastingSkillPower = 10

    // attackingArmy := createGreatDrakeArmy(&attackingPlayer)
    attackingArmy := createSettlerArmy(attackingPlayer, 3)
    attackingArmy.LayoutUnits(combat.TeamAttacker)

    return combat.MakeCombatScreen(cache, defendingArmy, attackingArmy, attackingPlayer, combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, 0, 0)
}

func makeScenario3(cache *lbx.LbxCache) *combat.CombatScreen {
    defendingPlayer := player.MakePlayer(setup.WizardCustom{
            Name: "Lair",
            Banner: data.BannerBlue,
        }, false, 0, 0, nil)

    // defendingArmy := createWarlockArmy(&defendingPlayer)
    defendingArmy := createSettlerArmy(defendingPlayer, 3)
    defendingArmy.LayoutUnits(combat.TeamDefender)

    allSpells, err := spellbook.ReadSpellsFromCache(cache)
    if err != nil {
        log.Printf("Unable to read spells: %v", err)
        allSpells = spellbook.Spells{}
    }

    attackingPlayer := player.MakePlayer(setup.WizardCustom{
            Name: "Merlin",
            Banner: data.BannerRed,
        }, true, 0, 0, nil)

    attackingPlayer.CastingSkillPower = 10000
    attackingPlayer.Mana = 10000

    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Entangle"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Terror"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Wrack"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Call Lightning"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Flame Strike"))

    // attackingArmy := createGreatDrakeArmy(&attackingPlayer)
    attackingArmy := createHeroArmy(attackingPlayer, cache)
    attackingArmy.LayoutUnits(combat.TeamAttacker)

    return combat.MakeCombatScreen(cache, defendingArmy, attackingArmy, attackingPlayer, combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, 0, 0)
}

// fight in an unwalled city with a fortress
func makeScenario4(cache *lbx.LbxCache) *combat.CombatScreen {
    defendingPlayer := player.MakePlayer(setup.WizardCustom{
            Name: "Enemy",
            Banner: data.BannerBlue,
        }, false, 0, 0, nil)

    // defendingArmy := createWarlockArmy(defendingPlayer)
    // defendingArmy := createSettlerArmy(defendingPlayer, 3)
    defendingArmy := createLizardmenArmy(defendingPlayer, 3)
    defendingArmy.LayoutUnits(combat.TeamDefender)

    allSpells, err := spellbook.ReadSpellsFromCache(cache)
    if err != nil {
        log.Printf("Unable to read spells: %v", err)
        allSpells = spellbook.Spells{}
    }

    attackingPlayer := player.MakePlayer(setup.WizardCustom{
            Name: "Merlin",
            Banner: data.BannerRed,
            Race: data.RaceHighMen,
        }, true, 0, 0, nil)

    attackingPlayer.CastingSkillPower = 10000
    attackingPlayer.Mana = 1000
    attackingPlayer.TaxRate = fraction.Zero()

    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Fireball"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Disrupt"))

    attackingArmy := createGreatDrakeArmy(attackingPlayer)
    // attackingArmy := createWeakArmy(attackingPlayer)
    // attackingArmy := createHighMenBowmanArmy(attackingPlayer)
    // attackingArmy := createHeroArmy(attackingPlayer)
    attackingArmy.LayoutUnits(combat.TeamAttacker)

    city := citylib.MakeCity("xyz", 10, 10, attackingPlayer.Wizard.Race, nil, nil, nil, attackingPlayer)
    city.Buildings.Insert(buildinglib.BuildingFortress)
    city.Buildings.Insert(buildinglib.BuildingCityWalls)

    city.AddEnchantment(data.CityEnchantmentWallOfFire, defendingPlayer.Wizard.Banner)
    // city.AddEnchantment(data.CityEnchantmentWallOfDarkness, defendingPlayer.Wizard.Banner)

    return combat.MakeCombatScreen(cache, defendingArmy, attackingArmy, attackingPlayer, combat.CombatLandscapeGrass, data.PlaneMyrror, combat.ZoneType{City: city}, 0, 0)
}

// fight in a tower of wizardy
func makeScenario5(cache *lbx.LbxCache) *combat.CombatScreen {
    defendingPlayer := player.MakePlayer(setup.WizardCustom{
            Name: "Enemy",
            Banner: data.BannerBlue,
        }, false, 0, 0, nil)

    // defendingArmy := createWarlockArmy(&defendingPlayer)
    defendingArmy := createSettlerArmy(defendingPlayer, 3)
    defendingArmy.LayoutUnits(combat.TeamDefender)

    /*
    allSpells, err := spellbook.ReadSpellsFromCache(cache)
    if err != nil {
        log.Printf("Unable to read spells: %v", err)
        allSpells = spellbook.Spells{}
    }
    */

    attackingPlayer := player.MakePlayer(setup.WizardCustom{
            Name: "Merlin",
            Banner: data.BannerRed,
            Race: data.RaceHighMen,
        }, true, 0, 0, nil)

    attackingPlayer.CastingSkillPower = 10
    attackingPlayer.TaxRate = fraction.Zero()

    // attackingArmy := createGreatDrakeArmy(&attackingPlayer)
    attackingArmy := createHeroArmy(attackingPlayer, cache)
    attackingArmy.LayoutUnits(combat.TeamAttacker)

    city := citylib.MakeCity("xyz", 10, 10, attackingPlayer.Wizard.Race, nil, nil, nil, attackingPlayer)
    city.Buildings.Insert(buildinglib.BuildingFortress)

    return combat.MakeCombatScreen(cache, defendingArmy, attackingArmy, attackingPlayer, combat.CombatLandscapeMountain, data.PlaneArcanus, combat.ZoneType{ChaosNode: true}, 0, 0)
}

func makeScenario6(cache *lbx.LbxCache) *combat.CombatScreen {
    defendingPlayer := player.MakePlayer(setup.WizardCustom{
            Name: "Enemy",
            Banner: data.BannerBlue,
        }, false, 0, 0, nil)

    // defendingArmy := createWarlockArmy(&defendingPlayer)
    defendingArmy := createSettlerArmy(defendingPlayer, 3)
    defendingArmy.LayoutUnits(combat.TeamDefender)

    /*
    allSpells, err := spellbook.ReadSpellsFromCache(cache)
    if err != nil {
        log.Printf("Unable to read spells: %v", err)
        allSpells = spellbook.Spells{}
    }
    */

    attackingPlayer := player.MakePlayer(setup.WizardCustom{
            Name: "Merlin",
            Banner: data.BannerRed,
            Race: data.RaceHighMen,
        }, true, 0, 0, nil)

    attackingPlayer.CastingSkillPower = 10
    attackingPlayer.TaxRate = fraction.Zero()

    // attackingArmy := createGreatDrakeArmy(&attackingPlayer)
    attackingArmy := createArchAngelArmy(attackingPlayer)
    attackingArmy.LayoutUnits(combat.TeamAttacker)

    city := citylib.MakeCity("xyz", 10, 10, attackingPlayer.Wizard.Race, nil, nil, nil, attackingPlayer)
    city.Buildings.Insert(buildinglib.BuildingFortress)

    return combat.MakeCombatScreen(cache, defendingArmy, attackingArmy, attackingPlayer, combat.CombatLandscapeMountain, data.PlaneArcanus, combat.ZoneType{ChaosNode: true}, 0, 0)
}

// combat on water
func makeScenario7(cache *lbx.LbxCache) *combat.CombatScreen {
    defendingPlayer := player.MakePlayer(setup.WizardCustom{
            Name: "Enemy",
            Banner: data.BannerBlue,
        }, false, 0, 0, nil)

    // defendingArmy := createWarlockArmy(&defendingPlayer)
    defendingArmy := createSettlerArmy(defendingPlayer, 3)
    defendingArmy.LayoutUnits(combat.TeamDefender)

    /*
    allSpells, err := spellbook.ReadSpellsFromCache(cache)
    if err != nil {
        log.Printf("Unable to read spells: %v", err)
        allSpells = spellbook.Spells{}
    }
    */

    attackingPlayer := player.MakePlayer(setup.WizardCustom{
            Name: "Merlin",
            Banner: data.BannerRed,
            Race: data.RaceHighMen,
        }, true, 0, 0, nil)

    attackingPlayer.CastingSkillPower = 10
    attackingPlayer.TaxRate = fraction.Zero()

    // attackingArmy := createGreatDrakeArmy(&attackingPlayer)
    attackingArmy := createArchAngelArmy(attackingPlayer)
    attackingArmy.LayoutUnits(combat.TeamAttacker)

    city := citylib.MakeCity("xyz", 10, 10, attackingPlayer.Wizard.Race, nil, nil, nil, attackingPlayer)
    city.Buildings.Insert(buildinglib.BuildingFortress)

    return combat.MakeCombatScreen(cache, defendingArmy, attackingArmy, attackingPlayer, combat.CombatLandscapeWater, data.PlaneArcanus, combat.ZoneType{}, 0, 0)
}

// life fantastic creatures vs death
func makeScenario8(cache *lbx.LbxCache) *combat.CombatScreen {
    defendingPlayer := player.MakePlayer(setup.WizardCustom{
            Name: "Enemy",
            Banner: data.BannerBlue,
        }, false, 0, 0, nil)

    // defendingArmy := createWarlockArmy(&defendingPlayer)
    defendingArmy := createDeathCreatureArmy(defendingPlayer)
    defendingArmy.LayoutUnits(combat.TeamDefender)

    /*
    allSpells, err := spellbook.ReadSpellsFromCache(cache)
    if err != nil {
        log.Printf("Unable to read spells: %v", err)
        allSpells = spellbook.Spells{}
    }
    */

    allSpells, _ := spellbook.ReadSpellsFromCache(cache)

    attackingPlayer := player.MakePlayer(setup.WizardCustom{
            Name: "Merlin",
            Banner: data.BannerRed,
            Race: data.RaceHighMen,
        }, true, 0, 0, nil)

    attackingPlayer.TaxRate = fraction.Zero()

    spells := []string{"High Prayer", "Prayer", "True Light", "Call Lightning", "Entangle",
                       "Blur", "Counter Magic", "Mass Invisibility", "Metal Fires", "Warp Reality",
                       "Black Prayer", "Darkness", "Mana Leak", "Terror", "Wrack",
                       "Disenchant Area", "Disenchant True",
                       "Creature Binding", "Mind Storm",
                       "Fire Bolt", "Ice Bolt", "Star Fires", "Dispel Evil", "Life Drain",
                       "Holy Word", "Cracks Call", "Banish", "Disintegrate", "Warp Wood", "Death Spell",
                       "Word of Death", "Dispel Magic True", "Web", "Petrify", "Raise Dead",
                   }


    for _, spellName := range spells {
        spell := allSpells.FindByName(spellName)
        if spell.Invalid() {
            log.Printf("Unknown spell: %v", spellName)
        } else {
            attackingPlayer.KnownSpells.AddSpell(spell)
        }
    }

    attackingPlayer.Mana = 10000
    attackingPlayer.CastingSkillPower = 10000

    attackingArmy := createArchAngelArmy(attackingPlayer)

    attackingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.GiantSpiders, 1, 1, data.PlaneArcanus, attackingPlayer.Wizard.Banner, attackingPlayer.MakeExperienceInfo()))

    attackingArmy.LayoutUnits(combat.TeamAttacker)

    attackingArmy.AddEnchantment(data.CombatEnchantmentWrack)

    for _, unit := range attackingArmy.GetUnits() {
        unit.AddCurse(data.UnitCurseVertigo)
        unit.AddCurse(data.UnitCurseShatter)
    }

    defendingArmy.AddEnchantment(data.CombatEnchantmentEntangle)

    for _, unit := range defendingArmy.GetUnits() {
        unit.AddEnchantment(data.UnitEnchantmentGiantStrength)
        unit.AddEnchantment(data.UnitEnchantmentHolyArmor)
    }

    city := citylib.MakeCity("xyz", 10, 10, attackingPlayer.Wizard.Race, nil, nil, nil, attackingPlayer)
    city.Buildings.Insert(buildinglib.BuildingFortress)

    return combat.MakeCombatScreen(cache, defendingArmy, attackingArmy, attackingPlayer, combat.CombatLandscapeMountain, data.PlaneArcanus, combat.ZoneType{ChaosNode: true}, 0, 0)
}

func NewEngine(scenario int) (*Engine, error) {
    cache := lbx.AutoCache()

    var combatScreen *combat.CombatScreen

    switch scenario {
        case 1: combatScreen = makeScenario1(cache)
        case 2: combatScreen = makeScenario2(cache)
        case 3: combatScreen = makeScenario3(cache)
        case 4: combatScreen = makeScenario4(cache)
        case 5: combatScreen = makeScenario5(cache)
        case 6: combatScreen = makeScenario6(cache)
        case 7: combatScreen = makeScenario7(cache)
        case 8: combatScreen = makeScenario8(cache)
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

    inputmanager.Update()

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

    mouse.Mouse.Draw(screen)
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

    profile, err := os.Create("profile.cpu.combat")
    if err != nil {
        log.Printf("Error creating profile: %v", err)
    } else {
        defer profile.Close()
        pprof.StartCPUProfile(profile)
        defer pprof.StopCPUProfile()
    }

    monitorWidth, _ := ebiten.Monitor().Size()
    size := monitorWidth / 390
    ebiten.SetWindowSize(data.ScreenWidth * size, data.ScreenHeight * size)

    ebiten.SetWindowTitle("combat screen")
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
