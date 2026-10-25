package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/hkdf"
	"crypto/sha256"
	"fmt"
	"log"
	"sync"
	"time"

	netlib "github.com/kazzmir/master-of-magic/lib/net"
)

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

// the actual limit is 2^32
const MAX_ENCRYPTED_PACKETS = 1 << 31

type encryptedConn struct {
	conn            netlib.PacketConn
	sendCipher      cipher.AEAD
	sendLock        sync.Mutex
	sentPackets     uint64
	receiveCipher   cipher.AEAD
	receivedPackets uint64
}

func (encrypted *encryptedConn) SetReadDeadline(deadline time.Time) error {
	return encrypted.conn.SetReadDeadline(deadline)
}

func (encrypted *encryptedConn) Read() ([]byte, error) {
	encryptedData, err := encrypted.conn.Read()
	if err != nil {
		return nil, err
	}

	encrypted.receivedPackets += 1
	if encrypted.receivedPackets > MAX_ENCRYPTED_PACKETS {
		return nil, fmt.Errorf("Maximum number of encrypted packets exceeded")
	}

	return encrypted.receiveCipher.Open(nil, nil, encryptedData, nil)
}

func (encrypted *encryptedConn) Send(data []byte) error {
	encrypted.sendLock.Lock()
	defer encrypted.sendLock.Unlock()
	encrypted.sentPackets += 1
	if encrypted.sentPackets > MAX_ENCRYPTED_PACKETS {
		return fmt.Errorf("Maximum number of encrypted packets exceeded")
	}

	encryptedData := encrypted.sendCipher.Seal(nil, nil, data, nil)
	return encrypted.conn.Send(encryptedData)
}

func NewEncryptedConn(conn netlib.PacketConn) (netlib.PacketConn, error) {
	privateKey, err := ecdh.X25519().GenerateKey(nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to generate private x25519 key: %v", err)
	}

	sentKey := make(chan error, 1)
	go func() {
		sentKey <- conn.Send(privateKey.PublicKey().Bytes())
	}()

	var sendCipher cipher.AEAD
	var receiveCipher cipher.AEAD

	publicKeyBytes, err := conn.Read()
	if err != nil {
		return nil, err
	}
	if err := <-sentKey; err != nil {
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
		conn:          conn,
		sendCipher:    sendCipher,
		receiveCipher: receiveCipher,
	}, nil
}
