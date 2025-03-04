package load

import (
    "io"
    "log"
    "fmt"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/set"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
)

// hero indexes as dos mom defines them
type HeroIndex int
const (
    HeroBrax HeroIndex = iota
    HeroGunther
    HeroZaldron
    HeroBShan
    HeroRakir
    HeroValana
    HeroBahgtru
    HeroSerena
    HeroShuri
    HeroTheria
    HeroGreyfairer
    HeroTaki
    HeroReywind
    HeroMalleus
    HeroTumu
    HeroJaer
    HeroMarcus
    HeroFang
    HeroMorgana
    HeroAureus
    HeroShinBo
    HeroSpyder
    HeroShalla
    HeroYramrag
    HeroMysticX
    HeroAerie
    HeroDethStryke
    HeroElana
    HeroRoland
    HeroMortu
    HeroAlorra
    HeroSirHarold
    HeroRavashack
    HeroWarrax
    HeroTorin
)

func (heroIndex HeroIndex) GetHero() herolib.HeroType {
    switch heroIndex {
        case HeroBrax: return herolib.HeroBrax
        case HeroGunther: return herolib.HeroGunther
        case HeroZaldron: return herolib.HeroZaldron
        case HeroBShan: return herolib.HeroBShan
        case HeroRakir: return herolib.HeroRakir
        case HeroValana: return herolib.HeroValana
        case HeroBahgtru: return herolib.HeroBahgtru
        case HeroSerena: return herolib.HeroSerena
        case HeroShuri: return herolib.HeroShuri
        case HeroTheria: return herolib.HeroTheria
        case HeroGreyfairer: return herolib.HeroGreyfairer
        case HeroTaki: return herolib.HeroTaki
        case HeroReywind: return herolib.HeroReywind
        case HeroMalleus: return herolib.HeroMalleus
        case HeroTumu: return herolib.HeroTumu
        case HeroJaer: return herolib.HeroJaer
        case HeroMarcus: return herolib.HeroMarcus
        case HeroFang: return herolib.HeroFang
        case HeroMorgana: return herolib.HeroMorgana
        case HeroAureus: return herolib.HeroAureus
        case HeroShinBo: return herolib.HeroShinBo
        case HeroSpyder: return herolib.HeroSpyder
        case HeroShalla: return herolib.HeroShalla
        case HeroYramrag: return herolib.HeroYramrag
        case HeroMysticX: return herolib.HeroMysticX
        case HeroAerie: return herolib.HeroAerie
        case HeroDethStryke: return herolib.HeroDethStryke
        case HeroElana: return herolib.HeroElana
        case HeroRoland: return herolib.HeroRoland
        case HeroMortu: return herolib.HeroMortu
        case HeroAlorra: return herolib.HeroAlorra
        case HeroSirHarold: return herolib.HeroSirHarold
        case HeroRavashack: return herolib.HeroRavashack
        case HeroWarrax: return herolib.HeroWarrax
        case HeroTorin: return herolib.HeroTorin
    }

    return herolib.HeroBrax
}

type HeroAbility uint32

const (
    HeroAbility_LEADERSHIP HeroAbility = 0x00000001
    HeroAbility_LEADERSHIP2 HeroAbility = 0x00000002
    HeroAbility_LEGENDARY HeroAbility      = 0x00000008
    HeroAbility_LEGENDARY2 HeroAbility     = 0x00000010
    HeroAbility_BLADEMASTER HeroAbility    = 0x00000040
    HeroAbility_BLADEMASTER2 HeroAbility   = 0x00000080
    HeroAbility_ARMSMASTER HeroAbility     = 0x00000200
    HeroAbility_ARMSMASTER2 HeroAbility    = 0x00000400
    HeroAbility_CONSTITUTION HeroAbility   = 0x00001000
    HeroAbility_CONSTITUTION2 HeroAbility  = 0x00002000
    HeroAbility_MIGHT HeroAbility          = 0x00008000
    HeroAbility_MIGHT2 HeroAbility         = 0x00010000
    HeroAbility_ARCANE_POWER HeroAbility   = 0x00040000
    HeroAbility_ARCANE_POWER2 HeroAbility  = 0x00080000
    HeroAbility_SAGE HeroAbility           = 0x00200000
    HeroAbility_SAGE2 HeroAbility          = 0x00400000
    HeroAbility_PRAYERMASTER HeroAbility   = 0x01000000
    HeroAbility_PRAYERMASTER2 HeroAbility  = 0x02000000
    HeroAbility_AGILITY2 HeroAbility       = 0x04000000
    HeroAbility_LUCKY HeroAbility          = 0x08000000
    HeroAbility_CHARMED HeroAbility        = 0x10000000
    HeroAbility_NOBLE HeroAbility          = 0x20000000
    HeroAbility_FEMALE HeroAbility         = 0x40000000
    HeroAbility_AGILITY HeroAbility        = 0x80000000
)

func (ability HeroAbility) String() string {
    switch ability {
        case HeroAbility_LEADERSHIP: return "Leadership"
        case HeroAbility_LEADERSHIP2: return "Leadership2"
        case HeroAbility_LEGENDARY: return "Legendary"
        case HeroAbility_LEGENDARY2: return "Legendary2"
        case HeroAbility_BLADEMASTER: return "Blademaster"
        case HeroAbility_BLADEMASTER2: return "Blademaster2"
        case HeroAbility_ARMSMASTER: return "Armsmaster"
        case HeroAbility_ARMSMASTER2: return "Armsmaster2"
        case HeroAbility_CONSTITUTION: return "Constitution"
        case HeroAbility_CONSTITUTION2: return "Constitution2"
        case HeroAbility_MIGHT: return "Might"
        case HeroAbility_MIGHT2: return "Might2"
        case HeroAbility_ARCANE_POWER: return "Arcane Power"
        case HeroAbility_ARCANE_POWER2: return "Arcane Power2"
        case HeroAbility_SAGE: return "Sage"
        case HeroAbility_SAGE2: return "Sage2"
        case HeroAbility_PRAYERMASTER: return "Prayermaster"
        case HeroAbility_PRAYERMASTER2: return "Prayermaster2"
        case HeroAbility_AGILITY2: return "Agility2"
        case HeroAbility_LUCKY: return "Lucky"
        case HeroAbility_CHARMED: return "Charmed"
        case HeroAbility_NOBLE: return "Noble"
        case HeroAbility_FEMALE: return "Female"
        case HeroAbility_AGILITY: return "Agility"
    }

    return ""
}

type HeroData struct {
    Level int16
    Abilities uint32
    AbilitySet *set.Set[HeroAbility]
    CastingSkill int8
    Spells [4]uint8
}

func makeHeroAbilities(abilities uint32) *set.Set[HeroAbility] {
    out := set.NewSet[HeroAbility]()

    all := []HeroAbility{
        HeroAbility_LEADERSHIP,
        HeroAbility_LEADERSHIP2,
        HeroAbility_LEGENDARY,
        HeroAbility_LEGENDARY2,
        HeroAbility_BLADEMASTER,
        HeroAbility_BLADEMASTER2,
        HeroAbility_ARMSMASTER,
        HeroAbility_ARMSMASTER2,
        HeroAbility_CONSTITUTION,
        HeroAbility_CONSTITUTION2,
        HeroAbility_MIGHT,
        HeroAbility_MIGHT2,
        HeroAbility_ARCANE_POWER,
        HeroAbility_ARCANE_POWER2,
        HeroAbility_SAGE,
        HeroAbility_SAGE2,
        HeroAbility_PRAYERMASTER,
        HeroAbility_PRAYERMASTER2,
        HeroAbility_AGILITY2,
        HeroAbility_LUCKY,
        HeroAbility_CHARMED,
        HeroAbility_NOBLE,
        HeroAbility_FEMALE,
        HeroAbility_AGILITY,
    }

    for _, ability := range all {
        if abilities & uint32(ability) != 0 {
            out.Insert(ability)
        }
    }

    return out
}

type SaveGame struct {
}

func loadHeroData(reader io.Reader) (HeroData, error) {
    var heroData HeroData
    var err error

    heroData.Level, err = lbx.ReadN[int16](reader)
    if err != nil {
        return HeroData{}, err
    }

    heroData.Abilities, err = lbx.ReadN[uint32](reader)
    if err != nil {
        return HeroData{}, err
    }

    heroData.AbilitySet = makeHeroAbilities(heroData.Abilities)

    heroData.CastingSkill, err = lbx.ReadN[int8](reader)
    if err != nil {
        return HeroData{}, err
    }

    for i := range 4 {
        heroData.Spells[i], err = lbx.ReadN[uint8](reader)
        if err != nil {
            return HeroData{}, err
        }
    }

    _, err = lbx.ReadByte(reader) // skip 1 byte
    if err != nil {
        return HeroData{}, err
    }

    return heroData, nil
}

const (
    WorldWidth = 60
    WorldHeight = 40
)

type TerrainData struct {
    Data [][]uint16
}

func LoadTerrain(reader io.Reader) (TerrainData, error) {
    data := make([][]uint16, WorldWidth)
    for i := range WorldWidth {
        data[i] = make([]uint16, WorldHeight)
    }

    for y := range(WorldHeight) {
        for x := range(WorldWidth) {
            value, err := lbx.ReadUint16(reader)
            if err != nil {
                return TerrainData{}, err
            }
            data[x][y] = value
        }
    }

    return TerrainData{Data: data}, nil
}

// load a dos savegame file
func LoadSaveGame(reader io.Reader) (*SaveGame, error) {
    numHeroes := 35

    for player := range 6 {
        for i := range numHeroes {
            heroData, err := loadHeroData(reader)
            if err != nil {
                return nil, err
            }
            if player == 0 {
                log.Printf("Read hero %v: %+v", HeroIndex(i).GetHero().GetUnit().Name, heroData)
            }
        }
    }

    numPlayers, err := lbx.ReadN[int16](reader)
    if err != nil {
        return nil, err
    }

    landSize, err := lbx.ReadN[int16](reader)
    if err != nil {
        return nil, err
    }

    magic, err := lbx.ReadN[int16](reader)
    if err != nil {
        return nil, err
    }

    difficulty, err := lbx.ReadN[int16](reader)
    if err != nil {
        return nil, err
    }

    cities, err := lbx.ReadN[int16](reader)
    if err != nil {
        return nil, err
    }

    units, err := lbx.ReadN[int16](reader)
    if err != nil {
        return nil, err
    }

    turn, err := lbx.ReadN[int16](reader)
    if err != nil {
        return nil, err
    }

    unit, err := lbx.ReadN[int16](reader)
    if err != nil {
        return nil, err
    }

    log.Printf("numPlayers: %v", numPlayers)
    log.Printf("landSize: %v", landSize)
    log.Printf("magic: %v", magic)
    log.Printf("difficulty: %v", difficulty)
    log.Printf("cities: %v", cities)
    log.Printf("units: %v", units)
    log.Printf("turn: %v", turn)
    log.Printf("unit: %v", unit)

    // FIXME: LoadTerrain

    return nil, fmt.Errorf("unfinished")
}
