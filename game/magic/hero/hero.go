package hero

import (
    "fmt"
    "slices"
    "cmp"
    "math/rand/v2"
    "math"

    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/lib/fraction"
)

type HeroStatus int
const (
    StatusAvailable HeroStatus = iota
    StatusEmployed
    StatusDead
)

type HeroType int
const (
    HeroTorin HeroType = iota
    HeroFang
    HeroBShan
    HeroMorgana
    HeroWarrax
    HeroMysticX
    HeroBahgtru
    HeroDethStryke
    HeroSpyder
    HeroSirHarold
    HeroBrax
    HeroRavashack
    HeroGreyfairer
    HeroShalla
    HeroRoland
    HeroMalleus
    HeroMortu
    HeroGunther
    HeroRakir
    HeroJaer
    HeroTaki
    HeroYramrag
    HeroValana
    HeroElana
    HeroAerie
    HeroMarcus
    HeroReywind
    HeroAlorra
    HeroZaldron
    HeroShinBo
    HeroSerena
    HeroShuri
    HeroTheria
    HeroTumu
    HeroAureus

    HeroNone
)

func (heroType HeroType) GetRequiredFame() int {
    switch heroType {
        case HeroTorin: return 0
        case HeroFang: return 10
        case HeroBShan: return 0
        case HeroMorgana: return 10
        case HeroWarrax: return 40
        case HeroMysticX: return 20
        case HeroBahgtru: return 0
        case HeroDethStryke: return 40
        case HeroSpyder: return 20
        case HeroSirHarold: return 40
        case HeroBrax: return 0
        case HeroRavashack: return 40
        case HeroGreyfairer: return 5
        case HeroShalla: return 20
        case HeroRoland: return 40
        case HeroMalleus: return 5
        case HeroMortu: return 40
        case HeroGunther: return 0
        case HeroRakir: return 0
        case HeroJaer: return 10
        case HeroTaki: return 5
        case HeroYramrag: return 20
        case HeroValana: return 0
        case HeroElana: return 40
        case HeroAerie: return 40
        case HeroMarcus: return 10
        case HeroReywind: return 5
        case HeroAlorra: return 40
        case HeroZaldron: return 0
        case HeroShinBo: return 20
        case HeroSerena: return 0
        case HeroShuri: return 0
        case HeroTheria: return 0
        case HeroTumu: return 5
        case HeroAureus: return 10
    }

    return 0
}

func (heroType HeroType) RandomAbilityCount() int {
    switch heroType {
        case HeroTorin: return 2
        case HeroFang: return 2
        case HeroBShan: return 0
        case HeroMorgana: return 2
        case HeroWarrax: return 3
        case HeroMysticX: return 5
        case HeroBahgtru: return 1
        case HeroDethStryke: return 1
        case HeroSpyder: return 1
        case HeroSirHarold: return 1
        case HeroBrax: return 0
        case HeroRavashack: return 2
        case HeroGreyfairer: return 0
        case HeroShalla: return 1
        case HeroRoland: return 1
        case HeroMalleus: return 1
        case HeroMortu: return 1
        case HeroGunther: return 0
        case HeroRakir: return 0
        case HeroJaer: return 1
        case HeroTaki: return 1
        case HeroYramrag: return 1
        case HeroValana: return 0
        case HeroElana: return 0
        case HeroAerie: return 2
        case HeroMarcus: return 0
        case HeroReywind: return 1
        case HeroAlorra: return 3
        case HeroZaldron: return 0
        case HeroShinBo: return 2
        case HeroSerena: return 1
        case HeroShuri: return 1
        case HeroTheria: return 0
        case HeroTumu: return 1
        case HeroAureus: return 2
    }

    return 0
}

func (heroType HeroType) RandomAbilityType() abilityChoice {
    switch heroType {
        case HeroTorin: return abilityChoiceAny
        case HeroFang: return abilityChoiceFighter
        case HeroBShan: return abilityChoiceAny
        case HeroMorgana: return abilityChoiceMage
        case HeroWarrax: return abilityChoiceAny
        case HeroMysticX: return abilityChoiceAny
        case HeroBahgtru: return abilityChoiceFighter
        case HeroDethStryke: return abilityChoiceFighter
        case HeroSpyder: return abilityChoiceFighter
        case HeroSirHarold: return abilityChoiceFighter
        case HeroBrax: return abilityChoiceAny
        case HeroRavashack: return abilityChoiceMage
        case HeroGreyfairer: return abilityChoiceAny
        case HeroShalla: return abilityChoiceFighter
        case HeroRoland: return abilityChoiceFighter
        case HeroMalleus: return abilityChoiceMage
        case HeroMortu: return abilityChoiceFighter
        case HeroGunther: return abilityChoiceAny
        case HeroRakir: return abilityChoiceAny
        case HeroJaer: return abilityChoiceMage
        case HeroTaki: return abilityChoiceFighter
        case HeroYramrag: return abilityChoiceMage
        case HeroValana: return abilityChoiceAny
        case HeroElana: return abilityChoiceAny
        case HeroAerie: return abilityChoiceMage
        case HeroMarcus: return abilityChoiceAny
        case HeroReywind: return abilityChoiceAny
        case HeroAlorra: return abilityChoiceAny
        case HeroZaldron: return abilityChoiceAny
        case HeroShinBo: return abilityChoiceFighter
        case HeroSerena: return abilityChoiceMage
        case HeroShuri: return abilityChoiceFighter
        case HeroTheria: return abilityChoiceAny
        case HeroTumu: return abilityChoiceFighter
        case HeroAureus: return abilityChoiceAny
    }

    return abilityChoiceAny
}

func (heroType HeroType) GetUnit() units.Unit {
    switch heroType {
        case HeroTorin: return units.HeroTorin
        case HeroFang: return units.HeroFang
        case HeroBShan: return units.HeroBShan
        case HeroMorgana: return units.HeroMorgana
        case HeroWarrax: return units.HeroWarrax
        case HeroMysticX: return units.HeroMysticX
        case HeroBahgtru: return units.HeroBahgtru
        case HeroDethStryke: return units.HeroDethStryke
        case HeroSpyder: return units.HeroSpyder
        case HeroSirHarold: return units.HeroSirHarold
        case HeroBrax: return units.HeroBrax
        case HeroRavashack: return units.HeroRavashack
        case HeroGreyfairer: return units.HeroGreyfairer
        case HeroShalla: return units.HeroShalla
        case HeroRoland: return units.HeroRoland
        case HeroMalleus: return units.HeroMalleus
        case HeroMortu: return units.HeroMortu
        case HeroGunther: return units.HeroGunther
        case HeroRakir: return units.HeroRakir
        case HeroJaer: return units.HeroJaer
        case HeroTaki: return units.HeroTaki
        case HeroYramrag: return units.HeroYramrag
        case HeroValana: return units.HeroValana
        case HeroElana: return units.HeroElana
        case HeroAerie: return units.HeroAerie
        case HeroMarcus: return units.HeroMarcus
        case HeroReywind: return units.HeroReywind
        case HeroAlorra: return units.HeroAlorra
        case HeroZaldron: return units.HeroZaldron
        case HeroShinBo: return units.HeroShinBo
        case HeroSerena: return units.HeroSerena
        case HeroShuri: return units.HeroShuri
        case HeroTheria: return units.HeroTheria
        case HeroTumu: return units.HeroTumu
        case HeroAureus: return units.HeroAureus
    }
    return units.HeroRakir
}

func (heroType HeroType) DefaultName() string {
    return heroType.GetUnit().Name
}

func AllHeroTypes() []HeroType {
    return []HeroType{
        HeroTorin,
        HeroFang,
        HeroBShan,
        HeroMorgana,
        HeroWarrax,
        HeroMysticX,
        HeroBahgtru,
        HeroDethStryke,
        HeroSpyder,
        HeroSirHarold,
        HeroBrax,
        HeroRavashack,
        HeroGreyfairer,
        HeroShalla,
        HeroRoland,
        HeroMalleus,
        HeroMortu,
        HeroGunther,
        HeroRakir,
        HeroJaer,
        HeroTaki,
        HeroYramrag,
        HeroValana,
        HeroElana,
        HeroAerie,
        HeroMarcus,
        HeroReywind,
        HeroAlorra,
        HeroZaldron,
        HeroShinBo,
        HeroSerena,
        HeroShuri,
        HeroTheria,
        HeroTumu,
        HeroAureus,
    }
}

type Hero struct {
    *units.OverworldUnit
    HeroType HeroType
    Name string
    Status HeroStatus

    // set at start of game
    Abilities map[data.AbilityType]data.Ability

    Equipment [3]*artifact.Artifact
}

func MakeHeroSimple(heroType HeroType) *Hero {
    unit := units.MakeOverworldUnit(heroType.GetUnit(), 0, 0, data.PlaneArcanus)
    unit.ExperienceInfo = &units.NoExperienceInfo{}
    unit.GlobalEnchantments = &units.NoEnchantments{}
    return MakeHero(unit, heroType, heroType.DefaultName())
}

func abilitiesToMap(abilities []data.Ability) map[data.AbilityType]data.Ability {
    out := make(map[data.AbilityType]data.Ability)

    for _, ability := range abilities {
        out[ability.Ability] = ability
    }

    return out
}

func MakeHero(unit *units.OverworldUnit, heroType HeroType, name string) *Hero {
    hero := &Hero{
        OverworldUnit: unit,
        Name: name,
        HeroType: heroType,
        Abilities: abilitiesToMap(unit.GetAbilities()),
        Status: StatusAvailable,
    }
    hero.SetParent(hero)
    return hero
}

type abilityChoice int
const (
    abilityChoiceFighter abilityChoice = iota
    abilityChoiceMage
    abilityChoiceAny
)

func selectAbility(kind abilityChoice) data.AbilityType {
    anyChoices := []data.AbilityType{
        data.AbilityCharmed,
        data.AbilityLucky,
        data.AbilityNoble,
    }

    fighterChoices := []data.AbilityType{
        data.AbilityAgility,
        data.AbilityArmsmaster,
        data.AbilityBlademaster,
        data.AbilityConstitution,
        data.AbilityLeadership,
        data.AbilityLegendary,
        data.AbilityMight,
    }

    mageChoices := []data.AbilityType{
        data.AbilityArcanePower,
        data.AbilityCaster,
        data.AbilityPrayermaster,
        data.AbilitySage,
    }

    var use []data.AbilityType
    switch kind {
        case abilityChoiceFighter:
            use = append(fighterChoices, anyChoices...)
        case abilityChoiceMage:
            use = append(mageChoices, anyChoices...)
        case abilityChoiceAny:
            use = append(append(fighterChoices, mageChoices...), anyChoices...)
    }

    return use[rand.N(len(use))]
}

func superVersion(ability data.AbilityType) data.AbilityType {
    switch ability {
        case data.AbilityAgility: return data.AbilitySuperAgility
        case data.AbilityArmsmaster: return data.AbilitySuperArmsmaster
        case data.AbilityBlademaster: return data.AbilitySuperBlademaster
        case data.AbilityConstitution: return data.AbilitySuperConstitution
        case data.AbilityLeadership: return data.AbilitySuperLeadership
        case data.AbilityLegendary: return data.AbilitySuperLegendary
        case data.AbilityMight: return data.AbilitySuperMight
        case data.AbilityArcanePower: return data.AbilitySuperArcanePower
        case data.AbilityPrayermaster: return data.AbilitySuperPrayermaster
        case data.AbilitySage: return data.AbilitySuperSage
    }

    return data.AbilityNone
}

// returns true if the ability is added. some abilities cannot be added in case the
// hero already has a super version of that ability, or the limit of 1 is reached for others
func (hero *Hero) AddAbility(ability data.AbilityType) bool {
    limit1 := []data.AbilityType{data.AbilityCharmed, data.AbilityLucky, data.AbilityNoble}

    if slices.Contains(limit1, ability) && hero.HasAbility(ability) {
        return false
    }

    if hero.HasAbility(superVersion(ability)) {
        return false
    }

    if ability == data.AbilityCaster {
        old := hero.Abilities[ability]
        hero.Abilities[ability] = data.MakeAbilityValue(ability, old.Value + 2.5)
        return true
    }

    // upgrade from regular ability to super version
    if hero.HasAbility(ability) {
        delete(hero.Abilities, ability)
        newAbility := superVersion(ability)
        hero.Abilities[newAbility] = data.MakeAbility(newAbility)
    } else {
        hero.Abilities[ability] = data.MakeAbility(ability)
    }

    return true
}

// add N random abilities
func (hero *Hero) SetExtraAbilities() {
    // totalLoops := 0
    for range hero.HeroType.RandomAbilityCount() {

        // this loop could run for a while, so possibly have some way to force ability selection
        // to be determinstic rather than purely random (such as removing abilities that cannot be chosen)
        for {
            // totalLoops += 1
            randomAbility := selectAbility(hero.HeroType.RandomAbilityType())
            if hero.AddAbility(randomAbility) {
                break
            }
        }
    }

    // fmt.Printf("Hero %v took %v loops: %v\n", hero.ShortName(), totalLoops, hero.GetAbilities())
}

func (hero *Hero) IsFemale() bool {
    switch hero.HeroType {
        case HeroValana, HeroSerena, HeroShuri, HeroTheria, HeroMorgana,
             HeroShalla, HeroElana, HeroAlorra: return true
    }

    return false
}

func (hero *Hero) IsChampion() bool {
    switch hero.HeroType {
        case HeroTorin, HeroWarrax, HeroMysticX, HeroDethStryke, HeroSirHarold,
             HeroRavashack, HeroRoland, HeroMortu, HeroElana, HeroAerie,
             HeroAlorra: return true

        case HeroAureus, HeroBahgtru, HeroTumu, HeroTheria, HeroBrax,
             HeroBShan, HeroFang, HeroGreyfairer, HeroGunther, HeroJaer,
             HeroMalleus, HeroMarcus, HeroMorgana, HeroRakir, HeroReywind,
             HeroSerena, HeroShalla, HeroShinBo, HeroShuri, HeroSpyder,
             HeroTaki, HeroValana, HeroYramrag, HeroZaldron: return false
    }

    return false
}

func (hero *Hero) SetStatus(status HeroStatus) {
    hero.Status = status
}

func (hero *Hero) GetName() string {
    return hero.Name
}

func (hero *Hero) SetName(name string) {
    hero.Name = name
}

func (hero *Hero) GetFullName() string {
    return fmt.Sprintf("%v the %v", hero.GetName(), hero.GetTitle())
}

func (hero *Hero) GetPortraitLbxInfo() (string, int) {
    lbxFile := "portrait.lbx"
    switch hero.HeroType {
        case HeroTorin: return lbxFile, 0
        case HeroFang: return lbxFile, 1
        case HeroBShan: return lbxFile, 2
        case HeroMorgana: return lbxFile, 3
        case HeroWarrax: return lbxFile, 4
        case HeroMysticX: return lbxFile, 5
        case HeroBahgtru: return lbxFile, 6
        case HeroDethStryke: return lbxFile, 7
        case HeroSpyder: return lbxFile, 8
        case HeroSirHarold: return lbxFile, 9
        case HeroBrax: return lbxFile, 10
        case HeroRavashack: return lbxFile, 11
        case HeroGreyfairer: return lbxFile, 12
        case HeroShalla: return lbxFile, 13
        case HeroRoland: return lbxFile, 14
        case HeroMalleus: return lbxFile, 15
        case HeroMortu: return lbxFile, 16
        case HeroGunther: return lbxFile, 17
        case HeroRakir: return lbxFile, 18
        case HeroJaer: return lbxFile, 19
        case HeroTaki: return lbxFile, 20
        case HeroYramrag: return lbxFile, 21
        case HeroValana: return lbxFile, 22
        case HeroElana: return lbxFile, 23
        case HeroAerie: return lbxFile, 24
        case HeroMarcus: return lbxFile, 25
        case HeroReywind: return lbxFile, 26
        case HeroAlorra: return lbxFile, 27
        case HeroZaldron: return lbxFile, 28
        case HeroShinBo: return lbxFile, 29
        case HeroSerena: return lbxFile, 30
        case HeroShuri: return lbxFile, 31
        case HeroTheria: return lbxFile, 32
        case HeroTumu: return lbxFile, 33
        case HeroAureus: return lbxFile, 34
    }

    return "", -1
}

func (hero *Hero) GetRequiredFame() int {
    return hero.HeroType.GetRequiredFame()
}

// fee is halved if the hiring wizard is charismatic, handle that elsewhere
func (hero *Hero) GetHireFee() int {
    base := 100 + hero.HeroType.GetRequiredFame() * 10

    levelInt := 1

    level := hero.GetHeroExperienceLevel()
    switch level {
        case units.ExperienceHero: levelInt = 1
        case units.ExperienceMyrmidon: levelInt = 2
        case units.ExperienceCaptain: levelInt = 3
        case units.ExperienceCommander: levelInt = 4
        case units.ExperienceChampionHero: levelInt = 5
        case units.ExperienceLord: levelInt = 6
        case units.ExperienceGrandLord: levelInt = 7
        case units.ExperienceSuperHero: levelInt = 8
        case units.ExperienceDemiGod: levelInt = 9
    }

    return base * (3 + levelInt) / 4
}

func (hero *Hero) AdjustHealth(amount int) {
    hero.OverworldUnit.Damage -= amount
    if hero.OverworldUnit.Damage < 0 {
        hero.OverworldUnit.Damage = 0
    }
    if hero.OverworldUnit.Damage > hero.GetMaxHealth() {
        hero.OverworldUnit.Damage = hero.GetMaxHealth()
    }

    if hero.GetHealth() <= 0 {
        hero.SetStatus(StatusDead)
    }
}

func (hero *Hero) HasItem(itemType artifact.ArtifactType) bool {
    for _, item := range hero.Equipment {
        if item != nil && item.Type == itemType {
            return true
        }
    }

    return false
}

func (hero *Hero) CanTouchAttack(damage units.Damage) bool {
    switch damage {
        case units.DamageMeleePhysical:
            return hero.HasItem(artifact.ArtifactTypeSword) || hero.HasItem(artifact.ArtifactTypeAxe) || hero.HasItem(artifact.ArtifactTypeMace)
        case units.DamageRangedMagical, units.DamageRangedPhysical:
            return hero.HasItem(artifact.ArtifactTypeBow) || hero.HasItem(artifact.ArtifactTypeStaff) || hero.HasItem(artifact.ArtifactTypeWand)
    }

    return false
}

// for mythril/adamantium, heroes dont use those
func (hero *Hero) SetWeaponBonus(bonus data.WeaponBonus) {
}

func (hero *Hero) GetWeaponBonus() data.WeaponBonus {
    return data.WeaponNone
}

func (hero *Hero) GetHealth() int {
    return hero.GetMaxHealth() - hero.GetDamage()
}

func (hero *Hero) GetMaxHealth() int {
    return hero.GetFullHitPoints() * hero.GetCount()
}

func (hero *Hero) AddExperience(amount int) {
    if hero.Status != StatusDead {
        hero.OverworldUnit.AddExperience(amount)
    }
}

func (hero *Hero) GetEnchantments() []data.UnitEnchantment {
    var artifactsEnchantments []data.UnitEnchantment

    for _, item := range hero.Equipment {
        if item != nil {
            artifactsEnchantments = append(artifactsEnchantments, item.GetEnchantments()...)
        }
    }

    return append(hero.OverworldUnit.GetEnchantments(), artifactsEnchantments...)
}

func (hero *Hero) HasEnchantment(enchantment data.UnitEnchantment) bool {
    return hero.OverworldUnit.HasEnchantment(enchantment) || slices.ContainsFunc(hero.Equipment[:], func (a *artifact.Artifact) bool {
        return a != nil && a.HasEnchantment(enchantment)
    })
}

func (hero *Hero) IsHero() bool {
    return true
}

func (hero *Hero) GetSpellChargeSpells() map[spellbook.Spell]int {
    out := make(map[spellbook.Spell]int)

    for _, item := range hero.Equipment {
        if item != nil {
            spell, count := item.GetSpellCharge()
            if count > 0 {
                _, ok := out[spell]
                if !ok {
                    out[spell] = 0
                }

                out[spell] += count
            }
        }
    }

    return out
}

func (hero *Hero) GetKnownSpells() []string {
    switch hero.HeroType {
        case HeroAerie: return []string{"Psionic Blast", "Vertigo", "Mind Storm"}
        case HeroAlorra: return []string{"Resist Magic", "Flight"}
        case HeroElana: return []string{"Dispel Evil", "Healing", "Prayer", "Holy Word"}
        case HeroGreyfairer: return []string{"Ice Bolt", "Petrify", "Web"}
        case HeroJaer: return []string{"Guardian Wind", "Word of Recall"}
        case HeroMalleus: return []string{"Fire Bolt", "Fireball", "Flame Strike", "Fire Elemental"}
        case HeroMarcus: return []string{"Resist Elements", "Stone Skin"}
        case HeroMorgana: return []string{"Darkness", "Possession", "Black Prayer", "Mana Leak"}
        case HeroRakir: return []string{"Resist Elements"}
        case HeroRavashack: return []string{"Weakness", "Black Sleep", "Animate Dead", "Wrack"}
        case HeroReywind: return []string{"Flame Blade", "Shatter", "Eldritch Weapon"}
        case HeroSerena: return []string{"Healing"}
        case HeroTorin: return []string{"Healing", "Holy Armor", "Lionheart"}
        case HeroValana: return []string{"Confusion", "Vertigo"}
        case HeroYramrag: return []string{"Lightning Bolt", "Doom Bolt", "Warp Lightning"}
        case HeroZaldron: return []string{"Counter Magic", "Dispel Magic True"}
    }

    return nil
}

func (hero *Hero) GetToHitMelee() int {
    // FIXME: add in equipment tohit bonuses
    base := 30

    level := hero.GetHeroExperienceLevel()
    switch level {
        case units.ExperienceHero:
        case units.ExperienceMyrmidon:
        case units.ExperienceCaptain: base += 10
        case units.ExperienceCommander: base += 10
        case units.ExperienceChampionHero: base += 10
        case units.ExperienceLord: base += 20
        case units.ExperienceGrandLord: base += 20
        case units.ExperienceSuperHero: base += 20
        case units.ExperienceDemiGod: base += 30
    }

    if hero.HasAbility(data.AbilityLucky) {
        base += 10
    }

    if hero.HasEnchantment(data.UnitEnchantmentHolyWeapon) {
        base += 10
    }

    return base + hero.GetAbilityToHit()
}

func (hero *Hero) GetToDefend() int {
    base := 30
    modifier := 0

    if hero.HasAbility(data.AbilityLucky) {
        modifier += 10
    }

    return base + modifier
}

func (hero *Hero) NaturalHeal(rate float64) {
    if hero.IsUndead() {
        return
    }

    amount := float64(hero.GetMaxHealth()) * rate
    if amount < 1 {
        amount = 1
    }
    hero.AdjustHealth(int(amount))
}

func (hero *Hero) GetMovesLeft(overworld bool) fraction.Fraction {
    return fraction.Zero().Max(hero.GetMovementSpeed(overworld).Subtract(hero.OverworldUnit.MovesUsed))
}

func (hero *Hero) SetMovesLeft(overworld bool, moves fraction.Fraction) {
    hero.OverworldUnit.MovesUsed = hero.GetMovementSpeed(overworld).Subtract(moves)
}

func (hero *Hero) GetCasterValue() float32 {
    caster, ok := hero.Abilities[data.AbilityCaster]
    if !ok || caster.Value == 0 {
        return 0
    }

    level := hero.GetHeroExperienceLevel()
    return caster.Value * float32(level.ToInt() + 1)
}

func (hero *Hero) GetThrownValue() int {
    thrown, ok := hero.Abilities[data.AbilityThrown]
    if !ok || thrown.Value == 0 {
        return 0
    }

    level := hero.GetHeroExperienceLevel()
    return int(thrown.Value) + level.ToInt()
}

func (hero *Hero) GetFireBreathValue() int {
    fireBreath, ok := hero.Abilities[data.AbilityFireBreath]
    if !ok || fireBreath.Value == 0 {
        return 0
    }

    level := hero.GetHeroExperienceLevel()
    return int(fireBreath.Value) + level.ToInt()
}

func (hero *Hero) GetAbilityValue(ability data.AbilityType) float32 {
    // melee bonus applies to thrown and breath attacks
    if ability == data.AbilityThrown {
        value := hero.GetThrownValue()
        if value == 0 {
            return 0
        }

        modifier := float32(0)

        if hero.GetRealm() == data.ChaosMagic && hero.OverworldUnit.GlobalEnchantments.HasEnchantment(data.EnchantmentChaosSurge) {
            modifier += 2
        }

        abilityBonus := hero.GetAbilityLeadership()
        // FIXME: maybe only apply the leadership bonus in combat
        return float32(value) + float32(abilityBonus) / 2 + modifier
    }

    if ability == data.AbilityFireBreath {
        value := hero.GetFireBreathValue()
        if value == 0 {
            return 0
        }

        modifier := float32(0)

        if hero.GetRealm() == data.ChaosMagic && hero.OverworldUnit.GlobalEnchantments.HasEnchantment(data.EnchantmentChaosSurge) {
            modifier += 2
        }

        // FIXME: maybe only apply the leadership bonus in combat
        abilityBonus := hero.GetAbilityLeadership()
        return float32(value) + float32(abilityBonus) / 2 + modifier
    }

    if ability == data.AbilityCaster {
        return hero.GetCasterValue()
    }

    heroAbility, ok := hero.Abilities[ability]
    if ok {
        return heroAbility.Value
    }

    // will likely be 0 except for special cases such as chaos channel fire breath
    return hero.Unit.GetAbilityValue(ability)
}

// FIXME: maybe this method shouldn't exist, and instead just iterate through GetAbilities()
/*
func (hero *Hero) GetAbilityReference(ability data.AbilityType) *data.Ability {
    for i := range len(hero.Abilities) {
        if hero.Abilities[i].Ability == ability {
            return &hero.Abilities[i]
        }
    }

    return nil
}
*/

func (hero *Hero) HasAbility(ability data.AbilityType) bool {
    _, ok := hero.Abilities[ability]
    if ok {
        return true
    }

    for _, item := range hero.Equipment {
        if item != nil && item.HasAbility(ability) {
            return true
        }
    }

    return hero.Unit.HasAbility(ability)
}

func (hero *Hero) HasItemAbility(ability data.ItemAbility) bool {
    return slices.ContainsFunc(hero.Equipment[:], func (a *artifact.Artifact) bool {
        return a != nil && a.HasItemAbility(ability)
    })
}

func (hero *Hero) IsInvisible() bool {
    return hero.HasAbility(data.AbilityInvisibility)
}

func (hero *Hero) IsSailing() bool {
    return false
}

func (hero *Hero) GetCount() int {
    return 1
}

func (hero *Hero) GetVisibleCount() int {
    return 1
}

func (hero *Hero) GetUpkeepGold() int {
    if hero.HasAbility(data.AbilityNoble) {
        return 0
    }

    return hero.OverworldUnit.GetUpkeepGold()
}

func (hero *Hero) GetMovementSpeed(overworld bool) fraction.Fraction {
    base := hero.OverworldUnit.GetBaseMovementSpeed(overworld)

    for _, item := range hero.Equipment {
        if item != nil {
            base += item.MovementBonus()
        }
    }

    return hero.OverworldUnit.MovementSpeedEnchantmentBonus(fraction.FromInt(base), hero.GetEnchantments())
}

func (hero *Hero) GetExperienceData() units.ExperienceData {
    level := hero.GetHeroExperienceLevel()
    return &level
}

func (hero *Hero) GetExperienceLevel() units.NormalExperienceLevel {
    return units.ExperienceRecruit
}

func (hero *Hero) GetHeroExperienceLevel() units.HeroExperienceLevel {

    experience := hero.OverworldUnit.Experience

    if experience < 120 && hero.HasEnchantment(data.UnitEnchantmentHeroism) {
        experience = 120
    }

    if hero.OverworldUnit.ExperienceInfo != nil {
        return units.GetHeroExperienceLevel(experience, hero.OverworldUnit.ExperienceInfo.HasWarlord(), hero.OverworldUnit.ExperienceInfo.Crusade())
    }

    return units.ExperienceHero
}

func (hero *Hero) ResetOwner() {
    hero.SetExperienceInfo(&units.NoExperienceInfo{})
}

// force hero to go up one level
func (hero *Hero) GainLevel(maxLevel units.HeroExperienceLevel) {
    if hero.GetHeroExperienceLevel() >= maxLevel {
        return
    }

    levels := []units.HeroExperienceLevel{
        units.ExperienceHero, units.ExperienceMyrmidon,
        units.ExperienceCaptain, units.ExperienceCommander,
        units.ExperienceChampionHero, units.ExperienceLord,
        units.ExperienceGrandLord, units.ExperienceSuperHero,
        units.ExperienceDemiGod,
    }

    currentLevel := hero.GetHeroExperienceLevel()

    // add just enough experience to make it to the next level
    for i := range len(levels) - 1 {
        if currentLevel == levels[i] {
            hero.AddExperience(levels[i + 1].ExperienceRequired(false, false) - hero.OverworldUnit.GetExperience())
            break
        }
    }
}

func (hero *Hero) GetBaseMeleeAttackPower() int {
    level := hero.GetHeroExperienceLevel()
    return hero.OverworldUnit.GetBaseMeleeAttackPower() + hero.getBaseMeleeAttackPowerProgression(level)
}

func (hero *Hero) getBaseMeleeAttackPowerProgression(level units.HeroExperienceLevel) int {
    switch level {
        case units.ExperienceHero: return 0
        case units.ExperienceMyrmidon: return 1
        case units.ExperienceCaptain: return 2
        case units.ExperienceCommander: return 3
        case units.ExperienceChampionHero: return 4
        case units.ExperienceLord: return 5
        case units.ExperienceGrandLord: return 6
        case units.ExperienceSuperHero: return 7
        case units.ExperienceDemiGod: return 8
    }
    return 0
}

func (hero *Hero) GetMeleeAttackPower() int {
    base := hero.GetBaseMeleeAttackPower()

    for _, item := range hero.Equipment {
        if item != nil {
            base += item.MeleeBonus()
        }
    }

    for _, enchantment := range hero.GetEnchantments() {
        base += hero.MeleeEnchantmentBonus(enchantment)
    }

    if hero.GetRealm() == data.ChaosMagic && hero.OverworldUnit.GlobalEnchantments.HasEnchantment(data.EnchantmentChaosSurge) {
        base += 2
    }

    return base + hero.GetAbilityMelee()
}

// returns a number that corresponds to the bonus this ability would apply.
// Might on a champion would return 5
// Super Agility on a captain would return 4
func (hero *Hero) GetAbilityBonus(ability data.AbilityType) int {
    switch ability {
        case data.AbilityAgility, data.AbilitySuperAgility: return hero.GetAbilityDefense()
        case data.AbilityConstitution, data.AbilitySuperConstitution: return hero.GetAbilityHealth()
        case data.AbilityLeadership, data.AbilitySuperLeadership: return hero.GetAbilityLeadership()
        case data.AbilitySage, data.AbilitySuperSage: return hero.GetAbilityResearch()
        case data.AbilityPrayermaster, data.AbilitySuperPrayermaster: return hero.GetAbilityResistance()
        case data.AbilityArcanePower, data.AbilitySuperArcanePower: return hero.GetAbilityMagicRangedAttack()
        case data.AbilityMight, data.AbilitySuperMight: return hero.GetAbilityMelee() - hero.GetAbilityLeadership() // hack because GetAbilityMelee() includes leadership
        case data.AbilityArmsmaster, data.AbilitySuperArmsmaster: return hero.GetAbilityExperienceBonus()
        case data.AbilityBlademaster, data.AbilitySuperBlademaster: return hero.GetAbilityToHit()
        case data.AbilityLegendary, data.AbilitySuperLegendary: return hero.GetAbilityFame()
    }

    return 0
}

func (hero *Hero) GetBaseRangedAttackPower() int {
    base := hero.Unit.GetBaseRangedAttackPower()
    if base == 0 {
        return 0
    }

    level := hero.GetHeroExperienceLevel()
    return base + hero.getBaseRangedAttackPowerProgression(level)
}

func (hero *Hero) getBaseRangedAttackPowerProgression(level units.HeroExperienceLevel) int {
    switch level {
        case units.ExperienceHero: return 0
        case units.ExperienceMyrmidon: return 1
        case units.ExperienceCaptain: return 2
        case units.ExperienceCommander: return 3
        case units.ExperienceChampionHero: return 4
        case units.ExperienceLord: return 5
        case units.ExperienceGrandLord: return 6
        case units.ExperienceSuperHero: return 7
        case units.ExperienceDemiGod: return 8
    }
    return 0
}

func (hero *Hero) GetRangedAttackPower() int {
    base := hero.GetBaseRangedAttackPower()

    if base == 0 {
        return 0
    }

    for _, item := range hero.Equipment {
        if item != nil {
            base += item.RangedAttackBonus()
        }
    }

    bonus := 0

    if hero.Unit.GetRangedAttackDamageType() == units.DamageRangedMagical {
        bonus += hero.GetAbilityMagicRangedAttack()
    } else {
        bonus += hero.GetAbilityRangedAttack()
    }

    for _, enchantment := range hero.GetEnchantments() {
        base += hero.RangedEnchantmentBonus(enchantment)
    }

    return base + bonus
}

func (hero *Hero) GetBaseDefense() int {
    level := hero.GetHeroExperienceLevel()
    return hero.Unit.GetBaseDefense() + hero.getBaseDefenseProgression(level)
}

func (hero *Hero) getBaseDefenseProgression(level units.HeroExperienceLevel) int {
    switch level {
        case units.ExperienceHero: return 0
        case units.ExperienceMyrmidon: return 1
        case units.ExperienceCaptain: return 1
        case units.ExperienceCommander: return 2
        case units.ExperienceChampionHero: return 2
        case units.ExperienceLord: return 3
        case units.ExperienceGrandLord: return 3
        case units.ExperienceSuperHero: return 4
        case units.ExperienceDemiGod: return 4
    }
    return 0
}

func (hero *Hero) GetDefense() int {
    base := hero.GetBaseDefense()

    for _, item := range hero.Equipment {
        if item != nil {
            base += item.DefenseBonus()
        }
    }

    for _, enchantment := range hero.GetEnchantments() {
        base += hero.DefenseEnchantmentBonus(enchantment)
    }

    return base
}

func (hero *Hero) GetBaseResistance() int {
    level := hero.GetHeroExperienceLevel()

    modifier := 0
    if hero.HasAbility(data.AbilityLucky) {
        modifier += 1
    }

    return hero.Unit.GetBaseResistance() + hero.getBaseResistanceProgression(level) + modifier
}

func (hero *Hero) getBaseResistanceProgression(level units.HeroExperienceLevel) int {
    switch level {
        case units.ExperienceHero: return 0
        case units.ExperienceMyrmidon: return 1
        case units.ExperienceCaptain: return 2
        case units.ExperienceCommander: return 3
        case units.ExperienceChampionHero: return 4
        case units.ExperienceLord: return 5
        case units.ExperienceGrandLord: return 6
        case units.ExperienceSuperHero: return 7
        case units.ExperienceDemiGod: return 8
    }
    return 0
}

// any added resistance from abilities (agility)
func (hero *Hero) GetAbilityDefense() int {
    level := hero.GetHeroExperienceLevel()
    extra := 0
    if hero.HasAbility(data.AbilityAgility) {
        extra = level.ToInt() + 1
    } else if hero.HasAbility(data.AbilitySuperAgility) {
        extra = int(float64(level.ToInt() + 1) * 1.5)
    }

    return extra
}

func (hero *Hero) GetAbilityToHit() int {
    level := hero.GetHeroExperienceLevel()

    if hero.HasAbility(data.AbilityBlademaster) {
        switch level {
            case units.ExperienceHero: return 0
            case units.ExperienceMyrmidon: return 10
            case units.ExperienceCaptain: return 10
            case units.ExperienceCommander: return 20
            case units.ExperienceChampionHero: return 20
            case units.ExperienceLord: return 30
            case units.ExperienceGrandLord: return 30
            case units.ExperienceSuperHero: return 40
            case units.ExperienceDemiGod: return 40
        }
    } else if hero.HasAbility(data.AbilitySuperBlademaster) {
        switch level {
            case units.ExperienceHero: return 0
            case units.ExperienceMyrmidon: return 10
            case units.ExperienceCaptain: return 20
            case units.ExperienceCommander: return 30
            case units.ExperienceChampionHero: return 30
            case units.ExperienceLord: return 40
            case units.ExperienceGrandLord: return 50
            case units.ExperienceSuperHero: return 60
            case units.ExperienceDemiGod: return 80
        }
    }

    return 0
}

func (hero *Hero) GetAbilityMagicRangedAttack() int {
    level := hero.GetHeroExperienceLevel()
    extra := 0

    if hero.HasAbility(data.AbilityArcanePower) {
        extra = level.ToInt() + 1
    } else if hero.HasAbility(data.AbilitySuperArcanePower) {
        extra = int(float64(level.ToInt() + 1) * 1.5)
    }

    return extra
}

// extra experience points to apply to all normal units in the same stack as the hero
func (hero *Hero) GetAbilityExperienceBonus() int {
    level := hero.GetHeroExperienceLevel()
    extra := 0

    if hero.HasAbility(data.AbilityArmsmaster) {
        extra = (level.ToInt() + 1) * 2
    } else if hero.HasAbility(data.AbilitySuperArmsmaster) {
        extra = int(float64((level.ToInt() + 1) * 2) * 1.5)
    }

    return extra
}

func (hero *Hero) GetAbilityHealth() int {
    level := hero.GetHeroExperienceLevel()
    extra := 0

    if hero.HasAbility(data.AbilityConstitution) {
        extra = level.ToInt() + 1
    } else if hero.HasAbility(data.AbilitySuperConstitution) {
        extra = int(float64(level.ToInt() + 1) * 1.5)
    }

    return extra
}

func (hero *Hero) GetAbilityLeadership() int {
    level := hero.GetHeroExperienceLevel()

    if hero.HasAbility(data.AbilityLeadership) {
        switch level {
            case units.ExperienceHero: return 0
            case units.ExperienceMyrmidon: return 0
            case units.ExperienceCaptain: return 1
            case units.ExperienceCommander: return 1
            case units.ExperienceChampionHero: return 1
            case units.ExperienceLord: return 2
            case units.ExperienceGrandLord: return 2
            case units.ExperienceSuperHero: return 2
            case units.ExperienceDemiGod: return 3
        }
    } else if hero.HasAbility(data.AbilitySuperLeadership) {
        switch level {
            case units.ExperienceHero: return 0
            case units.ExperienceMyrmidon: return 1
            case units.ExperienceCaptain: return 1
            case units.ExperienceCommander: return 2
            case units.ExperienceChampionHero: return 2
            case units.ExperienceLord: return 3
            case units.ExperienceGrandLord: return 3
            case units.ExperienceSuperHero: return 4
            case units.ExperienceDemiGod: return 4
        }
    }

    return 0
}

// added fame to the wizard
func (hero *Hero) GetAbilityFame() int {
    level := hero.GetHeroExperienceLevel()
    extra := 0

    if hero.HasAbility(data.AbilityLegendary) {
        extra = (level.ToInt() + 1) * 3
    } else if hero.HasAbility(data.AbilitySuperLegendary) {
        extra = int(float64((level.ToInt() + 1) * 3) * 1.5)
    }

    return extra
}

func (hero *Hero) GetAbilityMelee() int {
    level := hero.GetHeroExperienceLevel()
    extra := 0

    if hero.HasAbility(data.AbilityMight) {
        extra = level.ToInt() + 1
    } else if hero.HasAbility(data.AbilitySuperMight) {
        extra = int(float64(level.ToInt() + 1) * 1.5)
    }

    return extra + hero.GetAbilityLeadership()
}

func (hero *Hero) GetAbilityRangedAttack() int {
    return hero.GetAbilityMelee() / 2
}

func (hero *Hero) GetAbilityResistance() int {
    level := hero.GetHeroExperienceLevel()
    extra := 0

    if hero.HasAbility(data.AbilityPrayermaster) {
        extra = level.ToInt() + 1
    } else if hero.HasAbility(data.AbilitySuperPrayermaster) {
        extra = int(float64(level.ToInt() + 1) * 1.5)
    }

    return extra
}

// extra research points to apply at each turn
func (hero *Hero) GetAbilityResearch() int {
    level := hero.GetHeroExperienceLevel()
    extra := 0

    if hero.HasAbility(data.AbilitySage) {
        extra = (level.ToInt() + 1) * 3
    } else if hero.HasAbility(data.AbilitySuperSage) {
        extra = int(float64((level.ToInt() + 1) * 3) * 1.5)
    }

    return extra
}

func (hero *Hero) GetResistance() int {
    base := hero.GetBaseResistance()

    for _, item := range hero.Equipment {
        if item != nil {
            base += item.ResistanceBonus()
        }
    }

    for _, enchantment := range hero.GetEnchantments() {
        base += hero.ResistanceEnchantmentBonus(enchantment)
    }

    return base + hero.GetAbilityResistance()
}

func (hero *Hero) GetFullHitPoints() int {
    base := hero.GetBaseHitPoints()

    for _, enchantment := range hero.GetEnchantments() {
        base += hero.HitPointsEnchantmentBonus(enchantment)
    }

    if hero.OverworldUnit.GlobalEnchantments.HasFriendlyEnchantment(data.EnchantmentCharmOfLife) {
        base = int(math.Ceil(float64(base) * 1.25))
    }

    return base + hero.GetAbilityHealth()
}

func (hero *Hero) GetHitPoints() int {
    return (hero.GetMaxHealth() - hero.GetDamage()) / hero.GetCount()
}

func (hero *Hero) GetBaseHitPoints() int {
    level := hero.GetHeroExperienceLevel()
    return hero.OverworldUnit.GetBaseHitPoints() + hero.getBaseHitPointsProgression(level)
}

func (hero *Hero) getBaseHitPointsProgression(level units.HeroExperienceLevel) int {
    switch level {
        case units.ExperienceHero: return 0
        case units.ExperienceMyrmidon: return 1
        case units.ExperienceCaptain: return 2
        case units.ExperienceCommander: return 3
        case units.ExperienceChampionHero: return 4
        case units.ExperienceLord: return 5
        case units.ExperienceGrandLord: return 6
        case units.ExperienceSuperHero: return 7
        case units.ExperienceDemiGod: return 8
    }
    return 0
}

func (hero *Hero) GetBaseProgression() []string {
    var improvements []string

    level := hero.GetHeroExperienceLevel()
    if level <= units.ExperienceHero {
        return improvements
    }

    hitPoints := hero.getBaseHitPointsProgression(level) - hero.getBaseHitPointsProgression(level - 1)
    if hitPoints > 0 {
        improvements = append(improvements, fmt.Sprintf("+%v Hit Points", hitPoints))
    }

    resistance := hero.getBaseResistanceProgression(level) - hero.getBaseResistanceProgression(level - 1)
    if resistance > 0 {
        improvements = append(improvements, fmt.Sprintf("+%v Resistance", resistance))
    }

    meleeAttack := hero.getBaseMeleeAttackPowerProgression(level) - hero.getBaseMeleeAttackPowerProgression(level - 1)
    rangedAttack := hero.getBaseRangedAttackPowerProgression(level) - hero.getBaseRangedAttackPowerProgression(level - 1)
    if meleeAttack > 0 || rangedAttack > 0{
        improvements = append(improvements, fmt.Sprintf("+%v Attack", min(meleeAttack, rangedAttack)))
    }

    defense := hero.getBaseDefenseProgression(level) - hero.getBaseDefenseProgression(level - 1)
    if defense > 0 {
        improvements = append(improvements, fmt.Sprintf("+%v Defense", defense))
    }

    return improvements
}

// these abilities include the bonuses from hero level progression
func (hero *Hero) GetAbilities() []data.Ability {
    var enchantmentAbilities []data.Ability
    for _, enchantment := range hero.GetEnchantments() {
        enchantmentAbilities = append(enchantmentAbilities, enchantment.Abilities()...)
    }

    var outAbilities []data.Ability
    for _, ability := range hero.Abilities {
        newAbility := ability
        switch newAbility.Ability {
            case data.AbilityThrown:
                newAbility.Value = float32(hero.GetThrownValue())
            case data.AbilityCaster:
                newAbility.Value = hero.GetCasterValue()
            case data.AbilityFireBreath:
                newAbility.Value = float32(hero.GetFireBreathValue())
            default:
                newAbility.Value = float32(hero.GetAbilityBonus(ability.Ability))
        }
        outAbilities = append(outAbilities, newAbility)
    }

    return slices.SortedFunc(slices.Values(append(outAbilities, enchantmentAbilities...)), func (a, b data.Ability) int {
        return cmp.Compare(a.Ability, b.Ability)
    })
}

func (hero *Hero) GetTitle() string {
    switch hero.HeroType {
        case HeroTorin: return "Chosen"
        case HeroFang: return "Draconian"
        case HeroBShan: return "Dervish"
        case HeroMorgana: return "Witch"
        case HeroWarrax: return "Chaos Warrior"
        case HeroMysticX: return "Unknown"
        case HeroBahgtru: return "Orc Warrior"
        case HeroDethStryke: return "Swordsman"
        case HeroSpyder: return "Rogue"
        case HeroSirHarold: return "Knight"
        case HeroBrax: return "Dwarf"
        case HeroRavashack: return "Necromancer"
        case HeroGreyfairer: return "Druid"
        case HeroShalla: return "Amazon"
        case HeroRoland: return "Paladin"
        case HeroMalleus: return "Magician"
        case HeroMortu: return "Black Knight"
        case HeroGunther: return "Barbarian"
        case HeroRakir: return "Beastmaster"
        case HeroJaer: return "Wind Mage"
        case HeroTaki: return "War Monk"
        case HeroYramrag: return "Warlock"
        case HeroValana: return "Bard"
        case HeroElana: return "Priestess"
        case HeroAerie: return "Illusionist"
        case HeroMarcus: return "Ranger"
        case HeroReywind: return "Warrior Mage"
        case HeroAlorra: return "Elven Archer"
        case HeroZaldron: return "Sage"
        case HeroShinBo: return "Ninja"
        case HeroSerena: return "Healer"
        case HeroShuri: return "Huntress"
        case HeroTheria: return "Thief"
        case HeroTumu: return "Assassin"
        case HeroAureus: return "Golden One"
    }

    return ""
}

func (hero *Hero) GetArtifacts() []*artifact.Artifact {
    return hero.Equipment[:]
}

func (hero *Hero) GetArtifactSlots() []artifact.ArtifactSlot {
    if hero.OverworldUnit.Unit.Equals(units.HeroTorin) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroFang) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroBShan) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotRangedWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroMorgana) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMagicWeapon, artifact.ArtifactSlotJewelry, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroWarrax) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotAnyWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroMysticX) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotAnyWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroBahgtru) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroDethStryke) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroSpyder) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroSirHarold) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroBrax) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroRavashack) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMagicWeapon, artifact.ArtifactSlotJewelry, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroGreyfairer) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMagicWeapon, artifact.ArtifactSlotJewelry, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroShalla) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroRoland) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroMalleus) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMagicWeapon, artifact.ArtifactSlotJewelry, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroMortu) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroGunther) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroRakir) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroJaer) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMagicWeapon, artifact.ArtifactSlotJewelry, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroTaki) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroYramrag) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMagicWeapon, artifact.ArtifactSlotJewelry, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroValana) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroElana) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMagicWeapon, artifact.ArtifactSlotJewelry, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroAerie) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMagicWeapon, artifact.ArtifactSlotJewelry, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroMarcus) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotRangedWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroReywind) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotAnyWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroAlorra) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotRangedWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroZaldron) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMagicWeapon, artifact.ArtifactSlotJewelry, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroShinBo) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroSerena) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroShuri) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotRangedWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroTheria) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroTumu) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.OverworldUnit.Unit.Equals(units.HeroAureus) {
        return []artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    return []artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotArmor}
}
