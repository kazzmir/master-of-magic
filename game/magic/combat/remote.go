package combat

import (
    "net"
)

type Remote struct {
    IsServer bool
    Peer     net.Conn
}

func MakeRemote(isServer bool, peer net.Conn) *Remote {
    return &Remote{
        IsServer: isServer,
        Peer:     peer,
    }
}
