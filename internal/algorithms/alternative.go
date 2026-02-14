package algorithms

import (
	"fmt"
	"net"
	"sort"

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
		fmt.Printf("[Algo] Node %d: Starting Alternative Sort (Value: %d)\n", n.ID, n.Value.Value)
	}

	for round := 1; round < n.TotalNode; round++ {
		phaseIndex := (round + 1) % 3
		var phaseStartID int
		switch phaseIndex {
		case 0:
			phaseStartID = 2
		case 1:
			phaseStartID = 0
		default:
			phaseStartID = 1
		}

		isCenterNode := (n.ID >= phaseStartID) && ((n.ID-phaseStartID)%3 == 0)

		msg := types.Message[AlternativePayload]{
			SenderID: n.ID,
			Round:    round,
			Body:     n.Value,
			Type:     types.MsgData,
		}

		if isCenterNode {
			var leftValue, rightValue int
			var hasLeftNeighbor, hasRightNeighbor bool

			if n.Position != types.Head {
				leftMsg, _ := simulator.WaitForNeighbors(n, round, leftBuf, rightBuf)
				if leftMsg != nil {
					leftValue = leftMsg.Body.Value
					hasLeftNeighbor = true
				}
			}

			if n.Position != types.Tail {
				_, rightMsg := simulator.WaitForNeighbors(n, round, leftBuf, rightBuf)
				if rightMsg != nil {
					rightValue = rightMsg.Body.Value
					hasRightNeighbor = true
				}
			}

			leftCandidate, centerCandidate, rightCandidate := leftValue, n.Value.Value, rightValue

			if !hasLeftNeighbor && hasRightNeighbor {
				if centerCandidate > rightCandidate {
					centerCandidate, rightCandidate = rightCandidate, centerCandidate
				}
			} else if hasLeftNeighbor && !hasRightNeighbor {
				if leftCandidate > centerCandidate {
					leftCandidate, centerCandidate = centerCandidate, leftCandidate
				}
			} else if hasLeftNeighbor && hasRightNeighbor {
				ordered := []int{leftValue, n.Value.Value, rightValue}
				sort.Ints(ordered)
				leftCandidate, centerCandidate, rightCandidate = ordered[0], ordered[1], ordered[2]
			}

			n.Value.Value = centerCandidate

			if hasLeftNeighbor {
				msg.ReceiverID = n.ID - 1
				msg.Body.Value = leftCandidate
				_ = sendFunc(n.LeftConn, msg)
			}
			if hasRightNeighbor {
				msg.ReceiverID = n.ID + 1
				msg.Body.Value = rightCandidate
				_ = sendFunc(n.RightConn, msg)
			}

		} else {
			isLeftWing := (n.ID+1 >= phaseStartID) && ((n.ID+1-phaseStartID)%3 == 0) && n.Position != types.Tail
			isRightWing := (n.ID-1 >= phaseStartID) && ((n.ID-1-phaseStartID)%3 == 0) && n.Position != types.Head

			if isLeftWing {
				msg.ReceiverID = n.ID + 1
				_ = sendFunc(n.RightConn, msg)

				_, rightMsg := simulator.WaitForNeighbors(n, round, leftBuf, rightBuf)
				if rightMsg != nil {
					n.Value.Value = rightMsg.Body.Value
				}

			} else if isRightWing {
				msg.ReceiverID = n.ID - 1
				_ = sendFunc(n.LeftConn, msg)

				leftMsg, _ := simulator.WaitForNeighbors(n, round, leftBuf, rightBuf)
				if leftMsg != nil {
					n.Value.Value = leftMsg.Body.Value
				}
			}
		}

		engine.IncrementClock(n)
	}

	if debug {
		fmt.Printf("[Algo] Node %d: Alternative Sort Complete. Final: %d\n", n.ID, n.Value.Value)
	}
}
