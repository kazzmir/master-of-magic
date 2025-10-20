package main

import (
    "log"
    "math/rand/v2"
    // "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/set"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    "github.com/kazzmir/master-of-magic/game/magic/cityview"
    "github.com/kazzmir/master-of-magic/game/magic/camera"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
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
    Player *playerlib.Player
}

type NoCityProvider struct {
}

func (provider *NoCityProvider) FindRoadConnectedCities(city *citylib.City) []*citylib.City {
    return nil
}

func (provider *NoCityProvider) GoodMoonActive() bool {
    return false
}

func (provider *NoCityProvider) BadMoonActive() bool {
    return false
}

func (provider *NoCityProvider) PopulationBoomActive(city *citylib.City) bool {
    return false
}

func (provider *NoCityProvider) PlagueActive(city *citylib.City) bool {
    return false
}

func (provider *NoCityProvider) GetAllGlobalEnchantments() map[data.BannerType]*set.Set[data.Enchantment] {
    enchantments := make(map[data.BannerType]*set.Set[data.Enchantment])
    return enchantments
}

func NewEngine() (*Engine, error) {
    cache := lbx.AutoCache()

    player := playerlib.Player{
        Wizard: setup.WizardCustom{
            Name: "Billy",
            Banner: data.BannerRed,
            Books: []data.WizardBook{
                {Magic: data.ChaosMagic, Count: 11},
            },
            Retorts: []data.Retort{
                data.RetortInfernalPower,
            },
        },
        TaxRate: fraction.Make(3, 1),
        // TaxRate: fraction.Zero(),
        GlobalEnchantments: set.MakeSet[data.Enchantment](),
    }

    player.Gold = 500

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

    city := citylib.MakeCity("Boston", rand.N(20), rand.N(13) + 4, data.RaceHighElf, buildingInfo, &gameMap, &NoCityProvider{}, &player)
    city.Population = 24000
    city.Farmers = 4
    city.Workers = 2
    city.Production = 18
    city.ProducingBuilding = buildinglib.BuildingNone

    city.AddBuilding(buildinglib.BuildingFortress)
    city.AddBuilding(buildinglib.BuildingGranary)
    city.AddBuilding(buildinglib.BuildingFarmersMarket)
    // city.AddBuilding(buildinglib.BuildingMarketplace)
    city.AddBuilding(buildinglib.BuildingMinersGuild)
    city.AddBuilding(buildinglib.BuildingSawmill)
    // city.AddBuilding(buildinglib.BuildingMechaniciansGuild)
    city.AddBuilding(buildinglib.BuildingBuildersHall)
    city.AddBuilding(buildinglib.BuildingCityWalls)
    city.AddBuilding(buildinglib.BuildingWizardsGuild)
    city.AddBuilding(buildinglib.BuildingSmithy)
    city.AddBuilding(buildinglib.BuildingSummoningCircle)
    city.AddBuilding(buildinglib.BuildingOracle)
    city.AddBuilding(buildinglib.BuildingShrine)
    // city.AddBuilding(buildinglib.BuildingTemple)
    // city.AddBuilding(buildinglib.BuildingParthenon)
    city.AddBuilding(buildinglib.BuildingCathedral)

    // for _, building := range buildinglib.Buildings() {
    //     city.AddBuilding(building)
    // }

    city.ProducingBuilding = buildinglib.BuildingHousing
    // city.ProducingUnit = units.HighElfSpearmen
    city.ResetCitizens()
        // ProducingUnit: units.UnitNone,

    city.AddEnchantment(data.CityEnchantmentWallOfFire, data.BannerRed)
    // city.AddEnchantment(data.CityEnchantmentWallOfDarkness, data.BannerRed)
    city.AddEnchantment(data.CityEnchantmentNaturesEye, data.BannerRed)
    city.AddEnchantment(data.CityEnchantmentProsperity, data.BannerRed)
    city.AddEnchantment(data.CityEnchantmentInspirations, data.BannerRed)
    city.AddEnchantment(data.CityEnchantmentConsecration, data.BannerRed)
    city.AddEnchantment(data.CityEnchantmentAstralGate, data.BannerRed)
    city.AddEnchantment(data.CityEnchantmentAltarOfBattle, data.BannerRed)
    city.AddEnchantment(data.CityEnchantmentStreamOfLife, data.BannerRed)
    city.AddEnchantment(data.CityEnchantmentEarthGate, data.BannerRed)
    city.AddEnchantment(data.CityEnchantmentDarkRituals, data.BannerRed)
    city.AddEnchantment(data.CityEnchantmentEvilPresence, data.BannerGreen)

    city.AddEnchantment(data.CityEnchantmentLifeWard, data.BannerRed)
    city.AddEnchantment(data.CityEnchantmentDeathWard, data.BannerRed)
    city.AddEnchantment(data.CityEnchantmentChaosWard, data.BannerRed)
    city.AddEnchantment(data.CityEnchantmentNatureWard, data.BannerRed)
    city.AddEnchantment(data.CityEnchantmentSorceryWard, data.BannerRed)

    for i := 0; i < 2; i++ {
        unit := units.MakeOverworldUnitFromUnit(units.HighElfSpearmen, city.X, city.Y, city.Plane, city.GetBanner(), player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider())
        unit.AddEnchantment(data.UnitEnchantmentGiantStrength)
        player.AddUnit(unit)
    }
    for i := 0; i < 4; i++ {
        unit := units.MakeOverworldUnitFromUnit(units.HighElfSwordsmen, city.X, city.Y, city.Plane, city.GetBanner(), player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider())
        player.AddUnit(unit)
    }

    city.UpdateUnrest()

    cityScreen := cityview.MakeCityScreen(cache, city, &player, buildinglib.BuildingWizardsGuild)

    return &Engine{
        LbxCache: cache,
        CityScreen: cityScreen,
        ImageCache: util.MakeImageCache(cache),
        Map: &gameMap,
        Player: &player,
    }, nil
}

func (engine *Engine) togglePlane() {
    if engine.CityScreen.City.Plane == data.PlaneArcanus {
        engine.CityScreen.City.Plane = data.PlaneMyrror
    } else {
        engine.CityScreen.City.Plane = data.PlaneArcanus
    }
}

func (engine *Engine) toggleEnchantment(enchantment data.CityEnchantment) {
    if engine.CityScreen.City.HasEnchantment(enchantment) {
        engine.CityScreen.City.CancelEnchantment(enchantment, data.BannerBlue)
    } else {
        engine.CityScreen.City.AddEnchantment(enchantment, data.BannerBlue)
    }
    engine.CityScreen.City.UpdateUnrest()
    engine.CityScreen.UI = engine.CityScreen.MakeUI(buildinglib.BuildingNone)
}

func (engine *Engine) Update() error {

    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyEscape, ebiten.KeyCapsLock: return ebiten.Termination
            case ebiten.KeyP: engine.togglePlane()
            case ebiten.Key1: engine.toggleEnchantment(data.CityEnchantmentFlyingFortress)
            case ebiten.Key2: engine.toggleEnchantment(data.CityEnchantmentFamine)
            case ebiten.Key3: engine.toggleEnchantment(data.CityEnchantmentCursedLands)
            case ebiten.Key4: engine.toggleEnchantment(data.CityEnchantmentGaiasBlessing)
            case ebiten.Key5: engine.toggleEnchantment(data.CityEnchantmentChaosRift)
            case ebiten.Key6: engine.toggleEnchantment(data.CityEnchantmentHeavenlyLight)
            case ebiten.Key7: engine.toggleEnchantment(data.CityEnchantmentCloudOfShadow)
            case ebiten.Key8: engine.toggleEnchantment(data.CityEnchantmentPestilence)
            case ebiten.Key9: engine.toggleEnchantment(data.CityEnchantmentEvilPresence)
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
    return scale.Scale2(data.ScreenWidth, data.ScreenHeight)
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    monitorWidth, _ := ebiten.Monitor().Size()
    size := monitorWidth / 390
    ebiten.SetWindowSize(data.ScreenWidth * size, data.ScreenHeight * size)
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
