package main

import (
    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
)

type NewWizardScreen struct {
    Active bool
}

func (screen *NewWizardScreen) IsActive() bool {
    return screen.Active
}

func (screen *NewWizardScreen) Activate() {
    screen.Active = true
}

func (screen *NewWizardScreen) Deactivate() {
    screen.Active = false
}

func (screen *NewWizardScreen) Update() {
}

func (screen *NewWizardScreen) Load(cache *lbx.LbxCache) {
}

func (screen *NewWizardScreen) Draw(window *ebiten.Image) {
}

func MakeNewWizardScreen() *NewWizardScreen {
    return &NewWizardScreen{}
}
