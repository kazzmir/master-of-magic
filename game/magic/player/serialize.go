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
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
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

func reconstructWizard(serialized SerializedWizard) setup.WizardCustom {
    return setup.WizardCustom{
        Name: serialized.Name,
        Base: serialized.Base,
        Retorts: serialized.Retorts,
        Books: serialized.Books,
        Race: serialized.Race,
        Banner: serialized.Banner,
    }
}

type SerializedWork struct {
    Location image.Point `json:"location"`
    Progress float64 `json:"progress"`
}

// a unit in a stack can either be a regular unit or a reference to a hero
type SerializedUnitStackElement struct {
    Unit *units.SerializedOverworldUnit `json:"unit,omitempty"`
    Hero *herolib.HeroType `json:"hero,omitempty"`
}

type SerializedUnitStack struct {
    Units []SerializedUnitStackElement `json:"units"`
    Active []bool `json:"active"`

    CurrentPath pathfinding.Path `json:"current-path"`
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
    Stacks []SerializedUnitStack `json:"stacks"`
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

func serializeStacks(stacks []*UnitStack) []SerializedUnitStack {
    out := make([]SerializedUnitStack, 0)

    for _, stack := range stacks {
        serializedStack := SerializedUnitStack{
            CurrentPath: stack.CurrentPath,
        }

        for _, unitRaw := range stack.units {
            var serializedUnit SerializedUnitStackElement

            switch unit := unitRaw.(type) {
                case *units.OverworldUnit:
                    serialized := units.SerializeOverworldUnit(unit)
                    serializedUnit.Unit = &serialized
                case *herolib.Hero:
                    serializedUnit.Hero = &unit.HeroType
            }

            serializedStack.Active = append(serializedStack.Active, stack.IsActive(unitRaw))
            serializedStack.Units = append(serializedStack.Units, serializedUnit)
        }

        out = append(out, serializedStack)
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
        Stacks: serializeStacks(player.Stacks),
        HeroUnits: serializeHeros(player.Heroes[:]),
        VaultEquipment: serializeVaultEquipment(player.VaultEquipment[:]),
        CreateArtifact: serializeCreateArtifact(player.CreateArtifact),
        HeroPool: serializeHeros(slices.Collect(maps.Values(player.HeroPool))),
    }
}

func reconstructSpells(spellNames []string, allSpells spellbook.Spells) spellbook.Spells {
    var out spellbook.Spells

    for _, name := range spellNames {
        spell := allSpells.FindByName(name)
        if spell.Valid() {
            out.AddSpell(spell)
        }
    }

    return out
}

func reconstructEnchantments(enchantmentNames []string) *set.Set[data.Enchantment] {
    out := set.MakeSet[data.Enchantment]()

    for _, name := range enchantmentNames {
        out.Insert(data.GetEnchantmentByName(name))
    }

    return out
}

func reconstructHeroPool(serialized []herolib.SerializedHeroUnit, allSpells spellbook.Spells, globalEnchantmentProvider units.GlobalEnchantmentProvider, experienceInfo units.ExperienceInfo) map[herolib.HeroType]*herolib.Hero {
    out := make(map[herolib.HeroType]*herolib.Hero)

    for _, serializedHero := range serialized {
        hero := herolib.ReconstructHero(&serializedHero, allSpells, globalEnchantmentProvider, experienceInfo)
        out[hero.HeroType] = hero
    }

    return out
}

func reconstructHeroes(serialized []herolib.SerializedHeroUnit, allSpells spellbook.Spells, globalEnchantmentProvider units.GlobalEnchantmentProvider, experienceInfo units.ExperienceInfo) [6]*herolib.Hero {
    var out [6]*herolib.Hero

    for i, serializedHero := range serialized {
        if i < len(out) {
            hero := herolib.ReconstructHero(&serializedHero, allSpells, globalEnchantmentProvider, experienceInfo)
            out[i] = hero
        }
    }

    return out
}

func reconstructEquipment(serialized []artifact.SerializedArtifact, allSpells spellbook.Spells) [4]*artifact.Artifact {
    var out [4]*artifact.Artifact

    for i, serializedArtifact := range serialized {
        if i < len(out) {
            out[i] = artifact.ReconstructArtifact(&serializedArtifact, allSpells)
        }
    }

    return out
}

func reconstructArtifact(serialized *artifact.SerializedArtifact, allSpells spellbook.Spells) *artifact.Artifact {
    if serialized == nil {
        return nil
    }

    return artifact.ReconstructArtifact(serialized, allSpells)
}

func reconstructWork(serialized []SerializedWork) map[image.Point]float64 {
    out := make(map[image.Point]float64)

    for _, work := range serialized {
        out[work.Location] = work.Progress
    }

    return out
}

func reconstructCities(serialized []citylib.SerializedCity, arcanusCatchmentProvider citylib.CatchmentProvider, myrrorCatchmentProvider citylib.CatchmentProvider, cityServices citylib.CityServicesProvider, reignProvider citylib.ReignProvider, buildingInfo buildinglib.BuildingInfos) map[data.PlanePoint]*citylib.City {
    out := make(map[data.PlanePoint]*citylib.City)

    for _, serializedCity := range serialized {
        point := data.PlanePoint{
            Plane: serializedCity.Plane,
            X: serializedCity.X,
            Y: serializedCity.Y,
        }

        catchmentProvider := arcanusCatchmentProvider
        if serializedCity.Plane == data.PlaneMyrror {
            catchmentProvider = myrrorCatchmentProvider
        }

        city := citylib.ReconstructCity(&serializedCity, catchmentProvider, cityServices, reignProvider, buildingInfo)


        out[point] = city
    }

    return out
}

func reconstructStacks(serialized []SerializedUnitStack, heroes [6]*herolib.Hero, globalEnchantmentProvider units.GlobalEnchantmentProvider, experienceInfo units.ExperienceInfo) []*UnitStack {
    var out []*UnitStack

    for _, serializedStack := range serialized {
        stack := MakeUnitStack()

        stack.CurrentPath = serializedStack.CurrentPath

        for i, serializedUnit := range serializedStack.Units {
            if serializedUnit.Unit != nil {
                unit := units.ReconstructOverworldUnit(serializedUnit.Unit, globalEnchantmentProvider, experienceInfo)
                stack.AddUnit(unit)
                stack.SetActive(unit, serializedStack.Active[i])
            } else if serializedUnit.Hero != nil {
                for _, hero := range heroes {
                    if hero != nil && hero.HeroType == *serializedUnit.Hero {
                        stack.AddUnit(hero)
                        stack.SetActive(hero, serializedStack.Active[i])
                        break
                    }
                }
            }
        }

        out = append(out, stack)
    }

    return out
}

// returns a player object and a function to initialize its cities once the maps are available
func ReconstructPlayer(serialized *SerializedPlayer, globalEnchantmentsProvider GlobalEnchantmentsProvider, allSpells spellbook.Spells, buildingInfo buildinglib.BuildingInfos, cityServices citylib.CityServicesProvider) (*Player, func(*maplib.Map, *maplib.Map)) {
    player := &Player{
        ArcanusFog: serialized.ArcanusFog,
        MyrrorFog: serialized.MyrrorFog,
        TaxRate: serialized.TaxRate,
        GlobalEnchantmentsProvider: globalEnchantmentsProvider,
        Human: serialized.Human,

        Gold: serialized.Gold,
        Mana: serialized.Mana,
        Defeated: serialized.Defeated,
        Fame: serialized.Fame,
        BookOrderSeed1: serialized.BookOrderSeed1,
        BookOrderSeed2: serialized.BookOrderSeed2,
        Banished: serialized.Banished,
        CastingSpellPage: serialized.CastingSpellPage,
        PowerDistribution: serialized.PowerDistribution,
        SpellOfMasteryCost: serialized.SpellOfMasteryCost,
        CastingSkillPower: serialized.CastingSkillPower,
        RemainingCastingSkill: serialized.RemainingCastingSkill,
        ResearchProgress: serialized.ResearchProgress,
        CastingSpellProgress: serialized.CastingSpellProgress,
        KnownSpells: reconstructSpells(serialized.KnownSpells, allSpells),
        ResearchPoolSpells: reconstructSpells(serialized.ResearchPoolSpells, allSpells),
        ResearchCandidateSpells: reconstructSpells(serialized.ResearchCandidateSpells, allSpells),
        GlobalEnchantments: reconstructEnchantments(serialized.GlobalEnchantments),
        ResearchingSpell: allSpells.FindByName(serialized.ResearchingSpell),
        CastingSpell: allSpells.FindByName(serialized.CastingSpell),
        Wizard: reconstructWizard(serialized.Wizard),
        VaultEquipment: reconstructEquipment(serialized.VaultEquipment, allSpells),
        CreateArtifact: reconstructArtifact(serialized.CreateArtifact, allSpells),
        PowerHistory: serialized.PowerHistory,
        RoadWorkArcanus: reconstructWork(serialized.RoadWorkArcanus),
        RoadWorkMyrror: reconstructWork(serialized.RoadWorkMyrror),
        PurifyWorkArcanus: reconstructWork(serialized.PurifyWorkArcanus),
        PurifyWorkMyrror: reconstructWork(serialized.PurifyWorkMyrror),

        /*
        // relations with other players (treaties, etc)
        PlayerRelations map[*Player]*Relationship

        */
    }

    player.HeroPool = reconstructHeroPool(serialized.HeroPool, allSpells, player.MakeUnitEnchantmentProvider(), player.MakeExperienceInfo())
    player.Heroes = reconstructHeroes(serialized.HeroUnits, allSpells, player.MakeUnitEnchantmentProvider(), player.MakeExperienceInfo())

    player.Stacks = reconstructStacks(serialized.Stacks, player.Heroes, player.MakeUnitEnchantmentProvider(), player.MakeExperienceInfo())

    initializeCities := func(arcanusMap *maplib.Map, myrrorMap *maplib.Map) {
        player.Cities = reconstructCities(serialized.Cities, arcanusMap, myrrorMap, cityServices, player, buildingInfo)
    }

    return player, initializeCities
}
