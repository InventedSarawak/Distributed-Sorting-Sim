package algorithms

import (
	"fmt"
	"net"

	"github.com/InventedSarawak/Distributed-Sorting-Sim/internal/simulator"
	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/types"
)

type OddEvenPayload struct {
	Value int `json:"value"`
}

func RunOddEven(
	n *types.Node[OddEvenPayload],
	engine *simulator.SimulatorEngine[OddEvenPayload],
	leftBuf, rightBuf *simulator.RoundBuffer[OddEvenPayload],
	sendFunc func(net.Conn, types.Message[OddEvenPayload]) error,
	debug bool,
) {
	if debug {
		fmt.Printf("[Algo] Node %d: Starting Sort (Value: %d)\n", n.ID, n.Value.Value)
	}

	for round := 0; round < n.TotalNode; round++ {
		isEvenRound := round%2 == 0
		isEvenNode := n.ID%2 == 0

		var targetID int
		var isLeftExchange bool

		if isEvenRound {
			if isEvenNode {
				targetID = n.ID + 1
				isLeftExchange = false
			} else {
				targetID = n.ID - 1
				isLeftExchange = true
			}
		} else {
			if isEvenNode {
				targetID = n.ID - 1
				isLeftExchange = true
			} else {
				targetID = n.ID + 1
				isLeftExchange = false
			}
		}

		if targetID >= 0 && targetID < n.TotalNode {
			msg := types.Message[OddEvenPayload]{
				SenderID:   n.ID,
				ReceiverID: targetID,
				Round:      round,
				Body:       n.Value,
				Type:       types.MsgData,
			}

			conn := n.RightConn
			if isLeftExchange {
				conn = n.LeftConn
			}

			_ = sendFunc(conn, msg)

			var neighborMsg types.Message[OddEvenPayload]
			if isLeftExchange {
				neighborMsg = leftBuf.GetStepMessage(round)
			} else {
				neighborMsg = rightBuf.GetStepMessage(round)
			}

			oldVal := n.Value.Value
			neighborVal := neighborMsg.Body.Value

			if isLeftExchange {
				if n.Value.Value < neighborVal {
					n.Value.Value = neighborVal
				}
			} else {
				if n.Value.Value > neighborVal {
					n.Value.Value = neighborVal
				}
			}

			if debug && oldVal != n.Value.Value {
				fmt.Printf("[Algo] Node %d: Swapped %d -> %d (Round %d)\n", n.ID, oldVal, n.Value.Value, round)
			}
		}
		engine.IncrementClock(n)
	}

	if debug {
		fmt.Printf("[Algo] Node %d: Sort Complete. Final Value: %d\n", n.ID, n.Value.Value)
	}
}
