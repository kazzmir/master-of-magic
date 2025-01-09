package audio

import (
    "fmt"
    "bytes"
    "encoding/binary"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/voc"
    audiolib "github.com/hajimehoshi/ebiten/v2/audio"
)

var Context *audiolib.Context
const SampleRate = 44100

type MakePlayerFunc func () (*audiolib.Player, error)

func Initialize(){
    Context = audiolib.NewContext(SampleRate)
}

func convertToS16(u8samples []byte) []byte {
    var out bytes.Buffer

    for _, sample := range u8samples {
        binary.Write(&out, binary.LittleEndian, (int16(sample) - 128) * 256)
        binary.Write(&out, binary.LittleEndian, (int16(sample) - 128) * 256)
    }

    return out.Bytes()
}

// precomputes the resampled sound data so all the client has to do is invoke the returned function
// f, err := GetSoundMaker(soundLbx, index)
// player, err := f()
// player.Play()
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

    resampled := audiolib.Resample(bytes.NewReader(s16Samples), int64(len(s16Samples)), int(vocData.SampleRate()), SampleRate)

    return func() (*audiolib.Player, error){
        return Context.NewPlayer(resampled)
    }, nil
}

func LoadSoundFromLbx(soundLbx *lbx.LbxFile, index int) (*audiolib.Player, error){
    maker, err := GetSoundMaker(soundLbx, index)
    if err != nil {
        return nil, err
    }

    return maker()
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

func LoadSound(cache *lbx.LbxCache, index int) (*audiolib.Player, error){
    if Context == nil {
        return nil, fmt.Errorf("audio has not been initialized")
    }

    soundLbx, err := cache.GetLbxFile("soundfx.lbx")
    if err != nil {
        return nil, err
    }

    return LoadSoundFromLbx(soundLbx, index)
}
