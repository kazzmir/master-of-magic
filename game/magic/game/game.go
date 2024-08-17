package game

import (
    "fmt"
    "image/color"
    "strings"
    "log"

    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/hajimehoshi/ebiten/v2"
)

type ImageCache struct {
    LbxCache *lbx.LbxCache
    Cache map[string][]*ebiten.Image
}

func (cache *ImageCache) GetImages(lbxPath string, index int) ([]*ebiten.Image, error) {
    lbxPath = strings.ToLower(lbxPath)
    key := fmt.Sprintf("%s:%d", lbxPath, index)

    if images, ok := cache.Cache[key]; ok {
        return images, nil
    }

    lbxFile, err := cache.LbxCache.GetLbxFile(lbxPath)
    if err != nil {
        return nil, err
    }

    sprites, err := lbxFile.ReadImages(index)
    if err != nil {
        return nil, err
    }

    var out []*ebiten.Image
    for i := 0; i < len(sprites); i++ {
        out = append(out, ebiten.NewImageFromImage(sprites[i]))
    }

    cache.Cache[key] = out

    return out, nil
}

func (cache *ImageCache) GetImage(lbxFile string, spriteIndex int, animationIndex int) (*ebiten.Image, error) {
    images, err := cache.GetImages(lbxFile, spriteIndex)
    if err != nil {
        return nil, err
    }

    if animationIndex < len(images) {
        return images[animationIndex], nil
    }

    return nil, fmt.Errorf("invalid animation index: %d for %v:%v", animationIndex, lbxFile, spriteIndex)
}

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
    game := &Game{
        active: false,
        Map: MakeMap(),
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

func (game *Game) Update(){
}

func (game *Game) GetMainImage(index int) (*ebiten.Image, error) {
    image, err := game.ImageCache.GetImage("main.lbx", index, 0)

    if err != nil {
        log.Printf("Error: image in main.lbx is missing: %v\n", err)
    }

    return image, err
}

func (game *Game) Draw(screen *ebiten.Image){
    var options ebiten.DrawImageOptions

    game.Map.Draw(screen)

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
