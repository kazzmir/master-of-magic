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
