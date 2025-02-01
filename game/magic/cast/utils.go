package cast

import (
    "github.com/kazzmir/master-of-magic/game/magic/units"
)

func SummonUnitForSpell(spellName string) units.Unit {
    switch spellName {
        case "Magic Spirit":
            return units.MagicSpirit
        case "Angel":
            return units.Angel
        case "Arch Angel":
            return units.ArchAngel
        case "Guardian Spirit":
            return units.GuardianSpirit
        case "Unicorns":
            return units.Unicorn
        case "Basilisk":
            return units.Basilisk
        case "Behemoth":
            return units.Behemoth
        case "Cockatrices":
            return units.Cockatrice
        case "Colossus":
            return units.Colossus
        case "Earth Elemental":
            return units.EarthElemental
        case "Giant Spiders":
            return units.GiantSpider
        case "Gorgons":
            return units.Gorgon
        case "Great Wyrm":
            return units.GreatWyrm
        case "Sprites":
            return units.Sprite
        case "Stone Giant":
            return units.StoneGiant
        case "War Bears":
            return units.WarBear
        case "Air Elemental":
            return units.AirElemental
        case "Djinn":
            return units.Djinn
        case "Floating Island":
            return units.FloatingIsland
        case "Nagas":
            return units.Nagas
        case "Phantom Beast":
            return units.PhantomBeast
        case "Phantom Warriors":
            return units.PhantomWarrior
        case "Sky Drake":
            return units.SkyDrake
        case "Storm Giant":
            return units.StormGiant
        case "Chaos Spawn":
            return units.ChaosSpawn
        case "Chimeras":
            return units.Chimeras
        case "Doom Bat":
            return units.DoomBat
        case "Efreet":
            return units.Efreet
        case "Fire Elemental":
            return units.FireElemental
        case "Fire Giant":
            return units.FireGiant
        case "Gargoyles":
            return units.Gargoyle
        case "Great Drake":
            return units.GreatDrake
        case "Hell Hounds":
            return units.HellHounds
        case "Hydra":
            return units.Hydra
        case "Death Knights":
            return units.DeathKnight
        case "Demon Lord":
            return units.DemonLord
        case "Ghouls":
            return units.Ghoul
        case "Night Stalker":
            return units.NightStalker
        case "Shadow Demons":
            return units.ShadowDemon
        case "Skeletons":
            return units.Skeleton
        case "Wraiths":
            return units.Wraith
    }

    return units.UnitNone
}
