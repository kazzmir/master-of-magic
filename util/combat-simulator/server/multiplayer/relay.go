package multiplayer

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

const combatBytesPerSecond = 5 * 1024
const relayChunkSize = 1024

// A shared leaky-bucket clock throttles BOTH directions together. Forwarding
// small chunks avoids releasing a large, delayed packet as an unbounded burst.
type relayPacer struct {
	lock CriticalSection
	next time.Time
}

func (pacer *relayPacer) wait(ctx context.Context, size int) error {
	until, _ := pacer.lock.With2(func() (time.Time, error) {
		now := time.Now()
		if pacer.next.Before(now) {
			pacer.next = now
		}
		pacer.next = pacer.next.Add(time.Duration(size) * time.Second / combatBytesPerSecond)
		return pacer.next, nil
	})
	timer := time.NewTimer(time.Until(until))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type relaySlot struct {
	token      string
	connection *websocket.Conn
}
type combatRelay struct {
	lock         CriticalSection
	slots        [2]relaySlot
	ctx          context.Context
	cancel       context.CancelFunc
	paired, done chan struct{}
	once         sync.Once
	onDone       func()
	pacer        relayPacer
}

func newCombatRelay(onDone func()) *combatRelay {
	ctx, cancel := context.WithCancel(context.Background())
	relay := &combatRelay{ctx: ctx, cancel: cancel, paired: make(chan struct{}), done: make(chan struct{}), onDone: onDone}
	for i := range relay.slots {
		relay.slots[i].token = rand.Text()
	}
	go func() {
		timer := time.NewTimer(30 * time.Second)
		defer timer.Stop()
		select {
		case <-relay.paired:
		case <-relay.done:
		case <-timer.C:
			relay.stop()
		}
	}()
	return relay
}

func (relay *combatRelay) stop() {
	relay.once.Do(func() {
		relay.cancel()
		close(relay.done)
		// Never perform network close operations while the manager is locked.
		go func() {
			relay.lock.With(func() error {
				for _, slot := range relay.slots {
					if slot.connection != nil {
						slot.connection.CloseNow()
					}
				}
				return nil
			})
			if relay.onDone != nil {
				relay.onDone()
			}
		}()
	})
}

func (relay *combatRelay) attach(token string, connection *websocket.Conn) error {
	return relay.lock.With(func() error {
		if relay.ctx.Err() != nil {
			return errors.New("Combat relay has closed")
		}
		index := -1
		for i := range relay.slots {
			if relay.slots[i].token == token {
				index = i
			}
		}
		if index == -1 {
			return errors.New("Invalid combat token")
		}
		if relay.slots[index].connection != nil {
			return errors.New("Combat connection already attached")
		}
		relay.slots[index].connection = connection
		if relay.slots[0].connection != nil && relay.slots[1].connection != nil {
			close(relay.paired)
			go relay.run(relay.slots[0].connection, relay.slots[1].connection)
		}
		return nil
	})
}

func (relay *combatRelay) run(first, second *websocket.Conn) {
	defer relay.stop()
	for _, connection := range []*websocket.Conn{first, second} {
		ctx, cancel := context.WithTimeout(relay.ctx, 5*time.Second)
		err := wsjson.Write(ctx, connection, protocol.Message{Type: "relay-ready"})
		cancel()
		if err != nil {
			return
		}
	}
	done := make(chan error, 2)
	go func() { done <- relay.forward(first, second) }()
	go func() { done <- relay.forward(second, first) }()
	<-done
}

func (relay *combatRelay) forward(from, to *websocket.Conn) error {
	buffer := make([]byte, relayChunkSize)
	for {
		kind, reader, err := from.Reader(relay.ctx)
		if err != nil {
			return err
		}
		if kind != websocket.MessageBinary {
			return errors.New("Combat data must be binary")
		}
		for {
			n, err := reader.Read(buffer)
			if n > 0 {
				if err := relay.pacer.wait(relay.ctx, n); err != nil {
					return err
				}
				ctx, cancel := context.WithTimeout(relay.ctx, 10*time.Second)
				writeErr := to.Write(ctx, websocket.MessageBinary, buffer[:n])
				cancel()
				if writeErr != nil {
					return writeErr
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
		}
	}
}

// startCombat runs under manager.lock. Capabilities are private to each lobby
// connection; game IDs/client IDs alone cannot authorize a combat socket.
func (manager *Manager) startCombat(game *Game) {
	if game.relay != nil {
		return
	}
	var relay *combatRelay
	relay = newCombatRelay(func() {
		manager.lock.With(func() error {
			if manager.Games[game.id] == game && game.relay == relay {
				manager.leaveGame(game.players[game.hostID], game.hostID)
			}
			return nil
		})
	})
	game.relay = relay
	var defenderID uint64
	for id := range game.players {
		if id != game.attackerID {
			defenderID = id
		}
	}
	log.Printf("Game %d starting in room %q: attacker client=%d (%q), defender client=%d (%q), creator client=%d", game.id, game.room, game.attackerID, game.players[game.attackerID].name, defenderID, game.players[defenderID].name, game.hostID)
	index := 1
	for id, user := range game.players {
		slot := index
		if id == game.hostID {
			slot = 0
		}
		user.send(protocol.Message{Type: "combat-start", GameID: game.id, RelayToken: relay.slots[slot].token, Attacker: id == game.attackerID})
	}
}

func (manager *Manager) acceptCombat(writer http.ResponseWriter, request *http.Request) {
	connection, err := websocket.Accept(writer, request, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	defer connection.CloseNow()
	connection.SetReadLimit(1024*1024 + 4)
	ctx, cancel := context.WithTimeout(request.Context(), 10*time.Second)
	var hello protocol.Message
	err = wsjson.Read(ctx, connection, &hello)
	cancel()
	if err != nil {
		return
	}
	relay, _ := manager.lock.With2(func() (*combatRelay, error) {
		if game := manager.Games[hello.GameID]; game != nil && game.started {
			return game.relay, nil
		}
		return nil, nil
	})
	if relay == nil {
		// Slow invalid game-ID probes without holding the manager lock.
		timer := time.NewTimer(time.Second)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-request.Context().Done():
			return
		}
		connection.Close(websocket.StatusPolicyViolation, "Combat game is not available")
		return
	}
	if err := relay.attach(hello.RelayToken, connection); err != nil {
		connection.Close(websocket.StatusPolicyViolation, fmt.Sprint(err))
		return
	}
	<-relay.done
}
