package building

import (
    "log"
    "bytes"

    "github.com/kazzmir/master-of-magic/lib/lbx"
)

type BuildingDescriptions struct {
    Descriptions []string
}

func MakeBuildDescriptions(cache *lbx.LbxCache) *BuildingDescriptions {
    buildDescriptionLbx, err := cache.GetLbxFile("buildesc.lbx")
    if err == nil {
    } else {
        log.Printf("Unable to read building descriptions")
        return nil
    }

    descriptions := readBuildDescriptions(buildDescriptionLbx)

    return &BuildingDescriptions{
        Descriptions: descriptions,
    }
}

func (descriptions *BuildingDescriptions) Get(building Building) string {
    switch building {
        case BuildingTradeGoods: return descriptions.Descriptions[1]
        case BuildingHousing: return descriptions.Descriptions[2]
        case BuildingBarracks: return descriptions.Descriptions[3]
        case BuildingArmory: return descriptions.Descriptions[4]
        case BuildingFightersGuild: return descriptions.Descriptions[5]
        case BuildingArmorersGuild: return descriptions.Descriptions[6]
        case BuildingWarCollege: return descriptions.Descriptions[7]
        case BuildingSmithy: return descriptions.Descriptions[8]
        case BuildingStables: return descriptions.Descriptions[9]
        case BuildingAnimistsGuild: return descriptions.Descriptions[10]
        case BuildingFantasticStable: return descriptions.Descriptions[11]
        case BuildingShipwrightsGuild: return descriptions.Descriptions[12]
        case BuildingShipYard: return descriptions.Descriptions[13]
        case BuildingMaritimeGuild: return descriptions.Descriptions[14]
        case BuildingSawmill: return descriptions.Descriptions[15]
        case BuildingLibrary: return descriptions.Descriptions[16]
        case BuildingSagesGuild: return descriptions.Descriptions[17]
        case BuildingOracle: return descriptions.Descriptions[18]
        case BuildingAlchemistsGuild: return descriptions.Descriptions[19]
        case BuildingUniversity: return descriptions.Descriptions[20]
        case BuildingWizardsGuild: return descriptions.Descriptions[21]
        case BuildingShrine: return descriptions.Descriptions[22]
        case BuildingTemple: return descriptions.Descriptions[23]
        case BuildingParthenon: return descriptions.Descriptions[24]
        case BuildingCathedral: return descriptions.Descriptions[25]
        case BuildingMarketplace: return descriptions.Descriptions[26]
        case BuildingBank: return descriptions.Descriptions[27]
        case BuildingMerchantsGuild: return descriptions.Descriptions[28]
        case BuildingGranary: return descriptions.Descriptions[29]
        case BuildingFarmersMarket: return descriptions.Descriptions[30]
        case BuildingForestersGuild: return descriptions.Descriptions[31]
        case BuildingBuildersHall: return descriptions.Descriptions[32]
        case BuildingMechaniciansGuild: return descriptions.Descriptions[33]
        case BuildingMinersGuild: return descriptions.Descriptions[34]
        case BuildingCityWalls: return descriptions.Descriptions[35]
    }

    return ""
}

func readBuildDescriptions(buildDescriptionLbx *lbx.LbxFile) []string {
    entries, err := buildDescriptionLbx.RawData(0)
    if err != nil {
        return nil
    }

    reader := bytes.NewReader(entries)

    count, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil
    }

    if count > 10000 {
        return nil
    }

    size, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil
    }

    if size > 10000 {
        return nil
    }

    var descriptions []string

    for i := 0; i < int(count); i++ {
        data := make([]byte, size)
        _, err := reader.Read(data)

        if err != nil {
            break
        }

        nullByte := bytes.IndexByte(data, 0)
        if nullByte != -1 {
            descriptions = append(descriptions, string(data[0:nullByte]))
        } else {
            descriptions = append(descriptions, string(data))
        }
    }

    return descriptions
}
