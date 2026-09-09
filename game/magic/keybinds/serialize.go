package keybinds

import (
    "github.com/hajimehoshi/ebiten/v2"
)

const unboundName = "unbound"

func keyName(key ebiten.Key) string {
    if key == Unbound {
        return unboundName
    }

    name := key.String()
    if name == "" {
        return unboundName
    }

    return name
}

func KeyFromName(name string) (ebiten.Key, bool) {
    if name == "" || name == unboundName {
        return Unbound, true
    }

    for key := ebiten.Key(0); key <= ebiten.KeyMax; key++ {
        if key.String() == name {
            return key, true
        }
    }

    return Unbound, false
}

// Serialize writes every action as ID -> ebiten key name (or "unbound").
func (keybindings *Keybindings) Serialize() map[string]string {
    out := make(map[string]string, len(AllActions))
    if keybindings == nil {
        return out
    }

    for _, action := range AllActions {
        out[action.ID()] = keyName(keybindings.Get(action))
    }

    return out
}

// ApplySerialized overlays saved bindings onto the current map. Unknown action
// IDs and key names are skipped so a newer binary can still load an older save
// (new actions keep their defaults). A key claimed by two actions unbinds the
// previous owner, matching the Keys screen.
func (keybindings *Keybindings) ApplySerialized(serialized map[string]string) {
    if keybindings == nil || serialized == nil {
        return
    }

    for _, action := range AllActions {
        name, ok := serialized[action.ID()]
        if !ok {
            continue
        }

        key, ok := KeyFromName(name)
        if !ok {
            continue
        }

        if key != Unbound {
            if other, conflict := keybindings.ConflictingActionForKey(key); conflict && other != action {
                keybindings.Set(other, Unbound)
            }
        }

        keybindings.Set(action, key)
    }
}
