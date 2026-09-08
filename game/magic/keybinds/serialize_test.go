package keybinds

import (
    "testing"

    "github.com/hajimehoshi/ebiten/v2"
)

func TestAllActionsHaveUniqueIDs(test *testing.T) {
    seen := make(map[string]Action)

    for _, action := range AllActions {
        id := action.ID()
        if id == "" {
            test.Errorf("action %v (%s) has no ID()", int(action), action.Name())
            continue
        }

        if other, ok := seen[id]; ok {
            test.Errorf("%v and %v share ID %q", action.Name(), other.Name(), id)
        }
        seen[id] = action

        got, ok := ActionByID(id)
        if !ok || got != action {
            test.Errorf("ActionByID(%q) = %v ok=%v, want %v", id, got, ok, action)
        }
    }
}

func TestKeyFromNameRoundTrip(test *testing.T) {
    checks := []ebiten.Key{ebiten.KeyN, ebiten.KeyF1, ebiten.KeySpace, ebiten.KeyEscape}

    for _, key := range checks {
        got, ok := KeyFromName(key.String())
        if !ok || got != key {
            test.Errorf("%v: KeyFromName(%q) = %v ok=%v", key, key.String(), got, ok)
        }
    }

    unbound, ok := KeyFromName(unboundName)
    if !ok || unbound != Unbound {
        test.Errorf("unbound name should parse as Unbound")
    }
    unbound, ok = KeyFromName("")
    if !ok || unbound != Unbound {
        test.Errorf("empty name should parse as Unbound")
    }
    if _, ok := KeyFromName("not-a-key"); ok {
        test.Errorf("unknown key name should not parse")
    }
}

func TestKeybindingsSerializeRoundTrip(test *testing.T) {
    original := MakeKeybindings()
    original.Set(ActionNextTurn, ebiten.KeySpace)
    original.Set(ActionCitiesScreen, ebiten.KeyC)
    original.Set(ActionQuitWithoutSaving, Unbound)

    restored := MakeKeybindings()
    restored.ApplySerialized(original.Serialize())

    for _, action := range AllActions {
        if restored.Get(action) != original.Get(action) {
            test.Errorf("%s: got %v, want %v", action.Name(), restored.Get(action), original.Get(action))
        }
    }
}

func TestApplySerializedKeepsDefaultForMissingAction(test *testing.T) {
    restored := MakeKeybindings()
    restored.ApplySerialized(map[string]string{
        "next-turn": "Space",
    })

    if restored.Get(ActionNextTurn) != ebiten.KeySpace {
        test.Errorf("next-turn should be Space, got %v", restored.Get(ActionNextTurn))
    }
    if restored.Get(ActionSurveyor) != ActionSurveyor.Default() {
        test.Errorf("missing actions should keep defaults")
    }
}

func TestApplySerializedUnbindsConflict(test *testing.T) {
    restored := MakeKeybindings()
    restored.ApplySerialized(map[string]string{
        "game-screen": "N",
    })

    if restored.Get(ActionGameScreen) != ebiten.KeyN {
        test.Errorf("game-screen should take N")
    }
    if restored.Get(ActionNextTurn) != Unbound {
        test.Errorf("next-turn should be unbound after losing N, got %v", restored.Get(ActionNextTurn))
    }
}

func TestApplySerializedSkipsUnknown(test *testing.T) {
    restored := MakeKeybindings()
    restored.ApplySerialized(map[string]string{
        "not-an-action": "N",
        "next-turn": "not-a-key",
    })

    if restored.Get(ActionNextTurn) != ActionNextTurn.Default() {
        test.Errorf("unknown key name should leave next-turn at default")
    }
}
