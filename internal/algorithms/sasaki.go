package algorithms

import (
	"fmt"
	"net"

	"github.com/InventedSarawak/Distributed-Sorting-Sim/internal/simulator"
	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/types"
)

type SasakiPayload struct {
	Value    int  `json:"value"`
	IsMarked bool `json:"is_marked"`
}

func RunSasaki(
	n *types.Node[SasakiPayload],
	engine *simulator.SimulatorEngine[SasakiPayload],
	leftBuf, rightBuf *simulator.RoundBuffer[SasakiPayload],
	sendFunc func(net.Conn, types.Message[SasakiPayload]) error,
	debug bool,
) {
	if debug {
		fmt.Printf("[Algo] Node %d: Starting Sasaki Sort\n", n.ID)
	}

	for round := 0; round < n.TotalNode; round++ {
		engine.IncrementClock(n)
	}
}
