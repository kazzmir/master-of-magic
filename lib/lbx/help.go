package lbx

import (
    "strings"
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

