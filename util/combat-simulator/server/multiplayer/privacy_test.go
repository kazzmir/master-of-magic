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

func TestPrivateAndPasswordRooms(t *testing.T) {
	mux := http.NewServeMux()
	InitMux(mux, testAuth)
	server := httptest.NewServer(mux)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
		if message.Type != kind || message.Password != "" {
			t.Fatalf("unexpected response: %+v", message)
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
		write(conn, map[string]string{"auth": testAuth, "name": "Player1234"})
		read(conn, "connected")
		if version := read(conn, "version"); version.Version != 1 {
			t.Fatal(version)
		}
		write(conn, protocol.Message{Type: "version", Version: protocol.ClientVersion})
		return conn
	}
	creator := connect()
	rooms := []protocol.Message{
		{Type: "create-room", Room: "Open"},
		{Type: "create-room", Room: "Hidden", Private: true},
		{Type: "create-room", Room: "Protected", Password: "secret"},
		{Type: "create-room", Room: "Hidden protected", Private: true, Password: "different secret"},
	}
	for _, command := range rooms {
		write(creator, command)
		if reply := read(creator, "create-room"); reply.Room != command.Room {
			t.Fatal(reply)
		}
	}
	write(creator, protocol.Message{Type: "create-room", Room: "Too long password", Password: strings.Repeat("x", 73)})
	if reply := read(creator, ""); reply.Error == "" {
		t.Fatal("oversized password accepted")
	}

	guest := connect()
	write(guest, protocol.Message{Type: "list-rooms"})
	listed := read(guest, "list-rooms")
	for _, room := range rooms {
		if slices.Contains(listed.Rooms, room.Room) == room.Private {
			t.Fatal("incorrect public room visibility", room.Room, listed.Rooms)
		}
	}
	lastRoom := ""
	for _, room := range rooms {
		if room.Password != "" {
			for _, password := range []string{"", "wrong", room.Password + " "} {
				write(guest, protocol.Message{Type: "join-room", Room: room.Room, Password: password})
				if reply := read(guest, ""); reply.Error != "Incorrect room password" {
					t.Fatal(reply)
				}
				// A failed join must not evict the guest from their current room.
				if lastRoom != "" {
					write(guest, protocol.Message{Type: "chat", Text: "still here"})
					if reply := read(guest, "chat"); reply.Room != lastRoom {
						t.Fatal(reply)
					}
				}
			}
		}
		write(guest, protocol.Message{Type: "join-room", Room: room.Room, Password: room.Password})
		if reply := read(guest, "join-room"); reply.Room != room.Room {
			t.Fatal(reply)
		}
		if reply := read(guest, "room-users"); len(reply.Users) != 1 {
			t.Fatal(reply)
		}
		lastRoom = room.Room
	}
}
