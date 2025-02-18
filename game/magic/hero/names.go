package hero

// read the names of the heroes per wizard from names.lbx
// each wizarc has their own set of heroes, and each wizard gives a unique name to their hero

import (
    "log"
    "fmt"
    "io"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/lib/lbx"
)

func ReadNamesPerWizard(cache *lbx.LbxCache) map[data.WizardBase]map[HeroType]string {
    names, err := ReadNames(cache)

    if err != nil {
        log.Printf("Unable to read hero names: %v", err)
        return make(map[data.WizardBase]map[HeroType]string)
    }

    wizardOrder := []data.WizardBase{
        data.WizardMerlin,
        data.WizardRaven,
        data.WizardSharee,
        data.WizardLoPan,
        data.WizardJafar,
        data.WizardOberic,
        data.WizardRjak,
        data.WizardSssra,
        data.WizardTauron,
        data.WizardFreya,
        data.WizardHorus,
        data.WizardAriel,
        data.WizardTlaloc,
        data.WizardKali,
    }

    heroOrder := []HeroType{
        HeroBrax,
        HeroGunther,
        HeroZaldron,
        HeroBShan,
        HeroRakir,
        HeroValana,
        HeroBahgtru,
        HeroSerena,
        HeroShuri,
        HeroTheria,
        HeroGreyfairer,
        HeroTaki,
        HeroReywind,
        HeroMalleus,
        HeroTumu,
        HeroJaer,
        HeroMarcus,
        HeroFang,
        HeroMorgana,
        HeroAureus,
        HeroShinBo,
        HeroSpyder,
        HeroShalla,
        HeroYramrag,
        HeroMysticX,
        HeroAerie,
        HeroDethStryke,
        HeroElana,
        HeroRoland,
        HeroMortu,
        HeroAlorra,
        HeroSirHarold,
        HeroRavashack,
        HeroWarrax,
        HeroTorin,
    }

    wizards := make(map[data.WizardBase]map[HeroType]string)

    nameIndex := 0
    for _, wizard := range wizardOrder {
        wizards[wizard] = make(map[HeroType]string)
        for _, hero := range heroOrder {
            if nameIndex >= len(names) {
                break
            }
            wizards[wizard][hero] = names[nameIndex]
            nameIndex += 1
        }
    }

    return wizards
}

func ReadNames(cache *lbx.LbxCache) ([]string, error) {
    lbxFile, err := cache.GetLbxFile("names.lbx")
    if err != nil {
        return nil, err
    }

    reader, err := lbxFile.GetReader(0)
    if err != nil {
        return nil, err
    }

    count, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    if count > 10000 {
        return nil, fmt.Errorf("Name count was too high: %v", count)
    }

    size, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    if size > 10000 {
        return nil, fmt.Errorf("Size of each name entry was too high: %v", size)
    }

    var out []string

    data := make([]byte, size)
    for range count {
        n, err := io.ReadFull(reader, data)
        if err != nil {
            return nil, err
        }
        if n != len(data) {
            return nil, fmt.Errorf("Failed to read all of the name data (%v)", n)
        }

        log.Printf("Name: '%v'", string(data))
        out = append(out, string(data))
    }

    return out, nil
}
