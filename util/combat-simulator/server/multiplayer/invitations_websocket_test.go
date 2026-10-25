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

func TestInvitationWebsocketFlow(t *testing.T) {
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
	readKind := func(conn *websocket.Conn, kind string) protocol.Message {
		t.Helper()
		for {
			var message protocol.Message
			if err := wsjson.Read(ctx, conn, &message); err != nil {
				t.Fatal(err)
			}
			if message.Error != "" && kind != "" {
				t.Fatal(message)
			}
			if message.Type == kind {
				return message
			}
		}
	}
	connect := func(name string) (*websocket.Conn, uint64) {
		t.Helper()
		conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/multiplayer", nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { conn.CloseNow() })
		write(conn, map[string]string{"auth": "x", "name": name})
		id := readKind(conn, "connected").ClientID
		readKind(conn, "version")
		write(conn, protocol.Message{Type: "version", Version: 1})
		write(conn, protocol.Message{Type: "join-room", Room: "Public"})
		readKind(conn, "join-room")
		return conn, id
	}
	requester, requesterID := connect("Host")
	target, targetID := connect("Guest")
	write(requester, protocol.Message{Type: "request-game", TargetID: targetID})
	sent := readKind(requester, "invite-sent")
	invite := readKind(target, "game-invite")
	if invite.InviteID != sent.InviteID || invite.ClientID != requesterID || invite.Name != "Host" {
		t.Fatal(invite)
	}
	write(target, protocol.Message{Type: "respond-game-invite", InviteID: invite.InviteID, Accept: true})
	var gameID uint64
	for _, conn := range []*websocket.Conn{requester, target} {
		readKind(conn, "invite-closed")
		state := readKind(conn, "game-state").Game
		if state == nil || len(state.Players) != 2 || state.HostID != requesterID || state.Players[1].ClientID != targetID {
			t.Fatal(state)
		}
		if gameID != 0 && gameID != state.ID {
			t.Fatal("players received different games")
		}
		gameID = state.ID
	}
	other, _ := connect("Other")
	write(other, protocol.Message{Type: "request-game", TargetID: targetID})
	if reply := readKind(other, ""); !strings.Contains(reply.Error, "Guest is already in a game") {
		t.Fatal(reply)
	}
	requester.CloseNow()
	readKind(target, "game-left")
	write(other, protocol.Message{Type: "request-game", TargetID: targetID})
	readKind(other, "invite-sent")
	secondInvite := readKind(target, "game-invite")
	other.CloseNow()
	if closed := readKind(target, "invite-closed"); closed.InviteID != secondInvite.InviteID || !strings.Contains(closed.Text, "left the room") {
		t.Fatal(closed)
	}
}
