package meltysynth

import (
	"reflect"
	"testing"
	"time"
)

func crossingLoopNotes() *MidiFile {
	return &MidiFile{
		messages: []message{
			common3b(0x90, 60, 100), // Intro note lasts three seconds.
			common3b(0xb0, 116, 0),
			common3b(0x90, 60, 90),   // Each section note lasts 1.5 seconds.
			common3b(0xb0, 117, 127), // Section is only one second long.
			common3b(0x80, 60, 127),
			common3b(0x80, 60, 127),
			endOfTrack(),
		},
		times: []time.Duration{0, time.Second, time.Second, 2 * time.Second, 2500 * time.Millisecond, 3 * time.Second, 3 * time.Second},
	}
}

func TestXMIDINoteOffDeadlinesSurviveSectionJumps(t *testing.T) {
	p := NewXMIDIPlayback(crossingLoopNotes(), true)
	var offs []uint64
	var ons []uint64
	for _, test := range []struct {
		at                time.Duration
		wantOns, wantOffs []uint64
	}{
		{0, []uint64{1}, nil},
		{time.Second, []uint64{1, 2}, nil},
		{2 * time.Second, []uint64{1, 2, 3}, nil},
		{2500 * time.Millisecond, []uint64{1, 2, 3}, []uint64{2}},
		{3 * time.Second, []uint64{1, 2, 3, 4}, []uint64{2, 1}},
		{3500 * time.Millisecond, []uint64{1, 2, 3, 4}, []uint64{2, 1, 3}},
	} {
		p.position += test.at - p.elapsed
		p.elapsed = test.at
		p.processEvents(func(msg message) {
			if msg.command == 0x90 && msg.data2 != 0 {
				ons = append(ons, msg.noteID)
			} else if msg.command == 0x80 {
				offs = append(offs, msg.noteID)
			}
		}, nil)
		if !reflect.DeepEqual(ons, test.wantOns) || !reflect.DeepEqual(offs, test.wantOffs) {
			t.Fatalf("at %v: note-ons=%v note-offs=%v, want %v / %v", test.at, ons, offs, test.wantOns, test.wantOffs)
		}
	}
}

func TestXMIDIScheduledOffPreservesSameTickOrdering(t *testing.T) {
	file := &MidiFile{
		messages: []message{
			common3b(0x90, 60, 100),
			common2b(0xc0, 5),
			common3b(0x90, 62, 100),
			common3b(0x80, 62, 127),
			common3b(0x80, 60, 127),
			endOfTrack(),
		},
		times: []time.Duration{0, 0, 0, 0, 0, time.Second},
	}
	p := NewXMIDIPlayback(file, false)
	var got []message
	p.processEvents(func(msg message) { msg.noteID = 0; got = append(got, msg) }, nil)
	if want := file.messages[:5]; !reflect.DeepEqual(got, want) {
		t.Fatalf("events = %v, want %v", got, want)
	}
}

func TestXMIDIDurationDoesNotExpandWithNestedLoops(t *testing.T) {
	file := &MidiFile{
		messages: []message{
			common3b(0x90, 60, 100),
			common3b(0xb0, 116, 2),
			common3b(0xb0, 116, 2),
			common3b(0xb0, 117, 127),
			common3b(0xb0, 117, 127),
			common3b(0x80, 60, 127),
			endOfTrack(),
		},
		times: []time.Duration{0, time.Second, 2 * time.Second, 3 * time.Second, 4 * time.Second, 6 * time.Second, 6 * time.Second},
	}
	p := NewXMIDIPlayback(file, false)
	offs := 0
	emit := func(msg message) {
		if msg.command == 0x80 {
			offs++
		}
	}
	p.position, p.elapsed = 6*time.Second, 6*time.Second
	p.processEvents(emit, nil)
	if offs != 1 {
		t.Fatalf("six-second note should end at six elapsed seconds, got %d offs", offs)
	}
	p.position += 5 * time.Second
	p.elapsed += 5 * time.Second
	p.processEvents(emit, nil)
	if offs != 1 {
		t.Fatalf("source note-off was delivered a second time after loops: %d", offs)
	}
}

func TestSectionJumpDoesNotReleaseUnexpiredVoices(t *testing.T) {
	synth := sustainTestSynth()
	old := newVoice(synth)
	old.channel, old.key, old.noteID = 5, 60, 1
	old.voiceLength = synth.minimumVoiceDuration
	synth.voices.voices = []*voice{old}
	synth.voices.activeVoiceCount = 1
	state := midiLoopState{sustain: [16]bool{5: true}}
	synth.restartLoop(state, false)
	if old.voiceState != voice_Playing || !old.ignoreHoldPedal {
		t.Fatal("section jump ended an unexpired note or failed to prevent stuck sustain")
	}
	fresh := newVoice(synth)
	fresh.channel, fresh.key, fresh.noteID = 5, 60, 2
	synth.voices.voices = append(synth.voices.voices, fresh)
	synth.voices.activeVoiceCount++
	synth.noteOffInstance(1)
	old.releaseIfNecessary(synth.channels[5])
	if old.voiceState != voice_Released || fresh.voiceState != voice_Playing {
		t.Fatal("old scheduled note-off did not release independently of the new same-key note")
	}
}

func TestWholeSongRestartStillEndsUnexpiredVoices(t *testing.T) {
	synth := sustainTestSynth()
	v := newVoice(synth)
	synth.voices.voices = []*voice{v}
	synth.voices.activeVoiceCount = 1
	synth.restartLoop(midiLoopState{}, true)
	if v.voiceState != voice_ReleaseRequested {
		t.Fatal("whole-song restart no longer cleans up unexpired notes")
	}
}

func TestXMIDINoteIDsAreAssignedToAllVoiceLayers(t *testing.T) {
	font := loadGM(t)
	synth, err := NewSynthesizer(font, NewSynthesizerSettings(44100))
	if err != nil {
		t.Fatal(err)
	}
	synth.noteOn(0, 60, 100, 1)
	synth.noteOn(0, 60, 100, 2)
	synth.noteOffInstance(1)
	first, second := 0, 0
	for _, v := range synth.voices.voices[:synth.voices.activeVoiceCount] {
		switch v.noteID {
		case 1:
			first++
			if v.voiceState != voice_ReleaseRequested {
				t.Fatal("old layer missed its note-off")
			}
		case 2:
			second++
			if v.voiceState != voice_Playing {
				t.Fatal("new layer ended with the old note")
			}
		default:
			t.Fatal("a voice layer was not assigned a note ID")
		}
	}
	if first == 0 || second == 0 {
		t.Fatal("test did not create both note instances")
	}
}
