package engine

import (
	"fmt"
	"image"
	"image/color"

	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

type matchmakingUI struct {
	game                *lobbyGame
	container           *widget.Container
	list                *widget.List
	create, ready, exit *widget.Button
	heading, status     *widget.Text
	players             [2]*widget.Text
	window              *widget.Window
	waiting             map[uint64]protocol.GameState
	active              protocol.GameState
	pending             bool
}

func newMatchmakingUI(game *lobbyGame, theme *widget.Theme) *matchmakingUI {
	match := &matchmakingUI{game: game}
	label := func(value string) *widget.Text { return widget.NewText(widget.TextOpts.Text(value, nil, color.White)) }
	button := func(value string, action func()) *widget.Button {
		return widget.NewButton(widget.ButtonOpts.Text(value, nil, nil),
			widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) { action() }))
	}
	match.container = lobbyColumn(0, 1)
	header := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(
		widget.GridLayoutOpts.Columns(2), widget.GridLayoutOpts.Weights([]float64{1, 0}, nil),
		widget.GridLayoutOpts.Spacing(8, 0),
	)))
	match.heading = label("Games waiting for players — join a room")
	match.create = button("Create game", func() { match.send(protocol.Message{Type: "create-game"}) })
	header.AddChild(match.heading, match.create)
	match.list = widget.NewList(
		widget.ListOpts.EntryLabelFunc(func(entry any) string {
			state := match.waiting[entry.(uint64)]
			host := clientLabel(state.HostID)
			for _, player := range state.Players {
				if player.ClientID == state.HostID {
					host = player.Name
				}
			}
			return fmt.Sprintf("Game #%d — %s (1/2) — click to join", state.ID, host)
		}),
		widget.ListOpts.AllowReselect(),
		widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
			if match.active.ID == 0 && !match.pending {
				match.send(protocol.Message{Type: "join-game", GameID: args.Entry.(uint64)})
			}
		}),
	)
	match.container.AddChild(header, match.list)

	contents := lobbyColumn(0, 0, 0, 1, 0, 0)
	contents.GetWidget().SetTheme(theme)
	contents.SetBackgroundImage(uiimage.NewNineSliceColor(color.NRGBA{R: 24, G: 28, B: 36, A: 255}))
	match.players[0], match.players[1] = label(""), label("")
	match.status = label("")
	match.ready = button("Ready", func() {
		ready := false
		for _, player := range match.active.Players {
			if player.ClientID == game.clientID {
				ready = player.Ready
			}
		}
		if match.active.ID != 0 && !match.active.Started {
			match.send(protocol.Message{Type: "game-ready", GameID: match.active.ID, Ready: !ready})
		}
	})
	match.exit = button("Exit", func() {
		if match.active.ID != 0 {
			match.send(protocol.Message{Type: "leave-game", GameID: match.active.ID})
		}
	})
	buttons := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(
		widget.RowLayoutOpts.Direction(widget.DirectionHorizontal), widget.RowLayoutOpts.Spacing(8),
	)))
	buttons.AddChild(match.ready, match.exit)
	contents.AddChild(match.players[0], match.players[1], label("Game setup"), label("Setup options will be added here."), match.status, buttons)
	title := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(
		widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(6)),
	)), widget.ContainerOpts.BackgroundImage(uiimage.NewNineSliceColor(color.NRGBA{R: 42, G: 50, B: 64, A: 255})))
	title.AddChild(label("Game setup"))
	match.window = widget.NewWindow(widget.WindowOpts.Contents(contents), widget.WindowOpts.TitleBar(title, 36),
		widget.WindowOpts.Draggable(), widget.WindowOpts.Resizeable(), widget.WindowOpts.MinSize(500, 300))
	return match
}

func (match *matchmakingUI) send(message protocol.Message) {
	if match.pending || !match.game.connected || match.game.room == "" {
		return
	}
	if match.game.send(message) {
		match.pending = true
		match.setEnabled()
	}
}

func (match *matchmakingUI) setEnabled() {
	available := match.game.connected && match.game.room != "" && !match.pending
	match.create.GetWidget().Disabled = !available || match.active.ID != 0
	match.list.GetWidget().Disabled = !available || match.active.ID != 0
	match.ready.GetWidget().Disabled = !available || match.active.ID == 0 || match.active.Started
	match.exit.GetWidget().Disabled = !available || match.active.ID == 0
}

func (match *matchmakingUI) updateList(states []protocol.GameState) {
	match.waiting = make(map[uint64]protocol.GameState)
	entries := make([]any, 0, len(states))
	for _, state := range states {
		if state.Room == match.game.room && !state.Started && len(state.Players) == 1 {
			match.waiting[state.ID] = state
			entries = append(entries, state.ID)
		}
	}
	match.list.SetEntries(entries)
	match.heading.Label = fmt.Sprintf("Games waiting for players (%d)", len(entries))
}

func (match *matchmakingUI) reset() {
	match.active = protocol.GameState{}
	match.pending = false
	match.window.Close()
	match.updateList(nil)
	match.setEnabled()
}

func (match *matchmakingUI) handleMessage(message protocol.Message) bool {
	switch message.Type {
	case "list-games":
		if message.Room == match.game.room && match.game.room != "" {
			match.updateList(message.Games)
		}
	case "game-left":
		if message.GameID == match.active.ID {
			match.active = protocol.GameState{}
			match.pending = false
			match.window.Close()
			match.setEnabled()
		}
	case "game-state":
		state := message.Game
		if state == nil || state.Room != match.game.room || match.game.room == "" {
			return true
		}
		member := false
		for _, player := range state.Players {
			member = member || player.ClientID == match.game.clientID
		}
		if !member {
			return true
		}
		match.active = *state
		match.pending = false
		match.ready.SetText("Ready")
		for i := range match.players {
			value := fmt.Sprintf("Player %d: waiting for a player…", i+1)
			if i < len(state.Players) {
				player := state.Players[i]
				status := "not ready"
				if player.Ready {
					status = "ready"
				}
				value = fmt.Sprintf("Player %d: %s — %s", i+1, player.Name, status)
				if player.ClientID == match.game.clientID && player.Ready {
					match.ready.SetText("Not ready")
				}
			}
			match.players[i].Label = value
		}
		match.status.Label = "Both players must be ready."
		if len(state.Players) < 2 {
			match.status.Label = "Waiting for a second player."
		}
		if state.Started {
			match.status.Label = "Both players ready — game start is not implemented yet."
		}
		if !match.game.ui.IsWindowOpen(match.window) {
			rect := match.game.ui.Container.GetWidget().Rect
			x, y := max(0, (rect.Dx()-600)/2), max(0, (rect.Dy()-340)/2)
			match.window.SetLocation(image.Rect(x, y, x+600, y+340))
			match.game.ui.AddWindow(match.window)
		}
		match.setEnabled()
	default:
		return false
	}
	return true
}
