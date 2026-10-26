package net

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func TestWebsocketPacketDeadlinesAndConcurrentSend(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.CloseNow()
		packet := NewPacketNetConn(websocket.NetConn(r.Context(), conn, websocket.MessageBinary), 65536)
		for {
			data, err := packet.Read()
			if err != nil {
				return
			}
			if packet.Send(data) != nil {
				return
			}
		}
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	peer := NewWebsocketPacketConn(ctx, conn, 65536)
	defer peer.Close()
	peer.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
	if _, err := peer.Read(); err == nil {
		t.Fatal("deadline not enforced")
	}
	peer.SetReadDeadline(time.Time{})
	var wait sync.WaitGroup
	errors := make(chan error, 10)
	for i := range 10 {
		wait.Go(func() { errors <- peer.Send([]byte{byte(i), 42, 43}) })
	}
	wait.Wait()
	for range 10 {
		if err := <-errors; err != nil {
			t.Fatal(err)
		}
	}
	seen := make(map[byte]bool)
	for range 10 {
		data, err := peer.Read()
		if err != nil || len(data) != 3 || data[1] != 42 || data[2] != 43 {
			t.Fatal("packet framing corrupted", data, err)
		}
		seen[data[0]] = true
	}
	if len(seen) != 10 {
		t.Fatal("concurrent packets lost")
	}
	cancel()
	select {
	case <-peer.Done():
	case <-time.After(time.Second):
		t.Fatal("transport leaked after cancellation")
	}
}
