package algorithms

import (
	"github.com/InventedSarawak/Distributed-Sorting-Sim/internal/simulator"
	"github.com/InventedSarawak/Distributed-Sorting-Sim/internal/transport"
	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/types"
)

// OddEvenPayload represents the data exchanged during the sort.
type OddEvenPayload struct {
	Value int `json:"value"`
}

// RunOddEven runs the $N$ rounds of the transposition sort.
func RunOddEven(n *types.Node[OddEvenPayload], engine *simulator.SimulatorEngine[OddEvenPayload], leftBuf, rightBuf *simulator.RoundBuffer[OddEvenPayload]) {
	// Total rounds required for Odd-Even is equal to the number of nodes
	for round := 0; round < n.TotalNode; round++ {
		isEvenRound := round%2 == 0
		isEvenNode := n.ID%2 == 0

		// Determine if this node should compare with Left or Right this round
		var targetID int
		var isLeftExchange bool

		if isEvenRound {
			if isEvenNode {
				targetID = n.ID + 1 // Even node pairs with Right (ID+1)
				isLeftExchange = false
			} else {
				targetID = n.ID - 1 // Odd node pairs with Left (ID-1)
				isLeftExchange = true
			}
		} else {
			if isEvenNode {
				targetID = n.ID - 1 // Even node pairs with Left (ID-1)
				isLeftExchange = true
			} else {
				targetID = n.ID + 1 // Odd node pairs with Right (ID+1)
				isLeftExchange = false
			}
		}

		// Execute exchange if neighbor exists
		if targetID >= 0 && targetID < n.TotalNode {
			// 1. Send current value to neighbor
			msg := types.Message[OddEvenPayload]{
				SenderID: n.ID,
				Round:    round,
				Body:     n.Value,
				Type:     types.MsgData,
			}

			conn := n.RightConn
			if isLeftExchange {
				conn = n.LeftConn
			}

			transport.SendMessage(conn, msg)

			// 2. Receive neighbor's value using the RoundBuffer barrier
			var neighborMsg types.Message[OddEvenPayload]
			if isLeftExchange {
				neighborMsg = leftBuf.GetStepMessage(round)
			} else {
				neighborMsg = rightBuf.GetStepMessage(round)
			}

			// 3. Compare and Swap
			neighborVal := neighborMsg.Body.Value
			if isLeftExchange {
				// If I am the right side of the pair, I take the MAX
				if n.Value.Value < neighborVal {
					n.Value.Value = neighborVal
				}
			} else {
				// If I am the left side of the pair, I take the MIN
				if n.Value.Value > neighborVal {
					n.Value.Value = neighborVal
				}
			}
		}

		engine.IncrementClock(n)
	}
}
