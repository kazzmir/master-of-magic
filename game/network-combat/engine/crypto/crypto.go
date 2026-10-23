package crypto

import (
    "log"
    "fmt"
    "crypto/ecdh"
    "crypto/hkdf"
    "crypto/sha256"
    "crypto/cipher"
    "crypto/aes"
)

type PacketConn interface {
    Read() ([]byte, error)
    Send([]byte) error
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

type encryptedConn struct {
    conn PacketConn
    sendCipher cipher.AEAD
    receiveCipher cipher.AEAD
}

func (encrypted *encryptedConn) Read() ([]byte, error) {
    encryptedData, err := encrypted.conn.Read()
    if err != nil {
        return nil, err
    }

    return encrypted.receiveCipher.Open(nil, nil, encryptedData, nil)
}

func (encrypted *encryptedConn) Send(data []byte) error {
    encryptedData := encrypted.sendCipher.Seal(nil, nil, data, nil)
    return encrypted.conn.Send(encryptedData)
}

func NewEncryptedConn(conn PacketConn) (PacketConn, error) {
    privateKey, err := ecdh.X25519().GenerateKey(nil)
    if err != nil {
        return nil, fmt.Errorf("Unable to generate private x25519 key: %v", err)
    }

    go func() {
        localPublicKey := privateKey.PublicKey()
        err = conn.Send(localPublicKey.Bytes())
        if err != nil {
            log.Printf("Error sending public key: %v", err)
        }
    }()

    var sendCipher cipher.AEAD
    var receiveCipher cipher.AEAD

    publicKeyBytes, err := conn.Read()
    if err != nil {
        return nil, err
    }
    log.Printf("Received public key from peer: %x", publicKeyBytes)

    peerPublicKey, err := ecdh.X25519().NewPublicKey(publicKeyBytes)
    if err != nil {
        return nil, err
    }

    sharedSecret, err := privateKey.ECDH(peerPublicKey)
    if err != nil {
        return nil, err
    }

    sendCipher, err = deriveCipher(sharedSecret, privateKey.PublicKey().Bytes(), peerPublicKey.Bytes())
    if err != nil {
        return nil, err
    }

    receiveCipher, err = deriveCipher(sharedSecret, peerPublicKey.Bytes(), privateKey.PublicKey().Bytes())
    if err != nil {
        return nil, err
    }

    return &encryptedConn{
        conn: conn,
        sendCipher: sendCipher,
        receiveCipher: receiveCipher,
    }, nil
}
