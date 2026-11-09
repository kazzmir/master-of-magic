package music

import (
	"sync"
	"testing"
	"time"
)

func requireQueuedVolume(t *testing.T, music *Music, want float64) {
	t.Helper()
	select {
	case event := <-music.MusicEvents:
		update, ok := event.(MusicEventSetVolume)
		if !ok || update.Volume != want {
			t.Fatalf("queued update = %+v, want volume %v", event, want)
		}
	default:
		t.Fatal("missing volume update")
	}
	if len(music.MusicEvents) != 0 {
		t.Fatal("stale volume updates remain queued")
	}
	if got := music.GetVolume(); got != want {
		t.Fatalf("stored volume = %v, want %v", got, want)
	}
}

func TestSetVolumeKeepsLatestUpdate(t *testing.T) {
	music := MakeMusic(nil)
	defer music.Stop()
	music.MusicEvents = make(chan MusicEvent, 1)
	music.SetVolume(0.25)
	music.SetVolume(0.75)
	requireQueuedVolume(t, music, 0.75)

	// An update already consumed by playback must not prevent the next one.
	music.SetVolume(0.5)
	requireQueuedVolume(t, music, 0.5)
	music.SetVolume(0)
	requireQueuedVolume(t, music, 0)
}

func TestSetVolumeBurstDoesNotBlock(t *testing.T) {
	music := MakeMusic(nil)
	defer music.Stop()
	music.MusicEvents = make(chan MusicEvent, 1)
	done := make(chan struct{})
	go func() {
		for i := range 10000 {
			music.SetVolume(float64(i%100) / 100)
		}
		music.SetVolume(0)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("SetVolume blocked on a full queue")
	}
	requireQueuedVolume(t, music, 0)
}

func TestSetVolumeWithConcurrentPlaybackReceiver(t *testing.T) {
	music := MakeMusic(nil)
	events := make(chan MusicEvent, 1)
	music.MusicEvents = events
	stop, done, final := make(chan struct{}), make(chan struct{}), make(chan struct{})
	const want = 0.123456
	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			case event := <-events:
				if event.(MusicEventSetVolume).Volume == want {
					close(final)
				}
			}
		}
	}()
	defer func() {
		close(stop)
		<-done
		music.Stop()
	}()
	for i := range 10000 {
		music.SetVolume(float64(i%100) / 100)
	}
	music.SetVolume(want)
	select {
	case <-final:
	case <-time.After(5 * time.Second):
		t.Fatal("playback never received the final volume update")
	}
}

func TestSetVolumeWithoutPlayback(t *testing.T) {
	music := MakeMusic(nil)
	if music.GetVolume() != 1 {
		t.Fatal("incorrect initial volume")
	}
	music.SetVolume(0.25)
	if music.GetVolume() != 0.25 {
		t.Fatal("volume was not stored without a player")
	}
	music.Stop()
	music.SetVolume(0.75)
	if music.GetVolume() != 0.75 {
		t.Fatal("volume was not stored after stopping playback")
	}
}

func TestVolumeConcurrentReadersAndWriters(t *testing.T) {
	music := MakeMusic(nil)
	defer music.Stop()
	music.MusicEvents = make(chan MusicEvent, 1)
	var wait sync.WaitGroup
	for worker := range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for i := range 1000 {
				music.SetVolume(float64((worker+i)%100) / 100)
				music.GetVolume()
			}
		}()
	}
	wait.Wait()
	music.SetVolume(0.42)
	requireQueuedVolume(t, music, 0.42)
}
