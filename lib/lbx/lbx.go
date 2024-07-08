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
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff},
    color.RGBA{R: 0x8,  G: 0x4,  B: 0x4, A: 0xff},
    color.RGBA{R: 0x24, G: 0x1c, B: 0x18, A: 0xff},
    color.RGBA{R: 0x38, G: 0x30, B: 0x2c, A: 0xff},
    color.RGBA{R: 0x48, G: 0x40, B: 0x3c, A: 0xff},
    color.RGBA{R: 0x58, G: 0x50, B: 0x4c, A: 0xff},
    color.RGBA{R: 0x68, G: 0x60, B: 0x5c, A: 0xff},
    color.RGBA{R: 0x7c, G: 0x74, B: 70, A: 0xff},
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

        palette[i + firstColor] = color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 0xff}
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

        fmt.Printf("Read palette with %v colors\n", len(palette))
    }

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
        images = append(images, img)
    }

    return images, nil
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
