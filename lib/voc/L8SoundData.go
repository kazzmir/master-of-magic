package voc

// L8SoundData is a in-memory sound buffer.
type L8SoundData struct {
	sampleRate float32
	samples    []byte
}

// NewL8SoundData returns a new sound data instance with the given data.
func NewL8SoundData(sampleRate float32, samples []byte) *L8SoundData {
	data := &L8SoundData{
		sampleRate: sampleRate,
		samples:    samples,
    }

	return data
}

// SampleRate returns the amount of samples for one second.
func (data *L8SoundData) SampleRate() float32 {
	return data.sampleRate
}

// SampleCount returns the count of samples available in this data.
func (data *L8SoundData) SampleCount() int {
	return len(data.samples)
}

// Samples returns the samples in the given range
func (data *L8SoundData) Samples(from, to int) []byte {
	return data.samples[from:to]
}

func (data *L8SoundData) AllSamples() []byte {
    return data.samples
}
