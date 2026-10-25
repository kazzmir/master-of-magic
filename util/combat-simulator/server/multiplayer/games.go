package multiplayer

import (
	"slices"

	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

// All matchmaking state is accessed within manager.lock.With.
type Game struct {
	id      uint64
	room    string
	hostID  uint64
	players map[uint64]*User
	ready   map[uint64]bool
	started bool
}

func (game *Game) snapshot() protocol.GameState {
	state := protocol.GameState{ID: game.id, Room: game.room, HostID: game.hostID, Started: game.started}
	ids := make([]uint64, 0, len(game.players))
	for id := range game.players {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	// Keep the host in the first slot even if a guest has a lower ID.
	state.Players = append(state.Players, protocol.GamePlayer{ClientID: game.hostID, Name: game.players[game.hostID].name, Ready: game.ready[game.hostID]})
	for _, id := range ids {
		if id != game.hostID {
			state.Players = append(state.Players, protocol.GamePlayer{ClientID: id, Name: game.players[id].name, Ready: game.ready[id]})
		}
	}
	return state
}

func (manager *Manager) waitingGames(room string) []protocol.GameState {
	ids := make([]uint64, 0)
	for id, game := range manager.Games {
		if game.room == room && len(game.players) < 2 && !game.started {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	out := make([]protocol.GameState, 0, len(ids))
	for _, id := range ids {
		out = append(out, manager.Games[id].snapshot())
	}
	return out
}

func (manager *Manager) broadcastGames(room string) {
	if lobby := manager.Lobbies[room]; lobby != nil {
		broadcast(lobby, protocol.Message{Type: "list-games", Room: room, Games: manager.waitingGames(room)})
	}
}

func (game *Game) broadcastState() {
	state := game.snapshot()
	for _, user := range game.players {
		user.send(protocol.Message{Type: "game-state", Game: &state})
	}
}

// newGame is called only after checking that both participants are available.
func (manager *Manager) newGame(hostID uint64, host *User, guestID uint64, guest *User) *Game {
	manager.cancelInvitations(hostID, "Game request cancelled: a player joined another game")
	if guest != nil {
		manager.cancelInvitations(guestID, "Game request cancelled: a player joined another game")
	}
	if manager.Games == nil {
		manager.Games = make(map[uint64]*Game)
	}
	manager.nextGameID++
	game := &Game{id: manager.nextGameID, room: host.room, hostID: hostID, players: map[uint64]*User{hostID: host}, ready: map[uint64]bool{hostID: false}}
	manager.Games[game.id] = game
	host.gameID = game.id
	if guest != nil {
		game.players[guestID] = guest
		game.ready[guestID] = false
		guest.gameID = game.id
	}
	game.broadcastState()
	manager.broadcastGames(host.room)
	return game
}

func (manager *Manager) leaveGame(user *User, clientID uint64) {
	game := manager.Games[user.gameID]
	if game == nil {
		user.gameID = 0
		return
	}
	if clientID == game.hostID {
		// The host owns the game. Their departure releases both players.
		for _, player := range game.players {
			player.gameID = 0
			player.send(protocol.Message{Type: "game-left", GameID: game.id})
		}
		delete(manager.Games, game.id)
	} else {
		delete(game.players, clientID)
		delete(game.ready, clientID)
		user.gameID = 0
		user.send(protocol.Message{Type: "game-left", GameID: game.id})
		game.started = false
		for id := range game.ready {
			game.ready[id] = false
		}
		game.broadcastState()
	}
	manager.broadcastGames(game.room)
}

// handleGameCommand returns false for commands belonging to other subsystems.
// Each player may participate in one game, which also limits hosts to one game.
func (manager *Manager) handleGameCommand(user *User, clientID uint64, command protocol.Message) bool {
	switch command.Type {
	case "list-games", "create-game", "join-game", "game-ready", "leave-game":
	default:
		return false
	}
	fail := func(reason string) { user.send(protocol.Message{ID: command.ID, Error: reason}) }
	if manager.Lobbies[user.room] == nil {
		fail("Join a room first")
		return true
	}
	switch command.Type {
	case "list-games":
		user.send(protocol.Message{Type: "list-games", ID: command.ID, Room: user.room, Games: manager.waitingGames(user.room)})
	case "create-game":
		if user.gameID != 0 {
			fail("Exit your current game first")
			break
		}
		manager.newGame(clientID, user, 0, nil)
	case "join-game":
		game := manager.Games[command.GameID]
		if user.gameID != 0 {
			fail("Exit your current game first")
		} else if game == nil || game.room != user.room {
			fail("Game not found in this room")
		} else if len(game.players) >= 2 || game.started {
			fail("Game is full")
		} else {
			manager.cancelInvitations(clientID, "Game request cancelled: a player joined another game")
			game.players[clientID] = user
			user.gameID = game.id
			for id := range game.players {
				game.ready[id] = false
			}
			game.broadcastState()
			manager.broadcastGames(user.room)
		}
	case "game-ready":
		game := manager.Games[user.gameID]
		if game == nil || command.GameID != user.gameID {
			fail("You are not in that game")
		} else if game.started {
			fail("Both players are already ready")
		} else {
			game.ready[clientID] = command.Ready
			if len(game.players) == 2 {
				game.started = true
				for id := range game.players {
					game.started = game.started && game.ready[id]
				}
			}
			// Future hook: launch combat when started becomes true.
			game.broadcastState()
		}
	case "leave-game":
		if user.gameID == 0 || command.GameID != user.gameID {
			fail("You are not in that game")
		} else {
			manager.leaveGame(user, clientID)
		}
	}
	return true
}
