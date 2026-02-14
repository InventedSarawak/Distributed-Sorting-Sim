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
		isEvenNodeID := n.ID%2 == 0

		var partnerID int
		var exchangeWithLeft bool

		if isEvenRound {
			if isEvenNodeID {
				partnerID = n.ID + 1
				exchangeWithLeft = false
			} else {
				partnerID = n.ID - 1
				exchangeWithLeft = true
			}
		} else {
			if isEvenNodeID {
				partnerID = n.ID - 1
				exchangeWithLeft = true
			} else {
				partnerID = n.ID + 1
				exchangeWithLeft = false
			}
		}

		if partnerID >= 0 && partnerID < n.TotalNode {
			msg := types.Message[OddEvenPayload]{
				SenderID:   n.ID,
				ReceiverID: partnerID,
				Round:      round,
				Body:       n.Value,
				Type:       types.MsgData,
			}

			conn := n.RightConn
			if exchangeWithLeft {
				conn = n.LeftConn
			}

			_ = sendFunc(conn, msg)

			var neighborMsg types.Message[OddEvenPayload]
			if exchangeWithLeft {
				neighborMsg = leftBuf.GetStepMessage(round)
			} else {
				neighborMsg = rightBuf.GetStepMessage(round)
			}

			previousValue := n.Value.Value
			neighborValue := neighborMsg.Body.Value

			if exchangeWithLeft {
				if n.Value.Value < neighborValue {
					n.Value.Value = neighborValue
				}
			} else {
				if n.Value.Value > neighborValue {
					n.Value.Value = neighborValue
				}
			}

			if debug && previousValue != n.Value.Value {
				fmt.Printf("[Algo] Node %d: Swapped %d -> %d (Round %d)\n", n.ID, previousValue, n.Value.Value, round)
			}
		}
		engine.IncrementClock(n)
	}

	if debug {
		fmt.Printf("[Algo] Node %d: Sort Complete. Final Value: %d\n", n.ID, n.Value.Value)
	}
}
