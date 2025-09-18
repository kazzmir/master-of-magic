package units

import (
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

type ExperienceInfo interface {
    HasWarlord() bool
    Crusade() bool
}

type NoExperienceInfo struct {
}

func (noInfo *NoExperienceInfo) HasWarlord() bool {
    return false
}

func (noInfo *NoExperienceInfo) Crusade() bool {
    return false
}

type ExperienceData interface {
    ToInt() int
    Name() string
}

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

func (level *NormalExperienceLevel) ToInt() int {
    return int(*level)
}

func (level NormalExperienceLevel) Name() string {
    switch level {
        case ExperienceRecruit: return "Recruit"
        case ExperienceRegular: return "Regular"
        case ExperienceVeteran: return "Veteran"
        case ExperienceElite: return "Elite"
        case ExperienceUltraElite: return "Ultra Elite"
        case ExperienceChampionNormal: return "Champion"
    }

    return ""
}

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

func (level HeroExperienceLevel) ExperienceRequired(warlordRetort bool, crusade bool) int {
    switch level {
        case ExperienceHero: return 0
        case ExperienceMyrmidon:
            if warlordRetort || crusade {
                return 0
            }
            return 20
        case ExperienceCaptain:
            if warlordRetort && crusade {
                return 0
            }
            if warlordRetort || crusade {
                return 20
            }
            return 60
        case ExperienceCommander:
            if warlordRetort && crusade {
                return 20
            }
            if warlordRetort || crusade {
                return 60
            }
            return 120
        case ExperienceChampionHero:
            if warlordRetort && crusade {
                return 60
            }
            if warlordRetort || crusade {
                return 120
            }
            return 200
        case ExperienceLord:
            if warlordRetort && crusade {
                return 120
            }
            if warlordRetort || crusade {
                return 200
            }
            return 300
        case ExperienceGrandLord:
            if warlordRetort && crusade {
                return 200
            }
            if warlordRetort || crusade {
                return 300
            }
            return 450
        case ExperienceSuperHero:
            if warlordRetort && crusade {
                return 300
            }
            if warlordRetort || crusade {
                return 450
            }
            return 600
        case ExperienceDemiGod:
            if warlordRetort && crusade {
                return 450
            }
            if warlordRetort || crusade {
                return 600
            }
            return 1000
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
        if experience < levels[i + 1].ExperienceRequired(warlordRetort, crusade) {
            return levels[i]
        }
    }

    return levels[len(levels) - 1]
}

func (hero *HeroExperienceLevel) ToInt() int {
    return int(*hero)
}

func (hero *HeroExperienceLevel) Name() string {
    switch *hero {
        case ExperienceHero: return "Hero"
        case ExperienceMyrmidon: return "Myrmidon"
        case ExperienceCaptain: return "Captain"
        case ExperienceCommander: return "Commander"
        case ExperienceChampionHero: return "Champion"
        case ExperienceLord: return "Lord"
        case ExperienceGrandLord: return "Grand Lord"
        case ExperienceSuperHero: return "Super Hero"
        case ExperienceDemiGod: return "Demigod"
    }

    return ""
}

type Badge int
const (
    BadgeNone Badge = iota
    BadgeSilver
    BadgeGold
    BadgeRed
)

// index into main.lbx of the icon images
func (badge Badge) IconLbxIndex() int {
    switch badge {
        case BadgeNone: return -1
        case BadgeSilver: return 51
        case BadgeGold: return 52
        case BadgeRed: return 53
    }

    return -1
}

type ExperienceBadge struct {
    Badge Badge
    Count int
}

type BadgeUnit interface {
    GetRace() data.Race
    GetHeroExperienceLevel() HeroExperienceLevel
    GetExperienceLevel() NormalExperienceLevel
}

// return an object that contains the badge type and the number of icons
func GetExperienceBadge(unit BadgeUnit) ExperienceBadge {
    if unit.GetRace() == data.RaceHero {
        switch unit.GetHeroExperienceLevel() {
            case ExperienceHero:
            case ExperienceMyrmidon:
                return ExperienceBadge{
                    Badge: BadgeSilver,
                    Count: 1,
                }
            case ExperienceCaptain:
                return ExperienceBadge{
                    Badge: BadgeSilver,
                    Count: 2,
                }
            case ExperienceCommander:
                return ExperienceBadge{
                    Badge: BadgeSilver,
                    Count: 3,
                }
            case ExperienceChampionHero:
                return ExperienceBadge{
                    Badge: BadgeGold,
                    Count: 1,
                }
            case ExperienceLord:
                return ExperienceBadge{
                    Badge: BadgeGold,
                    Count: 2,
                }
            case ExperienceGrandLord:
                return ExperienceBadge{
                    Badge: BadgeGold,
                    Count: 3,
                }
            case ExperienceSuperHero:
                return ExperienceBadge{
                    Badge: BadgeRed,
                    Count: 1,
                }
            case ExperienceDemiGod:
                return ExperienceBadge{
                    Badge: BadgeRed,
                    Count: 2,
                }
        }
    } else {

        switch unit.GetExperienceLevel() {
            case ExperienceRecruit:
                // nothing
            case ExperienceRegular:
                // one white circle
                return ExperienceBadge{
                    Badge: BadgeSilver,
                    Count: 1,
                }
            case ExperienceVeteran:
                // two white circles
                return ExperienceBadge{
                    Badge: BadgeSilver,
                    Count: 2,
                }
            case ExperienceElite:
                // three white circles
                return ExperienceBadge{
                    Badge: BadgeSilver,
                    Count: 3,
                }
            case ExperienceUltraElite:
                // one yellow
                return ExperienceBadge{
                    Badge: BadgeGold,
                    Count: 1,
                }
            case ExperienceChampionNormal:
                // two yellow
                return ExperienceBadge{
                    Badge: BadgeGold,
                    Count: 2,
                }
        }
    }

    return ExperienceBadge{}
}
