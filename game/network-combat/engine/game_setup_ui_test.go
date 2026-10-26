package engine

import (
	"image"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

func setupEditorGame(t *testing.T, clientID uint64) (*lobbyGame, *lobbySession, protocol.GameState) {
	t.Helper()
	game := newLobbyGame()
	session := &lobbySession{cancel: func() {}, commands: make(chan protocol.Message, 8)}
	game.session = session
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "connected", ClientID: clientID}})
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "join-room", Room: "Public"}})
	state := protocol.GameState{ID: 12345678901, Room: "Public", HostID: 1, AttackerID: 1, Budget: 1000, Players: []protocol.GamePlayer{
		{ClientID: 1, Name: "Host", Loadout: protocol.DefaultLoadout(true)}, {ClientID: 2, Name: "Guest", Loadout: protocol.DefaultLoadout(false)},
	}}
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "game-state", Game: &state}})
	game.Update()
	return game, session, state
}

func TestArmyListStaysAboveSpendingFooter(t *testing.T) {
	game, _, state := setupEditorGame(t, 1)
	unit := state.Players[0].Loadout.Army[0]
	state.Players[0].Loadout.Army = make([]protocol.UnitID, protocol.MaxArmy)
	for i := range state.Players[0].Loadout.Army {
		state.Players[0].Loadout.Army[i] = unit
	}
	game.handleEvent(lobbyEvent{session: game.session, message: protocol.Message{Type: "game-state", Game: &state}})
	game.Update()
	for _, size := range []image.Point{{1000, 660}, {800, 580}, {1000, 500}} {
		game.matches.window.SetLocation(image.Rect(0, 0, size.X, size.Y))
		game.Draw(ebiten.NewImage(1100, 700))
		editor := game.matches.editor
		footerTop := editor.spent.GetWidget().Rect.Min.Y
		for _, list := range []*widget.List{editor.army, editor.spells} {
			if list.GetWidget().Rect.Max.Y > footerTop {
				t.Fatal("selection list overlaps spending footer", size, list.GetWidget().Rect, editor.spent.GetWidget().Rect)
			}
		}
	}
}

func TestMagicBookIcons(t *testing.T) {
	game, _, _ := setupEditorGame(t, 1)
	editor := game.matches.editor
	for i, icon := range editor.bookIcons {
		if icon == nil || icon.Image == nil || icon.Image.Bounds().Dx() != 24 || icon.Image.Bounds().Dy() != 24 {
			t.Fatal("missing magic book icon", protocol.MagicNames[i])
		}
		if len(icon.GetWidget().ToolTips) != 1 {
			t.Fatal("missing realm-name tooltip", protocol.MagicNames[i])
		}
	}
	if editor.imageCache != nil && !editor.bookIconsReady {
		t.Fatal("unable to load native magic symbols")
	}
	game.Draw(ebiten.NewImage(1100, 700))
}

func TestHostBudgetAndRoleEditor(t *testing.T) {
	game, session, state := setupEditorGame(t, 1)
	editor := game.matches.editor
	if editor.budget.GetWidget().Disabled {
		t.Fatal("creator budget field disabled")
	}
	game.Draw(ebiten.NewImage(1100, 700))
	frame := game.matches.window.GetContainer().GetWidget().Rect
	for _, control := range []widget.HasWidget{editor.budget, editor.mana, editor.books[4], editor.save, game.matches.ready} {
		if !control.GetWidget().Rect.In(frame) {
			t.Fatal("setup control outside floating window", control.GetWidget().Rect, frame)
		}
	}
	editor.budget.SetText(strconv.FormatUint(math.MaxUint64, 10))
	game.Update()
	editor.role.Click()
	game.Update()
	editor.apply.Click()
	game.Update()
	select {
	case command := <-session.commands:
		if command.Type != "game-settings" || command.Budget != math.MaxUint64 || command.AttackerID != 2 {
			t.Fatal(command)
		}
	default:
		t.Fatal("host settings not sent")
	}
	state.Budget, state.AttackerID, state.Revision = math.MaxUint64, 2, 1
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "game-state", Game: &state}})
	game.Update()
	if editor.settingsDirty || editor.budget.GetText() != strconv.FormatUint(math.MaxUint64, 10) {
		t.Fatal("settings acknowledgement not reflected")
	}
	if !strings.Contains(game.matches.players[0].Label, "defender") || !strings.Contains(game.matches.players[1].Label, "attacker") {
		t.Fatal("roles not displayed")
	}
}

func TestPlayerLoadoutEditor(t *testing.T) {
	game, session, state := setupEditorGame(t, 2)
	match := game.matches
	editor := match.editor
	if !editor.budget.GetWidget().Disabled || !editor.role.GetWidget().Disabled {
		t.Fatal("guest may change shared settings")
	}
	editor.mana.SetText("12")
	editor.books[0].SetText("1")
	editor.portrait.SetSelectedEntry(5)
	game.Update()
	editor.unitChoice.SetSelectedEntry(protocol.UnitID{LbxFile: "units2.lbx", Index: 17})
	game.Update()
	editor.addUnit.Click()
	game.Update()
	spellName := ""
	for _, spell := range protocol.SpellsCatalog() {
		if spell.Magic == "life" && spell.Cost <= 200 {
			spellName = spell.Name
			break
		}
	}
	if spellName == "" {
		t.Fatal("no Life spell in catalog")
	}
	editor.spellChoice.SetSelectedEntry(spellName)
	game.Update()
	editor.addSpell.Click()
	game.Update()
	if !editor.dirty || !match.ready.GetWidget().Disabled {
		t.Fatal("unsaved loadout can become ready")
	}
	editor.save.Click()
	game.Update()
	var saved protocol.Loadout
	select {
	case command := <-session.commands:
		if command.Type != "game-loadout" || command.Loadout == nil {
			t.Fatal(command)
		}
		saved = command.Loadout.Clone()
		if saved.Mana != 12 || saved.Books[0] != 1 || saved.Portrait != 5 || len(saved.Army) != 3 || len(saved.Spells) != 1 || saved.Spells[0] != spellName {
			t.Fatal("loadout selections missing", saved)
		}
	default:
		t.Fatal("loadout not sent")
	}
	state.Players[1].Loadout = saved
	state.Revision = 1
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "game-state", Game: &state}})
	game.Update()
	if editor.dirty || match.ready.GetWidget().Disabled {
		t.Fatal("saved loadout still blocked from ready")
	}
	match.ready.Click()
	game.Update()
	select {
	case command := <-session.commands:
		if command.Type != "game-ready" || !command.Ready || command.Revision != 1 {
			t.Fatal(command)
		}
	default:
		t.Fatal("ready not sent with setup revision")
	}
	state.Players[1].Ready = true
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "game-state", Game: &state}})
	game.Update()
	if !editor.mana.GetWidget().Disabled || match.ready.GetWidget().Disabled {
		t.Fatal("must unready before editing, but should still be able to unready")
	}
}
