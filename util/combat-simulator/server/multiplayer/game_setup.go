package multiplayer

import (
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

func (manager *Manager) handleGameSetup(user *User, clientID uint64, command protocol.Message) {
	fail := func(reason string) {
		user.send(protocol.Message{ID: command.ID, GameID: command.GameID, Error: reason})
	}
	game := manager.Games[user.gameID]
	if game == nil || game.id != command.GameID {
		fail("You are not in that game")
		return
	}
	if game.started {
		fail("Combat has already started")
		return
	}
	changed := false
	switch command.Type {
	case "game-settings":
		if game.hostID != clientID {
			fail("Only the creator can change the budget or player roles")
			return
		}
		if game.players[command.AttackerID] == nil {
			fail("Select a player in this game as attacker")
			return
		}
		changed = game.budget != command.Budget || game.attackerID != command.AttackerID
		game.budget, game.attackerID = command.Budget, command.AttackerID
	case "game-loadout":
		if command.Revision != game.revision {
			fail("Game setup changed; review and save your loadout again")
			return
		}
		if command.Loadout == nil {
			fail("Missing player loadout")
			return
		}
		if _, err := protocol.ValidateLoadout(*command.Loadout, game.budget, false); err != nil {
			fail(err.Error())
			return
		}
		changed = !game.loadouts[clientID].Equal(*command.Loadout)
		game.loadouts[clientID] = command.Loadout.Clone()
	}
	if changed {
		game.revision++
		for id := range game.ready {
			game.ready[id] = false
		}
	}
	game.broadcastState()
	manager.broadcastGames(game.room)
}
