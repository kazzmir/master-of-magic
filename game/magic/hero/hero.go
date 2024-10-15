package hero

import (
    "fmt"

    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
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

type Hero struct {
    Unit *units.OverworldUnit
    HeroType HeroType
    Name string

    Equipment [3]*artifact.Artifact
}

func MakeHero(unit *units.OverworldUnit, heroType HeroType, name string) *Hero {
    return &Hero{
        Unit: unit,
        Name: name,
        HeroType: heroType,
    }
}

func (hero *Hero) GetName() string {
    return fmt.Sprintf("%v the %v", hero.Unit.GetName(), hero.Title())
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

func (hero *Hero) AdjustHealth(amount int) {
    hero.Unit.AdjustHealth(amount)
}

func (hero *Hero) GetCombatRangeIndex(facing units.Facing) int {
    return hero.Unit.GetCombatRangeIndex(facing)
}

func (hero *Hero) GetHealth() int {
    return hero.Unit.GetHealth()
}

func (hero *Hero) GetMaxHealth() int {
    return hero.Unit.GetMaxHealth()
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

func (hero *Hero) GetBaseMeleeAttackPower() int {
    return hero.Unit.GetMeleeAttackPower()
}

func (hero *Hero) GetMeleeAttackPower() int {
    base := hero.GetBaseMeleeAttackPower()

    for _, item := range hero.Equipment {
        if item != nil {
            base += item.MeleeBonus()
        }
    }

    return base
}

func (hero *Hero) GetBaseRangedAttackPower() int {
    return hero.Unit.GetRangedAttackPower()
}

func (hero *Hero) GetRangedAttackPower() int {
    base := hero.Unit.GetRangedAttackPower()

    for _, item := range hero.Equipment {
        if item != nil {
            base += item.RangedAttackBonus()
        }
    }

    return base
}

func (hero *Hero) GetBaseDefense() int {
    return hero.Unit.GetDefense()
}

func (hero *Hero) GetDefense() int {
    base := hero.Unit.GetDefense()

    for _, item := range hero.Equipment {
        if item != nil {
            base += item.DefenseBonus()
        }
    }

    return base
}

func (hero *Hero) GetBaseResistance() int {
    return hero.Unit.GetResistance()
}

func (hero *Hero) GetResistance() int {
    base := hero.Unit.GetResistance()

    for _, item := range hero.Equipment {
        if item != nil {
            base += item.ResistanceBonus()
        }
    }

    return base
}

func (hero *Hero) GetHitPoints() int {
    return hero.Unit.GetHitPoints()
}

func (hero *Hero) GetBaseHitPoints() int {
    return hero.Unit.GetHitPoints()
}

func (hero *Hero) GetAbilities() []units.Ability {
    return hero.Unit.GetAbilities()
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
