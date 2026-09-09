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
    ActionQuitWithoutSaving
    ActionDefaultItemEditor
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
    ActionQuitWithoutSaving,
    ActionDefaultItemEditor,
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
        case ActionQuitWithoutSaving: return "Quit Without Saving"
        case ActionDefaultItemEditor: return "Default Item Editor"
    }

    return "Unknown"
}

// ID is the stable save-file name for this action. Display names in Name() can
// change; these IDs should not, so old saves keep resolving after a rename.
func (action Action) ID() string {
    switch action {
        case ActionGameScreen: return "game-screen"
        case ActionOpenSpellbook: return "open-spellbook"
        case ActionArmiesScreen: return "armies-screen"
        case ActionCitiesScreen: return "cities-screen"
        case ActionMagicScreen: return "magic-screen"
        case ActionAdvisors: return "advisors"
        case ActionSwitchPlanes: return "switch-planes"
        case ActionSurveyor: return "surveyor"
        case ActionCartographer: return "cartographer"
        case ActionApprentice: return "apprentice"
        case ActionHistorian: return "historian"
        case ActionAstrologer: return "astrologer"
        case ActionChancellor: return "chancellor"
        case ActionTaxCollector: return "tax-collector"
        case ActionGrandVizier: return "grand-vizier"
        case ActionMirror: return "mirror"
        case ActionNextTurn: return "next-turn"
        case ActionQuitWithoutSaving: return "quit-without-saving"
        case ActionDefaultItemEditor: return "default-item-editor"
    }

    return ""
}

func ActionByID(id string) (Action, bool) {
    if id == "" {
        return ActionGameScreen, false
    }

    for _, action := range AllActions {
        if action.ID() == id {
            return action, true
        }
    }

    return ActionGameScreen, false
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
        case ActionQuitWithoutSaving: return Unbound
        // remake default: original/CP left this unbound. E is free.
        case ActionDefaultItemEditor: return ebiten.KeyE
    }

    return Unbound
}

// Keybindings holds the current key bound to each action. Bindings are written
// into remake save files with the rest of Settings; music volume is still
// session-only.
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

// ConflictingActionForKey returns the action currently bound to key, or the
// zero-value ActionGameScreen and false if no action holds that key. It is the
// primitive the Keys screen consults before applying a rebind so a key can
// never end up driving two actions at once.
func (keybindings *Keybindings) ConflictingActionForKey(key ebiten.Key) (Action, bool) {
    if key == Unbound {
        return ActionGameScreen, false
    }

    for action, bound := range keybindings.bindings {
        if bound == key {
            return action, true
        }
    }

    return ActionGameScreen, false
}

func (keybindings *Keybindings) Set(action Action, key ebiten.Key) {
    keybindings.bindings[action] = key
}

// ResetToDefaults puts every action back to the original game's default key (see
// Action.Default), reusing the same map the actions already live in so the Keys
// screen can reset in place and the player reviews, or further tweaks, the
// bindings before closing the screen.
func (keybindings *Keybindings) ResetToDefaults() {
    for _, action := range AllActions {
        keybindings.bindings[action] = action.Default()
    }
}
