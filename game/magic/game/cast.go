package game

import (
    "fmt"
    "log"
    "image"

    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/font"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/cityview"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/summon"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"

    "github.com/hajimehoshi/ebiten/v2"
)

type LocationType int
const (
    LocationTypeAny LocationType = iota
    LocationTypeFriendlyCity
    LocationTypeEnemyCity
    LocationTypeFriendlyUnit
    LocationTypeEnemyUnit
    LocationTypeChangeTerrain
    LocationTypeTransmute
)

func (game *Game) doCastSpell(yield coroutine.YieldFunc, player *playerlib.Player, spell spellbook.Spell) {
    switch spell.Name {
        case "Earth Lore":
            tileX, tileY, cancel := game.selectLocationForSpell(yield, spell, player, LocationTypeAny)

            if cancel {
                return
            }

            game.Camera.Center(tileX, tileY)

            game.doCastEarthLore(yield, player)

            player.LiftFogSquare(tileX, tileY, 5, game.Plane)
        case "Create Artifact", "Enchant Item":
            showSummon := summon.MakeSummonArtifact(game.Cache, player.Wizard.Base)

            game.doSummon(yield, showSummon)

            select {
                case game.Events <- &GameEventVault{CreatedArtifact: player.CreateArtifact}:
                default:
            }

            player.CreateArtifact = nil
        case "Wall of Fire":
            tileX, tileY, cancel := game.selectLocationForSpell(yield, spell, player, LocationTypeFriendlyCity)

            if cancel {
                return
            }

            game.Camera.Center(tileX, tileY)
            chosenCity := player.FindCity(tileX, tileY, game.Plane)
            if chosenCity == nil {
                return
            }

            chosenCity.AddEnchantment(data.CityEnchantmentWallOfFire, player.GetBanner())

            yield()
            cityview.PlayEnchantmentSound(game.Cache)
            game.showCityEnchantment(yield, chosenCity, player, spell.Name)
        case "Change Terrain":
            tileX, tileY, cancel := game.selectLocationForSpell(yield, spell, player, LocationTypeChangeTerrain)

            if cancel {
                return
            }

            game.doCastChangeTerrain(yield, tileX, tileY)
        case "Transmute":
            tileX, tileY, cancel := game.selectLocationForSpell(yield, spell, player, LocationTypeTransmute)

            if cancel {
                return
            }

            game.doCastTransmute(yield, tileX, tileY)
    }
}

func (game *Game) showCityEnchantment(yield coroutine.YieldFunc, city *citylib.City, player *playerlib.Player, spellName string) {
    ui, quit, err := cityview.MakeEnchantmentView(game.Cache, city, player, spellName)
    if err != nil {
        log.Printf("Error making enchantment view: %v", err)
        return
    }

    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    game.Drawer = func(screen *ebiten.Image, game *Game){
        oldDrawer(screen, game)
        ui.Draw(ui, screen)
    }

    for quit.Err() == nil {
        game.Counter += 1
        ui.StandardUpdate()
        yield()
    }

    // absorb left click
    yield()
}

/* return x,y and true/false, where true means cancelled, and false means something was selected */
// FIXME: this copies a lot of code from the surveyor, try to combine the two with shared functions/code
func (game *Game) selectLocationForSpell(yield coroutine.YieldFunc, spell spellbook.Spell, player *playerlib.Player, locationType LocationType) (int, int, bool) {
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
        Camera: game.Camera,
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

    var selectMessage string

    switch locationType {
        case LocationTypeAny, LocationTypeChangeTerrain, LocationTypeTransmute: selectMessage = fmt.Sprintf("Select a space as the target for an %v spell.", spell.Name)
        case LocationTypeFriendlyCity: selectMessage = fmt.Sprintf("Select a friendly city to cast %v on.", spell.Name)
        default:
            selectMessage = fmt.Sprintf("unhandled location type %v", locationType)
    }

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

            whiteFont.PrintWrapCenter(screen, 280, 120, float64(cancelBackground.Bounds().Dx() - 5), 1, ebiten.ColorScale{}, selectMessage)
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
        overworld.Camera = game.Camera
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

    for !quit {
        if game.Camera.GetZoom() > 0.9 {
            overworld.Counter += 1
        }

        zoomed := game.doInputZoom(yield)
        _ = zoomed
        ui.StandardUpdate()

        x, y := inputmanager.MousePosition()

        // within the viewable area
        if game.InOverworldArea(x, y) {
            tileX, tileY := game.ScreenToTile(float64(x), float64(y))

            // right click should move the camera
            rightClick := inputmanager.RightClick()
            if rightClick /*|| zoomed */ {
                game.doMoveCamera(yield, tileX, tileY)
            }

            if inputmanager.LeftClick() {
                switch locationType {
                    case LocationTypeAny: return tileX, tileY, false
                    case LocationTypeFriendlyCity:
                        city := player.FindCity(tileX, tileY, game.Plane)
                        if city != nil {
                            return tileX, tileY, false
                        }
                    case LocationTypeChangeTerrain:
                        fog := player.GetFog(game.Plane)

                        if fog[tileX][tileY] {
                            terrainType := overworld.Map.GetTile(tileX, tileY).Tile.TerrainType()
                            switch terrainType {
                                case terrain.Desert, terrain.Forest, terrain.Hill,
                                     terrain.Swamp, terrain.Grass, terrain.Volcano,
                                     terrain.Mountain:
                                    return tileX, tileY, false
                            }
                        }
                    case LocationTypeTransmute:
                        fog := player.GetFog(game.Plane)

                        if fog[tileX][tileY] {
                            bonusType := overworld.Map.GetBonusTile(tileX, tileY)
                            switch bonusType {
                                case data.BonusCoal, data.BonusGem, data.BonusIronOre,
                                     data.BonusGoldOre, data.BonusSilverOre, data.BonusMithrilOre:
                                    return tileX, tileY, false
                            }
                        }

                    case LocationTypeEnemyCity:
                        // TODO

                    case LocationTypeFriendlyUnit:
                        // TODO

                    case LocationTypeEnemyUnit:
                        // TODO
                }
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


func (game *Game) doCastChangeTerrain(yield coroutine.YieldFunc, tileX int, tileY int) {
    game.Camera.Center(tileX, tileY)

    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    pics, _ := game.ImageCache.GetImages("specfx.lbx", 8)

    animation := util.MakeAnimation(pics, false)

    x := 130
    y := 100

    game.Drawer = func(screen *ebiten.Image, game *Game) {
        oldDrawer(screen, game)

        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(x - animation.Frame().Bounds().Dx() / 2), float64(y - animation.Frame().Bounds().Dy() / 2))
        screen.DrawImage(animation.Frame(), &options)
    }

    sound, err := audio.LoadNewSound(game.Cache, 18)
    if err == nil {
        sound.Play()
    }

    changeTerrain := func (x int, y int) {
        mapObject := game.CurrentMap()
        switch mapObject.GetTile(x, y).Tile.TerrainType() {
            case terrain.Desert, terrain.Forest, terrain.Hill, terrain.Swamp:
                mapObject.Map.SetTerrainAt(x, y, terrain.Grass, mapObject.Data, mapObject.Plane)
            case terrain.Grass:
                mapObject.Map.SetTerrainAt(x, y, terrain.Forest, mapObject.Data, mapObject.Plane)
            case terrain.Volcano:
                mapObject.Map.SetTerrainAt(x, y, terrain.Mountain, mapObject.Data, mapObject.Plane)
            case terrain.Mountain:
                mapObject.Map.SetTerrainAt(x, y, terrain.Hill, mapObject.Data, mapObject.Plane)
        }
    }

    quit := false
    for !quit {
        game.Counter += 1

        quit = false
        if game.Counter % 6 == 0 {
            quit = !animation.Next()
            if animation.CurrentFrame == 7 {
                changeTerrain(tileX, tileY)
            }
        }

        yield()
    }
}


func (game *Game) doCastTransmute(yield coroutine.YieldFunc, tileX int, tileY int) {
    game.Camera.Center(tileX, tileY)

    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    pics, _ := game.ImageCache.GetImages("specfx.lbx", 0)

    animation := util.MakeAnimation(pics, false)

    x := 130
    y := 100

    game.Drawer = func(screen *ebiten.Image, game *Game) {
        oldDrawer(screen, game)

        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(x - animation.Frame().Bounds().Dx() / 2), float64(y - animation.Frame().Bounds().Dy() / 2))
        screen.DrawImage(animation.Frame(), &options)
    }

    sound, err := audio.LoadNewSound(game.Cache, 18)
    if err == nil {
        sound.Play()
    }

    transmute := func (x int, y int) {
        mapObject := game.CurrentMap()
        if y >= 0 || y < mapObject.Map.Rows() {
            x = mapObject.WrapX(x)
            switch mapObject.GetBonusTile(x, y) {
                case data.BonusCoal: mapObject.SetBonus(x, y, data.BonusGem)
                case data.BonusGem: mapObject.SetBonus(x, y, data.BonusCoal)
                case data.BonusIronOre: mapObject.SetBonus(x, y, data.BonusGoldOre)
                case data.BonusGoldOre: mapObject.SetBonus(x, y, data.BonusIronOre)
                case data.BonusSilverOre: mapObject.SetBonus(x, y, data.BonusMithrilOre)
                case data.BonusMithrilOre: mapObject.SetBonus(x, y, data.BonusSilverOre)
            }
        }
    }

    quit := false
    for !quit {
        game.Counter += 1

        quit = false
        if game.Counter % 6 == 0 {
            quit = !animation.Next()
            if animation.CurrentFrame == 6 {
                transmute(tileX, tileY)
            }
        }

        yield()
    }
}
