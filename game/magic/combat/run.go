package combat

import (
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/unitview"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/log"
)

type ProxySpellSystem struct {
    Model *CombatModel
}

func (system *ProxySpellSystem) CreateFireballProjectile(target *ArmyUnit, cost int) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateFireballProjectileEffect(cost, &FakeDamageIndicators{}),
    }
}

func (system *ProxySpellSystem) CreateIceBoltProjectile(target *ArmyUnit, cost int) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateIceBoltProjectileEffect(cost, &FakeDamageIndicators{}),
    }
}

func (system *ProxySpellSystem) CreateStarFiresProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateStarFiresProjectileEffect(&FakeDamageIndicators{}),
    }
}

func (system *ProxySpellSystem) CreatePsionicBlastProjectile(target *ArmyUnit, cost int) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreatePsionicBlastProjectileEffect(cost, &FakeDamageIndicators{}),
    }
}

func (system *ProxySpellSystem) CreateDoomBoltProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateDoomBoltProjectileEffect(&FakeDamageIndicators{}),
    }
}

func (system *ProxySpellSystem) CreateFireBoltProjectile(target *ArmyUnit, cost int) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateFireBoltProjectileEffect(cost, &FakeDamageIndicators{}),
    }
}

func (system *ProxySpellSystem) CreateLightningBoltProjectile(target *ArmyUnit, cost int) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateLightningBoltProjectileEffect(cost, &FakeDamageIndicators{}),
    }
}

func (system *ProxySpellSystem) CreateWarpLightningProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateWarpLightningProjectileEffect(&FakeDamageIndicators{}),
    }
}

func (system *ProxySpellSystem) CreateFlameStrikeProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateFlameStrikeProjectileEffect(&FakeDamageIndicators{}),
    }
}

func (system *ProxySpellSystem) CreateLifeDrainProjectile(target *ArmyUnit, reduceResistance int, player ArmyPlayer, unitCaster *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateLifeDrainProjectileEffect(reduceResistance, player, unitCaster, &FakeDamageIndicators{}),
    }
}

func (system *ProxySpellSystem) CreateDispelEvilProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateDispelEvilProjectileEffect(&FakeDamageIndicators{}),
    }
}

func (system *ProxySpellSystem) CreateHealingProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateHealingProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateHolyWordProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateHolyWordProjectileEffect(&FakeDamageIndicators{}),
    }
}

func (system *ProxySpellSystem) CreateRecallHeroProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateRecallHeroProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateCracksCallProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateCracksCallProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateWebProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateWebProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateBanishProjectile(target *ArmyUnit, reduceResistance int) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateBanishProjectileEffect(reduceResistance, &FakeDamageIndicators{}),
    }
}

func (system *ProxySpellSystem) CreateDispelMagicProjectile(target *ArmyUnit, caster ArmyPlayer, dispelStrength int) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateDispelMagicProjectileEffect(caster, dispelStrength),
    }
}

func (system *ProxySpellSystem) CreateWordOfRecallProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateWordOfRecallProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateDisintegrateProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateDisintegrateProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateDisruptProjectile(x int, y int) *Projectile {
    fakeTarget := &ArmyUnit{
        X: x,
        Y: y,
    }

    return &Projectile{
        Target: fakeTarget,
        Effect: system.Model.CreateDisruptProjectileEffect(x, y),
    }
}

func (system *ProxySpellSystem) CreateMagicVortex(x int, y int) *MagicVortex {
    return &MagicVortex{
        X: x,
        Y: y,
    }
}

func (system *ProxySpellSystem) CreateWarpWoodProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateWarpWoodProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateDeathSpellProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateDeathSpellProjectileEffect(&FakeDamageIndicators{}),
    }
}

func (system *ProxySpellSystem) CreateWordOfDeathProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateWordOfDeathProjectileEffect(&FakeDamageIndicators{}),
    }
}

func (system *ProxySpellSystem) CreateSummoningCircle(x int, y int) *Projectile {
    // doesnt do anything
    return &Projectile{
        Effect: func(_ *ArmyUnit) {
        },
    }
}

func (system *ProxySpellSystem) CreateMindStormProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateMindStormProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateBlessProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateBlessProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateWeaknessProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateWeaknessProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateBlackSleepProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateBlackSleepProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateVertigoProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateVertigoProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateShatterProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateShatterProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateWarpCreatureProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateWarpCreatureProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateConfusionProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateConfusionProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreatePossessionProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreatePossessionProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateCreatureBindingProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateCreatureBindingProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreatePetrifyProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreatePetrifyProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateChaosChannelsProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateChaosChannelsProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateHeroismProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateHeroismProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateHolyArmorProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateHolyArmorProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateHolyWeaponProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateHolyWeaponProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateInvulnerabilityProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateInvulnerabilityProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateLionHeartProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateLionHeartProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateRighteousnessProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateRighteousnessProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateTrueSightProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateTrueSightProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateElementalArmorProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateElementalArmorProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateGiantStrengthProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateGiantStrengthProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateIronSkinProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateIronSkinProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateStoneSkinProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateStoneSkinProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateRegenerationProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateRegenerationProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateResistElementsProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateResistElementsProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateFlightProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateFlightProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateGuardianWindProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateGuardianWindProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateHasteProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateHasteProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateInvisibilityProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateInvisibilityProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateMagicImmunityProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateMagicImmunityProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateResistMagicProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateResistMagicProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateSpellLockProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateSpellLockProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateEldritchWeaponProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateEldritchWeaponProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateFlameBladeProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateFlameBladeProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateImmolationProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateImmolationProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateBerserkProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateBerserkProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateCloakOfFearProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateCloakOfFearProjectileEffect(),
    }
}

func (system *ProxySpellSystem) CreateWraithFormProjectile(target *ArmyUnit) *Projectile {
    return &Projectile{
        Target: target,
        Effect: system.Model.CreateWraithFormProjectileEffect(),
    }
}

func (system *ProxySpellSystem) PlaySound(spell spellbook.Spell) {
    // nothing
}

type FakeDamageIndicators struct {
}

func (indicators *FakeDamageIndicators) AddDamageIndicator(unit *ArmyUnit, damage int) {
    // nothing
}

type ProxyActions struct {
    Model *CombatModel
}

func (actions *ProxyActions) CreateRangeAttack(attacker *ArmyUnit, defender RangeTarget) {
    switch target := defender.(type) {
        case *ArmyUnit:
            effect := actions.Model.CreateRangeAttackEffect(attacker, &FakeDamageIndicators{})
            for range unitview.CombatPoints(attacker.Figures()) {
                effect(target)
            }

        case *WallTarget:
            effect := actions.Model.CreateRangeAttackWallEffect(attacker, target.X, target.Y)
            for range unitview.CombatPoints(attacker.Figures()) {
                effect(nil)
            }
    }
}

func (actions *ProxyActions) RangeAttack(attacker *ArmyUnit, defender RangeTarget) {
    actions.Model.rangeAttack(attacker, defender, actions)
}

func (actions *ProxyActions) MeleeAttack(attacker *ArmyUnit, defender *ArmyUnit) {
    actions.Model.meleeAttack(attacker, defender)
}

func (actions *ProxyActions) MeleeAttackWall(attacker *ArmyUnit, x int, y int) {
    actions.Model.meleeAttackWall(attacker, x, y)
}

func (actions *ProxyActions) MoveUnit(mover *ArmyUnit, path pathfinding.Path) {
    // log.Printf("Move %v along path: %+v", mover.Unit.GetName(), path)
    for len(path) > 0 && mover.MovesLeft.GreaterThan(fraction.FromInt(0)) {
        targetX, targetY := path[0].X, path[0].Y
        died := actions.Model.MoveUnit(mover, targetX, targetY)
        if died {
            return
        }

        path = path[1:]
    }

    // not really necessary as current path is only used when drawing the UI
    mover.CurrentPath = nil
}

func (actions *ProxyActions) Teleport(unit *ArmyUnit, x, y int, merge bool) {
    actions.Model.Teleport(unit, x, y)
}

func (actions *ProxyActions) DoProjectiles() {
    // just immediately apply all projectiles to their target
    for _, projectile := range actions.Model.Projectiles {
        projectile.Effect(projectile.Target)
    }

    actions.Model.Projectiles = nil
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

    spellSystem := &ProxySpellSystem{
        Model: model,
    }

    state := CombatStateRunning
    for state == CombatStateRunning {
        model.Update(spellSystem, actions, false, 0, 0)

        stop := false
        for !stop {
            select {
                case event := <-model.Events:
                    _ = event
                    log.Debug("Discarding event: %+v", event)
                default:
                    stop = true
            }
        }

        state = model.FinalState()
    }

    return state
}
