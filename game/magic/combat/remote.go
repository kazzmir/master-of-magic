package combat

import (
    "log"
    "time"
    "net"
    "io"
    "bytes"
    "context"
    "encoding/binary"
    "encoding/json/v2"

    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
)

const (
    RemoteTeleportType = "teleport"
    RemoteMoveType     = "move"
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

// event should be some event type from below
func (remote *Remote) SendEvent(event RemoteEvent) error {
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
        return nil, err
    }

    switch eventType {
        case RemoteTeleportType:
            var event RemoteTeleportEvent
            err = json.Unmarshal(data, &event)
            if err != nil {
                return nil, err
            }
            return &event, nil
        case RemoteMoveType:
            var event RemoteMoveEvent
            err = json.Unmarshal(data, &event)
            if err != nil {
                return nil, err
            }
            return &event, nil
        default:
            return nil, err
    }
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

type RemoteMoveEvent struct {
    Id uint64 `json:"id"`
    Type string `json:"type"`
    Path pathfinding.Path `json:"path"`
}

func (remote *RemoteMoveEvent) GetType() string {
    return remote.Type
}
