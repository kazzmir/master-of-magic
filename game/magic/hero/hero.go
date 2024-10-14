package hero

import (
    "fmt"

    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
)

type Hero struct {
    Unit *units.OverworldUnit
    Title string

    Equipment [3]*artifact.Artifact
}

func (hero *Hero) GetName() string {
    return fmt.Sprintf("%v the %v", hero.Unit.GetName(), hero.Title)
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

func (hero *Hero) GetResistance() int {
    return hero.Unit.GetResistance()
}

func (hero *Hero) GetHitPoints() int {
    return hero.Unit.GetHitPoints()
}

func (hero *Hero) GetAbilities() []units.Ability {
    return hero.Unit.GetAbilities()
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

func (hero *Hero) PortraitIndex() int {
    if hero.Unit.Unit.Equals(units.HeroTorin) {
        return 0
    }

    if hero.Unit.Unit.Equals(units.HeroFang) {
        return 1
    }

    if hero.Unit.Unit.Equals(units.HeroBShan) {
        return 2
    }

    if hero.Unit.Unit.Equals(units.HeroMorgana) {
        return 3
    }

    if hero.Unit.Unit.Equals(units.HeroWarrax) {
        return 4
    }

    if hero.Unit.Unit.Equals(units.HeroMysticX) {
        return 5
    }

    if hero.Unit.Unit.Equals(units.HeroBahgtru) {
        return 6
    }

    if hero.Unit.Unit.Equals(units.HeroDethStryke) {
        return 7
    }

    if hero.Unit.Unit.Equals(units.HeroSpyder) {
        return 8
    }

    if hero.Unit.Unit.Equals(units.HeroSirHarold) {
        return 9
    }

    if hero.Unit.Unit.Equals(units.HeroBrax) {
        return 10
    }

    if hero.Unit.Unit.Equals(units.HeroRavashack) {
        return 11
    }

    if hero.Unit.Unit.Equals(units.HeroGreyfairer) {
        return 12
    }

    if hero.Unit.Unit.Equals(units.HeroShalla) {
        return 13
    }

    if hero.Unit.Unit.Equals(units.HeroRoland) {
        return 14
    }

    if hero.Unit.Unit.Equals(units.HeroMalleus) {
        return 15
    }

    if hero.Unit.Unit.Equals(units.HeroMortu) {
        return 16
    }

    if hero.Unit.Unit.Equals(units.HeroGunther) {
        return 17
    }

    if hero.Unit.Unit.Equals(units.HeroRakir) {
        return 18
    }

    if hero.Unit.Unit.Equals(units.HeroJaer) {
        return 19
    }

    if hero.Unit.Unit.Equals(units.HeroTaki) {
        return 20
    }

    if hero.Unit.Unit.Equals(units.HeroYramrag) {
        return 21
    }

    if hero.Unit.Unit.Equals(units.HeroValana) {
        return 22
    }

    if hero.Unit.Unit.Equals(units.HeroElana) {
        return 23
    }

    if hero.Unit.Unit.Equals(units.HeroAerie) {
        return 24
    }

    if hero.Unit.Unit.Equals(units.HeroMarcus) {
        return 25
    }

    if hero.Unit.Unit.Equals(units.HeroReywind) {
        return 26
    }

    if hero.Unit.Unit.Equals(units.HeroAlorra) {
        return 27
    }

    if hero.Unit.Unit.Equals(units.HeroZaldron) {
        return 28
    }

    if hero.Unit.Unit.Equals(units.HeroShinBo) {
        return 29
    }

    if hero.Unit.Unit.Equals(units.HeroSerena) {
        return 30
    }

    if hero.Unit.Unit.Equals(units.HeroShuri) {
        return 31
    }

    if hero.Unit.Unit.Equals(units.HeroTheria) {
        return 32
    }

    if hero.Unit.Unit.Equals(units.HeroTumu) {
        return 33
    }

    if hero.Unit.Unit.Equals(units.HeroAureus) {
        return 34
    }

    return -1
}
