package net

import (
	"context"
	"io"
	gonet "net"
	"os"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type packetResult struct {
	data []byte
	err  error
}

// WebsocketPacketConn preserves packet framing over a WebSocket byte stream.
// Read deadlines are logical: a temporary timeout must not kill the WebSocket
// when combat switches from its army-exchange receive loop to its game loop.
type WebsocketPacketConn struct {
	packet     *PacketNetConn
	connection *websocket.Conn
	ctx        context.Context
	cancel     context.CancelFunc
	results    chan packetResult
	done       chan struct{}
	once       sync.Once
	sendLock   sync.Mutex
	lock       sync.Mutex
	deadline   time.Time
	changed    chan struct{}
	err        error
}

func NewWebsocketPacketConn(ctx context.Context, connection *websocket.Conn, maxPacketSize uint32) *WebsocketPacketConn {
	ctx, cancel := context.WithCancel(ctx)
	conn := &WebsocketPacketConn{connection: connection, ctx: ctx, cancel: cancel, results: make(chan packetResult, 16), done: make(chan struct{}), changed: make(chan struct{})}
	conn.packet = NewPacketNetConn(websocket.NetConn(ctx, connection, websocket.MessageBinary), maxPacketSize)
	go conn.receive()
	return conn
}

func (conn *WebsocketPacketConn) finish(err error) {
	conn.once.Do(func() {
		conn.lock.Lock()
		conn.err = err
		conn.lock.Unlock()
		close(conn.done)
		conn.cancel()
		conn.connection.CloseNow()
	})
}

func (conn *WebsocketPacketConn) receive() {
	defer close(conn.results)
	for {
		data, err := conn.packet.Read()
		if err != nil {
			conn.finish(err)
			return
		}
		select {
		case conn.results <- packetResult{data: data}:
		case <-conn.ctx.Done():
			conn.finish(conn.ctx.Err())
			return
		}
	}
}

func (conn *WebsocketPacketConn) Done() <-chan struct{} { return conn.done }
func (conn *WebsocketPacketConn) Err() error {
	conn.lock.Lock()
	defer conn.lock.Unlock()
	return conn.err
}
func (conn *WebsocketPacketConn) Close() { conn.finish(gonet.ErrClosed) }

func (conn *WebsocketPacketConn) Send(data []byte) error {
	conn.sendLock.Lock()
	defer conn.sendLock.Unlock()
	err := conn.packet.Send(data)
	if err != nil {
		conn.finish(err)
	}
	return err
}

func (conn *WebsocketPacketConn) SetReadDeadline(deadline time.Time) error {
	conn.lock.Lock()
	defer conn.lock.Unlock()
	conn.deadline = deadline
	close(conn.changed)
	conn.changed = make(chan struct{})
	return nil
}

func (conn *WebsocketPacketConn) Read() ([]byte, error) {
	for {
		conn.lock.Lock()
		deadline, changed := conn.deadline, conn.changed
		conn.lock.Unlock()
		var timer *time.Timer
		var timeout <-chan time.Time
		if !deadline.IsZero() {
			if !time.Now().Before(deadline) {
				return nil, os.ErrDeadlineExceeded
			}
			timer = time.NewTimer(time.Until(deadline))
			timeout = timer.C
		}
		var result packetResult
		var ok bool
		select {
		case result, ok = <-conn.results:
		case <-changed:
			if timer != nil {
				timer.Stop()
			}
			continue
		case <-timeout:
			return nil, os.ErrDeadlineExceeded
		case <-conn.ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			return nil, gonet.ErrClosed
		}
		if timer != nil {
			timer.Stop()
		}
		if !ok {
			return nil, io.EOF
		}
		return result.data, result.err
	}
}
