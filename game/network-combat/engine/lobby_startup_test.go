package engine

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestLobbyDisconnectBeforeFirstUpdate(t *testing.T) {
	game := newLobbyGame()
	game.server.SetText("invalid://server")
	game.connect()
	game.disconnect("Disconnected")
	if game.session != nil || game.connected {
		t.Fatal("disconnect did not reset session")
	}
	if err := game.Update(); err != nil {
		t.Fatal(err)
	}
	if game.toggle.Text().Label != "Connect" {
		t.Fatal("incorrect button label after early disconnect")
	}
}

func TestLobbyOfflineStartup(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()

	game := newLobbyGame()
	game.server.SetText(address)
	// Launch connects before the UI's first Update/validation.
	game.connect()
	defer game.disconnect("Disconnected")
	select {
	case event := <-game.events:
		if event.err == nil {
			t.Fatal("expected connection failure")
		}
		game.handleEvent(event)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for offline server failure")
	}
	if game.session != nil || game.connected || game.server.GetWidget().Disabled {
		t.Fatal("offline startup did not restore disconnected state")
	}
	if !strings.HasPrefix(game.status.Label, "Connection closed:") {
		t.Fatal(game.status.Label)
	}
	if err := game.Update(); err != nil {
		t.Fatal(err)
	}
	if game.toggle.Text().Label != "Connect" {
		t.Fatal("incorrect button label after startup failure")
	}
}
