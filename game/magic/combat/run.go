package combat

import (
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/unitview"
    "github.com/kazzmir/master-of-magic/lib/fraction"
)

type ProxySpellSystem struct {
}

func (system *ProxySpellSystem) CreateFireballProjectile(target *ArmyUnit, cost int) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateIceBoltProjectile(target *ArmyUnit, cost int) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateStarFiresProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreatePsionicBlastProjectile(target *ArmyUnit, cost int) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateDoomBoltProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateFireBoltProjectile(target *ArmyUnit, cost int) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateLightningBoltProjectile(target *ArmyUnit, cost int) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateWarpLightningProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateFlameStrikeProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateLifeDrainProjectile(target *ArmyUnit, reduceResistance int, player ArmyPlayer, unitCaster *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateDispelEvilProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateHealingProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateHolyWordProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateRecallHeroProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateCracksCallProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateWebProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateBanishProjectile(target *ArmyUnit, reduceResistance int) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateDispelMagicProjectile(target *ArmyUnit, caster ArmyPlayer, dispelStrength int) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateWordOfRecallProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateDisintegrateProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateDisruptProjectile(x int, y int) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateMagicVortex(x int, y int) *OtherUnit {
    return nil
}

func (system *ProxySpellSystem) CreateWarpWoodProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateDeathSpellProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateWordOfDeathProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateSummoningCircle(x int, y int) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateMindStormProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateBlessProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateWeaknessProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateBlackSleepProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateVertigoProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateShatterProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateWarpCreatureProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateConfusionProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreatePossessionProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateCreatureBindingProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreatePetrifyProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateChaosChannelsProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateHeroismProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateHolyArmorProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateHolyWeaponProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateInvulnerabilityProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateLionHeartProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateRighteousnessProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateTrueSightProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateElementalArmorProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateGiantStrengthProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateIronSkinProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateStoneSkinProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateRegenerationProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateResistElementsProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateFlightProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateGuardianWindProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateHasteProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateInvisibilityProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateMagicImmunityProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateResistMagicProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateSpellLockProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateEldritchWeaponProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateFlameBladeProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateImmolationProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateBerserkProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateCloakOfFearProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) CreateWraithFormProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *ProxySpellSystem) PlaySound(spell spellbook.Spell) {
}

type FakeDamageIndicators struct {
}

func (indicators *FakeDamageIndicators) AddDamageIndicator(unit *ArmyUnit, damage int) {
    // nothing
}

type ProxyActions struct {
    Model *CombatModel
}

func (actions *ProxyActions) RangeAttack(attacker *ArmyUnit, defender *ArmyUnit) {
    effect := actions.Model.CreateRangeAttackEffect(attacker, &FakeDamageIndicators{})

    for range unitview.CombatPoints(attacker.Figures()) {
        effect(defender)
    }
}

func (actions *ProxyActions) MeleeAttack(attacker *ArmyUnit, defender *ArmyUnit) {
    actions.Model.meleeAttack(attacker, defender)
}

func (actions *ProxyActions) MoveUnit(mover *ArmyUnit, path pathfinding.Path) {
    for len(path) > 0 && mover.MovesLeft.GreaterThan(fraction.FromInt(0)) {
        targetX, targetY := path[0].X, path[0].Y
        died := actions.Model.MoveUnit(mover, targetX, targetY)
        if died {
            return
        }

        mover.MoveX = float64(targetX)
        mover.MoveY = float64(targetY)

        path = path[1:]
    }

    // not really necessary as current path is only used when drawing the UI
    mover.CurrentPath = nil
}

func (actions *ProxyActions) Teleport(unit *ArmyUnit, x, y int, merge bool) {
    actions.Model.Teleport(unit, x, y)
}

func (actiosn *ProxyActions) DoProjectiles() {
}

func (actions *ProxyActions) ExtraControl() bool {
    return false
}

func (actions *ProxyActions) SingleAuto() bool {
    return false
}

// Run combat without a UI, and no user input. This is useful for simulating combat
// scenarios, running automated tests, or benchmarking performance.
func Run(model *CombatModel) CombatState {

    actions := &ProxyActions{
        Model: model,
    }

    state := CombatStateRunning
    for state == CombatStateRunning {
        model.Update(&ProxySpellSystem{}, actions, false, 0, 0)
        state = model.FinalState()
    }

    return state
}
