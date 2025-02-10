package game

import (
    "fmt"
    "github.com/kazzmir/master-of-magic/lib/lbx"
)

type EventType int

const (
    EventDisjunction EventType = iota
)

/* an event is something shown on screen to the user, like when a new building is created
 */
type Event struct {
    Type EventType
    BirthYear int // year/turn the event started
    Message string // messages are supposed to come out of eventmsg.lbx, but we mostly just hardcode them
    LbxIndex int // picture from events.lbx
    IsConjunction bool // only one conjunction event can be active at a time
    CityEvent bool // if true, then this event targets a city
}

func MakeDisjunctionEvent(year int) *Event {
    return &Event{
        Type: EventDisjunction,
        BirthYear: year,
        Message: "Disjunction! The fabric of magic has been torn asunder destroying all overland enchantments.",
        LbxIndex: 2,
        IsConjunction: false,
        CityEvent: false,
    }
}

type EventData struct {
    // these strings contain bytes that indicate a placeholder to insert some other value, such as the wizard's name or a city name
    Events []string
}

func parseString(s []byte) []byte {
    // strings are null terminated
    for i, c := range s {
        if c == 0 {
            return s[:i]
        }
    }
    return s
}

func ReadEventData(cache *lbx.LbxCache) (*EventData, error) {
    lbxData, err := cache.GetLbxFile("eventmsg.lbx")
    if err != nil {
        return nil, err
    }

    reader, err := lbxData.GetReader(0)
    if err != nil {
        return nil, err
    }

    count, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    size, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    events := make([]string, count)

    for i := 0; i < int(count); i++ {
        data := make([]byte, size)
        n, err := reader.Read(data)
        if err != nil {
            return nil, err
        }
        if n != int(size) {
            return nil, fmt.Errorf("did not read all %v bytes", size)
        }

        events[i] = string(parseString(data))
    }

    return &EventData{Events: events}, nil
}
