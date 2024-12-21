package main

import (
    "fmt"
    "strings"
    "os"

    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

func main(){
    specifyRace := ""
    if len(os.Args) > 1 {
        specifyRace = os.Args[1]
    }

    total := 0
    missingTotal := 0
    for _, unit := range units.AllUnits {
        if specifyRace != "" && strings.ToLower(specifyRace) != strings.ToLower(unit.Race.String()) {
            continue
        }

        total += 1
        var missing []string

        if unit.CombatLbxFile == "" {
            missing = append(missing, "CombatLbxFile")
        }

        if unit.MovementSpeed == 0 {
            missing = append(missing, "MovementSpeed")
        }

        if unit.MovementSound == units.MovementSoundNone {
            missing = append(missing, "MovementSound")
        }

        if unit.AttackSound == units.AttackSoundNone {
            missing = append(missing, "AttackSound")
        }

        if unit.RangedAttackPower > 0 {
            if unit.RangeAttackSound == units.RangeAttackSoundNone {
                missing = append(missing, "RangeAttackSound")
            }

            if unit.RangeAttackIndex == 0 {
                missing = append(missing, "RangeAttackIndex")
            }
        }

        if unit.Defense == 0 {
            missing = append(missing, "Defense")
        }

        if unit.HitPoints == 0 {
            missing = append(missing, "HitPoints")
        }

        if unit.Count == 0 {
            missing = append(missing, "Count")
        }

        if unit.Resistance == 0 {
            missing = append(missing, "Resistance")
        }

        if unit.Name == "" {
            missing = append(missing, "Name")
        }

        if unit.Name != "Settlers" && unit.MeleeAttackPower == 0 && unit.RangedAttackPower == 0 {
            missing = append(missing, "MeleeAttackPower or RangedAttackPower")
        }

        if unit.Race == data.RaceNone {
            missing = append(missing, "Race")
        }

        if unit.Race != data.RaceHero && unit.Race != data.RaceFantastic && unit.ProductionCost == 0 {
            missing = append(missing, "ProductionCost")
        }

        if unit.Race != data.RaceHero && unit.Race != data.RaceFantastic && unit.UpkeepFood == 0 {
            missing = append(missing, "UpkeepFood")
        }

        if len(missing) > 0 {
            fmt.Printf("Unit %s %s (%v:%v) is missing the following fields: %v\n", unit.Name, unit.Race, unit.LbxFile, unit.Index, strings.Join(missing, ", "))
            missingTotal += 1
        }
    }

    fmt.Printf("Total: %v Missing fields: %v\n", total, missingTotal)

}
