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

func TestAcceptAnyOrigin(t *testing.T) {
	mux := http.NewServeMux()
	InitMux(mux)
	server := httptest.NewServer(mux)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/multiplayer", &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": []string{"http://localhost:8080"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	if err := wsjson.Write(ctx, conn, map[string]string{"auth": "x"}); err != nil {
		t.Fatal(err)
	}
	var connected protocol.Message
	if err := wsjson.Read(ctx, conn, &connected); err != nil {
		t.Fatal(err)
	}
	if connected.Type != "connected" {
		t.Fatal(connected)
	}
}
