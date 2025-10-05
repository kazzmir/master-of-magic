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

func writeItem(writer io.Writer, data *ItemData) error {
    err := writeSlice(writer, data.Name)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.IconIndex)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, data.Slot)
    if err != nil {
        return err
    }

    err = writeN[byte](writer, data.Type)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.Cost)
    if err != nil {
        return err
    }

    err = writeN[byte](writer, data.Attack)
    if err != nil {
        return err
    }

    err = writeN[byte](writer, data.ToHit)
    if err != nil {
        return err
    }

    err = writeN[byte](writer, data.Defense)
    if err != nil {
        return err
    }

    err = writeN[byte](writer, data.Movement)
    if err != nil {
        return err
    }

    err = writeN[byte](writer, data.Resistance)
    if err != nil {
        return err
    }

    err = writeN[byte](writer, data.SpellSkill)
    if err != nil {
        return err
    }

    err = writeN[byte](writer, data.SpellSave)
    if err != nil {
        return err
    }

    err = writeN[byte](writer, data.Spell)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, data.Charges)
    if err != nil {
        return err
    }

    err = writeN[uint32](writer, data.Abilities)
    if err != nil {
        return err
    }

    return nil
}

func writeCity(writer io.Writer, data *CityData) error {
    err := writeSlice(writer, data.Name)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.Race)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.X)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.Y)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.Plane)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.Owner)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.Size)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.Population)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.Farmers)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.SoldBuilding)
    if err != nil {
        return err
    }

    err = writeN[byte](writer, data.Unknown1)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.Population10)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, data.PlayerBits)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, data.Unknown2)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.Construction)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.NumBuildings)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.Buildings)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.Enchantments)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.ProductionUnits)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.Production)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, data.Gold)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.Upkeep)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.ManaUpkeep)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.Research)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.Food)
    if err != nil {
        return err
    }

    err = writeSlice(writer, data.RoadConnections)
    if err != nil {
        return err
    }

    return nil
}

func writeUnit(writer io.Writer, data *UnitData) error {
    err := writeN[int8](writer, data.X)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.Y)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.Plane)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.Owner)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.MovesMax)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, data.TypeIndex)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.HeroSlot)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.Finished)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.Moves)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.DestinationX)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.DestinationY)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.Status)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.Level)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, data.Unknown2)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.Experience)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.MoveFailed)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.Damage)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.DrawPriority)
    if err != nil {
        return err
    }

    err = writeN[uint8](writer, data.Unknown3)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.InTower)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.SightRange)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.Mutations)
    if err != nil {
        return err
    }

    err = writeN[uint32](writer, data.Enchantments)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RoadTurns)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RoadX)
    if err != nil {
        return err
    }

    err = writeN[int8](writer, data.RoadY)
    if err != nil {
        return err
    }

    err = writeN(writer, data.Unknown1)
    if err != nil {
        return err
    }

    return nil
}

func writeMapData[T any](writer io.Writer, data [][]T) error {
    for y := range WorldHeight {
        for x := range WorldWidth {
            err := writeN[T](writer, data[x][y])
            if err != nil {
                return err
            }
        }
    }

    return nil
}

func writeMovementCostData(writer io.Writer, data *MovementCostData) error {
    err := writeMapData(writer, data.Moves)
    if err != nil {
        return err
    }

    err = writeMapData(writer, data.Walking)
    if err != nil {
        return err
    }

    err = writeMapData(writer, data.Forester)
    if err != nil {
        return err
    }

    err = writeMapData(writer, data.Mountaineer)
    if err != nil {
        return err
    }

    err = writeMapData(writer, data.Swimming)
    if err != nil {
        return err
    }

    err = writeMapData(writer, data.Sailing)
    if err != nil {
        return err
    }

    return nil
}

func writeEvents(writer io.Writer, data *EventData) error {
    err := writeN[int16](writer, data.LastEvent)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.MeteorStatus)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.MeteorPlayer)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.MeteorData)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.GiftStatus)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.GiftPlayer)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.GiftData)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.DisjunctionStatus)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.MarriageStatus)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.MarriagePlayer)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.MarriageNeutralCity)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.MarriagePlayerCity)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.EarthquakeStatus)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.EarthquakePlayer)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.EarthquakeData)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.PirateStatus)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.PiratePlayer)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.PirateData)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.PlagueStatus)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.PlaguePlayer)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.PlagueData)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.PlagueDuration)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.RebellionStatus)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.RebellionPlayer)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.RebellionData)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.DonationStatus)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.DonationPlayer)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.DonationData)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.DepletionStatus)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.DepletionPlayer)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.DepletionData)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.MineralsStatus)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.MineralsData)
    if err != nil {
        return err
    }

    // FIXME: this might be out of order with MineralsData
    err = writeN[int16](writer, data.MineralsPlayer)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.PopulationBoomStatus)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.PopulationBoomData)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.PopulationBoomPlayer)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.PopulationBoomDuration)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.GoodMoonStatus)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.GoodMoonDuration)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.BadMoonStatus)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.BadMoonDuration)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.ConjunctionChaosStatus)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.ConjunctionChaosDuration)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.ConjunctionNatureStatus)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.ConjunctionNatureDuration)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.ConjunctionSorceryStatus)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.ConjunctionSorceryDuration)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.ManaShortageStatus)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.ManaShortageDuration)
    if err != nil {
        return err
    }

    return nil
}

func writeHeroNameData(writer io.Writer, data *HeroNameData) error {
    err := writeSlice(writer, data.Name)
    if err != nil {
        return err
    }

    err = writeN[int16](writer, data.Experience)
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

    for i := range saveGame.Items {
        err = writeItem(writer, &saveGame.Items[i])
        if err != nil {
            return err
        }
    }

    for i := range saveGame.Cities {
        err = writeCity(writer, &saveGame.Cities[i])
        if err != nil {
            return err
        }
    }

    for i := range saveGame.Units {
        err = writeUnit(writer, &saveGame.Units[i])
        if err != nil {
            return err
        }
    }

    err = writeMapData(writer, saveGame.ArcanusTerrainSpecials)
    if err != nil {
        return err
    }

    err = writeMapData(writer, saveGame.MyrrorTerrainSpecials)
    if err != nil {
        return err
    }

    err = writeMapData(writer, saveGame.ArcanusExplored)
    if err != nil {
        return err
    }

    err = writeMapData(writer, saveGame.MyrrorExplored)
    if err != nil {
        return err
    }

    err = writeMovementCostData(writer, &saveGame.ArcanusMovementCost)
    if err != nil {
        return err
    }

    err = writeMovementCostData(writer, &saveGame.MyrrorMovementCost)
    if err != nil {
        return err
    }

    err = writeEvents(writer, &saveGame.Events)
    if err != nil {
        return err
    }

    err = writeMapData(writer, saveGame.ArcanusMapSquareFlags)
    if err != nil {
        return err
    }

    err = writeMapData(writer, saveGame.MyrrorMapSquareFlags)
    if err != nil {
        return err
    }

    err = writeN[uint16](writer, saveGame.GrandVizier)
    if err != nil {
        return err
    }

    for i := range saveGame.PremadeItems {
        // err = writePremadeItem(writer, saveGame.PremadeItems[i])
        err = writeN[byte](writer, saveGame.PremadeItems[i])
        if err != nil {
            return err
        }
    }

    for i := range saveGame.HeroNames {
        err = writeHeroNameData(writer, &saveGame.HeroNames[i])
        if err != nil {
            return err
        }
    }

    return nil
}
