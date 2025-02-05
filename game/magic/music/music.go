package music

import (
    "context"
    "log"
    "fmt"
    "sync"
    "strings"
    "time"

    "github.com/kazzmir/master-of-magic/game/magic/audio"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/xmi"

    "github.com/kazzmir/master-of-magic/lib/midi"
    "github.com/kazzmir/master-of-magic/lib/midi/smf"
    midiDrivers "github.com/kazzmir/master-of-magic/lib/midi/drivers"
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

func (music *Music) PlaySong(index int){
    music.cancel()

    music.done, music.cancel = context.WithCancel(context.Background())

    song, err := xmi.ReadMidiFromCache(music.Cache, "music.lbx", index)
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
        fmt.Printf("No midi driver available!\n")
        return
    }
    fmt.Printf("Got driver: %v\n", driver)
    outs, err := driver.Outs()
    if err != nil {
        fmt.Printf("Could not get midi output ports: %v\n", err)
    } else {
        fmt.Printf("Got midi output ports: %v\n", outs)
        if len(outs) > 0 {
            for _, out := range outs {

                if strings.Contains(out.String(), "Through"){
                    continue
                }

                send, err := midi.SendTo(out)
                if err != nil {
                    fmt.Printf("Could not send to midi output port: %v\n", err)
                    return
                }

                defer out.Close()

                // play forever
                for {
                    for _, event := range song.Tracks[0] {
                        // fmt.Printf("Sending event: %v\n", event)
                        err := send(event.Message.Bytes())
                        if err != nil {
                            fmt.Printf("Error: %v\n", err)
                        }

                        select {
                            case <-done.Done():
                                for _, message := range midi.SilenceChannel(-1) {
                                    send(message.Bytes())
                                }
                                return
                            // FIXME: use proper delay
                            case <-time.After(time.Millisecond * time.Duration(event.Delta) * 10):
                        }
                    }
                }
            }

            fmt.Printf("No playable output ports available!\n")

        } else {
            fmt.Printf("No midi output ports available!\n")
        }
    }
}
