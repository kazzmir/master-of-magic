package music

import (
    "context"
    "log"
    "bytes"
    "sync"
    "strings"
    "time"
    "math/rand/v2"
    "math"
    "os"

    "github.com/kazzmir/master-of-magic/game/magic/audio"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/xmi"

    "github.com/kazzmir/master-of-magic/lib/meltysynth"
    "github.com/kazzmir/master-of-magic/lib/midi"
    "github.com/kazzmir/master-of-magic/lib/midi/smf"
    midiPlayer "github.com/kazzmir/master-of-magic/lib/midi/smf/player"
    midiDrivers "github.com/kazzmir/master-of-magic/lib/midi/drivers"
)

type Song int

const (
    SongNone Song = -1

    // summoning circle, spell of return
    SongSummoningCircle Song = 0
    SongGreatUnsummoning Song = SongSummoningCircle
    SongDeathWish Song = SongSummoningCircle
    SongWallOfFire Song = 1
    // dark rituals, wall of darkness, cloud of shadow
    SongDarkRituals Song = 2
    // HeavenlyLight-Prosperity-AltarOfBattle-StreamOfLife-Inspirations-AstralGate-Consecration
    SongHeavenlyLight = 3
    // WallOfStone-NaturesEye-Earthquake-GaiasBlessing-MoveFortress-EarthGate
    SongWallOfStone = 4
    // SpellWard-FlyingFortress
    SongSpellWard Song = 5
    SongChaosRift Song = 7

    // evil presence, famine, cursed lands, pestilence
    SongEvilPresence Song = 8

    // spell of mastery start and end
    SongSpellOfMastery Song = 12

    SongCommonSummoningSpell Song = 13

    // uncommon summoning spell, planar seal
    SongUncommonSummoningSpell Song = 14

    SongRareSummoningSpell Song = 15

    // very rare summoning spell, hero, artifact
    SongVeryRareSummoningSpell Song = 16
    SongSupressMagicActivating Song = 17
    SongEternalNight Song = 19
    SongEvilOmens Song = 20
    SongZombieMastery Song = 21
    SongAuraOfMajesty Song = 22
    SongWindMastery Song = 23
    SongSuppressMagic Song = 24
    SongTimeStop Song = 25
    SongNatureAwareness Song = 26
    SongNaturesWrath Song = 27
    SongHerbMastery Song = 28
    SongChaosSurge Song = 29
    SongDoomMastery Song = 30
    SongGreatWasting Song = 31
    SongMeteorStorm Song = 32
    SongArmageddon Song = 33
    SongTranquility Song = 34
    SongLifeForce Song = 35
    SongCrusade Song = 36
    SongJustCause Song = 37
    SongHolyArms Song = 38

    // researched a spell, detect magic, awareness, charm of life
    SongLearnSpell Song = 40
    SongMerlin Song = 41
    SongMerlinMad Song = 42
    SongRaven Song = 43
    SongRavenMad Song = 44
    SongSharee Song = 45
    SongShareeMad Song = 46
    SongLoPan Song = 47
    SongLoPanMad Song = 48
    SongJafar Song = 49
    SongJafarMad Song = 50
    SongOberic Song = 51
    SongObericMad Song = 52
    SongRjak Song = 53
    SongRjakMad Song = 54
    SongSssra Song = 55
    SongSssraMad Song = 56
    SongTauron Song = 57
    SongTauronMad Song = 58
    SongFreya Song = 59
    SongFreyaMad Song = 60
    SongHorus Song = 61
    SongHorusMad Song = 62
    SongAriel Song = 63
    SongArielMad Song = 64
    SongTlaloc Song = 65
    SongTlalocMad Song = 66
    SongKali Song = 67
    SongKaliMad Song = 68
    SongCombatMerlin1 Song = 71
    SongCombatMerlin2 Song = 72
    SongCombatRaven1 Song = 73
    SongCombatRaven2 Song = 74
    SongCombatSharee1 Song = 75
    SongCombatSharee2 Song = 76
    SongCombatLoPan1 Song = 77
    SongCombatLoPan2 Song = 78
    SongCombatJafar1 Song = 79
    SongCombatJafar2 Song = 80
    SongCombatOberic1 Song = 81
    SongCombatOberic2 Song = 82
    SongCombatRjak1 Song = 83
    SongCombatRjak2 Song = 84
    SongCombatSssra1 Song = 85
    SongCombatSssra2 Song = 86
    SongCombatTauron1 Song = 87
    SongCombatTauron2 Song = 88
    SongCombatFreya1 Song = 89
    SongCombatFreya2 Song = 90
    SongCombatHorus1 Song = 91
    SongCombatHorus2 Song = 92
    SongCombatAriel1 Song = 93
    SongCombatAriel2 Song = 94
    SongCombatTlaloc1 Song = 95
    SongCombatTlaloc2 Song = 96
    SongCombatKali1 Song = 97
    SongCombatKali2 Song = 98
    SongBackground1 Song = 99
    SongBackground2 Song = 100
    SongBackground3 Song = 101
    SongCombat1 Song = 102
    SongCombat2 Song = 103
    SongTitle Song = 104
    SongSiteDiscovery Song = 105
    SongGoodEvent Song = 106
    SongBuildingFinished Song = 108
    SongYouWin Song = 109
    SongYouLose Song = 110
    SongIntro Song = 112
    SongBadEvent Song = 114
    SongHeroGainedALevel Song = 115
)

type Songs struct {
    Songs []Song
}

func (song Songs) Choose() Song {
    return randomChoose(song.Songs...)
}

type Music struct {
    Cache *lbx.LbxCache
    done context.Context
    cancel context.CancelFunc
    wait sync.WaitGroup

    XmiCache map[Song]*smf.SMF

    // queue of songs being played. a new song can be pushed on top, or popped off
    songQueue []Songs
}

func MakeMusic(cache *lbx.LbxCache) *Music {
    ctx, cancel := context.WithCancel(context.Background())
    return &Music{done: ctx, cancel: cancel, Cache: cache, XmiCache: make(map[Song]*smf.SMF)}
}

func randomChoose[T any](choices... T) T {
    return choices[rand.N(len(choices))]
}

func (music *Music) PushSongs(songs... Song) {
    music.songQueue = append(music.songQueue, Songs{Songs: songs})
    music.PlaySong(music.songQueue[len(music.songQueue)-1].Choose())
}

func (music *Music) PushSong(index Song){
    music.PushSongs(index)
}

func (music *Music) PopSong(){
    if len(music.songQueue) > 0 {
        music.songQueue = music.songQueue[:len(music.songQueue)-1]
        music.Stop()
        if len(music.songQueue) > 0 {
            music.PlaySong(music.songQueue[len(music.songQueue)-1].Choose())
        }
    }
}

func (music *Music) LoadSong(index Song) (*smf.SMF, error) {
    if song, ok := music.XmiCache[index]; ok {
        return song, nil
    }

    song, err := xmi.ReadMidiFromCache(music.Cache, "music.lbx", int(index))
    if err != nil {
        return nil, err
    }
    music.XmiCache[index] = song
    return song, nil
}

func (music *Music) PlaySong(index Song){
    log.Printf("Playing song %v", index)
    music.Stop()

    music.done, music.cancel = context.WithCancel(context.Background())

    song, err := music.LoadSong(index)

    if err != nil {
        log.Printf("Error: could not read midi %v: %v", index, err)
        return
    }

    soundFontPath := "/usr/share/sounds/sf2/FluidR3_GM.sf2"
    sf2, err := os.Open(soundFontPath)
    if err != nil {
        log.Printf("Error: could not open soundfont %v: %v", soundFontPath, err)
        return
    }
    defer sf2.Close()

    soundFont, err := meltysynth.NewSoundFont(sf2)
    if err != nil {
        log.Printf("Error: could not load soundfont %v: %v", soundFontPath, err)
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

        err := playMidi(song, music.done, soundFont)
        if err != nil {
            log.Printf("Error: could not play midi %v: %v", index, err)
        }
    }()
}

func (music *Music) Stop() {
    music.cancel()
    music.wait.Wait()
}

type MidiPlayer struct {
    Sequencer *meltysynth.MidiFileSequencer
    Left []float32
    Right []float32
}

func (player *MidiPlayer) Read(p []byte) (int, error) {
    samples := len(p) / 4 / 2

    /*
    engineLeft, engineRight := engine.PCMReader.Lock()
    defer engine.PCMReader.Unlock()
    */

    if len(player.Left) < samples {
        player.Left = make([]float32, samples)
        player.Right = make([]float32, samples)
    }

    left := player.Left
    right := player.Right

    // log.Printf("Render audio %v samples", len(left))
    player.Sequencer.Render(left, right)
    // log.Printf("Wrote midi samples %v", midiSamples)

    // myWriter := &ByteWriter{Data: p}

    for i := 0; i < samples; i++ {
        /*
        binary.Write(myWriter, binary.LittleEndian, left[i])
        binary.Write(myWriter, binary.LittleEndian, right[i])
        */

        v := math.Float32bits(left[i])
        p[i*8] = byte(v)
        p[i*8+1] = byte(v >> 8)
        p[i*8+2] = byte(v >> 16)
        p[i*8+3] = byte(v >> 24)
        // binary.LittleEndian.PutUint32(p[i*4:], v)
        v = math.Float32bits(right[i])
        // binary.LittleEndian.PutUint32(p[i*4+4:], v)
        p[i*8+4] = byte(v)
        p[i*8+5] = byte(v >> 8)
        p[i*8+6] = byte(v >> 16)
        p[i*8+7] = byte(v >> 24)
    }

    n := samples * 4 * 2

    return n, nil
}

// FIXME: replace smf.SMF with meltysynth.MidiFile
func playMidi(song *smf.SMF, done context.Context, soundFont *meltysynth.SoundFont) error {
    var buffer bytes.Buffer
    // write out a normal midi file
    song.WriteTo(&buffer)

    settings := meltysynth.NewSynthesizerSettings(audio.SampleRate)
    synthesizer, err := meltysynth.NewSynthesizer(soundFont, settings)
    if err != nil {
        return err
    }

    midi, err := meltysynth.NewMidiFile(&buffer)
    if err != nil {
        return err
    }

    sequencer := meltysynth.NewMidiFileSequencer(synthesizer)
    sequencer.Play(midi, true)

    midiPlayer := &MidiPlayer{
        Sequencer: sequencer,
    }

    player, err := audio.Context.NewPlayerF32(midiPlayer)
    if err != nil {
        return err
    }

    player.SetVolume(1.0)
    player.SetBufferSize(time.Second / 10)
    player.Play()

    <-done.Done()
    player.Close()

    return nil
}

func playMidi2(song *smf.SMF, done context.Context) {
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

                if strings.Contains(strings.ToLower(out.String()), "through"){
                    continue
                }

                log.Printf("Using midi output port: %v", out)

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
