package engine

import (
	"strings"
	"time"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

func (match *matchmakingUI) newGameChat() *widget.Container {
	panel := lobbyColumn(1, 0)
	match.chat = widget.NewTextArea(widget.TextAreaOpts.Text(""),
		widget.TextAreaOpts.ProcessBBCode(false), widget.TextAreaOpts.ShowVerticalScrollbar(),
		widget.TextAreaOpts.VerticalScrollMode(widget.ScrollEnd))
	match.chatInput = widget.NewTextInput(widget.TextInputOpts.Placeholder("Game message"),
		widget.TextInputOpts.AllowDuplicateSubmit(true),
		widget.TextInputOpts.SubmitHandler(func(*widget.TextInputChangedEventArgs) { match.sendChat() }))
	match.chatSend = widget.NewButton(widget.ButtonOpts.Text("Send", nil, nil),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) { match.sendChat() }))
	input := setupRow([]float64{1, 0})
	input.AddChild(match.chatInput, match.chatSend)
	panel.AddChild(match.chat, input)
	return panel
}

func (match *matchmakingUI) sendChat() {
	text := strings.TrimSpace(match.chatInput.GetText())
	if text == "" || match.active.ID == 0 || match.inviting {
		return
	}
	if len(text) > 2000 {
		match.status.Label = "Chat messages must be at most 2000 bytes"
		return
	}
	// Chat has its own echo event; it must not wait for a game-state update or
	// disable setup controls through match.pending.
	if match.game.send(protocol.Message{Type: "game-chat", GameID: match.active.ID, Text: text}) {
		match.chatInput.SetText("")
	}
}

func (match *matchmakingUI) clearChat() {
	match.chatHistory = nil
	match.chat.SetText("")
	match.chatInput.SetText("")
}

func (match *matchmakingUI) receiveChat(message protocol.Message) {
	if match.active.ID == 0 || message.GameID != match.active.ID {
		return
	}
	name := ""
	for _, player := range match.active.Players {
		if player.ClientID == message.ClientID {
			name = player.Name
			break
		}
	}
	if name == "" {
		return
	}
	if message.Name != "" {
		name = message.Name
	}
	match.chatHistory = append(match.chatHistory, time.Now().Local().Format("03:04:05 PM")+" "+name+": "+message.Text)
	if len(match.chatHistory) > 500 {
		match.chatHistory = match.chatHistory[len(match.chatHistory)-500:]
	}
	match.chat.SetText(strings.Join(match.chatHistory, "\n"))
}
