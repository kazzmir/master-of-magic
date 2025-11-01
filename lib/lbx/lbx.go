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
    "strings"
    "path/filepath"
)

func ReadByte(reader io.Reader) (byte, error) {
    var value byte
    err := binary.Read(reader, binary.BigEndian, &value)
    return value, err
}

func ReadUint16(reader io.Reader) (uint16, error) {
    var value uint16
    err := binary.Read(reader, binary.LittleEndian, &value)
    return value, err
}

func ReadN[T any](reader io.Reader) (T, error) {
    var value T
    err := binary.Read(reader, binary.LittleEndian, &value)
    return value, err
}

func ReadArrayN[T any](reader io.Reader, count int) ([]T, error) {
    var err error
    data := make([]T, count)
    for i := range count {
        data[i], err = ReadN[T](reader)
        if err != nil {
            return nil, err
        }
    }
    return data, nil
}

func ReadUint16Big(reader io.Reader) (uint16, error) {
    var value uint16
    err := binary.Read(reader, binary.BigEndian, &value)
    return value, err
}

func ReadUint32(reader io.Reader) (uint32, error) {
    var value uint32
    err := binary.Read(reader, binary.LittleEndian, &value)
    return value, err
}

func readInt32(reader io.Reader) (int32, error) {
    v, err := ReadUint32(reader)
    return int32(v), err
}

func WriteUint16(writer io.Writer, value uint16) error {
    return binary.Write(writer, binary.LittleEndian, value)
}

func WriteUint32(writer io.Writer, value uint32) error {
    return binary.Write(writer, binary.LittleEndian, value)
}

type LbxFile struct {
    Signature uint32
    Version uint16
    Data [][]byte
    // most lbx files have an extra strings section
    Strings []string
}

type PaletteInfo struct {
    Offset uint16
    Count uint16
    FirstColorIndex uint16
}

func premultiply(c color.RGBA) color.RGBA {
    a := float64(c.A) / 255.0
    return color.RGBA{
        R: uint8(float64(c.R) * a),
        G: uint8(float64(c.G) * a),
        B: uint8(float64(c.B) * a),
        A: c.A,
    }
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

    // the transparent color 232
    // color.RGBA{R: 0xa0, G: 0xa0, B: 0xb4, A: 0xff},
    premultiply(color.RGBA{R: 0x10, G: 0x10, B: 0x10, A: 160}),

    premultiply(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x80}),
    premultiply(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x80}),
    premultiply(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x80}),
    premultiply(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x80}),

    /*
    color.RGBA{R: 0x88, G: 0x88, B: 0xa4, A: 0xff},
    color.RGBA{R: 0x74, G: 0x74, B: 0x90, A: 0xff},
    color.RGBA{R: 0x60, G: 0x60, B: 0x80, A: 0xff},
    color.RGBA{R: 0x50, G: 0x4c, B: 0x70, A: 0xff},
    */
    color.RGBA{R: 0x40, G: 0x3c, B: 0x60, A: 0xff},
    color.RGBA{R: 0x30, G: 0x2c, B: 0x50, A: 0xff},
    color.RGBA{R: 0x24, G: 0x20, B: 0x40, A: 0xff},
    color.RGBA{R: 0x18, G: 0x14, B: 0x30, A: 0xff},
    color.RGBA{R: 0x10, G: 0xc,  B: 0x20, A: 0xff},
    color.RGBA{R: 0x14, G: 0xc,  B: 0x8, A: 0xff},
    color.RGBA{R: 0x18, G: 0x10, B: 0xc, A: 0xff},

    color.RGBA{R: 0x30, G: 0x30, B: 0x50, A: 0xff}, // 244
    color.RGBA{R: 0x28, G: 0x28, B: 0x48, A: 0xff}, // 245
    color.RGBA{R: 0x24, G: 0x24, B: 0x40, A: 0xff}, // 246
    color.RGBA{R: 0x20, G: 0x1c, B: 0x38, A: 0xff}, // 247

    /*
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff}, // 244
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff}, // 245
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff}, // 246
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff}, // 247
    */

    color.RGBA{R: 0x1c, G: 0x18, B: 0x34, A: 0xff}, // 248
    color.RGBA{R: 0x18, G: 0x14, B: 0x2c, A: 0xff}, // 249
    color.RGBA{R: 0x14, G: 0x10, B: 0x24, A: 0xff}, // 250
    color.RGBA{R: 0x10, G: 0xc,  B: 0x20, A: 0xff}, // 251
    color.RGBA{R: 0x40, G: 0x3c, B: 0x60, A: 0xff}, // 252

    /*
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff}, // 248
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff}, // 249
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff}, // 250
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff}, // 251
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff}, // 252
    */

    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff}, // 253
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff}, // 254
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff}, // 255
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

    offset, err := ReadUint16(reader)
    if err != nil {
        return PaletteInfo{}, err
    }

    firstColor, err := ReadUint16(reader)
    if err != nil {
        return PaletteInfo{}, err
    }

    count, err := ReadUint16(reader)
    if err != nil {
        return PaletteInfo{}, err
    }

    unknown1, err := ReadUint16(reader)
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

var debug = false

func copyImage(img *image.Paletted) *image.Paletted {
    out := image.NewPaletted(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()), img.Palette)

    for y := 0; y < img.Bounds().Dy(); y++ {
        for x := 0; x < img.Bounds().Dx(); x++ {
            out.SetColorIndex(x, y, img.ColorIndexAt(x, y))
        }
    }

    return out
}

// given a chunk of bytes that reader can read one byte at a time in sequential order, read a sprite out
// if the first byte is 0, then this is a delta frame and the lastImage is used as a base to modify
func readImage(reader io.Reader, lastImage *image.Paletted, width int, height int, palette color.Palette, startRleValue int) (*image.Paletted, error) {
    byteReader, ok := reader.(io.ByteReader)
    if !ok {
        byteReader = bufio.NewReader(reader)
    }

    reset, err := byteReader.ReadByte()
    if err != nil {
        return nil, err
    }

    deltaFrame := reset == 0

    var img *image.Paletted

    if deltaFrame {
        if lastImage == nil {
            return nil, fmt.Errorf("cannot have a delta frame without a full frame before it")
        }

        img = copyImage(lastImage)
    } else {
        img = image.NewPaletted(image.Rect(0, 0, width, height), palette)
    }

    // for each column of pixel data, read an operation and perform the operation.
    // an operation could be 'set the next pixel to value X' or an rle operation
    // which means to set the next N pixels to value X
    for x := 0; x < width; x++ {
        operation, err := byteReader.ReadByte()
        if err != nil {
            // return img, nil
            return nil, fmt.Errorf("Missing operation byte")
        }

        if operation == 0xff {
            continue
        }

        isRLE := false
        if operation == 0x80 {
            isRLE = true
        } else if operation == 0 {
            isRLE = false
        } else {
            return nil, fmt.Errorf("Invalid pixel operation: 0x%x", operation)
        }

        // Read some number of packets (size is packet_count)
        // each packet consists of a skip count followed by data bytes
        // a data byte can either be just a pixel value, or an RLE length
        // if the operation is RLE and the value > 0xdf

        b, err := byteReader.ReadByte()
        if err != nil {
            return nil, fmt.Errorf("Missing packet count byte")
        }
        packet_count := int(b)

        y := 0
        for packet_count > 0 {
            b, err := byteReader.ReadByte()
            if err != nil {
                return nil, fmt.Errorf("Missing data count byte")
            }
            packet_count -= 1
            data_count := int(b)

            skip_count, err := byteReader.ReadByte()
            if err != nil {
                return nil, fmt.Errorf("Missing skip count byte")
            }
            packet_count -= 1

            y += int(skip_count)
            packet_count -= data_count

            // read data_count bytes
            for i := 0; i < data_count; i++ {
                value, err := byteReader.ReadByte()
                if err != nil {
                    return nil, fmt.Errorf("Missing palette index byte")
                }

                if y < height {
                    if isRLE && value > 0xdf {
                        length := value - 0xdf
                        pixel, err := byteReader.ReadByte()
                        if err != nil {
                            return nil, fmt.Errorf("Missing pixel value")
                        }
                        i += 1

                        for j := byte(0); j < length; j++ {
                            if y < height {
                                img.SetColorIndex(x, y, pixel)
                                y += 1
                            }
                        }
                    } else {
                        img.SetColorIndex(x, y, value)
                        y += 1
                    }
                }
            }
        }
    }

    return img, nil
}

/* read an RLE encoded image using the given palette, and return the new image.
 * if this image is a delta frame (when the first byte is 0) then this image starts
 * with a copy of lastImage, and modifies it with the new data.
 */
/*
func readImage2(reader io.Reader, lastImage *image.Paletted, width int, height int, palette color.Palette, startRleValue int) (*image.Paletted, error) {
    byteReader, ok := reader.(io.ByteReader)
    if !ok {
        byteReader = bufio.NewReader(reader)
    }

    reset, err := byteReader.ReadByte()
    if err != nil {
        return nil, err
    }

    deltaFrame := reset == 0

    var img *image.Paletted

    if deltaFrame {

        if lastImage == nil {
            return nil, fmt.Errorf("cannot have a delta frame without a full frame before it")
        }

        img = copyImage(lastImage)
    } else {
        img = image.NewPaletted(image.Rect(0, 0, width, height), palette)
    }

    x := 0

    // for each column of pixel data, read an operation and perform the operation.
    // an operation could be 'set the next pixel to value X' or an rle operation
    // which means to set the next N pixels to value X
    for {
        v, err := byteReader.ReadByte()
        if err != nil {
            return img, nil
        }

        if debug {
            log.Printf("Read byte 0x%x\n", v)
        }

        // done with this column, just go to the next one
        if v == 0xff {
            x += 1
            continue
        }

        rle := startRleValue

        if v == 0 {
            rle = startRleValue
        } else if v == 0x80 {
            rle = 0xe0
        } else {
            return nil, fmt.Errorf("unexpected rle value 0x%x", v)
        }

        next_, err := byteReader.ReadByte()
        if err != nil {
            return nil, err
        }
        next := int(next_) - 2

        if next == 0 {
            return nil, fmt.Errorf("next bitmap location cannot be 0")
        }

        data, err := byteReader.ReadByte()
        if err != nil {
            return nil, err
        }

        y, err := byteReader.ReadByte()
        if err != nil {
            return nil, err
        }

        if debug {
            log.Printf("RLE: 0x%x, Next: %v, Data: %v, Y: %v\n", rle, next, data, y)
        }

        total := 0

        for total < next {
            for data > 0 {
                v2, err := byteReader.ReadByte()
                if err != nil {
                    return nil, err
                }

                total += 1

                if int(v2) >= rle {
                    length := int(v2) - int(rle) + 1
                    index, err := byteReader.ReadByte()
                    if err != nil {
                        return nil, err
                    }
                    total += 1

                    if length > img.Bounds().Dy() {
                        return nil, fmt.Errorf("rle length %v is greater than image height %v", length, img.Bounds().Dy())
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
                    return nil, err
                }
                newY, err := byteReader.ReadByte()
                if err != nil {
                    return nil, err
                }

                total += 2

                y += newY
                data = newData
            }
        }

        x += 1
    }
}
*/

func (lbx *LbxFile) GetReader(entry int) (*bytes.Reader, error) {
    if entry < 0 || entry >= len(lbx.Data) {
        return nil, fmt.Errorf("invalid lbx index %v, must be between 0 and %v", entry, len(lbx.Data) - 1)
    }

    return bytes.NewReader(lbx.Data[entry]), nil
}

/* use this to read a font entry, usually from fonts.lbx */
/*
func (lbx *LbxFile) ReadFonts(entry int) ([]*LbxFont, error) {
    if entry < 0 || entry >= len(lbx.Data) {
        return nil, fmt.Errorf("invalid lbx index %v, must be between 0 and %v", entry, len(lbx.Data) - 1)
    }

    reader := bytes.NewReader(lbx.Data[entry])
    return readFonts(reader)
}
*/

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
        width, err := ReadUint16(reader)
        if err != nil {
            break
        }

        height, err := ReadUint16(reader)
        if err != nil {
            return nil, err
        }

        for i := 0; i < 6; i++ {
            ReadUint16(reader)
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
            ReadUint16(reader)
        }

        images = append(images, img)
    }

    return images, nil
}

func (lbx *LbxFile) TotalEntries() int {
    return len(lbx.Data)
}

/* return just the palette for a given entry, or nil if there is no custom palette */
func (lbx *LbxFile) GetPalette(entry int) (color.Palette, error){
    reader := bytes.NewReader(lbx.Data[entry])

    _, err := ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    _, err = ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    _, err = ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    _, err = ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    _, err = ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    _, err = ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    _, err = ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    paletteOffset, err := ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    if paletteOffset > 0 {
        paletteInfo, err := readPaletteInfo(reader, int(paletteOffset))
        if err != nil {
            return nil, err
        }

        return readPalette(reader, int(paletteInfo.Offset), int(paletteInfo.FirstColorIndex), int(paletteInfo.Count))
    }

    return nil, nil
}

func (lbx *LbxFile) ReadImages(entry int) ([]*image.Paletted, error) {
    return lbx.ReadImagesWithPalette(entry, defaultPalette, false)
}

/* a nil palette means to use the default palette */
func (lbx *LbxFile) ReadImagesWithPalette(entry int, palette color.Palette, forcePalette bool) ([]*image.Paletted, error) {
    if entry < 0 || entry >= len(lbx.Data) {
        return nil, fmt.Errorf("invalid lbx index %v, must be between 0 and %v", entry, len(lbx.Data) - 1)
    }

    reader := bytes.NewReader(lbx.Data[entry])

    width, err := ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    height, err := ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    currentFrame, err := ReadUint16(reader)
    if err != nil {
        return nil, err
    }
    _ = currentFrame

    bitmapCount, err := ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    loopCount, err := ReadUint16(reader)
    if err != nil {
        return nil, err
    }
    _ = loopCount

    unknown3, err := ReadUint16(reader)
    if err != nil {
        return nil, err
    }
    _ = unknown3

    unknown4, err := ReadUint16(reader)
    if err != nil {
        return nil, err
    }
    _ = unknown4

    paletteOffset, err := ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    unknown5, err := ReadUint16(reader)
    if err != nil {
        return nil, err
    }
    _ = unknown5

    if debug {
        log.Printf("currentFrame: %v loopCount: %v unknown3: %v unknown4: %v unknown5: %v\n", currentFrame, loopCount, unknown3, unknown4, unknown5)
    }

    var offsets []uint32
    for i := 0; i < int(bitmapCount) + 1; i++ {
        offset, err := ReadUint32(reader)
        if err != nil {
            return nil, err
        }

        offsets = append(offsets, offset)
    }

    if debug {
        log.Printf("%v: Width: %v\n", entry, width)
        log.Printf("%v: Height: %v\n", entry, height)
        log.Printf("%v: Bitmap count: %v\n", entry, bitmapCount)
        log.Printf("%v: Palette offset: %v\n", entry, paletteOffset)
    }

    var paletteInfo PaletteInfo

    // if forcePalette is true then ignore the built-in palette in the image
    if paletteOffset > 0 && !forcePalette {

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
            log.Printf("Entry %v custom palette", entry)
            for i, c := range palette {
                r, g, b, _ := c.RGBA()
                fmt.Printf("  %v: r=0x%x g=0x%x b=0x%x\n", i, r/255, g/255, b/255)
            }
        }
    }

    if palette == nil {
        palette = defaultPalette
    }

    /*
    r, g, b, _ := palette[232].RGBA()
    palette[232] = premultiply(color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 96})
    */

    // palette[241] = color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff}

    /* if the palette is empty then just use the default palette */
    if paletteInfo.Count == 0 {
        paletteInfo.FirstColorIndex = 0
        paletteInfo.Count = 255
    }

    if debug {
        log.Printf("Palette info: %+v\n", paletteInfo)
    }

    var images []*image.Paletted

    var lastImage *image.Paletted
    for i := 0; i < int(bitmapCount); i++ {
        end := offsets[i+1]
        if debug {
            log.Printf("Read entry %v image %v at offset %v size %v\n", entry, i, offsets[i], end - offsets[i])
        }

        reader.Seek(int64(offsets[i]), io.SeekStart)

        imageReader := io.LimitReader(reader, int64(end - offsets[i]))

        // img := image.NewPaletted(image.Rect(0, 0, int(width), int(height)), palette)

        img, err := readImage(imageReader, lastImage, int(width), int(height), palette, int(paletteInfo.FirstColorIndex + paletteInfo.Count))
        if err != nil {
            return nil, err
        }

        lastImage = img

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

func readStringsSection(reader io.ReadSeeker, start int64, end uint32) []string {
    var out []string

    if int64(end) > start {
        reader.Seek(start, io.SeekStart)

        data := make([]byte, int64(end) - start)
        reader.Read(data)

        parts := bytes.Split(data, []byte{0})

        for _, part := range parts {
            if len(part) > 0 {
                out = append(out, string(part))
            }
        }
    }

    return out
}

const LbxSignature = 0x0000fead

func ReadLbx(reader io.ReadSeeker) (LbxFile, error) {
    numFiles, err := ReadUint16(reader)
    if err != nil {
        return LbxFile{}, err
    }

    signature, err := ReadUint32(reader)
    if err != nil {
        return LbxFile{}, err
    }

    if signature != LbxSignature {
        return LbxFile{}, fmt.Errorf("Invalid lbx signature, was 0x%x but expected 0x%x\n", signature, LbxSignature)
    }

    version, err := ReadUint16(reader)
    if err != nil {
        return LbxFile{}, err
    }

    // fmt.Printf("Version: %v\n", version)

    var offsets []uint32

    for i := 0; i < int(numFiles); i++ {
        offset, err := ReadUint32(reader)
        if err != nil {
            return LbxFile{}, err
        }

        // fmt.Printf("Offset %v: 0x%x\n", i, offset)

        offsets = append(offsets, offset)
    }

    // the last 4 bytes are the size of the file
    ReadUint32(reader)

    currentPosition, _ := reader.Seek(0, io.SeekCurrent)

    // but lets just get the true end of the file ourselves
    reader.Seek(0, io.SeekEnd)
    lastByte, _ := reader.Seek(0, io.SeekCurrent)

    var lbx LbxFile

    lbx.Signature = signature
    lbx.Version = version

    /*
    if len(offsets) > 0 {
        log.Printf("Position after offsets 0x%x first offset 0x%x difference 0x%x\n", currentPosition, offsets[0], offsets[0] - uint32(currentPosition))
    }
    */

    lbx.Strings = readStringsSection(reader, currentPosition, offsets[0])

    // log.Printf("Strings: %v", strings)

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

func SaveLbx(lbx LbxFile, writer io.WriteSeeker) error {
    err := WriteUint16(writer, uint16(len(lbx.Data)))
    if err != nil {
        return err
    }

    err = WriteUint32(writer, LbxSignature)
    if err != nil {
        return err
    }

    // FIXME: does the version matter?
    err = WriteUint16(writer, 0)
    if err != nil {
        return err
    }

    // fmt.Printf("Version: %v\n", version)

    var offsets []uint32

    offsetPosition, _ := writer.Seek(0, io.SeekCurrent)

    for range lbx.Data {
        WriteUint32(writer, 0) // placeholder
    }

    WriteUint32(writer, 0) // placeholder for the final size

    // write null-terminated strings
    for _, s := range lbx.Strings {
        writer.Write([]byte(s))
        writer.Write([]byte{0})
    }

    // log.Printf("Strings: %v", strings)

    for _, data := range lbx.Data {
        offset, err := writer.Seek(0, io.SeekCurrent)
        if err != nil {
            return err
        }

        offsets = append(offsets, uint32(offset))

        reader := bytes.NewReader(data)
        io.Copy(writer, reader)
    }

    _, err = writer.Seek(offsetPosition, io.SeekStart)
    if err != nil {
        return err
    }

    for _, offset := range offsets {
        err := WriteUint32(writer, offset)
        if err != nil {
            return err
        }
    }

    currentPosition, _ := writer.Seek(0, io.SeekCurrent)

    // the last 4 bytes are the size of the file
    err = WriteUint32(writer, uint32(currentPosition + 4))
    if err != nil {
        return err
    }

    return nil
}

/* some lbx images implicitly use a palette from a specific entry
 * return a mapping from lbx entry -> palette to be used for that entry, e.g.
 *   lbx.ReadImagesWithPalette(3, overrideMap[3])
 */
func GetPaletteOverrideMap(cache *LbxCache, lbxFile *LbxFile, filename string) (map[int]color.Palette, error) {
    clean := strings.ToLower(filepath.Base(filename))
    out := make(map[int]color.Palette)

    switch clean {
        case "resource.lbx":
            palette, err := lbxFile.GetPalette(7)
            if err != nil {
                return nil, err
            }
            out[5] = palette
            out[6] = palette
            out[8] = palette
            out[9] = palette
            out[10] = palette

            paletteEvent, err := lbxFile.GetPalette(40)
            if err != nil {
                return nil, err
            }

            out[41] = paletteEvent
            out[42] = paletteEvent

            paletteHire, err := lbxFile.GetPalette(43)
            if err != nil {
                return nil, err
            }

            out[44] = paletteHire
            out[45] = paletteHire

            paletteItem, err := lbxFile.GetPalette(46)
            if err != nil {
                return nil, err
            }

            out[47] = paletteItem
            out[48] = paletteItem

            shatterPalette := clonePalette(defaultPalette)
            shatterPalette[254] = premultiply(color.RGBA{R: 255, G: 0, B: 0, A: 128})
            out[79] = shatterPalette

        case "spellose.lbx":
            palette, err := lbxFile.GetPalette(28)
            if err != nil {
                return nil, err
            }

            out[-1] = palette
        case "mainscrn.lbx":
            palette, err := lbxFile.GetPalette(0)
            if err != nil {
                return nil, err
            }

            r, g, b, _ := palette[0].RGBA()
            palette[0] = color.RGBA{R: uint8(r/255), G: uint8(g/255), B: uint8(b/255), A: 0}

            out[-1] = palette
        case "load.lbx":
            palette, err := lbxFile.GetPalette(0)
            if err != nil {
                return nil, err
            }
            out[-1] = palette
        case "conquest.lbx":
            wizlab, err := cache.GetLbxFile("wizlab.lbx")
            if err == nil {
                palette, err := wizlab.GetPalette(19)
                if err != nil {
                    return nil, err
                }
                paletteTransparent := clonePalette(palette)
                paletteTransparent[0] = color.RGBA{R: 0, G: 0, B: 0, A: 0}

                out[-1] = paletteTransparent
            }
        case "magic.lbx":
            palette, _ := lbxFile.GetPalette(0)
            out[-1] = palette
        case "lose.lbx":
        case "splmastr.lbx":
            wizlab, err := cache.GetLbxFile("wizlab.lbx")
            if err == nil {
                palette, err := wizlab.GetPalette(19)
                if err != nil {
                    return nil, err
                }
                paletteTransparent := clonePalette(palette)
                paletteTransparent[0] = color.RGBA{R: 0, G: 0, B: 0, A: 0}

                out[-1] = paletteTransparent
            }
        case "wizlab.lbx":
            palette, err := lbxFile.GetPalette(19)
            if err != nil {
                return nil, err
            }

            paletteTransparent := clonePalette(palette)
            paletteTransparent[0] = color.RGBA{R: 0, G: 0, B: 0, A: 0}

            out[-1] = paletteTransparent
        case "win.lbx":
            for i := 3; i <= 16; i++ {
                palette, err := lbxFile.GetPalette(i)
                if err != nil {
                    continue
                }
                palette = clonePalette(palette)
                palette[0] = color.RGBA{R: 0, G: 0, B: 0, A: 0}
                out[i] = palette
            }

        case "unitview.lbx":
            palette, err := lbxFile.GetPalette(0)
            if err != nil {
                return nil, err
            }
            out[-1] = palette
        case "vortex.lbx":
            // FIXME

            /*
            palette, err := lbxFile.GetPalette(0)
            if err != nil {
                return nil, err
            }

            out[-1] = palette
            */
    }

    return out, nil
}
