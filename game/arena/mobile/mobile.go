//go:build android

// Package mobile binds the arena to an Android EbitenView.
package mobile

import (
    "archive/zip"
    "bytes"
    _ "embed"
    "fmt"

    "github.com/kazzmir/master-of-magic/game/arena/engine"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/mouse"
    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
    ebitenmobile "github.com/hajimehoshi/ebiten/v2/mobile"
)

// The build script stages legally supplied LBX files here, never in git.
//go:embed assets/game.zip
var gameData []byte

type game struct {
    engine *engine.Engine
}

func (game *game) Update() error {
    if game.engine == nil {
        archive, err := zip.NewReader(bytes.NewReader(gameData), int64(len(gameData)))
        if err != nil {
            return fmt.Errorf("open bundled game data: %w", err)
        }
        audio.Initialize()
        mouse.Initialize()
        game.engine = engine.MakeEngine(lbx.MakeLbxCache(archive))
    }
    return game.engine.Update()
}

func (game *game) Draw(screen *ebiten.Image) {
    if game.engine != nil {
        game.engine.Draw(screen)
    }
}

func (game *game) Layout(width, height int) (int, int) {
    // Keep the desktop shop usable regardless of the device's pixel density.
    if game.engine != nil {
        return game.engine.Layout(1200, 1050)
    }
    return 1200, 1050
}

func init() {
    ebitenmobile.SetGame(&game{})
}

// Dummy ensures gomobile generates bindings for this package.
func Dummy() {}
