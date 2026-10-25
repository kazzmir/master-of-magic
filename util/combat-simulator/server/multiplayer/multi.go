package multiplayer

import (
    "fmt"
    "errors"
    "slices"
    "time"
    "log"
    "context"
    "encoding/json/v2"
    "net/http"

    netlib "github.com/kazzmir/master-of-magic/lib/net"

    "github.com/coder/websocket"
    "github.com/coder/websocket/wsjson"
)

// have a list of lobbies, public and private
// users in a lobby can chat with each other
// anyone in a lobby can host a game, and someone else in the same lobby can join
// public and private lobbies can have passwords
// maximum of 100 players in a lobby

type WebsocketError struct {
    Error string `json:"error"`
}

func isValidAuth(auth string) bool {
    return true
}

type WebsocketMessage struct {
    Type string `json:"type"`
    Id uint64 `json:"id"`
}

type User struct {
}

type Lobby struct {
    Public bool
    Name string
    Users map[string]*User
}

type Manager struct {
    // map from name to lobby
    Lobbies map[string]*Lobby
}

func (manager *Manager) ListPublicLobbies() []string {
    out := make([]string, 0, len(manager.Lobbies))
    for _, lobby := range manager.Lobbies {
        if lobby.Public {
            out = append(out, lobby.Name)
        }
    }

    slices.Sort(out)
    return out
}

func handleWebsocket(connection *websocket.Conn, manager *Manager) {
    stop := false
    needAuth := true

    rateLimiter := netlib.RateLimiter{
        WindowSize: 1 * time.Second,
        MaxEvents: 2,
    }

    mainQuit, mainCancel := context.WithCancel(context.Background())
    defer mainCancel()

    for !stop {
        f := func() error {
            rateLimiter.AddAndLimit(mainQuit)

            var message map[string]any

            var quit context.Context
            var cancel context.CancelFunc

            if needAuth {
                quit, cancel = context.WithTimeout(context.Background(), 10*time.Second)
                defer cancel()
            } else {
                quit, cancel = mainQuit, mainCancel
            }

            err := wsjson.Read(quit, connection, &message)
            if err != nil {
                if websocket.CloseStatus(err) == websocket.StatusNoStatusRcvd {
                    stop = true
                    return nil
                }

                return fmt.Errorf("Error reading websocket message: %v", err)
            }

            if len(message) == 0 {
                reason := "Empty message"
                wsjson.Write(quit, connection, WebsocketError{Error: reason})
                return errors.New(reason)
            }

            if needAuth {
                auth, ok := message["auth"].(string)
                if !ok || !isValidAuth(auth) {
                    reason := "Missing or invalid auth field"
                    wsjson.Write(quit, connection, WebsocketError{Error: reason})
                    return errors.New(reason)
                }
                needAuth = false
                return nil
            }

            marshalled, err := json.Marshal(message)
            if err != nil {
                wsjson.Write(quit, connection, WebsocketError{Error: "Internal error"})
                return fmt.Errorf("Error marshalling websocket message: %v", err)
            }

            var command WebsocketMessage
            err = json.Unmarshal(marshalled, &command)
            if err != nil {
                wsjson.Write(quit, connection, WebsocketError{Error: "Invalid message format"})
                return fmt.Errorf("Error unmarshalling websocket message: %v", err)
            }

            switch command.Type {
                case "list-rooms":
                    publicLobbies := manager.ListPublicLobbies()
                    wsjson.Write(quit, connection, map[string]any{
                        "type": "list-rooms",
                        "id": command.Id,
                        "rooms": publicLobbies,
                    })
                case "join-room":
                    // TODO
                    // if room is password protected then require a password
                case "create-room":
                    // TODO
                case "leave-room":
                    // TODO
                default:
                    wsjson.Write(quit, connection, WebsocketError{Error: fmt.Sprintf("Unknown type '%s'", command.Type)})
            }

            // log.Printf("Received websocket message: %v", message)
            return nil
        }

        err := f()
        if err != nil {
            log.Printf("Error: %v", err)
            return
        }
    }
}

func InitMux(mux *http.ServeMux) {
    var manager Manager
    mux.HandleFunc("/multiplayer", func(writer http.ResponseWriter, request *http.Request) {
        connection, err := websocket.Accept(writer, request, nil)
        if err != nil {
            log.Printf("Error accepting websocket connection: %v", err)
            return
        }

        defer connection.CloseNow()

        log.Printf("Websocket connection established from %s", request.RemoteAddr)
        defer log.Printf("Websocket connection closed from %s", request.RemoteAddr)

        handleWebsocket(connection, &manager)
    })
}
