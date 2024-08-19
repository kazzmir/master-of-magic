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
)

type Unit struct {
    Unit units.Unit
    Banner data.BannerType
    X int
    Y int
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
}

func (player *Player) AddCity(city City) {
    player.Cities = append(player.Cities, &city)
}

func (player *Player) AddUnit(unit Unit) {
    player.Units = append(player.Units, &unit)
}

type GameState int
const (
    GameStateRunning GameState = iota
)

type Game struct {
    active bool

    ImageCache ImageCache
    WhiteFont *font.Font

    InfoFontYellow *font.Font
    Counter uint64

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

    return GameStateRunning
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

func (game *Game) Draw(screen *ebiten.Image){
    game.Map.Draw(0, 0, game.Counter / 4, screen)

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
                options.GeoM.Translate(float64(city.X * game.Map.TileWidth()), float64(city.Y * game.Map.TileHeight()))
                screen.DrawImage(cityPic, &options)
            }
        }

        for _, unit := range player.Units {
            var options ebiten.DrawImageOptions
            unitBack, err := game.GetUnitBackgroundImage(unit.Banner)
            if err == nil {
                options.GeoM.Translate(float64(unit.X * game.Map.TileWidth()), float64(unit.Y * game.Map.TileHeight()))
                screen.DrawImage(unitBack, &options)
            }

            pic, err := game.GetUnitImage(unit.Unit)
            if err == nil {
                options.GeoM.Translate(1, 1)
                screen.DrawImage(pic, &options)
            }
        }
    }

    // FIXME: render fog

    game.DrawHud(screen)
}
