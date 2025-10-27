package main

import (
    "log"
    "context"

    "github.com/kazzmir/master-of-magic/lib/fraction"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    // "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/cityview"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/set"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    Counter uint64
    UI *uilib.UI
    Quit context.Context
    Cache *lbx.LbxCache
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

func (provider *NoCityProvider) GetSpellByName(name string) spellbook.Spell {
    return spellbook.Spell{}
}

func (provider *NoCityProvider) GetAllGlobalEnchantments() map[data.BannerType]*set.Set[data.Enchantment] {
    enchantments := make(map[data.BannerType]*set.Set[data.Enchantment])
    return enchantments
}

func NewEngine() (*Engine, error) {
    cache := lbx.AutoCache()
    engine := &Engine{
        Counter: 0,
        Cache: cache,
    }

    var err error
    engine.UI, engine.Quit, err = engine.MakeUI()
    return engine, err
}

func (engine *Engine) MakeUI() (*uilib.UI, context.Context, error) {
    buildingInfo, _ := buildinglib.ReadBuildingInfo(engine.Cache)

    terrainLbx, err := engine.Cache.GetLbxFile("terrain.lbx")
    if err != nil {
        return nil, context.Background(), err
    }

    terrainData, err := terrain.ReadTerrainData(terrainLbx)
    if err != nil {
        return nil, context.Background(), err
    }

    gameMap := maplib.Map{
        Data: terrainData,
        Map: terrain.GenerateLandCellularAutomata(20, 20, terrainData, data.PlaneArcanus),
        TileCache: make(map[int]*ebiten.Image),
    }

    player := &playerlib.Player{
        Wizard: setup.WizardCustom{
            Name: "joe",
            Banner: data.BannerBlue,
        },
        TaxRate: fraction.Make(2, 1),
        GlobalEnchantments: set.MakeSet[data.Enchantment](),
    }

    city := citylib.MakeCity("Boston", 3, 8, data.RaceHighElf, buildingInfo, &gameMap, &NoCityProvider{}, player)
    city.Population = 12000
    city.Farmers = 4
    city.Workers = 2
    city.Production = 18
    city.ProducingBuilding = buildinglib.BuildingNone
    city.Buildings.Insert(buildinglib.BuildingGranary)
    city.Buildings.Insert(buildinglib.BuildingFarmersMarket)
    city.Buildings.Insert(buildinglib.BuildingMarketplace)
    city.Buildings.Insert(buildinglib.BuildingMinersGuild)
    city.Buildings.Insert(buildinglib.BuildingSawmill)
    city.Buildings.Insert(buildinglib.BuildingMechaniciansGuild)
    city.ProducingBuilding = buildinglib.BuildingHousing
    city.ResetCitizens()

    city.AddEnchantment(data.CityEnchantmentWallOfFire, data.BannerRed)

    return cityview.MakeEnchantmentView(engine.Cache, city, player, data.CityEnchantmentWallOfFire)

    /*
    allSpells, err := spellbook.ReadSpellsFromCache(engine.Cache)
    if err != nil {
        log.Printf("Unable to read spells: %v", err)
        allSpells = spellbook.Spells{}
    }

    spells := spellbook.Spells{}
    spells.AddSpell(allSpells.FindByName("War Bears"))
    spells.AddSpell(allSpells.FindByName("Guardian Spirit"))
    spells.AddSpell(allSpells.FindByName("Sprites"))
    spells.AddSpell(allSpells.FindByName("Magic Spirit"))
    spells.AddSpell(allSpells.FindByName("Great Wyrm"))
    spells.AddSpell(allSpells.FindByName("Gorgons"))
    spells.AddSpell(allSpells.FindByName("Arch Angel"))
    spells.AddSpell(allSpells.FindByName("Bless"))
    spells.AddSpell(allSpells.FindByName("Iron Skin"))

    spells = allSpells

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image) {
            ui.IterateElementsByLayer(func(element *uilib.UIElement) {
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })
        },
    }
    ui.SetElementsFromArray(nil)

    more := spellbook.MakeSpellBookCastUI(ui, engine.Cache, spells, 60, spellbook.Spell{}, 0, true, func (result spellbook.Spell, picked bool){
        if picked {
            log.Printf("Picked spell %v", result)
        }
    })
    ui.AddElements(more)

    return ui
    */
}

func (engine *Engine) Update() error {
    engine.Counter += 1
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    engine.UI.StandardUpdate()

    if engine.Quit.Err() != nil {
        return ebiten.Termination
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image){
    engine.UI.Draw(engine.UI, screen)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return scale.Scale2(data.ScreenWidth, data.ScreenHeight)
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(data.ScreenWidth * 2, data.ScreenHeight * 2)
    ebiten.SetWindowTitle("city enchantment")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    audio.Initialize()

    engine, err := NewEngine()

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
