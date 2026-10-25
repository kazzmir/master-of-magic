package multiplayer

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
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

func TestCreateRoom(t *testing.T) {
	mux := http.NewServeMux()
	InitMux(mux, testAuth)
	server := httptest.NewServer(mux)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	read := func(conn *websocket.Conn) protocol.Message {
		t.Helper()
		var message protocol.Message
		if err := wsjson.Read(ctx, conn, &message); err != nil {
			t.Fatal(err)
		}
		return message
	}
	write := func(conn *websocket.Conn, message any) {
		t.Helper()
		if err := wsjson.Write(ctx, conn, message); err != nil {
			t.Fatal(err)
		}
	}
	connect := func() *websocket.Conn {
		t.Helper()
		conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/multiplayer", nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { conn.CloseNow() })
		write(conn, map[string]string{"auth": testAuth, "name": "Creator"})
		if reply := read(conn); reply.Type != "connected" {
			t.Fatal(reply)
		}
		if reply := read(conn); reply.Type != "version" || reply.Version != 1 {
			t.Fatal(reply)
		}
		write(conn, protocol.Message{Type: "version", Version: protocol.ClientVersion})
		return conn
	}
	creator := connect()
	for _, name := range []string{"   ", "bad\nname", strings.Repeat("a", 31), strings.Repeat("界", 31)} {
		write(creator, protocol.Message{Type: "create-room", ID: 1, Room: name})
		if reply := read(creator); reply.Error == "" || reply.ID != 1 {
			t.Fatal(reply)
		}
	}
	write(creator, protocol.Message{Type: "create-room", ID: 2, Room: "  New lobby  "})
	if reply := read(creator); reply.Type != "create-room" || reply.Room != "New lobby" || reply.ID != 2 {
		t.Fatal(reply)
	}
	write(creator, protocol.Message{Type: "create-room", Room: "New lobby"})
	if reply := read(creator); reply.Error != "Room already exists" {
		t.Fatal(reply)
	}

	for _, name := range []string{strings.Repeat("a", 30), strings.Repeat("界", 30)} {
		write(creator, protocol.Message{Type: "create-room", Room: name})
		if reply := read(creator); reply.Type != "create-room" || reply.Room != name {
			t.Fatal("30-character room name rejected", reply)
		}
	}

	other := connect()
	write(other, protocol.Message{Type: "list-rooms"})
	if reply := read(other); reply.Type != "list-rooms" || !slices.Contains(reply.Rooms, "New lobby") {
		t.Fatal(reply)
	}
	write(other, protocol.Message{Type: "join-room", Room: "New lobby"})
	if reply := read(other); reply.Type != "join-room" || reply.Room != "New lobby" {
		t.Fatal(reply)
	}
	if reply := read(other); reply.Type != "room-users" || len(reply.Users) != 1 {
		t.Fatal(reply)
	}
}
