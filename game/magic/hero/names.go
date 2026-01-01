package hero

// read the names of the heroes per wizard from names.lbx
// there are 5 sets of names, one for each player. each player gives a unique name to their hero

import (
    "log"
    "fmt"
    "io"

    "github.com/kazzmir/master-of-magic/lib/lbx"
)

func ReadNamesPerWizard(cache *lbx.LbxCache) map[int]map[HeroType]string {
    names, err := ReadNames(cache)

    if err != nil {
        log.Printf("Unable to read hero names: %v", err)
        return make(map[int]map[HeroType]string)
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

    choices := make(map[int]map[HeroType]string)

    nameIndex := 0
    for i := range 5 {
        choices[i] = make(map[HeroType]string)
        for _, hero := range heroOrder {
            if nameIndex >= len(names) {
                break
            }
            choices[i][hero] = names[nameIndex]
            nameIndex += 1
        }
    }

    return choices
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

        // log.Printf("Name: '%v'", string(data))
        name := string(data)
        // trim null terminators
        for i, c := range name {
            if c == 0 {
                name = name[:i]
                break
            }
        }

        out = append(out, name)
    }

    return out, nil
}
