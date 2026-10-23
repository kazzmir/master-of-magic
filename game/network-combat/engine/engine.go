package engine

import (
    "flag"
    "log"
    "fmt"
    "time"
    "net"
    "context"
    "sync"
    "os"
    "os/signal"
    "syscall"
)

func doRun(peer net.Conn, quit context.Context) error {
    return nil
}

func doListen(port int, quit context.Context) error {
    listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        return err
    }

    defer listen.Close()

    ok := make(chan struct{})
    defer close(ok)

    stop := false

    go func() {
        select {
            case <-quit.Done():
                stop = true
                log.Printf("Shutting down listener")
                listen.Close()
            case <-ok:
        }
    }()

    for !stop {
        peer, err := listen.Accept()
        if err != nil {
            if stop {
                break
            }

            log.Printf("Failed to accept connection: %v", err)
            continue
        }

        log.Printf("Accepted connection from %s", peer.RemoteAddr())

        listen.Close()
        return doRun(peer, quit)
    }

    return fmt.Errorf("Listener stopped")
}

func doConnect(peer string, quit context.Context) error {
    conn, err := net.Dial("tcp", peer)
    if err != nil {
        return err
    }

    return doRun(conn, quit)
}

func Launch() {
    log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

    port := flag.Int("port", 0, "Port to listen on")
    peer := flag.String("peer", "", "Peer address to connect to")
    flag.Parse()

    quit, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()

    var wait sync.WaitGroup

    if *port != 0 {
        log.Printf("Listen for connections on port %d", *port)
        err := doListen(*port, quit)
        if err != nil {
            log.Fatalf("Failed to listen on port %d: %v", *port, err)
        }
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
