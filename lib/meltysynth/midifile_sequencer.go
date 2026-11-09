package meltysynth

import (
	"math"
	"time"
)

type MidiFileSequencer struct {
	synthesizer *Synthesizer
	playback    *MidiFilePlayback
	blockWrote  int32
}

func NewMidiFileSequencer(s *Synthesizer) *MidiFileSequencer {
	result := new(MidiFileSequencer)
	result.synthesizer = s
	return result
}

func (seq *MidiFileSequencer) Play(midiFile *MidiFile, loop bool) {
	seq.play(midiFile, loop, false)
}

// PlayXMI interprets the XMIDI loop controllers preserved by XMI-to-MIDI
// conversion. Ordinary Play leaves these otherwise undefined MIDI CCs alone.
func (seq *MidiFileSequencer) PlayXMI(midiFile *MidiFile, loop bool) {
	seq.play(midiFile, loop, true)
}

func (seq *MidiFileSequencer) play(midiFile *MidiFile, loop, xmi bool) {
	seq.playback = newMidiFilePlayback(midiFile, loop, xmi)
	seq.blockWrote = seq.synthesizer.BlockSize
	seq.synthesizer.Reset()
}

func (seq *MidiFileSequencer) Stop() {
	seq.playback = nil

	seq.synthesizer.Reset()
}

func (seq *MidiFileSequencer) Render(left []float32, right []float32) {
	var wrote int32
	length := int32(len(left))
	for wrote < length {
		if seq.blockWrote == seq.synthesizer.BlockSize {
			seq.processEvents()
			seq.blockWrote = 0
			if seq.playback != nil {
				seq.playback.position += time.Duration(float64(time.Second) * float64(seq.synthesizer.BlockSize) / float64(seq.synthesizer.SampleRate))
			}
		}

		srcRem := seq.synthesizer.BlockSize - seq.blockWrote
		dstRem := length - wrote
		rem := int32(math.Min(float64(srcRem), float64(dstRem)))

		seq.synthesizer.Render(left[wrote:wrote+rem], right[wrote:wrote+rem])

		seq.blockWrote += rem
		wrote += rem
	}
}

func (seq *MidiFileSequencer) processEvents() {
	if seq.playback == nil {
		return
	}
	seq.playback.processEvents(func(msg message) {
		seq.synthesizer.ProcessMidiMessage(int32(msg.channel), int32(msg.command), int32(msg.data1), int32(msg.data2))
	}, func(sustain [16]bool) {
		seq.synthesizer.restartLoop(sustain)
	})
}
