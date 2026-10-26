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

const testAuth = "multiplayer-test-key"

func TestSharedKeyAuthentication(t *testing.T) {
	for _, test := range []struct {
		name, expected, supplied string
		accepted                 bool
	}{
		{"matching", testAuth, testAuth, true},
		{"trim-config", "\n" + testAuth + "\n", testAuth, true},
		{"old-placeholder", testAuth, "x", false},
		{"incorrect", testAuth, "wrong-password", false},
		{"missing", testAuth, "", false},
		{"empty-config", "", "", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			mux := http.NewServeMux()
			InitMux(mux, test.expected)
			server := httptest.NewServer(mux)
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/multiplayer", nil)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.CloseNow()
			if err := wsjson.Write(ctx, conn, map[string]string{"auth": test.supplied}); err != nil {
				t.Fatal(err)
			}
			if !test.accepted {
				_, _, err := conn.Read(ctx)
				if websocket.CloseStatus(err) != websocket.StatusPolicyViolation {
					t.Fatal("invalid auth was not rejected before sending connected/version", err)
				}
				return
			}
			for _, kind := range []string{"connected", "version"} {
				var message protocol.Message
				if err := wsjson.Read(ctx, conn, &message); err != nil {
					t.Fatal(err)
				}
				if message.Type != kind {
					t.Fatal("unexpected authenticated response", message.Type)
				}
			}
		})
	}
}
