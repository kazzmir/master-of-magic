package player

import (
    "image"

    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
)

type SerializedWizard struct {
    Name string
    Base data.WizardBase
    Retorts []data.Retort
    Books []data.WizardBook
    Race data.Race
    Banner data.BannerType
}

func serializeWizard(wizard setup.WizardCustom) SerializedWizard {
    return SerializedWizard{
        Name: wizard.Name,
        Base: wizard.Base,
        Retorts: wizard.Retorts,
        Books: wizard.Books,
        Race: wizard.Race,
        Banner: wizard.Banner,
    }
}

type SerializedWork struct {
    Location image.Point `json:"location"`
    Progress float64 `json:"progress"`
}

type SerializedPlayer struct {
    ArcanusFog [][]data.FogType `json:"arcanus-fog"`
    MyrrorFog [][]data.FogType `json:"myrror-fog"`
    TaxRate map[string]int `json:"tax-rate"`
    Gold int `json:"gold"`
    Mana int `json:"mana"`
    Human bool `json:"human"`
    Defeated bool `json:"defeated"`
    Fame int `json:"fame"`
    BookOrderSeed1 uint64 `json:"book-order-seed_1"`
    BookOrderSeed2 uint64 `json:"book-order-seed_2"`
    Banished bool
    KnownSpells []string `json:"known-spells"`
    ResearchPoolSpells []string `json:"research-pool-spells"`
    ResearchCandidateSpells []string `json:"research-candidate-spells"`
    CastingSpellPage int `json:"casting-spell-page"`
    GlobalEnchantments []string `json:"global-enchantments"`
    Wizard SerializedWizard `json:"wizard"`
    PowerDistribution PowerDistribution `json:"power-distribution"`
    SpellOfMasteryCost int `json:"spell-of-mastery-cost"`
    CastingSkillPower int `json:"casting-skill-power"`
    RemainingCastingSkill int `json:"remaining-casting-skill"`
    ResearchingSpell string `json:"researching-spell"`
    ResearchProgress int `json:"research-progress"`
    CastingSpell string `json:"casting-spell"`
    CastingSpellProgress int `json:"casting-spell-progress"`
    PowerHistory []WizardPower `json:"power-history,omitempty"`
    RoadWorkArcanus []SerializedWork `json:"road-work-arcanus"`
    RoadWorkMyrror []SerializedWork `json:"road-work-myrror"`
    PurifyWorkArcanus []SerializedWork `json:"purify-work-arcanus"`
    PurifyWorkMyrror []SerializedWork `json:"purify-work-myrror"`

    // TODO
    // PlayerRelations map[*Player]*Relationship
    // HeroPool map[herolib.HeroType]*herolib.Hero
    // Heroes [6]*herolib.Hero
    // VaultEquipment [4]*artifact.Artifact
    // CreateArtifact *artifact.Artifact
    // Units []units.StackUnit
    // Cities map[data.PlanePoint]*citylib.City
}

func serializeFraction(frac fraction.Fraction) map[string]int {
    return map[string]int{
        "n": frac.Numerator,
        "d": frac.Denominator,
    }
}

func serializeWork(work map[image.Point]float64) []SerializedWork {
    out := make([]SerializedWork, 0)

    for location, progress := range work {
        out = append(out, SerializedWork{
            Location: location,
            Progress: progress,
        })
    }

    return out
}

func spellNames(spells spellbook.Spells) []string {
    out := make([]string, 0)

    for _, spell := range spells.Spells {
        out = append(out, spell.Name)
    }

    return out
}

func globalEnchantmentNames(enchantments *set.Set[data.Enchantment]) []string {
    out := make([]string, 0)

    for _, enchantment := range enchantments.Values() {
        out = append(out, enchantment.String())
    }

    return out
}

func SerializePlayer(player *Player) SerializedPlayer {
    return SerializedPlayer{
        ArcanusFog: player.ArcanusFog,
        MyrrorFog: player.MyrrorFog,
        TaxRate: serializeFraction(player.TaxRate),
        Gold: player.Gold,
        Mana: player.Mana,
        Human: player.Human,
        Defeated: player.Defeated,
        Fame: player.Fame,
        BookOrderSeed1: player.BookOrderSeed1,
        BookOrderSeed2: player.BookOrderSeed2,
        Banished: player.Banished,
        KnownSpells: spellNames(player.KnownSpells),
        ResearchPoolSpells: spellNames(player.ResearchPoolSpells),
        ResearchCandidateSpells: spellNames(player.ResearchCandidateSpells),
        CastingSpellPage: player.CastingSpellPage,
        GlobalEnchantments: globalEnchantmentNames(player.GlobalEnchantments),
        Wizard: serializeWizard(player.Wizard),
        PowerDistribution: player.PowerDistribution,
        SpellOfMasteryCost: player.SpellOfMasteryCost,
        CastingSkillPower: player.CastingSkillPower,
        RemainingCastingSkill: player.RemainingCastingSkill,
        ResearchingSpell: player.ResearchingSpell.Name,
        ResearchProgress: player.ResearchProgress,
        CastingSpell: player.CastingSpell.Name,
        CastingSpellProgress: player.CastingSpellProgress,
        PowerHistory: player.PowerHistory,
        RoadWorkArcanus: serializeWork(player.RoadWorkArcanus),
        RoadWorkMyrror: serializeWork(player.RoadWorkMyrror),
        PurifyWorkArcanus: serializeWork(player.PurifyWorkArcanus),
        PurifyWorkMyrror: serializeWork(player.PurifyWorkMyrror),
    }
}
