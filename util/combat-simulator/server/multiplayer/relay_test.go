package multiplayer

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	peercrypto "github.com/kazzmir/master-of-magic/game/network-combat/engine/crypto"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
	netlib "github.com/kazzmir/master-of-magic/lib/net"
)

func TestCombatRelay(t *testing.T) {
	mux := http.NewServeMux()
	InitMux(mux)
	server := httptest.NewServer(mux)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	base := "ws" + strings.TrimPrefix(server.URL, "http")
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
	connectLobby := func() *websocket.Conn {
		t.Helper()
		conn, _, err := websocket.Dial(ctx, base+"/multiplayer", nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { conn.CloseNow() })
		write(conn, map[string]string{"auth": "x"})
		readKind(conn, "connected")
		readKind(conn, "version")
		write(conn, protocol.Message{Type: "join-room", Room: "Public"})
		readKind(conn, "join-room")
		return conn
	}
	host, guest := connectLobby(), connectLobby()
	write(host, protocol.Message{Type: "create-game"})
	id := readKind(host, "game-state").Game.ID
	write(guest, protocol.Message{Type: "join-game", GameID: id})
	readKind(host, "game-state")
	readKind(guest, "game-state")
	write(host, protocol.Message{Type: "game-ready", GameID: id, Ready: true})
	write(guest, protocol.Message{Type: "game-ready", GameID: id, Ready: true})
	starts := []protocol.Message{readKind(host, "combat-start"), readKind(guest, "combat-start")}
	if !starts[0].Attacker || starts[1].Attacker || starts[0].RelayToken == "" || starts[0].RelayToken == starts[1].RelayToken {
		t.Fatal("incorrect private combat capabilities/roles")
	}
	dialCombat := func(hello protocol.Message) *websocket.Conn {
		t.Helper()
		conn, _, err := websocket.Dial(ctx, base+"/multiplayer/combat", nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { conn.CloseNow() })
		write(conn, hello)
		return conn
	}
	invalid := dialCombat(protocol.Message{GameID: id, RelayToken: "wrong"})
	if _, _, err := invalid.Read(ctx); websocket.CloseStatus(err) != websocket.StatusPolicyViolation {
		t.Fatal("invalid token accepted", err)
	}
	first := dialCombat(starts[0])
	second := dialCombat(starts[1])
	readKind(first, "relay-ready")
	readKind(second, "relay-ready")
	duplicate := dialCombat(starts[0])
	if _, _, err := duplicate.Read(ctx); websocket.CloseStatus(err) != websocket.StatusPolicyViolation {
		t.Fatal("duplicate connection accepted", err)
	}
	packets := []*netlib.WebsocketPacketConn{netlib.NewWebsocketPacketConn(ctx, first, 1024*1024), netlib.NewWebsocketPacketConn(ctx, second, 1024*1024)}
	defer packets[0].Close()
	defer packets[1].Close()
	var encrypted [2]netlib.PacketConn
	var failures [2]error
	var wait sync.WaitGroup
	for i := range packets {
		wait.Go(func() { encrypted[i], failures[i] = peercrypto.NewEncryptedConn(packets[i]) })
	}
	wait.Wait()
	for _, err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	// A temporary read timeout must NOT destroy the websocket between combat's
	// army exchange and its long-running event receive loop.
	packets[0].SetReadDeadline(time.Now().Add(10 * time.Millisecond))
	if _, err := encrypted[0].Read(); err == nil {
		t.Fatal("read deadline ignored")
	}
	packets[0].SetReadDeadline(time.Time{})
	payload := bytes.Repeat([]byte("x"), combatBytesPerSecond)
	began := time.Now()
	for i := range encrypted {
		wait.Go(func() { failures[i] = encrypted[i].Send(payload) })
	}
	for i := range encrypted {
		data, err := encrypted[i].Read()
		if err != nil || !bytes.Equal(data, payload) {
			t.Fatal("encrypted relay roundtrip failed", err)
		}
	}
	wait.Wait()
	for _, err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	if time.Since(began) < 1900*time.Millisecond {
		t.Fatal("combined 5 KiB/s bandwidth cap exceeded", time.Since(began))
	}
	packets[0].Close()
	if message := readKind(guest, "game-left"); message.GameID != id {
		t.Fatal("relay disconnect did not close game")
	}
}

func TestRelayPacerCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var pacer relayPacer
	if err := pacer.wait(ctx, combatBytesPerSecond); err == nil {
		t.Fatal("cancelled relay kept waiting")
	}
}
