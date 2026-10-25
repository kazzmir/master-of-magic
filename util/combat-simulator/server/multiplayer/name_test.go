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

func TestServerUserNameLimits(t *testing.T) {
	for _, character := range []string{"a", "界"} {
		for _, phase := range []string{"auth", "rename"} {
			t.Run(phase+"/"+character, func(t *testing.T) {
				mux := http.NewServeMux()
				InitMux(mux)
				server := httptest.NewServer(mux)
				defer server.Close()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/multiplayer", nil)
				if err != nil {
					t.Fatal(err)
				}
				defer conn.CloseNow()
				write := func(message any) {
					t.Helper()
					if err := wsjson.Write(ctx, conn, message); err != nil {
						t.Fatal(err)
					}
				}
				read := func() protocol.Message {
					t.Helper()
					var message protocol.Message
					if err := wsjson.Read(ctx, conn, &message); err != nil {
						t.Fatal(err)
					}
					return message
				}
				tooLong := strings.Repeat(character, 31)
				if phase == "auth" {
					write(map[string]string{"auth": "x", "name": tooLong})
					_, _, err := conn.Read(ctx)
					if websocket.CloseStatus(err) != websocket.StatusPolicyViolation {
						t.Fatal("overlong initial name accepted", err)
					}
					return
				}
				write(map[string]string{"auth": "x", "name": "Original"})
				connected := read()
				if connected.Type != "connected" {
					t.Fatal(connected)
				}
				if version := read(); version.Type != "version" || version.Version != 1 {
					t.Fatal(version)
				}
				write(protocol.Message{Type: "version", Version: protocol.ClientVersion})
				write(protocol.Message{Type: "set-name", ID: 1, Name: tooLong})
				if reply := read(); reply.Error == "" || reply.ID != 1 {
					t.Fatal("overlong rename accepted", reply)
				}
				write(protocol.Message{Type: "join-room", Room: "Public"})
				if reply := read(); reply.Type != "join-room" {
					t.Fatal(reply)
				}
				if reply := read(); reply.Type != "room-users" || reply.Names[connected.ClientID] != "Original" {
					t.Fatal("rejected rename changed stored name", reply)
				}
			})
		}
	}
}

func TestUserNameLength(t *testing.T) {
	for _, character := range []string{"a", "界"} {
		if !validName(strings.Repeat(character, 30)) {
			t.Fatal("30-character name rejected")
		}
		if validName(strings.Repeat(character, 31)) {
			t.Fatal("31-character name accepted")
		}
	}
}
