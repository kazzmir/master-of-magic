package lbx

import (
    "fmt"
    "log"
    "io"
    "bufio"
    "bytes"
    "encoding/binary"
    "image"
    "image/color"
)

func readUint16(reader io.Reader) (uint16, error) {
    var value uint16
    err := binary.Read(reader, binary.LittleEndian, &value)
    return value, err
}

func readUint16Big(reader io.Reader) (uint16, error) {
    var value uint16
    err := binary.Read(reader, binary.BigEndian, &value)
    return value, err
}

func readUint32(reader io.Reader) (uint32, error) {
    var value uint32
    err := binary.Read(reader, binary.LittleEndian, &value)
    return value, err
}

func readInt32(reader io.Reader) (int32, error) {
    v, err := readUint32(reader)
    return int32(v), err
}

type LbxFile struct {
    Signature uint32
    Version uint16
    Data [][]byte
}

type PaletteInfo struct {
    Offset uint16
    Count uint16
    FirstColorIndex uint16
}

var defaultPalette = color.Palette {
    // FIXME: set color 0 to transparent, double check that this is correct
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0x0},
    color.RGBA{R: 0x8,  G: 0x4,  B: 0x4, A: 0xff},
    color.RGBA{R: 0x24, G: 0x1c, B: 0x18, A: 0xff},
    color.RGBA{R: 0x38, G: 0x30, B: 0x2c, A: 0xff},
    color.RGBA{R: 0x48, G: 0x40, B: 0x3c, A: 0xff},
    color.RGBA{R: 0x58, G: 0x50, B: 0x4c, A: 0xff},
    color.RGBA{R: 0x68, G: 0x60, B: 0x5c, A: 0xff},
    color.RGBA{R: 0x7c, G: 0x74, B: 0x70, A: 0xff},
    color.RGBA{R: 0x8c, G: 0x84, B: 0x80, A: 0xff},
    color.RGBA{R: 0x9c, G: 0x94, B: 0x90, A: 0xff},
    color.RGBA{R: 0xac, G: 0xa4, B: 0xa0, A: 0xff},
    color.RGBA{R: 0xc0, G: 0xb8, B: 0xb4, A: 0xff},
    color.RGBA{R: 0xd0, G: 0xc8, B: 0xc4, A: 0xff},
    color.RGBA{R: 0xe0, G: 0xd8, B: 0xd4, A: 0xff},
    color.RGBA{R: 0xf0, G: 0xe8, B: 0xe4, A: 0xff},
    color.RGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
    color.RGBA{R: 0x38, G: 0x20, B: 0x1c, A: 0xff},
    color.RGBA{R: 0x40, G: 0x2c, B: 0x24, A: 0xff},
    color.RGBA{R: 0x48, G: 0x34, B: 0x2c, A: 0xff},
    color.RGBA{R: 0x50, G: 0x3c, B: 0x30, A: 0xff},
    color.RGBA{R: 0x58, G: 0x40, B: 0x34, A: 0xff},
    color.RGBA{R: 0x5c, G: 0x44, B: 0x38, A: 0xff},
    color.RGBA{R: 0x60, G: 0x48, B: 0x3c, A: 0xff},
    color.RGBA{R: 0x64, G: 0x4c, B: 0x3c, A: 0xff},
    color.RGBA{R: 0x68, G: 0x50, B: 0x40, A: 0xff},
    color.RGBA{R: 0x70, G: 0x54, B: 0x44, A: 0xff},
    color.RGBA{R: 0x78, G: 0x5c, B: 0x4c, A: 0xff},
    color.RGBA{R: 0x80, G: 0x64, B: 0x50, A: 0xff},
    color.RGBA{R: 0x8c, G: 0x70, B: 0x58, A: 0xff},
    color.RGBA{R: 0x94, G: 0x74, B: 0x5c, A: 0xff},
    color.RGBA{R: 0x9c, G: 0x7c, B: 0x64, A: 0xff},
    color.RGBA{R: 0xa4, G: 0x84, B: 0x68, A: 0xff},
    color.RGBA{R: 0xec, G: 0xc0, B: 0xd4, A: 0xff},
    color.RGBA{R: 0xd4, G: 0x98, B: 0xb4, A: 0xff},
    color.RGBA{R: 0xbc, G: 0x74, B: 0x94, A: 0xff},
    color.RGBA{R: 0xa4, G: 0x54, B: 0x7c, A: 0xff},
    color.RGBA{R: 0x8c, G: 0x38, B: 0x60, A: 0xff},
    color.RGBA{R: 0x74, G: 0x24, B: 0x4c, A: 0xff},
    color.RGBA{R: 0x5c, G: 0x10, B: 0x34, A: 0xff},
    color.RGBA{R: 0x44, G: 0x4,  B: 0x20, A: 0xff},
    color.RGBA{R: 0xec, G: 0xc0, B: 0xc0, A: 0xff},
    color.RGBA{R: 0xd4, G: 0x94, B: 0x94, A: 0xff},
    color.RGBA{R: 0xbc, G: 0x74, B: 0x74, A: 0xff},
    color.RGBA{R: 0xa4, G: 0x54, B: 0x54, A: 0xff},
    color.RGBA{R: 0x8c, G: 0x38, B: 0x38, A: 0xff},
    color.RGBA{R: 0x74, G: 0x24, B: 0x24, A: 0xff},
    color.RGBA{R: 0x5c, G: 0x10, B: 0x10, A: 0xff},
    color.RGBA{R: 0x44, G: 0x4,  B: 0x4, A: 0xff},
    color.RGBA{R: 0xec, G: 0xd4, B: 0xc0, A: 0xff},
    color.RGBA{R: 0xd4, G: 0xb4, B: 0x98, A: 0xff},
    color.RGBA{R: 0xbc, G: 0x98, B: 0x74, A: 0xff},
    color.RGBA{R: 0xa4, G: 0x7c, B: 0x54, A: 0xff},
    color.RGBA{R: 0x8c, G: 0x60, B: 0x38, A: 0xff},
    color.RGBA{R: 0x74, G: 0x4c, B: 0x24, A: 0xff},
    color.RGBA{R: 0x5c, G: 0x34, B: 0x10, A: 0xff},
    color.RGBA{R: 0x44, G: 0x24, B: 0x4, A: 0xff},
    color.RGBA{R: 0xec, G: 0xec, B: 0xc0, A: 0xff},
    color.RGBA{R: 0xd4, G: 0xd4, B: 0x94, A: 0xff},
    color.RGBA{R: 0xbc, G: 0xbc, B: 0x74, A: 0xff},
    color.RGBA{R: 0xa4, G: 0xa4, B: 0x54, A: 0xff},
    color.RGBA{R: 0x8c, G: 0x8c, B: 0x38, A: 0xff},
    color.RGBA{R: 0x74, G: 0x74, B: 0x24, A: 0xff},
    color.RGBA{R: 0x5c, G: 0x5c, B: 0x10, A: 0xff},
    color.RGBA{R: 0x44, G: 0x44, B: 0x4, A: 0xff},
    color.RGBA{R: 0xd4, G: 0xec, B: 0xbc, A: 0xff},
    color.RGBA{R: 0xb8, G: 0xd4, B: 0x98, A: 0xff},
    color.RGBA{R: 0x98, G: 0xbc, B: 0x74, A: 0xff},
    color.RGBA{R: 0x7c, G: 0xa4, B: 0x54, A: 0xff},
    color.RGBA{R: 0x60, G: 0x8c, B: 0x38, A: 0xff},
    color.RGBA{R: 0x4c, G: 0x74, B: 0x24, A: 0xff},
    color.RGBA{R: 0x38, G: 0x5c, B: 0x10, A: 0xff},
    color.RGBA{R: 0x24, G: 0x44, B: 0x4, A: 0xff},
    color.RGBA{R: 0xc0, G: 0xec, B: 0xc0, A: 0xff},
    color.RGBA{R: 0x98, G: 0xd4, B: 0x98, A: 0xff},
    color.RGBA{R: 0x74, G: 0xbc, B: 0x74, A: 0xff},
    color.RGBA{R: 0x54, G: 0xa4, B: 0x54, A: 0xff},
    color.RGBA{R: 0x38, G: 0x8c, B: 0x38, A: 0xff},
    color.RGBA{R: 0x24, G: 0x74, B: 0x24, A: 0xff},
    color.RGBA{R: 0x10, G: 0x5c, B: 0x10, A: 0xff},
    color.RGBA{R: 0x4,  G: 0x44, B: 0x4, A: 0xff},
    color.RGBA{R: 0xc0, G: 0xec, B: 0xd8, A: 0xff},
    color.RGBA{R: 0x98, G: 0xd4, B: 0xb8, A: 0xff},
    color.RGBA{R: 0x74, G: 0xbc, B: 0x98, A: 0xff},
    color.RGBA{R: 0x54, G: 0xa4, B: 0x7c, A: 0xff},
    color.RGBA{R: 0x38, G: 0x8c, B: 0x60, A: 0xff},
    color.RGBA{R: 0x24, G: 0x74, B: 0x4c, A: 0xff},
    color.RGBA{R: 0x10, G: 0x5c, B: 0x38, A: 0xff},
    color.RGBA{R: 0x4,  G: 0x44, B: 0x24, A: 0xff},
    color.RGBA{R: 0xf4, G: 0xc0, B: 0xa0, A: 0xff},
    color.RGBA{R: 0xe0, G: 0xa0, B: 0x84, A: 0xff},
    color.RGBA{R: 0xcc, G: 0x84, B: 0x6c, A: 0xff},
    color.RGBA{R: 0xc8, G: 0x8c, B: 0x68, A: 0xff},
    color.RGBA{R: 0xa8, G: 0x78, B: 0x54, A: 0xff},
    color.RGBA{R: 0x98, G: 0x68, B: 0x4c, A: 0xff},
    color.RGBA{R: 0x8c, G: 0x60, B: 0x44, A: 0xff},
    color.RGBA{R: 0x7c, G: 0x50, B: 0x3c, A: 0xff},
    color.RGBA{R: 0xc0, G: 0xd8, B: 0xec, A: 0xff},
    color.RGBA{R: 0x94, G: 0xb4, B: 0xd4, A: 0xff},
    color.RGBA{R: 0x70, G: 0x98, B: 0xbc, A: 0xff},
    color.RGBA{R: 0x54, G: 0x7c, B: 0xa4, A: 0xff},
    color.RGBA{R: 0x38, G: 0x64, B: 0x8c, A: 0xff},
    color.RGBA{R: 0x24, G: 0x4c, B: 0x74, A: 0xff},
    color.RGBA{R: 0x10, G: 0x38, B: 0x5c, A: 0xff},
    color.RGBA{R: 0x4,  G: 0x24, B: 0x44, A: 0xff},
    color.RGBA{R: 0xc0, G: 0xc0, B: 0xec, A: 0xff},
    color.RGBA{R: 0x98, G: 0x98, B: 0xd4, A: 0xff},
    color.RGBA{R: 0x74, G: 0x74, B: 0xbc, A: 0xff},
    color.RGBA{R: 0x54, G: 0x54, B: 0xa4, A: 0xff},
    color.RGBA{R: 0x3c, G: 0x38, B: 0x8c, A: 0xff},
    color.RGBA{R: 0x24, G: 0x24, B: 0x74, A: 0xff},
    color.RGBA{R: 0x10, G: 0x10, B: 0x5c, A: 0xff},
    color.RGBA{R: 0x4,  G: 0x4,  B: 0x44, A: 0xff},
    color.RGBA{R: 0xd8, G: 0xc0, B: 0xec, A: 0xff},
    color.RGBA{R: 0xb8, G: 0x98, B: 0xd4, A: 0xff},
    color.RGBA{R: 0x98, G: 0x74, B: 0xbc, A: 0xff},
    color.RGBA{R: 0x7c, G: 0x54, B: 0xa4, A: 0xff},
    color.RGBA{R: 0x60, G: 0x38, B: 0x8c, A: 0xff},
    color.RGBA{R: 0x4c, G: 0x24, B: 0x74, A: 0xff},
    color.RGBA{R: 0x38, G: 0x10, B: 0x5c, A: 0xff},
    color.RGBA{R: 0x24, G: 0x4,  B: 0x44, A: 0xff},
    color.RGBA{R: 0xec, G: 0xc0, B: 0xec, A: 0xff},
    color.RGBA{R: 0xd4, G: 0x98, B: 0xd4, A: 0xff},
    color.RGBA{R: 0xbc, G: 0x74, B: 0xbc, A: 0xff},
    color.RGBA{R: 0xa4, G: 0x54, B: 0xa4, A: 0xff},
    color.RGBA{R: 0x8c, G: 0x38, B: 0x8c, A: 0xff},
    color.RGBA{R: 0x74, G: 0x24, B: 0x74, A: 0xff},
    color.RGBA{R: 0x5c, G: 0x10, B: 0x5c, A: 0xff},
    color.RGBA{R: 0x44, G: 0x4,  B: 0x44, A: 0xff},
    color.RGBA{R: 0xd8, G: 0xd0, B: 0xd0, A: 0xff},
    color.RGBA{R: 0xc0, G: 0xb0, B: 0xb0, A: 0xff},
    color.RGBA{R: 0xa4, G: 0x90, B: 0x90, A: 0xff},
    color.RGBA{R: 0x8c, G: 0x74, B: 0x74, A: 0xff},
    color.RGBA{R: 0x78, G: 0x5c, B: 0x5c, A: 0xff},
    color.RGBA{R: 0x68, G: 0x4c, B: 0x4c, A: 0xff},
    color.RGBA{R: 0x5c, G: 0x3c, B: 0x3c, A: 0xff},
    color.RGBA{R: 0x48, G: 0x2c, B: 0x2c, A: 0xff},
    color.RGBA{R: 0xd0, G: 0xd8, B: 0xd0, A: 0xff},
    color.RGBA{R: 0xb0, G: 0xc0, B: 0xb0, A: 0xff},
    color.RGBA{R: 0x90, G: 0xa4, B: 0x90, A: 0xff},
    color.RGBA{R: 0x74, G: 0x8c, B: 0x74, A: 0xff},
    color.RGBA{R: 0x5c, G: 0x78, B: 0x5c, A: 0xff},
    color.RGBA{R: 0x4c, G: 0x68, B: 0x4c, A: 0xff},
    color.RGBA{R: 0x3c, G: 0x5c, B: 0x3c, A: 0xff},
    color.RGBA{R: 0x2c, G: 0x48, B: 0x2c, A: 0xff},
    color.RGBA{R: 0xc8, G: 0xc8, B: 0xd4, A: 0xff},
    color.RGBA{R: 0xb0, G: 0xb0, B: 0xc0, A: 0xff},
    color.RGBA{R: 0x90, G: 0x90, B: 0xa4, A: 0xff},
    color.RGBA{R: 0x74, G: 0x74, B: 0x8c, A: 0xff},
    color.RGBA{R: 0x5c, G: 0x5c, B: 0x78, A: 0xff},
    color.RGBA{R: 0x4c, G: 0x4c, B: 0x68, A: 0xff},
    color.RGBA{R: 0x3c, G: 0x3c, B: 0x5c, A: 0xff},
    color.RGBA{R: 0x2c, G: 0x2c, B: 0x48, A: 0xff},
    color.RGBA{R: 0xd8, G: 0xdc, B: 0xec, A: 0xff},
    color.RGBA{R: 0xc8, G: 0xcc, B: 0xdc, A: 0xff},
    color.RGBA{R: 0xb8, G: 0xc0, B: 0xd4, A: 0xff},
    color.RGBA{R: 0xa8, G: 0xb8, B: 0xcc, A: 0xff},
    color.RGBA{R: 0x9c, G: 0xb0, B: 0xcc, A: 0xff},
    color.RGBA{R: 0x94, G: 0xac, B: 0xcc, A: 0xff},
    color.RGBA{R: 0x88, G: 0xa4, B: 0xcc, A: 0xff},
    color.RGBA{R: 0x88, G: 0x94, B: 0xdc, A: 0xff},
    color.RGBA{R: 0xfc, G: 0xf0, B: 0x90, A: 0xff},
    color.RGBA{R: 0xfc, G: 0xe4, B: 0x60, A: 0xff},
    color.RGBA{R: 0xfc, G: 0xc8, B: 0x24, A: 0xff},
    color.RGBA{R: 0xfc, G: 0xac, B: 0xc, A: 0xff},
    color.RGBA{R: 0xfc, G: 0x78, B: 0x10, A: 0xff},
    color.RGBA{R: 0xd0, G: 0x1c, B: 0x0, A: 0xff},
    color.RGBA{R: 0x98, G: 0x0,  B: 0x0, A: 0xff},
    color.RGBA{R: 0x58, G: 0x0,  B: 0x0, A: 0xff},
    color.RGBA{R: 0x90, G: 0xf0, B: 0xfc, A: 0xff},
    color.RGBA{R: 0x60, G: 0xe4, B: 0xfc, A: 0xff},
    color.RGBA{R: 0x24, G: 0xc8, B: 0xfc, A: 0xff},
    color.RGBA{R: 0xc,  G: 0xac, B: 0xfc, A: 0xff},
    color.RGBA{R: 0x10, G: 0x78, B: 0xfc, A: 0xff},
    color.RGBA{R: 0x0,  G: 0x1c, B: 0xd0, A: 0xff},
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x98, A: 0xff},
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x58, A: 0xff},
    color.RGBA{R: 0xfc, G: 0xc8, B: 0x64, A: 0xff},
    color.RGBA{R: 0xfc, G: 0xb4, B: 0x2c, A: 0xff},
    color.RGBA{R: 0xec, G: 0xa4, B: 0x24, A: 0xff},
    color.RGBA{R: 0xdc, G: 0x94, B: 0x1c, A: 0xff},
    color.RGBA{R: 0xcc, G: 0x88, B: 0x18, A: 0xff},
    color.RGBA{R: 0xbc, G: 0x7c, B: 0x14, A: 0xff},
    color.RGBA{R: 0xa4, G: 0x6c, B: 0x1c, A: 0xff},
    color.RGBA{R: 0x8c, G: 0x60, B: 0x24, A: 0xff},
    color.RGBA{R: 0x78, G: 0x54, B: 0x24, A: 0xff},
    color.RGBA{R: 0x60, G: 0x44, B: 0x24, A: 0xff},
    color.RGBA{R: 0x48, G: 0x38, B: 0x24, A: 0xff},
    color.RGBA{R: 0x34, G: 0x28, B: 0x1c, A: 0xff},
    color.RGBA{R: 0x90, G: 0x68, B: 0x34, A: 0xff},
    color.RGBA{R: 0x90, G: 0x64, B: 0x2c, A: 0xff},
    color.RGBA{R: 0x94, G: 0x6c, B: 0x34, A: 0xff},
    color.RGBA{R: 0x94, G: 0x70, B: 0x40, A: 0xff},
    color.RGBA{R: 0x8c, G: 0x5c, B: 0x24, A: 0xff},
    color.RGBA{R: 0x90, G: 0x64, B: 0x2c, A: 0xff},
    color.RGBA{R: 0x90, G: 0x68, B: 0x30, A: 0xff},
    color.RGBA{R: 0x98, G: 0x78, B: 0x4c, A: 0xff},
    color.RGBA{R: 0x60, G: 0x3c, B: 0x2c, A: 0xff},
    color.RGBA{R: 0x54, G: 0xa4, B: 0xa4, A: 0xff},
    color.RGBA{R: 0xc0, G: 0x0,  B: 0x0, A: 0xff},
    color.RGBA{R: 0xfc, G: 0x88, B: 0xe0, A: 0xff},
    color.RGBA{R: 0xfc, G: 0x58, B: 0x84, A: 0xff},
    color.RGBA{R: 0xf4, G: 0x0,  B: 0xc, A: 0xff},
    color.RGBA{R: 0xd4, G: 0x0,  B: 0x0, A: 0xff},
    color.RGBA{R: 0xac, G: 0x0,  B: 0x0, A: 0xff},
    color.RGBA{R: 0xe8, G: 0xa8, B: 0xfc, A: 0xff},
    color.RGBA{R: 0xe0, G: 0x7c, B: 0xfc, A: 0xff},
    color.RGBA{R: 0xd0, G: 0x3c, B: 0xfc, A: 0xff},
    color.RGBA{R: 0xc4, G: 0x0,  B: 0xfc, A: 0xff},
    color.RGBA{R: 0x90, G: 0x0,  B: 0xbc, A: 0xff},
    color.RGBA{R: 0xfc, G: 0xf4, B: 0x7c, A: 0xff},
    color.RGBA{R: 0xfc, G: 0xe4, B: 0x0, A: 0xff},
    color.RGBA{R: 0xe4, G: 0xd0, B: 0x0, A: 0xff},
    color.RGBA{R: 0xa4, G: 0x98, B: 0x0, A: 0xff},
    color.RGBA{R: 0x64, G: 0x58, B: 0x0, A: 0xff},
    color.RGBA{R: 0xac, G: 0xfc, B: 0xa8, A: 0xff},
    color.RGBA{R: 0x74, G: 0xe4, B: 0x70, A: 0xff},
    color.RGBA{R: 0x0,  G: 0xbc, B: 0x0, A: 0xff},
    color.RGBA{R: 0x0,  G: 0xa4, B: 0x0, A: 0xff},
    color.RGBA{R: 0x0,  G: 0x7c, B: 0x0, A: 0xff},
    color.RGBA{R: 0xac, G: 0xa8, B: 0xfc, A: 0xff},
    color.RGBA{R: 0x80, G: 0x7c, B: 0xfc, A: 0xff},
    color.RGBA{R: 0x0,  G: 0x0,  B: 0xfc, A: 0xff},
    color.RGBA{R: 0x0,  G: 0x0,  B: 0xbc, A: 0xff},
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x7c, A: 0xff},
    color.RGBA{R: 0x30, G: 0x30, B: 0x50, A: 0xff},
    color.RGBA{R: 0x28, G: 0x28, B: 0x48, A: 0xff},
    color.RGBA{R: 0x24, G: 0x24, B: 0x40, A: 0xff},
    color.RGBA{R: 0x20, G: 0x1c, B: 0x38, A: 0xff},
    color.RGBA{R: 0x1c, G: 0x18, B: 0x34, A: 0xff},
    color.RGBA{R: 0x18, G: 0x14, B: 0x2c, A: 0xff},
    color.RGBA{R: 0x14, G: 0x10, B: 0x24, A: 0xff},
    color.RGBA{R: 0x10, G: 0xc,  B: 0x20, A: 0xff},
    color.RGBA{R: 0xa0, G: 0xa0, B: 0xb4, A: 0xff},
    color.RGBA{R: 0x88, G: 0x88, B: 0xa4, A: 0xff},
    color.RGBA{R: 0x74, G: 0x74, B: 0x90, A: 0xff},
    color.RGBA{R: 0x60, G: 0x60, B: 0x80, A: 0xff},
    color.RGBA{R: 0x50, G: 0x4c, B: 0x70, A: 0xff},
    color.RGBA{R: 0x40, G: 0x3c, B: 0x60, A: 0xff},
    color.RGBA{R: 0x30, G: 0x2c, B: 0x50, A: 0xff},
    color.RGBA{R: 0x24, G: 0x20, B: 0x40, A: 0xff},
    color.RGBA{R: 0x18, G: 0x14, B: 0x30, A: 0xff},
    color.RGBA{R: 0x10, G: 0xc,  B: 0x20, A: 0xff},
    color.RGBA{R: 0x14, G: 0xc,  B: 0x8, A: 0xff},
    color.RGBA{R: 0x18, G: 0x10, B: 0xc, A: 0xff},
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff},
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff},
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff},
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff},
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff},
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff},
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff},
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff},
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff},
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff},
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff},
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff},
}

func GetDefaultPalette() color.Palette {
    return clonePalette(defaultPalette)
}

func clonePalette(p color.Palette) color.Palette {
    newPalette := make(color.Palette, len(p))
    copy(newPalette, p)
    return newPalette
}

func readPaletteInfo(reader io.ReadSeeker, index int) (PaletteInfo, error) {
    reader.Seek(int64(index), io.SeekStart)

    offset, err := readUint16(reader)
    if err != nil {
        return PaletteInfo{}, err
    }

    firstColor, err := readUint16(reader)
    if err != nil {
        return PaletteInfo{}, err
    }

    count, err := readUint16(reader)
    if err != nil {
        return PaletteInfo{}, err
    }

    unknown1, err := readUint16(reader)
    if err != nil {
        return PaletteInfo{}, err
    }
    _ = unknown1

    info := PaletteInfo{
        Offset: offset,
        FirstColorIndex: firstColor,
        Count: count,
    }

    return info, nil
}

func readPalette(reader io.ReadSeeker, index int, firstColor int, count int) (color.Palette, error) {
    reader.Seek(int64(index), io.SeekStart)

    palette := clonePalette(defaultPalette)

    byteReader := bufio.NewReader(reader)

    for i := 0; i < count; i++ {
        r, err := byteReader.ReadByte()
        if err != nil {
            return nil, err
        }
        g, err := byteReader.ReadByte()
        if err != nil {
            return nil, err
        }
        b, err := byteReader.ReadByte()
        if err != nil {
            return nil, err
        }

        if i + firstColor >= len(palette) {
            return nil, fmt.Errorf("invalid color index %v, palette only has %v colors", i + firstColor, len(palette))
        }

        // log.Printf("Palette[%v] = %v %v %v\n", i + firstColor, r, g, b)

        // color values are multiplied by 4, but i'm not entirely sure why (other than the pascal source does the same thing)
        palette[i + firstColor] = color.RGBA{R: uint8(r << 2), G: uint8(g << 2), B: uint8(b << 2), A: 0xff}
    }

    return palette, nil
    // return palette, nil
}

const debug = false

/* read an RLE encoded image using the given palette, storing pixels into 'img'
 */
func readImage(reader io.Reader, img *image.Paletted, palette color.Palette, startRleValue int) error {
    byteReader, ok := reader.(io.ByteReader)
    if !ok {
        byteReader = bufio.NewReader(reader)
    }

    reset, err := byteReader.ReadByte()
    if err != nil {
        return err
    }

    if reset == 1 {
        // TODO: reset image to blank, which is magic pink 0xff00ff in the original code
    }

    x := 0

    for {
        v, err := byteReader.ReadByte()
        if err != nil {
            return nil
        }

        if debug {
            log.Printf("Read byte 0x%x\n", v)
        }

        if v == 0xff {
            continue
        }

        rle := startRleValue

        if v == 0 {
            rle = startRleValue
        } else if v == 0x80 {
            rle = 0xe0
        } else {
            return fmt.Errorf("unexpected rle value 0x%x", v)
        }

        next_, err := byteReader.ReadByte()
        if err != nil {
            return err
        }
        next := int(next_) - 2

        if next == 0 {
            return fmt.Errorf("next bitmap location cannot be 0")
        }

        data, err := byteReader.ReadByte()
        if err != nil {
            return err
        }

        y, err := byteReader.ReadByte()
        if err != nil {
            return err
        }

        if debug {
            log.Printf("RLE: 0x%x, Next: %v, Data: %v, Y: %v\n", rle, next, data, y)
        }

        total := 0

        for total < next {
            for data > 0 {
                v2, err := byteReader.ReadByte()
                if err != nil {
                    return err
                }

                total += 1

                if int(v2) >= rle {
                    length := int(v2) - int(rle) + 1
                    index, err := byteReader.ReadByte()
                    if err != nil {
                        return err
                    }
                    total += 1

                    if length > img.Bounds().Dy() {
                        return fmt.Errorf("rle length %v is greater than image height %v", length, img.Bounds().Dy())
                    }

                    if debug {
                        log.Printf("rle length=%v index=%v x=%v y=%v\n", length, index, x, y)
                    }

                    for i := 0; i < length; i++ {
                        img.SetColorIndex(x, int(y), uint8(index))
                        y += 1
                    }

                    data -= 2

                } else {
                    if debug {
                        log.Printf("normal pixel x=%v y=%v v=%v\n", x, y, v2)
                    }
                    img.SetColorIndex(x, int(y), uint8(v2))
                    y += 1
                    data -= 1
                }
            }

            if total < next {
                if debug {
                    log.Printf("Have to read two more bytes total=%v next=%v\n", total, next)
                }
                newData, err := byteReader.ReadByte()
                if err != nil {
                    return err
                }
                newY, err := byteReader.ReadByte()
                if err != nil {
                    return err
                }

                total += 2

                y += newY
                data = newData
            }
        }

        x += 1
    }

    return nil
}

/* use this to read a font entry, usually from fonts.lbx */
func (lbx *LbxFile) ReadFonts(entry int) ([]*Font, error) {
    if entry < 0 || entry >= len(lbx.Data) {
        return nil, fmt.Errorf("invalid lbx index %v, must be between 0 and %v", entry, len(lbx.Data) - 1)
    }

    reader := bytes.NewReader(lbx.Data[entry])
    return readFonts(reader)
}

/* terrain.lbx has a special format that is different from regular graphics */
func (lbx *LbxFile) ReadTerrainImages(entry int) ([]image.Image, error) {
    /* skip first 192 bytes */

    if entry < 0 || entry >= len(lbx.Data) {
        return nil, fmt.Errorf("invalid lbx index %v, must be between 0 and %v", entry, len(lbx.Data) - 1)
    }

    reader := bytes.NewReader(lbx.Data[entry])
    if reader.Size() < 192 {
        return nil, fmt.Errorf("invalid format for terrain, size is %v but must be at least 192", reader.Size())
    }

    reader.Seek(192, io.SeekStart)

    var images []image.Image
    for reader.Len() > 0 {
        width, err := readUint16(reader)
        if err != nil {
            break
        }

        height, err := readUint16(reader)
        if err != nil {
            return nil, err
        }

        for i := 0; i < 6; i++ {
            readUint16(reader)
        }

        img := image.NewPaletted(image.Rect(0, 0, int(width), int(height)), defaultPalette)
        for x := 0; x < int(width); x++ {
            for y := 0; y < int(height); y++ {
                index, err := reader.ReadByte()
                if err != nil {
                    return nil, err
                }

                img.SetColorIndex(x, y, index)
            }
        }

        for i := 0; i < 4; i++ {
            readUint16(reader)
        }

        images = append(images, img)
    }

    return images, nil
}

func (lbx *LbxFile) TotalEntries() int {
    return len(lbx.Data)
}

func (lbx *LbxFile) ReadSpells(entry int) (Spells, error) {
    if entry < 0 || entry >= len(lbx.Data) {
        return Spells{}, fmt.Errorf("invalid lbx index %v, must be between 0 and %v", entry, len(lbx.Data) - 1)
    }

    reader := bytes.NewReader(lbx.Data[entry])

    numEntries, err := readUint16(reader)
    if err != nil {
        return Spells{}, err
    }

    entrySize, err := readUint16(reader)
    if err != nil {
        return Spells{}, err
    }

    var spells Spells

    type MagicData struct {
        Magic SpellMagic
        Rarity SpellRarity
    }

    spellMagicIterator := (func() chan MagicData {
        out := make(chan MagicData)

        go func() {
            defer close(out)

            out <- MagicData{Magic: SpellMagicNone}
            order := []SpellMagic{SpellMagicNature, SpellMagicSorcery, SpellMagicChaos, SpellMagicLife, SpellMagicDeath}
            rarities := []SpellRarity{SpellRarityCommon, SpellRarityUncommon, SpellRarityRare, SpellRarityVeryRare}

            for _, magic := range order {
                // 10 types of common, uncommon, rare, very rare for each book of magic
                for _, rarity := range rarities {
                    for i := 0; i < 10; i++ {
                        out <- MagicData{Magic: magic, Rarity: rarity}
                    }
                }
            }

            // for arcane the spells are
            // common: magic spirit, dispel magic, spell of return, summoning circle
            // uncommon: detect magic, recall hero, disenchant area, enchant item, summon hero
            // rare: awareness, disjunction, create artifact, summon champion
            // very rare: spell of mastery

            for i := 0; i < 4; i++ {
                out <- MagicData{Magic: SpellMagicArcane, Rarity: SpellRarityCommon}
            }

            for i := 0; i < 5; i++ {
                out <- MagicData{Magic: SpellMagicArcane, Rarity: SpellRarityUncommon}
            }

            for i := 0; i < 4; i++ {
                out <- MagicData{Magic: SpellMagicArcane, Rarity: SpellRarityRare}
            }

            out <- MagicData{Magic: SpellMagicArcane, Rarity: SpellRarityVeryRare}
        }()

        return out
    })()

    for i := 0; i < int(numEntries); i++ {
        data := make([]byte, entrySize)
        n, err := reader.Read(data)
        if err != nil {
            return Spells{}, fmt.Errorf("Error reading help index %v: %v", i, err)
        }

        buffer := bytes.NewBuffer(data[0:n])

        nameData := buffer.Next(18)
        // fmt.Printf("Spell %v\n", i)

        name, err := bytes.NewBuffer(nameData).ReadString(0)
        if err != nil {
            name = string(nameData)
        } else {
            name = name[0:len(name)-1]
        }
        // fmt.Printf("  Name: %v\n", string(name))

        aiGroup, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }
        // fmt.Printf("  AI Group: %v\n", aiGroup)

        aiValue, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }
        // fmt.Printf("  AI Value: %v\n", aiValue)

        spellType, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }

        // fmt.Printf("  Spell Type: %v\n", spellType)

        section, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }

        // fmt.Printf("  Section: %v\n", section)

        realm, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }

        // fmt.Printf("  Magic Realm: %v\n", realm)

        eligibility, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }

        // fmt.Printf("  Caster Eligibility: %v\n", eligibility)

        buffer.Next(1) // ignore extra unused byte from 2-byte alignment

        castCost, err := readUint16Big(buffer)
        if err != nil {
            return Spells{}, err
        }

        // fmt.Printf("  Casting Cost: %v\n", castCost)

        researchCost, err := readUint16Big(buffer)
        if err != nil {
            return Spells{}, err
        }

        // fmt.Printf("  Research Cost: %v\n", researchCost)

        sound, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }

        // fmt.Printf("  Sound effect: %v\n", sound)

        // skip extra byte due to 2-byte alignment
        buffer.ReadByte()

        summoned, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }

        // fmt.Printf("  Summoned: %v\n", summoned)

        flag1, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }

        flag2, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }

        // FIXME: should this be a uint16?
        flag3, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }

        // fmt.Printf("  Flag1=%v Flag2=%v Flag3=%v\n", flag1, flag2, flag3)

        magicData := <-spellMagicIterator

        spells.AddSpell(Spell{
            Name: name,
            AiGroup: int(aiGroup),
            AiValue: int(aiValue),
            SpellType: int(spellType),
            Section: int(section),
            Realm: int(realm),
            Eligibility: int(eligibility),
            CastCost: int(castCost),
            ResearchCost: int(researchCost),
            Sound: int(sound),
            Summoned: int(summoned),
            Flag1: int(flag1),
            Flag2: int(flag2),
            Flag3: int(flag3),

            Magic: magicData.Magic,
            Rarity: magicData.Rarity,
        })
    }

    return spells, nil
}

func (lbx *LbxFile) ReadHelp(entry int) (Help, error) {
    if entry < 0 || entry >= len(lbx.Data) {
        return Help{}, fmt.Errorf("invalid lbx index %v, must be between 0 and %v", entry, len(lbx.Data) - 1)
    }

    reader := bytes.NewReader(lbx.Data[entry])

    numEntries, err := readUint16(reader)
    if err != nil {
        return Help{}, err
    }

    entrySize, err := readUint16(reader)
    if err != nil {
        return Help{}, err
    }

    if entrySize == 0 {
        return Help{}, fmt.Errorf("entry size was 0 in help")
    }

    if numEntries * entrySize > uint16(reader.Len()) {
        return Help{}, fmt.Errorf("too many entries in the help file entries=%v size=%v len=%v", numEntries, entrySize, reader.Len())
    }

    var help []HelpEntry

    // fmt.Printf("num entries: %v\n", numEntries)

    for i := 0; i < int(numEntries); i++ {
        data := make([]byte, entrySize)
        n, err := reader.Read(data)
        if err != nil {
            return Help{}, fmt.Errorf("Error reading help index %v: %v", i, err)
        }

        buffer := bytes.NewBuffer(data[0:n])

        headlineData := buffer.Next(30)

        b2 := bytes.NewBuffer(headlineData)
        headline, err := b2.ReadString(0)
        if err != nil {
            headline = string(headlineData)
        } else {
            headline = headline[0:len(headline)-1]
        }

        // fmt.Printf("  headline: %v\n", string(headline))

        // fmt.Printf("  at position 0x%x\n", n - buffer.Len())

        pictureLbxData := buffer.Next(14)
        b2 = bytes.NewBuffer(pictureLbxData)
        pictureLbx, err := b2.ReadString(0)
        // fmt.Printf("  lbx: %v\n", string(pictureLbx))

        pictureLbx = pictureLbx[0:len(pictureLbx)-1]

        // fmt.Printf("  at position 0x%x\n", n - buffer.Len())

        pictureIndex, err := readUint16(buffer)
        if err != nil {
            return Help{}, fmt.Errorf("Error reading help index %v: %v", i, err)
        }

        // fmt.Printf("  lbx index: %v\n", pictureIndex)

        appendHelpText, err := readUint16(buffer)
        if err != nil {
            return Help{}, err
        }

        // fmt.Printf("  appended help text: 0x%x\n", appendHelpText)

        info, err := buffer.ReadString(0)
        if err != nil {
            return Help{}, err
        }

        info = info[0:len(info)-1]

        // fmt.Printf("  text: '%v'\n", info)

        help = append(help, HelpEntry{
            Headline: headline,
            Lbx: pictureLbx,
            LbxIndex: int(pictureIndex),
            AppendHelpIndex: int(appendHelpText),
            Text: info,
        })
    }

    out := Help{Entries: help}
    out.updateMap()
    return out, nil
}

func (lbx *LbxFile) ReadImages(entry int) ([]image.Image, error) {

    if entry < 0 || entry >= len(lbx.Data) {
        return nil, fmt.Errorf("invalid lbx index %v, must be between 0 and %v", entry, len(lbx.Data) - 1)
    }

    reader := bytes.NewReader(lbx.Data[entry])

    width, err := readUint16(reader)
    if err != nil {
        return nil, err
    }

    height, err := readUint16(reader)
    if err != nil {
        return nil, err
    }

    unknown1, err := readUint16(reader)
    if err != nil {
        return nil, err
    }
    _ = unknown1

    bitmapCount, err := readUint16(reader)
    if err != nil {
        return nil, err
    }

    unknown2, err := readUint16(reader)
    if err != nil {
        return nil, err
    }
    _ = unknown2

    unknown3, err := readUint16(reader)
    if err != nil {
        return nil, err
    }
    _ = unknown3

    unknown4, err := readUint16(reader)
    if err != nil {
        return nil, err
    }
    _ = unknown4

    paletteOffset, err := readUint16(reader)
    if err != nil {
        return nil, err
    }

    unknown5, err := readUint16(reader)
    if err != nil {
        return nil, err
    }
    _ = unknown5

    var offsets []uint32
    for i := 0; i < int(bitmapCount) + 1; i++ {
        offset, err := readUint32(reader)
        if err != nil {
            return nil, err
        }

        offsets = append(offsets, offset)
    }

    if debug {
        fmt.Printf("Width: %v\n", width)
        fmt.Printf("Height: %v\n", height)
        fmt.Printf("Bitmap count: %v\n", bitmapCount)
        fmt.Printf("Palette offset: %v\n", paletteOffset)
    }

    var paletteInfo PaletteInfo
    palette := defaultPalette

    if paletteOffset > 0 {
        paletteInfo, err = readPaletteInfo(reader, int(paletteOffset))
        if err != nil {
            return nil, err
        }

        palette, err = readPalette(reader, int(paletteInfo.Offset), int(paletteInfo.FirstColorIndex), int(paletteInfo.Count))
        if err != nil {
            return nil, err
        }

        if debug {
            fmt.Printf("Read palette with %v colors first color %v count %v\n", len(palette), paletteInfo.FirstColorIndex, paletteInfo.Count)
        }
    }

    // palette[241] = color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff}

    /* if the palette is empty then just use the default palette */
    if paletteInfo.Count == 0 {
        paletteInfo.FirstColorIndex = 0
        paletteInfo.Count = 255
    }

    if debug {
        fmt.Printf("Palette info: %+v\n", paletteInfo)
    }

    var images []image.Image

    for i := 0; i < int(bitmapCount); i++ {
        end := offsets[i+1]
        if debug {
            fmt.Printf("Read image %v at offset %v size %v\n", i, offsets[i], end - offsets[i])
        }

        reader.Seek(int64(offsets[i]), io.SeekStart)

        imageReader := io.LimitReader(reader, int64(end - offsets[i]))

        img := image.NewPaletted(image.Rect(0, 0, int(width), int(height)), palette)

        err = readImage(imageReader, img, palette, int(paletteInfo.FirstColorIndex + paletteInfo.Count))
        if err != nil {
            return nil, err
        }

        if debug {
            x := 39
            y := 20
            index := img.ColorIndexAt(x, y)
            fmt.Printf("(%v,%v)=%v\n", x, y, index)
            /*
            for x := 0; x < img.Bounds().Dx(); x++ {
                for y := 0; y < img.Bounds().Dy(); y++ {
                    index := img.ColorIndexAt(x, y)
                    fmt.Printf("(%v,%v)=%v\n", x, y, index)
                }
            }
            */
        }

        images = append(images, img)
    }

    return images, nil
}

func (lbxFile *LbxFile) RawData(entry int) ([]byte, error) {
    if entry < 0 || entry >= len(lbxFile.Data) {
        return nil, fmt.Errorf("invalid lbx index %v, must be between 0 and %v", entry, len(lbxFile.Data) - 1)
    }

    return lbxFile.Data[entry], nil
}

const LbxSignature = 0x0000fead

func ReadLbx(reader io.ReadSeeker) (LbxFile, error) {
    numFiles, err := readUint16(reader)
    if err != nil {
        return LbxFile{}, err
    }

    signature, err := readUint32(reader)
    if err != nil {
        return LbxFile{}, err
    }

    if signature != LbxSignature {
        return LbxFile{}, fmt.Errorf("Invalid lbx signature, was 0x%x but expected 0x%x\n", signature, LbxSignature)
    }

    version, err := readUint16(reader)
    if err != nil {
        return LbxFile{}, err
    }

    // fmt.Printf("Version: %v\n", version)

    var offsets []uint32

    for i := 0; i < int(numFiles); i++ {
        offset, err := readUint32(reader)
        if err != nil {
            return LbxFile{}, err
        }

        // fmt.Printf("Offset %v: 0x%x\n", i, offset)

        offsets = append(offsets, offset)
    }

    reader.Seek(0, io.SeekEnd)
    lastByte, _ := reader.Seek(0, io.SeekCurrent)

    var lbx LbxFile

    lbx.Signature = signature
    lbx.Version = version

    for i, offset := range offsets {
        reader.Seek(int64(offset), io.SeekStart)
        end := uint32(0)
        if i < len(offsets) - 1 {
            end = offsets[i+1]
        } else {
            end = uint32(lastByte)
        }

        limitedReader := io.LimitReader(reader, int64(end - offset))
        var buffer bytes.Buffer
        io.Copy(&buffer, limitedReader)

        // fmt.Printf("File %v: %v bytes\n", i, buffer.Len())
        lbx.Data = append(lbx.Data, buffer.Bytes())
    }

    return lbx, nil
}
