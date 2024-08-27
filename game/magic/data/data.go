package data

import (
    _ "embed"
)

const ScreenWidth = 320
const ScreenHeight = 200

//go:embed data.zip
var DataZip []byte

type BannerType int
const (
    BannerGreen BannerType = iota
    BannerBlue
    BannerRed
    BannerPurple
    BannerYellow
    BannerBrown
)

func (banner BannerType) String() string {
    switch banner {
        case BannerGreen: return "green"
        case BannerBlue: return "blue"
        case BannerRed: return "red"
        case BannerPurple: return "purple"
        case BannerYellow: return "yellow"
        case BannerBrown: return "brown"
    }

    return ""
}

type Race int

const (
    RaceNone Race = iota
    RaceLizard
    RaceNomad
    RaceOrc
    RaceTroll
    RaceFantastic
    RaceHero
    RaceBarbarian
    RaceBeastmen
    RaceDarkElf
    RaceDraconian
    RaceDwarf
    RaceGnoll
    RaceHalfling
    RaceHighElf
    RaceHighMen
    RaceKlackon
)

func (race Race) String() string {
    switch race {
        case RaceNone: return "none"
        case RaceLizard: return "Lizardmen"
        case RaceNomad: return "Nomad"
        case RaceOrc: return "Orc"
        case RaceTroll: return "Troll"
        case RaceBarbarian: return "Barbarian"
        case RaceBeastmen: return "Beastmen"
        case RaceDarkElf: return "Dark Elf"
        case RaceDraconian: return "Draconian"
        case RaceDwarf: return "Dwarf"
        case RaceGnoll: return "Gnoll"
        case RaceHalfling: return "Halfling"
        case RaceHighElf: return "High Elf"
        case RaceHighMen: return "High Men"
        case RaceKlackon: return "Klackon"
    }

    return "?"
}

type Plane int

const (
    PlaneArcanus Plane = iota
    PlaneMyrror
)

