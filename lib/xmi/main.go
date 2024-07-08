package main

import (
    "os"
    "io"
    "fmt"
    "bytes"
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

type IFFSubChunkReader struct {
    reader *bytes.Reader
}

func (reader *IFFSubChunkReader) HasMore() bool {
    return reader.reader.Len() > 0
}

func (reader *IFFSubChunkReader) ReadTimbreChunk() (IFFChunk, error) {
    return IFFChunk{}, fmt.Errorf("error handle TIMB")
}

func (reader *IFFSubChunkReader) ReadEventChunk() (IFFChunk, error) {
    return IFFChunk{}, fmt.Errorf("error handle EVNT")
}

func (reader *IFFSubChunkReader) NextChunk() (IFFChunk, error) {
    id := make([]byte, 4)
    reader.reader.Read(id)

    if string(id) == "TIMB" {
        return reader.ReadTimbreChunk()
    }

    if string(id) == "EVNT" {
        return reader.ReadEventChunk()
    }

    return IFFChunk{}, fmt.Errorf("unknown iff chunk id: %v", string(id))
}

type IFFChunk struct {
    ID   []byte
    Data []byte
}

func (chunk *IFFChunk) IsForm() bool {
    return string(chunk.ID) == "FORM"
}

func (chunk *IFFChunk) GetFormType() string {
    return string(chunk.Data[:4])
}

func (chunk *IFFChunk) IsInfo() bool {
    return string(chunk.ID) == "INFO"
}

func (chunk *IFFChunk) GetInfoSequence() int {
    b1 := int(chunk.Data[0])
    b2 := int(chunk.Data[1])
    return b1 + b2 * 256
}

func (chunk *IFFChunk) SubChunkReader() *IFFSubChunkReader {
    return &IFFSubChunkReader{reader: bytes.NewReader(chunk.Data)}
}

func (chunk *IFFChunk) SubChunk() IFFChunk {
    if chunk.IsForm() || chunk.IsCat() {
        return IFFChunk{ID: chunk.Data[4:8], Data: chunk.Data[8:]}
    }

    return IFFChunk{ID: chunk.Data[0:4], Data: chunk.Data[4:]}
}

func (chunk *IFFChunk) IsTimbre() bool {
    return string(chunk.ID) == "TIMB"
}

func (chunk *IFFChunk) IsEvent() bool {
    return string(chunk.ID) == "EVNT"
}

func (chunk *IFFChunk) IsCat() bool {
    return string(chunk.ID) == "CAT "
}

func (chunk *IFFChunk) Name() string {
    return string(chunk.ID)
}

func (chunk *IFFChunk) Size() int {
    return len(chunk.Data)
}

func (chunk *IFFChunk) RawData() []byte {
    return chunk.Data
}

type IFFReader struct {
    reader *bufio.Reader
}

func (reader *IFFReader) HasMore() bool {
    _, err := reader.reader.Peek(4)
    return err == nil
}

func (reader *IFFReader) ReadChunk() (IFFChunk, error) {
    id := make([]byte, 4)
    _, err := reader.reader.Read(id)
    if err != nil {
        return IFFChunk{}, err
    }

    size, err := readUint32BE(reader.reader)
    if err != nil {
        return IFFChunk{}, err
    }

    data := make([]byte, size)
    _, err = reader.reader.Read(data)
    if err != nil {
        return IFFChunk{}, err
    }

    if size % 2 != 0 {
        // padding byte
        _, err := reader.reader.ReadByte()
        if err != nil {
            return IFFChunk{}, err
        }
    }

    return IFFChunk{ID: id, Data: data}, nil
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
        chunk, err := iffReader.ReadChunk()
        if err != nil {
            fmt.Printf("Error: %s\n", err)
            return
        }
        fmt.Printf("Chunk name=%v size=%v\n", chunk.Name(), chunk.Size())
        if chunk.IsForm() && chunk.GetFormType() == "XDIR" {
            subChunk := chunk.SubChunk()
            if subChunk.IsInfo() {
                fmt.Printf("  sequences: %v\n", subChunk.GetInfoSequence())
            }
            // fmt.Printf("  %v\n", string(chunk.RawData()))
        } else if chunk.IsCat() {
            subChunk := chunk.SubChunk()
            fmt.Printf("Cat subchunk name=%v size=%v form type=%v\n", subChunk.Name(), subChunk.Size(), subChunk.GetFormType())
            sub2 := subChunk.SubChunk()
            fmt.Printf("Cat sub2 name=%v size=%v\n", sub2.Name(), sub2.Size())

            subChunkReader := sub2.SubChunkReader()

            for subChunkReader.HasMore() {
                next, err := subChunkReader.NextChunk()
                if err != nil {
                    fmt.Printf("Error reading subchunk: %v\n", err)
                    break
                }

                fmt.Printf("  next subchunk name=%v size=%v\n", next.Name(), next.Size())
            }

            /*
            xmidChunk := sub2.SubChunk()
            fmt.Printf("  xmid chunk name=%v size=%v\n", xmidChunk.Name(), xmidChunk.Size())
            */
        }
    }
}
