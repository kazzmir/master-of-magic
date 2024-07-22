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
    // NEWGAME.LBX entry 8 contains boxes for wizard names
    // 9-23 are backgrounds for names
    // 24-26 are life books
    // 27-29 are sorcery/blue books
    // 30-32 are nature/green books
    // 33-35 are death books
    // 36-38 are chaos/red books
    // 41 is custom screen

}

func (screen *NewWizardScreen) Draw(window *ebiten.Image) {
}

func MakeNewWizardScreen() *NewWizardScreen {
    return &NewWizardScreen{}
}
