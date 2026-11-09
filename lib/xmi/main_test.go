package xmi

import (
	"reflect"
	"testing"

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
			name:     "tempo scaling also applies to final note off",
			messages: []MidiMessage{&MidiMessageTempoSetting{Tempo: 375000}, note(60, 10)},
			want:     smf.Track{{Delta: 0, Message: smf.MetaTempo(160).Bytes()}, on(0, 60), off(20, 60), eot(0)},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event := MidiEvent{Messages: append(test.messages, &MidiMessageEndOfTrack{})}
			got := event.ConvertToSMF().Tracks[0]
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("track = %v, want %v", got, test.want)
			}
		})
	}
}
