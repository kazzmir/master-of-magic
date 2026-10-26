package multiplayer

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

type lockedLogBuffer struct {
	lock sync.Mutex
	data bytes.Buffer
}

func (buffer *lockedLogBuffer) Write(data []byte) (int, error) {
	buffer.lock.Lock()
	defer buffer.lock.Unlock()
	return buffer.data.Write(data)
}
func (buffer *lockedLogBuffer) String() string {
	buffer.lock.Lock()
	defer buffer.lock.Unlock()
	return buffer.data.String()
}

func TestLifecycleLogging(t *testing.T) {
	var captured lockedLogBuffer
	previous := log.Writer()
	log.SetOutput(&captured)
	t.Cleanup(func() { log.SetOutput(previous) })
	mux := http.NewServeMux()
	InitMux(mux, testAuth)
	server := httptest.NewServer(mux)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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
				t.Fatal(message.Error)
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
		return conn
	}
	host, guest := connect("LoggerHost"), connect("LoggerGuest")
	write(host, protocol.Message{Type: "create-room", Room: "Logging public"})
	readKind(host, "create-room")
	const password = "logging-test-room-password"
	write(host, protocol.Message{Type: "create-room", Room: "Logging private", Private: true, Password: password})
	readKind(host, "create-room")
	for _, conn := range []*websocket.Conn{host, guest} {
		write(conn, protocol.Message{Type: "join-room", Room: "Logging private", Password: password})
		readKind(conn, "join-room")
	}
	write(host, protocol.Message{Type: "create-game"})
	id := readKind(host, "game-state").Game.ID
	write(guest, protocol.Message{Type: "join-game", GameID: id})
	readKind(host, "game-state")
	readKind(guest, "game-state")
	for _, conn := range []*websocket.Conn{host, guest} {
		write(conn, protocol.Message{Type: "game-ready", GameID: id, Ready: true})
	}
	first, second := readKind(host, "combat-start"), readKind(guest, "combat-start")
	write(host, protocol.Message{Type: "leave-room"})
	readKind(host, "leave-room")
	text := captured.String()
	for _, expected := range []string{
		`Client 1 ("LoggerHost") connected`,
		`Client 2 ("LoggerGuest") connected`,
		`created room "Logging public" (public, password_protected=false)`,
		`created room "Logging private" (private, password_protected=true)`,
		`Client 1 ("LoggerHost") joined room "Logging private"`,
		`Client 2 ("LoggerGuest") joined room "Logging private"`,
		`Client 1 ("LoggerHost") left room "Logging private"`,
		`starting in room "Logging private": attacker client=1 ("LoggerHost"), defender client=2 ("LoggerGuest"), creator client=1`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatal("missing lifecycle log", expected)
		}
	}
	for _, secret := range []string{testAuth, password, first.RelayToken, second.RelayToken} {
		if secret != "" && strings.Contains(text, secret) {
			t.Fatal("sensitive value included in server logs")
		}
	}
}
