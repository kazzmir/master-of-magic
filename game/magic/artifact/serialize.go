package artifact

import (
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

type SerializedArtifact struct {
    Type ArtifactType `json:"type"`
    Image int `json:"image"`
    Name string `json:"name"`
    Cost int `json:"cost"`
    Powers []SerializedPower `json:"powers"`
    Requirements []Requirement `json:"requirements"`
}

type SerializedPower struct {
    Type PowerType
    Amount int // for an ability this is the number of books of the Magic needed
    Name string
    Ability data.ItemAbility
    Magic data.MagicType // for abilities

    Spell string
    SpellCharges int

    // powers are sorted by how they are defined in itempow.lbx, so we just use that number here
    // this field has no utility other than sorting
    Index int
}

func serializePowers(powers []Power) []SerializedPower {
    serialized := make([]SerializedPower, 0, len(powers))

    for _, power := range powers {
        serialized = append(serialized, SerializedPower{
            Type: power.Type,
            Amount: power.Amount,
            Name: power.Name,
            Ability: power.Ability,
            Magic: power.Magic,
            Spell: power.Spell.Name,
            SpellCharges: power.SpellCharges,
            Index: power.Index,
        })
    }

    return serialized
}

func SerializeArtifact(artifact *Artifact) SerializedArtifact {
    return SerializedArtifact{
        Type: artifact.Type,
        Image: artifact.Image,
        Name: artifact.Name,
        Cost: artifact.Cost,
        Powers: serializePowers(artifact.Powers),
        Requirements: append(make([]Requirement, 0), artifact.Requirements...),
    }
}
