package player

import (
    "image"

    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
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
    TaxRate fraction.Fraction `json:"tax-rate"`
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
    Cities []citylib.SerializedCity `json:"cities"`
    NormalUnits []SerializedOverworldUnit `json:"units"`
    HeroUnits []SerializedHeroUnit `json:"hero-units"`

    // TODO
    // PlayerRelations map[*Player]*Relationship
    // HeroPool map[herolib.HeroType]*herolib.Hero
    // Heroes [6]*herolib.Hero
    // VaultEquipment [4]*artifact.Artifact
    // CreateArtifact *artifact.Artifact
    // Units []units.StackUnit
    // Cities map[data.PlanePoint]*citylib.City
}

type SerializedHeroUnit struct {
}

type SerializedOverworldUnit struct {
    Unit units.SerializedUnit `json:"unit"`
    MovesUsed fraction.Fraction `json:"moves-used"`
    Banner data.BannerType `json:"banner"`
    Plane data.Plane `json:"plane"`
    X int `json:"x"`
    Y int `json:"y"`
    Damage int `json:"damage"`
    Experience int `json:"experience"`
    WeaponBonus data.WeaponBonus `json:"weapon-bonus"`
    Undead bool `json:"undead"`

    Busy units.BusyStatus `json:"busy"`

    // for engineers to follow
    BuildRoadPath pathfinding.Path `json:"build-road-path"`

    Enchantments []data.UnitEnchantment `json:"enchantments"`
}

func serializeUnits(stackUnits []units.StackUnit) []SerializedOverworldUnit {
    out := make([]SerializedOverworldUnit, 0)

    for _, unit := range stackUnits {
        overworldUnit, ok := unit.(*units.OverworldUnit)
        if ok {
            out = append(out, SerializedOverworldUnit{
                Unit: units.SerializeUnit(overworldUnit.Unit),
                MovesUsed: overworldUnit.MovesUsed,
                Banner: overworldUnit.Banner,
                Plane: overworldUnit.Plane,
                X: overworldUnit.X,
                Y: overworldUnit.Y,
                Damage: overworldUnit.Damage,
                Experience: overworldUnit.Experience,
                WeaponBonus: overworldUnit.WeaponBonus,
                Undead: overworldUnit.Undead,
                Busy: overworldUnit.Busy,
                BuildRoadPath: append(make(pathfinding.Path, 0), overworldUnit.BuildRoadPath...),
                Enchantments: append(make([]data.UnitEnchantment, 0), overworldUnit.Enchantments...),
            })
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
    }
}
