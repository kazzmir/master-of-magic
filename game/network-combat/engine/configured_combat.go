package engine

import (
	"encoding/json/v2"
	"fmt"
	"time"

	buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
	citylib "github.com/kazzmir/master-of-magic/game/magic/city"
	"github.com/kazzmir/master-of-magic/game/magic/combat"
	"github.com/kazzmir/master-of-magic/game/magic/data"
	"github.com/kazzmir/master-of-magic/game/magic/player"
	"github.com/kazzmir/master-of-magic/game/magic/setup"
	"github.com/kazzmir/master-of-magic/game/magic/spellbook"
	"github.com/kazzmir/master-of-magic/game/magic/units"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
	"github.com/kazzmir/master-of-magic/lib/lbx"
	netlib "github.com/kazzmir/master-of-magic/lib/net"
	"github.com/kazzmir/master-of-magic/lib/optional"
)

func exchangeCombatSetup(peer netlib.PacketConn, localID uint64, state protocol.GameState) (protocol.CombatSetup, protocol.CombatSetup, error) {
	var local, other protocol.CombatSetup
	if len(state.Players) != 2 || !state.Started {
		return local, other, fmt.Errorf("game must have two ready players")
	}
	var own, opponent *protocol.GamePlayer
	for i := range state.Players {
		p := &state.Players[i]
		if p.ClientID == localID {
			own = p
		} else {
			opponent = p
		}
	}
	if own == nil || opponent == nil || state.AttackerID != own.ClientID && state.AttackerID != opponent.ClientID {
		return local, other, fmt.Errorf("invalid combat players/roles")
	}
	if _, err := protocol.ValidateLoadout(own.Loadout, state.Budget, true); err != nil {
		return local, other, err
	}
	local = protocol.CombatSetup{Type: "combat-setup", GameID: state.ID, ClientID: localID, Name: own.Name, Budget: state.Budget, AttackerID: state.AttackerID, Revision: state.Revision, Loadout: own.Loadout.Clone()}
	peer.SetReadDeadline(time.Now().Add(30 * time.Second))
	defer peer.SetReadDeadline(time.Time{})
	send := func() error {
		bytes, err := json.Marshal(local)
		if err != nil {
			return err
		}
		return peer.Send(bytes)
	}
	receive := func() error {
		bytes, err := peer.Read()
		if err != nil {
			return err
		}
		return json.Unmarshal(bytes, &other)
	}
	// Deterministic ordering works with both blocking packet streams and websockets.
	if localID == state.AttackerID {
		if err := send(); err != nil {
			return local, other, err
		}
		if err := receive(); err != nil {
			return local, other, err
		}
	} else {
		if err := receive(); err != nil {
			return local, other, err
		}
		if err := send(); err != nil {
			return local, other, err
		}
	}
	if other.Type != "combat-setup" || other.GameID != state.ID || other.ClientID != opponent.ClientID || other.Budget != state.Budget || other.AttackerID != state.AttackerID || other.Revision != state.Revision || !other.Loadout.Equal(opponent.Loadout) {
		return local, other, fmt.Errorf("peer combat setup does not match the approved game setup")
	}
	if _, err := protocol.ValidateLoadout(other.Loadout, state.Budget, true); err != nil {
		return local, other, err
	}
	other.Name = opponent.Name
	return local, other, nil
}

func makeConfiguredPlayer(spec protocol.CombatSetup, attacker, human bool, allSpells spellbook.Spells) (*player.Player, *combat.Army, error) {
	if _, err := protocol.ValidateLoadout(spec.Loadout, spec.Budget, true); err != nil {
		return nil, nil, err
	}
	books := []data.WizardBook{}
	for i, count := range spec.Loadout.Books {
		if count > 0 {
			books = append(books, data.WizardBook{Magic: data.MagicType(i + 1), Count: int(count)})
		}
	}
	banner := data.BannerBrown
	if attacker {
		banner = data.BannerGreen
	}
	wizard := setup.WizardCustom{Name: spec.Name, Base: data.WizardBase(spec.Loadout.Portrait), Portrait: spec.Loadout.Portrait, Books: books, Banner: banner}
	owner := player.MakePlayer(wizard, human, 0, 0, nil, &noGlobalEnchantments{})
	owner.Mana = int(spec.Loadout.Mana)
	owner.CastingSkillPower = 50000
	for _, name := range spec.Loadout.Spells {
		spell := allSpells.FindByName(name)
		if !spell.Valid() {
			return nil, nil, fmt.Errorf("spell %q unavailable in local game assets", name)
		}
		owner.KnownSpells.AddSpell(spell)
	}
	// A fortress enables wizard casting, but the battle itself is an open field
	// without the CLI scenario's free city enchantments/spells.
	fortress := citylib.MakeCity("Fortress", 10, 10, wizard.Race, nil, &BasicCatchment{}, nil, owner)
	fortress.Buildings.Insert(buildinglib.BuildingFortress)
	owner.AddCity(fortress)
	army := &combat.Army{Player: owner}
	for _, id := range spec.Loadout.Army {
		unit := units.GetUnitById(units.UnitId{LbxFile: id.LbxFile, Index: id.Index})
		if unit.Name == "" {
			return nil, nil, fmt.Errorf("army unit unavailable in local game assets")
		}
		army.AddUnit(units.MakeOverworldUnitFromUnit(unit, 1, 1, data.PlaneArcanus, banner, owner.MakeExperienceInfo(), owner.MakeUnitEnchantmentProvider()))
	}
	return owner, army, nil
}

func newConfiguredCombat(peer netlib.PacketConn, localID uint64, state protocol.GameState) (*Engine, error) {
	local, other, err := exchangeCombatSetup(peer, localID, state)
	if err != nil {
		return nil, err
	}
	cache := lbx.AutoCache()
	if cache == nil {
		return nil, fmt.Errorf("game assets not found")
	}
	spells, err := spellbook.ReadSpellsFromCache(cache)
	if err != nil {
		return nil, err
	}
	attacker, defender := local, other
	localAttacker := localID == state.AttackerID
	if !localAttacker {
		attacker, defender = other, local
	}
	attackingPlayer, attackingArmy, err := makeConfiguredPlayer(attacker, true, localAttacker, spells)
	if err != nil {
		return nil, err
	}
	defendingPlayer, defendingArmy, err := makeConfiguredPlayer(defender, false, !localAttacker, spells)
	if err != nil {
		return nil, err
	}
	remote := combat.MakeRemote(!localAttacker, !localAttacker, peer)
	model := combat.MakeCombatModel(spells, defendingArmy, attackingArmy, combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, data.MagicNone, 10, 25, make(chan combat.CombatEvent, 10), remote)
	human := attackingPlayer
	if !localAttacker {
		human = defendingPlayer
	}
	screen := combat.MakeCombatScreen(cache, defendingArmy, attackingArmy, optional.Of[combat.ArmyPlayer](human), combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, model)
	if screen == nil {
		return nil, fmt.Errorf("unable to create combat screen")
	}
	go remote.RunReceiveLoop(screen.Quit)
	return makeCombatEngine(model, screen), nil
}
