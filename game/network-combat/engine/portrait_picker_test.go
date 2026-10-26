package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

func TestPortraitGridPicker(t *testing.T) {
	game, session, state := setupEditorGame(t, 1)
	editor := game.matches.editor
	original := editor.draft.Portrait
	choice := (original + 1) % len(protocol.PortraitNames)
	editor.portraitButton.Click()
	game.Update()
	window := editor.portraitWindow
	if window == nil || !game.ui.IsWindowOpen(window) {
		t.Fatal("clicking current portrait did not open picker")
	}
	if len(editor.portraitChoices) != len(protocol.PortraitNames) || len(editor.portraitGrid.Children()) != len(protocol.PortraitNames) {
		t.Fatal("picker does not show all portraits")
	}
	game.Draw(ebiten.NewImage(1100, 700))
	first, second, third, fourth := editor.portraitChoices[0].GetWidget().Rect, editor.portraitChoices[1].GetWidget().Rect, editor.portraitChoices[2].GetWidget().Rect, editor.portraitChoices[3].GetWidget().Rect
	if first.Min.Y != second.Min.Y || second.Min.Y != third.Min.Y || first.Min.X >= second.Min.X || second.Min.X >= third.Min.X || fourth.Min.Y <= first.Min.Y || fourth.Min.X != first.Min.X {
		t.Fatal("portrait picker is not a three-column grid", first, second, third, fourth)
	}
	editor.openPortraitPicker()
	if editor.portraitWindow != window {
		t.Fatal("duplicate picker created")
	}
	editor.portraitChoices[choice].Click()
	game.Update()
	if game.ui.IsWindowOpen(window) || editor.draft.Portrait != choice || !editor.dirty {
		t.Fatal("selection did not update draft and close picker")
	}
	if editor.portraitButton.portrait != editor.images[choice] {
		t.Fatal("current portrait did not update")
	}
	if len(session.commands) != 0 {
		t.Fatal("portrait selection bypassed Save loadout")
	}
	editor.portraitButton.Click()
	game.Update()
	if !game.ui.IsWindowOpen(window) || editor.portraitWindow != window {
		t.Fatal("picker did not reopen")
	}
	state.Players[0].Loadout.Portrait = choice
	state.Players[0].Ready = true
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "game-state", Game: &state}})
	if game.ui.IsWindowOpen(window) || !editor.portraitButton.GetWidget().Disabled {
		t.Fatal("ready player can still select portrait")
	}
}

func TestPortraitPickerClosesOnGameExit(t *testing.T) {
	game, session, state := setupEditorGame(t, 1)
	editor := game.matches.editor
	editor.portraitButton.Click()
	game.Update()
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "game-left", GameID: state.ID}})
	if game.ui.IsWindowOpen(editor.portraitWindow) {
		t.Fatal("portrait picker outlived the game")
	}
}
