package maplib

import (
    "image"
    // "log"
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/lib/set"
)

/* choose X points surrounding the node. 0,0 is the node itself. for arcanus, choose 5-10 points from a 4x4 square.
 * for myrror choose 10-20 points from a 5x5 square.
 */
func makeZone(plane data.Plane) []image.Point {
    // choose X points
    // maxSize := 4
    numPoints := 0
    if plane == data.PlaneArcanus {
        // maxSize = 4
        numPoints = 5 + rand.IntN(5)
    } else if plane == data.PlaneMyrror {
        // maxSize = 5
        numPoints = 10 + rand.IntN(10)
    }

    // chosen := make(map[image.Point]bool)
    out := make([]image.Point, 0, numPoints)

    // always choose the center, which is where the node itself is
    // chosen[image.Pt(0, 0)] = true
    out = append(out, image.Pt(0, 0))

    numPoints -= 1

    abs := func (x int) int {
        if x < 0 {
            return -x
        }
        return x
    }

    // return the points that are distance away from the center, including the corners
    getPoints := func (distance int) []image.Point {
        var possible []image.Point
        for x := -distance; x <= distance; x++ {
            for y := -distance; y <= distance; y++ {
                if abs(x) == distance || abs(y) == distance {
                    possible = append(possible, image.Pt(x, y))
                }
            }
        }
        return possible
    }

    // take the first N points
    distance1 := getPoints(1)
    for _, choice := range rand.Perm(len(distance1)) {
        if numPoints == 0 {
            break
        }
        out = append(out, distance1[choice])
        numPoints -= 1
    }

    distance2 := getPoints(2)
    for _, choice := range rand.Perm(len(distance2)) {
        if numPoints == 0 {
            break
        }
        out = append(out, distance2[choice])
        numPoints -= 1
    }

    return out
}

/* budget for making encounter monsters is zone size + bonus
 */
func computeEncounterBudget(magicSetting data.MagicSetting, difficultySetting data.DifficultySetting, zoneSize int) int {
    budget := 0

    // these formulas come from the master of magic wiki
    switch magicSetting {
        case data.MagicSettingWeak:
            budget = (rand.IntN(11) + 4) * (zoneSize * zoneSize) / 2
        case data.MagicSettingNormal:
            budget = (rand.IntN(11) + 4) * (zoneSize * zoneSize)
        case data.MagicSettingPowerful:
            budget = (rand.IntN(11) + 4) * (zoneSize * zoneSize) * 3 / 2
    }

    bonus := float64(0)
    switch difficultySetting {
        case data.DifficultyIntro: bonus = -0.75
        case data.DifficultyEasy: bonus = -0.5
        case data.DifficultyAverage: bonus = -0.25
        case data.DifficultyHard: bonus = 0
        case data.DifficultyExtreme: bonus = 0.25
        case data.DifficultyImpossible: bonus = 0.50
    }

    return budget + int(float64(budget) * bonus)
}

/* divide budget by some divisor in range 1 to numChoices. find the enemy with the largest cost
 * that fits in the divided result.
 */
func chooseEnemy[E comparable](enemyCosts map[E]int, budget int, numChoices int) E {
    if numChoices < 0 {
        numChoices = 0
    }

    choices := rand.Perm(numChoices)
    var zero E

    for _, choice := range choices {
        divisor := choice + 1

        var enemyChoice E
        maxCost := 0

        for unit, cost := range enemyCosts {
            if cost > maxCost && cost <= budget / divisor {
                enemyChoice = unit
                maxCost = cost
            }
        }

        if enemyChoice != zero {
            return enemyChoice
        }
    }

    return zero
}

func chooseGuardianAndSecondary[E comparable](enemyCosts map[E]int, makeUnit func(E) units.Unit, budget int) ([]units.Unit, []units.Unit) {
    var guardians []units.Unit
    var secondary []units.Unit

    enemyChoice := chooseEnemy(enemyCosts, budget, 4)

    var zero E

    // chose no enemies!
    if enemyChoice == zero {
        return nil, nil
    }

    numGuardians := budget / enemyCosts[enemyChoice]
    if numGuardians > 9 {
        numGuardians = 9
    }

    for i := 0; i < numGuardians; i++ {
        guardians = append(guardians, makeUnit(enemyChoice))
    }

    remainingBudget := budget - numGuardians * enemyCosts[enemyChoice]

    enemyChoice = chooseEnemy(enemyCosts, remainingBudget, 10 - numGuardians)

    if enemyChoice != zero {
        secondaryCount := remainingBudget / enemyCosts[enemyChoice]

        if secondaryCount > 9 - numGuardians {
            secondaryCount = 9 - numGuardians
        }

        for i := 0; i < secondaryCount; i++ {
            secondary = append(secondary, makeUnit(enemyChoice))
        }
    }

    return guardians, secondary
}

/* returns guardian units and secondary units
 */
func computeNatureNodeEnemies(budget int) ([]units.Unit, []units.Unit) {
    type Enemy int
    const (
        None Enemy = iota
        WarBear
        Sprites
        EarthElemental
        GiantSpiders
        Cockatrices
        Basilisk
        StoneGiant
        Gorgons
        Behemoth
        Colossus
        GreatWyrm
    )

    makeUnit := func(enemy Enemy) units.Unit {
        switch enemy {
            case WarBear: return units.WarBear
            case Sprites: return units.Sprites
            case EarthElemental: return units.EarthElemental
            case GiantSpiders: return units.GiantSpiders
            case Cockatrices: return units.Cockatrices
            case Basilisk: return units.Basilisk
            case StoneGiant: return units.StoneGiant
            case Gorgons: return units.Gorgon
            case Behemoth: return units.Behemoth
            case Colossus: return units.Colossus
            case GreatWyrm: return units.GreatWyrm
        }

        return units.UnitNone
    }

    enemyCosts := map[Enemy]int{
        None: 0,
        WarBear: units.WarBear.CastingCost,
        Sprites: units.Sprites.CastingCost,
        EarthElemental: units.EarthElemental.CastingCost,
        GiantSpiders: units.GiantSpiders.CastingCost,
        Cockatrices: units.Cockatrices.CastingCost,
        Basilisk: units.Basilisk.CastingCost,
        StoneGiant: units.StoneGiant.CastingCost,
        Gorgons: units.Gorgon.CastingCost,
        Behemoth: units.Behemoth.CastingCost,
        Colossus: units.Colossus.CastingCost,
        GreatWyrm: units.GreatWyrm.CastingCost,
    }

    return chooseGuardianAndSecondary(enemyCosts, makeUnit, budget)
}

func computeSorceryNodeEnemies(budget int) ([]units.Unit, []units.Unit) {
    type Enemy int
    const (
        None Enemy = iota
        PhantomWarriors
        Naga
        AirElemental
        PhantomBeast
        StormGiant
        Djinn
        SkyDrake
    )

    makeUnit := func(enemy Enemy) units.Unit {
        switch enemy {
            case PhantomWarriors: return units.PhantomWarrior
            case Naga: return units.Nagas
            case AirElemental: return units.AirElemental
            case PhantomBeast: return units.PhantomBeast
            case StormGiant: return units.StormGiant
            case Djinn: return units.Djinn
            case SkyDrake: return units.SkyDrake
        }

        return units.UnitNone
    }

    enemyCosts := map[Enemy]int{
        None: 0,
        PhantomWarriors: units.PhantomWarrior.CastingCost,
        Naga: units.Nagas.CastingCost,
        AirElemental: units.AirElemental.CastingCost,
        PhantomBeast: units.PhantomBeast.CastingCost,
        StormGiant: units.StormGiant.CastingCost,
        Djinn: units.Djinn.CastingCost,
        SkyDrake: units.SkyDrake.CastingCost,
    }

    return chooseGuardianAndSecondary(enemyCosts, makeUnit, budget)
}

func computeDeathNodeEnemies(budget int) ([]units.Unit, []units.Unit) {
    type Enemy int
    const (
        None Enemy = iota
        Skeletons
        Zombies
        Ghouls
        Demons
        NightStalker
        Werewolves
        ShadowDemons
        Wraiths
        DeathKnights
        DemonLord
    )

    makeUnit := func(enemy Enemy) units.Unit {
        switch enemy {
            case Skeletons: return units.Skeleton
            case Zombies: return units.Zombie
            case Ghouls: return units.Ghoul
            case Demons: return units.Demon
            case NightStalker: return units.NightStalker
            case Werewolves: return units.WereWolf
            case ShadowDemons: return units.ShadowDemons
            case Wraiths: return units.Wraith
            case DeathKnights: return units.DeathKnights
            case DemonLord: return units.DemonLord
        }

        return units.UnitNone
    }

    enemyCosts := map[Enemy]int{
        None: 0,
        Skeletons: units.Skeleton.CastingCost,
        Zombies: units.Zombie.CastingCost,
        Ghouls: units.Ghoul.CastingCost,
        Demons: units.Demon.CastingCost,
        NightStalker: units.NightStalker.CastingCost,
        Werewolves: units.WereWolf.CastingCost,
        ShadowDemons: units.ShadowDemons.CastingCost,
        Wraiths: units.Wraith.CastingCost,
        DeathKnights: units.DeathKnights.CastingCost,
        DemonLord: units.DemonLord.CastingCost,
    }

    return chooseGuardianAndSecondary(enemyCosts, makeUnit, budget)
}

func computeLifeNodeEnemies(budget int) ([]units.Unit, []units.Unit) {
    type Enemy int
    const (
        None Enemy = iota
        GuardianSpirit
        Unicorns
        Angel
        ArchAngel
    )

    makeUnit := func(enemy Enemy) units.Unit {
        switch enemy {
            case GuardianSpirit: return units.GuardianSpirit
            case Unicorns: return units.Unicorn
            case Angel: return units.Angel
            case ArchAngel: return units.ArchAngel
        }

        return units.UnitNone
    }

    enemyCosts := map[Enemy]int{
        None: 0,
        GuardianSpirit: 75,
        Unicorns: 250,
        Angel: 550,
        ArchAngel: 950,
    }

    return chooseGuardianAndSecondary(enemyCosts, makeUnit, budget)
}

func computeChaosNodeEnemies(budget int) ([]units.Unit, []units.Unit) {
    type Enemy int
    const (
        None Enemy = iota
        HellHounds
        FireElemental
        FireGiant
        Gargoyles
        DoomBat
        Chimera
        ChaosSpawn
        Efreet
        Hydra
        GreatDrake
    )

    makeUnit := func(enemy Enemy) units.Unit {
        switch enemy {
            case HellHounds: return units.HellHounds
            case FireElemental: return units.FireElemental
            case FireGiant: return units.FireGiant
            case Gargoyles: return units.Gargoyle
            case DoomBat: return units.DoomBat
            case Chimera: return units.Chimeras
            case ChaosSpawn: return units.ChaosSpawn
            case Efreet: return units.Efreet
            case Hydra: return units.Hydra
            case GreatDrake: return units.GreatDrake
        }

        return units.UnitNone
    }

    enemyCosts := map[Enemy]int{
        None: 0,
        HellHounds: 40,
        FireElemental: 100,
        FireGiant: 150,
        Gargoyles: 200,
        DoomBat: 300,
        Chimera: 350,
        ChaosSpawn: 400,
        Efreet: 550,
        Hydra: 650,
        GreatDrake: 900,
    }

    return chooseGuardianAndSecondary(enemyCosts, makeUnit, budget)
}

func MakeMagicNode(kind MagicNode, magicSetting data.MagicSetting, difficulty data.DifficultySetting, plane data.Plane) (*ExtraMagicNode, *ExtraEncounter) {
    zone := makeZone(plane)

    rand.Shuffle(len(zone), func(i, j int) {
        zone[i], zone[j] = zone[j], zone[i]
    })

    var guardians []units.Unit
    var secondary []units.Unit

    budget := computeEncounterBudget(magicSetting, difficulty, len(zone))

    var encouterType EncounterType

    for len(guardians) == 0 {
        switch kind {
            case MagicNodeNature:
                guardians, secondary = computeNatureNodeEnemies(budget)
                encouterType = EncounterTypeNatureNode
                // log.Printf("Created nature node guardians: %v secondary: %v", guardians, secondary)
            case MagicNodeSorcery:
                guardians, secondary = computeSorceryNodeEnemies(budget)
                encouterType = EncounterTypeSorceryNode
                // log.Printf("Created sorcery node guardians: %v secondary: %v", guardians, secondary)
            case MagicNodeChaos:
                guardians, secondary = computeChaosNodeEnemies(budget)
                encouterType = EncounterTypeChaosNode
                // log.Printf("Created chaos node guardians: %v secondary: %v", guardians, secondary)
        }

        // failed, so increase the budget
        if len(guardians) == 0 {
            budget += 50
        }
    }

    magicNode := ExtraMagicNode{
        Kind: kind,
        Zone: zone,
    }

    encounter := ExtraEncounter{
        Type: encouterType,
        Units: append(guardians, secondary...),
        Budget: budget,
        ExploredBy: set.MakeSet[Wizard](),
    }

    return &magicNode, &encounter
}
