package keybinds

import (
    "testing"

    "github.com/hajimehoshi/ebiten/v2"
)

func TestAllActionsHaveANonEmptyName(test *testing.T) {
    for _, action := range AllActions {
        if action.Name() == "" || action.Name() == "Unknown" {
            test.Errorf("action %v has no Name() case", int(action))
        }
    }
}

func TestMakeKeybindingsSeedsDefaults(test *testing.T) {
    keybindings := MakeKeybindings()

    for _, action := range AllActions {
        if keybindings.Get(action) != action.Default() {
            test.Errorf("expected %v to default to %v but got %v", action.Name(), action.Default(), keybindings.Get(action))
        }
    }
}

func TestSetOverridesBinding(test *testing.T) {
    keybindings := MakeKeybindings()

    keybindings.Set(ActionNextTurn, ebiten.KeySpace)

    if keybindings.Get(ActionNextTurn) != ebiten.KeySpace {
        test.Errorf("expected Next Turn to be rebound to Space but got %v", keybindings.Get(ActionNextTurn))
    }

    // rebinding one action should not disturb another
    if keybindings.Get(ActionSurveyor) != ActionSurveyor.Default() {
        test.Errorf("expected Surveyor to remain at its default after rebinding a different action")
    }
}

func TestSetToUnbound(test *testing.T) {
    keybindings := MakeKeybindings()

    keybindings.Set(ActionNextTurn, Unbound)

    if keybindings.Get(ActionNextTurn) != Unbound {
        test.Errorf("expected Next Turn to be unbound but got %v", keybindings.Get(ActionNextTurn))
    }
}

// the original game's default bindings should be conflict-free - if two
// actions defaulted to the same key, only one would ever fire
func TestDefaultBindingsHaveNoDuplicates(test *testing.T) {
    seen := make(map[ebiten.Key]Action)

    for _, action := range AllActions {
        key := action.Default()
        if key == Unbound {
            continue
        }

        if other, ok := seen[key]; ok {
            test.Errorf("%v and %v both default to key %v", action.Name(), other.Name(), key)
        }
        seen[key] = action
    }
}

// ConflictingActionForKey backs the Keys screen: before applying a rebind, it
// asks the current bindings which action (if any) already holds the target key.
func TestConflictingActionForKeyFindsCurrentHolder(test *testing.T) {
    keybindings := MakeKeybindings()

    // ActionNextTurn defaults to KeyN, so that key should resolve back to
    // ActionNextTurn in a fresh keyset.
    got, ok := keybindings.ConflictingActionForKey(ebiten.KeyN)
    if !ok {
        test.Fatalf("expected KeyN to resolve to an action in defaults")
     }
    if got != ActionNextTurn {
        test.Errorf("expected KeyN to be held by %v, got %v", ActionNextTurn.Name(), got.Name())
     }
}

func TestConflictingActionForKeyUnboundReturnsNo(test *testing.T) {
    keybindings := MakeKeybindings()

    // Unbound is the sentinel for "no binding" and must never resolve to
    // any real action.
    if _, ok := keybindings.ConflictingActionForKey(Unbound); ok {
        test.Errorf("Unbound should not resolve to an action")
     }
}

func TestConflictingActionForKeyFreeKeyReturnsNo(test *testing.T) {
    keybindings := MakeKeybindings()

    // KeyX is not used by any default action.
    if _, ok := keybindings.ConflictingActionForKey(ebiten.KeyX); ok {
        test.Errorf("did not expect KeyX to be held by any action")
     }
}

// The Keys screen's flow unbinds the previous owner before applying a
// rebind, so the map stays conflict-free and ConflictingActionForKey keeps
// resolving to exactly one action per key.
func TestConflictingActionForKeyTracksRebind(test *testing.T) {
    keybindings := MakeKeybindings()

    // Simulate a confirmed rebind of ActionGameScreen onto ActionNextTurn's key.
    keybindings.Set(ActionNextTurn, Unbound)
    keybindings.Set(ActionGameScreen, ebiten.KeyN)

    got, ok := keybindings.ConflictingActionForKey(ebiten.KeyN)
    if !ok {
        test.Fatalf("expected KeyN to still resolve after rebind")
     }
    if got != ActionGameScreen {
        test.Errorf("expected KeyN to be held by %v after rebind, got %v", ActionGameScreen.Name(), got.Name())
     }

    // the old owner should now be unbound, and KeyN must resolve to exactly
    // one action (the new owner) with nothing else also claiming it.
    if keybindings.Get(ActionNextTurn) != Unbound {
        test.Errorf("expected old owner %v to be unbound after rebind", ActionNextTurn.Name())
     }
}
