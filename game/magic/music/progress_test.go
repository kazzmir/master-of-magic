package music

import (
	"bytes"
	"testing"
	"time"

	"github.com/kazzmir/master-of-magic/lib/meltysynth"
	"github.com/kazzmir/master-of-magic/lib/xmi"
)

func testXMIDIPlayback(t *testing.T, count uint8) *meltysynth.MidiFilePlayback {
	t.Helper()
	event := xmi.MidiEvent{Messages: []xmi.MidiMessage{
		&xmi.MidiMessageProgramChange{Program: 1},
		&xmi.MidiMessageDelay{Delay: 120},
		&xmi.MidiMessageControlChange{Controller: 116, Value: count},
		&xmi.MidiMessageProgramChange{Program: 2},
		&xmi.MidiMessageDelay{Delay: 240},
		&xmi.MidiMessageControlChange{Controller: 117, Value: 127},
		&xmi.MidiMessageDelay{Delay: 120},
		&xmi.MidiMessageProgramChange{Program: 9},
		&xmi.MidiMessageDelay{Delay: 120},
		&xmi.MidiMessageEndOfTrack{},
	}}
	var buffer bytes.Buffer
	if _, err := event.ConvertToSMF().WriteTo(&buffer); err != nil {
		t.Fatal(err)
	}
	file, err := meltysynth.NewMidiFile(&buffer)
	if err != nil {
		t.Fatal(err)
	}
	return meltysynth.NewXMIDIPlayback(file, true)
}

func TestPlaybackProgressWithXMIDIIntro(t *testing.T) {
	playback := testXMIDIPlayback(t, 0)
	for _, test := range []struct {
		total, position time.Duration
		loops           int64
	}{
		{0, 0, 0},
		{500 * time.Millisecond, 500 * time.Millisecond, 0},
		{2 * time.Second, 2 * time.Second, 0},
		{3 * time.Second, time.Second, 1},
		{4 * time.Second, 2 * time.Second, 1},
		{7 * time.Second, time.Second, 3},
	} {
		got := playbackProgress(playback, test.total)
		want := PlaybackProgress{Position: test.position, Duration: 3 * time.Second, Total: test.total, Loops: test.loops}
		if got != want {
			t.Fatalf("progress = %+v, want %+v", got, want)
		}
	}
}

func TestPlaybackProgressWithFiniteXMIDILoop(t *testing.T) {
	playback := testXMIDIPlayback(t, 3)
	for _, test := range []struct {
		total, position time.Duration
		loops           int64
	}{
		{3 * time.Second, time.Second, 1},
		{5 * time.Second, time.Second, 2},
		{7 * time.Second, 3 * time.Second, 2},
		{8 * time.Second, 4 * time.Second, 2},
		{9 * time.Second, 0, 3}, // All three iterations plus the tail, then restart.
	} {
		got := playbackProgress(playback, test.total)
		want := PlaybackProgress{Position: test.position, Duration: 5 * time.Second, Total: test.total, Loops: test.loops}
		if got != want {
			t.Fatalf("progress = %+v, want %+v", got, want)
		}
	}
}
