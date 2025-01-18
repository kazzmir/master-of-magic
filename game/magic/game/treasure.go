package game

import (
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

type TreasureItem interface {
}

type TreasureGold struct {
    Amount int
}

type TreasureMana struct {
    Amount int
}

type TreasureMagicalItem struct {
    Artifact artifact.Artifact
}

type TreasurePrisonerHero struct {
    Name string
}

type TreasureSpell struct {
    Spell spellbook.Spell
}

// always gives one spellbook of the specified magic type
type TreasureSpellbook struct {
    Magic data.MagicType
}

type TreasureRetort struct {
    Retort setup.WizardAbility
}

type Treasure struct {
    Treasures []TreasureItem
}

func makeTreasure() *Treasure {
    return &Treasure{}
}
