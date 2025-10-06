package combat

import (
    "slices"
)

type CombatObserver interface {
    ThrowAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int)
    PoisonTouchAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int)
    LifeStealTouchAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int)
    StoningTouchAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int)
    DispelEvilTouchAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int)
    DeathTouchAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int)
    DestructionAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int)
    StoneGazeAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int)
    DeathGazeAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int)
    DoomGazeAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int)
    FireBreathAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int)
    LightningBreathAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int)
    ImmolationAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int)
    MeleeAttack(attacker *ArmyUnit, defender *ArmyUnit, damageRoll []int)
    CauseFear(attacker *ArmyUnit, defender *ArmyUnit, fear int)
    WallOfFire(defender *ArmyUnit, damage int)
    UnitKilled(unit *ArmyUnit)
}

type CombatObservers struct {
    Observers []CombatObserver
}

func (observer *CombatObservers) AddObserver(add CombatObserver) {
    observer.Observers = append(observer.Observers, add)
}

func (observer *CombatObservers) RemoveObserver(remove CombatObserver) {
    observer.Observers = slices.DeleteFunc(observer.Observers, func (check CombatObserver) bool {
        return check == remove
    })
}

func (observer *CombatObservers) UnitKilled(unit *ArmyUnit) {
    for _, notify := range observer.Observers {
        notify.UnitKilled(unit)
    }
}

func (observer *CombatObservers) WallOfFire(defender *ArmyUnit, damage int) {
    for _, notify := range observer.Observers {
        notify.WallOfFire(defender, damage)
    }
}

func (observer *CombatObservers) ThrowAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int) {
    for _, notify := range observer.Observers {
        notify.ThrowAttack(attacker, defender, damage)
    }
}

func (observer *CombatObservers) MeleeAttack(attacker *ArmyUnit, defender *ArmyUnit, damageRoll []int) {
    for _, notify := range observer.Observers {
        notify.MeleeAttack(attacker, defender, damageRoll)
    }
}

func (observer *CombatObservers) CauseFear(attacker *ArmyUnit, defender *ArmyUnit, fear int) {
    for _, notify := range observer.Observers {
        notify.CauseFear(attacker, defender, fear)
    }
}

func (observer *CombatObservers) ImmolationAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int) {
    for _, notify := range observer.Observers {
        notify.ImmolationAttack(attacker, defender, damage)
    }
}

func (observer *CombatObservers) FireBreathAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int) {
    for _, notify := range observer.Observers {
        notify.FireBreathAttack(attacker, defender, damage)
    }
}

func (observer *CombatObservers) LightningBreathAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int) {
    for _, notify := range observer.Observers {
        notify.LightningBreathAttack(attacker, defender, damage)
    }
}

func (observer *CombatObservers) StoneGazeAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int) {
    for _, notify := range observer.Observers {
        notify.StoneGazeAttack(attacker, defender, damage)
    }
}

func (observer *CombatObservers) DeathGazeAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int) {
    for _, notify := range observer.Observers {
        notify.DeathGazeAttack(attacker, defender, damage)
    }
}

func (observer *CombatObservers) DoomGazeAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int) {
    for _, notify := range observer.Observers {
        notify.DoomGazeAttack(attacker, defender, damage)
    }
}

func (observer *CombatObservers) PoisonTouchAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int) {
    for _, notify := range observer.Observers {
        notify.PoisonTouchAttack(attacker, defender, damage)
    }
}

func (observer *CombatObservers) LifeStealTouchAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int) {
    for _, notify := range observer.Observers {
        notify.LifeStealTouchAttack(attacker, defender, damage)
    }
}

func (observer *CombatObservers) StoningTouchAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int) {
    for _, notify := range observer.Observers {
        notify.StoningTouchAttack(attacker, defender, damage)
    }
}

func (observer *CombatObservers) DispelEvilTouchAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int) {
    for _, notify := range observer.Observers {
        notify.DispelEvilTouchAttack(attacker, defender, damage)
    }
}

func (observer *CombatObservers) DeathTouchAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int) {
    for _, notify := range observer.Observers {
        notify.DeathTouchAttack(attacker, defender, damage)
    }
}

func (observer *CombatObservers) DestructionAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int) {
    for _, notify := range observer.Observers {
        notify.DestructionAttack(attacker, defender, damage)
    }
}
