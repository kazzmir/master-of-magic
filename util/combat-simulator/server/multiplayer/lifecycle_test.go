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

func TestDeleteEmptyRooms(t *testing.T) {
	mux := http.NewServeMux()
	InitMux(mux)
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
	read := func(conn *websocket.Conn, kind string) protocol.Message {
		t.Helper()
		var message protocol.Message
		if err := wsjson.Read(ctx, conn, &message); err != nil {
			t.Fatal(err)
		}
		if message.Type != kind {
			t.Fatal(message)
		}
		return message
	}
	connect := func() *websocket.Conn {
		t.Helper()
		conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/multiplayer", nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { conn.CloseNow() })
		write(conn, map[string]string{"auth": "x"})
		read(conn, "connected")
		return conn
	}
	observer, member := connect(), connect()
	join := func(room, password string) {
		t.Helper()
		write(member, protocol.Message{Type: "join-room", Room: room, Password: password})
		read(member, "join-room")
		if reply := read(member, "room-users"); len(reply.Users) != 1 {
			t.Fatal(reply)
		}
	}
	listed := func() []string {
		t.Helper()
		write(observer, protocol.Message{Type: "list-rooms"})
		return read(observer, "list-rooms").Rooms
	}
	join("Public", "")
	write(member, protocol.Message{Type: "leave-room"})
	read(member, "leave-room")
	if !slices.Contains(listed(), "Public") {
		t.Fatal("original Public room deleted when empty")
	}
	join("Public", "")
	write(member, protocol.Message{Type: "create-room", Room: "Temporary"})
	read(member, "create-room")
	join("Temporary", "")
	join("Temporary", "") // Rejoining must not delete the room or orphan membership.
	if !slices.Contains(listed(), "Temporary") {
		t.Fatal("rejoining deleted room")
	}
	write(member, protocol.Message{Type: "leave-room"})
	read(member, "leave-room")
	if slices.Contains(listed(), "Temporary") {
		t.Fatal("last explicit departure left room behind")
	}
	write(member, protocol.Message{Type: "join-room", Room: "Temporary"})
	if reply := read(member, ""); reply.Error != "Room not found" {
		t.Fatal(reply)
	}

	// Deleted names can be reused, with new password requirements.
	write(member, protocol.Message{Type: "create-room", Room: "Temporary", Password: "new password"})
	read(member, "create-room")
	write(member, protocol.Message{Type: "join-room", Room: "Temporary"})
	if reply := read(member, ""); reply.Error != "Incorrect room password" {
		t.Fatal(reply)
	}
	join("Temporary", "new password")
	write(member, protocol.Message{Type: "create-room", Room: "Second"})
	read(member, "create-room")
	join("Second", "")
	if rooms := listed(); slices.Contains(rooms, "Temporary") || !slices.Contains(rooms, "Public") {
		t.Fatal("incorrect room persistence after switching", rooms)
	}
	member.CloseNow()
	for slices.Contains(listed(), "Second") {
		select {
		case <-ctx.Done():
			t.Fatal("disconnect did not delete room", ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
}
