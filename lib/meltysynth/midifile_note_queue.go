package meltysynth

import "time"

// Recover the note durations preserved by XMI conversion. Match tied note-offs
// in reverse note-on order, as lib/xmi emits them. Pairing is done in the source
// timeline, not across loop iterations.
func pairNoteOffs(file *MidiFile) ([]int, []bool) {
	ends := make([]int, len(file.messages))
	matched := make([]bool, len(file.messages))
	for i := range ends {
		ends[i] = -1
	}
	var active [16][128][]int
	for i, msg := range file.messages {
		if msg.getMessageType() != msg_Normal || msg.channel >= 16 || msg.data1 >= 128 {
			continue
		}
		notes := &active[msg.channel][msg.data1]
		if msg.command == 0x90 && msg.data2 != 0 {
			*notes = append(*notes, i)
		} else if (msg.command == 0x80 || (msg.command == 0x90 && msg.data2 == 0)) && len(*notes) > 0 {
			last := len(*notes) - 1
			ends[(*notes)[last]] = i
			*notes = (*notes)[:last]
			matched[i] = true
		}
	}
	return ends, matched
}

type scheduledNoteOff struct {
	at      time.Duration // Absolute playback time, unaffected by section jumps.
	offset  time.Duration // Playback time minus source time when scheduled.
	index   int           // Source event order for events at the same timestamp.
	message message       // Carries the individual note instance's ID.
}

type noteOffQueue []scheduledNoteOff

func (q noteOffQueue) Len() int { return len(q) }
func (q noteOffQueue) Less(i, j int) bool {
	if q[i].at != q[j].at {
		return q[i].at < q[j].at
	}
	if q[i].offset != q[j].offset {
		return q[i].offset < q[j].offset
	}
	return q[i].index < q[j].index
}
func (q noteOffQueue) Swap(i, j int)   { q[i], q[j] = q[j], q[i] }
func (q *noteOffQueue) Push(value any) { *q = append(*q, value.(scheduledNoteOff)) }
func (q *noteOffQueue) Pop() any {
	last := len(*q) - 1
	value := (*q)[last]
	*q = (*q)[:last]
	return value
}
