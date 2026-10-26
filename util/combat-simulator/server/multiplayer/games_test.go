package multiplayer

import (
	"sync"
	"testing"

	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

func testGameManager() (*Manager, map[uint64]*User) {
	manager := &Manager{Lobbies: map[string]*Lobby{
		"Public": {Name: "Public", Public: true, persistent: true, Users: make(map[uint64]*User)},
		"Other":  {Name: "Other", Public: true, persistent: true, Users: make(map[uint64]*User)},
	}}
	users := make(map[uint64]*User)
	for id := uint64(1); id <= 4; id++ {
		room := "Public"
		if id == 4 {
			room = "Other"
		}
		user := &User{name: "Player", room: room, messages: make(chan protocol.Message, 256)}
		users[id] = user
		manager.Lobbies[room].Users[id] = user
	}
	return manager, users
}

func gameCommand(manager *Manager, user *User, id uint64, command protocol.Message) {
	manager.lock.With(func() error { manager.handleGameCommand(user, id, command); return nil })
}

func drainGameMessages(user *User) []protocol.Message {
	var messages []protocol.Message
	for {
		select {
		case message := <-user.messages:
			messages = append(messages, message)
		default:
			return messages
		}
	}
}

func requireGameError(t *testing.T, user *User) {
	t.Helper()
	for _, message := range drainGameMessages(user) {
		if message.Error != "" {
			return
		}
	}
	t.Fatal("command was not rejected")
}

func TestGameLifecycle(t *testing.T) {
	manager, users := testGameManager()
	host, guest, third, outsider := users[1], users[2], users[3], users[4]
	gameCommand(manager, host, 1, protocol.Message{Type: "create-game"})
	id := host.gameID
	if id == 0 || len(manager.Games) != 1 || len(manager.waitingGames("Public")) != 1 {
		t.Fatal("game not created")
	}
	if len(manager.waitingGames("Other")) != 0 {
		t.Fatal("games leaked across lobbies")
	}
	gameCommand(manager, host, 1, protocol.Message{Type: "create-game"})
	requireGameError(t, host)
	if len(manager.Games) != 1 {
		t.Fatal("duplicate hosting allowed")
	}
	gameCommand(manager, outsider, 4, protocol.Message{Type: "join-game", GameID: id})
	requireGameError(t, outsider)
	gameCommand(manager, guest, 2, protocol.Message{Type: "join-game", GameID: id})
	if guest.gameID != id || len(manager.Games[id].players) != 2 || len(manager.waitingGames("Public")) != 0 {
		t.Fatal("guest did not fill game")
	}
	gameCommand(manager, third, 3, protocol.Message{Type: "join-game", GameID: id})
	requireGameError(t, third)
	gameCommand(manager, third, 3, protocol.Message{Type: "game-ready", GameID: id, Ready: true})
	requireGameError(t, third)
	gameCommand(manager, host, 1, protocol.Message{Type: "game-ready", GameID: id, Ready: true})
	if manager.Games[id].started {
		t.Fatal("game started with one ready player")
	}
	gameCommand(manager, guest, 2, protocol.Message{Type: "game-ready", GameID: id, Ready: true})
	if !manager.Games[id].started {
		t.Fatal("both players ready did not mark start hook")
	}
	gameCommand(manager, guest, 2, protocol.Message{Type: "leave-game", GameID: id})
	if guest.gameID != 0 || manager.Games[id].started || manager.Games[id].ready[1] || len(manager.waitingGames("Public")) != 1 {
		t.Fatal("guest exit did not reset/reopen game")
	}
	gameCommand(manager, guest, 2, protocol.Message{Type: "create-game"})
	guestGame := guest.gameID
	if guestGame == 0 || guestGame == id {
		t.Fatal("released guest could not host a game")
	}
	gameCommand(manager, host, 1, protocol.Message{Type: "join-game", GameID: guestGame})
	requireGameError(t, host)
	gameCommand(manager, third, 3, protocol.Message{Type: "join-game", GameID: id})
	manager.lock.With(func() error { manager.leave(host, 1); return nil })
	if manager.Games[id] != nil || host.gameID != 0 || third.gameID != 0 {
		t.Fatal("host room departure did not release players")
	}
	found := false
	for _, message := range drainGameMessages(third) {
		found = found || message.Type == "game-left" && message.GameID == id
	}
	if !found {
		t.Fatal("guest not notified of host departure")
	}
	manager.lock.With(func() error { manager.leave(guest, 2); return nil })
	if len(manager.Games) != 0 {
		t.Fatal("room departure leaked games")
	}
}

func TestConcurrentGameJoins(t *testing.T) {
	manager, users := testGameManager()
	gameCommand(manager, users[1], 1, protocol.Message{Type: "create-game"})
	id := users[1].gameID
	var wait sync.WaitGroup
	for _, clientID := range []uint64{2, 3} {
		wait.Go(func() {
			gameCommand(manager, users[clientID], clientID, protocol.Message{Type: "join-game", GameID: id})
		})
	}
	wait.Wait()
	if len(manager.Games[id].players) != 2 {
		t.Fatal("game did not enforce two-player limit")
	}
	if (users[2].gameID == id) == (users[3].gameID == id) {
		t.Fatal("expected exactly one successful guest")
	}
}
