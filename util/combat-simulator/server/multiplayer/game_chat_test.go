package multiplayer

import (
	"strings"
	"testing"

	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

func TestPrivateGameChat(t *testing.T) {
	manager, users := testGameManager()
	gameCommand(manager, users[1], 1, protocol.Message{Type: "create-game"})
	id := users[1].gameID
	gameCommand(manager, users[2], 2, protocol.Message{Type: "join-game", GameID: id})
	gameCommand(manager, users[1], 1, protocol.Message{Type: "game-ready", GameID: id, Ready: true})
	for _, user := range users {
		drainGameMessages(user)
	}
	gameCommand(manager, users[1], 1, protocol.Message{Type: "game-chat", GameID: id, Text: "  hello opponent  ", ClientID: 3, Name: "spoofed"})
	for _, clientID := range []uint64{1, 2} {
		messages := drainGameMessages(users[clientID])
		if len(messages) != 1 {
			t.Fatal("missing/extra game chat messages", messages)
		}
		message := messages[0]
		if message.Type != "game-chat" || message.GameID != id || message.ClientID != 1 || message.Name != users[1].name || message.Text != "hello opponent" {
			t.Fatal(message)
		}
	}
	for _, clientID := range []uint64{3, 4} {
		if len(drainGameMessages(users[clientID])) != 0 {
			t.Fatal("game chat leaked to other users")
		}
	}
	if !manager.Games[id].ready[1] || manager.Games[id].revision != 0 {
		t.Fatal("chat changed readiness/setup")
	}
	gameCommand(manager, users[3], 3, protocol.Message{Type: "game-chat", GameID: id, Text: "intruder"})
	requireGameError(t, users[3])
	for _, text := range []string{"", "   ", strings.Repeat("x", 2001)} {
		gameCommand(manager, users[1], 1, protocol.Message{Type: "game-chat", GameID: id, Text: text})
		requireGameError(t, users[1])
		if len(drainGameMessages(users[2])) != 0 {
			t.Fatal("invalid chat delivered")
		}
	}
	gameCommand(manager, users[2], 2, protocol.Message{Type: "leave-game", GameID: id})
	for _, user := range users {
		drainGameMessages(user)
	}
	gameCommand(manager, users[1], 1, protocol.Message{Type: "game-chat", GameID: id, Text: "after exit"})
	if len(drainGameMessages(users[1])) != 1 || len(drainGameMessages(users[2])) != 0 {
		t.Fatal("departed player received game chat")
	}
}
