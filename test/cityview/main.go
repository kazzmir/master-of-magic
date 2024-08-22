package main

import (
    "log"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/set"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/game"
    "github.com/kazzmir/master-of-magic/game/magic/util"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    CityScreen *citylib.CityScreen
    ImageCache util.ImageCache
    Map game.Map
}

func NewEngine() (*Engine, error) {
    cache := lbx.AutoCache()

    city := citylib.City{
        Population: 6000,
        Farmers: 4,
        Workers: 2,
        Rebels: 1,
        Name: "Boston",
        Wall: false,
        Race: data.RaceHighElf,
        FoodProduction: 3,
        WorkProduction: 3,
        MoneyProduction: 4,
        MagicProduction: 3,
        Producing: citylib.BuildingHousing,
        X: 3,
        Y: 8,
        Buildings: set.MakeSet[citylib.Building](),
    }

    // city.AddBuilding(citylib.BuildingWizardsGuild)
    city.AddBuilding(citylib.BuildingSmithy)
    city.AddBuilding(citylib.BuildingSummoningCircle)
    city.AddBuilding(citylib.BuildingOracle)
    city.AddBuilding(citylib.BuildingWizardTower)

    cityScreen := citylib.MakeCityScreen(cache, &city)

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

    engine.CityScreen.Update()

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    cameraX := engine.CityScreen.City.X - 2
    cameraY := engine.CityScreen.City.Y - 2

    engine.CityScreen.Draw(screen, func (where *ebiten.Image, geom ebiten.GeoM, counter uint64) {
        where.Fill(color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff})

        engine.Map.Draw(cameraX, cameraY, counter / 10, where, geom)

        city := engine.CityScreen.City
        var cityPic *ebiten.Image
        var err error
        if city.Wall {
            cityPic, err = game.GetCityWallImage(city.GetSize(), &engine.ImageCache)
        } else {
            cityPic, err = game.GetCityNoWallImage(city.GetSize(), &engine.ImageCache)
        }

        if err == nil {
            var options ebiten.DrawImageOptions
            options.GeoM = geom

            options.GeoM.Translate(float64(city.X - cameraX) * float64(engine.Map.TileWidth()), float64(city.Y - cameraY) * float64(engine.Map.TileHeight()))
            where.DrawImage(cityPic, &options)
        }

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
