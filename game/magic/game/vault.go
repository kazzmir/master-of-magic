package game

import (
    "log"
    "image/color"

    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
)

type VaultFonts struct {
    ItemName *font.Font
    PowerFont *font.Font
}

func makeFonts(cache *lbx.LbxCache) *VaultFonts {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Could not read fonts: %v", err)
        return &VaultFonts{}
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Could not read fonts: %v", err)
        return &VaultFonts{}
    }

    orange := color.RGBA{R: 0xc7, G: 0x82, B: 0x1b, A: 0xff}
    namePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        util.Lighten(orange, 0),
        util.Lighten(orange, 20),
        util.Lighten(orange, 50),
        util.Lighten(orange, 80),
        orange,
        orange,
    }

    itemName := font.MakeOptimizedFontWithPalette(fonts[4], namePalette)
    powerFont := font.MakeOptimizedFont(fonts[2])

    return &VaultFonts{
        ItemName: itemName,
        PowerFont: powerFont,
    }
}

func (game *Game) showVaultScreen(createdArtifact *artifact.Artifact, heroes []*hero.Hero) (func(coroutine.YieldFunc), func (*ebiten.Image)) {

    imageCache := util.MakeImageCache(game.Cache)

    fonts := makeFonts(game.Cache)

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            background, _ := imageCache.GetImage("armylist.lbx", 5, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(data.ScreenWidth / 2 - background.Bounds().Dx() / 2), 2)
            screen.DrawImage(background, &options)

            if createdArtifact != nil {
                itemBackground, _ := imageCache.GetImage("itemisc.lbx", 25, 0)
                options.GeoM.Translate(32, 48)
                screen.DrawImage(itemBackground, &options)

                itemImage, _ := imageCache.GetImage("items.lbx", createdArtifact.Image, 0)
                options.GeoM.Translate(10, 8)
                screen.DrawImage(itemImage, &options)

                x, y := options.GeoM.Apply(float64(itemImage.Bounds().Max.X) + 3, 4)

                fonts.ItemName.Print(screen, x, y, 1, ebiten.ColorScale{}, createdArtifact.Name)
            }
        },
    }

    quit := false

    drawer := func (screen *ebiten.Image){
        ui.Draw(ui, screen)
    }

    logic := func (yield coroutine.YieldFunc) {
        for !quit {
            yield()
        }
    }

    return logic, drawer
}
