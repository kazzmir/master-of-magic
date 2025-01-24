package game

import (
    "fmt"
    "log"
    "image"
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/font"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
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
    "github.com/kazzmir/master-of-magic/game/magic/camera"

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
    LocationTypeRaiseVolcano
)

func (game *Game) doCastSpell(yield coroutine.YieldFunc, player *playerlib.Player, spell spellbook.Spell) {
    switch spell.Name {
        case "Earth Lore":
            tileX, tileY, cancel := game.selectLocationForSpell(yield, spell, player, LocationTypeAny)

            if cancel {
                return
            }

            game.doCastEarthLore(yield, tileX, tileY, player)
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

            game.doMoveCamera(yield, tileX, tileY)
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
        case "Raise Volcano":
            tileX, tileY, cancel := game.selectLocationForSpell(yield, spell, player, LocationTypeRaiseVolcano)

            if cancel {
                return
            }

            game.doCastRaiseVolcano(yield, tileX, tileY)
        case "Summon Hero":
            var choices []*herolib.Hero
            for _, hero := range game.Heroes {
                if hero.Status == herolib.StatusAvailable && !hero.IsChampion() {
                    choices = append(choices, hero)
                }
            }

            if len(choices) > 0 {
                hero := choices[rand.N(len(choices))]

                summonEvent := GameEventSummonHero{
                    Wizard: player.Wizard.Base,
                    Champion: false,
                }

                select {
                    case game.Events <- &summonEvent:
                    default:
                }

                event := GameEventHireHero{
                    Hero: hero,
                    Player: player,
                    Cost: 0,
                }

                select {
                    case game.Events <- &event:
                    default:
                }
            }
        case "Enchant Road":
            tileX, tileY, cancel := game.selectLocationForSpell(yield, spell, player, LocationTypeAny)

            if cancel {
                return
            }

            game.doCastEnchantRoad(yield, tileX, tileY)
        default:
            log.Printf("Warning: casting unhandled spell %v", spell.Name)
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
        case LocationTypeAny, LocationTypeChangeTerrain, LocationTypeTransmute, LocationTypeRaiseVolcano:
            selectMessage = fmt.Sprintf("Select a space as the target for an %v spell.", spell.Name)
        case LocationTypeFriendlyCity:
            selectMessage = fmt.Sprintf("Select a friendly city to cast %v on.", spell.Name)
        default:
            selectMessage = fmt.Sprintf("unhandled location type %v", locationType)
    }

    ui := &uilib.UI{
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

            game.WhiteFont.PrintRight(screen, float64(276 * data.ScreenScale), float64(68 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("%v GP", game.Players[0].Gold))
            game.WhiteFont.PrintRight(screen, float64(313 * data.ScreenScale), float64(68 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("%v MP", game.Players[0].Mana))

            castingFont.PrintCenter(screen, float64(280 * data.ScreenScale), float64(81 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, "Casting")

            whiteFont.PrintWrapCenter(screen, float64(280 * data.ScreenScale), float64(120 * data.ScreenScale), float64(cancelBackground.Bounds().Dx() - 5 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, selectMessage)
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
    ui.AddElement(makeButton(7, 270, 4))

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
                        if tileY >= 0 && tileY < overworld.Map.Map.Rows() {
                            tileX = overworld.Map.WrapX(tileX)

                            if player.IsTileVisible(tileX, tileY, game.Plane) {
                                terrainType := overworld.Map.GetTile(tileX, tileY).Tile.TerrainType()
                                switch terrainType {
                                    case terrain.Desert, terrain.Forest, terrain.Hill,
                                         terrain.Swamp, terrain.Grass, terrain.Volcano,
                                         terrain.Mountain:
                                        return tileX, tileY, false
                                }
                            }
                        }
                    case LocationTypeTransmute:
                        if tileY >= 0 && tileY < overworld.Map.Map.Rows() {
                            tileX = overworld.Map.WrapX(tileX)

                            if player.IsTileVisible(tileX, tileY, game.Plane) {
                                bonusType := overworld.Map.GetBonusTile(tileX, tileY)
                                switch bonusType {
                                    case data.BonusCoal, data.BonusGem, data.BonusIronOre,
                                         data.BonusGoldOre, data.BonusSilverOre, data.BonusMithrilOre:
                                        return tileX, tileY, false
                                }
                            }
                        }
                    case LocationTypeRaiseVolcano:
                        if tileY >= 0 && tileY < overworld.Map.Map.Rows() {
                            tileX = overworld.Map.WrapX(tileX)

                            if player.IsTileVisible(tileX, tileY, game.Plane) {
                                terrainType := overworld.Map.GetTile(tileX, tileY).Tile.TerrainType()
                                switch terrainType {
                                    case terrain.Desert, terrain.Forest, terrain.Swamp, terrain.Grass, terrain.Tundra:
                                        return tileX, tileY, false
                                }
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

type UpdateTerrainFunction func (int, int, int)

func (game *Game) doCastOnTerrain(yield coroutine.YieldFunc, tileX int, tileY int, animationIndex int, newSound bool, soundIndex int, terrainFunction UpdateTerrainFunction) {
    game.Camera.Zoom = camera.ZoomDefault
    game.doMoveCamera(yield, tileX, tileY)

    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    pics, _ := game.ImageCache.GetImages("specfx.lbx", animationIndex)

    animation := util.MakeAnimation(pics, false)

    // FIXME: need some function in Game that returns the pixel coordinates for a given tile
    x := 130 * data.ScreenScale
    y := 99 * data.ScreenScale

    game.Drawer = func(screen *ebiten.Image, game *Game) {
        oldDrawer(screen, game)

        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(x - animation.Frame().Bounds().Dx() / 2), float64(y - animation.Frame().Bounds().Dy() / 2))
        screen.DrawImage(animation.Frame(), &options)
    }

    if newSound {
        sound, err := audio.LoadNewSound(game.Cache, soundIndex)
        if err == nil {
            sound.Play()
        }
    } else {
        sound, err := audio.LoadSound(game.Cache, soundIndex)
        if err == nil {
            sound.Play()
        }
    }

    quit := false
    for !quit {
        game.Counter += 1

        quit = false
        if game.Counter % 6 == 0 {
            terrainFunction(tileX, tileY, animation.CurrentFrame)
            quit = !animation.Next()
        }

        yield()
    }
}

func (game *Game) doCastEnchantRoad(yield coroutine.YieldFunc, tileX int, tileY int) {
    update := func (x int, y int, frame int) {}

    game.doCastOnTerrain(yield, tileX, tileY, 46, false, 86, update)

    useMap := game.CurrentMap()

    // all roads in a 5x5 square around the target tile should become enchanted
    for dx := -2; dx <= 2; dx++ {
        for dy := -2; dy <= 2; dy++ {
            cx := useMap.WrapX(tileX + dx)
            cy := tileY + dy
            if cy < 0 || cy >= useMap.Height() {
                continue
            }

            if useMap.ContainsRoad(cx, cy) {
                useMap.SetRoad(cx, cy, true)
            }
        }
    }
}

func (game *Game) doCastEarthLore(yield coroutine.YieldFunc, tileX int, tileY int, player *playerlib.Player) {
    update := func (x int, y int, frame int) {}

    game.doCastOnTerrain(yield, tileX, tileY, 45, true, 18, update)

    player.LiftFogSquare(tileX, tileY, 5, game.Plane)
}

func (game *Game) doCastChangeTerrain(yield coroutine.YieldFunc, tileX int, tileY int) {
    update := func (x int, y int, frame int) {
        if frame == 7 {
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
    }

    game.doCastOnTerrain(yield, tileX, tileY, 8, false, 28, update)
}


func (game *Game) doCastTransmute(yield coroutine.YieldFunc, tileX int, tileY int) {
    update := func (x int, y int, frame int) {
        if frame == 6 {
            mapObject := game.CurrentMap()
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

    game.doCastOnTerrain(yield, tileX, tileY, 0, false, 28, update)
}


func (game *Game) doCastRaiseVolcano(yield coroutine.YieldFunc, tileX int, tileY int) {
    update := func (x int, y int, frame int) {
        if frame == 8 {
            mapObject := game.CurrentMap()
            mapObject.Map.SetTerrainAt(x, y, terrain.Grass, mapObject.Data, mapObject.Plane)
            mapObject.SetBonus(tileX, tileY, data.BonusNone)
        }
    }

    game.doCastOnTerrain(yield, tileX, tileY, 11, false, 98, update)

    mapObject := game.CurrentMap()
    mapObject.Map.SetTerrainAt(tileX, tileY, terrain.Volcano, mapObject.Data, mapObject.Plane)

    // FIXME: Destruct buildings in towns
    // FIXME: Raise wizard's power
    // FIXME: Chance of reverting with chance of generating minerals
}