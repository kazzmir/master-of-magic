package units

// FIXME: probably move this somewhere else
type Race int

const (
    RaceLizard Race = iota
)

type Unit struct {
    LbxFile string
    Index int
    Race Race
}

var LizardSpearmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 0,
    Race: RaceLizard,
}

var LizardSwordsmen Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 1,
    Race: RaceLizard,
}

var LizardHalberdiers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 2,
    Race: RaceLizard,
}

var LizardJavelineers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 3,
    Race: RaceLizard,
}

var LizardShamans Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 4,
    Race: RaceLizard,
}

var Settlers Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 5,
    Race: RaceLizard,
}

var DragonTurtle Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 6,
    Race: RaceLizard,
}

var GiantBat Unit = Unit{
    LbxFile: "units2.lbx",
    Index: 41,
    Race: RaceLizard,
}
