package music

import (
	"testing"

	"github.com/kazzmir/master-of-magic/game/magic/audio"
	"github.com/kazzmir/master-of-magic/lib/meltysynth"
)

func TestMusicPolyphonyDefaultAndOverrides(t *testing.T) {
	for _, test := range []struct {
		name    string
		options []PlayOptions
		want    int32
	}{
		{"default", nil, 256},
		{"zero means default", []PlayOptions{{MaximumPolyphony: 0}}, 256},
		{"minimum", []PlayOptions{{MaximumPolyphony: meltysynth.MinimumSupportedPolyphony}}, 8},
		{"override", []PlayOptions{{MaximumPolyphony: 128}}, 128},
		{"progress-only option preserves budget", []PlayOptions{{MaximumPolyphony: 128}, {}}, 128},
		{"last explicit option wins", []PlayOptions{{MaximumPolyphony: 128}, {MaximumPolyphony: 256}}, 256},
	} {
		t.Run(test.name, func(t *testing.T) {
			settings, err := synthesizerSettings(test.options)
			if err != nil {
				t.Fatal(err)
			}
			if settings.MaximumPolyphony != test.want || settings.SampleRate != audio.SampleRate {
				t.Fatalf("settings = %+v, want %d voices at %d Hz", settings, test.want, audio.SampleRate)
			}
			synth, err := meltysynth.NewSynthesizer(&meltysynth.SoundFont{}, settings)
			if err != nil {
				t.Fatal(err)
			}
			if synth.MaximumPolyphony != test.want {
				t.Fatalf("synth voice budget = %d, want %d", synth.MaximumPolyphony, test.want)
			}
		})
	}
}

func TestMusicPolyphonyRejectsInvalidBudgets(t *testing.T) {
	for _, budget := range []int32{-1, 1, 7, 257, 1024} {
		if _, err := synthesizerSettings([]PlayOptions{{MaximumPolyphony: budget}}); err == nil {
			t.Fatalf("accepted invalid voice budget %d", budget)
		}
	}
}
