package audio

import (
    "fmt"
    "bytes"
    "io"
    "encoding/binary"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/voc"
    audiolib "github.com/hajimehoshi/ebiten/v2/audio"
)

var Context *audiolib.Context
const SampleRate = 44100

// from soundfx.lbx
const SoundClick = 2

type MakePlayerFunc func () (*audiolib.Player)

func Initialize(){
    Context = audiolib.NewContext(SampleRate)
}

func IsReady() bool {
    return Context != nil && Context.IsReady()
}

func convertToS16(u8samples []byte) []byte {
    var out bytes.Buffer

    for _, sample := range u8samples {
        binary.Write(&out, binary.LittleEndian, (int16(sample) - 128) * 256)
        binary.Write(&out, binary.LittleEndian, (int16(sample) - 128) * 256)
    }

    return out.Bytes()
}

func SaveVoc(outputFile io.Writer, soundLbx *lbx.LbxFile, entryIndex int) error {
    data, err := soundLbx.RawData(entryIndex)
    if err != nil {
        return err
    }

    reader := bytes.NewReader(data)
    reader.Seek(16, io.SeekStart)

    vocFile, err := voc.Load(reader)
    if err != nil {
        return err
    }

    voc.Save(outputFile, vocFile.SampleRate(), vocFile.AllSamples())

    return nil
}


func SaveWav(outputFile io.Writer, soundLbx *lbx.LbxFile, index int) error {
    data, err := soundLbx.RawData(index)
    if err != nil {
        return err
    }

    reader := bytes.NewReader(data)
    reader.Seek(16, 0)

    vocData, err := voc.Load(reader)
    if err != nil {
        return err
    }

    s16Samples := convertToS16(vocData.AllSamples())

    resampled := audiolib.ResampleReader(bytes.NewReader(s16Samples), int64(len(s16Samples)), int(vocData.SampleRate()), SampleRate)

    pcmData, err := io.ReadAll(resampled)
    if err != nil {
        return err
    }

    channels := 2
    bitsPerSample := 16
    dataLength := len(pcmData)
    bytePerBloc := channels * bitsPerSample / 8
    bytePerSec := SampleRate * bytePerBloc
    binary.Write(outputFile, binary.LittleEndian, []byte("RIFF"))
    binary.Write(outputFile, binary.LittleEndian, uint32(dataLength + 36))
    binary.Write(outputFile, binary.LittleEndian, []byte("WAVE"))
    binary.Write(outputFile, binary.LittleEndian, []byte("fmt "))
    binary.Write(outputFile, binary.LittleEndian, uint32(16))  // BlocSize
    binary.Write(outputFile, binary.LittleEndian, uint16(1))   // AudioFormat
    binary.Write(outputFile, binary.LittleEndian, uint16(channels))
    binary.Write(outputFile, binary.LittleEndian, uint32(SampleRate))
    binary.Write(outputFile, binary.LittleEndian, uint32(bytePerSec))
    binary.Write(outputFile, binary.LittleEndian, uint16(bytePerBloc))
    binary.Write(outputFile, binary.LittleEndian, uint16(16)) // BitsPerSample
    binary.Write(outputFile, binary.LittleEndian, []byte("data"))
    binary.Write(outputFile, binary.LittleEndian, uint32(dataLength))
    binary.Write(outputFile, binary.LittleEndian, pcmData)

    return nil
}

// precomputes the resampled sound data so all the client has to do is invoke the returned function. this is useful if
// you want to play the same sound multiple times
//   f, err := GetSoundMaker(soundLbx, index)
//   player, err := f()
//   player.Play()
//    // play again
//   player, err = f()
//   player.Play()
func GetSoundMaker(soundLbx *lbx.LbxFile, index int) (MakePlayerFunc, error) {
    data, err := soundLbx.RawData(index)
    if err != nil {
        return nil, err
    }

    reader := bytes.NewReader(data)
    reader.Seek(16, 0)

    vocData, err := voc.Load(reader)
    if err != nil {
        return nil, err
    }

    s16Samples := convertToS16(vocData.AllSamples())

    resampled := audiolib.ResampleReader(bytes.NewReader(s16Samples), int64(len(s16Samples)), int(vocData.SampleRate()), SampleRate)

    resampledData, err := io.ReadAll(resampled)
    if err != nil {
        return nil, err
    }

    return func() (*audiolib.Player){
        // return Context.NewPlayer(resampled)
        return Context.NewPlayerFromBytes(resampledData)
    }, nil
}

func LoadSoundFromLbx(soundLbx *lbx.LbxFile, index int) (*audiolib.Player, error){
    maker, err := GetSoundMaker(soundLbx, index)
    if err != nil {
        return nil, err
    }

    return maker(), nil
}

func LoadCombatSound(cache *lbx.LbxCache, index int) (*audiolib.Player, error){
    if Context == nil {
        return nil, fmt.Errorf("audio has not been initialized")
    }

    soundLbx, err := cache.GetLbxFile("cmbtsnd.lbx")
    if err != nil {
        return nil, err
    }

    return LoadSoundFromLbx(soundLbx, index)
}

func LoadNewSound(cache *lbx.LbxCache, index int) (*audiolib.Player, error){
    if Context == nil {
        return nil, fmt.Errorf("audio has not been initialized")
    }

    soundLbx, err := cache.GetLbxFile("newsound.lbx")
    if err != nil {
        return nil, err
    }

    return LoadSoundFromLbx(soundLbx, index)
}

func LoadSoundMaker(cache *lbx.LbxCache, index int) (MakePlayerFunc, error){
    if Context == nil {
        return nil, fmt.Errorf("audio has not been initialized")
    }

    soundLbx, err := cache.GetLbxFile("soundfx.lbx")
    if err != nil {
        return nil, err
    }

    return GetSoundMaker(soundLbx, index)
}

func LoadSound(cache *lbx.LbxCache, index int) (*audiolib.Player, error){
    // FIXME: what is the lowest index here? There are 21 new sounds, so probably 256 - 21
    if index > 230 {
        return LoadNewSound(cache, 256 - index)
    }

    maker, err := LoadSoundMaker(cache, index)
    if err != nil {
        return nil, err
    }

    return maker(), nil
}
