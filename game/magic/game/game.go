package game

import (
    "image/color"
    "image"
    "math/rand/v2"
    "log"
    "math"
    "fmt"
    "strings"
    "slices"

    "github.com/kazzmir/master-of-magic/game/magic/ai"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
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
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/cityview"
    "github.com/kazzmir/master-of-magic/game/magic/armyview"
    "github.com/kazzmir/master-of-magic/game/magic/citylistview"
    "github.com/kazzmir/master-of-magic/game/magic/magicview"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/summon"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/music"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    mouselib "github.com/kazzmir/master-of-magic/lib/mouse"
    helplib "github.com/kazzmir/master-of-magic/game/magic/help"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/functional"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/colorm"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

func (game *Game) GetFogImage() *ebiten.Image {
    if game.Fog != nil {
        return game.Fog
    }

    game.Fog = ebiten.NewImage(game.CurrentMap().TileWidth(), game.CurrentMap().TileHeight())
    game.Fog.Fill(color.RGBA{R: 8, G: 4, B: 4, A: 0xff})
    return game.Fog
}

type GameEvent interface {
}

type GameEventMagicView struct {
}

type GameEventArmyView struct {
}

type GameEventRefreshUI struct {
}

type GameEventSurveyor struct {
}

type GameEventNextTurn struct {
}

type GameEventCityListView struct {
}

type GameEventApprenticeUI struct {
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
}

type GameEventNewOutpost struct {
    City *citylib.City
    Stack *playerlib.UnitStack
    Player *playerlib.Player
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
    Wizard data.WizardBase
}

type GameEventSummonHero struct {
    Wizard data.WizardBase
    Champion bool
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

type GameState int
const (
    GameStateRunning GameState = iota
    GameStateUnitMoving
    GameStateQuit
)

type Game struct {
    Cache *lbx.LbxCache
    ImageCache util.ImageCache
    WhiteFont *font.Font

    Music *music.Music

    Settings setup.NewGameSettings

    InfoFontYellow *font.Font
    InfoFontRed *font.Font
    Counter uint64
    Fog *ebiten.Image
    Drawer func (*ebiten.Image, *Game)
    State GameState
    Plane data.Plane

    TurnNumber uint64

    Heroes map[herolib.HeroType]*herolib.Hero

    ArtifactPool map[string]*artifact.Artifact

    MouseData *mouselib.MouseData

    Events chan GameEvent
    BuildingInfo buildinglib.BuildingInfos

    MovingStack *playerlib.UnitStack

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

    Camera camera.Camera
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

/* for each book type, there are X number of spells that can be researched per rarity type.
 * for example, books=3 yields 6 common, 3 uncommon, 2 rare, 1 very rare
 */
func (game *Game) InitializeResearchableSpells(spells *spellbook.Spells, player *playerlib.Player){
    commonCount := func(books int) int {
        if books == 1 {
            return 3
        }

        return int(math.Min(10, float64(3 + books)))
    }

    uncommonCount := func(books int) int {
        if books <= 6 {
            return books
        }

        if books == 7 {
            return 8
        }

        if books == 8 {
            return 10
        }

        return 10
    }

    rareCount := func(books int) int {
        if books == 1 {
            return 0
        }

        if books <= 8 {
            return books - 1
        }

        return int(math.Min(10, float64(books)))
    }

    veryRareCount := func(books int) int {
        if books <= 2 {
            return 0
        }

        if books <= 10 {
            return books - 2
        }

        return 10
    }

    type CountFunc func(int) int

    rarityCount := make(map[spellbook.SpellRarity]CountFunc)
    rarityCount[spellbook.SpellRarityCommon] = commonCount
    rarityCount[spellbook.SpellRarityUncommon] = uncommonCount
    rarityCount[spellbook.SpellRarityRare] = rareCount
    rarityCount[spellbook.SpellRarityVeryRare] = veryRareCount

    for _, book := range player.Wizard.Books {
        realmSpells := spells.GetSpellsByMagic(book.Magic)

        for rarity, countFunc := range rarityCount {
            raritySpells := realmSpells.GetSpellsByRarity(rarity)

            alreadyKnown := player.KnownSpells.GetSpellsByMagic(book.Magic).GetSpellsByRarity(rarity)

            raritySpells.RemoveSpells(alreadyKnown)
            raritySpells.ShuffleSpells()

            // if the player can research 6 spells but already has 3 selected, then they can research 3 more
            for i := 0; i < countFunc(book.Count) - len(alreadyKnown.Spells); i++ {
                player.ResearchPoolSpells.AddSpell(raritySpells.Spells[i])
            }
        }
    }
}

func (game *Game) AddPlayer(wizard setup.WizardCustom, human bool) *playerlib.Player{
    newPlayer := playerlib.MakePlayer(wizard, human, game.CurrentMap().Width(), game.CurrentMap().Height())

    if !human {
        newPlayer.AIBehavior = ai.MakeEnemyAI()
        newPlayer.StrategicCombat = true
    }

    allSpells := game.AllSpells()

    startingSpells := []string{"Magic Spirit", "Spell of Return"}
    if wizard.AbilityEnabled(setup.AbilityArtificer) {
        startingSpells = append(startingSpells, "Enchant Item", "Create Artifact")
    }

    newPlayer.ResearchPoolSpells = wizard.StartingSpells.Copy()
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

    game.InitializeResearchableSpells(&allSpells, newPlayer)
    newPlayer.UpdateResearchCandidates()

    // log.Printf("Research spells: %v", newPlayer.ResearchPoolSpells)

    // famous wizards get a head start of 10 fame
    if wizard.AbilityEnabled(setup.AbilityFamous) {
        newPlayer.Fame += 10
    }

    game.Players = append(game.Players, newPlayer)
    return newPlayer
}

func createHeroes() map[herolib.HeroType]*herolib.Hero {
    heroes := make(map[herolib.HeroType]*herolib.Hero)

    for _, heroType := range herolib.AllHeroTypes() {
        hero := herolib.MakeHeroSimple(heroType)
        hero.SetExtraAbilities()
        heroes[heroType] = hero
    }

    return heroes
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

    helpLbx, err := lbxCache.GetLbxFile("help.lbx")
    if err != nil {
        return nil
    }

    help, err := helplib.ReadHelp(helpLbx, 2)
    if err != nil {
        return nil
    }

    fontLbx, err := lbxCache.GetLbxFile("FONTS.LBX")
    if err != nil {
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        return nil
    }

    orange := color.RGBA{R: 0xc7, G: 0x82, B: 0x1b, A: 0xff}

    yellowPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        orange,
        orange,
        orange,
        orange,
        orange,
        orange,
    }

    infoFontYellow := font.MakeOptimizedFontWithPalette(fonts[0], yellowPalette)

    red := color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}
    redPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        red, red, red,
        red, red, red,
    }

    infoFontRed := font.MakeOptimizedFontWithPalette(fonts[0], redPalette)

    whitePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.White, color.White, color.White, color.White,
    }

    whiteFont := font.MakeOptimizedFontWithPalette(fonts[0], whitePalette)

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
        InfoFontYellow: infoFontYellow,
        InfoFontRed: infoFontRed,
        Heroes: createHeroes(),
        ArtifactPool: createArtifactPool(lbxCache),
        WhiteFont: whiteFont,
        BuildingInfo: buildingInfo,
        TurnNumber: 1,
        CurrentPlayer: -1,
        Camera: camera.MakeCamera(),

        RoadWorkArcanus: make(map[image.Point]float64),
        RoadWorkMyrror: make(map[image.Point]float64),

        PurifyWorkArcanus: make(map[image.Point]float64),
        PurifyWorkMyrror: make(map[image.Point]float64),
    }

    game.ArcanusMap = maplib.MakeMap(terrainData, settings.LandSize, settings.Magic, settings.Difficulty, data.PlaneArcanus, game)
    game.MyrrorMap = maplib.MakeMap(terrainData, settings.LandSize, settings.Magic, settings.Difficulty, data.PlaneMyrror, game)

    game.HudUI = game.MakeHudUI()
    game.Drawer = func(screen *ebiten.Image, game *Game){
        game.DrawGame(screen)
    }

    game.Music.PushSong(music.SongBackground1)

    return game
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

func (game *Game) FindCity(x int, y int, plane data.Plane) *citylib.City {
    for _, player := range game.Players {
        city := player.FindCity(x, y, plane)
        if city != nil {
            return city
        }
    }

    return nil
}

func (game *Game) ContainsCity(x int, y int, plane data.Plane) bool {
    return game.FindCity(x, y, plane) != nil
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

func (game *Game) FindValidCityLocation(plane data.Plane) (int, int, bool) {
    mapUse := game.GetMap(plane)
    continents := mapUse.Map.FindContinents()

    for i := 0; i < 10; i++ {
        continentIndex := rand.IntN(len(continents))
        continent := continents[continentIndex]
        if len(continent) > 100 {
            index := rand.IntN(len(continent))
            x := continent[index].X
            y := continent[index].Y

            tile := terrain.GetTile(mapUse.Map.Terrain[x][y])
            if y > 3 && y < mapUse.Map.Columns() - 3 && tile.IsLand() && !tile.IsMagic() && mapUse.GetEncounter(x, y) == nil && !game.ContainsCity(x, y, plane) {
                return x, y, true
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
            for _, index := range rand.Perm(continent.Size()) {
                tile := mapUse.GetTile(continent[index].X, continent[index].Y)
                if tile.Tile.IsLand() && !tile.Tile.IsMagic() {
                    return continent[index].X, continent[index].Y
                }
            }
        }
    }

    return 0, 0
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

    drawMinimap := func (screen *ebiten.Image, x int, y int, fog data.FogMap, counter uint64){
        game.CurrentMap().DrawMinimap(screen, citiesMiniMap, x, y, 1, fog, counter, false)
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

    game.HudUI = game.MakeHudUI()

    // absorb most recent left click
    yield()
}

/* how much power the player has.
 * add up all melded node tiles, all buildings that produce power, etc
 */
func (game *Game) ComputePower(player *playerlib.Player) int {
    power := float64(0)

    for _, city := range player.Cities {
        power += float64(city.ComputePower(player.Wizard.TotalBooks()))
    }

    magicBonus := float64(1)

    switch game.Settings.Magic {
        case data.MagicSettingWeak: magicBonus = 0.5
        case data.MagicSettingNormal: magicBonus = 1
        case data.MagicSettingPowerful: magicBonus = 1.5
    }

    for _, node := range game.ArcanusMap.GetMeldedNodes(player) {
        power += node.GetPower(magicBonus)
    }

    for _, node := range game.MyrrorMap.GetMeldedNodes(player) {
        power += node.GetPower(magicBonus)
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

func (game *Game) doMagicView(yield coroutine.YieldFunc) {

    oldDrawer := game.Drawer
    magicScreen := magicview.MakeMagicScreen(game.Cache, game.Players[0], game.Players[0].GetKnownPlayers(), game.ComputePower(game.Players[0]))

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

func (game *Game) doInput(yield coroutine.YieldFunc, title string, name string, topX int, topY int) string {
    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    fontLbx, err := game.Cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return ""
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return ""
    }

    bluish := color.RGBA{R: 0xcf, G: 0xef, B: 0xf9, A: 0xff}
    // red := color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}
    namePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        util.Lighten(bluish, -30),
        util.Lighten(bluish, -20),
        util.Lighten(bluish, -10),
        util.Lighten(bluish, 0),
    }

    orange := color.RGBA{R: 0xed, G: 0xa7, B: 0x12, A: 0xff}
    titlePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        util.Lighten(orange, -30),
        util.Lighten(orange, -20),
        util.Lighten(orange, -10),
        util.Lighten(orange, 0),
    }

    maxLength := float64(84 * data.ScreenScale)

    nameFont := font.MakeOptimizedFontWithPalette(fonts[4], namePalette)

    titleFont := font.MakeOptimizedFontWithPalette(fonts[4], titlePalette)

    quit := false

    source := ebiten.NewImage(1, 1)
    source.Fill(color.RGBA{R: 0xcf, G: 0xef, B: 0xf9, A: 0xff})

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })
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

            for len(name) > 0 && nameFont.MeasureTextWidth(name, float64(data.ScreenScale)) > maxLength {
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
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := game.ImageCache.GetImage("backgrnd.lbx", 33, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(topX * data.ScreenScale), float64(topY * data.ScreenScale))
            screen.DrawImage(background, &options)

            x, y := options.GeoM.Apply(float64(13 * data.ScreenScale), float64(20 * data.ScreenScale))

            nameFont.Print(screen, x, y, float64(data.ScreenScale), options.ColorScale, name)

            tx, ty := options.GeoM.Apply(float64(9 * data.ScreenScale), float64(6 * data.ScreenScale))
            titleFont.Print(screen, tx, ty, float64(data.ScreenScale), options.ColorScale, title)

            // draw cursor
            cursorX := x + nameFont.MeasureTextWidth(name, float64(data.ScreenScale))

            util.DrawTextCursor(screen, source, cursorX, y, game.Counter)
        },
    }

    ui.AddElement(input)
    ui.FocusElement(input, name)

    for !quit {
        game.Counter += 1
        ui.StandardUpdate()
        yield()
    }

    return name
}

func (game *Game) showNewBuilding(yield coroutine.YieldFunc, city *citylib.City, building buildinglib.Building, player *playerlib.Player){
    drawer := game.Drawer
    defer func(){
        game.Drawer = drawer
    }()

    fontLbx, err := game.Cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return
    }

    yellow := color.RGBA{R: 0xea, G: 0xb6, B: 0x00, A: 0xff}
    yellowPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        yellow, yellow, yellow,
        yellow, yellow, yellow,
        yellow, yellow, yellow,
        yellow, yellow, yellow,
    }

    bigFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowPalette)

    background, _ := game.ImageCache.GetImage("resource.lbx", 40, 0)

    // devil: 51
    // cat: 52
    // bird: 53
    // snake: 54
    // beetle: 55
    animalIndex := 51
    switch player.Wizard.MostBooks() {
        case data.NatureMagic: animalIndex = 54
        case data.SorceryMagic: animalIndex = 55
        case data.ChaosMagic: animalIndex = 51
        case data.LifeMagic: animalIndex = 53
        case data.DeathMagic: animalIndex = 52
    }
    animal, _ := game.ImageCache.GetImageTransform("resource.lbx", animalIndex, 0, "crop", util.AutoCrop)

    wrappedText := bigFont.CreateWrappedText(float64(175 * data.ScreenScale), float64(1 * data.ScreenScale), fmt.Sprintf("The %s of %s has completed the construction of a %s.", city.GetSize(), city.Name, game.BuildingInfo.Name(building)))

    rightSide, _ := game.ImageCache.GetImage("resource.lbx", 41, 0)

    getAlpha := util.MakeFadeIn(7, &game.Counter)

    buildingPics, err := game.ImageCache.GetImagesTransform("cityscap.lbx", buildinglib.GetBuildingIndex(building), "crop", util.AutoCrop)

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
        options.GeoM.Translate(float64(8 * data.ScreenScale), float64(60 * data.ScreenScale))
        screen.DrawImage(background, &options)
        iconOptions := options
        iconOptions.GeoM.Translate(float64(6 * data.ScreenScale), float64(-10 * data.ScreenScale))
        screen.DrawImage(animal, &iconOptions)

        x, y := options.GeoM.Apply(float64(8 * data.ScreenScale + animal.Bounds().Dx()), float64(9 * data.ScreenScale))
        bigFont.RenderWrapped(screen, x, y, wrappedText, options.ColorScale, false)

        options.GeoM.Translate(float64(background.Bounds().Dx()), 0)
        screen.DrawImage(rightSide, &options)

        x, y = options.GeoM.Apply(float64(4 * data.ScreenScale), float64(6 * data.ScreenScale))
        buildingSpace := screen.SubImage(image.Rect(int(x), int(y), int(x) + 45 * data.ScreenScale, int(y) + 47 * data.ScreenScale)).(*ebiten.Image)

        // buildingSpace.Fill(color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff})
        // vector.DrawFilledRect(buildingSpace, float32(x), float32(y), float32(buildingSpace.Bounds().Dx()), float32(buildingSpace.Bounds().Dy()), color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, false)

        landOptions := options
        landOptions.GeoM.Translate(float64(-10 * data.ScreenScale), float64(-10 * data.ScreenScale))
        buildingSpace.DrawImage(landBackground, &landOptions)

        buildingOptions := options
        // translate to the center of the building space, and then draw the image centered by translating
        // by -width/2, -height/2
        buildingOptions.GeoM.Reset()
        buildingOptions.GeoM.Translate(x, y)
        buildingOptions.GeoM.Translate(float64(buildingSpace.Bounds().Dx()/2), float64(buildingSpace.Bounds().Dy()) / 2)
        buildingOptions.GeoM.Translate(float64(buildingPicsAnimation.Frame().Bounds().Dx()) / -2, float64(buildingPicsAnimation.Frame().Bounds().Dy()) / -2)
        buildingSpace.DrawImage(buildingPicsAnimation.Frame(), &buildingOptions)
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

func (game *Game) showScroll(yield coroutine.YieldFunc, title string, text string){
    fontLbx, err := game.Cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return
    }

    red := util.Lighten(color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}, -60)
    redPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        red, red, red,
        red, red, red,
    }

    red2 := util.Lighten(color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}, -80)
    redPalette2 := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        red2, red2, red2,
        red2, red2, red2,
    }

    bigFont := font.MakeOptimizedFontWithPalette(fonts[4], redPalette)

    smallFont := font.MakeOptimizedFontWithPalette(fonts[1], redPalette2)
    wrappedText := smallFont.CreateWrappedText(float64(180 * data.ScreenScale), float64(data.ScreenScale), text)

    scrollImages, _ := game.ImageCache.GetImages("scroll.lbx", 2)

    totalImages := int((wrappedText.TotalHeight + float64(bigFont.Height() * data.ScreenScale)) / float64(5 * data.ScreenScale)) + 1

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

    scrollLength := 30 * data.ScreenScale

    getAlpha := util.MakeFadeIn(7, &game.Counter)

    game.Drawer = func (screen *ebiten.Image, game *Game){
        drawer(screen, game)

        var options ebiten.DrawImageOptions
        options.ColorScale.ScaleAlpha(getAlpha())

        options.GeoM.Translate(float64(65 * data.ScreenScale), float64(25 * data.ScreenScale))

        middleY := pageBackground.Bounds().Dy() / 2
        length := scrollLength / 2
        if length > middleY {
            length = middleY
        }
        pagePart := pageBackground.SubImage(image.Rect(0, middleY - length, pageBackground.Bounds().Dx(), middleY + length)).(*ebiten.Image)

        pageOptions := options
        pageOptions.GeoM.Translate(0, float64(middleY - length) + float64(5 * data.ScreenScale))
        screen.DrawImage(pagePart, &pageOptions)

        x, y := options.GeoM.Apply(float64(pageBackground.Bounds().Dx()) / 2, float64(middleY) - wrappedText.TotalHeight / 2 - float64(bigFont.Height() * data.ScreenScale) / 2 + 5)
        bigFont.PrintCenter(screen, x, y, float64(data.ScreenScale), options.ColorScale, title)
        y += float64(bigFont.Height() * data.ScreenScale) + 1
        smallFont.RenderWrapped(screen, x, y, wrappedText, options.ColorScale, true)

        scrollOptions := options
        scrollOptions.GeoM.Translate(float64(-63 * data.ScreenScale), float64(-20 * data.ScreenScale))
        screen.DrawImage(scrollAnimation.Frame(), &scrollOptions)
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
                scrollLength += 10 * data.ScreenScale
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
                scrollLength -= 10 * data.ScreenScale
            } else {
                quit = true
            }
        }

        yield()
    }

    // fade out
    getAlpha = util.MakeFadeOut(7, &game.Counter)
    for i := 0; i < 7; i++ {
        game.Counter += 1
        yield()
    }
}

func (game *Game) showOutpost(yield coroutine.YieldFunc, city *citylib.City, stack *playerlib.UnitStack, rename bool){
    drawer := game.Drawer
    defer func(){
        game.Drawer = drawer
    }()

    fontLbx, err := game.Cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return
    }

    // red := color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}
    yellow := util.Lighten(util.RotateHue(color.RGBA{R: 0xff, G: 0xff, B: 0x00, A: 0xff}, -0.60), 0)
    yellowPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        yellow,
        util.Lighten(yellow, -20),
        util.Lighten(yellow, -20),
        util.Lighten(yellow, -15),
        util.Lighten(yellow, -30),
        util.Lighten(yellow, -10),
        util.Lighten(yellow, -15),
        util.Lighten(yellow, -10),
        util.Lighten(yellow, -10),
        util.Lighten(yellow, -35),
        util.Lighten(yellow, -45),
        yellow,
        yellow,
        yellow,
    }

    bigFont := font.MakeOptimizedFontWithPalette(fonts[5], yellowPalette)

    game.Drawer = func (screen *ebiten.Image, game *Game){
        drawer(screen, game)

        background, _ := game.ImageCache.GetImage("backgrnd.lbx", 32, 0)

        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(30 * data.ScreenScale), float64(50 * data.ScreenScale))
        screen.DrawImage(background, &options)

        numHouses := city.GetOutpostHouses()
        maxHouses := 10

        houseOptions := options
        houseOptions.GeoM.Translate(float64(7 * data.ScreenScale), float64(31 * data.ScreenScale))

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
            screen.DrawImage(house, &houseOptions)
            houseOptions.GeoM.Translate(float64(house.Bounds().Dx()) + 1, 0)
        }

        emptyHouse, _ := game.ImageCache.GetImage("backgrnd.lbx", emptyHouseIndex, 0)
        for i := numHouses; i < maxHouses; i++ {
            screen.DrawImage(emptyHouse, &houseOptions)
            houseOptions.GeoM.Translate(float64(emptyHouse.Bounds().Dx() + 1 * data.ScreenScale), 0)
        }

        if stack != nil {
            stackOptions := options
            stackOptions.GeoM.Translate(float64(7 * data.ScreenScale), float64(55 * data.ScreenScale))

            for _, unit := range stack.Units() {
                pic, _ := GetUnitImage(unit, &game.ImageCache, city.Banner)
                screen.DrawImage(pic, &stackOptions)
                stackOptions.GeoM.Translate(float64(pic.Bounds().Dx() + 1 * data.ScreenScale), 0)
            }
        }

        x, y := options.GeoM.Apply(float64(6 * data.ScreenScale), float64(22 * data.ScreenScale))
        game.InfoFontYellow.Print(screen, x, y, float64(data.ScreenScale), options.ColorScale, city.Race.String())

        x, y = options.GeoM.Apply(float64(20 * data.ScreenScale), float64(5 * data.ScreenScale))
        if rename {
            bigFont.Print(screen, x, y, float64(data.ScreenScale), options.ColorScale, "New Outpost Founded")
        } else {
            bigFont.Print(screen, x, y, float64(data.ScreenScale), options.ColorScale, fmt.Sprintf("Outpost Of %v", city.Name))
        }

        cityScapeOptions := options
        cityScapeOptions.GeoM.Translate(float64(185 * data.ScreenScale), float64(30 * data.ScreenScale))
        x, y = cityScapeOptions.GeoM.Apply(0, 0)
        cityScape := screen.SubImage(image.Rect(int(x), int(y), int(x) + 72 * data.ScreenScale, int(y) + 66 * data.ScreenScale)).(*ebiten.Image)

        cityScapeBackground, _ := game.ImageCache.GetImage("cityscap.lbx", 0, 0)
        cityScape.DrawImage(cityScapeBackground, &cityScapeOptions)

        // regular house
        houseIndex := 25

        switch city.Race {
            case data.RaceDarkElf, data.RaceHighElf: houseIndex = 30
            case data.RaceGnoll, data.RaceKlackon, data.RaceLizard, data.RaceTroll: houseIndex = 35
        }

        cityHouse, _ := game.ImageCache.GetImage("cityscap.lbx", houseIndex, 0)
        options2 := cityScapeOptions
        options2.GeoM.Translate(float64(30 * data.ScreenScale), float64(20 * data.ScreenScale))
        cityScape.DrawImage(cityHouse, &options2)

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
    }

    quit := false
    for !quit {
        if inputmanager.LeftClick() {
            quit = true
        }

        yield()
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
func (game *Game) ComputeTerrainCost(stack *playerlib.UnitStack, sourceX int, sourceY int, destX int, destY int, mapUse *maplib.Map) (fraction.Fraction, bool) {
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

    // can't move from land to ocean unless all units are flyers or swimmers
    if tileFrom.Tile.IsLand() && !tileTo.Tile.IsLand() {
        if !stack.AllFlyers() && !stack.AllSwimmers() {
            return fraction.Zero(), false
        }
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

    if stack.AllFlyers() {
        return fraction.FromInt(1), true
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

        vector.DrawFilledRect(screen, 0, 0, float32(screen.Bounds().Dx()), float32(screen.Bounds().Dy()), scale.Apply(color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}), false)
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

func (game *Game) FindPath(oldX int, oldY int, newX int, newY int, stack *playerlib.UnitStack, fog data.FogMap) pathfinding.Path {

    useMap := game.GetMap(stack.Plane())

    if newY < 0 || newY >= useMap.Height() {
        return nil
    }

    normalized := func (a image.Point) image.Point {
        return image.Pt(useMap.WrapX(a.X), a.Y)
    }

    // check equality of two points taking wrapping into account
    tileEqual := func (a image.Point, b image.Point) bool {
        return normalized(a) == normalized(b)
    }

    // cache the containsEnemy result
    // true if the given coordinates contain an enemy unit or city
    containsEnemy := functional.Memoize2(func (x int, y int) bool {
        for _, player := range game.Players {
            if player.GetBanner() != stack.GetBanner() {
                enemyStack := player.FindStack(x, y, stack.Plane())
                if enemyStack != nil {
                    return true
                }

                enemyCity := player.FindCity(x, y, stack.Plane())
                if enemyCity != nil {
                    return true
                }
            }
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

        /*
        tileFrom := useMap.GetTile(x1, y1)
        tileTo := useMap.GetTile(x2, y2)
        */

        // FIXME: consider terrain type, roads, and unit abilities

        baseCost := float64(1)

        if x1 != x2 && y1 != y2 {
            baseCost = 1.5
        }

        // don't know what the cost is, assume we can move there
        // FIXME: how should this behave for different FogTypes?
        if x2 >= 0 && x2 < len(fog) && y2 >= 0 && y2 < len(fog[x2]) && fog[x2][y2] == data.FogTypeUnexplored {
            // increase cost of unknown tile very slightly so we prefer to move to known tiles
            return baseCost + 0.1
        }

        cost, ok := game.ComputeTerrainCost(stack, x1, y1, x2, y2, useMap)
        if !ok {
            return pathfinding.Infinity
        }

        return cost.ToFloat()

        // can't move from land to ocean unless all units are flyers
        /*
        if tileFrom.Tile.IsLand() && !tileTo.Tile.IsLand() {
            if !stack.AllFlyers() {
                return pathfinding.Infinity
            }
        }

        return baseCost
        */
    }

    neighbors := func (x int, y int) []image.Point {
        out := make([]image.Point, 0, 8)

        // up left
        if y > 0 {
            out = append(out, image.Pt(x - 1, y - 1))
        }

        // left
        out = append(out, image.Pt(x - 1, y))

        // down left
        if y < useMap.Height() - 1 {
            out = append(out, image.Pt(x - 1, y + 1))
        }

        // up right
        if y > 0 {
            out = append(out, image.Pt(x + 1, y - 1))
        }

        // up
        if y > 0 {
            out = append(out, image.Pt(x, y - 1))
        }

        // right
        out = append(out, image.Pt(x + 1, y))

        // down right
        if y < useMap.Height() - 1 {
            out = append(out, image.Pt(x + 1, y + 1))
        }

        // down
        if y < useMap.Height() - 1 {
            out = append(out, image.Pt(x, y + 1))
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
    fontLbx, err := game.Cache.GetLbxFile("FONTS.LBX")
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }

    bluish := util.Lighten(color.RGBA{R: 0x00, G: 0x00, B: 0xff, A: 0xff}, 90)

    optionPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        bluish, bluish, bluish, bluish,
        bluish, bluish, bluish, bluish,
    }

    optionFont := font.MakeOptimizedFontWithPalette(fonts[2], optionPalette)

    var elements []*uilib.UIElement

    var makeElements func()

    makeElements = func() {
        *background, _ = imageCache.GetImage("load.lbx", 11, 0)
        ok, _ := imageCache.GetImage("load.lbx", 4, 0)
        ui.RemoveElements(elements)
        elements = nil

        elements = append(elements, &uilib.UIElement{
            Rect: util.ImageRect(266 * data.ScreenScale, 176 * data.ScreenScale, ok),
            LeftClick: func(element *uilib.UIElement){
                onOk()
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))
                screen.DrawImage(ok, &options)
            },
        })

        resolutionBackground, _ := imageCache.GetImage("load.lbx", 5, 0)

        elements = append(elements, &uilib.UIElement{
            Rect: util.ImageRect(20 * data.ScreenScale, 40 * data.ScreenScale, resolutionBackground),
            LeftClick: func(element *uilib.UIElement){
                selected := func(name string, scale int, algorithm data.ScaleAlgorithm) string {
                    if data.ScreenScale == scale && data.ScreenScaleAlgorithm == algorithm {
                        return name + "*"
                    }
                    return name
                }

                update := func(scale int, algorithm data.ScaleAlgorithm){
                    data.ScreenScale = scale
                    data.ScreenScaleAlgorithm = algorithm
                    data.ScreenWidth = data.ScreenWidthOriginal * scale
                    data.ScreenHeight = data.ScreenHeightOriginal * scale
                    game.UpdateImages()
                    *imageCache = util.MakeImageCache(game.Cache)
                    makeElements()

                }

                makeChoices := func (name string, scales []int, algorithm data.ScaleAlgorithm) []uilib.Selection {
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

                normalChoices := makeChoices("Normal", []int{1, 2, 3, 4}, data.ScaleAlgorithmNormal)
                scaleChoices := makeChoices("Scale", []int{2, 3, 4}, data.ScaleAlgorithmScale)
                xbrChoices := makeChoices("XBR", []int{2, 3, 4}, data.ScaleAlgorithmXbr)

                choices := append(append(normalChoices, scaleChoices...), xbrChoices...)

                ui.AddElements(uilib.MakeSelectionUI(ui, game.Cache, imageCache, 40, 10, "Resolution", choices))
            },
            Draw: func (element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))
                screen.DrawImage(resolutionBackground, &options)

                x, y := options.GeoM.Apply(float64(3 * data.ScreenScale), float64(3 * data.ScreenScale))
                optionFont.Print(screen, x, y, float64(data.ScreenScale), options.ColorScale, "Screen")
            },
        })

        ui.AddElements(elements)
    }

    makeElements()
}

func (game *Game) doGameMenu(yield coroutine.YieldFunc) {
    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    quit := false

    imageCache := util.MakeImageCache(game.Cache)

    background, _ := imageCache.GetImage("load.lbx", 0, 0)

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            screen.DrawImage(background, &options)

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })
        },
    }

    var elements []*uilib.UIElement

    makeButton := func (index int, x int, y int, action func()) *uilib.UIElement {
        useImage, _ := imageCache.GetImage("load.lbx", index, 0)
        return &uilib.UIElement{
            Rect: util.ImageRect(x * data.ScreenScale, y * data.ScreenScale, useImage),
            PlaySoundLeftClick: true,
            LeftClick: func(element *uilib.UIElement){
                action()
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))
                screen.DrawImage(useImage, &options)
            },
        }
    }

    // quit
    elements = append(elements, makeButton(2, 43, 171, func(){
        game.State = GameStateQuit
        quit = true
    }))

    // load
    elements = append(elements, makeButton(1, 83, 171, func(){
        // FIXME
    }))

    // save
    elements = append(elements, makeButton(3, 122, 171, func(){
        // FIXME
    }))

    // settings
    elements = append(elements, makeButton(12, 172, 171, func(){
        ui.RemoveElements(elements)

        game.MakeSettingsUI(&imageCache, ui, &background, func(){
            quit = true
            // re-enter the game menu
            game.Events <- &GameEventGameMenu{}
        })
    }))

    // ok
    elements = append(elements, makeButton(4, 231, 171, func(){
        quit = true
    }))

    ui.SetElementsFromArray(elements)

    game.Drawer = func (screen *ebiten.Image, game *Game){
        ui.Draw(ui, screen)
    }

    yield()
    for !quit {
        ui.StandardUpdate()

        yield()
    }
    yield()

    game.RefreshUI()
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
    if player.Wizard.AbilityEnabled(setup.AbilityFamous) {
        chance *= 2
    }

    if chance > 10 {
        chance = 10
    }

    if rand.N(100) < chance {
        var heroCandidates []*herolib.Hero
        for _, hero := range game.Heroes {
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
        quit = true
        if hired {
            if player.AddHero(hero) {
                player.Gold -= cost
                hero.SetStatus(herolib.StatusEmployed)
                game.RefreshUI()
            }
        } else {
            hero.GainLevel(units.ExperienceChampionHero)
        }
    }

    game.HudUI.AddElements(MakeHireHeroScreenUI(game.Cache, game.HudUI, hero, cost, result))

    for !quit {
        game.Counter += 1
        game.HudUI.StandardUpdate()
        yield()
    }
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
    if player.Wizard.AbilityEnabled(setup.AbilityFamous) {
        chance *= 2
    }
    if chance > 10 {
        chance = 10
    }
    if rand.N(100) >= 10 {
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
    if player.Wizard.AbilityEnabled(setup.AbilityCharismatic) {
        cost /= 2
    }
    if player.Gold < cost {
        return
    }

    // create units
    var overworldUnits []*units.OverworldUnit
    for i := 0; i < count; i++ {
        overworldUnit := units.MakeOverworldUnitFromUnit(*unit, fortressCity.X, fortressCity.Y, fortressCity.Plane, player.Wizard.Banner, player.MakeExperienceInfo())
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
            }
            player.Gold -= cost
            game.RefreshUI()
        }
    }

    game.HudUI.AddElements(MakeHireMercenariesScreenUI(game.Cache, game.HudUI, units[0], len(units), cost, result))

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
    if player.Wizard.AbilityEnabled(setup.AbilityFamous) {
        chance *= 2
    }
    if chance > 10 {
        chance = 10
    }
    if rand.N(100) >= 10 {
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
    if player.Wizard.AbilityEnabled(setup.AbilityCharismatic) {
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

/* show the given message in an error popup on the screen
 */
func (game *Game) doNotice(yield coroutine.YieldFunc, message string) {
    quit := false
    game.HudUI.AddElement(uilib.MakeErrorElement(game.HudUI, game.Cache, &game.ImageCache, message, func(){
        quit = true
    }))

    for !quit {
        game.Counter += 1
        game.HudUI.StandardUpdate()
        yield()
    }
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

        yes := func(){
            quit = true
            doit = true
        }

        no := func(){
            quit = true
        }

        game.HudUI.AddElements(uilib.MakeConfirmDialog(game.HudUI, game.Cache, &game.ImageCache, message, true, yes, no))

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
        warlord := player.Wizard.AbilityEnabled(setup.AbilityWarlord)
        crusade := player.GlobalEnchantments.Contains(data.EnchantmentCrusade)

        level_before := units.GetHeroExperienceLevel(unit.GetExperience(), warlord, crusade)

        unit.AddExperience(amount)

        level_after := units.GetHeroExperienceLevel(unit.GetExperience(), warlord, crusade)

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

func (game *Game) ProcessEvents(yield coroutine.YieldFunc) {
    // keep processing events until we don't receive one in the events channel
    for {
        select {
            case event := <-game.Events:
                switch event.(type) {
                    case *GameEventMagicView:
                        game.doMagicView(yield)
                    case *GameEventRefreshUI:
                        game.HudUI = game.MakeHudUI()
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
                    case *GameEventNextTurn:
                        game.doNextTurn(yield)
                    case *GameEventSurveyor:
                        game.doSurveyor(yield)
                    case *GameEventApprenticeUI:
                        game.ShowApprenticeUI(yield, game.Players[0])
                    case *GameEventArmyView:
                        game.doArmyView(yield)
                    case *GameEventShowBanish:
                        banishEvent := event.(*GameEventShowBanish)
                        game.doBanish(yield, banishEvent.Attacker, banishEvent.Defender)
                    case *GameEventNotice:
                        notice := event.(*GameEventNotice)
                        game.doNotice(yield, notice.Message)
                    case *GameEventCastSpellBook:
                        game.ShowSpellBookCastUI(yield, game.Players[0])
                    case *GameEventCityListView:
                        game.doCityListView(yield)
                    case *GameEventNewOutpost:
                        outpost := event.(*GameEventNewOutpost)
                        if outpost.Player.IsHuman() {
                            game.showOutpost(yield, outpost.City, outpost.Stack, true)
                        }
                    case *GameEventVault:
                        vaultEvent := event.(*GameEventVault)
                        game.doVault(yield, vaultEvent.CreatedArtifact)
                    case *GameEventScroll:
                        scroll := event.(*GameEventScroll)
                        game.showScroll(yield, scroll.Title, scroll.Text)
                    case *GameEventLearnedSpell:
                        learnedSpell := event.(*GameEventLearnedSpell)
                        game.doLearnSpell(yield, learnedSpell.Player, learnedSpell.Spell)
                    case *GameEventResearchSpell:
                        researchSpell := event.(*GameEventResearchSpell)
                        game.ResearchNewSpell(yield, researchSpell.Player)
                    case *GameEventCastGlobalEnchantment:
                        castGlobal := event.(*GameEventCastGlobalEnchantment)
                        game.doCastGlobalEnchantment(yield, castGlobal.Player, castGlobal.Enchantment)
                    case *GameEventSelectLocationForSpell:
                        selectLocation := event.(*GameEventSelectLocationForSpell)
                        tileX, tileY, cancel := game.selectLocationForSpell(yield, selectLocation.Spell, selectLocation.Player, selectLocation.LocationType)
                        if !cancel {
                            selectLocation.SelectedFunc(yield, tileX, tileY)
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

                        if player.IsHuman() {
                            game.Music.PushSong(music.SongCommonSummoningSpell)
                            game.doSummon(yield, summon.MakeSummonUnit(game.Cache, summonUnit.Unit, player.Wizard.Base))
                            game.Music.PopSong()
                        }
                    case *GameEventSummonArtifact:
                        summonArtifact := event.(*GameEventSummonArtifact)
                        game.Music.PushSong(music.SongVeryRareSummoningSpell)
                        game.doSummon(yield, summon.MakeSummonArtifact(game.Cache, summonArtifact.Wizard))
                        game.Music.PopSong()
                    case *GameEventSummonHero:
                        summonHero := event.(*GameEventSummonHero)
                        game.Music.PushSong(music.SongVeryRareSummoningSpell)
                        game.doSummon(yield, summon.MakeSummonHero(game.Cache, summonHero.Wizard, summonHero.Champion))
                        game.Music.PopSong()
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
            default:
                return
        }
    }
}

func ChooseUniqueWizard(players []*playerlib.Player, allSpells spellbook.Spells) (setup.WizardCustom, bool) {
    // pick a new wizard with an unused wizard base and banner color, and race
    // if on myrror then select a myrran race

    chooseBase := func() (setup.WizardSlot, bool) {
        choices := slices.Clone(setup.DefaultWizardSlots())
        choices = slices.DeleteFunc(choices, func (wizard setup.WizardSlot) bool {
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

    race, ok := chooseRace(wizard.ExtraAbility == setup.AbilityMyrran)
    if !ok {
        return setup.WizardCustom{}, false
    }

    banner, ok := chooseBanner()
    if !ok {
        return setup.WizardCustom{}, false
    }

    var abilities []setup.WizardAbility
    if wizard.ExtraAbility != setup.AbilityNone {
        abilities = []setup.WizardAbility{wizard.ExtraAbility}
    }

    customWizard := setup.WizardCustom{
        Name: wizard.Name,
        Base: wizard.Base,
        Race: race,
        Books: slices.Clone(wizard.Books),
        Banner: banner,
        Abilities: abilities,
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
func (game *Game) TileToScreen(tileX int, tileY int) (int, int) {
    tileWidth := game.CurrentMap().TileWidth()
    tileHeight := game.CurrentMap().TileHeight()

    var geom ebiten.GeoM

    x, y := (game.CurrentMap().XDistance(game.Camera.GetX(), tileX) + game.Camera.GetX()) * tileWidth, tileY * tileHeight
    geom.Translate(float64(x + tileWidth / 2.0), float64(y + tileHeight / 2.0))
    geom.Translate(-game.Camera.GetZoomedX() * float64(tileWidth), -game.Camera.GetZoomedY() * float64(tileHeight))
    geom.Scale(game.Camera.GetAnimatedZoom(), game.Camera.GetAnimatedZoom())

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
    for camera.GetZoomedY() < -1 {
        y += 1
        camera.Center(x, y)
    }

    for camera.GetZoomedMaxY() >= float64(game.CurrentMap().Height()) {
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

func (game *Game) doMoveFleeingDefender(player *playerlib.Player, stack *playerlib.UnitStack) {
    // try to relocate a fleeing stack, kills units that are unable
    x := stack.X()
    y := stack.Y()
    plane := stack.Plane()
    mapUse := game.GetMap(plane)
    canMoveToWater := stack.AllFlyers() || stack.AllSwimmers()

    var positions []image.Point
    var waterPositions []image.Point
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

            // can not countain encounter
            if mapUse.GetEncounter(cx, cy) != nil {
                continue
            }

            if mapUse.GetTile(cx, cy).Tile.IsWater() && !canMoveToWater {
                waterPositions = append(waterPositions, image.Pt(cx, cy))
            } else {
                positions = append(positions, image.Pt(cx, cy))
            }
        }
    }

    // kill units that can not move to water if only water is available
    if len(positions) == 0 && len(waterPositions) != 0 {
        for _, unit := range stack.Units() {
            if !unit.IsFlying() && !unit.IsSwimmer() {
                player.RemoveUnit(unit)
            }
        }

        positions = waterPositions
    }

    // kill whole stack if no position found
    if len(positions) == 0 {
        for _, unit := range stack.Units() {
            player.RemoveUnit(unit)
        }
        return
    }

    // set to a random position
    position := positions[rand.IntN(len(positions))]
    stack.SetX(position.X)
    stack.SetY(position.Y)

    allStacks := player.FindAllStacks(position.X, position.Y, stack.Plane())
    for i := 1; i < len(allStacks); i++ {
        player.MergeStacks(allStacks[0], allStacks[i])
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
        defender.RemoveCity(city)
        attacker.AddCity(city)
        city.Banner = attacker.Wizard.Banner
        city.RulingRace = attacker.Wizard.Race
        city.UpdateTaxRate(attacker.TaxRate, attackerStack.Units())

        city.Buildings.Remove(buildinglib.BuildingFortress)
        city.Buildings.Remove(buildinglib.BuildingSummoningCircle)
    }

    if containedFortress {
        defender.Banished = true

        if attacker.IsHuman() || defender.IsHuman() {
            game.Events <- &GameEventShowBanish{Attacker: attacker, Defender: defender}
        }
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

func (game *Game) doMoveSelectedUnit(yield coroutine.YieldFunc, player *playerlib.Player) {
    stack := player.SelectedStack
    if stack == nil || len(stack.ActiveUnits()) == 0 {
        return
    }

    mapUse := game.GetMap(stack.Plane())

    stepsTaken := 0
    stopMoving := false
    var mergeStack *playerlib.UnitStack

    quitMoving:
    for i, step := range stack.CurrentPath {
        if stack.OutOfMoves() {
            break
        }

        oldX := stack.X()
        oldY := stack.Y()

        terrainCost, canMove := game.ComputeTerrainCost(stack, stack.X(), stack.Y(), step.X, step.Y, mapUse)

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

            for _, otherPlayer := range game.Players[1:] {
                // FIXME: this should get all stacks at the given location and merge them into a single stack for combat
                otherStack := otherPlayer.FindStack(stack.X(), stack.Y(), stack.Plane())
                if otherStack != nil {
                    otherCity := otherPlayer.FindCity(stack.X(), stack.Y(), stack.Plane())
                    zone := combat.ZoneType{
                        City: otherCity,
                    }

                    state := game.doCombat(yield, player, stack, otherPlayer, otherStack, zone)
                    if state == combat.CombatStateAttackerFlee {
                        stack.SetX(oldX)
                        stack.SetY(oldY)
                    } else if state == combat.CombatStateDefenderFlee {
                        game.doMoveFleeingDefender(otherPlayer, otherStack)
                    }

                    stack.ExhaustMoves()
                    game.RefreshUI()

                    stopMoving = true
                    break quitMoving
                }

                // defeat any unguarded cities immediately
                otherCity := otherPlayer.FindCity(stack.X(), stack.Y(), stack.Plane())
                if otherCity != nil {
                    raze, gold := game.defeatCity(yield, player, stack, otherPlayer, otherCity)

                    // FIXME: show a notice about any fame won
                    player.Fame += otherCity.FameForCaptureOrRaze(!raze)
                    otherPlayer.Fame += otherCity.FameForCaptureOrRaze(false)
                    player.Gold += gold
                    otherPlayer.Gold -= gold

                    stack.ExhaustMoves()
                    game.RefreshUI()

                    stopMoving = true
                    break quitMoving

                }
            }

            // some units in the stack might not have any moves left
            beforeActive := len(stack.ActiveUnits())
            stack.EnableMovers()
            afterActive := len(stack.ActiveUnits())
            if afterActive > 0 && afterActive != beforeActive {
                // stopMoving = true
                break
            }
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

    // only merge stacks if both stacks are stopped, otherwise they can move through each other
    if len(stack.CurrentPath) == 0 && mergeStack != nil {
        stack = player.MergeStacks(mergeStack, stack)
        player.SelectedStack = stack
        game.RefreshUI()
    }

    // update unrest for new units in the city
    newCity := player.FindCity(stack.X(), stack.Y(), stack.Plane())
    if newCity != nil {
        newCity.UpdateUnrest(stack.Units())
    }

    if stepsTaken > 0 {
        if stack.AnyOutOfMoves() {
            stack.ExhaustMoves()
            game.DoNextUnit(player)
        }

        game.RefreshUI()
    }
}

// given a position on the screen in pixels, return true if the position is within the area of the ui designated for the overworld
func (game *Game) InOverworldArea(x int, y int) bool {
    return x < 240 * data.ScreenScale && y > 18 * data.ScreenScale
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

    if player.SelectedStack != nil {
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

                        path := game.FindPath(oldX, oldY, newX, newY, stack, player.GetFog(game.Plane))
                        if path == nil {
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
                } else {
                    // make a copy of the unit stack to activate all units, because path finding only checks active units for terrain constraints
                    path := game.FindPath(oldX, oldY, newX, newY, playerlib.MakeUnitStackFromUnits(stack.Units()), player.GetFog(game.Plane))
                    if path == nil {
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
                        game.showOutpost(yield, city, player.FindStack(city.X, city.Y, city.Plane), false)
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
                            } else {
                                enemyStack := otherPlayer.FindStack(tileX, tileY, game.Plane)
                                if enemyStack != nil {
                                    quit := false
                                    clicked := func(){
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

func (game *Game) doAiMoveUnit(yield coroutine.YieldFunc, player *playerlib.Player, move *playerlib.AIMoveStackDecision) {
    stack := move.Stack
    to := move.Location
    log.Printf("  moving stack %v to %v, %v", stack, to.X, to.Y)
    terrainCost, ok := game.ComputeTerrainCost(stack, stack.X(), stack.Y(), to.X, to.Y, game.GetMap(stack.Plane()))
    if ok {
        oldX := stack.X()
        oldY := stack.Y()

        mapUse := game.GetMap(stack.Plane())

        encounter := mapUse.GetEncounter(mapUse.WrapX(to.X), to.Y)
        if encounter != nil && move.ConfirmEncounter != nil {
            if !move.ConfirmEncounter(encounter) {
                move.Invalid()
                return
            }
        }

        stack.Move(to.X - stack.X(), to.Y - stack.Y(), terrainCost, game.GetNormalizeCoordinateFunc())

        if game.Players[0].IsVisible(oldX, oldY, stack.Plane()) {
            game.showMovement(yield, oldX, oldY, stack, false)
        }

        player.LiftFogSquare(stack.X(), stack.Y(), stack.GetSightRange(), stack.Plane())

        if encounter != nil {
            game.doEncounter(yield, player, stack, encounter, mapUse, stack.X(), stack.Y())
            return
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

                return
            }

            city := enemy.FindCity(stack.X(), stack.Y(), stack.Plane())
            if city != nil {
                raze, gold := game.defeatCity(yield, player, stack, enemy, city)

                player.Fame += city.FameForCaptureOrRaze(!raze)
                enemy.Fame += city.FameForCaptureOrRaze(false)
                player.Gold += gold
                enemy.Gold -= gold

                // FIXME: if the wizard is neutral and decides to raze the town, then the town could become
                // an encounter zone

                return
            }
        }
    } else if move.Invalid != nil {
        move.Invalid()
    }
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
                    game.doAiMoveUnit(yield, player, decision.(*playerlib.AIMoveStackDecision))

                case *playerlib.AICreateUnitDecision:
                    create := decision.(*playerlib.AICreateUnitDecision)
                    log.Printf("ai %v creating %+v", player.Wizard.Name, create)

                    existingStack := player.FindStack(create.X, create.Y, create.Plane)
                    if existingStack == nil || len(existingStack.Units()) < 9 {
                        overworldUnit := units.MakeOverworldUnitFromUnit(create.Unit, create.X, create.Y, create.Plane, player.Wizard.Banner, player.MakeExperienceInfo())
                        player.AddUnit(overworldUnit)
                    }
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

    if len(decisions) == 0 {
        game.DoNextTurn()
    }
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

    oldDrawer := game.Drawer
    game.Drawer = func(screen *ebiten.Image, game *Game){
        cityScreen.Draw(screen, func (mapView *ebiten.Image, geom ebiten.GeoM, counter uint64){
            overworld.DrawOverworld(mapView, geom)
        }, mapUse.TileWidth(), mapUse.TileHeight())
    }

    for cityScreen.Update() == cityview.CityScreenStateRunning {
        overworld.Counter += 1
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
            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })
        },
    }

    ui.SetElementsFromArray(nil)

    ui.AddElements(uilib.MakeConfirmDialogWithLayerFull(ui, game.Cache, &game.ImageCache, 1, "Do you wish to completely destroy this city?", true, no, yes, noImages, yesImages))

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
        animation = util.MakePaletteRotateAnimation(reloadLbx, &game.ImageCache, lairIndex, rotateIndexLow, rotateIndexHigh)
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

    defender := playerlib.Player{
        Wizard: setup.WizardCustom{
            Name: encounter.Type.ShortName(),
        },
        StrategicCombat: true,
    }

    var enemies []units.StackUnit

    for _, unit := range encounter.Units {
        enemies = append(enemies, units.MakeOverworldUnit(unit))
    }

    zone := combat.ZoneType{
    }

    switch encounter.Type {
        case maplib.EncounterTypeLair, maplib.EncounterTypeCave: zone.Lair = true
        case maplib.EncounterTypePlaneTower: zone.Tower = true
        case maplib.EncounterTypeAncientTemple: zone.AncientTemple = true
        case maplib.EncounterTypeFallenTemple: zone.FallenTemple = true
        case maplib.EncounterTypeRuins: zone.Ruins = true
        case maplib.EncounterTypeAbandonedKeep: zone.AbandonedKeep = true
        case maplib.EncounterTypeDungeon: zone.Dungeon = true
        case maplib.EncounterTypeNatureNode: zone.NatureNode = true
        case maplib.EncounterTypeSorceryNode: zone.SorceryNode = true
        case maplib.EncounterTypeChaosNode: zone.ChaosNode = true
    }

    result := game.doCombat(yield, player, stack, &defender, playerlib.MakeUnitStackFromUnits(enemies), zone)
    if result == combat.CombatStateAttackerWin {
        mapUse.RemoveEncounter(x, y)

        game.createTreasure(encounter.Type, encounter.Budget, player)
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
        for _, hero := range game.Heroes {
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

func (game *Game) doTreasurePopup(yield coroutine.YieldFunc, player *playerlib.Player, treasure Treasure){
    uiDone := false

    fontLbx, err := game.Cache.GetLbxFile("FONTS.LBX")
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }

    orange := color.RGBA{R: 0xc7, G: 0x82, B: 0x1b, A: 0xff}

    yellowPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        orange,
        util.Lighten(orange, 15),
        util.Lighten(orange, 30),
        util.Lighten(orange, 50),
        orange,
        orange,
    }

    treasureFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowPalette)

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
            options.GeoM.Translate(float64(10 * data.ScreenScale), float64(50 * data.ScreenScale))

            fontX, fontY := options.GeoM.Apply(float64(10 * data.ScreenScale), float64(10 * data.ScreenScale))

            screen.DrawImage(left, &options)
            right, _ := game.ImageCache.GetImage("resource.lbx", 58, 0)
            options.GeoM.Translate(float64(left.Bounds().Dx()), 0)
            rightGeom := options.GeoM

            chest, _ := game.ImageCache.GetImage("reload.lbx", 20, 0)
            options.GeoM.Translate(float64(6 * data.ScreenScale), float64(8 * data.ScreenScale))
            screen.DrawImage(chest, &options)

            options.GeoM = rightGeom
            screen.DrawImage(right, &options)

            treasureFont.PrintWrap(screen, fontX, fontY, float64(left.Bounds().Dx() - 5 * data.ScreenScale), float64(data.ScreenScale), options.ColorScale, treasure.String())
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
                // FIXME: somehow recompute the research spell pool for the player
                player.Wizard.AddMagicLevel(spellbook.Magic, spellbook.Count)
            case *TreasureRetort:
                retort := item.(*TreasureRetort)
                player.Wizard.EnableAbility(retort.Retort)
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
        case terrain.Ocean: return combat.CombatLandscapeGrass
        case terrain.Volcano: return combat.CombatLandscapeGrass
        case terrain.Lake: return combat.CombatLandscapeGrass
        case terrain.NatureNode: return combat.CombatLandscapeGrass
        case terrain.SorceryNode: return combat.CombatLandscapeGrass
        case terrain.ChaosNode: return combat.CombatLandscapeMountain
    }

    return combat.CombatLandscapeGrass
}

/* run the tactical combat screen. returns the combat state as a result (attackers win, defenders win, flee, etc)
 * this also shows the raze city ui so that fame can be incorporated based on whether the city is razed or not
 */
func (game *Game) doCombat(yield coroutine.YieldFunc, attacker *playerlib.Player, attackerStack *playerlib.UnitStack, defender *playerlib.Player, defenderStack *playerlib.UnitStack, zone combat.ZoneType) combat.CombatState {
    attackingArmy := combat.Army{
        Player: attacker,
    }

    for _, unit := range attackerStack.Units() {
        attackingArmy.AddUnit(unit)
    }

    defendingArmy := combat.Army{
        Player: defender,
    }

    for _, unit := range defenderStack.Units() {
        defendingArmy.AddUnit(unit)
    }

    attackingArmy.LayoutUnits(combat.TeamAttacker)
    defendingArmy.LayoutUnits(combat.TeamDefender)

    var state combat.CombatState
    var defeatedDefenders int
    var defeatedAttackers int

    oldDrawer := game.Drawer
    var combatScreen *combat.CombatScreen

    if attacker.StrategicCombat && defender.StrategicCombat {
        state, defeatedAttackers, defeatedDefenders = combat.DoStrategicCombat(&attackingArmy, &defendingArmy)
        log.Printf("Strategic combat result state=%v", state)
    } else {

        defer mouse.Mouse.SetImage(game.MouseData.Normal)

        landscape := game.GetCombatLandscape(attackerStack.X(), attackerStack.Y(), attackerStack.Plane())

        // FIXME: take plane into account for the landscape/terrain
        combatScreen = combat.MakeCombatScreen(game.Cache, &defendingArmy, &attackingArmy, game.Players[0], landscape, attackerStack.Plane(), zone)

        // ebiten.SetCursorMode(ebiten.CursorModeHidden)

        game.Drawer = func (screen *ebiten.Image, game *Game){
            combatScreen.Draw(screen)
        }

        game.Music.PushSong(music.SongCombat1)

        state = combat.CombatStateRunning
        for state == combat.CombatStateRunning {
            state = combatScreen.Update(yield)
            yield()
        }

        game.Music.PopSong()

        defeatedDefenders = combatScreen.Model.DefeatedDefenders
        defeatedAttackers = combatScreen.Model.DefeatedAttackers
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
    } else if state == combat.CombatStateDefenderWin || state == combat.CombatStateAttackerFlee {
        winner, loser := distributeFame(defender, attacker, attackerStack, defeatedAttackers)
        defenderFame += winner
        attackerFame += loser
    }

    attacker.Fame = max(0, attacker.Fame + attackerFame)
    defender.Fame = max(0, defender.Fame + defenderFame)

    // Show end screen
    if !attacker.StrategicCombat || !defender.StrategicCombat {
        result := combat.CombatEndScreenResultLoose
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
        endScreen := combat.MakeCombatEndScreen(game.Cache, combatScreen, result, combatScreen.Model.DiedWhileFleeing, fame)
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
                    case game.Events <- &GameEventVault{CreatedArtifact: item}:
                    default:
                }
            }
        }
    }

    // ebiten.SetCursorMode(ebiten.CursorModeVisible)

    // remove dead units
    for _, unit := range attackerStack.Units() {
        if unit.GetHealth() <= 0 {
            attacker.RemoveUnit(unit)

            if unit.IsHero() && attacker.IsHuman() {
                hero := unit.(*herolib.Hero)
                distributeEquipment(attacker, hero)
            }
        }
    }

    for _, unit := range defenderStack.Units() {
        if unit.GetHealth() <= 0 {
            defender.RemoveUnit(unit)

            if unit.IsHero() && defender.IsHuman() {
                hero := unit.(*herolib.Hero)
                distributeEquipment(defender, hero)
            }

        }
    }

    if showHeroNotice {
        game.doNotice(yield, "One or more heroes died in combat. You must redistribute their equipment.")
    }

    return state
}

func GetUnitImage(unit units.StackUnit, imageCache *util.ImageCache, banner data.BannerType) (*ebiten.Image, error) {
    image, err := imageCache.GetImageTransform(unit.GetLbxFile(), unit.GetLbxIndex(), 0, banner.String(), units.MakeUpdateUnitColorsFunc(banner))

    if err != nil {
        log.Printf("Error: unit '%v' image in lbx file %v is missing: %v", unit.GetName(), unit.GetLbxFile(), err)
    }

    return image, err
}

func GetCityNoWallImage(size citylib.CitySize, cache *util.ImageCache) (*ebiten.Image, error) {
    var index int = 0

    switch size {
        case citylib.CitySizeHamlet: index = 0
        case citylib.CitySizeVillage: index = 1
        case citylib.CitySizeTown: index = 2
        case citylib.CitySizeCity: index = 3
        case citylib.CitySizeCapital: index = 4
    }

    // the city image is a sub-frame of animation 21
    return cache.GetImage("mapback.lbx", 21, index)
}

func GetCityWallImage(city *citylib.City, cache *util.ImageCache) (*ebiten.Image, error) {
    var index int = 0

    switch city.GetSize() {
        case citylib.CitySizeHamlet: index = 0
        case citylib.CitySizeVillage: index = 1
        case citylib.CitySizeTown: index = 2
        case citylib.CitySizeCity: index = 3
        case citylib.CitySizeCapital: index = 4
    }

    // the city image is a sub-frame of animation 20
    // return cache.GetImageTransform("mapback.lbx", 20, index, city.Banner.String(), util.ComposeImageTransform(units.MakeUpdateUnitColorsFunc(city.Banner), util.AutoCropGeneric))
    return cache.GetImageTransform("mapback.lbx", 20, index, city.Banner.String(), units.MakeUpdateUnitColorsFunc(city.Banner))
}

func (game *Game) ShowGrandVizierUI(){
    yes := func(){
        // FIXME: enable grand vizier
    }

    no := func(){
        // FIXME: disable grand vizier
    }

    game.HudUI.AddElements(uilib.MakeConfirmDialogWithLayer(game.HudUI, game.Cache, &game.ImageCache, 1, "Do you wish to allow the Grand Vizier to select what buildings your cities create?", true, yes, no))
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

    game.HudUI.AddElements(uilib.MakeSelectionUI(game.HudUI, game.Cache, &game.ImageCache, cornerX, cornerY, "Tax Per Population", taxes))
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
    spellbook.ShowSpellBook(yield, game.Cache, player.ResearchPoolSpells, player.KnownSpells, player.ResearchCandidateSpells, player.ResearchingSpell, player.ResearchProgress, int(player.SpellResearchPerTurn(power)), player.ComputeCastingSkill(), spellbook.Spell{}, false, nil, &newDrawer)
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
        spellbook.ShowSpellBook(yield, game.Cache, player.ResearchPoolSpells, player.KnownSpells, player.ResearchCandidateSpells, spellbook.Spell{}, 0, int(player.SpellResearchPerTurn(power)), player.ComputeCastingSkill(), spellbook.Spell{}, true, &player.ResearchingSpell, &newDrawer)
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
            Action: func(){},
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
            Action: func(){},
            Hotkey: "(F4)",
        },
        uilib.Selection{
            Name: "Astrologer",
            Action: func(){},
            Hotkey: "(F5)",
        },
        uilib.Selection{
            Name: "Chancellor",
            Action: func(){},
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
                if len(game.Players) > 0 {
                    game.HudUI.AddElement(mirror.MakeMirrorUI(game.Cache, game.Players[0], game.HudUI))
                }
            },
            Hotkey: "(F9)",
        },
    }

    return uilib.MakeSelectionUI(game.HudUI, game.Cache, &game.ImageCache, cornerX, cornerY, "Select An Advisor", advisors)
}

func (game *Game) ShowSpellBookCastUI(yield coroutine.YieldFunc, player *playerlib.Player){
    game.HudUI.AddElements(spellbook.MakeSpellBookCastUI(game.HudUI, game.Cache, player.KnownSpells.OverlandSpells(), make(map[spellbook.Spell]int), player.ComputeCastingSkill(), player.CastingSpell, player.CastingSpellProgress, true, func (spell spellbook.Spell, picked bool){
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

                created, cancel := artifact.ShowCreateArtifactScreen(yield, game.Cache, creation, &player.Wizard, player.Wizard.AbilityEnabled(setup.AbilityArtificer), player.Wizard.AbilityEnabled(setup.AbilityRunemaster), player.KnownSpells.CombatSpells(), &drawFunc)
                if cancel {
                    return
                }

                log.Printf("Create artifact %v", created)
                spell.OverrideCost = created.Cost

                player.CreateArtifact = created
            }

            castingCost := spell.Cost(true)

            // FIXME: if the player has runemaster and the spell is arcane, then apply a -25% reduction. Don't apply
            // to create artifact or enchant item because the reduction has already been applied

            if castingCost <= player.Mana && castingCost <= player.RemainingCastingSkill {
                player.Mana -= castingCost
                player.RemainingCastingSkill -= castingCost

                select {
                    case game.Events<- &GameEventCastSpell{Player: player, Spell: spell}:
                    default:
                }
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

func (game *Game) CreateOutpost(settlers units.StackUnit, player *playerlib.Player) *citylib.City {
    cityName := game.SuggestCityName(settlers.GetRace())

    newCity := citylib.MakeCity(cityName, settlers.GetX(), settlers.GetY(), settlers.GetRace(), settlers.GetBanner(), player.TaxRate, game.BuildingInfo, game.GetMap(settlers.GetPlane()), game)
    newCity.Plane = settlers.GetPlane()
    newCity.RulingRace = player.Wizard.Race
    newCity.Population = 300
    newCity.Outpost = true
    newCity.Banner = player.Wizard.Banner
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
                    unit.SetMovesLeft(fraction.Zero())
                }
            }

            player.SelectedStack.EnableMovers()

            // player.SelectedStack.ExhaustMoves()
            game.RefreshUI()

        } else if powers.BuildRoad {

            for _, unit := range player.SelectedStack.ActiveUnits() {
                if unit.HasAbility(data.AbilityConstruction) {
                    unit.SetBusy(units.BusyStatusBuildRoad)
                    unit.SetMovesLeft(fraction.Zero())
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
            screen.DrawImage(mainHud, &options)

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })
        },
        HandleKeys: func(keys []ebiten.Key){
            for _, key := range keys {
                switch key {
                    case ebiten.KeySpace:
                        stack := game.Players[0].SelectedStack

                        if stack == nil {
                            select {
                                case game.Events <- &GameEventNextTurn{}:
                                default:
                            }
                        } else {
                            select {
                                case game.Events <- &GameEventMoveCamera{Plane: stack.Plane(), X: stack.X(), Y: stack.Y()}:
                                default:
                            }
                        }
                }
            }
        },
    }

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
                screen.DrawImage(buttons[index], &options)
            },
        }
    }

    var elements []*uilib.UIElement

    // game button
    elements = append(elements, makeButton(1, 7 * data.ScreenScale, 4 * data.ScreenScale, false, func(){
        select {
            case game.Events <- &GameEventGameMenu{}:
            default:
        }
    }))

    // spell button
    elements = append(elements, makeButton(2, 47 * data.ScreenScale, 4 * data.ScreenScale, false, func(){
        select {
            case game.Events <- &GameEventCastSpellBook{}:
            default:
        }
    }))

    // army button
    elements = append(elements, makeButton(3, 89 * data.ScreenScale, 4 * data.ScreenScale, false, func(){
        select {
            case game.Events<- &GameEventArmyView{}:
            default:
        }
    }))

    // cities button
    elements = append(elements, makeButton(4, 140 * data.ScreenScale, 4 * data.ScreenScale, false, func(){
        select {
            case game.Events<- &GameEventCityListView{}:
            default:
        }
    }))

    // magic button
    elements = append(elements, makeButton(5, 184 * data.ScreenScale, 4 * data.ScreenScale, false, func(){
        select {
            case game.Events<- &GameEventMagicView{}:
            default:
        }
    }))

    // info button
    elements = append(elements, makeButton(6, 226 * data.ScreenScale, 4 * data.ScreenScale, true, func(){
        ui.AddElements(game.MakeInfoUI(60, 25))
    }))

    // plane button
    elements = append(elements, makeButton(7, 270 * data.ScreenScale, 4 * data.ScreenScale, false, func(){
        game.SwitchPlane()

        game.RefreshUI()
    }))

    if len(game.Players) > 0 && game.Players[0].SelectedStack != nil {
        player := game.Players[0]
        // stack := player.SelectedStack

        unitX1 := 246 * data.ScreenScale
        unitY1 := 79 * data.ScreenScale

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
                        stack.ToggleActive(unit)
                        select {
                            case game.Events<- &GameEventMoveUnit{Player: player}:
                            default:
                        }

                        updateMinMoves()
                    },
                    RightClick: func(this *uilib.UIElement){
                        ui.AddElements(unitview.MakeUnitContextMenu(game.Cache, ui, unit, disband))
                    },
                    Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                        var options ebiten.DrawImageOptions
                        options.GeoM.Translate(float64(unitRect.Min.X), float64(unitRect.Min.Y))
                        screen.DrawImage(unitBackground, &options)

                        options.GeoM.Translate(float64(data.ScreenScale), float64(data.ScreenScale))

                        if stack.IsActive(unit){
                            unitBack, _ := units.GetUnitBackgroundImage(unit.GetBanner(), &game.ImageCache)
                            screen.DrawImage(unitBack, &options)
                        }

                        options.GeoM.Translate(float64(data.ScreenScale), float64(data.ScreenScale))
                        unitImage, err := GetUnitImage(unit, &game.ImageCache, unit.GetBanner())
                        if err == nil {

                            if unit.GetBusy() != units.BusyStatusNone {
                                var patrolOptions colorm.DrawImageOptions
                                var matrix colorm.ColorM
                                patrolOptions.GeoM = options.GeoM
                                matrix.ChangeHSV(0, 0, 1)
                                colorm.DrawImage(screen, unitImage, matrix, &patrolOptions)
                            } else {
                                screen.DrawImage(unitImage, &options)
                            }

                            // draw the first enchantment on the unit
                            for _, enchantment := range unit.GetEnchantments() {
                                x, y := options.GeoM.Apply(0, 0)
                                util.DrawOutline(screen, &game.ImageCache, unitImage, x, y, options.ColorScale, game.Counter/10, enchantment.Color())
                                break
                            }
                        }

                        if unit.GetHealth() < unit.GetMaxHealth() {
                            highHealth := color.RGBA{R: 0, G: 0xff, B: 0, A: 0xff}
                            mediumHealth := color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff}
                            lowHealth := color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}

                            healthWidth := float64(10 * data.ScreenScale)
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

                            x, y := options.GeoM.Apply(float64(4 * data.ScreenScale), float64(19 * data.ScreenScale))
                            vector.StrokeLine(screen, float32(x), float32(y), float32(x + healthLength), float32(y), 1, useColor, false)
                        }

                        silverBadge := 51
                        goldBadge := 52
                        redBadge := 53
                        count := 0
                        index := 0

                        // draw experience badges
                        if unit.GetRace() == data.RaceHero {
                            switch units.GetHeroExperienceLevel(unit.GetExperience(), player.Wizard.AbilityEnabled(setup.AbilityWarlord), player.GlobalEnchantments.Contains(data.EnchantmentCrusade)) {
                            case units.ExperienceHero:
                            case units.ExperienceMyrmidon:
                                count = 1
                                index = silverBadge
                            case units.ExperienceCaptain:
                                count = 2
                                index = silverBadge
                            case units.ExperienceCommander:
                                count = 3
                                index = silverBadge
                            case units.ExperienceChampionHero:
                                count = 1
                                index = goldBadge
                            case units.ExperienceLord:
                                count = 2
                                index = goldBadge
                            case units.ExperienceGrandLord:
                                count = 3
                                index = goldBadge
                            case units.ExperienceSuperHero:
                                count = 1
                                index = redBadge
                            case units.ExperienceDemiGod:
                                count = 2
                                index = redBadge
                            }
                        } else {

                            switch units.GetNormalExperienceLevel(unit.GetExperience(), player.Wizard.AbilityEnabled(setup.AbilityWarlord), player.GlobalEnchantments.Contains(data.EnchantmentCrusade)) {
                            case units.ExperienceRecruit:
                                // nothing
                            case units.ExperienceRegular:
                                // one white circle
                                count = 1
                                index = silverBadge
                            case units.ExperienceVeteran:
                                // two white circles
                                count = 2
                                index = silverBadge
                            case units.ExperienceElite:
                                // three white circles
                                count = 3
                                index = silverBadge
                            case units.ExperienceUltraElite:
                                // one yellow
                                count = 1
                                index = goldBadge
                            case units.ExperienceChampionNormal:
                                // two yellow
                                count = 2
                                index = goldBadge
                            }
                        }

                        badgeOptions := options
                        badgeOptions.GeoM.Translate(float64(1 * data.ScreenScale), float64(21 * data.ScreenScale))
                        for i := 0; i < count; i++ {
                            pic, _ := game.ImageCache.GetImage("main.lbx", index, 0)
                            screen.DrawImage(pic, &badgeOptions)
                            badgeOptions.GeoM.Translate(float64(4 * data.ScreenScale), 0)
                        }

                        weaponOptions := options
                        weaponOptions.GeoM.Translate(float64(12 * data.ScreenScale), float64(18 * data.ScreenScale))
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
                            screen.DrawImage(weapon, &weaponOptions)
                        }

                        // draw a G on the unit if they are moving, P if purify, and B if building road
                        if unit.GetBusy() == units.BusyStatusBuildRoad {
                            x, y := options.GeoM.Apply(float64(1 * data.ScreenScale), float64(1 * data.ScreenScale))
                            game.WhiteFont.Print(screen, x, y, float64(data.ScreenScale), options.ColorScale, "B")
                        } else if unit.GetBusy() == units.BusyStatusPurify {
                            x, y := options.GeoM.Apply(float64(1 * data.ScreenScale), float64(1 * data.ScreenScale))
                            game.WhiteFont.Print(screen, x, y, float64(data.ScreenScale), options.ColorScale, "P")
                        } else if len(stack.CurrentPath) != 0 {
                            x, y := options.GeoM.Apply(float64(1 * data.ScreenScale), float64(1 * data.ScreenScale))
                            game.WhiteFont.Print(screen, x, y, float64(data.ScreenScale), options.ColorScale, "G")
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
            doneRect := util.ImageRect(246 * data.ScreenScale, 176 * data.ScreenScale, doneImages[0])
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
                    screen.DrawImage(doneImages[doneIndex], &options)
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
            patrolRect := util.ImageRect(280 * data.ScreenScale, 176 * data.ScreenScale, patrolImages[0])
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
                    screen.DrawImage(patrolImages[patrolIndex], &options)
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
            waitRect := util.ImageRect(246 * data.ScreenScale, 186 * data.ScreenScale, waitImages[0])
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
                    screen.DrawImage(waitImages[waitIndex], &options)
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
            buildRect := util.ImageRect(280 * data.ScreenScale, 186 * data.ScreenScale, buildImages[0])
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
                        if node != nil && !node.Warped{
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
                    x := float64(246.0 * data.ScreenScale)
                    y := float64(167.0 * data.ScreenScale)
                    game.WhiteFont.Print(screen, x, y, float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("Moves:%v", minMoves.ToFloat()))

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
                    options.GeoM.Translate(x + float64(60 * data.ScreenScale), y)
                    screen.DrawImage(useIcon, &options)
                }
            },
        })


    } else {
        // next turn
        nextTurnImage, _ := game.ImageCache.GetImage("main.lbx", 35, 0)
        nextTurnImageClicked, _ := game.ImageCache.GetImage("main.lbx", 58, 0)
        nextTurnRect := image.Rect(240 * data.ScreenScale, 174 * data.ScreenScale, 240 * data.ScreenScale + nextTurnImage.Bounds().Dx(), 174 * data.ScreenScale + nextTurnImage.Bounds().Dy())
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
                    ui.AddElement(uilib.MakeHelpElementWithLayer(ui, game.Cache, &game.ImageCache, 1, helpEntries[0], helpEntries[1:]...))
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(240 * data.ScreenScale), float64(174 * data.ScreenScale))
                screen.DrawImage(nextTurnImage, &options)
                if nextTurnClicked {
                    options.GeoM.Translate(float64(6 * data.ScreenScale), float64(5 * data.ScreenScale))
                    screen.DrawImage(nextTurnImageClicked, &options)
                }
            },
        })

        if len(game.Players) > 0 {
            player := game.Players[0]

            goldPerTurn := player.GoldPerTurn()
            foodPerTurn := player.FoodPerTurn()
            manaPerTurn := player.ManaPerTurn(game.ComputePower(player), game)

            elements = append(elements, &uilib.UIElement{
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    goldFood, _ := game.ImageCache.GetImage("main.lbx", 34, 0)
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(240 * data.ScreenScale), float64(77 * data.ScreenScale))
                    screen.DrawImage(goldFood, &options)

                    negativeScale := ebiten.ColorScale{}

                    // v is in range 0.5-1
                    v := (math.Cos(float64(game.Counter) / 7) + 1) / 4 + 0.5
                    negativeScale.SetR(float32(v))

                    if goldPerTurn < 0 {
                        game.InfoFontRed.PrintCenter(screen, float64(278 * data.ScreenScale), float64(103 * data.ScreenScale), float64(data.ScreenScale), negativeScale, fmt.Sprintf("%v Gold", goldPerTurn))
                    } else {
                        game.InfoFontYellow.PrintCenter(screen, float64(278 * data.ScreenScale), float64(103 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("%v Gold", goldPerTurn))
                    }

                    if foodPerTurn < 0 {
                        game.InfoFontRed.PrintCenter(screen, float64(278 * data.ScreenScale), float64(135 * data.ScreenScale), float64(data.ScreenScale), negativeScale, fmt.Sprintf("%v Food", foodPerTurn))
                    } else {
                        game.InfoFontYellow.PrintCenter(screen, float64(278 * data.ScreenScale), float64(135 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("%v Food", foodPerTurn))
                    }

                    if manaPerTurn < 0 {
                        game.InfoFontRed.PrintCenter(screen, float64(278 * data.ScreenScale), float64(167 * data.ScreenScale), float64(data.ScreenScale), negativeScale, fmt.Sprintf("%v Mana", manaPerTurn))
                    } else {
                        game.InfoFontYellow.PrintCenter(screen, float64(278 * data.ScreenScale), float64(167 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("%v Mana", manaPerTurn))
                    }
                },
            })
        }
    }

    elements = append(elements, &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            game.WhiteFont.PrintRight(screen, float64(276 * data.ScreenScale), float64(68 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("%v GP", game.Players[0].Gold))
        },
    })

    elements = append(elements, &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            game.WhiteFont.PrintRight(screen, float64(313 * data.ScreenScale), float64(68 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("%v MP", game.Players[0].Mana))
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

        select {
            case game.Events<- &GameEventMoveUnit{Player: player}:
            default:
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
        enchantment.City.RemoveEnchantment(enchantment.Enchantment.Enchantment, enchantment.Enchantment.Owner)
    }

    // FIXME: dissipate unit enchantments
}

func (game *Game) StartPlayerTurn(player *playerlib.Player) {
    disbandedMessages := game.DisbandUnits(player)

    if player.IsHuman() && len(disbandedMessages) > 0 {
        select {
            case game.Events<- &GameEventScroll{Title: "", Text: strings.Join(disbandedMessages, "\n")}:
            default:
        }
    }

    power := game.ComputePower(player)

    game.DissipateEnchantments(player, power)

    player.Gold += player.GoldPerTurn()
    if player.Gold < 0 {
        player.Gold = 0
    }

    player.Mana += player.ManaPerTurn(power, game)
    if player.Mana < 0 {
        player.Mana = 0
    }

    if !player.CastingSpell.Invalid() {
        // mana spent on the skill is the minimum of {player's mana, casting skill, remaining cost for spell}
        manaSpent := player.Mana
        if manaSpent > player.RemainingCastingSkill {
            manaSpent = player.RemainingCastingSkill
        }

        remainingMana := player.CastingSpell.Cost(true) - player.CastingSpellProgress
        if remainingMana < manaSpent {
            manaSpent = remainingMana
        }

        player.CastingSpellProgress += manaSpent
        player.Mana -= manaSpent

        if player.CastingSpell.Cost(true) <= player.CastingSpellProgress {

            if player.IsHuman() {
                select {
                    case game.Events<- &GameEventCastSpell{Player: player, Spell: player.CastingSpell}:
                    default:
                        log.Printf("Error: unable to invoke cast spell because event queue is full")
                }
            } else {
                game.doCastSpell(player, player.CastingSpell)
            }
            player.CastingSpell = spellbook.Spell{}
            player.CastingSpellProgress = 0
        }
    }

    if player.ResearchingSpell.Valid() {
        // log.Printf("wizard %v power=%v researching=%v progress=%v/%v perturn=%v", player.Wizard.Name, power, player.ResearchingSpell.Name, player.ResearchProgress, player.ResearchingSpell.ResearchCost, player.SpellResearchPerTurn(power))
        player.ResearchProgress += int(player.SpellResearchPerTurn(power))
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
    } else if game.TurnNumber > 1 {

        if player.IsHuman() {
            select {
                case game.Events<- &GameEventResearchSpell{Player: player}:
                default:
            }
        }
    }

    player.CastingSkillPower += player.CastingSkillPerTurn(power)

    // reset casting skill for this turn
    player.RemainingCastingSkill = player.ComputeCastingSkill()

    var removeCities []*citylib.City

    for _, city := range player.Cities {
        cityEvents := city.DoNextTurn(player.GetUnits(city.X, city.Y, city.Plane), game.GetMap(city.Plane))
        for _, event := range cityEvents {
            switch event.(type) {
            case *citylib.CityEventPopulationGrowth:
                growthEvent := event.(*citylib.CityEventPopulationGrowth)

                verb := "grown"
                if !growthEvent.Grow {
                    verb = "shrunk"
                }

                if growthEvent.Size == 0 {
                    removeCities = append(removeCities, city)
                    if player.IsHuman() {
                        select {
                            case game.Events<- &GameEventNotice{Message: fmt.Sprintf("The city of %v has been abandoned.", city.Name)}:
                            default:
                        }
                    }
                } else if player.IsHuman() {
                    scrollEvent := GameEventScroll{
                        Title: "CITY GROWTH",
                        Text: fmt.Sprintf("%v has %v to a population of %v.", city.Name, verb, city.Citizens()),
                    }

                    select {
                        case game.Events<- &scrollEvent:
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
                overworldUnit := units.MakeOverworldUnitFromUnit(newUnit.Unit, city.X, city.Y, city.Plane, city.Banner, player.MakeExperienceInfo())
                // only normal units get weapon bonuses
                if overworldUnit.GetRace() != data.RaceFantastic {
                    overworldUnit.SetWeaponBonus(newUnit.WeaponBonus)
                }
                player.AddUnit(overworldUnit)

                if player.AIBehavior != nil {
                    player.AIBehavior.ProducedUnit(city, player)
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
        }

        // any healer in the same stack provides an additional 20% healing rate
        for _, unit := range stack.Units() {
            if unit.HasAbility(data.AbilityHealer) {
                rate += 0.2
                break
            }
        }

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

    // game.CenterCamera(player.Cities[0].X, player.Cities[0].Y)
    game.DoNextUnit(player)
    if player.IsHuman() {
        game.RefreshUI()
    }
}

func (game *Game) revertVolcanos() {
    mapObjects := []*maplib.Map{game.ArcanusMap, game.MyrrorMap}
    for _, mapObject := range mapObjects {
        for location, extras := range mapObject.ExtraMap {
            _, ok := extras[maplib.ExtraKindVolcano]
            if ok {
                if rand.N(100) < 2 {
                    mapObject.RemoveVolcano(location.X, location.Y)
                }
            }
        }
    }
}

func (game *Game) EndOfTurn() {
    // put stuff here that should happen when all players have taken their turn

    game.revertVolcanos()

    game.TurnNumber += 1
}

func (game *Game) DoNextTurn(){
    game.CurrentPlayer += 1
    if game.CurrentPlayer >= len(game.Players) {
        // all players did their turn, so the next global turn starts
        game.EndOfTurn()
        game.CurrentPlayer = 0
    }

    if len(game.Players) > 0 {
        player := game.Players[game.CurrentPlayer]

        if player.Wizard.Banner != data.BannerBrown {
            game.StartPlayerTurn(player)
        } else {
            // neutral enemies should reset their moves each turn
            for _, stack := range player.Stacks {
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
            screen.DrawImage(overworld.FogBlack, &options)
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
            screen.DrawImage(FogEdge_N_E, &options)
        } else if n {
            screen.DrawImage(FogEdge_N, &options)
        } else if e {
            screen.DrawImage(FogEdge_E, &options)
        } else if ne {
            screen.DrawImage(FogCorner_NE, &options)
        }

        if s && e {
            screen.DrawImage(FogEdge_S_E, &options)
        } else if s {
            screen.DrawImage(FogEdge_S, &options)
        } else if se {
            screen.DrawImage(FogCorner_SE, &options)
        }

        if n && w {
            screen.DrawImage(FogEdge_N_W, &options)
        } else if w {
            screen.DrawImage(FogEdge_W, &options)
        } else if nw {
            screen.DrawImage(FogCorner_NW, &options)
        }

        if s && w {
            screen.DrawImage(FogEdge_S_W, &options)
        } else if sw {
            screen.DrawImage(FogCorner_SW, &options)
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

func (overworld *Overworld) DrawOverworld(screen *ebiten.Image, geom ebiten.GeoM){

    screen.Fill(color.RGBA{R: 32, G: 32, B: 32, A: 0xff})

    tileWidth := overworld.Map.TileWidth()
    tileHeight := overworld.Map.TileHeight()

    geom.Translate(-overworld.Camera.GetZoomedX() * float64(tileWidth), -overworld.Camera.GetZoomedY() * float64(tileHeight))
    geom.Scale(overworld.Camera.GetAnimatedZoom(), overworld.Camera.GetAnimatedZoom())

    overworld.Map.DrawLayer1(overworld.Camera, overworld.Counter / 8, overworld.ImageCache, screen, geom)

    convertTileCoordinates := func(x int, y int) (int, int) {
        outX := x * tileWidth
        outY := y * tileHeight
        return outX, outY
    }

    cityPositions := make(map[image.Point]struct{})

    for _, city := range overworld.Cities {
        var cityPic *ebiten.Image
        var err error
        cityPositions[image.Point{city.X, city.Y}] = struct{}{}
        cityPic, err = GetCityWallImage(city, overworld.ImageCache)
        /*
        if city.Wall {
            cityPic, err = GetCityWallImage(city.GetSize(), overworld.ImageCache)
        } else {
            cityPic, err = GetCityNoWallImage(city.GetSize(), overworld.ImageCache)
        }
        */

        if err == nil {
            var options ebiten.DrawImageOptions

            cityX, cityY := overworld.ToCameraCoordinates(city.X, city.Y)

            x, y := convertTileCoordinates(cityX, cityY)
            // x, y := cityX, cityY
            // options.GeoM = geom
            // draw the city in the center of the tile
            // first compute center of tile
            options.GeoM.Translate(float64(x + tileWidth / 2.0), float64(y + tileHeight / 2.0))
            // then move the city image so that the center of the image is at the center of the tile
            options.GeoM.Translate(float64(-cityPic.Bounds().Dx()) / 2.0, float64(-cityPic.Bounds().Dy()) / 2.0)
            options.GeoM.Concat(geom)
            screen.DrawImage(cityPic, &options)

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
            stackX, stackY := overworld.ToCameraCoordinates(stack.X(), stack.Y())

            // log.Printf("World %v, %v -> camera %v, %v. Camera: %v, %v", stack.X(), stack.Y(), stackX, stackY, overworld.Camera.GetX(), overworld.Camera.GetY())

            // x, y := convertTileCoordinates(stackX, stackY)
            x, y := float64(stackX), float64(stackY)

            // nx := overworld.Map.WrapX(x - overworld.Camera.GetX()) + overworld.Camera.GetX() + 6

            options.GeoM.Translate((x + float64(stack.OffsetX())) * float64(tileWidth), (y + float64(stack.OffsetY())) * float64(tileHeight))
            options.GeoM.Concat(geom)

            leader := stack.Leader()

            unitBack, err := units.GetUnitBackgroundImage(leader.GetBanner(), overworld.ImageCache)
            if err == nil {
                screen.DrawImage(unitBack, &options)
            }

            pic, err := GetUnitImage(leader, overworld.ImageCache, leader.GetBanner())
            if err == nil {
                options.GeoM.Translate(float64(data.ScreenScale), float64(data.ScreenScale))

                if leader.GetBusy() != units.BusyStatusNone {
                    var patrolOptions colorm.DrawImageOptions
                    var matrix colorm.ColorM
                    patrolOptions.GeoM = options.GeoM
                    matrix.ChangeHSV(0, 0, 1)
                    colorm.DrawImage(screen, pic, matrix, &patrolOptions)
                } else {
                    screen.DrawImage(pic, &options)
                }

                enchantment := util.First(leader.GetEnchantments(), data.UnitEnchantmentNone)
                if enchantment != data.UnitEnchantmentNone {
                    x, y := options.GeoM.Apply(0, 0)
                    util.DrawOutline(screen, overworld.ImageCache, pic, x, y, options.ColorScale, overworld.Counter/10, enchantment.Color())
                }
            }

        }
    }

    overworld.Map.DrawLayer2(overworld.Camera, overworld.Counter / 8, overworld.ImageCache, screen, geom)

    if overworld.Fog != nil {
        overworld.DrawFog(screen, geom)
    }

    // draw current path on top of fog
    if overworld.SelectedStack != nil {
        boot, _ := overworld.ImageCache.GetImage("compix.lbx", 72, 0)
        for pointI, point := range overworld.SelectedStack.CurrentPath {
            var options ebiten.DrawImageOptions
            x, y := convertTileCoordinates(overworld.ToCameraCoordinates(point.X, point.Y))
            options.GeoM.Translate(float64(x), float64(y))
            options.GeoM.Translate(float64(tileWidth) / 2, float64(tileHeight) / 2)
            options.GeoM.Translate(float64(boot.Bounds().Dx()) / -2, float64(boot.Bounds().Dy()) / -2)
            options.GeoM.Concat(geom)

            v := float32(1 + (math.Sin(float64(overworld.Counter * 4 + uint64(pointI) * 60) * math.Pi / 180) / 2 + 0.5) / 2)
            options.ColorScale.Scale(v, v, v, 1)

            screen.DrawImage(boot, &options)
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
    if data.ScreenScale == 1 && game.Camera.GetZoom() < 0.9 {
        useCounter = 1
    }

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

    overworldScreen := screen.SubImage(image.Rect(0, 18 * data.ScreenScale, 240 * data.ScreenScale, data.ScreenHeight)).(*ebiten.Image)
    overworld.DrawOverworld(overworldScreen, ebiten.GeoM{})

    var miniGeom ebiten.GeoM
    miniGeom.Translate(float64(250 * data.ScreenScale), float64(20 * data.ScreenScale))
    mx, my := miniGeom.Apply(0, 0)
    miniWidth := 60 * data.ScreenScale
    miniHeight := 31 * data.ScreenScale
    mini := screen.SubImage(image.Rect(int(mx), int(my), int(mx) + miniWidth, int(my) + miniHeight)).(*ebiten.Image)
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
