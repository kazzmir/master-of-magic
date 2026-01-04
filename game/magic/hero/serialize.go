package hero

import (
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
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

    Equipment []artifact.SerializedArtifact `json:"items"`
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

func serializeArtifacts(artifacts []*artifact.Artifact) []artifact.SerializedArtifact {
    out := make([]artifact.SerializedArtifact, 0, len(artifacts))

    for _, art := range artifacts {
        // we have to maintain the order of artifacts including empty slots
        if art == nil {
            out = append(out, artifact.SerializedArtifact{})
        } else {
            out = append(out, artifact.SerializeArtifact(art))
        }
    }

    return out
}

func SerializeHero(hero *Hero) SerializedHeroUnit {

    return SerializedHeroUnit{
        Base: units.SerializeOverworldUnit(hero.OverworldUnit),
        HeroType: hero.HeroType,
        Name: hero.Name,
        Status: hero.Status,
        Abilities: serializeAbilities(hero.Abilities),
        Equipment: serializeArtifacts(hero.Equipment[:]),
    }
}

func reconstructAbilities(serialized []SerializedAbility) map[data.AbilityType]data.Ability {
    out := make(map[data.AbilityType]data.Ability)

    for _, ability := range serialized {
        out[ability.Type] = data.Ability{
            Ability: ability.Type,
            Value: ability.Value,
        }
    }

    return out
}

func reconstructArtifacts(serialized []artifact.SerializedArtifact, allSpells spellbook.Spells) [3]*artifact.Artifact {
    var out [3]*artifact.Artifact

    for i, art := range serialized {
        if art.Type == artifact.ArtifactTypeNone {
            continue
        }

        if i < len(out) {
            out[i] = artifact.ReconstructArtifact(&art, allSpells)
        }
    }

    return out
}

func ReconstructHero(serialized *SerializedHeroUnit, allSpells spellbook.Spells, globalEnchantmentProvider units.GlobalEnchantmentProvider, experienceInfo units.ExperienceInfo) *Hero {
    hero := &Hero{
        Name: serialized.Name,
        HeroType: serialized.HeroType,
        Status: serialized.Status,
        Abilities: reconstructAbilities(serialized.Abilities),
        Equipment: reconstructArtifacts(serialized.Equipment, allSpells),
        OverworldUnit: units.ReconstructOverworldUnit(&serialized.Base, globalEnchantmentProvider, experienceInfo),
    }

    hero.OverworldUnit.SetParent(hero)

    return hero
}
