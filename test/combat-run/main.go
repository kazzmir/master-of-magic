package main

import (
    "log"
    "time"
    "bytes"
    "os"
    "runtime/pprof"
    "slices"
    "math/rand/v2"
    _ "embed"

    "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/lib/lbx"
)

//go:embed spelldat.lbx
var spellDataLbx []byte

func LoadSpellData() (spellbook.Spells, error) {
    buffer := bytes.NewReader(spellDataLbx)
    spellLbx, err := lbx.ReadLbx(buffer)
    if err != nil {
        return spellbook.Spells{}, err
    }
    return spellbook.ReadSpells(&spellLbx, 0)
}

type noGlobalEnchantments struct {
}

func (*noGlobalEnchantments) HasEnchantment(enchantment data.Enchantment) bool {
    return false
}

func (*noGlobalEnchantments) HasRivalEnchantment(player *player.Player, enchantment data.Enchantment) bool {
    return false
}

func RunCombat1(allSpells spellbook.Spells) {
    defendingPlayer := player.MakePlayer(setup.WizardCustom{
        Name: "AI-1",
        Banner: data.BannerBrown,
    }, false, 0, 0, nil, &noGlobalEnchantments{})

    attackingPlayer := player.MakePlayer(setup.WizardCustom{
        Name: "AI-2",
        Banner: data.BannerRed,
    }, false, 0, 0, nil, &noGlobalEnchantments{})

    attackingArmy := &combat.Army{
        Player: attackingPlayer,
    }

    defendingArmy := &combat.Army{
        Player: defendingPlayer,
    }

    for range 9 {
        attackingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.LizardSwordsmen, 1, 1, data.PlaneArcanus, attackingPlayer.Wizard.Banner, attackingPlayer.MakeExperienceInfo(), attackingPlayer.MakeUnitEnchantmentProvider()))
    }

    for range 9 {
        // defendingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.BarbarianSpearmen, 1, 1, data.PlaneArcanus, defendingPlayer.Wizard.Banner, defendingPlayer.MakeExperienceInfo(), defendingPlayer.MakeUnitEnchantmentProvider()))
        defendingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.BeastmenPriest, 1, 1, data.PlaneArcanus, defendingPlayer.Wizard.Banner, defendingPlayer.MakeExperienceInfo(), defendingPlayer.MakeUnitEnchantmentProvider()))
    }

    model := combat.MakeCombatModel(allSpells, defendingArmy, attackingArmy, combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, data.MagicNone, 0, 0, make(chan combat.CombatEvent, 10))

    start := time.Now()
    state := combat.Run(model)
    end := time.Now()
    log.Printf("Combat simulation took %v", end.Sub(start))
    log.Printf("Final state: %+v", state)
}

func RunCombat2(allSpells spellbook.Spells) {
    defendingPlayer := player.MakePlayer(setup.WizardCustom{
        Name: "AI-1",
        Banner: data.BannerBrown,
    }, false, 0, 0, nil, &noGlobalEnchantments{})

    attackingPlayer := player.MakePlayer(setup.WizardCustom{
        Name: "AI-2",
        Banner: data.BannerRed,
    }, false, 0, 0, nil, &noGlobalEnchantments{})

    attackingArmy := &combat.Army{
        Player: attackingPlayer,
    }

    defendingArmy := &combat.Army{
        Player: defendingPlayer,
    }

    randomUnit := func() units.Unit {
        choices := slices.Clone(units.AllUnits)
        choices = slices.DeleteFunc(choices, func(u units.Unit) bool {
            if u.IsSettlers() {
                return true
            }
            return false
        })
        return choices[rand.N(len(choices))]
    }

    for range 9 {
        unit := attackingArmy.AddUnit(units.MakeOverworldUnitFromUnit(randomUnit(), 1, 1, data.PlaneArcanus, attackingPlayer.Wizard.Banner, attackingPlayer.MakeExperienceInfo(), attackingPlayer.MakeUnitEnchantmentProvider()))
        log.Printf("Attacker: %v %v", unit.Unit.GetRace(), unit.Unit.GetName())
    }

    for range 9 {
        // defendingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.BarbarianSpearmen, 1, 1, data.PlaneArcanus, defendingPlayer.Wizard.Banner, defendingPlayer.MakeExperienceInfo(), defendingPlayer.MakeUnitEnchantmentProvider()))
        unit := defendingArmy.AddUnit(units.MakeOverworldUnitFromUnit(randomUnit(), 1, 1, data.PlaneArcanus, defendingPlayer.Wizard.Banner, defendingPlayer.MakeExperienceInfo(), defendingPlayer.MakeUnitEnchantmentProvider()))
        log.Printf("Defender: %v %v", unit.Unit.GetRace(), unit.Unit.GetName())
    }

    model := combat.MakeCombatModel(allSpells, defendingArmy, attackingArmy, combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, data.MagicNone, 0, 0, make(chan combat.CombatEvent, 10))

    start := time.Now()
    state := combat.Run(model)
    end := time.Now()
    log.Printf("Combat simulation took %v", end.Sub(start))
    log.Printf("Final state: %+v", state)
}

func RunCombat3(allSpells spellbook.Spells) {
    defendingPlayer := player.MakePlayer(setup.WizardCustom{
        Name: "AI-1",
        Banner: data.BannerBrown,
    }, false, 0, 0, nil, &noGlobalEnchantments{})

    attackingPlayer := player.MakePlayer(setup.WizardCustom{
        Name: "AI-2",
        Banner: data.BannerRed,
    }, false, 0, 0, nil, &noGlobalEnchantments{})

    attackingArmy := &combat.Army{
        Player: attackingPlayer,
    }

    defendingArmy := &combat.Army{
        Player: defendingPlayer,
    }

    attackingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.GnollSwordsmen, 1, 1, data.PlaneArcanus, attackingPlayer.Wizard.Banner, attackingPlayer.MakeExperienceInfo(), attackingPlayer.MakeUnitEnchantmentProvider()))

    defendingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.OrcSpearmen, 1, 1, data.PlaneArcanus, defendingPlayer.Wizard.Banner, defendingPlayer.MakeExperienceInfo(), defendingPlayer.MakeUnitEnchantmentProvider()))

    model := combat.MakeCombatModel(allSpells, defendingArmy, attackingArmy, combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, data.MagicNone, 0, 0, make(chan combat.CombatEvent, 10))

    start := time.Now()
    state := combat.Run(model)
    end := time.Now()
    log.Printf("Combat simulation took %v", end.Sub(start))
    log.Printf("Final state: %+v", state)
}

func main() {
    log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)

    allSpells, err := LoadSpellData()
    if err != nil {
        log.Fatalf("Failed to load spell data: %v", err)
    }

    profile, err := os.Create("profile.cpu.combat-run")
    if err != nil {
        log.Printf("Error creating profile: %v", err)
    } else {
        defer profile.Close()
        pprof.StartCPUProfile(profile)
        defer pprof.StopCPUProfile()
    }

    /*
    for range 100 {
        RunCombat1(allSpells)
    }
    */
    RunCombat2(allSpells)

    memoryProfile, err := os.Create("profile.mem.combat-run")
    if err != nil {
        log.Printf("Error creating memory profile: %v", err)
    } else {
        defer memoryProfile.Close()
        pprof.WriteHeapProfile(memoryProfile)
    }
}
