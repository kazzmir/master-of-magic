package main

import (
    "log"
    "time"
    "cmp"
    "fmt"
    "sync"
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
    magiclog "github.com/kazzmir/master-of-magic/lib/log"
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

func runBattle(allSpells spellbook.Spells, attacker units.Unit, defender units.Unit, count int) combat.CombatState {
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

    for range count {
        attackingArmy.AddUnit(units.MakeOverworldUnitFromUnit(attacker, 1, 1, data.PlaneArcanus, attackingPlayer.Wizard.Banner, attackingPlayer.MakeExperienceInfo(), attackingPlayer.MakeUnitEnchantmentProvider()))
    }

    for range count {
        defendingArmy.AddUnit(units.MakeOverworldUnitFromUnit(defender, 1, 1, data.PlaneArcanus, defendingPlayer.Wizard.Banner, defendingPlayer.MakeExperienceInfo(), defendingPlayer.MakeUnitEnchantmentProvider()))
    }

    model := combat.MakeCombatModel(allSpells, defendingArmy, attackingArmy, combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, data.MagicNone, 0, 0, make(chan combat.CombatEvent, 10))

    return combat.Run(model)
}

func RunAll(allSpells spellbook.Spells) {
    type UnitKey struct {
        LbxFile string
        Index int
    }

    getKey := func(u units.Unit) UnitKey {
        return UnitKey{
            LbxFile: u.LbxFile,
            Index: u.Index,
        }
    }

    findUnit := func(key UnitKey) units.Unit {
        for _, u := range units.AllUnits {
            if u.LbxFile == key.LbxFile && u.Index == key.Index {
                return u
            }
        }

        return units.UnitNone
    }

    // useUnits := units.UnitsByRace(data.RaceOrc)
    useUnits := units.AllUnits

    unitResults := make(map[UnitKey]int)
    var lock sync.Mutex
    var group sync.WaitGroup

    for _, attacker := range useUnits {
        group.Add(1)
        key := getKey(attacker)

        go func(){
            defer group.Done()
            for _, defender := range useUnits {
                value := 0
                for range 5 {
                    switch runBattle(allSpells, attacker, defender, 2) {
                        case combat.CombatStateAttackerWin:
                            value += 1
                        case combat.CombatStateDefenderWin:
                            value -= 1
                    }

                }

                lock.Lock()
                unitResults[key] += value
                lock.Unlock()
            }

            fmt.Printf(".")

            // log.Printf("Attacker %v %v: score %v", attacker.Race, attacker.GetName(), unitResults[key])
        }()
    }

    group.Wait()
    fmt.Println()

    type Result struct {
        Unit units.Unit
        Score int
    }

    var results []Result
    for unit, score := range unitResults {
        results = append(results, Result{
            Unit: findUnit(unit),
            Score: score,
        })
    }

    slices.SortFunc(results, func(a, b Result) int {
        return cmp.Compare(a.Score, b.Score)
    })

    var strongestNormal Result

    for _, unit := range results {
        log.Printf("Unit %v %v: score %v", unit.Unit.Race, unit.Unit.GetName(), unit.Score)
        if unit.Unit.Race != data.RaceFantastic && unit.Score > strongestNormal.Score {
            strongestNormal = unit
        }
    }

    log.Printf("Strongest normal unit: %v %v with score %v", strongestNormal.Unit.Race, strongestNormal.Unit.GetName(), strongestNormal.Score)
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

    magiclog.SetLevel(magiclog.LogLevelDisabled)

    /*
    for range 100 {
        RunCombat1(allSpells)
    }
    */
    // RunCombat2(allSpells)
    RunAll(allSpells)

    memoryProfile, err := os.Create("profile.mem.combat-run")
    if err != nil {
        log.Printf("Error creating memory profile: %v", err)
    } else {
        defer memoryProfile.Close()
        pprof.WriteHeapProfile(memoryProfile)
    }
}
