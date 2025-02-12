package game

import (
    "fmt"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
)

type EventType int

const (
    EventNone EventType = iota
    EventBadMoon
    EventConjunctionChaos
    EventConjunctionNature
    EventConjunctionSorcery
    EventDepletion
    EventDiplomaticMarriage
    EventDisjunction
    EventDonation
    EventEarthquake
    EventGift
    EventGoodMoon
    EventGreatMeteor
    EventManaShort
    EventNewMinerals
    EventPiracy
    EventPlague
    EventPopulationBoom
    EventRebellion
)

func (event EventType) IsGood() bool {
    switch event {
        case EventDiplomaticMarriage, EventDonation, EventGoodMoon,
             EventGift, EventNewMinerals, EventPopulationBoom,
             // FIXME: double check on the conjunction events
             EventConjunctionChaos, EventConjunctionNature, EventConjunctionSorcery:
             return true
         default: return false
    }
}

/* an event is something shown on screen to the user, like when a new building is created
 */
type Event struct {
    Type EventType
    BirthYear int // year/turn the event started
    Message string // messages are supposed to come out of eventmsg.lbx, but we mostly just hardcode them
    MessageStop string // for events that end after some time
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

func MakeBadMoonEvent(year int) *Event {
    return &Event{
        Type: EventBadMoon,
        BirthYear: year,
        Message: "Bad Moon! The moon controlling the powers over evil waxes, doubling the power from evil temples.",
        MessageStop: "The bad moon has waned.",
        LbxIndex: 13,
        CityEvent: false,
        IsConjunction: true,
    }
}

func MakeConjunctionChaosEvent(year int) *Event {
    return &Event{
        Type: EventConjunctionChaos,
        BirthYear: year,
        Message: "The rising triad of red stars come together, doubling all power gained from red nodes and halving all others.",
        MessageStop: "The conjunction of chaos has ended.",
        LbxIndex: 14,
        CityEvent: false,
        IsConjunction: true,
    }
}

func MakeConjunctionNatureEvent(year int) *Event {
    return &Event{
        Type: EventConjunctionNature,
        BirthYear: year,
        Message: "The rising triad of green stars come together, doubling all power gained from green nodes and halving all others.",
        MessageStop: "The conjunction of nature has ended.",
        LbxIndex: 15,
        CityEvent: false,
        IsConjunction: true,
    }
}

func MakeConjunctionSorceryEvent(year int) *Event {
    return &Event{
        Type: EventConjunctionSorcery,
        BirthYear: year,
        Message: "The rising triad of blue stars come together, doubling all power gained from blue nodes and halving all others.",
        MessageStop: "The conjunction of sorcery has ended.",
        LbxIndex: 16,
        CityEvent: false,
        IsConjunction: true,
    }
}

func MakeDepletionEvent(year int, bonus data.BonusType, cityName string) *Event {
    return &Event{
        Type: EventDepletion,
        BirthYear: year,
        Message: fmt.Sprintf("Depletion! A %v mine within %v has become depleted and can no longer be mined.", bonus, cityName),
        LbxIndex: 9,
        CityEvent: false,
        IsConjunction: false,
    }
}

func MakeDiplomaticMarriageEvent(year int, city *citylib.City) *Event {
    return &Event{
        Type: EventDiplomaticMarriage,
        BirthYear: year,
        Message: fmt.Sprintf("Diplomatic Marriage! The neutral %v of %v has offered to join your cause.", city.GetSize(), city.Name),
        LbxIndex: 3,
        CityEvent: true,
        IsConjunction: false,
    }
}

func MakeDonationEvent(year int, amount int) *Event {
    return &Event{
        Type: EventDonation,
        BirthYear: year,
        Message: fmt.Sprintf("Donation! A wealthy merchant has decided to support your cause with a contribution of %v gold.", amount),
        LbxIndex: 8,
        CityEvent: false,
        IsConjunction: false,
    }
}

func MakeEarthquakeEvent(year int, cityName string, people int, units int, buildings int) *Event {
    return &Event{
        Type: EventEarthquake,
        BirthYear: year,
        Message: fmt.Sprintf("Earthquake! A violent quake struck %v, killing %v people and %v units, and destroying %v buildings.", cityName, people, units, buildings),
        LbxIndex: 4,
        CityEvent: true,
        IsConjunction: false,
    }
}

/*
    EventGift
    EventGoodMoon
    EventGreatMeteor
    EventManaShort
    EventNewMinerals
    EventPiracy
    EventPlague
    EventPopulationBoom
    EventRebellion
    */


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
