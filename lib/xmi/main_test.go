package xmi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/kazzmir/master-of-magic/lib/meltysynth"

	"github.com/kazzmir/master-of-magic/lib/midi"
	"github.com/kazzmir/master-of-magic/lib/midi/smf"
)

func TestReadEventAdditiveDelays(t *testing.T) {
	tests := []struct {
		name   string
		delays []byte
		want   int64
	}{
		{"zero", []byte{0}, 0},
		{"single", []byte{10}, 10},
		{"127 before status", []byte{127}, 127},
		{"consecutive 127 before status", []byte{127, 127}, 254},
		{"127 with zero terminator", []byte{127, 0}, 127},
		{"adjacent short delays", []byte{10, 20}, 30},
		{"mixed delays", []byte{10, 127, 20}, 157},
		{"adjacent zero delays", []byte{0, 0}, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := append(append([]byte{}, test.delays...), 0xc0, 5, 0xff, 0x2f, 0)
			event, err := (&IFFChunk{Data: data}).ReadEvent()
			if err != nil {
				t.Fatal(err)
			}
			wantMessages := []MidiMessage{
				&MidiMessageDelay{Delay: test.want},
				&MidiMessageProgramChange{Channel: 0, Program: 5},
				&MidiMessageEndOfTrack{},
			}
			if !reflect.DeepEqual(event.Messages, wantMessages) {
				t.Fatalf("messages = %#v, want %#v", event.Messages, wantMessages)
			}
			wantTrack := smf.Track{
				{Delta: 0, Message: smf.MetaTempo(120).Bytes()},
				{Delta: uint32(test.want), Message: midi.ProgramChange(0, 5).Bytes()},
				{Delta: 0, Message: smf.EOT},
			}
			if got := event.ConvertToSMF().Tracks[0]; !reflect.DeepEqual(got, wantTrack) {
				t.Fatalf("track = %v, want %v", got, wantTrack)
			}
		})
	}
}

// MUSIC.LBX entry 115 contains 7f 93 30 6f 25 at EVNT offset 0x2b1.
// The delay must not consume the note-on status or its parameters.
func TestReadEventEntry115DelayBeforeNote(t *testing.T) {
	data := []byte{0x7f, 0x93, 0x30, 0x6f, 0x25, 0x25, 0xff, 0x2f, 0}
	event, err := (&IFFChunk{Data: data}).ReadEvent()
	if err != nil {
		t.Fatal(err)
	}
	wantMessages := []MidiMessage{
		&MidiMessageDelay{Delay: 127},
		&MidiMessageNoteOn{Channel: 3, Note: 48, Velocity: 111, Duration: 37},
		&MidiMessageDelay{Delay: 37},
		&MidiMessageEndOfTrack{},
	}
	if !reflect.DeepEqual(event.Messages, wantMessages) {
		t.Fatalf("messages = %#v, want %#v", event.Messages, wantMessages)
	}
	wantTrack := smf.Track{
		{Delta: 0, Message: smf.MetaTempo(120).Bytes()},
		{Delta: 127, Message: midi.NoteOn(3, 48, 111).Bytes()},
		{Delta: 37, Message: midi.NoteOffVelocity(3, 48, 127).Bytes()},
		{Delta: 0, Message: smf.EOT},
	}
	if got := event.ConvertToSMF().Tracks[0]; !reflect.DeepEqual(got, wantTrack) {
		t.Fatalf("track = %v, want %v", got, wantTrack)
	}
}

func TestConvertToSMFFlushNotesAtEndOfTrack(t *testing.T) {
	note := func(key uint8, duration int64) MidiMessage {
		return &MidiMessageNoteOn{Channel: 0, Note: key, Velocity: 100, Duration: duration}
	}
	on := func(delta uint32, key uint8) smf.Event {
		return smf.Event{Delta: delta, Message: midi.NoteOn(0, key, 100).Bytes()}
	}
	off := func(delta uint32, key uint8) smf.Event {
		return smf.Event{Delta: delta, Message: midi.NoteOffVelocity(0, key, 127).Bytes()}
	}
	eot := func(delta uint32) smf.Event {
		return smf.Event{Delta: delta, Message: smf.EOT}
	}
	tests := []struct {
		name     string
		messages []MidiMessage
		want     smf.Track
	}{
		{
			name:     "no notes preserves trailing silence",
			messages: []MidiMessage{&MidiMessageDelay{Delay: 20}},
			want:     smf.Track{eot(20)},
		},
		{
			name:     "consecutive delays preserve absolute time",
			messages: []MidiMessage{&MidiMessageDelay{Delay: 10}, &MidiMessageDelay{Delay: 20}},
			want:     smf.Track{eot(30)},
		},
		{
			name:     "note with no subsequent delay",
			messages: []MidiMessage{note(60, 100)},
			want:     smf.Track{on(0, 60), off(100, 60), eot(0)},
		},
		{
			name:     "zero duration note",
			messages: []MidiMessage{note(60, 0)},
			want:     smf.Track{on(0, 60), off(0, 60), eot(0)},
		},
		{
			name:     "partially elapsed note",
			messages: []MidiMessage{note(60, 100), &MidiMessageDelay{Delay: 20}},
			want:     smf.Track{on(0, 60), off(100, 60), eot(0)},
		},
		{
			name:     "sort pending notes and preserve tie ordering",
			messages: []MidiMessage{note(60, 100), note(62, 30), note(64, 30)},
			want:     smf.Track{on(0, 60), on(0, 62), on(0, 64), off(30, 64), off(0, 62), off(70, 60), eot(0)},
		},
		{
			name:     "pending delay after an already emitted note off",
			messages: []MidiMessage{note(60, 10), note(62, 30), &MidiMessageDelay{Delay: 20}},
			want:     smf.Track{on(0, 60), on(0, 62), off(10, 60), off(20, 62), eot(0)},
		},
		{
			name:     "note added after a delay",
			messages: []MidiMessage{note(60, 100), &MidiMessageDelay{Delay: 20}, note(62, 30)},
			want:     smf.Track{on(0, 60), on(20, 62), off(30, 62), off(50, 60), eot(0)},
		},
		{
			name:     "all notes already finished preserves trailing silence",
			messages: []MidiMessage{note(60, 10), &MidiMessageDelay{Delay: 20}},
			want:     smf.Track{on(0, 60), off(10, 60), eot(10)},
		},
		{
			name:     "source tempo does not scale final note off",
			messages: []MidiMessage{&MidiMessageTempoSetting{Tempo: 375000}, note(60, 10)},
			want:     smf.Track{on(0, 60), off(10, 60), eot(0)},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event := MidiEvent{Messages: append(test.messages, &MidiMessageEndOfTrack{})}
			got := event.ConvertToSMF().Tracks[0]
			want := append(smf.Track{{Delta: 0, Message: smf.MetaTempo(120).Bytes()}}, test.want...)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("track = %v, want %v", got, want)
			}
		})
	}
}

func midiLength(t *testing.T, song *smf.SMF) time.Duration {
	t.Helper()
	if song.TimeFormat != smf.MetricTicks(60) {
		t.Fatalf("time format = %v, want 60 PPQN", song.TimeFormat)
	}
	var buffer bytes.Buffer
	if _, err := song.WriteTo(&buffer); err != nil {
		t.Fatal(err)
	}
	file, err := meltysynth.NewMidiFile(&buffer)
	if err != nil {
		t.Fatal(err)
	}
	return file.GetLength()
}

func TestConvertToSMFFixedClock(t *testing.T) {
	for _, tempo := range []float64{0, 375000, 500000, 1500000} {
		event := MidiEvent{}
		if tempo != 0 {
			event.Messages = append(event.Messages, &MidiMessageTempoSetting{Tempo: tempo})
		}
		event.Messages = append(event.Messages,
			&MidiMessageNoteOn{Channel: 0, Note: 60, Velocity: 100, Duration: 120},
			&MidiMessageDelay{Delay: 120},
			&MidiMessageEndOfTrack{},
		)
		if got := midiLength(t, event.ConvertToSMF()); got != time.Second {
			t.Fatalf("tempo %v: 120 XMI ticks played for %v, want 1s", tempo, got)
		}
	}
}

func TestConvertToSMFTempoChangesDoNotMoveNotes(t *testing.T) {
	event := MidiEvent{Messages: []MidiMessage{
		&MidiMessageTempoSetting{Tempo: 600000},
		&MidiMessageDelay{Delay: 30},
		&MidiMessageNoteOn{Channel: 0, Note: 60, Velocity: 100, Duration: 120},
		&MidiMessageDelay{Delay: 30},
		&MidiMessageTempoSetting{Tempo: 1000000},
		&MidiMessageNoteOn{Channel: 0, Note: 62, Velocity: 100, Duration: 30},
		&MidiMessageDelay{Delay: 10},
		&MidiMessageTempoSetting{Tempo: 250000},
		&MidiMessageDelay{Delay: 20},
		&MidiMessageEndOfTrack{},
	}}
	song := event.ConvertToSMF()
	want := smf.Track{
		{Delta: 0, Message: smf.MetaTempo(120).Bytes()},
		{Delta: 30, Message: midi.NoteOn(0, 60, 100).Bytes()},
		{Delta: 30, Message: midi.NoteOn(0, 62, 100).Bytes()},
		{Delta: 30, Message: midi.NoteOffVelocity(0, 62, 127).Bytes()},
		{Delta: 60, Message: midi.NoteOffVelocity(0, 60, 127).Bytes()},
		{Delta: 0, Message: smf.EOT},
	}
	if !reflect.DeepEqual(song.Tracks[0], want) {
		t.Fatalf("track = %v, want %v", song.Tracks[0], want)
	}
	if got := midiLength(t, song); got != 1250*time.Millisecond {
		t.Fatalf("duration = %v, want 1.25s", got)
	}
}

func TestConvertToSMFSmallDelaysDoNotDrift(t *testing.T) {
	event := MidiEvent{Messages: []MidiMessage{
		&MidiMessageTempoSetting{Tempo: 1500000},
		&MidiMessageNoteOn{Channel: 0, Note: 60, Velocity: 100, Duration: 120},
	}}
	for range 120 {
		event.Messages = append(event.Messages,
			&MidiMessageDelay{Delay: 1},
			&MidiMessageProgramChange{Channel: 0, Program: 5},
		)
	}
	event.Messages = append(event.Messages, &MidiMessageEndOfTrack{})
	song := event.ConvertToSMF()
	// MeltySynth truncates each event's timestamp to nanoseconds; allow that
	// sub-microsecond loss, but verify the MIDI ticks themselves exactly below.
	if got := midiLength(t, song); got < time.Second-time.Microsecond || got > time.Second+time.Microsecond {
		t.Fatalf("120 one-tick delays played for %v, want approximately 1s", got)
	}
	var tick uint32
	for _, event := range song.Tracks[0] {
		tick += event.Delta
		if reflect.DeepEqual(event.Message.Bytes(), midi.NoteOffVelocity(0, 60, 127).Bytes()) && tick != 120 {
			t.Fatalf("note off at tick %d, want 120", tick)
		}
	}
	if tick != 120 {
		t.Fatalf("track ended at tick %d, want 120", tick)
	}
}

func testIFFChunk(id string, data []byte) []byte {
	chunk := append([]byte(id), 0, 0, 0, 0)
	binary.BigEndian.PutUint32(chunk[4:], uint32(len(data)))
	chunk = append(chunk, data...)
	if len(data)%2 != 0 {
		chunk = append(chunk, 0)
	}
	return chunk
}

func testXMID(children []byte) []byte {
	form := testIFFChunk("FORM", append([]byte("XMID"), children...))
	return testIFFChunk("CAT ", append([]byte("XMID"), form...))
}

func TestIFFReaderShortReads(t *testing.T) {
	data := append(testIFFChunk("EVNT", []byte{1, 2, 3}), testIFFChunk("INFO", []byte{1, 0})...)
	for _, test := range []struct {
		name   string
		reader io.Reader
	}{
		{"one byte", iotest.OneByteReader(bytes.NewReader(data))},
		{"half reads", iotest.HalfReader(bytes.NewReader(data))},
		{"data with EOF", iotest.DataErrReader(bytes.NewReader(data))},
	} {
		t.Run(test.name, func(t *testing.T) {
			reader := NewIFFReader(test.reader)
			for _, want := range []IFFChunk{
				{ID: []byte("EVNT"), Data: []byte{1, 2, 3}},
				{ID: []byte("INFO"), Data: []byte{1, 0}},
			} {
				got, err := reader.ReadChunk()
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("chunk = %v, want %v", got, want)
				}
			}
			if _, err := reader.ReadChunk(); err != io.EOF {
				t.Fatalf("end of stream = %v, want EOF", err)
			}
		})
	}
}

func TestIFFReaderTruncatedChunks(t *testing.T) {
	complete := testIFFChunk("DATA", []byte{1, 2})
	for _, size := range []int{1, 3, 4, 6, 8, 9} {
		reader := NewIFFReader(bytes.NewReader(complete[:size]))
		if !reader.HasMore() {
			t.Fatalf("HasMore hid %d trailing bytes", size)
		}
		if _, err := reader.ReadChunk(); !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Fatalf("%d-byte prefix: error = %v, want unexpected EOF", size, err)
		}
	}
	odd := testIFFChunk("DATA", []byte{1})
	reader := NewIFFReader(bytes.NewReader(odd[:len(odd)-1]))
	if _, err := reader.ReadChunk(); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("missing padding: error = %v, want unexpected EOF", err)
	}
}

func TestReadEventMetaPayloadAcrossBufferBoundary(t *testing.T) {
	// The meta header ends at byte 4095, leaving only one of its four payload
	// bytes in bufio's buffer. A single Read would incorrectly return short.
	data := bytes.Repeat([]byte{0xc0, 5}, 2046)
	data = append(data, 0xff, 0x58, 4, 4, 2, 24, 8, 0xff, 0x2f, 0)
	event, err := (&IFFChunk{Data: data}).ReadEvent()
	if err != nil {
		t.Fatal(err)
	}
	want := &MidiMessageTimeSignature{Numerator: 4, Denominator: 2, Metronome: 24, DemiSemiQuaverPerQuarter: 8}
	if len(event.Messages) != 2048 || !reflect.DeepEqual(event.Messages[2046], want) {
		t.Fatalf("incorrect event count or time signature: %d messages", len(event.Messages))
	}
}

func TestReadEventRejectsTruncation(t *testing.T) {
	for _, data := range [][]byte{
		nil,
		{0xc0, 5},             // Complete event, but no EOT.
		{10},                  // Delay at EOF.
		{0x90, 60, 100},       // Missing duration.
		{0xff, 0x58, 4, 4, 2}, // Short meta payload.
		{0xff, 0x2f},          // Missing EOT length.
		{0xff, 0x2f, 1, 0},    // Invalid EOT payload.
	} {
		if _, err := (&IFFChunk{Data: data}).ReadEvent(); err == nil {
			t.Fatalf("accepted truncated or invalid event stream: % x", data)
		}
	}
	if _, err := (&IFFChunk{Data: []byte{0xc0, 5}}).ReadEvent(); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("missing EOT: error = %v, want unexpected EOF", err)
	}
	// A valid EOT ends the event stream; padding must not become more events.
	event, err := (&IFFChunk{Data: []byte{0xff, 0x2f, 0, 0}}).ReadEvent()
	if err != nil || len(event.Messages) != 1 {
		t.Fatalf("EOT with padding: event = %v, error = %v", event, err)
	}
}

func TestConvertToMidiPropagatesReadErrors(t *testing.T) {
	valid := testXMID(testIFFChunk("EVNT", []byte{0xff, 0x2f, 0}))
	for _, nested := range [][]byte{
		[]byte("E"),                            // Partial chunk ID.
		[]byte("EVNT"),                         // Missing size.
		{'E', 'V', 'N', 'T', 0, 0, 0, 4, 0xff}, // Truncated payload.
	} {
		// A later valid sequence must not hide an earlier broken subchunk.
		data := append(testXMID(nested), valid...)
		if _, err := ConvertToMidi(bytes.NewReader(data)); !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Fatalf("nested % x: error = %v, want unexpected EOF", nested, err)
		}
	}
	children := append(testIFFChunk("TIMB", []byte{1, 0}), testIFFChunk("EVNT", []byte{0xff, 0x2f, 0})...)
	if _, err := ConvertToMidi(bytes.NewReader(testXMID(children))); err == nil || !strings.Contains(err.Error(), "TIMB") {
		t.Fatalf("broken timbre: error = %v, want TIMB error", err)
	}
	for _, data := range [][]byte{[]byte("F"), []byte("FORM")} {
		if _, err := ConvertToMidi(bytes.NewReader(data)); !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Fatalf("root % x: error = %v, want unexpected EOF", data, err)
		}
	}
	failure := errors.New("test read failure")
	reader := io.MultiReader(bytes.NewReader(testIFFChunk("JUNK", nil)), iotest.ErrReader(failure))
	if _, err := ConvertToMidi(reader); !errors.Is(err, failure) {
		t.Fatalf("I/O error = %v, want original read failure", err)
	}
	if _, err := ConvertToMidi(iotest.OneByteReader(bytes.NewReader(valid))); err != nil {
		t.Fatalf("valid file with short reads: %v", err)
	}
}
