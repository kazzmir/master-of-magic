package game

import (
    "image/color"
    "log"

    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Unit struct {
    Unit units.Unit
    Banner data.BannerType
    X int
    Y int
    Id uint64

    Movement int
    MoveX int
    MoveY int
}

const MovementLimit = 10

func (unit *Unit) Move(dx int, dy int){
    unit.Movement = MovementLimit

    unit.MoveX = unit.X
    unit.MoveY = unit.Y

    unit.X += dx
    unit.Y += dy

    // FIXME: can't move off of map

    if unit.X < 0 {
        unit.X = 0
    }

    if unit.Y < 0 {
        unit.Y = 0
    }
}

type City struct {
    Population int
    Wall bool
    X int
    Y int
}

func (city *City) GetSize() CitySize {
    if city.Population < 5000 {
        return CitySizeHamlet
    }

    if city.Population < 9000 {
        return CitySizeVillage
    }

    if city.Population < 13000 {
        return CitySizeTown
    }

    if city.Population < 17000 {
        return CitySizeCity
    }

    return CitySizeCapital
}

type Player struct {
    // matrix the same size as the map, where true means the player can see the tile
    // and false means the tile has not yet been discovered
    ArcanusFog [][]bool
    MyrrorFog [][]bool

    Wizard setup.WizardCustom

    Units []*Unit
    Cities []*City

    UnitId uint64
    SelectedUnit *Unit
}

func (player *Player) SetSelectedUnit(unit *Unit){
    player.SelectedUnit = unit
}

/* make anything within the given radius viewable by the player */
func (player *Player) LiftFog(x int, y int, radius int){

    // FIXME: make this a parameter
    fog := player.ArcanusFog

    for dx := -radius; dx <= radius; dx++ {
        for dy := -radius; dy <= radius; dy++ {
            if x + dx < 0 || x + dx >= len(fog) || y + dy < 0 || y + dy >= len(fog[0]) {
                continue
            }

            if dx * dx + dy * dy <= radius * radius {
                fog[x + dx][y + dy] = true
            }
        }
    }

}

func (player *Player) AddCity(city City) {
    player.Cities = append(player.Cities, &city)
}

func (player *Player) AddUnit(unit Unit) *Unit {
    unit.Id = player.UnitId
    player.UnitId += 1
    unit_ptr := &unit
    player.Units = append(player.Units, unit_ptr)
    return unit_ptr
}

func (game *Game) GetFogImage() *ebiten.Image {
    if game.Fog != nil {
        return game.Fog
    }

    game.Fog = ebiten.NewImage(game.Map.TileWidth(), game.Map.TileHeight())
    game.Fog.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 0xff})
    return game.Fog
}

type GameState int
const (
    GameStateRunning GameState = iota
    GameStateUnitMoving
)

type Game struct {
    active bool

    ImageCache ImageCache
    WhiteFont *font.Font

    InfoFontYellow *font.Font
    Counter uint64
    Fog *ebiten.Image
    State GameState

    cameraX int
    cameraY int


    // FIXME: need one map for arcanus and one for myrran
    Map *Map

    Players []*Player
}

func (game *Game) MakeFog() [][]bool {
    fog := make([][]bool, game.Map.Width())
    for x := 0; x < game.Map.Width(); x++ {
        fog[x] = make([]bool, game.Map.Height())
    }

    return fog
}

func (game *Game) AddPlayer(wizard setup.WizardCustom) *Player{
    newPlayer := &Player{
        ArcanusFog: game.MakeFog(),
        MyrrorFog: game.MakeFog(),
        Wizard: wizard,
    }

    game.Players = append(game.Players, newPlayer)
    return newPlayer
}

func (game *Game) Load(cache *lbx.LbxCache) error {
    fontLbx, err := cache.GetLbxFile("FONTS.LBX")
    if err != nil {
        return err
    }

    fonts, err := fontLbx.ReadFonts(0)
    if err != nil {
        return err
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

    game.InfoFontYellow = font.MakeOptimizedFontWithPalette(fonts[0], yellowPalette)

    whitePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.White, color.White, color.White, color.White,
    }

    game.WhiteFont = font.MakeOptimizedFontWithPalette(fonts[0], whitePalette)

    return nil
}

func MakeGame(wizard setup.WizardCustom, lbxCache *lbx.LbxCache) *Game {

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

    game := &Game{
        active: false,
        Map: MakeMap(terrainData),
        State: GameStateRunning,
        ImageCache: ImageCache{
            LbxCache: lbxCache,
            Cache: make(map[string][]*ebiten.Image),
        },
    }
    return game
}

func (game *Game) IsActive() bool {
    return game.active
}

func (game *Game) Activate() {
    game.active = true
}

func (game *Game) Update() GameState {
    game.Counter += 1

    tilesPerRow := data.ScreenWidth / game.Map.TileWidth()
    tilesPerColumn := data.ScreenHeight / game.Map.TileHeight()

    if game.State == GameStateRunning {
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

        if game.Players[0].SelectedUnit != nil {
            unit := game.Players[0].SelectedUnit
            game.cameraX = unit.X - tilesPerRow / 2
            game.cameraY = unit.Y - tilesPerColumn / 2

            if game.cameraX < 0 {
                game.cameraX = 0
            }

            if game.cameraY < 0 {
                game.cameraY = 0
            }

            if dx != 0 || dy != 0 {
                unit.Move(dx, dy)
                game.Players[0].LiftFog(unit.X, unit.Y, 2)
                game.State = GameStateUnitMoving
            }
        }
    } else if game.State == GameStateUnitMoving {
        unit := game.Players[0].SelectedUnit
        unit.Movement -= 1
        if unit.Movement == 0 {
            game.State = GameStateRunning
        }
    }

    return game.State
}

func (game *Game) GetMainImage(index int) (*ebiten.Image, error) {
    image, err := game.ImageCache.GetImage("main.lbx", index, 0)

    if err != nil {
        log.Printf("Error: image in main.lbx is missing: %v", err)
    }

    return image, err
}

type CitySize int
const (
    CitySizeHamlet CitySize = iota
    CitySizeVillage
    CitySizeTown
    CitySizeCity
    CitySizeCapital
)

func (game *Game) GetUnitBackgroundImage(banner data.BannerType) (*ebiten.Image, error) {
    index := -1
    switch banner {
        case data.BannerBlue: index = 14
        case data.BannerGreen: index = 15
        case data.BannerPurple: index = 16
        case data.BannerRed: index = 17
        case data.BannerYellow: index = 18
        case data.BannerBrown: index = 19
    }

    image, err := game.ImageCache.GetImage("mapback.lbx", index, 0)
    if err != nil {
        log.Printf("Error: image in mapback.lbx is missing: %v", err)
    }

    return image, err
}

func (game *Game) GetUnitImage(unit units.Unit) (*ebiten.Image, error) {
    image, err := game.ImageCache.GetImage(unit.LbxFile, unit.Index, 0)

    if err != nil {
        log.Printf("Error: image in %v is missing: %v", unit.LbxFile, err)
    }

    return image, err
}

func (game *Game) GetCityNoWallImage(size CitySize) (*ebiten.Image, error) {
    var index int = 0

    switch size {
        case CitySizeHamlet: index = 0
        case CitySizeVillage: index = 1
        case CitySizeTown: index = 2
        case CitySizeCity: index = 3
        case CitySizeCapital: index = 4
    }

    // the city image is a sub-frame of animation 20
    return game.ImageCache.GetImage("mapback.lbx", 20, index)
}

func (game *Game) GetCityWallImage(size CitySize) (*ebiten.Image, error) {
    var index int = 0

    switch size {
        case CitySizeHamlet: index = 0
        case CitySizeVillage: index = 1
        case CitySizeTown: index = 2
        case CitySizeCity: index = 3
        case CitySizeCapital: index = 4
    }

    // the city image is a sub-frame of animation 21
    return game.ImageCache.GetImage("mapback.lbx", 21, index)
}

func (game *Game) DrawHud(screen *ebiten.Image){
    var options ebiten.DrawImageOptions

    // draw hud on top of map
    mainHud, err := game.GetMainImage(0)
    if err == nil {
        screen.DrawImage(mainHud, &options)
    }

    options.GeoM.Reset()
    x := float64(7)
    y := float64(4)
    options.GeoM.Translate(x, y)

    gameButton1, err := game.GetMainImage(1)
    if err == nil {
        screen.DrawImage(gameButton1, &options)
        x += float64(gameButton1.Bounds().Dx())
        options.GeoM.Translate(float64(gameButton1.Bounds().Dx()) + 1, 0)
    }

    spellButton, err := game.GetMainImage(2)
    if err == nil {
        screen.DrawImage(spellButton, &options)
        options.GeoM.Translate(float64(spellButton.Bounds().Dx()) + 1, 0)
    }

    armyButton, err := game.GetMainImage(3)
    if err == nil {
        screen.DrawImage(armyButton, &options)
        options.GeoM.Translate(float64(armyButton.Bounds().Dx()) + 1, 0)
    }

    cityButton, err := game.GetMainImage(4)
    if err == nil {
        screen.DrawImage(cityButton, &options)
        options.GeoM.Translate(float64(cityButton.Bounds().Dx()) + 1, 0)
    }

    magicButton, err := game.GetMainImage(5)
    if err == nil {
        screen.DrawImage(magicButton, &options)
        options.GeoM.Translate(float64(magicButton.Bounds().Dx()) + 1, 0)
    }

    infoButton, err := game.GetMainImage(6)
    if err == nil {
        screen.DrawImage(infoButton, &options)
        options.GeoM.Translate(float64(infoButton.Bounds().Dx()) + 1, 0)
    }

    planeButton, err := game.GetMainImage(7)
    if err == nil {
        screen.DrawImage(planeButton, &options)
    }

    options.GeoM.Reset()

    goldFood, err := game.GetMainImage(34)
    if err == nil {
        options.GeoM.Translate(240, 77)
        screen.DrawImage(goldFood, &options)
    }

    game.InfoFontYellow.PrintCenter(screen, 278, 103, 1, "1 Gold")
    game.InfoFontYellow.PrintCenter(screen, 278, 135, 1, "1 Food")
    game.InfoFontYellow.PrintCenter(screen, 278, 167, 1, "1 Mana")

    game.WhiteFont.Print(screen, 257, 68, 1, "75 GP")
    game.WhiteFont.Print(screen, 298, 68, 1, "0 MP")

    /*
    options.GeoM.Reset()
    options.GeoM.Translate(245, 180)
    screen.DrawImage(game.NextTurnBackground, &options)
    */

    nextTurn, err := game.GetMainImage(35)
    if err == nil {
        options.GeoM.Reset()
        options.GeoM.Translate(240, 174)
        screen.DrawImage(nextTurn, &options)
    }
}

func (game *Game) DrawFog(screen *ebiten.Image, fog [][]bool, cameraX int, cameraY int){

    fogImage := func(index int) *ebiten.Image {
        img, err := game.ImageCache.GetImage("mapback.lbx", index, 0)
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

    fogBlack := game.GetFogImage()

    tilesPerRow := data.ScreenWidth / game.Map.TileWidth()
    tilesPerColumn := data.ScreenHeight / game.Map.TileHeight()
    var options ebiten.DrawImageOptions

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

        return !fog[x - 1][y]
    }

    for x := 0; x < tilesPerRow; x++ {
        for y := 0; y < tilesPerColumn; y++ {

            tileX := x + cameraX
            tileY := y + cameraY

            options.GeoM.Reset()
            options.GeoM.Translate(float64(x * game.Map.TileWidth()), float64(y * game.Map.TileHeight()))

            // nw := fogNW(tileX, tileY)
            n := fogN(tileX, tileY)
            // ne := fogNE(tileX, tileY)
            e := fogE(tileX, tileY)
            // se := fogSE(tileX, tileY)
            s := fogS(tileX, tileY)
            // sw := fogSW(tileX, tileY)
            w := fogW(tileX, tileY)

            if fog[tileX][tileY] {

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
                screen.DrawImage(fogBlack, &options)
            }
        }
    }

}

func (game *Game) Draw(screen *ebiten.Image){
    game.Map.Draw(game.cameraX, game.cameraY, game.Counter / 4, screen)

    if len(game.Players) > 0 {
        player := game.Players[0]

        for _, city := range player.Cities {
            var cityPic *ebiten.Image
            var err error
            if city.Wall {
                cityPic, err = game.GetCityWallImage(city.GetSize())
            } else {
                cityPic, err = game.GetCityNoWallImage(city.GetSize())
            }

            if err == nil {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64((city.X - game.cameraX) * game.Map.TileWidth()), float64((city.Y - game.cameraY) * game.Map.TileHeight()))
                screen.DrawImage(cityPic, &options)
            }
        }

        for _, unit := range player.Units {
            if player.SelectedUnit != unit || game.State == GameStateUnitMoving || game.Counter / 55 % 2 == 0 {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64((unit.X - game.cameraX) * game.Map.TileWidth()), float64((unit.Y - game.cameraY) * game.Map.TileHeight()))

                if game.State == GameStateUnitMoving && player.SelectedUnit == unit {
                    dx := float64(float64(unit.MoveX - unit.X) * float64(game.Map.TileWidth() * unit.Movement) / float64(MovementLimit))
                    dy := float64(float64(unit.MoveY - unit.Y) * float64(game.Map.TileHeight() * unit.Movement) / float64(MovementLimit))
                    options.GeoM.Translate(dx, dy)
                }

                unitBack, err := game.GetUnitBackgroundImage(unit.Banner)
                if err == nil {
                    screen.DrawImage(unitBack, &options)
                }

                pic, err := game.GetUnitImage(unit.Unit)
                if err == nil {
                    options.GeoM.Translate(1, 1)
                    screen.DrawImage(pic, &options)
                }
            }
        }

        // FIXME: render the proper plane
        game.DrawFog(screen, player.ArcanusFog, game.cameraX, game.cameraY)
    }

    game.DrawHud(screen)
}
