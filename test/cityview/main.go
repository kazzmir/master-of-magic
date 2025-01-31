package main

import (
    "log"
    // "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    "github.com/kazzmir/master-of-magic/game/magic/cityview"
    "github.com/kazzmir/master-of-magic/game/magic/camera"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/game"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/units"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    CityScreen *cityview.CityScreen
    ImageCache util.ImageCache
    Map *maplib.Map
}

type NoCityProvider struct {
}

func (provider *NoCityProvider) FindRoadConnectedCities(city *citylib.City) []*citylib.City {
    return nil
}

func NewEngine() (*Engine, error) {
    cache := lbx.AutoCache()

    player := playerlib.Player{
        Wizard: setup.WizardCustom{
            Name: "Billy",
            Banner: data.BannerRed,
        },
    }

    buildingInfo, _ := buildinglib.ReadBuildingInfo(cache)

    terrainLbx, err := cache.GetLbxFile("terrain.lbx")
    if err != nil {
        return nil, err
    }

    terrainData, err := terrain.ReadTerrainData(terrainLbx)
    if err != nil {
        return nil, err
    }


    gameMap := maplib.Map{
        Data: terrainData,
        Map: terrain.GenerateLandCellularAutomata(20, 20, terrainData, data.PlaneArcanus),
        TileCache: make(map[int]*ebiten.Image),
    }

    city := citylib.MakeCity("Boston", 3, 8, data.RaceHighElf, player.Wizard.Banner, fraction.Make(2, 1), buildingInfo, &gameMap, &NoCityProvider{})
    city.Population = 12000
    city.Farmers = 4
    city.Workers = 2
    city.Wall = false
    city.Production = 18
    city.ProducingBuilding = buildinglib.BuildingNone
    city.Banner = data.BannerBlue
    city.Buildings.Insert(buildinglib.BuildingGranary)
    city.Buildings.Insert(buildinglib.BuildingFarmersMarket)
    city.Buildings.Insert(buildinglib.BuildingMarketplace)
    city.Buildings.Insert(buildinglib.BuildingMinersGuild)
    city.Buildings.Insert(buildinglib.BuildingSawmill)
    city.Buildings.Insert(buildinglib.BuildingMechaniciansGuild)
    city.Buildings.Insert(buildinglib.BuildingBuildersHall)
    city.Buildings.Insert(buildinglib.BuildingCityWalls)
    city.ProducingBuilding = buildinglib.BuildingHousing
    // city.ProducingUnit = units.HighElfSpearmen
    city.ResetCitizens(nil)
        // ProducingUnit: units.UnitNone,

    /*
    city.AddEnchantment(data.CityEnchantmentWallOfFire, data.BannerRed)
    city.AddEnchantment(data.CityEnchantmentWallOfDarkness, data.BannerGreen)
    */

    var garrison []units.StackUnit
    for i := 0; i < 2; i++ {
        unit := units.MakeOverworldUnitFromUnit(units.HighElfSpearmen, city.X, city.Y, city.Plane, city.Banner, player.MakeExperienceInfo())
        player.AddUnit(unit)
        garrison = append(garrison, unit)
    }
    for i := 0; i < 4; i++ {
        unit := units.MakeOverworldUnitFromUnit(units.HighElfSwordsmen, city.X, city.Y, city.Plane, city.Banner, player.MakeExperienceInfo())
        player.AddUnit(unit)
        garrison = append(garrison, unit)
    }

    city.UpdateUnrest(garrison)

    city.AddBuilding(buildinglib.BuildingWizardsGuild)
    city.AddBuilding(buildinglib.BuildingSmithy)
    city.AddBuilding(buildinglib.BuildingSummoningCircle)
    city.AddBuilding(buildinglib.BuildingOracle)
    city.AddBuilding(buildinglib.BuildingFortress)
    city.AddBuilding(buildinglib.BuildingShrine)
    city.AddBuilding(buildinglib.BuildingTemple)

    cityScreen := cityview.MakeCityScreen(cache, city, &player, buildinglib.BuildingShrine)

    return &Engine{
        LbxCache: cache,
        CityScreen: cityScreen,
        ImageCache: util.MakeImageCache(cache),
        Map: &gameMap,
    }, nil
}

func (engine *Engine) Update() error {

    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    switch engine.CityScreen.Update() {
        case cityview.CityScreenStateRunning:
        case cityview.CityScreenStateDone:
            return ebiten.Termination
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    cameraX := engine.CityScreen.City.X - 2
    cameraY := engine.CityScreen.City.Y - 2

    engine.CityScreen.Draw(screen, func (where *ebiten.Image, geom ebiten.GeoM, counter uint64) {
        overworld := game.Overworld{
            Camera: camera.MakeCameraAt(cameraX, cameraY),
            Map: engine.Map,
            Cities: []*citylib.City{engine.CityScreen.City},
            Stacks: []*playerlib.UnitStack{},
            SelectedStack: nil,
            ImageCache: &engine.ImageCache,
            Counter: counter,
            Fog: nil,
            ShowAnimation: false,
            FogBlack: nil,
        }

        overworld.DrawOverworld(where, geom)
    }, engine.Map.TileWidth(), engine.Map.TileHeight())
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return data.ScreenWidth, data.ScreenHeight
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    monitorWidth, _ := ebiten.Monitor().Size()
    size := monitorWidth / 390

    ebiten.SetWindowSize(data.ScreenWidth / data.ScreenScale * size, data.ScreenHeight / data.ScreenScale * size)
    ebiten.SetWindowTitle("city view")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    audio.Initialize()

    engine, err := NewEngine()

    if err != nil {
        log.Printf("Error: unable to load engine: %v", err)
        return
    }

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
