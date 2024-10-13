package hero

import (
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
)

type Hero struct {
    Unit *units.OverworldUnit
    Title string

    Equipment [3]*artifact.Artifact
}

func (hero *Hero) PortraitIndex() int {
    if hero.Unit.Unit.Equals(units.HeroTorin) {
        return 0
    }

    if hero.Unit.Unit.Equals(units.HeroFang) {
        return 1
    }

    if hero.Unit.Unit.Equals(units.HeroBShan) {
        return 2
    }

    if hero.Unit.Unit.Equals(units.HeroMorgana) {
        return 3
    }

    if hero.Unit.Unit.Equals(units.HeroWarrax) {
        return 4
    }

    if hero.Unit.Unit.Equals(units.HeroMysticX) {
        return 5
    }

    if hero.Unit.Unit.Equals(units.HeroBahgtru) {
        return 6
    }

    if hero.Unit.Unit.Equals(units.HeroDethStryke) {
        return 7
    }

    if hero.Unit.Unit.Equals(units.HeroSpyder) {
        return 8
    }

    if hero.Unit.Unit.Equals(units.HeroSirHarold) {
        return 9
    }

    if hero.Unit.Unit.Equals(units.HeroBrax) {
        return 10
    }

    if hero.Unit.Unit.Equals(units.HeroRavashack) {
        return 11
    }

    if hero.Unit.Unit.Equals(units.HeroGreyfairer) {
        return 12
    }

    if hero.Unit.Unit.Equals(units.HeroShalla) {
        return 13
    }

    if hero.Unit.Unit.Equals(units.HeroRoland) {
        return 14
    }

    if hero.Unit.Unit.Equals(units.HeroMalleus) {
        return 15
    }

    if hero.Unit.Unit.Equals(units.HeroMortu) {
        return 16
    }

    if hero.Unit.Unit.Equals(units.HeroGunther) {
        return 17
    }

    if hero.Unit.Unit.Equals(units.HeroRakir) {
        return 18
    }

    if hero.Unit.Unit.Equals(units.HeroJaer) {
        return 19
    }

    if hero.Unit.Unit.Equals(units.HeroTaki) {
        return 20
    }

    if hero.Unit.Unit.Equals(units.HeroYramrag) {
        return 21
    }

    if hero.Unit.Unit.Equals(units.HeroValana) {
        return 22
    }

    if hero.Unit.Unit.Equals(units.HeroElana) {
        return 23
    }

    if hero.Unit.Unit.Equals(units.HeroAerie) {
        return 24
    }

    if hero.Unit.Unit.Equals(units.HeroMarcus) {
        return 25
    }

    if hero.Unit.Unit.Equals(units.HeroReywind) {
        return 26
    }

    if hero.Unit.Unit.Equals(units.HeroAlorra) {
        return 27
    }

    if hero.Unit.Unit.Equals(units.HeroZaldron) {
        return 28
    }

    if hero.Unit.Unit.Equals(units.HeroShinBo) {
        return 29
    }

    if hero.Unit.Unit.Equals(units.HeroSerena) {
        return 30
    }

    if hero.Unit.Unit.Equals(units.HeroShuri) {
        return 31
    }

    if hero.Unit.Unit.Equals(units.HeroTheria) {
        return 32
    }

    if hero.Unit.Unit.Equals(units.HeroTumu) {
        return 33
    }

    if hero.Unit.Unit.Equals(units.HeroAureus) {
        return 34
    }

    return -1
}
