package engine

import (
	"strings"
	"testing"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

func TestPlayerContextMenuRequestsGame(t *testing.T) {
	game := newLobbyGame()
	session := &lobbySession{cancel: func() {}, commands: make(chan protocol.Message, 4)}
	game.session = session
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "connected", ClientID: 1}})
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "join-room", Room: "Public"}})
	game.Update()
	invites := game.invitations
	if invites.contextMenu(1) != nil {
		t.Fatal("self invitation menu available")
	}
	menu := invites.contextMenu(2)
	menu.Children()[0].(*widget.Button).Click()
	game.Update()
	select {
	case command := <-session.commands:
		if command.Type != "request-game" || command.TargetID != 2 {
			t.Fatal(command)
		}
	default:
		t.Fatal("context menu did not request the clicked player")
	}
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Error: "Busy player is already in a game"}})
	if invites.requestPending || !strings.Contains(game.status.Label, "already in a game") || game.ui.IsWindowOpen(invites.window) {
		t.Fatal("busy target rejection not shown to requester")
	}
	invites.request(2)
	<-session.commands
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "invite-sent", InviteID: 5, TargetID: 2}})
	if invites.outgoingID != 5 || invites.requestPending {
		t.Fatal("outgoing request not tracked")
	}
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "invite-closed", InviteID: 5, Text: "Player declined the game request"}})
	if invites.outgoingID != 0 || game.status.Label != "Player declined the game request" {
		t.Fatal("decline was not displayed")
	}
}

func TestGameInvitationWindow(t *testing.T) {
	for _, accept := range []bool{false, true} {
		game := newLobbyGame()
		session := &lobbySession{cancel: func() {}, commands: make(chan protocol.Message, 4)}
		game.session = session
		deliver := func(message protocol.Message) { game.handleEvent(lobbyEvent{session: session, message: message}) }
		deliver(protocol.Message{Type: "connected", ClientID: 2})
		deliver(protocol.Message{Type: "join-room", Room: "Public"})
		game.Update()
		invites := game.invitations
		deliver(protocol.Message{Type: "game-invite", InviteID: 8, ClientID: 1, Room: "Other", Name: "Host"})
		if game.ui.IsWindowOpen(invites.window) {
			t.Fatal("request from another room displayed")
		}
		deliver(protocol.Message{Type: "game-invite", InviteID: 8, ClientID: 1, Room: "Public", Name: "Host"})
		if !game.ui.IsWindowOpen(invites.window) || !strings.Contains(invites.prompt.Label, "Host") {
			t.Fatal("invitation prompt missing")
		}
		game.Update()
		game.Draw(ebiten.NewImage(1100, 700))
		if accept {
			invites.accept.Click()
		} else {
			invites.decline.Click()
		}
		game.Update()
		select {
		case command := <-session.commands:
			if command.Type != "respond-game-invite" || command.InviteID != 8 || command.Accept != accept {
				t.Fatal(command)
			}
		default:
			t.Fatal("invitation response not sent")
		}
		if !invites.accept.GetWidget().Disabled || !invites.decline.GetWidget().Disabled {
			t.Fatal("response buttons not disabled while pending")
		}
		deliver(protocol.Message{Type: "invite-closed", InviteID: 8, Text: "Request resolved"})
		if game.ui.IsWindowOpen(invites.window) || invites.incomingID != 0 {
			t.Fatal("resolved prompt remained open")
		}
		if accept {
			state := protocol.GameState{ID: 11, Room: "Public", HostID: 1, Players: []protocol.GamePlayer{{ClientID: 1, Name: "Host"}, {ClientID: 2, Name: "Guest"}}}
			deliver(protocol.Message{Type: "game-state", Game: &state})
			if !game.ui.IsWindowOpen(game.matches.window) {
				t.Fatal("accepted request did not open new game")
			}
		}
		game.disconnect("Disconnected")
		if game.ui.IsWindowOpen(invites.window) || invites.outgoingID != 0 || invites.incomingID != 0 {
			t.Fatal("invitation state leaked on disconnect")
		}
	}
}
