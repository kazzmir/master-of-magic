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

func TestMultiplayerLobby(t *testing.T) {
	mux := http.NewServeMux()
	InitMux(mux, testAuth)
	server := httptest.NewServer(mux)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	connect := func(name string) *websocket.Conn {
		t.Helper()
		conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/multiplayer", nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { conn.CloseNow() })
		if err := wsjson.Write(ctx, conn, map[string]string{"auth": testAuth, "name": name}); err != nil {
			t.Fatal(err)
		}
		return conn
	}
	write := func(conn *websocket.Conn, message protocol.Message) {
		t.Helper()
		if err := wsjson.Write(ctx, conn, message); err != nil {
			t.Fatal(err)
		}
	}
	read := func(conn *websocket.Conn, kind string) protocol.Message {
		t.Helper()
		var message protocol.Message
		if err := wsjson.Read(ctx, conn, &message); err != nil {
			t.Fatal(err)
		}
		if message.Type != kind {
			t.Fatalf("expected %s, got %+v", kind, message)
		}
		return message
	}
	first := connect("Alice")
	firstConnected := read(first, "connected")
	if version := read(first, "version"); version.Version != 1 {
		t.Fatal(version)
	}
	write(first, protocol.Message{Type: "version", Version: protocol.ClientVersion})
	firstID := firstConnected.ClientID
	if firstConnected.Name != "Alice" {
		t.Fatal(firstConnected)
	}
	if firstID == 0 {
		t.Fatal("missing assigned client ID")
	}
	write(first, protocol.Message{Type: "list-rooms", ID: 12})
	rooms := read(first, "list-rooms")
	if rooms.ID != 12 || len(rooms.Rooms) != 1 || rooms.Rooms[0] == "" {
		t.Fatal(rooms)
	}
	room := rooms.Rooms[0]
	write(first, protocol.Message{Type: "join-room", Room: "missing"})
	if message := read(first, ""); message.Error == "" {
		t.Fatal("missing room accepted")
	}
	write(first, protocol.Message{Type: "join-room", Room: room})
	joined := read(first, "join-room")
	if joined.Room != room || joined.ClientID != firstID {
		t.Fatal(joined)
	}
	if users := read(first, "room-users"); !slices.Equal(users.Users, []uint64{firstID}) {
		t.Fatal(users)
	}

	second := connect("Bob")
	secondID := read(second, "connected").ClientID
	if version := read(second, "version"); version.Version != 1 {
		t.Fatal(version)
	}
	// A different version is accepted until compatibility checks are defined.
	write(second, protocol.Message{Type: "version", Version: 999})
	if secondID == 0 || secondID == firstID {
		t.Fatal("client IDs must be unique", firstID, secondID)
	}
	write(second, protocol.Message{Type: "set-name", Name: "Bob outside"})
	if updated := read(second, "name-updated"); updated.ClientID != secondID || updated.Name != "Bob outside" || updated.Room != "" {
		t.Fatal(updated)
	}
	write(second, protocol.Message{Type: "chat", Text: "not in a room"})
	if message := read(second, ""); message.Error == "" {
		t.Fatal("chat outside a room accepted")
	}
	write(second, protocol.Message{Type: "join-room", Room: room})
	secondJoined := read(second, "join-room")
	if secondJoined.ClientID != secondID {
		t.Fatal(secondJoined)
	}
	if users := read(second, "room-users"); !slices.Equal(users.Users, []uint64{firstID, secondID}) || users.Names[firstID] != "Alice" || users.Names[secondID] != "Bob outside" {
		t.Fatal(users)
	}
	if event := read(first, "user-joined"); event.ClientID != secondID || event.Name != "Bob outside" || event.Room != room {
		t.Fatal(event)
	}
	if users := read(first, "room-users"); !slices.Equal(users.Users, []uint64{firstID, secondID}) {
		t.Fatal(users)
	}

	// A client cannot impersonate another client by supplying their ID.
	write(first, protocol.Message{Type: "chat", Text: "hello", ClientID: secondID})
	for _, conn := range []*websocket.Conn{first, second} {
		message := read(conn, "chat")
		if message.Text != "hello" || message.ClientID != firstID || message.Room != room {
			t.Fatal(message)
		}
	}
	write(second, protocol.Message{Type: "set-name", Name: "Bob renamed", ClientID: firstID})
	for _, conn := range []*websocket.Conn{first, second} {
		updated := read(conn, "name-updated")
		if updated.ClientID != secondID || updated.Room != room || updated.Name != "Bob renamed" {
			t.Fatal(updated)
		}
	}
	write(second, protocol.Message{Type: "set-name", Name: "   "})
	if reply := read(second, ""); reply.Error == "" {
		t.Fatal("blank name accepted")
	}
	write(second, protocol.Message{Type: "leave-room", ID: 24})
	if reply := read(second, "leave-room"); reply.ID != 24 {
		t.Fatal(reply)
	}
	if left := read(first, "user-left"); left.Room != room || left.ClientID != secondID || len(left.Users) != 0 {
		t.Fatal(left)
	}
	write(second, protocol.Message{Type: "join-room", Room: room})
	read(second, "join-room")
	if users := read(second, "room-users"); users.Names[secondID] != "Bob renamed" {
		t.Fatal(users)
	}
	if event := read(first, "user-joined"); event.ClientID != secondID || event.Name != "Bob renamed" || event.Room != room {
		t.Fatal(event)
	}
	read(first, "room-users")
	second.CloseNow()
	if left := read(first, "user-left"); left.Room != room || left.ClientID != secondID || len(left.Users) != 0 {
		t.Fatal(left)
	}
}

func TestMissingAuth(t *testing.T) {
	mux := http.NewServeMux()
	InitMux(mux, testAuth)
	server := httptest.NewServer(mux)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/multiplayer", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	if err := wsjson.Write(ctx, conn, protocol.Message{Type: "list-rooms"}); err != nil {
		t.Fatal(err)
	}
	_, _, err = conn.Read(ctx)
	if websocket.CloseStatus(err) != websocket.StatusPolicyViolation {
		t.Fatal(err)
	}
}
