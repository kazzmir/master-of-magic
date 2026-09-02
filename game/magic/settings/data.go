package settings

import (
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/keybinds"
)

// Settings is the single top-level holder for every user preference that
// needs to be readable/settable from both the main menu and an in-progress
// game (the same set of screens that already share Music). Everything here
// is session-only, matching how Music's own volume already behaves -
// nothing is persisted to disk.
type Settings struct {
    EndOfTurnWait bool
    StrategicCombatOnly bool
    RandomEvents bool
    // AggressiveAI is session-only (same as RandomEvents). When on, Enemy2
    // uses the meaner expansion/offense knobs; default off is Classic pacing.
    AggressiveAI bool
    Keybindings *keybinds.Keybindings
}

func MakeSettings(cache *lbx.LbxCache) *Settings {
    return &Settings{
        EndOfTurnWait: true,
        StrategicCombatOnly: false,
        RandomEvents: true,
        AggressiveAI: false,
        Keybindings: keybinds.MakeKeybindings(),
    }
}
