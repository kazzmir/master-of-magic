package xmi

import (
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/midi/smf"
)

func ReadMidi(file *lbx.LbxFile, index int) (*smf.SMF, error) {
    reader, err := file.GetReader(index)
    if err != nil {
        return nil, err
    }

    return ConvertToMidi(reader)
}

func ReadMidiFromCache(cache *lbx.LbxCache, name string, index int) (*smf.SMF, error) {
    file, err := cache.GetLbxFile(name)
    if err != nil {
        return nil, err
    }

    return ReadMidi(file, index)
}
