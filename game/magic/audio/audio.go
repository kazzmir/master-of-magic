package audio

import (
    "bytes"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/voc"
    audiolib "github.com/hajimehoshi/ebiten/v2/audio"
)

var Context *audiolib.Context
const SampleRate = 44100

func Initialize(){
    Context = audiolib.NewContext(SampleRate)
}

func LoadSound(cache *lbx.LbxCache, index int) (*audiolib.Player, error){
    if Context == nil {
        return nil, fmt.Errorf("audio has not been initialized")
    }

    soundLbx, err := cache.GetLbxFile("soundfx.lbx")
    if err != nil {
        return nil, err
    }

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

    resampled := audiolib.Resample(bytes.NewReader(vocData.AllSamples()), int64(vocData.SampleCount()), int(vocData.SampleRate()), SampleRate)

    return Context.NewPlayer(resampled)
}
