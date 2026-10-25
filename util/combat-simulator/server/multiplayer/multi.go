package multiplayer

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
	netlib "github.com/kazzmir/master-of-magic/lib/net"
)

type User struct {
	messages   chan protocol.Message
	connection *websocket.Conn
	room       string
}

type Lobby struct {
	Public bool
	Name   string
	Users  map[string]*User
}

// All lobby and membership state is protected by mu. Network writes never
// hold mu: each connection has a bounded outbound queue and a single writer.
type Manager struct {
	mu      sync.Mutex
	Lobbies map[string]*Lobby
}

func (manager *Manager) ListPublicLobbies() []string {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	out := make([]string, 0, len(manager.Lobbies))
	for _, lobby := range manager.Lobbies {
		if lobby.Public {
			out = append(out, lobby.Name)
		}
	}
	slices.Sort(out)
	return out
}

func (user *User) send(message protocol.Message) {
	select {
	case user.messages <- message:
	default:
		// A stalled client must not block everyone else in the room.
		user.connection.CloseNow()
	}
}

func broadcast(lobby *Lobby, message protocol.Message) {
	for _, user := range lobby.Users {
		user.send(message)
	}
}

func broadcastUsers(lobby *Lobby) {
	users := make([]string, 0, len(lobby.Users))
	for name := range lobby.Users {
		users = append(users, name)
	}
	slices.Sort(users)
	broadcast(lobby, protocol.Message{Type: "room-users", Room: lobby.Name, Users: users})
}

// leave is called with manager.mu held.
func (manager *Manager) leave(user *User, name string) {
	if lobby := manager.Lobbies[user.room]; lobby != nil {
		delete(lobby.Users, name)
		broadcastUsers(lobby)
	}
	user.room = ""
}

func handleWebsocket(connection *websocket.Conn, manager *Manager, connectionID uint64) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	connection.SetReadLimit(16 * 1024)

	authCtx, authCancel := context.WithTimeout(ctx, 10*time.Second)
	var auth struct {
		Auth string `json:"auth"`
	}
	err := wsjson.Read(authCtx, connection, &auth)
	authCancel()
	if err != nil || auth.Auth == "" {
		connection.Close(websocket.StatusPolicyViolation, "Missing or invalid auth field")
		return
	}
	// Authentication is deliberately a placeholder for now.
	name := fmt.Sprintf("Player-%06x", connectionID)
	user := &User{messages: make(chan protocol.Message, 64), connection: connection}
	defer func() {
		manager.mu.Lock()
		defer manager.mu.Unlock()
		manager.leave(user, name)
	}()

	go func() {
		defer connection.CloseNow()
		for {
			select {
			case <-ctx.Done():
				return
			case message := <-user.messages:
				writeCtx, writeCancel := context.WithTimeout(ctx, 5*time.Second)
				err := wsjson.Write(writeCtx, connection, message)
				writeCancel()
				if err != nil {
					return
				}
			}
		}
	}()

	limiter := netlib.RateLimiter{WindowSize: time.Second, MaxEvents: 2}
	for {
		limiter.AddAndLimit(ctx)
		var command protocol.Message
		if err := wsjson.Read(ctx, connection, &command); err != nil {
			return
		}
		if command.Type == "list-rooms" {
			user.send(protocol.Message{Type: "list-rooms", ID: command.ID, Rooms: manager.ListPublicLobbies()})
			continue
		}
		manager.mu.Lock()
		switch command.Type {
		case "join-room":
			lobby := manager.Lobbies[command.Room]
			if lobby == nil || !lobby.Public {
				user.send(protocol.Message{ID: command.ID, Error: "Room not found"})
			} else if len(lobby.Users) >= 100 && user.room != command.Room {
				user.send(protocol.Message{ID: command.ID, Error: "Room is full"})
			} else {
				manager.leave(user, name)
				user.room = lobby.Name
				lobby.Users[name] = user
				user.send(protocol.Message{Type: "join-room", ID: command.ID, Room: lobby.Name, User: name})
				broadcastUsers(lobby)
			}
		case "leave-room":
			manager.leave(user, name)
			user.send(protocol.Message{Type: "leave-room", ID: command.ID})
		case "chat":
			text := strings.TrimSpace(command.Text)
			if lobby := manager.Lobbies[user.room]; lobby == nil {
				user.send(protocol.Message{ID: command.ID, Error: "Join a room first"})
			} else if text == "" || len(text) > 2000 {
				user.send(protocol.Message{ID: command.ID, Error: "Chat must contain 1–2000 bytes"})
			} else {
				broadcast(lobby, protocol.Message{Type: "chat", Room: lobby.Name, User: name, Text: text})
			}
		default:
			user.send(protocol.Message{ID: command.ID, Error: fmt.Sprintf("Unknown type '%s'", command.Type)})
		}
		manager.mu.Unlock()
	}
}

func InitMux(mux *http.ServeMux) {
	manager := &Manager{Lobbies: map[string]*Lobby{
		"Public": {Public: true, Name: "Public", Users: make(map[string]*User)},
	}}
	var counter atomic.Uint64
	mux.HandleFunc("/multiplayer", func(writer http.ResponseWriter, request *http.Request) {
		connection, err := websocket.Accept(writer, request, nil)
		if err != nil {
			log.Printf("Error accepting websocket connection: %v", err)
			return
		}
		defer connection.CloseNow()
		handleWebsocket(connection, manager, counter.Add(1))
	})
}
