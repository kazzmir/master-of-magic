package multiplayer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

func TestGameWebsocketFlow(t *testing.T) {
	mux := http.NewServeMux()
	InitMux(mux, testAuth)
	server := httptest.NewServer(mux)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	write := func(conn *websocket.Conn, message any) {
		t.Helper()
		if err := wsjson.Write(ctx, conn, message); err != nil {
			t.Fatal(err)
		}
	}
	readKind := func(conn *websocket.Conn, kind string) protocol.Message {
		t.Helper()
		for {
			var message protocol.Message
			if err := wsjson.Read(ctx, conn, &message); err != nil {
				t.Fatal(err)
			}
			if message.Error != "" {
				t.Fatal(message)
			}
			if message.Type == kind {
				return message
			}
		}
	}
	connect := func(name string) *websocket.Conn {
		t.Helper()
		conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/multiplayer", nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { conn.CloseNow() })
		write(conn, map[string]string{"auth": testAuth, "name": name})
		readKind(conn, "connected")
		readKind(conn, "version")
		write(conn, protocol.Message{Type: "version", Version: 1})
		write(conn, protocol.Message{Type: "join-room", Room: "Public"})
		readKind(conn, "join-room")
		return conn
	}
	host, guest := connect("Host"), connect("Guest")
	write(host, protocol.Message{Type: "create-game"})
	created := readKind(host, "game-state").Game
	if created == nil || len(created.Players) != 1 || created.Players[0].Name != "Host" {
		t.Fatal(created)
	}
	if list := readKind(guest, "list-games"); len(list.Games) != 1 || list.Games[0].ID != created.ID {
		t.Fatal(list)
	}
	write(guest, protocol.Message{Type: "join-game", GameID: created.ID})
	for _, conn := range []*websocket.Conn{host, guest} {
		if state := readKind(conn, "game-state").Game; state == nil || len(state.Players) != 2 || state.Players[1].Name != "Guest" {
			t.Fatal(state)
		}
	}
	write(host, protocol.Message{Type: "game-chat", GameID: created.ID, Text: "hello guest"})
	for _, conn := range []*websocket.Conn{host, guest} {
		if message := readKind(conn, "game-chat"); message.GameID != created.ID || message.ClientID != created.HostID || message.Name != "Host" || message.Text != "hello guest" {
			t.Fatal(message)
		}
	}
	write(host, protocol.Message{Type: "game-ready", GameID: created.ID, Ready: true})
	for _, conn := range []*websocket.Conn{host, guest} {
		if state := readKind(conn, "game-state").Game; state == nil || state.Started || !state.Players[0].Ready {
			t.Fatal(state)
		}
	}
	write(guest, protocol.Message{Type: "game-ready", GameID: created.ID, Ready: true})
	for _, conn := range []*websocket.Conn{host, guest} {
		if state := readKind(conn, "game-state").Game; state == nil || !state.Started {
			t.Fatal(state)
		}
	}
	write(guest, protocol.Message{Type: "leave-game", GameID: created.ID})
	if state := readKind(host, "game-state").Game; state == nil || len(state.Players) != 1 || state.Players[0].Ready || state.Started {
		t.Fatal(state)
	}
	if message := readKind(guest, "game-left"); message.GameID != created.ID {
		t.Fatal(message)
	}
	write(guest, protocol.Message{Type: "join-game", GameID: created.ID})
	readKind(host, "game-state")
	readKind(guest, "game-state")
	host.CloseNow()
	if message := readKind(guest, "game-left"); message.GameID != created.ID {
		t.Fatal(message)
	}
}
