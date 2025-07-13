package load

import (
    "io"
    "bufio"
    "encoding/binary"
)

func writeN[T any](writer io.Writer, value T) error {
    return binary.Write(writer, binary.LittleEndian, value)
}

func writeSlice[T any](writer io.Writer, data []T) error {
    for i := range data {
        err := writeN[T](writer, data[i])
        if err != nil {
            return err
        }
    }
    /*
    for len(data) > 0 {
        n, err := writer.Write(data)
        if err != nil {
            return err
        }
        data = data[n:]
    }
    */

    return nil
}

func writeHeroData(writer io.Writer, heroData *HeroData) error {

    err := writeN[int16](writer, heroData.Level)
    if err != nil {
        return err
    }

    err = writeN[uint32](writer, heroData.Abilities)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, heroData.CastingSkill)
    if err != nil {
        return err
    }

    for i := range 4 {
        err = writeN[uint8](writer, heroData.Spells[i])
        if err != nil {
            return err
        }
    }

    err = writeN[uint8](writer, heroData.ExtraByte)
    if err != nil {
        return err
    }

    return nil
}

func writePlayerHeroData(writer io.Writer, heroData *PlayerHeroData) error {

    err := writeN[int16](writer, heroData.Unit)
    if err != nil {
        return err
    }

    heroName := make([]byte, 14)
    copy(heroName, []byte(heroData.Name)[:min(len(heroData.Name), len(heroName))])
    err = writeSlice(writer, heroName)
    if err != nil {
        return err
    }

    for _, item := range heroData.Items {
        err = writeN[int16](writer, item)
        if err != nil {
            return err
        }
    }

    for _, itemSlot := range heroData.ItemSlot {
        err = writeN[int16](writer, itemSlot)
        if err != nil {
            return err
        }
    }

    return nil
}

func writeDiplomacy(writer io.Writer, data *DiplomacyData) error {

    err := writeSlice(writer, data.Contacted)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.TreatyInterest)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.PeaceInterest)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.TradeInterest)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.VisibleRelations)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.DiplomacyStatus)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.Strength)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.Action)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.Spell)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.City)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.DefaultRelations)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.ContactProgress)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.BrokenTreaty)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.Unknown1)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.HiddenRelations)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.Unknown2)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.TributeSpell)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.Unknown3)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.TributeGold)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.Unknown4)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.Unknown5)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.Unknown6)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.WarningProgress)
    if err != nil {
        return err
    }

    return nil
}

func writeAstrology(writer io.Writer, data *AstrologyData) error {
    err := writeN[int16](writer, data.MagicPower)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.SpellResearch)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.ArmyStrength)
    if err != nil {
        return err
    }

    return nil
}

func writePlayerData(writer io.Writer, data *PlayerData) error {
    err := writeN[uint8](writer, data.WizardId)
    if err != nil {
        return err
    }

    wizardName := make([]byte, 20)
    copy(wizardName, []byte(data.WizardName)[:min(len(data.WizardName), 20)])

    err = writeSlice(writer, wizardName)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, data.CapitalRace)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, data.BannerId)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, data.Unknown1)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.Personality)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.Objective)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.Unknown2)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.MasteryResearch)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.Fame)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.PowerBase)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.Volcanoes)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, data.ResearchRatio)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, data.ManaRatio)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, data.SkillRatio)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, data.VolcanoPower)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.SummonX)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.SummonY)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.SummonPlane)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.ResearchSpells)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.Unknown3)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.AverageUnitCost)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.Unknown4)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.CombatSkillLeft)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.CastingCostRemaining)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.CastingCostOriginal)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.CastingSpellIndex)
    if err != nil {
        return err
    }
    
    err = writeN[uint16](writer, data.SkillLeft)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.NominalSkill)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.TaxRate)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.SpellRanks)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RetortAlchemy)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RetortWarlord)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RetortChaosMastery)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RetortNatureMastery)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RetortSorceryMastery)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RetortInfernalPower)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RetortDivinePower)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RetortSageMaster)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RetortChanneler)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RetortMyrran)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RetortArchmage)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RetortManaFocusing)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RetortNodeMastery)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RetortFamous)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RetortRunemaster)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RetortConjurer)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RetortCharismatic)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RetortArtificer)
    if err != nil {
        return err
    }

    for i := range NumPlayerHeroes {
        err := writePlayerHeroData(writer, &data.HeroData[i])
        if err != nil {
            return err
        }
    }

    err = writeN[int16](writer, data.Unknown5)
    if err != nil {
        return err
    }

    for _, vaultItem := range data.VaultItems {
        err = writeN[int16](writer, vaultItem)
        if err != nil {
            return err
        }
    }

    err = writeDiplomacy(writer, &data.Diplomacy)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.ResearchCostRemaining)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.ManaReserve)
    if err != nil {
        return err
    }

    err = writeN[int32](writer, data.SpellCastingSkill)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.ResearchingSpellIndex)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.SpellsList)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.DefeatedWizards)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.GoldReserve)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.Unknown6)
    if err != nil {
        return err
    }

    err = writeAstrology(writer, &data.Astrology)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.Population)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.Historian)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.GlobalEnchantments)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.MagicStrategy)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.Unknown7)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.Hostility)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.ReevaluateHostilityCountdown)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.ReevaluateMagicStrategyCountdown)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.ReevaluateMagicPowerCountdown)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.PeaceDuration)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, data.Unknown8)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, data.Unknown9)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.Unknown10)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.TargetWizard)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, data.Unknown11)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, data.Unknown12)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.Unknown13)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.PrimaryRealm)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.SecondaryRealm)
    if err != nil {
        return err
    }

    return nil
}

func writeTerrain(writer io.Writer, terrain *TerrainData) error {
    for y := range WorldHeight {
        for x := range WorldWidth {
            err := writeN[uint16](writer, terrain.Data[x][y])
            if err != nil {
                return err
            }
        }
    }

    return nil
}

func writeLandMass(writer io.Writer, landMasses [][]uint8) error {
    for y := range WorldHeight {
        for x := range WorldWidth {
            err := writeN[uint8](writer, landMasses[x][y])
            if err != nil {
                return err
            }
        }
    }

    return nil
}

func writeNode(writer io.Writer, node *NodeData) error {
    err := writeN[int8](writer, node.X)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, node.Y)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, node.Plane)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, node.Owner)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, node.Power)
    if err != nil {
        return err
    }

    err = writeSlice(writer, node.AuraX)
    if err != nil {
        return err
    }

    err = writeSlice(writer, node.AuraY)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, node.NodeType)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, node.Flags)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, node.Unknown1)
    if err != nil {
        return err
    }

    return nil
}

func writeFortress(writer io.Writer, fortress *FortressData) error {
    err := writeN[int8](writer, fortress.X)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, fortress.Y)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, fortress.Plane)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, fortress.Active)
    if err != nil {
        return err
    }

    return nil
}

func writeTower(writer io.Writer, tower *TowerData) error {
    err := writeN[int8](writer, tower.X)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, tower.Y)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, tower.Owner)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, tower.Unknown1)
    if err != nil {
        return err
    }

    return nil
}

func writeLair(writer io.Writer, lair *LairData) error {
    err := writeN[int8](writer, lair.X)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, lair.Y)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, lair.Plane)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, lair.Intact)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, lair.Kind)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, lair.Guard1_unit_type)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, lair.Guard1_unit_count)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, lair.Guard2_unit_type)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, lair.Guard2_unit_count)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, lair.Unknown1)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, lair.Gold)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, lair.Mana)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, lair.SpellSpecial)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, lair.Flags)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, lair.ItemCount)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, lair.Unknown2)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, lair.Item1)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, lair.Item2)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, lair.Item3)
    if err != nil {
        return err
    }

    return nil
}

// write the save game object to the given writer
func WriteSaveGame(saveGame *SaveGame, writer1 io.Writer) error {
    writer := bufio.NewWriter(writer1)
    defer writer.Flush()

    for player := range saveGame.HeroData {
        for hero := range saveGame.HeroData[player] {
            err := writeHeroData(writer, &saveGame.HeroData[player][hero])
            if err != nil {
                return err
            }
        }
    }

    err := writeN[int16](writer, saveGame.NumPlayers)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, saveGame.LandSize)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, saveGame.Magic)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, saveGame.Difficulty)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, saveGame.NumCities)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, saveGame.NumUnits)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, saveGame.Turn)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, saveGame.Unit)
    if err != nil {
        return err
    }

    for i := range saveGame.PlayerData {
        err = writePlayerData(writer, &saveGame.PlayerData[i])
        if err != nil {
            return err
        }
    }

    err = writeTerrain(writer, &saveGame.ArcanusMap)
    if err != nil {
        return err
    }

    err = writeTerrain(writer, &saveGame.MyrrorMap)
    if err != nil {
        return err
    }

    err = writeSlice(writer, saveGame.UU_table_1)
    if err != nil {
        return err
    }

    err = writeSlice(writer, saveGame.UU_table_2)
    if err != nil {
        return err
    }

    err = writeLandMass(writer, saveGame.ArcanusLandMasses)
    if err != nil {
        return err
    }

    err = writeLandMass(writer, saveGame.MyrrorLandMasses)
    if err != nil {
        return err
    }

    for i := range saveGame.Nodes {
        err = writeNode(writer, &saveGame.Nodes[i])
        if err != nil {
            return err
        }
    }

    for i := range saveGame.Fortresses {
        err = writeFortress(writer, &saveGame.Fortresses[i])
        if err != nil {
            return err
        }
    }

    for i := range saveGame.Towers {
        err = writeTower(writer, &saveGame.Towers[i])
        if err != nil {
            return err
        }
    }

    for i := range saveGame.Lairs {
        err = writeLair(writer, &saveGame.Lairs[i])
        if err != nil {
            return err
        }
    }

    return nil
}
