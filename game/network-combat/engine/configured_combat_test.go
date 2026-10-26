package engine

import (
	"context"
	"encoding/json/v2"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/kazzmir/master-of-magic/game/magic/data"
	"github.com/kazzmir/master-of-magic/game/network-combat/engine/crypto"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
	"github.com/kazzmir/master-of-magic/lib/lbx"
	netlib "github.com/kazzmir/master-of-magic/lib/net"
)

func configuredTestState() protocol.GameState {
	host, guest := protocol.DefaultLoadout(true), protocol.DefaultLoadout(false)
	host.Mana = 55
	host.Portrait = 5
	host.Books[4] = 1
	if _, ok := protocol.FindSpell("Fire Bolt"); ok {
		host.Spells = []string{"Fire Bolt"}
	}
	guest.Mana = 88
	guest.Portrait = 2
	return protocol.GameState{ID: 12345678901, HostID: 1, AttackerID: 2, Budget: 10000, Revision: 5, Started: true, Room: "Public", Players: []protocol.GamePlayer{
		{ClientID: 1, Name: "Host", Ready: true, Loadout: host}, {ClientID: 2, Name: "Guest", Ready: true, Loadout: guest},
	}}
}

func combatTestPeers(t *testing.T) [2]netlib.PacketConn {
	t.Helper()
	first, second := net.Pipe()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(func() { cancel(); first.Close(); second.Close() })
	go func() { <-ctx.Done(); first.Close(); second.Close() }()
	return [2]netlib.PacketConn{netlib.NewPacketNetConn(first, MAX_PACKET_SIZE), netlib.NewPacketNetConn(second, MAX_PACKET_SIZE)}
}

func TestConfiguredPeerSetupExchange(t *testing.T) {
	peers := combatTestPeers(t)
	state := configuredTestState()
	var own, other [2]protocol.CombatSetup
	var failures [2]error
	var wait sync.WaitGroup
	for i := range peers {
		wait.Go(func() {
			encrypted, err := crypto.NewEncryptedConn(peers[i])
			if err != nil {
				failures[i] = err
				return
			}
			own[i], other[i], failures[i] = exchangeCombatSetup(encrypted, uint64(i+1), state)
		})
	}
	wait.Wait()
	for i, err := range failures {
		if err != nil {
			t.Fatal(err)
		}
		if !own[i].Loadout.Equal(state.Players[i].Loadout) || !other[i].Loadout.Equal(state.Players[1-i].Loadout) {
			t.Fatal("mana/books/spells/army/portrait not exchanged")
		}
	}
}

type alteredSetupPeer struct{ netlib.PacketConn }

func (peer alteredSetupPeer) Read() ([]byte, error) {
	data, err := peer.PacketConn.Read()
	if err != nil {
		return nil, err
	}
	var setup protocol.CombatSetup
	if err = json.Unmarshal(data, &setup); err != nil {
		return nil, err
	}
	setup.Loadout.Mana++
	return json.Marshal(setup)
}
func TestRejectTamperedCombatLoadout(t *testing.T) {
	peers := combatTestPeers(t)
	peers[0] = alteredSetupPeer{peers[0]}
	state := configuredTestState()
	var failures [2]error
	var wait sync.WaitGroup
	for i := range peers {
		wait.Go(func() { _, _, failures[i] = exchangeCombatSetup(peers[i], uint64(i+1), state) })
	}
	wait.Wait()
	if failures[0] == nil {
		t.Fatal("peer changed approved mana without detection")
	}
}

func TestConfiguredCombatModels(t *testing.T) {
	if lbx.AutoCache() == nil {
		t.Skip("combat game assets unavailable")
	}
	peers := combatTestPeers(t)
	state := configuredTestState()
	var engines [2]*Engine
	var failures [2]error
	var wait sync.WaitGroup
	for i := range peers {
		wait.Go(func() { engines[i], failures[i] = newConfiguredCombat(peers[i], uint64(i+1), state) })
	}
	wait.Wait()
	for _, engine := range engines {
		if engine != nil {
			defer closeCombatEngine(engine)
		}
	}
	for _, err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, engine := range engines {
		attacker, defender := engine.Model.AttackingArmy, engine.Model.DefendingArmy
		if attacker.Player.GetWizard().Name != "Guest" || defender.Player.GetWizard().Name != "Host" {
			t.Fatal("creator's attacker/defender assignment ignored")
		}
		if attacker.Player.GetMana() != 88 || defender.Player.GetMana() != 55 {
			t.Fatal("chosen mana not used")
		}
		if defender.Player.GetWizard().Base != data.WizardOberic || defender.Player.GetWizard().MagicLevel(data.ChaosMagic) != 1 {
			t.Fatal("chosen wizard/books not used")
		}
		known := defender.Player.GetKnownSpells()
		if !known.FindByName("Fire Bolt").Valid() {
			t.Fatal("chosen spell not known")
		}
		if known.FindByName("Wall of Fire").Valid() {
			t.Fatal("unbought spell granted")
		}
		if len(attacker.GetUnits()) != 2 || len(defender.GetUnits()) != 2 {
			t.Fatal("chosen armies not used")
		}
		if attacker.GetUnits()[0].Unit.GetRawUnit().GetId().Index != 0 || defender.GetUnits()[0].Unit.GetRawUnit().GetId().Index != 17 {
			t.Fatal("chosen units changed when swapping roles")
		}
	}
}
