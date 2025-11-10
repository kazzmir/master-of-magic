package game

import (
    _ "fmt"
    "image"

    "github.com/kazzmir/master-of-magic/lib/coroutine"
    // "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    _ "github.com/kazzmir/master-of-magic/game/magic/maplib"

    "github.com/hajimehoshi/ebiten/v2"
)

func (game *Game) ShowRoadBuilder(yield coroutine.YieldFunc, engineerStack *playerlib.UnitStack) {
    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    fonts := fontslib.MakeSurveyorFonts(game.Cache)

    var cityMap map[image.Point]*citylib.City

    makeOverworld := func () Overworld {
        cityMap = make(map[image.Point]*citylib.City)

        var cities []*citylib.City
        var stacks []*playerlib.UnitStack
        var fog data.FogMap

        for i, player := range game.Players {
            for _, city := range player.Cities {
                if city.Plane == game.Plane {
                    cities = append(cities, city)
                    cityMap[image.Pt(city.X, city.Y)] = city
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

        return Overworld{
            Camera: game.Camera,
            Counter: game.Counter,
            Map: game.CurrentMap(),
            Cities: cities,
            Stacks: stacks,
            SelectedStack: nil,
            ImageCache: &game.ImageCache,
            Fog: fog,
            ShowAnimation: game.State == GameStateUnitMoving,
            FogBlack: game.GetFogImage(),
        }
    }

    overworld := makeOverworld()

    selectedPoint := image.Pt(-1, -1)

    var cityInfoText font.WrappedText
    type Resources struct {
        Enabled bool
        MaximumPopulation int
        ProductionBonus int
        GoldBonus int
    }

    var resources Resources

    cancelBackground, _ := game.ImageCache.GetImage("main.lbx", 47, 0)

    ui := &uilib.UI{
        Cache: game.Cache,
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            mainHud, _ := game.ImageCache.GetImage("main.lbx", 0, 0)
            scale.DrawScaled(screen, mainHud, &options)

            landImage, _ := game.ImageCache.GetImage("main.lbx", 45, 0)
            options.GeoM.Translate(float64(240), float64(77))
            scale.DrawScaled(screen, landImage, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(240), float64(174))
            scale.DrawScaled(screen, cancelBackground, &options)

            ui.StandardDraw(screen)

            // player := game.Players[0]

            fonts.SurveyorFont.PrintCenter(screen, float64(280), float64(81), scale.ScaleAmount, ebiten.ColorScale{}, "Road Building")
        },
    }

    ui.SetElementsFromArray(nil)

    makeButton := func(lbxIndex int, x int, y int) *uilib.UIElement {
        button, _ := game.ImageCache.GetImage("main.lbx", lbxIndex, 0)
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(x), float64(y))
        return &uilib.UIElement{
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                scale.DrawScaled(screen, button, &options)
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
    ui.AddElement((func () *uilib.UIElement {
        buttons, _ := game.ImageCache.GetImages("main.lbx", 7)
        x := 270
        y := 4

        clicked := false

        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(x), float64(y))

        return &uilib.UIElement{
            Rect: util.ImageRect(x, y, buttons[0]),
            PlaySoundLeftClick: true,
            LeftClick: func(element *uilib.UIElement){
                clicked = true
            },
            LeftClickRelease: func(element *uilib.UIElement){
                clicked = false
                game.SwitchPlane()
                overworld = makeOverworld()
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                if clicked {
                    scale.DrawScaled(screen, buttons[1], &options)
                } else {
                    scale.DrawScaled(screen, buttons[0], &options)
                }
            },
        }
    })())

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
            scale.DrawScaled(screen, cancel[cancelIndex], &options)
        },
    })

    game.Drawer = func(screen *ebiten.Image, game *Game){
        overworld.Camera = game.Camera

        overworld.DrawOverworld(screen, ebiten.GeoM{})

        var miniGeom ebiten.GeoM
        miniGeom.Translate(float64(250), float64(20))
        mx, my := miniGeom.Apply(0, 0)
        miniWidth := 60
        miniHeight := 31
        mini := screen.SubImage(scale.ScaleRect(image.Rect(int(mx), int(my), int(mx) + miniWidth, int(my) + miniHeight))).(*ebiten.Image)
        overworld.DrawMinimap(mini)

        ui.Draw(ui, screen)
    }

    for !quit {
        overworld.Counter += 1
        zoomed := game.doInputZoom(yield)
        _ = zoomed

        ui.StandardUpdate()

        x, y := inputmanager.MousePosition()

        // within the viewable area
        if game.InOverworldArea(x, y) {
            newX, newY := game.ScreenToTile(float64(x), float64(y))
            newPoint := image.Pt(newX, newY)

            // right click should move the camera
            rightClick := inputmanager.RightClick()
            if rightClick /*|| zoomed*/ {
                game.doMoveCamera(yield, newX, newY)
            }

            if selectedPoint != newPoint {
                selectedPoint = newPoint
                resources.Enabled = false

                text := ""

                tile := game.CurrentMap().GetTile(newX, newY)

                if !tile.Tile.IsLand() {
                    text = "Cannot build cities on water."
                } else if cityMap[newPoint] == nil && game.NearCity(newPoint, 3, game.Plane) {
                    text = "Cities cannot be built less than 3 squares from any other city."
                } else {
                    text = "City Resources"
                    resources.Enabled = true
                    // FIXME: compute proper values for these
                    resources.MaximumPopulation = game.ComputeMaximumPopulation(newX, newY, game.Plane)
                    resources.ProductionBonus = game.CityProductionBonus(newX, newY, game.Plane)
                    resources.GoldBonus = game.CityGoldBonus(newX, newY, game.Plane)
                }

                cityInfoText = fonts.YellowFont.CreateWrappedText(float64(cancelBackground.Bounds().Dx() - 9), 1, text)
            }
        } else {
            cityInfoText.Clear()
            resources.Enabled = false
        }

        yield()
    }
}
