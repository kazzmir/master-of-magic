package engine

import (
    "flag"
    "log"
    "fmt"
    "time"
    "net"
    "math"
    "context"
    "sync"
    "image"
    "os"
    "os/signal"
    "syscall"

    "github.com/kazzmir/master-of-magic/game/network-combat/engine/crypto"
    netlib "github.com/kazzmir/master-of-magic/lib/net"
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

type Engine struct {
    Model *combat.CombatModel
    Combat *combat.CombatScreen
    Coroutine *coroutine.Coroutine
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

func sendArmy(remote *combat.Remote, team combat.Team, army *combat.Army) {
    var units []combat.RemoteUnit

    for _, unit := range army.GetUnits() {
        units = append(units, combat.RemoteUnit{
            UnitId: unit.Unit.GetRawUnit().GetId(),
        })
    }

    event := combat.RemoteArmyEvent{
        Team: team,
        Type: combat.RemoteArmyType,
        Units: units,
    }

    remote.SendEvent(&event)
}

func receiveArmy(remote *combat.Remote, player *player.Player) (*combat.Army, error) {
    select {
        case <-time.After(5 * time.Second):
            return nil, fmt.Errorf("timeout waiting for defending army from client")
        case event := <-remote.Events:
            switch event.GetType() {
                case combat.RemoteArmyType:
                    armyEvent := event.(*combat.RemoteArmyEvent)

                    log.Printf("Received army from remote: %v", armyEvent)

                    army := &combat.Army{
                        Player: player,
                    }

                    experienceInfo := player.MakeExperienceInfo()
                    enchantmentProvider := player.MakeUnitEnchantmentProvider()

                    for _, unitId := range armyEvent.Units {
                        made := units.MakeOverworldUnitFromUnit(units.GetUnitById(unitId.UnitId), 1, 1, data.PlaneArcanus, player.Wizard.Banner, experienceInfo, enchantmentProvider)
                        army.AddUnit(made)
                    }

                    return army, nil
            }
    }

    return nil, fmt.Errorf("did not receive an event from remote")
}

func makeGenericScenario(isServer bool, remote *combat.Remote, defendingUnit units.Unit, defendingCount int, attackingUnit units.Unit, attackingCount int) (*combat.CombatModel, *combat.CombatScreen, error) {
    cache := lbx.AutoCache()

    allSpells, err := spellbook.ReadSpellsFromCache(cache)
    if err != nil {
        return nil, nil, err
    }

    quit, cancel := context.WithCancel(context.Background())
    defer cancel()

    var waiter sync.WaitGroup
    waiter.Go(func(){ remote.RunReceiveLoop(quit) })

    // connecting player is always the defending player
    defendingPlayer := player.MakePlayer(setup.WizardCustom{
        Name: "Lair",
        Banner: data.BannerBrown,
    }, !isServer, 0, 0, nil, &noGlobalEnchantments{})

    defendingFortressCity := citylib.MakeCity("xyz", 10, 10, defendingPlayer.Wizard.Race, nil, &BasicCatchment{}, nil, defendingPlayer)
    defendingFortressCity.Buildings.Insert(buildinglib.BuildingFortress)
    // defendingFortressCity.Buildings.Insert(buildinglib.BuildingCityWalls)
    defendingFortressCity.AddEnchantment(data.CityEnchantmentWallOfFire, defendingPlayer.GetBanner())
    defendingPlayer.AddCity(defendingFortressCity)

    defendingPlayer.CastingSkillPower = 50000
    defendingPlayer.Mana = 1000

    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Wall of Fire"))
    defendingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Wall of Darkness"))

    var defendingArmy *combat.Army

    if !isServer {
        defendingArmy = createArmyN(defendingPlayer, defendingUnit, defendingCount)
        sendArmy(remote, combat.TeamDefender, defendingArmy)
    } else {
        defendingArmy, err = receiveArmy(remote, defendingPlayer)
        if err != nil {
            return nil, nil, err
        }
    }

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

    attackingPlayer.CastingSkillPower = 50000
    attackingPlayer.Mana = 1000

    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Call Chaos"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Disintegrate"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Flame Strike"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Holy Word"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Mass Healing"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Death Spell"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Earth to Mud"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Disrupt"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Phantom Warriors"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Disenchant Area"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("High Prayer"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Raise Dead"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Call Lightning"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Animate Dead"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Life Drain"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Wrack"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Terror"))
    attackingPlayer.KnownSpells.AddSpell(allSpells.FindByName("Magic Vortex"))

    var attackingArmy *combat.Army

    if isServer {
        attackingArmy = createArmyN(attackingPlayer, attackingUnit, attackingCount)
        sendArmy(remote, combat.TeamAttacker, attackingArmy)
    } else {
        attackingArmy, err = receiveArmy(remote, attackingPlayer)
        if err != nil {
            return nil, nil, err
        }
    }

    cancel()
    waiter.Wait()

    model := combat.MakeCombatModel(allSpells, defendingArmy, attackingArmy, combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{City: defendingFortressCity}, data.MagicNone, 10, 25, make(chan combat.CombatEvent, 10), remote)

    player := attackingPlayer
    if !isServer {
        player = defendingPlayer
    }

    combatScreen := combat.MakeCombatScreen(cache, defendingArmy, attackingArmy, optional.Of[combat.ArmyPlayer](player), combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, model)

    go remote.RunReceiveLoop(combatScreen.Quit)

    return model, combatScreen, nil
}

func MakeScenario1(isServer bool, remote *combat.Remote) (*combat.CombatModel, *combat.CombatScreen, error) {
    return makeGenericScenario(isServer, remote, units.DragonTurtle, 2, units.StagBeetle, 2)
}

func MakeScenario2(isServer bool, remote *combat.Remote) (*combat.CombatModel, *combat.CombatScreen, error) {
    return makeGenericScenario(isServer, remote, units.ArchAngel, 2, units.Griffin, 2)
}

func MakeScenario3(isServer bool, remote *combat.Remote) (*combat.CombatModel, *combat.CombatScreen, error) {
    return makeGenericScenario(isServer, remote, units.GnollSpearmen, 3, units.OrcSpearmen, 3)
}

func MakeScenario4(isServer bool, remote *combat.Remote) (*combat.CombatModel, *combat.CombatScreen, error) {
    return makeGenericScenario(isServer, remote, units.BeastmenPriest, 3, units.ArchAngel, 3)
}

func MakeScenario5(isServer bool, remote *combat.Remote) (*combat.CombatModel, *combat.CombatScreen, error) {
    return makeGenericScenario(isServer, remote, units.Cockatrices, 3, units.GreatDrake, 3)
}

func MakeScenario(scenario int, isServer bool, remote *combat.Remote) (*combat.CombatModel, *combat.CombatScreen, error) {
    switch scenario {
        case 1: return MakeScenario1(isServer, remote)
        case 2: return MakeScenario2(isServer, remote)
        case 3: return MakeScenario3(isServer, remote)
        case 4: return MakeScenario4(isServer, remote)
        case 5: return MakeScenario5(isServer, remote)
        default: return MakeScenario1(isServer, remote)
    }
}

func NewEngine(isServer bool, peer netlib.PacketConn, scenario int) (*Engine, error) {
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

func runGame(conn netlib.PacketConn, quit context.Context, isServer bool) error {
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

    engine, err := NewEngine(isServer, conn, 1)

    if err != nil {
        log.Printf("Error: unable to load engine: %v", err)
        return err
    }

    log.Printf("Running")

    return ebiten.RunGame(engine)
}

const MAX_PACKET_SIZE = 1024 * 1024 // 1 MB

func doRun(peer net.Conn, quit context.Context, isServer bool) error {
    defer peer.Close()

    go func() {
        <-quit.Done()
        log.Printf("Shutting down connection to %s", peer.RemoteAddr())
        peer.Close()
    }()

    packetConn := netlib.NewPacketNetConn(peer, MAX_PACKET_SIZE)

    didExchange := make(chan netlib.PacketConn)

    go func() {
        log.Printf("Starting key exchange with %s", peer.RemoteAddr())
        encryptedConn, err := crypto.NewEncryptedConn(packetConn)
        if err != nil {
            log.Printf("Failed to create encrypted connection: %v", err)
            return
        }
        didExchange <- encryptedConn
    }()

    var encryptedConn netlib.PacketConn

    select {
        case encryptedConn = <-didExchange:
            log.Printf("Key exchange completed with %s", peer.RemoteAddr())
        case <-quit.Done():
            return fmt.Errorf("Connection closed before key exchange completed")
        case <-time.After(5 * time.Second):
            return fmt.Errorf("Timeout waiting for key exchange to complete")
    }

    /*
    go encryptedConn.Send([]byte("Hello from this side!"))
    message, err := encryptedConn.Read()
    if err != nil {
        return err
    }
    
    log.Printf("Received message: %s", string(message))
    */

    return runGame(encryptedConn, quit, isServer)
}

func doListen(port int, quit context.Context) error {
    listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        return err
    }

    defer listen.Close()

    ok := make(chan struct{})
    defer close(ok)

    stop := false

    go func() {
        select {
            case <-quit.Done():
                stop = true
                log.Printf("Shutting down listener")
                listen.Close()
            case <-ok:
        }
    }()

    for !stop {
        peer, err := listen.Accept()
        if err != nil {
            if stop {
                break
            }

            log.Printf("Failed to accept connection: %v", err)
            continue
        }

        log.Printf("Accepted connection from %s", peer.RemoteAddr())

        listen.Close()
        return doRun(peer, quit, true)
    }

    return fmt.Errorf("Listener stopped")
}

func doConnect(peer string, quit context.Context) error {
    conn, err := net.Dial("tcp", peer)
    if err != nil {
        return err
    }

    return doRun(conn, quit, false)
}

func Launch() {
    log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

    port := flag.Int("port", 0, "Port to listen on")
    peer := flag.String("peer", "", "Peer address to connect to")
    flag.Parse()

    quit, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()

    var wait sync.WaitGroup

    if *port != 0 {
        log.Printf("Listen for connections on port %d", *port)
        err := doListen(*port, quit)
        if err != nil {
            log.Fatalf("Failed to listen on port %d: %v", *port, err)
        }
    } else if *peer != "" {
        // connect to peer
        doConnect(*peer, quit)
    } else {
        log.Fatal("Either -port or -peer must be specified")
    }

    done := make(chan struct{})
    go func() {
        wait.Wait()
        close(done)
    }()

    select {
        case <-done:
        case <-time.After(2 * time.Second):
            log.Println("Timeout waiting for goroutines to finish")
    }

    log.Printf("Shutting down")
}
