package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

// chatText verifies local-time prefixes while allowing content assertions to
// remain independent of the wall clock.
func chatText(t *testing.T, game *lobbyGame) string {
	t.Helper()
	value := game.chat.GetText()
	if value == "" {
		return ""
	}
	lines := strings.Split(value, "\n")
	for i, line := range lines {
		if len(line) < 12 {
			t.Fatalf("missing chat timestamp: %q", line)
		}
		if _, err := time.ParseInLocation("03:04:05 PM", line[:11], time.Local); err != nil || line[11] != ' ' {
			t.Fatalf("invalid chat timestamp: %q", line)
		}
		lines[i] = line[12:]
	}
	return strings.Join(lines, "\n")
}

func TestMultiplayerURL(t *testing.T) {
	for input, want := range map[string]string{
		"localhost:5000":                "ws://localhost:5000/multiplayer",
		" https://example.com/ignored ": "wss://example.com/multiplayer",
		"ws://localhost:5000":           "ws://localhost:5000/multiplayer",
	} {
		got, err := multiplayerURL(input)
		if err != nil || got != want {
			t.Fatalf("%q: %q, %v", input, got, err)
		}
	}
	for _, input := range []string{"", "ftp://example.com", "ws://"} {
		if _, err := multiplayerURL(input); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
}

func TestLobbyUI(t *testing.T) {
	game := newLobbyGame()
	if !regexp.MustCompile(`^Player[0-9]{4}$`).MatchString(game.name.GetText()) {
		t.Fatal("invalid default name", game.name.GetText())
	}
	if err := game.Update(); err != nil {
		t.Fatal(err)
	}
	game.Draw(ebiten.NewImage(1100, 700))
	session := &lobbySession{cancel: func() {}, commands: make(chan protocol.Message, 4)}
	game.session = session
	for _, message := range []protocol.Message{
		{Type: "connected", ClientID: 1},
		{Type: "list-rooms", Rooms: []string{"Lobby"}},
		{Type: "join-room", Room: "Lobby", ClientID: 1},
		{Type: "room-users", Room: "Lobby", Users: []uint64{1, 2}, Names: map[uint64]string{1: "Me", 2: "Other"}},
		{Type: "chat", Room: "Lobby", ClientID: 2, Text: "[hello]"},
	} {
		game.handleEvent(lobbyEvent{session: session, message: message})
	}
	if game.clientID != 1 || !game.connected {
		t.Fatal("server-assigned client ID not stored")
	}
	if chatText(t, game) != "Other: [hello]" {
		t.Fatal(game.chat.GetText())
	}
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "name-updated", Room: "Elsewhere", ClientID: 2, Name: "Ignored"}})
	if game.userLabel(2) != "Other" {
		t.Fatal("name update from another room applied")
	}
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "name-updated", Room: "Lobby", ClientID: 2, Name: "New name"}})
	if game.userLabel(2) != "New name" || !slices.Equal(game.userIDs, []uint64{1, 2}) {
		t.Fatal("rename changed identity")
	}
	wantChat := "Other: [hello]\n** Other is now known as New name"
	if chatText(t, game) != wantChat {
		t.Fatal("missing rename announcement", game.chat.GetText())
	}
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "name-updated", Room: "Lobby", ClientID: 2, Name: "New name"}})
	if chatText(t, game) != wantChat {
		t.Fatal("unchanged name produced an announcement", game.chat.GetText())
	}
	game.name.SetText("My new name")
	game.nameChanged = time.Now().Add(-time.Second)
	game.Update()
	if command := <-session.commands; command.Type != "set-name" || command.Name != "My new name" {
		t.Fatal(command)
	}
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "user-left", Room: "Elsewhere", ClientID: 2}})
	if !slices.Equal(game.userIDs, []uint64{1, 2}) {
		t.Fatal("departure from another room changed membership", game.userIDs)
	}
	for _, id := range []uint64{2, 2, 99} {
		game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "user-left", Room: "Lobby", ClientID: id}})
		if !slices.Equal(game.userIDs, []uint64{1}) {
			t.Fatal("incorrect membership after departure", game.userIDs)
		}
	}
	if _, exists := game.userNames[2]; exists {
		t.Fatal("departed user's name retained")
	}
	game.input.SetText("hello")
	game.sendChat()
	if command := <-session.commands; command.Type != "chat" || command.Text != "hello" {
		t.Fatal(command)
	}
	game.disconnect("Disconnected")
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "connected", ClientID: 1}})
	if game.connected || game.clientID != 0 || game.room != "" || game.chat.GetText() != "" || len(game.userIDs) != 0 {
		t.Fatal("stale session restored state")
	}
	game.Update()
	game.Draw(ebiten.NewImage(800, 500))
}

func TestLobbyMembershipAnnouncements(t *testing.T) {
	game := newLobbyGame()
	session := &lobbySession{cancel: func() {}}
	game.session = session
	deliver := func(message protocol.Message) {
		game.handleEvent(lobbyEvent{session: session, message: message})
	}
	deliver(protocol.Message{Type: "connected", ClientID: 1})
	deliver(protocol.Message{Type: "join-room", Room: "Lobby", ClientID: 1})
	deliver(protocol.Message{Type: "room-users", Room: "Lobby", Users: []uint64{1, 3}, Names: map[uint64]string{1: "Me", 3: "Existing"}})
	deliver(protocol.Message{Type: "user-joined", Room: "Elsewhere", ClientID: 2, Name: "Other"})
	deliver(protocol.Message{Type: "user-joined", Room: "Lobby", ClientID: 1, Name: "Me"})
	if game.chat.GetText() != "" {
		t.Fatal("snapshot or unrelated join produced announcement")
	}
	for range 2 {
		deliver(protocol.Message{Type: "user-joined", Room: "Lobby", ClientID: 2, Name: "Player1234"})
	}
	if chatText(t, game) != "** Player1234 has joined" {
		t.Fatal(game.chat.GetText())
	}
	if !slices.Equal(game.userIDs, []uint64{1, 2, 3}) {
		t.Fatal(game.userIDs)
	}
	deliver(protocol.Message{Type: "name-updated", Room: "Lobby", ClientID: 2, Name: "newname"})
	deliver(protocol.Message{Type: "user-left", Room: "Elsewhere", ClientID: 2})
	for range 2 {
		deliver(protocol.Message{Type: "user-left", Room: "Lobby", ClientID: 2})
	}
	want := "** Player1234 has joined\n** Player1234 is now known as newname\n** newname has left"
	if chatText(t, game) != want {
		t.Fatal(game.chat.GetText())
	}
	if !slices.Equal(game.userIDs, []uint64{1, 3}) {
		t.Fatal(game.userIDs)
	}
}

func TestRoomsRefreshTime(t *testing.T) {
	game := newLobbyGame()
	if game.roomsRefreshed.Label != "Last refreshed: Never" {
		t.Fatal(game.roomsRefreshed.Label)
	}
	session := &lobbySession{cancel: func() {}}
	game.session = session
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "list-rooms"}})
	stamp := strings.TrimPrefix(game.roomsRefreshed.Label, "Last refreshed: ")
	if _, err := time.ParseInLocation("03:04:05 PM", stamp, time.Local); err != nil {
		t.Fatal("invalid room refresh timestamp", game.roomsRefreshed.Label, err)
	}
	previous := game.roomsRefreshed.Label
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Error: "Refresh failed"}})
	if game.roomsRefreshed.Label != previous {
		t.Fatal("failed refresh changed last successful refresh time")
	}
}

func TestJoinSelectedRoomButton(t *testing.T) {
	game := newLobbyGame()
	if !game.joinButton.GetWidget().Disabled {
		t.Fatal("join enabled while disconnected")
	}
	session := &lobbySession{cancel: func() {}, commands: make(chan protocol.Message, 4)}
	game.session = session
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "connected", ClientID: 1}})
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "list-rooms", Rooms: []string{"First", "Selected"}}})
	game.Update()
	game.joinButton.Click()
	game.Update()
	if len(session.commands) != 0 {
		t.Fatal("join sent without a selection")
	}
	game.rooms.SetSelectedEntry("Selected")
	game.joinButton.Click()
	game.Update()
	select {
	case command := <-session.commands:
		if command.Type != "join-room" || command.Room != "Selected" {
			t.Fatal(command)
		}
	default:
		t.Fatal("join button did not send command")
	}
}

func TestCreateRoomUI(t *testing.T) {
	game := newLobbyGame()
	if !game.createButton.GetWidget().Disabled || !game.newRoom.GetWidget().Disabled {
		t.Fatal("room creation enabled while disconnected")
	}
	session := &lobbySession{cancel: func() {}, commands: make(chan protocol.Message, 4)}
	game.session = session
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "connected", ClientID: 1}})
	game.Update()
	game.createButton.Click()
	game.Update()
	if len(session.commands) != 0 {
		t.Fatal("empty room creation sent")
	}
	game.newRoom.SetText("  New lobby  ")
	game.privateRoom.SetState(widget.WidgetChecked)
	game.newRoomPassword.SetText("secret")
	game.createButton.Click()
	game.Update()
	select {
	case command := <-session.commands:
		if command.Type != "create-room" || command.Room != "New lobby" || !command.Private || command.Password != "secret" {
			t.Fatal(command)
		}
	default:
		t.Fatal("create button did not send command")
	}
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "create-room", Room: "New lobby"}})
	if game.newRoom.GetText() != "" || game.newRoomPassword.GetText() != "" || len(game.pendingRoomPasswords) != 0 {
		t.Fatal("room inputs or pending credentials not cleared")
	}
	for _, want := range []protocol.Message{{Type: "list-rooms"}, {Type: "join-room", Room: "New lobby", Password: "secret"}} {
		select {
		case got := <-session.commands:
			if got.Type != want.Type || got.Room != want.Room || got.Password != want.Password {
				t.Fatal(got)
			}
		default:
			t.Fatal("missing room refresh/join command")
		}
	}
}

func TestJoinPrivateRoomByName(t *testing.T) {
	game := newLobbyGame()
	session := &lobbySession{cancel: func() {}, commands: make(chan protocol.Message, 4)}
	game.session = session
	game.handleEvent(lobbyEvent{session: session, message: protocol.Message{Type: "connected", ClientID: 1}})
	game.Update()
	game.joinRoomName.SetText("Hidden room")
	game.joinPassword.SetText("password")
	game.joinButton.Click()
	game.Update()
	select {
	case command := <-session.commands:
		if command.Type != "join-room" || command.Room != "Hidden room" || command.Password != "password" {
			t.Fatal(command)
		}
	default:
		t.Fatal("no private room join command")
	}
}

func TestLobbySessionHandshake(t *testing.T) {
	errors := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/multiplayer" {
			t.Errorf("path: %s", r.URL.Path)
		}
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			errors <- err
			return
		}
		defer conn.CloseNow()
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		var auth map[string]string
		if err := wsjson.Read(ctx, conn, &auth); err != nil {
			errors <- err
			return
		}
		if auth["auth"] != "x" || auth["name"] != "Player1234" {
			t.Error("missing placeholder auth or initial name", auth)
		}
		if err := wsjson.Write(ctx, conn, protocol.Message{Type: "connected", ClientID: 42}); err != nil {
			errors <- err
			return
		}
		var command protocol.Message
		if err := wsjson.Read(ctx, conn, &command); err != nil {
			errors <- err
			return
		}
		if command.Type != "list-rooms" {
			t.Errorf("first command: %+v", command)
		}
		errors <- wsjson.Write(ctx, conn, protocol.Message{Type: "list-rooms", Rooms: []string{"Lobby"}})
		conn.Read(ctx)
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	session := &lobbySession{cancel: cancel, commands: make(chan protocol.Message, 4)}
	events := make(chan lobbyEvent, 8)
	go runLobbySession(ctx, session, strings.TrimPrefix(server.URL, "http://"), "Player1234", events)
	var clientID uint64
	for {
		select {
		case event := <-events:
			if event.err != nil {
				t.Fatal(event.err)
			}
			if event.message.Type == "connected" {
				clientID = event.message.ClientID
			}
			if event.message.Type == "list-rooms" {
				if clientID != 42 {
					t.Fatal("missing server-assigned client ID", clientID)
				}
				if err := <-errors; err != nil {
					t.Fatal(err)
				}
				return
			}
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
}
