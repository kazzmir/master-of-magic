package combat

import (
    "log"
    "time"
    "net"
    "io"
    "errors"
    "bytes"
    "context"
    "encoding/binary"
    "encoding/json/v2"

    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
)

const (
    RemoteTeleportType = "teleport"
    RemoteMoveType     = "move"
    RemoteRangeAttackType = "range_attack"
    RemoteProjectileFinishedType = "projectile_finished"
    RemoteDamageType = "damage"
    RemoteDamageIndicatorType = "damage_indicator"
    RemoteMeleeAttackType = "melee_attack"
    RemoteDoneTurnType = "done_turn"
    RemoteFinishMeleeAttackType = "melee_finished"
    RemoteHealType = "heal"
    RemoteKillUnitType = "kill_unit"
)

type Remote struct {
    // true if the remote side is the server side that will handle combat logic (random numbers)
    IsServer bool
    // true if remote side is attacker. false if defender
    Attacker bool
    Peer     net.Conn

    Events chan RemoteEvent
}

func MakeRemote(isServer bool, isAttacker bool, peer net.Conn) *Remote {
    return &Remote{
        IsServer: isServer,
        Attacker: isAttacker,
        Peer:     peer,
        Events:   make(chan RemoteEvent, 100),
    }
}

type RemoteEvent interface {
    GetType() string
}

var ErrorEmptyType = errors.New("event type is empty")

// event should be some event type from below
func (remote *Remote) SendEvent(event RemoteEvent) error {
    if event.GetType() == "" {
        log.Printf("Error: event type is empty")
        return ErrorEmptyType
    }

    var out bytes.Buffer
    err := json.MarshalWrite(&out, event)
    if err != nil {
        return err
    }

    // write 4-byte uint32 of json data
    size := uint32(out.Len())
    err = binary.Write(remote.Peer, binary.BigEndian, size)
    if err != nil {
        return err
    }

    // write json data
    _, err = remote.Peer.Write(out.Bytes())

    log.Printf("Sent json data: %s", out.String())
    return err
}

// caller should run this in a goroutine
func (remote *Remote) RunReceiveLoop(quit context.Context) {

    for quit.Err() == nil {
        done := make(chan struct{})
        go func() {
            select {
                case <-quit.Done():
                    // force read to error
                    remote.Peer.SetReadDeadline(time.Now())
                case <-done:
            }
        }()

        event, err := remote.ReceiveEvent()
        close(done)
        if err != nil {
            log.Printf("Error receiving event: %v", err)
            return
        }

        if event == nil {
            log.Printf("Received nil event")
            return
        }

        remote.Events <- event
    }
}

func (remote *Remote) ReceiveEvent() (RemoteEvent, error) {
    // read 4-byte uint32

    var size uint32
    err := binary.Read(remote.Peer, binary.BigEndian, &size)
    if err != nil {
        return nil, err
    }

    // read json data
    data := make([]byte, size)
    _, err = io.ReadFull(remote.Peer, data)
    if err != nil {
        return nil, err
    }

    log.Printf("Received json data: %s", string(data))

    // unmarshal into a map to get the type
    var m map[string]any
    err = json.Unmarshal(data, &m)
    if err != nil {
        return nil, err
    }

    // check type
    eventType, ok := m["type"].(string)
    if !ok {
        log.Printf("Error: event type is not a string")
        return nil, err
    }

    switch eventType {
        case RemoteTeleportType: return convert[*RemoteTeleportEvent](data)
        case RemoteRangeAttackType: return convert[*RemoteRangeAttackEvent](data)
        case RemoteMoveType: return convert[*RemoteMoveEvent](data)
        case RemoteProjectileFinishedType: return convert[*RemoteProjectileFinishedEvent](data)
        case RemoteDamageType: return convert[*RemoteDamageEvent](data)
        case RemoteDamageIndicatorType: return convert[*RemoteDamageIndicatorEvent](data)
        case RemoteMeleeAttackType: return convert[*RemoteMeleeAttackEvent](data)
        case RemoteDoneTurnType: return convert[*RemoteDoneTurnEvent](data)
        case RemoteFinishMeleeAttackType: return convert[*RemoteFinishMeleeAttackEvent](data)
        case RemoteHealType: return convert[*RemoteHealEvent](data)
        case RemoteKillUnitType: return convert[*RemoteKillUnitEvent](data)
        default:
            log.Printf("Error: unknown event type: %s", eventType)
            return nil, err
    }
}

func convert[T RemoteEvent](data []byte) (RemoteEvent, error) {
    var event T
    err := json.Unmarshal(data, &event)
    if err != nil {
        return nil, err
    }
    return event, nil
}

// sent when a unit teleports to a location
type RemoteTeleportEvent struct {
    Id uint64 `json:"id"`
    Type string `json:"type"`
    X int    `json:"x"`
    Y int    `json:"y"`
}

func (remote *RemoteTeleportEvent) GetType() string {
    return remote.Type
}

type RemoteProjectileFinishedEvent struct {
    Type string `json:"type"`
}

func (remote *RemoteProjectileFinishedEvent) GetType() string {
    return remote.Type
}

type RemoteRangeAttackEvent struct {
    AttackerId uint64 `json:"attacker_id"`
    DefenderId uint64 `json:"defender_id"`
    Type string `json:"type"`
}

func (remote *RemoteRangeAttackEvent) GetType() string {
    return remote.Type
}

type RemoteMoveEvent struct {
    Id uint64 `json:"id"`
    Type string `json:"type"`
    Path pathfinding.Path `json:"path"`
}

func (remote *RemoteMoveEvent) GetType() string {
    return remote.Type
}

type RemoteDamageEvent struct {
    Id uint64 `json:"id"`
    Type string `json:"type"`
    DamageKind DamageType `json:"damage_kind"`
    Damage int `json:"damage"`
}

func (remote *RemoteDamageEvent) GetType() string {
    return remote.Type
}

type RemoteDamageIndicatorEvent struct {
    Id uint64 `json:"id"`
    Type string `json:"type"`
    Damage int `json:"damage"`
}

func (remote *RemoteDamageIndicatorEvent) GetType() string {
    return remote.Type
}

type RemoteMeleeAttackEvent struct {
    Type string `json:"type"`
    AttackerId uint64 `json:"attacker_id"`
    DefenderId uint64 `json:"defender_id"`
}

func (remote *RemoteMeleeAttackEvent) GetType() string {
    return remote.Type
}

type RemoteDoneTurnEvent struct {
    Type string `json:"type"`
    Id uint64 `json:"id"`
}

func (remote *RemoteDoneTurnEvent) GetType() string {
    return remote.Type
}

type RemoteFinishMeleeAttackEvent struct {
    Type string `json:"type"`
}

func (remote *RemoteFinishMeleeAttackEvent) GetType() string {
    return remote.Type
}

type RemoteHealEvent struct {
    Type string `json:"type"`
    Id uint64 `json:"id"`
    Heal int `json:"heal"`
}

func (remote *RemoteHealEvent) GetType() string {
    return remote.Type
}

type RemoteKillUnitEvent struct {
    Type string `json:"type"`
    Id uint64 `json:"id"`
}

func (remote *RemoteKillUnitEvent) GetType() string {
    return remote.Type
}
