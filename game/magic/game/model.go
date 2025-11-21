package game

import (
    "image"
    "slices"
    "math"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/ai"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/lib/functional"
    "github.com/kazzmir/master-of-magic/lib/set"
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

    ArtifactPool map[string]*artifact.Artifact

    Settings setup.NewGameSettings

    heroNames map[int]map[herolib.HeroType]string
    allSpells spellbook.Spells

    // water bodies holds all the points that are part of the same body of water.
    // a plane can have mulitple bodies of water if some bodies are landlocked (enclosed entirely by a continent)
    // this is lazily initialized on first use
    WaterBodies map[data.Plane][]*set.Set[image.Point]

    TurnNumber uint64

    // FIXME: maybe put these in the Map object?
    RoadWorkArcanus map[image.Point]float64
    RoadWorkMyrror map[image.Point]float64

    // work done on purifying tiles
    PurifyWorkArcanus map[image.Point]float64
    PurifyWorkMyrror map[image.Point]float64
}

func MakeGameModel(terrainData *terrain.TerrainData, settings setup.NewGameSettings,
                   startingPlane data.Plane, cityProvider maplib.CityProvider,
                   heroNames map[int]map[herolib.HeroType]string, allSpells spellbook.Spells,
                   artifactPool map[string]*artifact.Artifact,
               ) *GameModel {

    planeTowers := maplib.GeneratePlaneTowerPositions(settings.LandSize, 6)

    return &GameModel{
        ArcanusMap: maplib.MakeMap(terrainData, settings.LandSize, settings.Magic, settings.Difficulty, data.PlaneArcanus, cityProvider, planeTowers),
        MyrrorMap: maplib.MakeMap(terrainData, settings.LandSize, settings.Magic, settings.Difficulty, data.PlaneMyrror, cityProvider, planeTowers),
        ArtifactPool: artifactPool,
        Settings: settings,
        heroNames: heroNames,
        allSpells: allSpells,
        Plane: startingPlane,
        RoadWorkArcanus: make(map[image.Point]float64),
        RoadWorkMyrror: make(map[image.Point]float64),

        PurifyWorkArcanus: make(map[image.Point]float64),
        PurifyWorkMyrror: make(map[image.Point]float64),
    }
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

func (model *GameModel) AddPlayer(wizard setup.WizardCustom, human bool) *playerlib.Player {
    useNames := model.heroNames[len(model.Players)]
    if useNames == nil {
        useNames = make(map[herolib.HeroType]string)
    }

    newPlayer := playerlib.MakePlayer(wizard, human, model.CurrentMap().Width(), model.CurrentMap().Height(), useNames, model)

    if !human {
        newPlayer.AIBehavior = ai.MakeEnemyAI()
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

func (model *GameModel) FindPath(oldX int, oldY int, newX int, newY int, player *playerlib.Player, stack playerlib.PathStack, fog data.FogMap) pathfinding.Path {

    useMap := model.GetMap(stack.Plane())

    if newY < 0 || newY >= useMap.Height() {
        return nil
    }

    allFlyers := stack.AllFlyers()

    // this is to avoid doing path finding at all so that we don't spend time trying to compute an impossible path
    // such as a water unit trying to move to land
    if fog.GetFog(useMap.WrapX(newX), newY) != data.FogTypeUnexplored {
        tileTo := useMap.GetTile(newX, newY)
        tileFrom := useMap.GetTile(oldX, oldY)

        // if this is a water unit that cannot walk on land then just return nil immediately since the move is impossible
        if tileTo.Tile.IsLand() && !stack.CanMoveOnLand(true) {
            return nil
        }

        // if this is a water unit and it is moving from water to more water, but the destination water tile
        // is landlocked and the origin tile is not part of the same body of water, then there cannot be a valid
        // path between the two water tiles
        if tileTo.Tile.IsWater() && tileFrom.Tile.IsWater() && !stack.CanMoveOnLand(true) {
            if model.GetWaterBody(useMap, newX, newY) != model.GetWaterBody(useMap, oldX, oldY) {
                return nil
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
        out := make([]image.Point, 0, 8)

        // cardinals first, followed by diagonals
        // left
        out = append(out, image.Pt(x - 1, y))

        // up
        if y > 0 {
            out = append(out, image.Pt(x, y - 1))
        }

        // right
        out = append(out, image.Pt(x + 1, y))

        // down
        if y < useMap.Height() - 1 {
            out = append(out, image.Pt(x, y + 1))
        }

        // up left
        if y > 0 {
            out = append(out, image.Pt(x - 1, y - 1))
        }

        // down left
        if y < useMap.Height() - 1 {
            out = append(out, image.Pt(x - 1, y + 1))
        }

        // up right
        if y > 0 {
            out = append(out, image.Pt(x + 1, y - 1))
        }

        // down right
        if y < useMap.Height() - 1 {
            out = append(out, image.Pt(x + 1, y + 1))
        }

        return out
    }

    path, ok := pathfinding.FindPath(image.Pt(oldX, oldY), image.Pt(newX, newY), 10000, tileCost, neighbors, tileEqual)
    if ok {
        return path[1:]
    }

    return nil
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

    // sailing units cannot move onto land
    if tileTo.Tile.IsLand() {
        if !stack.CanMoveOnLand(true) {
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
    arcanusBuilds := make(map[image.Point]struct{})
    myrrorBuilds := make(map[image.Point]struct{})

    for _, stack := range slices.Clone(player.Stacks) {
        plane := stack.Plane()

        engineerCount := ComputeEngineerCount(stack, true)

        if engineerCount > 0 {
            x, y := stack.X(), stack.Y()
            roads := model.RoadWorkArcanus
            if plane == data.PlaneMyrror {
                roads = model.RoadWorkMyrror
            }

            amount, ok := roads[image.Pt(x, y)]
            if !ok {
                amount = 0
            }

            tileWork := model.ComputeRoadBuildEffort(x, y, plane) // just to get the work map

            amount += math.Pow(tileWork.WorkPerEngineer, float64(engineerCount))
            if amount >= tileWork.TotalWork {
                model.GetMap(plane).SetRoad(x, y, plane == data.PlaneMyrror)

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
    for point, _ := range model.RoadWorkArcanus {
        _, ok := arcanusBuilds[point]
        if !ok {
            toDelete = append(toDelete, point)
        }
    }

    for _, point := range toDelete {
        // log.Printf("remove point %v", point)
        delete(model.RoadWorkArcanus, point)
    }

    toDelete = nil
    for point, _ := range model.RoadWorkMyrror {
        _, ok := myrrorBuilds[point]
        if !ok {
            toDelete = append(toDelete, point)
        }
    }

    for _, point := range toDelete {
        // log.Printf("remove point %v", point)
        delete(model.RoadWorkMyrror, point)
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
            purify := model.PurifyWorkArcanus
            if plane == data.PlaneMyrror {
                purify = model.PurifyWorkMyrror
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
    for point, _ := range model.PurifyWorkArcanus {
        _, ok := arcanusBuilds[point]
        if !ok {
            toDelete = append(toDelete, point)
        }
    }

    for _, point := range toDelete {
        // log.Printf("remove point %v", point)
        delete(model.PurifyWorkArcanus, point)
    }

    toDelete = nil
    for point, _ := range model.PurifyWorkMyrror {
        _, ok := myrrorBuilds[point]
        if !ok {
            toDelete = append(toDelete, point)
        }
    }

    for _, point := range toDelete {
        // log.Printf("remove point %v", point)
        delete(model.PurifyWorkMyrror, point)
    }
}
