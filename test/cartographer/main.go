package main

import (
    "log"
    "strconv"
    "os"
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/cartographer"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/setup"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    Coroutine *coroutine.Coroutine
    DrawScene func (*ebiten.Image)
}

type cityProvider struct{}
func (c *cityProvider) ContainsCity(x int, y int, plane data.Plane) bool {
    return false
}

func NewEngine(scenario int) (*Engine, error) {
    cache := lbx.AutoCache()

    player1 := playerlib.Player{
        Wizard: setup.WizardCustom{
            Banner: data.BannerBlue,
            Name: "bob",
        },
    }

    player2 := playerlib.Player{
        Wizard: setup.WizardCustom{
            Banner: data.BannerPurple,
            Name: "Kali",
        },
    }

    terrainLbx, err := cache.GetLbxFile("terrain.lbx")
    if err != nil {
        return nil, err
    }

    terrainData, err := terrain.ReadTerrainData(terrainLbx)
    if err != nil {
        return nil, err
    }

    makeFog := func(map_ *maplib.Map) data.FogMap {
        fog := make(data.FogMap, map_.Width())
        for x := range map_.Width() {
            fog[x] = make([]data.FogType, map_.Height())
        }
        return fog
    }

    arcanusMap := maplib.MakeMap(terrainData, 1, data.MagicSettingNormal, data.DifficultyAverage, data.PlaneArcanus, &cityProvider{}, nil)
    arcanusFog := makeFog(arcanusMap)

    /*
    for x := range arcanusMap.Width() {
        for y := range arcanusMap.Height() {
            arcanusMap.Map.Terrain[x][y] = terrain.TileOcean.Index(data.PlaneArcanus)
            // arcanusMap.Map.Terrain[x][y] = terrain.IndexBugGrass
        }
    }
    */

    for x := range len(arcanusFog) {
        for y := range len(arcanusFog[x]) {
            arcanusFog[x][y] = data.FogTypeVisible
        }
    }

    myrrorMap := maplib.MakeMap(terrainData, 1, data.MagicSettingNormal, data.DifficultyAverage, data.PlaneMyrror, &cityProvider{}, nil)
    myrrorFog := makeFog(myrrorMap)

    for x := range len(myrrorFog) {
        for y := range len(myrrorFog[x]) {
            myrrorFog[x][y] = data.FogTypeVisible
        }
    }

    var cities []*citylib.City

    city1 := citylib.MakeCity("city1", 10, 10, data.RaceBarbarian, nil, nil, nil, &player1)
    player1.AddCity(city1)

    city2 := citylib.MakeCity("city2", 30, 40, data.RaceKlackon, nil, nil, nil, &player2)
    player2.AddCity(city2)

    cities = append(cities, city1, city2)

    makeStack := func(player *playerlib.Player) *playerlib.UnitStack {
        x := rand.N(50)
        y := rand.N(50)

        unit := units.MakeOverworldUnitFromUnit(units.LizardSpearmen, x, y, data.PlaneArcanus, player.GetBanner(), player.MakeExperienceInfo(), &units.NoEnchantments{})

        return playerlib.MakeUnitStackFromUnits([]units.StackUnit{unit})
    }

    var stacks []*playerlib.UnitStack
    stacks = append(stacks, makeStack(&player1))
    stacks = append(stacks, makeStack(&player2))

    logic, draw := cartographer.MakeCartographer(cache, cities, stacks, arcanusMap, arcanusFog, myrrorMap, myrrorFog)

    return &Engine{
        LbxCache: cache,
        DrawScene: draw,
        Coroutine: coroutine.MakeCoroutine(logic),
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

    if engine.Coroutine.Run() != nil {
        return ebiten.Termination
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    engine.DrawScene(screen)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return scale.Scale2(data.ScreenWidth, data.ScreenHeight)
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    monitorWidth, _ := ebiten.Monitor().Size()
    size := monitorWidth / 390

    ebiten.SetWindowSize(data.ScreenWidth * size, data.ScreenHeight * size)
    ebiten.SetWindowTitle("cartographer")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    scenario := 1

    if len(os.Args) >= 2 {
        x, err := strconv.Atoi(os.Args[1])
        if err != nil {
            log.Fatalf("Error with scenario: %v", err)
        }

        scenario = x
    }

    engine, err := NewEngine(scenario)

    if err != nil {
        log.Printf("Error: unable to load engine: %v", err)
        return
    }

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
