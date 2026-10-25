package multiplayer

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
	netlib "github.com/kazzmir/master-of-magic/lib/net"
	"golang.org/x/crypto/bcrypt"
)

const MaxRooms = 1000
const MaxUsers = 10000

type User struct {
	messages   chan protocol.Message
	connection *websocket.Conn
	room       string
	name       string
	gameID     uint64
}

type Lobby struct {
	Public       bool
	Name         string
	Users        map[uint64]*User
	passwordHash []byte
	persistent   bool
}

type CriticalSection struct {
	mu sync.Mutex
}

func (cs *CriticalSection) With(f func() error) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return f()
}

func (cs *CriticalSection) With2[T any](f func() (T, error)) (T, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return f()
}

// All lobby and membership state is protected by lock. Network writes never
// hold lock: each connection has a bounded outbound queue and a single writer.
type Manager struct {
	lock         CriticalSection
	authKey      string
	userSlots    int // Lobby connections, including pending authentication.
	Lobbies      map[string]*Lobby
	Games        map[uint64]*Game
	Invitations  map[uint64]*GameInvitation
	nextInviteID uint64
}

func (manager *Manager) ListPublicLobbies() []string {
	out, _ := manager.lock.With2(func() ([]string, error) {
		out := make([]string, 0, len(manager.Lobbies))
		for _, lobby := range manager.Lobbies {
			if lobby.Public {
				out = append(out, lobby.Name)
			}
		}
		slices.Sort(out)
		return out, nil
	})
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
	users := make([]uint64, 0, len(lobby.Users))
	names := make(map[uint64]string, len(lobby.Users))
	for clientID, user := range lobby.Users {
		users = append(users, clientID)
		names[clientID] = user.name
	}
	slices.Sort(users)
	broadcast(lobby, protocol.Message{Type: "room-users", Room: lobby.Name, Users: users, Names: names})
}

// leave is called with manager.lock held.
func (manager *Manager) leave(user *User, clientID uint64) {
	manager.cancelInvitations(clientID, "Game request cancelled: a player left the room")
	manager.leaveGame(user, clientID)
	if lobby := manager.Lobbies[user.room]; lobby != nil {
		if _, exists := lobby.Users[clientID]; exists {
			log.Printf("Client %d (%q) left room %q", clientID, user.name, lobby.Name)
			delete(lobby.Users, clientID)
			if len(lobby.Users) == 0 && !lobby.persistent {
				delete(manager.Lobbies, lobby.Name)
			} else {
				broadcast(lobby, protocol.Message{Type: "user-left", Room: lobby.Name, ClientID: clientID})
			}
		}
	}
	user.room = ""
}

func validName(name string) bool {
	return name != "" && utf8.ValidString(name) && utf8.RuneCountInString(name) <= protocol.MaxUserNameLength && !strings.ContainsFunc(name, unicode.IsControl)
}

func handleWebsocket(connection *websocket.Conn, manager *Manager, connectionID uint64) {
	if !manager.reserveUserSlot() {
		connection.Close(websocket.StatusTryAgainLater, "Server user limit reached (10000 users)")
		return
	}
	defer manager.releaseUserSlot()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	connection.SetReadLimit(16 * 1024)

	authCtx, authCancel := context.WithTimeout(ctx, 10*time.Second)
	var auth struct {
		Auth string `json:"auth"`
		Name string `json:"name"`
	}
	err := wsjson.Read(authCtx, connection, &auth)
	authCancel()
	if err != nil || manager.authKey == "" || subtle.ConstantTimeCompare([]byte(auth.Auth), []byte(manager.authKey)) != 1 {
		connection.Close(websocket.StatusPolicyViolation, "Missing or invalid auth field")
		return
	}
	// Shared-key authentication; per-user accounts may be added later.
	name := strings.TrimSpace(auth.Name)
	if auth.Name == "" {
		name = fmt.Sprintf("Player%04d", connectionID%10000)
	}
	if !validName(name) {
		connection.Close(websocket.StatusPolicyViolation, "Name must contain 1–30 characters without control characters")
		return
	}
	user := &User{messages: make(chan protocol.Message, 64), connection: connection, name: name}
	log.Printf("Client %d (%q) connected", connectionID, user.name)
	defer manager.lock.With(func() error {
		manager.leave(user, connectionID)
		log.Printf("Client %d (%q) disconnected", connectionID, user.name)
		return nil
	})

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

	user.send(protocol.Message{Type: "connected", ClientID: connectionID, Name: user.name})
	user.send(protocol.Message{Type: "version", Version: protocol.ServerVersion})

	limiter := netlib.RateLimiter{WindowSize: time.Second, MaxEvents: 2}
	for {
		limiter.AddAndLimit(ctx)
		var command protocol.Message
		if err := wsjson.Read(ctx, connection, &command); err != nil {
			return
		}
		if command.Type == "version" {
			// Versions are informational until compatibility rules are defined.
			continue
		}
		if command.Type == "list-rooms" {
			user.send(protocol.Message{Type: "list-rooms", ID: command.ID, Rooms: manager.ListPublicLobbies()})
			continue
		}
		// Password hashing/comparison is deliberately outside the manager lock.
		// Room passwords are immutable after creation. Keep the room identity
		// too, since a deleted room's name can be reused during verification.
		var joiningLobby *Lobby
		var passwordHash []byte
		if command.Type == "create-room" && command.Password != "" {
			var err error
			passwordHash, err = bcrypt.GenerateFromPassword([]byte(command.Password), bcrypt.DefaultCost)
			if err != nil {
				user.send(protocol.Message{ID: command.ID, Room: strings.TrimSpace(command.Room), Error: "Room password must be at most 72 bytes"})
				continue
			}
		}
		if command.Type == "join-room" {
			joiningLobby, _ = manager.lock.With2(func() (*Lobby, error) {
				return manager.Lobbies[command.Room], nil
			})
			if joiningLobby != nil && len(joiningLobby.passwordHash) != 0 && bcrypt.CompareHashAndPassword(joiningLobby.passwordHash, []byte(command.Password)) != nil {
				user.send(protocol.Message{ID: command.ID, Error: "Incorrect room password"})
				continue
			}
		}
		manager.lock.With(func() error {
			if manager.handleInvitationCommand(user, connectionID, command) {
				return nil
			}
			if manager.handleGameCommand(user, connectionID, command) {
				return nil
			}
			switch command.Type {
			case "set-name":
				name := strings.TrimSpace(command.Name)
				if !validName(name) {
					user.send(protocol.Message{ID: command.ID, Error: "Name must contain 1–30 characters without control characters"})
				} else {
					user.name = name
					message := protocol.Message{Type: "name-updated", ID: command.ID, ClientID: connectionID, Room: user.room, Name: name}
					if lobby := manager.Lobbies[user.room]; lobby != nil {
						broadcast(lobby, message)
					} else {
						user.send(message)
					}
					if game := manager.Games[user.gameID]; game != nil {
						game.broadcastState()
						manager.broadcastGames(user.room)
					}
				}
			case "create-room":
				room := strings.TrimSpace(command.Room)
				if room == "" || !utf8.ValidString(room) || utf8.RuneCountInString(room) > protocol.MaxRoomNameLength || strings.ContainsFunc(room, unicode.IsControl) {
					user.send(protocol.Message{ID: command.ID, Room: room, Error: "Room name must contain 1–30 characters without control characters"})
				} else if _, exists := manager.Lobbies[room]; exists {
					user.send(protocol.Message{ID: command.ID, Room: room, Error: "Room already exists"})
				} else if len(manager.Lobbies) >= MaxRooms {
					user.send(protocol.Message{ID: command.ID, Room: room, Error: "Server room limit reached (1000 rooms)"})
				} else {
					manager.Lobbies[room] = &Lobby{Public: !command.Private, Name: room, Users: make(map[uint64]*User), passwordHash: passwordHash}
					visibility := "public"
					if command.Private {
						visibility = "private"
					}
					log.Printf("Client %d (%q) created room %q (%s, password_protected=%t)", connectionID, user.name, room, visibility, len(passwordHash) != 0)
					user.send(protocol.Message{Type: "create-room", ID: command.ID, Room: room})
				}
			case "join-room":
				lobby := manager.Lobbies[command.Room]
				if lobby == nil || lobby != joiningLobby {
					user.send(protocol.Message{ID: command.ID, Error: "Room not found"})
				} else if len(lobby.Users) >= 100 && user.room != command.Room {
					user.send(protocol.Message{ID: command.ID, Error: "Room is full"})
				} else {
					if user.room == lobby.Name {
						user.send(protocol.Message{Type: "join-room", ID: command.ID, Room: lobby.Name, ClientID: connectionID, Games: manager.waitingGames(lobby.Name)})
						broadcastUsers(lobby)
						if game := manager.Games[user.gameID]; game != nil {
							game.broadcastState()
						}
						return nil
					}
					manager.leave(user, connectionID)
					user.room = lobby.Name
					// Notify existing members before inserting the joining player.
					broadcast(lobby, protocol.Message{Type: "user-joined", Room: lobby.Name, ClientID: connectionID, Name: user.name})
					lobby.Users[connectionID] = user
					log.Printf("Client %d (%q) joined room %q", connectionID, user.name, lobby.Name)
					user.send(protocol.Message{Type: "join-room", ID: command.ID, Room: lobby.Name, ClientID: connectionID, Games: manager.waitingGames(lobby.Name)})
					broadcastUsers(lobby)
				}
			case "leave-room":
				manager.leave(user, connectionID)
				user.send(protocol.Message{Type: "leave-room", ID: command.ID})
			case "chat":
				text := strings.TrimSpace(command.Text)
				if lobby := manager.Lobbies[user.room]; lobby == nil {
					user.send(protocol.Message{ID: command.ID, Error: "Join a room first"})
				} else if text == "" || len(text) > 2000 {
					user.send(protocol.Message{ID: command.ID, Error: "Chat must contain 1–2000 bytes"})
				} else {
					broadcast(lobby, protocol.Message{Type: "chat", Room: lobby.Name, ClientID: connectionID, Text: text})
				}
			default:
				user.send(protocol.Message{ID: command.ID, Error: fmt.Sprintf("Unknown type '%s'", command.Type)})
			}
			return nil
		})
	}
}

func InitMux(mux *http.ServeMux, authKey string) {
	manager := &Manager{authKey: strings.TrimSpace(authKey), Lobbies: map[string]*Lobby{
		"Public": {Public: true, Name: "Public", Users: make(map[uint64]*User), persistent: true},
	}}
	mux.HandleFunc("/multiplayer/combat", manager.acceptCombat)
	var counter atomic.Uint64
	mux.HandleFunc("/multiplayer", func(writer http.ResponseWriter, request *http.Request) {
		connection, err := websocket.Accept(writer, request, &websocket.AcceptOptions{
			InsecureSkipVerify: true, // Allow browser clients hosted on any origin.
		})
		if err != nil {
			log.Printf("Error accepting websocket connection: %v", err)
			return
		}
		defer connection.CloseNow()
		handleWebsocket(connection, manager, counter.Add(1))
	})
}

// vim: set noexpandtab tabstop=4 shiftwidth=4:
