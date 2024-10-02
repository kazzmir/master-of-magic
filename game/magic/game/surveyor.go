package game

import (
    "fmt"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"

    "github.com/hajimehoshi/ebiten/v2"
)

func makeSurveyorFont(cache *lbx.LbxCache) *font.Font {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        return nil
    }

    white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
    palette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        white, white, white,
        white, white, white,
    }

    return font.MakeOptimizedFontWithPalette(fonts[4], palette)
}

func (game *Game) doSurveyor(yield coroutine.YieldFunc) {
    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    var cities []*citylib.City
    var fog [][]bool

    for i, player := range game.Players {
        for _, city := range player.Cities {
            if city.Plane == game.Plane {
                cities = append(cities, city)
            }
        }

        if i == 0 {
            fog = player.GetFog(game.Plane)
        }
    }

    surveyorFont := makeSurveyorFont(game.Cache)

    // just draw cities and land, but no units
    overworld := Overworld{
        CameraX: game.cameraX,
        CameraY: game.cameraY,
        Counter: game.Counter,
        Map: game.Map,
        Cities: cities,
        Stacks: nil,
        SelectedStack: nil,
        ImageCache: &game.ImageCache,
        Fog: fog,
        ShowAnimation: game.State == GameStateUnitMoving,
        FogBlack: game.GetFogImage(),
    }

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            mainHud, _ := game.ImageCache.GetImage("main.lbx", 0, 0)
            screen.DrawImage(mainHud, &options)

            landImage, _ := game.ImageCache.GetImage("main.lbx", 57, 0)
            options.GeoM.Translate(240, 77)
            screen.DrawImage(landImage, &options)

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })

            game.WhiteFont.PrintRight(screen, 276, 68, 1, ebiten.ColorScale{}, fmt.Sprintf("%v GP", game.Players[0].Gold))
            game.WhiteFont.PrintRight(screen, 313, 68, 1, ebiten.ColorScale{}, fmt.Sprintf("%v MP", game.Players[0].Mana))

            surveyorFont.PrintCenter(screen, 280, 81, 1, ebiten.ColorScale{}, "Surveyor")
        },
    }

    game.Drawer = func(screen *ebiten.Image, game *Game){
        overworld.DrawOverworld(screen, ebiten.GeoM{})

        var miniGeom ebiten.GeoM
        miniGeom.Translate(250, 20)
        mx, my := miniGeom.Apply(0, 0)
        miniWidth := 60
        miniHeight := 31
        mini := screen.SubImage(image.Rect(int(mx), int(my), int(mx) + miniWidth, int(my) + miniHeight)).(*ebiten.Image)
        overworld.DrawMinimap(mini)

        ui.Draw(ui, screen)
    }

    quit := false
    for !quit {
        yield()
    }
}
