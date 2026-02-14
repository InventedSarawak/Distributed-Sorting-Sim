package algorithms

import (
	"fmt"
	"math"
	"net"

	"github.com/InventedSarawak/Distributed-Sorting-Sim/internal/simulator"
	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/types"
)

type SasakiElement struct {
	Value    int  `json:"value"`
	IsMarked bool `json:"is_marked"`
}

type SasakiPayload struct {
	LValue SasakiElement `json:"l_value"`
	RValue SasakiElement `json:"r_value"`
	Area   int           `json:"area"`

	Value int `json:"value"`
}

func RunSasaki(
	n *types.Node[SasakiPayload],
	engine *simulator.SimulatorEngine[SasakiPayload],
	leftBuf, rightBuf *simulator.RoundBuffer[SasakiPayload],
	sendFunc func(net.Conn, types.Message[SasakiPayload]) error,
	debug bool,
) {
	initialValue := n.Value.Value

	switch n.Position {
	case types.Head:
		n.Value.LValue = SasakiElement{Value: math.MinInt32, IsMarked: false}
		n.Value.RValue = SasakiElement{Value: initialValue, IsMarked: true}
		n.Value.Area = -1
	case types.Tail:
		n.Value.LValue = SasakiElement{Value: initialValue, IsMarked: true}
		n.Value.RValue = SasakiElement{Value: math.MaxInt32, IsMarked: false}
		n.Value.Area = 0
	default:
		n.Value.LValue = SasakiElement{Value: initialValue, IsMarked: false}
		n.Value.RValue = SasakiElement{Value: initialValue, IsMarked: false}
		n.Value.Area = 0
	}

	if debug {
		fmt.Printf("[Algo] Node %d: Init Sasaki Area=%d L=%v R=%v\n", n.ID, n.Value.Area, n.Value.LValue, n.Value.RValue)
	}

	for round := 1; round < n.TotalNode; round++ {
		msg := types.Message[SasakiPayload]{
			SenderID: n.ID,
			Round:    round,
			Type:     types.MsgData,
		}

		if n.Position != types.Tail {
			msg.ReceiverID = n.ID + 1
			msg.Body.LValue = n.Value.RValue
			_ = sendFunc(n.RightConn, msg)
		}

		if n.Position != types.Head {
			msg.ReceiverID = n.ID - 1
			msg.Body.RValue = n.Value.LValue
			_ = sendFunc(n.LeftConn, msg)
		}

		leftMsg, rightMsg := simulator.WaitForNeighbors(n, round, leftBuf, rightBuf)

		if n.Position != types.Head && leftMsg != nil {
			leftIncomingRight := leftMsg.Body.RValue
			if leftIncomingRight.Value > n.Value.LValue.Value {
				if leftIncomingRight.IsMarked {
					n.Value.Area--
				}
				if n.Value.LValue.IsMarked {
					n.Value.Area++
				}
				n.Value.LValue = leftIncomingRight
			}
		}

		if n.Position != types.Tail && rightMsg != nil {
			rightIncomingLeft := rightMsg.Body.LValue
			if rightIncomingLeft.Value < n.Value.RValue.Value {
				if rightIncomingLeft.IsMarked {
					n.Value.Area++
				}
				if n.Value.RValue.IsMarked {
					n.Value.Area--
				}
				n.Value.RValue = rightIncomingLeft
			}
		}

		if n.Value.LValue.Value > n.Value.RValue.Value {
			leftTemp := n.Value.LValue
			n.Value.LValue = n.Value.RValue
			n.Value.RValue = leftTemp
		}

		engine.IncrementClock(n)
	}

	if n.Value.Area == -1 {
		n.Value.Value = n.Value.RValue.Value
	} else {
		n.Value.Value = n.Value.LValue.Value
	}

	if debug {
		fmt.Printf("[Algo] Node %d: Sasaki Complete. Final: %d (Area: %d)\n", n.ID, n.Value.Value, n.Value.Area)
	}
}
