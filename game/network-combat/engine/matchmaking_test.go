package engine

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

func TestMatchmakingUI(t *testing.T) {
	game := newLobbyGame()
	if !game.matches.create.GetWidget().Disabled {
		t.Fatal("game creation enabled outside room")
	}
	session := &lobbySession{cancel: func() {}, commands: make(chan protocol.Message, 8)}
	game.session = session
	deliver := func(message protocol.Message) { game.handleEvent(lobbyEvent{session: session, message: message}) }
	deliver(protocol.Message{Type: "connected", ClientID: 1})
	deliver(protocol.Message{Type: "join-room", Room: "Public"})
	game.Update()
	match := game.matches
	match.create.Click()
	game.Update()
	select {
	case command := <-session.commands:
		if command.Type != "create-game" {
			t.Fatal(command)
		}
	default:
		t.Fatal("create game did not send a command")
	}
	if !match.create.GetWidget().Disabled {
		t.Fatal("creation not disabled while pending")
	}
	state := protocol.GameState{ID: 7, Room: "Public", HostID: 1, Players: []protocol.GamePlayer{{ClientID: 1, Name: "Host"}}}
	deliver(protocol.Message{Type: "game-state", Game: &state})
	if !game.ui.IsWindowOpen(match.window) || match.active.ID != 7 || !match.create.GetWidget().Disabled {
		t.Fatal("host game window missing")
	}
	if match.indicators[0].Image != match.notReadyIcon || match.readiness[0].Label != "not ready" {
		t.Fatal("not-ready player missing red X")
	}
	if match.indicators[1].GetWidget().IsVisible() {
		t.Fatal("empty player slot has readiness icon")
	}
	game.Update()
	game.Draw(ebiten.NewImage(1100, 700))
	location := image.Rect(10, 20, 610, 360)
	match.window.SetLocation(location)
	state.Players = append(state.Players, protocol.GamePlayer{ClientID: 2, Name: "Guest"})
	deliver(protocol.Message{Type: "game-state", Game: &state})
	if match.window.GetContainer().GetWidget().Rect != location {
		t.Fatal("state update recreated game window")
	}
	match.ready.Click()
	game.Update()
	select {
	case command := <-session.commands:
		if command.Type != "game-ready" || command.GameID != 7 || !command.Ready {
			t.Fatal(command)
		}
	default:
		t.Fatal("Ready did not send command")
	}
	state.Players[0].Ready, state.Players[1].Ready, state.Started = true, true, true
	deliver(protocol.Message{Type: "game-state", Game: &state})
	for i := range match.indicators {
		if match.indicators[i].Image != match.readyIcon || match.readiness[i].Label != "ready" || !match.indicators[i].GetWidget().IsVisible() {
			t.Fatal("ready player missing green check mark")
		}
	}
	if !match.ready.GetWidget().Disabled || match.status.Label != "Both players ready — connecting to combat…" {
		t.Fatal("start placeholder not shown")
	}
	match.exit.Click()
	game.Update()
	select {
	case command := <-session.commands:
		if command.Type != "leave-game" || command.GameID != 7 {
			t.Fatal(command)
		}
	default:
		t.Fatal("Exit did not send command")
	}
	deliver(protocol.Message{Type: "game-left", GameID: 7})
	if game.ui.IsWindowOpen(match.window) || match.active.ID != 0 || match.create.GetWidget().Disabled {
		t.Fatal("game exit did not release UI")
	}
}

func TestWaitingGameClickAndCleanup(t *testing.T) {
	game := newLobbyGame()
	session := &lobbySession{cancel: func() {}, commands: make(chan protocol.Message, 8)}
	game.session = session
	deliver := func(message protocol.Message) { game.handleEvent(lobbyEvent{session: session, message: message}) }
	deliver(protocol.Message{Type: "connected", ClientID: 2})
	state := protocol.GameState{ID: 9, Room: "Public", HostID: 1, Players: []protocol.GamePlayer{{ClientID: 1, Name: "Host"}}}
	deliver(protocol.Message{Type: "join-room", Room: "Public", Games: []protocol.GameState{state}})
	game.Update()
	match := game.matches
	if len(match.waiting) != 1 {
		t.Fatal("initial game list missing")
	}
	match.list.SetSelectedEntry(uint64(9))
	game.Update()
	select {
	case command := <-session.commands:
		if command.Type != "join-game" || command.GameID != 9 {
			t.Fatal(command)
		}
	default:
		t.Fatal("game selection did not join")
	}
	deliver(protocol.Message{Error: "Game is full"})
	if match.pending || match.list.GetWidget().Disabled {
		t.Fatal("failed join left controls pending")
	}
	state.Players = append(state.Players, protocol.GamePlayer{ClientID: 2, Name: "Guest"})
	deliver(protocol.Message{Type: "game-state", Game: &state})
	game.disconnect("Disconnected")
	if match.active.ID != 0 || game.ui.IsWindowOpen(match.window) || len(match.waiting) != 0 {
		t.Fatal("disconnect left stale game state")
	}
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "game-state", Game: &state}})
	if match.active.ID != 0 {
		t.Fatal("stale session reopened game")
	}
}
