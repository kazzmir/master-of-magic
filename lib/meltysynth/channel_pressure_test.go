package meltysynth

import (
	"bytes"
	"math"
	"reflect"
	"testing"
	"time"
)

func TestChannelPressureMessage(t *testing.T) {
	synth := sustainTestSynth()
	synth.ProcessMidiMessage(5, 0xd0, 64, 127)
	if synth.channels[5].pressure != 64 || synth.channels[0].pressure != 0 {
		t.Fatal("channel pressure used the wrong data byte or affected another channel")
	}
	volume, expression := synth.channels[5].volume, synth.channels[5].expression
	for _, value := range []int32{0, 32, 127, -1, 128} {
		synth.ProcessMidiMessage(5, 0xd0, value, 0)
		if got, want := synth.channels[5].pressure, max(0, min(value, 127)); got != want {
			t.Fatalf("pressure = %d, want %d", got, want)
		}
	}
	if synth.channels[5].volume != volume || synth.channels[5].expression != expression {
		t.Fatal("channel pressure must not replace volume or expression")
	}
	synth.ProcessMidiMessage(-1, 0xd0, 127, 0)
	synth.ProcessMidiMessage(16, 0xd0, 127, 0)
}

func TestChannelPressureDefaultVibrato(t *testing.T) {
	ch := newChannel(nil, false)
	for _, pressure := range []int32{0, 32, 64, 127} {
		ch.setPressure(pressure)
		want := 0.5 * float64(pressure) / 127
		if got := ch.getVibratoDepth(); math.Abs(float64(got)-want) > 1e-6 {
			t.Fatalf("pressure %d: vibrato depth = %v semitones, want %v", pressure, got, want)
		}
	}
	ch.setModulationCoarse(127)
	ch.setModulationFine(127)
	if got := ch.getVibratoDepth(); math.Abs(float64(got)-1) > 1e-6 {
		t.Fatalf("pressure and modulation should add: depth = %v, want 1 semitone", got)
	}
	ch.setPressure(0)
	if got := ch.getVibratoDepth(); math.Abs(float64(got)-0.5) > 1e-6 {
		t.Fatalf("zero pressure must preserve the modulation wheel: depth = %v", got)
	}
}

func TestChannelPressureResets(t *testing.T) {
	synth := sustainTestSynth()
	for _, reset := range []func(){
		func() { synth.ProcessMidiMessage(5, 0xb0, 121, 0) },
		synth.ResetAllControllers,
		synth.Reset,
	} {
		synth.ProcessMidiMessage(5, 0xd0, 127, 0)
		reset()
		if synth.channels[5].pressure != 0 || synth.channels[5].getVibratoDepth() != 0 {
			t.Fatal("reset left channel pressure enabled")
		}
	}
}

func TestChannelPressureMIDIRunningStatus(t *testing.T) {
	// Pressure has one data byte. The second D0 event uses running status.
	data := []byte("MThd\x00\x00\x00\x06\x00\x00\x00\x01\x00\x3cMTrk\x00\x00\x00\x0c")
	data = append(data, 0, 0xd0, 16, 60, 127, 60, 0xd1, 64, 0, 0xff, 0x2f, 0)
	file, err := NewMidiFile(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	synth := sustainTestSynth()
	seq := NewMidiFileSequencer(synth)
	seq.PlayXMI(file, false)
	for _, test := range []struct {
		time               time.Duration
		channel0, channel1 int32
	}{
		{0, 16, 0},
		{500 * time.Millisecond, 127, 0},
		{time.Second, 127, 64},
	} {
		seq.playback.position = test.time
		seq.processEvents()
		if synth.channels[0].pressure != test.channel0 || synth.channels[1].pressure != test.channel1 {
			t.Fatalf("at %v: pressure = %d/%d, want %d/%d", test.time, synth.channels[0].pressure, synth.channels[1].pressure, test.channel0, test.channel1)
		}
	}
}

func TestChannelPressureLoopState(t *testing.T) {
	for _, test := range []struct {
		name                  string
		xmi, resetBeforeStart bool
		want                  int32
	}{
		{"whole song", false, false, 0},
		{"section restores intro pressure", true, false, 25},
		{"controller reset before loop start", true, true, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			file := &MidiFile{
				messages: []message{common2b(0xd5, 100), endOfTrack()},
				times:    []time.Duration{time.Second, 3 * time.Second},
			}
			if test.xmi {
				file = &MidiFile{messages: []message{common2b(0xd5, 25)}, times: []time.Duration{0}}
				if test.resetBeforeStart {
					file.messages = append(file.messages, common3b(0xb5, 121, 0))
					file.times = append(file.times, 500*time.Millisecond)
				}
				file.messages = append(file.messages, common3b(0xb0, 116, 0), common2b(0xd5, 100), common3b(0xb0, 117, 127), endOfTrack())
				file.times = append(file.times, time.Second, 2*time.Second, 3*time.Second, 4*time.Second)
			}
			synth := sustainTestSynth()
			seq := NewMidiFileSequencer(synth)
			seq.play(file, true, test.xmi)
			seq.playback.position = 2 * time.Second
			seq.processEvents()
			if synth.channels[5].pressure != 100 {
				t.Fatal("pressure not set before restart")
			}
			seq.playback.position = 3 * time.Second
			seq.processEvents()
			if got := synth.channels[5].pressure; got != test.want {
				t.Fatalf("pressure after restart = %d, want %d", got, test.want)
			}
		})
	}
}

func TestChannelPressureChangesActiveVoiceAudio(t *testing.T) {
	font := loadGM(t)
	settings := NewSynthesizerSettings(44100)
	settings.EnableReverbAndChorus = false
	plain, err := NewSynthesizer(font, settings)
	if err != nil {
		t.Fatal(err)
	}
	pressure, err := NewSynthesizer(font, settings)
	if err != nil {
		t.Fatal(err)
	}
	leftA, rightA := make([]float32, 11025), make([]float32, 11025)
	leftB, rightB := make([]float32, 11025), make([]float32, 11025)
	plain.NoteOn(0, 60, 100)
	pressure.NoteOn(0, 60, 100)
	plain.Render(leftA, rightA)
	pressure.Render(leftB, rightB)
	if !reflect.DeepEqual(leftA, leftB) {
		t.Fatal("identical initial voices rendered differently")
	}
	pressure.ProcessMidiMessage(0, 0xd0, 127, 0)
	plain.Render(leftA, rightA)
	pressure.Render(leftB, rightB)
	if reflect.DeepEqual(leftA, leftB) {
		t.Fatal("channel pressure did not affect the already sounding voice")
	}
}
