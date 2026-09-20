package combat

import (
    "log"
    "net"
    "bytes"
    "encoding/binary"
    "encoding/json/v2"
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
}

func MakeRemote(isServer bool, isAttacker bool, peer net.Conn) *Remote {
    return &Remote{
        IsServer: isServer,
        Attacker: isAttacker,
        Peer:     peer,
    }
}

// event should be some event type from below
func (remote *Remote) SendEvent(event any) error {
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

// sent when a unit teleports to a location
type RemoteTeleportEvent struct {
    Id uint64 `json:"id"`
    Type string `json:"type"`
    X int    `json:"x"`
    Y int    `json:"y"`
}

type RemoteMoveEvent struct {
    Id uint64 `json:"id"`
    Type string `json:"type"`
    X int    `json:"x"`
    Y int    `json:"y"`
}
