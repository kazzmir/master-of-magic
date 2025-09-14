package main

import (
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

func getUnitCost(unit *units.Unit) uint64 {
    // Settlers always cost 10
    if unit.IsSettlers() {
        return 10
    }

    if unit.Race == data.RaceHero {
        return 500000
    }

    if unit.Equals(units.LizardSpearmen) {
        return 100
    }

    if unit.Equals(units.LizardSwordsmen) {
        return 150
    }

    if unit.Equals(units.LizardHalberdiers) {
        return 200
    }

    if unit.Equals(units.LizardJavelineers) {
        return 200
    }

    if unit.Equals(units.LizardShamans) {
        return 300
    }

    if unit.Equals(units.DragonTurtle) {
        return 400
    }

    if unit.Equals(units.NomadSpearmen) {
        return 100
    }

    if unit.Equals(units.NomadSwordsmen) {
        return 150
    }

    if unit.Equals(units.NomadBowmen) {
        return 200
    }

    if unit.Equals(units.NomadPriest) {
        return 300
    }

    if unit.Equals(units.NomadHorsebowemen) {
        return 300
    }

    if unit.Equals(units.NomadPikemen) {
        return 250
    }

    if unit.Equals(units.NomadRangers) {
        return 300
    }

    if unit.Equals(units.Griffin) {
        return 450
    }

    if unit.Equals(units.OrcSpearmen) {
        return 100
    }

    if unit.Equals(units.OrcSwordsmen) {
        return 150
    }

    if unit.Equals(units.OrcHalberdiers) {
        return 200
    }

    if unit.Equals(units.OrcBowmen) {
        return 200
    }

    if unit.Equals(units.OrcCavalry) {
        return 300
    }

    if unit.Equals(units.OrcShamans) {
        return 300
    }

    if unit.Equals(units.OrcMagicians) {
        return 500
    }

    if unit.Equals(units.OrcEngineers) {
        return 50
    }

    if unit.Equals(units.WyvernRiders) {
        return 500
    }

    if unit.Equals(units.TrollSpearmen) {
        return 200
    }

    if unit.Equals(units.TrollSwordsmen) {
        return 400
    }

    if unit.Equals(units.TrollHalberdiers) {
        return 600
    }

    if unit.Equals(units.TrollShamans) {
        return 800
    }

    if unit.Equals(units.WarTrolls) {
        return 1000
    }

    if unit.Equals(units.WarMammoths) {
        return 800
    }

    if unit.Equals(units.MagicSpirit) {
        return 200
    }

    if unit.Equals(units.HellHounds) {
        return 500
    }

    if unit.Equals(units.Gargoyle) {
        return 600
    }

    if unit.Equals(units.FireGiant) {
        return 1000
    }

    if unit.Equals(units.FireElemental) {
        return 800
    }

    if unit.Equals(units.ChaosSpawn) {
        return 1200
    }

    if unit.Equals(units.Chimeras) {
        return 1300
    }

    if unit.Equals(units.DoomBat) {
        return 1100
    }

    if unit.Equals(units.Efreet) {
        return 1000
    }

    if unit.Equals(units.Hydra) {
        return 9000
    }

    if unit.Equals(units.GreatDrake) {
        return 12000
    }

    if unit.Equals(units.Skeleton) {
        return 120
    }

    if unit.Equals(units.Ghoul) {
        return 120
    }

    if unit.Equals(units.NightStalker) {
        return 200
    }

    if unit.Equals(units.WereWolf) {
        return 500
    }

    if unit.Equals(units.Demon) {
        return 700
    }

    if unit.Equals(units.Wraith) {
        return 1300
    }

    if unit.Equals(units.ShadowDemons) {
        return 1200
    }

    if unit.Equals(units.DeathKnights) {
        return 2500
    }

    if unit.Equals(units.DemonLord) {
        return 9000
    }

    if unit.Equals(units.Zombie) {
        return 300
    }

    if unit.Equals(units.Unicorn) {
        return 900
    }

    if unit.Equals(units.GuardianSpirit) {
        return 600
    }

    if unit.Equals(units.Angel) {
        return 1200
    }

    if unit.Equals(units.ArchAngel) {
        return 2000
    }

    if unit.Equals(units.WarBear) {
        return 450
    }

    if unit.Equals(units.Sprites) {
        return 500
    }

    if unit.Equals(units.Cockatrices) {
        return 900
    }

    if unit.Equals(units.Basilisk) {
        return 1300
    }

    if unit.Equals(units.GiantSpiders) {
        return 450
    }

    if unit.Equals(units.StoneGiant) {
        return 1200
    }

    if unit.Equals(units.Colossus) {
        return 2000
    }

    if unit.Equals(units.Gorgon) {
        return 1800
    }

    if unit.Equals(units.EarthElemental) {
        return 1800
    }

    if unit.Equals(units.Behemoth) {
        return 2500
    }

    if unit.Equals(units.GreatWyrm) {
        return 10000
    }

    if unit.Equals(units.FloatingIsland) {
        return 1000
    }

    if unit.Equals(units.PhantomBeast) {
        return 900
    }

    if unit.Equals(units.PhantomWarrior) {
        return 500
    }

    if unit.Equals(units.StormGiant) {
        return 2000
    }

    if unit.Equals(units.AirElemental) {
        return 1100
    }

    if unit.Equals(units.Djinn) {
        return 1600
    }

    if unit.Equals(units.SkyDrake) {
        return 12000
    }

    if unit.Equals(units.Nagas) {
        return 900
    }

    if unit.Equals(units.Catapult) {
        return 800
    }

    if unit.Equals(units.BarbarianSpearmen) {
        return 100
    }

    if unit.Equals(units.BarbarianSwordsmen) {
        return 150
    }

    if unit.Equals(units.BarbarianBowmen) {
        return 200
    }

    if unit.Equals(units.BarbarianCavalry) {
        return 350
    }

    if unit.Equals(units.BarbarianShaman) {
        return 400
    }

    if unit.Equals(units.Berserkers) {
        return 500
    }

    if unit.Equals(units.BeastmenSpearmen) {
        return 100
    }

    if unit.Equals(units.BeastmenSwordsmen) {
        return 150
    }

    if unit.Equals(units.BeastmenHalberdiers) {
        return 200
    }

    if unit.Equals(units.BeastmenBowmen) {
        return 300
    }

    if unit.Equals(units.BeastmenPriest) {
        return 400
    }

    if unit.Equals(units.BeastmenMagician) {
        return 700
    }

    if unit.Equals(units.BeastmenEngineer) {
        return 50
    }

    if unit.Equals(units.Centaur) {
        return 400
    }

    if unit.Equals(units.Manticore) {
        return 800
    }

    if unit.Equals(units.Minotaur) {
        return 800
    }

    if unit.Equals(units.DarkElfSpearmen) {
        return 200
    }

    if unit.Equals(units.DarkElfSwordsmen) {
        return 300
    }

    if unit.Equals(units.DarkElfHalberdiers) {
        return 400
    }

    if unit.Equals(units.DarkElfCavalry) {
        return 500
    }

    if unit.Equals(units.DarkElfPriests) {
        return 500
    }

    if unit.Equals(units.Nightblades) {
        return 500
    }

    if unit.Equals(units.Warlocks) {
        return 1200
    }

    if unit.Equals(units.Nightmares) {
        return 900
    }

    if unit.Equals(units.DraconianSpearmen) {
        return 200
    }

    if unit.Equals(units.DraconianSwordsmen) {
        return 300
    }

    if unit.Equals(units.DraconianHalberdiers) {
        return 400
    }

    if unit.Equals(units.DraconianBowmen) {
        return 400
    }

    if unit.Equals(units.DraconianShaman) {
        return 400
    }

    if unit.Equals(units.DraconianMagician) {
        return 800
    }

    if unit.Equals(units.DoomDrake) {
        return 900
    }

    if unit.Equals(units.AirShip) {
        return 1500
    }

    if unit.Equals(units.DwarfSwordsmen) {
        return 150
    }

    if unit.Equals(units.DwarfHalberdiers) {
        return 200
    }

    if unit.Equals(units.DwarfEngineer) {
        return 50
    }

    if unit.Equals(units.Hammerhands) {
        return 600
    }

    if unit.Equals(units.SteamCannon) {
        return 600
    }

    if unit.Equals(units.Golem) {
        return 800
    }

    if unit.Equals(units.GnollSpearmen) {
        return 100
    }

    if unit.Equals(units.GnollSwordsmen) {
        return 150
    }

    if unit.Equals(units.GnollHalberdiers) {
        return 200
    }

    if unit.Equals(units.GnollBowmen) {
        return 250
    }

    if unit.Equals(units.WolfRiders) {
        return 300
    }

    if unit.Equals(units.HalflingSpearmen) {
        return 100
    }

    if unit.Equals(units.HalflingSwordsmen) {
        return 150
    }

    if unit.Equals(units.HalflingBowmen) {
        return 200
    }

    if unit.Equals(units.HalflingShamans) {
        return 300
    }

    if unit.Equals(units.Slingers) {
        return 450
    }

    if unit.Equals(units.HighElfSpearmen) {
        return 100
    }

    if unit.Equals(units.HighElfSwordsmen) {
        return 150
    }

    if unit.Equals(units.HighElfHalberdiers) {
        return 200
    }

    if unit.Equals(units.HighElfCavalry) {
        return 300
    }

    if unit.Equals(units.HighElfMagician) {
        return 500
    }

    if unit.Equals(units.Longbowmen) {
        return 400
    }

    if unit.Equals(units.ElvenLord) {
        return 600
    }

    if unit.Equals(units.Pegasai) {
        return 600
    }

    if unit.Equals(units.HighMenEngineer) {
        return 50
    }

    if unit.Equals(units.HighMenSpearmen) {
        return 100
    }

    if unit.Equals(units.HighMenSwordsmen) {
        return 150
    }

    if unit.Equals(units.HighMenBowmen) {
        return 200
    }

    if unit.Equals(units.HighMenCavalry) {
        return 250
    }

    if unit.Equals(units.HighMenPriest) {
        return 300
    }

    if unit.Equals(units.HighMenMagician) {
        return 600
    }

    if unit.Equals(units.HighMenPikemen) {
        return 400
    }

    if unit.Equals(units.Paladin) {
        return 700
    }

    if unit.Equals(units.KlackonSpearmen) {
        return 100
    }

    if unit.Equals(units.KlackonSwordsmen) {
        return 150
    }

    if unit.Equals(units.KlackonHalberdiers) {
        return 250
    }

    if unit.Equals(units.KlackonEngineer) {
        return 50
    }

    if unit.Equals(units.StagBeetle) {
        return 600
    }

    return 1
}

func getUnitCost2(unit *units.Unit) uint64 {
    // Settlers always cost 10
    if unit.IsSettlers() {
        return 10
    }

    if unit.Race == data.RaceHero {
        return 500000
    }

    // Base cost: health, attack, defense, abilities
    health := unit.GetHitPoints() * unit.GetCount() / 2
    melee := unit.GetMeleeAttackPower() * unit.GetCount()
    ranged := unit.GetRangedAttackPower() * unit.RangedAttacks * unit.GetCount()
    var rangedMultiplier float32 = 1.0
    if unit.GetRangedAttackDamageType() == units.DamageRangedMagical {
        rangedMultiplier = 2
    }

    defense := unit.GetDefense() * unit.GetCount()
    resistance := unit.GetResistance() * unit.GetCount()

    // Ability modifier: +10% per ability, +20% for some strong ones
    var abilityValue float32 = 1
    for _, ability := range unit.GetAbilities() {
        switch ability.Ability {
            case data.AbilityArmorPiercing: abilityValue = 1.3
            case data.AbilityCauseFear: abilityValue = 1.8
            case data.AbilityColdImmunity: abilityValue = 1.1
            case data.AbilityDeathGaze: abilityValue = 1.5 * ability.Value
            case data.AbilityDeathImmunity: abilityValue = 2
            case data.AbilityDispelEvil: abilityValue = 1.1
            case data.AbilityDoomBoltSpell: abilityValue = 1.5
            case data.AbilityDoomGaze: abilityValue = 1.8 * ability.Value
            case data.AbilityDeathTouch: abilityValue = 2
            case data.AbilityFireballSpell: abilityValue = 2
            case data.AbilityFireBreath: abilityValue = 1.3 * ability.Value
            case data.AbilityFireImmunity: abilityValue = 1.2
            case data.AbilityFirstStrike: abilityValue = 1.5
            case data.AbilityHealingSpell: abilityValue = 1.4
            case data.AbilityHolyBonus: abilityValue = 1.3
            case data.AbilityIllusion: abilityValue = 1.6
            case data.AbilityIllusionsImmunity: abilityValue = 1.7
            case data.AbilityImmolation: abilityValue = 1.9
            case data.AbilityInvisibility: abilityValue = 2.4
            case data.AbilityLargeShield: abilityValue = 1.1
            case data.AbilityLifeSteal: abilityValue = 1.5 * -ability.Value
            case data.AbilityLightningBreath: abilityValue = 1.8 * ability.Value
            case data.AbilityLongRange: abilityValue = 1.1
            case data.AbilityMagicImmunity: abilityValue = 2.5
            case data.AbilityMerging: abilityValue = 2
            case data.AbilityMissileImmunity: abilityValue = 2
            case data.AbilityNegateFirstStrike: abilityValue = 1.3
            case data.AbilityNonCorporeal: abilityValue = 1.2
            case data.AbilityPathfinding: abilityValue = 1.1
            case data.AbilityPoisonImmunity: abilityValue = 1.3
            case data.AbilityPoisonTouch: abilityValue = 1.7 * ability.Value
            case data.AbilityRegeneration: abilityValue = 2.5
            case data.AbilityResistanceToAll: abilityValue = 1.9
            case data.AbilityStoningGaze: abilityValue = 1.6 * ability.Value
            case data.AbilityStoningImmunity: abilityValue = 1.5
            case data.AbilityStoningTouch: abilityValue = 1.6 * ability.Value
            case data.AbilitySummonDemons: abilityValue = 1.3
            case data.AbilityToHit: abilityValue = 1.3 * ability.Value / 10
            case data.AbilityTeleporting: abilityValue = 2
            case data.AbilityThrown: abilityValue = 1.4 * ability.Value
            case data.AbilityWallCrusher: abilityValue = 1.1
            case data.AbilityWeaponImmunity: abilityValue = 1.3
            case data.AbilityWebSpell: abilityValue = 1.3
        }
    }

    // Magic/fantastic units: add casting cost if present
    magicCost := 0
    if unit.CastingCost > 0 {
        magicCost = unit.CastingCost * 5
    }

    // Main cost formula
    cost := (float32(health)*3 + float32(melee)*4 + float32(ranged)*3*rangedMultiplier + float32(defense)*2 + float32(resistance) + float32(magicCost)) * abilityValue
    if cost < 10 {
        cost = 10
    }
    return uint64(cost)
}

func getEnchantmentRequirements(enchantment data.UnitEnchantment) data.WizardBook {
    life := func(n int) data.WizardBook {
        return data.WizardBook{Magic: data.LifeMagic, Count: n}
    }

    nature := func(n int) data.WizardBook {
        return data.WizardBook{Magic: data.NatureMagic, Count: n}
    }

    sorcery := func(n int) data.WizardBook {
        return data.WizardBook{Magic: data.SorceryMagic, Count: n}
    }

    death := func(n int) data.WizardBook {
        return data.WizardBook{Magic: data.DeathMagic, Count: n}
    }

    chaos := func(n int) data.WizardBook {
        return data.WizardBook{Magic: data.ChaosMagic, Count: n}
    }

    switch enchantment {
        case data.UnitEnchantmentGiantStrength: return nature(1)
        case data.UnitEnchantmentLionHeart: return life(2)
        case data.UnitEnchantmentHaste: return sorcery(7)
        case data.UnitEnchantmentImmolation: return chaos(4)
        case data.UnitEnchantmentResistElements: return nature(1)
        case data.UnitEnchantmentResistMagic: return sorcery(2)
        case data.UnitEnchantmentElementalArmor: return nature(3)
        case data.UnitEnchantmentBless: return life(2)
        case data.UnitEnchantmentRighteousness: return life(5)
        case data.UnitEnchantmentCloakOfFear: return death(4)
        case data.UnitEnchantmentTrueSight: return sorcery(4)
        case data.UnitEnchantmentPathFinding: return nature(1)
        case data.UnitEnchantmentFlight: return sorcery(5)
        case data.UnitEnchantmentChaosChannelsDemonWings: return chaos(4)
        case data.UnitEnchantmentChaosChannelsDemonSkin: return chaos(4)
        case data.UnitEnchantmentChaosChannelsFireBreath: return chaos(4)
        case data.UnitEnchantmentEndurance: return life(2)
        case data.UnitEnchantmentHeroism: return life(3)
        case data.UnitEnchantmentHolyArmor: return life(3)
        case data.UnitEnchantmentHolyWeapon: return life(3)
        case data.UnitEnchantmentInvulnerability: return life(8)
        case data.UnitEnchantmentIronSkin: return nature(5)
        case data.UnitEnchantmentRegeneration: return nature(7)
        case data.UnitEnchantmentStoneSkin: return nature(2)
        case data.UnitEnchantmentGuardianWind: return sorcery(5)
        case data.UnitEnchantmentInvisibility: return sorcery(7)
        case data.UnitEnchantmentMagicImmunity: return sorcery(9)
        case data.UnitEnchantmentSpellLock: return sorcery(6)
        case data.UnitEnchantmentEldritchWeapon: return chaos(2)
        case data.UnitEnchantmentFlameBlade: return chaos(3)
        case data.UnitEnchantmentBerserk: return death(4)
        case data.UnitEnchantmentBlackChannels: return death(5)
        case data.UnitEnchantmentWraithForm: return death(6)
    }

    return data.WizardBook{}
}

func getEnchantmentCost(enchantment data.UnitEnchantment) int {
    switch enchantment {
        case data.UnitEnchantmentGiantStrength: return 100
        case data.UnitEnchantmentLionHeart: return 200
        case data.UnitEnchantmentHaste: return 3000
        case data.UnitEnchantmentImmolation: return 200
        case data.UnitEnchantmentResistElements: return 100
        case data.UnitEnchantmentResistMagic: return 200
        case data.UnitEnchantmentElementalArmor: return 200
        case data.UnitEnchantmentBless: return 150
        case data.UnitEnchantmentRighteousness: return 300
        case data.UnitEnchantmentCloakOfFear: return 200
        case data.UnitEnchantmentTrueSight: return 150
        case data.UnitEnchantmentPathFinding: return 100
        case data.UnitEnchantmentFlight: return 300
        case data.UnitEnchantmentChaosChannelsDemonWings: return 200
        case data.UnitEnchantmentChaosChannelsDemonSkin: return 200
        case data.UnitEnchantmentChaosChannelsFireBreath: return 200
        case data.UnitEnchantmentEndurance: return 100
        case data.UnitEnchantmentHeroism: return 200
        case data.UnitEnchantmentHolyArmor: return 100
        case data.UnitEnchantmentHolyWeapon: return 200
        case data.UnitEnchantmentInvulnerability: return 1000
        case data.UnitEnchantmentIronSkin: return 100
        case data.UnitEnchantmentRegeneration: return 3000
        case data.UnitEnchantmentStoneSkin: return 100
        case data.UnitEnchantmentGuardianWind: return 300
        case data.UnitEnchantmentInvisibility: return 500
        case data.UnitEnchantmentMagicImmunity: return 500
        case data.UnitEnchantmentSpellLock: return 300
        case data.UnitEnchantmentEldritchWeapon: return 100
        case data.UnitEnchantmentFlameBlade: return 150
        case data.UnitEnchantmentBerserk: return 200
        case data.UnitEnchantmentBlackChannels: return 400
        case data.UnitEnchantmentWraithForm: return 300
    }

    return 0
}

/*
 * f(0) = 0
 * f(n) = f(n-1) + base * (1 + 0.1 * n)
 */
func computeManaCost(amount int) uint64 {
    var total float64 = 0

    baseCost := 2.0

    for i := range amount {
        total = total + baseCost * (1 + 0.1 * float64(i))
    }

    return uint64(total)
}

// how many spells of each rarity type the player can currently buy
func getCommonSpells(books int) int {
    if books >= 5 {
        return 10
    }

    if books <= 0 {
        return 0
    }

    return books * 2
}

func getUncommonSpells(books int) int {
    books = books - 3

    if books >= 5 {
        return 10
    }

    if books <= 0 {
        return 0
    }

    return books * 2
}

func getRareSpells(books int) int {
    books = books - 6

    if books >= 5 {
        return 10
    }

    if books <= 0 {
        return 0
    }

    return books * 2
}

func getVeryRareSpells(books int) int {
    if books >= 11 {
        return 10
    }

    books = books - 9
    if books >= 5 {
        return 10
    }

    if books <= 0 {
        return 0
    }

    return books
}
