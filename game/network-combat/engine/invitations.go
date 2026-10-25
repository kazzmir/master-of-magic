package engine

import (
	"fmt"
	"image"
	"image/color"

	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

type gameInvitationsUI struct {
	game                       *lobbyGame
	window                     *widget.Window
	prompt                     *widget.Text
	accept, decline            *widget.Button
	incomingID, outgoingID     uint64
	requestPending, responding bool
}

func newGameInvitationsUI(game *lobbyGame, theme *widget.Theme) *gameInvitationsUI {
	invites := &gameInvitationsUI{game: game}
	contents := lobbyColumn(1, 0)
	contents.GetWidget().SetTheme(theme)
	contents.SetBackgroundImage(uiimage.NewNineSliceColor(color.NRGBA{R: 24, G: 28, B: 36, A: 255}))
	invites.prompt = widget.NewText(widget.TextOpts.Text("", nil, color.White), widget.TextOpts.MaxWidth(520), widget.TextOpts.ProcessBBCode(false))
	invites.accept = widget.NewButton(widget.ButtonOpts.Text("Accept", nil, nil), widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) { invites.respond(true) }))
	invites.decline = widget.NewButton(widget.ButtonOpts.Text("Decline", nil, nil), widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) { invites.respond(false) }))
	buttons := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionHorizontal), widget.RowLayoutOpts.Spacing(8))))
	buttons.AddChild(invites.accept, invites.decline)
	contents.AddChild(invites.prompt, buttons)
	title := lobbyColumn(0)
	title.AddChild(widget.NewText(widget.TextOpts.Text("Game request", nil, color.White)))
	invites.window = widget.NewWindow(widget.WindowOpts.Contents(contents), widget.WindowOpts.TitleBar(title, 36),
		widget.WindowOpts.Draggable(), widget.WindowOpts.Modal(), widget.WindowOpts.MinSize(500, 200))
	return invites
}

func (invites *gameInvitationsUI) contextMenu(clientID uint64) *widget.Container {
	if clientID == invites.game.clientID {
		return nil
	}
	menu := lobbyColumn(0)
	menu.GetWidget().SetTheme(invites.game.ui.PrimaryTheme)
	menu.AddChild(widget.NewButton(widget.ButtonOpts.Text("Request to join a game", nil, nil),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) { invites.request(clientID) })))
	return menu
}

func (invites *gameInvitationsUI) request(clientID uint64) {
	game := invites.game
	if !game.connected || game.room == "" || clientID == game.clientID {
		return
	}
	if game.matches.active.ID != 0 {
		game.status.Label = "Exit your current game before requesting another player"
		return
	}
	if invites.requestPending || invites.outgoingID != 0 || invites.incomingID != 0 {
		game.status.Label = "You already have a pending game request"
		return
	}
	if game.send(protocol.Message{Type: "request-game", TargetID: clientID}) {
		invites.requestPending = true
	}
}

func (invites *gameInvitationsUI) respond(accept bool) {
	if invites.incomingID == 0 || invites.responding {
		return
	}
	if invites.game.send(protocol.Message{Type: "respond-game-invite", InviteID: invites.incomingID, Accept: accept}) {
		invites.responding = true
		invites.accept.GetWidget().Disabled = true
		invites.decline.GetWidget().Disabled = true
	}
}

func (invites *gameInvitationsUI) reset() {
	invites.incomingID, invites.outgoingID = 0, 0
	invites.requestPending, invites.responding = false, false
	invites.window.Close()
}

func (invites *gameInvitationsUI) handleError(message protocol.Message) {
	invites.requestPending = false
	if message.InviteID != 0 && message.InviteID == invites.incomingID {
		invites.incomingID = 0
		invites.responding = false
		invites.window.Close()
	}
}

func (invites *gameInvitationsUI) handleMessage(message protocol.Message) bool {
	switch message.Type {
	case "invite-sent":
		invites.requestPending = false
		invites.outgoingID = message.InviteID
		invites.game.status.Label = "Game request sent to " + invites.game.userLabel(message.TargetID)
	case "game-invite":
		if message.Room != invites.game.room || invites.game.room == "" {
			return true
		}
		invites.incomingID = message.InviteID
		invites.responding = false
		invites.accept.GetWidget().Disabled = false
		invites.decline.GetWidget().Disabled = false
		invites.prompt.Label = fmt.Sprintf("Game request from %s\nJoin a new game together?", message.Name)
		if !invites.game.ui.IsWindowOpen(invites.window) {
			rect := invites.game.ui.Container.GetWidget().Rect
			x, y := max(0, (rect.Dx()-550)/2), max(0, (rect.Dy()-220)/2)
			invites.window.SetLocation(image.Rect(x, y, x+550, y+220))
			invites.game.ui.AddWindow(invites.window)
		}
	case "invite-closed":
		matched := false
		if message.InviteID == invites.incomingID && invites.incomingID != 0 {
			invites.incomingID = 0
			invites.responding = false
			invites.window.Close()
			matched = true
		}
		if message.InviteID == invites.outgoingID && invites.outgoingID != 0 {
			invites.outgoingID = 0
			invites.requestPending = false
			matched = true
		}
		if matched {
			invites.game.status.Label = message.Text
		}
	default:
		return false
	}
	return true
}
