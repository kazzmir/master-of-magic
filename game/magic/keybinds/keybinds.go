package keybinds

import (
    "github.com/hajimehoshi/ebiten/v2"
)

// Unbound is the sentinel value for an action with no key bound to it,
// displayed as '---' in the Keys screen. ebiten.Key is a plain int with no
// negative values of its own, so -1 can never collide with a real key.
const Unbound ebiten.Key = -1

type Action int

const (
    ActionGameScreen Action = iota
    ActionOpenSpellbook
    ActionArmiesScreen
    ActionCitiesScreen
    ActionMagicScreen
    ActionAdvisors
    ActionSwitchPlanes
    ActionSurveyor
    ActionCartographer
    ActionApprentice
    ActionHistorian
    ActionAstrologer
    ActionChancellor
    ActionTaxCollector
    ActionGrandVizier
    ActionMirror
    ActionNextTurn
)

// AllActions lists every rebindable action, in the order they should be
// shown in the Keys screen (matches the original game's ordering).
var AllActions = []Action{
    ActionGameScreen,
    ActionOpenSpellbook,
    ActionArmiesScreen,
    ActionCitiesScreen,
    ActionMagicScreen,
    ActionAdvisors,
    ActionSwitchPlanes,
    ActionSurveyor,
    ActionCartographer,
    ActionApprentice,
    ActionHistorian,
    ActionAstrologer,
    ActionChancellor,
    ActionTaxCollector,
    ActionGrandVizier,
    ActionMirror,
    ActionNextTurn,
}

func (action Action) Name() string {
    switch action {
        case ActionGameScreen: return "Game Screen"
        case ActionOpenSpellbook: return "Open Spellbook"
        case ActionArmiesScreen: return "Armies Screen"
        case ActionCitiesScreen: return "Cities Screen"
        case ActionMagicScreen: return "Magic Screen"
        case ActionAdvisors: return "Advisors"
        case ActionSwitchPlanes: return "Switch Planes"
        case ActionSurveyor: return "Surveyor"
        case ActionCartographer: return "Cartographer"
        case ActionApprentice: return "Apprentice"
        case ActionHistorian: return "Historian"
        case ActionAstrologer: return "Astrologer"
        case ActionChancellor: return "Chancellor"
        case ActionTaxCollector: return "Tax Collector"
        case ActionGrandVizier: return "Grand Vizier"
        case ActionMirror: return "Mirror"
        case ActionNextTurn: return "Next Turn"
    }

    return "Unknown"
}

// Default returns the original game's default key binding for this action,
// or Unbound if the original game leaves it unbound by default.
func (action Action) Default() ebiten.Key {
    switch action {
        case ActionGameScreen: return ebiten.KeyG
        case ActionOpenSpellbook: return ebiten.KeyS
        case ActionArmiesScreen: return ebiten.KeyA
        case ActionCitiesScreen: return Unbound
        case ActionMagicScreen: return ebiten.KeyM
        case ActionAdvisors: return ebiten.KeyI
        case ActionSwitchPlanes: return ebiten.KeyP
        case ActionSurveyor: return ebiten.KeyF1
        case ActionCartographer: return ebiten.KeyF2
        case ActionApprentice: return ebiten.KeyF3
        case ActionHistorian: return ebiten.KeyF4
        case ActionAstrologer: return ebiten.KeyF5
        case ActionChancellor: return ebiten.KeyF6
        case ActionTaxCollector: return ebiten.KeyF7
        case ActionGrandVizier: return ebiten.KeyF8
        case ActionMirror: return ebiten.KeyF9
        case ActionNextTurn: return ebiten.KeyN
    }

    return Unbound
}

// Keybindings holds the current key bound to each action. Session-only,
// like every other setting in this codebase (e.g. music volume) - nothing
// here is persisted to disk.
type Keybindings struct {
    bindings map[Action]ebiten.Key
}

func MakeKeybindings() *Keybindings {
    bindings := make(map[Action]ebiten.Key)
    for _, action := range AllActions {
        bindings[action] = action.Default()
    }

    return &Keybindings{
        bindings: bindings,
    }
}

func (keybindings *Keybindings) Get(action Action) ebiten.Key {
    return keybindings.bindings[action]
}

func (keybindings *Keybindings) Set(action Action, key ebiten.Key) {
    keybindings.bindings[action] = key
}
