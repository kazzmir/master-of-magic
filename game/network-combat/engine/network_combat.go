package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kazzmir/master-of-magic/game/magic/audio"
	"github.com/kazzmir/master-of-magic/game/magic/mouse"
	"github.com/kazzmir/master-of-magic/game/network-combat/engine/crypto"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
	netlib "github.com/kazzmir/master-of-magic/lib/net"
)

type preparedCombat struct {
	engine *Engine
	packet *netlib.WebsocketPacketConn
	err    error
}

type networkCombatRun struct {
	gameID   uint64
	cancel   context.CancelFunc
	prepared chan preparedCombat
	engine   *Engine
	packet   *netlib.WebsocketPacketConn
}

func connectCombat(ctx context.Context, address string, start protocol.Message) (netlib.PacketConn, *netlib.WebsocketPacketConn, error) {
	endpoint, err := multiplayerURL(address)
	if err != nil {
		return nil, nil, err
	}
	dialCtx, dialCancel := context.WithTimeout(ctx, 10*time.Second)
	connection, _, err := websocket.Dial(dialCtx, endpoint+"/combat", nil)
	dialCancel()
	if err != nil {
		return nil, nil, err
	}
	ok := false
	defer func() {
		if !ok {
			connection.CloseNow()
		}
	}()
	helloCtx, helloCancel := context.WithTimeout(ctx, 30*time.Second)
	defer helloCancel()
	if err := wsjson.Write(helloCtx, connection, protocol.Message{GameID: start.GameID, RelayToken: start.RelayToken}); err != nil {
		return nil, nil, err
	}
	var ready protocol.Message
	if err := wsjson.Read(helloCtx, connection, &ready); err != nil {
		return nil, nil, err
	}
	if ready.Type != "relay-ready" {
		return nil, nil, fmt.Errorf("combat relay did not become ready")
	}
	packet := netlib.NewWebsocketPacketConn(ctx, connection, MAX_PACKET_SIZE)
	packet.SetReadDeadline(time.Now().Add(20 * time.Second))
	encrypted, err := crypto.NewEncryptedConn(packet)
	packet.SetReadDeadline(time.Time{})
	if err != nil {
		packet.Close()
		return nil, nil, fmt.Errorf("combat key exchange: %w", err)
	}
	ok = true
	return encrypted, packet, nil
}

func closeCombatEngine(engine *Engine) {
	if engine == nil {
		return
	}
	engine.Combat.Cancel()
	engine.Coroutine.Stop()
}

func (game *lobbyGame) startNetworkCombat(message protocol.Message) {
	if message.GameID == 0 || message.RelayToken == "" || game.matches.active.ID != message.GameID || !game.matches.active.Started {
		return
	}
	if game.networkCombat != nil && game.networkCombat.gameID == message.GameID {
		return
	}
	game.stopNetworkCombat("", false)
	if audio.Context == nil {
		audio.Initialize()
	}
	if mouse.Mouse == nil {
		mouse.Initialize()
	}
	ctx, cancel := context.WithCancel(context.Background())
	run := &networkCombatRun{gameID: message.GameID, cancel: cancel, prepared: make(chan preparedCombat)}
	game.networkCombat = run
	game.matches.status.Label = "Connecting to combat and exchanging keys…"
	address := game.server.GetText()
	go func() {
		encrypted, packet, err := connectCombat(ctx, address, message)
		var engine *Engine
		if err == nil {
			engine, err = NewEngine(message.Attacker, encrypted, networkCombatScenario)
		}
		if err != nil && packet != nil {
			packet.Close()
		}
		select {
		case run.prepared <- preparedCombat{engine: engine, packet: packet, err: err}:
		case <-ctx.Done():
			closeCombatEngine(engine)
			if packet != nil {
				packet.Close()
			}
		}
	}()
}

func (game *lobbyGame) stopNetworkCombat(status string, notify bool) {
	run := game.networkCombat
	if run == nil {
		return
	}
	game.networkCombat = nil
	run.cancel()
	if run.packet != nil {
		run.packet.Close()
	}
	closeCombatEngine(run.engine)
	ebiten.SetCursorMode(ebiten.CursorModeVisible)
	ebiten.SetWindowTitle("Master of Magic — Multiplayer")
	if status != "" {
		game.status.Label = status
	}
	// Once attached, closing the relay socket ends the game server-side.
	// Avoid racing that cleanup with a redundant leave-game command.
	if notify && run.packet == nil {
		game.send(protocol.Message{Type: "leave-game", GameID: run.gameID})
	}
}

// updateNetworkCombat runs on Ebiten's update thread. The existing RunGame is
// reused: Ebiten does not support starting a second game loop inside the lobby.
func (game *lobbyGame) updateNetworkCombat() bool {
	run := game.networkCombat
	if run == nil {
		return false
	}
	select {
	case ready := <-run.prepared:
		if ready.err != nil {
			run.packet = ready.packet
			game.stopNetworkCombat("Unable to start combat: "+ready.err.Error(), true)
			return false
		}
		run.engine, run.packet = ready.engine, ready.packet
		ebiten.SetCursorMode(ebiten.CursorModeHidden)
		ebiten.SetWindowTitle("Master of Magic — Network Combat")
	default:
	}
	if run.engine == nil {
		return false
	}
	select {
	case <-run.packet.Done():
		game.stopNetworkCombat("Combat connection closed", false)
		return false
	default:
	}
	if err := run.engine.Update(); err != nil {
		game.stopNetworkCombat("Combat ended", true)
		return false
	}
	return true
}
