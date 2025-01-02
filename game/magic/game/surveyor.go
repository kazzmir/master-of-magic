package game

import (
    "log"
    "fmt"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/coroutine"
    // "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"

    "github.com/hajimehoshi/ebiten/v2"
)

func makeSurveyorFont(fonts []*font.LbxFont) *font.Font {
    white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
    palette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        white, white, white,
        white, white, white,
    }

    return font.MakeOptimizedFontWithPalette(fonts[4], palette)
}

func makeYellowFont(fonts []*font.LbxFont) *font.Font {
    yellow := util.RotateHue(color.RGBA{R: 255, G: 255, B: 0, A: 255}, -0.15)

    palette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        yellow, yellow, yellow,
        yellow, yellow, yellow,
    }

    return font.MakeOptimizedFontWithPalette(fonts[1], palette)
}

func makeWhiteFont(fonts []*font.LbxFont) *font.Font {
    white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
    palette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        white, white, white,
        white, white, white,
    }

    return font.MakeOptimizedFontWithPalette(fonts[1], palette)
}

func (game *Game) doSurveyor(yield coroutine.YieldFunc) {
    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    var cities []*citylib.City
    var stacks []*playerlib.UnitStack
    var fog [][]bool

    for i, player := range game.Players {
        for _, city := range player.Cities {
            if city.Plane == game.Plane {
                cities = append(cities, city)
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

    fontLbx, err := game.Cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Error reading fonts: %v", err)
        return
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Error reading fonts: %v", err)
        return
    }

    surveyorFont := makeSurveyorFont(fonts)
    yellowFont := makeYellowFont(fonts)
    whiteFont := makeWhiteFont(fonts)

    overworld := Overworld{
        CameraX: game.cameraX,
        CameraY: game.cameraY,
        Counter: game.Counter,
        Map: game.CurrentMap(),
        Cities: cities,
        Stacks: stacks,
        SelectedStack: nil,
        ImageCache: &game.ImageCache,
        Fog: fog,
        ShowAnimation: game.State == GameStateUnitMoving,
        FogBlack: game.GetFogImage(),
        Zoom: game.GetZoom(),
    }

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
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            mainHud, _ := game.ImageCache.GetImage("main.lbx", 0, 0)
            screen.DrawImage(mainHud, &options)

            landImage, _ := game.ImageCache.GetImage("main.lbx", 57, 0)
            options.GeoM.Translate(240, 77)
            screen.DrawImage(landImage, &options)

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

            if selectedPoint.X >= 0 && selectedPoint.X < game.CurrentMap().Width() && selectedPoint.Y >= 0 && selectedPoint.Y < game.CurrentMap().Height() {
                if fog[selectedPoint.X][selectedPoint.Y] {
                    tile := game.CurrentMap().GetTile(selectedPoint.X, selectedPoint.Y)
                    y := float64(93)
                    yellowFont.PrintCenter(screen, 280, y, 1, ebiten.ColorScale{}, tile.Tile.Name())
                    y += float64(yellowFont.Height())

                    foodBonus := tile.Tile.FoodBonus()
                    if !foodBonus.IsZero() {
                        whiteFont.PrintCenter(screen, 280, y, 1, ebiten.ColorScale{}, fmt.Sprintf("%v food", foodBonus.NormalString()))
                        y += float64(whiteFont.Height())
                    }

                    productionBonus := tile.Tile.ProductionBonus()
                    if productionBonus != 0 {
                        whiteFont.PrintCenter(screen, 280, y, 1, ebiten.ColorScale{}, fmt.Sprintf("+%v%% production", productionBonus))
                        y += float64(whiteFont.Height())
                    }

                    y += float64(whiteFont.Height())

                    showBonus := func (name string, bonus string) {
                        yellowFont.PrintCenter(screen, 280, y, 1, ebiten.ColorScale{}, name)
                        y += float64(yellowFont.Height())
                        whiteFont.PrintWrapCenter(screen, 280, y, float64(cancelBackground.Bounds().Dx() - 5), 1, ebiten.ColorScale{}, bonus)
                    }

                    bonus := tile.GetBonus()
                    switch bonus {
                        case data.BonusNone: // nothing
                        case data.BonusGoldOre: showBonus("Gold Ore", fmt.Sprintf("+%v gold", bonus.GoldBonus()))
                        case data.BonusSilverOre: showBonus("Silver Ore", fmt.Sprintf("+%v gold", bonus.GoldBonus()))
                        case data.BonusWildGame: showBonus("Wild Game", fmt.Sprintf("+%v food", bonus.FoodBonus()))
                        case data.BonusNightshade: showBonus("Nightshade", "")
                        case data.BonusIronOre: showBonus("Iron Ore", fmt.Sprintf("Reduces normal unit cost by %v%%", bonus.UnitReductionBonus()))
                        case data.BonusCoal: showBonus("Coal", fmt.Sprintf("Reduces normal unit cost by %v%%", bonus.UnitReductionBonus()))
                        case data.BonusMithrilOre: showBonus("Mithril Ore", fmt.Sprintf("+%v power", bonus.PowerBonus()))
                        case data.BonusAdamantiumOre: showBonus("Adamantium Ore", fmt.Sprintf("+%v power", bonus.PowerBonus()))
                        case data.BonusGem: showBonus("Gem", fmt.Sprintf("+%v gold", bonus.GoldBonus()))
                        case data.BonusQuorkCrystal: showBonus("Quork Crystal", fmt.Sprintf("+%v power", bonus.PowerBonus()))
                        case data.BonusCrysxCrystal: showBonus("Crysx Crystal", fmt.Sprintf("+%v power", bonus.PowerBonus()))
                    }

                    // FIXME: show lair/node/tower

                    y = 160 - cityInfoText.TotalHeight

                    if resources.Enabled {
                        y = 170 - float64(whiteFont.Height()) * 3 - cityInfoText.TotalHeight
                    }

                    yellowFont.RenderWrapped(screen, 245, y, cityInfoText, ebiten.ColorScale{}, false)
                    y += cityInfoText.TotalHeight

                    if resources.Enabled {
                        whiteFont.Print(screen, 245, y, 1, ebiten.ColorScale{}, "Maximum Pop")
                        whiteFont.PrintRight(screen, 308, y, 1, ebiten.ColorScale{}, fmt.Sprintf("%v", resources.MaximumPopulation))
                        y += float64(whiteFont.Height())
                        whiteFont.Print(screen, 245, y, 1, ebiten.ColorScale{}, "Prod Bonus")
                        whiteFont.PrintRight(screen, 314, y, 1, ebiten.ColorScale{}, fmt.Sprintf("+%v%%", resources.ProductionBonus))
                        y += float64(whiteFont.Height())
                        whiteFont.Print(screen, 245, y, 1, ebiten.ColorScale{}, "Gold Bonus")
                        whiteFont.PrintRight(screen, 314, y, 1, ebiten.ColorScale{}, fmt.Sprintf("+%v%%", resources.GoldBonus))
                    }
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
        if game.GetZoom() >= 0.9 {
            overworld.Counter += 1
        }
        zoomed := game.doInputZoom()

        overworld.Zoom = game.GetZoom()

        ui.StandardUpdate()

        x, y := inputmanager.MousePosition()

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
            realX, realY := game.RealToTile(float64(x), float64(y))

            newX := game.cameraX + realX
            newY := game.cameraY + realY
            newPoint := image.Pt(newX, newY)

            // right click should move the camera
            rightClick := inputmanager.RightClick()
            if rightClick || zoomed {
                moveCamera = selectedPoint.Add(image.Pt(-5, -5))
                if moveCamera.X < 0 {
                    moveCamera.X = 0
                }
                if moveCamera.Y < 0 {
                    moveCamera.Y = 0
                }
                if moveCamera.Y >= game.CurrentMap().Height() - 11 {
                    moveCamera.Y = game.CurrentMap().Height() - 11
                }
            }

            if selectedPoint != newPoint {
                selectedPoint = newPoint
                resources.Enabled = false

                text := ""

                tile := game.CurrentMap().GetTile(newX, newY)

                if !tile.Tile.IsLand() {
                    text = "Cannot build cities on water."
                } else if game.NearCity(newPoint, 3) {
                    text = "Cities cannot be built less than 3 squares from any other city."
                } else {
                    text = "City Resources"
                    resources.Enabled = true
                    // FIXME: compute proper values for these
                    resources.MaximumPopulation = game.ComputeMaximumPopulation(newX, newY, game.Plane)
                    resources.ProductionBonus = game.CityProductionBonus(newX, newY, game.Plane)
                    resources.GoldBonus = game.CityGoldBonus(newX, newY, game.Plane)
                }

                cityInfoText = yellowFont.CreateWrappedText(float64(cancelBackground.Bounds().Dx()) - 9, 1, text)
            }
        } else {
            cityInfoText.Clear()
            resources.Enabled = false
        }

        yield()
    }
}
