package engine

import (
	"strings"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

func TestGameChatUI(t *testing.T) {
	game, session, state := setupEditorGame(t, 1)
	match := game.matches
	game.Draw(ebiten.NewImage(1100, 700))
	if match.chat.GetWidget().Rect.Min.Y < match.editor.spent.GetWidget().Rect.Max.Y || match.chatInput.GetWidget().Rect.Max.Y > match.status.GetWidget().Rect.Min.Y {
		t.Fatal("game chat not between spending footer and readiness message")
	}
	state.Players[0].Ready = true
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "game-state", Game: &state}})
	match.pending = true
	match.setEnabled()
	if match.chatInput.GetWidget().Disabled {
		t.Fatal("chat disabled during setup action")
	}
	match.chatInput.SetText(" hello ")
	match.chatSend.Click()
	game.Update()
	select {
	case command := <-session.commands:
		if command.Type != "game-chat" || command.GameID != state.ID || command.Text != "hello" {
			t.Fatal(command)
		}
	default:
		t.Fatal("game chat send failed")
	}
	if !match.pending || match.chatInput.GetText() != "" {
		t.Fatal("chat changed action state or did not clear input")
	}
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "game-chat", GameID: state.ID, ClientID: 2, Name: "Guest", Text: "[hello]"}})
	text := match.chat.GetText()
	if len(text) < 12 || text[12:] != "Guest: [hello]" {
		t.Fatal("chat content missing", text)
	}
	if _, err := time.ParseInLocation("03:04:05 PM", text[:11], time.Local); err != nil {
		t.Fatal("chat timestamp missing", text)
	}
	if !match.active.Players[0].Ready {
		t.Fatal("chat changed readiness")
	}
	for _, message := range []protocol.Message{
		{Type: "game-chat", GameID: state.ID + 1, ClientID: 2, Text: "wrong game"},
		{Type: "game-chat", GameID: state.ID, ClientID: 99, Text: "nonmember"},
	} {
		game.handleEvent(lobbyEvent{session: session, message: message})
	}
	if match.chat.GetText() != text {
		t.Fatal("unrelated game/user chat displayed")
	}
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "game-state", Game: &state}})
	if match.chat.GetText() != text {
		t.Fatal("game state update cleared chat")
	}
	match.chatInput.SetText(strings.Repeat("x", 2001))
	match.chatSend.Click()
	game.Update()
	if len(session.commands) != 0 || !strings.Contains(match.status.Label, "2000") {
		t.Fatal("oversized game chat not rejected")
	}
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "game-left", GameID: state.ID}})
	if match.chat.GetText() != "" || match.chatInput.GetText() != "" || !match.chatInput.GetWidget().Disabled {
		t.Fatal("game exit left chat state")
	}
}
