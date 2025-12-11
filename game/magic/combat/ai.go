package combat

import (
    "math/rand/v2"
    "slices"
    "cmp"
    "image"

    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/log"
)

type AIUnitActionsInterface interface {
    RangeAttack(attacker *ArmyUnit, defender RangeTarget)
    MeleeAttack(attacker *ArmyUnit, defender *ArmyUnit)
    MeleeAttackWall(attacker *ArmyUnit, x int, y int)
    MoveUnit(unit *ArmyUnit, path pathfinding.Path)
    Teleport(unit *ArmyUnit, x, y int, merge bool)
    DoProjectiles()
}

func doAI(model *CombatModel, spellSystem SpellSystem, aiActions AIUnitActionsInterface, aiUnit *ArmyUnit) {
    // aiArmy := combat.GetArmy(combat.SelectedUnit)
    army := model.GetArmy(aiUnit)
    otherArmy := model.GetOtherArmy(aiUnit)

    isConfused := false
    if aiUnit.ConfusionAction == ConfusionActionEnemyControl {
        otherArmy = model.GetArmy(aiUnit)
        isConfused = true
    }

    // for now, disallow confused enemies from casting spells
    if !isConfused && aiUnit.CanCast() && rand.N(100) < 20 {
        for spell, charges := range aiUnit.SpellCharges {
            if charges > 0 {
                casted := false
                // try to cast this spell
                // FIXME: what to do if the unit is confused?
                model.InvokeSpell(spellSystem, army, nil, spell, func(success bool){
                    casted = true

                    aiUnit.SpellCharges[spell] -= 1

                    if success {
                        log.Debug("AI unit %v cast %v with strength %v", aiUnit.Unit.GetName(), spell.Name, spell.Cost(false))
                        spellSystem.PlaySound(spell)
                    }
                })

                if casted {
                    aiUnit.MovesLeft = fraction.FromInt(0)
                    aiActions.DoProjectiles()
                    return
                }
            }
        }

        // FIXME: cast a spell if the unit has mana (caster ability)
        // this can be a little tricky because typically the unit has a choice between a ranged magical attack
        // and casting a spell, but sometimes the spells might not be as good
    }

    // if the selected unit has ranged attacks, then try to use that
    // otherwise, if in melee range of some enemy then attack them
    // otherwise walk towards the enemy

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
        if model.withinMeleeRange(aiUnit, unit) && model.canMeleeAttack(aiUnit, unit, true) {
            aiActions.MeleeAttack(aiUnit, unit)
            return
        }
    }

    moved := false
    if aiUnit.CanTeleport() {
        moved = doAIMovementTeleport(model, aiActions, aiUnit, otherArmy)
    } else {
        moved = doAIMovementPathfinding(model, aiActions, aiUnit, otherArmy)
    }

    if !moved {
        // didn't make a choice, just exhaust moves left
        aiUnit.MovesLeft = fraction.FromInt(0)
    }
}

func doAIMovementTeleport(model *CombatModel, aiActions AIUnitActionsInterface, aiUnit *ArmyUnit, otherArmy *Army) bool {
    // get a list of units that we can melee attack
    var filterCanAttack []*ArmyUnit
    for _, unit := range otherArmy.units {
        // if this is a confused unit, don't self attack
        if unit == aiUnit {
            continue
        }

        if model.canMeleeAttack(aiUnit, unit, false) {
            filterCanAttack = append(filterCanAttack, unit)
        }
    }

    // nothing to attack, so no where to go
    if len(filterCanAttack) == 0 {
        return false
    }

    abs := func (x int) int {
        if x < 0 {
            return -x
        }
        return x
    }

    teleportDistance := 10

    hasMerge := aiUnit.HasAbility(data.AbilityMerging)

    validSquare := func (x int, y int) bool {
        if model.IsInsideMap(x, y) {
            distance := abs(x - aiUnit.X) + abs(y - aiUnit.Y)
            if distance > teleportDistance {
                return false
            }

            if model.ContainsWallTower(x, y) {
                return false
            }

            // don't teleport into or out of clouds
            if !aiUnit.IsFlying() && model.IsCloudTile(x, y) != model.IsCloudTile(aiUnit.X, aiUnit.Y) {
                return false
            }

            if model.GetUnit(x, y) != nil {
                return false
            }

            // log.Printf("considering teleporting to %d,%d (distance %d)", x, y, distance)
            return true
        }

        return false
    }

    for _, index := range rand.Perm(len(filterCanAttack)) {
        unit := filterCanAttack[index]

        // try all 8 squares around the unit
        for dx := -1; dx <= 1; dx++ {
            for dy := -1; dy <= 1; dy++ {
                if dx == 0 && dy == 0 {
                    continue
                }

                cx := unit.X + dx
                cy := unit.Y + dy

                if validSquare(cx, cy) {
                    aiActions.Teleport(aiUnit, cx, cy, hasMerge)
                    return true
                }
            }
        }
    }

    for _, unit := range filterCanAttack {
        shortestDistance := 100000
        var bestX, bestY int
        found := false

        for dx := -teleportDistance; dx <= teleportDistance; dx++ {
            for dy := -teleportDistance; dy <= teleportDistance; dy++ {
                if dx == 0 && dy == 0 {
                    continue
                }

                cx := aiUnit.X + dx
                cy := aiUnit.Y + dy

                distance := abs(cx - unit.X) + abs(cy - unit.Y)
                if distance < shortestDistance && validSquare(cx, cy) {
                    shortestDistance = distance
                    bestX = cx
                    bestY = cy
                    found = true
                }
            }
        }

        if found {
            aiActions.Teleport(aiUnit, bestX, bestY, hasMerge)
            return true
        }
    }

    // couldn't find a unit to teleport next to

    return false
}

func doAIMovementPathfinding(model *CombatModel, aiActions AIUnitActionsInterface, aiUnit *ArmyUnit, otherArmy *Army) bool {
    aiUnit.Paths = make(map[image.Point]pathfinding.Path)

    paths := make(map[*ArmyUnit]pathfinding.Path)

    getPath := func (unit *ArmyUnit) pathfinding.Path {
        path, found := paths[unit]
        if !found {
            // pretend that there is no unit at the tile. this is a sin of the highest order

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
            if !model.canMeleeAttack(aiUnit, unit, false) {
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
        path := getPath(closestEnemy)

        // a path of length 2 contains the position of the aiUnit and the position of the enemy, so they are right next to each other
        if len(path) == 2 && model.canMeleeAttack(aiUnit, closestEnemy, true) {
            aiActions.MeleeAttack(aiUnit, closestEnemy)
            return true
        } else if len(path) > 2 && aiUnit.CanSee(closestEnemy) {
            // ignore path[0], thats where we are now. also ignore the last element, since we can't move onto the enemy

            last := path[len(path)-1]
            if last.X == closestEnemy.X && last.Y == closestEnemy.Y {
                path = path[:len(path)-1]
            }

            lastIndex := 0
            for lastIndex < len(path) {
                lastIndex += 1
                if !aiUnit.CanFollowPath(path[0:lastIndex], false) {
                    lastIndex -= 1
                    break
                }
            }

            if lastIndex >= 1 && lastIndex <= len(path) {
                aiActions.MoveUnit(aiUnit, path[1:lastIndex])
                return true
            }
        }
    }

    // no enemy to move towards, then possibly move towards gate
    if aiUnit.Team == TeamDefender && model.InsideCityWall(aiUnit.X, aiUnit.Y) {
        // if inside a city wall, then move towards the gate
        gateX, gateY := model.GetCityGateCoordinates()
        if gateX != -1 && gateY != -1 {
            path, ok := model.computePath(aiUnit.X, aiUnit.Y, gateX, gateY, aiUnit.CanTraverseWall(), aiUnit.IsFlying())
            if ok && len(path) > 1 && aiUnit.CanFollowPath(path, false) {
                aiActions.MoveUnit(aiUnit, path[1:])
                return true
            }
        }
    }

    return false
}
