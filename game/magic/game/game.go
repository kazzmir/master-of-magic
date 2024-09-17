package game

import (
    "image/color"
    "image"
    "math/rand"
    "log"
    "math"
    "fmt"
    "strings"

    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/game/magic/unitview"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/cityview"
    "github.com/kazzmir/master-of-magic/game/magic/magicview"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/draw"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    _ "github.com/hajimehoshi/ebiten/v2/vector"
)

func (game *Game) GetFogImage() *ebiten.Image {
    if game.Fog != nil {
        return game.Fog
    }

    game.Fog = ebiten.NewImage(game.Map.TileWidth(), game.Map.TileHeight())
    game.Fog.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 0xff})
    return game.Fog
}

type GameEvent interface {
}

type GameEventMagicView struct {
}

type GameEventNewOutpost struct {
    City *citylib.City
    Stack *playerlib.UnitStack
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

    InfoFontYellow *font.Font
    Counter uint64
    Fog *ebiten.Image
    Drawer func (*ebiten.Image, *Game)
    State GameState
    Plane data.Plane

    Events chan GameEvent

    BookOrder []int

    cameraX int
    cameraY int

    HudUI *uilib.UI
    Help lbx.Help

    // FIXME: need one map for arcanus and one for myrran
    Map *Map

    Players []*playerlib.Player
}

// create an array of N integers where each integer is some value between 0 and 2
// these values correlate to the index of the book image to draw under the wizard portrait
func randomizeBookOrder(books int) []int {
    order := make([]int, books)
    for i := 0; i < books; i++ {
        order[i] = rand.Intn(3)
    }
    return order
}

func (game *Game) MakeFog() [][]bool {
    fog := make([][]bool, game.Map.Width())
    for x := 0; x < game.Map.Width(); x++ {
        fog[x] = make([]bool, game.Map.Height())
    }

    return fog
}

func (game *Game) CenterCamera(x int, y int){
    game.cameraX = x - 5
    game.cameraY = y - 5

    if game.cameraX < 0 {
        game.cameraX = 0
    }

    if game.cameraY < 0 {
        game.cameraY = 0
    }
}

func (game *Game) AddPlayer(wizard setup.WizardCustom) *playerlib.Player{
    newPlayer := &playerlib.Player{
        ArcanusFog: game.MakeFog(),
        MyrrorFog: game.MakeFog(),
        Wizard: wizard,
    }

    game.Players = append(game.Players, newPlayer)
    return newPlayer
}

func MakeGame(lbxCache *lbx.LbxCache) *Game {

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
        orange,
        orange,
        orange,
        orange,
        orange,
        orange,
    }

    infoFontYellow := font.MakeOptimizedFontWithPalette(fonts[0], yellowPalette)

    whitePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.White, color.White, color.White, color.White,
    }

    whiteFont := font.MakeOptimizedFontWithPalette(fonts[0], whitePalette)

    game := &Game{
        Cache: lbxCache,
        Help: help,
        Events: make(chan GameEvent, 1),
        Map: MakeMap(terrainData),
        State: GameStateRunning,
        BookOrder: randomizeBookOrder(12),
        ImageCache: util.MakeImageCache(lbxCache),
        InfoFontYellow: infoFontYellow,
        WhiteFont: whiteFont,
    }

    game.HudUI = game.MakeHudUI()
    game.Drawer = func(screen *ebiten.Image, game *Game){
        game.DrawGame(screen)
    }

    return game
}

func (game *Game) FindValidCityLocation() (int, int) {
    continents := game.Map.Map.FindContinents()

    for i := 0; i < 10; i++ {
        continentIndex := rand.Intn(len(continents))
        continent := continents[continentIndex]
        if len(continent) > 100 {
            index := rand.Intn(len(continent))
            x := continent[index].X
            y := continent[index].Y

            if game.Map.Map.Terrain[x][y] == terrain.TileLand.Index {
                return x, y
            }
        }
    }

    return 0, 0
}

func (game *Game) doMagicView(yield coroutine.YieldFunc) {

    oldDrawer := game.Drawer
    magicScreen := magicview.MakeMagicScreen(game.Cache, game.Players[0])

    game.Drawer = func (screen *ebiten.Image, game *Game){
        magicScreen.Draw(screen)
    }

    for magicScreen.Update() == magicview.MagicScreenStateRunning {
        yield()
    }

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

    game.Drawer = func (screen *ebiten.Image, game *Game){
        oldDrawer(screen, game)

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

        width := float64(4)
        height := float64(8)

        yOffset := float64((game.Counter/3) % 16) - 8

        vertices := [4]ebiten.Vertex{
            ebiten.Vertex{
                DstX: float32(cursorX),
                DstY: float32(y - yOffset),
                SrcX: 0,
                SrcY: 0,
                ColorA: 1,
                ColorB: 1 ,
                ColorG: 1,
                ColorR: 1,
            },
            ebiten.Vertex{
                DstX: float32(cursorX + width),
                DstY: float32(y - yOffset),
                SrcX: 0,
                SrcY: 0,
                ColorA: 1,
                ColorB: 1 ,
                ColorG: 1,
                ColorR: 1,
            },
            ebiten.Vertex{
                DstX: float32(cursorX + width),
                DstY: float32(y + height - yOffset),
                SrcX: 0,
                SrcY: 0,
                ColorA: 0.1,
                ColorB: 1 ,
                ColorG: 1,
                ColorR: 1,
            },
            ebiten.Vertex{
                DstX: float32(cursorX),
                DstY: float32(y + height - yOffset),
                SrcX: 0,
                SrcY: 0,
                ColorA: 0.1,
                ColorB: 1 ,
                ColorG: 1,
                ColorR: 1,
            },
        }

        cursorArea := screen.SubImage(image.Rect(int(cursorX), int(y), int(cursorX + width), int(y + height))).(*ebiten.Image)
        cursorArea.DrawTriangles(vertices[:], []uint16{0, 1, 2, 2, 3, 0}, source, nil)
    }

    repeats := make(map[ebiten.Key]uint64)

    doInput := func(){
        keys := make([]ebiten.Key, 0)
        keys = inpututil.AppendJustReleasedKeys(keys)
        for _, key := range keys {
            delete(repeats, key)
        }

        runes := make([]rune, 0)
        runes = ebiten.AppendInputChars(runes)
        for _, r := range runes {
            if validNameString(string(r)) && nameFont.MeasureTextWidth(name + string(r), 1) < maxLength {
                name += string(r)
            }
        }

        keys = make([]ebiten.Key, 0)
        keys = inpututil.AppendPressedKeys(keys)
        for _, key := range keys {
            repeat, ok := repeats[key]
            if !ok {
                repeats[key] = game.Counter
                repeat = game.Counter
            }

            diff := game.Counter - repeat
            // log.Printf("repeat %v diff=%v", key, diff)
            if diff == 0 {
                // use = true
            } else if diff < 15 {
                continue
            } else if diff % 5 == 0 {
                // use = true
            } else {
                continue
            }

            switch key {
                case ebiten.KeyEnter:
                    if len(name) > 0 {
                        quit = true
                    }
                case ebiten.KeySpace:
                    if nameFont.MeasureTextWidth(name + " ", 1) < maxLength {
                        name += " "
                    }
                case ebiten.KeyBackspace:
                    if len(name) > 0 {
                        name = name[:len(name) - 1]
                    }
            }
        }
    }

    for !quit {
        game.Counter += 1
        doInput()
        yield()
    }

    return name
}

func (game *Game) showOutpost(yield coroutine.YieldFunc, city *citylib.City, stack *playerlib.UnitStack){
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
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        yellow, yellow, yellow,
        yellow, yellow, yellow,
        yellow, yellow, yellow,
        yellow, yellow, yellow,
    }

    bigFont := font.MakeOptimizedFontWithPalette(fonts[5], yellowPalette)

    game.Drawer = func (screen *ebiten.Image, game *Game){
        drawer(screen, game)

        background, _ := game.ImageCache.GetImage("backgrnd.lbx", 32, 0)

        var options ebiten.DrawImageOptions
        options.GeoM.Translate(30, 50)
        screen.DrawImage(background, &options)

        numHouses := 3
        maxHouses := 10

        houseOptions := options
        houseOptions.GeoM.Translate(7, 31)

        house, _ := game.ImageCache.GetImage("backgrnd.lbx", 34, 0)

        for i := 0; i < numHouses; i++ {
            screen.DrawImage(house, &houseOptions)
            houseOptions.GeoM.Translate(float64(house.Bounds().Dx()) + 1, 0)
        }

        emptyHouse, _ := game.ImageCache.GetImage("backgrnd.lbx", 37, 0)
        for i := numHouses; i < maxHouses; i++ {
            screen.DrawImage(emptyHouse, &houseOptions)
            houseOptions.GeoM.Translate(float64(emptyHouse.Bounds().Dx()) + 1, 0)
        }

        if stack != nil {
            stackOptions := options
            stackOptions.GeoM.Translate(7, 55)

            for _, unit := range stack.Units() {
                pic, _ := GetUnitImage(unit.Unit, &game.ImageCache)
                screen.DrawImage(pic, &stackOptions)
                stackOptions.GeoM.Translate(float64(pic.Bounds().Dx()) + 1, 0)
            }
        }

        x, y := options.GeoM.Apply(6, 22)
        game.InfoFontYellow.Print(screen, x, y, 1, options.ColorScale, city.Race.String())

        x, y = options.GeoM.Apply(20, 5)
        bigFont.Print(screen, x, y, 1, options.ColorScale, "New Outpost Founded")

        cityScapeOptions := options
        cityScapeOptions.GeoM.Translate(185, 30)
        x, y = cityScapeOptions.GeoM.Apply(0, 0)
        cityScape := screen.SubImage(image.Rect(int(x), int(y), int(x + 72), int(y + 66))).(*ebiten.Image)

        cityScapeBackground, _ := game.ImageCache.GetImage("cityscap.lbx", 0, 0)
        cityScape.DrawImage(cityScapeBackground, &cityScapeOptions)

        cityHouse, _ := game.ImageCache.GetImage("cityscap.lbx", 25, 0)
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
        if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
            quit = true
        }

        yield()
    }

    city.Name = game.doInput(yield, "New Outpost", city.Name, 80, 100)
}

func (game *Game) showMovement(yield coroutine.YieldFunc, oldX int, oldY int, stack *playerlib.UnitStack){
    // the number of frames it takes to move a unit one tile
    frames := 10

    dx := float64(oldX - stack.X())
    dy := float64(oldY - stack.Y())

    game.State = GameStateUnitMoving

    for i := 0; i < frames; i++ {
        game.Counter += 1
        stack.SetOffset(dx * float64(frames - i) / float64(frames), dy * float64(frames - i) / float64(frames))
        yield()
    }

    game.State = GameStateRunning

    stack.SetOffset(0, 0)
    game.CenterCamera(stack.X(), stack.Y())
}

func (game *Game) Update(yield coroutine.YieldFunc) GameState {
    game.Counter += 1

    /*
    if game.Counter % 10 == 0 {
        log.Printf("TPS: %v FPS: %v", ebiten.ActualTPS(), ebiten.ActualFPS())
    }
    */

    /*
    tilesPerRow := game.Map.TilesPerRow(data.ScreenWidth)
    tilesPerColumn := game.Map.TilesPerColumn(data.ScreenHeight)
    */

    select {
        case event := <-game.Events:
            switch event.(type) {
                case *GameEventMagicView:
                    game.doMagicView(yield)
                case *GameEventNewOutpost:
                    outpost := event.(*GameEventNewOutpost)
                    game.showOutpost(yield, outpost.City, outpost.Stack)
                case *GameEventCityName:
                    cityEvent := event.(*GameEventCityName)
                    city := cityEvent.City
                    city.Name = game.doInput(yield, cityEvent.Title, city.Name, cityEvent.X, cityEvent.Y)
            }
        default:
    }

    switch game.State {
        case GameStateRunning:
            game.HudUI.StandardUpdate()

            // kind of a hack to not allow player to interact with anything other than the current ui modal
            if game.HudUI.GetHighestLayerValue() == 0 {

                // log.Printf("Game.Update")
                keys := make([]ebiten.Key, 0)
                keys = inpututil.AppendJustPressedKeys(keys)

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

                if len(game.Players) > 0 {
                    player := game.Players[0]
                    if player.SelectedStack != nil {
                        stack := player.SelectedStack

                        // game.CenterCamera(stack.X(), stack.Y())

                        if dx != 0 || dy != 0 {
                            activeUnits := stack.ActiveUnits()
                            if len(activeUnits) > 0 {
                                inactiveUnits := stack.InactiveUnits()
                                if len(inactiveUnits) > 0 {
                                    stack.RemoveUnits(inactiveUnits)
                                    player.AddStack(playerlib.MakeUnitStackFromUnits(inactiveUnits))
                                    game.HudUI = game.MakeHudUI()
                                }

                                oldX := stack.X()
                                oldY := stack.Y()

                                newX := stack.X() + dx
                                newY := stack.Y() + dy

                                mergeStack := player.FindStack(newX, newY)

                                stack.Move(dx, dy)

                                game.showMovement(yield, oldX, oldY, stack)

                                player.LiftFog(stack.X(), stack.Y(), 2)

                                // game.State = GameStateUnitMoving

                                if mergeStack != nil {
                                    stack = player.MergeStacks(mergeStack, stack)
                                    player.SelectedStack = stack
                                    game.HudUI = game.MakeHudUI()
                                }

                                for _, otherPlayer := range game.Players[1:] {
                                    otherStack := otherPlayer.FindStack(stack.X(), stack.Y())
                                    if otherStack != nil {
                                        game.doCombat(yield, player, stack, otherPlayer, otherStack)
                                    }
                                }
                            }
                        }
                    }

                    rightClick := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)
                    if rightClick {
                        mouseX, mouseY := ebiten.CursorPosition()

                        // can only click into the area not hidden by the hud
                        if mouseX < 240 && mouseY > 18 {
                            // log.Printf("Click at %v, %v", mouseX, mouseY)
                            tileX := game.cameraX + mouseX / game.Map.TileWidth()
                            tileY := game.cameraY + mouseY / game.Map.TileHeight()

                            game.CenterCamera(tileX, tileY)

                            for _, city := range player.Cities {
                                if city.X == tileX && city.Y == tileY {
                                    game.doCityScreen(yield, city, player)
                                }
                            }
                        }
                    }
                }
            }
    }

    return game.State
}

func (game *Game) doCityScreen(yield coroutine.YieldFunc, city *citylib.City, player *playerlib.Player){
    cityScreen := cityview.MakeCityScreen(game.Cache, city, game.Players[0])

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

    overworld := Overworld{
        CameraX: city.X - 2,
        CameraY: city.Y - 2,
        Counter: 0,
        Map: game.Map,
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
        })
    }

    for cityScreen.Update() == cityview.CityScreenStateRunning {
        overworld.Counter += 1
        yield()
    }

    game.Drawer = oldDrawer
}

func (game *Game) doCombat(yield coroutine.YieldFunc, attacker *playerlib.Player, attackerStack *playerlib.UnitStack, defender *playerlib.Player, defenderStack *playerlib.UnitStack){
    attackingArmy := combat.Army{
        Player: attacker,
    }

    for _, unit := range attackerStack.Units() {
        attackingArmy.AddUnit(unit.Unit)
    }

    defendingArmy := combat.Army{
        Player: defender,
    }

    for _, unit := range defenderStack.Units() {
        defendingArmy.AddUnit(unit.Unit)
    }

    attackingArmy.LayoutUnits(combat.TeamAttacker)
    defendingArmy.LayoutUnits(combat.TeamDefender)

    combatScreen := combat.MakeCombatScreen(game.Cache, &defendingArmy, &attackingArmy, attacker)
    oldDrawer := game.Drawer

    ebiten.SetCursorMode(ebiten.CursorModeHidden)

    game.Drawer = func (screen *ebiten.Image, game *Game){
        combatScreen.Draw(screen)
    }

    state := combat.CombatStateRunning
    for state == combat.CombatStateRunning {
        state = combatScreen.Update()
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

    ebiten.SetCursorMode(ebiten.CursorModeVisible)
    game.Drawer = oldDrawer
}

func (game *Game) GetMainImage(index int) (*ebiten.Image, error) {
    image, err := game.ImageCache.GetImage("main.lbx", index, 0)

    if err != nil {
        log.Printf("Error: image in main.lbx is missing: %v", err)
    }

    return image, err
}

func GetUnitBackgroundImage(banner data.BannerType, imageCache *util.ImageCache) (*ebiten.Image, error) {
    index := -1
    switch banner {
        case data.BannerBlue: index = 14
        case data.BannerGreen: index = 15
        case data.BannerPurple: index = 16
        case data.BannerRed: index = 17
        case data.BannerYellow: index = 18
        case data.BannerBrown: index = 19
    }

    image, err := imageCache.GetImage("mapback.lbx", index, 0)
    if err != nil {
        log.Printf("Error: image in mapback.lbx is missing: %v", err)
    }

    return image, err
}

func GetUnitImage(unit units.Unit, imageCache *util.ImageCache) (*ebiten.Image, error) {
    image, err := imageCache.GetImage(unit.LbxFile, unit.Index, 0)

    if err != nil {
        log.Printf("Error: image in %v is missing: %v", unit.LbxFile, err)
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

    // the city image is a sub-frame of animation 20
    return cache.GetImage("mapback.lbx", 20, index)
}

func GetCityWallImage(size citylib.CitySize, cache *util.ImageCache) (*ebiten.Image, error) {
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

func (game *Game) MakeUnitContextMenu(ui *uilib.UI, unit *playerlib.Unit) []*uilib.UIElement {
    fontLbx, err := game.Cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
    }

    descriptionPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 90}),
        util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}),
        util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 200}),
        util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 200}),
        util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 200}),
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
    }

    descriptionFont := font.MakeOptimizedFontWithPalette(fonts[4], descriptionPalette)
    smallFont := font.MakeOptimizedFontWithPalette(fonts[1], descriptionPalette)
    mediumFont := font.MakeOptimizedFontWithPalette(fonts[2], descriptionPalette)

    yellowGradient := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0},
        color.RGBA{R: 0xed, G: 0xa4, B: 0x00, A: 0xff},
        color.RGBA{R: 0xff, G: 0xbc, B: 0x00, A: 0xff},
        color.RGBA{R: 0xff, G: 0xd6, B: 0x11, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
    }

    okDismissFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowGradient)

    var elements []*uilib.UIElement

    const fadeSpeed = 7

    getAlpha := ui.MakeFadeIn(fadeSpeed)

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := game.ImageCache.GetImage("unitview.lbx", 1, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(31, 6)
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(background, &options)

            options.GeoM.Translate(25, 30)
            unitview.RenderCombatImage(screen, &game.ImageCache, &unit.Unit, options)

            options.GeoM.Reset()
            options.GeoM.Translate(31, 6)
            options.GeoM.Translate(10, 50)
            unitview.RenderUnitInfoStats(screen, &game.ImageCache, &unit.Unit, descriptionFont, smallFont, options)

            options.GeoM.Translate(0, 60)
            unitview.RenderUnitAbilities(screen, &game.ImageCache, &unit.Unit, mediumFont, options)
        },
    })

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            box, _ := game.ImageCache.GetImage("unitview.lbx", 2, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(248, 139)
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(box, &options)
        },
    })

    buttonBackgrounds, _ := game.ImageCache.GetImages("backgrnd.lbx", 24)
    // dismiss button
    cancelRect := util.ImageRect(257, 149, buttonBackgrounds[0])
    cancelIndex := 0
    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Rect: cancelRect,
        LeftClick: func(this *uilib.UIElement){
            cancelIndex = 1

            var confirmElements []*uilib.UIElement

            yes := func(){
                ui.RemoveElements(elements)
                // FIXME: disband unit
            }

            no := func(){
            }

            confirmElements = uilib.MakeConfirmDialogWithLayer(ui, game.Cache, &game.ImageCache, 2, fmt.Sprintf("Do you wish to disband the unit of %v?", unit.Unit.Name), yes, no)

            ui.AddElements(confirmElements)
        },
        LeftClickRelease: func(this *uilib.UIElement){
            cancelIndex = 0
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(cancelRect.Min.X), float64(cancelRect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(buttonBackgrounds[cancelIndex], &options)

            x := float64(cancelRect.Min.X + cancelRect.Max.X) / 2
            y := float64(cancelRect.Min.Y + cancelRect.Max.Y) / 2
            okDismissFont.PrintCenter(screen, x, y - 5, 1, options.ColorScale, "Dismiss")
        },
    })

    okRect := util.ImageRect(257, 169, buttonBackgrounds[0])
    okIndex := 0
    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        Rect: okRect,
        LeftClick: func(this *uilib.UIElement){
            okIndex = 1
        },
        LeftClickRelease: func(this *uilib.UIElement){
            getAlpha = ui.MakeFadeOut(fadeSpeed)

            ui.AddDelay(fadeSpeed, func(){
                ui.RemoveElements(elements)
            })
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(okRect.Min.X), float64(okRect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(buttonBackgrounds[okIndex], &options)

            x := float64(okRect.Min.X + okRect.Max.X) / 2
            y := float64(okRect.Min.Y + okRect.Max.Y) / 2
            okDismissFont.PrintCenter(screen, x, y - 5, 1, options.ColorScale, "Ok")
        },
    })

    return elements
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

func (game *Game) ShowMirrorUI(){
    cornerX := 50
    cornerY := 1

    fontLbx, err := game.Cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Could not read fonts: %v", err)
        return
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Could not read fonts: %v", err)
        return
    }

    yellow := color.RGBA{R: 0xea, G: 0xb6, B: 0x00, A: 0xff}
    yellowPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        yellow, yellow, yellow,
        yellow, yellow, yellow,
    }

    smallFont := font.MakeOptimizedFontWithPalette(fonts[0], yellowPalette)

    heroFont := font.MakeOptimizedFontWithPalette(fonts[2], yellowPalette)

    var element *uilib.UIElement

    getAlpha := game.HudUI.MakeFadeIn(7)

    var portrait *ebiten.Image

    if len(game.Players) == 0 {
        return
    }

    player := game.Players[0]

    imageCache := util.MakeImageCache(game.Cache)

    bannerIndex := 0
    switch player.Wizard.Banner {
        case data.BannerBlue: bannerIndex = 0
        case data.BannerGreen: bannerIndex = 1
        case data.BannerPurple: bannerIndex = 2
        case data.BannerRed: bannerIndex = 3
        case data.BannerYellow: bannerIndex = 4
    }

    wizardIndex := 0

    switch player.Wizard.Base {
        case data.WizardMerlin: wizardIndex = 0
        case data.WizardRaven: wizardIndex = 5
        case data.WizardSharee: wizardIndex = 10
        case data.WizardLoPan: wizardIndex = 15
        case data.WizardJafar: wizardIndex = 20
        case data.WizardOberic: wizardIndex = 25
        case data.WizardRjak: wizardIndex = 30
        case data.WizardSssra: wizardIndex = 35
        case data.WizardTauron: wizardIndex = 40
        case data.WizardFreya: wizardIndex = 45
        case data.WizardHorus: wizardIndex = 50
        case data.WizardAriel: wizardIndex = 55
        case data.WizardTlaloc: wizardIndex = 60
        case data.WizardKali: wizardIndex = 65
    }

    portrait, _ = imageCache.GetImage("lilwiz.lbx", wizardIndex + bannerIndex, 0)

    doClose := func(){
        getAlpha = game.HudUI.MakeFadeOut(7)
        game.HudUI.AddDelay(7, func(){
            game.HudUI.RemoveElement(element)
        })
    }

    element = &uilib.UIElement{
        Layer: 1,
        LeftClick: func(this *uilib.UIElement){
            doClose()
        },
        NotLeftClicked: func(this *uilib.UIElement){
            doClose()
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("backgrnd.lbx", 4, 0)

            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(cornerX), float64(cornerY))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(background, &options)

            if portrait != nil {
                options.GeoM.Translate(11, 11)
                screen.DrawImage(portrait, &options)
            }

            smallFont.PrintCenter(screen, float64(cornerX + 30), float64(cornerY + 75), 1, options.ColorScale, fmt.Sprintf("%v GP", player.Gold))
            smallFont.PrintRight(screen, float64(cornerX + 170), float64(cornerY + 75), 1, options.ColorScale, fmt.Sprintf("%v MP", player.Mana))

            options.GeoM.Translate(34, 55)
            draw.DrawBooks(screen, options, &imageCache, player.Wizard.Books, game.BookOrder)

            smallFont.Print(screen, float64(cornerX + 13), float64(cornerY + 112), 1, options.ColorScale, setup.JoinAbilities(player.Wizard.Abilities))

            heroFont.PrintCenter(screen, float64(cornerX + 90), float64(cornerY + 131), 1, options.ColorScale, "Heroes")
            // FIXME: draw hero portraits here
        },
    }

    game.HudUI.AddElement(element)
}

func (game *Game) ShowTaxCollectorUI(cornerX int, cornerY int){
    taxes := []uilib.Selection{
        uilib.Selection{
            Name: "0 gold, 0% unrest",
            Action: func(){},
        },
        uilib.Selection{
            Name: "0.5 gold, 10% unrest",
            Action: func(){},
        },
        uilib.Selection{
            Name: "1 gold, 20% unrest",
            Action: func(){},
        },
        uilib.Selection{
            Name: "1.5 gold, 30% unrest",
            Action: func(){},
        },
        uilib.Selection{
            Name: "2 gold, 45% unrest",
            Action: func(){},
        },
        uilib.Selection{
            Name: "2.5 gold, 60% unrest",
            Action: func(){},
        },
        uilib.Selection{
            Name: "3 gold, 75% unrest",
            Action: func(){},
        },
    }

    game.HudUI.AddElements(uilib.MakeSelectionUI(game.HudUI, game.Cache, &game.ImageCache, cornerX, cornerY, "Tax Per Population", taxes))
}

func (game *Game) ShowApprenticeUI(){
    game.HudUI.AddElements(spellbook.MakeSpellBookUI(game.HudUI, game.Cache))
}

// advisor ui
func (game *Game) MakeInfoUI(cornerX int, cornerY int) []*uilib.UIElement {
    advisors := []uilib.Selection{
        uilib.Selection{
            Name: "Surveyor",
            Action: func(){},
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
                game.ShowApprenticeUI()
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
                game.ShowMirrorUI()
            },
            Hotkey: "(F9)",
        },
    }

    return uilib.MakeSelectionUI(game.HudUI, game.Cache, &game.ImageCache, cornerX, cornerY, "Select An Advisor", advisors)
}

func (game *Game) ShowSpellBookCastUI(){
    game.HudUI.AddElements(spellbook.MakeSpellBookCastUI(game.HudUI, game.Cache, spellbook.Spells{}, game.Players[0].CastingSkill, func (spell spellbook.Spell, picked bool){
        if picked {
            // FIXME: do something
        }
    }))
}

func (game *Game) CreateOutpost(settlers *playerlib.Unit, player *playerlib.Player){
    newCity := citylib.MakeCity("New City", settlers.X, settlers.Y, settlers.Unit.Race)
    newCity.Plane = settlers.Plane
    newCity.Population = 1000

    player.RemoveUnit(settlers)
    player.SelectedStack = nil
    game.HudUI = game.MakeHudUI()
    cityPtr := player.AddCity(newCity)

    stack := player.FindStack(newCity.X, newCity.Y)

    select {
        case game.Events<- &GameEventNewOutpost{City: cityPtr, Stack: stack}:
        default:
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
        // TODO
    }))

    // spell button
    elements = append(elements, makeButton(2, 47, 4, false, func(){
        game.ShowSpellBookCastUI()
    }))

    // army button
    elements = append(elements, makeButton(3, 89, 4, false, func(){
        // TODO
    }))

    // cities button
    elements = append(elements, makeButton(4, 140, 4, false, func(){
        // TODO
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
        // TODO
    }))

    if len(game.Players) > 0 && game.Players[0].SelectedStack != nil {
        stack := game.Players[0].SelectedStack

        unitX1 := 246
        unitY1 := 79

        unitX := unitX1
        unitY := unitY1

        row := 0
        for _, unit := range stack.Units() {
            // show a unit element for each unit in the stack
            // image index increases by 1 for each unit, indexes 24-32
            unitBackground, _ := game.ImageCache.GetImage("main.lbx", 24, 0)
            unitRect := util.ImageRect(unitX, unitY, unitBackground)
            elements = append(elements, &uilib.UIElement{
                Rect: unitRect,
                LeftClick: func(this *uilib.UIElement){
                    stack.ToggleActive(unit)
                },
                RightClick: func(this *uilib.UIElement){
                    ui.AddElements(game.MakeUnitContextMenu(ui, unit))
                },
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(unitRect.Min.X), float64(unitRect.Min.Y))
                    screen.DrawImage(unitBackground, &options)

                    options.GeoM.Translate(1, 1)

                    if stack.IsActive(unit){
                        unitBack, _ := GetUnitBackgroundImage(unit.Banner, &game.ImageCache)
                        screen.DrawImage(unitBack, &options)
                    }

                    options.GeoM.Translate(1, 1)
                    unitImage, _ := GetUnitImage(unit.Unit, &game.ImageCache)
                    screen.DrawImage(unitImage, &options)
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
                game.DoNextUnit()
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
                game.DoNextUnit()
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
                game.DoNextUnit()
            },
        })

        // FIXME: use index 15 to show inactive build button
        buildImages, _ := game.ImageCache.GetImages("main.lbx", 11)
        buildIndex := 0
        buildRect := util.ImageRect(280, 186, buildImages[0])
        buildCounter := uint64(0)
        elements = append(elements, &uilib.UIElement{
            Rect: buildRect,
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                colorScale := ebiten.ColorScale{}

                if buildCounter > 0 {
                    v := float32(1 + (math.Sin(float64(buildCounter / 4)) / 2 + 0.5) / 2)
                    colorScale.Scale(v, v, v, 1)
                }

                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(buildRect.Min.X), float64(buildRect.Min.Y))
                options.ColorScale.ScaleWithColorScale(colorScale)
                screen.DrawImage(buildImages[buildIndex], &options)
            },
            Inside: func(this *uilib.UIElement, x int, y int){
                buildCounter += 1
            },
            NotInside: func(this *uilib.UIElement){
                buildCounter = 0
            },
            LeftClick: func(this *uilib.UIElement){
                buildIndex = 1
            },
            LeftClickRelease: func(this *uilib.UIElement){
                buildIndex = 0

                player := game.Players[0]
                if player.SelectedStack != nil {
                    // search for the settlers (the only unit with the create outpost ability
                    for _, settlers := range player.SelectedStack.ActiveUnits() {
                        // FIXME: check if this tile is valid to build an outpost on
                        if settlers.Unit.HasAbility(units.AbilityCreateOutpost) {
                            game.CreateOutpost(settlers, player)
                            break
                        }
                    }
                }
            },
        })

    } else {
        // next turn
        nextTurnImage, _ := game.ImageCache.GetImage("main.lbx", 35, 0)
        nextTurnRect := image.Rect(240, 174, 240 + nextTurnImage.Bounds().Dx(), 174 + nextTurnImage.Bounds().Dy())
        elements = append(elements, &uilib.UIElement{
            Rect: nextTurnRect,
            LeftClick: func(this *uilib.UIElement){
                game.DoNextTurn()
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
            },
        })

        elements = append(elements, &uilib.UIElement{
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                goldFood, _ := game.ImageCache.GetImage("main.lbx", 34, 0)
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(240, 77)
                screen.DrawImage(goldFood, &options)

                game.InfoFontYellow.PrintCenter(screen, 278, 103, 1, ebiten.ColorScale{}, "1 Gold")
                game.InfoFontYellow.PrintCenter(screen, 278, 135, 1, ebiten.ColorScale{}, "1 Food")
                game.InfoFontYellow.PrintCenter(screen, 278, 167, 1, ebiten.ColorScale{}, "1 Mana")
            },
        })
    }

    elements = append(elements, &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            game.WhiteFont.Print(screen, 257, 68, 1, ebiten.ColorScale{}, fmt.Sprintf("%v GP", game.Players[0].Gold))
        },
    })

    elements = append(elements, &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            game.WhiteFont.Print(screen, 298, 68, 1, ebiten.ColorScale{}, fmt.Sprintf("%v MP", game.Players[0].Mana))
        },
    })

    ui.SetElementsFromArray(elements)

    return ui
}

func (game *Game) DoNextUnit(){
    if len(game.Players) > 0 {
        game.Players[0].SelectedStack = nil
    }

    game.HudUI = game.MakeHudUI()
}

func (game *Game) DoNextTurn(){
    // FIXME

    if len(game.Players) > 0 {
        player := game.Players[0]
        if len(player.Stacks) > 0 {
            player.SelectedStack = player.Stacks[0]
            game.CenterCamera(player.SelectedStack.X(), player.SelectedStack.Y())
        } else {
            game.CenterCamera(player.Cities[0].X, player.Cities[0].Y)
        }
        game.HudUI = game.MakeHudUI()
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

    fogN := func(x int, y int) bool {
        if y == 0 {
            return false
        }

        if x >= len(fog) || y - 1 >= len(fog[x]) {
            return false
        }

        return !fog[x][y - 1]
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
        if x == len(fog) - 1 {
            return false
        }

        if x + 1 >= len(fog) || y >= len(fog[x + 1]) {
            return false
        }

        return !fog[x + 1][y]
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
        if y == len(fog[0]) - 1 {
            return false
        }

        if x >= len(fog) || y + 1 >= len(fog[x]) {
            return false
        }

        return !fog[x][y + 1]
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
        if x == 0 {
            return false
        }

        if x - 1 < 0 || y >= len(fog[x - 1]) {
            return false
        }

        return !fog[x - 1][y]
    }

    for x := 0; x < tilesPerRow; x++ {
        for y := 0; y < tilesPerColumn; y++ {

            tileX := x + overworld.CameraX
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
    Map *Map
    Cities []*citylib.City
    Stacks []*playerlib.UnitStack
    SelectedStack *playerlib.UnitStack
    ImageCache *util.ImageCache
    Fog [][]bool
    ShowAnimation bool
    FogBlack *ebiten.Image
}

func (overworld *Overworld) DrawMinimap(screen *ebiten.Image){
    overworld.Map.DrawMinimap(screen, overworld.Cities, overworld.CameraX, overworld.CameraY)
}

func (overworld *Overworld) DrawOverworld(screen *ebiten.Image, geom ebiten.GeoM){
    overworld.Map.Draw(overworld.CameraX, overworld.CameraY, overworld.Counter / 8, screen, geom)

    tileWidth := float64(overworld.Map.TileWidth())
    tileHeight := float64(overworld.Map.TileHeight())

    convertTileCoordinates := func(x float64, y float64) (float64, float64) {
        outX := (x - float64(overworld.CameraX)) * tileWidth
        outY := (y - float64(overworld.CameraY)) * tileHeight
        return outX, outY
    }

    for _, city := range overworld.Cities {
        var cityPic *ebiten.Image
        var err error
        if city.Wall {
            cityPic, err = GetCityWallImage(city.GetSize(), overworld.ImageCache)
        } else {
            cityPic, err = GetCityNoWallImage(city.GetSize(), overworld.ImageCache)
        }

        if err == nil {
            var options ebiten.DrawImageOptions
            x, y := convertTileCoordinates(float64(city.X), float64(city.Y))
            options.GeoM = geom
            // draw the city in the center of the tile
            // first compute center of tile
            options.GeoM.Translate(x + tileWidth / 2.0, y + tileHeight / 2.0)
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
        if stack != overworld.SelectedStack || overworld.ShowAnimation || overworld.Counter / 55 % 2 == 0 {
            var options ebiten.DrawImageOptions
            options.GeoM = geom
            x, y := convertTileCoordinates(float64(stack.X()) + stack.OffsetX(), float64(stack.Y()) + stack.OffsetY())
            options.GeoM.Translate(x, y)

            unitBack, err := GetUnitBackgroundImage(stack.Leader().Banner, overworld.ImageCache)
            if err == nil {
                screen.DrawImage(unitBack, &options)
            }

            pic, err := GetUnitImage(stack.Leader().Unit, overworld.ImageCache)
            if err == nil {
                options.GeoM.Translate(1, 1)
                screen.DrawImage(pic, &options)
            }
        }
    }

    if overworld.Fog != nil {
        overworld.DrawFog(screen, geom)
    }
}

func (game *Game) Draw(screen *ebiten.Image){
    game.Drawer(screen, game)
}

func (game *Game) DrawGame(screen *ebiten.Image){

    var cities []*citylib.City
    var stacks []*playerlib.UnitStack
    var selectedStack *playerlib.UnitStack
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
        Map: game.Map,
        Cities: cities,
        Stacks: stacks,
        SelectedStack: selectedStack,
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
    miniHeight := 30
    mini := screen.SubImage(image.Rect(int(mx), int(my), int(mx) + miniWidth, int(my) + miniHeight)).(*ebiten.Image)
    overworld.DrawMinimap(mini)

    game.HudUI.Draw(game.HudUI, screen)
}
