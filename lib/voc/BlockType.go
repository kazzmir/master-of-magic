package voc

type blockType byte

const (
	terminator = blockType(0x00)
	soundData  = blockType(0x01)
    text = blockType(0x5)
)
