package combat

import (
    "net"
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
