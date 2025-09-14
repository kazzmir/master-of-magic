package combat

import (
    "slices"
    "cmp"
    "image"

    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/lib/fraction"
)

type AIUnitActionsInterface interface {
    RangeAttack(attacker *ArmyUnit, defender *ArmyUnit)
    MeleeAttack(attacker *ArmyUnit, defender *ArmyUnit)
    MoveUnit(unit *ArmyUnit, path pathfinding.Path)
}

func doAI(model *CombatModel, aiActions AIUnitActionsInterface, aiUnit *ArmyUnit) {
    // aiArmy := combat.GetArmy(combat.SelectedUnit)
    otherArmy := model.GetOtherArmy(aiUnit)
    if aiUnit.ConfusionAction == ConfusionActionEnemyControl {
        otherArmy = model.GetArmy(aiUnit)
    }

    // FIXME: cast a spell if the unit has mana (caster ability)

    // try a ranged attack first
    if aiUnit.CanRangeAttack() {
        candidates := slices.Clone(otherArmy.units)
        slices.SortFunc(candidates, func (a *ArmyUnit, b *ArmyUnit) int {
            return cmp.Compare(computeTileDistance(aiUnit.X, aiUnit.Y, a.X, a.Y), computeTileDistance(aiUnit.X, aiUnit.Y, b.X, b.Y))
        })

        for _, candidate := range candidates {
           if model.withinArrowRange(aiUnit, candidate) && model.canRangeAttack(aiUnit, candidate) {
               aiActions.RangeAttack(aiUnit, candidate)
               return
           }
        }
    }

    for _, unit := range otherArmy.units {
        if model.withinMeleeRange(aiUnit, unit) && model.canMeleeAttack(aiUnit, unit) {
            aiActions.MeleeAttack(aiUnit, unit)
            return
        }
    }

    aiUnit.Paths = make(map[image.Point]pathfinding.Path)

    // if the selected unit has ranged attacks, then try to use that
    // otherwise, if in melee range of some enemy then attack them
    // otherwise walk towards the enemy

    paths := make(map[*ArmyUnit]pathfinding.Path)

    getPath := func (unit *ArmyUnit) pathfinding.Path {
        path, found := paths[unit]
        if !found {
            model.Tiles[unit.Y][unit.X].Unit = nil
            var ok bool
            path, ok = model.computePath(aiUnit.X, aiUnit.Y, unit.X, unit.Y, aiUnit.CanTraverseWall(), aiUnit.IsFlying())
            model.Tiles[unit.Y][unit.X].Unit = unit
            if ok {
                paths[unit] = path
            } else {
                paths[unit] = nil
            }
        }

        return path
    }

    filterReachable := func (units []*ArmyUnit) []*ArmyUnit {
        var out []*ArmyUnit

        aiInWall := model.InsideAnyWall(aiUnit.X, aiUnit.Y)

        for _, unit := range units {
            // skip enemies that we can't melee anyway
            if !model.canMeleeAttack(aiUnit, unit) {
                continue
            }

            enemyInWall := model.InsideAnyWall(unit.X, unit.Y)

            // if the unit is inside a wall (fire/darkness/brick) but the target is outside, then don't move
            if aiUnit.Team == TeamDefender && aiInWall && !enemyInWall {
                continue
            }

            path := getPath(unit)
            if len(path) > 0 {
                out = append(out, unit)
            }
        }
        return out
    }

    // should filter by enemies that we can attack, so non-flyers do not move toward flyers
    candidates := filterReachable(slices.Clone(otherArmy.units))

    slices.SortFunc(candidates, func (a *ArmyUnit, b *ArmyUnit) int {
        aPath := getPath(a)
        bPath := getPath(b)

        return cmp.Compare(len(aPath), len(bPath))
    })


    // find a path to some enemy
    for _, closestEnemy := range candidates {
        // pretend that there is no unit at the tile. this is a sin of the highest order

        path := getPath(closestEnemy)

        // a path of length 2 contains the position of the aiUnit and the position of the enemy, so they are right next to each other
        if len(path) == 2 && model.canMeleeAttack(aiUnit, closestEnemy) {
            aiActions.MeleeAttack(aiUnit, closestEnemy)
            return
        } else if len(path) > 2 {
            // ignore path[0], thats where we are now. also ignore the last element, since we can't move onto the enemy

            last := path[len(path)-1]
            if last.X == closestEnemy.X && last.Y == closestEnemy.Y {
                path = path[:len(path)-1]
            }

            lastIndex := 0
            for lastIndex < len(path) {
                lastIndex += 1
                if !aiUnit.CanFollowPath(path[0:lastIndex]) {
                    lastIndex -= 1
                    break
                }
            }

            if lastIndex >= 1 && lastIndex <= len(path) {
                aiActions.MoveUnit(aiUnit, path[1:lastIndex])
                return
            }
        }
    }

    // no enemy to move towards, then possibly move towards gate
    if aiUnit.Team == TeamDefender && model.InsideCityWall(aiUnit.X, aiUnit.Y) {
        // if inside a city wall, then move towards the gate
        gateX, gateY := model.GetCityGateCoordinates()
        if gateX != -1 && gateY != -1 {
            path, ok := model.computePath(aiUnit.X, aiUnit.Y, gateX, gateY, aiUnit.CanTraverseWall(), aiUnit.IsFlying())
            if ok && len(path) > 1 && aiUnit.CanFollowPath(path) {
                aiActions.MoveUnit(aiUnit, path[1:])
                return
            }
        }
    }

    // didn't make a choice, just exhaust moves left
    aiUnit.MovesLeft = fraction.FromInt(0)
}
