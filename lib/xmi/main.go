package main

import (
    "os"
    "io"
    "fmt"
    "bufio"
    "encoding/binary"
)

func readUint16LE(reader io.Reader) (uint16, error) {
    var value uint16
    err := binary.Read(reader, binary.LittleEndian, &value)
    return value, err
}

func readUint32LE(reader io.Reader) (uint32, error) {
    var value uint32
    err := binary.Read(reader, binary.LittleEndian, &value)
    return value, err
}

func readUint32BE(reader io.Reader) (uint32, error) {
    var value uint32
    err := binary.Read(reader, binary.BigEndian, &value)
    return value, err
}

func readInt32LE(reader io.Reader) (int32, error) {
    v, err := readUint32LE(reader)
    return int32(v), err
}

type IFFReader struct {
    reader *bufio.Reader
}

func (reader *IFFReader) HasMore() bool {
    _, err := reader.reader.Peek(4)
    return err == nil
}

func (reader *IFFReader) ReadChunk() ([]byte, []byte, error) {
    id := make([]byte, 4)
    _, err := reader.reader.Read(id)
    if err != nil {
        return nil, nil, err
    }

    size, err := readUint32BE(reader.reader)
    if err != nil {
        return nil, nil, err
    }

    data := make([]byte, size)
    _, err = reader.reader.Read(data)
    if err != nil {
        return nil, nil, err
    }

    if size % 2 != 0 {
        // padding byte
        _, err := reader.reader.ReadByte()
        if err != nil {
            return nil, nil, err
        }
    }

    return id, data, nil
}

func NewIFFReader(reader io.Reader) *IFFReader {
    return &IFFReader{
        reader: bufio.NewReader(reader),
    }
}

func main(){
    if len(os.Args) < 2 {
        return
    }
    file := os.Args[1]

    data, err := os.Open(file)
    if err != nil {
        fmt.Printf("Error: %s\n", err)
        return
    }
    defer data.Close()

    reader := bufio.NewReader(data)
    reader.Discard(16)

    iffReader := NewIFFReader(reader)

    for iffReader.HasMore() {
        form, chunk, err := iffReader.ReadChunk()
        if err != nil {
            fmt.Printf("Error: %s\n", err)
            return
        }
        fmt.Printf("Name: %v\n", string(form))
        fmt.Printf("Size: %v\n", len(chunk))
    }
}
