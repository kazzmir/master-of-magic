package voc

import (
	"encoding/binary"
	"fmt"
	"io"
)

var errNotACreativeVoiceSound = fmt.Errorf("Not a Creative Voice Sound")

// Load reads from the provided source a Creative Voice Sound and returns the data.
func Load(source io.Reader) (data *L8SoundData, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()

	if source == nil {
		panic(fmt.Errorf("source is nil"))
	}

	readAndVerifyHeader(source)
	data = readSoundData(source)

	return
}

func readAndVerifyHeader(source io.Reader) {
	start := make([]byte, len(fileHeader))
	headerSize := uint16(0)
	version := uint16(0)
	versionValidity := uint16(0)

	source.Read(start)
	binary.Read(source, binary.LittleEndian, &headerSize)
	binary.Read(source, binary.LittleEndian, &version)
	binary.Read(source, binary.LittleEndian, &versionValidity)

	calculated := uint16(^version + versionCheckValue)
	if calculated != versionValidity {
		panic(fmt.Errorf("Version validity failed: 0x%04X != 0x%04X", calculated, versionValidity))
	}

	skip := make([]byte, headerSize-standardHeaderSize)
	source.Read(skip)
}

func readSoundData(source io.Reader) (data *L8SoundData) {
	sampleRate := float32(0.0)
	var samples []byte
	done := false

	for !done {
		blockStart := make([]byte, 4)

		source.Read(blockStart)
		switch blockType(blockStart[0]) {
		case soundData:
			{
				meta := make([]byte, 2)
				source.Read(meta)
				sampleRate = divisorToSampleRate(meta[0])

				newCount := lengthFromBlockStart(blockStart) - len(meta)
				buf := make([]byte, newCount)
				source.Read(buf)

				oldCount := len(samples)
				newSamples := make([]byte, oldCount+newCount)
				copy(newSamples, samples)
				copy(newSamples[oldCount:], buf)
				samples = newSamples
			}
		case terminator:
			{
				done = true
			}
		}
	}

	if len(samples) == 0 {
		panic(fmt.Errorf("No audio found"))
	} else {
		data = NewL8SoundData(sampleRate, samples)
	}

	return
}
