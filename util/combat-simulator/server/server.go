package main

import (
    "log"
    "io"
    "errors"
    "bytes"
    "net/http"
    "fmt"
    "time"
    "crypto/tls"
    "embed"
    "strings"
    "context"
    "flag"
    "slices"
    "encoding/json"

    "golang.org/x/crypto/acme/autocert"
    // "golang.org/x/crypto/acme"

    "github.com/sendgrid/sendgrid-go"
    "github.com/sendgrid/sendgrid-go/helpers/mail"
    "github.com/coder/websocket"
    "github.com/coder/websocket/wsjson"
)

//go:embed key/*
var keys embed.FS

func loadKey() (string, error) {
    keyBytes, err := keys.ReadFile("key/key.txt")
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(keyBytes)), nil
}

func loadSendGridKey() (string, error) {
    keyBytes, err := keys.ReadFile("key/sendgrid")
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(keyBytes)), nil
}

// send an email to me with the report
func doSendEmail(report string) {
    apiKey, err := loadSendGridKey()
    if err != nil {
        log.Printf("Unable to send email: %v", err)
        return
    }

    fromAddress := "magic@jonrafkind.com"
    toAddress := "jon@rafkind.com"

    from := mail.NewEmail("Magic", fromAddress)
    to := mail.NewEmail("Me", toAddress)

    subject := "Magic Combat Simulator Bug Report"

    // replace \n with <br>
    replaceNewline := func(s string) string {
        return strings.ReplaceAll(s, "\n", "<br>")
    }

    message := mail.NewSingleEmail(from, subject, to, report, replaceNewline(report))
    client := sendgrid.NewSendClient(apiKey)
    response, err := client.Send(message)
    if err != nil {
        log.Printf("Unable to send email: %v", err)
        return
    } else {
        log.Printf("Email sent: %v", response)
    }
}

type TLSGetCertificateFunc func(hello *tls.ClientHelloInfo) (*tls.Certificate, error)

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

type MultiplayerManager struct {
    // map from name to lobby
    Lobbies map[string]*Lobby
}

func (manager *MultiplayerManager) ListPublicLobbies() []string {
    out := make([]string, 0, len(manager.Lobbies))
    for _, lobby := range manager.Lobbies {
        if lobby.Public {
            out = append(out, lobby.Name)
        }
    }

    slices.Sort(out)
    return out
}

func handleWebsocket(connection *websocket.Conn, multiplayerManager *MultiplayerManager) {
    stop := false
    needAuth := true
    for !stop {
        f := func() error {
            var message map[string]any
            quit, cancel := context.WithTimeout(context.Background(), 10*time.Second)
            defer cancel()

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
                    publicLobbies := multiplayerManager.ListPublicLobbies()
                    wsjson.Write(quit, connection, map[string]any{
                        "type": "list-rooms",
                        "id": command.Id,
                        "rooms": publicLobbies,
                    })
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

func runServer(certificateFunc TLSGetCertificateFunc) error {
    key, err := loadKey()
    if err != nil {
        return err
    }

    mux := http.NewServeMux()
    mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
        fmt.Fprintf(writer, "OK")
    })

    mux.HandleFunc("OPTIONS /report", func(writer http.ResponseWriter, request *http.Request) {
        writer.Header().Set("Access-Control-Allow-Origin", "*")
        writer.Header().Set("Access-Control-Allow-Methods", "POST")
        writer.Header().Set("Access-Control-Allow-Headers", "X-Report-Key")
        writer.Header().Set("Access-Control-Max-Age", "86400")
        writer.WriteHeader(http.StatusOK)
    })

    mux.HandleFunc("POST /report", func(writer http.ResponseWriter, request *http.Request) {
        apiKey := request.Header.Get("X-Report-Key")
        if apiKey != key {
            http.Error(writer, "Unauthorized", http.StatusUnauthorized)
            return
        }

        log.Printf("Received report from %s", request.RemoteAddr)

        var buffer bytes.Buffer
        _, err := io.CopyN(&buffer, request.Body, 1 << 16)
        if err == nil || errors.Is(err, io.EOF) {
            log.Printf("Report: %v", buffer.String())

            go func() {
                doSendEmail(buffer.String())
            }()

        }

        writer.Header().Set("Access-Control-Allow-Origin", "*")
        writer.Header().Set("Access-Control-Allow-Methods", "POST")
        writer.Header().Set("Access-Control-Allow-Headers", "X-Report-Key")
        writer.Header().Set("Access-Control-Max-Age", "86400")
        fmt.Fprintf(writer, "OK")
    })

    // have a list of lobbies, public and private
    // users in a lobby can chat with each other
    // anyone in a lobby can host a game, and someone else in the same lobby can join
    // public and private lobbies can have passwords
    // maximum of 100 players in a lobby

    var multiplayerManager MultiplayerManager

    mux.HandleFunc("/multiplayer", func(writer http.ResponseWriter, request *http.Request) {
        connection, err := websocket.Accept(writer, request, nil)
        if err != nil {
            log.Printf("Error accepting websocket connection: %v", err)
            return
        }

        defer connection.CloseNow()

        log.Printf("Websocket connection established from %s", request.RemoteAddr)
        defer log.Printf("Websocket connection closed from %s", request.RemoteAddr)

        handleWebsocket(connection, &multiplayerManager)

    })

    server := &http.Server{
        Addr: ":5000",
        Handler: mux,
        ReadTimeout: 10 * time.Second,
        WriteTimeout: 10 * time.Second,
    }

    if certificateFunc != nil {
        server.TLSConfig = &tls.Config{
            GetCertificate: certificateFunc,
            MinVersion: tls.VersionTLS12,
        }

        log.Printf("HTTPS server listening on :5000")

        return server.ListenAndServeTLS("", "")
    } else {
        log.Printf("HTTP server listening on :5000")

        return server.ListenAndServe()
    }
}

func main(){
    log.Printf("Server starting")

    tls := flag.Bool("tls", true, "Run server with TLS")
    flag.Parse()

    if *tls {
        // acmeClient := &acme.Client{DirectoryURL: "https://acme-staging-v02.api.letsencrypt.org/directory"}

        certManager := autocert.Manager{
            Prompt: autocert.AcceptTOS,
            HostPolicy: autocert.HostWhitelist("magic.jonrafkind.com"),
            // Client: acmeClient,
            Cache: autocert.DirCache("certs"),
        }

        log.Printf("HTTP server listening on :5001")
        // lets encrypt stuff listens on http
        go http.ListenAndServe(":5001", certManager.HTTPHandler(nil))

        err := runServer(certManager.GetCertificate)
        if err != nil {
            log.Fatalf("Error running server: %v", err)
        }
    } else {
        err := runServer(nil)
        if err != nil {
            log.Fatalf("Error running server: %v", err)
        }
    }
}
