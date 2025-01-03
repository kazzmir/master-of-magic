package game

import (
    "fmt"
    "log"
    "image"

    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/font"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/summon"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    // "github.com/kazzmir/master-of-magic/game/magic/data"

    "github.com/hajimehoshi/ebiten/v2"
)

func (game *Game) doCastSpell(yield coroutine.YieldFunc, player *playerlib.Player, spell spellbook.Spell) {
    switch spell.Name {
        case "Earth Lore":
            screenX, screenY, cancel := game.selectLocationForSpell(yield, spell)

            if cancel {
                return
            }

            tileX := game.CurrentMap().WrapX(game.cameraX + screenX / game.CurrentMap().TileWidth())
            tileY := game.cameraY + screenY / game.CurrentMap().TileHeight()
            game.CenterCamera(tileX, tileY)

            game.doCastEarthLore(yield, player)

            player.LiftFogSquare(tileX, tileY, 5, game.Plane)
        case "Create Artifact", "Enchant Item":
            showSummon := summon.MakeSummonArtifact(game.Cache, player.Wizard.Base)

            drawer := game.Drawer
            defer func(){
                game.Drawer = drawer
            }()

            game.Drawer = func(screen *ebiten.Image, game *Game){
                drawer(screen, game)
                showSummon.Draw(screen)
            }

            for showSummon.Update() != summon.SummonStateDone {
                if inputmanager.LeftClick() {
                    break
                }
                yield()
            }

            // FIXME: show vault with artifact

            select {
                case game.Events <- &GameEventVault{CreatedArtifact: player.CreateArtifact}:
                default:
            }

            player.CreateArtifact = nil
    }
}

/* return x,y and true/false, where true means cancelled, and false means something was selected */
// FIXME: this copies a lot of code from the surveyor, try to combine the two with shared functions/code
func (game *Game) selectLocationForSpell(yield coroutine.YieldFunc, spell spellbook.Spell) (int, int, bool) {
    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    var cities []*citylib.City
    var citiesMiniMap []maplib.MiniMapCity
    var stacks []*playerlib.UnitStack
    var fog [][]bool

    for i, player := range game.Players {
        for _, city := range player.Cities {
            if city.Plane == game.Plane {
                cities = append(cities, city)
                citiesMiniMap = append(citiesMiniMap, city)
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
        return 0, 0, true
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Error reading fonts: %v", err)
        return 0, 0, true
    }

    castingFont := makeSurveyorFont(fonts)
    // yellowFont := makeYellowFont(fonts)
    whiteFont := makeWhiteFont(fonts)

    overworld := Overworld{
        CameraX: game.cameraX,
        CameraY: game.cameraY,
        Counter: game.Counter,
        Map: game.CurrentMap(),
        Cities: cities,
        CitiesMiniMap: citiesMiniMap,
        Stacks: stacks,
        SelectedStack: nil,
        ImageCache: &game.ImageCache,
        Fog: fog,
        ShowAnimation: game.State == GameStateUnitMoving,
        FogBlack: game.GetFogImage(),
    }

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

            castingFont.PrintCenter(screen, 280, 81, 1, ebiten.ColorScale{}, "Casting")

            whiteFont.PrintWrapCenter(screen, 280, 120, float64(cancelBackground.Bounds().Dx() - 5), 1, ebiten.ColorScale{}, fmt.Sprintf("Select a space as the target for an %v spell.", spell.Name))
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
            newX := game.cameraX + x / game.CurrentMap().TileWidth()
            newY := game.cameraY + y / game.CurrentMap().TileHeight()
            newPoint := image.Pt(newX, newY)

            // right click should move the camera
            rightClick := inputmanager.RightClick()
            if rightClick {
                moveCamera = newPoint.Add(image.Pt(-5, -5))
                if moveCamera.Y < 0 {
                    moveCamera.Y = 0
                }

                if moveCamera.Y >= game.CurrentMap().Height() - 11 {
                    moveCamera.Y = game.CurrentMap().Height() - 11
                }
            }

            if inputmanager.LeftClick() {
                return x, y, false
            }
        }

        yield()
    }

    return 0, 0, true
}

func (game *Game) doCastEarthLore(yield coroutine.YieldFunc, player *playerlib.Player) {
    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    pics, _ := game.ImageCache.GetImages("specfx.lbx", 45)

    animation := util.MakeAnimation(pics, false)

    x := 120
    y := 90

    game.Drawer = func(screen *ebiten.Image, game *Game) {
        oldDrawer(screen, game)

        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(x - animation.Frame().Bounds().Dx() / 2), float64(y - animation.Frame().Bounds().Dy() / 2))
        screen.DrawImage(animation.Frame(), &options)
    }

    // FIXME: play earth lore sound
    // probably soundfx.lbx sound 28
    // newsound.fx 18

    sound, err := audio.LoadNewSound(game.Cache, 18)
    if err == nil {
        sound.Play()
    }

    quit := false
    for !quit {
        game.Counter += 1

        quit = false
        if game.Counter % 6 == 0 {
            quit = !animation.Next()
        }

        yield()
    }
}
