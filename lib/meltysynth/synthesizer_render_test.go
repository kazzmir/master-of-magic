package meltysynth

import (
	"reflect"
	"testing"
)

func newRenderTestSynth(t *testing.T, effects bool) *Synthesizer {
	t.Helper()
	settings := NewSynthesizerSettings(44100)
	settings.MaximumPolyphony = 8
	settings.EnableReverbAndChorus = effects
	// These tests construct voices directly; no presets or audio device needed.
	synth, err := NewSynthesizer(&SoundFont{}, settings)
	if err != nil {
		t.Fatal(err)
	}
	return synth
}

func expiredRenderVoice(synth *Synthesizer) *voice {
	v := newVoice(synth)
	v.noteGain = 1
	v.volEnv.stage = env_Release
	v.volEnv.releaseLevel = 0 // Expires on the next process call.
	v.previousMixGainLeft, v.currentMixGainLeft = 1, 1
	v.previousMixGainRight, v.currentMixGainRight = 1, 1
	v.previousChorusSend, v.currentChorusSend = 1, 1
	v.previousReverbSend, v.currentReverbSend = 1, 1
	for i := range v.block {
		v.block[i] = 1 // Stale samples from the previous block must not be mixed.
	}
	return v
}

func liveRenderVoice(synth *Synthesizer) *voice {
	v := newVoice(synth)
	v.noteGain = 0.5
	v.key = 60
	v.instrumentReverb, v.instrumentChorus = 0.5, 0.5
	v.volEnv.start(0, 0.001, 0, 1, 1, 0.1)
	v.modEnv.start(0, 0.001, 0, 1, 1, 0.1)
	data := make([]int16, 65)
	for i := range data {
		data[i] = 8192
	}
	v.oscillator.start(data, loop_Continuous, synth.SampleRate, 0, 64, 0, 64, 60, 0, 0, 100)
	v.filter.setLowPassFilter(13500, 0)
	return v
}

func requireSilent(t *testing.T, name string, samples []float32) {
	t.Helper()
	for i, sample := range samples {
		if sample != 0 {
			t.Fatalf("%s[%d] = %v, want silence", name, i, sample)
		}
	}
}

func TestRenderBlockDoesNotMixExpiredVoices(t *testing.T) {
	for _, effects := range []bool{false, true} {
		for _, count := range []int{1, 3} {
			synth := newRenderTestSynth(t, effects)
			for i := 0; i < count; i++ {
				synth.voices.voices[i] = expiredRenderVoice(synth)
			}
			synth.voices.activeVoiceCount = int32(count)
			synth.renderBlock()
			if synth.voices.activeVoiceCount != 0 {
				t.Fatal("expired voices were not removed")
			}
			requireSilent(t, "left", synth.blockLeft)
			requireSilent(t, "right", synth.blockRight)
			if effects {
				requireSilent(t, "chorus left input", synth.chorusInputLeft)
				requireSilent(t, "chorus right input", synth.chorusInputRight)
				requireSilent(t, "reverb input", synth.reverbInput)
			}
		}
	}
}

func TestRenderBlockMixesSurvivorAfterCompaction(t *testing.T) {
	for _, effects := range []bool{false, true} {
		baseline := newRenderTestSynth(t, effects)
		baseline.voices.voices[0] = liveRenderVoice(baseline)
		baseline.voices.activeVoiceCount = 1
		synth := newRenderTestSynth(t, effects)
		for i := range 3 {
			synth.voices.voices[i] = expiredRenderVoice(synth)
		}
		survivor := liveRenderVoice(synth)
		synth.voices.voices[3] = survivor
		synth.voices.activeVoiceCount = 4
		for range 4 {
			baseline.renderBlock()
			synth.renderBlock()
			if synth.voices.activeVoiceCount != 1 || synth.voices.voices[0] != survivor {
				t.Fatal("survivor was not compacted into the active prefix")
			}
			for _, pair := range [][2][]float32{
				{synth.blockLeft, baseline.blockLeft},
				{synth.blockRight, baseline.blockRight},
				{synth.chorusInputLeft, baseline.chorusInputLeft},
				{synth.chorusInputRight, baseline.chorusInputRight},
				{synth.reverbInput, baseline.reverbInput},
			} {
				if !reflect.DeepEqual(pair[0], pair[1]) {
					t.Fatal("expired voices changed the survivor's dry mix or effects input")
				}
			}
		}
		if reflect.DeepEqual(synth.blockLeft, make([]float32, synth.BlockSize)) {
			t.Fatal("test survivor produced no audio")
		}
	}
}

func TestRenderBlockPreservesExistingEffectTails(t *testing.T) {
	for _, effect := range []string{"chorus", "reverb"} {
		t.Run(effect, func(t *testing.T) {
			synth := newRenderTestSynth(t, true)
			if effect == "chorus" {
				synth.chorusInputLeft[0], synth.chorusInputRight[0] = 1, 1
				synth.chorus.process(synth.chorusInputLeft, synth.chorusInputRight, synth.chorusOutputLeft, synth.chorusOutputRight)
			} else {
				synth.reverbInput[0] = 1
				synth.reverb.process(synth.reverbInput, synth.reverbOutputLeft, synth.reverbOutputRight)
			}
			tail := false
			for range 256 {
				synth.renderBlock()
				for i := range synth.blockLeft {
					if synth.blockLeft[i] != 0 || synth.blockRight[i] != 0 {
						tail = true
					}
				}
				requireSilent(t, "chorus left input", synth.chorusInputLeft)
				requireSilent(t, "chorus right input", synth.chorusInputRight)
				requireSilent(t, "reverb input", synth.reverbInput)
			}
			if !tail {
				t.Fatalf("%s tail was lost when no voices remained", effect)
			}
		})
	}
}
