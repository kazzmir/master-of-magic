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

func (help *Help) GetEntriesByName(name string) []HelpEntry {
    entry, ok := help.headlineMap[strings.ToLower(name)]
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

