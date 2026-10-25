package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

func TestMultiplayerURL(t *testing.T) {
	for input, want := range map[string]string{
		"localhost:5000":                "ws://localhost:5000/multiplayer",
		" https://example.com/ignored ": "wss://example.com/multiplayer",
		"ws://localhost:5000":           "ws://localhost:5000/multiplayer",
	} {
		got, err := multiplayerURL(input)
		if err != nil || got != want {
			t.Fatalf("%q: %q, %v", input, got, err)
		}
	}
	for _, input := range []string{"", "ftp://example.com", "ws://"} {
		if _, err := multiplayerURL(input); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
}

func TestLobbyUI(t *testing.T) {
	game := newLobbyGame()
	if err := game.Update(); err != nil {
		t.Fatal(err)
	}
	game.Draw(ebiten.NewImage(1100, 700))
	session := &lobbySession{cancel: func() {}, commands: make(chan protocol.Message, 4)}
	game.session = session
	for _, message := range []protocol.Message{
		{Type: "connected", ClientID: 1},
		{Type: "list-rooms", Rooms: []string{"Lobby"}},
		{Type: "join-room", Room: "Lobby", ClientID: 1},
		{Type: "room-users", Room: "Lobby", Users: []uint64{1, 2}},
		{Type: "chat", Room: "Lobby", ClientID: 2, Text: "[hello]"},
	} {
		game.handleEvent(lobbyEvent{session: session, message: message})
	}
	if game.clientID != 1 || !game.connected {
		t.Fatal("server-assigned client ID not stored")
	}
	if game.chat.GetText() != "Player-000002: [hello]" {
		t.Fatal(game.chat.GetText())
	}
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "user-left", Room: "Elsewhere", ClientID: 2}})
	if !slices.Equal(game.userIDs, []uint64{1, 2}) {
		t.Fatal("departure from another room changed membership", game.userIDs)
	}
	for _, id := range []uint64{2, 2, 99} {
		game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "user-left", Room: "Lobby", ClientID: id}})
		if !slices.Equal(game.userIDs, []uint64{1}) {
			t.Fatal("incorrect membership after departure", game.userIDs)
		}
	}
	game.input.SetText("hello")
	game.sendChat()
	if command := <-session.commands; command.Type != "chat" || command.Text != "hello" {
		t.Fatal(command)
	}
	game.disconnect("Disconnected")
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "connected", ClientID: 1}})
	if game.connected || game.clientID != 0 || game.room != "" || game.chat.GetText() != "" || len(game.userIDs) != 0 {
		t.Fatal("stale session restored state")
	}
	game.Update()
	game.Draw(ebiten.NewImage(800, 500))
}

func TestLobbySessionHandshake(t *testing.T) {
	errors := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/multiplayer" {
			t.Errorf("path: %s", r.URL.Path)
		}
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			errors <- err
			return
		}
		defer conn.CloseNow()
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		var auth map[string]string
		if err := wsjson.Read(ctx, conn, &auth); err != nil {
			errors <- err
			return
		}
		if auth["auth"] != "x" {
			t.Error("missing placeholder auth")
		}
		if err := wsjson.Write(ctx, conn, protocol.Message{Type: "connected", ClientID: 42}); err != nil {
			errors <- err
			return
		}
		var command protocol.Message
		if err := wsjson.Read(ctx, conn, &command); err != nil {
			errors <- err
			return
		}
		if command.Type != "list-rooms" {
			t.Errorf("first command: %+v", command)
		}
		errors <- wsjson.Write(ctx, conn, protocol.Message{Type: "list-rooms", Rooms: []string{"Lobby"}})
		conn.Read(ctx)
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	session := &lobbySession{cancel: cancel, commands: make(chan protocol.Message, 4)}
	events := make(chan lobbyEvent, 8)
	go runLobbySession(ctx, session, strings.TrimPrefix(server.URL, "http://"), events)
	var clientID uint64
	for {
		select {
		case event := <-events:
			if event.err != nil {
				t.Fatal(event.err)
			}
			if event.message.Type == "connected" {
				clientID = event.message.ClientID
			}
			if event.message.Type == "list-rooms" {
				if clientID != 42 {
					t.Fatal("missing server-assigned client ID", clientID)
				}
				if err := <-errors; err != nil {
					t.Fatal(err)
				}
				return
			}
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
}
