package load

import (
    "io"
    "log"
    "fmt"
    "bytes"

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

func loadPlayerData(reader io.Reader) error {
    var err error
    playerData := make([]byte, 1224)
    _, err = io.ReadFull(reader, playerData)
    if err != nil {
        return err
    }

    playerReader := bytes.NewReader(playerData)

    wizardId, err := lbx.ReadN[uint8](playerReader)
    if err != nil {
        return err
    }

    wizardName := make([]byte, 20)
    _, err = io.ReadFull(playerReader, wizardName)
    if err != nil {
        return err
    }

    // FIXME: read more fields

    log.Printf("Player %v: %v", wizardId, string(wizardName))

    return nil
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

func loadNodes(reader io.Reader) error {
    for range 30 {
        nodeData := make([]byte, 48)
        _, err := io.ReadFull(reader, nodeData)
        if err != nil {
            return err
        }

        nodeReader := bytes.NewReader(nodeData)

        x, err := lbx.ReadN[int8](nodeReader)
        if err != nil {
            return err
        }

        y, err := lbx.ReadN[int8](nodeReader)
        if err != nil {
            return err
        }

        plane, err := lbx.ReadN[int8](nodeReader)
        if err != nil {
            return err
        }

        owner, err := lbx.ReadN[int8](nodeReader)
        if err != nil {
            return err
        }

        power, err := lbx.ReadN[int8](nodeReader)
        if err != nil {
            return err
        }

        auraX := make([]byte, 20)
        _, err = io.ReadFull(nodeReader, auraX)
        if err != nil {
            return err
        }

        auraY := make([]byte, 20)
        _, err = io.ReadFull(nodeReader, auraY)
        if err != nil {
            return err
        }

        nodeType, err := lbx.ReadN[int8](nodeReader)
        if err != nil {
            return err
        }

        meld, err := lbx.ReadN[int8](nodeReader)
        if err != nil {
            return err
        }

        log.Printf("Node: x=%v y=%v plane=%v owner=%v power=%v auraX=%v auraY=%v nodeType=%v meld=%v", x, y, plane, owner, power, auraX, auraY, nodeType, meld)
    }

    return nil
}

func loadFortresses(reader io.Reader) error {
    for range 6 {
        fortressData := make([]byte, 4)
        _, err := io.ReadFull(reader, fortressData)
        if err != nil {
            return err
        }

        fortressReader := bytes.NewReader(fortressData)
        x, err := lbx.ReadN[int8](fortressReader)
        if err != nil {
            return err
        }

        y, err := lbx.ReadN[int8](fortressReader)
        if err != nil {
            return err
        }

        plane, err := lbx.ReadN[int8](fortressReader)
        if err != nil {
            return err
        }

        active, err := lbx.ReadN[int8](fortressReader)
        if err != nil {
            return err
        }

        log.Printf("Fortress: x=%v y=%v plane=%v active=%v", x, y, plane, active)
    }

    return nil
}

func loadTowers(reader io.Reader) error {
    for range 6 {
        towerData := make([]byte, 4)
        _, err := io.ReadFull(reader, towerData)
        if err != nil {
            return err
        }

        towerReader := bytes.NewReader(towerData)

        x, err := lbx.ReadN[int8](towerReader)
        if err != nil {
            return err
        }

        y, err := lbx.ReadN[int8](towerReader)
        if err != nil {
            return err
        }

        owner, err := lbx.ReadN[int8](towerReader)
        if err != nil {
            return err
        }

        log.Printf("Tower: x=%v y=%v owner=%v", x, y, owner)
    }

    return nil
}

func loadLairs(reader io.Reader) error {
    for range 102 {
        lairData := make([]byte, 24)
        _, err := io.ReadFull(reader, lairData)
        if err != nil {
            return err
        }

        lairReader := bytes.NewReader(lairData)

        x, err := lbx.ReadN[int8](lairReader)
        if err != nil {
            return err
        }

        y, err := lbx.ReadN[int8](lairReader)
        if err != nil {
            return err
        }

        plane, err := lbx.ReadN[int8](lairReader)
        if err != nil {
            return err
        }

        intact, err := lbx.ReadN[int8](lairReader)
        if err != nil {
            return err
        }

        kind, err := lbx.ReadN[int8](lairReader)
        if err != nil {
            return err
        }

        guard1_unit_type, err := lbx.ReadN[uint8](lairReader)
        if err != nil {
            return err
        }

        guard1_unit_count, err := lbx.ReadN[uint8](lairReader)
        if err != nil {
            return err
        }

        guard2_unit_type, err := lbx.ReadN[uint8](lairReader)
        if err != nil {
            return err
        }

        guard2_unit_count, err := lbx.ReadN[uint8](lairReader)
        if err != nil {
            return err
        }

        // 1 byte of padding
        _, err = lbx.ReadByte(lairReader)
        if err != nil {
            return err
        }

        gold, err := lbx.ReadN[int16](lairReader)
        if err != nil {
            return err
        }

        mana, err := lbx.ReadN[int16](lairReader)
        if err != nil {
            return err
        }

        spellSpecial, err := lbx.ReadN[int8](lairReader)
        if err != nil {
            return err
        }

        flags, err := lbx.ReadN[uint8](lairReader)
        if err != nil {
            return err
        }

        itemCount, err := lbx.ReadN[int8](lairReader)
        if err != nil {
            return err
        }

        _, err = lbx.ReadByte(lairReader) // skip 1 byte
        if err != nil {
            return err
        }

        item1, err := lbx.ReadN[int16](lairReader)
        if err != nil {
            return err
        }

        item2, err := lbx.ReadN[int16](lairReader)
        if err != nil {
            return err
        }

        item3, err := lbx.ReadN[int16](lairReader)
        if err != nil {
            return err
        }

        log.Printf("lair x=%v y=%v plane=%v intact=%v kind=%v guard1_unit_type=%v guard1_unit_count=%v guard2_unit_type=%v guard2_unit_count=%v gold=%v mana=%v spellSpecial=%v flags=%v itemCount=%v item1=%v item2=%v item3=%v", x, y, plane, intact, kind, guard1_unit_type, guard1_unit_count, guard2_unit_type, guard2_unit_count, gold, mana, spellSpecial, flags, itemCount, item1, item2, item3)
    }

    return nil
}

func loadItems(reader io.Reader) error {
    for range 138 {
        itemData := make([]byte, 50)
        _, err := io.ReadFull(reader, itemData)
        if err != nil {
            return err
        }

        itemReader := bytes.NewReader(itemData)
        name := make([]byte, 30)
        _, err = io.ReadFull(itemReader, name)
        if err != nil {
            return err
        }

        log.Printf("Item name=%v", string(name))
    }

    return nil
}

func loadCities(reader io.Reader) error {
    for range 100 {
        cityData := make([]byte, 114)
        _, err := io.ReadFull(reader, cityData)
        if err != nil {
            return err
        }

        cityReader := bytes.NewReader(cityData)

        name := make([]byte, 14)
        _, err = io.ReadFull(cityReader, name)
        if err != nil {
            return err
        }

        race, err := lbx.ReadN[int8](cityReader)
        if err != nil {
            return err
        }

        x, err := lbx.ReadN[int8](cityReader)
        if err != nil {
            return err
        }

        y, err := lbx.ReadN[int8](cityReader)
        if err != nil {
            return err
        }

        plane, err := lbx.ReadN[int8](cityReader)
        if err != nil {
            return err
        }

        owner, err := lbx.ReadN[int8](cityReader)
        if err != nil {
            return err
        }

        size, err := lbx.ReadN[int8](cityReader)
        if err != nil {
            return err
        }

        population, err := lbx.ReadN[int8](cityReader)
        if err != nil {
            return err
        }

        farmers, err := lbx.ReadN[int8](cityReader)
        if err != nil {
            return err
        }

        soldBuilding, err := lbx.ReadN[int8](cityReader)
        if err != nil {
            return err
        }

        _, err = lbx.ReadByte(cityReader) // skip 1 byte

        population10, err := lbx.ReadN[int16](cityReader)
        if err != nil {
            return err
        }

        playerBits, err := lbx.ReadN[uint8](cityReader)
        if err != nil {
            return err
        }

        _, err = lbx.ReadByte(cityReader) // skip 1 byte

        construction, err := lbx.ReadN[int16](cityReader)
        if err != nil {
            return err
        }

        numBuildings, err := lbx.ReadN[int8](cityReader)
        if err != nil {
            return err
        }

        buildings := make([]byte, 36)
        _, err = io.ReadFull(cityReader, buildings)
        if err != nil {
            return err
        }

        enchantments := make([]byte, 26)
        _, err = io.ReadFull(cityReader, enchantments)
        if err != nil {
            return err
        }

        productionUnits, err := lbx.ReadN[int8](cityReader)
        if err != nil {
            return err
        }

        production, err := lbx.ReadN[int16](cityReader)
        if err != nil {
            return err
        }

        gold, err := lbx.ReadN[uint8](cityReader)
        if err != nil {
            return err
        }

        upkeep, err := lbx.ReadN[int8](cityReader)
        if err != nil {
            return err
        }

        manaUpkeep, err := lbx.ReadN[int8](cityReader)
        if err != nil {
            return err
        }

        research, err := lbx.ReadN[int8](cityReader)
        if err != nil {
            return err
        }

        food, err := lbx.ReadN[int8](cityReader)
        if err != nil {
            return err
        }

        roadConnections := make([]byte, 13)
        _, err = io.ReadFull(cityReader, roadConnections)
        if err != nil {
            return err
        }

        _ = buildings
        _ = enchantments
        _ = roadConnections

        log.Printf("City name=%v race=%v x=%v y=%v plane=%v owner=%v size=%v population=%v farmers=%v soldBuilding=%v population10=%v playerBits=%v construction=%v numBuildings=%v buildings=%v enchantments=%v productionUnits=%v production=%v gold=%v upkeep=%v manaUpkeep=%v research=%v food=%v roadConnections=%v", string(name), race, x, y, plane, owner, size, population, farmers, soldBuilding, population10, playerBits, construction, numBuildings, buildings, enchantments, productionUnits, production, gold, upkeep, manaUpkeep, research, food, roadConnections)

    }

    return nil
}

func loadUnits(reader io.Reader) error {
    for range 1009 {
        unitData := make([]byte, 32)
        _, err := io.ReadFull(reader, unitData)
        if err != nil {
            return err
        }

        unitReader := bytes.NewReader(unitData)

        x, err := lbx.ReadN[int8](unitReader)
        if err != nil {
            return err
        }

        y, err := lbx.ReadN[int8](unitReader)
        if err != nil {
            return err
        }

        plane, err := lbx.ReadN[int8](unitReader)
        if err != nil {
            return err
        }

        owner, err := lbx.ReadN[int8](unitReader)
        if err != nil {
            return err
        }

        movesMax, err := lbx.ReadN[int8](unitReader)
        if err != nil {
            return err
        }

        typeIndex, err := lbx.ReadN[uint8](unitReader)
        if err != nil {
            return err
        }

        hero, err := lbx.ReadN[int8](unitReader)
        if err != nil {
            return err
        }

        finished, err := lbx.ReadN[int8](unitReader)
        if err != nil {
            return err
        }

        moves, err := lbx.ReadN[int8](unitReader)
        if err != nil {
            return err
        }

        destinationX, err := lbx.ReadN[int8](unitReader)
        if err != nil {
            return err
        }

        destinationY, err := lbx.ReadN[int8](unitReader)
        if err != nil {
            return err
        }

        status, err := lbx.ReadN[int8](unitReader)
        if err != nil {
            return err
        }

        level, err := lbx.ReadN[int8](unitReader)
        if err != nil {
            return err
        }

        experience, err := lbx.ReadN[int16](unitReader)
        if err != nil {
            return err
        }

        moveFailed, err := lbx.ReadN[int8](unitReader)
        if err != nil {
            return err
        }

        damage, err := lbx.ReadN[int8](unitReader)
        if err != nil {
            return err
        }

        drawPriority, err := lbx.ReadN[int8](unitReader)
        if err != nil {
            return err
        }

        inTower, err := lbx.ReadN[int16](unitReader)
        if err != nil {
            return err
        }

        sightRange, err := lbx.ReadN[int8](unitReader)
        if err != nil {
            return err
        }

        mutations, err := lbx.ReadN[int8](unitReader)
        if err != nil {
            return err
        }

        enchantments, err := lbx.ReadN[uint32](unitReader)
        if err != nil {
            return err
        }

        roadTurns, err := lbx.ReadN[int8](unitReader)
        if err != nil {
            return err
        }

        roadX, err := lbx.ReadN[int8](unitReader)
        if err != nil {
            return err
        }

        roadY, err := lbx.ReadN[int8](unitReader)
        if err != nil {
            return err
        }

        log.Printf("Unit x=%v y=%v plane=%v owner=%v movesMax=%v typeIndex=%v hero=%v finished=%v moves=%v destinationX=%v destinationY=%v status=%v level=%v experience=%v moveFailed=%v damage=%v drawPriority=%v inTower=%v sightRange=%v mutations=%v enchantments=%v roadTurns=%v roadX=%v roadY=%v", x, y, plane, owner, movesMax, typeIndex, hero, finished, moves, destinationX, destinationY, status, level, experience, moveFailed, damage, drawPriority, inTower, sightRange, mutations, enchantments, roadTurns, roadX, roadY)
    }

    return nil
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
    numHeroes := 35

    reader := &ReadMonitor{reader: reader1}

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

    for range 6 {
        err := loadPlayerData(reader)
        if err != nil {
            return nil, err
        }
    }

    arcanusMap, err := LoadTerrain(reader)
    if err != nil {
        return nil, err
    }

    myrrorMap, err := LoadTerrain(reader)
    if err != nil {
        return nil, err
    }

    _ = arcanusMap
    _ = myrrorMap

    // FIXME: what is this for?
    uu_table_1 := make([]byte, 96 * 2)
    _, err = io.ReadFull(reader, uu_table_1)
    if err != nil {
        return nil, err
    }

    // FIXME: what is this for?
    uu_table_2 := make([]byte, 96 * 2)
    _, err = io.ReadFull(reader, uu_table_2)
    if err != nil {
        return nil, err
    }

    log.Printf("Offset: 0x%x", reader.BytesRead)

    arcanusLandMasses, err := LoadLandMass(reader)
    if err != nil {
        return nil, err
    }

    myrrorLandMasses, err := LoadLandMass(reader)
    if err != nil {
        return nil, err
    }

    _ = arcanusLandMasses
    _ = myrrorLandMasses

    log.Printf("Offset: 0x%x", reader.BytesRead)

    err = loadNodes(reader)
    if err != nil {
        return nil, err
    }

    log.Printf("Offset: 0x%x", reader.BytesRead)

    err = loadFortresses(reader)
    if err != nil {
        return nil, err
    }

    log.Printf("Offset: 0x%x", reader.BytesRead)

    err = loadTowers(reader)
    if err != nil {
        return nil, err
    }

    log.Printf("Offset: 0x%x", reader.BytesRead)

    err = loadLairs(reader)
    if err != nil {
        return nil, err
    }

    log.Printf("Offset: 0x%x", reader.BytesRead)

    err = loadItems(reader)
    if err != nil {
        return nil, err
    }

    err = loadCities(reader)
    if err != nil {
        return nil, err
    }

    err = loadUnits(reader)
    if err != nil {
        return nil, err
    }

    return nil, fmt.Errorf("unfinished")
}
