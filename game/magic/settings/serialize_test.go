package settings

import (
    "encoding/json"
    "testing"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/kazzmir/master-of-magic/game/magic/keybinds"
)

func TestSerializeSettingsRoundTrip(test *testing.T) {
    original := MakeSettings(nil)
    original.EndOfTurnWait = false
    original.StrategicCombatOnly = true
    original.RandomEvents = false
    original.AggressiveAI = true
    original.Keybindings.Set(keybinds.ActionNextTurn, ebiten.KeySpace)

    restored := MakeSettings(nil)
    SerializeSettings(original).Apply(restored)

    if restored.EndOfTurnWait {
        test.Errorf("EndOfTurnWait should restore false")
    }
    if !restored.StrategicCombatOnly {
        test.Errorf("StrategicCombatOnly should restore true")
    }
    if restored.RandomEvents {
        test.Errorf("RandomEvents should restore false")
    }
    if !restored.AggressiveAI {
        test.Errorf("AggressiveAI should restore true")
    }
    if restored.Keybindings.Get(keybinds.ActionNextTurn) != ebiten.KeySpace {
        test.Errorf("next-turn should restore as Space")
    }
}

func TestApplyNilLeavesDefaults(test *testing.T) {
    settings := MakeSettings(nil)
    var serialized *SerializedSettings
    serialized.Apply(settings)

    if !settings.EndOfTurnWait || settings.StrategicCombatOnly || !settings.RandomEvents {
        test.Errorf("nil preferences should leave defaults, got %+v", *settings)
    }
}

func TestOldSaveJSONDoesNotZeroBools(test *testing.T) {
    // gzip JSON from before preferences existed has no such object at all.
    var serialized SerializedSettings
    if err := json.Unmarshal([]byte(`{}`), &serialized); err != nil {
        test.Fatalf("unmarshal: %v", err)
    }

    settings := MakeSettings(nil)
    serialized.Apply(settings)

    if !settings.EndOfTurnWait {
        test.Errorf("missing end-of-turn-wait must not become false")
    }
    if settings.StrategicCombatOnly {
        test.Errorf("missing strategic-combat-only should stay default false")
    }
    if !settings.RandomEvents {
        test.Errorf("missing random-events must not become false")
    }
}

func TestSerializeSettingsJSONKeys(test *testing.T) {
    raw, err := json.Marshal(SerializeSettings(MakeSettings(nil)))
    if err != nil {
        test.Fatalf("marshal: %v", err)
    }

    var payload map[string]any
    if err := json.Unmarshal(raw, &payload); err != nil {
        test.Fatalf("unmarshal: %v", err)
    }

    for _, key := range []string{"end-of-turn-wait", "strategic-combat-only", "random-events", "aggressive-ai", "keybindings"} {
        if _, ok := payload[key]; !ok {
            test.Errorf("expected %q in preferences JSON", key)
        }
    }
}
