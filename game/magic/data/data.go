package data

var ScreenScale = 2
var ScreenWidth = 320*ScreenScale
var ScreenHeight = 200*ScreenScale
var ScreenScaleAlgorithm = ScaleAlgorithmScale

type ScaleAlgorithm int

const (
    // the scale2x
    // https://www.scale2x.it/
    ScaleAlgorithmScale ScaleAlgorithm = iota
    ScaleAlgorithmXbr
    ScaleAlgorithmNormal
)

func (algorithm ScaleAlgorithm) String() string {
    switch algorithm {
        case ScaleAlgorithmScale: return "scale"
        case ScaleAlgorithmXbr: return "xbr"
        case ScaleAlgorithmNormal: return "normal"
    }

    return ""
}

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
    RaceAll
)

func ArcanianRaces() []Race {
    return []Race{
        RaceBarbarian,
        RaceGnoll,
        RaceHalfling,
        RaceHighElf,
        RaceHighMen,
        RaceKlackon,
        RaceLizard,
        RaceNomad,
        RaceOrc,
    }
}

func MyrranRaces() []Race {
    return []Race{
        RaceBeastmen,
        RaceDarkElf,
        RaceDraconian,
        RaceDwarf,
        RaceTroll,
    }
}

// technically 'Lizardmen' should be 'Lizardman' and 'Dwarf' should be 'Dwarven', but the help has them listed as
// 'Lizardmen Townsfolk' and 'Dwarf Townsfolk'
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
        case RaceHero: return "Hero"
        case RaceFantastic: return "Fantastic"
        case RaceAll: return "All"
    }

    return "?"
}

type Plane int

const (
    PlaneArcanus Plane = iota
    PlaneMyrror
)

type WizardBase int

const (
    WizardMerlin WizardBase = iota
    WizardRaven
    WizardSharee
    WizardLoPan
    WizardJafar
    WizardOberic
    WizardRjak
    WizardSssra
    WizardTauron
    WizardFreya
    WizardHorus
    WizardAriel
    WizardTlaloc
    WizardKali
)

type MagicType int

const (
    MagicNone MagicType = iota
    LifeMagic
    SorceryMagic
    NatureMagic
    DeathMagic
    ChaosMagic
    ArcaneMagic
)

func (magic MagicType) String() string {
    switch magic {
        case LifeMagic: return "Life"
        case SorceryMagic: return "Sorcery"
        case NatureMagic: return "Nature"
        case DeathMagic: return "Death"
        case ChaosMagic: return "Chaos"
        case ArcaneMagic: return "Arcane"
    }

    return ""
}

/* the number of books a wizard has of a specific magic type */
type WizardBook struct {
    Magic MagicType
    Count int
}

type MagicSetting int
const (
    MagicSettingWeak MagicSetting = iota
    MagicSettingNormal
    MagicSettingPowerful
)

type DifficultySetting int
const (
    DifficultyIntro DifficultySetting = iota
    DifficultyEasy
    DifficultyAverage
    DifficultyHard
    DifficultyExtreme
    DifficultyImpossible
)

type BonusType int
const (
    BonusNone BonusType = iota
    BonusWildGame
    BonusNightshade
    BonusSilverOre
    BonusGoldOre
    BonusIronOre
    BonusCoal
    BonusMithrilOre
    BonusAdamantiumOre
    BonusGem
    BonusQuorkCrystal
    BonusCrysxCrystal
)

func (bonus BonusType) FoodBonus() int {
    if bonus == BonusWildGame {
        return 2
    }

    return 0
}

func (bonus BonusType) GoldBonus() int {
    switch bonus {
        case BonusSilverOre: return 2
        case BonusGoldOre: return 3
        case BonusGem: return 5
        default: return 0
    }
}

func (bonus BonusType) PowerBonus() int {
    switch bonus {
        case BonusMithrilOre: return 1
        case BonusAdamantiumOre: return 2
        case BonusQuorkCrystal: return 3
        case BonusCrysxCrystal: return 5
        default: return 0
    }
}

// returns a percent that unit costs are reduced by, 10 -> -10%
func (bonus BonusType) UnitReductionBonus() int {
    switch bonus {
        case BonusIronOre: return 5
        case BonusCoal: return 10
        default: return 0
    }
}

type WeaponBonus int
const (
    WeaponNone WeaponBonus = iota
    WeaponMagic
    WeaponMythril
    WeaponAdamantium
)

type TreatyType int
const (
    TreatyNone TreatyType = iota
    TreatyPact
    TreatyAlliance
    TreatyWar
)

func (treaty TreatyType) String() string {
    switch treaty {
        case TreatyNone: return "none"
        case TreatyPact: return "pact"
        case TreatyAlliance: return "alliance"
        case TreatyWar: return "war"
    }

    return ""
}
