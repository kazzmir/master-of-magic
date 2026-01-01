package hero

import (
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

type SerializedAbility struct {
    Type data.AbilityType `json:"type"`
    Value float32 `json:"value"`
}

type SerializedHeroUnit struct {
    Base units.SerializedOverworldUnit `json:"base"`
    HeroType HeroType `json:"hero_type"`
    Name string `json:"name"`
    Status HeroStatus `json:"status"`

    // set at start of game
    Abilities []SerializedAbility `json:"abilities"`

    // Equipment [3]*artifact.Artifact
}

func serializeAbilities(abilities map[data.AbilityType]data.Ability) []SerializedAbility {
    serialized := make([]SerializedAbility, 0, len(abilities))
    for abilityType, ability := range abilities {
        serialized = append(serialized, SerializedAbility{
            Type: abilityType,
            Value: ability.Value,
        })
    }
    return serialized
}

func SerializeHero(hero *Hero) SerializedHeroUnit {

    return SerializedHeroUnit{
        Base: units.SerializeOverworldUnit(hero.OverworldUnit),
        HeroType: hero.HeroType,
        Name: hero.Name,
        Status: hero.Status,
        Abilities: serializeAbilities(hero.Abilities),
        // Equipment: hero.Equipment,
    }
}

