package settings

import (
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/keybinds"
    musiclib "github.com/kazzmir/master-of-magic/game/magic/music"
)

// Settings is the single top-level holder for every user preference that
// needs to be readable/settable from both the main menu and an in-progress
// game (the same set of screens that already share Music). Everything here
// is session-only, matching how Music's own volume already behaves -
// nothing is persisted to disk.
type Settings struct {
    Music *musiclib.Music
    EndOfTurnWait bool
    StrategicCombatOnly bool
    Keybindings *keybinds.Keybindings
}

func MakeSettings(cache *lbx.LbxCache) *Settings {
    return &Settings{
        Music: musiclib.MakeMusic(cache),
        EndOfTurnWait: true,
        StrategicCombatOnly: false,
        Keybindings: keybinds.MakeKeybindings(),
    }
}
