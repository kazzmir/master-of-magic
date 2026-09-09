package game

import (
    "encoding/json"
    "testing"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/kazzmir/master-of-magic/game/magic/keybinds"
    settingslib "github.com/kazzmir/master-of-magic/game/magic/settings"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
)

func TestSerializedGameOldJSONLeavesPreferencesNil(test *testing.T) {
    var serialized SerializedGame
    err := json.Unmarshal([]byte(`{"metadata":{"version":1,"name":"Gabe"},"settings":{"Difficulty":2,"Opponents":2,"LandSize":2,"Magic":1},"turn":1}`), &serialized)
    if err != nil {
        test.Fatalf("unmarshal: %v", err)
    }
    if serialized.Preferences != nil {
        test.Errorf("old save must not invent a preferences object")
    }
    if serialized.Settings.Difficulty != 2 {
        test.Errorf("new-game settings should still decode")
    }
}

func TestSerializedGamePreferencesRoundTrip(test *testing.T) {
    prefs := settingslib.MakeSettings(nil)
    prefs.EndOfTurnWait = false
    prefs.StrategicCombatOnly = true
    prefs.Keybindings.Set(keybinds.ActionNextTurn, ebiten.KeySpace)

    original := SerializedGame{
        Settings: setup.NewGameSettings{Opponents: 3},
        Preferences: settingslib.SerializeSettings(prefs),
        Turn: 12,
    }

    raw, err := json.Marshal(original)
    if err != nil {
        test.Fatalf("marshal: %v", err)
    }

    var loaded SerializedGame
    if err := json.Unmarshal(raw, &loaded); err != nil {
        test.Fatalf("unmarshal: %v", err)
    }
    if loaded.Turn != 12 || loaded.Settings.Opponents != 3 {
        test.Errorf("game fields: turn=%d opponents=%d", loaded.Turn, loaded.Settings.Opponents)
    }
    if loaded.Preferences == nil {
        test.Fatalf("preferences missing after round trip")
    }

    restored := settingslib.MakeSettings(nil)
    loaded.Preferences.Apply(restored)
    if restored.EndOfTurnWait || !restored.StrategicCombatOnly {
        test.Errorf("bools: wait=%v strategic=%v", restored.EndOfTurnWait, restored.StrategicCombatOnly)
    }
    if restored.Keybindings.Get(keybinds.ActionNextTurn) != ebiten.KeySpace {
        test.Errorf("keybindings did not round trip")
    }
}
