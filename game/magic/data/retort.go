package data

type Retort int
const (
    RetortAlchemy Retort = iota
    RetortWarlord
    RetortChanneler
    RetortArchmage
    RetortArtificer
    RetortConjurer
    RetortSageMaster
    RetortMyrran
    RetortDivinePower
    RetortFamous
    RetortRunemaster
    RetortCharismatic
    RetortChaosMastery
    RetortNatureMastery
    RetortSorceryMastery
    RetortInfernalPower
    RetortManaFocusing
    RetortNodeMastery
    RetortNone
)

func (retort Retort) MarshalJSON() ([]byte, error) {
    return []byte(`"` + retort.String() + `"`), nil
}

func (retort Retort) DependencyExplanation() string {
    switch retort {
        case RetortAlchemy: return ""
        case RetortWarlord: return ""
        case RetortChanneler: return ""
        case RetortArchmage: return "To select Archmage you need: 4 picks in any Realm of Magic"
        case RetortArtificer: return ""
        case RetortConjurer: return ""
        case RetortSageMaster: return "To select Sage Master you need: 1 pick in any 2 Realms of Magic"
        case RetortMyrran: return ""
        case RetortDivinePower: return "To select Divine Power you need: 4 picks in Life Magic"
        case RetortFamous: return ""
        case RetortRunemaster: return "To select Runemaster you need: 2 picks in any 3 Realms of Magic"
        case RetortCharismatic: return ""
        case RetortChaosMastery: return "To select Chaos Mastery you need: 4 picks in Chaos Magic"
        case RetortNatureMastery: return "To select Nature Mastery you need: 4 picks in Nature Magic"
        case RetortSorceryMastery: return "To select Sorcery Mastery you need: 4 picks in Sorcery Magic"
        case RetortInfernalPower: return "To select Infernal Power you need: 4 picks in Death Magic"
        case RetortManaFocusing: return "To select Mana Focusing you need: 4 picks in any Realm of Magic"
        case RetortNodeMastery: return "To select Node Mastery you need: 1 pick in Chaos Magic, 1 pick in Nature Magic, 1 pick in Sorcery Magic"
        case RetortNone: return ""
        default: return ""
    }
}

func (Retort Retort) String() string {
    switch Retort {
        case RetortAlchemy: return "Alchemy"
        case RetortWarlord: return "Warlord"
        case RetortChanneler: return "Channeler"
        case RetortArchmage: return "Archmage"
        case RetortArtificer: return "Artificer"
        case RetortConjurer: return "Conjurer"
        case RetortSageMaster: return "Sage Master"
        case RetortMyrran: return "Myrran"
        case RetortDivinePower: return "Divine Power"
        case RetortFamous: return "Famous"
        case RetortRunemaster: return "Runemaster"
        case RetortCharismatic: return "Charismatic"
        case RetortChaosMastery: return "Chaos Mastery"
        case RetortNatureMastery: return "Nature Mastery"
        case RetortSorceryMastery: return "Sorcery Mastery"
        case RetortInfernalPower: return "Infernal Power"
        case RetortManaFocusing: return "Mana Focusing"
        case RetortNodeMastery: return "Node Mastery"
        case RetortNone: return "invalid"
    }

    return "?"
}

// number of picks this retort costs when choosing a custom wizard
func (retort Retort) PickCost() int {
    switch retort {
        case RetortAlchemy: return 1
        case RetortWarlord: return 2
        case RetortChanneler: return 2
        case RetortArchmage: return 1
        case RetortArtificer: return 1
        case RetortConjurer: return 1
        case RetortSageMaster: return 1
        case RetortMyrran: return 3
        case RetortDivinePower: return 2
        case RetortFamous: return 2
        case RetortRunemaster: return 1
        case RetortCharismatic: return 1
        case RetortChaosMastery: return 1
        case RetortNatureMastery: return 1
        case RetortSorceryMastery: return 1
        case RetortInfernalPower: return 2
        case RetortManaFocusing: return 1
        case RetortNodeMastery: return 1
        case RetortNone: return 0
    }

    return 1
}
