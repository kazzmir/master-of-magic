package load

import (
    "io"
    "log"
    "fmt"
    "bytes"
    "bufio"
    "errors"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/set"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
)

const NumHeroes = 35
const NumPlayers = 6
const NumPlayerHeroes = 6
const NumSpellsPerMagicRealm = 40
const NumMagicRealms = 6
const NumSpells = NumMagicRealms * NumSpellsPerMagicRealm

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
    NumPlayers int16
    LandSize int16
    Magic int16
    Difficulty int16
    NumCities int16
    NumUnits int16
    Turn int16
    Unit int16

    // first index player, second index hero
    HeroData [][]HeroData

    // indexed by player number
    PlayerData []PlayerData

    ArcanusMap TerrainData
    MyrrorMap TerrainData

    GrandVizier uint16

    UU_table_1 []byte
    UU_table_2 []byte

    ArcanusLandMasses [][]uint8
    MyrrorLandMasses [][]uint8

    Nodes []NodeData
    Fortresses []FortressData
    Towers []TowerData
    Lairs []LairData
    Items []ItemData
    Cities []CityData
    Units []UnitData

    ArcanusTerrainSpecials [][]uint8
    MyrrorTerrainSpecials [][]uint8

    ArcanusExplored [][]int8
    MyrrorExplored [][]int8

    ArcanusMovementCost MovementCostData
    MyrrorMovementCost MovementCostData

    Events EventData

    ArcanusMapSquareFlags [][]uint8
    MyrrorMapSquareFlags [][]uint8

    PremadeItems []byte

    HeroNames []HeroNameData
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

type PlayerHeroData struct {
    Unit int16
    Name string
    Items []int16
    ItemSlot []int16
}

func loadPlayerHeroData(reader io.Reader) (PlayerHeroData, error) {
    data := make([]byte, 28)
    _, err := io.ReadFull(reader, data)
    if err != nil {
        return PlayerHeroData{}, err
    }

    dataReader := bytes.NewReader(data)

    unit, err := lbx.ReadN[int16](dataReader)
    if err != nil {
        return PlayerHeroData{}, err
    }

    name := make([]byte, 14)
    _, err = io.ReadFull(dataReader, name)
    if err != nil {
        return PlayerHeroData{}, err
    }

    items := make([]int16, 3)
    for i := range len(items) {
        items[i], err = lbx.ReadN[int16](dataReader)
        if err != nil {
            return PlayerHeroData{}, err
        }
    }

    itemSlots := make([]int16, 3)
    for i := range len(itemSlots) {
        itemSlots[i], err = lbx.ReadN[int16](dataReader)
        if err != nil {
            return PlayerHeroData{}, err
        }
    }

    return PlayerHeroData{
        Unit: unit,
        Name: string(name),
        Items: items,
        ItemSlot: itemSlots,
    }, nil
}

type DiplomacyData struct {
    Contacted []int8
    TreatyInterest []int16
    PeaceInterest []int16
    TradeInterest []int16
    VisibleRelations []int8
    DiplomacyStatus []int8
    Strength []int16
    Action []int8
    Spell []int16
    City []int8
    DefaultRelations []int8
    ContactProgress []int8
    BrokenTreaty []int8
    HiddenRelations []int8
    TributeSpell []int8
    TributeGold []int16
    WarningProgress []int8
}

func loadDiplomacy(reader io.Reader) (DiplomacyData, error) {
    data := make([]byte, 306)
    _, err := io.ReadFull(reader, data)
    if err != nil {
        return DiplomacyData{}, fmt.Errorf("Unable to read diplomacy data: %v", err)
    }
    dataReader := bytes.NewReader(data)

    var out DiplomacyData

    out.Contacted, err = lbx.ReadArrayN[int8](dataReader, NumPlayers)
    if err != nil {
        return DiplomacyData{}, err
    }

    out.TreatyInterest, err = lbx.ReadArrayN[int16](dataReader, NumPlayers)
    if err != nil {
        return DiplomacyData{}, err
    }

    out.PeaceInterest, err = lbx.ReadArrayN[int16](dataReader, NumPlayers)
    if err != nil {
        return DiplomacyData{}, err
    }

    out.TradeInterest, err = lbx.ReadArrayN[int16](dataReader, NumPlayers)
    if err != nil {
        return DiplomacyData{}, err
    }

    out.VisibleRelations, err = lbx.ReadArrayN[int8](dataReader, NumPlayers)
    if err != nil {
        return DiplomacyData{}, err
    }

    out.DiplomacyStatus, err = lbx.ReadArrayN[int8](dataReader, NumPlayers)
    if err != nil {
        return DiplomacyData{}, err
    }

    out.Strength, err = lbx.ReadArrayN[int16](dataReader, NumPlayers)
    if err != nil {
        return DiplomacyData{}, err
    }

    out.Action, err = lbx.ReadArrayN[int8](dataReader, NumPlayers)
    if err != nil {
        return DiplomacyData{}, err
    }

    out.Spell, err = lbx.ReadArrayN[int16](dataReader, NumPlayers)
    if err != nil {
        return DiplomacyData{}, err
    }

    out.City, err = lbx.ReadArrayN[int8](dataReader, NumPlayers)
    if err != nil {
        return DiplomacyData{}, err
    }

    out.DefaultRelations, err = lbx.ReadArrayN[int8](dataReader, NumPlayers)
    if err != nil {
        return DiplomacyData{}, err
    }

    out.ContactProgress, err = lbx.ReadArrayN[int8](dataReader, NumPlayers)
    if err != nil {
        return DiplomacyData{}, err
    }

    out.BrokenTreaty, err = lbx.ReadArrayN[int8](dataReader, NumPlayers)
    if err != nil {
        return DiplomacyData{}, err
    }

    unknown1, err := lbx.ReadArrayN[int16](dataReader, NumPlayers)
    if err != nil {
        return DiplomacyData{}, err
    }

    out.HiddenRelations, err = lbx.ReadArrayN[int8](dataReader, NumPlayers)
    if err != nil {
        return DiplomacyData{}, err
    }

    unknown2, err := lbx.ReadArrayN[int8](dataReader, 24)
    if err != nil {
        return DiplomacyData{}, err
    }

    out.TributeSpell, err = lbx.ReadArrayN[int8](dataReader, NumPlayers)
    if err != nil {
        return DiplomacyData{}, err
    }

    unknown3, err := lbx.ReadArrayN[int8](dataReader, 90)
    if err != nil {
        return DiplomacyData{}, err
    }

    out.TributeGold, err = lbx.ReadArrayN[int16](dataReader, NumPlayers)
    if err != nil {
        return DiplomacyData{}, err
    }

    unknown4, err := lbx.ReadArrayN[int8](dataReader, 30)
    if err != nil {
        return DiplomacyData{}, err
    }

    unknown5, err := lbx.ReadArrayN[int8](dataReader, NumPlayers)
    if err != nil {
        return DiplomacyData{}, err
    }

    unknown6, err := lbx.ReadArrayN[int8](dataReader, NumPlayers)
    if err != nil {
        return DiplomacyData{}, err
    }

    out.WarningProgress, err = lbx.ReadArrayN[int8](dataReader, NumPlayers)
    if err != nil {
        return DiplomacyData{}, err
    }

    _ = unknown1
    _ = unknown2
    _ = unknown3
    _ = unknown4
    _ = unknown5
    _ = unknown6

    return out, nil
}

type AstrologyData struct {
    MagicPower int16
    SpellResearch int16
    ArmyStrength int16
}

func loadAstrology(reader io.Reader) (AstrologyData, error) {
    var out AstrologyData
    var err error

    out.MagicPower, err = lbx.ReadN[int16](reader)
    if err != nil {
        return AstrologyData{}, err
    }

    out.SpellResearch, err = lbx.ReadN[int16](reader)
    if err != nil {
        return AstrologyData{}, err
    }

    out.ArmyStrength, err = lbx.ReadN[int16](reader)
    if err != nil {
        return AstrologyData{}, err
    }

    return out, nil
}

type PlayerData struct {
    WizardId uint8
    WizardName []byte
    CapitalRace uint8
    BannerId uint8
    Personality uint16
    Objective uint16
    MasteryResearch uint16
    Fame uint16
    PowerBase uint16
    Volcanoes uint16
    ResearchRatio uint8
    ManaRatio uint8
    SkillRatio uint8
    VolcanoPower uint8
    SummonX int16
    SummonY int16
    SummonPlane int16
    ResearchSpells []uint16
    AverageUnitCost uint16
    CombatSkillLeft uint16
    CastingCostRemaining uint16
    CastingCostOriginal uint16
    CastingSpellIndex uint16
    SkillLeft uint16
    NominalSkill uint16
    TaxRate uint16
    SpellRanks []int16
    RetortAlchemy int8
    RetortWarlord int8
    RetortChaosMastery int8
    RetortNatureMastery int8
    RetortSorceryMastery int8
    RetortInfernalPower int8
    RetortDivinePower int8
    RetortSageMaster int8
    RetortChanneler int8
    RetortMyrran int8
    RetortArchmage int8
    RetortNodeMastery int8
    RetortManaFocusing int8
    RetortFamous int8
    RetortRunemaster int8
    RetortConjurer int8
    RetortCharismatic int8
    RetortArtificer int8

    VaultItems []int16
    Diplomacy DiplomacyData

    ResearchCostRemaining uint16
    ManaReserve uint16
    SpellCastingSkill int32
    ResearchingSpellIndex int16
    SpellsList []uint8
    DefeatedWizards uint16
    GoldReserve uint16
    Astrology AstrologyData
    Population uint16
    Historian []uint8
    GlobalEnchantments []uint8
    MagicStrategy uint16
    Hostility []uint16
    ReevaluateHostilityCountdown uint16
    ReevaluateMagicStrategyCountdown uint16
    ReevaluateMagicPowerCountdown uint16
    PeaceDuration []uint8
    TargetWizard uint16
    PrimaryRealm uint16
    SecondaryRealm uint16
}

func loadPlayerData(reader io.Reader) (PlayerData, error) {
    var err error
    var out PlayerData

    playerData := make([]byte, 1224)
    _, err = io.ReadFull(reader, playerData)
    if err != nil {
        return PlayerData{}, err
    }

    playerReader := &ReadMonitor{reader: bytes.NewReader(playerData)}

    out.WizardId, err = lbx.ReadN[uint8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.WizardName = make([]byte, 20)
    _, err = io.ReadFull(playerReader, out.WizardName)
    if err != nil {
        return PlayerData{}, err
    }

    // log.Printf("Player offset capitalRace: 0x%x", playerReader.BytesRead)

    out.CapitalRace, err = lbx.ReadN[uint8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.BannerId, err = lbx.ReadN[uint8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    _, err = lbx.ReadByte(playerReader) // skip 1 byte
    if err != nil {
        return PlayerData{}, err
    }

    out.Personality, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.Objective, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    _, err = lbx.ReadArrayN[uint8](playerReader, 6) // skip 6 bytes
    if err != nil {
        return PlayerData{}, err
    }

    // not sure
    out.MasteryResearch, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.Fame, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.PowerBase, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.Volcanoes, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.ResearchRatio, err = lbx.ReadN[uint8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.ManaRatio, err = lbx.ReadN[uint8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.SkillRatio, err = lbx.ReadN[uint8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.VolcanoPower, err = lbx.ReadN[uint8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.SummonX, err = lbx.ReadN[int16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.SummonY, err = lbx.ReadN[int16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.SummonPlane, err = lbx.ReadN[int16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.ResearchSpells, err = lbx.ReadArrayN[uint16](playerReader, 8)
    if err != nil {
        return PlayerData{}, err
    }

    _, err = lbx.ReadArrayN[uint8](playerReader, 4) // skip 4 byte
    if err != nil {
        return PlayerData{}, err
    }

    // log.Printf("Player offset averageUnitCost: 0x%x", playerReader.BytesRead)

    out.AverageUnitCost, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    _, err = lbx.ReadN[uint16](playerReader) // skip 2 byte
    if err != nil {
        return PlayerData{}, err
    }

    out.CombatSkillLeft, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.CastingCostRemaining, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.CastingCostOriginal, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.CastingSpellIndex, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.SkillLeft, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.NominalSkill, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.TaxRate, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.SpellRanks, err = lbx.ReadArrayN[int16](playerReader, 5)
    if err != nil {
        return PlayerData{}, err
    }

    // log.Printf("Player offset retorts: 0x%x", playerReader.BytesRead)

    out.RetortAlchemy, err = lbx.ReadN[int8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.RetortWarlord, err = lbx.ReadN[int8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.RetortChaosMastery, err = lbx.ReadN[int8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.RetortNatureMastery, err = lbx.ReadN[int8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.RetortSorceryMastery, err = lbx.ReadN[int8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.RetortInfernalPower, err = lbx.ReadN[int8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.RetortDivinePower, err = lbx.ReadN[int8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.RetortSageMaster, err = lbx.ReadN[int8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.RetortChanneler, err = lbx.ReadN[int8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.RetortMyrran, err = lbx.ReadN[int8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.RetortArchmage, err = lbx.ReadN[int8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.RetortManaFocusing, err = lbx.ReadN[int8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.RetortNodeMastery, err = lbx.ReadN[int8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.RetortFamous, err = lbx.ReadN[int8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.RetortRunemaster, err = lbx.ReadN[int8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.RetortConjurer, err = lbx.ReadN[int8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.RetortCharismatic, err = lbx.ReadN[int8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.RetortArtificer, err = lbx.ReadN[int8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    // log.Printf("Player offset heroes: 0x%x", playerReader.BytesRead)

    heroData := make([]PlayerHeroData, NumHeroes)

    for i := range NumPlayerHeroes {
        heroData[i], err = loadPlayerHeroData(playerReader)
        if err != nil {
            return PlayerData{}, err
        }
    }

    _, err = lbx.ReadN[int16](playerReader) // skip 2 bytes

    out.VaultItems, err = lbx.ReadArrayN[int16](playerReader, 4)
    if err != nil {
        return PlayerData{}, err
    }

    log.Printf("Vault items: %v", out.VaultItems)

    // log.Printf("Player offset diplomacy: 0x%x", playerReader.BytesRead)

    out.Diplomacy, err = loadDiplomacy(playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.ResearchCostRemaining, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.ManaReserve, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.SpellCastingSkill, err = lbx.ReadN[int32](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.ResearchingSpellIndex, err = lbx.ReadN[int16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.SpellsList, err = lbx.ReadArrayN[uint8](playerReader, NumSpells)
    if err != nil {
        return PlayerData{}, err
    }

    out.DefeatedWizards, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.GoldReserve, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    unknown1, err := lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.Astrology, err = loadAstrology(playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.Population, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.Historian, err = lbx.ReadArrayN[uint8](playerReader, 288)
    if err != nil {
        return PlayerData{}, err
    }

    out.GlobalEnchantments, err = lbx.ReadArrayN[uint8](playerReader, 24)
    if err != nil {
        return PlayerData{}, err
    }

    // log.Printf("Player offset magic strategy: 0x%x", playerReader.BytesRead)

    out.MagicStrategy, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    unknown2, err := lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.Hostility, err = lbx.ReadArrayN[uint16](playerReader, NumPlayers)
    if err != nil {
        return PlayerData{}, err
    }

    out.ReevaluateHostilityCountdown, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.ReevaluateMagicStrategyCountdown, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.ReevaluateMagicPowerCountdown, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.PeaceDuration, err = lbx.ReadArrayN[uint8](playerReader, NumPlayers)
    if err != nil {
        return PlayerData{}, err
    }

    unknown3, err := lbx.ReadN[uint8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    unknown4, err := lbx.ReadN[uint8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    unknown5, err := lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    // log.Printf("Player offset target wizard: 0x%x", playerReader.BytesRead)

    out.TargetWizard, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    unknown6, err := lbx.ReadN[uint8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    unknown7, err := lbx.ReadN[uint8](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    unknown8, err := lbx.ReadArrayN[uint8](playerReader, 6)
    if err != nil {
        return PlayerData{}, err
    }

    out.PrimaryRealm, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    out.SecondaryRealm, err = lbx.ReadN[uint16](playerReader)
    if err != nil {
        return PlayerData{}, err
    }

    _ = unknown1
    _ = unknown2
    _ = unknown3
    _ = unknown4
    _ = unknown5
    _ = unknown6
    _ = unknown7
    _ = unknown8

    return out, nil
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

    for y := range WorldHeight {
        for x := range WorldWidth {
            value, err := lbx.ReadUint16(reader)
            if err != nil {
                return TerrainData{}, err
            }
            data[x][y] = value
        }
    }

    return TerrainData{Data: data}, nil
}

func LoadLandMass(reader io.Reader) ([][]uint8, error) {
    data := make([][]uint8, WorldWidth)
    for i := range WorldWidth {
        data[i] = make([]uint8, WorldHeight)
    }

    for y := range WorldHeight {
        for x := range WorldWidth {
            value, err := lbx.ReadN[uint8](reader)
            if err != nil {
                return nil, err
            }
            data[x][y] = value
        }
    }

    return data, nil
}

type NodeType int8
const (
    NodeTypeSorcery NodeType = iota
    NodeTypeNature
    NodeTypeChaos
)

type NodeData struct {
    X int8
    Y int8
    Plane int8
    Owner int8
    Power int8
    AuraX []byte
    AuraY []byte
    NodeType int8
    Flags int8
}

func loadNodes(reader io.Reader) ([]NodeData, error) {
    var out []NodeData
    for range 30 {
        nodeData := make([]byte, 48)
        _, err := io.ReadFull(reader, nodeData)
        if err != nil {
            return nil, err
        }

        nodeReader := bytes.NewReader(nodeData)

        var data NodeData

        data.X, err = lbx.ReadN[int8](nodeReader)
        if err != nil {
            return nil, err
        }

        data.Y, err = lbx.ReadN[int8](nodeReader)
        if err != nil {
            return nil, err
        }

        data.Plane, err = lbx.ReadN[int8](nodeReader)
        if err != nil {
            return nil, err
        }

        data.Owner, err = lbx.ReadN[int8](nodeReader)
        if err != nil {
            return nil, err
        }

        data.Power, err = lbx.ReadN[int8](nodeReader)
        if err != nil {
            return nil, err
        }

        data.AuraX, err = lbx.ReadArrayN[byte](nodeReader, 20)
        if err != nil {
            return nil, err
        }

        data.AuraY, err = lbx.ReadArrayN[byte](nodeReader, 20)
        if err != nil {
            return nil, err
        }

        data.NodeType, err = lbx.ReadN[int8](nodeReader)
        if err != nil {
            return nil, err
        }

        data.Flags, err = lbx.ReadN[int8](nodeReader)
        if err != nil {
            return nil, err
        }

        out = append(out, data)

        // log.Printf("Node: x=%v y=%v plane=%v owner=%v power=%v auraX=%v auraY=%v nodeType=%v meld=%v", x, y, plane, owner, power, auraX, auraY, nodeType, meld)
    }

    return out, nil
}

type FortressData struct {
    X int8
    Y int8
    Plane int8
    Active int8
}

func loadFortresses(reader io.Reader) ([]FortressData, error) {
    var out []FortressData

    for range NumPlayers {
        fortressData := make([]byte, 4)
        _, err := io.ReadFull(reader, fortressData)
        if err != nil {
            return nil, err
        }

        fortressReader := bytes.NewReader(fortressData)
        var data FortressData

        data.X, err = lbx.ReadN[int8](fortressReader)
        if err != nil {
            return nil, err
        }

        data.Y, err = lbx.ReadN[int8](fortressReader)
        if err != nil {
            return nil, err
        }

        data.Plane, err = lbx.ReadN[int8](fortressReader)
        if err != nil {
            return nil, err
        }

        data.Active, err = lbx.ReadN[int8](fortressReader)
        if err != nil {
            return nil, err
        }

        // log.Printf("Fortress: x=%v y=%v plane=%v active=%v", x, y, plane, active)
        out = append(out, data)
    }

    return out, nil
}

type TowerData struct {
    X int8
    Y int8
    Owner int8
}

func loadTowers(reader io.Reader) ([]TowerData, error) {
    var out []TowerData
    for range 6 {
        towerData := make([]byte, 4)
        _, err := io.ReadFull(reader, towerData)
        if err != nil {
            return nil, err
        }

        towerReader := bytes.NewReader(towerData)

        var data TowerData

        data.X, err = lbx.ReadN[int8](towerReader)
        if err != nil {
            return nil, err
        }

        data.Y, err = lbx.ReadN[int8](towerReader)
        if err != nil {
            return nil, err
        }

        data.Owner, err = lbx.ReadN[int8](towerReader)
        if err != nil {
            return nil, err
        }

        // log.Printf("Tower: x=%v y=%v owner=%v", x, y, owner)
        out = append(out, data)
    }

    return out, nil
}

type LairData struct {
    X int8
    Y int8
    Plane int8
    Intact int8
    Kind int8
    Guard1_unit_type uint8
    Guard1_unit_count uint8
    Guard2_unit_type uint8
    Guard2_unit_count uint8
    Gold int16
    Mana int16
    SpellSpecial int8
    Flags uint8
    ItemCount int8
    Item1 int16
    Item2 int16
    Item3 int16
}

func loadLairs(reader io.Reader) ([]LairData, error) {
    var out []LairData
    for range 102 {
        lairData := make([]byte, 24)
        _, err := io.ReadFull(reader, lairData)
        if err != nil {
            return nil, err
        }

        lairReader := bytes.NewReader(lairData)

        var data LairData

        data.X, err = lbx.ReadN[int8](lairReader)
        if err != nil {
            return nil, err
        }

        data.Y, err = lbx.ReadN[int8](lairReader)
        if err != nil {
            return nil, err
        }

        data.Plane, err = lbx.ReadN[int8](lairReader)
        if err != nil {
            return nil, err
        }

        data.Intact, err = lbx.ReadN[int8](lairReader)
        if err != nil {
            return nil, err
        }

        data.Kind, err = lbx.ReadN[int8](lairReader)
        if err != nil {
            return nil, err
        }

        data.Guard1_unit_type, err = lbx.ReadN[uint8](lairReader)
        if err != nil {
            return nil, err
        }

        data.Guard1_unit_count, err = lbx.ReadN[uint8](lairReader)
        if err != nil {
            return nil, err
        }

        data.Guard2_unit_type, err = lbx.ReadN[uint8](lairReader)
        if err != nil {
            return nil, err
        }

        data.Guard2_unit_count, err = lbx.ReadN[uint8](lairReader)
        if err != nil {
            return nil, err
        }

        // 1 byte of padding
        _, err = lbx.ReadByte(lairReader)
        if err != nil {
            return nil, err
        }

        data.Gold, err = lbx.ReadN[int16](lairReader)
        if err != nil {
            return nil, err
        }

        data.Mana, err = lbx.ReadN[int16](lairReader)
        if err != nil {
            return nil, err
        }

        data.SpellSpecial, err = lbx.ReadN[int8](lairReader)
        if err != nil {
            return nil, err
        }

        data.Flags, err = lbx.ReadN[uint8](lairReader)
        if err != nil {
            return nil, err
        }

        data.ItemCount, err = lbx.ReadN[int8](lairReader)
        if err != nil {
            return nil, err
        }

        _, err = lbx.ReadByte(lairReader) // skip 1 byte
        if err != nil {
            return nil, err
        }

        data.Item1, err = lbx.ReadN[int16](lairReader)
        if err != nil {
            return nil, err
        }

        data.Item2, err = lbx.ReadN[int16](lairReader)
        if err != nil {
            return nil, err
        }

        data.Item3, err = lbx.ReadN[int16](lairReader)
        if err != nil {
            return nil, err
        }

        // log.Printf("lair x=%v y=%v plane=%v intact=%v kind=%v guard1_unit_type=%v guard1_unit_count=%v guard2_unit_type=%v guard2_unit_count=%v gold=%v mana=%v spellSpecial=%v flags=%v itemCount=%v item1=%v item2=%v item3=%v", x, y, plane, intact, kind, guard1_unit_type, guard1_unit_count, guard2_unit_type, guard2_unit_count, gold, mana, spellSpecial, flags, itemCount, item1, item2, item3)
        out = append(out, data)
    }

    return out, nil
}

type ItemData struct {
    Name []byte
    IconIndex uint16
    Slot byte
    Type byte
    Cost uint16
    Attack byte
    ToHit byte
    Defense byte
    Movement byte
    Resistance byte
    SpellSkill byte
    SpellSave byte
    Spell byte
    Charges uint16
    Abilities uint32
}

func loadItems(reader io.Reader) ([]ItemData, error) {
    var out []ItemData
    for range 138 {
        itemData := make([]byte, 50)
        _, err := io.ReadFull(reader, itemData)
        if err != nil {
            return nil, err
        }

        itemReader := bytes.NewReader(itemData)

        var data ItemData

        data.Name = make([]byte, 30)
        _, err = io.ReadFull(itemReader, data.Name)
        if err != nil {
            return nil, err
        }

        data.IconIndex, err = lbx.ReadN[uint16](itemReader)
        if err != nil {
            return nil, err
        }

        data.Slot, err = lbx.ReadN[byte](itemReader)
        if err != nil {
            return nil, err
        }

        data.Type, err = lbx.ReadN[byte](itemReader)
        if err != nil {
            return nil, err
        }

        data.Cost, err = lbx.ReadN[uint16](itemReader)
        if err != nil {
            return nil, err
        }

        data.Attack, err = lbx.ReadN[byte](itemReader)
        if err != nil {
            return nil, err
        }

        data.ToHit, err = lbx.ReadN[byte](itemReader)
        if err != nil {
            return nil, err
        }

        data.Defense, err = lbx.ReadN[byte](itemReader)
        if err != nil {
            return nil, err
        }

        data.Movement, err = lbx.ReadN[byte](itemReader)
        if err != nil {
            return nil, err
        }

        data.Resistance, err = lbx.ReadN[byte](itemReader)
        if err != nil {
            return nil, err
        }

        data.SpellSkill, err = lbx.ReadN[byte](itemReader)
        if err != nil {
            return nil, err
        }

        data.SpellSave, err = lbx.ReadN[byte](itemReader)
        if err != nil {
            return nil, err
        }

        data.Spell, err = lbx.ReadN[byte](itemReader)
        if err != nil {
            return nil, err
        }

        data.Charges, err = lbx.ReadN[uint16](itemReader)
        if err != nil {
            return nil, err
        }

        data.Abilities, err = lbx.ReadN[uint32](itemReader)
        if err != nil {
            return nil, err
        }

        out = append(out, data)

        // log.Printf("Item name=%v", string(name))
    }

    return out, nil
}

type CityData struct {
    Name []byte
    Race int8
    X int8
    Y int8
    Plane int8
    Owner int8
    Size int8
    Population int8
    Farmers int8
    SoldBuilding int8
    Population10 int16
    PlayerBits uint8
    Construction int16
    NumBuildings int8
    Buildings []byte
    Enchantments []byte
    ProductionUnits int8
    Production int16
    Gold uint8
    Upkeep int8
    ManaUpkeep int8
    Research int8
    Food int8
    RoadConnections []byte
}

func loadCities(reader io.Reader) ([]CityData, error) {
    var out []CityData
    for range 100 {
        cityData := make([]byte, 114)
        _, err := io.ReadFull(reader, cityData)
        if err != nil {
            return nil, err
        }

        cityReader := bytes.NewReader(cityData)

        var data CityData

        data.Name = make([]byte, 14)
        _, err = io.ReadFull(cityReader, data.Name)
        if err != nil {
            return nil, err
        }

        data.Race, err = lbx.ReadN[int8](cityReader)
        if err != nil {
            return nil, err
        }

        data.X, err = lbx.ReadN[int8](cityReader)
        if err != nil {
            return nil, err
        }

        data.Y, err = lbx.ReadN[int8](cityReader)
        if err != nil {
            return nil, err
        }

        data.Plane, err = lbx.ReadN[int8](cityReader)
        if err != nil {
            return nil, err
        }

        data.Owner, err = lbx.ReadN[int8](cityReader)
        if err != nil {
            return nil, err
        }

        data.Size, err = lbx.ReadN[int8](cityReader)
        if err != nil {
            return nil, err
        }

        data.Population, err = lbx.ReadN[int8](cityReader)
        if err != nil {
            return nil, err
        }

        data.Farmers, err = lbx.ReadN[int8](cityReader)
        if err != nil {
            return nil, err
        }

        data.SoldBuilding, err = lbx.ReadN[int8](cityReader)
        if err != nil {
            return nil, err
        }

        _, err = lbx.ReadByte(cityReader) // skip 1 byte

        data.Population10, err = lbx.ReadN[int16](cityReader)
        if err != nil {
            return nil, err
        }

        data.PlayerBits, err = lbx.ReadN[uint8](cityReader)
        if err != nil {
            return nil, err
        }

        _, err = lbx.ReadByte(cityReader) // skip 1 byte

        data.Construction, err = lbx.ReadN[int16](cityReader)
        if err != nil {
            return nil, err
        }

        data.NumBuildings, err = lbx.ReadN[int8](cityReader)
        if err != nil {
            return nil, err
        }

        data.Buildings = make([]byte, 36)
        _, err = io.ReadFull(cityReader, data.Buildings)
        if err != nil {
            return nil, err
        }

        data.Enchantments = make([]byte, 26)
        _, err = io.ReadFull(cityReader, data.Enchantments)
        if err != nil {
            return nil, err
        }

        data.ProductionUnits, err = lbx.ReadN[int8](cityReader)
        if err != nil {
            return nil, err
        }

        data.Production, err = lbx.ReadN[int16](cityReader)
        if err != nil {
            return nil, err
        }

        data.Gold, err = lbx.ReadN[uint8](cityReader)
        if err != nil {
            return nil, err
        }

        data.Upkeep, err = lbx.ReadN[int8](cityReader)
        if err != nil {
            return nil, err
        }

        data.ManaUpkeep, err = lbx.ReadN[int8](cityReader)
        if err != nil {
            return nil, err
        }

        data.Research, err = lbx.ReadN[int8](cityReader)
        if err != nil {
            return nil, err
        }

        data.Food, err = lbx.ReadN[int8](cityReader)
        if err != nil {
            return nil, err
        }

        data.RoadConnections = make([]byte, 13)
        _, err = io.ReadFull(cityReader, data.RoadConnections)
        if err != nil {
            return nil, err
        }

        /*
        _ = buildings
        _ = enchantments
        _ = roadConnections
        */

        // log.Printf("City name=%v race=%v x=%v y=%v plane=%v owner=%v size=%v population=%v farmers=%v soldBuilding=%v population10=%v playerBits=%v construction=%v numBuildings=%v buildings=%v enchantments=%v productionUnits=%v production=%v gold=%v upkeep=%v manaUpkeep=%v research=%v food=%v roadConnections=%v", string(name), race, x, y, plane, owner, size, population, farmers, soldBuilding, population10, playerBits, construction, numBuildings, buildings, enchantments, productionUnits, production, gold, upkeep, manaUpkeep, research, food, roadConnections)

        out = append(out, data)
    }

    return out, nil
}

type UnitData struct {
    X int8
    Y int8
    Plane int8
    Owner int8
    MovesMax int8
    TypeIndex uint8
    Hero int8
    Finished int8
    Moves int8
    DestinationX int8
    DestinationY int8
    Status int8
    Level int8
    Experience int16
    MoveFailed int8
    Damage int8
    DrawPriority int8
    InTower int16
    SightRange int8
    Mutations int8
    Enchantments uint32
    RoadTurns int8
    RoadX int8
    RoadY int8
}

func loadUnits(reader io.Reader) ([]UnitData, error) {
    var out []UnitData
    for range 1009 {
        unitData := make([]byte, 32)
        _, err := io.ReadFull(reader, unitData)
        if err != nil {
            return nil, err
        }

        unitReader := bytes.NewReader(unitData)

        var data UnitData

        data.X, err = lbx.ReadN[int8](unitReader)
        if err != nil {
            return nil, err
        }

        data.Y, err = lbx.ReadN[int8](unitReader)
        if err != nil {
            return nil, err
        }

        data.Plane, err = lbx.ReadN[int8](unitReader)
        if err != nil {
            return nil, err
        }

        data.Owner, err = lbx.ReadN[int8](unitReader)
        if err != nil {
            return nil, err
        }

        data.MovesMax, err = lbx.ReadN[int8](unitReader)
        if err != nil {
            return nil, err
        }

        data.TypeIndex, err = lbx.ReadN[uint8](unitReader)
        if err != nil {
            return nil, err
        }

        data.Hero, err = lbx.ReadN[int8](unitReader)
        if err != nil {
            return nil, err
        }

        data.Finished, err = lbx.ReadN[int8](unitReader)
        if err != nil {
            return nil, err
        }

        data.Moves, err = lbx.ReadN[int8](unitReader)
        if err != nil {
            return nil, err
        }

        data.DestinationX, err = lbx.ReadN[int8](unitReader)
        if err != nil {
            return nil, err
        }

        data.DestinationY, err = lbx.ReadN[int8](unitReader)
        if err != nil {
            return nil, err
        }

        data.Status, err = lbx.ReadN[int8](unitReader)
        if err != nil {
            return nil, err
        }

        data.Level, err = lbx.ReadN[int8](unitReader)
        if err != nil {
            return nil, err
        }

        data.Experience, err = lbx.ReadN[int16](unitReader)
        if err != nil {
            return nil, err
        }

        data.MoveFailed, err = lbx.ReadN[int8](unitReader)
        if err != nil {
            return nil, err
        }

        data.Damage, err = lbx.ReadN[int8](unitReader)
        if err != nil {
            return nil, err
        }

        data.DrawPriority, err = lbx.ReadN[int8](unitReader)
        if err != nil {
            return nil, err
        }

        data.InTower, err = lbx.ReadN[int16](unitReader)
        if err != nil {
            return nil, err
        }

        data.SightRange, err = lbx.ReadN[int8](unitReader)
        if err != nil {
            return nil, err
        }

        data.Mutations, err = lbx.ReadN[int8](unitReader)
        if err != nil {
            return nil, err
        }

        data.Enchantments, err = lbx.ReadN[uint32](unitReader)
        if err != nil {
            return nil, err
        }

        data.RoadTurns, err = lbx.ReadN[int8](unitReader)
        if err != nil {
            return nil, err
        }

        data.RoadX, err = lbx.ReadN[int8](unitReader)
        if err != nil {
            return nil, err
        }

        data.RoadY, err = lbx.ReadN[int8](unitReader)
        if err != nil {
            return nil, err
        }

        // log.Printf("Unit x=%v y=%v plane=%v owner=%v movesMax=%v typeIndex=%v hero=%v finished=%v moves=%v destinationX=%v destinationY=%v status=%v level=%v experience=%v moveFailed=%v damage=%v drawPriority=%v inTower=%v sightRange=%v mutations=%v enchantments=%v roadTurns=%v roadX=%v roadY=%v", x, y, plane, owner, movesMax, typeIndex, hero, finished, moves, destinationX, destinationY, status, level, experience, moveFailed, damage, drawPriority, inTower, sightRange, mutations, enchantments, roadTurns, roadX, roadY)

        out = append(out, data)
    }

    return out, nil
}

// T should usually be int8 or uint8
func loadMapData[T any](reader io.Reader) ([][]T, error) {
    data := make([][]T, WorldWidth)
    for i := range WorldWidth {
        data[i] = make([]T, WorldHeight)
    }

    for y := range WorldHeight {
        for x := range WorldWidth {
            value, err := lbx.ReadN[T](reader)
            if err != nil {
                return nil, err
            }
            data[x][y] = value
        }
    }

    return data, nil
}

type MovementCostData struct {
    Moves [][]int8
    Walking [][]int8
    Forester [][]int8
    Mountaineer [][]int8
    Swimming [][]int8
    Sailing [][]int8
}

func loadMovementCostData(reader io.Reader) (MovementCostData, error) {
    moves, err := loadMapData[int8](reader)
    if err != nil {
        return MovementCostData{}, err
    }

    walking, err := loadMapData[int8](reader)
    if err != nil {
        return MovementCostData{}, err
    }

    forester, err := loadMapData[int8](reader)
    if err != nil {
        return MovementCostData{}, err
    }

    mountaineer, err := loadMapData[int8](reader)
    if err != nil {
        return MovementCostData{}, err
    }

    swimming, err := loadMapData[int8](reader)
    if err != nil {
        return MovementCostData{}, err
    }

    sailing, err := loadMapData[int8](reader)
    if err != nil {
        return MovementCostData{}, err
    }

    return MovementCostData{
        Moves: moves,
        Walking: walking,
        Forester: forester,
        Mountaineer: mountaineer,
        Swimming: swimming,
        Sailing: sailing,
    }, nil
}

type EventData struct {
    LastEvent int16
    MeteorStatus int16
    MeteorPlayer int16
    MeteorData int16

    GiftStatus int16
    GiftPlayer int16
    GiftData int16

    DisjunctionStatus int16

    MarriageStatus int16
    MarriagePlayer int16
    MarriageNeutralCity int16
    MarriagePlayerCity int16

    EarthquakeStatus int16
    EarthquakePlayer int16
    EarthquakeData int16

    PirateStatus int16
    PiratePlayer int16
    PirateData int16

    PlagueStatus int16
    PlaguePlayer int16
    PlagueData int16
    PlagueDuration int16

    RebellionStatus int16
    RebellionPlayer int16
    RebellionData int16

    DonationStatus int16
    DonationPlayer int16
    DonationData int16

    DepletionStatus int16
    DepletionPlayer int16
    DepletionData int16

    MineralsStatus int16
    MineralsData int16
    MineralsPlayer int16

    PopulationBoomStatus int16
    // FIXME: are data and player swapped here?
    PopulationBoomData int16
    PopulationBoomPlayer int16
    PopulationBoomDuration int16

    GoodMoonStatus int16
    GoodMoonDuration int16

    BadMoonStatus int16
    BadMoonDuration int16

    ConjunctionChaosStatus int16
    ConjunctionChaosDuration int16

    ConjunctionNatureStatus int16
    ConjunctionNatureDuration int16

    ConjunctionSorceryStatus int16
    ConjunctionSorceryDuration int16

    ManaShortageStatus int16
    ManaShortageDuration int16
}

func loadEvents(reader io.Reader) (EventData, error) {
    data := make([]byte, 100)
    _, err := io.ReadFull(reader, data)
    if err != nil {
        return EventData{}, err
    }

    eventReader := bytes.NewReader(data)
    var eventData EventData

    eventData.LastEvent, _ = lbx.ReadN[int16](eventReader)
    eventData.MeteorStatus, _ = lbx.ReadN[int16](eventReader)
    eventData.MeteorPlayer, _ = lbx.ReadN[int16](eventReader)
    eventData.MeteorData, _ = lbx.ReadN[int16](eventReader)
    eventData.GiftStatus, _ = lbx.ReadN[int16](eventReader)
    eventData.GiftPlayer, _ = lbx.ReadN[int16](eventReader)
    eventData.GiftData, _ = lbx.ReadN[int16](eventReader)

    eventData.DisjunctionStatus, _ = lbx.ReadN[int16](eventReader)
    eventData.MarriageStatus, _ = lbx.ReadN[int16](eventReader)
    eventData.MarriagePlayer, _ = lbx.ReadN[int16](eventReader)
    eventData.MarriageNeutralCity, _ = lbx.ReadN[int16](eventReader)
    eventData.MarriagePlayerCity, _ = lbx.ReadN[int16](eventReader)
    eventData.EarthquakeStatus, _ = lbx.ReadN[int16](eventReader)
    eventData.EarthquakePlayer, _ = lbx.ReadN[int16](eventReader)
    eventData.EarthquakeData, _ = lbx.ReadN[int16](eventReader)
    eventData.PirateStatus, _ = lbx.ReadN[int16](eventReader)
    eventData.PiratePlayer, _ = lbx.ReadN[int16](eventReader)
    eventData.PirateData, _ = lbx.ReadN[int16](eventReader)
    eventData.PlagueStatus, _ = lbx.ReadN[int16](eventReader)
    eventData.PlaguePlayer, _ = lbx.ReadN[int16](eventReader)
    eventData.PlagueData, _ = lbx.ReadN[int16](eventReader)
    eventData.PlagueDuration, _ = lbx.ReadN[int16](eventReader)
    eventData.RebellionStatus, _ = lbx.ReadN[int16](eventReader)
    eventData.RebellionPlayer, _ = lbx.ReadN[int16](eventReader)
    eventData.RebellionData, _ = lbx.ReadN[int16](eventReader)
    eventData.DonationStatus, _ = lbx.ReadN[int16](eventReader)
    eventData.DonationPlayer, _ = lbx.ReadN[int16](eventReader)
    eventData.DonationData, _ = lbx.ReadN[int16](eventReader)
    eventData.DepletionStatus, _ = lbx.ReadN[int16](eventReader)
    eventData.DepletionPlayer, _ = lbx.ReadN[int16](eventReader)
    eventData.DepletionData, _ = lbx.ReadN[int16](eventReader)
    eventData.MineralsStatus, _ = lbx.ReadN[int16](eventReader)
    eventData.MineralsData, _ = lbx.ReadN[int16](eventReader)
    eventData.MineralsPlayer, _ = lbx.ReadN[int16](eventReader)
    eventData.PopulationBoomStatus, _ = lbx.ReadN[int16](eventReader)
    eventData.PopulationBoomData, _ = lbx.ReadN[int16](eventReader)
    eventData.PopulationBoomPlayer, _ = lbx.ReadN[int16](eventReader)
    eventData.PopulationBoomDuration, _ = lbx.ReadN[int16](eventReader)
    eventData.GoodMoonStatus, _ = lbx.ReadN[int16](eventReader)
    eventData.GoodMoonDuration, _ = lbx.ReadN[int16](eventReader)
    eventData.BadMoonStatus, _ = lbx.ReadN[int16](eventReader)
    eventData.BadMoonDuration, _ = lbx.ReadN[int16](eventReader)
    eventData.ConjunctionChaosStatus, _ = lbx.ReadN[int16](eventReader)
    eventData.ConjunctionChaosDuration, _ = lbx.ReadN[int16](eventReader)
    eventData.ConjunctionNatureStatus, _ = lbx.ReadN[int16](eventReader)
    eventData.ConjunctionNatureDuration, _ = lbx.ReadN[int16](eventReader)
    eventData.ConjunctionSorceryStatus, _ = lbx.ReadN[int16](eventReader)
    eventData.ConjunctionSorceryDuration, _ = lbx.ReadN[int16](eventReader)
    eventData.ManaShortageStatus, _ = lbx.ReadN[int16](eventReader)
    eventData.ManaShortageDuration, _ = lbx.ReadN[int16](eventReader)

    return eventData, nil
}

type HeroNameData struct {
    Name []byte
    Experience int16
}

// hero names and experience for the human player's heroes
func loadHeroNames(reader io.Reader) ([]HeroNameData, error) {
    var out []HeroNameData
    for range NumHeroes {
        name := make([]byte, 14)
        _, err := io.ReadFull(reader, name)
        if err != nil {
            return nil, err
        }

        experience, err := lbx.ReadN[int16](reader)

        out = append(out, HeroNameData{Name: name, Experience: experience})
    }

    return out, nil
}

func loadPremadeItems(reader io.Reader) ([]byte, error) {
    data := make([]byte, 250)
    _, err := io.ReadFull(reader, data)
    if err != nil {
        return nil, err
    }

    // I think data is just an array of 0's and 1's where a 1 at some index means that premade item is available
    // TODO: parse data

    return data, nil
}

type ReadMonitor struct {
    reader io.Reader
    BytesRead int
}

func (read *ReadMonitor) Read(p []byte) (n int, err error) {
    n, err = read.reader.Read(p)
    read.BytesRead += n
    return n, err
}

// load a dos savegame file
func LoadSaveGame(reader1 io.Reader) (*SaveGame, error) {
    var err error
    var saveGame SaveGame

    reader := &ReadMonitor{reader: bufio.NewReader(reader1)}

    var heroData [][]HeroData
    for range NumPlayers {
        var data []HeroData
        for range NumHeroes {
            heroData, err := loadHeroData(reader)
            if err != nil {
                return nil, err
            }
            /*
            if player == 0 {
                log.Printf("Read hero %v: %+v", HeroIndex(i).GetHero().GetUnit().Name, heroData)
            }
            */
            data = append(data, heroData)
        }

        heroData = append(heroData, data)
    }

    saveGame.HeroData = heroData

    saveGame.NumPlayers, err = lbx.ReadN[int16](reader)
    if err != nil {
        return nil, err
    }

    saveGame.LandSize, err = lbx.ReadN[int16](reader)
    if err != nil {
        return nil, err
    }

    saveGame.Magic, err = lbx.ReadN[int16](reader)
    if err != nil {
        return nil, err
    }

    saveGame.Difficulty, err = lbx.ReadN[int16](reader)
    if err != nil {
        return nil, err
    }

    saveGame.NumCities, err = lbx.ReadN[int16](reader)
    if err != nil {
        return nil, err
    }

    saveGame.NumUnits, err = lbx.ReadN[int16](reader)
    if err != nil {
        return nil, err
    }

    saveGame.Turn, err = lbx.ReadN[int16](reader)
    if err != nil {
        return nil, err
    }

    saveGame.Unit, err = lbx.ReadN[int16](reader)
    if err != nil {
        return nil, err
    }

    /*
    log.Printf("numPlayers: %v", numPlayers)
    log.Printf("landSize: %v", landSize)
    log.Printf("magic: %v", magic)
    log.Printf("difficulty: %v", difficulty)
    log.Printf("cities: %v", cities)
    log.Printf("units: %v", units)
    log.Printf("turn: %v", turn)
    log.Printf("unit: %v", unit)
    */

    saveGame.PlayerData = make([]PlayerData, NumPlayers)
    for i := range NumPlayers {
        saveGame.PlayerData[i], err = loadPlayerData(reader)
        if err != nil {
            return nil, err
        }
    }

    saveGame.ArcanusMap, err = LoadTerrain(reader)
    if err != nil {
        return nil, err
    }

    saveGame.MyrrorMap, err = LoadTerrain(reader)
    if err != nil {
        return nil, err
    }

    // FIXME: what is this for?
    saveGame.UU_table_1 = make([]byte, 96 * 2)
    _, err = io.ReadFull(reader, saveGame.UU_table_1)
    if err != nil {
        return nil, err
    }

    // FIXME: what is this for?
    saveGame.UU_table_2 = make([]byte, 96 * 2)
    _, err = io.ReadFull(reader, saveGame.UU_table_2)
    if err != nil {
        return nil, err
    }

    // log.Printf("Offset: 0x%x", reader.BytesRead)

    saveGame.ArcanusLandMasses, err = LoadLandMass(reader)
    if err != nil {
        return nil, err
    }

    saveGame.MyrrorLandMasses, err = LoadLandMass(reader)
    if err != nil {
        return nil, err
    }

    // log.Printf("Offset: 0x%x", reader.BytesRead)

    saveGame.Nodes, err = loadNodes(reader)
    if err != nil {
        return nil, err
    }

    // log.Printf("Offset: 0x%x", reader.BytesRead)

    saveGame.Fortresses, err = loadFortresses(reader)
    if err != nil {
        return nil, err
    }

    // log.Printf("Offset: 0x%x", reader.BytesRead)

    saveGame.Towers, err = loadTowers(reader)
    if err != nil {
        return nil, err
    }

    // log.Printf("Offset: 0x%x", reader.BytesRead)

    saveGame.Lairs, err = loadLairs(reader)
    if err != nil {
        return nil, err
    }

    // log.Printf("Offset: 0x%x", reader.BytesRead)

    saveGame.Items, err = loadItems(reader)
    if err != nil {
        return nil, err
    }

    saveGame.Cities, err = loadCities(reader)
    if err != nil {
        return nil, err
    }

    saveGame.Units, err = loadUnits(reader)
    if err != nil {
        return nil, err
    }

    saveGame.ArcanusTerrainSpecials, err = loadMapData[uint8](reader)
    if err != nil {
        return nil, err
    }

    saveGame.MyrrorTerrainSpecials, err = loadMapData[uint8](reader)
    if err != nil {
        return nil, err
    }

    saveGame.ArcanusExplored, err = loadMapData[int8](reader)
    if err != nil {
        return nil, err
    }

    saveGame.MyrrorExplored, err = loadMapData[int8](reader)
    if err != nil {
        return nil, err
    }

    saveGame.ArcanusMovementCost, err = loadMovementCostData(reader)
    if err != nil {
        return nil, err
    }

    saveGame.MyrrorMovementCost, err = loadMovementCostData(reader)
    if err != nil {
        return nil, err
    }

    // log.Printf("Offset: 0x%x", reader.BytesRead)

    saveGame.Events, err = loadEvents(reader)
    if err != nil {
        return nil, err
    }

    saveGame.ArcanusMapSquareFlags, err = loadMapData[uint8](reader)
    if err != nil {
        return nil, err
    }

    saveGame.MyrrorMapSquareFlags, err = loadMapData[uint8](reader)
    if err != nil {
        return nil, err
    }

    // log.Printf("Offset: 0x%x", reader.BytesRead)

    saveGame.GrandVizier, err = lbx.ReadN[uint16](reader)
    if err != nil {
        return nil, err
    }

    // log.Printf("Grand vizier: %v", grandVizier)

    // log.Printf("Offset: 0x%x", reader.BytesRead)

    saveGame.PremadeItems, err = loadPremadeItems(reader)
    if err != nil {
        return nil, err
    }

    // log.Printf("Offset: 0x%x", reader.BytesRead)

    saveGame.HeroNames, err = loadHeroNames(reader)
    if err != nil {
        return nil, err
    }

    _, err = lbx.ReadByte(reader)
    if !errors.Is(err, io.EOF) {
        log.Printf("leftover data")
    }

    return &saveGame, nil
}
