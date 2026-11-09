package meltysynth

import (
	"container/heap"
	"time"
)

type midiLoopState struct {
	sustain  [16]bool
	pressure [16]int32
}

type midiLoop struct {
	index     int
	start     time.Duration
	remaining byte
	state     midiLoopState
}

// MidiFilePlayback follows a MIDI event timeline without rendering audio. Its
// clock can therefore follow the audio player's position without accessing the
// audio thread's sequencer state. It is not safe for concurrent use.
type MidiFilePlayback struct {
	file        *MidiFile
	loop        bool
	xmi         bool
	index       int
	position    time.Duration
	elapsed     time.Duration
	duration    time.Duration
	loops       int64
	stack       []midiLoop
	state       midiLoopState
	noteEnds    []int
	matchedOffs []bool
	nextNoteID  uint64
	pending     noteOffQueue
}

func newMidiFilePlayback(file *MidiFile, loop, xmi bool) *MidiFilePlayback {
	p := &MidiFilePlayback{file: file, loop: loop, xmi: xmi, duration: file.GetLength()}
	if xmi {
		p.noteEnds, p.matchedOffs = pairNoteOffs(file)
	}
	if loop && xmi {
		// An infinite XMIDI loop never reaches the events after its NEXT.
		var stack []midiLoop
		for i, msg := range file.messages {
			if msg.command != 0xb0 || msg.getMessageType() != msg_Normal {
				continue
			}
			switch msg.data1 {
			case 116:
				stack = append(stack, midiLoop{start: file.times[i], remaining: msg.data2})
			case 117:
				if len(stack) == 0 {
					continue
				}
				frame := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if msg.data2 >= 64 && frame.remaining == 0 && file.times[i] > frame.start {
					p.duration = file.times[i]
					return p
				}
			}
		}
	}
	return p
}

// NewXMIDIPlayback follows XMIDI FOR (CC116) and NEXT/BREAK (CC117) controls
// preserved in a converted MIDI file. A FOR count of zero repeats forever when
// loop is true; finite counts are honored even when loop is false. With loop
// true, reaching the end of the file also restarts the whole file. Matched
// note-offs keep their elapsed-time deadlines when a section jumps backward.
func NewXMIDIPlayback(file *MidiFile, loop bool) *MidiFilePlayback {
	return newMidiFilePlayback(file, loop, true)
}

// AdvanceTo advances to an absolute elapsed playback time, including repeats.
// Backward movement is ignored; construct a new tracker to restart playback.
func (p *MidiFilePlayback) AdvanceTo(elapsed time.Duration) {
	if elapsed < p.elapsed {
		return
	}
	p.position += elapsed - p.elapsed
	p.elapsed = elapsed
	p.processEvents(nil, nil)
}

func (p *MidiFilePlayback) Position() time.Duration {
	return min(p.position, p.duration)
}

func (p *MidiFilePlayback) Duration() time.Duration { return p.duration }
func (p *MidiFilePlayback) Loops() int64            { return p.loops }

func (p *MidiFilePlayback) processEvents(emit func(message), onLoop func(midiLoopState, bool)) {
	for {
		offset := p.elapsed - p.position
		if len(p.pending) > 0 {
			off := p.pending[0]
			beforeNext := p.index == len(p.file.messages)
			if !beforeNext {
				nextTime := offset + p.file.times[p.index]
				beforeNext = off.at < nextTime || (off.at == nextTime && (off.offset != offset || off.index <= p.index))
			}
			if off.at <= p.elapsed && beforeNext {
				heap.Pop(&p.pending)
				if emit != nil {
					emit(off.message)
				}
				continue
			}
		}
		if p.index == len(p.file.messages) {
			length := p.file.GetLength()
			if !p.loop || length <= 0 {
				return
			}
			p.position -= length
			p.index = 0
			p.stack = p.stack[:0]
			p.loops++
			p.state = midiLoopState{}
			p.pending = p.pending[:0]
			if onLoop != nil {
				onLoop(p.state, true)
			}
			continue
		}
		if p.file.times[p.index] > p.position {
			return
		}
		messageIndex := p.index
		timestamp := p.file.times[messageIndex]
		msg := p.file.messages[messageIndex]
		p.index++
		if msg.getMessageType() != msg_Normal {
			continue
		}
		if p.xmi {
			if p.matchedOffs[messageIndex] {
				continue // Delivered by the absolute-time queue, possibly after a jump.
			}
			if end := p.noteEnds[messageIndex]; end >= 0 {
				p.nextNoteID++
				msg.noteID = p.nextNoteID
				off := p.file.messages[end]
				off.noteID = msg.noteID
				heap.Push(&p.pending, scheduledNoteOff{
					at: offset + p.file.times[end], offset: offset,
					index: end, message: off,
				})
			}
		}
		if msg.command == 0xd0 && msg.channel < 16 {
			p.state.pressure[msg.channel] = int32(msg.data1)
		}
		if msg.command == 0xb0 && msg.channel < 16 {
			switch msg.data1 {
			case 64:
				p.state.sustain[msg.channel] = msg.data2 >= 64
			case 121: // Reset All Controllers releases pedal and aftertouch.
				p.state.sustain[msg.channel] = false
				p.state.pressure[msg.channel] = 0
			}
		}
		if p.xmi && msg.command == 0xb0 {
			switch msg.data1 {
			case 116:
				p.stack = append(p.stack, midiLoop{index: p.index, start: timestamp, remaining: msg.data2, state: p.state})
				continue
			case 117:
				if len(p.stack) == 0 {
					continue
				}
				frame := &p.stack[len(p.stack)-1]
				period := timestamp - frame.start
				// NEXT values below 64 break out. Never branch around a
				// zero-duration region, which could spin forever.
				if msg.data2 < 64 || period <= 0 || frame.remaining == 1 || (frame.remaining == 0 && !p.loop) {
					p.stack = p.stack[:len(p.stack)-1]
					continue
				}
				if frame.remaining > 0 {
					frame.remaining--
				}
				p.position -= period
				p.index = frame.index
				p.loops++
				p.state = frame.state
				if onLoop != nil {
					onLoop(p.state, false)
				}
				continue
			}
		}
		if emit != nil {
			emit(msg)
		}
	}
}
