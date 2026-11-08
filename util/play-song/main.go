package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/kazzmir/master-of-magic/game/magic/audio"
	"github.com/kazzmir/master-of-magic/game/magic/music"
	"github.com/kazzmir/master-of-magic/lib/lbx"
)

type playerGame struct {
	ctx   context.Context
	entry int
}

func (game *playerGame) Update() error {
	if game.ctx.Err() != nil || ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	return nil
}

func (game *playerGame) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Playing MUSIC.LBX entry %d (looping)\nEscape or close the window to stop.", game.entry))
}

func (game *playerGame) Layout(_, _ int) (int, int) {
	return 480, 80
}

func loadCache(path string) (*lbx.LbxCache, error) {
	if path == "" {
		cache := lbx.AutoCache()
		if cache == nil {
			return nil, fmt.Errorf("could not find game data; specify -lbx /path/to/MUSIC.LBX")
		}
		return cache, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open music archive: %w", err)
	}
	defer file.Close()

	archive, err := lbx.ReadLbx(file)
	if err != nil {
		return nil, fmt.Errorf("read music archive %q: %w", path, err)
	}
	return lbx.MakeCacheFromLbxFiles(map[string]*lbx.LbxFile{"music.lbx": &archive}), nil
}

func run(args []string) error {
	flags := flag.NewFlagSet("play-song", flag.ContinueOnError)
	path := flags.String("lbx", "", "path to MUSIC.LBX (default: auto-discover game data)")
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "Usage: play-song [-lbx /path/to/MUSIC.LBX] ENTRY")
		fmt.Fprintln(flags.Output(), "ENTRY is a zero-based music entry. Playback loops until stopped.")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		flags.Usage()
		return fmt.Errorf("specify exactly one music entry")
	}
	entry, err := strconv.Atoi(flags.Arg(0))
	if err != nil || entry < 0 {
		return fmt.Errorf("invalid music entry %q: expected a non-negative integer", flags.Arg(0))
	}

	cache, err := loadCache(*path)
	if err != nil {
		return err
	}
	player := music.MakeMusic(cache)
	defer player.Stop()

	// Preload so errors are reported to the CLI rather than just logged by PlaySong.
	if _, err := player.LoadSong(music.Song(entry)); err != nil {
		return fmt.Errorf("load music entry %d: %w", entry, err)
	}
	if _, err := player.LoadSoundFont(); err != nil {
		return fmt.Errorf("load soundfont: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	audio.Initialize()
	player.PlaySong(music.Song(entry))
	ebiten.SetWindowTitle(fmt.Sprintf("MUSIC.LBX entry %d", entry))
	ebiten.SetWindowSize(480, 80)
	return ebiten.RunGame(&playerGame{ctx: ctx, entry: entry})
}

func main() {
	if err := run(os.Args[1:]); err != nil && !errors.Is(err, flag.ErrHelp) {
		log.Print(err)
		os.Exit(1)
	}
}
