package midi_test

import (
	"testing"

	"github.com/kazzmir/master-of-magic/lib/midi"
)

func TestSysCommon(t *testing.T) {

	tests := []struct {
		msg      midi.Message
		expected string
	}{
		{
			midi.MTC(3),
			"MTC mtc: 3",
		},
		{
			midi.Tune(),
			"Tune",
		},
		{
			midi.SongSelect(5),
			"SongSelect song: 5",
		},
		{
			midi.SPP(4),
			"SPP spp: 4",
		},
		{
			midi.SPP(4000),
			"SPP spp: 4000",
		},
	}

	for n, test := range tests {
		//m := midi.Message(test.msg)
		m := test.msg

		if got, want := m.String(), test.expected; got != want {
			t.Errorf("[%v] (% X).String() = %#v; want %#v", n, test.msg, got, want)
		}

	}
}
