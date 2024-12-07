package main

import (
    "log"
    "net/http"
    "fmt"
    "time"
    "crypto/tls"
    "embed"
    "strings"

    "golang.org/x/crypto/acme/autocert"
    // "golang.org/x/crypto/acme"
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

func runServer(certManager *autocert.Manager) error {
    log.Printf("HTTPS server listening on :5000")

    key, err := loadKey()
    if err != nil {
        return err
    }

    mux := http.NewServeMux()
    mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
        fmt.Fprintf(writer, "OK")
    })

    mux.HandleFunc("POST /report", func(writer http.ResponseWriter, request *http.Request) {
        apiKey := request.Header.Get("X-Report-Key")
        if apiKey != key {
            http.Error(writer, "Unauthorized", http.StatusUnauthorized)
            return
        }

        log.Printf("Received report from %s", request.RemoteAddr)

        fmt.Fprintf(writer, "OK")
    })

    server := &http.Server{
        Addr: ":5000",
        Handler: mux,
        ReadTimeout: 10 * time.Second,
        WriteTimeout: 10 * time.Second,
        TLSConfig: &tls.Config{
            GetCertificate: certManager.GetCertificate,
            MinVersion: tls.VersionTLS12,
        },
    }

    return server.ListenAndServeTLS("", "")
}

func main(){
    log.Printf("Server starting")

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

    err := runServer(&certManager)
    if err != nil {
        log.Fatalf("Error running server: %v", err)
    }
}
