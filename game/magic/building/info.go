package building

import (
    "fmt"
    "bytes"
    "github.com/kazzmir/master-of-magic/lib/lbx"
)

func parseName(name []byte) string {
    for i, b := range name {
        if b == 0 {
            return string(name[:i])
        }
    }
    return string(name)
}

const ForrestTerrain = 101
const WaterTerrain = 110
const MineralTerrain = 200

type BuildingInfos []BuildingInfo

type BuildingInfo struct {
    Name string

    // index of building that must exist first, or 0 if no dependency
    BuildingDependency1 int
    BuildingDependency2 int

    // -1 for no terrain, otherwise specifies a tile index that the building can be built on
    TerrainDependency int

    // replaces the given building
    BuildingReplace int

    Grant20XP bool
    Grant60XP bool
    // grants magic weapons to new units if appropriate minerals around
    Alchemist bool

    // required gold to maintain this building
    UpkeepGold int
    UpkeepPower int
    PopulationGrowth int
    Religion int
    // points of research produced each turn
    Research int

    ConstructionCost int
    Animation int

    // 0:  None, Trade, Housing
    // 1:  Marketplace, Bank, Merchants Guild, Maritime Guild
    // 2:  Shrine, Temple, Parthenon, Cathedral
    // 3:  Library, Sages Guild, Oracle, Alchemists Guild, University, Wizards Guild
    // 4:  Barracks, Armory, Fighters Guild, Armorers Guild, War College, Smithy, Stables, Fantastic Stable, Mechanicians Guild, City Walls
    Category int
}

/* return the buildings that have the provided building as a dependency */
func (info BuildingInfos) Allows(building Building) []Building {
    var out []Building

    buildingIndex := info.GetBuildingIndex(building)

    for i, check := range info {
        if check.BuildingDependency1 == buildingIndex {
            out = append(out, indexToBuilding(i))
        }

        if check.BuildingDependency2 == buildingIndex {
            out = append(out, indexToBuilding(i))
        }
    }

    return out
}

func (info BuildingInfos) GetBuildingByName(name string) *BuildingInfo {
    for i := 0; i < len(info); i++ {
        if info[i].Name == name {
            return &info[i]
        }
    }
    return nil
}

func (info BuildingInfos) Dependencies(building Building) []Building {
    var out []Building
    data := info.BuildingInfo(building)
    if data.BuildingDependency1 != 0 {
        out = append(out, indexToBuilding(data.BuildingDependency1))
    }

    if data.BuildingDependency2 != 0 {
        out = append(out, indexToBuilding(data.BuildingDependency2))
    }

    return out
}

func (info BuildingInfos) BuildingInfo(building Building) BuildingInfo {
    return info[info.GetBuildingIndex(building)]
}

func (info BuildingInfos) ProductionCost(building Building) int {
    return info[info.GetBuildingIndex(building)].ConstructionCost
}

func (info BuildingInfos) UpkeepCost(building Building) int {
    return info[info.GetBuildingIndex(building)].UpkeepGold
}

func (info BuildingInfos) ManaProduction(building Building) int {
    return info[info.GetBuildingIndex(building)].Religion
}

func (info BuildingInfos) ResearchProduction(building Building) int {
    return info[info.GetBuildingIndex(building)].Research
}

func (info BuildingInfos) Name(building Building) string {
    if building == BuildingFortress {
        return "Fortress"
    }

    if building == BuildingSummoningCircle {
        return "Summoning Circle"
    }

    return info[info.GetBuildingIndex(building)].Name
}

func (info BuildingInfos) GetBuildingIndex(building Building) int {
    switch building {
        case BuildingNone: return 0
        case BuildingTradeGoods: return 1
        case BuildingHousing: return 2
        case BuildingBarracks: return 3
        case BuildingArmory: return 4
        case BuildingFightersGuild: return 5
        case BuildingArmorersGuild: return 6
        case BuildingWarCollege: return 7
        case BuildingSmithy: return 8
        case BuildingStables: return 9
        case BuildingAnimistsGuild: return 10
        case BuildingFantasticStable: return 11
        case BuildingShipwrightsGuild: return 12
        case BuildingShipYard: return 13
        case BuildingMaritimeGuild: return 14
        case BuildingSawmill: return 15
        case BuildingLibrary: return 16
        case BuildingSagesGuild: return 17
        case BuildingOracle: return 18
        case BuildingAlchemistsGuild: return 19
        case BuildingUniversity: return 20
        case BuildingWizardsGuild: return 21
        case BuildingShrine: return 22
        case BuildingTemple: return 23
        case BuildingParthenon: return 24
        case BuildingCathedral: return 25
        case BuildingMarketplace: return 26
        case BuildingBank: return 27
        case BuildingMerchantsGuild: return 28
        case BuildingGranary: return 29
        case BuildingFarmersMarket: return 30
        case BuildingForestersGuild: return 31
        case BuildingBuildersHall: return 32
        case BuildingMechaniciansGuild: return 33
        case BuildingMinersGuild: return 34
        case BuildingCityWalls: return 35
    }

    return 0
}

func indexToBuilding(index int) Building {
    switch index {
        case 0: return BuildingNone
        case 1: return BuildingTradeGoods
        case 2: return BuildingHousing
        case 3: return BuildingBarracks
        case 4: return BuildingArmory
        case 5: return BuildingFightersGuild
        case 6: return BuildingArmorersGuild
        case 7: return BuildingWarCollege
        case 8: return BuildingSmithy
        case 9: return BuildingStables
        case 10: return BuildingAnimistsGuild
        case 11: return BuildingFantasticStable
        case 12: return BuildingShipwrightsGuild
        case 13: return BuildingShipYard
        case 14: return BuildingMaritimeGuild
        case 15: return BuildingSawmill
        case 16: return BuildingLibrary
        case 17: return BuildingSagesGuild
        case 18: return BuildingOracle
        case 19: return BuildingAlchemistsGuild
        case 20: return BuildingUniversity
        case 21: return BuildingWizardsGuild
        case 22: return BuildingShrine
        case 23: return BuildingTemple
        case 24: return BuildingParthenon
        case 25: return BuildingCathedral
        case 26: return BuildingMarketplace
        case 27: return BuildingBank
        case 28: return BuildingMerchantsGuild
        case 29: return BuildingGranary
        case 30: return BuildingFarmersMarket
        case 31: return BuildingForestersGuild
        case 32: return BuildingBuildersHall
        case 33: return BuildingMechaniciansGuild
        case 34: return BuildingMinersGuild
        case 35: return BuildingCityWalls
    }

    return BuildingNone
}

func ReadBuildingInfo(cache *lbx.LbxCache) (BuildingInfos, error) {
    data, err := cache.GetLbxFile("builddat.lbx")
    if err != nil {
        return nil, fmt.Errorf("Unable to read builddat.lbx: %v", err)
    }

    reader, err := data.GetReader(0)
    if err != nil {
        return nil, fmt.Errorf("unable to read entry 0 in builddat.lbx: %v", err)
    }

    numBuildings, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil, fmt.Errorf("read error: %v", err)
    }

    entrySize, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil, fmt.Errorf("read error: %v", err)
    }

    var out BuildingInfos

    for i := 0; i < int(numBuildings); i++ {
        buildingData := make([]byte, entrySize)
        n, err := reader.Read(buildingData)
        if err != nil || n != int(entrySize) {
            return nil, fmt.Errorf("unable to read building info %v: %v", i, err)
        }

        buildingReader := bytes.NewReader(buildingData)
        name := make([]byte, 20)
        _, err = buildingReader.Read(name)
        if err != nil {
            return nil, fmt.Errorf("unable to read building name %v: %v", i, err)
        }

        dependency1, err := lbx.ReadUint16(buildingReader)
        if err != nil {
            return nil, fmt.Errorf("unable to read building dependency1 %v: %v", i, err)
        }

        dependency2, err := lbx.ReadUint16(buildingReader)
        if err != nil {
            return nil, fmt.Errorf("unable to read building dependency2 %v: %v", i, err)
        }

        replace, err := lbx.ReadUint16(buildingReader)
        if err != nil {
            return nil, fmt.Errorf("unable to read building replace %v: %v", i, err)
        }

        if replace == 65535 {
            replace = 0
        }

        grant20xp, err := lbx.ReadUint16(buildingReader)
        if err != nil {
            return nil, fmt.Errorf("unable to read building grant20xp %v: %v", i, err)
        }

        grant60xp, err := lbx.ReadUint16(buildingReader)
        if err != nil {
            return nil, fmt.Errorf("unable to read building grant60xp %v: %v", i, err)
        }

        alchemist, err := lbx.ReadUint16(buildingReader)
        if err != nil {
            return nil, fmt.Errorf("unable to read building alchemist %v: %v", i, err)
        }

        upkeepGold, err := lbx.ReadUint16(buildingReader)
        if err != nil {
            return nil, fmt.Errorf("unable to read building upkeepGold %v: %v", i, err)
        }

        populationGrowth, err := lbx.ReadUint16(buildingReader)
        if err != nil {
            return nil, fmt.Errorf("unable to read building populationGrowth %v: %v", i, err)
        }

        unknown1, err := lbx.ReadUint16(buildingReader)
        if err != nil {
            return nil, fmt.Errorf("unable to read building unknown1 %v: %v", i, err)
        }
        _ = unknown1

        unknown2, err := lbx.ReadUint16(buildingReader)
        if err != nil {
            return nil, fmt.Errorf("unable to read building unknown2 %v: %v", i, err)
        }
        _ = unknown2

        religion, err := lbx.ReadUint16(buildingReader)
        if err != nil {
            return nil, fmt.Errorf("unable to read building religion %v: %v", i, err)
        }

        research, err := lbx.ReadUint16(buildingReader)
        if err != nil {
            return nil, fmt.Errorf("unable to read building research %v: %v", i, err)
        }

        constructionCost, err := lbx.ReadUint16(buildingReader)
        if err != nil {
            return nil, fmt.Errorf("unable to read building constructionCost %v: %v", i, err)
        }

        unknown3, err := lbx.ReadUint16(buildingReader)
        if err != nil {
            return nil, fmt.Errorf("unable to read building unknown3 %v: %v", i, err)
        }
        _ = unknown3

        animation, err := lbx.ReadUint16(buildingReader)
        if err != nil {
            return nil, fmt.Errorf("unable to read building animation %v: %v", i, err)
        }

        category, err := lbx.ReadUint16(buildingReader)
        if err != nil {
            return nil, fmt.Errorf("unable to read building category %v: %v", i, err)
        }

        terrainDependency := uint16(0)
        if dependency1 > 100 {
            terrainDependency = dependency1
            dependency1 = 0
        }

        info := BuildingInfo{
            Name: parseName(name),
            BuildingDependency1: int(dependency1),
            BuildingDependency2: int(int16(dependency2)),
            TerrainDependency: int(int16(terrainDependency)),
            BuildingReplace: int(replace),
            Grant20XP: grant20xp == 1,
            Grant60XP: grant60xp == 1,
            Alchemist: alchemist == 1,
            UpkeepGold: -int(int16(upkeepGold)),
            PopulationGrowth: int(populationGrowth),
            Religion: int(religion),
            Research: int(research),
            ConstructionCost: int(constructionCost),
            Animation: int(int16(animation)),
            Category: int(category),
        }

        // fmt.Printf("Building %v: %+v\n", i, info)
        out = append(out, info)
    }

    // builddat has research as 7, but its supposed to be 5
    university := out.GetBuildingByName("University")
    if university != nil {
        university.Research = 5
    }

    // FIXME: does this value live in builddat.lbx?
    wizardsGuild := out.GetBuildingByName("Wizards' Guild")
    if wizardsGuild != nil {
        wizardsGuild.UpkeepPower = 3
    }

    oracle := out.GetBuildingByName("Oracle")
    if oracle != nil {
        // oracle should not provide research
        oracle.Research = 0
    }

    none := out.GetBuildingByName("None")
    if none != nil {
        none.Name = ""
    }

    return out, nil
}
