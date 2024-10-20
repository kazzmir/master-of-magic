package hero

import (
    "fmt"
    "slices"
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
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
    Unit *units.OverworldUnit
    HeroType HeroType
    Name string
    Status HeroStatus

    // set at start of game
    Abilities []units.Ability

    Equipment [3]*artifact.Artifact
}

type noInfo struct {
}

func (noInfo *noInfo) HasWarlord() bool {
    return false
}

func (noInfo *noInfo) Crusade() bool {
    return false
}

func MakeHeroSimple(heroType HeroType) *Hero {
    unit := units.MakeOverworldUnit(heroType.GetUnit())
    unit.ExperienceInfo = &noInfo{}
    return MakeHero(unit, heroType, heroType.DefaultName())
}

func MakeHero(unit *units.OverworldUnit, heroType HeroType, name string) *Hero {
    return &Hero{
        Unit: unit,
        Name: name,
        HeroType: heroType,
        Abilities: slices.Clone(unit.GetAbilities()),
        Status: StatusAvailable,
    }
}

type abilityChoice int
const (
    abilityChoiceFighter abilityChoice = iota
    abilityChoiceMage
    abilityChoiceAny
)

func selectAbility(kind abilityChoice) units.Ability {
    anyChoices := []units.Ability{
        units.AbilityCharmed,
        units.AbilityLucky,
        units.AbilityNoble,
        units.AbilitySage,
    }

    fighterChoices := []units.Ability{
        units.AbilityAgility,
        units.AbilityArmsmaster,
        units.AbilityBlademaster,
        units.AbilityConstitution,
        units.AbilityLeadership,
        units.AbilityLegendary,
        units.AbilityMight,
    }

    mageChoices := []units.Ability{
        units.AbilityArcanePower,
        units.AbilityCaster,
        units.AbilityPrayermaster,
        units.AbilitySage,
    }

    var use []units.Ability
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

func (hero *Hero) SetExtraAbilities() {
}

func (hero *Hero) SetStatus(status HeroStatus) {
    hero.Status = status
}

func (hero *Hero) GetName() string {
    return fmt.Sprintf("%v the %v", hero.Unit.GetName(), hero.Title())
}

func (hero *Hero) ShortName() string {
    return hero.Unit.GetName()
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

// fee is halved if the hiring wizard is charismatic, handle that elsewhere
func (hero *Hero) GetHireFee() int {
    base := 100 + hero.HeroType.GetRequiredFame() * 10

    levelInt := 1

    level := hero.GetExperienceLevel()
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
    hero.Unit.AdjustHealth(amount)

    if hero.GetHealth() <= 0 {
        hero.SetStatus(StatusDead)
    }
}

func (hero *Hero) GetCombatRangeIndex(facing units.Facing) int {
    return hero.Unit.GetCombatRangeIndex(facing)
}

func (hero *Hero) GetHealth() int {
    return hero.Unit.GetHealth()
}

func (hero *Hero) GetMaxHealth() int {
    return hero.GetHitPoints()
}

func (hero *Hero) AddExperience(amount int) {
    hero.Unit.AddExperience(amount)
}

func (hero *Hero) GetExperience() int {
    return hero.Unit.GetExperience()
}

func (hero *Hero) GetToHitMelee() int {
    base := 30

    level := hero.GetExperienceLevel()
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

    return base
}

func (hero *Hero) GetLbxFile() string {
    return hero.Unit.GetLbxFile()
}

func (hero *Hero) GetLbxIndex() int {
    return hero.Unit.GetLbxIndex()
}

func (hero *Hero) GetPatrol() bool {
    return hero.Unit.GetPatrol()
}

func (hero *Hero) SetPatrol(patrol bool) {
    hero.Unit.SetPatrol(patrol)
}

func (hero *Hero) GetPlane() data.Plane {
    return hero.Unit.GetPlane()
}

func (hero *Hero) GetRace() data.Race {
    return hero.Unit.GetRace()
}

func (hero *Hero) GetRawUnit() units.Unit {
    return hero.Unit.GetRawUnit()
}

func (hero *Hero) GetX() int {
    return hero.Unit.GetX()
}

func (hero *Hero) GetY() int {
    return hero.Unit.GetY()
}

func (hero *Hero) Move(dx int, dy int, cost fraction.Fraction){
    hero.Unit.Move(dx, dy, cost)
}

func (hero *Hero) NaturalHeal(rate float64) {
    amount := float64(hero.GetMaxHealth()) * rate
    if amount < 1 {
        amount = 1
    }
    hero.AdjustHealth(int(amount))
}

func (hero *Hero) ResetMoves() {
    hero.Unit.ResetMoves()
}

func (hero *Hero) SetId(id uint64) {
    hero.Unit.SetId(id)
}

func (hero *Hero) GetMovesLeft() fraction.Fraction {
    return hero.Unit.GetMovesLeft()
}

func (hero *Hero) SetMovesLeft(moves fraction.Fraction) {
    hero.Unit.SetMovesLeft(moves)
}

func (hero *Hero) GetAttackSound() units.AttackSound {
    return hero.Unit.GetAttackSound()
}

func (hero *Hero) GetMovementSound() units.MovementSound {
    return hero.Unit.GetMovementSound()
}

func (hero *Hero) GetRangeAttackSound() units.RangeAttackSound {
    return hero.Unit.GetRangeAttackSound()
}

func (hero *Hero) GetRangedAttackDamageType() units.Damage {
    return hero.Unit.GetRangedAttackDamageType()
}

func (hero *Hero) GetRangedAttacks() int {
    return hero.Unit.GetRangedAttacks()
}

func (hero *Hero) HasAbility(ability units.Ability) bool {
    return hero.Unit.HasAbility(ability)
}

func (hero *Hero) IsFlying() bool {
    return hero.Unit.IsFlying()
}

func (hero *Hero) GetBanner() data.BannerType {
    return hero.Unit.GetBanner()
}

func (hero *Hero) GetCombatLbxFile() string {
    return hero.Unit.GetCombatLbxFile()
}

func (hero *Hero) GetCombatIndex(facing units.Facing) int {
    return hero.Unit.GetCombatIndex(facing)
}

func (hero *Hero) GetCount() int {
    return 1
}

func (hero *Hero) GetUpkeepGold() int {
    return hero.Unit.GetUpkeepGold()
}

func (hero *Hero) GetUpkeepFood() int {
    return hero.Unit.GetUpkeepFood()
}

func (hero *Hero) GetUpkeepMana() int {
    return hero.Unit.GetUpkeepMana()
}

func (hero *Hero) GetMovementSpeed() int {
    return hero.Unit.GetMovementSpeed()
}

func (hero *Hero) GetProductionCost() int {
    return hero.Unit.GetProductionCost()
}

func (hero *Hero) GetExperienceLevel() units.HeroExperienceLevel {
    if hero.Unit.ExperienceInfo != nil {
        return units.GetHeroExperienceLevel(hero.Unit.Experience, hero.Unit.ExperienceInfo.HasWarlord(), hero.Unit.ExperienceInfo.Crusade())
    }

    return units.ExperienceHero
}

func (hero *Hero) SetExperienceInfo(info units.ExperienceInfo) {
    hero.Unit.ExperienceInfo = info
}

func (hero *Hero) GetBaseMeleeAttackPower() int {
    return hero.Unit.GetBaseMeleeAttackPower()
}

func (hero *Hero) GetMeleeAttackPower() int {
    base := hero.GetBaseMeleeAttackPower()

    for _, item := range hero.Equipment {
        if item != nil {
            base += item.MeleeBonus()
        }
    }

    level := hero.GetExperienceLevel()
    switch level {
        case units.ExperienceHero:
        case units.ExperienceMyrmidon: base += 1
        case units.ExperienceCaptain: base += 2
        case units.ExperienceCommander: base += 3
        case units.ExperienceChampionHero: base += 4
        case units.ExperienceLord: base += 5
        case units.ExperienceGrandLord: base += 6
        case units.ExperienceSuperHero: base += 7
        case units.ExperienceDemiGod: base += 8
    }

    return base
}

func (hero *Hero) GetBaseRangedAttackPower() int {
    return hero.Unit.GetBaseRangedAttackPower()
}

func (hero *Hero) GetRangedAttackPower() int {
    base := hero.Unit.GetBaseRangedAttackPower()

    if base == 0 {
        return 0
    }

    for _, item := range hero.Equipment {
        if item != nil {
            base += item.RangedAttackBonus()
        }
    }

    level := hero.GetExperienceLevel()
    switch level {
        case units.ExperienceHero:
        case units.ExperienceMyrmidon: base += 1
        case units.ExperienceCaptain: base += 2
        case units.ExperienceCommander: base += 3
        case units.ExperienceChampionHero: base += 4
        case units.ExperienceLord: base += 5
        case units.ExperienceGrandLord: base += 6
        case units.ExperienceSuperHero: base += 7
        case units.ExperienceDemiGod: base += 8
    }

    return base
}

func (hero *Hero) GetBaseDefense() int {
    return hero.Unit.GetBaseDefense()
}

func (hero *Hero) GetDefense() int {
    base := hero.Unit.GetBaseDefense()

    for _, item := range hero.Equipment {
        if item != nil {
            base += item.DefenseBonus()
        }
    }

    level := hero.GetExperienceLevel()
    switch level {
        case units.ExperienceHero:
        case units.ExperienceMyrmidon: base += 1
        case units.ExperienceCaptain: base += 1
        case units.ExperienceCommander: base += 2
        case units.ExperienceChampionHero: base += 2
        case units.ExperienceLord: base += 3
        case units.ExperienceGrandLord: base += 3
        case units.ExperienceSuperHero: base += 4
        case units.ExperienceDemiGod: base += 4
    }

    return base
}

func (hero *Hero) GetBaseResistance() int {
    return hero.Unit.GetBaseResistance()
}

func (hero *Hero) GetResistance() int {
    base := hero.Unit.GetBaseResistance()

    for _, item := range hero.Equipment {
        if item != nil {
            base += item.ResistanceBonus()
        }
    }

    level := hero.GetExperienceLevel()
    switch level {
        case units.ExperienceHero:
        case units.ExperienceMyrmidon: base += 1
        case units.ExperienceCaptain: base += 2
        case units.ExperienceCommander: base += 3
        case units.ExperienceChampionHero: base += 4
        case units.ExperienceLord: base += 5
        case units.ExperienceGrandLord: base += 6
        case units.ExperienceSuperHero: base += 7
        case units.ExperienceDemiGod: base += 8
    }

    return base
}

func (hero *Hero) GetHitPoints() int {
    base := hero.Unit.GetBaseHitPoints()

    level := hero.GetExperienceLevel()
    switch level {
        case units.ExperienceHero:
        case units.ExperienceMyrmidon: base += 1
        case units.ExperienceCaptain: base += 2
        case units.ExperienceCommander: base += 3
        case units.ExperienceChampionHero: base += 4
        case units.ExperienceLord: base += 5
        case units.ExperienceGrandLord: base += 6
        case units.ExperienceSuperHero: base += 7
        case units.ExperienceDemiGod: base += 8
    }

    return base
}

func (hero *Hero) GetBaseHitPoints() int {
    return hero.Unit.GetBaseHitPoints()
}

func (hero *Hero) GetAbilities() []units.Ability {
    return hero.Abilities
}

func (hero *Hero) Title() string {
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

func (hero *Hero) Slots() [3]artifact.ArtifactSlot {
    if hero.Unit.Unit.Equals(units.HeroTorin) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroFang) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroBShan) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotRangedWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroMorgana) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMagicWeapon, artifact.ArtifactSlotJewelry, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroWarrax) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotAnyWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroMysticX) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotAnyWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroBahgtru) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroDethStryke) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroSpyder) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroSirHarold) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroBrax) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroRavashack) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMagicWeapon, artifact.ArtifactSlotJewelry, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroGreyfairer) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMagicWeapon, artifact.ArtifactSlotJewelry, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroShalla) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroRoland) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroMalleus) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMagicWeapon, artifact.ArtifactSlotJewelry, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroMortu) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroGunther) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroRakir) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroJaer) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMagicWeapon, artifact.ArtifactSlotJewelry, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroTaki) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroYramrag) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMagicWeapon, artifact.ArtifactSlotJewelry, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroValana) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroElana) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMagicWeapon, artifact.ArtifactSlotJewelry, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroAerie) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMagicWeapon, artifact.ArtifactSlotJewelry, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroMarcus) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotRangedWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroReywind) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotAnyWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroAlorra) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotRangedWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroZaldron) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMagicWeapon, artifact.ArtifactSlotJewelry, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroShinBo) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroSerena) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroShuri) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotRangedWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroTheria) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroTumu) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    if hero.Unit.Unit.Equals(units.HeroAureus) {
        return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotJewelry}
    }

    return [3]artifact.ArtifactSlot{artifact.ArtifactSlotMeleeWeapon, artifact.ArtifactSlotArmor, artifact.ArtifactSlotArmor}
}
