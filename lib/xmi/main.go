package main

// https://moddingwiki.shikadi.net/wiki/XMI_Format
// https://www.vgmpf.com/Wiki/index.php?title=XMI

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

type IFFTimbreEntry struct {
    Patch uint8
    Bank uint8
}

type IFFTimbre struct {
    Entries []IFFTimbreEntry
}

type IFFEvent struct {
}

type IFFChunk struct {
    ID   []byte
    Data []byte
}

func (chunk *IFFChunk) IsForm() bool {
    return string(chunk.ID) == "FORM"
}

func (chunk *IFFChunk) IsEvent() bool {
    return string(chunk.ID) == "EVNT"
}

func (chunk *IFFChunk) IsCat() bool {
    return string(chunk.ID) == "CAT "
}

func (chunk *IFFChunk) GetFormType() string {
    return string(chunk.Data[:4])
}

func (chunk *IFFChunk) IsInfo() bool {
    return string(chunk.ID) == "INFO"
}

func (chunk *IFFChunk) IsTimbre() bool {
    return string(chunk.ID) == "TIMB"
}

func (chunk *IFFChunk) GetInfoSequence() int {
    b1 := int(chunk.Data[0])
    b2 := int(chunk.Data[1])
    return b1 + b2 * 256
}

func (chunk *IFFChunk) ReadTimbre() (IFFTimbre, error) {
    reader := bufio.NewReader(bytes.NewReader(chunk.Data))

    count, err := readUint16LE(reader)
    if err != nil {
        return IFFTimbre{}, err
    }

    var entries []IFFTimbreEntry

    for i := 0; i < int(count); i++ {
        patch, err := reader.ReadByte()
        if err != nil {
            return IFFTimbre{}, err
        }

        bank, err := reader.ReadByte()
        if err != nil {
            return IFFTimbre{}, err
        }

        entries = append(entries, IFFTimbreEntry{Patch: patch, Bank: bank})
    }

    return IFFTimbre{Entries: entries}, nil
}

func readMidiLength(reader *bufio.Reader) (int, error) {
    length := 0
    for {
        b, err := reader.ReadByte()
        if err != nil {
            return 0, err
        }

        length = (length << 7) | (int(b) & 0x7f)
        if b & 0x80 == 0 {
            break
        }
    }

    return length, nil
}

// https://www.ccarh.org/courses/253/handout/smf/
type MidiMetaEventKind uint8
const (
    MidiEventChannelPrefix MidiMetaEventKind = 0x20
    MidiEventEndOfTrack MidiMetaEventKind = 0x2f
    MidiEventTempoSetting MidiMetaEventKind = 0x51
    MidiEventSMPTEOffset MidiMetaEventKind = 0x54
    MidiEventTimeSignature MidiMetaEventKind = 0x58
    MidiEventKeySignature MidiMetaEventKind = 0x59
)

const (
    MidiMetaEvent uint8 = 0xff
)

type MidiMessage uint8
const (
    MidiMessageNoteOn = 0b1001
    MidiMessageControlChange = 0b1011
    MidiMessageProgramChange = 0b1100
    MidiMessageChannelPressure = 0b1101 // after touch
    MidiMessagePitchWheelChange = 0b1110
)

func (chunk *IFFChunk) ReadEvent() (IFFEvent, error) {
    fmt.Printf("Data: %v\n", chunk.Data[0:20])

    reader := bufio.NewReader(bytes.NewReader(chunk.Data))

    for {
        value, err := reader.ReadByte()
        if err != nil {
            break
        }

        fmt.Printf("Event 0x%x\n", value)

        // check high bit to see if its a delay
        isDelay := value & 0x80 == 0

        if !isDelay {
            switch value {
                // meta event
                case MidiMetaEvent:
                    kind, err := reader.ReadByte()
                    if err != nil {
                        return IFFEvent{}, err
                    }

                    length, err := readMidiLength(reader)
                    if err != nil {
                        return IFFEvent{}, err
                    }

                    data := make([]byte, length)
                    n, err := reader.Read(data)
                    if n != len(data) {
                        return IFFEvent{}, fmt.Errorf("Expected %v bytes, got %v", len(data), n)
                    }
                    if err != nil {
                        return IFFEvent{}, err
                    }

                    fmt.Printf("  Meta event: 0x%x length=%v data=%v\n", kind, length, data)

                    switch MidiMetaEventKind(kind) {
                    case MidiEventSMPTEOffset:
                        if len(data) != 5 {
                            return IFFEvent{}, fmt.Errorf("SMPTE event type has invalid length: %v", len(data))
                        }
                        hours := uint8(data[0])
                        rate := hours >> 5

                        switch rate {
                        case 0:
                            // 24 frames per second
                        case 1:
                            // 25 frames per second
                        case 2:
                            // 29.97 frames per second
                        case 3:
                            // 30 frames per second
                        }

                        // data[0] & 0x1f is hours
                        // data[1] is minutes
                        // data[2] is seconds
                        // data[3] is frames
                        // data[4] is sub-frames
                    case MidiEventKeySignature:
                        if len(data) != 2 {
                            return IFFEvent{}, fmt.Errorf("Key signature event type has invalid length: %v", len(data))
                        }

                        flats := int8(data[0])
                        major := data[1]

                        // flats: -7 to 7
                        // major: 0 for major key, 1 for minor key
                        _ = flats
                        _ = major
                    case MidiEventTimeSignature:
                        // TODO
                    case MidiEventTempoSetting:
                        // TODO
                    case MidiEventChannelPrefix:
                        // TODO
                    case MidiEventEndOfTrack:
                        // TODO
                    default:
                        fmt.Printf("unknown midi meta event type: 0x%x\n", kind)
                    }
                default:
                    message := value >> 4
                    channel := value & 0x0f

                    _ = channel

                    switch MidiMessage(message) {
                        case MidiMessageNoteOn:
                            // The first difference is "Note On" event contains 3 parameters - the note number, velocity level (same as standard MIDI), and also duration in ticks. Duration is stored as variable-length value in concatenated bits format. Since note events store information about its duration, there are no "Note Off" events.

                            note, err := reader.ReadByte()
                            if err != nil {
                                return IFFEvent{}, err
                            }
                            velocity, err := reader.ReadByte()
                            if err != nil {
                                return IFFEvent{}, err
                            }

                            duration, err := readMidiLength(reader)
                            if err != nil {
                                return IFFEvent{}, err
                            }

                            fmt.Printf("  note on note=%v velocity=%v duration=%v\n", note, velocity, duration)

                        case MidiMessageControlChange:
                            controller, err := reader.ReadByte()
                            if err != nil {
                                return IFFEvent{}, err
                            }
                            newValue, err := reader.ReadByte()
                            if err != nil {
                                return IFFEvent{}, err
                            }

                            _ = controller
                            _ = newValue
                            // TODO: handle values
                        case MidiMessageProgramChange:
                            program, err := reader.ReadByte()
                            if err != nil {
                                return IFFEvent{}, err
                            }
                            _ = program
                            // TODO: handle program

                        case MidiMessageChannelPressure:
                            pressure, err := reader.ReadByte()
                            if err != nil {
                                return IFFEvent{}, err
                            }
                            _ = pressure
                            // TODO: handle pressure

                        case MidiMessagePitchWheelChange:
                            low, err := reader.ReadByte()
                            if err != nil {
                                return IFFEvent{}, err
                            }
                            high, err := reader.ReadByte()
                            if err != nil {
                                return IFFEvent{}, err
                            }

                            total := int(high) << 7 | int(low)
                            _ = total
                            // TODO: handle pitch wheel change

                        default:
                            fmt.Printf("  unknown midi event type: 0x%x\n", value)
                            return IFFEvent{}, fmt.Errorf("Unknown midi event type: 0x%x", value)
                    }
            }
        } else {
            delay := int64(value)
            for value == 0x7f {
                value, err = reader.ReadByte()
                if err != nil {
                    return IFFEvent{}, err
                }
                delay += int64(value)
            }
        }
    }

    return IFFEvent{}, nil
}

func (chunk *IFFChunk) SubChunkReader() *IFFReader {
    return &IFFReader{reader: bufio.NewReader(bytes.NewReader(chunk.Data))}
}

func (chunk *IFFChunk) SubChunk() IFFChunk {
    if chunk.IsForm() || chunk.IsCat() {
        return IFFChunk{ID: chunk.Data[4:8], Data: chunk.Data[8:]}
    }

    return IFFChunk{ID: chunk.Data[0:4], Data: chunk.Data[4:]}
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
                next, err := subChunkReader.ReadChunk()
                if err != nil {
                    fmt.Printf("Error reading subchunk: %v\n", err)
                    break
                }

                /*
                timbre, ok := next.(*IFFTimbre)
                if ok {
                    fmt.Printf("  timbre entries: %v\n", len(timbre.Entries))
                }
                */

                fmt.Printf("  next subchunk name=%v size=%v\n", next.Name(), next.Size())

                if next.IsTimbre() {
                    timbre, err := next.ReadTimbre()
                    if err != nil {
                        fmt.Printf("Error reading timbre: %v\n", err)
                    } else {
                        fmt.Printf("  timbre entries: %v\n", len(timbre.Entries))
                    }
                } else if next.IsEvent() {
                    event, err := next.ReadEvent()
                    if err != nil {
                        fmt.Printf("Error reading event: %v\n", err)
                    } else {
                        fmt.Printf("  event: %v\n", event)
                    }
                }

            }

            /*
            xmidChunk := sub2.SubChunk()
            fmt.Printf("  xmid chunk name=%v size=%v\n", xmidChunk.Name(), xmidChunk.Size())
            */
        }
    }
}
