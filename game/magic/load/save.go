package load

import (
    "io"
    "bufio"
    "encoding/binary"
)

func writeN[T any](writer io.Writer, value T) error {
    return binary.Write(writer, binary.LittleEndian, value)
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

    err = writeN[uint8](writer, 0) // empty byte
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

    return nil
}
