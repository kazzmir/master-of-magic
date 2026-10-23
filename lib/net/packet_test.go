package net

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	gonet "net"
	"strings"
	"testing"
	"time"
)

func packetTestPipe(t *testing.T) (gonet.Conn, gonet.Conn) {
	t.Helper()
	left, right := gonet.Pipe()
	t.Cleanup(func() { left.Close() })
	t.Cleanup(func() { right.Close() })
	for _, conn := range []gonet.Conn{left, right} {
		if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
			t.Fatalf("set deadline: %v", err)
		}
	}
	return left, right
}

func TestPacketNetConnRoundTrip(t *testing.T) {
	left, right := packetTestPipe(t)
	sender := NewPacketNetConn(left, 16)
	receiver := NewPacketNetConn(right, 16)
	messages := [][]byte{[]byte("hello"), {}, {0, 1, 128, 255}, bytes.Repeat([]byte{42}, 16)}

	sent := make(chan error, 1)
	go func() {
		for _, message := range messages {
			if err := sender.Send(message); err != nil {
				sent <- err
				return
			}
		}
		sent <- nil
	}()

	for i, want := range messages {
		got, err := receiver.Read()
		if err != nil {
			t.Fatalf("read packet %d: %v", i, err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("packet %d: got %x, want %x", i, got, want)
		}
	}
	if err := <-sent; err != nil {
		t.Fatalf("send packets: %v", err)
	}
}

func TestPacketNetConnRead(t *testing.T) {
    // the test bypasses the Send method so we manually insert the prefix length header
	frame := func(length uint32, payload []byte) []byte {
		data := make([]byte, 4)
		binary.BigEndian.PutUint32(data, length)
		return append(data, payload...)
	}

	for _, test := range []struct {
		name    string
		wire    []byte
		want    []byte
		wantErr error
		tooBig  bool
	}{
		{name: "fragmented packet", wire: frame(3, []byte("abc")), want: []byte("abc")},
		{name: "empty packet", wire: frame(0, nil), want: []byte{}},
		{name: "maximum size", wire: frame(4, []byte("abcd")), want: []byte("abcd")},
		{name: "oversized packet", wire: frame(5, nil), tooBig: true},
		{name: "closed connection", wantErr: io.EOF},
		{name: "truncated header", wire: []byte{0, 0}, wantErr: io.ErrUnexpectedEOF},
		{name: "truncated payload", wire: frame(3, []byte("ab")), wantErr: io.ErrUnexpectedEOF},
	} {
		t.Run(test.name, func(t *testing.T) {
			writer, reader := packetTestPipe(t)
			go func() {
				defer writer.Close()
				// Separate writes force the reader to assemble fragmented data.
				for _, b := range test.wire {
					if _, err := writer.Write([]byte{b}); err != nil {
						return
					}
				}
			}()

			got, err := NewPacketNetConn(reader, 4).Read()
			if test.tooBig {
				if err == nil || !strings.Contains(err.Error(), "exceeds maximum allowed size") {
					t.Fatalf("expected packet size error, got %v", err)
				}
				return
			}
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("read error: got %v, want %v", err, test.wantErr)
			}
			if !bytes.Equal(got, test.want) {
				t.Fatalf("read payload: got %x, want %x", got, test.want)
			}
		})
	}
}
