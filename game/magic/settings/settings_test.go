package settings

import "testing"

// New settings must default to their original-game values. A regression that
// silently flips a default on/off would change player experience without anyone
// touching the setting.
func TestMakeSettingsDefaults(test *testing.T) {
    s := MakeSettings(nil)

    if !s.EndOfTurnWait {
        test.Errorf("EndOfTurnWait should default to true")
     }

    if s.StrategicCombatOnly {
        test.Errorf("StrategicCombatOnly should default to false")
     }

    if !s.RandomEvents {
        test.Errorf("RandomEvents should default to true")
     }
}
