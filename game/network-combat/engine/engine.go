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
    "crypto/ecdh"
    "crypto/hkdf"
    "crypto/sha256"
    "crypto/cipher"
    "crypto/aes"
)

const MAX_PACKET_SIZE = 1024 * 1024 // 1 MB

func readBytes(conn net.Conn) ([]byte, error) {
    // read 4 byte length prefix
    var length uint32
    err := binary.Read(conn, binary.BigEndian, &length)
    if err != nil {
        return nil, err
    }

    if length > MAX_PACKET_SIZE {
        return nil, fmt.Errorf("Packet size %d exceeds maximum allowed size %d", length, MAX_PACKET_SIZE)
    }

    data := make([]byte, length)
    _, err = io.ReadFull(conn, data)
    if err != nil {
        return nil, err
    }

    return data, nil
}

func sendBytes(conn net.Conn, data []byte) error {
    length := uint32(len(data))
    err := binary.Write(conn, binary.BigEndian, length)
    if err != nil {
        return err
    }

    _, err = conn.Write(data)
    if err != nil {
        return err
    }

    return nil
}

func deriveCipher(secret []byte, from []byte, to []byte) (cipher.AEAD, error) {
    key, err := hkdf.Key(sha256.New, secret, nil, fmt.Sprintf("key/%s/%s", string(from), string(to)), 32)
    if err != nil {
        return nil, err
    }

    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    return cipher.NewGCMWithRandomNonce(block)
}

func doRun(peer net.Conn, quit context.Context) error {
    defer peer.Close()

    privateKey, err := ecdh.X25519().GenerateKey(nil)
    if err != nil {
        return fmt.Errorf("Unable to generate private x25519 key: %v", err)
    }

    go func() {
        <-quit.Done()
        log.Printf("Shutting down connection to %s", peer.RemoteAddr())
        peer.Close()
    }()

    var wait sync.WaitGroup

    wait.Go(func() {
        const STATE_READ_PUBLIC_KEY = 0
        const STATE_EXCHANCE_MESSAGES = 1
        state := STATE_READ_PUBLIC_KEY

        var sendCipher cipher.AEAD
        var receiveCipher cipher.AEAD

        for quit.Err() == nil {
            switch state {
                case STATE_READ_PUBLIC_KEY:
                    publicKeyBytes, err := readBytes(peer)
                    if err != nil {
                        log.Printf("Failed to read public key from peer: %v", err)
                        return
                    }
                    log.Printf("Received public key from peer: %x", publicKeyBytes)
                    state = STATE_EXCHANCE_MESSAGES

                    peerPublicKey, err := ecdh.X25519().NewPublicKey(publicKeyBytes)
                    if err != nil {
                        log.Printf("Failed to create public key from bytes: %v", err)
                        return
                    }

                    sharedSecret, err := privateKey.ECDH(peerPublicKey)
                    if err != nil {
                        log.Printf("Failed to compute shared secret: %v", err)
                        return
                    }

                    sendCipher, err = deriveCipher(sharedSecret, privateKey.PublicKey().Bytes(), peerPublicKey.Bytes())
                    if err != nil {
                        log.Printf("Failed to derive send cipher: %v", err)
                        return
                    }

                    receiveCipher, err = deriveCipher(sharedSecret, peerPublicKey.Bytes(), privateKey.PublicKey().Bytes())
                    if err != nil {
                        log.Printf("Failed to derive receive cipher: %v", err)
                        return
                    }

                    go func() {
                        encrypted := sendCipher.Seal(nil, nil, []byte("Hello from peer!"), nil)
                        sendBytes(peer, encrypted)
                    }()

                case STATE_EXCHANCE_MESSAGES:
                    message, err := readBytes(peer)
                    if err != nil {
                        log.Printf("Failed to read message from peer: %v", err)
                        return
                    }

                    decrypted, err := receiveCipher.Open(nil, nil, message, nil)
                    if err != nil {
                        log.Printf("Failed to decrypt message from peer: %v", err)
                        return
                    }

                    log.Printf("Decrypted message from peer: %s", string(decrypted))
            }
        }

    })

    localPublicKey := privateKey.PublicKey()
    err = sendBytes(peer, localPublicKey.Bytes())
    if err != nil {
        return err
    }

    wait.Wait()

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
