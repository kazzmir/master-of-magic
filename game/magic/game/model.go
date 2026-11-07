package game

import (
    "image"
    "slices"
    "sync"
    "math"
    "math/rand/v2"
    "errors"
    "strings"
    "fmt"
    "log"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/ai"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/combat"
    settingslib "github.com/kazzmir/master-of-magic/game/magic/settings"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    "github.com/kazzmir/master-of-magic/lib/functional"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
)

type GameModel struct {
    ArcanusMap *maplib.Map
    MyrrorMap *maplib.Map
    Players []*playerlib.Player
    Plane data.Plane

    ArtifactPool *artifact.Catalog

    Settings setup.NewGameSettings
    GameSettings *settingslib.Settings

    // live read of the session-only Aggressive AI checkbox (not serialized).
    // Set by MakeGameWithModel from game Settings. Nil means Classic.
    IsAggressiveAI func() bool

    AIMode ai.Mode

    heroNames map[int]map[herolib.HeroType]string
    allSpells spellbook.Spells

    // water bodies holds all the points that are part of the same body of water.
    // a plane can have mulitple bodies of water if some bodies are landlocked (enclosed entirely by a continent)
    // this is lazily initialized on first use
    WaterBodies map[data.Plane][]*set.Set[image.Point]

    CurrentPlayer int

    TurnNumber uint64

    // for communication with the UI
    Events chan GameEvent

    BuildingInfo buildinglib.BuildingInfos
    LbxCache *lbx.LbxCache

    // the scroll events that occurred this turn
    ScrollEvents []*GameEventScroll

    // https://masterofmagic.fandom.com/wiki/Event
    RandomEvents []*RandomEvent
    LastEventTurn uint64
}

func MakeGameModel(terrainData *terrain.TerrainData, settings setup.NewGameSettings,
                   gameSettings *settingslib.Settings,
                   startingPlane data.Plane, events chan GameEvent,
                   heroNames map[int]map[herolib.HeroType]string, allSpells spellbook.Spells,
                   artifactPool *artifact.Catalog,
                   buildingInfo buildinglib.BuildingInfos,
                   lbxCache *lbx.LbxCache,
               ) *GameModel {

    planeTowers := maplib.GeneratePlaneTowerPositions(settings.LandSize, 6)

    model := GameModel{
        ArtifactPool: artifactPool,
        Settings: settings,
        GameSettings: gameSettings,
        heroNames: heroNames,
        allSpells: allSpells,
        Plane: startingPlane,
        CurrentPlayer: -1,
        Events: events,
        BuildingInfo: buildingInfo,
        LbxCache: lbxCache,
    }

    model.ArcanusMap = maplib.MakeMap(terrainData, settings.LandSize, settings.Magic, settings.Difficulty, data.PlaneArcanus, &model, planeTowers)
    model.MyrrorMap = maplib.MakeMap(terrainData, settings.LandSize, settings.Magic, settings.Difficulty, data.PlaneMyrror, &model, planeTowers)
    return &model
}

func (model *GameModel) CurrentMap() *maplib.Map {
    if model.Plane == data.PlaneArcanus {
        return model.ArcanusMap
    }

    return model.MyrrorMap
}

func (model *GameModel) SwitchPlane() {
    switch model.Plane {
        case data.PlaneArcanus: model.Plane = data.PlaneMyrror
        case data.PlaneMyrror: model.Plane = data.PlaneArcanus
    }
}

/* initial casting skill power is computed as follows:
 * skill = total number of magic books * 2
 * power = (skill-1)^2 + skill
 */
func computeInitialCastingSkillPower(books []data.WizardBook) int {
    total := 0
    for _, book := range books {
        total += book.Count
    }

    if total == 0 {
        return 0
    }

    total *= 2

    v := total - 1

    return v * v + total
}

func (model *GameModel) AddPlayer(wizard setup.WizardCustom, human bool) *playerlib.Player {
    useNames := model.heroNames[len(model.Players)]
    if useNames == nil {
        useNames = make(map[herolib.HeroType]string)
    }

    newPlayer := playerlib.MakePlayer(wizard, human, model.CurrentMap().Width(), model.CurrentMap().Height(), useNames, model)

    if !human {
        newPlayer.AIBehavior = ai.MakeEnemy2AI()

        switch model.AIMode {
            case ai.AIEnemy2: newPlayer.AIBehavior = ai.MakeEnemy2AI()
            case ai.AINet: newPlayer.AIBehavior = ai.MakeEnemyNetAI()
        }

        newPlayer.StrategicCombat = true
    }

    startingSpells := []string{"Magic Spirit", "Spell of Return"}
    if wizard.RetortEnabled(data.RetortArtificer) {
        startingSpells = append(startingSpells, "Enchant Item", "Create Artifact")
    }

    newPlayer.ResearchPoolSpells = wizard.StartingSpells.Copy()

    // not sure its necessary to add the starting spells to the research pool
    for _, spell := range startingSpells {
        newPlayer.ResearchPoolSpells.AddSpell(model.allSpells.FindByName(spell))
    }

    // every wizard gets all arcane spells by default
    newPlayer.ResearchPoolSpells.AddAllSpells(model.allSpells.GetSpellsByMagic(data.ArcaneMagic))

    newPlayer.KnownSpells = wizard.StartingSpells.Copy()
    for _, spell := range startingSpells {
        newPlayer.KnownSpells.AddSpell(model.allSpells.FindByName(spell))
    }
    newPlayer.CastingSkillPower = computeInitialCastingSkillPower(newPlayer.Wizard.Books)

    newPlayer.InitializeResearchableSpells(&model.allSpells)
    newPlayer.UpdateResearchCandidates()

    // log.Printf("Research spells: %v", newPlayer.ResearchPoolSpells)

    // famous wizards get a head start of 10 fame
    if wizard.RetortEnabled(data.RetortFamous) {
        newPlayer.Fame += 10
    }

    model.Players = append(model.Players, newPlayer)
    return newPlayer
}

// true if any alive player has the given enchantment enabled
func (model *GameModel) HasEnchantment(enchantment data.Enchantment) bool {
    for _, player := range model.Players {
        if !player.Defeated && player.HasEnchantment(enchantment) {
            return true
        }
    }

    return false
}

// true if any alive player that is not the given one has the given enchantment enabled
func (model *GameModel) HasRivalEnchantment(original *playerlib.Player, enchantment data.Enchantment) bool {
    for _, player := range model.Players {
        if !player.Defeated && player != original && player.HasEnchantment(enchantment) {
            return true
        }
    }

    return false
}

func (model *GameModel) GetMap(plane data.Plane) *maplib.Map {
    switch plane {
        case data.PlaneArcanus: return model.ArcanusMap
        case data.PlaneMyrror: return model.MyrrorMap
    }

    return nil
}

func (model *GameModel) FindPath(oldX int, oldY int, newX int, newY int, player *playerlib.Player, stack playerlib.PathStack, fog data.FogMap) (pathfinding.Path, bool) {

    useMap := model.GetMap(stack.Plane())

    if newY < 0 || newY >= useMap.Height() {
        return nil, false
    }

    oldX = useMap.WrapX(oldX)
    newX = useMap.WrapX(newX)

    if oldX == newX && oldY == newY {
        return nil, true
    }

    allFlyers := stack.AllFlyers()

    // this is to avoid doing path finding at all so that we don't spend time trying to compute an impossible path
    // such as a water unit trying to move to land
    if fog.GetFog(useMap.WrapX(newX), newY) != data.FogTypeUnexplored {
        tileTo := useMap.GetTile(newX, newY)
        tileFrom := useMap.GetTile(oldX, oldY)

        // if this stack contains a water-bound unit (a ship) that cannot walk on
        // land then the move onto land is impossible. Check ALL units, not just
        // the active ones: AI movement does not split inactive units off the
        // stack (see doAiMoveUnit), so an inactive ship must not be dragged
        // ashore when the active land units walk inland.
        if tileTo.Tile.IsLand() && !stack.CanMoveOnLand(false) {
            return nil, false
        }

        // if this is a water unit and it is moving from water to more water, but the destination water tile
        // is landlocked and the origin tile is not part of the same body of water, then there cannot be a valid
        // path between the two water tiles
        if tileTo.Tile.IsWater() && tileFrom.Tile.IsWater() && !stack.CanMoveOnLand(true) {
            if model.GetWaterBody(useMap, newX, newY) != model.GetWaterBody(useMap, oldX, oldY) {
                return nil, false
            }
        }

        /*
        if !tileTo.Tile.IsLand() && !allFlyers && stack.AnyLandWalkers() {

            // the stack might already contain a sailing unit
            if !stack.HasSailingUnits(true) {
                maybeStack := player.FindStack(useMap.WrapX(newX), newY, stack.Plane())
                if maybeStack != nil && maybeStack.HasSailingUnits(false) {
                    // ok, can move there because there is a ship
                } else {
                    return nil
                }
            }
        }
        */
    }

    normalized := func (a image.Point) image.Point {
        return image.Pt(useMap.WrapX(a.X), a.Y)
    }

    // check equality of two points taking wrapping into account
    tileEqual := func (a image.Point, b image.Point) bool {
        return normalized(a) == normalized(b)
    }

    getStack := func (x int, y int) (playerlib.PathStack, bool) {
        found := player.FindStack(x, y, stack.Plane())
        return found, found != nil
    }

    // cache locations of enemies
    enemyStacks := make(map[image.Point]struct{})
    enemyCities := make(map[image.Point]struct{})

    for _, enemy := range model.Players {
        if enemy != player {
            for _, enemyStack := range enemy.Stacks {
                enemyStacks[image.Pt(enemyStack.X(), enemyStack.Y())] = struct{}{}
            }
            for _, enemyCity := range enemy.Cities {
                enemyCities[image.Pt(enemyCity.X, enemyCity.Y)] = struct{}{}
            }
        }
    }

    // cache the containsEnemy result
    // true if the given coordinates contain an enemy unit or city
    containsEnemy := functional.Memoize2(func (x int, y int) bool {
        _, ok := enemyStacks[image.Pt(x, y)]
        if ok {
            return true
        }
        _, ok = enemyCities[image.Pt(x, y)]
        if ok {
            return true
        }

        return false
    })

    tileCost := func (x1 int, y1 int, x2 int, y2 int) float64 {
        x1 = useMap.WrapX(x1)
        x2 = useMap.WrapX(x2)

        if x1 < 0 || x1 >= useMap.Width() || y1 < 0 || y1 >= useMap.Height() {
            return pathfinding.Infinity
        }

        if x2 < 0 || x2 >= useMap.Width() || y2 < 0 || y2 >= useMap.Height() {
            return pathfinding.Infinity
        }

        // FIXME: it might be more optimal to put the infinity cases into the neighbors function instead

        // avoid encounters
        encounter := useMap.GetEncounter(x2, y2)
        if encounter != nil {
            if !tileEqual(image.Pt(x2, y2), image.Pt(newX, newY)) {
                return pathfinding.Infinity
            }
        }

        // avoid enemy units/cities
        if !tileEqual(image.Pt(x2, y2), image.Pt(newX, newY)) && containsEnemy(x2, y2) {
            return pathfinding.Infinity
        }

        baseCost := float64(1)

        if x1 != x2 && y1 != y2 {
            baseCost = 1.5
        }

        // don't know what the cost is, assume we can move there
        // FIXME: how should this behave for different FogTypes?
        // flying stacks can move anywhere, so don't penalize them
        if !allFlyers && x2 >= 0 && x2 < len(fog) && y2 >= 0 && y2 < len(fog[x2]) && fog[x2][y2] == data.FogTypeUnexplored {
            // increase cost of unknown tile by a lot so we prefer to move to known tiles
            return baseCost + 3
        }

        cost, ok := model.ComputeTerrainCost(stack, x1, y1, x2, y2, useMap, getStack)
        if !ok {
            return pathfinding.Infinity
        }

        return cost.ToFloat()
    }

    neighbors := func (x int, y int) []image.Point {
        return pathfinding.NeighborsWrapX(x, y, useMap.Height(), useMap.WrapX)
    }

    path, ok := pathfinding.FindPath(image.Pt(oldX, oldY), image.Pt(newX, newY), 10000, tileCost, neighbors, tileEqual)
    if ok {
        // ignore the start point
        return path[1:], true
    }

    return nil, false
}

func (model *GameModel) AllCities() []*citylib.City {
    var out []*citylib.City

    for _, player := range model.Players {
        for _, city := range player.Cities {
            out = append(out, city)
        }
    }

    return out
}

func (model *GameModel) GetWaterBody(mapUse *maplib.Map, x int, y int) *set.Set[image.Point] {
    if model.WaterBodies == nil {
        model.WaterBodies = make(map[data.Plane][]*set.Set[image.Point])
    }

    sets, ok := model.WaterBodies[mapUse.Plane]
    if !ok {
        sets = mapUse.GetWaterBodies()

        // log.Printf("Found %d water bodies on plane %d", len(sets), mapUse.Plane)

        model.WaterBodies[mapUse.Plane] = sets
    }

    find := image.Pt(mapUse.WrapX(x), y)

    // find the body of water that contains the given tile
    for _, set := range sets {
        if set.Contains(find) {
            return set
        }
    }

    return nil
}

/* return the cost to move from the current position the stack is on to the new given coordinates.
 * also return true/false if the move is even possible
 * FIXME: some values used by this logic could be precomputed and passed in as an argument. Things like 'containsFriendlyCity' could be a map of all cities
 * on the same plane as the unit, thus avoiding the expensive player.FindCity() call
 */
func (model *GameModel) ComputeTerrainCost(stack playerlib.PathStack, sourceX int, sourceY int, destX int, destY int, mapUse *maplib.Map, getStack func(int, int) (playerlib.PathStack, bool)) (fraction.Fraction, bool) {
    /*
    if stack.OutOfMoves() {
        return fraction.Zero(), false
    }
    */

    if sourceX == destX && sourceY == destY {
        return fraction.Zero(), true
    }

    tileFrom := mapUse.GetTile(sourceX, sourceY)
    tileTo := mapUse.GetTile(destX, destY)

    if !tileTo.Valid() {
        return fraction.Zero(), false
    }

    if stack.AllFlyers() {
        return fraction.FromInt(1), true
    }

    // can't move from land to ocean unless all units are flyers or swimmers
    if /* tileFrom.Tile.IsLand() && */ !tileTo.Tile.IsLand() {
        // a land walker can move onto a friendly stack on the ocean if that stack has sailing units
        if stack.AnyLandWalkers() {
            // if the stack already contains a sailing unit, then it is legal to move into water
            if stack.HasSailingUnits(true) {
                return fraction.FromInt(1), true
            }

            maybeStack, ok := getStack(destX, destY)
            if ok && maybeStack.HasSailingUnits(false) {
                return fraction.FromInt(1), true
            }
            return fraction.Zero(), false
        }
        /*
        if !stack.AllFlyers() && !stack.AllSwimmers() {
            return fraction.Zero(), false
        }
        */
    }

    containsFriendlyCity := func (x int, y int) bool {
        for _, player := range model.Players {
            if player.GetBanner() == stack.GetBanner() {
                if player.FindCity(x, y, stack.Plane()) != nil {
                    return true
                }
            }
        }

        return false
    }

    // sailing units cannot move onto land. Check ALL units (not just active
    // ones) so an inactive ship in the stack is never dragged ashore when the
    // active land units walk onto land.
    if tileTo.Tile.IsLand() {
        if !stack.CanMoveOnLand(false) {
            return fraction.Zero(), false
        }
    }

    // this feels like it can be improved
    if tileFrom.Tile.IsWater() && tileTo.Tile.IsWater() && !stack.CanMoveOnLand(true) {
        dx := mapUse.XDistance(sourceX, destX)
        dy := destY - sourceY
        if !tileFrom.CanTraverse(terrain.ToDirection(dx, dy), maplib.TraverseWater) {
            return fraction.Zero(), false
        }
    }

    road_v, ok := tileTo.Extras[maplib.ExtraKindRoad]
    if ok {
        road := road_v.(*maplib.ExtraRoad)
        if road.Enchanted {
            if stack.ActiveUnitsDoesntHaveAbility(data.AbilityNonCorporeal) {
                return fraction.Zero(), true
            }
        }

        return fraction.Make(1, 2), true
    }

    if containsFriendlyCity(destX, destY) {
        return fraction.Make(1, 2), true
    }

    if stack.HasPathfinding() {
        return fraction.Make(1, 2), true
    }

    // FIXME: handle swimming, sailing properties
    switch tileTo.Tile.TerrainType() {
        case terrain.Desert: return fraction.FromInt(1), true
        case terrain.SorceryNode: return fraction.FromInt(1), true
        case terrain.Grass: return fraction.FromInt(1), true
        case terrain.Forest:
            if stack.ActiveUnitsHasAbility(data.AbilityForester) {
                return fraction.FromInt(1), true
            }
            return fraction.FromInt(2), true
        case terrain.River: return fraction.FromInt(2), true
        case terrain.Tundra: return fraction.FromInt(2), true
        case terrain.Hill:
            if stack.ActiveUnitsHasAbility(data.AbilityMountaineer) {
                return fraction.FromInt(1), true
            }
            return fraction.FromInt(3), true
        case terrain.Swamp: return fraction.FromInt(3), true
        case terrain.Mountain:
            if stack.ActiveUnitsHasAbility(data.AbilityMountaineer) {
                return fraction.FromInt(1), true
            }
            return fraction.FromInt(4), true
        case terrain.Volcano:
            if stack.ActiveUnitsHasAbility(data.AbilityMountaineer) {
                return fraction.FromInt(1), true
            }

            return fraction.FromInt(4), true
    }

    return fraction.FromInt(1), true
}

func (model *GameModel) ComputeCityStackInfo() playerlib.CityStackInfo {
    out := playerlib.CityStackInfo{
        ArcanusStacks: make(map[image.Point]*playerlib.UnitStack),
        MyrrorStacks: make(map[image.Point]*playerlib.UnitStack),
        ArcanusCities: make(map[image.Point]*citylib.City),
        MyrrorCities: make(map[image.Point]*citylib.City),
    }

    for _, player := range model.Players {
        for _, stack := range player.Stacks {
            switch stack.Plane() {
                case data.PlaneArcanus: out.ArcanusStacks[image.Pt(stack.X(), stack.Y())] = stack
                case data.PlaneMyrror: out.MyrrorStacks[image.Pt(stack.X(), stack.Y())] = stack
            }
        }

        for _, city := range player.Cities {
            switch city.Plane {
                case data.PlaneArcanus: out.ArcanusCities[image.Pt(city.X, city.Y)] = city
                case data.PlaneMyrror: out.MyrrorCities[image.Pt(city.X, city.Y)] = city
            }
        }
    }

    return out
}

// return the city and its owner
func (model *GameModel) FindCity(x int, y int, plane data.Plane) (*citylib.City, *playerlib.Player) {
    for _, player := range model.Players {
        city := player.FindCity(x, y, plane)
        if city != nil {
            return city, player
        }
    }

    return nil, nil
}

func (model *GameModel) FindSettlableLocations(x int, y int, plane data.Plane, fog data.FogMap) []image.Point {
    tiles := model.GetMap(plane).GetContinentTiles(x, y)

    // compute all pointes that we can't build a city on because they are too close to another city
    unavailable := make(map[image.Point]bool)
    for _, city := range model.AllCities() {
        if city.Plane == plane {
            // keep a distance of 5 tiles from any other city
            for dx := -5; dx <= 5; dx++ {
                for dy := -5; dy <= 5; dy++ {
                    cx := model.CurrentMap().WrapX(city.X + dx)
                    cy := city.Y + dy

                    unavailable[image.Pt(cx, cy)] = true
                }
            }
        }
    }

    var out []image.Point

    for _, tile := range tiles {
        _, ok := unavailable[image.Pt(tile.X, tile.Y)]
        if ok {
            continue
        }

        if fog[tile.X][tile.Y] == data.FogTypeUnexplored {
            continue
        }

        if tile.Corrupted() || tile.HasEncounter() || tile.Tile.IsMagic() {
            continue
        }

        out = append(out, image.Pt(tile.X, tile.Y))
    }

    return out
}

func (model *GameModel) FindStack(x int, y int, plane data.Plane) (*playerlib.UnitStack, *playerlib.Player) {
    for _, player := range model.Players {
        stack := player.FindStack(x, y, plane)
        if stack != nil {
            return stack, player
        }
    }

    return nil, nil
}

func (model *GameModel) GetDifficulty() data.DifficultySetting {
    return model.Settings.Difficulty
}

func (model *GameModel) GetAggressiveAI() bool {
    if model.IsAggressiveAI == nil {
        return false
    }
    return model.IsAggressiveAI()
}

func (model *GameModel) GetTurnNumber() uint64 {
    return model.TurnNumber
}

/* true if a settler can build a city here
 * a tile must be land, not corrupted, not have an encounter, not have a magic node, and not be too close to another city
 */
func (model *GameModel) IsSettlableLocation(x int, y int, plane data.Plane) bool {
    if !model.NearCity(image.Pt(x, y), 3, plane) {
        mapUse := model.GetMap(plane)
        if mapUse.HasCorruption(x, y) || mapUse.GetEncounter(x, y) != nil || mapUse.GetMagicNode(x, y) != nil {
            return false
        }

        return mapUse.GetTile(x, y).Tile.IsLand()
    }

    return false
}

func (model *GameModel) NearCity(point image.Point, squares int, plane data.Plane) bool {
    for _, city := range model.AllCities() {
        if city.Plane == plane {
            xDiff := model.CurrentMap().XDistance(city.X, point.X)
            yDiff := city.Y - point.Y

            if xDiff < 0 {
                xDiff = -xDiff
            }

            if yDiff < 0 {
                yDiff = -yDiff
            }

            if xDiff <= squares && yDiff <= squares {
                return true
            }
        }
    }

    return false
}

// find all engineers that are currently building a road
// compute the work done by each engineer according to the terrain
//   total work = work per engineer ^ engineers building on that tile
// add total work to some counter, and when that total reaches the threshold for the terrain type
// then set a road on that tile and make the engineers no longer busy
func (model *GameModel) DoBuildRoads(player *playerlib.Player) {
    // keeps track of roads that are still being built
    arcanusBuilds := make(map[image.Point]struct{})
    myrrorBuilds := make(map[image.Point]struct{})

    for _, stack := range slices.Clone(player.Stacks) {
        plane := stack.Plane()

        engineerCount := ComputeEngineerCount(stack, true)

        if engineerCount > 0 {
            x, y := stack.X(), stack.Y()
            roads := player.RoadWorkArcanus
            if plane == data.PlaneMyrror {
                roads = player.RoadWorkMyrror
            }

            amount, ok := roads[image.Pt(x, y)]
            if !ok {
                amount = 0
            }

            tileWork := model.ComputeRoadBuildEffort(x, y, plane) // just to get the work map

            amount += math.Pow(tileWork.WorkPerEngineer, float64(engineerCount))
            if amount >= tileWork.TotalWork {
                model.GetMap(plane).SetRoad(x, y, plane == data.PlaneMyrror)

                player.BuiltRoad(x, y, plane)

                for _, unit := range stack.Units() {
                    if unit.GetBusy() == units.BusyStatusBuildRoad {
                        unit.SetBusy(units.BusyStatusNone)
                    }
                }
            } else {
                roads[image.Pt(x, y)] = amount
                if plane == data.PlaneArcanus {
                    arcanusBuilds[image.Pt(x, y)] = struct{}{}
                } else {
                    myrrorBuilds[image.Pt(x, y)] = struct{}{}
                }
            }
        }
    }

    // remove all points that are no longer being built

    var toDelete []image.Point
    for point, _ := range player.RoadWorkArcanus {
        _, ok := arcanusBuilds[point]
        if !ok {
            toDelete = append(toDelete, point)
        }
    }

    for _, point := range toDelete {
        // log.Printf("remove point %v", point)
        delete(player.RoadWorkArcanus, point)
    }

    toDelete = nil
    for point, _ := range player.RoadWorkMyrror {
        _, ok := myrrorBuilds[point]
        if !ok {
            toDelete = append(toDelete, point)
        }
    }

    for _, point := range toDelete {
        // log.Printf("remove point %v", point)
        delete(player.RoadWorkMyrror, point)
    }

}

type RoadWork struct {
    WorkPerEngineer float64
    TotalWork float64
}

// returns how many turns it takes to build a road on the given tile with the given stack
func (model *GameModel) ComputeRoadBuildEffort(x int, y int, plane data.Plane) RoadWork {

    computeWork := func (oneEngineerTurn int, twoEngineerTurn int) RoadWork {
        workPerEngineer := float64(oneEngineerTurn) / float64(twoEngineerTurn)
        totalWork := float64(oneEngineerTurn) * workPerEngineer
        return RoadWork{WorkPerEngineer: workPerEngineer, TotalWork: totalWork}
    }

    work := make(map[terrain.TerrainType]RoadWork)
    work[terrain.Grass] = computeWork(3, 1)
    work[terrain.Desert] = computeWork(4, 2)
    work[terrain.River] = computeWork(5, 2)
    work[terrain.Forest] = computeWork(6, 3)
    work[terrain.Tundra] = computeWork(6, 3)
    work[terrain.Hill] = computeWork(6, 3)
    work[terrain.Swamp] = computeWork(8, 4)
    work[terrain.Mountain] = computeWork(8, 4)
    work[terrain.Volcano] = computeWork(8, 4)
    work[terrain.ChaosNode] = computeWork(8, 4)
    work[terrain.NatureNode] = computeWork(5, 2)
    work[terrain.SorceryNode] = computeWork(4, 2)

    tile := model.GetMap(plane).GetTile(x, y)

    return work[tile.Tile.TerrainType()]
}

func (model *GameModel) DoPurify(player *playerlib.Player) {
    type PurifyWork struct {
        WorkPerUnit float64
        TotalWork float64
    }

    computeWork := func (oneUnitTurn int, twoUnitTurns int) PurifyWork {
        workPerUnit := float64(oneUnitTurn) / float64(twoUnitTurns)
        totalWork := float64(oneUnitTurn) * workPerUnit
        return PurifyWork{WorkPerUnit: workPerUnit, TotalWork: totalWork}
    }

    work := computeWork(5, 3)

    arcanusBuilds := make(map[image.Point]struct{})
    myrrorBuilds := make(map[image.Point]struct{})

    for _, stack := range player.Stacks {
        plane := stack.Plane()

        unitCount := 0
        for _, unit := range stack.Units() {
            if unit.GetBusy() == units.BusyStatusPurify {
                unitCount += 1
            }
        }

        if unitCount > 0 {
            x, y := stack.X(), stack.Y()
            // log.Printf("building a road at %v, %v with %v engineers", x, y, engineerCount)
            purify := player.PurifyWorkArcanus
            if plane == data.PlaneMyrror {
                purify = player.PurifyWorkMyrror
            }

            amount, ok := purify[image.Pt(x, y)]
            if !ok {
                amount = 0
            }

            amount += math.Pow(work.WorkPerUnit, float64(unitCount))
            // log.Printf("  amount is now %v. total work is %v", amount, tileWork.TotalWork)
            if amount >= work.TotalWork {
                model.GetMap(plane).RemoveCorruption(x, y)

                for _, unit := range stack.Units() {
                    if unit.GetBusy() == units.BusyStatusPurify {
                        unit.SetBusy(units.BusyStatusNone)
                    }
                }

            } else {
                purify[image.Pt(x, y)] = amount
                if plane == data.PlaneArcanus {
                    arcanusBuilds[image.Pt(x, y)] = struct{}{}
                } else {
                    myrrorBuilds[image.Pt(x, y)] = struct{}{}
                }
            }
        }
    }

    // remove all points that are no longer being built

    var toDelete []image.Point
    for point, _ := range player.PurifyWorkArcanus {
        _, ok := arcanusBuilds[point]
        if !ok {
            toDelete = append(toDelete, point)
        }
    }

    for _, point := range toDelete {
        // log.Printf("remove point %v", point)
        delete(player.PurifyWorkArcanus, point)
    }

    toDelete = nil
    for point, _ := range player.PurifyWorkMyrror {
        _, ok := myrrorBuilds[point]
        if !ok {
            toDelete = append(toDelete, point)
        }
    }

    for _, point := range toDelete {
        // log.Printf("remove point %v", point)
        delete(player.PurifyWorkMyrror, point)
    }
}

func (model *GameModel) ManaShortActive() bool {
    return slices.ContainsFunc(model.RandomEvents, func(event *RandomEvent) bool {
        return event.Type == RandomEventManaShort
    })
}

func (model *GameModel) PopulationBoomActive(city *citylib.City) bool {
    return slices.ContainsFunc(model.RandomEvents, func(event *RandomEvent) bool {
        return event.Type == RandomEventPopulationBoom && event.TargetCity == city
    })
}

func (model *GameModel) PlagueActive(city *citylib.City) bool {
    return slices.ContainsFunc(model.RandomEvents, func(event *RandomEvent) bool {
        return event.Type == RandomEventPlague && event.TargetCity == city
    })
}

func (model *GameModel) GoodMoonActive() bool {
    return slices.ContainsFunc(model.RandomEvents, func(event *RandomEvent) bool {
        return event.Type == RandomEventGoodMoon
    })
}

func (model *GameModel) BadMoonActive() bool {
    return slices.ContainsFunc(model.RandomEvents, func(event *RandomEvent) bool {
        return event.Type == RandomEventBadMoon
    })
}

func (model *GameModel) ConjunctionChaosActive() bool {
    return slices.ContainsFunc(model.RandomEvents, func(event *RandomEvent) bool {
        return event.Type == RandomEventConjunctionChaos
    })
}

func (model *GameModel) ConjunctionNatureActive() bool {
    return slices.ContainsFunc(model.RandomEvents, func(event *RandomEvent) bool {
        return event.Type == RandomEventConjunctionNature
    })
}

func (model *GameModel) ConjunctionSorceryActive() bool {
    return slices.ContainsFunc(model.RandomEvents, func(event *RandomEvent) bool {
        return event.Type == RandomEventConjunctionSorcery
    })
}

func (model *GameModel) DoRandomEvents() {
    // maybe create a new event
    eventModifier := fraction.FromInt(1)
    switch model.Settings.Difficulty {
        case data.DifficultyIntro: eventModifier = fraction.Make(1, 2)
        case data.DifficultyEasy: eventModifier = fraction.Make(2, 3)
        case data.DifficultyAverage: eventModifier = fraction.Make(3, 4)
        case data.DifficultyHard: eventModifier = fraction.Make(4, 5)
        case data.DifficultyExtreme: eventModifier = fraction.Make(1, 1)
        case data.DifficultyImpossible: eventModifier = fraction.Make(6, 5)
    }

    // for testing purposes
    // eventModifier = fraction.FromInt(10)

    eventProbability := fraction.FromInt(int(model.TurnNumber - model.LastEventTurn)).Multiply(eventModifier)
    if model.TurnNumber < 50 || model.TurnNumber - model.LastEventTurn < 5 {
        eventProbability = fraction.Zero()
    }

    if rand.N(512) < int(eventProbability.ToFloat()) {
        choices := set.NewSet[RandomEventType](
            RandomEventBadMoon,
            RandomEventConjunctionChaos,
            RandomEventConjunctionNature,
            RandomEventConjunctionSorcery,
            RandomEventDepletion,
            RandomEventDiplomaticMarriage,
            RandomEventDisjunction,
            RandomEventDonation,
            RandomEventEarthquake,
            RandomEventGift,
            RandomEventGoodMoon,
            RandomEventGreatMeteor,
            RandomEventManaShort,
            RandomEventNewMinerals,
            RandomEventPiracy,
            RandomEventPlague,
            RandomEventPopulationBoom,
            RandomEventRebellion,
        )

        // remove events that can't occur because they are already occurring or
        // there is some mutually exclusive other event
        for _, event := range model.RandomEvents {
            choices.Remove(event.Type)
            // remove all conjunctions because only one conjunction can be active at a time
            if event.IsConjunction {
                choices.Remove(RandomEventBadMoon)
                choices.Remove(RandomEventGoodMoon)
                choices.Remove(RandomEventConjunctionChaos)
                choices.Remove(RandomEventConjunctionNature)
                choices.Remove(RandomEventConjunctionSorcery)
                choices.Remove(RandomEventManaShort)
            }
        }

        if model.TurnNumber < 150 {
            choices.Remove(RandomEventDiplomaticMarriage)
            choices.Remove(RandomEventGreatMeteor)
        }

        if choices.Size() > 0 {
            choice := choices.Values()[rand.N(choices.Size())]

            // return a RandomEvent object to show, and also cause the event to occur (if instant)
            makeEvent := func (choice RandomEventType, target *playerlib.Player) (*RandomEvent, GameEvent) {
                usedCities := set.NewSet[*citylib.City]()
                for _, event := range model.RandomEvents {
                    if event.TargetCity != nil {
                        usedCities.Insert(event.TargetCity)
                    }
                }

                switch choice {
                    case RandomEventBadMoon: return MakeBadMoonEvent(model.TurnNumber), nil
                    case RandomEventGoodMoon: return MakeGoodMoonEvent(model.TurnNumber), nil
                    case RandomEventConjunctionChaos: return MakeConjunctionChaosEvent(model.TurnNumber), nil
                    case RandomEventConjunctionNature: return MakeConjunctionNatureEvent(model.TurnNumber), nil
                    case RandomEventConjunctionSorcery: return MakeConjunctionSorceryEvent(model.TurnNumber), nil
                    case RandomEventManaShort: return MakeManaShortEvent(model.TurnNumber), nil
                    case RandomEventDisjunction:
                        // there must be at least one global enchantment for this event to occur
                        hasGlobalEnchantment := false

                        for _, player := range model.Players {
                            if player.GlobalEnchantments.Size() > 0 {
                                hasGlobalEnchantment = true
                                break
                            }
                        }

                        if !hasGlobalEnchantment {
                            return nil, nil
                        }

                        // remove all global enchantments
                        for _, player := range model.Players {
                            player.GlobalEnchantments.Clear()
                        }

                        return MakeDisjunctionEvent(model.TurnNumber), nil
                    case RandomEventDonation:
                        // FIXME: what are the bounds here?
                        gold := rand.N(2000) + 100
                        target.Gold += gold

                        return MakeDonationEvent(model.TurnNumber, gold, target), nil
                    case RandomEventPiracy:
                        if target.Gold < 100 {
                            return nil, nil
                        }

                        // between 30-50%, compute random number between 0-20%, add 30%
                        gold := rand.N(target.Gold / 5) + target.Gold * 3 / 10
                        target.Gold = max(0, target.Gold - gold)

                        return MakePiracyEvent(model.TurnNumber, gold, target), nil
                    case RandomEventGift:
                        var out []*artifact.Artifact
                        for _, candidate := range model.ArtifactPool.AvailableArtifacts() {
                            if canUseArtifact(candidate, target.Wizard) {
                                out = append(out, candidate)
                            }
                        }

                        // couldn't find a valid artifact
                        if len(out) == 0 {
                            return nil, nil
                        }

                        use := out[rand.N(len(out))]

                        model.ArtifactPool.Award(use)

                        // returning GameEventVault here is ugly but we need a way to have the vault event
                        // be added to game.Events after the random event
                        return MakeGiftEvent(model.TurnNumber, use.Name, target), &GameEventVault{CreatedArtifact: use, Player: target}
                    case RandomEventDepletion:
                        // choose a random town that has a mineral bonus in its catchment area,
                        // and then remove the bonus from the map
                        cities := target.GetCities()
                        for _, cityIndex := range rand.Perm(len(cities)) {
                            city := cities[cityIndex]
                            mapUse := model.GetMap(city.Plane)
                            catchment := mapUse.GetCatchmentArea(city.X, city.Y)
                            var choices []maplib.FullTile
                            for _, tile := range catchment {
                                switch tile.GetBonus() {
                                    case data.BonusSilverOre, data.BonusGoldOre, data.BonusIronOre, data.BonusCoal,
                                         data.BonusMithrilOre, data.BonusAdamantiumOre, data.BonusGem:
                                        choices = append(choices, tile)
                                }
                            }

                            if len(choices) > 0 {
                                tile := choices[rand.N(len(choices))]
                                mapUse.RemoveBonus(tile.X, tile.Y)
                                return MakeDepletionEvent(model.TurnNumber, tile.GetBonus(), city.Name), nil
                            }
                        }

                        return nil, nil

                    case RandomEventDiplomaticMarriage:
                        neutral := model.GetNeutralPlayer()
                        if neutral != nil {
                            if len(neutral.Cities) > 0 {
                                cities := neutral.GetCities()
                                city := cities[rand.N(len(cities))]
                                // if the owner of the city has a stack garrisoned there then the garrison is disbanded
                                stack := neutral.FindStack(city.X, city.Y, city.Plane)
                                if stack != nil {
                                    for _, unit := range stack.Units() {
                                        neutral.RemoveUnit(unit)
                                    }
                                }

                                ChangeCityOwner(city, neutral, target, ChangeCityRemoveAllEnchantments)

                                return MakeDiplomaticMarriageEvent(model.TurnNumber, city), nil
                            }
                        }

                        return nil, nil

                    case RandomEventEarthquake:
                        choices := model.AllCities()
                        if len(choices) == 0 {
                            return nil, nil
                        }

                        city := choices[rand.N(len(choices))]

                        people, units, buildings := model.DoEarthquake(city, target)

                        return MakeEarthquakeEvent(model.TurnNumber, city.Name, people, units, len(buildings)), nil

                    case RandomEventGreatMeteor:
                        choices := model.AllCities()
                        if len(choices) == 0 {
                            return nil, nil
                        }

                        city := choices[rand.N(len(choices))]

                        people, units, buildings := model.doCallTheVoid(city, target)

                        return MakeGreatMeteorEvent(model.TurnNumber, city.Name, people, units, buildings), nil

                    case RandomEventNewMinerals:
                        cities := target.GetCities()
                        for _, cityIndex := range rand.Perm(len(cities)) {
                            city := cities[cityIndex]
                            mapUse := model.GetMap(city.Plane)
                            catchment := mapUse.GetCatchmentArea(city.X, city.Y)
                            var choices []maplib.FullTile
                            for _, tile := range catchment {
                                terrainType := tile.Tile.TerrainType()
                                if tile.GetBonus() == data.BonusNone && (terrainType == terrain.Hill || terrainType == terrain.Mountain) {
                                    choices = append(choices, tile)
                                }
                            }

                            if len(choices) > 0 {
                                tile := choices[rand.N(len(choices))]

                                bonusChoices := []data.BonusType{data.BonusGoldOre, data.BonusCoal, data.BonusMithrilOre, data.BonusAdamantiumOre, data.BonusGem}
                                bonus := bonusChoices[rand.N(len(bonusChoices))]

                                mapUse.SetBonus(tile.X, tile.Y, bonus)
                                return MakeNewMineralsEvent(model.TurnNumber, bonus, city), nil
                            }
                        }

                    case RandomEventPlague:
                        cities := target.GetCities()
                        for _, cityIndex := range rand.Perm(len(cities)) {
                            city := cities[cityIndex]
                            if !usedCities.Contains(city) {
                                return MakePlagueEvent(model.TurnNumber, city), nil
                            }
                        }

                        return nil, nil

                    case RandomEventPopulationBoom:
                        cities := target.GetCities()
                        for _, cityIndex := range rand.Perm(len(cities)) {
                            city := cities[cityIndex]
                            if !usedCities.Contains(city) {
                                return MakePopulationBoomEvent(model.TurnNumber, city), nil
                            }
                        }

                        return nil, nil

                    case RandomEventRebellion:
                        if len(target.Cities) == 0 {
                            return nil, nil
                        }

                        var neutralPlayer *playerlib.Player
                        for _, neutral := range model.Players {
                            if neutral.GetBanner() == data.BannerBrown {
                                neutralPlayer = neutral
                                break
                            }
                        }

                        if neutralPlayer != nil {
                            var choices []*citylib.City
                            for _, city := range target.Cities {
                                if city.HasFortress() || city.HasSummoningCircle() {
                                    continue
                                }

                                // cannot target a city with a hero in it
                                stack := target.FindStack(city.X, city.Y, city.Plane)
                                if stack != nil && stack.HasHero() {
                                    continue
                                }

                                choices = append(choices, city)
                            }

                            if len(choices) > 0 {
                                city := choices[rand.N(len(choices))]

                                // disband any fantastic units garrisoned at the city, and convert to neutral all other normal units
                                stack := target.FindStack(city.X, city.Y, city.Plane)
                                if stack != nil {
                                    for _, unit := range stack.Units() {
                                        target.RemoveUnit(unit)
                                        if unit.GetRace() != data.RaceFantastic {
                                            unit.SetBanner(neutralPlayer.GetBanner())
                                            neutralPlayer.AddUnit(unit)
                                        }
                                    }
                                }

                                ChangeCityOwner(city, target, neutralPlayer, ChangeCityRemoveAllEnchantments)

                                // plague/population boom might still be active for the city. just leave them for now

                                return MakeRebellionEvent(model.TurnNumber, city), nil
                            }
                        }

                        return nil, nil
                }

                return nil, nil
            }

            validTargets := slices.DeleteFunc(slices.Clone(model.Players), func(player *playerlib.Player) bool {
                return player.Defeated || player.Skip || player.GetBanner() == data.BannerBrown
            })

            if len(validTargets) > 0 {
                targetWizard := validTargets[rand.N(len(validTargets))]
                newEvent, extraEvent := makeEvent(choice, targetWizard)
                if newEvent != nil {
                    model.LastEventTurn = model.TurnNumber

                    // log.Printf("Random event occurred: %+v", newEvent)

                    if !newEvent.Instant {
                        model.RandomEvents = append(model.RandomEvents, newEvent)
                    }

                    // FIXME: if the event is targeting an AI wizard then the event message should change slightly
                    model.Events <- &GameEventShowRandomEvent{Event: newEvent, Starting: true}

                    if extraEvent != nil {
                        model.Events <- extraEvent
                    }

                    model.RefreshUI()
                }
            }
        }
    }

    var keep []*RandomEvent
    // add events to the 'keep' array to keep them for the next turn
    for _, event := range model.RandomEvents {

        // once citizens has reached 2, plague will dissipate automatically
        if event.Type == RandomEventPlague && event.TargetCity.Citizens() <= 2 {
            model.Events <- &GameEventShowRandomEvent{Event: event, Starting: false}
            continue
        }

        // a random event can end after 5 turns, and the chances of it ending are 5% per turn
        turns := model.TurnNumber - event.BirthYear
        if turns < 5 {
            keep = append(keep, event)
            continue
        }
        step := uint64(5)
        if event.IsConjunction {
            step = 10
        }

        chance := (turns - 5) * step

        if uint64(rand.N(100)) < chance {
            // don't keep
            model.Events <- &GameEventShowRandomEvent{Event: event, Starting: false}
        } else {
            keep = append(keep, event)
        }
    }

    model.RandomEvents = keep
}

func (model *GameModel) DoNextTurn() {
    // if time stop is enabled then don't move to the other players, just keep doing the current player
    if model.CurrentPlayer >= 0 && model.Players[model.CurrentPlayer].HasEnchantment(data.EnchantmentTimeStop) {
        model.EndOfTurn()
    } else {
        model.CurrentPlayer += 1
        if model.CurrentPlayer >= len(model.Players) {
            // all players did their turn, so the next global turn starts
            model.EndOfTurn()
            model.CurrentPlayer = 0
        }
    }

    if len(model.Players) > 0 {
        player := model.Players[model.CurrentPlayer]

        aiPlayer := model.Players[model.CurrentPlayer]

        if aiPlayer.AIBehavior != nil {
            aiPlayer.AIBehavior.PreTurn(aiPlayer)
        }

        if player.Wizard.Banner != data.BannerBrown {
            model.StartPlayerTurn(player)
        } else {
            // neutral enemies should reset their moves each turn
            for _, stack := range player.Stacks {
                handleStasis(stack)
                stack.ResetMoves()
                stack.EnableMovers()
            }
        }

        if aiPlayer.AIBehavior != nil {
            aiPlayer.AIBehavior.NewTurn(aiPlayer)
        }
    }
}

/* return a tuple of booleans where each boolean is true if the corresponding resource is not enough
 * to support the units.
 * (gold, food, mana)
 *
 * (false, false, false) means all units are supported.
 */
func (model *GameModel) CheckDisband(player *playerlib.Player) (bool, bool, bool) {

    unitsNeedGold := false
    unitsNeedFood := false
    unitsNeedMana := false

    for unit := range player.Units() {
        // dont need to keep checking in this case
        if unitsNeedGold && unitsNeedFood && unitsNeedMana {
            break
        }

        unitsNeedGold = unitsNeedGold || unit.GetUpkeepGold() > 0
        unitsNeedFood = unitsNeedFood || unit.GetUpkeepFood() > 0
        unitsNeedMana = unitsNeedMana || unit.GetUpkeepMana() > 0
    }

    goldPerTurn := player.GoldPerTurn()
    goldIssue := player.Gold + goldPerTurn < 0 && unitsNeedGold
    foodIssue := player.FoodPerTurn() < 0 && unitsNeedFood

    // FIXME: can the power be passed in so it doesn't have to be computed multiple times?
    manaPerTurn := player.ManaPerTurn(model.ComputePower(player), model)

    manaIssue := player.Mana + manaPerTurn < 0 && unitsNeedMana

    return goldIssue, foodIssue, manaIssue
}

/* disband units due to lack of resources, return an array of messages about units that were lost
 */
func (model *GameModel) DisbandUnits(player *playerlib.Player) []string {
    /*
    if 2 > 1 {
        return nil
    }
    */

    // keep removing units until the upkeep value can be paid
    ok := false
    var disbandedMessages []string
    for player.UnitCount() > 0 && !ok {
        ok = true

        goldIssue, foodIssue, manaIssue := model.CheckDisband(player)

        if goldIssue || foodIssue || manaIssue {
            ok = false
            disbanded := false

            allUnits := slices.Collect(player.Units())

            // try to disband one unit that is taking up resources
            for i := len(allUnits) - 1; i >= 0; i-- {
                unit := allUnits[i]
                // disband the unit for the right reason
                if goldIssue && unit.GetUpkeepGold() > 0 {
                    log.Printf("Disband %v due to lack of gold", unit)
                    disbandedMessages = append(disbandedMessages, fmt.Sprintf("%v disbanded due to lack of gold", unit.GetName()))
                    player.RemoveUnit(unit)
                    disbanded = true
                    break
                }

                if foodIssue && unit.GetUpkeepFood() > 0 {
                    log.Printf("Disband %v due to lack of food", unit)
                    disbandedMessages = append(disbandedMessages, fmt.Sprintf("%v disbanded due to lack of food", unit.GetName()))
                    player.RemoveUnit(unit)
                    disbanded = true
                    break
                }

                if manaIssue && unit.GetUpkeepMana() > 0 {
                    log.Printf("Disband %v due to lack of mana", unit)
                    disbandedMessages = append(disbandedMessages, fmt.Sprintf("%v disbanded due to lack of mana", unit.GetName()))
                    player.RemoveUnit(unit)
                    disbanded = true
                    break
                }
            }

            if !disbanded {
                // fail safe to make sure we exit the loop in case somehow a unit was not disbanded
                break
            }

            var toRemove []units.StackUnit
            // check land walkers on ocean tiles that do not have valid transport
            // FIXME: handle not enough transports
            for _, stack := range player.Stacks {
                // if the stack can fly for whatever reason then no units will drown
                if stack.AllFlyers() {
                    continue
                }
                mapUse := model.GetMap(stack.Plane())
                hasTransport := stack.HasSailingUnits(false)
                if !hasTransport && !mapUse.GetTile(stack.X(), stack.Y()).Tile.IsLand() {
                    for _, unit := range stack.Units() {
                        if unit.IsLandWalker() {
                            toRemove = append(toRemove, unit)
                        }
                    }
                }
            }

            for _, unit := range toRemove {
                player.RemoveUnit(unit)
            }
        }
    }

    return disbandedMessages
}

// turn off enchantments that can not be afforded
func (model *GameModel) DissipateEnchantments(player *playerlib.Player, power int) {
    // if time stop is in effect then only check that spell for dissipation, all other spells have no upkeep
    if player.HasEnchantment(data.EnchantmentTimeStop) {
        if player.Mana < data.EnchantmentTimeStop.UpkeepMana() {
            player.RemoveEnchantment(data.EnchantmentTimeStop)
        }

        return
    }

    isManaIssue := func() bool {
        manaPerTurn := player.ManaPerTurn(power, model)
        return player.Mana + manaPerTurn < 0
    }

    // keep removing enchantments until there is no more mana issue
    for player.GlobalEnchantments.Size() > 0 && isManaIssue() {
        enchantments := player.GlobalEnchantments.Values()
        enchantment := enchantments[rand.N(len(enchantments))]
        player.GlobalEnchantments.Remove(enchantment)
    }

    // keep removing city enchantments until there is no more mana issue
    for {
        enchantments := model.GetCityEnchantmentsByBanner(player.GetBanner())
        if len(enchantments) == 0 || !isManaIssue() {
            break
        }

        enchantment := enchantments[rand.N(len(enchantments))]
        enchantment.City.CancelEnchantment(enchantment.Enchantment.Enchantment, enchantment.Enchantment.Owner)
    }

    var enchantedUnits []units.StackUnit
    for unit := range player.Units() {
        if len(unit.GetEnchantments()) > 0 {
            enchantedUnits = append(enchantedUnits, unit)
        }
    }

    for len(enchantedUnits) > 0 && isManaIssue() {
        unit := enchantedUnits[rand.N(len(enchantedUnits))]
        enchantments := unit.GetEnchantments()
        enchantment := enchantments[rand.N(len(enchantments))]
        unit.RemoveEnchantment(enchantment)
        if len(unit.GetEnchantments()) == 0 {
            enchantedUnits = slices.DeleteFunc(enchantedUnits, func(u units.StackUnit) bool {
                return u == unit
            })
        }
    }
}

// FIXME: add a "reason" of fizzling, like "Spell X was fizzled because of Y"
func (model *GameModel) ShowFizzleSpell(spell spellbook.Spell, caster *playerlib.Player) {
    if caster.IsHuman() {
        model.Events <- &GameEventNotice{
            // FIXME: align this message with how dos mom does it
            Message: fmt.Sprintf("The spell %v has fizzled.", spell.Name),
        }
    }
}

// FIXME: maybe this should be generalized so it can handle any enchantment that cuases a spell to fizzle?
func (model *GameModel) ShowTranquilityFizzle(tranquilityOwner *playerlib.Player, caster *playerlib.Player, spell spellbook.Spell) {
    event := GameEventTranquilityFizzle{
        TranquilityOwner: tranquilityOwner,
        Caster: caster,
        Spell: spell,
    }

    select {
        case model.Events <- &event:
        default:
    }
}

// the amount of experience a unit in a stack should get at the end of each turn
// if there is a hero in the stack then the hero's armsmaster ability applies
func (model *GameModel) GetExperienceBonus(stack *playerlib.UnitStack) int {
    base := 1
    bonus := 0

    // only the highest armsmaster bonus applies
    for _, unit := range stack.Units() {
        if unit.GetRace() == data.RaceHero {
            hero := unit.(*herolib.Hero)
            more := hero.GetAbilityExperienceBonus()
            if more > bonus {
                bonus = more
            }
        }
    }

    return base + bonus
}

func (model *GameModel) AddExperience(player *playerlib.Player, unit units.StackUnit, amount int) {
    if player.IsHuman() && unit.IsHero() {
        level_before := unit.GetHeroExperienceLevel()

        unit.AddExperience(amount)

        level_after := unit.GetHeroExperienceLevel()

        if level_before != level_after {
            hero := unit.(*herolib.Hero)
            model.Events <- &GameEventHeroLevelUp{
                Hero: hero,
            }
        }
    } else {
        unit.AddExperience(amount)
    }
}

/* random chance to create a hire hero event
 */
func (model *GameModel) maybeHireHero(player *playerlib.Player) {
    if len(player.AliveHeroes()) >= 6 {
        return
    }
    // chance as an integer between 0-100

    // every 25 fame increases chance by 1
    // every hero reduces chance by a fraction (1 hero = halve chance. 2 heroes = 1/3 chance)
    chance := (3 + player.GetFame() / 25) / ((len(player.AliveHeroes()) + 3) / 2)
    if player.Wizard.RetortEnabled(data.RetortFamous) {
        chance *= 2
    }

    if chance > 10 {
        chance = 10
    }

    if rand.N(100) < chance {
        var heroCandidates []*herolib.Hero
        for _, hero := range player.HeroPool {
            // torin can never be hired
            if hero.HeroType == herolib.HeroTorin {
                continue
            }

            if hero.Status == herolib.StatusAvailable {
                if hero.GetRequiredFame() <= player.GetFame() {
                    heroCandidates = append(heroCandidates, hero)
                }
            }
        }

        if len(heroCandidates) > 0 {
            hero := heroCandidates[rand.N(len(heroCandidates))]

            fee := hero.GetHireFee()
            if fee > player.Gold {
                // hero gains a level if the player can't afford to hire them
                hero.GainLevel(units.ExperienceChampionHero)
            } else {
                select {
                    case model.Events <- &GameEventHireHero{Cost: fee, Hero: hero, Player: player}:
                    default:
                }
            }
        }
    }
}

/* random chance to create a hire mercenaries event
 */
func (model *GameModel) maybeHireMercenaries(player *playerlib.Player) {
    if model.TurnNumber <= 30 {
        return
    }

    fortressCity := player.FindFortressCity()
    if fortressCity == nil {
        return  // player is being banished
    }

    // chance to create an event
    chance := 1 + player.GetFame() / 20
    if player.Wizard.RetortEnabled(data.RetortFamous) {
        chance *= 2
    }
    if chance > 10 {
        chance = 10
    }

    if rand.N(100) >= chance {
        return
    }

    // unit type
    presentOnArcanus := false
    presentOnMyrror := false
    for _, city := range player.Cities {
        if city.Plane == data.PlaneArcanus {
            presentOnArcanus = true
        }
        if city.Plane == data.PlaneMyrror {
            presentOnMyrror = true
        }
    }

    var unitCandidates []*units.Unit
    for _, unit := range units.AllUnits {
        if unit.Race == data.RaceFantastic || unit.Race == data.RaceHero {
            continue
        }

        if unit.IsSettlers() {
            continue
        }

        if unit.Name == "Trireme" || unit.Name == "Galley" || unit.Name == "Warship" {
            continue
        }

        isFromPresentPlane := false
        if presentOnArcanus {
            for _, race := range data.ArcanianRaces() {
                if unit.Race == race {
                    isFromPresentPlane = true
                    break
                }
            }
        }
        if !isFromPresentPlane && presentOnMyrror {
            for _, race := range data.MyrranRaces() {
                if unit.Race == race {
                    isFromPresentPlane = true
                    break
                }
            }
        }
        if !isFromPresentPlane {
            continue
        }

        unitCandidates = append(unitCandidates, &unit)
    }
    if len(unitCandidates) == 0 {
        return
    }

    unit := unitCandidates[rand.IntN(len(unitCandidates))]

    // number of units
    count := 1
    countRoll := rand.N(100) + player.GetFame()
    switch {
        case countRoll > 90: count = 3
        case countRoll > 60: count = 2
    }

    // experience
    level := 1
    experience := 20
    experienceRoll := rand.N(100) + player.GetFame()
    switch {
        case experienceRoll > 90:
            level = 3
            experience = 120
        case experienceRoll > 60:
            level = 2
            experience = 60
    }

    // cost
    cost := count * unit.ProductionCost * (level + 3) / 2
    if player.Wizard.RetortEnabled(data.RetortCharismatic) {
        cost /= 2
    }
    if player.Gold < cost {
        return
    }

    // create units
    var overworldUnits []*units.OverworldUnit
    for range count {
        overworldUnit := units.MakeOverworldUnitFromUnit(*unit, fortressCity.X, fortressCity.Y, fortressCity.Plane, player.Wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider())
        overworldUnit.Experience = experience
        overworldUnits = append(overworldUnits, overworldUnit)
    }

    select {
        case model.Events <- &GameEventHireMercenaries{Cost: cost, Units: overworldUnits, Player: player}:
        default:
    }
}

/* random chance to create a merchant event
 */
func (model *GameModel) maybeBuyFromMerchant(player *playerlib.Player) {
    // chance to create an event
    chance := 2 + player.GetFame() / 25
    if player.Wizard.RetortEnabled(data.RetortFamous) {
        chance *= 2
    }
    if chance > 10 {
        chance = 10
    }

    if rand.N(100) >= chance {
        return
    }

    var artifactCandidates []*artifact.Artifact
    for _, candidate := range model.ArtifactPool.AvailableArtifacts() {
        requirementsMet := true
        for _, requirement := range candidate.Requirements {
            if requirement.Amount > 12 {
                requirementsMet = false
                break
            }
        }
        if !requirementsMet {
            continue
        }

        artifactCandidates = append(artifactCandidates, candidate)
    }
    if len(artifactCandidates) == 0 {
        return
    }

    artifact := artifactCandidates[rand.IntN(len(artifactCandidates))]

    // cost
    cost := artifact.Cost
    if player.Wizard.RetortEnabled(data.RetortCharismatic) {
        cost /= 2
    }
    if player.Gold < cost {
        return
    }

    select {
        case model.Events <- &GameEventMerchant{Cost: cost, Artifact: artifact, Player: player}:
        default:
    }
}

func (model *GameModel) DoNextUnit(player *playerlib.Player){
    choose := true
    for choose {
        choose = false
        startingIndex := 0
        if player.SelectedStack != nil {
            for i, stack := range player.Stacks {
                if stack == player.SelectedStack {
                    startingIndex = i + 1
                    break
                }
            }
        }

        player.SelectedStack = nil

        for i := 0; i < len(player.Stacks); i++ {
            index := (i + startingIndex) % len(player.Stacks)
            stack := player.Stacks[index]
            if stack.HasMoves() {
                player.SelectedStack = stack
                stack.EnableMovers()
                stack.DisableNonTransport()

                if player.IsHuman() {
                    select {
                        case model.Events <- &GameEventMoveCamera{Plane: stack.Plane(), X: stack.X(), Y: stack.Y(), Instant: true}:
                        default:
                    }
                }
                /*
                game.Plane = stack.Plane()
                game.Camera.Center(stack.X(), stack.Y())
                */
                break
            }
        }

        if player.SelectedStack != nil && len(player.SelectedStack.CurrentPath) == 0 {
            if model.MaybeBuildRoads(player.SelectedStack, player) {
                choose = true
            }
        }
    }

    if player.IsHuman() {
        if player.SelectedStack != nil {
            if len(player.SelectedStack.CurrentPath) > 0 {
                select {
                    case model.Events<- &GameEventMoveUnit{Player: player}:
                    default:
                }
            }
        }

        model.RefreshUI()
    }
}

func (model *GameModel) StartPlayerTurn(player *playerlib.Player) {
    if player.IsHuman() {
        model.ScrollEvents = nil
    }

    if player.Skip {
        return
    }

    player.MovedStacksThisTurn = 0
    player.SuppressAutoNextTurn = false

    disbandedMessages := model.DisbandUnits(player)

    if player.IsHuman() && len(disbandedMessages) > 0 {
        select {
            case model.Events<- &GameEventScroll{Title: "", Text: strings.Join(disbandedMessages, "\n")}:
            default:
        }
    }

    power := model.ComputePower(player)

    model.DissipateEnchantments(player, power)

    // timestop may have dissipated by now
    timeStop := player.HasEnchantment(data.EnchantmentTimeStop)

    player.Gold += player.GoldPerTurn()
    if player.Gold < 0 {
        player.Gold = 0
    }

    player.Mana += player.ManaPerTurn(power, model)

    if timeStop {
        player.Mana -= data.EnchantmentTimeStop.UpkeepMana()
    }

    player.Mana = max(0, player.Mana)

    if !player.CastingSpell.Invalid() {
        // mana spent on the skill is the minimum of {player's mana, casting skill, remaining cost for spell}
        manaSpent := player.Mana
        if manaSpent > player.RemainingCastingSkill {
            manaSpent = player.RemainingCastingSkill
        }

        spellCost := player.ComputeEffectiveSpellCost(player.CastingSpell, true)

        remainingMana := spellCost - player.CastingSpellProgress
        if remainingMana < manaSpent {
            manaSpent = remainingMana
        }

        player.IncreaseCastingSkillProgress(manaSpent)

        if spellCost <= player.CastingSpellProgress {
            model.doCastSpell(player, player.CastingSpell)
            player.CastingSpell = spellbook.Spell{}
            player.CastingSpellProgress = 0
        }
    }

    if player.ResearchingSpell.Valid() {
        if !timeStop {
            // log.Printf("wizard %v power=%v researching=%v progress=%v/%v perturn=%v", player.Wizard.Name, power, player.ResearchingSpell.Name, player.ResearchProgress, player.ResearchingSpell.ResearchCost, player.SpellResearchPerTurn(power))
            player.ResearchProgress += player.ComputeEffectiveResearchPerTurn(player.SpellResearchPerTurn(power), player.ResearchingSpell)
            if player.ResearchProgress >= player.ResearchingSpell.ResearchCost {

                if player.IsHuman() {
                    select {
                        case model.Events<- &GameEventLearnedSpell{Player: player, Spell: player.ResearchingSpell}:
                        default:
                    }
                }

                log.Printf("wizard %v learned %v", player.Wizard.Name, player.ResearchingSpell.Name)

                player.LearnSpell(player.ResearchingSpell)

                if player.IsHuman() {
                    select {
                        case model.Events<- &GameEventResearchSpell{Player: player}:
                        default:
                    }
                }
            }
        }
    } else if model.TurnNumber > 0 {

        if player.IsHuman() {
            select {
                case model.Events<- &GameEventResearchSpell{Player: player}:
                default:
            }
        }
    }

    player.CastingSkillPower += player.CastingSkillPerTurn(power)

    // reset casting skill for this turn
    player.RemainingCastingSkill = player.ComputeOverworldCastingSkill()

    var removeCities []*citylib.City

    if !timeStop {
        for _, city := range player.Cities {
            cityEvents := city.DoNextTurn(model.GetMap(city.Plane))
            for _, event := range cityEvents {
                switch event.(type) {
                case *citylib.CityEventPopulationGrowth:
                    if player.IsHuman() {
                        growthEvent := event.(*citylib.CityEventPopulationGrowth)

                        verb := "grown"
                        if !growthEvent.Grow {
                            verb = "shrunk"
                        }

                        scrollEvent := GameEventScroll{
                            Title: "CITY GROWTH",
                            Text: fmt.Sprintf("%v has %v to a population of %v.", city.Name, verb, city.Citizens()),
                        }

                        select {
                            case model.Events<- &scrollEvent:
                            default:
                        }
                    }
                case *citylib.CityEventCityAbandoned:
                    removeCities = append(removeCities, city)
                    if player.IsHuman() {
                        select {
                            case model.Events<- &GameEventNotice{Message: fmt.Sprintf("The city of %v has been abandoned.", city.Name)}:
                            default:
                        }
                    }
                case *citylib.CityEventNewBuilding:
                    newBuilding := event.(*citylib.CityEventNewBuilding)

                    if player.IsHuman() {
                        select {
                            case model.Events<- &GameEventNewBuilding{City: city, Building: newBuilding.Building, Player: player}:
                            default:
                        }
                    } else {
                        log.Printf("Year=%v AI %v(%v) city %v created %v", model.TurnNumber, player.Wizard.Name, player.GetBanner(), city.Name, model.BuildingInfo.Name(newBuilding.Building))
                    }
                case *citylib.CityEventOutpostDestroyed:
                    removeCities = append(removeCities, city)
                    if player.IsHuman() {
                        select {
                            case model.Events<- &GameEventNotice{Message: fmt.Sprintf("The outpost of %v has been deserted.", city.Name)}:
                            default:
                        }
                    }
                case *citylib.CityEventOutpostHamlet:
                    if player.IsHuman() {
                        select {
                            case model.Events<- &GameEventNotice{Message: fmt.Sprintf("The outpost of %v has grown into a hamlet.", city.Name)}:
                            default:
                        }
                    }
                case *citylib.CityEventNewUnit:
                    newUnit := event.(*citylib.CityEventNewUnit)
                    overworldUnit := units.MakeOverworldUnitFromUnit(newUnit.Unit, city.X, city.Y, city.Plane, city.GetBanner(), player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider())
                    // only normal units get weapon bonuses
                    if overworldUnit.GetRace() != data.RaceFantastic {
                        overworldUnit.SetWeaponBonus(newUnit.WeaponBonus)
                    }

                    // automatically apply chaos channels to new normal units
                    // checking the race is probably redundant because a new unit built by the city will never be a hero nor fantastic
                    if overworldUnit.GetRace() != data.RaceHero && overworldUnit.GetRace() != data.RaceFantastic && player.HasEnchantment(data.EnchantmentDoomMastery) {
                        model.applyChaosChannels(overworldUnit)
                    }

                    overworldUnit.AddExperience(newUnit.Experience)
                    player.CreateUnit(overworldUnit)
                    model.ResolveStackAt(city.X, city.Y, city.Plane)

                    if player.AIBehavior != nil {
                        player.AIBehavior.ProducedUnit(city, player)
                        log.Printf("Year=%v AI %v(%v) city %v created unit %v", model.TurnNumber, player.Wizard.Name, player.GetBanner(), city.Name, overworldUnit.GetName())
                    }
                }
            }
        }
    }

    model.DoBuildRoads(player)
    model.DoPurify(player)

    for _, stack := range player.Stacks {

        // every unit gains 1 experience at each turn
        unitExperience := model.GetExperienceBonus(stack)
        for _, unit := range stack.Units() {
            switch unit.GetRace() {
                case data.RaceHero: model.AddExperience(player, unit, 1)
                case data.RaceFantastic: // nothing
                default:
                    model.AddExperience(player, unit, unitExperience)
            }
        }

        // base healing rate is 5%. in a town is 10%, with animists guild is 16.67%
        rate := 0.05

        city := player.FindCity(stack.X(), stack.Y(), stack.Plane())

        if city != nil {
            rate = 0.1
            if city.Buildings.Contains(buildinglib.BuildingAnimistsGuild) {
                rate = 0.1667
            }

            if city.HasEnchantment(data.CityEnchantmentStreamOfLife) {
                rate = 1
            }
        }

        // any healer in the same stack provides an additional 20% healing rate
        for _, unit := range stack.Units() {
            if unit.HasAbility(data.AbilityHealer) {
                rate += 0.2
                break
            }
        }

        // heal all the way up
        if player.HasEnchantment(data.EnchantmentHerbMastery) {
            rate = 1
        }

        handleStasis(stack)

        stack.NaturalHeal(rate)
        stack.ResetMoves()
        stack.EnableMovers()
    }

    for _, city := range removeCities {
        player.RemoveCity(city)
    }

    model.maybeHireHero(player)
    model.maybeHireMercenaries(player)
    model.maybeBuyFromMerchant(player)

    player.UpdateFogVisibility()

    if player.GlobalEnchantments.Contains(data.EnchantmentAwareness) {
        model.doExploreFogForAwareness(player)
    }

    model.discoverWizards()

    // game.CenterCamera(player.Cities[0].X, player.Cities[0].Y)
    model.DoNextUnit(player)
    if player.IsHuman() {
        model.RefreshUI()
    }
}

func (model *GameModel) applyChaosChannels(unit *units.OverworldUnit) {
    choices := set.NewSet(
        data.UnitEnchantmentChaosChannelsDemonSkin,
        data.UnitEnchantmentChaosChannelsDemonWings,
        data.UnitEnchantmentChaosChannelsFireBreath,
    )

    if unit.IsFlying() {
        choices.Remove(data.UnitEnchantmentChaosChannelsDemonWings)
    }

    if unit.GetAbilityValue(data.AbilityFireBreath) > 0 {
        choices.Remove(data.UnitEnchantmentChaosChannelsFireBreath)
    }

    slice := choices.Values()
    choice := slice[rand.N(len(slice))]
    unit.AddEnchantment(choice)
}

// 5% chance to destroy a building in every city, and 15% chance to destroy a garrisoned unit
func (model *GameModel) maybeDoNaturesWrath(caster *playerlib.Player) {
    if model.HasRivalEnchantment(caster, data.EnchantmentNaturesWrath) {
        for _, city := range caster.Cities {
            var toRemove []buildinglib.Building
            for _, building := range city.Buildings.Values() {
                if rand.N(100) < 5 {
                    toRemove = append(toRemove, building)
                }
            }

            // FIXME: emit a message for each building that was destroyed
            for _, building := range toRemove {
                log.Printf("Natures wrath: destroy building %v", building)
                city.RemoveBuilding(building)
            }

            stack := caster.FindStack(city.X, city.Y, city.Plane)
            var removeUnits []units.StackUnit
            if stack != nil {
                for _, unit := range stack.Units() {
                    if rand.N(100) < 15 {
                        log.Printf("Natures wrath: destroy unit %v", unit.GetName())
                        removeUnits = append(removeUnits, unit)
                    }
                }
            }

            for _, unit := range removeUnits {
                caster.RemoveUnit(unit)
            }
        }
    }
}

func (model *GameModel) ApplyAIDecisions(player *playerlib.Player, decisions []playerlib.AIDecision) {
    for _, decision := range decisions {
        switch decision.(type) {
            case *playerlib.AIMoveStackDecision:
                moveDecision := decision.(*playerlib.AIMoveStackDecision)
                useStack := moveDecision.Stack

                if len(moveDecision.Units) > 0 && len(moveDecision.Units) != len(moveDecision.Stack.Units()) {
                    useStack = player.SplitStack(moveDecision.Stack, moveDecision.Units)
                }

                useStack.CurrentPath = moveDecision.Path

            // mainly for the raider ai
            case *playerlib.AICreateUnitDecision:
                create := decision.(*playerlib.AICreateUnitDecision)
                log.Printf("ai %v creating %+v", player.Wizard.Name, create)

                overworldUnit := units.MakeOverworldUnitFromUnit(create.Unit, create.X, create.Y, create.Plane, player.Wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider())
                if create.Patrol {
                    overworldUnit.SetBusy(units.BusyStatusPatrol)
                }
                player.CreateUnit(overworldUnit)
                model.ResolveStackAt(create.X, create.Y, create.Plane)
            case *playerlib.AIUpdateCityDecision:
                update := decision.(*playerlib.AIUpdateCityDecision)

                update.City.Farmers = update.Farmers
                update.City.Workers = update.Workers
                update.City.ResetCitizens()

            case *playerlib.AIBuildOutpostDecision:
                build := decision.(*playerlib.AIBuildOutpostDecision)

                stack := build.Stack.GetActiveUnitWithAbility(data.AbilityCreateOutpost)
                if stack != nil {
                    model.CreateOutpost(stack, player)
                }
            case *playerlib.AIBuildRoadDecision:
                build := decision.(*playerlib.AIBuildRoadDecision)
                roadStack := build.Stack
                path := model.FindRoadPath(roadStack.X(), roadStack.Y(), build.X, build.Y, player, roadStack, player.GetFog(roadStack.Plane()))
                if len(path) > 0 {
                    for _, unit := range roadStack.Units() {
                        if unit.HasAbility(data.AbilityConstruction) {
                            unit.SetBuildRoadPath(path)
                        }
                    }
                    model.MaybeBuildRoads(roadStack, player)
                }
            case *playerlib.AIMeldNodeDecision:
                meld := decision.(*playerlib.AIMeldNodeDecision)
                model.tryMeldNode(meld.Stack, player)
            case *playerlib.AIPlaneShiftDecision:
                shift := decision.(*playerlib.AIPlaneShiftDecision)
                if err := model.PlaneShift(shift.Stack, player); err == nil {
                    shift.Stack.CurrentPath = nil
                }
            case *playerlib.AIProduceDecision:
                produce := decision.(*playerlib.AIProduceDecision)
                log.Printf("Year=%v AI %v(%v) city %v producing %v %v", model.TurnNumber, player.Wizard.Name, player.GetBanner(), produce.City.Name, model.BuildingInfo.Name(produce.Building), produce.Unit.Name)
                produce.City.ProducingBuilding = produce.Building
                produce.City.ProducingUnit = produce.Unit
            case *playerlib.AIResearchSpellDecision:
                research := decision.(*playerlib.AIResearchSpellDecision)
                if player.ResearchingSpell.Invalid() {
                    player.ResearchingSpell = research.Spell
                }

            case *playerlib.AICastSpellDecision:
                cast := decision.(*playerlib.AICastSpellDecision)

                if player.CastingSpell.Invalid() {
                    player.CastingSpell = cast.Spell
                }
            case *playerlib.AICastUnitSpellDecision:
                cast := decision.(*playerlib.AICastUnitSpellDecision)
                if player.CastingSpell.Invalid() {
                    player.CastingSpell = cast.Spell
                    player.CastingSpellTarget = cast.Target
                }
        }
    }
}

func (model *GameModel) IsGlobalEnchantmentActive(enchantment data.Enchantment) bool {
    return slices.ContainsFunc(model.Players, func (player *playerlib.Player) bool {
        return player.GlobalEnchantments.Contains(enchantment)
    })
}

func (model *GameModel) PlaneShift(stack *playerlib.UnitStack, player *playerlib.Player) error {
    if model.IsGlobalEnchantmentActive(data.EnchantmentPlanarSeal) {
        return errors.New("Cannot plane shift while the Planar Seal is in effect")
    }

    canMove := false

    cityOppositePlane := player.FindCity(stack.X(), stack.Y(), stack.Plane().Opposite())

    mapPlane := model.GetMap(stack.Plane().Opposite())

    tile := mapPlane.GetTile(stack.X(), stack.Y())
    canMove = cityOppositePlane != nil || stack.AllFlyers() || tile.Tile.IsLand()

    if cityOppositePlane != nil {
        for _, unit := range stack.ActiveUnits() {
            if !cityOppositePlane.CanEnter(unit) {
                canMove = false
                break
            }
        }
    }

    // cannot planar travel if there is an encounter node
    if canMove && mapPlane.GetEncounter(stack.X(), stack.Y()) != nil {
        canMove = false
    }

    if canMove && tile.Tile.IsWater() && stack.AllLandWalkers() {
        canMove = false
    }

    cityStackInfo := model.ComputeCityStackInfo()

    if cityStackInfo.ContainsEnemy(stack.X(), stack.Y(), stack.Plane().Opposite(), player) {
        canMove = false
    }

    // no matter what the reason, just emit a message that planar travel is not possible
    if !canMove {
        return errors.New("The selected units cannot planar travel at this location.")
    }

    if canMove {
        // if there is a friendly stack at the new location then merge the stacks
        mergeStack := cityStackInfo.FindFriendlyStack(stack.X(), stack.Y(), stack.Plane().Opposite(), player)
        stack.SetPlane(stack.Plane().Opposite())
        if mergeStack != nil {
            player.MergeStacks(stack, mergeStack)
        }
        player.UpdateFogVisibility()
        model.discoverWizards()
    }

    return nil
}

func (model *GameModel) discoverWizards() {
    select {
        case model.Events <- &GameEventDiscoverWizards{}:
        default:
    }
}

// tryMeldNode melds the magic node the given stack is standing on, if the stack
// contains a unit able to meld (a magic spirit) and the node is unowned by this
// player and not warped. Returns true if a meld happened.
func (model *GameModel) tryMeldNode(stack *playerlib.UnitStack, player *playerlib.Player) bool {
    useMap := model.GetMap(stack.Plane())
    if useMap == nil {
        return false
    }
    if useMap.GetEncounter(stack.X(), stack.Y()) != nil {
        return false
    }
    node := useMap.GetMagicNode(stack.X(), stack.Y())
    if node == nil || node.Warped || node.MeldingWizard == player {
        return false
    }
    for _, unit := range stack.Units() {
        if unit.HasAbility(data.AbilityMeld) {
            if model.DoMeld(unit, player, node) {
                return true
            }
        }
    }
    return false
}

func (model *GameModel) DoMeld(unit units.StackUnit, player *playerlib.Player, node *maplib.ExtraMagicNode) bool {
    melded := node.Meld(player, unit.GetRawUnit())
    // whether melding succeeded or not the melding unit is removed
    player.RemoveUnit(unit)
    return melded
}

// return true if the stack is going to build a road this turn
func (model *GameModel) MaybeBuildRoads(stack *playerlib.UnitStack, player *playerlib.Player) bool {
    var buildRoadUnits []units.StackUnit
    for _, unit := range stack.Units() {
        if len(unit.GetBuildRoadPath()) > 0 {
            buildRoadUnits = append(buildRoadUnits, unit)
        }
    }

    if len(buildRoadUnits) > 0 {
        if len(buildRoadUnits) < len(stack.Units()) {
            stack = player.SplitStack(stack, buildRoadUnits)
        }

        stack.EnableMovers()
        roadPath := buildRoadUnits[0].GetBuildRoadPath()
        stack.CurrentPath = nil

        useMap := model.GetMap(stack.Plane())

        hasRoad := useMap.ContainsRoad(stack.X(), stack.Y())

        // build a road here
        if !hasRoad {
            for _, unit := range buildRoadUnits {
                unit.SetBusy(units.BusyStatusBuildRoad)
                stack.SetActive(unit, false)
                unit.SetBuildRoadPath(roadPath[1:])
            }

            return true
        } else {

            // i=1
            // [R, X, X]
            // path = [R, X]
            // roadPath = [X, X]

            // move along the road path until we find a point that doesn't have a road
            i := 0
            for i < len(roadPath) {
                point := roadPath[i]
                hasRoad := useMap.ContainsRoad(point.X, point.Y)
                if !hasRoad {
                    break
                }

                i += 1

                // log.Printf("Unit stack %v is building road along path %v. Current %v, %v hasRoad=%v", stack, roadPath, stack.X(), stack.Y(), hasRoad)
            }

            if i >= len(roadPath) {
                i = len(roadPath) - 1
            }

            stack.CurrentPath = roadPath[:i+1]
            for _, unit := range buildRoadUnits {
                unit.SetBuildRoadPath(roadPath[i:])
            }
        }
    }

    return false
}

func (model *GameModel) SuggestCityName(race data.Race) (string) {
    allCities := model.AllCities()
    fallback := fmt.Sprintf("City %d", len(allCities)+1)

    predefinedNames, err := citylib.ReadCityNames(model.LbxCache)
    if err != nil {
        log.Printf("Unable to read predefined city names: %v", err)
        return fallback
    }

    if len(predefinedNames) % 14 != 0 {
        log.Printf("Predefined city names not dividable by number of races")
        return fallback
    }

    raceIndex := 0
    switch race {
        case data.RaceBarbarian: raceIndex = 0
        case data.RaceBeastmen: raceIndex = 1
        case data.RaceDarkElf: raceIndex = 2
        case data.RaceDraconian: raceIndex = 3
        case data.RaceDwarf: raceIndex = 4
        case data.RaceGnoll: raceIndex = 5
        case data.RaceHalfling: raceIndex = 6
        case data.RaceHighElf: raceIndex = 7
        case data.RaceHighMen: raceIndex = 8
        case data.RaceKlackon: raceIndex = 9
        case data.RaceLizard: raceIndex = 10
        case data.RaceNomad: raceIndex = 11
        case data.RaceOrc: raceIndex = 12
        case data.RaceTroll: raceIndex = 13
    }

    var existingNames []string
    for _, city := range allCities {
        existingNames = append(existingNames, city.Name)
    }

    entriesPerRace := len(predefinedNames) / 14
    choices := predefinedNames[raceIndex * entriesPerRace: (raceIndex + 1) * entriesPerRace]

    return chooseCityName(existingNames, choices)
}

func (model *GameModel) CreateOutpost(settlers units.StackUnit, player *playerlib.Player) *citylib.City {
    cityName := model.SuggestCityName(settlers.GetRace())

    newCity := citylib.MakeCity(cityName, settlers.GetX(), settlers.GetY(), settlers.GetRace(), model.BuildingInfo, model.GetMap(settlers.GetPlane()), model, player)
    newCity.Plane = settlers.GetPlane()
    newCity.Population = 300
    newCity.Outpost = true
    newCity.ProducingBuilding = buildinglib.BuildingHousing
    newCity.ProducingUnit = units.UnitNone

    model.GetMap(settlers.GetPlane()).SetRoad(settlers.GetX(), settlers.GetY(), settlers.GetPlane() == data.PlaneMyrror)

    player.RemoveUnit(settlers)
    player.SelectedStack = nil
    model.RefreshUI()
    player.AddCity(newCity)

    stack := player.FindStack(newCity.X, newCity.Y, newCity.Plane)

    select {
        case model.Events<- &GameEventNewOutpost{City: newCity, Stack: stack, Player: player}:
        default:
    }

    return newCity
}

func (model *GameModel) ResolveStackAt(x int, y int, plane data.Plane) {
    stack, player := model.FindStack(x, y, plane)
    if stack == nil {
        return
    }

    count := len(stack.Units())
    if count <= data.MaxUnitsInStack {
        return
    }

    // try to move random units to a nearby tile
    stackUnits := stack.Units()
    for _, i := range rand.Perm(len(stackUnits)) {
        unit := stackUnits[i]

        positions := model.FindEscapePosition(player, unit)

        if len(positions) != 0 {
            // set to a random position
            position := positions[rand.IntN(len(positions))]
            unit.SetX(position.X)
            unit.SetY(position.Y)

            // merge stacks
            stack.RemoveUnit(unit)
            player.AddStack(playerlib.MakeUnitStackFromUnits([]units.StackUnit{unit}))
            allStacks := player.FindAllStacks(position.X, position.Y, stack.Plane())
            for i := 1; i < len(allStacks); i++ {
                player.MergeStacks(allStacks[0], allStacks[i])
            }

            count -= 1
            if count <= data.MaxUnitsInStack {
                break
            }
        }
    }

    // kill units until enough room
    if count > data.MaxUnitsInStack {
        stackUnits = stack.Units()
        slices.SortFunc(stackUnits, func(unitA, unitB units.StackUnit) int {
            // non-heros before heroes
            if unitA.IsHero() != unitB.IsHero() {
                if unitA.IsHero() {
                    return 1
                }
                return -1
            }

            // low-leveled heroes first
            if unitA.IsHero() && unitB.IsHero() {
                return unitA.GetExperience() - unitB.GetExperience()
            }

            // low production or casting cost first
            minCostA := min(unitA.GetRawUnit().ProductionCost, unitA.GetRawUnit().CastingCost)
            minCostB := min(unitB.GetRawUnit().ProductionCost, unitB.GetRawUnit().CastingCost)
            return minCostA - minCostB
        })

        for _, unit := range stackUnits {
            log.Printf("Unit %v killed by ResolveStack", unit)
            player.RemoveUnit(unit)
            count -= 1
            if count <= data.MaxUnitsInStack {
                break
            }
        }
    }
}

/* how much power the player has.
 * add up all melded node tiles, all buildings that produce power, etc
 */
func (model *GameModel) ComputePower(player *playerlib.Player) int {
    if model.ManaShortActive() {
        return 0
    }

    power := float64(0)

    for _, city := range player.Cities {
        power += float64(city.ComputePower())
    }

    magicBonus := float64(1)

    switch model.Settings.Magic {
        case data.MagicSettingWeak: magicBonus = 0.5
        case data.MagicSettingNormal: magicBonus = 1
        case data.MagicSettingPowerful: magicBonus = 1.5
    }

    // the active conjunction type
    magicConjunction := maplib.MagicNodeNone

    if model.ConjunctionChaosActive() {
        magicConjunction = maplib.MagicNodeChaos
    }
    if model.ConjunctionNatureActive() {
        magicConjunction = maplib.MagicNodeNature
    }
    if model.ConjunctionSorceryActive() {
        magicConjunction = maplib.MagicNodeSorcery
    }

    // compute the power a node gives off taking active conjunctions into account
    applyConjunction := func (node *maplib.ExtraMagicNode) float64 {
        nodePower := node.GetPower(magicBonus)

        if nodePower < 0 {
            return nodePower
        }

        multiplier := 1.0

        if magicConjunction != maplib.MagicNodeNone {
            if magicConjunction != node.Kind {
                multiplier *= 0.5
            } else {
                multiplier *= 2
            }
        }

        if player.Wizard.RetortEnabled(data.RetortNodeMastery) {
            multiplier *= 2
        }

        if player.Wizard.RetortEnabled(data.RetortChaosMastery) && node.Kind == maplib.MagicNodeChaos {
            multiplier *= 2
        }

        if player.Wizard.RetortEnabled(data.RetortNatureMastery) && node.Kind == maplib.MagicNodeNature {
            multiplier *= 2
        }

        if player.Wizard.RetortEnabled(data.RetortSorceryMastery) && node.Kind == maplib.MagicNodeSorcery {
            multiplier *= 2
        }

        return nodePower * multiplier
    }

    for _, node := range model.ArcanusMap.GetMeldedNodes(player) {
        power += applyConjunction(node)
    }

    for _, node := range model.MyrrorMap.GetMeldedNodes(player) {
        power += applyConjunction(node)
    }

    power += float64(len(model.ArcanusMap.GetCastedVolcanoes(player)))
    power += float64(len(model.MyrrorMap.GetCastedVolcanoes(player)))

    if power < 0 {
        power = 0
    }

    return int(power)
}

// returns all cities that are connected to this one via roads
func (model *GameModel) FindRoadConnectedCities(city *citylib.City) []*citylib.City {
    // first check if there is at least one tile around the city that is a road

    hasRoad := false

    mapUse := model.GetMap(city.Plane)

    road_check:
    for dx := -1; dx <= 1; dx++ {
        for dy := -1; dy <= 1; dy++ {
            if dx == 0 && dy == 0 {
                continue
            }

            cx := mapUse.WrapX(city.X + dx)
            cy := city.Y + dy

            if dy < 0 || dy >= mapUse.Height() {
                continue
            }

            if mapUse.ContainsRoad(cx, cy) {
                hasRoad = true
                break road_check
            }
        }
    }

    if !hasRoad {
        return nil
    }

    var out []*citylib.City

    for _, otherCity := range model.AllCities() {
        if otherCity == city {
            continue
        }

        if otherCity.Plane == city.Plane && model.IsCityRoadConnected(city, otherCity) {
            out = append(out, otherCity)
        }
    }

    return out
}

// returns true if the two cities are connected by a road
func (model *GameModel) IsCityRoadConnected(fromCity *citylib.City, toCity *citylib.City) bool {
    if fromCity.Plane != toCity.Plane {
        return false
    }

    mapUse := model.GetMap(fromCity.Plane)

    normalized := func (a image.Point) image.Point {
        return image.Pt(mapUse.WrapX(a.X), a.Y)
    }

    // check equality of two points taking wrapping into account
    tileEqual := func (a image.Point, b image.Point) bool {
        return normalized(a) == normalized(b)
    }

    // cost doesn't matter
    tileCost := func (x1 int, y1 int, x2 int, y2 int) float64 {
        return 1
    }

    neighbors := func (x int, y int) []image.Point {
        var out []image.Point
        for dx := -1; dx <= 1; dx++ {
            for dy := -1; dy <= 1; dy++ {
                if dx == 0 && dy == 0 {
                    continue
                }

                cx := mapUse.WrapX(x + dx)
                cy := y + dy

                if cy < 0 || cy >= mapUse.Height() {
                    continue
                }

                if mapUse.ContainsRoad(cx, cy) || model.ContainsCity(cx, cy, fromCity.Plane) {
                    out = append(out, image.Pt(cx, cy))
                }
            }
        }

        return out
    }

    _, ok := pathfinding.FindPath(image.Pt(fromCity.X, fromCity.Y), image.Pt(toCity.X, toCity.Y), 10000, tileCost, neighbors, tileEqual)

    return ok
}

func (model *GameModel) ContainsCity(x int, y int, plane data.Plane) bool {
    city, _ := model.FindCity(x, y, plane)
    return city != nil
}

func (model *GameModel) GetAllGlobalEnchantments() map[data.BannerType]*set.Set[data.Enchantment] {
    enchantments := make(map[data.BannerType]*set.Set[data.Enchantment])
    for _, player := range model.Players {
        enchantments[player.GetBanner()] = player.GlobalEnchantments.Clone()
    }
    return enchantments
}


func (model *GameModel) GetSpellByName(name string) spellbook.Spell {
    return model.allSpells.FindByName(name)
}

func (model *GameModel) RefreshUI() {
    select {
        case model.Events <- &GameEventRefreshUI{}:
        default:
    }
}

// only normal town buildings can be destroyed in an earthequake
func earthquakeBuildingProtected(building buildinglib.Building, intact *set.Set[buildinglib.Building]) bool {
    if !building.IsTownBuilding() {
        return true
    }

    replacement := building.ReplacedBy()
    return replacement != buildinglib.BuildingNone && intact.Contains(replacement)
}

// create a vault event for all items the hero owned, and return vault events for each item
func (model *GameModel) distributeEquipment(player *playerlib.Player, hero *herolib.Hero) []*GameEventVault {
    var events []*GameEventVault

    for index, item := range hero.Equipment {
        if item != nil {
            events = append(events, &GameEventVault{CreatedArtifact: item, Player: player})
        }

        // no matter what the hero loses the item
        hero.Equipment[index] = nil
    }

    return events
}

// returns the number of people, units, buildings that were lost
func (model *GameModel) DoEarthquake(city *citylib.City, player *playerlib.Player) (int, int, []buildinglib.Building) {
    // https://masterofmagic.fandom.com/wiki/Earthquake
    // Spare a building if ReplacedBy() is still intact at the start of the quake
    // (only the top of a line, e.g. Parthenon when Shrine+Temple+Parthenon exist, can collapse).

    // earthquake never kills any citizens
    people := 0

    var killedUnits []units.StackUnit

    stack := player.FindStack(city.X, city.Y, city.Plane)
    if stack != nil {
        for _, unit := range stack.Units() {
            if unit.IsFlying() || unit.HasAbility(data.AbilityNonCorporeal) {
                continue
            }

            roll := rand.N(100)
            if roll < 25 {
                killedUnits = append(killedUnits, unit)
            }
        }
    }

    var showNotice sync.Once

    for _, unit := range killedUnits {
        if unit.IsHero() {
            hero := unit.(*herolib.Hero)
            vaultEvents := model.distributeEquipment(player, hero)
            if player.IsHuman() {
                showNotice.Do(func() {
                    noticeEvent := &GameEventNotice{Message: "One or more heroes died in combat. You must redistribute their equipment."}
                    select {
                        case model.Events <- noticeEvent:
                        default:
                    }
                })
            }

            for _, event := range vaultEvents {
                select {
                    case model.Events <- event:
                    default:
                }
            }
        }

        player.RemoveUnit(unit)
    }

    intact := set.NewSet(city.Buildings.Values()...)
    var destroyedBuildings []buildinglib.Building
    for _, building := range intact.Values() {
        if earthquakeBuildingProtected(building, intact) {
            continue
        }

        roll := rand.N(100)
        if roll < 15 {
            destroyedBuildings = append(destroyedBuildings, building)
            city.Buildings.Remove(building)
        }
    }

    return people, len(killedUnits), destroyedBuildings
}

// returns number of citizens killed, units killed, and buildings destroyed
func (model *GameModel) doCallTheVoid(city *citylib.City, player *playerlib.Player) (int, int, int) {
    // https://masterofmagic.fandom.com/wiki/Call_the_Void

    var destroyedBuildings []buildinglib.Building

    for _, building := range city.Buildings.Values() {
        if rand.N(2) == 0 {
            destroyedBuildings = append(destroyedBuildings, building)
            city.Buildings.Remove(building)
        }
    }

    killedCitizens := 0
    for range city.Citizens() - 1 {
        if rand.N(2) == 0 {
            killedCitizens += 1
        }
    }

    city.Population -= killedCitizens * 1000

    stack := player.FindStack(city.X, city.Y, city.Plane)
    killedUnits := 0
    if stack != nil {
        for _, unit := range stack.Units() {
            // some units are immune
            if unit.HasAbility(data.AbilityMagicImmunity) || unit.HasAbility(data.AbilityRegeneration) || unit.HasEnchantment(data.UnitEnchantmentRighteousness) {
                continue
            }

            if rand.N(2) == 0 {
                unit.AdjustHealth(-10)
                if unit.GetHealth() <= 0 {
                    player.RemoveUnit(unit)
                    killedUnits += 1
                }
            }
        }
    }

    city.ResetCitizens()

    mapUse := model.GetMap(city.Plane)

    // corrupt surrouding tiles
    for dx := -2; dx <= 2; dx++ {
        for dy := -2; dy <= 2; dy++ {
            cx := mapUse.WrapX(city.X + dx)
            cy := city.Y + dy
            if cy < 0 || cy >= mapUse.Height() {
                continue
            }

            if mapUse.GetTile(cx, cy).Tile.IsLand() && rand.N(2) == 0 {
                mapUse.SetCorruption(cx, cy)
            }
        }
    }

    return killedCitizens, killedUnits, len(destroyedBuildings)
}

func (model *GameModel) GetCityEnchantmentsByBanner(banner data.BannerType) []playerlib.CityEnchantment {
    var result []playerlib.CityEnchantment

    for _, player := range model.Players {
        for _, city := range player.Cities {
            for _, enchantment := range city.GetEnchantmentsCastBy(banner) {
                result = append(result, playerlib.CityEnchantment{City: city, Enchantment: enchantment})
            }
        }
    }

    return result
}

// get all alive players that are not the current player
func (model *GameModel) GetEnemies(player *playerlib.Player) []*playerlib.Player {
    var out []*playerlib.Player
    for _, enemy := range model.Players {
        // if enemy != player && len(enemy.Cities) > 0 {
        if enemy != player && !enemy.Defeated {
            out = append(out, enemy)
        }
    }
    return out
}

func (model *GameModel) ComputeMaximumPopulation(x int, y int, plane data.Plane) int {
    // find catchment area of x, y
    // for each square, compute food production
    // maximum pop is food production
    maybeCity, _ := model.FindCity(x, y, plane)
    if maybeCity != nil {
        return maybeCity.MaximumCitySize()
    }

    mapUse := model.GetMap(plane)
    catchment := mapUse.GetCatchmentArea(x, y)

    food := fraction.Zero()

    for _, tile := range catchment {
        food = food.Add(tile.FoodBonus())
        bonus := tile.GetBonus()
        food = food.Add(fraction.FromInt(bonus.FoodBonus()))
    }

    maximum := int(food.ToFloat())
    if maximum > 25 {
        maximum = 25
    }

    return maximum
}

// all the cities on the continent containing the given x, y with the given plane, beloning to the given player
func (model *GameModel) FindCitiesOnContinent(x int, y int, plane data.Plane, player *playerlib.Player) []*citylib.City {
    tiles := model.GetMap(plane).GetContinentTiles(x, y)
    tileSet := set.NewSet[image.Point]()
    for _, tile := range tiles {
        tileSet.Insert(image.Pt(tile.X, tile.Y))
    }

    var out []*citylib.City

    for _, city := range player.Cities {
        if city.Plane == plane && tileSet.Contains(image.Pt(city.X, city.Y)) {
            out = append(out, city)
        }
    }

    return out
}

// all the stacks on the continent containing the given x, y with the given plane, beloning to the given player
func (model *GameModel) FindStacksOnContinent(x int, y int, plane data.Plane, player *playerlib.Player) []*playerlib.UnitStack {
    tiles := model.GetMap(plane).GetContinentTiles(x, y)
    tileSet := set.NewSet[image.Point]()
    for _, tile := range tiles {
        tileSet.Insert(image.Pt(tile.X, tile.Y))
    }

    var out []*playerlib.UnitStack

    for _, stack := range player.Stacks {
        if stack.Plane() == plane && tileSet.Contains(image.Pt(stack.X(), stack.Y())) {
            out = append(out, stack)
        }
    }

    return out
}

type MovementHandler interface {
    ShowMovement(x int, y int, stack *playerlib.UnitStack, center bool)
    DoEncounter(player *playerlib.Player, stack *playerlib.UnitStack, encounter *maplib.ExtraEncounter, map_ *maplib.Map, x int, y int) combat.CombatState
    DoCombat(player *playerlib.Player, stack *playerlib.UnitStack, enemy *playerlib.Player, enemyStack *playerlib.UnitStack, zone combat.ZoneType) combat.CombatState
    DefeatCity(player *playerlib.Player, stack *playerlib.UnitStack, enemy *playerlib.Player, city *citylib.City) (bool, int)
    DiscoverWizards()
}

// FIXME: can this just use doMoveSelectedUnit?
// returns the rest of the path the stack should walk (nil if the path ends)
func (model *GameModel) doAiMoveUnit(handlers MovementHandler, player *playerlib.Player, stack *playerlib.UnitStack) pathfinding.Path {
    if len(stack.Units()) > 0 {
        // stack = player.SplitStack(stack, stack.ActiveUnits())
        for _, unit := range stack.Units() {
            unit.SetBusy(units.BusyStatusNone)
        }
    }

    // FIXME: split the stack into just the active units in case some are busy or out of moves
    // path := move.Path
    path := stack.CurrentPath

    if len(path) == 0 {
        return nil
    }

    to := path[0]
    path = path[1:]

    // log.Printf("  moving stack %v to %v, %v", stack, to.X, to.Y)
    getStack := func(x int, y int) (playerlib.PathStack, bool) {
        found := player.FindStack(x, y, stack.Plane())
        return found, found != nil
    }
    terrainCost, ok := model.ComputeTerrainCost(stack, stack.X(), stack.Y(), to.X, to.Y, model.GetMap(stack.Plane()), getStack)
    if ok {
        oldX := stack.X()
        oldY := stack.Y()

        mapUse := model.GetMap(stack.Plane())

        encounter := mapUse.GetEncounter(mapUse.WrapX(to.X), to.Y)
        if encounter != nil {
            if !player.AIBehavior.ConfirmEncounter(stack, encounter) {
                player.AIBehavior.InvalidMove(stack)
                return nil
            }
        }

        newCity, _ := model.FindCity(to.X, to.Y, stack.Plane())
        if newCity != nil {
            for _, unit := range stack.ActiveUnits() {
                if !newCity.CanEnter(unit) {
                    player.AIBehavior.InvalidMove(stack)
                    return nil
                }
            }
        }

        stack.Move(to.X - stack.X(), to.Y - stack.Y(), terrainCost, model.GetNormalizeCoordinateFunc())

        if model.GetHumanPlayer().IsVisible(oldX, oldY, stack.Plane()) {
            handlers.ShowMovement(oldX, oldY, stack, false)
        }

        path = player.AIBehavior.MovedStack(stack, path)

        player.LiftFogSquare(stack.X(), stack.Y(), stack.GetSightRange(), stack.Plane())
        handlers.DiscoverWizards()

        if encounter != nil {
            // game.doEncounter(yield, player, stack, encounter, mapUse, stack.X(), stack.Y())
            state := handlers.DoEncounter(player, stack, encounter, mapUse, stack.X(), stack.Y())

            // if we fled the encounter, step back off its tile instead of
            // lingering on top of the (still-guarded) lair / node / tower
            if state == combat.CombatStateAttackerFlee {
                stack.SetX(oldX)
                stack.SetY(oldY)
            }

            return nil
        }

        for _, enemy := range model.GetEnemies(player) {
            // FIXME: this should get all stacks at the given location and merge them into a single stack for combat
            enemyStack := enemy.FindStack(stack.X(), stack.Y(), stack.Plane())
            if enemyStack != nil {
                city := enemy.FindCity(stack.X(), stack.Y(), stack.Plane())
                zone := combat.ZoneType{
                    City: city,
                }

                // state := game.doCombat(yield, player, stack, enemy, enemyStack, zone)
                state := handlers.DoCombat(player, stack, enemy, enemyStack, zone)

                if state == combat.CombatStateAttackerFlee {
                    stack.SetX(oldX)
                    stack.SetY(oldY)
                } else if state == combat.CombatStateDefenderFlee {
                    model.doMoveFleeingDefender(enemy, enemyStack)
                }

                return nil
            }

            city := enemy.FindCity(stack.X(), stack.Y(), stack.Plane())
            if city != nil {
                // raze, gold := game.defeatCity(yield, player, stack, enemy, city)
                raze, gold := handlers.DefeatCity(player, stack, enemy, city)

                player.Fame = max(0, player.Fame + city.FameForCaptureOrRaze(!raze))
                enemy.Fame = max(0, enemy.Fame + city.FameForCaptureOrRaze(false))
                player.Gold += gold
                enemy.Gold -= gold

                // FIXME: if the wizard is neutral and decides to raze the town, then the town could become
                // an encounter zone

                return nil
            }
        }
    } else {
        player.AIBehavior.InvalidMove(stack)
        return nil
    }

    return path
}

// try to relocate a fleeing stack, kills units that are unable
func (model *GameModel) doMoveFleeingDefender(player *playerlib.Player, stack *playerlib.UnitStack) {
    stackUnits := stack.Units()

    for _, i := range rand.Perm(len(stackUnits)) {
        unit := stackUnits[i]
        positions := model.FindEscapePosition(player, unit)

        // kill unit if it can not move
        if len(positions) == 0 {
            player.RemoveUnit(unit)
            continue
        }

        // set to a random position
        position := positions[rand.IntN(len(positions))]
        unit.SetX(position.X)
        unit.SetY(position.Y)

        // merge stacks
        stack.RemoveUnit(unit)
        player.AddStack(playerlib.MakeUnitStackFromUnits([]units.StackUnit{unit}))
        allStacks := player.FindAllStacks(position.X, position.Y, unit.GetPlane())
        for i := 1; i < len(allStacks); i++ {
            player.MergeStacks(allStacks[0], allStacks[i])
        }
    }
}

// try to find a nearby position that the given unit can move to
func (model *GameModel) FindEscapePosition(player *playerlib.Player, unit units.StackUnit) []image.Point {
    x := unit.GetX()
    y := unit.GetY()
    plane := unit.GetPlane()
    mapUse := model.GetMap(plane)
    canMoveToWater := unit.IsFlying() || unit.IsSwimmer() || unit.IsSailing()

    var positions []image.Point
    for dx := -1; dx <= 1; dx++ {
        for dy := -1; dy <= 1; dy++ {
            if dx == 0 && dy == 0 {
                continue
            }

            cx := mapUse.WrapX(x + dx)
            cy := y + dy

            if dy < 0 || dy >= mapUse.Height() {
                continue
            }

            // can not contain an enemy stack or city
            occupied := false
            for _, enemy := range model.GetEnemies(player) {

                if enemy.FindStack(cx, cy, plane) != nil {
                    occupied = true
                    break
                }

                if enemy.FindCity(cx, cy, plane) != nil {
                    occupied = true
                    break
                }
            }

            if occupied {
                continue
            }

            // can not contain a friendly full stack
            existing := player.FindStack(cx, cy, plane)
            if existing != nil && len(existing.Units()) >= data.MaxUnitsInStack {
                continue
            }

            // can not countain encounter
            if mapUse.GetEncounter(cx, cy) != nil {
                continue
            }

            if !mapUse.GetTile(cx, cy).Tile.IsWater() || canMoveToWater {
                positions = append(positions, image.Pt(cx, cy))
            }
        }
    }

    return positions
}

// At the beginning of each turn, Awareness clears the fog from all cities for enchantment's owner (newly built included)
func (model *GameModel) doExploreFogForAwareness(awarenessOwner *playerlib.Player) {
    for _, city := range model.AllCities() {
        if city.GetBanner() == awarenessOwner.GetBanner() {
            continue // No need, those cities do already provide vision
        }
        awarenessOwner.ExploreFogSquare(city.X, city.Y, 1, city.Plane)
        _, owner := model.FindCity(city.X, city.Y, city.Plane)
        if owner != nil {
            model.meetWizards(awarenessOwner, owner)
        }
    }
}

func (model *GameModel) GetNormalizeCoordinateFunc() units.NormalizeCoordinateFunc {
    return func (x int, y int) (int, int) {
        return model.CurrentMap().WrapX(x), y
    }
}

func (model *GameModel) GetNeutralPlayer() *playerlib.Player {
    for _, player := range model.Players {
        if player.GetBanner() == data.BannerBrown {
            return player
        }
    }

    return nil
}

// FIXME: make this able to return nil
func (model *GameModel) GetHumanPlayer() *playerlib.Player {
    return model.Players[0]
}

func (model *GameModel) GetBuildingInfos() buildinglib.BuildingInfos {
    return model.BuildingInfo
}

func (model *GameModel) RelocateUnit(player *playerlib.Player, unit units.StackUnit) {
    summonCity := player.FindSummoningCity()
    if summonCity == nil {
        return
    }

    player.UpdateUnitLocation(unit, summonCity.X, summonCity.Y, summonCity.Plane)

    allStacks := player.FindAllStacks(summonCity.X, summonCity.Y, summonCity.Plane)
    for i := 1; i < len(allStacks); i++ {
        player.MergeStacks(allStacks[0], allStacks[i])
    }

    model.ResolveStackAt(summonCity.X, summonCity.Y, summonCity.Plane)

    unit.SetBusy(units.BusyStatusNone)
}

func (model *GameModel) ComputeWizardPower(player *playerlib.Player) playerlib.WizardPower {
    armyStrength := 0

    for _, stack := range player.Stacks {
        for _, unit := range stack.Units() {
            armyStrength += computeUnitPower(unit)
        }
    }

    manaPerTurn := player.ManaPerTurn(model.ComputePower(player), model)

    research := 0
    for _, spell := range player.KnownSpells.Spells {
        research += spell.ResearchCost
    }

    return playerlib.WizardPower{
        Army: armyStrength / 2,
        Magic: max(0, player.Mana / 2 + manaPerTurn),
        SpellResearch: research / 20,
    }
}

func (model *GameModel) revertVolcanos() {
    mapObjects := []*maplib.Map{model.ArcanusMap, model.MyrrorMap}
    for _, mapObject := range mapObjects {
        for location, _ := range mapObject.ExtraMap {
            if mapObject.HasVolcano(location.X, location.Y) {
                if rand.N(100) < 2 {
                    mapObject.RemoveVolcano(location.X, location.Y)
                }
            }
        }
    }
}

// raises 4 to 6 volcanoes on random tiles
func (model *GameModel) doArmageddon() {
    info := model.ComputeCityStackInfo()
    for _, player := range model.Players {
        if player.GlobalEnchantments.Contains(data.EnchantmentArmageddon) {
            // get a list of valid map tiles on both planes
            var points []data.PlanePoint
            catchment := player.GetAllCatchmentArea()
            mapObjects := []*maplib.Map{model.ArcanusMap, model.MyrrorMap}
            for _, mapObject := range mapObjects {
                for x := range mapObject.Map.Columns() {
                    for y := range mapObject.Map.Rows() {
                        point := data.PlanePoint{X: x, Y: y, Plane: mapObject.Plane}
                        tile := terrain.GetTile(mapObject.Map.Terrain[x][y])
                        if tile.IsWater() || tile.IsRiver() || mapObject.HasVolcano(x, y) || mapObject.HasMagicNode(x, y) || catchment.Contains(point) {
                            continue
                        }

                        city := info.FindCity(x, y, mapObject.Plane)
                        if city != nil && (city.HasEnchantment(data.CityEnchantmentConsecration) || city.HasEnchantment(data.CityEnchantmentChaosWard)) {
                            continue
                        }

                        points = append(points, point)
                    }
                }
            }

            // create 4 to 6 volcanoes
            for _, index := range rand.Perm(len(points))[:min(len(points), 4 + rand.IntN(2))] {
                point := points[index]
                model.GetMap(point.Plane).SetVolcano(point.X, point.Y, player)
            }
        }
    }
}

// corrupts 3-6 random tiles
func (model *GameModel) doGreatWasting() {
    info := model.ComputeCityStackInfo()
    for _, player := range model.Players {
        if player.GlobalEnchantments.Contains(data.EnchantmentGreatWasting) {
            // get a list of valid map tiles on both planes
            var points []data.PlanePoint
            catchment := player.GetAllCatchmentArea()
            mapObjects := []*maplib.Map{model.ArcanusMap, model.MyrrorMap}
            for _, mapObject := range mapObjects {
                for x := range mapObject.Map.Columns() {
                    for y := range mapObject.Map.Rows() {
                        point := data.PlanePoint{X: x, Y: y, Plane: mapObject.Plane}
                        tile := terrain.GetTile(mapObject.Map.Terrain[x][y])
                        if tile.IsWater() || tile.IsRiver() || mapObject.HasCorruption(x, y) || catchment.Contains(point) {
                            continue
                        }

                        city := info.FindCity(x, y, mapObject.Plane)
                        if city != nil && (city.HasEnchantment(data.CityEnchantmentConsecration) || city.HasEnchantment(data.CityEnchantmentChaosWard)) {
                            continue
                        }

                        points = append(points, point)
                    }
                }
            }

            // corrupt 3 to 6 tiles
            for _, index := range rand.Perm(len(points))[:min(len(points), 3 + rand.IntN(3))] {
                point := points[index]
                model.GetMap(point.Plane).SetCorruption(point.X, point.Y)
            }
        }
    }
}

func (model *GameModel) doChaosRift() {
    for _, city := range model.AllCities() {
        if city.HasEnchantment(data.CityEnchantmentChaosRift) {

            // do 5 magical attacks of strength 8 to units in the city
            stack, player := model.FindStack(city.X, city.Y, city.Plane)
            if stack != nil && !stack.IsEmpty() {
                stackUnits := stack.Units()
                for range 5 {
                    choice := stackUnits[rand.N(len(stackUnits))]

                    // regeneration units are never hurt by overland spells
                    if choice.HasAbility(data.AbilityRegeneration) {
                        continue
                    }

                    wrapper := &UnitDamageWrapper{StackUnit: choice}

                    combat.ApplyDamage(wrapper, []int{combat.ComputeRoll(8, 30)}, units.DamageRangedMagical, combat.DamageSourceSpell, combat.DamageModifiers{ArmorPiercing: true, Magic: data.ChaosMagic})
                }

                // check for dead units
                for _, unit := range stackUnits {
                    if unit.GetHealth() <= 0 {
                        player.RemoveUnit(unit)
                    }
                }
            }

            // each building has a 5% chance of being destroyed
            var destroyedBuildings []buildinglib.Building
            for _, building := range city.Buildings.Values() {
                if model.BuildingInfo.ProductionCost(building) != 0 {
                    if rand.N(100) < 5 {
                        destroyedBuildings = append(destroyedBuildings, building)
                    }
                }
            }

            for _, building := range destroyedBuildings {
                // emit a notice?
                city.Buildings.Remove(building)
            }
        }
    }
}

func (model *GameModel) doMeteorStorm() {
    immolate := func (unit units.StackUnit) int {
        damage, _ := combat.ApplyAreaDamage(&UnitDamageWrapper{StackUnit: unit}, 4, units.DamageImmolation, 0)
        return damage
    }

    if !model.HasEnchantment(data.EnchantmentMeteorStorm) {
        return
    }

    affectedPlayers := set.NewSet[*playerlib.Player]()
    for _, player := range model.Players {
        if player.Defeated {
            continue
        }

        // add every other player to the affected set.
        // if another player also cast meteor storm, then ultimately the current player will also
        // be added to the set, meaning both players' meteor storms affects the other
        if player.HasEnchantment(data.EnchantmentMeteorStorm) {
            for _, otherPlayer := range model.Players {
                if otherPlayer != player {
                    affectedPlayers.Insert(otherPlayer)
                }
            }
        }
    }

    entityInfo := model.ComputeCityStackInfo()

    for _, player := range model.Players {
        if player.Defeated {
            continue
        }

        // non-garrisoned units take immolation damage
        var toRemove []units.StackUnit
        for unit := range player.Units() {
            if entityInfo.FindCity(unit.GetX(), unit.GetY(), unit.GetPlane()) == nil {
                immolate(unit)
                if unit.GetHealth() <= 0 {
                    toRemove = append(toRemove, unit)
                }
            }
        }

        for _, unit := range toRemove {
            player.RemoveUnit(unit)
        }

        if affectedPlayers.Contains(player) {
            var removeCities []*citylib.City
            for _, city := range player.Cities {
                // chaos ward and consecration protect the city
                if city.HasEnchantment(data.CityEnchantmentChaosWard) || city.HasEnchantment(data.CityEnchantmentConsecration) {
                    continue
                }

                // destroy all outposts
                if city.Outpost {
                    removeCities = append(removeCities, city)
                    continue
                }

                // buildings hae a 1% chance of being destroyed
                var destroyedBuildings []buildinglib.Building
                for _, building := range city.Buildings.Values() {
                    if rand.N(100) == 0 {
                        // FIXME: include nightshade protection
                        destroyedBuildings = append(destroyedBuildings, building)
                    }
                }

                for _, building := range destroyedBuildings {
                    city.Buildings.Remove(building)
                }

            }

            for _, city := range removeCities {
                player.RemoveCity(city)
            }
        }
    }
}

func (model *GameModel) EndOfTurn() {
    // put stuff here that should happen when all players have taken their turn

    for _, player := range model.Players {
        if !player.Defeated {
            player.AddPowerHistory(model.ComputeWizardPower(player))
        }
    }

    model.revertVolcanos()

    // FIXME: the wiki says armageddon will not do anything while time stop is in effect.
    // figure out what other global spells don't have any effect (great wasting, chaos rift, meteor storm)
    if !model.HasEnchantment(data.EnchantmentTimeStop) {
        model.doArmageddon()
    }

    model.doGreatWasting()

    model.doChaosRift()

    model.doMeteorStorm()

    model.TurnNumber += 1

     // gate random-event rolls behind the user toggle; in-flight events continue
     // their normal decay path inside DoRandomEvents and this just skips new rolls.
    if model.GameSettings.RandomEvents {
        model.DoRandomEvents()
    }

    for _, player := range model.Players {
        if player.Defeated || player.Banished {
            continue
        }

        player.UpdateDiplomaticRelations()
    }

    // FIXME
    /*
    if model.WatchMode {
        log.Printf("---- End of Year %v Summary ----", model.TurnNumber - 1)
        for _, player := range model.Players {
            if !player.IsHuman() {
                power := model.ComputePower(player)
                log.Printf("Wizard %v (%v) - Gold: %v (%v), Food: %v, Mana: %v (%v), Cities: %v, Units: %v, Tax: %v, Research: %v (%v)",
                    player.Wizard.Name, player.GetBanner(), player.Gold, player.GoldPerTurn(), player.FoodPerTurn(), player.Mana, player.ManaPerTurn(power, model),
                    len(player.Cities), player.UnitCount(), player.TaxRate,
                    player.ResearchingSpell.Name, player.ResearchProgress)
            }
        }
    }
    */
}
