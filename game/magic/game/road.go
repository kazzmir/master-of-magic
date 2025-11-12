package game

import (
    "fmt"
    "image"
    "math"

    "github.com/kazzmir/master-of-magic/lib/coroutine"
    // "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"

    "github.com/hajimehoshi/ebiten/v2"
)

func (game *Game) ComputeRoadTime(path []image.Point, stack *playerlib.UnitStack) int {
    turns := float64(0)

    engineerCount := 0
    for _, unit := range stack.Units() {
        if unit.HasAbility(data.AbilityConstruction) {
            engineerCount += 1

            if unit.GetRace() == data.RaceDwarf {
                engineerCount += 1
            }

            if unit.HasEnchantment(data.UnitEnchantmentEndurance) {
                engineerCount += 1
            }
        }
    }

    mapUse := game.GetMap(stack.Plane())

    for _, point := range path {

        // no cost if road is already there
        // FIXME: take into account time to walk over road
        if mapUse.ContainsRoad(point.X, point.Y) {
            continue
        }

        work := game.ComputeRoadBuildEffort(point.X, point.Y, stack.Plane())
        turns += work.TotalWork / math.Pow(work.WorkPerEngineer, float64(engineerCount))
    }

    return int(turns)
}

type RoadMap struct {
    *maplib.Map

    CurrentPath pathfinding.Path
}

func (game *Game) ShowRoadBuilder(yield coroutine.YieldFunc, engineerStack *playerlib.UnitStack, player *playerlib.Player) {
    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    fonts := fontslib.MakeSurveyorFonts(game.Cache)

    var cityMap map[image.Point]*citylib.City

    roadMap := RoadMap{
        Map: game.CurrentMap(),
    }

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
            Map: &roadMap,
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

    cancelBackground, _ := game.ImageCache.GetImage("main.lbx", 47, 0)

    currentPath := []image.Point{image.Pt(engineerStack.X(), engineerStack.Y())}

    roadTurns := game.ComputeRoadTime(currentPath, engineerStack)

    roadMap.CurrentPath = pathfinding.Path{image.Pt(engineerStack.X(), engineerStack.Y())}

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

            fonts.SurveyorFont.PrintCenter(screen, float64(280), float64(81), scale.ScaleAmount, ebiten.ColorScale{}, "Road")
            fonts.SurveyorFont.PrintCenter(screen, float64(280), float64(81 + fonts.SurveyorFont.Height()), scale.ScaleAmount, ebiten.ColorScale{}, "Building")

            fonts.YellowFont.PrintWrap(screen, float64(249), float64(110), 68, font.FontOptions{DropShadow: true, Scale: scale.ScaleAmount}, fmt.Sprintf("It will take %d turns to complete the construction of this road.", roadTurns))
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
            scale.DrawScaled(screen, cancel[cancelIndex], &options)
        },
    })

    game.Drawer = func(screen *ebiten.Image, game *Game){
        overworld.Camera = game.Camera

        overworld.DrawOverworld(screen, ebiten.GeoM{})
        mini := screen.SubImage(game.GetMinimapRect()).(*ebiten.Image)
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

            leftClick := inputmanager.LeftClick()

            if leftClick && selectedPoint != newPoint {
                selectedPoint = newPoint

                newPath := game.FindPath(engineerStack.X(), engineerStack.Y(), newX, newY, player, engineerStack, player.GetFog(engineerStack.Plane()))

                if len(newPath) > 0 {

                    // if any point along the path is unexplored, we cannot build there
                    ok := true
                    for _, point := range newPath {
                        if !player.IsExplored(point.X, point.Y, engineerStack.Plane()) {
                            ok = false
                            break
                        }
                    }

                    if ok {
                        roadTurns = game.ComputeRoadTime(newPath, engineerStack)
                        roadMap.CurrentPath = newPath
                    }
                }
            }
        }

        yield()
    }
}
