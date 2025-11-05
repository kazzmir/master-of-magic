package game

import (
    "fmt"
    "slices"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
)

type RandomEventType int

const (
    RandomEventNone RandomEventType = iota
    RandomEventBadMoon
    RandomEventConjunctionChaos
    RandomEventConjunctionNature
    RandomEventConjunctionSorcery
    RandomEventDepletion
    RandomEventDiplomaticMarriage
    RandomEventDisjunction
    RandomEventDonation
    RandomEventEarthquake
    RandomEventGift
    RandomEventGoodMoon
    RandomEventGreatMeteor
    RandomEventManaShort
    RandomEventNewMinerals
    RandomEventPiracy
    RandomEventPlague
    RandomEventPopulationBoom
    RandomEventRebellion
)

func AllRandomEvents() []RandomEventType {
    return []RandomEventType{
        RandomEventBadMoon,
        RandomEventConjunctionChaos,
        RandomEventConjunctionNature,
        RandomEventConjunctionSorcery,
        RandomEventDepletion,
        RandomEventDiplomaticMarriage,
        RandomEventDisjunction,
        RandomEventDonation,
        RandomEventEarthquake,
        RandomEventGift,
        RandomEventGoodMoon,
        RandomEventGreatMeteor,
        RandomEventManaShort,
        RandomEventNewMinerals,
        RandomEventPiracy,
        RandomEventPlague,
        RandomEventPopulationBoom,
        RandomEventRebellion,
    }
}

func RandomCityEvents() []RandomEventType {
    return slices.DeleteFunc(AllRandomEvents(), func(event RandomEventType) bool {
        return event.IsCityEvent()
    })
}

func (event RandomEventType) IsGood() bool {
    switch event {
        case RandomEventDiplomaticMarriage, RandomEventDonation, RandomEventGoodMoon,
             RandomEventGift, RandomEventNewMinerals, RandomEventPopulationBoom,
             // FIXME: double check on the conjunction events
             RandomEventConjunctionChaos, RandomEventConjunctionNature, RandomEventConjunctionSorcery:
             return true
         default: return false
    }
}

func (event RandomEventType) IsCityEvent() bool {
    switch event {
        case RandomEventDepletion, RandomEventDiplomaticMarriage, RandomEventEarthquake,
             RandomEventGreatMeteor, RandomEventNewMinerals, RandomEventPlague,
             RandomEventPopulationBoom, RandomEventRebellion: return true
        default: return false
    }
}

/* an event is something shown on screen to the user, like when a new building is created
 */
type RandomEvent struct {
    Type RandomEventType
    BirthYear uint64 // year/turn the event started
    Message string // messages are supposed to come out of eventmsg.lbx, but we mostly just hardcode them
    MessageStop string // for events that end after some time
    LbxIndex int // picture from events.lbx
    Instant bool // true if there is no duration
    IsConjunction bool // only one conjunction event can be active at a time
    TargetCity *citylib.City // if not nil, then this event targets a city
    TargetPlayer *playerlib.Player // if not nil, then this event targets a player
}

func MakeDisjunctionEvent(year uint64) *RandomEvent {
    return &RandomEvent{
        Type: RandomEventDisjunction,
        BirthYear: year,
        Message: "Disjunction! The fabric of magic has been torn asunder destroying all overland enchantments.",
        LbxIndex: 2,
        IsConjunction: false,
        Instant: true,
    }
}

func MakeBadMoonEvent(year uint64) *RandomEvent {
    return &RandomEvent{
        Type: RandomEventBadMoon,
        BirthYear: year,
        Message: "Bad Moon! The moon controlling the powers over evil waxes, doubling the power from evil temples.",
        MessageStop: "The bad moon has waned.",
        LbxIndex: 13,
        IsConjunction: true,
    }
}

func MakeGoodMoonEvent(year uint64) *RandomEvent {
    return &RandomEvent{
        Type: RandomEventGoodMoon,
        BirthYear: year,
        Message: "Good Moon! The moon controlling the powers over good waxes, doubling the power of good temples.",
        MessageStop: "The good moon has waned.",
        LbxIndex: 12,
        IsConjunction: true,
    }
}

func MakeConjunctionChaosEvent(year uint64) *RandomEvent {
    return &RandomEvent{
        Type: RandomEventConjunctionChaos,
        BirthYear: year,
        Message: "The rising triad of red stars come together, doubling all power gained from red nodes and halving all others.",
        MessageStop: "The conjunction of chaos has ended.",
        LbxIndex: 14,
        IsConjunction: true,
    }
}

func MakeConjunctionNatureEvent(year uint64) *RandomEvent {
    return &RandomEvent{
        Type: RandomEventConjunctionNature,
        BirthYear: year,
        Message: "The rising triad of green stars come together, doubling all power gained from green nodes and halving all others.",
        MessageStop: "The conjunction of nature has ended.",
        LbxIndex: 15,
        IsConjunction: true,
    }
}

func MakeConjunctionSorceryEvent(year uint64) *RandomEvent {
    return &RandomEvent{
        Type: RandomEventConjunctionSorcery,
        BirthYear: year,
        Message: "The rising triad of blue stars come together, doubling all power gained from blue nodes and halving all others.",
        MessageStop: "The conjunction of sorcery has ended.",
        LbxIndex: 16,
        IsConjunction: true,
    }
}

func MakeDepletionEvent(year uint64, bonus data.BonusType, cityName string) *RandomEvent {
    return &RandomEvent{
        Type: RandomEventDepletion,
        BirthYear: year,
        Message: fmt.Sprintf("Depletion! A %v mine within %v has become depleted and can no longer be mined.", bonus, cityName),
        LbxIndex: 9,
        IsConjunction: false,
        Instant: true,
    }
}

func MakeDiplomaticMarriageEvent(year uint64, city *citylib.City) *RandomEvent {
    return &RandomEvent{
        Type: RandomEventDiplomaticMarriage,
        BirthYear: year,
        Message: fmt.Sprintf("Diplomatic Marriage! The neutral %v of %v has offered to join your cause.", city.GetSize(), city.Name),
        LbxIndex: 3,
        IsConjunction: false,
        Instant: true,
    }
}

func MakeDonationEvent(year uint64, amount int, target *playerlib.Player) *RandomEvent {
    return &RandomEvent{
        Type: RandomEventDonation,
        BirthYear: year,
        Message: fmt.Sprintf("Donation! A wealthy merchant has decided to support your cause with a contribution of %v gold.", amount),
        LbxIndex: 8,
        IsConjunction: false,
        Instant: true,
        TargetPlayer: target,
    }
}

func MakeEarthquakeEvent(year uint64, cityName string, people int, units int, buildings int) *RandomEvent {
    return &RandomEvent{
        Type: RandomEventEarthquake,
        BirthYear: year,
        Message: fmt.Sprintf("Earthquake! A violent quake struck %v, killing %v people and %v units, and destroying %v buildings.", cityName, people, units, buildings),
        LbxIndex: 4,
        IsConjunction: false,
        Instant: true,
    }
}

func MakeGiftEvent(year uint64, name string, target *playerlib.Player) *RandomEvent {
    return &RandomEvent{
        Type: RandomEventGift,
        BirthYear: year,
        Message: fmt.Sprintf("The Gift! An ancient God has returned, bearing the relic of %v to aid your cause.", name),
        LbxIndex: 1,
        IsConjunction: false,
        Instant: true,
        TargetPlayer: target,
    }
}

func MakeGreatMeteorEvent(year uint64, city string, people int, units int, buildings int) *RandomEvent {
    return &RandomEvent{
        Type: RandomEventGreatMeteor,
        BirthYear: year,
        Message: fmt.Sprintf("A meteor has hit %v killing %v townsfolk and %v units, and destroying %v buildings.", city, people, units, buildings),
        LbxIndex: 0,
        IsConjunction: false,
        Instant: true,
    }
}

func MakeManaShortEvent(year uint64) *RandomEvent {
    return &RandomEvent{
        Type: RandomEventManaShort,
        BirthYear: year,
        Message: "Magic Short! All sources of magical power have been shorted out.",
        MessageStop: "The mana short has ended and magic has returned to normal.",
        LbxIndex: 17,
        IsConjunction: true,
    }
}

func MakeNewMineralsEvent(year uint64, bonus data.BonusType, city *citylib.City) *RandomEvent {
    return &RandomEvent{
        Type: RandomEventNewMinerals,
        BirthYear: year,
        Message: fmt.Sprintf("New Mine! Surveyors find a %v deposit near the %v of %v.", bonus, city.GetSize(), city.Name),
        LbxIndex: 10,
        IsConjunction: false,
        Instant: true,
    }
}

func MakePiracyEvent(year uint64, gold int, target *playerlib.Player) *RandomEvent {
    return &RandomEvent{
        Type: RandomEventPiracy,
        BirthYear: year,
        Message: fmt.Sprintf("Pirates! Pirates have plundered your gold reserve, looting and stealing %v gold.", gold),
        LbxIndex: 5,
        IsConjunction: false,
        Instant: true,
        TargetPlayer: target,
    }
}

func MakePlagueEvent(year uint64, city *citylib.City) *RandomEvent {
    return &RandomEvent{
        Type: RandomEventPlague,
        BirthYear: year,
        Message: fmt.Sprintf("PLAGUE! A virulent plague has broken out in the %v of %v.", city.GetSize(), city.Name),
        MessageStop: fmt.Sprintf("The plague in %v has ended.", city.Name),
        LbxIndex: 6,
        IsConjunction: false,
        TargetCity: city,
    }
}

func MakePopulationBoomEvent(year uint64, city *citylib.City) *RandomEvent {
    return &RandomEvent{
        Type: RandomEventPopulationBoom,
        BirthYear: year,
        Message: fmt.Sprintf("Population Boom! A sudden population boom doubles the population growth rate of the %v of %v.", city.GetSize(), city.Name),
        MessageStop: fmt.Sprintf("The population boom in %v has ended.", city.Name),
        LbxIndex: 11,
        IsConjunction: false,
        TargetCity: city,
    }
}

func MakeRebellionEvent(year uint64, city *citylib.City) *RandomEvent {
    return &RandomEvent{
        Type: RandomEventRebellion,
        BirthYear: year,
        Message: fmt.Sprintf("Rebellion! The %v of %v has rebelled and become a neutral city.", city.GetSize(), city.Name),
        LbxIndex: 7,
        IsConjunction: false,
        Instant: true,
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
