package game

import (
    "image/color"
    "log"

    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/hajimehoshi/ebiten/v2"
)

type GameState int
const (
    GameStateRunning GameState = iota
)

type Game struct {
    active bool

    ImageCache ImageCache
    WhiteFont *font.Font

    InfoFontYellow *font.Font

    // FIXME: need one map for arcanus and one for myrran
    Map *Map
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
    return GameStateRunning
}

func (game *Game) GetMainImage(index int) (*ebiten.Image, error) {
    image, err := game.ImageCache.GetImage("main.lbx", index, 0)

    if err != nil {
        log.Printf("Error: image in main.lbx is missing: %v", err)
    }

    return image, err
}

type Flag int
const (
    FlagBlue Flag = iota
    FlagGreen
    FlagPurple
    FlagRed
    FlagYellow
    FlagBrown
)

type CitySize int
const (
    CitySizeHamlet CitySize = iota
    CitySizeVillage
    CitySizeTown
    CitySizeCity
    CitySizeCapital
)

func (game *Game) GetUnitBackgroundImage(flag Flag) (*ebiten.Image, error) {
    index := -1
    switch flag {
        case FlagBlue: index = 14
        case FlagGreen: index = 15
        case FlagPurple: index = 16
        case FlagRed: index = 17
        case FlagYellow: index = 18
        case FlagBrown: index = 19
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

func (game *Game) Draw(screen *ebiten.Image){
    var options ebiten.DrawImageOptions

    game.Map.Draw(0, 0, screen)

    city1, err := game.GetCityNoWallImage(CitySizeCity)
    if err == nil {
        tileX := 4
        tileY := 4

        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(tileX * game.Map.TileWidth()), float64(tileY * game.Map.TileHeight()))
        screen.DrawImage(city1, &options)
    }

    unitBack, err := game.GetUnitBackgroundImage(FlagBlue)
    if err == nil {
        unitTileX := 8
        unitTileY := 8

        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(unitTileX * game.Map.TileWidth()), float64(unitTileY * game.Map.TileHeight()))
        screen.DrawImage(unitBack, &options)

        bat, err := game.GetUnitImage(units.DoomBat)
        if err == nil {
            options.GeoM.Translate(1, 1)
            screen.DrawImage(bat, &options)
        }

    }

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
