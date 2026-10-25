package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/kazzmir/master-of-magic/game/network-combat/engine/crypto"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
	netlib "github.com/kazzmir/master-of-magic/lib/net"
)

func TestConnectCombat(t *testing.T) {
	failures := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/multiplayer/combat" {
			t.Error(r.URL.Path)
		}
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			failures <- err
			return
		}
		defer conn.CloseNow()
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		var hello protocol.Message
		if err := wsjson.Read(ctx, conn, &hello); err != nil {
			failures <- err
			return
		}
		if hello.GameID != 7 || hello.RelayToken != "capability" {
			t.Error("missing combat authorization", hello)
		}
		if err := wsjson.Write(ctx, conn, protocol.Message{Type: "relay-ready"}); err != nil {
			failures <- err
			return
		}
		packet := netlib.NewWebsocketPacketConn(ctx, conn, MAX_PACKET_SIZE)
		defer packet.Close()
		encrypted, err := crypto.NewEncryptedConn(packet)
		if err != nil {
			failures <- err
			return
		}
		data, err := encrypted.Read()
		if err != nil {
			failures <- err
			return
		}
		failures <- encrypted.Send(data)
		encrypted.Read() // Keep the socket alive until client cleanup.
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	encrypted, packet, err := connectCombat(ctx, server.URL, protocol.Message{GameID: 7, RelayToken: "capability"})
	if err != nil {
		t.Fatal(err)
	}
	defer packet.Close()
	if err := encrypted.Send([]byte("combat test")); err != nil {
		t.Fatal(err)
	}
	data, err := encrypted.Read()
	if err != nil || string(data) != "combat test" {
		t.Fatal(string(data), err)
	}
	if err := <-failures; err != nil {
		t.Fatal(err)
	}
}
