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
    "flag"

    "github.com/kazzmir/master-of-magic/util/combat-simulator/server/multiplayer"

    "golang.org/x/crypto/acme/autocert"
    // "golang.org/x/crypto/acme"

    "github.com/sendgrid/sendgrid-go"
    "github.com/sendgrid/sendgrid-go/helpers/mail"
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

func runServer(certificateFunc TLSGetCertificateFunc) error {
    key, err := loadKey()
    if err != nil {
        return err
    }
    if key == "" {
        return errors.New("key/key.txt must contain a non-empty auth key")
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

    multiplayer.InitMux(mux, key)

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
    log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
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
