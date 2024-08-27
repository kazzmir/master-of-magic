package game

import (
    "image/color"
    "log"

    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/lib/lbx"
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

type GameState int
const (
    GameStateRunning GameState = iota
    GameStateUnitMoving
)

type Game struct {
    active bool

    ImageCache util.ImageCache
    WhiteFont *font.Font

    InfoFontYellow *font.Font
    Counter uint64
    Fog *ebiten.Image
    State GameState
    Plane data.Plane

    cameraX int
    cameraY int


    // FIXME: need one map for arcanus and one for myrran
    Map *Map

    Players []*player.Player
}

func (game *Game) MakeFog() [][]bool {
    fog := make([][]bool, game.Map.Width())
    for x := 0; x < game.Map.Width(); x++ {
        fog[x] = make([]bool, game.Map.Height())
    }

    return fog
}

func (game *Game) AddPlayer(wizard setup.WizardCustom) *player.Player{
    newPlayer := &player.Player{
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

    fontLbx, err := lbxCache.GetLbxFile("FONTS.LBX")
    if err != nil {
        return nil
    }

    fonts, err := fontLbx.ReadFonts(0)
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
        active: false,
        Map: MakeMap(terrainData),
        State: GameStateRunning,
        ImageCache: util.MakeImageCache(lbxCache),
        InfoFontYellow: infoFontYellow,
        WhiteFont: whiteFont,
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

    tilesPerRow := game.Map.TilesPerRow(data.ScreenWidth)
    tilesPerColumn := game.Map.TilesPerColumn(data.ScreenHeight)

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

        if len(game.Players) > 0 && game.Players[0].SelectedUnit != nil {
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

            tileX := x + overworld.CameraX
            tileY := y + overworld.CameraY

            options.GeoM = geom
            options.GeoM.Translate(float64(x * tileWidth), float64(y * tileHeight))

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
    Units []*player.Unit
    SelectedUnit *player.Unit
    ImageCache *util.ImageCache
    Fog [][]bool
    ShowAnimation bool
    FogBlack *ebiten.Image
}

func (overworld *Overworld) DrawOverworld(screen *ebiten.Image, geom ebiten.GeoM){
    overworld.Map.Draw(overworld.CameraX, overworld.CameraY, overworld.Counter / 8, screen, geom)

    tileWidth := overworld.Map.TileWidth()
    tileHeight := overworld.Map.TileHeight()

    convertTileCoordinates := func(x int, y int) (int, int) {
        outX := (x - overworld.CameraX) * tileWidth
        outY := (y - overworld.CameraY) * tileHeight
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
            x, y := convertTileCoordinates(city.X, city.Y)
            options.GeoM = geom
            // draw the city in the center of the tile
            // first compute center of tile
            options.GeoM.Translate(float64(x) + float64(tileWidth) / 2.0, float64(y) + float64(tileHeight) / 2.0)
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

    for _, unit := range overworld.Units {
        if overworld.SelectedUnit != unit || overworld.ShowAnimation || overworld.Counter / 55 % 2 == 0 {
            var options ebiten.DrawImageOptions
            options.GeoM = geom
            x, y := convertTileCoordinates(unit.X, unit.Y)
            options.GeoM.Translate(float64(x), float64(y))

            if overworld.ShowAnimation && overworld.SelectedUnit == unit {
                dx := float64(float64(unit.MoveX - unit.X) * float64(tileWidth * unit.Movement) / float64(player.MovementLimit))
                dy := float64(float64(unit.MoveY - unit.Y) * float64(tileHeight * unit.Movement) / float64(player.MovementLimit))
                options.GeoM.Translate(dx, dy)
            }

            unitBack, err := GetUnitBackgroundImage(unit.Banner, overworld.ImageCache)
            if err == nil {
                screen.DrawImage(unitBack, &options)
            }

            pic, err := GetUnitImage(unit.Unit, overworld.ImageCache)
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

    var cities []*citylib.City
    var units []*player.Unit
    var selectedUnit *player.Unit
    var fog [][]bool

    for i, player := range game.Players {
        for _, city := range player.Cities {
            if city.Plane == game.Plane {
                cities = append(cities, city)
            }
        }

        for _, unit := range player.Units {
            if unit.Plane == game.Plane {
                units = append(units, unit)
            }
        }

        if i == 0 {
            selectedUnit = player.SelectedUnit
            fog = player.GetFog(game.Plane)
        }
    }

    overworld := Overworld{
        CameraX: game.cameraX,
        CameraY: game.cameraY,
        Counter: game.Counter,
        Map: game.Map,
        Cities: cities,
        Units: units,
        SelectedUnit: selectedUnit,
        ImageCache: &game.ImageCache,
        Fog: fog,
        ShowAnimation: game.State == GameStateUnitMoving,
        FogBlack: game.GetFogImage(),
    }

    overworld.DrawOverworld(screen, ebiten.GeoM{})

    game.DrawHud(screen)
}
