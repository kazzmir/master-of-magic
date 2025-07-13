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

        if i == 0 {
            break
        }
    }

    return nil
}
