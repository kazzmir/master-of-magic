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
    "github.com/kazzmir/master-of-magic/game/magic/maplib"

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
            screen.DrawImage(mainHud, &options)

            landImage, _ := game.ImageCache.GetImage("main.lbx", 57, 0)
            options.GeoM.Translate(float64(240 * data.ScreenScale), float64(77 * data.ScreenScale))
            screen.DrawImage(landImage, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(240 * data.ScreenScale), float64(174 * data.ScreenScale))
            screen.DrawImage(cancelBackground, &options)

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })

            player := game.Players[0]

            game.WhiteFont.PrintRight(screen, float64(276 * data.ScreenScale), float64(68 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("%v GP", player.Gold))
            game.WhiteFont.PrintRight(screen, float64(313 * data.ScreenScale), float64(68 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("%v MP", player.Mana))

            surveyorFont.PrintCenter(screen, float64(280 * data.ScreenScale), float64(81 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, "Surveyor")

            if selectedPoint.X >= 0 && selectedPoint.X < game.CurrentMap().Width() && selectedPoint.Y >= 0 && selectedPoint.Y < game.CurrentMap().Height() {
                if overworld.Fog[selectedPoint.X][selectedPoint.Y] != data.FogTypeUnexplored {
                    mapObject := game.CurrentMap()
                    tile := mapObject.GetTile(selectedPoint.X, selectedPoint.Y)
                    node := mapObject.GetMagicNode(selectedPoint.X, selectedPoint.Y)

                    y := float64(93 * data.ScreenScale)

                    // Terrain
                    name := tile.Name(mapObject)
                    if node != nil {
                        switch node.Kind {
                            case maplib.MagicNodeNature: name = "Forest"
                            case maplib.MagicNodeSorcery: name = "Grasslands"
                            case maplib.MagicNodeChaos: name = "Mountain"
                        }
                    }
                    yellowFont.PrintCenter(screen, float64(280 * data.ScreenScale), y, float64(data.ScreenScale), ebiten.ColorScale{}, name)
                    y += float64(yellowFont.Height() * data.ScreenScale)

                    // Terrain bonuses
                    if tile.Corrupted() {
                        whiteFont.PrintCenter(screen, float64(280 * data.ScreenScale), y, float64(data.ScreenScale), ebiten.ColorScale{}, "Corruption")
                        y += float64(whiteFont.Height() * data.ScreenScale)
                    }

                    foodBonus := tile.FoodBonus()
                    if !foodBonus.IsZero() {
                        whiteFont.PrintCenter(screen, float64(280 * data.ScreenScale), y, float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("%v food", foodBonus.NormalString()))
                        y += float64(whiteFont.Height() * data.ScreenScale)
                    }

                    productionBonus := tile.ProductionBonus()
                    if productionBonus != 0 {
                        whiteFont.PrintCenter(screen, float64(280 * data.ScreenScale), y, float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("+%v%% production", productionBonus))
                        y += float64(whiteFont.Height() * data.ScreenScale)
                    }

                    goldBonus := tile.GoldBonus(mapObject)
                    if goldBonus != 0 {
                        whiteFont.PrintCenter(screen, float64(280 * data.ScreenScale), y, float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("+%v%% gold", goldBonus))
                        y += float64(whiteFont.Height() * data.ScreenScale)
                    }

                    y += float64(whiteFont.Height() * data.ScreenScale)

                    // Bonuses
                    bonus := tile.GetBonus()
                    if bonus != data.BonusNone {
                        yellowFont.PrintCenter(screen, float64(280 * data.ScreenScale), y, float64(data.ScreenScale), ebiten.ColorScale{}, bonus.String())
                        y += float64(yellowFont.Height() * data.ScreenScale)

                        food := bonus.FoodBonus()
                        if food != 0 {
                            whiteFont.PrintWrapCenter(screen, float64(280 * data.ScreenScale), y, float64(cancelBackground.Bounds().Dx() - 5 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("+%v food", food))
                            y += float64(whiteFont.Height() * data.ScreenScale)
                        }

                        gold := bonus.GoldBonus()
                        if gold != 0 {
                            whiteFont.PrintWrapCenter(screen, float64(280 * data.ScreenScale), y, float64(cancelBackground.Bounds().Dx() - 5 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("+%v gold", gold))
                            y += float64(whiteFont.Height() * data.ScreenScale)
                        }

                        power := bonus.PowerBonus()
                        if power != 0 {
                            whiteFont.PrintWrapCenter(screen, float64(280 * data.ScreenScale), y, float64(cancelBackground.Bounds().Dx() - 5 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("+%v power", power))
                            y += float64(whiteFont.Height() * data.ScreenScale)
                        }

                        reduction := bonus.UnitReductionBonus()
                        if reduction != 0 {
                            whiteFont.PrintWrapCenter(screen, float64(280 * data.ScreenScale), y, float64(cancelBackground.Bounds().Dx() - 5 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("Reduces normal unit cost by %v%%", reduction))
                            y += float64(whiteFont.Height() * data.ScreenScale)
                        }
                    }

                    // Nodes
                    if node != nil {
                        yellowFont.PrintWrapCenter(screen, float64(280 * data.ScreenScale), y, float64(cancelBackground.Bounds().Dx() - 5 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, node.Kind.Name())
                        y += float64(yellowFont.Height() * data.ScreenScale)

                        if node.Warped {
                            whiteFont.PrintCenter(screen, float64(280 * data.ScreenScale), y, float64(data.ScreenScale), ebiten.ColorScale{}, "Warped")
                            y += float64(whiteFont.Height() * data.ScreenScale)
                        } else if node.MeldingWizard != nil && node.GuardianSpiritMeld {
                            whiteFont.PrintCenter(screen, float64(280 * data.ScreenScale), y, float64(data.ScreenScale), ebiten.ColorScale{}, "Guardian Spirit")
                            y += float64(whiteFont.Height() * data.ScreenScale)
                        } else if node.MeldingWizard != nil {
                            whiteFont.PrintCenter(screen, float64(280 * data.ScreenScale), y, float64(data.ScreenScale), ebiten.ColorScale{}, "Magic Spirit")
                            y += float64(whiteFont.Height() * data.ScreenScale)
                        }
                    }

                    // Lairs
                    encounter := mapObject.GetEncounter(selectedPoint.X, selectedPoint.Y)
                    if encounter != nil && encounter.Type != maplib.EncounterTypeChaosNode && encounter.Type != maplib.EncounterTypeNatureNode && encounter.Type != maplib.EncounterTypeSorceryNode {
                        yellowFont.PrintCenter(screen, float64(280 * data.ScreenScale), y, float64(data.ScreenScale), ebiten.ColorScale{}, encounter.Type.Name())
                        y += float64(yellowFont.Height() * data.ScreenScale)
                    }

                    // Enemies
                    if encounter != nil {
                        text := "Unexplored"
                        if encounter.ExploredBy.Contains(player) {
                            text = "Empty"
                            if len(encounter.Units) > 0 {
                                text = encounter.Units[0].Name
                            }
                        }
                        whiteFont.PrintCenter(screen, float64(280 * data.ScreenScale), y, float64(data.ScreenScale), ebiten.ColorScale{}, text)
                        y += float64(whiteFont.Height() * data.ScreenScale)
                    }

                    // FIXME: how should this behave for different fog types?
                    if cityMap[selectedPoint] != nil {
                        city := cityMap[selectedPoint]
                        yellowFont.PrintWrapCenter(screen, float64(280 * data.ScreenScale), y, float64(cancelBackground.Bounds().Dx() - 5 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, city.String())
                    }

                    y = float64(170 * data.ScreenScale) - cityInfoText.TotalHeight

                    if resources.Enabled {
                        y = float64(170 * data.ScreenScale) - float64(whiteFont.Height() * data.ScreenScale) * 3 - cityInfoText.TotalHeight
                    }

                    yellowFont.RenderWrapped(screen, float64(245 * data.ScreenScale), y, cityInfoText, ebiten.ColorScale{}, false)
                    y += cityInfoText.TotalHeight

                    if resources.Enabled {
                        whiteFont.Print(screen, float64(245 * data.ScreenScale), y, float64(data.ScreenScale), ebiten.ColorScale{}, "Maximum Pop")
                        whiteFont.PrintRight(screen, float64(308 * data.ScreenScale), y, float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("%v", resources.MaximumPopulation))
                        y += float64(whiteFont.Height() * data.ScreenScale)
                        whiteFont.Print(screen, float64(245 * data.ScreenScale), y, float64(data.ScreenScale), ebiten.ColorScale{}, "Prod Bonus")
                        whiteFont.PrintRight(screen, float64(314 * data.ScreenScale), y, float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("+%v%%", resources.ProductionBonus))
                        y += float64(whiteFont.Height() * data.ScreenScale)
                        whiteFont.Print(screen, float64(245 * data.ScreenScale), y, float64(data.ScreenScale), ebiten.ColorScale{}, "Gold Bonus")
                        whiteFont.PrintRight(screen, float64(314 * data.ScreenScale), y, float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("+%v%%", resources.GoldBonus))
                    }
                }
            }
        },
    }

    ui.SetElementsFromArray(nil)

    makeButton := func(lbxIndex int, x int, y int) *uilib.UIElement {
        button, _ := game.ImageCache.GetImage("main.lbx", lbxIndex, 0)
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(x * data.ScreenScale), float64(y * data.ScreenScale))
        return &uilib.UIElement{
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
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
    ui.AddElement((func () *uilib.UIElement {
        buttons, _ := game.ImageCache.GetImages("main.lbx", 7)
        x := 270 * data.ScreenScale
        y := 4 * data.ScreenScale

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
                    screen.DrawImage(buttons[1], &options)
                } else {
                    screen.DrawImage(buttons[0], &options)
                }
            },
        }
    })())

    quit := false

    // cancel button at bottom
    cancel, _ := game.ImageCache.GetImages("main.lbx", 41)
    cancelIndex := 0
    cancelRect := util.ImageRect(263 * data.ScreenScale, 182 * data.ScreenScale, cancel[0])
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
        overworld.Camera = game.Camera

        overworld.DrawOverworld(screen, ebiten.GeoM{})

        var miniGeom ebiten.GeoM
        miniGeom.Translate(float64(250 * data.ScreenScale), float64(20 * data.ScreenScale))
        mx, my := miniGeom.Apply(0, 0)
        miniWidth := 60 * data.ScreenScale
        miniHeight := 31 * data.ScreenScale
        mini := screen.SubImage(image.Rect(int(mx), int(my), int(mx) + miniWidth, int(my) + miniHeight)).(*ebiten.Image)
        overworld.DrawMinimap(mini)

        ui.Draw(ui, screen)
    }

    for !quit {
        if game.Camera.GetZoom() >= 0.9 {
            overworld.Counter += 1
        }
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
                } else if cityMap[newPoint] == nil && game.NearCity(newPoint, 3) {
                    text = "Cities cannot be built less than 3 squares from any other city."
                } else {
                    text = "City Resources"
                    resources.Enabled = true
                    // FIXME: compute proper values for these
                    resources.MaximumPopulation = game.ComputeMaximumPopulation(newX, newY, game.Plane)
                    resources.ProductionBonus = game.CityProductionBonus(newX, newY, game.Plane)
                    resources.GoldBonus = game.CityGoldBonus(newX, newY, game.Plane)
                }

                cityInfoText = yellowFont.CreateWrappedText(float64(cancelBackground.Bounds().Dx() - 9 * data.ScreenScale), float64(data.ScreenScale), text)
            }
        } else {
            cityInfoText.Clear()
            resources.Enabled = false
        }

        yield()
    }
}
