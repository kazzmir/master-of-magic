package main

import (
    "log"
    // "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/set"
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

    city := citylib.City{
        Population: 6000,
        Farmers: 4,
        Workers: 2,
        Rebels: 1,
        Name: "Boston",
        Wall: false,
        Race: data.RaceHighElf,
        FoodProductionRate: 3,
        WorkProductionRate: 3,
        MoneyProductionRate: 4,
        MagicProductionRate: 3,
        Production: 18,
        ProducingBuilding: citylib.BuildingNone,
        // ProducingBuilding: citylib.BuildingBarracks,
        ProducingUnit: units.HighElfSpearmen,
        // ProducingUnit: units.UnitNone,
        X: 3,
        Y: 8,
        Buildings: set.MakeSet[citylib.Building](),
    }

    // city.AddBuilding(citylib.BuildingWizardsGuild)
    city.AddBuilding(citylib.BuildingSmithy)
    city.AddBuilding(citylib.BuildingSummoningCircle)
    city.AddBuilding(citylib.BuildingOracle)
    city.AddBuilding(citylib.BuildingFortress)

    cityScreen := cityview.MakeCityScreen(cache, &city, &player)

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

    ebiten.SetWindowSize(data.ScreenWidth * 5, data.ScreenHeight * 5)
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
