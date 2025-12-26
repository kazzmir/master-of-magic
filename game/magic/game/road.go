package game

import (
    "fmt"
    "image"
    "image/color"
    "math"
    _ "log"

    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/functional"
    // "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/camera"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

// return the number of engineers in the stack. only consider ones currently building a road if buildingRoad is true
func ComputeEngineerCount(stack *playerlib.UnitStack, buildingRoad bool) int {
    engineerCount := 0
    for _, unit := range stack.Units() {
        if unit.HasAbility(data.AbilityConstruction) && (!buildingRoad || unit.GetBusy() == units.BusyStatusBuildRoad) {
            engineerCount += 1

            if unit.GetRace() == data.RaceDwarf {
                engineerCount += 1
            }

            if unit.HasEnchantment(data.UnitEnchantmentEndurance) {
                engineerCount += 1
            }
        }
    }

    return engineerCount
}

func (game *Game) ComputeRoadTime(path []image.Point, stack *playerlib.UnitStack) int {
    turns := float64(0)

    engineerCount := ComputeEngineerCount(stack, false)

    mapUse := game.GetMap(stack.Plane())

    for _, point := range path {

        // no cost if road is already there
        // FIXME: take into account time to walk over road
        if mapUse.ContainsRoad(point.X, point.Y) {
            continue
        }

        work := game.Model.ComputeRoadBuildEffort(point.X, point.Y, stack.Plane())
        turns += work.TotalWork / math.Pow(work.WorkPerEngineer, float64(engineerCount))
    }

    return int(turns)
}

type RoadMap struct {
    *maplib.Map

    CurrentRoad pathfinding.Path
    // the value is an index into CurrentRoad of which path element this is
    Road map[image.Point]int
}

func (roadMap *RoadMap) UpdateRoad(path pathfinding.Path) {
    // log.Printf("New road: %v", path)
    roadMap.CurrentRoad = path
    roadMap.Road = make(map[image.Point]int)

    for i, point := range path {
        roadMap.Road[point] = i
    }
}

func (roadMap *RoadMap) DrawLayer2(camera_ camera.Camera, animationCounter uint64, imageCache *util.ImageCache, screen *ebiten.Image, geom ebiten.GeoM) {
    roadMap.DrawLayer2Internal(camera_, animationCounter, imageCache, screen, geom, roadMap)
}

func (roadMap *RoadMap) DrawTileLayer2(screen *ebiten.Image, imageCache *util.ImageCache, getOptions func() *ebiten.DrawImageOptions, animationCounter uint64, tileX int, tileY int){
    roadMap.Map.DrawTileLayer2(screen, imageCache, getOptions, animationCounter, tileX, tileY)

    index, ok := roadMap.Road[image.Pt(tileX, tileY)]
    if !ok {
        index, ok = roadMap.Road[image.Pt(tileX - roadMap.Width(), tileY)]
        if !ok {
            index, ok = roadMap.Road[image.Pt(tileX + roadMap.Width(), tileY)]
        }
    }

    if ok {
        options := getOptions()

        middleX, middleY := float64(roadMap.TileWidth()) / 2, float64(roadMap.TileHeight()) / 2

        drawSegment := func(cx int, cy int) {
            x, y := middleX, middleY

            xDistance := roadMap.XDistance(tileX, cx)

            if xDistance < 0 {
                x = 0
            } else if xDistance > 0 {
                x = float64(roadMap.TileWidth())
            }
            if cy < tileY {
                y = 0
            } else if cy > tileY {
                y = float64(roadMap.TileHeight())
            }

            x1, y1 := options.GeoM.Apply(middleX, middleY)
            x2, y2 := options.GeoM.Apply(x, y)

            vector.StrokeLine(screen, float32(scale.Scale(x1)), float32(scale.Scale(y1)), float32(scale.Scale(x2)), float32(scale.Scale(y2)), 1, color.RGBA{R: 255, A: 255}, true)
        }

        if index > 0 {
            drawSegment(roadMap.CurrentRoad[index - 1].X, roadMap.CurrentRoad[index - 1].Y)
        }
        if index < len(roadMap.CurrentRoad) - 1 {
            drawSegment(roadMap.CurrentRoad[index + 1].X, roadMap.CurrentRoad[index + 1].Y)
        }

        // x, y := options.GeoM.Apply(float64(roadMap.TileWidth() / 2), float64(roadMap.TileHeight() / 2))

        mx, my := options.GeoM.Apply(middleX, middleY)
        vector.FillCircle(screen, float32(scale.Scale(mx)), float32(scale.Scale(my)), 2, color.RGBA{R: 255, A: 255}, true)
    }
}

func (game *Game) FindRoadPath(oldX int, oldY int, newX int, newY int, player *playerlib.Player, stack playerlib.PathStack, fog data.FogMap) pathfinding.Path {
    useMap := game.GetMap(stack.Plane())

    if newY < 0 || newY >= useMap.Height() {
        return nil
    }

    normalized := func (a image.Point) image.Point {
        return image.Pt(useMap.WrapX(a.X), a.Y)
    }

    // check equality of two points taking wrapping into account
    tileEqual := func (a image.Point, b image.Point) bool {
        return normalized(a) == normalized(b)
    }

    // cache locations of enemies
    enemyStacks := make(map[image.Point]struct{})
    enemyCities := make(map[image.Point]struct{})

    for _, enemy := range game.Model.Players {
        if enemy != player {
            for _, enemyStack := range enemy.Stacks {
                enemyStacks[image.Pt(enemyStack.X(), enemyStack.Y())] = struct{}{}
            }
            for _, enemyCity := range enemy.Cities {
                enemyCities[image.Pt(enemyCity.X, enemyCity.Y)] = struct{}{}
            }
        }
    }

    // cache the containsEnemy result
    // true if the given coordinates contain an enemy unit or city
    containsEnemy := functional.Memoize2(func (x int, y int) bool {
        _, ok := enemyStacks[image.Pt(x, y)]
        if ok {
            return true
        }
        _, ok = enemyCities[image.Pt(x, y)]
        if ok {
            return true
        }

        return false
    })

    tileCost := func (x1 int, y1 int, x2 int, y2 int) float64 {
        x1 = useMap.WrapX(x1)
        x2 = useMap.WrapX(x2)

        if x1 < 0 || x1 >= useMap.Width() || y1 < 0 || y1 >= useMap.Height() {
            return pathfinding.Infinity
        }

        if x2 < 0 || x2 >= useMap.Width() || y2 < 0 || y2 >= useMap.Height() {
            return pathfinding.Infinity
        }

        if fog[x2][y2] == data.FogTypeUnexplored {
            return pathfinding.Infinity
        }

        tile := useMap.GetTile(x2, y2)
        if tile.Tile.IsWater() {
            return pathfinding.Infinity
        }

        // FIXME: it might be more optimal to put the infinity cases into the neighbors function instead

        // avoid encounters
        encounter := useMap.GetEncounter(x2, y2)
        if encounter != nil {
            if !tileEqual(image.Pt(x2, y2), image.Pt(newX, newY)) {
                return pathfinding.Infinity
            }
        }

        // avoid enemy units/cities
        if !tileEqual(image.Pt(x2, y2), image.Pt(newX, newY)) && containsEnemy(x2, y2) {
            return pathfinding.Infinity
        }

        if tile.HasRoad() {
            if stack.Plane() == data.PlaneMyrror {
                // should non-corporeal engineers take more than 0 time to cross a road?
                return 0
            }
            return 0.5
        }

        if x1 != x2 && y1 != y2 {
            // make diagonals more expensive
            return 1.1
        }

        return 1
    }

    neighbors := func (x int, y int) []image.Point {
        out := make([]image.Point, 0, 8)

        // cardinals first, followed by diagonals
        // left
        out = append(out, image.Pt(x - 1, y))

        // up
        if y > 0 {
            out = append(out, image.Pt(x, y - 1))
        }

        // right
        out = append(out, image.Pt(x + 1, y))

        // down
        if y < useMap.Height() - 1 {
            out = append(out, image.Pt(x, y + 1))
        }

        // up left
        if y > 0 {
            out = append(out, image.Pt(x - 1, y - 1))
        }

        // down left
        if y < useMap.Height() - 1 {
            out = append(out, image.Pt(x - 1, y + 1))
        }

        // up right
        if y > 0 {
            out = append(out, image.Pt(x + 1, y - 1))
        }

        // down right
        if y < useMap.Height() - 1 {
            out = append(out, image.Pt(x + 1, y + 1))
        }

        return out
    }

    path, ok := pathfinding.FindPath(image.Pt(oldX, oldY), image.Pt(newX, newY), 10000, tileCost, neighbors, tileEqual)
    if ok {
        return path
    }

    return nil
}

// returns the path of road that the given engineers should build
func (game *Game) ShowRoadBuilder(yield coroutine.YieldFunc, engineerStack *playerlib.UnitStack, player *playerlib.Player) pathfinding.Path {
    fonts := fontslib.MakeSurveyorFonts(game.Cache)

    var cityMap map[image.Point]*citylib.City

    roadMap := RoadMap{
        Map: game.Model.CurrentMap(),
    }

    makeOverworld := func () Overworld {
        cityMap = make(map[image.Point]*citylib.City)

        var cities []*citylib.City
        var stacks []*playerlib.UnitStack
        var fog data.FogMap

        for i, player := range game.Model.Players {
            for _, city := range player.Cities {
                if city.Plane == game.Model.Plane {
                    cities = append(cities, city)
                    cityMap[image.Pt(city.X, city.Y)] = city
                }
            }

            for _, stack := range player.Stacks {
                if stack.Plane() == game.Model.Plane {
                    stacks = append(stacks, stack)
                }
            }

            if i == 0 {
                fog = player.GetFog(game.Model.Plane)
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

    currentPath := []image.Point{image.Pt(engineerStack.X(), engineerStack.Y())}

    roadTurns := game.ComputeRoadTime(currentPath, engineerStack)

    roadMap.UpdateRoad(pathfinding.Path{image.Pt(engineerStack.X(), engineerStack.Y())})

    ui := &uilib.UI{
        Cache: game.Cache,
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            mainHud, _ := game.ImageCache.GetImage("main.lbx", 0, 0)
            scale.DrawScaled(screen, mainHud, &options)

            landImage, _ := game.ImageCache.GetImage("main.lbx", 45, 0)
            options.GeoM.Translate(float64(240), float64(77))
            scale.DrawScaled(screen, landImage, &options)

            button, _ := game.ImageCache.GetImage("main.lbx", 48, 0)
            options.GeoM.Reset()
            options.GeoM.Translate(float64(240), float64(173))
            scale.DrawScaled(screen, button, &options)

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
    success := false

    // cancel button at bottom
    cancel, _ := game.ImageCache.GetImages("main.lbx", 41)
    cancelIndex := 0
    cancelRect := util.ImageRect(280, 181, cancel[0])
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

    okImages, _ := game.ImageCache.GetImages("main.lbx", 46)
    okIndex := 0
    okRect := util.ImageRect(246, 181, okImages[0])
    ui.AddElement(&uilib.UIElement{
        Rect: okRect,
        LeftClick: func(element *uilib.UIElement){
            okIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            okIndex = 0
            success = true
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(okRect.Min.X), float64(okRect.Min.Y))
            scale.DrawScaled(screen, okImages[okIndex], &options)
        },
    })

    game.PushDrawer(func(screen *ebiten.Image){
        overworld.Camera = game.Camera

        overworld.DrawOverworld(screen, ebiten.GeoM{})
        mini := screen.SubImage(game.GetMinimapRect()).(*ebiten.Image)
        overworld.DrawMinimap(mini)

        ui.Draw(ui, screen)
    })
    defer game.PopDrawer()

    for !quit && !success {
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

                newPath := game.FindRoadPath(engineerStack.X(), engineerStack.Y(), newX, newY, player, engineerStack, player.GetFog(engineerStack.Plane()))

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
                        roadMap.UpdateRoad(newPath)
                    }
                }
            }
        }

        yield()
    }

    if success {
        return roadMap.CurrentRoad
    }

    return nil
}
