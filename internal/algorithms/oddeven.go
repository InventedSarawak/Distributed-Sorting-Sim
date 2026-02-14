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
) {
	fmt.Printf("[Algo] Node %d: Starting Sort (Value: %d)\n", n.ID, n.Value.Value)

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
			// Send
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

			// fmt.Printf("[Algo] Node %d: Round %d - Sending to %d\n", n.ID, round, targetID)
			_ = sendFunc(conn, msg)

			// Receive
			// fmt.Printf("[Algo] Node %d: Round %d - Waiting for %d\n", n.ID, round, targetID)
			var neighborMsg types.Message[OddEvenPayload]
			if isLeftExchange {
				neighborMsg = leftBuf.GetStepMessage(round)
			} else {
				neighborMsg = rightBuf.GetStepMessage(round)
			}

			// Compare
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

			if oldVal != n.Value.Value {
				fmt.Printf("[Algo] Node %d: Swapped %d -> %d (Round %d)\n", n.ID, oldVal, n.Value.Value, round)
			}
		}

		engine.IncrementClock(n)
	}
	fmt.Printf("[Algo] Node %d: Sort Complete. Final Value: %d\n", n.ID, n.Value.Value)
}
