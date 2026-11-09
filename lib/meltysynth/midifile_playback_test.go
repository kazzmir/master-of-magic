package meltysynth

import (
	"reflect"
	"testing"
	"time"
)

func loopTestFile(count byte) *MidiFile {
	return &MidiFile{
		messages: []message{
			common2b(0xc0, 1), // Intro.
			common3b(0xb0, 116, count),
			common2b(0xc0, 2), // Loop body.
			common3b(0xb0, 117, 127),
			common2b(0xc0, 9), // Tail.
			endOfTrack(),
		},
		times: []time.Duration{0, time.Second, time.Second, 3 * time.Second, 4 * time.Second, 5 * time.Second},
	}
}

func advancePlayback(p *MidiFilePlayback, elapsed time.Duration, programs *[]byte) {
	p.position += elapsed - p.elapsed
	p.elapsed = elapsed
	p.processEvents(func(msg message) {
		if msg.command == 0xc0 {
			*programs = append(*programs, msg.data1)
		}
	}, nil)
}

func TestXMIDIInfiniteLoopSkipsIntroAndTail(t *testing.T) {
	p := NewXMIDIPlayback(loopTestFile(0), true)
	var programs []byte
	advancePlayback(p, 0, &programs)
	advancePlayback(p, 3*time.Second, &programs)
	advancePlayback(p, 5*time.Second, &programs)
	if want := []byte{1, 2, 2, 2}; !reflect.DeepEqual(programs, want) {
		t.Fatalf("programs = %v, want %v", programs, want)
	}
	if p.Position() != time.Second || p.Duration() != 3*time.Second || p.Loops() != 2 {
		t.Fatalf("position=%v duration=%v loops=%d", p.Position(), p.Duration(), p.Loops())
	}
	// Overshoot within an audio block must survive a jump rather than adding
	// a block's worth of delay to every loop.
	p.AdvanceTo(7*time.Second + 500*time.Microsecond)
	if p.Position() != time.Second+500*time.Microsecond || p.Loops() != 3 {
		t.Fatalf("overshoot lost: position=%v loops=%d", p.Position(), p.Loops())
	}
}

func TestXMIDIFiniteLoopCount(t *testing.T) {
	for _, count := range []byte{1, 3, 127} {
		p := NewXMIDIPlayback(loopTestFile(count), false)
		var programs []byte
		end := 5*time.Second + time.Duration(count-1)*2*time.Second
		advancePlayback(p, end, &programs)
		want := []byte{1}
		for range int(count) {
			want = append(want, 2)
		}
		want = append(want, 9)
		if !reflect.DeepEqual(programs, want) {
			t.Fatalf("count=%d programs=%v, want %v", count, programs, want)
		}
		if p.Position() != 5*time.Second || p.Loops() != int64(count-1) {
			t.Fatalf("count=%d position=%v loops=%d", count, p.Position(), p.Loops())
		}
	}
}

func TestXMIDINestedLoops(t *testing.T) {
	file := &MidiFile{
		messages: []message{
			common3b(0xb0, 116, 2),
			common2b(0xc0, 1),
			common3b(0xb0, 116, 3),
			common2b(0xc0, 2),
			common3b(0xb0, 117, 127),
			common3b(0xb0, 117, 127),
			common2b(0xc0, 9),
			endOfTrack(),
		},
		times: []time.Duration{0, 0, time.Second, time.Second, 2 * time.Second, 3 * time.Second, 4 * time.Second, 5 * time.Second},
	}
	p := NewXMIDIPlayback(file, false)
	var programs []byte
	advancePlayback(p, 12*time.Second, &programs)
	want := []byte{1, 2, 2, 2, 1, 2, 2, 2, 9}
	if !reflect.DeepEqual(programs, want) || p.Loops() != 5 || p.Position() != 5*time.Second {
		t.Fatalf("programs=%v loops=%d position=%v", programs, p.Loops(), p.Position())
	}
}

func TestXMIDIBreakAndDisabledInfiniteLoop(t *testing.T) {
	for _, test := range []struct {
		name            string
		loop, breakLoop bool
	}{
		{"infinite disabled", false, false},
		{"break", true, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			file := loopTestFile(0)
			if test.breakLoop {
				file.messages[3].data2 = 0
			}
			p := NewXMIDIPlayback(file, test.loop)
			var programs []byte
			advancePlayback(p, 4*time.Second, &programs)
			if !reflect.DeepEqual(programs, []byte{1, 2, 9}) || p.Loops() != 0 {
				t.Fatalf("programs=%v loops=%d", programs, p.Loops())
			}
		})
	}
}

func TestXMIDIZeroLengthLoopDoesNotSpin(t *testing.T) {
	file := loopTestFile(0)
	file.times[3] = time.Second
	p := NewXMIDIPlayback(file, true)
	p.AdvanceTo(4 * time.Second)
	if p.Position() != 4*time.Second || p.Loops() != 0 {
		t.Fatalf("position=%v loops=%d", p.Position(), p.Loops())
	}
}

func TestOrdinaryMIDIIgnoresXMIDIControls(t *testing.T) {
	p := newMidiFilePlayback(loopTestFile(0), true, false)
	var programs []byte
	advancePlayback(p, 5*time.Second, &programs)
	if !reflect.DeepEqual(programs, []byte{1, 2, 9, 1}) || p.Loops() != 1 || p.Position() != 0 {
		t.Fatalf("programs=%v loops=%d position=%v", programs, p.Loops(), p.Position())
	}
}

func TestSequencerXMIDIRenderMatchesProgressClock(t *testing.T) {
	// Program-change-only playback needs no soundfont or audio device.
	synth := &Synthesizer{
		BlockSize:  10,
		SampleRate: 1000,
		voices:     &voiceCollection{},
		blockLeft:  make([]float32, 10),
		blockRight: make([]float32, 10),
	}
	synth.channels = []*channel{newChannel(synth, false)}
	seq := NewMidiFileSequencer(synth)
	file := loopTestFile(0)
	seq.PlayXMI(file, true)
	left, right := make([]float32, 10), make([]float32, 10)
	for range 501 {
		seq.Render(left, right)
	}
	tracker := NewXMIDIPlayback(file, true)
	tracker.AdvanceTo(5010 * time.Millisecond)
	if seq.playback.Loops() != tracker.Loops() || seq.playback.Position() != tracker.Position() || synth.channels[0].patchNumber != 2 {
		t.Fatalf("rendered position=%v loops=%d patch=%d; tracker position=%v loops=%d", seq.playback.Position(), seq.playback.Loops(), synth.channels[0].patchNumber, tracker.Position(), tracker.Loops())
	}
	seq.Play(file, true)
	for range 501 {
		seq.Render(left, right)
	}
	if synth.channels[0].patchNumber != 1 || seq.playback.Loops() != 1 {
		t.Fatalf("ordinary MIDI playback changed: patch=%d loops=%d", synth.channels[0].patchNumber, seq.playback.Loops())
	}
}

func TestXMIDIUnmatchedStartFallsBackToWholeFileLoop(t *testing.T) {
	file := loopTestFile(0)
	file.messages[3] = common2b(0xc0, 3)
	p := NewXMIDIPlayback(file, true)
	p.AdvanceTo(15 * time.Second)
	if p.Loops() != 3 || p.Position() != 0 || len(p.stack) != 0 {
		t.Fatalf("position=%v loops=%d stack=%v", p.Position(), p.Loops(), p.stack)
	}
}

func sustainTestSynth() *Synthesizer {
	synth := &Synthesizer{
		BlockSize:            64,
		SampleRate:           44100,
		minimumVoiceDuration: 88,
		voices:               &voiceCollection{},
	}
	for range 16 {
		synth.channels = append(synth.channels, newChannel(synth, false))
	}
	return synth
}

func sustainedTestVoice(synth *Synthesizer, channel int32) *voice {
	v := newVoice(synth)
	v.channel = channel
	v.voiceLength = synth.minimumVoiceDuration
	v.noteGain = 0.5
	v.end()
	return v
}

func TestLoopRestartReleasesSustainedVoices(t *testing.T) {
	for _, test := range []struct {
		name            string
		xmi             bool
		reenableAtStart bool
	}{
		{"whole song", false, false},
		{"section", true, false},
		{"section immediately reenables pedal", true, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			file := &MidiFile{
				messages: []message{common3b(0xb5, 64, 127), endOfTrack()},
				times:    []time.Duration{time.Second, 2 * time.Second},
			}
			if test.xmi {
				pedalTime := time.Second
				if test.reenableAtStart {
					pedalTime = 0
				}
				file = &MidiFile{
					messages: []message{common3b(0xb0, 116, 0), common3b(0xb5, 64, 127), common3b(0xb0, 117, 127), endOfTrack()},
					times:    []time.Duration{0, pedalTime, 2 * time.Second, 3 * time.Second},
				}
			}
			synth := sustainTestSynth()
			seq := NewMidiFileSequencer(synth)
			seq.play(file, true, test.xmi)
			seq.playback.position = time.Second
			seq.processEvents()
			if !synth.channels[5].holdPedal {
				t.Fatal("test failed to enable sustain")
			}
			old := sustainedTestVoice(synth, 5)
			synth.voices.voices = []*voice{old}
			synth.voices.activeVoiceCount = 1
			synth.channels[5].volume = 73 << 7
			synth.channels[5].pitchBend = 0.25
			seq.playback.position = 2 * time.Second
			seq.processEvents()
			if synth.channels[5].holdPedal != test.reenableAtStart {
				t.Fatalf("incorrect pedal state after restart: %v", synth.channels[5].holdPedal)
			}
			old.releaseIfNecessary(synth.channels[5])
			if old.voiceState != voice_Released || old.volEnv.stage != env_Release || old.modEnv.stage != env_Release || old.noteGain != 0.5 {
				t.Fatalf("old voice did not enter its normal release envelope: state=%d gain=%v", old.voiceState, old.noteGain)
			}
			if synth.channels[5].volume != 73<<7 || synth.channels[5].pitchBend != 0.25 {
				t.Fatal("loop restart reset unrelated controllers")
			}
			if test.reenableAtStart {
				fresh := sustainedTestVoice(synth, 5)
				fresh.releaseIfNecessary(synth.channels[5])
				if fresh.voiceState != voice_ReleaseRequested {
					t.Fatal("new iteration's voices no longer honor sustain")
				}
			}
		})
	}
}

func TestSectionLoopRestoresSustainAtStart(t *testing.T) {
	file := &MidiFile{
		messages: []message{
			common3b(0xb6, 64, 127),
			common3b(0xb0, 116, 0),
			common3b(0xb6, 64, 0),
			common3b(0xb0, 117, 127),
			endOfTrack(),
		},
		times: []time.Duration{0, time.Second, 2 * time.Second, 3 * time.Second, 4 * time.Second},
	}
	synth := sustainTestSynth()
	seq := NewMidiFileSequencer(synth)
	seq.PlayXMI(file, true)
	seq.playback.position = 2 * time.Second
	seq.processEvents()
	if synth.channels[6].holdPedal {
		t.Fatal("pedal should be off at the end of the loop body")
	}
	old := sustainedTestVoice(synth, 6)
	synth.voices.voices = []*voice{old}
	synth.voices.activeVoiceCount = 1
	seq.playback.position = 3 * time.Second
	seq.processEvents()
	if !synth.channels[6].holdPedal {
		t.Fatal("loop-start pedal state was not restored")
	}
	old.releaseIfNecessary(synth.channels[6])
	if old.voiceState != voice_Released {
		t.Fatal("restored sustain prevented the old voice from releasing")
	}
}

func TestLoopReleaseStillHonorsMinimumVoiceDuration(t *testing.T) {
	synth := sustainTestSynth()
	v := sustainedTestVoice(synth, 0)
	v.voiceLength = 0
	synth.voices.voices = []*voice{v}
	synth.voices.activeVoiceCount = 1
	sustain := [16]bool{true}
	synth.restartLoop(sustain)
	v.releaseIfNecessary(synth.channels[0])
	if v.voiceState != voice_ReleaseRequested {
		t.Fatal("loop restart bypassed minimum voice duration")
	}
	v.voiceLength = synth.minimumVoiceDuration
	v.releaseIfNecessary(synth.channels[0])
	if v.voiceState != voice_Released {
		t.Fatal("old voice stayed sustained after minimum duration")
	}
}
