package settings

import (
    "github.com/kazzmir/master-of-magic/game/magic/keybinds"
)

// SerializedSettings is the remake-only preferences blob stored next to the
// game state. Pointers distinguish "field omitted" (old save) from false.
type SerializedSettings struct {
    EndOfTurnWait *bool `json:"end-of-turn-wait,omitempty"`
    StrategicCombatOnly *bool `json:"strategic-combat-only,omitempty"`
    RandomEvents *bool `json:"random-events,omitempty"`
    AggressiveAI *bool `json:"aggressive-ai,omitempty"`
    Keybindings map[string]string `json:"keybindings,omitempty"`
}

func SerializeSettings(settings *Settings) *SerializedSettings {
    if settings == nil {
        return nil
    }

    out := &SerializedSettings{
        EndOfTurnWait: new(settings.EndOfTurnWait),
        StrategicCombatOnly: new(settings.StrategicCombatOnly),
        RandomEvents: new(settings.RandomEvents),
        AggressiveAI: new(settings.AggressiveAI),
    }

    if settings.Keybindings != nil {
        out.Keybindings = settings.Keybindings.Serialize()
    }

    return out
}

// Apply copies saved preferences onto settings. Nil fields are left alone so
// an old save without preferences does not flip EndOfTurnWait/RandomEvents
// from their defaults to false.
func (serialized *SerializedSettings) Apply(settings *Settings) {
    if serialized == nil || settings == nil {
        return
    }

    if serialized.EndOfTurnWait != nil {
        settings.EndOfTurnWait = *serialized.EndOfTurnWait
    }
    if serialized.StrategicCombatOnly != nil {
        settings.StrategicCombatOnly = *serialized.StrategicCombatOnly
    }
    if serialized.RandomEvents != nil {
        settings.RandomEvents = *serialized.RandomEvents
    }
    if serialized.AggressiveAI != nil {
        settings.AggressiveAI = *serialized.AggressiveAI
    }
    if serialized.Keybindings != nil {
        if settings.Keybindings == nil {
            settings.Keybindings = keybinds.MakeKeybindings()
        }
        settings.Keybindings.ApplySerialized(serialized.Keybindings)
    }
}
