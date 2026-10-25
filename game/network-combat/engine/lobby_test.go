package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
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
	game.handleEvent(lobbyEvent{session: session, connected: true})
	for _, message := range []protocol.Message{
		{Type: "list-rooms", Rooms: []string{"Lobby"}},
		{Type: "join-room", Room: "Lobby", User: "Me"},
		{Type: "room-users", Room: "Lobby", Users: []string{"Me", "Other"}},
		{Type: "chat", Room: "Lobby", User: "Other", Text: "[hello]"},
	} {
		game.handleEvent(lobbyEvent{session: session, message: message})
	}
	if game.chat.GetText() != "Other: [hello]" {
		t.Fatal(game.chat.GetText())
	}
	game.input.SetText("hello")
	game.sendChat()
	if command := <-session.commands; command.Type != "chat" || command.Text != "hello" {
		t.Fatal(command)
	}
	game.disconnect("Disconnected")
	game.handleEvent(lobbyEvent{session: session, connected: true})
	if game.connected || game.room != "" || game.chat.GetText() != "" {
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
	for {
		select {
		case event := <-events:
			if event.err != nil {
				t.Fatal(event.err)
			}
			if event.message.Type == "list-rooms" {
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
