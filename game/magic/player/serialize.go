package player

import (
    "image"
    "maps"
    "slices"

    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
)

type SerializedWizard struct {
    Name string `json:"name"`
    Base data.WizardBase `json:"base"`
    Retorts []data.Retort `json:"retorts"`
    Books []data.WizardBook `json:"books"`
    Race data.Race `json:"race"`
    Banner data.BannerType `json:"banner"`
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
    TaxRate fraction.Fraction `json:"tax-rate"`
    Gold int `json:"gold"`
    Mana int `json:"mana"`
    Human bool `json:"human"`
    Defeated bool `json:"defeated"`
    Fame int `json:"fame"`
    BookOrderSeed1 uint64 `json:"book-order-seed_1"`
    BookOrderSeed2 uint64 `json:"book-order-seed_2"`
    Banished bool `json:"banished"`
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
    Cities []citylib.SerializedCity `json:"cities"`
    NormalUnits []units.SerializedOverworldUnit `json:"units"`
    HeroUnits []herolib.SerializedHeroUnit `json:"hero-units"`

    VaultEquipment []artifact.SerializedArtifact `json:"vault-equipment"`
    CreateArtifact *artifact.SerializedArtifact `json:"create-artifact,omitempty"`

    HeroPool []herolib.SerializedHeroUnit `json:"hero-pool"`

    // TODO
    // PlayerRelations map[*Player]*Relationship
}

func serializeHeros(heroes []*herolib.Hero) []herolib.SerializedHeroUnit {
    out := make([]herolib.SerializedHeroUnit, 0)
    for _, hero := range heroes {
        if hero != nil {
            out = append(out, herolib.SerializeHero(hero))
        }
    }
    return out
}

func serializeUnits(stackUnits []units.StackUnit) []units.SerializedOverworldUnit {
    out := make([]units.SerializedOverworldUnit, 0)

    for _, unit := range stackUnits {
        overworldUnit, ok := unit.(*units.OverworldUnit)
        if ok {
            out = append(out, units.SerializeOverworldUnit(overworldUnit))
        }
    }

    return out
}

func serializeCities(cities map[data.PlanePoint]*citylib.City) []citylib.SerializedCity {
    out := make([]citylib.SerializedCity, 0)

    for _, city := range cities {
        out = append(out, citylib.SerializeCity(city))
    }

    return out
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

func serializeVaultEquipment(artifacts []*artifact.Artifact) []artifact.SerializedArtifact {
    out := make([]artifact.SerializedArtifact, 0)

    for _, art := range artifacts {
        if art != nil {
            out = append(out, artifact.SerializeArtifact(art))
        }
    }

    return out
}

func serializeCreateArtifact(art *artifact.Artifact) *artifact.SerializedArtifact {
    if art == nil {
        return nil
    }

    serialized := artifact.SerializeArtifact(art)
    return &serialized
}

func SerializePlayer(player *Player) SerializedPlayer {
    return SerializedPlayer{
        ArcanusFog: player.ArcanusFog,
        MyrrorFog: player.MyrrorFog,
        TaxRate: player.TaxRate,
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
        Cities: serializeCities(player.Cities),
        NormalUnits: serializeUnits(player.Units),
        HeroUnits: serializeHeros(player.Heroes[:]),
        VaultEquipment: serializeVaultEquipment(player.VaultEquipment[:]),
        CreateArtifact: serializeCreateArtifact(player.CreateArtifact),
        HeroPool: serializeHeros(slices.Collect(maps.Values(player.HeroPool))),
    }
}

func ReconstructPlayer(serialized *SerializedPlayer, globalEnchantmentsProvider GlobalEnchantmentsProvider) *Player {
    player := &Player{
        ArcanusFog: serialized.ArcanusFog,
        MyrrorFog: serialized.MyrrorFog,
        TaxRate: serialized.TaxRate,
        GlobalEnchantmentsProvider: globalEnchantmentsProvider,
        GlobalEnchantments: set.MakeSet[data.Enchantment](),
        Human: serialized.Human,
    }

    return player
}
