package ai

import (
    "log"
    "math/rand/v2"
    "image"

    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/lib/functional"
)

type RaiderAI struct {
    MonsterAccumulator int
    // MovedStacks map[*playerlib.UnitStack]bool
}

func MakeRaiderAI() *RaiderAI {
    return &RaiderAI{
        MonsterAccumulator: 0,
        // MovedStacks: make(map[*playerlib.UnitStack]bool),
    }
}

func (raider *RaiderAI) NewTurn(player *playerlib.Player) {
    // raider.MovedStacks = make(map[*playerlib.UnitStack]bool)
}

// return a random number between low and high inclusive
func randomRange(low int, high int) int {
    if low > high {
        return 0
    }

    if low == high {
        return low
    }

    return rand.N(high-low+1) + low
}

func (raider *RaiderAI) ConfirmEncounter(stack *playerlib.UnitStack, encounter *maplib.ExtraEncounter) bool {
    return false
}

func (raider *RaiderAI) InvalidMove(stack *playerlib.UnitStack) {
}

func (raider *RaiderAI) MovedStack(stack *playerlib.UnitStack, path pathfinding.Path) pathfinding.Path {
    return path
}

func (raider *RaiderAI) MoveStacks(player *playerlib.Player, enemies []*playerlib.Player, aiServices playerlib.AIServices) []playerlib.AIDecision {
    cityStackInfo := aiServices.ComputeCityStackInfo()

    // log.Printf("Raider stacks: %v", len(player.Stacks))

    var decisions []playerlib.AIDecision
    for _, stack := range player.Stacks {
        fog := player.GetFog(stack.Plane())
        map_ := aiServices.GetMap(stack.Plane())

        // if the stack is in a city and all units are in a patrol state, then choose some subset of the units and make them move around

        if !stack.OutOfMoves() {
            // FIXME: if the unit walked by a previously unknown city, they should stop their current path and possibly attack the city
            if stack.CurrentPath != nil {

                foundCity := false
                sightRange := stack.GetSightRange()
                check:
                for dx := -sightRange; dx <= sightRange; dx += sightRange {
                    for dy := -sightRange; dy <= sightRange; dy += sightRange {
                        city := cityStackInfo.FindCity(map_.WrapX(stack.X() + dx), stack.Y() + dy, stack.Plane())
                        if city != nil && city.GetBanner() != player.GetBanner() {
                            path, _ := aiServices.FindPath(stack.X(), stack.Y(), city.X, city.Y, player, stack, fog)
                            if len(path) > 0 {
                                foundCity = true
                                break check
                            }
                        }
                    }
                }

                if !foundCity {
                    decisions = append(decisions, &playerlib.AIMoveStackDecision{
                        Stack: stack,
                        Path: stack.CurrentPath,
                    })
                    // raider.MovedStacks[stack] = true
                    continue
                }
            }

            var currentPath pathfinding.Path

            for _, enemy := range enemies {
                for _, city := range enemy.Cities {
                    if city.Plane != stack.Plane() {
                        continue
                    }

                    // don't know about the city
                    // FIXME: how should this behave for different fog types?
                    if fog[city.X][city.Y] == data.FogTypeUnexplored {
                        continue
                    }

                    // log.Printf("ai stack %v found enemy city %v", stack, city)

                    path, _ := aiServices.FindPath(stack.X(), stack.Y(), city.X, city.Y, player, stack, fog)
                    if len(path) > 0 {
                        if len(currentPath) == 0 {
                            currentPath = path
                        } else {
                            if len(path) < len(currentPath) {
                                currentPath = path
                            }
                        }
                    }
                }
            }

            // choose a random unexplored tile on this continent
            if len(currentPath) == 0 {
                // allow flying/swimming units to walk randomly over the map
                if stack.AnyLandWalkers() {
                    continent := aiServices.GetMap(stack.Plane()).GetContinentTiles(stack.X(), stack.Y())
                    attempts := 6
                    for _, tileIndex := range rand.Perm(len(continent)) {
                        tile := &continent[tileIndex]
                        if fog[tile.X][tile.Y] == data.FogTypeUnexplored {
                            attempts -= 1
                            if attempts <= 0 {
                                break
                            }

                            currentPath, _ = aiServices.FindPath(stack.X(), stack.Y(), tile.X, tile.Y, player, stack, fog)
                            if len(currentPath) > 0 {
                                break
                            }
                        }
                    }
                }

                if len(currentPath) == 0 {
                    // just move randomly because all tiles have been explored
                    whereX := stack.X() + randomRange(-5, 5)
                    whereY := stack.Y() + randomRange(-5, 5)
                    currentPath, _ = aiServices.FindPath(stack.X(), stack.Y(), whereX, whereY, player, stack, fog)
                }
            }

            if len(currentPath) > 0 {
                decisions = append(decisions, &playerlib.AIMoveStackDecision{
                    Stack: stack,
                    Path: currentPath,
                })

            } else {
                stack.ExhaustMoves()
            }

            // raider.MovedStacks[stack] = true
        } else if player.FindCity(stack.X(), stack.Y(), stack.Plane()) != nil && rand.N(10) == 0 {
            busy := 0
            for _, unit := range stack.Units() {
                if unit.GetBusy() == units.BusyStatusPatrol {
                    busy += 1
                }
            }

            if busy == stack.Size() {
                var moveUnits []units.StackUnit
                maxUnits := rand.N(max(1, stack.Size() - 2))
                if maxUnits > 0 {
                    // log.Printf("Raiders moving %v units", maxUnits)
                    stackUnits := stack.Units()
                    for _, i := range rand.Perm(stack.Size()) {
                        if maxUnits == 0 {
                            break
                        }
                        moveUnits = append(moveUnits, stackUnits[i])
                        maxUnits -= 1
                    }

                    var paths []pathfinding.Path
                    for dx := -1; dx <= 1; dx += 1 {
                        for dy := -1; dy <= 1; dy += 1 {
                            if dx == 0 && dy == 0 {
                                continue
                            }
                            path, _ := aiServices.FindPath(stack.X(), stack.Y(), stack.X() + dx, stack.Y() + dy, player, stack, fog)
                            if len(path) > 0 {
                                paths = append(paths, path)
                            }
                        }
                    }

                    if len(paths) > 0 {
                        chosenPath := paths[rand.N(len(paths))]
                        /*
                        log.Printf("  chosen path %v", chosenPath)
                        for _, unit := range moveUnits {
                            log.Printf("  move unit %v", unit.GetName())
                        }
                        */

                        decisions = append(decisions, &playerlib.AIMoveStackDecision{
                            Stack: stack,
                            Units: moveUnits,
                            Path: chosenPath,
                        })
                    }
                }
            }
        }
    }
    return decisions
}

func (raider *RaiderAI) GetRampageRate(difficulty data.DifficultySetting) int {
    switch difficulty {
        case data.DifficultyIntro:
            return 1
        case data.DifficultyEasy:
            return rand.N(2) + 1
        case data.DifficultyAverage:
            return rand.N(3) + 1
        case data.DifficultyHard:
            return rand.N(4) + 1
        case data.DifficultyExtreme:
            return rand.N(5) + 1
        case data.DifficultyImpossible:
            return rand.N(6) + 1
    }

    return 0
}

func first[A any, B any](a A, b B) A {
    return a
}

func (raider *RaiderAI) CreateRampagingMonsters(player *playerlib.Player, aiServices playerlib.AIServices) []playerlib.AIDecision {
    isValidEncounterType := func (encounterType maplib.EncounterType) bool {
        switch encounterType {
            case maplib.EncounterTypeLair, maplib.EncounterTypeCave,
            maplib.EncounterTypeAncientTemple, maplib.EncounterTypeFallenTemple,
            maplib.EncounterTypeRuins, maplib.EncounterTypeAbandonedKeep, maplib.EncounterTypeDungeon:
            return true
        default:
            return false
        }
    }

    type FoundPoint struct {
        Point image.Point
        Plane data.Plane
        Found bool
    }

    type FindPoint func (plane data.Plane, continent terrain.Continent) FoundPoint

    findPoint := func (plane data.Plane, continent terrain.Continent) FoundPoint {
        // find a random encounter tile on the continent

        encounterTiles := set.MakeSet[image.Point]()
        map_ := aiServices.GetMap(plane)

        for _, point := range continent.Values() {
            encounter := map_.GetEncounter(point.X, point.Y)
            // encounter must be lair, dungeon, etc, and not have a stack on it already
            if encounter != nil && isValidEncounterType(encounter.Type) && first(aiServices.FindStack(point.X, point.Y, plane)) == nil {
                encounterTiles.Insert(point)
            }
        }

        choices := encounterTiles.Values()
        for _, i := range rand.Perm(len(choices)) {
            return FoundPoint{Point: choices[i], Plane: plane, Found: true}
        }

        return FoundPoint{}
    }

    type FindPointPlane func (terrain.Continent) FoundPoint

    findPointPlane := functional.Curry2(findPoint)


    var creators []func() FoundPoint
    allCities := aiServices.AllCities()
    for _, plane := range []data.Plane{data.PlaneArcanus, data.PlaneMyrror} {
        candidateContinents := set.MakeSet[terrain.Continent]()

        // get all continents with cities on them
        continents := aiServices.GetMap(plane).GetContinents()
        for _, city := range allCities {
            if city.Plane == plane && city.GetBanner() != player.GetBanner() {
                for _, continent := range continents {
                    if continent.Contains(image.Pt(city.X, city.Y)) {
                        candidateContinents.Insert(continent)
                    }
                }
            }
        }

        for _, continent := range candidateContinents.Values() {
            f := functional.Curry1(findPointPlane(plane))(continent)
            creators = append(creators, f)
        }
    }

    for _, i := range rand.Perm(len(creators)) {
        f := creators[i]
        found := f()
        if found.Found {
            map_ := aiServices.GetMap(found.Plane)
            encounter := map_.GetEncounter(found.Point.X, found.Point.Y)
            // selected a point on the map, now create monsters
            return createMonsters(found.Point, found.Plane, encounter.Type, aiServices.GetDifficulty(), aiServices.GetTurnNumber())
        }
    }

    return nil
}

// create some fantastic creatures at the given point based on the realm of the encounter type
// FIXME: this will create monsters on top of the encounter zone, but if a battle occurs then the wizard will fight the overland
// monsters rather than the monsters guarding the encounter zone. after defeating the overland monsters the wizard must
// then re-enter the encounter zone to fight the monsters guarding it.
func createMonsters(point image.Point, plane data.Plane, encounterType maplib.EncounterType, difficulty data.DifficultySetting, turn uint64) []playerlib.AIDecision {
    budget := 20
    switch difficulty {
        case data.DifficultyIntro: budget = 20
        case data.DifficultyEasy: budget = 30
        case data.DifficultyAverage: budget = 40
        case data.DifficultyHard: budget = 50
        case data.DifficultyExtreme: budget = 60
        case data.DifficultyImpossible: budget = 70
    }

    // by turn 200, the budget is increased by 3x
    budget = int(float64(budget) * (1.0 + float64(min(150, turn - 50)) / 50.0))

    var decisions []playerlib.AIDecision

    // each encounter type can create monsters of a specific realm
    var realms []data.MagicType
    switch encounterType {
        case maplib.EncounterTypeLair: realms = []data.MagicType{data.ChaosMagic, data.DeathMagic, data.NatureMagic}
        case maplib.EncounterTypeCave: realms = []data.MagicType{data.ChaosMagic, data.DeathMagic, data.NatureMagic}
        case maplib.EncounterTypeAncientTemple: realms = []data.MagicType{data.DeathMagic, data.LifeMagic}
        case maplib.EncounterTypeFallenTemple: realms = []data.MagicType{data.DeathMagic, data.LifeMagic}
        case maplib.EncounterTypeRuins: realms = []data.MagicType{data.DeathMagic, data.LifeMagic}
        case maplib.EncounterTypeAbandonedKeep: realms = []data.MagicType{data.ChaosMagic, data.DeathMagic, data.NatureMagic}
        case maplib.EncounterTypeDungeon: realms = []data.MagicType{data.ChaosMagic, data.DeathMagic, data.NatureMagic}
    }

    if len(realms) == 0 {
        return nil
    }

    // choose one of the available realm options
    choose := realms[rand.N(len(realms))]
    allUnits := units.UnitsByRealm(choose)

    for budget > 0 {
        // log.Printf("monster budget: %v", budget)
        added := false
        // find a random unit that can be created within the budget
        for _, i := range rand.Perm(len(allUnits)) {
            if allUnits[i].CastingCost > 0 && allUnits[i].CastingCost <= budget {
                budget -= allUnits[i].CastingCost
                decision := &playerlib.AICreateUnitDecision{
                    Unit: allUnits[i],
                    X: point.X,
                    Y: point.Y,
                    Plane: plane,
                }

                decisions = append(decisions, decision)
                added = true
                break
            }
        }

        // couldnt add any more units, so stop
        if !added {
            budget = 0
        }
    }

    return decisions
}

func (raider *RaiderAI) CreateUnits(player *playerlib.Player, aiServices playerlib.AIServices) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision

    // don't create too many stacks
    if len(player.Stacks) > 20 {
        return decisions
    }

    /*
    getContinent := functional.Memoize3(func(x int, y int, plane data.Plane) []maplib.FullTile {
        return aiServices.GetMap(plane).GetContinentTiles(x, y)
    })

    findEnemyCity := func (city *citylib.City) bool {
        tiles := getContinent(city.X, city.Y, city.Plane)
        for _, tile := range tiles {
            otherCity, otherPlayer := aiServices.FindCity(tile.X, tile.Y, city.Plane)
            if otherCity != nil && otherPlayer != player {
                return true
            }
        }

        return false
    }
    */

    // after turn 50, every turn there is a N% chance (based on difficulty) to create a pack of monsters
    // in an encounter zone (but not a tower or node)
    // the monsters should try to walk towards the nearest city, but will roam around randomly if no city is found
    raider.MonsterAccumulator += raider.GetRampageRate(aiServices.GetDifficulty())
    if raider.MonsterAccumulator >= 50 && aiServices.GetTurnNumber() >= 50 {
        raider.MonsterAccumulator = 0
        log.Printf("Create rampaging monsters")
        decisions = append(decisions, raider.CreateRampagingMonsters(player, aiServices)...)
    }

    for _, city := range player.Cities {
        stack := player.FindStack(city.X, city.Y, city.Plane)

        makeUnit := false
        // always make a unit if there is no stack in the city
        if stack == nil || stack.IsEmpty() {
            makeUnit = true
        } else if rand.N(15) == 0 && (stack != nil && stack.Size() < 6) {
            makeUnit = true
        }

        if makeUnit {
            decisions = append(decisions, &playerlib.AICreateUnitDecision{
                // FIXME: use some sort of budget for the unit so that mostly low level units are created early in the game
                Unit: units.ChooseRandomUnit(city.Race),
                X: city.X,
                Y: city.Y,
                Plane: city.Plane,
                Patrol: true,
            })
        }
    }

    return decisions
}

// always force all raider cities to have the maximum number of farmers
func (raider *RaiderAI) UpdateCities(self *playerlib.Player) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision

    for _, city := range self.Cities {
        // request the number of farmers to be the maximum possible
        decisions = append(decisions, &playerlib.AIUpdateCityDecision{
            City: city,
            Farmers: city.Citizens(),
            Workers: 0,
        })
    }

    return decisions
}

func (raider *RaiderAI) Update(player *playerlib.Player, aiServices playerlib.AIServices) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision

    enemies := aiServices.GetEnemies(player)

    decisions = append(decisions, raider.MoveStacks(player, enemies, aiServices)...)
    decisions = append(decisions, raider.CreateUnits(player, aiServices)...)
    decisions = append(decisions, raider.UpdateCities(player)...)

    return decisions
}

func (raider *RaiderAI) PostUpdate(self *playerlib.Player, aiServices playerlib.AIServices) {
}

func (raider *RaiderAI) ProducedUnit(city *citylib.City, player *playerlib.Player) {
}

func (raider *RaiderAI) ConfirmRazeTown(city *citylib.City) bool {
    return true
}

func (raider *RaiderAI) HandleMerchantItem(self *playerlib.Player, item *artifact.Artifact, cost int) bool {
    return false
}
