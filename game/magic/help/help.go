package help

import (
    "bytes"
    "fmt"
    "strings"

    "github.com/kazzmir/master-of-magic/lib/lbx"
)

// stuff from help.lbx
type HelpEntry struct {
    // shows up on top of scroll
    Headline string
    // lbx file to pull an icon out, shown at top left of scroll
    Lbx string
    // index of sprite in the lbx file
    LbxIndex int
    // extra text to append
    AppendHelpIndex int
    // text displayed in scroll
    Text string
}

type Help struct {
    Entries []HelpEntry
    // map from headline name to its entry
    headlineMap map[string]int
}

func (help *Help) GetRawEntry(entry int) HelpEntry {
    return help.Entries[entry]
}

func (help *Help) updateMap(){
    help.headlineMap = make(map[string]int)

    for i, entry := range help.Entries {
        help.headlineMap[strings.ToLower(entry.Headline)] = i
    }
}

func (help *Help) autocorrectName(name string) string {
    switch name {
        // spells
        case "Basilisk": return "Basilisks"
        case "Nature Awareness": return "Nature's Awareness"
        case "Floating Island": return "Floating Islands"
        case "Invisiblity": return "Invisibility"
        case "Chimeras": return "Chimera"
        case "Armageddon": return "Armagedon"
        case "Lionheart": return "Lion Heart"
        case "Arch Angel": return "Archangel"
        // buildings
        case "Fighters Guild": return "Fighter's Guild"
        case "Armorers Guild": return "Armorer's Guild"
        case "Animists Guild": return "Animist's Guild"
        case "Ship Yard": return "Shipyard"
        case "Sages Guild": return "Sage's Guild"
        case "Wizards Guild": return "Wizard's Guild"
        case "Merchants Guild": return "Merchant's Guild"
        case "Farmers Market": return "Farmer's Market"
        case "Foresters Guild": return "Forester's Guild"
        case "Builders Hall": return "Builder's Hall"
        case "Mechanicians Guild": return "Mechanician's Guild"
        case "Miners Guild": return "Miner's Guild"
    }
    return name
}

func (help *Help) GetEntriesByName(name string) []HelpEntry {
    entry, ok := help.headlineMap[strings.ToLower(name)]
    if ok {
        return help.GetEntries(entry)
    }

    // FIXME: Use a more generic approach like
    entry, ok = help.headlineMap[strings.ToLower(help.autocorrectName(name))]
    if ok {
        return help.GetEntries(entry)
    }

    return nil
}

func (help *Help) GetEntries(entry int) []HelpEntry {
    if entry < 0 || entry >= len(help.Entries) {
        return nil
    }

    var out []HelpEntry

    use := help.Entries[entry]
    out = append(out, use)
    if use.AppendHelpIndex != 0 {
        var next []HelpEntry
        // special value means append next entry
        if use.AppendHelpIndex == 0xffff {
            next = help.GetEntries(entry+1)
        } else {
            next = help.GetEntries(use.AppendHelpIndex)
        }
        out = append(out, next...)
    }

    return out
}

func ReadHelp(lbxFile *lbx.LbxFile, entry int) (Help, error) {
    if entry < 0 || entry >= len(lbxFile.Data) {
        return Help{}, fmt.Errorf("invalid lbx index %v, must be between 0 and %v", entry, len(lbxFile.Data) - 1)
    }

    reader := bytes.NewReader(lbxFile.Data[entry])

    numEntries, err := lbx.ReadUint16(reader)
    if err != nil {
        return Help{}, err
    }

    entrySize, err := lbx.ReadUint16(reader)
    if err != nil {
        return Help{}, err
    }

    if entrySize == 0 {
        return Help{}, fmt.Errorf("entry size was 0 in help")
    }

    if numEntries * entrySize > uint16(reader.Len()) {
        return Help{}, fmt.Errorf("too many entries in the help file entries=%v size=%v len=%v", numEntries, entrySize, reader.Len())
    }

    var help []HelpEntry

    // fmt.Printf("num entries: %v\n", numEntries)

    for i := 0; i < int(numEntries); i++ {
        data := make([]byte, entrySize)
        n, err := reader.Read(data)
        if err != nil {
            return Help{}, fmt.Errorf("Error reading help index %v: %v", i, err)
        }

        buffer := bytes.NewBuffer(data[0:n])

        headlineData := buffer.Next(30)

        b2 := bytes.NewBuffer(headlineData)
        headline, err := b2.ReadString(0)
        if err != nil {
            headline = string(headlineData)
        } else {
            headline = headline[0:len(headline)-1]
        }

        // fmt.Printf("  headline: %v\n", string(headline))

        // fmt.Printf("  at position 0x%x\n", n - buffer.Len())

        pictureLbxData := buffer.Next(14)
        b2 = bytes.NewBuffer(pictureLbxData)
        pictureLbx, err := b2.ReadString(0)
        // fmt.Printf("  lbx: %v\n", string(pictureLbx))

        pictureLbx = pictureLbx[0:len(pictureLbx)-1]

        // fmt.Printf("  at position 0x%x\n", n - buffer.Len())

        pictureIndex, err := lbx.ReadUint16(buffer)
        if err != nil {
            return Help{}, fmt.Errorf("Error reading help index %v: %v", i, err)
        }

        // fmt.Printf("  lbx index: %v\n", pictureIndex)

        appendHelpText, err := lbx.ReadUint16(buffer)
        if err != nil {
            return Help{}, err
        }

        // fmt.Printf("  appended help text: 0x%x\n", appendHelpText)

        info, err := buffer.ReadString(0)
        if err != nil {
            return Help{}, err
        }

        info = info[0:len(info)-1]

        // fmt.Printf("  text: '%v'\n", info)

        help = append(help, HelpEntry{
            Headline: headline,
            Lbx: pictureLbx,
            LbxIndex: int(pictureIndex),
            AppendHelpIndex: int(appendHelpText),
            Text: info,
        })
    }

    out := Help{Entries: help}
    out.updateMap()
    return out, nil
}

func ReadHelpFromCache(cache *lbx.LbxCache) (Help, error) {
    helpLbx, err := cache.GetLbxFile("help.lbx")
    if err != nil {
        return Help{}, nil
    }

    return ReadHelp(helpLbx, 2)
}
