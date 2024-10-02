package game

import (
    "fmt"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
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

    selectedPoint := image.Pt(-1, -1)

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            mainHud, _ := game.ImageCache.GetImage("main.lbx", 0, 0)
            screen.DrawImage(mainHud, &options)

            landImage, _ := game.ImageCache.GetImage("main.lbx", 57, 0)
            options.GeoM.Translate(240, 77)
            screen.DrawImage(landImage, &options)

            cancelBackground, _ := game.ImageCache.GetImage("main.lbx", 47, 0)
            options.GeoM.Reset()
            options.GeoM.Translate(240, 174)
            screen.DrawImage(cancelBackground, &options)

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })

            game.WhiteFont.PrintRight(screen, 276, 68, 1, ebiten.ColorScale{}, fmt.Sprintf("%v GP", game.Players[0].Gold))
            game.WhiteFont.PrintRight(screen, 313, 68, 1, ebiten.ColorScale{}, fmt.Sprintf("%v MP", game.Players[0].Mana))

            surveyorFont.PrintCenter(screen, 280, 81, 1, ebiten.ColorScale{}, "Surveyor")

            if selectedPoint.X >= 0 && selectedPoint.X < game.Map.Width() && selectedPoint.Y >= 0 && selectedPoint.Y < game.Map.Height() {
                if fog[selectedPoint.X][selectedPoint.Y] {
                    game.WhiteFont.PrintCenter(screen, 280, 93, 1, ebiten.ColorScale{}, fmt.Sprintf("Forest"))
                }
            }
        },
    }

    ui.SetElementsFromArray(nil)

    makeButton := func(lbxIndex int, x int, y int) *uilib.UIElement {
        button, _ := game.ImageCache.GetImage("main.lbx", lbxIndex, 0)
        return &uilib.UIElement{
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(x), float64(y))
                screen.DrawImage(button, &options)
            },
        }
    }

    // game
    ui.AddElement(makeButton(1, 7, 4))

    // spells
    ui.AddElement(makeButton(2, 47, 4))

    // army button
    ui.AddElement(makeButton(3, 89, 4))

    // cities button
    ui.AddElement(makeButton(4, 140, 4))

    // magic button
    ui.AddElement(makeButton(5, 184, 4))

    // info button
    ui.AddElement(makeButton(6, 226, 4))

    // plane button
    ui.AddElement(makeButton(7, 270, 4))

    quit := false

    // cancel button at bottom
    cancel, _ := game.ImageCache.GetImages("main.lbx", 41)
    cancelIndex := 0
    cancelRect := util.ImageRect(263, 182, cancel[0])
    ui.AddElement(&uilib.UIElement{
        Rect: cancelRect,
        LeftClick: func(element *uilib.UIElement){
            cancelIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            cancelIndex = 0
            quit = true
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(cancelRect.Min.X), float64(cancelRect.Min.Y))
            screen.DrawImage(cancel[cancelIndex], &options)
        },
    })

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

    moveCamera := image.Pt(game.cameraX, game.cameraY)
    for !quit {
        overworld.Counter += 1

        ui.StandardUpdate()

        x, y := ebiten.CursorPosition()

        if overworld.Counter % 5 == 0 && (moveCamera.X != game.cameraX || moveCamera.Y != game.cameraY) {
            if moveCamera.X < game.cameraX {
                game.cameraX -= 1
            } else if moveCamera.X > game.cameraX {
                game.cameraX += 1
            }

            if moveCamera.Y < game.cameraY {
                game.cameraY -= 1
            } else if moveCamera.Y > game.cameraY {
                game.cameraY += 1
            }

            overworld.CameraX = game.cameraX
            overworld.CameraY = game.cameraY
        }

        // within the viewable area
        if x < 240 && y > 18 {
            newX := game.cameraX + x / game.Map.TileWidth()
            newY := game.cameraY + y / game.Map.TileHeight()
            selectedPoint = image.Pt(newX, newY)

            // right click should move the camera
            rightClick := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)
            if rightClick {
                moveCamera = selectedPoint.Add(image.Pt(-5, -5))
                if moveCamera.X < 0 {
                    moveCamera.X = 0
                }
                if moveCamera.Y < 0 {
                    moveCamera.Y = 0
                }
                if moveCamera.Y >= game.Map.Height() - 11 {
                    moveCamera.Y = game.Map.Height() - 11
                }
            }
        }

        yield()
    }
}
