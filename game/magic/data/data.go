package data

import (
    "fmt"
    "strconv"
    "image/color"
)

const ScreenWidth = 320
const ScreenHeight = 200

const MaxUnitsInStack = 9

type BannerType int
const (
    BannerGreen BannerType = iota
    BannerBlue
    BannerRed
    BannerPurple
    BannerYellow
    BannerBrown
)

func AllBanners() []BannerType {
    return []BannerType{
        BannerGreen,
        BannerBlue,
        BannerRed,
        BannerPurple,
        BannerYellow,
        BannerBrown,
    }
}

func (banner BannerType) MarshalJSON() ([]byte, error) {
    return []byte(fmt.Sprintf(`"%v"`, banner.String())), nil
}

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

func (banner BannerType) Color() color.RGBA {
    switch banner {
        case BannerGreen: return color.RGBA{R: 0x20, G: 0x80, B: 0x2c, A: 0xff}
        case BannerBlue: return color.RGBA{R: 0x15, G: 0x1d, B: 0x9d, A: 0xff}
        case BannerRed: return color.RGBA{R: 0x9d, G: 0x15, B: 0x15, A: 0xff}
        case BannerPurple: return color.RGBA{R: 0x6d, G: 0x15, B: 0x9d, A: 0xff}
        case BannerYellow: return color.RGBA{R: 0x9d, G: 0x9d, B: 0x15, A: 0xff}
        case BannerBrown: return color.RGBA{R: 0x82, G: 0x60, B: 0x12, A: 0xff}
    }

    return color.RGBA{}
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

func (race Race) MarshalJSON() ([]byte, error) {
    return []byte(fmt.Sprintf(`"%v"`, race.String())), nil
}

func (race *Race) UnmarshalJSON(data []byte) error {
    str, err := strconv.Unquote(string(data))
    if err != nil {
        return err
    }

    *race = getRaceByName(str)
    return nil
}

func getRaceByName(name string) Race {
    switch name {
        case "none": return RaceNone
        case "LizardMen": return RaceLizard
        case "Nomad": return RaceNomad
        case "Orc": return RaceOrc
        case "Troll": return RaceTroll
        case "Barbarian": return RaceBarbarian
        case "Beastmen": return RaceBeastmen
        case "Dark Elf": return RaceDarkElf
        case "Draconian": return RaceDraconian
        case "Dwarf": return RaceDwarf
        case "Gnoll": return RaceGnoll
        case "Halfling": return RaceHalfling
        case "High Elf": return RaceHighElf
        case "High Men": return RaceHighMen
        case "Klackon": return RaceKlackon
        case "Hero": return RaceHero
        case "Fantastic": return RaceFantastic
        case "All": return RaceAll
    }

    return RaceNone
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

type HouseType int

const (
    HouseTypeTree HouseType = iota
    HouseTypeHut
    HouseTypeNormal
)

func (race Race) HouseType() HouseType {
    switch race {
        case RaceDarkElf, RaceHighElf: return HouseTypeTree
        case RaceTroll, RaceGnoll, RaceKlackon, RaceLizard: return HouseTypeHut
        case RaceHighMen, RaceHalfling, RaceBarbarian, RaceDwarf,
             RaceNomad, RaceOrc, RaceDraconian, RaceBeastmen: return HouseTypeNormal

    }

    return HouseTypeNormal
}

type Plane int

const (
    PlaneArcanus Plane = iota
    PlaneMyrror
)

func (plane Plane) String() string {
    switch plane {
        case PlaneArcanus: return "Arcanus"
        case PlaneMyrror: return "Myrror"
    }

    return ""
}

func (plane Plane) Opposite() Plane {
    switch plane {
        case PlaneArcanus: return PlaneMyrror
        case PlaneMyrror: return PlaneArcanus
    }

    return PlaneArcanus
}

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
    // FIXME: Those constants are not equal to LBX ones. The proper order is Nature, Sorcery, Chaos, Life, Death, Arcane.
    MagicNone MagicType = iota
    LifeMagic
    SorceryMagic
    NatureMagic
    DeathMagic
    ChaosMagic
    ArcaneMagic
)

func (magic MagicType) MarshalJSON() ([]byte, error) {
    return []byte(fmt.Sprintf(`"%v"`, magic.String())), nil
}

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

func (bonus BonusType) String() string {
    switch bonus {
        case BonusGoldOre: return "Gold Ore"
        case BonusSilverOre: return "Silver Ore"
        case BonusWildGame: return "Wild Game"
        case BonusNightshade: return "Nightshade"
        case BonusIronOre: return "Iron Ore"
        case BonusCoal: return "Coal"
        case BonusMithrilOre: return "Mithril Ore"
        case BonusAdamantiumOre: return "Adamantium Ore"
        case BonusGem: return "Gem"
        case BonusQuorkCrystal: return "Quork Crystal"
        case BonusCrysxCrystal: return "Crysx Crystal"
    }

    return ""
}

func (bonus BonusType) LbxIndex() int {
    switch bonus {
        case BonusWildGame: return 92
        case BonusNightshade: return 91
        case BonusSilverOre: return 80
        case BonusGoldOre: return 81
        case BonusIronOre: return 78
        case BonusCoal: return 79
        case BonusMithrilOre: return 83
        case BonusAdamantiumOre: return 84
        case BonusGem: return 82
        case BonusQuorkCrystal: return 85
        case BonusCrysxCrystal: return 86
    }
    return -1
}

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

func (weapon WeaponBonus) String() string {
    switch weapon {
        case WeaponNone: return "none"
        case WeaponMagic: return "magic"
        case WeaponMythril: return "mythril"
        case WeaponAdamantium: return "adamantium"
    }

    return "none"
}

func getWeaponBonusByName(name string) WeaponBonus {
    switch name {
        case "none": return WeaponNone
        case "magic": return WeaponMagic
        case "mythril": return WeaponMythril
        case "adamantium": return WeaponAdamantium
    }

    return WeaponNone
}

func (weapon WeaponBonus) MarshalJSON() ([]byte, error) {
    return []byte(fmt.Sprintf(`"%v"`, weapon.String())), nil
}

func (weapon *WeaponBonus) UnmarshalJSON(data []byte) error {
    str, err := strconv.Unquote(string(data))
    if err != nil {
        return err
    }

    *weapon = getWeaponBonusByName(str)
    return nil
}

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

type FogType int

const (
    FogTypeUnexplored FogType = iota
    FogTypeExplored
    FogTypeVisible
)

type FogMap [][]FogType

func (fog FogMap) GetFog(x int, y int) FogType {
    if x < 0 || y < 0 || x >= len(fog) || y >= len(fog[x]) {
        return FogTypeUnexplored
    }

    return fog[x][y]
}

type PlanePoint struct {
    X int
    Y int
    Plane Plane
}
