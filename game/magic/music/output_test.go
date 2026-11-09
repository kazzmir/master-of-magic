package music

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	"github.com/kazzmir/master-of-magic/data"
	"github.com/kazzmir/master-of-magic/lib/meltysynth"
	"github.com/kazzmir/master-of-magic/lib/midi"
	"github.com/kazzmir/master-of-magic/lib/midi/smf"
)

func TestMusicSynthesizerReservesHeadroom(t *testing.T) {
	synth, err := newMusicSynthesizer(&meltysynth.SoundFont{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if synth.MasterVolume != 0.35 || synth.MaximumPolyphony != DefaultMaximumPolyphony {
		t.Fatalf("gain=%v polyphony=%d", synth.MasterVolume, synth.MaximumPolyphony)
	}
	// The measured worst bundled-font peak, at the old 0.5 synth gain,
	// should now remain below the soft knee, not merely below full scale.
	peak := 1.228326 * float64(synth.MasterVolume) / 0.5
	if peak >= musicSoftLimitKnee {
		t.Fatalf("insufficient headroom for measured peak: %v", peak)
	}
}

func TestStereoLimiterPreservesQuietSamples(t *testing.T) {
	for _, sample := range [][2]float32{
		{0, 0},
		{0.1, -0.2},
		{-0.8, 0.75},
		{0.9, -0.9},
	} {
		left, right := limitStereo(sample[0], sample[1])
		if left != sample[0] || right != sample[1] {
			t.Fatalf("quiet frame %v changed to %v/%v", sample, left, right)
		}
	}
}

func TestStereoLimiterBoundsAndBalance(t *testing.T) {
	for _, sample := range [][2]float32{
		{1, -0.5},
		{-1.228326, 0.3},
		{0.5, -2},
		{-4, 4},
		{math.MaxFloat32, -math.MaxFloat32},
	} {
		left, right := limitStereo(sample[0], sample[1])
		if math.IsNaN(float64(left)) || math.IsNaN(float64(right)) || math.Abs(float64(left)) > 1 || math.Abs(float64(right)) > 1 {
			t.Fatalf("frame %v produced unsafe output %v/%v", sample, left, right)
		}
		leftGain := float64(left) / float64(sample[0])
		rightGain := float64(right) / float64(sample[1])
		if math.Abs(leftGain-rightGain) > 1e-6*math.Max(leftGain, rightGain) {
			t.Fatalf("stereo balance changed for %v: gains %v/%v", sample, leftGain, rightGain)
		}
		if leftGain <= 0 || leftGain > 1 {
			t.Fatalf("limiter amplified or inverted frame %v", sample)
		}
	}
}

func TestStereoLimiterHasSmoothKnee(t *testing.T) {
	previous := float32(0)
	for i := 0; i <= 3000; i++ {
		input := float32(i) / 1000
		output, _ := limitStereo(input, 0)
		if output < previous || output > input || output > 1 {
			t.Fatalf("invalid transfer curve at %v: previous=%v output=%v", input, previous, output)
		}
		previous = output
	}
	low, _ := limitStereo(0.95, 0)
	high, _ := limitStereo(1, 0)
	if !(low < high && high < 1) {
		t.Fatalf("overload protection flattened peaks instead of soft limiting: %v/%v", low, high)
	}
}

func TestMidiPlayerWritesSafeFloat32Output(t *testing.T) {
	file, err := data.Data.Open("data/TimGM6mb.sf2")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	font, err := meltysynth.NewSoundFont(file)
	if err != nil {
		t.Fatal(err)
	}
	song := smf.New()
	song.TimeFormat = smf.MetricTicks(60)
	var track smf.Track
	track.Add(0, midi.NoteOn(0, 60, 127).Bytes())
	track.Add(120, midi.NoteOff(0, 60).Bytes())
	track.Close(0)
	song.Add(track)
	var buffer bytes.Buffer
	if _, err := song.WriteTo(&buffer); err != nil {
		t.Fatal(err)
	}
	midifile, err := meltysynth.NewMidiFile(&buffer)
	if err != nil {
		t.Fatal(err)
	}
	for _, gain := range []float32{musicSynthesizerGain, 20} {
		synth, err := newMusicSynthesizer(font, nil)
		if err != nil {
			t.Fatal(err)
		}
		// Stress the fallback with a deliberately much louder synth too.
		synth.MasterVolume = gain
		sequencer := meltysynth.NewMidiFileSequencer(synth)
		sequencer.PlayXMI(midifile, false)
		player := MidiPlayer{Sequencer: sequencer}
		pcm := make([]byte, 8192*8)
		n, err := player.Read(pcm)
		if err != nil || n != len(pcm) {
			t.Fatalf("Read = %d, %v", n, err)
		}
		rawPeak := float64(0)
		for i := range player.Left {
			left := math.Float32frombits(binary.LittleEndian.Uint32(pcm[i*8:]))
			right := math.Float32frombits(binary.LittleEndian.Uint32(pcm[i*8+4:]))
			if math.IsNaN(float64(left)) || math.IsNaN(float64(right)) || math.Abs(float64(left)) > 1 || math.Abs(float64(right)) > 1 {
				t.Fatalf("unsafe serialized sample at gain %v: %v/%v", gain, left, right)
			}
			rawPeak = math.Max(rawPeak, math.Max(math.Abs(float64(player.Left[i])), math.Abs(float64(player.Right[i]))))
			if gain == musicSynthesizerGain && (left != player.Left[i] || right != player.Right[i]) {
				t.Fatal("normal piano output was changed by the fallback limiter")
			}
		}
		if gain == 20 && rawPeak <= 1 {
			t.Fatal("test did not exercise overload protection")
		}
		if rawPeak == 0 {
			t.Fatal("test produced no audio")
		}
	}
}

func TestStereoLimiterRejectsNonfiniteFrames(t *testing.T) {
	for _, sample := range [][2]float32{
		{float32(math.Inf(1)), 0.5},
		{0.5, float32(math.Inf(-1))},
		{float32(math.NaN()), 0.5},
	} {
		left, right := limitStereo(sample[0], sample[1])
		if left != 0 || right != 0 {
			t.Fatal("nonfinite audio frame was not silenced")
		}
	}
}
