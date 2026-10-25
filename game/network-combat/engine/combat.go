package engine

import (
	"context"
	"fmt"
	"image"
	"log"
	"math"
	"time"

	"github.com/kazzmir/master-of-magic/game/magic/audio"
	buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
	citylib "github.com/kazzmir/master-of-magic/game/magic/city"
	"github.com/kazzmir/master-of-magic/game/magic/combat"
	"github.com/kazzmir/master-of-magic/game/magic/data"
	"github.com/kazzmir/master-of-magic/game/magic/inputmanager"
	"github.com/kazzmir/master-of-magic/game/magic/maplib"
	"github.com/kazzmir/master-of-magic/game/magic/mouse"
	"github.com/kazzmir/master-of-magic/game/magic/player"
	"github.com/kazzmir/master-of-magic/game/magic/scale"
	"github.com/kazzmir/master-of-magic/game/magic/setup"
	"github.com/kazzmir/master-of-magic/game/magic/spellbook"
	"github.com/kazzmir/master-of-magic/game/magic/units"
	"github.com/kazzmir/master-of-magic/lib/coroutine"
	"github.com/kazzmir/master-of-magic/lib/lbx"
	netlib "github.com/kazzmir/master-of-magic/lib/net"
	"github.com/kazzmir/master-of-magic/lib/optional"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const networkCombatScenario = 6

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
	return int(math.Sqrt(float64(dx*dx + dy*dy)))
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
	Model     *combat.CombatModel
	Combat    *combat.CombatScreen
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

	if err := engine.Coroutine.Run(); err != nil {
		return ebiten.Termination
	}

	return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
	engine.Combat.Draw(screen)
	mouse.Mouse.Draw(screen)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (int, int) {
	return scale.Scale2(data.ScreenWidth, data.ScreenHeight)
}

func sendArmy(remote *combat.Remote, team combat.Team, army *combat.Army) error {
	var units []combat.RemoteUnit

	for _, unit := range army.GetUnits() {
		units = append(units, combat.RemoteUnit{
			UnitId: unit.Unit.GetRawUnit().GetId(),
		})
	}

	event := combat.RemoteArmyEvent{
		Team:  team,
		Type:  combat.RemoteArmyType,
		Units: units,
	}

	return remote.SendEvent(&event)
}

func receiveArmy(remote *combat.Remote, player *player.Player) (*combat.Army, error) {
	// Read exactly the army packet. A temporary background receive loop could
	// consume and discard the peer's first combat event during startup.
	remote.Peer.SetReadDeadline(time.Now().Add(10 * time.Second))
	defer remote.Peer.SetReadDeadline(time.Time{})
	event, err := remote.ReceiveEvent()
	if err != nil {
		return nil, fmt.Errorf("receiving remote army: %w", err)
	}
	armyEvent, ok := event.(*combat.RemoteArmyEvent)
	if !ok {
		return nil, fmt.Errorf("expected remote army, got %T", event)
	}
	army := &combat.Army{Player: player}
	experienceInfo := player.MakeExperienceInfo()
	enchantmentProvider := player.MakeUnitEnchantmentProvider()
	for _, unitId := range armyEvent.Units {
		made := units.MakeOverworldUnitFromUnit(units.GetUnitById(unitId.UnitId), 1, 1, data.PlaneArcanus, player.Wizard.Banner, experienceInfo, enchantmentProvider)
		army.AddUnit(made)
	}
	return army, nil
}

func makeGenericScenario(isServer bool, remote *combat.Remote, defendingUnit units.Unit, defendingCount int, attackingUnit units.Unit, attackingCount int) (*combat.CombatModel, *combat.CombatScreen, error) {
	cache := lbx.AutoCache()

	allSpells, err := spellbook.ReadSpellsFromCache(cache)
	if err != nil {
		return nil, nil, err
	}

	// connecting player is always the defending player
	defendingPlayer := player.MakePlayer(setup.WizardCustom{
		Name:   "Lair",
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
		if err := sendArmy(remote, combat.TeamDefender, defendingArmy); err != nil {
			return nil, nil, err
		}
	} else {
		defendingArmy, err = receiveArmy(remote, defendingPlayer)
		if err != nil {
			return nil, nil, err
		}
	}

	// server is attacker
	attackingPlayer := player.MakePlayer(setup.WizardCustom{
		Name:   "Merlin",
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
		if err := sendArmy(remote, combat.TeamAttacker, attackingArmy); err != nil {
			return nil, nil, err
		}
	} else {
		attackingArmy, err = receiveArmy(remote, attackingPlayer)
		if err != nil {
			return nil, nil, err
		}
	}

	model := combat.MakeCombatModel(allSpells, defendingArmy, attackingArmy, combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{City: defendingFortressCity}, data.MagicNone, 10, 25, make(chan combat.CombatEvent, 10), remote)

	player := attackingPlayer
	if !isServer {
		player = defendingPlayer
	}

	combatScreen := combat.MakeCombatScreen(cache, defendingArmy, attackingArmy, optional.Of[combat.ArmyPlayer](player), combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, model)

	if combatScreen == nil {
		return nil, nil, fmt.Errorf("unable to create combat screen (check game assets)")
	}
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
	case 1:
		return MakeScenario1(isServer, remote)
	case 2:
		return MakeScenario2(isServer, remote)
	case 3:
		return MakeScenario3(isServer, remote)
	case 4:
		return MakeScenario4(isServer, remote)
	case 5:
		return MakeScenario5(isServer, remote)
	case networkCombatScenario:
		return makeGenericScenario(isServer, remote, units.LizardSpearmen, 2, units.OrcSpearmen, 2)
	default:
		return MakeScenario1(isServer, remote)
	}
}

func NewEngine(isServer bool, peer netlib.PacketConn, scenario int) (*Engine, error) {
	model, combatScreen, err := MakeScenario(scenario, isServer, combat.MakeRemote(!isServer, !isServer, peer))

	if err != nil {
		return nil, err
	}

	return makeCombatEngine(model, combatScreen), nil
}

func makeCombatEngine(model *combat.CombatModel, combatScreen *combat.CombatScreen) *Engine {
	run := func(yield coroutine.YieldFunc) error {
		for combatScreen.Quit.Err() == nil && combatScreen.Update(yield) == combat.CombatStateRunning {
			if err := yield(); err != nil {
				return err
			}
		}

		return ebiten.Termination
	}

	return &Engine{
		Model:     model,
		Combat:    combatScreen,
		Coroutine: coroutine.MakeCoroutine(run),
	}
}

func runGame(conn netlib.PacketConn, quit context.Context, isServer bool) error {
	monitorWidth, _ := ebiten.Monitor().Size()
	size := monitorWidth / 390
	ebiten.SetWindowSize(data.ScreenWidth*size, data.ScreenHeight*size)

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
