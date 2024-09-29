package hero

type HireScreenResult int

const (
    HireScreenRunning HireScreenResult = iota
    HireScreenHire
    HireScreenReject
)

type HireScreen struct {
    State HireScreenResult
}

func (hire *HireScreen) Update() HireScreenResult {
    return hire.State
}

func (hire *HireScreen) Draw(screen *ebiten.Image) {
}
