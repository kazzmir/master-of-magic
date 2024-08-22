package data

const ScreenWidth = 320
const ScreenHeight = 200

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

