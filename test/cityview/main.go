package main

import (
    "log"
    // "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/cityview"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/game"
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
    Map game.Map
}

func NewEngine() (*Engine, error) {
    cache := lbx.AutoCache()

    player := playerlib.Player{
        Wizard: setup.WizardCustom{
            Name: "Billy",
        },
    }

    city := citylib.MakeCity("Boston", 3, 8, data.RaceHighElf, fraction.Make(1, 1))
    city.Population = 6000
    city.Farmers = 4
    city.Workers = 2
    city.Wall = false
    city.MagicProductionRate = 3
    city.Production = 18
    city.ProducingBuilding = citylib.BuildingNone
    city.Banner = data.BannerBlue
    // ProducingBuilding: citylib.BuildingBarracks,
    city.ProducingUnit = units.HighElfSpearmen
        // ProducingUnit: units.UnitNone,

    city.AddGarrisonUnit(units.GreatDrake)
    city.AddGarrisonUnit(units.GreatDrake)
    for i := 0; i < 4; i++ {
        city.AddGarrisonUnit(units.FireElemental)
    }

    // city.AddBuilding(citylib.BuildingWizardsGuild)
    city.AddBuilding(citylib.BuildingSmithy)
    city.AddBuilding(citylib.BuildingSummoningCircle)
    city.AddBuilding(citylib.BuildingOracle)
    city.AddBuilding(citylib.BuildingFortress)

    cityScreen := cityview.MakeCityScreen(cache, city, &player)

    terrainLbx, err := cache.GetLbxFile("terrain.lbx")
    if err != nil {
        return nil, err
    }

    terrainData, err := terrain.ReadTerrainData(terrainLbx)
    if err != nil {
        return nil, err
    }

    return &Engine{
        LbxCache: cache,
        CityScreen: cityScreen,
        ImageCache: util.MakeImageCache(cache),
        Map: game.Map{
            Data: terrainData,
            Map: terrain.GenerateLandCellularAutomata(20, 20, terrainData),
            TileCache: make(map[int]*ebiten.Image),
        },
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
            CameraX: cameraX,
            CameraY: cameraY,
            Map: &engine.Map,
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
    })
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return data.ScreenWidth, data.ScreenHeight
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    monitorWidth, _ := ebiten.Monitor().Size()
    size := monitorWidth / 390

    ebiten.SetWindowSize(data.ScreenWidth * size, data.ScreenHeight * size)
    ebiten.SetWindowTitle("city view")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

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
