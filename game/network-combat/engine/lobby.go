package engine

import (
	"context"
	"fmt"
	"image/color"
	"maps"
	"math/rand/v2"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/themes"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

func multiplayerURL(address string) (string, error) {
	address = strings.TrimSpace(address)
	if !strings.Contains(address, "://") {
		address = "ws://" + address
	}
	u, err := url.Parse(address)
	if err != nil {
		return "", err
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	case "ws", "wss":
	default:
		return "", fmt.Errorf("use a server address or ws/wss URL")
	}
	if u.Hostname() == "" {
		return "", fmt.Errorf("server address is empty")
	}
	u.Path, u.RawPath, u.RawQuery, u.Fragment = "/multiplayer", "", "", ""
	return u.String(), nil
}

type lobbySession struct {
	cancel   context.CancelFunc
	commands chan protocol.Message
}

type lobbyEvent struct {
	session *lobbySession
	message protocol.Message
	err     error
}

// runLobbySession keeps all network I/O off Ebiten's update thread. A session
// owns its connection; events from cancelled/older sessions are ignored by UI.
func runLobbySession(ctx context.Context, session *lobbySession, address, name string, events chan<- lobbyEvent) {
	defer session.cancel()
	emit := func(event lobbyEvent) {
		event.session = session
		select {
		case events <- event:
		case <-ctx.Done():
		}
	}
	endpoint, err := multiplayerURL(address)
	if err != nil {
		emit(lobbyEvent{err: err})
		return
	}
	dialCtx, dialCancel := context.WithTimeout(ctx, 10*time.Second)
	connection, _, err := websocket.Dial(dialCtx, endpoint, nil)
	dialCancel()
	if err != nil {
		emit(lobbyEvent{err: err})
		return
	}
	defer connection.CloseNow()
	connection.SetReadLimit(1024 * 1024)
	write := func(message any) error {
		writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return wsjson.Write(writeCtx, connection, message)
	}
	if err := write(map[string]string{"auth": "x", "name": name}); err != nil {
		emit(lobbyEvent{err: err})
		return
	}
	if err := write(protocol.Message{Type: "list-rooms", ID: 1}); err != nil {
		emit(lobbyEvent{err: err})
		return
	}
	readCtx, readCancel := context.WithCancel(ctx)
	defer readCancel()
	readDone := make(chan error, 1)
	go func() {
		for {
			var message protocol.Message
			if err := wsjson.Read(readCtx, connection, &message); err != nil {
				readDone <- err
				return
			}
			emit(lobbyEvent{message: message})
		}
	}()
	id := uint64(1)
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-readDone:
			emit(lobbyEvent{err: err})
			return
		case command := <-session.commands:
			id++
			command.ID = id
			if err := write(command); err != nil {
				emit(lobbyEvent{err: err})
				return
			}
		}
	}
}

type lobbyGame struct {
	ui                                     ebitenui.UI
	events                                 chan lobbyEvent
	session                                *lobbySession
	connected                              bool
	clientID                               uint64
	server                                 *widget.TextInput
	name                                   *widget.TextInput
	nameDirty                              bool
	nameChanged                            time.Time
	userNames                              map[uint64]string
	toggle                                 *widget.Button
	status, roomLabel                      *widget.Text
	rooms, users                           *widget.List
	chat                                   *widget.TextArea
	input                                  *widget.TextInput
	sendButton, leaveButton, refreshButton *widget.Button
	room                                   string
	history                                []string
	userIDs                                []uint64
	lastRoom                               string
	lastClick                              time.Time
}

func lobbyColumn(weights ...float64) *widget.Container {
	return widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(
		widget.GridLayoutOpts.Columns(1),
		widget.GridLayoutOpts.Weights([]float64{1}, weights),
		widget.GridLayoutOpts.Spacing(8, 8),
		widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(8)),
	)))
}

func newLobbyGame() *lobbyGame {
	game := &lobbyGame{events: make(chan lobbyEvent, 128)}
	label := func(text string) *widget.Text {
		return widget.NewText(widget.TextOpts.Text(text, nil, color.White))
	}
	button := func(text string, action func()) *widget.Button {
		return widget.NewButton(widget.ButtonOpts.Text(text, nil, nil),
			widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) { action() }))
	}
	root := lobbyColumn(0, 0, 0, 1)
	theme := themes.GetBasicDarkTheme()
	root.GetWidget().SetTheme(theme)
	game.ui = ebitenui.UI{Container: root, PrimaryTheme: theme}
	connectionRow := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(
		widget.GridLayoutOpts.Columns(3), widget.GridLayoutOpts.Weights([]float64{0, 1, 0}, nil),
		widget.GridLayoutOpts.Spacing(8, 0),
	)))
	game.server = widget.NewTextInput(widget.TextInputOpts.Placeholder("localhost:5000"))
	game.server.SetText("localhost:5000")
	game.toggle = button("Connect", func() {
		if game.session != nil {
			game.disconnect("Disconnected")
		} else {
			game.connect()
		}
	})
	connectionRow.AddChild(label("Server"), game.server, game.toggle)
	nameRow := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(
		widget.GridLayoutOpts.Columns(2), widget.GridLayoutOpts.Weights([]float64{0, 1}, nil),
		widget.GridLayoutOpts.Spacing(8, 0),
	)))
	game.name = widget.NewTextInput(widget.TextInputOpts.Placeholder("Your name"),
		widget.TextInputOpts.ChangedHandler(func(*widget.TextInputChangedEventArgs) {
			game.nameDirty = true
			game.nameChanged = time.Now()
		}))
	game.name.SetText(fmt.Sprintf("Player%04d", rand.IntN(10000)))
	nameRow.AddChild(label("Name"), game.name)
	game.status = label("Disconnected")
	root.AddChild(connectionRow, nameRow, game.status)

	body := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(
		widget.GridLayoutOpts.Columns(3),
		widget.GridLayoutOpts.Weights([]float64{0.2, 0.2, 0.6}, []float64{1}),
	)))
	rooms := lobbyColumn(0, 1, 0)
	game.rooms = widget.NewList(
		widget.ListOpts.EntryLabelFunc(func(entry any) string { return entry.(string) }),
		widget.ListOpts.AllowReselect(),
		widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
			room := args.Entry.(string)
			now := time.Now()
			if game.connected && room == game.lastRoom && now.Sub(game.lastClick) < 400*time.Millisecond {
				game.send(protocol.Message{Type: "join-room", Room: room})
				game.lastRoom = ""
			} else {
				game.lastRoom, game.lastClick = room, now
			}
		}),
	)
	game.refreshButton = button("Refresh rooms", func() { game.send(protocol.Message{Type: "list-rooms"}) })
	rooms.AddChild(label("Rooms"), game.rooms, game.refreshButton)
	users := lobbyColumn(0, 1, 0)
	game.roomLabel = label("Not in a room")
	game.users = widget.NewList(widget.ListOpts.EntryLabelFunc(func(entry any) string { return game.userLabel(entry.(uint64)) }))
	game.leaveButton = button("Leave room", func() { game.send(protocol.Message{Type: "leave-room"}) })
	users.AddChild(game.roomLabel, game.users, game.leaveButton)
	chat := lobbyColumn(0, 1, 0, 0)
	game.chat = widget.NewTextArea(widget.TextAreaOpts.Text(""),
		widget.TextAreaOpts.ProcessBBCode(false), widget.TextAreaOpts.ShowVerticalScrollbar(),
		widget.TextAreaOpts.VerticalScrollMode(widget.ScrollEnd))
	game.input = widget.NewTextInput(widget.TextInputOpts.Placeholder("Type a chat message"),
		widget.TextInputOpts.AllowDuplicateSubmit(true),
		widget.TextInputOpts.SubmitHandler(func(*widget.TextInputChangedEventArgs) { game.sendChat() }))
	game.sendButton = button("Send", game.sendChat)
	chat.AddChild(label("Chat"), game.chat, game.input, game.sendButton)
	body.AddChild(rooms, users, chat)
	root.AddChild(body)
	game.setEnabled()
	return game
}

func (game *lobbyGame) setEnabled() {
	game.server.GetWidget().Disabled = game.session != nil
	game.rooms.GetWidget().Disabled = !game.connected
	game.refreshButton.GetWidget().Disabled = !game.connected
	inRoom := game.connected && game.room != ""
	game.input.GetWidget().Disabled = !inRoom
	game.sendButton.GetWidget().Disabled = !inRoom
	game.leaveButton.GetWidget().Disabled = !inRoom
}

func (game *lobbyGame) connect() {
	ctx, cancel := context.WithCancel(context.Background())
	session := &lobbySession{cancel: cancel, commands: make(chan protocol.Message, 32)}
	game.session = session
	game.toggle.SetText("Disconnect")
	game.status.Label = "Connecting..."
	game.setEnabled()
	game.nameDirty = false
	go runLobbySession(ctx, session, game.server.GetText(), strings.TrimSpace(game.name.GetText()), game.events)
}

func (game *lobbyGame) clearRoom() {
	game.room = ""
	game.roomLabel.Label = "Not in a room"
	game.userIDs = nil
	game.userNames = nil
	game.users.SetEntries(nil)
	game.history = nil
	game.chat.SetText("")
	game.input.SetText("")
	game.setEnabled()
}

func (game *lobbyGame) disconnect(status string) {
	if game.session != nil {
		game.session.cancel()
	}
	game.session = nil
	game.connected = false
	game.clientID = 0
	game.toggle.SetText("Connect")
	game.status.Label = status
	game.rooms.SetEntries(nil)
	game.lastRoom = ""
	game.clearRoom()
}

func (game *lobbyGame) send(message protocol.Message) bool {
	if !game.connected || game.session == nil {
		return false
	}
	select {
	case game.session.commands <- message:
		return true
	default:
		game.status.Label = "Too many pending commands; try again shortly"
		return false
	}
}

func (game *lobbyGame) sendChat() {
	text := strings.TrimSpace(game.input.GetText())
	if game.room == "" || text == "" {
		return
	}
	if len(text) > 2000 {
		game.status.Label = "Chat messages must be at most 2000 bytes"
		return
	}
	if game.send(protocol.Message{Type: "chat", Text: text}) {
		game.input.SetText("")
	}
}

func stringEntries(values []string) []any {
	out := make([]any, len(values))
	for i, value := range values {
		out[i] = value
	}
	return out
}

// Labels are presentation only; membership and events always use client IDs.
func clientLabel(id uint64) string { return fmt.Sprintf("Player-%06x", id) }

func (game *lobbyGame) userLabel(id uint64) string {
	if name := game.userNames[id]; name != "" {
		return name
	}
	return clientLabel(id)
}

func (game *lobbyGame) updateUsers() {
	entries := make([]any, len(game.userIDs))
	for i, id := range game.userIDs {
		entries[i] = id
	}
	game.users.SetEntries(entries)
}

func (game *lobbyGame) appendChat(message string) {
	game.history = append(game.history, message)
	if len(game.history) > 500 {
		game.history = game.history[len(game.history)-500:]
	}
	game.chat.SetText(strings.Join(game.history, "\n"))
}

func (game *lobbyGame) handleEvent(event lobbyEvent) {
	if event.session != game.session {
		return
	}
	if event.err != nil {
		game.disconnect("Connection closed: " + event.err.Error())
		return
	}
	message := event.message
	if message.Error != "" {
		game.status.Label = message.Error
		return
	}
	switch message.Type {
	case "connected":
		game.clientID = message.ClientID
		game.connected = true
		game.status.Label = "Connected as " + message.Name
		game.setEnabled()
	case "list-rooms":
		game.rooms.SetEntries(stringEntries(message.Rooms))
		game.lastRoom = ""
		game.status.Label = fmt.Sprintf("Connected — %d rooms", len(message.Rooms))
	case "join-room":
		game.clearRoom()
		game.room = message.Room
		game.roomLabel.Label = "Users in " + message.Room
		game.status.Label = "Joined " + message.Room
		game.setEnabled()
	case "leave-room":
		game.clearRoom()
	case "room-users":
		if message.Room == game.room {
			game.userIDs = slices.Clone(message.Users)
			game.userNames = maps.Clone(message.Names)
			game.updateUsers()
		}
	case "user-joined":
		if message.Room == game.room && game.room != "" && message.ClientID != game.clientID && !slices.Contains(game.userIDs, message.ClientID) {
			game.userIDs = append(game.userIDs, message.ClientID)
			slices.Sort(game.userIDs)
			if game.userNames == nil {
				game.userNames = make(map[uint64]string)
			}
			game.userNames[message.ClientID] = message.Name
			game.updateUsers()
			game.appendChat("** " + game.userLabel(message.ClientID) + " has joined")
		}
	case "user-left":
		if message.Room == game.room {
			if message.ClientID != game.clientID && slices.Contains(game.userIDs, message.ClientID) {
				game.appendChat("** " + game.userLabel(message.ClientID) + " has left")
			}
			game.userIDs = slices.DeleteFunc(game.userIDs, func(id uint64) bool { return id == message.ClientID })
			delete(game.userNames, message.ClientID)
			game.updateUsers()
		}
	case "name-updated":
		if message.Room == game.room && game.room != "" && slices.Contains(game.userIDs, message.ClientID) {
			oldName := game.userLabel(message.ClientID)
			if game.userNames == nil {
				game.userNames = make(map[uint64]string)
			}
			game.userNames[message.ClientID] = message.Name
			game.updateUsers()
			if oldName != message.Name {
				game.appendChat("** " + oldName + " is now known as " + message.Name)
			}
		}
	case "chat":
		if message.Room == game.room {
			game.appendChat(game.userLabel(message.ClientID) + ": " + message.Text)
		}
	}
}

func (game *lobbyGame) Update() error {
	// Coalesce typing so name edits do not flood the server's command queue.
	if game.connected && game.nameDirty && time.Since(game.nameChanged) >= 400*time.Millisecond {
		name := strings.TrimSpace(game.name.GetText())
		if name != "" && game.send(protocol.Message{Type: "set-name", Name: name}) {
			game.nameDirty = false
		}
	}
	for i := 0; i < 128; i++ {
		select {
		case event := <-game.events:
			game.handleEvent(event)
		default:
			game.ui.Update()
			return nil
		}
	}
	game.ui.Update()
	return nil
}

func (game *lobbyGame) Draw(screen *ebiten.Image) {
	screen.Fill(color.NRGBA{R: 24, G: 28, B: 36, A: 255})
	game.ui.Draw(screen)
}

func (game *lobbyGame) Layout(width, height int) (int, int) { return width, height }
