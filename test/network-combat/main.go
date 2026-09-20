package main

import (
    "io"
    "fmt"
    "log"
    "time"
    "image"
    "math"
    "flag"
    "net"
    "strconv"

    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/optional"
    "github.com/kazzmir/master-of-magic/lib/coroutine"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type noGlobalEnchantments struct {
}

func (*noGlobalEnchantments) HasEnchantment(enchantment data.Enchantment) bool {
    return false
}

func (*noGlobalEnchantments) HasRivalEnchantment(player *player.Player, enchantment data.Enchantment) bool {
    return false
}

type BasicCatchment struct {
}

func (basic *BasicCatchment) GetCatchmentArea(x int, y int) map[image.Point]maplib.FullTile {
    return map[image.Point]maplib.FullTile{}
}

func (basic *BasicCatchment) GetGoldBonus(x int, y int) int {
    return 0
}

func (basic *BasicCatchment) OnShore(x int, y int) bool {
    return false
}

func (basic *BasicCatchment) ByRiver(x int, y int) bool {
    return false
}

func (basic *BasicCatchment) TileDistance(x1 int, y1 int, x2 int, y2 int) int {
    dx := x1 - x2
    dy := y1 - y2
    return int(math.Sqrt(float64(dx * dx + dy * dy)))
}


type Engine struct {
    Model *combat.CombatModel
    Combat *combat.CombatScreen
    Coroutine *coroutine.Coroutine
}

func createArmyN(player *player.Player, unit units.Unit, count int) *combat.Army {
    army := combat.Army{
        Player: player,
    }

    for i := 0; i < count; i++ {
        made := units.MakeOverworldUnitFromUnit(unit, 1, 1, data.PlaneArcanus, player.Wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider())
        army.AddUnit(made)
    }

    return &army
}

func MakeScenario1(isServer bool, remote *combat.Remote) (*combat.CombatModel, *combat.CombatScreen, error) {
    cache := lbx.AutoCache()

    allSpells, err := spellbook.ReadSpellsFromCache(cache)
    if err != nil {
        return nil, nil, err
    }

    // remote player is always the defending player
    defendingPlayer := player.MakePlayer(setup.WizardCustom{
        Name: "Lair",
        Banner: data.BannerBrown,
    }, !isServer, 0, 0, nil, &noGlobalEnchantments{})

    // defendingArmy := createWarlockArmy(&defendingPlayer)
    // defendingArmy := createHighMenBowmanArmyN(defendingPlayer, 3)
    defendingArmy := createArmyN(defendingPlayer, units.Hydra, 1)

    defendingFortressCity := citylib.MakeCity("xyz", 10, 10, defendingPlayer.Wizard.Race, nil, &BasicCatchment{}, nil, defendingPlayer)
    defendingFortressCity.Buildings.Insert(buildinglib.BuildingFortress)
    defendingPlayer.AddCity(defendingFortressCity)

    defendingPlayer.CastingSkillPower = 1000
    defendingPlayer.Mana = 1000

    // server is attacker
    attackingPlayer := player.MakePlayer(setup.WizardCustom{
            Name: "Merlin",
            Banner: data.BannerGreen,
            Books: []data.WizardBook{
                data.WizardBook{
                    Magic: data.ChaosMagic,
                    Count: 8,
                },
            },
        }, isServer, 0, 0, nil, &noGlobalEnchantments{})

    fortressCity := citylib.MakeCity("xyz", 10, 10, attackingPlayer.Wizard.Race, nil, &BasicCatchment{}, nil, attackingPlayer)
    fortressCity.Buildings.Insert(buildinglib.BuildingFortress)
    attackingPlayer.AddCity(fortressCity)

    attackingPlayer.CastingSkillPower = 1000
    attackingPlayer.Mana = 1000

    attackingArmy := createArmyN(attackingPlayer, units.Warlocks, 3)

    model := combat.MakeCombatModel(allSpells, defendingArmy, attackingArmy, combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, data.MagicNone, 10, 25, make(chan combat.CombatEvent, 10), remote)
    combatScreen := combat.MakeCombatScreen(cache, defendingArmy, attackingArmy, optional.Of[combat.ArmyPlayer](attackingPlayer), combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, model)

    return model, combatScreen, nil
}

func MakeScenario2(isServer bool, remote *combat.Remote) (*combat.CombatModel, *combat.CombatScreen, error) {
    cache := lbx.AutoCache()

    allSpells, err := spellbook.ReadSpellsFromCache(cache)
    if err != nil {
        return nil, nil, err
    }

    // remote player is always the defending player
    defendingPlayer := player.MakePlayer(setup.WizardCustom{
        Name: "Lair",
        Banner: data.BannerBrown,
    }, !isServer, 0, 0, nil, &noGlobalEnchantments{})

    // defendingArmy := createWarlockArmy(&defendingPlayer)
    // defendingArmy := createHighMenBowmanArmyN(defendingPlayer, 3)
    defendingArmy := createArmyN(defendingPlayer, units.ArchAngel, 1)

    defendingFortressCity := citylib.MakeCity("xyz", 10, 10, defendingPlayer.Wizard.Race, nil, &BasicCatchment{}, nil, defendingPlayer)
    defendingFortressCity.Buildings.Insert(buildinglib.BuildingFortress)
    defendingPlayer.AddCity(defendingFortressCity)

    defendingPlayer.CastingSkillPower = 1000
    defendingPlayer.Mana = 1000

    // server is attacker
    attackingPlayer := player.MakePlayer(setup.WizardCustom{
            Name: "Merlin",
            Banner: data.BannerGreen,
            Books: []data.WizardBook{
                data.WizardBook{
                    Magic: data.ChaosMagic,
                    Count: 8,
                },
            },
        }, isServer, 0, 0, nil, &noGlobalEnchantments{})

    fortressCity := citylib.MakeCity("xyz", 10, 10, attackingPlayer.Wizard.Race, nil, &BasicCatchment{}, nil, attackingPlayer)
    fortressCity.Buildings.Insert(buildinglib.BuildingFortress)
    attackingPlayer.AddCity(fortressCity)

    attackingPlayer.CastingSkillPower = 1000
    attackingPlayer.Mana = 1000

    attackingArmy := createArmyN(attackingPlayer, units.Griffin, 1)

    model := combat.MakeCombatModel(allSpells, defendingArmy, attackingArmy, combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, data.MagicNone, 10, 25, make(chan combat.CombatEvent, 10), remote)
    combatScreen := combat.MakeCombatScreen(cache, defendingArmy, attackingArmy, optional.Of[combat.ArmyPlayer](attackingPlayer), combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, model)

    return model, combatScreen, nil
}

func MakeScenario(scenario int, isServer bool, remote *combat.Remote) (*combat.CombatModel, *combat.CombatScreen, error) {
    switch scenario {
        case 1: return MakeScenario1(isServer, remote)
        case 2: return MakeScenario2(isServer, remote)
        default: return MakeScenario1(isServer, remote)
    }
}

func NewEngine(isServer bool, peer net.Conn, scenario int) (*Engine, error) {
    model, combatScreen, err := MakeScenario(scenario, isServer, combat.MakeRemote(!isServer, !isServer, peer))

    if err != nil {
        return nil, err
    }

    run := func(yield coroutine.YieldFunc) error {
        for combatScreen.Update(yield) == combat.CombatStateRunning {
            yield()
        }

        return ebiten.Termination
    }

    return &Engine{
        Model: model,
        Combat: combatScreen,
        Coroutine: coroutine.MakeCoroutine(run),
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

    inputmanager.Update()

    engine.Coroutine.Run()

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    engine.Combat.Draw(screen)
    mouse.Mouse.Draw(screen)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (int, int) {
    return scale.Scale2(data.ScreenWidth, data.ScreenHeight)
}

func isPort(address string) bool {
    n, err := strconv.Atoi(address)
    return err == nil && n > 0 && n < 65536
}

func resolveAddress(address string) string {
    if isPort(address) {
        return net.JoinHostPort("localhost", address)
    }

    return address
}

// send a string to the other side to ensure we are connecting to another network combat client
func doHandshake(conn net.Conn, isClient bool) error {
    start := time.Now()
    defer func() {
        log.Printf("Handshake took %v", time.Since(start))
    }()

    magicString := "network-combat-handshake"

    success := true

    done := make(chan struct{})
    // read response asynchronously
    go func() {
        defer close(done)
        buffer := make([]byte, len(magicString))
        _, err := io.ReadFull(conn, buffer)
        if err != nil {
            success = false
            return
        }

        if string(buffer) != magicString {
            success = false
            return
        }
    }()

    _, err := io.WriteString(conn, magicString)
    if err != nil {
        success = false
    } else {
        conn.SetReadDeadline(time.Now().Add(5 * time.Second))
        select {
            case <-time.After(5 * time.Second):
                success = false
            case <-done:
        }
        conn.SetReadDeadline(time.Time{})
    }

    if success {
        return nil
    }

    return fmt.Errorf("handshake failed")
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    scenario := 0

    connectAddress := flag.String("connect", "", "Connect address")
    listenAddress := flag.String("listen", "", "Listen address")
    flag.Parse()

    for _, arg := range flag.Args() {
        v, err := strconv.Atoi(arg)
        if err == nil {
            scenario = v
        }
    }

    log.Printf("Starting combat screen, scenario %d", scenario)

    var peerConnection net.Conn
    isServer := false

    // connecting to a server
    if *connectAddress != "" {
        log.Printf("Connecting to server at %s", *connectAddress)

        address := resolveAddress(*connectAddress)

        connection, err := net.Dial("tcp", address)
        if err != nil {
            log.Printf("Error: unable to connect to server: %v", err)
            return
        }
        defer connection.Close()
        peerConnection = connection

        err = doHandshake(peerConnection, true)
        if err != nil {
            log.Printf("Error: handshake failed: %v", err)
            return
        }
    }

    if listenAddress != nil && *listenAddress != "" {
        address := resolveAddress(*listenAddress)
        server, err := net.Listen("tcp", address)
        if err != nil {
            log.Printf("Error: unable to listen on %s: %v", address, err)
            return
        }

        defer server.Close()

        log.Printf("Listening on %s, waiting..", address)
        clientConnection, err := server.Accept()
        if err != nil {
            log.Printf("Error: unable to accept connection: %v", err)
            return
        }
        defer clientConnection.Close()

        server.Close()

        err = doHandshake(clientConnection, false)
        if err != nil {
            log.Printf("Error: handshake failed: %v", err)
            return
        }

        peerConnection = clientConnection
        isServer = true
    }

    log.Printf("Initializing")

    monitorWidth, _ := ebiten.Monitor().Size()
    size := monitorWidth / 390
    ebiten.SetWindowSize(data.ScreenWidth * size, data.ScreenHeight * size)

    name := "client"
    if isServer {
        name = "server"
    }
    ebiten.SetWindowTitle(fmt.Sprintf("combat screen (%s)", name))
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
    ebiten.SetCursorMode(ebiten.CursorModeHidden)

    audio.Initialize()
    mouse.Initialize()

    engine, err := NewEngine(isServer, peerConnection, scenario)

    if err != nil {
        log.Printf("Error: unable to load engine: %v", err)
        return
    }

    log.Printf("Running")

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }

    log.Printf("Bye")
}
