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
    "slices"
    "cmp"
    "io/fs"
    "fmt"

    "github.com/kazzmir/master-of-magic/game/magic/audio"

    "github.com/kazzmir/master-of-magic/data"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/xmi"

    "github.com/kazzmir/master-of-magic/lib/meltysynth"
    "github.com/kazzmir/master-of-magic/lib/midi/smf"
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

type MusicEvent interface {
}

type MusicEventSetVolume struct {
    Volume float64
}

type Songs struct {
    Songs []Song
}

func (song Songs) Choose() Song {
    return randomChoose(song.Songs...)
}

// PlaybackProgress describes audio played, excluding startup and buffering time.
type PlaybackProgress struct {
    Position time.Duration // Position in the source song; jumps at loop boundaries.
    Duration time.Duration // Source song length, or the first infinite loop's end.
    Total time.Duration // Total time played, including completed loops.
    Loops int64 // Number of section or whole-song loop restarts.
}

// Music uses a larger budget than MeltySynth's general-purpose default: layered
// SoundFont instruments and release tails can exhaust 64 voices in MUSIC.LBX.
const DefaultMaximumPolyphony = meltysynth.MaximumSupportedPolyphony

type PlayOptions struct {
    // MaximumPolyphony is the voice budget (8–256). Zero uses the music default.
    MaximumPolyphony int32
    // Progress receives snapshots about every 50 ms. Updates are dropped when
    // the receiver is not ready; callers must not close the channel during play.
    Progress chan<- PlaybackProgress
}

type Music struct {
    Cache *lbx.LbxCache
    done context.Context
    cancel context.CancelFunc
    wait sync.WaitGroup

    Enabled bool

    MusicEvents chan MusicEvent

    SoundFont *meltysynth.SoundFont
    XmiCache map[Song]*smf.SMF
    volumeMutex sync.Mutex // Guards volume and publishing/replacing MusicEvents.
    volume float64 // 1 for loud, 0 for silence

    // queue of songs being played. a new song can be pushed on top, or popped off
    songQueue []Songs
}

func MakeMusic(cache *lbx.LbxCache) *Music {
    ctx, cancel := context.WithCancel(context.Background())
    return &Music{
        done: ctx,
        cancel: cancel,
        Cache: cache,
        XmiCache: make(map[Song]*smf.SMF),
        Enabled: true,
        volume: 1.0,
    }
}

func randomChoose[T any](choices... T) T {
    return choices[rand.N(len(choices))]
}

func (music *Music) GetVolume() float64 {
    music.volumeMutex.Lock()
    defer music.volumeMutex.Unlock()
    return music.volume
}

func (music *Music) SetVolume(volume float64) {
    music.volumeMutex.Lock()
    defer music.volumeMutex.Unlock()
    music.volume = volume

    if music.MusicEvents != nil {
        // The one-slot queue currently carries only volume updates. Replace
        // an older pending update, rather than dropping the newest value.
        select {
            case <-music.MusicEvents:
            default:
        }
        select {
            case music.MusicEvents <- MusicEventSetVolume{Volume: volume}:
            default:
        }
    }
}

func (music *Music) IsMusicEnabled() bool {
    return music.Enabled
}

func (music *Music) SetMusicEnabled(enabled bool) {
    music.Enabled = enabled

    if !enabled {
        music.Stop()
    } else if len(music.songQueue) > 0 {
        // resume whatever song context is currently active, rather than waiting
        // for the next natural PlaySong/PushSong call somewhere else in the game
        music.PlaySong(music.songQueue[len(music.songQueue) - 1].Choose())
    }
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

func (music *Music) ClearSongQueue(){
    for len(music.songQueue) > 0 {
        music.PopSong()
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

func (music *Music) LoadSoundFont() (*meltysynth.SoundFont, error) {
    if music.SoundFont != nil {
        return music.SoundFont, nil
    }

    type candidate struct {
        File fs.File
        Name string
        Size int
    }

    var candidates []candidate

    isSoundFont := func(name string) bool {
        lower := strings.ToLower(name)
        return strings.HasSuffix(lower, ".sf2") || strings.HasSuffix(lower, ".sf3")
    }

    addCandidate := func (useFs fs.FS, path string, entry fs.DirEntry) {
        file, err := useFs.Open(path)
        if err != nil {
            log.Printf("Error opening file %v: %v", path, err)
            return
        }

        // FIXME: handle symlinks

        info, err := entry.Info()
        if err != nil {
            log.Printf("Error getting file info %v: %v", path, err)
            file.Close()
            return
        }

        log.Printf("Found soundfont candidate %v size %v", path, info.Size())

        // FIXME: check that the file is really a soundfont by opening it

        candidates = append(candidates, candidate{File: file, Size: int(info.Size()), Name: path})
    }

    makeWalkFunc := func(useFs fs.FS) fs.WalkDirFunc {
        return func(path string, entry fs.DirEntry, err error) error {
            if err != nil {
                return fs.SkipDir
            }

            if entry.IsDir() {
                return nil
            }

            if isSoundFont(entry.Name()) {
                addCandidate(useFs, path, entry)
            }

            return nil
        }
    }

    fs.WalkDir(data.Data, ".", makeWalkFunc(data.Data))
    // any other places to check by default?
    soundsDir := os.DirFS("/usr/share/sounds")
    fs.WalkDir(soundsDir, ".", makeWalkFunc(soundsDir))

    hereDir := os.DirFS(".")
    entries, err := fs.ReadDir(hereDir, ".")
    if err == nil {
        for _, entry := range entries {
            if !isSoundFont(entry.Name()) {
                continue
            }

            addCandidate(hereDir, entry.Name(), entry)
        }
    }

    if len(candidates) == 0 {
        return nil, fmt.Errorf("no soundfont candidates found")
    }

    defer func() {
        for _, candidate := range candidates {
            candidate.File.Close()
        }
    }()

    candidates = slices.SortedFunc(slices.Values(candidates), func (a, b candidate) int {
        return cmp.Compare(a.Size, b.Size)
    })

    slices.Reverse(candidates)

    // try to open the largest soundfont first
    for _, choose := range candidates {
        log.Printf("Opening soundfont %v", choose.Name)

        soundFont, err := meltysynth.NewSoundFont(choose.File)
        if err == nil {
            music.SoundFont = soundFont
            return soundFont, nil
        }
    }

    return nil, fmt.Errorf("could not open any soundfont")
}

func (music *Music) PlaySong(index Song, options... PlayOptions){
    if !music.Enabled {
        return
    }

    log.Printf("Playing song %v", index)
    music.Stop()

    music.done, music.cancel = context.WithCancel(context.Background())

    song, err := music.LoadSong(index)

    if err != nil {
        log.Printf("Error: could not read midi %v: %v", index, err)
        return
    }

    soundFont, err := music.LoadSoundFont()
    if err != nil {
        log.Printf("Error: could not load soundfont: %v", err)
        return
    }

    musicEvents := make(chan MusicEvent, 1)
    music.volumeMutex.Lock()
    music.MusicEvents = musicEvents
    music.volumeMutex.Unlock()

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

        err := playMidi(song, music.done, soundFont, music.GetVolume(), musicEvents, options)
        if err != nil {
            log.Printf("Error: could not play midi %v: %v", index, err)
        }
    }()
}

func (music *Music) Stop() {
    music.cancel()
    music.wait.Wait()

    music.volumeMutex.Lock()
    music.MusicEvents = nil
    music.volumeMutex.Unlock()
}

// implements the io.Reader interface for ebiten's NewPlayerF32 function
type MidiPlayer struct {
    Sequencer *meltysynth.MidiFileSequencer
    Left []float32
    Right []float32
}

func (player *MidiPlayer) Read(p []byte) (int, error) {
    // 4 bytes per sample, 2 channels
    samples := len(p) / 4 / 2

    if len(player.Left) < samples {
        player.Left = make([]float32, samples)
        player.Right = make([]float32, samples)
    }

    left := player.Left[:samples]
    right := player.Right[:samples]

    // actually generate the audio
    player.Sequencer.Render(left, right)

    // log.Printf("Wrote midi samples %v", midiSamples)

    // write left and right channels to p
    for i := 0; i < samples; i++ {
        v := math.Float32bits(left[i])
        p[i*8] = byte(v)
        p[i*8+1] = byte(v >> 8)
        p[i*8+2] = byte(v >> 16)
        p[i*8+3] = byte(v >> 24)

        v = math.Float32bits(right[i])
        p[i*8+4] = byte(v)
        p[i*8+5] = byte(v >> 8)
        p[i*8+6] = byte(v >> 16)
        p[i*8+7] = byte(v >> 24)
    }

    n := samples * 4 * 2

    return n, nil
}

func playbackProgress(playback *meltysynth.MidiFilePlayback, total time.Duration) PlaybackProgress {
    playback.AdvanceTo(total)
    return PlaybackProgress{
        Position: playback.Position(),
        Duration: playback.Duration(),
        Total: total,
        Loops: playback.Loops(),
    }
}

func synthesizerSettings(options []PlayOptions) (*meltysynth.SynthesizerSettings, error) {
    settings := meltysynth.NewSynthesizerSettings(audio.SampleRate)
    settings.MaximumPolyphony = DefaultMaximumPolyphony
    for _, option := range options {
        if option.MaximumPolyphony == 0 {
            continue
        }
        if option.MaximumPolyphony < meltysynth.MinimumSupportedPolyphony || option.MaximumPolyphony > meltysynth.MaximumSupportedPolyphony {
            return nil, fmt.Errorf("music polyphony must be between %d and %d, got %d", meltysynth.MinimumSupportedPolyphony, meltysynth.MaximumSupportedPolyphony, option.MaximumPolyphony)
        }
        settings.MaximumPolyphony = option.MaximumPolyphony
    }
    return settings, nil
}

// FIXME: replace smf.SMF with meltysynth.MidiFile
func playMidi(song *smf.SMF, done context.Context, soundFont *meltysynth.SoundFont, volume float64, events chan MusicEvent, options []PlayOptions) error {
    var buffer bytes.Buffer
    // write out a normal midi file
    song.WriteTo(&buffer)

    settings, err := synthesizerSettings(options)
    if err != nil {
        return err
    }
    synthesizer, err := meltysynth.NewSynthesizer(soundFont, settings)
    if err != nil {
        return err
    }

    midi, err := meltysynth.NewMidiFile(&buffer)
    if err != nil {
        return err
    }

    sequencer := meltysynth.NewMidiFileSequencer(synthesizer)
    sequencer.PlayXMI(midi, true)

    midiPlayer := &MidiPlayer{
        Sequencer: sequencer,
    }

    player, err := audio.Context.NewPlayerF32(midiPlayer)
    if err != nil {
        return err
    }

    player.SetVolume(volume)
    player.SetBufferSize(time.Second / 10)
    player.Play()

    // Position is the audio player's monotonic playback clock, even when the
    // sequencer loops. Do not read the sequencer from this goroutine: Render
    // runs on the audio thread.
    var progressTicks <-chan time.Time
    if len(options) > 0 {
        ticker := time.NewTicker(50 * time.Millisecond)
        defer ticker.Stop()
        progressTicks = ticker.C
    }
    playback := meltysynth.NewXMIDIPlayback(midi, true)
    reportProgress := func() {
        progress := playbackProgress(playback, player.Position())
        for _, option := range options {
            if option.Progress != nil {
                select {
                    case option.Progress <- progress:
                    default:
                }
            }
        }
    }
    reportProgress()

    quit := false
    for !quit {
        select {
            case <-done.Done():
                quit = true
            case <-progressTicks:
                reportProgress()
            case event := <-events:
                switch e := event.(type) {
                    case MusicEventSetVolume:
                        player.SetVolume(e.Volume)
                }
        }
    }

    player.Close()

    return nil
}
