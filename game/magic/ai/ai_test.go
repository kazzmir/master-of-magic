package ai

import (
    "image"
    "testing"

    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    "github.com/kazzmir/master-of-magic/lib/set"
)

func TestUnitAttackPowerUsesToHit(test *testing.T) {
    swordsmen := units.MakeOverworldUnit(units.HighMenSwordsmen, 0, 0, data.PlaneArcanus)
    if unitAttackPower(swordsmen) <= 0 {
        test.Errorf("swordsmen should have positive attack power after to-hit scaling")
    }

    settler := units.MakeOverworldUnit(units.HighMenSettlers, 0, 0, data.PlaneArcanus)
    if unitAttackPower(settler) != 0 {
        test.Errorf("settlers should still count as 0 attack power, got %v", unitAttackPower(settler))
    }
}

func TestBuyItem(test *testing.T) {
    enemy := MakeEnemyAI()

    self := playerlib.MakePlayer(setup.WizardCustom{}, false, 2, 2, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{})
    self.Gold = 1000

    weapon := artifact.Artifact{
    }

    if !enemy.HandleMerchantItem(self, &weapon, 300) {
        test.Errorf("expected to buy item, but did not")
    }

    artifacts := 0
    for _, item := range self.VaultEquipment {
        if item != nil {
            artifacts += 1
        }
    }

    if artifacts != 1 {
        test.Errorf("expected 1 artifact in vault, got %d", artifacts)
    }

    if self.Gold != 700 {
        test.Errorf("expected 700 gold after purchase, got %d", self.Gold)
    }
}

// stubAIServices is a no-LBX stand-in for GameModel so Enemy2 gates can be tested.
type stubAIServices struct {
    difficulty data.DifficultySetting
    aggressive bool
    enemies []*playerlib.Player
}

func (s *stubAIServices) GetCityEnchantmentsByBanner(banner data.BannerType) []playerlib.CityEnchantment {
    return nil
}

func (s *stubAIServices) FindPath(oldX int, oldY int, newX int, newY int, player *playerlib.Player, stack playerlib.PathStack, fog data.FogMap) (pathfinding.Path, bool) {
    return pathfinding.Path{image.Pt(oldX, oldY), image.Pt(newX, newY)}, true
}

func (s *stubAIServices) FindSettlableLocations(x int, y int, plane data.Plane, fog data.FogMap) []image.Point {
    return []image.Point{image.Pt(x + 3, y)}
}

func (s *stubAIServices) IsSettlableLocation(x int, y int, plane data.Plane) bool {
    return true
}

func (s *stubAIServices) GetDifficulty() data.DifficultySetting {
    return s.difficulty
}

func (s *stubAIServices) GetAggressiveAI() bool {
    return s.aggressive
}

func (s *stubAIServices) GetMap(data.Plane) *maplib.Map {
    return nil
}

func (s *stubAIServices) FindCitiesOnContinent(x int, y int, plane data.Plane, player *playerlib.Player) []*citylib.City {
    return nil
}

func (s *stubAIServices) FindStacksOnContinent(x int, y int, plane data.Plane, player *playerlib.Player) []*playerlib.UnitStack {
    return nil
}

func (s *stubAIServices) GetTurnNumber() uint64 {
    return 20
}

func (s *stubAIServices) ComputeMaximumPopulation(int, int, data.Plane) int {
    return 10
}

func (s *stubAIServices) ComputePower(player *playerlib.Player) int {
    return 0
}

func (s *stubAIServices) AllCities() []*citylib.City {
    return nil
}

func (s *stubAIServices) FindStack(x int, y int, plane data.Plane) (*playerlib.UnitStack, *playerlib.Player) {
    return nil, nil
}

func (s *stubAIServices) FindCity(x int, y int, plane data.Plane) (*citylib.City, *playerlib.Player) {
    return nil, nil
}

func (s *stubAIServices) ComputeCityStackInfo() playerlib.CityStackInfo {
    return playerlib.CityStackInfo{}
}

func (s *stubAIServices) GetEnemies(player *playerlib.Player) []*playerlib.Player {
    return s.enemies
}

func (s *stubAIServices) GetBuildingInfos() buildinglib.BuildingInfos {
    return nil
}

func testPlayer() *playerlib.Player {
    return playerlib.MakePlayer(setup.WizardCustom{}, false, 20, 20, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{})
}

func testCity(name string, x int, y int) *citylib.City {
    city := citylib.MakeCity(name, x, y, data.RaceHighMen, nil, nil, nil, nil)
    city.Plane = data.PlaneArcanus
    return city
}

func TestEnemy2KnobsClassicVsAggressive(test *testing.T) {
    classic := enemy2Knobs{Aggressive: false, Difficulty: data.DifficultyAverage}
    if classic.connectCitiesMin() != 2 {
        test.Errorf("classic Average should start roads at 2 cities, got %d", classic.connectCitiesMin())
    }
    if classic.canQueueSettler(0) {
        test.Errorf("classic Average should not queue a settler at 0 food")
    }
    if !classic.canQueueSettler(1) {
        test.Errorf("classic Average should queue a settler when food is surplus")
    }
    if classic.settlerChance() != 60 {
        test.Errorf("classic settler chance should stay 60, got %d", classic.settlerChance())
    }

    aggressive := enemy2Knobs{Aggressive: true, Difficulty: data.DifficultyAverage}
    if aggressive.connectCitiesMin() != 2 {
        test.Errorf("Aggressive AI should start roads at 2 cities, got %d", aggressive.connectCitiesMin())
    }
    if !aggressive.canQueueSettler(0) {
        test.Errorf("Aggressive AI should queue a settler at 0 food")
    }
    if aggressive.canQueueSettler(-1) {
        test.Errorf("Aggressive AI should still not queue a settler while starving")
    }

    hard := enemy2Knobs{Aggressive: false, Difficulty: data.DifficultyHard}
    if hard.connectCitiesMin() != 2 {
        test.Errorf("Hard should start roads at 2 cities, got %d", hard.connectCitiesMin())
    }
    if !hard.canQueueSettler(0) {
        test.Errorf("Hard should queue a settler at 0 food")
    }

    intro := enemy2Knobs{Aggressive: false, Difficulty: data.DifficultyIntro}
    if intro.connectCitiesMin() != 3 {
        test.Errorf("Intro should keep the 3-city road gate, got %d", intro.connectCitiesMin())
    }
    if intro.canQueueSettler(0) {
        test.Errorf("Intro should not queue a settler at 0 food")
    }

    easy := enemy2Knobs{Aggressive: false, Difficulty: data.DifficultyEasy}
    if easy.connectCitiesMin() != 3 {
        test.Errorf("Easy Classic should keep the 3-city road gate, got %d", easy.connectCitiesMin())
    }
}

func TestMakeEnemy2KnobsUsesDifficultyAndAggressive(test *testing.T) {
    services := &stubAIServices{difficulty: data.DifficultyAverage, aggressive: false}
    knobs := makeEnemy2Knobs(services)
    if knobs.Difficulty != data.DifficultyAverage || knobs.Aggressive {
        test.Errorf("expected Average/classic knobs, got difficulty=%v aggressive=%v", knobs.Difficulty, knobs.Aggressive)
    }

    services.difficulty = data.DifficultyImpossible
    services.aggressive = true
    knobs = makeEnemy2Knobs(services)
    if knobs.Difficulty != data.DifficultyImpossible || !knobs.Aggressive {
        test.Errorf("GetDifficulty/GetAggressiveAI were ignored: difficulty=%v aggressive=%v", knobs.Difficulty, knobs.Aggressive)
    }
}

func hasGoalType(goals []EnemyGoal, goalType GoalType) bool {
    for _, goal := range goals {
        if goal.Goal == goalType {
            return true
        }
    }
    return false
}

func TestComputeGoalsConnectCitiesUsesDifficulty(test *testing.T) {
    self := testPlayer()
    self.AddCity(testCity("Alpha", 1, 1))
    self.AddCity(testCity("Beta", 5, 5))

    ai := MakeEnemy2AI()
    intro := ai.ComputeGoals(self, &stubAIServices{difficulty: data.DifficultyIntro})
    if hasGoalType(intro, GoalConnectCities) {
        test.Errorf("Intro with 2 cities should not enable GoalConnectCities")
    }

    classic := ai.ComputeGoals(self, &stubAIServices{difficulty: data.DifficultyAverage})
    if !hasGoalType(classic, GoalConnectCities) {
        test.Errorf("classic Average with 2 cities should enable GoalConnectCities")
    }

    hard := ai.ComputeGoals(self, &stubAIServices{difficulty: data.DifficultyHard})
    if !hasGoalType(hard, GoalConnectCities) {
        test.Errorf("Hard with 2 cities should enable GoalConnectCities")
    }

    aggressive := ai.ComputeGoals(self, &stubAIServices{difficulty: data.DifficultyAverage, aggressive: true})
    if !hasGoalType(aggressive, GoalConnectCities) {
        test.Errorf("Aggressive AI with 2 cities should enable GoalConnectCities")
    }
}

func countMoveDecisions(decisions []playerlib.AIDecision) int {
    count := 0
    for _, decision := range decisions {
        if _, ok := decision.(*playerlib.AIMoveStackDecision); ok {
            count++
        }
    }
    return count
}

func TestGoalDefeatEnemiesAttacksVisibleCityWithoutEnemyStack(test *testing.T) {
    self := testPlayer()
    for range 4 {
        self.AddUnit(units.MakeOverworldUnit(units.HighMenSwordsmen, 1, 1, data.PlaneArcanus))
    }
    // a settler must not be sent on the attack
    self.AddUnit(units.MakeOverworldUnit(units.HighMenSettlers, 2, 2, data.PlaneArcanus))

    enemy := testPlayer()
    enemy.AddCity(testCity("Rival", 8, 8))

    self.LiftFogAll(data.PlaneArcanus)

    services := &stubAIServices{
        difficulty: data.DifficultyAverage,
        enemies: []*playerlib.Player{enemy},
    }

    ai := MakeEnemy2AI()
    seenGoals := set.MakeSet[GoalType]()
    decisions, _ := ai.GoalDecisions(self, services, EnemyGoal{Goal: GoalDefeatEnemies, Weight: 0.9}, seenGoals, &AIData{
        FoodPerTurn: func() int { return 1 },
        GoldPerTurn: func() int { return 1 },
    })

    if countMoveDecisions(decisions) < 1 {
        test.Errorf("expected a city march with no visible enemy stack, got %#v", decisions)
    }

    for _, decision := range decisions {
        move, ok := decision.(*playerlib.AIMoveStackDecision)
        if !ok {
            continue
        }
        if stackHasSettler(move.Stack) {
            test.Errorf("settler stack was sent to attack a city")
        }
    }
}

func TestGoalDefeatEnemiesAttacksRememberedCityWithoutVisibleStack(test *testing.T) {
    self := testPlayer()
    for range 4 {
        self.AddUnit(units.MakeOverworldUnit(units.HighMenSwordsmen, 1, 1, data.PlaneArcanus))
    }

    enemy := testPlayer()
    enemy.AddCity(testCity("Fogged", 8, 8))

    services := &stubAIServices{
        difficulty: data.DifficultyAverage,
        enemies: []*playerlib.Player{enemy},
    }

    ai := MakeEnemy2AI()
    ai.KnownEnemyCities[CityKey{X: 8, Y: 8, Plane: data.PlaneArcanus}] = &EnemyCityInfo{X: 8, Y: 8, Plane: data.PlaneArcanus}

    seenGoals := set.MakeSet[GoalType]()
    decisions, _ := ai.GoalDecisions(self, services, EnemyGoal{Goal: GoalDefeatEnemies, Weight: 0.9}, seenGoals, &AIData{
        FoodPerTurn: func() int { return 1 },
        GoldPerTurn: func() int { return 1 },
    })

    if countMoveDecisions(decisions) < 1 {
        test.Errorf("expected a march on a remembered city with no visible stacks, got %#v", decisions)
    }
}

func TestGoalDefeatEnemiesAggressiveMarchesSmallStackOnVisibleCity(test *testing.T) {
    self := testPlayer()
    self.AddUnit(units.MakeOverworldUnit(units.HighMenSwordsmen, 1, 1, data.PlaneArcanus))

    enemy := testPlayer()
    enemy.AddCity(testCity("Hamlet", 8, 8))
    self.LiftFogAll(data.PlaneArcanus)

    services := &stubAIServices{
        difficulty: data.DifficultyAverage,
        aggressive: true,
        enemies: []*playerlib.Player{enemy},
    }

    ai := MakeEnemy2AI()
    seenGoals := set.MakeSet[GoalType]()
    decisions, _ := ai.GoalDecisions(self, services, EnemyGoal{Goal: GoalDefeatEnemies, Weight: 0.9}, seenGoals, &AIData{
        FoodPerTurn: func() int { return 1 },
        GoldPerTurn: func() int { return 1 },
    })

    if countMoveDecisions(decisions) < 1 {
        test.Errorf("Aggressive AI should march a single swordsmen stack on a visible city, got %#v", decisions)
    }
}
