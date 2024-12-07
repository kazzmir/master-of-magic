package main

import (
    "log"
    "net/http"
    "fmt"
    "crypto/tls"

    "golang.org/x/crypto/acme/autocert"
    // "golang.org/x/crypto/acme"
)

func main(){
    log.Printf("Server starting")

    // acmeClient := &acme.Client{DirectoryURL: "https://acme-staging-v02.api.letsencrypt.org/directory"}

    certManager := autocert.Manager{
        Prompt: autocert.AcceptTOS,
        HostPolicy: autocert.HostWhitelist("magic.jonrafkind.com"),
        // Client: acmeClient,
        Cache: autocert.DirCache("certs"),
    }

    http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
        fmt.Fprintf(writer, "Hello, World!")
    })

    server := &http.Server{
        Addr: ":5000",
        TLSConfig: &tls.Config{
            GetCertificate: certManager.GetCertificate,
            MinVersion: tls.VersionTLS12,
        },
    }

    go http.ListenAndServe(":5001", certManager.HTTPHandler(nil))

    log.Fatal(server.ListenAndServeTLS("", ""))
}
