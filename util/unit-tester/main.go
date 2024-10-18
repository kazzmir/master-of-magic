package main

import (
    "fmt"
    "strings"

    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

func main(){
    total := 0
    missingTotal := 0
    for _, unit := range units.AllUnits {
        total += 1
        var missing []string

        if unit.CombatLbxFile == "" {
            missing = append(missing, "CombatLbxFile")
        }

        if unit.MovementSpeed == 0 {
            missing = append(missing, "MovementSpeed")
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

        if unit.MeleeAttackPower == 0 && unit.RangedAttackPower == 0 {
            missing = append(missing, "MeleeAttackPower or RangedAttackPower")
        }

        if unit.Race == data.RaceNone {
            missing = append(missing, "Race")
        }

        if unit.Race != data.RaceHero && unit.Race != data.RaceFantastic && unit.ProductionCost == 0 {
            missing = append(missing, "ProductionCost")
        }

        if len(missing) > 0 {
            fmt.Printf("Unit %s %s is missing the following fields: %v\n", unit.Name, unit.Race, strings.Join(missing, ", "))
            missingTotal += 1
        }
    }

    fmt.Printf("Total: %v Missing fields: %v\n", total, missingTotal)

}
