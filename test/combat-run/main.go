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

func RunAll(allSpells spellbook.Spells, unitsPerSide int) {
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

    // useUnits := units.UnitsByRace(data.RaceOrc)
    useUnits := units.AllUnits

    battles := 5

    log.Printf("Units per side: %v, battles per matchup: %v", unitsPerSide, battles)

    unitResults := make(map[UnitKey]map[UnitKey]int)
    var group sync.WaitGroup

    for _, attacker := range useUnits {
        group.Add(1)
        key := getKey(attacker)

        unitResults[key] = make(map[UnitKey]int)

        useMap := unitResults[key]

        go func(){
            defer group.Done()
            for _, defender := range useUnits {
                value := 0
                for range battles {
                    switch runBattle(allSpells, attacker, defender, unitsPerSide) {
                        case combat.CombatStateAttackerWin:
                            value += 1
                        case combat.CombatStateDefenderWin:
                            value -= 1
                    }

                }

                useMap[getKey(defender)] = value
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

    unitMap := make(map[UnitKey]units.Unit)
    for _, unit := range useUnits {
        unitMap[getKey(unit)] = unit
    }

    csvName := fmt.Sprintf("combat-results-%vv%v.csv", unitsPerSide, unitsPerSide)
    csvFile, err := os.Create(csvName)
    if err != nil {
        log.Fatalf("Failed to create CSV file: %v", err)
    }
    defer csvFile.Close()

    fmt.Fprintf(csvFile, "Unit,Total Score")

    for _, unit := range useUnits {
        fmt.Fprintf(csvFile, ",%v %v", unit.Race, unit.GetName())
    }
    fmt.Fprintln(csvFile)

    var results []Result
    for unitKey, scores := range unitResults {

        totalScore := 0

        unit := unitMap[unitKey]

        for _, score := range scores {
            totalScore += score
        }

        results = append(results, Result{
            Unit: unit,
            Score: totalScore,
        })
    }

    slices.SortFunc(results, func(a, b Result) int {
        return cmp.Compare(b.Score, a.Score)
    })

    // var strongestNormal Result

    for _, unit := range results {

        fmt.Fprintf(csvFile, "%v %v", unit.Unit.Race, unit.Unit.GetName())

        /*
        log.Printf("Unit %v %v: score %v", unit.Unit.Race, unit.Unit.GetName(), unit.Score)
        if unit.Unit.Race != data.RaceFantastic && unit.Score > strongestNormal.Score {
            strongestNormal = unit
        }
        */

        fmt.Fprintf(csvFile, ",%v", unit.Score)

        // write all the columns, which is the score against that opponent
        useMap := unitResults[getKey(unit.Unit)]
        for _, opponent := range useUnits {
            fmt.Fprintf(csvFile, ",%v", useMap[getKey(opponent)])
        }
        fmt.Fprintln(csvFile)
    }

    // log.Printf("Strongest normal unit: %v %v with score %v", strongestNormal.Unit.Race, strongestNormal.Unit.GetName(), strongestNormal.Score)

    log.Printf("Results written to %v", csvName)
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
    RunAll(allSpells, 4)

    memoryProfile, err := os.Create("profile.mem.combat-run")
    if err != nil {
        log.Printf("Error creating memory profile: %v", err)
    } else {
        defer memoryProfile.Close()
        pprof.WriteHeapProfile(memoryProfile)
    }
}
