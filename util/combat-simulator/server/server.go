package main

import (
    "log"
    "net/http"
    "fmt"
    "time"
    "crypto/tls"

    "golang.org/x/crypto/acme/autocert"
    // "golang.org/x/crypto/acme"
)

func runServer(certManager *autocert.Manager){
    log.Printf("HTTPS server listening on :5000")

    mux := http.NewServeMux()
    mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
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

    log.Fatal(server.ListenAndServeTLS("", ""))
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

    runServer(&certManager)
}
