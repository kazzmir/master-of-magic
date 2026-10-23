package net

// send and receive a complete packet, not a stream of bytes
// each packet is prefixed with a 4 byte uint32 length
type PacketConn interface {
    Read() ([]byte, error)
    Send([]byte) error
}

