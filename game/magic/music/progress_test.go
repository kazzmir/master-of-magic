package music

import (
	"testing"
	"time"
)

func TestPlaybackProgress(t *testing.T) {
	duration := 10 * time.Second
	period := duration + time.Millisecond
	tests := []struct {
		name     string
		total    time.Duration
		position time.Duration
		loops    int64
	}{
		{"start", 0, 0, 0},
		{"first loop", 3 * time.Second, 3 * time.Second, 0},
		{"block boundary padding", duration + time.Microsecond, duration, 0},
		{"loop reset", period, 0, 1},
		{"several loops", 3*period + 2*time.Second, 2 * time.Second, 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := playbackProgress(test.total, duration, period)
			want := PlaybackProgress{Position: test.position, Duration: duration, Total: test.total, Loops: test.loops}
			if got != want {
				t.Fatalf("progress = %+v, want %+v", got, want)
			}
		})
	}
	if got := playbackProgress(time.Second, 0, 0); got.Position != 0 || got.Loops != 0 {
		t.Fatalf("invalid period should not divide by zero: %+v", got)
	}
}

func TestLoopDuration(t *testing.T) {
	// At 1 kHz with 10-sample blocks, loops align to 10 ms boundaries.
	for _, test := range []struct{ duration, want time.Duration }{
		{0, 10 * time.Millisecond},
		{time.Millisecond, 10 * time.Millisecond},
		{10 * time.Millisecond, 10 * time.Millisecond},
		{11 * time.Millisecond, 20 * time.Millisecond},
	} {
		if got := loopDuration(test.duration, 10, 1000); got != test.want {
			t.Fatalf("loopDuration(%v) = %v, want %v", test.duration, got, test.want)
		}
	}
	// Match the sequencer's truncated block clock at the engine's sample rate.
	blockSize, sampleRate := int32(64), int32(44100)
	blockTime := time.Duration(float64(time.Second) * float64(blockSize) / float64(sampleRate))
	got := loopDuration(1000*blockTime, blockSize, sampleRate)
	want := time.Duration(float64(1000*blockSize) * float64(time.Second) / float64(sampleRate))
	if got != want {
		t.Fatalf("loop period = %v, want %v", got, want)
	}
}
