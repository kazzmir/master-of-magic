package main

import (
    "fmt"
    "log"

    "github.com/kazzmir/master-of-magic/game/magic/building"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    helplib "github.com/kazzmir/master-of-magic/game/magic/help"
    "github.com/kazzmir/master-of-magic/lib/lbx"
)

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    cache := lbx.AutoCache()

    lbxFile, err := cache.GetLbxFile("help.lbx")
    if err != nil {
        log.Printf("Unable to read help.lbx: %v", err)
        return
    }

    help, err := helplib.ReadHelp(lbxFile, 2)
    if err != nil {
        log.Printf("Unable to read help: %v", err)
        return
    }

    // Spells

    lbxFile, err = cache.GetLbxFile("spelldat.lbx")
    if err != nil {
        log.Printf("Unable to read spelldat.lbx: %v", err)
        return
    }

    spells, err := spellbook.ReadSpells(lbxFile, 0)
    if err != nil {
        log.Printf("Unable to read help: %v", err)
        return
    }

    for _, spell := range spells.Spells {
        helpEntries := help.GetEntriesByName(spell.Name)
        if helpEntries == nil {
            fmt.Printf("No entry found for spell %v\n", spell.Name)
        } else {
            if len(helpEntries) > 1 {
                fmt.Printf("Spell %v contains more than one entry\n", spell.Name)
            }
        }
    }

    // abilities
    abilities := []data.Retort{
        data.RetortAlchemy,
        data.RetortWarlord,
        data.RetortChanneler,
        data.RetortArchmage,
        data.RetortArtificer,
        data.RetortConjurer,
        data.RetortSageMaster,
        data.RetortMyrran,
        data.RetortDivinePower,
        data.RetortFamous,
        data.RetortRunemaster,
        data.RetortCharismatic,
        data.RetortChaosMastery,
        data.RetortNatureMastery,
        data.RetortSorceryMastery,
        data.RetortInfernalPower,
        data.RetortManaFocusing,
        data.RetortNodeMastery,
    }

    for _, ability := range abilities {
        helpEntries := help.GetEntriesByName(ability.String())
        if helpEntries == nil {
            fmt.Printf("No entry found for ability %v\n", ability.String())
        } else {
            if len(helpEntries) > 1 {
                fmt.Printf("Ability %v contains more than one entry\n", ability.String())
            }
        }
    }

    // Races
    races := data.ArcanianRaces()
    races = append(races, data.MyrranRaces()...)
    for _, race := range races {
        name := fmt.Sprintf("%v townsfolk", race)
        helpEntries := help.GetEntriesByName(name)
        if helpEntries == nil {
            fmt.Printf("No entry found for race %v\n", name)
        } else {
            if len(helpEntries) < 1 {
                fmt.Printf("Race %v contains only one entry\n", name)
            }
        }
    }

    // Buildings
    buildings, err := building.ReadBuildingInfo(cache)
    if err != nil {
        log.Printf("Unable to read buildings: %v", err)
        return
    }

    for _, building := range buildings {
        helpEntries := help.GetEntriesByName(building.Name)
        if helpEntries == nil {
            fmt.Printf("No entry found for building %v\n", building.Name)
        } else {
            if len(helpEntries) > 1 {
                fmt.Printf("Building %v contains more than one entry\n", building.Name)
            }
        }
    }
}
