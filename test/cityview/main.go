package main

import (
    "log"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/set"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/data"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    CityScreen *citylib.CityScreen
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

    return &Engine{
        LbxCache: cache,
        CityScreen: cityScreen,
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
    engine.CityScreen.Draw(screen, func (where *ebiten.Image, options ebiten.DrawImageOptions) {
        where.Fill(color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff})
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
