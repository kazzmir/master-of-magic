package units

type NormalExperienceLevel int

const (
    ExperienceRecruit NormalExperienceLevel = iota
    ExperienceRegular
    ExperienceVeteran
    ExperienceElite
    // also needs either warlord retort or crusade enchantment
    ExperienceUltraElite
    // needs both warlord retort and crusade enchantment in play
    ExperienceChampionNormal
)

type HeroExperienceLevel int

const (
    ExperienceHero HeroExperienceLevel = iota
    ExperienceMyrmidon
    ExperienceCaptain
    ExperienceCommander
    ExperienceChampionHero
    ExperienceLord
    ExperienceGrandLord
    ExperienceSuperHero
    ExperienceDemiGod
)

func (level NormalExperienceLevel) ExperienceRequired(warlordRetort bool, crusade bool) int {
    switch level {
        case ExperienceRecruit: return 0
        case ExperienceRegular:
            if warlordRetort || crusade {
                return 0
            }
            return 20
        case ExperienceVeteran:
            if warlordRetort && crusade {
                return 0
            }
            if warlordRetort || crusade {
                return 20
            }
            return 60
        case ExperienceElite:
            if warlordRetort && crusade {
                return 20
            }
            if warlordRetort || crusade {
                return 60
            }
            return 120
        case ExperienceUltraElite:
            // needs at least one
            if warlordRetort && crusade {
                return 60
            }
            return 120
        case ExperienceChampionNormal:
            // needs both
            return 120
    }

    return 0
}

func GetNormalExperienceLevel(experience int, warlordRetort bool, crusade bool) NormalExperienceLevel {
    if experience < ExperienceRegular.ExperienceRequired(warlordRetort, crusade) {
        return ExperienceRecruit
    }
    if experience < ExperienceVeteran.ExperienceRequired(warlordRetort, crusade) {
        return ExperienceRegular
    }
    if experience < ExperienceElite.ExperienceRequired(warlordRetort, crusade) {
        return ExperienceVeteran
    }

    if warlordRetort && crusade && experience >= ExperienceChampionNormal.ExperienceRequired(warlordRetort, crusade) {
        return ExperienceChampionNormal
    }

    if (warlordRetort || crusade) && experience >= ExperienceUltraElite.ExperienceRequired(warlordRetort, crusade) {
        return ExperienceUltraElite
    }

    return ExperienceElite
}

func (level HeroExperienceLevel) ExperienceRequired() int {
    switch level {
        case ExperienceHero: return 0
        case ExperienceMyrmidon: return 20
        case ExperienceCaptain: return 60
        case ExperienceCommander: return 120
        case ExperienceChampionHero: return 200
        case ExperienceLord: return 300
        case ExperienceGrandLord: return 450
        case ExperienceSuperHero: return 600
        case ExperienceDemiGod: return 1000
    }

    return 0
}

func GetHeroExperienceLevel(experience int, warlordRetort bool, crusade bool) HeroExperienceLevel {
    levels := []HeroExperienceLevel{
        ExperienceHero, ExperienceMyrmidon, ExperienceCaptain,
        ExperienceCommander, ExperienceChampionHero, ExperienceLord,
        ExperienceGrandLord, ExperienceSuperHero, ExperienceDemiGod,
    }

    for i := 0; i < len(levels) - 1; i++ {
        if experience < levels[i + 1].ExperienceRequired() {
            return levels[i]
        }
    }

    return levels[len(levels) - 1]
}
