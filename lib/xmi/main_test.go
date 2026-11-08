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
