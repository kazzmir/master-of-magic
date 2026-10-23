package engine

import (
    "flag"
    "log"
    "fmt"
    "time"
    "net"
    "context"
    "sync"
    "encoding/binary"
    "io"
    "os"
    "os/signal"
    "syscall"

    "github.com/kazzmir/master-of-magic/game/network-combat/engine/crypto"
)

type PacketNetConn struct {
    net.Conn
}

const MAX_PACKET_SIZE = 1024 * 1024 // 1 MB

func (packet *PacketNetConn) Read() ([]byte, error) {
    // read 4 byte length prefix
    var length uint32
    err := binary.Read(packet.Conn, binary.BigEndian, &length)
    if err != nil {
        return nil, err
    }

    if length > MAX_PACKET_SIZE {
        return nil, fmt.Errorf("Packet size %d exceeds maximum allowed size %d", length, MAX_PACKET_SIZE)
    }

    data := make([]byte, length)
    _, err = io.ReadFull(packet.Conn, data)
    if err != nil {
        return nil, err
    }

    return data, nil
}

func (packet *PacketNetConn) Send(data []byte) error {
    length := uint32(len(data))
    err := binary.Write(packet.Conn, binary.BigEndian, length)
    if err != nil {
        return err
    }

    _, err = packet.Conn.Write(data)
    if err != nil {
        return err
    }

    return nil
}

func doRun(peer net.Conn, quit context.Context) error {
    defer peer.Close()

    go func() {
        <-quit.Done()
        log.Printf("Shutting down connection to %s", peer.RemoteAddr())
        peer.Close()
    }()

    packetConn := PacketNetConn{peer}

    didExchange := make(chan crypto.PacketConn)

    go func() {
        log.Printf("Starting key exchange with %s", peer.RemoteAddr())
        encryptedConn, err := crypto.NewEncryptedConn(&packetConn)
        if err != nil {
            log.Printf("Failed to create encrypted connection: %v", err)
            return
        }
        didExchange <- encryptedConn
    }()

    var encryptedConn crypto.PacketConn

    select {
        case encryptedConn = <-didExchange:
            log.Printf("Key exchange completed with %s", peer.RemoteAddr())
        case <-quit.Done():
            return fmt.Errorf("Connection closed before key exchange completed")
        case <-time.After(5 * time.Second):
            return fmt.Errorf("Timeout waiting for key exchange to complete")
    }

    go encryptedConn.Send([]byte("Hello from this side!"))
    message, err := encryptedConn.Read()
    if err != nil {
        return err
    }
    
    log.Printf("Received message: %s", string(message))

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
