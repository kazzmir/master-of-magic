package multiplayer

import (
	"math"
	"testing"

	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

func TestGameBudgetRolesAndLoadouts(t *testing.T) {
	manager, users := testGameManager()
	gameCommand(manager, users[1], 1, protocol.Message{Type: "create-game"})
	id := users[1].gameID
	gameCommand(manager, users[2], 2, protocol.Message{Type: "join-game", GameID: id})
	game := manager.Games[id]
	if game.budget != protocol.DefaultBudget || game.attackerID != 1 {
		t.Fatal("invalid defaults")
	}
	gameCommand(manager, users[2], 2, protocol.Message{Type: "game-settings", GameID: id, Budget: 1, AttackerID: 2})
	requireGameError(t, users[2])
	if game.budget != protocol.DefaultBudget || game.attackerID != 1 {
		t.Fatal("guest changed host settings")
	}
	gameCommand(manager, users[1], 1, protocol.Message{Type: "game-settings", GameID: id, Budget: 1, AttackerID: 3})
	requireGameError(t, users[1])
	if game.budget != protocol.DefaultBudget || game.attackerID != 1 {
		t.Fatal("nonparticipant assigned a combat role")
	}
	gameCommand(manager, users[1], 1, protocol.Message{Type: "game-ready", GameID: id, Ready: true})
	if !game.ready[1] {
		t.Fatal("host not ready")
	}
	loadout := game.loadouts[2].Clone()
	loadout.Mana = 10
	loadout.Portrait = 5
	loadout.Books[0] = 1
	gameCommand(manager, users[2], 2, protocol.Message{Type: "game-loadout", GameID: id, Loadout: &loadout})
	if game.ready[1] || game.revision != 1 || !game.loadouts[2].Equal(loadout) {
		t.Fatal("loadout update did not reset readiness/save selections")
	}
	if game.loadouts[1].Mana != 0 {
		t.Fatal("guest modified host's loadout")
	}
	gameCommand(manager, users[1], 1, protocol.Message{Type: "game-ready", GameID: id, Ready: true, Revision: 0})
	requireGameError(t, users[1])
	gameCommand(manager, users[1], 1, protocol.Message{Type: "game-settings", GameID: id, Budget: math.MaxUint64, AttackerID: 2})
	if game.budget != math.MaxUint64 || game.attackerID != 2 || game.revision != 2 {
		t.Fatal("uint64 budget/role assignment not saved")
	}
	state := game.snapshot()
	if state.Budget != math.MaxUint64 || state.AttackerID != 2 || !state.Players[1].Loadout.Equal(loadout) {
		t.Fatal("setup missing from snapshot")
	}
	gameCommand(manager, users[1], 1, protocol.Message{Type: "game-settings", GameID: id, Budget: 0, AttackerID: 2})
	gameCommand(manager, users[1], 1, protocol.Message{Type: "game-ready", GameID: id, Ready: true, Revision: game.revision})
	requireGameError(t, users[1])
	if game.ready[1] {
		t.Fatal("over-budget player became ready")
	}
	gameCommand(manager, users[1], 1, protocol.Message{Type: "game-settings", GameID: id, Budget: protocol.DefaultBudget, AttackerID: 2})
	bad := loadout.Clone()
	bad.Mana = protocol.DefaultBudget
	gameCommand(manager, users[2], 2, protocol.Message{Type: "game-loadout", GameID: id, Revision: game.revision, Loadout: &bad})
	requireGameError(t, users[2])
	if !game.loadouts[2].Equal(loadout) {
		t.Fatal("rejected overspending changed saved loadout")
	}
	for _, clientID := range []uint64{1, 2} {
		gameCommand(manager, users[clientID], clientID, protocol.Message{Type: "game-ready", GameID: id, Ready: true, Revision: game.revision})
	}
	if !game.started {
		t.Fatal("valid configured game did not start")
	}
	for _, clientID := range []uint64{1, 2} {
		found := false
		for _, message := range drainGameMessages(users[clientID]) {
			if message.Type == "combat-start" {
				found = true
				if message.Attacker != (clientID == 2) {
					t.Fatal("host's role assignment ignored")
				}
			}
		}
		if !found {
			t.Fatal("no configured combat start")
		}
	}
	manager.lock.With(func() error { manager.leaveGame(users[1], 1); return nil })
}
