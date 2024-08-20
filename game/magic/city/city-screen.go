package city

import (
    "log"
    "fmt"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"

    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/hajimehoshi/ebiten/v2"
)

type CityScreen struct {
    LbxCache *lbx.LbxCache
    ImageCache util.ImageCache
    Font *font.Font
    City *City

    Counter uint64
}

func MakeCityScreen(cache *lbx.LbxCache, city *City) *CityScreen {

    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    fonts, err := fontLbx.ReadFonts(0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
    }

    yellowPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xfa, G: 0xe1, B: 0x16, A: 0xff},
            color.RGBA{R: 0xff, G: 0xef, B: 0x2f, A: 0xff},
            color.RGBA{R: 0xff, G: 0xef, B: 0x2f, A: 0xff},
            color.RGBA{R: 0xe0, G: 0x8a, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xd2, G: 0x7f, B: 0x0, A: 0xff},
            color.RGBA{R: 0xd2, G: 0x7f, B: 0x0, A: 0xff},
            color.RGBA{R: 0x99, G: 0x4f, B: 0x0, A: 0xff},
            color.RGBA{R: 0x99, G: 0x4f, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0x99, G: 0x4f, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        }

    bigFont := font.MakeOptimizedFontWithPalette(fonts[5], yellowPalette)

    cityScreen := &CityScreen{
        LbxCache: cache,
        ImageCache: util.MakeImageCache(cache),
        City: city,
        Font: bigFont,
    }
    return cityScreen
}

func (cityScreen *CityScreen) Update() {
    cityScreen.Counter += 1
}

func (cityScreen *CityScreen) Draw(screen *ebiten.Image) {
    animationCounter := cityScreen.Counter / 8

    // 5 is grasslands
    landBackground, err := cityScreen.ImageCache.GetImage("cityscap.lbx", 0, 4)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(4, 102)
        screen.DrawImage(landBackground, &options)
    }

    river, err := cityScreen.ImageCache.GetImages("cityscap.lbx", 3)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(5, 100)
        screen.DrawImage(river[0], &options)

        index := animationCounter % 5

        screen.DrawImage(river[index + 1], &options)
    }

    ui, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 6, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        screen.DrawImage(ui, &options)
    }

    cityScreen.Font.Print(screen, 20, 3, 1, fmt.Sprintf("%v of %s", cityScreen.City.GetSize(), cityScreen.City.Name))

    smallFood, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 40, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(6, 52)
        for i := 0; i < cityScreen.City.FoodProduction; i++ {
            screen.DrawImage(smallFood, &options)
            options.GeoM.Translate(float64(smallFood.Bounds().Dx() + 1), 0)
        }
    }

    // big food is 88
    // hammer is 41
    smallWork, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 41, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(6, 60)
        for i := 0; i < cityScreen.City.WorkProduction; i++ {
            screen.DrawImage(smallWork, &options)
            options.GeoM.Translate(float64(smallWork.Bounds().Dx() + 1), 0)
        }
    }

    // big hammer is 89
    // coin is 42
    smallCoin, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 42, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(6, 68)
        for i := 0; i < cityScreen.City.MoneyProduction; i++ {
            screen.DrawImage(smallCoin, &options)
            options.GeoM.Translate(float64(smallCoin.Bounds().Dx() + 1), 0)
        }
    }

    // big coin is 90
    // small magic is 43
    smallMagic, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 43, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(6, 76)
        for i := 0; i < cityScreen.City.MagicProduction; i++ {
            screen.DrawImage(smallMagic, &options)
            options.GeoM.Translate(float64(smallMagic.Bounds().Dx() + 1), 0)
        }
    }
    // big magic is 91
}
