package engine

import (
    "flag"
    "log"
    "time"
    "context"
    "sync"
    "os"
    "os/signal"
    "syscall"
)

func doListen(port int, quit context.Context) {
}

func doConnect(peer string, quit context.Context) {
}

func Launch() {
    log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

    port := flag.Int("port", 0, "Port to listen on")
    peer := flag.String("peer", "", "Peer address to connect to")
    flag.Parse()

    quit, cancel := context.WithCancel(context.Background())
    defer cancel()

    signal.NotifyContext(quit, os.Interrupt, syscall.SIGTERM)

    var wait sync.WaitGroup

    if *port != 0 {
        log.Printf("Listen for connections on port %d", *port)
        doListen(*port, quit)
    } else if *peer != "" {
        // connect to peer
        doConnect(*peer, quit)
    } else {
        log.Fatal("Either -port or -peer must be specified")
    }

    done := make(chan struct{})
    go func() {
        wait.Wait()
        close(done)
    }()

    select {
        case <-done:
        case <-time.After(2 * time.Second):
            log.Println("Timeout waiting for goroutines to finish")
    }

    log.Printf("Shutting down")
}
