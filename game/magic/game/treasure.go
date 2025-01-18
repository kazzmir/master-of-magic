package main

type TreasureItem interface {
}

type TreasureGold struct {
}

type TreasureMana struct {
}

type TreasureMagicalItem struct {
}

type TreasurePrisonerHero struct {
}

type TreasureSpell struct {
}

type TreasureSpellbook struct {
}

type TreasureRetort struct {
}

type Treasure struct {
    Treasures []TreasureItem
}

func makeTreasure() *Treasure {
    return &Treasure{}
}
