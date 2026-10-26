package engine

import (
	"fmt"
	"image"
	"image/color"

	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

type matchmakingUI struct {
	game                    *lobbyGame
	container               *widget.Container
	list                    *widget.List
	create, ready, exit     *widget.Button
	heading, status         *widget.Text
	players                 [2]*widget.Text
	readiness               [2]*widget.Text
	indicators              [2]*widget.Graphic
	readyIcon, notReadyIcon *ebiten.Image
	window                  *widget.Window
	waiting                 map[uint64]protocol.GameState
	active                  protocol.GameState
	pending                 bool
	inviting                bool
	editor                  *gameSetupUI
	chat                    *widget.TextArea
	chatInput               *widget.TextInput
	chatSend                *widget.Button
	chatHistory             []string
}

func makeReadinessIcon(ready bool) *ebiten.Image {
	icon := ebiten.NewImage(18, 18)
	if ready {
		green := color.NRGBA{R: 75, G: 220, B: 100, A: 255}
		vector.StrokeLine(icon, 3, 9, 7, 14, 3, green, true)
		vector.StrokeLine(icon, 7, 14, 15, 3, 3, green, true)
	} else {
		red := color.NRGBA{R: 235, G: 65, B: 65, A: 255}
		vector.StrokeLine(icon, 4, 4, 14, 14, 3, red, true)
		vector.StrokeLine(icon, 14, 4, 4, 14, 3, red, true)
	}
	return icon
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

	contents := lobbyColumn(0, 0, 0.75, 0.25, 0, 0)
	contents.GetWidget().SetTheme(theme)
	contents.SetBackgroundImage(uiimage.NewNineSliceColor(color.NRGBA{R: 24, G: 28, B: 36, A: 255}))
	match.readyIcon, match.notReadyIcon = makeReadinessIcon(true), makeReadinessIcon(false)
	var playerRows [2]*widget.Container
	for i := range match.players {
		match.players[i], match.readiness[i] = label(""), label("")
		match.indicators[i] = widget.NewGraphic(widget.GraphicOpts.Image(match.notReadyIcon))
		playerRows[i] = widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal), widget.RowLayoutOpts.Spacing(6),
		)))
		playerRows[i].AddChild(match.players[i], match.indicators[i], match.readiness[i])
	}
	match.status = label("")
	match.ready = button("Ready", func() {
		ready := false
		for _, player := range match.active.Players {
			if player.ClientID == game.clientID {
				ready = player.Ready
			}
		}
		if match.active.ID != 0 && !match.active.Started {
			match.send(protocol.Message{Type: "game-ready", GameID: match.active.ID, Ready: !ready, Revision: match.active.Revision})
		}
	})
	match.exit = button("Exit", func() {
		if match.inviting {
			game.invitations.cancelRequest()
			return
		}
		if match.active.ID != 0 {
			match.send(protocol.Message{Type: "leave-game", GameID: match.active.ID})
		}
	})
	buttons := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(
		widget.RowLayoutOpts.Direction(widget.DirectionHorizontal), widget.RowLayoutOpts.Spacing(8),
	)))
	buttons.AddChild(match.ready, match.exit)
	match.editor = newGameSetupUI(match, theme)
	contents.AddChild(playerRows[0], playerRows[1], match.editor.container, match.newGameChat(), match.status, buttons)
	title := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(
		widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(6)),
	)), widget.ContainerOpts.BackgroundImage(uiimage.NewNineSliceColor(color.NRGBA{R: 42, G: 50, B: 64, A: 255})))
	title.AddChild(label("Game setup"))
	match.window = widget.NewWindow(widget.WindowOpts.Contents(contents), widget.WindowOpts.TitleBar(title, 36),
		widget.WindowOpts.Draggable(), widget.WindowOpts.Resizeable(), widget.WindowOpts.MinSize(500, 300))
	frame := match.window.GetContainer().(*widget.Container)
	frame.GetWidget().SetTheme(theme)
	frame.SetBackgroundImage(uiimage.NewBorderedNineSliceColor(
		color.NRGBA{R: 24, G: 28, B: 36, A: 255}, color.NRGBA{R: 230, G: 215, B: 140, A: 255}, 2,
	))
	widget.ContainerOpts.Layout(widget.NewGridLayout(
		widget.GridLayoutOpts.Columns(1),
		widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true}),
		widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(2)),
	))(frame)
	return match
}

func (match *matchmakingUI) send(message protocol.Message) {
	if match.pending || match.inviting || !match.game.connected || match.game.room == "" {
		return
	}
	if match.game.send(message) {
		match.pending = true
		match.setEnabled()
	}
}

func (match *matchmakingUI) setEnabled() {
	match.editor.setEnabled()
	chatEnabled := match.game.connected && match.active.ID != 0 && !match.inviting
	match.chatInput.GetWidget().Disabled = !chatEnabled
	match.chatSend.GetWidget().Disabled = !chatEnabled
	available := match.game.connected && match.game.room != "" && !match.pending
	match.create.GetWidget().Disabled = !available || match.active.ID != 0 || match.inviting
	match.list.GetWidget().Disabled = !available || match.active.ID != 0 || match.inviting
	match.ready.GetWidget().Disabled = !available || match.active.ID == 0 || match.active.Started || match.inviting || !match.editor.canReady()
	match.exit.GetWidget().Disabled = !available || match.active.ID == 0
	if match.inviting {
		invites := match.game.invitations
		match.exit.GetWidget().Disabled = !available || invites.outgoingID == 0 || invites.cancelPending
	}
}

func (match *matchmakingUI) openWindow() {
	if match.game.ui.IsWindowOpen(match.window) {
		return
	}
	rect := match.game.ui.Container.GetWidget().Rect
	if rect.Empty() {
		rect = image.Rect(0, 0, 1100, 700)
	}
	width, height := min(1000, max(760, rect.Dx()-40)), min(680, max(520, rect.Dy()-40))
	x, y := max(0, (rect.Dx()-width)/2), max(0, (rect.Dy()-height)/2)
	match.window.SetLocation(image.Rect(x, y, x+width, y+height))
	match.game.ui.AddWindow(match.window)
}

func (match *matchmakingUI) showInvitation(targetID uint64) {
	match.inviting = true
	match.clearChat()
	match.editor.container.GetWidget().SetVisibility(widget.Visibility_Hide)
	match.players[0].Label = match.game.userLabel(match.game.clientID)
	match.players[1].Label = match.game.userLabel(targetID) + ": waiting for player"
	for i := range match.players {
		match.indicators[i].GetWidget().SetVisibility(widget.Visibility_Hide)
		match.readiness[i].GetWidget().SetVisibility(widget.Visibility_Hide)
	}
	match.ready.SetText("Ready")
	match.status.Label = "Waiting for the player to accept your game request."
	match.openWindow()
	match.setEnabled()
}

func (match *matchmakingUI) closeInvitationPreview() {
	if !match.inviting {
		return
	}
	match.inviting = false
	match.window.Close()
	match.setEnabled()
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
	match.clearChat()
	match.pending = false
	match.inviting = false
	match.window.Close()
	match.updateList(nil)
	match.setEnabled()
}

func (match *matchmakingUI) handleMessage(message protocol.Message) bool {
	switch message.Type {
	case "game-chat":
		match.receiveChat(message)
	case "list-games":
		if message.Room == match.game.room && match.game.room != "" {
			match.updateList(message.Games)
		}
	case "game-left":
		if message.GameID == match.active.ID {
			match.active = protocol.GameState{}
			match.clearChat()
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
		if match.active.ID != state.ID {
			match.clearChat()
		}
		match.active = *state
		match.inviting = false
		match.editor.container.GetWidget().SetVisibility(widget.Visibility_Show)
		match.editor.sync()
		match.pending = false
		match.ready.SetText("Ready")
		for i := range match.players {
			value := fmt.Sprintf("Player %d: waiting for a player…", i+1)
			match.indicators[i].GetWidget().SetVisibility(widget.Visibility_Hide)
			match.readiness[i].GetWidget().SetVisibility(widget.Visibility_Hide)
			if i < len(state.Players) {
				player := state.Players[i]
				match.indicators[i].GetWidget().SetVisibility(widget.Visibility_Show)
				match.readiness[i].GetWidget().SetVisibility(widget.Visibility_Show)
				match.indicators[i].Image = match.notReadyIcon
				match.readiness[i].Label = "not ready"
				if player.Ready {
					match.indicators[i].Image = match.readyIcon
					match.readiness[i].Label = "ready"
				}
				role := "defender"
				attackerID := state.AttackerID
				if attackerID == 0 {
					attackerID = state.HostID
				}
				if player.ClientID == attackerID {
					role = "attacker"
				}
				value = fmt.Sprintf("Player %d: %s (%s) —", i+1, player.Name, role)
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
			match.status.Label = "Both players ready — connecting to combat…"
		}
		match.openWindow()
		match.setEnabled()
	default:
		return false
	}
	return true
}
