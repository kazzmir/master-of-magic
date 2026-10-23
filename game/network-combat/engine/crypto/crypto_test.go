package crypto

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

type testPacketConn struct {
	incoming chan []byte
	outgoing chan []byte
}

func (conn *testPacketConn) Read() ([]byte, error) {
	select {
	case packet := <-conn.incoming:
		return packet, nil
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timed out reading packet")
	}
}

func (conn *testPacketConn) Send(packet []byte) error {
	select {
	case conn.outgoing <- bytes.Clone(packet):
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timed out sending packet")
	}
}

func TestNewEncryptedConn(t *testing.T) {
	leftToRight := make(chan []byte, 1)
	rightToLeft := make(chan []byte, 1)
	leftTransport := &testPacketConn{incoming: rightToLeft, outgoing: leftToRight}
	rightTransport := &testPacketConn{incoming: leftToRight, outgoing: rightToLeft}

	type result struct {
		conn PacketConn
		err  error
	}
	leftResult := make(chan result, 1)
	go func() {
		conn, err := NewEncryptedConn(leftTransport)
		leftResult <- result{conn: conn, err: err}
	}()

	right, err := NewEncryptedConn(rightTransport)
	if err != nil {
		t.Fatalf("create right encrypted connection: %v", err)
	}
	left := <-leftResult
	if left.err != nil {
		t.Fatalf("create left encrypted connection: %v", left.err)
	}

	for _, direction := range []struct {
		name      string
		sender    PacketConn
		receiver  PacketConn
		transport *testPacketConn
	}{
		{"left to right", left.conn, right, rightTransport},
		{"right to left", right, left.conn, leftTransport},
	} {
		t.Run(direction.name, func(t *testing.T) {
			for _, message := range [][]byte{[]byte("hello, encrypted peer"), {}, {0, 1, 127, 128, 255}} {
				if err := direction.sender.Send(message); err != nil {
					t.Fatalf("send packet: %v", err)
				}

				wire, err := direction.transport.Read()
				if err != nil {
					t.Fatalf("read wire packet: %v", err)
				}
				if bytes.Equal(wire, message) {
					t.Fatal("wire packet is not encrypted")
				}
				// Put the captured packet back so the encrypted connection can read it.
				direction.transport.incoming <- wire

				got, err := direction.receiver.Read()
				if err != nil {
					t.Fatalf("decrypt packet: %v", err)
				}
				if !bytes.Equal(got, message) {
					t.Fatalf("received %x, want %x", got, message)
				}
			}
		})
	}
}
