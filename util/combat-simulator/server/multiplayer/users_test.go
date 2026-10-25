package multiplayer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

func TestConcurrentUserLimit(t *testing.T) {
	manager := &Manager{userSlots: MaxUsers - 1}
	var wait sync.WaitGroup
	var accepted atomic.Int32
	for range 20 {
		wait.Go(func() {
			if manager.reserveUserSlot() {
				accepted.Add(1)
			}
		})
	}
	wait.Wait()
	if accepted.Load() != 1 || manager.userSlots != MaxUsers {
		t.Fatal("global user limit exceeded", accepted.Load(), manager.userSlots)
	}
	manager.releaseUserSlot()
	if !manager.reserveUserSlot() {
		t.Fatal("released slot could not be reused")
	}
}

func userSlotCount(manager *Manager) int {
	count, _ := manager.lock.With2(func() (int, error) { return manager.userSlots, nil })
	return count
}

func TestUserLimitAdmissionAndCleanup(t *testing.T) {
	for _, authenticate := range []bool{false, true} {
		name := "pending-auth"
		if authenticate {
			name = "authenticated"
		}
		t.Run(name, func(t *testing.T) {
			manager := &Manager{authKey: testAuth, userSlots: MaxUsers - 1}
			done := make(chan struct{}, 4)
			var counter atomic.Uint64
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, err := websocket.Accept(w, r, nil)
				if err != nil {
					return
				}
				defer conn.CloseNow()
				handleWebsocket(conn, manager, counter.Add(1))
				done <- struct{}{}
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			dial := func() *websocket.Conn {
				t.Helper()
				conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http"), nil)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { conn.CloseNow() })
				return conn
			}
			auth := func(conn *websocket.Conn) {
				t.Helper()
				if err := wsjson.Write(ctx, conn, map[string]string{"auth": testAuth}); err != nil {
					t.Fatal(err)
				}
				var message protocol.Message
				if err := wsjson.Read(ctx, conn, &message); err != nil {
					t.Fatal(err)
				}
				if message.Type != "connected" {
					t.Fatal(message)
				}
			}
			waitDone := func() {
				t.Helper()
				select {
				case <-done:
				case <-ctx.Done():
					t.Fatal("user slot was not released", ctx.Err())
				}
			}
			first := dial()
			if authenticate {
				auth(first)
			}
			for userSlotCount(manager) != MaxUsers {
				select {
				case <-ctx.Done():
					t.Fatal("slot not reserved")
				case <-time.After(time.Millisecond):
				}
			}
			rejected := dial()
			if _, _, err := rejected.Read(ctx); websocket.CloseStatus(err) != websocket.StatusTryAgainLater {
				t.Fatal("extra user was not rejected", err)
			}
			waitDone()
			if userSlotCount(manager) != MaxUsers {
				t.Fatal("rejected connection changed slot count")
			}
			first.CloseNow()
			waitDone()
			if userSlotCount(manager) != MaxUsers-1 {
				t.Fatal("disconnect leaked user slot")
			}
			replacement := dial()
			auth(replacement)
			replacement.CloseNow()
			waitDone()
			if userSlotCount(manager) != MaxUsers-1 {
				t.Fatal("replacement leaked user slot")
			}
		})
	}
}

func TestFailedAuthenticationReleasesUserSlot(t *testing.T) {
	manager := &Manager{authKey: testAuth}
	done := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.CloseNow()
		handleWebsocket(conn, manager, 1)
		close(done)
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	if err := wsjson.Write(ctx, conn, map[string]string{"auth": "wrong"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := conn.Read(ctx); websocket.CloseStatus(err) != websocket.StatusPolicyViolation {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if userSlotCount(manager) != 0 {
		t.Fatal("failed authentication leaked user slot")
	}
}
