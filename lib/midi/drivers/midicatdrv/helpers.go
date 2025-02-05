package midicatdrv

import (
	"github.com/kazzmir/master-of-magic/lib/midi/drivers"
)

type inPorts []drivers.In

func (i inPorts) Len() int {
	return len(i)
}

func (i inPorts) Swap(a, b int) {
	i[a], i[b] = i[b], i[a]
}

func (i inPorts) Less(a, b int) bool {
	return i[a].Number() < i[b].Number()
}

type outPorts []drivers.Out

func (i outPorts) Len() int {
	return len(i)
}

func (i outPorts) Swap(a, b int) {
	i[a], i[b] = i[b], i[a]
}

func (i outPorts) Less(a, b int) bool {
	return i[a].Number() < i[b].Number()
}
