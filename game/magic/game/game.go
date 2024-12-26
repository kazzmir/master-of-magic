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

    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/mirror"
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
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    mouselib "github.com/kazzmir/master-of-magic/lib/mouse"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/fraction"
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
    game.Fog.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 0xff})
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

type GameEventHireHero struct {
    Hero *herolib.Hero
    Player *playerlib.Player
    Cost int
}

type GameEventVault struct {
    CreatedArtifact *artifact.Artifact
}

type GameEventNewOutpost struct {
    City *citylib.City
    Stack *playerlib.UnitStack
}

type GameEventLearnedSpell struct {
    Player *playerlib.Player
    Spell spellbook.Spell
}

type GameEventResearchSpell struct {
    Player *playerlib.Player
}

type GameEventLoadMenu struct {
}

type GameEventCastSpell struct {
    Player *playerlib.Player
    Spell spellbook.Spell
}

type GameEventSummonUnit struct {
    Wizard data.WizardBase
    Unit units.Unit
}

type GameEventSummonArtifact struct {
    Wizard data.WizardBase
}

type GameEventSummonHero struct {
    Wizard data.WizardBase
    Champion bool
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

    MouseData *mouselib.MouseData

    Events chan GameEvent
    BuildingInfo buildinglib.BuildingInfos

    MovingStack *playerlib.UnitStack

    cameraX int
    cameraY int

    HudUI *uilib.UI
    Help lbx.Help

    ArcanusMap *maplib.Map
    MyrrorMap *maplib.Map

    Players []*playerlib.Player
    CurrentPlayer int
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
    }

    return powers
}

// a true value in fog means the tile is visible, false means not visible
func (game *Game) MakeFog() [][]bool {
    fog := make([][]bool, game.CurrentMap().Width())
    for x := 0; x < game.CurrentMap().Width(); x++ {
        fog[x] = make([]bool, game.CurrentMap().Height())
    }

    return fog
}

func (game *Game) CenterCamera(x int, y int){
    game.cameraX = x - 5
    game.cameraY = y - 5

    /*
    if game.cameraX < 0 {
        game.cameraX = 0
    }
    */

    if game.cameraY < 0 {
        game.cameraY = 0
    }

    if game.cameraY >= game.CurrentMap().Height() - 11 {
        game.cameraY = game.CurrentMap().Height() - 11
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
    newPlayer := playerlib.MakePlayer(wizard, human, game.MakeFog(), game.MakeFog())

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

    help, err := helpLbx.ReadHelp(2)
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
        color.White, color.White, color.White, color.White,
    }

    whiteFont := font.MakeOptimizedFontWithPalette(fonts[0], whitePalette)

    buildingInfo, err := buildinglib.ReadBuildingInfo(lbxCache)
    if err != nil {
        log.Printf("Unable to read building info: %v", err)
        return nil
    }

    mouseData, err := mouselib.MakeMouseData(lbxCache)
    if err != nil {
        log.Printf("Unable to read mouse data: %v", err)
        return nil
    }

    game := &Game{
        Cache: lbxCache,
        Help: help,
        MouseData: mouseData,
        Events: make(chan GameEvent, 1000),
        ArcanusMap: maplib.MakeMap(terrainData, settings.LandSize, data.PlaneArcanus),
        MyrrorMap: maplib.MakeMap(terrainData, settings.LandSize, data.PlaneMyrror),
        Plane: data.PlaneArcanus,
        State: GameStateRunning,
        Settings: settings,
        ImageCache: util.MakeImageCache(lbxCache),
        InfoFontYellow: infoFontYellow,
        InfoFontRed: infoFontRed,
        Heroes: createHeroes(),
        WhiteFont: whiteFont,
        BuildingInfo: buildingInfo,
        TurnNumber: 1,
        CurrentPlayer: -1,
    }

    game.HudUI = game.MakeHudUI()
    game.Drawer = func(screen *ebiten.Image, game *Game){
        game.DrawGame(screen)
    }

    return game
}

func (game *Game) NearCity(point image.Point, squares int) bool {
    for _, player := range game.Players {
        for _, city := range player.Cities {
            if city.Plane == game.Plane {
                diff := image.Pt(city.X, city.Y).Sub(point)

                total := int(math.Abs(float64(diff.X)) + math.Abs(float64(diff.Y)))

                if total <= squares {
                    return true
                }
            }
        }
    }

    return false
}

func (game *Game) FindValidCityLocation() (int, int) {
    mapUse := game.CurrentMap()
    continents := mapUse.Map.FindContinents(mapUse.Plane)

    for i := 0; i < 10; i++ {
        continentIndex := rand.IntN(len(continents))
        continent := continents[continentIndex]
        if len(continent) > 100 {
            index := rand.IntN(len(continent))
            x := continent[index].X
            y := continent[index].Y

            if mapUse.Map.Terrain[x][y] == terrain.TileLand.Index(mapUse.Plane) {
                return x, y
            }
        }
    }

    return 0, 0
}

func (game *Game) FindValidCityLocationOnContinent(x int, y int) (int, int) {
    mapUse := game.CurrentMap()
    continents := mapUse.Map.FindContinents(mapUse.Plane)

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
    for i := 0; i < entriesPerRace; i++ {
        name := predefinedNames[entriesPerRace * raceIndex + i]
        if !slices.Contains(existingNames, name) {
            return name
        }
    }
    
    for _, name := range predefinedNames {
        if !slices.Contains(existingNames, name) {
            return name
        }
    }

    return fallback
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

    citiesMiniMap := make([]maplib.MiniMapCity, 0, len(cities))
    for _, city := range cities {
        citiesMiniMap = append(citiesMiniMap, city)
    }

    drawMinimap := func (screen *ebiten.Image, x int, y int, fog [][]bool, counter uint64){
        game.CurrentMap().DrawMinimap(screen, citiesMiniMap, x, y, fog, counter, false)
    }

    var showCity *citylib.City
    selectCity := func(city *citylib.City){
        // ignore outpost
        if city.Citizens() >= 1 {
            showCity = city
        }
        game.CenterCamera(city.X, city.Y)
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

    drawMinimap := func (screen *ebiten.Image, x int, y int, fog [][]bool, counter uint64){
        game.CurrentMap().DrawMinimap(screen, citiesMiniMap, x, y, fog, counter, false)
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
        power += float64(city.ComputePower())
    }

    power += float64(player.Wizard.TotalBooks())

    magicBonus := float64(1)

    switch game.Settings.Magic {
        case data.MagicSettingWeak: magicBonus = 0.5
        case data.MagicSettingNormal: magicBonus = 1
        case data.MagicSettingPowerful: magicBonus = 1.5
    }

    for _, node := range game.ArcanusMap.GetMeldedNodes(player) {
        power += float64(len(node.Zone)) * magicBonus
    }

    for _, node := range game.MyrrorMap.GetMeldedNodes(player) {
        power += float64(len(node.Zone)) * magicBonus
    }

    if power < 0 {
        power = 0
    }

    return int(power)
}

// enemy wizards, but not including the raider ai
func (game *Game) GetEnemyWizards() []*playerlib.Player {
    var out []*playerlib.Player

    for _, player := range game.Players {
        if !player.Human && player.Wizard.Banner != data.BannerBrown {
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

    maxLength := float64(84)

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

            for len(name) > 0 && nameFont.MeasureTextWidth(name, 1) > maxLength {
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
            options.GeoM.Translate(float64(topX), float64(topY))
            screen.DrawImage(background, &options)

            x, y := options.GeoM.Apply(13, 20)

            nameFont.Print(screen, x, y, 1, options.ColorScale, name)

            tx, ty := options.GeoM.Apply(9, 6)
            titleFont.Print(screen, tx, ty, 1, options.ColorScale, title)

            // draw cursor
            cursorX := x + nameFont.MeasureTextWidth(name, 1)

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

func (game *Game) showNewBuilding(yield coroutine.YieldFunc, city *citylib.City, building buildinglib.Building){
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
    snake, _ := game.ImageCache.GetImageTransform("resource.lbx", 54, 0, "crop", util.AutoCrop)

    wrappedText := bigFont.CreateWrappedText(180, 1, fmt.Sprintf("The %s of %s has completed the construction of a %s.", city.GetSize(), city.Name, game.BuildingInfo.Name(building)))

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
        options.GeoM.Translate(8, 60)
        screen.DrawImage(background, &options)
        iconOptions := options
        iconOptions.GeoM.Translate(6, -10)
        screen.DrawImage(snake, &iconOptions)

        x, y := options.GeoM.Apply(8 + float64(snake.Bounds().Dx()), 9)
        bigFont.RenderWrapped(screen, x, y, wrappedText, options.ColorScale, false)

        options.GeoM.Translate(float64(background.Bounds().Dx()), 0)
        screen.DrawImage(rightSide, &options)

        x, y = options.GeoM.Apply(4, 6)
        buildingSpace := screen.SubImage(image.Rect(int(x), int(y), int(x + 45), int(y + 47))).(*ebiten.Image)

        // vector.DrawFilledRect(buildingSpace, float32(x), float32(y), float32(buildingSpace.Bounds().Dx()), float32(buildingSpace.Bounds().Dy()), color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, false)

        landOptions := options
        landOptions.GeoM.Translate(-10, -10)
        buildingSpace.DrawImage(landBackground, &landOptions)

        buildingOptions := options
        buildingOptions.GeoM.Translate(float64(buildingSpace.Bounds().Dx()) / 2, float64(buildingSpace.Bounds().Dy()) / 2)
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
    wrappedText := smallFont.CreateWrappedText(180, 1, text)

    scrollImages, _ := game.ImageCache.GetImages("scroll.lbx", 2)

    totalImages := int((wrappedText.TotalHeight + float64(bigFont.Height())) / 5) + 1

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

        options.GeoM.Translate(65, 25)

        middleY := pageBackground.Bounds().Dy() / 2
        length := scrollLength / 2
        if length > middleY {
            length = middleY
        }
        pagePart := pageBackground.SubImage(image.Rect(0, middleY - length, pageBackground.Bounds().Dx(), middleY + length)).(*ebiten.Image)

        pageOptions := options
        pageOptions.GeoM.Translate(0, float64(middleY - length) + 5)
        screen.DrawImage(pagePart, &pageOptions)

        x, y := options.GeoM.Apply(float64(pageBackground.Bounds().Dx()) / 2, float64(middleY) - wrappedText.TotalHeight / 2 - float64(bigFont.Height()) / 2 + 5)
        bigFont.PrintCenter(screen, x, y, 1, options.ColorScale, title)
        y += float64(bigFont.Height()) + 1
        smallFont.RenderWrapped(screen, x, y, wrappedText, options.ColorScale, true)

        scrollOptions := options
        scrollOptions.GeoM.Translate(-63, -20)
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
        options.GeoM.Translate(30, 50)
        screen.DrawImage(background, &options)

        numHouses := city.GetOutpostHouses()
        maxHouses := 10

        houseOptions := options
        houseOptions.GeoM.Translate(7, 31)

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
            houseOptions.GeoM.Translate(float64(emptyHouse.Bounds().Dx()) + 1, 0)
        }

        if stack != nil {
            stackOptions := options
            stackOptions.GeoM.Translate(7, 55)

            for _, unit := range stack.Units() {
                pic, _ := GetUnitImage(unit, &game.ImageCache, city.Banner)
                screen.DrawImage(pic, &stackOptions)
                stackOptions.GeoM.Translate(float64(pic.Bounds().Dx()) + 1, 0)
            }
        }

        x, y := options.GeoM.Apply(6, 22)
        game.InfoFontYellow.Print(screen, x, y, 1, options.ColorScale, city.Race.String())

        x, y = options.GeoM.Apply(20, 5)
        if rename {
            bigFont.Print(screen, x, y, 1, options.ColorScale, "New Outpost Founded")
        } else {
            bigFont.Print(screen, x, y, 1, options.ColorScale, fmt.Sprintf("Outpost Of %v", city.Name))
        }

        cityScapeOptions := options
        cityScapeOptions.GeoM.Translate(185, 30)
        x, y = cityScapeOptions.GeoM.Apply(0, 0)
        cityScape := screen.SubImage(image.Rect(int(x), int(y), int(x + 72), int(y + 66))).(*ebiten.Image)

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
        options2.GeoM.Translate(30, 20)
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

func (game *Game) showMovement(yield coroutine.YieldFunc, oldX int, oldY int, stack *playerlib.UnitStack){
    drawer := game.Drawer
    defer func(){
        game.Drawer = drawer
    }()

    // the number of frames it takes to move a unit one tile
    frames := 10

    tileWidth := float64(game.CurrentMap().TileWidth())
    tileHeight := float64(game.CurrentMap().TileHeight())

    convertTileCoordinates := func(x float64, y float64) (float64, float64) {
        outX := (x - float64(game.cameraX)) * tileWidth
        outY := (y - float64(game.cameraY)) * tileHeight
        return outX, outY
    }

    dx := float64(oldX - stack.X())
    dy := float64(oldY - stack.Y())

    game.State = GameStateUnitMoving

    game.MovingStack = stack

    boot, _ := game.ImageCache.GetImage("compix.lbx", 72, 0)

    game.Drawer = func (screen *ebiten.Image, game *Game){
        drawer(screen, game)

        // draw boot images on the map that show where the unit is moving to
        for _, point := range stack.CurrentPath {
            var options ebiten.DrawImageOptions
            x, y := convertTileCoordinates(float64(point.X), float64(point.Y))
            options.GeoM.Translate(x, y)
            options.GeoM.Translate(float64(tileWidth) / 2, float64(tileHeight) / 2)
            options.GeoM.Translate(float64(boot.Bounds().Dx()) / -2, float64(boot.Bounds().Dy()) / -2)
            screen.DrawImage(boot, &options)
        }

    }

    for i := 0; i < frames; i++ {
        game.Counter += 1
        stack.SetOffset(dx * float64(frames - i) / float64(frames), dy * float64(frames - i) / float64(frames))
        yield()
    }

    game.State = GameStateRunning
    game.MovingStack = nil

    stack.SetOffset(0, 0)
    game.CenterCamera(stack.X(), stack.Y())
}

/* return the cost to move from the current position the stack is on to the new given coordinates.
 * also return true/false if the move is even possible
 */
func (game *Game) ComputeTerrainCost(stack *playerlib.UnitStack, x int, y int, mapUse *maplib.Map) (fraction.Fraction, bool) {
    if stack.OutOfMoves() {
        return fraction.Zero(), false
    }

    tileFrom := mapUse.GetTile(stack.X(), stack.Y())
    tileTo := mapUse.GetTile(x, y)

    // can't move from land to ocean unless all units are flyers
    if tileFrom.Tile.IsLand() && !tileTo.Tile.IsLand() {
        if !stack.AllFlyers() {
            return fraction.Zero(), false
        }
    }

    oldX := stack.X()
    oldY := stack.Y()

    xDiff := int(math.Abs(float64(x - oldX)))
    yDiff := int(math.Abs(float64(y - oldY)))

    if xDiff == 1 && yDiff == 1 {
        return fraction.Make(3, 2), true
    }

    if xDiff == 1 || yDiff == 1 {
        return fraction.FromInt(1), true
    }

    return fraction.Zero(), false
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

func (game *Game) FindPath(oldX int, oldY int, newX int, newY int, stack *playerlib.UnitStack, fog [][]bool) pathfinding.Path {

    useMap := game.GetMap(stack.Plane())

    tileCost := func (x1 int, y1 int, x2 int, y2 int) float64 {
        x1 = useMap.WrapX(x1)
        x2 = useMap.WrapX(x2)

        if x1 < 0 || x1 >= useMap.Width() || y1 < 0 || y1 >= useMap.Height() {
            return pathfinding.Infinity
        }

        if x2 < 0 || x2 >= useMap.Width() || y2 < 0 || y2 >= useMap.Height() {
            return pathfinding.Infinity
        }

        tileFrom := useMap.GetTile(x1, y1)
        tileTo := useMap.GetTile(x2, y2)

        // FIXME: consider terrain type, roads, and unit abilities

        baseCost := float64(1)

        if x1 != x2 && y1 != y2 {
            baseCost = 1.5
        }

        // don't know what the cost is, assume we can move there
        if x2 >= 0 && x2 < len(fog) && y2 >= 0 && y2 < len(fog[x2]) && !fog[x2][y2] {
            return baseCost
        }

        // can't move from land to ocean unless all units are flyers
        if tileFrom.Tile.IsLand() && !tileTo.Tile.IsLand() {
            if !stack.AllFlyers() {
                return pathfinding.Infinity
            }
        }

        return baseCost
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

    path, ok := pathfinding.FindPath(image.Pt(oldX, oldY), image.Pt(newX, newY), 10000, tileCost, neighbors)
    if ok {
        return path[1:]
    }

    return nil
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

func (game *Game) doLoadMenu(yield coroutine.YieldFunc) {
    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    quit := false

    imageCache := util.MakeImageCache(game.Cache)

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            background, _ := imageCache.GetImage("load.lbx", 0, 0)
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
            Rect: util.ImageRect(x, y, useImage),
            LeftClick: func(element *uilib.UIElement){
                action()
                quit = true
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(x), float64(y))
                screen.DrawImage(useImage, &options)
            },
        }
    }

    // quit
    elements = append(elements, makeButton(2, 43, 171, func(){
        game.State = GameStateQuit
    }))

    // load
    elements = append(elements, makeButton(1, 83, 171, func(){
    }))

    // save
    elements = append(elements, makeButton(3, 122, 171, func(){
    }))

    // settings
    elements = append(elements, makeButton(12, 172, 171, func(){
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
    chance := (3 + player.Fame / 25) / ((len(player.AliveHeroes()) + 3) / 2)
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
                if hero.GetRequiredFame() <= player.Fame {
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

    game.HudUI.AddElements(MakeHireScreenUI(game.Cache, game.HudUI, hero, cost, result))

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

        game.HudUI.AddElements(uilib.MakeConfirmDialog(game.HudUI, game.Cache, &game.ImageCache, message, yes, no))

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
                        if hire.Player.Human {
                            game.doHireHero(yield, hire.Cost, hire.Hero, hire.Player)
                        }
                    case *GameEventNextTurn:
                        game.doNextTurn(yield)
                    case *GameEventSurveyor:
                        game.doSurveyor(yield)
                    case *GameEventApprenticeUI:
                        game.ShowApprenticeUI(yield, game.Players[0])
                    case *GameEventArmyView:
                        game.doArmyView(yield)
                    case *GameEventNotice:
                        notice := event.(*GameEventNotice)
                        game.doNotice(yield, notice.Message)
                    case *GameEventCastSpellBook:
                        game.ShowSpellBookCastUI(yield, game.Players[0])
                    case *GameEventCityListView:
                        game.doCityListView(yield)
                    case *GameEventNewOutpost:
                        outpost := event.(*GameEventNewOutpost)
                        game.showOutpost(yield, outpost.City, outpost.Stack, true)
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
                    case *GameEventCastSpell:
                        castSpell := event.(*GameEventCastSpell)
                        // in cast.go
                        game.doCastSpell(yield, castSpell.Player, castSpell.Spell)
                    case *GameEventNewBuilding:
                        buildingEvent := event.(*GameEventNewBuilding)
                        game.CenterCamera(buildingEvent.City.X, buildingEvent.City.Y)
                        game.showNewBuilding(yield, buildingEvent.City, buildingEvent.Building)
                        game.doCityScreen(yield, buildingEvent.City, buildingEvent.Player, buildingEvent.Building)
                    case *GameEventCityName:
                        cityEvent := event.(*GameEventCityName)
                        city := cityEvent.City
                        city.Name = game.doInput(yield, cityEvent.Title, city.Name, cityEvent.X, cityEvent.Y)
                    case *GameEventSummonUnit:
                        summonUnit := event.(*GameEventSummonUnit)
                        game.doSummon(yield, summon.MakeSummonUnit(game.Cache, summonUnit.Unit, summonUnit.Wizard))
                    case *GameEventSummonArtifact:
                        summonArtifact := event.(*GameEventSummonArtifact)
                        game.doSummon(yield, summon.MakeSummonArtifact(game.Cache, summonArtifact.Wizard))
                    case *GameEventSummonHero:
                        summonHero := event.(*GameEventSummonHero)
                        game.doSummon(yield, summon.MakeSummonHero(game.Cache, summonHero.Wizard, summonHero.Champion))
                    case *GameEventLoadMenu:
                        game.doLoadMenu(yield)
                }
            default:
                return
        }
    }
}

/* returns a wizard definition and true if successful, otherwise false if no more wizards can be created
 */
func (game *Game) ChooseWizard() (setup.WizardCustom, bool) {
    // pick a new wizard with an unused wizard base and banner color, and race
    // if on myrror then select a myrran race

    chooseBase := func() (setup.WizardSlot, bool) {
        choices := slices.Clone(setup.DefaultWizardSlots())
        choices = slices.DeleteFunc(choices, func (wizard setup.WizardSlot) bool {
            for _, player := range game.Players {
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
            for _, player := range game.Players {
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
            for _, player := range game.Players {
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

    customWizard.StartingSpells.AddAllSpells(setup.GetStartingSpells(&customWizard, game.AllSpells()))
    return customWizard, true
}

func (game *Game) RefreshUI() {
    select {
        case game.Events <- &GameEventRefreshUI{}:
        default:
    }
}

func (game *Game) doPlayerUpdate(yield coroutine.YieldFunc, player *playerlib.Player) {
    // log.Printf("Game.Update")
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    if player.SelectedStack != nil {
        stack := player.SelectedStack
        mapUse := game.GetMap(stack.Plane())
        oldX := stack.X()
        oldY := stack.Y()

        if len(stack.CurrentPath) == 0 {

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

            newX := stack.X() + dx
            newY := stack.Y() + dy

            leftClick := inputmanager.LeftClick()
            if leftClick {
                mouseX, mouseY := inputmanager.MousePosition()

                // can only click into the area not hidden by the hud
                if mouseX < 240 && mouseY > 18 {
                    // log.Printf("Click at %v, %v", mouseX, mouseY)
                    newX = game.cameraX + mouseX / mapUse.TileWidth()
                    newY = game.cameraY + mouseY / mapUse.TileHeight()
                }
            }

            if newX != oldX || newY != oldY {
                activeUnits := stack.ActiveUnits()
                if len(activeUnits) > 0 {
                    if newY > 0 && newY < mapUse.Height() {

                        inactiveUnits := stack.InactiveUnits()
                        if len(inactiveUnits) > 0 {
                            stack.RemoveUnits(inactiveUnits)
                            player.AddStack(playerlib.MakeUnitStackFromUnits(inactiveUnits))
                            game.RefreshUI()
                        }

                        path := game.FindPath(oldX, oldY, newX, newY, stack, player.GetFog(game.Plane))
                        if path == nil {
                            game.blinkRed(yield)
                        } else {
                            stack.CurrentPath = path
                        }
                    }
                }
            }
        }

        stepsTaken := 0
        stopMoving := false
        var mergeStack *playerlib.UnitStack

        quitMoving:
        for i, step := range stack.CurrentPath {
            if stack.OutOfMoves() {
                break
            }

            terrainCost, canMove := game.ComputeTerrainCost(stack, step.X, step.Y, mapUse)

            if canMove {
                node := mapUse.GetMagicNode(step.X, step.Y)
                if node != nil && !node.Empty {
                    if game.confirmMagicNodeEncounter(yield, node) {

                        stack.Move(step.X - stack.X(), step.Y - stack.Y(), terrainCost)
                        game.showMovement(yield, oldX, oldY, stack)
                        player.LiftFog(stack.X(), stack.Y(), 2, stack.Plane())

                        game.doMagicEncounter(yield, player, stack, node)

                        game.RefreshUI()
                    }

                    stopMoving = true
                    break quitMoving
                }

                lair := mapUse.GetLair(step.X, step.Y)
                if lair != nil {
                    if game.confirmLairEncounter(yield, lair) {
                        stack.Move(step.X - stack.X(), step.Y - stack.Y(), terrainCost)
                        game.showMovement(yield, oldX, oldY, stack)
                        player.LiftFog(stack.X(), stack.Y(), 2, stack.Plane())

                        game.doLairEncounter(yield, player, stack, lair)

                        game.RefreshUI()
                    }

                    stopMoving = true
                    break quitMoving
                }

                stepsTaken = i + 1
                mergeStack = player.FindStack(step.X, step.Y)

                stack.Move(step.X - stack.X(), step.Y - stack.Y(), terrainCost)
                game.showMovement(yield, oldX, oldY, stack)
                player.LiftFog(stack.X(), stack.Y(), 2, stack.Plane())

                for _, otherPlayer := range game.Players[1:] {
                    otherStack := otherPlayer.FindStack(stack.X(), stack.Y())
                    if otherStack != nil {
                        zone := combat.ZoneType{
                            City: otherPlayer.FindCity(stack.X(), stack.Y()),
                        }

                        game.doCombat(yield, player, stack, otherPlayer, otherStack, zone)

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
                    stopMoving = true
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

        if mergeStack != nil {
            stack = player.MergeStacks(mergeStack, stack)
            player.SelectedStack = stack
            game.RefreshUI()
        }

        // update unrest for new units in the city
        newCity := player.FindCity(stack.X(), stack.Y())
        if newCity != nil {
            newCity.UpdateUnrest(stack.Units())
        }

        if stepsTaken > 0 && stack.OutOfMoves() {
            game.DoNextUnit(player)
        }
    }

    rightClick := inputmanager.RightClick()
    if rightClick {
        mapUse := game.CurrentMap()
        mouseX, mouseY := inputmanager.MousePosition()

        // can only click into the area not hidden by the hud
        if mouseX < 240 && mouseY > 18 {
            // log.Printf("Click at %v, %v", mouseX, mouseY)
            tileX := game.CurrentMap().WrapX(game.cameraX + mouseX / mapUse.TileWidth())
            tileY := game.cameraY + mouseY / mapUse.TileHeight()

            game.CenterCamera(tileX, tileY)

            city := player.FindCity(tileX, tileY)
            if city != nil {
                if city.Outpost {
                    game.showOutpost(yield, city, player.FindStack(city.X, city.Y), false)
                } else {
                    game.doCityScreen(yield, city, player, buildinglib.BuildingNone)
                }
                game.RefreshUI()
            } else {
                stack := player.FindStack(tileX, tileY)
                if stack != nil {
                    player.SelectedStack = stack
                    game.RefreshUI()
                } else {

                    for _, otherPlayer := range game.Players {
                        if otherPlayer == player {
                            continue
                        }

                        city := otherPlayer.FindCity(tileX, tileY)
                        if city != nil {
                            game.doEnemyCityView(yield, city, otherPlayer)
                        }

                        enemyStack := otherPlayer.FindStack(tileX, tileY)
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

                if player.Human {
                    if game.HudUI.GetHighestLayerValue() == 0 {
                        game.doPlayerUpdate(yield, player)
                    }
                } else {
                    log.Printf("AI year %v: make decisions", game.TurnNumber)

                    var decisions []playerlib.AIDecision

                    if player.AIBehavior != nil {
                        decisions = player.AIBehavior.Update(player, game.GetEnemies(player), game)
                        log.Printf("AI Decisions: %v", decisions)

                        for _, decision := range decisions {
                            switch decision.(type) {
                                case *playerlib.AIMoveStackDecision:
                                    moveDecision := decision.(*playerlib.AIMoveStackDecision)
                                    stack := moveDecision.Stack
                                    to := moveDecision.Location
                                    log.Printf("  moving stack %v to %v, %v", stack, to.X, to.Y)
                                    terrainCost, _ := game.ComputeTerrainCost(stack, to.X, to.Y, game.GetMap(stack.Plane()))
                                    oldX := stack.X()
                                    oldY := stack.Y()
                                    stack.Move(to.X - stack.X(), to.Y - stack.Y(), terrainCost)
                                    game.showMovement(yield, oldX, oldY, stack)
                                    player.LiftFog(stack.X(), stack.Y(), 2, stack.Plane())

                                    for _, enemy := range game.GetEnemies(player) {
                                        enemyStack := enemy.FindStack(stack.X(), stack.Y())
                                        if enemyStack != nil {
                                            zone := combat.ZoneType{
                                                City: enemy.FindCity(stack.X(), stack.Y()),
                                            }
                                            game.doCombat(yield, player, stack, enemy, enemyStack, zone)
                                        }
                                    }
                                case *playerlib.AICreateUnitDecision:
                                    create := decision.(*playerlib.AICreateUnitDecision)
                                    log.Printf("ai creating %+v", create)

                                    existingStack := player.FindStack(create.X, create.Y)
                                    if existingStack == nil || len(existingStack.Units()) < 9 {
                                        overworldUnit := units.MakeOverworldUnitFromUnit(create.Unit, create.X, create.Y, create.Plane, player.Wizard.Banner, player.MakeExperienceInfo())
                                        player.AddUnit(overworldUnit)
                                    }
                            }
                        }
                    }

                    if len(decisions) == 0 {
                        game.DoNextTurn()
                    }
                }
            }
    }

    return game.State
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

func (game *Game) doEnemyCityView(yield coroutine.YieldFunc, city *citylib.City, player *playerlib.Player){
    drawer := game.Drawer
    defer func(){
        game.Drawer = drawer
    }()

    logic, draw := cityview.SimplifiedView(game.Cache, city, player)

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
    var fog [][]bool

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
        CameraX: city.X - 2,
        CameraY: city.Y - 2,
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

    game.Drawer = oldDrawer
}

func (game *Game) confirmMagicNodeEncounter(yield coroutine.YieldFunc, node *maplib.ExtraMagicNode) bool {
    reloadLbx, err := game.Cache.GetLbxFile("reload.lbx")
    if err != nil {
        return false
    }

    lairIndex := 11
    nodeName := "nature"

    switch node.Kind {
        case maplib.MagicNodeChaos:
            lairIndex = 10
            nodeName = "chaos"
        case maplib.MagicNodeNature:
            lairIndex = 11
            nodeName = "nature"
        case maplib.MagicNodeSorcery:
            lairIndex = 12
            nodeName = "sorcery"
    }

    guardianName := ""
    if len(node.Guardians) > 0 {
        guardianName = node.Guardians[0].Name
    }

    rotateIndexLow := 247
    rotateIndexHigh := 254

    animation := util.MakePaletteRotateAnimation(reloadLbx, lairIndex, rotateIndexLow, rotateIndexHigh)

    return game.confirmEncounter(yield, fmt.Sprintf("You have found a %v node. Scouts have spotted %v within the %v node. Do you wish to enter?", nodeName, guardianName, nodeName), animation)
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

func (game *Game) confirmLairEncounter(yield coroutine.YieldFunc, encounter *maplib.ExtraEncounter) bool {
    lairIndex := 13
    encounterName := ""
    article := ""

    switch encounter.Type {
        case maplib.EncounterTypeLair:
            lairIndex = 17
            encounterName = "monster lair"
            article = "a"
        case maplib.EncounterTypeCave:
            lairIndex = 17
            encounterName = "mysterious cave"
            article = "a"
        case maplib.EncounterTypePlaneTower:
            lairIndex = 9
            encounterName = "tower"
            article = "a"
        case maplib.EncounterTypeAncientTemple:
            lairIndex = 15
            encounterName = "ancient temple"
            article = "a"
        case maplib.EncounterTypeFallenTemple:
            lairIndex = 19
            encounterName = "fallen temple"
            article = "a"
        case maplib.EncounterTypeRuins:
            lairIndex = 18
            encounterName = "ruins"
            article = "some"
        case maplib.EncounterTypeAbandonedKeep:
            lairIndex = 16
            encounterName = "abandoned keep"
            article = "an"
        case maplib.EncounterTypeDungeon:
            lairIndex = 14
            encounterName = "dungeon"
            article = "a"
    }

    guardianName := ""
    if len(encounter.Units) > 0 {
        guardianName = encounter.Units[0].Name
    }

    pic, _ := game.ImageCache.GetImage("reload.lbx", lairIndex, 0)

    return game.confirmEncounter(yield, fmt.Sprintf("You have found %v %v. Scouts have spotted %v within the %v. Do you wish to enter?", article, encounterName, guardianName, encounterName), util.MakeAnimation([]*ebiten.Image{pic}, true))
}

func (game *Game) doLairEncounter(yield coroutine.YieldFunc, player *playerlib.Player, stack *playerlib.UnitStack, encounter *maplib.ExtraEncounter){
    defender := playerlib.Player{
        Wizard: setup.WizardCustom{
            Name: "Node",
        },
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
    }

    result := game.doCombat(yield, player, stack, &defender, playerlib.MakeUnitStackFromUnits(enemies), zone)
    if result == combat.CombatStateAttackerWin {
        // FIXME: give treasure
    } else {
        // FIXME: remove killed defenders
    }

    // absorb extra clicks
    yield()
}

func (game *Game) doMagicEncounter(yield coroutine.YieldFunc, player *playerlib.Player, stack *playerlib.UnitStack, node *maplib.ExtraMagicNode){

    defender := playerlib.Player{
        Wizard: setup.WizardCustom{
            Name: "Node",
        },
    }

    var enemies []units.StackUnit

    for _, unit := range node.Guardians {
        enemies = append(enemies, units.MakeOverworldUnit(unit))
    }

    for _, unit := range node.Secondary {
        enemies = append(enemies, units.MakeOverworldUnit(unit))
    }

    zone := combat.ZoneType{
    }

    switch node.Kind {
        case maplib.MagicNodeNature: zone.NatureNode = true
        case maplib.MagicNodeSorcery: zone.SorceryNode = true
        case maplib.MagicNodeChaos: zone.ChaosNode = true
    }

    result := game.doCombat(yield, player, stack, &defender, playerlib.MakeUnitStackFromUnits(enemies), zone)
    if result == combat.CombatStateAttackerWin {
        // node should have no guardians
        node.Empty = true

        // FIXME: give treasure
    }

    // absorb extra clicks
    yield()
}

func (game *Game) GetCombatLandscape(x int, y int, plane data.Plane) combat.CombatLandscape {
    tile := game.GetMap(plane).GetTile(x, y)

    switch tile.Tile.TerrainType() {
        case terrain.Land, terrain.Hill, terrain.Grass,
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
 */
func (game *Game) doCombat(yield coroutine.YieldFunc, attacker *playerlib.Player, attackerStack *playerlib.UnitStack, defender *playerlib.Player, defenderStack *playerlib.UnitStack, zone combat.ZoneType) combat.CombatState {
    defer mouse.Mouse.SetImage(game.MouseData.Normal)

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

    landscape := game.GetCombatLandscape(attackerStack.X(), attackerStack.Y(), attackerStack.Plane())

    // FIXME: take plane into account for the landscape/terrain
    combatScreen := combat.MakeCombatScreen(game.Cache, &defendingArmy, &attackingArmy, game.Players[0], landscape, attackerStack.Plane(), zone)
    oldDrawer := game.Drawer

    // ebiten.SetCursorMode(ebiten.CursorModeHidden)

    game.Drawer = func (screen *ebiten.Image, game *Game){
        combatScreen.Draw(screen)
    }

    state := combat.CombatStateRunning
    for state == combat.CombatStateRunning {
        state = combatScreen.Update(yield)
        yield()
    }

    endScreen := combat.MakeCombatEndScreen(game.Cache, combatScreen, state == combat.CombatStateAttackerWin)
    game.Drawer = func (screen *ebiten.Image, game *Game){
        endScreen.Draw(screen)
    }

    state2 := combat.CombatEndScreenRunning
    for state2 == combat.CombatEndScreenRunning {
        state2 = endScreen.Update()
        yield()
    }

    if state == combat.CombatStateAttackerWin {
        for _, unit := range attackerStack.Units() {
            if unit.GetRace() != data.RaceFantastic {
                unit.AddExperience(combatScreen.Model.DefeatedDefenders * 2)
            }
        }
    } else if state == combat.CombatStateDefenderWin {
        for _, unit := range defenderStack.Units() {
            if unit.GetRace() != data.RaceFantastic {
                unit.AddExperience(combatScreen.Model.DefeatedAttackers * 2)
            }
        }
    }

    // ebiten.SetCursorMode(ebiten.CursorModeVisible)
    game.Drawer = oldDrawer

    for _, unit := range attackerStack.Units() {
        if unit.GetHealth() <= 0 {
            attacker.RemoveUnit(unit)
        }
    }

    for _, unit := range defenderStack.Units() {
        if unit.GetHealth() <= 0 {
            defender.RemoveUnit(unit)
        }
    }

    return state
}

func (game *Game) GetMainImage(index int) (*ebiten.Image, error) {
    image, err := game.ImageCache.GetImage("main.lbx", index, 0)

    if err != nil {
        log.Printf("Error: image in main.lbx is missing: %v", err)
    }

    return image, err
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
    return cache.GetImageTransform("mapback.lbx", 20, index, city.Banner.String(), units.MakeUpdateUnitColorsFunc(city.Banner))
}

func (game *Game) ShowGrandVizierUI(){
    yes := func(){
        // FIXME: enable grand vizier
    }

    no := func(){
        // FIXME: disable grand vizier
    }

    game.HudUI.AddElements(uilib.MakeConfirmDialogWithLayer(game.HudUI, game.Cache, &game.ImageCache, 1, "Do you wish to allow the Grand Vizier to select what buildings your cities create?", yes, no))
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
    game.HudUI.AddElements(spellbook.MakeSpellBookCastUI(game.HudUI, game.Cache, player.KnownSpells.OverlandSpells(), player.ComputeCastingSkill(), player.CastingSpell, player.CastingSpellProgress, true, func (spell spellbook.Spell, picked bool){
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

                created, cancel := artifact.ShowCreateArtifactScreen(yield, game.Cache, creation, &drawFunc)
                if cancel {
                    return
                }

                log.Printf("Create artifact %v", created)
                spell.OverrideCost = created.Cost()

                player.CreateArtifact = created
            }

            castingCost := spell.Cost(true)

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

func (game *Game) CatchmentArea(x int, y int) []image.Point {
    var out []image.Point

    for dx := -2; dx <= 2; dx++ {
        for dy := -2; dy <= 2; dy++ {
            // ignore corners
            if int(math.Abs(float64(dx)) + math.Abs(float64(dy))) == 4 {
                continue
            }

            out = append(out, image.Point{X: x + dx, Y: y + dy})
        }
    }

    return out
}

func (game *Game) ComputeMaximumPopulation(x int, y int, plane data.Plane) int {
    // find catchment area of x, y
    // for each square, compute food production
    // maximum pop is food production
    catchment := game.CatchmentArea(x, y)

    food := fraction.Zero()

    for _, point := range catchment {
        tile := game.GetMap(plane).GetTile(point.X, point.Y)
        food = food.Add(tile.Tile.FoodBonus())
        // FIXME: get bonus directly from tile.Extra
        bonus := game.GetMap(plane).GetBonusTile(point.X, point.Y)
        food = food.Add(fraction.FromInt(bonus.FoodBonus()))
    }

    maximum := int(food.ToFloat())
    if maximum > 25 {
        maximum = 25
    }

    return maximum
}

func (game *Game) CityGoldBonus(x int, y int, plane data.Plane) int {
    gold := 0
    tile := game.GetMap(plane).GetTile(x, y)
    if tile.Tile.TerrainType() == terrain.River {
        gold += 20
    }

    // check tiles immediately touching the city
    touchingShore := false
    for dx := -1; dx <= 1; dx++ {
        for dy := -1; dy <= 1; dy++ {
            if dx == 0 && dy == 0 {
                continue
            }

            tile := game.GetMap(plane).GetTile(x + dx, y + dy)
            if tile.Tile.TerrainType() == terrain.Shore {
                touchingShore = true
            }
        }
    }

    if touchingShore {
        gold += 10
    }

    return gold
}

func (game *Game) CityProductionBonus(x int, y int, plane data.Plane) int {
    catchment := game.CatchmentArea(x, y)

    production := 0

    for _, point := range catchment {
        tile := game.GetMap(plane).GetTile(point.X, point.Y)
        production += tile.Tile.ProductionBonus()
    }

    return production
}

func (game *Game) CreateOutpost(settlers units.StackUnit, player *playerlib.Player) *citylib.City {
    cityName := game.SuggestCityName(settlers.GetRace())

    newCity := citylib.MakeCity(cityName, settlers.GetX(), settlers.GetY(), settlers.GetRace(), settlers.GetBanner(), player.TaxRate, game.BuildingInfo, game.GetMap(settlers.GetPlane()))
    newCity.Plane = settlers.GetPlane()
    newCity.Population = 300
    newCity.Outpost = true
    newCity.Banner = player.Wizard.Banner
    newCity.ProducingBuilding = buildinglib.BuildingHousing
    newCity.ProducingUnit = units.UnitNone

    player.RemoveUnit(settlers)
    player.SelectedStack = nil
    game.RefreshUI()
    player.AddCity(newCity)

    stack := player.FindStack(newCity.X, newCity.Y)

    select {
        case game.Events<- &GameEventNewOutpost{City: newCity, Stack: stack}:
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
                // FIXME: check if this tile is valid to build an outpost on
                if settlers.HasAbility(data.AbilityCreateOutpost) {
                    game.CreateOutpost(settlers, player)
                    game.RefreshUI()
                    break
                }
            }
        } else if powers.Meld {
            node := game.GetMap(player.SelectedStack.Plane()).GetMagicNode(player.SelectedStack.X(), player.SelectedStack.Y())
            for _, melder := range player.SelectedStack.ActiveUnits() {
                if melder.HasAbility(data.AbilityMeld) {
                    game.DoMeld(melder, player, node)
                    game.RefreshUI()
                    break
                }
            }
        }
    }
}

func (game *Game) MakeHudUI() *uilib.UI {
    ui := &uilib.UI{
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
                        if game.Players[0].SelectedStack == nil {
                            select {
                                case game.Events <- &GameEventNextTurn{}:
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
    elements = append(elements, makeButton(1, 7, 4, false, func(){
        select {
            case game.Events <- &GameEventLoadMenu{}:
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
        switch game.Plane {
            case data.PlaneArcanus: game.Plane = data.PlaneMyrror
            case data.PlaneMyrror: game.Plane = data.PlaneArcanus
        }

        game.RefreshUI()
    }))

    if len(game.Players) > 0 && game.Players[0].SelectedStack != nil {
        player := game.Players[0]
        stack := player.SelectedStack

        unitX1 := 246
        unitY1 := 79

        unitX := unitX1
        unitY := unitY1

        row := 0
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
                LeftClick: func(this *uilib.UIElement){
                    stack.ToggleActive(unit)
                },
                RightClick: func(this *uilib.UIElement){
                    ui.AddElements(unitview.MakeUnitContextMenu(game.Cache, ui, unit, disband))
                },
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(unitRect.Min.X), float64(unitRect.Min.Y))
                    screen.DrawImage(unitBackground, &options)

                    options.GeoM.Translate(1, 1)

                    if stack.IsActive(unit){
                        unitBack, _ := units.GetUnitBackgroundImage(unit.GetBanner(), &game.ImageCache)
                        screen.DrawImage(unitBack, &options)
                    }

                    options.GeoM.Translate(1, 1)
                    unitImage, err := GetUnitImage(unit, &game.ImageCache, player.Wizard.Banner)
                    if err == nil {
                        screen.DrawImage(unitImage, &options)

                        // draw the first enchantment on the unit
                        for _, enchantment := range unit.GetEnchantments() {
                            x, y := options.GeoM.Apply(0, 0)
                            util.DrawOutline(screen, &game.ImageCache, unitImage, x, y, game.Counter/10, enchantment.Color())
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

                        x, y := options.GeoM.Apply(4, 19)
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
                    badgeOptions.GeoM.Translate(1, 21)
                    for i := 0; i < count; i++ {
                        pic, _ := game.ImageCache.GetImage("main.lbx", index, 0)
                        screen.DrawImage(pic, &badgeOptions)
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
                        screen.DrawImage(weapon, &weaponOptions)
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
                screen.DrawImage(doneImages[doneIndex], &options)
            },
            Inside: func(this *uilib.UIElement, x int, y int){
                doneCounter += 1
            },
            NotInside: func(this *uilib.UIElement){
                doneCounter = 0
            },
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
                screen.DrawImage(patrolImages[patrolIndex], &options)
            },
            Inside: func(this *uilib.UIElement, x int, y int){
                patrolCounter += 1
            },
            NotInside: func(this *uilib.UIElement){
                patrolCounter = 0
            },
            LeftClick: func(this *uilib.UIElement){
                patrolIndex = 1
            },
            LeftClickRelease: func(this *uilib.UIElement){
                patrolIndex = 0

                if player.SelectedStack != nil {
                    for _, unit := range player.SelectedStack.ActiveUnits() {
                        unit.SetPatrol(true)
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
                screen.DrawImage(waitImages[waitIndex], &options)
            },
            Inside: func(this *uilib.UIElement, x int, y int){
                waitCounter += 1
            },
            NotInside: func(this *uilib.UIElement){
                waitCounter = 0
            },
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
        buildIndex := 0
        buildRect := util.ImageRect(280, 186, buildImages[0])
        buildCounter := uint64(0)
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
                } else if powers.Meld {
                    use = meldImages[buildIndex]

                    canMeld := false
                    node := game.GetMap(player.SelectedStack.Plane()).GetMagicNode(player.SelectedStack.X(), player.SelectedStack.Y())
                    if node != nil && node.Empty {
                        canMeld = true
                    }

                    if !canMeld {
                        matrix.ChangeHSV(0, 0, 1)
                    }
                }

                colorm.DrawImage(screen, use, matrix, &options)
            },
            Inside: func(this *uilib.UIElement, x int, y int){
                buildCounter += 1
            },
            NotInside: func(this *uilib.UIElement){
                buildCounter = 0
            },
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
                    node := game.GetMap(player.SelectedStack.Plane()).GetMagicNode(player.SelectedStack.X(), player.SelectedStack.Y())
                    if node != nil && node.Empty {
                        canMeld = true
                    }

                    if canMeld {
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

    } else {
        // next turn
        nextTurnImage, _ := game.ImageCache.GetImage("main.lbx", 35, 0)
        nextTurnImageClicked, _ := game.ImageCache.GetImage("main.lbx", 58, 0)
        nextTurnRect := image.Rect(240, 174, 240 + nextTurnImage.Bounds().Dx(), 174 + nextTurnImage.Bounds().Dy())
        nextTurnClicked := false
        elements = append(elements, &uilib.UIElement{
            Rect: nextTurnRect,
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
                options.GeoM.Translate(240, 174)
                screen.DrawImage(nextTurnImage, &options)
                if nextTurnClicked {
                    options.GeoM.Translate(6, 5)
                    screen.DrawImage(nextTurnImageClicked, &options)
                }
            },
        })

        if len(game.Players) > 0 {
            player := game.Players[0]

            goldPerTurn := player.GoldPerTurn()
            foodPerTurn := player.FoodPerTurn()
            manaPerTurn := player.ManaPerTurn(game.ComputePower(player))

            elements = append(elements, &uilib.UIElement{
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    goldFood, _ := game.ImageCache.GetImage("main.lbx", 34, 0)
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(240, 77)
                    screen.DrawImage(goldFood, &options)

                    negativeScale := ebiten.ColorScale{}

                    // v is in range 0.5-1
                    v := (math.Cos(float64(game.Counter) / 7) + 1) / 4 + 0.5
                    negativeScale.SetR(float32(v))

                    if goldPerTurn < 0 {
                        game.InfoFontRed.PrintCenter(screen, 278, 103, 1, negativeScale, fmt.Sprintf("%v Gold", goldPerTurn))
                    } else {
                        game.InfoFontYellow.PrintCenter(screen, 278, 103, 1, ebiten.ColorScale{}, fmt.Sprintf("%v Gold", goldPerTurn))
                    }

                    if foodPerTurn < 0 {
                        game.InfoFontRed.PrintCenter(screen, 278, 135, 1, negativeScale, fmt.Sprintf("%v Food", foodPerTurn))
                    } else {
                        game.InfoFontYellow.PrintCenter(screen, 278, 135, 1, ebiten.ColorScale{}, fmt.Sprintf("%v Food", foodPerTurn))
                    }

                    if manaPerTurn < 0 {
                        game.InfoFontRed.PrintCenter(screen, 278, 167, 1, negativeScale, fmt.Sprintf("%v Mana", manaPerTurn))
                    } else {
                        game.InfoFontYellow.PrintCenter(screen, 278, 167, 1, ebiten.ColorScale{}, fmt.Sprintf("%v Mana", manaPerTurn))
                    }
                },
            })
        }
    }

    elements = append(elements, &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            game.WhiteFont.PrintRight(screen, 276, 68, 1, ebiten.ColorScale{}, fmt.Sprintf("%v GP", game.Players[0].Gold))
        },
    })

    elements = append(elements, &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            game.WhiteFont.PrintRight(screen, 313, 68, 1, ebiten.ColorScale{}, fmt.Sprintf("%v MP", game.Players[0].Mana))
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
        if stack.HasMoves() && len(stack.ActiveUnits()) > 0 {
            player.SelectedStack = stack
            stack.EnableMovers()
            game.Plane = stack.Plane()
            game.CenterCamera(stack.X(), stack.Y())
            break
        }
    }

    if player.Human {
        /*
        if player.SelectedStack == nil {
            fortressCity := player.FindFortressCity()
            if fortressCity != nil {
                game.CenterCamera(fortressCity.X, fortressCity.Y)
            }
        }
        */
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
    goldIssue := player.GoldPerTurn() < 0 && player.GoldPerTurn() > player.Gold
    foodIssue := player.FoodPerTurn() < 0

    manaPerTurn := player.ManaPerTurn(game.ComputePower(player))

    manaIssue := manaPerTurn < 0 && manaPerTurn > player.Mana

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

func (game *Game) StartPlayerTurn(player *playerlib.Player) {
    disbandedMessages := game.DisbandUnits(player)

    if player.Human && len(disbandedMessages) > 0 {
        select {
            case game.Events<- &GameEventScroll{Title: "", Text: strings.Join(disbandedMessages, "\n")}:
            default:
        }
    }

    power := game.ComputePower(player)

    player.Gold += player.GoldPerTurn()
    if player.Gold < 0 {
        player.Gold = 0
    }

    player.Mana += player.ManaPerTurn(power)
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

            if player.Human {
                select {
                    case game.Events<- &GameEventCastSpell{Player: player, Spell: player.CastingSpell}:
                    default:
                        log.Printf("Error: unable to invoke cast spell because event queue is full")
                }
            }
            player.CastingSpell = spellbook.Spell{}
            player.CastingSpellProgress = 0
        }
    }

    if player.ResearchingSpell.Valid() {
        player.ResearchProgress += int(player.SpellResearchPerTurn(power))
        if player.ResearchProgress >= player.ResearchingSpell.ResearchCost {

            if player.Human {
                select {
                    case game.Events<- &GameEventLearnedSpell{Player: player, Spell: player.ResearchingSpell}:
                    default:
                }
            }

            player.LearnSpell(player.ResearchingSpell)

            if player.Human {
                select {
                    case game.Events<- &GameEventResearchSpell{Player: player}:
                    default:
                }
            }
        }
    } else if game.TurnNumber > 1 {

        if player.Human {
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
        cityEvents := city.DoNextTurn(player.GetUnits(city.X, city.Y))
        for _, event := range cityEvents {
            switch event.(type) {
            case *citylib.CityEventPopulationGrowth:
                // growth := event.(*citylib.CityEventPopulationGrowth)

                scrollEvent := GameEventScroll{
                    Title: "CITY GROWTH",
                    // FIXME: 'has shrunk' if growth is negative?
                    Text: fmt.Sprintf("%v has grown to a population of %v.", city.Name, city.Citizens()),
                }

                if player.Human {
                    select {
                        case game.Events<- &scrollEvent:
                        default:
                    }
                }

                /*
                if growth.Size > 0 {
                    log.Printf("City grew by %v to %v", growth.Size, city.Citizens())
                } else {
                    log.Printf("City shrunk by %v to %v", -growth.Size, city.Citizens())
                }
                */
            case *citylib.CityEventNewBuilding:
                newBuilding := event.(*citylib.CityEventNewBuilding)

                if player.Human {
                    select {
                        case game.Events<- &GameEventNewBuilding{City: city, Building: newBuilding.Building, Player: player}:
                        default:
                    }
                }
            case *citylib.CityEventOutpostDestroyed:
                removeCities = append(removeCities, city)
                if player.Human {
                    select {
                        case game.Events<- &GameEventNotice{Message: fmt.Sprintf("The outpost of %v has been deserted.", city.Name)}:
                        default:
                    }
                }
            case *citylib.CityEventOutpostHamlet:
                if player.Human {
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
            }
        }
    }

    for _, stack := range player.Stacks {

        // every unit gains 1 experience at each turn
        for _, unit := range stack.Units() {
            if unit.GetRace() != data.RaceFantastic {
                unit.AddExperience(1)
            }
        }

        // base healing rate is 5%. in a town is 10%, with animists guild is 16.67%
        rate := 0.05

        city := player.FindCity(stack.X(), stack.Y())

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

    // game.CenterCamera(player.Cities[0].X, player.Cities[0].Y)
    game.DoNextUnit(player)
    game.RefreshUI()
}

func (game *Game) DoNextTurn(){
    game.CurrentPlayer += 1
    if game.CurrentPlayer >= len(game.Players) {
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

        if game.Players[game.CurrentPlayer].AIBehavior != nil {
            game.Players[game.CurrentPlayer].AIBehavior.NewTurn()
        }
    }

    game.TurnNumber += 1
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
    FogEdge_S_W := fogImage(3)
    FogEdge_S := fogImage(5)
    FogEdge_N_W := fogImage(7)
    FogEdge_N := fogImage(8)
    FogEdge_W := fogImage(9)

    /*
    FogEdge_SW := fogImage(6)
    FogEdge_SW_W_NW_N_NE := fogImage(7)
    FogEdge_NW_N_NE := fogImage(8)
    FogEdge_SW_N := fogImage(9)
    FogEdge_NW_W := fogImage(10)
    FogEdge_SW_W_NW_N := fogImage(11)
    */

    // fogBlack := game.GetFogImage()

    tileWidth := overworld.Map.TileWidth()
    tileHeight := overworld.Map.TileHeight()

    tilesPerRow := overworld.Map.TilesPerRow(screen.Bounds().Dx())
    tilesPerColumn := overworld.Map.TilesPerColumn(screen.Bounds().Dy())
    var options ebiten.DrawImageOptions

    fog := overworld.Fog

    /*
    fogNW := func(x int, y int) bool {
        if x == 0 || y == 0 {
            return false
        }

        return fog[x - 1][y - 1]
    }
    */

    checkFog := func(x int, y int) bool {
        x = overworld.Map.WrapX(x)
        if x < 0 || x >= len(fog) || y >= len(fog[x]) || y < 0{
            return false
        }

        return !fog[x][y]
    }

    fogN := func(x int, y int) bool {
        return checkFog(x, y - 1)
    }

    /*
    fogNE := func(x int, y int) bool {
        if x == len(fog) - 1 || y == 0 {
            return false
        }

        return fog[x + 1][y - 1]
    }
    */

    fogE := func(x int, y int) bool {
        return checkFog(x+1, y)
    }

    /*
    fogSE := func(x int, y int) bool {
        if x == len(fog) - 1 || y == len(fog[0]) - 1 {
            return false
        }

        return fog[x + 1][y + 1]
    }
    */

    fogS := func(x int, y int) bool {
        return checkFog(x, y + 1)
    }

    /*
    fogSW := func(x int, y int) bool {
        if x == 0 || y == len(fog[0]) - 1 {
            return false
        }

        return fog[x - 1][y + 1]
    }
    */

    fogW := func(x int, y int) bool {
        return checkFog(x - 1, y)
    }

    for x := 0; x < tilesPerRow; x++ {
        for y := 0; y < tilesPerColumn; y++ {

            tileX := overworld.Map.WrapX(x + overworld.CameraX)
            tileY := y + overworld.CameraY

            options.GeoM = geom
            options.GeoM.Translate(float64(x * tileWidth), float64(y * tileHeight))

            if tileX >= 0 && tileY >= 0 && tileX < len(fog) && tileY < len(fog[tileX]) && fog[tileX][tileY] {
                // nw := fogNW(tileX, tileY)
                n := fogN(tileX, tileY)
                // ne := fogNE(tileX, tileY)
                e := fogE(tileX, tileY)
                // se := fogSE(tileX, tileY)
                s := fogS(tileX, tileY)
                // sw := fogSW(tileX, tileY)
                w := fogW(tileX, tileY)

                if n && e {
                    screen.DrawImage(FogEdge_N_E, &options)
                } else if n {
                    screen.DrawImage(FogEdge_N, &options)
                } else if e {
                    options.GeoM.Scale(1, -1)
                    screen.DrawImage(FogEdge_W, &options)
                }

                if s && e {
                    screen.DrawImage(FogEdge_S_E, &options)
                } else if s {
                    screen.DrawImage(FogEdge_S, &options)
                }

                if n && w {
                    screen.DrawImage(FogEdge_N_W, &options)
                } else if w && !s {
                    screen.DrawImage(FogEdge_W, &options)
                }

                if s && w {
                    screen.DrawImage(FogEdge_S_W, &options)
                }

                /*
                if nw && n && ne && e && se && !s && !sw && !w {
                    screen.DrawImage(FogEdge_NW_N_NE_E_SE, &options)
                } else if sw && s && se && e && ne && !n && !nw && !w {
                    screen.DrawImage(FogEdge_SW_S_SE_E_NE, &options)
                } else if nw && w && sw && s && se && !e && !ne && !n {
                    screen.DrawImage(FogEdge_NW_W_SW_S_SE, &options)
                } else if nw && s && !n && !ne && !e && !se && !sw && !w {
                    screen.DrawImage(FogEdge_NW_S, &options)
                } else if sw && s && se && !n && !ne && !e && !nw && !w {
                    screen.DrawImage(FogEdge_SW_S_SE, &options)
                } else if sw && !n && !ne && !e && !se && !s && !nw && !w {
                    screen.DrawImage(FogEdge_SW, &options)
                } else if nw && w && sw && !s && n && ne && !e && !se {
                    screen.DrawImage(FogEdge_SW_W_NW_N_NE, &options)
                } else if nw && n && ne && !s && !se && !e && !sw && !w {
                    screen.DrawImage(FogEdge_NW_N_NE, &options)
                } else if sw && n && !ne && !e && !se && !s && !nw && !w {
                    screen.DrawImage(FogEdge_SW_N, &options)
                } else if nw && w && !n && !ne && !e && !se && !s && !sw {
                    screen.DrawImage(FogEdge_NW_W, &options)
                } else if sw && w && nw && n && !ne && !e && !se && !s {
                    screen.DrawImage(FogEdge_SW_W_NW_N, &options)
                }
                */
            } else {

                if overworld.FogBlack != nil {
                    screen.DrawImage(overworld.FogBlack, &options)
                }
            }
        }
    }

}

type Overworld struct {
    CameraX int
    CameraY int
    Counter uint64
    Map *maplib.Map
    Cities []*citylib.City
    CitiesMiniMap []maplib.MiniMapCity
    Stacks []*playerlib.UnitStack
    SelectedStack *playerlib.UnitStack
    MovingStack *playerlib.UnitStack
    ImageCache *util.ImageCache
    Fog [][]bool
    ShowAnimation bool
    FogBlack *ebiten.Image
}

func (overworld *Overworld) DrawMinimap(screen *ebiten.Image){
    overworld.Map.DrawMinimap(screen, overworld.CitiesMiniMap, overworld.CameraX + 5, overworld.CameraY + 5, overworld.Fog, overworld.Counter, true)
}

func (overworld *Overworld) DrawOverworld(screen *ebiten.Image, geom ebiten.GeoM){
    overworld.Map.DrawLayer1(overworld.CameraX, overworld.CameraY, overworld.Counter / 8, overworld.ImageCache, screen, geom)

    tileWidth := overworld.Map.TileWidth()
    tileHeight := overworld.Map.TileHeight()

    convertTileCoordinates := func(x int, y int) (int, int) {
        outX := overworld.Map.WrapX(x - overworld.CameraX) * tileWidth
        outY := (y - overworld.CameraY) * tileHeight
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
            x, y := convertTileCoordinates(city.X, city.Y)
            options.GeoM = geom
            // draw the city in the center of the tile
            // first compute center of tile
            options.GeoM.Translate(float64(x + tileWidth / 2.0), float64(y + tileHeight / 2.0))
            // then move the city image so that the center of the image is at the center of the tile
            options.GeoM.Translate(float64(-cityPic.Bounds().Dx()) / 2.0, float64(-cityPic.Bounds().Dy()) / 2.0)
            screen.DrawImage(cityPic, &options)

            /*
            tx, ty := options.GeoM.Apply(float64(0), float64(0))
            vector.StrokeRect(screen, float32(tx), float32(ty), float32(cityPic.Bounds().Dx()), float32(cityPic.Bounds().Dy()), 1, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, true)
            vector.DrawFilledCircle(screen, float32(x) + float32(tileWidth) / 2, float32(y) + float32(tileHeight) / 2, 2, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, true)
            */
        }
    }

    for _, stack := range overworld.Stacks {
        doDraw := false
        if stack.Leader() == nil {
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
            options.GeoM = geom
            x, y := convertTileCoordinates(stack.X(), stack.Y())
            options.GeoM.Translate(float64(x) + stack.OffsetX() * float64(tileWidth), float64(y) + stack.OffsetY() * float64(tileHeight))

            leader := stack.Leader()

            unitBack, err := units.GetUnitBackgroundImage(leader.GetBanner(), overworld.ImageCache)
            if err == nil {
                screen.DrawImage(unitBack, &options)
            }

            pic, err := GetUnitImage(leader, overworld.ImageCache, leader.GetBanner())
            if err == nil {
                options.GeoM.Translate(1, 1)
                screen.DrawImage(pic, &options)

                enchantment := util.First(leader.GetEnchantments(), data.UnitEnchantmentNone)
                if enchantment != data.UnitEnchantmentNone {
                    x, y := options.GeoM.Apply(0, 0)
                    util.DrawOutline(screen, overworld.ImageCache, pic, x, y, overworld.Counter/10, enchantment.Color())
                }
            }
        }
    }

    overworld.Map.DrawLayer2(overworld.CameraX, overworld.CameraY, overworld.Counter / 8, overworld.ImageCache, screen, geom)

    if overworld.Fog != nil {
        overworld.DrawFog(screen, geom)
    }
}

func (game *Game) Draw(screen *ebiten.Image){
    game.Drawer(screen, game)
}

func (game *Game) DrawGame(screen *ebiten.Image){

    var cities []*citylib.City
    var citiesMiniMap []maplib.MiniMapCity
    var stacks []*playerlib.UnitStack
    var selectedStack *playerlib.UnitStack
    var fog [][]bool

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

    overworld := Overworld{
        CameraX: game.cameraX,
        CameraY: game.cameraY,
        Counter: game.Counter,
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

    overworld.DrawOverworld(screen, ebiten.GeoM{})

    var miniGeom ebiten.GeoM
    miniGeom.Translate(250, 20)
    mx, my := miniGeom.Apply(0, 0)
    miniWidth := 60
    miniHeight := 31
    mini := screen.SubImage(image.Rect(int(mx), int(my), int(mx) + miniWidth, int(my) + miniHeight)).(*ebiten.Image)
    overworld.DrawMinimap(mini)

    game.HudUI.Draw(game.HudUI, screen)
}
