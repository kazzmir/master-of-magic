//go:build examples

package midi_test

import (
	"fmt"
	"os"

	. "github.com/kazzmir/master-of-magic/lib/midi"
	"github.com/kazzmir/master-of-magic/lib/midi/gm"

	// testdrv has one in port and one out port which is connected to the in port
	_ "github.com/kazzmir/master-of-magic/lib/midi/drivers/testdrv"
	//_ "github.com/kazzmir/master-of-magic/lib/midi/drivers/rtmididrv"
	// when using rtmidi ("for real"), replace with the line above
)

func Example() {

	var eachMessage = func(msg Message, timestampms int32) {
		if msg.Is(RealTimeMsg) {
			// ignore realtime messages
			return
		}
		var channel, key, velocity, cc, val, prog uint8
		switch {

		// is better, than to use GetNoteOn (handles note on messages with velocity of 0 as expected)
		case msg.GetNoteStart(&channel, &key, &velocity):
			fmt.Printf("note started channel: %v key: %v (%s) velocity: %v\n", channel, key, Note(key), velocity)

		// is better, than to use GetNoteOff (handles note on messages with velocity of 0 as expected)
		case msg.GetNoteEnd(&channel, &key):
			fmt.Printf("note ended channel: %v key: %v (%s)\n", channel, key, Note(key))

		case msg.GetControlChange(&channel, &cc, &val):
			fmt.Printf("control change %v (%s) channel: %v value: %v\n", cc, ControlChangeName[cc], channel, val)

		case msg.GetProgramChange(&channel, &prog):
			fmt.Printf("program change %v (%s) channel: %v\n", prog, gm.Instr(prog), channel)

		default:
			fmt.Printf("%s\n", msg)
		}
	}

	// always good to close the driver at the end
	defer CloseDriver()

	// allows you to get the ports when using "real" drivers like rtmididrv or portmididrv
	if len(os.Args) == 2 && os.Args[1] == "list" {
		fmt.Printf("MIDI IN Ports\n")
		fmt.Println(GetInPorts())
		fmt.Printf("\n\nMIDI OUT Ports\n")
		fmt.Println(GetOutPorts())
		fmt.Printf("\n\n")
		return
	}

	var out, _ = OutPort(0)
	// takes the first out port, for real, consider
	// var out = OutByName("my synth")

	// creates a sender function to the out port
	send, _ := SendTo(out)

	var in, _ = InPort(0)
	// here we take first in port, for real, consider
	// var in = InByName("my midi keyboard")

	// listens to the in port and calls eachMessage for every message.
	// any running status bytes are converted and only complete messages are passed to the eachMessage.
	stop, _ := ListenTo(in, eachMessage)

	{ // send some messages
		send(NoteOn(0, Db(5), 100))
		send(NoteOff(0, Db(5)))
		send(Pitchbend(0, -12))
		send(ProgramChange(1, gm.Instr_AcousticBass.Value()))
		send(ControlChange(2, FootPedalMSB, On))
	}

	// stops listening
	stop()

	// Output:
	// note started channel: 0 key: 61 (Db5) velocity: 100
	// note ended channel: 0 key: 61 (Db5)
	// PitchBend channel: 0 pitch: -12 (8180)
	// program change 32 (AcousticBass) channel: 1
	// control change 4 (Foot Pedal (MSB)) channel: 2 value: 127

}
