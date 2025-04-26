package xmi

// https://moddingwiki.shikadi.net/wiki/XMI_Format
// https://www.vgmpf.com/Wiki/index.php?title=XMI

import (
    "io"
    "fmt"
    "bytes"
    "bufio"
    "encoding/binary"
    "math"
    "cmp"
    "slices"

    "github.com/kazzmir/master-of-magic/lib/midi/smf"
    "github.com/kazzmir/master-of-magic/lib/midi"
)

const debug = false

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

func readMidiLength(reader *bufio.Reader) (int64, error) {
    length := int64(0)
    for {
        b, err := reader.ReadByte()
        if err != nil {
            return 0, err
        }

        length = (length << 7) | (int64(b) & 0x7f)

        if b & 0x80 == 0 {
            break
        }
    }

    return length, nil
}

// https://www.ccarh.org/courses/253/handout/smf/
type MidiMetaEventKind uint8
const (
    MidiEventSequenceNumber MidiMetaEventKind = 0x00
    MidiEventTextEvent MidiMetaEventKind = 0x01
    MidiEventCopyrightNotice MidiMetaEventKind = 0x02
    MidiEventSequenceTrackName MidiMetaEventKind = 0x03
    MidiEventInstrumentName MidiMetaEventKind = 0x04
    MidiEventLyric MidiMetaEventKind = 0x05
    MidiEventMarker MidiMetaEventKind = 0x06
    MidiEventCuePoint MidiMetaEventKind = 0x07

    MidiEventChannelPrefix MidiMetaEventKind = 0x20
    MidiEventPortPrefix MidiMetaEventKind = 0x21
    MidiEventEndOfTrack MidiMetaEventKind = 0x2f
    MidiEventTempoSetting MidiMetaEventKind = 0x51
    MidiEventSMPTEOffset MidiMetaEventKind = 0x54
    MidiEventTimeSignature MidiMetaEventKind = 0x58
    MidiEventKeySignature MidiMetaEventKind = 0x59
)

const (
    MidiMetaEvent uint8 = 0xff
)

type MidiMessageValue uint8
const (
    MidiMessageNoteOnValue = 0b1001
    MidiMessageControlChangeValue = 0b1011
    MidiMessageProgramChangeValue = 0b1100
    MidiMessageChannelPressureValue = 0b1101 // after touch
    MidiMessagePitchWheelChangeValue = 0b1110
)

type MidiEvent struct {
    Messages []MidiMessage
}

type NoteOffDuration struct {
    Channel uint8
    Note uint8
    Duration int64
    // used for sorting
    index int
}

// FIXME: add a ConvertToMeltySynthMidi instead of smf

// convert from an XMI midi event to a standard midi event by keeping track of note on durations
func (event *MidiEvent) ConvertToSMF() *smf.SMF {
    object := smf.New()

    // FIXME: where does 97 come from? all the master of magic files seem to use it
    object.TimeFormat = smf.MetricTicks(97)

    var track smf.Track

    var future []NoteOffDuration

    // this value will get adjusted once we see a tempo setting (MidiMessageTempoSetting)
    trueTempo := 80.0

    /* tempo setting hex value: 19 f3 8d
      at bpm: 35.28
      delay 34 -> delta 15
      delay 50 -> delta 22
      delay 66 -> delta 29

      tempo setting hex value: 09 f3 8d
      at bpm: 92
      delay 34 -> delta 39
      delay 50 -> delta 58
      delay 66 -> delta 76

      it looks like the formula is delay * (tempo/80)
    */
    adjustDelay := func(delay int64) int64 {
        // return int64(math.Round(float64(delay) * (float64(trueTempo) / 80.0)))
        return int64(math.Round(float64(delay) * (float64(trueTempo) / 80.0)))
    }

    currentDelay := uint32(0)
    for messageIndex, message := range event.Messages {
        switch message.(type) {
            case *MidiMessageSMPTEOffset:
                offset := message.(*MidiMessageSMPTEOffset)
                // FIXME: what to do with fps?
                if debug {
                    fmt.Printf(" emit smpte hour: %v\n", offset.Hour)
                }
                track.Add(currentDelay, smf.MetaSMPTE(offset.Hour, offset.Minute, offset.Second, offset.Frame, offset.SubFrame).Bytes())
                currentDelay = 0
            case *MidiMessageTimeSignature:
                signature := message.(*MidiMessageTimeSignature)
                if debug {
                    fmt.Printf("Set time signature to %+v\n", signature)
                }
                track.Add(currentDelay, smf.MetaTimeSig(signature.Numerator, 1<<signature.Denominator, signature.Metronome, signature.DemiSemiQuaverPerQuarter).Bytes())
                // track.Add(currentDelay, smf.MetaTimeSig(4, 4, signature.Metronome, signature.DemiSemiQuaverPerQuarter).Bytes())
                currentDelay = 0
            case *MidiMessageChannelPrefix:
                prefix := message.(*MidiMessageChannelPrefix)
                track.Add(currentDelay, smf.MetaChannel(prefix.Channel).Bytes())
                currentDelay = 0
            case *MidiMessagePortPrefix:
                prefix := message.(*MidiMessagePortPrefix)
                track.Add(currentDelay, smf.MetaPort(prefix.Port).Bytes())
                currentDelay = 0
            case *MidiMessageTempoSetting:
                tempo := message.(*MidiMessageTempoSetting)
                // TODO: figure out where this constant comes from
                trueTempo = 60000000.0 / tempo.Tempo
                track.Add(currentDelay, smf.MetaTempo(trueTempo).Bytes())

                // track.Add(currentDelay, smf.MetaTempo(131).Bytes())

                currentDelay = 0
            case *MidiMessageProgramChange:
                program := message.(*MidiMessageProgramChange)
                track.Add(currentDelay, midi.ProgramChange(program.Channel, program.Program).Bytes())
                currentDelay = 0
            case *MidiMessagePitchWheelChange:
                pitch := message.(*MidiMessagePitchWheelChange)
                track.Add(currentDelay, midi.Pitchbend(pitch.Channel, pitch.Value).Bytes())
                currentDelay = 0
            case *MidiMessageKeySignature:
                key := message.(*MidiMessageKeySignature)
                flats := key.Flats
                if flats < 0 {
                    flats = -flats
                }
                track.Add(currentDelay, smf.MetaKey(key.Major, key.Major == 0, uint8(flats), key.Flats < 0).Bytes())
                currentDelay = 0
            case *MidiMessageNoteOn:
                note := message.(*MidiMessageNoteOn)
                if debug {
                    fmt.Printf("[%d] Note on: channel=%v note=%v velocity=%v duration=%v\n", messageIndex, note.Channel, note.Note, note.Velocity, adjustDelay(note.Duration))
                }
                track.Add(currentDelay, midi.NoteOn(note.Channel, note.Note, note.Velocity).Bytes())
                future = append(future, NoteOffDuration{Channel: note.Channel, Note: note.Note, Duration: adjustDelay(note.Duration), index: messageIndex})
                // track.Add(uint32(note.Duration), midi.NoteOffVelocity(note.Channel, note.Note, 127).Bytes())
                currentDelay = 0
            case *MidiMessageChannelPressure:
                pressure := message.(*MidiMessageChannelPressure)
                track.Add(currentDelay, midi.AfterTouch(pressure.Channel, pressure.Pressure).Bytes())
                currentDelay = 0
            case *MidiMessageControlChange:
                control := message.(*MidiMessageControlChange)
                track.Add(currentDelay, midi.ControlChange(control.Channel, control.Controller, control.Value).Bytes())
                currentDelay = 0
            case *MidiMessageDelay:
                delay := message.(*MidiMessageDelay)
                currentDelay = uint32(adjustDelay(delay.Delay))

                // sort by duration so we emit a note off for the shortest duration first
                slices.SortFunc(future, func (a, b NoteOffDuration) int {
                    if a.Duration == b.Duration {
                        // this sort of matches what xmi2mid does
                        // reverse order for some reason, I don't know why
                        return cmp.Compare(b.index, a.index)
                    }
                    return cmp.Compare(a.Duration, b.Duration)
                })

                /* compute which notes to turn off */
                sofar := int64(0)
                var more []NoteOffDuration
                for _, note := range future {
                    note.Duration -= sofar
                    if note.Duration < 0 {
                        note.Duration = 0
                    }
                    if note.Duration - int64(currentDelay) <= 0 {
                        track.Add(uint32(note.Duration), midi.NoteOffVelocity(note.Channel, note.Note, 127).Bytes())
                        if debug {
                            fmt.Printf("[%d] emit note off: delta=%v channel=%v note=%v\n", messageIndex, note.Duration, note.Channel, note.Note)
                        }
                        sofar += note.Duration
                        currentDelay -= uint32(note.Duration)
                    } else {
                        note.Duration -= int64(currentDelay)
                        more = append(more, note)
                    }
                }
                future = more
                if debug {
                    if len(future) > 0 {
                        fmt.Printf("future notes: %+v\n", future)
                    }
                }

            case *MidiMessageEndOfTrack:
                if debug {
                    fmt.Printf("[%d] End of track\n", messageIndex)
                }
                track.Close(currentDelay)
            default:
                fmt.Printf("Unhandled midi message in conversion: %T\n", message)
        }
    }

    object.Add(track)

    return object
}

// midi message events that XMI supports. Notably there is no note off
type MidiMessage interface {
}

type MidiMessageSMPTEOffset struct {
    Fps float32
    Hour uint8
    Minute uint8
    Second uint8
    Frame uint8
    SubFrame uint8
}

type MidiMessageKeySignature struct {
    Flats int8
    Major uint8
}

type MidiMessageTimeSignature struct {
    Numerator uint8
    Denominator uint8
    Metronome uint8
    DemiSemiQuaverPerQuarter uint8
}

type MidiMessageTempoSetting struct {
    Tempo float64
}

type MidiMessageChannelPrefix struct {
    Channel uint8
}

type MidiMessagePortPrefix struct {
    Port uint8
}

type MidiMessageEndOfTrack struct {
}

type MidiMessageNoteOn struct {
    Channel uint8
    Note uint8
    Velocity uint8
    Duration int64
}

type MidiMessageControlChange struct {
    Channel uint8
    Controller uint8
    Value uint8
}

type MidiMessageProgramChange struct {
    Channel uint8
    Program uint8
}

type MidiMessageChannelPressure struct {
    Channel uint8
    Pressure uint8
}

type MidiMessagePitchWheelChange struct {
    Channel uint8
    Value int16
}

type MidiMessageDelay struct {
    Delay int64
}

func (chunk *IFFChunk) ReadEvent() (MidiEvent, error) {
    // fmt.Printf("Data: %v\n", chunk.Data[0:20])

    reader := bufio.NewReader(bytes.NewReader(chunk.Data))

    var messages []MidiMessage

    for {
        value, err := reader.ReadByte()
        if err != nil {
            break
        }

        // fmt.Printf("Event 0x%x\n", value)

        // check high bit to see if its a delay
        isDelay := value & 0x80 == 0

        if !isDelay {
            switch value {
                // meta event
                case MidiMetaEvent:
                    kind, err := reader.ReadByte()
                    if err != nil {
                        return MidiEvent{}, err
                    }

                    length, err := readMidiLength(reader)
                    if err != nil {
                        return MidiEvent{}, err
                    }

                    data := make([]byte, length)
                    n, err := reader.Read(data)
                    if n != len(data) {
                        return MidiEvent{}, fmt.Errorf("Expected %v bytes, got %v", len(data), n)
                    }
                    if err != nil {
                        return MidiEvent{}, err
                    }

                    // fmt.Printf("  Meta event: 0x%x length=%v data=%v\n", kind, length, data)

                    switch MidiMetaEventKind(kind) {
                    case MidiEventSMPTEOffset:
                        if len(data) != 5 {
                            return MidiEvent{}, fmt.Errorf("SMPTE event type has invalid length: %v", len(data))
                        }
                        hours := uint8(data[0])
                        rate := hours >> 6
                        hours = hours & 0x3f

                        fps := float32(24.0)

                        switch rate {
                            case 0: fps = 24.0
                            case 1: fps = 25.0
                            case 2: fps = 29.97
                            case 3: fps = 30.0
                        }

                        // data[0] & 0x1f is hours
                        // data[1] is minutes
                        // data[2] is seconds
                        // data[3] is frames
                        // data[4] is sub-frames

                        messages = append(messages, &MidiMessageSMPTEOffset{
                            Hour: hours,
                            Minute: data[1],
                            Second: data[2],
                            Frame: data[3],
                            SubFrame: data[4],
                            Fps: fps,
                        })
                    case MidiEventKeySignature:
                        if len(data) != 2 {
                            return MidiEvent{}, fmt.Errorf("Key signature event type has invalid length: %v", len(data))
                        }

                        flats := int8(data[0])
                        major := data[1]

                        // flats: -7 to 7
                        // major: 0 for major key, 1 for minor key
                        _ = flats
                        _ = major

                        messages = append(messages, &MidiMessageKeySignature{
                            Flats: flats,
                            Major: major,
                        })

                    case MidiEventTimeSignature:

                        if len(data) != 4 {
                            return MidiEvent{}, fmt.Errorf("Time signature event type has invalid length: %v", len(data))
                        }

                        messages = append(messages, &MidiMessageTimeSignature{
                            Numerator: data[0],
                            Denominator: data[1],
                            Metronome: data[2],
                            DemiSemiQuaverPerQuarter: data[3],
                        })
                    case MidiEventTempoSetting:
                        if len(data) != 3 {
                            return MidiEvent{}, fmt.Errorf("Tempo setting event type has invalid length: %v", len(data))
                        }

                        // tempo events are ignored
                        messages = append(messages, &MidiMessageTempoSetting{
                            Tempo: float64(binary.BigEndian.Uint32(append([]byte{0}, data...))),
                        })
                    case MidiEventChannelPrefix:
                        if len(data) != 1 {
                            return MidiEvent{}, fmt.Errorf("Channel prefix event type has invalid length: %v", len(data))
                        }

                        messages = append(messages, &MidiMessageChannelPrefix{
                            Channel: data[0],
                        })
                    case MidiEventPortPrefix:
                        if len(data) != 1 {
                            return MidiEvent{}, fmt.Errorf("Port prefix event type has invalid length: %v", len(data))
                        }

                        messages = append(messages, &MidiMessagePortPrefix{
                            Port: data[0],
                        })
                    case MidiEventEndOfTrack:
                        messages = append(messages, &MidiMessageEndOfTrack{
                        })
                    default:
                        return MidiEvent{}, fmt.Errorf("Unknown midi meta event type: 0x%x", kind)
                    }
                default:
                    message := value >> 4
                    channel := value & 0x0f

                    switch MidiMessageValue(message) {
                        case MidiMessageNoteOnValue:
                            // The first difference is "Note On" event contains 3 parameters - the note number, velocity level (same as standard MIDI), and also duration in ticks. Duration is stored as variable-length value in concatenated bits format. Since note events store information about its duration, there are no "Note Off" events.

                            note, err := reader.ReadByte()
                            if err != nil {
                                return MidiEvent{}, err
                            }

                            velocity, err := reader.ReadByte()
                            if err != nil {
                                return MidiEvent{}, err
                            }

                            duration, err := readMidiLength(reader)
                            if err != nil {
                                return MidiEvent{}, err
                            }

                            // fmt.Printf("  note on note=%v velocity=%v duration=%v\n", note, velocity, duration)

                            messages = append(messages, &MidiMessageNoteOn{
                                Channel: channel,
                                Note: note,
                                Velocity: velocity,
                                Duration: duration,
                            })

                            if debug {
                                fmt.Printf("[%d] Note on: channel=%v note=%v velocity=%v duration=%v\n", len(messages), channel, note, velocity, duration)
                            }

                        case MidiMessageControlChangeValue:
                            controller, err := reader.ReadByte()
                            if err != nil {
                                return MidiEvent{}, err
                            }
                            newValue, err := reader.ReadByte()
                            if err != nil {
                                return MidiEvent{}, err
                            }

                            messages = append(messages, &MidiMessageControlChange{
                                Channel: channel,
                                Controller: controller,
                                Value: newValue,
                            })

                            if debug {
                                fmt.Printf("[%d] Control change: channel=%v controller=%v value=%v\n", len(messages), channel, controller, newValue)
                            }

                        case MidiMessageProgramChangeValue:
                            program, err := reader.ReadByte()
                            if err != nil {
                                return MidiEvent{}, err
                            }

                            messages = append(messages, &MidiMessageProgramChange{
                                Channel: channel,
                                Program: program,
                            })

                        case MidiMessageChannelPressureValue:
                            pressure, err := reader.ReadByte()
                            if err != nil {
                                return MidiEvent{}, err
                            }

                            messages = append(messages, &MidiMessageChannelPressure{
                                Channel: channel,
                                Pressure: pressure,
                            })
                        case MidiMessagePitchWheelChangeValue:
                            low, err := reader.ReadByte()
                            if err != nil {
                                return MidiEvent{}, err
                            }
                            high, err := reader.ReadByte()
                            if err != nil {
                                return MidiEvent{}, err
                            }

                            total := int16(high) << 7 | int16(low)
                            /* 8192 is in the middle, move it to 0 (neutral) */
                            total -= 8192

                            messages = append(messages, &MidiMessagePitchWheelChange{
                                Channel: channel,
                                Value: int16(total),
                            })

                            /*
                        case MidiEventSequenceNumber:
                        case MidiEventTextEvent:
                        case MidiEventCopyrightNotice:
                        case MidiEventSequenceTrackName:
                        case MidiEventInstrumentName:
                        case MidiEventLyric:
                        case MidiEventMarker:
                        case MidiEventCuePoint:
                            */

                        default:
                            return MidiEvent{}, fmt.Errorf("Unknown midi event type: 0x%x", value)
                    }
            }
        } else {
            delay := int64(value)
            for value == 0x7f {
                value, err = reader.ReadByte()
                if err != nil {
                    return MidiEvent{}, err
                }
                delay += int64(value)
            }


            messages = append(messages, &MidiMessageDelay{
                Delay: delay,
            })

            if debug {
                fmt.Printf("[%d] Delay of %v\n", len(messages), delay)
            }
        }
    }

    return MidiEvent{Messages: messages}, nil
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
    n, err := io.ReadFull(reader.reader, data)
    if err != nil {
        return IFFChunk{}, err
    }

    if n != int(size) {
        return IFFChunk{}, fmt.Errorf("Expected %v bytes, got %v", size, n)
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

func ConvertToMidi(data io.Reader) (*smf.SMF, error) {
    reader := bufio.NewReader(data)
    // if reading from an lbx file
    // reader.Discard(16)

    iffReader := NewIFFReader(reader)

    for iffReader.HasMore() {
        chunk, err := iffReader.ReadChunk()
        if err != nil {
            return nil, err
        }
        if debug {
            fmt.Printf("Chunk name=%v size=%v\n", chunk.Name(), chunk.Size())
        }
        if chunk.IsForm() && chunk.GetFormType() == "XDIR" {
            subChunk := chunk.SubChunk()
            if debug {
                if subChunk.IsInfo() {
                    fmt.Printf("  sequences: %v\n", subChunk.GetInfoSequence())
                }
            }
            // fmt.Printf("  %v\n", string(chunk.RawData()))
        } else if chunk.IsCat() {
            subChunk := chunk.SubChunk()
            if debug {
                fmt.Printf("Cat subchunk name=%v size=%v form type=%v\n", subChunk.Name(), subChunk.Size(), subChunk.GetFormType())
            }
            sub2 := subChunk.SubChunk()
            if debug {
                fmt.Printf("Cat sub2 name=%v size=%v\n", sub2.Name(), sub2.Size())
            }

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

                if debug {
                    fmt.Printf("  next subchunk name=%v size=%v\n", next.Name(), next.Size())
                }

                if next.IsTimbre() {
                    timbre, err := next.ReadTimbre()
                    if err != nil {
                        fmt.Printf("Error reading timbre: %v\n", err)
                    } else {
                        if debug {
                            fmt.Printf("  timbre entries: %v\n", len(timbre.Entries))
                        }
                    }
                } else if next.IsEvent() {
                    event, err := next.ReadEvent()
                    if err != nil {
                        return nil, err
                        // fmt.Printf("Error reading event: %v\n", err)
                        // continue
                    }

                    smfObject := event.ConvertToSMF()
                    return smfObject, nil
                } else {
                    fmt.Printf("  unknown subchunk\n")
                }
            }

            /*
            xmidChunk := sub2.SubChunk()
            fmt.Printf("  xmid chunk name=%v size=%v\n", xmidChunk.Name(), xmidChunk.Size())
            */
        }
    }

    return nil, fmt.Errorf("no midi data found")
}
