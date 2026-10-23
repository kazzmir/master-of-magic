package net

import (
    gonet "net"
    "encoding/binary"
    "fmt"
    "io"
)

// send and receive a complete packet, not a stream of bytes
// each packet is prefixed with a 4 byte uint32 length
type PacketConn interface {
    Read() ([]byte, error)
    Send([]byte) error
}

type PacketNetConn struct {
    gonet.Conn
    MaxPacketSize uint32
}

// convert a net.Conn to a PacketConn
// a single packet cannot exceed maxPacketSize, otherwise an error is returned
func NewPacketNetConn(peer gonet.Conn, maxPacketSize uint32) *PacketNetConn {
    return &PacketNetConn{
        Conn: peer,
        MaxPacketSize: maxPacketSize,
    }
}

func (packet *PacketNetConn) Read() ([]byte, error) {
    // read 4 byte length prefix
    var length uint32
    err := binary.Read(packet.Conn, binary.BigEndian, &length)
    if err != nil {
        return nil, err
    }

    if length > packet.MaxPacketSize {
        return nil, fmt.Errorf("Packet size %d exceeds maximum allowed size %d", length, packet.MaxPacketSize)
    }

    data := make([]byte, length)
    _, err = io.ReadFull(packet.Conn, data)
    if err != nil {
        return nil, err
    }

    return data, nil
}

func (packet *PacketNetConn) Send(data []byte) error {
    length := uint32(len(data))

    if length > packet.MaxPacketSize {
        return fmt.Errorf("Packet size %d exceeds maximum allowed size %d", length, packet.MaxPacketSize)
    }

    err := binary.Write(packet.Conn, binary.BigEndian, length)
    if err != nil {
        return err
    }

    _, err = packet.Conn.Write(data)
    if err != nil {
        return err
    }

    return nil
}
