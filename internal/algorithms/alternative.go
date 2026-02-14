package algorithms

import (
	"fmt"
	"net"

	"github.com/InventedSarawak/Distributed-Sorting-Sim/internal/simulator"
	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/types"
)

type AlternativePayload struct {
	Value int `json:"value"`
}

func RunAlternative(
	n *types.Node[AlternativePayload],
	engine *simulator.SimulatorEngine[AlternativePayload],
	leftBuf, rightBuf *simulator.RoundBuffer[AlternativePayload],
	sendFunc func(net.Conn, types.Message[AlternativePayload]) error,
	debug bool,
) {
	if debug {
		fmt.Printf("[Algo] Node %d: Starting Alternative Sort\n", n.ID)
	}

	for round := 0; round < n.TotalNode; round++ {
		engine.IncrementClock(n)
	}
}
