package music

import (
    "context"
    "log"
    "bytes"
    "sync"
    "strings"
    "time"

    "github.com/kazzmir/master-of-magic/game/magic/audio"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/xmi"

    "github.com/kazzmir/master-of-magic/lib/midi"
    "github.com/kazzmir/master-of-magic/lib/midi/smf"
    midiPlayer "github.com/kazzmir/master-of-magic/lib/midi/smf/player"
    midiDrivers "github.com/kazzmir/master-of-magic/lib/midi/drivers"
)

type Song int

const (
    SongOverworld Song = 100
    SongBuildingFinished Song = 113
    SongCombat1 Song = 102
)

type Music struct {
    Cache *lbx.LbxCache
    done context.Context
    cancel context.CancelFunc
    wait sync.WaitGroup
}

func MakeMusic(cache *lbx.LbxCache) *Music {
    ctx, cancel := context.WithCancel(context.Background())
    return &Music{done: ctx, cancel: cancel, Cache: cache}
}

func (music *Music) PlaySong(index Song){
    music.cancel()

    music.done, music.cancel = context.WithCancel(context.Background())

    song, err := xmi.ReadMidiFromCache(music.Cache, "music.lbx", int(index))
    if err != nil {
        log.Printf("Error: could not read midi from cache: %v", err)
        return
    }

    music.wait.Add(1)
    go func(){
        defer music.wait.Done()

        for !audio.IsReady() {
            select {
                case <-music.done.Done():
                    return
                case <-time.After(time.Millisecond * 100):
            }
        }

        playMidi(song, music.done)
    }()
}

func (music *Music) Stop() {
    music.cancel()
    music.wait.Wait()
}

func playMidi(song *smf.SMF, done context.Context) {
    driver := midiDrivers.Get()
    if driver == nil {
        log.Printf("No midi driver available!\n")
        return
    }
    log.Printf("Got driver: %v\n", driver)
    outs, err := driver.Outs()
    if err != nil {
        log.Printf("Could not get midi output ports: %v\n", err)
    } else {
        log.Printf("Got midi output ports: %v\n", outs)
        if len(outs) > 0 {
            for _, out := range outs {

                if strings.Contains(out.String(), "Through"){
                    continue
                }

                send, err := midi.SendTo(out)
                if err != nil {
                    log.Printf("Could not send to midi output port: %v\n", err)
                    return
                }

                defer out.Close()

                var data bytes.Buffer

                _, err = song.WriteTo(&data)
                if err != nil {
                    log.Printf("Could not write midi to buffer: %v", err)
                    return
                }

                // play forever
                for done.Err() == nil {
                    _, err := midiPlayer.Play(out, bytes.NewReader(data.Bytes()), done, 0)
                    if err != nil {
                        log.Printf("Could not play midi: %v", err)
                        break
                    }


                    /* not really sure why this doesn't work, but the notes are played too fast somehow
                    _, err := midiPlayer.PlaySMF(out, song, done)
                    if err != nil {
                        log.Printf("Could not play midi: %v", err)
                        break
                    }
                    */
                }

                // turn off all notes
                for _, message := range midi.SilenceChannel(-1) {
                    send(message.Bytes())
                }

                return
            }

            log.Printf("No playable output ports available!\n")

        } else {
            log.Printf("No midi output ports available!\n")
        }
    }
}
