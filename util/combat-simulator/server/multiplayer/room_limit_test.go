package multiplayer

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

func TestTotalRoomLimit(t *testing.T) {
	manager := &Manager{authKey: testAuth, Lobbies: map[string]*Lobby{"Public": {Public: true, Name: "Public", persistent: true, Users: make(map[uint64]*User)}}}
	// Mixed public/private rooms and the permanent Public room all count.
	for i := 1; i < MaxRooms-1; i++ {
		name := fmt.Sprintf("Room%d", i)
		manager.Lobbies[name] = &Lobby{Name: name, Public: i%2 == 0, Users: make(map[uint64]*User)}
	}
	var counter atomic.Uint64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.CloseNow()
		handleWebsocket(conn, manager, counter.Add(1))
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
	var clients [2]*websocket.Conn
	for i := range clients {
		conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http"), nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.CloseNow()
		clients[i] = conn
		write(conn, map[string]string{"auth": testAuth})
		read(conn)
		read(conn)
	}
	// Two connections compete for the final slot; only one can succeed.
	for i, conn := range clients {
		write(conn, protocol.Message{Type: "create-room", Room: fmt.Sprintf("Final%d", i), Private: i == 1})
	}
	accepted, rejected := 0, 0
	for _, conn := range clients {
		reply := read(conn)
		if reply.Type == "create-room" {
			accepted++
		} else if strings.Contains(reply.Error, "1000") {
			rejected++
		} else {
			t.Fatal(reply)
		}
	}
	if accepted != 1 || rejected != 1 {
		t.Fatal("room cap not enforced across connections", accepted, rejected)
	}
	for i, conn := range clients {
		write(conn, protocol.Message{Type: "create-room", Room: fmt.Sprintf("Extra%d", i), Private: i == 0})
		if reply := read(conn); !strings.Contains(reply.Error, "1000") {
			t.Fatal("room created over cap", reply)
		}
	}
	count, _ := manager.lock.With2(func() (int, error) { return len(manager.Lobbies), nil })
	if count != MaxRooms {
		t.Fatal("incorrect total room count", count)
	}
}
