package game

import (
    "image/color"
    "image"
    "math/rand/v2"
    "log"
    "math"
    "fmt"
    "context"
    "strings"
    "slices"
    "errors"

    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/ai"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/mastery"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/mirror"
    "github.com/kazzmir/master-of-magic/game/magic/camera"
    "github.com/kazzmir/master-of-magic/game/magic/banish"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/game/magic/unitview"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/cityview"
    "github.com/kazzmir/master-of-magic/game/magic/armyview"
    "github.com/kazzmir/master-of-magic/game/magic/citylistview"
    "github.com/kazzmir/master-of-magic/game/magic/magicview"
    "github.com/kazzmir/master-of-magic/game/magic/diplomacy"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/summon"
    "github.com/kazzmir/master-of-magic/game/magic/cartographer"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/music"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/gamemenu"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    mouselib "github.com/kazzmir/master-of-magic/lib/mouse"
    helplib "github.com/kazzmir/master-of-magic/game/magic/help"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/functional"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/colorm"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

type GameFonts struct {
    InfoFontYellow *font.Font
    InfoFontRed *font.Font
    WhiteFont *font.Font
}

func MakeGameFonts(cache *lbx.LbxCache) *GameFonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Error loading fonts: %v", err)
        return nil
    }

    return &GameFonts{
        InfoFontYellow: loader(fontslib.SmallYellow),
        InfoFontRed: loader(fontslib.SmallRed),
        WhiteFont: loader(fontslib.SmallerWhite),
    }
}

type GameEvent interface {
}

type GameEventMagicView struct {
}

type GameEventDiplomacy struct {
    Player *playerlib.Player
    Enemy *playerlib.Player
}

type GameEventArmyView struct {
}

type GameEventRefreshUI struct {
}

type GameEventSurveyor struct {
}

type GameEventCartographer struct {
}

type GameEventNextTurn struct {
}

type GameEventCityListView struct {
}

type GameEventApprenticeUI struct {
}

type GameEventAstrologer struct {
}

type GameEventHistorian struct {
}

type GameEventCastSpellBook struct {
}

type GameEventNotice struct {
    Message string
}

type GameEventTreasure struct {
    Treasure Treasure
    Player *playerlib.Player
}

type GameEventHireHero struct {
    Hero *herolib.Hero
    Player *playerlib.Player
    Cost int
}

type GameEventHireMercenaries struct {
    Units []*units.OverworldUnit
    Player *playerlib.Player
    Cost int
}

type GameEventMerchant struct {
    Artifact *artifact.Artifact
    Player *playerlib.Player
    Cost int
}

type GameEventVault struct {
    CreatedArtifact *artifact.Artifact
    Player *playerlib.Player
}

// invoke an arbitrary routine
type GameEventInvokeRoutine struct {
    Routine func (yield coroutine.YieldFunc)
}

type GameEventNewOutpost struct {
    City *citylib.City
    Stack *playerlib.UnitStack
    Player *playerlib.Player
}

// add the group to the ui and continue executing the main coroutine until the quit context is cancelled
type GameEventRunUI struct {
    Group *uilib.UIElementGroup
    Quit context.Context
    Song music.Song
}

type GameEventSelectLocationForSpell struct {
    Spell spellbook.Spell
    Player *playerlib.Player
    LocationType LocationType
    SelectedFunc func (yield coroutine.YieldFunc, tileX int, tileY int)
}

type GameEventLearnedSpell struct {
    Player *playerlib.Player
    Spell spellbook.Spell
}

type GameEventResearchSpell struct {
    Player *playerlib.Player
}

type GameEventCastGlobalEnchantment struct {
    Player *playerlib.Player
    Enchantment data.Enchantment
    // if non-nil, then invoke this function after the enchantment animation
    After func()
}

type GameEventGameMenu struct {
}

type GameEventCastSpell struct {
    Player *playerlib.Player
    Spell spellbook.Spell
}

type GameEventSummonUnit struct {
    Player *playerlib.Player
    Unit units.Unit
}

type GameEventSummonArtifact struct {
    Player *playerlib.Player
}

type GameEventSummonHero struct {
    Player *playerlib.Player
    Champion bool
    Female bool
}

type GameEventShowBanish struct {
    Attacker *playerlib.Player
    Defender *playerlib.Player
}

type GameEventNewBuilding struct {
    City *citylib.City
    Building buildinglib.Building
    Player *playerlib.Player
}

type GameEventScroll struct {
    Title string
    Text string
    Old bool // a replayed event, don't add it again
}

type GameEventCityName struct {
    Title string
    City *citylib.City
    // position on screen where to show the input box
    X int
    Y int
}

type GameEventHeroLevelUp struct {
    Hero *herolib.Hero
}

type GameEventMoveCamera struct {
    Plane data.Plane
    X int
    Y int
    Instant bool // set to true to have the camera move instantly, rather than smoothly scroll
}

// https://masterofmagic.fandom.com/wiki/Event
type GameEventShowRandomEvent struct {
    Event *RandomEvent
    // true if the event is just starting, or false if it is ending
    Starting bool
}

type GameEventMoveUnit struct {
    Player *playerlib.Player
}

func StartingCityEvent(city *citylib.City) *GameEventCityName {
    return &GameEventCityName{
        Title: "New Starting City",
        City: city,
        X: 60,
        Y: 28,
    }
}

type ChangeCityEnchantments int
const (
    ChangeCityKeepEnchantments ChangeCityEnchantments = iota
    ChangeCityRemoveOwnerEnchantments
    ChangeCityRemoveAllEnchantments
)

type CastPlayer struct {
    player *playerlib.Player
    remainingCastingSkill int
    castingSkill int
    manaPerTurn int
    mana int
}

func (castPlayer *CastPlayer) ComputeEffectiveSpellCost(spell spellbook.Spell, overland bool) int {
    return castPlayer.player.ComputeEffectiveSpellCost(spell, overland)
}

func (castPlayer *CastPlayer) ComputeEffectiveResearchPerTurn(research float64, spell spellbook.Spell) int {
    return castPlayer.player.ComputeEffectiveResearchPerTurn(research, spell)
}

func (castPlayer *CastPlayer) ComputeTurnsToCast(cost int) int {
    infiniteTurns := 1000
    if castPlayer.castingSkill <= 0 {
        return infiniteTurns
    }

    turns := 0

    usableMana := castPlayer.mana

    castingSkill := castPlayer.remainingCastingSkill 

    for cost > 0 {
        if cost <= usableMana {
            cost -= castingSkill
            usableMana -= castingSkill
        } else {
            // if mana is 0 and manaPerTurn is negative then we basically can never cast the spell
            if usableMana <= 0 && castPlayer.manaPerTurn <= 0 {
                // maybe try to return +infinity?
                return infiniteTurns
            }

            // there is probably a closed-form equation for this
            spend := max(1, min(castingSkill, usableMana))
            cost -= spend
            usableMana -= spend
            if castPlayer.manaPerTurn > 0 {
                usableMana += castPlayer.manaPerTurn
            }
        }

        if cost > 0 {
            turns += 1
        }

        castingSkill = castPlayer.castingSkill
    }

    return turns
}

type GameState int
const (
    GameStateRunning GameState = iota
    GameStateUnitMoving
    GameStateQuit
)

type Game struct {
    Cache *lbx.LbxCache
    ImageCache util.ImageCache

    GameLoader gamemenu.GameLoader

    Music *music.Music

    Fonts *GameFonts

    Settings setup.NewGameSettings

    Counter uint64
    Fog *ebiten.Image
    Drawer func (*ebiten.Image, *Game)
    State GameState
    Plane data.Plane

    TurnNumber uint64

    ArtifactPool map[string]*artifact.Artifact

    MouseData *mouselib.MouseData

    Events chan GameEvent
    BuildingInfo buildinglib.BuildingInfos

    MovingStack *playerlib.UnitStack

    // https://masterofmagic.fandom.com/wiki/Event
    RandomEvents []*RandomEvent
    LastEventTurn uint64

    HudUI *uilib.UI
    Help helplib.Help

    ArcanusMap *maplib.Map
    MyrrorMap *maplib.Map

    // FIXME: maybe put these in the Map object?
    RoadWorkArcanus map[image.Point]float64
    RoadWorkMyrror map[image.Point]float64

    // work done on purifying tiles
    PurifyWorkArcanus map[image.Point]float64
    PurifyWorkMyrror map[image.Point]float64

    Players []*playerlib.Player
    CurrentPlayer int

    // the scroll events that occurred this turn
    ScrollEvents []*GameEventScroll

    Camera camera.Camera
}

func (game *Game) GetFogImage() *ebiten.Image {
    if game.Fog != nil {
        return game.Fog
    }

    game.Fog = ebiten.NewImage(game.CurrentMap().TileWidth(), game.CurrentMap().TileHeight())
    game.Fog.Fill(color.RGBA{R: 8, G: 4, B: 4, A: 0xff})
    return game.Fog
}

func (game *Game) GetMap(plane data.Plane) *maplib.Map {
    switch plane {
        case data.PlaneArcanus: return game.ArcanusMap
        case data.PlaneMyrror: return game.MyrrorMap
    }

    return nil
}

func (game *Game) CurrentMap() *maplib.Map {
    if game.Plane == data.PlaneArcanus {
        return game.ArcanusMap
    }

    return game.MyrrorMap
}

type UnitBuildPowers struct {
    CreateOutpost bool
    Meld bool
    BuildRoad bool
    Purify bool
}

// to be able to use the artifact, the wizard must have enough magic books to satisfy the artifact's requirements
func canUseArtifact(check *artifact.Artifact, wizard setup.WizardCustom) bool {
    // all artifact requirements must be satisfied
    for _, requirement := range check.Requirements {
        if wizard.MagicLevel(requirement.MagicType) < requirement.Amount {
            return false
        }
    }

    // and all ability requirements must be satisfied
    for _, power := range check.Powers {
        if power.Type == artifact.PowerTypeAbility1 || power.Type == artifact.PowerTypeAbility2 || power.Type == artifact.PowerTypeAbility3 {
            if wizard.MagicLevel(power.Magic) < power.Amount {
                return false
            }
        }
    }

    return true
}

func computeUnitBuildPowers(stack *playerlib.UnitStack) UnitBuildPowers {
    var powers UnitBuildPowers

    for _, check := range stack.ActiveUnits() {
        if check.HasAbility(data.AbilityCreateOutpost) {
            powers.CreateOutpost = true
        }

        if check.HasAbility(data.AbilityMeld) {
            powers.Meld = true
        }

        if check.HasAbility(data.AbilityConstruction) {
            powers.BuildRoad = true
        }

        if check.HasAbility(data.AbilityPurify) {
            powers.Purify = true
        }
    }

    return powers
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

func (game *Game) AllSpells() spellbook.Spells {
    spells, err := spellbook.ReadSpellsFromCache(game.Cache)
    if err != nil {
        log.Printf("Could not read spells from cache: %v", err)
        return spellbook.Spells{}
    }

    return spells
}

func (game *Game) AddPlayer(wizard setup.WizardCustom, human bool) *playerlib.Player{
    heroNames := herolib.ReadNamesPerWizard(game.Cache)
    useNames := heroNames[len(game.Players)]
    if useNames == nil {
        useNames = make(map[herolib.HeroType]string)
    }

    newPlayer := playerlib.MakePlayer(wizard, human, game.CurrentMap().Width(), game.CurrentMap().Height(), useNames, game)

    if !human {
        newPlayer.AIBehavior = ai.MakeEnemyAI()
        newPlayer.StrategicCombat = true
    }

    allSpells := game.AllSpells()

    startingSpells := []string{"Magic Spirit", "Spell of Return"}
    if wizard.RetortEnabled(data.RetortArtificer) {
        startingSpells = append(startingSpells, "Enchant Item", "Create Artifact")
    }

    newPlayer.ResearchPoolSpells = wizard.StartingSpells.Copy()

    // not sure its necessary to add the starting spells to the research pool
    for _, spell := range startingSpells {
        newPlayer.ResearchPoolSpells.AddSpell(allSpells.FindByName(spell))
    }

    // every wizard gets all arcane spells by default
    newPlayer.ResearchPoolSpells.AddAllSpells(allSpells.GetSpellsByMagic(data.ArcaneMagic))

    newPlayer.KnownSpells = wizard.StartingSpells.Copy()
    for _, spell := range startingSpells {
        newPlayer.KnownSpells.AddSpell(allSpells.FindByName(spell))
    }
    newPlayer.CastingSkillPower = computeInitialCastingSkillPower(newPlayer.Wizard.Books)

    newPlayer.InitializeResearchableSpells(&allSpells)
    newPlayer.UpdateResearchCandidates()

    // log.Printf("Research spells: %v", newPlayer.ResearchPoolSpells)

    // famous wizards get a head start of 10 fame
    if wizard.RetortEnabled(data.RetortFamous) {
        newPlayer.Fame += 10
    }

    game.Players = append(game.Players, newPlayer)
    return newPlayer
}

func createArtifactPool(lbxCache *lbx.LbxCache) map[string]*artifact.Artifact {
    artifacts, err := artifact.ReadArtifacts(lbxCache)
    if err != nil {
        log.Printf("Error reading artifacts")
        return nil
    }

    pool := make(map[string]*artifact.Artifact)
    for _, artifact := range artifacts {
        pool[artifact.Name] = &artifact
    }

    return pool
}

func MakeGame(lbxCache *lbx.LbxCache, settings setup.NewGameSettings) *Game {

    terrainLbx, err := lbxCache.GetLbxFile("terrain.lbx")
    if err != nil {
        log.Printf("Error: could not load terrain: %v", err)
        return nil
    }

    terrainData, err := terrain.ReadTerrainData(terrainLbx)
    if err != nil {
        log.Printf("Error: could not load terrain: %v", err)
        return nil
    }

    help, err := helplib.ReadHelpFromCache(lbxCache)
    if err != nil {
        return nil
    }

    fonts := MakeGameFonts(lbxCache)

    buildingInfo, err := buildinglib.ReadBuildingInfo(lbxCache)
    if err != nil {
        log.Printf("Unable to read building info: %v", err)
        return nil
    }

    imageCache := util.MakeImageCache(lbxCache)

    mouseData, err := mouselib.MakeMouseData(lbxCache, &imageCache)
    if err != nil {
        log.Printf("Unable to read mouse data: %v", err)
        return nil
    }

    game := &Game{
        Cache: lbxCache,
        Help: help,
        Music: music.MakeMusic(lbxCache),
        MouseData: mouseData,
        Events: make(chan GameEvent, 1000),
        Plane: data.PlaneArcanus,
        State: GameStateRunning,
        Settings: settings,
        ImageCache: imageCache,
        Fonts: fonts,
        ArtifactPool: createArtifactPool(lbxCache),
        BuildingInfo: buildingInfo,
        TurnNumber: 0,
        CurrentPlayer: -1,
        Camera: camera.MakeCamera(),

        RoadWorkArcanus: make(map[image.Point]float64),
        RoadWorkMyrror: make(map[image.Point]float64),

        PurifyWorkArcanus: make(map[image.Point]float64),
        PurifyWorkMyrror: make(map[image.Point]float64),
    }

    planeTowers := maplib.GeneratePlaneTowerPositions(settings.LandSize, 6)

    game.ArcanusMap = maplib.MakeMap(terrainData, settings.LandSize, settings.Magic, settings.Difficulty, data.PlaneArcanus, game, planeTowers)
    game.MyrrorMap = maplib.MakeMap(terrainData, settings.LandSize, settings.Magic, settings.Difficulty, data.PlaneMyrror, game, planeTowers)

    game.HudUI = game.MakeHudUI()
    game.Drawer = func(screen *ebiten.Image, game *Game){
        game.DrawGame(screen)
    }

    game.Music.PushSongs(music.SongBackground1, music.SongBackground2, music.SongBackground3)

    return game
}

func (game *Game) GetDifficulty() data.DifficultySetting {
    return game.Settings.Difficulty
}

func (game *Game) Shutdown() {
    game.Music.Stop()
}

func (game *Game) UpdateImages() {
    game.ImageCache = util.MakeImageCache(game.Cache)
    game.Fog = nil
    game.ArcanusMap.ResetCache()
    game.MyrrorMap.ResetCache()

    mouseData, err := mouselib.MakeMouseData(game.Cache, &game.ImageCache)
    if err != nil {
        log.Printf("Unable to read mouse data: %v", err)
    } else {
        game.MouseData = mouseData
        mouse.Mouse.SetImage(game.MouseData.Normal)
    }
}

func (game *Game) GetPlayerByBanner(banner data.BannerType) *playerlib.Player {
    for _, player := range game.Players {
        if player.GetBanner() == banner {
            return player
        }
    }

    return nil
}

// return the city and its owner
func (game *Game) FindCity(x int, y int, plane data.Plane) (*citylib.City, *playerlib.Player) {
    for _, player := range game.Players {
        city := player.FindCity(x, y, plane)
        if city != nil {
            return city, player
        }
    }

    return nil, nil
}

func (game *Game) FindStack(x int, y int, plane data.Plane) (*playerlib.UnitStack, *playerlib.Player) {
    for _, player := range game.Players {
        stack := player.FindStack(x, y, plane)
        if stack != nil {
            return stack, player
        }
    }

    return nil, nil
}

func (game *Game) ContainsCity(x int, y int, plane data.Plane) bool {
    city, _ := game.FindCity(x, y, plane)
    return city != nil
}

func (game *Game) NearCity(point image.Point, squares int, plane data.Plane) bool {
    for _, city := range game.AllCities() {
        if city.Plane == plane {
            xDiff := game.CurrentMap().XDistance(city.X, point.X)
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

type CityValidArea map[image.Point]bool

func (area CityValidArea) FindLocation() (int, int, bool) {
    choices := make([]image.Point, 0, len(area))
    for point, ok := range area {
        if ok {
            choices = append(choices, point)
        }
    }

    if len(choices) == 0 {
        return 0, 0, false
    }

    point := choices[rand.N(len(choices))]
    return point.X, point.Y, true
}

func (game *Game) MakeCityValidArea(plane data.Plane) CityValidArea {
    out := make(CityValidArea)

    mapUse := game.GetMap(plane)
    continents := mapUse.Map.FindContinents()

    for _, continent := range continents {
        if continent.Size() > 100 {
            for _, point := range continent.Values() {
                tile := terrain.GetTile(mapUse.Map.Terrain[point.X][point.Y])
                if point.Y > 3 && point.Y < mapUse.Map.Rows() - 3 && tile.IsLand() && !tile.IsMagic() && mapUse.GetEncounter(point.X, point.Y) == nil {
                    out[point] = true
                }
            }
        }
    }

    return out
}

func (game *Game) FindValidCityLocation(plane data.Plane) (int, int, bool) {
    mapUse := game.GetMap(plane)
    continents := mapUse.Map.FindContinents()

    for i := 0; i < 10; i++ {
        continentIndex := rand.IntN(len(continents))
        continent := continents[continentIndex]
        if continent.Size() > 100 {
            points := continent.Values()
            index := rand.IntN(len(points))
            x := points[index].X
            y := points[index].Y

            tile := terrain.GetTile(mapUse.Map.Terrain[x][y])
            if y > 3 && y < mapUse.Map.Rows() - 3 && tile.IsLand() && !tile.IsMagic() && mapUse.GetEncounter(x, y) == nil && !game.ContainsCity(x, y, plane) {
                return x, y, true
            }
        }
    }

    return 0, 0, false
}

func (game *Game) FindValidCityLocationOnShore(plane data.Plane) (int, int, bool) {
    mapUse := game.GetMap(plane)
    continents := mapUse.Map.FindContinents()

    for i := 0; i < 10; i++ {
        continentIndex := rand.N(len(continents))
        continent := continents[continentIndex]
        if continent.Size() > 100 {

            var candidates []image.Point
            for _, point := range continent.Values() {
                x := point.X
                y := point.Y
                tile := terrain.GetTile(mapUse.Map.Terrain[x][y])
                if y > 3 && y < mapUse.Map.Rows() - 3 && tile.IsLand() && !tile.IsMagic() && mapUse.GetEncounter(x, y) == nil && !game.ContainsCity(x, y, plane) {

                    found := false
                    for dx := -1; dx <= 1; dx++ {
                        for dy := -1; dy <= 1; dy++ {
                            maybe := terrain.GetTile(mapUse.Map.Terrain[mapUse.WrapX(x+dx)][y+dy])
                            if maybe.TerrainType() == terrain.Shore {
                                found = true
                            }
                        }
                    }

                    if found {
                        candidates = append(candidates, point)
                    }
                }
            }

            if len(candidates) > 0 {
                choice := rand.N(len(candidates))
                return candidates[choice].X, candidates[choice].Y, true
            }
        }
    }

    return 0, 0, false
}

func (game *Game) FindValidCityLocationOnContinent(plane data.Plane, x int, y int) (int, int) {
    mapUse := game.GetMap(plane)
    continents := mapUse.Map.FindContinents()

    for _, continent := range continents {
        if continent.Contains(image.Pt(x, y)) {
            points := continent.Values()
            for _, index := range rand.Perm(len(points)) {
                tile := mapUse.GetTile(points[index].X, points[index].Y)
                if tile.Tile.IsLand() && !tile.Tile.IsMagic() {
                    return points[index].X, points[index].Y
                }
            }
        }
    }

    return 0, 0
}

func randomChoose[T any](choices... T) T {
    return choices[rand.N(len(choices))]
}

// given a list of allNames 'A', 'B', 'C', 'A 1', 'B 1', and a list of choices 'A', 'B', 'C'
// choose the next name that is not in allNames but possibly with some monotonically increasing counter
// In the above example we would choose 'C 1'
func chooseCityName(allNames []string, choices []string) string {

    // try to find a unused city name
    for _, i := range rand.Perm(len(choices)) {
        name := choices[i]
        if !slices.Contains(allNames, name) {
            return name
        }
    }

    // find a name by appending some number to a predefined choice
    suffix := 1
    for {
        for _, i := range rand.Perm(len(choices)) {
            name := fmt.Sprintf("%s %v", choices[i], suffix)
            if !slices.Contains(allNames, name) {
                return name
            }
        }

        suffix += 1
    }
}

func (game *Game) SuggestCityName(race data.Race) (string) {
    allCities := game.AllCities()
    fallback := fmt.Sprintf("City %d", len(allCities)+1)

    predefinedNames, err := citylib.ReadCityNames(game.Cache)
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

func (game *Game) AllUnits() []units.StackUnit{
    var out []units.StackUnit

    for _, player := range game.Players {
        for _, unit := range player.Units {
            out = append(out, unit)
        }
    }

    return out
}

func (game *Game) AllCities() []*citylib.City {
    var out []*citylib.City

    for _, player := range game.Players {
        for _, city := range player.Cities {
            out = append(out, city)
        }
    }

    return out
}

func (game *Game) doCityListView(yield coroutine.YieldFunc) {
    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    cities := game.AllCities()

    var arcanusCities []maplib.MiniMapCity
    var myrrorCities []maplib.MiniMapCity

    for _, city := range cities {
        if city.Plane == data.PlaneArcanus {
            arcanusCities = append(arcanusCities, city)
        } else {
            myrrorCities = append(myrrorCities, city)
        }
    }

    player := game.Players[0]

    drawMinimap := func (screen *ebiten.Image, x int, y int, plane data.Plane, counter uint64){
        use := arcanusCities
        if plane == data.PlaneMyrror {
            use = myrrorCities
        }
        game.GetMap(plane).DrawMinimap(screen, use, x, y, 1, player.GetFog(plane), counter, false)
    }

    var showCity *citylib.City
    selectCity := func(city *citylib.City){
        // ignore outpost
        if city.Citizens() >= 1 {
            showCity = city
        }
        game.Plane = city.Plane
        game.Camera.Center(city.X, city.Y)
    }

    view := citylistview.MakeCityListScreen(game.Cache, game.Players[0], drawMinimap, selectCity)

    game.Drawer = func (screen *ebiten.Image, game *Game){
        view.Draw(screen)
    }

    for view.Update() == citylistview.CityListScreenStateRunning {
        yield()
    }

    // absorb most recent left click
    yield()

    if showCity != nil {
        game.doCityScreen(yield, showCity, game.Players[0], buildinglib.BuildingNone)
    }

    // absorb last click
    yield()

    game.RefreshUI()
}

func (game *Game) doArmyView(yield coroutine.YieldFunc) {
    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    cities := game.AllCities()

    citiesMiniMap := make([]maplib.MiniMapCity, 0, len(cities))
    for _, city := range cities {
        citiesMiniMap = append(citiesMiniMap, city)
    }

    drawMinimap := func (screen *ebiten.Image, x int, y int, fog data.FogMap, plane data.Plane, counter uint64){
        game.GetMap(plane).DrawMinimap(screen, citiesMiniMap, x, y, 1, fog, counter, false)
    }

    showVault := func(){
        game.doVault(yield, nil)
    }

    army := armyview.MakeArmyScreen(game.Cache, game.Players[0], drawMinimap, showVault)

    game.Drawer = func (screen *ebiten.Image, game *Game){
        army.Draw(screen)
    }

    for army.Update() == armyview.ArmyScreenStateRunning {
        yield()
    }

    if army.Player.SelectedStack != nil {
        stack := army.Player.SelectedStack
        select {
            case game.Events <- &GameEventMoveCamera{Plane: stack.Plane(), X: stack.X(), Y: stack.Y()}:
            default:
        }
    }

    game.HudUI = game.MakeHudUI()

    // absorb most recent left click
    yield()
}

/* how much power the player has.
 * add up all melded node tiles, all buildings that produce power, etc
 */
func (game *Game) ComputePower(player *playerlib.Player) int {
    if game.ManaShortActive() {
        return 0
    }

    power := float64(0)

    for _, city := range player.Cities {
        power += float64(city.ComputePower())
    }

    magicBonus := float64(1)

    switch game.Settings.Magic {
        case data.MagicSettingWeak: magicBonus = 0.5
        case data.MagicSettingNormal: magicBonus = 1
        case data.MagicSettingPowerful: magicBonus = 1.5
    }

    // the active conjunction type
    magicConjunction := maplib.MagicNodeNone

    if game.ConjunctionChaosActive() {
        magicConjunction = maplib.MagicNodeChaos
    }
    if game.ConjunctionNatureActive() {
        magicConjunction = maplib.MagicNodeNature
    }
    if game.ConjunctionSorceryActive() {
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

    for _, node := range game.ArcanusMap.GetMeldedNodes(player) {
        power += applyConjunction(node)
    }

    for _, node := range game.MyrrorMap.GetMeldedNodes(player) {
        power += applyConjunction(node)
    }

    power += float64(len(game.ArcanusMap.GetCastedVolcanoes(player)))
    power += float64(len(game.MyrrorMap.GetCastedVolcanoes(player)))

    if power < 0 {
        power = 0
    }

    return int(power)
}

// enemy wizards, but not including the raider ai
func (game *Game) GetEnemyWizards() []*playerlib.Player {
    var out []*playerlib.Player

    for _, player := range game.Players {
        if !player.IsHuman() && player.Wizard.Banner != data.BannerBrown {
            out = append(out, player)
        }
    }

    return out
}

func (game *Game) EnterDiplomacy(player *playerlib.Player, enemy *playerlib.Player) {
    game.Events <- &GameEventDiplomacy{
        Player: player,
        Enemy: enemy,
    }

    game.Events <- &GameEventMagicView{}
}

func (game *Game) doDiplomacy(yield coroutine.YieldFunc, player *playerlib.Player, enemy *playerlib.Player) {
    logic, draw := diplomacy.ShowDiplomacyScreen(game.Cache, player, enemy, 1400 + int(game.TurnNumber / 12))

    oldDrawer := game.Drawer
    game.Drawer = func (screen *ebiten.Image, game *Game){
        draw(screen)
    }

    game.Music.PushSong(diplomacy.GetSong(player, enemy))
    defer game.Music.PopSong()

    logic(yield)

    game.Drawer = oldDrawer
    yield()
    game.RefreshUI()
}

func (game *Game) doMagicView(yield coroutine.YieldFunc) {

    oldDrawer := game.Drawer
    magicScreen := magicview.MakeMagicScreen(game.Cache, game.Players[0], game.GetEnemies(game.Players[0]), game.ComputePower(game.Players[0]), game)

    game.Drawer = func (screen *ebiten.Image, game *Game){
        magicScreen.Draw(screen)
    }

    for magicScreen.Update() == magicview.MagicScreenStateRunning {
        yield()
    }

    yield()

    game.Drawer = oldDrawer

    game.RefreshUI()
}

func validNameString(s string) bool {
    if len(s) != 1 {
        return false
    }

    return strings.ContainsAny(s, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-~@^")
}

type InputFonts struct {
    NameFont *font.Font
    TitleFont *font.Font
}

func MakeInputFonts(cache *lbx.LbxCache) *InputFonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Unable to load fonts: %v", err)
        return nil
    }

    return &InputFonts{
        NameFont: loader(fontslib.NameFont),
        TitleFont: loader(fontslib.TitleFontOrange),
    }
}

func (game *Game) doInput(yield coroutine.YieldFunc, title string, name string, topX int, topY int) string {
    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    fonts := MakeInputFonts(game.Cache)

    maxLength := float64(84)

    quit := false

    source := ebiten.NewImage(1, 1)
    source.Fill(color.RGBA{R: 0xcf, G: 0xef, B: 0xf9, A: 0xff})

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            ui.StandardDraw(screen)
        },
    }
    ui.SetElementsFromArray(nil)

    game.Drawer = func (screen *ebiten.Image, game *Game){
        oldDrawer(screen, game)

        ui.Draw(ui, screen)
    }

    input := &uilib.UIElement{
        TextEntry: func(element *uilib.UIElement, text string) string {
            name = text

            for len(name) > 0 && fonts.NameFont.MeasureTextWidth(name, 1) > maxLength {
                name = name[:len(name)-1]
            }

            return name
        },
        HandleKeys: func(keys []ebiten.Key) {
            for _, key := range keys {
                switch key {
                    case ebiten.KeyEnter:
                        if len(name) > 0 {
                            quit = true
                        }
                    case ebiten.KeyBackspace:
                        if len(name) > 0 {
                            name = name[:len(name) - 1]
                        }
                }
            }
        },
        // Emulating the original game behavior.
        NotLeftClicked: func(element *uilib.UIElement) {
            if len(name) > 0 {
                quit = true
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := game.ImageCache.GetImage("backgrnd.lbx", 33, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(topX), float64(topY))
            scale.DrawScaled(screen, background, &options)

            x, y := options.GeoM.Apply(float64(13), float64(20))

            fonts.NameFont.Print(screen, x, y, scale.ScaleAmount, options.ColorScale, name)

            tx, ty := options.GeoM.Apply(float64(9), float64(6))
            fonts.TitleFont.Print(screen, tx, ty, scale.ScaleAmount, options.ColorScale, title)

            // draw cursor
            cursorX := x + fonts.NameFont.MeasureTextWidth(name, 1)

            util.DrawTextCursor(screen, source, cursorX, y, game.Counter)
        },
    }

    ui.AddElement(input)
    ui.FocusElement(input, name)
    defer ui.UnfocusElement()

    for !quit {
        game.Counter += 1
        ui.StandardUpdate()
        yield()
    }

    return name
}

type NewBuildingFonts struct {
    BigFont *font.Font
}

func MakeNewBuildingFonts(cache *lbx.LbxCache) *NewBuildingFonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Unable to load fonts: %v", err)
        return nil
    }

    return &NewBuildingFonts{
        BigFont: loader(fontslib.YellowBig2),
    }
}

func (game *Game) showNewBuilding(yield coroutine.YieldFunc, city *citylib.City, building buildinglib.Building, player *playerlib.Player){
    drawer := game.Drawer
    defer func(){
        game.Drawer = drawer
    }()

    fonts := MakeNewBuildingFonts(game.Cache)

    background, _ := game.ImageCache.GetImage("resource.lbx", 40, 0)

    animal := game.GetWizardAnimal(player.Wizard)

    wrappedText := fonts.BigFont.CreateWrappedText(float64(175), 1, fmt.Sprintf("The %s of %s has completed the construction of a %s.", city.GetSize(), city.Name, game.BuildingInfo.Name(building)))

    rightSide, _ := game.ImageCache.GetImage("resource.lbx", 41, 0)

    getAlpha := util.MakeFadeIn(7, &game.Counter)

    buildingPics, err := game.ImageCache.GetImagesTransform("cityscap.lbx", building.Index(), "crop", util.AutoCrop)

    if err != nil {
        log.Printf("Error: Unable to get building picture for %v: %v", game.BuildingInfo.Name(building), err)
        return
    }

    buildingPicsAnimation := util.MakeAnimation(buildingPics, true)

    // FIXME: pick background based on tile the land is on?
    landBackground, _ := game.ImageCache.GetImage("cityscap.lbx", 0, 4)

    game.Drawer = func (screen *ebiten.Image, game *Game){
        drawer(screen, game)

        var options ebiten.DrawImageOptions
        options.ColorScale.ScaleAlpha(getAlpha())
        options.GeoM.Translate(float64(8), float64(60))
        scale.DrawScaled(screen, background, &options)
        iconOptions := options
        iconOptions.GeoM.Translate(float64(2), float64(-10))
        scale.DrawScaled(screen, animal, &iconOptions)

        x, y := options.GeoM.Apply(80, 9)
        fonts.BigFont.RenderWrapped(screen, x, y, wrappedText, font.FontOptions{Scale: scale.ScaleAmount, Options: &options})

        options.GeoM.Translate(float64(background.Bounds().Dx()), 0)
        scale.DrawScaled(screen, rightSide, &options)

        x, y = options.GeoM.Apply(float64(4), float64(6))
        buildingSpace := screen.SubImage(scale.ScaleRect(image.Rect(int(x), int(y), int(x) + 45, int(y) + 47))).(*ebiten.Image)

        // buildingSpace.Fill(color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff})
        // vector.DrawFilledRect(buildingSpace, float32(x), float32(y), float32(buildingSpace.Bounds().Dx()), float32(buildingSpace.Bounds().Dy()), color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, false)

        landOptions := options
        landOptions.GeoM.Translate(float64(-10), float64(-10))
        scale.DrawScaled(buildingSpace, landBackground, &landOptions)

        buildingOptions := options
        // translate to the center of the building space, and then draw the image centered by translating
        // by -width/2, -height/2
        buildingOptions.GeoM.Reset()
        buildingOptions.GeoM.Translate(x, y)
        buildingOptions.GeoM.Translate(scale.Unscale(float64(buildingSpace.Bounds().Dx()/2)), scale.Unscale(float64(buildingSpace.Bounds().Dy() - 10)))
        buildingOptions.GeoM.Translate(float64(buildingPicsAnimation.Frame().Bounds().Dx()) / -2, -float64(buildingPicsAnimation.Frame().Bounds().Dy()))
        scale.DrawScaled(buildingSpace, buildingPicsAnimation.Frame(), &buildingOptions)
    }

    quit := false
    quitCounter := 0

    yield()

    for !quit || quitCounter < 7 {
        game.Counter += 1

        if quit {
            quitCounter += 1
        } else {
            if inputmanager.LeftClick() || inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
                quit = true
                getAlpha = util.MakeFadeOut(7, &game.Counter)
            }

            if game.Counter % 8 == 0 {
                buildingPicsAnimation.Next()
            }
        }

        yield()
    }

}

type ScrollFonts struct {
    BigFont *font.Font
    SmallFont *font.Font
}

func MakeScrollFonts(cache *lbx.LbxCache) *ScrollFonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Unable to load fonts: %v", err)
        return nil
    }

    return &ScrollFonts{
        BigFont: loader(fontslib.BigRed2),
        SmallFont: loader(fontslib.SmallRed2),
    }
}

func (game *Game) showScroll(yield coroutine.YieldFunc, title string, text string){
    fonts := MakeScrollFonts(game.Cache)

    wrappedText := fonts.SmallFont.CreateWrappedText(float64(180), 1, text)

    scrollImages, _ := game.ImageCache.GetImages("scroll.lbx", 2)

    totalImages := int((wrappedText.TotalHeight + float64(fonts.BigFont.Height())) / float64(5)) + 1

    if totalImages < 3 {
        totalImages = 3
    }

    if totalImages > len(scrollImages) {
        totalImages = len(scrollImages)
    }

    // only show some of the scroll being unwound
    scrollAnimation := util.MakeAnimation(scrollImages[:totalImages], false)
    pageBackground, _ := game.ImageCache.GetImage("scroll.lbx", 5, 0)

    drawer := game.Drawer
    defer func(){
        game.Drawer = drawer
    }()

    scrollLength := 30

    getAlpha := util.MakeFadeIn(7, &game.Counter)

    game.Drawer = func (screen *ebiten.Image, game *Game){
        drawer(screen, game)

        var options ebiten.DrawImageOptions
        options.ColorScale.ScaleAlpha(getAlpha())

        options.GeoM.Translate(float64(65), float64(25))

        middleY := pageBackground.Bounds().Dy() / 2
        length := scrollLength / 2
        if length > middleY {
            length = middleY
        }
        pagePart := pageBackground.SubImage(image.Rect(0, middleY - length, pageBackground.Bounds().Dx(), middleY + length)).(*ebiten.Image)

        pageOptions := options
        pageOptions.GeoM.Translate(0, float64(middleY - length) + float64(5))
        scale.DrawScaled(screen, pagePart, &pageOptions)

        // make the text fade out a little more than the rest of the scroll
        textScale := options.ColorScale
        textScale.ScaleAlpha(getAlpha())

        x, y := options.GeoM.Apply(float64(pageBackground.Bounds().Dx()) / 2, float64(middleY) - wrappedText.TotalHeight / 2 - float64(fonts.BigFont.Height()) / 2 + 5)
        fonts.BigFont.PrintCenter(screen, x, y, scale.ScaleAmount, textScale, title)
        y += float64(fonts.BigFont.Height()) + 1

        var textOptions ebiten.DrawImageOptions
        textOptions.ColorScale = textScale
        fonts.SmallFont.RenderWrapped(screen, x, y, wrappedText, font.FontOptions{Justify: font.FontJustifyCenter, Scale: scale.ScaleAmount, Options: &textOptions})

        scrollOptions := options
        scrollOptions.GeoM.Translate(float64(-63), float64(-20))
        scale.DrawScaled(screen, scrollAnimation.Frame(), &scrollOptions)
    }

    quit := false

    animationSpeed := uint64(6)

    // absorb clicks and key presses
    yield()

    // show scroll opening up
    for !quit {
        game.Counter += 1

        if inputmanager.LeftClick() || inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
            quit = true
        }

        if game.Counter % animationSpeed == 0 {
            if scrollAnimation.Next() {
                scrollLength += 10
            }
        }

        yield()
    }

    // scroll closes
    scrollAnimation = util.MakeReverseAnimation(scrollImages[:totalImages], false)
    quit = false
    for !quit {
        game.Counter += 1

        if game.Counter % animationSpeed == 0 {
            if scrollAnimation.Next() {
                scrollLength -= 10
            } else {
                quit = true
            }
        }

        yield()
    }

    // offset because the scroll shrunk too much
    scrollLength += 10

    // fade out
    getAlpha = util.MakeFadeOut(7, &game.Counter)
    for range 7 {
        game.Counter += 1
        yield()
    }
}

type OutpostFonts struct {
    BigFont *font.Font
    SmallFont *font.Font
}

func MakeOutpostFonts(cache *lbx.LbxCache) *OutpostFonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Unable to load fonts: %v", err)
        return nil
    }

    return &OutpostFonts{
        BigFont: loader(fontslib.TitleYellowFont),
        SmallFont: loader(fontslib.SmallWhite70),
    }
}

func (game *Game) showOutpost(yield coroutine.YieldFunc, city *citylib.City, stack *playerlib.UnitStack, player *playerlib.Player, rename bool){
    fonts := MakeOutpostFonts(game.Cache)

    group := uilib.MakeGroup()

    game.HudUI.AddGroup(group)
    defer game.HudUI.RemoveGroup(group)

    quit := false

    background, _ := game.ImageCache.GetImage("backgrnd.lbx", 32, 0)
    var uiOptions ebiten.DrawImageOptions

    x1 := 30
    y1 := 50

    uiOptions.GeoM.Translate(float64(x1), float64(y1))
    rect := util.ImageRect(x1, y1, background)
    group.AddElement(&uilib.UIElement{
        LeftClick: func(element *uilib.UIElement){
            quit = true
        },
        Order: 0,
        Layer: 1,
        Rect: rect,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            scale.DrawScaled(screen, background, &uiOptions)

            numHouses := city.GetOutpostHouses()
            maxHouses := 10

            houseOptions := uiOptions
            houseOptions.GeoM.Translate(float64(7), float64(31))

            fullHouseIndex := 34
            emptyHouseIndex := 37

            switch city.Race {
            case data.RaceDarkElf, data.RaceHighElf:
                fullHouseIndex = 35
                emptyHouseIndex = 38
            case data.RaceGnoll, data.RaceKlackon, data.RaceLizard, data.RaceTroll:
                fullHouseIndex = 36
                emptyHouseIndex = 39
            }

            house, _ := game.ImageCache.GetImage("backgrnd.lbx", fullHouseIndex, 0)

            for i := 0; i < numHouses; i++ {
                scale.DrawScaled(screen, house, &houseOptions)
                houseOptions.GeoM.Translate(float64(house.Bounds().Dx()) + 1, 0)
            }

            emptyHouse, _ := game.ImageCache.GetImage("backgrnd.lbx", emptyHouseIndex, 0)
            for i := numHouses; i < maxHouses; i++ {
                scale.DrawScaled(screen, emptyHouse, &houseOptions)
                houseOptions.GeoM.Translate(float64(emptyHouse.Bounds().Dx() + 1), 0)
            }

            x, y := uiOptions.GeoM.Apply(float64(6), float64(22))
            game.Fonts.InfoFontYellow.Print(screen, x, y, scale.ScaleAmount, uiOptions.ColorScale, city.Race.String())

            x, y = uiOptions.GeoM.Apply(float64(20), float64(5))
            if rename {
                fonts.BigFont.Print(screen, x, y, scale.ScaleAmount, uiOptions.ColorScale, "New Outpost Founded")
            } else {
                fonts.BigFont.Print(screen, x, y, scale.ScaleAmount, uiOptions.ColorScale, fmt.Sprintf("Outpost Of %v", city.Name))
            }

            cityScapeOptions := uiOptions
            cityScapeOptions.GeoM.Translate(float64(185), float64(30))
            x, y = cityScapeOptions.GeoM.Apply(0, 0)
            cityScape := screen.SubImage(scale.ScaleRect(image.Rect(int(x), int(y), int(x) + 72, int(y) + 66))).(*ebiten.Image)

            cityScapeBackground, _ := game.ImageCache.GetImage("cityscap.lbx", 0, 0)
            scale.DrawScaled(cityScape, cityScapeBackground, &cityScapeOptions)

            // regular house
            houseIndex := 25

            switch city.Race {
                case data.RaceDarkElf, data.RaceHighElf: houseIndex = 30
                case data.RaceGnoll, data.RaceKlackon, data.RaceLizard, data.RaceTroll: houseIndex = 35
            }

            cityHouse, _ := game.ImageCache.GetImage("cityscap.lbx", houseIndex, 0)
            options2 := cityScapeOptions
            options2.GeoM.Translate(float64(30), float64(20))
            scale.DrawScaled(cityScape, cityHouse, &options2)

            /*
            x, y = options2.GeoM.Apply(0, 0)
            vector.DrawFilledRect(cityScape, float32(x), float32(y), float32(cityScape.Bounds().Dx()), float32(cityScape.Bounds().Dy()), color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, false)
            log.Printf("cityscape at %v, %v", x, y)
            x = 30
            */
            // vector.DrawFilledCircle(cityScape, float32(x), float32(y), 3, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, false)
            // vector.DrawFilledCircle(screen, float32(x), float32(y), 3, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, false)
            // vector.StrokeRect(cityScape, float32(x+1), float32(y+1), float32(cityScape.Bounds().Dx())-1, float32(cityScape.Bounds().Dy())-1, 1, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, false)
            // vector.DrawFilledRect(cityScape, 0, 0, 320, 200, util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x80}), false)

        },
    })

    if stack != nil {

        var matrix colorm.ColorM
        matrix.ChangeHSV(0, 0, 1)

        for i, unit := range stack.Units() {
            pic, _ := unitview.GetUnitOverworldImage(&game.ImageCache, unit)
            stackOptions := uiOptions
            stackOptions.GeoM.Translate(float64(7), float64(55))
            stackOptions.GeoM.Translate(float64(i % 5) * float64(pic.Bounds().Dx()), float64(i / 5) * float64(pic.Bounds().Dy() + 1))

            var patrolOptions colorm.DrawImageOptions
            patrolOptions.GeoM = scale.ScaleGeom(stackOptions.GeoM)

            x, y := stackOptions.GeoM.Apply(0, 0)
            group.AddElement(&uilib.UIElement{
                Order: 1,
                Layer: 1,
                Rect: util.ImageRect(int(x), int(y), pic),
                LeftClick: func(element *uilib.UIElement){
                    player.SelectedStack = stack
                    game.RefreshUI()
                    quit = true
                },
                Tooltip: func(element *uilib.UIElement) (string, *font.Font) {
                    return unit.GetName(), fonts.SmallFont
                },
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    if unit.GetBusy() != units.BusyStatusNone {
                        colorm.DrawImage(screen, pic, matrix, &patrolOptions)
                    } else {
                        scale.DrawScaled(screen, pic, &stackOptions)
                    }

                    // draw the first enchantment on the unit
                    for _, enchantment := range unit.GetEnchantments() {
                        util.DrawOutline(screen, &game.ImageCache, pic, scale.ScaleGeom(stackOptions.GeoM), stackOptions.ColorScale, game.Counter/8, enchantment.Color())
                        break
                    }
                },
            })

        }
    }

    // FIXME: try to replace this with game.doRunUI(), but the issue is that we want the doInput() to run on top of the
    // outpost screen, where doRunUI() would remove the ui group first
    for !quit {
        game.Counter += 1

        game.HudUI.StandardUpdate()
        if yield() != nil {
            break
        }
    }

    if rename {
        city.Name = game.doInput(yield, "New Outpost", city.Name, 80, 100)
    }
}

func (game *Game) showMovement(yield coroutine.YieldFunc, oldX int, oldY int, stack *playerlib.UnitStack, center bool){
    // the number of frames it takes to move a unit one tile
    frames := 10

    dx := float64(game.CurrentMap().XDistance(stack.X(), oldX))
    dy := float64(oldY - stack.Y())

    game.State = GameStateUnitMoving

    game.MovingStack = stack

    for i := 0; i < frames; i++ {
        game.Counter += 1

        interpolate := float64(frames - i) / float64(frames)

        stack.SetOffset(dx * interpolate, dy * interpolate)
        yield()
    }

    game.State = GameStateRunning
    game.MovingStack = nil

    stack.SetOffset(0, 0)
    if center {
        game.Camera.Center(stack.X(), stack.Y())
    }
}

/* return the cost to move from the current position the stack is on to the new given coordinates.
 * also return true/false if the move is even possible
 * FIXME: some values used by this logic could be precomputed and passed in as an argument. Things like 'containsFriendlyCity' could be a map of all cities
 * on the same plane as the unit, thus avoiding the expensive player.FindCity() call
 */
func (game *Game) ComputeTerrainCost(stack playerlib.PathStack, sourceX int, sourceY int, destX int, destY int, mapUse *maplib.Map, getStack func(int, int) (playerlib.PathStack, bool)) (fraction.Fraction, bool) {
    /*
    if stack.OutOfMoves() {
        return fraction.Zero(), false
    }
    */

    tileFrom := mapUse.GetTile(sourceX, sourceY)
    tileTo := mapUse.GetTile(destX, destY)

    if !tileTo.Valid() {
        return fraction.Zero(), false
    }

    if stack.AllFlyers() {
        return fraction.FromInt(1), true
    }

    // can't move from land to ocean unless all units are flyers or swimmers
    if tileFrom.Tile.IsLand() && !tileTo.Tile.IsLand() {
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
        for _, player := range game.Players {
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

/* blink the game screen red to indicate the user attempted to do something invalid
 */
func (game *Game) blinkRed(yield coroutine.YieldFunc) {
    drawer := game.Drawer
    defer func(){
        game.Drawer = drawer
    }()

    fadeSpeed := uint64(6)

    counter := uint64(0)
    getAlpha := util.MakeFadeIn(fadeSpeed, &counter)

    game.Drawer = func (screen *ebiten.Image, game *Game){
        drawer(screen, game)

        var scale colorm.ColorM
        scale.Scale(1, 1, 1, float64(getAlpha() / 2))

        vector.FillRect(screen, 0, 0, float32(screen.Bounds().Dx()), float32(screen.Bounds().Dy()), scale.Apply(color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}), false)
    }

    for i := uint64(0); i < fadeSpeed; i++ {
        counter += 1
        yield()
    }

    getAlpha = util.MakeFadeOut(fadeSpeed, &counter)

    for i := uint64(0); i < fadeSpeed; i++ {
        counter += 1
        yield()
    }
}

func (game *Game) GetNormalizeCoordinateFunc() units.NormalizeCoordinateFunc {
    return func (x int, y int) (int, int) {
        return game.CurrentMap().WrapX(x), y
    }
}

// returns all cities that are connected to this one via roads
func (game *Game) FindRoadConnectedCities(city *citylib.City) []*citylib.City {

    // first check if there is at least one tile around the city that is a road

    hasRoad := false

    mapUse := game.GetMap(city.Plane)

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

    for _, otherCity := range game.AllCities() {
        if otherCity == city {
            continue
        }

        if otherCity.Plane == city.Plane && game.IsCityRoadConnected(city, otherCity) {
            out = append(out, otherCity)
        }
    }

    return out
}

// returns true if the two cities are connected by a road
func (game *Game) IsCityRoadConnected(fromCity *citylib.City, toCity *citylib.City) bool {
    if fromCity.Plane != toCity.Plane {
        return false
    }

    mapUse := game.GetMap(fromCity.Plane)

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

                if mapUse.ContainsRoad(cx, cy) || game.ContainsCity(cx, cy, fromCity.Plane) {
                    out = append(out, image.Pt(cx, cy))
                }
            }
        }

        return out
    }

    _, ok := pathfinding.FindPath(image.Pt(fromCity.X, fromCity.Y), image.Pt(toCity.X, toCity.Y), 10000, tileCost, neighbors, tileEqual)

    return ok
}

func (game *Game) FindPath(oldX int, oldY int, newX int, newY int, player *playerlib.Player, stack playerlib.PathStack, fog data.FogMap) pathfinding.Path {

    useMap := game.GetMap(stack.Plane())

    if newY < 0 || newY >= useMap.Height() {
        return nil
    }

    allFlyers := stack.AllFlyers()

    // FIXME: we might not need this check at all, and just let tileCost handle it
    /*
    if fog.GetFog(useMap.WrapX(newX), newY) != data.FogTypeUnexplored {
        tileTo := useMap.GetTile(newX, newY)
        if tileTo.Tile.IsLand() && !stack.CanMoveOnLand(true) {
            return nil
        }

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

    }
    */

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

    for _, enemy := range game.Players {
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

        cost, ok := game.ComputeTerrainCost(stack, x1, y1, x2, y2, useMap, getStack)
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

/* true if a settler can build a city here
 * a tile must be land, not corrupted, not have an encounter, not have a magic node, and not be too close to another city
 */
func (game *Game) IsSettlableLocation(x int, y int, plane data.Plane) bool {
    if !game.NearCity(image.Pt(x, y), 3, plane) {
        mapUse := game.GetMap(plane)
        if mapUse.HasCorruption(x, y) || mapUse.GetEncounter(x, y) != nil || mapUse.GetMagicNode(x, y) != nil {
            return false
        }

        return mapUse.GetTile(x, y).Tile.IsLand()
    }

    return false
}

func (game *Game) FindSettlableLocations(x int, y int, plane data.Plane, fog data.FogMap) []image.Point {
    tiles := game.GetMap(plane).GetContinentTiles(x, y)

    // compute all pointes that we can't build a city on because they are too close to another city
    unavailable := make(map[image.Point]bool)
    for _, city := range game.AllCities() {
        if city.Plane == plane {
            // keep a distance of 5 tiles from any other city
            for dx := -5; dx <= 5; dx++ {
                for dy := -5; dy <= 5; dy++ {
                    cx := game.CurrentMap().WrapX(city.X + dx)
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

func (game *Game) doSummon(yield coroutine.YieldFunc, summonObject *summon.Summon) {
    drawer := game.Drawer
    defer func(){
        game.Drawer = drawer
    }()

    game.Drawer = func (screen *ebiten.Image, game *Game){
        drawer(screen, game)
        summonObject.Draw(screen)
    }

    for summonObject.Update() == summon.SummonStateRunning {
        leftClick := inputmanager.LeftClick()
        if leftClick {
            break
        }

        yield()
    }

    // absorb left click
    yield()
}

// mutates the ui by adding/removing elements
// FIXME: its a hack to pass in the background image as a double pointer so we can mutate it
func (game *Game) MakeSettingsUI(imageCache *util.ImageCache, ui *uilib.UI, background **ebiten.Image, onOk func()) {
    fonts := fontslib.MakeSettingsFonts(game.Cache)

    var elements []*uilib.UIElement

    var makeElements func()

    makeElements = func() {
        *background, _ = imageCache.GetImage("load.lbx", 11, 0)
        ok, _ := imageCache.GetImage("load.lbx", 4, 0)
        ui.RemoveElements(elements)
        elements = nil

        elements = append(elements, &uilib.UIElement{
            Rect: util.ImageRect(266, 176, ok),
            LeftClick: func(element *uilib.UIElement){
                onOk()
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))
                scale.DrawScaled(screen, ok, &options)
            },
        })

        resolutionBackground, _ := imageCache.GetImage("load.lbx", 5, 0)

        elements = append(elements, &uilib.UIElement{
            Rect: util.ImageRect(20, 40, resolutionBackground),
            LeftClick: func(element *uilib.UIElement){
                selected := func(name string, scale int, algorithm scale.ScaleAlgorithm) string {
                    /*
                    if data.ScreenScale == scale && data.ScreenScaleAlgorithm == algorithm {
                        return name + "*"
                    }
                    */
                    return name
                }

                update := func(scale int, algorithm scale.ScaleAlgorithm){
                    /*
                    data.ScreenScale = scale
                    data.ScreenScaleAlgorithm = algorithm
                    data.ScreenWidth = data.ScreenWidthOriginal * scale
                    data.ScreenHeight = data.ScreenHeightOriginal * scale
                    game.UpdateImages()
                    *imageCache = util.MakeImageCache(game.Cache)
                    makeElements()
                    */
                }

                makeChoices := func (name string, scales []int, algorithm scale.ScaleAlgorithm) []uilib.Selection {
                    var out []uilib.Selection
                    for _, value := range scales {
                        out = append(out, uilib.Selection{
                            Name: selected(fmt.Sprintf("%v %vx", name, value), value, algorithm),
                            Action: func(){
                                update(value, algorithm)
                            },
                        })
                    }
                    return out
                }

                normalChoices := makeChoices("Normal", []int{1, 2, 3, 4}, scale.ScaleAlgorithmNormal)
                scaleChoices := makeChoices("Scale", []int{2, 3, 4}, scale.ScaleAlgorithmScale)
                xbrChoices := makeChoices("XBR", []int{2, 3, 4}, scale.ScaleAlgorithmXbr)

                choices := append(append(normalChoices, scaleChoices...), xbrChoices...)

                ui.AddElements(uilib.MakeSelectionUI(ui, game.Cache, imageCache, 40, 10, "Resolution", choices, true))
            },
            Draw: func (element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))
                scale.DrawScaled(screen, resolutionBackground, &options)

                x, y := options.GeoM.Apply(float64(3), float64(3))
                fonts.OptionFont.Print(screen, x, y, scale.ScaleAmount, options.ColorScale, "Screen")
            },
        })

        ui.AddElements(elements)
    }

    makeElements()
}

func (game *Game) doGameMenu(yield coroutine.YieldFunc) {
    gameMenu, quit := gamemenu.MakeGameMenuUI(game.Cache, game.GameLoader, func(){
        game.State = GameStateQuit
    })

    event := GameEventRunUI{
        Group: gameMenu,
        Quit: quit,
        Song: music.SongNone,
    }

    select {
        case game.Events <- &event:
            game.RefreshUI()
        default:
    }
}

func (game *Game) doVault(yield coroutine.YieldFunc, newArtifact *artifact.Artifact) {
    drawer := game.Drawer
    defer func(){
        game.Drawer = drawer
    }()

    vaultLogic, vaultDrawer := game.showVaultScreen(newArtifact, game.Players[0])

    if newArtifact != nil {
        itemLogic, itemDrawer := game.showItemPopup(newArtifact, game.Cache, &game.ImageCache, nil)

        game.Drawer = func (screen *ebiten.Image, game *Game){
            drawer(screen, game)
            vaultDrawer(screen)
            itemDrawer(screen)
        }

        itemLogic(yield)
    }

    game.Drawer = func (screen *ebiten.Image, game *Game){
        drawer(screen, game)
        vaultDrawer(screen)
    }

    vaultLogic(yield)
}

/* random chance to create a hire hero event
 */
func (game *Game) maybeHireHero(player *playerlib.Player) {
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
                    case game.Events <- &GameEventHireHero{Cost: fee, Hero: hero, Player: player}:
                    default:
                }
            }
        }
    }
}

/* show the hire hero popup, and if the user clicks 'hire' then add the hero to the player's list of heroes
 */
func (game *Game) doHireHero(yield coroutine.YieldFunc, cost int, hero *herolib.Hero, player *playerlib.Player) {
    // ensure the player can actually afford to hire the hero
    if cost > player.Gold {
        return
    }

    quit := false

    result := func(hired bool) {
        if hired {
            if player.AddHeroToFortress(hero) {
                player.Gold -= cost
                hero.SetStatus(herolib.StatusEmployed)

                name := game.doInput(yield, "Hero Name", hero.GetName(), 70, 50)
                hero.SetName(name)

                game.ResolveStackAt(hero.GetX(), hero.GetY(), hero.GetPlane())

                game.RefreshUI()
            }
        } else {
            hero.GainLevel(units.ExperienceChampionHero)
        }
    }

    fadeOut := func() {
        quit = true
    }

    game.HudUI.AddGroup(MakeHireHeroScreenUI(game.Cache, game.HudUI, hero, cost, result, fadeOut))

    for !quit {
        game.Counter += 1
        game.HudUI.StandardUpdate()
        if yield() != nil {
            return
        }
    }

    yield()
}

/* random chance to create a hire mercenaries event
 */
func (game *Game) maybeHireMercenaries(player *playerlib.Player) {
    if game.TurnNumber <= 30 {
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
    for i := 0; i < count; i++ {
        overworldUnit := units.MakeOverworldUnitFromUnit(*unit, fortressCity.X, fortressCity.Y, fortressCity.Plane, player.Wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider())
        overworldUnit.Experience = experience
        overworldUnits = append(overworldUnits, overworldUnit)
    }

    select {
        case game.Events <- &GameEventHireMercenaries{Cost: cost, Units: overworldUnits, Player: player}:
        default:
    }
}

/* show the hire mercenaries popup, and if the user clicks 'hire' then add the units to the player's list of units
 */
func (game *Game) doHireMercenaries(yield coroutine.YieldFunc, cost int, units []*units.OverworldUnit, player *playerlib.Player) {
    if len(units) < 1 {
        return
    }

    if cost > player.Gold {
        return
    }

    quit := false

    result := func(hired bool) {
        quit = true
        if hired {
            for _, unit := range units {
                player.AddUnit(unit)
                game.ResolveStackAt(unit.GetX(), unit.GetY(), unit.GetPlane())
            }
            player.Gold -= cost
            game.RefreshUI()
        }
    }

    game.HudUI.AddGroup(MakeHireMercenariesScreenUI(game.Cache, game.HudUI, units[0], len(units), cost, result))

    for !quit {
        game.Counter += 1
        game.HudUI.StandardUpdate()
        yield()
    }
}

/* random chance to create a merchant event
 */
func (game *Game) maybeBuyFromMerchant(player *playerlib.Player) {
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
    for _, artifact := range game.ArtifactPool {
        requirementsMet := true
        for _, requirement := range artifact.Requirements {
            if requirement.Amount > 12 {
                requirementsMet = false
                break
            }
        }
        if !requirementsMet {
            continue
        }

        artifactCandidates = append(artifactCandidates, artifact)
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
        case game.Events <- &GameEventMerchant{Cost: cost, Artifact: artifact, Player: player}:
        default:
    }
}

/* show the merchant popup, and if the user clicks 'buy' then add the artifact to the player's vault and remove it from the pool
 */
 func (game *Game) doMerchant(yield coroutine.YieldFunc, cost int, artifact *artifact.Artifact, player *playerlib.Player) {
     if cost > player.Gold {
         return
     }

    quit := false

    result := func(bought bool) {
        quit = true
        if bought {
            delete(game.ArtifactPool, artifact.Name)
            game.doVault(yield, artifact)
        }
    }

    game.HudUI.AddElements(MakeMerchantScreenUI(game.Cache, game.HudUI, artifact, cost, result))

    for !quit {
        game.Counter += 1
        game.HudUI.StandardUpdate()
        yield()
    }
}

// FIXME: add a "reason" of fizzling, like "Spell X was fizzled because of Y"
func (game *Game) ShowFizzleSpell(spell spellbook.Spell, caster *playerlib.Player) {
    if caster.IsHuman() {
        game.Events <- &GameEventNotice{
            // FIXME: align this message with how dos mom does it
            Message: fmt.Sprintf("The spell %v has fizzled.", spell.Name),
        }
    }
}

/* show the given message in an error popup on the screen
 */
func (game *Game) doNotice(yield coroutine.YieldFunc, ui *uilib.UI, message string) {
    beep, err := audio.LoadSound(game.Cache, 0)
    if err == nil {
        beep.Play()
    }

    quit := false
    ui.AddElement(uilib.MakeErrorElement(ui, game.Cache, &game.ImageCache, message, func(){
        quit = true
    }))

    yield()

    for !quit {
        game.Counter += 1
        ui.StandardUpdate()
        yield()
    }

    yield()
}

/* player has clicked the 'next turn' button, so attempt to start the next turn
 * but first do some checks on disbanding units to confirm the player really wants to end the turn.
 */
func (game *Game) doNextTurn(yield coroutine.YieldFunc) {
    player := game.Players[0]
    goldIssue, foodIssue, manaIssue := game.CheckDisband(player)

    if goldIssue || foodIssue || manaIssue {

        quit := false
        doit := false

        message := ""

        if goldIssue {
            message = "Some units do not have enough gold and will disband unless you make more gold. Do you wish to allow them to disband?"
        } else if foodIssue {
            message = "Some units do not have enough food and will die unless you allocate more farmers in a city. Do you wish to allow them to die?"
        } else if manaIssue {
            message = "Some units do not have enough mana and will disband unless you make more mana. Do you wish to allow them to disband?"
        }

        group := uilib.MakeGroup()

        yes := func(){
            quit = true
            doit = true

            game.HudUI.RemoveGroup(group)
        }

        no := func(){
            quit = true
            game.HudUI.RemoveGroup(group)
        }

        group.AddElements(uilib.MakeConfirmDialog(group, game.Cache, &game.ImageCache, message, true, yes, no))
        game.HudUI.AddGroup(group)

        for !quit {
            game.Counter += 1
            game.HudUI.StandardUpdate()
            yield()
        }

        if !doit {
            return
        }
        yield()

    }

    game.DoNextTurn()
}

func (game *Game) AddExperience(player *playerlib.Player, unit units.StackUnit, amount int) {
    if player.IsHuman() && unit.IsHero() {
        level_before := unit.GetHeroExperienceLevel()

        unit.AddExperience(amount)

        level_after := unit.GetHeroExperienceLevel()

        if level_before != level_after {
            hero := unit.(*herolib.Hero)
            game.Events <- &GameEventHeroLevelUp{
                Hero: hero,
            }
        }
    } else {
        unit.AddExperience(amount)
    }
}

type RandomEventFonts struct {
    BigFont *font.Font
}

func MakeRandomEventFonts(cache *lbx.LbxCache) *RandomEventFonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Unable to load fonts: %v", err)
        return nil
    }

    return &RandomEventFonts{
        BigFont: loader(fontslib.BigOrangeGradient2),
    }
}

func (game *Game) doRandomEvent(yield coroutine.YieldFunc, event *RandomEvent, start bool, wizard setup.WizardCustom) {
    drawer := game.Drawer
    defer func(){
        game.Drawer = drawer
    }()

    fonts := MakeRandomEventFonts(game.Cache)

    background, _ := game.ImageCache.GetImage("resource.lbx", 40, 0)

    animal := game.GetWizardAnimal(wizard)

    message := event.Message
    if !start {
        message = event.MessageStop
    }
    wrappedText := fonts.BigFont.CreateWrappedText(float64(175), 1, message)

    rightSide, _ := game.ImageCache.GetImage("resource.lbx", 41, 0)

    getAlpha := util.MakeFadeIn(7, &game.Counter)

    eventPic, err := game.ImageCache.GetImage("events.lbx", event.LbxIndex, 0)

    if err != nil {
        log.Printf("Error: Unable to get event picture for %v: %v", event.Type, err)
        return
    }

    if event.Type.IsGood() {
        game.Music.PushSong(music.SongGoodEvent)
    } else {
        game.Music.PushSong(music.SongBadEvent)
    }

    defer game.Music.PopSong()

    game.Drawer = func (screen *ebiten.Image, game *Game){
        drawer(screen, game)

        var options ebiten.DrawImageOptions
        options.ColorScale.ScaleAlpha(getAlpha())
        options.GeoM.Translate(float64(8), float64(60))
        scale.DrawScaled(screen, background, &options)
        iconOptions := options
        iconOptions.GeoM.Translate(float64(34), float64(28))
        iconOptions.GeoM.Translate(float64(-animal.Bounds().Dx() / 2), float64(-animal.Bounds().Dy() / 2))
        scale.DrawScaled(screen, animal, &iconOptions)

        x, y := options.GeoM.Apply(float64(75), float64(9))
        fonts.BigFont.RenderWrapped(screen, x, y, wrappedText, font.FontOptions{Scale: scale.ScaleAmount, Options: &options})

        options.GeoM.Translate(float64(background.Bounds().Dx()), 0)

        shiftX := float64(6)
        shiftY := float64(8)
        options.GeoM.Translate(shiftX, shiftY)
        scale.DrawScaled(screen, eventPic, &options)
        options.GeoM.Translate(-shiftX, -shiftY)
        scale.DrawScaled(screen, rightSide, &options)

        /*
        x, y = options.GeoM.Apply(float64(4 * data.ScreenScale), float64(6 * data.ScreenScale))
        buildingSpace := screen.SubImage(image.Rect(int(x), int(y), int(x) + 45 * data.ScreenScale, int(y) + 47 * data.ScreenScale)).(*ebiten.Image)

        // buildingSpace.Fill(color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff})
        // vector.DrawFilledRect(buildingSpace, float32(x), float32(y), float32(buildingSpace.Bounds().Dx()), float32(buildingSpace.Bounds().Dy()), color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, false)
        */
    }

    quit := false

    for !quit {
        game.Counter += 1
        if inputmanager.LeftClick() {
            quit = true
        }
        if yield() != nil {
            return
        }
    }

    getAlpha = util.MakeFadeOut(7, &game.Counter)

    for range 7 {
        game.Counter += 1
        yield()
    }
}

func (game *Game) ProcessEvents(yield coroutine.YieldFunc) {
    // keep processing events until we don't receive one in the events channel
    var lastEvent GameEvent
    for {
        select {
            case event := <-game.Events:
                switch event.(type) {
                    case *GameEventMagicView:
                        game.doMagicView(yield)
                    case *GameEventDiplomacy:
                        diplomacy := event.(*GameEventDiplomacy)
                        game.doDiplomacy(yield, diplomacy.Player, diplomacy.Enemy)
                    case *GameEventRefreshUI:
                        // compress ui refreshes
                        switch lastEvent.(type) {
                            case *GameEventRefreshUI: // nothing, since we just did a refresh
                            default: game.HudUI = game.MakeHudUI()
                        }
                    case *GameEventHireHero:
                        hire := event.(*GameEventHireHero)
                        if hire.Player.IsHuman() {
                            game.doHireHero(yield, hire.Cost, hire.Hero, hire.Player)
                        }
                    case *GameEventHireMercenaries:
                        hire := event.(*GameEventHireMercenaries)
                        if hire.Player.IsHuman() {
                            game.doHireMercenaries(yield, hire.Cost, hire.Units, hire.Player)
                        }
                    case *GameEventMerchant:
                        merchant := event.(*GameEventMerchant)
                        if merchant.Player.IsHuman() {
                            game.doMerchant(yield, merchant.Cost, merchant.Artifact, merchant.Player)
                        }
                    case *GameEventRunUI:
                        runUI := event.(*GameEventRunUI)
                        if runUI.Song != music.SongNone {
                            game.Music.PushSong(runUI.Song)
                        }

                        game.doRunUI(yield, runUI.Group, runUI.Quit)

                        if runUI.Song != music.SongNone {
                            game.Music.PopSong()
                        }
                    case *GameEventNextTurn:
                        game.doNextTurn(yield)
                    case *GameEventSurveyor:
                        game.doSurveyor(yield)
                    case *GameEventCartographer:
                        game.doCartographer(yield)
                    case *GameEventAstrologer:
                        game.ShowAstrologer(yield)
                    case *GameEventHistorian:
                        game.ShowHistorian(yield)
                    case *GameEventApprenticeUI:
                        game.ShowApprenticeUI(yield, game.Players[0])
                    case *GameEventArmyView:
                        game.doArmyView(yield)
                    case *GameEventShowBanish:
                        banishEvent := event.(*GameEventShowBanish)
                        game.doBanish(yield, banishEvent.Attacker, banishEvent.Defender)
                    case *GameEventNotice:
                        notice := event.(*GameEventNotice)
                        game.doNotice(yield, game.HudUI, notice.Message)
                    case *GameEventCastSpellBook:
                        game.ShowSpellBookCastUI(yield, game.Players[0])
                    case *GameEventCityListView:
                        game.doCityListView(yield)
                    case *GameEventNewOutpost:
                        outpost := event.(*GameEventNewOutpost)
                        if outpost.Player.IsHuman() {
                            game.showOutpost(yield, outpost.City, outpost.Stack, outpost.Player, true)
                        }
                    case *GameEventVault:
                        vaultEvent := event.(*GameEventVault)
                        if vaultEvent.Player.IsHuman() {
                            game.doVault(yield, vaultEvent.CreatedArtifact)
                        } else {
                            // FIXME: give item to AI
                        }
                    case *GameEventShowRandomEvent:
                        randomEvent := event.(*GameEventShowRandomEvent)
                        game.doRandomEvent(yield, randomEvent.Event, randomEvent.Starting, game.Players[0].Wizard)
                    case *GameEventScroll:
                        scroll := event.(*GameEventScroll)
                        game.showScroll(yield, scroll.Title, scroll.Text)
                        if !scroll.Old {
                            game.ScrollEvents = append(game.ScrollEvents, scroll)
                        }
                    case *GameEventLearnedSpell:
                        learnedSpell := event.(*GameEventLearnedSpell)
                        game.doLearnSpell(yield, learnedSpell.Player, learnedSpell.Spell)
                    case *GameEventResearchSpell:
                        researchSpell := event.(*GameEventResearchSpell)
                        game.ResearchNewSpell(yield, researchSpell.Player)
                    case *GameEventInvokeRoutine:
                        invokeRoutine := event.(*GameEventInvokeRoutine)
                        invokeRoutine.Routine(yield)
                    case *GameEventCastGlobalEnchantment:
                        castGlobal := event.(*GameEventCastGlobalEnchantment)
                        player := castGlobal.Player

                        if player.IsHuman() || game.CastingDetectableByHuman(player) {
                            after := castGlobal.After
                            if after == nil {
                                after = func(){}
                            }
                            game.doCastGlobalEnchantment(yield, player, castGlobal.Enchantment, after)
                        }
                    case *GameEventSelectLocationForSpell:
                        selectLocation := event.(*GameEventSelectLocationForSpell)
                        if selectLocation.Player.IsHuman() {
                            tileX, tileY, cancel := game.selectLocationForSpell(yield, selectLocation.Spell, selectLocation.Player, selectLocation.LocationType)
                            if !cancel {
                                selectLocation.SelectedFunc(yield, tileX, tileY)
                            }
                        } else {
                            // FIXME: implement AI location selection
                        }
                    case *GameEventCastSpell:
                        castSpell := event.(*GameEventCastSpell)
                        // in cast.go
                        game.doCastSpell(castSpell.Player, castSpell.Spell)
                    case *GameEventTreasure:
                        treasure := event.(*GameEventTreasure)
                        if treasure.Player.IsHuman() {
                            game.doTreasurePopup(yield, treasure.Player, treasure.Treasure)
                        }

                        game.ApplyTreasure(yield, treasure.Player, treasure.Treasure)
                        yield()

                    case *GameEventNewBuilding:
                        buildingEvent := event.(*GameEventNewBuilding)
                        game.Camera.Center(buildingEvent.City.X, buildingEvent.City.Y)
                        game.Music.PushSong(music.SongBuildingFinished)
                        game.showNewBuilding(yield, buildingEvent.City, buildingEvent.Building, buildingEvent.Player)
                        game.Music.PopSong()
                        game.doCityScreen(yield, buildingEvent.City, buildingEvent.Player, buildingEvent.Building)
                    case *GameEventCityName:
                        cityEvent := event.(*GameEventCityName)
                        city := cityEvent.City
                        city.Name = game.doInput(yield, cityEvent.Title, city.Name, cityEvent.X, cityEvent.Y)
                    case *GameEventSummonUnit:
                        summonUnit := event.(*GameEventSummonUnit)
                        player := summonUnit.Player

                        if player.IsHuman() || game.CastingDetectableByHuman(player) {
                            game.Music.PushSong(music.SongCommonSummoningSpell)
                            game.doSummon(yield, summon.MakeSummonUnit(game.Cache, summonUnit.Unit, player.Wizard.Base, !player.IsHuman()))
                            game.Music.PopSong()
                        }
                    case *GameEventSummonArtifact:
                        summonArtifact := event.(*GameEventSummonArtifact)
                        player := summonArtifact.Player

                        if player.IsHuman() || game.CastingDetectableByHuman(player) {
                            game.Music.PushSong(music.SongVeryRareSummoningSpell)
                            game.doSummon(yield, summon.MakeSummonArtifact(game.Cache, player.Wizard.Base, !player.IsHuman()))
                            game.Music.PopSong()
                        }
                    case *GameEventSummonHero:
                        summonHero := event.(*GameEventSummonHero)
                        player := summonHero.Player

                        if player.IsHuman() || game.CastingDetectableByHuman(player) {
                            game.Music.PushSong(music.SongVeryRareSummoningSpell)
                            game.doSummon(yield, summon.MakeSummonHero(game.Cache, player.Wizard.Base, summonHero.Champion, !player.IsHuman(), summonHero.Female))
                            game.Music.PopSong()
                        }
                    case *GameEventGameMenu:
                        game.doGameMenu(yield)
                    case *GameEventHeroLevelUp:
                        levelEvent := event.(*GameEventHeroLevelUp)
                        game.Music.PushSong(music.SongHeroGainedALevel)
                        game.showHeroLevelUpPopup(yield, levelEvent.Hero)
                        game.Music.PopSong()
                    case *GameEventMoveCamera:
                        moveCamera := event.(*GameEventMoveCamera)
                        game.Plane = moveCamera.Plane

                        if moveCamera.Instant {
                            game.Camera.Center(moveCamera.X, moveCamera.Y)
                        } else {
                            game.doMoveCamera(yield, moveCamera.X, moveCamera.Y)
                        }
                    case *GameEventMoveUnit:
                        moveUnit := event.(*GameEventMoveUnit)
                        game.doMoveSelectedUnit(yield, moveUnit.Player)
                }

                lastEvent = event
            default:
                return
        }
    }
}

func (game *Game) GetTurnNumber() uint64 {
    return game.TurnNumber
}

// the turn as a readable date, such as June 1450
func (game *Game) TurnDate() string {
    base := uint64(1400)

    month := game.TurnNumber % 12
    years := game.TurnNumber / 12

    monthNames := []string{
        "January",
        "February",
        "March",
        "April",
        "May",
        "June",
        "July",
        "August",
        "September",
        "October",
        "November",
        "December",
    }

    return fmt.Sprintf("%s %v", monthNames[month], base + years)
}

func (game *Game) ShowHistorian(yield coroutine.YieldFunc) {
    group := uilib.MakeGroup()

    fade := group.MakeFadeIn(7)

    type Fonts struct {
        Title *font.Font
        Date *font.Font
        Year *font.Font
        BannerFonts map[data.BannerType]*font.Font
    }

    fonts, err := (func() (Fonts, error){
        loader, err := fontslib.Loader(game.Cache)
        if err != nil {
            return Fonts{}, err
        }

        return Fonts{
            Title: loader(fontslib.BigOrangeGradient2),
            Date: loader(fontslib.LightFontSmall),
            Year: loader(fontslib.SmallWhite),
            BannerFonts: fontslib.MakeBannerFonts(game.Cache, 3),
        }, nil
    })()

    if err != nil {
        log.Printf("Error: historian: unable to load font: %v", err)
        return
    }

    quit, cancel := context.WithCancel(context.Background())

    generateImage := func() *ebiten.Image {
        background, _ := game.ImageCache.GetImage("reload.lbx", 0, 0)

        mainImage := ebiten.NewImage(background.Bounds().Dx(), background.Bounds().Dy())
        var options ebiten.DrawImageOptions
        mainImage.DrawImage(background, &options)

        fonts.Title.PrintOptions(mainImage, float64(mainImage.Bounds().Dx() / 2), 10, font.FontOptions{DropShadow: true, Scale: 1, Justify: font.FontJustifyCenter, Options: &options}, "History Of Wizards Power")

        fonts.Date.PrintOptions(mainImage, float64(mainImage.Bounds().Dx() - 8), 11, font.FontOptions{DropShadow: true, Scale: 1, Justify: font.FontJustifyRight, Options: &options}, game.TurnDate())

        xStart := 10
        xEnd := mainImage.Bounds().Dx() - 10

        players := append(game.GetHumanPlayer().GetKnownPlayers(), game.GetHumanPlayer())

        // draw bottom line of graph with tick marks
        tickLineY := mainImage.Bounds().Dy() - 20
        tickLineColor := color.RGBA{R: 0xec, G: 0x8d, B: 0x13, A: 0xff}
        black := color.RGBA{A: 96}
        vector.StrokeLine(mainImage, float32(xStart), float32(tickLineY+1), float32(xEnd), float32(tickLineY+1), 1, black, false)
        vector.StrokeLine(mainImage, float32(xStart), float32(tickLineY), float32(xEnd), float32(tickLineY), 1, tickLineColor, false)
        ticks := 23
        for i := range ticks {
            x := float64(xStart) + (float64(i) + 0.5) * float64(xEnd - xStart) / float64(ticks)
            vector.StrokeLine(mainImage, float32(x), float32(tickLineY - 2), float32(x), float32(tickLineY + 1), 1, tickLineColor, false)
        }

        dates := uint64(10)

        maxYear := game.TurnNumber / 12
        if maxYear < dates {
            maxYear = dates
        }
        for i := range dates {
            x := float64(xStart) + (float64(i) + 0.5) * float64(xEnd - xStart) / float64(dates)
            year := i * maxYear / dates
            // dont show the same year twice
            if i > 0 && (i-1) * maxYear / dates == year {
                break
            }

            fonts.Year.PrintOptions(mainImage, x, float64(tickLineY + 3), font.FontOptions{DropShadow: true, Scale: 1, Justify: font.FontJustifyCenter, Options: &options}, fmt.Sprintf("%v", 1400 + year))
        }

        // get maximum power of any player
        maxPower := 0
        for _, player := range players {
            for _, power := range player.PowerHistory {
                maxPower = max(maxPower, power.TotalPower())
            }
        }

        // have some minimum of the Y axis
        if maxPower < 1000 {
            maxPower = 1000
        }

        for i, player := range players {
            nameFont := fonts.BannerFonts[player.GetBanner()]

            nameFont.PrintOptions(mainImage, float64(10), float64(30 + i * nameFont.Height()), font.FontOptions{DropShadow: true, Scale: 1, Justify: font.FontJustifyLeft, Options: &options}, player.Wizard.Name)

            maxTurn := game.TurnNumber
            if maxTurn < maxYear * 12 {
                maxTurn = maxYear * 12
            }

            lineColor := player.GetBanner().Color()

            yHeight := 120
            baseLine := mainImage.Bounds().Dy() - 25

            first := true
            // lastX := float64(0)
            lastY := float64(0)

            // iterate through the X coordinates on the graph, and get the turn number associated with that coordinate
            // draw a small line between the point at that X coordinate and the previous X coordinate, where the Y value
            // is the power at that turn
            for x := range xEnd - xStart {
                turn := uint64(float64(x) * float64(maxTurn) / float64(xEnd - xStart))
                if turn >= game.TurnNumber {
                    break
                }

                history, ok := player.GetPowerHistoryForTurn(turn)
                if !ok {
                    break
                }
                power := history.TotalPower()
                if power < 0 {
                    power = 0
                }

                y := float64(baseLine - power * yHeight / maxPower)

                if !first {
                    vector.StrokeLine(mainImage, float32(xStart + x - 1), float32(lastY), float32(xStart + x), float32(y), 1, lineColor, false)
                }

                first = false
                lastY = y
            }
        }

        return mainImage
    }

    mainImage := generateImage()
    rect := util.ImageRect(0, 0, mainImage)

    group.AddElement(&uilib.UIElement{
        Layer: 1,
        Rect: rect,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(fade())
            options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))
            scale.DrawScaled(screen, mainImage, &options)
        },
        LeftClick: func(element *uilib.UIElement){
            fade = group.MakeFadeOut(7)
            group.AddDelay(7, func(){
                cancel()
            })
        },
    })

    game.doRunUI(yield, group, quit)
}

func (game *Game) ShowAstrologer(yield coroutine.YieldFunc) {
    group := uilib.MakeGroup()

    quit, cancel := context.WithCancel(context.Background())

    background, _ := game.ImageCache.GetImage("reload.lbx", 1, 0)
    rect := util.ImageRect(10, 10, background)

    fade := group.MakeFadeIn(7)

    type Fonts struct {
        Title *font.Font
        Subtitle *font.Font
        Name *font.Font
    }

    fonts, err := (func() (Fonts, error){
        loader, err := fontslib.Loader(game.Cache)
        if err != nil {
            return Fonts{}, err
        }

        return Fonts{
            Title: loader(fontslib.BigOrangeGradient2),
            Subtitle: loader(fontslib.NormalBlue),
            Name: loader(fontslib.SmallOrange),
        }, nil
    })()

    if err != nil {
        log.Printf("Error: astrologer: unable to load font: %v", err)
        cancel()
        return
    }

    generateImage := func() *ebiten.Image {
        mainImage := ebiten.NewImage(background.Bounds().Dx(), background.Bounds().Dy())
        var options ebiten.DrawImageOptions
        mainImage.DrawImage(background, &options)

        x, y := options.GeoM.Apply(float64(background.Bounds().Dx()) / 2, 9)
        fonts.Title.PrintOptions(mainImage, x, y, font.FontOptions{DropShadow: true, Scale: 1, Justify: font.FontJustifyCenter, Options: &options}, "Current Status Of Wizards")

        fonts.Subtitle.PrintOptions(mainImage, x, y + 14, font.FontOptions{DropShadow: true, Scale: 1, Justify: font.FontJustifyCenter, Options: &options}, "Army Strength")
        fonts.Subtitle.PrintOptions(mainImage, x, y + 63, font.FontOptions{DropShadow: true, Scale: 1, Justify: font.FontJustifyCenter, Options: &options}, "Magic Power")
        fonts.Subtitle.PrintOptions(mainImage, x, y + 113, font.FontOptions{DropShadow: true, Scale: 1, Justify: font.FontJustifyCenter, Options: &options}, "Spell Research")

        armyStart := y + 14 + 12
        magicStart := y + 63 + 12
        researchStart := y + 113 + 12

        // max values used to normalize the bars
        // initial values so that the bars don't immediately take up the full volume
        maxArmy := 400
        maxMagic := 400
        maxResearch := 400

        players := append(game.GetHumanPlayer().GetKnownPlayers(), game.GetHumanPlayer())

        for _, player := range players {
            if player.Defeated || player.IsNeutral() {
                continue
            }

            power := player.LatestWizardPower()
            maxArmy = max(maxArmy, power.Army)
            maxMagic = max(maxMagic, power.Magic)
            maxResearch = max(maxResearch, power.SpellResearch)
        }

        computeBarSize := func(value int, maxValue int) float64 {
            maxLength := 180

            return float64(value) / float64(maxValue) * float64(maxLength)
        }

        xStart := float64(14)
        barStart := xStart + 50
        black := color.RGBA{A: 255}

        drawBar := func (y float64, size float64, lineColor color.RGBA) {
            vector.FillRect(mainImage, float32(barStart + 1), float32(y + 1), float32(size), 2, black, false)
            vector.FillRect(mainImage, float32(barStart), float32(y), float32(size), 2, lineColor, false)
        }

        for i, player := range players {
            if player.Defeated || player.IsNeutral() {
                continue
            }

            lineColor := player.GetBanner().Color()

            power := player.LatestWizardPower()

            // log.Printf("Power for %v: %v", player.Wizard.Name, power)

            x, y := options.GeoM.Apply(xStart, armyStart + float64(i * fonts.Name.Height()))
            fonts.Name.PrintOptions(mainImage, x, y, font.FontOptions{DropShadow: true, Scale: 1, Justify: font.FontJustifyLeft, Options: &options}, player.Wizard.Name)

            armyBarLength := computeBarSize(power.Army, maxArmy)

            drawBar(y + 2, armyBarLength, lineColor)

            x, y = options.GeoM.Apply(xStart, magicStart + float64(i * fonts.Name.Height()))
            fonts.Name.PrintOptions(mainImage, x, y, font.FontOptions{DropShadow: true, Scale: 1, Justify: font.FontJustifyLeft, Options: &options}, player.Wizard.Name)

            magicBarLength := computeBarSize(power.Magic, maxMagic)

            drawBar(y + 2, magicBarLength, lineColor)

            x, y = options.GeoM.Apply(xStart, researchStart + float64(i * fonts.Name.Height()))
            fonts.Name.PrintOptions(mainImage, x, y, font.FontOptions{DropShadow: true, Scale: 1, Justify: font.FontJustifyLeft, Options: &options}, player.Wizard.Name)

            researchBarLength := computeBarSize(power.SpellResearch, maxResearch)
            drawBar(y + 2, researchBarLength, lineColor)
        }

        return mainImage
    }

    mainImage := generateImage()

    group.AddElement(&uilib.UIElement{
        Rect: rect,
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(fade())
            options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))
            scale.DrawScaled(screen, mainImage, &options)
        },
        LeftClick: func(element *uilib.UIElement){
            fade = group.MakeFadeOut(7)
            group.AddDelay(7, cancel)
        },
    })

    game.doRunUI(yield, group, quit)
}

func (game *Game) doCartographer(yield coroutine.YieldFunc) {

    var stacks []*playerlib.UnitStack
    for _, player := range game.Players {
        if !player.Defeated {
            stacks = append(stacks, player.Stacks...)
        }
    }

    knownPlayers := append([]*playerlib.Player{game.GetHumanPlayer()}, game.GetHumanPlayer().GetKnownPlayers()...)

    logic, draw := cartographer.MakeCartographer(game.Cache, game.AllCities(), stacks, knownPlayers, game.ArcanusMap, game.GetHumanPlayer().GetFog(data.PlaneArcanus), game.MyrrorMap, game.GetHumanPlayer().GetFog(data.PlaneMyrror))

    yield()
    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    game.Drawer = func (screen *ebiten.Image, game *Game){
        oldDrawer(screen, game)
        draw(screen)
    }

    logic(yield)
}

func (game *Game) doRunUI(yield coroutine.YieldFunc, group *uilib.UIElementGroup, quit context.Context) {
    game.HudUI.AddGroup(group)
    defer game.HudUI.RemoveGroup(group)

    yield()
    for quit.Err() == nil {
        game.Counter += 1
        game.HudUI.StandardUpdate()
        if yield() != nil {
            break
        }
    }

    yield()
}

func ChooseUniqueWizard(players []*playerlib.Player, allSpells spellbook.Spells) (setup.WizardCustom, bool) {
    // pick a new wizard with an unused wizard base and banner color, and race
    // if on myrror then select a myrran race

    chooseBase := func() (setup.WizardSlot, bool) {
        choices := slices.Clone(setup.DefaultWizardSlots())
        choices = slices.DeleteFunc(choices, func (wizard setup.WizardSlot) bool {
            if wizard.Name == "Custom" {
                return true
            }

            for _, player := range players {
                if player.Wizard.Base == wizard.Base {
                    return true
                }
            }

            return false
        })

        if len(choices) == 0 {
            return setup.WizardSlot{}, false
        }

        return choices[rand.N(len(choices))], true
    }

    chooseRace := func(myrror bool) (data.Race, bool) {
        var choices []data.Race
        if myrror {
            choices = slices.Clone(data.MyrranRaces())
        } else {
            choices = slices.Clone(data.ArcanianRaces())
        }

        choices = slices.DeleteFunc(choices, func (race data.Race) bool {
            for _, player := range players {
                if player.Wizard.Race == race {
                    return true
                }
            }

            return false
        })

        if len(choices) == 0 {
            return data.RaceNone, false
        }

        return choices[rand.N(len(choices))], true
    }

    chooseBanner := func() (data.BannerType, bool) {
        choices := []data.BannerType{data.BannerGreen, data.BannerBlue, data.BannerRed, data.BannerPurple, data.BannerYellow}
        choices = slices.DeleteFunc(choices, func (banner data.BannerType) bool {
            for _, player := range players {
                if player.Wizard.Banner == banner {
                    return true
                }
            }

            return false
        })

        if len(choices) == 0 {
            return data.BannerGreen, false
        }

        return choices[rand.N(len(choices))], true
    }

    wizard, ok := chooseBase()

    if !ok {
        return setup.WizardCustom{}, false
    }

    race, ok := chooseRace(wizard.ExtraRetort == data.RetortMyrran)
    if !ok {
        return setup.WizardCustom{}, false
    }

    banner, ok := chooseBanner()
    if !ok {
        return setup.WizardCustom{}, false
    }

    var retorts []data.Retort
    if wizard.ExtraRetort != data.RetortNone {
        retorts = []data.Retort{wizard.ExtraRetort}
    }

    customWizard := setup.WizardCustom{
        Name: wizard.Name,
        Base: wizard.Base,
        Race: race,
        Books: slices.Clone(wizard.Books),
        Banner: banner,
        Retorts: retorts,
    }

    customWizard.StartingSpells.AddAllSpells(setup.GetStartingSpells(&customWizard, allSpells))
    return customWizard, true

}

/* returns a wizard definition and true if successful, otherwise false if no more wizards can be created
 */
func (game *Game) ChooseWizard() (setup.WizardCustom, bool) {
    return ChooseUniqueWizard(game.Players, game.AllSpells())
}

func (game *Game) RefreshUI() {
    select {
        case game.Events <- &GameEventRefreshUI{}:
        default:
    }
}

// returns screen coordinates in pixels of the middle of the given tile
// warning: this does not apply the screen scaling to the coordinates so you must still use
// scale.DrawScaled() to draw the image at the correct size/position
func (game *Game) TileToScreen(tileX int, tileY int) (int, int) {
    tileWidth := game.CurrentMap().TileWidth()
    tileHeight := game.CurrentMap().TileHeight()

    var geom ebiten.GeoM

    x, y := (game.CurrentMap().XDistance(game.Camera.GetX(), tileX) + game.Camera.GetX()) * tileWidth, tileY * tileHeight
    geom.Translate(float64(x + tileWidth / 2.0), float64(y + tileHeight / 2.0))
    geom.Translate(-game.Camera.GetZoomedX() * float64(tileWidth), -game.Camera.GetZoomedY() * float64(tileHeight))
    geom.Scale(game.Camera.GetAnimatedZoom(), game.Camera.GetAnimatedZoom())
    // geom.Concat(scale.ScaledGeom)

    outX, outY := geom.Apply(0, 0)
    return int(outX), int(outY)
}

// convert real screen coordinates to tile coordinates
func (game *Game) ScreenToTile(inX float64, inY float64) (int, int) {
    tileWidth := game.CurrentMap().TileWidth()
    tileHeight := game.CurrentMap().TileHeight()

    var geom ebiten.GeoM

    camera := game.Camera

    /*
    geom.Translate(6, 5)
    geom.Scale(float64(tileWidth), float64(tileHeight))
    geom.Scale(camera.GetAnimatedZoom(), camera.GetAnimatedZoom())
    */

    geom.Translate(-camera.GetZoomedX() * float64(tileWidth), -camera.GetZoomedY() * float64(tileHeight))
    geom.Scale(camera.GetAnimatedZoom(), camera.GetAnimatedZoom())
    geom.Concat(scale.ScaledGeom)

    geom.Invert()

    tileX, tileY := geom.Apply(inX, inY)

    tileX /= float64(tileWidth)
    tileY /= float64(tileHeight)

    // log.Printf("relative tile %v, %v camera %v, %v", tileX, tileY, game.Camera.GetX(), game.Camera.GetY())

    // return int(tileX + float64(game.Camera.GetX())), int(tileY + float64(game.Camera.GetY()))

    // return int(math.Floor(tileX)), int(math.Floor(tileY))
    return game.CurrentMap().WrapX(int(math.Floor(tileX))), int(math.Floor(tileY))
}

func (game *Game) doInputZoom(yield coroutine.YieldFunc) bool {
    // FIXME: move most of this code to the camera module

    inputLoop:
    for {
        _, wheelY := inputmanager.Wheel()

        // zoomSpeed := 5
        zoomSpeed2 := 7

        if wheelY > 0 {
            oldZoom := game.Camera.Zoom
            game.Camera.Zoom = min(game.Camera.Zoom + 1, camera.ZoomMax)
            game.Camera.AnimatedZoom = float64(oldZoom - game.Camera.Zoom)

            if oldZoom != game.Camera.Zoom {
                /*
                for i := range zoomSpeed {
                    game.AnimatedZoom = float64(i) / float64(zoomSpeed) - 1.0
                    yield()
                }
                */

                for i := 0; i < 90; i += zoomSpeed2 {
                    game.Camera.AnimatedZoom = math.Sin(float64(i) * math.Pi / 180.0) - 1
                    game.Counter += 1
                    yield()

                    _, wheelY := ebiten.Wheel()
                    if wheelY > 0 || wheelY < 0 {
                        continue inputLoop
                    }

                }

                game.Camera.AnimatedZoom = 0
            }

            return true
        } else if wheelY < 0 {
            oldZoom := game.Camera.Zoom
            game.Camera.Zoom = max(game.Camera.Zoom - 1, camera.ZoomMin)
            game.Camera.AnimatedZoom = float64(oldZoom - game.Camera.Zoom)

            if oldZoom != game.Camera.Zoom {
                /*
                for i := range zoomSpeed {
                    game.AnimatedZoom = 1.0 - float64(i) / float64(zoomSpeed)
                    yield()
                }
                */

                for i := 0; i < 90; i += zoomSpeed2 {
                    game.Camera.AnimatedZoom = 1.0 - math.Sin(float64(i) * math.Pi / 180.0)
                    game.Counter += 1
                    yield()

                    _, wheelY := ebiten.Wheel()
                    if wheelY > 0 || wheelY < 0 {
                        continue inputLoop
                    }
                }

                game.Camera.AnimatedZoom = 0
            }

            return true
        }

        return false
    }
}

func (game *Game) doMoveCamera(yield coroutine.YieldFunc, x int, y int) {
    camera := game.Camera

    camera.Center(x, y)
    minY := math.Floor(-1 / camera.GetZoom())
    for camera.GetZoomedY() < minY {
        y += 1
        camera.Center(x, y)
    }

    for camera.GetZoomedMaxY() >= float64(game.CurrentMap().Height()) && camera.Y > 0 {
        y -= 1
        camera.Center(x, y)
    }

    /*
    if y < 0 {
        y = 0
    }
    */

    if y > game.CurrentMap().Height() {
        y = game.CurrentMap().Height()
    }

    dx := game.CurrentMap().XDistance(game.Camera.GetX(), x)
    dy := y - game.Camera.GetY()
    length := math.Sqrt(float64(dx * dx + dy * dy))

    angle := math.Atan2(float64(dy), float64(dx))
    angle_cos := math.Cos(angle)
    angle_sin := math.Sin(angle)

    steps := 10

    for i := range steps {
        value := float64(i) / float64(steps) * math.Pi / 2
        magnitude := length * math.Sin(value)
        game.Camera.SetOffset(angle_cos * magnitude, angle_sin * magnitude)
        yield()
    }

    game.Camera.SetOffset(0, 0)
    game.Camera.Center(game.CurrentMap().WrapX(x), y)
}

// try to find a nearby position that the given unit can move to
func (game *Game) FindEscapePosition(player *playerlib.Player, unit units.StackUnit) []image.Point {
    x := unit.GetX()
    y := unit.GetY()
    plane := unit.GetPlane()
    mapUse := game.GetMap(plane)
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
            for _, enemy := range game.GetEnemies(player) {

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

func (game *Game) ResolveStackAt(x int, y int, plane data.Plane) {
    stack, player := game.FindStack(x, y, plane)
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

        positions := game.FindEscapePosition(player, unit)

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

// try to relocate a fleeing stack, kills units that are unable
func (game *Game) doMoveFleeingDefender(player *playerlib.Player, stack *playerlib.UnitStack) {
    stackUnits := stack.Units()

    for _, i := range rand.Perm(len(stackUnits)) {
        unit := stackUnits[i]
        positions := game.FindEscapePosition(player, unit)

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

// returns true if the city was razed, and the amount of gold plundered from the city
func (game *Game) defeatCity(yield coroutine.YieldFunc, attacker *playerlib.Player, attackerStack *playerlib.UnitStack, defender *playerlib.Player, city *citylib.City) (bool, int) {
    raze := false
    gold := defender.ComputePlunderedGold(city)

    if attacker.IsHuman() {
        raze = game.confirmRazeTown(yield, city)
    } else {
        raze = attacker.AIBehavior.ConfirmRazeTown(city)
    }

    containedFortress := city.Buildings.Contains(buildinglib.BuildingFortress)

    if raze {
        defender.RemoveCity(city)
    } else {
        ChangeCityOwner(city, defender, attacker, ChangeCityRemoveOwnerEnchantments)
    }

    if containedFortress {
        defender.Banished = true

        if attacker.IsHuman() || defender.IsHuman() {
            game.Events <- &GameEventShowBanish{Attacker: attacker, Defender: defender}
        }

        // FIXME: automatically start casting spell of return if possible
    }

    return raze, gold
}

func (game *Game) doBanish(yield coroutine.YieldFunc, attacker *playerlib.Player, defender *playerlib.Player) {
    banishLogic, banishDraw := banish.ShowBanishAnimation(game.Cache, attacker, defender)

    oldDrawer := game.Drawer
    defer func() {
        game.Drawer = oldDrawer
    }()

    game.Drawer = func(screen *ebiten.Image, game *Game){
        banishDraw(screen)
    }

    banishLogic(yield)

    yield()
}

func (game *Game) GetStackOwner(stack *playerlib.UnitStack) *playerlib.Player {
    for _, player := range game.Players {
        if player.OwnsStack(stack) {
            return player
        }
    }

    return nil
}

func (game *Game) GetCityOwner(city *citylib.City) *playerlib.Player {
    for _, player := range game.Players {
        if player.OwnsCity(city) {
            return player
        }
    }

    return nil
}

func (game *Game) doMoveSelectedUnit(yield coroutine.YieldFunc, player *playerlib.Player) {
    stack := player.SelectedStack
    if stack == nil || len(stack.ActiveUnits()) == 0 {
        return
    }

    mapUse := game.GetMap(stack.Plane())

    stepsTaken := 0
    stopMoving := false
    var mergeStack *playerlib.UnitStack
    // kind of a hack, in case the stack couldn't move due to spell ward or something we attempt to merge
    // the stack with whatever stack it is standing on
    for _, otherStack := range player.FindAllStacks(stack.X(), stack.Y(), stack.Plane()) {
        if otherStack != stack {
            mergeStack = otherStack
        }
    }

    getStack := func(x int, y int) (playerlib.PathStack, bool) {
        found := player.FindStack(mapUse.WrapX(x), y, stack.Plane())
        return found, found != nil
    }

    entityInfo := game.ComputeCityStackInfo()

    quitMoving:
    for i, step := range stack.CurrentPath {
        if stack.AnyOutOfMoves() {
            break
        }

        oldX := stack.X()
        oldY := stack.Y()

        city := entityInfo.FindCity(step.X, step.Y, stack.Plane())
        if city != nil {
            // units might not be able to enter a city if the city has spell wards in effect
            for _, unit := range stack.ActiveUnits() {
                if !city.CanEnter(unit) {
                    stopMoving = true
                    game.Events <- &GameEventNotice{Message: fmt.Sprintf("%v can not enter the city of %v", unit.GetRawUnit().Name, city.Name)}
                    break quitMoving
                }
            }
        }

        terrainCost, canMove := game.ComputeTerrainCost(stack, stack.X(), stack.Y(), step.X, step.Y, mapUse, getStack)

        if canMove {

            encounter := mapUse.GetEncounter(mapUse.WrapX(step.X), step.Y)
            if encounter != nil {
                if game.confirmLairEncounter(yield, encounter) {
                    stack.Move(step.X - stack.X(), step.Y - stack.Y(), terrainCost, game.GetNormalizeCoordinateFunc())
                    game.showMovement(yield, oldX, oldY, stack, true)
                    player.LiftFogSquare(stack.X(), stack.Y(), stack.GetSightRange(), stack.Plane())

                    stack.ExhaustMoves()
                    state := game.doEncounter(yield, player, stack, encounter, mapUse, stack.X(), stack.Y())
                    if state == combat.CombatStateAttackerFlee {
                        stack.SetX(oldX)
                        stack.SetY(oldY)
                    }

                    game.RefreshUI()
                } else {
                    encounter.ExploredBy.Insert(player)
                }

                stopMoving = true
                break quitMoving
            }

            stepsTaken = i + 1
            mergeStack = player.FindStack(mapUse.WrapX(step.X), step.Y, stack.Plane())

            stack.Move(step.X - stack.X(), step.Y - stack.Y(), terrainCost, game.GetNormalizeCoordinateFunc())
            game.showMovement(yield, oldX, oldY, stack, true)
            player.LiftFogSquare(stack.X(), stack.Y(), stack.GetSightRange(), stack.Plane())

            if entityInfo.ContainsEnemy(stack.X(), stack.Y(), stack.Plane(), player) {
                // FIXME: this should get all stacks at the given location and merge them into a single stack for combat
                otherStack := entityInfo.FindStack(stack.X(), stack.Y(), stack.Plane())
                if otherStack != nil {
                    otherCity := entityInfo.FindCity(stack.X(), stack.Y(), stack.Plane())
                    zone := combat.ZoneType{
                        City: otherCity,
                    }

                    defenderPlayer := game.GetStackOwner(otherStack)

                    // note: doCombat will already call defeatCity if the attacker wins the battle
                    state := game.doCombat(yield, player, stack, defenderPlayer, otherStack, zone)
                    if state == combat.CombatStateAttackerFlee {
                        stack.SetX(oldX)
                        stack.SetY(oldY)
                    } else if state == combat.CombatStateDefenderFlee {
                        game.doMoveFleeingDefender(defenderPlayer, otherStack)
                    }

                    stack.ExhaustMoves()
                    game.RefreshUI()

                    stopMoving = true
                    break quitMoving
                }

                // defeat any unguarded cities immediately
                otherCity := entityInfo.FindCity(stack.X(), stack.Y(), stack.Plane())
                if otherCity != nil {
                    defenderPlayer := game.GetCityOwner(otherCity)
                    raze, gold := game.defeatCity(yield, player, stack, defenderPlayer, otherCity)

                    // FIXME: show a notice about any fame won
                    player.Fame = max(0, player.Fame + otherCity.FameForCaptureOrRaze(!raze))
                    defenderPlayer.Fame = max(0, defenderPlayer.Fame + otherCity.FameForCaptureOrRaze(false))
                    player.Gold += gold
                    defenderPlayer.Gold -= gold

                    stack.ExhaustMoves()
                    game.RefreshUI()

                    stopMoving = true
                    break quitMoving

                }
            }

            // have to force the ui to refresh because we are not processing events here
            game.HudUI = game.MakeHudUI()

            // some units in the stack might not have any moves left
            /*
            beforeActive := len(stack.ActiveUnits())
            stack.EnableMovers()
            afterActive := len(stack.ActiveUnits())
            if afterActive > 0 && afterActive != beforeActive {
                // stopMoving = true
                break
            }
            */
        } else {
            // can't move, so abort the rest of the path
            stopMoving = true
            break
        }
    }

    if stopMoving {
        stack.CurrentPath = nil
    } else if stepsTaken > 0 {
        stack.CurrentPath = stack.CurrentPath[stepsTaken:]
    }

    // there might be some units in the stack that can still move, but if they move as a group
    // then all units should be out of moves once the unit with least amount of movement points is done
    if stack.GetRemainingMoves().IsZero() {
        stack.ExhaustMoves()
        game.DoNextUnit(player)
        game.RefreshUI()
    }

    // only merge stacks if both stacks are stopped, otherwise they can move through each other
    if len(stack.CurrentPath) == 0 && mergeStack != nil && mergeStack.X() == stack.X() && mergeStack.Y() == stack.Y() {
        stack = player.MergeStacks(mergeStack, stack)
        player.SelectedStack = stack
        game.RefreshUI()
    }

    // update unrest for new units in the city
    newCity := player.FindCity(stack.X(), stack.Y(), stack.Plane())
    if newCity != nil {
        newCity.UpdateUnrest()
    }

    /*
    if stepsTaken > 0 {
        if stack.AnyOutOfMoves() {
            stack.ExhaustMoves()
            game.DoNextUnit(player)
        }

        game.RefreshUI()
    }
    */
}

// given a position on the screen in pixels, return true if the position is within the area of the ui designated for the overworld
func (game *Game) InOverworldArea(x int, y int) bool {
    scaledX, scaledY := scale.Scale2(240, 18)
    return x < scaledX && y > scaledY
}

func (game *Game) doPlayerUpdate(yield coroutine.YieldFunc, player *playerlib.Player) {
    // log.Printf("Game.Update")
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    zoomed := game.doInputZoom(yield)
    _ = zoomed

    mouseX, mouseY := inputmanager.MousePosition()
    leftClick := inputmanager.LeftClick()
    rightClick := inputmanager.RightClick()

    if player.SelectedStack != nil && player.SelectedStack.Plane() == game.Plane {
        stack := player.SelectedStack
        mapUse := game.GetMap(stack.Plane())
        oldX := stack.X()
        oldY := stack.Y()

        if true || len(stack.CurrentPath) == 0 || stack.OutOfMoves() {

            dx := 0
            dy := 0

            for _, key := range keys {
                switch key {
                    case ebiten.KeyUp: dy = -1
                    case ebiten.KeyDown: dy = 1
                    case ebiten.KeyLeft: dx = -1
                    case ebiten.KeyRight: dx = 1
                }
            }

            newX := game.CurrentMap().WrapX(stack.X() + dx)
            newY := stack.Y() + dy

            if leftClick {
                // can only click into the area not hidden by the hud
                if game.InOverworldArea(mouseX, mouseY) {
                    // log.Printf("Click at %v, %v", mouseX, mouseY)
                    /*
                    realX, realY := game.RealToTile(float64(mouseX), float64(mouseY))
                    newX = game.cameraX + realX
                    newY = game.cameraY + realY
                    */
                    newX, newY = game.ScreenToTile(float64(mouseX), float64(mouseY))
                    // log.Printf("Click at %v, %v -> %v, %v", mouseX, mouseY, newX, newY)
                    newX = game.CurrentMap().WrapX(newX)
                }
            }

            if newX != oldX || newY != oldY {
                activeUnits := stack.ActiveUnits()
                if len(activeUnits) > 0 {
                    if newY >= 0 && newY < mapUse.Height() {

                        var inactiveStack *playerlib.UnitStack

                        inactiveUnits := stack.InactiveUnits()
                        if len(inactiveUnits) > 0 {
                            stack.RemoveUnits(inactiveUnits)
                            inactiveStack = player.AddStack(playerlib.MakeUnitStackFromUnits(inactiveUnits))
                            game.RefreshUI()
                        }

                        oldCity := player.FindCity(oldX, oldY, stack.Plane())
                        newCity := player.FindCity(newX, newY, stack.Plane())

                        // unit can move instantly to the new city if they are standing on a city with earth gate
                        // and the new city also has earth gate
                        if oldCity != nil && newCity != nil && oldCity.HasEnchantment(data.CityEnchantmentEarthGate) && newCity.HasEnchantment(data.CityEnchantmentEarthGate) && stack.GetRemainingMoves().GreaterThan(fraction.Zero()) {
                            // the other city might be protected by spell ward
                            canMove := true
                            for _, unit := range stack.ActiveUnits() {
                                if !newCity.CanEnter(unit) {
                                    canMove = false
                                    game.Events <- &GameEventNotice{Message: fmt.Sprintf("%v can not enter the city of %v", unit.GetRawUnit().Name, newCity.Name)}
                                    break
                                }
                            }

                            if canMove {
                                stack.UseMovement(fraction.FromInt(1))
                                newCityStack := player.FindStack(newX, newY, stack.Plane())
                                if newCityStack != nil {
                                    player.MergeStacks(newCityStack, stack)
                                } else {
                                    stack.SetX(newX)
                                    stack.SetY(newY)
                                }
                                game.RefreshUI()
                            }
                        } else {
                            path := game.FindPath(oldX, oldY, newX, newY, player, stack, player.GetFog(game.Plane))
                            if len(path) == 0 {
                                game.blinkRed(yield)
                                if inactiveStack != nil {
                                    player.MergeStacks(stack, inactiveStack)
                                }
                            } else {
                                // FIXME: i'm not sure this can ever occur in practice
                                if inactiveStack != nil {
                                    inactiveStack.CurrentPath = stack.CurrentPath
                                }
                                stack.CurrentPath = path

                                select {
                                    case game.Events <- &GameEventMoveUnit{Player: player}:
                                    default:
                                }
                            }
                        }
                    }
                } else {
                    // make a copy of the unit stack to activate all units, because path finding only checks active units for terrain constraints
                    path := game.FindPath(oldX, oldY, newX, newY, player, playerlib.MakeUnitStackFromUnits(stack.Units()), player.GetFog(game.Plane))
                    if len(path) == 0 {
                        game.blinkRed(yield)
                    } else {
                        stack.CurrentPath = path
                    }
                }
            } else if leftClick && game.InOverworldArea(mouseX, mouseY) {
                stack.CurrentPath = nil
            }
        }
    }

    if rightClick /*|| zoomed*/ {
        // mapUse := game.CurrentMap()
        mouseX, mouseY := inputmanager.MousePosition()

        // can only click into the area not hidden by the hud
        if game.InOverworldArea(mouseX, mouseY) {
            // log.Printf("Click at %v, %v", mouseX, mouseY)
            // realX, realY := game.RealToTile(float64(mouseX), float64(mouseY))
            tileX, tileY := game.ScreenToTile(float64(mouseX), float64(mouseY))

            // log.Printf("camera %v, %v, right click at %v, %v", game.cameraX, game.cameraY, realX, realY)

            // log.Printf("height %v tile height %v", data.ScreenHeight, game.CurrentMap().TileHeight())

            // log.Printf("Click at %v, %v -> %v, %v", mouseX, mouseY, tileX, tileY)

            game.doMoveCamera(yield, tileX, tileY)
            tileX = game.CurrentMap().WrapX(tileX)

            if rightClick {
                city := player.FindCity(tileX, tileY, game.Plane)
                if city != nil {
                    if city.Outpost {
                        game.showOutpost(yield, city, player.FindStack(city.X, city.Y, city.Plane), player, false)
                    } else {
                        game.doCityScreen(yield, city, player, buildinglib.BuildingNone)
                    }
                    game.RefreshUI()
                } else {
                    stack := player.FindStack(tileX, tileY, game.Plane)
                    if stack != nil {
                        player.SelectedStack = stack
                        game.RefreshUI()
                    } else {
                        for _, otherPlayer := range game.Players {
                            if otherPlayer == player {
                                continue
                            }

                            city := otherPlayer.FindCity(tileX, tileY, game.Plane)
                            if city != nil {
                                if player.Admin {
                                    game.doCityScreen(yield, city, otherPlayer, buildinglib.BuildingNone)
                                } else {
                                    game.doEnemyCityView(yield, city, player, otherPlayer)
                                }
                            } else if player.IsVisible(tileX, tileY, game.Plane) {
                                enemyStack := otherPlayer.FindStack(tileX, tileY, game.Plane)
                                if enemyStack != nil {
                                    quit := false
                                    clicked := func(unit unitview.UnitView){
                                        quit = true
                                    }

                                    var unitViewElements []unitview.UnitView
                                    for _, unit := range enemyStack.Units() {
                                        unitViewElements = append(unitViewElements, unit)
                                    }

                                    game.HudUI.AddElements(unitview.MakeSmallListView(game.Cache, game.HudUI, unitViewElements, otherPlayer.Wizard.Name, clicked))
                                    for !quit {
                                        game.Counter += 1
                                        game.HudUI.StandardUpdate()
                                        yield()
                                    }
                                }
                            }
                        }

                    }
                }
            }
        }
    }
}

func (game *Game) Update(yield coroutine.YieldFunc) GameState {
    game.Counter += 1

    /*
    if game.Counter % 10 == 0 {
        log.Printf("TPS: %v FPS: %v", ebiten.ActualTPS(), ebiten.ActualFPS())
    }
    */

    game.ProcessEvents(yield)

    switch game.State {
        case GameStateRunning:
            game.HudUI.StandardUpdate()

            // kind of a hack to not allow player to interact with anything other than the current ui modal
            if len(game.Players) > 0 && game.CurrentPlayer >= 0 {
                player := game.Players[game.CurrentPlayer]

                if player.IsHuman() {
                    if game.HudUI.GetHighestLayerValue() == 0 {
                        game.doPlayerUpdate(yield, player)
                    }
                } else {
                    game.doAiUpdate(yield, player)
                }
            }
    }

    return game.State
}

// FIXME: can this just use doMoveSelectedUnit?
// returns the rest of the path the stack should walk (nil if the path ends)
func (game *Game) doAiMoveUnit(yield coroutine.YieldFunc, player *playerlib.Player, move *playerlib.AIMoveStackDecision) pathfinding.Path {
    stack := move.Stack

    if len(move.Units) > 0 {
        stack = player.SplitStack(stack, move.Units)
        for _, unit := range stack.Units() {
            unit.SetBusy(units.BusyStatusNone)
        }
    }

    // FIXME: split the stack into just the active units in case some are busy or out of moves
    path := move.Path

    if len(path) == 0 {
        return nil
    }

    to := path[0]
    path = path[1:]

    log.Printf("  moving stack %v to %v, %v", stack, to.X, to.Y)
    getStack := func(x int, y int) (playerlib.PathStack, bool) {
        found := player.FindStack(x, y, stack.Plane())
        return found, found != nil
    }
    terrainCost, ok := game.ComputeTerrainCost(stack, stack.X(), stack.Y(), to.X, to.Y, game.GetMap(stack.Plane()), getStack)
    if ok {
        oldX := stack.X()
        oldY := stack.Y()

        mapUse := game.GetMap(stack.Plane())

        encounter := mapUse.GetEncounter(mapUse.WrapX(to.X), to.Y)
        if encounter != nil {
            if !move.ConfirmEncounter(encounter) {
                move.Invalid()
                return nil
            }
        }

        newCity, _ := game.FindCity(to.X, to.Y, stack.Plane())
        if newCity != nil {
            for _, unit := range stack.ActiveUnits() {
                if !newCity.CanEnter(unit) {
                    move.Invalid()
                    return nil
                }
            }
        }

        stack.Move(to.X - stack.X(), to.Y - stack.Y(), terrainCost, game.GetNormalizeCoordinateFunc())

        if game.GetHumanPlayer().IsVisible(oldX, oldY, stack.Plane()) {
            game.showMovement(yield, oldX, oldY, stack, false)
        }

        move.Moved()

        player.LiftFogSquare(stack.X(), stack.Y(), stack.GetSightRange(), stack.Plane())

        if encounter != nil {
            game.doEncounter(yield, player, stack, encounter, mapUse, stack.X(), stack.Y())
            return nil
        }

        for _, enemy := range game.GetEnemies(player) {
            // FIXME: this should get all stacks at the given location and merge them into a single stack for combat
            enemyStack := enemy.FindStack(stack.X(), stack.Y(), stack.Plane())
            if enemyStack != nil {
                city := enemy.FindCity(stack.X(), stack.Y(), stack.Plane())
                zone := combat.ZoneType{
                    City: city,
                }
                state := game.doCombat(yield, player, stack, enemy, enemyStack, zone)

                if state == combat.CombatStateAttackerFlee {
                    stack.SetX(oldX)
                    stack.SetY(oldY)
                } else if state == combat.CombatStateDefenderFlee {
                    game.doMoveFleeingDefender(enemy, enemyStack)
                }

                return nil
            }

            city := enemy.FindCity(stack.X(), stack.Y(), stack.Plane())
            if city != nil {
                raze, gold := game.defeatCity(yield, player, stack, enemy, city)

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
        move.Invalid()
        return nil
    }

    return path
}

func (game *Game) doAiUpdate(yield coroutine.YieldFunc, player *playerlib.Player) {
    log.Printf("AI %v year %v: make decisions", player.Wizard.Name, game.TurnNumber)

    var decisions []playerlib.AIDecision

    if player.AIBehavior != nil {
        decisions = player.AIBehavior.Update(player, game.GetEnemies(player), game, player.ManaPerTurn(game.ComputePower(player), game))
        log.Printf("AI %v Decisions: %v", player.Wizard.Name, decisions)

        for _, decision := range decisions {
            switch decision.(type) {
                case *playerlib.AIMoveStackDecision:
                    moveDecision := decision.(*playerlib.AIMoveStackDecision)
                    newPath := game.doAiMoveUnit(yield, player, moveDecision)
                    moveDecision.Stack.CurrentPath = newPath

                // mainly for the raider ai
                case *playerlib.AICreateUnitDecision:
                    create := decision.(*playerlib.AICreateUnitDecision)
                    log.Printf("ai %v creating %+v", player.Wizard.Name, create)

                    overworldUnit := units.MakeOverworldUnitFromUnit(create.Unit, create.X, create.Y, create.Plane, player.Wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider())
                    if create.Patrol {
                        overworldUnit.SetBusy(units.BusyStatusPatrol)
                    }
                    player.AddUnit(overworldUnit)
                    game.ResolveStackAt(create.X, create.Y, create.Plane)
                case *playerlib.AIUpdateCityDecision:
                    update := decision.(*playerlib.AIUpdateCityDecision)

                    update.City.Farmers = update.Farmers
                    update.City.Workers = update.Workers
                    update.City.ResetCitizens()

                case *playerlib.AIBuildOutpostDecision:
                    build := decision.(*playerlib.AIBuildOutpostDecision)

                    var stack units.StackUnit
                    for _, unit := range build.Stack.Units() {
                        if unit.HasAbility(data.AbilityCreateOutpost) {
                            stack = unit
                            break
                        }
                    }

                    if stack != nil {
                        game.CreateOutpost(stack, player)
                    }
                case *playerlib.AIProduceDecision:
                    produce := decision.(*playerlib.AIProduceDecision)
                    log.Printf("ai %v producing %v %v", player.Wizard.Name, game.BuildingInfo.Name(produce.Building), produce.Unit.Name)
                    if produce.Building != buildinglib.BuildingNone {
                        produce.City.ProducingBuilding = produce.Building
                    } else {
                        produce.City.ProducingUnit = produce.Unit
                    }
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
            }
        }

        player.AIBehavior.PostUpdate(player, game.GetEnemies(player))
    }

    // if len(decisions) == 0 {
        game.DoNextTurn()
    // }
}

// get all alive players that are not the current player
func (game *Game) GetEnemies(player *playerlib.Player) []*playerlib.Player {
    var out []*playerlib.Player
    for _, enemy := range game.Players {
        if enemy != player && len(enemy.Cities) > 0 {
            out = append(out, enemy)
        }
    }
    return out
}

func (game *Game) doEnemyCityView(yield coroutine.YieldFunc, city *citylib.City, player *playerlib.Player, otherPlayer *playerlib.Player){
    drawer := game.Drawer
    defer func(){
        game.Drawer = drawer
    }()

    logic, draw := cityview.SimplifiedView(game.Cache, city, player, otherPlayer)

    game.Drawer = func(screen *ebiten.Image, game *Game){
        drawer(screen, game)
        draw(screen)
    }

    logic(yield, func(){
        game.Counter += 1
    })

    yield()
}

/* show a view of the city
 */
func (game *Game) doCityScreen(yield coroutine.YieldFunc, city *citylib.City, player *playerlib.Player, newBuilding buildinglib.Building){
    cityScreen := cityview.MakeCityScreen(game.Cache, city, player, newBuilding)

    var cities []*citylib.City
    var stacks []*playerlib.UnitStack
    var fog data.FogMap

    for i, player := range game.Players {
        for _, city := range player.Cities {
            if city.Plane == game.Plane {
                cities = append(cities, city)
            }
        }

        for _, stack := range player.Stacks {
            if stack.Plane() == game.Plane {
                stacks = append(stacks, stack)
            }
        }

        if i == 0 {
            fog = player.GetFog(game.Plane)
        }
    }

    mapUse := game.GetMap(city.Plane)

    overworld := Overworld{
        Camera: camera.MakeCameraAt(city.X, city.Y).UpdateSize(4, 4),
        Counter: 0,
        Map: mapUse,
        Cities: cities,
        Stacks: stacks,
        SelectedStack: nil,
        ImageCache: &game.ImageCache,
        Fog: fog,
        ShowAnimation: false,
        FogBlack: game.GetFogImage(),
    }

    catchment := mapUse.GetCatchmentArea(city.X, city.Y)

    var sharedPoints []image.Point
    for _, tile := range catchment {
        if tile.IsShared {
            sharedPoints = append(sharedPoints, image.Point{X: tile.X - city.X, Y: tile.Y - city.Y})
        }
    }

    oldDrawer := game.Drawer
    halfTile, _ := game.ImageCache.GetImage("backgrnd.lbx", 0, 0)
    game.Drawer = func(screen *ebiten.Image, game *Game){
        cityScreen.Draw(screen, func (mapView *ebiten.Image, geom ebiten.GeoM, counter uint64){
            overworld.DrawOverworld(mapView, geom)

            var options ebiten.DrawImageOptions
            for _, point := range sharedPoints {
                options.GeoM = geom
                // move to center of the map
                options.GeoM.Translate(float64(overworld.Camera.SizeX / 2 * mapUse.TileWidth()), float64(overworld.Camera.SizeY / 2 * mapUse.TileHeight()))
                // offset by the shared tile position
                options.GeoM.Translate(float64(point.X * mapUse.TileWidth()), float64(point.Y * mapUse.TileHeight()))
                options.GeoM.Translate(float64(mapUse.TileWidth() / 2 - halfTile.Bounds().Dx() / 2), float64(mapUse.TileHeight() / 2 - halfTile.Bounds().Dy() / 2))
                // zoom will not affect the tile
                scale.DrawScaled(mapView, halfTile, &options)
            }

        }, mapUse.TileWidth(), mapUse.TileHeight())
    }

    for cityScreen.Update() == cityview.CityScreenStateRunning {
        overworld.Counter += 1

        select {
            case event := <-cityScreen.Actions:
                switch event {
                    case cityview.CityScreenActionChangeName:
                        yield()
                        city.Name = game.doInput(yield, "Set Name", city.Name, 80, 100)
                        cityScreen.ResetUI()
                }
            default:
        }

        yield()
    }

    yield()

    game.Drawer = oldDrawer
}

// similar to confirmEncounter() but without the buttons
func (game *Game) showEncounter(yield coroutine.YieldFunc, message string, animation *util.Animation) {
    quit := false

    dismiss := func(){
        quit = true
    }

    game.HudUI.AddElements(uilib.MakeLairShowDialogWithLayer(game.HudUI, game.Cache, &game.ImageCache, animation, 1, message, dismiss))

    for !quit {
        game.Counter += 1
        if game.Counter % 6 == 0 {
            animation.Next()
        }
        game.HudUI.StandardUpdate()
        yield()
    }

    yield()
}

func (game *Game) confirmEncounter(yield coroutine.YieldFunc, message string, animation *util.Animation) bool {
    quit := false

    result := false

    yes := func(){
        quit = true
        result = true
    }

    no := func(){
        quit = true
    }

    game.HudUI.AddElements(uilib.MakeLairConfirmDialogWithLayer(game.HudUI, game.Cache, &game.ImageCache, animation, 1, message, yes, no))

    yield()
    for !quit {
        game.Counter += 1
        if game.Counter % 6 == 0 {
            animation.Next()
        }
        game.HudUI.StandardUpdate()
        yield()
    }

    return result
}

// returns true to raze the town, false to occupy it
func (game *Game) confirmRazeTown(yield coroutine.YieldFunc, city *citylib.City) bool {
    raze := false
    quit := false
    yes := func(){
        raze = true
        quit = true
    }

    no := func(){
        raze = false
        quit = true
    }

    yesImages, _ := game.ImageCache.GetImages("compix.lbx", 82)
    noImages, _ := game.ImageCache.GetImages("compix.lbx", 81)

    ui := &uilib.UI{
        Cache: game.Cache,
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            ui.StandardDraw(screen)
        },
    }

    ui.SetElementsFromArray(nil)

    group := uilib.MakeGroup()

    group.AddElements(uilib.MakeConfirmDialogWithLayerFull(group, game.Cache, &game.ImageCache, 1, "Do you wish to completely destroy this city?", true, no, yes, noImages, yesImages))
    ui.AddGroup(group)

    oldDrawer := game.Drawer
    game.Drawer = func(screen *ebiten.Image, game *Game){
        oldDrawer(screen, game)
        ui.Draw(ui, screen)
    }
    defer func(){
        game.Drawer = oldDrawer
    }()

    yield()
    for !quit {
        game.Counter += 1
        ui.StandardUpdate()
        yield()
    }

    yield()

    return raze
}

func (game *Game) confirmLairEncounter(yield coroutine.YieldFunc, encounter *maplib.ExtraEncounter) bool {
    reloadLbx, err := game.Cache.GetLbxFile("reload.lbx")
    if err != nil {
        return false
    }

    lairIndex := 13
    article := "a"
    animated := false

    switch encounter.Type {
        case maplib.EncounterTypeLair:
            lairIndex = 17
        case maplib.EncounterTypeCave:
            lairIndex = 17
        case maplib.EncounterTypePlaneTower:
            lairIndex = 9
        case maplib.EncounterTypeAncientTemple:
            lairIndex = 15
            article = "an"
        case maplib.EncounterTypeFallenTemple:
            lairIndex = 19
        case maplib.EncounterTypeRuins:
            lairIndex = 18
            article = "some"
        case maplib.EncounterTypeAbandonedKeep:
            lairIndex = 16
            article = "an"
        case maplib.EncounterTypeDungeon:
            lairIndex = 14
        case maplib.EncounterTypeChaosNode:
            lairIndex = 10
            animated = true
        case maplib.EncounterTypeNatureNode:
            lairIndex = 11
            animated = true
        case maplib.EncounterTypeSorceryNode:
            lairIndex = 12
            animated = true
    }

    guardianName := ""
    if len(encounter.Units) > 0 {
        guardianName = encounter.Units[0].Name
    }

    pic, _ := game.ImageCache.GetImage("reload.lbx", lairIndex, 0)

    animation := util.MakeAnimation([]*ebiten.Image{pic}, true)
    if animated {
        rotateIndexLow := 247
        rotateIndexHigh := 254
        animation = util.MakePaletteRotateAnimation(reloadLbx, lairIndex, rotateIndexLow, rotateIndexHigh)
    }

    game.Music.PushSong(music.SongSiteDiscovery)
    defer game.Music.PopSong()

    if len(encounter.Units) == 0 {
        game.showEncounter(yield, fmt.Sprintf("You have found %v %v.", article, encounter.Type.Name()), animation)
        return true
    }

    return game.confirmEncounter(yield, fmt.Sprintf("You have found %v %v. Scouts have spotted %v within the %v. Do you wish to enter?", article, encounter.Type.Name(), guardianName, encounter.Type.Name()), animation)
}

func (game *Game) doEncounter(yield coroutine.YieldFunc, player *playerlib.Player, stack *playerlib.UnitStack, encounter *maplib.ExtraEncounter, mapUse *maplib.Map, x int, y int) combat.CombatState {
    // there was nothing in the encounter, just give treasure
    if len(encounter.Units) == 0 {
        mapUse.RemoveEncounter(x, y)
        game.createTreasure(encounter.Type, encounter.Budget, player)
        yield()
        return combat.CombatStateNoCombat
    }

    wizard := setup.WizardCustom{
        Name: encounter.Type.ShortName(),
    }

    defender := playerlib.MakePlayer(wizard, false, 1, 1, make(map[herolib.HeroType]string), game)
    defender.StrategicCombat = true

    var enemies []units.StackUnit

    for _, unit := range encounter.Units {
        enemies = append(enemies, units.MakeOverworldUnit(unit, x, y, mapUse.Plane))
    }

    zone := combat.ZoneType{
    }

    switch encounter.Type {
        case maplib.EncounterTypeLair, maplib.EncounterTypeCave: zone.Encounter = combat.ZoneLair
        case maplib.EncounterTypePlaneTower: zone.Encounter = combat.ZoneTower
        case maplib.EncounterTypeAncientTemple: zone.Encounter = combat.ZoneAncientTemple
        case maplib.EncounterTypeFallenTemple: zone.Encounter = combat.ZoneFallenTemple
        case maplib.EncounterTypeRuins: zone.Encounter = combat.ZoneRuins
        case maplib.EncounterTypeAbandonedKeep: zone.Encounter = combat.ZoneAbandonedKeep
        case maplib.EncounterTypeDungeon: zone.Encounter = combat.ZoneDungeon
        case maplib.EncounterTypeNatureNode: zone.Encounter = combat.ZoneNatureNode
        case maplib.EncounterTypeSorceryNode: zone.Encounter = combat.ZoneSorceryNode
        case maplib.EncounterTypeChaosNode: zone.Encounter = combat.ZoneChaosNode
    }

    result := game.doCombat(yield, player, stack, defender, playerlib.MakeUnitStackFromUnits(enemies), zone)
    if result == combat.CombatStateAttackerWin {
        mapUse.RemoveEncounter(x, y)

        game.createTreasure(encounter.Type, encounter.Budget, player)

        // defeating a plane tower also removes the tower from the other plane
        if encounter.Type == maplib.EncounterTypePlaneTower {
            mapUse.SetPlaneTower(x, y)
            otherMap := game.GetMap(mapUse.Plane.Opposite())
            otherMap.RemoveEncounter(x, y)
            otherMap.SetPlaneTower(x, y)
        }

    } else {
        var remaining []units.Unit
        for index := range enemies {
            if enemies[index].GetHealth() > 0 {
                remaining = append(remaining, encounter.Units[index])
            }
        }
        encounter.Units = remaining
    }

    // absorb extra clicks
    yield()

    return result
}

func (game *Game) createTreasure(encounterType maplib.EncounterType, budget int, player *playerlib.Player){
    allSpells, err := spellbook.ReadSpellsFromCache(game.Cache)
    if err != nil {
        log.Printf("Error: unable to read spells: %v", err)
    } else {
        var heroes []*herolib.Hero
        for _, hero := range player.HeroPool {
            // only include available heroes that are not champions
            if hero.Status == herolib.StatusAvailable && !hero.IsChampion() {
                heroes = append(heroes, hero)
            }
        }

        makeArtifacts := func () []*artifact.Artifact {
            var out []*artifact.Artifact
            for _, artifact := range game.ArtifactPool {
                out = append(out, artifact)
            }
            return out
        }

        treasure := makeTreasure(game.Cache, encounterType, budget, player.Wizard, player.KnownSpells, allSpells, heroes, makeArtifacts)
        // FIXME: show treasure ui for human, otherwise just apply treasure for AI
        select {
            case game.Events <- &GameEventTreasure{Treasure: treasure, Player: player}:
            default:
        }
    }
}

type TreasureFonts struct {
    TreasureFont *font.Font
}

func MakeTreasureFonts(cache *lbx.LbxCache) *TreasureFonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Error loading fonts: %v", err)
        return nil
    }

    return &TreasureFonts{
        TreasureFont: loader(fontslib.LightGradient1),
    }
}

func (game *Game) doTreasurePopup(yield coroutine.YieldFunc, player *playerlib.Player, treasure Treasure){
    uiDone := false

    fonts := MakeTreasureFonts(game.Cache)

    getAlpha := util.MakeFadeIn(7, &game.Counter)

    element := &uilib.UIElement{
        Layer: 2,
        Rect: image.Rect(0, 0, data.ScreenWidth, data.ScreenHeight),
        LeftClick: func (element *uilib.UIElement){
            uiDone = true
        },
        Draw: func (element *uilib.UIElement, screen *ebiten.Image){
            left, _ := game.ImageCache.GetImage("resource.lbx", 56, 0)
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            options.GeoM.Translate(float64(10), float64(50))

            fontX, fontY := options.GeoM.Apply(float64(10), float64(10))

            scale.DrawScaled(screen, left, &options)
            right, _ := game.ImageCache.GetImage("resource.lbx", 58, 0)
            options.GeoM.Translate(float64(left.Bounds().Dx()), 0)
            rightGeom := options.GeoM

            chest, _ := game.ImageCache.GetImage("reload.lbx", 20, 0)
            options.GeoM.Translate(float64(6), float64(8))
            scale.DrawScaled(screen, chest, &options)

            options.GeoM = rightGeom
            scale.DrawScaled(screen, right, &options)

            fonts.TreasureFont.PrintWrap(screen, fontX, fontY, float64(left.Bounds().Dx() - 5), font.FontOptions{DropShadow: true, Options: &options, Scale: scale.ScaleAmount}, treasure.String())
        },
    }

    game.HudUI.AddElement(element)

    for !uiDone {
        game.Counter += 1
        game.HudUI.StandardUpdate()
        yield()
    }

    getAlpha = util.MakeFadeOut(7, &game.Counter)

    for range 7 {
        game.Counter += 1
        game.HudUI.StandardUpdate()
        yield()
    }

    yield()

    game.HudUI.RemoveElement(element)
}

func (game *Game) ApplyTreasure(yield coroutine.YieldFunc, player *playerlib.Player, treasure Treasure) {
    for _, item := range treasure.Treasures {
        switch item.(type) {
            case *TreasureGold:
                gold := item.(*TreasureGold)
                player.Gold += gold.Amount
            case *TreasureMana:
                mana := item.(*TreasureMana)
                player.Mana += mana.Amount
            case *TreasureMagicalItem:
                magicalItem := item.(*TreasureMagicalItem)
                if player.IsHuman() {
                    game.doVault(yield, magicalItem.Artifact)
                } else {
                    // FIXME: ai should get the vault item
                }
                // if the treasure was one of the premade artifacts, then remove it from the pool
                delete(game.ArtifactPool, magicalItem.Artifact.Name)
            case *TreasurePrisonerHero:
                hero := item.(*TreasurePrisonerHero)
                if player.IsHuman() {
                    game.doHireHero(yield, 0, hero.Hero, player)
                } else {
                    // FIXME: ai should get a chance to hire the hero
                }
            case *TreasureSpell:
                spell := item.(*TreasureSpell)
                if player.IsHuman() {
                    game.doLearnSpell(yield, player, spell.Spell)
                }
                player.LearnSpell(spell.Spell)
            case *TreasureSpellbook:
                spellbook := item.(*TreasureSpellbook)
                player.Wizard.AddMagicLevel(spellbook.Magic, spellbook.Count)
                allSpells := game.AllSpells()
                // add more researchable spells based on the new magic level
                player.InitializeResearchableSpells(&allSpells)
            case *TreasureRetort:
                retort := item.(*TreasureRetort)
                player.Wizard.EnableRetort(retort.Retort)
        }
    }
}

func (game *Game) GetCombatLandscape(x int, y int, plane data.Plane) combat.CombatLandscape {
    tile := game.GetMap(plane).GetTile(x, y)

    switch tile.Tile.TerrainType() {
        case terrain.Hill, terrain.Grass,
             terrain.Forest, terrain.River, terrain.Shore,
             terrain.Swamp: return combat.CombatLandscapeGrass

        case terrain.Desert: return combat.CombatLandscapeDesert
        case terrain.Mountain: return combat.CombatLandscapeMountain
        case terrain.Tundra: return combat.CombatLandscapeTundra

        // FIXME: these cases are special
        case terrain.Ocean: return combat.CombatLandscapeWater
        case terrain.Volcano: return combat.CombatLandscapeGrass
        case terrain.Lake: return combat.CombatLandscapeGrass
        case terrain.NatureNode: return combat.CombatLandscapeGrass
        case terrain.SorceryNode: return combat.CombatLandscapeGrass
        case terrain.ChaosNode: return combat.CombatLandscapeMountain
    }

    return combat.CombatLandscapeGrass
}

// get the kind of magic that is influencing the given tile
func (game *Game) GetInfluenceMagic(x int, y int, plane data.Plane) data.MagicType {
    map_ := game.GetMap(plane)

    node := map_.GetMagicInfluence(x, y)
    if node != nil {
        return node.Kind.MagicType()
    }

    return data.MagicNone
}

// true if any alive player has the given enchantment enabled
func (game *Game) HasEnchantment(enchantment data.Enchantment) bool {
    for _, player := range game.Players {
        if !player.Defeated && player.HasEnchantment(enchantment) {
            return true
        }
    }

    return false
}

// true if any alive player that is not the given one has the given enchantment enabled
func (game *Game) HasRivalEnchantment(original *playerlib.Player, enchantment data.Enchantment) bool {
    for _, player := range game.Players {
        if !player.Defeated && player != original && player.HasEnchantment(enchantment) {
            return true
        }
    }

    return false
}

// 5% chance to destroy a building in every city, and 15% chance to destroy a garrisoned unit
func (game *Game) maybeDoNaturesWrath(caster *playerlib.Player) {
    if game.HasRivalEnchantment(caster, data.EnchantmentNaturesWrath) {
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

/* run the tactical combat screen. returns the combat state as a result (attackers win, defenders win, flee, etc)
 * this also shows the raze city ui so that fame can be incorporated based on whether the city is razed or not
 */
func (game *Game) doCombat(yield coroutine.YieldFunc, attacker *playerlib.Player, attackerStack *playerlib.UnitStack, defender *playerlib.Player, defenderStack *playerlib.UnitStack, zone combat.ZoneType) combat.CombatState {
    landscape := game.GetCombatLandscape(defenderStack.X(), defenderStack.Y(), defenderStack.Plane())

    createArmy := func (player *playerlib.Player, stack *playerlib.UnitStack) *combat.Army {
        army := combat.Army{
            Player: player,
        }

        for _, unit := range stack.Units() {
            // flying units always get added to battle
            if !unit.IsFlying() {
                if landscape == combat.CombatLandscapeWater && unit.IsLandWalker() {
                    continue
                }

                // dont add sailing units to non-water combat
                if landscape != combat.CombatLandscapeWater && unit.IsSailing() {
                    continue
                }
            }

            army.AddUnit(unit)
        }

        return &army
    }

    attackingArmy := createArmy(attacker, attackerStack)
    defendingArmy := createArmy(defender, defenderStack)

    attackingArmy.LayoutUnits(combat.TeamAttacker)
    defendingArmy.LayoutUnits(combat.TeamDefender)

    var state combat.CombatState
    var defeatedDefenders int
    var defeatedAttackers int
    var recalledAttackers []units.StackUnit
    var recalledDefenders []units.StackUnit

    oldDrawer := game.Drawer
    var combatScreen *combat.CombatScreen

    strategicCombat := attacker.StrategicCombat && defender.StrategicCombat

    if strategicCombat {
        state, defeatedAttackers, defeatedDefenders = combat.DoStrategicCombat(attackingArmy, defendingArmy)
        log.Printf("Strategic combat result state=%v", state)
    } else {

        defer mouse.Mouse.SetImage(game.MouseData.Normal)

        combatScreen = combat.MakeCombatScreen(game.Cache, defendingArmy, attackingArmy, game.Players[0], landscape, attackerStack.Plane(), zone, game.GetInfluenceMagic(attackerStack.X(), attackerStack.Y(), attackerStack.Plane()), attackerStack.X(), attackerStack.Y())

        if zone.City != nil && zone.City.HasEnchantment(data.CityEnchantmentHeavenlyLight) {
            combatScreen.Model.AddGlobalEnchantment(data.CombatEnchantmentTrueLight)
        }
        if zone.City != nil && zone.City.HasEnchantment(data.CityEnchantmentCloudOfShadow) {
            combatScreen.Model.AddGlobalEnchantment(data.CombatEnchantmentDarkness)
        } else {
            for _, enchantments := range game.GetAllGlobalEnchantments() {
                if enchantments.Contains(data.EnchantmentEternalNight) {
                    combatScreen.Model.AddGlobalEnchantment(data.CombatEnchantmentDarkness)
                    break
                }
            }
        }

        // ebiten.SetCursorMode(ebiten.CursorModeHidden)

        game.Drawer = func (screen *ebiten.Image, game *Game){
            combatScreen.Draw(screen)
        }

        game.Music.PushSong(randomChoose(music.SongCombat1, music.SongCombat2))

        state = combat.CombatStateRunning
        for state == combat.CombatStateRunning {
            state = combatScreen.Update(yield)
            yield()
        }

        game.Music.PopSong()

        // FIXME: show create undead animation (cmbtfx.lbx 27) if there are new undead units

        defeatedDefenders = combatScreen.Model.DefeatedDefenders
        defeatedAttackers = combatScreen.Model.DefeatedAttackers

        // FIXME: resolve the attacker/defender stack at the end of combat?
        for _, unit := range combatScreen.Model.AttackingArmy.RecalledUnits {
            recalledAttackers = append(recalledAttackers, unit.Unit.(units.StackUnit))
        }
        for _, unit := range combatScreen.Model.DefendingArmy.RecalledUnits {
            recalledDefenders = append(recalledDefenders, unit.Unit.(units.StackUnit))
        }
    }

    // returns the fame that should be added to the winner and loser. the loser fame is negative
    distributeFame := func(winner *playerlib.Player, loser *playerlib.Player, loserStack *playerlib.UnitStack, defeatedUnits int) (int, int) {
        winnerFame := 0
        loserFame := 0

        if defeatedUnits >= 4 {
            winnerFame += 1
            loserFame -= 1
        }

        for _, unit := range loserStack.Units() {
            if unit.GetRawUnit().CastingCost >= 600 {
                winnerFame += 1
                loserFame -= 1
                break
            }
            if unit.IsHero() {
                hero := unit.(*herolib.Hero)
                loserFame -= (int(hero.GetExperienceLevel()) + 1) / 2
            }
        }

        return winnerFame, loserFame
    }

    // fame
    var attackerFame, defenderFame int

    if state == combat.CombatStateAttackerWin || state == combat.CombatStateDefenderFlee {
        if zone.City != nil {
            razeCity, gold := game.defeatCity(yield, attacker, attackerStack, defender, zone.City)
            // if razeCity is true then we pass in false to get the fame for capturing the city
            attackerFame += zone.City.FameForCaptureOrRaze(!razeCity)
            defenderFame += zone.City.FameForCaptureOrRaze(false)

            attacker.Gold += gold
            defender.Gold -= gold
        }

        winner, loser := distributeFame(attacker, defender, defenderStack, defeatedDefenders)
        attackerFame += winner
        defenderFame += loser

        // some units might have been raised from the dead but been owned by the opposite side
        for _, unit := range attackingArmy.GetUnits() {
            if unit.Unit.GetHealth() > 0 {
                attackerStack.AddUnit(unit.Unit)
                // FIXME: its annoying to have to set all 3 of these fields. maybe consolidate them into just
                // SetOwner(player)
                unit.Unit.SetBanner(attacker.GetBanner())
                unit.Unit.SetGlobalEnchantmentProvider(attacker.MakeUnitEnchantmentProvider())
                unit.Unit.SetExperienceInfo(attacker.MakeExperienceInfo())
                defender.RemoveUnit(unit.Unit)
            }
        }

    } else if state == combat.CombatStateDefenderWin || state == combat.CombatStateAttackerFlee {
        winner, loser := distributeFame(defender, attacker, attackerStack, defeatedAttackers)
        defenderFame += winner
        attackerFame += loser

        for _, unit := range defendingArmy.GetUnits() {
            if unit.Unit.GetHealth() > 0 {
                defenderStack.AddUnit(unit.Unit)
                unit.Unit.SetBanner(defender.GetBanner())
                unit.Unit.SetGlobalEnchantmentProvider(defender.MakeUnitEnchantmentProvider())
                unit.Unit.SetExperienceInfo(defender.MakeExperienceInfo())
                attacker.RemoveUnit(unit.Unit)
            }
        }
    }

    attacker.Fame = max(0, attacker.Fame + attackerFame)
    defender.Fame = max(0, defender.Fame + defenderFame)

    cityPopulationLoss := 0
    var cityBuildingLoss []buildinglib.Building

    if zone.City != nil && state == combat.CombatStateAttackerWin && !strategicCombat {
        // maximum chance is 50%, minimum is 10%
        chance := min(50, 10 + combatScreen.Model.CollateralDamage * 2)
        for range zone.City.Citizens() - 1 {
            if rand.N(100) < chance {
                cityPopulationLoss += 1
            }
        }

        minBuildingChance := 10
        if attacker.GetBanner() == data.BannerBrown {
            minBuildingChance = 50
        }

        chance = min(75, minBuildingChance + combatScreen.Model.CollateralDamage)
        for _, building := range zone.City.Buildings.Values() {
            if building == buildinglib.BuildingFortress || building == buildinglib.BuildingSummoningCircle {
                continue
            }

            if rand.N(100) < chance {
                cityBuildingLoss = append(cityBuildingLoss, building)
            }
        }

        zone.City.Population -= cityPopulationLoss * 1000
        for _, building := range cityBuildingLoss {
            zone.City.Buildings.Remove(building)
        }
        zone.City.ResetCitizens()
    }

    // Show end screen
    if !strategicCombat {
        result := combat.CombatEndScreenResultLose
        humanAttacker := attacker.IsHuman()
        fame := defenderFame

        if state == combat.CombatStateAttackerWin || state == combat.CombatStateDefenderFlee {
            fame = attackerFame
        }

        switch {
            case state == combat.CombatStateAttackerWin && humanAttacker,
                 state == combat.CombatStateDefenderWin && !humanAttacker:
                result = combat.CombatEndScreenResultWin
            case state == combat.CombatStateAttackerFlee && humanAttacker,
                 state == combat.CombatStateDefenderFlee && !humanAttacker:
                result = combat.CombatEndScreenResultRetreat
        }

        // FIXME: show how much gold was plundered (or lost)
        endScreen := combat.MakeCombatEndScreen(game.Cache, combatScreen, result, combatScreen.Model.DiedWhileFleeing, fame, cityPopulationLoss, len(cityBuildingLoss))
        game.Drawer = func (screen *ebiten.Image, game *Game){
            endScreen.Draw(screen)
        }

        state2 := combat.CombatEndScreenRunning
        for state2 == combat.CombatEndScreenRunning {
            state2 = endScreen.Update()
            yield()
        }

        game.Drawer = oldDrawer
    }

    // Redistribute equipment of died heros
    showHeroNotice := false

    distributeEquipment := func (player *playerlib.Player, hero *herolib.Hero){
        for _, item := range hero.Equipment {
            if item != nil {
                showHeroNotice = true
                select {
                    case game.Events <- &GameEventVault{CreatedArtifact: item, Player: player}:
                    default:
                }
            }
        }
    }

    // recall units
    relocateUnits := func(player *playerlib.Player, units []units.StackUnit) {
        for _, unit := range units {
            game.RelocateUnit(player, unit)
        }
    }

    relocateUnits(attacker, recalledAttackers)
    relocateUnits(defender, recalledDefenders)

    // remove dead units
    killUnits := func (player *playerlib.Player, stack *playerlib.UnitStack, landscape combat.CombatLandscape){
        // first remove sailing units
        for _, unit := range stack.Units() {
            if unit.IsSailing() && unit.GetHealth() <= 0 {
                player.RemoveUnit(unit)
            }
        }

        transport := stack.HasSailingUnits(false)

        for _, unit := range stack.Units() {
            dead := unit.GetHealth() <= 0

            // if combat was on water and there are no sailing ships left then all units should die
            // FIXME: handle the case that there were originally two ships and one died, thus not being able to transport some units
            if landscape == combat.CombatLandscapeWater && unit.IsLandWalker() && !transport {
                dead = true
            }

            if dead {
                player.RemoveUnit(unit)

                if unit.IsHero() {
                    hero := unit.(*herolib.Hero)
                    if player.IsHuman() {
                        distributeEquipment(player, hero)
                    }
                    // FIXME: what happens with the equipment in case of non-human players?
                    for index := range hero.Equipment {
                        hero.Equipment[index] = nil
                    }
                }
            }
        }
    }

    killUnits(attacker, attackerStack, landscape)
    killUnits(defender, defenderStack, landscape)

    if !strategicCombat {
        switch state {
            case combat.CombatStateAttackerWin, combat.CombatStateDefenderFlee:
                for _, unit := range combatScreen.Model.UndeadUnits {
                    defender.RemoveUnit(unit.Unit)
                    if len(attackerStack.Units()) < data.MaxUnitsInStack {
                        attacker.AddUnit(attacker.UpdateUnit(unit.Unit))
                    }
                }

                if attacker.HasEnchantment(data.EnchantmentZombieMastery) {
                    for _, unit := range append(slices.Clone(defendingArmy.KilledUnits), attackingArmy.KilledUnits...) {
                        if unit.GetRace() != data.RaceFantastic && unit.GetRace() != data.RaceHero && len(attackerStack.Units()) < data.MaxUnitsInStack {
                            attacker.AddUnit(units.MakeOverworldUnitFromUnit(units.Zombie, attackerStack.X(), attackerStack.Y(), attackerStack.Plane(), attacker.GetBanner(), attacker.MakeExperienceInfo(), attacker.MakeUnitEnchantmentProvider()))
                        }
                    }
                }

            case combat.CombatStateDefenderWin, combat.CombatStateAttackerFlee:
                for _, unit := range combatScreen.Model.UndeadUnits {
                    attacker.RemoveUnit(unit.Unit)
                    if len(defenderStack.Units()) < data.MaxUnitsInStack {
                        defender.AddUnit(defender.UpdateUnit(unit.Unit))
                    }
                }

                if defender.HasEnchantment(data.EnchantmentZombieMastery) {
                    for _, unit := range append(slices.Clone(defendingArmy.KilledUnits), attackingArmy.KilledUnits...) {
                        if unit.GetRace() != data.RaceFantastic && unit.GetRace() != data.RaceHero && len(defenderStack.Units()) < data.MaxUnitsInStack {
                            defender.AddUnit(units.MakeOverworldUnitFromUnit(units.Zombie, defenderStack.X(), defenderStack.Y(), defenderStack.Plane(), defender.GetBanner(), defender.MakeExperienceInfo(), defender.MakeUnitEnchantmentProvider()))
                        }
                    }
                }
        }
    }

    // experience
    if state == combat.CombatStateAttackerWin || state == combat.CombatStateDefenderFlee {
        for _, unit := range attackerStack.Units() {
            if unit.GetRace() != data.RaceFantastic {
                game.AddExperience(attacker, unit, defeatedDefenders * 2)
            }
        }
    } else if state == combat.CombatStateDefenderWin || state == combat.CombatStateAttackerFlee {
        for _, unit := range defenderStack.Units() {
            if unit.GetRace() != data.RaceFantastic {
                game.AddExperience(defender, unit, defeatedAttackers * 2)
            }
        }
    }

    if showHeroNotice {
        game.doNotice(yield, game.HudUI, "One or more heroes died in combat. You must redistribute their equipment.")
    }

    return state
}

// FIXME: maybe this should be generalized so it can handle any enchantment that cuases a spell to fizzle?
func (game *Game) ShowTranquilityFizzle(tranquilityOwner *playerlib.Player, caster *playerlib.Player, spell spellbook.Spell) {
    group, quit := game.makeTranquilityFizzleUI(tranquilityOwner, caster, spell)

    game.Events <- &GameEventRunUI{
        Group: group,
        Quit: quit,
        Song: music.SongSupressMagicActivating,
    }
}

func (game *Game) GetWizardAnimal(wizard setup.WizardCustom) *ebiten.Image {
    // devil: 51
    // cat: 52
    // bird: 53
    // snake: 54
    // beetle: 55
    animalIndex := 51
    switch wizard.MostBooks() {
        case data.NatureMagic: animalIndex = 54
        case data.SorceryMagic: animalIndex = 55
        case data.ChaosMagic: animalIndex = 51
        case data.LifeMagic: animalIndex = 53
        case data.DeathMagic: animalIndex = 52
    }
    animal, _ := game.ImageCache.GetImageTransform("resource.lbx", animalIndex, 0, "crop", util.AutoCrop)
    return animal
}

type FizzleFonts struct {
    Font *font.Font
}

func MakeFizzleFonts(cache *lbx.LbxCache) *FizzleFonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Error loading fizzle fonts: %v", err)
        return nil
    }

    return &FizzleFonts{
        Font: loader(fontslib.LightFont),
    }
}

// the spell was fizzled by the tranquility spell. show the fizzle picture of a broken wand
func (game *Game) makeTranquilityFizzleUI(tranquilityOwner *playerlib.Player, caster *playerlib.Player, spell spellbook.Spell) (*uilib.UIElementGroup, context.Context) {
    quit, cancel := context.WithCancel(context.Background())

    group := uilib.MakeGroup()

    rotateIndexLow := 244
    rotateIndexHigh := 254
    specfxLbx, _ := game.Cache.GetLbxFile("specfx.lbx")
    wandAnimation := util.MakePaletteRotateAnimation(specfxLbx, 50, rotateIndexLow, rotateIndexHigh)

    fader := group.MakeFadeIn(7)

    fonts := MakeFizzleFonts(game.Cache)

    clicked := false
    shutdown := func(){
        if !clicked {
            clicked = true
            fader = group.MakeFadeOut(7)
            group.AddDelay(7, func() {
                cancel()
            })
        }
    }

    uiX := 10
    uiY := 70

    group.Update = func() {
        if group.Counter % 3 == 0 {
            wandAnimation.Next()
        }
    }

    animal := game.GetWizardAnimal(tranquilityOwner.Wizard)

    left, _ := game.ImageCache.GetImage("resource.lbx", 43, 0)
    right, _ := game.ImageCache.GetImage("resource.lbx", 44, 0)

    owner := ""
    if tranquilityOwner == game.Players[0] {
        owner = "Your"
    } else {
        owner = tranquilityOwner.Wizard.Name + "'s"
    }

    textOffset := 70

    wrappedText := fonts.Font.CreateWrappedText(float64(left.Bounds().Dx() - textOffset), 1, fmt.Sprintf("%v Tranquility spell has caused %v's %v spell to fizzle.", owner, caster.Wizard.Name, spell.Name))

    rect := image.Rect(0, 0, left.Bounds().Dx() + right.Bounds().Dx(), left.Bounds().Dy()).Add(image.Pt(uiX, uiY))
    group.AddElement(&uilib.UIElement{
        Rect: rect,
        Draw: func (element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(fader())
            options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
            scale.DrawScaled(screen, left, &options)

            saveGeom := options.GeoM
            options.GeoM.Translate(35, 30)
            options.GeoM.Translate(float64(-animal.Bounds().Dx()/2), float64(-animal.Bounds().Dy()/2))
            scale.DrawScaled(screen, animal, &options)

            options.GeoM = saveGeom

            options.GeoM.Translate(float64(left.Bounds().Dx()), 0)
            scale.DrawScaled(screen, right, &options)

            options.GeoM.Translate(5, 7)

            scale.DrawScaled(screen, wandAnimation.Frame(), &options)

            fonts.Font.RenderWrapped(screen, float64(rect.Min.X + textOffset), float64(rect.Min.Y + 12), wrappedText, font.FontOptions{Scale: scale.ScaleAmount, DropShadow: true, Options: &options})
        },
        LeftClick: func (element *uilib.UIElement){
            shutdown()
        },
    })

    return group, quit
}

func GetCityImage(city *citylib.City, cache *util.ImageCache) (*ebiten.Image, error) {
    var spriteIndex int = 21
    var animationIndex int = 0

    if city.HasWall() {
        spriteIndex = 20
    }

    switch city.GetSize() {
        case citylib.CitySizeHamlet: animationIndex = 0
        case citylib.CitySizeVillage: animationIndex = 1
        case citylib.CitySizeTown: animationIndex = 2
        case citylib.CitySizeCity: animationIndex = 3
        case citylib.CitySizeCapital: animationIndex = 4
    }

    // the city image is a sub-frame of animation 20
    // return cache.GetImageTransform("mapback.lbx", 20, index, city.GetBanner().String(), util.ComposeImageTransform(units.MakeUpdateUnitColorsFunc(city.GetBanner()), util.AutoCropGeneric))
    return cache.GetImageTransform("mapback.lbx", spriteIndex, animationIndex, city.GetBanner().String(), units.MakeUpdateUnitColorsFunc(city.GetBanner()))
}

func (game *Game) ShowGrandVizierUI(){
    group := uilib.MakeGroup()

    yes := func(){
        // FIXME: enable grand vizier
        game.HudUI.RemoveGroup(group)
    }

    no := func(){
        // FIXME: disable grand vizier
        game.HudUI.RemoveGroup(group)
    }

    group.AddElements(uilib.MakeConfirmDialogWithLayer(group, game.Cache, &game.ImageCache, 1, "Do you wish to allow the Grand Vizier to select what buildings your cities create?", true, yes, no))
    game.HudUI.AddGroup(group)
}

func (game *Game) ShowTaxCollectorUI(cornerX int, cornerY int){
    player := game.Players[0]

    // put a * on the value that is currently selected
    selected := func(s string, use bool) string {
        if use {
            return fmt.Sprintf("%v*", s)
        }

        return s
    }

    update := func(rate fraction.Fraction){
        player.UpdateTaxRate(rate)
        game.RefreshUI()
    }

    taxes := []uilib.Selection{
        uilib.Selection{
            Name: selected("0 gold, 0% unrest", player.TaxRate.IsZero()),
            Action: func(){
                update(fraction.Zero())
            },
        },
        uilib.Selection{
            Name: selected("0.5 gold, 10% unrest", player.TaxRate.Equals(fraction.Make(1, 2))),
            Action: func(){
                update(fraction.Make(1, 2))
            },
        },
        uilib.Selection{
            Name: selected("1 gold, 20% unrest", player.TaxRate.Equals(fraction.Make(1, 1))),
            Action: func(){
                update(fraction.Make(1, 1))
            },
        },
        uilib.Selection{
            Name: selected("1.5 gold, 30% unrest", player.TaxRate.Equals(fraction.Make(3, 2))),
            Action: func(){
                update(fraction.Make(3, 2))
            },
        },
        uilib.Selection{
            Name: selected("2 gold, 45% unrest", player.TaxRate.Equals(fraction.Make(2, 1))),
            Action: func(){
                update(fraction.Make(2, 1))
            },
        },
        uilib.Selection{
            Name: selected("2.5 gold, 60% unrest", player.TaxRate.Equals(fraction.Make(5, 2))),
            Action: func(){
                update(fraction.Make(5, 2))
            },
        },
        uilib.Selection{
            Name: selected("3 gold, 75% unrest", player.TaxRate.Equals(fraction.Make(3, 1))),
            Action: func(){
                update(fraction.Make(3, 1))
            },
        },
    }

    game.HudUI.AddElements(uilib.MakeSelectionUI(game.HudUI, game.Cache, &game.ImageCache, cornerX, cornerY, "Tax Per Population", taxes, true))
}

func (game *Game) ShowApprenticeUI(yield coroutine.YieldFunc, player *playerlib.Player){
    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    newDrawer := func (screen *ebiten.Image){
    }

    game.Drawer = func (screen *ebiten.Image, game *Game){
        newDrawer(screen)
    }

    power := game.ComputePower(player)
    spellbook.ShowSpellBook(yield, game.Cache, player.ResearchPoolSpells, player.KnownSpells, player.ResearchCandidateSpells, player.ResearchingSpell, player.ResearchProgress, player.SpellResearchPerTurn(power), player.ComputeOverworldCastingSkill(), spellbook.Spell{}, false, nil, player, &newDrawer)
}

func (game *Game) ResearchNewSpell(yield coroutine.YieldFunc, player *playerlib.Player){
    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    newDrawer := func (screen *ebiten.Image){
    }

    game.Drawer = func (screen *ebiten.Image, game *Game){
        newDrawer(screen)
    }

    if len(player.ResearchCandidateSpells.Spells) > 0 {
        power := game.ComputePower(player)
        spellbook.ShowSpellBook(yield, game.Cache, player.ResearchPoolSpells, player.KnownSpells, player.ResearchCandidateSpells, spellbook.Spell{}, 0, player.SpellResearchPerTurn(power), player.ComputeOverworldCastingSkill(), spellbook.Spell{}, true, &player.ResearchingSpell, player, &newDrawer)
    }
}

// show all scroll events for this turn, or a message that no events occurred
func (game *Game) DoChancellor(){
    if len(game.ScrollEvents) == 0 {
        event := &GameEventScroll{
            Title: "NO EVENTS THIS MONTH",
        }
        select {
            case game.Events <- event:
            default:
        }
    } else {
        for _, event := range game.ScrollEvents {
            event.Old = true
            select {
                case game.Events <- event:
                default:
            }
        }
    }
}

func (game *Game) ShowMirror() {
    if len(game.Players) > 0 {
        game.HudUI.AddElement(mirror.MakeMirrorUI(game.Cache, game.Players[0], game.HudUI))
    }
}

// advisor ui
func (game *Game) MakeInfoUI(cornerX int, cornerY int) []*uilib.UIElement {
    advisors := []uilib.Selection{
        uilib.Selection{
            Name: "Surveyor",
            Action: func(){
                select {
                    case game.Events<- &GameEventSurveyor{}:
                    default:
                }
            },
            Hotkey: "(F1)",
        },
        uilib.Selection{
            Name: "Cartographer",
            Action: func(){
                select {
                    case game.Events<- &GameEventCartographer{}:
                    default:
                }
            },
            Hotkey: "(F2)",
        },
        uilib.Selection{
            Name: "Apprentice",
            Action: func(){
                select {
                    case game.Events<- &GameEventApprenticeUI{}:
                    default:
                }
            },
            Hotkey: "(F3)",
        },
        uilib.Selection{
            Name: "Historian",
            Action: func(){
                select {
                    case game.Events<- &GameEventHistorian{}:
                    default:
                }
            },
            Hotkey: "(F4)",
        },
        uilib.Selection{
            Name: "Astrologer",
            Action: func(){
                select {
                    case game.Events<- &GameEventAstrologer{}:
                    default:
                }
            },
            Hotkey: "(F5)",
        },
        uilib.Selection{
            Name: "Chancellor",
            Action: func(){
                game.DoChancellor()
            },
            Hotkey: "(F6)",
        },
        uilib.Selection{
            Name: "Tax Collector",
            Action: func(){
                game.ShowTaxCollectorUI(cornerX - 10, cornerY + 10)
            },
            Hotkey: "(F7)",
        },
        uilib.Selection{
            Name: "Grand Vizier",
            Action: func(){
                game.ShowGrandVizierUI()
            },
            Hotkey: "(F8)",
        },
        uilib.Selection{
            Name: "Mirror",
            Action: func(){
                game.ShowMirror()
            },
            Hotkey: "(F9)",
        },
    }

    return uilib.MakeSelectionUI(game.HudUI, game.Cache, &game.ImageCache, cornerX, cornerY, "Select An Advisor", advisors, true)
}

func (game *Game) ShowSpellBookCastUI(yield coroutine.YieldFunc, player *playerlib.Player){
    overlandSpells := player.KnownSpells.OverlandSpells()
    // don't show this spell in the spellbook, it will be cast automatically
    overlandSpells.RemoveSpell(overlandSpells.FindByName("Spell of Return"))

    castPlayer := CastPlayer{
        player: player,
        remainingCastingSkill: player.RemainingCastingSkill,
        castingSkill: player.ComputeOverworldCastingSkill(),
        manaPerTurn: player.ManaPerTurn(game.ComputePower(player), game),
        mana: player.Mana,
    }

    game.HudUI.AddElements(spellbook.MakeSpellBookCastUI(game.HudUI, game.Cache, overlandSpells, make(map[spellbook.Spell]int), player.ComputeOverworldCastingSkill(), player.CastingSpell, player.CastingSpellProgress, true, &castPlayer, &player.CastingSpellPage, func (spell spellbook.Spell, picked bool){
        if picked {
            if spell.Name == "Create Artifact" || spell.Name == "Enchant Item" {

                drawFunc := func(screen *ebiten.Image){}
                oldDrawer := game.Drawer
                defer func(){
                    game.Drawer = oldDrawer
                }()
                game.Drawer = func(screen *ebiten.Image, game *Game){
                    drawFunc(screen)
                }

                creation := artifact.CreationCreateArtifact
                switch spell.Name {
                    case "Create Artifact": creation = artifact.CreationCreateArtifact
                    case "Enchant Item": creation = artifact.CreationEnchantItem
                }

                created, cancel := artifact.ShowCreateArtifactScreen(yield, game.Cache, creation, &player.Wizard, player.KnownSpells.CombatSpells(false), &drawFunc)
                if cancel {
                    return
                }

                log.Printf("Create artifact %v", created)
                spell.OverrideCost = created.Cost

                player.CreateArtifact = created
            } else if spell.Name == "Spell of Mastery" {
                // show an animation that the spell of mastery is being cast first
                game.Music.PushSong(music.SongSpellOfMastery)
                logic, draw := mastery.ShowSpellOfMasteryScreen(game.Cache, player.Wizard.Name)

                oldDrawer := game.Drawer
                game.Drawer = func(screen *ebiten.Image, game *Game){
                    draw(screen)
                }

                logic(yield)

                game.Drawer = oldDrawer
                game.Music.PopSong()
            }

            castingCost := player.ComputeEffectiveSpellCost(spell, true)

            if spell.Valid() && castingCost <= player.Mana && castingCost <= player.RemainingCastingSkill {
                player.Mana -= castingCost
                player.RemainingCastingSkill -= castingCost

                game.doCastSpell(player, spell)
            } else {
                player.CastingSpell = spell
            }
        }
    }))
}

func (game *Game) ComputeMaximumPopulation(x int, y int, plane data.Plane) int {
    // find catchment area of x, y
    // for each square, compute food production
    // maximum pop is food production
    maybeCity, _ := game.FindCity(x, y, plane)
    if maybeCity != nil {
        return maybeCity.MaximumCitySize()
    }

    mapUse := game.GetMap(plane)
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

func (game *Game) CityGoldBonus(x int, y int, plane data.Plane) int {
    mapObject := game.GetMap(plane)
    tile := mapObject.GetTile(x, y)
    return tile.GoldBonus(mapObject)
}

func (game *Game) CityProductionBonus(x int, y int, plane data.Plane) int {
    mapUse := game.GetMap(plane)
    catchment := mapUse.GetCatchmentArea(x, y)

    production := 0

    for _, tile := range catchment {
        production += tile.ProductionBonus(false)
    }

    return production
}

// FIXME: cache the spells
func (game *Game) GetSpellByName(name string) spellbook.Spell {
    spells, err := spellbook.ReadSpellsFromCache(game.Cache)
    if err != nil {
        return spellbook.Spell{}
    }

    return spells.FindByName(name)
}

func (game *Game) CreateOutpost(settlers units.StackUnit, player *playerlib.Player) *citylib.City {
    cityName := game.SuggestCityName(settlers.GetRace())

    newCity := citylib.MakeCity(cityName, settlers.GetX(), settlers.GetY(), settlers.GetRace(), game.BuildingInfo, game.GetMap(settlers.GetPlane()), game, player)
    newCity.Plane = settlers.GetPlane()
    newCity.Population = 300
    newCity.Outpost = true
    newCity.ProducingBuilding = buildinglib.BuildingHousing
    newCity.ProducingUnit = units.UnitNone

    player.RemoveUnit(settlers)
    player.SelectedStack = nil
    game.RefreshUI()
    player.AddCity(newCity)

    stack := player.FindStack(newCity.X, newCity.Y, newCity.Plane)

    select {
        case game.Events<- &GameEventNewOutpost{City: newCity, Stack: stack, Player: player}:
        default:
    }

    return newCity
}

func (game *Game) DoMeld(unit units.StackUnit, player *playerlib.Player, node *maplib.ExtraMagicNode){
    node.Meld(player, unit.GetRawUnit())
    player.RemoveUnit(unit)
}

func (game *Game) DoBuildAction(player *playerlib.Player){
    if player.SelectedStack != nil {
        var powers UnitBuildPowers

        if player.SelectedStack != nil {
            powers = computeUnitBuildPowers(player.SelectedStack)
        }

        if powers.CreateOutpost {
            // search for the settlers (the only unit with the create outpost ability
            for _, settlers := range player.SelectedStack.ActiveUnits() {
                if game.IsSettlableLocation(settlers.GetX(), settlers.GetY(), settlers.GetPlane()) && settlers.HasAbility(data.AbilityCreateOutpost) {
                    game.CreateOutpost(settlers, player)
                    game.RefreshUI()
                    break
                }
            }
        } else if powers.Meld {
            node := game.GetMap(player.SelectedStack.Plane()).GetMagicNode(player.SelectedStack.X(), player.SelectedStack.Y())
            for _, melder := range player.SelectedStack.ActiveUnits() {
                if melder.HasAbility(data.AbilityMeld) && !node.Warped {
                    game.DoMeld(melder, player, node)
                    game.RefreshUI()
                    break
                }
            }
        } else if powers.Purify {

            for _, unit := range player.SelectedStack.ActiveUnits() {
                if unit.HasAbility(data.AbilityPurify) {
                    unit.SetBusy(units.BusyStatusPurify)
                    unit.SetMovesLeft(true, fraction.Zero())
                }
            }

            player.SelectedStack.EnableMovers()

            // player.SelectedStack.ExhaustMoves()
            game.RefreshUI()

        } else if powers.BuildRoad {

            for _, unit := range player.SelectedStack.ActiveUnits() {
                if unit.HasAbility(data.AbilityConstruction) {
                    unit.SetBusy(units.BusyStatusBuildRoad)
                    unit.SetMovesLeft(true, fraction.Zero())
                }
            }

            player.SelectedStack.EnableMovers()

            // player.SelectedStack.ExhaustMoves()
            game.RefreshUI()
        }
    }
}

// find all engineers that are currently building a road
// compute the work done by each engineer according to the terrain
//   total work = work per engineer ^ engineers building on that tile
// add total work to some counter, and when that total reaches the threshold for the terrain type
// then set a road on that tile and make the engineers no longer busy
func (game *Game) DoBuildRoads(player *playerlib.Player) {
    type RoadWork struct {
        WorkPerEngineer float64
        TotalWork float64
    }

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

    arcanusBuilds := make(map[image.Point]struct{})
    myrrorBuilds := make(map[image.Point]struct{})

    for _, stack := range player.Stacks {
        plane := stack.Plane()

        engineerCount := 0
        for _, unit := range stack.Units() {
            if unit.GetBusy() == units.BusyStatusBuildRoad {
                engineerCount += 1

                if unit.GetRace() == data.RaceDwarf {
                    engineerCount += 1
                }

                if unit.HasEnchantment(data.UnitEnchantmentEndurance) {
                    engineerCount += 1
                }
            }
        }

        if engineerCount > 0 {
            x, y := stack.X(), stack.Y()
            // log.Printf("building a road at %v, %v with %v engineers", x, y, engineerCount)
            roads := game.RoadWorkArcanus
            if plane == data.PlaneMyrror {
                roads = game.RoadWorkMyrror
            }

            amount, ok := roads[image.Pt(x, y)]
            if !ok {
                amount = 0
            }

            tileWork := work[game.GetMap(plane).GetTile(x, y).Tile.TerrainType()]

            amount += math.Pow(tileWork.WorkPerEngineer, float64(engineerCount))
            // log.Printf("  amount is now %v. total work is %v", amount, tileWork.TotalWork)
            if amount >= tileWork.TotalWork {
                game.GetMap(plane).SetRoad(x, y, plane == data.PlaneMyrror)

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
    for point, _ := range game.RoadWorkArcanus {
        _, ok := arcanusBuilds[point]
        if !ok {
            toDelete = append(toDelete, point)
        }
    }

    for _, point := range toDelete {
        // log.Printf("remove point %v", point)
        delete(game.RoadWorkArcanus, point)
    }

    toDelete = nil
    for point, _ := range game.RoadWorkMyrror {
        _, ok := myrrorBuilds[point]
        if !ok {
            toDelete = append(toDelete, point)
        }
    }

    for _, point := range toDelete {
        // log.Printf("remove point %v", point)
        delete(game.RoadWorkMyrror, point)
    }

}

func (game *Game) DoPurify(player *playerlib.Player) {
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
            purify := game.PurifyWorkArcanus
            if plane == data.PlaneMyrror {
                purify = game.PurifyWorkMyrror
            }

            amount, ok := purify[image.Pt(x, y)]
            if !ok {
                amount = 0
            }

            amount += math.Pow(work.WorkPerUnit, float64(unitCount))
            // log.Printf("  amount is now %v. total work is %v", amount, tileWork.TotalWork)
            if amount >= work.TotalWork {
                game.GetMap(plane).RemoveCorruption(x, y)

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
    for point, _ := range game.PurifyWorkArcanus {
        _, ok := arcanusBuilds[point]
        if !ok {
            toDelete = append(toDelete, point)
        }
    }

    for _, point := range toDelete {
        // log.Printf("remove point %v", point)
        delete(game.PurifyWorkArcanus, point)
    }

    toDelete = nil
    for point, _ := range game.PurifyWorkMyrror {
        _, ok := myrrorBuilds[point]
        if !ok {
            toDelete = append(toDelete, point)
        }
    }

    for _, point := range toDelete {
        // log.Printf("remove point %v", point)
        delete(game.PurifyWorkMyrror, point)
    }
}

func (game *Game) IsGlobalEnchantmentActive(enchantment data.Enchantment) bool {
    return slices.ContainsFunc(game.Players, func (player *playerlib.Player) bool {
        return player.GlobalEnchantments.Contains(enchantment)
    })
}

func (game *Game) ComputeCityStackInfo() playerlib.CityStackInfo {
    out := playerlib.CityStackInfo{
        ArcanusStacks: make(map[image.Point]*playerlib.UnitStack),
        MyrrorStacks: make(map[image.Point]*playerlib.UnitStack),
        ArcanusCities: make(map[image.Point]*citylib.City),
        MyrrorCities: make(map[image.Point]*citylib.City),
    }

    for _, player := range game.Players {
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

func (game *Game) GetHumanPlayer() *playerlib.Player {
    return game.Players[0]
}

func (game *Game) PlaneShift(stack *playerlib.UnitStack, player *playerlib.Player) error {
    if game.IsGlobalEnchantmentActive(data.EnchantmentPlanarSeal) {
        return errors.New("Cannot plane shift while the Planar Seal is in effect")
    }

    canMove := false

    cityOppositePlane := player.FindCity(stack.X(), stack.Y(), stack.Plane().Opposite())

    mapPlane := game.GetMap(stack.Plane().Opposite())

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

    cityStackInfo := game.ComputeCityStackInfo()

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
    }

    return nil
}

func (game *Game) doPlanarTraval() {
    // no switching planes if the global enchantment planar seal is in effect
    if !game.IsGlobalEnchantmentActive(data.EnchantmentPlanarSeal) {
        player := game.Players[0]

        stack := player.SelectedStack

        if stack != nil && stack.Plane() == game.Plane.Opposite() {
            activeStack := player.SplitActiveStack(stack)
            // nothing to do
            if len(activeStack.Units()) == 0 {
                return
            }

            travelEnabled := false

            if game.CurrentMap().HasOpenTower(activeStack.X(), activeStack.Y()) {
                travelEnabled = true
            } else {
                cityThisPlane := player.FindCity(activeStack.X(), activeStack.Y(), activeStack.Plane())
                cityOppositePlane := player.FindCity(activeStack.X(), activeStack.Y(), activeStack.Plane().Opposite())
                hasAstralGate := (cityThisPlane != nil && cityThisPlane.HasEnchantment(data.CityEnchantmentAstralGate)) ||
                                 (cityOppositePlane != nil && cityOppositePlane.HasEnchantment(data.CityEnchantmentAstralGate))

                hasPlanarTravel := activeStack.ActiveUnitsHasAbility(data.AbilityPlaneShift)

                if hasAstralGate || hasPlanarTravel {
                    travelEnabled = true
                }
            }

            moved := false
            if travelEnabled {
                err := game.PlaneShift(activeStack, player)
                if err != nil {
                    select {
                        case game.Events<- &GameEventNotice{Message: fmt.Sprintf("The selected units cannot planar travel at this location.")}:
                        default:
                    }
                } else {
                    moved = true
                    player.SelectedStack = stack
                    player.UpdateFogVisibility()
                }
            }

            if !moved {
                if activeStack != stack {
                    player.MergeStacks(stack, activeStack)
                }
            }
        }
    }
}

func (game *Game) SwitchPlane() {
    switch game.Plane {
        case data.PlaneArcanus: game.Plane = data.PlaneMyrror
        case data.PlaneMyrror: game.Plane = data.PlaneArcanus
    }
}

func (game *Game) MakeHudUI() *uilib.UI {
    ui := &uilib.UI{
        Cache: game.Cache,
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            mainHud, _ := game.ImageCache.GetImage("main.lbx", 0, 0)
            scale.DrawScaled(screen, mainHud, &options)

            ui.StandardDraw(screen)
        },
        HandleKeys: func(keys []ebiten.Key){
            player := game.Players[game.CurrentPlayer]
            if player.IsHuman() {
                if game.HudUI.GetHighestLayerValue() == 0 {
                    for _, key := range keys {
                        switch key {
                            case ebiten.KeySpace:
                                stack := game.Players[0].SelectedStack

                                if stack != nil {
                                    select {
                                        case game.Events <- &GameEventMoveCamera{Plane: stack.Plane(), X: stack.X(), Y: stack.Y()}:
                                        default:
                                    }
                                }
                            case ebiten.KeyN:
                                stack := game.Players[0].SelectedStack

                                if stack == nil || stack.OutOfMoves() {
                                    select {
                                        case game.Events <- &GameEventNextTurn{}:
                                        default:
                                    }
                                }
                            case ebiten.KeyF1:
                                select {
                                    case game.Events<- &GameEventSurveyor{}:
                                    default:
                                }
                            case ebiten.KeyF2:
                                select {
                                    case game.Events<- &GameEventCartographer{}:
                                    default:
                                }
                            case ebiten.KeyF3:
                                select {
                                    case game.Events<- &GameEventApprenticeUI{}:
                                    default:
                                }
                            case ebiten.KeyF4:
                                select {
                                    case game.Events<- &GameEventHistorian{}:
                                    default:
                                }
                            case ebiten.KeyF5:
                                select {
                                    case game.Events<- &GameEventAstrologer{}:
                                    default:
                                }
                            case ebiten.KeyF6:
                                game.DoChancellor()
                            case ebiten.KeyF7:
                                cornerX := 60
                                cornerY := 25
                                game.ShowTaxCollectorUI(cornerX - 10, cornerY + 10)
                            case ebiten.KeyF8:
                                game.ShowGrandVizierUI()
                            case ebiten.KeyF9:
                                game.ShowMirror()
                        }
                    }
                }
            }
        },
    }

    group := uilib.MakeGroup()
    ui.AddGroup(group)

    // onClick - true to perform the action when the left click occurs, false to perform the action when the left click is released
    makeButton := func(lbxIndex int, x int, y int, onClick bool, action func()) *uilib.UIElement {
        buttons, _ := game.ImageCache.GetImages("main.lbx", lbxIndex)
        rect := image.Rect(x, y, x + buttons[0].Bounds().Dx(), y + buttons[0].Bounds().Dy())
        index := 0
        counter := uint64(0)
        return &uilib.UIElement{
            Rect: rect,
            PlaySoundLeftClick: true,
            Inside: func(this *uilib.UIElement, x int, y int){
                counter += 1
            },
            NotInside: func(this *uilib.UIElement){
                counter = 0
            },
            LeftClick: func(this *uilib.UIElement){
                index = 1
                if onClick {
                    action()
                }
            },
            LeftClickRelease: func(this *uilib.UIElement){
                index = 0
                if !onClick {
                    action()
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                colorScale := ebiten.ColorScale{}

                if counter > 0 {
                    v := float32(1 + (math.Sin(float64(counter / 4)) / 2 + 0.5) / 2)
                    colorScale.Scale(v, v, v, 1)
                }

                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
                options.ColorScale.ScaleWithColorScale(colorScale)
                scale.DrawScaled(screen, buttons[index], &options)
            },
        }
    }

    var elements []*uilib.UIElement

    // game button
    elements = append(elements, makeButton(1, 7, 4, false, func(){
        select {
            case game.Events <- &GameEventGameMenu{}:
            default:
        }
    }))

    // spell button
    elements = append(elements, makeButton(2, 47, 4, false, func(){
        select {
            case game.Events <- &GameEventCastSpellBook{}:
            default:
        }
    }))

    // army button
    elements = append(elements, makeButton(3, 89, 4, false, func(){
        select {
            case game.Events<- &GameEventArmyView{}:
            default:
        }
    }))

    // cities button
    elements = append(elements, makeButton(4, 140, 4, false, func(){
        select {
            case game.Events<- &GameEventCityListView{}:
            default:
        }
    }))

    // magic button
    elements = append(elements, makeButton(5, 184, 4, false, func(){
        select {
            case game.Events<- &GameEventMagicView{}:
            default:
        }
    }))

    // info button
    elements = append(elements, makeButton(6, 226, 4, true, func(){
        ui.AddElements(game.MakeInfoUI(60, 25))
    }))

    // plane button
    elements = append(elements, makeButton(7, 270, 4, false, func(){
        game.SwitchPlane()
        game.doPlanarTraval()

        game.RefreshUI()
    }))

    // minimap
    minimapRect := scale.UnscaleRect(game.GetMinimapRect())
    var minimapPoint image.Point
    moveMinimap := func(){

        middleMapX := minimapRect.Bounds().Dx() / 2
        middleMapY := minimapRect.Bounds().Dy() / 2

        x := game.Camera.GetX() + (minimapPoint.X - middleMapX)
        y := game.Camera.GetY() + (minimapPoint.Y - middleMapY)

        select {
            case game.Events <- &GameEventMoveCamera{Plane: game.Plane, X: x, Y: y, Instant: false}:
            default:
        }
    }
    elements = append(elements, &uilib.UIElement{
        Rect: minimapRect,
        /*
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            util.DrawRect(screen, scale.ScaleRect(minimapRect), color.RGBA{R: 255, A: 255})
        },
        */
        Inside: func(this *uilib.UIElement, x int, y int){
            minimapPoint = image.Pt(x, y)
        },
        RightClick: func(this *uilib.UIElement){
            moveMinimap()
        },
        LeftClick: func(this *uilib.UIElement){
            moveMinimap()
        },
    })

    if len(game.Players) > 0 && game.Players[0].SelectedStack != nil {
        player := game.Players[0]
        // stack := player.SelectedStack

        unitX1 := 246
        unitY1 := 79

        unitX := unitX1
        unitY := unitY1

        minMoves := fraction.Zero()

        row := 0

        allStacks := player.FindAllStacks(player.SelectedStack.X(), player.SelectedStack.Y(), player.SelectedStack.Plane())

        updateMinMoves := func() {
            minMoves = fraction.Zero()
            smallest := fraction.Zero()
            first := true
            for _, stack := range allStacks {
                if first || stack.GetRemainingMoves().LessThan(smallest) {
                    smallest = stack.GetRemainingMoves()
                }

                first = false
            }

            minMoves = smallest
        }

        updateMinMoves()

        for _, stack := range allStacks {
            for _, unit := range stack.Units() {
                // show a unit element for each unit in the stack
                // image index increases by 1 for each unit, indexes 24-32
                disband := func(){
                    player.RemoveUnit(unit)
                    game.RefreshUI()
                    if player.SelectedStack == nil {
                        game.DoNextUnit(player)
                    }
                }

                unitBackground, _ := game.ImageCache.GetImage("main.lbx", 24, 0)
                unitRect := util.ImageRect(unitX, unitY, unitBackground)
                elements = append(elements, &uilib.UIElement{
                    Rect: unitRect,
                    PlaySoundLeftClick: true,
                    LeftClick: func(this *uilib.UIElement){
                        // cannot toggle stasis units
                        if unit.GetBusy() == units.BusyStatusStasis {
                            return
                        }

                        stack.ToggleActive(unit)

                        // FIXME: explain why this is here. Commenting it out doesn't seem to hurt
                        select {
                            case game.Events<- &GameEventMoveUnit{Player: player}:
                            default:
                        }

                        updateMinMoves()
                        game.RefreshUI()
                    },
                    RightClick: func(this *uilib.UIElement){
                        ui.AddGroup(unitview.MakeUnitContextMenu(game.Cache, ui, unit, disband))
                    },
                    Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                        var options ebiten.DrawImageOptions
                        options.GeoM.Translate(float64(unitRect.Min.X), float64(unitRect.Min.Y))
                        scale.DrawScaled(screen, unitBackground, &options)

                        options.GeoM.Translate(1, 1)

                        if stack.IsActive(unit){
                            unitBack, _ := units.GetUnitBackgroundImage(unit.GetBanner(), &game.ImageCache)
                            scale.DrawScaled(screen, unitBack, &options)
                        }

                        options.GeoM.Translate(1, 1)
                        unitImage, err := unitview.GetUnitOverworldImage(&game.ImageCache, unit)
                        if err == nil {

                            if unit.GetBusy() != units.BusyStatusNone {
                                var patrolOptions colorm.DrawImageOptions
                                var matrix colorm.ColorM
                                patrolOptions.GeoM = scale.ScaleGeom(options.GeoM)
                                matrix.ChangeHSV(0, 0, 1)
                                colorm.DrawImage(screen, unitImage, matrix, &patrolOptions)
                            } else {
                                scale.DrawScaled(screen, unitImage, &options)
                            }

                            // draw the first enchantment on the unit
                            for _, enchantment := range unit.GetEnchantments() {
                                util.DrawOutline(screen, &game.ImageCache, unitImage, scale.ScaleGeom(options.GeoM), options.ColorScale, game.Counter/8, enchantment.Color())
                                break
                            }
                        }

                        if unit.GetHealth() < unit.GetMaxHealth() {
                            highHealth := color.RGBA{R: 0, G: 0xff, B: 0, A: 0xff}
                            mediumHealth := color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff}
                            lowHealth := color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}

                            healthWidth := float64(10)
                            healthPercent := float64(unit.GetHealth()) / float64(unit.GetMaxHealth())
                            healthLength := healthWidth * healthPercent

                            // always show at least one point of health
                            if healthLength < 1 {
                                healthLength = 1
                            }

                            useColor := highHealth
                            if healthPercent < 0.33 {
                                useColor = lowHealth
                            } else if healthPercent < 0.66 {
                                useColor = mediumHealth
                            } else {
                                useColor = highHealth
                            }

                            x, y := options.GeoM.Apply(float64(4), float64(19))
                            vector.StrokeLine(screen, scale.Scale(float32(x)), scale.Scale(float32(y)), scale.Scale(float32(x + healthLength)), scale.Scale(float32(y)), float32(scale.ScaleAmount), useColor, false)
                        }

                        badgeInfo := units.GetExperienceBadge(unit)

                        badgeOptions := options
                        badgeOptions.GeoM.Translate(1, 21)
                        for range badgeInfo.Count {
                            pic, _ := game.ImageCache.GetImage("main.lbx", badgeInfo.Badge.IconLbxIndex(), 0)
                            scale.DrawScaled(screen, pic, &badgeOptions)
                            badgeOptions.GeoM.Translate(4, 0)
                        }

                        weaponOptions := options
                        weaponOptions.GeoM.Translate(12, 18)
                        var weapon *ebiten.Image
                        switch unit.GetWeaponBonus() {
                        case data.WeaponMagic:
                            weapon, _ = game.ImageCache.GetImage("main.lbx", 54, 0)
                        case data.WeaponMythril:
                            weapon, _ = game.ImageCache.GetImage("main.lbx", 55, 0)
                        case data.WeaponAdamantium:
                            weapon, _ = game.ImageCache.GetImage("main.lbx", 56, 0)
                        }

                        if weapon != nil {
                            scale.DrawScaled(screen, weapon, &weaponOptions)
                        }

                        useGeom := options.GeoM

                        // draw a G on the unit if they are moving, P if purify, and B if building road
                        if unit.GetBusy() == units.BusyStatusBuildRoad {
                            x, y := useGeom.Apply(float64(1), float64(1))
                            game.Fonts.WhiteFont.Print(screen, x, y, scale.ScaleAmount, options.ColorScale, "B")
                        } else if unit.GetBusy() == units.BusyStatusPurify {
                            x, y := useGeom.Apply(float64(1), float64(1))
                            game.Fonts.WhiteFont.Print(screen, x, y, scale.ScaleAmount, options.ColorScale, "P")
                        } else if len(stack.CurrentPath) != 0 {
                            x, y := useGeom.Apply(float64(1), float64(1))
                            game.Fonts.WhiteFont.Print(screen, x, y, scale.ScaleAmount, options.ColorScale, "G")
                        }
                    },
                })

                row += 1
                unitX += unitBackground.Bounds().Dx()
                if row >= 3 {
                    row = 0
                    unitX = unitX1
                    unitY += unitBackground.Bounds().Dy()
                }
            }

            doneImages, _ := game.ImageCache.GetImages("main.lbx", 8)
            doneIndex := 0
            doneRect := util.ImageRect(246, 176, doneImages[0])
            doneCounter := uint64(0)
            elements = append(elements, &uilib.UIElement{
                Rect: doneRect,
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    colorScale := ebiten.ColorScale{}

                    if doneCounter > 0 {
                        v := float32(1 + (math.Sin(float64(doneCounter / 4)) / 2 + 0.5) / 2)
                        colorScale.Scale(v, v, v, 1)
                    }

                    var options ebiten.DrawImageOptions
                    options.ColorScale.ScaleWithColorScale(colorScale)
                    options.GeoM.Translate(float64(doneRect.Min.X), float64(doneRect.Min.Y))
                    scale.DrawScaled(screen, doneImages[doneIndex], &options)
                },
                Inside: func(this *uilib.UIElement, x int, y int){
                    doneCounter += 1
                },
                NotInside: func(this *uilib.UIElement){
                    doneCounter = 0
                },
                PlaySoundLeftClick: true,
                LeftClick: func(this *uilib.UIElement){
                    doneIndex = 1
                },
                LeftClickRelease: func(this *uilib.UIElement){
                    doneIndex = 0

                    if player.SelectedStack != nil {
                        player.SelectedStack.ExhaustMoves()
                    }

                    game.DoNextUnit(player)
                },
            })

            patrolImages, _ := game.ImageCache.GetImages("main.lbx", 9)
            patrolIndex := 0
            patrolRect := util.ImageRect(280, 176, patrolImages[0])
            patrolCounter := uint64(0)
            elements = append(elements, &uilib.UIElement{
                Rect: patrolRect,
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    colorScale := ebiten.ColorScale{}

                    if patrolCounter > 0 {
                        v := float32(1 + (math.Sin(float64(patrolCounter / 4)) / 2 + 0.5) / 2)
                        colorScale.Scale(v, v, v, 1)
                    }

                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(patrolRect.Min.X), float64(patrolRect.Min.Y))
                    options.ColorScale.ScaleWithColorScale(colorScale)
                    scale.DrawScaled(screen, patrolImages[patrolIndex], &options)
                },
                Inside: func(this *uilib.UIElement, x int, y int){
                    patrolCounter += 1
                },
                NotInside: func(this *uilib.UIElement){
                    patrolCounter = 0
                },
                PlaySoundLeftClick: true,
                LeftClick: func(this *uilib.UIElement){
                    patrolIndex = 1
                },
                LeftClickRelease: func(this *uilib.UIElement){
                    patrolIndex = 0

                    if player.SelectedStack != nil {
                        for _, unit := range player.SelectedStack.ActiveUnits() {
                            unit.SetBusy(units.BusyStatusPatrol)
                        }
                    }

                    player.SelectedStack.EnableMovers()

                    game.DoNextUnit(player)
                },
            })

            waitImages, _ := game.ImageCache.GetImages("main.lbx", 10)
            waitIndex := 0
            waitRect := util.ImageRect(246, 186, waitImages[0])
            waitCounter := uint64(0)
            elements = append(elements, &uilib.UIElement{
                Rect: waitRect,
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    colorScale := ebiten.ColorScale{}

                    if waitCounter > 0 {
                        v := float32(1 + (math.Sin(float64(waitCounter / 4)) / 2 + 0.5) / 2)
                        colorScale.Scale(v, v, v, 1)
                    }

                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(waitRect.Min.X), float64(waitRect.Min.Y))
                    options.ColorScale.ScaleWithColorScale(colorScale)
                    scale.DrawScaled(screen, waitImages[waitIndex], &options)
                },
                Inside: func(this *uilib.UIElement, x int, y int){
                    waitCounter += 1
                },
                NotInside: func(this *uilib.UIElement){
                    waitCounter = 0
                },
                PlaySoundLeftClick: true,
                LeftClick: func(this *uilib.UIElement){
                    waitIndex = 1
                },
                LeftClickRelease: func(this *uilib.UIElement){
                    waitIndex = 0
                    game.DoNextUnit(player)
                },
            })

            // FIXME: use index 15 to show inactive build button
            inactiveBuild, _ := game.ImageCache.GetImages("main.lbx", 15)
            buildImages, _ := game.ImageCache.GetImages("main.lbx", 11)
            meldImages, _ := game.ImageCache.GetImages("main.lbx", 49)
            purifyImages, _ := game.ImageCache.GetImages("main.lbx", 42)
            inactivePurify, _ := game.ImageCache.GetImage("main.lbx", 43, 0)
            buildIndex := 0
            buildRect := util.ImageRect(280, 186, buildImages[0])
            buildCounter := uint64(0)

            hasRoad := game.GetMap(player.SelectedStack.Plane()).ContainsRoad(player.SelectedStack.X(), player.SelectedStack.Y())
            hasCity := game.ContainsCity(player.SelectedStack.X(), player.SelectedStack.Y(), player.SelectedStack.Plane())
            node := game.GetMap(player.SelectedStack.Plane()).GetMagicNode(player.SelectedStack.X(), player.SelectedStack.Y())
            isCorrupted := game.GetMap(player.SelectedStack.Plane()).HasCorruption(player.SelectedStack.X(), player.SelectedStack.Y())
            canSettle := player.SelectedStack.ActiveUnitsHasAbility(data.AbilityCreateOutpost) && game.IsSettlableLocation(player.SelectedStack.X(), player.SelectedStack.Y(), player.SelectedStack.Plane())

            elements = append(elements, &uilib.UIElement{
                Rect: buildRect,
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    var options colorm.DrawImageOptions
                    var matrix colorm.ColorM
                    options.GeoM.Translate(float64(buildRect.Min.X), float64(buildRect.Min.Y))
                    options.GeoM = scale.ScaleGeom(options.GeoM)

                    if buildCounter > 0 {
                        v := 1 + (math.Sin(float64(buildCounter / 4)) / 2 + 0.5) / 2
                        matrix.Scale(v, v, v, 1)
                    }

                    var use *ebiten.Image
                    use = inactiveBuild[0]

                    var powers UnitBuildPowers

                    if player.SelectedStack != nil {
                        powers = computeUnitBuildPowers(player.SelectedStack)
                    }

                    if powers.CreateOutpost {
                        use = buildImages[buildIndex]
                        if !canSettle {
                            use = inactiveBuild[0]
                        }
                    } else if powers.Meld {
                        use = meldImages[buildIndex]

                        canMeld := false
                        if node != nil && !node.Warped {
                            canMeld = true
                        }

                        if !canMeld {
                            matrix.ChangeHSV(0, 0, 1)
                        }
                    } else if powers.Purify {
                        if isCorrupted {
                            use = purifyImages[buildIndex]
                        } else {
                            use = inactivePurify
                        }
                    } else if powers.BuildRoad && !hasRoad && !hasCity {
                        use = buildImages[buildIndex]
                    }

                    colorm.DrawImage(screen, use, matrix, &options)
                },
                Inside: func(this *uilib.UIElement, x int, y int){
                    buildCounter += 1
                },
                NotInside: func(this *uilib.UIElement){
                    buildCounter = 0
                },
                PlaySoundLeftClick: true,
                LeftClick: func(this *uilib.UIElement){
                    var powers UnitBuildPowers

                    if player.SelectedStack != nil {
                        powers = computeUnitBuildPowers(player.SelectedStack)
                    }

                    if powers.CreateOutpost {
                        // FIXME: check if we can build an outpost here
                        buildIndex = 1
                    } else if powers.Meld {
                        canMeld := false
                        if node != nil && !node.Warped {
                            canMeld = true
                        }

                        if canMeld {
                            buildIndex = 1
                        }
                    } else if powers.Purify {
                        if isCorrupted {
                            buildIndex = 1
                        }
                    } else if powers.BuildRoad {
                        if !hasRoad && !hasCity {
                            buildIndex = 1
                        }
                    }

                },
                LeftClickRelease: func(this *uilib.UIElement){
                    // if couldn't left click, then release should do nothing
                    if buildIndex == 0 {
                        return
                    }

                    buildIndex = 0

                    game.DoBuildAction(player)
                },
            })
        }

        elements = append(elements, &uilib.UIElement{
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                if !minMoves.IsZero() {
                    game.Fonts.WhiteFont.PrintOptions(screen, 246, 167, font.FontOptions{DropShadow: true, Scale: scale.ScaleAmount}, fmt.Sprintf("Moves:%v", minMoves.ToFloat()))

                    sailingIcon, _ := game.ImageCache.GetImage("main.lbx", 18, 0)
                    swimmingIcon, _ := game.ImageCache.GetImage("main.lbx", 19, 0)
                    mountaineeringIcon, _ := game.ImageCache.GetImage("main.lbx", 20, 0)
                    foresterIcon, _ := game.ImageCache.GetImage("main.lbx", 21, 0)
                    flyingIcon, _ := game.ImageCache.GetImage("main.lbx", 22, 0)
                    pathfindingIcon, _ := game.ImageCache.GetImage("main.lbx", 23, 0)
                    planeTravelIcon, _ := game.ImageCache.GetImage("main.lbx", 36, 0)
                    windWalkingIcon, _ := game.ImageCache.GetImage("main.lbx", 37, 0)
                    walkingIcon, _ := game.ImageCache.GetImage("main.lbx", 38, 0)

                    _ = sailingIcon
                    _ = swimmingIcon
                    _ = planeTravelIcon
                    _ = windWalkingIcon

                    useIcon := walkingIcon

                    if player.SelectedStack != nil {
                        if player.SelectedStack.AllFlyers() {
                            useIcon = flyingIcon
                        } else if player.SelectedStack.HasPathfinding() {
                            useIcon = pathfindingIcon
                        } else if player.SelectedStack.ActiveUnitsHasAbility(data.AbilityMountaineer) {
                            useIcon = mountaineeringIcon
                        } else if player.SelectedStack.ActiveUnitsHasAbility(data.AbilityForester) {
                            useIcon = foresterIcon
                        } else if player.SelectedStack.AllSwimmers() {
                            useIcon = swimmingIcon
                        }
                    }

                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(246 + float64(60), 167)
                    scale.DrawScaled(screen, useIcon, &options)
                }
            },
        })


    } else {
        // next turn
        nextTurnImage, _ := game.ImageCache.GetImage("main.lbx", 35, 0)
        nextTurnImageClicked, _ := game.ImageCache.GetImage("main.lbx", 58, 0)
        nextTurnRect := image.Rect(240, 174, 240 + nextTurnImage.Bounds().Dx(), 174 + nextTurnImage.Bounds().Dy())
        nextTurnClicked := false
        elements = append(elements, &uilib.UIElement{
            Rect: nextTurnRect,
            PlaySoundLeftClick: true,
            LeftClick: func(this *uilib.UIElement){
                nextTurnClicked = true
            },
            LeftClickRelease: func(this *uilib.UIElement){
                nextTurnClicked = false
                select {
                    case game.Events <- &GameEventNextTurn{}:
                    default:
                }
            },
            RightClick: func(this *uilib.UIElement){
                helpEntries := game.Help.GetEntriesByName("Next Turn")
                if helpEntries != nil {
                    group.AddElement(uilib.MakeHelpElementWithLayer(group, game.Cache, &game.ImageCache, 1, helpEntries[0], helpEntries[1:]...))
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(240, 174)
                scale.DrawScaled(screen, nextTurnImage, &options)
                if nextTurnClicked {
                    options.GeoM.Translate(6, 5)
                    scale.DrawScaled(screen, nextTurnImageClicked, &options)
                }
            },
        })

        if len(game.Players) > 0 {
            player := game.Players[0]

            goldPerTurn := player.GoldPerTurn()
            foodPerTurn := player.FoodPerTurn()
            manaPerTurn := player.ManaPerTurn(game.ComputePower(player), game)

            conjunction, conjunctionColor := game.ActiveConjunctionName()

            goldFood, _ := game.ImageCache.GetImage("main.lbx", 34, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(240, 77)
            elements = append(elements, &uilib.UIElement{
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    scale.DrawScaled(screen, goldFood, &options)

                    negativeScale := ebiten.ColorScale{}

                    // v is in range 0.5-1
                    v := (math.Cos(float64(game.Counter) / 7) + 1) / 4 + 0.5
                    negativeScale.SetR(float32(v))

                    negative := options
                    negative.ColorScale = negativeScale

                    negativeOptions := font.FontOptions{Justify: font.FontJustifyCenter, Options: &negative, Scale: scale.ScaleAmount}
                    normalOptions := font.FontOptions{Justify: font.FontJustifyCenter, Options: &options, Scale: scale.ScaleAmount}

                    if goldPerTurn < 0 {
                        game.Fonts.InfoFontRed.PrintOptions(screen, 278, 103, negativeOptions, fmt.Sprintf("%v Gold", goldPerTurn))
                    } else {
                        game.Fonts.InfoFontYellow.PrintOptions(screen, 278, 103, normalOptions, fmt.Sprintf("%v Gold", goldPerTurn))
                    }

                    if foodPerTurn < 0 {
                        game.Fonts.InfoFontRed.PrintOptions(screen, 278, 135, negativeOptions, fmt.Sprintf("%v Food", foodPerTurn))
                    } else {
                        game.Fonts.InfoFontYellow.PrintOptions(screen, 278, 135, normalOptions, fmt.Sprintf("%v Food", foodPerTurn))
                    }

                    if manaPerTurn < 0 {
                        game.Fonts.InfoFontRed.PrintOptions(screen, 278, 167, negativeOptions, fmt.Sprintf("%v Mana", manaPerTurn))
                    } else {
                        game.Fonts.InfoFontYellow.PrintOptions(screen, 278, 167, normalOptions, fmt.Sprintf("%v Mana", manaPerTurn))
                    }

                    if conjunction != "" {
                        conjunctionOptions := options
                        conjunctionOptions.ColorScale.ScaleWithColor(conjunctionColor)
                        game.Fonts.WhiteFont.PrintOptions(screen, 278, 155, font.FontOptions{Justify: font.FontJustifyCenter, Options: &conjunctionOptions, Scale: scale.ScaleAmount}, conjunction)
                    }
                },
            })
        }
    }

    elements = append(elements, &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            game.Fonts.WhiteFont.PrintOptions(screen, 276, 68, font.FontOptions{Justify: font.FontJustifyRight, DropShadow: true, Scale: scale.ScaleAmount}, fmt.Sprintf("%v GP", game.Players[0].Gold))
        },
    })

    elements = append(elements, &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            game.Fonts.WhiteFont.PrintOptions(screen, 314, 68, font.FontOptions{Justify: font.FontJustifyRight, DropShadow: true, Scale: scale.ScaleAmount}, fmt.Sprintf("%v MP", game.Players[0].Mana))
        },
    })

    ui.SetElementsFromArray(elements)

    return ui
}

func (game *Game) DoNextUnit(player *playerlib.Player){
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

            if player.IsHuman() {
                select {
                    case game.Events <- &GameEventMoveCamera{Plane: stack.Plane(), X: stack.X(), Y: stack.Y(), Instant: true}:
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

    if player.IsHuman() {
        /*
        if player.SelectedStack == nil {
            fortressCity := player.FindFortressCity()
            if fortressCity != nil {
                game.CenterCamera(fortressCity.X, fortressCity.Y)
            }
        }
        */

        if player.SelectedStack != nil && len(player.SelectedStack.CurrentPath) > 0 {
            select {
                case game.Events<- &GameEventMoveUnit{Player: player}:
                default:
            }
        }

        game.RefreshUI()
    }
}

/* return a tuple of booleans where each boolean is true if the corresponding resource is not enough
 * to support the units.
 * (gold, food, mana)
 *
 * (false, false, false) means all units are supported.
 */
func (game *Game) CheckDisband(player *playerlib.Player) (bool, bool, bool) {

    unitsNeedGold := false
    unitsNeedFood := false
    unitsNeedMana := false

    for _, unit := range player.Units {
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
    manaPerTurn := player.ManaPerTurn(game.ComputePower(player), game)

    manaIssue := player.Mana + manaPerTurn < 0 && unitsNeedMana

    return goldIssue, foodIssue, manaIssue
}

/* disband units due to lack of resources, return an array of messages about units that were lost
 */
func (game *Game) DisbandUnits(player *playerlib.Player) []string {
    /*
    if 2 > 1 {
        return nil
    }
    */

    // keep removing units until the upkeep value can be paid
    ok := false
    var disbandedMessages []string
    for len(player.Units) > 0 && !ok {
        ok = true

        goldIssue, foodIssue, manaIssue := game.CheckDisband(player)

        if goldIssue || foodIssue || manaIssue {
            ok = false
            disbanded := false

            // try to disband one unit that is taking up resources
            for i := len(player.Units) - 1; i >= 0; i-- {
                unit := player.Units[i]
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
                mapUse := game.GetMap(stack.Plane())
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

// the amount of experience a unit in a stack should get at the end of each turn
// if there is a hero in the stack then the hero's armsmaster ability applies
func (game *Game) GetExperienceBonus(stack *playerlib.UnitStack) int {
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


func (game *Game) GetCityEnchantmentsByBanner(banner data.BannerType) []playerlib.CityEnchantment {
    var result []playerlib.CityEnchantment

    for _, player := range game.Players {
        for _, city := range player.Cities {
            for _, enchantment := range city.GetEnchantmentsCastBy(banner) {
                result = append(result, playerlib.CityEnchantment{City: city, Enchantment: enchantment})
            }
        }
    }

    return result
}

// turn off enchantments that can not be afforded
func (game *Game) DissipateEnchantments(player *playerlib.Player, power int) {
    // if time stop is in effect then only check that spell for dissipation, all other spells have no upkeep
    if player.HasEnchantment(data.EnchantmentTimeStop) {
        if player.Mana < data.EnchantmentTimeStop.UpkeepMana() {
            player.RemoveEnchantment(data.EnchantmentTimeStop)
        }

        return
    }

    isManaIssue := func() bool {
        manaPerTurn := player.ManaPerTurn(power, game)
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
        enchantments := game.GetCityEnchantmentsByBanner(player.GetBanner())
        if len(enchantments) == 0 || !isManaIssue() {
            break
        }

        enchantment := enchantments[rand.N(len(enchantments))]
        enchantment.City.CancelEnchantment(enchantment.Enchantment.Enchantment, enchantment.Enchantment.Owner)
    }

    var enchantedUnits []units.StackUnit
    for _, unit := range player.Units {
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

func (game *Game) applyChaosChannels(unit *units.OverworldUnit) {
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

func handleStasis(stack *playerlib.UnitStack) {
    // this seems like a reasonable place to handle stasis, but it could be moved elsewhere
    for _, unit := range stack.Units() {
        if unit.GetBusy() == units.BusyStatusStasis {
            if rand.N(10) + 1 < combat.GetResistanceFor(unit, data.SorceryMagic) - 5 {
                unit.SetBusy(units.BusyStatusNone)
            }
        }
    }
}

func (game *Game) StartPlayerTurn(player *playerlib.Player) {
    if player.IsHuman() {
        game.ScrollEvents = nil
    }

    disbandedMessages := game.DisbandUnits(player)

    if player.IsHuman() && len(disbandedMessages) > 0 {
        select {
            case game.Events<- &GameEventScroll{Title: "", Text: strings.Join(disbandedMessages, "\n")}:
            default:
        }
    }

    power := game.ComputePower(player)

    game.DissipateEnchantments(player, power)

    // timestop may have dissipated by now
    timeStop := player.HasEnchantment(data.EnchantmentTimeStop)

    player.Gold += player.GoldPerTurn()
    if player.Gold < 0 {
        player.Gold = 0
    }

    player.Mana += player.ManaPerTurn(power, game)

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

        player.CastingSpellProgress += manaSpent
        player.Mana -= manaSpent

        if spellCost <= player.CastingSpellProgress {
            game.doCastSpell(player, player.CastingSpell)
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
                        case game.Events<- &GameEventLearnedSpell{Player: player, Spell: player.ResearchingSpell}:
                        default:
                    }
                }

                // log.Printf("wizard %v learned %v", player.Wizard.Name, player.ResearchingSpell.Name)

                player.LearnSpell(player.ResearchingSpell)

                if player.IsHuman() {
                    select {
                        case game.Events<- &GameEventResearchSpell{Player: player}:
                        default:
                    }
                }
            }
        }
    } else if game.TurnNumber > 0 {

        if player.IsHuman() {
            select {
                case game.Events<- &GameEventResearchSpell{Player: player}:
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
            cityEvents := city.DoNextTurn(game.GetMap(city.Plane))
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
                            case game.Events<- &scrollEvent:
                            default:
                        }
                    }
                case *citylib.CityEventCityAbandoned:
                    removeCities = append(removeCities, city)
                    if player.IsHuman() {
                        select {
                            case game.Events<- &GameEventNotice{Message: fmt.Sprintf("The city of %v has been abandoned.", city.Name)}:
                            default:
                        }
                    }
                case *citylib.CityEventNewBuilding:
                    newBuilding := event.(*citylib.CityEventNewBuilding)

                    if player.IsHuman() {
                        select {
                            case game.Events<- &GameEventNewBuilding{City: city, Building: newBuilding.Building, Player: player}:
                            default:
                        }
                    } else {
                        log.Printf("ai created %v", game.BuildingInfo.Name(newBuilding.Building))
                    }
                case *citylib.CityEventOutpostDestroyed:
                    removeCities = append(removeCities, city)
                    if player.IsHuman() {
                        select {
                            case game.Events<- &GameEventNotice{Message: fmt.Sprintf("The outpost of %v has been deserted.", city.Name)}:
                            default:
                        }
                    }
                case *citylib.CityEventOutpostHamlet:
                    if player.IsHuman() {
                        select {
                            case game.Events<- &GameEventNotice{Message: fmt.Sprintf("The outpost of %v has grown into a hamlet.", city.Name)}:
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
                        game.applyChaosChannels(overworldUnit)
                    }

                    overworldUnit.AddExperience(newUnit.Experience)
                    player.AddUnit(overworldUnit)
                    game.ResolveStackAt(city.X, city.Y, city.Plane)

                    if player.AIBehavior != nil {
                        player.AIBehavior.ProducedUnit(city, player)
                    }
                }
            }
        }
    }

    game.DoBuildRoads(player)
    game.DoPurify(player)

    for _, stack := range player.Stacks {

        // every unit gains 1 experience at each turn
        unitExperience := game.GetExperienceBonus(stack)
        for _, unit := range stack.Units() {
            switch unit.GetRace() {
                case data.RaceHero: game.AddExperience(player, unit, 1)
                case data.RaceFantastic: // nothing
                default:
                    game.AddExperience(player, unit, unitExperience)
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

    game.maybeHireHero(player)
    game.maybeHireMercenaries(player)
    game.maybeBuyFromMerchant(player)

    player.UpdateFogVisibility()

    if player.GlobalEnchantments.Contains(data.EnchantmentAwareness) {
        game.doExploreFogForAwareness(player)
    }

    // game.CenterCamera(player.Cities[0].X, player.Cities[0].Y)
    game.DoNextUnit(player)
    if player.IsHuman() {
        game.RefreshUI()
    }
}

func (game *Game) revertVolcanos() {
    mapObjects := []*maplib.Map{game.ArcanusMap, game.MyrrorMap}
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

// returns the number of people, units, buildings that were lost
func (game *Game) doEarthquake(city *citylib.City, player *playerlib.Player) (int, int, []buildinglib.Building) {
    // FIXME: destroy buildings with 15% chance and non-flying units with 25% chance
    // https://masterofmagic.fandom.com/wiki/Earthquake

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

    for _, unit := range killedUnits {
        player.RemoveUnit(unit)
    }

    var destroyedBuildings []buildinglib.Building
    for _, building := range city.Buildings.Values() {
        roll := rand.N(100)
        if roll < 15 {
            destroyedBuildings = append(destroyedBuildings, building)
            city.Buildings.Remove(building)
        }
    }

    return people, len(killedUnits), destroyedBuildings
}

// At the beginning of each turn, Awareness clears the fog from all cities for enchantment's owner (newly built included)
func (game *Game) doExploreFogForAwareness(awarenessOwner *playerlib.Player) {
    for _, city := range game.AllCities() {
        if city.GetBanner() == awarenessOwner.GetBanner() {
            continue // No need, those cities do already provide vision
        }
        awarenessOwner.ExploreFogSquare(city.X, city.Y, 1, city.Plane)
    }
}

// returns number of citizens killed, units killed, and buildings destroyed
func (game *Game) doCallTheVoid(city *citylib.City, player *playerlib.Player) (int, int, int) {
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

    mapUse := game.GetMap(city.Plane)

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

// raises 4 to 6 volcanoes on random tiles
func (game *Game) doArmageddon() {
    info := game.ComputeCityStackInfo()
    for _, player := range game.Players {
        if player.GlobalEnchantments.Contains(data.EnchantmentArmageddon) {
            // get a list of valid map tiles on both planes
            var points []data.PlanePoint
            catchment := player.GetAllCatchmentArea()
            mapObjects := []*maplib.Map{game.ArcanusMap, game.MyrrorMap}
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
                game.GetMap(point.Plane).SetVolcano(point.X, point.Y, player)
            }
        }
    }
}

// corrupts 3-6 random tiles
func (game *Game) doGreatWasting() {
    info := game.ComputeCityStackInfo()
    for _, player := range game.Players {
        if player.GlobalEnchantments.Contains(data.EnchantmentGreatWasting) {
            // get a list of valid map tiles on both planes
            var points []data.PlanePoint
            catchment := player.GetAllCatchmentArea()
            mapObjects := []*maplib.Map{game.ArcanusMap, game.MyrrorMap}
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
                game.GetMap(point.Plane).SetCorruption(point.X, point.Y)
            }
        }
    }
}

// city is controlled by the newOwner instead of owner
func ChangeCityOwner(city *citylib.City, owner *playerlib.Player, newOwner *playerlib.Player, enchantmentChange ChangeCityEnchantments) {
    owner.RemoveCity(city)
    newOwner.AddCity(city)
    city.ReignProvider = newOwner

    city.Buildings.Remove(buildinglib.BuildingFortress)
    city.Buildings.Remove(buildinglib.BuildingSummoningCircle)

    switch enchantmentChange {
        case ChangeCityKeepEnchantments:
        case ChangeCityRemoveOwnerEnchantments:
            city.RemoveAllEnchantmentsByOwner(owner.GetBanner())
        case ChangeCityRemoveAllEnchantments:
            city.Enchantments.Clear()
    }

    city.UpdateUnrest()
}

func (game *Game) ManaShortActive() bool {
    return slices.ContainsFunc(game.RandomEvents, func(event *RandomEvent) bool {
        return event.Type == RandomEventManaShort
    })
}

func (game *Game) PopulationBoomActive(city *citylib.City) bool {
    return slices.ContainsFunc(game.RandomEvents, func(event *RandomEvent) bool {
        return event.Type == RandomEventPopulationBoom && event.TargetCity == city
    })
}

func (game *Game) PlagueActive(city *citylib.City) bool {
    return slices.ContainsFunc(game.RandomEvents, func(event *RandomEvent) bool {
        return event.Type == RandomEventPlague && event.TargetCity == city
    })
}

func (game *Game) GoodMoonActive() bool {
    return slices.ContainsFunc(game.RandomEvents, func(event *RandomEvent) bool {
        return event.Type == RandomEventGoodMoon
    })
}

func (game *Game) BadMoonActive() bool {
    return slices.ContainsFunc(game.RandomEvents, func(event *RandomEvent) bool {
        return event.Type == RandomEventBadMoon
    })
}

func (game *Game) ConjunctionChaosActive() bool {
    return slices.ContainsFunc(game.RandomEvents, func(event *RandomEvent) bool {
        return event.Type == RandomEventConjunctionChaos
    })
}

func (game *Game) ConjunctionNatureActive() bool {
    return slices.ContainsFunc(game.RandomEvents, func(event *RandomEvent) bool {
        return event.Type == RandomEventConjunctionNature
    })
}

func (game *Game) ConjunctionSorceryActive() bool {
    return slices.ContainsFunc(game.RandomEvents, func(event *RandomEvent) bool {
        return event.Type == RandomEventConjunctionSorcery
    })
}

func (game *Game) ActiveConjunctionName() (string, color.Color) {

    for _, event := range game.RandomEvents {
        switch event.Type {
            case RandomEventConjunctionChaos: return "Conjunction", color.RGBA{R: 255, G: 0, B: 0, A: 255}
            case RandomEventConjunctionNature: return "Conjunction", color.RGBA{R: 0, G: 255, B: 0, A: 255}
            case RandomEventConjunctionSorcery: return "Conjunction", color.RGBA{R: 0, G: 0, B: 255, A: 255}
            case RandomEventManaShort: return "Mana Short", color.RGBA{R: 0, G: 255, B: 0, A: 255}
            case RandomEventGoodMoon: return "Good Moon", color.RGBA{R: 255, G: 255, B: 255, A: 255}
            case RandomEventBadMoon: return "Bad Moon", color.RGBA{R: 0, G: 0, B: 0, A: 255}
        }
    }

    return "", color.RGBA{}
}

func (game *Game) DoRandomEvents() {
    // maybe create a new event
    eventModifier := fraction.FromInt(1)
    switch game.Settings.Difficulty {
        case data.DifficultyIntro: eventModifier = fraction.Make(1, 2)
        case data.DifficultyEasy: eventModifier = fraction.Make(2, 3)
        case data.DifficultyAverage: eventModifier = fraction.Make(3, 4)
        case data.DifficultyHard: eventModifier = fraction.Make(4, 5)
        case data.DifficultyExtreme: eventModifier = fraction.Make(1, 1)
        case data.DifficultyImpossible: eventModifier = fraction.Make(6, 5)
    }

    // for testing purposes
    // eventModifier = fraction.FromInt(10)

    eventProbability := fraction.FromInt(int(game.TurnNumber - game.LastEventTurn)).Multiply(eventModifier)
    if game.TurnNumber < 50 || game.TurnNumber - game.LastEventTurn < 5 {
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
        for _, event := range game.RandomEvents {
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

        if game.TurnNumber < 150 {
            choices.Remove(RandomEventDiplomaticMarriage)
            choices.Remove(RandomEventGreatMeteor)
        }

        if choices.Size() > 0 {
            choice := choices.Values()[rand.N(choices.Size())]

            // return a RandomEvent object to show, and also cause the event to occur (if instant)
            makeEvent := func (choice RandomEventType, target *playerlib.Player) (*RandomEvent, GameEvent) {
                usedCities := set.NewSet[*citylib.City]()
                for _, event := range game.RandomEvents {
                    if event.TargetCity != nil {
                        usedCities.Insert(event.TargetCity)
                    }
                }

                switch choice {
                    case RandomEventBadMoon: return MakeBadMoonEvent(game.TurnNumber), nil
                    case RandomEventGoodMoon: return MakeGoodMoonEvent(game.TurnNumber), nil
                    case RandomEventConjunctionChaos: return MakeConjunctionChaosEvent(game.TurnNumber), nil
                    case RandomEventConjunctionNature: return MakeConjunctionNatureEvent(game.TurnNumber), nil
                    case RandomEventConjunctionSorcery: return MakeConjunctionSorceryEvent(game.TurnNumber), nil
                    case RandomEventManaShort: return MakeManaShortEvent(game.TurnNumber), nil
                    case RandomEventDisjunction:
                        // there must be at least one global enchantment for this event to occur
                        hasGlobalEnchantment := false

                        for _, player := range game.Players {
                            if player.GlobalEnchantments.Size() > 0 {
                                hasGlobalEnchantment = true
                                break
                            }
                        }

                        if !hasGlobalEnchantment {
                            return nil, nil
                        }

                        // remove all global enchantments
                        for _, player := range game.Players {
                            player.GlobalEnchantments.Clear()
                        }

                        return MakeDisjunctionEvent(game.TurnNumber), nil
                    case RandomEventDonation:
                        // FIXME: what are the bounds here?
                        gold := rand.N(2000) + 100
                        target.Gold += gold

                        return MakeDonationEvent(game.TurnNumber, gold), nil
                    case RandomEventPiracy:
                        if target.Gold < 100 {
                            return nil, nil
                        }

                        // between 30-50%, compute random number between 0-20%, add 30%
                        gold := rand.N(target.Gold / 5) + target.Gold * 3 / 10
                        target.Gold = max(0, target.Gold - gold)

                        return MakePiracyEvent(game.TurnNumber, gold), nil
                    case RandomEventGift:
                        var out []*artifact.Artifact
                        for _, artifact := range game.ArtifactPool {
                            if canUseArtifact(artifact, target.Wizard) {
                                out = append(out, artifact)
                            }
                        }

                        // couldn't find a valid artifact
                        if len(out) == 0 {
                            return nil, nil
                        }

                        use := out[rand.N(len(out))]

                        delete(game.ArtifactPool, use.Name)

                        // returning GameEventVault here is ugly but we need a way to have the vault event
                        // be added to game.Events after the random event
                        return MakeGiftEvent(game.TurnNumber, use.Name), &GameEventVault{CreatedArtifact: use, Player: target}
                    case RandomEventDepletion:
                        // choose a random town that has a mineral bonus in its catchment area,
                        // and then remove the bonus from the map
                        cities := target.GetCities()
                        for _, cityIndex := range rand.Perm(len(cities)) {
                            city := cities[cityIndex]
                            mapUse := game.GetMap(city.Plane)
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
                                return MakeDepletionEvent(game.TurnNumber, tile.GetBonus(), city.Name), nil
                            }
                        }

                        return nil, nil

                    case RandomEventDiplomaticMarriage:
                        for _, player := range game.Players {
                            if player.GetBanner() == data.BannerBrown {
                                if len(player.Cities) > 0 {
                                    cities := target.GetCities()
                                    city := cities[rand.N(len(cities))]
                                    // if the owner of the city has a stack garrisoned there then the garrison is disbanded
                                    stack := player.FindStack(city.X, city.Y, city.Plane)
                                    if stack != nil {
                                        for _, unit := range stack.Units() {
                                            player.RemoveUnit(unit)
                                        }
                                    }

                                    ChangeCityOwner(city, player, target, ChangeCityRemoveAllEnchantments)

                                    return MakeDiplomaticMarriageEvent(game.TurnNumber, city), nil
                                }
                            }
                        }

                        return nil, nil

                    case RandomEventEarthquake:
                        choices := game.AllCities()
                        if len(choices) == 0 {
                            return nil, nil
                        }

                        city := choices[rand.N(len(choices))]

                        people, units, buildings := game.doEarthquake(city, target)

                        return MakeEarthquakeEvent(game.TurnNumber, city.Name, people, units, len(buildings)), nil

                    case RandomEventGreatMeteor:
                        choices := game.AllCities()
                        if len(choices) == 0 {
                            return nil, nil
                        }

                        city := choices[rand.N(len(choices))]

                        people, units, buildings := game.doCallTheVoid(city, target)

                        return MakeGreatMeteorEvent(game.TurnNumber, city.Name, people, units, buildings), nil

                    case RandomEventNewMinerals:
                        cities := target.GetCities()
                        for _, cityIndex := range rand.Perm(len(cities)) {
                            city := cities[cityIndex]
                            mapUse := game.GetMap(city.Plane)
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
                                return MakeNewMineralsEvent(game.TurnNumber, bonus, city), nil
                            }
                        }

                    case RandomEventPlague:
                        cities := target.GetCities()
                        for _, cityIndex := range rand.Perm(len(cities)) {
                            city := cities[cityIndex]
                            if !usedCities.Contains(city) {
                                return MakePlagueEvent(game.TurnNumber, city), nil
                            }
                        }

                        return nil, nil

                    case RandomEventPopulationBoom:
                        cities := target.GetCities()
                        for _, cityIndex := range rand.Perm(len(cities)) {
                            city := cities[cityIndex]
                            if !usedCities.Contains(city) {
                                return MakePopulationBoomEvent(game.TurnNumber, city), nil
                            }
                        }

                        return nil, nil

                    case RandomEventRebellion:
                        if len(target.Cities) == 0 {
                            return nil, nil
                        }

                        var neutralPlayer *playerlib.Player
                        for _, neutral := range game.Players {
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

                                return MakeRebellionEvent(game.TurnNumber, city), nil
                            }
                        }

                        return nil, nil
                }

                return nil, nil
            }

            targetWizard := game.Players[rand.N(len(game.Players))]
            newEvent, extraEvent := makeEvent(choice, targetWizard)
            if newEvent != nil {
                game.LastEventTurn = game.TurnNumber

                if !newEvent.Instant {
                    game.RandomEvents = append(game.RandomEvents, newEvent)
                }

                // FIXME: if the event is targeting an AI wizard then the event message should change slightly
                game.Events <- &GameEventShowRandomEvent{Event: newEvent, Starting: true}

                if extraEvent != nil {
                    game.Events <- extraEvent
                }

                game.RefreshUI()
            }
        }

    }

    var keep []*RandomEvent
    // add events to the 'keep' array to keep them for the next turn
    for _, event := range game.RandomEvents {

        // once citizens has reached 2, plague will dissipate automatically
        if event.Type == RandomEventPlague && event.TargetCity.Citizens() <= 2 {
            game.Events <- &GameEventShowRandomEvent{Event: event, Starting: false}
            continue
        }

        // a random event can end after 5 turns, and the chances of it ending are 5% per turn
        turns := game.TurnNumber - event.BirthYear
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
            game.Events <- &GameEventShowRandomEvent{Event: event, Starting: false}
        } else {
            keep = append(keep, event)
        }
    }

    game.RandomEvents = keep
}

func (game *Game) doChaosRift() {
    for _, city := range game.AllCities() {
        if city.HasEnchantment(data.CityEnchantmentChaosRift) {

            // do 5 magical attacks of strength 8 to units in the city
            stack, player := game.FindStack(city.X, city.Y, city.Plane)
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
                if game.BuildingInfo.ProductionCost(building) != 0 {
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

func (game *Game) doMeteorStorm() {
    immolate := func (unit units.StackUnit) int {
        damage, _ := combat.ApplyAreaDamage(&UnitDamageWrapper{StackUnit: unit}, 4, units.DamageImmolation, 0)
        return damage
    }

    if !game.HasEnchantment(data.EnchantmentMeteorStorm) {
        return
    }

    affectedPlayers := set.NewSet[*playerlib.Player]()
    for _, player := range game.Players {
        if player.Defeated {
            continue
        }

        // add every other player to the affected set.
        // if another player also cast meteor storm, then ultimately the current player will also
        // be added to the set, meaning both players' meteor storms affects the other
        if player.HasEnchantment(data.EnchantmentMeteorStorm) {
            for _, otherPlayer := range game.Players {
                if otherPlayer != player {
                    affectedPlayers.Insert(otherPlayer)
                }
            }
        }
    }

    entityInfo := game.ComputeCityStackInfo()

    for _, player := range game.Players {
        if player.Defeated {
            continue
        }

        // non-garrisoned units take immolation damage
        for _, unit := range slices.Clone(player.Units) {
            if entityInfo.FindCity(unit.GetX(), unit.GetY(), unit.GetPlane()) == nil {
                immolate(unit)
                if unit.GetHealth() <= 0 {
                    player.RemoveUnit(unit)
                }
            }
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

// stolen from combat/model.go
func computeUnitPower(unit units.StackUnit) int {
    power := 0

    power += unit.GetMaxHealth()
    power += unit.GetDefense()
    power += unit.GetResistance()
    power += unit.GetRangedAttackPower() * unit.GetCount()
    power += unit.GetMeleeAttackPower() * unit.GetCount()

    return power
}

func (game *Game) ComputeWizardPower(player *playerlib.Player) playerlib.WizardPower {
    armyStrength := 0

    for _, stack := range player.Stacks {
        for _, unit := range stack.Units() {
            armyStrength += computeUnitPower(unit)
        }
    }

    manaPerTurn := player.ManaPerTurn(game.ComputePower(player), game)

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

func (game *Game) EndOfTurn() {
    // put stuff here that should happen when all players have taken their turn

    for _, player := range game.Players {
        if !player.Defeated {
            player.AddPowerHistory(game.ComputeWizardPower(player))
        }
    }

    game.revertVolcanos()

    // FIXME: the wiki says armageddon will not do anything while time stop is in effect.
    // figure out what other global spells don't have any effect (great wasting, chaos rift, meteor storm)
    if !game.HasEnchantment(data.EnchantmentTimeStop) {
        game.doArmageddon()
    }

    game.doGreatWasting()

    game.doChaosRift()

    game.doMeteorStorm()

    game.TurnNumber += 1

    game.DoRandomEvents()

    for _, player := range game.Players {
        if player.Defeated || player.Banished {
            continue
        }

        player.UpdateDiplomaticRelations()
    }

}

func (game *Game) DoNextTurn(){
    // if time stop is enabled then don't move to the other players, just keep doing the current player
    if game.CurrentPlayer >= 0 && game.Players[game.CurrentPlayer].HasEnchantment(data.EnchantmentTimeStop) {
        game.EndOfTurn()
    } else {
        game.CurrentPlayer += 1
        if game.CurrentPlayer >= len(game.Players) {
            // all players did their turn, so the next global turn starts
            game.EndOfTurn()
            game.CurrentPlayer = 0
        }
    }

    if len(game.Players) > 0 {
        player := game.Players[game.CurrentPlayer]

        if player.Wizard.Banner != data.BannerBrown {
            game.StartPlayerTurn(player)
        } else {
            // neutral enemies should reset their moves each turn
            for _, stack := range player.Stacks {
                handleStasis(stack)
                stack.ResetMoves()
                stack.EnableMovers()
            }
        }

        aiPlayer := game.Players[game.CurrentPlayer]
        if aiPlayer.AIBehavior != nil {
            aiPlayer.AIBehavior.NewTurn(aiPlayer)
        }
    }
}

func (overworld *Overworld) DrawFog(screen *ebiten.Image, geom ebiten.GeoM){

    fogImage := func(index int) *ebiten.Image {
        img, err := overworld.ImageCache.GetImage("mapback.lbx", index, 0)
        if err != nil {
            log.Printf("Error: image in mapback.lbx is missing: %v", err)
            return nil
        }
        return img
    }

    FogEdge_N_E := fogImage(0)
    FogEdge_S_E := fogImage(1)
    FogEdge_E := fogImage(2)
    FogEdge_S_W := fogImage(3)
    FogEdge_S := fogImage(5)
    FogEdge_N_W := fogImage(7)
    FogEdge_N := fogImage(8)
    FogEdge_W := fogImage(11)
    FogCorner_NE := fogImage(10)
    FogCorner_SW := fogImage(13)
    FogCorner_SE := fogImage(6)
    FogCorner_NW := fogImage(12)

    tileWidth := overworld.Map.TileWidth()
    tileHeight := overworld.Map.TileHeight()

    /*
    tilesPerRow := overworld.Map.TilesPerRow(screen.Bounds().Dx())
    tilesPerColumn := overworld.Map.TilesPerColumn(screen.Bounds().Dy())
    */
    var options ebiten.DrawImageOptions

    fog := overworld.Fog

    checkFog := func(x int, y int, fogType data.FogType) bool {
        x = overworld.Map.WrapX(x)
        if x < 0 || x >= len(fog) || y >= len(fog[x]) || y < 0{
            return false
        }

        return fog[x][y] == fogType
    }

    fogN := func(x int, y int, fogType data.FogType) bool {
        return checkFog(x, y - 1, fogType)
    }

    fogE := func(x int, y int, fogType data.FogType) bool {
        return checkFog(x + 1, y, fogType)
    }

    fogS := func(x int, y int, fogType data.FogType) bool {
        return checkFog(x, y + 1, fogType)
    }

    fogW := func(x int, y int, fogType data.FogType) bool {
        return checkFog(x - 1, y, fogType)
    }

    fogNE := func(x int, y int, fogType data.FogType) bool {
        return checkFog(x + 1, y - 1, fogType)
    }

    fogSE := func(x int, y int, fogType data.FogType) bool {
        return checkFog(x + 1, y + 1, fogType)
    }

    fogNW := func(x int, y int, fogType data.FogType) bool {
        return checkFog(x - 1, y - 1, fogType)
    }

    fogSW := func(x int, y int, fogType data.FogType) bool {
        return checkFog(x - 1, y + 1, fogType)
    }

    minX, minY, maxX, maxY := overworld.Camera.GetTileBounds()

    drawFogTile := func(tileX int, tileY int) {
        if overworld.FogBlack != nil {
            scale.DrawScaled(screen, overworld.FogBlack, &options)
        }
    }

    drawFogBorder := func(tileX int, tileY int, fogType data.FogType) {
        n := fogN(tileX, tileY, fogType)
        e := fogE(tileX, tileY, fogType)
        s := fogS(tileX, tileY, fogType)
        w := fogW(tileX, tileY, fogType)
        ne := fogNE(tileX, tileY, fogType)
        se := fogSE(tileX, tileY, fogType)
        nw := fogNW(tileX, tileY, fogType)
        sw := fogSW(tileX, tileY, fogType)

        if n && e {
            scale.DrawScaled(screen, FogEdge_N_E, &options)
        } else if n {
            scale.DrawScaled(screen, FogEdge_N, &options)
        } else if e {
            scale.DrawScaled(screen, FogEdge_E, &options)
        } else if ne {
            scale.DrawScaled(screen, FogCorner_NE, &options)
        }

        if s && e {
            scale.DrawScaled(screen, FogEdge_S_E, &options)
        } else if s {
            scale.DrawScaled(screen, FogEdge_S, &options)
        } else if se {
            scale.DrawScaled(screen, FogCorner_SE, &options)
        }

        if n && w {
            scale.DrawScaled(screen, FogEdge_N_W, &options)
        } else if w {
            scale.DrawScaled(screen, FogEdge_W, &options)
        } else if nw {
            scale.DrawScaled(screen, FogCorner_NW, &options)
        }

        if s && w {
            scale.DrawScaled(screen, FogEdge_S_W, &options)
        } else if sw {
            scale.DrawScaled(screen, FogCorner_SW, &options)
        }
    }

    // log.Printf("fog min %v, %v max %v, %v", minX, minY, maxX, maxY)

    black := ebiten.ColorScale{}
    black.Scale(1, 1, 1, 1)

    lightTransparent := ebiten.ColorScale{}
    lightTransparent.Scale(1, 1, 1, 0.3)

    darkTransparent := ebiten.ColorScale{}
    darkTransparent.Scale(1, 1, 1, 0.5)

    for x := minX; x < maxX; x++ {
        for y := minY; y < maxY; y++ {
            tileX := overworld.Map.WrapX(x)
            tileY := y

            options.GeoM.Reset()
            options.GeoM.Translate(float64(x * tileWidth), float64(y * tileHeight))
            options.GeoM.Concat(geom)

            if tileX >= 0 && tileY >= 0 && tileX < len(fog) && tileY < len(fog[tileX]) {
                switch fog[tileX][tileY] {
                    case data.FogTypeUnexplored:
                        options.ColorScale = black
                        drawFogTile(tileX, tileY)

                    // FIXME: make drawing fog of war configurable?
                    // This would be with no fog of war like the original
                    // case data.FogTypeExplored, data.FogTypeVisible:
                    //     options.ColorScale = black
                    //     drawFogBorder(tileX, tileY, data.FogTypeUnexplored)

                    case data.FogTypeExplored:
                        options.ColorScale = darkTransparent
                        drawFogTile(tileX, tileY)

                        options.ColorScale = black
                        drawFogBorder(tileX, tileY, data.FogTypeUnexplored)

                    case data.FogTypeVisible:
                        options.ColorScale = lightTransparent
                        drawFogBorder(tileX, tileY, data.FogTypeExplored)

                        options.ColorScale = black
                        drawFogBorder(tileX, tileY, data.FogTypeUnexplored)
                }
            }
        }
    }

}

type Overworld struct {
    Camera camera.Camera
    Counter uint64
    Map *maplib.Map
    Cities []*citylib.City
    CitiesMiniMap []maplib.MiniMapCity
    Stacks []*playerlib.UnitStack
    SelectedStack *playerlib.UnitStack
    MovingStack *playerlib.UnitStack
    ImageCache *util.ImageCache
    Fog data.FogMap
    ShowAnimation bool
    FogBlack *ebiten.Image
}

func (overworld *Overworld) ToCameraCoordinates(x int, y int) (int, int) {
    return overworld.Map.XDistance(overworld.Camera.GetX(), x) + overworld.Camera.GetX(), y
}

func (overworld *Overworld) DrawMinimap(screen *ebiten.Image){
    overworld.Map.DrawMinimap(screen, overworld.CitiesMiniMap, overworld.Camera.GetX(), overworld.Camera.GetY(), overworld.Camera.GetZoom(), overworld.Fog, overworld.Counter, true)
}

// FIXME: pass in an UnscaledGeom here
func (overworld *Overworld) DrawOverworld(screen *ebiten.Image, geom ebiten.GeoM){

    screen.Fill(color.RGBA{R: 32, G: 32, B: 32, A: 0xff})

    tileWidth := overworld.Map.TileWidth()
    tileHeight := overworld.Map.TileHeight()

    geom.Translate(-overworld.Camera.GetZoomedX() * float64(tileWidth), -overworld.Camera.GetZoomedY() * float64(tileHeight))
    geom.Scale(overworld.Camera.GetAnimatedZoom(), overworld.Camera.GetAnimatedZoom())
    // geom.Concat(scale.ScaledGeom)

    overworld.Map.DrawLayer1(overworld.Camera, overworld.Counter / 8, overworld.ImageCache, screen, geom)

    convertTileCoordinates := func(x int, y int) (int, int) {
        outX := x * tileWidth
        outY := y * tileHeight
        return outX, outY
    }

    boundX1, _, boundX2, _ := overworld.Camera.GetTileBounds()

    // all sprites have to be drawn N times to account for the zoom level being high enough
    // that the sprite might be visible on both edges of the map
    mapXOffsets := []int{0}

    // log.Printf("Fit %d maps\n", (boundX2 - boundX1) / overworld.Map.Width())

    // compute how many maps can fit in the given range
    for fit := range ( (boundX2 - boundX1) / overworld.Map.Width() ) {
        n := fit + 1
        // add one width to the right
        mapXOffsets = append(mapXOffsets, n * overworld.Map.Width())
        // and to the left
        mapXOffsets = append(mapXOffsets, -n * overworld.Map.Width())
    }

    // mapXOffsets will contain [0, -mapWidth, +mapWidth, -2*mapWidth, +2*mapWidth, ...]

    cityPositions := make(map[image.Point]struct{})

    for _, city := range overworld.Cities {
        var cityPic *ebiten.Image
        var err error
        cityPositions[image.Point{city.X, city.Y}] = struct{}{}
        cityPic, err = GetCityImage(city, overworld.ImageCache)

        if err == nil {
            // var options ebiten.DrawImageOptions

            cityX, cityY := overworld.ToCameraCoordinates(city.X, city.Y)

            var cityGeom ebiten.GeoM
            cityGeom.Translate(float64(tileWidth / 2.0), float64(tileHeight / 2.0))
            cityGeom.Translate(float64(-cityPic.Bounds().Dx()) / 2.0, float64(-cityPic.Bounds().Dy()) / 2.0)
            cityGeom.Concat(geom)

            var drawOptions ebiten.DrawImageOptions
            for _, offset := range mapXOffsets {
                x, y := convertTileCoordinates(cityX + offset, cityY)
                drawOptions.GeoM.Reset()
                drawOptions.GeoM.Translate(float64(x), float64(y))
                drawOptions.GeoM.Concat(cityGeom)
                scale.DrawScaled(screen, cityPic, &drawOptions)
            }

            /*
            tx, ty := geom.Apply(float64(x), float64(y))
            vector.StrokeRect(screen, float32(tx), float32(ty), float32(cityPic.Bounds().Dx()), float32(cityPic.Bounds().Dy()), 1, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, true)
            vector.DrawFilledCircle(screen, float32(tx) + float32(tileWidth) / 2, float32(ty) + float32(tileHeight) / 2, 2, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, true)
            */
        }
    }

    for _, stack := range overworld.Stacks {
        doDraw := false
        if stack.Leader() == nil {
            continue
        }

        if overworld.Fog[stack.X()][stack.Y()] != data.FogTypeVisible {
            continue
        }

        location := image.Point{stack.X(), stack.Y()}
        _, hasCity := cityPositions[location]

        if stack == overworld.SelectedStack && (overworld.ShowAnimation || overworld.Counter / 55 % 2 == 0) {
            doDraw = true
        } else if stack == overworld.MovingStack {
            doDraw = true
        } else if stack != overworld.SelectedStack && !hasCity {
            doDraw = true
        }

        if doDraw {
            var options ebiten.DrawImageOptions
            // options.GeoM = geom

            /*
            stackX := float64(overworld.Map.XDistance(overworld.Camera.GetX(), stack.X()) + overworld.Camera.GetX())
            stackY := float64(stack.Y())
            */
            for _, offset := range mapXOffsets {
                options.GeoM.Reset()
                stackX, stackY := overworld.ToCameraCoordinates(stack.X(), stack.Y())

                // log.Printf("World %v, %v -> camera %v, %v. Camera: %v, %v", stack.X(), stack.Y(), stackX, stackY, overworld.Camera.GetX(), overworld.Camera.GetY())

                // x, y := convertTileCoordinates(stackX, stackY)
                x, y := float64(stackX), float64(stackY)

                // nx := overworld.Map.WrapX(x - overworld.Camera.GetX()) + overworld.Camera.GetX() + 6

                options.GeoM.Translate((x + float64(offset) + float64(stack.OffsetX())) * float64(tileWidth), (y + float64(stack.OffsetY())) * float64(tileHeight))

                leader := stack.Leader()

                unitBack, err := units.GetUnitBackgroundImage(leader.GetBanner(), overworld.ImageCache)
                if err == nil {
                    saveGeom := options.GeoM
                    options.GeoM.Concat(geom)
                    scale.DrawScaled(screen, unitBack, &options)
                    options.GeoM = saveGeom
                }

                pic, err := unitview.GetUnitOverworldImage(overworld.ImageCache, leader)
                if err == nil {
                    // screen scale is already taken into account, so we can translate by 1 pixel here
                    options.GeoM.Translate(1, 1)
                    options.GeoM.Concat(geom)

                    if leader.GetBusy() != units.BusyStatusNone {
                        var patrolOptions colorm.DrawImageOptions
                        var matrix colorm.ColorM
                        patrolOptions.GeoM = scale.ScaleGeom(options.GeoM)
                        matrix.ChangeHSV(0, 0, 1)
                        colorm.DrawImage(screen, pic, matrix, &patrolOptions)
                    } else {
                        scale.DrawScaled(screen, pic, &options)
                    }

                    enchantment := util.First(leader.GetEnchantments(), data.UnitEnchantmentNone)
                    if enchantment != data.UnitEnchantmentNone {
                        util.DrawOutline(screen, overworld.ImageCache, pic, scale.ScaleGeom(options.GeoM), options.ColorScale, overworld.Counter/8, enchantment.Color())
                    }
                }
            }
        }
    }

    overworld.Map.DrawLayer2(overworld.Camera, overworld.Counter / 8, overworld.ImageCache, screen, geom)

    if overworld.Fog != nil {
        overworld.DrawFog(screen, geom)
    }

    // draw current path on top of fog
    if overworld.SelectedStack != nil && len(overworld.SelectedStack.CurrentPath) > 0 {
        boot, _ := overworld.ImageCache.GetImage("compix.lbx", 72, 0)
        var options ebiten.DrawImageOptions

        var bootGeom ebiten.GeoM
        bootGeom.Translate(float64(tileWidth) / 2, float64(tileHeight) / 2)
        bootGeom.Translate(float64(boot.Bounds().Dx()) / -2, float64(boot.Bounds().Dy()) / -2)
        bootGeom.Concat(geom)

        for pointI, point := range overworld.SelectedStack.CurrentPath {
            for _, offset := range mapXOffsets {
                cx, cy := overworld.ToCameraCoordinates(point.X, point.Y)
                x, y := convertTileCoordinates(cx + offset, cy)
                options.GeoM.Reset()
                options.ColorScale.Reset()
                options.GeoM.Translate(float64(x), float64(y))
                options.GeoM.Concat(bootGeom)

                v := float32(1 + (math.Sin(float64(overworld.Counter * 4 + uint64(pointI) * 60) * math.Pi / 180) / 2 + 0.5) / 2)
                options.ColorScale.Scale(v, v, v, 1)

                scale.DrawScaled(screen, boot, &options)
            }
        }
    }


    /*
    for i := range int(200.0) {
        // y := int(float64(i) * float64(tileHeight) * overworld.Camera.GetZoom())
        _, y := geom.Apply(0, float64(i * tileHeight))
        vector.StrokeLine(screen, float32(0), float32(y), float32(320), float32(y), 1, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, false)
    }

    for i := range int(200) {
        // x := int(float64(i) * float64(tileWidth) * overworld.Camera.GetZoom() - 6.0/overworld.Camera.GetZoom()*float64(tileWidth))
        // x := int(float64(i) * float64(tileWidth) * overworld.Camera.GetZoom())
        // x := int((float64(i) - 6.0/overworld.Camera.GetZoom()) * float64(tileWidth) * overworld.Camera.GetZoom())
        // v1 := 6.0/overworld.Camera.GetZoom() * float64(tileWidth)
        // x := float32((float64(i) * float64(tileWidth) - v1) * overworld.Camera.GetZoom())
        x, _ := geom.Apply(float64(i * tileWidth), 0)
        vector.StrokeLine(screen, float32(x), float32(0), float32(x), float32(200), 1, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, false)
    }
    */
}

func (game *Game) Draw(screen *ebiten.Image){
    game.Drawer(screen, game)
}

func (game *Game) DrawGame(screen *ebiten.Image){

    var cities []*citylib.City
    var citiesMiniMap []maplib.MiniMapCity
    var stacks []*playerlib.UnitStack
    var selectedStack *playerlib.UnitStack
    var fog data.FogMap

    for i, player := range game.Players {
        for _, city := range player.Cities {
            if city.Plane == game.Plane {
                cities = append(cities, city)
                citiesMiniMap = append(citiesMiniMap, city)
            }
        }

        for _, stack := range player.Stacks {
            if stack.Plane() == game.Plane {
                stacks = append(stacks, stack)
            }
        }

        /*
        for _, unit := range player.Units {
            if unit.Plane == game.Plane {
                units = append(units, unit)
            }
        }
        */

        if i == 0 {
            selectedStack = player.SelectedStack
            fog = player.GetFog(game.Plane)
        }
    }

    useCounter := game.Counter
    /*
    if data.ScreenScale == 1 && game.Camera.GetZoom() < 0.9 {
        useCounter = 1
    }
    */

    overworld := Overworld{
        Camera: game.Camera,
        Counter: useCounter,
        Map: game.CurrentMap(),
        Cities: cities,
        CitiesMiniMap: citiesMiniMap,
        Stacks: stacks,
        SelectedStack: selectedStack,
        MovingStack: game.MovingStack,
        ImageCache: &game.ImageCache,
        Fog: fog,
        ShowAnimation: game.State == GameStateUnitMoving,
        FogBlack: game.GetFogImage(),
    }

    overworldScreen := screen.SubImage(image.Rect(0, scale.Scale(18), scale.Scale(240), scale.Scale(data.ScreenHeight))).(*ebiten.Image)
    overworld.DrawOverworld(overworldScreen, ebiten.GeoM{})

    mini := screen.SubImage(game.GetMinimapRect()).(*ebiten.Image)
    if mini.Bounds().Dx() > 0 {
        overworld.DrawMinimap(mini)
    }

    // test of TileToScreen
    /*
    mouseX, mouseY := inputmanager.MousePosition()
    px, py := game.TileToScreen(game.ScreenToTile(float64(mouseX), float64(mouseY)))
    vector.DrawFilledCircle(screen, float32(px), float32(py), 2, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, true)
    */

    game.HudUI.Draw(game.HudUI, screen)
}

func (game *Game) GetMinimapRect() image.Rectangle {
    var miniGeom ebiten.GeoM
    miniGeom.Translate(scale.Scale2(250.0, 20.0))
    mx, my := miniGeom.Apply(0, 0)
    miniWidth := scale.Scale(60)
    miniHeight := scale.Scale(31)
    return image.Rect(int(mx), int(my), int(mx) + miniWidth, int(my) + miniHeight)
}

func (game *Game) GetAllGlobalEnchantments() map[data.BannerType]*set.Set[data.Enchantment] {
    enchantments := make(map[data.BannerType]*set.Set[data.Enchantment])
    for _, player := range game.Players {
        enchantments[player.GetBanner()] = player.GlobalEnchantments.Clone()
    }
    return enchantments
}

func (game *Game) CastingDetectableByHuman(caster *playerlib.Player) bool {
    for _, player := range game.Players {
        if player.IsHuman() {
            if player.GlobalEnchantments.Contains(data.EnchantmentDetectMagic) {
                for _, enemy := range player.GetKnownPlayers() {
                    if enemy == caster {
                        return true
                    }
                }
            }
        }
    }
    return false
}

func (game *Game) RelocateUnit(player *playerlib.Player, unit units.StackUnit) {
    summonCity := player.FindSummoningCity()
    if summonCity == nil {
        return
    }

    player.UpdateUnitLocation(unit, summonCity.X, summonCity.Y, summonCity.Plane)

    allStacks := player.FindAllStacks(summonCity.X, summonCity.Y, summonCity.Plane)
    for i := 1; i < len(allStacks); i++ {
        player.MergeStacks(allStacks[0], allStacks[i])
    }

    game.ResolveStackAt(summonCity.X, summonCity.Y, summonCity.Plane)

    unit.SetBusy(units.BusyStatusNone)
}
