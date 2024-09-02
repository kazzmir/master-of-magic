package voc

import (
    "encoding/binary"
    "fmt"
    "io"
    "log"
)

var errNotACreativeVoiceSound = fmt.Errorf("Not a Creative Voice Sound")

func lengthFromBlockStart(blockStart []byte) int {
    b0 := int(blockStart[1]) << 0
    b1 := int(blockStart[2]) << 8
    b2 := int(blockStart[3]) << 16

    return b0 + b1 + b2
}

func divisorToSampleRate(divisor byte) float32 {
    return rateBase / float32(256-int(divisor))
}

// Load reads from the provided source a Creative Voice Sound and returns the data.
func Load(source io.Reader) (*L8SoundData, error) {
    if source == nil {
        return nil, fmt.Errorf("source is nil")
    }

    err := readAndVerifyHeader(source)
    if err != nil {
        return nil, err
    }
    return readSoundData(source)
}

func readAndVerifyHeader(source io.Reader) error {
    start := make([]byte, len(fileHeader))
    headerSize := uint16(0)
    version := uint16(0)
    versionValidity := uint16(0)

    source.Read(start)
    binary.Read(source, binary.LittleEndian, &headerSize)
    binary.Read(source, binary.LittleEndian, &version)
    binary.Read(source, binary.LittleEndian, &versionValidity)

    // log.Printf("Version: 0x%x", version)

    calculated := uint16(^version + versionCheckValue)
    if calculated != versionValidity {
        return fmt.Errorf("Version validity failed: 0x%04X != 0x%04X", calculated, versionValidity)
    }

    skip := make([]byte, headerSize-standardHeaderSize)
    source.Read(skip)

    return nil
}

func readSoundData(source io.Reader) (*L8SoundData, error) {
    sampleRate := float32(0.0)
    var samples []byte
    done := false

    blockStart := make([]byte, 4)
    meta := make([]byte, 2)
    for !done {
        source.Read(blockStart)

        // log.Printf("Read block %v", blockStart[0])
        switch blockType(blockStart[0]) {
        case soundData:
            source.Read(meta)
            sampleRate = divisorToSampleRate(meta[0])

            // log.Printf("Sample rate: %v from 0x%x", sampleRate, meta[0])
            // log.Printf("Codec: %v", meta[1])

            newCount := lengthFromBlockStart(blockStart) - len(meta)
            buf := make([]byte, newCount)
            source.Read(buf)

            oldCount := len(samples)
            newSamples := make([]byte, oldCount+newCount)
            copy(newSamples, samples)
            copy(newSamples[oldCount:], buf)
            samples = newSamples
        case terminator:
            done = true
        default:
            log.Printf("Unknown block type: %v", blockStart[0])
        }
    }

    if len(samples) == 0 {
        return nil, fmt.Errorf("No audio found")
    } else {
        return NewL8SoundData(sampleRate, samples), nil
    }
}
