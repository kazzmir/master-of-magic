package music

import (
	"math"

	"github.com/kazzmir/master-of-magic/lib/meltysynth"
)

// The general-purpose synth defaults to 0.5. Music reserves roughly 3 dB more
// headroom so the bundled soundfont's dense passages need no peak limiting.
const musicSynthesizerGain float32 = 0.35
const musicSoftLimitKnee = 0.9

func newMusicSynthesizer(font *meltysynth.SoundFont, options []PlayOptions) (*meltysynth.Synthesizer, error) {
	settings, err := synthesizerSettings(options)
	if err != nil {
		return nil, err
	}
	synth, err := meltysynth.NewSynthesizer(font, settings)
	if err != nil {
		return nil, err
	}
	synth.MasterVolume = musicSynthesizerGain
	return synth, nil
}

// limitStereo keeps float32 output in [-1, 1] with a smooth knee rather than
// hard clipping. Quieter frames are unchanged. Link the gain of both channels
// to the louder sample so overload protection does not alter stereo balance.
func limitStereo(left, right float32) (float32, float32) {
	l, r := float64(left), float64(right)
	peak := math.Max(math.Abs(l), math.Abs(r))
	if peak <= musicSoftLimitKnee {
		return left, right
	}
	if math.IsNaN(peak) || math.IsInf(peak, 0) {
		return 0, 0
	}
	const headroom = 1 - musicSoftLimitKnee
	limited := musicSoftLimitKnee + headroom*math.Tanh((peak-musicSoftLimitKnee)/headroom)
	gain := limited / peak
	return float32(l * gain), float32(r * gain)
}
